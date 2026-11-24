// Package suppressions manages per-file, per-rule error counts for the file CLI.
package suppressions

import (
	"context"
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
	cwd     string // Physical working directory.
	path    string
	options Options
	entries entries
}

// FileResult identifies a linted target. A completed rule with no diagnostics
// has a count of zero.
type FileResult struct {
	Path         string // Absolute, slash-separated filename of the lint result.
	SkippedRules []string
	Incomplete   bool // Some results for this file are unavailable.
}

type checkedFile struct {
	key             string
	counts          map[string]int
	skippedRules    []string
	preserveEntries bool
}

// Apply checks the final lint results against the baseline, updating counts in
// record or prune mode. Clean files must be included in checked.
func (f *File) Apply(ctx context.Context, checked []FileResult, diagnostics []rule.RuleDiagnostic) ([]rule.RuleDiagnostic, bool, error) {
	files := make(map[string]*checkedFile)
	// Named recording needs only the selected rule's result. Updating the
	// entire file requires a complete file result.
	recordNamedRules := f.options.Mode == Record && f.options.Rules != nil
	// An empty, successful target set can still create a baseline or prune deleted files.
	canUpdate := len(checked) == 0
	for _, result := range checked {
		key := f.fileKey(result.Path)
		if f.options.Mode == Check && f.entries[key] == nil {
			continue
		}
		preserve := result.Incomplete && !recordNamedRules
		files[result.Path] = &checkedFile{
			key:             key,
			skippedRules:    result.SkippedRules,
			preserveEntries: preserve,
		}
		canUpdate = canUpdate || !preserve
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
	if f.options.Mode != Check && canUpdate {
		err := f.update(ctx, func(latest entries) bool {
			if f.options.Mode == Record {
				return f.record(latest, files)
			}
			return f.prune(latest, files)
		})
		if err != nil {
			return diagnostics, false, err
		}
	}

	unused := false
	for _, file := range files {
		for id, recorded := range f.entries[file.key] {
			if file.counts[id] < recorded.Count && !slices.Contains(file.skippedRules, id) {
				unused = true
			}
		}
	}
	visible := slices.DeleteFunc(diagnostics, func(diagnostic rule.RuleDiagnostic) bool {
		file := files[diagnostic.FilePath]
		if file == nil || !eligible(diagnostic) {
			return false
		}
		count := file.counts[diagnostic.RuleName]
		return count != 0 && count <= f.entries[file.key][diagnostic.RuleName].Count
	})
	return visible, unused, nil
}

func eligible(d rule.RuleDiagnostic) bool {
	return d.Origin == rule.DiagnosticOriginLint && d.Severity == rule.SeverityError && d.RuleName != ""
}

func (f *File) record(data entries, files map[string]*checkedFile) bool {
	changed := false
	for _, file := range files {
		if file.preserveEntries {
			continue
		}
		for id, count := range file.counts {
			if data[file.key][id].Count == count || f.options.Rules != nil && !slices.Contains(f.options.Rules, id) {
				continue
			}
			if data[file.key] == nil {
				data[file.key] = make(map[string]entry)
			}
			data[file.key][id] = entry{Count: count}
			changed = true
		}
	}
	return changed
}

func (f *File) prune(data entries, files map[string]*checkedFile) bool {
	changed := false
	preserved := make(map[string]bool)
	for _, file := range files {
		if file.preserveEntries {
			preserved[file.key] = true
			continue
		}
		rules := data[file.key]
		for id, recorded := range rules {
			count := file.counts[id]
			if count >= recorded.Count || slices.Contains(file.skippedRules, id) {
				continue
			}
			changed = true
			if count == 0 {
				delete(rules, id)
			} else {
				rules[id] = entry{Count: count}
			}
		}
		if rules != nil && len(rules) == 0 {
			delete(data, file.key)
			changed = true
		}
	}
	for key := range data {
		if !preserved[key] && f.fileMissing(key) {
			delete(data, key)
			changed = true
		}
	}
	return changed
}
