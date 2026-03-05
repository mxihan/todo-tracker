# TODO Tracker Makefile
# 项目构建和管理命令

# 变量定义
BINARY_NAME := todo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin
GO := go
GOFLAGS := -v

# 构建信息
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# 默认目标
.PHONY: all
all: clean build

# ============================================================================
# 构建相关
# ============================================================================

## build: 构建当前平台的二进制文件
.PHONY: build
build:
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/todo

## build-all: 构建所有平台
.PHONY: build-all
build-all:
	@echo "构建所有平台..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/todo
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/todo
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/todo
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/todo
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/todo

## install: 安装到系统
.PHONY: install
install:
	@echo "安装 $(BINARY_NAME)..."
	$(GO) install ./cmd/todo

# ============================================================================
# 测试相关
# ============================================================================

## test: 运行所有测试
.PHONY: test
test:
	@echo "运行测试..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

## test-coverage: 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage: test
	@echo "生成覆盖率报告..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

## test-short: 运行短测试
.PHONY: test-short
test-short:
	$(GO) test -short -v ./...

# ============================================================================
# 代码质量
# ============================================================================

## lint: 运行代码检查
.PHONY: lint
lint:
	@echo "运行代码检查..."
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint" && exit 1)
	golangci-lint run ./...

## lint-fix: 自动修复代码问题
.PHONY: lint-fix
lint-fix:
	@echo "自动修复代码问题..."
	golangci-lint run --fix ./...

## fmt: 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	$(GO) fmt ./...

## vet: 运行go vet
.PHONY: vet
vet:
	@echo "运行go vet..."
	$(GO) vet ./...

# ============================================================================
# 依赖管理
# ============================================================================

## deps: 下载依赖
.PHONY: deps
deps:
	@echo "下载依赖..."
	$(GO) mod download

## deps-update: 更新依赖
.PHONY: deps-update
deps-update:
	@echo "更新依赖..."
	$(GO) mod tidy
	$(GO) get -u ./...

## deps-clean: 清理依赖缓存
.PHONY: deps-clean
deps-clean:
	@echo "清理依赖缓存..."
	$(GO) clean -modcache

# ============================================================================
# 发布相关
# ============================================================================

## release: 使用goreleaser发布
.PHONY: release
release:
	@echo "发布新版本..."
	@which goreleaser > /dev/null || (echo "请先安装 goreleaser" && exit 1)
	goreleaser release --rm-dist

## release-snapshot: 创建快照版本（不发布）
.PHONY: release-snapshot
release-snapshot:
	@echo "创建快照版本..."
	goreleaser release --snapshot --rm-dist

# ============================================================================
# 开发工具
# ============================================================================

## run: 本地运行
.PHONY: run
run:
	$(GO) run ./cmd/todo

## dev: 开发模式（热重载）
.PHONY: dev
dev:
	@echo "开发模式..."
	@which air > /dev/null || (echo "请先安装 air: go install github.com/cosmtrek/air@latest" && exit 1)
	air

## clean: 清理构建产物
.PHONY: clean
clean:
	@echo "清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# ============================================================================
# 帮助
# ============================================================================

## help: 显示帮助信息
.PHONY: help
help:
	@echo "TODO Tracker Makefile 帮助"
	@echo ""
	@echo "使用方法: make [目标]"
	@echo ""
	@echo "可用目标:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'