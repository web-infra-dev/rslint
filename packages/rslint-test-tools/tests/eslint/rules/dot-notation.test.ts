import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { config } from '../config';

// Skipped cases retain their upstream records; re-enabling them also requires complete expectations.
new RuleTester(config).run('dot-notation', {
  valid: [
    'a.b;',
    'a.b.c;',
    "a['12'];",
    'a[b];',
    'a[0];',
    {
      code: 'a.b.c;',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a.arguments;',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a.let;',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a.yield;',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a.eval;',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a[0];',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: "a['while'];",
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: "a['true'];",
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: "a['null'];",
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a[true];',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a[null];',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    {
      code: 'a.true;',
      options: [
        {
          allowKeywords: true,
        },
      ],
    },
    {
      code: 'a.null;',
      options: [
        {
          allowKeywords: true,
        },
      ],
    },
    {
      code: "a['snake_case'];",
      options: [
        {
          allowPattern: '^[a-z]+(_[a-z]+)+$',
        },
      ],
    },
    {
      code: "a['lots_of_snake_case'];",
      options: [
        {
          allowPattern: '^[a-z]+(_[a-z]+)+$',
        },
      ],
    },
    'a[`time${range}`];',
    {
      code: 'a[`while`];',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
    'a[`time range`];',
    'a.true;',
    'a.null;',
    'a[undefined];',
    'a[void 0];',
    'a[b()];',
    'a[/(?<zero>0)/];',
    "class C { foo() { this['#a'] } }",
    {
      code: 'class C { #in; foo() { this.#in; } }',
      options: [
        {
          allowKeywords: false,
        },
      ],
    },
  ],
  invalid: [
    {
      code: 'a.true;',
      options: [
        {
          allowKeywords: false,
        },
      ],
      errors: [
        {
          messageId: 'useBrackets',
          message: '.true is a syntax error.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 7,
        },
      ],
      output: 'a["true"];',
    },
    {
      code: "a['true'];",
      errors: [
        {
          messageId: 'useDot',
          message: '["true"] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: 'a.true;',
    },
    {
      code: 'a[`time`];',
      errors: [
        {
          messageId: 'useDot',
          message: '[`time`] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: 'a.time;',
    },
    {
      code: 'a[null];',
      errors: [
        {
          messageId: 'useDot',
          message: '[null] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 7,
        },
      ],
      output: 'a.null;',
    },
    {
      code: 'a[true];',
      errors: [
        {
          messageId: 'useDot',
          message: '[true] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 7,
        },
      ],
      output: 'a.true;',
    },
    {
      code: 'a[false];',
      errors: [
        {
          messageId: 'useDot',
          message: '[false] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 8,
        },
      ],
      output: 'a.false;',
    },
    {
      code: "a['b'];",
      errors: [
        {
          messageId: 'useDot',
          message: '["b"] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 6,
        },
      ],
      output: 'a.b;',
    },
    {
      code: "a.b['c'];",
      errors: [
        {
          messageId: 'useDot',
          message: '["c"] is better written in dot notation.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
        },
      ],
      output: 'a.b.c;',
    },
    {
      code: "a['_dangle'];",
      options: [
        {
          allowPattern: '^[a-z]+(_[a-z]+)+$',
        },
      ],
      errors: [
        {
          messageId: 'useDot',
          message: '["_dangle"] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: 'a._dangle;',
    },
    {
      code: "a['SHOUT_CASE'];",
      options: [
        {
          allowPattern: '^[a-z]+(_[a-z]+)+$',
        },
      ],
      errors: [
        {
          messageId: 'useDot',
          message: '["SHOUT_CASE"] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 15,
        },
      ],
      output: 'a.SHOUT_CASE;',
    },
    {
      code: `a
  ['SHOUT_CASE'];`,
      errors: [
        {
          messageId: 'useDot',
          message: '["SHOUT_CASE"] is better written in dot notation.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 16,
        },
      ],
      output: `a
  .SHOUT_CASE;`,
    },
    {
      code: `getResource()
    .then(function(){})
    ["catch"](function(){})
    .then(function(){})
    ["catch"](function(){});`,
      errors: [
        {
          messageId: 'useDot',
          message: '["catch"] is better written in dot notation.',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 13,
        },
        {
          messageId: 'useDot',
          message: '["catch"] is better written in dot notation.',
          line: 5,
          column: 6,
          endLine: 5,
          endColumn: 13,
        },
      ],
      output: `getResource()
    .then(function(){})
    .catch(function(){})
    .then(function(){})
    .catch(function(){});`,
    },
    {
      code: `foo
  .while;`,
      options: [
        {
          allowKeywords: false,
        },
      ],
      errors: [
        {
          messageId: 'useBrackets',
          message: '.while is a syntax error.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 9,
        },
      ],
      output: `foo
  ["while"];`,
    },
    {
      code: "foo[ /* comment */ 'bar' ]",
      errors: [
        {
          messageId: 'useDot',
          message: '["bar"] is better written in dot notation.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "foo[ 'bar' /* comment */ ]",
      errors: [
        {
          messageId: 'useDot',
          message: '["bar"] is better written in dot notation.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: "foo[    'bar'    ];",
      errors: [
        {
          messageId: 'useDot',
          message: '["bar"] is better written in dot notation.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 14,
        },
      ],
      output: 'foo.bar;',
    },
    {
      code: 'foo. /* comment */ while',
      options: [
        {
          allowKeywords: false,
        },
      ],
      errors: [
        {
          messageId: 'useBrackets',
          message: '.while is a syntax error.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "foo[('bar')]",
      errors: [
        {
          messageId: 'useDot',
          message: '["bar"] is better written in dot notation.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 11,
        },
      ],
      output: 'foo.bar',
    },
    {
      code: 'foo[(null)]',
      errors: [
        {
          messageId: 'useDot',
          message: '[null] is better written in dot notation.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'foo.null',
    },
    {
      code: "(foo)['bar']",
      errors: [
        {
          messageId: 'useDot',
          message: '["bar"] is better written in dot notation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 12,
        },
      ],
      output: '(foo).bar',
    },
    {
      code: "1['toString']",
      errors: [
        {
          messageId: 'useDot',
          message: '["toString"] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: '1 .toString',
    },
    {
      code: "foo['bar']instanceof baz",
      errors: [
        {
          messageId: 'useDot',
          message: '["bar"] is better written in dot notation.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'foo.bar instanceof baz',
    },
    {
      code: 'let.if()',
      options: [
        {
          allowKeywords: false,
        },
      ],
      errors: [
        {
          messageId: 'useBrackets',
          message: '.if is a syntax error.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: "5['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 9,
        },
      ],
      output: '5 .prop',
    },
    {
      code: "-5['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: '-5 .prop',
    },
    {
      skip: 'The native TypeScript parser rejects this legacy leading-zero numeric literal before rule execution.',
      code: "01['prop']",
      errors: [
        {
          messageId: 'useDot',
        },
      ],
    } as unknown as TestCase,
    {
      skip: 'The native TypeScript parser rejects this legacy leading-zero numeric literal before rule execution.',
      code: "01234567['prop']",
      errors: [
        {
          messageId: 'useDot',
        },
      ],
    } as unknown as TestCase,
    {
      skip: 'The native TypeScript parser rejects this legacy leading-zero numeric literal before rule execution.',
      code: "08['prop']",
      errors: [
        {
          messageId: 'useDot',
        },
      ],
    } as unknown as TestCase,
    {
      skip: 'The native TypeScript parser rejects this legacy leading-zero numeric literal before rule execution.',
      code: "090['prop']",
      errors: [
        {
          messageId: 'useDot',
        },
      ],
    } as unknown as TestCase,
    {
      skip: 'The native TypeScript parser rejects this legacy leading-zero numeric literal before rule execution.',
      code: "018['prop']",
      errors: [
        {
          messageId: 'useDot',
        },
      ],
    } as unknown as TestCase,
    {
      code: "5_000['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: '5_000 .prop',
    },
    {
      code: "5_000_00['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output: '5_000_00 .prop',
    },
    {
      code: "5.000_000['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
        },
      ],
      output: '5.000_000.prop',
    },
    {
      code: "0b1010_1010['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 19,
        },
      ],
      output: '0b1010_1010.prop',
    },
    {
      code: "obj?.['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
        },
      ],
      output: 'obj?.prop',
    },
    {
      code: "0?.['prop']",
      errors: [
        {
          messageId: 'useDot',
          message: '["prop"] is better written in dot notation.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 11,
        },
      ],
      output: '0?.prop',
    },
    {
      code: 'obj?.true',
      options: [
        {
          allowKeywords: false,
        },
      ],
      errors: [
        {
          messageId: 'useBrackets',
          message: '.true is a syntax error.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'obj?.["true"]',
    },
    {
      code: 'let?.true',
      options: [
        {
          allowKeywords: false,
        },
      ],
      errors: [
        {
          messageId: 'useBrackets',
          message: '.true is a syntax error.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
      output: 'let?.["true"]',
    },
  ],
});
