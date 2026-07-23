package rule

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
)

func TestNormalizeOptions(t *testing.T) {
	// nil → empty (non-nil) slice: native rules key on len==0 for their defaults
	// branch, and the eslint-plugin host serializes context.options to JSON and
	// needs [] not null.
	got := NormalizeOptions(nil)
	if got == nil || len(got) != 0 {
		t.Errorf("nil → empty non-nil slice, got %#v", got)
	}

	// eslint-format array → passthrough preserving elements.
	arr := NormalizeOptions([]interface{}{"a", "b"})
	if len(arr) != 2 || arr[0] != "a" || arr[1] != "b" {
		t.Errorf("array → passthrough, got %v", arr)
	}

	// empty array → len 0 so the defaults branch still fires.
	if e := NormalizeOptions([]interface{}{}); len(e) != 0 {
		t.Errorf("empty array → len 0, got %v", e)
	}

	// bare single option (config.rules unwrapped a single map) → wrapped
	// 1-element array, so optArray[0] yields the option object.
	single := NormalizeOptions(map[string]interface{}{"k": 1})
	if len(single) != 1 || !reflect.DeepEqual(single[0], map[string]interface{}{"k": 1}) {
		t.Errorf("single map → [map], got %v", single)
	}

	// bare single string option → wrapped (e.g. no-cond-assign "always").
	str := NormalizeOptions("always")
	if len(str) != 1 || str[0] != "always" {
		t.Errorf("single string → [string], got %v", str)
	}

	// A lone array-valued option (["error", ["a","b"]] → Options [["a","b"]])
	// must surface as context.options == [["a","b"]]: one element that is the
	// array itself, not the two strings flattened into the options list.
	nested := NormalizeOptions([]interface{}{[]interface{}{"a", "b"}})
	if len(nested) != 1 || !reflect.DeepEqual(nested[0], []interface{}{"a", "b"}) {
		t.Errorf("lone array option → [[a,b]], got %v", nested)
	}
}

func TestRuleContextReportWithoutReporterPanics(t *testing.T) {
	textRange := core.NewTextRange(0, 0)
	message := RuleMessage{Description: "must not be dropped silently"}
	tests := []struct {
		name   string
		report func(*RuleContext)
	}{
		{name: "range", report: func(ctx *RuleContext) { ctx.ReportRange(textRange, message) }},
		{name: "range fixes", report: func(ctx *RuleContext) { ctx.ReportRangeWithFixes(textRange, message) }},
		{name: "range suggestions", report: func(ctx *RuleContext) { ctx.ReportRangeWithSuggestions(textRange, message) }},
		{name: "range combined", report: func(ctx *RuleContext) { ctx.ReportRangeWithFixesAndSuggestions(textRange, message, nil, nil) }},
		{name: "node", report: func(ctx *RuleContext) { ctx.ReportNode(nil, message) }},
		{name: "node fixes", report: func(ctx *RuleContext) { ctx.ReportNodeWithFixes(nil, message) }},
		{name: "node suggestions", report: func(ctx *RuleContext) { ctx.ReportNodeWithSuggestions(nil, message) }},
		{name: "node combined", report: func(ctx *RuleContext) { ctx.ReportNodeWithFixesAndSuggestions(nil, message, nil, nil) }},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got != "rule: uninitialized RuleContext reporter" {
					t.Fatalf("panic = %v, want uninitialized reporter failure", got)
				}
			}()

			var ctx RuleContext
			testCase.report(&ctx)
		})
	}
}
