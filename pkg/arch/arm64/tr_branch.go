package arm64

import (
	"fmt"

	"github.com/vmpacker/pkg/vm"
)

// ============================================================
// 分支翻译 — B / B.cond / BL / BLR / BR / TBZ
// CSEL/CBZ 已迁移到 tr_stack.go (trStackCSEL/trStackCBZ)
// ============================================================

func (t *Translator) trBranch(inst vm.Instruction) error {
	target := int64(inst.Offset) + inst.Imm

	if target < 0 || target > int64(t.funcSize) {
		absoluteTarget, ok := addSignedOffset(t.funcAddr, target)
		if !ok {
			return fmt.Errorf("分支目标地址溢出: 函数 0x%X + 偏移 %d", t.funcAddr, target)
		}
		if _, allowed := t.externalBranchTargets[absoluteTarget]; !allowed {
			return fmt.Errorf("分支目标 0x%X 超出函数范围 [0x%X, 0x%X]，且不是已知函数入口",
				absoluteTarget, t.funcAddr, t.funcAddr+uint64(t.funcSize))
		}

		// AArch64 的外部 B 是尾调用：目标返回时应直接返回当前函数的调用者。
		t.emit(vm.OpCallNative)
		t.emitU64(absoluteTarget)
		t.emit(vm.OpRet, 0)
		return nil
	}

	t.emit(vm.OpJmp)
	fixPos := t.pos()
	t.emitU32(0)
	t.fixups = append(t.fixups, branchFixup{vmOffset: fixPos, arm64Target: int(target)})
	return nil
}

func addSignedOffset(base uint64, offset int64) (uint64, bool) {
	if offset >= 0 {
		delta := uint64(offset)
		if delta > ^uint64(0)-base {
			return 0, false
		}
		return base + delta, true
	}
	delta := uint64(-(offset + 1)) + 1
	if delta > base {
		return 0, false
	}
	return base - delta, true
}

func (t *Translator) trBranchCond(inst vm.Instruction) error {
	target := inst.Offset + int(inst.Imm)

	if target < 0 || target > t.funcSize {
		return fmt.Errorf("条件分支目标 0x%X 超出函数范围 [0, 0x%X]", target, t.funcSize)
	}

	var vmOp byte
	switch inst.Cond {
	case COND_EQ:
		vmOp = vm.OpJe
	case COND_NE:
		vmOp = vm.OpJne
	case COND_LT:
		vmOp = vm.OpJl
	case COND_GE:
		vmOp = vm.OpJge
	case COND_GT:
		vmOp = vm.OpJgt
	case COND_LE:
		vmOp = vm.OpJle
	case COND_CS:
		vmOp = vm.OpJae
	case COND_CC:
		vmOp = vm.OpJb
	case COND_HI:
		vmOp = vm.OpJa
	case COND_LS:
		vmOp = vm.OpJbe
	case COND_MI:
		vmOp = vm.OpJl // MI: N==1 → FL_SIGN set
	case COND_PL:
		vmOp = vm.OpJge // PL: N==0 → FL_SIGN not set
	default:
		return fmt.Errorf("不支持的条件码 0x%X", inst.Cond)
	}

	t.emit(vmOp)
	fixPos := t.pos()
	t.emitU32(0)
	t.fixups = append(t.fixups, branchFixup{vmOffset: fixPos, arm64Target: target})
	return nil
}

func (t *Translator) trBL(inst vm.Instruction) error {
	target := uint64(int64(t.funcAddr) + int64(inst.Offset) + inst.Imm)

	t.emit(vm.OpCallNative)
	t.emitU64(target)
	return nil
}

func (t *Translator) trBLR(inst vm.Instruction) error {
	rn, err := t.mapReg(inst.Rn)
	if err != nil {
		return err
	}
	t.emit(vm.OpCallReg, rn)
	return nil
}

func (t *Translator) trBR(inst vm.Instruction) error {
	rn, err := t.mapReg(inst.Rn)
	if err != nil {
		return err
	}
	t.emit(vm.OpBrReg, rn)
	return nil
}

// trTBZ 翻译 TBZ/TBNZ — test bit and branch
// 字节码: [OpTbz/OpTbnz][reg][bit][target32] = 7B
// inst.Shift = bit number (b5:b40), inst.Imm = offset (已乘4)
func (t *Translator) trTBZ(inst vm.Instruction, isZero bool) error {
	target := inst.Offset + int(inst.Imm)

	if target < 0 || target > t.funcSize {
		return fmt.Errorf("TBZ/TBNZ 分支目标 0x%X 超出函数范围 [0, 0x%X)", target, t.funcSize)
	}

	rd, err := t.mapReg(inst.Rd)
	if err != nil {
		return err
	}

	var vmOp byte
	if isZero {
		vmOp = vm.OpTbz
	} else {
		vmOp = vm.OpTbnz
	}

	t.emit(vmOp, rd, byte(inst.Shift))
	fixPos := t.pos()
	t.emitU32(0)
	t.fixups = append(t.fixups, branchFixup{vmOffset: fixPos, arm64Target: target})
	return nil
}
