package no_unused_vars

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsReferenceIndexAdversarial(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{
			Code: `export {};
const value = 1;
function consume(value: number) { return value; }
console.log(value, consume(2));`,
		},
		{
			Code: `export namespace Outer {
  const value = 1;
  export { value };
}`,
		},
		{
			Code: `interface Model { value: number }
const Model = { value: 1 };
export { Model };`,
		},
		{
			Code: `namespace Empty {}
export const value = Empty;`,
		},
		{
			Code: `namespace EmptyA {}
namespace EmptyB {}
export const values = [EmptyA, EmptyB];`,
		},
		{
			Code: `import { join as localJoin } from "path";
export const joined = localJoin("a", "b");`,
		},
		{
			Code: `interface Foo {
  bar: string;
}
export = Foo;`,
		},
		{
			Tsx: true,
			Code: `import React from "react";
export const node = <div />;`,
		},
		{
			Code: `export class Example {
  constructor(value: string, private name: string) {
    name = "";
    console.log(this.name);
  }
}`,
		},
	}

	invalid := []rule_tester.InvalidTestCase{
		{
			Code: `export {};
const value = 1;
function consume(value: number) { return value; }
consume(2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
			},
		},
		{
			Code: `export {};
const value = 1;
function consume(value: number) { return 0; }
console.log(value);
consume(2);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 3, Column: 18},
			},
		},
		{
			Code: `export class Example {
  constructor(value: string, private readonly name: string) {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 15},
			},
		},
		{
			Code: `export {};
const prop = 1;
const label = 2;
declare const obj: any;
label: for (;;) { obj.prop; break label; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
				{MessageId: "unusedVar", Line: 3, Column: 7},
			},
		},
		{
			Code: `namespace Empty {}
function consume(Empty: number) { return Empty; }
export { consume };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 11},
			},
		},
		{
			Code: `const value = 0;
export namespace Container {
  const value = 1;
  export { value };
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 7},
			},
		},
		{
			Code: `interface Foo {
  bar: string;
}
type Bar = 1;
export = Foo;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 4, Column: 6},
			},
		},
		{
			Tsx: true,
			Code: `export {};
const attr = 1;
declare const Component: any;
const rendered = <Component attr="value" />;
console.log(rendered);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 7},
			},
		},
		{
			Code: `export declare namespace A {
  const ambientMember: number;
}
export declare namespace B {
  const privateMember: number;
  export {};
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 5, Column: 9},
			},
		},
		{
			FileName: "reference-cache.d.ts",
			Code: `declare const publicValue: number;
declare const privateValue: number;
export { publicValue };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 2, Column: 15},
			},
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		valid,
		invalid,
	)
}

func TestNoUnusedVarsTypeOnlyLocalExports(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			{Code: `type T = {}; export type { T };`},
			{Code: `interface M {} const M = 1; export type { M };`},
			{Code: `import { V } from "./foo"; export type { V };`},
			{Code: `const value = 1; export namespace N { export { value }; } consume(N);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `const value = 1; export type { value };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
			},
			{
				Code:   "function f() {}\nexport { type f as F };",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
			},
			{
				Code: "type Token = {};\nnamespace Box {\n  const Token = 1;\n  export type { Token };\n}\nconsume(Box);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unusedVar", Line: 3, Column: 9},
				},
			},
			{
				Code:   "let assigned;\nexport type { assigned };\nassigned = 1;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Column: 1}},
			},
		},
	)
}

func TestNoUnusedVarsPatternCacheBounded(t *testing.T) {
	for i := range maxCachedPatterns + 32 {
		cachedPattern(fmt.Sprintf("^cache-attack-%d$", i))
	}

	patternCache.RLock()
	defer patternCache.RUnlock()
	if got := len(patternCache.entries); got > maxCachedPatterns {
		t.Fatalf("pattern cache has %d entries, want at most %d", got, maxCachedPatterns)
	}
	if len(patternCache.order) != len(patternCache.entries) {
		t.Fatalf("pattern cache order has %d entries, map has %d", len(patternCache.order), len(patternCache.entries))
	}
	seen := make(map[string]struct{}, len(patternCache.order))
	for _, pattern := range patternCache.order {
		if _, ok := patternCache.entries[pattern]; !ok {
			t.Fatalf("pattern cache order contains missing entry %q", pattern)
		}
		if _, duplicate := seen[pattern]; duplicate {
			t.Fatalf("pattern cache order contains duplicate entry %q", pattern)
		}
		seen[pattern] = struct{}{}
	}
}
