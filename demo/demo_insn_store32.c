#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_store32(void){
    uint32_t buf[2]={0};
    __asm__ volatile(
        "mov w2, #42\n"
        "str w2, [%[p]]\n"
        :
        : [p] "r"(buf)
        : "w2","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (int)buf[0];
}

int main(void){int v=check_store32();if(v==42){printf("PASS:STORE32:%d\\n",v);return 0;}printf("FAIL:STORE32:%d\\n",v);return 1;}
