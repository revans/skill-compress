package skillcompress

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Eval holds the fields from evals.yml needed for compression validation.
type Eval struct {
	ID     string
	Tests  string
	Input  string
	Assert []Assertion
}

// Assertion is a single assertion entry from an eval's assert list.
type Assertion struct {
	Type  string
	Value string
}

type rawEval struct {
	ID     string              `yaml:"id"`
	Tests  string              `yaml:"tests"`
	Input  string              `yaml:"input"`
	Assert []map[string]string `yaml:"assert"`
}

// LoadEvalsForSubstrate reads evals.yml and returns all evals where tests == substrateID.
// Returns nil (not an error) when the evals file does not exist.
func LoadEvalsForSubstrate(evalsFile, substrateID string) ([]Eval, error) {
	data, err := os.ReadFile(evalsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading evals file %q: %w", evalsFile, err)
	}

	var raws []rawEval
	if err := yaml.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("parsing evals file %q: %w", evalsFile, err)
	}

	var matched []Eval
	for _, r := range raws {
		if r.Tests != substrateID {
			continue
		}
		assertions := make([]Assertion, 0, len(r.Assert))
		for _, raw := range r.Assert {
			for k, v := range raw {
				assertions = append(assertions, Assertion{Type: k, Value: v})
				break
			}
		}
		matched = append(matched, Eval{
			ID:     r.ID,
			Tests:  r.Tests,
			Input:  r.Input,
			Assert: assertions,
		})
	}
	return matched, nil
}
