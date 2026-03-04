#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_store64(void){
    uint64_t buf[2]={0};
    __asm__ volatile(
        "mov x2, #42\n"
        "str x2, [%[p]]\n"
        :
        : [p] "r"(buf)
        : "x2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (int)buf[0];
}

int main(void){int v=check_store64();if(v==42){printf("PASS:STORE64:%d\\n",v);return 0;}printf("FAIL:STORE64:%d\\n",v);return 1;}
