package config

import (
	"reflect"
	"testing"
)

// TestParseArrayRuleConfig_OptionShapes pins how an array-style rule config's
// post-severity args map to RuleConfig.Options. The load-bearing case is a lone
// option that is itself an array (["error", ["a","b"]]): it must keep the outer
// wrapper ([["a","b"]]) so the eslint-plugin dispatch reconstructs
// context.options == [["a","b"]] instead of collapsing it to ["a","b"], which is
// indistinguishable from a two-element option list (["error","a","b"]).
func TestParseArrayRuleConfig_OptionShapes(t *testing.T) {
	tests := []struct {
		name string
		in   []interface{}
		want interface{}
	}{
		{
			name: "single string option unwraps to the value",
			in:   []interface{}{"error", "both"},
			want: "both",
		},
		{
			name: "single object option unwraps to the value",
			in:   []interface{}{"error", map[string]interface{}{"k": float64(1)}},
			want: map[string]interface{}{"k": float64(1)},
		},
		{
			name: "single array option keeps its wrapper",
			in:   []interface{}{"error", []interface{}{"a", "b"}},
			want: []interface{}{[]interface{}{"a", "b"}},
		},
		{
			name: "multiple options pass through as the args list",
			in:   []interface{}{"error", "a", "b"},
			want: []interface{}{"a", "b"},
		},
		{
			name: "multiple options including an object",
			in:   []interface{}{"error", "both", map[string]interface{}{"k": float64(1)}},
			want: []interface{}{"both", map[string]interface{}{"k": float64(1)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := parseArrayRuleConfig(tt.in)
			if rc == nil {
				t.Fatal("parseArrayRuleConfig returned nil")
			}
			if rc.Level != "error" {
				t.Errorf("Level = %q, want \"error\"", rc.Level)
			}
			if !reflect.DeepEqual(rc.Options, tt.want) {
				t.Errorf("Options = %#v, want %#v", rc.Options, tt.want)
			}
		})
	}
}
