package arm64

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/vmpacker/pkg/vm"
)

func TestTranslateBranchLocal(t *testing.T) {
	translator := NewTranslator(0x1000, 12)
	result, err := translator.Translate([]vm.Instruction{
		{Op: int(B), Offset: 0, Imm: 8},
		{Op: int(NOP), Offset: 4},
		{Op: int(RET), Offset: 8},
	})
	if err != nil {
		t.Fatalf("Translate() error: %v", err)
	}
	if len(result.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported instructions: %v", result.Unsupported)
	}
	if result.Bytecode[0] != vm.OpJmp {
		t.Fatalf("opcode = 0x%X, want JMP", result.Bytecode[0])
	}
	if target := binary.LittleEndian.Uint32(result.Bytecode[1:5]); target != 6 {
		t.Fatalf("VM branch target = %d, want 6", target)
	}
}

func TestTranslateBranchExternalKnownEntryAsTailCall(t *testing.T) {
	translator := NewTranslator(0x1000, 4)
	translator.SetExternalBranchTargets([]uint64{0x1100})
	result, err := translator.Translate([]vm.Instruction{
		{Op: int(B), Offset: 0, Imm: 0x100},
	})
	if err != nil {
		t.Fatalf("Translate() error: %v", err)
	}
	if len(result.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported instructions: %v", result.Unsupported)
	}
	if result.Bytecode[0] != vm.OpCallNative {
		t.Fatalf("opcode = 0x%X, want CALL_NATIVE", result.Bytecode[0])
	}
	if target := binary.LittleEndian.Uint64(result.Bytecode[1:9]); target != 0x1100 {
		t.Fatalf("native target = 0x%X, want 0x1100", target)
	}
	if result.Bytecode[9] != vm.OpRet || result.Bytecode[10] != 0 {
		t.Fatalf("tail call is not followed by RET R0: % X", result.Bytecode[:11])
	}
}

func TestTranslateBranchExternalUnknownOrMidFunctionRejected(t *testing.T) {
	translator := NewTranslator(0x1000, 4)
	translator.SetExternalBranchTargets([]uint64{0x1100})
	result, err := translator.Translate([]vm.Instruction{
		{Op: int(B), Offset: 0, Imm: 0x104},
	})
	if err != nil {
		t.Fatalf("Translate() error: %v", err)
	}
	if len(result.Unsupported) != 1 {
		t.Fatalf("unsupported count = %d, want 1", len(result.Unsupported))
	}
	if !strings.Contains(result.Unsupported[0], "0x1104") ||
		!strings.Contains(result.Unsupported[0], "不是已知函数入口") {
		t.Fatalf("unexpected diagnostic: %s", result.Unsupported[0])
	}
}

