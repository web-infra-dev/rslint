// Upstream tests and documentation: eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-module.js
import { describe } from 'rstack/test';
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: 'unicorn',
  filename: 'upstream.js',
  languageOptions: {
    globals: {
      exports: 'readonly',
      require: 'readonly',
      module: 'readonly',
      __filename: 'readonly',
      __dirname: 'readonly',
    },
    parserOptions: {
      project: false,
      projectService: false,
    },
  },
});

describe('Strict mode', () => {
  ruleTester.run('prefer-module', {
    valid: [
      'const foo = "use strict";',
      '("use strict")',
      '"use strong";',
      'eval("\'use strict\'; var x = 42; x;");',
      'new Function("\'use strict\'; var x = 42; return x;");',
    ],
    invalid: [
      {
        code: `'use strict';

console.log(1);`,
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: 'console.log(1);',
              },
            ],
          },
        ],
      },
      {
        code: `"use strict";

console.log(1);`,
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: 'console.log(1);',
              },
            ],
          },
        ],
      },
      {
        code: `function foo () {
	"use strict";
	console.log(1);
}`,
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: `function foo () {
	console.log(1);
}`,
              },
            ],
          },
        ],
      },
      {
        code: `'use strict';

console.log(1);`,
        filename: 'example.mjs',
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [],
          },
        ],
        output: 'console.log(1);',
      },
    ],
  });
});

// rslint already parses top-level returns without Espree's
// ecmaFeatures.globalReturn parser flag.
describe('Global return', () => {
  ruleTester.run('prefer-module', {
    valid: [
      `function a() {
	return;
}`,
    ],
    invalid: [
      {
        code: `if (foo) {
	return;
}`,
        languageOptions: {
          sourceType: 'script',
        },
        errors: [
          {
            messageId: 'error/global-return',
            message: '"return" should be used inside a function.',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('__dirname and __filename', () => {
  ruleTester.run('prefer-module', {
    valid: [
      `const __filename = 1;
const foo = __filename;`,
      `const __dirname = 1;
const foo = __dirname;`,
      'import {__filename as filename} from "foo.mjs"',
      'const foo = 1;export {foo as __dirname}',
    ],
    invalid: [
      {
        code: 'const dirname = __dirname;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'const dirname = import.meta.dirname;',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'const dirname = path.dirname(url.fileURLToPath(import.meta.url));',
              },
            ],
          },
        ],
      },
      {
        code: 'const dirname = __filename;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'const dirname = import.meta.filename;',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output: 'const dirname = url.fileURLToPath(import.meta.url);',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = { __dirname};',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'const foo = { __dirname: import.meta.dirname};',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'const foo = { __dirname: path.dirname(url.fileURLToPath(import.meta.url))};',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = {__filename, };',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'const foo = {__filename: import.meta.filename, };',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'const foo = {__filename: url.fileURLToPath(import.meta.url), };',
              },
            ],
          },
        ],
      },
      {
        code: 'if (__dirname.startsWith("/project/src/")) {}',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output:
                  'if (import.meta.dirname.startsWith("/project/src/")) {}',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'if (path.dirname(url.fileURLToPath(import.meta.url)).startsWith("/project/src/")) {}',
              },
            ],
          },
        ],
      },
      {
        code: 'if (__filename.endsWith(".js")) {}',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'if (import.meta.filename.endsWith(".js")) {}',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'if (url.fileURLToPath(import.meta.url).endsWith(".js")) {}',
              },
            ],
          },
        ],
      },
    ],
  });
});

