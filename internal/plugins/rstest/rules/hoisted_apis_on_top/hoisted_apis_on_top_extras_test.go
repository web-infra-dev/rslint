// TestHoistedApisOnTopExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise: the full set of module-mock APIs Rstest
// lifts and the non-hoisted counterparts it offers for them, the receiver and
// property shapes the Rstest build does and does not rewrite, the top-level
// shapes accepted for a value-producing call, and every position where the move
// suggestion is withheld while the diagnostic stands. Each case carries an
// inline comment pointing at the branch, Dimension 4 row, or tsgo AST quirk it
// covers.
package hoisted_apis_on_top

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func moveSuggestion(output string) rule_tester.InvalidTestCaseSuggestion {
	return rule_tester.InvalidTestCaseSuggestion{
		MessageId: "suggestMoveHoistedApiToTop",
		Output:    output,
	}
}

func nonHoistedSuggestion(output string) rule_tester.InvalidTestCaseSuggestion {
	return rule_tester.InvalidTestCaseSuggestion{
		MessageId: "suggestUseNonHoistedApi",
		Output:    output,
	}
}

func reported(
	code string,
	line, column, endColumn int,
	suggestions ...rule_tester.InvalidTestCaseSuggestion,
) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId:   "hoistedApisOnTop",
			Line:        line,
			Column:      column,
			EndLine:     line,
			EndColumn:   endColumn,
			Suggestions: suggestions,
		}},
	}
}

func TestHoistedApisOnTopExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&HoistedApisOnTopRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional chain (tsgo flags the chain rather
			// than wrapping it) ----
			// The Rstest build leaves an optional chain alone, so the call runs
			// where it is written and throws instead of being lifted.
			{Code: `if (c) { rs?.mock('./a'); }`},
			{Code: `if (c) { rs.mock?.('./a'); }`},
			{Code: `if (c) { rs?.mock?.('./a'); }`},

			// ---- Dimension 4: access / key forms ----
			// Only a plain dotted member is rewritten.
			{Code: `if (c) { rs['mock']('./a'); }`},
			{Code: "if (c) { rs[`mock`]('./a'); }"},
			{Code: `if (c) { rs[apiName]('./a'); }`},
			// A call reached through import.meta.rstest is not rewritten.
			{Code: `if (c) { import.meta.rstest.rs.mock('./a'); }`},
			// The callee itself is parenthesized rather than the receiver.
			{Code: `if (c) { (rs.mock)('./a'); }`},

			// Locks in hoistedAPIs: the non-hoisted counterparts run where they
			// are written, so they belong in a runtime location.
			{Code: `if (c) { rs.doMock('./a'); }`},
			{Code: `if (c) { rs.doMockRequire('./a'); }`},
			{Code: `if (c) { rs.doUnmock('./a'); }`},
			{Code: `if (c) { rs.doUnmockRequire('./a'); }`},
			// Locks in hoistedAPIs: the rest of the utilities object is not
			// lifted either.
			{Code: `if (c) { rs.fn(); }`},
			{Code: `if (c) { rs.spyOn(target, 'method'); }`},
			{Code: `if (c) { rs.mockObject(target); }`},
			{Code: `if (c) { rs.resetModules(); }`},
			{Code: `if (c) { rs.importActual('./a'); }`},
			{Code: `if (c) { rs.importMock('./a'); }`},
			{Code: `if (c) { rs.requireActual('./a'); }`},
			{Code: `if (c) { rs.mocked(target); }`},

			// Locks in namespaceNames: another receiver spelling is another
			// library's API.
			{Code: `if (c) { vi.mock('./a'); }`},
			{Code: `if (c) { jest.mock('./a'); }`},
			{Code: `if (c) { helpers.mock('./a'); }`},
			{Code: `if (c) { mock('./a'); }`},

			// ---- Dimension 4: parenthesized statement expression ----
			// tsgo keeps the parentheses that ESTree drops; the call is still
			// written as a statement of the file.
			{Code: `(rs.mock('./a'));`},
			{Code: `((rs.mock('./a')));`},

			// Locks in isWrittenAtTopLevel: every shape a value-producing
			// hoisted call is designed for.
			{Code: `const shared = rs.hoisted(() => ({ id: 1 }));`},
			{Code: `let shared = rs.hoisted(() => ({ id: 1 }));`},
			{Code: `var shared = rs.hoisted(() => ({ id: 1 }));`},
			{Code: `export const shared = rs.hoisted(() => ({ id: 1 }));`},
			{Code: `const { id } = rs.hoisted(() => ({ id: 1 }));`},
			{Code: `const [first] = rs.hoisted(() => [1]);`},
			{Code: `const shared = (rs.hoisted(() => 1));`},
			{Code: `const shared = (await rs.hoisted(async () => 1));`},
			{Code: `const a = 1, shared = rs.hoisted(() => 1);`},

			// ---- Dimension 4: receiver wrappers, at the top of the file ----
			{Code: `(rs).mock('./a');`},
			{Code: `rs!.mock('./a');`},
			{Code: `(rs as any).mock('./a');`},

			// ---- Dimension 4: empty argument list / empty file ----
			{Code: `rs.mock();`},
			{Code: `rs.hoisted();`},
			{Code: `;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver (tsgo preserves the
			// wrapper ESTree flattens) ----
			// TypeScript erases these before the module-mock rewrite runs, so
			// the call is lifted all the same.
			reported(
				`if (c) { (rs).mock('./a'); }`,
				1, 10, 26,
				moveSuggestion(`(rs).mock('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { (rs).doMock('./a'); }`),
			),
			reported(
				`if (c) { ((rs)).mock('./a'); }`,
				1, 10, 28,
				moveSuggestion(`((rs)).mock('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { ((rs)).doMock('./a'); }`),
			),
			// ---- Dimension 4: TS non-null assertion on the receiver ----
			reported(
				`if (c) { rs!.mock('./a'); }`,
				1, 10, 25,
				moveSuggestion(`rs!.mock('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { rs!.doMock('./a'); }`),
			),
			// ---- Dimension 4: TS type-expression wrappers on the receiver ----
			reported(
				`if (c) { (rs as any).mock('./a'); }`,
				1, 10, 33,
				moveSuggestion(`(rs as any).mock('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { (rs as any).doMock('./a'); }`),
			),
			reported(
				`if (c) { (rs satisfies unknown).mock('./a'); }`,
				1, 10, 44,
				moveSuggestion(`(rs satisfies unknown).mock('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { (rs satisfies unknown).doMock('./a'); }`),
			),

			// Locks in hoistedAPIs: the two CommonJS module operations and
			// their non-hoisted counterparts.
			reported(
				`if (c) { rs.mockRequire('./a'); }`,
				1, 10, 31,
				moveSuggestion(`rs.mockRequire('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { rs.doMockRequire('./a'); }`),
			),
			reported(
				`if (c) { rs.unmockRequire('./a'); }`,
				1, 10, 33,
				moveSuggestion(`rs.unmockRequire('./a');
if (c) { ; }`),
				nonHoistedSuggestion(`if (c) { rs.doUnmockRequire('./a'); }`),
			),

			// Locks in namespaceNames: the receiver is matched by the name
			// written at the call site, with no import or scope analysis, which
			// is exactly what the Rstest build does. A local binding that
			// shadows the name does not stop the call from being lifted.
			reported(
				`const rs = { mock(_path: string) {} };
if (c) {
  rs.mock('./a');
}`,
				3, 3, 17,
				moveSuggestion(`rs.mock('./a');
const rs = { mock(_path: string) {} };
if (c) {
  ;
}`),
				nonHoistedSuggestion(`const rs = { mock(_path: string) {} };
if (c) {
  rs.doMock('./a');
}`),
			),
			reported(
				`import * as rs from '@rstest/core';
if (c) {
  rs.mock('./a');
}`,
				3, 3, 17,
				moveSuggestion(`import * as rs from '@rstest/core';
rs.mock('./a');
if (c) {
  ;
}`),
				nonHoistedSuggestion(`import * as rs from '@rstest/core';
if (c) {
  rs.doMock('./a');
}`),
			),

			// Locks in lastImportEnd: the last import of the file wins, and a
			// type-only import counts like any other.
			reported(
				`import a from './a';
import type { B } from './b';
if (c) {
  rs.mock('./a');
}`,
				4, 3, 17,
				moveSuggestion(`import a from './a';
import type { B } from './b';
rs.mock('./a');
if (c) {
  ;
}`),
				nonHoistedSuggestion(`import a from './a';
import type { B } from './b';
if (c) {
  rs.doMock('./a');
}`),
			),
			// Locks in lastImportEnd: an import-equals declaration is not an
			// import declaration, so the call goes to the very start.
			reported(
				`import fs = require('fs');
if (c) {
  rs.mock('./a');
}`,
				3, 3, 17,
				moveSuggestion(`rs.mock('./a');
import fs = require('fs');
if (c) {
  ;
}`),
				nonHoistedSuggestion(`import fs = require('fs');
if (c) {
  rs.doMock('./a');
}`),
			),

			// Locks in moveToTopFixes arm 2: a value-producing call bound by a
			// single-declarator statement moves with its binding. Lifting only
			// the call would leave the binding assigned when its block runs
			// while the factory has already executed.
			reported(
				`if (c) {
  const shared = rs.hoisted(() => ({ id: 1 }));
}`,
				2, 18, 47,
				moveSuggestion(`const shared = rs.hoisted(() => ({ id: 1 }));
if (c) {
  
}`),
			),
			reported(
				`if (c) {
  const shared = await rs.hoisted(async () => ({ id: 1 }));
}`,
				2, 24, 59,
				moveSuggestion(`const shared = await rs.hoisted(async () => ({ id: 1 }));
if (c) {
  
}`),
			),
			// Locks in movableVariableStatement: a second declarator would move
			// with the binding, so nothing moves and the diagnostic stands
			// alone.
			reported(
				`if (c) {
  const shared = rs.hoisted(() => 1), other = 2;
}`,
				2, 18, 37,
			),
			// Locks in variableStatementOf: a `for` header is not a statement
			// of its own.
			reported(
				`for (const shared = rs.hoisted(() => 1); ; ) {}`,
				1, 21, 40,
			),
			// Locks in moveToTopFixes: a value-producing call whose value is
			// consumed by a surrounding expression has no text to stand in for
			// it, so only the diagnostic is reported.
			reported(
				`if (c) { register(rs.hoisted(() => 1)); }`,
				1, 19, 38,
			),
			reported(
				`const shared = { value: rs.hoisted(() => 1) };`,
				1, 25, 44,
			),

			// Locks in moveToTopFixes arm 3: the other hoisted APIs evaluate to
			// undefined, so `undefined` is an exact stand-in wherever their
			// value is consumed — including at the top of the file, where a
			// binding is still not the shape these APIs are written in.
			reported(
				`const done = rs.mock('./a');`,
				1, 14, 28,
				moveSuggestion(`rs.mock('./a');
const done = undefined;`),
				nonHoistedSuggestion(`const done = rs.doMock('./a');`),
			),
			reported(
				`const run = () => rs.unmock('./a');`,
				1, 19, 35,
				moveSuggestion(`rs.unmock('./a');
const run = () => undefined;`),
				nonHoistedSuggestion(`const run = () => rs.doUnmock('./a');`),
			),
			reported(
				`if (c) { c && rs.mock('./a'); }`,
				1, 15, 29,
				moveSuggestion(`rs.mock('./a');
if (c) { c && undefined; }`),
				nonHoistedSuggestion(`if (c) { c && rs.doMock('./a'); }`),
			),

			// Locks in moveToTopFixes arm 1: an `if` written without braces
			// keeps a body, because the statement becomes an empty statement
			// rather than being deleted.
			reported(
				`if (c) rs.mock('./a')`,
				1, 8, 22,
				moveSuggestion(`rs.mock('./a');
if (c) ;`),
				nonHoistedSuggestion(`if (c) rs.doMock('./a')`),
			),

			// ---- Dimension 4: nesting / traversal boundaries ----
			// A class static block is a runtime location like any other.
			reported(
				`class Suite {
  static {
    rs.mock('./a');
  }
}`,
				3, 5, 19,
				moveSuggestion(`rs.mock('./a');
class Suite {
  static {
    ;
  }
}`),
				nonHoistedSuggestion(`class Suite {
  static {
    rs.doMock('./a');
  }
}`),
			),

			// ---- Real-user: the shape the Rstest mocking guide warns about,
			// where a module mock is written inside the test that needs it ----
			reported(
				`import { rs, test } from '@rstest/core';

test('reads the cached value', () => {
  rs.mock('./cache');
});`,
				4, 3, 21,
				moveSuggestion(`import { rs, test } from '@rstest/core';
rs.mock('./cache');

test('reads the cached value', () => {
  ;
});`),
				nonHoistedSuggestion(`import { rs, test } from '@rstest/core';

test('reads the cached value', () => {
  rs.doMock('./cache');
});`),
			),
			// ---- Real-user: the "share values with a hoisted mock factory"
			// pattern, written inside the suite that uses it ----
			reported(
				`import { describe, rs } from '@rstest/core';

describe('cache', () => {
  const mocks = rs.hoisted(() => ({ read: rs.fn() }));
});`,
				4, 17, 54,
				moveSuggestion(`import { describe, rs } from '@rstest/core';
const mocks = rs.hoisted(() => ({ read: rs.fn() }));

describe('cache', () => {
  
});`),
			),
		},
	)
}

