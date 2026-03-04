/*
 * demo_ldadd_basic.c — Phase B: LDADD/CAS 基础翻译测试（不用 X15）
 * 排除 R15 冲突因素，纯测 VMP 翻译是否正确
 */
#include <stdint.h>
#include <stdio.h>

/* 基础 LDADD: X0 做 Rs, 正常寄存器 */
__attribute__((noinline)) int test_ldadd_basic(void) {
  int fail = 0;
  uint64_t mem = 100;
  uint64_t addend = 42;
  uint64_t old_val;
  /* LDADD addend, old_val, [&mem] → old_val=100, mem=142 */
  __asm__ volatile("ldadd %2, %0, [%1]"
                   : "=r"(old_val)
                   : "r"(&mem), "r"(addend)
                   : "memory");
  if (old_val != 100) {
    printf("  FAIL: ldadd_basic old exp=100 got=%lu\n", (unsigned long)old_val);
    fail++;
  } else {
    printf("  PASS: ldadd_basic old=%lu\n", (unsigned long)old_val);
  }
  if (mem != 142) {
    printf("  FAIL: ldadd_basic mem exp=142 got=%lu\n", (unsigned long)mem);
    fail++;
  } else {
    printf("  PASS: ldadd_basic mem=%lu\n", (unsigned long)mem);
  }
  return fail;
}

/* 基础 CAS: 正常寄存器 */
__attribute__((noinline)) int test_cas_basic(void) {
  int fail = 0;
  uint64_t mem = 0xAAAA;
  uint64_t expected = 0xAAAA;
  uint64_t newval = 0xBBBB;
  /* CAS expected, newval, [&mem]
   * 比较: mem(0xAAAA) == expected(0xAAAA) → 相等
   * 结果: mem = newval(0xBBBB), expected = old_mem(0xAAAA) */
  __asm__ volatile("cas %0, %2, [%1]"
                   : "+r"(expected)
                   : "r"(&mem), "r"(newval)
                   : "memory");
  if (expected != 0xAAAA) {
    printf("  FAIL: cas_basic expected exp=0xAAAA got=0x%lx\n",
           (unsigned long)expected);
    fail++;
  } else {
    printf("  PASS: cas_basic expected=0x%lx\n", (unsigned long)expected);
  }
  if (mem != 0xBBBB) {
    printf("  FAIL: cas_basic mem exp=0xBBBB got=0x%lx\n", (unsigned long)mem);
    fail++;
  } else {
    printf("  PASS: cas_basic mem=0x%lx\n", (unsigned long)mem);
  }
  return fail;
}

int main(void) {
  int total = 0;
  printf("=== Phase B: LDADD/CAS 基础测试 ===\n");
  total += test_ldadd_basic();
  total += test_cas_basic();
  printf("===================================\n");
  if (total == 0)
    printf("ALL PASS\n");
  else
    printf("%d FAIL(s)\n", total);
  return total;
}
