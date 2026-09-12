package no_new_buffer_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_new_buffer "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_new_buffer"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewBufferExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_new_buffer.NoNewBufferRule,
		[]rule_tester.ValidTestCase{
			{Code: `new globalThis.Buffer(1)`, FileName: "file.js"},
			{Code: `new BufferFactory(1)`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			withFileName(fixedNewBufferCase(`new Buffer<string>`, `new Buffer<string>`, "from", `Buffer.from<string>()`), "file.ts"),
			fixedNewBufferCase("new\vBuffer(1)", "new\vBuffer(1)", "alloc", "Buffer.alloc(1)"),
			fixedNewBufferCase("new\fBuffer(1)", "new\fBuffer(1)", "alloc", "Buffer.alloc(1)"),
			fixedNewBufferCase("new\u00a0Buffer(1)", "new\u00a0Buffer(1)", "alloc", "Buffer.alloc(1)"),
			fixedNewBufferCase(`const number = 1; new Buffer(number)`, `new Buffer(number)`, "alloc", `const number = 1; Buffer.alloc(number)`),
			fixedNewBufferCase(`const string = "x"; new Buffer(string)`, `new Buffer(string)`, "from", `const string = "x"; Buffer.from(string)`),
			fixedNewBufferCase(`const bytes = [1]; new Buffer(bytes)`, `new Buffer(bytes)`, "from", `const bytes = [1]; Buffer.from(bytes)`),
			suggestedNewBufferCase(`let number = 1; new Buffer(number)`, `new Buffer(number)`),
			suggestedNewBufferCase(`var string = "x"; new Buffer(string)`, `new Buffer(string)`),
			fixedNewBufferCase(`new Buffer((1 === 1) ? 1 : value)`, `new Buffer((1 === 1) ? 1 : value)`, "alloc", `Buffer.alloc((1 === 1) ? 1 : value)`),
			fixedNewBufferCase(`new Buffer(undefined ?? 1)`, `new Buffer(undefined ?? 1)`, "alloc", `Buffer.alloc(undefined ?? 1)`),
			fixedNewBufferCase(`new Buffer({} && 1)`, `new Buffer({} && 1)`, "alloc", `Buffer.alloc({} && 1)`),
			// A known condition can select a numeric expression whose value is
			// unknown. Do not require the selected branch itself to fold.
			fixedNewBufferCase(`new Buffer(true ? Math.min(size, 8) : value)`, `new Buffer(true ? Math.min(size, 8) : value)`, "alloc", `Buffer.alloc(true ? Math.min(size, 8) : value)`),
			fixedNewBufferCase(`new Buffer(false ? value : bytes.length)`, `new Buffer(false ? value : bytes.length)`, "alloc", `Buffer.alloc(false ? value : bytes.length)`),
			fixedNewBufferCase(`const enabled = true; new Buffer(enabled ? Number(value) : unknown)`, `new Buffer(enabled ? Number(value) : unknown)`, "alloc", `const enabled = true; Buffer.alloc(enabled ? Number(value) : unknown)`),
			fixedNewBufferCase(`new Buffer("x" ? parseInt(value) : unknown)`, `new Buffer("x" ? parseInt(value) : unknown)`, "alloc", `Buffer.alloc("x" ? parseInt(value) : unknown)`),
			fixedNewBufferCase(`new Buffer(0 ? unknown : bytes.length)`, `new Buffer(0 ? unknown : bytes.length)`, "alloc", `Buffer.alloc(0 ? unknown : bytes.length)`),
			fixedNewBufferCase(`new Buffer(null ? unknown : bytes.length)`, `new Buffer(null ? unknown : bytes.length)`, "alloc", `Buffer.alloc(null ? unknown : bytes.length)`),
			fixedNewBufferCase(`new Buffer({} ? bytes.length : unknown)`, `new Buffer({} ? bytes.length : unknown)`, "alloc", `Buffer.alloc({} ? bytes.length : unknown)`),
			withFileName(fixedNewBufferCase(`function f(size: number) { new Buffer(true ? size : unknown) }`, `new Buffer(true ? size : unknown)`, "alloc", `function f(size: number) { Buffer.alloc(true ? size : unknown) }`), "file.ts"),
			fixedNewBufferCase(`new Buffer(enabled ? bytes.length : Number(value))`, `new Buffer(enabled ? bytes.length : Number(value))`, "alloc", `Buffer.alloc(enabled ? bytes.length : Number(value))`),
			suggestedNewBufferCase(`new Buffer(enabled ? bytes.length : unknown)`, `new Buffer(enabled ? bytes.length : unknown)`),
			suggestedNewBufferCase(`let enabled = true; new Buffer(enabled ? bytes.length : unknown)`, `new Buffer(enabled ? bytes.length : unknown)`),
			suggestedNewBufferCase(`new Buffer(false ? bytes.length : unknown)`, `new Buffer(false ? bytes.length : unknown)`),
			// The restricted operand can contain member access, calls, or other
			// expressions above the constructor in the AST.
			fixedNewBufferCase("function f() { return new // bytes\nBuffer([42]).length; }", "new // bytes\nBuffer([42])", "from", "function f() { return ( // bytes\nBuffer.from([42]).length); }"),
			fixedNewBufferCase("function f() { throw new // bytes\nBuffer([42]).toString(); }", "new // bytes\nBuffer([42])", "from", "function f() { throw ( // bytes\nBuffer.from([42]).toString()); }"),
			fixedNewBufferCase("function* f() { yield new // bytes\nBuffer([42]).length + 1; }", "new // bytes\nBuffer([42])", "from", "function* f() { yield ( // bytes\nBuffer.from([42]).length + 1); }"),
			withFileName(fixedNewBufferCase("function f() { return new // bytes\nBuffer([42])!; }", "new // bytes\nBuffer([42])", "from", "function f() { return ( // bytes\nBuffer.from([42])!); }"), "file.ts"),
			// A line terminator between `new` and the callee would turn this
			// into a bare yield if the fixer only removed `new`.
			fixedNewBufferCase("function* values() {\n\tyield new // yield\n\t\tBuffer(1);\n}", "new // yield\n\t\tBuffer(1)", "alloc", "function* values() {\n\tyield ( // yield\n\t\tBuffer.alloc(1));\n}"),
			// Delegating yields need the same operand grouping after `yield*`.
			fixedNewBufferCase("function* values() {\n\tyield* new // yield-star\n\t\tBuffer([1]);\n}", "new // yield-star\n\t\tBuffer([1])", "from", "function* values() {\n\tyield* ( // yield-star\n\t\tBuffer.from([1]));\n}"),
			// ECMAScript line and paragraph separators also trigger ASI.
			fixedNewBufferCase("() => {\n\treturn new\u2028\tBuffer(1);\n}", "new\u2028\tBuffer(1)", "alloc", "() => {\n\treturn ( Buffer.alloc(1));\n}"),
			fixedNewBufferCase("() => {\n\tthrow new\u2029\tBuffer(1);\n}", "new\u2029\tBuffer(1)", "alloc", "() => {\n\tthrow ( Buffer.alloc(1));\n}"),
			fixedNewBufferCase(`new Buffer((1))`, `new Buffer((1))`, "alloc", `Buffer.alloc((1))`),
			fixedNewBufferCase(`new Buffer(([1]))`, `new Buffer(([1]))`, "from", `Buffer.from(([1]))`),
			fixedNewBufferCase(`const Buffer = factory; new Buffer(1)`, `new Buffer(1)`, "alloc", `const Buffer = factory; Buffer.alloc(1)`),
			withFileName(suggestedNewBufferCase(`new Buffer((value) as string)`, `new Buffer((value) as string)`), "file.ts"),
			suggestedNewBufferCase(`new Buffer(Math.unknown())`, `new Buffer(Math.unknown())`),
			suggestedNewBufferCase(`new Buffer(foo?.length)`, `new Buffer(foo?.length)`),
			fixedNewBufferCase(`new Buffer(-1)`, `new Buffer(-1)`, "alloc", `Buffer.alloc(-1)`),
			fixedNewBufferCase(`new Buffer(1 + 2)`, `new Buffer(1 + 2)`, "alloc", `Buffer.alloc(1 + 2)`),
			fixedNewBufferCase(`new Buffer(Number.parseInt(value))`, `new Buffer(Number.parseInt(value))`, "alloc", `Buffer.alloc(Number.parseInt(value))`),
			withFileName(fixedNewBufferCase(`new Buffer(value as number)`, `new Buffer(value as number)`, "alloc", `Buffer.alloc(value as number)`), "file.ts"),
			withFileName(fixedNewBufferCase(`function f(value: number) { new Buffer(value) }`, `new Buffer(value)`, "alloc", `function f(value: number) { Buffer.alloc(value) }`), "file.ts"),
			fixedNewBufferCase(`new Buffer("a" + "b")`, `new Buffer("a" + "b")`, "from", `Buffer.from("a" + "b")`),
			fixedNewBufferCase(`new Buffer(true ? 1 : value)`, `new Buffer(true ? 1 : value)`, "alloc", `Buffer.alloc(true ? 1 : value)`),
			fixedNewBufferCase(`new Buffer(false ? value : 1)`, `new Buffer(false ? value : 1)`, "alloc", `Buffer.alloc(false ? value : 1)`),
			fixedNewBufferCase(`const enabled = true; new Buffer(enabled ? 1 : value)`, `new Buffer(enabled ? 1 : value)`, "alloc", `const enabled = true; Buffer.alloc(enabled ? 1 : value)`),
			fixedNewBufferCase(`const size = 1; new Buffer(size || value)`, `new Buffer(size || value)`, "alloc", `const size = 1; Buffer.alloc(size || value)`),
			fixedNewBufferCase(`let value; new Buffer(value = 1)`, `new Buffer(value = 1)`, "alloc", `let value; Buffer.alloc(value = 1)`),
			fixedNewBufferCase(`new Buffer("x".indexOf(value))`, `new Buffer("x".indexOf(value))`, "alloc", `Buffer.alloc("x".indexOf(value))`),
			fixedNewBufferCase(`function f(Math) { new Buffer(Math.min(x, y)) }`, `new Buffer(Math.min(x, y))`, "alloc", `function f(Math) { Buffer.alloc(Math.min(x, y)) }`),
			fixedNewBufferCase(`function f(Number) { new Buffer(Number(value)) }`, `new Buffer(Number(value))`, "alloc", `function f(Number) { Buffer.alloc(Number(value)) }`),
			fixedNewBufferCase(`new Buffer(1 || value)`, `new Buffer(1 || value)`, "alloc", `Buffer.alloc(1 || value)`),
			suggestedNewBufferCase(`let value = 1; new Buffer(value++)`, `new Buffer(value++)`),
		},
	)
}

