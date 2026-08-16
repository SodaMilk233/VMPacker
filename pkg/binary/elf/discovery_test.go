package elf

import (
	"bytes"
	stdelf "debug/elf"
	"encoding/binary"
	"testing"
)

func TestIsAArch64PLTStub(t *testing.T) {
	words := []uint32{0x90001FF0, 0xF9449611, 0x9124A210, 0xD61F0220}
	code := make([]byte, 16)
	for i, word := range words {
		binary.LittleEndian.PutUint32(code[i*4:], word)
	}
	if !isAArch64PLTStub(code) {
		t.Fatal("valid AArch64 PLT entry was not recognized")
	}
	code[12] ^= 1
	if isAArch64PLTStub(code) {
		t.Fatal("non-PLT instruction sequence was accepted")
	}
}

func TestParseEHFrameHeader(t *testing.T) {
	memory, headerAddr, functionAddr, functionSize := buildEHFixture()
	const imageBase = uint64(0x1000)
	headerOffset := int(headerAddr - imageBase)
	header := memory[headerOffset:]

	image := &elfVirtualImage{
		order:       binary.LittleEndian,
		addressSize: 8,
		regions: []virtualRegion{{
			addr: imageBase, size: uint64(len(memory)), reader: bytes.NewReader(memory),
		}},
	}

	functions, failures, err := parseEHFrameHeader(image, header[:20], headerAddr)
	if err != nil {
		t.Fatalf("parseEHFrameHeader failed: %v", err)
	}
	if failures != 0 {
		t.Fatalf("unexpected malformed FDE count: %d", failures)
	}
	if len(functions) != 1 {
		t.Fatalf("expected one function, got %d", len(functions))
	}
	if functions[0].Addr != functionAddr || functions[0].Size != uint64(functionSize) {
		t.Fatalf("unexpected range: 0x%X + 0x%X", functions[0].Addr, functions[0].Size)
	}
	if functions[0].Source != ".eh_frame" || functions[0].Confidence != "high" {
		t.Fatalf("unexpected metadata: %+v", functions[0])
	}
}

func TestDiscoverFunctionsFromPTGNUEHFrame(t *testing.T) {
	memory, headerAddr, functionAddr, functionSize := buildEHFixture()
	const imageBase = uint64(0x1000)
	headerOffset := int(headerAddr - imageBase)

	f := &stdelf.File{
		FileHeader: stdelf.FileHeader{
			Class: stdelf.ELFCLASS64, Data: stdelf.ELFDATA2LSB,
			ByteOrder: binary.LittleEndian, Machine: stdelf.EM_AARCH64,
		},
		Progs: []*stdelf.Prog{
			{
				ProgHeader: stdelf.ProgHeader{
					Type: stdelf.PT_LOAD, Flags: stdelf.PF_R | stdelf.PF_X,
					Vaddr: imageBase, Filesz: uint64(len(memory)), Memsz: uint64(len(memory)),
				},
				ReaderAt: bytes.NewReader(memory),
			},
			{
				ProgHeader: stdelf.ProgHeader{
					Type: ptGNUEHFrame, Flags: stdelf.PF_R,
					Vaddr: headerAddr, Filesz: 20, Memsz: 20,
				},
				ReaderAt: bytes.NewReader(memory[headerOffset : headerOffset+20]),
			},
		},
	}

	result := DiscoverFunctions(f)
	if len(result.Warnings) != 0 {
		t.Fatalf("unexpected discovery warnings: %v", result.Warnings)
	}
	if len(result.Functions) != 1 {
		t.Fatalf("expected one function, got %d", len(result.Functions))
	}
	function := result.Functions[0]
	if function.Addr != functionAddr || function.Size != functionSize || function.Source != ".eh_frame" {
		t.Fatalf("unexpected candidate: %+v", function)
	}
}

func TestParseEHFrameSection(t *testing.T) {
	memory, _, functionAddr, functionSize := buildEHFixture()
	const (
		imageBase = uint64(0x1000)
		fdeEnd    = 0x14 + 4 + 13
	)
	image := &elfVirtualImage{
		order:       binary.LittleEndian,
		addressSize: 8,
		regions: []virtualRegion{{
			addr: imageBase, size: uint64(len(memory)), reader: bytes.NewReader(memory),
		}},
	}
	functions, failures, err := parseEHFrameSection(image, memory[:fdeEnd+4], imageBase)
	if err != nil || failures != 0 {
		t.Fatalf("parseEHFrameSection failed: failures=%d err=%v", failures, err)
	}
	if len(functions) != 1 || functions[0].Addr != functionAddr || functions[0].Size != functionSize {
		t.Fatalf("unexpected functions: %+v", functions)
	}
}

