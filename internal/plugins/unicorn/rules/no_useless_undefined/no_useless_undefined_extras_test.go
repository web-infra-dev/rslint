package no_useless_undefined_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_undefined"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUselessUndefinedExtras(t *testing.T) {
	bothOff := map[string]any{"checkArguments": false, "checkArrowFunctionBody": false}
	bothOn := map[string]any{"checkArguments": true, "checkArrowFunctionBody": true}
	escapedIdentifier := upstreamErrorAt(1, 29)
	escapedIdentifier[0].EndColumn = 43
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_undefined.NoUselessUndefinedRule, []rule_tester.ValidTestCase{
		{FileName: "file.ts", Code: "function f(): ((number | undefined)) { return undefined; }"},
		{FileName: "file.ts", Code: "function f() { return undefined as undefined; }"},
		{FileName: "file.ts", Code: "const f = () => undefined!;"},
		{FileName: "file.js", Code: "function f() { return void 0; }"},
		{FileName: "file.js", Code: "let undefined;"},
		{FileName: "file.js", Code: `function inspect(undefined) {
	let value = undefined;
	const arrow = () => undefined;
	function* generator() { yield undefined; }
	function accept(input = undefined) {}
	collect(undefined);
	const indexed = index >= 0 ? items[index] : undefined;
	return undefined;
}`},
		{FileName: "file.js", Code: "({[target = undefined]: x} = source);"},
		{FileName: "file.js", Code: "([x = (target = undefined)] = source);"},
		{FileName: "file.js", Code: "target = undefined;"},
		{FileName: "file.js", Code: "import(undefined);"},
		{FileName: "file.js", Code: `const a = [1, 2]; a["splice"](0, undefined); a["push"](undefined);`},
		{FileName: "file.js", Code: `items["sp" + "lice"](0, undefined);`},
		{FileName: "file.js", Code: "(items?.[`push`])(undefined);"},
		{FileName: "file.js", Code: `foo["bind"](value, undefined); (foo?.bind)(value, undefined);`},
		{FileName: "file.js", Code: `Reflect.get(obj, "x", undefined); Reflect["construct"](Object, [], undefined);`},
		{FileName: "file.js", Code: `globalThis["Reflect"]?.["get"](obj, "x", undefined); Reflect.get(...args, undefined);`},
		{FileName: "file.js", Code: `const value = i < items["length"] ? items[i] : undefined;`},
		{FileName: "file.js", Code: "const value = i <= items.length - 2 ? items[i] : undefined;"},
		{FileName: "file.js", Code: "const value = i >= 9007199254740992 ? items[i - 9007199254740992] : undefined;"},
		{FileName: "file.js", Code: "const value = i >= 0 ? items[++i] : undefined;"},
		{FileName: "file.js", Code: "const value = i >= 0 ? [get?.()][i] : undefined;"},
		{FileName: "file.js", Code: "const value = i >= 0 ? [String.raw`x`][i] : undefined;"},
		{FileName: "file.js", Code: "const value = i < 0 ? items[i] : undefined;"},
		{FileName: "file.js", Code: "const value = i >= 0 ? undefined : undefined;"},
		{FileName: "file.js", Code: "const noop = () => undefined; foo(undefined);", Options: bothOff},
	}, []rule_tester.InvalidTestCase{
		{
			FileName: "file.ts", Code: "function f(): (void) { return undefined; }",
			Output: []string{"function f(): (void) { return; }"}, Errors: upstreamErrorAt(1, 31),
		},
		{
			FileName: "file.ts", Code: "function f(): (undefined) { return undefined; }",
			Output: []string{"function f(): (undefined) { return; }"}, Errors: upstreamErrorAt(1, 36),
		},
		{
			FileName: "file.ts", Code: "function f({name}: any = undefined) {}",
			Errors: upstreamErrorAt(1, 26),
		},
		{
			FileName: "file.ts", Code: "class C { constructor(public value: number | undefined = undefined) {} } new C();",
			Errors: upstreamErrorAt(1, 58),
		},
		{
			FileName: "file.ts", Code: "function f(value: undefined = undefined, name: string) {}",
			Errors: upstreamErrorAt(1, 31),
		},
		{
			FileName: "file.ts", Code: "function f(value: undefined = undefined, name?: string) {}",
			Output: []string{"function f(value?: undefined, name?: string) {}"}, Errors: upstreamErrorAt(1, 31),
		},
		{
			FileName: "file.ts", Code: "function f(value: undefined = undefined, ...args: string[]) {}",
			Output: []string{"function f(value?: undefined, ...args: string[]) {}"}, Errors: upstreamErrorAt(1, 31),
		},
		{
			FileName: "file.ts", Code: "function f(value /* keep */ : T = undefined) {}",
			Output: []string{"function f(value /* keep */ ?: T) {}"}, Errors: upstreamErrorAt(1, 35),
		},
		{
			FileName: "file.ts", Code: "let value: T = (undefined);",
			Output: []string{"let value: T;"}, Errors: upstreamErrorAt(1, 17),
		},
		{
			FileName: "file.js", Code: "function f() { return /** @type {undefined} */ (undefined); }",
			Output: []string{"function f() { return /** @type {undefined} */; }"}, Errors: upstreamErrorAt(1, 49),
		},
		{
			FileName: "file.js", Code: "function f() { return(\r\nundefined); }",
			Output: []string{"function f() { return\r\n; }"}, Errors: upstreamErrorAt(2, 1),
		},
		{
			// cspell:disable-next-line
			FileName: "file.js", Code: `const emoji = "😀"; let x = u\u006edefined;`,
			Output: []string{`const emoji = "😀"; let x;`}, Errors: escapedIdentifier,
		},
		{
			FileName: "file.js", Code: "({x = (undefined)} = source);",
			Output: []string{"({x} = source);"}, Errors: upstreamErrorAt(1, 8),
		},
		{
			FileName: "file.js", Code: "({x: target.value = (undefined)} = source);",
			Output: []string{"({x: target.value} = source);"}, Errors: upstreamErrorAt(1, 22),
		},
		{
			FileName: "file.js", Code: "({key: (target.value) = undefined} = source);",
			Output: []string{"({key: (target.value)} = source);"}, Errors: upstreamErrorAt(1, 25),
		},
		{
			FileName: "file.js", Code: "[target.value = undefined] = source;",
			Output: []string{"[target.value] = source;"}, Errors: upstreamErrorAt(1, 17),
		},
		{
			FileName: "file.js", Code: "for ({x = undefined} of sources) {}",
			Output: []string{"for ({x} of sources) {}"}, Errors: upstreamErrorAt(1, 11),
		},
		{FileName: "file.js", Code: "foo(undefined /* keep */,);", Errors: upstreamErrorAt(1, 5)},
		{FileName: "file.js", Code: "foo(value, /* keep */ (undefined),);", Errors: upstreamErrorAt(1, 24)},
		{
			FileName: "file.js", Code: "foo(value, undefined /* keep */,);",
			Output: []string{"foo(value /* keep */,);"}, Errors: upstreamErrorAt(1, 12),
		},
		{
			FileName: "file.js", Code: "foo(/* keep */ undefined);",
			Output: []string{"foo(/* keep */ );"}, Errors: upstreamErrorAt(1, 16),
		},
		{
			FileName: "file.js", Code: `Reflect.get(obj, "x", undefined, undefined);`,
			Output: []string{`Reflect.get(obj, "x", undefined);`}, Errors: upstreamErrorAt(1, 34),
		},
		{
			FileName: "file.js", Code: "Reflect.get(obj, undefined);",
			Output: []string{"Reflect.get(obj);"}, Errors: upstreamErrorAt(1, 18),
		},
		{
			FileName: "file.js", Code: `const Reflect = {get() {}}; Reflect.get(obj, "x", undefined);`,
			Output: []string{`const Reflect = {get() {}}; Reflect.get(obj, "x");`}, Errors: upstreamErrorAt(1, 51),
		},
		{
			FileName: "file.js", Code: "class C { #ref() {} f() { this.#ref(undefined); } }",
			Output: []string{"class C { #ref() {} f() { this.#ref(); } }"}, Errors: upstreamErrorAt(1, 37),
		},
		{
			FileName: "file.js", Code: "foo(...args, undefined);",
			Output: []string{"foo(...args);"}, Errors: upstreamErrorAt(1, 14),
		},
		{
			FileName: "file.js", Code: "foo?.(undefined);",
			Output: []string{"foo?.();"}, Errors: upstreamErrorAt(1, 7),
		},
		{
			FileName: "file.jsx", Code: "const view = <Widget value={foo(undefined)} />;",
			Output: []string{"const view = <Widget value={foo()} />;"}, Errors: upstreamErrorAt(1, 33),
		},
		{
			FileName: "file.js", Code: "const value = i < items.length ? items[i] : undefined;",
			Errors: upstreamErrorAt(1, 45, "const value = items[i];"),
		},
		{
			FileName: "file.js", Code: "const value = items.length - +1 >= i ? items[i] : undefined;",
			Errors: upstreamErrorAt(1, 51, "const value = items[i];"),
		},
		{
			FileName: "file.js", Code: "const value = items.length - 1 < i ? undefined : items[i];",
			Errors: upstreamErrorAt(1, 38, "const value = items[i];"),
		},
		{
			FileName: "file.js", Code: "const value = i > -(-(-1)) ? items[i] : undefined;",
			Errors: upstreamErrorAt(1, 41, "const value = items[i];"),
		},
		{
			FileName: "file.js", Code: "const value = 4 >= i ? undefined : items[i - 5];",
			Errors: upstreamErrorAt(1, 24, "const value = items[i - 5];"),
		},
		{
			FileName: "file.js", Code: "const value = 4 < i ? items[i - 5] : undefined;",
			Errors: upstreamErrorAt(1, 38, "const value = items[i - 5];"),
		},
		{
			FileName: "file.js", Code: "function f() { return undefined; }", Options: bothOff,
			Output: []string{"function f() { return; }"}, Errors: upstreamErrorAt(1, 23),
		},
		{
			FileName: "file.js", Code: "const noop = () => undefined;", Options: bothOn,
			Output: []string{"const noop = () => {};"}, Errors: upstreamErrorAt(1, 20),
		},
		{
			FileName: "file.js", Code: "foo(undefined);", Options: map[string]any{},
			Output: []string{"foo();"}, Errors: upstreamErrorAt(1, 5),
		},
	})
}

