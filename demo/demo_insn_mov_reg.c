#include <stdio.h>

__attribute__((noinline)) int check_mov_reg(int x) {
    unsigned long out = 0;
    __asm__ volatile(
        "mov x1, %x[i]\n"
        "mov x0, x1\n"
        "mov %x[o], x0\n"
        : [o] "=r" (out)
        : [i] "r" ((unsigned long)x)
        : "x0", "x1", "cc", "memory");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (out == (unsigned long)x) ? 42 : -1;
}

int main(void){int v=check_mov_reg(42);if(v==42){printf("PASS:MOV_REG:%d\n",v);return 0;}printf("FAIL:MOV_REG:%d\n",v);return 1;}

