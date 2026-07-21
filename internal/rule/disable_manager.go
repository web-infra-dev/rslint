package rule

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// directiveEntry represents one (rule, directive-occurrence) pair for a
// disable-kind directive (block disable, disable-line, disable-next-line).
// It exists purely for --report-unused-disable-directives bookkeeping: the
// suppression algorithms below (isBlockDisabled, IsRuleDisabled) mark an
// entry's Used field the first time it actually suppresses a diagnostic.
// rule == "" means the entry is the wildcard form (e.g. `eslint-disable`
// with no rule list), which applies to every rule.
type directiveEntry struct {
	rule string
	rng  core.TextRange
	used bool
}

// UnusedDirective describes a disable directive that never suppressed a
// diagnostic. RuleName is "" for the wildcard form.
type UnusedDirective struct {
	Range    core.TextRange
	RuleName string
}

// blockDirective represents a block-level disable or enable event at a specific line
type blockDirective struct {
	line      int
	isDisable bool     // true = disable, false = enable
	rules     []string // nil means all rules (wildcard)
	// entries is populated only for isDisable directives, parallel to rules
	// (or a single wildcard entry when rules is nil). Used to mark usage once
	// isBlockDisabled determines this directive is the one suppressing a
	// diagnostic.
	entries []*directiveEntry
}

// entryForRule returns the tracked entry for ruleName ("" for the wildcard
// entry), or nil if this directive carries no entries (e.g. an enable
// directive, which is never reported as unused).
func (d *blockDirective) entryForRule(ruleName string) *directiveEntry {
	for _, e := range d.entries {
		if e.rule == ruleName {
			return e
		}
	}
	return nil
}

// directiveKind represents the type of an inline directive comment.
type directiveKind int

const (
	directiveNone     directiveKind = iota
	directiveBlock                  // rslint-disable / eslint-disable (block)
	directiveEnable                 // rslint-enable / eslint-enable
	directiveLine                   // rslint-disable-line / eslint-disable-line
	directiveNextLine               // rslint-disable-next-line / eslint-disable-next-line
)

// directivePrefix defines the comment prefixes for disable/enable directives.
type directivePrefix struct {
	disable string // e.g. "rslint-disable"
	enable  string // e.g. "rslint-enable"
}

// directivePrefixes lists the supported directive prefixes.
// Both rslint- and eslint- prefixes are supported and fully equivalent.
var directivePrefixes = []directivePrefix{
	{"rslint-disable", "rslint-enable"},
	{"eslint-disable", "eslint-enable"},
}

// DisableManager tracks which rules are disabled at different locations in a file
type DisableManager struct {
	sourceFile            *ast.SourceFile
	comments              *CommentStore
	parsed                bool
	blockDirectives       []blockDirective          // block disable/enable events in source order
	lineDisabledRules     map[int][]*directiveEntry // Rules disabled for specific lines
	nextLineDisabledRules map[int][]*directiveEntry // Rules disabled for the next line
	// disableEntries collects every disable-kind directiveEntry (block,
	// line, next-line) in source order, for UnusedDirectives().
	disableEntries []*directiveEntry
}

// NewDisableManager creates a manager whose directives are parsed on the first
// disable check. The manager does not materialize comments without a directive.
func NewDisableManager(sourceFile *ast.SourceFile, comments *CommentStore) *DisableManager {
	return &DisableManager{
		sourceFile: sourceFile,
		comments:   comments,
	}
}

func (dm *DisableManager) ensureParsed() {
	if dm == nil || dm.parsed {
		return
	}
	dm.parsed = true
	if dm.sourceFile == nil || !mayContainDisableDirective(dm.sourceFile.Text()) {
		return
	}
	dm.parseDirectives(dm.comments.All())
}

