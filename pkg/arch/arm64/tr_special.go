package arm64

import (
	"fmt"

	"github.com/vmpacker/pkg/vm"
)

// ============================================================
// 特殊指令翻译 — ADRP / ADR
// ============================================================

func (t *Translator) trADRP(instructions []vm.Instruction, idx int) (int, error) {
	inst := instructions[idx]
	rd, err := t.mapReg(inst.Rd)
	if err != nil {
		return 0, err
	}

	pc, ok := addSignedOffset(t.funcAddr, int64(inst.Offset))
	if !ok {
		return 0, fmt.Errorf("ADRP PC 地址溢出")
	}
	pageBase := pc &^ 0xFFF
	adrpResult, ok := addSignedOffset(pageBase, inst.Imm)
	if !ok {
		return 0, fmt.Errorf("ADRP 目标地址溢出")
	}

	if idx+1 < len(instructions) {
		next := instructions[idx+1]
		if Op(next.Op) == ADD_IMM && next.Rd == inst.Rd && next.Rn == inst.Rd {
			finalAddr, ok := addSignedOffset(adrpResult, next.Imm)
			if !ok {
				return 0, fmt.Errorf("ADRP+ADD 目标地址溢出")
			}
			t.emit(vm.OpMovImage, rd)
			t.emitU64(finalAddr)
			return 1, nil
		}
	}

	t.emit(vm.OpMovImage, rd)
	t.emitU64(adrpResult)
	return 0, nil
}

func (t *Translator) trADR(inst vm.Instruction) (int, error) {
	rd, err := t.mapReg(inst.Rd)
	if err != nil {
		return 0, err
	}
	pc, ok := addSignedOffset(t.funcAddr, int64(inst.Offset))
	if !ok {
		return 0, fmt.Errorf("ADR PC 地址溢出")
	}
	addr, ok := addSignedOffset(pc, inst.Imm)
	if !ok {
		return 0, fmt.Errorf("ADR 目标地址溢出")
	}
	t.emit(vm.OpMovImage, rd)
	t.emitU64(addr)
	return 0, nil
}

// trSVC 翻译 SVC #imm16
// 字节码: [OpSvc][imm16_lo][imm16_hi] = 3B
// handler 使用 inline asm 执行 svc #0，从 VM 寄存器传递 syscall 参数
func (t *Translator) trSVC(inst vm.Instruction) error {
	imm16 := uint16(inst.Imm)
	t.emit(vm.OpSvc, byte(imm16), byte(imm16>>8))
	return nil
}
