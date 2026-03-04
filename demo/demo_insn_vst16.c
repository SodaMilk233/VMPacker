#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int check_vst16(void){
    uint8_t in[16]={42,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15};
    uint8_t out[16]={0};
    __asm__ volatile(
        "ld1 {v0.16b}, [%[src]]\n"
        "st1 {v0.16b}, [%[dst]]\n"
        :
        : [src] "r"(in), [dst] "r"(out)
        : "v0","memory","cc");
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");
    return (int)out[0];
}
int main(void){int v=check_vst16();if(v==42){printf("PASS:VST16:%d\\n",v);return 0;}printf("FAIL:VST16:%d\\n",v);return 1;}
