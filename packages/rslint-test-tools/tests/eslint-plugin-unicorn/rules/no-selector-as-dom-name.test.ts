// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-selector-as-dom-name.js
import assert from 'node:assert/strict';
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { test } from 'rstack/test';

function invalid(
  code: string,
  ranges: [number, number, number, number][],
  output: string | null,
  filename = 'file.js',
): TestCase {
  return {
    code,
    filename,
    ...(output === null ? {} : { output }),
    errors: ranges.map(([line, column, endLine, endColumn]) => ({
      messageId: 'no-selector-as-dom-name',
      message: 'Do not use selector syntax in DOM names.',
      line,
      column,
      endLine,
      endColumn,
    })),
  };
}

const tester = new RuleTester({ plugin: 'unicorn' });
tester.run('no-selector-as-dom-name', {
  valid: [
    'element.classList.add("foo");',
    'element.classList.remove("foo");',
    'element.classList.contains("foo");',
    'element.classList.toggle("foo");',
    'element.classList.replace("old", "new");',
    'element.classList.toggle("foo", ".bar");',
    'element.classList.replace("old", "new", ".extra");',
    'element.classList.add("foo.bar");',
    'element.classList.add(className);',
    'element.classList.add(`foo`);',
    'element.classList.add(`${className}`);',
    'element.classList.add(...classNames);',
    'element.classList.replace(...tokens, ".new");',
    { code: '("string" as any).classList.add(".foo");', filename: 'file.ts' },
    '([]).classList.add(".foo");',
    '([element]).classList.add(".foo");',
    '([...elements]).classList.add(".foo");',
    '(() => {}).classList.add(".foo");',
    '(class Node {}).classList.add(".foo");',
    '(function() {}).classList.add(".foo");',
    '(0).classList.add(".foo");',
    '(1).classList.add(".foo");',
    '(0.1).classList.add(".foo");',
    '("").classList.add(".foo");',
    '("string").classList.add(".foo");',
    '(/regex/).classList.add(".foo");',
    '(null).classList.add(".foo");',
    '(0n).classList.add(".foo");',
    '(1n).classList.add(".foo");',
    '(true).classList.add(".foo");',
    '(false).classList.add(".foo");',
    '({}).classList.add(".foo");',
    '(`templateLiteral`).classList.add(".foo");',
    '(undefined).classList.add(".foo");',
    'element.classList["add"](".foo");',
    'element.classList[method](".foo");',
    'element.notClassList.add(".foo");',
    'classList.add(".foo");',
    'document.getElementsByClassName("foo");',
    'element.getElementsByClassName("foo");',
    'document.getElementsByClassName(className);',
    'document.getElementsByClassName(`foo`);',
    'document.getElementsByClassName(`${className}`);',
    { code: 'element.classList.add("foo" as string);', filename: 'file.ts' },
    {
      code: '(undefined as any).getElementsByClassName(".foo");',
      filename: 'file.ts',
    },
    '([]).getElementsByClassName(".foo");',
    '([element]).getElementsByClassName(".foo");',
    '([...elements]).getElementsByClassName(".foo");',
    '(() => {}).getElementsByClassName(".foo");',
    '(class Node {}).getElementsByClassName(".foo");',
    '(function() {}).getElementsByClassName(".foo");',
    '(0).getElementsByClassName(".foo");',
    '(1).getElementsByClassName(".foo");',
    '(0.1).getElementsByClassName(".foo");',
    '("").getElementsByClassName(".foo");',
    '("string").getElementsByClassName(".foo");',
    '(/regex/).getElementsByClassName(".foo");',
    '(null).getElementsByClassName(".foo");',
    '(0n).getElementsByClassName(".foo");',
    '(1n).getElementsByClassName(".foo");',
    '(true).getElementsByClassName(".foo");',
    '(false).getElementsByClassName(".foo");',
    '({}).getElementsByClassName(".foo");',
    '(`templateLiteral`).getElementsByClassName(".foo");',
    '(undefined).getElementsByClassName(".foo");',
    'document.getElementById("foo");',
    'document.getElementById(id);',
    'document.getElementById(`foo`);',
    'document.getElementById(`${id}`);',
    '([]).getElementById("#foo");',
    '([element]).getElementById("#foo");',
    '([...elements]).getElementById("#foo");',
    '(() => {}).getElementById("#foo");',
    '(class Node {}).getElementById("#foo");',
    '(function() {}).getElementById("#foo");',
    '(0).getElementById("#foo");',
    '(1).getElementById("#foo");',
    '(0.1).getElementById("#foo");',
    '("").getElementById("#foo");',
    '("string").getElementById("#foo");',
    '(/regex/).getElementById("#foo");',
    '(null).getElementById("#foo");',
    '(0n).getElementById("#foo");',
    '(1n).getElementById("#foo");',
    '(true).getElementById("#foo");',
    '(false).getElementById("#foo");',
    '({}).getElementById("#foo");',
    '(`templateLiteral`).getElementById("#foo");',
    '(undefined).getElementById("#foo");',
    'document.querySelector("#foo");',
    'element.setAttribute("class", ".foo");',
    'element.setAttribute("id", "#foo");',
    "element.classList.add('active');\nelement.classList.remove('hidden');\nelement.classList.contains('selected');\nelement.classList.toggle('expanded');\nelement.classList.replace('old', 'new');\ndocument.getElementsByClassName('item');\ndocument.getElementById('app');",
  ],
  invalid: [
    invalid(
      'element.classList.add(".foo");',
      [[1, 23, 1, 29]],
      'element.classList.add("foo");',
    ),
    invalid(
      'element.classList.add((".foo"));',
      [[1, 24, 1, 30]],
      'element.classList.add(("foo"));',
    ),
    invalid(
      "element.classList.add('.foo');",
      [[1, 23, 1, 29]],
      "element.classList.add('foo');",
    ),
    invalid('element.classList.add("#foo");', [[1, 23, 1, 29]], null),
    invalid(
      'element?.classList.add(".foo");',
      [[1, 24, 1, 30]],
      'element?.classList.add("foo");',
    ),
    invalid(
      'element.classList?.add(".foo");',
      [[1, 24, 1, 30]],
      'element.classList?.add("foo");',
    ),
    invalid(
      'element.classList.add?.(".foo");',
      [[1, 25, 1, 31]],
      'element.classList.add?.("foo");',
    ),
    invalid(
      'element.classList.remove(".foo");',
      [[1, 26, 1, 32]],
      'element.classList.remove("foo");',
    ),
    invalid(
      'element.classList.contains(".foo");',
      [[1, 28, 1, 34]],
      'element.classList.contains("foo");',
    ),
    invalid(
      'element.classList.toggle(".foo");',
      [[1, 26, 1, 32]],
      'element.classList.toggle("foo");',
    ),
    invalid(
      'element.classList.replace(".old", "new");',
      [[1, 27, 1, 33]],
      'element.classList.replace("old", "new");',
    ),
    invalid(
      'element.classList.replace("old", ".new");',
      [[1, 34, 1, 40]],
      'element.classList.replace("old", "new");',
    ),
    invalid(
      'element.classList.add(".foo", ".bar");',
      [
        [1, 23, 1, 29],
        [1, 31, 1, 37],
      ],
      'element.classList.add("foo", "bar");',
    ),
    invalid(
      'element.classList.add(`.foo`);',
      [[1, 23, 1, 29]],
      'element.classList.add(`foo`);',
    ),
    invalid(
      'element.classList.add(`.${className}`);',
      [[1, 23, 1, 38]],
      'element.classList.add(`${className}`);',
    ),
    invalid('element.classList.add(`#${className}`);', [[1, 23, 1, 38]], null),
    invalid(
      'element.classList.add(`.${className}.bar`);',
      [[1, 23, 1, 42]],
      null,
    ),
    invalid('element.classList.add(`.foo${suffix}`);', [[1, 23, 1, 38]], null),
    invalid(
      'element.classList.add(".foo" as string);',
      [[1, 23, 1, 29]],
      'element.classList.add("foo" as string);',
      'file.ts',
    ),
    invalid(
      'element.classList.replace("old", ".new" as string);',
      [[1, 34, 1, 40]],
      'element.classList.replace("old", "new" as string);',
      'file.ts',
    ),
    invalid(
      'element.classList.add(".foo"!);',
      [[1, 23, 1, 29]],
      'element.classList.add("foo"!);',
      'file.ts',
    ),
    invalid(
      'element.classList.add(".foo" satisfies string);',
      [[1, 23, 1, 29]],
      'element.classList.add("foo" satisfies string);',
      'file.ts',
    ),
    invalid(
      'document.getElementsByClassName(<string>".foo");',
      [[1, 41, 1, 47]],
      'document.getElementsByClassName(<string>"foo");',
      'file.ts',
    ),
    invalid(
      'document.getElementsByClassName(".foo");',
      [[1, 33, 1, 39]],
      'document.getElementsByClassName("foo");',
    ),
    invalid('document.getElementsByClassName("#foo");', [[1, 33, 1, 39]], null),
    invalid(
      'document.getElementsByClassName(`.${className}`);',
      [[1, 33, 1, 48]],
      'document.getElementsByClassName(`${className}`);',
    ),
    invalid(
      'document.getElementsByClassName(`#${className}`);',
      [[1, 33, 1, 48]],
      null,
    ),
    invalid(
      'document?.getElementsByClassName(".foo");',
      [[1, 34, 1, 40]],
      'document?.getElementsByClassName("foo");',
    ),
    invalid(
      'document.getElementsByClassName?.(".foo");',
      [[1, 35, 1, 41]],
      'document.getElementsByClassName?.("foo");',
    ),
    invalid(
      'element.getElementsByClassName(".foo");',
      [[1, 32, 1, 38]],
      'element.getElementsByClassName("foo");',
    ),
    invalid('element.classList.add("\\u002efoo");', [[1, 23, 1, 34]], null),
    invalid('element.classList.add(`\\u002efoo`);', [[1, 23, 1, 34]], null),
    invalid('element.classList.add(".foo.bar");', [[1, 23, 1, 33]], null),
    invalid('element.classList.add(".foo .bar");', [[1, 23, 1, 34]], null),
    invalid('element.classList.add(".foo, .bar");', [[1, 23, 1, 35]], null),
    invalid('element.classList.add(".foo:hover");', [[1, 23, 1, 35]], null),
    invalid('element.classList.add(".123");', [[1, 23, 1, 29]], null),
    invalid(
      'element.classList.add(".foo", ".bar.baz");',
      [
        [1, 23, 1, 29],
        [1, 31, 1, 41],
      ],
      'element.classList.add("foo", ".bar.baz");',
    ),
    invalid(
      'document.getElementsByClassName(".foo.bar");',
      [[1, 33, 1, 43]],
      null,
    ),
    invalid(
      'document.getElementsByClassName(".foo .bar");',
      [[1, 33, 1, 44]],
      null,
    ),
    invalid(
      'document.getElementById("#foo");',
      [[1, 25, 1, 31]],
      'document.getElementById("foo");',
    ),
    invalid(
      'document.getElementById(("#foo"));',
      [[1, 26, 1, 32]],
      'document.getElementById(("foo"));',
    ),
    invalid('document.getElementById(".foo");', [[1, 25, 1, 31]], null),
    invalid(
      "document.getElementById('#foo');",
      [[1, 25, 1, 31]],
      "document.getElementById('foo');",
    ),
    invalid(
      'document?.getElementById("#foo");',
      [[1, 26, 1, 32]],
      'document?.getElementById("foo");',
    ),
    invalid(
      'document.getElementById?.("#foo");',
      [[1, 27, 1, 33]],
      'document.getElementById?.("foo");',
    ),
    invalid(
      'document.getElementById(`#foo`);',
      [[1, 25, 1, 31]],
      'document.getElementById(`foo`);',
    ),
    invalid(
      'document.getElementById(`#${id}`);',
      [[1, 25, 1, 33]],
      'document.getElementById(`${id}`);',
    ),
    invalid('document.getElementById(`.${id}`);', [[1, 25, 1, 33]], null),
    invalid('document.getElementById(`#${id}:hover`);', [[1, 25, 1, 39]], null),
    invalid(
      'document.getElementById("#foo" as string);',
      [[1, 25, 1, 31]],
      'document.getElementById("foo" as string);',
      'file.ts',
    ),
    invalid('document.getElementById("\\u0023foo");', [[1, 25, 1, 36]], null),
    invalid('document.getElementById(`\\u0023foo`);', [[1, 25, 1, 36]], null),
    invalid('document.getElementById("#foo.bar");', [[1, 25, 1, 35]], null),
    invalid('document.getElementById("#foo .bar");', [[1, 25, 1, 36]], null),
    invalid('document.getElementById("#foo, #bar");', [[1, 25, 1, 37]], null),
    invalid('document.getElementById("#foo:hover");', [[1, 25, 1, 37]], null),
    invalid('document.getElementById("#123");', [[1, 25, 1, 31]], null),
    invalid(
      "element.classList.add('.active');\nelement.classList.remove('.hidden');\nelement.classList.contains('.selected');\nelement.classList.toggle('.expanded');\nelement.classList.replace('.old', '.new');\ndocument.getElementsByClassName('.item');\ndocument.getElementById('#app');",
      [
        [1, 23, 1, 32],
        [2, 26, 2, 35],
        [3, 28, 3, 39],
        [4, 26, 4, 37],
        [5, 27, 5, 33],
        [5, 35, 5, 41],
        [6, 33, 6, 40],
        [7, 25, 7, 31],
      ],
      "element.classList.add('active');\nelement.classList.remove('hidden');\nelement.classList.contains('selected');\nelement.classList.toggle('expanded');\nelement.classList.replace('old', 'new');\ndocument.getElementsByClassName('item');\ndocument.getElementById('app');",
    ),
  ],
});

