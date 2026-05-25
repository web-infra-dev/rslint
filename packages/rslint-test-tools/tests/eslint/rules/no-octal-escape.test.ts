import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-octal-escape', {
  valid: [
    // Hex escapes
    'var foo = "\\x51";',
    // Escaped backslash followed by digits
    'var foo = "foo \\\\251 bar";',
    // Regex backreference (not a string literal)
    'var foo = /([abc]) \\1/g;',
    // \\0 alone is a valid NULL character
    "var foo = '\\0';",
    "'\\0'",
    // \\0 followed by space or other non-digit
    "'\\0 '",
    "' \\0'",
    "'a\\0'",
    "'\\0a'",
    // Escaped backslash
    "'\\\\'",
    "'\\\\0'",
    "'\\\\08'",
    "'\\\\1'",
    "'\\\\01'",
    "'\\\\12'",
    // Escaped backslash followed by \\0
    "'\\\\\\\\\\0'",
    // \\0 followed by escaped backslash
    "'\\0\\\\'",
    // Plain digits (not escape sequences)
    "'0'",
    "'1'",
    "'8'",
    "'01'",
    "'08'",
    "'80'",
    "'12'",
    // Other escape sequences
    "'\\n'",
  ],
  // Invalid cases cannot be tested through the JS test framework because the
  // TypeScript parser reports octal escape sequences as syntax errors (TS1487),
  // preventing program creation. The detection logic is comprehensively tested
  // via Go unit tests (TestFindOctalEscape in no_octal_escape_test.go).
  invalid: [],
});
