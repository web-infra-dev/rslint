// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('order', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "\n        import { fromAbi } from '../../.checkpoint/models';\n        import { handleVotingPowerValidationMetadata } from '../common/ipfs';\n        import ExecutionStrategyAbi from './abis/executionStrategy.json';\n        import L1AvatarExecutionStrategyAbi from './abis/l1/L1AvatarExectionStrategy';\n        import { FullConfig } from './config';\n      ",
      options: [
        {
          groups: ['builtin', 'external', 'internal', 'type'],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 0',
      code: "import { Buffer } from 'buffer';\n\nimport { includes, isUndefined } from 'lodash-es';",
      options: [
        {
          groups: ['builtin', 'external', 'internal'],
          alphabetize: {
            order: 'desc',
            caseInsensitive: true,
          },
          'newlines-between': 'always',
        },
      ],
      settings: {
        'import/external-module-folders': ['src/order-internal-modules'],
      },
    },
    {
      name: 'run 1 valid 1',
      code: "import checkpoint from '../../checkpoint/models';\nimport common from '../common/files';\nimport avatar from './adapters/avatar';\nimport execution from './adapters/execution';\nimport config from './config';",
      options: [
        {
          groups: ['builtin', 'external', 'internal', 'type'],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 2',
      code: "\n        var fs = require('fs');\n        var async = require('async');\n        var relParent1 = require('../foo');\n        var relParent2 = require('../foo/bar');\n        var relParent3 = require('../');\n        var relParent4 = require('..');\n        var sibling = require('./foo');\n        var index = require('./');",
    },
    {
      name: 'run 1 valid 3',
      code: "\n        import fs from 'fs';\n        import async, {foo1} from 'async';\n        import relParent1 from '../foo';\n        import relParent2, {foo2} from '../foo/bar';\n        import relParent3 from '../';\n        import sibling, {foo3} from './foo';\n        import index from './';",
    },
    {
      name: 'run 1 valid 4',
      code: "\n        var fs = require('fs');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var async = require('async');",
    },
    {
      name: 'run 1 valid 5',
      code: "\n        var index = require('./');\n        var sibling = require('./foo');\n        var relParent3 = require('../');\n        var relParent2 = require('../foo/bar');\n        var relParent1 = require('../foo');\n        var async = require('async');\n        var fs = require('fs');\n      ",
      options: [
        {
          groups: ['index', 'sibling', 'parent', 'external', 'builtin'],
        },
      ],
    },
    {
      name: 'run 1 valid 6',
      code: "\n        var path = require('path');\n        var _ = require('lodash');\n        var async = require('async');\n        var fs = require('f' + 's');",
    },
    {
      name: 'run 1 valid 7',
      code: "\n        var path = require('path');\n        var result = add(1, 2);\n        var _ = require('lodash');",
    },
    {
      name: 'run 1 valid 8',
      code: "\n        var index = require('./');\n        function foo() {\n          var fs = require('fs');\n        }\n        () => require('fs');\n        if (a) {\n          require('fs');\n        }",
    },
    {
      name: 'run 1 valid 9',
      code: "\n        const foo = [\n          require('./foo'),\n          require('fs'),\n        ]",
    },
    {
      name: 'run 1 valid 10',
      code: "const foo = `${require('./a')} ${require('fs')}`",
    },
    {
      name: 'run 1 valid 11',
      code: "\n        var unknown1 = require('/unknown1');\n        var fs = require('fs');\n        var unknown2 = require('/unknown2');\n        var async = require('async');\n        var unknown3 = require('/unknown3');\n        var foo = require('../foo');\n        var unknown4 = require('/unknown4');\n        var bar = require('../foo/bar');\n        var unknown5 = require('/unknown5');\n        var parent = require('../');\n        var unknown6 = require('/unknown6');\n        var foo = require('./foo');\n        var unknown7 = require('/unknown7');\n        var index = require('./');\n        var unknown8 = require('/unknown8');\n    ",
    },
    {
      name: 'run 1 valid 12',
      code: "\n        require('./foo');\n        require('fs');\n        var path = require('path');\n    ",
    },
    {
      name: 'run 1 valid 13',
      code: "\n        import './foo';\n        import 'fs';\n        import path from 'path';\n    ",
    },
    {
      name: 'run 1 valid 14',
      code: '\n        function add(a, b) {\n          return a + b;\n        }\n        var foo;\n    ',
    },
    {
      name: 'run 1 valid 15',
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
    {
      name: 'run 1 valid 16',
      code: "\n        import async from 'async';\n        import fs from 'fs';\n        import path from 'path';\n\n        import index from '.';\n        import relParent3 from '../';\n        import relParent1 from '../foo';\n        import sibling from './foo';\n      ",
      options: [
        {
          groups: [['builtin', 'external']],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 17',
      code: "\n      import { fooz } from '../baz.js'\n      import { foo } from './bar.js'\n      ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
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
    {
      name: 'run 1 valid 18',
      code: "\n        var index = require('./');\n        var path = require('path');\n      ",
      options: [
        {
          groups: ['index', ['sibling', 'parent', 'external']],
        },
      ],
    },
    {
      name: 'run 1 valid 19',
      code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        import sibling, {foo3} from './foo';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var index = require('./');\n      ",
    },
    {
      name: 'run 1 valid 20',
      code: "\n          import async, {foo1} from 'async';\n          import relParent2, {foo2} from '../foo/bar';\n          import sibling, {foo3} from './foo';\n          var fs = require('fs');\n          var util = require(\"util\");\n          var relParent1 = require('../foo');\n          var relParent3 = require('../');\n          var index = require('./');\n        ",
    },
    {
      name: 'run 1 valid 21',
      code: '\n          export import CreateSomething = _CreateSomething;\n        ',
    },
    {
      name: 'run 1 valid 22',
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
    {
      name: 'run 1 valid 23',
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
    {
      name: 'run 1 valid 24',
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
    {
      name: 'run 1 valid 25',
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import { Button } from '#/components/Button';\n        import { add } from './helper';",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '#/**',
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 26',
      code: "\n        import fs from 'fs';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        import { Button } from '#/components/Button';\n        import _ from 'lodash';\n        import { add } from './helper';",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
            },
            {
              pattern: '#/**',
              group: 'external',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 27',
      code: "\n        import fs from 'fs';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '#/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';",
      options: [
        {
          'newlines-between': 'always',
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'before',
            },
            {
              pattern: '#/**',
              group: 'external',
              position: 'before',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 28',
      code: "\n        import fs from 'fs';\n\n        import _ from 'lodash';\n\n        import { Input } from '~/components/Input';\n\n        import { Button } from '!/components/Button';\n\n        import { add } from './helper';",
      options: [
        {
          'newlines-between': 'always',
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '!/**',
              patternOptions: {
                nonegate: true,
              },
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 29',
      code: "\n        import fs from 'fs';\n\n        import { Input } from '@app/components/Input';\n\n        import { Button } from '@app2/components/Button';\n\n        import _ from 'lodash';\n\n        import { add } from './helper';",
      options: [
        {
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['builtin'],
          pathGroups: [
            {
              pattern: '@app/**',
              group: 'external',
              position: 'before',
            },
            {
              pattern: '@app2/**',
              group: 'external',
              position: 'before',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 30',
      code: "\n        import fs from 'fs';\n        import external from 'external';\n        import externalTooPlease from './make-me-external';\n\n        import sibling from './sibling';",
      options: [
        {
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            {
              pattern: './make-me-external',
              group: 'external',
            },
          ],
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
    {
      name: 'run 1 valid 31',
      code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      settings: {
        'import/external-module-folders': ['node_modules', 'symlinked-module'],
      },
    },
    {
      name: 'run 1 valid 32',
      code: "\n        import _ from 'lodash';\n        import m from '@test-scope/some-module';\n\n        import bar from './bar';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      settings: {
        'import/external-module-folders': ['node_modules', 'symlinked-module'],
      },
    },
    {
      name: 'run 1 valid 33',
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n\n\n        var sibling = require('./foo');\n\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 34',
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 35',
      code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n      var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'ignore',
        },
      ],
    },
    {
      name: 'run 1 valid 36',
      code: "\n      var fs = require('fs');\n\n      var index = require('./');\n      var path = require('path');\n      var sibling = require('./foo');\n\n\n      var relParent1 = require('../foo');\n\n      var relParent3 = require('../');\n\n      var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
        },
      ],
    },
    {
      name: 'run 1 valid 37',
      code: "\n        import path from 'path';\n\n        import {\n            I,\n            Want,\n            Couple,\n            Imports,\n            Here\n        } from 'bar';\n        import external from 'external'\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 38',
      code: "\n        import path from 'path';\n        import net\n          from 'net';\n\n        import external from 'external'\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 39',
      code: "\n        import foo\n          from '../../../../this/will/be/very/long/path/and/therefore/this/import/has/to/be/in/two/lines';\n\n        import bar\n          from './sibling';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 40',
      code: "\n        import path from 'path';\n\n        import 'loud-rejection';\n        import 'something-else';\n\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 41',
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 42',
      code: "\n        var path = require('path');\n\n        require('loud-rejection');\n        require('something-else');\n\n        var _ = require('lodash');\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 43',
      code: "\n        var path = require('path');\n        require('loud-rejection');\n        require('something-else');\n        var _ = require('lodash');\n      ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 44',
      code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 45',
      code: "\n        var some = require('asdas');\n        var config = {\n          port: 4444,\n          runner: {\n            server_path: require('runner-binary').path,\n            cli_args: {\n                'webdriver.chrome.driver': require('browser-binary').path\n            }\n          }\n        }\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 46',
      code: "\n        var fs = require('fs');\n        var path = require('path');\n\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n        var relParent2 = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 47',
      code: "\n        var fs = require('fs');\n        var path = require('path');\n        var util = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent2 } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var sibling = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 48',
      code: "\n        import a from 'foo';\n        import b from 'bar';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'ignore',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 49',
      code: "\n        import c from 'Bar';\n        import b from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 50',
      code: "\n        import a from 'foo';\n        import b from 'bar';\n        import c from 'Bar';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 51',
      code: '\n        import a from "foo";\n        import c from "foo/bar";\n        import d from "foo/barfoo";\n        import b from "foo-bar";\n      ',
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 52',
      code: '\n        import a from "foo";\n        import c from "foo/foobar/bar";\n        import d from "foo/foobar/barfoo";\n        import b from "foo-bar";\n      ',
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 53',
      code: '\n        import b from "foo-bar";\n        import d from "foo/barfoo";\n        import c from "foo/bar";\n        import a from "foo";\n      ',
      options: [
        {
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 54',
      code: '\n        import b from "foo-bar";\n        import c from "foo,bar";\n        import d from "foo/barfoo";\n        import a from "foo";',
      options: [
        {
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 55',
      code: "\n        import b from 'Bar';\n        import c from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 56',
      code: "\n        import { hello } from './hello';\n        import { int } from './int';\n        const blah = require('./blah');\n        const { cello } = require('./cello');\n      ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 57',
      code: "\n        import React from 'react';\n        import { BrowserRouter } from 'react-router-dom';\n      ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 58',
      code: "\n        import { UserInputError } from 'apollo-server-express';\n\n        import { new as assertNewEmail } from '~/Assertions/Email';\n      ",
      options: [
        {
          alphabetize: {
            caseInsensitive: true,
            order: 'asc',
          },
          pathGroups: [
            {
              pattern: '~/*',
              group: 'internal',
            },
          ],
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
    {
      name: 'run 1 valid 59',
      code: "\n        import { ReactElement, ReactNode } from 'react';\n\n        import { util } from 'Internal/lib';\n\n        import { parent } from '../parent';\n\n        import { sibling } from './sibling';\n      ",
      options: [
        {
          alphabetize: {
            caseInsensitive: true,
            order: 'asc',
          },
          pathGroups: [
            {
              pattern: 'Internal/**/*',
              group: 'internal',
            },
          ],
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
    {
      name: 'run 1 valid 60',
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
    {
      name: 'run 1 valid 61',
      code: "\n        import A from 'a';\n\n        import C from 'c';\n\n        import B from 'b';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: true,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            {
              pattern: 'a',
              group: 'external',
              position: 'before',
            },
            {
              pattern: 'b',
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 62',
      code: "\n        import A from 'a';\n        import C from 'c';\n        import B from 'b';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            {
              pattern: 'a',
              group: 'external',
              position: 'before',
            },
            {
              pattern: 'b',
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 63',
      code: "\n        import A from 'a';\n\n        import b from './b';\n        import B from './B';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            {
              pattern: 'a',
              group: 'external',
            },
            {
              pattern: 'b',
              group: 'internal',
              position: 'before',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 64',
      code: '\n        import A from "baz";\n        import B from "Bar";\n        import C from "Foo";\n\n        import D from "..";\n        import E from "../";\n        import F from "../baz";\n        import G from "../Bar";\n        import H from "../Foo";\n\n        import I from ".";\n        import J from "./baz";\n        import K from "./Bar";\n        import L from "./Foo";\n      ',
      options: [
        {
          alphabetize: {
            caseInsensitive: false,
            order: 'asc',
          },
          'newlines-between': 'always',
          groups: [
            ['builtin', 'external', 'internal', 'unknown', 'object', 'type'],
            'parent',
            ['sibling', 'index'],
          ],
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            {
              pattern: './',
              group: 'sibling',
              position: 'before',
            },
            {
              pattern: '.',
              group: 'sibling',
              position: 'before',
            },
            {
              pattern: '..',
              group: 'parent',
              position: 'before',
            },
            {
              pattern: '../',
              group: 'parent',
              position: 'before',
            },
            {
              pattern: '[a-z]*',
              group: 'external',
              position: 'before',
            },
            {
              pattern: '../[a-z]*',
              group: 'parent',
              position: 'before',
            },
            {
              pattern: './[a-z]*',
              group: 'sibling',
              position: 'before',
            },
          ],
        },
      ],
    },
    {
      name: 'run 1 valid 65',
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
    {
      name: 'run 1 valid 66',
      code: '\n        import { a, B as C, Z } from \'./Z\';\n        const { D, n: c, Y } = require(\'./Z\');\n        export { C, D };\n        export { A, B, C as default } from "./Z";\n\n        const { ["ignore require-statements with non-identifier imports"]: z, d } = require("./Z");\n        exports = { ["ignore exports statements with non-identifiers"]: Z, D };\n      ',
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 67',
      code: "\n        const { b, A } = require('./Z');\n      ",
      options: [
        {
          named: true,
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 68',
      code: '\n        import { A, B } from "./Z";\n        export { Z, A } from "./Z";\n        export { N, P } from "./Z";\n        const { X, Y } = require("./Z");\n      ',
      options: [
        {
          named: {
            require: true,
            import: true,
            export: false,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 69',
      code: '\n        import { B, A } from "./Z";\n        const { D, C } = require("./Z");\n        export { B, A } from "./Z";\n      ',
      options: [
        {
          named: {
            require: false,
            import: false,
            export: false,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 70',
      code: '\n        import { B, A, R } from "foo";\n        const { D, O, G } = require("tunes");\n        export { B, A, Z } from "foo";\n      ',
      options: [
        {
          named: {
            enabled: false,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 71',
      code: '\n        import { A as A, A as B, A as C } from "./Z";\n        const { a, a: b, a: c } = require("./Z");\n      ',
      options: [
        {
          named: true,
        },
      ],
    },
    {
      name: 'run 1 valid 72',
      code: '\n        import { A, B, C } from "./Z";\n        exports = { A, B, C };\n        module.exports = { a: A, b: B, c: C };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 73',
      code: '\n        module.exports.A = { };\n        module.exports.A.B = { };\n        module.exports.B = { };\n        exports.C = { };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 74',
      code: '\n        var exports = null;\n        var module = null;\n        exports = { };\n        module = { };\n        module.exports = { };\n        module.exports.U = { };\n        module.exports.N = { };\n        module.exports.C = { };\n        exports.L = { };\n        exports.E = { };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 75',
      code: '\n        exports["B"] = { };\n        exports["C"] = { };\n        exports["A"] = { };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 76',
      code: "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 77',
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 78',
      code: "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 79',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { C } from 'dirA/Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
        },
      ],
    },
    {
      name: 'run 1 valid 80',
      code: "\n              import c from 'Bar';\n              import type { A } from 'foo';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    {
      name: 'run 1 valid 81',
      code: "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 82',
      code: "\n              import { Partner } from '@models/partner/partner';\n              import { PartnerId } from '@models/partner/partner-id';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 83',
      code: "\n              import { serialize, parse, mapFieldErrors } from '@vtaits/form-schema';\n              import type { GetFieldSchema } from '@vtaits/form-schema';\n              import { useMemo, useCallback } from 'react';\n              import type { ReactElement, ReactNode } from 'react';\n              import { Form } from 'react-final-form';\n              import type { FormProps as FinalFormProps } from 'react-final-form';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 84',
      code: "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 85',
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
            {
              pattern: 'react',
              group: 'external',
              position: 'before',
            },
          ],
          'newlines-between': 'always',
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 86',
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
    {
      name: 'run 1 valid 87',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    {
      name: 'run 1 valid 88',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: false,
        },
      ],
    },
    {
      name: 'run 1 valid 89',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 90',
      code: "\n              import c from 'Bar';\n              import type { AA } from 'abc';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    {
      name: 'run 1 valid 91',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 92',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 93',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 94',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'ignore',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 95',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 96',
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 97',
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
    {
      name: 'run 1 valid 98',
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
    {
      name: 'run 1 valid 99',
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
    {
      name: 'run 1 valid 100',
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
    {
      name: 'run 1 valid 101',
      code: "\n              import fs from 'fs';\n\n              import '@scoped/package';\n              import type { B } from 'fs';\n\n              import type { A1 } from '/bad/bad/bad/bad';\n              import './a/b/c';\n              import type { A2 } from '/bad/bad/bad/bad';\n              import type { A3 } from '/bad/bad/bad/bad';\n              import type { D1 } from '/bad/bad/not/good';\n              import type { D2 } from '/bad/bad/not/good';\n              import type { D3 } from '/bad/bad/not/good';\n\n              import type { C } from '@something/else';\n\n              import type { E } from './index.js';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['builtin', 'type', 'unknown', 'external'],
          sortTypesGroup: true,
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 102',
      code: "\n              // https://prettier.io/docs/en/options.html\n\n              module.exports = {\n                  ...require('@xxxx/.prettierrc.js'),\n              };\n            ",
      options: [
        {
          named: {
            enabled: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 103',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 104',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 105',
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n\n              import { createDebugLogger } from 'multiverse+rejoinder';\n\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n\n              import {\n                $artificiallyInvoked,\n                $canonical,\n                $exists,\n                $genesis\n              } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: {
            enabled: true,
            types: 'types-last',
          },
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
    {
      name: 'run 1 valid 106',
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n              import { createDebugLogger } from 'multiverse+rejoinder';\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n              import { $artificiallyInvoked } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: {
            enabled: true,
            types: 'types-last',
          },
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
    {
      name: 'run 1 valid 107',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 108',
      code: "\n              import assert from 'assert';\n              import { isNativeError } from 'util/types';\n\n              import { runNoRejectOnBadExit } from '@-xun/run';\n              import { TrialError } from 'named-app-errors';\n              import { resolve as resolverLibrary } from 'resolve.exports';\n\n              import { toAbsolutePath, type AbsolutePath } from 'rootverse+project-utils:src/fs.ts';\n\n              import type { PackageJson } from 'type-fest';\n              // Some comment about remembering to do something\n              import type { XPackageJson } from 'rootverse:src/assets/config/_package.json.ts';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: {
            enabled: true,
            types: 'types-last',
          },
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
    {
      name: 'run 1 valid 109',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 110',
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 111',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 112',
      code: "\n              import blist2 from 'blist';\n              import blist from 'BList';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 113',
      code: "\n              import blist from 'BList';\n              import blist2 from 'blist';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: false,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 114',
      code: "\n              import { apply, compose } from 'xcompose';\n            ",
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 115',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import './styles.css';\n            ",
      options: [
        {
          warnOnUnassignedImports: true,
        },
      ],
    },
    {
      name: 'run 1 valid 116',
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: {
            order: 'asc',
          },
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 117',
      code: '\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 118',
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n\n              var async = require('async');\n\n              var relParent1 = require('../foo');\n\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n\n              var relParent3 = require('../bar');\n\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 119',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 120',
      code: "\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['type', 'external', 'internal', 'index'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 121',
      code: "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 122',
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 123',
      code: "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 124',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { C } from 'dirA/Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
        },
      ],
    },
    {
      name: 'run 1 valid 125',
      code: "\n              import c from 'Bar';\n              import type { A } from 'foo';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    {
      name: 'run 1 valid 126',
      code: "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 127',
      code: "\n              import { Partner } from '@models/partner/partner';\n              import { PartnerId } from '@models/partner/partner-id';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 128',
      code: "\n              import { serialize, parse, mapFieldErrors } from '@vtaits/form-schema';\n              import type { GetFieldSchema } from '@vtaits/form-schema';\n              import { useMemo, useCallback } from 'react';\n              import type { ReactElement, ReactNode } from 'react';\n              import { Form } from 'react-final-form';\n              import type { FormProps as FinalFormProps } from 'react-final-form';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 129',
      code: "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 130',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    {
      name: 'run 1 valid 131',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: false,
        },
      ],
    },
    {
      name: 'run 1 valid 132',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['type'],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 133',
      code: "\n              import c from 'Bar';\n              import type { AA } from 'abc';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import b from 'dirA/bar';\n              import type { D } from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
        },
      ],
    },
    {
      name: 'run 1 valid 134',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 135',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 136',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n              import type { AA } from 'abc';\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 137',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n              import type { C } from 'dirA/Bar';\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'ignore',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 138',
      code: "\n              import c from 'Bar';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA } from 'abc';\n\n              import type { A } from 'foo';\n\n              import type { C } from 'dirA/Bar';\n\n              import type { D } from 'dirA/bar';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 139',
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 140',
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
    {
      name: 'run 1 valid 141',
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
    {
      name: 'run 1 valid 142',
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
    {
      name: 'run 1 valid 143',
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
    {
      name: 'run 1 valid 144',
      code: "\n              import fs from 'fs';\n\n              import '@scoped/package';\n              import type { B } from 'fs';\n\n              import type { A1 } from '/bad/bad/bad/bad';\n              import './a/b/c';\n              import type { A2 } from '/bad/bad/bad/bad';\n              import type { A3 } from '/bad/bad/bad/bad';\n              import type { D1 } from '/bad/bad/not/good';\n              import type { D2 } from '/bad/bad/not/good';\n              import type { D3 } from '/bad/bad/not/good';\n\n              import type { C } from '@something/else';\n\n              import type { E } from './index.js';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['builtin', 'type', 'unknown', 'external'],
          sortTypesGroup: true,
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 145',
      code: "\n              // https://prettier.io/docs/en/options.html\n\n              module.exports = {\n                  ...require('@xxxx/.prettierrc.js'),\n              };\n            ",
      options: [
        {
          named: {
            enabled: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 146',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 147',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 148',
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n\n              import { createDebugLogger } from 'multiverse+rejoinder';\n\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n\n              import {\n                $artificiallyInvoked,\n                $canonical,\n                $exists,\n                $genesis\n              } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: {
            enabled: true,
            types: 'types-last',
          },
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
    {
      name: 'run 1 valid 149',
      code: "\n              import makeVanillaYargs from 'yargs/yargs';\n              import { createDebugLogger } from 'multiverse+rejoinder';\n              import { globalDebuggerNamespace } from 'rootverse+bfe:src/constant.ts';\n              import { ErrorMessage, type KeyValueEntry } from 'rootverse+bfe:src/error.ts';\n              import { $artificiallyInvoked } from 'rootverse+bfe:src/symbols.ts';\n\n              import type {\n                Entries,\n                LiteralUnion,\n                OmitIndexSignature,\n                Promisable,\n                StringKeyOf\n              } from 'type-fest';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: {
            enabled: true,
            types: 'types-last',
          },
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
    {
      name: 'run 1 valid 150',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'never',
          'newlines-between-types': 'always-and-inside-groups',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 151',
      code: "\n              import assert from 'assert';\n              import { isNativeError } from 'util/types';\n\n              import { runNoRejectOnBadExit } from '@-xun/run';\n              import { TrialError } from 'named-app-errors';\n              import { resolve as resolverLibrary } from 'resolve.exports';\n\n              import { toAbsolutePath, type AbsolutePath } from 'rootverse+project-utils:src/fs.ts';\n\n              import type { PackageJson } from 'type-fest';\n              // Some comment about remembering to do something\n              import type { XPackageJson } from 'rootverse:src/assets/config/_package.json.ts';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            orderImportKind: 'asc',
            caseInsensitive: true,
          },
          named: {
            enabled: true,
            types: 'types-last',
          },
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
    {
      name: 'run 1 valid 152',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
    },
    {
      name: 'run 1 valid 153',
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 154',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
    },
    {
      name: 'run 1 valid 155',
      code: "\n              import blist2 from 'blist';\n              import blist from 'BList';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 156',
      code: "\n              import blist from 'BList';\n              import blist2 from 'blist';\n              import * as classnames from 'classnames';\n              import aTypes from 'prop-types';\n              import React, { PureComponent } from 'react';\n              import { compose, apply } from 'xcompose';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: false,
          },
        },
      ],
    },
    {
      name: 'run 1 valid 157',
      code: "\n              import { apply, compose } from 'xcompose';\n            ",
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
    },
    {
      name: 'run 1 valid 158',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import './styles.css';\n            ",
      options: [
        {
          warnOnUnassignedImports: true,
        },
      ],
    },
    {
      name: 'run 1 valid 159',
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: {
            order: 'asc',
          },
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 160',
      code: '\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n            ',
      options: [
        {
          groups: ['builtin', 'parent', 'sibling', 'index', 'type'],
          sortTypesGroup: true,
        },
      ],
    },
    {
      name: 'run 1 valid 161',
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n\n              var async = require('async');\n\n              var relParent1 = require('../foo');\n\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n\n              var relParent3 = require('../bar');\n\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 162',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
          ],
          'newlines-between': 'always-and-inside-groups',
          'newlines-between-types': 'never',
          pathGroupsExcludedImportTypes: [],
          sortTypesGroup: true,
          consolidateIslands: 'inside-groups',
        },
      ],
    },
    {
      name: 'run 1 valid 163',
      code: "\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['type', 'external', 'internal', 'index'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import a from './foo';\nimport b from '../bar';\n",
      options: [
        {
          groups: ['builtin', 'external', 'internal', 'type'],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`../bar` import should occur before import of `./foo`',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 24,
        },
      ],
      output: "import b from '../bar';\nimport a from './foo';\n",
    },
    {
      name: 'run 0 invalid 1',
      code: "\n        var async = require('async');\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 2',
      code: "\n        var async = require('async');\n        var fs = require('fs'); \n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output:
        "\n        var fs = require('fs'); \n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 3',
      code: "\n        var async = require('async');\n        var fs = require('fs'); /* comment */\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output:
        "\n        var fs = require('fs'); /* comment */\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 4',
      code: "\n        /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 47,
        },
      ],
      output:
        "\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n        /* comment1 */  var async = require('async'); /* comment2 */\n      ",
    },
    {
      name: 'run 0 invalid 5',
      code: "\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 47,
        },
      ],
      output:
        "\n        /* comment3 */  var fs = require('fs'); /* comment4 */\n        /* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\n      ",
    },
    {
      name: 'run 0 invalid 6',
      code: "/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n/* comment3 */  var fs = require('fs'); /* comment4 */\r\n",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 2,
          column: 26,
          endLine: 2,
          endColumn: 39,
        },
      ],
      output:
        "/* comment3 */  var fs = require('fs'); /* comment4 */\r\n/* comment0 */  /* comment1 */  var async = require('async'); /* comment2 */\r\n",
    },
    {
      name: 'run 0 invalid 7',
      code: "\n        /* multi-line1\n          comment1 */  var async = require('async'); /* multi-line2\n          comment2 */  var fs = require('fs'); /* multi-line3\n          comment3 */\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 33,
          endLine: 4,
          endColumn: 46,
        },
      ],
      output:
        "\n        /* multi-line1\n          comment1 */  var fs = require('fs'); \n  var async = require('async'); /* multi-line2\n          comment2 *//* multi-line3\n          comment3 */\n      ",
    },
    {
      name: 'run 0 invalid 8',
      code: "\n        var {b} = require('async');\n        var {a} = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 32,
        },
      ],
      output:
        "\n        var {a} = require('fs');\n        var {b} = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 9',
      code: "\n        var async = require('async');\n        var fs =\n          require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 24,
        },
      ],
      output:
        "\n        var fs =\n          require('fs');\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 10',
      code: "\n        var async = require('async');\n        var fs = require('fs');",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var async = require('async');\n",
    },
    {
      name: 'run 0 invalid 11',
      code: "\n        import async from 'async';\n        import fs from 'fs';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import async from 'async';\n      ",
    },
    {
      name: 'run 0 invalid 12',
      code: "\n        var async = require('async');\n        import fs from 'fs';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 13',
      code: "\n        var parent = require('../parent');\n        var async = require('async');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`async` import should occur before import of `../parent`',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 37,
        },
      ],
      output:
        "\n        var async = require('async');\n        var parent = require('../parent');\n      ",
    },
    {
      name: 'run 0 invalid 14',
      code: "\n        var sibling = require('./sibling');\n        var parent = require('../parent');\n      ",
      errors: [
        {
          messageId: 'order',
          message:
            '`../parent` import should occur before import of `./sibling`',
          line: 3,
          column: 22,
          endLine: 3,
          endColumn: 42,
        },
      ],
      output:
        "\n        var parent = require('../parent');\n        var sibling = require('./sibling');\n      ",
    },
    {
      name: 'run 0 invalid 15',
      code: "\n        var index = require('./');\n        var sibling = require('./sibling');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./sibling` import should occur before import of `./`',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 43,
        },
      ],
      output:
        "\n        var sibling = require('./sibling');\n        var index = require('./');\n      ",
    },
    {
      name: 'run 0 invalid 16',
      code: "\n          var sibling = require('./sibling');\n          var async = require('async');\n          var fs = require('fs');\n        ",
      errors: [
        {
          messageId: 'order',
          message: '`async` import should occur before import of `./sibling`',
          line: 3,
          column: 23,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `./sibling`',
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 33,
        },
      ],
      output:
        "\n          var fs = require('fs');\n          var async = require('async');\n          var sibling = require('./sibling');\n        ",
    },
    {
      name: 'run 0 invalid 17',
      code: "\n        var index = require('./');\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./` import should occur after import of `bar`',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 34,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var path = require('path');\n        var _ = require('lodash');\n        var foo = require('foo');\n        var bar = require('bar');\n        var index = require('./');\n      ",
    },
    {
      name: 'run 0 invalid 18',
      code: "\n        var fs = require('fs');\n        var index = require('./');\n      ",
      options: [
        {
          groups: ['index', 'sibling', 'parent', 'external', 'builtin'],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`./` import should occur before import of `fs`',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 34,
        },
      ],
      output:
        "\n        var index = require('./');\n        var fs = require('fs');\n      ",
    },
    {
      name: 'run 0 invalid 19',
      code: "\n        var foo = require('./foo').bar;\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `./foo`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: "\n        var foo = require('./foo').bar.bar.bar;\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `./foo`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: "\n        var foo = require('./foo').bar\n          .bar\n          .bar;\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `./foo`',
          line: 5,
          column: 18,
          endLine: 5,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: "\n        var foo = require('./foo');\n        var fs = require('fs').bar\n          .bar\n          .bar;\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `./foo`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
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
        {
          messageId: 'order',
          message: '`path` import should occur before import of `./foo`',
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 35,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n      ",
    },
    {
      name: 'run 0 invalid 24',
      code: "\n        var path = require('path');\n        var async = require('async');\n      ",
      options: [
        {
          groups: ['index', ['sibling', 'parent', 'external', 'internal']],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`async` import should occur before import of `path`',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 37,
        },
      ],
      output:
        "\n        var async = require('async');\n        var path = require('path');\n      ",
    },
    {
      name: 'run 0 invalid 25',
      code: "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        import sibling, {foo3} from './foo';\n        var index = require('./');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./foo` import should occur before import of `fs`',
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 45,
        },
      ],
      output:
        "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        import sibling, {foo3} from './foo';\n        var fs = require('fs');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var index = require('./');\n      ",
    },
    {
      name: 'run 0 invalid 26',
      code: "\n        var fs = require('fs');\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur after import of `../foo/bar`',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 31,
        },
      ],
      output:
        "\n        import async, {foo1} from 'async';\n        import relParent2, {foo2} from '../foo/bar';\n        var fs = require('fs');\n      ",
    },
    {
      name: 'run 0 invalid 27',
      code: "\n          var fs = require('fs');\n          import async, {foo1} from 'async';\n          import bar = require(\"../foo/bar\");\n        ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur after import of `../foo/bar`',
          line: 2,
          column: 20,
          endLine: 2,
          endColumn: 33,
        },
      ],
      output:
        "\n          import async, {foo1} from 'async';\n          import bar = require(\"../foo/bar\");\n          var fs = require('fs');\n        ",
    },
    {
      name: 'run 0 invalid 28',
      code: "\n          var async = require('async');\n          var fs = require('fs');\n        ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output:
        "\n          var fs = require('fs');\n          var async = require('async');\n        ",
    },
    {
      name: 'run 0 invalid 29',
      code: "\n          import sync = require('sync');\n          import async, {foo1} from 'async';\n\n          import index from './';\n        ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`async` import should occur before import of `sync`',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 45,
        },
      ],
      output:
        "\n          import async, {foo1} from 'async';\n          import sync = require('sync');\n\n          import index from './';\n        ",
    },
    {
      name: 'run 0 invalid 30',
      code: "\n          import log = console.log;\n          import blah = require('./blah');",
      errors: [
        {
          messageId: 'order',
          message:
            '`./blah` import should occur before import of `console.log`',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'run 0 invalid 31',
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
          messageId: 'order',
          message:
            '`fs` import should occur before import of `-/components/Button`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 29,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import { Button } from '-/components/Button';\n        import { add } from './helper';\n      ",
    },
    {
      name: 'run 0 invalid 32',
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
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 29,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 56,
        },
      ],
      output:
        "\n        import fs from 'fs';\n\n        import { Button } from '-/components/Button';\n        import { LinkButton } from '-/components/Link';\n\n        import { add } from './helper';\n      ",
    },
    {
      name: 'run 0 invalid 33',
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'never',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 6,
          column: 23,
          endLine: 6,
          endColumn: 39,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 34',
      code: "\n        var fs = require('fs'); /* comment */\n\n        var index = require('./');\n      ",
      options: [
        {
          groups: [['builtin'], ['index']],
          'newlines-between': 'never',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 31,
        },
      ],
      output:
        "\n        var fs = require('fs'); /* comment */\n        var index = require('./');\n      ",
    },
    {
      name: 'run 0 invalid 35',
      code: "\n        var fs = require('fs'); /* multi-line\n        comment */\n\n        var index = require('./');\n      ",
      options: [
        {
          groups: [['builtin'], ['index']],
          'newlines-between': 'never',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 36',
      code: "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n        var sibling = require('./foo');\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
      options: [
        {
          groups: [['builtin', 'index'], ['sibling'], ['parent', 'external']],
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 23,
          endLine: 5,
          endColumn: 39,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var index = require('./');\n        var path = require('path');\n\n        var sibling = require('./foo');\n\n        var relParent1 = require('../foo');\n        var relParent3 = require('../');\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 37',
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
        {
          messageId: 'withinGroupNewline',
          message: 'There should be no empty line within import group',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 31,
        },
        {
          messageId: 'withinGroupNewline',
          message: 'There should be no empty line within import group',
          line: 7,
          column: 23,
          endLine: 7,
          endColumn: 39,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var path = require('path');\n        var index = require('./');\n\n        var sibling = require('./foo');\n        var async = require('async');\n      ",
    },
    {
      name: 'run 0 invalid 38',
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'never',
          warnOnUnassignedImports: false,
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 39',
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'never',
          warnOnUnassignedImports: true,
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output:
        "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ",
    },
    {
      name: 'run 0 invalid 40',
      code: "\n        import path from 'path';\n        export const abc = 123;\n\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 41',
      code: "\n        import path from 'path';\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
      output:
        "\n        import path from 'path';\n\n        import 'loud-rejection';\n        import 'something-else';\n        import _ from 'lodash';\n      ",
    },
    {
      name: 'run 0 invalid 42',
      code: "\n        import path from 'path'; // comment\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
      output:
        "\n        import path from 'path'; // comment\n\n        import _ from 'lodash';\n      ",
    },
    {
      name: 'run 0 invalid 43',
      code: "\n        import path from 'path'; /* comment */ /* comment */\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
      output:
        "\n        import path from 'path'; /* comment */ /* comment */\n\n        import _ from 'lodash';\n      ",
    },
    {
      name: 'run 0 invalid 44',
      code: "\n        import path from 'path'; /* 1\n        2 */\n        import _ from 'lodash';\n      ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
      ],
      output:
        "\n        import path from 'path';\n\n /* 1\n        2 */\n        import _ from 'lodash';\n      ",
    },
    {
      name: 'run 0 invalid 45',
      code: "\n        const local = require('./local');\n\n        fn_call();\n\n        const global1 = require('global1');\n        const global2 = require('global2');\n\n        fn_call();\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./local` import should occur after import of `global2`',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'run 0 invalid 46',
      code: "\n        const local = require('./local');\n        fn_call();\n        const global1 = require('global1');\n        const global2 = require('global2');\n\n        fn_call();\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./local` import should occur after import of `global2`',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'run 0 invalid 47',
      code: "\n        const local1 = require('./local1');\n        const local2 = require('./local2');\n        const local3 = require('./local3');\n        const local4 = require('./local4');\n        fn_call();\n        const global1 = require('global1');\n        const global2 = require('global2');\n        const global3 = require('global3');\n        const global4 = require('global4');\n        const global5 = require('global5');\n        fn_call();\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./local1` import should occur after import of `global5`',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 43,
        },
        {
          messageId: 'order',
          message: '`./local2` import should occur after import of `global5`',
          line: 3,
          column: 24,
          endLine: 3,
          endColumn: 43,
        },
        {
          messageId: 'order',
          message: '`./local3` import should occur after import of `global5`',
          line: 4,
          column: 24,
          endLine: 4,
          endColumn: 43,
        },
        {
          messageId: 'order',
          message: '`./local4` import should occur after import of `global5`',
          line: 5,
          column: 24,
          endLine: 5,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'run 0 invalid 48',
      code: "\n        const local = require('./local');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const global3 = require('global3');\n\n        fn_call();\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./local` import should occur after import of `global3`',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 41,
        },
      ],
    },
    {
      name: 'run 0 invalid 49',
      code: "\n        const local1 = require('./local1');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const local2 = require('./local2');\n        const global3 = require('global3');\n        const global4 = require('global4');\n\n        fn_call();\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./local1` import should occur after import of `global4`',
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 43,
        },
        {
          messageId: 'order',
          message: '`./local2` import should occur after import of `global4`',
          line: 6,
          column: 24,
          endLine: 6,
          endColumn: 43,
        },
      ],
      output:
        "\n        const local1 = require('./local1');\n        const global1 = require('global1');\n        const global2 = require('global2');\n        fn_call();\n        const global3 = require('global3');\n        const global4 = require('global4');\n        const local2 = require('./local2');\n\n        fn_call();\n      ",
    },
    {
      name: 'run 0 invalid 50',
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        ",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`~/components/Input` import should occur before import of `./helper`',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 52,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import { add } from './helper';\n        ",
    },
    {
      name: 'run 0 invalid 51',
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        ",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`./helper` import should occur after import of `async`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 40,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { Input } from '~/components/Input';\n        import async from 'async';\n        import { add } from './helper';\n        ",
    },
    {
      name: 'run 0 invalid 52',
      code: "\n        import fs from 'fs';\n        import _ from 'lodash';\n        import { add } from './helper';\n        import { Input } from '~/components/Input';\n        ",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'before',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`~/components/Input` import should occur before import of `lodash`',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 52,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import { Input } from '~/components/Input';\n        import _ from 'lodash';\n        import { add } from './helper';\n        ",
    },
    {
      name: 'run 0 invalid 53',
      code: "\n        import fs from 'fs';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        import { Export } from '-/components/Export';\n        ",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '#/**',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '-/**',
              group: 'external',
              position: 'before',
            },
            {
              pattern: '$/**',
              group: 'external',
              position: 'before',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`-/components/Export` import should occur before import of `$/components/Import`',
          line: 8,
          column: 9,
          endLine: 8,
          endColumn: 54,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        ",
    },
    {
      name: 'run 0 invalid 54',
      code: "\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        import { Output } from '~/components/Output';\n        ",
      options: [
        {
          pathGroups: [
            {
              pattern: '~/**',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '#/**',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '-/**',
              group: 'external',
              position: 'before',
            },
            {
              pattern: '$/**',
              group: 'external',
              position: 'before',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`~/components/Output` import should occur before import of `#/components/Input`',
          line: 8,
          column: 9,
          endLine: 8,
          endColumn: 54,
        },
      ],
      output:
        "\n        import fs from 'fs';\n        import { Export } from '-/components/Export';\n        import { Import } from '$/components/Import';\n        import _ from 'lodash';\n        import { Output } from '~/components/Output';\n        import { Input } from '#/components/Input';\n        import { add } from './helper';\n        ",
    },
    {
      name: 'run 0 invalid 55',
      code: "\n        import path from 'path';\n        import { namespace } from '@namespace';\n        import { a } from 'a';\n        import { b } from 'b';\n        import { c } from 'c';\n        import { d } from 'd';\n        import { e } from 'e';\n        import { f } from 'f';\n        import { g } from 'g';\n        import { h } from 'h';\n        import { i } from 'i';\n        import { j } from 'j';\n        import { k } from 'k';",
      options: [
        {
          groups: ['builtin', 'external', 'internal'],
          pathGroups: [
            {
              pattern: '@namespace',
              group: 'external',
              position: 'after',
            },
            {
              pattern: 'a',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'b',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'c',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'd',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'e',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'f',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'g',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'h',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'i',
              group: 'internal',
              position: 'before',
            },
          ],
          'newlines-between': 'always',
          pathGroupsExcludedImportTypes: ['builtin'],
        },
      ],
      settings: {
        'import/internal-regex': '^(a|b|c|d|e|f|g|h|i|j|k)(\\/|$)',
      },
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 48,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 6,
          column: 9,
          endLine: 6,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 7,
          column: 9,
          endLine: 7,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 8,
          column: 9,
          endLine: 8,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 9,
          column: 9,
          endLine: 9,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 10,
          column: 9,
          endLine: 10,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 11,
          column: 9,
          endLine: 11,
          endColumn: 31,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 12,
          column: 9,
          endLine: 12,
          endColumn: 31,
        },
      ],
      output:
        "\n        import path from 'path';\n\n        import { namespace } from '@namespace';\n\n        import { a } from 'a';\n\n        import { b } from 'b';\n\n        import { c } from 'c';\n\n        import { d } from 'd';\n\n        import { e } from 'e';\n\n        import { f } from 'f';\n\n        import { g } from 'g';\n\n        import { h } from 'h';\n\n        import { i } from 'i';\n\n        import { j } from 'j';\n        import { k } from 'k';",
    },
    {
      name: 'run 0 invalid 56',
      code: "\n        import external from 'external';\n        import a from '@namespace/a';\n        import b from '@namespace/b';\n        import { parent } from '../../parent';\n        import local from './local';\n        import './side-effect';",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
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
            {
              pattern: '@namespace',
              group: 'external',
              position: 'after',
            },
            {
              pattern: '@namespace/**',
              group: 'external',
              position: 'after',
            },
          ],
          pathGroupsExcludedImportTypes: [],
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 41,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 38,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 47,
        },
      ],
      output:
        "\n        import external from 'external';\n\n        import a from '@namespace/a';\n        import b from '@namespace/b';\n\n        import { parent } from '../../parent';\n\n        import local from './local';\n        import './side-effect';",
    },
    {
      name: 'run 0 invalid 57',
      code: "\n        var async = require('async');\n        fn_call();\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 58',
      code: "\n        const env = require('./config');\n\n        Object.keys(env);\n\n        const http = require('http');\n        const express = require('express');\n\n        http.createServer(express());\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`./config` import should occur after import of `express`',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 59',
      code: "\n        var async = require('async');\n        var a = require('./value.js')(a);\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 60',
      code: "\n        var async = require('async');\n        var fs = require('fs')(a);\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 61',
      code: "\n        var async = require('async')(a);\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 62',
      code: "\n        var async = require('async');\n        require('./aa');\n        var fs = require('fs');\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 63',
      code: "\n        import async from 'async';\n        fn_call();\n        import fs from 'fs';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 64',
      code: "\n        import async from 'async';\n        var a = 1;\n        import fs from 'fs';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 65',
      code: "\n        import async from 'async';\n        var a = require('./value.js')(a);\n        import fs from 'fs';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 66',
      code: "\n        import async from 'async';\n        import './aa';\n        import fs from 'fs';\n      ",
      errors: [
        {
          messageId: 'order',
          message: '`fs` import should occur before import of `async`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 67',
      code: "\n        import b from 'bar';\n        import c from 'Bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`Bar` import should occur before import of `bar`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output:
        "\n        import c from 'Bar';\n        import b from 'bar';\n        import a from 'foo';\n\n        import index from './';\n      ",
    },
    {
      name: 'run 0 invalid 68',
      code: "\n        import a from 'foo';\n        import c from 'Bar';\n        import b from 'bar';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur before import of `Bar`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 29,
        },
      ],
      output:
        "\n        import a from 'foo';\n        import b from 'bar';\n        import c from 'Bar';\n\n        import index from './';\n      ",
    },
    {
      name: 'run 0 invalid 69',
      code: '\n        import a from "foo";\n        import b from "foo-bar";\n        import c from "foo/bar";\n        import d from "foo/barfoo";\n      ',
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`foo-bar` import should occur after import of `foo/barfoo`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output:
        '\n        import a from "foo";\n        import c from "foo/bar";\n        import d from "foo/barfoo";\n        import b from "foo-bar";\n      ',
    },
    {
      name: 'run 0 invalid 70',
      code: "\n        import b from 'foo';\n        import a from 'Bar';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`Bar` import should occur before import of `foo`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output:
        "\n        import a from 'Bar';\n        import b from 'foo';\n\n        import index from './';\n      ",
    },
    {
      name: 'run 0 invalid 71',
      code: "\n        import a from 'Bar';\n        import b from 'foo';\n\n        import index from './';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
            caseInsensitive: true,
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`foo` import should occur before import of `Bar`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 29,
        },
      ],
      output:
        "\n        import b from 'foo';\n        import a from 'Bar';\n\n        import index from './';\n      ",
    },
    {
      name: 'run 0 invalid 72',
      code: "\n        const b = require('./b').get();\n        const a = require('./a');\n      ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`./a` import should occur before import of `./b`',
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 33,
        },
      ],
      output:
        "\n        const a = require('./a');\n        const b = require('./b').get();\n      ",
    },
    {
      name: 'run 0 invalid 73',
      code: "\n        import a from '../a';\n        import p from '..';\n      ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`..` import should occur before import of `../a`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 28,
        },
      ],
      output:
        "\n        import p from '..';\n        import a from '../a';\n      ",
    },
    {
      name: 'run 0 invalid 74',
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
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 27,
        },
      ],
      output:
        "\n        import A from 'a';\n\n        import C from './c';\n      ",
    },
    {
      name: 'run 0 invalid 75',
      code: "\n        import A from 'a';\n\n        import C from 'c';\n      ",
      options: [
        {
          'newlines-between': 'always',
          distinctGroup: false,
          pathGroupsExcludedImportTypes: [],
          pathGroups: [
            {
              pattern: 'a',
              group: 'external',
              position: 'before',
            },
            {
              pattern: 'c',
              group: 'external',
              position: 'after',
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'withinGroupNewline',
          message: 'There should be no empty line within import group',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 27,
        },
      ],
      output:
        "\n        import A from 'a';\n        import C from 'c';\n      ",
    },
    {
      name: 'run 0 invalid 76',
      code: '\n        var { B, A: R } = require("./Z");\n        import { O as G, D } from "./Z";\n        import { K, L, J } from "./Z";\n        export { Z, X, Y } from "./Z";\n      ',
      filename: 'src/virtual.js',
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` import should occur before import of `B`',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: 'order',
          message: '`D` import should occur before import of `O`',
          line: 3,
          column: 26,
          endLine: 3,
          endColumn: 27,
        },
        {
          messageId: 'order',
          message: '`J` import should occur before import of `K`',
          line: 4,
          column: 24,
          endLine: 4,
          endColumn: 25,
        },
        {
          messageId: 'order',
          message: '`Z` export should occur after export of `Y`',
          line: 5,
          column: 18,
          endLine: 5,
          endColumn: 19,
        },
      ],
      output:
        '\n        var { A: R, B } = require("./Z");\n        import { D, O as G } from "./Z";\n        import { J, K, L } from "./Z";\n        export { X, Y, Z } from "./Z";\n      ',
    },
    {
      name: 'run 0 invalid 77',
      code: '\n        import { D, C } from "./Z";\n        var { B, A } = require("./Z");\n        export { B, A };\n      ',
      options: [
        {
          named: {
            require: false,
            import: true,
            export: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`C` import should occur before import of `D`',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 22,
        },
        {
          messageId: 'order',
          message: '`A` export should occur before export of `B`',
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 22,
        },
      ],
      output:
        '\n        import { C, D } from "./Z";\n        var { B, A } = require("./Z");\n        export { A, B };\n      ',
    },
    {
      name: 'run 0 invalid 78',
      code: '\n        import { A as B, A as C, A } from "./Z";\n        export { A, A as D, A as B, A as C } from "./Z";\n        const { a: b, a: c, a } = require("./Z");\n      ',
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` import should occur before import of `A as B`',
          line: 2,
          column: 34,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'order',
          message: '`A as D` export should occur after export of `A as C`',
          line: 3,
          column: 21,
          endLine: 3,
          endColumn: 27,
        },
        {
          messageId: 'order',
          message: '`a` import should occur before import of `a as b`',
          line: 4,
          column: 29,
          endLine: 4,
          endColumn: 30,
        },
      ],
      output:
        '\n        import { A, A as B, A as C } from "./Z";\n        export { A, A as B, A as C, A as D } from "./Z";\n        const { a, a: b, a: c } = require("./Z");\n      ',
    },
    {
      name: 'run 0 invalid 79',
      code: '\n        import { A, B, C } from "./Z";\n        exports = { B, C, A };\n        module.exports = { c: C, a: A, b: B };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` export should occur before export of `B`',
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 28,
        },
        {
          messageId: 'order',
          message: '`c` export should occur after export of `b`',
          line: 4,
          column: 28,
          endLine: 4,
          endColumn: 32,
        },
      ],
      output:
        '\n        import { A, B, C } from "./Z";\n        exports = { A, B, C };\n        module.exports = { a: A, b: B, c: C };\n      ',
    },
    {
      name: 'run 0 invalid 80',
      code: '\n        exports.B = { };\n        module.exports.A = { };\n        module.exports.C = { };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` export should occur before export of `B`',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 31,
        },
      ],
      output:
        '\n        module.exports.A = { };\n        exports.B = { };\n        module.exports.C = { };\n      ',
    },
    {
      name: 'run 0 invalid 81',
      code: '\n        exports.A.C = { };\n        module.exports.A.A = { };\n        exports.A.B = { };\n      ',
      options: [
        {
          named: {
            cjsExports: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A.C` export should occur after export of `A.B`',
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 26,
        },
      ],
      output:
        '\n        module.exports.A.A = { };\n        exports.A.B = { };\n        exports.A.C = { };\n      ',
    },
    {
      name: 'run 0 invalid 82',
      code: '\n        const {\n          F: O,\n          O: B,\n          /* Hello World */\n          A: R\n        } = require("./Z");\n        import {\n          Y,\n          X,\n        } from "./Z";\n        export {\n          Z, A,\n          B\n        } from "./Z";\n        module.exports = {\n          a: A, o: O,\n          b: B\n        };\n      ',
      filename: 'src/virtual.js',
      options: [
        {
          named: {
            enabled: true,
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` import should occur before import of `F`',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 15,
        },
        {
          messageId: 'order',
          message: '`X` import should occur before import of `Y`',
          line: 10,
          column: 11,
          endLine: 10,
          endColumn: 12,
        },
        {
          messageId: 'order',
          message: '`Z` export should occur after export of `B`',
          line: 13,
          column: 11,
          endLine: 13,
          endColumn: 12,
        },
        {
          messageId: 'order',
          message: '`b` export should occur before export of `o`',
          line: 18,
          column: 11,
          endLine: 18,
          endColumn: 15,
        },
      ],
      output:
        '\n        const {\n          /* Hello World */\n          A: R,\n          F: O,\n          O: B\n        } = require("./Z");\n        import {\n          X,\n          Y,\n        } from "./Z";\n        export { A,\n          B,\n          Z\n        } from "./Z";\n        module.exports = {\n          a: A,\n          b: B, o: O\n        };\n      ',
    },
    {
      name: 'run 0 invalid 83',
      code: "\n          const { cello } = require('./cello');\n          import { int } from './int';\n          const blah = require('./blah');\n          import { hello } from './hello';\n        ",
      errors: [
        {
          messageId: 'order',
          message: '`./int` import should occur before import of `./cello`',
          line: 3,
          column: 11,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: 'order',
          message: '`./hello` import should occur before import of `./cello`',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 43,
        },
      ],
      output:
        "\n          import { int } from './int';\n          import { hello } from './hello';\n          const { cello } = require('./cello');\n          const blah = require('./blah');\n        ",
    },
    {
      name: 'run 0 invalid 84',
      code: "\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n        var async = require('async');\n        var relParent1 = require('../foo');\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n        var relParent3 = require('../bar');\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 39,
          endLine: 4,
          endColumn: 54,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 21,
          endLine: 5,
          endColumn: 37,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 6,
          column: 26,
          endLine: 6,
          endColumn: 43,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 27,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 13,
          column: 26,
          endLine: 13,
          endColumn: 43,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 15,
          column: 34,
          endLine: 15,
          endColumn: 50,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ",
    },
    {
      name: 'run 0 invalid 85',
      code: "\n        var fs = require('fs');\n\n        var path = require('path');\n\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n\n        var sibling3 = require('./foobar');\n      ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 31,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 24,
          column: 24,
          endLine: 24,
          endColumn: 40,
        },
      ],
      output:
        "\n        var fs = require('fs');\n        var path = require('path');\n        var { util1, util2, util3 } = require('util');\n\n        var async = require('async');\n\n        var relParent1 = require('../foo');\n\n        var {\n          relParent21,\n          relParent22,\n          relParent23,\n          relParent24,\n        } = require('../');\n\n        var relParent3 = require('../bar');\n\n        var { sibling1,\n          sibling2, sibling3 } = require('./foo');\n\n        var sibling2 = require('./bar');\n        var sibling3 = require('./foobar');\n      ",
    },
    {
      name: 'run 0 invalid 86',
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur after type import of `Bar`',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 35,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 87',
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur before import of `Bar`',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 35,
        },
      ],
      output:
        "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 88',
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`Bar` import should occur before import of `bar`',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'order',
          message: '`Bar` type import should occur before type import of `foo`',
          line: 9,
          column: 15,
          endLine: 9,
          endColumn: 44,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
    },
    {
      name: 'run 0 invalid 89',
      code: "\n              import a from 'foo';\n              import c from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur before import of `Bar`',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'order',
          message: '`foo` type import should occur before type import of `Bar`',
          line: 9,
          column: 15,
          endLine: 9,
          endColumn: 44,
        },
      ],
      output:
        "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
    },
    {
      name: 'run 0 invalid 90',
      code: "\n              import './local1';\n              import global from 'global1';\n              import local from './local2';\n              import 'global2';\n            ",
      options: [
        {
          warnOnUnassignedImports: true,
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`global1` import should occur before import of `./local1`',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 44,
        },
        {
          messageId: 'order',
          message: '`global2` import should occur before import of `./local1`',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 91',
      code: "\n              import local from './local';\n\n              import 'global1';\n\n              import global2 from 'global2';\n              import global3 from 'global3';\n            ",
      options: [
        {
          warnOnUnassignedImports: true,
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`./local` import should occur after import of `global3`',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 43,
        },
      ],
    },
    {
      name: 'run 0 invalid 92',
      code: "\n              import type { ParsedPath } from 'path';\n              import type { CopyOptions } from 'fs';\n\n              declare module 'my-module' {\n                import type { ParsedPath } from 'path';\n                import type { CopyOptions } from 'fs';\n              }\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`fs` type import should occur before type import of `path`',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 53,
        },
        {
          messageId: 'order',
          message: '`fs` type import should occur before type import of `path`',
          line: 7,
          column: 17,
          endLine: 7,
          endColumn: 55,
        },
      ],
      output:
        "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ",
    },
    {
      name: 'run 0 invalid 93',
      code: '\n              import { type Z, A } from "./Z";\n              import type N, { E, D } from "./Z";\n              import type { L, G } from "./Z";\n            ',
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` import should occur before type import of `Z`',
          line: 2,
          column: 32,
          endLine: 2,
          endColumn: 33,
        },
        {
          messageId: 'order',
          message: '`D` import should occur before import of `E`',
          line: 3,
          column: 35,
          endLine: 3,
          endColumn: 36,
        },
        {
          messageId: 'order',
          message: '`G` import should occur before import of `L`',
          line: 4,
          column: 32,
          endLine: 4,
          endColumn: 33,
        },
      ],
      output:
        '\n              import { A, type Z } from "./Z";\n              import type N, { D, E } from "./Z";\n              import type { G, L } from "./Z";\n            ',
    },
    {
      name: 'run 0 invalid 94',
      code: '\n              const { B, /* Hello World */ A } = require("./Z");\n              export { B, A } from "./Z";\n            ',
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` import should occur before import of `B`',
          line: 2,
          column: 44,
          endLine: 2,
          endColumn: 45,
        },
        {
          messageId: 'order',
          message: '`A` export should occur before export of `B`',
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 28,
        },
      ],
      output:
        '\n              const { /* Hello World */ A, B } = require("./Z");\n              export { A, B } from "./Z";\n            ',
    },
    {
      name: 'run 0 invalid 95',
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
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 41,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 46,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 45,
        },
      ],
      output:
        '\n              import type A from "fs";\n              import type B from "path";\n\n              import type C from "../foo.js";\n\n              import type D from "./bar.js";\n\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
    },
    {
      name: 'run 0 invalid 96',
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
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 39,
        },
      ],
      output:
        '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
    },
    {
      name: 'run 0 invalid 97',
      code: "\n              import c from 'Bar';\n\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 60,
          column: 15,
          endLine: 60,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 76,
          column: 15,
          endLine: 76,
          endColumn: 49,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 98',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 4,
          column: 15,
          endLine: 12,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 13,
          column: 15,
          endLine: 21,
          endColumn: 29,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 22,
          column: 15,
          endLine: 22,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 40,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 24,
          column: 15,
          endLine: 24,
          endColumn: 38,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 25,
          column: 15,
          endLine: 26,
          endColumn: 37,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 27,
          column: 15,
          endLine: 27,
          endColumn: 45,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 28,
          column: 15,
          endLine: 31,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 32,
          column: 15,
          endLine: 32,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 33,
          column: 15,
          endLine: 37,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 38,
          column: 15,
          endLine: 38,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 39,
          column: 15,
          endLine: 39,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 40,
          column: 15,
          endLine: 44,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 46,
          column: 15,
          endLine: 46,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 47,
          column: 15,
          endLine: 47,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 48,
          column: 15,
          endLine: 52,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 53,
          column: 15,
          endLine: 53,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 54,
          column: 15,
          endLine: 54,
          endColumn: 51,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 99',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 22,
          column: 15,
          endLine: 22,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 40,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 24,
          column: 15,
          endLine: 24,
          endColumn: 38,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 28,
          column: 15,
          endLine: 31,
          endColumn: 28,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 38,
          column: 15,
          endLine: 38,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 46,
          column: 15,
          endLine: 46,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 53,
          column: 15,
          endLine: 53,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 54,
          column: 15,
          endLine: 54,
          endColumn: 51,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 100',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 43,
        },
      ],
      output:
        "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 101',
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 39,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 43,
        },
      ],
      output:
        "\n              import fs from 'fs';\n\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 102',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 43,
        },
      ],
      output:
        "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 103',
      code: "\n                import React, { PureComponent } from 'react';\n                import aTypes from 'prop-types';\n                import { compose, apply } from 'xcompose';\n                import * as classnames from 'classnames';\n                import blist2 from 'blist';\n                import blist from 'BList';\n              ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`prop-types` import should occur before import of `react`',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 49,
        },
        {
          messageId: 'order',
          message: '`classnames` import should occur before import of `react`',
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 58,
        },
        {
          messageId: 'order',
          message: '`blist` import should occur before import of `react`',
          line: 6,
          column: 17,
          endLine: 6,
          endColumn: 44,
        },
        {
          messageId: 'order',
          message: '`BList` import should occur before import of `react`',
          line: 7,
          column: 17,
          endLine: 7,
          endColumn: 43,
        },
      ],
      output:
        "\n                import blist2 from 'blist';\n                import blist from 'BList';\n                import * as classnames from 'classnames';\n                import aTypes from 'prop-types';\n                import React, { PureComponent } from 'react';\n                import { compose, apply } from 'xcompose';\n              ",
    },
    {
      name: 'run 0 invalid 104',
      code: "\n              import { compose, apply } from 'xcompose';\n            ",
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`apply` import should occur before import of `compose`',
          line: 2,
          column: 33,
          endLine: 2,
          endColumn: 38,
        },
      ],
      output:
        "\n              import { apply, compose } from 'xcompose';\n            ",
    },
    {
      name: 'run 0 invalid 105',
      code: "\n              import fs from 'fs';\n              import './styles.css';\n              import path from 'path';\n            ",
      options: [
        {
          warnOnUnassignedImports: true,
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`path` import should occur before import of `./styles.css`',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 106',
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`../foo.js` type import should occur before type import of `fs`',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 46,
        },
        {
          messageId: 'order',
          message:
            '`./bar.js` type import should occur before type import of `fs`',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 45,
        },
        {
          messageId: 'order',
          message: '`./` type import should occur before type import of `fs`',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 39,
        },
      ],
      output:
        '\n              import type C from "../foo.js";\n              import type E from \'./\';\n              import type D from "./bar.js";\n              import type A from "fs";\n              import type B from "path";\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
    },
    {
      name: 'run 0 invalid 107',
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n              var async = require('async');\n              var relParent1 = require('../foo');\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n              var relParent3 = require('../bar');\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 45,
          endLine: 4,
          endColumn: 60,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 27,
          endLine: 5,
          endColumn: 43,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 6,
          column: 32,
          endLine: 6,
          endColumn: 49,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 12,
          column: 19,
          endLine: 12,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 13,
          column: 32,
          endLine: 13,
          endColumn: 49,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 15,
          column: 40,
          endLine: 15,
          endColumn: 56,
        },
      ],
      output:
        "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n\n              var async = require('async');\n\n              var relParent1 = require('../foo');\n\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n\n              var relParent3 = require('../bar');\n\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
    },
    {
      name: 'run 0 invalid 108',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 4,
          column: 15,
          endLine: 12,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 13,
          column: 15,
          endLine: 21,
          endColumn: 29,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 22,
          column: 15,
          endLine: 22,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 40,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 24,
          column: 15,
          endLine: 24,
          endColumn: 38,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 25,
          column: 15,
          endLine: 26,
          endColumn: 37,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 27,
          column: 15,
          endLine: 27,
          endColumn: 45,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 28,
          column: 15,
          endLine: 31,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 32,
          column: 15,
          endLine: 32,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 33,
          column: 15,
          endLine: 37,
          endColumn: 33,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 39,
          column: 15,
          endLine: 39,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 40,
          column: 15,
          endLine: 44,
          endColumn: 33,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 47,
          column: 15,
          endLine: 47,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 48,
          column: 15,
          endLine: 52,
          endColumn: 33,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 109',
      code: '\n                export { type B, A };\n              ',
      options: [
        {
          named: {
            enabled: true,
            types: 'mixed',
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` export should occur before type export of `B`',
          line: 2,
          column: 34,
          endLine: 2,
          endColumn: 35,
        },
      ],
      output: '\n                export { A, type B };\n              ',
    },
    {
      name: 'run 0 invalid 110',
      code: '\n                import { type B, A, default as C } from "./Z";\n              ',
      options: [
        {
          named: {
            import: true,
            types: 'types-last',
          },
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`B` type import should occur after import of `default`',
          line: 2,
          column: 26,
          endLine: 2,
          endColumn: 32,
        },
      ],
      output:
        '\n                import { A, default as C, type B } from "./Z";\n              ',
    },
    {
      name: 'run 0 invalid 111',
      code: '\n                export { A, type Z } from "./Z";\n              ',
      options: [
        {
          named: {
            enabled: true,
            types: 'types-first',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`Z` type export should occur before export of `A`',
          line: 2,
          column: 29,
          endLine: 2,
          endColumn: 35,
        },
      ],
      output:
        '\n                export { type Z, A } from "./Z";\n              ',
    },
    {
      name: 'run 0 invalid 112',
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
          messageId: 'order',
          message:
            '`node:child_process` import should occur before import of `express`',
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 56,
        },
      ],
      output:
        "\n                import chpro from 'node:child_process';\n                import express from 'express';\n                import log4js from 'log4js';\n                // import fsp from 'node:fs/promises';\n              ",
    },
    {
      name: 'run 0 invalid 113',
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur after type import of `Bar`',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 35,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n              import type { A } from 'foo';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 114',
      code: "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n            ",
      options: [
        {
          groups: ['external', 'index'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur before import of `Bar`',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 35,
        },
      ],
      output:
        "\n              import a from 'foo';\n              import type { A } from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n              import type { C } from 'Bar';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 115',
      code: "\n              import b from 'bar';\n              import c from 'Bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`Bar` import should occur before import of `bar`',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'order',
          message: '`Bar` type import should occur before type import of `foo`',
          line: 9,
          column: 15,
          endLine: 9,
          endColumn: 44,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import b from 'bar';\n              import a from 'foo';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
    },
    {
      name: 'run 0 invalid 116',
      code: "\n              import a from 'foo';\n              import c from 'Bar';\n              import b from 'bar';\n\n              import index from './';\n\n              import type { C } from 'Bar';\n              import type { A } from 'foo';\n            ",
      options: [
        {
          groups: ['external', 'index', 'type'],
          alphabetize: {
            order: 'desc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`bar` import should occur before import of `Bar`',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 35,
        },
        {
          messageId: 'order',
          message: '`foo` type import should occur before type import of `Bar`',
          line: 9,
          column: 15,
          endLine: 9,
          endColumn: 44,
        },
      ],
      output:
        "\n              import a from 'foo';\n              import b from 'bar';\n              import c from 'Bar';\n\n              import index from './';\n\n              import type { A } from 'foo';\n              import type { C } from 'Bar';\n            ",
    },
    {
      name: 'run 0 invalid 117',
      code: "\n              import type { ParsedPath } from 'path';\n              import type { CopyOptions } from 'fs';\n\n              declare module 'my-module' {\n                import type { ParsedPath } from 'path';\n                import type { CopyOptions } from 'fs';\n              }\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`fs` type import should occur before type import of `path`',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 53,
        },
        {
          messageId: 'order',
          message: '`fs` type import should occur before type import of `path`',
          line: 7,
          column: 17,
          endLine: 7,
          endColumn: 55,
        },
      ],
      output:
        "\n              import type { CopyOptions } from 'fs';\n              import type { ParsedPath } from 'path';\n\n              declare module 'my-module' {\n                import type { CopyOptions } from 'fs';\n                import type { ParsedPath } from 'path';\n              }\n            ",
    },
    {
      name: 'run 0 invalid 118',
      code: '\n              const { B, /* Hello World */ A } = require("./Z");\n              export { B, A } from "./Z";\n            ',
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`A` import should occur before import of `B`',
          line: 2,
          column: 44,
          endLine: 2,
          endColumn: 45,
        },
        {
          messageId: 'order',
          message: '`A` export should occur before export of `B`',
          line: 3,
          column: 27,
          endLine: 3,
          endColumn: 28,
        },
      ],
      output:
        '\n              const { /* Hello World */ A, B } = require("./Z");\n              export { A, B } from "./Z";\n            ',
    },
    {
      name: 'run 0 invalid 119',
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
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 41,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 46,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 45,
        },
      ],
      output:
        '\n              import type A from "fs";\n              import type B from "path";\n\n              import type C from "../foo.js";\n\n              import type D from "./bar.js";\n\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
    },
    {
      name: 'run 0 invalid 120',
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
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 39,
        },
      ],
      output:
        '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n              import a from "fs";\n              import b from "path";\n\n              import c from "../foo.js";\n\n              import d from "./bar.js";\n\n              import e from "./";\n            ',
    },
    {
      name: 'run 0 invalid 121',
      code: "\n              import c from 'Bar';\n\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 60,
          column: 15,
          endLine: 60,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be no empty lines between this single-line import and the single-line import that follows it',
          line: 76,
          column: 15,
          endLine: 76,
          endColumn: 49,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 122',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 4,
          column: 15,
          endLine: 12,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 13,
          column: 15,
          endLine: 21,
          endColumn: 29,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 22,
          column: 15,
          endLine: 22,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 40,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 24,
          column: 15,
          endLine: 24,
          endColumn: 38,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 25,
          column: 15,
          endLine: 26,
          endColumn: 37,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 27,
          column: 15,
          endLine: 27,
          endColumn: 45,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 28,
          column: 15,
          endLine: 31,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 32,
          column: 15,
          endLine: 32,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 33,
          column: 15,
          endLine: 37,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 38,
          column: 15,
          endLine: 38,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 39,
          column: 15,
          endLine: 39,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 40,
          column: 15,
          endLine: 44,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 46,
          column: 15,
          endLine: 46,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 47,
          column: 15,
          endLine: 47,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 48,
          column: 15,
          endLine: 52,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 53,
          column: 15,
          endLine: 53,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 54,
          column: 15,
          endLine: 54,
          endColumn: 51,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 123',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 22,
          column: 15,
          endLine: 22,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 40,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 24,
          column: 15,
          endLine: 24,
          endColumn: 38,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 28,
          column: 15,
          endLine: 31,
          endColumn: 28,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 38,
          column: 15,
          endLine: 38,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 46,
          column: 15,
          endLine: 46,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 53,
          column: 15,
          endLine: 53,
          endColumn: 50,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 54,
          column: 15,
          endLine: 54,
          endColumn: 51,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n\n              import type { F } from './index.js';\n\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 124',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 43,
        },
      ],
      output:
        "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 125',
      code: "\n              import fs from 'fs';\n\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 39,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 43,
        },
      ],
      output:
        "\n              import fs from 'fs';\n\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 126',
      code: "\n              import fs from 'fs';\n              import path from 'path';\n\n              import sibling from './foo';\n\n              import index from './';\n            ",
      options: [
        {
          'newlines-between': 'never',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 39,
        },
        {
          messageId: 'groupNewline',
          message: 'There should be no empty line between import groups',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 43,
        },
      ],
      output:
        "\n              import fs from 'fs';\n              import path from 'path';\n              import sibling from './foo';\n              import index from './';\n            ",
    },
    {
      name: 'run 0 invalid 127',
      code: "\n                import React, { PureComponent } from 'react';\n                import aTypes from 'prop-types';\n                import { compose, apply } from 'xcompose';\n                import * as classnames from 'classnames';\n                import blist2 from 'blist';\n                import blist from 'BList';\n              ",
      options: [
        {
          alphabetize: {
            order: 'asc',
            caseInsensitive: true,
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`prop-types` import should occur before import of `react`',
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 49,
        },
        {
          messageId: 'order',
          message: '`classnames` import should occur before import of `react`',
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 58,
        },
        {
          messageId: 'order',
          message: '`blist` import should occur before import of `react`',
          line: 6,
          column: 17,
          endLine: 6,
          endColumn: 44,
        },
        {
          messageId: 'order',
          message: '`BList` import should occur before import of `react`',
          line: 7,
          column: 17,
          endLine: 7,
          endColumn: 43,
        },
      ],
      output:
        "\n                import blist2 from 'blist';\n                import blist from 'BList';\n                import * as classnames from 'classnames';\n                import aTypes from 'prop-types';\n                import React, { PureComponent } from 'react';\n                import { compose, apply } from 'xcompose';\n              ",
    },
    {
      name: 'run 0 invalid 128',
      code: "\n              import { compose, apply } from 'xcompose';\n            ",
      options: [
        {
          named: true,
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`apply` import should occur before import of `compose`',
          line: 2,
          column: 33,
          endLine: 2,
          endColumn: 38,
        },
      ],
      output:
        "\n              import { apply, compose } from 'xcompose';\n            ",
    },
    {
      name: 'run 0 invalid 129',
      code: "\n              import fs from 'fs';\n              import './styles.css';\n              import path from 'path';\n            ",
      options: [
        {
          warnOnUnassignedImports: true,
        },
      ],
      errors: [
        {
          messageId: 'order',
          message: '`path` import should occur before import of `./styles.css`',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 130',
      code: '\n              import type A from "fs";\n              import type B from "path";\n              import type C from "../foo.js";\n              import type D from "./bar.js";\n              import type E from \'./\';\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
      options: [
        {
          groups: ['type', 'builtin', 'parent', 'sibling', 'index'],
          alphabetize: {
            order: 'asc',
          },
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`../foo.js` type import should occur before type import of `fs`',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 46,
        },
        {
          messageId: 'order',
          message:
            '`./bar.js` type import should occur before type import of `fs`',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 45,
        },
        {
          messageId: 'order',
          message: '`./` type import should occur before type import of `fs`',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 39,
        },
      ],
      output:
        '\n              import type C from "../foo.js";\n              import type E from \'./\';\n              import type D from "./bar.js";\n              import type A from "fs";\n              import type B from "path";\n\n              import a from "fs";\n              import b from "path";\n              import c from "../foo.js";\n              import d from "./bar.js";\n              import e from "./";\n            ',
    },
    {
      name: 'run 0 invalid 131',
      code: "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n              var async = require('async');\n              var relParent1 = require('../foo');\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n              var relParent3 = require('../bar');\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
      options: [
        {
          'newlines-between': 'always-and-inside-groups',
          consolidateIslands: 'inside-groups',
        },
      ],
      errors: [
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 4,
          column: 45,
          endLine: 4,
          endColumn: 60,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 5,
          column: 27,
          endLine: 5,
          endColumn: 43,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 6,
          column: 32,
          endLine: 6,
          endColumn: 49,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 12,
          column: 19,
          endLine: 12,
          endColumn: 33,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 13,
          column: 32,
          endLine: 13,
          endColumn: 49,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 15,
          column: 40,
          endLine: 15,
          endColumn: 56,
        },
      ],
      output:
        "\n              var fs = require('fs');\n              var path = require('path');\n              var { util1, util2, util3 } = require('util');\n\n              var async = require('async');\n\n              var relParent1 = require('../foo');\n\n              var {\n                relParent21,\n                relParent22,\n                relParent23,\n                relParent24,\n              } = require('../');\n\n              var relParent3 = require('../bar');\n\n              var { sibling1,\n                sibling2, sibling3 } = require('./foo');\n\n              var sibling2 = require('./bar');\n              var sibling3 = require('./foobar');\n            ",
    },
    {
      name: 'run 0 invalid 132',
      code: "\n              import c from 'Bar';\n              import d from 'bar';\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n              import a from 'foo';\n              import b from 'dirA/bar';\n              import index from './';\n              import type { AA,\n                BB, CC } from 'abc';\n              import type { Z } from 'fizz';\n              import type {\n                A,\n                B\n              } from 'foo';\n              import type { C2 } from 'dirB/Bar';\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
      options: [
        {
          alphabetize: {
            order: 'asc',
          },
          groups: ['external', 'internal', 'index', 'type'],
          pathGroups: [
            {
              pattern: 'dirA/**',
              group: 'internal',
              position: 'after',
            },
            {
              pattern: 'dirB/**',
              group: 'internal',
              position: 'before',
            },
            {
              pattern: 'dirC/**',
              group: 'internal',
            },
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
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 4,
          column: 15,
          endLine: 12,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 13,
          column: 15,
          endLine: 21,
          endColumn: 29,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 22,
          column: 15,
          endLine: 22,
          endColumn: 35,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 23,
          column: 15,
          endLine: 23,
          endColumn: 40,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 24,
          column: 15,
          endLine: 24,
          endColumn: 38,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 25,
          column: 15,
          endLine: 26,
          endColumn: 37,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 27,
          column: 15,
          endLine: 27,
          endColumn: 45,
        },
        {
          messageId: 'groupNewline',
          message:
            'There should be at least one empty line between import groups',
          line: 28,
          column: 15,
          endLine: 31,
          endColumn: 28,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 32,
          column: 15,
          endLine: 32,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 33,
          column: 15,
          endLine: 37,
          endColumn: 33,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 39,
          column: 15,
          endLine: 39,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 40,
          column: 15,
          endLine: 44,
          endColumn: 33,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this import and the multi-line import that follows it',
          line: 47,
          column: 15,
          endLine: 47,
          endColumn: 50,
        },
        {
          messageId: 'consolidate',
          message:
            'There should be at least one empty line between this multi-line import and the import that follows it',
          line: 48,
          column: 15,
          endLine: 52,
          endColumn: 33,
        },
      ],
      output:
        "\n              import c from 'Bar';\n              import d from 'bar';\n\n              import {\n                aa,\n                bb,\n                cc,\n                dd,\n                ee,\n                ff,\n                gg\n              } from 'baz';\n\n              import {\n                hh,\n                ii,\n                jj,\n                kk,\n                ll,\n                mm,\n                nn\n              } from 'fizz';\n\n              import a from 'foo';\n\n              import b from 'dirA/bar';\n\n              import index from './';\n\n              import type { AA,\n                BB, CC } from 'abc';\n\n              import type { Z } from 'fizz';\n\n              import type {\n                A,\n                B\n              } from 'foo';\n\n              import type { C2 } from 'dirB/Bar';\n\n              import type {\n                D2,\n                X2,\n                Y2\n              } from 'dirB/bar';\n\n              import type { E2 } from 'dirB/baz';\n              import type { C3 } from 'dirC/Bar';\n\n              import type {\n                D3,\n                X3,\n                Y3\n              } from 'dirC/bar';\n\n              import type { E3 } from 'dirC/baz';\n              import type { F3 } from 'dirC/caz';\n              import type { C1 } from 'dirA/Bar';\n\n              import type {\n                D1,\n                X1,\n                Y1\n              } from 'dirA/bar';\n\n              import type { E1 } from 'dirA/baz';\n              import type { F } from './index.js';\n              import type { G } from './aaa.js';\n              import type { H } from './bbb';\n            ",
    },
    {
      name: 'run 0 invalid 133',
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
              patternOptions: {
                matchBase: true,
              },
            },
            {
              pattern: '*.+(css|svg)',
              group: 'type',
              position: 'after',
              patternOptions: {
                matchBase: true,
              },
            },
          ],
          pathGroupsExcludedImportTypes: [],
          alphabetize: {
            order: 'asc',
          },
          'newlines-between': 'always',
        },
      ],
      errors: [
        {
          messageId: 'order',
          message:
            '`path/path/path/helpers` import should occur before import of `path/path/path/src/Cfg`',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 58,
        },
        {
          messageId: 'order',
          message:
            '`path/path/tip` import should occur before import of `path/src/l10n`',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 45,
        },
        {
          messageId: 'order',
          message:
            '`../../../../path/path/path/component` import should occur before import of `../../../../path/path/path/controller`',
          line: 8,
          column: 9,
          endLine: 8,
          endColumn: 74,
        },
      ],
      output:
        "\n        import { helpers } from 'path/path/path/helpers';\n        import { cfg } from 'path/path/path/src/Cfg';\n        import { tip } from 'path/path/tip';\n        import { l10n } from 'path/src/l10n';\n\n        import { component } from '../../../../path/path/path/component';\n        import { controller } from '../../../../path/path/path/controller';\n      ",
    },
  ],
});
