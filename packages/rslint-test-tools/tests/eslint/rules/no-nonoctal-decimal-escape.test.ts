import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-nonoctal-decimal-escape', {
  valid: [
    // Plain digits (no backslash) — not escapes at all.
    "'8'",
    "'9'",
    "'foo8'",
    "'foo9bar'",
    // Escaped backslash before the digit — `\\8` and `\\9` are a literal
    // backslash followed by a plain digit, not a `\8` / `\9` escape.
    "'\\\\8'",
    "'\\\\9'",
    "'\\\\8\\\\9'",
    "'\\\\\\\\9'",
    // Standard escape sequences.
    "'\\n'",
    "'\\0'",
    // Non-string nodes are out of scope; the rule listens on string literals.
    'var x = 8;',
    'var \\u8888',
    'var re = /\\8/;',
  ],
  // Invalid cases cannot be tested through the JS test framework because the
  // TypeScript parser reports `\8` / `\9` as syntax errors (TS1488),
  // preventing program creation. The detector is exercised in
  // TestScanDecimalEscapes (Go), and the diagnostic-emission pipeline
  // (positions + suggestion outputs) is exercised in
  // TestNoNonoctalDecimalEscapeDiagnostics (Go) using a lenient program.
  invalid: [],
});
