package prefer_reflect_apply_test

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_reflect_apply"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "prefer-reflect-apply"
const messageText = "Prefer `Reflect.apply()` over `Function#apply()`."

func applyInvalid(code, reported string, output ...string) rule_tester.InvalidTestCase {
	start := strings.Index(code, reported)
	if start < 0 || reported == "" {
		panic("missing reported call in prefer-reflect-apply test")
	}
	position := func(offset int) (int, int) {
		prefix := code[:offset]
		line := strings.Count(prefix, "\n") + 1
		column := len(utf16.Encode([]rune(prefix[strings.LastIndex(prefix, "\n")+1:]))) + 1
		return line, column
	}
	line, column := position(start)
	endLine, endColumn := position(start + len(reported))
	return rule_tester.InvalidTestCase{
		Code: code, FileName: "file.js", Output: output,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID, Message: messageText,
			Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
		}},
	}
}

// Complete test/prefer-reflect-apply.js and documentation examples from
// https://github.com/sindresorhus/eslint-plugin-unicorn/tree/v75.0.0
func TestPreferReflectApplyUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `foo.apply();`},
		{Code: `foo.apply(null);`},
		{Code: `foo.apply(this);`},
		{Code: `foo.apply(null, 42);`},
		{Code: `foo.apply(this, 42);`},
		{Code: `foo.apply(bar, arguments);`},
		{Code: `[].apply(null, [42]);`},
		{Code: `foo.apply(bar);`},
		{Code: `foo.apply(bar, []);`},
		{Code: `foo.apply;`},
		{Code: `apply;`},
		{Code: `Reflect.apply(foo, null);`},
		{Code: `Reflect.apply(foo, null, [bar]);`},
		// Upstream does not pass a scope to getPropertyName.
		{Code: `const apply = "apply"; foo[apply](null, [42]);`},
	}
	invalid := []rule_tester.InvalidTestCase{}
	for _, test := range []struct{ code, output string }{
		{`foo.apply(null, [42]);`, `Reflect.apply(foo, null, [42]);`},
		{`foo.apply(null, []);`, `Reflect.apply(foo, null, []);`},
		{`(foo.bar).apply(null, [42]);`, `Reflect.apply(foo.bar, null, [42]);`},
		{`foo.bar.apply(null, [42]);`, `Reflect.apply(foo.bar, null, [42]);`},
		{`Function.prototype.apply.call(foo, null, [42]);`, `Reflect.apply(foo, null, [42]);`},
		{`Function.prototype.apply.call(foo.bar, null, [42]);`, `Reflect.apply(foo.bar, null, [42]);`},
		{`foo.apply(null, arguments);`, `Reflect.apply(foo, null, arguments);`},
		{`Function.prototype.apply.call(foo, null, arguments);`, `Reflect.apply(foo, null, arguments);`},
		{`foo.apply(this, [42]);`, `Reflect.apply(foo, this, [42]);`},
		{`Function.prototype.apply.call(foo, this, [42]);`, `Reflect.apply(foo, this, [42]);`},
		{`foo.apply(this, arguments);`, `Reflect.apply(foo, this, arguments);`},
		{`Function.prototype.apply.call(foo, this, arguments);`, `Reflect.apply(foo, this, arguments);`},
		{`foo["apply"](null, [42]);`, `Reflect.apply(foo, null, [42]);`},
	} {
		invalid = append(invalid, applyInvalid(test.code, strings.TrimSuffix(test.code, ";"), test.output))
	}

	// Each documentation example declares foo before the direct/prototype
	// calls. Keep those bindings as well as all four receiver/list pairings.
	for _, receiver := range []string{"null", "this"} {
		for _, arguments := range []string{"[42]", "arguments"} {
			prefix := "function foo() {}\n"
			output := prefix + "Reflect.apply(foo, " + receiver + ", " + arguments + ");"
			valid = append(valid, rule_tester.ValidTestCase{Code: output})
			for _, call := range []string{
				"foo.apply(" + receiver + ", " + arguments + ")",
				"Function.prototype.apply.call(foo, " + receiver + ", " + arguments + ")",
			} {
				invalid = append(invalid, applyInvalid(prefix+call+";", call, output))
			}
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_reflect_apply.PreferReflectApplyRule, valid, invalid)
}