func mayContainDisableDirective(text string) bool {
	const marker = "lint-"
	for searchStart := 0; searchStart < len(text); {
		offset := strings.Index(text[searchStart:], marker)
		if offset < 0 {
			return false
		}
		start := searchStart + offset
		if start >= 2 {
			prefix := text[start-2 : start]
			rest := text[start+len(marker):]
			if (prefix == "rs" || prefix == "es") &&
				(strings.HasPrefix(rest, "disable") || strings.HasPrefix(rest, "enable")) {
				return true
			}
		}
		searchStart = start + len(marker)
	}
	return false
}

// newDisableEntries builds one directiveEntry per rule name (or a single
// wildcard entry when rules is empty), all sharing the same comment range.
// It also appends every entry to dm.disableEntries for UnusedDirectives().
func (dm *DisableManager) newDisableEntries(rng core.TextRange, rules []string) []*directiveEntry {
	var entries []*directiveEntry
	if len(rules) == 0 {
		entries = []*directiveEntry{{rule: "", rng: rng}}
	} else {
		entries = make([]*directiveEntry, len(rules))
		for i, r := range rules {
			entries[i] = &directiveEntry{rule: r, rng: rng}
		}
	}
	dm.disableEntries = append(dm.disableEntries, entries...)
	return entries
}

// parseDirectives parses disable/enable directive comments from the source text.
// Both rslint- and eslint- prefixed directives are recognized.
func (dm *DisableManager) parseDirectives(comments []*ast.CommentRange) {
	if dm.sourceFile.Text() == "" || len(comments) == 0 {
		return
	}

	text := dm.sourceFile.Text()

	for _, comment := range comments {
		var commentContent string
		switch comment.Kind {
		case ast.KindSingleLineCommentTrivia:
			commentContent = strings.TrimSpace(text[comment.Pos()+2 : comment.End()])
		case ast.KindMultiLineCommentTrivia:
			commentContent = strings.TrimSpace(text[comment.Pos()+2 : comment.End()-2])
		default:
			continue
		}

		kind, rules := matchDirective(commentContent)
		if kind == directiveNone {
			continue
		}

		lineNum, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.Pos())
		commentRange := core.NewTextRange(comment.Pos(), comment.End())

		switch kind {
		case directiveLine:
			entries := dm.newDisableEntries(commentRange, rules)
			if dm.lineDisabledRules == nil {
				dm.lineDisabledRules = make(map[int][]*directiveEntry)
			}
			dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], entries...)
		case directiveNextLine:
			nextLineNum := lineNum + 1
			entries := dm.newDisableEntries(commentRange, rules)
			if dm.nextLineDisabledRules == nil {
				dm.nextLineDisabledRules = make(map[int][]*directiveEntry)
			}
			dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], entries...)
		case directiveBlock:
			dm.blockDirectives = append(dm.blockDirectives, blockDirective{
				line:      lineNum,
				isDisable: true,
				rules:     rules,
				entries:   dm.newDisableEntries(commentRange, rules),
			})
		case directiveEnable:
			dm.blockDirectives = append(dm.blockDirectives, blockDirective{
				line:      lineNum,
				isDisable: false,
				rules:     rules,
			})
		}
	}
}

// matchDirective checks if a comment content string is a disable/enable directive.
// Returns the directive kind and any specified rule names.
func matchDirective(commentContent string) (directiveKind, []string) {
	for _, p := range directivePrefixes {
		if strings.HasPrefix(commentContent, p.disable) {
			rest := commentContent[len(p.disable):]
			if strings.HasPrefix(rest, "-line") {
				return directiveLine, parseRuleNames(rest[len("-line"):])
			}
			if strings.HasPrefix(rest, "-next-line") {
				return directiveNextLine, parseRuleNames(rest[len("-next-line"):])
			}
			return directiveBlock, parseRuleNames(rest)
		}
		if strings.HasPrefix(commentContent, p.enable) {
			return directiveEnable, parseRuleNames(commentContent[len(p.enable):])
		}
	}
	return directiveNone, nil
}

