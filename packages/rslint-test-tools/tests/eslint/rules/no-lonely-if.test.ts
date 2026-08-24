import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors the upstream no-lonely-if valid/invalid semantic set (Layer 1). It
// verifies rule registration + wire protocol + ESLint-compatible diagnostic
// shape; the exhaustive edge-shape / branch lock-in coverage lives in the Go
// suite.
ruleTester.run('no-lonely-if', {
  valid: [
    'if (a) {;} else if (b) {;}',
    'if (a) {;} else { if (b) {;} ; }',
    'if (a) if (a) {} else { if (b) {} } else {}',
  ],
  invalid: [
    {
      code: 'if (a) {;} else { if (b) {;} }',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else {\n  if (b) {\n    bar();\n  }\n}',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else /* comment */ {\n  if (b) {\n    bar();\n  }\n}',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else {\n  /* otherwise, do the other thing */ if (b) {\n    bar();\n  }\n}',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else {\n  if /* this comment is ok */ (b) {\n    bar();\n  }\n}',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else {\n  if (b) {\n    bar();\n  } /* this comment will prevent this test case from being autofixed. */\n}',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (foo) {} else { if (bar) baz(); }',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (foo) {} else { if (bar) baz() } qux();',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (foo) {} else { if (bar) baz(); } qux();',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (foo) {\n} else {\n  if (bar) baz()\n}\n[1, 2, 3].forEach(foo);',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (foo) {\n} else {\n  if (bar) baz++\n}\nfoo;',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (foo) {\n} else {\n  if (bar) baz++;\n}\nfoo;',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else {\n  if (b) bar()\n}\n`template literal`;',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
    {
      code: 'if (a) {\n  foo();\n} else {\n  if (b) {\n    bar();\n  } else if (c) {\n    baz();\n  } else {\n    qux();\n  }\n}',
      errors: [{ messageId: 'unexpectedLonelyIf' }],
    },
  ],
});
