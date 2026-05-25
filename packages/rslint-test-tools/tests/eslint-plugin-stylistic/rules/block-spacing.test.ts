import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('block-spacing', null as never, {
  valid: [
    // default/always
    { code: '{ foo(); }', options: ['always'] },
    { code: '{ foo(); }' },
    { code: '{ foo();\n}' },
    { code: '{\nfoo(); }' },
    { code: '{\r\nfoo();\r\n}' },
    { code: 'if (a) { foo(); }' },
    { code: 'if (a) {} else { foo(); }' },
    { code: 'switch (a) {}' },
    { code: 'switch (a) { case 0: foo(); }' },
    { code: 'while (a) { foo(); }' },
    { code: 'do { foo(); } while (a);' },
    { code: 'for (;;) { foo(); }' },
    { code: 'for (var a in b) { foo(); }' },
    { code: 'for (var a of b) { foo(); }' },
    { code: 'try { foo(); } catch (e) { foo(); }' },
    { code: 'function foo() { bar(); }' },
    { code: '(function() { bar(); });' },
    { code: '(() => { bar(); });' },
    { code: 'if (a) { /* comment */ foo(); /* comment */ }' },
    { code: 'if (a) { //comment\n foo(); }' },
    { code: 'class C { static {} }' },
    { code: 'class C { static { foo; } }' },
    { code: 'class C { static { /* comment */foo;/* comment */ } }' },

    // never
    { code: '{foo();}', options: ['never'] },
    { code: '{foo();\n}', options: ['never'] },
    { code: '{\nfoo();}', options: ['never'] },
    { code: '{\r\nfoo();\r\n}', options: ['never'] },
    { code: 'if (a) {foo();}', options: ['never'] },
    { code: 'if (a) {} else {foo();}', options: ['never'] },
    { code: 'switch (a) {}', options: ['never'] },
    { code: 'switch (a) {case 0: foo();}', options: ['never'] },
    { code: 'while (a) {foo();}', options: ['never'] },
    { code: 'do {foo();} while (a);', options: ['never'] },
    { code: 'for (;;) {foo();}', options: ['never'] },
    { code: 'for (var a in b) {foo();}', options: ['never'] },
    { code: 'for (var a of b) {foo();}', options: ['never'] },
    { code: 'try {foo();} catch (e) {foo();}', options: ['never'] },
    { code: 'function foo() {bar();}', options: ['never'] },
    { code: '(function() {bar();});', options: ['never'] },
    { code: '(() => {bar();});', options: ['never'] },
    {
      code: 'if (a) {/* comment */ foo(); /* comment */}',
      options: ['never'],
    },
    { code: 'if (a) { //comment\n foo();}', options: ['never'] },
    { code: 'class C { static { } }', options: ['never'] },
    { code: 'class C { static {foo;} }', options: ['never'] },
    {
      code: 'class C { static {/* comment */ foo; /* comment */} }',
      options: ['never'],
    },
    {
      code: 'class C { static { // line comment is allowed\n foo;\n} }',
      options: ['never'],
    },
    { code: 'class C { static {\nfoo;\n} }', options: ['never'] },
    { code: 'class C { static { \n foo; \n } }', options: ['never'] },
  ],
  invalid: [
    // default/always
    {
      code: '{foo();}',
      output: '{ foo(); }',
      options: ['always'],
      errors: [
        {
          line: 1,
          column: 1,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
        {
          line: 1,
          column: 8,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: '{foo();}',
      output: '{ foo(); }',
      errors: [
        {
          line: 1,
          column: 1,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
        {
          line: 1,
          column: 8,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: '{ foo();}',
      output: '{ foo(); }',
      errors: [
        {
          line: 1,
          column: 9,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: '{foo(); }',
      output: '{ foo(); }',
      errors: [
        {
          line: 1,
          column: 1,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
      ],
    },
    {
      code: '{\nfoo();}',
      output: '{\nfoo(); }',
      errors: [
        {
          line: 2,
          column: 7,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: '{foo();\n}',
      output: '{ foo();\n}',
      errors: [
        {
          line: 1,
          column: 1,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
      ],
    },
    {
      code: 'if (a) {foo();}',
      output: 'if (a) { foo(); }',
      errors: [
        {
          line: 1,
          column: 8,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
        {
          line: 1,
          column: 15,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: 'switch (a) {case 0: foo();}',
      output: 'switch (a) { case 0: foo(); }',
      errors: [
        {
          line: 1,
          column: 12,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
        {
          line: 1,
          column: 27,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: 'function foo() {bar();}',
      output: 'function foo() { bar(); }',
      errors: [
        {
          line: 1,
          column: 16,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
        {
          line: 1,
          column: 23,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: '(() => {bar();});',
      output: '(() => { bar(); });',
      errors: [
        {
          line: 1,
          column: 8,
          messageId: 'missing',
          data: { location: 'after', token: '{' },
        },
        {
          line: 1,
          column: 15,
          messageId: 'missing',
          data: { location: 'before', token: '}' },
        },
      ],
    },
    {
      code: 'class C { static {foo;} }',
      output: 'class C { static { foo; } }',
      errors: [
        {
          messageId: 'missing',
          data: { location: 'after', token: '{' },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
        {
          messageId: 'missing',
          data: { location: 'before', token: '}' },
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // never
    {
      code: '{ foo(); }',
      output: '{foo();}',
      options: ['never'],
      errors: [
        {
          messageId: 'extra',
          data: { location: 'after', token: '{' },
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 3,
        },
        {
          messageId: 'extra',
          data: { location: 'before', token: '}' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'if (a) { foo(); }',
      output: 'if (a) {foo();}',
      options: ['never'],
      errors: [
        {
          messageId: 'extra',
          data: { location: 'after', token: '{' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'extra',
          data: { location: 'before', token: '}' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'switch (a) { case 0: foo(); }',
      output: 'switch (a) {case 0: foo();}',
      options: ['never'],
      errors: [
        {
          messageId: 'extra',
          data: { location: 'after', token: '{' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'extra',
          data: { location: 'before', token: '}' },
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'function foo() { bar(); }',
      output: 'function foo() {bar();}',
      options: ['never'],
      errors: [
        {
          messageId: 'extra',
          data: { location: 'after', token: '{' },
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'extra',
          data: { location: 'before', token: '}' },
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'class C { static { foo; } }',
      output: 'class C { static {foo;} }',
      options: ['never'],
      errors: [
        {
          messageId: 'extra',
          data: { location: 'after', token: '{' },
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'extra',
          data: { location: 'before', token: '}' },
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
  ],
});
