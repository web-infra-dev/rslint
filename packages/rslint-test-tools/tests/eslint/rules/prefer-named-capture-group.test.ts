import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-named-capture-group', {
  valid: [
    '/normal_regex/',
    '/(?:[0-9]{4})/',
    '/(?<year>[0-9]{4})/',
    '/\\u{1F680}/u',
    'new RegExp()',
    'new RegExp(foo)',
    "new RegExp('')",
    "new RegExp('(?<year>[0-9]{4})')",
    'RegExp()',
    'RegExp(foo)',
    "RegExp('')",
    "RegExp('(?<year>[0-9]{4})')",
    "RegExp('(')", // invalid regexp should be ignored
    "RegExp('\\\\u{1F680}', 'u')",

    // Mirrors upstream ecmaVersion 6 / 2017 cases: globalThis doesn't exist as a
    // global before ES2020, so globalThis.RegExp isn't tracked. The JS test
    // harness doesn't expose a per-case ecmaVersion override; the Go upstream
    // suite (prefer_named_capture_group_upstream_test.go) covers these directly.
    { code: "new globalThis.RegExp('([0-9]{4})')", skip: true },
    { code: "new globalThis.RegExp('([0-9]{4})')", skip: true },
    { code: "new globalThis.RegExp('([0-9]{4})')", skip: true },

    'new globalThis.RegExp()',
    'new globalThis.RegExp(foo)',
    'globalThis.RegExp(foo)',
    `
                var globalThis = bar;
                globalThis.RegExp(foo);
                `,
    `
                function foo () {
                    var globalThis = bar;
                    new globalThis.RegExp(baz);
                }
                `,

    // ES2024
    "new RegExp('(?<c>[[A--B]])', 'v')",

    // Checks that the 'v' flag is understood: without it `([\\q])` would be a
    // valid (unnamed-capturing-group) regex; with 'v' it's a SyntaxError.
    "new RegExp('([\\\\q])', 'v')",

    // ES2025
    '/(?i:foo)bar/',
    "new RegExp('(?i:foo)bar')",
    '/(?-i:foo)bar/',
    "new RegExp('(?-i:foo)bar')",
  ],
  invalid: [
    {
      code: '/([0-9]{4})/',
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp('([0-9]{4})')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "RegExp('([0-9]{4})')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: 'new RegExp(`a(bc)d`)',
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp('ሴ噸(?:a)(b)');",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp('\\\\u1234\\\\u5678(?:a)(b)');",
      errors: [{ messageId: 'required' }],
    },
    {
      code: '/([0-9]{4})-(\\w{5})/',
      errors: [{ messageId: 'required' }, { messageId: 'required' }],
    },
    {
      code: '/([0-9]{4})-(5)/',
      errors: [{ messageId: 'required' }, { messageId: 'required' }],
    },
    {
      code: '/(?<temp2>(a))/',
      errors: [{ messageId: 'required' }],
    },
    {
      code: '/(?<temp2>(a)(?<temp5>b))/',
      errors: [{ messageId: 'required' }],
    },
    {
      code: '/(?<temp1>[0-9]{4})-(\\w{5})/',
      errors: [{ messageId: 'required' }],
    },
    {
      code: '/(?<temp1>[0-9]{4})-(5)/',
      errors: [{ messageId: 'required' }],
    },
    {
      code: '/(?<temp1>a)(?<temp2>a)(a)(?<temp3>a)/',
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp('(' + 'a)')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp('a(bc)d' + 'e')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: 'new RegExp("foo" + "(a)" + "(b)");',
      errors: [{ messageId: 'required' }, { messageId: 'required' }],
    },
    {
      code: 'new RegExp("foo" + "(?:a)" + "(b)");',
      errors: [{ messageId: 'required' }],
    },
    {
      code: "RegExp('(a)'+'')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "RegExp( '' + '(ab)')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp(`(ab)${''}`)",
      errors: [{ messageId: 'required' }],
    },
    {
      code: 'new RegExp(`(a)\n`)',
      errors: [{ messageId: 'required' }],
    },
    {
      code: 'RegExp(`a(b\nc)d`)',
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new RegExp('a(b)\\'')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "RegExp('(a)\\\\d')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: 'RegExp(`\\\\a(b)`)',
      errors: [{ messageId: 'required' }],
    },
    {
      code: "new globalThis.RegExp('([0-9]{4})')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: "globalThis.RegExp('([0-9]{4})')",
      errors: [{ messageId: 'required' }],
    },
    {
      code: `
                function foo() { var globalThis = bar; }
                new globalThis.RegExp('([0-9]{4})');
            `,
      errors: [{ messageId: 'required' }],
    },

    // ES2024
    {
      code: "new RegExp('([[A--B]])', 'v')",
      errors: [{ messageId: 'required' }],
    },
  ],
});
