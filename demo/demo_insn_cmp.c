#include <stdio.h>

__attribute__((noinline)) int check_cmp(int a, int b) {
    int out = -1;
    __asm__ volatile(
        "mov w2, #0\n"
        "cmp %w[x], %w[y]\n"
        "b.ne 1f\n"
        "mov w2, #42\n"
        "1:\n"
        "mov %w[o], w2\n"
        : [o] "=r" (out)
        : [x] "r" (a), [y] "r" (b)
        : "w2", "cc", "memory");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_cmp(41,41);if(v==42){printf("PASS:CMP:%d\n",v);return 0;}printf("FAIL:CMP:%d\n",v);return 1;}

