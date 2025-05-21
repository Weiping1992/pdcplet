# Go相关配置
GO = go
BINARY_NAME = pdcplet
BINARY_DIR = bin/
SRC_FILES = $(wildcard *.go)
TEST_OUTPUT_DIR = ./test/

# MOCK配置
MOCK_DIR = mock/
MOCK_BIN = mockserver
MOCK_SRC = mockserver.go

# 构建参数
BUILD_FLAGS = -ldflags="-s -w" # 减小二进制体积
RACE_FLAGS = -race             # 竞态检测

.PHONY: all build run test clean help

## 默认目标：构建项目
all: build

## 编译项目
build:
	$(GO) build $(BUILD_FLAGS) -o $(BINARY_DIR)$(BINARY_NAME) main.go
	$(GO) build $(BUILD_FLAGS) -o $(BINARY_DIR)$(MOCK_BIN) $(MOCK_DIR)$(MOCK_SRC)

## 编译并运行
run: build
	./$(BINARY_NAME)

## 运行单元测试
test:
	mkdir -p $(TEST_OUTPUT_DIR)
	$(GO) test -v -covermode=atomic -coverprofile=$(TEST_OUTPUT_DIR)/coverage.out ./...

## 生成测试覆盖率报告
coverage: test
	$(GO) tool cover -html=$(TEST_OUTPUT_DIR)/coverage.out -o $(TEST_OUTPUT_DIR)/coverage.html

## 多平台交叉编译
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-linux-amd64

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY_NAME)-windows-amd64.exe

build-mac:
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BINARY_NAME)-darwin-arm64

## 代码格式化
fmt:
	$(GO) fmt ./...
	gofmt -s -w .

## 静态代码检查
lint:
	golangci-lint run --enable-all

## 清理构建产物
clean:
	$(GO) clean
	rm -f $(BINARY_DIR)$(BINARY_NAME)*
	rm -f $(BINARY_DIR)$(MOCK_BIN)*
	rm -rf $(TEST_OUTPUT_DIR)

## 显示帮助
help:
	@echo "可用命令:"
	@echo "  make          编译项目"
	@echo "  make run      编译并运行"
	@echo "  make test     运行单元测试"
	@echo "  make coverage 生成测试覆盖率报告"
	@echo "  make fmt      格式化代码"
	@echo "  make lint     静态代码检查"
	@echo "  make clean    清理构建产物"
	@echo "  make help     显示帮助信息"
	@echo ""
	@echo "交叉编译命令:"
	@echo "  make build-linux   生成Linux版本"
	@echo "  make build-windows 生成Windows版本"
	@echo "  make build-mac     生成macOS版本"