const interactions = [
  [
    'file.js',
    'document.getElementsByClassName(".foo");',
    'document.querySelectorAll(".foo");',
  ],
  [
    'file.js',
    'document.getElementsByClassName(".foo")[0];',
    'document.querySelector(".foo");',
  ],
  // Rslint's prefer-query-selector omits :scope from generated selectors.
  [
    'file.js',
    'element.getElementsByClassName(".foo");',
    'element.querySelectorAll(".foo");',
  ],
  [
    'file.js',
    'document.getElementById("#foo");',
    'document.querySelector("#foo");',
  ],
  [
    'file.ts',
    'document.getElementsByClassName(".foo" as string);',
    'document.querySelectorAll(".foo" as string);',
  ],
  [
    'file.ts',
    'document.getElementById("#foo" as string);',
    'document.querySelector("#foo" as string);',
  ],
] as const;

for (const [filename, code, output] of interactions) {
  test('fixes DOM names before prefer-query-selector: ' + code, async () => {
    const root = import.meta.dirname;
    const result = await lint({
      workingDirectory: root,
      configDirectory: root,
      config: [
        {
          plugins: ['unicorn'],
          languageOptions: { parserOptions: { projectService: false } },
          rules: {
            'unicorn/no-selector-as-dom-name': 'error',
            'unicorn/prefer-query-selector': 'error',
          },
        },
      ],
      fileContents: { [path.join(root, filename)]: code },
      fix: true,
    });
    assert.equal(result.fileCount, 1);
    assert.equal(result.ruleCount, 2);
    assert.deepEqual(result.output, { [filename]: output });
    assert.deepEqual(result.diagnostics, []);
  });
}
