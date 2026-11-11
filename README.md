# aigit

[中文文档 (Chinese Documentation)](./README_CN.md) | English Documentation

The most powerful git commit assistant ever!

It's a command-line tool that streamlines the git commit process by automatically generating meaningful and standardized commit messages, including title and body.

`aigit commit` is as simple as `git commit`.

## Supported 🤖 AI Providers

aigit works with any **OpenAI-compatible** API (OpenAI, DeepSeek, Qwen, Groq, Together AI, OpenRouter, vLLM, ...).

```shell
aigit auth add <api_key> <model> --base-url <base_url>
```

## Getting Started

### Installation

#### Option 1: Download Binary

1. Using `go install`:

```shell
go install github.com/YouCD/aigit@latest
```

2. Download from releases:

- Go to the [releases page](https://github.com/YouCD/aigit/releases) and download the binary for your platform.
- Rename the binary to `aigit` and move it to `/usr/local/bin/aigit`.

```shell
chmod +x aigit && sudo mv aigit /usr/local/bin/aigit
```

#### Option 2: Build from Source

```shell
git clone https://github.com/YouCD/aigit.git
cd aigit
go build -o aigit .
sudo mv aigit /usr/local/bin/aigit
```

### Generate commit message

Run `aigit commit` in a repository with staged changes:

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

- If nothing is staged, aigit asks whether to run `git add .` first.
- After committing, aigit asks whether to push to the remote.

Useful flags:

```shell
aigit commit -y       # skip all confirmations and commit directly
aigit commit -l zh    # generate the message in Chinese (en|zh, per-run only)
aigit commit -t 120   # LLM request timeout in seconds (per-run only)
```

### Configure your AI provider

```shell
$ aigit auth add sk-xxx deepseek-chat --base-url https://api.deepseek.com/v1
Successfully added API key for openai-compatible
```

## Configuration

The config file lives at `~/.config/aigit/config.json`:

| Field | Description | Default |
|---|---|---|
| `language` | Commit message language (`en` / `zh`) | auto-detected from `LANG` |
| `emoji` | Include emojis in the commit message | `true` |
| `timeout` | LLM request timeout in seconds | `300` |

The message language can also be overridden per run with `aigit commit -l en|zh`.

## Update check

aigit checks for new releases in the background (at most once every 24 hours, cached
in `~/.config/aigit/update-check.json`). When a newer version is available, it asks after the
command finishes whether to upgrade; if you agree, aigit builds the new release from
source and upgrades itself in place. Set `AIGIT_NO_UPDATE_CHECK=1` to disable the check.
