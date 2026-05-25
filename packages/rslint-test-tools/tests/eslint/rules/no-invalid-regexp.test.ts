import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-invalid-regexp', {
  valid: [
    "RegExp('.')",
    "new RegExp('.')",
    "new RegExp('.', 'im')",
    "new RegExp('.', 'gmi')",
    "new RegExp('.', 'dgimsuy')",
    "new RegExp(pattern, 'g')",
    "new RegExp('.', flags)",
    // No arguments
    'RegExp()',
    // Non-string pattern types
    'RegExp(`pattern`)',
    'RegExp(pattern)',
    "RegExp('[' + '')",
    'RegExp(123)',
    // Non-RegExp callee
    "global.RegExp('.', 'z')",
    "window.RegExp('.', 'z')",
    "foo.RegExp('.', 'z')",
    "regexp('.', 'z')",
    // Non-literal flags
    "new RegExp('.', `g`)",
    "new RegExp('.', 'g' + 'i')",
    // Non-literal pattern + non-literal flags
    'new RegExp(pattern, flags)',
    // Non-literal pattern + valid flags
    "new RegExp(foo, 'gi')",
    "new RegExp(pattern, '')",
    // Empty flags
    "new RegExp('.', '')",
    // All flags with v (without u)
    "new RegExp('.', 'dgimsvy')",

    // Skipped: regexp2 engine limitations (FP — valid in ESLint but regexp2 rejects)
    // Unicode property long names
    { code: "new RegExp('\\\\p{Letter}', 'u');", skip: true },
    // Unicode Script= syntax
    { code: "new RegExp('\\\\p{Script=Nandinagari}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Cpmn}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Cypro_Minoan}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Old_Uyghur}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Ougr}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Tangsa}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Tnsa}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Toto}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Vith}', 'u');", skip: true },
    { code: "new RegExp('\\\\p{Script=Vithkuqi}', 'u');", skip: true },
    // v-flag set notation
    { code: "new RegExp('[A--B]', 'v');", skip: true },
    { code: "new RegExp('[A--[0-9]]', 'v');", skip: true },
    {
      code: "new RegExp('[\\\\p{Basic_Emoji}--\\\\q{a|bc|def}]', 'v');",
      skip: true,
    },
    { code: "new RegExp('[A--B]', flags);", skip: true },
    // Surrogate pair named capture groups
    { code: "new RegExp('(?<\\\\ud835\\\\udc9c>.)', 'g');", skip: true },
    { code: "new RegExp('(?<\\\\u{1d49c}>.)', 'g');", skip: true },
  ],
  invalid: [
    // Invalid flags
    {
      code: "RegExp('.', 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('.', 'G');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('.', 'abc');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('.', 'gz');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Duplicate flags
    {
      code: "new RegExp('.', 'gg');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('.', 'dd');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('.', 'giig');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // u+v conflict
    {
      code: "RegExp('.', 'uv');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('.', 'guv');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Invalid pattern
    {
      code: "RegExp('[');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "RegExp('(');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Invalid pattern with non-literal flags
    {
      code: "new RegExp('[', flags);",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Non-literal pattern + flag errors
    {
      code: "RegExp(foo, 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp(foo, 'gg');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "RegExp(foo, 'uv');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Template/binary pattern + invalid flags
    {
      code: "RegExp(`test`, 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "RegExp('a' + 'b', 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Parenthesized callee
    {
      code: "(RegExp)('.', 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Nesting: inside function call
    {
      code: "foo(new RegExp('['));",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Nesting: inside array
    {
      code: "[RegExp('[')];",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Nesting: arrow function
    {
      code: "const fn = () => RegExp('.', 'z');",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Nesting: class method
    {
      code: "class C { m() { new RegExp('['); } }",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Nesting: default parameter
    {
      code: "function f(x = new RegExp('[')) {}",
      errors: [{ messageId: 'regexMessage' }],
    },
    // Nesting: logical expression
    {
      code: "true && new RegExp('.', 'zz');",
      errors: [{ messageId: 'regexMessage' }],
    },

    // Skipped: regexp2 engine limitations (FN — invalid in ESLint but regexp2 misses)
    // Invalid escape in unicode mode
    {
      code: "new RegExp('\\\\a', 'u');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    // v-flag specific parsing
    {
      code: "new RegExp('[[]', 'v');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('[[]\\\\u{0}*', 'v');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    // Duplicate named capture groups outside alternatives
    {
      code: "new RegExp('(?<k>a)(?<k>b)');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    // Inline modifier validation
    {
      code: "new RegExp('(?ii:foo)');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('(?-ii:foo)');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('(?i-i:foo)');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('(?-:foo)');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    {
      code: "new RegExp('(?-u:foo)');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
    // Trailing backslash
    {
      code: "new RegExp('\\\\');",
      skip: true,
      errors: [{ messageId: 'regexMessage' }],
    },
  ],
});
