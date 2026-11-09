// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-string-raw.js
import { RuleTester, type ValidTestCase } from '../rule-tester';

const prefer = {
  messageId: 'prefer-string-raw',
  message: '`String.raw` should be used to avoid escaping `\\`.',
};
const unnecessary = {
  messageId: 'unnecessary-string-raw',
  message:
    'Using `String.raw` is unnecessary as the string does not contain any `\\`.',
};

const syntax = (code: string, filename = 'case.js'): ValidTestCase => ({
  code,
  filename: `src/${filename}`,
  languageOptions: { parserOptions: { project: false, projectService: false } },
});

const typed = (code: string): ValidTestCase => ({
  code,
  filename: 'src/virtual.ts',
});

const withErrors = (input: ValidTestCase, ...errors: (typeof prefer)[]) => ({
  ...input,
  errors,
});

// The Go upstream suite also asserts diagnostic ranges and applied edits.
new RuleTester().run('prefer-string-raw', {} as never, {
  valid: [
    // TypeScript literal and enum contexts.
    syntax('type Path = `a\\\\b`;', 'case.ts'),
    syntax('enum Paths {File = `a\\\\b`}', 'case.ts'),
    // String literals with and without type information.
    typed("const path = 'a\\\\b';"),
    typed("const path: 'a\\\\b' = 'a\\\\b';"),
    typed(
      "type PathText = 'a\\\\b' | undefined; const path: PathText = 'a\\\\b'; const exact: 'a\\\\b' = path;",
    ),
    typed("declare function use(path: 'a\\\\b'): void; use('a\\\\b');"),
    typed("const options = {path: 'a\\\\b'} as const;"),
    typed("const options = ['a\\\\b'] as const;"),
    typed("const options = {nested: {path: 'a\\\\b'}} as const;"),
    typed("class Paths {readonly path = 'a\\\\b';}"),
    typed("const path = 'a\\\\b' satisfies string;"),
    typed("const path = 'a\\\\b' satisfies 'a\\\\b';"),
    typed(
      "declare function use<const Type>(value: Type): Type; const path = use('a\\\\b');",
    ),
    typed(
      "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use('a\\\\b');",
    ),
    typed("const path = (/* keep */ 'a\\\\b');"),
    typed("function path(): string {return 'a\\\\b';}"),
    typed(
      "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = 'a\\\\b';\nif (getTag(path) === 'string') {}",
    ),
    // String literals under const assertions.
    syntax("const path = 'a\\\\b' as const;", 'file.ts'),
    syntax("const path = <const>'a\\\\b';", 'file.ts'),
    typed("const path = 'a\\\\b' as const;"),
    typed("const path = <const>'a\\\\b';"),
    // Template literals with and without type information.
    typed('const path = `a\\\\b`;'),
    typed("const path: 'a\\\\b' = `a\\\\b`;"),
    typed(
      "type PathText = 'a\\\\b' | undefined; const path: PathText = `a\\\\b`; const exact: 'a\\\\b' = path;",
    ),
    typed("declare function use(path: 'a\\\\b'): void; use(`a\\\\b`);"),
    typed('const options = {path: `a\\\\b`} as const;'),
    typed('const options = [`a\\\\b`] as const;'),
    typed('const options = {nested: {path: `a\\\\b`}} as const;'),
    typed('class Paths {readonly path = `a\\\\b`;}'),
    typed('const path = `a\\\\b` satisfies string;'),
    typed("const path = `a\\\\b` satisfies 'a\\\\b';"),
    typed(
      'declare function use<const Type>(value: Type): Type; const path = use(`a\\\\b`);',
    ),
    typed(
      "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use(`a\\\\b`);",
    ),
    typed('const path = (/* keep */ `a\\\\b`);'),
    typed('function path(): string {return `a\\\\b`;}'),
    typed(
      "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = `a\\\\b`;\nif (getTag(path) === 'string') {}",
    ),
    // Template literals under const assertions.
    syntax('const path = `a\\\\b` as const;', 'file.ts'),
    syntax('const path = <const>`a\\\\b`;', 'file.ts'),
    typed('const path = `a\\\\b` as const;'),
    typed('const path = <const>`a\\\\b`;'),
    // Type-sensitive templates and TypeScript extensions.
    typed(
      'declare const suffix: string; declare function use<const Type>(value: Type): Type; const result = use({path: `a\\\\b${suffix}`});',
    ),
    typed(
      'declare const suffix: string; const path: `a\\\\b${string}` = `a\\\\b${suffix}`;',
    ),
    // Snapshot cases for string literals.
    syntax("a = '\\''"),
    syntax('a = "\\\\a \\\n\tb"'),
    syntax("a = 'a\\\\b\\u{51}c'"),
    syntax('a = "a\\\\b`"'),
    syntax('a = "a\\\\b${foo}"'),
    syntax('a = "a\\\\b\\tc"'),
    syntax('a = "a\\\\b\\nc"'),
    syntax("a = '\\\\'"),
    syntax("a = 'a\\\\b\\\"'"),
    // Snapshot cases for template literals.
    syntax('a = `a`'),
    syntax('a = `${foo}`'),
    syntax('a = `a${100}b`'),
    syntax('a = `a\\t${foo.bar}b\\\\c`'),
    syntax('a = `${foo}\\\\a${bar}\\``'),
    syntax('a = `a\\${`'),
    syntax('a = `a$\\{`'),
    syntax("a = `${a}\\'${b}\\\\`"),
    syntax('a = `\\"a\\\\b`'),
    syntax('a = `\\\\a${foo}b\\\\${foo}`'),
    syntax('a = String.raw`a\\\\b`'),
    syntax('a = `\\\\a \\\nb`'),
    // Snapshot cases for unnecessary String.raw.
    syntax('a = foo`ab`'),
    syntax('a = foo().bar`ab`'),
    syntax('a = foo.bar()`ab`'),
    syntax('a = String["raw"]`ab`'),
    syntax('a = foo.raw`ab`'),
    syntax('a = String.foo`ab`'),
    syntax('a = String.raw`a\\b`'),
    syntax('a = String.raw`a\\b${foo}cd`'),
    syntax('a = String.raw`ab${foo}c\\nd`'),
    syntax('a = String.raw`a\n\tb\\c\n\tde`'),
    syntax('a = String.raw\n\t// Comment\n\t`ab\\c`'),
    // Shadowing, comments and yield.
    syntax('const String = {raw() {}};\na = String.raw`abc`'),
    // Snapshot cases for restricted syntax and Jest inline snapshots.
    syntax("'a\\\\b';", 'case.ts'),
    syntax("import 'a\\\\b';", 'case.ts'),
    syntax("export {} from 'a\\\\b';", 'case.ts'),
    syntax("export * from 'a\\\\b';", 'case.ts'),
    syntax("import 'm' with {'a\\\\b': 'v'};", 'case.ts'),
    syntax("export {} from 'm' with {'a\\\\b': 'v'};", 'case.ts'),
    syntax("import 'm' with {k: 'a\\\\b'};", 'case.ts'),
    syntax("export {} from 'm' with {k: 'a\\\\b'};", 'case.ts'),
    syntax("import {'a\\\\b' as s} from 'm';", 'case.ts'),
    syntax("export {'a\\\\b' as s} from 'm';", 'case.ts'),
    syntax("export {s as 'a\\\\b'} from 'm';", 'case.ts'),
    syntax("export * as 'a\\\\b' from 'm';", 'case.ts'),
    syntax("<Component attribute='a\\\\b' />", 'case.tsx'),
    syntax("enum E {'a\\\\b' = 1}", 'case.ts'),
    syntax("enum E {K = 'a\\\\b'}", 'case.ts'),
    syntax("module 'a\\\\b' {}", 'case.ts'),
    syntax("import type T = require('a\\\\b');", 'case.ts'),
    syntax("type T = 'a\\\\b';", 'case.ts'),
    syntax("type T = import('a\\\\b');", 'case.ts'),
    syntax("({ 'a\\\\b': 1 })", 'case.ts'),
    syntax("class C { 'a\\\\b' = 1 }", 'case.ts'),
    syntax("class C { 'a\\\\b'(){} }", 'case.ts'),
    syntax("class C { accessor 'a\\\\b' = 1 }", 'case.ts'),
    syntax("abstract class C { abstract 'a\\\\b' }", 'case.ts'),
    syntax("abstract class C { abstract ['a\\\\b'] }", 'case.ts'),
    syntax("abstract class C { abstract 'a\\\\b'() }", 'case.ts'),
    syntax("abstract class C { abstract ['a\\\\b']() }", 'case.ts'),
    syntax("abstract class C { abstract accessor 'a\\\\b' }", 'case.ts'),
    syntax("abstract class C { abstract accessor ['a\\\\b'] }", 'case.ts'),
    syntax("interface I { 'a\\\\b' }", 'case.ts'),
    syntax("interface I { ['a\\\\b'] }", 'case.ts'),
    syntax("interface I { 'a\\\\b'(): void }", 'case.ts'),
    syntax("interface I { ['a\\\\b'](): void }", 'case.ts'),
    syntax("expect(foo).toMatchInlineSnapshot('a\\\\b')", 'case.ts'),
    syntax('expect(foo).toMatchInlineSnapshot(`a\\\\b`)', 'case.ts'),
    // Documentation examples.
    syntax('const file = String.raw`C:\\windows\\style\\path\\to\\file.js`;'),
    syntax('const regexp = new RegExp(String.raw`foo\\.bar`);'),
    syntax(
      'const file = String.raw`C:\\windows\\temp\\my-app-${process.pid}.log`;',
    ),
    syntax("const noBackslash = 'foobar'"),
    // Shared upstream empty-file smoke test.
    syntax(''),
  ],
  invalid: [
    // TypeScript literal and enum contexts.
    withErrors(
      syntax(
        "type Separator = 'a\\\\b' | 'c\\\\d'; const separator: Separator = 'a\\\\b';",
        'case.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax(
        "const options = {path: 'C:\\\\Users'} as const; const exact: 'C:\\\\Users' = options.path;",
        'case.ts',
      ),
      prefer,
    ),
    // String literals with and without type information.
    withErrors(syntax("const path = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(syntax("const path: 'a\\\\b' = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(
      syntax(
        "type PathText = 'a\\\\b' | undefined; const path: PathText = 'a\\\\b'; const exact: 'a\\\\b' = path;",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax(
        "declare function use(path: 'a\\\\b'): void; use('a\\\\b');",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax("const options = {path: 'a\\\\b'} as const;", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("const options = ['a\\\\b'] as const;", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("const options = {nested: {path: 'a\\\\b'}} as const;", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("class Paths {readonly path = 'a\\\\b';}", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("const path = 'a\\\\b' satisfies string;", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("const path = 'a\\\\b' satisfies 'a\\\\b';", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax(
        "declare function use<const Type>(value: Type): Type; const path = use('a\\\\b');",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax(
        "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use('a\\\\b');",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax("const path = (/* keep */ 'a\\\\b');", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("function path(): string {return 'a\\\\b';}", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax(
        "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = 'a\\\\b';\nif (getTag(path) === 'string') {}",
        'file.ts',
      ),
      prefer,
    ),
    // Widened string bindings.
    withErrors(syntax("const path: string = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(syntax("let path: string = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(syntax("let path = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(syntax("var path = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(
      syntax("const path: string = (/* keep */ 'a\\\\b');", 'file.ts'),
      prefer,
    ),
    withErrors(typed("const path: string = 'a\\\\b';"), prefer),
    withErrors(typed("let path: string = 'a\\\\b';"), prefer),
    withErrors(typed("let path = 'a\\\\b';"), prefer),
    withErrors(typed("var path = 'a\\\\b';"), prefer),
    withErrors(typed("const path: string = (/* keep */ 'a\\\\b');"), prefer),
    // Nullable and aliased string bindings.
    withErrors(
      syntax(
        "type PathText = string; const path: PathText = 'a\\\\b';",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      typed("type PathText = string; const path: PathText = 'a\\\\b';"),
      prefer,
    ),
    withErrors(
      syntax("const path: string | undefined = 'a\\\\b';", 'file.ts'),
      prefer,
    ),
    withErrors(typed("const path: string | undefined = 'a\\\\b';"), prefer),
    withErrors(
      syntax("let path: string | null = 'a\\\\b';", 'file.ts'),
      prefer,
    ),
    withErrors(typed("let path: string | null = 'a\\\\b';"), prefer),
    withErrors(
      syntax(
        "type PathText = string | undefined; const path: PathText = 'a\\\\b';",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      typed(
        "type PathText = string | undefined; const path: PathText = 'a\\\\b';",
      ),
      prefer,
    ),
    // Template literals with and without type information.
    withErrors(syntax('const path = `a\\\\b`;', 'file.ts'), prefer),
    withErrors(syntax("const path: 'a\\\\b' = `a\\\\b`;", 'file.ts'), prefer),
    withErrors(
      syntax(
        "type PathText = 'a\\\\b' | undefined; const path: PathText = `a\\\\b`; const exact: 'a\\\\b' = path;",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax(
        "declare function use(path: 'a\\\\b'): void; use(`a\\\\b`);",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax('const options = {path: `a\\\\b`} as const;', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax('const options = [`a\\\\b`] as const;', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax('const options = {nested: {path: `a\\\\b`}} as const;', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax('class Paths {readonly path = `a\\\\b`;}', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax('const path = `a\\\\b` satisfies string;', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax("const path = `a\\\\b` satisfies 'a\\\\b';", 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax(
        'declare function use<const Type>(value: Type): Type; const path = use(`a\\\\b`);',
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax(
        "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use(`a\\\\b`);",
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      syntax('const path = (/* keep */ `a\\\\b`);', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax('function path(): string {return `a\\\\b`;}', 'file.ts'),
      prefer,
    ),
    withErrors(
      syntax(
        "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = `a\\\\b`;\nif (getTag(path) === 'string') {}",
        'file.ts',
      ),
      prefer,
    ),
    // Widened template bindings.
    withErrors(syntax('const path: string = `a\\\\b`;', 'file.ts'), prefer),
    withErrors(syntax('let path: string = `a\\\\b`;', 'file.ts'), prefer),
    withErrors(syntax('let path = `a\\\\b`;', 'file.ts'), prefer),
    withErrors(syntax('var path = `a\\\\b`;', 'file.ts'), prefer),
    withErrors(
      syntax('const path: string = (/* keep */ `a\\\\b`);', 'file.ts'),
      prefer,
    ),
    withErrors(typed('const path: string = `a\\\\b`;'), prefer),
    withErrors(typed('let path: string = `a\\\\b`;'), prefer),
    withErrors(typed('let path = `a\\\\b`;'), prefer),
    withErrors(typed('var path = `a\\\\b`;'), prefer),
    withErrors(typed('const path: string = (/* keep */ `a\\\\b`);'), prefer),
    // Nullable and aliased template bindings.
    withErrors(
      syntax(
        'type PathText = string; const path: PathText = `a\\\\b`;',
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      typed('type PathText = string; const path: PathText = `a\\\\b`;'),
      prefer,
    ),
    withErrors(
      syntax('const path: string | undefined = `a\\\\b`;', 'file.ts'),
      prefer,
    ),
    withErrors(typed('const path: string | undefined = `a\\\\b`;'), prefer),
    withErrors(
      syntax('let path: string | null = `a\\\\b`;', 'file.ts'),
      prefer,
    ),
    withErrors(typed('let path: string | null = `a\\\\b`;'), prefer),
    withErrors(
      syntax(
        'type PathText = string | undefined; const path: PathText = `a\\\\b`;',
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      typed(
        'type PathText = string | undefined; const path: PathText = `a\\\\b`;',
      ),
      prefer,
    ),
    // An interpolated template in a return statement.
    withErrors(
      syntax(
        'declare const suffix: string; function path(): string {return`a\\\\b${suffix}`;}',
        'file.ts',
      ),
      prefer,
    ),
    withErrors(
      typed(
        'declare const suffix: string; function path(): string {return`a\\\\b${suffix}`;}',
      ),
      prefer,
    ),
    // Type-sensitive templates and TypeScript extensions.
    withErrors(syntax("const path = 'a\\\\b';", 'file.ts'), prefer),
    withErrors(syntax("const path = 'a\\\\b';", 'file.mts'), prefer),
    withErrors(syntax("const path = 'a\\\\b';", 'file.cts'), prefer),
    withErrors(syntax("const path = 'a\\\\b';", 'file.tsx'), prefer),
    // Upstream uses the TypeScript parser on file.js. The native parser selects JavaScript behavior.
    withErrors(syntax("const path = 'a\\\\b';", 'file.js'), prefer),
    withErrors(
      typed('declare const suffix: string; const path = `a\\\\b${suffix}`;'),
      prefer,
    ),
    withErrors(syntax('const path = String.raw`abc`;', 'file.ts'), unnecessary),
    // Snapshot cases for string literals.
    withErrors(syntax("TEST_STRING = 'a\\\\b';"), prefer),
    withErrors(syntax("function a() {return'a\\\\b'}"), prefer),
    withErrors(syntax('const foo = "foo \\\\x46";'), prefer),
    withErrors(syntax("a = 'a\\\\b\\''"), prefer),
    withErrors(syntax('a = "a\\\\b\\""'), prefer),
    // Snapshot cases for template literals.
    withErrors(syntax('a = `a\\\\b`'), prefer),
    withErrors(syntax('function a() {return`a\\\\b${foo}cd`}'), prefer),
    withErrors(syntax('a = {[`a\\\\b${foo}cd${foo.bar}e\\\\f`]: b}'), prefer),
    withErrors(syntax('a = `a${foo}${foo.bar}b\\\\c`'), prefer),
    withErrors(syntax('a = `a\\\\b${"c\\\\d"}e`'), prefer, prefer),
    withErrors(syntax('a = `\\\\a\nb`'), prefer),
    withErrors(syntax('a = `\\\\a${foo}\nb${bar}c\nd\\\\\\\\e`'), prefer),
    withErrors(syntax('a = `a\\\\b${ foo /* bar */}c\\\\d`'), prefer),
    withErrors(syntax('a = `a\\\\b${ foo + bar }`'), prefer),
    withErrors(syntax('a = `${ foo .bar }a\\\\b`'), prefer),
    // Snapshot cases for unnecessary String.raw.
    withErrors(syntax('a = String.raw`abc`'), unnecessary),
    withErrors(syntax('a = String.raw`ab${foo}cd`'), unnecessary),
    withErrors(syntax('a = String.raw`ab"c`'), unnecessary),
    withErrors(syntax("a = String.raw`ab'c`"), unnecessary),
    withErrors(syntax('a = String.raw`ab\'"c`'), unnecessary),
    withErrors(syntax('a = String.raw`ab\r\nc`'), unnecessary),
    withErrors(syntax('a = String.raw`a\n\tbc\n\tde`'), unnecessary),
    withErrors(syntax('a = String.raw`\na${foo}b\n${bar}cd`'), unnecessary),
    withErrors(syntax('a\nString.raw`abc`'), unnecessary),
    withErrors(syntax('a\nString.raw`a${b}`'), unnecessary),
    withErrors(syntax('a = String.raw\n// Comment\n`ab\nc`'), unnecessary),
    withErrors(syntax('a = String.raw  /* comment */  `ab\nc`'), unnecessary),
    withErrors(
      syntax('function foo() {\n\treturn String.raw`abc`\n}'),
      unnecessary,
    ),
    withErrors(
      syntax(
        'function foo() {\n\treturn String.raw\n\t\t\t// comment\n\t\t\t`abc`\n}',
      ),
      unnecessary,
    ),
    withErrors(
      syntax(
        'function foo() {\n\treturn (String.raw\n\t\t\t// already parenthesized\n\t\t\t`abc`)\n}',
      ),
      unnecessary,
    ),
    withErrors(
      syntax(
        'function foo() {\n\treturn (String.raw) // remove parenthesis\n\t\t\t`abc`\n}',
      ),
      unnecessary,
    ),
    withErrors(syntax('String.raw`use strict`'), unnecessary),
    // Shadowing, comments and yield.
    withErrors(syntax('String /* keep */ .raw`abc`;'), unnecessary),
    withErrors(
      syntax('function* foo() {\n\tyield String.raw\n\t\t`abc`;\n}'),
      unnecessary,
    ),
    withErrors(
      syntax('function* foo() {\n\tyield String.raw\n\t\t`a${value}`;\n}'),
      unnecessary,
    ),
    // Snapshot cases for restricted syntax and Jest inline snapshots.
    withErrors(syntax("({ ['a\\\\b']: 1 })", 'case.ts'), prefer),
    withErrors(syntax("class C { ['a\\\\b'] = 1 }", 'case.ts'), prefer),
    withErrors(syntax("class C { ['a\\\\b'](){} }", 'case.ts'), prefer),
    withErrors(
      syntax("class C { accessor ['a\\\\b'] = 1 }", 'case.ts'),
      prefer,
    ),
    withErrors(
      syntax('expect(\'a\\\\b\').toMatchInlineSnapshot("")', 'case.ts'),
      prefer,
    ),
    withErrors(
      syntax('expect(`a\\\\b`).toMatchInlineSnapshot(``)', 'case.ts'),
      prefer,
    ),
    // Line terminators.
    withErrors(syntax('a = String.raw`ab\rd`;'), unnecessary),
    withErrors(syntax('a = String.raw`ab\nd`;'), unnecessary),
    withErrors(syntax('a = String.raw`ab\u2028d`;'), unnecessary),
    withErrors(syntax('a = String.raw`abc`;'), unnecessary),
    // Removing the tag after return, throw or yield.
    withErrors(
      syntax('function read(ready) { return String.raw\n`x` && ready; }'),
      unnecessary,
    ),
    withErrors(
      syntax('function read(ready) { throw String.raw\n`x` && ready; }'),
      unnecessary,
    ),
    withErrors(
      syntax('function* read(ready) { yield String.raw\n`x` && ready; }'),
      unnecessary,
    ),
    withErrors(
      syntax('function read(ready) { return String.raw`x` && ready; }'),
      unnecessary,
    ),
    withErrors(
      syntax('function read(ready) { return (String.raw\n`x` && ready); }'),
      unnecessary,
    ),
    withErrors(
      syntax(
        'function read(ready) { return String.raw /* keep */\n`x` && ready; }',
      ),
      unnecessary,
    ),
    // Documentation examples.
    withErrors(
      syntax('const file = "C:\\\\windows\\\\style\\\\path\\\\to\\\\file.js";'),
      prefer,
    ),
    withErrors(syntax("const regexp = new RegExp('foo\\\\.bar');"), prefer),
    withErrors(
      syntax(
        'const file = `C:\\\\windows\\\\temp\\\\my-app-${process.pid}.log`;',
      ),
      prefer,
    ),
    withErrors(syntax('const noBackslash = String.raw`foobar`'), unnecessary),
  ],
});
