package prefer_date_now_test

import (
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_date_now"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	dateMessageID   = "prefer-date"
	methodMessageID = "prefer-date-now-over-methods"
	numberMessageID = "prefer-date-now-over-number-data-object"
)

func invalidCase(code, output, messageID string, targets ...string) rule_tester.InvalidTestCase {
	result := rule_tester.InvalidTestCase{Code: code, FileName: "file.js", Output: []string{output}}
	searchFrom := 0
	for _, target := range targets {
		offset := strings.Index(code[searchFrom:], target)
		if offset < 0 || target == "" {
			panic("missing diagnostic target: " + target)
		}
		start := searchFrom + offset
		end := start + len(target)
		searchFrom = end
		position := func(offset int) (int, int) {
			prefix := code[:offset]
			lineStart := strings.LastIndex(prefix, "\n") + 1
			return strings.Count(prefix, "\n") + 1, len(utf16.Encode([]rune(prefix[lineStart:]))) + 1
		}
		line, column := position(start)
		endLine, endColumn := position(end)
		message := "Prefer `Date.now()` over `new Date()`."
		switch messageID {
		case methodMessageID:
			message = "Prefer `Date.now()` over `Date#" + target + "()`."
		case numberMessageID:
			message = "Prefer `Date.now()` over `Number(new Date())`."
		}
		result.Errors = append(result.Errors, rule_tester.InvalidTestCaseError{
			MessageId: messageID, Message: message,
			Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
		})
	}
	if len(result.Errors) == 0 {
		panic("invalid case requires diagnostic targets")
	}
	return result
}

