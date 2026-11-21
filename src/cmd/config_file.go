package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileConfig is the content of the config file passed with --config (-c).
//
//	# noverify.yaml
//	paths:
//	  - ../..
//	include:
//	  - ../../../main
//	exclude-files:
//	  - ../../../main/cli/bootstrap.php
//	exclude-checks:
//	  - useExitOrDie
//
// The relative paths are relative to the directory of the config file,
// so the config works the same way no matter where noverify is run from.
type FileConfig struct {
	// Paths are the files and directories to analyze.
	// The paths from the command line replace them.
	Paths []string `yaml:"paths"`

	// Include are the files and directories that are only indexed (not analyzed),
	// the functions and classes from them are used for the type inference.
	// It is the same as --index-only-files.
	Include []string `yaml:"include"`

	// ExcludeFiles are the files and directories that are excluded from both
	// the indexing and the analysis. It is the same as --exclude-files.
	ExcludeFiles []string `yaml:"exclude-files"`

	// ExcludeChecks are the names of the checks to disable.
	// It is the same as --exclude-checks.
	ExcludeChecks []string `yaml:"exclude-checks"`

	// dir is the directory of the config file, the base of the relative paths.
	dir string
}

// LoadFileConfig reads and validates the config file.
//
// The relative paths of the result are made absolute (relative to the config file).
// It is an error if the file has an unknown key, or if a path from `paths` or
// `include` does not exist.
func LoadFileConfig(filename string) (*FileConfig, error) {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("config %s: %v", filename, err)
	}

	data, err := os.ReadFile(absFilename)
	if err != nil {
		return nil, fmt.Errorf("read config: %v", err)
	}

	var config FileConfig

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)                                                  // typos in the keys must not be silently ignored
	if err := decoder.Decode(&config); err != nil && !errors.Is(err, io.EOF) { // io.EOF is an empty file
		return nil, fmt.Errorf("config %s: %v", filename, err)
	}

	config.dir = filepath.Dir(absFilename)

	config.Paths = config.resolve(config.Paths)
	config.Include = config.resolve(config.Include)
	config.ExcludeFiles = config.resolve(config.ExcludeFiles)

	for _, key := range []struct {
		name  string
		paths []string
	}{
		{"paths", config.Paths},
		{"include", config.Include},
	} {
		for _, p := range key.paths {
			if _, err := os.Stat(p); err != nil {
				return nil, fmt.Errorf("config %s: %s: %v", filename, key.name, err)
			}
		}
	}

	return &config, nil
}

// resolve makes the relative paths absolute (relative to the directory of the config).
func (c *FileConfig) resolve(paths []string) []string {
	var result []string
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(c.dir, p)
		}
		result = append(result, filepath.Clean(p))
	}
	return result
}

// apply adds the values of the config to the flags and returns the paths to analyze.
//
// The lists are merged with the flags of the command line, but the paths
// from the command line (args) replace the paths of the config.
func (c *FileConfig) apply(flags *ParsedFlags, args []string) []string {
	flags.IndexOnlyFiles = joinList(flags.IndexOnlyFiles, c.Include)
	flags.ExcludeFiles = joinList(flags.ExcludeFiles, c.ExcludeFiles)
	flags.ReportsExcludeChecks = joinList(flags.ReportsExcludeChecks, c.ExcludeChecks)

	if len(args) != 0 {
		return args
	}
	return c.Paths
}

// joinList appends the items to the comma-separated list.
func joinList(list string, items []string) string {
	if len(items) == 0 {
		return list
	}
	all := make([]string, 0, len(items)+1)
	if list != "" {
		all = append(all, list)
	}
	all = append(all, items...)
	return strings.Join(all, ",")
}

// applyFileConfig loads the config file passed with --config and applies it.
func applyFileConfig(ctx *AppContext) error {
	if ctx.ParsedFlags.ConfigFile == "" {
		return nil
	}

	config, err := LoadFileConfig(ctx.ParsedFlags.ConfigFile)
	if err != nil {
		return err
	}

	ctx.ParsedArgs = config.apply(&ctx.ParsedFlags, ctx.ParsedArgs)

	if len(ctx.ParsedArgs) == 0 && ctx.ParsedFlags.GitRepo == "" {
		return fmt.Errorf("config %s: no paths to analyze, add `paths` to the config or pass them in the command line", ctx.ParsedFlags.ConfigFile)
	}

	return nil
}
