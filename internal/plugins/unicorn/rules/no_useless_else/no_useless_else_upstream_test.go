// Ported from eslint-plugin-unicorn v79.0.0 test/no-useless-else.js and documentation. See LICENSE.
package no_useless_else_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_else"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func elseAt(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-useless-else",
		Message:   "Unexpected `else` after a statement that exits.",
		Line:      line, Column: column, EndLine: line, EndColumn: column + 4,
	}
}

func TestNoUselessElseUpstreamSnapshots(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		return;
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			do {} while (maybeThrow(), process.exit(1));
		} catch {
			cleanup();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			throw error;
			process.exit(1);
		} catch {
			cleanup();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			process.exit(maybeThrow());
		} catch {
			cleanup();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			process.exit((maybeThrow(), 1));
		} catch {
			cleanup();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			maybeThrow();
			process.exit();
		} catch {
			cleanup();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `if (foo) {
	bar();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `const MODE_GROUP = new Set(['foo', 'bar']);

const args = new Map();

const modes = new Set();
if (args.has('modes')) {
	for (const mode of args.get('modes').split(',')) {
		const trimmed = mode.trim();
		if (MODE_GROUP.has(trimmed)) {
			modes.add(trimmed);
		}
	}

	if (!modes.size) {
		throw new Error('No valid modes');
	}
} else {
	for (const mode of MODE_GROUP) {
		modes.add(mode);
	}
}`},
		{FileName: "case.js", Code: `const modes = new Set(['foo']);
modes.clear();
if (condition) {
	modes.size ? process.exit() : doSomething();
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
Object.defineProperty(object, 'value', {get() { return false; }});

if (condition) {
	if (object.value) {
		throw new Error();
	}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
Object.defineProperty(object, 'value', {get() { throw new Error(); }});

if (condition) {
	try {
		if (object.value && true) {
			throw new Error();
		}
	} catch {}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
Object.defineProperty(object, 'value', {get() { throw new Error(); }});

if (condition) {
	try {
		if (object?.value) {
			throw new Error();
		}
	} catch {}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
const getterKey = 'value';
Object.defineProperty(object, 'value', {get() { throw new Error(); }});

if (condition) {
	try {
		if (object[getterKey]) {
			throw new Error();
		}
	} catch {}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
Object.defineProperty(object, 'value', {get() { throw new Error(); }});

if (condition) {
	try {
		object.value(process.exit(1));
	} catch {}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
Object.defineProperty(object, 'value', {get() { throw new Error(); }});

if (condition) {
	try {
		object.value.foo(process.exit(1));
	} catch {}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `const object = {value: true};
Object.defineProperty(object, 'value', {get() { throw new Error(); }});

if (condition) {
	try {
		(object?.value)[process.exit(1)];
	} catch {}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `let modes = new Set(['foo']);

if (condition) {
	if ((modes = new Set()).size) {
		throw new Error();
	}
} else {
	doSomethingElse();
}`},
		{FileName: "case.js", Code: `function qux() {
	const modes = new Set(['foo']);
	modes.clear();

	if (foo) {
		if (modes.size) {
			return;
		}
		bar();
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		if (bar) {
			return;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `const value = foo
	? bar()
	: baz();`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			return bar();
		} catch {}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `if (foo) {
	try {
		class Example {
			static {
				maybeThrow();
				process.exit(1);
			}
		}
	} catch {}
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	while (bar) {
		break;
	}
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	switch (bar) {
		case baz:
			break;
	}
} else {
	qux();
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		switch (bar) {
			case 1:
				return;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		switch (bar) {
			case 1:
				qux();
				break;
			default:
				return;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `for (const foo of bar)
	if (foo) {
		continue;
	} else {
		baz();
	}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		outer: for (const a of b) {
			for (const c of d) {
				break outer;
			}
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		switch (bar) {
			case 1:
				return;
			default:
				break;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		switch (value) {
			case 1:
				break;
		default:
			break;
		case process.exit(1):
			{}
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		switch (value) {
			case 1:
				return;
		default:
			break;
		case process.exit(1):
			{}
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			doSomething();
		} catch {
			handle();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		if (inner) {
			return;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		do {
			if (bar) {
				break;
			}
		} while (true);
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		for (const x of xs) {
			return x;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		try {
			return doSomething();
		} catch {
			handle();
		} finally {
			cleanup();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		const inner = () => {
			return 1;
		};
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		function inner() {
			return 1;
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	return;
	if (foo) {
		return;
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	return;
	if (foo) {
		process.exit();
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		false && process.exit();
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux(process) {
	if (foo) {
		process.exit();
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `import process from 'node:process';

if (foo) {
	process.exit();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	process.exitCode = 1;
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	process?.exit();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	process.exit?.();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	callback?.(process.exit(1));
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	foo?.bar[process.exit(1)];
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	process['exit']();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	lib.process.exit();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `if (foo) {
	new process.exit();
} else {
	baz();
}`},
		{FileName: "case.js", Code: `function qux() {
	if (foo) {
		if (bar) {
			process.exit();
		}
	} else {
		baz();
	}
}`},
		{FileName: "case.js", Code: `function qux(value) {
	if (foo) {
		try {
			switch (value) {
			default:
				break;
			case process.exit(1):
				{}
			}
		} catch {
			cleanup();
		}
	} else {
		baz();
	}
}`},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		outer: do {
			while (true) {
				continue outer;
			}
		} while (process.exit(1));
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		outer: do {
			while (true) {
				continue outer;
			}
		} while (process.exit(1));
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux(condition) {
	if (foo) {
		try {
			condition ? process.exit(1) : process.exit(2);
		} catch {}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(6, 4)},
			Output: []string{
				`function qux(condition) {
	if (foo) {
		try {
			condition ? process.exit(1) : process.exit(2);
		} catch {}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux(condition) {
	if (foo) {
		try {
			if (condition) {
				process.exit(1);
			} else {
				process.exit(2);
			}
		} catch {}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(6, 6), elseAt(10, 4)},
			Output: []string{
				`function qux(condition) {
	if (foo) {
		try {
			if (condition) {
				process.exit(1);
			} else {
				process.exit(2);
			}
		} catch {}
	}
	baz();
}`,
				`function qux(condition) {
	if (foo) {
		try {
			if (condition) {
				process.exit(1);
			}
			process.exit(2);
		} catch {}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		try {
			return bar();
		} finally {
			cleanup();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		try {
			return bar();
		} finally {
			cleanup();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		switch (bar) {
			case 1:
				return;
			default:
				throw new Error();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(9, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		switch (bar) {
			case 1:
				return;
			default:
				throw new Error();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		while (true) {
			doSomething();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(6, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		while (true) {
			doSomething();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		for (;;) {
			doSomething();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(6, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		for (;;) {
			doSomething();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		do {
			doSomething();
		} while (true);
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(6, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		do {
			doSomething();
		} while (true);
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		switch (bar) {
			case 1:
			case 2:
				return;
			default:
				throw new Error();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(10, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		switch (bar) {
			case 1:
			case 2:
				return;
			default:
				throw new Error();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		try {
			return doSomething();
		} catch {
			throw new Error();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		try {
			return doSomething();
		} catch {
			throw new Error();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		if (a) {
			return;
		} else if (b) {
			throw new Error();
		} else {
			return;
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 5), elseAt(10, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		if (a) {
			return;
		} else if (b) {
			throw new Error();
		} else {
			return;
		}
	}
	baz();
}`,
				`function qux() {
	if (foo) {
		if (a) {
			return;
		}
		if (b) {
			throw new Error();
		} else {
			return;
		}
	}
	baz();
}`,
				`function qux() {
	if (foo) {
		if (a) {
			return;
		}
		if (b) {
			throw new Error();
		}
		return;
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	outer: for (const a of b) {
		if (foo) {
			continue outer;
		} else {
			baz();
		}
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 5)},
			Output: []string{
				`function qux() {
	outer: for (const a of b) {
		if (foo) {
			continue outer;
		}
		baz();
	}
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		try {
			doSomething();
		} finally {
			return cleanup();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		try {
			doSomething();
		} finally {
			return cleanup();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else { bar(); }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `if (foo) {
	throw new Error();
} else {
	bar();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(3, 3)},
			Output: []string{
				`if (foo) {
	throw new Error();
}
bar();`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (false) {
		return;
	} else {
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (false) {
		return;
	}
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (true) {
		return;
	} else {
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (true) {
		return;
	}
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `while (foo) {
	if (bar) {
		break;
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`while (foo) {
	if (bar) {
		break;
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `while (foo) {
	if (bar) {
		continue;
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`while (foo) {
	if (bar) {
		continue;
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `switch (foo) {
	case bar:
		if (baz) {
			break;
		} else {
			qux();
		}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 5)},
			Output: []string{
				`switch (foo) {
	case bar:
		if (baz) {
			break;
		}
		qux();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `class Foo {
	static {
		if (foo) {
			throw new Error();
		} else {
			bar();
		}
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 5)},
			Output: []string{
				`class Foo {
	static {
		if (foo) {
			throw new Error();
		}
		bar();
	}
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo)
		return;
	else
		bar();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 2)},
			Output: []string{
				`function qux() {
	if (foo)
		return;
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else if (bar) {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	if (bar) {
		baz();
	}
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		if (bar) {
			return;
		} else {
			throw new Error();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 5), elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		if (bar) {
			return;
		} else {
			throw new Error();
		}
	}
	baz();
}`,
				`function qux() {
	if (foo) {
		if (bar) {
			return;
		}
		throw new Error();
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		// Keep this comment.
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	// Keep this comment.
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		/*
		Keep this comment.
		*/
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	/*
	Keep this comment.
	*/
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code:     "function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar(`foo`);\n\t}\n}",
			Errors:   []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				"function qux() {\n\tif (foo) {\n\t\treturn;\n\t}\n\tbar(`foo`);\n}",
			},
		},
		{
			FileName: "case.js",
			Code:     "function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar(`\n\tleft\n`);\n\t}\n}",
			Errors:   []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		bar('foo\
bar');
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.jsx",
			Code:     "function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\treturn (\n\t\t\t<div>{`\n\t\t\t\tmultiline\n\t\t\t\ttemplate\n\t\t\t`}</div>\n\t\t);\n\t}\n}",
			Errors:   []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.jsx",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		return <pre>
			line
		</pre>;
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	return <pre>
		line
	</pre>;
}`,
			},
		},
		{
			FileName: "case.jsx",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		<Widget title="first
			second" />;
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.jsx",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		render(<div>
			text
		</div>);
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		return;
	}
	render(<div>
		text
	</div>);
}`,
			},
		},
		{
			FileName: "case.jsx",
			Code: `function init() {
	if (tagName) {
		addTagToFooter(tagName);
		return;
	} else {
		void addReleaseBanner(
			<>
				No <ExplanationLink>stable version tags</ExplanationLink> for this PR.
			</>,
			signal,
		);
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 4)},
			Output: []string{
				`function init() {
	if (tagName) {
		addTagToFooter(tagName);
		return;
	}
	void addReleaseBanner(
		<>
			No <ExplanationLink>stable version tags</ExplanationLink> for this PR.
		</>,
		signal,
	);
}`,
			},
		},
		{
			FileName: "case.jsx",
			Code:     "function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\trender(<div>\n\t\t\ttext\n\t\t</div>, css`\n\t\t\tcolor: red;\n\t\t`);\n\t}\n}",
			Errors:   []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		const bar = 1;
		baz(bar);
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		let bar = 1;
		baz(bar);
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		class Bar {}
		baz(Bar);
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		function bar() {}
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else /* comment */ {
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} /* comment */ else {
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 18)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		bar();
	} // trailing
	baz();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux(foo) {
	if (foo) {
		return;
	} else function bar() {}
}`,
			LanguageOptions: rule.LanguageOptions{SourceType: "script", ECMAVersion: 5},
			Errors:          []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		bar()
	} baz()
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		bar = function() {}
	}
	(baz)();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.jsx",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		<Foo>
			text
		</Foo>
	}
	<Bar />
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo)
		return
	else
		[bar].forEach(baz)
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 2)},
		},
		{
			FileName: "case.ts",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		type Bar = string;
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.ts",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		interface Bar {}
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.ts",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		enum Bar {}
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.ts",
			Code: `function qux() {
	if (foo) {
		return;
	} else {
		namespace Bar {}
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		process.exit();
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		process.exit();
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		true && process.exit();
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		true && process.exit();
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		[...process.exit()];
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		[...process.exit()];
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo)
		process.exit(1);
	else
		baz();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 2)},
			Output: []string{
				`function qux() {
	if (foo)
		process.exit(1);
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		setup();
		process.exit(1);
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		setup();
		process.exit(1);
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		{
			process.exit();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(6, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		{
			process.exit();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		switch (value) {
			case 1:
				if (false) {
					break;
				}

				process.exit(1);
			default:
				process.exit(2);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(13, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		switch (value) {
			case 1:
				if (false) {
					break;
				}

				process.exit(1);
			default:
				process.exit(2);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		switch (value) {
			case 1:
				if (true) {
					return;
				}
				break;
			default:
				process.exit(1);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(12, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		switch (value) {
			case 1:
				if (true) {
					return;
				}
				break;
			default:
				process.exit(1);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		label: {
			if (false) {
				break label;
			}

			process.exit(1);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(10, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		label: {
			if (false) {
				break label;
			}

			process.exit(1);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		process.exit(1);
	} else if (bar) {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		process.exit(1);
	}
	if (bar) {
		baz();
	}
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		try {
			process.exit(1);
			cleanup();
		} finally {
			cleanup();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(9, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		try {
			process.exit(1);
			cleanup();
		} finally {
			cleanup();
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		if (bar) {
			return;
		} else {
			process.exit(1);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(5, 5), elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		if (bar) {
			return;
		} else {
			process.exit(1);
		}
	}
	baz();
}`,
				`function qux() {
	if (foo) {
		if (bar) {
			return;
		}
		process.exit(1);
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		try {
			throw error;
		} catch {
			process.exit(1);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		try {
			throw error;
		} catch {
			process.exit(1);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		switch (value) {
			case 1:
				local: {
					break local;
				}

				process.exit(1);
			default:
				process.exit(2);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(13, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		switch (value) {
			case 1:
				local: {
					break local;
				}

				process.exit(1);
			default:
				process.exit(2);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `outer: while (condition) {
	if (foo) {
		switch (value) {
			case 1:
				break outer;
		default:
				process.exit(1);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(9, 4)},
			Output: []string{
				`outer: while (condition) {
	if (foo) {
		switch (value) {
			case 1:
				break outer;
		default:
				process.exit(1);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `outer: while (condition) {
	if (foo) {
		switch (value) {
			case 1:
				continue outer;
		default:
				process.exit(1);
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(9, 4)},
			Output: []string{
				`outer: while (condition) {
	if (foo) {
		switch (value) {
			case 1:
				continue outer;
		default:
				process.exit(1);
		}
	}
	baz();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `function qux() {
	if (foo) {
		try {
			process.exit();
		} finally {
			cleanup();
		}
	} else {
		baz();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(8, 4)},
			Output: []string{
				`function qux() {
	if (foo) {
		try {
			process.exit();
		} finally {
			cleanup();
		}
	}
	baz();
}`,
			},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_else.NoUselessElseRule, valid, invalid)
}

func TestNoUselessElseUpstreamCRLF(t *testing.T) {
	invalid := []rule_tester.InvalidTestCase{
		{
			FileName: "case.js",
			Code:     "function run(item) {\r\n\tif (item.isActive) {\r\n\t\treturn process(item);\r\n\t} else {\r\n\t\tsave(item);\r\n\t\tlog(item);\r\n\t}\r\n}\r\n",
			Errors:   []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				"function run(item) {\r\n\tif (item.isActive) {\r\n\t\treturn process(item);\r\n\t}\r\n\tsave(item);\r\n\tlog(item);\r\n}\r\n",
			},
		},
		{
			FileName: "case.js",
			Code:     "function run(item) {\r\n\tif (item.isActive) {\r\n\t\treturn process(item);\r\n\t} else save(item);\r\n}\r\n",
			Errors:   []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				"function run(item) {\r\n\tif (item.isActive) {\r\n\t\treturn process(item);\r\n\t}\r\n\tsave(item);\r\n}\r\n",
			},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_else.NoUselessElseRule, nil, invalid)
}

func TestNoUselessElseUpstreamIndentation(t *testing.T) {
	invalid := []rule_tester.InvalidTestCase{
		{
			FileName: "case.js",
			Code: `function run(item) {
  if (item.isActive) {
    return process(item);
  } else {
    if (item.isReady) {
      save(item);
    }

    log(item);
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 5)},
			Output: []string{
				`function run(item) {
  if (item.isActive) {
    return process(item);
  }
  if (item.isReady) {
    save(item);
  }

  log(item);
}`,
			},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_else.NoUselessElseRule, nil, invalid)
}

func TestNoUselessElseUpstreamEmptyBlocks(t *testing.T) {
	invalid := []rule_tester.InvalidTestCase{
		{
			FileName: "case.js",
			Code: `function run(item) {
	if (item.isActive) {
		return process(item);
	} else {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function run(item) {
	if (item.isActive) {
		return process(item);
	}
}
`,
			},
		},
		{
			FileName: "case.js",
			Code: `function run(item) {
	if (item.isActive) {
		return process(item);
	} else {
	}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function run(item) {
	if (item.isActive) {
		return process(item);
	}
}
`,
			},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_else.NoUselessElseRule, nil, invalid)
}

func TestNoUselessElseUpstreamDocumentation(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{FileName: "case.js", Code: `function example() {
	if (foo) {
		return;
	}

	bar();
}`},
		{FileName: "case.js", Code: `if (foo) {
	throw new Error();
}

if (bar) {
	baz();
}`},
	}
	invalid := []rule_tester.InvalidTestCase{
		{
			FileName: "case.js",
			Code: `function example() {
	if (foo) {
		return;
	} else {
		bar();
	}
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(4, 4)},
			Output: []string{
				`function example() {
	if (foo) {
		return;
	}
	bar();
}`,
			},
		},
		{
			FileName: "case.js",
			Code: `if (foo) {
	throw new Error();
} else if (bar) {
	baz();
}`,
			Errors: []rule_tester.InvalidTestCaseError{elseAt(3, 3)},
			Output: []string{
				`if (foo) {
	throw new Error();
}
if (bar) {
	baz();
}`,
			},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_else.NoUselessElseRule, valid, invalid)
}
