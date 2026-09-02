package prefer_array_flat_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_array_flat"
)

// TestPreferArrayFlatExtrasSchema locks the public option schema to unicorn
// v74, including string-only function entries, tuple length, uniqueness, and
// additionalProperties. Behavioral cases live in the sibling upstream and
// extras files.
func TestPreferArrayFlatExtrasSchema(t *testing.T) {
	tests := []struct {
		name    string
		options []any
		wantErr bool
	}{
		{name: "no options"},
		{name: "empty object", options: []any{map[string]any{}}},
		{
			name: "string functions",
			options: []any{map[string]any{
				"functions": []any{"flat", "utils.flat"},
			}},
		},
		{
			name: "function items must be strings",
			options: []any{map[string]any{
				"functions": []any{1.0, nil, true, map[string]any{}},
			}},
			wantErr: true,
		},
		{
			name: "duplicate functions",
			options: []any{map[string]any{
				"functions": []any{"flat", "flat"},
			}},
			wantErr: true,
		},
		{
			name: "functions must be array",
			options: []any{map[string]any{
				"functions": "flat",
			}},
			wantErr: true,
		},
		{
			name: "unknown property",
			options: []any{map[string]any{
				"unknown": true,
			}},
			wantErr: true,
		},
		{
			name:    "only one option object",
			options: []any{map[string]any{}, map[string]any{}},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := prefer_array_flat.PreferArrayFlatRule.Schema.Validate(test.options)
			if (err != nil) != test.wantErr {
				t.Fatalf("Schema.Validate(%#v) error = %v, wantErr %v",
					test.options, err, test.wantErr)
			}
		})
	}
}
