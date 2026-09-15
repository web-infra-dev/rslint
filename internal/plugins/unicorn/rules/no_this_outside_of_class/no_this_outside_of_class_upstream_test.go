// TestNoThisOutsideOfClassUpstream migrates the complete valid/invalid suite from
// upstream test/no-this-outside-of-class.js at v72.0.0 (plus PR #3596 TypeScript
// this-parameter tests). Position assertions cover every invalid case.
// rslint-specific edge shapes, real-user cases, and branch lock-ins live in
// no_this_outside_of_class_extras_test.go.
package no_this_outside_of_class_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_this_outside_of_class"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID   = "no-this-outside-of-class"
	messageText = "Do not use `this` outside of classes."
)

func thisOccurrenceRange(code string, occurrence int) (line, column, endLine, endColumn int) {
	searchFrom := 0
	offset := -1
	for index := 1; index <= occurrence; index++ {
		found := strings.Index(code[searchFrom:], "this")
		if found < 0 {
			panic("this occurrence not found in: " + code)
		}
		offset = searchFrom + found
		searchFrom = offset + len("this")
	}

	before := code[:offset]
	line = strings.Count(before, "\n") + 1
	lastNewline := strings.LastIndex(before, "\n")
	column = offset - lastNewline
	return line, column, line, column + len("this")
}

func invalidCase(code string, occurrences ...int) rule_tester.InvalidTestCase {
	if len(occurrences) == 0 {
		occurrences = []int{1}
	}
	errors := make([]rule_tester.InvalidTestCaseError, len(occurrences))
	for i, occ := range occurrences {
		line, col, endLine, endCol := thisOccurrenceRange(code, occ)
		errors[i] = rule_tester.InvalidTestCaseError{
			MessageId: messageID,
			Message:   messageText,
			Line:      line,
			Column:    col,
			EndLine:   endLine,
			EndColumn: endCol,
		}
	}
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: errors,
	}
}

func TestNoThisOutsideOfClassUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_this_outside_of_class.NoThisOutsideOfClassRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
class Foo {
	constructor() {
		this.value = 1;
	}

	method() {
		return this.value;
	}

	methodWithDefault(value = this.value) {
		return value;
	}

	static method() {
		return this.value;
	}

	get value() {
		return this._value;
	}

	set value(value) {
		this._value = value;
	}

	#method() {
		return this.value;
	}
}
`,
			},
			{
				Code: `
const Foo = class {
	value = this.defaultValue;
	static value = this.defaultValue;
};
`,
			},
			{
				Code: `
class Foo {
	static {
		this.value = 1;
	}
}
`,
			},
			{
				Code: `
class Foo {
	method() {
		const getValue = () => this.value;
		return getValue();
	}
}
`,
			},
			{
				Code: `
class Foo {
	value = () => this.defaultValue;

	static {
		const update = () => this.value = 1;
		update();
	}
}
`,
			},
			{
				Code: `
class Foo {
	method() {
		class Bar {
			method() {
				return this.value;
			}
		}

		return Bar;
	}
}
`,
			},
			{
				Code: `
class Foo {
	method() {
		class Bar extends this.Base {
			[this.key]() {}
		}

		return Bar;
	}
}
`,
			},
			{
				Code: `
class Foo {
	method() {
		class Bar {
			[this.key] = 1;
		}

		return Bar;
	}
}
`,
			},
			// TypeScript: accessor property
			{
				Code: `
class Foo {
	accessor value = this.defaultValue;
}
`,
			},
			// TypeScript: non-arrow function with explicit this parameter
			{
				Code: `
const foo = {
	validator(this: TrackedModel, value: Date | null) {
		const getValue = () => this.value;
		return getValue() === value;
	},
};
`,
			},
			{
				Code: `
function validator(this: TrackedModel) {
	return this.value;
}
`,
			},
			{
				Code: `
const validator = function (this: TrackedModel) {
	return this.value;
};
`,
			},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase("this.value;"),
			invalidCase("const getValue = () => this.value;"),
			invalidCase(`
function foo() {
	return this.value;
}
`),
			invalidCase(`
function Foo(value) {
	this.value = value;
}
`),
			invalidCase(`
const foo = function () {
	return this.value;
};
`),
			invalidCase(`
(function () {
	this.value = 1;
})();
`),
			invalidCase(`
function foo() {
	const getValue = () => this.value;
	return getValue();
}
`),
			invalidCase(`
class Foo {
	method() {
		function getValue() {
			return this.value;
		}

		return getValue();
	}
}
`),
			invalidCase(`
class Foo {
	value = function () {
		return this.value;
	};
}
`),
			invalidCase(`
class Foo {
	static {
		function update() {
			this.value = 1;
		}

		update();
	}
}
`),
			invalidCase(`
const foo = {
	method() {
		return this.value;
	}
};
`),
			invalidCase(`
const foo = {
	get value() {
		return this._value;
	},
	set value(value) {
		this._value = value;
	}
};
`, 1, 2),
			invalidCase(`
const foo = {
	method: function () {
		return this.value;
	}
};
`),
			invalidCase(`
Foo.prototype.method = function () {
	this.value();
};
`),
			invalidCase(`
new SDK({
	onReady: function () {
		this.value;
	}
});
`),
			invalidCase(`
export default {
	methods: {
		refresh() {
			this.list = getList();
		}
	}
};
`),
			invalidCase(`
class Foo {
	[this.key]() {}
}
`),
			invalidCase(`
class Foo {
	[this.key] = 1;
}
`),
			invalidCase("class Foo extends this.Base {}"),
			// TypeScript: accessor property with computed key
			invalidCase(`
class Foo {
	accessor [this.key] = 1;
}
`),
			// TypeScript: function without this parameter inside function with this parameter
			invalidCase(`
const foo = {
	validator(this: TrackedModel) {
		function getValue() {
			return this.value;
		}

		return getValue();
	},
};
`, 2),
			// TypeScript: method without this parameter
			invalidCase(`
const foo = {
	method() {
		return this.value;
	},
};
`),
			// TypeScript: arrow function with (pseudo-)this parameter
			invalidCase("const validator = (this: TrackedModel) => this.value;", 2),
			// TypeScript: function with non-this parameter
			invalidCase(`
function validator(value: TrackedModel) {
	return this.value;
}
`),
		},
	)
}
