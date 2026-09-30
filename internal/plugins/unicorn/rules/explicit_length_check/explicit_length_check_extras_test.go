// Additional scope, AST, mutation, safe-fix and edit-demand cases.
package explicit_length_check_test

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/explicit_length_check"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"reflect"
	"testing"
)

func TestExplicitLengthCheckExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_length_check.ExplicitLengthCheckRule, []rule_tester.ValidTestCase{
		// Unicorn folds constant receiver keys when recognizing shape/non-zero guards.
		{Code: `if (groups["a" + "b"].width && groups.ab.length) {}`, FileName: "case.js"},
		{Code: `if (groups[1 + 1].size && groups[2].size > 0) {}`, FileName: "case.js"},

		{Code: "if (foo?.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo?.bar.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo[\"length\"]) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class Box { #length = 0; check() { if (this.#length) {} } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "class Box { check() { if ((this).length) {} } }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Tsx: true, Code: "const x = <foo.length />;", FileName: "case.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "interface Box extends foo.length {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "type T = typeof foo.length;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (Boolean?.(foo.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (Boolean(...foo.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "function f(Boolean) { if ((Boolean)(foo.length)) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo.length as number) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if ((foo.length satisfies number)) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo.length! >= 1) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const n = foo.length ?? bar.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo.length === 0n) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo.length === -0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (foo.length === ZERO) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (box.width && !box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (box.length && box.depth) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (boxes[0].width && boxes[\"0\"].length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if ((box.length > 0 && ready) && box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "if (({length: -1}).length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "let box = {length: -1}; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: 9007199254740992}; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {size: 1n}; if (box.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {size: undefined}; if (box.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {}; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; box.length++; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; box.length += 2; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; delete box.length; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; Object.defineProperty(box, \"length\", {get() {return 2}}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; Object.defineProperty(box, \"height\", {value: 2}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; Object.defineProperties(box, {height: {value: 2}}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; Object.assign(box, {[key]: 2}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; Object.assign(box, {...values}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; let n = 2; box.length = n; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; for (box.length of []) {} if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; for (box.length of [1, \"x\"]) {} if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; [box.length = 2] = []; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; ({length: box.length} = {[key]: 2}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: -1}; class A { static { box.length = 2; } } if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		{Code: "const box = {length: 2}; class A { get field() { box.length = \"x\"; } } if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Tagged templates use Tag in tsgo, including when tracking receiver escapes.
		{Code: "const box = {length: -1}; box.length`x`; if (box.length) {}", FileName: "case.js", Output: []string{"const box = {length: -1}; box.length`x`; if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 46, EndLine: 1, EndColumn: 56}}},
		{Code: "const box = {size: -1}; (box.size)`x`; if (box.size) {}", FileName: "case.js", Output: []string{"const box = {size: -1}; (box.size)`x`; if (box.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 44, EndLine: 1, EndColumn: 52}}},
		// Preserve grouping and statement boundaries, including safer fixes than upstream.
		{Code: "const x = 1 + !items.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = 1 + (items.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
		{Code: "const x = Boolean(items.length) + 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (items.length > 0) + 1;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = 1 === !items.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = 1 === (items.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 30}}},
		{Code: "const x = Boolean(items.length).valueOf();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (items.length > 0).valueOf();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = Boolean(items.length)[key];", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (items.length > 0)[key];"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = +Boolean(items.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = +(items.length > 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 12, EndLine: 1, EndColumn: 33}}},
		{Code: "async function f(){return await Boolean(items.length);}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"async function f(){return await (items.length > 0);}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 33, EndLine: 1, EndColumn: 54}}},
		{Code: "const x = Boolean(items.length)();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (items.length > 0)();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = new (Boolean(items.length))();", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = new (items.length > 0)();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 16, EndLine: 1, EndColumn: 37}}},
		{Code: "const x = Boolean(items.length)`x`;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (items.length > 0)`x`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = Boolean(items.length)!;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (items.length > 0)!;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = <boolean>Boolean(items.length);", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = <boolean>(items.length > 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 41}}},
		{Code: "const x = items.length > 0 > 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}, Output: []string{"const x = (items.length !== 0) > 1;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 27}}},
		{Code: "foo()\n![].length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo()\n;[].length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 1, EndLine: 2, EndColumn: 11}}},
		{Code: "foo()\n!({length:1}).length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo()\n;({length:1}).length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 1, EndLine: 2, EndColumn: 21}}},
		{Code: "foo()\n!`x`.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo()\n;`x`.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 1, EndLine: 2, EndColumn: 12}}},
		{Code: "foo();\n![].length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"foo();\n[].length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 1, EndLine: 2, EndColumn: 11}}},
		{Code: "if (ready) ![].length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (ready) [].length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 12, EndLine: 1, EndColumn: 22}}},
		{Code: "function f(){return !\n[].length;}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function f(){return [].length === 0;}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 21, EndLine: 2, EndColumn: 10}}},
		{Code: "if (Boolean(items.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 26}}},
		{Code: "consume(Boolean(items.length));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"consume(items.length > 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 9, EndLine: 1, EndColumn: 30}}},
		{Code: "consume(object[Boolean(items.length)]);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"consume(object[items.length > 0]);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 16, EndLine: 1, EndColumn: 37}}},
		{Code: "const x = Boolean(items.length) && ready;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = items.length > 0 && ready;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 32}}},
		{Code: "const x = (1 + Boolean(items.length));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = (1 + (items.length > 0));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 16, EndLine: 1, EndColumn: 37}}},
		{Code: "const x = 1 + (Boolean(items.length));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const x = 1 + (items.length > 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 16, EndLine: 1, EndColumn: 37}}},
		{Code: "const box = {length: -1}; if (box.length) {} box.length = 2; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; if (box.length) {} box.length = 2; if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 66, EndLine: 1, EndColumn: 76}}},
		{Code: "const box = {length: -1, size: 2}; if (box.length) {} if (box.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1, size: 2}; if (box.length) {} if (box.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 59, EndLine: 1, EndColumn: 67}}},

		// Disabling built-in globals must also disable their static values.
		{Code: "const foo = {length: NaN}; if (foo.length) {}", FileName: "case.js", Globals: map[string]any{"NaN": "off"}, Output: []string{"const foo = {length: NaN}; if (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 32, EndLine: 1, EndColumn: 42}}},
		{Code: "const foo = {length: Infinity}; if (foo.length) {}", FileName: "case.js", Globals: map[string]any{"Infinity": "off"}, Output: []string{"const foo = {length: Infinity}; if (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 37, EndLine: 1, EndColumn: 47}}},

		{Code: "if (box.width || box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (box.width || box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28}}},
		{Code: "if (box?.width && box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (box?.width && box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
		{Code: "if (box[\"width\"] && box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (box[\"width\"] && box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}}},
		{Code: "if (box.length && (box.length !== 0 || ready)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (box.length > 0 && (box.length > 0 || ready)) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 15}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 36}}},
		{Code: "if (box.length && Boolean?.(box.length > 0)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (box.length > 0 && Boolean?.(box.length > 0)) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 15}}},
		{Code: "function f(Boolean) { return box.length && Boolean(box.length > 0); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 30, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(Boolean) { return box.length > 0 && Boolean(box.length > 0); }"}}}}},
		{Code: "if (getBox().width && getBox().length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (getBox().width && getBox().length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 23, EndLine: 1, EndColumn: 38}}},
		{Code: "if (getBox().length && getBox().length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (getBox().length > 0 && getBox().length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20}}},
		{Code: "\"😀\";\nif (\n  !items /* keep */.length\n) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"\"😀\";\nif (\n  items /* keep */.length === 0\n) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 3, Column: 3, EndLine: 3, EndColumn: 27}}},
		{Code: "if (!/* removed upstream */(items.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.length === 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 42}}},
		{Code: "if ((items.length) >= (0x1)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 28}}},
		{Code: "if (items.length == 0x0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.length === 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24}}},
		{Code: "if (1e0 <= items.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 22}}},
		{Code: "if ((items.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{}}, Output: []string{"if ((items.length > 0)) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 6, EndLine: 1, EndColumn: 18}}},
		{Code: "if (items.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "greater-than"}}, Output: []string{"if (items.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 17}}},
		{Code: "const n = items.size && ready;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size !== 0` when checking size is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const n = items.size !== 0 && ready;"}}}}},
		{Code: "const n = !Boolean(!Boolean(items.size));", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const n = items.size > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 41}}},
		{Code: "const n = void (!items.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const n = void (items.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 30}}},
		{Code: "async function f() { return await (!items.length); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"async function f() { return await (items.length === 0); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 36, EndLine: 1, EndColumn: 49}}},
		{Code: "if ((!items.length) === true) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 6, EndLine: 1, EndColumn: 19}}},
		{Code: "const n = !items.length + 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 11, EndLine: 1, EndColumn: 24}}},
		{Code: "const n = 1 + !items.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const n = 1 + (items.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28}}},
		{Code: "if ((items as string[]).length) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if ((items as string[]).length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 31}}},
		{Code: "if (items!.length) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items!.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 18}}},
		{Code: "if (/** @type {number} */ (items.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (/** @type {number} */ (items.length > 0)) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 28, EndLine: 1, EndColumn: 40}}},
		{Code: "if (/** @type {unknown[]} */ (items).length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (/** @type {unknown[]} */ (items).length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 30, EndLine: 1, EndColumn: 44}}},
		{Code: "if (Boolean(/** @type {number} */ (items.length))) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 50}}},
		{Tsx: true, Code: "const view = <div>{items.length && <span />}</div>;", FileName: "case.tsx", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const view = <div>{items.length > 0 && <span />}</div>;"}}}}},
		{Code: "const box = {length: -1, other: unknown}; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1, other: unknown}; if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 47, EndLine: 1, EndColumn: 57}}},
		{Code: "const box = {get length() { return -1; }}; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {get length() { return -1; }}; if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 48, EndLine: 1, EndColumn: 58}}},
		{Code: "const box = {length: -1}; consume(box); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; consume(box); if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 45, EndLine: 1, EndColumn: 55}}},
		{Code: "const box = {length: -1}; Object.defineProperties(box, {length: {value: 2}}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; Object.defineProperties(box, {length: {value: 2}}); if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 83, EndLine: 1, EndColumn: 93}}},
		{Code: "const box = {length: -1}; const n = 2; box.length = n; if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; const n = 2; box.length = n; if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 60, EndLine: 1, EndColumn: 70}}},
		{Code: "const box = {length: -1}; const values = [1, 2]; for (box.length of values) {} if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; const values = [1, 2]; for (box.length of values) {} if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 84, EndLine: 1, EndColumn: 94}}},
		{Code: "const box = {length: -1}; ({a: [box.length]} = {a: [2]}); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; ({a: [box.length]} = {a: [2]}); if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 63, EndLine: 1, EndColumn: 73}}},
		{Code: "const box = {length: -1}; do {box.length = 2;} while (false); if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; do {box.length = 2;} while (false); if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 67, EndLine: 1, EndColumn: 77}}},
		{Code: "const box = {length: -1}; try {} finally {box.length = 2;} if (box.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const box = {length: -1}; try {} finally {box.length = 2;} if (box.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74}}},
	})
}

func TestExplicitLengthCheckArtifactsFollowDemand(t *testing.T) {
	for _, testCase := range []struct {
		source, output string
		options        []any
		suggestions    []struct{ messageID, description, output string }
	}{
		{source: "if (items.length) {}", output: "if (items.length > 0) {}"},
		{source: "const x = Boolean(items.length).valueOf();", output: "const x = (items.length > 0).valueOf();"},
		{source: "foo()\n![].length", output: "foo()\n;[].length === 0"},
		{source: "const x = !items.size;", output: "const x = items.size === 0;"},
		{source: "if (items.size) {}", output: "if (items.size !== 0) {}", options: []any{map[string]any{"non-zero": "not-equal"}}},
		{source: "const x = items.length && render();", output: "const x = items.length && render();", suggestions: []struct{ messageID, description, output string }{{"suggestion", "Replace `.length` with `.length > 0`.", "const x = items.length > 0 && render();"}}},
		{source: "const x = items.size && render();", output: "const x = items.size && render();", options: []any{map[string]any{"non-zero": "not-equal"}}, suggestions: []struct{ messageID, description, output string }{{"suggestion", "Replace `.size` with `.size !== 0`.", "const x = items.size !== 0 && render();"}}},
		{source: "if (!items.length > 0) {}", output: "if (!items.length > 0) {}"},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, sourceFile, err := helper.CreateTestProgram(testCase.source, "edit-demand.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				var got []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: explicit_length_check.ExplicitLengthCheckRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								return explicit_length_check.ExplicitLengthCheckRule.Run(ctx, testCase.options)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
				})
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				diagnostics[demand] = got[0]
			}
			baseline := diagnostics[rule.EditDemandNone]
			all := diagnostics[rule.EditDemandAll]
			for demand, diagnostic := range diagnostics {
				if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.Severity != baseline.Severity {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := testCase.output != testCase.source && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Errorf("demand %d: unexpected autofix artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofix artifacts", demand)
				}
				expectedOutput := testCase.source
				if wantFix {
					expectedOutput = testCase.output
				}
				output, _, fixed := linter.ApplyRuleFixes(testCase.source, []rule.RuleDiagnostic{diagnostic})
				if output != expectedOutput || fixed != wantFix {
					t.Errorf("demand %d: unexpected autofix %q", demand, output)
				}
				wantSuggestions := len(testCase.suggestions) > 0 && (demand == rule.EditDemandSuggestion || demand == rule.EditDemandAll)
				if !wantSuggestions {
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced suggestions without demand", demand)
					}
					continue
				}
				if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != len(testCase.suggestions) || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Fatalf("demand %d: inconsistent suggestions", demand)
				}
				for index, expected := range testCase.suggestions {
					suggestion := (*diagnostic.Suggestions)[index]
					if suggestion.Message.Id != expected.messageID || suggestion.Message.Description != expected.description {
						t.Errorf("demand %d: wrong suggestion message", demand)
					}
					output, _, fixed := linter.ApplyRuleFixes(testCase.source, (*diagnostic.Suggestions)[index:index+1])
					if !fixed || output != expected.output {
						t.Errorf("demand %d: unexpected suggestion %q", demand, output)
					}
				}
			}
		})
	}
}
