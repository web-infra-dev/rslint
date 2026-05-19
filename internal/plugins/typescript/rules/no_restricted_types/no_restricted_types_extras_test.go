// TestNoRestrictedTypesExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
package no_restricted_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedTypesRule, []rule_tester.ValidTestCase{
		// ---- Branch lock-in checkBannedTypes() arm 1: bannedType === null disables the ban ----
		{
			Code:    `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": nil}}),
		},
		// ---- Branch lock-in checkBannedTypes() arm 1: bannedType === false disables the ban ----
		{
			Code:    `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": false}}),
		},
		// ---- Branch lock-in: keyword listener is not wired when its name is absent ----
		// bigint is NOT in the types map → the keyword listener is never
		// registered, so an unconfigured primitive keyword stays valid.
		{
			Code:    `let value: bigint;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Branch lock-in TSTupleType arm 2: non-empty tuple does not match "[]" key ----
		// Upstream gates the empty-tuple report on `node.elementTypes.length === 0`.
		// `[number]` has one element, so even though `[]` is banned the rule must stay silent.
		{
			Code:    `let value: [number];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"[]": "Use unknown[] instead."}}),
		},
		// ---- Branch lock-in TSTypeLiteral arm 2: non-empty literal does not match "{}" key ----
		{
			Code:    `let value: { a: number };`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"{}": "Use object instead."}}),
		},
		// ---- Branch lock-in IsClassImplementsOrInterfaceExtends: `class X extends Y` is NOT a heritage match ----
		// Upstream registers TSClassImplements + TSInterfaceHeritage but
		// deliberately omits the class-extends listener. `Banned` here is
		// an expression-position identifier, not a TypeReference, so neither
		// listener fires.
		{
			Code:    `class Derived extends Banned {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Real-user: ban Object in expression-position usage stays silent ----
		// Same as the upstream Object-call valid cases — we re-assert here on
		// a value-position expression that *looks* like a type reference but
		// is parsed as an Identifier expression. The rule must not bleed
		// into value position.
		{
			Code:    `const foo = Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Real-user: shadowing a banned name with a local type alias should still report ----
		// Upstream's rule is *name-string* matching, not symbol-aware — any
		// `Banned` text in a type position reports, even when a local
		// `type Banned = ...` alias would have resolved it.
		// This valid case asserts the opposite: when the type-reference name
		// is NOT in the banned set, the local alias is treated normally.
		{
			Code: `
				type Allowed = string;
				let v: Allowed;
			`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Real-user / Branch lock-in: `any` is not in TYPE_KEYWORDS upstream ----
		// Upstream's TYPE_KEYWORDS deliberately omits `any` (handled by
		// `no-explicit-any` instead). Locking in that banning `any` here
		// has no effect, so a future extension of keywordTypeNames doesn't
		// silently start firing on `any`.
		{
			Code: `let value: any;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"any": map[string]interface{}{
					"message": "Use unknown instead.",
					"suggest": []interface{}{"unknown"},
				},
			}}),
		},
		// ---- Dimension 4 / Receiver wrappers: N/A ----
		// N/A: TypeReference cannot itself be wrapped by `(...)` /
		// non-null / `as` / optional-chain — those are value-expression
		// wrappers. A `(Banned)` in type position is `ParenthesizedType`
		// containing a TypeReference, and the TypeReference listener fires
		// on the inner node — covered by the corresponding invalid case
		// below rather than as a valid no-match case.
	}, []rule_tester.InvalidTestCase{
		// ---- Branch lock-in checkBannedTypes() arm: bannedType === true → empty customMessage ----
		// Already covered for Banned in upstream; this locks in the same
		// shape for an empty type literal.
		{
			Code:    `let value: {};`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"{}": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `{}` as a type.", Line: 1, Column: 12},
			},
		},
		// ---- Branch lock-in getCustomMessage() arm: object without message → empty customMessage ----
		{
			Code: `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok"},
			}}),
			Output: []string{`let value: Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 12},
			},
		},
		// ---- Branch lock-in getCustomMessage() arm: object with empty-string message → empty customMessage ----
		// Upstream's `if (bannedType.message)` falsifies on '' so customMessage stays empty.
		{
			Code: `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"message": ""},
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 12},
			},
		},
		// ---- Branch lock-in checkBannedTypes() arm: object with suggest → emit suggestion(s) ----
		{
			Code: `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{
					"message": "Use Ok instead.",
					"suggest": []interface{}{"Ok"},
				},
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Message:   "Don't use `Banned` as a type. Use Ok instead.",
					Line:      1, Column: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "bannedTypeReplacement", Output: `let value: Ok;`},
					},
				},
			},
		},
		// ---- Branch lock-in: multiple suggestions are emitted in order ----
		{
			Code: `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{
					"message": "Use a real type.",
					"suggest": []interface{}{"OkOne", "OkTwo"},
				},
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Message:   "Don't use `Banned` as a type. Use a real type.",
					Line:      1, Column: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "bannedTypeReplacement", Output: `let value: OkOne;`},
						{MessageId: "bannedTypeReplacement", Output: `let value: OkTwo;`},
					},
				},
			},
		},
		// ---- Branch lock-in: object with BOTH fixWith and suggest emits fix + suggestions ----
		// Upstream `context.report({...fix, suggest})` keeps both; we
		// match by sending the diagnostic via
		// ReportNodeWithFixesAndSuggestions.
		{
			Code: `let value: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{
					"fixWith": "FixedOk",
					"message": "Use Ok instead.",
					"suggest": []interface{}{"SuggestedOk"},
				},
			}}),
			Output: []string{`let value: FixedOk;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Message:   "Don't use `Banned` as a type. Use Ok instead.",
					Line:      1, Column: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "bannedTypeReplacement", Output: `let value: SuggestedOk;`},
					},
				},
			},
		},
		// ---- Branch lock-in TSTypeReference arm 2: typeArguments → also check whole-node text ----
		// Same shape as `Banned<any>` upstream case, but this version
		// also has `Banned` banned bare. Both reports fire: one on the
		// TypeName, one on the whole node text.
		{
			Code: `let value: Banned<X>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned":    true,
				"Banned<X>": "Don't parameterize Banned with X.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 12},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<X>` as a type. Don't parameterize Banned with X.", Line: 1, Column: 12},
			},
		},
		// ---- Branch lock-in TSClassImplements typeArguments → check both expression and whole node ----
		{
			Code: `class C implements Banned<X> {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned":    true,
				"Banned<X>": "Don't parameterize Banned.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 20},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<X>` as a type. Don't parameterize Banned.", Line: 1, Column: 20},
			},
		},
		// ---- Branch lock-in TSInterfaceHeritage typeArguments → check both expression and whole node ----
		{
			Code: `interface I extends Banned<X> {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned":    true,
				"Banned<X>": "Don't parameterize Banned.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 21},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<X>` as a type. Don't parameterize Banned.", Line: 1, Column: 21},
			},
		},
		// ---- Dimension 4 / Receiver wrappers: ParenthesizedType wrapping TypeReference ----
		// tsgo parses `(Banned)` in type position as ParenthesizedType containing
		// a TypeReference. ESTree flattens parens, so upstream's listener still
		// fires on the inner TSTypeReference. Lock in that the wrapper does not
		// prevent the report (and that the column anchors at the inner
		// TypeReference, not the opening paren).
		{
			Code:    `let value: (Banned);`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 13},
			},
		},
		// ---- Dimension 4 / Receiver wrappers: nested ParenthesizedType ----
		// `((Banned))` becomes ParenthesizedType(ParenthesizedType(TypeReference)).
		// The TypeReference listener fires on the innermost node regardless of
		// how many parens wrap it.
		{
			Code:    `let value: ((Banned));`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 14},
			},
		},
		// ---- Dimension 4 / Nesting: TypeReference inside Array<...> still fires ----
		// `Array<Banned>` is TypeReference(Array, [TypeReference(Banned)]). The
		// outer TypeReference's TypeName is `Array` (no match) but the inner
		// TypeReference fires on `Banned`.
		{
			Code:    `let value: Array<Banned>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 18},
			},
		},
		// ---- Dimension 4 / Nesting: empty TypeLiteral nested inside TypeLiteral ----
		// Outer `{ a: {} }` has one member so the empty-literal check skips
		// it; the inner `{}` is empty so the rule reports on it.
		{
			Code:    `let value: { a: {} };`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"{}": "Use object instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `{}` as a type. Use object instead.", Line: 1, Column: 17},
			},
		},
		// ---- Dimension 4 / Nesting: same-kind nesting boundary on empty tuple ----
		// `[Banned, []]` has two elements so empty-tuple check skips it; the
		// inner `[]` is empty so it triggers.
		{
			Code: `let value: [Banned, []];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"[]":     "Use unknown[] instead.",
				"Banned": true,
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 13},
				{MessageId: "bannedTypeMessage", Message: "Don't use `[]` as a type. Use unknown[] instead.", Line: 1, Column: 21},
			},
		},
		// ---- Dimension 4 / Access forms: QualifiedName text comparison strips inner whitespace ----
		// `NS . Banned` in source matches the bannedTypes key "NS.Banned"
		// after whitespace normalization. Locks in that `removeSpaces`
		// is applied to the *node text*, not just to user keys.
		{
			Code: `let value: NS . Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"NS.Banned": "Use NS.Ok instead.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `NS.Banned` as a type. Use NS.Ok instead.", Line: 1, Column: 12},
			},
		},
		// ---- Real-user: namespace + parameterized type ban ----
		// Locks in that whole-node match works through QualifiedName with
		// type arguments.
		{
			Code: `let value: NS.Banned<X>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"NS.Banned<X>": "Avoid NS.Banned<X>.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `NS.Banned<X>` as a type. Avoid NS.Banned<X>.", Line: 1, Column: 12},
			},
		},
		// ---- Dimension 4 / Type context: function-type parameter and return type ----
		// FunctionType nests parameter TypeReferences and a return TypeReference.
		// Each occurrence fires independently — locking that the listener walks
		// into every TypeReference in a FunctionTypeNode, not just top-level.
		{
			Code:    `let f: (x: Banned) => Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 12},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 23},
			},
		},
		// ---- Dimension 4 / Type context: generic type parameter constraint ----
		// `<T extends Banned>` puts a TypeReference inside a TypeParameter's
		// `constraint` slot. The listener must still fire there.
		{
			Code:    `function f<T extends Banned>() {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 22},
			},
		},
		// ---- Dimension 4 / Type context: conditional type (extends + true + false branches) ----
		// `T extends Banned ? Banned : never` puts TypeReferences in the
		// extends/true/false slots of a ConditionalType. Locks in that the
		// listener traverses all three.
		{
			Code:    `type T<X> = X extends Banned ? Banned : never;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 23},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 32},
			},
		},
		// ---- Dimension 4 / Type context: mapped type value ----
		// `{ [K in keyof T]: Banned }` puts a TypeReference inside a MappedType's
		// `type` slot. The listener must reach it.
		{
			Code:    `type M<T> = { [K in keyof T]: Banned };`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 31},
			},
		},
		// ---- Dimension 4 / Type context: keyof Banned (TypeOperator) ----
		// `keyof Banned` wraps a TypeReference in a TypeOperator node. The
		// listener fires on the inner TypeReference.
		{
			Code:    `type K = keyof Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 16},
			},
		},
		// ---- Dimension 4 / Type context: indexed access type ----
		// `Banned[K]` is IndexedAccessType(TypeReference, TypeReference). Both
		// the object and the index can be TypeReferences; here only the object
		// matches Banned.
		{
			Code:    `type I<K extends string> = Banned[K];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 28},
			},
		},
		// ---- Dimension 4 / Type context: `satisfies` expression ----
		// `x satisfies Banned` puts a TypeReference in a SatisfiesExpression's
		// type slot. Locks in that the listener walks expression-position
		// type annotations as well as declaration-position ones.
		{
			Code:    `1 satisfies Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 13},
			},
		},
		// ---- Dimension 4 / Type context: `new` with type arguments ----
		// `new Foo<Banned>()` puts the TypeReference inside the NewExpression's
		// type-argument list — a TypeArguments-of-NewExpression slot that is
		// distinct from the TypeArguments-of-TypeReference slot.
		{
			Code:    `const x = new Foo<Banned>();`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok instead."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 19},
			},
		},
		// ---- Dimension 4 / Same-kind nesting: TypeReference inside TypeReference ----
		// `Banned<Banned>` is TypeReference(Banned, [TypeReference(Banned)]).
		// With only `Banned` banned, three reports fire:
		//   1. outer TypeName  → "Banned"
		//   2. outer whole node (because TypeArguments != nil) → "Banned<Banned>"
		//      (no match; not banned)
		//   3. inner TypeName → "Banned"
		// So we see exactly two `Banned` reports.
		{
			Code: `let value: Banned<Banned>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": true,
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 12},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 19},
			},
		},
		// ---- tsgo AST: KindNullKeyword fires inside union/array/tuple positions ----
		// In tsgo, `null` in type position is `KindNullKeyword` directly (not
		// wrapped in `KindLiteralType`). The keyword listener must fire on it
		// regardless of how deeply nested it is, not just at the top of the
		// type annotation. Three positions, three reports.
		{
			Code: `let a: string | null; let b: null[]; let c: [null];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"null": "Use `unknown` or be explicit.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `null` as a type. Use `unknown` or be explicit.", Line: 1, Column: 17},
				{MessageId: "bannedTypeMessage", Message: "Don't use `null` as a type. Use `unknown` or be explicit.", Line: 1, Column: 30},
				{MessageId: "bannedTypeMessage", Message: "Don't use `null` as a type. Use `unknown` or be explicit.", Line: 1, Column: 46},
			},
		},
		// ---- tsgo AST: KindUndefinedKeyword fires inside union/array/type-arg positions ----
		// Same shape lock-in for undefined as for null — both are top-level
		// keyword kinds in tsgo (not literal-type wrapped), so the keyword
		// listener should fire in any nested position.
		{
			Code: `let a: string | undefined; let b: Array<undefined>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"undefined": "Prefer optional ?: instead.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `undefined` as a type. Prefer optional ?: instead.", Line: 1, Column: 17},
				{MessageId: "bannedTypeMessage", Message: "Don't use `undefined` as a type. Prefer optional ?: instead.", Line: 1, Column: 41},
			},
		},
		// ---- tsgo AST: KindBigIntKeyword fires inside intersection / type-arg / nested arrays ----
		// Locks in that primitive keyword listeners reach arbitrary nesting,
		// not just declaration-direct positions.
		{
			Code: `type X = bigint & { x: 1 }; type Y = Set<bigint>; let z: bigint[][];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"bigint": "Use number.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `bigint` as a type. Use number.", Line: 1, Column: 10},
				{MessageId: "bannedTypeMessage", Message: "Don't use `bigint` as a type. Use number.", Line: 1, Column: 42},
				{MessageId: "bannedTypeMessage", Message: "Don't use `bigint` as a type. Use number.", Line: 1, Column: 58},
			},
		},
		// ---- tsgo AST: PropertyAccess as heritage expression ----
		// `class X implements ns.Banned` — in tsgo the expression slot of
		// ExpressionWithTypeArguments can be a PropertyAccessExpression
		// (lowercase namespace bound in expression scope) rather than a
		// QualifiedName. Locks in that `stringifyNode` works on both shapes,
		// and the helper still matches the heritage gate.
		{
			Code:    `const ns = { Banned: class {} }; class C implements ns.Banned {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"ns.Banned": "Use ns.Ok."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `ns.Banned` as a type. Use ns.Ok.", Line: 1, Column: 53},
			},
		},
		// ---- tsgo AST: PropertyAccess heritage with type arguments ----
		// Outer-node match on `ns.Banned<T>` after the expression-only match
		// on `ns.Banned`. Locks in the typeArguments branch for the
		// PropertyAccess-shaped heritage.
		{
			Code: `const ns = { Banned: class<T>{} }; class C implements ns.Banned<number> {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"ns.Banned":         "Don't use the bare class.",
				"ns.Banned<number>": "Don't use the numeric specialization.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `ns.Banned` as a type. Don't use the bare class.", Line: 1, Column: 55},
				{MessageId: "bannedTypeMessage", Message: "Don't use `ns.Banned<number>` as a type. Don't use the numeric specialization.", Line: 1, Column: 55},
			},
		},
		// ---- tsgo AST: NamedTupleMember keeps tuple non-empty ----
		// `[name: Banned]` is a TupleType with one NamedTupleMember element.
		// Two lock-ins in one shape:
		//   1. The empty-tuple check skips (Elements has 1 node), so `[]` ban
		//      does NOT fire on this construct (verified via the absence of a
		//      `[]` error in Errors).
		//   2. The inner TypeReference still fires for `Banned`.
		{
			Code: `let v: [name: Banned];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"[]":     "Use unknown[] instead.",
				"Banned": "Use Ok instead.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 15},
			},
		},
		// ---- tsgo AST: OptionalType in tuple keeps tuple non-empty ----
		// `[Banned?]` is TupleType with one OptionalType element (wrapping a
		// TypeReference). Same lock-in: empty-tuple ban must not fire, inner
		// TypeReference must.
		{
			Code: `let v: [Banned?];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"[]":     "Use unknown[] instead.",
				"Banned": "Use Ok instead.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 9},
			},
		},
		// ---- tsgo AST: RestType in tuple keeps tuple non-empty ----
		// `[...Banned[]]` — RestType wrapping ArrayType wrapping TypeReference.
		// Inner TypeReference fires; empty-tuple ban must not.
		{
			Code: `let v: [...Banned[]];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"[]":     "Use unknown[] instead.",
				"Banned": "Use Ok instead.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok instead.", Line: 1, Column: 12},
			},
		},
		// ---- tsgo AST: `readonly` TypeOperator over banned tuple/array ----
		// `readonly Banned[]` and `readonly [Banned]` — TypeOperator(readonly, ...)
		// must not block descent into its operand.
		{
			Code: `let a: readonly Banned[]; let b: readonly [Banned];`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": true,
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 17},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 44},
			},
		},
		// ---- tsgo AST: type predicate ----
		// `function f(x): x is Banned` puts a TypeReference inside a
		// TypePredicateNode. Locks in that the type-predicate type slot is
		// walked by the listener.
		{
			Code:    `function check(x: unknown): x is Banned { return true; }`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok.", Line: 1, Column: 34},
			},
		},
		// ---- tsgo AST: template literal type ----
		// In tsgo, `${Banned}` inside a TemplateLiteralType holds a
		// TypeReference. Locks in that the listener fires inside the
		// template-substitution slot.
		{
			Code: `type T = ` + "`${Banned}`" + `;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": "Use a concrete type.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use a concrete type.", Line: 1, Column: 13},
			},
		},
		// ---- tsgo AST: TypeAssertion (`<Banned>x` old-style) ----
		// Locks in the angle-bracket assertion shape — different from `as`.
		// The TypeReference inside the assertion is reachable.
		{
			Code:    `const x = <Banned>1;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok.", Line: 1, Column: 12},
			},
			TSConfig: "tsconfig.json",
			Tsx:      false,
		},
		// ---- tsgo AST: `as` cast with ParenthesizedType ----
		// `1 as (Banned)` — AsExpression with type=ParenthesizedType(TypeReference).
		// Listener fires on the inner TypeReference regardless of the paren
		// wrapper.
		{
			Code:    `1 as (Banned);`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok.", Line: 1, Column: 7},
			},
		},
		// ---- tsgo AST: comment inside the type annotation ----
		// A comment between the colon and the TypeReference is leading
		// trivia; TrimNodeTextRange must skip it so the report still anchors
		// on `Banned`, not on the comment.
		{
			Code:    `let v: /* legacy */ Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": "Use Ok."}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok.", Line: 1, Column: 21},
			},
		},
		// ---- tsgo AST: multi-line type annotation with internal whitespace ----
		// Whitespace + newline inside `Banned<\n  number\n>` must collapse to
		// `Banned<number>` after removeSpaces — matching the same configured
		// key.
		{
			Code: "let v: Banned<\n  number\n>;",
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned<number>": "Don't parameterize Banned with number.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<number>` as a type. Don't parameterize Banned with number.", Line: 1, Column: 8},
			},
		},
		// ---- tsgo AST: multi-level QualifiedName (A.B.C.Banned) ----
		// Locks in that QualifiedName text extraction handles arbitrary
		// nesting depth, not just two-level NS.X.
		{
			Code: `let v: A.B.C.Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"A.B.C.Banned": "Reach for a flatter type.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `A.B.C.Banned` as a type. Reach for a flatter type.", Line: 1, Column: 8},
			},
		},
		// ---- Real-user: ban legacy `Function` keyword (no-unsafe-function-type style) ----
		// Common deprecation-period config: ban a primitive-ish global with
		// a fix suggestion. Locks in that `Function` is matched via
		// TypeReference TypeName (it is *not* a TS keyword node in either
		// ESTree or tsgo).
		{
			Code: `let f: Function;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Function": map[string]interface{}{
					"message": "Use a specific function signature instead.",
					"fixWith": "() => void",
				},
			}}),
			Output: []string{`let f: () => void;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Function` as a type. Use a specific function signature instead.", Line: 1, Column: 8},
			},
		},
		// ---- Real-user: ban `React.FC` with a suggested replacement ----
		// Frequently asked-for ban in React codebases. The suggestion replaces
		// only the *TypeName* (`React.FC`), not the outer TypeReference, so
		// the original `<Props>` argument carries over into the suggested
		// replacement. The replacement type `React.FunctionComponent` itself
		// accepts type arguments, so the resulting `React.FunctionComponent<Props>`
		// remains valid TypeScript — the test demonstrates the design intent
		// while keeping output syntactically valid.
		{
			Code: `type Props = {}; const C: React.FC<Props> = () => null;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"React.FC": map[string]interface{}{
					"message": "Use React.FunctionComponent instead.",
					"suggest": []interface{}{"React.FunctionComponent"},
				},
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedTypeMessage",
					Message:   "Don't use `React.FC` as a type. Use React.FunctionComponent instead.",
					Line:      1, Column: 27,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "bannedTypeReplacement", Output: `type Props = {}; const C: React.FunctionComponent<Props> = () => null;`},
					},
				},
			},
		},
		// ---- Design intent: fixWith on the TypeName preserves TypeArguments ----
		// This is the *positive-output* counterpart to the upstream
		// `Omit<Foo, 'a'>` → `Ok<Foo, 'a'>` migration test, written in this
		// extras suite so the typeName-only-replacement design is locked in
		// with a directly-named test, not just as a side effect of an upstream
		// case. If a future refactor anchors fix/report on the outer
		// TypeReference instead of the TypeName, both this and the upstream
		// Omit case fail together.
		{
			Code: `let v: Banned<X>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": map[string]interface{}{"fixWith": "Ok", "message": "Use Ok."},
			}}),
			Output: []string{`let v: Ok<X>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Ok.", Line: 1, Column: 8},
			},
		},
		// ---- Design intent: whole-node ban with fixWith replaces the whole TypeReference ----
		// Mirror of the design-intent lock-in above, this time configuring the
		// parameterized form `Banned<X>` directly. Here the fix DOES rewrite
		// the entire `Banned<X>` to `Ok` because the bannedTypes key matches
		// the whole-node text. Locks in that registering the parameterized
		// key replaces the parameterized form end-to-end.
		{
			Code: `let v: Banned<X>;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned<X>": map[string]interface{}{"fixWith": "Ok", "message": "Replace whole Banned<X>."},
			}}),
			Output: []string{`let v: Ok;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned<X>` as a type. Replace whole Banned<X>.", Line: 1, Column: 8},
			},
		},
		// ---- Branch lock-in: unknown JSON value shape collapses to bare ban ----
		// Upstream getCustomMessage treats any truthy non-string non-object
		// (e.g. a stray number `1` slipped past schema validation) as a bare
		// ban — customMessage = ''. parseBannedTypes' default arm mirrors
		// this: a number value goes through Kind="true", reporting with no
		// custom-message suffix.
		{
			Code: `let v: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": 1,
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 8},
			},
		},
		// ---- Real-user: same name banned and not-banned in same file ----
		// User configures `Banned` but a *different* identifier `Allowed`
		// appears in the same scope — must not bleed into reports.
		{
			Code: `type Allowed = { ok: true }; let a: Allowed; let b: Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": "Use Allowed instead.",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type. Use Allowed instead.", Line: 1, Column: 53},
			},
		},
		// ---- Real-user: class field, method param/return, and static field ----
		// Every type-annotation slot inside a class body is reachable.
		{
			Code: `
				class C {
					field: Banned = null as any;
					static staticField: Banned;
					method(x: Banned): Banned { return x; }
					get acc(): Banned { return null as any; }
					set acc(v: Banned) {}
				}
			`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": true,
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 3, Column: 13},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 4, Column: 26},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 5, Column: 16},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 5, Column: 25},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 6, Column: 17},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 7, Column: 17},
			},
		},
		// ---- Real-user: interface body with property, method, index signature ----
		{
			Code: `
				interface I {
					p: Banned;
					m(x: Banned): Banned;
					[k: string]: Banned;
				}
			`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Banned": true,
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 3, Column: 9},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 4, Column: 11},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 4, Column: 20},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 5, Column: 19},
			},
		},
		// ---- Real-user: generic default ----
		// `<T = Banned>` puts a TypeReference in the TypeParameter's
		// `default` slot, separate from the `constraint` slot.
		{
			Code:    `function f<T = Banned>() {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 16},
			},
		},
		// ---- Real-user: rest parameter & optional parameter ----
		// `?:` makes the parameter optional but the type slot stays a regular
		// TypeReference; `...` makes it rest, type slot is ArrayType.
		{
			Code:    `function f(a?: Banned, ...rest: Banned[]) {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 16},
				{MessageId: "bannedTypeMessage", Message: "Don't use `Banned` as a type.", Line: 1, Column: 33},
			},
		},
		// ---- Real-user: ban Object both bare and as global identifier ----
		// `let f = Object();` (value position) — must NOT report; `let v: Object;`
		// (type position) — must report. Lock in the two-position invariant.
		{
			Code: `let f = Object(); let v: Object;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"Object": "Use object (lowercase).",
			}}),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "bannedTypeMessage", Message: "Don't use `Object` as a type. Use object (lowercase).", Line: 1, Column: 26},
			},
		},
	})
}

// TestNoRestrictedTypesNonFalsePositives is a second extras suite focused on
// shapes that should NOT fire — defensive against future broadening of the
// listener set. Each case is a real syntax form that *looks* close to one of
// the banned shapes but, per upstream's listener wiring, must stay silent.
func TestNoRestrictedTypesNonFalsePositives(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedTypesRule, []rule_tester.ValidTestCase{
		// ---- Non-FP: `boolean` ban does NOT match `true` / `false` literal types ----
		// Upstream's TYPE_KEYWORDS maps "boolean" only to TSBooleanKeyword.
		// `true` / `false` are LiteralType(TrueKeyword/FalseKeyword), not
		// BooleanKeyword. Banning "boolean" must not fire on them.
		{
			Code:    `let t: true; let f: false;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"boolean": true}}),
		},
		// ---- Non-FP: keyword ban does NOT match a same-named TypeReference ----
		// `Bigint` (capital B) is a TypeReference, not the bigint keyword. We
		// don't have an entry for capital-B Bigint, so it must stay valid.
		{
			Code:    `type Bigint = number; let v: Bigint;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"bigint": true}}),
		},
		// ---- Non-FP: `import('module').Banned` is not a TypeReference ----
		// ImportType has its own AST kind; the qualifier "Banned" is not a
		// TypeReference. Upstream doesn't fire on it either. Locks in that we
		// don't accidentally bridge ImportType qualifiers to the TypeReference
		// listener.
		{
			Code:    `let v: import('./mod').Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Non-FP: `typeof Banned` does NOT fire ----
		// `typeof X` is TypeQuery containing an Identifier — not a
		// TypeReference. Upstream doesn't fire either.
		{
			Code:    `const Banned = {}; type T = typeof Banned;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Non-FP: `Banned` as a value identifier in expression position ----
		// `Banned()` (value call), `new Banned()` (value new), `Banned.foo`
		// (value access) — none are type positions. None must fire.
		{
			Code: `
				class Banned {
					static foo() {}
				}
				Banned();
				new Banned();
				Banned.foo();
			`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Non-FP: object literal key named like a banned type ----
		// `{ Banned: 1 }` — Banned is a PropertyName, not a TypeReference.
		// Must stay silent.
		{
			Code:    `const x = { Banned: 1 };`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Non-FP: `enum` declaration with banned name ----
		// The enum *declaration* names a type but it's not a TypeReference
		// usage. References *to* the enum would fire; the declaration site
		// shouldn't.
		{
			Code:    `enum Banned { A, B }`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Non-FP: empty interface body / empty class body ----
		// Empty `interface I {}` body is *not* an empty TypeLiteral — it's an
		// InterfaceDeclaration. Banning `{}` must not fire on it.
		{
			Code:    `interface I {} class C {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"{}": "Use object."}}),
		},
		// ---- Non-FP: `class extends` with banned identifier ----
		// Lock-in alongside the in-positive-suite test: `class X extends Banned {}`
		// must stay silent because upstream omits the class-extends listener.
		// This second copy lives in the false-positive defense suite to
		// catch any drift where a future listener-set widening starts firing.
		{
			Code:    `class X extends Banned {}`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{"Banned": true}}),
		},
		// ---- Non-FP: empty bannedTypes (default options) means nothing fires ----
		// Important real-user default: if the user does not supply `types`,
		// the rule degrades to a no-op. Verifies that an empty map (and an
		// entirely missing options) doesn't break the listener wiring.
		{
			Code: `let v: bigint | string; let w: {}; let x: []; class C implements Whatever {}`,
		},
		// ---- Non-FP: explicit `null` value disables an otherwise-active key ----
		// Already covered in the positive-suite valid cases; repeating in the
		// FP suite locks the disable-via-null branch into the FP audit.
		{
			Code: `let v: bigint;`,
			Options: optionsArr(map[string]interface{}{"types": map[string]interface{}{
				"bigint": nil,
			}}),
		},
	}, []rule_tester.InvalidTestCase{})
}
