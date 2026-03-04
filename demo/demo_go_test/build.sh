#!/bin/bash
# 交叉编译 Go 程序为 ARM64 Linux ELF (带符号表)
# 
# 用法: bash demo/demo_go_test/build.sh
#
# 注意: 
#   -ldflags="-compressdwarf=false" 保留完整调试信息
#   CGO_ENABLED=0 确保纯 Go 编译，无 C 依赖
#   -gcflags="-N -l" 禁用优化和内联，保证函数不被优化掉

set -e

cd "$(dirname "$0")"

echo "[*] Building Go ARM64 ELF..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -gcflags="-N -l" \
    -o ../../build/demo_go_test \
    .

echo "[+] Output: build/demo_go_test"
echo ""
echo "[*] Checking symbols..."
# 用 Go 的 objdump 或 nm 查看符号
go tool nm ../../build/demo_go_test 2>/dev/null | grep -i "checkKey" || echo "(use 'go tool nm' or 'aarch64-linux-gnu-nm' to inspect)"
echo ""
echo "[*] Done. Try:"
echo "  ./build/vmpacker.exe -info build/demo_go_test"
echo "  ./build/vmpacker.exe -func main.checkKey -v -debug build/demo_go_test"
