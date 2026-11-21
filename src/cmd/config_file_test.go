package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newConfigTestTree creates modules/{main,mod/lib,mod/dev/phpat} and returns modules.
func newConfigTestTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"main", "mod/lib", "mod/dev/phpat"} {
		if err := os.MkdirAll(filepath.Join(root, "modules", dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "modules")
}

func TestLoadFileConfig(t *testing.T) {
	modules := newConfigTestTree(t)
	config := filepath.Join(modules, "mod/dev/phpat/noverify.yaml")
	writeTestFile(t, config, `
# comments are allowed
paths:
  - ../../lib
  - ..
include:
  - ../../../main
exclude-files:
  - ../../lib/legacy.php
  - ../../../main/cli/bootstrap.php
  - /absolute/path/stub.php
  - ""
exclude-checks:
  - useExitOrDie
  - notSafeCall
`)

	have, err := LoadFileConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	want := &FileConfig{
		Paths:   []string{filepath.Join(modules, "mod/lib"), filepath.Join(modules, "mod/dev")},
		Include: []string{filepath.Join(modules, "main")},
		ExcludeFiles: []string{
			filepath.Join(modules, "mod/lib/legacy.php"),
			filepath.Join(modules, "main/cli/bootstrap.php"),
			filepath.Clean("/absolute/path/stub.php"),
		},
		ExcludeChecks: []string{"useExitOrDie", "notSafeCall"},
		dir:           filepath.Join(modules, "mod/dev/phpat"),
	}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("config mismatch:\nhave %+v\nwant %+v", have, want)
	}
}

func TestLoadFileConfigRelativeToTheFileNotToTheWorkingDir(t *testing.T) {
	modules := newConfigTestTree(t)
	config := filepath.Join(modules, "mod/dev/phpat/noverify.yaml")
	writeTestFile(t, config, "paths:\n  - ../../lib\n")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	// The config is passed with a path relative to the working directory.
	if err := os.Chdir(filepath.Join(modules, "mod")); err != nil {
		t.Fatal(err)
	}
	have, err := LoadFileConfig("dev/phpat/noverify.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{filepath.Join(modules, "mod/lib")}; !reflect.DeepEqual(have.Paths, want) {
		t.Errorf("have %v, want %v", have.Paths, want)
	}

	if err := os.Chdir(modules); err != nil {
		t.Fatal(err)
	}
	have, err = LoadFileConfig("mod/dev/phpat/noverify.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{filepath.Join(modules, "mod/lib")}; !reflect.DeepEqual(have.Paths, want) {
		t.Errorf("have %v, want %v", have.Paths, want)
	}
}

func TestLoadFileConfigEmpty(t *testing.T) {
	modules := newConfigTestTree(t)

	for name, contents := range map[string]string{
		"empty file":    "",
		"only comments": "# nothing\n",
		"empty lists":   "paths: []\ninclude:\nexclude-files: []\n",
	} {
		config := filepath.Join(modules, "mod/dev/phpat/empty.yaml")
		writeTestFile(t, config, contents)

		have, err := LoadFileConfig(config)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(have.Paths)+len(have.Include)+len(have.ExcludeFiles)+len(have.ExcludeChecks) != 0 {
			t.Errorf("%s: the config must be empty: %+v", name, have)
		}
	}
}

func TestLoadFileConfigErrors(t *testing.T) {
	modules := newConfigTestTree(t)
	config := filepath.Join(modules, "mod/dev/phpat/noverify.yaml")

	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{"unknown key", "path:\n  - ../../lib\n", "field path not found"},
		{"misspelled key", "exclude-file:\n  - a.php\n", "field exclude-file not found"},
		{"a string instead of a list", "paths: ../../lib\n", "cannot unmarshal"},
		{"a map instead of a list", "include:\n  a: b\n", "cannot unmarshal"},
		{"bad yaml", "paths: [../../lib\n", "yaml"},
		{"path does not exist", "paths:\n  - ../../missing\n", "paths:"},
		{"included path does not exist", "paths:\n  - ../../lib\ninclude:\n  - ../../../missing\n", "include:"},
	}

	for _, test := range tests {
		writeTestFile(t, config, test.contents)

		_, err := LoadFileConfig(config)
		if err == nil {
			t.Errorf("%s: error expected", test.name)
			continue
		}
		if !strings.Contains(err.Error(), test.want) {
			t.Errorf("%s: error %q must contain %q", test.name, err, test.want)
		}
	}

	if _, err := LoadFileConfig(filepath.Join(modules, "missing.yaml")); err == nil {
		t.Errorf("the missing config file must be an error")
	}

	// An excluded path may not exist.
	writeTestFile(t, config, "paths:\n  - ../../lib\nexclude-files:\n  - ../../lib/missing.php\n")
	if _, err := LoadFileConfig(config); err != nil {
		t.Errorf("an excluded path may not exist: %v", err)
	}
}

