#include <stdio.h>

__attribute__((noinline)) int check_mov_imm(void) {
    unsigned long out = 0;
    __asm__ volatile(
        "mov x0, #42\n"
        "mov %x[o], x0\n"
        : [o] "=r" (out)
        :
        : "x0", "cc", "memory");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (out == 42UL) ? 42 : -1;
}

int main(void){int v=check_mov_imm();if(v==42){printf("PASS:MOV_IMM:%d\n",v);return 0;}printf("FAIL:MOV_IMM:%d\n",v);return 1;}

