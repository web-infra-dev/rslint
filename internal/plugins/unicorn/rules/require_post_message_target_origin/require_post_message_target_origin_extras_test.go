package require_post_message_target_origin_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_post_message_target_origin"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequirePostMessageTargetOriginExtras(t *testing.T) {
	invalid := []rule_tester.InvalidTestCase{
		invalidCase("(window).postMessage(message)", "file.js", 1, 29, 1, 30,
			"(window).postMessage(message, window.location.origin)",
			"(window).postMessage(message, '*')",
		),
		invalidCase("(window.postMessage)(message)", "file.js", 1, 29, 1, 30,
			"(window.postMessage)(message, window.location.origin)",
			"(window.postMessage)(message, '*')",
		),
		invalidCase("window?.child.postMessage(message)", "file.js", 1, 34, 1, 35,
			"window?.child.postMessage(message, self.location.origin)",
			"window?.child.postMessage(message, '*')",
		),
		invalidCase("(window?.child).postMessage(message)", "file.js", 1, 36, 1, 37,
			"(window?.child).postMessage(message, self.location.origin)",
			"(window?.child).postMessage(message, '*')",
		),
		invalidCase("window['child'].postMessage(message)", "file.js", 1, 36, 1, 37,
			"window['child'].postMessage(message, self.location.origin)",
			"window['child'].postMessage(message, '*')",
		),
		invalidCase("(window as Window).postMessage<string>(message)", "file.ts", 1, 47, 1, 48,
			"(window as Window).postMessage<string>(message, self.location.origin)",
			"(window as Window).postMessage<string>(message, '*')",
		),
		invalidCase("window.postMessage<string>(message)", "file.ts", 1, 35, 1, 36,
			"window.postMessage<string>(message, window.location.origin)",
			"window.postMessage<string>(message, '*')",
		),
		invalidCase("(/** @type {Window} */ (window)).postMessage(message)", "file.js", 1, 53, 1, 54,
			"(/** @type {Window} */ (window)).postMessage(message, window.location.origin)",
			"(/** @type {Window} */ (window)).postMessage(message, '*')",
		),
		invalidCase("wind\\u006fw.postMessage(message)", "file.js", 1, 32, 1, 33,
			"wind\\u006fw.postMessage(message, window.location.origin)",
			"wind\\u006fw.postMessage(message, '*')",
		),
		invalidCase("window.postMessage(\"😀\");", "file.js", 1, 24, 1, 25,
			"window.postMessage(\"😀\", window.location.origin);",
			"window.postMessage(\"😀\", '*');",
		),
		invalidCase("window.postMessage(\n  message, // keep\n)", "file.js", 2, 11, 3, 2,
			"window.postMessage(\n  message, // keep\n window.location.origin,)",
			"window.postMessage(\n  message, // keep\n '*',)",
		),
		invalidCase("window.postMessage((message) /* keep */ )", "file.js", 1, 29, 1, 42,
			"window.postMessage((message) /* keep */ , window.location.origin)",
			"window.postMessage((message) /* keep */ , '*')",
		),
		invalidCase("window.postMessage(/[,)]/)", "file.js", 1, 26, 1, 27,
			"window.postMessage(/[,)]/, window.location.origin)",
			"window.postMessage(/[,)]/, '*')",
		),
		invalidCase("window.postMessage(`value: ${value}`)", "file.js", 1, 37, 1, 38,
			"window.postMessage(`value: ${value}`, window.location.origin)",
			"window.postMessage(`value: ${value}`, '*')",
		),
		invalidCase("window.postMessage(<span>{message}</span>)", "file.jsx", 1, 42, 1, 43,
			"window.postMessage(<span>{message}</span>, window.location.origin)",
			"window.postMessage(<span>{message}</span>, '*')",
		),
		invalidCase("class Sender extends Base { send() { super.postMessage(message); } }", "file.js", 1, 63, 1, 64,
			"class Sender extends Base { send() { super.postMessage(message, self.location.origin); } }",
			"class Sender extends Base { send() { super.postMessage(message, '*'); } }",
		),
		invalidCase("class Child extends window.postMessage(message) {}", "file.ts", 1, 47, 1, 48,
			"class Child extends window.postMessage(message, window.location.origin) {}",
			"class Child extends window.postMessage(message, '*') {}",
		),
		invalidCase("new Worker('worker.js').postMessage(message)", "file.js", 1, 44, 1, 45,
			"new Worker('worker.js').postMessage(message, self.location.origin)",
			"new Worker('worker.js').postMessage(message, '*')",
		),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&require_post_message_target_origin.RequirePostMessageTargetOriginRule,
		[]rule_tester.ValidTestCase{
			{Code: "window?.postMessage?.(message)", FileName: "file.js"},
			{Code: "(window?.postMessage)(message)", FileName: "file.js"},
			{Code: "window.postMessage(...[message])", FileName: "file.js"},
			{Code: "new window.postMessage(message)", FileName: "file.js"},
			{Code: "window['postMessage']?.(message)", FileName: "file.js"},
			{Code: "class Sender { #postMessage() {} send() { this.#postMessage(message); } }", FileName: "file.js"},
			{Code: "(window.postMessage as Function)(message)", FileName: "file.ts"},
			{Code: "window.postMessage!(message)", FileName: "file.ts"},
			{Code: "window.postMessage(message, undefined)", FileName: "file.js"},
		}, invalid)
}

func TestRequirePostMessageTargetOriginEditDemand(t *testing.T) {
	const source = "foo.postMessage(message)"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "edit-demand.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: require_post_message_target_origin.RequirePostMessageTargetOriginRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return require_post_message_target_origin.RequirePostMessageTargetOriginRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
		}
		diagnostics[demand] = got[0]
		output, _, fixed := linter.ApplyRuleFixes(source, got)
		if fixed || output != source {
			t.Fatalf("demand %d: unexpected autofix %q", demand, output)
		}
	}
	base := diagnostics[rule.EditDemandNone]
	for demand, d := range diagnostics {
		if d.Range != base.Range || !reflect.DeepEqual(d.Message, base.Message) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
		if d.FixesPtr != nil {
			t.Errorf("demand %d produced an autofix", demand)
		}
	}
	if diagnostics[rule.EditDemandNone].Suggestions != nil || diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions produced without demand")
	}
	suggestions := diagnostics[rule.EditDemandAll].Suggestions
	if suggestions == nil || len(*suggestions) != 3 || !reflect.DeepEqual(suggestions, diagnostics[rule.EditDemandSuggestion].Suggestions) {
		t.Fatal("inconsistent suggestion artifacts")
	}
	for i, origin := range []string{"foo.location.origin", "self.location.origin", "'*'"} {
		suggestion := (*suggestions)[i]
		if suggestion.Message.Id != "suggestion" || suggestion.Message.Description != "Use `"+origin+"`." {
			t.Errorf("suggestion %d: unexpected message %+v", i, suggestion.Message)
		}
		output, _, fixed := linter.ApplyRuleFixes(source, []rule.RuleSuggestion{suggestion})
		if !fixed || output != "foo.postMessage(message, "+origin+")" {
			t.Errorf("suggestion %d: unexpected output %q", i, output)
		}
	}
}