func TestFileConfigApply(t *testing.T) {
	config := &FileConfig{
		Paths:         []string{"/m/mod/lib"},
		Include:       []string{"/m/main", "/m/rest"},
		ExcludeFiles:  []string{"/m/main/cli/bootstrap.php"},
		ExcludeChecks: []string{"useExitOrDie", "notSafeCall"},
	}

	t.Run("an empty command line", func(t *testing.T) {
		flags := &ParsedFlags{}
		paths := config.apply(flags, nil)

		if !reflect.DeepEqual(paths, []string{"/m/mod/lib"}) {
			t.Errorf("paths: %v", paths)
		}
		if flags.IndexOnlyFiles != "/m/main,/m/rest" {
			t.Errorf("IndexOnlyFiles: %q", flags.IndexOnlyFiles)
		}
		if flags.ExcludeFiles != "/m/main/cli/bootstrap.php" {
			t.Errorf("ExcludeFiles: %q", flags.ExcludeFiles)
		}
		if flags.ReportsExcludeChecks != "useExitOrDie,notSafeCall" {
			t.Errorf("ReportsExcludeChecks: %q", flags.ReportsExcludeChecks)
		}
	})

	t.Run("the command line is merged, the paths replace the paths of the config", func(t *testing.T) {
		flags := &ParsedFlags{
			IndexOnlyFiles:       "vendor",
			ExcludeFiles:         "a.php",
			ReportsExcludeChecks: "unused",
		}
		paths := config.apply(flags, []string{"other/dir"})

		if !reflect.DeepEqual(paths, []string{"other/dir"}) {
			t.Errorf("paths: %v", paths)
		}
		if flags.IndexOnlyFiles != "vendor,/m/main,/m/rest" {
			t.Errorf("IndexOnlyFiles: %q", flags.IndexOnlyFiles)
		}
		if flags.ExcludeFiles != "a.php,/m/main/cli/bootstrap.php" {
			t.Errorf("ExcludeFiles: %q", flags.ExcludeFiles)
		}
		if flags.ReportsExcludeChecks != "unused,useExitOrDie,notSafeCall" {
			t.Errorf("ReportsExcludeChecks: %q", flags.ReportsExcludeChecks)
		}
	})

	t.Run("an empty config keeps the flags", func(t *testing.T) {
		flags := &ParsedFlags{IndexOnlyFiles: "vendor"}
		paths := (&FileConfig{}).apply(flags, nil)

		if len(paths) != 0 || flags.IndexOnlyFiles != "vendor" || flags.ExcludeFiles != "" || flags.ReportsExcludeChecks != "" {
			t.Errorf("paths %v, flags %+v", paths, flags)
		}
	})
}

func TestApplyFileConfigWithoutPaths(t *testing.T) {
	modules := newConfigTestTree(t)
	config := filepath.Join(modules, "mod/dev/phpat/noverify.yaml")
	writeTestFile(t, config, "exclude-checks:\n  - unused\n")

	ctx := &AppContext{ParsedFlags: ParsedFlags{ConfigFile: config}}
	err := applyFileConfig(ctx)
	if err == nil || !strings.Contains(err.Error(), "no paths to analyze") {
		t.Fatalf("error expected, have %v", err)
	}

	// The paths from the command line are enough.
	ctx = &AppContext{ParsedFlags: ParsedFlags{ConfigFile: config}, ParsedArgs: []string{"src"}}
	if err := applyFileConfig(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ctx.ParsedFlags.ReportsExcludeChecks != "unused" {
		t.Errorf("ReportsExcludeChecks: %q", ctx.ParsedFlags.ReportsExcludeChecks)
	}

	// Without the config file nothing changes.
	ctx = &AppContext{}
	if err := applyFileConfig(ctx); err != nil || len(ctx.ParsedArgs) != 0 {
		t.Errorf("no config file: %v %v", err, ctx.ParsedArgs)
	}
}
