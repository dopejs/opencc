# opencc

[English](../README.md) | [简体中文](README.zh-CN.md) | [Español](README.es.md)

多 CLI 環境切換器，支援 Claude Code、Codex、OpenCode，具備 API 代理自動故障轉移。

## 功能

- **多 CLI 支援** — 支援 Claude Code、Codex、OpenCode 三種 CLI，可依專案設定
- **多組態管理** — 在 `~/.opencc/opencc.json` 中統一管理所有 API 組態
- **代理故障轉移** — 內建 HTTP 代理，當主要 provider 無法使用時自動切換至備用
- **場景路由** — 根據請求特徵（thinking、image、longContext 等）智慧路由
- **專案綁定** — 將目錄綁定至特定 profile 與 CLI，實現專案層級自動組態
- **環境變數設定** — 在 provider 層級為每個 CLI 分別設定環境變數
- **TUI 設定介面** — 互動式終端介面，支援 Dashboard 與傳統兩種模式
- **Web 管理介面** — 瀏覽器視覺化管理 provider、profile 與專案綁定
- **自動更新** — `opencc upgrade` 一鍵升級，支援 semver 版本比對
- **Shell 補全** — 支援 zsh / bash / fish

## 安裝

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
```

解除安裝：

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall
```

## 快速開始

```sh
# 開啟 TUI 設定介面，建立第一個 provider
opencc config

# 啟動（使用預設 profile）
opencc

# 使用指定 profile
opencc -p work

# 使用指定 CLI
opencc --cli codex
```

## 命令一覽

| 命令 | 說明 |
|------|------|
| `opencc` | 啟動 CLI（使用專案綁定或預設組態） |
| `opencc -p <profile>` | 使用指定 profile 啟動 |
| `opencc -p` | 互動選擇 profile |
| `opencc --cli <cli>` | 使用指定 CLI（claude/codex/opencode） |
| `opencc use <provider>` | 直接使用指定 provider（不經代理） |
| `opencc pick` | 互動選擇 provider 啟動 |
| `opencc list` | 列出所有 provider 與 profile |
| `opencc config` | 開啟 TUI 設定介面 |
| `opencc config --legacy` | 使用傳統 TUI 介面 |
| `opencc bind <profile>` | 將目前目錄綁定至 profile |
| `opencc bind --cli <cli>` | 將目前目錄綁定使用指定 CLI |
| `opencc unbind` | 解除目前目錄綁定 |
| `opencc status` | 顯示目前目錄綁定狀態 |
| `opencc web start` | 啟動 Web 管理介面 |
| `opencc web open` | 在瀏覽器中開啟 Web 介面 |
| `opencc web stop` | 停止 Web 服務 |
| `opencc upgrade` | 升級至最新版本 |
| `opencc version` | 顯示版本 |

## 多 CLI 支援

opencc 支援三種 AI 程式設計助手 CLI：

| CLI | 說明 | API 格式 |
|-----|------|---------|
| `claude` | Claude Code（預設） | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### 設定預設 CLI

```sh
# 透過 TUI
opencc config  # Settings → Default CLI

# 透過 Web UI
opencc web open  # Settings 頁面
```

### 依專案設定 CLI

```sh
cd ~/work/project
opencc bind --cli codex  # 此目錄使用 Codex
```

### 臨時使用其他 CLI

```sh
opencc --cli opencode  # 本次使用 OpenCode
```

## Profile 管理

Profile 是一組 provider 的有序清單，用於故障轉移。

### 組態範例

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
# 使用預設 profile
opencc

# 使用指定 profile
opencc -p work

# 互動選擇
opencc -p
```

## 專案綁定

將目錄綁定至特定 profile 和／或 CLI，實現專案層級自動組態。

```sh
cd ~/work/company-project

# 綁定 profile
opencc bind work-profile

# 綁定 CLI
opencc bind --cli codex

# 同時綁定
opencc bind work-profile --cli codex

# 檢視狀態
opencc status

# 解除綁定
opencc unbind
```

**優先順序**：命令列參數 > 專案綁定 > 全域預設

## TUI 設定介面

```sh
opencc config
```

v1.5 提供全新 Dashboard 介面：

- **左側清單**：Providers、Profiles、Project Bindings
- **右側詳情**：選取項目的詳細資訊
- **快捷鍵**：
  - `a` - 新增項目
  - `e` - 編輯選取項目
  - `d` - 刪除選取項目
  - `Tab` - 切換焦點
  - `q` - 返回／離開

使用 `--legacy` 切換至傳統介面。

## Web 管理介面

```sh
# 啟動（背景執行，連接埠 19840）
opencc web start

# 開啟瀏覽器
opencc web open

# 停止
opencc web stop
```

Web UI 功能：
- Provider 與 Profile 管理
- 專案綁定管理
- 全域設定（預設 CLI、預設 Profile、連接埠）
- 請求日誌檢視
- 模型欄位自動補全

## 環境變數設定

每個 provider 可以為不同 CLI 設定獨立的環境變數：

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

### Claude Code 常用環境變數

| 變數 | 說明 |
|------|------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 最大輸出 token |
| `MAX_THINKING_TOKENS` | 擴展思考預算 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 最大上下文窗口 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 預設逾時 |

## 場景路由

根據請求特徵自動路由至不同 provider：

| 場景 | 觸發條件 |
|------|---------|
| `think` | 啟用 thinking 模式 |
| `image` | 包含圖片內容 |
| `longContext` | 內容超過閾值 |
| `webSearch` | 使用 web_search 工具 |
| `background` | 使用 Haiku 模型 |

**Fallback 機制**：若場景設定的 providers 全部失敗，會自動 fallback 至 profile 的預設 providers。

組態範例：

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

## 組態檔案

| 檔案 | 說明 |
|------|------|
| `~/.opencc/opencc.json` | 主組態檔案 |
| `~/.opencc/proxy.log` | 代理日誌 |
| `~/.opencc/web.log` | Web 服務日誌 |

### 完整組態範例

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

## 升級

```sh
# 最新版本
opencc upgrade

# 指定版本
opencc upgrade 1.5
opencc upgrade 1.5.0
```

## 從舊版遷移

若先前使用 `~/.cc_envs/` 格式，opencc 會自動遷移至 `~/.opencc/opencc.json`。

## 開發

```sh
# 建置
go build -o opencc .

# 測試
go test ./...
```

發佈：推送 tag 後 GitHub Actions 自動建置。

```sh
git tag v1.5.1
git push origin v1.5.1
```

## License

MIT