func TestNoUselessUndefinedEditDemand(t *testing.T) {
	t.Parallel()
	const code = "let value = undefined;\nconst item = index >= 0 ? array[index] : undefined;"
	program, file := undefinedTestProgram(t, "file.js", code)
	all := lintUndefined(program, file, rule.EditDemandAll)
	if len(all) != 2 || len(all[0].Fixes()) != 1 || all[1].Suggestions == nil || len(*all[1].Suggestions) != 1 {
		t.Fatalf("expected one autofix and one suggestion, got %+v", all)
	}
	if got := (*all[1].Suggestions)[0].Message.Description; got != "Use the indexed access directly." {
		t.Errorf("unexpected suggestion message %q", got)
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		diagnostics := lintUndefined(program, file, demand)
		if len(diagnostics) != len(all) {
			t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), len(all))
		}
		for i, got := range diagnostics {
			want := all[i]
			if demand == rule.EditDemandNone || demand == rule.EditDemandSuggestion {
				want.FixesPtr = nil
			}
			if demand == rule.EditDemandNone || demand == rule.EditDemandAutofix {
				want.Suggestions = nil
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d, diagnostic %d:\ngot  %+v\nwant %+v", demand, i, got, want)
			}
		}
	}
}

func undefinedTestProgram(t *testing.T, name, code string) (*program.Program, string) {
	t.Helper()
	root := fixtures.GetRootDir()
	file := tspath.ResolvePath(root.Dir, name)
	fs := utils.NewOverlayVFS(root.FS, map[string]string{file: code})
	prog, err := program.NewFromRoots(program.RootOptions{
		RootFileNames:   []string{file},
		Host:            utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: program.SourceOnlyCompilerOptions(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return prog, file
}

func lintUndefined(prog *program.Program, file string, demand rule.EditDemand) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: prog,
		File:    file,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			r := no_useless_undefined.NoUselessUndefinedRule
			return []rule.ConfiguredRule{{Name: r.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return r.Run(ctx, nil)
			}}}
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}
