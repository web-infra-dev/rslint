// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/explicit-length-check.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/explicit-length-check.md
package explicit_length_check_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/explicit_length_check"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestExplicitLengthCheckUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_length_check.ExplicitLengthCheckRule, []rule_tester.ValidTestCase{
		// Upstream valid #1
		{Code: "if (foo.notLength) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #2
		{Code: "if (length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #3
		{Code: "if (foo[length]) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #4
		{Code: "if (foo[\"length\"]) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #5
		{Code: "foo.length === 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #6
		{Code: "foo.length > 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #7
		{Code: "const bar = foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #8
		{Code: "const bar = +foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #9
		{Code: "const x = Boolean(foo.length, foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #10
		{Code: "const x = new Boolean(foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #11
		{Code: "const x = NotBoolean(foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #12
		{Code: "const Boolean = value => value; const isNotEmpty = Boolean(foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #13
		{Code: "function unicorn(Boolean) { if (Boolean(foo.length)) {} }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #14
		{Code: "const length = foo.length ?? 0", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #15
		{Code: "if (foo.length ?? bar) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #16
		{Code: "if (foo.length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #17
		{Code: "if (foo.length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "greater-than"}}},
		// Upstream valid #18
		{Code: "if (foo.length !== 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}},
		// Upstream valid #19
		{Code: "const object: {size: number | null} = {size: 123}; if (object.size && object.size > 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #20
		{Code: "if (foo.length!) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #21
		{Code: "const object: {length: number | undefined} = {length: 123}; if (object.length && object.length > 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #22
		{Code: "const object: {size: number | undefined} = {size: 123}; if (object.size && object.size !== 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}},
		// Upstream valid #23
		{Code: "const object: {size: number | undefined} = {size: 123}; if (object.size && object.size! > 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #24
		{Code: "const object: {size: number | undefined} = {size: 123}; if (object.size && (object.size as number) > 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #25
		{Code: "const object: {size: number | undefined} = {size: 123}; if (object.size && (<number>object.size) > 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #26
		{Code: "const object = {size: 123}; if (object.size && (object.size satisfies number) > 0) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #27
		{Code: "const object: {size: number | undefined} = {size: 123}; if (object.size! >= 1) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #28
		{Code: "const object: {size: number | undefined} = {size: 123}; if ((object.size as number) >= 1) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #29
		{Code: "const object: {size: number | undefined} = {size: 123}; if ((<number>object.size) >= 1) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #30
		{Code: "const object = {size: 123}; if ((object.size satisfies number) >= 1) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #31
		{Code: "const object: {size: number | undefined} = {size: 123}; if (object.size && object.size! >= 1) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #32
		{Code: "if (foo.length === 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #33
		{Code: "const bar = foo.length === 0 ? 1 : 2", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #34
		{Code: "while (foo.length > 0) {\n\tfoo.pop();\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #35
		{Code: "do {\n\tfoo.pop();\n} while (foo.length > 0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #36
		{Code: "for (; foo.length > 0; foo.pop());", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #37
		{Code: "if (foo.length !== 1) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #38
		{Code: "if (foo.length > 1) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #39
		{Code: "if (foo.length < 2) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #40
		{Code: "const foo = { size: \"small\" }; if (foo.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #41
		{Code: "const foo = { length: -1 }; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #42
		{Code: "const foo = { length: 1.5 }; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #43
		{Code: "const foo = { length: NaN }; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #44
		{Code: "const foo = { length: Infinity }; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #45
		{Code: "const x = foo.length || 2", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #46
		{Code: "const A_NUMBER = 2; const x = foo.length || A_NUMBER", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #47
		{Code: "const x = foo.length || \"bar\"", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #48
		{Code: "const x = foo.length || `bar`", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #49
		{Code: "const A_STRING = \"bar\"; const x = foo.length || A_STRING", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #50
		{Code: "const size = props.size || \"mini\"", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #51
		{Code: "const x = foo.length || unknown", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #52
		{Code: "something(options.length || 500)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #53
		{Code: "const itemCount = result.totalCount || result.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #54
		{Code: "if (\n\tdimensions.width &&\n\tdimensions.height &&\n\tdimensions.length\n) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #55
		{Code: "if (packagingData.dimensions.width && packagingData.dimensions.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #56
		{Code: "if (dimensions.width && dimensions.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #57
		{Code: "if (dimensions.height && dimensions.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Upstream valid #58
		{Code: "if (dimensions.depth && dimensions.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Upstream invalid #1
		{Code: "if (!foo.length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
		// Upstream invalid #2
		{Code: "if (!foo.length === 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
		// Upstream invalid #3
		{Code: "() => foo.length && bar()", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "() => foo.length > 0 && bar()"}}}}},
		// Upstream invalid #4
		{Code: "alert(foo.length && bar())", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 7, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "alert(foo.length > 0 && bar())"}}}}},
		// Upstream invalid #5
		{Code: "if (items.length && items.every(Boolean)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.length > 0 && items.every(Boolean)) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 17}}},
		// Upstream invalid #6
		{Code: "if (items.map && items.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (items.map && items.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 18, EndLine: 1, EndColumn: 30}}},
		// Upstream invalid #7
		{Code: "if (text.trim && text.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (text.trim && text.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 18, EndLine: 1, EndColumn: 29}}},
		// Upstream invalid #8
		{Code: "if (set.has && set.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (set.has && set.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24}}},
		// Upstream invalid #9
		{Code: "if (bytes.subarray && bytes.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (bytes.subarray && bytes.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 23, EndLine: 1, EndColumn: 35}}},
		// Upstream invalid #10
		{Code: "if (typedArray.BYTES_PER_ELEMENT && typedArray.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (typedArray.BYTES_PER_ELEMENT && typedArray.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 37, EndLine: 1, EndColumn: 54}}},
		// Upstream invalid #11
		{Code: "if (container.width && items.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (container.width && items.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 24, EndLine: 1, EndColumn: 36}}},
		// Upstream invalid #12
		{Code: "if (dimensions.width > 0 && dimensions.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (dimensions.width > 0 && dimensions.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 29, EndLine: 1, EndColumn: 46}}},
		// Upstream invalid #13
		{Code: "if (object.size && object.size >= 1) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 36}}},
		// Upstream invalid #14
		{Code: "if (0 < object.size && object.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size > 0 && object.size) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20}}},
		// Upstream invalid #15
		{Code: "if (object.size && object.size > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}, Output: []string{"if (object.size && object.size !== 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size !== 0` when checking size is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 35}}},
		// Upstream invalid #16
		{Code: "if (object.size && other.size > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size > 0 && other.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
		// Upstream invalid #17
		{Code: "if (object.size && !(object.size === 0)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 40}}},
		// Upstream invalid #18
		{Code: "if (object.size && Boolean(object.size !== 0)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46}}},
		// Upstream invalid #19
		{Code: "if (object.size && Boolean(object.size > 0)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 44}}},
		// Upstream invalid #20
		{Code: "if (object.size && !Boolean(object.size === 0)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 47}}},
		// Upstream invalid #21
		{Code: "if (object.size >= 1 && object.size !== 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size > 0 && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21}, {MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 25, EndLine: 1, EndColumn: 42}}},
		// Upstream invalid #22
		{Code: "if (!!object.size && object.size > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size > 0 && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 18}}},
		// Upstream invalid #23
		{Code: "if (!object.size && object.size > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (object.size === 0 && object.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.size === 0` when checking size is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 17}}},
	})
}

func TestExplicitLengthCheckSnapshots(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_length_check.ExplicitLengthCheckRule, []rule_tester.ValidTestCase{
		// Snapshots valid #1
		{Code: "class A {\n\ta() {\n\t\tif (this.length);\n\t\twhile (!this.size || foo);\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #2
		{Code: "const foo = {length: 123}; mutate(); if (foo.length) {} function mutate() { Object.defineProperty(foo, 'length', {get() { return 'x'; }}); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #3
		{Code: "const foo = {length: -1}; function mutate() { foo.length = 123; } if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #4
		{Code: "const foo = {length: 123}; if (foo.length) {} mutate(); function mutate() { Object.defineProperty(foo, 'length', {get() { return 'x'; }}); }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #5
		{Code: "const foo = {length: -1}; foo.length = 'x'; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #6
		{Code: "const foo = {length: 123}; foo.length = 'x'; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #7
		{Code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #8
		{Code: "const foo = {length: 123}; Object.defineProperty(foo, 'length', {value: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #9
		{Code: "const foo = {length: -1}; const descriptor = {value: 'x'}; Object.defineProperty(foo, 'length', descriptor); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #10
		{Code: "const foo = {length: -1}; Object.assign(foo, {length: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #11
		{Code: "const foo = {length: -1}; const values = {length: 'x'}; Object.assign(foo, values); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #12
		{Code: "const foo = {length: -1}; const values = {length: 'x'}; Object.assign(foo, {length: 123}, values); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #13
		{Code: "const foo = {length: -1}; Object.assign(foo, {length: 123, length: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #14
		{Code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123, value: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #15
		{Code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123, ...descriptor}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #16
		{Code: "const foo = {length: -1}; Object.defineProperties(foo, {length: {value: 123}, length: {value: 'x'}}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #17
		{Code: "const foo = {length: -1}; ({length: foo.length} = {length: 123, length: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #18
		{Code: "const foo = {length: -1}; ({length: foo.length = 123} = {length: 456}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #19
		{Code: "const foo = {length: -1}; const {value = (foo.length = 123)} = {value: 0}; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #20
		{Code: "const foo = {length: -1}; try { if (condition) throw new Error(); foo.length = 123; } catch {} if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #21
		{Code: "const foo = {length: -1}; if (condition) foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #22
		{Code: "const foo = {length: -1}; condition && (foo.length = 123); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #23
		{Code: "const foo = {length: -1}; condition ? foo.length = 123 : 0; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #24
		{Code: "const foo = {length: -1}; while (condition) foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #25
		{Code: "const foo = {length: -1}; for (; condition;) foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #26
		{Code: "const foo = {length: -1}; switch (value) { case 1: foo.length = 123; } if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #27
		{Code: "const foo = {length: -1}; try {} catch { foo.length = 123; } if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #28
		{Code: "const foo = {length: -1}; Object.assign?.(foo, {length: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #29
		{Code: "const foo = {length: -1}; object?.method(foo.length = 123); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #30
		{Code: "const foo = {length: -1}; Object.defineProperties(foo, definitions); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #31
		{Code: "const foo = {length: -1}; Object.defineProperties(foo, {...definitions}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #32
		{Code: "const foo = {length: -1}; Object.defineProperty(foo, propertyName, {value: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #33
		{Code: "const foo = {length: -1}; Object.assign(foo, {other: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #34
		{Code: "const foo = {length: -1}; maybe?.[foo.length = 123]; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #35
		{Code: "const foo = {length: -1}; maybe?.property[foo.length = 123]; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #36
		{Code: "const foo = {length: -1}; maybe?.property.method(foo.length = 123); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #37
		{Code: "const foo = {length: 123}; Object.assign?.(foo, {length: 'x'}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #38
		{Code: "const foo = {length: -1}; if (foo.length) {} foo.length = 123;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #39
		{Code: "const foo = {length: -1}; if (foo.length) {} Object.assign(foo, {length: 123});", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #40
		{Code: "const foo = {length: -1}; if (foo.length) {} Object.defineProperty(foo, 'length', {value: 123});", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #41
		{Code: "const foo = {length: -1}; class A {field = (foo.length = \"x\");} new A(); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #42
		{Code: "const foo = {length: -1}; class A {[foo.length = \"x\"]() {}} if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #43
		{Code: "const foo = {length: -1}; class A {accessor field = (foo.length = \"x\");} new A(); if (foo.length) {}", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #44
		{Code: "const foo = {length: -1}; [...foo.length] = [[123]]; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #45
		{Code: "const foo = {length: -1}; for (foo.length in {key: true}) {} if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #46
		{Code: "const foo = {length: 123}; for (foo.length in {key: true}) {} if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots valid #47
		{Code: "const foo = {length: -1}; for (foo.length of ['x']) {} if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},

		// Keep upstream mutation/escape cases; rslint leaves these local objects unchecked.
		// Snapshots invalid #4 (conservative local-object handling)
		{Code: "const foo = {length: -1}; foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #5 (conservative local-object handling)
		{Code: "const foo = {length: -1}; Object.assign(foo, {length: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #6 (conservative local-object handling)
		{Code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #7 (conservative local-object handling)
		{Code: "const foo = {length: -1}; [foo.length] = [123]; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #8 (conservative local-object handling)
		{Code: "const foo = {length: -1}; ({length: foo.length} = {length: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #9 (conservative local-object handling)
		{Code: "const foo = {length: -1}; for (foo.length of [123]) {} if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #49 (conservative local-object handling)
		{Code: "const foo = {length: -1}; if (true) foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #50 (conservative local-object handling)
		{Code: "const foo = {length: -1}; if (false) {} else foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #51 (conservative local-object handling)
		{Code: "const foo = {length: -1}; true ? foo.length = 123 : 0; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #52 (conservative local-object handling)
		{Code: "const foo = {length: -1}; false ? 0 : foo.length = 123; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #53 (conservative local-object handling)
		{Code: "const foo = {length: -1}; Object.assign(foo, {length: 'x', length: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #54 (conservative local-object handling)
		{Code: "const foo = {length: -1}; Object.defineProperty(foo, 'length', {value: 'x', value: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #55 (conservative local-object handling)
		{Code: "const foo = {length: -1}; Object.defineProperties(foo, {length: {value: 'x'}, length: {value: 123}}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #56 (conservative local-object handling)
		{Code: "const foo = {length: -1}; ({length: foo.length} = {length: 'x', length: 123}); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #57 (conservative local-object handling)
		{Code: "const foo = {length: 123}; Object.assign(foo); if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Snapshots invalid #58 (conservative local-object handling)
		{Code: "const foo = {length: -1}; switch (value) { default: foo.length = 123; } if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Snapshots invalid #1
		{Code: "if (\n\t!!!(\n\t\t!foo.length &&\n\t\tfoo.length == 0 &&\n\t\tfoo.length < 1 &&\n\t\tfoo.length <= 0 &&\n\t\t0 === foo.length &&\n\t\t0 == foo.length &&\n\t\t1 > foo.length &&\n\t\t0 >= foo.length\n\t) ||\n\t!(\n\t\tfoo.length ||\n\t\t!!foo.length ||\n\t\tfoo.length !== 0 ||\n\t\tfoo.length != 0 ||\n\t\tfoo.length >= 1 ||\n\t\t0 !== foo.length ||\n\t\t0 != foo.length ||\n\t\t0 < foo.length ||\n\t\t1 <= foo.length\n\t)\n) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (\n\t!!!(\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0 &&\n\t\tfoo.length === 0\n\t) ||\n\t!(\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0 ||\n\t\tfoo.length > 0\n\t)\n) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 3, Column: 3, EndLine: 3, EndColumn: 14}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 4, Column: 3, EndLine: 4, EndColumn: 18}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 5, Column: 3, EndLine: 5, EndColumn: 17}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 6, Column: 3, EndLine: 6, EndColumn: 18}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 7, Column: 3, EndLine: 7, EndColumn: 19}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 8, Column: 3, EndLine: 8, EndColumn: 18}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 9, Column: 3, EndLine: 9, EndColumn: 17}, {MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 10, Column: 3, EndLine: 10, EndColumn: 18}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 13, Column: 3, EndLine: 13, EndColumn: 13}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 14, Column: 3, EndLine: 14, EndColumn: 15}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 15, Column: 3, EndLine: 15, EndColumn: 19}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 16, Column: 3, EndLine: 16, EndColumn: 18}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 17, Column: 3, EndLine: 17, EndColumn: 18}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 18, Column: 3, EndLine: 18, EndColumn: 19}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 19, Column: 3, EndLine: 19, EndColumn: 18}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 20, Column: 3, EndLine: 20, EndColumn: 17}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 21, Column: 3, EndLine: 21, EndColumn: 18}}},
		// Snapshots invalid #2
		{Code: "if (\n\tfoo.length ||\n\t!!foo.length ||\n\tfoo.length != 0 ||\n\tfoo.length > 0 ||\n\tfoo.length >= 1 ||\n\t0 !== foo.length ||\n\t0 != foo.length ||\n\t0 < foo.length ||\n\t1 <= foo.length\n) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}, Output: []string{"if (\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0 ||\n\tfoo.length !== 0\n) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 2, Column: 2, EndLine: 2, EndColumn: 12}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 3, Column: 2, EndLine: 3, EndColumn: 14}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 4, Column: 2, EndLine: 4, EndColumn: 17}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 5, Column: 2, EndLine: 5, EndColumn: 16}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 6, Column: 2, EndLine: 6, EndColumn: 17}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 7, Column: 2, EndLine: 7, EndColumn: 18}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 8, Column: 2, EndLine: 8, EndColumn: 17}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 9, Column: 2, EndLine: 9, EndColumn: 16}, {MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 10, Column: 2, EndLine: 10, EndColumn: 17}}},
		// Snapshots invalid #3
		{Code: "const foo = { length: 123 }; if (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}, Output: []string{"const foo = { length: 123 }; if (foo.length !== 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length !== 0` when checking length is not zero.", Line: 1, Column: 34, EndLine: 1, EndColumn: 44}}},
		// Snapshots invalid #10
		{Code: "if (foo.bar && foo.bar.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.bar && foo.bar.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 16, EndLine: 1, EndColumn: 30}}},
		// Snapshots invalid #11
		{Code: "if (foo.length || foo.bar()) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.length > 0 || foo.bar()) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 15}}},
		// Snapshots invalid #12
		{Code: "if (!!(!!foo.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 21}}},
		// Snapshots invalid #13
		{Code: "if (!(foo.length === 0)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 24}}},
		// Snapshots invalid #14
		{Code: "while (foo.length >= 1) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"while (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// Snapshots invalid #15
		{Code: "do {} while (foo.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"do {} while (foo.length > 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24}}},
		// Snapshots invalid #16
		{Code: "for (let i = 0; (bar && !foo.length); i ++) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for (let i = 0; (bar && foo.length === 0); i ++) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 25, EndLine: 1, EndColumn: 36}}},
		// Snapshots invalid #17
		{Code: "const isEmpty = foo.length < 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 31}}},
		// Snapshots invalid #18
		{Code: "const isEmpty = foo.length < 1 ? true : false;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isEmpty = foo.length === 0 ? true : false;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 31}}},
		// Snapshots invalid #19
		{Code: "const isEmpty = foo.length <= 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 32}}},
		// Snapshots invalid #20
		{Code: "if (0 >= foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.length === 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20}}},
		// Snapshots invalid #21
		{Code: "bar(foo.length >= 1)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"bar(foo.length > 0)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 20}}},
		// Snapshots invalid #22
		{Code: "bar(!foo.length || foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"bar(foo.length === 0 || foo.length)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
		// Snapshots invalid #23
		{Code: "const bar = void !foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const bar = void (foo.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 18, EndLine: 1, EndColumn: 29}}},
		// Snapshots invalid #24
		{Code: "const isNotEmpty = Boolean(foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isNotEmpty = foo.length > 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 39}}},
		// Snapshots invalid #25
		{Code: "if (!!Boolean(foo.length)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 26}}},
		// Snapshots invalid #26
		{Code: "const isNotEmpty = Boolean(foo.length || bar)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isNotEmpty = Boolean(foo.length > 0 || bar)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 28, EndLine: 1, EndColumn: 38}}},
		// Snapshots invalid #27
		{Code: "const isEmpty = Boolean(!foo.length)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isEmpty = foo.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 37}}},
		// Snapshots invalid #28
		{Code: "const isEmpty = Boolean(foo.length === 0)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isEmpty = foo.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 42}}},
		// Snapshots invalid #29
		{Code: "const isNotEmpty = !Boolean(foo.length === 0)", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isNotEmpty = foo.length > 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 20, EndLine: 1, EndColumn: 46}}},
		// Snapshots invalid #30
		{Code: "const isEmpty = !Boolean(!Boolean(foo.length === 0))", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"const isEmpty = foo.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 53}}},
		// Snapshots invalid #31
		{Code: "if (foo.size) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.size > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
		// Snapshots invalid #32
		{Code: "if (foo.size && bar.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if (foo.size > 0 && bar.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.size > 0` when checking size is not zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27}}},
		// Snapshots invalid #33
		{Code: "function foo() {return!foo.length}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function foo() {return foo.length === 0}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 23, EndLine: 1, EndColumn: 34}}},
		// Snapshots invalid #34
		{Code: "function foo() {throw!foo.length}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function foo() {throw foo.length === 0}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 22, EndLine: 1, EndColumn: 33}}},
		// Snapshots invalid #35
		{Code: "async function foo() {await!foo.length}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"async function foo() {await (foo.length === 0)}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 28, EndLine: 1, EndColumn: 39}}},
		// Snapshots invalid #36
		{Code: "function * foo() {yield!foo.length}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function * foo() {yield foo.length === 0}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 24, EndLine: 1, EndColumn: 35}}},
		// Snapshots invalid #37
		{Code: "function * foo() {yield*!foo.length}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"function * foo() {yield*foo.length === 0}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 25, EndLine: 1, EndColumn: 36}}},
		// Snapshots invalid #38
		{Code: "delete!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"delete (foo.length === 0)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 7, EndLine: 1, EndColumn: 18}}},
		// Snapshots invalid #39
		{Code: "typeof!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"typeof (foo.length === 0)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 7, EndLine: 1, EndColumn: 18}}},
		// Snapshots invalid #40
		{Code: "void!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"void (foo.length === 0)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
		// Snapshots invalid #41
		// Preserve the right operand grouping; upstream omits these parentheses.
		{Code: "a instanceof!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"a instanceof (foo.length === 0)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 13, EndLine: 1, EndColumn: 24}}},
		// Snapshots invalid #42
		// Preserve the right operand grouping; upstream omits these parentheses.
		{Code: "a in!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"a in (foo.length === 0)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},
		// Snapshots invalid #43
		{Code: "export default!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"export default foo.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26}}},
		// Snapshots invalid #44
		{Code: "if(true){}else!foo.length", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"if(true){}else foo.length === 0"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26}}},
		// Snapshots invalid #45
		{Code: "do!foo.length;while(true) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"do foo.length === 0;while(true) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 3, EndLine: 1, EndColumn: 14}}},
		// Snapshots invalid #46
		{Code: "switch(foo){case!foo.length:{}}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"switch(foo){case foo.length === 0:{}}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 17, EndLine: 1, EndColumn: 28}}},
		// Snapshots invalid #47
		{Code: "for(const a of!foo.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for(const a of foo.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26}}},
		// Snapshots invalid #48
		{Code: "for(const a in!foo.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"for(const a in foo.length === 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26}}},
	})
}

func TestExplicitLengthCheckVue(t *testing.T) {
	// Vue parser services and template AST nodes are not available in rslint.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_length_check.ExplicitLengthCheckRule, []rule_tester.ValidTestCase{
		// Vue valid #1
		{Code: "<not-template><div v-if=\"foo.length\"></div></not-template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue valid #2
		{Code: "<template><div v-not-if=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue valid #3
		{Code: "<template><div v-if=\"foo.notLength\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue valid #4
		{Code: "<template><div v-SHoW=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue valid #5
		{Code: "<template><div hidden=\"!foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue valid #6
		{Code: "<template><img :width=\"foo.length\"/></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
	}, []rule_tester.InvalidTestCase{
		// Vue invalid #1
		{Code: "<template><div v-if=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #2
		{Code: "<template>\n\t<div>\n\t\t<div v-if=\"foo\"></div>\n\t\t<div v-else-if=\"bar.length\"></div>\n\t</div>\n</template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #3
		{Code: "<template><div v-if=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #4
		{Code: "<template><div v-if=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "not-equal"}}, Skip: true},
		// Vue invalid #5
		{Code: "<template><div v-if=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Options: []any{map[string]any{"non-zero": "greater-than"}}, Skip: true},
		// Vue invalid #6
		{Code: "<template><div v-if=\"foo.length && bar\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #7
		{Code: "<script>if (foo.length) {}</script>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #8
		{Code: "<template><div v-show=\"foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #9
		{Code: "<template><div :hidden=\"foo.length >= 1\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #10
		{Code: "<template><div @click=\"foo.length >= 1\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #11
		{Code: "<template><div @click=\"method($event, foo.length >= 1)\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #12
		{Code: "<template><div v-bind:hidden=\"0 === foo.length\"></div></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #13
		{Code: "<template><input :disabled=\"Boolean(foo.length)\"></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
		// Vue invalid #14
		{Code: "<template><custom-component :custom-property=\"!foo.length\"></custom-component></template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
	})
}

func TestExplicitLengthCheckDocs(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_length_check.ExplicitLengthCheckRule, []rule_tester.ValidTestCase{
		// Docs valid #9
		{Code: "// ✅\nconst isEmpty = foo.length === 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #11
		{Code: "// ✅\nconst isEmptySet = foo.size === 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #22
		{Code: "// ✅\nconst isNotEmpty = foo.length > 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #24
		{Code: "// ✅\nif (foo.length > 0 || bar.length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #26
		{Code: "// ✅\nconst unicorn = foo.length > 0 ? 1 : 2;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #28
		{Code: "// ✅\nwhile (foo.length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #30
		{Code: "// ✅\ndo {} while (foo.length > 0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #32
		{Code: "// ✅\nfor (; foo.length > 0; ) {};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #34
		{Code: "if (bothNotEmpty(foo, bar)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #35
		{Code: "const bothNotEmpty = (a, b) => a.length > 0 && b.length > 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
		// Docs valid #36
		{Code: "if (bothNotEmpty(foo, bar)) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}},
	}, []rule_tester.InvalidTestCase{
		// Docs invalid #1
		{Code: "// ❌\nconst isEmpty = !foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 28}}},
		// Docs invalid #2
		{Code: "// ❌\nconst isEmpty = foo.length == 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 32}}},
		// Docs invalid #3
		{Code: "// ❌\nconst isEmpty = foo.length < 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 31}}},
		// Docs invalid #4
		{Code: "// ❌\nconst isEmpty = foo.length <= 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 32}}},
		// Docs invalid #5
		{Code: "// ❌\nconst isEmpty = 0 === foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 33}}},
		// Docs invalid #6
		{Code: "// ❌\nconst isEmpty = 0 == foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 32}}},
		// Docs invalid #7
		{Code: "// ❌\nconst isEmpty = 1 > foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 31}}},
		// Docs invalid #8
		{Code: "// ❌\n// Negative style is disallowed too\nconst isEmpty = !(foo.length > 0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\n// Negative style is disallowed too\nconst isEmpty = foo.length === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 3, Column: 17, EndLine: 3, EndColumn: 34}}},
		// Docs invalid #10
		{Code: "// ❌\nconst isEmptySet = !foo.size;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isEmptySet = foo.size === 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.size === 0` when checking size is zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 29}}},
		// Docs invalid #13
		{Code: "// ❌\nconst isNotEmpty = foo.length !== 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 36}}},
		// Docs invalid #14
		{Code: "// ❌\nconst isNotEmpty = foo.length != 0;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 35}}},
		// Docs invalid #15
		{Code: "// ❌\nconst isNotEmpty = foo.length >= 1;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 35}}},
		// Docs invalid #16
		{Code: "// ❌\nconst isNotEmpty = 0 !== foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 36}}},
		// Docs invalid #17
		{Code: "// ❌\nconst isNotEmpty = 0 != foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 35}}},
		// Docs invalid #18
		{Code: "// ❌\nconst isNotEmpty = 0 < foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 34}}},
		// Docs invalid #19
		{Code: "// ❌\nconst isNotEmpty = 1 <= foo.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 35}}},
		// Docs invalid #20
		{Code: "// ❌\nconst isNotEmpty = Boolean(foo.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 20, EndLine: 2, EndColumn: 39}}},
		// Docs invalid #21
		{Code: "// ❌\n// Negative style is disallowed too\nconst isNotEmpty = !(foo.length === 0);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\n// Negative style is disallowed too\nconst isNotEmpty = foo.length > 0;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 3, Column: 20, EndLine: 3, EndColumn: 39}}},
		// Docs invalid #23
		{Code: "// ❌\nif (foo.length || bar.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nif (foo.length > 0 || bar.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 5, EndLine: 2, EndColumn: 15}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 19, EndLine: 2, EndColumn: 29}}},
		// Docs invalid #25
		{Code: "// ❌\nconst unicorn = foo.length ? 1 : 2;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nconst unicorn = foo.length > 0 ? 1 : 2;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 17, EndLine: 2, EndColumn: 27}}},
		// Docs invalid #27
		{Code: "// ❌\nwhile (foo.length) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nwhile (foo.length > 0) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 8, EndLine: 2, EndColumn: 18}}},
		// Docs invalid #29
		{Code: "// ❌\ndo {} while (foo.length);", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\ndo {} while (foo.length > 0);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 14, EndLine: 2, EndColumn: 24}}},
		// Docs invalid #31
		{Code: "// ❌\nfor (; foo.length; ) {};", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{"// ❌\nfor (; foo.length > 0; ) {};"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 2, Column: 8, EndLine: 2, EndColumn: 18}}},
		// Docs invalid #33
		{Code: "const bothNotEmpty = (a, b) => a.length && b.length;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 32, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const bothNotEmpty = (a, b) => a.length > 0 && b.length;"}}}, {MessageId: "non-zero", Message: "Use `.length > 0` when checking length is not zero.", Line: 1, Column: 44, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const bothNotEmpty = (a, b) => a.length && b.length > 0;"}}}}},
		// Docs invalid #37
		{Code: "// ❌\nif (!foo.length > 0) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "zero", Message: "Use `.length === 0` when checking length is zero.", Line: 2, Column: 5, EndLine: 2, EndColumn: 16}}},
	})
}

func TestExplicitLengthCheckDocsVue(t *testing.T) {
	// Vue parser services and template AST nodes are not available in rslint.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &explicit_length_check.ExplicitLengthCheckRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// DocsVue invalid #12
		{Code: "<template>\n\t<!-- ❌ -->\n\t<div v-if=\"!foo.length\">Vue</div>\n\n\t<!-- ✅ -->\n\t<div v-if=\"foo.length === 0\">Vue</div>\n</template>", FileName: "case.vue", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Skip: true},
	})
}