func TestParseEHFrameHeaderRejectsTruncatedTable(t *testing.T) {
	image := &elfVirtualImage{order: binary.LittleEndian, addressSize: 8}
	header := []byte{1, dwEHPOmit, 0x03, 0x3B, 1, 0, 0, 0}
	if _, _, err := parseEHFrameHeader(image, header, 0x1000); err == nil {
		t.Fatal("expected truncated table error")
	}
}

func TestParseAddrSpecExplicitRanges(t *testing.T) {
	tests := []struct {
		input string
		addr  uint64
		end   uint64
		name  string
	}{
		{input: "0x4000-0x4080:verify", addr: 0x4000, end: 0x4080, name: "verify"},
		{input: "0x4000:0x80:verify", addr: 0x4000, end: 0x4080, name: "verify"},
		{input: "0x4000:0x80", addr: 0x4000, end: 0x4080, name: "sub_4000"},
		{input: "0x4000", addr: 0x4000, end: 0, name: "sub_4000"},
	}
	for _, test := range tests {
		spec, err := ParseAddrSpec(test.input)
		if err != nil {
			t.Fatalf("ParseAddrSpec(%q): %v", test.input, err)
		}
		if spec.Addr != test.addr || spec.End != test.end || spec.Name != test.name {
			t.Fatalf("ParseAddrSpec(%q) = %+v", test.input, spec)
		}
	}
}

func TestParseAddrSpecRejectsInvalidSize(t *testing.T) {
	for _, input := range []string{"", "0x4000:0", "0x4000-0x4000", "0xFFFFFFFFFFFFFFFF:2"} {
		if _, err := ParseAddrSpec(input); err == nil {
			t.Fatalf("ParseAddrSpec(%q) unexpectedly succeeded", input)
		}
	}
}

func buildEHFixture() ([]byte, uint64, uint64, uint64) {
	const (
		imageBase    = uint64(0x1000)
		cieOffset    = 0x00
		fdeOffset    = 0x14
		headerOffset = 0x80
		functionAddr = uint64(0x4000)
		functionSize = uint32(0x80)
	)
	memory := make([]byte, 0x4000)

	cieBody := []byte{
		0, 0, 0, 0,
		1, 'z', 'R', 0,
		1, 0x78, 30,
		1, 0x1B,
		0, 0, 0,
	}
	binary.LittleEndian.PutUint32(memory[cieOffset:], uint32(len(cieBody)))
	copy(memory[cieOffset+4:], cieBody)

	fdeAddr := imageBase + fdeOffset
	fdeBody := make([]byte, 13)
	binary.LittleEndian.PutUint32(fdeBody[0:], uint32(fdeAddr+4-imageBase))
	startFieldAddr := fdeAddr + 8
	binary.LittleEndian.PutUint32(fdeBody[4:], uint32(int32(functionAddr-startFieldAddr)))
	binary.LittleEndian.PutUint32(fdeBody[8:], functionSize)
	fdeBody[12] = 0
	binary.LittleEndian.PutUint32(memory[fdeOffset:], uint32(len(fdeBody)))
	copy(memory[fdeOffset+4:], fdeBody)

	headerAddr := imageBase + headerOffset
	header := memory[headerOffset:]
	header[0] = 1
	header[1] = 0x1B
	header[2] = 0x03
	header[3] = 0x3B
	binary.LittleEndian.PutUint32(header[4:], uint32(int32(int64(imageBase)-int64(headerAddr+4))))
	binary.LittleEndian.PutUint32(header[8:], 1)
	binary.LittleEndian.PutUint32(header[12:], uint32(int32(functionAddr-headerAddr)))
	binary.LittleEndian.PutUint32(header[16:], uint32(int32(int64(fdeAddr)-int64(headerAddr))))

	return memory, headerAddr, functionAddr, uint64(functionSize)
}
