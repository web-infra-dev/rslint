// Upstream tests and documentation: eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-query-selector.js
// Inputs mirror upstream; expectations include the documented autofix safety differences.
import assert from 'node:assert/strict';
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import { RuleTester } from '../rule-tester';

function valid(
  code: string,
  options: { allowWithVariables?: boolean }[] = [],
  filename = 'src/virtual.js',
) {
  return { code, options, filename };
}

function invalid(
  code: string,
  output: string | null,
  locations: [
    method: string,
    replacement: string,
    line: number,
    column: number,
  ][],
  options: { allowWithVariables?: boolean }[] = [],
  filename = 'src/virtual.js',
) {
  return {
    ...valid(code, options, filename),
    output,
    errors: locations.map(([method, replacement, line, column]) => ({
      messageId: 'prefer-query-selector',
      message: `Prefer \`.${replacement}()\` over \`.${method}()\`.`,
      line,
      column,
      endLine: line,
      endColumn: column + method.length,
    })),
  };
}

const cases = {
  valid: [
    // test/prefer-query-selector.js, including all not-dom-node-types fixtures.
    valid('new document.getElementById(foo);'),
    valid('getElementById(foo);'),
    valid("document['getElementById'](bar);"),
    valid('document[getElementById](bar);'),
    valid('document.foo(bar);'),
    valid('document.getElementById();'),
    valid('document.getElementsByClassName("foo", "bar");'),
    valid('document.getElementById(...["id"]);'),
    valid('([]).getElementById(foo)'),
    valid('([element]).getElementById(foo)'),
    valid('([...elements]).getElementById(foo)'),
    valid('(() => {}).getElementById(foo)'),
    valid('(class Node {}).getElementById(foo)'),
    valid('(function() {}).getElementById(foo)'),
    valid('(0).getElementById(foo)'),
    valid('(1).getElementById(foo)'),
    valid('(0.1).getElementById(foo)'),
    valid('("").getElementById(foo)'),
    valid('("string").getElementById(foo)'),
    valid('(/regex/).getElementById(foo)'),
    valid('(null).getElementById(foo)'),
    valid('(0n).getElementById(foo)'),
    valid('(1n).getElementById(foo)'),
    valid('(true).getElementById(foo)'),
    valid('(false).getElementById(foo)'),
    valid('({}).getElementById(foo)'),
    valid('(`templateLiteral`).getElementById(foo)'),
    valid('(undefined).getElementById(foo)'),
    valid('document.querySelector("#foo");'),
    valid('document.querySelector(".bar");'),
    valid('document.querySelector("main #foo .bar");'),
    valid('document.querySelectorAll(".foo .bar");'),
    valid('document.querySelectorAll("li a");'),
    valid('document.querySelector("li").querySelectorAll("a");'),
    valid('document.getElementsByName();'),
    valid('document.getElementById(someId);', [{ allowWithVariables: true }]),
    valid('document.getElementsByClassName(someClass);', [
      { allowWithVariables: true },
    ]),
    valid('document.getElementsByClassName(fn());', [
      { allowWithVariables: true },
    ]),
    valid('document.getElementsByClassName(`${someClass}`);', [
      { allowWithVariables: true },
    ]),
    valid('document.getElementById(`${someId}`);', [
      { allowWithVariables: true },
    ]),
    // Upstream documentation examples.
    valid("document.querySelector('#foo');"),
    valid("document.querySelectorAll('.foo');"),
    valid("document.querySelectorAll('.foo.bar');"),
    valid("document.querySelectorAll('main');"),
    valid(
      "const singleClassName = 'foo';\ndocument.querySelectorAll(`.${CSS.escape(singleClassName)}`);",
    ),
    valid("document.querySelector('form');"),
    valid(
      'document.getElementById(someId);\ndocument.getElementsByClassName(someClass);',
      [{ allowWithVariables: true }],
    ),
  ],
  invalid: [
    // test/prefer-query-selector.js.
    invalid(
      'document.getElementById("foo");',
      'document.querySelector("#foo");',
      [['getElementById', 'querySelector', 1, 10]],
    ),
    invalid('document.getElementById("#foo");', null, [
      ['getElementById', 'querySelector', 1, 10],
    ]),
    invalid('document.getElementById(".foo");', null, [
      ['getElementById', 'querySelector', 1, 10],
    ]),
    invalid('document.getElementById("foo bar");', null, [
      ['getElementById', 'querySelector', 1, 10],
    ]),
    invalid('document.getElementById("foo:bar");', null, [
      ['getElementById', 'querySelector', 1, 10],
    ]),
    invalid(
      'document.getElementById("foo" as string);',
      'document.querySelector("#foo" as string);',
      [['getElementById', 'querySelector', 1, 10]],
      [],
      'src/virtual.ts',
    ),
    invalid(
      'document.getElementById("#foo" as string);',
      null,
      [['getElementById', 'querySelector', 1, 10]],
      [],
      'src/virtual.ts',
    ),
    invalid(
      'document.getElementsByClassName("foo");',
      'document.querySelectorAll(".foo");',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid('document.getElementsByClassName(".foo");', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName("#foo");', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName("foo:bar");', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'document.getElementsByClassName("foo" as string);',
      'document.querySelectorAll(".foo" as string);',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [],
      'src/virtual.ts',
    ),
    invalid(
      'document.getElementsByClassName(".foo" as string);',
      null,
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [],
      'src/virtual.ts',
    ),
    invalid(
      'element.getElementsByClassName("foo");',
      'element.querySelectorAll(".foo");',
      [['getElementsByClassName', 'querySelectorAll', 1, 9]],
    ),
    invalid('element.getElementsByClassName(".foo");', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 9],
    ]),
    invalid(
      'document.getElementsByClassName("foo bar");',
      'document.querySelectorAll(".foo.bar");',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByTagName("foo");',
      'document.querySelectorAll("foo");',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid('document.getElementsByTagName("foo:bar");', null, [
      ['getElementsByTagName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'document.getElementsByTagName("*");',
      'document.querySelectorAll("*");',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'element.getElementsByTagName("foo");',
      'element.querySelectorAll("foo");',
      [['getElementsByTagName', 'querySelectorAll', 1, 9]],
    ),
    invalid('element.getElementsByTagName("foo bar");', null, [
      ['getElementsByTagName', 'querySelectorAll', 1, 9],
    ]),
    invalid('document.getElementById("");', null, [
      ['getElementById', 'querySelector', 1, 10],
    ]),
    invalid(
      "document.getElementById('foo');",
      "document.querySelector('#foo');",
      [['getElementById', 'querySelector', 1, 10]],
    ),
    invalid(
      "document.getElementsByClassName('foo');",
      "document.querySelectorAll('.foo');",
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "document.getElementsByClassName('foo bar');",
      "document.querySelectorAll('.foo.bar');",
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "document.getElementsByTagName('foo');",
      "document.querySelectorAll('foo');",
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid("document.getElementsByClassName('');", null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'document.getElementById(`foo`);',
      'document.querySelector(`#foo`);',
      [['getElementById', 'querySelector', 1, 10]],
    ),
    invalid('document.getElementById(`#foo`);', null, [
      ['getElementById', 'querySelector', 1, 10],
    ]),
    invalid(
      'document.getElementsByClassName(`foo`);',
      'document.querySelectorAll(`.foo`);',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid('document.getElementsByClassName(`.foo`);', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'element.getElementsByClassName(`foo`);',
      'element.querySelectorAll(`.foo`);',
      [['getElementsByClassName', 'querySelectorAll', 1, 9]],
    ),
    invalid(
      'document.getElementsByClassName(`foo bar`);',
      'document.querySelectorAll(`.foo.bar`);',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByTagName(`foo`);',
      'document.querySelectorAll(`foo`);',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'element.getElementsByTagName(`foo`);',
      'element.querySelectorAll(`foo`);',
      [['getElementsByTagName', 'querySelectorAll', 1, 9]],
    ),
    invalid('document.getElementsByTagName(``);', null, [
      ['getElementsByTagName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName(`${fn()}`);', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName(`foo ${undefined}`);', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName(null);', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'document.getElementsByTagName(null);',
      'document.querySelectorAll(null);',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid('document.getElementsByClassName(fn());', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName("foo" + fn());', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByClassName(foo + "bar");', null, [
      ['getElementsByClassName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'for (const div of document.body.getElementById("id").getElementsByClassName("class")) {\n\tconsole.log(div.getElementsByTagName("div"));\n}',
      'for (const div of document.body.querySelector("#id").querySelectorAll(".class")) {\n\tconsole.log(div.querySelectorAll("div"));\n}',
      [
        ['getElementById', 'querySelector', 1, 33],
        ['getElementsByClassName', 'querySelectorAll', 1, 54],
        ['getElementsByTagName', 'querySelectorAll', 2, 18],
      ],
    ),
    invalid('e.getElementById(3)', null, [
      ['getElementById', 'querySelector', 1, 3],
    ]),
    invalid(
      'document.getElementsByName("foo");',
      'document.querySelectorAll("[name=\'foo\']");',
      [['getElementsByName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "document.getElementsByName('foo');",
      'document.querySelectorAll(\'[name="foo"]\');',
      [['getElementsByName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByName(`foo`);',
      "document.querySelectorAll(`[name='foo']`);",
      [['getElementsByName', 'querySelectorAll', 1, 10]],
    ),
    invalid("document.getElementsByName(`${'foo'}`);", null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByName(null);', null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByName("");', null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByName(foo + "bar");', null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'document.getElementsByName("multiple name should be fixable");',
      'document.querySelectorAll("[name=\'multiple name should be fixable\']");',
      [['getElementsByName', 'querySelectorAll', 1, 10]],
    ),
    invalid('document.getElementsByName("foo\'bar");', null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid("document.getElementsByName('foo\"bar');", null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid("document.getElementsByName(`foo'bar`);", null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByName("foo\\\\bar");', null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid('document.getElementsByName("foo\\nbar");', null, [
      ['getElementsByName', 'querySelectorAll', 1, 10],
    ]),
    invalid(
      'document.getElementsByName("foo\\\\bar" as string);',
      null,
      [['getElementsByName', 'querySelectorAll', 1, 10]],
      [],
      'src/virtual.ts',
    ),
    invalid(
      'document.getElementsByTagName("form")[0].addEventListener("submit", submitFunction);',
      'document.querySelectorAll("form")[0].addEventListener("submit", submitFunction);',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByTagName("form").item(0).submit();',
      'document.querySelector("form").submit();',
      [['getElementsByTagName', 'querySelector', 1, 10]],
    ),
    invalid(
      'document.getElementsByClassName("submit-button").at(0).click();',
      'document.querySelectorAll(".submit-button").at(0).click();',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByName("email")[0].focus();',
      'document.querySelectorAll("[name=\'email\']")[0].focus();',
      [['getElementsByName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByName("email").item(0).focus();',
      'document.querySelector("[name=\'email\']").focus();',
      [['getElementsByName', 'querySelector', 1, 10]],
    ),
    invalid(
      'delete document.getElementsByTagName("form").at(0);',
      'delete document.querySelectorAll("form").at(0);',
      [['getElementsByTagName', 'querySelectorAll', 1, 17]],
    ),
    invalid(
      'for (document.getElementsByTagName("form")[0] in object) {}',
      'for (document.querySelectorAll("form")[0] in object) {}',
      [['getElementsByTagName', 'querySelectorAll', 1, 15]],
    ),
    invalid(
      'for (document.getElementsByTagName("form")[0] of iterable) {}',
      'for (document.querySelectorAll("form")[0] of iterable) {}',
      [['getElementsByTagName', 'querySelectorAll', 1, 15]],
    ),
    invalid(
      'document.getElementsByTagName("form") /* keep */ [0].addEventListener("submit", submitFunction);',
      'document.querySelectorAll("form") /* keep */ [0].addEventListener("submit", submitFunction);',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementsByTagName("form").item(/* keep */ 0).addEventListener("submit", submitFunction);',
      'document.querySelectorAll("form").item(/* keep */ 0).addEventListener("submit", submitFunction);',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      'document.getElementById("foo");',
      'document.querySelector("#foo");',
      [['getElementById', 'querySelector', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementById("foo" as string);',
      'document.querySelector("#foo" as string);',
      [['getElementById', 'querySelector', 1, 10]],
      [{ allowWithVariables: true }],
      'src/virtual.ts',
    ),
    invalid(
      'document.getElementsByClassName("foo");',
      'document.querySelectorAll(".foo");',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName("foo"!);',
      'document.querySelectorAll(".foo"!);',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
      'src/virtual.ts',
    ),
    invalid(
      'document.getElementsByTagName("foo");',
      'document.querySelectorAll("foo");',
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByName("foo");',
      'document.querySelectorAll("[name=\'foo\']");',
      [['getElementsByName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByTagName(someTag);',
      null,
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByName(someName);',
      null,
      [['getElementsByName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName(null);',
      null,
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName(`foo`);',
      'document.querySelectorAll(`.foo`);',
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName(variable + "x");',
      null,
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName("foo" + fn());',
      null,
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName(foo + "bar");',
      null,
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementById(`x${someId}`);',
      null,
      [['getElementById', 'querySelector', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByClassName(`foo ${someClass}`);',
      null,
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'window.frame.getElementById("bar");',
      'window.frame.querySelector("#bar");',
      [['getElementById', 'querySelector', 1, 14]],
    ),
    invalid(
      'class Foo { #document; bar() { return window.#document.getElementById("bar"); } }',
      'class Foo { #document; bar() { return window.#document.querySelector("#bar"); } }',
      [['getElementById', 'querySelector', 1, 56]],
    ),
    invalid(
      'for (const element of document.getElementsByClassName("foo")[0]) {}',
      'for (const element of document.querySelectorAll(".foo")[0]) {}',
      [['getElementsByClassName', 'querySelectorAll', 1, 32]],
    ),
    // Upstream documentation examples.
    invalid(
      "document.getElementById('foo');",
      "document.querySelector('#foo');",
      [['getElementById', 'querySelector', 1, 10]],
    ),
    invalid(
      "document.getElementsByClassName('foo');",
      "document.querySelectorAll('.foo');",
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "document.getElementsByClassName('foo bar');",
      "document.querySelectorAll('.foo.bar');",
      [['getElementsByClassName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "document.getElementsByTagName('main');",
      "document.querySelectorAll('main');",
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "const singleClassName = 'foo';\ndocument.getElementsByClassName(singleClassName);",
      null,
      [['getElementsByClassName', 'querySelectorAll', 2, 10]],
    ),
    invalid(
      "document.getElementsByTagName('form')[0];",
      "document.querySelectorAll('form')[0];",
      [['getElementsByTagName', 'querySelectorAll', 1, 10]],
    ),
    invalid(
      "document.getElementById('foo');\ndocument.getElementsByClassName('foo');",
      "document.querySelector('#foo');\ndocument.querySelectorAll('.foo');",
      [
        ['getElementById', 'querySelector', 1, 10],
        ['getElementsByClassName', 'querySelectorAll', 2, 10],
      ],
      [{ allowWithVariables: true }],
    ),
    invalid(
      'document.getElementsByTagName(someTag);\ndocument.getElementsByName(someName);',
      null,
      [
        ['getElementsByTagName', 'querySelectorAll', 1, 10],
        ['getElementsByName', 'querySelectorAll', 2, 10],
      ],
      [{ allowWithVariables: true }],
    ),
  ],
};

new RuleTester().run('prefer-query-selector', null as never, cases);

// The suite wrapper checks messages; assert ranges and edits through IPC here.
test('upstream diagnostic ranges and autofixes', async () => {
  for (const item of cases.invalid) {
    const filename = path.resolve(import.meta.dirname, item.filename);
    const result = await lint({
      config: [
        {
          plugins: ['unicorn'],
          languageOptions: {
            parserOptions: { project: false, projectService: false },
          },
          rules: {
            'unicorn/prefer-query-selector': ['error', ...item.options],
          },
        },
      ],
      configDirectory: import.meta.dirname,
      workingDirectory: process.cwd(),
      fileContents: { [filename]: item.code },
    });
    assert.equal(result.fileCount, 1, item.code);
    assert.equal(result.ruleCount, 1, item.code);
    assert.equal(result.diagnostics.length, item.errors.length, item.code);
    for (const [index, diagnostic] of result.diagnostics.entries()) {
      const expected = item.errors[index];
      assert.equal(
        diagnostic.ruleName,
        'unicorn/prefer-query-selector',
        item.code,
      );
      assert.equal(diagnostic.messageId, expected.messageId, item.code);
      assert.equal(diagnostic.message, expected.message, item.code);
      assert.deepEqual(
        diagnostic.range,
        {
          start: { line: expected.line, column: expected.column },
          end: { line: expected.endLine, column: expected.endColumn },
        },
        item.code,
      );
      assert.equal(diagnostic.suggestions?.length ?? 0, 0, item.code);
    }
    const fixes = result.diagnostics
      .flatMap((diagnostic) => diagnostic.fixes ?? [])
      .sort((a, b) => b.startPos - a.startPos);
    let output = item.code;
    for (const fix of fixes)
      output =
        output.slice(0, fix.startPos) + fix.text + output.slice(fix.endPos);
    assert.equal(fixes.length ? output : null, item.output, item.code);
  }
});
