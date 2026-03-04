/*
 * demo_simple.c — 最简 VMP 测试: 纯计算，无外部调用
 *
 * check_simple(a) = a * 3 + 7
 *
 * 编译: aarch64-linux-gnu-gcc -O1 -o demo_simple demo_simple.c
 * 保护: vmpacker -func check_simple -v -o demo_simple.vmp demo_simple
 * 测试: ./demo_simple.vmp    → 应输出 37
 */

int check_simple(int a) {
  int b = a * 3;
  int c = b + 7;
  return c;
}

/* 最简 _start 入口, 避免使用 libc */
void _start(void) {
  int result = check_simple(10); /* 期望: 37 */

  /* 使用 syscall write(1, &ch, 1) 输出单个字符 */
  char ch = '0' + (result % 10); /* '7' */
  char nl = '\n';

  /* syscall: write(fd=1, buf=&ch, len=1) */
  register long x0 asm("x0") = 1;
  register long x1 asm("x1") = (long)&ch;
  register long x2 asm("x2") = 1;
  register long x8 asm("x8") = 64; /* __NR_write */
  asm volatile("svc #0" : : "r"(x0), "r"(x1), "r"(x2), "r"(x8));

  /* write newline */
  x1 = (long)&nl;
  asm volatile("svc #0" : : "r"(x0), "r"(x1), "r"(x2), "r"(x8));

  /* syscall: exit(result) */
  x0 = result;
  x8 = 93; /* __NR_exit */
  asm volatile("svc #0" : : "r"(x0), "r"(x8));
}
