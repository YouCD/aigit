package llm

import "testing"

const commitJSON = `{"type":"feat","scope":"cli","description":"Add commit command","body":["Adds a commit subcommand"]}`

const wantRendered = "✨ feat(cli): Add commit command\n\n- Adds a commit subcommand"

// openTag/closeTag model the thinking-tag wrappers some providers emit around
// the JSON payload. Built via concatenation so the test source carries no
// literal newlines inside a double-quoted string.
var (
	openTag  = "<" + "thinking>"
	closeTag = "<" + "/thinking>"
)

func TestParseAndRenderCommitWithThinkingTags(t *testing.T) {
	cases := map[string]string{
		"closed tags":    "\n" + openTag + "\nanalysing diff\n" + closeTag + "\n" + commitJSON + "\n" + closeTag + "\n",
		"unclosed tag":   "\n" + openTag + "\nanalysing diff\n" + commitJSON,
		"tags and fence": "\n" + openTag + "\nreasoning\n" + closeTag + "\n```json\n" + commitJSON + "\n```\n" + closeTag + "\n",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := parseAndRenderCommit(raw, true)
			if err != nil {
				t.Fatalf("parseAndRenderCommit(%q) error: %v", raw, err)
			}
			if got != wantRendered {
				t.Errorf("parseAndRenderCommit:\n got: %q\nwant: %q", got, wantRendered)
			}
		})
	}
}

func TestIsRawCommitJSONWithThinkingTags(t *testing.T) {
	cases := map[string]struct {
		in   string
		want bool
	}{
		"closed tags":   {"\n" + openTag + "\nreasoning\n" + closeTag + "\n" + commitJSON + "\n" + closeTag + "\n", true},
		"unclosed tag":  {"\n" + openTag + "\nreasoning\n" + commitJSON, true},
		"plain json":    {commitJSON, true},
		"prose only":    {"Here is the commit message:", false},
		"thinking only": {"\n" + openTag + "\nlet me think\n" + closeTag + "\n", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsRawCommitJSON(tc.in); got != tc.want {
				t.Errorf("IsRawCommitJSON(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseAndRenderCommitStillHandlesFences(t *testing.T) {
	raw := "```json\n" + commitJSON + "\n```"
	got, err := parseAndRenderCommit(raw, true)
	if err != nil {
		t.Fatalf("parseAndRenderCommit(%q) error: %v", raw, err)
	}
	if got != wantRendered {
		t.Errorf("parseAndRenderCommit:\n got: %q\nwant: %q", got, wantRendered)
	}
}
