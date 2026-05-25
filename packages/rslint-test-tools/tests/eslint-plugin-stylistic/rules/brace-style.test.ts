import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('brace-style', null as never, {
  valid: [
    // 1tbs (default) baseline
    {
      code: 'function f() {\n    if (true)\n        return {x: 1}\n    else {\n        var y = 2\n        return y\n    }\n}',
    },
    {
      code: 'if (tag === 1) glyph.id = pbf.readVarint();\nelse if (tag === 2) glyph.bitmap = pbf.readBytes();',
    },
    { code: 'function foo () {\n  return;\n}' },
    { code: 'function a(b,\nc,\nd) { }' },
    { code: '!function foo () {\n  return;\n}' },
    { code: '!function a(b,\nc,\nd) { }' },
    { code: 'if (foo) {\n  bar();\n}' },
    { code: 'if (a) {\n  b();\n} else {\n  c();\n}' },
    { code: 'while (foo) {\n  bar();\n}' },
    { code: 'for (;;) {\n  bar();\n}' },
    { code: "switch (foo) {\n  case 'bar': break;\n}" },
    { code: 'try {\n  bar();\n} catch (e) {\n  baz();\n}' },
    { code: 'do {\n  bar();\n} while (true)' },
    { code: 'for (foo in bar) {\n  baz();\n}' },
    { code: 'if (a &&\n  b &&\n  c) {\n}' },
    { code: 'switch(0) {\n}' },
    { code: 'class Foo {\n}' },
    { code: '(class {\n})' },
    { code: 'class\nFoo {\n}' },
    { code: 'class Foo {\n    bar() {\n    }\n}' },

    // stroustrup
    { code: 'if (foo) {\n}\nelse {\n}', options: ['stroustrup'] },
    {
      code: 'try {\n  bar();\n}\ncatch (e) {\n  baz();\n}',
      options: ['stroustrup'],
    },
    { code: 'class Foo {\n}', options: ['stroustrup'] },
    { code: '(class {\n})', options: ['stroustrup'] },

    // allman
    { code: 'if (foo)\n{\n}\nelse\n{\n}', options: ['allman'] },
    {
      code: 'try\n{\n  bar();\n}\ncatch (e)\n{\n  baz();\n}',
      options: ['allman'],
    },
    { code: 'switch(x)\n{\n  case 1:\n    bar();\n}', options: ['allman'] },
    { code: 'class Foo\n{\n}', options: ['allman'] },
    { code: '(class\n{\n})', options: ['allman'] },
    { code: 'class\nFoo\n{\n}', options: ['allman'] },

    // 1tbs + allowSingleLine: true
    {
      code: 'function foo () { return; }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'function foo () { a(); b(); return; }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'function a(b,c,d) { }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: '!function foo () { return; }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: '!function a(b,c,d) { }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'if (foo) {  bar(); }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'if (a) { b(); } else { c(); }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'while (foo) {  bar(); }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'for (;;) {  bar(); }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'switch (foo) {  case "bar": break; }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'try {  bar(); } catch (e) { baz();  }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'do {  bar(); } while (true)',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'for (foo in bar) {  baz();  }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'if (a && b && c) {  }',
      options: ['1tbs', { allowSingleLine: true }],
    },
    { code: 'switch(0) {}', options: ['1tbs', { allowSingleLine: true }] },
    {
      code: 'if (foo) {}\nelse {}',
      options: ['stroustrup', { allowSingleLine: true }],
    },
    {
      code: 'try {  bar(); }\ncatch (e) { baz();  }',
      options: ['stroustrup', { allowSingleLine: true }],
    },
    {
      code: 'var foo = () => { return; }',
      options: ['stroustrup', { allowSingleLine: true }],
    },
    {
      code: 'if (foo) {}\nelse {}',
      options: ['allman', { allowSingleLine: true }],
    },
    {
      code: 'try {  bar(); }\ncatch (e) { baz();  }',
      options: ['allman', { allowSingleLine: true }],
    },
    {
      code: 'var foo = () => { return; }',
      options: ['allman', { allowSingleLine: true }],
    },
    {
      code: 'if (foo) { baz(); } else {\n  boom();\n}',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'if (foo) { baz(); } else if (bar) {\n  boom();\n}',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'if (foo) { baz(); } else\nif (bar) {\n  boom();\n}',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'try { somethingRisky(); } catch(e) {\n  handleError();\n}',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'if (tag === 1) fontstack.name = pbf.readString();\nelse if (tag === 2) fontstack.range = pbf.readString();\nelse if (tag === 3) {\n  var glyph = pbf.readMessage(readGlyph, {});\n  fontstack.glyphs[glyph.id] = glyph;\n}',
    },
    { code: 'switch(x) {}', options: ['allman', { allowSingleLine: true }] },
    { code: 'class Foo {}', options: ['1tbs', { allowSingleLine: true }] },
    { code: 'class Foo {}', options: ['allman', { allowSingleLine: true }] },
    { code: '(class {})', options: ['1tbs', { allowSingleLine: true }] },
    { code: '(class {})', options: ['allman', { allowSingleLine: true }] },

    // Standalone / Program / SwitchCase block parents — skipped per
    // STATEMENT_LIST_PARENTS.
    { code: '{}' },
    { code: 'if (foo) {\n}\n{\n}' },
    {
      code: 'switch (foo) {\n  case bar:\n    baz();\n    {\n      qux();\n    }\n}',
    },
    { code: '{\n}' },
    { code: '{\n  {\n  }\n}' },

    // https://github.com/eslint/eslint/issues/7974
    { code: 'class Ball {\n  throw() {}\n  catch() {}\n}' },
    { code: '({\n  and() {},\n  finally() {}\n})' },
    { code: '(class {\n  or() {}\n  else() {}\n})' },
    { code: 'if (foo) bar = function() {}\nelse baz()' },

    // Class static blocks
    {
      code: 'class C {\n    static {\n        foo;\n    }\n}',
      options: ['1tbs'],
    },
    {
      code: 'class C {\n    static {}\n\n    static {\n    }\n}',
      options: ['1tbs'],
    },
    {
      code: 'class C {\n    static { foo; }\n}',
      options: ['1tbs', { allowSingleLine: true }],
    },
    {
      code: 'class C {\n    static {\n        foo;\n    }\n}',
      options: ['stroustrup'],
    },
    {
      code: 'class C {\n    static {}\n\n    static {\n    }\n}',
      options: ['stroustrup'],
    },
    {
      code: 'class C {\n    static { foo; }\n}',
      options: ['stroustrup', { allowSingleLine: true }],
    },
    {
      code: 'class C\n{\n    static\n    {\n        foo;\n    }\n}',
      options: ['allman'],
    },
    { code: 'class C\n{\n    static\n    {}\n}', options: ['allman'] },
    {
      code: 'class C\n{\n    static {}\n\n    static { foo; }\n\n    static\n    { foo; }\n}',
      options: ['allman', { allowSingleLine: true }],
    },
    {
      code: 'class C {\n    static {\n        {\n            foo;\n        }\n    }\n}',
      options: ['1tbs'],
    },

    // TS `with` statement — parsed by tsgo even under strict mode.
    {
      code: 'with (foo) {\n  bar();\n}',
      filename: 'src/virtual.ts',
    },
    {
      code: 'with (foo) {  bar(); }',
      options: ['1tbs', { allowSingleLine: true }],
      filename: 'src/virtual.ts',
    },

    // TS namespace / module bodies — non-.tsx so generic-like syntax doesn't
    // collide with JSX.
    {
      code: 'module "Foo" {\n}',
      options: ['1tbs'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'module "Foo" {\n}',
      options: ['stroustrup'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'module "Foo"\n{\n}',
      options: ['allman'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'namespace Foo {\n}',
      options: ['1tbs'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'namespace Foo {\n}',
      options: ['stroustrup'],
      filename: 'src/virtual.ts',
    },
    {
      code: 'namespace Foo\n{\n}',
      options: ['allman'],
      filename: 'src/virtual.ts',
    },
  ],
  invalid: [
    {
      code: 'if (f) {\n  bar;\n}\nelse\n  baz;',
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'var foo = () => { return; }',
      errors: [
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'function foo() { return; }',
      errors: [
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'function foo() \n { \n return; }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    {
      code: '!function foo() \n { \n return; }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    {
      code: 'if (foo) \n { \n bar(); }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    {
      code: 'if (a) { \nb();\n } else \n { c(); }',
      errors: [
        { messageId: 'nextLineOpen' },
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'while (foo) \n { \n bar(); }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    {
      code: 'for (;;) \n { \n bar(); }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    {
      code: 'switch (foo) \n { \n case "bar": break; }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    { code: 'switch (foo) \n { }', errors: [{ messageId: 'nextLineOpen' }] },
    {
      code: 'try \n { \n bar(); \n } catch (e) {}',
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'try { \n bar(); \n } catch (e) \n {}',
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'do \n { \n bar(); \n} while (true)',
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'for (foo in bar) \n { \n baz(); \n }',
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'for (foo of bar) \n { \n baz(); \n }',
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'try { \n bar(); \n }\ncatch (e) {\n}',
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'try { \n bar(); \n } catch (e) {\n}\n finally {\n}',
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'if (a) { \nb();\n } \n else { \nc();\n }',
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'try { \n bar(); \n }\ncatch (e) {\n} finally {\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'try { \n bar(); \n } catch (e) {\n}\n finally {\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'if (a) { \nb();\n } else { \nc();\n }',
      options: ['stroustrup'],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'if (foo) {\nbaz();\n} else if (bar) {\nbaz();\n}\nelse {\nqux();\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'if (foo) {\npoop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'try { \n bar(); \n }\n catch (e) {\n}\n finally {\n}',
      options: ['allman'],
      errors: [
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineOpen' },
      ],
    },
    {
      code: 'switch(x) { case 1: \nbar(); }\n ',
      options: ['allman'],
      errors: [
        { messageId: 'sameLineOpen' },
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'if (a) { \nb();\n } else { \nc();\n }',
      options: ['allman'],
      errors: [
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineClose' },
        { messageId: 'sameLineOpen' },
      ],
    },
    {
      code: 'if (foo) {\nbaz();\n} else if (bar) {\nbaz();\n}\nelse {\nqux();\n}',
      options: ['allman'],
      errors: [
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineClose' },
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineOpen' },
      ],
    },
    {
      code: 'if (foo)\n{ poop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}',
      options: ['allman'],
      errors: [
        { messageId: 'blockSameLine' },
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineClose' },
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineOpen' },
      ],
    },
    {
      code: 'if (foo)\n{\n  bar(); }',
      options: ['allman'],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'try\n{\n  somethingRisky();\n} catch (e)\n{\n  handleError()\n}',
      options: ['allman'],
      errors: [{ messageId: 'sameLineClose' }],
    },

    // allowSingleLine: true
    {
      code: 'function foo() { return; \n}',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'function foo() { a(); b(); return; \n}',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'function foo() { \n return; }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'function foo() {\na();\nb();\nreturn; }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: '!function foo() { \n return; }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'if (a) { b();\n } else { c(); }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'if (a) { b(); }\nelse { c(); }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'while (foo) { \n bar(); }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'for (;;) { bar(); \n }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'switch (foo) \n { \n case "bar": break; }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
    },
    {
      code: 'switch (foo) \n { }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'try {  bar(); }\ncatch (e) { baz();  }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'try \n { \n bar(); \n } catch (e) {}',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'try { \n bar(); \n } catch (e) \n {}',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'do \n { \n bar(); \n} while (true)',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'for (foo in bar) \n { \n baz(); \n }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'try { \n bar(); \n }\ncatch (e) {\n}',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'try { \n bar(); \n } catch (e) {\n}\n finally {\n}',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'if (a) { \nb();\n } \n else { \nc();\n }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'nextLineClose' }],
    },
    {
      code: 'try { \n bar(); \n }\ncatch (e) {\n} finally {\n}',
      options: ['stroustrup', { allowSingleLine: true }],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'try { \n bar(); \n } catch (e) {\n}\n finally {\n}',
      options: ['stroustrup', { allowSingleLine: true }],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'if (a) { \nb();\n } else { \nc();\n }',
      options: ['stroustrup', { allowSingleLine: true }],
      errors: [{ messageId: 'sameLineClose' }],
    },
    {
      code: 'if (foo)\n{ poop();\n} \nelse if (bar) {\nbaz();\n} else if (thing) {\nboom();\n}\nelse {\nqux();\n}',
      options: ['allman', { allowSingleLine: true }],
      errors: [
        { messageId: 'blockSameLine' },
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineClose' },
        { messageId: 'sameLineOpen' },
        { messageId: 'sameLineOpen' },
      ],
    },

    // Comment interferes with fix
    {
      code: 'if (foo) // comment \n{\nbar();\n}',
      errors: [{ messageId: 'nextLineOpen' }],
    },

    // https://github.com/eslint/eslint/issues/7493
    {
      code: 'if (foo) {\n bar\n.baz }',
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'if (foo)\n{\n bar\n.baz }',
      options: ['allman'],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'if (foo) { bar\n.baz }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'if (foo) { bar\n.baz }',
      options: ['allman', { allowSingleLine: true }],
      errors: [
        { messageId: 'sameLineOpen' },
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'switch (x) {\n case 1: foo() }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'singleLineClose' }],
    },
    { code: 'class Foo\n{\n}', errors: [{ messageId: 'nextLineOpen' }] },
    { code: '(class\n{\n})', errors: [{ messageId: 'nextLineOpen' }] },
    {
      code: 'class Foo{\n}',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
    },
    {
      code: '(class {\n})',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
    },
    {
      code: 'class Foo {\nbar() {\n}}',
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: '(class Foo {\nbar() {\n}})',
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'class\nFoo{}',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
    },

    // https://github.com/eslint/eslint/issues/7621
    {
      code: 'if (foo)\n{\n    bar\n}\nelse {\n    baz\n}',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'nextLineClose' }],
    },

    // Class static blocks
    {
      code: 'class C {\n    static\n    {\n        foo;\n    }\n}',
      options: ['1tbs'],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'class C {\n    static {foo;\n    }\n}',
      options: ['1tbs'],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'class C {\n    static {\n        foo;}\n}',
      options: ['1tbs'],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'class C {\n    static\n    {foo;}\n}',
      options: ['1tbs'],
      errors: [
        { messageId: 'nextLineOpen' },
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'class C {\n    static\n    {}\n}',
      options: ['1tbs'],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'class C {\n    static\n    {\n        foo;\n    }\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'class C {\n    static {foo;\n    }\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'class C {\n    static {\n        foo;}\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'class C {\n    static\n    {foo;}\n}',
      options: ['stroustrup'],
      errors: [
        { messageId: 'nextLineOpen' },
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'class C {\n    static\n    {}\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'nextLineOpen' }],
    },
    {
      code: 'class C\n{\n    static{\n        foo;\n    }\n}',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
    },
    {
      code: 'class C\n{\n    static\n    {foo;\n    }\n}',
      options: ['allman'],
      errors: [{ messageId: 'blockSameLine' }],
    },
    {
      code: 'class C\n{\n    static\n    {\n        foo;}\n}',
      options: ['allman'],
      errors: [{ messageId: 'singleLineClose' }],
    },
    {
      code: 'class C\n{\n    static{foo;}\n}',
      options: ['allman'],
      errors: [
        { messageId: 'sameLineOpen' },
        { messageId: 'blockSameLine' },
        { messageId: 'singleLineClose' },
      ],
    },
    {
      code: 'class C\n{\n    static{}\n}',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
    },

    // TS `with` statement
    {
      code: 'with (foo) \n { \n bar(); }',
      errors: [{ messageId: 'nextLineOpen' }, { messageId: 'singleLineClose' }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'with (foo) { bar(); \n }',
      options: ['1tbs', { allowSingleLine: true }],
      errors: [{ messageId: 'blockSameLine' }],
      filename: 'src/virtual.ts',
    },

    // TS namespace / module
    {
      code: 'module "Foo"\n{\n}',
      errors: [{ messageId: 'nextLineOpen' }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'module "Foo"\n{\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'nextLineOpen' }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'module "Foo" { \n }',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'namespace Foo\n{\n}',
      errors: [{ messageId: 'nextLineOpen' }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'namespace Foo\n{\n}',
      options: ['stroustrup'],
      errors: [{ messageId: 'nextLineOpen' }],
      filename: 'src/virtual.ts',
    },
    {
      code: 'namespace Foo { \n }',
      options: ['allman'],
      errors: [{ messageId: 'sameLineOpen' }],
      filename: 'src/virtual.ts',
    },
  ],
});
