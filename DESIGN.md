# ctools 设计文档

## 项目概览

ctools 是基于 Wails 的多合一桌面工具箱。后端采用 Go 编写，提供网络诊断、密码学算法以及其他实验性功能；前端使用 Vue 3、Vite 与 Vuetify（Material Design 3 风格）构建图形界面。项目分为四大类工具：文本处理、网络、密码学与其他实验工具。

主要特性：

- **网络能力**：批量 Ping、精简 POST 调试、服务器资产管理。
- **密码工具**：密钥解析/生成、非对称与对称运算、哈希/HMAC、证书签发与解析、DER 结构解析等。
- **其他工具**：SOCKS5 代理、GMSSL 测试、Mermaid 与 PlantUML 绘图工作台。
- **跨平台桌面体验**：通过 Wails 同时启动 Go 后端与 Vue 前端，支持热更新与独立构建。

## 后端架构

### 服务划分

后端位于 `backend/` 目录，遵循功能分层：

- `network`: 网络探测、HTTP 请求、服务器资产管理等。
- `crypto`: 密钥管理、杂凑、对称/非对称算法、证书处理等密码功能。
- `other`: 杂项服务（SOCKS5、GMSSL、Mermaid/PlantUML 辅助等）。

`main.go` 实例化 `NetworkService`、`CryptoService`、`OtherService` 并绑定到 Wails 应用；`app.go` 将 Wails 上下文注入各服务。

### 配置与持久化

- 所有持久化数据（如 `crypto_keys.json`、`ping_history.json` 等）默认写入 `os.UserConfigDir()/WailsToolbox`。
- 通过新增环境变量 `CTOOLS_CONFIG_DIR` 可以覆盖配置目录，便于测试与企业部署。
- 密钥文件现在以 `0600` 权限写入，避免明文密钥被同机用户读取。

### 密码工具改进

1. **Mermaid/PlantUML 辅助**：`OtherService` 新增 `RenderPlantUML`，接受 diagram 文本并通过 HTTP POST 方式调用指定 PlantUML 服务器，返回 Base64 编码的渲染结果。默认指向内网 Docker 服务器（`http://127.0.0.1:18080/plantuml`），同时支持自定义地址。
2. **安全性修复**：`extractPEMOrDER` 现正确处理裸 DER/HEX 数据，并在 PEM 模式下拦截尾随垃圾数据，解决此前无法解析非 PEM 输入的问题。
3. **配置目录控制**：新的环境变量避免测试污染真实用户配置，同时便于将数据指向受控磁盘。
4. **文件权限**：密钥存储改为 0600，同时在 JSON 写入失败时输出日志，提升可观测性。

### 单元测试

新增 `backend/crypto/service_test.go`，覆盖：

- `extractPEMOrDER` 能接受十六进制 DER 输入，并对附加垃圾数据报错。
- `GenerateKeyPair` 存储流程：在自定义目录内生成 RSA 密钥，确保文件存在且（非 Windows 环境下）权限为 0600。

运行命令：`go test ./backend/...`。CI 应保证 `crypto` 包覆盖率不低于 80%（项目保护要求）。

## 前端架构

### 结构

- 入口：`frontend/src/main.js`（Vite/Vue 3）。
- 配置：`frontend/src/config/menu.js`（侧边栏与路由 id 映射）。
- 路由：`frontend/src/router/index.js`，采用 Hash History。
- 视图：按类别划分到 `views/text|network|crypto|other`。
- UI：Vuetify 3，采用 Material Design 3 风格，组件统一两空格缩进。

### Mermaid 工作台优化

- 左右布局改进，预览区支持扩展为全宽，并依据视窗高度自适应。
- 新增缩放控制（放大/缩小/复位）与预览百分比显示。
- 支持自定义主题 JSON（传入 `themeVariables`），并在本地缓存主题、缩放、自动渲染等偏好。
- 预览区增加背景填充与缩放容器，保持大图滚动体验。

### PlantUML 工作台

- 新增 `/tool/other/plantuml` 路由与菜单，集成在「其他工具」分类。
- 支持选择 SVG/PNG/TXT 输出格式、渲染服务器预设（本地 Docker / 官方 / 自定义）、超时及自动预览。
- 前端调用后端 `RenderPlantUML`，避免 CORS 限制，同时允许调用内网 PlantUML 服务。
- 提供模板、缩放、全屏、下载、复制文本等辅助功能，并提示如何在内网通过 Docker 运行服务。

### 构建与依赖

- 安装：`npm --prefix frontend install`
- 开发：`wails dev`
- 构建：`npm --prefix frontend run build && wails build`
- 新增依赖：仅后端 Go 代码变更，无额外前端 npm 依赖。

构建时 Vite 会提示部分大体积 chunk（主要来自 Vuetify 与图形组件），目前通过文档提示即可。

## 数据流与交互

1. **前端调用**：Vue 组件通过 `frontend/wailsjs/go/**` 直接调用 Go 服务。新增方法 `RenderPlantUML` 已由 `wails generate module` 生成 TypeScript 绑定。
2. **数据缓存**：Mermaid 与 PlantUML 工具将草稿与 UI 偏好写入 `localStorage`，保持会话一致性。
3. **文件导出**：Mermaid/PlantUML 渲染结果可导出为 SVG/PNG，PNG 通过 Canvas 缩放导出，SVG 直接写入 Blob。

## 测试与验证

- 后端：`go test ./backend/...`
- 前端：`npm --prefix frontend run build`（确保无编译错误）
- 集成功能：建议手动验证
  1. Mermaid 工具：输入默认模板、切换主题、自定义主题 JSON、缩放、导出图片。
  2. PlantUML 工具：启动本地 PlantUML Server（Docker 即可），验证 SVG/PNG/TXT 输出、自动预览、缩放、导出。
  3. 密钥管理：在设置 `CTOOLS_CONFIG_DIR` 后运行一次 RSA 生成，确认密钥 JSON 存在且权限正确。

## 展望

- 可在 `other` 服务中继续扩展离线渲染能力（如内置 PlantUML 引擎或 Kroki 兼容层），减少对外部服务依赖。
- 密码工具未来可补充更多测试覆盖（特别是 ECC/SM 系列）与更严格的输入校验。
- 前端可考虑为大体积依赖开启 `build.rollupOptions.output.manualChunks` 以降低主包体积。
