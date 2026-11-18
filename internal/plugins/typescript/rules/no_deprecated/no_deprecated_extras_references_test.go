package no_deprecated

import (
	"encoding/json"
	"strings"
	"testing"
)

// These expectations preserve deprecated value reads and avoid diagnostics on
// annotated prototype definitions; see Differences from upstream in the docs.
func TestNoDeprecatedValueAndPrototypeReferences(t *testing.T) {
	type diagnostic struct {
		id, message string
		start, end  int
	}
	for _, tc := range []struct {
		name, code, filename, options string
		files                         map[string]string
		want                          []diagnostic
	}{
		{name: "key-plain-and-parenthesized", code: `/** @deprecated variable */
const key='old';
const obj={old:1};obj[key];obj[(key)];`, filename: "main.ts", want: []diagnostic{{"deprecatedWithReason", "`key` is deprecated. variable", 67, 70}, {"deprecatedWithReason", "`key` is deprecated. variable", 77, 80}}},
		{name: "key-wrapper-call-optional", code: `/** @deprecated variable */
const key='old';
const obj={old(){}};obj?.[key]();obj[((key))]();obj[key as 'old']();`, filename: "main.ts", want: []diagnostic{{"deprecatedWithReason", "`key` is deprecated. variable", 71, 74}, {"deprecatedWithReason", "`key` is deprecated. variable", 84, 87}, {"deprecatedWithReason", "`key` is deprecated. variable", 97, 100}}},
		{name: "key-and-property", code: `/** @deprecated variable */
const key='old';
const obj={
/** @deprecated property */ old:1
};
obj[key];obj[(key)];`, filename: "main.ts", want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. property", 98, 101}, {"deprecatedWithReason", "`key` is deprecated. variable", 98, 101}, {"deprecatedWithReason", "`old` is deprecated. property", 108, 111}, {"deprecatedWithReason", "`key` is deprecated. variable", 108, 111}}},
		{name: "property-only", code: `const key='old';
const obj={
/** @deprecated property */ old:1
};
obj[key];`, filename: "main.ts", want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. property", 70, 73}}},
		{name: "allow-key", code: `/** @deprecated variable */
const key='old';
const obj={
/** @deprecated property */ old:1
};
obj[key];`, filename: "main.ts", options: `[{"allow": ["key"]}]`, want: []diagnostic{{"deprecatedWithReason", "`old` is deprecated. property", 98, 101}}},
		{name: "allow-receiver", code: `/** @deprecated variable */
const key='old';
interface Holder {
/** @deprecated property */ old:number;
}
declare const obj:Holder;obj[key];`, filename: "main.ts", options: `[{"allow": ["Holder"]}]`, want: []diagnostic{{"deprecatedWithReason", "`key` is deprecated. variable", 135, 138}}},
		{name: "non-literal-key", code: `/** @deprecated variable */
declare const key:string;declare const obj:Record<string,number>;obj[key];`, filename: "main.ts", want: []diagnostic{{"deprecatedWithReason", "`key` is deprecated. variable", 97, 100}}},
		{name: "key-on-write-delete", code: `/** @deprecated variable */
const key='old';
const obj={old:1};obj[key]=2;delete obj[key];`, filename: "main.ts", want: []diagnostic{{"deprecatedWithReason", "`key` is deprecated. variable", 67, 70}, {"deprecatedWithReason", "`key` is deprecated. variable", 85, 88}}},
		{name: "key-shadowed", code: `/** @deprecated variable */
const key='old';
const obj={old:1};function f(key:'old'){return obj[key];}`, filename: "main.ts", want: []diagnostic{}},
		{name: "prototype-definition-only", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
`, filename: "main.js", want: []diagnostic{}},
		{name: "prototype-call", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
new Box().legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 95, 101}}},
		{name: "prototype-read", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
new Box().legacy;`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 95, 101}}},
		{name: "prototype-direct-prototype", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
Box.prototype.legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 99, 105}}},
		{name: "prototype-instance-alias", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
let box=new Box();const alias=box;alias.legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 125, 131}}},
		{name: "prototype-instance-reassigned", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
let box=new Box();box={legacy(){}};box.legacy();`, filename: "main.js", want: []diagnostic{}},
		{name: "prototype-constructor-reassigned", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
Box=function Other(){};new Box().legacy();`, filename: "main.js", want: []diagnostic{}},
		{name: "prototype-constructor-shadowed", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
function f(){function Box(){};new Box().legacy();}`, filename: "main.js", want: []diagnostic{}},
		{name: "prototype-unknown-receiver", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
function f(box){box.legacy();}`, filename: "main.js", want: []diagnostic{}},
		{name: "prototype-subsequent-write", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
Box.prototype.legacy=function(){};new Box().legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 99, 105}, {"deprecatedWithReason", "`legacy` is deprecated. Use current.", 129, 135}}},
		{name: "prototype-compound-write", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
Box.prototype.legacy ||= function(){};`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 99, 105}}},
		{name: "prototype-function-expression", code: `const Box=function(){};
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
new Box().legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 101, 107}}},
		{name: "prototype-constructor-alias", code: `function Box(){}
const Alias=Box;
/** @deprecated old */
Alias.prototype.legacy=function(){};
new Box().legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 104, 110}}},
		{name: "prototype-other-constructor", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
function Other(){};Other.prototype.legacy=function(){};new Other().legacy();new Box().legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 171, 177}}},
		{name: "prototype-computed", code: `function Box() {}
/** @deprecated Use current. */
Box.prototype['legacy']=function(){};
new Box()['legacy']();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 98, 106}}},
		{name: "prototype-parenthesized", code: `function Box() {}
/** @deprecated Use current. */
(Box.prototype).legacy=(function(){});
(new Box()).legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 101, 107}}},
		{name: "prototype-read-before-definition", code: `function Box(){}
new Box().legacy();
/** @deprecated old */
Box.prototype.legacy=function(){};`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. old", 27, 33}}},
		{name: "prototype-no-deprecation", code: `function Box(){};Box.prototype.legacy=function(){};new Box().legacy();`, filename: "main.js", want: []diagnostic{}},
		{name: "prototype-class", code: `class Box {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
new Box().legacy();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 90, 96}}},
		{name: "prototype-computed-deprecated-key", code: `function Box(){}
/** @deprecated key */
const key='legacy';
/** @deprecated method */
Box.prototype[key]=function(){};
new Box()[key]();`, filename: "main.js", want: []diagnostic{{"deprecatedWithReason", "`key` is deprecated. key", 100, 103}, {"deprecatedWithReason", "`legacy` is deprecated. method", 129, 132}, {"deprecatedWithReason", "`key` is deprecated. key", 129, 132}}},
		{name: "prototype-cross-file", code: `import {Box} from './box.js';
new Box().legacy();`, filename: "main.js", files: map[string]string{"box.js": `export function Box() {}
/** @deprecated Use current. */
Box.prototype.legacy=function(){};
`}, want: []diagnostic{{"deprecatedWithReason", "`legacy` is deprecated. Use current.", 40, 46}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{tc.filename: tc.code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"target":"esnext","module":"commonjs","strict":true,"types":[]}}`}
			for name, text := range tc.files {
				files[name] = text
			}
			var options any
			if tc.options != "" {
				if err := json.Unmarshal([]byte(tc.options), &options); err != nil {
					t.Fatal(err)
				}
			}
			got := runNoDeprecatedDiagnosticsForFiles(t, files, tc.filename, options)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d diagnostics, want %d: %#v", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				d := got[i]
				if d.Message.Id != w.id || d.Message.Description != w.message || d.Range.Pos() != w.start || d.Range.End() != w.end {
					t.Errorf("got %s %q [%d,%d), want %s %q [%d,%d)", d.Message.Id, d.Message.Description, d.Range.Pos(), d.Range.End(), w.id, w.message, w.start, w.end)
				}
				if d.FixesPtr != nil || d.Suggestions != nil {
					t.Error("unexpected fix or suggestion")
				}
			}
		})
	}
}

func TestNoDeprecatedPrototypeReplacement(t *testing.T) {
	for _, code := range []string{
		`function Box(){}
/** @deprecated old */
Box.prototype.legacy=function(){};
Box.prototype={legacy(){}};
new Box().legacy();`,
		`export function Box(){}
/** @deprecated old */
Box.prototype.legacy=function(){};
Box=function Other(){};
new Box().legacy();`,
	} {
		got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.js": code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"module":"commonjs","types":[]}}`}, "main.js", nil)
		if len(got) != 0 {
			t.Fatalf("unexpected diagnostics: %#v", got)
		}
	}
	const code = `function Box(){}
Box.prototype={};
/** @deprecated old */
Box.prototype.legacy=function(){};
new Box().legacy();`
	got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.js": code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"module":"commonjs","types":[]}}`}, "main.js", nil)
	start := strings.LastIndex(code, "legacy")
	if len(got) != 1 || got[0].Message.Id != "deprecatedWithReason" || got[0].Message.Description != "`legacy` is deprecated. old" || got[0].Range.Pos() != start || got[0].Range.End() != start+len("legacy") {
		t.Fatalf("unexpected diagnostics: %#v", got)
	}
}

