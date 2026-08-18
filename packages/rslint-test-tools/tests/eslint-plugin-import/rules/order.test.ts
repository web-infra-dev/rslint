// cspell:ignore asdas barfoo blist chpro Promisable rootverse vtaits xcompose yargs
// Migrated from eslint-plugin-import v2.32.0 tests/src/rules/order.js.
// Semantic duplicates across the core, TypeScript, and Babel suites are collapsed: 162 valid + 132 invalid.
// Four Flow-only `import typeof` cases are omitted because the TypeScript parser rejects them; the generic Flow-suite case remains below.
import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
const rule = null as never;

ruleTester.run('order', rule, {
  valid: [],
  invalid: [
    // Upstream case 0:0.
    {
      code: "\n        var async = require('async');\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:1.
    {
      code: "\n        var async = require('async');\n        var fs = require('fs'); \n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:2.
    {
      code: "\n        var async = require('async');\n        var fs = require('fs'); /* comment */\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:3.
    {
      code: "\n        /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:4.
    {
      code: "\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:5.
    {
      code: "/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n/* comment3 */  var fs = require('fs'); /* comment4 */\r\n",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:6.
    {
      code: "\n        /* multi-line1\n          comment1 */  var async = require('async'); /* multi-line2\n          comment2 */  var fs = require('fs'); /* multi-line3\n          comment3 */\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:7.
    {
      code: "\n        var {b} = require('async');\n        var {a} = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:8.
    {
      code: "\n        var async = require('async');\n        var fs =\n          require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:9.
    {
      code: "\n        var async = require('async');\n        var fs = require('fs');",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:10.
    {
      code: "\n        import async from 'async';\n        import fs from 'fs';\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:11.
    {
      code: "\n        var async = require('async');\n        import fs from 'fs';\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:12.
    {
      code: "\n        var parent = require('../parent');\n        var async = require('async');\n      ",
      errors: [
        { message: '`async` import should occur before import of `../parent`' },
      ],
    },
    // Upstream case 0:13.
    {
      code: "\n        var sibling = require('./sibling');\n        var parent = require('../parent');\n      ",
      errors: [
        {
          message:
            '`../parent` import should occur before import of `./sibling`',
        },
      ],
    },
    // Upstream case 0:14.
    {
      code: "\n        var index = require('./');\n        var sibling = require('./sibling');\n      ",
      errors: [
        { message: '`./sibling` import should occur before import of `./`' },
      ],
    },
    // Upstream case 0:15.
    {
      code: "\n          var sibling = require('./sibling');\n          var async = require('async');\n          var fs = require('fs');\n        ",
      errors: [
        { message: '`async` import should occur before import of `./sibling`' },
        { message: '`fs` import should occur before import of `./sibling`' },
      ],
    },
    // Upstream case 0:16.
    {
      code: "\n        var index = require('./');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n      ",
      errors: [{ message: '`./` import should occur after import of `bar`' }],
    },
    // Upstream case 0:17.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n      ",
      options: [
        { groups: ['index', 'sibling', 'parent', 'external', 'builtin'] },
      ],
      errors: [{ message: '`./` import should occur before import of `fs`' }],
    },
    // Upstream case 0:18.
    {
      code: "\n        var foo = require('./foo').bar;\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `./foo`' },
      ],
    },
    // Upstream case 0:19.
    {
      code: "\n        var foo = require('./foo').bar.bar.bar;\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `./foo`' },
      ],
    },
    // Upstream case 0:20.
    {
      code: "\n        var foo = require('./foo').bar\n          .bar\n          .bar;\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `./foo`' },
      ],
    },
    // Upstream case 0:21.
    {
      code: "\n        var foo = require('./foo');\n        var fs = require('fs').bar\n          .bar\n          .bar;\n      ",
      errors: [
        { message: '`fs` import should occur before import of `./foo`' },
      ],
    },
    // Upstream case 0:22.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var sibling = require('./foo');\n        var path = require('path');\n      ",
      options: [
        {
          groups: [
            ['builtin', 'index'],
            ['sibling', 'parent', 'external'],
          ],
        },
      ],
      errors: [
        { message: '`path` import should occur before import of `./foo`' },
      ],
    },
    // Upstream case 0:23.
    {
      code: "\n        var path = require('path');\n        var async = require('async');\n      ",
      options: [
        { groups: ['index', ['sibling', 'parent', 'external', 'internal']] },
      ],
      errors: [
        { message: '`async` import should occur before import of `path`' },
      ],
    },
    // Upstream case 0:24.
    {
      code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        import sibling, {foo3} from './foo';\n        var index = require('./');\n      ",
      errors: [
        { message: '`./foo` import should occur before import of `fs`' },
      ],
    },
    // Upstream case 0:25.
    {
      code: "\n        var fs = require('fs');\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n      ",
      errors: [
        { message: '`fs` import should occur after import of `../foo/bar`' },
      ],
    },
    // Upstream case 0:26.
    {
      code: "\n          var fs = require('fs');\n          import async, {foo1} from 'async';\n          import bar = require(\"../foo/bar\");\n        ",
      errors: [
        { message: '`fs` import should occur after import of `../foo/bar`' },
      ],
    },
    // Upstream case 0:27.
    {
      code: "\n          var async = require('async');\n          var fs = require('fs');\n        ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:28.
    {
      code: "\n          import sync = require('sync');\n          import async, {foo1} from 'async';\n\n          import index from './';\n        ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
      errors: [
        { message: '`async` import should occur before import of `sync`' },
      ],
    },
    // Upstream case 0:29.
    {
      code: "\n          import log = console.log;\n          import blah = require('./blah');",
      errors: [
        {
          message:
            '`./blah` import should occur before import of `console.log`',
        },
      ],
    },
    // Upstream case 0:30.
    {
      code: "\n        import { Button } from '-/components/Button';\n        import { add } from './helper';\n        import fs from 'fs';\n      ",
      options: [
        {
          groups: [
            'builtin',
            'external',
            'unknown',
            'parent',
            'sibling',
            'index',
          ],
        },
      ],
      errors: [
        {
          message:
            '`fs` import should occur before import of `-/components/Button`',
        },
      ],
    },
    // Upstream case 0:31.
    {
      code: "\n        import fs from 'fs';\n        import { Button } from '-/components/Button';\n        import { LinkButton } from '-/components/Link';\n        import { add } from './helper';\n      ",
      options: [
        {
          groups: [
            'builtin',
            'external',
            'unknown',
            'parent',
            'sibling',
            'index',
          ],
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:32.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'never',
        },
      ],
      errors: [
        { message: 'There should be no empty line between import groups' },
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 0:33.
    {
      code: "\n        var fs = require('fs'); /* comment */\n\n        var index = require('./');\n      ",
      options: [
        { groups: [['builtin'], ['index']], 'newlines-between': 'never' },
      ],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 0:34.
    {
      code: "\n        var fs = require('fs'); /* multi-line\n        comment */\n\n        var index = require('./');\n      ",
      options: [
        { groups: [['builtin'], ['index']], 'newlines-between': 'never' },
      ],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 0:35.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:36.
    {
      code: "\n        var fs = require('fs');\n\n        var path = require('path');\n        var index = require('./');\n\n        var sibling = require('./foo');\n\n        var async = require('async');\n      ",
      options: [
        {
          groups: [
            ['builtin', 'index'],
            ['sibling', 'parent', 'external'],
          ],
          'newlines-between': 'always',
        },
      ],
      errors: [
        { message: 'There should be no empty line within import group' },
        { message: 'There should be no empty line within import group' },
      ],
    },
    // Upstream case 0:37.
    {
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [
        { 'newlines-between': 'never', warnOnUnassignedImports: false },
      ],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 0:38.
    {
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'never', warnOnUnassignedImports: true }],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 0:39.
    {
      code: "\n        import path from 'path';\n        export const abc = 123;\n\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'never' }],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 0:40.
    {
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:41.
    {
      code: "\n        import path from 'path'; // comment\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:42.
    {
      code: "\n        import path from 'path'; /* comment */ /* comment */\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:43.
    {
      code: "\n        import path from 'path'; /* 1\n        2 */\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:44.
    {
      code: "\n        const local = require('./local');\n\n        fn_call();\n\n        const global1 = require('global1');\n        const global2 = require('global2');\n\n        fn_call();\n      ",
      errors: [
        { message: '`./local` import should occur after import of `global2`' },
      ],
    },
    // Upstream case 0:45.
    {
      code: "\n        const local = require('./local');\n        fn_call();\n        const global1 = require('global1');\n        const global2 = require('global2');\n\n        fn_call();\n      ",
      errors: [
        { message: '`./local` import should occur after import of `global2`' },
      ],
    },
    // Upstream case 0:46.
    {
      code: "\n        const local1 = require('./local1');\n        const local2 = require('./local2');\n        const local3 = require('./local3');\n        const local4 = require('./local4');\n        fn_call();\n        const global1 = require('global1');\n        const global2 = require('global2');\n        const global3 = require('global3');\n        const global4 = require('global4');\n        const global5 = require('global5');\n        fn_call();\n      ",
      errors: [
        { message: '`./local1` import should occur after import of `global5`' },
        { message: '`./local2` import should occur after import of `global5`' },
        { message: '`./local3` import should occur after import of `global5`' },
        { message: '`./local4` import should occur after import of `global5`' },
      ],
    },
    // Upstream case 0:47.
    {
      code: "\n        const local = require('./local');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const global3 = require('global3');\n\n        fn_call();\n      ",
      errors: [
        { message: '`./local` import should occur after import of `global3`' },
      ],
    },
    // Upstream case 0:48.
    {
      code: "\n        const local1 = require('./local1');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const local2 = require('./local2');\n        const global3 = require('global3');\n        const global4 = require('global4');\n\n        fn_call();\n      ",
      errors: [
        { message: '`./local1` import should occur after import of `global4`' },
        { message: '`./local2` import should occur after import of `global4`' },
      ],
    },
    // Upstream case 0:49.
    {
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        ",
      options: [
        {
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'after' },
          ],
        },
      ],
      errors: [
        {
          message:
            '`~/components/Input` import should occur before import of `./helper`',
        },
      ],
    },
    // Upstream case 0:50.
    {
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        ",
      options: [{ pathGroups: [{ pattern: '~/**', group: 'external' }] }],
      errors: [
        { message: '`./helper` import should occur after import of `async`' },
      ],
    },
    // Upstream case 0:51.
    {
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        ",
      options: [
        {
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'before' },
          ],
        },
      ],
      errors: [
        {
          message:
            '`~/components/Input` import should occur before import of `lodash`',
        },
      ],
    },
    // Upstream case 0:52.
    {
      code: "\n        import fs from 'fs';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        import { Export } from '-/components/Export';\n        ",
      options: [
        {
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'after' },
            { pattern: '#/**', group: 'external', position: 'after' },
            { pattern: '-/**', group: 'external', position: 'before' },
            { pattern: '$/**', group: 'external', position: 'before' },
          ],
        },
      ],
      errors: [
        {
          message:
            '`-/components/Export` import should occur before import of `$/components/Import`',
        },
      ],
    },
    // Upstream case 0:53.
    {
      code: "\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        import { Output } from '~/components/Output';\n        ",
      options: [
        {
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'after' },
            { pattern: '#/**', group: 'external', position: 'after' },
            { pattern: '-/**', group: 'external', position: 'before' },
            { pattern: '$/**', group: 'external', position: 'before' },
          ],
        },
      ],
      errors: [
        {
          message:
            '`~/components/Output` import should occur before import of `#/components/Input`',
        },
      ],
    },
    // Upstream case 0:54.
    {
      code: "\n        import path from 'path';\n        import { namespace } from '@namespace';\n        import { a } from 'a';\n        import { b } from 'b';\n        import { c } from 'c';\n        import { d } from 'd';\n        import { e } from 'e';\n        import { f } from 'f';\n        import { g } from 'g';\n        import { h } from 'h';\n        import { i } from 'i';\n        import { j } from 'j';\n        import { k } from 'k';",
      options: [
        {
          groups: ['builtin', 'external', 'internal'],
          pathGroups: [
            { pattern: '@namespace', group: 'external', position: 'after' },
            { pattern: 'a', group: 'internal', position: 'before' },
            { pattern: 'b', group: 'internal', position: 'before' },
            { pattern: 'c', group: 'internal', position: 'before' },
            { pattern: 'd', group: 'internal', position: 'before' },
            { pattern: 'e', group: 'internal', position: 'before' },
            { pattern: 'f', group: 'internal', position: 'before' },
            { pattern: 'g', group: 'internal', position: 'before' },
            { pattern: 'h', group: 'internal', position: 'before' },
            { pattern: 'i', group: 'internal', position: 'before' },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['builtin'],
        },
      ],
      settings: { 'import/internal-regex': '^(a|b|c|d|e|f|g|h|i|j|k)(\\/|$)' },
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:55.
    {
      code: "\n        import external from 'external';\n        import a from '@namespace/a';\n        import b from '@namespace/b';\n        import { parent } from '../../parent';\n        import local from './local';\n        import './side-effect';",
      options: [
        {
          alphabetize: { order: 'asc', caseInsensitive: true },
          groups: [
            'type',
            'builtin',
            'external',
            'internal',
            'parent',
            'sibling',
            'index',
            'object',
          ],
          'newlines-between': 'always',
          pathGroups: [
            { pattern: '@namespace', group: 'external', position: 'after' },
            { pattern: '@namespace/**', group: 'external', position: 'after' },
          ],
          pathGroupsExcludedImportTypes: [],
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:56.
    {
      code: "\n        var async = require('async');\n        fn_call();\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:57.
    {
      code: "\n        const env = require('./config');\n\n        Object.keys(env);\n\n        const http = require('http');\n        const express = require('express');\n\n        http.createServer(express());\n      ",
      errors: [
        { message: '`./config` import should occur after import of `express`' },
      ],
    },
    // Upstream case 0:58.
    {
      code: "\n        var async = require('async');\n        var a = require('./value.js')(a);\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:59.
    {
      code: "\n        var async = require('async');\n        var fs = require('fs')(a);\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:60.
    {
      code: "\n        var async = require('async')(a);\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:61.
    {
      code: "\n        var async = require('async');\n        require('./aa');\n        var fs = require('fs');\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:62.
    {
      code: "\n        import async from 'async';\n        fn_call();\n        import fs from 'fs';\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:63.
    {
      code: "\n        import async from 'async';\n        var a = 1;\n        import fs from 'fs';\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:64.
    {
      code: "\n        import async from 'async';\n        var a = require('./value.js')(a);\n        import fs from 'fs';\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:65.
    {
      code: "\n        import async from 'async';\n        import './aa';\n        import fs from 'fs';\n      ",
      errors: [
        { message: '`fs` import should occur before import of `async`' },
      ],
    },
    // Upstream case 0:66.
    {
      code: "\n        import b from 'bar';\n        import c from 'Bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
      errors: [{ message: '`Bar` import should occur before import of `bar`' }],
    },
    // Upstream case 0:67.
    {
      code: "\n        import a from 'foo';\n        import c from 'Bar';\n        import b from 'bar';\n\n        import index from './';\n      ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'desc' } },
      ],
      errors: [{ message: '`bar` import should occur before import of `Bar`' }],
    },
    // Upstream case 0:68.
    {
      code: '\n        import a from "foo";\n        import b from "foo-bar";\n        import c from "foo/bar";\n        import d from "foo/barfoo";\n      ',
      options: [{ alphabetize: { order: 'asc' } }],
      errors: [
        {
          message: '`foo-bar` import should occur after import of `foo/barfoo`',
        },
      ],
    },
    // Upstream case 0:69.
    {
      code: "\n        import b from 'foo';\n        import a from 'Bar';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: { order: 'asc', caseInsensitive: true },
        },
      ],
      errors: [{ message: '`Bar` import should occur before import of `foo`' }],
    },
    // Upstream case 0:70.
    {
      code: "\n        import a from 'Bar';\n        import b from 'foo';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: { order: 'desc', caseInsensitive: true },
        },
      ],
      errors: [{ message: '`foo` import should occur before import of `Bar`' }],
    },
    // Upstream case 0:71.
    {
      code: "\n        const b = require('./b').get();\n        const a = require('./a');\n      ",
      options: [{ alphabetize: { order: 'asc' } }],
      errors: [{ message: '`./a` import should occur before import of `./b`' }],
    },
    // Upstream case 0:72.
    {
      code: "\n        import a from '../a';\n        import p from '..';\n      ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
      errors: [{ message: '`..` import should occur before import of `../a`' }],
    },
    // Upstream case 0:73.
    {
      code: "\n        import A from 'a';\n        import C from './c';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 0:74.
    {
      code: "\n        import A from 'a';\n\n        import C from 'c';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            { pattern: 'a', group: 'external', position: 'before' },
            { pattern: 'c', group: 'external', position: 'after' },
          ],
        },
      ],
      errors: [
        { message: 'There should be no empty line within import group' },
      ],
    },
    // Upstream case 0:75.
    {
      code: '\n        var { B, A: R } = require("./Z");\n        import { O as G, D } from "./Z";\n        import { K, L, J } from "./Z";\n        export { Z, X, Y } from "./Z";\n      ',
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` import should occur before import of `B`' },
        { message: '`D` import should occur before import of `O`' },
        { message: '`J` import should occur before import of `K`' },
        { message: '`Z` export should occur after export of `Y`' },
      ],
    },
    // Upstream case 0:76.
    {
      code: '\n        import { D, C } from "./Z";\n        var { B, A } = require("./Z");\n        export { B, A };\n      ',
      options: [
        {
          named: { require: false, import: true, export: true },
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        { message: '`C` import should occur before import of `D`' },
        { message: '`A` export should occur before export of `B`' },
      ],
    },
    // Upstream case 0:77.
    {
      code: '\n        import { A as B, A as C, A } from "./Z";\n        export { A, A as D, A as B, A as C } from "./Z";\n        const { a: b, a: c, a } = require("./Z");\n      ',
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` import should occur before import of `A as B`' },
        { message: '`A as D` export should occur after export of `A as C`' },
        { message: '`a` import should occur before import of `a as b`' },
      ],
    },
    // Upstream case 0:78.
    {
      code: '\n        import { A, B, C } from "./Z";\n        exports = { B, C, A };\n        module.exports = { c: C, a: A, b: B };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` export should occur before export of `B`' },
        { message: '`c` export should occur after export of `b`' },
      ],
    },
    // Upstream case 0:79.
    {
      code: '\n        exports.B = { };\n        module.exports.A = { };\n        module.exports.C = { };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
      errors: [{ message: '`A` export should occur before export of `B`' }],
    },
    // Upstream case 0:80.
    {
      code: '\n        exports.A.C = { };\n        module.exports.A.A = { };\n        exports.A.B = { };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
      errors: [{ message: '`A.C` export should occur after export of `A.B`' }],
    },
    // Upstream case 0:81.
    {
      code: '\n        const {\n          F: O,\n          O: B,\n          /* Hello World */\n          A: R\n        } = require("./Z");\n        import {\n          Y,\n          X,\n        } from "./Z";\n        export {\n          Z, A,\n          B\n        } from "./Z";\n        module.exports = {\n          a: A, o: O,\n          b: B\n        };\n      ',
      options: [{ named: { enabled: true }, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` import should occur before import of `F`' },
        { message: '`X` import should occur before import of `Y`' },
        { message: '`Z` export should occur after export of `B`' },
        { message: '`b` export should occur before export of `o`' },
      ],
    },
    // Upstream case 0:82.
    {
      code: "\n          const { cello } = require('./cello');\n          import { int } from './int';\n          const blah = require('./blah');\n          import { hello } from './hello';\n        ",
      errors: [
        { message: '`./int` import should occur before import of `./cello`' },
        { message: '`./hello` import should occur before import of `./cello`' },
      ],
    },
    // Upstream case 0:83.
    {
      code: "\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n        var async = require('async');\n        var relParent1 = require('../foo');\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n        var relParent3 = require('../bar');\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
      ],
    },
    // Upstream case 0:84.
    {
      code: "\n        var fs = require('fs');\n\n        var path = require('path');\n\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
      ],
    },
    // Upstream case 1:0.
    {
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
      errors: [
        { message: '`bar` import should occur after type import of `Bar`' },
      ],
    },
    // Upstream case 1:1.
    {
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'desc' } },
      ],
      errors: [{ message: '`bar` import should occur before import of `Bar`' }],
    },
    // Upstream case 1:2.
    {
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        { message: '`Bar` import should occur before import of `bar`' },
        {
          message: '`Bar` type import should occur before type import of `foo`',
        },
      ],
    },
    // Upstream case 1:3.
    {
      code: "\n              import a from 'foo';\n              import c from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'desc' },
        },
      ],
      errors: [
        { message: '`bar` import should occur before import of `Bar`' },
        {
          message: '`foo` type import should occur before type import of `Bar`',
        },
      ],
    },
    // Upstream case 1:4, 2:4.
    {
      code: "\n              import './local1';\n              import global from 'global1';\n              import local from './local2';\n              import 'global2';\n            ",
      options: [{ warnOnUnassignedImports: true }],
      errors: [
        {
          message: '`global1` import should occur before import of `./local1`',
        },
        {
          message: '`global2` import should occur before import of `./local1`',
        },
      ],
    },
    // Upstream case 1:5, 2:5.
    {
      code: "\n              import local from './local';\n\n              import 'global1';\n\n              import global2 from 'global2';\n              import global3 from 'global3';\n            ",
      options: [{ warnOnUnassignedImports: true }],
      errors: [
        { message: '`./local` import should occur after import of `global3`' },
      ],
    },
    // Upstream case 1:6.
    {
      code: "\n              import type { ParsedPath } from 'path';\n              import type { CopyOptions } from 'fs';\n\n              declare module 'my-module' {\n                import type { ParsedPath } from 'path';\n                import type { CopyOptions } from 'fs';\n              }\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
      errors: [
        {
          message: '`fs` type import should occur before type import of `path`',
        },
        {
          message: '`fs` type import should occur before type import of `path`',
        },
      ],
    },
    // Upstream case 1:7.
    // Also covers 2:7's ordering contract. tsgo recovers TypeScript's removed
    // `import type Default, { Named }` form with TypeScript-parser semantics.
    {
      code: '\n              import { type Z, A } from "./Z";\n              import type N, { E, D } from "./Z";\n              import type { L, G } from "./Z";\n            ',
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` import should occur before type import of `Z`' },
        { message: '`D` import should occur before import of `E`' },
        { message: '`G` import should occur before import of `L`' },
      ],
    },
    // Upstream case 1:8.
    {
      code: '\n              const { B, /* Hello World */ A } = require("./Z");\n              export { B, A } from "./Z";\n            ',
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` import should occur before import of `B`' },
        { message: '`A` export should occur before export of `B`' },
      ],
    },
    // Upstream case 1:9, 1:21.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 1:10, 1:22.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'always',
          'newlines-between-types': 'never',
        },
      ],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 1:11.
    {
      code: "\n              import c from 'Bar';\n\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
      ],
    },
    // Upstream case 1:12.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 1:13.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'never',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 1:14.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 1:15.
    {
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always-and-inside-groups' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 1:16.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'never' }],
      errors: [
        { message: 'There should be no empty line between import groups' },
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 1:17.
    {
      code: "\n                import React, { PureComponent } from 'react';\n                import aTypes from 'prop-types';\n                import { compose, apply } from 'xcompose';\n                import * as classnames from 'classnames';\n                import blist2 from 'blist';\n                import blist from 'BList';\n              ",
      options: [{ alphabetize: { order: 'asc', caseInsensitive: true } }],
      errors: [
        {
          message: '`prop-types` import should occur before import of `react`',
        },
        {
          message: '`classnames` import should occur before import of `react`',
        },
        { message: '`blist` import should occur before import of `react`' },
        { message: '`BList` import should occur before import of `react`' },
      ],
    },
    // Upstream case 1:18.
    {
      code: "\n              import { compose, apply } from 'xcompose';\n            ",
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`apply` import should occur before import of `compose`' },
      ],
    },
    // Upstream case 1:19.
    {
      code: "\n              import fs from 'fs';\n              import './styles.css';\n              import path from 'path';\n            ",
      options: [{ warnOnUnassignedImports: true }],
      errors: [
        {
          message: '`path` import should occur before import of `./styles.css`',
        },
      ],
    },
    // Upstream case 1:20.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        {
          message:
            '`../foo.js` type import should occur before type import of `fs`',
        },
        {
          message:
            '`./bar.js` type import should occur before type import of `fs`',
        },
        { message: '`./` type import should occur before type import of `fs`' },
      ],
    },
    // Upstream case 1:23.
    {
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n              var async = require('async');\n              var relParent1 = require('../foo');\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n              var relParent3 = require('../bar');\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
      ],
    },
    // Upstream case 1:24.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
      ],
    },
    // Upstream case 1:25.
    {
      code: '\n                export { type B, A };\n              ',
      options: [
        {
          named: { enabled: true, types: 'mixed' },
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        { message: '`A` export should occur before type export of `B`' },
      ],
    },
    // Upstream case 1:26.
    {
      code: '\n                import { type B, A, default as C } from "./Z";\n              ',
      options: [
        {
          named: { import: true, types: 'types-last' },
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        { message: '`B` type import should occur after import of `default`' },
      ],
    },
    // Upstream case 1:27.
    {
      code: '\n                export { A, type Z } from "./Z";\n              ',
      options: [{ named: { enabled: true, types: 'types-first' } }],
      errors: [
        { message: '`Z` type export should occur before export of `A`' },
      ],
    },
    // Upstream case 1:28, 2:25.
    {
      code: "\n                import express from 'express';\n                import log4js from 'log4js';\n                import chpro from 'node:child_process';\n                // import fsp from 'node:fs/promises';\n              ",
      options: [
        {
          groups: [
            'builtin',
            'external',
            'internal',
            'parent',
            'sibling',
            'index',
            'object',
            'type',
          ],
        },
      ],
      errors: [
        {
          message:
            '`node:child_process` import should occur before import of `express`',
        },
      ],
    },
    // Upstream case 2:0.
    {
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
      errors: [
        { message: '`bar` import should occur after type import of `Bar`' },
      ],
    },
    // Upstream case 2:1.
    {
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'desc' } },
      ],
      errors: [{ message: '`bar` import should occur before import of `Bar`' }],
    },
    // Upstream case 2:2.
    {
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        { message: '`Bar` import should occur before import of `bar`' },
        {
          message: '`Bar` type import should occur before type import of `foo`',
        },
      ],
    },
    // Upstream case 2:3.
    {
      code: "\n              import a from 'foo';\n              import c from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'desc' },
        },
      ],
      errors: [
        { message: '`bar` import should occur before import of `Bar`' },
        {
          message: '`foo` type import should occur before type import of `Bar`',
        },
      ],
    },
    // Upstream case 2:6.
    {
      code: "\n              import type { ParsedPath } from 'path';\n              import type { CopyOptions } from 'fs';\n\n              declare module 'my-module' {\n                import type { ParsedPath } from 'path';\n                import type { CopyOptions } from 'fs';\n              }\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
      errors: [
        {
          message: '`fs` type import should occur before type import of `path`',
        },
        {
          message: '`fs` type import should occur before type import of `path`',
        },
      ],
    },
    // Upstream case 2:8.
    {
      code: '\n              const { B, /* Hello World */ A } = require("./Z");\n              export { B, A } from "./Z";\n            ',
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`A` import should occur before import of `B`' },
        { message: '`A` export should occur before export of `B`' },
      ],
    },
    // Upstream case 2:9, 2:21.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 2:10, 2:22.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'always',
          'newlines-between-types': 'never',
        },
      ],
      errors: [
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 2:11.
    {
      code: "\n              import c from 'Bar';\n\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
        {
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
        },
      ],
    },
    // Upstream case 2:12.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 2:13.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'never',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 2:14.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 2:15.
    {
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always-and-inside-groups' }],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
      ],
    },
    // Upstream case 2:16.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'never' }],
      errors: [
        { message: 'There should be no empty line between import groups' },
        { message: 'There should be no empty line between import groups' },
      ],
    },
    // Upstream case 2:17.
    {
      code: "\n                import React, { PureComponent } from 'react';\n                import aTypes from 'prop-types';\n                import { compose, apply } from 'xcompose';\n                import * as classnames from 'classnames';\n                import blist2 from 'blist';\n                import blist from 'BList';\n              ",
      options: [{ alphabetize: { order: 'asc', caseInsensitive: true } }],
      errors: [
        {
          message: '`prop-types` import should occur before import of `react`',
        },
        {
          message: '`classnames` import should occur before import of `react`',
        },
        { message: '`blist` import should occur before import of `react`' },
        { message: '`BList` import should occur before import of `react`' },
      ],
    },
    // Upstream case 2:18.
    {
      code: "\n              import { compose, apply } from 'xcompose';\n            ",
      options: [{ named: true, alphabetize: { order: 'asc' } }],
      errors: [
        { message: '`apply` import should occur before import of `compose`' },
      ],
    },
    // Upstream case 2:19.
    {
      code: "\n              import fs from 'fs';\n              import './styles.css';\n              import path from 'path';\n            ",
      options: [{ warnOnUnassignedImports: true }],
      errors: [
        {
          message: '`path` import should occur before import of `./styles.css`',
        },
      ],
    },
    // Upstream case 2:20.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: { order: 'asc' },
        },
      ],
      errors: [
        {
          message:
            '`../foo.js` type import should occur before type import of `fs`',
        },
        {
          message:
            '`./bar.js` type import should occur before type import of `fs`',
        },
        { message: '`./` type import should occur before type import of `fs`' },
      ],
    },
    // Upstream case 2:23.
    {
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n              var async = require('async');\n              var relParent1 = require('../foo');\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n              var relParent3 = require('../bar');\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
      ],
    },
    // Upstream case 2:24.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between import groups',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
        },
        {
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
        },
      ],
    },
    // Upstream case 3:3. Generic Flow-suite case with no Flow-only syntax.
    {
      code: "\n        import { cfg } from 'path/path/path/src/Cfg';\n        import { l10n } from 'path/src/l10n';\n        import { helpers } from 'path/path/path/helpers';\n        import { tip } from 'path/path/tip';\n\n        import { controller } from '../../../../path/path/path/controller';\n        import { component } from '../../../../path/path/path/component';\n      ",
      options: [
        {
          groups: [
            ['builtin', 'external'],
            'internal',
            ['sibling', 'parent'],
            'object',
            'type',
          ],
          pathGroups: [
            {
              pattern: 'react',
              group: 'builtin',
              position: 'before',
              patternOptions: { matchBase: true },
            },
            {
              pattern: '*.+(css|svg)',
              group: 'type',
              position: 'after',
              patternOptions: { matchBase: true },
            },
          ],
          pathGroupsExcludedImportTypes: [],
          alphabetize: { order: 'asc' },
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          message:
            '`path/path/path/helpers` import should occur before import of `path/path/path/src/Cfg`',
        },
        {
          message:
            '`path/path/tip` import should occur before import of `path/src/l10n`',
        },
        {
          message:
            '`../../../../path/path/path/component` import should occur before import of `../../../../path/path/path/controller`',
        },
      ],
    },
  ],
});

ruleTester.run('order', rule, {
  valid: [
    // Upstream case 0:0.
    {
      code: "\n        var fs = require('fs');\n        var async = require('async');\n        var relParent1 = require('../foo');\n        var relParent2 = require('../foo/bar');\n        var relParent3 = require('../');\n        var relParent4 = require('..');\n        var sibling = require('./foo');\n        var index = require('./');",
    },
    // Upstream case 0:1.
    {
      code: "\n        import fs from 'fs';\n        import async, {foo1} from 'async';\n        import relParent1 from '../foo';\n        import relParent2, {foo2} from '../foo/bar';\n        import relParent3 from '../';\n        import sibling, {foo3} from './foo';\n        import index from './';",
    },
    // Upstream case 0:2.
    {
      code: "\n        var fs = require('fs');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var async = require('async');",
    },
    // Upstream case 0:3.
    {
      code: "\n        var index = require('./');\n        var sibling = require('./foo');\n        var relParent3 = require('../');\n        var relParent2 = require('../foo/bar');\n        var relParent1 = require('../foo');\n        var async = require('async');\n        var fs = require('fs');\n      ",
      options: [
        { groups: ['index', 'sibling', 'parent', 'external', 'builtin'] },
      ],
    },
    // Upstream case 0:4.
    {
      code: "\n        var path = require('path');\n        var _ = require('lodash');\n        var async = require('async');\n        var fs = require('f' + 's');",
    },
    // Upstream case 0:5.
    {
      code: "\n        var path = require('path');\n        var result = add(1, 2);\n        var _ = require('lodash');",
    },
    // Upstream case 0:6.
    {
      code: "\n        var index = require('./');\n        function foo() {\n          var fs = require('fs');\n        }\n        () => require('fs');\n        if (a) {\n          require('fs');\n        }",
    },
    // Upstream case 0:7.
    {
      code: "\n        const foo = [\n          require('./foo'),\n          require('fs'),\n        ]",
    },
    // Upstream case 0:8.
    {
      code: "const foo = `${require('./a')} ${require('fs')}`",
    },
    // Upstream case 0:9.
    {
      code: "\n        var unknown1 = require('/unknown1');\n        var fs = require('fs');\n        var unknown2 = require('/unknown2');\n        var async = require('async');\n        var unknown3 = require('/unknown3');\n        var foo = require('../foo');\n        var unknown4 = require('/unknown4');\n        var bar = require('../foo/bar');\n        var unknown5 = require('/unknown5');\n        var parent = require('../');\n        var unknown6 = require('/unknown6');\n        var foo = require('./foo');\n        var unknown7 = require('/unknown7');\n        var index = require('./');\n        var unknown8 = require('/unknown8');\n    ",
    },
    // Upstream case 0:10.
    {
      code: "\n        require('./foo');\n        require('fs');\n        var path = require('path');\n    ",
    },
    // Upstream case 0:11.
    {
      code: "\n        import './foo';\n        import 'fs';\n        import path from 'path';\n    ",
    },
    // Upstream case 0:12.
    {
      code: '\n        function add(a, b) {\n          return a + b;\n        }\n        var foo;\n    ',
    },
    // Upstream case 0:13.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n        var relParent3 = require('../');\n        var async = require('async');\n        var relParent1 = require('../foo');\n      ",
      options: [
        {
          groups: [
            ['builtin', 'index'],
            ['sibling', 'parent', 'external'],
          ],
        },
      ],
    },
    // Upstream case 0:14.
    {
      code: "\n        import async from 'async';\n        import fs from 'fs';\n        import path from 'path';\n\n        import index from '.';\n        import relParent3 from '../';\n        import relParent1 from '../foo';\n        import sibling from './foo';\n      ",
      options: [
        {
          groups: [['builtin', 'external']],
          alphabetize: { order: 'asc', caseInsensitive: true },
        },
      ],
    },
    // Upstream case 0:15.
    {
      code: "\n      import { fooz } from '../baz.js'\n      import { foo } from './bar.js'\n      ",
      options: [
        {
          alphabetize: { order: 'asc', caseInsensitive: true },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            'object',
          ],
          'newlines-between': 'always',
          warnOnUnassignedImports: true,
        },
      ],
    },
    // Upstream case 0:16.
    {
      code: "\n        var index = require('./');\n        var path = require('path');\n      ",
      options: [{ groups: ['index', ['sibling', 'parent', 'external']] }],
    },
    // Upstream case 0:17.
    {
      code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        import sibling, {foo3} from './foo';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var index = require('./');\n      ",
    },
    // Upstream case 0:18.
    {
      code: "\n          import async, {foo1} from 'async';\n          import relParent2, {foo2} from '../foo/bar';\n          import sibling, {foo3} from './foo';\n          var fs = require('fs');\n          var util = require(\"util\");\n          var relParent1 = require('../foo');\n          var relParent3 = require('../');\n          var index = require('./');\n        ",
    },
    // Upstream case 0:19.
    {
      code: '\n          export import CreateSomething = _CreateSomething;\n        ',
    },
    // Upstream case 0:20.
    {
      code: "\n        import fs from 'fs';\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n        import { add } from './helper';",
      options: [
        {
          groups: [
            'builtin',
            'external',
            'unknown',
            'parent',
            'sibling',
            'index',
          ],
        },
      ],
    },
    // Upstream case 0:21.
    {
      code: "\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n        import fs from 'fs';\n        import { add } from './helper';",
      options: [
        {
          groups: [
            'unknown',
            'builtin',
            'external',
            'parent',
            'sibling',
            'index',
          ],
        },
      ],
    },
    // Upstream case 0:22.
    {
      code: "\n        import fs from 'fs';\n\n        import { Input } from '-/components/Input';\n        import { Button } from '-/components/Button';\n\n        import p from '..';\n        import q from '../';\n\n        import { add } from './helper';\n\n        import i from '.';\n        import j from './';",
      options: [
        {
          'newlines-between': 'always',
          groups: [
            'builtin',
            'external',
            'unknown',
            'parent',
            'sibling',
            'index',
          ],
        },
      ],
    },
    // Upstream case 0:23.
    {
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import { Button } from '#/components/Button';\n        import { add } from './helper';",
      options: [
        {
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'after' },
            { pattern: '#/**', group: 'external', position: 'after' },
          ],
        },
      ],
    },
    // Upstream case 0:24.
    {
      code: "\n        import fs from 'fs';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        import { Button } from '#/components/Button';\n        import _ from 'lodash';\n        import { add } from './helper';",
      options: [
        {
          pathGroups: [
            { pattern: '~/**', group: 'external' },
            { pattern: '#/**', group: 'external' },
          ],
        },
      ],
    },
    // Upstream case 0:25.
    {
      code: "\n        import fs from 'fs';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '#/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';",
      options: [
        {
          'newlines-between': 'always',
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'before' },
            { pattern: '#/**', group: 'external', position: 'before' },
          ],
        },
      ],
    },
    // Upstream case 0:26.
    {
      code: "\n        import fs from 'fs';\n\n        import _ from 'lodash';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '!/components/Button';\n\n        import { add } from './helper';",
      options: [
        {
          'newlines-between': 'always',
          pathGroups: [
            { pattern: '~/**', group: 'external', position: 'after' },
            {
              pattern: '!/**',
              patternOptions: { nonegate: true },
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
    },
    // Upstream case 0:27.
    {
      code: "\n        import fs from 'fs';\n\n        import { Input } from '@app/components/Input';\n\n        import { Button } from '@app2/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';",
      options: [
        {
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['builtin'],
          pathGroups: [
            { pattern: '@app/**', group: 'external', position: 'before' },
            { pattern: '@app2/**', group: 'external', position: 'before' },
          ],
        },
      ],
    },
    // Upstream case 0:28.
    {
      code: "\n        import fs from 'fs';\n        import external from 'external';\n        import externalTooPlease from './make-me-external';\n\n        import sibling from './sibling';",
      options: [
        {
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          pathGroups: [{ pattern: './make-me-external', group: 'external' }],
          groups: [
            ['builtin', 'external'],
            'internal',
            'parent',
            'sibling',
            'index',
          ],
        },
      ],
    },
    // Upstream case 0:29.
    {
      code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ",
      options: [{ 'newlines-between': 'always' }],
      settings: {
        'import/external-module-folders': ['node_modules', 'symlinked-module'],
      },
    },
    // Upstream case 0:30.
    {
      code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ",
      options: [{ 'newlines-between': 'always' }],
      settings: {
        'import/external-module-folders': ['node_modules', 'symlinked-module'],
      },
    },
    // Upstream case 0:31.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n\n\n        var sibling = require('./foo');\n\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'always',
        },
      ],
    },
    // Upstream case 0:32.
    {
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'never',
        },
      ],
    },
    // Upstream case 0:33.
    {
      code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n      var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'ignore',
        },
      ],
    },
    // Upstream case 0:34.
    {
      code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n\n      var async = require('async');\n      ",
      options: [
        { groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']] },
      ],
    },
    // Upstream case 0:35.
    {
      code: "\n        import path from 'path';\n\n        import {\n            I,\n            Want,\n            Couple,\n            Imports,\n            Here\n        } from 'bar';\n        import external from 'external'\n      ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 0:36.
    {
      code: "\n        import path from 'path';\n        import net\n          from 'net';\n\n        import external from 'external'\n      ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 0:37.
    {
      code: "\n        import foo\n          from '../../../../this/will/be/very/long/path/and/therefore/this/import/has/to/be/in/two/lines';\n\n        import bar\n          from './sibling';\n      ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 0:38.
    {
      code: "\n        import path from 'path';\n\n        import 'loud-rejection';\n        import 'something-else';\n\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 0:39.
    {
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [{ 'newlines-between': 'never' }],
    },
    // Upstream case 0:40.
    {
      code: "\n        var path = require('path');\n\n        require('loud-rejection');\n        require('something-else');\n\n        var _ = require('lodash');\n      ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 0:41.
    {
      code: "\n        var path = require('path');\n        require('loud-rejection');\n        require('something-else');\n        var _ = require('lodash');\n      ",
      options: [{ 'newlines-between': 'never' }],
    },
    // Upstream case 0:42.
    {
      code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ",
      options: [{ 'newlines-between': 'never' }],
    },
    // Upstream case 0:43.
    {
      code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 0:44.
    {
      code: "\n        var fs = require('fs');\n        var path = require('path');\n\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n        var relParent2 = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ",
      options: [{ 'newlines-between': 'always-and-inside-groups' }],
    },
    // Upstream case 0:45.
    {
      code: "\n        var fs = require('fs');\n        var path = require('path');\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent2 } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 0:46.
    {
      code: "\n        import a from 'foo';\n        import b from 'bar';\n\n        import index from './';\n      ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'ignore' } },
      ],
    },
    // Upstream case 0:47.
    {
      code: "\n        import c from 'Bar';\n        import b from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
    },
    // Upstream case 0:48.
    {
      code: "\n        import a from 'foo';\n        import b from 'bar';\n        import c from 'Bar';\n\n        import index from './';\n      ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'desc' } },
      ],
    },
    // Upstream case 0:49.
    {
      code: '\n        import a from "foo";\n        import c from "foo/bar";\n        import d from "foo/barfoo";\n        import b from "foo-bar";\n      ',
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:50.
    {
      code: '\n        import a from "foo";\n        import c from "foo/foobar/bar";\n        import d from "foo/foobar/barfoo";\n        import b from "foo-bar";\n      ',
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:51.
    {
      code: '\n        import b from "foo-bar";\n        import d from "foo/barfoo";\n        import c from "foo/bar";\n        import a from "foo";\n      ',
      options: [{ alphabetize: { order: 'desc' } }],
    },
    // Upstream case 0:52.
    {
      code: '\n        import b from "foo-bar";\n        import c from "foo,bar";\n        import d from "foo/barfoo";\n        import a from "foo";',
      options: [{ alphabetize: { order: 'desc' } }],
    },
    // Upstream case 0:53.
    {
      code: "\n        import b from 'Bar';\n        import c from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: { order: 'asc' },
          'newlines-between': 'always',
        },
      ],
    },
    // Upstream case 0:54.
    {
      code: "\n        import { hello } from './hello';\n        import { int } from './int';\n        const blah = require('./blah');\n        const { cello } = require('./cello');\n      ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:55.
    {
      code: "\n        import React from 'react';\n        import { BrowserRouter } from 'react-router-dom';\n      ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:56.
    {
      code: "\n        import { UserInputError } from 'apollo-server-express';\n\n        import { new as assertNewEmail } from '~/Assertions/Email';\n      ",
      options: [
        {
          alphabetize: { caseInsensitive: true, order: 'asc' },
          pathGroups: [{ pattern: '~/*', group: 'internal' }],
          groups: [
            'builtin',
            'external',
            'internal',
            'parent',
            'sibling',
            'index',
          ],
          'newlines-between': 'always',
        },
      ],
    },
    // Upstream case 0:57.
    {
      code: "\n        import { ReactElement, ReactNode } from 'react';\n\n        import { util } from 'Internal/lib';\n\n        import { parent } from '../parent';\n\n        import { sibling } from './sibling';\n      ",
      options: [
        {
          alphabetize: { caseInsensitive: true, order: 'asc' },
          pathGroups: [{ pattern: 'Internal/**/*', group: 'internal' }],
          groups: [
            'builtin',
            'external',
            'internal',
            'parent',
            'sibling',
            'index',
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 0:58.
    {
      code: "\n        import fs from 'fs';\n\n        import express from 'express';\n\n        import service from '@/api/service';\n\n        import fooParent from '../foo';\n\n        import fooSibling from './foo';\n\n        import index from './';\n\n        import internalDoesNotExistSoIsUnknown from '@/does-not-exist';\n      ",
      options: [
        {
          groups: [
            'builtin',
            'external',
            'internal',
            'parent',
            'sibling',
            'index',
            'unknown',
          ],
          'newlines-between': 'always',
        },
      ],
    },
    // Upstream case 0:59.
    {
      code: "\n        import A from 'a';\n\n        import C from 'c';\n\n        import B from 'b';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: true,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            { pattern: 'a', group: 'external', position: 'before' },
            { pattern: 'b', group: 'external', position: 'after' },
          ],
        },
      ],
    },
    // Upstream case 0:60.
    {
      code: "\n        import A from 'a';\n        import C from 'c';\n        import B from 'b';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            { pattern: 'a', group: 'external', position: 'before' },
            { pattern: 'b', group: 'external', position: 'after' },
          ],
        },
      ],
    },
    // Upstream case 0:61.
    {
      code: "\n        import A from 'a';\n\n        import b from './b';\n        import B from './B';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            { pattern: 'a', group: 'external' },
            { pattern: 'b', group: 'internal', position: 'before' },
          ],
        },
      ],
    },
    // Upstream case 0:62.
    {
      code: '\n        import A from "baz";\n        import B from "Bar";\n        import C from "Foo";\n\n        import D from "..";\n        import E from "../";\n        import F from "../baz";\n        import G from "../Bar";\n        import H from "../Foo";\n\n        import I from ".";\n        import J from "./baz";\n        import K from "./Bar";\n        import L from "./Foo";\n      ',
      options: [
        {
          alphabetize: { caseInsensitive: false, order: 'asc' },
          'newlines-between': 'always',
          groups: [
            ['builtin', 'external', 'internal', 'unknown', 'object', 'type'],
            'parent',
            ['sibling', 'index'],
          ],
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            { pattern: './', group: 'sibling', position: 'before' },
            { pattern: '.', group: 'sibling', position: 'before' },
            { pattern: '..', group: 'parent', position: 'before' },
            { pattern: '../', group: 'parent', position: 'before' },
            { pattern: '[a-z]*', group: 'external', position: 'before' },
            { pattern: '../[a-z]*', group: 'parent', position: 'before' },
            { pattern: './[a-z]*', group: 'sibling', position: 'before' },
          ],
        },
      ],
    },
    // Upstream case 0:63.
    {
      code: "\n        import B from './B';\n        import b from './b';\n      ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    // Upstream case 0:64.
    {
      code: '\n        import { a, B as C, Z } from \'./Z\';\n        const { D, n: c, Y } = require(\'./Z\');\n        export { C, D };\n        export { A, B, C as default } from "./Z";\n\n        const { ["ignore require-statements with non-identifier imports"]: z, d } = require("./Z");\n        exports = { ["ignore exports statements with non-identifiers"]: Z, D };\n      ',
      options: [
        { named: true, alphabetize: { order: 'asc', caseInsensitive: true } },
      ],
    },
    // Upstream case 0:65.
    {
      code: "\n        const { b, A } = require('./Z');\n      ",
      options: [{ named: true, alphabetize: { order: 'desc' } }],
    },
    // Upstream case 0:66.
    {
      code: '\n        import { A, B } from "./Z";\n        export { Z, A } from "./Z";\n        export { N, P } from "./Z";\n        const { X, Y } = require("./Z");\n      ',
      options: [{ named: { require: true, import: true, export: false } }],
    },
    // Upstream case 0:67.
    {
      code: '\n        import { B, A } from "./Z";\n        const { D, C } = require("./Z");\n        export { B, A } from "./Z";\n      ',
      options: [{ named: { require: false, import: false, export: false } }],
    },
    // Upstream case 0:68.
    {
      code: '\n        import { B, A, R } from "foo";\n        const { D, O, G } = require("tunes");\n        export { B, A, Z } from "foo";\n      ',
      options: [{ named: { enabled: false } }],
    },
    // Upstream case 0:69.
    {
      code: '\n        import { A as A, A as B, A as C } from "./Z";\n        const { a, a: b, a: c } = require("./Z");\n      ',
      options: [{ named: true }],
    },
    // Upstream case 0:70.
    {
      code: '\n        import { A, B, C } from "./Z";\n        exports = { A, B, C };\n        module.exports = { a: A, b: B, c: C };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:71.
    {
      code: '\n        module.exports.A = { };\n        module.exports.A.B = { };\n        module.exports.B = { };\n        exports.C = { };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:72.
    {
      code: '\n        var exports = null;\n        var module = null;\n        exports = { };\n        module = { };\n        module.exports = { };\n        module.exports.U = { };\n        module.exports.N = { };\n        module.exports.C = { };\n        exports.L = { };\n        exports.E = { };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
    },
    // Upstream case 0:73.
    {
      code: '\n        exports["B"] = { };\n        exports["C"] = { };\n        exports["A"] = { };\n      ',
      options: [{ named: { cjsExports: true }, alphabetize: { order: 'asc' } }],
    },
    // Upstream case 1:0.
    {
      code: "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
    },
    // Upstream case 1:1.
    {
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'desc' } },
      ],
    },
    // Upstream case 1:2.
    {
      code: "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'asc' },
        },
      ],
    },
    // Upstream case 1:3.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { C } from 'dirA/Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
        },
      ],
    },
    // Upstream case 1:4.
    {
      code: "\n              import c from 'Bar';\n              import type { A } from 'foo';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 1:5.
    {
      code: "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'desc' },
        },
      ],
    },
    // Upstream case 1:6.
    {
      code: "\n              import { Partner } from '@models/partner/partner';\n              import { PartnerId } from '@models/partner/partner-id';\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 1:7.
    {
      code: "\n              import { serialize, parse, mapFieldErrors } from '@vtaits/form-schema';\n              import type { GetFieldSchema } from '@vtaits/form-schema';\n              import { useMemo, useCallback } from 'react';\n              import type { ReactElement, ReactNode } from 'react';\n              import { Form } from 'react-final-form';\n              import type { FormProps as FinalFormProps } from 'react-final-form';\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 1:8.
    {
      code: "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 1:9, 2:9.
    {
      code: '\n              import { useLazyQuery, useQuery } from "@apollo/client";\n              import { useEffect } from "react";\n            ',
      options: [
        {
          groups: [
            'builtin',
            'external',
            'internal',
            'parent',
            'sibling',
            'index',
            'object',
            'type',
          ],
          pathGroups: [
            { pattern: 'react', group: 'external', position: 'before' },
          ],
          'newlines-between': 'always',
          alphabetize: { order: 'asc', caseInsensitive: true },
        },
      ],
    },
    // Upstream case 1:10, 2:10.
    {
      code: "\n                import express from 'express';\n                import log4js from 'log4js';\n                import chpro from 'node:child_process';\n                // import fsp from 'node:fs/promises';\n              ",
      options: [
        {
          groups: [
            [
              'builtin',
              'external',
              'internal',
              'parent',
              'sibling',
              'index',
              'object',
              'type',
            ],
          ],
        },
      ],
    },
    // Upstream case 1:11.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 1:12.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: false,
        },
      ],
    },
    // Upstream case 1:13.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:14.
    {
      code: "\n              import c from 'Bar';\n              import type { AA } from 'abc';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 1:15.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:16.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'never',
          'newlines-between-types': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:17.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:18.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'never',
          'newlines-between-types': 'ignore',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:19.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:20.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: { order: 'asc', caseInsensitive: true },
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:21, 1:42.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'always',
          'newlines-between-types': 'ignore',
        },
      ],
    },
    // Upstream case 1:22, 1:43.
    {
      code: '\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
          'newlines-between': 'always',
          'newlines-between-types': 'ignore',
        },
      ],
    },
    // Upstream case 1:23, 1:44.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n\n              import type C from "../foo.js";\n\n              import type D from "./bar.js";\n\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'never',
          'newlines-between-types': 'always',
        },
      ],
    },
    // Upstream case 1:24, 1:45.
    {
      code: '\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n\n              import type C from "../foo.js";\n\n              import type D from "./bar.js";\n\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
          'newlines-between': 'never',
          'newlines-between-types': 'always',
        },
      ],
    },
    // Upstream case 1:25.
    {
      code: "\n              import fs from 'fs';\n\n              import '@scoped/package';\n              import type { B } from 'fs';\n\n              import type { A1 } from '/bad/bad/bad/bad';\n              import './a/b/c';\n              import type { A2 } from '/bad/bad/bad/bad';\n              import type { A3 } from '/bad/bad/bad/bad';\n              import type { D1 } from '/bad/bad/not/good';\n              import type { D2 } from '/bad/bad/not/good';\n              import type { D3 } from '/bad/bad/not/good';\n\n              import type { C } from '@something/else';\n\n              import type { E } from './index.js';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['builtin', 'type', 'unknown', 'external'],
          sortTypesGroup: true,
          'newlines-between': 'always',
        },
      ],
    },
    // Upstream case 1:26.
    {
      code: "\n              // https://prettier.io/docs/en/options.html\n\n              module.exports = {\n                  ...require('@xxxx/.prettierrc.js'),\n              };\n            ",
      options: [{ named: { enabled: true } }],
    },
    // Upstream case 1:27.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:28.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'never',
        },
      ],
    },
    // Upstream case 1:29.
    {
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n\n              import { createDebugLogger } from 'multiverse+rejoinder';\n\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n\n              import {\n                $artificiallyInvoked,\n                $canonical,\n                $exists,\n                $genesis\n              } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: { enabled: true, types: 'types-last' },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            ['object', 'type'],
          ],
          pathGroups: [
            {
              pattern: 'multiverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'rootverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'universe{*,*/**}',
              group: 'external',
              position: 'after',
            },
          ],
          distinctGroup: true,
          pathGroupsExcludedImportTypes: ['builtin', 'object'],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:30.
    {
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n              import { createDebugLogger } from 'multiverse+rejoinder';\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n              import { $artificiallyInvoked } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: { enabled: true, types: 'types-last' },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            ['object', 'type'],
          ],
          pathGroups: [
            {
              pattern: 'multiverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'rootverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'universe{*,*/**}',
              group: 'external',
              position: 'after',
            },
          ],
          distinctGroup: true,
          pathGroupsExcludedImportTypes: ['builtin', 'object'],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:31.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:32.
    {
      code: "\n              import assert from 'assert';\n              import { isNativeError } from 'util/types';\n\n              import { runNoRejectOnBadExit } from '@-xun/run';\n              import { TrialError } from 'named-app-errors';\n              import { resolve as resolverLibrary } from 'resolve.exports';\n\n              import { toAbsolutePath, type AbsolutePath } from 'rootverse+project-utils:src/fs.ts';\n\n              import type { PackageJson } from 'type-fest';\n              // Some comment about remembering to do something\n              import type { XPackageJson } from 'rootverse:src/assets/config/_package.json.ts';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: { enabled: true, types: 'types-last' },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            ['object', 'type'],
          ],
          pathGroups: [
            {
              pattern: 'rootverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
          ],
          distinctGroup: true,
          pathGroupsExcludedImportTypes: ['builtin', 'object'],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:33.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 1:34.
    {
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always-and-inside-groups' }],
    },
    // Upstream case 1:35.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'never' }],
    },
    // Upstream case 1:36.
    {
      code: "\n              import blist2 from 'blist';\n              import blist from 'BList';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [{ alphabetize: { order: 'asc', caseInsensitive: true } }],
    },
    // Upstream case 1:37.
    {
      code: "\n              import blist from 'BList';\n              import blist2 from 'blist';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [{ alphabetize: { order: 'asc', caseInsensitive: false } }],
    },
    // Upstream case 1:38.
    {
      code: "\n              import { apply, compose } from 'xcompose';\n            ",
      options: [{ named: true, alphabetize: { order: 'asc' } }],
    },
    // Upstream case 1:39.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import './styles.css';\n            ",
      options: [{ warnOnUnassignedImports: true }],
    },
    // Upstream case 1:40.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: { order: 'asc' },
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:41.
    {
      code: '\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 1:46.
    {
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n\n              var async = require('async');\n\n              var relParent1 = require('../foo');\n\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n\n              var relParent3 = require('../bar');\n\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:47.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 1:48.
    {
      code: "\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['type', 'external', 'internal', 'index'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:0.
    {
      code: "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'asc' } },
      ],
    },
    // Upstream case 2:1.
    {
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ",
      options: [
        { groups: ['external', 'index'], alphabetize: { order: 'desc' } },
      ],
    },
    // Upstream case 2:2.
    {
      code: "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'asc' },
        },
      ],
    },
    // Upstream case 2:3.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { C } from 'dirA/Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
        },
      ],
    },
    // Upstream case 2:4.
    {
      code: "\n              import c from 'Bar';\n              import type { A } from 'foo';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 2:5.
    {
      code: "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: { order: 'desc' },
        },
      ],
    },
    // Upstream case 2:6.
    {
      code: "\n              import { Partner } from '@models/partner/partner';\n              import { PartnerId } from '@models/partner/partner-id';\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 2:7.
    {
      code: "\n              import { serialize, parse, mapFieldErrors } from '@vtaits/form-schema';\n              import type { GetFieldSchema } from '@vtaits/form-schema';\n              import { useMemo, useCallback } from 'react';\n              import type { ReactElement, ReactNode } from 'react';\n              import { Form } from 'react-final-form';\n              import type { FormProps as FinalFormProps } from 'react-final-form';\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 2:8.
    {
      code: "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ",
      options: [{ alphabetize: { order: 'asc' } }],
    },
    // Upstream case 2:11.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 2:12.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: false,
        },
      ],
    },
    // Upstream case 2:13.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:14.
    {
      code: "\n              import c from 'Bar';\n              import type { AA } from 'abc';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    // Upstream case 2:15.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:16.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'never',
          'newlines-between-types': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:17.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'always',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:18.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'never',
          'newlines-between-types': 'ignore',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:19.
    {
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [{ pattern: 'dirA/**', group: 'internal' }],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:20.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: { order: 'asc', caseInsensitive: true },
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:21, 2:42.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'always',
          'newlines-between-types': 'ignore',
        },
      ],
    },
    // Upstream case 2:22, 2:43.
    {
      code: '\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
          'newlines-between': 'always',
          'newlines-between-types': 'ignore',
        },
      ],
    },
    // Upstream case 2:23, 2:44.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n\n              import type C from "../foo.js";\n\n              import type D from "./bar.js";\n\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          sortTypesGroup: true,
          'newlines-between': 'never',
          'newlines-between-types': 'always',
        },
      ],
    },
    // Upstream case 2:24, 2:45.
    {
      code: '\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n\n              import type C from "../foo.js";\n\n              import type D from "./bar.js";\n\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
          'newlines-between': 'never',
          'newlines-between-types': 'always',
        },
      ],
    },
    // Upstream case 2:25.
    {
      code: "\n              import fs from 'fs';\n\n              import '@scoped/package';\n              import type { B } from 'fs';\n\n              import type { A1 } from '/bad/bad/bad/bad';\n              import './a/b/c';\n              import type { A2 } from '/bad/bad/bad/bad';\n              import type { A3 } from '/bad/bad/bad/bad';\n              import type { D1 } from '/bad/bad/not/good';\n              import type { D2 } from '/bad/bad/not/good';\n              import type { D3 } from '/bad/bad/not/good';\n\n              import type { C } from '@something/else';\n\n              import type { E } from './index.js';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['builtin', 'type', 'unknown', 'external'],
          sortTypesGroup: true,
          'newlines-between': 'always',
        },
      ],
    },
    // Upstream case 2:26.
    {
      code: "\n              // https://prettier.io/docs/en/options.html\n\n              module.exports = {\n                  ...require('@xxxx/.prettierrc.js'),\n              };\n            ",
      options: [{ named: { enabled: true } }],
    },
    // Upstream case 2:27.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:28.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'never',
        },
      ],
    },
    // Upstream case 2:29.
    {
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n\n              import { createDebugLogger } from 'multiverse+rejoinder';\n\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n\n              import {\n                $artificiallyInvoked,\n                $canonical,\n                $exists,\n                $genesis\n              } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: { enabled: true, types: 'types-last' },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            ['object', 'type'],
          ],
          pathGroups: [
            {
              pattern: 'multiverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'rootverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'universe{*,*/**}',
              group: 'external',
              position: 'after',
            },
          ],
          distinctGroup: true,
          pathGroupsExcludedImportTypes: ['builtin', 'object'],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:30.
    {
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n              import { createDebugLogger } from 'multiverse+rejoinder';\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n              import { $artificiallyInvoked } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: { enabled: true, types: 'types-last' },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            ['object', 'type'],
          ],
          pathGroups: [
            {
              pattern: 'multiverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'rootverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'universe{*,*/**}',
              group: 'external',
              position: 'after',
            },
          ],
          distinctGroup: true,
          pathGroupsExcludedImportTypes: ['builtin', 'object'],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:31.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:32.
    {
      code: "\n              import assert from 'assert';\n              import { isNativeError } from 'util/types';\n\n              import { runNoRejectOnBadExit } from '@-xun/run';\n              import { TrialError } from 'named-app-errors';\n              import { resolve as resolverLibrary } from 'resolve.exports';\n\n              import { toAbsolutePath, type AbsolutePath } from 'rootverse+project-utils:src/fs.ts';\n\n              import type { PackageJson } from 'type-fest';\n              // Some comment about remembering to do something\n              import type { XPackageJson } from 'rootverse:src/assets/config/_package.json.ts';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: { enabled: true, types: 'types-last' },
          groups: [
            'builtin',
            'external',
            'internal',
            ['parent', 'sibling', 'index'],
            ['object', 'type'],
          ],
          pathGroups: [
            {
              pattern: 'rootverse{*,*/**}',
              group: 'external',
              position: 'after',
            },
          ],
          distinctGroup: true,
          pathGroupsExcludedImportTypes: ['builtin', 'object'],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:33.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always' }],
    },
    // Upstream case 2:34.
    {
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'always-and-inside-groups' }],
    },
    // Upstream case 2:35.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [{ 'newlines-between': 'never' }],
    },
    // Upstream case 2:36.
    {
      code: "\n              import blist2 from 'blist';\n              import blist from 'BList';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [{ alphabetize: { order: 'asc', caseInsensitive: true } }],
    },
    // Upstream case 2:37.
    {
      code: "\n              import blist from 'BList';\n              import blist2 from 'blist';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [{ alphabetize: { order: 'asc', caseInsensitive: false } }],
    },
    // Upstream case 2:38.
    {
      code: "\n              import { apply, compose } from 'xcompose';\n            ",
      options: [{ named: true, alphabetize: { order: 'asc' } }],
    },
    // Upstream case 2:39.
    {
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import './styles.css';\n            ",
      options: [{ warnOnUnassignedImports: true }],
    },
    // Upstream case 2:40.
    {
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: { order: 'asc' },
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:41.
    {
      code: '\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
        },
      ],
    },
    // Upstream case 2:46.
    {
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n\n              var async = require('async');\n\n              var relParent1 = require('../foo');\n\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n\n              var relParent3 = require('../bar');\n\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:47.
    {
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    // Upstream case 2:48.
    {
      code: "\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: { order: 'asc' },
          groups: ['type', 'external', 'internal', 'index'],
          pathGroups: [
            { pattern: 'dirA/**', group: 'internal', position: 'after' },
            { pattern: 'dirB/**', group: 'internal', position: 'before' },
            { pattern: 'dirC/**', group: 'internal' },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
  ],
  invalid: [],
});
