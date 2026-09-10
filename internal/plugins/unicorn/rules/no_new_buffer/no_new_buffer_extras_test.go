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
			// A line terminator between `new` and the callee would turn this
			// into a bare yield if the fixer only removed `new`.
			fixedNewBufferCase("function* values() {\n\tyield new // yield\n\t\tBuffer(1);\n}", "new // yield\n\t\tBuffer(1)", "alloc", "function* values() {\n\tyield ( // yield\n\t\tBuffer.alloc(1));\n}"),
			// Delegating yields need the same operand grouping after `yield*`.
			fixedNewBufferCase("function* values() {\n\tyield* new // yield-star\n\t\tBuffer([1]);\n}", "new // yield-star\n\t\tBuffer([1])", "from", "function* values() {\n\tyield* ( // yield-star\n\t\tBuffer.from([1]));\n}"),
			// ECMAScript line and paragraph separators also trigger ASI.
			fixedNewBufferCase("() => {\n\treturn new\u2028\tBuffer(1);\n}", "new\u2028\tBuffer(1)", "alloc", "() => {\n\treturn ( \u2028\tBuffer.alloc(1));\n}"),
			fixedNewBufferCase("() => {\n\tthrow new\u2029\tBuffer(1);\n}", "new\u2029\tBuffer(1)", "alloc", "() => {\n\tthrow ( \u2029\tBuffer.alloc(1));\n}"),
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
