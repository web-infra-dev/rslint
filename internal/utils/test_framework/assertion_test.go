package test_framework

import (
	"reflect"
	"testing"
)

func TestParseAssertionFunctionOptions(t *testing.T) {
	tests := []struct {
		name       string
		options    []any
		wantAssert []string
		wantBlocks []string
	}{
		{
			name:       "uses defaults when option is absent",
			wantAssert: []string{"expect", "assert"},
			wantBlocks: []string{},
		},
		{
			name:       "preserves an explicitly empty assertion list",
			options:    []any{map[string]interface{}{"assertFunctionNames": []interface{}{}}},
			wantAssert: []string{},
			wantBlocks: []string{},
		},
		{
			name: "parses both lists",
			options: []any{map[string]interface{}{
				"assertFunctionNames":          []interface{}{"verify", 1},
				"additionalTestBlockFunctions": []interface{}{"suite.test"},
			}},
			wantAssert: []string{"verify"},
			wantBlocks: []string{"suite.test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ParseAssertionFunctionOptions(test.options, []string{"expect", "assert"})
			if !reflect.DeepEqual(got.AssertFunctionNames, test.wantAssert) {
				t.Errorf("AssertFunctionNames = %#v, want %#v", got.AssertFunctionNames, test.wantAssert)
			}
			if !reflect.DeepEqual(got.AdditionalTestBlockFunctions, test.wantBlocks) {
				t.Errorf("AdditionalTestBlockFunctions = %#v, want %#v", got.AdditionalTestBlockFunctions, test.wantBlocks)
			}
		})
	}
}

func TestAssertPatterns(t *testing.T) {
	compiled := CompileAssertPatterns([]string{"expect", "request.**.expect", "VERIFY"})
	tests := []struct {
		name string
		want bool
	}{
		{"expect", true},
		{"expect.toBe", true},
		{"request.get.expect", true},
		{"request.get.user.expect", true},
		{"verify", true},
		{"request.expectation", false},
		{"", false},
	}

	for _, test := range tests {
		if got := MatchesAssertName(test.name, compiled); got != test.want {
			t.Errorf("MatchesAssertName(%q) = %v, want %v", test.name, got, test.want)
		}
	}

	if MatchesAssertName("expect", CompileAssertPatterns([]string{"("})) {
		t.Error("malformed pattern must not match")
	}
}
