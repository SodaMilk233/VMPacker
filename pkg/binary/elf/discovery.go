package elf

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
)

const (
	ptGNUEHFrame = elf.ProgType(0x6474e550)

	dwEHPAbsptr   = 0x00
	dwEHPOmit     = 0xFF
	dwEHPPcrel    = 0x10
	dwEHPTextrel  = 0x20
	dwEHPDatarel  = 0x30
	dwEHPFuncrel  = 0x40
	dwEHPAligned  = 0x50
	dwEHPIndirect = 0x80

	maxEHRecordSize = 64 << 20
	maxFDECount     = 1 << 20
)

// FunctionCandidate is a function range recovered from ELF metadata.
type FunctionCandidate struct {
	Name       string
	Addr       uint64
	Size       uint64
	Source     string
	Confidence string
}

// FunctionDiscovery contains usable candidates and non-fatal parser warnings.
type FunctionDiscovery struct {
	Functions []FunctionCandidate
	Warnings  []string
}

// DiscoverFunctions merges symbol tables and unwind FDE ranges. Missing symbol
// tables are expected for stripped binaries and are not treated as errors.
func DiscoverFunctions(f *elf.File) FunctionDiscovery {
	result := FunctionDiscovery{}
	byAddress := make(map[uint64]FunctionCandidate)

	add := func(candidate FunctionCandidate) {
		if candidate.Name == "" {
			candidate.Name = fmt.Sprintf("sub_%X", candidate.Addr)
		}
		if !isExecutableFileRange(f, candidate.Addr, candidate.Size) {
			return
		}

		current, exists := byAddress[candidate.Addr]
		if !exists || candidatePriority(candidate.Source) > candidatePriority(current.Source) {
			byAddress[candidate.Addr] = candidate
		}
	}

	addSymbols := func(symbols []elf.Symbol, source string) {
		for _, symbol := range symbols {
			if elf.ST_TYPE(symbol.Info) != elf.STT_FUNC || symbol.Section == elf.SHN_UNDEF || symbol.Size == 0 {
				continue
			}
			add(FunctionCandidate{
				Name:       symbol.Name,
				Addr:       symbol.Value,
				Size:       symbol.Size,
				Source:     source,
				Confidence: "high",
			})
		}
	}

	if symbols, err := f.Symbols(); err == nil {
		addSymbols(symbols, ".symtab")
	} else if !errors.Is(err, elf.ErrNoSymbols) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("reading .symtab failed: %v", err))
	}
	if symbols, err := f.DynamicSymbols(); err == nil {
		addSymbols(symbols, ".dynsym")
	} else if !errors.Is(err, elf.ErrNoSymbols) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("reading .dynsym failed: %v", err))
	}

	unwindFunctions, unwindWarnings := discoverUnwindFunctions(f)
	for _, candidate := range unwindFunctions {
		add(candidate)
	}
	result.Warnings = append(result.Warnings, unwindWarnings...)

	result.Functions = make([]FunctionCandidate, 0, len(byAddress))
	for _, candidate := range byAddress {
		result.Functions = append(result.Functions, candidate)
	}
	sort.Slice(result.Functions, func(i, j int) bool {
		if result.Functions[i].Addr == result.Functions[j].Addr {
			return result.Functions[i].Size < result.Functions[j].Size
		}
		return result.Functions[i].Addr < result.Functions[j].Addr
	})
	return result
}

