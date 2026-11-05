---
id: llm-commit-response-parsing
type: module-design
title: llm commit-message response parsing
status: draft
---

# llm commit-message response parsing

The `llm` package turns a provider's raw chat reply into a commit message. This spec
covers the response-parsing contract in `model.go` — how `parseAndRenderCommit` and
`IsRawCommitJSON` interpret whatever the model returns.

## Contract

- **A valid reply is the commit JSON object** with mandatory `type` and `description`
  (plus optional `scope` and `body`). `extractCommitData` accepts a reply iff it
  contains such an object; `parseAndRenderCommit` then renders it through the commit
  template, and `IsRawCommitJSON` reports `true` for the same condition. Both share one
  extractor so they cannot drift apart.
- **Wrappers are tolerated, not required.** Providers may wrap the JSON in markdown code
  fences, prose, or thinking-tag markers (`<thinking>` / `</thinking>`). The parser
  recovers the JSON object from any of these shapes rather than demanding a bare object.
- **Extraction is two-pass.** First the whole cleaned response (markers and fences
  removed, content kept) is tried as a JSON object. If that fails, a balanced-brace scan
  recovers the outermost JSON object embedded anywhere in the reply. Only when no object
  parses does `parseAndRenderCommit` fall back to returning the cleaned prose as the
  message.
- **Thinking-only replies carry no message.** A reply containing only thinking content
  and no JSON object is not commit data: `IsRawCommitJSON` is `false` and the rendered
  message is the tag-free prose — a degenerate model failure, not a parseable commit.

## Decisions

- **Recover the object instead of stripping blocks.** The first fix attempt deleted
  complete `


` blocks; that broke the dominant real shape where the model emits the JSON
  *inside* its thinking block (opening tag, reasoning, JSON, closing tag), discarding
  the payload. Embedded extraction works for JSON-inside-tags, JSON-after-tags, and
  plain objects alike, so block deletion was dropped.
- **Markers are removed, their content kept.** `cleanCommitResponse` deletes the
  `<thinking>` / `</thinking>` marker strings but leaves the text between them, so an
  in-block JSON survives the whole-response pass and the tag-free prose fallback loses
  no payload.
- **`IsRawCommitJSON` means "contains a commit JSON object",** i.e. the reply still
  needs rendering. Callers in `main.go` use it to detect an unrendered raw reply and
  warn; it must therefore agree with `extractCommitData` on every input, which the
  shared extractor guarantees.
