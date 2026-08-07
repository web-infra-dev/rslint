// TestNoUnnecessaryTypeParametersExtras locks in branches and edge shapes
// that the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without
// breaking a named lock-in.
package no_unnecessary_type_parameters

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeParametersExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeParametersRule, []rule_tester.ValidTestCase{
		// ---- Dimension 1: AST node types ----
		{Code: `
// Async function: T used in param and (via await) the resolved return type.
async function identity<T>(arg: T): Promise<T> {
  return arg;
}
`},
		{Code: `
// Generator function: T used in the yielded value type and the parameter.
function* gen<T>(seed: T): Generator<T> {
  yield seed;
}
`},
		{Code: `
// Async generator.
async function* gen<T>(seed: T): AsyncGenerator<T> {
  yield seed;
}
`},

		// ---- Dimension 4: TSConstructSignatureDeclaration is never checked ----
		// Locks in upstream's selector list omission of
		// TSConstructSignatureDeclaration: even a type parameter used only
		// once in an interface's construct signature must never be reported,
		// because the rule's selectors never visit this node kind at all.
		{Code: `
interface I {
  new <T>(value: T): void;
}
`},

		// ---- Real-user: typescript-eslint/typescript-eslint#9961 (open) ----
		// Upstream's own class handling has a known, intentionally-unfixed
		// false positive: a class whose single member uses T only once is
		// still reported, even though (unlike a bare function) a reader might
		// expect per-member reasoning. rslint matches this for parity — see
		// the invalid cases below for the interface contrast this issue
		// documents (interface member usage is correctly NOT reported).
		{Code: `
interface ValueInterface<T> {
  value: T;
}
`},

		// ---- Real-user: typescript-eslint/typescript-eslint#11528 (not planned) ----
		// The maintainers confirmed this specific shape is intentional: TData
		// used as a direct generic argument of the *return type* (IError<TData>)
		// counts as "used multiple times" via the AST-level direct-generic-
		// argument shortcut, even though it's spelled only once. See the
		// invalid case below for the sibling shape from the same thread that
		// the maintainers confirmed SHOULD still be reported (TData used only
		// in the return type itself, with no other signature position).
		{Code: `
interface IError<TData> {
  error: true;
  data: TData;
}

declare const checkIsError: <TData>(anything: unknown) => anything is IError<TData>;

const extractErrorData = <TData,>(error?: unknown): IError<TData> | null => {
  return checkIsError<TData>(error) ? error : null;
};
`},

		// ---- Branch lock-in: isTypeParameterRepeatedInAST self-reference exclusion ----
		// Locks in upstream isTypeParameterRepeatedInAST() arm: a reference to
		// T inside T's own constraint clause doesn't count toward repetition,
		// so this needs the type-checker phase; T's constraint (Base<T>) and
		// its use as a direct generic argument there both correctly resolve
		// to "used multiple times" once the checker phase runs.
		{Code: `
interface Base<T> {
  value: T;
}
declare function recursive<T extends Base<T>>(x: T): T;
`},

		// ---- Branch lock-in: countTypeParameterUsage() KindConstructor arm ----
		// A class's constructor is a member visited with fromClass=true; the
		// special functionLikeType branch (visitSignature via
		// GetSignatureFromDeclaration) must fire for it rather than
		// GetTypeAtLocation, or the class's own T count comes out too low.
		{Code: `
class Box<T> {
  value: T;
  constructor(value: T) {
    this.value = value;
  }
}
`},

		// ---- Branch lock-in: collectTypeParameterUsageCounts() readonly array vs mutable array ----
		{Code: `declare function makeReadonlyArrayParam<T>(input: readonly T[]): T;`},
		{Code: `
// Mutable array used as a parameter (not a return type) counts as single-use,
// matching upstream's isReturnType-gated special case for Array/ReadonlyArray.
declare function identityArray<T>(input: T[]): T[];
`},

		// ---- Branch lock-in: generic type-alias reference (type.aliasTypeArguments) ----
		{Code: `
type Wrapper<T> = { value: T };
declare function wrap<T>(value: T): Wrapper<T>;
`},

		// ---- Dimension 4: private identifier / computed key on a class ----
		{Code: `
class Container<T> {
  #value: T;
  constructor(value: T) {
    this.#value = value;
  }
  get(): T {
    return this.#value;
  }
}
`},

		// ---- Dimension 4: polymorphic `this` type in a member signature ----
		// A `this` type carries the same type-parameter type flag as a real type
		// parameter, but its symbol is the declaring interface, so its first
		// declaration is not a type parameter declaration. It must be skipped
		// rather than blindly treated as one. Here T is used twice, so nothing
		// is reported; see the invalid cases for the single-use counterpart.
		{Code: `
interface Box<T> {
  value: T;
  take(value: T): void;
  self(): this;
}
declare function f<T>(b: Box<T>): void;
`},

		// ---- Branch lock-in: canHaveTypeArgumentsList() TypeQuery arm ----
		// `typeof g<T>` attaches its type arguments to a TypeQuery rather than a
		// TypeReference; T must still register as a direct generic-argument use.
		{Code: `
declare function g<U>(u: U): U;
declare function f<T>(x: typeof g<T>): void;
`},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: polymorphic `this` type in a class member ----
		// Regression lock-in: `this` types reach the type-parameter branch of the
		// type walk with a class symbol, whose first declaration is a
		// ClassDeclaration. Reporting T here (used only once, in `value: T`)
		// matches ESLint; the `this` return contributes nothing.
		{
			Code: `
class Box<T> {
  value: T;
  self(): this {
    return this;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the class signature.",
				Line:      2,
				Column:    11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class Box {
  value: unknown;
  self(): this {
    return this;
  }
}
`,
				}},
			}},
		},
		// The same shape with an inferred (rather than annotated) `this` return.
		{
			Code: `
class Box<T> {
  value: T;
  self() {
    return this;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class Box {
  value: unknown;
  self() {
    return this;
  }
}
`,
				}},
			}},
		},

		// ---- Real-user: typescript-eslint/typescript-eslint#11528 (not planned) ----
		// The maintainers confirmed this sibling shape from the same thread
		// SHOULD be reported: TData appears only in the return type itself
		// (not as a generic argument of some other type there); a call inside
		// the body passing TData explicitly doesn't count, since body usage
		// is outside the declaring signature.
		{
			Code: `
interface IError<TData> {
  error: true;
  data: TData;
}

declare const checkIsError: <TData>(anything: unknown) => anything is IError<TData>;

const extractErrorData = <TData,>(error?: unknown): TData | null => {
  return checkIsError<TData>(error) ? error.data : null;
};
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      9,
				Column:    27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
interface IError<TData> {
  error: true;
  data: TData;
}

declare const checkIsError: <TData>(anything: unknown) => anything is IError<TData>;

const extractErrorData = (error?: unknown): unknown | null => {
  return checkIsError<unknown>(error) ? error.data : null;
};
`,
				}},
			}},
		},

		// ---- Dimension 4: TS non-null assertion receiver on a reference ----
		{
			Code: `declare function get<T>(): T!;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): unknown!;`,
				}},
			}},
		},

		// ---- Dimension 4: parenthesized receiver, single and multi-level ----
		{
			Code: `declare function get<T>(): (T);`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): (unknown);`,
				}},
			}},
		},
		{
			Code: `declare function get<T>(): ((T));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): ((unknown));`,
				}},
			}},
		},

		// ---- Real-user: typescript-eslint/typescript-eslint#9961 (open) ----
		// Class member usages (get accessor and method) still get reported
		// even though the constrained value is only ever read, matching
		// upstream's acknowledged-but-unfixed class-member false positive.
		{
			Code: `