// NativeBranchTargets collects exact destinations that may safely receive an
// external B tail call: recovered function starts plus verified PLT thunks.
func NativeBranchTargets(f *elf.File, functions []FunctionCandidate) []uint64 {
	seen := make(map[uint64]struct{}, len(functions))
	for _, function := range functions {
		seen[function.Addr] = struct{}{}
	}
	for _, target := range discoverPLTBranchTargets(f) {
		seen[target] = struct{}{}
	}
	targets := make([]uint64, 0, len(seen))
	for target := range seen {
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

// discoverPLTBranchTargets returns verified AArch64 PLT thunk entries. These
// are valid native tail-call destinations but are not protectable functions.
func discoverPLTBranchTargets(f *elf.File) []uint64 {
	targets := make(map[uint64]struct{})
	scan := func(base uint64, data []byte) {
		for offset := 0; offset+16 <= len(data); offset += 4 {
			if isAArch64PLTStub(data[offset : offset+16]) {
				targets[base+uint64(offset)] = struct{}{}
			}
		}
	}

	if section := f.Section(".plt"); section != nil {
		if data, err := section.Data(); err == nil {
			scan(section.Addr, data)
		}
	}
	if len(targets) == 0 {
		// Section headers may be stripped. The instruction signature is strict
		// enough to scan executable file-backed segments without trusting names.
		for _, program := range f.Progs {
			if program.Type != elf.PT_LOAD || program.Flags&elf.PF_X == 0 || program.Filesz < 16 {
				continue
			}
			data, err := io.ReadAll(program.Open())
			if err == nil {
				scan(program.Vaddr, data)
			}
		}
	}

	result := make([]uint64, 0, len(targets))
	for target := range targets {
		result = append(result, target)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func isAArch64PLTStub(code []byte) bool {
	if len(code) < 16 {
		return false
	}
	adrp := binary.LittleEndian.Uint32(code[0:4])
	ldr := binary.LittleEndian.Uint32(code[4:8])
	add := binary.LittleEndian.Uint32(code[8:12])
	br := binary.LittleEndian.Uint32(code[12:16])
	return adrp&0x9F00001F == 0x90000010 &&
		ldr&0xFFC003FF == 0xF9400211 &&
		add&0xFFC003FF == 0x91000210 &&
		br == 0xD61F0220
}

func candidatePriority(source string) int {
	switch source {
	case ".symtab":
		return 3
	case ".dynsym":
		return 2
	case ".eh_frame":
		return 1
	default:
		return 0
	}
}

func isExecutableFileRange(f *elf.File, addr, size uint64) bool {
	if size == 0 || addr&3 != 0 || size&3 != 0 {
		return false
	}
	end, ok := checkedAdd(addr, size)
	if !ok {
		return false
	}
	for _, program := range f.Progs {
		if program.Type != elf.PT_LOAD || program.Flags&elf.PF_X == 0 {
			continue
		}
		programEnd, ok := checkedAdd(program.Vaddr, program.Filesz)
		if ok && addr >= program.Vaddr && end <= programEnd {
			return true
		}
	}
	return false
}

func checkedAdd(a, b uint64) (uint64, bool) {
	result := a + b
	return result, result >= a
}

type virtualRegion struct {
	addr   uint64
	size   uint64
	reader io.ReaderAt
}

type elfVirtualImage struct {
	order       binary.ByteOrder
	addressSize int
	textBase    uint64
	regions     []virtualRegion
}

func newELFVirtualImage(f *elf.File) *elfVirtualImage {
	addressSize := 4
	if f.Class == elf.ELFCLASS64 {
		addressSize = 8
	}
	image := &elfVirtualImage{order: f.ByteOrder, addressSize: addressSize}
	textBaseSet := false
	for _, program := range f.Progs {
		if program.Filesz > 0 {
			image.regions = append(image.regions, virtualRegion{
				addr: program.Vaddr, size: program.Filesz, reader: program,
			})
		}
		if program.Type == elf.PT_LOAD && program.Flags&elf.PF_X != 0 &&
			(!textBaseSet || program.Vaddr < image.textBase) {
			image.textBase = program.Vaddr
			textBaseSet = true
		}
	}
	for _, section := range f.Sections {
		if section.Size == 0 || section.Type == elf.SHT_NOBITS {
			continue
		}
		image.regions = append(image.regions, virtualRegion{
			addr: section.Addr, size: section.Size, reader: section,
		})
	}
	return image
}

func (image *elfVirtualImage) read(addr, size uint64) ([]byte, error) {
	if size == 0 || size > maxEHRecordSize {
		return nil, fmt.Errorf("invalid virtual read size 0x%X", size)
	}
	end, ok := checkedAdd(addr, size)
	if !ok {
		return nil, fmt.Errorf("virtual address overflow at 0x%X", addr)
	}
	for _, region := range image.regions {
		regionEnd, valid := checkedAdd(region.addr, region.size)
		if !valid || addr < region.addr || end > regionEnd {
			continue
		}
		buffer := make([]byte, int(size))
		n, err := region.reader.ReadAt(buffer, int64(addr-region.addr))
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if n != len(buffer) {
			return nil, io.ErrUnexpectedEOF
		}
		return buffer, nil
	}
	return nil, fmt.Errorf("virtual address range 0x%X-0x%X is not file-backed", addr, end)
}

func discoverUnwindFunctions(f *elf.File) ([]FunctionCandidate, []string) {
	image := newELFVirtualImage(f)
	var warnings []string

	headerData, headerAddr, headerErr := findEHFrameHeader(f)
	if headerErr != nil {
		warnings = append(warnings, headerErr.Error())
	}
	if len(headerData) > 0 {
		functions, failures, err := parseEHFrameHeader(image, headerData, headerAddr)
		if failures > 0 {
			warnings = append(warnings, fmt.Sprintf(".eh_frame_hdr skipped %d malformed FDE entries", failures))
		}
		if err == nil && len(functions) > 0 {
			return functions, warnings
		}
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("parsing .eh_frame_hdr failed: %v", err))
		}
	}

	section := f.Section(".eh_frame")
	if section == nil {
		return nil, warnings
	}
	if section.Size > maxEHRecordSize {
		warnings = append(warnings, fmt.Sprintf(".eh_frame is too large: 0x%X", section.Size))
		return nil, warnings
	}
	data, err := section.Data()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("reading .eh_frame failed: %v", err))
		return nil, warnings
	}
	functions, failures, err := parseEHFrameSection(image, data, section.Addr)
	if failures > 0 {
		warnings = append(warnings, fmt.Sprintf(".eh_frame skipped %d malformed FDE records", failures))
	}
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("parsing .eh_frame failed: %v", err))
	}
	return functions, warnings
}

