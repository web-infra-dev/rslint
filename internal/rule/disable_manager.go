package rule

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// ESLintDirective represents an ESLint disable/enable directive
type ESLintDirective struct {
	Kind      ESLintDirectiveKind
	Line      int
	RuleNames []string
}

type ESLintDirectiveKind int

const (
	ESLintDirectiveDisable ESLintDirectiveKind = iota
	ESLintDirectiveEnable
	ESLintDirectiveDisableLine
	ESLintDirectiveDisableNextLine
)

// DisableManager tracks which rules are disabled at different locations in a file
type DisableManager struct {
	sourceFile            *ast.SourceFile
	disabledRules         map[string]bool  // Rules disabled for the entire file
	lineDisabledRules     map[int][]string // Rules disabled for specific lines
	nextLineDisabledRules map[int][]string // Rules disabled for the next line
}

// NewDisableManager creates a new DisableManager for the given source file
func NewDisableManager(sourceFile *ast.SourceFile, comments []*ast.CommentRange) *DisableManager {
	dm := &DisableManager{
		sourceFile:            sourceFile,
		disabledRules:         make(map[string]bool),
		lineDisabledRules:     make(map[int][]string),
		nextLineDisabledRules: make(map[int][]string),
	}

	dm.parseESLintDirectives(comments)
	return dm
}

// parseESLintDirectives parses ESLint-style disable/enable comments from the source text
func (dm *DisableManager) parseESLintDirectives(comments []*ast.CommentRange) {
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
		}

		lineNum, _ := scanner.GetECMALineAndCharacterOfPosition(dm.sourceFile, comment.Pos())
		rulePos := 0

		if strings.HasPrefix(commentContent, "eslint-disable") {
			rulePos += 14
			text := commentContent[rulePos:]

			if strings.HasPrefix(text, "-line") {
				// Check for eslint-disable-line
				rulePos += 5
				rules := parseRuleNames(commentContent[rulePos:])
				if len(rules) == 0 {
					dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], "*")
				} else {
					dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], rules...)
				}
			} else if strings.HasPrefix(text, "-next-line") {
				// Check for eslint-disable-next-line
				rulePos += 10
				rules := parseRuleNames(commentContent[rulePos:])
				nextLineNum := lineNum + 1
				if len(rules) == 0 {
					dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], "*")
				} else {
					dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], rules...)
				}
			} else {
				// Check for eslint-disable (block comments)
				rules := parseRuleNames(commentContent[rulePos:])
				if len(rules) == 0 {
					dm.disabledRules["*"] = true
				} else {
					for _, rule := range rules {
						dm.disabledRules[rule] = true
					}
				}
			}
		} else if strings.HasPrefix(commentContent, "eslint-enable") {
			rulePos += 13

			// Check for eslint-enable (block comments)
			rules := parseRuleNames(commentContent[rulePos:])
			if len(rules) == 0 {
				// Enable all rules
				for key := range dm.disabledRules {
					delete(dm.disabledRules, key)
				}
			} else {
				for _, rule := range rules {
					delete(dm.disabledRules, rule)
				}
			}
		}
	}
}

// parseRuleNames parses rule names from a string like "rule1, rule2, rule3"
func parseRuleNames(rulesStr string) []string {
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
	// Check if rule is disabled for the entire file
	if dm.disabledRules[ruleName] || dm.disabledRules["*"] {
		return true
	}

	// Get the line number for the position
	line, _ := scanner.GetECMALineAndCharacterOfPosition(dm.sourceFile, pos)

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
