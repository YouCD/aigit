package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/openai/openai-go"
	openaioption "github.com/openai/openai-go/option"
)

const ProviderOpenAICompatible = "openai-compatible"

// Debug controls whether debug information is printed to stderr.
var Debug bool

const englishPrompt = `Analyze the git diff and output ONLY a valid JSON object with these fields:
- "type": commit type from the list below
- "scope": affected component (optional, empty string if none)
- "description": short imperative description (max 72 chars)
- "body": list of strings for detailed explanation, each bullet point as a separate string

Type selection (by priority):
BREAKING CHANGE: API breaking changes
fix: Bug fixes, crashes, errors, security issues
feat: New features, APIs, capabilities
perf: Performance improvements
refactor: Code restructuring without functional changes
docs: Documentation only
style: Formatting, whitespace, imports
test: Test changes
build: Build system, dependencies
ci: CI/CD changes
chore: Maintenance, version updates

Rules:
- Body should explain WHAT and WHY, not HOW
- Each line should be less than 72 characters

Example: {"type":"feat","scope":"auth","description":"add user login endpoint","body":["implement JWT token generation","add password hashing"]}

IMPORTANT: Output ONLY the JSON object. No other text.

Diff:`

const chinesePrompt = `分析 git diff，只输出一个有效的 JSON 对象，包含以下字段：
- "type": 提交类型，从下方列表选择
- "scope": 影响范围（可选，没有则填空字符串）
- "description": 简短的描述
- "body": 字符串数组，每条占一行

类型选择（按优先级）：
BREAKING CHANGE: API 不兼容变更
fix: 修复 bug、崩溃、错误、安全问题
feat: 新功能、API、能力
perf: 性能改进
refactor: 代码重构，无功能变化
docs: 仅文档变更
style: 格式化、空白、导入
test: 测试变更
build: 构建系统、依赖
ci: CI/CD 变更
chore: 维护、版本更新

规则：
- body 说明 WHAT 和 WHY，不是 HOW
- 每条不超过 72 个字符
- description 和 body 必须用中文

示例：{"type":"feat","scope":"auth","description":"添加用户登录接口","body":["实现 JWT token 生成","添加密码哈希中间件"]}

重要：只输出 JSON 对象，不要其他文字。

Diff:`

func getPrompt(language string, _ bool) string {
	if language == "zh" {
		return chinesePrompt
	}
	return englishPrompt
}

type CommitData struct {
	Type        string   `json:"type"`
	Scope       string   `json:"scope"`
	Description string   `json:"description"`
	Body        []string `json:"body"`
}

func emojiForType(t string) string {
	switch t {
	case "feat":
		return "✨"
	case "fix":
		return "🐛"
	case "docs":
		return "📝"
	case "style":
		return "🎨"
	case "refactor":
		return "♻️"
	case "perf":
		return "⚡️"
	case "test":
		return "✅"
	case "build":
		return "🔧"
	case "ci":
		return "🚀"
	case "chore":
		return "🔖"
	default:
		return ""
	}
}

const commitTemplateText = `{{if .Emoji}}{{.Emoji}} {{end}}{{.Type}}{{if .Scope}}({{.Scope}}){{end}}: {{.Description}}{{if .Body}}

{{.Body}}{{end}}`

type commitTemplateData struct {
	Type        string
	Scope       string
	Description string
	Body        string
	Emoji       string
}

func renderCommitMessage(data *CommitData, useEmoji bool) string {
	body := ""
	if len(data.Body) > 0 {
		for i, line := range data.Body {
			data.Body[i] = "- " + line
		}
		body = strings.Join(data.Body, "\n")
	}
	tmplData := commitTemplateData{
		Type:        data.Type,
		Scope:       data.Scope,
		Description: data.Description,
		Body:        body,
	}
	if useEmoji {
		tmplData.Emoji = emojiForType(data.Type)
	}
	tmpl := template.Must(template.New("commit").Parse(commitTemplateText))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return data.Description
	}
	return buf.String()
}

func parseAndRenderCommit(raw string, useEmoji bool) (string, error) {
	data, _ := extractCommitData(raw)
	if data == nil {
		return cleanCommitResponse(raw), nil
	}
	return renderCommitMessage(data, useEmoji), nil
}

func IsRawCommitJSON(s string) bool {
	return looksLikeCommitJSON(s)
}

