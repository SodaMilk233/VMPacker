#include <stdio.h>
#include <stdint.h>

/*
 * demo_insn_ldrsb.c — LDRSB 指令测试
 *
 * LDRSB Xd, [Xn, #imm] : 从内存加载 8 位有符号值，符号扩展到 64 位
 * 测试正数和负数两种情况
 */

__attribute__((noinline)) int64_t check_ldrsb(void) {
    /* 正数: 0x7F = 127 → 符号扩展后仍为 0x000000000000007F */
    /* 负数: 0x80 = -128 (i8) → 符号扩展后为 0xFFFFFFFFFFFFFF80 */
    int8_t buf[2] = { -128, 127 };
    int64_t r1 = 0, r2 = 0;

    __asm__ volatile(
        "ldrsb %[o1], [%[p]]\n"       /* buf[0] = -128 → sign-extend to i64 */
        "ldrsb %[o2], [%[p], #1]\n"   /* buf[1] = 127 → sign-extend */
        : [o1] "=r"(r1), [o2] "=r"(r2)
        : [p] "r"(buf)
        : "memory"
    );
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");

    /* r1 应该是 -128 (0xFFFFFFFFFFFFFF80) */
    /* r2 应该是 127 (0x000000000000007F) */
    if (r1 == -128 && r2 == 127) {
        return 1; /* PASS */
    }
    return 0; /* FAIL */
}

int main(void) {
    int64_t v = check_ldrsb();
    if (v == 1) {
        printf("PASS:LDRSB\n");
        return 0;
    }
    printf("FAIL:LDRSB r=%ld\n", v);
    return 1;
}
