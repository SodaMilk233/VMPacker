#include <stdio.h>

/*
 * 目标：触发 ARM64 BR(寄存器间接跳转)，对应 VM 的 OP_BR_REG。
 * 方法：GNU C computed goto（goto *ptr）。
 */
__attribute__((noinline)) int check_br_reg(int x) {
    static void *tbl[] = { &&L_FAIL, &&L_OK };
    void *target = tbl[(x == 41) ? 1 : 0];
    goto *target;

L_FAIL:
    return -1;

L_OK:
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return 42;
}

int main(void) {
    int v = check_br_reg(41);
    if (v == 42) {
        printf("PASS:BR_REG:%d\n", v);
        return 0;
    }
    printf("FAIL:BR_REG:%d\n", v);
    return 1;
}

