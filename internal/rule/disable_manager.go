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
	mu         sync.Mutex
	tracked    []trackedDirective
	sourceFile ast.SourceFileLike
	comments   *CommentStore
	parsed     bool
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
			commentContent = ecmascript.StringTrim(text[comment.Pos()+2 : comment.End()])
		case ast.KindMultiLineCommentTrivia:
			commentContent = ecmascript.StringTrim(text[comment.Pos()+2 : comment.End()-2])
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

		lineNum, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.Pos())
		endLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.End())
		if kind == directiveLine && lineNum != endLine {
			continue
		}
		targetLine := lineNum
		if kind == directiveNextLine {
			targetLine = endLine + 1
		}
		tracked := trackedDirective{kind: kind, location: core.NewTextRange(comment.Pos(), comment.End()), line: targetLine}
		if len(rules) == 0 {
			tracked.entries = []directiveEntry{{}}
		} else {
			for _, name := range rules {
				tracked.entries = append(tracked.entries, directiveEntry{name: name})
			}
		}
		tracked.locateEntries(text, commentContent)
		dm.tracked = append(dm.tracked, tracked)

	}
}

// matchDirective checks if a comment content string is a disable/enable directive.
// Returns the directive kind and any specified rule names.
func matchDirective(commentContent string) (directiveKind, []string) {
	for _, p := range directivePrefixes {
		for _, candidate := range []struct {
			keyword string
			kind    directiveKind
		}{
			{p.disable + "-next-line", directiveNextLine}, {p.disable + "-line", directiveLine}, {p.disable, directiveBlock}, {p.enable, directiveEnable},
		} {
			if !strings.HasPrefix(commentContent, candidate.keyword) {
				continue
			}
			rest := commentContent[len(candidate.keyword):]
			if rest != "" {
				first, _ := utf8.DecodeRuneInString(rest)
				if !ecmascript.IsWhiteSpaceOrLineTerminator(first) {
					continue
				}
			}
			return candidate.kind, parseRuleNames(rest)
		}
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

	var rules []string
	seen := map[string]bool{}
	for _, rule := range strings.Split(rulesStr, ",") {
		rule = ecmascript.StringTrim(rule)
		if len(rule) >= 2 && (rule[0] == '\'' || rule[0] == '"') && rule[len(rule)-1] == rule[0] {
			rule = rule[1 : len(rule)-1]
		}
		if rule != "" && !seen[rule] {
			rules = append(rules, rule)
			seen[rule] = true
		}
	}
	return rules
}

// IsRuleDisabled checks if a rule is disabled at the given position
func (dm *DisableManager) IsRuleDisabled(ruleName string, pos int) bool {
	if dm == nil || (dm.parsed && len(dm.tracked) == 0) {
		return false
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.ensureParsed()
	if len(dm.tracked) == 0 {
		return false
	}
	var active *directiveEntry
	for i := range dm.tracked {
		d := &dm.tracked[i]
		if d.location.Pos() > pos {
			break
		}
		if d.kind != directiveBlock && d.kind != directiveEnable {
			continue
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
	for i := range dm.tracked {
		d := &dm.tracked[i]
		if (d.kind != directiveLine && d.kind != directiveNextLine) || d.line != line {
			continue
		}
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

// ESLint recognizes a whitespace-delimited run of at least two hyphens.
func directiveExplanationStart(text string) int {
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
