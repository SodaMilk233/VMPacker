/*
 * demo_license.c — VMP 端到端测试 demo
 *
 * 编译:  aarch64-linux-gnu-gcc -O1 -o demo_license demo_license.c
 * 保护:  vmpacker -func check_license -v -o demo_license.vmp demo_license
 * 运行:  ./demo_license.vmp 12345678
 *        ./demo_license.vmp 00000000
 */

#include <stdio.h>
#include <string.h>

/*
 * check_license — 简单的许可证验证函数
 *
 * 这个函数将被 VMP 保护。它使用了翻译器已支持的指令:
 *   - MOV (MOVZ/MOVK)
 *   - ADD/SUB/MUL/XOR/AND/OR/LSL/LSR
 *   - CMP + B.cond
 *   - LDR/STR (字节加载)
 *   - 循环
 *   - RET
 *
 * 算法: 对输入 key 的每个字节做 hash，与期望值比较
 *   hash = 0
 *   for each byte b in key:
 *       hash = (hash * 31 + b) ^ 0x5A
 *   return hash == EXPECTED
 */
int check_license(const char *key) {
  if (key == 0)
    return 0;

  unsigned long hash = 0;
  int len = 0;

  /* 计算长度 (不用 strlen 避免外部调用) */
  const char *p = key;
  while (*p != 0) {
    len++;
    p++;
  }

  /* 密钥长度必须为 8 */
  if (len != 8)
    return 0;

  /* 计算 hash */
  for (int i = 0; i < 8; i++) {
    unsigned char b = (unsigned char)key[i];
    hash = (hash * 31 + b) ^ 0x5A;
  }

  /* 期望值: check_license("12345678") == 1 */
  unsigned long expected = 0;
  const char *valid = "12345678";
  for (int i = 0; i < 8; i++) {
    unsigned char b = (unsigned char)valid[i];
    expected = (expected * 31 + b) ^ 0x5A;
  }

  return (hash == expected) ? 1 : 0;
}

int main(int argc, char *argv[]) {
  if (argc < 2) {
    printf("Usage: %s <license-key>\n", argv[0]);
    return 1;
  }

  int result = check_license(argv[1]);

  if (result) {
    printf("[+] License valid!\n");
  } else {
    printf("[-] License invalid.\n");
  }

  return result ? 0 : 1;
}
