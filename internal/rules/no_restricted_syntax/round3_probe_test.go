package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedSyntax_Round3Probes targets ESTree↔tsgo divergences
// the previous rounds did not exercise. Each probe encodes a real-world
// selector pattern that should match (positive) or must NOT match
// (negative) — a single failing probe reveals a real implementation bug.
func TestNoRestrictedSyntax_Round3Probes(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			// MethodDefinition[kind='constructor'] should NOT match a
			// regular class method.
			{
				Code:    `class C { foo() {} }`,
				Options: []interface{}{"MethodDefinition[kind='constructor']"},
			},
			// MethodDefinition[kind='method'] should NOT match a
			// constructor.
			{
				Code:    `class C { constructor() {} }`,
				Options: []interface{}{"MethodDefinition[kind='method']"},
			},
			// Property[shorthand=true] should NOT match a non-shorthand prop.
			{
				Code:    `({ a: 1 });`,
				Options: []interface{}{"Property[shorthand=true]"},
			},
			// Property[method=true] should NOT match a regular property.
			{
				Code:    `({ a: 1 });`,
				Options: []interface{}{"Property[method=true]"},
			},
			// Property[method=true] should NOT match a class method
			// (those are MethodDefinition).
			{
				Code:    `class C { foo() {} }`,
				Options: []interface{}{"Property[method=true]"},
			},
			// RestElement should NOT match a regular function parameter.
			{
				Code:    `function f(a) {}`,
				Options: []interface{}{"RestElement"},
			},
			// AssignmentPattern should NOT match a regular function param.
			{
				Code:    `function f(a) {}`,
				Options: []interface{}{"AssignmentPattern"},
			},
			// ClassDeclaration[superClass] should NOT match a class without
			// extends clause.
			{
				Code:    `class C {}`,
				Options: []interface{}{"ClassDeclaration[superClass]"},
			},
			// SwitchCase[test=null] should NOT match a non-default case.
			{
				Code:    `switch (x) { case 1: y; }`,
				Options: []interface{}{"SwitchCase[test=null]"},
			},
			// ForStatement[update] should NOT match a for without update.
			{
				Code:    `for (var i = 0;;) {}`,
				Options: []interface{}{"ForStatement[update]"},
			},
			// PropertyDefinition[value] should NOT match an uninitialized
			// class field.
			{
				Code:    `class C { x; }`,
				Options: []interface{}{"PropertyDefinition[value]"},
			},
			// TemplateLiteral[expressions.length=0] does NOT match a
			// template with substitutions.
			{
				Code:    "const s = `a${b}c`;",
				Options: []interface{}{"TemplateLiteral[expressions.length=0]"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// MethodDefinition[kind='method'] — class regular methods.
			{
				Code:    `class C { foo() {} }`,
				Options: []interface{}{"MethodDefinition[kind='method']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition[kind='method']' is not allowed."},
				},
			},
			// MethodDefinition[kind='constructor']
			{
				Code:    `class C { constructor() {} }`,
				Options: []interface{}{"MethodDefinition[kind='constructor']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition[kind='constructor']' is not allowed."},
				},
			},
			// MethodDefinition[kind='get']
			{
				Code:    `class C { get x() { return 1; } }`,
				Options: []interface{}{"MethodDefinition[kind='get']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition[kind='get']' is not allowed."},
				},
			},
			// Property[shorthand=true] — object literal shorthand.
			{
				Code:    `const a = 1; ({ a });`,
				Options: []interface{}{"Property[shorthand=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property[shorthand=true]' is not allowed."},
				},
			},
			// Property[method=true] — object method shorthand.
			{
				Code:    `({ foo() {} });`,
				Options: []interface{}{"Property[method=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property[method=true]' is not allowed."},
				},
			},
			// RestElement on function parameter.
			{
				Code:    `function f(...args) {}`,
				Options: []interface{}{"RestElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'RestElement' is not allowed."},
				},
			},
			// AssignmentPattern on parameter default.
			{
				Code:    `function f(a = 1) {}`,
				Options: []interface{}{"AssignmentPattern"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'AssignmentPattern' is not allowed."},
				},
			},
			// ClassDeclaration[superClass] on extending class.
			{
				Code:    `class C extends Base {}`,
				Options: []interface{}{"ClassDeclaration[superClass]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ClassDeclaration[superClass]' is not allowed."},
				},
			},
			// SwitchCase default — test=null.
			{
				Code:    `switch (x) { default: y; }`,
				Options: []interface{}{"SwitchCase[test=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'SwitchCase[test=null]' is not allowed."},
				},
			},
			// SwitchCase non-default — has test, length > 0 consequent.
			{
				Code:    `switch (x) { case 1: y; break; }`,
				Options: []interface{}{"SwitchCase[consequent.length>0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'SwitchCase[consequent.length>0]' is not allowed."},
				},
			},
			// ForStatement[update] — has update.
			{
				Code:    `for (var i = 0; i < 10; i++) {}`,
				Options: []interface{}{"ForStatement[update]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'ForStatement[update]' is not allowed."},
				},
			},
			// PropertyDefinition[value] — initialized class field.
			{
				Code:    `class C { x = 1; }`,
				Options: []interface{}{"PropertyDefinition[value]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'PropertyDefinition[value]' is not allowed."},
				},
			},
			// TemplateLiteral with substitutions matches expressions.length>0.
			{
				Code:    "const s = `a${b}c`;",
				Options: []interface{}{"TemplateLiteral[expressions.length>0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'TemplateLiteral[expressions.length>0]' is not allowed."},
				},
			},
			// "use strict" directive — ESTree marks ExpressionStatement.directive.
			{
				Code:    `"use strict"; foo();`,
				Options: []interface{}{`ExpressionStatement[directive='use strict']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ExpressionStatement[directive='use strict']' is not allowed.`},
				},
			},
			// Literal[bigint] — BigInt literal detection.
			{
				Code:    `const n = 1n;`,
				Options: []interface{}{"Literal[bigint]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal[bigint]' is not allowed."},
				},
			},
			// Literal[regex] — regex literal detection (vs other literals).
			{
				Code:    `const r = /abc/;`,
				Options: []interface{}{"Literal[regex]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Literal[regex]' is not allowed."},
				},
			},
			// Property[computed=true] — object key bracketed.
			{
				Code:    `({ [x]: 1 });`,
				Options: []interface{}{"Property[computed=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property[computed=true]' is not allowed."},
				},
			},
			// MethodDefinition[static=true] — static class method.
			{
				Code:    `class C { static foo() {} }`,
				Options: []interface{}{"MethodDefinition[static=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'MethodDefinition[static=true]' is not allowed."},
				},
			},
			// PropertyDefinition[static=true] — static class field.
			{
				Code:    `class C { static x = 1; }`,
				Options: []interface{}{"PropertyDefinition[static=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'PropertyDefinition[static=true]' is not allowed."},
				},
			},
			// YieldExpression[delegate=true] — `yield*`.
			{
				Code:    `function* g() { yield* x; }`,
				Options: []interface{}{"YieldExpression[delegate=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'YieldExpression[delegate=true]' is not allowed."},
				},
			},
			// CatchClause > BlockStatement — body access.
			{
				Code:    `try { f(); } catch (e) { g(); }`,
				Options: []interface{}{"CatchClause > BlockStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'CatchClause > BlockStatement' is not allowed."},
				},
			},
			// CallExpression > Super — `super()` invocation.
			{
				Code:    `class C extends B { constructor() { super(); } }`,
				Options: []interface{}{"CallExpression > Super"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'CallExpression > Super' is not allowed."},
				},
			},
			// ArrowFunctionExpression as PropertyDefinition.value.
			{
				Code:    `class C { onClick = () => {}; }`,
				Options: []interface{}{`PropertyDefinition[value.type='ArrowFunctionExpression']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'PropertyDefinition[value.type='ArrowFunctionExpression']' is not allowed.`},
				},
			},
			// Class extends specific name.
			{
				Code:    `class C extends Component {}`,
				Options: []interface{}{`ClassDeclaration[superClass.name='Component']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'ClassDeclaration[superClass.name='Component']' is not allowed.`},
				},
			},
			// Object-literal computed key with a string.
			{
				Code:    `({ ['x']: 1 });`,
				Options: []interface{}{"Property[computed=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'Property[computed=true]' is not allowed."},
				},
			},
			// `:has(...)` against statements inside SwitchCase.
			{
				Code:    `switch (x) { case 1: foo(); break; }`,
				Options: []interface{}{`SwitchCase:has(BreakStatement)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'SwitchCase:has(BreakStatement)' is not allowed.`},
				},
			},
			// JSXAttribute name access
			{
				Code:    `const x = <Foo bar="baz" />;`,
				Tsx:     true,
				Options: []interface{}{`JSXAttribute[name.name='bar']`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: `Using 'JSXAttribute[name.name='bar']' is not allowed.`},
				},
			},
			// Object destructuring with default — AssignmentPattern at
			// BindingElement with Initializer (no rest).
			{
				Code:    `const { a = 1 } = obj;`,
				Options: []interface{}{"AssignmentPattern"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'AssignmentPattern' is not allowed."},
				},
			},
			// Array destructuring with rest — RestElement at
			// BindingElement with DotDotDotToken.
			{
				Code:    `const [first, ...tail] = arr;`,
				Options: []interface{}{"RestElement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'RestElement' is not allowed."},
				},
			},
			// Static class block — ESTree StaticBlock.
			{
				Code:    `class C { static { initialize(); } }`,
				Options: []interface{}{"StaticBlock"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'StaticBlock' is not allowed."},
				},
			},
		},
	)
}