// cleanCommitResponse removes thinking-tag markers, markdown code fences and
// the "json" language prefix, returning a trimmed response. Tag markers are
// removed while keeping their content, so a JSON payload the model places
// inside a thinking block survives intact.
func cleanCommitResponse(raw string) string {
	s := strings.ReplaceAll(raw, "<thinking>", "")
	s = strings.ReplaceAll(s, "</thinking>", "")
	if idx := strings.Index(s, "```"); idx >= 0 {
		s = s[idx+3:]
		if endIdx := strings.LastIndex(s, "```"); endIdx >= 0 {
			s = s[:endIdx]
		}
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "json")
	return strings.TrimSpace(s)
}

// tryParseCommitJSON decodes s and requires the two mandatory fields.
func tryParseCommitJSON(s string) (*CommitData, error) {
	var data CommitData
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return nil, err
	}
	if data.Type == "" || data.Description == "" {
		return nil, fmt.Errorf("missing required fields")
	}
	return &data, nil
}

// balancedJSONSpan returns the substring of s beginning at the '{' at start
// and ending at its matching '}', honouring nested objects and braces inside
// string literals.
func balancedJSONSpan(s string, start int) (string, bool) {
	depth, inStr, escape := 0, false, false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case escape:
				escape = false
			case c == '\\':
				escape = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

func extractCommitData(raw string) (*CommitData, error) {
	// The prompt asks for a bare JSON object, so try the cleaned response as a
	// whole first (covers plain objects and fenced blocks).
	if data, err := tryParseCommitJSON(cleanCommitResponse(raw)); err == nil {
		return data, nil
	}
	// Some providers wrap the JSON in thinking tags or prose; recover the
	// outermost JSON object embedded anywhere in the response.
	start := 0
	for {
		i := strings.IndexByte(raw[start:], '{')
		if i < 0 {
			break
		}
		i += start
		if span, ok := balancedJSONSpan(raw, i); ok {
			if data, err := tryParseCommitJSON(span); err == nil {
				return data, nil
			}
		}
		start = i + 1
	}
	return nil, fmt.Errorf("no commit JSON found in response")
}

func looksLikeCommitJSON(s string) bool {
	_, err := extractCommitData(s)
	return err == nil
}

var _ MessageGenerator = (*OpenAICompatibleGenerator)(nil)

// MessageGenerator Define a commit message generator
type MessageGenerator interface {
	GenerateCommitMessage(diff string) (string, error)
}

// OpenAICompatibleGenerator Implementation for OpenAI-compatible providers (e.g. Groq, Together AI, OpenRouter)
type OpenAICompatibleGenerator struct {
	apiKey   string
	model    string
	baseURL  string
	language string
	emoji    bool
	timeout  int
}

func NewOpenAICompatibleGenerator(apiKey, model, baseURL, language string, emoji bool, timeout int) *OpenAICompatibleGenerator {
	return &OpenAICompatibleGenerator{apiKey: apiKey, model: model, baseURL: baseURL, language: language, emoji: emoji, timeout: timeout}
}

func (g *OpenAICompatibleGenerator) GenerateCommitMessage(diff string) (string, error) {
	return generateOpenAICompatibleCommitMessage(diff, g.apiKey, g.model, g.baseURL, g.language, g.emoji, g.timeout)
}

func generateOpenAICompatibleCommitMessage(diff, apiKey, model, baseURL, language string, emoji bool, timeout int) (string, error) {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	client := openai.NewClient(
		openaioption.WithAPIKey(apiKey),
		openaioption.WithBaseURL(baseURL),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	prompt := fmt.Sprintf("%s\n%s", getPrompt(language, emoji), diff)

	if Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] === LLM Request ===\n")
		fmt.Fprintf(os.Stderr, "[DEBUG] Provider: openai-compatible\n")
		fmt.Fprintf(os.Stderr, "[DEBUG] Model: %s\n", model)
		fmt.Fprintf(os.Stderr, "[DEBUG] Base URL: %s\n", baseURL)
		fmt.Fprintf(os.Stderr, "[DEBUG] Prompt total length: %d chars\n", len(prompt))
		if len(prompt) > 5000 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Prompt (first 5000 chars):\n%s\n\n... (truncated, full length=%d)\n", prompt[:5000], len(prompt))
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] Prompt:\n%s\n", prompt)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] === End LLM Request ===\n")
	}

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: model,
	})
	if err != nil {
		return "", fmt.Errorf("generating commit message: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		return parseAndRenderCommit(strings.TrimSpace(chatCompletion.Choices[0].Message.Content), emoji)
	}

	return "", fmt.Errorf("no commit message generated by OpenAI-compatible provider")
}
