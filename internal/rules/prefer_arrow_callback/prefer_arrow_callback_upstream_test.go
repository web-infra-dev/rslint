// TestPreferArrowCallbackUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/prefer-arrow-callback.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in the prefer_arrow_callback_extras_test.go file.
// cspell:ignore amet
package prefer_arrow_callback

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const preferArrowCallbackMessageText = "Unexpected function expression."

func preferArrowErrorAt(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		Line:      line,
		Column:    column,
		MessageId: "preferArrowCallback",
		Message:   preferArrowCallbackMessageText,
	}
}

func preferArrowErrors(code string, count int) []rule_tester.InvalidTestCaseError {
	errors := make([]rule_tester.InvalidTestCaseError, 0, count)
	offset := 0
	for len(errors) < count {
		nextAsync := strings.Index(code[offset:], "async function")
		nextFunction := strings.Index(code[offset:], "function")
		if nextAsync >= 0 {
			nextAsync += offset
		}
		if nextFunction >= 0 {
			nextFunction += offset
		}
		pos := nextFunction
		if nextAsync >= 0 && (pos < 0 || nextAsync < pos) {
			pos = nextAsync
		}
		if pos < 0 {
			panic("missing function token in test case")
		}
		line, column := lineColumnForOffset(code, pos)
		errors = append(errors, preferArrowErrorAt(line, column))
		offset = pos + len("function")
	}
	return errors
}

