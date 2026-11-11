# aigit

[English Documentation](README.md) | 中文文档

最强大的 Git 提交助手！

一个命令行工具，通过自动生成有意义、标准化的提交信息（包含标题和正文）来简化 git 提交流程。

`aigit commit` 就像 `git commit` 一样简单。

## 支持的大模型/AI

aigit 支持任意 **OpenAI 兼容** 的 API（OpenAI、DeepSeek、Qwen、Groq、Together AI、OpenRouter、vLLM 等）。

```shell
aigit auth add <api_key> <model> --base-url <base_url>
```

## 快速开始

### 安装

#### 选项 1：下载二进制文件

1. 使用 `go install`：

```shell
go install github.com/YouCD/aigit@latest
```

2. 从发布页面下载：

- 前往 [发布页面](https://github.com/YouCD/aigit/releases) 下载适合您平台的二进制文件。
- 将二进制文件重命名为 `aigit` 并移动到 `/usr/local/bin/aigit`。

```shell
chmod +x aigit && sudo mv aigit /usr/local/bin/aigit
```

#### 选项 2：从源码构建

```shell
git clone https://github.com/YouCD/aigit.git
cd aigit
go build -o aigit .
sudo mv aigit /usr/local/bin/aigit
```

### 生成提交信息

在有暂存变更的仓库中运行 `aigit commit`：

```shell
$ aigit commit

🤖 Generating commit message by openai-compatible

📝 Generated commit message:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
feat(llm): add support for volcengine-go-sdk

This commit adds support for the volcengine-go-sdk for integrating with
Doubao LLM service.

The following changes were made:

- Provider type and APIKey field were added to the llm.Config struct.
- generateDoubaoCommitMessage function was updated to use the volcengine-go-sdk.
- The client is initialized with the apiKey and endpointId.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
? Choose an action ❯ Commit this message  Regenerate message
```

- 如果没有暂存的变更，aigit 会询问是否先执行 `git add .`。
- 提交完成后，aigit 会询问是否推送到远程仓库。

常用参数：

```shell
aigit commit -y       # 跳过所有确认，直接提交
aigit commit -l zh    # 生成中文提交信息（en|zh，仅当次生效）
aigit commit -t 120   # LLM 请求超时秒数（仅当次生效）
```

### 配置 AI 提供商

```shell
$ aigit auth add sk-xxx deepseek-chat --base-url https://api.deepseek.com/v1
Successfully added API key for openai-compatible
```

## 配置

配置文件位于 `~/.config/aigit/config.json`：

| 字段 | 说明 | 默认值 |
|---|---|---|
| `language` | 提交信息语言（`en` / `zh`） | 根据 `LANG` 环境变量自动检测 |
| `emoji` | 提交信息中是否包含 emoji | `true` |
| `timeout` | LLM 请求超时时间（秒） | `300` |

也可以通过 `aigit commit -l en|zh` 在单次运行时临时覆盖语言。

## 版本更新检查

aigit 会在后台检查新版本（每 24 小时最多一次，结果缓存在 `~/.config/aigit/update-check.json`）。
当有新版本可用时，会在命令执行结束后询问是否升级；如果您同意，aigit 会从源码构建新版本并原地替换。
设置 `AIGIT_NO_UPDATE_CHECK=1` 可以禁用该功能。