func TestNoNewBufferUnsafeMemberReads(t *testing.T) {
	var invalid []rule_tester.InvalidTestCase
	for _, value := range []string{`[1]`, `1`, `"x"`} {
		setup := `const object = {value: ` + value + `}; Object.defineProperty(object, "value", {get() { return unknown; }}); `
		for _, test := range []struct {
			declarations string
			argument     string
		}{
			{argument: `object.value`},
			{argument: `object["value"]`},
			{declarations: `const alias = object.value; `, argument: `alias`},
			{declarations: `const first = object["value"]; const alias = first; `, argument: `alias`},
		} {
			expression := `new Buffer(` + test.argument + `)`
			invalid = append(invalid, suggestedNewBufferCase(setup+test.declarations+expression+`;`, expression))
		}
	}
	// An aliased getter must not select an array branch or hide inside an array initializer.
	setup := `const object = {value: true}; Object.defineProperty(object, "value", {get() { return unknown; }}); `
	invalid = append(invalid,
		suggestedNewBufferCase(setup+`const enabled = object.value; new Buffer(enabled ? [1] : 1);`, `new Buffer(enabled ? [1] : 1)`),
		suggestedNewBufferCase(setup+`const values = [object.value]; new Buffer(values);`, `new Buffer(values)`),
		fixedNewBufferCase(`const bytes = [1]; const alias = bytes; new Buffer(alias);`, `new Buffer(alias)`, "from", `const bytes = [1]; const alias = bytes; Buffer.from(alias);`),
		fixedNewBufferCase(`const size = 1; const alias = size; new Buffer(alias);`, `new Buffer(alias)`, "alloc", `const size = 1; const alias = size; Buffer.alloc(alias);`),
		fixedNewBufferCase(`const text = "x"; const alias = text; new Buffer(alias);`, `new Buffer(alias)`, "from", `const text = "x"; const alias = text; Buffer.from(alias);`),
	)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_buffer.NoNewBufferRule, nil, invalid)
}

