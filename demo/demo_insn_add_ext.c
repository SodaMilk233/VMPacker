/*
 * demo_insn_add_ext.c — ADD/SUB extended register 测试
 *
 * T4: 纯组合路径，无需新增 VM opcode
 *
 * 覆盖编码变体:
 *   0x8B21C262 = ADD X2, X19, W1, SXTW
 *     sf=1, op=0, S=0, 01011_00_1, Rm=00001, option=110(SXTW), imm3=000, Rn=10011, Rd=00010
 *
 *   0x8B3C0333 = ADD X19, X25, W28, UXTB
 *     sf=1, op=0, S=0, 01011_00_1, Rm=11100, option=000(UXTB), imm3=000, Rn=11001, Rd=10011
 *
 *   0xCB3063FF = SUB SP, SP, X16, UXTX
 *     sf=1, op=1, S=0, 01011_00_1, Rm=10000, option=011(UXTX), imm3=000, Rn=11111, Rd=11111
 *
 *   0x8B3063FF = ADD SP, SP, X16, UXTX
 *     sf=1, op=0, S=0, 01011_00_1, Rm=10000, option=011(UXTX), imm3=000, Rn=11111, Rd=11111
 *
 * 交叉编译:
 *   aarch64-linux-gnu-gcc -O1 -static -o build/demo_insn_add_ext demo/demo_insn_add_ext.c
 */
#include <stdio.h>
#include <stdint.h>

__attribute__((noinline))
int test_add_ext(uint64_t a, uint64_t b, uint64_t c, uint64_t d) {
    /*
     * ABI: a=X0, b=X1, c=X2, d=X3
     * 使用 X9-X16, X19, X25, X28 作为工作寄存器
     *
     * Test 1: ADD X2, X19, W1, SXTW  (0x8B21C262)
     *   X19 = a, W1 = (int32_t)b → X2 = a + sign_extend(b[31:0])
     *
     * Test 2: ADD X19, X25, W28, UXTB (0x8B3C0333)
     *   X25 = c, W28 = d → X19 = c + zero_extend(d[7:0])
     *
     * Test 3+4: SUB SP, SP, X16, UXTX (0xCB3063FF)
     *           ADD SP, SP, X16, UXTX (0x8B3063FF)
     *   先 SUB 再 ADD 同一值，SP 应恢复原值
     */
    uint64_t r1, r2, sp_before, sp_after;

    __asm__ volatile(
        /* ---- Test 1: ADD X2, X19, W1, SXTW ---- */
        /* 设置: X19 = a, X1 = b */
        "mov x19, %[a]\n\t"
        "mov x1,  %[b]\n\t"
        ".inst 0x8B21C262\n\t"       /* ADD X2, X19, W1, SXTW */
        "mov %[r1], x2\n\t"

        /* ---- Test 2: ADD X19, X25, W28, UXTB ---- */
        /* 设置: X25 = c, W28 = d */
        "mov x25, %[c]\n\t"
        "mov x28, %[d]\n\t"
        ".inst 0x8B3C0333\n\t"       /* ADD X19, X25, W28, UXTB */
        "mov %[r2], x19\n\t"

        /* ---- Test 3+4: SUB/ADD SP with UXTX ---- */
        /* 保存 SP，用 X16=64 做 SUB 再 ADD 恢复 */
        "mov %[spb], sp\n\t"
        "mov x16, #64\n\t"
        ".inst 0xCB3063FF\n\t"       /* SUB SP, SP, X16, UXTX */
        ".inst 0x8B3063FF\n\t"       /* ADD SP, SP, X16, UXTX */
        "mov %[spa], sp\n\t"

        /* padding NOPs to ensure function > 72 bytes */
        "nop\n\t" "nop\n\t" "nop\n\t" "nop\n\t"
        "nop\n\t" "nop\n\t" "nop\n\t" "nop\n\t"
        "nop\n\t" "nop\n\t" "nop\n\t" "nop\n\t"

        : [r1] "=r"(r1), [r2] "=r"(r2),
          [spb] "=r"(sp_before), [spa] "=r"(sp_after)
        : [a] "r"(a), [b] "r"(b), [c] "r"(c), [d] "r"(d)
        : "x1", "x2", "x16", "x19", "x25", "x28", "memory"
    );

    /*
     * 验证:
     * Test 1: ADD X2, X19, W1, SXTW
     *   r1 = a + (int64_t)(int32_t)(b & 0xFFFFFFFF)
     *
     * Test 2: ADD X19, X25, W28, UXTB
     *   r2 = c + (d & 0xFF)
     *
     * Test 3+4: SP 应恢复
     *   sp_before == sp_after
     */
    int64_t sext_b = (int64_t)(int32_t)(uint32_t)(b & 0xFFFFFFFF);
    uint64_t exp1 = a + (uint64_t)sext_b;
    uint64_t exp2 = c + (d & 0xFF);

    if (r1 != exp1) {
        printf("ADD_EXT FAIL test1: ADD SXTW got=0x%lX expected=0x%lX\n",
               (unsigned long)r1, (unsigned long)exp1);
        return 1;
    }
    if (r2 != exp2) {
        printf("ADD_EXT FAIL test2: ADD UXTB got=0x%lX expected=0x%lX\n",
               (unsigned long)r2, (unsigned long)exp2);
        return 2;
    }
    if (sp_before != sp_after) {
        printf("ADD_EXT FAIL test3: SUB/ADD UXTX SP mismatch before=0x%lX after=0x%lX\n",
               (unsigned long)sp_before, (unsigned long)sp_after);
        return 3;
    }

    return 0;
}

int main(void) {
    /*
     * 测试值:
     *   a = 0x0000000100000000 (4GB)
     *   b = 0x00000000FFFFFFFE (W 部分 = 0xFFFFFFFE = -2 signed)
     *   c = 0x0000001000000000
     *   d = 0x00000000000000AB (UXTB 取低 8 位 = 0xAB)
     */
    uint64_t a = 0x0000000100000000ULL;
    uint64_t b = 0x00000000FFFFFFFEULL;  /* SXTW → -2 → 0xFFFFFFFFFFFFFFFE */
    uint64_t c = 0x0000001000000000ULL;
    uint64_t d = 0x12345678000000ABULL;  /* UXTB → 0xAB */

    int ret = test_add_ext(a, b, c, d);
    if (ret == 0) {
        printf("ADD_EXT PASS\n");
        return 0;
    }
    return ret;
}