// TestHoistedApisOnTopEditDemand locks in that the diagnostic is identical
// under every edit demand, that this rule never produces an autofix, and that
// the suggestions appear only when they are asked for.
func TestHoistedApisOnTopEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`import a from './a';
if (c) {
  rs.mock('./a');
  rs.hoisted(() => 1);
}`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     HoistedApisOnTopRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return HoistedApisOnTopRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		want := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
	}

	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		for _, diagnostic := range diagnostics {
			if diagnostic.FixesPtr != nil {
				t.Fatal("hoisted-apis-on-top unexpectedly materialized an autofix")
			}
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly} {
		for _, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Fatal("suggestions materialized without being demanded")
			}
		}
	}
	// `mock` carries the move and the non-hoisted-counterpart suggestion;
	// `hoisted` has no non-hoisted counterpart and carries only the move.
	for index, want := range []int{2, 1} {
		if suggestionOnly[index].Suggestions == nil {
			t.Fatalf("diagnostic %d materialized no suggestions", index)
		}
		if got := len(*suggestionOnly[index].Suggestions); got != want {
			t.Fatalf("diagnostic %d produced %d suggestions, want %d", index, got, want)
		}
		if !reflect.DeepEqual(suggestionOnly[index].Suggestions, allEdits[index].Suggestions) {
			t.Fatalf("diagnostic %d suggestions differ between demands", index)
		}
	}
}
