import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const alphabetical = { messageId: 'sortImportsAlphabetically' };
const ignoreCase = [{ ignoreCase: true }] as any;

ruleTester.run('sort-imports', {
  valid: [
    "import a from 'foo.js';\nimport b from 'bar.js';\nimport c from 'baz.js';\n",
    "import * as B from 'foo.js';\nimport A from 'bar.js';",
    "import * as B from 'foo.js';\nimport {a, b} from 'bar.js';",
    "import {b, c} from 'bar.js';\nimport A from 'foo.js';",
    {
      code: "import A from 'bar.js';\nimport {b, c} from 'foo.js';",
      options: [
        {
          memberSyntaxSortOrder: ['single', 'multiple', 'none', 'all'],
        },
      ] as any,
    },
    "import {a, b} from 'bar.js';\nimport {c, d} from 'foo.js';",
    "import A from 'foo.js';\nimport B from 'bar.js';",
    "import A from 'foo.js';\nimport a from 'bar.js';",
    "import a, * as b from 'foo.js';\nimport c from 'bar.js';",
    "import 'foo.js';\n import a from 'bar.js';",
    "import B from 'foo.js';\nimport a from 'bar.js';",
    {
      code: "import a from 'foo.js';\nimport B from 'bar.js';",
      options: ignoreCase,
    },
    "import {a, b, c, d} from 'foo.js';",
    {
      code: "import a from 'foo.js';\nimport B from 'bar.js';",
      options: [{ ignoreDeclarationSort: true }] as any,
    },
    {
      code: "import {b, A, C, d} from 'foo.js';",
      options: [{ ignoreMemberSort: true }] as any,
    },
    {
      code: "import {B, a, C, d} from 'foo.js';",
      options: [{ ignoreMemberSort: true }] as any,
    },
    {
      code: "import {a, B, c, D} from 'foo.js';",
      options: ignoreCase,
    },
    "import a, * as b from 'foo.js';",
    "import * as a from 'foo.js';\n\nimport b from 'bar.js';",
    "import * as bar from 'bar.js';\nimport * as foo from 'foo.js';",
    {
      code: "import 'foo';\nimport bar from 'bar';",
      options: ignoreCase,
    },
    "import React, {Component} from 'react';",
    {
      code: "import b from 'b';\n\nimport a from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import a from 'a';\n\nimport 'b';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import { b } from 'b';\n\n\nimport { a } from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import b from 'b';\n// comment\nimport a from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import b from 'b';\nfoo();\nimport a from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import { b } from 'b';/*\n comment \n*/import { a } from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import b from\n'b';\n\nimport\n a from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import c from 'c';\n\nimport a from 'a';\nimport b from 'b';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
    {
      code: "import c from 'c';\n\nimport b from 'b';\n\nimport a from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
    },
  ],
  invalid: [
    {
      code: "import a from 'foo.js';\nimport A from 'bar.js';",
      errors: [alphabetical],
    },
    {
      code: "import b from 'foo.js';\nimport a from 'bar.js';",
      errors: [alphabetical],
    },
    {
      code: "import {b, c} from 'foo.js';\nimport {a, d} from 'bar.js';",
      errors: [alphabetical],
    },
    {
      code: "import * as foo from 'foo.js';\nimport * as bar from 'bar.js';",
      errors: [alphabetical],
    },
    {
      code: "import a from 'foo.js';\nimport {b, c} from 'bar.js';",
      errors: [{ messageId: 'unexpectedSyntaxOrder' }],
    },
    {
      code: "import a from 'foo.js';\nimport * as b from 'bar.js';",
      errors: [{ messageId: 'unexpectedSyntaxOrder' }],
    },
    {
      code: "import a from 'foo.js';\nimport 'bar.js';",
      errors: [{ messageId: 'unexpectedSyntaxOrder' }],
    },
    {
      code: "import b from 'bar.js';\nimport * as a from 'foo.js';",
      options: [
        {
          memberSyntaxSortOrder: ['all', 'single', 'multiple', 'none'],
        },
      ] as any,
      errors: [{ messageId: 'unexpectedSyntaxOrder' }],
    },
    {
      code: "import {b, a, d, c} from 'foo.js';",
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import {b, a, d, c} from 'foo.js';\nimport {e, f, g, h} from 'bar.js';",
      options: [{ ignoreDeclarationSort: true }] as any,
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import {a, B, c, D} from 'foo.js';",
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import {zzzzz, /* comment */ aaaaa} from 'foo.js';",
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import {zzzzz /* comment */, aaaaa} from 'foo.js';",
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import {/* comment */ zzzzz, aaaaa} from 'foo.js';",
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import {zzzzz, aaaaa /* comment */} from 'foo.js';",
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: `
              import {
                boop,
                foo,
                zoo,
                baz as qux,
                bar,
                beep
              } from 'foo.js';
            `,
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
    {
      code: "import b from 'b';\nimport a from 'a';",
      errors: [alphabetical],
    },
    {
      code: "import b from 'b';\n\nimport a from 'a';",
      errors: [alphabetical],
    },
    {
      code: "import b from 'b';\nimport a from 'a';",
      options: [{}] as any,
      errors: [alphabetical],
    },
    {
      code: "import b from 'b';\nimport a from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import b from 'b';import a from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import b from 'b'; /* comment */ import a from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import b from 'b'; // comment\nimport a from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import b from 'b'; // comment 1\n/* comment 2 */import a from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import { b } from 'b'; /* comment line 1 \n comment line 2 */ import { a } from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import b\nfrom 'b'; import a\nfrom 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import { b } from \n'b'; /* comment */ import\n { a } from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import { b } from \n'b';\nimport\n { a } from 'a';",
      options: [{ allowSeparatedGroups: false }] as any,
      errors: [alphabetical],
    },
    {
      code: "import c from 'c';\n\nimport b from 'b';\nimport a from 'a';",
      options: [{ allowSeparatedGroups: true }] as any,
      errors: [{ messageId: 'sortImportsAlphabetically', line: 4 }],
    },
    {
      code: "import b from 'b';\n\nimport { c, a } from 'c';",
      options: [{ allowSeparatedGroups: true }] as any,
      errors: [{ messageId: 'sortMembersAlphabetically' }],
    },
  ],
});
