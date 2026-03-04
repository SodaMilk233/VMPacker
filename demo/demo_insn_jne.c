#include <stdio.h>
__attribute__((noinline)) int check_jne(int x){int out=-1;__asm__ volatile(
"mov w0,#7\n"
"mov w1,%w[i]\n"
"cmp w1,#0\n"
"b.ne 1f\n"
"mov w0,#13\n"
"1:\n"
"mov w0,#42\n"
"mov %w[o],w0\n"
:[o]"=r"(out):[i]"r"(x):"w0","w1","cc","memory");
__asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");return out;}
int main(void){int v=check_jne(1);if(v==42){printf("PASS:JNE:%d\n",v);return 0;}printf("FAIL:JNE:%d\n",v);return 1;}