func findEHFrameHeader(f *elf.File) ([]byte, uint64, error) {
	if section := f.Section(".eh_frame_hdr"); section != nil {
		if section.Size > maxEHRecordSize {
			return nil, 0, fmt.Errorf(".eh_frame_hdr is too large: 0x%X", section.Size)
		}
		data, err := section.Data()
		return data, section.Addr, err
	}
	for _, program := range f.Progs {
		if program.Type != ptGNUEHFrame || program.Filesz == 0 {
			continue
		}
		if program.Filesz > maxEHRecordSize {
			return nil, 0, fmt.Errorf("PT_GNU_EH_FRAME is too large: 0x%X", program.Filesz)
		}
		data := make([]byte, int(program.Filesz))
		n, err := program.ReadAt(data, 0)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, 0, err
		}
		if n != len(data) {
			return nil, 0, io.ErrUnexpectedEOF
		}
		return data, program.Vaddr, nil
	}
	return nil, 0, nil
}

type ehBases struct {
	text     uint64
	data     uint64
	function uint64
}

func parseEHFrameHeader(image *elfVirtualImage, data []byte, addr uint64) ([]FunctionCandidate, int, error) {
	if len(data) < 4 {
		return nil, 0, io.ErrUnexpectedEOF
	}
	if data[0] != 1 {
		return nil, 0, fmt.Errorf("unsupported .eh_frame_hdr version %d", data[0])
	}

	ehFrameEncoding := data[1]
	countEncoding := data[2]
	tableEncoding := data[3]
	if countEncoding == dwEHPOmit || tableEncoding == dwEHPOmit {
		return nil, 0, fmt.Errorf(".eh_frame_hdr omits the FDE table")
	}

	cursor := 4
	bases := ehBases{text: image.textBase, data: addr}
	if ehFrameEncoding != dwEHPOmit {
		if _, err := decodeEHValue(image, data, addr, &cursor, ehFrameEncoding, image.addressSize, bases); err != nil {
			return nil, 0, fmt.Errorf("decoding eh_frame pointer: %w", err)
		}
	}
	count, err := decodeEHValue(image, data, addr, &cursor, countEncoding, image.addressSize, bases)
	if err != nil {
		return nil, 0, fmt.Errorf("decoding FDE count: %w", err)
	}
	if count > maxFDECount {
		return nil, 0, fmt.Errorf("unreasonable FDE count %d", count)
	}

	functions := make([]FunctionCandidate, 0, int(count))
	failures := 0
	for i := uint64(0); i < count; i++ {
		tableStart, err := decodeEHValue(image, data, addr, &cursor, tableEncoding, image.addressSize, bases)
		if err != nil {
			return functions, failures, fmt.Errorf("decoding table entry %d start: %w", i, err)
		}
		fdeAddr, err := decodeEHValue(image, data, addr, &cursor, tableEncoding, image.addressSize, bases)
		if err != nil {
			return functions, failures, fmt.Errorf("decoding table entry %d FDE: %w", i, err)
		}
		start, size, err := parseFDEAt(image, fdeAddr, addr)
		if err != nil {
			failures++
			continue
		}
		if start != tableStart {
			failures++
			continue
		}
		functions = append(functions, FunctionCandidate{
			Name:       fmt.Sprintf("sub_%X", start),
			Addr:       start,
			Size:       size,
			Source:     ".eh_frame",
			Confidence: "high",
		})
	}
	return functions, failures, nil
}

