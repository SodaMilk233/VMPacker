#include <stdio.h>

__attribute__((noinline)) int check_not(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #-43\n"
        "mvn x3, x1\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_not();if(v==42){printf("PASS:NOT:%d\\n",v);return 0;}printf("FAIL:NOT:%d\\n",v);return 1;}
