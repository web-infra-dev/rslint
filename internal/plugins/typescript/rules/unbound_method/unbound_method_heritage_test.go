package unbound_method

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUnboundMethodHeritage(t *testing.T) {
	// The upstream rule also checks heritage members in incomplete or invalid
	// types. A compiler AST change must not silently remove those diagnostics.
	var invalid []rule_tester.InvalidTestCase
	for _, declaration := range []struct {
		code   string
		member string
		width  int
	}{
		{"class Cls { static method() {} }", "Cls.method", 10},
		{"declare const Cls: { method(): void };", "Cls.method", 10},
		{"namespace NS { export class Cls { static method() {} } }", "NS.Cls.method", 13},
		{"namespace NS { export class Base { static method() {} } } import Cls = NS.Base;", "Cls.method", 10},
		{"export {}; const Math = { round() {} };", "Math.round", 10},
		{"const alias = console;", "alias.log", 9},
	} {
		for _, context := range []struct {
			code   string
			column int
		}{
			{"interface I extends ", 21},
			{"class C implements ", 20},
		} {
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: declaration.code + "\n" + context.code + declaration.member + " {}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unboundWithoutThisAnnotation",
					Message:   "Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.",
					Line:      2, Column: context.column, EndLine: 2, EndColumn: context.column + declaration.width,
				}},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnboundMethodRule,
		[]rule_tester.ValidTestCase{
			{Code: `namespace NS { export class Base {} } interface I extends NS.Base {}`},
			{Code: `namespace NS { export class Base {} } class C implements NS.Base {}`},
			{Code: `class Cls { static method() {} } type I = Cls.method;`},
			{Code: `class Cls { static method() {} } interface I extends Box<Cls.method> {}`},
			{Code: `class Cls { static method(this: void) {} } interface I extends Cls.method {}`},
			{Code: `class Cls { static method() {} } class C implements Cls.method {}`, Options: map[string]any{"ignoreStatic": true}},
			{Code: `interface I extends console.log {}`},
			{Code: `class C implements Math.round {}`},
		}, invalid)
}
