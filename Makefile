.PHONY: build build-cli run test clean install

# 项目名称
APP_NAME := meme-server
CLI_NAME := meme-cli
BUILD_DIR := ./build

# Go 参数
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 构建 MCP Server
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# 构建 CLI 工具
build-cli:
	@echo "Building $(CLI_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(CLI_NAME) ./cmd/cli

# 构建所有
build-all: build build-cli

# 跨平台构建
build-release:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/server
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_NAME)-darwin-amd64 ./cmd/cli
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(CLI_NAME)-darwin-arm64 ./cmd/cli
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(CLI_NAME)-linux-amd64 ./cmd/cli

# 运行 MCP Server
run: build
	@echo "Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME)

# ============ 本地测试命令 ============

# 测试搜索 (使用 CLI)
search: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "$(word 2,$(MAKECMDGOALS))" -v

# 快速搜索测试
test-cat: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "猫" -l 5 -v

test-dog: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "狗" -l 5 -v

test-happy: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "开心" -l 5 -v

# 测试指定源
test-source: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "猫" -s pdan -l 5 -v

# 测试多个源
test-multi: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "猫" -s pdan,sougou -l 3 -v

# 列出所有源
list: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -list

# JSON 输出测试
test-json: build-cli
	$(BUILD_DIR)/$(CLI_NAME) -k "猫" -l 3 -json

# ============ MCP 协议测试 ============

# 测试 MCP list_tools
test-mcp-tools: build
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | $(BUILD_DIR)/$(APP_NAME)

# 测试 MCP search_meme
test-mcp-search: build
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search_meme","arguments":{"keyword":"猫","limit":5}}}' | $(BUILD_DIR)/$(APP_NAME)

# 测试 MCP list_sources
test-mcp-list: build
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_sources","arguments":{}}}' | $(BUILD_DIR)/$(APP_NAME)

# ============ 开发工具 ============

# 单元测试
test:
	go test -v ./...

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# 安装依赖
deps:
	go mod tidy
	go mod download

# 安装到系统
install: build build-cli
	@echo "Installing to /usr/local/bin..."
	@cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@cp $(BUILD_DIR)/$(CLI_NAME) /usr/local/bin/

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 允许传递额外参数
%:
	@:
