package no_useless_spread_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_spread"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessSpreadEditDemand(t *testing.T) {
	t.Parallel()
	const source = `const x = [...[foo]]; const y = [...foo.concat(bar)]; Object.assign(target, {...source}); new Set(...items);`
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/edit-demand.js", Path: "/edit-demand.js",
	}, source, core.ScriptKindJS)
	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		comments := rule.NewCommentStore(file)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile: file, Comments: comments, DisableManager: rule.NewDisableManager(file, comments),
		}.WithDiagnosticConsumer(no_useless_spread.NoUselessSpreadRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
		})
		listeners := no_useless_spread.NoUselessSpreadRule.Run(ctx, nil)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if listener := listeners[node.Kind]; listener != nil {
				listener(node)
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if len(diagnostics) != 4 {
			t.Fatalf("demand %d: got %d diagnostics, want 4", demand, len(diagnostics))
		}
		return diagnostics
	}
	all := run(rule.EditDemandAll)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		for index, actual := range run(demand) {
			expected := all[index]
			if demand&rule.EditDemandAutofix == 0 {
				expected.FixesPtr = nil
			}
			if demand&rule.EditDemandSuggestion == 0 {
				expected.Suggestions = nil
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("demand %d changed diagnostic %d or produced unexpected edits: got %+v, want %+v", demand, index, actual, expected)
			}
		}
	}
	want := `const x = [foo]; const y = foo.concat(bar); Object.assign(target, {...source}); new Set(...items);`
	if output, _, fixed := linter.ApplyRuleFixes(source, all); !fixed || output != want {
		t.Fatalf("autofix = %q, want %q", output, want)
	}
	want = `const x = [...[foo]]; const y = [...foo.concat(bar)]; Object.assign(target, source); new Set(...items);`
	if all[2].Suggestions == nil || len(*all[2].Suggestions) != 1 {
		t.Fatal("expected one Object.assign suggestion")
	}
	if output, _, fixed := linter.ApplyRuleFixes(source, *all[2].Suggestions); !fixed || output != want {
		t.Fatalf("suggestion = %q, want %q", output, want)
	}
}

// Additional runtime shapes and edit boundaries checked against Unicorn v77.0.0.
func TestNoUselessSpreadExtrasRuntimeShapes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_spread.NoUselessSpreadRule,
		[]rule_tester.ValidTestCase{
			// JavaScript inference must not act like upstream TypeScript parser
			// services when classifying a method's return type.
			{Code: "const a = [1,2]; const b = [...a.map(x => x)];", FileName: "file.js"},
			{Code: "const a = []; const b = [...a.filter(Boolean)];", FileName: "file.js"},
			{Code: "[...foo?.flat()]", FileName: "file.js"},
			{Code: "[...foo.flat?.()]", FileName: "file.js"},
			{Code: "[...(foo?.flat)()]", FileName: "file.js"},
			{Code: "[...foo['flat']()]", FileName: "file.js"},
			{Code: "[...foo[`flat`]()]", FileName: "file.js"},
			{Code: "class C { #flat() {} f() { return [...this.#flat()]; } }", FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "async function f(){for await (const value of [...items]);}", FileName: "file.js", Output: []string{"async function f(){for await (const value of items);}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 46, EndLine: 1, EndColumn: 56}}},
			{Code: "const b = [...((a?.b).flat)()];", FileName: "file.js", Output: []string{"const b = ((a?.b).flat)();"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 31}}},
			{Code: "const a = [1,2]; const b = [...a.slice(1)];", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 28, EndLine: 1, EndColumn: 43}}},
			{Code: "[...(foo.flat)()]", FileName: "file.js", Output: []string{"(foo.flat)()"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			{Code: "[...((foo.toReversed()))]", FileName: "file.js", Output: []string{"((foo.toReversed()))"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26}}},
			{Code: "Promise.all([...Array.from(foo)])", FileName: "file.js", Output: []string{"Promise.all(Array.from(foo))"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Promise.all(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 13, EndLine: 1, EndColumn: 33}, {MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 13, EndLine: 1, EndColumn: 33}}},
			{Code: "(Promise.all)(([...items]))", FileName: "file.js", Output: []string{"(Promise.all)((items))"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Promise.all(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 16, EndLine: 1, EndColumn: 26}}},
			{Code: "new (Set)([...items])", FileName: "file.js", Output: []string{"new (Set)(items)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`new Set(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 11, EndLine: 1, EndColumn: 21}}},
			{Code: "class C extends foo(...[bar]) {}", FileName: "file.js", Output: []string{"class C extends foo(bar) {}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in arguments is unnecessary.", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
			{Code: "function f(Set) { return new Set([...items]); }", FileName: "file.js", Output: []string{"function f(Set) { return new Set(items); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`new Set(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 34, EndLine: 1, EndColumn: 44}}},
			{Code: "Object.assign(target, ({...a, ...b}))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 25, EndLine: 1, EndColumn: 28, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, (a, b))"}}}}},
			{Code: "Object.assign(target, {...a}, ...rest)", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, a, ...rest)"}}}}},
			{Code: "foo?.(...[a,,b])", FileName: "file.js", Output: []string{"foo?.(a,undefined,b)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in arguments is unnecessary.", Line: 1, Column: 7, EndLine: 1, EndColumn: 10}}},
			{Code: "superFn(...[,,a,])", FileName: "file.js", Output: []string{"superFn(undefined,undefined,a)"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in arguments is unnecessary.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
			{Code: "const 名字 = \"😀\";\nfoo(\n  ...[a]\n);", FileName: "file.js", Output: []string{"const 名字 = \"😀\";\nfoo(\n  a\n);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in arguments is unnecessary.", Line: 3, Column: 3, EndLine: 3, EndColumn: 6}}},
		})
}