func TestAddSignedOffset(t *testing.T) {
	tests := []struct {
		name   string
		base   uint64
		offset int64
		want   uint64
		ok     bool
	}{
		{name: "positive", base: 0x1000, offset: 0x20, want: 0x1020, ok: true},
		{name: "negative", base: 0x1000, offset: -0x20, want: 0xFE0, ok: true},
		{name: "underflow", base: 0, offset: -1, ok: false},
		{name: "overflow", base: ^uint64(0), offset: 1, ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := addSignedOffset(test.base, test.offset)
			if got != test.want || ok != test.ok {
				t.Fatalf("addSignedOffset() = (0x%X, %v), want (0x%X, %v)", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestTranslateImageRelativeAddresses(t *testing.T) {
	t.Run("ADR", func(t *testing.T) {
		translator := NewTranslator(0x4000, 4)
		result, err := translator.Translate([]vm.Instruction{
			{Op: int(ADR), Offset: 0, Imm: 0x20, Rd: 3},
		})
		if err != nil || len(result.Unsupported) != 0 {
			t.Fatalf("Translate() = (%v, %v)", err, result.Unsupported)
		}
		if result.Bytecode[0] != vm.OpMovImage || result.Bytecode[1] != 3 {
			t.Fatalf("ADR bytecode prefix = % X", result.Bytecode[:2])
		}
		if target := binary.LittleEndian.Uint64(result.Bytecode[2:10]); target != 0x4020 {
			t.Fatalf("ADR image VA = 0x%X, want 0x4020", target)
		}
	})

	t.Run("LDR literal", func(t *testing.T) {
		translator := NewTranslator(0x4000, 8)
		result, err := translator.Translate([]vm.Instruction{
			{Op: int(LDR_LIT), Offset: 4, Imm: -4, Rd: 0, SF: true},
		})
		if err != nil || len(result.Unsupported) != 0 {
			t.Fatalf("Translate() = (%v, %v)", err, result.Unsupported)
		}
		if result.Bytecode[0] != vm.OpSPushImage {
			t.Fatalf("LDR literal opcode = 0x%X, want S_PUSH_IMAGE", result.Bytecode[0])
		}
		if target := binary.LittleEndian.Uint64(result.Bytecode[1:9]); target != 0x4000 {
			t.Fatalf("LDR literal image VA = 0x%X, want 0x4000", target)
		}
	})
}

func TestTranslateSBFMGeneralForms(t *testing.T) {
	tests := []struct {
		name  string
		sf    bool
		immr  int64
		imms  int
		input uint64
		want  uint64
	}{
		{name: "SBFIZ signed word shifted", sf: true, immr: 58, imms: 31, input: 0xFFFFFFFF, want: 0xFFFFFFFFFFFFFFC0},
		{name: "SBFX byte", sf: true, immr: 8, imms: 15, input: 0x8000, want: 0xFFFFFFFFFFFFFF80},
		{name: "ASR alias", sf: true, immr: 4, imms: 63, input: 0xFFFFFFFFFFFFFFF0, want: 0xFFFFFFFFFFFFFFFF},
		{name: "SXTB alias", sf: true, immr: 0, imms: 7, input: 0x80, want: 0xFFFFFFFFFFFFFF80},
		{name: "32-bit truncation", sf: false, immr: 0, imms: 7, input: 0x80, want: 0xFFFFFF80},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			translator := NewTranslator(0x1000, 4)
			result, err := translator.Translate([]vm.Instruction{{
				Op: int(SBFM), Offset: 0, Rd: 0, Rn: 1,
				Imm: test.immr, Shift: test.imms, SF: test.sf,
			}})
			if err != nil || len(result.Unsupported) != 0 {
				t.Fatalf("Translate() = (%v, %v)", err, result.Unsupported)
			}
			registers := [32]uint64{1: test.input}
			if got := runStackTestBytecode(t, result.Bytecode[:result.CodeLen], &registers); got != test.want {
				t.Fatalf("result = 0x%X, want 0x%X", got, test.want)
			}
		})
	}
}

func runStackTestBytecode(t *testing.T, code []byte, registers *[32]uint64) uint64 {
	t.Helper()
	stack := make([]uint64, 0, 8)
	pop := func() uint64 {
		value := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return value
	}
	for pc := 0; pc < len(code); {
		switch code[pc] {
		case vm.OpSVload:
			stack = append(stack, registers[code[pc+1]&31])
			pc += 2
		case vm.OpSVstore:
			registers[code[pc+1]&31] = pop()
			pc += 2
		case vm.OpSPushImm32:
			stack = append(stack, uint64(binary.LittleEndian.Uint32(code[pc+1:])))
			pc += 5
		case vm.OpSPushImm64:
			stack = append(stack, binary.LittleEndian.Uint64(code[pc+1:]))
			pc += 9
		case vm.OpSAnd:
			b, a := pop(), pop()
			stack = append(stack, a&b)
			pc++
		case vm.OpSShr:
			b, a := pop(), pop()
			stack = append(stack, a>>(b&63))
			pc++
		case vm.OpSShl:
			b, a := pop(), pop()
			stack = append(stack, a<<(b&63))
			pc++
		case vm.OpSAsr:
			b, a := pop(), pop()
			stack = append(stack, uint64(int64(a)>>(b&63)))
			pc++
		case vm.OpSTrunc32:
			stack = append(stack, pop()&0xFFFFFFFF)
			pc++
		case vm.OpHalt:
			return registers[0]
		default:
			t.Fatalf("unexpected opcode 0x%X at %d", code[pc], pc)
		}
	}
	return registers[0]
}