class ValueClassGetter<T> {
  get value(): T {
    return null!;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the class signature.",
				Line:      2,
				Column:    24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class ValueClassGetter {
  get value(): unknown {
    return null!;
  }
}
`,
				}},
			}},
		},
		{
			Code: `
class ValueClassMethod<T> {
  getValue(): T {
    return null!;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    24,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class ValueClassMethod {
  getValue(): unknown {
    return null!;
  }
}
`,
				}},
			}},
		},

		// ---- Dimension 2: deeply nested (3+ levels) function scoping ----
		// Only the innermost `identity`'s own T is reported; the outer two
		// same-named T's are each used twice in their own signatures (param +
		// return) and must not "bleed" into each other's counts.
		{
			Code: `
function outer<T>(a: T): T {
  function middle<T>(b: T): T {
    function identity<T>(c: T): void {}
    return b;
  }
  return a;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      4,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function outer<T>(a: T): T {
  function middle<T>(b: T): T {
    function identity(c: unknown): void {}
    return b;
  }
  return a;
}
`,
				}},
			}},
		},

		// ---- Dimension 4: graceful degradation — overload signature (body-absent) ----
		{
			Code: `
declare class Store {
  get<T>(key: string): T;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare class Store {
  get(key: string): unknown;
}
`,
				}},
			}},
		},

		// ---- Branch lock-in: buildReplaceWithConstraintFixes() union constraint
		// needs parens when substituted into an intersection position ----
		{
			Code: `
function f<T extends string | number>(x: T & { tag: true }): void {}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function f(x: (string | number) & { tag: true }): void {}
`,
				}},
			}},
		},

		// ---- Branch lock-in: buildReplaceWithConstraintFixes() union constraint
		// needs parens when substituted into an indexed-access position ----
		{
			Code: `
type Lookup = { a: 1; b: 2 };
function f<T extends 'a' | 'b'>(x: Lookup[T]): void {}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
type Lookup = { a: 1; b: 2 };
function f(x: Lookup[('a' | 'b')]): void {}
`,
				}},
			}},
		},

		// ---- Branch lock-in: no constraint at all replaces with "unknown" ----
		{
			Code: `function bare<T>(x: T) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `function bare(x: unknown) {}`,
				}},
			}},
		},
	})
}