func parseEHFrameSection(image *elfVirtualImage, data []byte, addr uint64) ([]FunctionCandidate, int, error) {
	var functions []FunctionCandidate
	failures := 0
	for offset := 0; offset < len(data); {
		total, idOffset, idSize, terminator, err := ehRecordLayout(data[offset:], image.order)
		if err != nil {
			return functions, failures, fmt.Errorf("record at offset 0x%X: %w", offset, err)
		}
		if terminator {
			break
		}
		if offset+total > len(data) {
			return functions, failures, io.ErrUnexpectedEOF
		}
		record := data[offset : offset+total]
		id := readSizedUnsigned(record[idOffset:idOffset+idSize], image.order)
		if id != 0 {
			start, size, err := parseFDEAt(image, addr+uint64(offset), addr)
			if err != nil {
				failures++
			} else {
				functions = append(functions, FunctionCandidate{
					Name:       fmt.Sprintf("sub_%X", start),
					Addr:       start,
					Size:       size,
					Source:     ".eh_frame",
					Confidence: "high",
				})
			}
		}
		offset += total
	}
	return functions, failures, nil
}

type cieInfo struct {
	fdeEncoding byte
	addressSize int
}

func parseFDEAt(image *elfVirtualImage, addr, dataBase uint64) (uint64, uint64, error) {
	record, idOffset, idSize, err := readEHRecord(image, addr)
	if err != nil {
		return 0, 0, err
	}
	cieDelta := readSizedUnsigned(record[idOffset:idOffset+idSize], image.order)
	if cieDelta == 0 {
		return 0, 0, fmt.Errorf("record at 0x%X is a CIE", addr)
	}
	idAddr := addr + uint64(idOffset)
	if cieDelta > idAddr {
		return 0, 0, fmt.Errorf("invalid CIE back-reference 0x%X", cieDelta)
	}
	cieAddr := idAddr - cieDelta
	cieRecord, cieIDOffset, cieIDSize, err := readEHRecord(image, cieAddr)
	if err != nil {
		return 0, 0, fmt.Errorf("reading CIE at 0x%X: %w", cieAddr, err)
	}
	if readSizedUnsigned(cieRecord[cieIDOffset:cieIDOffset+cieIDSize], image.order) != 0 {
		return 0, 0, fmt.Errorf("record at 0x%X is not a CIE", cieAddr)
	}
	cie, err := parseCIE(image, cieRecord, cieAddr, cieIDOffset+cieIDSize, dataBase)
	if err != nil {
		return 0, 0, err
	}

	cursor := idOffset + idSize
	bases := ehBases{text: image.textBase, data: dataBase}
	start, err := decodeEHValue(image, record, addr, &cursor, cie.fdeEncoding, cie.addressSize, bases)
	if err != nil {
		return 0, 0, fmt.Errorf("decoding FDE start: %w", err)
	}
	rangeEncoding := cie.fdeEncoding & 0x0F
	size, err := decodeEHValue(image, record, addr, &cursor, rangeEncoding, cie.addressSize, ehBases{})
	if err != nil {
		return 0, 0, fmt.Errorf("decoding FDE range: %w", err)
	}
	if size == 0 {
		return 0, 0, fmt.Errorf("zero-length FDE at 0x%X", addr)
	}
	return start, size, nil
}

