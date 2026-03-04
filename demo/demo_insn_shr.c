#include <stdio.h>

__attribute__((noinline)) int check_shr(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #84\n"
        "mov x2, #1\n"
        "lsr x3, x1, x2\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x2","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_shr();if(v==42){printf("PASS:SHR:%d\\n",v);return 0;}printf("FAIL:SHR:%d\\n",v);return 1;}
