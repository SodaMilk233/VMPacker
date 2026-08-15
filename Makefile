# ============================================================
# VMP 工具链 Makefile
# make all    → 编译 C stub → 嵌入 Go → 输出到 build/
# make stub   → 仅编译 VM 解释器 blob
# make packer → 仅编译 Go packer（需先 make stub）
# make demo   → 交叉编译 demo 程序
# make test   → 运行 Go 单元测试
# make clean  → 清理所有产物
# ============================================================

# Cross toolchain
GO ?= go
POWERSHELL ?= powershell

ifeq ($(OS),Windows_NT)
NDK_ROOT ?= C:/Users/Administrator/AppData/Local/Android/Sdk/ndk/28.1.13356709
NDK_HOST_TAG ?= windows-x86_64
NDK_API ?= 21
NDK_LLVM_BIN := $(NDK_ROOT)/toolchains/llvm/prebuilt/$(NDK_HOST_TAG)/bin
HOST_LLVM_ROOT ?= C:/Program Files/LLVM
HOST_LLVM_BIN := $(HOST_LLVM_ROOT)/bin

CC = "$(NDK_LLVM_BIN)/clang.exe"
LD = "$(HOST_LLVM_BIN)/ld.lld.exe"
OBJCOPY = "$(HOST_LLVM_BIN)/llvm-objcopy.exe"
NM = "$(HOST_LLVM_BIN)/llvm-nm.exe"
READELF = "$(NDK_LLVM_BIN)/llvm-readelf.exe"
TARGET_CFLAGS = --target=aarch64-linux-android$(NDK_API) -mno-outline-atomics
STUB_TARGET_CFLAGS = $(TARGET_CFLAGS) -fno-pic -fno-pie
PACKER_EXT = .exe
else
CROSS ?= aarch64-linux-gnu-
CC = $(CROSS)gcc
LD = $(CROSS)ld
OBJCOPY = $(CROSS)objcopy
NM = $(CROSS)nm
READELF = $(CROSS)readelf
TARGET_CFLAGS =
STUB_TARGET_CFLAGS =
PACKER_EXT =
endif

# 目录
STUB_DIR   = stub/linux/arm64
CMD_DIR    = cmd/vmpacker
DEMO_DIR   = demo
BUILD_DIR  = build

# ------ VM 解释器 blob ------
STUB_SRC   = $(STUB_DIR)/vm_interp.c
STUB_LDS   = $(STUB_DIR)/vm_interp.lds
STUB_O     = $(BUILD_DIR)/stub/vm_interp.o
STUB_ELF   = $(BUILD_DIR)/stub/vm_interp.elf
STUB_BIN   = $(CMD_DIR)/vm_interp.bin
STUB_RAW   = $(BUILD_DIR)/vm_interp_raw.bin
STUB_META  = scripts/make_stub_blob.ps1

# ------ Go packer ------
PACKER     = $(BUILD_DIR)/vmpacker$(PACKER_EXT)

# ------ Demo ------
DEMO_LICENSE     = $(BUILD_DIR)/demo_license
DEMO_SIMPLE      = $(BUILD_DIR)/demo_simple

# 编译选项 (必须 -mcmodel=tiny，禁止 -fPIC)
STUB_CFLAGS = $(STUB_TARGET_CFLAGS) -c -Os -mcmodel=tiny -fno-stack-protector \
              -fno-builtin -nostdlib -march=armv8-a \
              -DVM_INDIRECT_DISPATCH -DVM_FUNC_SPLIT -DVM_TOKEN_ENTRY

DEMO_CFLAGS = $(TARGET_CFLAGS) -static -O0 -march=armv8-a
DEMO_LDFLAGS = -Wl,--build-id

# ============================================================
.PHONY: all stub packer demo test clean help toolchain-info gui gui-installer sync-public

all: stub packer
	@echo ""
	@echo "[+] Build complete: $(BUILD_DIR)/"

toolchain-info:
	@$(CC) $(TARGET_CFLAGS) --version
	@$(LD) --version
	@$(OBJCOPY) --version
	@$(NM) --version

# ------ VM 解释器 blob ------
stub: $(STUB_BIN)

$(STUB_O): $(STUB_SRC) | $(BUILD_DIR)/stub
	$(CC) $(STUB_CFLAGS) $< -o $@

$(STUB_ELF): $(STUB_O) $(STUB_LDS)
	$(LD) -T $(STUB_LDS) -o $@ $(STUB_O)

