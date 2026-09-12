package no_unnecessary_assertion

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed testdata/types.txtar
var typeFixtures []byte

func typeRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	archive, err := txtarfs.Parse("types.txtar", typeFixtures)
	if err != nil {
		t.Fatal(err)
	}
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	root := fixtures.GetRootDir()
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

func TestNoUnnecessaryAssertionCompilerOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		typeRoot(t), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.json"},
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.strict-null.json"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.unstrict.json", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStrictNullCheck"}}},
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.override.json", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStrictNullCheck"}}},
			{Code: "expect(value).toBe(other)", TSConfig: "tsconfig.default.json", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noStrictNullCheck"}}},
		},
	)
}

func TestNoUnnecessaryAssertionRstestMatchers(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "declare const value: string | null; expect(value).to.be.null;"},
			{Code: "declare const value: string | undefined; expect(value).to.be.undefined;"},
			{Code: "declare const value: string | number; expect(value).to.be.NaN;"},
			{Code: "declare const value: any; expect(value).toBeNull(); expect(value).to.be.NaN;"},
			{Code: "declare const value: unknown; expect(value).toBeUndefined(); expect(value).to.be.null;"},
			{Code: "declare const value: Promise<string>; expect(value).resolves.toBeNull();"},
			{Code: "declare const value: Promise<string>; expect(value).rejects.not.toBeUndefined();"},
			{Code: "expect.poll((): string | null => null).toBeNull();"},
			{Code: "expect.poll(async (): Promise<string | undefined> => undefined).toBeDefined();"},
			{Code: "function check<T>(value: T) { expect(value).toBeNull(); }"},
			{Code: "function check<T>(read: () => T) { expect.poll(read).toBeNaN(); }"},
			{Code: "function check<T>(value: T extends string ? null : number) { expect(value).toBeNull(); }"},
			{Code: "function check<T>(key: keyof T) { expect(key).toBeNaN(); }"},
			{Code: "type UserId = number & { readonly __brand: unique symbol }; declare const value: UserId; expect(value).toBeNaN();"},
			{Code: "function check<T>(value: T & {}) { expect(value).toBeNaN(); }"},
			{Code: "declare function cleanup(): void; expect(cleanup()).toBeUndefined();"},
			{Code: "declare function cleanup(): void; expect(cleanup()).toBeDefined();"},
			{Code: "const read: () => void = () => null; expect(read()).toBeNull();"},
			{Code: "const read: () => void = () => NaN; expect(read()).toBeNaN();"},
			{Code: "expect.poll((): string => 'ready').to.be.null;"},
			{Code: "expect.poll(nonCallable).toBeNull();"},
			{Code: "expect.element(locator).toBeNull();"},
			{Code: "expect('ready').to.exist; expect('ready').to.exists; expect('ready').toBeNullable();"},
			{Code: "expect({}).to.have.property('missing').undefined;"},
			{Code: "expect('ready').to.equal('ready').and.be.null;"},
			{Code: "expect('ready')[matcherName];"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "expect('ready').to.be.null;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssertion", Message: "Unnecessary assertion, subject cannot be null", Line: 1, Column: 1, EndLine: 1, EndColumn: 27}}},
			{Code: "expect('ready').to.be.null.and.empty;", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "expect('ready').not.to.be.undefined;", Errors: []rule_tester.InvalidTestCaseError{diagnostic("undefined", 1)}},
			{Code: "expect('ready').to.be.NaN;", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 1)}},
			{Code: "expect.soft('ready').to.not.be.null;", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "expect.poll((): string => 'ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "expect.poll(async (): Promise<string> => 'ready').toBeUndefined();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("undefined", 1)}},
			{Code: "expect('ready')['to']['be']['undefined'];", Errors: []rule_tester.InvalidTestCaseError{diagnostic("undefined", 1)}},
		},
	)
}

func TestNoUnnecessaryAssertionTypesAndNarrowing(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "declare const value: null | string; expect(value).toBeNull();"},
			{Code: "declare const value: undefined | string; expect(value).toBeDefined();"},
			{Code: "declare const value: number | object; expect(value).toBeNaN();"},
			{Code: "function check<T extends string | null>(value: T) { expect(value).toBeNull(); }"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "function check<T extends string>(value: T) { expect(value).toBeNull(); }", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "function check<T extends number>(value: T) { expect(value).toBeNull(); }", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "function check<T>(key: keyof T) { expect(key).toBeNull(); }", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "type Name = string & { readonly __brand: unique symbol }; declare const value: Name; expect(value).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 1)}},
			{Code: "function check(value: string | null) { if (value !== null) { expect(value).toBeNull(); } }", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "declare const value: bigint; expect(value).toBeNaN();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("a number", 1)}},
		},
	)
}

func TestNoUnnecessaryAssertionCrossFileAndAmbientTypes(t *testing.T) {
	entry, err := txtarfs.Parse("types.txtar", typeFixtures)
	if err != nil {
		t.Fatal(err)
	}
	data, err := entry.ReadFile("entry.ts")
	if err != nil {
		t.Fatal(err)
	}
	rule_tester.RunRuleTester(
		typeRoot(t), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		nil,
		[]rule_tester.InvalidTestCase{{
			Code:     string(data),
			FileName: "entry.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				diagnostic("null", 2),
				diagnostic("undefined", 6),
				diagnostic("null", 10),
				diagnostic("a number", 12),
			},
		}},
	)
}

