#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_load32(void){
    uint32_t buf[2]={42,0};
    int out=-1;
    __asm__ volatile(
        "ldr w2, [%[p]]\n"
        "mov %w[o], w2\n"
        : [o] "=r"(out)
        : [p] "r"(buf)
        : "w2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return out;
}

int main(void){int v=check_load32();if(v==42){printf("PASS:LOAD32:%d\\n",v);return 0;}printf("FAIL:LOAD32:%d\\n",v);return 1;}
