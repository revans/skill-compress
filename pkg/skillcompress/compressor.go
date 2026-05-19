package skillcompress

import (
	"fmt"
	"strings"
)

const oraclePrompt = `You are a prompt compression specialist. Make the following prompt as concise as possible while preserving every behavioral instruction.

Remove:
- Hedge language ("you might want to", "it's generally good practice to", "consider doing", "it's worth noting", "remember that", "keep in mind")
- Explanations of WHY a rule exists — keep the rule, drop the rationale prose
- Restatements and summaries ("in other words", "to summarize", "as noted above", "put differently", "to reiterate")
- Redundant examples — if multiple examples demonstrate the same pattern, keep only the most illustrative one
- Preamble before the actual instruction ("Now that we've established X...", "In order to achieve X, you should...")
- Meta-commentary ("this is important because", "note that", "it should be noted that")
- Passive and indirect constructions — convert to direct ("it should be noted that X" → "X")

Keep:
- Every concrete instruction
- Every specific pattern, code example, or constraint that demonstrates something distinct
- Every edge case or exception that changes behavior
- Negative examples that show what NOT to do, when they demonstrate a non-obvious failure mode

Return only the compressed prompt text. No explanation, no wrapper, no preamble — just the compressed content.

---

`

// Compress calls Claude to produce a concise version of promptText.
// The oracle prompt instructs Claude on what to remove and what to keep.
func Compress(promptText, model string, timeoutSeconds int) (string, error) {
	full := oraclePrompt + promptText
	result, err := RunClaude(full, model, timeoutSeconds)
	if err != nil {
		return "", fmt.Errorf("compression oracle: %w", err)
	}
	return strings.TrimSpace(result.Output), nil
}
