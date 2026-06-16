/**
 * @fileoverview Tests for one-true-brace rule.
 * @author Ian Christian Myers
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/brace-style/brace-style._js_.test.ts
 *   packages/eslint-plugin/rules/brace-style/brace-style._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('brace-style', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - `parserOptions` (ecmaVersion) dropped — rslint always parses at esnext.
 *  - The rule's messageIds (`nextLineOpen` / `sameLineOpen` / `blockSameLine` /
 *    `nextLineClose` / `singleLineClose` / `sameLineClose`) take no `data`, so they
 *    map 1:1 to a fixed message; the RuleTester renders them from the plugin's meta.
 *
 * KNOWN GAPS: none. Every upstream case — including the `with (foo) { ... }`,
 * string-named ambient `module "Foo" { ... }`, and `namespace Foo { ... }` TS
 * cases — parses under rslint's ts-go parser and produces byte-identical
 * diagnostics and autofix output. (`with` is not rejected here: the generated
 * tsconfig does not enable `strict`, so ts-go parses the statement.) No
 * Babel/Flow cases, no octal/`\8` syntax, no `assert` import attributes, and no
 * output-only invalid cases exist for this rule, so nothing was isolated.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('brace-style', null as never, {
  valid: [
    // ---- from brace-style._js_.test.ts ----
    "function f() {\n    if (true)\n        return {x: 1}\n    else {\n        var y = 2\n        return y\n    }\n}",
    "if (tag === 1) glyph.id = pbf.readVarint();\nelse if (tag === 2) glyph.bitmap = pbf.readBytes();",
    "function foo () {\n  return;\n}",
    "function a(b,\nc,\nd) { }",
    "!function foo () {\n  return;\n}",
    "!function a(b,\nc,\nd) { }",
    "if (foo) {\n  bar();\n}",
    "if (a) {\n  b();\n} else {\n  c();\n}",
    "while (foo) {\n  bar();\n}",
    "for (;;) {\n  bar();\n}",
    "switch (foo) {\n  case 'bar': break;\n}",
    "try {\n  bar();\n} catch (e) {\n  baz();\n}",
    "do {\n  bar();\n} while (true)",
    "for (foo in bar) {\n  baz();\n}",
    "if (a &&\n  b &&\n  c) {\n}",
    "switch(0) {\n}",
    "class Foo {\n}",
    "(class {\n})",
    "class\nFoo {\n}",
    "class Foo {\n    bar() {\n    }\n}",
    {
      code: "if (foo) {\n}\nelse {\n}",
      options: ["stroustrup"],
    },
    {
      code: "if (foo)\n{\n}\nelse\n{\n}",
      options: ["allman"],
    },
    {
      code: "try {\n  bar();\n}\ncatch (e) {\n  baz();\n}",
      options: ["stroustrup"],
    },
    {
      code: "try\n{\n  bar();\n}\ncatch (e)\n{\n  baz();\n}",
      options: ["allman"],
    },
    {
      code: "function foo () { return; }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "function foo () { a(); b(); return; }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "function a(b,c,d) { }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "!function foo () { return; }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "!function a(b,c,d) { }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (foo) {  bar(); }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (a) { b(); } else { c(); }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "while (foo) {  bar(); }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "for (;;) {  bar(); }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "switch (foo) {  case \"bar\": break; }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "try {  bar(); } catch (e) { baz();  }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "do {  bar(); } while (true)",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "for (foo in bar) {  baz();  }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (a && b && c) {  }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "switch(0) {}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (foo) {}\nelse {}",
      options: ["stroustrup",{"allowSingleLine":true}],
    },
    {
      code: "try {  bar(); }\ncatch (e) { baz();  }",
      options: ["stroustrup",{"allowSingleLine":true}],
    },
    {
      code: "var foo = () => { return; }",
      options: ["stroustrup",{"allowSingleLine":true}],
    },
    {
      code: "if (foo) {}\nelse {}",
      options: ["allman",{"allowSingleLine":true}],
    },
    {
      code: "try {  bar(); }\ncatch (e) { baz();  }",
      options: ["allman",{"allowSingleLine":true}],
    },
    {
      code: "var foo = () => { return; }",
      options: ["allman",{"allowSingleLine":true}],
    },
    {
      code: "if (foo) { baz(); } else {\n  boom();\n}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (foo) { baz(); } else if (bar) {\n  boom();\n}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (foo) { baz(); } else\nif (bar) {\n  boom();\n}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "try { somethingRisky(); } catch(e) {\n  handleError();\n}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "if (tag === 1) fontstack.name = pbf.readString();\nelse if (tag === 2) fontstack.range = pbf.readString();\nelse if (tag === 3) {\n  var glyph = pbf.readMessage(readGlyph, {});\n  fontstack.glyphs[glyph.id] = glyph;\n}",
      options: ["1tbs"],
    },
    {
      code: "if (tag === 1) fontstack.name = pbf.readString();\nelse if (tag === 2) fontstack.range = pbf.readString();\nelse if (tag === 3) {\n  var glyph = pbf.readMessage(readGlyph, {});\n  fontstack.glyphs[glyph.id] = glyph;\n}",
      options: ["stroustrup"],
    },
    {
      code: "switch(x)\n{\n  case 1:\n    bar();\n}",
      options: ["allman"],
    },
    {
      code: "switch(x) {}",
      options: ["allman",{"allowSingleLine":true}],
    },
    {
      code: "class Foo {\n}",
      options: ["stroustrup"],
    },
    {
      code: "(class {\n})",
      options: ["stroustrup"],
    },
    {
      code: "class Foo\n{\n}",
      options: ["allman"],
    },
    {
      code: "(class\n{\n})",
      options: ["allman"],
    },
    {
      code: "class\nFoo\n{\n}",
      options: ["allman"],
    },
    {
      code: "class Foo {}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "class Foo {}",
      options: ["allman",{"allowSingleLine":true}],
    },
    {
      code: "(class {})",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "(class {})",
      options: ["allman",{"allowSingleLine":true}],
    },
    "{}",
    "if (foo) {\n}\n{\n}",
    "switch (foo) {\n  case bar:\n    baz();\n    {\n      qux();\n    }\n}",
    "{\n}",
    "{\n  {\n  }\n}",
    "class Ball {\n  throw() {}\n  catch() {}\n}",
    "({\n  and() {},\n  finally() {}\n})",
    "(class {\n  or() {}\n  else() {}\n})",
    "if (foo) bar = function() {}\nelse baz()",
    {
      code: "class C {\n    static {\n        foo;\n    }\n}",
      options: ["1tbs"],
    },
    {
      code: "class C {\n    static {}\n\n    static {\n    }\n}",
      options: ["1tbs"],
    },
    {
      code: "class C {\n    static { foo; }\n}",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "class C {\n    static {\n        foo;\n    }\n}",
      options: ["stroustrup"],
    },
    {
      code: "class C {\n    static {}\n\n    static {\n    }\n}",
      options: ["stroustrup"],
    },
    {
      code: "class C {\n    static { foo; }\n}",
      options: ["stroustrup",{"allowSingleLine":true}],
    },
    {
      code: "class C\n{\n    static\n    {\n        foo;\n    }\n}",
      options: ["allman"],
    },
    {
      code: "class C\n{\n    static\n    {}\n}",
      options: ["allman"],
    },
    {
      code: "class C\n{\n    static {}\n\n    static { foo; }\n\n    static\n    { foo; }\n}",
      options: ["allman",{"allowSingleLine":true}],
    },
    {
      code: "class C {\n    static {\n        {\n            foo;\n        }\n    }\n}",
      options: ["1tbs"],
    },

    // ---- from brace-style._ts_.test.ts ----
    "with (foo) {\n  bar();\n}",
    {
      code: "with (foo) {  bar(); }",
      options: ["1tbs",{"allowSingleLine":true}],
    },
    {
      code: "module \"Foo\" {\n}",
      options: ["1tbs"],
    },
    {
      code: "module \"Foo\" {\n}",
      options: ["stroustrup"],
    },
    {
      code: "module \"Foo\"\n{\n}",
      options: ["allman"],
    },
    {
      code: "namespace Foo {\n}",
      options: ["1tbs"],
    },
    {
      code: "namespace Foo {\n}",
      options: ["stroustrup"],
    },
    {
      code: "namespace Foo\n{\n}",
      options: ["allman"],
    },
  ],

  invalid: [
    // ---- from brace-style._js_.test.ts ----
    {
      code: "if (f) {\n  bar;\n}\nelse\n  baz;",
      output: "if (f) {\n  bar;\n} else\n  baz;",
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "var foo = () => { return; }",
      output: "var foo = () => {\n return; \n}",
      errors: [
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "function foo() { return; }",
      output: "function foo() {\n return; \n}",
      errors: [
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "function foo() \n { \n return; }",
      output: "function foo() { \n return; \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "!function foo() \n { \n return; }",
      output: "!function foo() { \n return; \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "if (foo) \n { \n bar(); }",
      output: "if (foo) { \n bar(); \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "if (a) { \nb();\n } else \n { c(); }",
      output: "if (a) { \nb();\n } else {\n c(); \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "while (foo) \n { \n bar(); }",
      output: "while (foo) { \n bar(); \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "for (;;) \n { \n bar(); }",
      output: "for (;;) { \n bar(); \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "switch (foo) \n { \n case \"bar\": break; }",
      output: "switch (foo) { \n case \"bar\": break; \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "switch (foo) \n { }",
      output: "switch (foo) { }",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "try \n { \n bar(); \n } catch (e) {}",
      output: "try { \n bar(); \n } catch (e) {}",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "try { \n bar(); \n } catch (e) \n {}",
      output: "try { \n bar(); \n } catch (e) {}",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "do \n { \n bar(); \n} while (true)",
      output: "do { \n bar(); \n} while (true)",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "for (foo in bar) \n { \n baz(); \n }",
      output: "for (foo in bar) { \n baz(); \n }",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "for (foo of bar) \n { \n baz(); \n }",
      output: "for (foo of bar) { \n baz(); \n }",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "try { \n bar(); \n }\ncatch (e) {\n}",
      output: "try { \n bar(); \n } catch (e) {\n}",
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
      output: "try { \n bar(); \n } catch (e) {\n} finally {\n}",
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "if (a) { \nb();\n } \n else { \nc();\n }",
      output: "if (a) { \nb();\n } else { \nc();\n }",
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "try { \n bar(); \n }\ncatch (e) {\n} finally {\n}",
      output: "try { \n bar(); \n }\ncatch (e) {\n}\n finally {\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
      output: "try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "if (a) { \nb();\n } else { \nc();\n }",
      output: "if (a) { \nb();\n }\n else { \nc();\n }",
      options: ["stroustrup"],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "if (foo) {\nbaz();\n} else if (bar) {\nbaz();\n}\nelse {\nqux();\n}",
      output: "if (foo) {\nbaz();\n}\n else if (bar) {\nbaz();\n}\nelse {\nqux();\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "if (foo) {\npoop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
      output: "if (foo) {\npoop();\n} \nelse if (bar) {\nbaz();\n}\n else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}",
      output: "try \n{ \n bar(); \n }\n catch (e) \n{\n}\n finally \n{\n}",
      options: ["allman"],
      errors: [
        { messageId: "sameLineOpen", line: 1 },
        { messageId: "sameLineOpen", line: 4 },
        { messageId: "sameLineOpen", line: 6 },
      ],
    },
    {
      code: "switch(x) { case 1: \nbar(); }\n ",
      output: "switch(x) \n{\n case 1: \nbar(); \n}\n ",
      options: ["allman"],
      errors: [
        { messageId: "sameLineOpen", line: 1 },
        { messageId: "blockSameLine", line: 1 },
        { messageId: "singleLineClose", line: 2 },
      ],
    },
    {
      code: "if (a) { \nb();\n } else { \nc();\n }",
      output: "if (a) \n{ \nb();\n }\n else \n{ \nc();\n }",
      options: ["allman"],
      errors: [
        { messageId: "sameLineOpen" },
        { messageId: "sameLineClose" },
        { messageId: "sameLineOpen" },
      ],
    },
    {
      code: "if (foo) {\nbaz();\n} else if (bar) {\nbaz();\n}\nelse {\nqux();\n}",
      output: "if (foo) \n{\nbaz();\n}\n else if (bar) \n{\nbaz();\n}\nelse \n{\nqux();\n}",
      options: ["allman"],
      errors: [
        { messageId: "sameLineOpen" },
        { messageId: "sameLineClose" },
        { messageId: "sameLineOpen" },
        { messageId: "sameLineOpen" },
      ],
    },
    {
      code: "if (foo)\n{ poop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
      output: "if (foo)\n{\n poop();\n} \nelse if (bar) \n{\nbaz();\n}\n else if (thing) \n{\nboom();\n}\nelse \n{\nqux();\n}",
      options: ["allman"],
      errors: [
        { messageId: "blockSameLine" },
        { messageId: "sameLineOpen" },
        { messageId: "sameLineClose" },
        { messageId: "sameLineOpen" },
        { messageId: "sameLineOpen" },
      ],
    },
    {
      code: "if (foo)\n{\n  bar(); }",
      output: "if (foo)\n{\n  bar(); \n}",
      options: ["allman"],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "try\n{\n  somethingRisky();\n} catch (e)\n{\n  handleError()\n}",
      output: "try\n{\n  somethingRisky();\n}\n catch (e)\n{\n  handleError()\n}",
      options: ["allman"],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "function foo() { return; \n}",
      output: "function foo() {\n return; \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "function foo() { a(); b(); return; \n}",
      output: "function foo() {\n a(); b(); return; \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "function foo() { \n return; }",
      output: "function foo() { \n return; \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "function foo() {\na();\nb();\nreturn; }",
      output: "function foo() {\na();\nb();\nreturn; \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "!function foo() { \n return; }",
      output: "!function foo() { \n return; \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "if (a) { b();\n } else { c(); }",
      output: "if (a) {\n b();\n } else { c(); }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "if (a) { b(); }\nelse { c(); }",
      output: "if (a) { b(); } else { c(); }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "while (foo) { \n bar(); }",
      output: "while (foo) { \n bar(); \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "for (;;) { bar(); \n }",
      output: "for (;;) {\n bar(); \n }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "switch (foo) \n { \n case \"bar\": break; }",
      output: "switch (foo) { \n case \"bar\": break; \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "switch (foo) \n { }",
      output: "switch (foo) { }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "try {  bar(); }\ncatch (e) { baz();  }",
      output: "try {  bar(); } catch (e) { baz();  }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "try \n { \n bar(); \n } catch (e) {}",
      output: "try { \n bar(); \n } catch (e) {}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "try { \n bar(); \n } catch (e) \n {}",
      output: "try { \n bar(); \n } catch (e) {}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "do \n { \n bar(); \n} while (true)",
      output: "do { \n bar(); \n} while (true)",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "for (foo in bar) \n { \n baz(); \n }",
      output: "for (foo in bar) { \n baz(); \n }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "try { \n bar(); \n }\ncatch (e) {\n}",
      output: "try { \n bar(); \n } catch (e) {\n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
      output: "try { \n bar(); \n } catch (e) {\n} finally {\n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "if (a) { \nb();\n } \n else { \nc();\n }",
      output: "if (a) { \nb();\n } else { \nc();\n }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "nextLineClose" }],
    },
    {
      code: "try { \n bar(); \n }\ncatch (e) {\n} finally {\n}",
      output: "try { \n bar(); \n }\ncatch (e) {\n}\n finally {\n}",
      options: ["stroustrup",{"allowSingleLine":true}],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "try { \n bar(); \n } catch (e) {\n}\n finally {\n}",
      output: "try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}",
      options: ["stroustrup",{"allowSingleLine":true}],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "if (a) { \nb();\n } else { \nc();\n }",
      output: "if (a) { \nb();\n }\n else { \nc();\n }",
      options: ["stroustrup",{"allowSingleLine":true}],
      errors: [{ messageId: "sameLineClose" }],
    },
    {
      code: "if (foo)\n{ poop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}",
      output: "if (foo)\n{\n poop();\n} \nelse if (bar) \n{\nbaz();\n}\n else if (thing) \n{\nboom();\n}\nelse \n{\nqux();\n}",
      options: ["allman",{"allowSingleLine":true}],
      errors: [
        { messageId: "blockSameLine" },
        { messageId: "sameLineOpen" },
        { messageId: "sameLineClose" },
        { messageId: "sameLineOpen" },
        { messageId: "sameLineOpen" },
      ],
    },
    {
      code: "if (foo) // comment \n{\nbar();\n}",
      output: null,
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "if (foo) {\n bar\n.baz }",
      output: "if (foo) {\n bar\n.baz \n}",
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "if (foo)\n{\n bar\n.baz }",
      output: "if (foo)\n{\n bar\n.baz \n}",
      options: ["allman"],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "if (foo) { bar\n.baz }",
      output: "if (foo) {\n bar\n.baz \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "if (foo) { bar\n.baz }",
      output: "if (foo) \n{\n bar\n.baz \n}",
      options: ["allman",{"allowSingleLine":true}],
      errors: [
        { messageId: "sameLineOpen" },
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "switch (x) {\n case 1: foo() }",
      output: "switch (x) {\n case 1: foo() \n}",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "class Foo\n{\n}",
      output: "class Foo {\n}",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "(class\n{\n})",
      output: "(class {\n})",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "class Foo{\n}",
      output: "class Foo\n{\n}",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },
    {
      code: "(class {\n})",
      output: "(class \n{\n})",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },
    {
      code: "class Foo {\nbar() {\n}}",
      output: "class Foo {\nbar() {\n}\n}",
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "(class Foo {\nbar() {\n}})",
      output: "(class Foo {\nbar() {\n}\n})",
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "class\nFoo{}",
      output: "class\nFoo\n{}",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },
    {
      code: "if (foo)\n{\n    bar\n}\nelse {\n    baz\n}",
      output: "if (foo) {\n    bar\n} else {\n    baz\n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "nextLineClose" },
      ],
    },
    {
      code: "class C {\n    static\n    {\n        foo;\n    }\n}",
      output: "class C {\n    static {\n        foo;\n    }\n}",
      options: ["1tbs"],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "class C {\n    static {foo;\n    }\n}",
      output: "class C {\n    static {\nfoo;\n    }\n}",
      options: ["1tbs"],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "class C {\n    static {\n        foo;}\n}",
      output: "class C {\n    static {\n        foo;\n}\n}",
      options: ["1tbs"],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "class C {\n    static\n    {foo;}\n}",
      output: "class C {\n    static {\nfoo;\n}\n}",
      options: ["1tbs"],
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "class C {\n    static\n    {}\n}",
      output: "class C {\n    static {}\n}",
      options: ["1tbs"],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "class C {\n    static\n    {\n        foo;\n    }\n}",
      output: "class C {\n    static {\n        foo;\n    }\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "class C {\n    static {foo;\n    }\n}",
      output: "class C {\n    static {\nfoo;\n    }\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "class C {\n    static {\n        foo;}\n}",
      output: "class C {\n    static {\n        foo;\n}\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "class C {\n    static\n    {foo;}\n}",
      output: "class C {\n    static {\nfoo;\n}\n}",
      options: ["stroustrup"],
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "class C {\n    static\n    {}\n}",
      output: "class C {\n    static {}\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "class C\n{\n    static{\n        foo;\n    }\n}",
      output: "class C\n{\n    static\n{\n        foo;\n    }\n}",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },
    {
      code: "class C\n{\n    static\n    {foo;\n    }\n}",
      output: "class C\n{\n    static\n    {\nfoo;\n    }\n}",
      options: ["allman"],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "class C\n{\n    static\n    {\n        foo;}\n}",
      output: "class C\n{\n    static\n    {\n        foo;\n}\n}",
      options: ["allman"],
      errors: [{ messageId: "singleLineClose" }],
    },
    {
      code: "class C\n{\n    static{foo;}\n}",
      output: "class C\n{\n    static\n{\nfoo;\n}\n}",
      options: ["allman"],
      errors: [
        { messageId: "sameLineOpen" },
        { messageId: "blockSameLine" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "class C\n{\n    static{}\n}",
      output: "class C\n{\n    static\n{}\n}",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },

    // ---- from brace-style._ts_.test.ts ----
    {
      code: "with (foo) \n { \n bar(); }",
      output: "with (foo) { \n bar(); \n}",
      errors: [
        { messageId: "nextLineOpen" },
        { messageId: "singleLineClose" },
      ],
    },
    {
      code: "with (foo) { bar(); \n }",
      output: "with (foo) {\n bar(); \n }",
      options: ["1tbs",{"allowSingleLine":true}],
      errors: [{ messageId: "blockSameLine" }],
    },
    {
      code: "module \"Foo\"\n{\n}",
      output: "module \"Foo\" {\n}",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "module \"Foo\"\n{\n}",
      output: "module \"Foo\" {\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "module \"Foo\" { \n }",
      output: "module \"Foo\" \n{ \n }",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },
    {
      code: "namespace Foo\n{\n}",
      output: "namespace Foo {\n}",
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "namespace Foo\n{\n}",
      output: "namespace Foo {\n}",
      options: ["stroustrup"],
      errors: [{ messageId: "nextLineOpen" }],
    },
    {
      code: "namespace Foo { \n }",
      output: "namespace Foo \n{ \n }",
      options: ["allman"],
      errors: [{ messageId: "sameLineOpen" }],
    },
  ],
});