func parseCIE(image *elfVirtualImage, record []byte, addr uint64, cursor int, dataBase uint64) (cieInfo, error) {
	info := cieInfo{fdeEncoding: dwEHPAbsptr, addressSize: image.addressSize}
	if cursor >= len(record) {
		return info, io.ErrUnexpectedEOF
	}
	version := record[cursor]
	cursor++
	augmentation, err := readCString(record, &cursor)
	if err != nil {
		return info, err
	}
	if version == 4 {
		if cursor+2 > len(record) {
			return info, io.ErrUnexpectedEOF
		}
		info.addressSize = int(record[cursor])
		segmentSize := record[cursor+1]
		cursor += 2
		if segmentSize != 0 || (info.addressSize != 4 && info.addressSize != 8) {
			return info, fmt.Errorf("unsupported CIE address/segment size %d/%d", info.addressSize, segmentSize)
		}
	}
	if _, err := decodeULEB(record, &cursor); err != nil {
		return info, err
	}
	if _, err := decodeSLEB(record, &cursor); err != nil {
		return info, err
	}
	if version == 1 {
		if cursor >= len(record) {
			return info, io.ErrUnexpectedEOF
		}
		cursor++
	} else if _, err := decodeULEB(record, &cursor); err != nil {
		return info, err
	}

	if len(augmentation) == 0 {
		return info, nil
	}
	if augmentation[0] != 'z' {
		return info, fmt.Errorf("unsupported CIE augmentation %q", augmentation)
	}
	augmentationLength, err := decodeULEB(record, &cursor)
	if err != nil {
		return info, err
	}
	augmentationEnd64, ok := checkedAdd(uint64(cursor), augmentationLength)
	if !ok || augmentationEnd64 > uint64(len(record)) {
		return info, io.ErrUnexpectedEOF
	}
	augmentationEnd := int(augmentationEnd64)
	bases := ehBases{text: image.textBase, data: dataBase}
	for i := 1; i < len(augmentation); i++ {
		switch augmentation[i] {
		case 'L':
			if cursor >= augmentationEnd {
				return info, io.ErrUnexpectedEOF
			}
			cursor++
		case 'R':
			if cursor >= augmentationEnd {
				return info, io.ErrUnexpectedEOF
			}
			info.fdeEncoding = record[cursor]
			cursor++
		case 'P':
			if cursor >= augmentationEnd {
				return info, io.ErrUnexpectedEOF
			}
			encoding := record[cursor]
			cursor++
			if _, err := decodeEHValue(image, record, addr, &cursor, encoding, info.addressSize, bases); err != nil {
				return info, fmt.Errorf("decoding personality pointer: %w", err)
			}
		case 'S':
			// Signal frame marker has no augmentation payload.
		default:
			return info, fmt.Errorf("unsupported CIE augmentation character %q", augmentation[i])
		}
	}
	return info, nil
}

