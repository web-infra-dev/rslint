import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('dot-notation', {
  valid: [
    'a.b;',
    'a.b.c;',
    "a['12'];",
    'a[b];',
    'a[0];',
    { code: 'a.b.c;', options: { allowKeywords: false } },
    { code: 'a.arguments;', options: { allowKeywords: false } },
    { code: 'a.let;', options: { allowKeywords: false } },
    { code: 'a.yield;', options: { allowKeywords: false } },
    { code: 'a.eval;', options: { allowKeywords: false } },
    { code: 'a[0];', options: { allowKeywords: false } },
    { code: "a['while'];", options: { allowKeywords: false } },
    { code: "a['true'];", options: { allowKeywords: false } },
    { code: "a['null'];", options: { allowKeywords: false } },
    { code: 'a[true];', options: { allowKeywords: false } },
    { code: 'a[null];', options: { allowKeywords: false } },
    { code: 'a.true;', options: { allowKeywords: true } },
    { code: 'a.null;', options: { allowKeywords: true } },
    {
      code: "a['snake_case'];",
      options: { allowPattern: '^[a-z]+(_[a-z]+)+$' },
    },
    {
      code: "a['lots_of_snake_case'];",
      options: { allowPattern: '^[a-z]+(_[a-z]+)+$' },
    },
    'a[`time${range}`];',
    { code: 'a[`while`];', options: { allowKeywords: false } },
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
      options: { allowKeywords: false },
    },
  ],
  invalid: [
    {
      code: 'a.true;',
      options: { allowKeywords: false },
      errors: [{ messageId: 'useBrackets' }],
    },
    {
      code: "a['true'];",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'a[`time`];',
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'a[null];',
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'a[true];',
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'a[false];',
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "a['b'];",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "a.b['c'];",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "a['_dangle'];",
      options: { allowPattern: '^[a-z]+(_[a-z]+)+$' },
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "a['SHOUT_CASE'];",
      options: { allowPattern: '^[a-z]+(_[a-z]+)+$' },
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "a\n  ['SHOUT_CASE'];",
      errors: [{ messageId: 'useDot', line: 2, column: 4 }],
    },
    {
      code:
        'getResource()\n' +
        '    .then(function(){})\n' +
        '    ["catch"](function(){})\n' +
        '    .then(function(){})\n' +
        '    ["catch"](function(){});',
      errors: [{ messageId: 'useDot' }, { messageId: 'useDot' }],
    },
    {
      code: 'foo\n  .while;',
      options: { allowKeywords: false },
      errors: [{ messageId: 'useBrackets' }],
    },
    {
      code: "foo[ /* comment */ 'bar' ]",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "foo[ 'bar' /* comment */ ]",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "foo[    'bar'    ];",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'foo. /* comment */ while',
      options: { allowKeywords: false },
      errors: [{ messageId: 'useBrackets' }],
    },
    {
      code: "foo[('bar')]",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'foo[(null)]',
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "(foo)['bar']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "1['toString']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "foo['bar']instanceof baz",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'let.if()',
      options: { allowKeywords: false },
      errors: [{ messageId: 'useBrackets' }],
    },
    {
      code: "5['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "-5['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    // SKIP: legacy octal-style numeric literals (leading zero) are a hard
    // syntax error under tsgo's parser (TS1121 / TS1489), even in
    // non-strict scripts - a framework gap, not a rule-semantic one. See
    // dot_notation_upstream_test.go for the equivalent Go-side skips.
    {
      code: "01['prop']",
      errors: [{ messageId: 'useDot' }],
      skip: true,
    },
    {
      code: "01234567['prop']",
      errors: [{ messageId: 'useDot' }],
      skip: true,
    },
    {
      code: "08['prop']",
      errors: [{ messageId: 'useDot' }],
      skip: true,
    },
    {
      code: "090['prop']",
      errors: [{ messageId: 'useDot' }],
      skip: true,
    },
    {
      code: "018['prop']",
      errors: [{ messageId: 'useDot' }],
      skip: true,
    },
    {
      code: "5_000['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "5_000_00['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "5.000_000['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "0b1010_1010['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "obj?.['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: "0?.['prop']",
      errors: [{ messageId: 'useDot' }],
    },
    {
      code: 'obj?.true',
      options: { allowKeywords: false },
      errors: [{ messageId: 'useBrackets' }],
    },
    {
      code: 'let?.true',
      options: { allowKeywords: false },
      errors: [{ messageId: 'useBrackets' }],
    },
  ],
});
