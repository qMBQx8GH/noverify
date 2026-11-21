package checkers

import (
	"os"
	"path/filepath"
	"testing"
)

// The layout is like in Bitrix: the config is in the module, next to the phpstan config,
// and noverify is run from the directory of the module.
func TestConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("builds noverify, skipped in -short mode")
	}

	exe := buildNoVerify(t)

	modules := t.TempDir()
	files := map[string]string{
		"main/lib/core.php": `<?php
/** */
class CoreHelper {
  /** */
  public static function help(): int { return 1; }
}
`,
		"main/lib/bootstrap.php": "<?php\nfunction stub_only() {}\n", // the incomplete stub
		"mod/lib/a.php":          "<?php\necho CoreHelper::help();\necho stub_only();\necho undefined_in_a();\nexit(1);\n",
		"mod/lib/legacy.php":     "<?php\necho undefined_in_legacy();\n",
		"mod/other/b.php":        "<?php\necho undefined_in_b();\n",
		"other/o.php":            "<?php\necho undefined_in_other();\n",
		"mod/dev/phpat/noverify.yaml": `# the paths are relative to this file
paths:
  - ../../lib
include:
  - ../../../main
exclude-files:
  - ../../lib/legacy.php
  - ../../../main/lib/bootstrap.php
exclude-checks:
  - useExitOrDie
`,
		"mod/dev/phpat/empty.yaml": "# no paths\nexclude-checks:\n  - unused\n",
		"mod/dev/phpat/typo.yaml":  "pths:\n  - ../../lib\n",
		"mod/dev/phpat/bad.yaml":   "paths:\n  - ../../missing\n",
	}
	for name, contents := range files {
		path := filepath.Join(modules, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	moduleDir := filepath.Join(modules, "mod")

	t.Run("from the module directory, like phpstan -c", func(t *testing.T) {
		out := runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/noverify.yaml")

		mustContain(t, out, "undefined function undefined_in_a")
		// `include` is indexed, but it is not analyzed.
		mustNotContain(t, out, "core.php")
		mustNotContain(t, out, "Call to undefined function help")
		mustNotContain(t, out, "CoreHelper")
		// `exclude-files` are not indexed and not analyzed.
		mustContain(t, out, "undefined function stub_only")
		mustNotContain(t, out, "undefined_in_legacy")
		// `exclude-checks`.
		mustNotContain(t, out, "useExitOrDie")
		// `paths` don't contain mod/other.
		mustNotContain(t, out, "undefined_in_b")
	})

	t.Run("the same with --config and from another directory", func(t *testing.T) {
		out := runNoVerify(t, exe, t.TempDir(), "--config", filepath.Join(moduleDir, "dev/phpat/noverify.yaml"))

		mustContain(t, out, "undefined function undefined_in_a")
		mustContain(t, out, "undefined function stub_only")
		mustNotContain(t, out, "undefined_in_legacy")
		mustNotContain(t, out, "useExitOrDie")
	})

	t.Run("the paths of the command line replace the paths of the config", func(t *testing.T) {
		out := runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/noverify.yaml", "other")

		mustContain(t, out, "undefined function undefined_in_b")
		mustNotContain(t, out, "undefined_in_a")
		// The rest of the config still works.
		mustNotContain(t, out, "useExitOrDie")
	})

	t.Run("the flags are merged with the config", func(t *testing.T) {
		out := runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/noverify.yaml",
			"--exclude-checks", "undefinedFunction", "--exclude-files", "lib/a.php")

		mustNotContain(t, out, "undefinedFunction")
		mustNotContain(t, out, "undefined_in_a")
		mustNotContain(t, out, "useExitOrDie") // from the config
	})

	t.Run("without the config the same paths give the other result", func(t *testing.T) {
		out := runNoVerify(t, exe, moduleDir, "lib")

		mustContain(t, out, "undefined_in_legacy")
		mustContain(t, out, "Call to undefined method \\CoreHelper::help()")
		mustContain(t, out, "undefined function stub_only")
		mustContain(t, out, "useExitOrDie")
	})

	t.Run("the paths are required", func(t *testing.T) {
		out := runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/empty.yaml")
		mustContain(t, out, "no paths to analyze")
		mustNotContain(t, out, "Indexing")

		out = runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/empty.yaml", "lib")
		mustContain(t, out, "undefined_in_a")
	})

	t.Run("the errors of the config", func(t *testing.T) {
		out := runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/typo.yaml")
		mustContain(t, out, "field pths not found")

		out = runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/bad.yaml")
		mustContain(t, out, "paths:")
		mustContain(t, out, "missing")

		out = runNoVerify(t, exe, moduleDir, "-c", "dev/phpat/no_such_config.yaml")
		mustContain(t, out, "read config")
	})
}
