package no_empty_interface

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyInterfaceRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyInterfaceRule, []rule_tester.ValidTestCase{
		// Valid cases
		{Code: `
interface Foo {
  name: string;
}
`},
		{Code: `
interface Foo {
  name: string;
}

interface Bar {
  age: number;
}

// valid because extending multiple interfaces can be used instead of a union type
interface Baz extends Foo, Bar {}
`},
		{
			Code: `
interface Foo {
  name: string;
}

interface Bar extends Foo {}
`,
			Options: map[string]interface{}{"allowSingleExtends": true},
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Bar {}
`,
			Options: map[string]interface{}{"allowSingleExtends": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases
		{
			Code: "interface Foo {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmpty",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `interface Foo extends {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmpty",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Baz {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			Output: []string{`
interface Foo {
  props: string;
}

type Bar = Foo

class Baz {}
`},
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

class Bar {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			// No output when merged with class
		},
		{
			Code: `
interface Foo {
  props: string;
}

interface Bar extends Foo {}

const bar = class Bar {};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			Output: []string{`
interface Foo {
  props: string;
}

type Bar = Foo

const bar = class Bar {};
`},
		},
		{
			Code: `
interface Foo {
  name: string;
}

interface Bar extends Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      6,
					Column:    11,
					EndLine:   6,
					EndColumn: 14,
				},
			},
			Options: map[string]interface{}{"allowSingleExtends": false},
			Output: []string{`
interface Foo {
  name: string;
}

type Bar = Foo
`},
		},
		{
			Code: "interface Foo extends Array<number> {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				},
			},
			Output: []string{`type Foo = Array<number>`},
		},
		{
			Code: "interface Foo extends Array<number | {}> {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 14,
				},
			},
			Output: []string{`type Foo = Array<number | {}>`},
		},
		{
			Code: `
interface Bar {
  bar: string;
}
interface Foo extends Array<Bar> {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      5,
					Column:    11,
					EndLine:   5,
					EndColumn: 14,
				},
			},
			Output: []string{`
interface Bar {
  bar: string;
}
type Foo = Array<Bar>
`},
		},
		{
			Code: `
type R = Record<string, unknown>;
interface Foo extends R {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      3,
					Column:    11,
					EndLine:   3,
					EndColumn: 14,
				},
			},
			Output: []string{`
type R = Record<string, unknown>;
type Foo = R
`},
		},
		{
			Code: `
interface Foo<T> extends Bar<T> {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noEmptyWithSuper",
					Line:      2,
					Column:    11,
					EndLine:   2,
					EndColumn: 14,
				},
			},
			Output: []string{`
type Foo<T> = Bar<T>
`},
		},
	})
}

func TestNoEmptyInterfaceRuleAmbient(t *testing.T) {
	const declaredModuleSource = `declare module FooBar {
  type Baz = typeof baz;
  export interface Bar extends Baz {}
}`
	const declaredModuleSuggestion = `declare module FooBar {
  type Baz = typeof baz;
  export type Bar = Baz
}`

	invalid := make([]rule_tester.InvalidTestCase, 0, 8)
	for _, fileName := range []string{"ambient-standard.d.ts", "ambient-module.d.mts", "ambient-commonjs.d.cts", "ambient-custom.d.custom.ts"} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:     declaredModuleSource,
			FileName: fileName,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      3,
				Column:    20,
				EndLine:   3,
				EndColumn: 23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "noEmptyWithSuper",
					Output:    declaredModuleSuggestion,
				}},
			}},
		})
	}

	invalid = append(invalid,
		rule_tester.InvalidTestCase{
			Code: `declare namespace Outer {
  namespace Inner {
    interface Nested extends Base {}
  }
}`,
			FileName: "nested.d.ts",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      3,
				Column:    15,
				EndLine:   3,
				EndColumn: 21,
			}},
			Output: []string{`declare namespace Outer {
  namespace Inner {
    type Nested = Base
  }
}`},
		},
		rule_tester.InvalidTestCase{
			Code: `declare namespace Outer.Inner {
  interface Nested extends Base {}
}`,
			FileName: "dotted.d.ts",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      2,
				Column:    13,
				EndLine:   2,
				EndColumn: 19,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "noEmptyWithSuper",
					Output: `declare namespace Outer.Inner {
  type Nested = Base
}`,
				}},
			}},
		},
		rule_tester.InvalidTestCase{
			Code:     "interface TopLevel extends Base {}",
			FileName: "top-level.d.ts",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      1,
				Column:    11,
				EndLine:   1,
				EndColumn: 19,
			}},
			Output: []string{"type TopLevel = Base"},
		},
		rule_tester.InvalidTestCase{
			Code: `declare module FooBar {
  interface Ordinary extends Base {}
}`,
			FileName: "ordinary.ts",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      2,
				Column:    13,
				EndLine:   2,
				EndColumn: 21,
			}},
			Output: []string{`declare module FooBar {
  type Ordinary = Base
}`},
		},
	)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyInterfaceRule, nil, invalid)
}

