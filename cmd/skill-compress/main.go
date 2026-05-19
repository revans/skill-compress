package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/revans/skill-compress/pkg/skillcompress"
)

func main() {
	apply := flag.Bool("apply", false, "overwrite the original file with the compressed copy")
	config := flag.String("config", ".skill-eval.yml", "path to config file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: skill-compress <path> [--apply] [--config <path>]\n\n")
		fmt.Fprintf(os.Stderr, "  Compresses a skill prompt and validates it against matching tests.\n")
		fmt.Fprintf(os.Stderr, "  The compressed copy is written to evals/compress/{id}/.\n\n")
		fmt.Fprintf(os.Stderr, "  Run without --apply to compress and validate.\n")
		fmt.Fprintf(os.Stderr, "  Run with --apply to promote the compressed copy over the original.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	os.Exit(run(flag.Arg(0), *apply, *config))
}

func run(originalPath string, apply bool, configPath string) int {
	cfg, err := skillcompress.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if apply {
		return runApply(originalPath, cfg)
	}
	return runCompress(originalPath, cfg)
}

func runCompress(originalPath string, cfg skillcompress.Config) int {
	workKey := skillcompress.WorkKeyFromPath(originalPath)

	originalBytes, err := os.ReadFile(originalPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading %q: %v\n", originalPath, err)
		return 2
	}
	originalText := string(originalBytes)

	fmt.Printf("\nCompressing %s...\n", filepath.Base(originalPath))

	compressedText, err := skillcompress.Compress(originalText, cfg.DefaultModel, cfg.CompressTimeoutSeconds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	workDir := skillcompress.CompressWorkDir(cfg.CompressDir, workKey)

	compressedPath, err := skillcompress.WriteCompressedCopy(workDir, originalPath, compressedText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if err := skillcompress.WriteDiff(workDir, originalPath, compressedPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write diff: %v\n", err)
	}

	origChars := len(originalText)
	compChars := len(compressedText)
	reductionPct := 0
	if origChars > 0 {
		reductionPct = int(float64(origChars-compChars) / float64(origChars) * 100)
	}

	fmt.Printf("\n  Original:   %d chars\n", origChars)
	fmt.Printf("  Compressed: %d chars  (-%d%%)\n", compChars, reductionPct)
	fmt.Printf("  Written to: %s\n", compressedPath)

	matched, err := skillcompress.LoadEvalsFor(cfg.EvalsFile, workKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load test suite: %v\n", err)
	}

	if len(matched) == 0 {
		fmt.Printf("\n  No tests found for %s — skipping validation.\n", workKey)
		fmt.Printf("  Review %s before applying.\n", compressedPath)

		result := skillcompress.BuildResult(workKey, originalPath, compressedPath,
			cfg.DefaultModel, originalText, compressedText, nil, nil, false)
		if err := skillcompress.WriteResult(workDir, result); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not write result: %v\n", err)
		}
		return 0
	}

	fmt.Printf("\n  Validating against %d test(s) for %s...\n\n", len(matched), workKey)

	validationResults, allPassed := skillcompress.ValidateAll(
		matched, compressedText, cfg.DefaultModel, cfg.PerEvalTimeoutSeconds)

	for _, vr := range validationResults {
		if vr.Err != nil {
			fmt.Printf("  ERROR  %s  (%v)\n", vr.ID, vr.Err)
		} else if vr.Passed {
			fmt.Printf("  PASS   %s\n", vr.ID)
		} else {
			fmt.Printf("  FAIL   %s\n", vr.ID)
			for _, ar := range vr.Assertions {
				if !ar.Passed {
					fmt.Printf("         %s %q — failed\n", ar.Type, ar.Value)
				}
			}
		}
	}

	result := skillcompress.BuildResult(workKey, originalPath, compressedPath,
		cfg.DefaultModel, originalText, compressedText, matched, validationResults, allPassed)
	if err := skillcompress.WriteResult(workDir, result); err != nil {
		fmt.Fprintf(os.Stderr, "\nwarning: could not write result: %v\n", err)
	}

	diffPath := filepath.Join(workDir, "diff.md")
	if allPassed {
		fmt.Printf("\n  All tests pass. Run with --apply to promote.\n")
		fmt.Printf("  Diff: %s\n", diffPath)
		return 0
	}

	fmt.Printf("\n  Validation failed. Review %s before applying.\n", diffPath)
	return 1
}

func runApply(originalPath string, cfg skillcompress.Config) int {
	workKey := skillcompress.WorkKeyFromPath(originalPath)

	workDir := skillcompress.CompressWorkDir(cfg.CompressDir, workKey)
	compressedPath := filepath.Join(workDir, filepath.Base(originalPath))

	if _, err := os.Stat(compressedPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: no compressed copy at %s\n", compressedPath)
		fmt.Fprintf(os.Stderr, "       run skill-compress %s first\n", originalPath)
		return 2
	}

	prior, err := skillcompress.ReadResult(workDir)
	if err == nil && prior.Overall == "fail" {
		fmt.Fprintf(os.Stderr, "error: last validation failed — review and re-run before applying\n")
		fmt.Fprintf(os.Stderr, "       %s\n", filepath.Join(workDir, "result.yml"))
		return 1
	}

	if err := skillcompress.Promote(originalPath, compressedPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	fmt.Printf("\nApplied: %s\n", originalPath)
	return 0
}