// parseRuleNames parses rule names from a string like " rule1, rule2, rule3"
// It also strips inline descriptions after " -- " (e.g., "rule1 -- reason")
func parseRuleNames(rulesStr string) []string {
	// Strip inline description after " -- " before trimming, so that
	// wildcard-with-description like " -- reason" is correctly handled.
	if idx := strings.Index(rulesStr, " -- "); idx != -1 {
		rulesStr = rulesStr[:idx]
	}

	rulesStr = strings.TrimSpace(rulesStr)
	if rulesStr == "" {
		return nil
	}

	var rules []string
	for _, rule := range strings.Split(rulesStr, ",") {
		rule = strings.TrimSpace(rule)
		if rule != "" {
			rules = append(rules, rule)
		}
	}
	return rules
}

// IsRuleDisabled checks if a rule is disabled at the given position. When it
// is, the responsible directive's usage is recorded for UnusedDirectives().
func (dm *DisableManager) IsRuleDisabled(ruleName string, pos int) bool {
	if dm == nil || dm.sourceFile == nil {
		return false
	}
	dm.ensureParsed()
	if len(dm.blockDirectives) == 0 && len(dm.lineDisabledRules) == 0 && len(dm.nextLineDisabledRules) == 0 {
		return false
	}

	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, pos)

	// Check block disable/enable directives (range-based)
	if disabled, entry := dm.isBlockDisabled(ruleName, line); disabled {
		if entry != nil {
			entry.used = true
		}
		return true
	}

	// Check if rule is disabled for this specific line
	if lineEntries, exists := dm.lineDisabledRules[line]; exists {
		for _, e := range lineEntries {
			if e.rule == ruleName || e.rule == "" {
				e.used = true
				return true
			}
		}
	}

	// Check if rule is disabled for this line via next-line directive
	if nextLineEntries, exists := dm.nextLineDisabledRules[line]; exists {
		for _, e := range nextLineEntries {
			if e.rule == ruleName || e.rule == "" {
				e.used = true
				return true
			}
		}
	}

	return false
}

// isBlockDisabled replays block directives in source order to determine
// whether a rule is disabled at the given line. When it is, it also returns
// the directiveEntry responsible, so the caller can mark it used.
func (dm *DisableManager) isBlockDisabled(ruleName string, line int) (bool, *directiveEntry) {
	allDisabled := false
	ruleDisabled := false
	hasRuleSpecific := false
	var allWinner *blockDirective
	var ruleWinner *blockDirective

	for i := range dm.blockDirectives {
		d := &dm.blockDirectives[i]
		if d.line > line {
			break
		}

		if len(d.rules) == 0 {
			// Wildcard directive: affects all rules and resets rule-specific state
			allDisabled = d.isDisable
			allWinner = d
			hasRuleSpecific = false
		} else {
			for _, r := range d.rules {
				if r == ruleName {
					ruleDisabled = d.isDisable
					ruleWinner = d
					hasRuleSpecific = true
				}
			}
		}
	}

	if hasRuleSpecific {
		if !ruleDisabled {
			return false, nil
		}
		return true, ruleWinner.entryForRule(ruleName)
	}
	if !allDisabled {
		return false, nil
	}
	if allWinner == nil {
		return false, nil
	}
	return true, allWinner.entryForRule("")
}

// UnusedDirectives returns every disable directive (block, line, or
// next-line) that never suppressed a diagnostic. `eslint-enable` directives
// are never included — ESLint only reports unused *disable* directives.
// Must be called after all rules have finished linting the file.
func (dm *DisableManager) UnusedDirectives() []UnusedDirective {
	if dm == nil {
		return nil
	}
	dm.ensureParsed()
	var out []UnusedDirective
	for _, e := range dm.disableEntries {
		if !e.used {
			out = append(out, UnusedDirective{Range: e.rng, RuleName: e.rule})
		}
	}
	return out
}
