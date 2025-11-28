<div align="center">

# Git Commit Guide CLI

一个可交互的 Git 提交信息助手，帮助团队快速生成符合约定式规范的 Commit Message。

</div>

## ✨ 功能特性

- **约定式提交**：统一输出 `<type>(<scope>): <desc>` 结构，并支持正文与脚注。
- **交互式引导**：问答式采集工单号、类型、影响范围、描述、正文、脚注。
- **分支推断**：自动从当前分支提取 `bcds-xxxx` 作为默认工单号。
- **输入记忆**：最近一次输入会保存到 `.git-guide/git-commit.json`，再次运行自动带出暗色占位。
- **安全校验**：限制描述长度、首字母格式、结尾符号等，避免不规范提交。
- **终端友好**：彩色提示、空行多行输入、Y/N 快捷确认等增强体验。

## 📦 安装

准备好 Go 1.21+（或你的 `go.mod` 指定版本），推荐直接使用 `go install`：

```bash
go install github.com/ibreez3/git-cm@latest
```

安装完成后，`git-cm` 将出现在 `$GOBIN`（默认 `$GOPATH/bin` 或 `$HOME/go/bin`）。将其加入 `PATH`，或创建别名 `alias gcm=git-cm` 方便调用。

如果你更喜欢手动编译：

```bash
git clone https://github.com/ibreez3/git-cm.git
cd git-cm
go build -o git-cm
```

## 🚀 使用

1. 在已有 Git 仓库中暂存改动：`git add ...`
2. 运行二进制或别名：`git-cm` / `gcm`
3. 按提示输入信息，确认后自动执行 `git commit -m "<message>"`。

示例对话（部分输出）：

```text
=== Git 提交助手 (Custom Commit CLI) ===
1. 请输入工单号 ...: bcds-2839

2. 请选择提交类型:
   1. feat     新功能 (A new feature)
   2. fix      Bug修复 (A bug fix)
   ...
   请输入编号 (1-10): 1

3. 请输入影响范围（可选，如 api、payment，回车跳过）: api

4. 请输入简短描述（≤50字符，首字母小写，无句号）:
   描述: improve migrate recovery orchestration

需要添加详细正文吗？ (y/N): y
请输入正文，每行 ≤72 字符 (直接回车结束):
重构迁移执行顺序
...
```

最后 CLI 会展示并确认完整信息，例如：

```text
feat(api): improve migrate recovery orchestration

重构迁移执行顺序

BREAKING CHANGE: 新的迁移策略要求升级 agent
Refs bcds-2839
```

## ⚙️ 缓存与配置

- 上次输入会序列化到 `.git-guide/git-commit.json`，随仓库一起保存。
- 默认提示优先使用缓存值，若无缓存则尝试从当前分支名中提取工单号。
- 不需要的缓存可以手动删除 `.git-guide/git-commit.json`。

## 🧪 开发与调试

常用命令：

```bash
# 运行 CLI
go run ./main.go

# 格式化
gofmt -w main.go

# 依赖检查 / 编译
go vet ./...
go build
```

提交前可运行 `go test ./...`（若未来加入测试）。

## 📝 约定

CLI 内置常见提交类型：

| 类型 | 场景 | SemVer 影响 |
| ---- | ---- | ----------- |
| feat | 新功能 | minor |
| fix | Bug 修复 | patch |
| docs | 文档更新 | - |
| style | 代码风格 | - |
| refactor | 重构 | - |
| perf | 性能优化 | patch |
| test | 测试相关 | - |
| build | 构建依赖 | patch |
| ci | CI 配置 | - |
| chore | 其他 | - |

破坏性变更需要在正文或脚注中添加 `BREAKING CHANGE: ...`。

## 🤝 贡献

欢迎通过 Issue 或 PR 反馈需求及问题。提交 PR 时请：

1. 确认 `gofmt` 已执行。
2. 使用本工具生成的规范提交信息。
3. 在描述中注明改动背景和验证方式。

## 📄 许可证

本项目基于 [MIT License](./LICENSE) 发布，可自由使用、修改与分发。

---

若你在团队中需要统一的 Git 提交流程，这个小工具可以帮助大家保持一致。欢迎 Star ⭐ 与分享！

