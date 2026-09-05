package rule

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func parseInlineGlobalsForTest(t *testing.T, source string) (map[string]utils.GlobalAccess, []InlineGlobal) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)

	return ParseInlineGlobals(sourceFile, NewCommentStore(sourceFile))
}

func assertInlineGlobals(t *testing.T, source string, got []InlineGlobal, want []struct {
	name      string
	access    utils.GlobalAccess
	positions []int
}) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d inline globals, want %d: %#v", len(got), len(want), got)
	}
	for i, expected := range want {
		actual := got[i]
		if actual.Name != expected.name || actual.Access != expected.access {
			t.Fatalf("inline global %d = (%q, %v), want (%q, %v)", i, actual.Name, actual.Access, expected.name, expected.access)
		}
		if len(actual.NameRanges) != len(expected.positions) {
			t.Fatalf("inline global %q has %d ranges, want %d", actual.Name, len(actual.NameRanges), len(expected.positions))
		}
		for rangeIndex, textRange := range actual.NameRanges {
			if textRange.Pos() != expected.positions[rangeIndex] || textRange.End() != expected.positions[rangeIndex]+len(expected.name) {
				t.Errorf("inline global %q range %d = %d:%d, want %d:%d", actual.Name, rangeIndex, textRange.Pos(), textRange.End(), expected.positions[rangeIndex], expected.positions[rangeIndex]+len(expected.name))
			}
			if rangeText := source[textRange.Pos():textRange.End()]; rangeText != expected.name {
				t.Errorf("inline global %q range %d contains %q", actual.Name, rangeIndex, rangeText)
			}
		}
	}
}

func TestParseInlineGlobals_MetadataAndSourceFiltering(t *testing.T) {
	source := "const fake = '/*global stringName*/';\n" +
		"// global lineName\n" +
		"/* Global upperName */\n" +
		"/* globalConfig prefixName */\n" +
		"/*globals foo, Array:readonly, offName:off, foo:writable -- reason, ignored */\n" +
		"/*global offName, foo:off */\n"

	values, globals := parseInlineGlobalsForTest(t, source)
	wantValues := map[string]utils.GlobalAccess{"foo": utils.GlobalAccessOff, "Array": utils.GlobalAccessReadonly, "offName": utils.GlobalAccessReadonly}
	if !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("values = %#v, want %#v", values, wantValues)
	}

	assertInlineGlobals(t, source, globals, []struct {
		name      string
		access    utils.GlobalAccess
		positions []int
	}{
		{name: "foo", access: utils.GlobalAccessOff, positions: []int{strings.Index(source, "foo, Array"), strings.LastIndex(source, "foo:off")}},
		{name: "Array", access: utils.GlobalAccessReadonly, positions: []int{strings.Index(source, "Array:readonly")}},
		{name: "offName", access: utils.GlobalAccessReadonly, positions: []int{strings.Index(source, "offName:off"), strings.LastIndex(source, "offName, foo")}},
	})
}

func TestParseInlineGlobals_AccessLevels(t *testing.T) {
	source := "/*globals bare, ro:readonly, alias:readable, rw:writable, rw2:writeable, gone:off */"

	values, globals := parseInlineGlobalsForTest(t, source)
	wantValues := map[string]utils.GlobalAccess{
		"bare":  utils.GlobalAccessReadonly,
		"ro":    utils.GlobalAccessReadonly,
		"alias": utils.GlobalAccessReadonly,
		"rw":    utils.GlobalAccessWritable,
		"rw2":   utils.GlobalAccessWritable,
		"gone":  utils.GlobalAccessOff,
	}
	if !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("values = %#v, want %#v", values, wantValues)
	}
	if len(globals) != len(wantValues) {
		t.Fatalf("got %d inline globals, want %d: %#v", len(globals), len(wantValues), globals)
	}
}

