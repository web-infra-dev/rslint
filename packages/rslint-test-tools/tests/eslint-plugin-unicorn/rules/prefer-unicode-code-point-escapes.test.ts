// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/prefer-unicode-code-point-escapes.js
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import { RuleTester } from '../rule-tester';
import { buildConfigForSettings } from '../../src/util/load-test-config';

const groups = [
  {
    label: 'Upstream',
    valid: [
      {
        code: "const foo = '\\u{7A}'",
        output: null,
        errors: [],
      },
      {
        code: "const foo = '\\u{1F4A9}'",
        output: null,
        errors: [],
      },
      {
        code: "const foo = '\\n\\t\\r\\\\\\'\\\"'",
        output: null,
        errors: [],
      },
      {
        code: "const foo = '\\0'",
        output: null,
        errors: [],
      },
      {
        code: "const foo = '\\8\\9\\08'",
        skip: true,
        languageOptions: {
          sourceType: 'script' as const,
        },
        output: null,
        errors: [],
      },
      {
        code: "const foo = '\\\\u2661'",
        output: null,
        errors: [],
      },
      {
        code: "const foo = '\\\\x7A'",
        output: null,
        errors: [],
      },
      {
        code: 'const foo = tag`\\u2661`',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = tag`\\123`',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = String.raw`\\u2661`',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{61}/u',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{61}/v',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /[\\uD83D\\uDCA9]/u',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /[\\uD83D\\uDCA9]/v',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /[[\\uD83D\\uDCA9]\\uD83D\\uDCA9]/v',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{XYZ}/',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{110000}/',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{}/',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{/',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\cK/',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = new RegExp("\\\\u0061")',
        output: null,
        errors: [],
      },
    ],
    invalid: [
      {
        code: "const foo = '\\x7A'",
        output: "const foo = '\\u{7A}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = "\\x7A"',
        output: 'const foo = "\\u{7A}"',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\xa9'",
        output: "const foo = '\\u{A9}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\u2661'",
        output: "const foo = '\\u{2661}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\uD83D\\uDCA9'",
        output: "const foo = '\\u{1F4A9}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 27,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\123'",
        skip: true,
        languageOptions: {
          sourceType: 'script' as const,
        },
        output: "const foo = '\\u{53}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\00'",
        skip: true,
        languageOptions: {
          sourceType: 'script' as const,
        },
        output: "const foo = '\\u{0}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\1\\12\\123\\4\\45'",
        skip: true,
        languageOptions: {
          sourceType: 'script' as const,
        },
        output: "const foo = '\\u{1}\\u{A}\\u{53}\\u{4}\\u{25}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\400'",
        skip: true,
        languageOptions: {
          sourceType: 'script' as const,
        },
        output: "const foo = '\\u{20}0'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: "const foo = '\\x7A\\u2661\\uD83D\\uDCA9'",
        output: "const foo = '\\u{7A}\\u{2661}\\u{1F4A9}'",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 37,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = `\\x7A${bar}\\u2661`',
        output: 'const foo = `\\u{7A}${bar}\\u{2661}`',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 31,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = `\\\\\\x7A`',
        output: 'const foo = `\\\\\\u{7A}`',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\x7A/u',
        output: 'const foo = /\\u{7A}/u',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\u0061/v',
        output: 'const foo = /\\u{61}/v',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\uD83D\\uDCA9/u',
        output: 'const foo = /\\u{1F4A9}/u',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\[\\uD83D\\uDCA9/u',
        output: 'const foo = /\\[\\u{1F4A9}/u',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 30,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /[\\x2D]/u',
        output: 'const foo = /[\\u{2D}]/u',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /[\\cA]/u',
        output: 'const foo = /[\\u{1}]/u',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\cA/u',
        output: 'const foo = /\\u{1}/u',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\cA/',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 18,
            suggestions: [
              {
                messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
                desc: 'Use Unicode code point escapes and add the `u` flag.',
                output: 'const foo = /\\u{1}/u',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = /\\u0061/',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
                desc: 'Use Unicode code point escapes and add the `u` flag.',
                output: 'const foo = /\\u{61}/u',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = /\\u{61}/',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
                desc: 'Use Unicode code point escapes and add the `u` flag.',
                output: 'const foo = /\\u{61}/u',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = /\\x7A/g',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
                desc: 'Use Unicode code point escapes and add the `u` flag.',
                output: 'const foo = /\\u{7A}/gu',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = /\\x61\\_/',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\u{61}\\_/',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 23,
            suggestions: [],
          },
        ],
      },
    ],
  },
  {
    label: 'Documentation',
    valid: [
      {
        code: "const foo = '\\u{7A}';\nconst bar = '\\u{2661}';\nconst baz = '\\u{1F4A9}';",
        output: null,
        errors: [],
      },
      {
        code: 'const foo = `\\u{7A}${bar}\\u{2661}`;',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{61}/u;',
        output: null,
        errors: [],
      },
      {
        code: 'const foo = /\\u{61}/u;',
        output: null,
        errors: [],
      },
    ],
    invalid: [
      {
        code: "const foo = '\\x7A';\nconst bar = '\\u2661';\nconst baz = '\\uD83D\\uDCA9';\n",
        output:
          "const foo = '\\u{7A}';\nconst bar = '\\u{2661}';\nconst baz = '\\u{1F4A9}';\n",
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 2,
            column: 13,
            endLine: 2,
            endColumn: 21,
            suggestions: [],
          },
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 3,
            column: 13,
            endLine: 3,
            endColumn: 27,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = `\\x7A${bar}\\u2661`;\n',
        output: 'const foo = `\\u{7A}${bar}\\u{2661}`;\n',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 31,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\u0061/u;\n',
        output: 'const foo = /\\u{61}/u;\n',
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = /\\u0061/;\n',
        output: null,
        errors: [
          {
            messageId: 'prefer-unicode-code-point-escapes',
            message: 'Prefer Unicode code point escapes.',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'prefer-unicode-code-point-escapes/add-unicode-flag',
                desc: 'Use Unicode code point escapes and add the `u` flag.',
                output: 'const foo = /\\u{61}/u;\n',
              },
            ],
          },
        ],
      },
    ],
  },
];
// tsgo rejects these five legacy script cases with TS1487/TS1488 before
// dispatching rules. Retain them as explicit skips; Go source-only tests verify
// their complete diagnostics and fixes.
const unsupported = new Set(
  groups
    .flatMap((group) => [...group.valid, ...group.invalid])
    .filter((item) => 'skip' in item && item.skip)
    .map((item) => item.code),
);
const supported = (item: { code: string }) => !unsupported.has(item.code);
for (const code of unsupported) {
  test.skip(`legacy script syntax rejected by tsgo: ${code}`, () => {});
}

