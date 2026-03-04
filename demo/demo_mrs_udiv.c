/* demo_mrs_udiv.c -- 测试 MRS 和 UDIV 指令的 VMP 翻译
 *
 * 功能:
 *   1. timing_checkpoint: 用 MRS cntvct_el0 读取计时器
 *   2. timing_check:     用 MRS cntfrq_el0 + UDIV 计算时间差
 *
 * 编译:
 *   aarch64-linux-gnu-gcc -O2 -static -o demo_mrs_udiv demo_mrs_udiv.c
 */
#include <stdint.h>
#include <stdio.h>


/* 读取 ARM64 计时器值 */
uint64_t my_timing_checkpoint(void) {
  uint64_t val;
  __asm__ volatile("mrs %0, cntvct_el0" : "=r"(val));
  return val;
}

/* 用 MRS + UDIV 计算耗时(微秒) */
int my_timing_check(uint64_t t_start, uint64_t t_end, uint32_t threshold_us) {
  uint64_t freq;
  __asm__ volatile("mrs %0, cntfrq_el0" : "=r"(freq));

  if (freq == 0 || t_end <= t_start)
    return 0;

  uint64_t elapsed_us = (t_end - t_start) * 1000000 / freq;
  return elapsed_us > (uint64_t)threshold_us ? 1 : 0;
}

int main(void) {
  uint64_t t0 = my_timing_checkpoint();

  /* 做一些计算来消耗时间 */
  volatile uint64_t sum = 0;
  for (int i = 0; i < 1000; i++) {
    sum += i * i;
  }

  uint64_t t1 = my_timing_checkpoint();

  /* 阈值 1 秒 - 循环 1000 次不会超过 */
  int slow = my_timing_check(t0, t1, 1000000);
  /* 阈值 0 微秒 - 任何操作都会超过 */
  int fast = my_timing_check(t0, t1, 0);

  if (slow == 0 && fast == 1) {
    printf("MRS_UDIV PASS\n");
    return 0;
  } else {
    printf("MRS_UDIV FAIL slow=%d fast=%d\n", slow, fast);
    return 1;
  }
}
