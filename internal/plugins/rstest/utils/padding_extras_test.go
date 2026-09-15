package utils_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_after_all_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_after_each_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_all"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_before_all_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_before_each_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_describe_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_expect_groups"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/padding_around_test_blocks"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func paddingErrors(count int) []rule_tester.InvalidTestCaseError {
	errors := make([]rule_tester.InvalidTestCaseError, count)
	for i := range errors {
		errors[i].MessageId = "missingPadding"
	}
	return errors
}

func TestRstestPaddingRuleMappings(t *testing.T) {
	tests := []struct {
		name   string
		lint   *rule.Rule
		code   string
		output string
		errors int
	}{
		{name: "after-all", lint: &padding_around_after_all_blocks.PaddingAroundAfterAllBlocksRule, code: "setup();\nafterAll(cleanup);\nfinish();", output: "setup();\n\nafterAll(cleanup);\n\nfinish();", errors: 2},
		{name: "after-each", lint: &padding_around_after_each_blocks.PaddingAroundAfterEachBlocksRule, code: "setup();\nafterEach(cleanup);\nfinish();", output: "setup();\n\nafterEach(cleanup);\n\nfinish();", errors: 2},
		{name: "before-all", lint: &padding_around_before_all_blocks.PaddingAroundBeforeAllBlocksRule, code: "setup();\nbeforeAll(connect);\nfinish();", output: "setup();\n\nbeforeAll(connect);\n\nfinish();", errors: 2},
		{name: "before-each", lint: &padding_around_before_each_blocks.PaddingAroundBeforeEachBlocksRule, code: "setup();\nbeforeEach(reset);\nfinish();", output: "setup();\n\nbeforeEach(reset);\n\nfinish();", errors: 2},
		{name: "describe", lint: &padding_around_describe_blocks.PaddingAroundDescribeBlocksRule, code: "setup();\ndescribe('suite', run);\nfinish();", output: "setup();\n\ndescribe('suite', run);\n\nfinish();", errors: 2},
		{name: "expect-groups", lint: &padding_around_expect_groups.PaddingAroundExpectGroupsRule, code: "const value = load();\nexpect(value).toBe(true);\nexpect(value).toBeDefined();\nfinish();", output: "const value = load();\n\nexpect(value).toBe(true);\nexpect(value).toBeDefined();\n\nfinish();", errors: 2},
		{name: "test", lint: &padding_around_test_blocks.PaddingAroundTestBlocksRule, code: "setup();\ntest('works', run);\nit('also works', run);", output: "setup();\n\ntest('works', run);\n\nit('also works', run);", errors: 2},
		{name: "all", lint: &padding_around_all.PaddingAroundAllRule, code: "setup();\nbeforeAll(connect);\ntest('works', run);", output: "setup();\n\nbeforeAll(connect);\n\ntest('works', run);", errors: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule_tester.RunRuleTester(
				fixtures.GetRootDir(),
				"tsconfig.json",
				t,
				test.lint,
				[]rule_tester.ValidTestCase{{Code: test.output}},
				[]rule_tester.InvalidTestCase{{
					Code: test.code, Output: []string{test.output}, Errors: paddingErrors(test.errors),
				}},
			)
		})
	}
}

func TestRstestPaddingUsesRstestApiTable(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&padding_around_all.PaddingAroundAllRule,
		[]rule_tester.ValidTestCase{
			{Code: "setup();\nfit('Jest only', run);\nfinish();"},
			{Code: `setup();
xit('Jest only', run);
finish();`},
			{Code: "setup();\nxdescribe('Jest only', run);\nfinish();"},
			{Code: "setup();\nexpectTypeOf(value).toEqualTypeOf<string>();\nfinish();"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "setup();\nit('Rstest alias', run);\nfinish();",
				Output: []string{"setup();\n\nit('Rstest alias', run);\n\nfinish();"},
				Errors: paddingErrors(2),
			},
		},
	)
}

func TestRstestAtomicPaddingRuleOwnsAggregateBoundary(t *testing.T) {
	const code = "setup();\nbeforeAll(connect);"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	disableManager := rule.NewDisableManager(sourceFile, comments)
	cache := rule.NewFileCache()
	var diagnostics []rule.RuleDiagnostic
	var listenersForTraversal []rule.RuleListeners

	for _, lintRule := range []*rule.Rule{
		&padding_around_all.PaddingAroundAllRule,
		&padding_around_before_all_blocks.PaddingAroundBeforeAllBlocksRule,
	} {
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: disableManager,
		}.WithFileCache(cache).WithDiagnosticConsumer(
			lintRule.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		)
		listenersForTraversal = append(listenersForTraversal, lintRule.Run(ctx, nil))
	}

	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		for _, listeners := range listenersForTraversal {
			if listener := listeners[node.Kind]; listener != nil {
				listener(node)
			}
		}
		return node.ForEachChild(visit)
	}
	visit(sourceFile.AsNode())
	for _, listeners := range listenersForTraversal {
		if listener := listeners[rule.ListenerOnExit(ast.KindEndOfFile)]; listener != nil {
			listener(nil)
		}
	}

	if len(diagnostics) != 1 || diagnostics[0].RuleName != "rstest/padding-around-before-all-blocks" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}
