import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-restricted-imports', {
  valid: [
    // No options
    'import os from "os";',
    // String option — different module
    { code: 'import os from "os";', options: ['osx'] as any },
    { code: 'import fs from "fs";', options: ['crypto'] as any },
    {
      code: 'import path from "path";',
      options: ['crypto', 'stream', 'os'] as any,
    },
    // Side-effect import not matching
    { code: 'import "foo"', options: ['crypto'] as any },
    // Subpath not matching parent
    { code: 'import "foo/bar";', options: ['foo'] as any },
    // Paths option
    {
      code: 'import withPaths from "foo/bar";',
      options: { paths: ['foo', 'bar'] },
    },
    // Patterns option
    {
      code: 'import withPatterns from "foo/bar";',
      options: { patterns: ['foo/c*'] },
    },
    // Relative path not matching
    { code: "import foo from 'foo';", options: ['../foo'] as any },
    // Gitignore negation
    {
      code: 'import withGitignores from "foo/bar";',
      options: { patterns: ['foo/*', '!foo/bar'] },
    },
    // Pattern group with negation
    {
      code: 'import withPatterns from "foo/bar";',
      options: {
        patterns: [
          { group: ['foo/*', '!foo/bar'], message: 'foo is forbidden' },
        ],
      },
    },
    // Case sensitive not matching
    {
      code: "import x from 'foo';",
      options: { patterns: [{ group: ['FOO'], caseSensitive: true }] },
    },
    // importNames — allowed name
    {
      code: 'import { AllowedObject } from "foo";',
      options: { paths: [{ name: 'foo', importNames: ['DisallowedObject'] }] },
    },
    // Default import not restricted by named importNames
    {
      code: 'import DisallowedObject from "foo";',
      options: { paths: [{ name: 'foo', importNames: ['DisallowedObject'] }] },
    },
    // Aliased import — alias matches restriction name, but source name doesn't
    {
      code: 'import { AllowedObject as DisallowedObject } from "foo";',
      options: { paths: [{ name: 'foo', importNames: ['DisallowedObject'] }] },
    },
    // Side-effect import with importNames
    {
      code: 'import "foo";',
      options: { paths: [{ name: 'foo', importNames: ['DisallowedObject'] }] },
    },
    // Export * from non-restricted
    { code: 'export * from "foo";', options: ['bar'] as any },
    // allowImportNames
    {
      code: 'import { AllowedObject } from "foo";',
      options: {
        paths: [{ name: 'foo', allowImportNames: ['AllowedObject'] }],
      },
    },
    // allowImportNames in patterns
    {
      code: "import { foo } from 'foo';",
      options: { patterns: [{ group: ['foo'], allowImportNames: ['foo'] }] },
    },
    // Export allowImportNames
    {
      code: "export { bar } from 'foo';",
      options: { paths: [{ name: 'foo', allowImportNames: ['bar'] }] },
    },
    // Regex not matching
    {
      code: 'import x from "foo/bar";',
      options: { patterns: [{ regex: 'foo/baz' }] },
    },
    // Regex case sensitive not matching
    {
      code: "import x from 'foo';",
      options: { patterns: [{ regex: 'FOO', caseSensitive: true }] },
    },
    // Regex negative lookahead: "foo/bar" NOT matched by "foo/(?!bar)"
    {
      code: 'import x from "foo/bar";',
      options: { patterns: [{ regex: 'foo/(?!bar)' }] },
    },
    // Default import not reported by importNamePattern
    {
      code: "import Foo from 'foo';",
      options: { patterns: [{ group: ['foo'], importNamePattern: '^Foo' }] },
    },
    // Named import not matching importNamePattern
    {
      code: "import { Bar } from 'foo';",
      options: { patterns: [{ group: ['foo'], importNamePattern: '^Foo' }] },
    },
    // allowImportNamePattern matching
    {
      code: "import { Foo } from 'foo';",
      options: {
        patterns: [{ group: ['foo'], allowImportNamePattern: '^Foo' }],
      },
    },
    // Pattern importNames — different name
    {
      code: "import { Bar } from '../../my/relative-module';",
      options: {
        patterns: [{ group: ['**/my/relative-module'], importNames: ['Foo'] }],
      },
    },
    // TypeScript: type import allowed
    {
      code: "import type foo from 'import-foo';",
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
    },
    {
      code: "import type { Bar } from 'import-foo';",
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
    },
    {
      code: "export type { Bar } from 'import-foo';",
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
    },
    // TypeScript: type-only pattern allowed
    {
      code: "import type { Bar } from 'import-foo';",
      options: {
        patterns: [{ group: ['import-foo'], allowTypeImports: true }],
      },
    },
    // export type * allowed with allowTypeImports
    {
      code: 'export type * from "foo";',
      options: { paths: [{ name: 'foo', allowTypeImports: true }] },
    },
    // import type = require() allowed with pattern
    {
      code: 'import type fs = require("fs");',
      options: { patterns: [{ group: ['f*'], allowTypeImports: true }] },
    },
    // individual type specifier with importNames + allowTypeImports
    {
      code: "import { type Bar } from 'import-foo';",
      options: {
        paths: [
          { name: 'import-foo', importNames: ['Bar'], allowTypeImports: true },
        ],
      },
    },
    {
      code: "export { type Bar } from 'import-foo';",
      options: {
        paths: [
          { name: 'import-foo', importNames: ['Bar'], allowTypeImports: true },
        ],
      },
    },
    // all specifiers individually type → whole import is type-only
    {
      code: 'import { type A, type B } from "import-foo";',
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
    },
    // export with all specifiers individually type
    {
      code: 'export { type A, type B } from "import-foo";',
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
    },
    // Pattern importNames + allowTypeImports: type specifier skipped
    {
      code: 'import { Bar, type Baz } from "import/private/bar";',
      options: {
        patterns: [
          {
            group: ['import/private/*'],
            importNames: ['Baz'],
            allowTypeImports: true,
          },
        ],
      },
    },
    // Pattern allowImportNames + allowTypeImports
    {
      code: 'import { Foo, type Bar } from "import/private/bar";',
      options: {
        patterns: [
          {
            group: ['import/private/*'],
            allowImportNames: ['Foo'],
            allowTypeImports: true,
          },
        ],
      },
    },
    // export { bar, type baz } with path allowImportNames + allowTypeImports
    {
      code: 'export { bar, type baz } from "import-foo";',
      options: {
        paths: [
          {
            name: 'import-foo',
            allowImportNames: ['bar'],
            allowTypeImports: true,
          },
        ],
      },
    },
    // import = require() with importNames: no specifiers → no importName error
    {
      code: "import foo = require('foo');",
      options: { paths: [{ name: 'foo', importNames: ['foo'] }] },
    },
    // @scoped package in pattern with negation
    {
      code: "import { foo } from '@app/api/enums';",
      options: { patterns: [{ group: ['@app/api/*', '!@app/api/enums'] }] },
    },
    // String literal specifier: 'AllowedObject' not in restriction list
    {
      code: 'import { \'AllowedObject\' as bar } from "foo";',
      options: { paths: [{ name: 'foo', importNames: ['DisallowedObject'] }] },
    },
    // String literal ' ' (space) is not '' (empty)
    {
      code: 'import { \' \' as bar } from "foo";',
      options: { paths: [{ name: 'foo', importNames: [''] }] },
    },
    // export with string literal alias: source 'AllowedObject' not restricted
    {
      code: 'export { \'AllowedObject\' as DisallowedObject } from "foo";',
      options: { paths: [{ name: 'foo', importNames: ['DisallowedObject'] }] },
    },
  ],
  invalid: [
    // Basic string restriction
    {
      code: 'import "fs"',
      options: ['fs'] as any,
      errors: [{ messageId: 'path' }],
    },
    {
      code: 'import os from "os ";',
      options: ['fs', 'crypto ', 'stream', 'os'] as any,
      errors: [{ messageId: 'path' }],
    },
    // Subpath restriction
    {
      code: 'import "foo/bar";',
      options: ['foo/bar'] as any,
      errors: [{ messageId: 'path' }],
    },
    // Path option
    {
      code: 'import withPaths from "foo/bar";',
      options: { paths: ['foo/bar'] },
      errors: [{ messageId: 'path' }],
    },
    // Pattern matching
    {
      code: 'import withPatterns from "foo/bar";',
      options: { patterns: ['foo'] },
      errors: [{ messageId: 'patterns' }],
    },
    {
      code: 'import withPatterns from "foo/bar";',
      options: { patterns: ['bar'] },
      errors: [{ messageId: 'patterns' }],
    },
    // Pattern group with custom message
    {
      code: 'import withPatterns from "foo/baz";',
      options: {
        patterns: [
          {
            group: ['foo/*', '!foo/bar'],
            message: 'foo is forbidden, use foo/bar instead',
          },
        ],
      },
      errors: [{ messageId: 'patternWithCustomMessage' }],
    },
    // Case insensitive pattern (default)
    {
      code: "import x from 'foo';",
      options: { patterns: [{ group: ['FOO'] }] },
      errors: [{ messageId: 'patterns' }],
    },
    // Gitignore negation: not negated → matches
    {
      code: 'import x from "foo/bar";',
      options: { patterns: ['foo/*', '!foo/baz'] },
      errors: [{ messageId: 'patterns' }],
    },
    // Export * restriction
    {
      code: 'export * from "fs";',
      options: ['fs'] as any,
      errors: [{ messageId: 'path' }],
    },
    // Export * as ns restriction
    {
      code: 'export * as ns from "fs";',
      options: ['fs'] as any,
      errors: [{ messageId: 'path' }],
    },
    // Export named restriction
    {
      code: 'export {a} from "fs";',
      options: ['fs'] as any,
      errors: [{ messageId: 'path' }],
    },
    // Export named with importNames
    {
      code: 'export {foo as b} from "fs";',
      options: {
        paths: [
          { name: 'fs', importNames: ['foo'], message: "Don't import 'foo'." },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Custom message on path
    {
      code: 'import x from "foo";',
      options: {
        paths: [{ name: 'foo', message: "Please import from 'bar' instead." }],
      },
      errors: [{ messageId: 'pathWithCustomMessage' }],
    },
    // Default import restriction via importNames: ["default"]
    {
      code: 'import DisallowedObject from "foo";',
      options: {
        paths: [
          {
            name: 'foo',
            importNames: ['default'],
            message:
              "Please import the default import of 'foo' from /bar/ instead.",
          },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Star import with importNames
    {
      code: 'import * as All from "foo";',
      options: {
        paths: [
          {
            name: 'foo',
            importNames: ['DisallowedObject'],
            message: "Please import 'DisallowedObject' from /bar/ instead.",
          },
        ],
      },
      errors: [{ messageId: 'everythingWithCustomMessage' }],
    },
    // Named import — restricted name
    {
      code: 'import { DisallowedObject } from "foo";',
      options: {
        paths: [
          {
            name: 'foo',
            importNames: ['DisallowedObject'],
            message: "Please import 'DisallowedObject' from /bar/ instead.",
          },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Aliased import — source name restricted
    {
      code: 'import { DisallowedObject as AllowedObject } from "foo";',
      options: {
        paths: [
          {
            name: 'foo',
            importNames: ['DisallowedObject'],
            message: "Please import 'DisallowedObject' from /bar/ instead.",
          },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Multiple restricted names — two errors
    {
      code: 'import { DisallowedObjectOne, DisallowedObjectTwo, AllowedObject } from "foo";',
      options: {
        paths: [
          {
            name: 'foo',
            importNames: ['DisallowedObjectOne', 'DisallowedObjectTwo'],
          },
        ],
      },
      errors: [{ messageId: 'importName' }, { messageId: 'importName' }],
    },
    // Relative path
    {
      code: "import relative from '../foo';",
      options: ['../foo'] as any,
      errors: [{ messageId: 'path' }],
    },
    // Regex pattern matching
    {
      code: 'import x from "foo/baz";',
      options: {
        patterns: [
          { regex: 'foo/baz', message: 'foo is forbidden, use bar instead' },
        ],
      },
      errors: [{ messageId: 'patternWithCustomMessage' }],
    },
    // Regex case insensitive (default)
    {
      code: 'import x from "foo";',
      options: { patterns: [{ regex: 'FOO' }] },
      errors: [{ messageId: 'patterns' }],
    },
    // Regex negative lookahead: "foo/baz" matched by "foo/(?!bar)"
    {
      code: 'import x from "foo/baz";',
      options: {
        patterns: [
          {
            regex: 'foo/(?!bar)',
            message: 'foo is forbidden, use bar instead',
          },
        ],
      },
      errors: [{ messageId: 'patternWithCustomMessage' }],
    },
    // Pattern importNames
    {
      code: "import { Foo } from '../../my/relative-module';",
      options: {
        patterns: [{ group: ['**/my/relative-module'], importNames: ['Foo'] }],
      },
      errors: [{ messageId: 'patternAndImportName' }],
    },
    // Pattern importNamePattern
    {
      code: "import { Foo } from 'foo';",
      options: { patterns: [{ group: ['foo'], importNamePattern: '^Foo' }] },
      errors: [{ messageId: 'patternAndImportName' }],
    },
    // Star import with pattern importNames
    {
      code: "import * as All from 'foo-bar-baz';",
      options: { patterns: [{ group: ['**'], importNames: ['Foo', 'Bar'] }] },
      errors: [{ messageId: 'patternAndEverything' }],
    },
    // allowImportNames — disallowed name
    {
      code: 'import { AllowedObject, DisallowedObject } from "foo";',
      options: {
        paths: [{ name: 'foo', allowImportNames: ['AllowedObject'] }],
      },
      errors: [{ messageId: 'allowedImportName' }],
    },
    // allowImportNames with star import
    {
      code: 'import * as AllowedObject from "foo";',
      options: {
        paths: [{ name: 'foo', allowImportNames: ['AllowedObject'] }],
      },
      errors: [{ messageId: 'everythingWithAllowImportNames' }],
    },
    // Pattern allowImportNames — disallowed name
    {
      code: 'import { AllowedObject, DisallowedObject } from "foo";',
      options: {
        patterns: [{ group: ['foo'], allowImportNames: ['AllowedObject'] }],
      },
      errors: [{ messageId: 'allowedImportName' }],
    },
    // Pattern allowImportNamePattern — non-matching name
    {
      code: "import { Bar } from 'foo';",
      options: {
        patterns: [{ group: ['foo'], allowImportNamePattern: '^Foo' }],
      },
      errors: [{ messageId: 'allowedImportNamePattern' }],
    },
    // Export with allowImportNamePattern
    {
      code: "export { Bar } from 'foo';",
      options: {
        patterns: [{ group: ['foo'], allowImportNamePattern: '^Foo' }],
      },
      errors: [{ messageId: 'allowedImportNamePattern' }],
    },
    // Star with allowImportNamePattern
    {
      code: "import * as Foo from 'foo';",
      options: {
        patterns: [{ group: ['foo'], allowImportNamePattern: '^Foo' }],
      },
      errors: [{ messageId: 'everythingWithAllowedImportNamePattern' }],
    },
    // Star with importNamePattern
    {
      code: "import * as Foo from 'foo';",
      options: { patterns: [{ group: ['foo'], importNamePattern: '^Foo' }] },
      errors: [{ messageId: 'patternAndEverythingWithRegexImportName' }],
    },
    // TypeScript: regular import fails even with allowTypeImports
    {
      code: "import foo from 'import-foo';",
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
      errors: [{ messageId: 'path' }],
    },
    // TypeScript: import equals with external module reference
    {
      code: "import foo = require('import-foo');",
      options: { paths: [{ name: 'import-foo' }] },
      errors: [{ messageId: 'path' }],
    },
    // TypeScript: type import with restricted importNames still reported
    {
      code: 'import type { bar } from "mod";',
      options: {
        paths: [
          {
            name: 'mod',
            importNames: ['bar'],
            message: "don't import 'bar' at all",
          },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // TypeScript: mixed type and regular imports with allowTypeImports
    {
      code: 'import { Bar, type Baz } from "import-foo";',
      options: {
        paths: [
          {
            name: 'import-foo',
            importNames: ['Bar', 'Baz'],
            allowTypeImports: true,
          },
        ],
      },
      errors: [{ messageId: 'importName' }],
    },
    // Side-effect import with allowTypeImports: still restricted
    {
      code: "import 'import-foo';",
      options: { paths: [{ name: 'import-foo', allowTypeImports: true }] },
      errors: [{ messageId: 'path' }],
    },
    // import = require() with allowTypeImports: regular require restricted
    {
      code: "import foo = require('import-foo');",
      options: {
        paths: [
          {
            name: 'import-foo',
            allowTypeImports: true,
            message: 'Please use import-bar instead.',
          },
        ],
      },
      errors: [{ messageId: 'pathWithCustomMessage' }],
    },
    // export { Bar, type Baz } with allowTypeImports: only Bar reported
    {
      code: "export { Bar, type Baz } from 'import-foo';",
      options: {
        paths: [
          {
            name: 'import-foo',
            importNames: ['Bar', 'Baz'],
            allowTypeImports: true,
          },
        ],
      },
      errors: [{ messageId: 'importName' }],
    },
    // export type * + export * with allowTypeImports: only regular export * reported
    {
      code: 'export type * from "foo";\nexport * from "foo";',
      options: { paths: [{ name: 'foo', allowTypeImports: true }] },
      errors: [{ messageId: 'path' }],
    },
    // empty export {} with allowTypeImports: non-type export restricted
    {
      code: 'export { } from "mod";\nexport type { } from "mod";',
      options: { paths: [{ name: 'mod', allowTypeImports: true }] },
      errors: [{ messageId: 'path' }],
    },
    // allowTypeImports: false explicitly → type import restricted too
    {
      code: "import { Foo } from 'restricted-path';\nimport type { Bar } from 'restricted-path';",
      options: {
        paths: [
          {
            name: 'restricted-path',
            allowTypeImports: false,
            message: 'This import is restricted.',
          },
        ],
      },
      errors: [
        { messageId: 'pathWithCustomMessage' },
        { messageId: 'pathWithCustomMessage' },
      ],
    },
    // import = require() with pattern
    {
      code: 'import fs = require("fs");',
      options: { patterns: [{ group: ['f*'] }] },
      errors: [{ messageId: 'patterns' }],
    },
    // type-only export matched by pattern (no allowTypeImports)
    {
      code: "export type { InvalidTestCase } from '@typescript-eslint/utils/dist/ts-eslint';",
      options: { patterns: ['@typescript-eslint/utils/dist/*'] },
      errors: [{ messageId: 'patterns' }],
    },
    // allowImportNames + allowTypeImports: non-allowed name restricted
    {
      code: 'import { baz } from "foo";',
      options: {
        paths: [
          { name: 'foo', allowImportNames: ['bar'], allowTypeImports: true },
        ],
      },
      errors: [{ messageId: 'allowedImportName' }],
    },
    // Pattern with allowTypeImports + importNames: regular specifier reported
    {
      code: 'import { Foo, type Bar } from "import/private/bar";',
      options: {
        patterns: [
          {
            group: ['import/private/*'],
            importNames: ['Foo', 'Bar'],
            allowTypeImports: true,
          },
        ],
      },
      errors: [{ messageId: 'patternAndImportName' }],
    },
    // export type { bar } with two entries: one allowTypeImports, one not
    {
      code: 'export type { bar } from "mod";',
      options: {
        paths: [
          {
            name: 'mod',
            importNames: ['foo'],
            allowTypeImports: true,
            message: "import 'foo' only as type",
          },
          {
            name: 'mod',
            importNames: ['bar'],
            message: "don't import 'bar' at all",
          },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Both paths AND patterns matching same import → errors from both
    {
      code: "import foo from 'import1/private/foo';",
      options: {
        paths: ['import1/private/foo'],
        patterns: ['import1/private/*'],
      },
      errors: [{ messageId: 'path' }, { messageId: 'patterns' }],
    },
    // @scoped package matched by pattern
    {
      code: "import { foo } from '@app/api/bar';",
      options: { patterns: [{ group: ['@app/api/*', '!@app/api/enums'] }] },
      errors: [{ messageId: 'patterns' }],
    },
    // @scoped with regex + importNamePattern
    {
      code: "import { Foo_Enum } from '@app/api';",
      options: {
        patterns: [{ regex: '@app/api$', importNamePattern: '_Enum$' }],
      },
      errors: [{ messageId: 'patternAndImportName' }],
    },
    // String literal specifier: export { 'foo' as b } from "fs"
    {
      code: 'export { \'foo\' as b } from "fs";',
      options: {
        paths: [
          { name: 'fs', importNames: ['foo'], message: "Don't import 'foo'." },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // String literal specifier: export { 'foo' } from "fs" (no alias)
    {
      code: 'export { \'foo\' } from "fs";',
      options: {
        paths: [
          { name: 'fs', importNames: ['foo'], message: "Don't import 'foo'." },
        ],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Empty string import name: export { '' } from "fs"
    {
      code: 'export { \'\' } from "fs";',
      options: {
        paths: [{ name: 'fs', importNames: [''], message: "Don't import ''." }],
      },
      errors: [{ messageId: 'importNameWithCustomMessage' }],
    },
    // Complex glob: **/*-*-baz matching foo-bar-baz
    {
      code: "import * as All from 'foo-bar-baz';",
      options: {
        patterns: [
          {
            group: ['**/*-*-baz'],
            importNames: ['Foo', 'Bar'],
            message: "Use only 'Baz'.",
          },
        ],
      },
      errors: [{ messageId: 'patternAndEverythingWithCustomMessage' }],
    },
    // Pattern group with multiple exact paths
    {
      code: 'import x from "foo/baz";',
      options: {
        patterns: [
          {
            group: ['foo/bar', 'foo/baz'],
            message: 'some foo subimports are restricted',
          },
        ],
      },
      errors: [{ messageId: 'patternWithCustomMessage' }],
    },
  ],
});