func TestNoNewBufferEditDemand(t *testing.T) {
	const source = "new Buffer(unknown);\n"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}

	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     no_new_buffer.NoNewBufferRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return no_new_buffer.NoNewBufferRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				got = append(got, diagnostic)
			}},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
		}
		diagnostics[demand] = got[0]
	}

	base := diagnostics[rule.EditDemandNone]
	for demand, diagnostic := range diagnostics {
		if diagnostic.Range != base.Range || diagnostic.Message.Id != base.Message.Id || diagnostic.Message.Description != base.Message.Description || !reflect.DeepEqual(diagnostic.Message.Data, base.Message.Data) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d materialized an autofix for a suggestion-only rule", demand)
		}
	}
	if diagnostics[rule.EditDemandNone].Suggestions != nil || diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions materialized without suggestion demand")
	}
	suggestions := diagnostics[rule.EditDemandSuggestion].Suggestions
	allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
	if suggestions == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestions, *allSuggestions) {
		t.Fatal("suggestion artifacts differ between suggestion-only and all demand")
	}
	output, _, _ := linter.ApplyRuleFixes(source, []rule.RuleSuggestion{(*allSuggestions)[0]})
	if output != "Buffer.from(unknown);\n" {
		t.Fatalf("suggestion output = %q", output)
	}
}

