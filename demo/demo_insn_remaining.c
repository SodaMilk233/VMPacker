/*
 * demo_insn_remaining.c — TBZ/TBNZ, CCMP, CCMN, SVC combined test
 *
 * 4 test functions, each __attribute__((noinline)) with NOP padding > 72B.
 * Uses X9-X14 as safe temp registers.
 *
 * .inst encodings:
 *   TBZ  X9, #0,  +8  → 0x36000049
 *   TBZ  X9, #33, +8  → 0xB6080049
 *   TBNZ X9, #0,  +8  → 0x37000049
 *   CCMP X9, X10, #0, EQ → 0xFA4A0120
 *   CCMP X9, X11, #0, EQ → 0xFA4B0120
 *   CCMP X9, #5,  #4, NE → 0xFA451924
 *   CCMN X9, X10, #0, EQ → 0xBA4A0120
 *   CCMN X9, #3,  #4, NE → 0xBA431924
 *   SVC  #0              → 0xD4000001
 *
 * Cross-compile:
 *   aarch64-linux-gnu-gcc -O1 -static -o build/demo_insn_remaining demo/demo_insn_remaining.c
 */
#include <stdio.h>
#include <stdint.h>

/* ============================================================
 * test_tbz_tbnz: 4 sub-tests
 *   1) TBZ X9, #0, +8 with X9=2 (bit0=0 → taken, skip mov x10,#99)
 *   2) TBZ X9, #0, +8 with X9=3 (bit0=1 → not taken, executes mov x10,#99)
 *   3) TBNZ X9, #0, +8 with X9=3 (bit0=1 → taken, skip mov x10,#99)
 *   4) TBZ X9, #33, +8 with X9=(1<<33) (bit33=1 → not taken)
 * Returns: sum of sub-test results (0=pass each)
 * ============================================================ */
__attribute__((noinline))
uint64_t test_tbz_tbnz(void) {
    uint64_t result;
    __asm__ volatile(
        "mov x10, #0\n"          /* accumulator */

        /* sub-test 1: TBZ X9, #0, +8; X9=2 → bit0=0 → branch taken */
        "mov x9, #2\n"
        ".inst 0x36000049\n"     /* TBZ X9, #0, +8 */
        "mov x10, #99\n"         /* skipped if taken */
        /* land here: x10 should still be 0 */
        "add x10, x10, #0\n"    /* sub1 contribution = x10 (0 if pass) */
        "mov x11, x10\n"        /* save sub1 result */

        /* sub-test 2: TBZ X9, #0, +8; X9=3 → bit0=1 → not taken */
        "mov x9, #3\n"
        "mov x10, #0\n"
        ".inst 0x36000049\n"     /* TBZ X9, #0, +8 */
        "mov x10, #1\n"         /* executed (not taken) */
        /* x10=1 means not-taken happened (correct) */
        "cmp x10, #1\n"
        "cset x12, eq\n"        /* x12=1 if correct */
        "eor x12, x12, #1\n"   /* x12=0 if correct, 1 if wrong */
        "add x11, x11, x12\n"

        /* sub-test 3: TBNZ X9, #0, +8; X9=3 → bit0=1 → taken */
        "mov x9, #3\n"
        "mov x10, #0\n"
        ".inst 0x37000049\n"     /* TBNZ X9, #0, +8 */
        "mov x10, #99\n"         /* skipped if taken */
        "add x11, x11, x10\n"   /* +0 if pass */

        /* sub-test 4: TBZ X9, #33, +8; X9=(1<<33) → bit33=1 → not taken */
        "mov x9, #2\n"
        "lsl x9, x9, #32\n"     /* x9 = 1<<33 */
        "mov x10, #0\n"
        ".inst 0xB6080049\n"     /* TBZ X9, #33, +8 */
        "mov x10, #1\n"          /* executed (not taken, correct) */
        "cmp x10, #1\n"
        "cset x12, eq\n"
        "eor x12, x12, #1\n"
        "add x11, x11, x12\n"

        "mov %[out], x11\n"

        /* padding NOPs > 72 bytes */
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"

        : [out] "=r" (result)
        :
        : "x9", "x10", "x11", "x12", "memory"
    );
    return result;
}

/* ============================================================
 * test_ccmp: 3 sub-tests
 *   1) CCMP X9, X10, #0, EQ; prior CMP sets Z=1 → cond holds → real compare
 *   2) CCMP X9, X11, #0, EQ; prior CMP sets Z=0 → cond fails → nzcv=0
 *   3) CCMP X9, #5, #4, NE; prior CMP sets Z=0 → cond holds → real compare
 * ============================================================ */
__attribute__((noinline))
uint64_t test_ccmp(uint64_t a, uint64_t b) {
    uint64_t result;
    __asm__ volatile(
        "mov x9,  %[a]\n"
        "mov x10, %[b]\n"
        "mov x11, #0\n"          /* accumulator */

        /* sub-test 1: CMP X9, X9 (Z=1) then CCMP X9, X10, #0, EQ */
        "cmp x9, x9\n"           /* Z=1 → EQ holds */
        ".inst 0xFA4A0120\n"     /* CCMP X9, X10, #0, EQ → real compare X9 vs X10 */
        "cset x12, eq\n"         /* x12=1 if X9==X10 (a==b) */
        /* if a!=b, x12=0 → correct for a!=b case */

        /* sub-test 2: CMP X9, X10 (Z=0 if a!=b) then CCMP X9, X11_val, #0, EQ */
        "mov x13, #77\n"         /* x11 used as accum, use x13 as dummy */
        "cmp x9, x10\n"          /* Z=0 if a!=b → EQ fails */
        ".inst 0xFA4B0120\n"     /* CCMP X9, X11, #0, EQ → cond fails → nzcv=0 → Z=0 */
        "cset x12, eq\n"         /* x12=0 (Z=0 from default nzcv) → correct */
        "add x11, x11, x12\n"

        /* sub-test 3: CCMP X9, #5, #4, NE */
        "cmp x9, x10\n"          /* Z=0 if a!=b → NE holds */
        ".inst 0xFA451924\n"     /* CCMP X9, #5, #4, NE → real compare X9 vs 5 */
        "cset x12, eq\n"         /* x12=1 if a==5 */
        "add x11, x11, x12\n"

        "mov %[out], x11\n"

        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"

        : [out] "=r" (result)
        : [a] "r" (a), [b] "r" (b)
        : "x9", "x10", "x11", "x12", "x13", "memory"
    );
    return result;
}