func TestNoUnnecessaryAssertionRstestSources(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: "import { expect } from 'vitest'; expect('ready').toBeNull();"},
			{Code: "function check(expect: (value: unknown) => any) { expect('ready').toBeNull(); }"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "expect('ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "import { expect as check } from '@rstest/core'; check('ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "import * as rstest from '@rstest/core'; rstest.expect('ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "const { expect: check } = require('rstack/test'); check('ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "import.meta.rstest.expect('ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "test('value', ({ expect }) => { expect('ready').toBeNull(); });", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
			{Code: "import { expect as check } from '@rstest/playwright'; check('ready').toBeNull();", Errors: []rule_tester.InvalidTestCaseError{diagnostic("null", 1)}},
		},
	)
}

func TestNoUnnecessaryAssertionIsFilteredFromSourceOnlyPrograms(t *testing.T) {
	if !NoUnnecessaryAssertionRule.RequiresTypeInfo {
		t.Fatal("rstest/no-unnecessary-assertion must require type information")
	}
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: "expect('ready').toBeNull();"})
	host := utils.CreateCompilerHost(root.Dir, fs)
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	ruleRan := false
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{program},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             NoUnnecessaryAssertionRule.Name,
				Severity:         rule.SeverityError,
				RequiresTypeInfo: NoUnnecessaryAssertionRule.RequiresTypeInfo,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					ruleRan = true
					return NoUnnecessaryAssertionRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer:       rule.DiagnosticConsumer{Report: func(rule.RuleDiagnostic) {}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ruleRan {
		t.Fatal("type-aware rule ran in a source-only Program")
	}
	if _, ok := result.ExecutedRules[NoUnnecessaryAssertionRule.Name]; ok {
		t.Fatal("filtered rule was retained in ExecutedRules")
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("source-only test unexpectedly provided a TypeChecker")
	}
}

func TestTypeCapabilities(t *testing.T) {
	root := typeRoot(t)
	host := utils.CreateCompilerHost(root.Dir, root.FS)
	program, err := utils.CreateProgram(true, root.FS, root.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	tc, done := program.GetTypeChecker(t.Context())
	defer done()
	file := program.GetSourceFile("probe.ts")
	if file == nil {
		t.Fatal("missing probe.ts")
	}
	if !utils.IsStrictCompilerOptionEnabled(program.Options(), program.Options().StrictNullChecks) {
		t.Fatal("strictNullChecks not inherited")
	}
	want := map[string]checker.TypeFlags{
		"definite()":           checker.TypeFlagsString,
		"nullable()":           checker.TypeFlagsNull,
		"optional()":           checker.TypeFlagsUndefined,
		"numeric()":            checker.TypeFlagsNumberLike,
		"ambient()":            checker.TypeFlagsString,
		"ambientNullable()":    checker.TypeFlagsNull,
		"ambientUnknown()":     checker.TypeFlagsUnknown,
		"Promise.resolve(1)":   checker.TypeFlagsObject,
		"NaN":                  checker.TypeFlagsNumberLike,
		"conditional('value')": checker.TypeFlagsNull,
		"conditional(1)":       checker.TypeFlagsNumberLike,
		"promised()":           checker.TypeFlagsObject,
	}
	seen := make(map[string]bool)
	valueCount := 0
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsCallExpression(node) && ast.IsIdentifier(node.Expression()) && node.Expression().Text() == "inspect" {
			argument := node.Arguments()[0]
			text := strings.TrimSpace(file.Text()[argument.Pos():argument.End()])
			typ := tc.GetTypeAtLocation(argument)
			switch text {
			case "value":
				valueCount++
				if valueCount == 1 {
					if !utils.IsTypeFlagSetWithUnion(
						utils.GetConstrainedTypeAtLocation(tc, argument),
						checker.TypeFlagsNull,
					) {
						t.Fatalf("generic constraint unavailable: type=%v constrained=%v", checker.Type_flags(typ), checker.Type_flags(utils.GetConstrainedTypeAtLocation(tc, argument)))
					}
				} else if !utils.IsTypeFlagSet(typ, checker.TypeFlagsString) {
					t.Fatal("control-flow narrowing unavailable")
				}
			case "async () => promised()":
				sigs := utils.GetCallSignatures(tc, typ)
				if len(sigs) != 1 {
					t.Fatal("callback signature unavailable")
				}
				awaited := checker.Checker_getAwaitedType(tc, checker.Checker_getReturnTypeOfSignature(tc, sigs[0]))
				if !utils.IsTypeFlagSetWithUnion(awaited, checker.TypeFlagsUndefined) {
					t.Fatal("awaited callback return unavailable")
				}
			default:
				flag, ok := want[text]
				if !ok || !utils.IsTypeFlagSetWithUnion(typ, flag) {
					t.Fatalf("%s: flags %v, want %v", text, checker.Type_flags(typ), flag)
				}
				seen[text] = true
			}
			if text == "ambient()" || text == "NaN" {
				name := argument
				if ast.IsCallExpression(name) {
					name = name.Expression()
				}
				symbol := tc.GetSymbolAtLocation(name)
				if symbol == nil || len(symbol.Declarations) == 0 {
					t.Fatalf("%s: symbol unavailable", text)
				}
				source := ast.GetSourceFileOfNode(symbol.Declarations[0])
				if text == "ambient()" && !strings.HasSuffix(source.FileName(), "/globals.d.ts") {
					t.Fatal("ambient symbol did not resolve to globals.d.ts")
				}
				if text == "NaN" && !strings.Contains(source.FileName(), "/lib.es5.d.ts") {
					t.Fatal("NaN did not resolve to default lib")
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
	if len(seen) != len(want) || valueCount != 2 {
		t.Fatalf("incomplete capability probe: %v, values=%d", seen, valueCount)
	}
}
