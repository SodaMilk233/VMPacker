#include <stdio.h>

__attribute__((noinline)) int check_jmp(int x) {
    int out = -1;
    __asm__ volatile(
        "mov w0, #0\n"
        "b 1f\n"
        "mov w0, #7\n"
        "1:\n"
        "mov w0, #42\n"
        "mov %w[o], w0\n"
        : [o] "=r" (out)
        :
        : "w0", "cc", "memory");
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return out + (x - x);
}

int main(void) {
    int v = check_jmp(41);
    if (v == 42) {
        printf("PASS:JMP:%d\n", v);
        return 0;
    }
    printf("FAIL:JMP:%d\n", v);
    return 1;
}

