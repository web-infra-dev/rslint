package utils

import (
	"testing"
	"github.com/microsoft/typescript-go/shim/core"
)

func TestParseCommentDirectives(t *testing.T) {
	// Test parseDirectiveContent function directly
	testCases := []struct {
		input    string
		expected *CommentDirective
	}{
		{
			"rslint-disable",
			&CommentDirective{Type: DirectiveDisable, Rules: []string{}, AppliesTo: ApplyToFile},
		},
		{
			"rslint-disable await-thenable",
			&CommentDirective{Type: DirectiveDisable, Rules: []string{"await-thenable"}, AppliesTo: ApplyToNextLine},
		},
		{
			"rslint-disable rule1,rule2",
			&CommentDirective{Type: DirectiveDisable, Rules: []string{"rule1", "rule2"}, AppliesTo: ApplyToNextLine},
		},
		{
			"just a normal comment",
			nil,
		},
	}
	
	for _, tc := range testCases {
		result := parseDirectiveContent(tc.input)
		if tc.expected == nil {
			if result != nil {
				t.Errorf("Expected nil for input %q, got %+v", tc.input, result)
			}
		} else {
			if result == nil {
				t.Errorf("Expected %+v for input %q, got nil", tc.expected, tc.input)
				continue
			}
			if result.Type != tc.expected.Type {
				t.Errorf("Expected Type %v for input %q, got %v", tc.expected.Type, tc.input, result.Type)
			}
			if result.AppliesTo != tc.expected.AppliesTo {
				t.Errorf("Expected AppliesTo %v for input %q, got %v", tc.expected.AppliesTo, tc.input, result.AppliesTo)
			}
			if len(result.Rules) != len(tc.expected.Rules) {
				t.Errorf("Expected %d rules for input %q, got %d", len(tc.expected.Rules), tc.input, len(result.Rules))
				continue
			}
			for i, rule := range result.Rules {
				if rule != tc.expected.Rules[i] {
					t.Errorf("Expected rule %q at index %d for input %q, got %q", tc.expected.Rules[i], i, tc.input, rule)
				}
			}
		}
	}
}

func TestRuleDisableTracker(t *testing.T) {
	tracker := NewRuleDisableTracker()
	
	// Test file-wide disable
	tracker.fileDisabled = append(tracker.fileDisabled, core.TextPos(10))
	
	// Test specific rule disable
	tracker.disabledRules["test-rule"] = append(tracker.disabledRules["test-rule"], 
		core.NewTextRange(20, 30))
	
	// Test cases
	testCases := []struct {
		ruleName string
		pos      core.TextPos
		expected bool
	}{
		{"any-rule", core.TextPos(15), true},  // Should be disabled due to file-wide disable
		{"any-rule", core.TextPos(5), false},  // Before file-wide disable
		{"test-rule", core.TextPos(25), true}, // Within specific rule disable range 
		{"test-rule", core.TextPos(35), true}, // After specific rule disable range, but file-wide disabled
		{"other-rule", core.TextPos(25), true}, // After file-wide disable position
	}
	
	for _, tc := range testCases {
		result := tracker.IsRuleDisabled(tc.ruleName, tc.pos)
		if result != tc.expected {
			t.Errorf("Expected IsRuleDisabled(%q, %d) = %v, got %v", 
				tc.ruleName, tc.pos, tc.expected, result)
		}
	}
}