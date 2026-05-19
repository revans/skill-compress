package skillcompress

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// CompressResult is the structured output written to result.yml.
type CompressResult struct {
	SubstrateID     string                 `yaml:"substrate_id"`
	OriginalFile    string                 `yaml:"original_file"`
	CompressedFile  string                 `yaml:"compressed_file"`
	OriginalChars   int                    `yaml:"original_chars"`
	CompressedChars int                    `yaml:"compressed_chars"`
	ReductionPct    int                    `yaml:"reduction_pct"`
	Model           string                 `yaml:"model"`
	RanAt           string                 `yaml:"ran_at"`
	EvalsMatched    int                    `yaml:"evals_matched"`
	Validation      []evalValidationRecord `yaml:"validation,omitempty"`
	Overall         string                 `yaml:"overall"` // "pass", "fail", "unvalidated"
}

type evalValidationRecord struct {
	EvalID     string            `yaml:"eval_id"`
	Status     string            `yaml:"status"`
	Assertions []assertionRecord `yaml:"assertions"`
}

type assertionRecord struct {
	Type   string `yaml:"type"`
	Value  string `yaml:"value"`
	Result string `yaml:"result"`
}

// WorkKeyFromPath derives the key used for the compress working directory.
// For substrate rule files (e.g. RU-001-params-expect.md) it extracts the ID prefix.
// For other files it falls back to the parent directory name, then the filename stem.
func WorkKeyFromPath(path string) string {
	base := filepath.Base(path)
	if m := regexp.MustCompile(`^([A-Z]+-\d+)`).FindString(base); m != "" {
		return m
	}
	if parent := filepath.Base(filepath.Dir(path)); parent != "." && parent != "/" {
		return parent
	}
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}

// CompressWorkDir returns the working directory for a substrate ID under baseDir.
func CompressWorkDir(baseDir, substrateID string) string {
	return filepath.Join(baseDir, substrateID)
}

// WriteCompressedCopy writes compressedText to workDir/{basename of originalPath}.
// Returns the path to the written file.
func WriteCompressedCopy(workDir, originalPath, compressedText string) (string, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("creating compress dir: %w", err)
	}
	dest := filepath.Join(workDir, filepath.Base(originalPath))
	if err := os.WriteFile(dest, []byte(compressedText), 0644); err != nil {
		return "", fmt.Errorf("writing compressed copy: %w", err)
	}
	return dest, nil
}

// WriteDiff shells out to `diff -u` and writes the result to workDir/diff.md.
// Errors from diff (exit 1 when files differ) are expected and ignored.
func WriteDiff(workDir, originalPath, compressedPath string) error {
	out, _ := exec.Command("diff", "-u", originalPath, compressedPath).Output()
	content := "```diff\n" + string(out) + "```\n"
	return os.WriteFile(filepath.Join(workDir, "diff.md"), []byte(content), 0644)
}

// WriteResult serializes result to workDir/result.yml.
func WriteResult(workDir string, result CompressResult) error {
	data, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshaling result: %w", err)
	}
	return os.WriteFile(filepath.Join(workDir, "result.yml"), data, 0644)
}

// ReadResult deserializes result.yml from workDir.
func ReadResult(workDir string) (CompressResult, error) {
	var r CompressResult
	data, err := os.ReadFile(filepath.Join(workDir, "result.yml"))
	if err != nil {
		return r, err
	}
	return r, yaml.Unmarshal(data, &r)
}

// Promote overwrites originalPath with the contents of compressedPath.
func Promote(originalPath, compressedPath string) error {
	data, err := os.ReadFile(compressedPath)
	if err != nil {
		return fmt.Errorf("reading compressed copy: %w", err)
	}
	return os.WriteFile(originalPath, data, 0644)
}

// BuildResult constructs a CompressResult from compression and validation outputs.
func BuildResult(substrateID, originalPath, compressedPath, model,
	originalText, compressedText string,
	evals []Eval, validationResults []EvalValidationResult, allPassed bool) CompressResult {

	origChars := len(originalText)
	compChars := len(compressedText)
	reductionPct := 0
	if origChars > 0 {
		reductionPct = int(float64(origChars-compChars) / float64(origChars) * 100)
	}

	overall := "unvalidated"
	if len(evals) > 0 {
		if allPassed {
			overall = "pass"
		} else {
			overall = "fail"
		}
	}

	var records []evalValidationRecord
	for _, vr := range validationResults {
		status := "pass"
		if vr.Err != nil {
			status = "error"
		} else if !vr.Passed {
			status = "fail"
		}
		var assertions []assertionRecord
		for _, ar := range vr.Assertions {
			result := "pass"
			if !ar.Passed {
				result = "fail"
			}
			assertions = append(assertions, assertionRecord{Type: ar.Type, Value: ar.Value, Result: result})
		}
		records = append(records, evalValidationRecord{EvalID: vr.ID, Status: status, Assertions: assertions})
	}

	return CompressResult{
		SubstrateID:     substrateID,
		OriginalFile:    originalPath,
		CompressedFile:  compressedPath,
		OriginalChars:   origChars,
		CompressedChars: compChars,
		ReductionPct:    reductionPct,
		Model:           model,
		RanAt:           time.Now().UTC().Format(time.RFC3339),
		EvalsMatched:    len(evals),
		Validation:      records,
		Overall:         overall,
	}
}