func readEHRecord(image *elfVirtualImage, addr uint64) ([]byte, int, int, error) {
	prefix, err := image.read(addr, 12)
	if err != nil {
		return nil, 0, 0, err
	}
	total, idOffset, idSize, terminator, err := ehRecordLayout(prefix, image.order)
	if err != nil {
		return nil, 0, 0, err
	}
	if terminator {
		return nil, 0, 0, fmt.Errorf("unexpected .eh_frame terminator at 0x%X", addr)
	}
	record, err := image.read(addr, uint64(total))
	return record, idOffset, idSize, err
}

func ehRecordLayout(data []byte, order binary.ByteOrder) (int, int, int, bool, error) {
	if len(data) < 4 {
		return 0, 0, 0, false, io.ErrUnexpectedEOF
	}
	length32 := order.Uint32(data[:4])
	if length32 == 0 {
		return 4, 0, 0, true, nil
	}
	if length32 == 0xFFFFFFFF {
		if len(data) < 12 {
			return 0, 0, 0, false, io.ErrUnexpectedEOF
		}
		length64 := order.Uint64(data[4:12])
		if length64 < 8 || length64 > maxEHRecordSize-12 {
			return 0, 0, 0, false, fmt.Errorf("invalid 64-bit record length 0x%X", length64)
		}
		return int(12 + length64), 12, 8, false, nil
	}
	if length32 < 4 || length32 > maxEHRecordSize-4 {
		return 0, 0, 0, false, fmt.Errorf("invalid record length 0x%X", length32)
	}
	return int(4 + length32), 4, 4, false, nil
}

func decodeEHValue(image *elfVirtualImage, data []byte, dataAddr uint64, cursor *int, encoding byte, addressSize int, bases ehBases) (uint64, error) {
	if encoding == dwEHPOmit {
		return 0, fmt.Errorf("omitted encoded value")
	}
	application := encoding & 0x70
	if application == dwEHPAligned {
		currentAddr := dataAddr + uint64(*cursor)
		alignedAddr := (currentAddr + uint64(addressSize-1)) &^ uint64(addressSize-1)
		*cursor += int(alignedAddr - currentAddr)
		application = 0
	}
	valueAddr := dataAddr + uint64(*cursor)
	format := encoding & 0x0F

	var unsigned uint64
	var signed int64
	isSigned := false
	var err error
	switch format {
	case 0x00:
		unsigned, err = readUnsigned(data, cursor, addressSize, image.order)
	case 0x01:
		unsigned, err = decodeULEB(data, cursor)
	case 0x02:
		unsigned, err = readUnsigned(data, cursor, 2, image.order)
	case 0x03:
		unsigned, err = readUnsigned(data, cursor, 4, image.order)
	case 0x04:
		unsigned, err = readUnsigned(data, cursor, 8, image.order)
	case 0x08:
		signed, err = readSigned(data, cursor, addressSize, image.order)
		isSigned = true
	case 0x09:
		signed, err = decodeSLEB(data, cursor)
		isSigned = true
	case 0x0A:
		signed, err = readSigned(data, cursor, 2, image.order)
		isSigned = true
	case 0x0B:
		signed, err = readSigned(data, cursor, 4, image.order)
		isSigned = true
	case 0x0C:
		signed, err = readSigned(data, cursor, 8, image.order)
		isSigned = true
	default:
		return 0, fmt.Errorf("unsupported DW_EH_PE format 0x%X", format)
	}
	if err != nil {
		return 0, err
	}

	base := uint64(0)
	switch application {
	case 0:
	case dwEHPPcrel:
		base = valueAddr
	case dwEHPTextrel:
		base = bases.text
	case dwEHPDatarel:
		base = bases.data
	case dwEHPFuncrel:
		base = bases.function
	default:
		return 0, fmt.Errorf("unsupported DW_EH_PE application 0x%X", application)
	}

	var value uint64
	if isSigned {
		if signed < 0 {
			delta := uint64(-(signed + 1)) + 1
			if delta > base {
				return 0, fmt.Errorf("encoded pointer underflow")
			}
			value = base - delta
		} else {
			var ok bool
			value, ok = checkedAdd(base, uint64(signed))
			if !ok {
				return 0, fmt.Errorf("encoded pointer overflow")
			}
		}
	} else {
		var ok bool
		value, ok = checkedAdd(base, unsigned)
		if !ok {
			return 0, fmt.Errorf("encoded pointer overflow")
		}
	}

	if encoding&dwEHPIndirect != 0 {
		pointer, err := image.read(value, uint64(addressSize))
		if err != nil {
			return 0, err
		}
		value = readSizedUnsigned(pointer, image.order)
	}
	return value, nil
}

