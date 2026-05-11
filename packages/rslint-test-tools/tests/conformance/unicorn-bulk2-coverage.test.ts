import { describe, test, expect } from '@rstest/core';
import { pathToFileURL } from 'node:url';
import path from 'node:path';

import unicornPlugin from 'eslint-plugin-unicorn';
import {
  runConformance,
  formatReport,
  type ConformanceFixture,
} from '../../src/eslint-conformance.js';

/**
 * Second bulk unicorn coverage round — 30 more rules across DOM
 * helpers, modern-method preferences, array-method aliases, and
 * miscellaneous style. Pushes the matrix toward 75% / 80% of the
 * plugin's 149 rules.
 */

describe('unicorn bulk2-coverage conformance', () => {
  test('eslint and rslint agree on a second batch of unicorn rules', async () => {
    const baseUrl = pathToFileURL(
      path.resolve(__dirname, '..', '..', 'package.json'),
    ).href;

    const fixtures: ConformanceFixture[] = [
      {
        filePath: 'consistent-assert.js',
        text: `import { strict as assert } from 'assert';\nassert.equal(1, 1);`,
        rules: { 'unicorn/consistent-assert': 'error' },
      },
      {
        filePath: 'consistent-date-clone.js',
        text: `const d = new Date(); const c = new Date(d.getTime());\nvoid c;`,
        rules: { 'unicorn/consistent-date-clone': 'error' },
      },
      {
        filePath: 'consistent-destructuring.js',
        text: `const obj = { a: 1, b: 2 }; const a = obj.a; const b = obj.b; void a; void b;`,
        rules: { 'unicorn/consistent-destructuring': 'error' },
      },
      {
        filePath: 'consistent-existence-index-check.js',
        text: `if ([1,2,3].indexOf(2) !== -1) {}`,
        rules: { 'unicorn/consistent-existence-index-check': 'error' },
      },
      {
        filePath: 'consistent-template-literal-escape.js',
        text: 'const x = `a\\u0042b`;\nvoid x;',
        rules: { 'unicorn/consistent-template-literal-escape': 'error' },
      },
      {
        filePath: 'custom-error-definition.js',
        text: `class FooError extends Error { constructor(msg) { super(msg); } }\nvoid FooError;`,
        rules: { 'unicorn/custom-error-definition': 'error' },
      },
      {
        filePath: 'empty-brace-spaces.js',
        text: `function f() { }`,
        rules: { 'unicorn/empty-brace-spaces': 'error' },
      },
      {
        filePath: 'expiring-todo-comments.js',
        text: `// TODO [2020-01-01]: fixme`,
        rules: { 'unicorn/expiring-todo-comments': 'error' },
      },
      {
        filePath: 'no-accessor-recursion.js',
        text: `class C { get x() { return this.x; } }\nvoid C;`,
        rules: { 'unicorn/no-accessor-recursion': 'error' },
      },
      {
        filePath: 'no-array-reverse.js',
        text: `const a = [1,2,3]; a.reverse();`,
        rules: { 'unicorn/no-array-reverse': 'error' },
      },
      {
        filePath: 'no-array-sort.js',
        text: `const a = [1,2,3]; a.sort();`,
        rules: { 'unicorn/no-array-sort': 'error' },
      },
      {
        filePath: 'no-immediate-mutation.js',
        text: `const a = [1,2,3].sort();\nvoid a;`,
        rules: { 'unicorn/no-immediate-mutation': 'error' },
      },
      {
        filePath: 'no-length-as-slice-end.js',
        text: `[1,2,3].slice(0, [1,2,3].length);`,
        rules: { 'unicorn/no-length-as-slice-end': 'error' },
      },
      {
        filePath: 'no-named-default.js',
        text: `import { default as X } from "./mod"; void X;`,
        rules: { 'unicorn/no-named-default': 'error' },
      },
      {
        filePath: 'no-thenable.js',
        text: `const o = { then: () => 1 };\nvoid o;`,
        rules: { 'unicorn/no-thenable': 'error' },
      },
      {
        filePath: 'no-unnecessary-array-splice-count.js',
        text: `[1,2,3].splice(0, 0);`,
        rules: { 'unicorn/no-unnecessary-array-splice-count': 'error' },
      },
      {
        filePath: 'no-unnecessary-slice-end.js',
        text: `[1,2,3].slice(0, undefined);`,
        rules: { 'unicorn/no-unnecessary-slice-end': 'error' },
      },
      {
        filePath: 'no-unreadable-array-destructuring.js',
        text: `const [,,a] = [1,2,3]; void a;`,
        rules: { 'unicorn/no-unreadable-array-destructuring': 'error' },
      },
      {
        filePath: 'no-useless-error-capture-stack-trace.js',
        text: `class FooError extends Error { constructor(msg) { super(msg); Error.captureStackTrace(this, FooError); } }\nvoid FooError;`,
        rules: { 'unicorn/no-useless-error-capture-stack-trace': 'error' },
      },
      {
        filePath: 'no-useless-iterator-to-array.js',
        text: `const a = [...[1,2,3]]; void a;`,
        rules: { 'unicorn/no-useless-iterator-to-array': 'error' },
      },
      {
        filePath: 'no-useless-length-check.js',
        text: `const a = [1,2,3]; if (a.length === 0 || a.length > 0) {}`,
        rules: { 'unicorn/no-useless-length-check': 'error' },
      },
      {
        filePath: 'prefer-array-flat-map.js',
        text: `[1,2,3].map(x => [x]).flat();`,
        rules: { 'unicorn/prefer-array-flat-map': 'error' },
      },
      {
        filePath: 'prefer-at.js',
        text: `const a = [1,2,3]; const last = a[a.length - 1]; void last;`,
        rules: { 'unicorn/prefer-at': 'error' },
      },
      {
        filePath: 'prefer-blob-reading-methods.js',
        text: `const b = new Blob(); const fr = new FileReader(); fr.readAsArrayBuffer(b);`,
        rules: { 'unicorn/prefer-blob-reading-methods': 'error' },
      },
      {
        filePath: 'prefer-code-point.js',
        text: `'a'.charCodeAt(0);`,
        rules: { 'unicorn/prefer-code-point': 'error' },
      },
      {
        filePath: 'prefer-dom-node-append.js',
        text: `const p = document.createElement('p'); document.body.appendChild(p);`,
        rules: { 'unicorn/prefer-dom-node-append': 'error' },
      },
      {
        filePath: 'prefer-dom-node-dataset.js',
        text: `const el = document.body; el.setAttribute('data-x', '1');`,
        rules: { 'unicorn/prefer-dom-node-dataset': 'error' },
      },
      {
        filePath: 'prefer-dom-node-remove.js',
        text: `const el = document.body; el.parentNode.removeChild(el);`,
        rules: { 'unicorn/prefer-dom-node-remove': 'error' },
      },
      {
        filePath: 'prefer-dom-node-text-content.js',
        text: `const el = document.body; el.innerText = 'x';`,
        rules: { 'unicorn/prefer-dom-node-text-content': 'error' },
      },
      {
        filePath: 'prefer-event-target.js',
        text: `class C extends EventEmitter { }\nvoid C;`,
        rules: { 'unicorn/prefer-event-target': 'error' },
      },
      {
        filePath: 'prefer-global-this.js',
        text: `window.alert('x');`,
        rules: { 'unicorn/prefer-global-this': 'error' },
      },
      {
        filePath: 'prefer-json-parse-buffer.js',
        text: `import fs from 'fs'; const x = JSON.parse(fs.readFileSync('a.json', 'utf8'));\nvoid x;`,
        rules: { 'unicorn/prefer-json-parse-buffer': 'error' },
      },
      {
        filePath: 'prefer-keyboard-event-key.js',
        text: `document.addEventListener('keydown', e => e.keyCode);`,
        rules: { 'unicorn/prefer-keyboard-event-key': 'error' },
      },
      {
        filePath: 'prefer-modern-dom-apis.js',
        text: `const a = document.body; const b = document.createElement('p'); a.parentNode.replaceChild(b, a);`,
        rules: { 'unicorn/prefer-modern-dom-apis': 'error' },
      },
      {
        filePath: 'prefer-native-coercion-functions.js',
        text: `const f = x => Boolean(x);\nvoid f;`,
        rules: { 'unicorn/prefer-native-coercion-functions': 'error' },
      },
      {
        filePath: 'prefer-node-protocol.js',
        text: `import fs from 'fs';\nvoid fs;`,
        rules: { 'unicorn/prefer-node-protocol': 'error' },
      },
      {
        filePath: 'prefer-object-from-entries.js',
        text: `const o = [['a',1]].reduce((acc, [k,v]) => (acc[k] = v, acc), {}); void o;`,
        rules: { 'unicorn/prefer-object-from-entries': 'error' },
      },
      {
        filePath: 'prefer-query-selector.js',
        text: `document.getElementById('x');`,
        rules: { 'unicorn/prefer-query-selector': 'error' },
      },
      {
        filePath: 'prefer-reflect-apply.js',
        text: `function f() {} f.apply(null, [1,2,3]);`,
        rules: { 'unicorn/prefer-reflect-apply': 'error' },
      },
      {
        filePath: 'prefer-string-raw.js',
        text: 'const a = "a\\\\b"; void a;',
        rules: { 'unicorn/prefer-string-raw': 'error' },
      },
      {
        filePath: 'prefer-string-replace-all.js',
        text: `'aaa'.replace(/a/g, 'b');`,
        rules: { 'unicorn/prefer-string-replace-all': 'error' },
      },
      {
        filePath: 'prefer-structured-clone.js',
        text: `const a = JSON.parse(JSON.stringify({x:1}));\nvoid a;`,
        rules: { 'unicorn/prefer-structured-clone': 'error' },
      },
      {
        filePath: 'prefer-type-error.js',
        text: `function f(x) { if (typeof x !== 'number') { throw new Error('x must be a number'); } }\nvoid f;`,
        rules: { 'unicorn/prefer-type-error': 'error' },
      },
      {
        filePath: 'relative-url-style.js',
        text: `const u = new URL('./other.js', import.meta.url);\nvoid u;`,
        rules: { 'unicorn/relative-url-style': 'error' },
      },
      {
        filePath: 'string-content.js',
        text: `const x = "test'\\u2019Hello";\nvoid x;`,
        rules: { 'unicorn/string-content': 'error' },
      },
      {
        filePath: 'template-indent.js',
        text: 'const x = `\n  line\n`;\nvoid x;',
        rules: { 'unicorn/template-indent': 'error' },
      },
    ];

    const ruleNames = Array.from(
      new Set(fixtures.flatMap((f) => Object.keys(f.rules))),
    )
      .map((r) => r.replace(/^unicorn\//, ''))
      .sort();

    const report = await runConformance({
      plugin: {
        prefix: 'unicorn',
        plugin: unicornPlugin as never,
        specifier: 'eslint-plugin-unicorn',
        ruleNames,
      },
      fixtures,
      resolverBaseUrl: baseUrl,
      workerCount: 1,
    });

    if (report.mismatched > 0) {
      throw new Error('Conformance mismatch:\n' + formatReport(report));
    }
    expect(report.mismatched).toBe(0);
    // Vacuous-pass guard — see unicorn-extended-coverage for rationale.
    expect(
      report.fixtureResults.reduce((n, r) => n + r.eslint.length, 0),
    ).toBeGreaterThan(0);
  });
});
