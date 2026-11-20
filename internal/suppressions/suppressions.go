// Package suppressions manages per-file, per-rule error counts for the file CLI.
package suppressions

import (
	"context"
	"fmt"
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

type targetAlias struct {
	selected, other string
}

type File struct {
	cwd           string // Physical working directory, slash-separated and ending in a slash.
	path          string
	options       Options
	entries       entries
	targetAliases []targetAlias
}

// NoteTargetAlias records two selected filenames for which lint will run only once.
func (f *File) NoteTargetAlias(selected, other string) {
	alias := targetAlias{selected: f.fileKey(selected), other: f.fileKey(other)}
	if alias.selected != alias.other {
		f.targetAliases = append(f.targetAliases, alias)
	}
}

// CheckTargetAliases reports when recording or verifying counts requires both filenames.
func (f *File) CheckTargetAliases() error {
	return f.checkTargetAliases(f.entries)
}

func (f *File) checkTargetAliases(data entries) error {
	recordingRules := f.options.Mode == Record && (f.options.Rules == nil || len(f.options.Rules) > 0)
	for _, alias := range f.targetAliases {
		if recordingRules || len(data[alias.selected]) != 0 || len(data[alias.other]) != 0 {
			return fmt.Errorf("cannot verify suppressions for %q and %q together because they refer to the same file. Run rslint on each path separately or exclude one path", alias.selected, alias.other)
		}
	}
	return nil
}

// FileResult identifies a linted target. A completed rule with no diagnostics
// has a count of zero.
type FileResult struct {
	Path         string // Absolute, slash-separated filename of the lint result.
	SkippedRules []string
}

type checkedFile struct {
	key          string
	counts       map[string]int
	skippedRules []string
}

// Apply compares diagnostics with the baseline for the given files, including
// those that were checked and produced no diagnostics.
func (f *File) Apply(ctx context.Context, checked []FileResult, diagnostics []rule.RuleDiagnostic) ([]rule.RuleDiagnostic, bool, error) {
	files := make(map[string]*checkedFile)
	for _, result := range checked {
		key := f.fileKey(result.Path)
		if f.options.Mode == Check && f.entries[key] == nil {
			continue
		}
		files[result.Path] = &checkedFile{key: key, skippedRules: result.SkippedRules}
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
	var aliasErr error
	if f.options.Mode == Check {
		aliasErr = f.CheckTargetAliases()
	} else if err := f.update(ctx, files); err != nil {
		return diagnostics, false, err
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
	return visible, unused, aliasErr
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
			delete(data, file.key)
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
