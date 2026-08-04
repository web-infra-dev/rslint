import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// The JS suite mirrors upstream Layer 1 only: it verifies rule registration,
// the IPC wire protocol, and ESLint-compatible diagnostic shape. Edge-shape
// (Dimension 4), branch lock-in, and type-aware cases live in the Go
// prefer_array_some_extras_test.go / _upstream_test.go files.

const suggestionFind = (output: string, method = 'find') => [
  {
    messageId: 'some-suggestion',
    data: { method },
    output,
  },
];

const valid = (code: string) => ({ code });

ruleTester.run('prefer-array-some', null as never, {
  valid: [
    // `.find()` result not used as a boolean
    valid('const bar = foo.find(fn)'),
    valid('const bar = foo.find(fn) || baz'),
    valid('if (foo.find(fn) ?? bar) {}'),
    valid('let hasFoo = foo.find(fn); if (hasFoo) {}'),
    valid('const hasFoo = foo.find(fn); hasFoo.value;'),
    valid('const {hasFoo} = foo.find(fn); if (hasFoo) {}'),
    valid('export const hasFoo = foo.find(fn);'),
    valid('const hasFoo = foo.find(fn); if (hasFoo === undefined) {}'),
    valid('const Boolean = value => value; if (Boolean(foo.find(fn))) {}'),

    // receiver definitely not an array
    valid('if (new Foo().find(fn)) {}'),
    valid('if (new Foo().findLast(fn)) {}'),
    valid('new Foo().findIndex(fn) !== -1'),
    valid('new Foo().filter(fn).length > 0'),
    valid('if ("foo".find(fn)) {}'),

    // not a matched CallExpression
    valid('if (find(fn)) {}'),
    valid('if (foo["find"](fn)) {}'),
    valid('if (foo.notFind(fn)) {}'),
    valid('if (foo.find()) {}'),
    valid('if (foo.find(fn, thisArgument, extraArgument)) {}'),
    valid('if (foo.find(...argumentsArray)) {}'),

    // `.filter().length` valid comparisons
    valid('array.filter(fn).length > 0.'),
    valid('array.filter(fn).length < 0'),
    valid('array.filter(fn).length >= 0'),
    valid('array.filter(fn).length != 0'),
    valid('array.filter(fn).length === 0'),
    valid('array.filter(fn).length >= 1'),
    valid('array.filter(fn)?.length > 0'),
    valid('array.filter?.(fn).length > 0'),
    valid('array?.filter(fn).length > 0'),
    valid('$element.filter(":visible").length > 0'),

    // `.findIndex()` valid comparisons
    valid('foo.notMatchedMethod(bar) !== -1'),
    valid('new foo.findIndex(bar) !== -1'),
    valid('foo.findIndex(bar, extraArgument) !== -1'),
    valid('foo.findIndex(...bar) !== -1'),

    // compare-with-undefined valid
    valid('foo.find(fn) == 0'),
    valid('foo.find(fn) === null'),
    valid('foo.find(fn) >= undefined'),
    valid('typeof foo.find(fn) === "undefined"'),
  ],
  invalid: [
    // boolean-position `.find()` / `.findLast()`
    {
      code: 'const bar = !foo.find(fn)',
      errors: [
        {
          messageId: 'some',
          suggestions: suggestionFind('const bar = !foo.some(fn)'),
        },
      ],
    },
    {
      code: 'if (foo.find(fn)) {}',
      errors: [
        {
          messageId: 'some',
          suggestions: suggestionFind('if (foo.some(fn)) {}'),
        },
      ],
    },
    {
      code: 'if (foo.findLast(fn)) {}',
      errors: [
        {
          messageId: 'some',
          suggestions: suggestionFind('if (foo.some(fn)) {}', 'findLast'),
        },
      ],
    },
    {
      code: 'const bar = Boolean(foo.find(fn))',
      errors: [
        {
          messageId: 'some',
          suggestions: suggestionFind('const bar = Boolean(foo.some(fn))'),
        },
      ],
    },
    {
      code: 'const bar = foo.find(fn) ? 1 : 2',
      errors: [
        {
          messageId: 'some',
          suggestions: suggestionFind('const bar = foo.some(fn) ? 1 : 2'),
        },
      ],
    },
    // variable used only as boolean (array receiver)
    {
      code: 'const hasFoo = [].find(fn); if (hasFoo) {}',
      errors: [
        {
          messageId: 'some',
          suggestions: suggestionFind(
            'const hasFoo = [].some(fn); if (hasFoo) {}',
          ),
        },
      ],
    },
    // actual message text
    {
      code: 'if (bar.find(fn)) {}',
      errors: [
        {
          message: 'Prefer `.some(…)` over `.find(…)`.',
          suggestions: [
            {
              messageId: 'some-suggestion',
              data: { method: 'find' },
              output: 'if (bar.some(fn)) {}',
            },
          ],
        },
      ],
    },

    // `.findIndex()` compared → autofix
    {
      code: 'foo.findIndex(bar) !== -1',
      output: 'foo.some(bar) ',
      errors: [{ messageId: 'some' }],
    },
    {
      code: 'foo.findIndex(bar) === -1',
      output: '!foo.some(bar) ',
      errors: [{ messageId: 'some' }],
    },
    {
      code: 'foo.findLastIndex(bar) >= 0',
      output: 'foo.some(bar) ',
      errors: [{ messageId: 'some' }],
    },
    {
      code: 'foo.findIndex(bar) < 0',
      output: '!foo.some(bar) ',
      errors: [{ messageId: 'some' }],
    },

    // `.filter().length` → autofix
    {
      code: 'array.filter(fn).length > 0',
      output: 'array.some(fn)',
      errors: [{ messageId: 'filter' }],
    },
    {
      code: 'array.filter(fn).length !== 0',
      output: 'array.some(fn)',
      errors: [{ messageId: 'filter' }],
    },

    // compare-with-undefined → suggestion
    {
      code: 'foo.find(fn) == null',
      errors: [
        { messageId: 'some', suggestions: suggestionFind('!foo.some(fn)') },
      ],
    },
    {
      code: 'foo.find(fn) !== undefined',
      errors: [
        { messageId: 'some', suggestions: suggestionFind('foo.some(fn)') },
      ],
    },
  ],
});
