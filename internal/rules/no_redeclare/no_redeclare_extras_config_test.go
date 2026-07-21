package no_redeclare

import "testing"

func TestNoRedeclareParseOptionsMatrix(t *testing.T) {
	tests := []struct {
		name string
		raw  []any
		want options
	}{
		{name: "omitted", want: coreDefaults()},
		{name: "empty options array", raw: []any{}, want: coreDefaults()},
		{name: "empty object", raw: []any{map[string]any{}}, want: coreDefaults()},
		{
			name: "builtin globals enabled",
			raw:  []any{map[string]any{"builtinGlobals": true}},
			want: coreDefaults(),
		},
		{
			name: "builtin globals disabled",
			raw:  []any{map[string]any{"builtinGlobals": false}},
			want: options{builtinGlobals: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseOptionsWith(test.raw, coreDefaults(), false)
			if got != test.want {
				t.Fatalf("parseOptionsWith() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNoRedeclareSchemaMatrix(t *testing.T) {
	if NoRedeclareRule.Schema == nil {
		t.Fatal("no-redeclare must declare its upstream option schema")
	}

	valid := []struct {
		name    string
		options []any
	}{
		{name: "omitted"},
		{name: "empty options array", options: []any{}},
		{name: "empty object", options: []any{map[string]any{}}},
		{name: "builtin globals true", options: []any{map[string]any{"builtinGlobals": true}}},
		{name: "builtin globals false", options: []any{map[string]any{"builtinGlobals": false}}},
	}
	for _, test := range valid {
		t.Run("valid/"+test.name, func(t *testing.T) {
			if err := NoRedeclareRule.Schema.Validate(test.options); err != nil {
				t.Fatalf("unexpected schema error: %v", err)
			}
		})
	}

	invalid := []struct {
		name    string
		options []any
	}{
		{name: "non-object option", options: []any{true}},
		{name: "null option", options: []any{nil}},
		{name: "non-boolean value", options: []any{map[string]any{"builtinGlobals": "true"}}},
		{name: "unknown property", options: []any{map[string]any{"ignoreDeclarationMerge": true}}},
		{name: "too many options", options: []any{map[string]any{}, map[string]any{}}},
	}
	for _, test := range invalid {
		t.Run("invalid/"+test.name, func(t *testing.T) {
			if err := NoRedeclareRule.Schema.Validate(test.options); err == nil {
				t.Fatal("expected schema validation to reject options")
			}
		})
	}
}
