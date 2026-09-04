// TestNoThisOutsideOfClassExtras tests edge-shape augmentation (Dimension 1-4),
// real-user code shapes from upstream issues, and branch lock-ins.
package no_this_outside_of_class_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_this_outside_of_class"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisOutsideOfClassExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_this_outside_of_class.NoThisOutsideOfClassRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: Async, generator, and async generator class methods ----
			{
				Code: `
class Foo {
	async asyncMethod() {
		return this.value;
	}
	*genMethod() {
		yield this.value;
	}
	async *asyncGenMethod() {
		yield this.value;
	}
}
`,
			},
			// ---- Dimension 4: Parenthesized receiver and expression wrappers ----
			{
				Code: `
class Foo {
	method() {
		return (this).value;
	}
	nestedParens() {
		return ((this)).value;
	}
	wrappedField = ((this.defaultValue));
	parenArrowField = (() => (this.defaultValue));
}
`,
			},
			// ---- Dimension 4: TypeScript private identifiers and parameter properties ----
			{
				Code: `
class Foo {
	#privateField = this.value;
	#privateMethod() {
		return this.value;
	}
	get #privateGetter() {
		return this.value;
	}
	set #privateSetter(val) {
		this.value = val;
	}
	constructor(public x = this.value, readonly y = this.value) {}
}
`,
			},
			// ---- Dimension 4: Deeply nested arrow functions inside class method ----
			{
				Code: `
class Foo {
	deepArrows() {
		return () => () => () => this.value;
	}
}
`,
			},
			// ---- Dimension 4: Object literal properties inside class method ----
			{
				Code: `
class Foo {
	method() {
		const obj = {
			[this.key]: 1,
			val: this.value,
			arrow: () => this.value,
		};
		return obj;
	}
}
`,
			},
			// ---- Dimension 4: JSX inside class method ----
			{
				Tsx: true,
				Code: `
class Foo {
	render() {
		return <div title={this.title}>{this.content}</div>;
	}
}
`,
			},
			// ---- Dimension 4: JSX with <this /> tag name (ignored like ESTree) ----
			{
				Tsx:  true,
				Code: "<this />;",
			},
			// ---- Real-user: Issue #3593 - TypeScript this parameter in validator callback ----
			{
				Code: `
const config = {
	validate: {
		validator(this: TrackedModel, value: Date | null) {
			const requiresSpecificDate = this.notificationFrequency === 'once';
			const hasSpecificDate = value !== null;
			return requiresSpecificDate === hasSpecificDate;
		},
	},
};
`,
			},
			// ---- Real-user: Nested class inside function with this parameter ----
			{
				Code: `
function outer(this: OuterType) {
	class Inner {
		[this.key] = 1;
		[this.method]() {}
	}
	return Inner;
}
`,
			},
			// ---- Lock-in: Method overload signatures with this in implementation ----
			{
				Code: `
class Foo {
	method(x: number): void;
	method(x: string): void;
	method(x: any): void {
		this.value = x;
	}
}
`,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 1: Async and generator functions outside class ----
			invalidCase(`
async function foo() {
	return this.value;
}
`),
			invalidCase(`
function* foo() {
	yield this.value;
}
`),
			// ---- Dimension 4: Parenthesized this at top level ----
			invalidCase("(this).value;"),
			invalidCase("((this)).value;"),
			invalidCase("(((this)));"),
			// ---- Dimension 4: Deeply nested arrow functions at top level ----
			invalidCase("const f = () => () => () => this.value;"),
			// ---- Dimension 4: Object literal method inside class method (binds its own this) ----
			invalidCase(`
class Foo {
	method() {
		const obj = {
			inner() {
				return this.value;
			},
		};
	}
}
`),
			// ---- Dimension 4: JSX inside non-class function ----
			func() rule_tester.InvalidTestCase {
				tc := invalidCase(`
function Comp() {
	return <div title={this.title} />;
}
`)
				tc.Tsx = true
				return tc
			}(),
			// ---- Dimension 4: Type query (typeof this) at top level ----
			invalidCase("type T = typeof this;"),
			// ---- Dimension 4: Type query (typeof this) in class field type annotation ----
			invalidCase(`
class Foo {
	x: typeof this;
}
`),
			// ---- Real-user: Issue #2450 - UMD wrapper with self : this fallback ----
			invalidCase("(typeof self !== 'undefined' ? self : this);"),
			// ---- Real-user: Issue #2450 - Vue options API with this in method ----
			invalidCase(`
const component = {
	methods: {
		foo() {
			this.name = 'xxx';
		},
	},
};
`),
			// ---- Real-user: Issue #2450 - Vue options API with multiple methods using this ----
			invalidCase(`
const component = {
	methods: {
		greet() {
			return this.message;
		},
		update() {
			this.count++;
		},
	},
};
`, 1, 2),
			// ---- Real-user: Issue #3593 - Function inside validator with this parameter ----
			invalidCase(`
function validator(this: TrackedModel) {
	const helper = function () {
		return this.subValue;
	};
	return helper();
}
// Occurrence 1 is 'this: TrackedModel' (parameter name, ignored).
// Occurrence 2 is 'this.subValue' (expression inside helper, reported).
`, 2),
		},
	)
}
