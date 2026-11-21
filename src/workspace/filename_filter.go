package workspace

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monochromegane/go-gitignore"
)

type FilenameFilter struct {
	exclude *regexp.Regexp

	// skipPaths are the absolute paths of the files and directories that
	// are always ignored (see SetSkipPaths).
	skipPaths []string

	gitignoreEnabled bool
	matchers         []gitignore.IgnoreMatcher
	initialMatchers  map[string]struct{}
}

func NewFilenameFilter(exclude *regexp.Regexp) *FilenameFilter {
	return &FilenameFilter{
		exclude:         exclude,
		initialMatchers: make(map[string]struct{}),
	}
}

// SetSkipPaths sets the files and directories that are always ignored.
//
// Paths are resolved to absolute ones (relative paths are relative to the working
// directory). A directory is ignored together with everything inside of it.
func (filter *FilenameFilter) SetSkipPaths(paths []string) {
	filter.skipPaths = filter.skipPaths[:0]
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		filter.skipPaths = append(filter.skipPaths, filepath.Clean(p))
	}
}

func (filter *FilenameFilter) isSkipped(path string) bool {
	for _, skip := range filter.skipPaths {
		if path == skip || strings.HasPrefix(path, skip+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (filter *FilenameFilter) EnableGitignore() { filter.gitignoreEnabled = true }

func (filter *FilenameFilter) GitignoreIsEnabled() bool { return filter.gitignoreEnabled }

func (filter *FilenameFilter) InitialGitignorePush(path string, matcher gitignore.IgnoreMatcher) {
	if !filter.gitignoreEnabled {
		panic("add when gitignore is disabled")
	}
	filter.initialMatchers[path] = struct{}{}
	filter.matchers = append(filter.matchers, matcher)
}

func (filter *FilenameFilter) GitignorePush(path string, matcher gitignore.IgnoreMatcher) {
	if !filter.gitignoreEnabled {
		panic("pop when gitignore is disabled")
	}
	if _, ok := filter.initialMatchers[path]; ok {
		return
	}
	filter.matchers = append(filter.matchers, matcher)
}

func (filter *FilenameFilter) GitignorePop(path string) {
	if !filter.gitignoreEnabled {
		panic("pop when gitignore is disabled")
	}
	if _, ok := filter.initialMatchers[path]; ok {
		return
	}
	filter.matchers = filter.matchers[:len(filter.matchers)-1]
}

func (filter *FilenameFilter) IgnoreFile(path string) bool {
	return filter.ignore(path, false)
}

func (filter *FilenameFilter) IgnoreDir(path string) bool {
	return filter.ignore(path, true)
}

func (filter *FilenameFilter) ignore(path string, isDir bool) bool {
	if filter.isSkipped(path) {
		return true
	}
	if filter.exclude != nil && filter.exclude.MatchString(path) {
		return true
	}
	for _, matcher := range filter.matchers {
		if matcher.Match(path, isDir) {
			return true
		}
	}
	return false
}
