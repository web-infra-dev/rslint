import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-duplicate-imports', {
  valid: [
    // ---- ESLint upstream — JS suite ----
    `import os from "os";
import fs from "fs";`,
    `import { merge } from "lodash-es";`,
    `import _, { merge } from "lodash-es";`,
    `import * as Foobar from "async";`,
    `import "foo"`,
    `import os from "os";
export { something } from "os";`,
    `import * as bar from "os";
import { baz } from "os";`,
    `import foo, * as bar from "os";
import { baz } from "os";`,
    `import foo, { bar } from "os";
import * as baz from "os";`,
    {
      code: `import os from "os";
export { hello } from "hello";`,
      options: { includeExports: true },
    },
    {
      code: `import os from "os";
export * from "hello";`,
      options: { includeExports: true },
    },
    {
      code: `import os from "os";
export { hello as hi } from "hello";`,
      options: { includeExports: true },
    },
    {
      code: `import os from "os";
export default function(){};`,
      options: { includeExports: true },
    },
    {
      code: `import { merge } from "lodash-es";
export { merge as lodashMerge }`,
      options: { includeExports: true },
    },
    {
      code: `export { something } from "os";
export * as os from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import { something } from "os";
export * as os from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import * as os from "os";
export { something } from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import os from "os";
export * from "os";`,
      options: { includeExports: true },
    },
    {
      code: `export { something } from "os";
export * from "os";`,
      options: { includeExports: true },
    },

    // ---- ESLint upstream — TypeScript suite ----
    `import type { Os } from "os";
import type { Fs } from "fs";`,
    `import { type Os } from "os";
import type { Fs } from "fs";`,
    `import type { Merge } from "lodash-es";`,
    `import _, { type Merge } from "lodash-es";`,
    `import type * as Foobar from "async";`,
    `import type Os from "os";
export type { Something } from "os";`,
    `import type Os from "os";
export { type Something } from "os";`,
    `import type * as Bar from "os";
import { type Baz } from "os";`,
    `import foo, * as bar from "os";
import { type Baz } from "os";`,
    `import foo, { type bar } from "os";
import type * as Baz from "os";`,
    `import type { Merge } from "lodash-es";
import type _ from "lodash-es";`,
    {
      code: `import type Os from "os";
export { type Hello } from "hello";`,
      options: { includeExports: true },
    },
    {
      code: `import type Os from "os";
export type * from "hello";`,
      options: { includeExports: true },
    },
    {
      code: `import type Os from "os";
export { type Hello as Hi } from "hello";`,
      options: { includeExports: true },
    },
    {
      code: `import type Os from "os";
export default function(){};`,
      options: { includeExports: true },
    },
    {
      code: `import { type Merge } from "lodash-es";
export { Merge as lodashMerge }`,
      options: { includeExports: true },
    },
    {
      code: `export type { Something } from "os";
export * as os from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import { type Something } from "os";
export * as os from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import type * as Os from "os";
export { something } from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import type Os from "os";
export * from "os";`,
      options: { includeExports: true },
    },
    {
      code: `import type Os from "os";
export type { Something } from "os";`,
      options: { includeExports: true },
    },
    {
      code: `export type { Something } from "os";
export * from "os";`,
      options: { includeExports: true },
    },

    // ---- allowSeparateTypeImports ----
    {
      code: `import { foo, type Bar } from "module";`,
      options: { allowSeparateTypeImports: true },
    },
    {
      code: `import { foo } from "module";
import type { Bar } from "module";`,
      options: { allowSeparateTypeImports: true },
    },
    {
      code: `import { type Foo } from "module";
import type { Bar } from "module";`,
      options: { allowSeparateTypeImports: true },
    },
    {
      code: `import { foo, type Bar } from "module";
export { type Baz } from "module2";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
    },
    {
      code: `import type { Foo } from "module";
export { bar, type Baz } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
    },
    {
      code: `import { type Foo } from "module";
export type { Bar } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
    },
    {
      code: `import type * as Foo from "module";
export { type Bar } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
    },
    {
      code: `import { type Foo } from "module";
export type * as Bar from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
    },
  ],
  invalid: [
    // ---- ESLint upstream — JS suite ----
    {
      code: `import "fs";
import "fs"`,
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import { merge } from "lodash-es";
import { find } from "lodash-es";`,
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import { merge } from "lodash-es";
import _ from "lodash-es";`,
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import os from "os";
import { something } from "os";
import * as foobar from "os";`,
      errors: [
        { messageId: 'import', line: 2, column: 1 },
        { messageId: 'import', line: 3, column: 1 },
      ],
    },
    {
      code: `import * as modns from "lodash-es";
import { merge } from "lodash-es";
import { baz } from "lodash-es";`,
      errors: [{ messageId: 'import', line: 3, column: 1 }],
    },
    {
      code: `export { os } from "os";
export { something } from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'export', line: 2, column: 1 }],
    },
    {
      code: `import os from "os";
export { os as foobar } from "os";
export { something } from "os";`,
      options: { includeExports: true },
      errors: [
        { messageId: 'exportAs', line: 2, column: 1 },
        { messageId: 'export', line: 3, column: 1 },
        { messageId: 'exportAs', line: 3, column: 1 },
      ],
    },
    {
      code: `import os from "os";
export { something } from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `import os from "os";
export * as os from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `export * as os from "os";
import os from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'importAs', line: 2, column: 1 }],
    },
    {
      code: `import * as modns from "mod";
export * as  modns from "mod";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `export * from "os";
export * from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'export', line: 2, column: 1 }],
    },
    {
      code: `import "os";
export * from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },

    // ---- ESLint upstream — TypeScript suite ----
    {
      code: `import { type Merge } from "lodash-es";
import { type Find } from "lodash-es";`,
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import { type Merge } from "lodash-es";
import type { Find } from "lodash-es";`,
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import type { Merge } from "lodash-es";
import type { Find } from "lodash-es";`,
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import type Os from "os";
import type { Something } from "os";
import type * as Foobar from "os";`,
      errors: [{ messageId: 'import', line: 3, column: 1 }],
    },
    {
      code: `import type * as Modns from "lodash-es";
import type { Merge } from "lodash-es";
import type { Baz } from "lodash-es";`,
      errors: [{ messageId: 'import', line: 3, column: 1 }],
    },
    {
      code: `import { type Foo } from "module";
export type { Bar } from "module";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `export { os } from "os";
export type { Something } from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'export', line: 2, column: 1 }],
    },
    {
      code: `export type { Os } from "os";
export type { Something } from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'export', line: 2, column: 1 }],
    },
    {
      code: `import type { Os } from "os";
export type { Os as Foobar } from "os";
export type { Something } from "os";`,
      options: { includeExports: true },
      errors: [
        { messageId: 'exportAs', line: 2, column: 1 },
        { messageId: 'export', line: 3, column: 1 },
        { messageId: 'exportAs', line: 3, column: 1 },
      ],
    },
    {
      code: `import type { Os } from "os";
export type { Something } from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `import type Os from "os";
export type * as Os from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `import type * as Modns from "mod";
export type * as Modns from "mod";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },
    {
      code: `export type * from "os";
export type * from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'export', line: 2, column: 1 }],
    },
    {
      code: `import "os";
export type { Os } from "os";`,
      options: { includeExports: true },
      errors: [{ messageId: 'exportAs', line: 2, column: 1 }],
    },

    // ---- allowSeparateTypeImports invalid forms ----
    {
      code: `import { someValue } from 'module';
import { anotherValue } from 'module';`,
      options: { allowSeparateTypeImports: true },
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import type { Merge } from "lodash-es";
import type { Find } from "lodash-es";`,
      options: { allowSeparateTypeImports: true },
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `import { someValue, type Foo } from 'module';
import type { SomeType } from 'module';
import type { AnotherType } from 'module';`,
      options: { allowSeparateTypeImports: true },
      errors: [{ messageId: 'import', line: 3, column: 1 }],
    },
    {
      code: `import { type Foo } from 'module';
import { type Bar } from 'module';`,
      options: { allowSeparateTypeImports: true },
      errors: [{ messageId: 'import', line: 2, column: 1 }],
    },
    {
      code: `export type { Foo } from "module";
export type { Bar } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
      errors: [{ messageId: 'export', line: 2, column: 1 }],
    },
    {
      code: `import { type Foo } from "module";
export { type Bar } from "module";
export { type Baz } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
      errors: [
        { messageId: 'exportAs', line: 2, column: 1 },
        { messageId: 'export', line: 3, column: 1 },
        { messageId: 'exportAs', line: 3, column: 1 },
      ],
    },
    {
      code: `import { type Foo } from "module";
export { type Bar } from "module";
export { regular } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
      errors: [
        { messageId: 'exportAs', line: 2, column: 1 },
        { messageId: 'export', line: 3, column: 1 },
        { messageId: 'exportAs', line: 3, column: 1 },
      ],
    },
    {
      code: `import { type Foo } from "module";
import { regular } from "module";
export { type Bar } from "module";
export { regular as other } from "module";`,
      options: { allowSeparateTypeImports: true, includeExports: true },
      errors: [
        { messageId: 'import', line: 2, column: 1 },
        { messageId: 'exportAs', line: 3, column: 1 },
        { messageId: 'export', line: 4, column: 1 },
        { messageId: 'exportAs', line: 4, column: 1 },
      ],
    },
  ],
});
