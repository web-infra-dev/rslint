// TestAlwaysReturnUpstream migrates the full valid/invalid suite from upstream
// __tests__/always-return.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in always_return_extras_test.go.

package always_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/always_return"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const thenMsg = "Each then() should return a value or throw"

func TestAlwaysReturnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&always_return.AlwaysReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid ----
			{Code: `hey.then(x => x)`},
			{Code: `hey.then(x => ({}))`},
			{Code: `hey.then(x => { return; })`},
			{Code: `hey.then(x => { return x ? x.id : null })`},
			{Code: `hey.then(x => { return x * 10 })`},
			{Code: `hey.then(x => { process.exit(0); })`},
			{Code: `hey.then(x => { process.abort(); })`},
			{Code: `hey.then(function() { return 42; })`},
			{Code: `hey.then(function() { return new Promise(); })`},
			{Code: `hey.then(function() { return "x"; }).then(doSomethingWicked)`},
			{Code: `hey.then(x => x).then(function() { return "3" })`},
			{Code: `hey.then(function() { throw new Error("msg"); })`},
			{Code: `hey.then(function(x) { if (!x) { throw new Error("no x"); } return x; })`},
			{Code: `hey.then(function(x) { if (x) { return x; } throw new Error("no x"); })`},
			{Code: `hey.then(function(x) { if (x) { process.exit(0); } throw new Error("no x"); })`},
			{Code: `hey.then(function(x) { if (x) { process.abort(); } throw new Error("no x"); })`},
			{Code: `hey.then(x => { throw new Error("msg"); })`},
			{Code: `hey.then(x => { if (!x) { throw new Error("no x"); } return x; })`},
			{Code: `hey.then(x => { if (x) { return x; } throw new Error("no x"); })`},
			{Code: `hey.then(x => { var f = function() { }; return f; })`},
			{Code: `hey.then(x => { if (x) { return x; } else { return x; } })`},
			{Code: `hey.then(x => { return x; var y = "unreachable"; })`},
			{Code: `hey.then(x => { return x; return "unreachable"; })`},
			{Code: `hey.then(x => { return; }, err=>{ log(err); })`},
			{Code: `hey.then(x => { return x && x(); }, err=>{ log(err); })`},
			{Code: `hey.then(x => { return x.y || x(); }, err=>{ log(err); })`},
			{Code: `hey.then(x => {
  return anotherFunc({
    nested: {
      one: x === 1 ? 1 : 0,
      two: x === 2 ? 1 : 0
    }
  })
})`},
			{Code: `hey.then(({x, y}) => {
  if (y) {
    throw new Error(x || y)
  }
  return x
})`},
			// ---- upstream valid: ignoreLastCallback ----
			{Code: `hey.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `if(foo) { hey.then(x => { console.log(x) }) }`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `void hey.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `
async function foo() {
  await hey.then(x => { console.log(x) })
}`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `hey?.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `foo = (hey.then(x => { console.log(x) }), 42)`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `(42, hey.then(x => { console.log(x) }))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `
hey
  .then(x => { console.log(x) })
  .catch(e => console.error(e))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `
hey
  .then(x => { console.log(x) })
  .catch(e => console.error(e))
  .finally(() => console.error('end'))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `
hey
  .then(x => { console.log(x) })
  .finally(() => console.error('end'))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			// ---- upstream valid: ignoreAssignmentVariable (default globalThis) ----
			{Code: `hey.then(x => { globalThis = x })`},
			{Code: `hey.then(x => { globalThis[a] = x })`},
			{Code: `hey.then(x => { globalThis.a = x })`},
			{Code: `hey.then(x => { globalThis.a.n = x })`},
			{Code: `hey.then(x => { globalThis[12] = x })`},
			{Code: `hey.then(x => { globalThis['12']["test"] = x })`},
			{Code: `hey.then(x => { window['x'] = x })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"globalThis", "window"}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid ----
			{
				Code:   `hey.then(x => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 10}},
			},
			{
				Code:   `hey.then(function() { })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 10}},
			},
			{
				Code:   `hey.then(function() { }).then(x)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 10}},
			},
			{
				Code: `hey.then(function() { }).then(function() { })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "thenShouldReturnOrThrow", Message: thenMsg},
					{MessageId: "thenShouldReturnOrThrow", Message: thenMsg},
				},
			},
			{
				Code:   `hey.then(function() { return; }).then(function() { })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:   `hey.then(function() { doSomethingWicked(); })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 10}},
			},
			{
				Code:   `hey.then(function() { if (x) { return x; } })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 23}},
			},
			{
				Code:   `hey.then(function() { if (x) { return x; } else { }})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 23}},
			},
			{
				Code:   `hey.then(function() { if (x) { } else { return x; }})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 23}},
			},
			{
				Code:   `hey.then(function() { if (x) { process.chdir(); } else { return x; }})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 23}},
			},
			{
				Code:   `hey.then(function() { if (x) { return you.then(function() { return x; }); } })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 23}},
			},
			{
				Code:   `hey.then( x => { x ? x.id : null })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 11}},
			},
			{
				Code:   `hey.then(function(x) { x ? x.id : null })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 1, Column: 10}},
			},
			{
				Code: `(function() {
  return hey.then(x => {
    anotherFunc({
      nested: {
        one: x === 1 ? 1 : 0,
        two: x === 2 ? 1 : 0
      }
    })
  })
})()`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code: `
hey.then(({x, y}) => {
  if (y) {
    throw new Error(x || y)
  }
})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 3, Column: 3}},
			},
			{
				Code: `
hey.then(({x, y}) => {
  if (y) {
    return x
  }
})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 3, Column: 3}},
			},
			// ---- upstream invalid: ignoreLastCallback ----
			{
				Code: `
hey
  .then(function(x) { console.log(x) /* missing return here */ })
  .then(function(y) { console.log(y) /* no error here */ })`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg, Line: 3}},
			},
			{
				Code:    `const foo = hey.then(function(x) {});`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code: `
function foo() {
  return hey.then(function(x) {});
}`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code: `
async function foo() {
  return await hey.then(x => { console.log(x) })
}`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:    `const foo = hey?.then(x => { console.log(x) })`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:    `const foo = (42, hey.then(x => { console.log(x) }))`,
				Options: map[string]interface{}{"ignoreLastCallback": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			// ---- upstream invalid: ignoreAssignmentVariable ----
			{
				Code:   `hey.then(x => { invalid = x })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:   `hey.then(x => { invalid['x'] = x })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				// cspell:ignore windo
				Code:    `hey.then(x => { windo[x] = x })`,
				Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:    `hey.then(x => { windo['x'] = x })`,
				Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:    `hey.then(x => { windows['x'] = x })`,
				Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
			{
				Code:    `hey.then(x => { x() })`,
				Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: thenMsg}},
			},
		},
	)
}
