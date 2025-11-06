package array_callback_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrayCallbackReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ArrayCallbackReturnRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// forEach without checkForEach option
			{Code: `foo.forEach(function(x) { return x; });`},
			{Code: `foo.forEach(function() { if (a === b) { return; } });`},
			{Code: `foo.forEach(function() { return; });`},
			{Code: `foo.forEach(x => x);`},

			// Proper return statements
			{Code: `foo.map(function() { return true; });`},
			{Code: `foo.map(function() { return; });`, Options: []interface{}{map[string]interface{}{"allowImplicit": true}}},
			{Code: `foo.filter(function bar() { return true; });`},
			{Code: `foo.filter(function(x) { if (x) { return true; } else { return false; } });`},
			{Code: `foo.find(function() { return true; });`},
			{Code: `foo.reduce(function(acc, val) { return acc + val; }, 0);`},
			{Code: `foo.some(function() { return true; });`},
			{Code: `foo.every(function() { return true; });`},
			{Code: `foo.sort(function(a, b) { return a - b; });`},
			{Code: `foo.toSorted(function(a, b) { return a - b; });`},
			{Code: `foo.flatMap(function(x) { return [x]; });`},

			// Arrow functions with implicit returns
			{Code: `foo.map(x => x);`},
			{Code: `foo.map(x => { return x; });`},
			{Code: `foo.filter(x => x);`},
			{Code: `foo.every(x => x);`},
			{Code: `foo.some(x => x);`},

			// Array.from with proper callback
			{Code: `Array.from(x, function() { return true; });`},
			{Code: `Array.from(x, x => x);`},

			// Methods on non-arrays (we still check by method name)
			{Code: `foo.unknownMethod(function() {});`},
			{Code: `foo.unknownMethod(function() { return true; });`},

			// Nested functions
			{Code: `foo.map(function() { return function() { return true; }; });`},
			{Code: `foo.map(function() { return function() {}; });`},

			// allowImplicit option
			{
				Code: `foo.map(function() { return; });`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `foo.map(function() { if (a === b) { return; } return; });`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `foo.every(function() { return; });`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},
			{
				Code: `foo.filter(function() { return; });`,
				Options: []interface{}{map[string]interface{}{"allowImplicit": true}},
			},

			// checkForEach option - valid cases
			{
				Code: `foo.forEach(function(x) { bar(x); });`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
			},
			{
				Code: `foo.forEach(function(x) {});`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
			},
			{
				Code: `foo.forEach(function() { return; });`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
			},

			// checkForEach with allowVoid option
			{
				Code: `foo.forEach(x => void bar(x));`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true, "allowVoid": true}},
			},
			{
				Code: `foo.forEach((x, i) => void (i % 2 && bar(x)));`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true, "allowVoid": true}},
			},

			// Edge cases
			{Code: `foo.map(function() { try { return true; } catch(e) {} });`},
			{Code: `foo.map(async function() { return true; });`},
			{Code: `foo.map(async () => true);`},
			{Code: `foo.map(function*() { yield true; });`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Missing returns in various methods
			{
				Code: `Array.from(x, function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 15},
				},
			},
			{
				Code: `foo.every(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 11},
				},
			},
			{
				Code: `foo.filter(function foo() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 12},
				},
			},
			{
				Code: `foo.find(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 10},
				},
			},
			{
				Code: `foo.map(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 9},
				},
			},
			{
				Code: `foo.reduce(function(acc) {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 12},
				},
			},
			{
				Code: `foo.reduceRight(function(acc) {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 17},
				},
			},
			{
				Code: `foo.some(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 10},
				},
			},
			{
				Code: `foo.sort(function(a, b) {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 10},
				},
			},
			{
				Code: `foo.toSorted(function(a, b) {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 14},
				},
			},
			{
				Code: `foo.flatMap(function(x) {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 13},
				},
			},
			{
				Code: `foo.findLast(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 14},
				},
			},
			{
				Code: `foo.findLastIndex(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 19},
				},
			},
			{
				Code: `foo.findIndex(function() {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 15},
				},
			},

			// Arrow functions with missing returns
			{
				Code: `foo.map(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedReturnValue", Line: 1, Column: 9},
				},
			},
			{
				Code: `foo.filter(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedReturnValue", Line: 1, Column: 12},
				},
			},
			{
				Code: `Array.from(x, () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedReturnValue", Line: 1, Column: 15},
				},
			},

			// Incomplete returns (not all paths)
			{
				Code: `foo.map(function() { if (a === b) { return true; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 9},
				},
			},
			{
				Code: `foo.map(() => { if (a === b) { return true; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedInside", Line: 1, Column: 9},
				},
			},
			{
				Code: `foo.filter(function bar() { if (a === b) { return true; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 12},
				},
			},
			{
				Code: `foo.filter(function() { if (a) { return true; } else if (b) { return false; } });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 12},
				},
			},

			// Empty return statements without allowImplicit
			{
				Code: `foo.map(function() { return; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 9},
				},
			},
			{
				Code: `foo.every(function() { return; });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 11},
				},
			},

			// checkForEach option - invalid cases
			{
				Code: `foo.forEach(x => x);`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNoReturnValue", Line: 1, Column: 13},
				},
			},
			{
				Code: `foo.forEach(function(x) { return x; });`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNoReturnValue", Line: 1, Column: 13},
				},
			},
			{
				Code: `foo.forEach(function() { if (a === b) { return a; } });`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNoReturnValue", Line: 1, Column: 13},
				},
			},
			{
				Code: `foo.forEach(() => bar());`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNoReturnValue", Line: 1, Column: 13},
				},
			},

			// checkForEach with allowVoid - void is not used
			{
				Code: `foo.forEach(x => bar(x));`,
				Options: []interface{}{map[string]interface{}{"checkForEach": true, "allowVoid": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNoReturnValue", Line: 1, Column: 13},
				},
			},

			// Multiple methods chained
			{
				Code: `foo.map(x => x).filter(() => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedReturnValue", Line: 1, Column: 24},
				},
			},

			// Nested callbacks
			{
				Code: `foo.map(function() { return bar.filter(function() {}); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedAtEnd", Line: 1, Column: 40},
				},
			},
		},
	)
}