// A setting that spells none of the three levels is ignored, as ESLint does
// after reporting the bad directive. A name with no earlier inline setting is
// left to the config or built-in setting; one that already has a setting keeps
// it, and the ignored comment contributes no name range.
func TestParseInlineGlobals_InvalidSettingIsIgnored(t *testing.T) {
	source := "/*global nonsense:bar, empty:, kept:writable */\n" +
		"/*global wasValid:off */\n" +
		"/*global wasValid:bar, kept:bar */\n" +
		"/*global lateValid:bar */\n" +
		"/*global lateValid:writable */"

	values, globals := parseInlineGlobalsForTest(t, source)
	wantValues := map[string]utils.GlobalAccess{
		"kept":      utils.GlobalAccessWritable,
		"wasValid":  utils.GlobalAccessOff,
		"lateValid": utils.GlobalAccessWritable,
	}
	if !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("values = %#v, want %#v", values, wantValues)
	}
	assertInlineGlobals(t, source, globals, []struct {
		name      string
		access    utils.GlobalAccess
		positions []int
	}{
		{name: "kept", access: utils.GlobalAccessWritable, positions: []int{strings.Index(source, "kept")}},
		{name: "wasValid", access: utils.GlobalAccessOff, positions: []int{strings.Index(source, "wasValid")}},
		{name: "lateValid", access: utils.GlobalAccessWritable, positions: []int{strings.LastIndex(source, "lateValid")}},
	})
}

func TestParseInlineGlobalsKeepsOrdinaryCommentsLazy(t *testing.T) {
	source := "// ordinary comment\nconst value = 1;\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		Path:     "/test.ts",
	}, source, core.ScriptKindTS)
	store := NewCommentStore(sourceFile)

	values, globals := ParseInlineGlobals(sourceFile, store)
	if values != nil || globals != nil {
		t.Fatalf("ParseInlineGlobals returned (%#v, %#v), want nil results", values, globals)
	}
	if store.scanned {
		t.Fatal("ordinary comment caused inline-global parsing to scan comments")
	}
}

func TestParseInlineGlobals_DuplicateAndOverrideSemantics(t *testing.T) {
	source := "/*global a:off, a, b, b:off */\n" +
		"/*global a:off, b */\n" +
		"/*global a */"

	values, globals := parseInlineGlobalsForTest(t, source)
	wantValues := map[string]utils.GlobalAccess{"a": utils.GlobalAccessReadonly, "b": utils.GlobalAccessReadonly}
	if !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("values = %#v, want %#v", values, wantValues)
	}

	firstA := strings.Index(source, "a:off")
	secondA := strings.Index(source[firstA+1:], "a:off") + firstA + 1
	lastA := strings.LastIndex(source, "a */")
	assertInlineGlobals(t, source, globals, []struct {
		name      string
		access    utils.GlobalAccess
		positions []int
	}{
		{name: "a", access: utils.GlobalAccessReadonly, positions: []int{firstA, secondA, lastA}},
		{name: "b", access: utils.GlobalAccessReadonly, positions: []int{strings.Index(source, "b, b:off"), strings.LastIndex(source, "b */")}},
	})
}

func TestParseInlineGlobals_UnicodeWhitespaceAndNames(t *testing.T) {
	zeroWidthSpace := string(rune(0x200B))
	source := "/*global\u0085notADirective*/\n" +
		"/*global" + zeroWidthSpace + "alsoNotADirective*/\n" +
		"/*globals\u00A0π\u2003:\uFEFFwritable,\u2028𐐀\u2029名:off */"

	values, globals := parseInlineGlobalsForTest(t, source)
	wantValues := map[string]utils.GlobalAccess{"π": utils.GlobalAccessWritable, "𐐀": utils.GlobalAccessReadonly, "名": utils.GlobalAccessOff}
	if !reflect.DeepEqual(values, wantValues) {
		t.Fatalf("values = %#v, want %#v", values, wantValues)
	}
	assertInlineGlobals(t, source, globals, []struct {
		name      string
		access    utils.GlobalAccess
		positions []int
	}{
		{name: "π", access: utils.GlobalAccessWritable, positions: []int{strings.LastIndex(source, "π")}},
		{name: "𐐀", access: utils.GlobalAccessReadonly, positions: []int{strings.LastIndex(source, "𐐀")}},
		{name: "名", access: utils.GlobalAccessOff, positions: []int{strings.LastIndex(source, "名")}},
	})
}

func TestParseInlineGlobals_CommentOnlyFileAfterShebang(t *testing.T) {
	source := "#!/usr/bin/env node\n/*global onlyComment */"
	values, globals := parseInlineGlobalsForTest(t, source)
	if !reflect.DeepEqual(values, map[string]utils.GlobalAccess{"onlyComment": utils.GlobalAccessReadonly}) {
		t.Fatalf("values = %#v, want onlyComment declared", values)
	}
	assertInlineGlobals(t, source, globals, []struct {
		name      string
		access    utils.GlobalAccess
		positions []int
	}{
		{name: "onlyComment", access: utils.GlobalAccessReadonly, positions: []int{strings.Index(source, "onlyComment")}},
	})
}

