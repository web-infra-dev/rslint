package forbid_foreign_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestForbidForeignPropTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidForeignPropTypesRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{
			Code: `import { propTypes } from "SomeComponent";`,
			Tsx:  true,
		},
		{
			Code: `import { propTypes as someComponentPropTypes } from "SomeComponent";`,
			Tsx:  true,
		},
		{
			Code: `const foo = propTypes`,
			Tsx:  true,
		},
		{
			Code: `foo(propTypes)`,
			Tsx:  true,
		},
		{
			Code: `foo + propTypes`,
			Tsx:  true,
		},
		{
			Code: `const foo = [propTypes]`,
			Tsx:  true,
		},
		// Shorthand `{ propTypes }` in an ObjectLiteralExpression — NOT a
		// destructuring pattern, so the ObjectBindingPattern listener does
		// not fire. Locked in here.
		{
			Code: `const foo = { propTypes }`,
			Tsx:  true,
		},
		// LHS of dotted assignment — `isAssignmentLHS` excludes.
		{
			Code: `Foo.propTypes = propTypes`,
			Tsx:  true,
		},
		// LHS of bracket assignment — `isAssignmentLHS` excludes too.
		{
			Code: `Foo["propTypes"] = propTypes`,
			Tsx:  true,
		},
		// Computed access whose argument is a non-literal Identifier — does
		// NOT match upstream's `Literal` branch. Locked in.
		{
			Code: `const propTypes = "bar"; Foo[propTypes];`,
			Tsx:  true,
		},
		// allowInPropTypes: true + assignment-context propTypes use.
		{
			Code: `
        const Message = (props) => (<div>{props.message}</div>);
        Message.propTypes = {
          message: PropTypes.string
        };
        const Hello = (props) => (<Message>Hello {props.name}</Message>);
        Hello.propTypes = {
          name: Message.propTypes.message
        };
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},
		// allowInPropTypes: true + class-field-context propTypes use.
		{
			Code: `
        class MyComponent extends React.Component {
          static propTypes = {
            baz: Qux.propTypes.baz
          };
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},

		// ---- Universal edge shapes (Dimension 4) ----
		// Template-literal computed access — upstream does NOT match
		// TemplateLiteral, only Literal. Lock-in: locks upstream's
		// `property.type === 'Literal'` arm — without this case a future
		// refactor that broadens to NoSubstitutionTemplateLiteral would
		// silently drift. (Goes here, not invalid, because upstream
		// passes too.)
		{
			Code: "Foo[`propTypes`]",
			Tsx:  true,
		},
		// PrivateIdentifier `#propTypes` — upstream's
		// `property.type === 'Identifier'` excludes PrivateIdentifier.
		// Lock-in for the type-check we make on `KindIdentifier`.
		{
			Code: `class C { #propTypes = 1; m() { return this.#propTypes } }`,
			Tsx:  true,
		},
		// Computed access whose argument is a numeric literal — not 'propTypes'.
		{
			Code: `Foo[0]`,
			Tsx:  true,
		},
		// Property name "propTypes" used as a *string-literal key* in an
		// ObjectBindingPattern. Upstream's `'name' in property.key` check
		// excludes string-literal keys. Lock-in for the Identifier-only
		// destructuring gate.
		{
			Code: `var { "propTypes": x } = SomeComponent;`,
			Tsx:  true,
		},
		// Rest element `{ ...propTypes }` — upstream's
		// `property.type === 'Property'` filter excludes RestElement, so
		// the rest binding `propTypes` is not reported. Locked in via the
		// `DotDotDotToken != nil` skip in the listener.
		{
			Code: `var { ...propTypes } = SomeComponent;`,
			Tsx:  true,
		},
		// LHS of compound assignment — upstream `isAssignmentLHS` matches
		// any AssignmentExpression, regardless of operator. `+=` etc.
		// should NOT report on the LHS. Lock-in for `IsAssignmentOperator`
		// covering compound forms.
		{
			Code: `Foo.propTypes += {}`,
			Tsx:  true,
		},
		// Parenthesized LHS `(Foo.propTypes) = X` — ESTree flattens the
		// parens so upstream sees the property as the assignment LHS and
		// does NOT report. tsgo preserves the ParenthesizedExpression;
		// `effectiveParent` inside `isAssignmentLHS` walks through it so
		// we behave the same.
		{
			Code: `(Foo.propTypes) = {};`,
			Tsx:  true,
		},
		// Foreign propTypes used inside an arrow function body in a
		// propTypes assignment — closest enclosing AssignmentExpression
		// has LHS `Bar.propTypes`, so allowInPropTypes:true excuses it.
		// Locks in the parent walk through nested function bodies.
		{
			Code: `
        Bar.propTypes = {
          x: () => Foo.propTypes
        };
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},
		// Same as above, but for the class-field branch — the parent walk
		// must traverse a method/getter body without losing track of the
		// enclosing `static propTypes` field.
		{
			Code: `
        class C {
          static propTypes = {
            x: () => Foo.propTypes
          };
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},
		// Object destructuring with key `'propTypes'` written with a
		// computed-string literal `[...]` — upstream's `'name' in
		// property.key` check excludes ComputedPropertyName too.
		{
			Code: `var { ["propTypes"]: x } = SomeComponent;`,
			Tsx:  true,
		},
		// Object destructuring assignment where the KEY is a non-propTypes
		// identifier and the VALUE local is named `propTypes`. Upstream
		// gates on KEY.name === 'propTypes', so this is fine.
		{
			Code: `var { foo: propTypes } = SomeComponent;`,
			Tsx:  true,
		},

		// ---- TS type-level positions (NOT MemberExpression) ----
		// `typeof Foo.propTypes` in a type alias — tsgo represents the
		// type-level dotted name as `KindQualifiedName`, NOT
		// `KindPropertyAccessExpression`, so the listener correctly does
		// not fire. Mirrors @typescript-eslint/parser emitting
		// `TSQualifiedName` (also not MemberExpression).
		{
			Code: `type X = typeof Foo.propTypes;`,
			Tsx:  true,
		},
		{
			Code: `type X = { p: typeof Foo.propTypes };`,
			Tsx:  true,
		},
		// ---- allowInPropTypes: true + deeply nested foreign access ----
		// HOC pattern: spread of foreign propTypes inside own propTypes
		// assignment.
		{
			Code: `
        function withFoo(C) {
          C.propTypes = { ...Inner.propTypes };
          return C;
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},
		// Two-level nested function bodies — parent walk must traverse
		// both function expressions to find the enclosing assignment.
		{
			Code: `
        Foo.propTypes = {
          x: function() {
            return { y: function() { return Bar.propTypes; } };
          }
        };
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},
		// Conditional / logical expressions inside propTypes assignment.
		{
			Code: `
        Foo.propTypes = (cond ? Bar.propTypes : Baz.propTypes) || {};
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},
		// Same idea for class field — spread + conditional inside the
		// `static propTypes` initializer.
		{
			Code: `
        class C {
          static propTypes = { ...Other.propTypes, x: cond ? Inner.propTypes.y : null };
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
		},

		// ---- Regular ObjectLiteralExpression must NOT fire ----
		// A plain object containing a key called `propTypes` is not
		// destructuring; the OLE listener gates on `IsAssignmentTarget`.
		{
			Code: `const x = { propTypes: 1, foo: 2 };`,
			Tsx:  true,
		},
		// Nested OLE that is NEITHER an assignment target — same rule.
		{
			Code: `const x = { y: { propTypes: 1 } };`,
			Tsx:  true,
		},

		// ---- Bracket access with non-string literal arguments (NOT match) ----
		// Numeric separator literal — `KindNumericLiteral`, not `KindStringLiteral`.
		{
			Code: `Foo[1_000];`,
			Tsx:  true,
		},
		// BigInt literal — `KindBigIntLiteral`, not `KindStringLiteral`.
		{
			Code: `Foo[1n];`,
			Tsx:  true,
		},

		// ---- Destructuring with computed template-literal key (NOT match) ----
		// Upstream `'name' in property.key` excludes ComputedPropertyName
		// regardless of what's inside; lock both common forms.
		{
			Code: "var { [`propTypes`]: x } = SomeComponent;",
			Tsx:  true,
		},

		// ---- Function rest param destructuring (NOT match) ----
		// `function f({...propTypes})` — propTypes is the REST target,
		// not a regular key. `findPropTypesKey` skips `DotDotDotToken != nil`.
		{
			Code: `function f({ ...propTypes }) {}`,
			Tsx:  true,
		},

		// ---- Case-sensitivity ----
		// Exact-match only — `PropTypes` (uppercase P) is a different name.
		{
			Code: `Foo.PropTypes;`,
			Tsx:  true,
		},
		{
			Code: `Foo.proptypes;`,
			Tsx:  true,
		},
		{
			Code: `Foo["PropTypes"];`,
			Tsx:  true,
		},
		{
			Code: `var { PropTypes } = SomeComponent;`,
			Tsx:  true,
		},

		// ---- defaultProps / contextTypes / similar (NOT match) ----
		// The rule is specific to `propTypes`. Other React static fields
		// must not be reported.
		{
			Code: `Foo.defaultProps;`,
			Tsx:  true,
		},
		{
			Code: `Foo.contextTypes;`,
			Tsx:  true,
		},

		// ---- JSX dotted tag (NOT match) ----
		// `<Foo.propTypes />` and `<Foo.propTypes></Foo.propTypes>` —
		// tsgo represents dotted JSX tags as `PropertyAccessExpression`,
		// but ESTree's `JSXMemberExpression` is a DIFFERENT node type.
		// ESLint's `MemberExpression` listener does NOT fire on JSX
		// dotted tags; the `isJsxTagName` skip aligns us byte-for-byte.
		{
			Code: `var X = <Foo.propTypes />;`,
			Tsx:  true,
		},
		{
			Code: `var X = <Foo.propTypes></Foo.propTypes>;`,
			Tsx:  true,
		},
		// Chained dotted JSX tag — `<Foo.propTypes.Bar />`. The walk-up
		// in `isJsxTagName` follows the PA chain to its outermost link.
		{
			Code: `var X = <Foo.propTypes.Bar />;`,
			Tsx:  true,
		},
		// JSX dotted tag must not affect the rule when propTypes appears
		// inside an attribute value of the same element — that PA is
		// inside a JsxExpression, not the TagName position.
		// (Inverse already covered in invalid: `<Foo p={Bar.propTypes} />`)

		// ---- typeof in TS type-query (already covered) + value-context typeof ----
		// `typeof Foo.propTypes` as a TYPE QUERY (in a type alias) does
		// NOT fire — see earlier valid block. The inverse — `typeof X.p`
		// in VALUE context — IS a regular MemberExpression and fires;
		// see the invalid block below for the lock-in.
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `
        var Foo = createReactClass({
          propTypes: Bar.propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      3,
					Column:    26,
					EndLine:   3,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
        var Foo = createReactClass({
          propTypes: Bar["propTypes"],
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      3,
					Column:    26,
					EndLine:   3,
					EndColumn: 37,
				},
			},
		},
		{
			Code: `
        var { propTypes } = SomeComponent
        var Foo = createReactClass({
          propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 24,
				},
			},
		},
		{
			Code: `
        var { propTypes: things, ...foo } = SomeComponent
        var Foo = createReactClass({
          propTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 32,
				},
			},
		},
		// Upstream marks this `no-ts` (skipped under TS parsers) due to a
		// TS-mode bug in the parent walk. Our tsgo-based port correctly
		// finds the enclosing `static fooBar = {...}` PropertyDeclaration,
		// confirms it is NOT a `propTypes` field, and reports under
		// default options. Lock-in: tsgo handles the case upstream skips.
		{
			Code: `
        class MyComponent extends React.Component {
          static fooBar = {
            baz: Qux.propTypes.baz
          };
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      4,
					Column:    22,
					EndLine:   4,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
        var { propTypes: typesOfProps } = SomeComponent
        var Foo = createReactClass({
          propTypes: typesOfProps,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 38,
				},
			},
		},
		{
			Code: `
        const Message = (props) => (<div>{props.message}</div>);
        Message.propTypes = {
          message: PropTypes.string
        };
        const Hello = (props) => (<Message>Hello {props.name}</Message>);
        Hello.propTypes = {
          name: Message.propTypes.message
        };
      `,
			Options: map[string]interface{}{"allowInPropTypes": false},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      8,
					Column:    25,
					EndLine:   8,
					EndColumn: 34,
				},
			},
		},
		// Same caveat as case 5 — upstream `no-ts`, tsgo handles correctly.
		{
			Code: `
        class MyComponent extends React.Component {
          static propTypes = {
            baz: Qux.propTypes.baz
          };
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": false},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      4,
					Column:    22,
					EndLine:   4,
					EndColumn: 31,
				},
			},
		},

		// ---- Lock-in tests (universal edge shapes + upstream branch coverage) ----

		// Optional-chain receiver `Foo?.propTypes` — tsgo represents
		// optional chains as a flag on PropertyAccessExpression, not a
		// separate `ChainExpression` wrapper. Locked-in here so a future
		// shape change (e.g. a new wrapper kind) doesn't silently miss
		// the report.
		{
			Code: `Foo?.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    6,
					EndLine:   1,
					EndColumn: 15,
				},
			},
		},
		// Parenthesized receiver `(Foo).propTypes` — tsgo preserves
		// ParenthesizedExpression while ESTree flattens it. The receiver
		// being parenthesized must not affect the property check.
		{
			Code: `(Foo).propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    7,
					EndLine:   1,
					EndColumn: 16,
				},
			},
		},
		// Non-null-assertion receiver `Foo!.propTypes` — TS-only AST
		// shape. Confirms the rule fires on the property regardless of
		// receiver wrappers.
		{
			Code: `declare const Foo: any; Foo!.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 39,
				},
			},
		},
		// Chained access `Foo.propTypes.isRequired` — the inner
		// `Foo.propTypes` fires; the outer `.isRequired` does not. One
		// report total.
		{
			Code: `Foo.propTypes.isRequired;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},
		// allowInPropTypes: true does NOT shield foreign propTypes inside
		// an *unrelated* class field (`static fooBar = ...`). Mirrors
		// upstream `findParentClassProperty` returning the field but its
		// key NOT being `propTypes`.
		{
			Code: `
        class MyComponent extends React.Component {
          static fooBar = {
            baz: Qux.propTypes.baz
          };
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      4,
					Column:    22,
					EndLine:   4,
					EndColumn: 31,
				},
			},
		},
		// allowInPropTypes: true + assignment to NON-propTypes left-hand
		// side — should still report. Locks `assignment.left.property.name
		// === 'propTypes'` arm.
		{
			Code:    `Foo.notPropTypes = { x: Bar.propTypes.x };`,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 38,
				},
			},
		},
		// Bracket-keyed assignment LHS `Foo['propTypes'] = ...` does NOT
		// satisfy upstream's `assignment.left.property.name === 'propTypes'`
		// (the property is a Literal, has no `.name`). With
		// allowInPropTypes: true the rule must still report Bar.propTypes
		// inside the RHS, because the assignment is not recognized as a
		// propTypes-shaped assignment.
		{
			Code:    `Foo['propTypes'] = { x: Bar.propTypes.x };`,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 38,
				},
			},
		},

		// ---- Destructuring-assignment form (`(...) = X`) ----
		// Upstream's `ObjectPattern` listener fires on both var-declaration
		// and assignment forms — ESLint emits the same node kind. tsgo
		// splits them: declarations use `ObjectBindingPattern`, the
		// assignment form keeps the LHS as `ObjectLiteralExpression`. The
		// dedicated OLE listener (gated on `ast.IsAssignmentTarget`) covers
		// these.
		{
			Code: `({ propTypes } = SomeComponent);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    4,
					EndLine:   1,
					EndColumn: 13,
				},
			},
		},
		{
			Code: `({ propTypes: alias } = SomeComponent);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    4,
					EndLine:   1,
					EndColumn: 20,
				},
			},
		},
		// Nested destructuring assignment — the inner OLE is in an
		// assignment-target position via the outer object literal. tsgo's
		// `ast.IsAssignmentTarget` walks up through nested object literals
		// to confirm.
		{
			Code: `({ a: { propTypes } } = SomeComponent);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 18,
				},
			},
		},
		// Array-destructuring containing object-destructuring.
		// `[{propTypes}] = X` requires parens to be parseable in
		// expression position; the inner OLE still sits in an assignment
		// target via the array literal.
		{
			Code: `([{ propTypes }] = [SomeComponent]);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},

		// ---- Other nesting / scoping (Dimension 2) ----
		// Static class block — `findParentClassProperty` returns nil
		// (KindClassStaticBlockDeclaration is not KindPropertyDeclaration),
		// so even `allowInPropTypes: true` does not shield the access.
		{
			Code: `
        class C {
          static block = (() => {
            return Bar.propTypes;
          })();
        }
      `,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenPropType",
					Line:      4,
					Column:    24,
					EndLine:   4,
					EndColumn: 33,
				},
			},
		},
		// Multiple `propTypes` accesses in the same expression — each
		// PropertyAccessExpression / ElementAccessExpression fires the
		// listener independently.
		{
			Code: `var a = Foo.propTypes; var b = Bar["propTypes"];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				{MessageId: "forbiddenPropType", Line: 1, Column: 36, EndLine: 1, EndColumn: 47},
			},
		},

		// ---- JSX context ----
		// PA inside JSX expression child.
		{
			Code: `var X = <div>{Bar.propTypes}</div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
			},
		},
		// PA inside JSX attribute value.
		{
			Code: `var X = <Foo p={Bar.propTypes} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
			},
		},
		// PA inside JSX spread attribute `{...Bar.propTypes}`.
		{
			Code: `var X = <Foo {...Bar.propTypes} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
			},
		},

		// ---- TypeScript receiver / suffix wrappers ----
		// `as`-cast receiver — tsgo: PA(AsExpression(Foo, any), propTypes).
		{
			Code: `(Foo as any).propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 14, EndLine: 1, EndColumn: 23},
			},
		},
		// `satisfies`-receiver — same shape, different wrapper kind.
		{
			Code: `(Foo satisfies any).propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
			},
		},
		// Non-null suffix on the access result `Foo.propTypes!` — the
		// inner PA still fires.
		{
			Code: `Foo.propTypes!;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},
		// Non-null suffix between accesses — `Foo.propTypes!.x` reports
		// only on the inner `propTypes` (outer access is `.x`).
		{
			Code: `Foo.propTypes!.x;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},
		// Optional chain bracket access `Foo?.['propTypes']`.
		{
			Code: `Foo?.['propTypes'];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 7, EndLine: 1, EndColumn: 18},
			},
		},

		// ---- Spread / array / object literal positions ----
		// Spread of foreign propTypes inside an object literal (NOT an
		// assignment target).
		{
			Code: `const x = {...Foo.propTypes};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
			},
		},
		// PA used inside an array literal element.
		{
			Code: `const x = [Foo.propTypes];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			},
		},
		// Spread of foreign propTypes inside an array literal.
		{
			Code: `const x = [...Foo.propTypes];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
			},
		},

		// ---- Function parameter / default contexts ----
		// Default-value initializer in a parameter.
		{
			Code: `function f(x = Foo.propTypes) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 20, EndLine: 1, EndColumn: 29},
			},
		},
		// Object-destructuring parameter pulling out `propTypes`.
		{
			Code: `function f({ propTypes }) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 14, EndLine: 1, EndColumn: 23},
			},
		},
		// Nested destructuring parameter — inner ObjectBindingPattern fires.
		{
			Code: `function f({ a: { propTypes } }) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
			},
		},
		// Destructuring parameter with default — `{propTypes = X}` is
		// still keyed by `propTypes` Identifier.
		{
			Code: `function f({ propTypes = {} }) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 14, EndLine: 1, EndColumn: 28},
			},
		},
		// Array destructuring containing object destructuring (declaration).
		{
			Code: `var [{ propTypes }] = arr;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
			},
		},

		// ---- Class / decorator positions ----
		// `extends` clause uses value-level PA (not type-level).
		{
			Code: `class C extends Foo.propTypes {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
			},
		},
		// Computed method name `[Foo.propTypes]() {}`.
		{
			Code: `class C { [Foo.propTypes]() {} }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			},
		},
		// Class decorator `@Foo.propTypes`.
		{
			Code: `@Foo.propTypes class X {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 6, EndLine: 1, EndColumn: 15},
			},
		},
		// Interface heritage clause — tsgo represents
		// `interface I extends Foo.propTypes {}` with a value-level
		// PropertyAccessExpression in the heritage clause's
		// ExpressionWithTypeArguments (mirroring TypeScript's grammar:
		// `extends` takes a LeftHandSideExpression). Upstream's
		// MemberExpression listener fires equivalently.
		{
			Code: `interface I extends Foo.propTypes {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
			},
		},
		// Class method NAMED `propTypes` — `findParentClassProperty`
		// returns nil (MethodDeclaration is not PropertyDeclaration), so
		// `allowInPropTypes:true` does NOT shield foreign access inside
		// the body. Locks the field-vs-method asymmetry.
		{
			Code:    `class C { static propTypes() { return Bar.propTypes; } }`,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 43, EndLine: 1, EndColumn: 52},
			},
		},
		// Class getter NAMED `propTypes` — same story (GetAccessor is
		// not PropertyDeclaration).
		{
			Code:    `class C { static get propTypes() { return Bar.propTypes; } }`,
			Options: map[string]interface{}{"allowInPropTypes": true},
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 47, EndLine: 1, EndColumn: 56},
			},
		},

		// ---- Tagged template / call / new ----
		// Tagged template: tag is `Foo.propTypes`.
		{
			Code: "Foo.propTypes`x`;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},
		// Call on propTypes — `Foo.propTypes()`.
		{
			Code: `Foo.propTypes();`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},
		// `new` on propTypes — `new Foo.propTypes()`.
		{
			Code: `new Foo.propTypes();`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
			},
		},
		// Generic call — `Foo.propTypes<T>()`.
		{
			Code: `Foo.propTypes<any>();`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},
		// Template-literal substitution `${Foo.propTypes}`.
		{
			Code: "`${Foo.propTypes}`;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
			},
		},

		// ---- Loops / control flow ----
		{
			Code: `for (const k in Foo.propTypes) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
			},
		},
		{
			Code: `for (const v of Foo.propTypes) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
			},
		},
		{
			Code: `while (Foo.propTypes) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 12, EndLine: 1, EndColumn: 21},
			},
		},
		// Async / generator: await + yield expression operands.
		{
			Code: `async function f() { return await Foo.propTypes; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 39, EndLine: 1, EndColumn: 48},
			},
		},
		{
			Code: `function* g() { yield Foo.propTypes; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 27, EndLine: 1, EndColumn: 36},
			},
		},

		// ---- Destructuring assignment edge cases ----
		// Multiple keys with `propTypes` after another property — listener
		// reports the FIRST matching key (upstream uses `.find`).
		{
			Code: `({ foo: a, propTypes } = SomeComponent);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 12, EndLine: 1, EndColumn: 21},
			},
		},
		// Spread (`...rest`) is skipped; `propTypes` shorthand still
		// reports.
		{
			Code: `({ propTypes, ...rest } = SomeComponent);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 4, EndLine: 1, EndColumn: 13},
			},
		},
		// Default initializer in destructuring assignment shorthand —
		// `({propTypes = {}} = X)`. The ShorthandPropertyAssignment node
		// spans `propTypes = {}`.
		{
			Code: `({ propTypes = {} } = SomeComponent);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 4, EndLine: 1, EndColumn: 18},
			},
		},

		// ---- Receiver-type variations (rule is purely syntactic) ----
		// `this.propTypes` — `this` keyword as receiver. Rule fires; the
		// rule does not filter by what the receiver "is".
		{
			Code: `class C { m() { return this.propTypes; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 29, EndLine: 1, EndColumn: 38},
			},
		},
		// `super.propTypes` — `super` keyword as receiver.
		{
			Code: `class C extends D { m() { return super.propTypes; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 40, EndLine: 1, EndColumn: 49},
			},
		},
		// CallExpression receiver — `f().propTypes`. Outer PA fires.
		{
			Code: `f().propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},
		// ConditionalExpression receiver — `(cond ? A : B).propTypes`.
		{
			Code: `(cond ? A : B).propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			},
		},
		// Sequence-expression receiver `(a, B).propTypes` — tsgo represents
		// the comma operator as a `BinaryExpression(KindCommaToken)`.
		{
			Code: `(a, B).propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
			},
		},
		// ElementAccess receiver — `arr[0].propTypes`.
		{
			Code: `arr[0].propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
			},
		},
		// Dynamic import — `import('x').propTypes`. The receiver is a
		// CallExpression with `import` as the keyword tag.
		{
			Code: `import('x').propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
			},
		},

		// ---- Chained / repeated propTypes ----
		// `Foo.propTypes.propTypes` — outer PA's name is also `propTypes`,
		// so BOTH fire (inner reports on inner `propTypes`, outer reports
		// on outer `propTypes`). Mirrors upstream's per-MemberExpression
		// fire. tsgo's listener visits parent-first, so the outer report
		// is emitted before the inner one.
		{
			Code: `Foo.propTypes.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
				{MessageId: "forbiddenPropType", Line: 1, Column: 5, EndLine: 1, EndColumn: 14},
			},
		},

		// ---- ES2021 logical assignment operators on LHS (NOT match) ----
		// These emit BinaryExpression with `||=` / `&&=` / `??=` operator
		// tokens — `ast.IsAssignmentOperator` returns true for all three,
		// so `isAssignmentLHS` correctly excludes the LHS.
		// The RIGHT side is `Bar.propTypes` and DOES fire.
		{
			Code: `Foo.propTypes ||= Bar.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 23, EndLine: 1, EndColumn: 32},
			},
		},
		{
			Code: `Foo.propTypes &&= Bar.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 23, EndLine: 1, EndColumn: 32},
			},
		},
		{
			Code: `Foo.propTypes ??= Bar.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 23, EndLine: 1, EndColumn: 32},
			},
		},

		// ---- Operator / unary positions ----
		// `delete Foo.propTypes` — DeleteExpression with PA operand.
		{
			Code: `delete Foo.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 12, EndLine: 1, EndColumn: 21},
			},
		},
		// `typeof Foo.propTypes` in VALUE context — UnaryExpression /
		// TypeOfExpression with PA operand. Distinct from the type-level
		// `typeof Foo.propTypes` inside `type X = ...` (QualifiedName,
		// covered in valid).
		{
			Code: `var x = typeof Foo.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 20, EndLine: 1, EndColumn: 29},
			},
		},
		// `void Foo.propTypes` — VoidExpression.
		{
			Code: `void Foo.propTypes;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
			},
		},

		// ---- Class expression / nested classes ----
		// Class expression with extends: `const X = class extends Foo.propTypes {}`.
		{
			Code: `const X = class extends Foo.propTypes {};`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 29, EndLine: 1, EndColumn: 38},
			},
		},

		// ---- Catch parameter destructuring ----
		// `try {} catch ({ propTypes }) {}` — catch with destructuring
		// parameter. ObjectBindingPattern listener fires.
		{
			Code: `try {} catch ({ propTypes }) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 17, EndLine: 1, EndColumn: 26},
			},
		},

		// ---- JSX fragment ----
		// `<>{Foo.propTypes}</>` — fragment expression child.
		{
			Code: `var X = <>{Foo.propTypes}</>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			},
		},

		// ---- Spread argument in call ----
		// `f(...Foo.propTypes)` — spread argument.
		{
			Code: `f(...Foo.propTypes);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
			},
		},

		// ---- Decorator on class member ----
		// Member-level decorator on a class field — PA inside decorator
		// fires.
		{
			Code: `class C { @Foo.propTypes prop = 1; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
			},
		},
		// `for (const {propTypes} of arr) {}` — for-of with destructuring
		// in the loop variable position. The ObjectBindingPattern listener
		// fires.
		{
			Code: `for (const { propTypes } of arr) {}`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 14, EndLine: 1, EndColumn: 23},
			},
		},
		// JSX with foreign propTypes in attribute value AND a JSX dotted
		// tag — confirms only the attribute-value PA is reported, the
		// JSX tag PA is correctly skipped via `isJsxTagName`.
		{
			Code: `var X = <NS.Foo p={Bar.propTypes} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenPropType", Line: 1, Column: 24, EndLine: 1, EndColumn: 33},
			},
		},
	})
}
