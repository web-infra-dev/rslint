import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-useless-backreference', {
  valid: [
    // Not a regular expression
    String.raw`regExp('\\1(a)')`,
    String.raw`new Regexp('\\1(a)', 'u')`,
    String.raw`RegExp.foo('\\1(a)', 'u')`,
    String.raw`new foo.RegExp('\\1(a)')`,

    // Unknown pattern
    'RegExp(p)',
    "new RegExp(p, 'u')",
    String.raw`RegExp('\\1(a)' + suffix)`,
    'new RegExp(`${prefix}\\\\1(a)`)',

    // Not the global RegExp
    String.raw`let RegExp; new RegExp('\\1(a)');`,
    String.raw`function foo() { var RegExp; RegExp('\\1(a)', 'u'); }`,
    String.raw`function foo(RegExp) { new RegExp('\\1(a)'); }`,
    String.raw`if (foo) { const RegExp = bar; RegExp('\\1(a)'); }`,

    // No capturing groups
    '/(?:)/',
    '/(?:a)/',
    "new RegExp('')",
    "RegExp('(?:a)|(?:b)*')",
    String.raw`/^ab|[cd].\n$/`,

    // No backreferences
    '/(a)/',
    "RegExp('(a)|(b)')",
    String.raw`new RegExp('\\n\\d(a)')`,
    String.raw`/\0(a)/`,
    String.raw`/\0(a)/u`,
    '/(?<=(a))(b)(?=(c))/',
    '/(?<!(a))(b)(?!(c))/',
    '/(?<foo>a)/',

    // Not really a backreference
    String.raw`RegExp('\\\\1(a)')`,
    String.raw`/\\1(a)/`,
    String.raw`/\1/`,
    String.raw`/^\1$/`,
    String.raw`/\2(a)/`,
    String.raw`/\1(?:a)/`,
    String.raw`/\1(?=a)/`,
    String.raw`/\1(?!a)/`,
    String.raw`/^[\1](a)$/`,
    String.raw`new RegExp('[\\1](a)')`,
    String.raw`/\11(a)/`,
    String.raw`/\k<foo>(a)/`,
    String.raw`/^(a)\1\2$/`,

    // Valid backreferences
    String.raw`/(a)\1/`,
    String.raw`/(a).\1/`,
    String.raw`RegExp('(a)\\1(b)')`,
    String.raw`/(a)(b)\2(c)/`,
    String.raw`/(?<foo>a)\k<foo>/`,
    String.raw`new RegExp('(.)\\1')`,
    String.raw`RegExp('(a)\\1(?:b)')`,
    String.raw`/(a)b\1/`,
    String.raw`/((a)\2)/`,
    String.raw`/^(?:(a)\1)$/`,
    String.raw`/^((a)\2)$/`,
    String.raw`/^(((a)\3))|b$/`,
    String.raw`/(a)?(b)*(\1)(c)/`,
    String.raw`/(?<=(a))b\1/`,
    String.raw`/(?<=(?=(a)\1))b/`,

    // Backreference before the group in same lookbehind
    String.raw`/(?<!\1(a))b/`,
    String.raw`/(?<=\1(a))b/`,
    String.raw`/(?<!\1.(a))b/`,
    String.raw`/(?<=\1.(a))b/`,
    String.raw`/(?=(?<=\1(a)))b/`,
    String.raw`/(.)(?<=\2(a))b/`,

    // Not into another alternative
    String.raw`/^(a)\1|b/`,
    String.raw`/^a|(b)\1/`,
    String.raw`/^a|(b|c)\1/`,
    String.raw`/^(a)|(b)\2/`,
    String.raw`/^(?:(a)|(b)\2)$/`,

    // Not into a negative lookaround
    String.raw`/.(?=(b))\1/`,
    String.raw`/.(?<=(b))\1/`,
    String.raw`/a(?!(b)\1)./`,
    String.raw`/a(?<!\1(b))./`,
    String.raw`/(?<!(a))(b)(?!(c))\2/`,
    String.raw`/a(?!(b|c)\1)./`,

    // Syntax errors
    String.raw`RegExp('\\1(a)[')`,
    String.raw`new RegExp('\\1(a){', 'u')`,
    String.raw`new RegExp('\\1(a)\\2', 'ug')`,
    String.raw`const flags = 'gus'; RegExp('\\1(a){', flags);`,
    String.raw`RegExp('\\1(a)\\k<foo>', 'u')`,
    String.raw`new RegExp('\\k<foo>(?<foo>a)\\k<bar>')`,

    // ES2025 named-duplicate alternatives
    String.raw`/((?<foo>bar)\k<foo>|(?<foo>baz))/`,
  ],
  invalid: [
    // Nested
    {
      code: String.raw`new RegExp('(\\1)')`,
      errors: [{ messageId: 'nested' }],
    },
    {
      code: String.raw`/^(a\1)$/`,
      errors: [{ messageId: 'nested' }],
    },
    {
      code: String.raw`/^((a)\1)$/`,
      errors: [{ messageId: 'nested' }],
    },
    {
      code: String.raw`/(b)(\2a)/`,
      errors: [{ messageId: 'nested' }],
    },
    {
      code: String.raw`/a(?<foo>(.)b\1)/`,
      errors: [{ messageId: 'nested' }],
    },
    {
      code: String.raw`/a(?<foo>\k<foo>)b/`,
      errors: [{ messageId: 'nested' }],
    },
    {
      code: String.raw`/(?<=(a\1))b/`,
      errors: [{ messageId: 'nested' }],
    },

    // Forward
    {
      code: String.raw`/\1(a)/`,
      errors: [{ messageId: 'forward' }],
    },
    {
      code: String.raw`/\k<foo>(?<foo>bar)/`,
      errors: [{ messageId: 'forward' }],
    },
    {
      code: String.raw`/(\2)(a)/`,
      errors: [{ messageId: 'forward' }],
    },
    {
      code: String.raw`RegExp('(a)\\2(b)')`,
      errors: [{ messageId: 'forward' }],
    },
    {
      code: String.raw`/\1(?<=(a))./`,
      errors: [{ messageId: 'forward' }],
    },

    // Backward in same lookbehind
    {
      code: String.raw`/(?<=(a)\1)b/`,
      errors: [{ messageId: 'backward' }],
    },
    {
      code: String.raw`/(?<!(a)\1)b/`,
      errors: [{ messageId: 'backward' }],
    },
    {
      code: String.raw`/(.)(?<!(b|c)\2)d/`,
      errors: [{ messageId: 'backward' }],
    },

    // Disjunctive
    {
      code: String.raw`/(a)|\1b/`,
      errors: [{ messageId: 'disjunctive' }],
    },
    {
      code: String.raw`/^(?:(a)|\1b)$/`,
      errors: [{ messageId: 'disjunctive' }],
    },
    {
      code: String.raw`RegExp('(a|bc)|\\1')`,
      errors: [{ messageId: 'disjunctive' }],
    },

    // Into negative lookaround
    {
      code: String.raw`/a(?!(b)).\1/`,
      errors: [{ messageId: 'intoNegativeLookaround' }],
    },
    {
      code: String.raw`/(?<!(a))b\1/`,
      errors: [{ messageId: 'intoNegativeLookaround' }],
    },
    {
      code: String.raw`new RegExp('(?!(?<foo>\\n))\\1')`,
      errors: [{ messageId: 'intoNegativeLookaround' }],
    },

    // Multiple invalid
    {
      code: String.raw`/\1(a)\2(b)/`,
      errors: [{ messageId: 'forward' }, { messageId: 'forward' }],
    },
    {
      code: String.raw`/\1.(?<=(a)\1)/`,
      errors: [{ messageId: 'forward' }, { messageId: 'backward' }],
    },

    // Statically known expression
    {
      code: String.raw`const r = RegExp, p = '\\1', s = '(a)'; new r(p + s);`,
      errors: [{ messageId: 'forward' }],
    },

    // Non-evaluable flags assumed to lack 'u'
    {
      code: String.raw`RegExp('\\1(a){', flags);`,
      errors: [{ messageId: 'forward' }],
    },

    // ES2024 v-flag
    {
      code: String.raw`new RegExp('\\1([[A--B]])', 'v')`,
      errors: [{ messageId: 'forward' }],
    },

    // ES2025 named-duplicate alternatives
    {
      code: String.raw`/\k<foo>((?<foo>bar)|(?<foo>baz))/`,
      errors: [{ messageId: 'forward' }],
    },
    {
      code: String.raw`/((?<foo>bar)|\k<foo>|(?<foo>baz))/`,
      errors: [{ messageId: 'disjunctive' }],
    },
    {
      code: String.raw`/((?<foo>bar)|(?<foo>baz\k<foo>)|(?<foo>qux\k<foo>))/`,
      errors: [{ messageId: 'nested' }, { messageId: 'nested' }],
    },
  ],
});
