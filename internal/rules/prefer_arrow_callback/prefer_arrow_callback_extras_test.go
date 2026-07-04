// TestPreferArrowCallbackExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package prefer_arrow_callback

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferArrowCallbackExtras(t *testing.T) {
	disallowUnboundThisMap := map[string]any{"allowUnboundThis": false}
	allowNamedMap := map[string]any{"allowNamedFunctions": true}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferArrowCallbackRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: TS assertion wrappers are not transparent to ESLint's getCallbackInfo ----
			{Code: "foo((function() {}) as any);"},
			{Code: "foo((function() {})!);"},
			{Code: "foo((function() {}) satisfies Function);"},

			// ---- Dimension 4: generator and non-callback function container forms ----
			{Code: "foo(function* named() { yield 1; });"},
			{Code: "const callback = function() {};"},
			{Code: "const obj = { callback: function() {} };"},

			// Locks in upstream FunctionExpression:exit arm 1: allowNamedFunctions skips named callbacks.
			{Code: "foo(function keep() {});", Options: allowNamedMap},

			// Locks in upstream FunctionExpression:exit arm 3: recursive named functions are skipped.
			{Code: "foo(function recur(n) { return n ? recur(n - 1) : 0; });"},

			// Locks in upstream FunctionExpression:exit arm 3: recursive references from nested functions still count.
			{Code: "foo(function recur() { function inner() { return recur; } });"},

			// ---- Dimension 2: default parameters share the callback's `this`/`arguments` binding ----
			{Code: "foo(function(a = this.value) {});"},
			{Code: "foo(function(a = arguments[0]) {});"},

			// ---- Dimension 2: computed method names run in the outer callback scope ----
			{Code: "foo(function() { class C { [this.key]() {} } });"},
			{Code: "foo(function() { class C { [arguments[0]]() {} } });"},

			// Locks in upstream getCallbackInfo MemberExpression arm: non-bind members are not callbacks.
			{Code: "foo((function() {}).call);"},

			// Locks in upstream getCallbackInfo MemberExpression arm: `.bind` must be the callee.
			{Code: "foo((function() {}).bind.prop);"},

			// Locks in upstream FunctionExpression:exit arm 4: `arguments` in an arrow still belongs to the callback.
			{Code: "foo(function() { (() => arguments); });"},

			// Locks in upstream FunctionExpression:exit arm 4: implicit `arguments` remains unsafe even if a nested block shadows it.
			{Code: "foo(function() { arguments; { const arguments = 1; } });"},

			// ---- Real-user: dynamic `this` callbacks stay valid by default ----
			{Code: "button.addEventListener('click', function(event) { this.disabled = event.defaultPrevented; });"},
			{Code: "describe('suite', function() { this.timeout(1000); });"},
			{Code: "emitter.on('data', function(chunk) { this.emit('seen', chunk); });"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single parenthesized callback expression ----
			invalidCase("foo((function() {}));", "foo((() => {}));", nil, 1),

			// ---- Dimension 4: multi-level parenthesized callback expression ----
			invalidCase("foo(((function(a) { return a; })));", "foo((((a) => { return a; })));", nil, 1),

			// ---- Dimension 4: optional call target with callback argument ----
			invalidCase("foo?.bar?.(function(value) { return value; });", "foo?.bar?.((value) => { return value; });", nil, 1),

			// ---- Dimension 4: optional `.bind(this)` access ----
			invalidCase("foo(function() { return this.value; }?.bind(this));", "foo(() => { return this.value; });", nil, 1),

			// ---- Dimension 4: async callback with TypeScript return type ----
			invalidCase("foo(async function(value: number): Promise<number> { return value; });", "foo(async (value: number): Promise<number> => { return value; });", nil, 1),

			// ---- Dimension 4: scanner-based parameter paren lookup; comment text contains `(` ----
			invalidCase("foo(function /* comment ( */ (value) { return value; });", "foo( /* comment ( */ (value) => { return value; });", nil, 1),

			// ---- Dimension 4: same-kind nesting boundary; nested function `this` does not block outer report ----
			invalidCase("foo(function() { return function() { this; }; });", "foo(() => { return function() { this; }; });", nil, 1),

			// ---- Dimension 4: nested function's own `arguments` does not block the outer callback report ----
			invalidCase("foo(function() { function inner() { arguments; } return 1; });", "foo(() => { function inner() { arguments; } return 1; });", nil, 1),

			// ---- Dimension 4: local declaration named `arguments` shadows the implicit arguments binding ----
			invalidCase("foo(function() { function arguments() {} return arguments; });", "foo(() => { function arguments() {} return arguments; });", nil, 1),

			// ---- Dimension 4: catch binding named `arguments` is local, not the callback's implicit arguments object ----
			invalidCase("foo(function() { try { throw 1; } catch (arguments) { return arguments; } });", "foo(() => { try { throw 1; } catch (arguments) { return arguments; } });", nil, 1),

			// ---- Dimension 4: nested method `this` must not block the outer callback report ----
			invalidCase("foo(function() { return { m() { return this.x; } }; });", "foo(() => { return { m() { return this.x; } }; });", nil, 1),

			// ---- Dimension 4: member/access key forms are N/A; keys inside the callback do not affect matching ----
			invalidCase("foo(function() { return { ['x']: 1, '#x': 2 }; });", "foo(() => { return { ['x']: 1, '#x': 2 }; });", nil, 1),

			// ---- Dimension 4: graceful degradation for body-absent TS declarations inside callbacks ----
			invalidCase("foo(function() { declare function nested(): void; });", "foo(() => { declare function nested(): void; });", nil, 1),

			// ---- Dimension 4: empty arguments list ----
			invalidCase("new Foo(function() {});", "new Foo((() => {}));", nil, 1),

			// ---- Real-user: Promise callback ----
			invalidCase("Promise.resolve(1).then(function(value) { return value + 1; });", "Promise.resolve(1).then((value) => { return value + 1; });", nil, 1),

			// ---- Real-user: Array callback with bare object options uses ESLint defaults ----
			invalidCase("items.forEach(function(item) { item.save(); });", "items.forEach((item) => { item.save(); });", map[string]any{}, 1),

			// ---- Real-user: timer callback with lexical `this` binding ----
			invalidCase("setTimeout(function() { this.flush(); }.bind(this), 0);", "setTimeout(() => { this.flush(); }, 0);", nil, 1),

			// ---- Real-user: async task callback ----
			invalidCase("queueMicrotask(async function() { await flush(); });", "queueMicrotask(async () => { await flush(); });", nil, 1),

			// ---- Real-user: new Promise executor keeps upstream's parenthesized arrow fix ----
			invalidCase("new Promise(function(resolve, reject) { resolve(reject); });", "new Promise(((resolve, reject) => { resolve(reject); }));", nil, 1),

			// ---- Real-user: #16718 Allman-style body with a named callback ----
			invalidCase(`
            test(
                function named()
                { return 1; }
            );
            `, `
            test(
                () =>
                { return 1; }
            );
            `, nil, 1),

			// ---- Real-user: #16718 Allman-style body with parameter comment ----
			invalidCase(`
            queue.push(
                function (
                    value
                ) // keep parameter layout
                { return value; }
            );
            `, `
            queue.push(
                (
                    value
                ) => // keep parameter layout
                { return value; }
            );
            `, nil, 1),

			// Locks in upstream getCallbackInfo LogicalExpression arm.
			invalidCase("foo(left && function() {});", "foo(left && (() => {}));", nil, 1),

			// Locks in upstream getCallbackInfo ConditionalExpression arm.
			invalidCase("foo(cond ? noop : function() {});", "foo(cond ? noop : () => {});", nil, 1),

			// Locks in upstream FunctionExpression:exit arm 5: allowUnboundThis=false reports unbound `this` but emits no fix.
			invalidCase("foo(function() { return this.value; });", "", disallowUnboundThisMap, 1),

			// ---- Real-user: dynamic `this` callbacks report without fix when allowUnboundThis=false ----
			invalidCase("button.addEventListener('click', function(event) { this.disabled = event.defaultPrevented; });", "", disallowUnboundThisMap, 1),

			// Locks in upstream FunctionExpression:exit arm 5: `this` in default params is also unbound callback `this`.
			invalidCase("foo(function(a = this.value) {});", "", disallowUnboundThisMap, 1),

			// Locks in upstream FunctionExpression:exit arm 5: `this` in computed method names is outer callback `this`.
			invalidCase("foo(function() { class C { [this.key]() {} } });", "", disallowUnboundThisMap, 1),

			// Locks in upstream fix arm: duplicate simple params report but are not fixed.
			invalidCase("foo(function(a, a) { return a; });", "", nil, 1),

			// Locks in upstream fix arm: `.bind(this)` containing comments reports but is not fixed.
			invalidCase("foo(function() { return this; }.bind(/* keep */ this));", "", nil, 1),
		},
	)
}
