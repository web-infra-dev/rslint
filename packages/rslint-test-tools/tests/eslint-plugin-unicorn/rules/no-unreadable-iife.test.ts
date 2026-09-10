// Upstream tests and documentation from eslint-plugin-unicorn v74.0.0 (MIT).
// License: internal/plugins/unicorn/rules/no_unreadable_iife/LICENSE
import path from 'node:path';
import { lint } from '@rslint/core/internal';

const cases = {
  valid: [
    'const foo = (bar => bar)();',
    'const foo = (() => {\n\treturn a ? b : c\n})();',
    'const bar = getBar();\nconst foo = bar ? bar.baz : baz;',
    'const getBaz = bar => (bar ? bar.baz : baz);\nconst foo = getBaz(getBar());',
    'const foo = {bar, baz};',
    'const foo = (bar => {\n\treturn bar ? bar.baz : baz;\n})(getBar());',
  ],
  invalid: [
    {
      code: 'const foo = (() => (a ? b : c))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 20,
      endLine: 1,
      endColumn: 31,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: 'const foo = (() => { return a ? b : c; })();',
        },
      ],
    },
    {
      code: 'const foo = (() => (\n\ta ? b : c\n))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 20,
      endLine: 3,
      endColumn: 2,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: 'const foo = (() => { return a ? b : c; })();',
        },
      ],
    },
    {
      code: 'const foo = (\n\t() => (\n\t\ta ? b : c\n\t)\n)();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 2,
      column: 8,
      endLine: 4,
      endColumn: 3,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: 'const foo = (\n\t() => { return a ? b : c; }\n)();',
        },
      ],
    },
    {
      code: 'const foo = (() => (/* comment */ a ? b : c))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 20,
      endLine: 1,
      endColumn: 45,
      suggestions: [],
    },
    {
      code: 'const foo = (() => (\n\ta, b\n))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 20,
      endLine: 3,
      endColumn: 2,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: 'const foo = (() => { return a, b; })();',
        },
      ],
    },
    {
      code: 'const foo = (() => ({\n\ta: b,\n}))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 20,
      endLine: 3,
      endColumn: 3,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: 'const foo = (() => { return {\n\ta: b,\n}; })();',
        },
      ],
    },
    {
      code: 'const foo = (bar => (bar))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 21,
      endLine: 1,
      endColumn: 26,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: 'const foo = (bar => { return bar; })();',
        },
      ],
    },
    {
      code: '(async () => ({\n\tbar,\n}))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 14,
      endLine: 3,
      endColumn: 3,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: '(async () => { return {\n\tbar,\n}; })();',
        },
      ],
    },
    {
      code: 'const foo = (async (bar) => ({\n\tbar: await baz(),\n}))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 29,
      endLine: 3,
      endColumn: 3,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output:
            'const foo = (async (bar) => { return {\n\tbar: await baz(),\n}; })();',
        },
      ],
    },
    {
      code: '(async () => (( {bar} )))();',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 14,
      endLine: 1,
      endColumn: 25,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output: '(async () => { return {bar}; })();',
        },
      ],
    },
    {
      code: 'const foo = (bar => (bar ? bar.baz : baz))(getBar());',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 21,
      endLine: 1,
      endColumn: 42,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output:
            'const foo = (bar => { return bar ? bar.baz : baz; })(getBar());',
        },
      ],
    },
    {
      code: 'const foo = ((bar, baz) => ({bar, baz}))(bar, baz);',
      message:
        'IIFE with parenthesized arrow function body is considered unreadable.',
      messageId: 'no-unreadable-iife',
      line: 1,
      column: 28,
      endLine: 1,
      endColumn: 40,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use a block statement body.',
          output:
            'const foo = ((bar, baz) => { return {bar, baz}; })(bar, baz);',
        },
      ],
    },
  ],
};

const directory = path.resolve(import.meta.dirname, '..');
const filename = path.join(directory, 'src/virtual.tsx');
const run = (code: string, fix = false) =>
  lint({
    config: [
      {
        plugins: ['unicorn'],
        rules: { 'unicorn/no-unreadable-iife': 'error' },
      },
    ],
    configDirectory: directory,
    workingDirectory: directory,
    fileContents: { [filename]: code },
    fix,
  });

describe('unicorn/no-unreadable-iife', () => {
  test('upstream valid cases', async () => {
    for (const code of cases.valid) {
      const result = await run(code);
      expect(result.fileCount).toBe(1);
      expect(result.diagnostics).toEqual([]);
    }
  });
  test('upstream diagnostics and suggestions', async () => {
    for (const item of cases.invalid) {
      const result = await run(item.code);
      const autofix = await run(item.code, true);
      expect(autofix.output ?? {}).toEqual({});
      expect(autofix.fixableErrorCount).toBe(0);
      expect(result.diagnostics).toHaveLength(1);
      const [diagnostic] = result.diagnostics;
      expect(diagnostic.ruleName).toBe('unicorn/no-unreadable-iife');
      expect(diagnostic.messageId).toBe(item.messageId);
      expect(diagnostic.message).toBe(item.message);
      expect(diagnostic.range).toEqual({
        start: { line: item.line, column: item.column },
        end: { line: item.endLine, column: item.endColumn },
      });
      expect(diagnostic.fixes ?? []).toEqual([]);
      const suggestions = diagnostic.suggestions ?? [];
      expect(suggestions).toHaveLength(item.suggestions.length);
      for (const [index, expected] of item.suggestions.entries()) {
        const suggestion = suggestions[index];
        expect(suggestion.messageId).toBe(expected.messageId);
        expect(suggestion.message).toBe(expected.message);
        expect(suggestion.fixes).toHaveLength(1);
        const [edit] = suggestion.fixes ?? [];
        const output =
          item.code.slice(0, edit.startPos) +
          edit.text +
          item.code.slice(edit.endPos);
        expect(output).toBe(expected.output);
        expect((await run(output)).diagnostics).toEqual([]);
      }
    }
  });
});
