package config

import (
	"reflect"
	"testing"
)

// TestParseArrayRuleConfig_OptionShapes pins how an array-style rule config's
// post-severity args map to RuleConfig.Options: always the raw remaining
// slice, with no bare-value collapsing. This also covers the load-bearing
// case of a lone option that is itself an array (["error", ["a","b"]]): it
// naturally stays wrapped as [["a","b"]], distinguishable from a two-element
// option list (["error","a","b"]) which becomes ["a","b"].
func TestParseArrayRuleConfig_OptionShapes(t *testing.T) {
	tests := []struct {
		name string
		in   []interface{}
		want []interface{}
	}{
		{
			name: "no options",
			in:   []interface{}{"error"},
			want: nil,
		},
		{
			name: "single string option stays wrapped",
			in:   []interface{}{"error", "both"},
			want: []interface{}{"both"},
		},
		{
			name: "single object option stays wrapped",
			in:   []interface{}{"error", map[string]interface{}{"k": float64(1)}},
			want: []interface{}{map[string]interface{}{"k": float64(1)}},
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
