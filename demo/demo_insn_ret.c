#include <stdio.h>

__attribute__((noinline)) int check_ret(int x) {
    int y = x + 1;
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return y;
}

int main(void) {
    int v = check_ret(41);
    if (v == 42) {
        printf("PASS:RET:%d\n", v);
        return 0;
    }
    printf("FAIL:RET:%d\n", v);
    return 1;
}