func TestNoDeprecatedPrototypeConstructorResult(t *testing.T) {
	for _, tc := range []struct {
		body    string
		reports bool
	}{
		{"return {legacy(){}};", false},
		{"return 1;", true},
		{"return this;", true},
		{"function nested(){return {legacy(){}}}", true},
	} {
		code := "function Box(){" + tc.body + "}\n/** @deprecated old */\nBox.prototype.legacy=function(){};\nnew Box().legacy();"
		got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.js": code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"module":"commonjs","types":[]}}`}, "main.js", nil)
		if !tc.reports {
			if len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", got)
			}
			continue
		}
		start := strings.LastIndex(code, "legacy")
		if len(got) != 1 || got[0].Message.Description != "`legacy` is deprecated. old" || got[0].Range.Pos() != start || got[0].Range.End() != start+len("legacy") {
			t.Fatalf("unexpected diagnostics: %#v", got)
		}
	}
}

func TestNoDeprecatedPrototypeAllow(t *testing.T) {
	for _, tc := range []struct {
		name, access string
		reports      bool
	}{
		{"legacy", "box.legacy()", false},
		{"Box", "box['legacy']()", false},
		{"Other", "box['legacy']()", true},
		{"Box", "box.legacy()", true},
	} {
		code := "function Box(){}\n/** @deprecated old */\nBox.prototype.legacy=function(){};\nconst box=new Box();" + tc.access + ";"
		got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.js": code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"module":"commonjs","types":[]}}`}, "main.js", []any{map[string]any{"allow": []any{tc.name}}})
		if !tc.reports {
			if len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", got)
			}
			continue
		}
		span := "legacy"
		if strings.Contains(tc.access, "[") {
			span = "'legacy'"
		}
		start := strings.LastIndex(code, span)
		if len(got) != 1 || got[0].Message.Description != "`legacy` is deprecated. old" || got[0].Range.Pos() != start || got[0].Range.End() != start+len(span) {
			t.Fatalf("unexpected diagnostics: %#v", got)
		}
	}
}

