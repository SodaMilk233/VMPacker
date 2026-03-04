#include <stdio.h>

__attribute__((noinline)) int ext_add(int a, int b) {
    return a + b;
}

__attribute__((noinline)) int check_call_nat(int x) {
    int y = ext_add(x, 1);
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return y;
}

int main(void) {
    int v = check_call_nat(41);
    if (v == 42) {
        printf("PASS:CALL_NAT:%d\n", v);
        return 0;
    }
    printf("FAIL:CALL_NAT:%d\n", v);
    return 1;
}