func TestParseInlineGlobals_TrailingMultilineComment(t *testing.T) {
	source := "const value = 1; /*globals foo,\r\n    Array */\nfoo; Array;"
	values, globals := parseInlineGlobalsForTest(t, source)
	if !reflect.DeepEqual(values, map[string]utils.GlobalAccess{"foo": utils.GlobalAccessReadonly, "Array": utils.GlobalAccessReadonly}) {
		t.Fatalf("values = %#v, want foo and Array declared", values)
	}
	assertInlineGlobals(t, source, globals, []struct {
		name      string
		access    utils.GlobalAccess
		positions []int
	}{
		{name: "foo", access: utils.GlobalAccessReadonly, positions: []int{strings.Index(source, "foo,")}},
		{name: "Array", access: utils.GlobalAccessReadonly, positions: []int{strings.Index(source, "Array */")}},
	})
}

func TestParseInlineGlobals_NoDirective(t *testing.T) {
	for _, source := range []string{
		"",
		"const text = '/*global stringName*/';",
		"// global lineName\nconst value = 1;",
	} {
		values, globals := parseInlineGlobalsForTest(t, source)
		if values != nil || globals != nil {
			t.Errorf("ParseInlineGlobals(%q) = (%#v, %#v), want nil results", source, values, globals)
		}
	}
}

func TestMatchInlineGlobalsDirective(t *testing.T) {
	zeroWidthSpace := string(rune(0x200B))
	tests := []struct {
		name         string
		input        string
		expectedRest string
		expectedOK   bool
	}{
		{name: "global with names", input: "global foo, bar", expectedRest: "foo, bar", expectedOK: true},
		{name: "globals with names", input: "globals foo, bar", expectedRest: "foo, bar", expectedOK: true},
		{name: "bare global", input: "global", expectedOK: true},
		{name: "tab separator", input: "global\tfoo", expectedRest: "foo", expectedOK: true},
		{name: "unicode separator", input: "global\u00A0foo", expectedRest: "foo", expectedOK: true},
		{name: "BOM separator", input: "global\uFEFFfoo", expectedRest: "foo", expectedOK: true},
		{name: "not a directive", input: "eslint-disable no-console"},
		{name: "lookalike prefix", input: "globalConfig setup"},
		{name: "comma after label", input: "global,foo"},
		{name: "upper-case label", input: "Global foo"},
		{name: "next-line is not ECMAScript whitespace", input: "global\u0085foo"},
		{name: "zero-width space is not ECMAScript whitespace", input: "global" + zeroWidthSpace + "foo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rest, ok := matchInlineGlobalsDirective(test.input)
			if ok != test.expectedOK {
				t.Fatalf("ok = %v, want %v", ok, test.expectedOK)
			}
			if ok && rest != test.expectedRest {
				t.Errorf("rest = %q, want %q", rest, test.expectedRest)
			}
		})
	}
}

func TestParseGlobalNameList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{name: "empty", input: "", expected: map[string]string{}},
		{name: "single bare name", input: "foo", expected: map[string]string{"foo": ""}},
		{name: "comma separated", input: "foo, bar", expected: map[string]string{"foo": "", "bar": ""}},
		{name: "whitespace separated", input: "foo bar", expected: map[string]string{"foo": "", "bar": ""}},
		{name: "settings", input: "foo:readonly, bar:writable, baz:off", expected: map[string]string{"foo": "readonly", "bar": "writable", "baz": "off"}},
		{name: "spaces around separators", input: "  foo : writable ,  bar  ", expected: map[string]string{"foo": "writable", "bar": ""}},
		{name: "duplicate uses last setting", input: "foo:off foo:writable", expected: map[string]string{"foo": "writable"}},
		{name: "only first value segment matters", input: "foo:off:ignored bar:readonly:ignored", expected: map[string]string{"foo": "off", "bar": "readonly"}},
		{name: "unicode whitespace", input: "π\u2003:\uFEFFwritable\u2028名:off", expected: map[string]string{"π": "writable", "名": "off"}},
		{name: "next-line remains part of name", input: "foo\u0085bar", expected: map[string]string{"foo\u0085bar": ""}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseGlobalNameList(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("parseGlobalNameList(%q) = %#v, want %#v", test.input, result, test.expected)
			}
		})
	}
}
