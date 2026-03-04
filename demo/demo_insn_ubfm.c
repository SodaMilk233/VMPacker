#include <stdio.h>
#include <stdint.h>

/*
 * demo_insn_ubfm.c — UBFM width>=32 测试
 *
 * 测试 UBFX 提取位域宽度 >= 32 的情况
 * UBFX Xd, Xn, #lsb, #width  =>  UBFM Xd, Xn, #lsb, #(lsb+width-1)
 *
 * 测试1: UBFX X0, X0, #0, #33  (提取低33位)
 *   UBFM X0, X0, #0, #32  (immr=0, imms=32)
 *   输入: 0x3FFFFFFFF (34位) → 输出: 0x1FFFFFFFF (低33位)
 *
 * 测试2: UBFX X0, X0, #16, #33 (从bit16开始提取33位)
 *   UBFM X0, X0, #16, #48  (immr=16, imms=48)
 *   输入: 0x1FFFFFFFFFFFF (49位) → 输出: 0x1FFFFFFFF (33位)
 */

__attribute__((noinline)) int64_t check_ubfm(void) {
    uint64_t v1 = 0x3FFFFFFFFULL;  /* 低34位全1 */
    uint64_t r1 = 0;

    /* UBFX x0, x0, #0, #33 → 提取低33位 → 0x1FFFFFFFF */
    __asm__ volatile(
        "ubfx %[out], %[in], #0, #33\n"
        : [out] "=r"(r1)
        : [in] "r"(v1)
    );
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");

    uint64_t v2 = 0x1FFFFFFFFFFFFULL; /* 低49位全1 */
    uint64_t r2 = 0;

    /* UBFX x0, x0, #16, #33 → 从bit16提取33位 → 0x1FFFFFFFF */
    __asm__ volatile(
        "ubfx %[out], %[in], #16, #33\n"
        : [out] "=r"(r2)
        : [in] "r"(v2)
    );
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");

    if (r1 == 0x1FFFFFFFFULL && r2 == 0x1FFFFFFFFULL) return 1;
    return 0;
}

int main(void) {
    int64_t v = check_ubfm();
    if (v == 1) { printf("PASS:UBFM\n"); return 0; }
    printf("FAIL:UBFM r=%ld\n", v);
    return 1;
}
