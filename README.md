# opencc

Claude Code 多环境切换器，支持 API 代理自动故障转移。

## 功能

- **多配置管理** — 在 `~/.opencc/opencc.json` 中统一管理所有 API 配置，随时切换
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **Fallback Profiles** — 多个命名的故障转移配置，按场景快速切换（work / staging / …）
- **TUI 配置界面** — 交互式终端界面管理配置、profile 和故障转移顺序
- **自更新** — `opencc upgrade` 一键升级，支持 semver 版本匹配
- **Shell 补全** — 支持 zsh / bash / fish

## 安装

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
```

卸载：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall
```

## 使用

### 创建配置

通过 TUI 界面创建：

```sh
opencc config
```

或手动编辑 `~/.opencc/opencc.json`：

```json
{
  "providers": {
    "work": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "reasoning_model": "claude-sonnet-4-5-thinking",
      "haiku_model": "claude-haiku-4-5",
      "opus_model": "claude-opus-4-5",
      "sonnet_model": "claude-sonnet-4-5"
    },
    "backup": {
      "base_url": "https://backup.example.com",
      "auth_token": "sk-..."
    }
  },
  "profiles": {
    "default": ["work", "backup"],
    "staging": ["staging-provider"]
  }
}
```

### 命令一览

| 命令 | 说明 |
|------|------|
| `opencc` | 以代理模式启动 claude（使用 default profile） |
| `opencc -f work` | 使用名为 "work" 的 fallback profile 启动 |
| `opencc -f` | 交互选择一个 profile 后启动 |
| `opencc use <config>` | 使用指定配置直接启动 claude |
| `opencc pick` | 交互选择 provider 启动（不保存） |
| `opencc list` | 列出所有配置（按 fallback 顺序排列） |
| `opencc config` | 打开 TUI 配置管理界面 |
| `opencc upgrade` | 升级到最新版本 |
| `opencc upgrade 1.2` | 升级到 1.2.x 最新版本 |
| `opencc version` | 显示当前版本 |
| `opencc completion zsh/bash/fish` | 生成 shell 补全脚本 |

### 故障转移

opencc 支持多个命名的 fallback profile，用于不同使用场景。

Profile 配置在 `~/.opencc/opencc.json` 的 `profiles` 字段中：

```json
{
  "profiles": {
    "default": ["work", "backup", "personal"],
    "work": ["work-primary", "work-secondary"],
    "staging": ["staging-provider"]
  }
}
```

#### 使用 Profile

```sh
# 使用 default profile（等同于之前的行为）
opencc

# 使用指定 profile
opencc -f work

# 交互选择 profile
opencc -f
```

通过 `opencc config` 进入 TUI，按 `f` 键管理 fallback profiles — 可创建、编辑、删除 profile 及调整各 profile 内的 provider 顺序。

启动时 opencc 会启动一个本地 HTTP 代理，按顺序尝试各 provider。当前 provider 返回 429 或 5xx 时自动切换到下一个，并对失败的 provider 进行指数退避。

### 升级

```sh
# 升级到最新版本
opencc upgrade

# 升级到 1.x.x 最新版本
opencc upgrade 1

# 升级到 1.2.x 最新版本
opencc upgrade 1.2

# 升级到精确版本
opencc upgrade 1.2.3
```

### 配置文件说明

| 文件 | 说明 |
|------|------|
| `~/.opencc/opencc.json` | 统一 JSON 配置文件（providers + profiles） |
| `~/.opencc/proxy.log` | 代理运行日志 |

每个 provider 支持以下字段：

| 字段 | 必填 | 说明 |
|------|------|------|
| `base_url` | 是 | API 地址 |
| `auth_token` | 是 | API 密钥 |
| `model` | 否 | 主模型，默认 `claude-sonnet-4-5` |
| `reasoning_model` | 否 | 推理模型，默认 `claude-sonnet-4-5-thinking` |
| `haiku_model` | 否 | Haiku 模型，默认 `claude-haiku-4-5` |
| `opus_model` | 否 | Opus 模型，默认 `claude-opus-4-5` |
| `sonnet_model` | 否 | Sonnet 模型，默认 `claude-sonnet-4-5` |

### 从旧版迁移

如果之前使用 `~/.cc_envs/` 格式的配置文件，opencc 会在首次运行时自动迁移到 `~/.opencc/opencc.json`。旧目录不会被删除，可以手动清理。

## 开发

需要 Go 1.25+。

```sh
# 构建
go build -o opencc .

# 测试
go test ./...

# 构建当前平台二进制
./deploy.sh

# 构建所有平台二进制
./deploy.sh --all
```

发布流程：打 tag 后 GitHub Actions 自动构建并创建 Release。

```sh
git tag v1.2.0
git push origin v1.2.0
```

## 目录结构

```
├── main.go              # 入口
├── cmd/                 # CLI 命令 (cobra)
├── internal/
│   ├── config/          # 统一 JSON 配置管理（Store + 迁移）
│   └── proxy/           # HTTP 代理服务器
├── tui/                 # TUI 界面 (bubbletea)
├── install.sh           # 用户安装脚本
└── deploy.sh            # 构建发布脚本
```

## License

MIT
