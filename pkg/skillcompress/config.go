package skillcompress

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds configuration loaded from .skill-eval.yml.
// skill-compress shares the same config file as skill-eval; compress-specific
// fields (compress_dir, compress_timeout_seconds) are ignored by skill-eval.
type Config struct {
	DefaultModel           string `yaml:"default_model"`
	EvalsFile              string `yaml:"evals_file"`
	PerEvalTimeoutSeconds  int    `yaml:"per_eval_timeout_seconds"`
	CompressDir            string `yaml:"compress_dir"`
	CompressTimeoutSeconds int    `yaml:"compress_timeout_seconds"`
}

// LoadConfig reads .skill-eval.yml at path and applies defaults for missing fields.
func LoadConfig(path string) (Config, error) {
	cfg := Config{
		EvalsFile:              "evals.yml",
		PerEvalTimeoutSeconds:  60,
		CompressDir:            "evals/compress",
		CompressTimeoutSeconds: 120,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return cfg, fmt.Errorf("reading config %q: %w", path, err)
		}
	} else {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parsing config %q: %w", path, err)
		}
	}

	if cfg.DefaultModel == "" {
		return cfg, fmt.Errorf("default_model is required in .skill-eval.yml")
	}
	return cfg, nil
}
