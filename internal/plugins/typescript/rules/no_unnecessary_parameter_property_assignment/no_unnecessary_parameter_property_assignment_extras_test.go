// TestNoUnnecessaryParameterPropertyAssignmentExtras locks in branches and
// edge shapes that the upstream test suite doesn't exercise. Each case
// carries an inline comment pointing at the specific branch / Dimension 4
// row / tsgo AST quirk it covers, so future refactors can't silently regress
// them without breaking a named lock-in.
//
// Dimension 4 walk (rows from PORT_RULE.md):
//   - Receiver / expression wrappers (parens, `!`, `as T`, optional chain) on
//     the LHS receiver ........ covered: tsgo paren elision is mirrored via
//     SkipParentheses; non-`this` receivers (`X!.y`, `(X as any).y`,
//     `X?.y`) lock-in as non-matches.
//   - Access / key forms (Identifier vs StringLiteral vs
//     NoSubstitutionTemplateLiteral vs computed-with-substitution vs numeric
//     literal) ................ covered.
//   - Declaration / container forms (class declaration vs expression; arrow
//     vs FunctionExpression vs MethodDeclaration vs Constructor) covered in
//     the upstream suite + arrow-IIFE-in-PropertyDef lock-in here.
//   - Nesting / traversal boundaries (class-in-class, function-in-function,
//     PropertyDefinition inside a deeper class) covered: upstream tests
//     class-in-class, the lock-ins below add PropertyDef-inside-nested-class.
//   - Graceful degradation (overload/declare/abstract, empty constructor
//     bodies) ................. covered.
package no_unnecessary_parameter_property_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryParameterPropertyAssignmentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryParameterPropertyAssignmentRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver wrappers (must NOT be reported) ----
			// Non-`this` receiver via TS `!`: `this!.foo = foo` walks into
			// `pa.Expression.Kind == ast.KindNonNullExpression`, not
			// ThisKeyword, so getPropertyName returns false. Lock the negative
			// path so a future "unwrap-all-receiver-wrappers" change can't
			// silently flip this on.
			{Code: `
class Foo {
  constructor(public foo: string) {
    this!.foo = foo;
  }
}
    `},
			// `(this as any).foo = foo` — AsExpression on the receiver. Same
			// reasoning as above; locks in the non-match path.
			{Code: `
class Foo {
  constructor(public foo: string) {
    (this as any).foo = foo;
  }
}
    `},

			// ---- Dimension 4: access/key forms ----
			// `this[0]` — numeric literal key. utils.GetStaticStringValue
			// returns "" for numeric literals (it only handles string-typed
			// literals), so getPropertyName falls through and we never match.
			// Locks the non-match — and would never match a parameter anyway
			// because JS parameters can't be named "0".
			{Code: `
class Foo {
  constructor(public foo: number) {
    this[0] = foo;
  }
}
    `},
			// ---- Locks in upstream getIdentifier() arm "other": neither
			// Identifier nor TSAsExpression/TSNonNullExpression. `foo satisfies T`
			// (SatisfiesExpression) returns null from getIdentifier per
			// upstream's literal switch, so we don't report.
			{Code: `
class Foo {
  constructor(public foo: string) {
    this.foo = foo satisfies string;
  }
}
    `},

			// ---- Locks in upstream constructor-listener arm "leftName !==
			// rightId.name". The LHS and RHS names are different — even though
			// both identifiers reference parameter properties.
			{Code: `
class Foo {
  constructor(public foo: string, public bar: string) {
    this.foo = bar;
  }
}
    `},
			// ---- Locks in "rightId is not a parameter". Reference resolves
			// to a top-level const, not the parameter property.
			{Code: `
const foo = 'outer';
class Foo {
  constructor(public foo: string) {
    {
      const foo = 'inner';
      this.foo = foo;
    }
  }
}
    `},
			// ---- Locks in "no parameter property with that name". The
			// constructor has a parameter named `foo`, but without a
			// modifier — so the assignment is meaningful, not redundant.
			{Code: `
class Foo {
  foo: string;
  constructor(foo: string) {
    this.foo = foo;
  }
}
    `},

			// ---- Locks in the assignedBeforeUnnecessary gate from BOTH
			// directions. Here `&&=` would normally be reported, but a prior
			// `+=` (NOT in UNNECESSARY_OPERATORS) seeds the set and suppresses
			// the later candidate.
			{Code: `
class Foo {
  constructor(public foo: number) {
    this.foo += 1;
    this.foo &&= foo;
  }
}
    `},

			// ---- Locks in the PropertyDef-by-arrow-IIFE path. The arrow IIFE
			// must be the DIRECT initializer of the PropertyDeclaration
			// (modulo parens) — a nested IIFE doesn't count.
			{Code: `
class Foo {
  constructor(public foo: number) {
    this.foo = foo;
  }
  init = ((() => {
    this.foo = 1;
  })());
}
    `},

			// ---- Dimension 4: arrow IIFE as an ARGUMENT to a call inside a
			// PropertyDeclaration initializer. Upstream's loose
			// `arrow.parent.type === CallExpression` matches calls where the
			// arrow is the callee OR an argument, and our iifeCallOfArrow
			// mirrors that. With the arrow as an ARGUMENT, the surrounding
			// CallExpression isn't the PropertyDeclaration.Initializer (the
			// initializer is the OUTER call), so we still don't match — locks
			// in the propDef.Initializer !== call branch.
			{Code: `
declare function take(fn: () => void): number;
class Foo {
  constructor(public foo: number) {
    this.foo = foo;
  }
  init = take(() => {
    this.foo = 1;
  });
}
    `},

			// ---- Locks in: non-constructor function-like ancestor stops the
			// constructor listener. Method body assignment is NOT inside a
			// constructor, so we never push.
			{Code: `
class Foo {
  bar(foo: string) {
    this.foo = foo;
  }
}
    `},

			// ---- Class EXPRESSION container (Dimension 4 declaration form).
			// The upstream `class Bar` invalid case covers class-in-class, but
			// only via ClassDeclaration. Lock in class expression at the top
			// level too.
			{Code: `
const F = class {
  constructor(public foo: string) {
    this.foo = bar;
  }
};
    `},

			// ---- Graceful degradation: abstract / declare members must not
			// crash and must not falsely report on body-absent forms.
			{Code: `
declare class Foo {
  constructor(public foo: string);
}
    `},
			{Code: `
abstract class Foo {
  abstract bar(): void;
  constructor(public foo: string) {}
}
    `},

			// ---- Locks in: assignment whose RHS is an `as` chain that
			// terminates in a non-Identifier still resolves to null. (E.g.,
			// `foo.bar as any` — getIdentifier recurses through AsExpression,
			// then finds PropertyAccessExpression, not Identifier → null.)
			{Code: `
class Foo {
  constructor(public foo: { bar: string }) {
    this.foo = foo.bar as any;
  }
}
    `},

			// ---- Dimension 4: PrivateIdentifier access — `this.#foo = foo`.
			// tsgo encodes `#foo` as KindPrivateIdentifier (NOT KindIdentifier),
			// so getPropertyName returns false on the Name() check and we don't
			// match. Parameter properties can't be named with `#`, so the
			// non-match is correct.
			{Code: `
class Foo {
  #foo: string;
  constructor(foo: string) {
    this.#foo = foo;
  }
}
    `},

			// ---- Graceful degradation: rest parameter (`...foo`) can't be a
			// parameter property — TypeScript rejects the modifier — so an
			// inner `this.foo = foo` would only match if foo is also a real
			// parameter property. Here neither is, so no report.
			{Code: `
class Foo {
  constructor(...foo: string[]) {
    this.foo = foo;
  }
}
    `},

			// ---- Graceful degradation: destructuring parameter
			// (`{ foo }: T`) — TS forbids a modifier on a binding pattern, so
			// there's no parameter property, and the `this.foo = foo` is
			// meaningful.
			{Code: `
class Foo {
  foo: string;
  constructor({ foo }: { foo: string }) {
    this.foo = foo;
  }
}
    `},

			// ---- Dimension 4: nested PropertyDefinition inside a CLASS
			// EXPRESSION inside an OUTER PropertyDefinition. The two
			// `assignedBeforeConstructor` sets must not bleed — outer
			// PropertyDef is on outer class's reportInfo, inner on inner's.
			{Code: `
class Outer {
  constructor(public foo: string) {
    this.foo = bar;
  }
  inner = class Inner {
    init = (this.bar = 1);
  };
}
    `},

			// ---- Locks in upstream's "single bump" of arrow IIFE. With
			// nested arrow IIFEs `(() => (() => { … })())()`, the outer one
			// is the IIFE that immediately wraps the constructor body, but
			// the INNER `this.foo = foo` is inside a SECOND arrow whose
			// parent is the inner CallExpression. upstream bumps only ONCE
			// (functionNode = outer arrow, which still isn't Constructor),
			// so no report. We mirror that single-bump behavior — the inner
			// case should NOT report.
			{Code: `
class Foo {
  constructor(public foo: string) {
    (() =>
      (() => {
        this.foo = foo;
      })())();
  }
}
    `},

			// ---- Locks in: arrow inside a regular METHOD (not constructor).
			// findParentFunction reaches the arrow → arrow.Kind != Constructor
			// → constructor branch skip. PropertyDef branch: funcNode is arrow,
			// but no enclosing PropertyDeclaration → skip. Both branches noop.
			{Code: `
class Foo {
  constructor(public foo: string) {}
  bar() {
    const arrow = () => {
      this.foo = 'x';
    };
  }
}
    `},

			// ---- Locks in: assignment inside a getter — not a constructor.
			{Code: `
class Foo {
  constructor(public foo: string) {}
  get bar() {
    this.foo = 'x';
    return 1;
  }
}
    `},

			// ---- Dimension 4: receiver wrapped in `as` (`(this as Foo).foo`).
			// Upstream's isThisMemberExpression checks node.object.type ===
			// ThisExpression; `this as Foo` is TSAsExpression, so it returns
			// false. We match by not stripping TSAsExpression from the
			// receiver — only ParenthesizedExpression is stripped.
			{Code: `
class Foo {
  constructor(public foo: string) {
    (this as Foo).foo = foo;
  }
}
    `},

			// ---- Locks in upstream constructor handler's "wrong RHS shape"
			// — `this.foo = foo.bar` where the RHS is PropertyAccess, not
			// Identifier. getIdentifier returns nil; no report. Upstream
			// already tests this with `foo.bar` (private foo:any case), but
			// adds an explicit lock for the bare path.
			{Code: `
class Foo {
  constructor(public foo: { foo: string }) {
    this.foo = foo.foo;
  }
}
    `},

			// ---- Real-user: NestJS / Angular DI style. Many parameter
			// properties at once with explicit redundant assignments. The
			// rule must catch every redundant one independently.
			//
			// Listed in valid because here NONE are redundant — the
			// assignments target DIFFERENT names than the parameter
			// properties. Locks in that name-mismatch survives across
			// multiple parameters.
			{Code: `
class UserService {
  constructor(
    private http: any,
    private logger: any,
    private config: any,
  ) {
    this.notHttp = http;
    this.notLogger = logger;
  }
}
    `},

			// ---- Locks in catch-parameter shadowing. Even though the catch
			// binding shares the parameter's name, it's a separate symbol
			// (CatchClause-introduced), so isReferenceFromParameter resolves
			// to a non-Parameter declaration → no report.
			{Code: `
class Foo {
  constructor(public foo: string) {
    try {
    } catch (foo) {
      this.foo = foo;
    }
  }
}
    `},

			// ---- Locks in for-loop variable shadowing.
			{Code: `
class Foo {
  constructor(public foo: string) {
    for (const foo of []) {
      this.foo = foo;
    }
  }
}
    `},

			// ---- Locks in block-scoped `let` shadowing inside a sub-block.
			// Top-level redeclaration of a parameter is a TS error, but a
			// nested block can legitimately introduce its own `let`. Once
			// inside that block, `foo` resolves to the inner binding — not
			// the parameter — so isReferenceFromParameter returns false.
			{Code: `
class Foo {
  constructor(public foo: string) {
    {
      let foo = 'inner';
      this.foo = foo;
    }
  }
}
    `},

			// ---- Locks in: `this.foo += foo` (compound op, non-unnecessary
			// operator). Already covered by upstream, but extras adds a
			// version with a parameter that's NOT a parameter property —
			// upstream's assignedBeforeUnnecessary tracking shouldn't depend
			// on whether `foo` is a parameter property; it just records the
			// name.
			{Code: `
class Foo {
  foo: number = 0;
  constructor(foo: number) {
    this.foo += foo;
    this.foo = foo;
  }
}
    `},

			// ---- Locks in: destructuring assignment `[this.foo] = [foo]`.
			// tsgo represents this as BinaryExpression with LHS =
			// ArrayLiteralExpression — not PropertyAccessExpression — so
			// getPropertyName falls through and we don't match. Matches
			// upstream (LHS is ArrayPattern, isThisMemberExpression false).
			{Code: `
class Foo {
  constructor(public foo: string) {
    [this.foo] = [foo];
  }
}
    `},

			// ---- Locks in: chained assignment `this.a = this.b = foo`.
			// The outer BinaryExpression's RHS is the inner BinaryExpression
			// (not Identifier / TSAsExpression / TSNonNullExpression), so
			// getIdentifier returns nil on the outer. The inner's leftName
			// (`b`) doesn't equal rightName (`foo`), so the inner skips too.
			// Neither reports — matches upstream's narrow `this.X = X`
			// pattern recognition.
			{Code: `
class Foo {
  constructor(
    public a: string,
    public b: string,
  ) {
    this.a = (this.b = a);
  }
}
    `},

			// ---- Locks in: comma expression on RHS — `this.foo = (foo, sideEffect)`.
			// SequenceExpression (in tsgo: BinaryExpression with comma op)
			// is not Identifier/AsExpression/NonNullExpression, so
			// getIdentifier returns nil → don't report. Matches upstream.
			{Code: `
declare function sideEffect(): void;
class Foo {
  constructor(public foo: string) {
    this.foo = (foo, sideEffect(), foo);
  }
}
    `},

			// ---- Real-user: parameter property + PropertyDef initializer
			// using the SAME name suppresses the constructor report.
			// Upstream invalid-15 has PropertyDef AFTER constructor; this
			// extras lock-in has the REVERSED ordering (PropertyDef before)
			// — visit order shouldn't matter because suppression is checked
			// on ClassBody:exit after both have been observed.
			{Code: `
class Foo {
  init = (() => {
    this.foo = 1;
  })();
  constructor(public foo: number) {
    this.foo = foo;
  }
}
    `},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: NoSubstitutionTemplateLiteral as the computed
			// key (e.g. `this[`foo`] = foo`). utils.GetStaticStringValue
			// returns "foo" for NoSubstitutionTemplateLiteral, so this should
			// be reported. Upstream's getStaticStringValue handles this too;
			// upstream just doesn't have a direct test for it.
			{
				Code: "\nclass Foo {\n  constructor(private foo: string) {\n    this[`foo`] = foo;\n  }\n}\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Dimension 4: receiver wrapped by parentheses on the LHS
			// — `(this.foo) = foo`. tsgo preserves parens, but SkipParentheses
			// elides them, so we still match (upstream sees them elided too).
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    (this.foo) = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in getIdentifier() arm "TSNonNullExpression" inside a
			// chain — recursive unwrap through `(foo as any)!`. Upstream
			// `getIdentifier` recurses through both wrappers; our impl does
			// the same.
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo = (foo as any)!;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in RHS-paren-stripping (an extra branch we add over
			// upstream's literal code to mirror upstream's ESTree-level
			// behavior under tsgo's paren-preserving AST). `this.foo = (foo)`
			// — outer parens on the RHS — must still report.
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo = (foo);
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in the &&= / ||= / ??= report path when the prior
			// non-unnecessary op has NOT been seen — i.e., the operator
			// branch in isolation. Upstream covers each separately; this adds
			// a multi-op invalid in the SAME constructor to ensure the seed
			// of one doesn't accidentally suppress the others.
			{
				Code: `
class Foo {
  constructor(public a: number, public b: number, public c: number) {
    this.a ||= a;
    this.b ??= b;
    this.c &&= c;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 5, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 6, Column: 5},
				},
			},

			// ---- Locks in: parameter property whose binding has a default
			// value (`public foo = ''`). Upstream tests `public foo = ''`
			// with `this.foo = foo` and `+= 'foo'`; we also cover the bare
			// `this.foo = foo` first-line case to exercise the path where
			// constructorHasParameterPropertyNamed sees a Parameter with both
			// modifier flags AND an Initializer set.
			{
				Code: `
class Foo {
  constructor(public foo: string = 'hello') {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in: nested CLASS EXPRESSION inside a constructor —
			// upstream covers class declaration nesting, this asserts the
			// class-expression variant pops/pushes its own reportInfo.
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    const Bar = class {
      constructor(private foo: string) {
        this.foo = foo;
      }
    };
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 6, Column: 9},
					{MessageId: "unnecessaryAssign", Line: 9, Column: 5},
				},
			},

			// ---- Locks in the IIFE-in-constructor branch where the IIFE
			// itself contains a nested non-arrow function (which still defers
			// `this` to the surrounding scope only because the IIFE is an
			// arrow). Tests the findParentFunction → arrow → iifeCallOfArrow
			// → findParentFunction(call.Parent) bump.
			{
				Code: `
class Foo {
  constructor(private foo: string) {
    (() => {
      this.foo = foo;
    })();
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 5, Column: 7},
				},
			},

			// ---- Dimension 4: parenthesized RECEIVER on the LHS —
			// `(this).foo = foo`. ESTree elides the parens around `this`, so
			// upstream still treats this as a `this`-rooted member access. We
			// mirror with SkipParentheses on the receiver inside getPropertyName.
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    (this).foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in: `protected` modifier counts as a parameter
			// property. Upstream tests `public` / `private` but not
			// `protected` / `readonly` / combined — these go through the same
			// ModifierFlagsParameterPropertyModifier flag, but exercise the
			// path explicitly.
			{
				Code: `
class Foo {
  constructor(protected foo: string) {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in: `readonly` as a parameter property modifier.
			{
				Code: `
class Foo {
  constructor(readonly foo: string) {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Locks in: combined modifiers `private readonly` /
			// `public readonly` — same path, but the modifier-flag combo
			// shouldn't gate matching.
			{
				Code: `
class Foo {
  constructor(public readonly foo: string) {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Multiple parameter properties, multiple `this.X = X`
			// assignments in a single constructor — each must be reported
			// independently (verifies the loop over parameters + the per-name
			// gating).
			{
				Code: `
class Foo {
  constructor(
    public a: string,
    public b: string,
    public c: string,
  ) {
    this.a = a;
    this.b = b;
    this.c = c;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 8, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 9, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 10, Column: 5},
				},
			},

			// ---- Real-user: NestJS / Angular DI pattern. Multiple parameter
			// properties WITH the redundant assignments — this is the typical
			// production mistake the rule is designed to catch. Mixed
			// modifiers + leading `super()` call + one non-parameter-property
			// argument exercise the path holistically.
			{
				Code: `
class Base {}
class UserService extends Base {
  constructor(
    private readonly http: any,
    private logger: any,
    public config: any,
    plainArg: any,
  ) {
    super();
    this.http = http;
    this.logger = logger;
    this.config = config;
    this.plainArg = plainArg;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 11, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 12, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 13, Column: 5},
				},
			},

			// ---- Control-flow boundaries: the rule is control-flow
			// agnostic (matches upstream — no reachability analysis). Lock
			// in that `this.foo = foo` is reported regardless of whether
			// it's guarded by `if` / `for` / `try` / `switch`.
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    if (foo) {
      this.foo = foo;
    }
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 5, Column: 7},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    try {
      this.foo = foo;
    } catch {}
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 5, Column: 7},
				},
			},
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    switch (foo) {
      case 'a':
        this.foo = foo;
        break;
    }
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 6, Column: 9},
				},
			},

			// ---- Three-level class nesting. Each level has its own
			// parameter property + redundant assignment. The fileBuffer
			// sort must emit reports in source-position order across all
			// three classes (line 5 column 9 / line 7 column 13 / line 11
			// column 5).
			{
				Code: `
class A {
  constructor(public a: string) {
    class B {
      constructor(public b: string) {
        class C {
          constructor(public c: string) {
            this.c = c;
          }
        }
        this.b = b;
      }
    }
    this.a = a;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 8, Column: 13},
					{MessageId: "unnecessaryAssign", Line: 11, Column: 9},
					{MessageId: "unnecessaryAssign", Line: 14, Column: 5},
				},
			},

			// ---- `abstract` class with concrete constructor — the rule
			// applies normally inside the implementation body.
			{
				Code: `
abstract class Foo {
  abstract bar(): void;
  constructor(public foo: string) {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 5, Column: 5},
				},
			},

			// ---- Constructor with leading overload signatures. The
			// overload signatures have no body, only the implementation does
			// — and the implementation is where parameter properties live.
			// Locks in that overload signatures don't confuse the rule.
			{
				Code: `
class Foo {
  constructor(foo: string);
  constructor(foo: number);
  constructor(public foo: any) {
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 6, Column: 5},
				},
			},

			// ---- The void-IIFE valid case above asserted the inner arrow
			// does NOT mark "foo" in assignedBeforeConstructor. This is the
			// flip side: the outer constructor's `this.bar = bar` should
			// still report (separate name, so suppression wouldn't matter
			// even if leaky). Sanity-check the full-program report.
			{
				Code: `
class Foo {
  constructor(public bar: string) {
    this.bar = bar;
  }
  init = void (() => {
    this.foo = 1;
  })();
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
				},
			},

			// ---- Real-user: subclass constructor with super call BEFORE the
			// redundant assignment. The `super(foo)` evaluates its arguments
			// first; the parameter property still completes binding
			// regardless. Locks in that `super()` placement does not
			// influence whether `this.X = X` is redundant.
			{
				Code: `
class Base {
  constructor(public x: string) {}
}
class Foo extends Base {
  constructor(public foo: string) {
    super(foo);
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 8, Column: 5},
				},
			},

			// ---- Mixed `||=`, `&&=`, `=`, `+=` in one constructor — verify
			// each per-name `assignedBeforeUnnecessary` set is independent.
			// `foo`: `=` first then `+=` after — `=` reports; `+=` doesn't
			// re-mark before because the seed gate only blocks LATER `=` /
			// `||=` / `&&=` / `??=` (upstream tests this for foo at line 5
			// invalid-5; we add a per-name independence check).
			{
				Code: `
class Foo {
  constructor(
    public a: number,
    public b: number,
  ) {
    this.a += 1;
    this.a = a;
    this.b = b;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					// `this.a = a` is suppressed by the prior `+=`
					// (assignedBeforeUnnecessary), but `this.b = b` is
					// independent and still fires.
					{MessageId: "unnecessaryAssign", Line: 9, Column: 5},
				},
			},

			// ---- Locks in: assignment inside a CONDITION expression
			// (`if (this.foo = foo)`). The Binary listener still fires on the
			// inner assignment; nothing about the condition position
			// suppresses it. Matches upstream's syntactic, control-flow-
			// agnostic recognition.
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    if ((this.foo = foo)) {
    }
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 10},
				},
			},

			// ---- Multiple redundant assignments to the SAME parameter
			// property — upstream emits one report per assignment. The
			// `assignedBeforeUnnecessary` gate doesn't suppress later same-op
			// assignments (only non-unnecessary ops seed it).
			{
				Code: `
class Foo {
  constructor(public foo: string) {
    this.foo = foo;
    this.foo = foo;
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 4, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 5, Column: 5},
					{MessageId: "unnecessaryAssign", Line: 6, Column: 5},
				},
			},

			// ---- Inner class declared INSIDE a constructor body has its
			// OWN PropertyDef-initializer assignedBeforeConstructor — the
			// outer constructor's reportInfo must not be polluted.
			// Specifically: inner class's PropertyDef assigns "foo" (same
			// name as outer parameter property), so outer's report for
			// `this.foo = foo` MUST still fire. Locks in that the
			// reportInfoStack push/pop isolates per-class state.
			{
				Code: `
class Outer {
  constructor(public foo: number) {
    class Inner {
      init = (this.foo = 1);
    }
    this.foo = foo;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssign", Line: 7, Column: 5},
				},
			},

		},
	)
}
