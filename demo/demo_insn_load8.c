#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_load8(void){
    uint8_t buf[8]={42,0,0,0,0,0,0,0};
    int out=-1;
    __asm__ volatile(
        "ldrb w2, [%[p]]\n"
        "mov %w[o], w2\n"
        : [o] "=r"(out)
        : [p] "r"(buf)
        : "w2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_load8();if(v==42){printf("PASS:LOAD8:%d\\n",v);return 0;}printf("FAIL:LOAD8:%d\\n",v);return 1;}