func TestNoEmptyInterfaceRuleSafeEdits(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEmptyInterfaceRule, nil, []rule_tester.InvalidTestCase{
		{
			Code: `export /* keep export */ interface Foo<
  T /* keep type */,
> extends Base<T> {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      1,
				Column:    36,
				EndLine:   1,
				EndColumn: 39,
			}},
			Output: []string{`export /* keep export */ type Foo<
  T /* keep type */,
> = Base<T>`},
		},
		{
			Code: "export declare interface Foo extends Base {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      1,
				Column:    26,
				EndLine:   1,
				EndColumn: 29,
			}},
			Output: []string{"export type Foo = Base"},
		},
		{
			Code: "export default interface Foo extends Base {}",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      1,
				Column:    26,
				EndLine:   1,
				EndColumn: 29,
			}},
		},
		{
			Code: `interface Base { value: string }
class Merged<T> {}
interface Merged<T> extends Base {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      3,
				Column:    11,
				EndLine:   3,
				EndColumn: 17,
			}},
		},
		{
			Code: `interface Base { value: string }
class Separate {}
namespace Other {
  interface Separate extends Base {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "noEmptyWithSuper",
				Line:      4,
				Column:    13,
				EndLine:   4,
				EndColumn: 21,
			}},
			Output: []string{`interface Base { value: string }
class Separate {}
namespace Other {
  type Separate = Base
}`},
		},
	})
}

func TestNoEmptyInterfaceRuleDoesNotRequireTypeInformation(t *testing.T) {
	t.Parallel()
	if NoEmptyInterfaceRule.RequiresTypeInfo {
		t.Fatal("no-empty-interface must run without type information")
	}

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"interface Empty {}\ninterface Derived extends Base {}",
		"without-type-information.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := lintNoEmptyInterfaceForTest(lintprogram.NewFromCompiler(program), sourceFile, false, rule.EditDemandNone)
	if len(diagnostics) != 2 || diagnostics[0].Message.Id != "noEmpty" || diagnostics[1].Message.Id != "noEmptyWithSuper" {
		t.Fatalf("diagnostics without type information = %#v, want noEmpty and noEmptyWithSuper", diagnostics)
	}
}

func TestNoEmptyInterfaceRuleEditDemand(t *testing.T) {
	t.Parallel()

	const code = `interface Base { value: string }
interface Autofix extends Base {}
declare module Ambient {
  export interface Suggested extends Base {}
}
class Merged<T> {}
interface Merged<T> extends Base {}
export default interface Defaulted extends Base {}
// eslint-disable-next-line @typescript-eslint/no-empty-interface
interface Disabled extends Base {}`

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(code, "edit-demand.d.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	lintProgram := lintprogram.NewFromCompiler(program)

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{}
	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		diagnostics[demand] = lintNoEmptyInterfaceForTest(lintProgram, sourceFile, false, demand)
		if got := len(diagnostics[demand]); got != 4 {
			t.Fatalf("demand %d produced %d diagnostics, want 4", demand, got)
		}
	}

	for index, all := range diagnostics[rule.EditDemandAll] {
		for demand, result := range diagnostics {
			got := result[index]
			if got.Message.Id != all.Message.Id ||
				got.Message.Description != all.Message.Description ||
				got.Range != all.Range ||
				got.Severity != all.Severity ||
				got.RuleName != all.RuleName {
				t.Errorf("demand %d changed diagnostic %d identity", demand, index)
			}
		}
	}

	for _, demand := range []rule.EditDemand{rule.EditDemandAutofix, rule.EditDemandAll} {
		fixes := diagnostics[demand][0].FixesPtr
		if fixes == nil || len(*fixes) != 1 || (*fixes)[0].Text != "type Autofix = Base" {
			t.Fatalf("demand %d autofix = %#v, want Autofix replacement", demand, fixes)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandSuggestion} {
		if diagnostics[demand][0].FixesPtr != nil {
			t.Fatalf("demand %d unexpectedly materialized the autofix", demand)
		}
	}

	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		suggestions := diagnostics[demand][1].Suggestions
		if suggestions == nil || len(*suggestions) != 1 {
			t.Fatalf("demand %d suggestions = %#v, want one", demand, suggestions)
		}
		fixes := (*suggestions)[0].Fixes()
		if len(fixes) != 1 || fixes[0].Text != "type Suggested = Base" {
			t.Fatalf("demand %d suggestion fixes = %#v, want Suggested replacement", demand, fixes)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		if diagnostics[demand][1].Suggestions != nil {
			t.Fatalf("demand %d unexpectedly materialized the suggestion", demand)
		}
	}

	for demand, result := range diagnostics {
		for index, diagnostic := range result {
			if diagnostic.FixesPtr != nil && index != 0 {
				t.Errorf("demand %d diagnostic %d unexpectedly has an autofix", demand, index)
			}
			if diagnostic.Suggestions != nil && index != 1 {
				t.Errorf("demand %d diagnostic %d unexpectedly has suggestions", demand, index)
			}
		}
	}
}

func lintNoEmptyInterfaceForTest(
	program *lintprogram.Program,
	sourceFile *ast.SourceFile,
	hasTypeInfo bool,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	diagnostics := make([]rule.RuleDiagnostic, 0, 4)
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     program,
		File:        sourceFile.FileName(),
		HasTypeInfo: hasTypeInfo,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     "@typescript-eslint/no-empty-interface",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoEmptyInterfaceRule.Run(ctx, nil)
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
	return diagnostics
}
