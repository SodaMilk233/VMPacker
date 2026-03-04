#include <stdio.h>

__attribute__((noinline)) int check_push(void){
    long out=-1;
    __asm__ volatile(
        "mov x1, #42\n"
        "stp x1, xzr, [sp, #-16]!\n"
        "ldp x2, xzr, [sp], #16\n"
        "mov %[o], x2\n"
        : [o] "=r"(out)
        :
        : "x1","x2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (int)out;
}
int main(void){int v=check_push();if(v==42){printf("PASS:PUSH:%d\\n",v);return 0;}printf("FAIL:PUSH:%d\\n",v);return 1;}
