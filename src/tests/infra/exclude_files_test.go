package checkers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildNoVerify(t *testing.T) string {
	t.Helper()

	exe := filepath.Join(t.TempDir(), "noverify.exe")
	// A relative target avoids problems with modules/vendor/GOPATH.
	out, err := exec.Command("go", "build", "-o", exe, "../../../").CombinedOutput()
	if err != nil {
		t.Fatalf("build noverify: %v: %s", err, out)
	}
	return exe
}

func runNoVerify(t *testing.T, exe, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command(exe, append([]string{"check", "--disable-cache"}, args...)...)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput() // the exit code is not zero if there are issues
	return string(out)
}

func TestExcludeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("builds noverify, skipped in -short mode")
	}

	exe := buildNoVerify(t)

	dir := t.TempDir()
	files := map[string]string{
		// The incomplete stub that makes the code below look correct.
		"stubs/bootstrap.php": "<?php\nfunction stubbed_function() {}\n",
		"app/main.php":        "<?php\nstubbed_function();\nmissing_function();\n",
		"app/legacy.php":      "<?php\nlegacy_missing_function();\n",
		"app/legacy2.php":     "<?php\nlegacy_missing_function2();\n",
	}
	for name, contents := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("nothing is excluded", func(t *testing.T) {
		out := runNoVerify(t, exe, dir, "--index-only-files", "stubs", "app")
		mustContain(t, out, "Call to undefined function missing_function")
		mustContain(t, out, "Call to undefined function legacy_missing_function ")
		mustNotContain(t, out, "Call to undefined function stubbed_function")
	})

	t.Run("an excluded file is not indexed", func(t *testing.T) {
		out := runNoVerify(t, exe, dir,
			"--index-only-files", "stubs", "--exclude-files", "stubs/bootstrap.php", "app")
		mustContain(t, out, "Call to undefined function stubbed_function")
		mustContain(t, out, "Call to undefined function missing_function")
	})

	t.Run("an excluded file is not analyzed", func(t *testing.T) {
		out := runNoVerify(t, exe, dir,
			"--index-only-files", "stubs", "--exclude-files", "app/legacy.php", "app")
		mustContain(t, out, "Call to undefined function missing_function")
		mustContain(t, out, "Call to undefined function legacy_missing_function2")
		mustNotContain(t, out, "Call to undefined function legacy_missing_function ")
	})

	t.Run("several files and a directory, absolute paths", func(t *testing.T) {
		out := runNoVerify(t, exe, dir,
			"--index-only-files", "stubs",
			"--exclude-files", filepath.Join(dir, "app/legacy.php")+",app/legacy2.php,stubs",
			"app")
		mustContain(t, out, "Call to undefined function stubbed_function")
		mustNotContain(t, out, "legacy_missing_function")
	})

	t.Run("the exclude regexp still only hides reports", func(t *testing.T) {
		out := runNoVerify(t, exe, dir,
			"--index-only-files", "stubs", "--exclude", `stubs/`, "app")
		// The stub is indexed, so its function is known.
		mustNotContain(t, out, "Call to undefined function stubbed_function")
	})
}

func mustContain(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Errorf("the output must contain %q:\n%s", want, out)
	}
}

func mustNotContain(t *testing.T, out, bad string) {
	t.Helper()
	if strings.Contains(out, bad) {
		t.Errorf("the output must not contain %q:\n%s", bad, out)
	}
}
