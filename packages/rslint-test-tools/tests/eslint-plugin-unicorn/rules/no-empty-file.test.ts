// Upstream: eslint-plugin-unicorn v77.0.0 (MIT).
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-empty-file.js
import { RuleTester } from '../rule-tester';

const tester = new RuleTester();
const invalid = (code: string, filename: string, allowComments?: boolean) => ({
  code,
  filename,
  ...(allowComments === undefined ? {} : { options: [{ allowComments }] }),
  errors: [
    { messageId: 'no-empty-file', message: 'Empty files are not allowed.' },
  ],
  output: null,
});

// Complete upstream snapshot group, retaining unsupported cases as explained skips.
tester.run('no-empty-file', null as never, {
  valid: [
    { code: 'const x = 0;', filename: 'example.js' },
    { code: ';; const x = 0;', filename: 'example.js' },
    { code: '{{{;;const x = 0;}}}', filename: 'example.js' },
    { code: "'use strict';\nconst x = 0;", filename: 'example.js' },
    { code: ";;'use strict';", filename: 'example.js' },
    { code: "{'use strict';}", filename: 'example.js' },
    { code: '("use strict")', filename: 'example.js' },
    { code: '`use strict`', filename: 'example.js' },
    { code: '({})', filename: 'example.js' },
    {
      code: "#!/usr/bin/env node\nconsole.log('done');",
      filename: 'example.js',
    },
    { code: 'false', filename: 'example.js' },
    { code: '("")', filename: 'example.js' },
    { code: 'NaN', filename: 'example.js' },
    { code: 'undefined', filename: 'example.js' },
    { code: 'null', filename: 'example.js' },
    { code: '[]', filename: 'example.js' },
    { code: '(() => {})()', filename: 'example.js' },
    { code: '/// <reference types="example" />', filename: 'example.d.ts' },
    { code: '/// <reference types="example" />', filename: 'example.ts' },
    {
      code: '// comment',
      filename: 'example.js',
      options: [{ allowComments: true }],
    },
    {
      code: '/* comment */',
      filename: 'example.js',
      options: [{ allowComments: true }],
    },
    {
      code: '/**\n * @typedef {object} Foo\n * @property {string} bar\n */',
      filename: 'example.js',
      options: [{ allowComments: true }],
    },
    {
      code: '// No need to write tests here.',
      filename: 'example.test.ts',
      options: [{ allowComments: true }],
    },
  ],
  invalid: [
    invalid('', 'example.js'),
    invalid('\uFEFF', 'example.js'),
    invalid(' ', 'example.js'),
    invalid('\t', 'example.js'),
    invalid('\n', 'example.js'),
    invalid('\r', 'example.js'),
    invalid('\r\n', 'example.js'),
    invalid('', 'example.js'),
    invalid('// comment', 'example.js'),
    invalid('/* comment */', 'example.js'),
    invalid('#!/usr/bin/env node', 'example.js'),
    invalid("'use asm';", 'example.js'),
    invalid("'use strict';", 'example.js'),
    invalid('"use strict"', 'example.js'),
    invalid('""', 'example.js'),
    invalid(';', 'example.js'),
    invalid(';;', 'example.js'),
    invalid('{}', 'example.js'),
    invalid('{;;}', 'example.js'),
    invalid('{{}}', 'example.js'),
    invalid('', 'example.js', true),
    invalid(' ', 'example.js', true),
    invalid('#!/usr/bin/env node', 'example.js', true),
    invalid('#!/usr/bin/env node\n// comment', 'example.js', true),
    invalid('; // comment', 'example.js', true),
    invalid("'use strict'; // comment", 'example.js', true),
    invalid('{/* comment */}', 'example.js', true),
    invalid('{}', 'example.mjs'),
    invalid('{}', 'example.ts'),
    invalid('{}', 'example.tsx'),
    invalid('{}', 'example.jsx'),
    invalid('{}', 'example.cts'),
  ],
});

// Regression for upstream issue #2175.
tester.run('no-empty-file', null as never, {
  valid: [{ code: '(() => {})();', filename: 'example.ts' }],
  invalid: [
    invalid('"";', 'example.ts'),
    invalid('"use strict";', 'example.ts'),
  ],
});

// Additional documentation examples; the remaining examples are already in the snapshot group.
tester.run('no-empty-file', null as never, {
  valid: [
    { code: ';;\nconst x = 0;', filename: 'example.js' },
    { code: '{\n\tconst x = 0;\n}', filename: 'example.js' },
  ],
  invalid: [
    invalid('// Comment', 'example.js'),
    invalid('/* Comment */', 'example.js'),
    invalid('{\n}', 'example.js'),
  ],
});

