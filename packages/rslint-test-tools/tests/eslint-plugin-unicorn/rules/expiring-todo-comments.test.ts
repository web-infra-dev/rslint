// Upstream: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/expiring-todo-comments.js
// Includes documentation examples. RegExp objects use equivalent JSON string patterns.
// Other-language cases are retained as explained Go skips in the upstream suite.
import { RuleTester } from '../rule-tester';

new RuleTester().run('expiring-todo-comments', {} as never, {
  valid: [
    {
      code: '// TODO [2200-12-12]: Too long... Can you feel it?',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// FIXME [2200-12-12]: Too long... Can you feel it?',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// XXX [2200-12-12]: Too long... Can you feel it?',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO (lubien) [2200-12-12]: Too long... Can you feel it?',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// FIXME [2200-12-12] (lubien): Too long... Can you feel it?',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// Expire Condition [2200-12-12]: new term name',
      options: [
        {
          terms: ['Expire Condition'],
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// Expire Condition [2000-01-01]: new term name',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "// TODO [>2000]: We sure didn't past this version",
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [>1000]: partial version with > should use semver range semantics',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [find-up-simple@>1]: find-up-simple is 1.0.1 so >1 should not trigger',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:node@>22]: node engine is 22.x so >22 should not trigger',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [peer:eslint@>=11]: peer eslint floor is 10.x so >=11 should not trigger',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [peer:eslint@>10]: `>10` means `>=11.0.0`, so the 10.x floor should not trigger',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "// TODO [peer:does-not-exist@>=1]: a peer dependency we don't declare never triggers",
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [peer:find-up-simple@>=1]: `find-up-simple` is a dependency but not a peer dependency, so `peer:` never triggers',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [-find-up-simple]: We actually use this.',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "// TODO [+popura]: I think we won't need a broken package.",
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [+popura-2000-01-01]: A broken package with a date-like name.',
      options: [
        {
          checkDates: true,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "// TODO [semver@>1000]: Welp hopefully we won't get at that.",
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "// TODO [semver@>=1000]: Welp hopefully we won't get at that.",
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>=1.0.0]: we are using a pre-release',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>=1.0.0-gamma.1]: beta comes first from gamma',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>=1.0.0-beta.2]: we are in beta.1',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12, -find-up-simple]: Combo',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12, -find-up-simple, +popura]: Combo',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12, -find-up-simple, +popura, semver@>=1000]: Combo',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:node@>=100]: When we start supporting only >= 10',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2000-01-01]: Expired dates are ignored by default',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12]: Multiple\n\t\t// TODO [2200-12-12]: Lines',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '/*\n\t\t  * TODO [2200-12-12]: Yet\n\t\t  * TODO [engine:node@>=100]: Another\n\t\t  * TODO [+popura]: Way\n\t\t  */',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [invalid]',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [] might have [some] that [try [to trick] me]',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [but [it will]] [fallback] [[[ to the default ]]] rule [[',
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO ISSUE-123 fix later',
      options: [
        {
          allowWarningComments: false,
          ignore: ['ISSUE-\\d+'],
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [ISSUE-123] fix later',
      options: [
        {
          allowWarningComments: false,
          ignore: ['ISSUE-\\d+'],
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [1999-01-01, ISSUE-123] fix later',
      options: [
        {
          allowWarningComments: false,
          ignore: ['ISSUE-\\d+'],
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [Issue-123] fix later',
      options: [
        {
          allowWarningComments: false,
          ignore: ['[iI][sS][sS][uU][eE]-\\d+'],
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2001-01-01]: quite old',
      options: [
        {
          date: '2000-01-01',
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2000-01-01]: too old but ignored in all environments',
      options: [
        {
          checkDates: false,
          checkDatesOnPullRequests: true,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// eslint-disable-next-line unicorn/expiring-todo-comments\n\t\t\t\t   // TODO without a date',
      options: [
        {
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "/* eslint-disable unicorn/expiring-todo-comments */\n\t\t\t\t   // TODO without a date\n\t\t\t\t   // fixme [2000-01-01]: too old'\n\t\t\t\t   /* eslint-enable unicorn/expiring-todo-comments */",
      options: [
        {
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: "// ✅\n// TODO [2200-12-25]: Too long... Can you feel it?\n// FIXME [2200-12-25]: Too long... Can you feel it?\n\n// TODO (lubien) [2200-12-12]: You can add something before the arguments.\n// TODO @lubien [2200-12-12]: You can add something before the arguments.\n// FIXME [2200-12-25] (lubien): You can add something after the arguments, before the colon.\n// TODO [2200-12-12] No colon after argument.\n\n// TODO [+react]: Refactor this when we use React.\n// TODO [-lodash]: If we remove lodash we need to change this.\n\n// TODO [lodash@>10]: Lodash has a new way to do this; when we bump to its version let's use it.\n// TODO [lodash@>=10]: Lodash has a new way to do this; when we bump to its version let's use it.\n\n// TODO [2200-12-25, +popura, lodash@>10]: Combo.\n\n// TODO [peer:eslint@>=99]: When our minimum supported `eslint` reaches v99.\n\n// TODO [engine:node@>12]: When we bump to this Node version we can use import/export.\n\n/*\n * TODO [2200-12-25]: Yet\n * TODO [2200-12-25]: Another\n * TODO [2200-12-25]: Way\n */\n",
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
    },
  ],
  invalid: [
    {
      code: '// TODO [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '/*\n\t\t\t* TODO [2000-01-01]: Yet\n\t\t\t* TODO [2000-01-01]: Another\n\t\t\t* TODO [2000-01-01] Way\n\t\t\t*/',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Yet',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [],
        },
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Another',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [],
        },
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Way',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '/*\n\t\t\t* TODO [2000-01-01]: Invalid\n\t\t\t* TODO [2200-01-01]: Valid\n\t\t\t* TODO [2000-01-01]: Invalid\n\t\t\t*/',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Invalid',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [],
        },
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Invalid',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '/*\n\t\t\t* Something here\n\t\t\t* TODO [engine:node@>=8]: Invalid\n\t\t\t* Also something here\n\t\t\t*/',
      errors: [
        {
          messageId: 'unicorn/engineMatches',
          message: 'Due since Node.js version matched: node>=8. Invalid',
          line: 1,
          column: 1,
          endLine: 5,
          endColumn: 6,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// fixme [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// xxx [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// ToDo [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// fIxME [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// Todoist [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
          terms: ['Todoist'],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// Expire Condition [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
          terms: ['Expire Condition'],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// XxX [2000-01-01]: too old',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. too old',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12, 2200-12-12]: Multiple dates',
      errors: [
        {
          messageId: 'unicorn/avoidMultipleDates',
          message:
            'Avoid using multiple expiration dates: 2200-12-12, 2200-12-12. Multiple dates',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12, 2200-12-12]: Multiple dates are still invalid',
      errors: [
        {
          messageId: 'unicorn/avoidMultipleDates',
          message:
            'Avoid using multiple expiration dates: 2200-12-12, 2200-12-12. Multiple dates are still invalid',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 67,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [>1]: if your package.json version is >1',
      errors: [
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >1. if your package.json version is >1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [>1, >2]: multiple package versions',
      errors: [
        {
          messageId: 'unicorn/avoidMultiplePackageVersions',
          message:
            'Avoid using multiple package versions: >1, >2. multiple package versions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [>=1]: if your package.json version is >=1',
      errors: [
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >=1. if your package.json version is >=1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [+find-up-simple]: when you install `find-up-simple`',
      errors: [
        {
          messageId: 'unicorn/havePackage',
          message:
            'Due since find-up-simple was installed. when you install `find-up-simple`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [-popura]: when you uninstall `popura`',
      errors: [
        {
          messageId: 'unicorn/dontHavePackage',
          message: 'Due since popura was removed. when you uninstall `popura`',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [find-up-simple@>=1]: when `find-up-simple` version is >= 1',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: find-up-simple >= 1. when `find-up-simple` version is >= 1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 68,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:node@>=8]: when support is for node >= 8',
      errors: [
        {
          messageId: 'unicorn/engineMatches',
          message:
            'Due since Node.js version matched: node>=8. when support is for node >= 8',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [peer:eslint@>=9]: when the peer eslint floor reaches >= 9',
      errors: [
        {
          messageId: 'unicorn/peerVersionMatches',
          message:
            'Due since peer dependency version matched: eslint >= 9. when the peer eslint floor reaches >= 9',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 67,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [peer:eslint@>8]: when the peer eslint floor reaches > 8',
      errors: [
        {
          messageId: 'unicorn/peerVersionMatches',
          message:
            'Due since peer dependency version matched: eslint > 8. when the peer eslint floor reaches > 8',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 65,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [find-up-simple@>0.2.0]: when `find-up-simple` version is > 0.2.0',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: find-up-simple > 0.2.0. when `find-up-simple` version is > 0.2.0',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 74,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>=1.0.0-alfa.1]: when `@lubien/fixture-beta-package` version is >= 1.0.0-alfa.1',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: @lubien/fixture-beta-package >= 1.0.0-alfa.1. when `@lubien/fixture-beta-package` version is >= 1.0.0-alfa.1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 118,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>=1.0.0-beta.1]: when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.1',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: @lubien/fixture-beta-package >= 1.0.0-beta.1. when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 118,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>=1.0.0-beta.0]: when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.0',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: @lubien/fixture-beta-package >= 1.0.0-beta.0. when `@lubien/fixture-beta-package` version is >= 1.0.0-beta.0',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 118,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [@lubien/fixture-beta-package@>0.9]: when `@lubien/fixture-beta-package` prerelease version is > 0.9',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: @lubien/fixture-beta-package > 0.9. when `@lubien/fixture-beta-package` prerelease version is > 0.9',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 109,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [semver>1]: Missing @.',
      errors: [
        {
          messageId: 'unicorn/missingAtSymbol',
          message:
            "Missing '@' on TODO argument. On 'semver>1' use 'semver@>1'. Missing @.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [> 1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On '> 1' use '>1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [semver@> 1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'semver@> 1' use 'semver@>1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [semver @>1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'semver @>1' use 'semver@>1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 57,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [semver@>= 1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'semver@>= 1' use 'semver@>=1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 58,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [semver @>=1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'semver @>=1' use 'semver@>=1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 58,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:node @>=1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'engine:node @>=1' use 'engine:node@>=1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 63,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:node@>= 1]: Remove whitespace when it can fix.',
      errors: [
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'engine:node@>= 1' use 'engine:node@>=1'. Remove whitespace when it can fix.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 63,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO',
      errors: [
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO []',
      errors: [
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO []'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [no meaning at all]',
      errors: [
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO [no meaning at all]'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [] might have [some] that [try [to trick] me]',
      errors: [
        {
          messageId: 'unexpectedComment',
          message:
            "Unexpected 'todo': 'TODO [] might have [some] that [try [to...'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [but [it will]] [fallback] [[[ to the default ]]] rule [[[',
      errors: [
        {
          messageId: 'unexpectedComment',
          message:
            "Unexpected 'todo': 'TODO [but [it will]] [fallback] [[[ to...'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 67,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:npm@>=10000]: Unsupported engine',
      errors: [
        {
          messageId: 'unexpectedComment',
          message:
            "Unexpected 'todo': 'TODO [engine:npm@>=10000]: Unsupported...'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [engine:somethingrandom@>=10000]: Unsupported engine',
      errors: [
        {
          messageId: 'unexpectedComment',
          message:
            "Unexpected 'todo': 'TODO [engine:somethingrandom@>=10000]:...'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
          suggestions: [],
        },
      ],
      options: [
        {
          allowWarningComments: false,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2000-01-01, >1]: Combine date with package version',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message:
            'Past due date: 2000-01-01. Combine date with package version',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
          suggestions: [],
        },
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >1. Combine date with package version',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2200-12-12, >1, 2200-12-12, >2]: Multiple dates and package versions',
      errors: [
        {
          messageId: 'unicorn/avoidMultipleDates',
          message:
            'Avoid using multiple expiration dates: 2200-12-12, 2200-12-12. Multiple dates and package versions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 78,
          suggestions: [],
        },
        {
          messageId: 'unicorn/avoidMultiplePackageVersions',
          message:
            'Avoid using multiple package versions: >1, >2. Multiple dates and package versions',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 78,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [-popura, find-up-simple@>=1]: Combine not having a package with version match',
      errors: [
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since popura was removed. Combine not having a package with version match',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 87,
          suggestions: [],
        },
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: find-up-simple >= 1. Combine not having a package with version match',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 87,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [+find-up-simple, -popura]: Combine presence/absence of packages',
      errors: [
        {
          messageId: 'unicorn/havePackage',
          message:
            'Due since find-up-simple was installed. Combine presence/absence of packages',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since popura was removed. Combine presence/absence of packages',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// Expire Condition [2000-01-01, semver>1]: Expired TODO and missing symbol',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Expired TODO and missing symbol',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 76,
          suggestions: [],
        },
        {
          messageId: 'unicorn/missingAtSymbol',
          message:
            "Missing '@' on TODO argument. On 'semver>1' use 'semver@>1'. Expired TODO and missing symbol",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 76,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
          terms: ['Expire Condition'],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [semver @>=1, -popura]: Package uninstalled and whitespace error',
      errors: [
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since popura was removed. Package uninstalled and whitespace error',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'semver @>=1' use 'semver@>=1'. Package uninstalled and whitespace error",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// HUGETODO [semver @>=1, engine:node@>=8, 2000-01-01, -popura, >1, +find-up-simple, find-up-simple@>=1, peer:eslint@>=9]: Big mix',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2000-01-01. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/reachedPackageVersion',
          message: 'Past due package version: >1. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/dontHavePackage',
          message: 'Due since popura was removed. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/havePackage',
          message: 'Due since find-up-simple was installed. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: find-up-simple >= 1. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/peerVersionMatches',
          message:
            'Due since peer dependency version matched: eslint >= 9. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/engineMatches',
          message: 'Due since Node.js version matched: node>=8. Big mix',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
        {
          messageId: 'unicorn/removeWhitespaces',
          message:
            "Avoid using whitespace on TODO argument. On 'semver @>=1' use 'semver@>=1'. Big mix",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 131,
          suggestions: [],
        },
      ],
      options: [
        {
          checkDates: true,
          checkDatesOnPullRequests: true,
          terms: ['HUGETODO'],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [ISSUE-123] fix later',
      options: [
        {
          allowWarningComments: false,
          ignore: [],
        },
      ],
      errors: [
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO [ISSUE-123] fix later'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '\n\t\t\t// TODO fix later\n\t\t\t// TODO ISSUE-123 fix later\n\t\t\t',
      options: [
        {
          allowWarningComments: false,
          ignore: ['[iI][sS][sS][uU][eE]-\\d+'],
        },
      ],
      errors: [
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO fix later'.",
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '/*\n\t\t\tTODO Invalid\n\t\t\tTODO ISSUE-123 Valid\n\t\t\t*/',
      options: [
        {
          allowWarningComments: false,
          ignore: ['[iI][sS][sS][uU][eE]-\\d+'],
        },
      ],
      errors: [
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO Invalid'.",
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 6,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2999-12-01]: Y3K bug',
      options: [
        {
          date: '3000-01-01',
          checkDates: true,
          checkDatesOnPullRequests: true,
        },
      ],
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: 'Past due date: 2999-12-01. Y3K bug',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: null,
      filename: 'fixtures/expiring-todo-comments/input.ts',
    },
    {
      code: '// TODO [2019-11-15]: Refactor this code before the sprint ends.\n// TODO (@lubien) [2019-07-18]: When John delivers his code. I can reuse it.\n// TODO [2019-08-10]: I must refactor this for sure before I deliver.\n',
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message:
            'Past due date: 2019-11-15. Refactor this code before the sprint ends.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 65,
          suggestions: [],
        },
        {
          messageId: 'unicorn/expiredTodo',
          message:
            'Past due date: 2019-07-18. When John delivers his code. I can reuse it.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 77,
          suggestions: [],
        },
        {
          messageId: 'unicorn/expiredTodo',
          message:
            'Past due date: 2019-08-10. I must refactor this for sure before I deliver.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 70,
          suggestions: [],
        },
      ],
    },
    {
      code: '// TODO [>=1.0.0]: I should work around this when we reach v1.\n// TODO (@lubien) [>0]: For now this is fine but for a stable version we must refactor.\n// FIXME [>10]: This feature is deprecated and should be removed from the next major version.\n',
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >=1.0.0. I should work around this when we reach v1.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 63,
          suggestions: [],
        },
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >0. For now this is fine but for a stable version we must refactor.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 88,
          suggestions: [],
        },
      ],
    },
    {
      code: '// TODO [engine:node@>=8]: We can use async/await now.\n// FIXME [engine:node@>=20.0.0]: Hey, node can use import/export now, we should refactor.\n',
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/engineMatches',
          message:
            'Due since Node.js version matched: node>=8. We can use async/await now.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
          suggestions: [],
        },
      ],
    },
    {
      code: "// TODO [-vue-function-api]: When we remove `vue-function-api` we should refactor this.\n// FIXME [+read-pkg]: If we use this package we don't need to use this function below.\n// XXX @lubien [+react, -jquery]: We can use React for this widget instead of jQuery.\n",
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since vue-function-api was removed. When we remove `vue-function-api` we should refactor this.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 88,
          suggestions: [],
        },
        {
          messageId: 'unicorn/havePackage',
          message:
            "Due since read-pkg was installed. If we use this package we don't need to use this function below.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 87,
          suggestions: [],
        },
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since jquery was removed. We can use React for this widget instead of jQuery.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 86,
          suggestions: [],
        },
      ],
    },
    {
      code: "// TODO [vue@>=3]: Refactor to function API when it's stable.\n// FIXME [cerebro@>0.10.0]: This is a quickfix until cerebro fixes this.\n// XXX [popura@>=2.0.0]: This API is deprecated so we should not use it by then.\n",
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/versionMatches',
          message:
            "Due since package version matched: vue >= 3. Refactor to function API when it's stable.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 62,
          suggestions: [],
        },
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: cerebro > 0.10.0. This is a quickfix until cerebro fixes this.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 73,
          suggestions: [],
        },
      ],
    },
    {
      code: '// TODO [peer:eslint@>=9]: Drop the `CLIEngine` fallback once we require ESLint 9.\n',
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/peerVersionMatches',
          message:
            'Due since peer dependency version matched: eslint >= 9. Drop the `CLIEngine` fallback once we require ESLint 9.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 83,
          suggestions: [],
        },
      ],
    },
    {
      code: '// TODO [+react, -jquery]: We can use React for this widget instead of jQuery.\n// TODO [2019-07-15, +react]: Refactor this if we install React or if we reach that date.\n// TODO [-vue-function-api, vue@>=3]: Now we should use Vue native function API.\n',
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since jquery was removed. We can use React for this widget instead of jQuery.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 79,
          suggestions: [],
        },
        {
          messageId: 'unicorn/expiredTodo',
          message:
            'Past due date: 2019-07-15. Refactor this if we install React or if we reach that date.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 90,
          suggestions: [],
        },
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            'Due since vue-function-api was removed. Now we should use Vue native function API.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 81,
          suggestions: [],
        },
        {
          messageId: 'unicorn/versionMatches',
          message:
            'Due since package version matched: vue >= 3. Now we should use Vue native function API.',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 81,
          suggestions: [],
        },
      ],
    },
    {
      code: '/*\n * We should really make this code better.\n * When we support Node.js 12 we can refactor imports.\n * And we also can do [x], [y], [z].\n * TODO [engine:node@>=12]: Use import/export.\n */\n\n/*\n * This code would be so easy if we used `popura` package helpers.\n * When you can, install `popura`, use it and remove dead code.\n * TODO [+popura]: Refactor to use `popura`.\n *\n * You can also use `popura-cli` since we want help on [feature].\n * TODO [+popura-cli]: Document how to use `popura-cli`.\n */\n',
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/engineMatches',
          message:
            'Due since Node.js version matched: node>=12. Use import/export.',
          line: 1,
          column: 1,
          endLine: 6,
          endColumn: 4,
          suggestions: [],
        },
      ],
    },
    {
      code: "// ❌\n// With `checkDates: true`\n// TODO [2000-01-01]: I'll fix this next week.\n// TODO [2000-01-01, 2001-01-01]: Multiple dates won't work.\n\n// TODO [>1]: If your package.json version is > 1.\n// TODO [>=1]: If your package.json version is >= 1.\n// TODO [>1, >2]: Multiple package versions won't work.\n\n// TODO [+already-have-pkg]: Since we already have it, this reports.\n// TODO [-we-dont-have-this-package]: Since we don't have, trigger a report.\n\n// TODO [read-pkg@>1]: When `read-pkg` version is > 1 don't forget to do this.\n// TODO [read-pkg@>=5.1.1]: When `read-pkg` version is >= 5.1.1 don't forget to do that.\n\n// TODO [peer:eslint@>=8]: Whoops, our minimum supported `eslint` is already >= 8.\n\n// TODO [engine:node@>=8]: Whoops, we are already supporting it!\n\n// TODO: Add unicorns.\n",
      options: [
        {
          date: '2026-09-30',
          checkDates: true,
          checkDatesOnPullRequests: true,
          allowWarningComments: false,
        },
      ],
      filename: 'fixtures/expiring-todo-comments/docs/input.ts',
      errors: [
        {
          messageId: 'unicorn/expiredTodo',
          message: "Past due date: 2000-01-01. I'll fix this next week.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 47,
          suggestions: [],
        },
        {
          messageId: 'unicorn/avoidMultipleDates',
          message:
            "Avoid using multiple expiration dates: 2000-01-01, 2001-01-01. Multiple dates won't work.",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 61,
          suggestions: [],
        },
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >1. If your package.json version is > 1.',
          line: 6,
          column: 1,
          endLine: 6,
          endColumn: 51,
          suggestions: [],
        },
        {
          messageId: 'unicorn/reachedPackageVersion',
          message:
            'Past due package version: >=1. If your package.json version is >= 1.',
          line: 7,
          column: 1,
          endLine: 7,
          endColumn: 53,
          suggestions: [],
        },
        {
          messageId: 'unicorn/avoidMultiplePackageVersions',
          message:
            "Avoid using multiple package versions: >1, >2. Multiple package versions won't work.",
          line: 8,
          column: 1,
          endLine: 8,
          endColumn: 56,
          suggestions: [],
        },
        {
          messageId: 'unicorn/havePackage',
          message:
            'Due since already-have-pkg was installed. Since we already have it, this reports.',
          line: 10,
          column: 1,
          endLine: 10,
          endColumn: 69,
          suggestions: [],
        },
        {
          messageId: 'unicorn/dontHavePackage',
          message:
            "Due since we-dont-have-this-package was removed. Since we don't have, trigger a report.",
          line: 11,
          column: 1,
          endLine: 11,
          endColumn: 77,
          suggestions: [],
        },
        {
          messageId: 'unicorn/versionMatches',
          message:
            "Due since package version matched: read-pkg > 1. When `read-pkg` version is > 1 don't forget to do this.",
          line: 13,
          column: 1,
          endLine: 13,
          endColumn: 79,
          suggestions: [],
        },
        {
          messageId: 'unicorn/versionMatches',
          message:
            "Due since package version matched: read-pkg >= 5.1.1. When `read-pkg` version is >= 5.1.1 don't forget to do that.",
          line: 14,
          column: 1,
          endLine: 14,
          endColumn: 89,
          suggestions: [],
        },
        {
          messageId: 'unicorn/peerVersionMatches',
          message:
            'Due since peer dependency version matched: eslint >= 8. Whoops, our minimum supported `eslint` is already >= 8.',
          line: 16,
          column: 1,
          endLine: 16,
          endColumn: 83,
          suggestions: [],
        },
        {
          messageId: 'unicorn/engineMatches',
          message:
            'Due since Node.js version matched: node>=8. Whoops, we are already supporting it!',
          line: 18,
          column: 1,
          endLine: 18,
          endColumn: 65,
          suggestions: [],
        },
        {
          messageId: 'unexpectedComment',
          message: "Unexpected 'todo': 'TODO: Add unicorns.'.",
          line: 20,
          column: 1,
          endLine: 20,
          endColumn: 23,
          suggestions: [],
        },
      ],
    },
    {
      code: '\n\t\t// TODO [eslint@>9]: Drop fallback.\n\t\t// TODO [prettier@>3]: Drop dev fallback.\n\t\t// TODO [peer:eslint@>9]: Drop peer fallback.\n\t\t// TODO [+eslint]: Presence checks still work.\n\t',
      filename: 'fixtures/expiring-todo-comments/catalog/input.ts',
      errors: [
        {
          messageId: 'unicorn/unsupportedCatalogProtocol',
          message:
            'Cannot check dependency version because eslint uses the unsupported `catalog:` protocol. Drop fallback.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 38,
          suggestions: [],
        },
        {
          messageId: 'unicorn/unsupportedCatalogProtocol',
          message:
            'Cannot check dependency version because prettier uses the unsupported `catalog:` protocol. Drop dev fallback.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 44,
          suggestions: [],
        },
        {
          messageId: 'unicorn/unsupportedCatalogProtocol',
          message:
            'Cannot check peer dependency version because eslint uses the unsupported `catalog:` protocol. Drop peer fallback.',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 48,
          suggestions: [],
        },
        {
          messageId: 'unicorn/havePackage',
          message:
            'Due since eslint was installed. Presence checks still work.',
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 49,
          suggestions: [],
        },
      ],
    },
  ],
});
