// Package suppressions manages per-file, per-rule error counts for the file CLI.
package suppressions

import (
	"context"
	"slices"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	baseDir     string // Native process cwd, the origin of all serialized paths.
	physicalCwd string
	cwdAliases  []utils.DirectoryIdentity
	path        string
	options     Options
	entries     entries
}

// FileResult identifies a selected path. A completed rule with no diagnostics
// has a count of zero. An unchecked path only reserves its baseline keys.
type FileResult struct {
	Path         string
	InputPaths   []string // Spellings of CLI operands that selected this lint result.
	SkippedRules []string
	Unchecked    bool
}

type checkedFile struct {
	keys         []string
	literal      string
	portable     string
	counts       map[string]int
	skippedRules []string
}

// Apply compares diagnostics with the baseline for the given files, including
// those that were checked and produced no diagnostics.
func (f *File) Apply(ctx context.Context, checked []FileResult, diagnostics []rule.RuleDiagnostic) ([]rule.RuleDiagnostic, bool, error) {
	files := f.matchFiles(checked)
	if f.options.Mode == Check {
		selectExistingKeys(f.entries, files)
	}
	for _, diagnostic := range diagnostics {
		file := files[diagnostic.FilePath]
		if file == nil || !eligible(diagnostic) || slices.Contains(file.skippedRules, diagnostic.RuleName) {
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
		for _, key := range file.keys {
			for id, recorded := range f.entries[key] {
				if file.counts[id] < recorded.Count && !slices.Contains(file.skippedRules, id) {
					unused = true
				}
			}
		}
	}
	visible := slices.DeleteFunc(diagnostics, func(diagnostic rule.RuleDiagnostic) bool {
		file := files[diagnostic.FilePath]
		if file == nil || !eligible(diagnostic) {
			return false
		}
		count := file.counts[diagnostic.RuleName]
		if count == 0 {
			return false
		}
		for _, key := range file.keys {
			if count > f.entries[key][diagnostic.RuleName].Count {
				return false
			}
		}
		return true
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
			for _, key := range file.keys {
				if data[key] == nil {
					data[key] = make(map[string]entry)
				}
				data[key][id] = entry{Count: count}
			}
		}
	}
}

func (f *File) prune(data entries, files map[string]*checkedFile) error {
	for _, file := range files {
		for _, key := range file.keys {
			rules := data[key]
			for id, recorded := range rules {
				if slices.Contains(file.skippedRules, id) {
					continue
				}
				switch count := file.counts[id]; {
				case count == 0:
					delete(rules, id)
				case count < recorded.Count:
					rules[id] = entry{Count: count}
				}
			}
			if len(rules) == 0 {
				delete(data, key)
			}
		}
	}
	for key := range data {
		missing, err := f.fileMissing(key)
		if err != nil {
			return err
		}
		if missing {
			delete(data, key)
		}
	}
	return nil
}
