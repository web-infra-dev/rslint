package rule

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// matchInlineGlobalsDirective
// ---------------------------------------------------------------------------

func TestMatchInlineGlobalsDirective(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedRest string
		expectedOk   bool
	}{
		{"global with names", "global foo, bar", "foo, bar", true},
		{"globals with names", "globals foo, bar", "foo, bar", true},
		{"bare global", "global", "", true},
		{"bare globals", "globals", "", true},
		{"global with tab", "global\tfoo", "foo", true},
		{"global with newline", "global\nfoo", "foo", true},
		{"not a directive", "eslint-disable no-console", "", false},
		{"lookalike prefix is not a directive", "globalConfig setup here", "", false},
		{"globalsConfig lookalike is not a directive", "globalsConfig setup here", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rest, ok := matchInlineGlobalsDirective(tt.input)
			if ok != tt.expectedOk {
				t.Fatalf("ok = %v, want %v", ok, tt.expectedOk)
			}
			if ok && rest != tt.expectedRest {
				t.Errorf("rest = %q, want %q", rest, tt.expectedRest)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseGlobalNameList
// ---------------------------------------------------------------------------

func TestParseGlobalNameList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{"empty", "", map[string]string{}},
		{"single bare name", "foo", map[string]string{"foo": ""}},
		{"comma separated", "foo, bar", map[string]string{"foo": "", "bar": ""}},
		{"whitespace separated", "foo bar", map[string]string{"foo": "", "bar": ""}},
		{"with settings", "foo:readonly, bar:writable, baz:off", map[string]string{"foo": "readonly", "bar": "writable", "baz": "off"}},
		{"mixed separators and settings", "foo:writable bar, baz:off", map[string]string{"foo": "writable", "bar": "", "baz": "off"}},
		{"extra whitespace", "  foo : writable ,  bar  ", map[string]string{"foo": "writable", "bar": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGlobalNameList(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseGlobalNameList(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
