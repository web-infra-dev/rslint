package utils

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// CommentDirective represents a parsed rslint directive from a comment
type CommentDirective struct {
	Type      DirectiveType // disable, enable, etc.
	Rules     []string      // empty means all rules
	Range     core.TextRange
	AppliesTo ApplyTarget // next-line, file, etc.
}

type DirectiveType int

const (
	DirectiveDisable DirectiveType = iota
	DirectiveEnable
)

type ApplyTarget int

const (
	ApplyToNextLine ApplyTarget = iota
	ApplyToFile
)

// RuleDisableTracker tracks which rules are disabled for which ranges in a file
type RuleDisableTracker struct {
	// disabledRules maps rule names to ranges where they're disabled
	disabledRules map[string][]core.TextRange
	// fileDisabled tracks if all rules are disabled for the rest of the file from a certain position
	fileDisabled []core.TextPos
}

// NewRuleDisableTracker creates a new tracker
func NewRuleDisableTracker() *RuleDisableTracker {
	return &RuleDisableTracker{
		disabledRules: make(map[string][]core.TextRange),
		fileDisabled:  make([]core.TextPos, 0),
	}
}

// IsRuleDisabled checks if a rule is disabled at a given position
func (t *RuleDisableTracker) IsRuleDisabled(ruleName string, pos core.TextPos) bool {
	// Check if all rules are disabled for the file from this position
	for _, disablePos := range t.fileDisabled {
		if pos >= disablePos {
			return true
		}
	}

	// Check if specific rule is disabled
	if ranges, exists := t.disabledRules[ruleName]; exists {
		for _, disableRange := range ranges {
			if int(pos) >= disableRange.Pos() && int(pos) < disableRange.End() {
				return true
			}
		}
	}

	return false
}

// ParseCommentDirectives parses rslint directives from all comments in a source file
func ParseCommentDirectives(sourceFile *ast.SourceFile) []CommentDirective {
	var directives []CommentDirective
	text := sourceFile.Text()

	// Use regex to find comments - simpler and safer than scanner approach
	// Match single-line comments: // comment
	singleLineRegex := regexp.MustCompile(`//(.*)`)
	// Match multi-line comments: /* comment */
	multiLineRegex := regexp.MustCompile(`/\*(.*?)\*/`)

	// Find single-line comments
	matches := singleLineRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		commentStart := match[0]
		commentEnd := match[1]
		contentStart := match[2]
		contentEnd := match[3]
		
		if contentStart >= 0 && contentEnd >= 0 {
			content := text[contentStart:contentEnd]
			if directive := parseDirectiveContent(content); directive != nil {
				directive.Range = core.NewTextRange(commentStart, commentEnd)
				directives = append(directives, *directive)
			}
		}
	}

	// Find multi-line comments
	matches = multiLineRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		commentStart := match[0]
		commentEnd := match[1]
		contentStart := match[2]
		contentEnd := match[3]
		
		if contentStart >= 0 && contentEnd >= 0 {
			content := text[contentStart:contentEnd]
			if directive := parseDirectiveContent(content); directive != nil {
				directive.Range = core.NewTextRange(commentStart, commentEnd)
				directives = append(directives, *directive)
			}
		}
	}

	return directives
}

// parseDirectiveContent parses the content of a comment for rslint directives
func parseDirectiveContent(content string) *CommentDirective {
	// Regex patterns for rslint directives
	disableAllPattern := regexp.MustCompile(`^rslint-disable\s*$`)
	disableRulePattern := regexp.MustCompile(`^rslint-disable\s+(.+)$`)

	content = strings.TrimSpace(content)

	if disableAllPattern.MatchString(content) {
		return &CommentDirective{
			Type:      DirectiveDisable,
			Rules:     []string{}, // empty means all rules
			AppliesTo: ApplyToFile,
		}
	}

	if matches := disableRulePattern.FindStringSubmatch(content); matches != nil {
		rules := parseRuleNames(matches[1])
		return &CommentDirective{
			Type:      DirectiveDisable,
			Rules:     rules,
			AppliesTo: ApplyToNextLine,
		}
	}

	return nil
}

// parseRuleNames parses rule names from a string (comma or space separated)
func parseRuleNames(ruleStr string) []string {
	// Split by comma or space, clean up whitespace
	parts := regexp.MustCompile(`[,\s]+`).Split(strings.TrimSpace(ruleStr), -1)
	var rules []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			rules = append(rules, part)
		}
	}
	return rules
}

// ApplyDirectives applies comment directives to build a rule disable tracker
func ApplyDirectives(directives []CommentDirective, sourceFile *ast.SourceFile) *RuleDisableTracker {
	tracker := NewRuleDisableTracker()

	for _, directive := range directives {
		switch directive.Type {
		case DirectiveDisable:
			if len(directive.Rules) == 0 {
				// Disable all rules for rest of file
				tracker.fileDisabled = append(tracker.fileDisabled, core.TextPos(directive.Range.Pos()))
			} else {
				// Disable specific rules
				switch directive.AppliesTo {
				case ApplyToNextLine:
					// Find the next line and create a range for it
					nextLineRange := getNextLineRange(directive.Range, sourceFile)
					if nextLineRange != nil {
						for _, ruleName := range directive.Rules {
							tracker.disabledRules[ruleName] = append(tracker.disabledRules[ruleName], *nextLineRange)
						}
					}
				case ApplyToFile:
					// Apply to rest of file (should not happen for specific rules in current design)
					for _, ruleName := range directive.Rules {
						endOfFile := core.TextPos(len(sourceFile.Text()))
						tracker.disabledRules[ruleName] = append(tracker.disabledRules[ruleName], 
							core.NewTextRange(directive.Range.Pos(), int(endOfFile)))
					}
				}
			}
		}
	}

	return tracker
}

// getNextLineRange finds the range of the next line after a comment
func getNextLineRange(commentRange core.TextRange, sourceFile *ast.SourceFile) *core.TextRange {
	lineMap := sourceFile.LineMap()
	
	// Find which line the comment is on
	commentLine, _ := scanner.GetLineAndCharacterOfPosition(sourceFile, int(commentRange.End()))
	
	// Get the next line
	nextLine := commentLine + 1
	if nextLine >= len(lineMap) {
		return nil // No next line
	}
	
	// Calculate range of next line
	nextLineStart := lineMap[nextLine]
	var nextLineEnd core.TextPos
	if nextLine+1 < len(lineMap) {
		nextLineEnd = lineMap[nextLine+1] - 1 // Exclude newline
	} else {
		nextLineEnd = core.TextPos(len(sourceFile.Text()))
	}
	
	textRange := core.NewTextRange(int(nextLineStart), int(nextLineEnd))
	return &textRange
}