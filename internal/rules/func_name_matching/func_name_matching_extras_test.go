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
