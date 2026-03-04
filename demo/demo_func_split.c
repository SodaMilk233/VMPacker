/*
 * demo_func_split.c — 多函数分裂 Demo
 *
 * 验证 __attribute__((section(...))) 将函数分散到不同 ELF section
 * 后，通过函数指针间接调用仍能正确执行。
 *
 * 编译: aarch64-linux-gnu-gcc -static -O1 -nostdlib -march=armv8-a \
 *       -o demo_func_split demo/demo_func_split.c
 *
 * 预期输出: "AMBS" + 换行 (4个section的函数都正确执行)
 *   A = ALU, M = MEM, B = BRANCH, S = SYSTEM
 */

typedef unsigned long u64;
typedef unsigned int  u32;
typedef unsigned char u8;

/* ---- 模拟 handler 函数签名 ---- */
typedef u32 (*handler_fn)(u32 a, u32 b);

/* ---- Section 分裂: 4 个函数分别放到不同 section ---- */

__attribute__((section(".text.vm_alu"), noinline))
u32 h_alu_add(u32 a, u32 b) {
    return a + b;  /* ALU: 加法, 10+3=13 */
}

__attribute__((section(".text.vm_mem"), noinline))
u32 h_mem_load(u32 a, u32 b) {
    return a * b;  /* MEM: 乘法, 10*3=30 */
}

__attribute__((section(".text.vm_branch"), noinline))
u32 h_branch_cmp(u32 a, u32 b) {
    return (a > b) ? a - b : b - a;  /* BRANCH: 差值, |10-3|=7 */
}

__attribute__((section(".text.vm_system"), noinline))
u32 h_system_xor(u32 a, u32 b) {
    return a ^ b;  /* SYSTEM: 异或, 10^3=9 */
}

/* ---- _start ---- */
void _start(void) {
    /* 构建函数指针表 (volatile 防止编译器优化为直接调用) */
    volatile handler_fn table[4];
    table[0] = h_alu_add;
    table[1] = h_mem_load;
    table[2] = h_branch_cmp;
    table[3] = h_system_xor;

    u32 a = 10, b = 3;
    char buf[6];
    int idx = 0;

    /* 通过函数指针间接调用 */
    buf[idx++] = (table[0](a, b) == 13) ? 'A' : 'x';
    buf[idx++] = (table[1](a, b) == 30) ? 'M' : 'x';
    buf[idx++] = (table[2](a, b) ==  7) ? 'B' : 'x';
    buf[idx++] = (table[3](a, b) ==  9) ? 'S' : 'x';
    buf[idx++] = '\n';

    /* 单次 write syscall 输出全部结果 */
    register long x0 __asm__("x0") = 1;
    register long x1 __asm__("x1") = (long)buf;
    register long x2 __asm__("x2") = (long)idx;
    register long x8 __asm__("x8") = 64;
    __asm__ volatile("svc #0"
        : "+r"(x0)
        : "r"(x1), "r"(x2), "r"(x8)
        : "memory");

    /* exit(0) */
    x0 = 0;
    x8 = 93;
    __asm__ volatile("svc #0" : : "r"(x0), "r"(x8));
}
