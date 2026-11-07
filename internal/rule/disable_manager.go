package rule

import (
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

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
	sourceFile        ast.SourceFileLike
	comments          *CommentStore
	parsed            bool
	blockDirectives   []suppressionBlock
	lineDisabledRules map[int][]string
	*directiveTracking
}

// Only reporting managers share state with plugin completion. Suppression-only
// managers belong to one file's native shard and need neither a lock nor entries.
type directiveTracking struct {
	mu         sync.Mutex
	tracked    []trackedDirective
	blocks     []int
	lines      map[int][]int
	usageIndex map[directiveUsageKey]*directiveEntry
}

type suppressionBlock struct {
	pos      int
	rules    []string
	disabled bool
}

// NewDisableManager creates a manager whose directives are parsed on the first
// disable check. The manager does not materialize comments without a directive.
func NewDisableManager(sourceFile *ast.SourceFile, comments *CommentStore) *DisableManager {
	return &DisableManager{
		sourceFile:        sourceFile,
		comments:          comments,
		directiveTracking: &directiveTracking{},
	}
}

// NewSuppressionManager retains lazy matching without unused-reporting metadata.
func NewSuppressionManager(sourceFile *ast.SourceFile, comments *CommentStore) *DisableManager {
	return &DisableManager{sourceFile: sourceFile, comments: comments}
}

func (dm *DisableManager) ensureParsed() {
	if dm == nil || dm.parsed {
		return
	}
	dm.parsed = true
	if dm.sourceFile == nil || !MayContainDisableDirective(dm.sourceFile.Text()) {
		return
	}
	dm.parseDirectives(dm.comments.All())
	dm.comments = nil
}

// MayContainDisableDirective cheaply rejects text without a supported keyword.
func MayContainDisableDirective(text string) bool {
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
			commentContent = strings.TrimLeftFunc(text[comment.Pos()+2:comment.End()], ecmascript.IsWhiteSpaceOrLineTerminator)
		case ast.KindMultiLineCommentTrivia:
			commentContent = strings.TrimLeftFunc(text[comment.Pos()+2:comment.End()-2], ecmascript.IsWhiteSpaceOrLineTerminator)
		default:
			continue
		}

		kind, rules := matchDirective(commentContent)
		if kind == directiveNone {
			continue
		}
		// Only line-flavored directives can be carried by line comments.
		if comment.Kind == ast.KindSingleLineCommentTrivia && (kind == directiveBlock || kind == directiveEnable) {
			continue
		}

		if dm.directiveTracking == nil && (kind == directiveBlock || kind == directiveEnable) {
			dm.blockDirectives = append(dm.blockDirectives, suppressionBlock{pos: comment.Pos(), rules: rules, disabled: kind == directiveBlock})
			continue
		}
		lineNum, column := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.Pos())
		endLine := lineNum
		if comment.Kind == ast.KindMultiLineCommentTrivia && (kind == directiveLine || kind == directiveNextLine) {
			endLine, _ = scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.End())
		}
		if kind == directiveLine && lineNum != endLine {
			continue
		}
		targetLine := lineNum
		if kind == directiveNextLine {
			targetLine = endLine + 1
		}
		if dm.directiveTracking == nil {
			if dm.lineDisabledRules == nil {
				dm.lineDisabledRules = make(map[int][]string)
			}
			if len(rules) == 0 {
				rules = []string{""}
			}
			if previous, exists := dm.lineDisabledRules[targetLine]; exists {
				dm.lineDisabledRules[targetLine] = append(previous, rules...)
			} else {
				dm.lineDisabledRules[targetLine] = rules
			}
			continue
		}
		tracked := trackedDirective{kind: kind, location: core.NewTextRange(comment.Pos(), comment.End()), line: targetLine, sourceLine: lineNum, column: int(column)}
		if len(rules) == 0 {
			tracked.entries = []directiveEntry{{}}
		} else {
			for _, name := range rules {
				tracked.entries = append(tracked.entries, directiveEntry{name: name})
			}
		}
		tracked.locateEntries(text, commentContent)
		index := len(dm.tracked)
		dm.tracked = append(dm.tracked, tracked)
		if kind == directiveBlock || kind == directiveEnable {
			dm.blocks = append(dm.blocks, index)
		} else {
			if dm.lines == nil {
				dm.lines = map[int][]int{}
			}
			dm.lines[targetLine] = append(dm.lines[targetLine], index)
		}
	}
}

