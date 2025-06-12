# GoReleaser 配置改进

基于 [Grype 项目](https://github.com/anchore/grype) 的最佳实践，我们对 webhook-bridge 的 `.goreleaser.yml` 配置进行了以下改进：

## 主要改进

### 1. 版本和发布配置
- 添加了 `version: 2` 以使用 GoReleaser v2 格式
- 配置了自动预发布检测 (`prerelease: auto`)

### 2. 多架构支持扩展
- **新增架构支持**：
  - `ppc64le` (PowerPC 64-bit Little Endian)
  - `s390x` (IBM System z)
- **架构排除策略**：
  - Windows 和 macOS 跳过 ARM 和特殊架构构建
  - 保持主流平台的完整支持

### 3. 构建优化
- **可重现构建**：
  - 添加 `mod_timestamp: '{{ .CommitTimestamp }}'`
  - 确保构建时间戳一致性
- **增强的 ldflags**：
  - 添加 `-extldflags '-static'` 用于静态链接
  - 新增 `gitCommit` 和 `gitDescription` 构建信息
  - 使用 YAML 锚点 (`&build-ldflags`) 避免重复

### 4. 包管理器支持
- **新增 Linux 包支持**：
  - RPM 包 (Red Hat/CentOS/Fedora)
  - DEB 包 (Debian/Ubuntu)
- **包配置**：
  - 系统级配置文件安装到 `/etc/webhook-bridge/`
  - 共享资源安装到 `/usr/share/webhook-bridge/`
  - 依赖管理：`python3` (必需)，`python3-pip` (推荐)

### 5. Docker 多架构支持
- **分离的架构构建**：
  - AMD64: `webhook-bridge:latest-amd64`, `webhook-bridge:{{ .Tag }}-amd64`
  - ARM64: `webhook-bridge:{{ .Tag }}-arm64v8`
- **Docker Manifests**：
  - 自动创建多架构 manifest
  - 支持 `latest`, `{{ .Tag }}`, `v{{ .Major }}`, `v{{ .Major }}.{{ .Minor }}` 标签
- **增强的构建参数**：
  - 完整的 OCI 标签支持
  - 构建参数传递 (`BUILD_DATE`, `BUILD_VERSION`, `VCS_REF`, `VCS_URL`)

### 6. 代码签名准备
- 添加了 Cosign 签名配置模板（已注释）
- 支持 OIDC 签名流程
- 为未来的安全增强做准备

## 配置对比

### Grype 项目的优秀特性
✅ **已采用**：
- 多架构支持 (`ppc64le`, `s390x`)
- Docker 多架构构建和 manifests
- 可重现构建时间戳
- 增强的 ldflags 配置
- 代码签名框架

✅ **已适配**：
- Linux 包管理器支持 (RPM/DEB)
- 完整的 OCI 标签
- 环境变量优化

### webhook-bridge 项目的特色
🚀 **保持的优势**：
- Next.js 前端集成
- Python 执行器支持
- 完整的开发工具链
- Homebrew 和 Scoop 支持
- 丰富的文档和示例

## 使用方法

### 本地测试
```bash
# 快照构建（测试用）
go run dev.go release-snapshot

# 干运行（验证配置）
goreleaser release --skip=publish --clean
```

### 正式发布
```bash
# 创建标签
git tag v1.0.0
git push origin v1.0.0

# 自动触发 GitHub Actions 发布
# 或手动发布
goreleaser release --clean
```

### 多架构 Docker 使用
```bash
# 拉取多架构镜像（自动选择合适架构）
docker pull ghcr.io/loonghao/webhook-bridge:latest

# 指定架构
docker pull ghcr.io/loonghao/webhook-bridge:v1.0.0-amd64
docker pull ghcr.io/loonghao/webhook-bridge:v1.0.0-arm64v8
```

### Linux 包安装
```bash
# Debian/Ubuntu
sudo dpkg -i webhook-bridge_*.deb

# Red Hat/CentOS/Fedora
sudo rpm -i webhook-bridge_*.rpm
```

## 注意事项

1. **构建时间**：多架构支持会增加构建时间
2. **存储空间**：Docker 多架构镜像会占用更多存储
3. **测试覆盖**：建议在不同架构上测试关键功能
4. **依赖管理**：确保 Python 执行器在所有架构上正常工作

## 下一步计划

- [ ] 启用代码签名 (Cosign)
- [ ] 添加更多架构支持 (如需要)
- [ ] 集成安全扫描
- [ ] 优化构建缓存策略
