# Contributing to GameTrainerHub

感谢你对 **GameTrainerHub** 的关注！欢迎通过提交 Issue 或 Pull Request 参与项目改进。

## 开发环境要求

* **操作系统**：Windows 10 (1809+) 或 Windows 11 (64-bit)
* **Go**：1.26 或更高版本
* **Node.js**：20.x 或更高版本
* **Wails CLI**：`v2.17.1`
* **NSIS**：3.x（如需测试 Windows 安装包编译）

## 提交流程

1. **Fork 本仓库** 并创建功能或修复分支（例如 `feat/your-feature` 或 `fix/issue-desc`）。
2. **编写代码与测试**：
   * 后端 Go 代码修改需确保通过 `go vet ./...` 与 `go test -count=1 ./internal/...`。
   * 前端 Vue / TypeScript 修改需确保通过 `cd frontend && npm run build`（含 `vue-tsc` 类型检查）。
3. **检查敏感信息**：严禁在代码中包含硬编码密钥、Token、本地数据库文件（`*.db`）或个人环境配置。
4. **提交规范**：建议使用 Conventional Commits 格式：
   * `feat: ...` 新功能
   * `fix: ...` 问题修复
   * `docs: ...` 文档更新
   * `build: ...` 构建与 CI 配置调整
   * `refactor: ...` 代码重构
5. **发起 Pull Request**：清晰描述改动动机、实现方式及本地测试验证结果。