describe('require', () => {
  ruleTester.run('prefer-module', {
    valid: [
      `import {createRequire} from 'module';
const require = createRequire(import.meta.url);
const foo = require("foo");`,
      `const require = function require(require, id) {
	return require(id);
}`,
      `const require = class A {
	require(require) {
		return require(id);
	}
}`,
    ],
    invalid: [
      {
        code: 'require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "foo";',
              },
            ],
          },
        ],
      },
      {
        code: "require('foo');",
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import 'foo';",
              },
            ],
          },
        ],
      },
      {
        code: 'require( (("foo")) );',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: '((require))("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 3,
            endLine: 1,
            endColumn: 10,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "foo";',
              },
            ],
          },
        ],
      },
      {
        code: '(( require("foo") ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 4,
            endLine: 1,
            endColumn: 11,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: ' import "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo=require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = require.resolve("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
      },
      {
        code: `const foo
	=
	require("foo");`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = require( (("foo")) );',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = ((require))("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = (( require("foo") ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 23,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import foo from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo}=require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: `const {foo}
	=
	require("foo");`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo} = require( (("foo")) );',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo} = ((require))("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo} = (( require("foo") ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 25,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo: foo}=require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 25,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: `const {foo: foo}
	=
	require("foo");`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo: foo} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo: foo} = require( (("foo")) );',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo: foo} = ((require))("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 29,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo: foo} = (( require("foo") ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 23,
            endLine: 1,
            endColumn: 30,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as foo} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo:bar}=require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 17,
            endLine: 1,
            endColumn: 24,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: `const {foo:bar}
	=
	require("foo");`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo:bar} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo:bar} = require( (("foo")) );',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 26,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo:bar} = ((require))("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 28,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo:bar} = (( require("foo") ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 29,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {foo as bar} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz}=require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 47,
            endLine: 1,
            endColumn: 54,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: `const {
	a   :foo,
	b:   bar,
	default   :   baz,
}
	=
	require("foo");`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 7,
            column: 2,
            endLine: 7,
            endColumn: 9,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: `import {
	a as foo,
	b as bar,
	default as baz,
} from "foo";`,
              },
            ],
          },
        ],
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 49,
            endLine: 1,
            endColumn: 56,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = require( (("foo")) );',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 49,
            endLine: 1,
            endColumn: 56,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = ((require))("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 51,
            endLine: 1,
            endColumn: 58,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const {a   :foo, b:   bar, default   :   baz} = (( require("foo") ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 52,
            endLine: 1,
            endColumn: 59,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output:
                  'import {a as foo, b as bar, default as baz} from  "foo" ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const {} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 19,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import {} from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: 'const{   }=require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 19,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import{   } from "foo";',
              },
            ],
          },
        ],
      },
      {
        code: `const r = require;
const foo = r("foo");`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
      },
      {
        code: 'new require("foo")',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 12,
            suggestions: [],
          },
        ],
      },
      {
        code: 'require("foo", extraArgument)',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const a = require()',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
      },
      {
        code: 'require(..."foo")',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'require("../" + "file.js")',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'require(file)',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'a = require("foo")',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 12,
            suggestions: [],
          },
        ],
      },
      {
        code: 'function a(a = require("foo")) {}',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 16,
            endLine: 1,
            endColumn: 23,
            suggestions: [],
          },
        ],
      },
      {
        code: 'let foo = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 18,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = require("foo"), bar = require("bar");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 35,
            endLine: 1,
            endColumn: 42,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const {[foo]: foo} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 29,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const {["foo"]: foo} = require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 31,
            suggestions: [],
          },
        ],
      },
      {
        code: 'if (foo) require("foo");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 17,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = require`foo`;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
      },
      {
        code: `function loadModule() {
	return require("foo");
}`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 2,
            column: 9,
            endLine: 2,
            endColumn: 16,
            suggestions: [],
          },
        ],
      },
      {
        code: `function loadModule() {
	const foo = require("foo");
	return foo;
}`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 2,
            column: 14,
            endLine: 2,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = require("foo"), bar = 1;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('exports and module', () => {
  ruleTester.run('prefer-module', {
    valid: [
      'function exports(exports) {}',
      'function module(module) {}',
      `const exports = foo;
exports.bar = bar;`,
      'const exports = 1;',
      `const module = foo;
module.exports = bar;
module.exports.bar = bar;`,
      'const module = 1;',
    ],
    invalid: [
      {
        code: 'exports = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default foo;',
              },
            ],
          },
        ],
      },
      {
        code: 'module.exports = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default foo;',
              },
            ],
          },
        ],
      },
      {
        code: '(( ((exports)) = ((foo)) ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 13,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: ' export default  ((foo)) ;',
              },
            ],
          },
        ],
      },
      {
        code: '(( ((module.exports)) = ((foo)) ));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 12,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: ' export default  ((foo)) ;',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = 1;exports.foo = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as foo};',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = 1;module.exports.foo = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as foo};',
              },
            ],
          },
        ],
      },
      {
        code: 'exports["foo"] = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'module.exports["foo"] = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = exports;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = exports.foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 20,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = module.exports;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = module.exports.foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: 'module["exports"] = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
      },
      {
        code: 'module[exports] = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 15,
            suggestions: [],
          },
        ],
      },
      {
        code: 'module.exports.foo.bar = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = 1;exports.default = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as default};',
              },
            ],
          },
        ],
      },
      {
        code: 'const foo = 1;module.exports.default = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'const foo = 1;export {foo as default};',
              },
            ],
          },
        ],
      },
      {
        code: 'exports.foo.bar = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'exports = 1;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default 1;',
              },
            ],
          },
        ],
      },
      {
        code: 'exports.foo = [];',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'module.exports = function() {};',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default function() {};',
              },
            ],
          },
        ],
      },
      {
        code: 'module.exports.foo = foo || bar;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [],
          },
        ],
      },
      {
        code: 'exports += foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const foo = module.children',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 13,
            endLine: 1,
            endColumn: 19,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const parentModule = module.parent',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 22,
            endLine: 1,
            endColumn: 28,
            suggestions: [],
          },
        ],
      },
      {
        code: `function foo() {
	exports.foo = foo;
	module.exports.foo = foo;
}`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 9,
            suggestions: [],
          },
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 3,
            column: 2,
            endLine: 3,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('TypeScript', () => {
  ruleTester.run('prefer-module', {
    valid: [
      {
        code: 'type module = number[];',
        filename: 'upstream.ts',
      },
      {
        code: 'type ModuleRegistry = { [module: string]: string };',
        filename: 'upstream.ts',
      },
      {
        code: 'const module = 1; type ModuleRegistry = { [module: string]: string };',
        filename: 'upstream.ts',
      },
      {
        code: 'type module = number[]; type ModuleRegistry = { [module: string]: string };',
        filename: 'upstream.ts',
      },
      {
        code: 'type Data = { [module in keyof string]: number; };',
        filename: 'upstream.ts',
      },
      {
        code: 'type ModuleRegistry = { [exports: string]: string };',
        filename: 'upstream.ts',
      },
    ],
    invalid: [
      {
        code: 'const foo: Foo = require("foo");',
        filename: 'upstream.ts',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 18,
            endLine: 1,
            endColumn: 25,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const {foo}: Foo = require("foo");',
        filename: 'upstream.ts',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('JSON modules', () => {
  ruleTester.run('prefer-module', {
    valid: [],
    invalid: [
      {
        code: 'const data = require("./a.json");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import data from "./a.json" with {type: "json"};',
              },
            ],
          },
        ],
      },
      {
        code: 'require("./a.json");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "./a.json" with {type: "json"};',
              },
            ],
          },
        ],
      },
      {
        code: "const data = (require(('./a.json')));",
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import data from './a.json' with {type: 'json'};",
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo} = require("./a.json");',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 22,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('.cjs files', () => {
  ruleTester.run('prefer-module', {
    valid: [
      {
        code: '__dirname',
        filename: 'foo.cjs',
      },
      {
        code: '__dirname',
        filename: 'foo.cjS',
        skip: 'The native compiler excludes mixed-case .cjS filenames before rule execution; Go extras cover the suffix behavior.',
      },
    ],
    invalid: [
      {
        code: '__filename',
        filename: 'foo.mjs',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 11,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'import.meta.filename',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output: 'url.fileURLToPath(import.meta.url)',
              },
            ],
          },
        ],
      },
      {
        code: 'require("lodash")',
        filename: 'foo.js',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "lodash"',
              },
            ],
          },
        ],
      },
      {
        code: 'require("lodash")',
        filename: 'foo.cjs/foo.js',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: 'import "lodash"',
              },
            ],
          },
        ],
      },
    ],
  });
});

describe('Comments in named exports', () => {
  ruleTester.run('prefer-module', {
    valid: [],
    invalid: [
      {
        code: `const foo = 1;
exports /* note */ .foo = foo;`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: `const foo = 1;
exports./* c */foo = foo;`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: `const foo = 1;
exports.foo /* c */ = foo;`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: `const foo = 1;
module.exports /* c */ .foo = foo;`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 2,
            column: 1,
            endLine: 2,
            endColumn: 7,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('Constructor callees', () => {
  ruleTester.run('prefer-module', {
    valid: [],
    invalid: [
      {
        code: 'new __filename.constructor("text");',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new import.meta.filename.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url)).constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new (__filename).constructor("text");',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 16,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new (import.meta.filename).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url)).constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new __filename.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new (import.meta.filename).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url)).constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new __filename!.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __filename: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output: 'new (import.meta.filename)!.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  'new (url.fileURLToPath(import.meta.url))!.constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new __dirname.constructor("text");',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new import.meta.dirname.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url))).constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new (__dirname).constructor("text");',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 6,
            endLine: 1,
            endColumn: 15,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new (import.meta.dirname).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url))).constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new __dirname.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new (import.meta.dirname).constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url))).constructor("text");',
              },
            ],
          },
        ],
      },
      {
        code: 'new __dirname!.constructor("text");',
        filename: 'upstream.ts',
        languageOptions: {
          globals: {
            __dirname: 'readonly',
          },
        },
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 5,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output: 'new (import.meta.dirname)!.constructor("text");',
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  'new (path.dirname(url.fileURLToPath(import.meta.url)))!.constructor("text");',
              },
            ],
          },
        ],
      },
    ],
  });
});

describe('Static template module sources', () => {
  ruleTester.run('prefer-module', {
    valid: [],
    invalid: [
      {
        code: 'const file = require(`file`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from 'file';",
              },
            ],
          },
        ],
      },
      {
        code: 'require(`file`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import 'file';",
              },
            ],
          },
        ],
      },
      {
        code: 'const file = require(`./file.json`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './file.json' with {type: 'json'};",
              },
            ],
          },
        ],
      },
      {
        code: 'const {foo: bar} = require(`file`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 20,
            endLine: 1,
            endColumn: 27,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import {foo as bar} from 'file';",
              },
            ],
          },
        ],
      },
      {
        code: "const file = require(`./it's.js`);",
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './it\\'s.js';",
              },
            ],
          },
        ],
      },
      {
        code: `const file = require(\`./line
file.js\`);`,
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './line\\nfile.js';",
              },
            ],
          },
        ],
      },
      {
        code: 'const file = require(`./back\\`tick.js`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './back`tick.js';",
              },
            ],
          },
        ],
      },
      {
        code: 'const file = require(`./fi\\u006ce`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from './file';",
              },
            ],
          },
        ],
      },
      {
        code: 'const file = require(/* keep */ (`file`));',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import file from /* keep */ 'file';",
              },
            ],
          },
        ],
      },
      {
        code: 'const file = require(`${name}`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const file = require(tag`file`);',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 21,
            suggestions: [],
          },
        ],
      },
      {
        code: 'const file: Foo = require(`file`);',
        filename: 'upstream.ts',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 19,
            endLine: 1,
            endColumn: 26,
            suggestions: [],
          },
        ],
      },
    ],
  });
});

describe('Documentation examples', () => {
  ruleTester.run('prefer-module', {
    valid: [
      `const __dirname = import.meta.dirname;
const __filename = import.meta.filename;`,
      `import {fileURLToPath} from 'node:url';
import path from 'node:path';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(fileURLToPath(import.meta.url));`,
      `import {fileURLToPath} from 'node:url';
const foo = fileURLToPath(new URL('foo.js', import.meta.url));`,
      "const foo = new URL('foo.js', import.meta.url);",
      `function run() {
	if (foo) {
		return;
	}
}
run();`,
      "const file = path.join(import.meta.dirname, 'foo.js');",
      "const file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), 'foo.js');",
      "const content = fs.readFileSync(import.meta.filename, 'utf8');",
      "import {fromPairs} from 'lodash-es';",
      'export default foo;',
      'const foo = 1;export {foo};',
    ],
    invalid: [
      {
        code: "'use strict';",
        errors: [
          {
            messageId: 'error/use-strict-directive',
            message: 'Do not use "use strict" directive.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
            suggestions: [
              {
                messageId: 'suggestion/use-strict-directive',
                desc: 'Remove "use strict" directive.',
                output: '',
              },
            ],
          },
        ],
      },
      {
        code: `if (foo) {
	return;
}`,
        languageOptions: {
          sourceType: 'script',
        },
        errors: [
          {
            messageId: 'error/global-return',
            message: '"return" should be used inside a function.',
            line: 2,
            column: 2,
            endLine: 2,
            endColumn: 8,
            suggestions: [],
          },
        ],
      },
      {
        code: "const file = path.join(__dirname, 'foo.js');",
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__dirname".',
            line: 1,
            column: 24,
            endLine: 1,
            endColumn: 33,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-dirname',
                desc: 'Replace `__dirname` with `import.meta.dirname`.',
                output:
                  "const file = path.join(import.meta.dirname, 'foo.js');",
              },
              {
                messageId: 'suggestion/import-meta-url-to-dirname',
                desc: 'Replace `__dirname` with `…(import.meta.url)`.',
                output:
                  "const file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), 'foo.js');",
              },
            ],
          },
        ],
      },
      {
        code: "const content = fs.readFileSync(__filename, 'utf8');",
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "__filename".',
            line: 1,
            column: 33,
            endLine: 1,
            endColumn: 43,
            suggestions: [
              {
                messageId: 'suggestion/import-meta-filename',
                desc: 'Replace `__filename` with `import.meta.filename`.',
                output:
                  "const content = fs.readFileSync(import.meta.filename, 'utf8');",
              },
              {
                messageId: 'suggestion/import-meta-url-to-filename',
                desc: 'Replace `__filename` with `…(import.meta.url)`.',
                output:
                  "const content = fs.readFileSync(url.fileURLToPath(import.meta.url), 'utf8');",
              },
            ],
          },
        ],
      },
      {
        code: "const {fromPairs} = require('lodash');",
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "require".',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 28,
            suggestions: [
              {
                messageId: 'suggestion/import',
                desc: 'Switch to `import`.',
                output: "import {fromPairs} from 'lodash';",
              },
            ],
          },
        ],
      },
      {
        code: 'module.exports = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "module".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 7,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export default foo;',
              },
            ],
          },
        ],
      },
      {
        code: 'exports.foo = foo;',
        errors: [
          {
            messageId: 'error/identifier',
            message: 'Do not use "exports".',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 8,
            suggestions: [
              {
                messageId: 'suggestion/export',
                desc: 'Switch to `export`.',
                output: 'export {foo as foo};',
              },
            ],
          },
        ],
      },
    ],
  });
});
