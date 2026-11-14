package rule

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// blockDirective represents a block-level disable or enable event at a source location
type blockDirective struct {
	line      int
	column    core.UTF16Offset
	isDisable bool     // true = disable, false = enable
	rules     []string // nil means all rules (wildcard)
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

var disableDirectiveLabels = [...]struct {
	label string
	kind  directiveKind
}{
	{"disable", directiveBlock},
	{"enable", directiveEnable},
	{"disable-line", directiveLine},
	{"disable-next-line", directiveNextLine},
}

// An omitted list disables all rules. Empty names never appear in a parsed list.
const allLineRules = ""

// DisableManager tracks which rules are disabled at different locations in a file
type DisableManager struct {
	sourceFile            *ast.SourceFile
	comments              *CommentStore
	parsed                bool
	blockDirectives       []blockDirective // block disable/enable events in source order
	lineDisabledRules     map[int][]string // Rules disabled for specific lines
	nextLineDisabledRules map[int][]string // Rules disabled for the next line
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

// parseDirectives parses disable/enable directive comments from the source text.
// Both rslint- and eslint- prefixed directives are recognized.
func (dm *DisableManager) parseDirectives(comments []*ast.CommentRange) {
	if dm.sourceFile.Text() == "" || len(comments) == 0 {
		return
	}

	text := dm.sourceFile.Text()

	for _, comment := range comments {
		kind, rules := matchDirective(utils.CommentValue(text, comment))
		switch kind {
		case directiveLine:
			lineNum := scanner.GetECMALineOfPosition(dm.sourceFile, comment.Pos())
			if comment.Kind == ast.KindMultiLineCommentTrivia && scanner.GetECMALineOfPosition(dm.sourceFile, comment.End()) != lineNum {
				continue
			}
			if dm.lineDisabledRules == nil {
				dm.lineDisabledRules = make(map[int][]string)
			}
			if len(rules) == 0 {
				dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], allLineRules)
			} else {
				dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], rules...)
			}
		case directiveNextLine:
			nextLineNum := scanner.GetECMALineOfPosition(dm.sourceFile, comment.End()) + 1
			if dm.nextLineDisabledRules == nil {
				dm.nextLineDisabledRules = make(map[int][]string)
			}
			if len(rules) == 0 {
				dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], allLineRules)
			} else {
				dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], rules...)
			}
		case directiveBlock, directiveEnable:
			if comment.Kind != ast.KindMultiLineCommentTrivia {
				continue
			}
			lineNum, columnNum := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.Pos())
			dm.blockDirectives = append(dm.blockDirectives, blockDirective{
				line:      lineNum,
				column:    columnNum,
				isDisable: kind == directiveBlock,
				rules:     rules,
			})
		}
	}
}

// matchDirective checks if a comment content string is a disable/enable directive.
// Returns the directive kind and any specified rule names.
func matchDirective(commentContent string) (directiveKind, []string) {
	// Trailing whitespace can complete the separator for an empty description.
	commentContent = commentContent[ecmascript.SkipLeadingWhitespace(commentContent, 0, len(commentContent)):]
	suffix, ok := strings.CutPrefix(commentContent, "eslint-")
	if !ok {
		suffix, ok = strings.CutPrefix(commentContent, "rslint-")
		if !ok {
			return directiveNone, nil
		}
	}

	for _, directive := range disableDirectiveLabels {
		if value, ok := cutDirectiveLabel(suffix, directive.label); ok {
			return directive.kind, parseRuleNames(value)
		}
	}
	return directiveNone, nil
}

// parseRuleNames reads the comma-separated rule names before any description.
func parseRuleNames(rulesStr string) []string {
	if end := findDirectiveJustification(rulesStr, 0, len(rulesStr)); end >= 0 {
		rulesStr = rulesStr[:end]
	}
	var rules []string
	for rule := range strings.SplitSeq(rulesStr, ",") {
		start, end := trimECMAScriptWhitespaceRange(rule, 0, len(rule))
		start, end = stripMatchingQuoteRange(rule, start, end)
		if start < end {
			rules = append(rules, rule[start:end])
		}
	}
	return rules
}

// IsRuleDisabled checks if a rule is disabled at the given position
func (dm *DisableManager) IsRuleDisabled(ruleName string, pos int) bool {
	if dm == nil || dm.sourceFile == nil {
		return false
	}
	dm.ensureParsed()
	if len(dm.blockDirectives) == 0 && len(dm.lineDisabledRules) == 0 && len(dm.nextLineDisabledRules) == 0 {
		return false
	}

	line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, pos)

	// Check block disable/enable directives (range-based)
	if dm.isBlockDisabled(ruleName, line, column) {
		return true
	}

	// Check if rule is disabled for this specific line
	if lineRules, exists := dm.lineDisabledRules[line]; exists {
		for _, disabledRule := range lineRules {
			if disabledRule == ruleName || disabledRule == allLineRules {
				return true
			}
		}
	}

	// Check if rule is disabled for this line via next-line directive
	if nextLineRules, exists := dm.nextLineDisabledRules[line]; exists {
		for _, disabledRule := range nextLineRules {
			if disabledRule == ruleName || disabledRule == allLineRules {
				return true
			}
		}
	}

	return false
}

// isBlockDisabled replays block directives in source order to determine
// whether a rule is disabled at the given source location.
func (dm *DisableManager) isBlockDisabled(ruleName string, line int, column core.UTF16Offset) bool {
	allDisabled := false
	ruleDisabled := false
	hasRuleSpecific := false

	for _, d := range dm.blockDirectives {
		if d.line > line || (d.line == line && d.column > column) {
			break
		}

		if len(d.rules) == 0 {
			// Wildcard directive: affects all rules and resets rule-specific state
			allDisabled = d.isDisable
			hasRuleSpecific = false
		} else {
			for _, r := range d.rules {
				if r == ruleName {
					ruleDisabled = d.isDisable
					hasRuleSpecific = true
				}
			}
		}
	}

	if hasRuleSpecific {
		return ruleDisabled
	}
	return allDisabled
}
