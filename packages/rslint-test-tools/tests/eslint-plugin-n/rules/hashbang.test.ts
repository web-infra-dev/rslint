// Upstream: eslint-plugin-n v18.3.0 tests/lib/rules/hashbang.js (shebang alias).
/**
 * @author Toru Nagashima
 * See LICENSE file in root directory for full license.
 */

import { RuleTester } from '../rule-tester';
const rule = undefined;

/**
 * Makes a file path to a fixture.
 * @param {string} name - A name.
 * @returns {string} A file path to a fixture.
 */
function fixture(name: string) {
  return name;
}

const ruleTester = new RuleTester();
ruleTester.run('hashbang', rule, {
  valid: [
    {
      name: 'string-bin/bin/test.js',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env node\nhello();',
    },
    {
      name: 'string-bin/lib/test.js',
      filename: fixture('string-bin/lib/test.js'),
      code: 'hello();',
    },
    {
      name: 'object-bin/bin/a.js',
      filename: fixture('object-bin/bin/a.js'),
      code: '#!/usr/bin/env node\nhello();',
    },
    {
      name: 'object-bin/bin/b.js',
      filename: fixture('object-bin/bin/b.js'),
      code: '#!/usr/bin/env node\nhello();',
    },
    {
      name: 'string-bin/bin/test-env-flag.js',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env -S node\nhello();',
    },
    {
      name: 'string-bin/bin/test-env-flag-node-flag.js',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env -S node --loader tsm\nhello();',
    },
    {
      name: 'string-bin/bin/test-env-ignore-environment.js',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env --ignore-environment node\nhello();',
    },
    {
      name: 'string-bin/bin/test-env-flags-node-flag.js',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env -i -S node --loader tsm\nhello();',
    },
    {
      name: 'string-bin/bin/test-block-signal.js',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env --block-signal=SIGINT -S FOO=bar node --loader tsm\nhello();',
    },
    {
      name: 'object-bin/bin/c.js',
      filename: fixture('object-bin/bin/c.js'),
      code: 'hello();',
    },
    {
      name: 'no-bin-field/lib/test.js',
      filename: fixture('no-bin-field/lib/test.js'),
      code: 'hello();',
    },
    {
      name: '<input> with shebang',
      code: '#!/usr/bin/env node\nhello();',
    },
    {
      name: '<input> without shebang',
      code: 'hello();',
    },

    // convertPath
    {
      name: 'convertPath - string-bin/src/bin/test.js',
      filename: fixture('string-bin/src/bin/test.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
    },
    {
      name: 'convertPath - string-bin/src/lib/test.js',
      filename: fixture('string-bin/src/lib/test.js'),
      code: 'hello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
    },
    {
      name: 'convertPath - object-bin/src/bin/a.js',
      filename: fixture('object-bin/src/bin/a.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
    },
    {
      name: 'convertPath - object-bin/src/bin/b.js',
      filename: fixture('object-bin/src/bin/b.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
    },
    {
      name: 'convertPath - object-bin/src/bin/c.js',
      filename: fixture('object-bin/src/bin/c.js'),
      code: 'hello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
    },
    {
      name: 'convertPath - no-bin-field/src/lib/test.js',
      filename: fixture('no-bin-field/src/lib/test.js'),
      code: 'hello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
    },

    // Should work fine if the filename is relative.
    {
      name: 'relative path - string-bin/bin/test.js',
      filename: 'tests/fixtures/shebang/string-bin/bin/test.js',
      code: '#!/usr/bin/env node\nhello();',
    },
    {
      name: 'relative path - string-bin/lib/test.js',
      filename: 'tests/fixtures/shebang/string-bin/lib/test.js',
      code: 'hello();',
    },

    // BOM and \r\n
    {
      name: 'BOM without newline',
      filename: fixture('string-bin/lib/test.js'),
      code: '\uFEFFhello();',
    },
    {
      name: 'BOM with newline',
      filename: fixture('string-bin/lib/test.js'),
      code: '\uFEFFhello();\n',
    },
    {
      name: 'with windows newline',
      filename: fixture('string-bin/lib/test.js'),
      code: 'hello();\r\n',
    },
    {
      name: 'BOM with windows newline',
      filename: fixture('string-bin/lib/test.js'),
      code: '\uFEFFhello();\r\n',
    },

    // blank lines on the top of files.
    {
      name: 'blank lines on the top of files.',
      filename: fixture('string-bin/lib/test.js'),
      code: '\n\n\nhello();',
    },

    // https://github.com/mysticatea/eslint-plugin-node/issues/51
    {
      name: 'Shebang with CLI flags',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env node --harmony\nhello();',
    },

    // use node resolution
    {
      name: 'use node resolution',
      filename: fixture('object-bin/bin/index.js'),
      code: '#!/usr/bin/env node\nhello();',
    },

    // npm unpublished files are ignored
    {
      name: 'published file cant have shebang',
      filename: fixture('unpublished/published.js'),
      code: 'hello();',
      options: [{ ignoreUnpublished: true }],
    },
    {
      name: 'unpublished file can have shebang',
      filename: fixture('unpublished/unpublished.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ ignoreUnpublished: true }],
    },
    {
      name: 'unpublished file can have noshebang',
      filename: fixture('unpublished/unpublished.js'),
      code: 'hello();',
      options: [{ ignoreUnpublished: true }],
    },

    {
      name: 'file matching additionalExecutables',
      filename: fixture('unpublished/something.test.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ additionalExecutables: ['*.test.js'] }],
    },

    // executableMap
    {
      name: '.ts maps to ts-node',
      filename: fixture('object-bin/bin/t.ts'),
      code: '#!/usr/bin/env ts-node\nhello();',
      options: [{ executableMap: { '.ts': 'ts-node' } }],
    },
    {
      name: '.ts maps to ts-node',
      filename: fixture('object-bin/bin/a.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ executableMap: { '.ts': 'ts-node' } }],
    },
  ],
  invalid: [
    {
      name: 'bin: string - match - no shebang',
      filename: fixture('string-bin/bin/test.js'),
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'bin: string - match - incorrect shebang',
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/node\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'bin: string - no match - with shebang',
      filename: fixture('string-bin/lib/test.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },
    {
      name: 'bin: {a: "./bin/a.js"} - match - no shebang',
      filename: fixture('object-bin/bin/a.js'),
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'bin: {b: "./bin/b.js"} - match - no shebang',
      filename: fixture('object-bin/bin/b.js'),
      code: '#!/usr/bin/node\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'bin: {c: "./bin"} - no match - with shebang',
      filename: fixture('object-bin/bin/c.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },
    {
      name: 'bin: undefined - no match - with shebang',
      filename: fixture('no-bin-field/lib/test.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },

    // convertPath
    {
      name: 'convertPath in options',
      filename: fixture('string-bin/src/bin/test.js'),
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'convertPath in settings',
      filename: fixture('string-bin/src/bin/test.js'),
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
      settings: {
        node: { convertPath: { 'src/**': ['^src/(.+)$', '$1'] } },
      },
    },
    {
      name: 'converted path - string-bin/src/bin/test.js',
      filename: fixture('string-bin/src/bin/test.js'),
      code: '#!/usr/bin/node\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'converted path - string-bin/src/lib/test.js',
      filename: fixture('string-bin/src/lib/test.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs no shebang.'],
    },
    {
      name: 'converted path - object-bin/src/bin/a.js',
      filename: fixture('object-bin/src/bin/a.js'),
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'converted path - object-bin/src/bin/b.js',
      filename: fixture('object-bin/src/bin/b.js'),
      code: '#!/usr/bin/node\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'converted path - object-bin/src/bin/c.js',
      filename: fixture('object-bin/src/bin/c.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs no shebang.'],
    },
    {
      name: 'converted path - no-bin-field/src/lib/test.js',
      filename: fixture('no-bin-field/src/lib/test.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      options: [{ convertPath: { 'src/**': ['^src/(.+)$', '$1'] } }],
      errors: ['This file needs no shebang.'],
    },

    // Should work fine if the filename is relative.
    {
      name: 'relative path - string-bin/bin/test.js',
      filename: 'tests/fixtures/shebang/string-bin/bin/test.js',
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'relative path - string-bin/lib/test.js',
      filename: 'tests/fixtures/shebang/string-bin/lib/test.js',
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },

    // header comments
    {
      name: 'header comments',
      filename: fixture('string-bin/bin/test.js'),
      code: '/* header */\nhello();',
      output: '#!/usr/bin/env node\n/* header */\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },

    // BOM and \r\n
    {
      filename: fixture('string-bin/bin/test.js'),
      code: '\uFEFFhello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: 'hello();\n',
      output: '#!/usr/bin/env node\nhello();\n',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: 'hello();\r\n',
      output: '#!/usr/bin/env node\nhello();\r\n',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: '\uFEFFhello();\n',
      output: '#!/usr/bin/env node\nhello();\n',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: '\uFEFFhello();\r\n',
      output: '#!/usr/bin/env node\nhello();\r\n',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: '#!/usr/bin/env node\r\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file must have Unix linebreaks (LF).'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: '\uFEFF#!/usr/bin/env node\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file must not have Unicode BOM.'],
    },
    {
      filename: fixture('string-bin/bin/test.js'),
      code: '\uFEFF#!/usr/bin/env node\r\nhello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: [
        'This file must not have Unicode BOM.',
        'This file must have Unix linebreaks (LF).',
      ],
    },

    // https://github.com/mysticatea/eslint-plugin-node/issues/51
    {
      name: 'Shebang with CLI flags',
      filename: fixture('string-bin/lib/test.js'),
      code: '#!/usr/bin/env node --harmony\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },

    // use node resolution
    {
      name: 'use node resolution',
      filename: fixture('object-bin/bin/index.js'),
      code: 'hello();',
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },

    // npm unpublished files are ignored
    {
      name: 'unpublished file should not have shebang',
      filename: fixture('unpublished/unpublished.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },
    {
      name: 'published file should have shebang',
      filename: fixture('unpublished/published.js'),
      code: '#!/usr/bin/env node\nhello();',
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },

    {
      name: 'unpublished file shebang ignored',
      filename: fixture('unpublished/published.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ ignoreUnpublished: true }],
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },

    {
      name: 'executable in additionalExecutables without shebang',
      filename: fixture('unpublished/something.test.js'),
      code: 'hello();',
      options: [{ additionalExecutables: ['*.test.js'] }],
      output: '#!/usr/bin/env node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env node".'],
    },
    {
      name: 'file not in additionalExecutables with shebang',
      filename: fixture('unpublished/not-a-test.js'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ additionalExecutables: ['*.test.js'] }],
      output: 'hello();',
      errors: ['This file needs no shebang.'],
    },

    // executableMap
    {
      name: '.ts maps to ts-node',
      filename: fixture('object-bin/bin/t.ts'),
      code: 'hello();',
      options: [{ executableMap: { '.ts': 'ts-node' } }],
      output: '#!/usr/bin/env ts-node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env ts-node".'],
    },
    {
      name: '.ts maps to ts-node',
      filename: fixture('object-bin/bin/t.ts'),
      code: '#!/usr/bin/env node\nhello();',
      options: [{ executableMap: { '.ts': 'ts-node' } }],
      output: '#!/usr/bin/env ts-node\nhello();',
      errors: ['This file needs shebang "#!/usr/bin/env ts-node".'],
    },
  ],
});
