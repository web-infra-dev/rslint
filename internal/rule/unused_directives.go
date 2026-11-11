package rule

import (
	"fmt"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

type directiveEntry struct {
	rangeInSource core.TextRange
	name          string
	used          bool
	invalid       bool
}

type trackedDirective struct {
	listStart, listEnd int
	kind               directiveKind
	location           core.TextRange
	line               int
	sourceLine, column int
	entries            []directiveEntry
}

// DisableDirectiveUsage identifies a rule entry by its original comment's
// zero-based ECMAScript line and UTF-16 column. Nil RuleID means a wildcard.
type DisableDirectiveUsage struct {
	Line   int     `json:"line"`
	Column int     `json:"column"`
	RuleID *string `json:"ruleId"`
}

type directiveUsageKey struct {
	line, column int
	name         string
}

// MarkUsedDirectives validates the whole summary before changing usage. Unknown
// entries are protocol errors: treating them as unused could erase a valid disable.
func (dm *DisableManager) MarkUsedDirectives(usages []DisableDirectiveUsage) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.ensureParsed()
	if dm.usageIndex == nil {
		dm.usageIndex = map[directiveUsageKey]*directiveEntry{}
		for i := range dm.tracked {
			d := &dm.tracked[i]
			if d.kind == directiveEnable {
				continue
			}
			for j := range d.entries {
				e := &d.entries[j]
				if !e.invalid {
					dm.usageIndex[directiveUsageKey{d.sourceLine, d.column, e.name}] = e
				}
			}
		}
	}
	entries := make([]*directiveEntry, 0, len(usages))
	for _, usage := range usages {
		name := ""
		if usage.RuleID != nil {
			name = *usage.RuleID
			if name == "" {
				return fmt.Errorf("empty named disable entry at %d:%d", usage.Line, usage.Column)
			}
		}
		entry := dm.usageIndex[directiveUsageKey{usage.Line, usage.Column, name}]
		if entry == nil {
			return fmt.Errorf("unknown disable entry at %d:%d for %q", usage.Line, usage.Column, name)
		}
		entries = append(entries, entry)
	}
	for _, entry := range entries {
		entry.used = true
	}
	return nil
}

// Freeze parses while syntax is leased and replaces it with a text-only frame.
// The resulting manager can be shared by native reporting and plugin returns.
func (dm *DisableManager) Freeze(source ast.SourceFileLike) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.ensureParsed()
	dm.comments = nil
	dm.sourceFile = source
}

