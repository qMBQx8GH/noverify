package workspace

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("<?php\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readAll(t *testing.T, root string, filter *FilenameFilter) []string {
	t.Helper()

	ch := make(chan FileInfo, 64)
	ReadFilenames([]string{root}, filter, []string{"php"})(ch)
	close(ch)

	var names []string
	for f := range ch {
		rel, err := filepath.Rel(root, f.Name)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, filepath.ToSlash(rel))
	}
	sort.Strings(names)
	return names
}

func TestFilenameFilterSkipPaths(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{
		"main/cli/bootstrap.php",
		"main/cli/bootstrap.php.bak.php",
		"main/cli/bootstrap2.php",
		"main/cli/other.php",
		"main/dev/cli/annotations/bootstrap.php",
		"main/dev/cli/annotations/other.php",
		"main/lib/a.php",
		"crm/lib/b.php",
	} {
		writeFile(t, filepath.Join(root, f))
	}

	tests := []struct {
		name string
		skip []string
		want []string
	}{
		{
			name: "no skip paths",
			want: []string{
				"crm/lib/b.php",
				"main/cli/bootstrap.php",
				"main/cli/bootstrap.php.bak.php",
				"main/cli/bootstrap2.php",
				"main/cli/other.php",
				"main/dev/cli/annotations/bootstrap.php",
				"main/dev/cli/annotations/other.php",
				"main/lib/a.php",
			},
		},
		{
			name: "individual files, the names that only start with the same text stay",
			skip: []string{
				filepath.Join(root, "main/cli/bootstrap.php"),
				filepath.Join(root, "main/dev/cli/annotations/bootstrap.php"),
			},
			want: []string{
				"crm/lib/b.php",
				"main/cli/bootstrap.php.bak.php",
				"main/cli/bootstrap2.php",
				"main/cli/other.php",
				"main/dev/cli/annotations/other.php",
				"main/lib/a.php",
			},
		},
		{
			name: "directory is skipped with its content",
			skip: []string{filepath.Join(root, "main/dev")},
			want: []string{
				"crm/lib/b.php",
				"main/cli/bootstrap.php",
				"main/cli/bootstrap.php.bak.php",
				"main/cli/bootstrap2.php",
				"main/cli/other.php",
				"main/lib/a.php",
			},
		},
		{
			name: "directory with a trailing slash and spaces around",
			skip: []string{" " + filepath.Join(root, "main/cli") + "/ ", ""},
			want: []string{
				"crm/lib/b.php",
				"main/dev/cli/annotations/bootstrap.php",
				"main/dev/cli/annotations/other.php",
				"main/lib/a.php",
			},
		},
		{
			name: "a name prefix of a directory is not a directory",
			skip: []string{filepath.Join(root, "main/c")},
			want: []string{
				"crm/lib/b.php",
				"main/cli/bootstrap.php",
				"main/cli/bootstrap.php.bak.php",
				"main/cli/bootstrap2.php",
				"main/cli/other.php",
				"main/dev/cli/annotations/bootstrap.php",
				"main/dev/cli/annotations/other.php",
				"main/lib/a.php",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := NewFilenameFilter(nil)
			filter.SetSkipPaths(test.skip)

			have := readAll(t, root, filter)
			if len(have) != len(test.want) {
				t.Fatalf("files mismatch:\nhave %v\nwant %v", have, test.want)
			}
			for i := range have {
				if have[i] != test.want[i] {
					t.Fatalf("files mismatch:\nhave %v\nwant %v", have, test.want)
				}
			}
		})
	}
}

func TestFilenameFilterSkipPathsRelativeToWorkingDir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main/cli/bootstrap.php"))
	writeFile(t, filepath.Join(root, "main/lib/a.php"))

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	filter := NewFilenameFilter(nil)
	filter.SetSkipPaths([]string{"main/cli/bootstrap.php"})

	have := readAll(t, root, filter)
	if len(have) != 1 || have[0] != "main/lib/a.php" {
		t.Fatalf("have %v, want [main/lib/a.php]", have)
	}
}

func TestFilenameFilterSkipPathsAndExcludeRegexp(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.php"))
	writeFile(t, filepath.Join(root, "b.php"))
	writeFile(t, filepath.Join(root, "c.php"))

	filter := NewFilenameFilter(regexp.MustCompile(`b\.php$`))
	filter.SetSkipPaths([]string{filepath.Join(root, "a.php")})

	have := readAll(t, root, filter)
	if len(have) != 1 || have[0] != "c.php" {
		t.Fatalf("have %v, want [c.php]", have)
	}
}