// matchDirective checks if a comment content string is a disable/enable directive.
// Returns the directive kind and any specified rule names.
func matchDirective(commentContent string) (directiveKind, []string) {
	for _, p := range directivePrefixes {
		var rest string
		var kind directiveKind
		if strings.HasPrefix(commentContent, p.disable) {
			rest = commentContent[len(p.disable):]
			kind = directiveBlock
			if strings.HasPrefix(rest, "-next-line") {
				rest = rest[len("-next-line"):]
				kind = directiveNextLine
			} else if strings.HasPrefix(rest, "-line") {
				rest = rest[len("-line"):]
				kind = directiveLine
			}
		} else if strings.HasPrefix(commentContent, p.enable) {
			rest = commentContent[len(p.enable):]
			kind = directiveEnable
		} else {
			continue
		}
		if rest != "" {
			first, _ := utf8.DecodeRuneInString(rest)
			if !ecmascript.IsWhiteSpaceOrLineTerminator(first) {
				continue
			}
		}
		return kind, parseRuleNames(rest)
	}
	return directiveNone, nil
}

// parseRuleNames parses rule names from a string like " rule1, rule2, rule3"
// It also strips inline descriptions after " -- " (e.g., "rule1 -- reason")
func parseRuleNames(rulesStr string) []string {
	// Strip inline description after " -- " before trimming, so that
	// wildcard-with-description like " -- reason" is correctly handled.
	if idx := directiveExplanationStart(rulesStr); idx != -1 {
		rulesStr = rulesStr[:idx]
	}

	rulesStr = ecmascript.StringTrim(rulesStr)
	if rulesStr == "" {
		return nil
	}

	if !strings.Contains(rulesStr, ",") {
		name := trimRuleName(rulesStr)
		if name == "" {
			return nil
		}
		return []string{name}
	}
	var rules []string
	seen := map[string]bool{}
	for _, rule := range strings.Split(rulesStr, ",") {
		rule = trimRuleName(rule)
		if rule != "" && !seen[rule] {
			rules = append(rules, rule)
			seen[rule] = true
		}
	}
	return rules
}

func trimRuleName(name string) string {
	name = ecmascript.StringTrim(name)
	if len(name) >= 2 && (name[0] == '\'' || name[0] == '"') && name[len(name)-1] == name[0] {
		return name[1 : len(name)-1]
	}
	return name
}

// IsRuleDisabled checks if a rule is disabled at the given position
func (dm *DisableManager) IsRuleDisabled(ruleName string, pos int) bool {
	if dm == nil {
		return false
	}
	if dm.directiveTracking == nil {
		return dm.isSuppressed(ruleName, pos)
	}
	if dm.parsed && len(dm.tracked) == 0 {
		return false
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.ensureParsed()
	if len(dm.tracked) == 0 {
		return false
	}
	var active *directiveEntry
	for _, index := range dm.blocks {
		d := &dm.tracked[index]
		if d.location.Pos() > pos {
			break
		}
		for j := range d.entries {
			e := &d.entries[j]
			if e.invalid {
				continue
			}
			if e.name != "" && e.name != ruleName {
				continue
			}
			if d.kind == directiveEnable {
				active = nil
			} else {
				active = e
			}
		}
	}
	if active != nil {
		active.used = true
		return true
	}
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, pos)
	// Line directives form a separate pass. Their stable source ordering
	// determines attribution when next-line and same-line comments overlap.
	for _, index := range dm.lines[line] {
		d := &dm.tracked[index]
		for j := range d.entries {
			e := &d.entries[j]
			if e.invalid {
				continue
			}
			if e.name == "" || e.name == ruleName {
				active = e
			}
		}
	}
	if active != nil {
		active.used = true
		return true
	}
	return false
}

// isSuppressed is the file-local path: compact names, direct line lookup,
// and no usage bookkeeping or synchronization in the report hot loop.
func (dm *DisableManager) isSuppressed(ruleName string, pos int) bool {
	dm.ensureParsed()
	if len(dm.blockDirectives) == 0 && len(dm.lineDisabledRules) == 0 {
		return false
	}
	disabled := false
	for _, directive := range dm.blockDirectives {
		if directive.pos > pos {
			break
		}
		if len(directive.rules) == 0 {
			disabled = directive.disabled
		} else {
			for _, name := range directive.rules {
				if name == ruleName {
					disabled = directive.disabled
					break
				}
			}
		}
	}
	if disabled {
		return true
	}
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, pos)
	for _, name := range dm.lineDisabledRules[line] {
		if name == "" || name == ruleName {
			return true
		}
	}
	return false
}

// ESLint recognizes a whitespace-delimited run of at least two hyphens.
func directiveExplanationStart(text string) int {
	if !strings.Contains(text, "--") {
		return -1
	}
	for i := 1; i+2 < len(text); i++ {
		if text[i] != '-' || text[i+1] != '-' {
			continue
		}
		before, _ := utf8.DecodeLastRuneInString(text[:i])
		if !ecmascript.IsWhiteSpaceOrLineTerminator(before) {
			continue
		}
		end := i + 2
		for end < len(text) && text[end] == '-' {
			end++
		}
		after, _ := utf8.DecodeRuneInString(text[end:])
		if ecmascript.IsWhiteSpaceOrLineTerminator(after) {
			return i
		}
	}
	return -1
}
