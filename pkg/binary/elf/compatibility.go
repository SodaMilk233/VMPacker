package elf

import (
	"debug/elf"
	"fmt"

	"github.com/vmpacker/pkg/arch/arm64"
	"github.com/vmpacker/pkg/vm"
)

// FunctionCompatibility is a fail-closed translation preflight result.
type FunctionCompatibility struct {
	Compatible       bool
	Reason           string
	UnsupportedCount int
	TotalInsts       int
	TranslatedInsts  int
}

// CheckFunctionCompatibility decodes and translates a function without
// modifying the ELF. It is used by the GUI to disable unsupported candidates.
func (p *Packer) CheckFunctionCompatibility(f *elf.File, fi *vm.FuncInfo, branchTargets []uint64) FunctionCompatibility {
	code, err := p.ExtractFuncCode(f, fi)
	if err != nil {
		return FunctionCompatibility{Reason: err.Error()}
	}
	insts := p.DecodeFunction(code)
	translator := arm64.NewTranslator(fi.Addr, int(fi.Size))
	translator.SetExternalBranchTargets(branchTargets)
	result, err := translator.Translate(insts)
	if err != nil {
		return FunctionCompatibility{
			Reason:     fmt.Sprintf("translation failed: %v", err),
			TotalInsts: len(insts),
		}
	}
	compatibility := FunctionCompatibility{
		Compatible:       len(result.Unsupported) == 0,
		UnsupportedCount: len(result.Unsupported),
		TotalInsts:       result.TotalInsts,
		TranslatedInsts:  result.TransInsts,
	}
	if len(result.Unsupported) > 0 {
		compatibility.Reason = result.Unsupported[0]
		if len(result.Unsupported) > 1 {
			compatibility.Reason += fmt.Sprintf("（另有 %d 条）", len(result.Unsupported)-1)
		}
	}
	return compatibility
}
