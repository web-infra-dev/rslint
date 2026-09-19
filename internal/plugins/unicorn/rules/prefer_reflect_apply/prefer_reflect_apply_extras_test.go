package prefer_reflect_apply_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_reflect_apply"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferReflectApplyExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `new foo.apply(null, []);`},
		{Code: `apply(null, []);`},
		{Code: `foo.call(null, []);`},
		{Code: `foo.apply(null, [], extra);`},
		{Code: `foo.apply(undefined, []);`},
		{Code: `foo.apply(null, args);`},
		{Code: `foo.apply(...args);`},
		{Code: `foo.apply(null, ...arguments);`},
		{Code: `foo.apply?.(null, []);`},
		{Code: `foo?.apply(null, []);`},
		{Code: `(foo?.apply)(null, []);`},
		{Code: `foo?.["apply"](null, []);`},
		{Code: `foo[method](null, []);`},
		{Code: `foo["call"](null, []);`},
		{Code: `class C { #apply; f() { this.#apply(null, []); } }`},
		{Code: `"text".apply(null, []);`},
		{Code: `(42).apply(null, []);`},
		{Code: `42n.apply(null, []);`},
		{Code: `/pattern/.apply(null, []);`},
		{Code: `true.apply(null, []); false.apply(null, []); null.apply(null, []);`},
		{Code: `({}).apply(null, []);`},
		{Code: `Function.prototype.apply.call(foo, null);`},
		{Code: `Function.prototype.apply.call(foo, null, [], extra);`},
		{Code: `Function.prototype.apply.call(foo, other, []);`},
		{Code: `Function.prototype.apply.call(foo, null, args);`},
		{Code: `Other.prototype.apply.call(foo, null, []);`},
		{Code: `globalThis.Function.prototype.apply.call(foo, null, []);`},
		{Code: `Function.other.apply.call(foo, null, []);`},
		{Code: `Function.prototype.other.call(foo, null, []);`},
		{Code: `Function.apply.call(foo, null, []);`},
		{Code: `(Function?.prototype.apply).call(foo, null, []);`},
		// Authored TS wrappers are not erased by ESTree.
		{Code: `(foo.apply as Function)(null, []);`, FileName: "file.ts"},
		{Code: `foo.apply(null as any, []);`, FileName: "file.ts"},
		{Code: `foo.apply(null, [] as any[]);`, FileName: "file.ts"},
		{Code: `foo.apply(null, arguments!);`, FileName: "file.ts"},
		{Code: `foo.apply(null, arguments satisfies unknown);`, FileName: "file.ts"},
	}
	invalid := []rule_tester.InvalidTestCase{}
	for _, test := range []struct{ code, output string }{
		{`((foo)).apply((null), ([42]));`, `Reflect.apply(foo, null, [42]);`},
		{`(foo.apply)(this, (arguments));`, `Reflect.apply(foo, this, arguments);`},
		{`((Function).prototype.apply).call((foo), (null), ([]));`, `Reflect.apply(foo, null, []);`},
		{`foo["ap" + "ply"](null, []);`, `Reflect.apply(foo, null, []);`},
		{"foo[`apply`](null, []);", `Reflect.apply(foo, null, []);`},
		{`foo[true ? "apply" : "call"](null, []);`, `Reflect.apply(foo, null, []);`},
		{`foo[["apply"]](null, []);`, `Reflect.apply(foo, null, []);`},
		{`Function["prototype"]["ap" + "ply"]["call"](foo, null, []);`, `Reflect.apply(foo, null, []);`},
		{`f\u006fo.app\u006cy(null, []);`, `Reflect.apply(f\u006fo, null, []);`},
		{"`text`.apply(null, []);", "Reflect.apply(`text`, null, []);"},
		{`(function () {}).apply(this, []);`, `Reflect.apply(function () {}, this, []);`},
		{`(() => {}).apply(null, []);`, `Reflect.apply(() => {}, null, []);`},
		{`foo.apply(null, [first, ...rest]);`, `Reflect.apply(foo, null, [first, ...rest]);`},
		{`foo /* method */ .apply(/* receiver */null, [/* keep */42]);`, `Reflect.apply(foo, null, [/* keep */42]);`},
		// Source parentheses/JSDoc casts are transparent, unlike TS assertions.
		{`(/** @type {Function} */ (foo)).apply(null, []);`, `Reflect.apply(foo, null, []);`},
		{`foo.apply(/** @type {any} */ (null), []);`, `Reflect.apply(foo, null, []);`},
		// Narrow safety differences from upstream, documented beside the rule.
		{`(first, second).apply(null, []);`, `Reflect.apply((first, second), null, []);`},
		{`Function.prototype.apply.call((first, second), null, []);`, `Reflect.apply((first, second), null, []);`},
		{`(foo?.bar).apply(null, []);`, `Reflect.apply(foo?.bar, null, []);`},
		// The safety check must not reject effects that the fix preserves.
		{`getFunction().apply(null, [effect()]);`, `Reflect.apply(getFunction(), null, [effect()]);`},
		{`object[getKey()].apply(null, []);`, `Reflect.apply(object[getKey()], null, []);`},
		{`Function.prototype.apply.call(getFunction(), this, [effect()]);`, `Reflect.apply(getFunction(), this, [effect()]);`},
		{`fn[(0, "apply")](null, []);`, `Reflect.apply(fn, null, []);`},
		{`fn[({method: "apply"}).method](null, []);`, `Reflect.apply(fn, null, []);`},
		{`fn[(() => touch(), "apply")](null, []);`, `Reflect.apply(fn, null, []);`},
	} {
		invalid = append(invalid, applyInvalid(test.code, strings.TrimSuffix(test.code, ";"), test.output))
	}
	invalid = append(invalid,
		applyInvalid(`"😀"; fn.apply(null, []);`, `fn.apply(null, [])`, `"😀"; Reflect.apply(fn, null, []);`),
		applyInvalid("fn.apply(\n  null,\n  [42]\n);", "fn.apply(\n  null,\n  [42]\n)", `Reflect.apply(fn, null, [42]);`),
		applyInvalid(`foo?.bar.apply(null, []);`, `foo?.bar.apply(null, [])`),
		applyInvalid(`Function?.prototype.apply.call(foo, null, []);`, `Function?.prototype.apply.call(foo, null, [])`),
		applyInvalid(`class C extends B { f() { super.apply(null, []); } }`, `super.apply(null, [])`),
		applyInvalid(`class C extends B { f() { super.method.apply(this, arguments); } }`, `super.method.apply(this, arguments)`,
			`class C extends B { f() { Reflect.apply(super.method, this, arguments); } }`),
		// A statically known key can still execute a call or assignment.
		applyInvalid(`const events = []; function fn() { events.push("called"); } fn[(events.push("key"), "apply")](null, []); events;`,
			`fn[(events.push("key"), "apply")](null, [])`),
		applyInvalid(`let key; function fn() {} fn[key = "apply"](null, []); key;`, `fn[key = "apply"](null, [])`),
		applyInvalid(`const events = []; function fn() { events.push("called"); } Function[(events.push("key"), "prototype")].apply.call(fn, null, []); events;`,
			`Function[(events.push("key"), "prototype")].apply.call(fn, null, [])`),
		// Expanding the first argument changes the actual argument positions:
		// this call returns 0, but Reflect.apply(...targets, null, []) throws.
		applyInvalid(`function fn() { return arguments.length; } const targets = [fn, null]; Function.prototype.apply.call(...targets, null, []);`,
			`Function.prototype.apply.call(...targets, null, [])`),
	)
	for _, code := range []string{
		`Function.prototype[(touch(), "apply")].call(fn, null, []);`,
		`Function.prototype.apply[(touch(), "call")](fn, null, []);`,
		`fn[[(touch(), "apply")]](null, []);`,
		`fn[(value++, "apply")](null, []);`,
		`fn[(++value, "apply")](null, []);`,
		`fn[(delete object.key, "apply")](null, []);`,
		`fn[(void touch(), "apply")](null, []);`,
		// A property read may invoke a getter, even in an ignored operand.
		`fn[(object.key, "apply")](null, []);`,
		`fn[(object[key], "apply")](null, []);`,
		// Class creation can read getters in heritage expressions.
		`fn[(class extends object.Base {}, "apply")](null, []);`,
		// Tagged templates and spread can execute code without a CallExpression.
		"fn[((() => touch())`key`, \"apply\")](null, []);",
		`fn[([...iterable], "apply")](null, []);`,
		`fn[({...{get value() { touch(); }}}, "apply")](null, []);`,
		// Dropping an unresolved read would also suppress its ReferenceError.
		`fn[(unknownReference, "apply")](null, []);`,
		`fn[(this, "apply")](null, []);`,
		// Ignored operands can throw or invoke a coercion hook without calls
		// in the enclosing expression's AST.
		`fn[(1n / 0n, "apply")](null, []);`,
		`fn[(+1n, "apply")](null, []);`,
		"fn[(`${{toString: function () { touch(); return \"key\"; }}}`, \"apply\")](null, []);",
		// Conservatively reject calls even on a statically unreachable branch.
		`fn[true ? "apply" : touch()](null, []);`,
		// Even a one-element literal spread is deliberately left unchanged.
		`Function.prototype.apply.call(...[fn], this, arguments);`,
	} {
		invalid = append(invalid, applyInvalid(code, strings.TrimSuffix(code, ";")))
	}
	invalid = append(invalid,
		applyInvalid(`async function run() { fn[(await getKey(), "apply")](null, []); }`, `fn[(await getKey(), "apply")](null, [])`),
	)
	for _, code := range []string{
		`fn[(key = "apply") as string](null, []);`,
		`fn[((touch(), "apply") satisfies string)](null, []);`,
	} {
		item := applyInvalid(code, strings.TrimSuffix(code, ";"))
		item.FileName = "file.ts"
		invalid = append(invalid, item)
	}
	jsx := applyInvalid(`fn[(<Component />, "apply")](null, []);`, `fn[(<Component />, "apply")](null, [])`)
	jsx.FileName, jsx.Tsx = "file.tsx", true
	invalid = append(invalid, jsx)
	for _, test := range []struct{ code, output string }{
		{`(foo as Function).apply(null, []);`, `Reflect.apply(foo as Function, null, []);`},
		{`foo!.apply(null, []);`, `Reflect.apply(foo!, null, []);`},
		{`foo.apply<number>(null, []);`, `Reflect.apply(foo, null, []);`},
		{`foo["apply" as const](null, []);`, `Reflect.apply(foo, null, []);`},
	} {
		item := applyInvalid(test.code, strings.TrimSuffix(test.code, ";"), test.output)
		item.FileName = "file.ts"
		invalid = append(invalid, item)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_reflect_apply.PreferReflectApplyRule, valid, invalid)
}

