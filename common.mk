SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c

# このファイルのディレクトリを取得する
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
# ビルド用ディレクトリ
BUILD_DIR := $(COMMON_SELF_DIR)bin

TINYGO := tinygo
TINYGO_FLAGS := -target pico --serial uart
TINYGO_BUILD := $(TINYGO) build $(TINYGO_FLAGS)
TINYGO_GDB := $(TINYGO) gdb $(TINYGO_FLAGS)
ELF2UF2 := $(COMMON_SELF_DIR)elf2uf2-build/elf2uf2
OPENOCD := openocd
OPENOCD_FLAGS := -f interface/cmsis-dap.cfg -f target/rp2040.cfg -c "adapter speed 5000"
OPENOCD_START := $(OPENOCD) $(OPENOCD_FLAGS)
OPENOCD_LOAD := $(OPENOCD) $(OPENOCD_FLAGS) -c "tcl_port disabled" -c "gdb_port disabled"
MKDIR := mkdir -p
RM := rm -rf

# TARGET には、生成したい .elf ファイルを定義する
# TARGET は、Makefile 側の定義を優先する。
# デフォルト値は、Makefile の "ディレクトリ名.elf"
TARGET ?= $(notdir $(CURDIR)).elf

.PHONY: all build clean load openocd gdb

all: build

build: $(BUILD_DIR)/$(TARGET) $(BUILD_DIR)/$(TARGET:.elf=.uf2)

clean:
	$(RM) $(BUILD_DIR)/$(TARGET) $(BUILD_DIR)/$(TARGET:.elf=.uf2)

# .elf ファイル生成
$(BUILD_DIR)/$(TARGET):

# .uf2 ファイル生成
$(BUILD_DIR)/$(TARGET:.elf=.uf2): $(BUILD_DIR)/$(TARGET)

# .elf ファイル生成ルール
%.elf:
	$(MKDIR) $(BUILD_DIR)
	$(TINYGO_BUILD) -o $@

# .uf2 ファイル生成ルール
%.uf2: %.elf
	$(MKDIR) $(BUILD_DIR)
	$(ELF2UF2) $< $@

load: $(BUILD_DIR)/$(TARGET)
	$(OPENOCD_LOAD) -c "program $(BUILD_DIR)/$(TARGET) verify reset exit"

openocd:
	$(OPENOCD_START)

gdb:
	$(TINYGO_GDB)