/* ============================================================
 * test_ccmn: 3 sub-tests
 *   1) CCMN X9, X10, #0, EQ; prior CMP Z=1 → cond holds → CMN(X9, X10)
 *   2) CCMN X9, X10, #0, EQ; prior CMP Z=0 → cond fails → nzcv=0
 *   3) CCMN X9, #3, #4, NE; prior CMP Z=0 → cond holds → CMN(X9, 3)
 * ============================================================ */
__attribute__((noinline))
uint64_t test_ccmn(uint64_t a, uint64_t b) {
    uint64_t result;
    __asm__ volatile(
        "mov x9,  %[a]\n"
        "mov x10, %[b]\n"
        "mov x11, #0\n"

        /* sub-test 1: CMP X9, X9 (Z=1) then CCMN X9, X10, #0, EQ */
        "cmp x9, x9\n"
        ".inst 0xBA4A0120\n"     /* CCMN X9, X10, #0, EQ → CMN X9, X10 */
        "cset x12, eq\n"         /* Z=1 if (X9+X10)==0 */
        /* for a=10, b=5: 10+5=15 ≠ 0 → x12=0 → correct */

        /* sub-test 2: CMP X9, X10 (Z=0) then CCMN X9, X10, #0, EQ */
        "cmp x9, x10\n"
        ".inst 0xBA4A0120\n"     /* cond fails → nzcv=0 → Z=0 */
        "cset x12, eq\n"         /* x12=0 → correct */
        "add x11, x11, x12\n"

        /* sub-test 3: CCMN X9, #3, #4, NE */
        "cmp x9, x10\n"          /* Z=0 → NE holds */
        ".inst 0xBA431924\n"     /* CCMN X9, #3, #4, NE → CMN X9, 3 */
        "cset x12, eq\n"         /* Z=1 if (X9+3)==0 */
        "add x11, x11, x12\n"

        "mov %[out], x11\n"

        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"

        : [out] "=r" (result)
        : [a] "r" (a), [b] "r" (b)
        : "x9", "x10", "x11", "x12", "memory"
    );
    return result;
}

/* ============================================================
 * test_svc: SVC #0 syscall test
 *   Uses write(2, "SVC OK\n", 7) via SVC #0
 *   X8=64 (write), X0=2 (stderr), X1=buf, X2=7 (len)
 *   Returns bytes written (7 on success)
 * ============================================================ */
__attribute__((noinline))
uint64_t test_svc(void) {
    uint64_t result;
    /* "SVC OK\n" = 0x53 0x56 0x43 0x20 0x4F 0x4B 0x0A */
    __asm__ volatile(
        /* build string on stack */
        "sub sp, sp, #16\n"
        "mov x9, #0x5653\n"      /* 'SV' little-endian: 0x56='V', 0x53='S' → "SV" */
        "movk x9, #0x2043, lsl #16\n"  /* 'C ' */
        "movk x9, #0x4B4F, lsl #32\n"  /* 'OK' */
        "movk x9, #0x000A, lsl #48\n"  /* '\n\0' */
        "str x9, [sp]\n"

        /* syscall: write(2, sp, 7) */
        "mov x8, #64\n"          /* __NR_write */
        "mov x0, #2\n"           /* fd = stderr */
        "mov x1, sp\n"           /* buf */
        "mov x2, #7\n"           /* len */
        ".inst 0xD4000001\n"     /* SVC #0 */

        "mov %[out], x0\n"       /* result = bytes written */
        "add sp, sp, #16\n"

        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"
        "nop\n" "nop\n" "nop\n" "nop\n"

        : [out] "=r" (result)
        :
        : "x0", "x1", "x2", "x8", "x9", "memory"
    );
    return result;
}

int main(void) {
    int fail = 0;

    /* ---- TBZ/TBNZ ---- */
    uint64_t r1 = test_tbz_tbnz();
    printf("TBZ/TBNZ: result=%lu %s\n", r1, r1 == 0 ? "PASS" : "FAIL");
    if (r1 != 0) fail++;

    /* ---- CCMP (a=10, b=5) ---- */
    uint64_t r2 = test_ccmp(10, 5);
    printf("CCMP: result=%lu %s\n", r2, r2 == 0 ? "PASS" : "FAIL");
    if (r2 != 0) fail++;

    /* ---- CCMN (a=10, b=5) ---- */
    uint64_t r3 = test_ccmn(10, 5);
    printf("CCMN: result=%lu %s\n", r3, r3 == 0 ? "PASS" : "FAIL");
    if (r3 != 0) fail++;

    /* ---- SVC ---- */
    uint64_t r4 = test_svc();
    printf("SVC: wrote=%lu %s\n", r4, r4 == 7 ? "PASS" : "FAIL");
    if (r4 != 7) fail++;

    printf("\n=== %d/%d tests passed ===\n", 4 - fail, 4);
    return fail;
}
