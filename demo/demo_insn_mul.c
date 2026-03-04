#include <stdio.h>

__attribute__((noinline)) int check_mul(void){
    int out=-1;
    __asm__ volatile(
        "mov x1, #21\n"
        "mov x2, #2\n"
        "mul x3, x1, x2\n"
        "mov %w[o], w3\n"
        : [o] "=r"(out)
        :
        : "x1","x2","x3","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_mul();if(v==42){printf("PASS:MUL:%d\\n",v);return 0;}printf("FAIL:MUL:%d\\n",v);return 1;}
