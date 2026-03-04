#include <stdio.h>

__attribute__((noinline)) int check_shl_imm(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #21\n"
        "lsl x3, x1, #1\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}
int main(void){int v=check_shl_imm();if(v==42){printf("PASS:SHL_IMM:%d\\n",v);return 0;}printf("FAIL:SHL_IMM:%d\\n",v);return 1;}