// UnusedDiagnostics is called once, only after all producers finish successfully.
// ignored contains enabled rules skipped by execution policy, never off rules.
func (dm *DisableManager) UnusedDiagnostics(path string, severity DiagnosticSeverity, demand EditDemand, ignored map[string]bool) []RuleDiagnostic {
	if dm == nil || severity == SeverityOff {
		return nil
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.ensureParsed()
	// ESLint attributes enables in reverse order, independently of diagnostics.
	enabled := map[string]*directiveEntry{}
	for i := len(dm.tracked) - 1; i >= 0; i-- {
		d := &dm.tracked[i]
		if d.kind != directiveBlock && d.kind != directiveEnable {
			continue
		}
		for j := len(d.entries) - 1; j >= 0; j-- {
			e := &d.entries[j]
			if e.invalid {
				continue
			}
			if d.kind == directiveEnable {
				if e.name == "" {
					clear(enabled)
				}
				enabled[e.name] = e
			} else if e.name == "" {
				for _, enable := range enabled {
					enable.used = true
				}
				clear(enabled)
			} else if enable := enabled[e.name]; enable != nil {
				enable.used = true
			} else if enable := enabled[""]; enable != nil {
				enable.used = true
			}
		}
	}
	var diagnostics []RuleDiagnostic
	for _, d := range dm.tracked {
		var unused []string
		for _, e := range d.entries {
			if !e.invalid && !e.used && !ignored[e.name] && (e.name != "" || len(ignored) == 0) {
				unused = append(unused, e.name)
			}
		}
		if len(unused) == 0 {
			continue
		}
		whole := len(unused) == len(d.entries)
		groups := [][]string{unused}
		if !whole {
			groups = nil
			for _, name := range unused {
				groups = append(groups, []string{name})
			}
		}
		for _, names := range groups {
			description := directiveRuleDescription(names)
			message := "Unused eslint-disable directive (no problems were reported"
			if d.kind == directiveEnable {
				message = "Unused eslint-enable directive (no matching eslint-disable directives were found"
			}
			if description != "" {
				if d.kind == directiveEnable {
					message += " for "
				} else {
					message += " from "
				}
				message += description
			}
			message += ")."
			diagnostic := RuleDiagnostic{FilePath: path, SourceFile: dm.sourceFile, Range: core.NewTextRange(d.location.Pos(), d.location.Pos()), Severity: severity, Message: RuleMessage{Description: message}}
			if demand&EditDemandAutofix != 0 {
				fix := RuleFix{Range: d.location, Text: " "}
				if !whole {
					fix = dm.removeEntry(d, names[0])
				}
				fixes := []RuleFix{fix}
				diagnostic.FixesPtr = &fixes
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	return diagnostics
}

func directiveRuleDescription(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		if name != "" {
			quoted = append(quoted, "'"+name+"'")
		}
	}
	if len(quoted) <= 2 {
		return strings.Join(quoted, " or ")
	}
	return strings.Join(quoted[:len(quoted)-1], ", ") + ", or " + quoted[len(quoted)-1]
}

func (dm *DisableManager) removeEntry(d trackedDirective, name string) RuleFix {
	text := dm.sourceFile.Text()
	var selected directiveEntry
	for _, entry := range d.entries {
		if entry.name == name {
			selected = entry
			break
		}
	}
	a, b := selected.rangeInSource.Pos(), selected.rangeInSource.End()
	previous := strings.LastIndexByte(text[d.listStart:a], ',')
	next := strings.IndexByte(text[b:d.listEnd], ',')
	if previous >= 0 {
		a = d.listStart + previous
		if next >= 0 {
			b += next
		} else {
			a = len(strings.TrimRightFunc(text[:a], ecmascript.IsWhiteSpaceOrLineTerminator))
		}
	} else if next >= 0 {
		b += next + 1
		b += len(text[b:d.listEnd]) - len(strings.TrimLeftFunc(text[b:d.listEnd], ecmascript.IsWhiteSpaceOrLineTerminator))
	}
	return RuleFix{Range: core.NewTextRange(a, b)}
}

func (d *trackedDirective) locateEntries(text, content string) {
	contentStart := d.location.Pos() + 2 + strings.Index(text[d.location.Pos()+2:d.location.End()], content)
	keywordEnd := strings.IndexFunc(content, ecmascript.IsWhiteSpaceOrLineTerminator)
	if keywordEnd < 0 {
		keywordEnd = len(content)
	}
	d.listStart = contentStart + keywordEnd
	d.listEnd = contentStart + len(content)
	if explanation := directiveExplanationStart(text[d.listStart:d.listEnd]); explanation >= 0 {
		d.listEnd = d.listStart + explanation
	}
	offset := d.listStart
	for _, part := range strings.Split(text[d.listStart:d.listEnd], ",") {
		trimmed := ecmascript.StringTrim(part)
		start := offset + strings.Index(part, trimmed)
		name := trimmed
		if len(name) >= 2 && (name[0] == '\'' || name[0] == '"') && name[len(name)-1] == name[0] {
			name = name[1 : len(name)-1]
		}
		for i := range d.entries {
			if d.entries[i].name == name && d.entries[i].rangeInSource.End() == 0 {
				d.entries[i].rangeInSource = core.NewTextRange(start, start+len(trimmed))
			}
		}
		offset += len(part) + 1
	}
}

// ValidateRules excludes unavailable names from usage attribution and deletion.
// Invalid entries remain in their parent so a partial fix cannot erase them.
// The callback is consumed here, never retained across plugin completion.
func (dm *DisableManager) ValidateRules(exists func(string) bool) {
	if dm == nil || exists == nil {
		return
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.ensureParsed()
	for i := range dm.tracked {
		for j := range dm.tracked[i].entries {
			entry := &dm.tracked[i].entries[j]
			entry.invalid = entry.name != "" && !exists(entry.name)
		}
	}
}
