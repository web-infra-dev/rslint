import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
const rule = null as never;

ruleTester.run('order', rule, {
  valid: [
    // ============================================================
    // Default order
    // ============================================================
    {
      code: `
var fs = require('fs');
var async = require('async');
var relParent1 = require('../foo');
var relParent2 = require('../foo/bar');
var sibling = require('./foo');
var index = require('./');
`,
    },
    {
      code: `
import fs from 'fs';
import async from 'async';
import relParent1 from '../foo';
import sibling from './foo';
import index from './';
`,
    },
    {
      code: `
import fs = require('fs');
import async = require('async');
import sibling = require('./foo');
`,
    },
    // mixed import + require: imports first within same group
    {
      code: `
import fs from 'fs';
import async from 'async';
var path = require('path');
var _ = require('lodash');
`,
    },
    // Reverse order via groups
    {
      code: `
var index = require('./');
var sibling = require('./foo');
var parent = require('../foo');
var async = require('async');
var fs = require('fs');
`,
      options: [
        { groups: ['index', 'sibling', 'parent', 'external', 'builtin'] },
      ],
    },
    // ============================================================
    // newlines-between
    // ============================================================
    {
      code: `
import fs from 'fs';

import async from 'async';

import sibling from './foo';
`,
      options: [{ 'newlines-between': 'always' }],
    },
    {
      code: `
import fs from 'fs';
import async from 'async';
import sibling from './foo';
`,
      options: [{ 'newlines-between': 'never' }],
    },
    // ============================================================
    // alphabetize
    // ============================================================
    {
      code: `
import a from 'a';
import b from 'b';
`,
      options: [{ alphabetize: { order: 'asc' } }],
    },
    {
      code: `
import b from 'b';
import a from 'a';
`,
      options: [{ alphabetize: { order: 'desc' } }],
    },
    // ============================================================
    // pathGroups
    // ============================================================
    {
      code: `
import path from 'path';
import async from 'async';
import a from '@app/foo';
import sibling from './foo';
`,
      options: [
        {
          pathGroups: [
            { pattern: '@app/**', group: 'external', position: 'after' },
          ],
        },
      ],
    },
    // ============================================================
    // Type imports
    // ============================================================
    {
      code: `
import fs from 'fs';
import async from 'async';
import type { T } from 'foo';
`,
      options: [{ groups: ['builtin', 'external', 'type'] }],
    },
    // ============================================================
    // Named ordering
    // ============================================================
    {
      code: `
import { a, b, c } from 'foo';
`,
      options: [{ named: true, alphabetize: { order: 'asc' } }],
    },
    {
      code: `
import { type T, a, b } from 'foo';
`,
      options: [
        {
          named: { enabled: true, types: 'types-first' },
          alphabetize: { order: 'asc' },
        },
      ],
    },
    // ============================================================
    // Settings
    // ============================================================
    {
      code: `
import x from 'my-builtin';
import async from 'async';
`,
      settings: { 'import/core-modules': ['my-builtin'] },
    },
    // ============================================================
    // declare module — independent scope
    // ============================================================
    {
      code: `
import fs from 'fs';
import async from 'async';

declare module 'x' {
  import path from 'path';
  import lodash from 'lodash';
}
`,
    },
    // ============================================================
    // Side-effect imports ignored by default
    // ============================================================
    {
      code: `
import './styles.css';
import 'something-else';
import path from 'path';
`,
    },
    // ============================================================
    // Empty / comment-only files
    // ============================================================
    { code: '' },
    { code: '// just a comment' },
  ],
  invalid: [
    // ============================================================
    // Default order
    // ============================================================
    {
      code: `
var async = require('async');
var fs = require('fs');
`,
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
        },
      ],
    },
    {
      code: `
var sibling = require('./foo');
var parent = require('../foo');
`,
      errors: [{ messageId: 'order' }],
    },
    {
      code: `
import sibling from './foo';
import fs = require('fs');
`,
      errors: [{ messageId: 'order' }],
    },
    // ============================================================
    // Multiple out-of-order
    // ============================================================
    {
      code: `
import sibling from './foo';
import async from 'async';
import fs from 'fs';
`,
      errors: [{ messageId: 'order' }, { messageId: 'order' }],
    },
    // ============================================================
    // newlines-between
    // ============================================================
    {
      code: `
import fs from 'fs';
import async from 'async';
`,
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    {
      code: `
import fs from 'fs';

import async from 'async';
`,
      options: [{ 'newlines-between': 'never' }],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
        },
      ],
    },
    // ============================================================
    // alphabetize
    // ============================================================
    {
      code: `
import b from 'b';
import a from 'a';
`,
      options: [{ alphabetize: { order: 'asc' } }],
      errors: [
        {
          messageId: 'order',
          message: '`a` import should occur before import of `b`',
        },
      ],
    },
    {
      code: `
import a from 'a';
import b from 'b';
`,
      options: [{ alphabetize: { order: 'desc' } }],
      errors: [{ messageId: 'order' }],
    },
    // ============================================================
    // pathGroups
    // ============================================================
    {
      code: `
import a from '@app/foo';
import path from 'path';
`,
      options: [
        {
          pathGroups: [
            { pattern: '@app/**', group: 'external', position: 'after' },
          ],
        },
      ],
      errors: [{ messageId: 'order' }],
    },
    // ============================================================
    // Type import message wording
    // ============================================================
    {
      code: `
import async from 'async';
import type {T} from 'fs';
`,
      errors: [
        {
          messageId: 'order',
          message: '`fs` type import should occur before import of `async`',
        },
      ],
    },
    // ============================================================
    // warnOnUnassignedImports
    // ============================================================
    {
      code: `
import fs from 'fs';
import './styles.css';
import path from 'path';
`,
      options: [{ warnOnUnassignedImports: true }],
      errors: [{ messageId: 'order' }],
    },
    // ============================================================
    // Named ordering
    // ============================================================
    {
      code: `
import { c, a, b } from 'foo';
`,
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        {
          messageId: 'namedOrder',
          message: '`c` import should occur after import of `b`',
        },
      ],
    },
    // ============================================================
    // declare module body checked independently
    // ============================================================
    {
      code: `
import fs from 'fs';
declare module 'x' {
  import sibling from './foo';
  import path from 'path';
}
`,
      errors: [{ messageId: 'order' }],
    },
  ],
});
