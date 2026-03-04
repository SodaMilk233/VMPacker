/*
 * demo_insn_stur_call.c — 测试 STUR 负偏移 + BL 跨函数调用
 *
 * 编译: aarch64-linux-gnu-gcc -O1 -static -o demo_stur_call
 * demo_insn_stur_call.c 运行: ./demo_stur_call → "STUR_CALL PASS"
 */
#include <stdio.h>
#include <string.h>

/* 外部辅助函数 (BL 调用目标) */
__attribute__((noinline)) int helper_add(int a, int b) { return a + b; }

__attribute__((noinline)) int helper_mul(int a, int b) { return a * b; }

/* 测试 STUR 负偏移: 向后写入字节 */
__attribute__((noinline)) void test_stur_neg(unsigned char *buf, int len) {
  /* 编译器可能用 STUR/STURB 来实现 buf[i-1] 等操作 */
  for (int i = len; i > 0; i--) {
    buf[i] = buf[i - 1]; /* 右移一位 */
  }
  buf[0] = 0xFF;
}

/* 测试跨函数 BL 调用 + STUR */
__attribute__((noinline)) int test_call_stur(int x) {
  int a = helper_add(x, 10);
  int b = helper_mul(a, 3);
  unsigned char tmp[8] = {1, 2, 3, 4, 5, 6, 7, 8};
  test_stur_neg(tmp, 7);
  int sum = 0;
  for (int i = 0; i < 8; i++)
    sum += tmp[i];
  return b + sum;
}

int main(void) {
  int r = test_call_stur(5);
  /*
   * x=5 → a=5+10=15 → b=15*3=45
   * tmp 初始 = {1,2,3,4,5,6,7,8}
   * test_stur_neg: 右移1位 → {0xFF,1,2,3,4,5,6,7}
   * sum = 0xFF+1+2+3+4+5+6+7 = 255+28 = 283
   * result = 45 + 283 = 328
   */
  if (r == 328) {
    printf("STUR_CALL PASS\n");
  } else {
    printf("STUR_CALL FAIL: got %d, expect 328\n", r);
  }
  return 0;
}
