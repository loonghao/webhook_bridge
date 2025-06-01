# CI/CD 最佳实践指南

## 🔍 问题背景

在 GitHub Actions CI 环境中，经常遇到以下 Go 版本不匹配错误：

```
compile: version "go1.23.0" does not match go tool version "go1.22.12"
```

这个问题的根本原因是：
1. **Go 版本不一致**：不同的 CI job 使用了不同版本的 Go
2. **缓存冲突**：Go 工具链缓存了不同版本的编译器和工具
3. **依赖安装顺序**：protobuf 工具安装时使用了错误的 Go 版本

## 🚀 解决方案

### 1. 统一版本管理

**创建环境变量配置文件** (`.github/env`)：
```bash
# GitHub Actions Environment Variables
GO_VERSION=1.23
GOLANGCI_LINT_VERSION=v1.64.6
NODE_VERSION=20
PYTHON_VERSION=3.11
CGO_ENABLED=0
```

**在每个 CI job 中加载环境变量**：
```yaml
steps:
- name: Checkout code
  uses: actions/checkout@v4

- name: Add variables to environment file
  run: cat ".github/env" >> "$GITHUB_ENV"

- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    check-latest: true
    cache: true
```

### 2. CI 环境设置脚本

创建 `dev/ci-setup.sh` 脚本来：
- 验证 Go 版本一致性
- 清理 Go 缓存以解决版本冲突
- 安装必要的 Go 工具
- 设置 Go 模块

**使用方式**：
```yaml
- name: Setup CI environment
  shell: bash
  run: |
    chmod +x dev/ci-setup.sh
    ./dev/ci-setup.sh
```

### 3. 缓存清理命令

添加开发工具命令来清理 Go 缓存：

```bash
# 清理所有 Go 缓存
go run dev.go clean-cache
```

这个命令会清理：
- 构建缓存 (`go clean -cache`)
- 模块缓存 (`go clean -modcache`)
- 测试缓存 (`go clean -testcache`)
- 已安装的包 (`go clean -i all`)

## 📋 最佳实践

### 1. 版本管理策略

**参考知名项目**：
- **Helm**: 使用 `.github/env` 文件统一管理版本
- **Lazygit**: 在 workflow 顶部定义环境变量
- **Kubernetes**: 使用矩阵策略但保持版本一致性

**推荐做法**：
```yaml
# ✅ 好的做法：统一版本管理
env:
  GO_VERSION: '1.23'

strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
    # 不要在这里定义 go-version

# ❌ 避免的做法：多版本矩阵测试
strategy:
  matrix:
    go-version: ['1.21', '1.22', '1.23']  # 容易导致版本冲突
```

### 2. CI Job 优化

**简化 CI 配置**：
```yaml
steps:
- name: Checkout code
  uses: actions/checkout@v4

- name: Add variables to environment file
  run: cat ".github/env" >> "$GITHUB_ENV"

- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ env.GO_VERSION }}
    check-latest: true
    cache: true

- name: Setup CI environment
  shell: bash
  run: |
    chmod +x dev/ci-setup.sh
    ./dev/ci-setup.sh
```

### 3. 构建速度优化

**缓存策略**：
- 使用 `actions/setup-go@v5` 的内置缓存
- 避免在 CI 中清理模块缓存（除非必要）
- 使用 `check-latest: true` 确保版本一致性

**并行化策略**：
```yaml
# 测试、构建、linting 并行执行
jobs:
  test:
    # ...
  lint:
    # ...
  build:
    needs: [test, lint]  # 只有测试和 lint 通过才构建
```

## 🛠️ 故障排除

### 常见问题和解决方案

1. **版本不匹配错误**
   ```bash
   # 解决方案：清理缓存并重新安装工具
   go run dev.go clean-cache
   go mod download
   ```

2. **protobuf 工具版本冲突**
   ```bash
   # 解决方案：重新安装 protobuf 工具
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. **CI 缓存问题**
   ```yaml
   # 解决方案：在 CI 中强制清理缓存
   - name: Clean Go caches
     run: |
       go clean -cache
       go clean -modcache
       go clean -testcache
   ```

### 调试技巧

**检查环境信息**：
```bash
go version
go env GOPATH
go env GOCACHE
go env GOMODCACHE
which protoc
which protoc-gen-go
```

**验证工具版本**：
```bash
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

## 📚 参考资料

- [GitHub Actions Go 最佳实践](https://docs.github.com/en/actions/use-cases-and-examples/building-and-testing/building-and-testing-go)
- [Helm CI 配置](https://github.com/helm/helm/blob/main/.github/workflows/build-test.yml)
- [Lazygit CI 配置](https://github.com/jesseduffield/lazygit/blob/master/.github/workflows/ci.yml)
- [Go 模块最佳实践](https://go.dev/blog/using-go-modules)