func TestNoDeprecatedPrototypePackageAllow(t *testing.T) {
	const code = "import {Box} from 'demo-box'; new Box().legacy();"
	files := map[string]string{
		"main.js":                            code,
		"tsconfig.json":                      `{"compilerOptions":{"allowJs":true,"checkJs":true,"maxNodeModuleJsDepth":1,"module":"commonjs","types":[]}}`,
		"node_modules/demo-box/package.json": `{"name":"demo-box","version":"1.0.0","main":"index.js"}`,
		"node_modules/demo-box/index.js": `export function Box(){}
/** @deprecated old */
Box.prototype.legacy=function(){};`,
	}
	for _, pkg := range []string{"demo-box", "unrelated"} {
		got := runNoDeprecatedDiagnosticsForFiles(t, files, "main.js", []any{map[string]any{"allow": []any{map[string]any{"from": "package", "package": pkg, "name": "legacy"}}}})
		if pkg == "demo-box" {
			if len(got) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", got)
			}
			continue
		}
		start := strings.LastIndex(code, "legacy")
		if len(got) != 1 || got[0].Message.Description != "`legacy` is deprecated. old" || got[0].Range.Pos() != start || got[0].Range.End() != start+len("legacy") {
			t.Fatalf("unexpected diagnostics: %#v", got)
		}
	}
}

func TestNoDeprecatedOrdinaryPrototypeNamedProperty(t *testing.T) {
	const code = `const obj={prototype:{
/** @deprecated old */
legacy(){}
}};
/** @deprecated assignment */
obj.prototype.legacy=function(){};`
	got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.js": code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"module":"commonjs","types":[]}}`}, "main.js", nil)
	start := strings.LastIndex(code, "legacy")
	if len(got) != 1 || got[0].Message.Description != "`legacy` is deprecated. old" || got[0].Range.Pos() != start || got[0].Range.End() != start+len("legacy") {
		t.Fatalf("unexpected diagnostics: %#v", got)
	}
}

func TestNoDeprecatedPrototypeBigIntKeys(t *testing.T) {
	const code = `function Box(){}
/** @deprecated old */
Box.prototype[1n]=function(){};
new Box()[1n]();
new Box()[2n]();
new Box()['1']();`
	got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.js": code, "tsconfig.json": `{"compilerOptions":{"allowJs":true,"checkJs":true,"target":"esnext","module":"commonjs","types":[]}}`}, "main.js", nil)
	if len(got) != 2 {
		t.Fatalf("unexpected diagnostics: %#v", got)
	}
	for i, span := range []string{"1n", "'1'"} {
		start := strings.LastIndex(code, span)
		if got[i].Message.Id != "deprecatedWithReason" || got[i].Message.Description != "`1` is deprecated. old" || got[i].Range.Pos() != start || got[i].Range.End() != start+len(span) {
			t.Fatalf("unexpected diagnostic: %#v", got[i])
		}
	}
}
