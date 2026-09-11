// Upstream tests and documentation from eslint-plugin-unicorn v74.0.0 (MIT).
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/test/require-post-message-target-origin.js
import path from 'node:path';
import { lint } from '@rslint/core/internal';

const cases = {
  valid: [
    'window.postMessage(message, targetOrigin)',
    'postMessage(message)',
    'window.postMessage',
    'window.postMessage()',
    'window.postMessage(message, targetOrigin, transfer)',
    'window.postMessage(...message)',
    'window[postMessage](message)',
    'window["postMessage"](message)',
    'window.notPostMessage(message)',
    'window.postMessage?.(message)',
    "window.postMessage(sensitiveData, 'https://trusted-domain.com');",
    "window.postMessage({token: authToken}, 'https://api.example.com');",
    "window.postMessage({publicData: 'hello'}, '*');",
    "iframe.contentWindow.postMessage(data, 'https://expected-iframe-origin.com');",
  ],
  invalid: [
    {
      code: 'window.postMessage(message)',
      column: 27,
      line: 1,
      endLine: 1,
      endColumn: 28,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `window.location.origin`.',
          output: 'window.postMessage(message, window.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "window.postMessage(message, '*')",
        },
      ],
    },
    {
      code: 'self.postMessage(message)',
      column: 25,
      line: 1,
      endLine: 1,
      endColumn: 26,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'self.postMessage(message, self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "self.postMessage(message, '*')",
        },
      ],
    },
    {
      code: 'globalThis.postMessage(message)',
      column: 31,
      line: 1,
      endLine: 1,
      endColumn: 32,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `globalThis.location.origin`.',
          output: 'globalThis.postMessage(message, globalThis.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "globalThis.postMessage(message, '*')",
        },
      ],
    },
    {
      code: 'foo.postMessage(message )',
      column: 24,
      line: 1,
      endLine: 1,
      endColumn: 26,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `foo.location.origin`.',
          output: 'foo.postMessage(message , foo.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'foo.postMessage(message , self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "foo.postMessage(message , '*')",
        },
      ],
    },
    {
      code: 'foo?.postMessage(message )',
      column: 25,
      line: 1,
      endLine: 1,
      endColumn: 27,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `foo.location.origin`.',
          output: 'foo?.postMessage(message , foo.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'foo?.postMessage(message , self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "foo?.postMessage(message , '*')",
        },
      ],
    },
    {
      code: 'foo.postMessage( ((message)) )',
      column: 29,
      line: 1,
      endLine: 1,
      endColumn: 31,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `foo.location.origin`.',
          output: 'foo.postMessage( ((message)) , foo.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'foo.postMessage( ((message)) , self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "foo.postMessage( ((message)) , '*')",
        },
      ],
    },
    {
      code: 'foo.postMessage(message,)',
      column: 25,
      line: 1,
      endLine: 1,
      endColumn: 26,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `foo.location.origin`.',
          output: 'foo.postMessage(message, foo.location.origin,)',
        },
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'foo.postMessage(message, self.location.origin,)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "foo.postMessage(message, '*',)",
        },
      ],
    },
    {
      code: 'foo.postMessage(message , )',
      column: 26,
      line: 1,
      endLine: 1,
      endColumn: 28,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `foo.location.origin`.',
          output: 'foo.postMessage(message ,  foo.location.origin,)',
        },
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'foo.postMessage(message ,  self.location.origin,)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "foo.postMessage(message ,  '*',)",
        },
      ],
    },
    {
      code: 'foo.window.postMessage(message)',
      column: 31,
      line: 1,
      endLine: 1,
      endColumn: 32,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'foo.window.postMessage(message, self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "foo.window.postMessage(message, '*')",
        },
      ],
    },
    {
      code: 'document.defaultView.postMessage(message)',
      column: 41,
      line: 1,
      endLine: 1,
      endColumn: 42,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output:
            'document.defaultView.postMessage(message, self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "document.defaultView.postMessage(message, '*')",
        },
      ],
    },
    {
      code: 'getWindow().postMessage(message)',
      column: 32,
      line: 1,
      endLine: 1,
      endColumn: 33,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output: 'getWindow().postMessage(message, self.location.origin)',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "getWindow().postMessage(message, '*')",
        },
      ],
    },
    {
      code: 'window.postMessage(sensitiveData);',
      column: 33,
      line: 1,
      endLine: 1,
      endColumn: 34,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `window.location.origin`.',
          output: 'window.postMessage(sensitiveData, window.location.origin);',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "window.postMessage(sensitiveData, '*');",
        },
      ],
    },
    {
      code: 'window.postMessage({token: authToken});',
      column: 38,
      line: 1,
      endLine: 1,
      endColumn: 39,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `window.location.origin`.',
          output:
            'window.postMessage({token: authToken}, window.location.origin);',
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output: "window.postMessage({token: authToken}, '*');",
        },
      ],
    },
    {
      code: "const iframe = document.querySelector('iframe');\niframe.contentWindow.postMessage(data);",
      column: 38,
      line: 2,
      endLine: 2,
      endColumn: 39,
      suggestions: [
        {
          messageId: 'suggestion',
          message: 'Use `self.location.origin`.',
          output:
            "const iframe = document.querySelector('iframe');\niframe.contentWindow.postMessage(data, self.location.origin);",
        },
        {
          messageId: 'suggestion',
          message: "Use `'*'`.",
          output:
            "const iframe = document.querySelector('iframe');\niframe.contentWindow.postMessage(data, '*');",
        },
      ],
    },
  ],
};

const directory = path.resolve(import.meta.dirname, '..');
const filename = path.join(directory, 'src/virtual.js');
const ruleName = 'unicorn/require-post-message-target-origin';
const run = (code: string, fix = false) =>
  lint({
    config: [{ plugins: ['unicorn'], rules: { [ruleName]: 'error' } }],
    configDirectory: directory,
    workingDirectory: directory,
    fileContents: { [filename]: code },
    fix,
  });

describe(ruleName, () => {
  test('upstream valid cases and documentation', async () => {
    for (const code of cases.valid) {
      const result = await run(code);
      expect(result.fileCount).toBe(1);
      expect(result.diagnostics).toEqual([]);
    }
  });
  test('upstream diagnostics and suggestions', async () => {
    for (const item of cases.invalid) {
      const result = await run(item.code);
      expect(result.fileCount).toBe(1);
      expect(result.diagnostics).toHaveLength(1);
      const [diagnostic] = result.diagnostics;
      expect(diagnostic.ruleName).toBe(ruleName);
      expect(diagnostic.messageId).toBe('error');
      expect(diagnostic.message).toBe('Missing the `targetOrigin` argument.');
      expect(diagnostic.range).toEqual({
        start: { line: item.line, column: item.column },
        end: { line: item.endLine, column: item.endColumn },
      });
      expect(diagnostic.fixes ?? []).toEqual([]);
      const autofix = await run(item.code, true);
      expect(autofix.output ?? {}).toEqual({});
      expect(autofix.fixableErrorCount).toBe(0);
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
