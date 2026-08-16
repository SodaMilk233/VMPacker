package api

import (
	"context"
	"debug/elf"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	elfpacker "github.com/vmpacker/pkg/binary/elf"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed vm_interp.bin
var interpBlob []byte

// VMPEngine API Interface for Frontend
type VMPEngine struct {
	ctx     context.Context
	mu      sync.Mutex
	isBusy  bool
	dataDir string
}

// NewVMPEngine create new VMPEngine
func NewVMPEngine() *VMPEngine {
	// Use executable directory as data dir
	exe, _ := os.Executable()
	dataDir := filepath.Dir(exe)
	return &VMPEngine{dataDir: dataDir}
}

// Startup registers the Context
func (e *VMPEngine) Startup(ctx context.Context) {
	e.ctx = ctx
}

// SelectFile prompts user to select a file
func (e *VMPEngine) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(e.ctx, runtime.OpenDialogOptions{
		Title: "Select ELF Executable",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Executable Files (*.elf, *.exe, etc.)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return selection, nil
}

// SelectSaveFile prompts user to select a save path
func (e *VMPEngine) SelectSaveFile(defaultFilename string) (string, error) {
	selection, err := runtime.SaveFileDialog(e.ctx, runtime.SaveDialogOptions{
		Title:           "选择保存路径",
		DefaultFilename: defaultFilename,
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

// AnalyzeELF reads binary information, verifying ARM64 format and extracting functions
func (e *VMPEngine) AnalyzeELF(filePath string) (map[string]interface{}, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	fileName := filepath.Base(filePath)

	// Open the file as an ELF
	f, err := elf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ELF file: %v", err)
	}
	defer f.Close()

	if f.Machine != elf.EM_AARCH64 {
		return nil, fmt.Errorf("unsupported architecture: only ARM64 is supported")
	}

	discovery := elfpacker.DiscoverFunctions(f)
	branchTargets := elfpacker.NativeBranchTargets(f, discovery.Functions)
	checker := elfpacker.NewPacker(filePath, "", nil, nil, false, false, false, true, nil)
	funcs := []map[string]interface{}{}
	compatibleCount := 0
	for _, candidate := range discovery.Functions {
		isAddressCandidate := candidate.Source != ".symtab" && candidate.Source != ".dynsym"
		fi, findErr := checker.FindFunctionByAddr(f, elfpacker.AddrSpec{
			Addr: candidate.Addr,
			End:  candidate.Addr + candidate.Size,
			Name: candidate.Name,
		})
		compatibility := elfpacker.FunctionCompatibility{Reason: "function range could not be read"}
		if findErr != nil {
			compatibility.Reason = findErr.Error()
		} else {
			compatibility = checker.CheckFunctionCompatibility(f, fi, branchTargets)
		}
		if compatibility.Compatible {
			compatibleCount++
		}
		funcs = append(funcs, map[string]interface{}{
			"name":                candidate.Name,
			"address":             fmt.Sprintf("0x%X", candidate.Addr),
			"size":                candidate.Size,
			"source":              candidate.Source,
			"confidence":          candidate.Confidence,
			"protection":          "Virtualization",
			"isCustom":            isAddressCandidate,
			"compatible":          compatibility.Compatible,
			"compatibilityReason": compatibility.Reason,
			"unsupportedCount":    compatibility.UnsupportedCount,
		})
	}

	return map[string]interface{}{
		"fileName":          fileName,
		"filePath":          filePath,
		"arch":              "ARM64",
		"format":            "ELF",
		"functions":         funcs,
		"warnings":          discovery.Warnings,
		"compatibleCount":   compatibleCount,
		"incompatibleCount": len(funcs) - compatibleCount,
	}, nil
}

// Protect executes the protection process
func (e *VMPEngine) Protect(options map[string]interface{}) error {
	e.mu.Lock()
	if e.isBusy {
		e.mu.Unlock()
		return fmt.Errorf("engine is currently busy")
	}
	e.isBusy = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.isBusy = false
		e.mu.Unlock()
	}()

	runtime.EventsEmit(e.ctx, "vmp-log", "[*] 启动 Core Engine...")

	targetFile, ok := options["file"].(string)
	if !ok {
		return fmt.Errorf("invalid file parameter")
	}

	funcsParam, _ := options["functions"].([]interface{})
	opts, ok := options["options"].(map[string]interface{})
	if !ok {
		opts = make(map[string]interface{})
	}

	outPath, _ := opts["outputPath"].(string)
	if outPath == "" {
		outPath = targetFile + ".vmp"
	}

	enableDebug, _ := opts["debug"].(bool)
	stripSymbols, _ := opts["stripSymbols"].(bool)
	tokenEntry, _ := opts["tokenEntry"].(bool)
	verbose := true

	runtime.EventsEmit(e.ctx, "vmp-log", fmt.Sprintf("[+] 目标程序: %s", targetFile))

	var funcs []string
	var addrSpecs []elfpacker.AddrSpec
	for _, rawFn := range funcsParam {
		fMap, ok := rawFn.(map[string]interface{})
		if !ok {
			continue
		}
		isCustom, _ := fMap["isCustom"].(bool)
		if isCustom {
			// 手动添加的函数: 通过地址范围保护
			name, _ := fMap["name"].(string)
			addrStr, _ := fMap["address"].(string)
			sizeFloat, _ := fMap["size"].(float64) // JSON number → float64
			addrStr = strings.TrimPrefix(addrStr, "0x")
			addrStr = strings.TrimPrefix(addrStr, "0X")
			addr, err := strconv.ParseUint(addrStr, 16, 64)
			if err != nil {
				runtime.EventsEmit(e.ctx, "vmp-log", fmt.Sprintf("[!] 地址解析失败: %s — %v", addrStr, err))
				continue
			}
			spec := elfpacker.AddrSpec{
				Addr: addr,
				End:  addr + uint64(sizeFloat),
				Name: name,
			}
			addrSpecs = append(addrSpecs, spec)
			runtime.EventsEmit(e.ctx, "vmp-log", fmt.Sprintf("[+] 手动函数: %s @ 0x%X-0x%X", name, spec.Addr, spec.End))
		} else {
			name, _ := fMap["name"].(string)
			if name != "" {
				funcs = append(funcs, name)
			}
		}
	}

	totalCount := len(funcs) + len(addrSpecs)
	runtime.EventsEmit(e.ctx, "vmp-log", fmt.Sprintf("[+] 开始提取并编译 %d 个目标函数节点 (符号: %d, 地址: %d)...", totalCount, len(funcs), len(addrSpecs)))

	packer := elfpacker.NewPacker(targetFile, outPath, funcs, addrSpecs, verbose, stripSymbols, enableDebug, tokenEntry, interpBlob)

	// Temporarily override os.Stdout/os.Stderr or just let it process.
	// Since NewPacker prints to os.Stdout directly, the user wants logs in the GUI.
	// We'll trust that the quick process will just throw out outputs and we emit main events.

	if err := packer.Process(); err != nil {
		runtime.EventsEmit(e.ctx, "vmp-log", fmt.Sprintf("[x] 保护失败: %v", err))
		return err
	}

	runtime.EventsEmit(e.ctx, "vmp-log", fmt.Sprintf("[*] 初始化完成! 导出位置: %s", outPath))
	return nil
}

// recentFilePath returns the path to the recent files JSON
func (e *VMPEngine) recentFilePath() string {
	return filepath.Join(e.dataDir, "recent_files.json")
}

// GetRecentFiles returns the list of recent files (max 10)
func (e *VMPEngine) GetRecentFiles() []map[string]string {
	data, err := os.ReadFile(e.recentFilePath())
	if err != nil {
		return []map[string]string{}
	}
	var files []map[string]string
	if err := json.Unmarshal(data, &files); err != nil {
		return []map[string]string{}
	}
	return files
}

// AddRecentFile adds a file path to the recent files list
func (e *VMPEngine) AddRecentFile(filePath string) {
	files := e.GetRecentFiles()
	name := filepath.Base(filePath)

	// Remove duplicate
	filtered := make([]map[string]string, 0, len(files))
	for _, f := range files {
		if f["path"] != filePath {
			filtered = append(filtered, f)
		}
	}

	// Prepend new entry
	entry := map[string]string{"name": name, "path": filePath}
	filtered = append([]map[string]string{entry}, filtered...)

	// Cap at 10
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}

	data, _ := json.Marshal(filtered)
	_ = os.WriteFile(e.recentFilePath(), data, 0644)
}

// GetDataDir returns the application data directory
func (e *VMPEngine) GetDataDir() string {
	return e.dataDir
}
