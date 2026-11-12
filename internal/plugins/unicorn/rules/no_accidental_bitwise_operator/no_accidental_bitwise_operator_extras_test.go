package no_accidental_bitwise_operator_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_accidental_bitwise_operator"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAccidentalBitwiseOperatorExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_accidental_bitwise_operator.NoAccidentalBitwiseOperatorRule, []rule_tester.ValidTestCase{
		jsCase("obj & obj?.[key]; obj & (obj?.p).q; obj & (obj.p)?.q;"),
		jsCase("value | tag`text`; value | <div />;"),
		{Code: `options | ({} satisfies Shape); obj & obj!.prop; obj! & obj.prop; obj & (obj as Shape).prop; type T = A & A["prop"];`, FileName: "case.ts", LanguageOptions: module},
	}, []rule_tester.InvalidTestCase{
		// The receiver may be parenthesized and the two identifiers can use different spellings.
		suggestionCase(`((\u006Fbj)) & (((obj)[key?.value]));`, `((\u006Fbj)) && (((obj)[key?.value]));`, "&", 1, 14),
		suggestionCase("class Example { #field; check(obj) { return obj & obj.#field; } }", "class Example { #field; check(obj) { return obj && obj.#field; } }", "&", 1, 49),
		// JSDoc casts in JavaScript leave the operand's runtime form unchanged.
		suggestionCase("/** @type {Object} */ (obj) & (/** @type {Object} */ (obj)).prop;", "/** @type {Object} */ (obj) && (/** @type {Object} */ (obj)).prop;", "&", 1, 29),
		suggestionCase("value | /** @satisfies {Object} */ ({});", "value || /** @satisfies {Object} */ ({});", "|", 1, 7),
		suggestionCase("value | `text ${name}`;", "value || `text ${name}`;", "|", 1, 7),
		suggestionCase("value |= () => {};", "value ||= () => {};", "|=", 1, 7),
		suggestionCase("value | {} | fallback;", "(value || {}) | fallback;", "|", 1, 7),
		withFileName(suggestionCase("const element = <div name={ready && value | ''} />;", "const element = <div name={ready && (value || '')} />;", "|", 1, 43), "case.tsx"),
		withFileName(suggestionCase("new Foo<A<B>>\nvalue | '' ?? fallback", "new Foo<A<B>>\n;(value || '') ?? fallback", "|", 2, 7), "case.ts"),
		suggestionCase("foo()\n/* leading */ value | '' // trailing\n  ?? fallback;", "foo()\n/* leading */ ;(value || '') // trailing\n  ?? fallback;", "|", 2, 21),
		{
			Code: "result = value | {};\na, value | '';", FileName: "case.js", LanguageOptions: module,
			Errors: []rule_tester.InvalidTestCaseError{
				operatorError("|", 1, 16, "result = value || {};\na, value | '';"),
				operatorError("|", 2, 10, "result = value | {};\na, value || '';"),
			},
		},
		// Report only the operator, including after UTF-16 surrogate pairs and CRLF.
		{
			Code: "const text = \"😀\"; obj /* & */ & obj.p;\r\ninput\n/* |= */ |= false;", FileName: "case.js", LanguageOptions: module,
			Errors: []rule_tester.InvalidTestCaseError{
				operatorError("&", 1, 32, "const text = \"😀\"; obj /* & */ && obj.p;\r\ninput\n/* |= */ |= false;"),
				operatorError("|=", 3, 10, "const text = \"😀\"; obj /* & */ & obj.p;\r\ninput\n/* |= */ ||= false;"),
			},
		},
		{
			Code: "obj & obj.p | {};", FileName: "case.js", LanguageOptions: module,
			Errors: []rule_tester.InvalidTestCaseError{
				// The outer expression is visited first.
				operatorError("|", 1, 13, "obj & obj.p || {};"),
				operatorError("&", 1, 5, "(obj && obj.p) | {};"),
			},
		},
	})
}

func TestNoAccidentalBitwiseOperatorEditDemand(t *testing.T) {
	const code = "ready && value | '';"
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/case.js", Path: "/case.js"}, code, core.ScriptKindJS)
	node := source.Statements.Nodes[0].Expression().AsBinaryExpression().Right
	definition := no_accidental_bitwise_operator.NoAccidentalBitwiseOperatorRule
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var reported []rule.RuleDiagnostic
		ctx := (rule.RuleContext{SourceFile: source}).WithDiagnosticConsumer(definition.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(d rule.RuleDiagnostic) { reported = append(reported, d) },
		})
		definition.Run(ctx, nil)[ast.KindBinaryExpression](node)
		if len(reported) != 1 {
			t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(reported))
		}
		diagnostics[demand] = reported[0]
	}

	baseline := diagnostics[rule.EditDemandNone]
	all := diagnostics[rule.EditDemandAll]
	for demand, diagnostic := range diagnostics {
		if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) {
			t.Errorf("demand %d changed the diagnostic range or message", demand)
		}
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d produced an autofix", demand)
		}
		if demand == rule.EditDemandNone || demand == rule.EditDemandAutofix {
			if diagnostic.Suggestions != nil {
				t.Errorf("demand %d produced a suggestion", demand)
			}
			continue
		}
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 || !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
			t.Fatalf("demand %d produced inconsistent suggestions", demand)
		}
		message := (*diagnostic.Suggestions)[0].Message
		if message.Id != "no-accidental-bitwise-operator/suggestion" || message.Description != "Replace `|` with `||`." {
			t.Errorf("demand %d produced an unexpected suggestion message: %v", demand, message)
		}
	}
}
