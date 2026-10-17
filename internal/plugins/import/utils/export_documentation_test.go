package utils

import "testing"

// Expectations come from Doctrine 2.1.0 with unwrap:true, used by import v2.32.0.
func TestExportDocTags(t *testing.T) {
	tests := []struct {
		name, text, description string
		deprecated, module      bool
	}{
		{"no tags", "* ordinary documentation", "", false, false},
		{"empty", "* @deprecated", "", true, false},
		{"dash", "* @deprecated - use new", "use new", true, false},
		{"multiline", "* @deprecated first\n *   second\n * @returns {number} ok", "first\n  second", true, false},
		{"CRLF", "* @deprecated first\r\n * second", "first\r\nsecond", true, false},
		{"Unicode whitespace", "*\u00a0@deprecated\u00a0old", "old", true, false},
		{"Unicode newline", "* @deprecated first\u2028 * second", "first\u2028second", true, false},
		{"duplicate", "* @deprecated first\n * @deprecated second", "first", true, false},
		{"blank paragraph", "* @deprecated first\n *\n * second", "first\n\nsecond", true, false},
		{"module", "* @module sample\n * @deprecated old module", "old module", true, true},
		{"case-sensitive tag", "* @Deprecated old", "", false, false},
		{"inline text", "* prose @deprecated ignored", "", false, false},
		{"module without deprecation", "* @module sample", "", false, true},
		// Deliberate documented difference: malformed unrelated types are tolerated.
		{"malformed unrelated type", "* @param {broken\n * @deprecated old", "old", true, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deprecation, module := parseExportDoc(test.text)
			if module != test.module || (deprecation != nil) != test.deprecated {
				t.Fatalf("got (%#v, %v), want deprecated=%v, module=%v", deprecation, module, test.deprecated, test.module)
			}
			if deprecation != nil && deprecation.Description != test.description {
				t.Fatalf("description %q, want %q", deprecation.Description, test.description)
			}
		})
	}
}

func TestDocumentationSettingsCacheIdentity(t *testing.T) {
	defaults := moduleSettingsKey(nil)
	disabled := moduleSettingsKey(map[string]any{"import/docstyle": []any{}})
	jsFirst := moduleSettingsKey(map[string]any{"import/docstyle": []any{"jsdoc", "tomdoc"}})
	tomFirst := moduleSettingsKey(map[string]any{"import/docstyle": []any{"tomdoc", "jsdoc"}})
	keys := []string{defaults, disabled, jsFirst, tomFirst}
	for i, key := range keys {
		for j := range i {
			if key == keys[j] {
				t.Fatalf("documentation settings %d and %d share a cache key", i, j)
			}
		}
	}
}
