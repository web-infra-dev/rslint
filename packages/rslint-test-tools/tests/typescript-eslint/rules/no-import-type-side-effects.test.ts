import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  // @ts-ignore
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('no-import-type-side-effects', {
  valid: [
    "import T from 'mod';",
    "import * as T from 'mod';",
    "import { T } from 'mod';",
    "import type { T } from 'mod';",
    "import type { T, U } from 'mod';",
    "import { type T, U } from 'mod';",
    "import { T, type U } from 'mod';",
    "import type T from 'mod';",
    "import type T, { U } from 'mod';",
    "import T, { type U } from 'mod';",
    "import type * as T from 'mod';",
    "import 'mod';",
  ],
  invalid: [
    {
      code: "import { type A } from 'mod';",
      errors: [{ messageId: 'useTopLevelQualifier' }],
      output: "import type { A } from 'mod';",
    },
    {
      code: "import { type A as AA } from 'mod';",
      errors: [{ messageId: 'useTopLevelQualifier' }],
      output: "import type { A as AA } from 'mod';",
    },
    {
      code: "import { type A, type B } from 'mod';",
      errors: [{ messageId: 'useTopLevelQualifier' }],
      output: "import type { A, B } from 'mod';",
    },
    {
      code: "import { type A as AA, type B as BB } from 'mod';",
      errors: [{ messageId: 'useTopLevelQualifier' }],
      output: "import type { A as AA, B as BB } from 'mod';",
    },
  ],
});
