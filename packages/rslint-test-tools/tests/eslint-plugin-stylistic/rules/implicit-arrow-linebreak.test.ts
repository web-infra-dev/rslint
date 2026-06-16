/**
 * @fileoverview enforce the location of arrow function bodies
 * @author Sharmila Jesupaul
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/implicit-arrow-linebreak/implicit-arrow-linebreak.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('implicit-arrow-linebreak', null as never, { valid, invalid })`.
 *  - The `$` unindent template tag (`unindent` from eslint-vitest-rule-tester /
 *    @antfu/utils) is evaluated to its real multi-line string: leading/trailing
 *    fully-whitespace lines are dropped and the common leading indentation is
 *    stripped. Trailing whitespace inside a line is preserved.
 *  - The many PLAIN backtick (non-`$`) multi-line strings in the upstream `valid`
 *    array are kept VERBATIM — their literal indentation and leading/trailing
 *    newlines are preserved exactly as written (no dedent).
 *  - The local error helper constants are inlined to their final `{ messageId }`:
 *      EXPECTED_LINEBREAK   -> { messageId: 'expected' }
 *      UNEXPECTED_LINEBREAK -> { messageId: 'unexpected' }
 *  - `parserOptions` (ecmaVersion 8) dropped — rslint parses every fixture with
 *    ts-go via the generated tsconfig; async arrows are valid TS there.
 *
 * The upstream file is a single `run()` block — there is no second `if (!skipBabel)`
 * block and no Babel/Flow cases. The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule. No `suggestions` and no output-only invalid cases exist.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `KNOWN GAPS` block at the bottom, each annotated with what
 * upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('implicit-arrow-linebreak', null as never, {
  valid: [
    "(foo) => {\n        bar\n    }",
    "() => bar;",
    "() => (bar);",
    "() => bar => baz;",
    "() => ((((bar))));",
    "(foo) => (\n        bar\n    )",
    "(foo) => bar();",
    "\n    //comment\n    foo => bar;\n    ",
    "\n    foo => (\n        // comment\n        bar => (\n            // another comment\n            baz\n        )\n    )\n    ",
    "\n    foo => (\n        // comment\n        bar => baz\n    )\n    ",
    "\n    /* text */\n    () => bar;\n    ",
    "\n    /* foo */\n    const bar = () => baz;\n    ",
    "\n    (foo) => (\n            //comment\n                bar\n            )\n    ",
    "\n      [ // comment\n        foo => 'bar'\n      ]\n    ",
    "\n      /**\n     One two three four\n      Five six seven nine.\n      */\n      (foo) => bar\n    ",
    "\n    const foo = {\n      id: 'bar',\n      // comment\n      prop: (foo1) => 'returning this string',\n    }\n    ",
    "\n    // comment\n      \"foo\".split('').map((char) => char\n      )\n    ",
    { code: "async foo => () => bar;" },
    { code: "// comment\nasync foo => 'string'" },
    { code: "(foo) =>\n    (\n        bar\n    )", options: ["below"] },
    { code: "() =>\n    ((((bar))));", options: ["below"] },
    { code: "() =>\n    bar();", options: ["below"] },
    { code: "() =>\n    (bar);", options: ["below"] },
    { code: "() =>\n    bar =>\n        baz;", options: ["below"] },
  ],

  invalid: [
    { code: "(foo) =>\n    bar();", output: "(foo) => bar();", errors: [{ messageId: "unexpected" }] },
    { code: "() =>\n    (bar);", output: "() => (bar);", errors: [{ messageId: "unexpected" }] },
    { code: "() =>\n    bar =>\n        baz;", output: "() => bar => baz;", errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }] },
    { code: "() =>\n    ((((bar))));", output: "() => ((((bar))));", errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n    (\n        bar\n    )", output: "(foo) => (\n        bar\n    )", errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n  // test comment\n  bar", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "const foo = () =>\n// comment\n[]", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n    (\n    //comment\n        bar\n    )", output: "(foo) => (\n    //comment\n        bar\n    )", errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n    (\n        bar\n    //comment\n    )", output: "(foo) => (\n        bar\n    //comment\n    )", errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n // comment\n // another comment\n    bar", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n// comment\n(\n// another comment\nbar\n)", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "() => // comment \n bar", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "(foo) => //comment \n bar", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n  /* test comment */\n  bar", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "(foo) =>\n  // hi\n     bar =>\n       // there\n         baz;", output: null, errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }] },
    { code: "(foo) =>\n  // hi\n     bar => (\n       // there\n         baz\n     )", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "const foo = {\n  id: 'bar',\n  prop: (foo1) =>\n    // comment\n    'returning this string',\n}", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "[ foo =>\n  // comment\n  'bar'\n]", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "\"foo\".split('').map((char) =>\n// comment\nchar\n)", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "new Promise((resolve, reject) =>\n    // comment\n    resolve()\n)", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "() =>\n/**\nsuccinct\nexplanation\nof code\n*/\nbar", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "stepOne =>\n    /**\n    here is\n    what is\n    happening\n    */\n    stepTwo =>\n        // then this happens\n        stepThree", output: null, errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }] },
    { code: "() =>\n    /**\n    multi\n    line\n    */\n    bar =>\n        /**\n        many\n        lines\n        */\n        baz", output: null, errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }] },
    { code: "foo('', boo =>\n  // comment\n  bar\n)", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "async foo =>\n    // comment\n    'string'", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "async foo =>\n    // comment\n    // another\n    bar;", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "async (foo) =>\n    // comment\n    'string'", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "const foo = 1,\n  bar = 2,\n  baz = () => // comment\n    qux", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "const foo = () =>\n  //comment\n  qux,\n  bar = 2,\n  baz = 3", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "const foo = () =>\n    //two\n    1,\n    boo = () =>\n    //comment\n    2,\n    bop = \"what\"", output: null, errors: [{ messageId: "unexpected" }, { messageId: "unexpected" }] },
    { code: "start()\n    .then(() =>\n        /* If I put a comment here, eslint --fix breaks badly */\n        process && typeof process.send === 'function' && process.send('ready')\n    )\n    .catch(err => {\n    /* catch seems to be needed here */\n    console.log('Error: ', err)\n    })", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "hello(response =>\n    // comment\n    response, param => param)", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "start(\n    arr =>\n        // cometh\n        bod => {\n            // soon\n            yyyy\n        }\n)", output: null, errors: [{ messageId: "unexpected" }] },
    { code: "(foo) => bar();", options: ["below"], output: "(foo) => \nbar();", errors: [{ messageId: "expected" }] },
    { code: "(foo) => bar => baz;", options: ["below"], output: "(foo) => \nbar => \nbaz;", errors: [{ messageId: "expected" }, { messageId: "expected" }] },
    { code: "(foo) => (bar);", options: ["below"], output: "(foo) => \n(bar);", errors: [{ messageId: "expected" }] },
    { code: "(foo) => (((bar)));", options: ["below"], output: "(foo) => \n(((bar)));", errors: [{ messageId: "expected" }] },
    { code: "(foo) => (\n    bar\n)", options: ["below"], output: "(foo) => \n(\n    bar\n)", errors: [{ messageId: "expected" }] },
  ],
});
