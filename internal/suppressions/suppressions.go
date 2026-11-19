// Package suppressions manages per-file, per-rule error counts for the file CLI.
package suppressions

import (
	"context"
	"os"
	"path/filepath"
	"slices"

	"github.com/web-infra-dev/rslint/internal/rule"
)

type Mode uint8

const (
	Check Mode = iota
	Record
	Prune
)

type Options struct {
	Location string
	Mode     Mode
	Rules    []string // In Record mode nil selects all rules and an empty list selects none.
}

type entry struct {
	Count int `json:"count"`
}

type entries map[string]map[string]entry

type File struct {
	cwd     string
	path    string
	options Options
	entries entries
}

type checkedFile struct {
	key    string
	counts map[string]int
}

// Apply processes a completed lint result. checked includes clean files and
// excludes files whose rules could not run. Diagnostic paths are the same
// caller-visible paths used to project checked files.
func (f *File) Apply(ctx context.Context, checked []string, diagnostics []rule.RuleDiagnostic) ([]rule.RuleDiagnostic, bool, error) {
	size := len(checked)
	if f.options.Mode == Check {
		size = min(size, len(f.entries))
	}
	files := make(map[string]*checkedFile, size)
	for _, path := range checked {
		key, err := filepath.Rel(f.cwd, filepath.FromSlash(path))
		if err != nil {
			// On Windows, a target on another volume has no relative path.
			key = filepath.FromSlash(path)
		}
		key = filepath.ToSlash(key)
		if f.options.Mode != Check || len(f.entries[key]) > 0 {
			files[path] = &checkedFile{key: key}
		}
	}
	for _, diagnostic := range diagnostics {
		file := files[diagnostic.FilePath]
		if file == nil || !eligible(diagnostic) {
			continue
		}
		if file.counts == nil {
			file.counts = make(map[string]int)
		}
		file.counts[diagnostic.RuleName]++
	}
	if f.options.Mode != Check {
		if err := f.update(ctx, files); err != nil {
			return diagnostics, false, err
		}
	}

	unused := false
	for _, file := range files {
		for id, recorded := range f.entries[file.key] {
			if file.counts[id] < recorded.Count {
				unused = true
			}
		}
	}
	visible := slices.DeleteFunc(diagnostics, func(diagnostic rule.RuleDiagnostic) bool {
		file := files[diagnostic.FilePath]
		return file != nil && eligible(diagnostic) &&
			file.counts[diagnostic.RuleName] <= f.entries[file.key][diagnostic.RuleName].Count
	})
	return visible, unused, nil
}

func eligible(d rule.RuleDiagnostic) bool {
	return d.Origin == rule.DiagnosticOriginLint && d.Severity == rule.SeverityError && d.RuleName != ""
}

func (f *File) record(data entries, files map[string]*checkedFile) {
	for _, file := range files {
		for id, count := range file.counts {
			if f.options.Rules != nil && !slices.Contains(f.options.Rules, id) {
				continue
			}
			if data[file.key] == nil {
				data[file.key] = make(map[string]entry)
			}
			data[file.key][id] = entry{Count: count}
		}
	}
}

func (f *File) prune(data entries, files map[string]*checkedFile) error {
	for _, file := range files {
		rules := data[file.key]
		for id, recorded := range rules {
			switch count := file.counts[id]; {
			case count == 0:
				delete(rules, id)
			case count < recorded.Count:
				rules[id] = entry{Count: count}
			}
		}
		if len(rules) == 0 {
			delete(data, file.key)
		}
	}
	for key := range data {
		path := filepath.FromSlash(key)
		if !filepath.IsAbs(path) {
			path = filepath.Join(f.cwd, path)
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			delete(data, key)
		} else if err != nil {
			return err
		}
	}
	return nil
}
