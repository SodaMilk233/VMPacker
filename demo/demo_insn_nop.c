#include <stdio.h>

__attribute__((noinline)) int check_nop(int x) {
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return x + 1;
}

int main(void) {
    int v = check_nop(41);
    if (v == 42) {
        printf("PASS:NOP:%d\n", v);
        return 0;
    }
    printf("FAIL:NOP:%d\n", v);
    return 1;
}

