// TestSwitchExhaustivenessCheckExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
package switch_exhaustiveness_check

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSwitchExhaustivenessCheckExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: parenthesized discriminant (single level) ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch ((value)) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: parenthesized discriminant (multi-level) ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch (((value))) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: TS non-null assertion discriminant ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction | undefined;
switch (value!) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: "as" type-expression discriminant ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: unknown;
switch (value as Direction) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: "satisfies" discriminant ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: 'north' | 'south';
switch (value satisfies Direction) {
  case 'north':
    break;
  case 'south':
    break;
}
`},
			// ---- Dimension 4: optional-chain discriminant ----
			{Code: `
declare const obj: { prop: 'a' | 'b' } | undefined;
switch (obj?.prop) {
  case 'a':
    break;
  case 'b':
    break;
  case undefined:
    break;
}
`},
			// ---- Dimension 4: parenthesized case-test expression ----
			{Code: `
type Direction = 'north' | 'south';
declare const value: Direction;
switch (value) {
  case ('north'):
    break;
  case ('south'):
    break;
}
`},
			// ---- Dimension 4: fallthrough (grouped) case clauses cover every constituent ----
			{Code: `
type Direction = 'north' | 'south' | 'east' | 'west';
declare const value: Direction;
switch (value) {
  case 'north':
  case 'south':
    break;
  case 'east':
  case 'west':
    break;
}
`},
			// ---- Dimension 4: nested switch — inner switch's cases don't count toward the outer switch ----
			{Code: `
type Outer = 'a' | 'b';
type Inner = 'x' | 'y';
declare const outer: Outer;
declare const inner: Inner;
switch (outer) {
  case 'a':
    switch (inner) {
      case 'x':
        break;
      case 'y':
        break;
    }
    break;
  case 'b':
    break;
}
`},
			// ---- Real-user: literal case matching an intersection-branded union constituent behind a type alias (issue tracker pattern for branded types) ----
			{Code: `
type Brand<T, B extends string> = T & { readonly __brand: B };
type UserId = Brand<string, 'UserId'>;
type Status = 'active' | 'inactive';
declare const value: Status | UserId;
switch (value) {
  case 'active':
    break;
  case 'inactive':
    break;
  default:
    break;
}
`, Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": true}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: computed enum member name via multi-line template literal ----
			{
				Code: `
enum Enum {
  'a' = 1,
  ` + "[`key-with\n\n          new-line`]" + ` = 2,
}

declare const a: Enum;

switch (a) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum Enum {
  'a' = 1,
  ` + "[`key-with\n\n          new-line`]" + ` = 2,
}

declare const a: Enum;

switch (a) {
case Enum.a: { throw new Error('Not implemented yet: Enum.a case') }
case Enum['key-with\n\n          new-line']: { throw new Error('Not implemented yet: Enum[\'key-with\\n\\n          new-line\'] case') }
}
`,
							},
						},
					},
				},
			},
			// ---- Real-user (typescript-eslint#11842): a bare numeric literal `case 0:` does not
			// count as covering a numeric enum member with the same underlying value — TS treats
			// enum-member types and plain numeric-literal types as distinct type objects even when
			// their runtime values match, so the Set-based case-type lookup in getSwitchMetadata
			// (keyed by *checker.Type identity) does not consider it covered. This locks in the
			// checker-identity-based caseTypes.has(...) branch in getSwitchMetadata, matching
			// upstream's real reported behavior (closed as working-as-intended: use the qualified
			// enum member in the case test). ----
			{
				Code: `
enum DataType {
  First = 0,
  Second = 1,
}

function test(type: DataType) {
  switch (type) {
    case 0:
      break;
    case 1:
      break;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: DataType.First | DataType.Second",
						Line:      8,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum DataType {
  First = 0,
  Second = 1,
}

function test(type: DataType) {
  switch (type) {
    case 0:
      break;
    case 1:
      break;
    case DataType.First: { throw new Error('Not implemented yet: DataType.First case') }
    case DataType.Second: { throw new Error('Not implemented yet: DataType.Second case') }
  }
}
`,
							},
						},
					},
				},
			},
			// ---- Branch lock-in: checkSwitchUnnecessaryDefaultCase fires against a
			// comment-based default (not just a real `default:` clause). The switch below is
			// fully covered by explicit literal cases and the trailing comment matches the
			// (default) "no default" pattern, so getSwitchMetadata's defaultCase resolves via
			// getCommentDefaultCase rather than a real clause — this locks in that
			// checkSwitchUnnecessaryDefaultCase's `meta.defaultCase.exists()` check treats a
			// comment-based default the same as a real one, and that defaultCaseRef.reportRange
			// anchors the diagnostic on the comment when there is no clause node. ----
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  case 'b':
    break;
  // no default
}
`,
				Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase", Line: 8, Column: 3, EndLine: 8, EndColumn: 16},
				},
			},
			// ---- Branch lock-in: getCommentDefaultCase matches a block comment, not just a
			// line comment (the `/*...*/`-stripping branch in getCommentDefaultCase). ----
			{
				Code: `
declare const value: number;
switch (value) {
  case 0:
    break;
  case 1:
    break;
  /* no default */
}
`,
				Options: map[string]any{"requireDefaultForNonUnion": true},
				Errors:  nil,
			},
		},
	)

	t.Run("EditDemand", testSwitchExhaustivenessCheckEditDemand)
}

func testSwitchExhaustivenessCheckEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"type T = 1 | 2;\n"+
			"function test(value: T): number {\n"+
			"  switch (value) {\n"+
			"    case 1:\n"+
			"      return 1;\n"+
			"  }\n"+
			"}\n",
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
			Program:      program,
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:             SwitchExhaustivenessCheckRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: SwitchExhaustivenessCheckRule.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return SwitchExhaustivenessCheckRule.Run(ctx, nil)
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
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}

	want := withoutEdits(diagnostics[rule.EditDemandAll][0])
	for demand, demandDiagnostics := range diagnostics {
		if got := withoutEdits(demandDiagnostics[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("diagnostic changed for demand %d:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
		if demandDiagnostics[0].FixesPtr != nil {
			t.Errorf("demand %d unexpectedly has autofixes (rule only emits suggestions)", demand)
		}
	}

	suggestionOnly := diagnostics[rule.EditDemandSuggestion][0].Suggestions
	allEdits := diagnostics[rule.EditDemandAll][0].Suggestions
	if suggestionOnly == nil || !reflect.DeepEqual(suggestionOnly, allEdits) {
		t.Fatalf("suggestions differ between suggestion and all-edits demand")
	}
	if len(*suggestionOnly) != 1 || (*suggestionOnly)[0].Message.Id != "addMissingCases" {
		t.Fatalf("suggestions = %#v, want a single addMissingCases suggestion", *suggestionOnly)
	}

	if diagnostics[rule.EditDemandNone][0].Suggestions != nil ||
		diagnostics[rule.EditDemandAutofix][0].Suggestions != nil {
		t.Errorf("suggestions attached without suggestion demand")
	}
}
