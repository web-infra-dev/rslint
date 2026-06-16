/**
 * @fileoverview Tests for array-bracket-spacing rule.
 * @author Ian Christian Myers
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/array-bracket-spacing/array-bracket-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('array-bracket-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (ecmaVersion) dropped — rslint resolves via tsconfig.
 *  - `type` fields (deprecated AST node type) dropped (none were present).
 *
 * The upstream file contains NO `$` unindent template tags, NO spread/helper
 * error builders, NO `readFileSync` external-fixture cases, and NO `suggestions`.
 * The `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * The second upstream `run()` block (`array-bracket-spacing_babel`, guarded by
 * `if (!skipBabel)` and using `languageOptionsForBabelFlow`) is a Babel/Flow
 * parser suite — those cases are moved to `KNOWN GAPS` at the bottom (rslint has
 * no Babel/Flow parser).
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('array-bracket-spacing', null as never, {
  valid: [
    { code: 'var foo = obj[ 1 ]', options: ['always'] },
    { code: 'var foo = obj[ \'foo\' ];', options: ['always'] },
    { code: 'var foo = obj[ [ 1, 1 ] ];', options: ['always'] },

    // always - singleValue
    { code: 'var foo = [\'foo\']', options: ['always', { singleValue: false }] },
    { code: 'var foo = [2]', options: ['always', { singleValue: false }] },
    { code: 'var foo = [[ 1, 1 ]]', options: ['always', { singleValue: false }] },
    { code: 'var foo = [{ \'foo\': \'bar\' }]', options: ['always', { singleValue: false }] },
    { code: 'var foo = [bar]', options: ['always', { singleValue: false }] },

    // always - objectsInArrays
    { code: 'var foo = [{ \'bar\': \'baz\' }, 1,  5 ];', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [ 1, 5, { \'bar\': \'baz\' }];', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [{\n\'bar\': \'baz\', \n\'qux\': [{ \'bar\': \'baz\' }], \n\'quxx\': 1 \n}]', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [{ \'bar\': \'baz\' }]', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [{ \'bar\': \'baz\' }, 1, { \'bar\': \'baz\' }];', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [ 1, { \'bar\': \'baz\' }, 5 ];', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [ 1, { \'bar\': \'baz\' }, [{ \'bar\': \'baz\' }] ];', options: ['always', { objectsInArrays: false }] },
    { code: 'var foo = [ function(){} ];', options: ['always', { objectsInArrays: false }] },

    // always - arraysInArrays
    { code: 'var arr = [[ 1, 2 ], 2, 3, 4 ];', options: ['always', { arraysInArrays: false }] },
    { code: 'var arr = [[ 1, 2 ], [[[ 1 ]]], 3, 4 ];', options: ['always', { arraysInArrays: false }] },
    { code: 'var foo = [ arr[i], arr[j] ];', options: ['always', { arraysInArrays: false }] },

    // always - arraysInArrays, objectsInArrays
    { code: 'var arr = [[ 1, 2 ], 2, 3, { \'foo\': \'bar\' }];', options: ['always', { arraysInArrays: false, objectsInArrays: false }] },

    // always - arraysInArrays, objectsInArrays, singleValue
    { code: 'var arr = [[ 1, 2 ], [2], 3, { \'foo\': \'bar\' }];', options: ['always', { arraysInArrays: false, objectsInArrays: false, singleValue: false }] },

    // always
    { code: 'obj[ foo ]', options: ['always'] },
    { code: 'obj[\nfoo\n]', options: ['always'] },
    { code: 'obj[ \'foo\' ]', options: ['always'] },
    { code: 'obj[ \'foo\' + \'bar\' ]', options: ['always'] },
    { code: 'obj[ obj2[ foo ] ]', options: ['always'] },
    { code: 'obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })', options: ['always'] },
    { code: 'obj[ \'map\' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })', options: ['always'] },
    { code: 'obj[ \'for\' + \'Each\' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })', options: ['always'] },

    { code: 'var arr = [ 1, 2, 3, 4 ];', options: ['always'] },
    { code: 'var arr = [ [ 1, 2 ], 2, 3, 4 ];', options: ['always'] },
    { code: 'var arr = [\n1, 2, 3, 4\n];', options: ['always'] },
    { code: 'var foo = [];', options: ['always'] },

    // singleValue: false, objectsInArrays: true, arraysInArrays
    { code: 'this.db.mappings.insert([\n { alias: \'a\', url: \'http://www.amazon.de\' },\n { alias: \'g\', url: \'http://www.google.de\' }\n], function() {});', options: ['always', { singleValue: false, objectsInArrays: true, arraysInArrays: true }] },

    // always - destructuring assignment
    { code: 'var [ x, y ] = z', options: ['always'] },
    { code: 'var [ x,y ] = z', options: ['always'] },
    { code: 'var [ x, y\n] = z', options: ['always'] },
    { code: 'var [\nx, y ] = z', options: ['always'] },
    { code: 'var [\nx, y\n] = z', options: ['always'] },
    { code: 'var [\nx,,,\n] = z', options: ['always'] },
    { code: 'var [ ,x, ] = z', options: ['always'] },
    { code: 'var [\nx, ...y\n] = z', options: ['always'] },
    { code: 'var [\nx, ...y ] = z', options: ['always'] },
    { code: 'var [[ x, y ], z ] = arr;', options: ['always', { arraysInArrays: false }] },
    { code: 'var [ x, [ y, z ]] = arr;', options: ['always', { arraysInArrays: false }] },
    { code: '[{ x, y }, z ] = arr;', options: ['always', { objectsInArrays: false }] },
    { code: '[ x, { y, z }] = arr;', options: ['always', { objectsInArrays: false }] },

    // never
    { code: 'obj[foo]', options: ['never'] },
    { code: 'obj[\'foo\']', options: ['never'] },
    { code: 'obj[\'foo\' + \'bar\']', options: ['never'] },
    { code: 'obj[\'foo\'+\'bar\']', options: ['never'] },
    { code: 'obj[obj2[foo]]', options: ['never'] },
    { code: 'obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })', options: ['never'] },
    { code: 'obj[\'map\'](function(item) { return [\n1,\n2,\n3,\n4\n]; })', options: ['never'] },
    { code: 'obj[\'for\' + \'Each\'](function(item) { return [\n1,\n2,\n3,\n4\n]; })', options: ['never'] },
    { code: 'var arr = [1, 2, 3, 4];', options: ['never'] },
    { code: 'var arr = [[1, 2], 2, 3, 4];', options: ['never'] },
    { code: 'var arr = [\n1, 2, 3, 4\n];', options: ['never'] },
    { code: 'obj[\nfoo]', options: ['never'] },
    { code: 'obj[foo\n]', options: ['never'] },
    { code: 'var arr = [1,\n2,\n3,\n4\n];', options: ['never'] },
    { code: 'var arr = [\n1,\n2,\n3,\n4];', options: ['never'] },

    // never - destructuring assignment
    { code: 'var [x, y] = z', options: ['never'] },
    { code: 'var [x,y] = z', options: ['never'] },
    { code: 'var [x, y\n] = z', options: ['never'] },
    { code: 'var [\nx, y] = z', options: ['never'] },
    { code: 'var [\nx, y\n] = z', options: ['never'] },
    { code: 'var [\nx,,,\n] = z', options: ['never'] },
    { code: 'var [,x,] = z', options: ['never'] },
    { code: 'var [\nx, ...y\n] = z', options: ['never'] },
    { code: 'var [\nx, ...y] = z', options: ['never'] },
    { code: 'var [ [x, y], z] = arr;', options: ['never', { arraysInArrays: true }] },
    { code: 'var [x, [y, z] ] = arr;', options: ['never', { arraysInArrays: true }] },
    { code: '[ { x, y }, z] = arr;', options: ['never', { objectsInArrays: true }] },
    { code: '[x, { y, z } ] = arr;', options: ['never', { objectsInArrays: true }] },

    // never - singleValue
    { code: 'var foo = [ \'foo\' ]', options: ['never', { singleValue: true }] },
    { code: 'var foo = [ 2 ]', options: ['never', { singleValue: true }] },
    { code: 'var foo = [ [1, 1] ]', options: ['never', { singleValue: true }] },
    { code: 'var foo = [ {\'foo\': \'bar\'} ]', options: ['never', { singleValue: true }] },
    { code: 'var foo = [ bar ]', options: ['never', { singleValue: true }] },

    // never - objectsInArrays
    { code: 'var foo = [ {\'bar\': \'baz\'}, 1, 5];', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [1, 5, {\'bar\': \'baz\'} ];', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [ {\n\'bar\': \'baz\', \n\'qux\': [ {\'bar\': \'baz\'} ], \n\'quxx\': 1 \n} ]', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [ {\'bar\': \'baz\'} ]', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [ {\'bar\': \'baz\'}, 1, {\'bar\': \'baz\'} ];', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [1, {\'bar\': \'baz\'} , 5];', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [1, {\'bar\': \'baz\'}, [ {\'bar\': \'baz\'} ]];', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [function(){}];', options: ['never', { objectsInArrays: true }] },
    { code: 'var foo = [];', options: ['never', { objectsInArrays: true }] },

    // never - arraysInArrays
    { code: 'var arr = [ [1, 2], 2, 3, 4];', options: ['never', { arraysInArrays: true }] },
    { code: 'var foo = [arr[i], arr[j]];', options: ['never', { arraysInArrays: true }] },
    { code: 'var foo = [];', options: ['never', { arraysInArrays: true }] },

    // never - arraysInArrays, singleValue
    { code: 'var arr = [ [1, 2], [ [ [ 1 ] ] ], 3, 4];', options: ['never', { arraysInArrays: true, singleValue: true }] },

    // never - arraysInArrays, objectsInArrays
    { code: 'var arr = [ [1, 2], 2, 3, {\'foo\': \'bar\'} ];', options: ['never', { arraysInArrays: true, objectsInArrays: true }] },

    // should not warn
    { code: 'var foo = {};', options: ['never'] },
    { code: 'var foo = [];', options: ['never'] },

    { code: 'var foo = [{\'bar\':\'baz\'}, 1, {\'bar\': \'baz\'}];', options: ['never'] },
    { code: 'var foo = [{\'bar\': \'baz\'}];', options: ['never'] },
    { code: 'var foo = [{\n\'bar\': \'baz\', \n\'qux\': [{\'bar\': \'baz\'}], \n\'quxx\': 1 \n}]', options: ['never'] },
    { code: 'var foo = [1, {\'bar\': \'baz\'}, 5];', options: ['never'] },
    { code: 'var foo = [{\'bar\': \'baz\'}, 1,  5];', options: ['never'] },
    { code: 'var foo = [1, 5, {\'bar\': \'baz\'}];', options: ['never'] },
    { code: 'var obj = {\'foo\': [1, 2]}', options: ['never'] },
  ],

  invalid: [
    {
      code: 'var foo = [ ]',
      output: 'var foo = []',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // objectsInArrays
    {
      code: 'var foo = [ { \'bar\': \'baz\' }, 1,  5];',
      output: 'var foo = [{ \'bar\': \'baz\' }, 1,  5 ];',
      options: ['always', { objectsInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'missingSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'var foo = [1, 5, { \'bar\': \'baz\' } ];',
      output: 'var foo = [ 1, 5, { \'bar\': \'baz\' }];',
      options: ['always', { objectsInArrays: false }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'var foo = [ { \'bar\':\'baz\' }, 1, { \'bar\': \'baz\' } ];',
      output: 'var foo = [{ \'bar\':\'baz\' }, 1, { \'bar\': \'baz\' }];',
      options: ['always', { objectsInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },

    // singleValue
    {
      code: 'var obj = [ \'foo\' ];',
      output: 'var obj = [\'foo\'];',
      options: ['always', { singleValue: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var obj = [\'foo\' ];',
      output: 'var obj = [\'foo\'];',
      options: ['always', { singleValue: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var obj = [\'foo\'];',
      output: 'var obj = [ \'foo\' ];',
      options: ['never', { singleValue: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // always - arraysInArrays
    {
      code: 'var arr = [ [ 1, 2 ], 2, 3, 4 ];',
      output: 'var arr = [[ 1, 2 ], 2, 3, 4 ];',
      options: ['always', { arraysInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var arr = [ 1, 2, 2, [ 3, 4 ] ];',
      output: 'var arr = [ 1, 2, 2, [ 3, 4 ]];',
      options: ['always', { arraysInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'var arr = [[ 1, 2 ], 2, [ 3, 4 ] ];',
      output: 'var arr = [[ 1, 2 ], 2, [ 3, 4 ]];',
      options: ['always', { arraysInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'var arr = [ [ 1, 2 ], 2, [ 3, 4 ]];',
      output: 'var arr = [[ 1, 2 ], 2, [ 3, 4 ]];',
      options: ['always', { arraysInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var arr = [ [ 1, 2 ], 2, [ 3, 4 ] ];',
      output: 'var arr = [[ 1, 2 ], 2, [ 3, 4 ]];',
      options: ['always', { arraysInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // always - destructuring
    {
      code: 'var [x,y] = y',
      output: 'var [ x,y ] = y',
      options: ['always'],
      errors: [{
        messageId: 'missingSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }, {
        messageId: 'missingSpaceBefore',
        data: {
          tokenValue: ']',
        },
        line: 1,
        column: 9,
        endLine: 1,
        endColumn: 10,
      }],
    },
    {
      code: 'var [x,y ] = y',
      output: 'var [ x,y ] = y',
      options: ['always'],
      errors: [{
        messageId: 'missingSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'var [,,,x,,] = y',
      output: 'var [ ,,,x,, ] = y',
      options: ['always'],
      errors: [{
        messageId: 'missingSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }, {
        messageId: 'missingSpaceBefore',
        data: {
          tokenValue: ']',
        },
        line: 1,
        column: 12,
        endLine: 1,
        endColumn: 13,
      }],
    },
    {
      code: 'var [ ,,,x,,] = y',
      output: 'var [ ,,,x,, ] = y',
      options: ['always'],
      errors: [{
        messageId: 'missingSpaceBefore',
        data: {
          tokenValue: ']',
        },
        line: 1,
        column: 13,
        endLine: 1,
        endColumn: 14,
      }],
    },
    {
      code: 'var [...horse] = y',
      output: 'var [ ...horse ] = y',
      options: ['always'],
      errors: [{
        messageId: 'missingSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }, {
        messageId: 'missingSpaceBefore',
        data: {
          tokenValue: ']',
        },
        line: 1,
        column: 14,
        endLine: 1,
        endColumn: 15,
      }],
    },
    {
      code: 'var [...horse ] = y',
      output: 'var [ ...horse ] = y',
      options: ['always'],
      errors: [{
        messageId: 'missingSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'var [ [ x, y ], z ] = arr;',
      output: 'var [[ x, y ], z ] = arr;',
      options: ['always', { arraysInArrays: false }],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: '[ { x, y }, z ] = arr;',
      output: '[{ x, y }, z ] = arr;',
      options: ['always', { objectsInArrays: false }],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: {
          tokenValue: '[',
        },
        line: 1,
        column: 2,
        endLine: 1,
        endColumn: 3,
      }],
    },
    {
      code: '[ x, { y, z } ] = arr;',
      output: '[ x, { y, z }] = arr;',
      options: ['always', { objectsInArrays: false }],
      errors: [{
        messageId: 'unexpectedSpaceBefore',
        data: {
          tokenValue: ']',
        },
        line: 1,
        column: 14,
        endLine: 1,
        endColumn: 15,
      }],
    },

    // never -  arraysInArrays
    {
      code: 'var arr = [[1, 2], 2, [3, 4]];',
      output: 'var arr = [ [1, 2], 2, [3, 4] ];',
      options: ['never', { arraysInArrays: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var arr = [ ];',
      output: 'var arr = [];',
      options: ['never', { arraysInArrays: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // never -  objectsInArrays
    {
      code: 'var arr = [ ];',
      output: 'var arr = [];',
      options: ['never', { objectsInArrays: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // always
    {
      code: 'var arr = [1, 2, 3, 4];',
      output: 'var arr = [ 1, 2, 3, 4 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var arr = [1, 2, 3, 4 ];',
      output: 'var arr = [ 1, 2, 3, 4 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var arr = [ 1, 2, 3, 4];',
      output: 'var arr = [ 1, 2, 3, 4 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // never
    {
      code: 'var arr = [ 1, 2, 3, 4 ];',
      output: 'var arr = [1, 2, 3, 4];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var arr = [1, 2, 3, 4 ];',
      output: 'var arr = [1, 2, 3, 4];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var arr = [ 1, 2, 3, 4];',
      output: 'var arr = [1, 2, 3, 4];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var arr = [ [ 1], 2, 3, 4];',
      output: 'var arr = [[1], 2, 3, 4];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'var arr = [[1 ], 2, 3, 4 ];',
      output: 'var arr = [[1], 2, 3, 4];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // multiple spaces
    {
      code: 'var arr = [  1, 2   ];',
      output: 'var arr = [1, 2];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'function f( [   a, b  ] ) {}',
      output: 'function f( [a, b] ) {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'var arr = [ 1,\n   2   ];',
      output: 'var arr = [1,\n   2];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'var arr = [  1, [ 2, 3  ] ];',
      output: 'var arr = [1, [2, 3]];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: {
            tokenValue: '[',
          },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 25,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: {
            tokenValue: ']',
          },
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
  ],
});

/**
 * ===================== array-bracket-spacing — KNOWN GAPS =====================
 *
 * The cases below come from the SECOND upstream `run()` block
 * (`array-bracket-spacing_babel`, guarded by `if (!skipBabel)`), which parses
 * with `languageOptionsForBabelFlow` — i.e. the Babel parser with the Flow
 * plugin. rslint has no Babel/Flow parser; it parses every fixture with the
 * ts-go parser. These fixtures use Flow's `(... : Array<any>) => {}` arrow
 * typing on a destructuring parameter, which the upstream suite only exercises
 * under Babel/Flow. They are preserved here verbatim for the record, NOT run
 * through the green `ruleTester.run` above.
 *
 * ---- valid (upstream expects 0 diagnostics) ----
 *
 *   { code: '([ a, b ]: Array<any>) => {}',  options: ['always'] }   // Babel/Flow
 *   { code: '([a, b]: Array< any >) => {}',   options: ['never']  }   // Babel/Flow
 *
 * ---- invalid (upstream expects the given diagnostics + fix) ----
 *
 *   {
 *     code:   '([ a, b ]: Array<any>) => {}',
 *     output: '([a, b]: Array<any>) => {}',
 *     options: ['never'],   // Babel/Flow
 *     errors: [
 *       { messageId: 'unexpectedSpaceAfter',  data: { tokenValue: '[' }, line: 1, column: 3, endLine: 1, endColumn: 4 },
 *       { messageId: 'unexpectedSpaceBefore', data: { tokenValue: ']' }, line: 1, column: 8, endLine: 1, endColumn: 9 },
 *     ],
 *   }
 *   {
 *     code:   '([a, b]: Array< any >) => {}',
 *     output: '([ a, b ]: Array< any >) => {}',
 *     options: ['always'],  // Babel/Flow
 *     errors: [
 *       { messageId: 'missingSpaceAfter',  data: { tokenValue: '[' }, line: 1, column: 2, endLine: 1, endColumn: 3 },
 *       { messageId: 'missingSpaceBefore', data: { tokenValue: ']' }, line: 1, column: 7, endLine: 1, endColumn: 8 },
 *     ],
 *   }
 *
 * =============================================================================
 */
