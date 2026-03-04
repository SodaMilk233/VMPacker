package main

import "fmt"

// checkKey — 纯计算函数，用于测试 VMP 保护 Go 编译产物
// 算法: ((input * 7) + 42) ^ 0xFF
//
//go:noinline
func checkKey(input int) int {
	a := input * 7
	b := a + 42
	c := b ^ 0xFF
	return c
}

func main() {
	result := checkKey(10)
	fmt.Printf("checkKey(10) = %d\n", result)
	// 期望: ((10 * 7) + 42) ^ 0xFF = (70 + 42) ^ 255 = 112 ^ 255 = 143
	if result == 143 {
		fmt.Println("[+] OK")
	} else {
		fmt.Println("[-] FAIL")
	}
}
