#include <stdio.h>

__attribute__((noinline)) int check_sub(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #44\n"
        "mov x2, #2\n"
        "sub x3, x1, x2\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x2","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_sub();if(v==42){printf("PASS:SUB:%d\\n",v);return 0;}printf("FAIL:SUB:%d\\n",v);return 1;}