// These expectations were verified against eslint-plugin-unicorn v74.0.0.
func TestNoNewBufferControlFlowSafety(t *testing.T) {
	invalid := []rule_tester.InvalidTestCase{
		// The member safety proof uses JavaScript code units for property keys.
		fixedNewBufferCase(`new Buffer(({"😀": "x"})["\ud83d\ude00"])`, `new Buffer(({"😀": "x"})["\ud83d\ude00"])`, "from", `Buffer.from(({"😀": "x"})["\ud83d\ude00"])`),
		fixedNewBufferCase(`new Buffer(({"\ud83d\ude00": [1]})["😀"])`, `new Buffer(({"\ud83d\ude00": [1]})["😀"])`, "from", `Buffer.from(({"\ud83d\ude00": [1]})["😀"])`),
		fixedNewBufferCase("new Buffer(({value: \"x\"}).value)", "new Buffer(({value: \"x\"}).value)", "from", "Buffer.from(({value: \"x\"}).value)"),
		fixedNewBufferCase("new Buffer(({value: \"x\"})[\"value\"])", "new Buffer(({value: \"x\"})[\"value\"])", "from", "Buffer.from(({value: \"x\"})[\"value\"])"),
		fixedNewBufferCase("new Buffer(({value: [1]}).value)", "new Buffer(({value: [1]}).value)", "from", "Buffer.from(({value: [1]}).value)"),
		fixedNewBufferCase("new Buffer(({value: 1, value: \"x\"}).value)", "new Buffer(({value: 1, value: \"x\"}).value)", "from", "Buffer.from(({value: 1, value: \"x\"}).value)"),
		fixedNewBufferCase("const text = \"x\"; new Buffer(({text}).text)", "new Buffer(({text}).text)", "from", "const text = \"x\"; Buffer.from(({text}).text)"),
		fixedNewBufferCase("new Buffer([1][0])", "new Buffer([1][0])", "alloc", "Buffer.alloc([1][0])"),
		fixedNewBufferCase("new Buffer([[1]][0])", "new Buffer([[1]][0])", "from", "Buffer.from([[1]][0])"),
		fixedNewBufferCase("new Buffer([1][\"length\"])", "new Buffer([1][\"length\"])", "alloc", "Buffer.alloc([1][\"length\"])"),
		fixedNewBufferCase("new Buffer(\"x\"[0])", "new Buffer(\"x\"[0])", "from", "Buffer.from(\"x\"[0])"),
		fixedNewBufferCase("new Buffer(\"x\"[\"length\"])", "new Buffer(\"x\"[\"length\"])", "alloc", "Buffer.alloc(\"x\"[\"length\"])"),
		fixedNewBufferCase("const text = \"x\"; new Buffer(text[0])", "new Buffer(text[0])", "from", "const text = \"x\"; Buffer.from(text[0])"),
		fixedNewBufferCase("const key = 0; new Buffer([1][key])", "new Buffer([1][key])", "alloc", "const key = 0; Buffer.alloc([1][key])"),
		fixedNewBufferCase("const key = \"value\"; new Buffer(({value: \"x\"})[key])", "new Buffer(({value: \"x\"})[key])", "from", "const key = \"value\"; Buffer.from(({value: \"x\"})[key])"),
		fixedNewBufferCase("new Buffer(Object.freeze([1]))", "new Buffer(Object.freeze([1]))", "from", "Buffer.from(Object.freeze([1]))"),
		fixedNewBufferCase("new Buffer(Object.seal([1]))", "new Buffer(Object.seal([1]))", "from", "Buffer.from(Object.seal([1]))"),
		fixedNewBufferCase("new Buffer(Object.preventExtensions([1]))", "new Buffer(Object.preventExtensions([1]))", "from", "Buffer.from(Object.preventExtensions([1]))"),
		fixedNewBufferCase("new Buffer(Object.freeze(Object.seal([1])))", "new Buffer(Object.freeze(Object.seal([1])))", "from", "Buffer.from(Object.freeze(Object.seal([1])))"),
		fixedNewBufferCase("new Buffer(Object.freeze(({value: [1]}).value))", "new Buffer(Object.freeze(({value: [1]}).value))", "from", "Buffer.from(Object.freeze(({value: [1]}).value))"),
		fixedNewBufferCase("const frozen = Object.freeze([1]); const alias = frozen; new Buffer(alias)", "new Buffer(alias)", "from", "const frozen = Object.freeze([1]); const alias = frozen; Buffer.from(alias)"),
		fixedNewBufferCase("new Buffer(Math[\"PI\"])", "new Buffer(Math[\"PI\"])", "alloc", "Buffer.alloc(Math[\"PI\"])"),
		fixedNewBufferCase("new Buffer(Math[\"SQRT2\"])", "new Buffer(Math[\"SQRT2\"])", "alloc", "Buffer.alloc(Math[\"SQRT2\"])"),
		fixedNewBufferCase("new Buffer(Number[\"EPSILON\"])", "new Buffer(Number[\"EPSILON\"])", "alloc", "Buffer.alloc(Number[\"EPSILON\"])"),
		fixedNewBufferCase("new Buffer(Number[\"MAX_SAFE_INTEGER\"])", "new Buffer(Number[\"MAX_SAFE_INTEGER\"])", "alloc", "Buffer.alloc(Number[\"MAX_SAFE_INTEGER\"])"),
		fixedNewBufferCase("new Buffer(Number[\"MIN_VALUE\"])", "new Buffer(Number[\"MIN_VALUE\"])", "alloc", "Buffer.alloc(Number[\"MIN_VALUE\"])"),
		fixedNewBufferCase("new Buffer(Number[\"NaN\"])", "new Buffer(Number[\"NaN\"])", "alloc", "Buffer.alloc(Number[\"NaN\"])"),
		fixedNewBufferCase("new Buffer(Math[\"PI\"] ? 1 : unknown)", "new Buffer(Math[\"PI\"] ? 1 : unknown)", "alloc", "Buffer.alloc(Math[\"PI\"] ? 1 : unknown)"),
		fixedNewBufferCase("new Buffer(String.raw`x`)", "new Buffer(String.raw`x`)", "from", "Buffer.from(String.raw`x`)"),
		fixedNewBufferCase("new Buffer(({value: true}).value ? [1] : unknown)", "new Buffer(({value: true}).value ? [1] : unknown)", "from", "Buffer.from(({value: true}).value ? [1] : unknown)"),
		fixedNewBufferCase("new Buffer([true][0] ? 1 : unknown)", "new Buffer([true][0] ? 1 : unknown)", "alloc", "Buffer.alloc([true][0] ? 1 : unknown)"),
		fixedNewBufferCase("let value = 1; new Buffer(true ? [1] : value)", "new Buffer(true ? [1] : value)", "from", "let value = 1; Buffer.from(true ? [1] : value)"),
		fixedNewBufferCase("let value = 1; new Buffer([1] || value)", "new Buffer([1] || value)", "from", "let value = 1; Buffer.from([1] || value)"),
		fixedNewBufferCase("let flag; new Buffer(flag = 1)", "new Buffer(flag = 1)", "alloc", "let flag; Buffer.alloc(flag = 1)"),
		fixedNewBufferCase("let flag; new Buffer((flag = true) ? 1 : 2)", "new Buffer((flag = true) ? 1 : 2)", "alloc", "let flag; Buffer.alloc((flag = true) ? 1 : 2)"),
		fixedNewBufferCase("let flag; new Buffer(true ? 1 : (flag = \"x\"))", "new Buffer(true ? 1 : (flag = \"x\"))", "alloc", "let flag; Buffer.alloc(true ? 1 : (flag = \"x\"))"),
		suggestedNewBufferCase("new Buffer(({value: \"x\"}).missing)", "new Buffer(({value: \"x\"}).missing)"),
		suggestedNewBufferCase("new Buffer([,][0])", "new Buffer([,][0])"),
		suggestedNewBufferCase("new Buffer([1][1])", "new Buffer([1][1])"),
		suggestedNewBufferCase("new Buffer([1][-1])", "new Buffer([1][-1])"),
		suggestedNewBufferCase("new Buffer([1][\"00\"])", "new Buffer([1][\"00\"])"),
		suggestedNewBufferCase("new Buffer(\"x\"[1])", "new Buffer(\"x\"[1])"),
		suggestedNewBufferCase("new Buffer(({get value() { return \"x\"; }}).value)", "new Buffer(({get value() { return \"x\"; }}).value)"),
		suggestedNewBufferCase("new Buffer(({[\"value\"]: \"x\"}).value)", "new Buffer(({[\"value\"]: \"x\"}).value)"),
		suggestedNewBufferCase("new Buffer(({__proto__: {value: \"x\"}}).value)", "new Buffer(({__proto__: {value: \"x\"}}).value)"),
		suggestedNewBufferCase("new Buffer(({value: \"x\", __proto__: null}).value)", "new Buffer(({value: \"x\", __proto__: null}).value)"),
		suggestedNewBufferCase("const object = {value: \"x\"}; new Buffer(object.value)", "new Buffer(object.value)"),
		suggestedNewBufferCase("const values = [1]; new Buffer(values[0])", "new Buffer(values[0])"),
		suggestedNewBufferCase("const text = \"x\"; const alias = text; new Buffer(alias[0])", "new Buffer(alias[0])"),
		suggestedNewBufferCase("new Buffer(Object[\"freeze\"]([1]))", "new Buffer(Object[\"freeze\"]([1]))"),
		suggestedNewBufferCase("new Buffer(Object.freeze([1], 2))", "new Buffer(Object.freeze([1], 2))"),
		suggestedNewBufferCase("new Buffer(Object.freeze(...[[1]]))", "new Buffer(Object.freeze(...[[1]]))"),
		suggestedNewBufferCase("new Buffer(Object?.freeze([1]))", "new Buffer(Object?.freeze([1]))"),
		suggestedNewBufferCase("new Buffer(Object.freeze?.([1]))", "new Buffer(Object.freeze?.([1]))"),
		suggestedNewBufferCase("function f(Object) { new Buffer(Object.freeze([1])) }", "new Buffer(Object.freeze([1]))"),
		suggestedNewBufferCase("function f(Math) { new Buffer(Math[\"PI\"]) }", "new Buffer(Math[\"PI\"])"),
		suggestedNewBufferCase("function f(Number) { new Buffer(Number[\"EPSILON\"]) }", "new Buffer(Number[\"EPSILON\"])"),
		suggestedNewBufferCase("const key = \"PI\"; new Buffer(Math[key])", "new Buffer(Math[key])"),
		suggestedNewBufferCase("new Buffer(Math[\"unknown\"])", "new Buffer(Math[\"unknown\"])"),
		suggestedNewBufferCase("new Buffer(Math?.[\"PI\"])", "new Buffer(Math?.[\"PI\"])"),
		suggestedNewBufferCase("let key = 0; new Buffer([1][key])", "new Buffer([1][key])"),
		suggestedNewBufferCase("const value = ({value: \"x\"}).value; new Buffer(value)", "new Buffer(value)"),
		suggestedNewBufferCase("let flag; new Buffer((flag = true) ? 1 : unknown)", "new Buffer((flag = true) ? 1 : unknown)"),
		suggestedNewBufferCase("let flag; const enabled = flag = true; new Buffer(enabled ? 1 : unknown)", "new Buffer(enabled ? 1 : unknown)"),
		suggestedNewBufferCase("let flag; const first = flag = true; const enabled = first; new Buffer(enabled ? [1] : unknown)", "new Buffer(enabled ? [1] : unknown)"),
		suggestedNewBufferCase("let flag = 0; new Buffer((flag += 1) ? 1 : unknown)", "new Buffer((flag += 1) ? 1 : unknown)"),
		suggestedNewBufferCase("let flag = false; new Buffer((flag ||= true) ? 1 : unknown)", "new Buffer((flag ||= true) ? 1 : unknown)"),
		suggestedNewBufferCase("let flag = 0; new Buffer(flag++ ? [1] : unknown)", "new Buffer(flag++ ? [1] : unknown)"),
		suggestedNewBufferCase("let flag = 0; new Buffer(++flag ? [1] : unknown)", "new Buffer(++flag ? [1] : unknown)"),
		suggestedNewBufferCase("new Buffer((void effect()) ?? [1])", "new Buffer((void effect()) ?? [1])"),
		suggestedNewBufferCase("new Buffer((effect(), true) ? [1] : unknown)", "new Buffer((effect(), true) ? [1] : unknown)"),
		suggestedNewBufferCase("new Buffer(String(\"x\"))", "new Buffer(String(\"x\"))"),
		suggestedNewBufferCase("const text = String(\"x\"); new Buffer(text)", "new Buffer(text)"),
		suggestedNewBufferCase("new Buffer(Object.freeze([1]) && [1])", "new Buffer(Object.freeze([1]) && [1])"),
		suggestedNewBufferCase("const enabled = Object.freeze([1]) && true; new Buffer(enabled ? [1] : unknown)", "new Buffer(enabled ? [1] : unknown)"),
		suggestedNewBufferCase("new Buffer(Object.freeze((effect(), [1])))", "new Buffer(Object.freeze((effect(), [1])))"),
		suggestedNewBufferCase("const object = {}; new Buffer((delete object.value) ? [1] : unknown)", "new Buffer((delete object.value) ? [1] : unknown)"),
		suggestedNewBufferCase("let value = 1; new Buffer((void value) ?? [1])", "new Buffer((void value) ?? [1])"),
		suggestedNewBufferCase("let flag; new Buffer(true ? \"x\" : (flag = \"x\"))", "new Buffer(true ? \"x\" : (flag = \"x\"))"),
		suggestedNewBufferCase("let flag; const sideEffect = flag = \"x\"; new Buffer(true ? \"x\" : sideEffect)", "new Buffer(true ? \"x\" : sideEffect)"),
		suggestedNewBufferCase("const cycle = cycle; new Buffer(cycle)", "new Buffer(cycle)"),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_buffer.NoNewBufferRule, nil, invalid)
}
