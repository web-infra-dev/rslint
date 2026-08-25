// TestFuncNameMatchingExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension walk notes for func-name-matching:
//   - Dimension 3 (autofix boundaries): N/A — the rule has no autofix or
//     suggestions.
//   - Dimension 4 access/key forms: PrivateIdentifier keys are already
//     covered by the upstream migration (class D's `#foo`/`#bar` valid
//     cases, and the AssignmentExpression `this.#x = ...` / `a.b.#x = ...`
//     class-field-adjacent invalid/valid cases).
//   - Dimension 4 nesting/traversal boundaries and graceful degradation:
//     SpreadAssignment/RestElement never appear as this rule's own node
//     (PropertyAssignment/PropertyDeclaration/VariableDeclaration/
//     BinaryExpression) — N/A, nothing to degrade against.
//   - Dimension 4 optional-chain forms are already exercised by upstream's
//     own invalid migration (`Object?.defineProperty(...)`,
//     `(obj?.aaa).foo = ...`, `(Object?.defineProperties)(...)`).
package func_name_matching

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestFuncNameMatchingExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncNameMatchingRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: (module).exports — parenthesized receiver is
			// invisible to ESTree, so upstream's own isModuleExports check
			// would still recognize this; tsgo keeps the parens explicit and
			// the rule's AssignmentExpression handler unwraps them the same
			// way ----
			{Code: `(module).exports = function foo(name) {};`},

			// ---- Dimension 4: computed bracket access whose argument is a
			// TemplateLiteral, not an ESTree Literal — the computed-literal
			// guard rejects it outright, so this is a no-op regardless of
			// name/option ----
			{Code: "obj[`foo`] = function foo() {};"},

			// ---- Dimension 4: computed bracket access with an empty-string
			// literal key — passes the Literal guard, but the extracted name
			// is empty so nothing is compared ----
			{Code: `obj[''] = function foo() {};`},

			// ---- Dimension 4: numeric-literal computed key — passes the
			// Literal guard, but "0" is never identifier-shaped ----
			{Code: `obj[0] = function foo() {};`},

			// ---- Branch lock-in: the AssignmentExpression listener applies
			// its identifier-shape/reserved-word gate uniformly to both
			// dot-access and bracket-access property names — "for" is a
			// reserved word, so neither form is ever compared, regardless of
			// the mismatching function name. Contrast with the Property
			// listener's plain-Identifier-key invalid case below, where the
			// same word IS compared ----
			{Code: `obj.for = function bar() {};`},
			{Code: `obj['for'] = function bar() {};`},

			// ---- Branch lock-in: "yield" is not a reserved word under
			// esutils' ES5 keyword table (only ES6's unconditionally reserves
			// it), so it is still checked as a property name at
			// ecmaVersion 5 when it happens to already match the function
			// name ----
			{Code: `({'yield': function yield() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 5}},

			// ---- Default-version lock-in: rslint and ESLint flat config
			// default to the latest ECMAScript version, where "yield" is
			// reserved and a string-literal property key is not compared ----
			{Code: `({'yield': function y() {}})`},

			// ---- Branch lock-in: at ecmaVersion 2015, "yield" is reserved
			// (esutils' ES6 keyword table always rejects it, regardless of
			// strict mode), so the property key is left unchecked even
			// though the function name mismatches ----
			{Code: `({'yield': function y() {}})`, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015}},

			// ---- Branch lock-in: considerPropertyDescriptor's
			// Object.defineProperty branch requires the property-name
			// argument to be a string literal; a non-string second argument
			// leaves the descriptor object unchecked entirely (falls out of
			// the switch without reaching the plain fallback, since the
			// outer guard on propertyName === "value" already matched) ----
			{
				Code:    `Object.defineProperty(foo, 0, { value: function bar() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
			},

			// ---- Branch lock-in: considerPropertyDescriptor's
			// Object.defineProperties/Object.create branch only reads a
			// non-computed Identifier outer key (node.parent.parent.key);
			// a computed outer key is skipped even though it's the direct
			// argument of a matching call ----
			{
				Code:            `Object.defineProperties(foo, { [bar]: { value: function bar() {} } })`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
			},

			// ---- Divergence lock-in: upstream reads the outer key as
			// `node.parent.parent.key.name`, which is `undefined` for a
			// string- or numeric-literal key and then reported verbatim as
			// ``should match property name `undefined` ``. The port compares
			// only identifier outer keys, so these stay unreported ----
			{
				Code:    `Object.defineProperties(foo, { "bar": { value: function baz() {} } })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
			},
			{
				Code:    `Object.create(a, { 123: { value: function baz() {} } })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
			},

			// ---- Dimension 4: auto-accessor class fields are AccessorProperty
			// nodes in ESTree, which the rule's `Property,
			// PropertyDefinition[value]` selector never visits; tsgo spells
			// the same field as a property declaration with an `accessor`
			// modifier, so the listener has to opt out of it ----
			{Code: `class C { accessor x = function y() {}; }`},
			{
				Code:    `class C { static accessor value = function foo() {}; }`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
			},

			// ---- Divergence lock-in: with the property-name argument missing
			// entirely, upstream dereferences `arguments[1]` and throws a
			// TypeError that surfaces as a lint crash; the port checks the
			// argument count and leaves the descriptor unchecked ----
			{
				Code:    `Object.defineProperty({ value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
			},

			// ---- Branch lock-in: a shorthand `value` descriptor entry holds
			// an Identifier, not a FunctionExpression, so no listener applies
			// no matter which descriptor call encloses it ----
			{
				Code:    `Object.defineProperty(foo, 'bar', { value })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
			},

			// ---- Dimension 4: TS-only wrappers that ESTree models as their
			// own node types — a non-null assertion applied to the assignment
			// target is a TSNonNullExpression rather than a MemberExpression,
			// and a wrapped function expression is no longer a
			// FunctionExpression initializer ----
			{Code: `obj.foo! = function bar() {};`},
			{Code: `var foo = (function bar() {})!;`},
			{Code: `({ foo: function bar() {} as any })`},

			// ---- Branch lock-in: a named class expression is not a function
			// expression, so the assignment is never compared ----
			{Code: `foo = class bar {};`},

			// ---- Real-user: named CommonJS export assignment
			// (`exports.foo = function foo() {}`) — a very common Node.js
			// idiom distinct from `module.exports`; matches here since names
			// agree ----
			{Code: `exports.foo = function foo() {};`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized dot-access receiver — parens
			// are invisible to ESTree; tsgo's AssignmentExpression handler
			// unwraps them via ast.SkipParentheses before inspecting the
			// left-hand side's kind ----
			{
				Code: `(obj).foo = function bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},

			// ---- Dimension 4: TS non-null-assertion receiver nested inside
			// a dot-access target (`obj!.foo`) — the top-level node is still
			// a plain PropertyAccessExpression, so identification of the
			// assignment target is unaffected by what tsgo wraps around its
			// own Expression field ----
			{
				Code: `obj!.foo = function bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},

			// ---- Dimension 4: TS `as` type-assertion receiver, parenthesized
			// as TS syntax requires (`(obj as any).foo`) — same reasoning as
			// the non-null-assertion case above ----
			{
				Code: `(obj as any).foo = function bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},

			// ---- Branch lock-in: a reserved word used as a plain
			// (non-computed) Identifier object-literal property key IS
			// compared — the Property listener's isIdentifierName gate only
			// applies to its string-literal-key path, never to a syntactic
			// `.key` Identifier. Contrast with the AssignmentExpression
			// `obj.for` / `obj['for']` valid cases above, where the same
			// word is always gated ----
			{
				Code: `({for: function bar() {}})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 3},
				},
			},

			// ---- Branch lock-in: the AssignmentExpression listener's
			// identifier-shape check is hardcoded to ES5 rules independent of
			// the file's configured ecmaVersion (upstream calls its
			// isIdentifier helper with no ecmaVersion argument here). At
			// ecmaVersion 2020, "yield" would be ES6-reserved and skipped by
			// the Property listener's own check — but this is the
			// AssignmentExpression listener, so the check still fires ----
			{
				Code:            `obj['yield'] = function y() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},

			// ---- Explicit-version lock-in: unlike the default/latest case
			// above, ES5 accepts "yield" as an identifier-shaped property
			// name, so a mismatch is reported ----
			{
				Code:            `({'yield': function y() {}})`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 5},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `y` should match property name `yield`.", Line: 1, Column: 3},
				},
			},

			// ---- Branch lock-in: a class field named exactly "value" with
			// considerPropertyDescriptor enabled never enters the
			// descriptor-call branches, because its parent is a class body,
			// not an ObjectLiteralExpression — it falls through to the plain
			// property-name comparison instead ----
			{
				Code:            `class C { value = function foo() {}; }`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 11},
				},
			},

			// ---- Branch lock-in: the Object.defineProperties/Object.create
			// branch is selected by the shape of the call four levels up, but
			// the outer key it reads only exists when the enclosing node is a
			// property assignment. Nested array literals reach the call at the
			// same depth, so the descriptor object here is a plain object
			// literal whose own `value` key is what gets compared ----
			{
				Code:    `Object.create(a, [[{ value: function baz() {} }]]);`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `value`.", Line: 1, Column: 22},
				},
			},

			// ---- Branch lock-in: tsgo hangs class members directly off the
			// class, one level shallower than ESTree's ClassBody, so a class
			// field inside a descriptor call reaches it at the same depth as a
			// descriptor-map entry. The key compared is still the field's own
			// `value` property, not the field name ----
			{
				Code:            `Object.create(a, class { bar = { value: function baz() {} } });`,
				Options:         []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `value`.", Line: 1, Column: 34},
				},
			},

			// ---- Dimension 4: parenthesized descriptor map / descriptor
			// object — parens are invisible to ESTree, so upstream still walks
			// straight from the descriptor's `value` property to the enclosing
			// Object.create/defineProperties/defineProperty call ----
			{
				Code:    `Object.create(a, ({ bar: { value: function baz() {} } }));`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 28},
				},
			},
			{
				Code:    `Object.defineProperties(foo, { bar: ({ value: function baz() {} }) });`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 40},
				},
			},
			{
				Code:    `Object.defineProperty(foo, 'bar', ({ value: function baz() {} }));`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 38},
				},
			},

			// ---- Wrapper boundary: ESTree omits parentheses around the
			// descriptor-name argument. tsgo keeps them, so both direct
			// descriptor-call branches must unwrap the argument before
			// checking whether it is a string literal ----
			{
				Code:    `Object.defineProperty(foo, ("bar"), { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 39},
				},
			},
			{
				Code:    `Reflect.defineProperty(foo, ("bar"), { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 40},
				},
			},

			// ---- Branch lock-in: the reserved-word table esutils applies here
			// is its non-strict one, so `await`, `let`, `static` and
			// `implements` are all still identifier-shaped property names and
			// get compared — at every ecmaVersion, on both the string-literal
			// key path and the assignment path ----
			{
				Code:            `({'await': function foo() {}})`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `foo` should match property name `await`.", Line: 1, Column: 3},
				},
			},
			{
				Code:            `({'implements': function foo() {}})`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2015},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `foo` should match property name `implements`.", Line: 1, Column: 3},
				},
			},
			{
				Code:            `obj['static'] = function foo() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `foo` should match property name `static`.", Line: 1, Column: 1},
				},
			},

			// ---- Divergence lock-in: esutils uses its frozen Unicode 9
			// identifier table for the ES6 path too, so a post-Unicode-9
			// code point stays unchecked upstream at a modern ecmaVersion.
			// tsgo's current scanner table accepts U+10570 and reports ----
			{
				Code:            `({ "𐕰": function foo() {} })`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `foo` should match property name `𐕰`.", Line: 1, Column: 4},
				},
			},

			// ---- Divergence lock-in: the assignment path asks upstream
			// for its default ES5 identifier check at every configured
			// version, while rslint still uses tsgo's current table ----
			{
				Code:            `obj["ᢅ"] = function foo() {};`,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `foo` should match property name `ᢅ`.", Line: 1, Column: 1},
				},
			},

			// ---- Dimension 4: an optional call to a descriptor function is
			// still a CallExpression, and TS type arguments on the callee
			// leave it one too; a non-null assertion around the callee is a
			// TSNonNullExpression instead, which no longer matches
			// `Object.defineProperty` and leaves the plain `value` key as the
			// compared name ----
			{
				Code:    `Object.defineProperty?.(foo, 'bar', { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 39},
				},
			},
			{
				Code:    `Object.defineProperty<any>(foo, 'bar', { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `bar`.", Line: 1, Column: 42},
				},
			},
			{
				Code:    `Object.defineProperty!(foo, 'bar', { value: function baz() {} })`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `value`.", Line: 1, Column: 38},
				},
			},

			// ---- Dimension 4: a TS type assertion around the descriptor map
			// is a node of its own in both ASTs, so — unlike parentheses — it
			// breaks the walk from the descriptor up to the enclosing call ----
			{
				Code:    `Object.defineProperties(foo, { bar: { value: function baz() {} } } as any)`,
				Options: []any{"always", map[string]any{"considerPropertyDescriptor": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `baz` should match property name `value`.", Line: 1, Column: 39},
				},
			},

			// ---- Branch lock-in: the assignment listener is keyed on the
			// node type, not on the operator, so compound assignments are
			// compared like plain ones; a chained assignment is compared at
			// its innermost target, the only one whose right-hand side is the
			// function expression ----
			{
				Code: `foo += function bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Message: "Function name `bar` should match variable name `foo`.", Line: 1, Column: 1},
				},
			},
			{
				Code: `a = b = function foo() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Message: "Function name `foo` should match variable name `b`.", Line: 1, Column: 5},
				},
			},

			// ---- Dimension 4: async, generator and async-generator function
			// expressions are all plain FunctionExpressions ----
			{
				Code: `foo = async function* bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Message: "Function name `bar` should match variable name `foo`.", Line: 1, Column: 1},
				},
			},

			// ---- Reported positions: columns are counted in UTF-16 units
			// after an astral-plane literal, and a leading BOM does not shift
			// the first line's columns ----
			{
				Code: `var s = '😀'; ({ 'foo': function bar() {} });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `bar` should match property name `foo`.", Line: 1, Column: 18},
				},
			},
			{
				Code: "\ufeff({ 'foo': function bar() {} })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Message: "Function name `bar` should match property name `foo`.", Line: 1, Column: 4},
				},
			},

			// ---- Dimension 4: a TS namespace body is just another statement
			// container around the declaration ----
			{
				Code: `namespace N { var foo = function bar() {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchVariable", Message: "Function name `bar` should match variable name `foo`.", Line: 1, Column: 19},
				},
			},

			// ---- Real-user: assigning onto a property OF module.exports
			// (`module.exports.foo = ...`), a common named-export idiom —
			// isModuleExports only special-cases the exact `module.exports`
			// / `module['exports']` member expression, not deeper chains off
			// of it, so this is checked like any other property assignment ----
			{
				Code: `module.exports.foo = function bar() {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "matchProperty", Line: 1, Column: 1},
				},
			},
		},
	)
}