// Every case from eslint-plugin-unicorn v74.0.0/test/prefer-date-now.js.
func TestPreferDateNowUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_date_now.PreferDateNowRule,
		[]rule_tester.ValidTestCase{
			{Code: `const ts = Date.now()`},
			// Constructor shape and arguments.
			{Code: `+Date()`},
			{Code: `+ Date`},
			{Code: `+ new window.Date()`},
			{Code: `+ new Moments()`},
			{Code: `+ new Date(0)`},
			{Code: `+ new Date(...[])`},
			// Method calls.
			{Code: `new Date.getTime()`},
			{Code: `valueOf()`},
			{Code: `new Date()[getTime]()`},
			{Code: `new Date()["valueOf"]()`},
			{Code: `new Date().notListed(0)`},
			{Code: `new Date().getTime(0)`},
			{Code: `new Date().valueOf(...[])`},
			// Number and BigInt calls.
			{Code: `new Number(new Date())`},
			{Code: `window.BigInt(new Date())`},
			{Code: `toNumber(new Date())`},
			{Code: `BigInt()`},
			{Code: `Number(new Date(), extraArgument)`},
			{Code: `BigInt([...new Date()])`},
			// Unary, assignment and binary expressions.
			{Code: `throw new Date()`},
			{Code: `typeof new Date()`},
			{Code: `const foo = () => {return new Date()}`},
			{Code: `foo += new Date()`},
			{Code: `function * foo() {yield new Date()}`},
			{Code: `new Date() + new Date()`},
			{Code: `foo = new Date() | 0`},
			{Code: `foo &= new Date()`},
			{Code: `foo = new Date() >> 0`},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase(`const ts = new Date().getTime();`, `const ts = Date.now();`, methodMessageID, "getTime"),
			invalidCase(`const ts = (new Date).getTime();`, `const ts = Date.now();`, methodMessageID, "getTime"),
			invalidCase(`const ts = (new Date()).getTime();`, `const ts = Date.now();`, methodMessageID, "getTime"),
			invalidCase(`const ts = new Date().valueOf();`, `const ts = Date.now();`, methodMessageID, "valueOf"),
			invalidCase(`const ts = (new Date).valueOf();`, `const ts = Date.now();`, methodMessageID, "valueOf"),
			invalidCase(`const ts = (new Date()).valueOf();`, `const ts = Date.now();`, methodMessageID, "valueOf"),
			invalidCase(`const ts = /* 1 */ Number(/* 2 */ new /* 3 */ Date( /* 4 */ ) /* 5 */) /* 6 */`, `const ts = /* 1 */ Date.now() /* 6 */`, numberMessageID, "Number(/* 2 */ new /* 3 */ Date( /* 4 */ ) /* 5 */)"),
			invalidCase(`const tsBigInt = /* 1 */ BigInt(/* 2 */ new /* 3 */ Date( /* 4 */ ) /* 5 */) /* 6 */`, `const tsBigInt = /* 1 */ BigInt(/* 2 */ Date.now() /* 5 */) /* 6 */`, dateMessageID, "new /* 3 */ Date( /* 4 */ )"),
			invalidCase(`const ts = + /* 1 */ new Date;`, `const ts = Date.now();`, dateMessageID, "+ /* 1 */ new Date"),
			invalidCase(`const ts = - /* 1 */ new Date();`, `const ts = - /* 1 */ Date.now();`, dateMessageID, "new Date()"),
			invalidCase(`const ts = +(new Date());`, `const ts = Date.now();`, dateMessageID, "+(new Date())"),
			invalidCase(`const ts = -(new Date());`, `const ts = -(Date.now());`, dateMessageID, "new Date()"),
			invalidCase(`const ts = new Date() - 0`, `const ts = Date.now() - 0`, dateMessageID, "new Date()"),
			invalidCase(`const foo = bar - new Date`, `const foo = bar - Date.now()`, dateMessageID, "new Date"),
			invalidCase(`const foo = new Date() * bar`, `const foo = Date.now() * bar`, dateMessageID, "new Date()"),
			invalidCase(`const ts = new Date() / 1`, `const ts = Date.now() / 1`, dateMessageID, "new Date()"),
			invalidCase(`const ts = new Date() % Infinity`, `const ts = Date.now() % Infinity`, dateMessageID, "new Date()"),
			invalidCase(`const ts = new Date() ** 1`, `const ts = Date.now() ** 1`, dateMessageID, "new Date()"),
			invalidCase(`const zero = (new Date(/* 1 */) /* 2 */) /* 3 */ - /* 4 */new Date`, `const zero = (Date.now() /* 2 */) /* 3 */ - /* 4 */Date.now()`, dateMessageID, "new Date(/* 1 */)", "new Date"),
			invalidCase(`foo -= new Date()`, `foo -= Date.now()`, dateMessageID, "new Date()"),
			invalidCase(`foo *= new Date()`, `foo *= Date.now()`, dateMessageID, "new Date()"),
			invalidCase(`foo /= new Date`, `foo /= Date.now()`, dateMessageID, "new Date"),
			invalidCase(`foo %= new Date()`, `foo %= Date.now()`, dateMessageID, "new Date()"),
			invalidCase(`foo **= new Date()`, `foo **= Date.now()`, dateMessageID, "new Date()"),
			invalidCase(`function foo(){return+new Date}`, `function foo(){return Date.now()}`, dateMessageID, "+new Date"),
			invalidCase(`function foo(){return-new Date}`, `function foo(){return-Date.now()}`, dateMessageID, "new Date"),
		})
}

func TestPreferDateNowUpstreamDocumentation(t *testing.T) {
	// eslint-plugin-unicorn v74.0.0/docs/rules/prefer-date-now.md.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_date_now.PreferDateNowRule,
		[]rule_tester.ValidTestCase{
			{Code: `const foo = Date.now();`},
			{Code: `const foo = Date.now() * 2;`},
		}, []rule_tester.InvalidTestCase{
			invalidCase(`const foo = new Date().getTime();`, `const foo = Date.now();`, methodMessageID, "getTime"),
			invalidCase(`const foo = new Date().valueOf();`, `const foo = Date.now();`, methodMessageID, "valueOf"),
			invalidCase(`const foo = +new Date;`, `const foo = Date.now();`, dateMessageID, "+new Date"),
			invalidCase(`const foo = Number(new Date());`, `const foo = Date.now();`, numberMessageID, "Number(new Date())"),
			invalidCase(`const foo = new Date() * 2;`, `const foo = Date.now() * 2;`, dateMessageID, "new Date()"),
		})
}
