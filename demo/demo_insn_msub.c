#include <stdio.h>
#include <stdint.h>

__attribute__((noinline)) int64_t check_msub(void) {
    int64_t a = 100, b = 3, c = 7;
    int64_t r1 = 0, r2 = 0;

    /* msub r1, b, c, a  =>  r1 = a - b*c = 100 - 3*7 = 79 */
    __asm__ volatile(
        "msub %[o1], %[b], %[c], %[a]\n"
        : [o1] "=r"(r1)
        : [a] "r"(a), [b] "r"(b), [c] "r"(c)
    );
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");

    /* msub r2, b, c, xzr  =>  r2 = 0 - b*c = -21 */
    __asm__ volatile(
        "msub %[o2], %[b], %[c], xzr\n"
        : [o2] "=r"(r2)
        : [b] "r"(b), [c] "r"(c)
    );
    __asm__ volatile("nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop; nop;");

    if (r1 == 79 && r2 == -21) return 1;
    return 0;
}

int main(void) {
    int64_t v = check_msub();
    if (v == 1) { printf("PASS:MSUB\n"); return 0; }
    printf("FAIL:MSUB r=%ld\n", v);
    return 1;
}
