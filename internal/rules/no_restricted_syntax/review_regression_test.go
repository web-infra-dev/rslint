package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedSyntaxReviewRegressions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`ClassDeclaration > MethodDefinition`},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`MethodDefinition > BlockStatement`},
			},
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{`TSEnumDeclaration > TSEnumMember`},
			},
			{
				Code:    `enum E { A, B }`,
				Options: []interface{}{`TSEnumBody[body]`},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `const f = () => {};`,
				Options: []interface{}{`ArrowFunctionExpression > BlockStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `foo(); bar();`,
				Options: []interface{}{`ExpressionStatement + ExpressionStatement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { a() {} b() {} }`,
				Options: []interface{}{`ClassBody:has(> MethodDefinition)`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { a() {} b() {} }`,
				Options: []interface{}{`MethodDefinition:nth-child(2)`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{`JSXElement:has(> JSXOpeningElement)`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{`JSXOpeningElement[type='JSXOpeningElement']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo>hi</Foo>;`,
				Tsx:     true,
				Options: []interface{}{`JSXElement[type='JSXElement']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`FunctionExpression`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`:function`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`MethodDefinition > :function`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code:    `class C { m() {} }`,
				Options: []interface{}{`:function[body.type='BlockStatement']`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedSyntax",
					Line:      1, Column: 12, EndLine: 1, EndColumn: 17,
				}},
			},
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{`:is(JSXElement, JSXOpeningElement)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax"},
					{MessageId: "restrictedSyntax"},
				},
			},
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{`JSXElement > JSXOpeningElement`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo />;`,
				Tsx:     true,
				Options: []interface{}{`JSXOpeningElement > JSXIdentifier`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo a b={c} />;`,
				Tsx:     true,
				Options: []interface{}{`JSXOpeningElement[attributes.length=2]`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code:    `const x = <Foo a b={c} />;`,
				Tsx:     true,
				Options: []interface{}{`JSXOpeningElement[attributes.0.name.name='a']`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax"}},
			},
			{
				Code: `const x = <Foo />;`,
				Tsx:  true,
				Options: []interface{}{
					map[string]interface{}{"selector": `:is(JSXElement, JSXOpeningElement)`, "message": "union"},
					map[string]interface{}{"selector": `[type='JSXElement']`, "message": "physical"},
					map[string]interface{}{"selector": `JSXElement > JSXOpeningElement`, "message": "opening"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "union"},
					{MessageId: "restrictedSyntax", Message: "physical"},
					{MessageId: "restrictedSyntax", Message: "union"},
					{MessageId: "restrictedSyntax", Message: "opening"},
				},
			},
		},
	)
}
