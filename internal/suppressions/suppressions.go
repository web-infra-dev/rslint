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

// SelectsRule reports whether a rule is included by the requested rule filter.
func (o Options) SelectsRule(name string) bool {
	return o.Rules == nil || slices.Contains(o.Rules, name)
}

type entry struct {
	Count int `json:"count"`
}

type entries map[string]map[string]entry

type targetAlias struct {
	selected, other string
	mayRecord       bool
}

type File struct {
	cwd           string // Physical working directory, slash-separated and ending in a slash.
	cwdAlias      string // Working directory used for target selection when it differs from cwd.
	path          string
	options       Options
	entries       entries
	targetAliases []targetAlias
	protected     map[string]error
	pending       *stagedUpdate
}

// NoteTargetAlias records two selected filenames for which lint will run only once.
// mayRecord reports whether either enables an error rule selected for recording.
func (f *File) NoteTargetAlias(selected, other string, mayRecord bool) {
	alias := targetAlias{selected: f.fileKey(selected), other: f.fileKey(other), mayRecord: mayRecord}
	if alias.selected != alias.other {
		f.targetAliases = append(f.targetAliases, alias)
	}
}

// CheckTargetAliases reports when recording or verifying counts requires both filenames.
func (f *File) CheckTargetAliases() error {
	return f.checkTargetAliases(f.entries)
}

// PreserveCounts reserves the current counts for a target whose eligibility
// could not be verified. A read or an update leaving those counts intact is valid.
func (f *File) PreserveCounts(path string, reason error) {
	if f.protected == nil {
		f.protected = make(map[string]error)
	}
	f.protected[f.fileKey(path)] = reason
}

func (f *File) checkTargetAliases(data entries) error {
	for _, alias := range f.targetAliases {
		if alias.mayRecord || len(data[alias.selected]) != 0 || len(data[alias.other]) != 0 {
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

// Prepare compares diagnostics with the latest counts and stages any changes.
// The caller releases an update with CommitUpdate or DiscardUpdate.
func (f *File) Prepare(ctx context.Context, checked []FileResult, diagnostics []rule.RuleDiagnostic) ([]rule.RuleDiagnostic, bool, error) {
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
	counts := f.entries
	var aliasErr error
	if f.options.Mode == Check {
		aliasErr = f.CheckTargetAliases()
	} else {
		if err := f.prepareUpdate(ctx, files); err != nil {
			f.DiscardUpdate()
			return diagnostics, false, err
		}
		counts = f.pending.entries
	}

	unused := false
	for _, file := range files {
		for id, recorded := range counts[file.key] {
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
		return count != 0 && count <= counts[file.key][diagnostic.RuleName].Count
	})
	return visible, unused, aliasErr
}

func eligible(d rule.RuleDiagnostic) bool {
	return d.Origin == rule.DiagnosticOriginLint && d.Severity == rule.SeverityError && d.RuleName != ""
}

func (f *File) record(data entries, files map[string]*checkedFile) (bool, error) {
	changed := false
	for _, file := range files {
		for id, count := range file.counts {
			if !f.options.SelectsRule(id) {
				continue
			}
			if data[file.key][id].Count == count {
				continue
			}
			if err := f.protected[file.key]; err != nil {
				return false, err
			}
			if data[file.key] == nil {
				data[file.key] = make(map[string]entry)
			}
			data[file.key][id] = entry{Count: count}
			changed = true
		}
	}
	return changed, nil
}

func (f *File) prune(data entries, files map[string]*checkedFile) (bool, error) {
	changed := false
	for _, file := range files {
		rules := data[file.key]
		for id, recorded := range rules {
			if slices.Contains(file.skippedRules, id) {
				continue
			}
			count := file.counts[id]
			if count >= recorded.Count {
				continue
			}
			if err := f.protected[file.key]; err != nil {
				return false, err
			}
			changed = true
			if count == 0 {
				delete(rules, id)
			} else {
				rules[id] = entry{Count: count}
			}
		}
		if rules != nil && len(rules) == 0 && f.protected[file.key] == nil {
			delete(data, file.key)
			changed = true
		}
	}
	for key := range data {
		missing, err := f.fileMissing(key)
		if err != nil {
			return false, err
		}
		if missing {
			if err := f.protected[key]; err != nil {
				return false, err
			}
			delete(data, key)
			changed = true
		}
	}
	return changed, nil
}
