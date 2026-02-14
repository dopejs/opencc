# opencc

[English](../README.md) | [繁體中文](README.zh-TW.md) | [Español](README.es.md)

多 CLI 环境切换器，支持 Claude Code、Codex、OpenCode，带 API 代理自动故障转移。

## 功能

- **多 CLI 支持** — 支持 Claude Code、Codex、OpenCode 三种 CLI，可按项目配置
- **多配置管理** — 在 `~/.opencc/opencc.json` 中统一管理所有 API 配置
- **代理故障转移** — 内置 HTTP 代理，当主 provider 不可用时自动切换到备用
- **场景路由** — 根据请求特征（thinking、image、longContext 等）智能路由
- **项目绑定** — 将目录绑定到特定 profile 和 CLI，实现项目级自动配置
- **环境变量配置** — 在 provider 级别为每个 CLI 单独配置环境变量
- **TUI 配置界面** — 交互式终端界面，支持 Dashboard 和传统两种模式
- **Web 管理界面** — 浏览器可视化管理 provider、profile 和项目绑定
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

## 快速开始

```sh
# 打开 TUI 配置界面，创建第一个 provider
opencc config

# 启动（使用默认 profile）
opencc

# 使用指定 profile
opencc -p work

# 使用指定 CLI
opencc --cli codex
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `opencc` | 启动 CLI（使用项目绑定或默认配置） |
| `opencc -p <profile>` | 使用指定 profile 启动 |
| `opencc -p` | 交互选择 profile |
| `opencc --cli <cli>` | 使用指定 CLI（claude/codex/opencode） |
| `opencc use <provider>` | 直接使用指定 provider（无代理） |
| `opencc pick` | 交互选择 provider 启动 |
| `opencc list` | 列出所有 provider 和 profile |
| `opencc config` | 打开 TUI 配置界面 |
| `opencc config --legacy` | 使用传统 TUI 界面 |
| `opencc bind <profile>` | 绑定当前目录到 profile |
| `opencc bind --cli <cli>` | 绑定当前目录使用指定 CLI |
| `opencc unbind` | 解除当前目录绑定 |
| `opencc status` | 显示当前目录绑定状态 |
| `opencc web start` | 启动 Web 管理界面 |
| `opencc web open` | 在浏览器中打开 Web 界面 |
| `opencc web stop` | 停止 Web 服务 |
| `opencc upgrade` | 升级到最新版本 |
| `opencc version` | 显示版本 |

## 多 CLI 支持

opencc 支持三种 AI 编程助手 CLI：

| CLI | 说明 | API 格式 |
|-----|------|---------|
| `claude` | Claude Code（默认） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### 设置默认 CLI

```sh
# 通过 TUI
opencc config  # Settings → Default CLI

# 通过 Web UI
opencc web open  # Settings 页面
```

### 按项目配置 CLI

```sh
cd ~/work/project
opencc bind --cli codex  # 该目录使用 Codex
```

### 临时使用其他 CLI

```sh
opencc --cli opencode  # 本次使用 OpenCode
```

## Profile 管理

Profile 是一组 provider 的有序列表，用于故障转移。

### 配置示例

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

### 使用 Profile

```sh
# 使用默认 profile
opencc

# 使用指定 profile
opencc -p work

# 交互选择
opencc -p
```

## 项目绑定

将目录绑定到特定 profile 和/或 CLI，实现项目级自动配置。

```sh
cd ~/work/company-project

# 绑定 profile
opencc bind work-profile

# 绑定 CLI
opencc bind --cli codex

# 同时绑定
opencc bind work-profile --cli codex

# 查看状态
opencc status

# 解除绑定
opencc unbind
```

**优先级**：命令行参数 > 项目绑定 > 全局默认

## TUI 配置界面

```sh
opencc config
```

v1.5 提供全新 Dashboard 界面：

- **左侧列表**：Providers、Profiles、Project Bindings
- **右侧详情**：选中项的详细信息
- **快捷键**：
  - `a` - 添加新项
  - `e` - 编辑选中项
  - `d` - 删除选中项
  - `Tab` - 切换焦点
  - `q` - 返回/退出

使用 `--legacy` 切换到传统界面。

## Web 管理界面

```sh
# 启动（后台运行，端口 19840）
opencc web start

# 打开浏览器
opencc web open

# 停止
opencc web stop
```

Web UI 功能：
- Provider 和 Profile 管理
- 项目绑定管理
- 全局设置（默认 CLI、默认 Profile、端口）
- 请求日志查看
- 模型字段自动补全

## 环境变量配置

每个 provider 可以为不同 CLI 配置独立的环境变量：

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

### Claude Code 常用环境变量

| 变量 | 说明 |
|------|------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大输出 token |
| `MAX_THINKING_TOKENS` | 扩展思考预算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文窗口 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 默认超时 |

## 场景路由

根据请求特征自动路由到不同 provider：

| 场景 | 触发条件 |
|------|---------|
| `think` | 启用 thinking 模式 |
| `image` | 包含图片内容 |
| `longContext` | 内容超过阈值 |
| `webSearch` | 使用 web_search 工具 |
| `background` | 使用 Haiku 模型 |

**Fallback 机制**：如果场景配置的 providers 全部失败，会自动 fallback 到 profile 的默认 providers。

配置示例：

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```

## 配置文件

| 文件 | 说明 |
|------|------|
| `~/.opencc/opencc.json` | 主配置文件 |
| `~/.opencc/proxy.log` | 代理日志 |
| `~/.opencc/web.log` | Web 服务日志 |

### 完整配置示例

```json
{
  "version": 5,
  "default_profile": "default",
  "default_cli": "claude",
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "cli": "codex"
    }
  }
}
```

## 升级

```sh
# 最新版本
opencc upgrade

# 指定版本
opencc upgrade 1.5
opencc upgrade 1.5.0
```

## 从旧版迁移

如果之前使用 `~/.cc_envs/` 格式，opencc 会自动迁移到 `~/.opencc/opencc.json`。

## 开发

```sh
# 构建
go build -o opencc .

# 测试
go test ./...
```

发布：打 tag 后 GitHub Actions 自动构建。

```sh
git tag v1.5.1
git push origin v1.5.1
```

## License

MIT
