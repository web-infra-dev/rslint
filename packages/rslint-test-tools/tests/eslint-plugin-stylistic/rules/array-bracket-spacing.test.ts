import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('array-bracket-spacing', null as never, {
  valid: [
    { code: 'var foo = obj[ 1 ]', options: ['always'] },
    { code: "var foo = obj[ 'foo' ];", options: ['always'] },
    { code: 'var foo = obj[ [ 1, 1 ] ];', options: ['always'] },

    // always - singleValue
    { code: "var foo = ['foo']", options: ['always', { singleValue: false }] },
    { code: 'var foo = [2]', options: ['always', { singleValue: false }] },
    {
      code: 'var foo = [[ 1, 1 ]]',
      options: ['always', { singleValue: false }],
    },
    {
      code: "var foo = [{ 'foo': 'bar' }]",
      options: ['always', { singleValue: false }],
    },
    { code: 'var foo = [bar]', options: ['always', { singleValue: false }] },

    // always - objectsInArrays
    {
      code: "var foo = [{ 'bar': 'baz' }, 1,  5 ];",
      options: ['always', { objectsInArrays: false }],
    },
    {
      code: "var foo = [ 1, 5, { 'bar': 'baz' }];",
      options: ['always', { objectsInArrays: false }],
    },
    {
      code: "var foo = [{ 'bar': 'baz' }]",
      options: ['always', { objectsInArrays: false }],
    },
    {
      code: 'var foo = [ function(){} ];',
      options: ['always', { objectsInArrays: false }],
    },

    // always - arraysInArrays
    {
      code: 'var arr = [[ 1, 2 ], 2, 3, 4 ];',
      options: ['always', { arraysInArrays: false }],
    },
    {
      code: 'var foo = [ arr[i], arr[j] ];',
      options: ['always', { arraysInArrays: false }],
    },

    // always
    { code: 'obj[ foo ]', options: ['always'] },
    { code: "obj[ 'foo' ]", options: ['always'] },
    { code: 'var arr = [ 1, 2, 3, 4 ];', options: ['always'] },
    { code: 'var foo = [];', options: ['always'] },

    // always - destructuring assignment
    { code: 'var [ x, y ] = z', options: ['always'] },
    { code: 'var [ ,x, ] = z', options: ['always'] },
    {
      code: 'var [[ x, y ], z ] = arr;',
      options: ['always', { arraysInArrays: false }],
    },

    // never
    { code: 'obj[foo]', options: ['never'] },
    { code: "obj['foo']", options: ['never'] },
    { code: 'var arr = [1, 2, 3, 4];', options: ['never'] },
    { code: 'var arr = [[1, 2], 2, 3, 4];', options: ['never'] },

    // never - destructuring assignment
    { code: 'var [x, y] = z', options: ['never'] },
    { code: 'var [,x,] = z', options: ['never'] },
    {
      code: 'var [ [x, y], z] = arr;',
      options: ['never', { arraysInArrays: true }],
    },

    // never - singleValue
    { code: "var foo = [ 'foo' ]", options: ['never', { singleValue: true }] },
    { code: 'var foo = [ 2 ]', options: ['never', { singleValue: true }] },
    { code: 'var foo = [ bar ]', options: ['never', { singleValue: true }] },

    // never - objectsInArrays
    {
      code: "var foo = [ {'bar': 'baz'}, 1, 5];",
      options: ['never', { objectsInArrays: true }],
    },
    {
      code: "var foo = [1, 5, {'bar': 'baz'} ];",
      options: ['never', { objectsInArrays: true }],
    },
    { code: 'var foo = [];', options: ['never', { objectsInArrays: true }] },

    // never - arraysInArrays
    {
      code: 'var arr = [ [1, 2], 2, 3, 4];',
      options: ['never', { arraysInArrays: true }],
    },
    { code: 'var foo = [];', options: ['never', { arraysInArrays: true }] },

    // should not warn
    { code: 'var foo = {};', options: ['never'] },
    { code: 'var foo = [];', options: ['never'] },
    {
      code: "var foo = [{'bar':'baz'}, 1, {'bar': 'baz'}];",
      options: ['never'],
    },
    { code: "var obj = {'foo': [1, 2]}", options: ['never'] },
  ],
  invalid: [
    {
      code: 'var foo = [ ]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // objectsInArrays
    {
      code: "var foo = [ { 'bar': 'baz' }, 1,  5];",
      options: ['always', { objectsInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },

    // singleValue
    {
      code: "var obj = [ 'foo' ];",
      options: ['always', { singleValue: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "var obj = ['foo'];",
      options: ['never', { singleValue: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
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
      options: ['always', { arraysInArrays: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // always - destructuring
    {
      code: 'var [x,y] = y',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'var [...horse] = y',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },

    // never -  arraysInArrays
    {
      code: 'var arr = [[1, 2], 2, [3, 4]];',
      options: ['never', { arraysInArrays: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var arr = [ ];',
      options: ['never', { arraysInArrays: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
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
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // never
    {
      code: 'var arr = [ 1, 2, 3, 4 ];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // multiple spaces
    {
      code: 'var arr = [  1, 2   ];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
  ],
});
