/*
 * demo_token_entry.c — Token 32-bit 编码/解码 roundtrip 验证
 *
 * 布局:
 *   bits[31:24] = XOR key    (8  bits, 0-255)
 *   bits[23:12] = bc_offset  (12 bits, 0-4095)
 *   bits[11:0]  = func_id    (12 bits, 0-4095)
 */

#include <stdint.h>
#include <stdio.h>

/* ---- 解码宏 ---- */
#define TOKEN_FUNC_ID(tok)    ((tok) & 0xFFF)
#define TOKEN_BC_OFFSET(tok)  (((tok) >> 12) & 0xFFF)
#define TOKEN_XOR_KEY(tok)    (((tok) >> 24) & 0xFF)

/* ---- 编码宏 ---- */
#define TOKEN_ENCODE(func_id, bc_offset, xor_key) \
    (((uint32_t)(xor_key) << 24) | ((uint32_t)(bc_offset) << 12) | ((uint32_t)(func_id) & 0xFFF))

/* 单条测试用例 */
static int test_roundtrip(uint32_t func_id, uint32_t bc_offset, uint32_t xor_key)
{
    uint32_t tok = TOKEN_ENCODE(func_id, bc_offset, xor_key);
    uint32_t d_fid = TOKEN_FUNC_ID(tok);
    uint32_t d_off = TOKEN_BC_OFFSET(tok);
    uint32_t d_key = TOKEN_XOR_KEY(tok);

    int ok = (d_fid == func_id) && (d_off == bc_offset) && (d_key == xor_key);

    printf("  func_id=%-4u bc_offset=%-4u xor_key=0x%02X  =>  tok=0x%08X  =>  "
           "fid=%-4u off=%-4u key=0x%02X  [%s]\n",
           func_id, bc_offset, xor_key, tok,
           d_fid, d_off, d_key,
           ok ? "PASS" : "FAIL");
    return ok;
}

int main(void)
{
    int fail = 0;

    printf("=== Token 32-bit roundtrip 验证 ===\n\n");

    /* 边界值 */
    printf("[边界值]\n");
    fail += !test_roundtrip(0,    0,    0);
    fail += !test_roundtrip(4095, 4095, 255);
    fail += !test_roundtrip(4095, 0,    0);
    fail += !test_roundtrip(0,    4095, 0);
    fail += !test_roundtrip(0,    0,    255);

    /* 典型值 */
    printf("\n[典型值]\n");
    fail += !test_roundtrip(42,   100,  0xA5);
    fail += !test_roundtrip(1,    1,    1);
    fail += !test_roundtrip(256,  512,  128);

    /* 随机值 */
    printf("\n[随机值]\n");
    fail += !test_roundtrip(3333, 2048, 0xDE);
    fail += !test_roundtrip(1023, 3071, 0x7F);
    fail += !test_roundtrip(2222, 1111, 0x55);
    fail += !test_roundtrip(4000, 300,  0xAB);
    fail += !test_roundtrip(777,  888,  0x01);

    printf("\n=== 结果: %s (%d 项失败) ===\n",
           fail ? "FAIL" : "ALL PASS", fail);

    return fail ? 1 : 0;
}
