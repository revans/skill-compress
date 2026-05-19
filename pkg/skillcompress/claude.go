package skillcompress

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ClaudeResult holds the captured output and duration from a claude -p invocation.
type ClaudeResult struct {
	Output     string
	DurationMs int64
}

// RunClaude invokes claude -p with the given prompt and retries once on transient failure.
func RunClaude(prompt, model string, timeoutSeconds int) (ClaudeResult, error) {
	result, err := runClaudeOnce(prompt, model, timeoutSeconds)
	if err != nil {
		time.Sleep(2 * time.Second)
		result, err = runClaudeOnce(prompt, model, timeoutSeconds)
	}
	return result, err
}

func runClaudeOnce(prompt, model string, timeoutSeconds int) (ClaudeResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "-p", "--model", model)
	cmd.Stdin = strings.NewReader(prompt)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()
	out, err := cmd.Output()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return ClaudeResult{}, fmt.Errorf("timeout after %ds", timeoutSeconds)
	}
	if err != nil {
		if msg := stderr.String(); msg != "" {
			return ClaudeResult{}, fmt.Errorf("claude subprocess: %w; stderr: %s", err, msg)
		}
		return ClaudeResult{}, fmt.Errorf("claude subprocess: %w", err)
	}

	return ClaudeResult{
		Output:     string(out),
		DurationMs: duration.Milliseconds(),
	}, nil
}
