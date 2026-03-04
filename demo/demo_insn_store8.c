#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_store8(void){
    uint8_t buf[8]={0};
    __asm__ volatile(
        "mov w2, #42\n"
        "strb w2, [%[p]]\n"
        :
        : [p] "r"(buf)
        : "w2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (int)buf[0];
}

int main(void){int v=check_store8();if(v==42){printf("PASS:STORE8:%d\\n",v);return 0;}printf("FAIL:STORE8:%d\\n",v);return 1;}
