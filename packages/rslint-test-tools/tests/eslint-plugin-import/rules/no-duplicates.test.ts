import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
const rule = null as never;

ruleTester.run('no-duplicates', rule, {
  valid: [
    // Different modules
    { code: "import { x } from './foo'; import { y } from './bar'" },

    // Unresolved modules with different names
    { code: 'import foo from "234artaf"; import { shoop } from "234q25ad"' },

    // Namespace + named from same module is allowed (cannot be merged)
    { code: "import * as ns from './foo'; import {y} from './foo'" },
    { code: "import {y} from './foo'; import * as ns from './foo'" },

    // Type import + value import (not prefer-inline)
    { code: "import type { x } from './foo'; import y from './foo'" },

    // Different type imports
    { code: "import type x from './foo'; import type y from './bar'" },
    { code: "import type {x} from './foo'; import type {y} from './bar'" },

    // Type default + type named from same module
    { code: "import type x from './foo'; import type {y} from './foo'" },

    // Different query strings with considerQueryString
    {
      code: "import x from './bar?optionX'; import y from './bar?optionY';",
      options: [{ considerQueryString: true }],
    },

    // Inline type specifier + value import (not prefer-inline)
    { code: "import { type x } from './foo'; import y from './foo'" },
    { code: "import { type x } from './foo'; import { y } from './foo'" },
  ],
  invalid: [
    // Basic duplicate named imports
    {
      code: "import { x } from './foo'; import { y } from './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Three-way merge
    {
      code: "import {x} from './foo'; import {y} from './foo'; import { z } from './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Side-effect import + named import
    {
      code: "import './foo'; import {x} from './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Default + named import merge
    {
      code: "import def from './foo'; import {x} from './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Duplicate side-effect imports
    {
      code: "import './foo'; import './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Duplicate unresolved modules
    {
      code: "import foo from 'non-existent'; import bar from 'non-existent';",
      errors: [
        { message: "'non-existent' imported multiple times." },
        { message: "'non-existent' imported multiple times." },
      ],
    },

    // TypeScript: duplicate type-only named imports
    {
      code: "import type {x} from './foo'; import type {y} from './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // prefer-inline: type + value imports
    {
      code: "import {AValue} from './foo'; import type {AType} from './foo'",
      options: [{ 'prefer-inline': true }],
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // prefer-inline: both inline type imports
    {
      code: "import {type x} from './foo'; import {type y} from './foo'",
      options: [{ 'prefer-inline': true }],
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Namespace imports cannot be merged (but still reported)
    {
      code: "import * as ns1 from './foo'; import * as ns2 from './foo'",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },

    // Query strings without considerQueryString
    {
      code: "import x from './bar?optionX'; import y from './bar?optionY';",
      errors: [
        { message: "'./bar' imported multiple times." },
        { message: "'./bar' imported multiple times." },
      ],
    },

    // Multiline with trailing newline removal
    {
      code: "import { Foo } from './foo';\nimport { Bar } from './foo';\nexport const value = {}",
      errors: [
        { message: "'./foo' imported multiple times." },
        { message: "'./foo' imported multiple times." },
      ],
    },
  ],
});
