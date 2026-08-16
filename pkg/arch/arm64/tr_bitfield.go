package arm64

import (
	"github.com/vmpacker/pkg/vm"
)

// ============================================================
// 位域翻译 — SBFM (安全，无 temp 寄存器冲突)
// UBFM 已迁移到 tr_stack.go (trStackUBFM)
// ============================================================

func (t *Translator) trSBFM(inst vm.Instruction) error {
	rd, err := t.mapReg(inst.Rd)
	if err != nil {
		return err
	}
	rn, err := t.mapReg(inst.Rn)
	if err != nil {
		return err
	}
	immr := uint32(inst.Imm)
	imms := uint32(inst.Shift)

	regSize := uint32(32)
	if inst.SF {
		regSize = 64
	}

	// Work in the stack VM so Rd may safely alias Rn. W-register sources are
	// explicitly truncated because every VM register is physically 64-bit.
	t.sVload(rn)
	if !inst.SF {
		t.sPushImm32(0xFFFFFFFF)
		t.emit(vm.OpSAnd)
	}

	var width uint32
	var leftShift uint32
	if imms >= immr {
		// SBFX/ASR: extract [imms:immr], then sign-extend that field.
		width = imms - immr + 1
		if immr > 0 {
			t.sPushImm32(immr)
			t.emit(vm.OpSShr)
		}
	} else {
		// SBFIZ: sign-extend the low field and insert it at regSize-immr.
		width = imms + 1
		leftShift = regSize - immr
		if width < 64 {
			t.sPushImm(bitMask(width))
			t.emit(vm.OpSAnd)
		}
	}

	if width < 64 {
		shift := uint32(64) - width
		t.sPushImm32(shift)
		t.emit(vm.OpSShl)
		t.sPushImm32(shift)
		t.emit(vm.OpSAsr)
	}
	if leftShift > 0 {
		t.sPushImm32(leftShift)
		t.emit(vm.OpSShl)
	}
	if !inst.SF {
		t.emit(vm.OpSTrunc32)
	}
	t.sVstore(rd)
	return nil
}

func bitMask(width uint32) uint64 {
	if width >= 64 {
		return ^uint64(0)
	}
	return (uint64(1) << width) - 1
}
