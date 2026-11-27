package no_restricted_syntax

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Regression expectations were checked against ESLint 10.10.0, esquery 1.7.0,
// and @typescript-eslint/parser 8.69.0.
func TestNoRestrictedSyntaxStructureRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{
				// virtual sibling function
				Code:    "class C { a() {} b() {} }",
				Options: []any{"FunctionExpression + MethodDefinition"},
			},
			{
				// virtual sibling class body
				Code:    "class A {} class B {}",
				Options: []any{"ClassBody + ClassDeclaration"},
			},
			{
				// fields do not panic
				Code:    "class C { m(a,b) {} } const x = <Foo a b={c} />;",
				Options: []any{".body.params"},
				Tsx:     true,
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassBody:has(ClassDeclaration)"},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassBody:has(ClassDeclaration > ClassBody)"},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"MethodDefinition[body]"},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"MethodDefinition[params]"},
			},
			{
				// quoted constructor
				Code:    "class C { \"constructor\"(){} }",
				Options: []any{"Identifier.key"},
			},
			{
				// quoted constructor
				Code:    "class C { \"constructor\"(){} }",
				Options: []any{"MethodDefinition[key.name=\"constructor\"]"},
			},
			{
				// static block
				Code:    "class C { static { f(); } }",
				Options: []any{"BlockStatement"},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"CallExpression + CallExpression"},
			},
			{
				// attributes
				Code:    "f(); new F(); returnValue; function f(){ return x; }",
				Options: []any{"CallExpression[expression]"},
			},
			{
				// attributes
				Code:    "f(); new F(); returnValue; function f(){ return x; }",
				Options: []any{"ReturnStatement[expression]"},
			},
			{
				// array index
				Code:    "f(a,b);",
				Options: []any{"CallExpression[arguments.01.name=\"b\"]"},
			},
			{
				// jsx identity fields
				Code:    "const x=<Foo a />;",
				Options: []any{"JSXElement[name]"},
				Tsx:     true,
			},
			{
				// jsx identity fields
				Code:    "const x=<Foo a />;",
				Options: []any{"JSXElement[attributes]"},
				Tsx:     true,
			},
			{
				// parser token children
				Code:    "class C {; async m(){} }",
				Options: []any{"ClassDeclaration:has(> :not(Identifier,ClassBody))"},
			},
			{
				// parser token children
				Code:    "class C {; async m(){} }",
				Options: []any{"MethodDefinition:has(> :not(Identifier,FunctionExpression))"},
			},
			{
				// operator children
				Code:    "a + 1;",
				Options: []any{"BinaryExpression:has(> :not(Identifier,Literal))"},
			},
			{
				// operator children
				Code:    "a + 1;",
				Options: []any{"Program:has(> :not(ExpressionStatement))"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				// desc class method
				Code:    "class C { m() {} }",
				Options: []any{"ClassDeclaration MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// desc method body
				Code:    "class C { m() {} }",
				Options: []any{"MethodDefinition BlockStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// desc class call
				Code:    "class C { m() { foo(); } }",
				Options: []any{"ClassDeclaration CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// desc jsx name
				Code:    "const x = <Foo />;",
				Options: []any{"JSXElement JSXIdentifier"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
				},
			},
			{
				// child chain control
				Code:    "class C { m() {} }",
				Options: []any{"ClassDeclaration > ClassBody > MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// dispatch virtual parent (canonical visitor order: key before value)
				Code:    "class C { m() {} }",
				Options: []any{"[parent.type='MethodDefinition']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// dispatch virtual parent explicit
				Code:    "class C { m() {} }",
				Options: []any{"FunctionExpression[parent.type='MethodDefinition']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// dispatch virtual parent class (canonical visitor order: id before body)
				Code:    "class C {}",
				Options: []any{"[parent.type='ClassDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "restrictedSyntax", Line: 1, Column: 9, EndLine: 1, EndColumn: 11},
				},
			},
			{
				// jsx regex union
				Code:    "const x = <Foo a b={c} />;",
				Options: []any{"[type=/^(JSXElement|JSXOpeningElement)$/]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
				},
			},
			{
				// jsx regex child
				Code:    "const x = <Foo />;",
				Options: []any{"JSXElement > [type=/^JSXOpeningElement$/]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// jsx not
				Code:    "const x = <Foo />;",
				Options: []any{":not(Program,VariableDeclaration,VariableDeclarator,Identifier,JSXElement,JSXIdentifier)"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// method range call
				Code:    "class C { m(x = foo()) {} }",
				Options: []any{"FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
				},
			},
			{
				// method range type
				Code:    "class C { m(x: () => void) {} }",
				Options: []any{"FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 30},
				},
			},
			{
				// method range return type
				Code:    "class C { m(): () => void {} }",
				Options: []any{"FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 29},
				},
			},
			{
				// param position
				Code:    "class C { m(x, y) {} }",
				Options: []any{"FunctionExpression > Identifier:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// jsx opening attrs
				Code:    "const x = <Foo a b />;",
				Options: []any{"JSXOpeningElement[attributes.length=2]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// jsx opening self
				Code:    "const x = <Foo />;",
				Options: []any{"JSXOpeningElement[selfClosing=true]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// jsx expression empty
				Code:    "const x = <Foo>{/* c */}</Foo>;",
				Options: []any{"JSXExpressionContainer[expression.type='JSXEmptyExpression']"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// enum members field
				Code:    "enum E { A, B }",
				Options: []any{"TSEnumMember.members"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				// decorator expression field
				Code:    "@sealed() class C {}",
				Options: []any{"Decorator > CallExpression.expression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 2, EndLine: 1, EndColumn: 10},
				},
			},
			{
				// param position direct
				Code:    "class C { m(x, y) {} }",
				Options: []any{"Identifier:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// param position control
				Code:    "function m(x, y) {}",
				Options: []any{"Identifier:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				// dispatch virtual parent class explicit
				Code:    "class C {}",
				Options: []any{"ClassBody[parent.type='ClassDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 9, EndLine: 1, EndColumn: 11},
				},
			},
			{
				// jsx regex paired
				Code:    "const x = <Foo a b={c}></Foo>;",
				Options: []any{"[type=/^(JSXElement|JSXOpeningElement)$/]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 30},
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// fields do not panic
				Code:    "class C { m(a,b) {} } const x = <Foo a b={c} />;",
				Options: []any{".params"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				// fields do not panic
				Code:    "class C { m(a,b) {} } const x = <Foo a b={c} />;",
				Options: []any{":has(> .params)"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 20},
				},
			},
			{
				// fields do not panic
				Code:    "class C { m(a,b) {} } const x = <Foo a b={c} />;",
				Options: []any{"[params.length=2]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 20},
				},
			},
			{
				// program
				Code:    "if(a) f(); while(b) g();",
				Options: []any{"Program"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// program
				Code:    "if(a) f(); while(b) g();",
				Options: []any{"Program:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// program
				Code:    "if(a) f(); while(b) g();",
				Options: []any{":has(!IfStatement ~ WhileStatement)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// program
				Code:    "if(a) f(); while(b) g();",
				Options: []any{":has(IfStatement + !WhileStatement)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// program
				Code:    "if(a) f(); while(b) g();",
				Options: []any{"Program:has(Program > IfStatement)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// program
				Code:    "if(a) f(); while(b) g();",
				Options: []any{"[parent.type=\"Program\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// program empty
				Code:    "/* empty */",
				Options: []any{"Program"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 12},
				},
			},
			{
				// program empty
				Code:    "/* empty */",
				Options: []any{"Program:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 12},
				},
			},
			{
				// params
				Code:    "function f(a,b=1,...rest){} const g={m(a,b){}};",
				Options: []any{"FunctionExpression:has(> Identifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 39, EndLine: 1, EndColumn: 46},
				},
			},
			{
				// params
				Code:    "function f(a,b=1,...rest){} const g={m(a,b){}};",
				Options: []any{"Identifier.params"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
					{MessageId: "restrictedSyntax", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
				},
			},
			{
				// params
				Code:    "function f(a,b=1,...rest){} const g={m(a,b){}};",
				Options: []any{"Identifier:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
				},
			},
			{
				// params
				Code:    "function f(a,b=1,...rest){} const g={m(a,b){}};",
				Options: []any{"AssignmentPattern:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
				},
			},
			{
				// params
				Code:    "function f(a,b=1,...rest){} const g={m(a,b){}};",
				Options: []any{"RestElement:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
				},
			},
			{
				// params
				Code:    "function f(a,b=1,...rest){} const g={m(a,b){}};",
				Options: []any{"[parent.type=\"FunctionExpression\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
					{MessageId: "restrictedSyntax", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
					{MessageId: "restrictedSyntax", Line: 1, Column: 44, EndLine: 1, EndColumn: 46},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassDeclaration > Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"MethodDefinition:first-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"MethodDefinition:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"MethodDefinition + MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassBody[body.length=2]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 19, EndLine: 1, EndColumn: 37},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassDeclaration:has(ClassDeclaration > ClassBody)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassDeclaration[body.type=\"ClassBody\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"ClassBody.body"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 19, EndLine: 1, EndColumn: 37},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"MethodDefinition.body.body"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
					{MessageId: "restrictedSyntax", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
			},
			{
				// class boundaries
				Code:    "class C extends B {; m(){}; n(){}; }",
				Options: []any{"FunctionExpression.body.body.value"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
					{MessageId: "restrictedSyntax", Line: 1, Column: 30, EndLine: 1, EndColumn: 34},
				},
			},
			{
				// constructor
				Code:    "class C { constructor(x){} }",
				Options: []any{"Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// constructor
				Code:    "class C { constructor(x){} }",
				Options: []any{"MethodDefinition > Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// constructor
				Code:    "class C { constructor(x){} }",
				Options: []any{"Identifier.key"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// constructor
				Code:    "class C { constructor(x){} }",
				Options: []any{"MethodDefinition[key.name=\"constructor\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// constructor
				Code:    "class C { constructor(x){} }",
				Options: []any{"FunctionExpression:has(> Identifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// quoted constructor
				Code:    "class C { \"constructor\"(){} }",
				Options: []any{"Literal.key"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// quoted constructor
				Code:    "class C { \"constructor\"(){} }",
				Options: []any{"Literal[value=\"constructor\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// quoted constructor
				Code:    "class C { \"constructor\"(){} }",
				Options: []any{"MethodDefinition[key.type=\"Literal\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 28},
				},
			},
			{
				// generic method
				Code:    "class C { m<T extends { x: () => void }>(x:T): (() => void) {} }",
				Options: []any{"FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 63},
				},
			},
			{
				// generic method
				Code:    "class C { m<T extends { x: () => void }>(x:T): (() => void) {} }",
				Options: []any{"FunctionExpression > Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 42, EndLine: 1, EndColumn: 45},
				},
			},
			{
				// generic method
				Code:    "class C { m<T extends { x: () => void }>(x:T): (() => void) {} }",
				Options: []any{"FunctionExpression Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "restrictedSyntax", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "restrictedSyntax", Line: 1, Column: 42, EndLine: 1, EndColumn: 45},
					{MessageId: "restrictedSyntax", Line: 1, Column: 44, EndLine: 1, EndColumn: 45},
				},
			},
			{
				// generic method
				Code:    "class C { m<T extends { x: () => void }>(x:T): (() => void) {} }",
				Options: []any{"Identifier:has(> *)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 42, EndLine: 1, EndColumn: 45},
				},
			},
			{
				// bodyless
				Code:    "abstract class C { abstract m(x:string):void; m2(x:number):void; m2(x:any){} } interface I { get value():number; }",
				Options: []any{"FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 68, EndLine: 1, EndColumn: 77},
				},
			},
			{
				// bodyless
				Code:    "abstract class C { abstract m(x:string):void; m2(x:number):void; m2(x:any){} } interface I { get value():number; }",
				Options: []any{"TSEmptyBodyFunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 30, EndLine: 1, EndColumn: 46},
					{MessageId: "restrictedSyntax", Line: 1, Column: 49, EndLine: 1, EndColumn: 65},
				},
			},
			{
				// bodyless
				Code:    "abstract class C { abstract m(x:string):void; m2(x:number):void; m2(x:any){} } interface I { get value():number; }",
				Options: []any{"TSEmptyBodyFunctionExpression:has(> Identifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 30, EndLine: 1, EndColumn: 46},
					{MessageId: "restrictedSyntax", Line: 1, Column: 49, EndLine: 1, EndColumn: 65},
				},
			},
			{
				// bodyless
				Code:    "abstract class C { abstract m(x:string):void; m2(x:number):void; m2(x:any){} } interface I { get value():number; }",
				Options: []any{"TSAbstractMethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 20, EndLine: 1, EndColumn: 46},
				},
			},
			{
				// bodyless
				Code:    "abstract class C { abstract m(x:string):void; m2(x:number):void; m2(x:any){} } interface I { get value():number; }",
				Options: []any{"MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 47, EndLine: 1, EndColumn: 65},
					{MessageId: "restrictedSyntax", Line: 1, Column: 66, EndLine: 1, EndColumn: 77},
				},
			},
			{
				// bodyless
				Code:    "abstract class C { abstract m(x:string):void; m2(x:number):void; m2(x:any){} } interface I { get value():number; }",
				Options: []any{"FunctionExpression:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 68, EndLine: 1, EndColumn: 77},
				},
			},
			{
				// static block
				Code:    "class C { static { f(); } }",
				Options: []any{"StaticBlock > ExpressionStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// static block
				Code:    "class C { static { f(); } }",
				Options: []any{"StaticBlock:has(> ExpressionStatement)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXExpressionContainer.value"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"[value.type=\"JSXExpressionContainer\"]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXExpressionContainer:first-child"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 28, EndLine: 1, EndColumn: 35},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXSpreadChild:nth-child(2)"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 44},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXElement:nth-child(3)"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXOpeningElement[attributes.length=2]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 28},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXOpeningElement[selfClosing=false]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 28},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXSpreadChild[expression.type=\"Identifier\"]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 44},
				},
			},
			{
				// jsx edges
				Code:    "const x = <Foo.Bar a b={c}>{/*x*/}{...rest}<Bar />text</Foo.Bar>;",
				Options: []any{"JSXEmptyExpression.expression"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 29, EndLine: 1, EndColumn: 34},
				},
			},
			{
				// jsx fragment
				Code:    "const x = <><Foo />{a}<span /></>;",
				Options: []any{"JSXElement + JSXExpressionContainer"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 20, EndLine: 1, EndColumn: 23},
				},
			},
			{
				// jsx fragment
				Code:    "const x = <><Foo />{a}<span /></>;",
				Options: []any{"JSXElement ~ JSXElement"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 31},
				},
			},
			{
				// jsx fragment
				Code:    "const x = <><Foo />{a}<span /></>;",
				Options: []any{"JSXElement:first-child"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				// jsx fragment
				Code:    "const x = <><Foo />{a}<span /></>;",
				Options: []any{"JSXElement:last-child"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 31},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"ChainExpression > CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"[expression.type=\"CallExpression\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"[parent.type=\"ChainExpression\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"ChainExpression:first-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"ChainExpression:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"ChainExpression + CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 30, EndLine: 1, EndColumn: 33},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"CallExpression ~ ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"ChainExpression.elements"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
					{MessageId: "restrictedSyntax", Line: 1, Column: 35, EndLine: 1, EndColumn: 40},
				},
			},
			{
				// chain
				Code:    "a?.b?.(); const x = [a?.b(), f(), g?.()];",
				Options: []any{"CallExpression.elements"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 30, EndLine: 1, EndColumn: 33},
				},
			},
			{
				// attributes
				Code:    "f(); new F(); returnValue; function f(){ return x; }",
				Options: []any{"[expression.type=\"Identifier\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// attributes
				Code:    "f(); new F(); returnValue; function f(){ return x; }",
				Options: []any{"[body.type=\"BlockStatement\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 28, EndLine: 1, EndColumn: 53},
				},
			},
			{
				// parent roundtrip
				Code:    "const x = <Foo>{/*x*/}<Bar /></Foo>;",
				Options: []any{"JSXEmptyExpression[parent.expression.type=\"JSXEmptyExpression\"]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// parent roundtrip
				Code:    "const x = <Foo>{/*x*/}<Bar /></Foo>;",
				Options: []any{"JSXOpeningElement[parent.openingElement.type=\"JSXOpeningElement\"]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			{
				// parent roundtrip
				Code:    "const x = <Foo>{/*x*/}<Bar /></Foo>;",
				Options: []any{"JSXIdentifier[parent.type=\"JSXOpeningElement\"]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
					{MessageId: "restrictedSyntax", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			{
				// array index
				Code:    "f(a,b);",
				Options: []any{"CallExpression[arguments.1.name=\"b\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				// jsx identity fields
				Code:    "const x=<Foo a />;",
				Options: []any{"JSXOpeningElement[name]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// jsx identity fields
				Code:    "const x=<Foo a />;",
				Options: []any{"JSXOpeningElement[attributes.length=1]"},
				Tsx:     true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 9, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// static fields
				Code:    "class C { static { f(); g(); } }",
				Options: []any{"StaticBlock[body.length=2]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 31},
				},
			},
			{
				// static fields
				Code:    "class C { static { f(); g(); } }",
				Options: []any{"ExpressionStatement.body"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
					{MessageId: "restrictedSyntax", Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
				},
			},
			{
				// static fields
				Code:    "class C { static { f(); g(); } }",
				Options: []any{"ExpressionStatement:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
				},
			},
			{
				// quoted raw
				Code:    "class C { \"constructor\"(){} }",
				Options: []any{"Literal[raw=/constructor/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// nested chains
				Code:    "f?.(a?.b); obj?.[key?.x];",
				Options: []any{"ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
					{MessageId: "restrictedSyntax", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 25},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// nested chains
				Code:    "f?.(a?.b); obj?.[key?.x];",
				Options: []any{"ChainExpression > CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				// nested chains
				Code:    "f?.(a?.b); obj?.[key?.x];",
				Options: []any{"CallExpression > ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
				},
			},
			{
				// nested chains
				Code:    "f?.(a?.b); obj?.[key?.x];",
				Options: []any{"MemberExpression > ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				// computed key
				Code:    "class C { [name()](){} }",
				Options: []any{"MethodDefinition > CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// computed key
				Code:    "class C { [name()](){} }",
				Options: []any{"CallExpression.key"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 18},
				},
			},
			{
				// computed key
				Code:    "class C { [name()](){} }",
				Options: []any{"MethodDefinition[key.type=\"CallExpression\"]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 23},
				},
			},
			{
				// chain subjects
				Code:    "const x=[a?.b(), f()];",
				Options: []any{":matches(:not(*), !ChainExpression ~ CallExpression)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 21},
				},
			},
			{
				// chain subjects
				Code:    "const x=[a?.b(), f()];",
				Options: []any{":matches(:not(*), ChainExpression + !CallExpression)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 21},
				},
			},
		})
}

// TypeScript assertion expressions are their own ESTree node types:
// TSAsExpression, TSSatisfiesExpression, TSNonNullExpression and
// TSTypeAssertion. ESTree drops parentheses, so a parenthesized inner
// assertion is still a direct child of the outer one. A trailing non-null
// assertion on an optional chain (`o?.x!`) is inside the ChainExpression.
func TestNoRestrictedSyntaxTSAssertionExpressions(t *testing.T) {
	const assertions = "declare const v: unknown;\nconst a = v as string;\nconst b = v!;\nconst c = v satisfies unknown;\nconst d = <string>v;\n"
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{
				// A single cast has no assertion child.
				Code:    "declare const v: unknown;\nconst a = v as string;\n",
				Options: []any{"TSAsExpression > TSAsExpression"},
			},
			{
				// A trailing non-null assertion stays inside the optional chain.
				Code:    "declare const o: { x?: { y: number } } | undefined;\no?.x!;\n",
				Options: []any{"TSNonNullExpression > ChainExpression"},
			},
			{
				// Parentheses end the chain before the assertion.
				Code:    "declare const o: { x?: { y: number } } | undefined;\n(o?.x)!;\n",
				Options: []any{"ChainExpression > TSNonNullExpression"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    assertions,
				Options: []any{"TSAsExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'TSAsExpression' is not allowed.", Line: 2, Column: 11, EndLine: 2, EndColumn: 22},
				},
			},
			{
				Code:    assertions,
				Options: []any{"TSNonNullExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 3, Column: 11, EndLine: 3, EndColumn: 13},
				},
			},
			{
				Code:    assertions,
				Options: []any{"TSSatisfiesExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 4, Column: 11, EndLine: 4, EndColumn: 30},
				},
			},
			{
				Code:    assertions,
				Options: []any{"TSTypeAssertion"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 5, Column: 11, EndLine: 5, EndColumn: 20},
				},
			},
			{
				Code:    "const t = [1, 2] as const;\n",
				Options: []any{"TSAsExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 26},
				},
			},
			{
				// Double cast: the inner assertion is the child.
				Code: "declare const v: unknown;\nconst e = v as unknown as string;\n",
				Options: []any{map[string]any{
					"selector": "TSAsExpression > TSAsExpression",
					"message":  "Do not double cast.",
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Do not double cast.", Line: 2, Column: 11, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code:    "declare const v: unknown;\nconst e = (v as unknown) as string;\n",
				Options: []any{"TSAsExpression > TSAsExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 12, EndLine: 2, EndColumn: 24},
				},
			},
			{
				// Double cast seen from the outer assertion.
				Code:    "declare const v: unknown;\nconst e = v as unknown as string;\n",
				Options: []any{"TSAsExpression[expression.type='TSAsExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 11, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:    "declare const o: { x?: { y: number } };\no.x!.y;\n",
				Options: []any{"MemberExpression > TSNonNullExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
				},
			},
			{
				Code:    "declare const o: { x?: { y: number } } | undefined;\no?.x!;\n",
				Options: []any{"ChainExpression > TSNonNullExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			{
				Code:    "declare const o: { x?: { y: number } } | undefined;\no?.x!;\n",
				Options: []any{"ExpressionStatement > ChainExpression > TSNonNullExpression > MemberExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 5},
				},
			},
			{
				Code:    "declare const o: { x?: { y: number } } | undefined;\n(o?.x)!;\n",
				Options: []any{"TSNonNullExpression > ChainExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 2, EndLine: 2, EndColumn: 6},
				},
			},
			{
				Code:    "declare const o: { x?: { y: number } } | undefined;\no?.x!.y;\n",
				Options: []any{"ChainExpression > MemberExpression > TSNonNullExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			{
				Code:    "declare const o: { x?: { y: number } } | undefined;\no?.x!!;\n",
				Options: []any{"ChainExpression > TSNonNullExpression > TSNonNullExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
		})
}

// Attribute paths look through assertions, but a trailing non-null assertion
// on an optional chain (`o?.x!`) is the ChainExpression itself in ESTree, so
// the field must resolve to ChainExpression rather than the inner member.
func TestNoRestrictedSyntaxNonNullChainFields(t *testing.T) {
	const decl = "declare const o: { x?: { y: number }; f?: () => number } | undefined;\n"
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{Code: decl + "const a = o?.x!;\n", Options: []any{"VariableDeclarator[init.type='MemberExpression']"}},
			{Code: decl + "function f() { return o?.x!; }\n", Options: []any{"ReturnStatement[argument.type='MemberExpression']"}},
			{Code: decl + "(o?.f!)();\n", Options: []any{"CallExpression[callee.type='MemberExpression']"}},
			{Code: decl + "(o?.x!).y;\n", Options: []any{"MemberExpression[object.type='MemberExpression']"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    decl + "const a = o?.x!;\n",
				Options: []any{"VariableDeclarator[init.type='ChainExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 7, EndLine: 2, EndColumn: 16},
				},
			},
			{
				Code:    decl + "function f() { return o?.x!; }\n",
				Options: []any{"ReturnStatement[argument.type='ChainExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 16, EndLine: 2, EndColumn: 29},
				},
			},
			{
				Code:    decl + "(o?.f!)();\n",
				Options: []any{"CallExpression[callee.type='ChainExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 10},
				},
			},
			{
				Code:    decl + "(o?.x!).y;\n",
				Options: []any{"MemberExpression[object.type='ChainExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 10},
				},
			},
		})
}

// Import specifiers sit directly in ImportDeclaration.specifiers in ESTree:
// tsgo's ImportClause (when nameless or when climbing from named/namespace
// bindings) and NamedImports are transparent containers. When a default
// binding is present, ImportClause represents only ImportDefaultSpecifier and
// spans the default binding identifier alone.
// Expectations checked against ESLint 10.13.0 and @typescript-eslint/parser 8.65.0.
func TestNoRestrictedSyntaxImportSpecifiers(t *testing.T) {
	const issue2631 = "import * as Lib from 'some-lib'\nimport { x } from 'other-lib'\nexport const l = [Lib, x]\n"
	const defaultAndNamespace = "import Def, * as NS from 'lib';\n"
	const defaultAndNamed = "import Def, { x, y as z } from 'lib';\n"

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			// Nameless ImportClause must not match ImportDefaultSpecifier or act as a parent.
			{Code: issue2631, Options: []any{"ImportDefaultSpecifier"}},
			{Code: issue2631, Options: []any{"ImportDefaultSpecifier > ImportNamespaceSpecifier"}},
			{Code: issue2631, Options: []any{"ImportDefaultSpecifier > ImportSpecifier"}},
			{Code: "import type * as NS from 't2';\n", Options: []any{"ImportDefaultSpecifier"}},
			{Code: "import type { T } from 't3';\n", Options: []any{"ImportDefaultSpecifier"}},
			{Code: "import {} from 'empty';\n", Options: []any{"ImportDefaultSpecifier", "ImportSpecifier", "ImportDeclaration:has(> :not(Literal))"}},
			{Code: "import 'side-effect';\n", Options: []any{"ImportDefaultSpecifier", "ImportNamespaceSpecifier", "ImportSpecifier"}},

			// Combined default + namespace: namespace specifier is a sibling, not a child of default.
			{Code: defaultAndNamespace, Options: []any{"ImportDefaultSpecifier > ImportNamespaceSpecifier"}},
			{Code: defaultAndNamespace, Options: []any{"ImportDefaultSpecifier ImportNamespaceSpecifier"}},
			{Code: defaultAndNamespace, Options: []any{"ImportDefaultSpecifier:has(ImportNamespaceSpecifier)"}},
			{Code: defaultAndNamespace, Options: []any{"ImportDefaultSpecifier:has(> ImportNamespaceSpecifier)"}},
			{Code: defaultAndNamespace, Options: []any{"ImportDefaultSpecifier:has(Identifier[name='NS'])"}},
			{Code: defaultAndNamespace, Options: []any{"ImportNamespaceSpecifier:first-child"}},
			{Code: defaultAndNamespace, Options: []any{"ImportDeclaration > Identifier"}},

			// Combined default + named: named specifiers are siblings of default, not children.
			{Code: defaultAndNamed, Options: []any{"ImportDefaultSpecifier > ImportSpecifier"}},
			{Code: defaultAndNamed, Options: []any{"ImportDefaultSpecifier ImportSpecifier"}},
			{Code: defaultAndNamed, Options: []any{"ImportDefaultSpecifier:has(ImportSpecifier)"}},
			{Code: defaultAndNamed, Options: []any{"ImportDefaultSpecifier:has(Identifier[name='x'])"}},
			{Code: defaultAndNamed, Options: []any{"ImportSpecifier:first-child"}},
			{Code: defaultAndNamed, Options: []any{"ImportSpecifier.elements"}},

			// Specifiers and import attributes belong to separate arrays.
			{Code: "import { x } from 'str' with { type: 'json' };\n", Options: []any{"ImportSpecifier + ImportAttribute"}},

			// Reverse sibling-subject selectors must not strip right-hand attributes.
			{Code: "import { y, z } from 'a';\n", Options: []any{"!ImportSpecifier ~ ImportSpecifier[local.name='y']"}},
			{Code: "import { y, z } from 'a';\n", Options: []any{"ImportSpecifier + !ImportSpecifier[local.name='y']"}},
			{Code: "[a + b, c * d];\n", Options: []any{"!BinaryExpression ~ BinaryExpression[operator='+']"}},
			{Code: "[a && b, c + d];\n", Options: []any{"!LogicalExpression ~ :matches(BinaryExpression):matches(LogicalExpression)"}},

			// importKind and phase negative filters must not false-positive.
			{Code: "import type Def from 'a';\n", Options: []any{"ImportDeclaration[importKind!='type']"}},
			{Code: "import Def from 'a';\n", Options: []any{"ImportDeclaration[importKind='type']", "ImportDeclaration[phase!=null]", "ImportDefaultSpecifier[importKind]"}},
			{Code: "import * as NS from 'a';\n", Options: []any{"ImportNamespaceSpecifier[importKind]"}},
			{Code: "import { type A } from 'a';\n", Options: []any{"ImportSpecifier[importKind!='type']"}},
			{Code: "import type { T } from 'a';\n", Options: []any{"ImportSpecifier[importKind='type']"}},
			{Code: "import defer * as NS from 'a';\n", Options: []any{"ImportDeclaration[phase!='defer']"}},

			// Symmetric export specifier and namespace export boundaries.
			{Code: "export { x, y as z } from 'a';\n", Options: []any{"ExportSpecifier.elements", "ExportNamedDeclaration[kind='value']"}},
			{Code: "export * from 'a';\n", Options: []any{"ExportAllDeclaration[kind='value']", "ExportAllDeclaration[specifiers]"}},
			{Code: "export * as ns from 'lib';\n", Options: []any{"Identifier[parent.specifiers]", "ExportAllDeclaration[specifiers]"}},
			{Code: "export {} from 'a';\n", Options: []any{"ExportNamedDeclaration:has(> :not(Literal))"}},
			{Code: "export * as \"ns\" from 'a';\n", Options: []any{"ExportAllDeclaration:has(> :not(Literal))"}},
		},
		[]rule_tester.InvalidTestCase{
			// Issue #2631 exact reproduction with multiple selectors in specificity order.
			{
				Code: issue2631,
				Options: []any{
					map[string]any{"selector": `ImportDeclaration[source.value="some-lib"]`, "message": "A"},
					map[string]any{"selector": `ImportDeclaration > ImportNamespaceSpecifier`, "message": "B"},
					map[string]any{"selector": `ImportNamespaceSpecifier[parent.type="ImportDeclaration"]`, "message": "C"},
					map[string]any{"selector": `ImportDeclaration > ImportSpecifier`, "message": "D"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "A", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Message: "B", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Message: "C", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Message: "D", Line: 2, Column: 10, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code:    issue2631,
				Options: []any{`ImportNamespaceSpecifier[parent.source.value="some-lib"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    issue2631,
				Options: []any{`ImportSpecifier[parent.type="ImportDeclaration"][parent.source.value="other-lib"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 10, EndLine: 2, EndColumn: 11},
				},
			},
			{
				Code:    "import * as Lib from 'some-lib';\n",
				Options: []any{"[type='ImportNamespaceSpecifier']", "ImportDeclaration > ImportNamespaceSpecifier:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
				},
			},
			{
				// Bare wildcard visits Program plus every ESTree node in canonical order.
				Code:    "import * as Lib from 'some-lib';\n",
				Options: []any{"*"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 32},
				},
			},

			// Combined default + namespace: parent, range, siblings, fields, and :has().
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDeclaration > ImportDefaultSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDeclaration > ImportNamespaceSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier + ImportNamespaceSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportNamespaceSpecifier[parent.type='ImportDeclaration'][parent.source.value='lib']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier[parent.type='ImportDeclaration'][parent.source.value='lib']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier:first-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportNamespaceSpecifier:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportNamespaceSpecifier:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDeclaration:has(> ImportDefaultSpecifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDeclaration:has(> ImportNamespaceSpecifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier > Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"Identifier[parent.type='ImportDefaultSpecifier']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"Identifier[parent.type='ImportNamespaceSpecifier']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"[parent.type='ImportDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
					{MessageId: "restrictedSyntax", Line: 1, Column: 26, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDeclaration[specifiers.0.type='ImportDefaultSpecifier'][specifiers.1.type='ImportNamespaceSpecifier']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"Identifier.specifiers.local"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},

			// Combined default + named: parent, siblings, indices, fields, and :has().
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportDeclaration > ImportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportDefaultSpecifier + ImportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportDefaultSpecifier ~ ImportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportSpecifier + ImportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportSpecifier:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportSpecifier:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportDeclaration:has(> ImportSpecifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportSpecifier.specifiers"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportDefaultSpecifier.specifiers"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code:    "import Def, { a as b, c as d } from 'lib';\n",
				Options: []any{"Identifier[parent.type='ImportSpecifier']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
					{MessageId: "restrictedSyntax", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
				},
			},
			{
				// Shorthand ImportSpecifier (`{ x }`) emits both `imported` and `local`
				// Identifier nodes in ESTree, matching upstream ESLint.
				Code:    defaultAndNamed,
				Options: []any{"Identifier[parent.type='ImportSpecifier']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"ImportDeclaration[specifiers.1.type='ImportSpecifier']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},

			// Type-only imports, empty named imports, string-literal import name, and attributes.
			{
				Code:    "import type Def from 't1';\n",
				Options: []any{"ImportDeclaration > ImportDefaultSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "import type * as NS from 't2';\n",
				Options: []any{"ImportDeclaration > ImportNamespaceSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "import type { T } from 't3';\n",
				Options: []any{"ImportDeclaration > ImportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "import {} from 'empty';\n",
				Options: []any{"[parent.type='ImportDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `import { "a" as b } from 'str' with { type: 'json' };` + "\n",
				Options: []any{`ImportSpecifier[parent.source.value="str"]`, "ImportDeclaration > ImportAttribute"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
					{MessageId: "restrictedSyntax", Line: 1, Column: 39, EndLine: 1, EndColumn: 51},
				},
			},

			// :exit order on combined imports must exit ImportDefaultSpecifier
			// right after its binding identifier and before sibling bindings enter.
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier:exit", "ImportNamespaceSpecifier:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"Identifier:exit", "ImportDefaultSpecifier:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamespace,
				Options: []any{"ImportDefaultSpecifier:exit", "ImportNamespaceSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    defaultAndNamed,
				Options: []any{"[parent.type='ImportDeclaration']:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
					{MessageId: "restrictedSyntax", Line: 1, Column: 32, EndLine: 1, EndColumn: 37},
				},
			},

			// Reverse sibling-subject selectors with typed right-hand attributes.
			{
				Code:    "import { x, y } from 'a';\n",
				Options: []any{"!ImportSpecifier ~ ImportSpecifier[local.name='y']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "import { x, y } from 'a';\n",
				Options: []any{"ImportSpecifier + !ImportSpecifier[local.name='y']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "[a * b, c + d];\n",
				Options: []any{"!BinaryExpression ~ BinaryExpression[operator='+']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 2, EndLine: 1, EndColumn: 7},
					{MessageId: "restrictedSyntax", Line: 1, Column: 9, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ BinaryExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ :matches(BinaryExpression)[operator='+']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ :matches(BinaryExpression, CallExpression)[operator='+']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ :is(BinaryExpression)[operator='+']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ [operator='+']:matches(BinaryExpression)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{":matches(!LogicalExpression ~ BinaryExpression[operator='+'], Identifier[name='missing'])"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ BinaryExpression[operator='+'], Identifier[name='missing']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ :matches(BinaryExpression, LogicalExpression)[operator='+']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{"!LogicalExpression ~ :matches(BinaryExpression, [operator='+'])"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "[a && b, c + d];\n",
				Options: []any{":matches(!LogicalExpression ~ BinaryExpression[operator='+'], LogicalExpression[operator='||'])"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
				},
			},

			// importKind and phase attributes.
			{
				Code:    "import type Def from 'a';\nimport Def2 from 'b';\n",
				Options: []any{"ImportDeclaration[importKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "import type { T } from 'a';\nimport { type A, B } from 'b';\n",
				Options: []any{"ImportSpecifier[importKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 10, EndLine: 2, EndColumn: 16},
				},
			},
			{
				Code:    "import type { T } from 'a';\n",
				Options: []any{"ImportSpecifier[importKind='value']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "import defer * as NS from 'a';\nimport * as NS2 from 'b';\n",
				Options: []any{"ImportDeclaration[phase='defer']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "import defer * as NS from 'a';\nimport * as NS2 from 'b';\n",
				Options: []any{"ImportDeclaration[phase=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
				},
			},

			// Contextual keywords, comments, and default + empty braces.
			{
				Code:    "import type from 'a';\nimport type type from 'b';\nimport /* c1 */ Def /* c2 */, {} from 'c';\n",
				Options: []any{"ImportDeclaration > ImportDefaultSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 12},
					{MessageId: "restrictedSyntax", Line: 2, Column: 13, EndLine: 2, EndColumn: 17},
					{MessageId: "restrictedSyntax", Line: 3, Column: 17, EndLine: 3, EndColumn: 20},
				},
			},

			// Symmetric export specifier and namespace export selectors.
			{
				Code:    "export { x, y as z } from 'a';\n",
				Options: []any{"ExportNamedDeclaration > ExportSpecifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export { x, y as z } from 'a';\n",
				Options: []any{"ExportSpecifier[parent.type='ExportNamedDeclaration'][parent.source.value='a']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export { x, y as z } from 'a';\n",
				Options: []any{"ExportSpecifier.specifiers"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export { x, y as z } from 'a';\n",
				Options: []any{"ExportSpecifier + ExportSpecifier", "ExportNamedDeclaration:has(> ExportSpecifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export * as ns from 'a';\n",
				Options: []any{"ExportAllDeclaration > Identifier.exported", "Identifier[parent.type='ExportAllDeclaration']", "ExportAllDeclaration:has(> Identifier)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 15},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    "export * as \"ns\" from 'a';\n",
				Options: []any{"ExportAllDeclaration > Literal.exported", "ExportAllDeclaration[specifiers=undefined]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
				},
			},
		})
}

// Issue #2626 and full TSESTree node/attribute coverage: type keywords,
// typeAnnotation filters (positive and negative), assertion parent fields,
// declarations, generics, signatures, abstract properties, templates, regexes,
// and graceful handling of unsupported syntax.
func TestNoRestrictedSyntaxTypeScriptNodes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			// Negative typeAnnotation filters must not false-positive when matching.
			{Code: "declare const v: unknown;\nconst a = v as unknown;\n", Options: []any{`TSAsExpression[typeAnnotation.type!="TSUnknownKeyword"]`}},
			{Code: "const a = { x: 1 } as const;\n", Options: []any{`TSAsExpression[typeAnnotation.typeName.name!="const"]`}},
			{Code: "declare const v: string;\nconst a = v as string;\n", Options: []any{`TSAsExpression > TSAsExpression[typeAnnotation.type="TSUnknownKeyword"]`}},
			// Parent attribute paths do not look through TS assertions.
			{Code: "declare const v: unknown;\nconst a = v as string;\n", Options: []any{`VariableDeclarator[init.name="v"]`}},
			{Code: "declare const o: { x?: { y: number } };\no.x!.y;\n", Options: []any{`MemberExpression[object.type="MemberExpression"]`}},
			// Abstract property is TSAbstractPropertyDefinition, not PropertyDefinition.
			{Code: "abstract class C { abstract x: number; }\n", Options: []any{`PropertyDefinition`}},
			// Absent optional TS fields are undefined, not null.
			{Code: "function f(x) {}\n", Options: []any{`FunctionDeclaration[returnType=null]`}},
			{Code: "function f(x) {}\n", Options: []any{`Identifier[typeAnnotation=null]`}},
			{Code: "class C { m() {} }\n", Options: []any{`MethodDefinition[accessibility=null]`}},
			{Code: "foo();\n", Options: []any{`CallExpression[typeArguments=null]`}},
			// Valid ES2024 /v regex literal is not null.
			{Code: "const r = /[a-z]/v;\n", Options: []any{`Literal[value=null]`}},
		},
		[]rule_tester.InvalidTestCase{
			// Unsupported JSDoc nullable type authored in TS degrades locally
			// without hiding surrounding declarations or statements.
			{
				Code:    "type Bad = ?number;\n",
				Options: []any{`TSTypeAliasDeclaration`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20}},
			},
			// Exact issue #2626 reproduction.
			{
				Code: "export const v: unknown = 1\nexport const a = v as string\nexport const b = v!\nexport const c = v satisfies unknown\nenum E { A }\nexport { E }\n",
				Options: []any{
					"TSAsExpression",
					"TSNonNullExpression",
					"TSSatisfiesExpression",
					"TSUnknownKeyword",
					"TSEnumDeclaration",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Message: "Using 'TSUnknownKeyword' is not allowed.", Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
					{MessageId: "restrictedSyntax", Message: "Using 'TSAsExpression' is not allowed.", Line: 2, Column: 18, EndLine: 2, EndColumn: 29},
					{MessageId: "restrictedSyntax", Message: "Using 'TSNonNullExpression' is not allowed.", Line: 3, Column: 18, EndLine: 3, EndColumn: 20},
					{MessageId: "restrictedSyntax", Message: "Using 'TSSatisfiesExpression' is not allowed.", Line: 4, Column: 18, EndLine: 4, EndColumn: 37},
					{MessageId: "restrictedSyntax", Message: "Using 'TSUnknownKeyword' is not allowed.", Line: 4, Column: 30, EndLine: 4, EndColumn: 37},
					{MessageId: "restrictedSyntax", Message: "Using 'TSEnumDeclaration' is not allowed.", Line: 5, Column: 1, EndLine: 5, EndColumn: 13},
				},
			},
			// Issue #2626 double cast selector `as unknown as T`.
			{
				Code:    "declare const v: number;\nexport const a = v as unknown as string;\n",
				Options: []any{`TSAsExpression > TSAsExpression[typeAnnotation.type="TSUnknownKeyword"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 18, EndLine: 2, EndColumn: 30},
				},
			},
			// Disallow `as` except `as const`.
			{
				Code:    "const ok = { x: 1 } as const;\nconst bad = (1 as any);\n",
				Options: []any{`TSAsExpression[typeAnnotation.typeName.name!="const"]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 14, EndLine: 2, EndColumn: 22},
				},
			},
			// Parent attribute and field class selectors see assertions directly.
			{
				Code:    "declare const v: unknown;\nconst a = v as string;\n",
				Options: []any{`VariableDeclarator[init.type="TSAsExpression"]`, `VariableDeclarator > TSAsExpression.init`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 7, EndLine: 2, EndColumn: 22},
					{MessageId: "restrictedSyntax", Line: 2, Column: 11, EndLine: 2, EndColumn: 22},
				},
			},
			// TSTypeAnnotation wrapper and exact parent-child topology.
			{
				Code:    "const v: unknown = 1;\n",
				Options: []any{`Identifier > TSTypeAnnotation > TSUnknownKeyword`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// TS declarations, bodies, generics, instantiation, abstract properties, accessors, and module/export forms.
			{
				Code: "export interface I<T> { x: T; }\nexport type Alias<U> = U | number;\nnamespace N { export const x = 1; }\nconst inst = fn<string>;\nabstract class C { abstract p: number; accessor a = 1; }\n",
				Options: []any{
					"TSInterfaceDeclaration",
					"TSTypeAliasDeclaration",
					"TSTypeParameterDeclaration",
					"TSUnionType",
					"TSModuleDeclaration",
					"TSInstantiationExpression",
					"TSAbstractPropertyDefinition",
					"AccessorProperty",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 32},
					{MessageId: "restrictedSyntax", Line: 1, Column: 19, EndLine: 1, EndColumn: 22},
					{MessageId: "restrictedSyntax", Line: 2, Column: 8, EndLine: 2, EndColumn: 35},
					{MessageId: "restrictedSyntax", Line: 2, Column: 18, EndLine: 2, EndColumn: 21},
					{MessageId: "restrictedSyntax", Line: 2, Column: 24, EndLine: 2, EndColumn: 34},
					{MessageId: "restrictedSyntax", Line: 3, Column: 1, EndLine: 3, EndColumn: 36},
					{MessageId: "restrictedSyntax", Line: 4, Column: 14, EndLine: 4, EndColumn: 24},
					{MessageId: "restrictedSyntax", Line: 5, Column: 20, EndLine: 5, EndColumn: 39},
					{MessageId: "restrictedSyntax", Line: 5, Column: 40, EndLine: 5, EndColumn: 55},
				},
			},
			// TemplateElement.value.{cooked,raw} and RegExp Literal.value (canonical sorted flags + properties).
			{
				Code:    "const s = `hello`;\nconst r = /abc/ig;\n",
				Options: []any{`TemplateElement[value.raw="hello"]`, `Literal[value="/abc/gi"][value.ignoreCase=true]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
					{MessageId: "restrictedSyntax", Line: 2, Column: 11, EndLine: 2, EndColumn: 18},
				},
			},
			// Sparse array holes preserve index and break adjacent `+` while allowing `~`.
			{
				Code:    "const a = [1, , 3];\n",
				Options: []any{`Literal:nth-child(3)`, `Literal ~ Literal`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// Program.parent is null (not undefined).
			{
				Code:    "const x = 1;\n",
				Options: []any{`Program[parent=null]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
				},
			},
			// Absent optional TS fields compare equal to undefined.
			{
				Code:    "function f(x) {}\n",
				Options: []any{`FunctionDeclaration[returnType=undefined]`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
				},
			},
		})
}

func BenchmarkNoRestrictedSyntaxLargeSpecifierLists(b *testing.B) {
	var names strings.Builder
	for i := range 1000 {
		if i > 0 {
			names.WriteString(", ")
		}
		names.WriteByte('a')
		names.WriteString(strconv.Itoa(i))
	}
	list := names.String()

	var manyImports strings.Builder
	for i := range 100 {
		manyImports.WriteString("import { a")
		manyImports.WriteString(strconv.Itoa(i))
		manyImports.WriteString(" } from 'lib';\n")
	}

	cases := []struct {
		name     string
		code     string
		selector string
		kind     ast.Kind
		want     int
	}{
		{
			name:     "ExportNamed1000",
			code:     "export { " + list + " } from 'lib';\n",
			selector: "ExportSpecifier:first-child",
			kind:     ast.KindExportSpecifier,
			want:     1000,
		},
		{
			name:     "ImportNamed1000",
			code:     "import { " + list + " } from 'lib';\n",
			selector: "ImportSpecifier:first-child",
			kind:     ast.KindImportSpecifier,
			want:     1000,
		},
		{
			name:     "ImportDefaultAndNamed1000",
			code:     "import Def, { " + list + " } from 'lib';\n",
			selector: "ImportSpecifier:first-child",
			kind:     ast.KindImportSpecifier,
			want:     1000,
		},
		{
			name:     "ImportNamed100Declarations",
			code:     manyImports.String(),
			selector: "ImportDeclaration[specifiers.length=0]",
			kind:     ast.KindImportDeclaration,
			want:     100,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			sf := tsparser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/bench.ts",
				Path:     "/bench.ts",
			}, tc.code, core.ScriptKindTS)
			compiled, err := parseSelector(tc.selector)
			if err != nil {
				b.Fatalf("parseSelector(%q): %v", tc.selector, err)
			}
			var targetType estree.TypeID
			switch tc.kind {
			case ast.KindExportSpecifier:
				targetType, _ = estree.TypeIDByName("ExportSpecifier")
			case ast.KindImportSpecifier:
				targetType, _ = estree.TypeIDByName("ImportSpecifier")
			case ast.KindImportDeclaration:
				targetType, _ = estree.TypeIDByName("ImportDeclaration")
			}

			var builder estree.Builder
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				count := 0
				if err := builder.WithTree(sf, estree.TreeOptions{SourceType: "module"}, func(tree *estree.Tree) {
					var walk func(estree.NodeID)
					walk = func(id estree.NodeID) {
						if tree.Type(id) == targetType {
							count++
							matchesSelector(compiled, tree, id, estree.NoNode, nil)
						}
						tree.ForEachChild(id, func(child estree.NodeID) bool {
							walk(child)
							return false
						})
					}
					walk(tree.Root())
				}); err != nil {
					b.Fatal(err)
				}
				if count != tc.want {
					b.Fatalf("expected %d targets, got %d", tc.want, count)
				}
			}
		})
	}
}

func BenchmarkNoRestrictedSyntaxEndToEnd(b *testing.B) {
	normalTS := strings.Repeat("export function f(a: number): number { return a + 1; }\n", 1000)
	asCastTS := normalTS + "export const x = (1 as unknown as string);\n"
	manyStmtsTS := "function big() {\n" + strings.Repeat("const a = 1;\n", 1000) + "return 1;\n}\n"

	multi20 := []any{
		`CallExpression[callee.name='eval']`,
		`CallExpression[callee.name='Function']`,
		`CallExpression[callee.name='setTimeout']`,
		`CallExpression[callee.name='setInterval']`,
		`CallExpression[callee.property.name='log']`,
		`CallExpression[callee.property.name='warn']`,
		`CallExpression[callee.property.name='error']`,
		`CallExpression[callee.property.name='debug']`,
		`MemberExpression[property.name='__proto__']`,
		`MemberExpression[property.name='prototype']`,
		`MemberExpression[property.name='constructor']`,
		`MemberExpression[property.name='caller']`,
		`Literal[value='lodash']`,
		`Literal[value='moment']`,
		`Literal[value='jquery']`,
		`Literal[value='underscore']`,
		`TSAsExpression > TSAsExpression[typeAnnotation.type='TSUnknownKeyword']`,
		`TSAsExpression[typeAnnotation.type='TSAnyKeyword']`,
		`WithStatement`,
		`LabeledStatement`,
	}

	cases := []struct {
		name    string
		code    string
		options []any
		want    int
	}{
		{
			name:    "LazySkip_TSAsExpression",
			code:    normalTS,
			options: []any{`TSAsExpression > TSAsExpression[typeAnnotation.type='TSUnknownKeyword']`},
			want:    0,
		},
		{
			name:    "Match_TSAsExpression_2626",
			code:    asCastTS,
			options: []any{`TSAsExpression > TSAsExpression[typeAnnotation.type='TSUnknownKeyword']`},
			want:    1,
		},
		{
			name:    "MultiSelector_20Rules_Dispatch",
			code:    asCastTS,
			options: multi20,
			want:    1,
		},
		{
			name:    "MemoizedSibling_1000Stmts",
			code:    manyStmtsTS,
			options: []any{`VariableDeclaration ~ ReturnStatement`},
			want:    1,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			sf := tsparser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/bench.ts",
				Path:     "/bench.ts",
			}, tc.code, core.ScriptKindTS)
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				reports := 0
				ctx := rule.RuleContext{
					SourceFile: sf,
				}.WithReporter("no-restricted-syntax", rule.SeverityError, func(rule.RuleDiagnostic) {
					reports++
				})
				listeners := NoRestrictedSyntaxRule.Run(ctx, tc.options)
				if len(listeners) > 0 {
					var visit func(*ast.Node) bool
					visit = func(n *ast.Node) bool {
						if fn := listeners[n.Kind]; fn != nil {
							fn(n)
						}
						n.ForEachChild(visit)
						return false
					}
					visit(sf.AsNode())
				}
				if reports != tc.want {
					b.Fatalf("expected %d reports, got %d", tc.want, reports)
				}
			}
		})
	}
}