func lineColumnForOffset(code string, offset int) (int, int) {
	line, column := 1, 1
	for i, r := range code {
		if i >= offset {
			break
		}
		if r == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

func invalidCase(code string, output string, options any, count int) rule_tester.InvalidTestCase {
	tc := rule_tester.InvalidTestCase{
		Code:    code,
		Options: options,
		Errors:  preferArrowErrors(code, count),
	}
	if output != "" {
		tc.Output = []string{output}
	}
	return tc
}

func TestPreferArrowCallbackUpstream(t *testing.T) {
	allowNamed := []any{map[string]any{"allowNamedFunctions": true}}
	disallowNamed := []any{map[string]any{"allowNamedFunctions": false}}
	disallowUnboundThis := []any{map[string]any{"allowUnboundThis": false}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferArrowCallbackRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo(a => a);"},
			{Code: "foo(function*() {});"},
			{Code: "foo(function() { this; });"},
			{Code: "foo(function bar() {});", Options: allowNamed},
			{Code: "foo(function() { (() => this); });"},
			{Code: "foo(function() { this; }.bind(obj));"},
			{Code: "foo(function() { this; }.call(this));"},
			{Code: "foo(a => { (function() {}); });"},
			{Code: "var foo = function foo() {};"},
			{Code: "(function foo() {})();"},
			{Code: "foo(function bar() { bar; });"},
			{Code: "foo(function bar() { arguments; });"},
			{Code: "foo(function bar() { arguments; }.bind(this));"},
			{Code: "foo(function bar() { new.target; });"},
			{Code: "foo(function bar() { new.target; }.bind(this));"},
			{Code: "foo(function bar() { this; }.bind(this, somethingElse));"},
			{Code: "foo((function() {}).bind.bar)"},
			{Code: "foo((function() { this.bar(); }).bind(obj).bind(this))"},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase("foo(function bar() {});", "foo(() => {});", nil, 1),
			invalidCase("foo(function() {});", "foo(() => {});", allowNamed, 1),
			invalidCase("foo(function bar() {});", "foo(() => {});", disallowNamed, 1),
			invalidCase("foo(function() {});", "foo(() => {});", nil, 1),
			invalidCase("foo(nativeCb || function() {});", "foo(nativeCb || (() => {}));", nil, 1),
			invalidCase("foo(bar ? function() {} : function() {});", "foo(bar ? () => {} : () => {});", nil, 2),
			invalidCase("foo(function() { (function() { this; }); });", "foo(() => { (function() { this; }); });", nil, 1),
			invalidCase("foo(function() { this; }.bind(this));", "foo(() => { this; });", nil, 1),
			invalidCase("foo(bar || function() { this; }.bind(this));", "foo(bar || (() => { this; }));", nil, 1),
			invalidCase("foo(function() { (() => this); }.bind(this));", "foo(() => { (() => this); });", nil, 1),
			invalidCase("foo(function bar(a) { a; });", "foo((a) => { a; });", nil, 1),
			invalidCase("foo(function(a) { a; });", "foo((a) => { a; });", nil, 1),
			invalidCase("foo(function(arguments) { arguments; });", "foo((arguments) => { arguments; });", nil, 1),
			invalidCase("foo(function() { this; });", "", disallowUnboundThis, 1),
			invalidCase("foo(function() { (() => this); });", "", disallowUnboundThis, 1),
			invalidCase("qux(function(foo, bar, baz) { return foo * 2; })", "qux((foo, bar, baz) => { return foo * 2; })", nil, 1),
			invalidCase("qux(function(foo, bar, baz) { return foo * bar; }.bind(this))", "qux((foo, bar, baz) => { return foo * bar; })", nil, 1),
			invalidCase("qux(function(foo, bar, baz) { return foo * this.qux; }.bind(this))", "qux((foo, bar, baz) => { return foo * this.qux; })", nil, 1),
			invalidCase("foo(function() {}.bind(this, somethingElse))", "foo((() => {}).bind(this, somethingElse))", nil, 1),
			invalidCase("qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });", "qux((foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) => { return foo + bar; });", nil, 1),
			invalidCase("qux(function(baz, baz) { })", "", nil, 1),
			invalidCase("qux(function( /* no params */ ) { })", "qux(( /* no params */ ) => { })", nil, 1),
			invalidCase("qux(function( /* a */ foo /* b */ , /* c */ bar /* d */ , /* e */ baz /* f */ ) { return foo; })", "qux(( /* a */ foo /* b */ , /* c */ bar /* d */ , /* e */ baz /* f */ ) => { return foo; })", nil, 1),
			invalidCase("qux(async function (foo = 1, bar = 2, baz = 3) { return baz; })", "qux(async (foo = 1, bar = 2, baz = 3) => { return baz; })", nil, 1),
			invalidCase("qux(async function (foo = 1, bar = 2, baz = 3) { return this; }.bind(this))", "qux(async (foo = 1, bar = 2, baz = 3) => { return this; })", nil, 1),
			invalidCase("foo(async function /*\n*/ () { return 1; });", "", nil, 1),
			invalidCase("foo(async function // c\n () { return 1; });", "", nil, 1),
			invalidCase("foo(async function\n () { return 1; });", "", nil, 1),
			invalidCase("foo(async function /* c */ () { return 1; });", "foo(async  /* c */ () => { return 1; });", nil, 1),
			invalidCase("foo((bar || function() {}).bind(this))", "", nil, 1),
			invalidCase("foo(function() {}.bind(this).bind(obj))", "foo((() => {}).bind(obj))", nil, 1),

			// ---- Optional chaining ----
			invalidCase("foo?.(function() {});", "foo?.(() => {});", nil, 1),
			invalidCase("foo?.(function() { return this; }.bind(this));", "foo?.(() => { return this; });", nil, 1),
			invalidCase("foo(function() { return this; }?.bind(this));", "foo(() => { return this; });", nil, 1),
			invalidCase("foo((function() { return this; }?.bind)(this));", "", nil, 1),

			// ---- https://github.com/eslint/eslint/issues/16718 ----
			invalidCase(`
            test(
                function ()
                { }
            );
            `, `
            test(
                () =>
                { }
            );
            `, nil, 1),
			invalidCase(`
            test(
                function (
                    ...args
                ) /* Lorem ipsum
                dolor sit amet. */ {
                    return args;
                }
            );
            `, `
            test(
                (
                    ...args
                ) => /* Lorem ipsum
                dolor sit amet. */ {
                    return args;
                }
            );
            `, nil, 1),
		},
	)
}

func TestPreferArrowCallbackUpstreamTypeScript(t *testing.T) {
	allowNamed := []any{map[string]any{"allowNamedFunctions": true}}
	disallowNamed := []any{map[string]any{"allowNamedFunctions": false}}
	disallowUnboundThis := []any{map[string]any{"allowUnboundThis": false}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferArrowCallbackRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo(a => a);"},
			{Code: "foo((a:string) => a);"},
			{Code: "foo(function*() {});"},
			{Code: "foo(function() { this; });"},
			{Code: "foo(function bar(a:string) {});", Options: allowNamed},
			{Code: "foo(function() { (() => this); });"},
			{Code: "foo(function() { this; }.bind(obj));"},
			{Code: "foo(function() { this; }.call(this));"},
			{Code: "foo(a => { (function() {}); });"},
			{Code: "var foo = function foo() {};"},
			{Code: "(function foo() {})();"},
			{Code: "foo(function bar() { bar; });"},
			{Code: "foo(function bar() { arguments; });"},
			{Code: "foo(function bar() { arguments; }.bind(this));"},
			{Code: "foo(function bar() { new.target; });"},
			{Code: "foo(function bar() { new.target; }.bind(this));"},
			{Code: "foo(function bar() { this; }.bind(this, somethingElse));"},
			{Code: "foo((function() {}).bind.bar)"},
			{Code: "foo((function() { this.bar(); }).bind(obj).bind(this))"},
			{Code: "test('clean', function (this: any) { this.foo = 'Cleaned!';});"},
			{Code: "obj.test('clean', function (foo) { this.foo = 'Cleaned!'; });"},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase("foo(function bar() {});", "foo(() => {});", nil, 1),
			invalidCase("foo(function(a:string) {});", "foo((a:string) => {});", allowNamed, 1),
			invalidCase("foo(function bar() {});", "foo(() => {});", disallowNamed, 1),
			invalidCase("foo(function() {});", "foo(() => {});", nil, 1),
			invalidCase("foo(nativeCb || function() {});", "foo(nativeCb || (() => {}));", nil, 1),
			invalidCase("foo(bar ? function() {} : function() {});", "foo(bar ? () => {} : () => {});", nil, 2),
			invalidCase("foo(function() { (function() { this; }); });", "foo(() => { (function() { this; }); });", nil, 1),
			invalidCase("foo(function() { this; }.bind(this));", "foo(() => { this; });", nil, 1),
			invalidCase("foo(bar || function() { this; }.bind(this));", "foo(bar || (() => { this; }));", nil, 1),
			invalidCase("foo(function() { (() => this); }.bind(this));", "foo(() => { (() => this); });", nil, 1),
			invalidCase("foo(function bar(a:string) { a; });", "foo((a:string) => { a; });", nil, 1),
			invalidCase("foo(function(a:any) { a; });", "foo((a:any) => { a; });", nil, 1),
			invalidCase("foo(function(arguments:any) { arguments; });", "foo((arguments:any) => { arguments; });", nil, 1),
			invalidCase("foo(function(a:string) { this; });", "", disallowUnboundThis, 1),
			invalidCase("foo(function() { (() => this); });", "", disallowUnboundThis, 1),
			invalidCase("qux(function(foo:string, bar:number, baz:string) { return foo * 2; })", "qux((foo:string, bar:number, baz:string) => { return foo * 2; })", nil, 1),
			invalidCase("qux(function(foo:number, bar:number, baz:number) { return foo * bar; }.bind(this))", "qux((foo:number, bar:number, baz:number) => { return foo * bar; })", nil, 1),
			invalidCase("qux(function(foo:any, bar:any, baz:any) { return foo * this.qux; }.bind(this))", "qux((foo:any, bar:any, baz:any) => { return foo * this.qux; })", nil, 1),
			invalidCase("foo(function() {}.bind(this, somethingElse))", "foo((() => {}).bind(this, somethingElse))", nil, 1),
			invalidCase("qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });", "qux((foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) => { return foo + bar; });", nil, 1),
			invalidCase("qux(function(baz:string, baz:string) { })", "", nil, 1),
			invalidCase("qux(function( /* no params */ ) { })", "qux(( /* no params */ ) => { })", nil, 1),
			invalidCase("qux(function( /* a */ foo:string /* b */ , /* c */ bar:string /* d */ , /* e */ baz:string /* f */ ) { return foo; })", "qux(( /* a */ foo:string /* b */ , /* c */ bar:string /* d */ , /* e */ baz:string /* f */ ) => { return foo; })", nil, 1),
			invalidCase("qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return baz; })", "qux(async (foo:number = 1, bar:number = 2, baz:number = 3) => { return baz; })", nil, 1),
			invalidCase("qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return this; }.bind(this))", "qux(async (foo:number = 1, bar:number = 2, baz:number = 3) => { return this; })", nil, 1),
			invalidCase("foo((bar || function() {}).bind(this))", "", nil, 1),
			invalidCase("foo(function() {}.bind(this).bind(obj))", "foo((() => {}).bind(obj))", nil, 1),
			invalidCase("foo?.(function() {});", "foo?.(() => {});", nil, 1),
			invalidCase("foo?.(function() { return this; }.bind(this));", "foo?.(() => { return this; });", nil, 1),
			invalidCase("foo(function() { return this; }?.bind(this));", "foo(() => { return this; });", nil, 1),
			invalidCase("foo((function() { return this; }?.bind)(this));", "", nil, 1),
			invalidCase(`
            test(
                function ()
                { }
            );
            `, `
            test(
                () =>
                { }
            );
            `, nil, 1),
			invalidCase(`
            test(
                function (
                    ...args
                ) /* Lorem ipsum
                dolor sit amet. */ {
                    return args;
                }
            );
            `, `
            test(
                (
                    ...args
                ) => /* Lorem ipsum
                dolor sit amet. */ {
                    return args;
                }
            );
            `, nil, 1),
			invalidCase("foo(function():string { return 'foo' });", "foo(():string => { return 'foo' });", nil, 1),
			invalidCase("test('foo', function (this: any) {});", "", nil, 1),
		},
	)
}
