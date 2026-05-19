package skillcompress

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Init creates .skill-eval.yml if it does not exist, or adds default_model
// to it if the file exists but the field is missing.
func Init(configPath, model string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading %q: %w", configPath, err)
		}
		return createConfig(configPath, model)
	}

	var existing map[string]interface{}
	if err := yaml.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("parsing %q: %w", configPath, err)
	}

	if v, ok := existing["default_model"]; ok && v != "" {
		fmt.Printf("default_model already set to %q in %s — nothing to do.\n", v, configPath)
		return nil
	}

	return appendDefaultModel(configPath, data, model)
}

func createConfig(path, model string) error {
	content := fmt.Sprintf(`default_model: %s

# evals_file: evals.yml
# per_eval_timeout_seconds: 60
# compress_dir: evals/compress
# compress_timeout_seconds: 120
`, model)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("Created %s with default_model: %s\n", path, model)
	return nil
}

func appendDefaultModel(path string, existing []byte, model string) error {
	line := fmt.Sprintf("default_model: %s\n", model)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		line = "\n" + line
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %q: %w", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		return err
	}
	fmt.Printf("Added default_model: %s to %s\n", model, path)
	return nil
}
