package skillcompress

import (
	"regexp"
	"strings"
)

// AssertionResult holds the outcome of a single assertion check.
type AssertionResult struct {
	Type   string
	Value  string
	Passed bool
}

// EvalValidationResult holds the outcome of validating one test against the compressed prompt.
type EvalValidationResult struct {
	ID         string
	Assertions []AssertionResult
	Passed     bool
	Err        error
}

// ValidateEval runs a single eval against the compressed prompt and returns the result.
func ValidateEval(eval Eval, compressedPrompt, model string, timeoutSeconds int) EvalValidationResult {
	res := EvalValidationResult{ID: eval.ID}

	cr, err := RunClaude(compressedPrompt+"\n\n"+eval.Input, model, timeoutSeconds)
	if err != nil {
		res.Err = err
		return res
	}

	allPassed := true
	for _, a := range eval.Assert {
		ar := checkAssertion(a, cr.Output)
		res.Assertions = append(res.Assertions, ar)
		if !ar.Passed {
			allPassed = false
		}
	}
	res.Passed = allPassed
	return res
}

// ValidateAll runs every eval against the compressed prompt and returns results plus an overall pass.
func ValidateAll(evals []Eval, compressedPrompt, model string, timeoutSeconds int) ([]EvalValidationResult, bool) {
	results := make([]EvalValidationResult, len(evals))
	allPassed := true
	for i, e := range evals {
		results[i] = ValidateEval(e, compressedPrompt, model, timeoutSeconds)
		if results[i].Err != nil || !results[i].Passed {
			allPassed = false
		}
	}
	return results, allPassed
}

func checkAssertion(a Assertion, output string) AssertionResult {
	ar := AssertionResult{Type: a.Type, Value: a.Value}
	switch a.Type {
	case "contains":
		ar.Passed = strings.Contains(output, a.Value)
	case "not_contains":
		ar.Passed = !strings.Contains(output, a.Value)
	case "matches":
		matched, err := regexp.MatchString(a.Value, output)
		ar.Passed = err == nil && matched
	case "not_matches":
		matched, err := regexp.MatchString(a.Value, output)
		ar.Passed = err == nil && !matched
	}
	return ar
}
