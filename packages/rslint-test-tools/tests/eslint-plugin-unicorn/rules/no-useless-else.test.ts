// Ported from eslint-plugin-unicorn v79.0.0 test/no-useless-else.js and documentation. See LICENSE.
import { RuleTester } from '../rule-tester';

const error = {
  messageId: 'no-useless-else',
  message: 'Unexpected `else` after a statement that exits.',
};

new RuleTester().run('no-useless-else', {} as never, {
  valid: [
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tdo {} while (maybeThrow(), process.exit(1));\n\t\t} catch {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tthrow error;\n\t\t\tprocess.exit(1);\n\t\t} catch {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tprocess.exit(maybeThrow());\n\t\t} catch {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tprocess.exit((maybeThrow(), 1));\n\t\t} catch {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tmaybeThrow();\n\t\t\tprocess.exit();\n\t\t} catch {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tbar();\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: "const MODE_GROUP = new Set(['foo', 'bar']);\n\nconst args = new Map();\n\nconst modes = new Set();\nif (args.has('modes')) {\n\tfor (const mode of args.get('modes').split(',')) {\n\t\tconst trimmed = mode.trim();\n\t\tif (MODE_GROUP.has(trimmed)) {\n\t\t\tmodes.add(trimmed);\n\t\t}\n\t}\n\n\tif (!modes.size) {\n\t\tthrow new Error('No valid modes');\n\t}\n} else {\n\tfor (const mode of MODE_GROUP) {\n\t\tmodes.add(mode);\n\t}\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const modes = new Set(['foo']);\nmodes.clear();\nif (condition) {\n\tmodes.size ? process.exit() : doSomething();\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nObject.defineProperty(object, 'value', {get() { return false; }});\n\nif (condition) {\n\tif (object.value) {\n\t\tthrow new Error();\n\t}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nObject.defineProperty(object, 'value', {get() { throw new Error(); }});\n\nif (condition) {\n\ttry {\n\t\tif (object.value && true) {\n\t\t\tthrow new Error();\n\t\t}\n\t} catch {}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nObject.defineProperty(object, 'value', {get() { throw new Error(); }});\n\nif (condition) {\n\ttry {\n\t\tif (object?.value) {\n\t\t\tthrow new Error();\n\t\t}\n\t} catch {}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nconst getterKey = 'value';\nObject.defineProperty(object, 'value', {get() { throw new Error(); }});\n\nif (condition) {\n\ttry {\n\t\tif (object[getterKey]) {\n\t\t\tthrow new Error();\n\t\t}\n\t} catch {}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nObject.defineProperty(object, 'value', {get() { throw new Error(); }});\n\nif (condition) {\n\ttry {\n\t\tobject.value(process.exit(1));\n\t} catch {}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nObject.defineProperty(object, 'value', {get() { throw new Error(); }});\n\nif (condition) {\n\ttry {\n\t\tobject.value.foo(process.exit(1));\n\t} catch {}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "const object = {value: true};\nObject.defineProperty(object, 'value', {get() { throw new Error(); }});\n\nif (condition) {\n\ttry {\n\t\t(object?.value)[process.exit(1)];\n\t} catch {}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "let modes = new Set(['foo']);\n\nif (condition) {\n\tif ((modes = new Set()).size) {\n\t\tthrow new Error();\n\t}\n} else {\n\tdoSomethingElse();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: "function qux() {\n\tconst modes = new Set(['foo']);\n\tmodes.clear();\n\n\tif (foo) {\n\t\tif (modes.size) {\n\t\t\treturn;\n\t\t}\n\t\tbar();\n\t} else {\n\t\tbaz();\n\t}\n}",
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tif (bar) {\n\t\t\treturn;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const value = foo\n\t? bar()\n\t: baz();',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\treturn bar();\n\t\t} catch {}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\ttry {\n\t\tclass Example {\n\t\t\tstatic {\n\t\t\t\tmaybeThrow();\n\t\t\t\tprocess.exit(1);\n\t\t\t}\n\t\t}\n\t} catch {}\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\twhile (bar) {\n\t\tbreak;\n\t}\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tswitch (bar) {\n\t\tcase baz:\n\t\t\tbreak;\n\t}\n} else {\n\tqux();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (bar) {\n\t\t\tcase 1:\n\t\t\t\treturn;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (bar) {\n\t\t\tcase 1:\n\t\t\t\tqux();\n\t\t\t\tbreak;\n\t\t\tdefault:\n\t\t\t\treturn;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'for (const foo of bar)\n\tif (foo) {\n\t\tcontinue;\n\t} else {\n\t\tbaz();\n\t}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\touter: for (const a of b) {\n\t\t\tfor (const c of d) {\n\t\t\t\tbreak outer;\n\t\t\t}\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (bar) {\n\t\t\tcase 1:\n\t\t\t\treturn;\n\t\t\tdefault:\n\t\t\t\tbreak;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\tbreak;\n\t\tdefault:\n\t\t\tbreak;\n\t\tcase process.exit(1):\n\t\t\t{}\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\treturn;\n\t\tdefault:\n\t\t\tbreak;\n\t\tcase process.exit(1):\n\t\t\t{}\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tdoSomething();\n\t\t} catch {\n\t\t\thandle();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tif (inner) {\n\t\t\treturn;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tdo {\n\t\t\tif (bar) {\n\t\t\t\tbreak;\n\t\t\t}\n\t\t} while (true);\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tfor (const x of xs) {\n\t\t\treturn x;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\treturn doSomething();\n\t\t} catch {\n\t\t\thandle();\n\t\t} finally {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tconst inner = () => {\n\t\t\treturn 1;\n\t\t};\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tfunction inner() {\n\t\t\treturn 1;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\treturn;\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\treturn;\n\tif (foo) {\n\t\tprocess.exit();\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tfalse && process.exit();\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux(process) {\n\tif (foo) {\n\t\tprocess.exit();\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: "import process from 'node:process';\n\nif (foo) {\n\tprocess.exit();\n} else {\n\tbaz();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tprocess.exitCode = 1;\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tprocess?.exit();\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tprocess.exit?.();\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tcallback?.(process.exit(1));\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tfoo?.bar[process.exit(1)];\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: "if (foo) {\n\tprocess['exit']();\n} else {\n\tbaz();\n}",
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tlib.process.exit();\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tnew process.exit();\n} else {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tif (bar) {\n\t\t\tprocess.exit();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function qux(value) {\n\tif (foo) {\n\t\ttry {\n\t\t\tswitch (value) {\n\t\t\tdefault:\n\t\t\t\tbreak;\n\t\t\tcase process.exit(1):\n\t\t\t\t{}\n\t\t\t}\n\t\t} catch {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
    },
    // Documentation examples.
    {
      code: 'function example() {\n\tif (foo) {\n\t\treturn;\n\t}\n\n\tbar();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (foo) {\n\tthrow new Error();\n}\n\nif (bar) {\n\tbaz();\n}',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    {
      code: 'function qux() {\n\tif (foo) {\n\t\touter: do {\n\t\t\twhile (true) {\n\t\t\t\tcontinue outer;\n\t\t\t}\n\t\t} while (process.exit(1));\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux(condition) {\n\tif (foo) {\n\t\ttry {\n\t\t\tcondition ? process.exit(1) : process.exit(2);\n\t\t} catch {}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux(condition) {\n\tif (foo) {\n\t\ttry {\n\t\t\tif (condition) {\n\t\t\t\tprocess.exit(1);\n\t\t\t} else {\n\t\t\t\tprocess.exit(2);\n\t\t\t}\n\t\t} catch {}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error, error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\treturn bar();\n\t\t} finally {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (bar) {\n\t\t\tcase 1:\n\t\t\t\treturn;\n\t\t\tdefault:\n\t\t\t\tthrow new Error();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\twhile (true) {\n\t\t\tdoSomething();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tfor (;;) {\n\t\t\tdoSomething();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tdo {\n\t\t\tdoSomething();\n\t\t} while (true);\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (bar) {\n\t\t\tcase 1:\n\t\t\tcase 2:\n\t\t\t\treturn;\n\t\t\tdefault:\n\t\t\t\tthrow new Error();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\treturn doSomething();\n\t\t} catch {\n\t\t\tthrow new Error();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tif (a) {\n\t\t\treturn;\n\t\t} else if (b) {\n\t\t\tthrow new Error();\n\t\t} else {\n\t\t\treturn;\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error, error],
    },
    {
      code: 'function qux() {\n\touter: for (const a of b) {\n\t\tif (foo) {\n\t\t\tcontinue outer;\n\t\t} else {\n\t\t\tbaz();\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tdoSomething();\n\t\t} finally {\n\t\t\treturn cleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else { bar(); }\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'if (foo) {\n\tthrow new Error();\n} else {\n\tbar();\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (false) {\n\t\treturn;\n\t} else {\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (true) {\n\t\treturn;\n\t} else {\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'while (foo) {\n\tif (bar) {\n\t\tbreak;\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'while (foo) {\n\tif (bar) {\n\t\tcontinue;\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'switch (foo) {\n\tcase bar:\n\t\tif (baz) {\n\t\t\tbreak;\n\t\t} else {\n\t\t\tqux();\n\t\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'class Foo {\n\tstatic {\n\t\tif (foo) {\n\t\t\tthrow new Error();\n\t\t} else {\n\t\t\tbar();\n\t\t}\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo)\n\t\treturn;\n\telse\n\t\tbar();\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else if (bar) {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tif (bar) {\n\t\t\treturn;\n\t\t} else {\n\t\t\tthrow new Error();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error, error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\t// Keep this comment.\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\t/*\n\t\tKeep this comment.\n\t\t*/\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar(`foo`);\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar(`\n\tleft\n`);\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: "function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar('foo\\\nbar');\n\t}\n}",
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\treturn (\n\t\t\t<div>{`\n\t\t\t\tmultiline\n\t\t\t\ttemplate\n\t\t\t`}</div>\n\t\t);\n\t}\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\treturn <pre>\n\t\t\tline\n\t\t</pre>;\n\t}\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\t<Widget title="first\n\t\t\tsecond" />;\n\t}\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\trender(<div>\n\t\t\ttext\n\t\t</div>);\n\t}\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function init() {\n\tif (tagName) {\n\t\taddTagToFooter(tagName);\n\t\treturn;\n\t} else {\n\t\tvoid addReleaseBanner(\n\t\t\t<>\n\t\t\t\tNo <ExplanationLink>stable version tags</ExplanationLink> for this PR.\n\t\t\t</>,\n\t\t\tsignal,\n\t\t);\n\t}\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\trender(<div>\n\t\t\ttext\n\t\t</div>, css`\n\t\t\tcolor: red;\n\t\t`);\n\t}\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tconst bar = 1;\n\t\tbaz(bar);\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tlet bar = 1;\n\t\tbaz(bar);\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tclass Bar {}\n\t\tbaz(Bar);\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tfunction bar() {}\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else /* comment */ {\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} /* comment */ else {\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar();\n\t} // trailing\n\tbaz();\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux(foo) {\n\tif (foo) {\n\t\treturn;\n\t} else function bar() {}\n}',
      filename: 'src/virtual.js',
      languageOptions: { sourceType: 'script', ecmaVersion: 5 },
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar()\n\t} baz()\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar = function() {}\n\t}\n\t(baz)();\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\t<Foo>\n\t\t\ttext\n\t\t</Foo>\n\t}\n\t<Bar />\n}',
      filename: 'src/virtual.jsx',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo)\n\t\treturn\n\telse\n\t\t[bar].forEach(baz)\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\ttype Bar = string;\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.ts',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tinterface Bar {}\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.ts',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tenum Bar {}\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.ts',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tnamespace Bar {}\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.ts',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tprocess.exit();\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttrue && process.exit();\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\t[...process.exit()];\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo)\n\t\tprocess.exit(1);\n\telse\n\t\tbaz();\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tsetup();\n\t\tprocess.exit(1);\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\t{\n\t\t\tprocess.exit();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\tif (false) {\n\t\t\t\t\tbreak;\n\t\t\t\t}\n\n\t\t\t\tprocess.exit(1);\n\t\t\tdefault:\n\t\t\t\tprocess.exit(2);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\tif (true) {\n\t\t\t\t\treturn;\n\t\t\t\t}\n\t\t\t\tbreak;\n\t\t\tdefault:\n\t\t\t\tprocess.exit(1);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tlabel: {\n\t\t\tif (false) {\n\t\t\t\tbreak label;\n\t\t\t}\n\n\t\t\tprocess.exit(1);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tprocess.exit(1);\n\t} else if (bar) {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tprocess.exit(1);\n\t\t\tcleanup();\n\t\t} finally {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tif (bar) {\n\t\t\treturn;\n\t\t} else {\n\t\t\tprocess.exit(1);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error, error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tthrow error;\n\t\t} catch {\n\t\t\tprocess.exit(1);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\tlocal: {\n\t\t\t\t\tbreak local;\n\t\t\t\t}\n\n\t\t\t\tprocess.exit(1);\n\t\t\tdefault:\n\t\t\t\tprocess.exit(2);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'outer: while (condition) {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\tbreak outer;\n\t\tdefault:\n\t\t\t\tprocess.exit(1);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'outer: while (condition) {\n\tif (foo) {\n\t\tswitch (value) {\n\t\t\tcase 1:\n\t\t\t\tcontinue outer;\n\t\tdefault:\n\t\t\t\tprocess.exit(1);\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function qux() {\n\tif (foo) {\n\t\ttry {\n\t\t\tprocess.exit();\n\t\t} finally {\n\t\t\tcleanup();\n\t\t}\n\t} else {\n\t\tbaz();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    // CRLF.
    {
      code: 'function run(item) {\r\n\tif (item.isActive) {\r\n\t\treturn process(item);\r\n\t} else {\r\n\t\tsave(item);\r\n\t\tlog(item);\r\n\t}\r\n}\r\n',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function run(item) {\r\n\tif (item.isActive) {\r\n\t\treturn process(item);\r\n\t} else save(item);\r\n}\r\n',
      filename: 'src/virtual.js',
      errors: [error],
    },
    // Indentation.
    {
      code: 'function run(item) {\n  if (item.isActive) {\n    return process(item);\n  } else {\n    if (item.isReady) {\n      save(item);\n    }\n\n    log(item);\n  }\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    // EmptyBlocks.
    {
      code: 'function run(item) {\n\tif (item.isActive) {\n\t\treturn process(item);\n\t} else {}\n}\n',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'function run(item) {\n\tif (item.isActive) {\n\t\treturn process(item);\n\t} else {\n\t}\n}\n',
      filename: 'src/virtual.js',
      errors: [error],
    },
    // Documentation.
    {
      code: 'function example() {\n\tif (foo) {\n\t\treturn;\n\t} else {\n\t\tbar();\n\t}\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
    {
      code: 'if (foo) {\n\tthrow new Error();\n} else if (bar) {\n\tbaz();\n}',
      filename: 'src/virtual.js',
      errors: [error],
    },
  ],
});
