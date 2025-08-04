package rule

import (
	"regexp"
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
func NewDisableManager(sourceFile *ast.SourceFile) *DisableManager {
	dm := &DisableManager{
		sourceFile:            sourceFile,
		disabledRules:         make(map[string]bool),
		lineDisabledRules:     make(map[int][]string),
		nextLineDisabledRules: make(map[int][]string),
	}

	dm.parseESLintDirectives()
	return dm
}

// parseESLintDirectives parses ESLint-style disable/enable comments from the source text
func (dm *DisableManager) parseESLintDirectives() {
	if dm.sourceFile.Text() == "" {
		return
	}

	text := dm.sourceFile.Text()
	lines := strings.Split(text, "\n")

	// Regular expressions to match ESLint directives
	eslintDisableLineRe := regexp.MustCompile(`//\s*eslint-disable-line(?:\s+([^\r\n]+))?`)
	eslintDisableNextLineRe := regexp.MustCompile(`//\s*eslint-disable-next-line(?:\s+([^\r\n]+))?`)
	eslintDisableRe := regexp.MustCompile(`/\*\s*eslint-disable(?:\s+([^*]+))?\s*\*/`)
	eslintEnableRe := regexp.MustCompile(`/\*\s*eslint-enable(?:\s+([^*]+))?\s*\*/`)

	for i, line := range lines {
		lineNum := i // 0-based line numbers

		// Check for eslint-disable-line
		if matches := eslintDisableLineRe.FindStringSubmatch(line); matches != nil {
			rules := parseRuleNames(matches[1])
			if len(rules) == 0 {
				dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], "*")
			} else {
				dm.lineDisabledRules[lineNum] = append(dm.lineDisabledRules[lineNum], rules...)
			}
		}

		// Check for eslint-disable-next-line
		if matches := eslintDisableNextLineRe.FindStringSubmatch(line); matches != nil {
			rules := parseRuleNames(matches[1])
			nextLineNum := lineNum + 1
			if len(rules) == 0 {
				dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], "*")
			} else {
				dm.nextLineDisabledRules[nextLineNum] = append(dm.nextLineDisabledRules[nextLineNum], rules...)
			}
		}

		// Check for eslint-disable (block comments)
		if matches := eslintDisableRe.FindStringSubmatch(line); matches != nil {
			rules := parseRuleNames(matches[1])
			if len(rules) == 0 {
				dm.disabledRules["*"] = true
			} else {
				for _, rule := range rules {
					dm.disabledRules[rule] = true
				}
			}
		}

		// Check for eslint-enable (block comments)
		if matches := eslintEnableRe.FindStringSubmatch(line); matches != nil {
			rules := parseRuleNames(matches[1])
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
	line, _ := scanner.GetLineAndCharacterOfPosition(dm.sourceFile, pos)

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