$(STUB_BIN): $(STUB_ELF) $(STUB_META) | $(BUILD_DIR)
	$(OBJCOPY) -O binary $< $(STUB_RAW)
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File $(STUB_META) \
		-ElfPath "$<" -RawPath "$(STUB_RAW)" -OutputPath "$@" \
		-NmPath $(NM) -ReadElfPath $(READELF)
	@$(POWERSHELL) -NoProfile -Command "Copy-Item -LiteralPath '$(STUB_BIN)' -Destination '$(BUILD_DIR)/vm_interp.bin' -Force"

# ------ Go packer (embed vm_interp.bin) ------
packer: $(STUB_BIN) | $(BUILD_DIR)
	@$(POWERSHELL) -NoProfile -Command "if (Test-Path '$(PACKER)') { Remove-Item -Force '$(PACKER)' -ErrorAction SilentlyContinue }"
	$(GO) build -o $(PACKER) ./$(CMD_DIR)/
	@echo "[+] packer: $(PACKER)"

# ------ Demo 程序 ------
demo: $(DEMO_LICENSE) $(DEMO_SIMPLE)

$(DEMO_LICENSE): $(DEMO_DIR)/demo_license.c | $(BUILD_DIR)
	$(CC) $(DEMO_CFLAGS) $(DEMO_LDFLAGS) $< -o $@
	@echo "[+] demo: $@"

$(DEMO_SIMPLE): $(DEMO_DIR)/demo_simple.c | $(BUILD_DIR)
	$(CC) $(TARGET_CFLAGS) -static -O1 -nostdlib -march=armv8-a $(DEMO_LDFLAGS) $< -o $@
	@echo "[+] demo: $@"

# ------ 测试 ------
test:
	$(GO) test ./...

# ------ 目录创建 ------
$(BUILD_DIR):
	@$(POWERSHELL) -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(BUILD_DIR)' | Out-Null"

$(BUILD_DIR)/stub: | $(BUILD_DIR)
	@$(POWERSHELL) -NoProfile -Command "New-Item -ItemType Directory -Force -Path '$(BUILD_DIR)/stub' | Out-Null"

# ------ 清理 ------
clean:
	@$(POWERSHELL) -NoProfile -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '$(BUILD_DIR)', '$(STUB_BIN)'"
	@echo "[+] cleaned"

# ------ 帮助 ------
help:
	@echo "make all     - 编译 stub + packer (输出到 build/)"
	@echo "make toolchain-info - 显示当前 NDK/LLVM 工具链"
	@echo "make stub    - 仅编译 VM 解释器 blob"
	@echo "make packer  - 编译 Go packer (自动嵌入 blob)"
	@echo "make gui     - 编译 GUI Windows 可执行文件"
	@echo "make gui-installer - 编译 GUI + NSIS 安装包（需要 NSIS）"
	@echo "make demo    - 交叉编译 demo 程序"
	@echo "make test    - 运行单元测试"
	@echo "make clean        - 清理所有产物"
	@echo "make sync-public  - 同步到公开仓库 (vmpack remote)"

# ------ GUI 版本 (Wails + NSIS) ------
GUI_DIR = vmp-gui
GUI_STUB_BIN = $(GUI_DIR)/backend/api/vm_interp.bin
WAILS_VERSION ?= v2.11.0
WAILS_PACKAGE = github.com/wailsapp/wails/v2/cmd/wails@$(WAILS_VERSION)

gui: stub
	@$(POWERSHELL) -NoProfile -Command "Copy-Item -LiteralPath '$(STUB_BIN)' -Destination '$(GUI_STUB_BIN)' -Force"
	cd $(GUI_DIR) && $(GO) run $(WAILS_PACKAGE) build
	@echo "[+] GUI: $(GUI_DIR)/build/bin/vmp-gui.exe"

gui-installer: stub
	@$(POWERSHELL) -NoProfile -Command "Copy-Item -LiteralPath '$(STUB_BIN)' -Destination '$(GUI_STUB_BIN)' -Force"
	@$(POWERSHELL) -NoProfile -Command "$$env:PATH = 'C:\Program Files (x86)\NSIS;' + $$env:PATH; Set-Location '$(GUI_DIR)'; go run '$(WAILS_PACKAGE)' build -nsis"
	@echo "[+] GUI installer: $(GUI_DIR)/build/bin/"

# ------ 同步公开仓库 ------
sync-public:
	@powershell -ExecutionPolicy Bypass -File sync-public.ps1

