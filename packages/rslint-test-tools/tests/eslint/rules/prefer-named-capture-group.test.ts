import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
new RuleTester(config).run('prefer-named-capture-group', {
  valid: [
    {
      code: '/normal_regex/',
    },
    {
      code: '/(?:[0-9]{4})/',
    },
    {
      code: '/(?<year>[0-9]{4})/',
    },
    {
      code: '/\\u{1F680}/u',
    },
    {
      code: 'new RegExp()',
    },
    {
      code: 'new RegExp(foo)',
    },
    {
      code: "new RegExp('')",
    },
    {
      code: "new RegExp('(?<year>[0-9]{4})')",
    },
    {
      code: 'RegExp()',
    },
    {
      code: 'RegExp(foo)',
    },
    {
      code: "RegExp('')",
    },
    {
      code: "RegExp('(?<year>[0-9]{4})')",
    },
    {
      code: "RegExp('(')",
    },
    {
      code: "RegExp('\\\\u{1F680}', 'u')",
    },
    {
      code: "new globalThis.RegExp('([0-9]{4})')",
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      code: "new globalThis.RegExp('([0-9]{4})')",
      languageOptions: {
        ecmaVersion: 2015,
      },
    },
    {
      code: "new globalThis.RegExp('([0-9]{4})')",
      languageOptions: {
        ecmaVersion: 2017,
      },
    },
    {
      code: 'new globalThis.RegExp()',
    },
    {
      code: 'new globalThis.RegExp(foo)',
    },
    {
      code: 'globalThis.RegExp(foo)',
    },
    {
      code: '\n                var globalThis = bar;\n                globalThis.RegExp(foo);\n                ',
    },
    {
      code: '\n                function foo () {\n                    var globalThis = bar;\n                    new globalThis.RegExp(baz);\n                }\n                ',
    },
    {
      code: "new RegExp('(?<c>[[A--B]])', 'v')",
    },
    {
      code: "new RegExp('([\\\\q])', 'v')",
    },
    {
      code: '/(?i:foo)bar/',
    },
    {
      code: "new RegExp('(?i:foo)bar')",
    },
    {
      code: '/(?-i:foo)bar/',
    },
    {
      code: "new RegExp('(?-i:foo)bar')",
    },
  ],
  invalid: [
    {
      code: '/([0-9]{4})/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp1>[0-9]{4})/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?:[0-9]{4})/',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('([0-9]{4})')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: "new RegExp('(?<temp1>[0-9]{4})')",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: "new RegExp('(?:[0-9]{4})')",
            },
          ],
        },
      ],
    },
    {
      code: "RegExp('([0-9]{4})')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: "RegExp('(?<temp1>[0-9]{4})')",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: "RegExp('(?:[0-9]{4})')",
            },
          ],
        },
      ],
    },
    {
      code: 'new RegExp(`a(bc)d`)',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(bc)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: 'new RegExp(`a(?<temp1>bc)d`)',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: 'new RegExp(`a(?:bc)d`)',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('ሴ噸(?:a)(b)');",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(b)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: "new RegExp('ሴ噸(?:a)(?<temp1>b)');",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: "new RegExp('ሴ噸(?:a)(?:b)');",
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('\\\\u1234\\\\u5678(?:a)(b)');",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(b)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '/([0-9]{4})-(\\w{5})/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp1>[0-9]{4})-(\\w{5})/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?:[0-9]{4})-(\\w{5})/',
            },
          ],
        },
        {
          messageId: 'required',
          message:
            "Capture group '(\\w{5})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/([0-9]{4})-(?<temp1>\\w{5})/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/([0-9]{4})-(?:\\w{5})/',
            },
          ],
        },
      ],
    },
    {
      code: '/([0-9]{4})-(5)/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp1>[0-9]{4})-(5)/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?:[0-9]{4})-(5)/',
            },
          ],
        },
        {
          messageId: 'required',
          message:
            "Capture group '(5)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/([0-9]{4})-(?<temp1>5)/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/([0-9]{4})-(?:5)/',
            },
          ],
        },
      ],
    },
    {
      code: '/(?<temp2>(a))/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp2>(?<temp3>a))/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?<temp2>(?:a))/',
            },
          ],
        },
      ],
    },
    {
      code: '/(?<temp2>(a)(?<temp5>b))/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp2>(?<temp6>a)(?<temp5>b))/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?<temp2>(?:a)(?<temp5>b))/',
            },
          ],
        },
      ],
    },
    {
      code: '/(?<temp1>[0-9]{4})-(\\w{5})/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(\\w{5})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp1>[0-9]{4})-(?<temp2>\\w{5})/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?<temp1>[0-9]{4})-(?:\\w{5})/',
            },
          ],
        },
      ],
    },
    {
      code: '/(?<temp1>[0-9]{4})-(5)/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(5)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp1>[0-9]{4})-(?<temp2>5)/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?<temp1>[0-9]{4})-(?:5)/',
            },
          ],
        },
      ],
    },
    {
      code: '/(?<temp1>a)(?<temp2>a)(a)(?<temp3>a)/',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: '/(?<temp1>a)(?<temp2>a)(?<temp4>a)(?<temp3>a)/',
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: '/(?<temp1>a)(?<temp2>a)(?:a)(?<temp3>a)/',
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('(' + 'a)')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "new RegExp('a(bc)d' + 'e')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(bc)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'new RegExp("foo" + "(a)" + "(b)");',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
        {
          messageId: 'required',
          message:
            "Capture group '(b)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'new RegExp("foo" + "(?:a)" + "(b)");',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(b)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: "RegExp('(a)'+'')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "RegExp( '' + '(ab)')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(ab)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "new RegExp(`(ab)${''}`)",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(ab)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: `new RegExp(\`(a)
\`)`,
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 3,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: `new RegExp(\`(?<temp1>a)
\`)`,
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: `new RegExp(\`(?:a)
\`)`,
            },
          ],
        },
      ],
    },
    {
      code: `RegExp(\`a(b
c)d\`)`,
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(b\nc)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 6,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: `RegExp(\`a(?<temp1>b
c)d\`)`,
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: `RegExp(\`a(?:b
c)d\`)`,
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('a(b)\\'')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(b)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "RegExp('(a)\\\\d')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(a)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'RegExp(`\\\\a(b)`)',
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '(b)' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "new globalThis.RegExp('([0-9]{4})')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: "new globalThis.RegExp('(?<temp1>[0-9]{4})')",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: "new globalThis.RegExp('(?:[0-9]{4})')",
            },
          ],
        },
      ],
    },
    {
      code: "globalThis.RegExp('([0-9]{4})')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: "globalThis.RegExp('(?<temp1>[0-9]{4})')",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: "globalThis.RegExp('(?:[0-9]{4})')",
            },
          ],
        },
      ],
    },
    {
      code: "\n                function foo() { var globalThis = bar; }\n                new globalThis.RegExp('([0-9]{4})');\n            ",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 52,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output:
                "\n                function foo() { var globalThis = bar; }\n                new globalThis.RegExp('(?<temp1>[0-9]{4})');\n            ",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output:
                "\n                function foo() { var globalThis = bar; }\n                new globalThis.RegExp('(?:[0-9]{4})');\n            ",
            },
          ],
        },
      ],
    },
    {
      code: "new RegExp('([[A--B]])', 'v')",
      errors: [
        {
          messageId: 'required',
          message:
            "Capture group '([[A--B]])' should be converted to a named or non-capturing group.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'addGroupName',
              desc: 'Add name to capture group.',
              output: "new RegExp('(?<temp1>[[A--B]])', 'v')",
            },
            {
              messageId: 'addNonCapture',
              desc: 'Convert group to non-capturing.',
              output: "new RegExp('(?:[[A--B]])', 'v')",
            },
          ],
        },
      ],
    },
  ],
});
