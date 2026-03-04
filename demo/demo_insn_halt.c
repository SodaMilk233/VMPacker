#include <stdio.h>

/*
 * HALT 指令验证说明：
 * vmp 翻译器会在每个函数字节码尾部自动追加 OP_HALT。
 * 本 demo 保持函数逻辑极简，验证受保护函数执行完成后能正常停机并返回。
 */
__attribute__((noinline)) int check_halt(int x) {
    int y = x + 1;
    __asm__ volatile(
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop; "
        "nop; nop; nop; nop; nop; nop; nop; nop;");
    return y;
}

int main(void) {
    int v = check_halt(41);
    if (v == 42) {
        printf("PASS:HALT:%d\n", v);
        return 0;
    }
    printf("FAIL:HALT:%d\n", v);
    return 1;
}

