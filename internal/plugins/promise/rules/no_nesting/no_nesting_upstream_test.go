// TestNoNestingUpstream migrates the full valid/invalid suite from upstream
// __tests__/no-nesting.js 1:1. Position assertions cover line/column for every
// invalid case. rslint-specific lock-in cases live in no_nesting_extras_test.go.

package no_nesting_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_nesting"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const avoidNestingMessage = "Avoid nesting promises."

func TestNoNestingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_nesting.NoNestingRule,
		[]rule_tester.ValidTestCase{
			// ---- resolve and reject are sometimes okay ----
			{Code: `Promise.resolve(4).then(function(x) { return x })`},
			{Code: `Promise.reject(4).then(function(x) { return x })`},
			{Code: `Promise.resolve(4).then(function() {})`},
			{Code: `Promise.reject(4).then(function() {})`},

			// ---- throw and return are fine ----
			{Code: `doThing().then(function() { return 4 })`},
			{Code: `doThing().then(function() { throw 4 })`},
			{Code: `doThing().then(null, function() { return 4 })`},
			{Code: `doThing().then(null, function() { throw 4 })`},
			{Code: `doThing().catch(null, function() { return 4 })`},
			{Code: `doThing().catch(null, function() { throw 4 })`},

			// ---- arrow functions and other things ----
			{Code: `doThing().then(() => 4)`},
			{Code: `doThing().then(() => { throw 4 })`},
			{Code: `doThing().then(()=>{}, () => 4)`},
			{Code: `doThing().then(()=>{}, () => { throw 4 })`},
			{Code: `doThing().catch(() => 4)`},
			{Code: `doThing().catch(() => { throw 4 })`},

			// ---- random functions and callback methods ----
			{Code: `var x = function() { return Promise.resolve(4) }`},
			{Code: `function y() { return Promise.resolve(4) }`},
			{Code: `function then() { return Promise.reject() }`},
			{Code: `doThing(function(x) { return Promise.reject(x) })`},

			// ---- Promise statics and Promise.all are fine inside callbacks ----
			{Code: `doThing().then(function() { return Promise.all([a,b,c]) })`},
			{Code: `doThing().then(function() { return Promise.resolve(4) })`},
			{Code: `doThing().then(() => Promise.resolve(4))`},
			{Code: `doThing().then(() => Promise.all([a]))`},

			// ---- references vars in closure ----
			{Code: `doThing()
      .then(a => getB(a)
        .then(b => getC(a, b))
      )`},
			{Code: `doThing()
      .then(a => {
        const c = a * 2;
        return getB(c).then(b => getC(c, b))
      })`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `doThing().then(function() { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 31}},
			},
			{
				Code:   `doThing().then(function() { b.catch() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 31}},
			},
			{
				Code:   `doThing().then(function() { return a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 38}},
			},
			{
				Code:   `doThing().then(function() { return b.catch() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 38}},
			},
			{
				Code:   `doThing().then(() => { a.then() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 26}},
			},
			{
				Code:   `doThing().then(() => { b.catch() })`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 26}},
			},
			{
				Code:   `doThing().then(() => a.then())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 24}},
			},
			{
				Code:   `doThing().then(() => b.catch())`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 1, Column: 24}},
			},
			// ---- references vars in closure ----
			{
				Code: `
      doThing()
        .then(a => getB(a)
          .then(b => getC(b))
        )`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 4}},
			},
			{
				Code: `
      doThing()
        .then(a => getB(a)
          .then(b => getC(a, b)
            .then(c => getD(a, c))
          )
        )`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNesting", Message: avoidNestingMessage, Line: 5}},
			},
		},
	)
}
