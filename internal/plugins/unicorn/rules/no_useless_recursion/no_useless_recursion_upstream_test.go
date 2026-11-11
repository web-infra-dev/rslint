// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-recursion.js
package no_useless_recursion_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_recursion"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessRecursionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_recursion.NoUselessRecursionRule,
		[]rule_tester.ValidTestCase{
			// Also a documentation example.
			{Code: `function foo(bar) {
	while (bar.baz) {
		bar = bar.baz;
	}

	return bar;
}`, FileName: "case.js"},
			{Code: `function foo(bar) {
	for (; bar.baz; bar = bar.baz) {}

	return bar;
}`, FileName: "case.js"},
			// Also a documentation example.
			{Code: `function foo(bar) {
	if (Array.isArray(bar)) {
		return bar.map(baz => foo(baz));
	}

	return bar;
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	return 1 + foo(value);
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	return value ? foo(value.next) : value;
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	function bar() {
		return foo(value);
	}

	return bar();
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	const foo = () => value;
	return foo();
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	return foo?.(value);
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	return object.foo(value);
}`, FileName: "case.js"},
			{Code: `const foo = value => {
	return foo(value);
};`, FileName: "case.js"},
			{Code: `const foo = function (value) {
	return foo(value);
};`, FileName: "case.js"},
			{Code: `function * foo(value) {
	return foo(value);
}`, FileName: "case.js"},
			{Code: `const object = {
	foo(value) {
		return foo(value);
	},
};`, FileName: "case.js"},
			{Code: `class Foo {
	foo(value) {
		return foo(value);
	}
}`, FileName: "case.js"},
			{Code: `const object = {
	get foo() {
		return this.foo;
	},
};`, FileName: "case.js"},
			{Code: `function foo(value) {
	try {
		return foo(value.next);
	} finally {
		cleanup(value);
	}
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	try {
		return foo(value.next);
	} catch (error) {
		if (value.retry) {
			throw error;
		}

		return value;
	}
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	using resource = acquire(value);
	return foo(value.next);
}`, FileName: "case.js"},
			{Code: `async function foo(value) {
	await using resource = acquire(value);
	return foo(value.next);
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	for (using resource = acquire(value); value.next;) {
		return foo(value.next);
	}
}`, FileName: "case.js"},
			{Code: `async function foo(value) {
	for (await using resource = acquire(value); value.next;) {
		return foo(value.next);
	}
}`, FileName: "case.js"},
			{Code: `function foo(value) {
	for (using resource of resources) {
		return foo(value.next);
	}
}`, FileName: "case.js"},
			{Code: `async function foo(value) {
	for await (await using resource of resources) {
		return foo(value.next);
	}
}`, FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			// Also a documentation example.
			recursiveReturn(`function foo(bar) {
	if (bar.baz) {
		return foo(bar.baz);
	}

	return bar;
}`, "case.js", 3, 10, 13),
			recursiveReturn(`const bar = function foo(value) {
	return foo(value.next);
};`, "case.js", 2, 9, 12),
			recursiveReturn(`function foo(first, second) {
	return foo(first.next, second.next);
}`, "case.js", 2, 9, 12),
			recursiveReturn(`function foo() {
	return foo();
}`, "case.js", 2, 9, 12),
			recursiveReturn(`async function foo(value) {
	if (value) {
		return foo(value.next);
	}

	return value;
}`, "case.js", 3, 10, 13),
			// TypeScript parser.
			recursiveReturn(`function foo(bar: Bar) {
	return foo(bar.baz as Bar);
}`, "case.ts", 2, 9, 12),
		},
	)
}

func recursiveReturn(code, filename string, line, column, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: filename,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "no-useless-recursion",
			Message:   "Use a loop instead of this recursive function call.",
			Line:      line, Column: column,
			EndLine: line, EndColumn: endColumn,
		}},
	}
}
