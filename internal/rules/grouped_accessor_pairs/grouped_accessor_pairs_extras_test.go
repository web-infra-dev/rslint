package grouped_accessor_pairs

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestGroupedAccessorPairsExtras covers tsgo edge shapes, real-user examples,
// and every reachable upstream branch. The 1:1 ESLint migration lives in
// grouped_accessor_pairs_upstream_test.go.
func TestGroupedAccessorPairsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&GroupedAccessorPairsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver/expression wrappers ----
			// Parentheses around the complete dynamic key are transparent.
			{Code: `({ get [(key)](){}, set [key](value){} })`},
			// Parentheses nested within a dynamic key remain part of its token stream.
			{Code: `({ get [(key) + other](){}, middle: true, set [key + other](value){} })`},
			// TS assertions remain part of a dynamic key's token identity.
			{Code: `({ get [(key as string)](){}, middle: true, set [key](value){} })`},
			{Code: `({ get [key!](){}, middle: true, set [key](value){} })`},
			{Code: `({ get [key satisfies string](){}, middle: true, set [key](value){} })`},
			{Code: `({ get [source?.key](){}, set [source?.key](value){} })`},

			// ---- Dimension 4: access/key forms ----
			{Code: `({ get name(){}, set ['name'](value){} })`},
			{Code: `({ get 0(){}, set [0](value){} })`},
			{Code: `class C { get #name(){} middle(){} set '#name'(value){} }`},
			{Code: `({ get [left](){}, middle: true, set [right](value){} })`},
			{Code: `({ get [left + right](){}, set [left+right](value){} })`},
			{Code: `({ get [left + right](){}, middle: true, set [left - right](value){} })`},
			{Code: `({ get [+key](){}, middle: true, set [-key](value){} })`},
			{Code: `({ get [key++](){}, middle: true, set [key--](value){} })`},
			// N/A: element access is only an expression inside a computed key;
			// this rule does not inspect member-access receivers.
			{Code: `({ get [source['name']](){}, set [source['name']](value){} })`},

			// ---- Dimension 4: declaration/container forms ----
			{Code: `class C { get name(){} set name(value){} }`},
			{Code: `const C = class { get name(){} set name(value){} };`},
			{Code: `const value = { get name(){}, set name(value){} };`},
			{Code: `interface I { get name(): string; set name(value: string); }`, Options: []any{"anyOrder", map[string]any{"enforceForTSTypes": true}}},
			{Code: `type T = { get name(): string; set name(value: string); };`, Options: []any{"anyOrder", map[string]any{"enforceForTSTypes": true}}},
			// N/A: async and generator accessors are invalid JavaScript syntax.

			// ---- Dimension 4: nesting/traversal boundaries ----
			{Code: `class Outer { get name(){} set name(value){} method(){ return class Inner { get name(){} set name(value){} }; } }`},
			{Code: `const outer = { get name(){ return { get name(){}, set name(value){} }; }, set name(value){} };`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `({ ...source })`},
			{Code: `({})`},
			{Code: `class Empty {}`},
			{Code: `abstract class C { abstract get name(): string; abstract set name(value: string); }`},
			{Code: `declare class C { get name(): string; set name(value: string); }`},
			// N/A: rest elements in binding patterns are not object-literal members.

			// Locks in upstream areEqualKeys() cross-kind arm: a public string key
			// never matches a private key with the same visible spelling.
			{Code: `class C { get '#name'(){} middle(){} set #name(value){} }`},
			// Locks in upstream areEqualTokenLists() length/type/value mismatch arms.
			{Code: `({ get [a](){}, middle: true, set [a.b](value){} })`},
			{Code: `({ get [a + b](){}, middle: true, set [a - b](value){} })`},
			// Locks in upstream duplicate guard: duplicate getters suppress the pair.
			{Code: `({ get a(){}, middle: true, set a(value){}, get a(){} })`},
			// Locks in upstream class predicates: static and instance groups differ.
			{Code: `class C { get a(){} middle(){} static set a(value){} }`},
			// Locks in the default order fallback and explicit false TS option.
			{Code: `({ set a(value){}, get a(){} })`},
			{Code: `interface I { get a(): string; middle: true; set a(value: string); }`, Options: []any{"anyOrder", map[string]any{"enforceForTSTypes": false}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: ESLint issue #12277 object spread ----
			{
				Code: `({ get a(){}, ...source, set a(value){} })`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Message:   "Accessor pair getter 'a' and setter 'a' should be grouped.",
				}},
			},
			// ---- Real-user: ESLint issue #12277 intervening class method ----
			{
				Code: `class C { get a(){} method(){} set a(value){} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Message:   "Accessor pair getter 'a' and setter 'a' should be grouped.",
				}},
			},
			// ---- Real-user: ESLint issue #19860 interface accessors ----
			{
				Code:    `interface I { get prop(): string; between: true; set prop(value: string); }`,
				Options: []any{"anyOrder", map[string]any{"enforceForTSTypes": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Message:   "Accessor pair getter 'prop' and setter 'prop' should be grouped.",
				}},
			},
			// ---- Real-user: ESLint issue #19860 type-literal accessors ----
			{
				Code:    `type T = { get prop(): string; between: true; set prop(value: string); };`,
				Options: []any{"anyOrder", map[string]any{"enforceForTSTypes": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Message:   "Accessor pair getter 'prop' and setter 'prop' should be grouped.",
				}},
			},
			// Locks in upstream grouping-before-order branch.
			{
				Code:    `({ set a(value){}, middle: true, get a(){} })`,
				Options: []any{"getBeforeSet"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notGrouped"}},
			},
			// Locks in upstream getBeforeSet order arm.
			{
				Code:    `({ set a(value){}, get a(){} })`,
				Options: []any{"getBeforeSet"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "invalidOrder"}},
			},
			// Locks in upstream setBeforeGet order arm.
			{
				Code:    `class C { static get #a(){} static set #a(value){} }`,
				Options: []any{"setBeforeGet"},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "invalidOrder",
					Message:   "Expected static private setter #a to be before static private getter #a.",
				}},
			},
			// Locks in upstream static predicate with an intervening instance member.
			{
				Code:   `class C { static get a(){} method(){} static set a(value){} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "notGrouped"}},
			},
			// Locks in the class predicate's bodyless declaration arm: declare-class
			// accessors are MethodDefinition nodes upstream and remain enforceable.
			{
				Code: `declare class C { get a(): string; middle(): void; set a(value: string); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Message:   "Accessor pair getter 'a' and setter 'a' should be grouped.",
				}},
			},
			// Decorators are part of the MethodDefinition range used by upstream's
			// function-head location helper.
			{
				Code: `class C { @first() get a(){} field; @second(1) set a(value){} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Column:    37,
					EndColumn: 53,
				}},
			},
			// getFunctionNameWithKind stringifies a dynamic TSMethodSignature key
			// as the upstream null static-name result.
			{
				Code:    `interface I { get [key](): string; middle: true; set [key](value: string); }`,
				Options: []any{"anyOrder", map[string]any{"enforceForTSTypes": true}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "notGrouped",
					Message:   "Accessor pair getter 'null' and setter 'null' should be grouped.",
				}},
			},
		},
	)
}
