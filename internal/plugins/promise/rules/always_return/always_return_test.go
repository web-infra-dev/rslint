package always_return_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/always_return"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const message = "Each then() should return a value or throw"

func TestAlwaysReturn(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&always_return.AlwaysReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
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
			{Code: `hey.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `if(foo) { hey.then(x => { console.log(x) }) }`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `void hey.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `async function foo() {
        await hey.then(x => { console.log(x) })
      }`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `hey?.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `foo = (hey.then(x => { console.log(x) }), 42)`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `(42, hey.then(x => { console.log(x) }))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `hey
        .then(x => { console.log(x) })
        .catch(e => console.error(e))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `hey
        .then(x => { console.log(x) })
        .catch(e => console.error(e))
        .finally(() => console.error('end'))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `hey
        .then(x => { console.log(x) })
        .finally(() => console.error('end'))`, Options: map[string]interface{}{"ignoreLastCallback": true}},
			{Code: `hey.then(x => { globalThis = x })`},
			{Code: `hey.then(x => { globalThis[a] = x })`},
			{Code: `hey.then(x => { globalThis.a = x })`},
			{Code: `hey.then(x => { globalThis.a.n = x })`},
			{Code: `hey.then(x => { globalThis[12] = x })`},
			{Code: `hey.then(x => { globalThis['12']["test"] = x })`},
			{Code: `hey.then(x => { window['x'] = x })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"globalThis", "window"}}},

			// ---- Upstream semantic branches and tsgo shape locks ----
			// Locks in upstream isMemberCall(): computed `.then` is ignored.
			{Code: `hey['then'](x => { console.log(x) })`},
			// Locks in upstream isFirstArgument(): only the first callback is checked.
			{Code: `hey.then(ok, x => { console.log(x) })`},
			// Parentheses are transparent in ESTree; rslint unwraps them explicitly.
			{Code: `hey.then((function() { return 1; }))`},
			// Expression-bodied arrow is valid even if it is a void expression.
			{Code: `hey.then(x => void console.log(x))`},

			// ---- switch statement — false-positive regression (review feedback) ----
			// switch with default and every case terminating: should NOT error
			{Code: `hey.then(function(x) { switch (x) { case 1: return 'a'; case 2: return 'b'; default: return 'c'; } })`},
			// switch with fallthrough case (empty case falls through to returning case)
			{Code: `hey.then(function(x) { switch (x) { case 1: case 2: return 'a'; default: return 'b'; } })`},
			// switch where every case throws
			{Code: `hey.then(function(x) { switch (x) { case 1: throw new Error(); default: throw new Error(); } })`},

			// ---- infinite loop — false-positive regression (review feedback) ----
			// Infinite loops without an exiting break should be terminal even when the body falls through.
			{Code: `hey.then(function() { while (true) {} })`},
			{Code: `hey.then(function(x) { while (true) { if (x) return 1; } })`},
			{Code: `hey.then(function() { while (true) { x++; } })`},
			{Code: `hey.then(function() { do {} while (true) })`},
			{Code: `hey.then(function() { while (1) { return 1; } })`},
			{Code: `hey.then(function() { while ('truthy') { return 1; } })`},

			// ---- coverage gaps surfaced in review: for / labeled / do-while(false) ----
			{Code: `hey.then(function() { for (;;) { return 1; } })`},
			{Code: `hey.then(function() { outer: { return 1; } })`},
			{Code: `hey.then(function() { do { return 1; } while (false) })`},

			// ---- switch: a conditional break must not be masked, yet these still terminate ----
			{Code: `hey.then(function(x) { switch (x) { case 1: { return 1; } default: return 2; } })`},
			{Code: `hey.then(function(x) { switch (x) { case 1: if (y) return 1; else return 2; default: return 3; } })`},
			{Code: `hey.then(function(x) { switch (x) { case 1: if (y) return 1; default: return 2; } })`},
			{Code: `hey.then(function(x) { switch (x) { case 1: while (true) { break; } return 'a'; default: return 'b'; } })`},

			// ---- ignoreAssignmentVariable: compound assignment to an ignored var (default globalThis) ----
			{Code: `hey.then(x => { globalThis.a += x })`},
			{Code: `hey.then(x => { globalThis.a ||= x })`},
			{Code: `hey.then(x => { globalThis.a ??= x })`},
			{Code: `hey.then(x => { globalThis[x] -= 1 })`},
			{Code: `hey.then(x => { window.a += x })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}}},

			// ---- try/catch: catch unreachable because the try block cannot throw ----
			{Code: `hey.then(function() { try { return 1 } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return -1 } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return {a: 1} } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return [1, 2] } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return true ? 1 : 2 } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return 1; foo() } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { let a = 1; return 1 } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return "x" } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return void 0 } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { return 1 + 2 } catch (e) { log(e) } })`},
			{Code: "hey.then(function() { try { return `a${1}` } catch (e) { log(e) } })"},
			{Code: `hey.then(function() { try { ; return 1 } catch (e) { log(e) } })`},
			{Code: `hey.then(function() { try { 1; return 2 } catch (e) { log(e) } })`},

			// ---- switch: break reachability inside a clause ----
			{Code: `hey.then(function(x) { switch (x) { case 1: { return 1; break; } default: return 2; } })`},
			{Code: `hey.then(function(x) { switch (x) { case 1: ; return 1; default: return 2; } })`},
			{Code: `hey.then(function() { try { return 'a' in {} } catch (e) { log(e) } })`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ESLint upstream invalid cases ----
			{Code: `hey.then(x => {})`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message, Line: 1, Column: 10}}},
			{Code: `hey.then(function() { })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message, Line: 1, Column: 10}}},
			{Code: `hey.then(function() { }).then(x)`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { }).then(function() { })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}, {MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { return; }).then(function() { })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { doSomethingWicked(); })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { if (x) { return x; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { if (x) { return x; } else { }})`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { if (x) { } else { return x; }})`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { if (x) { process.chdir(); } else { return x; }})`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { if (x) { return you.then(function() { return x; }); } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then( x => { x ? x.id : null })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function(x) { x ? x.id : null })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `(function() {
        return hey.then(x => {
          anotherFunc({
            nested: {
              one: x === 1 ? 1 : 0,
              two: x === 2 ? 1 : 0
            }
          })
        })
      })()`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(({x, y}) => {
        if (y) {
          throw new Error(x || y)
        }
      })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(({x, y}) => {
        if (y) {
          return x
        }
      })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey
        .then(function(x) { console.log(x) /* missing return here */ })
        .then(function(y) { console.log(y) /* no error here */ })`, Options: map[string]interface{}{"ignoreLastCallback": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message, Line: 2}}},
			{Code: `const foo = hey.then(function(x) {});`, Options: map[string]interface{}{"ignoreLastCallback": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `function foo() {
        return hey.then(function(x) {});
      }`, Options: map[string]interface{}{"ignoreLastCallback": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `async function foo() {
        return await hey.then(x => { console.log(x) })
      }`, Options: map[string]interface{}{"ignoreLastCallback": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `const foo = hey?.then(x => { console.log(x) })`, Options: map[string]interface{}{"ignoreLastCallback": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `const foo = (42, hey.then(x => { console.log(x) }))`, Options: map[string]interface{}{"ignoreLastCallback": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { invalid = x })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { invalid['x'] = x })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { const value = x })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { notWindow[x] = x })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { notWindow['x'] = x })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { windows['x'] = x })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(x => { x() })`, Options: map[string]interface{}{"ignoreAssignmentVariable": []interface{}{"window"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},

			// Locks in upstream process matcher: only process.exit/abort is terminal.
			{Code: `hey.then(x => { process.exitCode = 1 })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},

			// ---- switch statement — false-positive regression (review feedback) ----
			// switch without default: entry via any case can fall through to end, not terminal
			{Code: `hey.then(function(x) { switch (x) { case 1: return 'a'; case 2: return 'b'; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			// switch where one case breaks before returning
			{Code: `hey.then(function(x) { switch (x) { case 1: return 'a'; case 2: break; default: return 'c'; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			// switch where default case body doesn't terminate
			{Code: `hey.then(function(x) { switch (x) { case 1: return 'a'; default: console.log('nope'); } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			// switch where a case conditionally breaks out before its return (review fix)
			{Code: `hey.then(function(x) { switch (x) { case 1: if (y) break; return 'a'; default: return 'b'; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			// same, with the conditional break nested inside a block
			{Code: `hey.then(function(x) { switch (x) { case 1: { if (y) break; } return 'a'; default: return 'b'; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			// compound assignment to a non-ignored variable is still reported
			{Code: `hey.then(x => { notGlobal.a += x })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},

			// ---- try/catch: catch reachable because the try block may throw ----
			{Code: `hey.then(function() { try { return foo() } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { return x } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { return {[k]: 1} } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { foo(); return 1 } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { return x.y } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { return [...x] } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { return {...x} } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: "hey.then(function() { try { return `a${x}` } catch (e) { log(e) } })", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { let a = foo(); return 1 } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { let {a} = x; return 1 } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},

			// ---- switch: break (else-branch / labeled) exits the switch, not terminal ----
			{Code: `hey.then(function(x) { switch (x) { case 1: if (y) return 1; else break; default: return 2; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function(x) { outer: switch (x) { case 1: if (y) break outer; return 1; default: return 2; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function(x) { switch (x) { case 1: lbl: if (y) break; return 1; default: return 2; } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { return x instanceof Object } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
			{Code: `hey.then(function() { try { throw new Error() } catch (e) { log(e) } })`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "thenShouldReturnOrThrow", Message: message}}},
		},
	)
}