func TestNoUselessSpreadExtrasCommentsAndASI(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_spread.NoUselessSpreadRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: "Object.assign(target, {...\"/* string */\"});", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, \"/* string */\");"}}}}},
			{Code: "Object.assign(target, {.../\\/\\//});", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, /\\/\\//);"}}}}},
			{Code: "Object.assign(target, {...`/* ${value} */`});", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, `/* ${value} */`);"}}}}},
			{Code: "foo(... /* a */ ( /* b */ [ /* c */ a, /* d */ ] /* e */ ))", FileName: "file.js", Output: []string{"foo( /* a */  /* b */  /* c */ a /* d */  /* e */ )"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in arguments is unnecessary.", Line: 1, Column: 5, EndLine: 1, EndColumn: 8}}},
			{Code: "Object.assign(target, /* outside */ {...source})", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 38, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, /* outside */ source)"}}}}},
			{Code: "Object.assign(target, {... /* inside */ source})", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27}}},
			{Code: "Object.assign(target, {...(a /* inside */, b)})", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27}}},
			{Code: "Promise.all([ /* a */ ... /* b */ (items) /* c */, /* d */ ])", FileName: "file.js", Output: []string{"Promise.all( /* a */  /* b */ (items) /* c */ /* d */ )"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Promise.all(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 13, EndLine: 1, EndColumn: 61}}},
			{Code: "function f(){throw[\n...Object.keys(value)\n]}", FileName: "file.js", Output: []string{"function f(){throw (\nObject.keys(value)\n)}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 19, EndLine: 3, EndColumn: 2}}},
			{Code: "function f(){return /* keep */ [\n...Object.keys(value)\n];}", FileName: "file.js", Output: []string{"function f(){return ( /* keep */ \nObject.keys(value)\n);}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 32, EndLine: 3, EndColumn: 2}}},
			{Code: "function f(){return[...\nObject.keys(value)];}", FileName: "file.js", Output: []string{"function f(){return (\nObject.keys(value));}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 20, EndLine: 2, EndColumn: 20}}},
			{Code: "function f(){return[...Object.keys(value)\n];}", FileName: "file.js", Output: []string{"function f(){return Object.keys(value)\n;}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 20, EndLine: 2, EndColumn: 2}}},
			{Code: "function f(){return [... /* comment\n */ Object.keys(value)];}", FileName: "file.js", Output: []string{"function f(){return (  /* comment\n */ Object.keys(value));}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 21, EndLine: 2, EndColumn: 24}}},
			{Code: "function f(){return[ ...Object.keys(value) ]}", FileName: "file.js", Output: []string{"function f(){return ( Object.keys(value) )}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 20, EndLine: 3, EndColumn: 2}}},
			{Code: "function f(){throw([\n...Object.keys(value)\n]);}", FileName: "file.js", Output: []string{"function f(){throw (\nObject.keys(value)\n);}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 20, EndLine: 3, EndColumn: 2}}},
			{Code: "[.../** @type {string[]} */ (Object.keys(value))]", FileName: "file.js", Output: []string{"/** @type {string[]} */ (Object.keys(value))"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 1, EndLine: 1, EndColumn: 50}}},
			{Code: "const a = [.../** @type {string[]} */ ([value])];", FileName: "file.js", Output: []string{"const a = [/** @type {string[]} */ value];"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in array literal is unnecessary.", Line: 1, Column: 12, EndLine: 1, EndColumn: 15}}},
			{Code: "Object.assign(target, /** @type {object} */ ({...source}))", FileName: "file.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-object-assign", Message: "`Object.assign(…)` source object with only spread properties is unnecessary.", Line: 1, Column: 47, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion/remove-object-assign-spread", Output: "Object.assign(target, /** @type {object} */ (source))"}}}}},
		})
}

func TestNoUselessSpreadExtrasTypeScriptAndJSX(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_spread.NoUselessSpreadRule,
		[]rule_tester.ValidTestCase{
			{Code: "const a = [...([value] as string[])];", FileName: "file.ts"},
			{Code: "const a = [...(value.flat() as string[])];", FileName: "file.ts"},
			{Code: "Promise.all(([...items] satisfies Iterable<string>));", FileName: "file.ts"},
			{Code: "const el = <C {...{a}} />;", FileName: "file.tsx"},
			{Code: "class C { #flat(): string[] {return []} f(){ return [...this.#flat()] }}", FileName: "file.ts"},
			{Code: "function f(a: Int16Array) { return [...a.map(x => x)]; }", FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "for (const value of [...fn<string>]);", FileName: "file.ts", Output: []string{"for (const value of (fn<string>));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 36}}},
			{Code: "for (const value of [...<Iterable<string>>items]);", FileName: "file.ts", Output: []string{"for (const value of <Iterable<string>>items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 49}}},
			{Code: "function f(a: number[]) { return [...a.slice(1)]; }", FileName: "file.ts", Output: []string{"function f(a: number[]) { return a.slice(1); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "clone-array", Message: "Unnecessarily cloning an array.", Line: 1, Column: 34, EndLine: 1, EndColumn: 49}}},
			{Code: "const el = <C values={[...[a]]} />;", FileName: "file.tsx", Output: []string{"const el = <C values={[a]} />;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "spread-in-list", Message: "Spread an array literal in array literal is unnecessary.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27}}},
			{Code: "for(const value of [...items!]);", FileName: "file.ts", Output: []string{"for(const value of items!);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 31}}},
			{Code: "for(const value of [...items as string[]]);", FileName: "file.ts", Output: []string{"for(const value of items as string[]);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 42}}},
			{Code: "for(const value of [...items satisfies Iterable<string>]);", FileName: "file.ts", Output: []string{"for(const value of items satisfies Iterable<string>);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 57}}},
		})
}

func TestNoUselessSpreadExtrasPrecedenceAndConsumers(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_spread.NoUselessSpreadRule,
		[]rule_tester.ValidTestCase{
			{Code: "[...Object.keys(...args)]", FileName: "file.js"},
			{Code: "[...await Promise.all(...args)]", FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "for(const value of [...a || b]);", FileName: "file.js", Output: []string{"for(const value of (a || b));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 31}}},
			{Code: "for(const value of [...(a || b)]);", FileName: "file.js", Output: []string{"for(const value of (a || b));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 33}}},
			{Code: "for(const value of [...(a, b)]);", FileName: "file.js", Output: []string{"for(const value of (a, b));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 31}}},
			{Code: "async function f(){for(const value of [...await values]);}", FileName: "file.js", Output: []string{"async function f(){for(const value of (await values));}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 39, EndLine: 1, EndColumn: 56}}},
			{Code: "function* f(){yield*[...a ? b : c];}", FileName: "file.js", Output: []string{"function* f(){yield*(a ? b : c);}"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-yield-star", Message: "`yield*` can delegate to an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 35}}},
			{Code: "for(const value of [...this]);", FileName: "file.js", Output: []string{"for(const value of this);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29}}},
			{Code: "for(const value of [...null]);", FileName: "file.js", Output: []string{"for(const value of null);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29}}},
			{Code: "for(const value of [...`a${b}`]);", FileName: "file.js", Output: []string{"for(const value of `a${b}`);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 32}}},
			{Code: "for(const value of [...{}]);", FileName: "file.js", Output: []string{"for(const value of ({}));"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array-in-for-of", Message: "`for…of` can iterate over an iterable, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 27}}},
			{Code: "Int8Array.from([...items]);", FileName: "file.js", Output: []string{"Int8Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Int8Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 16, EndLine: 1, EndColumn: 26}}},
			{Code: "Uint8Array.from([...items]);", FileName: "file.js", Output: []string{"Uint8Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Uint8Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27}}},
			{Code: "Uint8ClampedArray.from([...items]);", FileName: "file.js", Output: []string{"Uint8ClampedArray.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Uint8ClampedArray.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 24, EndLine: 1, EndColumn: 34}}},
			{Code: "Int16Array.from([...items]);", FileName: "file.js", Output: []string{"Int16Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Int16Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27}}},
			{Code: "Uint16Array.from([...items]);", FileName: "file.js", Output: []string{"Uint16Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Uint16Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28}}},
			{Code: "Int32Array.from([...items]);", FileName: "file.js", Output: []string{"Int32Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Int32Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 17, EndLine: 1, EndColumn: 27}}},
			{Code: "Uint32Array.from([...items]);", FileName: "file.js", Output: []string{"Uint32Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Uint32Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 18, EndLine: 1, EndColumn: 28}}},
			{Code: "Float16Array.from([...items]);", FileName: "file.js", Output: []string{"Float16Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Float16Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
			{Code: "Float32Array.from([...items]);", FileName: "file.js", Output: []string{"Float32Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Float32Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
			{Code: "Float64Array.from([...items]);", FileName: "file.js", Output: []string{"Float64Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`Float64Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
			{Code: "BigInt64Array.from([...items]);", FileName: "file.js", Output: []string{"BigInt64Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`BigInt64Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 20, EndLine: 1, EndColumn: 30}}},
			{Code: "BigUint64Array.from([...items]);", FileName: "file.js", Output: []string{"BigUint64Array.from(items);"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "iterable-to-array", Message: "`BigUint64Array.from(…)` accepts an iterable as an argument, it's unnecessary to convert to an array.", Line: 1, Column: 21, EndLine: 1, EndColumn: 31}}},
		})
}
