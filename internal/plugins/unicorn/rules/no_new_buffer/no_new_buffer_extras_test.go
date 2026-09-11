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
