#include <stdio.h>

__attribute__((noinline)) int check_jle(int x) {
    int out = -1;
    __asm__ volatile(
        "mov w0, #7\n"
        "mov w1, %w[i]\n"
        "cmp w1, #1\n"
        "b.le 1f\n"
        "mov w0, #13\n"
        "1:\n"
        "mov w0, #42\n"
        "mov %w[o], w0\n"
        : [o] "=r" (out)
        : [i] "r" (x)
        : "w0", "w1", "cc", "memory");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_jle(1); if(v==42){printf("PASS:JLE:%d\n",v); return 0;} printf("FAIL:JLE:%d\n",v); return 1;}