// The native wrapper has no parser/language or processor support. Retain each
// unsupported upstream case as a named skip, including its original options.
const unsupportedCases = [
  {
    code: '<template><div/></template>',
    filename: 'example.vue',
    reason: 'vue parser',
    expected: 'valid',
  },
  {
    code: '<template><div/></template>\n<script></script>',
    filename: 'example.vue',
    reason: 'vue parser',
    expected: 'valid',
  },
  {
    code: '<div>Hello</div>',
    filename: 'example.html',
    reason: 'html parser',
    expected: 'valid',
  },
  {
    code: '<!DOCTYPE html>',
    filename: 'example.html',
    reason: 'html parser',
    expected: 'valid',
  },
  {
    code: '<!-- comment -->',
    filename: 'example.html',
    options: [{ allowComments: true }],
    reason: 'html parser',
    expected: 'valid',
  },
  {
    code: 'a { color: red; }',
    filename: 'example.css',
    reason: 'css language',
    expected: 'valid',
  },
  {
    code: '/* comment */',
    filename: 'example.css',
    options: [{ allowComments: true }],
    reason: 'css language',
    expected: 'valid',
  },
  {
    code: '# Title',
    filename: 'example.md',
    reason: 'markdown language',
    expected: 'valid',
  },
  {
    code: '<!-- a --> text <!-- b -->',
    filename: 'example.md',
    reason: 'markdown language',
    expected: 'valid',
  },
  {
    code: '<!-- comment -->',
    filename: 'example.md',
    options: [{ allowComments: true }],
    reason: 'markdown language',
    expected: 'valid',
  },
  {
    code: 'key: value',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'valid',
  },
  {
    code: '&anchor value',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'valid',
  },
  {
    code: '---\n---\nkey: value',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'valid',
  },
  {
    code: '# comment',
    filename: 'example.yaml',
    options: [{ allowComments: true }],
    reason: 'yaml language',
    expected: 'valid',
  },
  {
    code: '# Logs\nnode_modules\n',
    filename: '.gitignore',
    reason: 'plain-text parser',
    expected: 'valid',
  },
  {
    code: 'key = "value"',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'valid',
  },
  {
    code: '[table]',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'valid',
  },
  {
    code: '[[table]]',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'valid',
  },
  {
    code: 'items = []',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'valid',
  },
  {
    code: 'table = {}',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'valid',
  },
  {
    code: '# comment',
    filename: 'example.toml',
    options: [{ allowComments: true }],
    reason: 'toml language',
    expected: 'valid',
  },
  {
    code: '{}',
    filename: 'example.cJs',
    reason:
      'mixed-case extensions (TypeScript file discovery is case-sensitive)',
    expected: 'invalid',
  },
  {
    code: '{}',
    filename: 'example.MTS',
    reason:
      'mixed-case extensions (TypeScript file discovery is case-sensitive)',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.md',
    reason: 'non-JS/TS extension',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.vue',
    reason: 'non-JS/TS extension',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.svelte',
    reason: 'non-JS/TS extension',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.astro',
    reason: 'non-JS/TS extension',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.css',
    reason: 'non-JS/TS extension',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.txt',
    reason: 'non-JS/TS extension',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.vue',
    reason: 'vue parser',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.html',
    reason: 'html parser',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: 'example.html',
    reason: 'html parser',
    expected: 'invalid',
  },
  {
    code: '<!-- comment -->',
    filename: 'example.html',
    reason: 'html parser',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.css',
    reason: 'css language',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: 'example.css',
    reason: 'css language',
    expected: 'invalid',
  },
  {
    code: '/* comment */',
    filename: 'example.css',
    reason: 'css language',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.md',
    reason: 'markdown language',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: 'example.md',
    reason: 'markdown language',
    expected: 'invalid',
  },
  {
    code: '<!-- comment -->',
    filename: 'example.md',
    reason: 'markdown language',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '# comment',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '%YAML 1.2\n# comment',
    filename: 'example.yaml',
    options: [{ allowComments: true }],
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '&anchor',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '&anchor\n# comment',
    filename: 'example.yaml',
    options: [{ allowComments: true }],
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '!tag\n# comment',
    filename: 'example.yaml',
    options: [{ allowComments: true }],
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '---\n# comment',
    filename: 'example.yaml',
    options: [{ allowComments: true }],
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '...\n# comment',
    filename: 'example.yaml',
    options: [{ allowComments: true }],
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '---\n---',
    filename: 'example.yaml',
    reason: 'yaml language',
    expected: 'invalid',
  },
  {
    code: '',
    filename: '.gitignore',
    reason: 'plain-text parser',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: '.gitignore',
    reason: 'plain-text parser',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'invalid',
  },
  {
    code: '# comment',
    filename: 'example.toml',
    reason: 'toml language',
    expected: 'invalid',
  },
  {
    code: '',
    filename: 'example.toml',
    options: [{ allowComments: true }],
    reason: 'toml language',
    expected: 'invalid',
  },
  {
    code: '   \n\t ',
    filename: 'example.toml',
    options: [{ allowComments: true }],
    reason: 'toml language',
    expected: 'invalid',
  },
];
for (const item of unsupportedCases) {
  test.skip(`upstream ${item.expected}: ${item.reason}, ${item.filename}, ${JSON.stringify(item.code)}, ${JSON.stringify(item.options)}`, () => {});
}
test.skip('upstream processor: empty block.js extracted from document.txt containing Physical file', () => {});
