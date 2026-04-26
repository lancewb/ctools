# ctools 设计文档

## 项目概览

ctools 是基于 Wails 的 Windows/Linux 常用小工具集合。后端采用 Go 标准库优先的实现策略，前端使用 Vue 3、Vite 与 Vuetify 构建图形界面。当前工具按 8 大类组织：文本工具、编码转换、网络调试、监控观测、密码运算、证书 ASN.1、开发辅助、其他小工具。

主要特性：

- **轻量化**：移除 Mermaid、PlantUML 与内置 jar，不再引入大体积绘图依赖。
- **网络调试**：HTTP 客户端、TCP 客户端、端口扫描、DNS 查询、群 Ping、服务器管理。
- **监控观测**：Prometheus `/metrics` 周期抓取，支持按间隔刷新并显示动态图表。
- **密码与证书**：密钥解析/生成、对称/非对称运算、哈希/HMAC、证书签发与解析、DER 结构解析、GMSSL 检测。
- **日常辅助**：JSON 格式化、正则测试、文本 Diff、URL/Base64/Hex、JWT、时间戳/UUID、颜色转换等。

## 后端架构

后端位于 `backend/` 目录，遵循功能分层：

- `network`: HTTP 请求、批量 Ping、DNS、端口扫描、TCP 客户端、Prometheus metrics 抓取、服务器资产管理。
- `crypto`: 密钥管理、杂凑、对称/非对称算法、证书处理。
- `other`: SOCKS5 代理、GMSSL/TLCP 测试等杂项服务。

`main.go` 实例化 `NetworkService`、`CryptoService`、`OtherService` 并绑定到 Wails 应用；`app.go` 将 Wails 上下文注入各服务。

## 配置与持久化

- 持久化数据默认写入 `os.UserConfigDir()/WailsToolbox`。
- `CTOOLS_CONFIG_DIR` 可覆盖配置目录，便于测试与部署。
- 密钥文件以 `0600` 权限写入。

## 构建

- 前端构建：`npm --prefix frontend run build`
- 后端测试：`go test ./...`
- Linux Wails 构建：`wails build -clean -tags webkit2_41`
- GitHub Actions 支持手动输入版本号，并在 Linux/Windows 构建归档后发布。

## 测试

Go 测试覆盖主要后端能力：

- 密码工具：哈希/HMAC、对称加解密、RSA/ECC/SM2 运算、证书与 DER。
- 网络工具：HTTP、DNS、端口扫描、TCP 客户端、Prometheus 解析、历史/配置持久化。
- 其他工具：SOCKS5 代理与 GMSSL 参数校验。

前端通过 Vite 构建验证所有视图与路由。
