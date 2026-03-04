#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_load64(void){
    uint64_t buf[2]={42,0};
    long out=-1;
    __asm__ volatile(
        "ldr x2, [%[p]]\n"
        "mov %[o], x2\n"
        : [o] "=r"(out)
        : [p] "r"(buf)
        : "x2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (int)out;
}

int main(void){int v=check_load64();if(v==42){printf("PASS:LOAD64:%d\\n",v);return 0;}printf("FAIL:LOAD64:%d\\n",v);return 1;}