const filename = 'src/virtual.js';
const ruleName = 'unicorn/prefer-unicode-code-point-escapes';
for (const group of groups) {
  describe(group.label, () => {
    new RuleTester().run('prefer-unicode-code-point-escapes', null as never, {
      valid: group.valid.filter(supported).map((c) => ({ ...c, filename })),
      invalid: group.invalid.filter(supported).map((c) => ({ ...c, filename })),
    });
  });
}

// The suite wrapper checks messages only. Verify every upstream range, edit,
// and suggestion through the native IPC API as well.
test('matches upstream ranges, autofixes, and suggestions', async () => {
  const { config, configDirectory } = await buildConfigForSettings(
    path.resolve(import.meta.dirname, '../rslint.config.mjs'),
    undefined,
  );
  const absoluteFilename = path.resolve(import.meta.dirname, '..', filename);
  const apply = (
    code: string,
    fixes: { startPos: number; endPos: number; text: string }[] = [],
  ) =>
    [...fixes]
      .sort((a, b) => b.startPos - a.startPos)
      .reduce(
        (text, fix) =>
          text.slice(0, fix.startPos) + fix.text + text.slice(fix.endPos),
        code,
      );
  for (const group of groups) {
    for (const item of group.invalid.filter(supported)) {
      const request = {
        workingDirectory: process.cwd(),
        configDirectory,
        config: [
          ...config,
          {
            ...('languageOptions' in item
              ? { languageOptions: item.languageOptions }
              : {}),
            rules: { [ruleName]: 'error' as const },
          },
        ],
        fileContents: { [absoluteFilename]: item.code },
      };
      const result = await lint(request);
      expect(result.fileCount).toBe(1);
      expect(result.ruleCount).toBe(1);
      expect(
        result.diagnostics.map((d) => ({
          ruleName: d.ruleName,
          messageId: d.messageId,
          message: d.message,
          range: d.range,
          suggestions: (d.suggestions ?? []).map((s) => ({
            messageId: s.messageId,
            desc: s.message,
            output: apply(item.code, s.fixes),
          })),
        })),
      ).toEqual(
        item.errors.map((e) => ({
          ruleName,
          messageId: e.messageId,
          message: e.message,
          range: {
            start: { line: e.line, column: e.column },
            end: { line: e.endLine, column: e.endColumn },
          },
          suggestions: e.suggestions,
        })),
      );
      const fixed = await lint({ ...request, fix: true });
      expect(Object.values(fixed.output ?? {})).toEqual(
        item.output === null ? [] : [item.output],
      );
      expect(fixed.diagnostics).toHaveLength(
        item.output === null ? item.errors.length : 0,
      );
    }
  }
});
