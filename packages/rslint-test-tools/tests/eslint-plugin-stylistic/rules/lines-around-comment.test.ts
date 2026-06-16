/**
 * @fileoverview Test enforcement of lines around comments.
 * @author Jamund Ferguson
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/lines-around-comment/lines-around-comment._js_.test.ts
 *   packages/eslint-plugin/rules/lines-around-comment/lines-around-comment._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('lines-around-comment', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - `parserOptions` (ecmaVersion) dropped — rslint resolves via tsconfig.
 *  - `linterOptions.reportUnusedDisableDirectives` dropped — it tunes an ESLint
 *    core feature (unused eslint-disable reporting) unrelated to this rule; the
 *    rslint CLI doesn't surface that directive lint, so the rule's own
 *    diagnostics are unaffected.
 *
 * Every upstream case has an explicit `errors` array (no output-only invalid
 * cases) and there are no `suggestions`, spreads, or custom error helpers. The
 * `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * KNOWN GAPS: none. Every upstream case (157 valid + 104 invalid across both
 * the _js_ and _ts_ files — TS `interface` / `type` / `enum` / `module`
 * boundaries, `static {}` class static blocks, destructuring, hashbang comments
 * and the eslint-pragma ignorePattern cases) parses under ts-go and produces the
 * exact same diagnostic count, message, line and autofix output as upstream —
 * verified by running the suite (and by perturbation: a deliberately wrong line
 * or autofix expectation fails, so the green is real, not a no-op).
 */
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('lines-around-comment', null as never, {
  valid: [
    // ---- from lines-around-comment._js_.test.ts ----
    "bar()\n\n/** block block block\n * block \n */\n\nvar a = 1;",
    "bar()\n\n/** block block block\n * block \n */\nvar a = 1;",
    "bar()\n// line line line \nvar a = 1;",
    "bar()\n\n// line line line\nvar a = 1;",
    "bar()\n// line line line\n\nvar a = 1;",
    {
      code: "bar()\n// line line line\n\nvar a = 1;",
      options: [
        {
          afterLineComment: true,
        },
      ],
    },
    {
      code: "foo()\n\n// line line line\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
        },
      ],
    },
    {
      code: "foo()\n\n// line line line\n\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
    },
    {
      code: "foo()\n\n// line line line\n// line line\n\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
    },
    {
      code: "// line line line\n// line line",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
    },
    {
      code: "bar()\n\n/** A Block comment with a an empty line after\n *\n */\nvar a = 1;",
      options: [
        {
          afterBlockComment: false,
          beforeBlockComment: true,
        },
      ],
    },
    {
      code: "bar()\n\n/** block block block\n * block \n */\nvar a = 1;",
      options: [
        {
          afterBlockComment: false,
        },
      ],
    },
    {
      code: "/** \nblock \nblock block\n */\n/* block \n block \n */",
      options: [
        {
          afterBlockComment: true,
          beforeBlockComment: true,
        },
      ],
    },
    {
      code: "bar()\n\n/** block block block\n * block \n */\n\nvar a = 1;",
      options: [
        {
          afterBlockComment: true,
          beforeBlockComment: true,
        },
      ],
    },
    {
      code: "foo() // An inline comment with a an empty line after\nvar a = 1;",
      options: [
        {
          afterLineComment: true,
          beforeLineComment: true,
        },
      ],
    },
    {
      code: "foo();\nbar() /* An inline block comment with a an empty line after\n *\n */\nvar a = 1;",
      options: [
        {
          beforeBlockComment: true,
        },
      ],
    },
    {
      code: "bar()\n\n/** block block block\n * block \n */\n//line line line\nvar a = 1;",
      options: [
        {
          afterBlockComment: true,
        },
      ],
    },
    {
      code: "bar()\n\n/** block block block\n * block \n */\n//line line line\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
        },
      ],
    },
    {
      code: "var a,\n\n// line\nb;",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "function foo(){   \n// line at block start\nvar g = 1;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "function foo(){// line at block start\nvar g = 1;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "var foo = function(){\n// line at block start\nvar g = 1;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "var foo = function(){\n// line at block start\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "if(true){\n// line at block start\nvar g = 1;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "if(true){\n\n// line at block start\nvar g = 1;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "if(true){\n// line at block start\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "if(true){ bar(); } else {\n// line at block start\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\n// line at switch case start\nbreak;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\n\n// line at switch case start\nbreak;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\n// line at switch case start\nbreak;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\n\n// line at switch case start\nbreak;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "function foo(){   \n/* block comment at block start */\nvar g = 1;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "function foo(){/* block comment at block start */\nvar g = 1;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "var foo = function(){\n/* block comment at block start */\nvar g = 1;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "if(true){\n/* block comment at block start */\nvar g = 1;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "if(true){\n\n/* block comment at block start */\nvar g = 1;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "while(true){\n\n/* \nblock comment at block start\n */\nvar g = 1;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "class A {\n/**\n* hi\n */\nconstructor() {}\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "class A {\n/**\n* hi\n */\nconstructor() {}\n}",
      options: [
        {
          allowClassStart: true,
        },
      ],
    },
    {
      code: "class A {\n/**\n* hi\n */\nconstructor() {}\n}",
      options: [
        {
          allowBlockStart: false,
          allowClassStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\n/* block comment at switch case start */\nbreak;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\n\n/* block comment at switch case start */\nbreak;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\n/* block comment at switch case start */\nbreak;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\n\n/* block comment at switch case start */\nbreak;\n}",
      options: [
        {
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        // line comment\n    }\n\n    static {\n        // line comment\n        foo();\n    }\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "class C {\n    static\n    {\n        // line comment\n    }\n\n    static\n    {\n        // line comment\n        foo();\n    }\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        /* block comment */\n    }\n\n    static {\n        /* block\n           comment */\n    }\n\n    static {\n        /* block comment */\n        foo();\n    }\n\n    static {\n        /* block\n           comment */\n        foo();\n    }\n}",
      options: [
        {
          beforeBlockComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "class C {\n    static\n    {\n        /* block comment */\n    }\n\n    static\n    {\n        /* block\n        comment */\n    }\n\n    static\n    {\n        /* block comment */\n        foo();\n    }\n\n    static\n    {\n        /* block\n        comment */\n        foo();\n    }\n}",
      options: [
        {
          beforeBlockComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "switch (foo) {\n// this comment is allowed by allowBlockStart: true \n    \ncase 1:    \n    bar();\n    break;\n    \n// this comment is allowed by allowBlockEnd: true\n}",
      options: [
        {
          allowBlockStart: true,
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch (foo)\n{\n// this comment is allowed by allowBlockStart: true \n    \ncase 1:    \n    bar();\n    break;\n}",
      options: [
        {
          allowBlockStart: true,
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
    },
    {
      code: "switch (\n    function(){}()\n)\n{\n    // this comment is allowed by allowBlockStart: true\n    case foo:\n        break;\n}",
      options: [
        {
          allowBlockStart: true,
          beforeLineComment: true,
        },
      ],
    },
    {
      code: "var a,\n// line\n\nb;",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "function foo(){\nvar g = 91;\n// line at block end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "function foo(){\nvar g = 61;\n\n\n// line at block end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "var foo = function(){\nvar g = 1;\n\n\n// line at block end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "if(true){\nvar g = 1;\n// line at block end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "if(true){\nvar g = 1;\n\n// line at block end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nvar g = 1;\n\n// line at switch case end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nvar g = 1;\n\n// line at switch case end\n\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\nvar g = 1;\n\n// line at switch case end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\nvar g = 1;\n\n// line at switch case end\n\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      options: [
        {
          afterLineComment: true,
          beforeLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
    },
    {
      code: "function foo(){   \nvar g = 1;\n/* block comment at block end */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "function foo(){\nvar g = 1;\n/* block comment at block end */}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "var foo = function(){\nvar g = 1;\n/* block comment at block end */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "if(true){\nvar g = 1;\n/* block comment at block end */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "if(true){\nvar g = 1;\n\n/* block comment at block end */\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "while(true){\n\nvar g = 1;\n\n/* \nblock comment at block end\n */}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "class B {\nconstructor() {}\n\n/**\n* hi\n */\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "class B {\nconstructor() {}\n\n/**\n* hi\n */\n}",
      options: [
        {
          afterBlockComment: true,
          allowClassEnd: true,
        },
      ],
    },
    {
      code: "class B {\nconstructor() {}\n\n/**\n* hi\n */\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: false,
          allowClassEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nvar g = 1;\n\n/* block comment at switch case end */\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nvar g = 1;\n\n/* block comment at switch case end */\n\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\nvar g = 1;\n\n/* block comment at switch case end */\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\nvar g = 1;\n\n/* block comment at switch case end */\n\n}",
      options: [
        {
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        // line comment\n    }\n\n    static {\n        foo();\n        // line comment\n    }\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        /* block comment */\n    }\n\n    static {\n        /* block\n           comment */\n    }\n\n    static {\n        foo();\n        /* block comment */\n    }\n\n    static {\n        foo();\n        /* block\n           comment */\n    }\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowBlockEnd: true,
        },
      ],
    },
    {
      code: "var a,\n\n// line\nb;",
      options: [
        {
          beforeLineComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "var obj = {\n  // line at object start\n  g: 1\n};",
      options: [
        {
          beforeLineComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    // hi\n    test: function() {\n    }\n  }\n}",
      options: [
        {
          beforeLineComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "var obj = {\n  /* block comment at object start*/\n  g: 1\n};",
      options: [
        {
          beforeBlockComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    /**\n    * hi\n    */\n    test: function() {\n    }\n  }\n}",
      options: [
        {
          beforeLineComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "const {\n  // line at object start\n  g: a\n} = {};",
      options: [
        {
          beforeLineComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "const {\n  // line at object start\n  g\n} = {};",
      options: [
        {
          beforeLineComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "const {\n  /* block comment at object-like start*/\n  g: a\n} = {};",
      options: [
        {
          beforeBlockComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "const {\n  /* block comment at object-like start*/\n  g\n} = {};",
      options: [
        {
          beforeBlockComment: true,
          allowObjectStart: true,
        },
      ],
    },
    {
      code: "var a,\n// line\n\nb;",
      options: [
        {
          afterLineComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "var obj = {\n  g: 1\n  // line at object end\n};",
      options: [
        {
          afterLineComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    test: function() {\n    }\n    // hi\n  }\n}",
      options: [
        {
          afterLineComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "var obj = {\n  g: 1\n  \n  /* block comment at object end*/\n};",
      options: [
        {
          afterBlockComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    test: function() {\n    }\n    \n    /**\n    * hi\n    */\n  }\n}",
      options: [
        {
          afterBlockComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "const {\n  g: a\n  // line at object end\n} = {};",
      options: [
        {
          afterLineComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "const {\n  g\n  // line at object end\n} = {};",
      options: [
        {
          afterLineComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "const {\n  g: a\n  \n  /* block comment at object-like end*/\n} = {};",
      options: [
        {
          afterBlockComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "const {\n  g\n  \n  /* block comment at object-like end*/\n} = {};",
      options: [
        {
          afterBlockComment: true,
          allowObjectEnd: true,
        },
      ],
    },
    {
      code: "var a,\n\n// line\nb;",
      options: [
        {
          beforeLineComment: true,
          allowArrayStart: true,
        },
      ],
    },
    {
      code: "var arr = [\n  // line at array start\n  1\n];",
      options: [
        {
          beforeLineComment: true,
          allowArrayStart: true,
        },
      ],
    },
    {
      code: "var arr = [\n  /* block comment at array start*/\n  1\n];",
      options: [
        {
          beforeBlockComment: true,
          allowArrayStart: true,
        },
      ],
    },
    {
      code: "const [\n  // line at array start\n  a\n] = [];",
      options: [
        {
          beforeLineComment: true,
          allowArrayStart: true,
        },
      ],
    },
    {
      code: "const [\n  /* block comment at array start*/\n  a\n] = [];",
      options: [
        {
          beforeBlockComment: true,
          allowArrayStart: true,
        },
      ],
    },
    {
      code: "var a,\n// line\n\nb;",
      options: [
        {
          afterLineComment: true,
          allowArrayEnd: true,
        },
      ],
    },
    {
      code: "var arr = [\n  1\n  // line at array end\n];",
      options: [
        {
          afterLineComment: true,
          allowArrayEnd: true,
        },
      ],
    },
    {
      code: "var arr = [\n  1\n  \n  /* block comment at array end*/\n];",
      options: [
        {
          afterBlockComment: true,
          allowArrayEnd: true,
        },
      ],
    },
    {
      code: "const [\n  a\n  // line at array end\n] = [];",
      options: [
        {
          afterLineComment: true,
          allowArrayEnd: true,
        },
      ],
    },
    {
      code: "const [\n  a\n  \n  /* block comment at array end*/\n] = [];",
      options: [
        {
          afterBlockComment: true,
          allowArrayEnd: true,
        },
      ],
    },
    {
      code: "foo;\n\n/* eslint-disable no-underscore-dangle */\n\nthis._values = values;\nthis._values2 = true;\n/* eslint-enable no-underscore-dangle */\nbar",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
        },
      ],
    },
    "foo;\n/* eslint */",
    "foo;\n/* jshint */",
    "foo;\n/* jslint */",
    "foo;\n/* istanbul */",
    "foo;\n/* global */",
    "foo;\n/* globals */",
    "foo;\n/* exported */",
    "foo;\n/* jscs */",
    {
      code: "foo\n/* this is pragmatic */",
      options: [
        {
          ignorePattern: "pragma",
        },
      ],
    },
    {
      code: "foo\n/* this is pragmatic */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
          ignorePattern: "pragma",
        },
      ],
    },
    {
      code: "#!comment\n\nvar a = 1;",
      options: [
        {
          afterHashbangComment: true,
        },
      ],
    },
    "#!comment\nvar a = 1;",
    {
      code: "#!comment\nvar a = 1;",
      options: [
        {},
      ],
    },
    {
      code: "#!comment\nvar a = 1;",
      options: [
        {
          afterHashbangComment: false,
        },
      ],
    },
    {
      code: "#!comment\nvar a = 1;",
      options: [
        {
          afterLineComment: true,
          afterBlockComment: true,
        },
      ],
    },

    // ---- from lines-around-comment._ts_.test.ts ----
    {
      code: "interface A {\n  // line\n  a: string;\n}",
      options: [
        {
          beforeLineComment: true,
          allowInterfaceStart: true,
        },
      ],
    },
    {
      code: "interface A {\n  /* block\n    comment */\n  a: string;\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceStart: true,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  // line\n}",
      options: [
        {
          afterLineComment: true,
          allowInterfaceEnd: true,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  /* block\n    comment */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowInterfaceEnd: true,
        },
      ],
    },
    {
      code: "type A = {\n  // line\n  a: string;\n}",
      options: [
        {
          beforeLineComment: true,
          allowTypeStart: true,
        },
      ],
    },
    {
      code: "type A = {\n  /* block\n    comment */\n  a: string;\n}",
      options: [
        {
          beforeBlockComment: true,
          allowTypeStart: true,
        },
      ],
    },
    {
      code: "type A = {\n  a: string;\n  // line\n}",
      options: [
        {
          afterLineComment: true,
          allowTypeEnd: true,
        },
      ],
    },
    {
      code: "type A = {\n  a: string;\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowTypeEnd: true,
        },
      ],
    },
    {
      code: "enum A {\n  // line\n  a,\n}",
      options: [
        {
          beforeLineComment: true,
          allowEnumStart: true,
        },
      ],
    },
    {
      code: "enum A {\n  /* block\n     comment */\n  a,\n}",
      options: [
        {
          beforeBlockComment: true,
          allowEnumStart: true,
        },
      ],
    },
    {
      code: "enum A {\n  a,\n  // line\n}",
      options: [
        {
          afterLineComment: true,
          allowEnumEnd: true,
        },
      ],
    },
    {
      code: "enum A {\n  a,\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowEnumEnd: true,
        },
      ],
    },
    {
      code: "declare module A {\n  // line\n  const a: string;\n}",
      options: [
        {
          beforeLineComment: true,
          allowModuleStart: true,
        },
      ],
    },
    {
      code: "declare module A {\n  /* block\n     comment */\n  const a: string;\n}",
      options: [
        {
          beforeBlockComment: true,
          allowModuleStart: true,
        },
      ],
    },
    {
      code: "declare module A {\n  const a: string;\n  // line\n}",
      options: [
        {
          afterLineComment: true,
          allowModuleEnd: true,
        },
      ],
    },
    {
      code: "declare module A {\n  const a: string;\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowModuleEnd: true,
        },
      ],
    },
    {
      code: "interface A {foo: string;\n\n/* eslint-disable no-underscore-dangle */\n\n_values: 2;\n_values2: true;\n/* eslint-enable no-underscore-dangle */\nbar: boolean}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
        },
      ],
    },
    "\ninterface A {\n  foo;\n  /* eslint */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* jshint */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* jslint */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* istanbul */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* global */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* globals */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* exported */\n}\n    ",
    "\ninterface A {\n  foo;\n  /* jscs */\n}\n    ",
    {
      code: "interface A {\n  foo: boolean;\n  /* this is pragmatic */\n}",
      options: [
        {
          ignorePattern: "pragma",
        },
      ],
    },
    {
      code: "interface A {\n  foo;\n  /* this is pragmatic */\n}",
      options: [
        {
          applyDefaultIgnorePatterns: false,
          ignorePattern: "pragma",
        },
      ],
    },
    {
      code: "interface A {\n  foo: string; // this is inline line comment\n}",
      options: [
        {
          beforeLineComment: true,
        },
      ],
    },
    {
      code: "interface A {\n  foo: string /* this is inline block comment */;\n}",
    },
    {
      code: "interface A {\n  /* this is inline block comment */ foo: string;\n}",
    },
    {
      code: "interface A {\n  /* this is inline block comment */ foo: string /* this is inline block comment */;\n}",
    },
    {
      code: "interface A {\n  /* this is inline block comment */ foo: string; // this is inline line comment ;\n}",
    },
  ],

  invalid: [
    // ---- from lines-around-comment._js_.test.ts ----
    {
      code: "bar()\n/** block block block\n * block \n */\nvar a = 1;",
      output: "bar()\n\n/** block block block\n * block \n */\nvar a = 1;",
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "baz()\n// A line comment with no empty line after\nvar a = 1;",
      output: "baz()\n// A line comment with no empty line after\n\nvar a = 1;",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
        },
      ],
    },
    {
      code: "baz()\n// A line comment with no empty line after\nvar a = 1;",
      output: "baz()\n\n// A line comment with no empty line after\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "// A line comment with no empty line after\nvar a = 1;",
      output: "// A line comment with no empty line after\n\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: "baz()\n// A line comment with no empty line after\nvar a = 1;",
      output: "baz()\n\n// A line comment with no empty line after\n\nvar a = 1;",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "bar()\n/**\n * block block block\n */\nvar a = 1;",
      output: "bar()\n\n/**\n * block block block\n */\n\nvar a = 1;",
      options: [
        {
          afterBlockComment: true,
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "bar()\n/* first block comment */ /* second block comment */\nvar a = 1;",
      output: "bar()\n\n/* first block comment */ /* second block comment */\n\nvar a = 1;",
      options: [
        {
          afterBlockComment: true,
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "bar()\n/* first block comment */ /* second block\n comment */\nvar a = 1;",
      output: "bar()\n\n/* first block comment */ /* second block\n comment */\n\nvar a = 1;",
      options: [
        {
          afterBlockComment: true,
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "bar()\n/**\n * block block block\n */\nvar a = 1;",
      output: "bar()\n/**\n * block block block\n */\n\nvar a = 1;",
      options: [
        {
          afterBlockComment: true,
          beforeBlockComment: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "bar()\n/**\n * block block block\n */\nvar a = 1;",
      output: "bar()\n\n/**\n * block block block\n */\nvar a = 1;",
      options: [
        {
          afterBlockComment: false,
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "var a,\n// line\nb;",
      output: "var a,\n\n// line\nb;",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "function foo(){\nvar a = 1;\n// line at block start\nvar g = 1;\n}",
      output: "function foo(){\nvar a = 1;\n\n// line at block start\nvar g = 1;\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "var a,\n// line\nb;",
      output: "var a,\n// line\n\nb;",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "function foo(){\nvar a = 1;\n\n// line at block start\nvar g = 1;\n}",
      output: "function foo(){\nvar a = 1;\n\n// line at block start\n\nvar g = 1;\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\n// line at switch case start\nbreak;\n}",
      output: "switch ('foo'){\ncase 'foo':\n\n// line at switch case start\nbreak;\n}",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\n// line at switch case start\nbreak;\n}",
      output: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\n\n// line at switch case start\nbreak;\n}",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 6,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      output: "while(true){\n// line at block start and end\n\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockStart: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "while(true){\n// line at block start and end\n}",
      output: "while(true){\n\n// line at block start and end\n}",
      options: [
        {
          beforeLineComment: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "class A {\n// line at class start\nconstructor() {}\n}",
      output: "class A {\n\n// line at class start\nconstructor() {}\n}",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "class A {\n// line at class start\nconstructor() {}\n}",
      output: "class A {\n\n// line at class start\nconstructor() {}\n}",
      options: [
        {
          allowBlockStart: true,
          allowClassStart: false,
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "class B {\nconstructor() {}\n\n// line at class end\n}",
      output: "class B {\nconstructor() {}\n\n// line at class end\n\n}",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "class B {\nconstructor() {}\n\n// line at class end\n}",
      output: "class B {\nconstructor() {}\n\n// line at class end\n\n}",
      options: [
        {
          afterLineComment: true,
          allowBlockEnd: true,
          allowClassEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nvar g = 1;\n\n// line at switch case end\n}",
      output: "switch ('foo'){\ncase 'foo':\nvar g = 1;\n\n// line at switch case end\n\n}",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 5,
        },
      ],
    },
    {
      code: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\nvar g = 1;\n\n// line at switch case end\n}",
      output: "switch ('foo'){\ncase 'foo':\nbreak;\n\ndefault:\nvar g = 1;\n\n// line at switch case end\n\n}",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 8,
        },
      ],
    },
    {
      code: "class C {\n    // line comment\n    static{}\n}",
      output: "class C {\n    // line comment\n\n    static{}\n}",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
          allowClassStart: true,
          allowClassEnd: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "class C {\n    /* block\n       comment */\n    static{}\n}",
      output: "class C {\n    /* block\n       comment */\n\n    static{}\n}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
          allowClassStart: true,
          allowClassEnd: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "class C {\n    static\n    // line comment\n    {}\n}",
      output: "class C {\n    static\n\n    // line comment\n\n    {}\n}",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
          allowClassStart: true,
          allowClassEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "class C {\n    static\n    /* block\n       comment */\n    {}\n}",
      output: "class C {\n    static\n\n    /* block\n       comment */\n\n    {}\n}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
          allowClassStart: true,
          allowClassEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        // line comment\n        foo();\n    }\n}",
      output: "class C {\n    static {\n        // line comment\n\n        foo();\n    }\n}",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        /* block\n           comment */\n        foo();\n    }\n}",
      output: "class C {\n    static {\n        /* block\n           comment */\n\n        foo();\n    }\n}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        foo();\n        // line comment\n    }\n}",
      output: "class C {\n    static {\n        foo();\n\n        // line comment\n    }\n}",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 4,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        foo();\n        /* block\n           comment */\n    }\n}",
      output: "class C {\n    static {\n        foo();\n\n        /* block\n           comment */\n    }\n}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 4,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        foo();\n        // line comment\n        bar();\n    }\n}",
      output: "class C {\n    static {\n        foo();\n\n        // line comment\n\n        bar();\n    }\n}",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 4,
        },
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "class C {\n    static {\n        foo();\n        /* block\n           comment */\n        bar();\n    }\n}",
      output: "class C {\n    static {\n        foo();\n\n        /* block\n           comment */\n\n        bar();\n    }\n}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 4,
        },
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "class C {\n    static{}\n    // line comment\n}",
      output: "class C {\n    static{}\n\n    // line comment\n}",
      options: [
        {
          beforeLineComment: true,
          afterLineComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
          allowClassStart: true,
          allowClassEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "class C {\n    static{}\n    /* block\n       comment */\n}",
      output: "class C {\n    static{}\n\n    /* block\n       comment */\n}",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          allowBlockStart: true,
          allowBlockEnd: true,
          allowClassStart: true,
          allowClassEnd: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "var obj = {\n  // line at object start\n  g: 1\n};",
      output: "var obj = {\n\n  // line at object start\n  g: 1\n};",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    // hi\n    test: function() {\n    }\n  }\n}",
      output: "function hi() {\n  return {\n\n    // hi\n    test: function() {\n    }\n  }\n}",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "var obj = {\n  /* block comment at object start*/\n  g: 1\n};",
      output: "var obj = {\n\n  /* block comment at object start*/\n  g: 1\n};",
      options: [
        {
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    /**\n    * hi\n    */\n    test: function() {\n    }\n  }\n}",
      output: "function hi() {\n  return {\n\n    /**\n    * hi\n    */\n    test: function() {\n    }\n  }\n}",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "const {\n  // line at object start\n  g: a\n} = {};",
      output: "const {\n\n  // line at object start\n  g: a\n} = {};",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "const {\n  // line at object start\n  g\n} = {};",
      output: "const {\n\n  // line at object start\n  g\n} = {};",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "const {\n  /* block comment at object-like start*/\n  g: a\n} = {};",
      output: "const {\n\n  /* block comment at object-like start*/\n  g: a\n} = {};",
      options: [
        {
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "const {\n  /* block comment at object-like start*/\n  g\n} = {};",
      output: "const {\n\n  /* block comment at object-like start*/\n  g\n} = {};",
      options: [
        {
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "var obj = {\n  g: 1\n  // line at object end\n};",
      output: "var obj = {\n  g: 1\n  // line at object end\n\n};",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    test: function() {\n    }\n    // hi\n  }\n}",
      output: "function hi() {\n  return {\n    test: function() {\n    }\n    // hi\n\n  }\n}",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 5,
        },
      ],
    },
    {
      code: "var obj = {\n  g: 1\n  \n  /* block comment at object end*/\n};",
      output: "var obj = {\n  g: 1\n  \n  /* block comment at object end*/\n\n};",
      options: [
        {
          afterBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "function hi() {\n  return {\n    test: function() {\n    }\n    \n    /**\n    * hi\n    */\n  }\n}",
      output: "function hi() {\n  return {\n    test: function() {\n    }\n    \n    /**\n    * hi\n    */\n\n  }\n}",
      options: [
        {
          afterBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 6,
        },
      ],
    },
    {
      code: "const {\n  g: a\n  // line at object end\n} = {};",
      output: "const {\n  g: a\n  // line at object end\n\n} = {};",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "const {\n  g\n  // line at object end\n} = {};",
      output: "const {\n  g\n  // line at object end\n\n} = {};",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "const {\n  g: a\n  \n  /* block comment at object-like end*/\n} = {};",
      output: "const {\n  g: a\n  \n  /* block comment at object-like end*/\n\n} = {};",
      options: [
        {
          afterBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "const {\n  g\n  \n  /* block comment at object-like end*/\n} = {};",
      output: "const {\n  g\n  \n  /* block comment at object-like end*/\n\n} = {};",
      options: [
        {
          afterBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "var arr = [\n  // line at array start\n  1\n];",
      output: "var arr = [\n\n  // line at array start\n  1\n];",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "var arr = [\n  /* block comment at array start*/\n  1\n];",
      output: "var arr = [\n\n  /* block comment at array start*/\n  1\n];",
      options: [
        {
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "const [\n  // line at array start\n  a\n] = [];",
      output: "const [\n\n  // line at array start\n  a\n] = [];",
      options: [
        {
          beforeLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "const [\n  /* block comment at array start*/\n  a\n] = [];",
      output: "const [\n\n  /* block comment at array start*/\n  a\n] = [];",
      options: [
        {
          beforeBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "var arr = [\n  1\n  // line at array end\n];",
      output: "var arr = [\n  1\n  // line at array end\n\n];",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "var arr = [\n  1\n  \n  /* block comment at array end*/\n];",
      output: "var arr = [\n  1\n  \n  /* block comment at array end*/\n\n];",
      options: [
        {
          afterBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "const [\n  a\n  // line at array end\n] = [];",
      output: "const [\n  a\n  // line at array end\n\n] = [];",
      options: [
        {
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "const [\n  a\n  \n  /* block comment at array end*/\n] = [];",
      output: "const [\n  a\n  \n  /* block comment at array end*/\n\n] = [];",
      options: [
        {
          afterBlockComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 4,
        },
      ],
    },
    {
      code: "foo;\n\n/* eslint-disable no-underscore-dangle */\n\nthis._values = values;\nthis._values2 = true;\n/* eslint-enable no-underscore-dangle */\nbar",
      output: "foo;\n\n/* eslint-disable no-underscore-dangle */\n\nthis._values = values;\nthis._values2 = true;\n\n/* eslint-enable no-underscore-dangle */\n\nbar",
      options: [
        {
          beforeBlockComment: true,
          afterBlockComment: true,
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 7,
        },
        {
          messageId: "after",
          line: 7,
        },
      ],
    },
    {
      code: "foo;\n/* eslint */",
      output: "foo;\n\n/* eslint */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* jshint */",
      output: "foo;\n\n/* jshint */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* jslint */",
      output: "foo;\n\n/* jslint */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* istanbul */",
      output: "foo;\n\n/* istanbul */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* global */",
      output: "foo;\n\n/* global */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* globals */",
      output: "foo;\n\n/* globals */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* exported */",
      output: "foo;\n\n/* exported */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* jscs */",
      output: "foo;\n\n/* jscs */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo\n/* something else */",
      output: "foo\n\n/* something else */",
      options: [
        {
          ignorePattern: "pragma",
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo\n/* eslint */",
      output: "foo\n\n/* eslint */",
      options: [
        {
          applyDefaultIgnorePatterns: false,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "foo;\n/* fallthrough */",
      output: "foo;\n\n/* fallthrough */",
      options: [],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "switch (\n// this comment is not allowed by allowBlockStart: true\n\n    foo\n)\n{   \ncase 1:    \n    bar();\n    break;\n}",
      output: "switch (\n\n// this comment is not allowed by allowBlockStart: true\n\n    foo\n)\n{   \ncase 1:    \n    bar();\n    break;\n}",
      options: [
        {
          allowBlockStart: true,
          beforeLineComment: true,
          afterLineComment: true,
        },
      ],
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "#!foo\nvar a = 1;",
      output: "#!foo\n\nvar a = 1;",
      options: [
        {
          afterHashbangComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
        },
      ],
    },

    // ---- from lines-around-comment._ts_.test.ts ----
    {
      code: "bar();\n/** block block block\n * block\n */\nvar a = 1;",
      output: "bar();\n\n/** block block block\n * block\n */\nvar a = 1;",
      errors: [
        {
          messageId: "before",
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  // line\n}",
      output: "interface A {\n  a: string;\n\n  // line\n}",
      options: [
        {
          beforeLineComment: true,
          allowInterfaceStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  /* block\n     comment */\n}",
      output: "interface A {\n  a: string;\n\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "interface A {\n  // line\n  a: string;\n}",
      output: "interface A {\n\n  // line\n  a: string;\n}",
      options: [
        {
          beforeLineComment: true,
          allowInterfaceStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "interface A {\n  /* block\n     comment */\n  a: string;\n}",
      output: "interface A {\n\n  /* block\n     comment */\n  a: string;\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  // line\n}",
      output: "interface A {\n  a: string;\n  // line\n\n}",
      options: [
        {
          afterLineComment: true,
          allowInterfaceEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  /* block\n     comment */\n}",
      output: "interface A {\n  a: string;\n  /* block\n     comment */\n\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowInterfaceEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "type A = {\n  a: string;\n  // line\n}",
      output: "type A = {\n  a: string;\n\n  // line\n}",
      options: [
        {
          beforeLineComment: true,
          allowInterfaceStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "type A = {\n  a: string;\n  /* block\n     comment */\n}",
      output: "type A = {\n  a: string;\n\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "type A = {\n  // line\n  a: string;\n}",
      output: "type A = {\n\n  // line\n  a: string;\n}",
      options: [
        {
          beforeLineComment: true,
          allowInterfaceStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "type A = {\n  /* block\n     comment */\n  a: string;\n}",
      output: "type A = {\n\n  /* block\n     comment */\n  a: string;\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "type A = {\n  a: string;\n  // line\n}",
      output: "type A = {\n  a: string;\n  // line\n\n}",
      options: [
        {
          afterLineComment: true,
          allowInterfaceEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "type A = {\n  a: string;\n  /* block\n     comment */\n}",
      output: "type A = {\n  a: string;\n  /* block\n     comment */\n\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowInterfaceEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "enum A {\n  a,\n  // line\n}",
      output: "enum A {\n  a,\n\n  // line\n}",
      options: [
        {
          beforeLineComment: true,
          allowEnumStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "enum A {\n  a,\n  /* block\n     comment */\n}",
      output: "enum A {\n  a,\n\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: true,
          allowEnumStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "enum A {\n  // line\n  a,\n}",
      output: "enum A {\n\n  // line\n  a,\n}",
      options: [
        {
          beforeLineComment: true,
          allowEnumStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "enum A {\n  /* block\n     comment */\n  a,\n}",
      output: "enum A {\n\n  /* block\n     comment */\n  a,\n}",
      options: [
        {
          beforeBlockComment: true,
          allowEnumStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "enum A {\n  a,\n  // line\n}",
      output: "enum A {\n  a,\n  // line\n\n}",
      options: [
        {
          afterLineComment: true,
          allowEnumEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "enum A {\n  a,\n  /* block\n     comment */\n}",
      output: "enum A {\n  a,\n  /* block\n     comment */\n\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowEnumEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "module A {\n  const a: string;\n  // line\n}",
      output: "module A {\n  const a: string;\n\n  // line\n}",
      options: [
        {
          beforeLineComment: true,
          allowModuleStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "module A {\n  const a: string;\n  /* block\n     comment */\n}",
      output: "module A {\n  const a: string;\n\n  /* block\n     comment */\n}",
      options: [
        {
          beforeBlockComment: true,
          allowModuleStart: true,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "module A {\n  // line\n  const a: string;\n}",
      output: "module A {\n\n  // line\n  const a: string;\n}",
      options: [
        {
          beforeLineComment: true,
          allowModuleStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "module A {\n  /* block\n     comment */\n  const a: string;\n}",
      output: "module A {\n\n  /* block\n     comment */\n  const a: string;\n}",
      options: [
        {
          beforeBlockComment: true,
          allowModuleStart: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 2,
        },
      ],
    },
    {
      code: "module A {\n  const a: string;\n  // line\n}",
      output: "module A {\n  const a: string;\n  // line\n\n}",
      options: [
        {
          afterLineComment: true,
          allowModuleEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "module A {\n  const a: string;\n  /* block\n     comment */\n}",
      output: "module A {\n  const a: string;\n  /* block\n     comment */\n\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowModuleEnd: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 3,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  /* block */ /* block */\n}",
      output: "interface A {\n  a: string;\n\n  /* block */ /* block */\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceEnd: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "interface A {\n  a: string;\n  /* block */ // line\n}",
      output: "interface A {\n  a: string;\n\n  /* block */ // line\n}",
      options: [
        {
          beforeBlockComment: true,
          allowInterfaceEnd: false,
        },
      ],
      errors: [
        {
          messageId: "before",
          line: 3,
        },
      ],
    },
    {
      code: "interface A {\n  /* block */ /* block */\n  a: string;\n}",
      output: "interface A {\n  /* block */ /* block */\n\n  a: string;\n}",
      options: [
        {
          beforeBlockComment: false,
          afterBlockComment: true,
          allowInterfaceStart: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "interface A {\n  /* block */ // line\n  a: string;\n}",
      output: "interface A {\n  /* block */ // line\n\n  a: string;\n}",
      options: [
        {
          beforeBlockComment: false,
          afterLineComment: true,
          allowInterfaceStart: false,
        },
      ],
      errors: [
        {
          messageId: "after",
          line: 2,
        },
      ],
    },
    {
      code: "#!foo\nvar a = 1;",
      output: "#!foo\n\nvar a = 1;",
      options: [
        {
          afterHashbangComment: true,
        },
      ],
      errors: [
        {
          messageId: "after",
        },
      ],
    },
  ],
});
