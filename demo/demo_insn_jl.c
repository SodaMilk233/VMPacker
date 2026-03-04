#include <stdio.h>
__attribute__((noinline)) int check_jl(int x){int out=-1;__asm__ volatile(
"mov w0,#7\n"
"mov w1,%w[i]\n"
"cmp w1,#1\n"
"b.lt 1f\n"
"mov w0,#13\n"
"1:\n"
"mov w0,#42\n"
"mov %w[o],w0\n"
:[o]"=r"(out):[i]"r"(x):"w0","w1","cc","memory");
__asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");return out;}
int main(void){int v=check_jl(0);if(v==42){printf("PASS:JL:%d\n",v);return 0;}printf("FAIL:JL:%d\n",v);return 1;}