func readUnsigned(data []byte, cursor *int, size int, order binary.ByteOrder) (uint64, error) {
	if size != 2 && size != 4 && size != 8 {
		return 0, fmt.Errorf("unsupported integer size %d", size)
	}
	if *cursor < 0 || *cursor+size > len(data) {
		return 0, io.ErrUnexpectedEOF
	}
	value := readSizedUnsigned(data[*cursor:*cursor+size], order)
	*cursor += size
	return value, nil
}

func readSigned(data []byte, cursor *int, size int, order binary.ByteOrder) (int64, error) {
	value, err := readUnsigned(data, cursor, size, order)
	if err != nil {
		return 0, err
	}
	switch size {
	case 2:
		return int64(int16(value)), nil
	case 4:
		return int64(int32(value)), nil
	case 8:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("unsupported signed integer size %d", size)
	}
}

func readSizedUnsigned(data []byte, order binary.ByteOrder) uint64 {
	switch len(data) {
	case 2:
		return uint64(order.Uint16(data))
	case 4:
		return uint64(order.Uint32(data))
	case 8:
		return order.Uint64(data)
	default:
		return 0
	}
}

func decodeULEB(data []byte, cursor *int) (uint64, error) {
	var result uint64
	for shift := uint(0); shift < 64; shift += 7 {
		if *cursor >= len(data) {
			return 0, io.ErrUnexpectedEOF
		}
		value := data[*cursor]
		(*cursor)++
		if shift == 63 && value&0x7E != 0 {
			return 0, fmt.Errorf("ULEB128 overflow")
		}
		result |= uint64(value&0x7F) << shift
		if value&0x80 == 0 {
			return result, nil
		}
	}
	return 0, fmt.Errorf("ULEB128 overflow")
}

func decodeSLEB(data []byte, cursor *int) (int64, error) {
	var result uint64
	var value byte
	shift := uint(0)
	for ; shift < 64; shift += 7 {
		if *cursor >= len(data) {
			return 0, io.ErrUnexpectedEOF
		}
		value = data[*cursor]
		(*cursor)++
		if shift == 63 && value&0x7E != 0 && value&0x7E != 0x7E {
			return 0, fmt.Errorf("SLEB128 overflow")
		}
		result |= uint64(value&0x7F) << shift
		if value&0x80 == 0 {
			shift += 7
			if shift < 64 && value&0x40 != 0 {
				result |= ^uint64(0) << shift
			}
			return int64(result), nil
		}
	}
	return 0, fmt.Errorf("SLEB128 overflow")
}

func readCString(data []byte, cursor *int) (string, error) {
	start := *cursor
	for *cursor < len(data) {
		if data[*cursor] == 0 {
			value := string(data[start:*cursor])
			(*cursor)++
			return value, nil
		}
		(*cursor)++
	}
	return "", io.ErrUnexpectedEOF
}
