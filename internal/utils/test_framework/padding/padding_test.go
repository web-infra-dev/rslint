package padding_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/test_framework/padding"
)

var names = padding.StatementNames{
	"beforeAll": padding.StatementBeforeAll,
	"describe":  padding.StatementDescribe,
	"expect":    padding.StatementExpect,
	"test":      padding.StatementTest,
}

func around(statementType padding.StatementType) []padding.Config {
	return []padding.Config{
		{Padding: padding.PaddingAlways, Previous: padding.Types(padding.StatementAny), Next: padding.Types(statementType)},
		{Padding: padding.PaddingAlways, Previous: padding.Types(statementType), Next: padding.Types(padding.StatementAny)},
	}
}

func expectGroups() []padding.Config {
	return append(around(padding.StatementExpect), padding.Config{
		Padding:  padding.PaddingAny,
		Previous: padding.Types(padding.StatementExpect),
		Next:     padding.Types(padding.StatementExpect),
	})
}

func runRule(
	t *testing.T,
	code string,
	statementNames padding.StatementNames,
	configs []padding.Config,
	demand rule.EditDemand,
) ([]rule.RuleDiagnostic, string) {
	t.Helper()
	return runRules(t, code, statementNames, [][]padding.Config{configs}, demand)
}

func runRules(
	t *testing.T,
	code string,
	statementNames padding.StatementNames,
	configSets [][]padding.Config,
	demand rule.EditDemand,
) ([]rule.RuleDiagnostic, string) {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer("test/padding", rule.SeverityError, rule.DiagnosticConsumer{
		Demand: demand,
		Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	cache := rule.NewFileCache()
	var eofListeners []func(*ast.Node)
	var listenersForTraversal []rule.RuleListeners
	for _, configs := range configSets {
		testRule := padding.NewRule(padding.Definition{
			Name: "test/padding", Family: "test", Names: statementNames, Configs: configs,
		})
		listeners := testRule.Run(ctx.WithFileCache(cache), nil)
		listenersForTraversal = append(listenersForTraversal, listeners)
		if listener := listeners[rule.ListenerOnExit(ast.KindEndOfFile)]; listener != nil {
			eofListeners = append(eofListeners, listener)
		}
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
	for _, listener := range eofListeners {
		listener(nil)
	}
	output, _, _ := linter.ApplyRuleFixes(code, diagnostics)
	return diagnostics, output
}

func runNamedRules(
	t *testing.T,
	code string,
	rules []struct {
		name    string
		configs []padding.Config
	},
) []rule.RuleDiagnostic {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	disableManager := rule.NewDisableManager(sourceFile, comments)
	cache := rule.NewFileCache()
	var diagnostics []rule.RuleDiagnostic
	var eof func(*ast.Node)
	var listenersForTraversal []rule.RuleListeners
	for _, configured := range rules {
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: disableManager,
		}.WithFileCache(cache).WithDiagnosticConsumer(
			configured.name,
			rule.SeverityError,
			rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		)
		listeners := padding.NewRule(padding.Definition{
			Name: configured.name, Family: "test", Names: names, Configs: configured.configs,
		}).Run(ctx, nil)
		listenersForTraversal = append(listenersForTraversal, listeners)
		if listener := listeners[rule.ListenerOnExit(ast.KindEndOfFile)]; listener != nil {
			eof = listener
		}
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
	if eof != nil {
		eof(nil)
	}
	return diagnostics
}

func TestPaddingAroundStatement(t *testing.T) {
	code := "const database = createDatabase();\nbeforeAll(connect);\ntest('loads', run);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "const database = createDatabase();\n\nbeforeAll(connect);\n\ntest('loads', run);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingExistingBlankLines(t *testing.T) {
	code := "setup();\n\nbeforeAll(connect);\n\ntest('loads', run);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 0 || output != code {
		t.Fatalf("diagnostics = %d, output = %q", len(diagnostics), output)
	}
}

func TestPaddingExpectGroupsUseLastMatchingConfig(t *testing.T) {
	code := "const user = getUser();\nexpect(user.name).toBe('Ada');\nexpect(user.active).toBe(true);\nconst account = getAccount();"
	diagnostics, output := runRule(t, code, names, expectGroups(), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "const user = getUser();\n\nexpect(user.name).toBe('Ada');\nexpect(user.active).toBe(true);\n\nconst account = getAccount();"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingStatementListsAreIndependent(t *testing.T) {
	code := "setup();\nfunction run() {\n  prepare();\n  beforeAll(connect);\n  cleanup();\n}\ntest('loads', run);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "setup();\nfunction run() {\n  prepare();\n\n  beforeAll(connect);\n\n  cleanup();\n}\ntest('loads', run);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingIgnoresTypeScriptModuleBlocks(t *testing.T) {
	code := "namespace Feature {\n  setup();\n  beforeAll(connect);\n  cleanup();\n}"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 0 || output != code {
		t.Fatalf("diagnostics = %d, output = %q", len(diagnostics), output)
	}
}

func TestPaddingIgnoresClassStaticBlocks(t *testing.T) {
	code := "class Feature {\n  static {\n    setup();\n    beforeAll(connect);\n    cleanup();\n  }\n}"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 0 || output != code {
		t.Fatalf("diagnostics = %d, output = %q", len(diagnostics), output)
	}
}

func TestPaddingSwitchClauseStatementList(t *testing.T) {
	code := "switch (kind) {\ncase 'a':\n  setup();\n  beforeAll(connect);\n  break;\ndefault:\n  cleanup();\n}"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "switch (kind) {\ncase 'a':\n  setup();\n\n  beforeAll(connect);\n\n  break;\ndefault:\n  cleanup();\n}"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingLabelsAndAwait(t *testing.T) {
	code := "async function run() {\n  setup();\n  label: beforeAll(connect);\n  await expect(load()).resolves.toBe(true);\n  cleanup();\n}"
	configs := append(around(padding.StatementBeforeAll), around(padding.StatementExpect)...)
	diagnostics, output := runRule(t, code, names, configs, rule.EditDemandAutofix)
	if len(diagnostics) != 3 {
		t.Fatalf("diagnostics = %d, want 3", len(diagnostics))
	}
	want := "async function run() {\n  setup();\n\n  label: beforeAll(connect);\n\n  await expect(load()).resolves.toBe(true);\n\n  cleanup();\n}"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingAwaitParenthesizedSubject(t *testing.T) {
	code := "async function run() {\n  setup();\n  await (expect(load()).resolves.toBe(true));\n  cleanup();\n}"
	diagnostics, output := runRule(t, code, names, around(padding.StatementExpect), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "async function run() {\n  setup();\n\n  await (expect(load()).resolves.toBe(true));\n\n  cleanup();\n}"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingUnwrapsSourceParentheses(t *testing.T) {
	code := "setup();\n(expect(value).toBe(true));\ncleanup();"
	diagnostics, output := runRule(t, code, names, around(padding.StatementExpect), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "setup();\n\n(expect(value).toBe(true));\n\ncleanup();"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingComments(t *testing.T) {
	code := "setup(); // setup\n// hook docs\nbeforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "setup(); // setup\n\n// hook docs\nbeforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingBlockCommentOnSameLine(t *testing.T) {
	code := "setup(); /* setup */ beforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "setup(); /* setup */\n\n beforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingExistingBlankLineBeforeComment(t *testing.T) {
	code := "setup();\n\n// hook docs\nbeforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 0 || output != code {
		t.Fatalf("diagnostics = %d, output = %q", len(diagnostics), output)
	}
}

func TestPaddingMatchesUpstreamLFInsertionsInCRLFSource(t *testing.T) {
	code := "setup();\r\nbeforeAll(connect);\r\ncleanup();"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "setup();\n\r\nbeforeAll(connect);\n\r\ncleanup();"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingSameLineStatements(t *testing.T) {
	code := "setup(); beforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "setup();\n\n beforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingSemicolonStatement(t *testing.T) {
	code := "setup()\n;beforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "setup()\n\n;beforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingReportsInSourceOrder(t *testing.T) {
	code := "function run() {\n  setup();\n  beforeAll(connect);\n}\nsetup();\nbeforeAll(connect);"
	diagnostics, _ := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandNone)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	if diagnostics[0].Range.Pos() >= diagnostics[1].Range.Pos() {
		t.Fatalf("diagnostics are not in source order: %v then %v", diagnostics[0].Range, diagnostics[1].Range)
	}
}

func TestPaddingFrameworkNamesStaySeparate(t *testing.T) {
	code := "setup();\nfit('focused', run);\nbeforeAll(connect);"
	diagnostics, output := runRules(t, code, names, [][]padding.Config{
		around(padding.StatementTest),
		around(padding.StatementBeforeAll),
	}, rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "setup();\nfit('focused', run);\n\nbeforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingRulesDeduplicateMatchingPairs(t *testing.T) {
	code := "setup();\nbeforeAll(connect);"
	diagnostics, output := runRules(t, code, names, [][]padding.Config{
		around(padding.StatementBeforeAll),
		around(padding.StatementBeforeAll),
	}, rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "setup();\n\nbeforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingAtomicRuleWinsOverAggregateRegardlessOfRegistrationOrder(t *testing.T) {
	const code = "setup();\nbeforeAll(connect);"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	cache := rule.NewFileCache()
	var diagnostics []rule.RuleDiagnostic
	definitions := []padding.Definition{
		{Name: "test/padding-around-all", Family: "test", Priority: 100, Names: names, Configs: around(padding.StatementBeforeAll)},
		{Name: "test/padding-around-before-all", Family: "test", Names: names, Configs: around(padding.StatementBeforeAll)},
	}
	var listenersForTraversal []rule.RuleListeners
	var eof func(*ast.Node)
	for _, definition := range definitions {
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithFileCache(cache).WithDiagnosticConsumer(
			definition.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		)
		listeners := padding.NewRule(definition).Run(ctx, nil)
		listenersForTraversal = append(listenersForTraversal, listeners)
		if listener := listeners[rule.ListenerOnExit(ast.KindEndOfFile)]; listener != nil {
			eof = listener
		}
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
	eof(nil)
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "test/padding-around-before-all" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestPaddingFamiliesReportIndependently(t *testing.T) {
	const code = "setup();\nbeforeAll(connect);"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	cache := rule.NewFileCache()
	var diagnostics []rule.RuleDiagnostic
	var listenersForTraversal []rule.RuleListeners
	var eof func(*ast.Node)
	for _, definition := range []padding.Definition{
		{Name: "jest/padding", Family: "jest", Names: names, Configs: around(padding.StatementBeforeAll)},
		{Name: "rstest/padding", Family: "rstest", Names: names, Configs: around(padding.StatementBeforeAll)},
	} {
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithFileCache(cache).WithDiagnosticConsumer(
			definition.Name,
			rule.SeverityError,
			rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		)
		listeners := padding.NewRule(definition).Run(ctx, nil)
		listenersForTraversal = append(listenersForTraversal, listeners)
		if listener := listeners[rule.ListenerOnExit(ast.KindEndOfFile)]; listener != nil {
			eof = listener
		}
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
	eof(nil)
	if len(diagnostics) != 2 || diagnostics[0].RuleName != "jest/padding" ||
		diagnostics[1].RuleName != "rstest/padding" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestPaddingDisabledRuleFallsThroughToNextMatchingRule(t *testing.T) {
	code := "setup();\n// rslint-disable-next-line first/padding\nbeforeAll(connect);\ncleanup();"
	diagnostics := runNamedRules(t, code, []struct {
		name    string
		configs []padding.Config
	}{
		{name: "first/padding", configs: around(padding.StatementBeforeAll)},
		{name: "second/padding", configs: around(padding.StatementBeforeAll)},
	})
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	if diagnostics[0].RuleName != "second/padding" || diagnostics[1].RuleName != "first/padding" {
		t.Fatalf("rule names = %q, %q", diagnostics[0].RuleName, diagnostics[1].RuleName)
	}
}

func TestPaddingUsesSyntaxNamesWithoutBindingResolution(t *testing.T) {
	code := "function beforeAll(callback) { callback(); }\nsetup();\nbeforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandAutofix)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	want := "function beforeAll(callback) { callback(); }\nsetup();\n\nbeforeAll(connect);"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingDecodesEscapedIdentifierToken(t *testing.T) {
	code := "setup();\nt\\u0065st('works', run);\ncleanup();"
	diagnostics, output := runRule(t, code, names, around(padding.StatementTest), rule.EditDemandAutofix)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(diagnostics))
	}
	want := "setup();\n\nt\\u0065st('works', run);\n\ncleanup();"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestPaddingEmptyConfigurationDoesNotRegisterListener(t *testing.T) {
	testRule := padding.NewRule(padding.Definition{Name: "test/padding", Family: "test", Names: names})
	if listeners := testRule.Run(rule.RuleContext{}, nil); listeners != nil {
		t.Fatalf("listeners = %#v, want nil", listeners)
	}
}

func TestPaddingDiagnosticIdentityAndRange(t *testing.T) {
	code := "setup();\nbeforeAll(connect);"
	diagnostics, _ := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandNone)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	diagnostic := diagnostics[0]
	if diagnostic.Message.Id != "missingPadding" || diagnostic.Message.Description != "Expected blank line before this statement." {
		t.Fatalf("message = %#v", diagnostic.Message)
	}
	if got := code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "beforeAll(connect);" {
		t.Fatalf("reported text = %q", got)
	}
}

func TestPaddingUsesDefinitionMessage(t *testing.T) {
	const code = "setup();\nbeforeAll(connect);"
	customMessage := rule.RuleMessage{Id: "custom", Description: "Custom padding message."}
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithFileCache(rule.NewFileCache()).WithDiagnosticConsumer(
		"test/padding",
		rule.SeverityError,
		rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	)
	listeners := padding.NewRule(padding.Definition{
		Name: "test/padding", Family: "test", Message: customMessage, Names: names,
		Configs: around(padding.StatementBeforeAll),
	}).Run(ctx, nil)
	listeners[rule.ListenerOnExit(ast.KindEndOfFile)](nil)
	if len(diagnostics) != 1 || diagnostics[0].Message.Id != customMessage.Id ||
		diagnostics[0].Message.Description != customMessage.Description {
		t.Fatalf("diagnostics = %#v, want message %#v", diagnostics, customMessage)
	}
}

func TestPaddingAutofixIsDeferred(t *testing.T) {
	code := "setup();\nbeforeAll(connect);"
	diagnostics, output := runRule(t, code, names, around(padding.StatementBeforeAll), rule.EditDemandNone)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(diagnostics))
	}
	if diagnostics[0].FixesPtr != nil || output != code {
		t.Fatalf("unexpected deferred fix: fixes = %#v, output = %q", diagnostics[0].FixesPtr, output)
	}
}

func TestPaddingCopiesInputs(t *testing.T) {
	mutableNames := padding.StatementNames{"beforeAll": padding.StatementBeforeAll}
	mutableConfigs := around(padding.StatementBeforeAll)
	testRule := padding.NewRule(padding.Definition{
		Name: "test/padding", Family: "test", Names: mutableNames, Configs: mutableConfigs,
	})
	mutableNames["beforeAll"] = padding.StatementTest
	mutableConfigs[0].Next[0] = padding.StatementTest

	code := "setup();\nbeforeAll(connect);"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	comments := rule.NewCommentStore(sourceFile)
	var diagnostics []rule.RuleDiagnostic
	ctx := rule.RuleContext{
		SourceFile:     sourceFile,
		Comments:       comments,
		DisableManager: rule.NewDisableManager(sourceFile, comments),
	}.WithDiagnosticConsumer("test/padding", rule.SeverityError, rule.DiagnosticConsumer{
		Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})
	listeners := testRule.Run(ctx, nil)
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		return node.ForEachChild(visit)
	}
	visit(sourceFile.AsNode())
	listeners[rule.ListenerOnExit(ast.KindEndOfFile)](nil)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics after mutating inputs = %d, want 1", len(diagnostics))
	}
}
