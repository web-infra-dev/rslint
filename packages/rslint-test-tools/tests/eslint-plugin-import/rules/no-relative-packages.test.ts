// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import {
  RuleTester,
  type TestCase,
  type ValidTestCase,
} from '@rslint/rule-tester';
import { test } from 'rstack/test';
import { fixtureConfig } from '../config';
import assert from 'node:assert/strict';
import { lint } from '@rslint/core/internal';
import {
  createFixtureRoot,
  removeFixtureRoot,
} from '../../../../rslint-rule-tester/src/fixture.ts';
import { buildLintRequest } from '../../../../rslint-rule-tester/src/request.ts';
import { checkResult } from '../../../../rslint-rule-tester/src/diagnostics.ts';
const config = fixtureConfig(
  'internal/plugins/import/rules/no_relative_packages/testdata/modules.txtar',
);
const cases: {
  valid: ValidTestCase[];
  invalid: TestCase[];
} = {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import foo from "./index.js"',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 1',
      code: 'import bar from "../bar"',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 2',
      code: 'import {foo} from "a"',
      filename: 'tests/files/package-named/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 3',
      code: 'const bar = require("../bar.js")',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 4',
      code: 'const bar = require("../not/a/file/path.js")',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 5',
      code: 'import "package"',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 6',
      code: 'require("../bar.js")',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 7',
      code: "import bar from 'bar';",
      filename: 'my-project/packages/foo/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 8',
      code: "import bar from 'bar';",
      filename: 'my-project/entry.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import foo from "./package-named"',
      filename: 'tests/files/bar.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `./package-named`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import foo from "../package-named"',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `../package-named`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import foo from "../package-scoped"',
      filename: 'tests/files/package/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `../package-scoped`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import bar from "../bar"',
      filename: 'tests/files/package-named/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `../bar`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import bar from '../bar';",
      filename: 'my-project/packages/foo/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `../bar`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "import entry from '../../entry.js';",
      filename: 'my-project/packages/foo/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `../../entry.js`',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import bar from './packages/bar';",
      filename: 'my-project/entry.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            'Relative import from another package is not allowed. Use a package import instead of `./packages/bar`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
  ],
};
new RuleTester(config).run('no-relative-packages', cases);
// Preserve the independent fix-pass check for the documented unsafe-upstream-fix difference.
test('upstream diagnostic ranges without unsafe autofixes', async () => {
  const root = await createFixtureRoot(config.fixtureFiles ?? {});
  try {
    for (const item of cases.invalid) {
      const { request } = buildLintRequest(
        config,
        'import/no-relative-packages',
        root,
        item,
      );
      const result = await lint({ ...request, fix: true });
      const errors = item.errors;
      checkResult(result, errors, {
        code: item.code,
        root,
        ruleName: 'import/no-relative-packages',
        hasOutput: false,
      });
      assert.ok(!result.output || Object.keys(result.output).length === 0);
    }
  } finally {
    await removeFixtureRoot(root);
  }
});
// Upstream autofix examples retained as coverage records, not native expectations.
// "import foo from \"./package-named\"" => "import foo from \"package-named\""
// "import foo from \"../package-named\"" => "import foo from \"package-named\""
// "import foo from \"../package-scoped\"" => "import foo from \"@scope/package-named\""
// "import bar from \"../bar\"" => "import bar from \"eslint-plugin-import/tests/files/bar\""
// "import bar from '../bar';" => "import bar from \"bar\";"
// "import entry from '../../entry.js';" => "import entry from \"my-project/entry.js\";"
// "import bar from './packages/bar';" => "import bar from \"bar\";"
