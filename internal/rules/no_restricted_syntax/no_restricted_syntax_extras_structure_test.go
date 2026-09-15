package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Regression expectations were checked against ESLint 10.10.0, esquery 1.7.0,
// and @typescript-eslint/parser 8.69.0. These cases exercise the tsgo facade.
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
				// dispatch virtual parent
				Code:    "class C { m() {} }",
				Options: []any{"[parent.type='MethodDefinition']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 12, EndLine: 1, EndColumn: 17},
					{MessageId: "restrictedSyntax", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
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
				// dispatch virtual parent class
				Code:    "class C {}",
				Options: []any{"[parent.type='ClassDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 9, EndLine: 1, EndColumn: 11},
					{MessageId: "restrictedSyntax", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
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
