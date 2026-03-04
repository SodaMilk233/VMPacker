#include <stdio.h>

__attribute__((noinline)) int check_xor(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #104\n"
        "mov x2, #66\n"
        "eor x3, x1, x2\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x2","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_xor();if(v==42){printf("PASS:XOR:%d\\n",v);return 0;}printf("FAIL:XOR:%d\\n",v);return 1;}
