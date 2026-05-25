import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('computed-property-spacing', null as never, {
  valid: [
    // default - never
    { code: 'obj[foo]' },
    { code: "obj['foo']" },
    { code: 'var x = {[b]: a}' },

    // always
    { code: 'obj[ foo ]', options: ['always'] },
    { code: 'obj[\nfoo\n]', options: ['always'] },
    { code: "obj[ 'foo' ]", options: ['always'] },
    { code: "obj[ 'foo' + 'bar' ]", options: ['always'] },
    { code: 'obj[ obj2[ foo ] ]', options: ['always'] },
    { code: 'var foo = obj[ 1 ]', options: ['always'] },
    { code: "var foo = obj[ 'foo' ];", options: ['always'] },
    { code: 'var foo = obj[ [1, 1] ];', options: ['always'] },

    // always - objectLiteralComputedProperties
    { code: 'var x = {[ "a" ]: a}', options: ['always'] },
    { code: 'var y = {[ x ]: a}', options: ['always'] },
    { code: 'var x = {[ "a" ]() {}}', options: ['always'] },
    { code: 'var y = {[ x ]() {}}', options: ['always'] },

    // never
    { code: 'obj[foo]', options: ['never'] },
    { code: "obj['foo']", options: ['never'] },
    { code: "obj['foo' + 'bar']", options: ['never'] },
    { code: 'obj[obj2[foo]]', options: ['never'] },
    { code: 'obj[\nfoo]', options: ['never'] },
    { code: 'obj[foo\n]', options: ['never'] },
    { code: 'var foo = obj[1]', options: ['never'] },
    { code: 'var foo = obj[[ 1, 1 ]];', options: ['never'] },

    // never - objectLiteralComputedProperties
    { code: 'var x = {["a"]: a}', options: ['never'] },
    { code: 'var y = {[x]: a}', options: ['never'] },
    { code: 'var x = {["a"]() {}}', options: ['never'] },
    { code: 'var y = {[x]() {}}', options: ['never'] },

    // Classes — enforceForClassMembers
    {
      code: 'class A { [ a ](){} }',
      options: ['never', { enforceForClassMembers: false }],
    },
    {
      code: 'A = class { [a](){} }',
      options: ['always', { enforceForClassMembers: false }],
    },
    {
      code: 'class A { [ a ]; }',
      options: ['never', { enforceForClassMembers: false }],
    },
    {
      code: 'class A { [a]; }',
      options: ['always', { enforceForClassMembers: false }],
    },

    // Classes — valid spacing
    {
      code: 'A = class { [a](){} }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { [ a ](){} }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { [a]; static [a]; [a] = 0; static [a] = 0; }',
      options: ['never', { enforceForClassMembers: true }],
    },

    // Destructuring Assignment
    { code: 'const { [a]: someProp } = obj;', options: ['never'] },
    { code: '({ [a]: someProp } = obj);', options: ['never'] },
    { code: 'const { [ a ]: someProp } = obj;', options: ['always'] },

    // TS — AccessorProperty & IndexedAccessType
    {
      code: 'class A { accessor [b] = 1 }',
      options: ['never', { enforceForClassMembers: true }],
    },
    { code: 'type Foo = A[B]' },
  ],

  invalid: [
    {
      code: 'var foo = obj[ 1];',
      output: 'var foo = obj[ 1 ];',
      options: ['always'],
      errors: [
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
    {
      code: 'var foo = obj[1 ];',
      output: 'var foo = obj[ 1 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'obj[ foo ]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'var foo = obj[1]',
      output: 'var foo = obj[ 1 ]',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // always - objectLiteralComputedProperties
    {
      code: 'var x = {[a]: b}',
      output: 'var x = {[ a ]: b}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },

    // never - objectLiteralComputedProperties
    {
      code: 'var x = {[ a ]: b}',
      output: 'var x = {[a]: b}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // default class behavior (enforceForClassMembers: true)
    {
      code: 'class A { [ a ](){} }',
      output: 'class A { [a](){} }',
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
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },

    // never - classes
    {
      code: 'class A { [ a](){} }',
      output: 'class A { [a](){} }',
      options: ['never', { enforceForClassMembers: true }],
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

    // always - classes
    {
      code: 'A = class { [a](){} }',
      output: 'A = class { [ a ](){} }',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },

    // Optional chaining
    {
      code: 'obj?.[1];',
      output: 'obj?.[ 1 ];',
      options: ['always'],
      errors: [
        { messageId: 'missingSpaceAfter', data: { tokenValue: '[' } },
        { messageId: 'missingSpaceBefore', data: { tokenValue: ']' } },
      ],
    },
    {
      code: 'obj?.[ 1 ];',
      output: 'obj?.[1];',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedSpaceAfter', data: { tokenValue: '[' } },
        { messageId: 'unexpectedSpaceBefore', data: { tokenValue: ']' } },
      ],
    },

    // Destructuring
    {
      code: 'const { [ a ]: someProp } = obj;',
      output: 'const { [a]: someProp } = obj;',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedSpaceAfter', data: { tokenValue: '[' } },
        { messageId: 'unexpectedSpaceBefore', data: { tokenValue: ']' } },
      ],
    },

    // TS — AccessorProperty + IndexedAccessType
    {
      code: 'class A { accessor [ a ] = 0 }',
      output: 'class A { accessor [a] = 0 }',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        { messageId: 'unexpectedSpaceAfter', column: 21, endColumn: 22 },
        { messageId: 'unexpectedSpaceBefore', column: 23, endColumn: 24 },
      ],
    },
    {
      code: 'type Foo = A[ B ]',
      output: 'type Foo = A[B]',
      errors: [
        { messageId: 'unexpectedSpaceAfter', data: { tokenValue: '[' } },
        { messageId: 'unexpectedSpaceBefore', data: { tokenValue: ']' } },
      ],
    },
    {
      code: 'type Foo = A[B]',
      output: 'type Foo = A[ B ]',
      options: ['always'],
      errors: [
        { messageId: 'missingSpaceAfter', data: { tokenValue: '[' } },
        { messageId: 'missingSpaceBefore', data: { tokenValue: ']' } },
      ],
    },
  ],
});