func TestPreferReflectApplyEditDemand(t *testing.T) {
	for _, test := range []struct{ code, output string }{
		{`fn.apply(null, [])`, `Reflect.apply(fn, null, [])`},
		{`Function.prototype.apply.call(fn, this, arguments)`, `Reflect.apply(fn, this, arguments)`},
		{`(first, second).apply(null, [])`, `Reflect.apply((first, second), null, [])`},
		{`foo?.bar.apply(null, [])`, ""},
		{`class C extends B { f() { super.apply(null, []); } }`, ""},
		{`fn[(touch(), "apply")](null, [])`, ""},
		{`Function[(touch(), "prototype")].apply.call(fn, null, [])`, ""},
		{`Function.prototype[(touch(), "apply")].call(fn, null, [])`, ""},
		{`Function.prototype.apply[(touch(), "call")](fn, null, [])`, ""},
		{`Function.prototype.apply.call(...targets, null, [])`, ""},
		{`object[getKey()].apply(null, [])`, `Reflect.apply(object[getKey()], null, [])`},
		{`fn[(class extends object.Base {}, "apply")](null, [])`, ""},
		{"fn[((() => touch())`key`, \"apply\")](null, [])", ""},
		{`fn[(1n / 0n, "apply")](null, [])`, ""},
	} {
		t.Run(test.code, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/edit-demand.js", Path: "/edit-demand.js",
			}, test.code, core.ScriptKindJS)
			var baseline rule.RuleDiagnostic
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll,
			} {
				var diagnostics []rule.RuleDiagnostic
				ctx := rule.RuleContext{SourceFile: sourceFile}.WithDiagnosticConsumer(
					prefer_reflect_apply.PreferReflectApplyRule.Name, rule.SeverityError,
					rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
						diagnostics = append(diagnostics, d)
					}})
				listeners := prefer_reflect_apply.PreferReflectApplyRule.Run(ctx, nil)
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if listener := listeners[node.Kind]; listener != nil {
						listener(node)
					}
					return node.ForEachChild(visit)
				}
				visit(sourceFile.AsNode())
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				diagnostic := diagnostics[0]
				if demand == rule.EditDemandNone {
					baseline = diagnostic
				}
				withoutFixes := diagnostic
				withoutFixes.FixesPtr = nil
				if !reflect.DeepEqual(withoutFixes, baseline) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := demand&rule.EditDemandAutofix != 0 && test.output != ""
				if (diagnostic.FixesPtr != nil) != wantFix {
					t.Fatalf("demand %d: unexpected fixes: %+v", demand, diagnostic.FixesPtr)
				}
				if wantFix {
					output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
					if !fixed || output != test.output {
						t.Errorf("demand %d: fixed = %v, output = %q, want %q", demand, fixed, output, test.output)
					}
				}
			}
		})
	}
}
