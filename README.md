# skill-compress

Compresses any prompt file — strips hedge language, redundant examples, and meta-commentary while preserving every behavioral instruction — then validates the compressed version against your eval suite before you apply it.

Works on system prompts, agent instructions, Claude Code skill files, or any plain-text prompt. Think of it as a lossless compressor for prompts: smaller file, same behavior.

A companion to [skill-eval](https://github.com/revans/skill-eval). Both tools share the same `.skill-eval.yml` config and `evals.yml` format, so if you're already running skill-eval in a project, skill-compress plugs straight in.

## Contents

- [Getting started](#getting-started)
- [Two-step workflow](#two-step-workflow)
- [What the compressor removes and keeps](#what-the-compressor-removes-and-keeps)
- [Validation](#validation)
- [Flag reference](#flag-reference)
- [Configuration](#configuration)
- [Eval YAML reference](#eval-yaml-reference)
- [Assertion types](#assertion-types)
- [Output structure](#output-structure)
- [Exit codes](#exit-codes)

## Getting started

### 1. Install prerequisites

- **Go 1.22+** — [install](https://go.dev/dl/)
- **Claude CLI** — `skill-compress` shells out to `claude -p` for compression and validation. Install and authenticate:

```bash
npm install -g @anthropic-ai/claude-code
claude                      # follow the login prompt
claude -p "hello" --model claude-sonnet-4-6   # verify it works
```

The Claude CLI must be on your `$PATH`.

### 2. Install skill-compress

```bash
go install github.com/revans/skill-compress/cmd/skill-compress@latest
```

Or build from source:

```bash
go build -o skill-compress ./cmd/skill-compress
```

### 3. Create `.skill-eval.yml`

```yaml
default_model: claude-sonnet-4-6
```

All other fields are optional — see `.skill-eval.yml.example` for defaults. If you already have `.skill-eval.yml` from skill-eval, no changes needed.

## Two-step workflow

**Step 1 — compress and validate:**

```bash
skill-compress path/to/my-skill.md
```

This sends the skill to Claude with a compression oracle prompt, writes the compressed copy to `evals/compress/<id>/`, runs any matching evals from `evals.yml`, and prints pass/fail per test.

```
Compressing my-skill.md...

  Original:   4820 chars
  Compressed: 2941 chars  (-39%)
  Written to: evals/compress/my-skill/my-skill.md

  Validating against 3 test(s) for my-skill...

  PASS   my-skill-basic
  PASS   my-skill-edge-no-input
  FAIL   my-skill-rejects-invalid
         contains "error" — failed

  Validation failed. Review evals/compress/my-skill/diff.md before applying.
```

**Step 2 — apply (once all tests pass):**

```bash
skill-compress path/to/my-skill.md --apply
```

Overwrites the original with the compressed copy. Blocked if the last validation run failed.

## What the compressor removes and keeps

**Removes:**
- Hedge language (`"you might want to"`, `"consider doing"`, `"it's worth noting"`)
- Explanations of *why* a rule exists — keeps the rule, drops the rationale prose
- Restatements and summaries (`"in other words"`, `"to summarize"`, `"as noted above"`)
- Redundant examples — when multiple examples show the same pattern, keeps the most illustrative one
- Preamble before the actual instruction
- Meta-commentary (`"this is important because"`, `"note that"`)
- Passive and indirect constructions (converts to direct)

**Keeps:**
- Every concrete instruction
- Every specific pattern, code example, or constraint that demonstrates something distinct
- Every edge case or exception that changes behavior
- Negative examples that show what NOT to do, when they demonstrate a non-obvious failure mode

## Validation

If you have an `evals.yml` (from skill-eval or written manually), skill-compress runs any evals whose `tests` field matches the prompt's ID against the compressed version before letting you apply it.

The prompt ID is derived from the filename:
- `my-agent.md` → `my-agent`
- `prompts/summarizer.md` → `summarizer` (parent dir name used as fallback when filename is generic)
- `RU-001-params.md` → `RU-001` (uppercase prefix pattern extracted automatically)

If no evals exist for a skill, compression still runs — validation is skipped and the result is marked `unvalidated`.

## Flag reference

| Flag | Default | Description |
|------|---------|-------------|
| `--apply` | false | Promote the compressed copy over the original |
| `--config` | `.skill-eval.yml` | Path to config file |

## Configuration

| Field | Default | Description |
|-------|---------|-------------|
| `default_model` | *(required)* | Claude model for compression and validation |
| `evals_file` | `evals.yml` | Path to your evals file |
| `per_eval_timeout_seconds` | `60` | Seconds allowed per eval validation call |
| `compress_dir` | `evals/compress` | Directory for compressed output and diffs |
| `compress_timeout_seconds` | `120` | Seconds allowed for the compression call |

## Eval YAML reference

See `evals.example.yml` for a working example. The format is the same as skill-eval.

```yaml
- id: my-skill-basic
  tests: my-skill          # prompt ID — links this eval to a skill
  input: "some user input"
  assert:
    - contains: "expected output"
    - not_contains: "bad output"
```

## Assertion types

| Type | Passes when |
|------|-------------|
| `contains` | output includes the string |
| `not_contains` | output does not include the string |
| `matches` | output matches the regex |
| `not_matches` | output does not match the regex |

## Output structure

```
evals/compress/
  <prompt-id>/
    my-skill.md    # compressed copy
    diff.md        # unified diff (original → compressed)
    result.yml     # structured result
```

`result.yml` fields: `prompt_id`, `original_chars`, `compressed_chars`, `reduction_pct`, `model`, `ran_at`, `evals_matched`, `validation`, `overall` (`pass` | `fail` | `unvalidated`).

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Compression succeeded; all tests passed (or no tests found) |
| `1` | Validation failed |
| `2` | Usage error or configuration problem |

## License

MIT
