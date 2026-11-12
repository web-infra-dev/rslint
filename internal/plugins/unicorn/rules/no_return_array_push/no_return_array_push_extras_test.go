package no_return_array_push_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_return_array_push"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoReturnArrayPushExtras(t *testing.T) {
	script := invalidCase("case.js", `function run() { return let[index].push(1); }`, "push", 1, 36, `function run() { (let[index].push(1)); return; }`)
	script.LanguageOptions.SourceType = "script"
	escaped := invalidCase("case.js", `const title = "🦄", length = items.p\u0075sh(value);`, "push", 1, 36, "")
	escaped.Errors[0].EndColumn = 45
	const queueCast = `class Queue {
  push(value) {
    return value;
  }
}

const queue = /** @type {Queue} */ (readQueue());
const value = queue.push(1);`
	const inferredStream = `function read() {
  return [];
}

const stream = read();
const length = stream.push(1);`
	const sameLineArray = `setup(); /** @param {Array<number>} stream */
function add(stream) {
  return stream.push(1);
}`

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_return_array_push.NoReturnArrayPushRule,
		[]rule_tester.ValidTestCase{
			{FileName: "case.js", Code: `function run() { return (items?.push)(value); }`},
			{FileName: "case.js", Code: `function run() { return process?.stdin.push(value); }`},
			{FileName: "case.js", Code: `class A { #push(x) {} run() { return this.#push(1); } }`},
			{FileName: "case.js", Code: `const x = new items.push(value);`},
			{FileName: "case.ts", Code: `(items.push as any)(value);`},
			{FileName: "case.ts", Code: `items.push(value)<string>;`},
			{FileName: "case.js", Code: `function run(...stream) { stream = replacement; return stream.push(1); }`},
			{FileName: "case.js", Code: `/** @param {URLSearchParams} values */ function run(values) { return values.push(1); }`},
			{FileName: "case.js", Code: `/** @returns {Set<number>} */ function get() {} function run() { return get().unshift(1); }`},
			// Only an ordinary assignment establishes the receiver from its right operand.
			{FileName: "case.js", Code: `let target; function run() { return (target = new Map()).push(1); }`},
			{FileName: "case.js", Code: `const second = []; const stream = input ?? second; const length = stream.push(1);`},
			{FileName: "case.js", Code: `/** @type {unknown} */ const stream = []; const length = stream.push(1);`},
			{FileName: "case.js", Code: `/** @returns {Array<number>} */ async function load() { return []; } const stream = load(); const length = stream.push(1);`},
			{FileName: "case.js", Code: `/** @returns {Array<number>} */ function* load() { yield 1; } const stream = load(); const length = stream.push(1);`},
			{FileName: "case.js", Code: `/** @type {() => Array<number>} */ const load = async () => []; const stream = load(); const length = stream.push(1);`},
			{FileName: "case.ts", Code: `function run() { return (items?.push!)(1); }`},
			{FileName: "case.js", Code: queueCast},
			{FileName: "case.js", Code: inferredStream},
			{FileName: "case.js", Code: `/** @param {function(): void} queue */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @type {function(): void} */
const queue = getQueue();
function add() { return queue.unshift(1); }`},
			{FileName: "case.js", Code: `setup(); /** @param {function(): void} queue */
function add(queue) { return queue.unshift(1); }`},
			{FileName: "case.js", Code: `setup(); /** @param {Set<number>} queue */
function add(queue) { return queue.unshift(1); }`},
			{FileName: "case.js", Code: `setup(); /** @param {() => void} queue */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @param {!function(): void} queue */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @type {?function(): {push: () => number}} */
const queue = getQueue();
function add() { return queue.unshift(1); }`},
			{FileName: "case.js", Code: `/** @param {(!function(): void) | string} queue */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @param {!function(): Set<number>} get */
function add(get) { return get().push(1); }`},
			{FileName: "case.js", Code: `/** @param {() => void} [queue] */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @param {Set<number>} [queue] */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @param {?Set<number>} queue */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `/** @param {Set<number>=} queue */
function add(queue) { return queue.push(1); }`},
			{FileName: "case.js", Code: `setup(); /** @param {Array<number>} [stream] */
function add(stream) { return stream?.push(1); }`},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase("case.js", `function run() { return (process?.stdin).push(value); }`, "push", 1, 42, `function run() { (process?.stdin).push(value); return; }`),
			invalidCase("case.js", `function run() { return this["router"].push(value); }`, "push", 1, 40, `function run() { this["router"].push(value); return; }`),
			invalidCase("case.js", `function run() { return items.push(...values); }`, "push", 1, 31, `function run() { items.push(...values); return; }`),
			invalidCase("case.js", `const x = target[items.push(value)];`, "push", 1, 24, ""),
			invalidCase("case.js", `const x = items.push(value)();`, "push", 1, 17, ""),
			invalidCase("case.jsx", `const el = <items.push title="🦄">{items.unshift(value)}</items.push>;`, "unshift", 1, 42, ""),
			invalidCase("case.ts", `function run() { return items.push(value)<string>; }`, "push", 1, 31, ""),
			invalidCase("case.ts", `function run() { return (items?.child)!.push(value); }`, "push", 1, 41, `function run() { (items?.child)!.push(value); return; }`),
			invalidCase("case.js", `function run(...stream) { return stream.push(1); }`, "push", 1, 41, `function run(...stream) { stream.push(1); return; }`),
			invalidCase("case.js", `function run() { const stream = Object.keys(obj); return stream.push("key"); }`, "push", 1, 65, `function run() { const stream = Object.keys(obj); stream.push("key"); return; }`),
			invalidCase("case.js", `function run() { const stream = [1, 2].sort(); return stream.push(3); }`, "push", 1, 62, `function run() { const stream = [1, 2].sort(); stream.push(3); return; }`),
			invalidCase("case.js", `/** @param {Array<number>} router */ function run(router) { return router.push(1); }`, "push", 1, 75, `/** @param {Array<number>} router */ function run(router) { router.push(1); return; }`),
			invalidCase("case.js", `/** @param {Array<number>} [items] */ function add(items) { return items.push(1); }`, "push", 1, 74, `/** @param {Array<number>} [items] */ function add(items) { items.push(1); return; }`),
			invalidCase("case.js", `/** @param {Array<!function(): void>} stream */ function add(stream) { return stream.push(1); }`, "push", 1, 86, `/** @param {Array<!function(): void>} stream */ function add(stream) { stream.push(1); return; }`),
			invalidCase("case.js", `/** @param {?function(): Set<number>} get */ function add(get) { return get().push(1); }`, "push", 1, 79, `/** @param {?function(): Set<number>} get */ function add(get) { get().push(1); return; }`),
			invalidCase("case.js", `const length = Buffer.from(data).push(value);`, "push", 1, 34, ""),
			invalidCase("case.js", `function run() { return (items.push(1) /* keep */); }`, "push", 1, 32, ""),
			invalidCase("case.js", "function run() { return (items.push(1)) // stay\n}", "push", 1, 32, "function run() { items.push(1); return; // stay\n}"),
			invalidCase("case.js", `function run() { return async function () {}().push(1); }`, "push", 1, 48, `function run() { (async function () {}().push(1)); return; }`),
			invalidCase("case.js", `function run() { return class { static list = [] }.list.push(1); }`, "push", 1, 57, `function run() { (class { static list = [] }.list.push(1)); return; }`),
			invalidCase("case.js", `function run() { return { values: items }.values.push(1); }`, "push", 1, 50, `function run() { ({ values: items }.values.push(1)); return; }`),
			invalidCase("case.js", "function run() { let local\nreturn [items].push(1); }", "push", 2, 16, "function run() { let local\n;[items].push(1); return; }"),
			invalidCase("case.js", "function run() { return\nreturn [items].push(1); }", "push", 2, 16, "function run() { return\n;[items].push(1); return; }"),
			invalidCase("case.js", "function run() { do {} while (ready)\nreturn [items].push(1); }", "push", 2, 16, "function run() { do {} while (ready)\n;[items].push(1); return; }"),
			invalidCase("case.js", `function run() { switch (x) { case 0: return items.push(x); } }`, "push", 1, 52, ""),
			invalidCase("case.js", `const first = []; const second = []; const stream = first ?? second; const length = stream.push(1);`, "push", 1, 92, ""),
			invalidCase("case.js", `let target; function run() { return (target ??= new Map()).push(1); }`, "push", 1, 60, `let target; function run() { (target ??= new Map()).push(1); return; }`),
			// The plain-array check for push and the indexed-collection check for unshift share this binding.
			invalidCase("case.js", `const stream = new Uint8Array(); const first = stream.push(1); const second = stream.unshift(1);`, "unshift", 1, 86, ""),
			invalidCase("case.js", `function run() { return /** @type {number} */ (items.push(1)); }`, "push", 1, 54, ""),
			invalidCase("case.ts", `function run() { return items.push<string>("x"); }`, "push", 1, 31, `function run() { items.push<string>("x"); return; }`),
			invalidCase("case.ts", `function make<T>(value: T): number[] { return []; } const stream = make(1); const length = stream.push(1);`, "push", 1, 99, ""),
			invalidCase("case.jsx", "function run() { setup = <View />\nreturn [items].push(1); }", "push", 2, 16, "function run() { setup = <View />\n;[items].push(1); return; }"),
			invalidCase("case.js", "/** @returns {number[]} */\n"+inferredStream, "push", 7, 23, ""),
			invalidCase("case.js", sameLineArray, "push", 3, 17, `setup(); /** @param {Array<number>} stream */
function add(stream) {
  stream.push(1); return;
}`),
			invalidCase("case.js", `setup(); /** @param {Set<number>} unused */
function add(unused, queue) {
  return queue.push(1);
}`, "push", 3, 16, `setup(); /** @param {Set<number>} unused */
function add(unused, queue) {
  queue.push(1); return;
}`),
			script,
			escaped,
		},
	)
}

func TestNoReturnArrayPushEditDemand(t *testing.T) {
	const source = "function f() { return (items.push(value)); }\nconst add = value => items.unshift(value);"
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, "case.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: source})
	compiler, err := utils.CreateProgramLenient(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	if compiler.GetSourceFile(file) == nil {
		t.Fatalf("missing source file %q", file)
	}
	sourceProgram := program.NewFromCompiler(compiler)
	lintRule := no_return_array_push.NoReturnArrayPushRule

	var all []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram,
			File:    file,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: lintRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return lintRule.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: got %d diagnostics, want 2", demand, len(diagnostics))
		}
		if demand == rule.EditDemandAll {
			all = diagnostics
			if all[0].Suggestions == nil || len(*all[0].Suggestions) != 1 || all[1].Suggestions != nil {
				t.Fatalf("expected one suggestion for the direct return: %#v", all)
			}
			if got := (*all[0].Suggestions)[0].Message.Data["method"]; got != "push" {
				t.Errorf("suggestion method = %q, want push", got)
			}
		}
		for i, got := range diagnostics {
			want := all[i]
			if demand&rule.EditDemandSuggestion == 0 {
				want.Suggestions = nil
			}
			if got.FixesPtr != nil || !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d, diagnostic %d: got %#v, want %#v", demand, i, got, want)
			}
		}
	}
}
