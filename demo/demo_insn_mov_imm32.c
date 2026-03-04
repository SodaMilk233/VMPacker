#include <stdio.h>

__attribute__((noinline)) int check_mov_imm32(void) {
    unsigned int out = 0;
    __asm__ volatile(
        "mov w0, #42\n"
        "mov %w[o], w0\n"
        : [o] "=r" (out)
        :
        : "w0", "cc", "memory");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (out == 42U) ? 42 : -1;
}

int main(void){int v=check_mov_imm32();if(v==42){printf("PASS:MOV_IMM32:%d\n",v);return 0;}printf("FAIL:MOV_IMM32:%d\n",v);return 1;}

