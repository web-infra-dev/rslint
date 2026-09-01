// TestNoThisAssignmentExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
package no_this_assignment_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_this_assignment"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisAssignmentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_this_assignment.NoThisAssignmentRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: Destructuring and rest targets are not identifiers ----
			{Code: `const {...rest} = this;`},
			{Code: `({...rest} = this);`},
			{Code: `[first, ...rest] = this;`},

			// ---- Dimension 4: TypeScript wrappers are not direct this expressions ----
			{Code: `const self = this!;`, FileName: "file.ts"},
			{Code: `const self = this as unknown;`, FileName: "file.ts"},
			{Code: `const self = this satisfies unknown;`, FileName: "file.ts"},

			// ---- Dimension 4: Declaration/container forms excluded upstream ----
			{Code: `class A { accessor self = this; }`, FileName: "file.ts"},
			{Code: `const object = {self: this};`},
			{Code: `function f(self = this) {}`},
			{Code: `for (const self of [this]) {}`},

			// ---- Dimension 4: Empty and body-absent forms degrade gracefully ----
			{Code: `const {} = this;`},
			{Code: `declare const self: unknown;`, FileName: "file.ts"},

			// Locks in upstream getProblem() arm 1: the target must be an identifier.
			{Code: `object.self = this;`},
			{Code: `this.self = this;`},

			// Locks in upstream getProblem() arm 2: the value must be exactly this.
			{Code: `const self = this.value;`},
			{Code: `self = getThis();`},

			// Locks in the AssignmentExpression listener gate.
			{Code: `self === this;`},
			{Code: `self + this;`},

			// ---- Review regression: destructuring defaults are ESTree AssignmentPattern nodes ----
			{Code: `[foo = this] = array;`},
			{Code: `({foo: bar = this} = object);`},
			{Code: `for ([foo = this] of array) {}`},
			{Code: `[[foo = this] = fallback] = array;`},
			{Code: `for ({key: self = this} in source) {}`},
			{Code: `for ([self = this] in source) {}`},
			{Code: `for ([[self = this] = fallback] in source) {}`},
		},
		[]rule_tester.InvalidTestCase{
			// Review regression: a TypeScript wrapper around an ordinary assignment
			// remains an AssignmentExpression, not an AssignmentPattern.
			invalid(`[(foo = this) as any] = arr`, `foo = this`, "foo"),

			// ---- Dimension 4: ESTree erases parenthesized expressions ----
			invalid(`const self = (this);`, `self = (this)`, "self"),
			invalid(`self = ((this));`, `self = ((this))`, "self"),

			// ---- Dimension 4: Variable declarations in loop headers ----
			invalid(`for (let self = this; condition; update) {}`, `self = this`, "self"),
			invalid(`for (self = this; condition; update) {}`, `self = this`, "self"),
			invalid(`using self = this;`, `self = this`, "self"),

			// Locks in upstream AssignmentExpression listener: compound/logical operators.
			invalid(`self += this;`, `self += this`, "self"),
			invalid(`self ||= this;`, `self ||= this`, "self"),
			invalid(`self ??= this;`, `self ??= this`, "self"),

			// ---- Review regression: ordinary expressions inside assignment targets are not patterns ----
			invalid(`({a: ({self: self = this}).x} = source);`, `self = this`, "self"),
			invalid(`({a: call({self: self = this})} = source);`, `self = this`, "self"),
			invalid(`for ({a: (condition ? {self: self = this} : fallback).x} in source) {}`, `self = this`, "self"),

			// ---- Real-user: issue #997 proposal shapes ----
			invalid(`const that = this; function foo() { return that.foo; }`, `that = this`, "that"),

			// ---- Real-user: issue #1108 nested callback aliases ----
			{
				Code: "class A {\n  test() {\n    const self = this;\n    let current;\n    return {mouseOver() { current = this; return self.input; }};\n  }\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError("class A {\n  test() {\n    const self = this;\n    let current;\n    return {mouseOver() { current = this; return self.input; }};\n  }\n}", `self = this`, "self"),
					expectedError("class A {\n  test() {\n    const self = this;\n    let current;\n    return {mouseOver() { current = this; return self.input; }};\n  }\n}", `current = this`, "current"),
				},
			},

			// Diagnostic contract: multi-line report range.
			invalid("const self =\n  this;", "self =\n  this", "self"),
			invalid("self =\n  this;", "self =\n  this", "self"),
		},
	)
}

// N/A: Access/key forms do not apply; the rule accepts only identifier targets.
// N/A: Optional chains do not produce direct identifier assignments to this.
// N/A: The rule has no autofix or suggestions.
