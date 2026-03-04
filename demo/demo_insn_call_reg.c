#include <stdio.h>

typedef int (*fn_add1_t)(int);

__attribute__((noinline)) int ext_add1(int x) {
    return x + 1;
}

__attribute__((noinline)) int check_call_reg(int x) {
    fn_add1_t fp = ext_add1;
    int y = fp(x);
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return y;
}

int main(void) {
    int v = check_call_reg(41);
    if (v == 42) {
        printf("PASS:CALL_REG:%d\n", v);
        return 0;
    }
    printf("FAIL:CALL_REG:%d\n", v);
    return 1;
}