// TestNoUnnecessaryTypeParametersEditDemand exercises the four edit-demand
// modes for a representative diagnostic (a sole-use suggestion), asserting
// diagnostic count/message/range are unaffected by demand while the
// suggestion artifact only materializes when requested (the rule has no
// autofix, only a suggestion).
func TestNoUnnecessaryTypeParametersEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"declare function take<T>(param: T): void;\n",
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
					Name:             NoUnnecessaryTypeParametersRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: NoUnnecessaryTypeParametersRule.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoUnnecessaryTypeParametersRule.Run(ctx, nil)
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
	for demand, got := range diagnostics {
		if got := withoutEdits(got[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("diagnostic changed for demand %d:\ngot:  %#v\nwant: %#v", demand, got, want)
		}
		if got[0].FixesPtr != nil {
			t.Errorf("demand %d unexpectedly has autofixes (rule only has suggestions)", demand)
		}
	}

	suggestionOnly := diagnostics[rule.EditDemandSuggestion][0].Suggestions
	allEdits := diagnostics[rule.EditDemandAll][0].Suggestions
	if suggestionOnly == nil || !reflect.DeepEqual(suggestionOnly, allEdits) {
		t.Fatalf("suggestions differ between suggestion-only and all-edits demand")
	}
	if len(*suggestionOnly) != 1 || (*suggestionOnly)[0].Message.Id != "replaceUsagesWithConstraint" {
		t.Fatalf("suggestions = %#v, want 1 replaceUsagesWithConstraint suggestion", *suggestionOnly)
	}
	if diagnostics[rule.EditDemandNone][0].Suggestions != nil ||
		diagnostics[rule.EditDemandAutofix][0].Suggestions != nil {
		t.Errorf("suggestions attached without suggestion demand")
	}
}
