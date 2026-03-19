package rule

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// blockDirective represents a block-level disable or enable event at a specific line
type blockDirective struct {
	line      int
	isDisable bool     // true = disable, false = enable
	rules     []string // nil means all rules (wildcard)
}

// DisableManager tracks which rules are disabled at different locations in a file
type DisableManager struct {
	sourceFile            *ast.SourceFile
	blockDirectives       []blockDirective // block disable/enable events in source order
	lineDisabledRules     map[int][]string // Rules disabled for specific lines
	nextLineDisabledRules map[int][]string // Rules disabled for the next line
}

// NewDisableManager creates a new DisableManager for the given source file
func NewDisableManager(sourceFile *ast.SourceFile, comments []*ast.CommentRange) *DisableManager {
	dm := &DisableManager{
		sourceFile:            sourceFile,
		lineDisabledRules:     make(map[int][]string),
		nextLineDisabledRules: make(map[int][]string),
	}

	dm.parseDirectives(comments)
	return dm
}

// parseDirectives parses eslint-disable/enable comments from the source text
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

		lineNum, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, comment.Pos())
		rulePos := 0

		if strings.HasPrefix(commentContent, "eslint-disable") {
			rulePos += 14
			rest := commentContent[rulePos:]

			if strings.HasPrefix(rest, "-line") {
				// eslint-disable-line
				rulePos += 5
				rules := parseRuleNames(commentContent[rulePos:])
				if len(rules) == 0 {
					dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], "*")
				} else {
					dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], rules...)
				}
			} else if strings.HasPrefix(rest, "-next-line") {
				// eslint-disable-next-line
				rulePos += 10
				rules := parseRuleNames(commentContent[rulePos:])
				nextLineNum := lineNum + 1
				if len(rules) == 0 {
					dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], "*")
				} else {
					dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], rules...)
				}
			} else {
				// eslint-disable (block)
				rules := parseRuleNames(commentContent[rulePos:])
				dm.blockDirectives = append(dm.blockDirectives, blockDirective{
					line:      lineNum,
					isDisable: true,
					rules:     rules,
				})
			}
		} else if strings.HasPrefix(commentContent, "eslint-enable") {
			rulePos += 13
			rules := parseRuleNames(commentContent[rulePos:])
			dm.blockDirectives = append(dm.blockDirectives, blockDirective{
				line:      lineNum,
				isDisable: false,
				rules:     rules,
			})
		}
	}
}

// parseRuleNames parses rule names from a string like " rule1, rule2, rule3"
func parseRuleNames(rulesStr string) []string {
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

// IsRuleDisabled checks if a rule is disabled at the given position
func (dm *DisableManager) IsRuleDisabled(ruleName string, pos int) bool {
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(dm.sourceFile, pos)

	// Check block disable/enable directives (range-based)
	if dm.isBlockDisabled(ruleName, line) {
		return true
	}

	// Check if rule is disabled for this specific line
	if lineRules, exists := dm.lineDisabledRules[line]; exists {
		for _, disabledRule := range lineRules {
			if disabledRule == ruleName || disabledRule == "*" {
				return true
			}
		}
	}

	// Check if rule is disabled for this line via next-line directive
	if nextLineRules, exists := dm.nextLineDisabledRules[line]; exists {
		for _, disabledRule := range nextLineRules {
			if disabledRule == ruleName || disabledRule == "*" {
				return true
			}
		}
	}

	return false
}

// isBlockDisabled replays block directives in source order to determine
// whether a rule is disabled at the given line.
func (dm *DisableManager) isBlockDisabled(ruleName string, line int) bool {
	allDisabled := false
	ruleDisabled := false
	hasRuleSpecific := false

	for _, d := range dm.blockDirectives {
		if d.line > line {
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
