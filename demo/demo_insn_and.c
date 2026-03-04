#include <stdio.h>

__attribute__((noinline)) int check_and(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #42\n"
        "mov x2, #255\n"
        "and x3, x1, x2\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x2","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_and();if(v==42){printf("PASS:AND:%d\\n",v);return 0;}printf("FAIL:AND:%d\\n",v);return 1;}
