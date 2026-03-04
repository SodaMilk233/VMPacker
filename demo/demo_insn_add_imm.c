#include <stdio.h>

__attribute__((noinline)) int check_add_imm(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #40\n"
        "add x3, x1, #2\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}
int main(void){int v=check_add_imm();if(v==42){printf("PASS:ADD_IMM:%d\\n",v);return 0;}printf("FAIL:ADD_IMM:%d\\n",v);return 1;}
