#include <stdio.h>

__attribute__((noinline)) int check_shr_imm(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #84\n"
        "lsr x3, x1, #1\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}
int main(void){int v=check_shr_imm();if(v==42){printf("PASS:SHR_IMM:%d\\n",v);return 0;}printf("FAIL:SHR_IMM:%d\\n",v);return 1;}
