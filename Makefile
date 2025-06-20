# HackMITM Makefile

# 变量定义
BINARY_NAME=hackmitm
MAIN_PATH=./cmd/hackmitm
BUILD_DIR=./build
DOCKER_IMAGE=hackmitm
VERSION?=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go 相关设置
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=0

# 默认目标
.PHONY: help
help: ## 显示帮助信息
	@echo "可用的命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 依赖管理
.PHONY: deps
deps: ## 安装依赖
	go mod download
	go mod tidy

# 构建
.PHONY: build
build: deps ## 构建可执行文件
	@echo "构建 $(BINARY_NAME) v$(VERSION)..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 交叉编译
.PHONY: build-all
build-all: deps ## 构建所有平台的可执行文件
	@echo "开始交叉编译..."
	mkdir -p $(BUILD_DIR)
	
	# Linux amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	# Linux arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows amd64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo "交叉编译完成"

# 安装
.PHONY: install
install: build ## 安装到系统路径
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin/"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "安装完成"

# 测试
.PHONY: test
test: ## 运行测试
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## 运行测试并显示覆盖率
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

# 代码质量
.PHONY: lint
lint: ## 运行代码检查
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，请运行: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: fmt
fmt: ## 格式化代码
	go fmt ./...

.PHONY: vet
vet: ## 运行 go vet
	go vet ./...

# Docker
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 8080:8080 \
		-v $(PWD)/certs:/app/certs \
		-v $(PWD)/configs:/app/configs \
		$(DOCKER_IMAGE):latest

.PHONY: docker-stop
docker-stop: ## 停止 Docker 容器
	docker stop $(DOCKER_IMAGE) || true
	docker rm $(DOCKER_IMAGE) || true

# 清理
.PHONY: clean
clean: ## 清理构建文件
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -rf ./certs/*.pem

.PHONY: clean-all
clean-all: clean docker-stop ## 清理所有文件和容器
	docker rmi $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest 2>/dev/null || true

# 发布
.PHONY: release
release: clean build-all ## 构建发布版本
	@echo "创建发布包..."
	mkdir -p $(BUILD_DIR)/release
	
	# 创建 tar.gz 包
	for binary in $(BUILD_DIR)/$(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			platform=$$(basename $$binary | sed 's/$(BINARY_NAME)-//'); \
			mkdir -p $(BUILD_DIR)/release/$(BINARY_NAME)-$$platform; \
			cp $$binary $(BUILD_DIR)/release/$(BINARY_NAME)-$$platform/$(BINARY_NAME); \
			cp README.md $(BUILD_DIR)/release/$(BINARY_NAME)-$$platform/; \
			cp -r configs $(BUILD_DIR)/release/$(BINARY_NAME)-$$platform/; \
			cp -r docs $(BUILD_DIR)/release/$(BINARY_NAME)-$$platform/; \
			cd $(BUILD_DIR)/release && tar -czf $(BINARY_NAME)-$$platform.tar.gz $(BINARY_NAME)-$$platform; \
			cd - > /dev/null; \
			rm -rf $(BUILD_DIR)/release/$(BINARY_NAME)-$$platform; \
		fi \
	done
	
	@echo "发布包创建完成: $(BUILD_DIR)/release/"

# 开发
.PHONY: dev
dev: build ## 开发模式运行
	./$(BUILD_DIR)/$(BINARY_NAME) -verbose -log-level debug

.PHONY: dev-watch
dev-watch: ## 监控文件变化并重新构建
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air 未安装，请运行: go install github.com/cosmtrek/air@latest"; \
		echo "或者手动运行: make dev"; \
	fi

# 证书管理
.PHONY: export-ca
export-ca: build ## 导出 CA 证书
	mkdir -p ./certs
	./$(BUILD_DIR)/$(BINARY_NAME) -export-ca ./certs/ca-cert.pem
	@echo "CA 证书已导出到: ./certs/ca-cert.pem"

.PHONY: clean-certs
clean-certs: ## 清理证书文件
	rm -rf ./certs/*.pem
	@echo "证书文件已清理"

# 配置管理
.PHONY: config-example
config-example: ## 生成示例配置文件
	mkdir -p ./configs
	@if [ ! -f ./configs/config.json ]; then \
		echo "配置文件已存在，跳过生成"; \
	else \
		echo "示例配置已在 ./configs/config.json"; \
	fi

# 基准测试
.PHONY: bench
bench: ## 运行基准测试
	go test -bench=. -benchmem ./...

# 性能分析
.PHONY: pprof-cpu
pprof-cpu: ## CPU 性能分析
	@echo "启动 CPU 性能分析 (需要服务器运行在 :6060)"
	go tool pprof http://localhost:6060/debug/pprof/profile

.PHONY: pprof-mem
pprof-mem: ## 内存性能分析
	@echo "启动内存性能分析 (需要服务器运行在 :6060)"
	go tool pprof http://localhost:6060/debug/pprof/heap

# 文档
.PHONY: docs
docs: ## 生成文档
	@echo "生成 Go 文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "访问 http://localhost:6060/pkg/hackmitm/ 查看文档"; \
		godoc -http=:6060; \
	else \
		echo "godoc 未安装，请运行: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"

# 环境检查
.PHONY: check-env
check-env: ## 检查开发环境
	@echo "检查开发环境..."
	@go version
	@echo "Go 模块: $$(go list -m)"
	@echo "GOPATH: $$(go env GOPATH)"
	@echo "GOROOT: $$(go env GOROOT)"
	@if command -v docker >/dev/null 2>&1; then echo "Docker: $$(docker --version)"; else echo "Docker: 未安装"; fi
	@if command -v git >/dev/null 2>&1; then echo "Git: $$(git --version)"; else echo "Git: 未安装"; fi 