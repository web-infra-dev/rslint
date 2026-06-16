/**
 * @fileoverview Tests for indent rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/indent/indent._js_.test.ts
 *   packages/eslint-plugin/rules/indent/indent._jsx_.test.ts
 *   packages/eslint-plugin/rules/indent/indent._ts_.test.ts
 * (the ._css_ test file is excluded — rslint has no CSS parser.)
 *
 * Transformations applied per the porting spec:
 *  - run({ name, rule, lang, parserOptions, valid, invalid }) -> ruleTester.run('indent', null as never, { valid, invalid }).
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - The local `expectedErrors([line, expected, actual])` helper is evaluated to its
 *    final { messageId: 'wrongIndentation', data: { expected, actual }, line } objects
 *    (data.expected = `<n> space(s)`/`<n> tab(s)` or a raw string; message interpolates to
 *    "Expected indentation of <expected> but found <actual>.").
 *  - The two external fixtures read via readFileSync (indent-invalid-fixture-1.js /
 *    indent-valid-fixture-1.js) are inlined verbatim as the code/output of one invalid case.
 *  - The ._ts_ `individualNodeTests` reduce is evaluated to its generated valid/invalid cases.
 *  - The ._jsx_ `valids()/invalids()` helpers expand each test across parsers; only the
 *    `@typescript-eslint/parser` variant (the parser rslint emulates) is kept — every base
 *    jsx case has exactly one, so each upstream jsx test maps to exactly one case here. The
 *    trailing `// features: ..., parser: ...` comment those helpers append is preserved
 *    verbatim (it sits at column 0 and is not itself an indentation violation).
 *  - parserOptions / lang / name / rule / type / features dropped — rslint parses at esnext.
 *
 * The upstream `if (!skipBabel)` block (3 Babel/Flow invalid cases) is NOT ported: under
 * the installed ESLint (>= 10) `skipBabel` is true, so upstream itself never runs it; it is
 * recorded under KNOWN GAPS (Babel/Flow) below.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they are moved
 * to the KNOWN GAPS block at the bottom, each annotated with upstream-expected vs. rslint.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('indent', null as never, {
  valid: [
    // ---- from indent._js_.test.ts ----
    {
      code: "bridge.callHandler(\n  'getAppVersion', 'test23', function(responseData) {\n    window.ah.mobileAppVersion = responseData;\n  }\n);",
      options: [2],
    },
    {
      code: "bridge.callHandler(\n  'getAppVersion', 'test23', function(responseData) {\n    window.ah.mobileAppVersion = responseData;\n  });",
      options: [2],
    },
    {
      code: "bridge.callHandler(\n  'getAppVersion',\n  null,\n  function responseCallback(responseData) {\n    window.ah.mobileAppVersion = responseData;\n  }\n);",
      options: [2],
    },
    {
      code: "bridge.callHandler(\n  'getAppVersion',\n  null,\n  function responseCallback(responseData) {\n    window.ah.mobileAppVersion = responseData;\n  });",
      options: [2],
    },
    {
      code: 'function doStuff(keys) {\n    _.forEach(\n        keys,\n        key => {\n            doSomething(key);\n        }\n    );\n}',
      options: [4],
    },
    {
      code: "example(\n    function () {\n        console.log('example');\n    }\n);",
      options: [4],
    },
    {
      code: 'let foo = somethingList\n    .filter(x => {\n        return x;\n    })\n    .map(x => {\n        return 100 * x;\n    });',
      options: [4],
    },
    {
      code: 'var x = 0 &&\n    {\n        a: 1,\n        b: 2\n    };',
      options: [4],
    },
    { code: 'var x = 0 &&\n\t{\n\t\ta: 1,\n\t\tb: 2\n\t};', options: ['tab'] },
    {
      code: 'var x = 0 &&\n    {\n        a: 1,\n        b: 2\n    }||\n    {\n        c: 3,\n        d: 4\n    };',
      options: [4],
    },
    { code: "var x = [\n    'a',\n    'b',\n    'c'\n];", options: [4] },
    { code: "var x = ['a',\n    'b',\n    'c',\n];", options: [4] },
    { code: 'var x = 0 && 1;', options: [4] },
    { code: 'var x = 0 && { a: 1, b: 2 };', options: [4] },
    { code: 'var x = 0 &&\n    (\n        1\n    );', options: [4] },
    {
      code: "require('http').request({hostname: 'localhost',\n  port: 80}, function(res) {\n  res.end();\n});",
      options: [2],
    },
    {
      code: 'function test() {\n  return client.signUp(email, PASSWORD, { preVerified: true })\n    .then(function (result) {\n      // hi\n    })\n    .then(function () {\n      return FunctionalHelpers.clearBrowserState(self, {\n        contentServer: true,\n        contentServer1: true\n      });\n    });\n}',
      options: [2],
    },
    {
      code: "it('should... some lengthy test description that is forced to be' +\n  'wrapped into two lines since the line length limit is set', () => {\n  expect(true).toBe(true);\n});",
      options: [2],
    },
    {
      code: 'function test() {\n    return client.signUp(email, PASSWORD, { preVerified: true })\n        .then(function (result) {\n            var x = 1;\n            var y = 1;\n        }, function(err){\n            var o = 1 - 2;\n            var y = 1 - 2;\n            return true;\n        })\n}',
      options: [4],
    },
    { code: 'import foo from "foo"\n\n;(() => {})()', options: [4] },
    {
      code: 'function test() {\n    return client.signUp(email, PASSWORD, { preVerified: true })\n    .then(function (result) {\n        var x = 1;\n        var y = 1;\n    }, function(err){\n        var o = 1 - 2;\n        var y = 1 - 2;\n        return true;\n    });\n}',
      options: [4, { MemberExpression: 0 }],
    },
    { code: '// hi', options: [2, { VariableDeclarator: 1, SwitchCase: 1 }] },
    {
      code: 'var Command = function() {\n  var fileList = [],\n      files = []\n\n  files.concat(fileList)\n};',
      options: [2, { VariableDeclarator: { var: 2, let: 2, const: 3 } }],
    },
    { code: '  ', options: [2, { VariableDeclarator: 1, SwitchCase: 1 }] },
    {
      code: "if(data) {\n  console.log('hi');\n  b = true;};",
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: "foo = () => {\n  console.log('hi');\n  return true;};",
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: "function test(data) {\n  console.log('hi');\n  return true;};",
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: "var test = function(data) {\n  console.log('hi');\n};",
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: "arr.forEach(function(data) {\n  otherdata.forEach(function(zero) {\n    console.log('hi');\n  }) });",
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'a = [\n    ,3\n]',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: "[\n  ['gzip', 'gunzip'],\n  ['gzip', 'unzip'],\n  ['deflate', 'inflate'],\n  ['deflateRaw', 'inflateRaw'],\n].forEach(function(method) {\n  console.log(method);\n});",
      options: [2, { SwitchCase: 1, VariableDeclarator: 2 }],
    },
    {
      code: 'test(123, {\n    bye: {\n        hi: [1,\n            {\n                b: 2\n            }\n        ]\n    }\n});',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var xyz = 2,\n    lmn = [\n        {\n            a: 1\n        }\n    ];',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'lmnn = [{\n    a: 1\n},\n{\n    b: 2\n}, {\n    x: 2\n}];',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    '[{\n    foo: 1\n}, {\n    foo: 2\n}, {\n    foo: 3\n}]',
    'foo([\n    bar\n], [\n    baz\n], [\n    qux\n]);',
    {
      code: "abc({\n    test: [\n        [\n            c,\n            xyz,\n            2\n        ].join(',')\n    ]\n});",
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'abc = {\n  test: [\n    [\n      c,\n      xyz,\n      2\n    ]\n  ]\n};',
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'abc(\n  {\n    a: 1,\n    b: 2\n  }\n);',
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'abc({\n    a: 1,\n    b: 2\n});',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var abc =\n  [\n    c,\n    xyz,\n    {\n      a: 1,\n      b: 2\n    }\n  ];',
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var abc = [\n  c,\n  xyz,\n  {\n    a: 1,\n    b: 2\n  }\n];',
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var abc = 5,\n    c = 2,\n    xyz =\n    {\n      a: 1,\n      b: 2\n    };',
      options: [
        2,
        { VariableDeclarator: 2, SwitchCase: 1, assignmentOperator: 0 },
      ],
    },
    'var\n    x = {\n        a: 1,\n    },\n    y = {\n        b: 2\n    }',
    'const\n    x = {\n        a: 1,\n    },\n    y = {\n        b: 2\n    }',
    'let\n    x = {\n        a: 1,\n    },\n    y = {\n        b: 2\n    }',
    'var foo = { a: 1 }, bar = {\n    b: 2\n};',
    'var foo = { a: 1 }, bar = {\n        b: 2\n    },\n    baz = {\n        c: 3\n    }',
    'const {\n        foo\n    } = 1,\n    bar = 2',
    {
      code: 'var foo = 1,\n  bar =\n    2',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: 'var foo = 1,\n  bar\n    = 2',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: 'var foo\n    = 1,\n  bar\n    = 2',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: 'var foo\n    =\n      1,\n  bar\n    =\n      2',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: 'var foo\n    = (1),\n  bar\n    = (2)',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: "let foo = 'foo',\n    bar = bar;\nconst a = 'a',\n      b = 'b';",
      options: [2, { VariableDeclarator: 'first' }],
    },
    {
      code: "let foo = 'foo',\n    bar = bar  // <-- no semicolon here\nconst a = 'a',\n      b = 'b'  // <-- no semicolon here",
      options: [2, { VariableDeclarator: 'first' }],
    },
    {
      code: 'var foo = 1,\n    bar = 2,\n    baz = 3\n;',
      options: [2, { VariableDeclarator: { var: 2 } }],
    },
    {
      code: 'var foo = 1,\n    bar = 2,\n    baz = 3\n    ;',
      options: [2, { VariableDeclarator: { var: 2 } }],
    },
    {
      code: "var foo = 'foo',\n    bar = bar;",
      options: [2, { VariableDeclarator: { var: 'first' } }],
    },
    {
      code: "var foo = 'foo',\n    bar = 'bar'  // <-- no semicolon here",
      options: [2, { VariableDeclarator: { var: 'first' } }],
    },
    {
      code: 'let foo = 1,\n    bar = 2,\n    baz',
      options: [2, { VariableDeclarator: 'first' }],
    },
    { code: 'let\n    foo', options: [4, { VariableDeclarator: 'first' }] },
    {
      code: 'let foo = 1,\n    bar =\n    2',
      options: [2, { VariableDeclarator: 'first', assignmentOperator: 0 }],
    },
    {
      code: 'var a = {\n  a: 1,\n  b: 2\n};',
      options: [2, { VariableDeclarator: 'first' }],
    },
    {
      code: 'var a = 2,\n    c = {\n      a: 1,\n      b: 2\n    },\n    b = 2;',
      options: [2, { VariableDeclarator: 'first' }],
    },
    {
      code: 'var x = {\n      a: 1,\n      b: 2\n    },\n    y = {\n      c: 1,\n      d: 3\n    },\n    z = 5;',
      options: [2, { VariableDeclarator: 'first' }],
    },
    {
      code: 'var abc =\n    {\n      a: 1,\n      b: 2\n    };',
      options: [
        2,
        { VariableDeclarator: 2, SwitchCase: 1, assignmentOperator: 2 },
      ],
    },
    {
      code: 'var a = new abc({\n        a: 1,\n        b: 2\n    }),\n    b = 2;',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var a = 2,\n  c = {\n    a: 1,\n    b: 2\n  },\n  b = 2;',
      options: [2, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var x = 2,\n    y = {\n      a: 1,\n      b: 2\n    },\n    b = 2;',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'var e = {\n      a: 1,\n      b: 2\n    },\n    b = 2;',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'var a = {\n  a: 1,\n  b: 2\n};',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'function test() {\n  if (true ||\n            false){\n    console.log(val);\n  }\n}',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    'var foo = bar ||\n    !(\n        baz\n    );',
    'for (var foo = 1;\n    foo < 10;\n    foo++) {}',
    'for (\n    var foo = 1;\n    foo < 10;\n    foo++\n) {}',
    {
      code: 'for (var val in obj)\n  if (true)\n    console.log(val);',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    { code: 'with (a)\n    b();', options: [4] },
    { code: 'with (a)\n    b();\nc();', options: [4] },
    {
      code: 'if(true)\n  if (true)\n    if (true)\n      console.log(val);',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'function hi(){     var a = 1;\n  y++;                   x++;\n}',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'for(;length > index; index++)if(NO_HOLES || index in self){\n  x++;\n}',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'function test(){\n  switch(length){\n    case 1: return function(a){\n      return fn.call(that, a);\n    };\n  }\n}',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'var geometry = 2,\nrotate = 2;',
      options: [2, { VariableDeclarator: 0 }],
    },
    {
      code: 'var geometry,\n    rotate;',
      options: [4, { VariableDeclarator: 1 }],
    },
    {
      code: 'var geometry,\n\trotate;',
      options: ['tab', { VariableDeclarator: 1 }],
    },
    {
      code: 'var geometry,\n  rotate;',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: 'var geometry,\n    rotate;',
      options: [2, { VariableDeclarator: 2 }],
    },
    {
      code: 'let geometry,\n    rotate;',
      options: [2, { VariableDeclarator: 2 }],
    },
    {
      code: 'const geometry = 2,\n    rotate = 3;',
      options: [2, { VariableDeclarator: 2 }],
    },
    {
      code: 'var geometry, box, face1, face2, colorT, colorB, sprite, padding, maxWidth,\n  height, rotate;',
      options: [2, { SwitchCase: 1 }],
    },
    {
      code: 'var geometry, box, face1, face2, colorT, colorB, sprite, padding, maxWidth;',
      options: [2, { SwitchCase: 1 }],
    },
    { code: 'if (1 < 2){\n//hi sd\n}', options: [2] },
    { code: 'while (1 < 2){\n  //hi sd\n}', options: [2] },
    { code: "while (1 < 2) console.log('hi');", options: [2] },
    {
      code: '[a, boop,\n    c].forEach((index) => {\n    index;\n});',
      options: [4],
    },
    {
      code: '[a, b,\n    c].forEach(function(index){\n    return index;\n});',
      options: [4],
    },
    { code: '[a, b, c].forEach((index) => {\n    index;\n});', options: [4] },
    {
      code: '[a, b, c].forEach(function(index){\n    return index;\n});',
      options: [4],
    },
    {
      code: '(foo)\n    .bar([\n        baz\n    ]);',
      options: [4, { MemberExpression: 1 }],
    },
    {
      code: 'switch (x) {\n    case "foo":\n        a();\n        break;\n    case "bar":\n        switch (y) {\n            case "1":\n                break;\n            case "2":\n                a = 6;\n                break;\n        }\n    case "test":\n        break;\n}',
      options: [4, { SwitchCase: 1 }],
    },
    {
      code: 'switch (x) {\n        case "foo":\n            a();\n            break;\n        case "bar":\n            switch (y) {\n                    case "1":\n                        break;\n                    case "2":\n                        a = 6;\n                        break;\n            }\n        case "test":\n            break;\n}',
      options: [4, { SwitchCase: 2 }],
    },
    {
      code: 'switch (a) {\ncase "foo":\n    a();\n    break;\ncase "bar":\n    switch(x){\n    case \'1\':\n        break;\n    case \'2\':\n        a = 6;\n        break;\n    }\n}',
      options: [4, { SwitchCase: 0 }],
    },
    {
      code: 'switch (a) {\ncase "foo":\n    a();\n    break;\ncase "bar":\n    if(x){\n        a = 2;\n    }\n    else{\n        a = 6;\n    }\n}',
      options: [4, { SwitchCase: 0 }],
    },
    {
      code: 'switch (a) {\ncase "foo":\n    a();\n    break;\ncase "bar":\n    if(x){\n        a = 2;\n    }\n    else\n        a = 6;\n}',
      options: [4, { SwitchCase: 0 }],
    },
    {
      code: 'switch (a) {\ncase "foo":\n    a();\n    break;\ncase "bar":\n    a(); break;\ncase "baz":\n    a(); break;\n}',
      options: [4, { SwitchCase: 0 }],
    },
    'switch (0) {\n}',
    {
      code: 'function foo() {\n    var a = "a";\n    switch(a) {\n    case "a":\n        return "A";\n    case "b":\n        return "B";\n    }\n}\nfoo();',
      options: [4, { SwitchCase: 0 }],
    },
    {
      code: 'switch(value){\n    case "1":\n    case "2":\n        a();\n        break;\n    default:\n        a();\n        break;\n}\nswitch(value){\n    case "1":\n        a();\n        break;\n    case "2":\n        break;\n    default:\n        break;\n}',
      options: [4, { SwitchCase: 1 }],
    },
    'var obj = {foo: 1, bar: 2};\nwith (obj) {\n    console.log(foo + bar);\n}',
    'if (a) {\n    (1 + 2 + 3); // no error on this line\n}',
    'switch(value){ default: a(); break; }',
    {
      code: "import {addons} from 'react/addons'\nimport React from 'react'",
      options: [2],
    },
    "import {\n    foo,\n    bar,\n    baz\n} from 'qux';",
    "var foo = 0, bar = 0; baz = 0;\nexport {\n    foo,\n    bar,\n    baz\n} from 'qux';",
    { code: 'var a = 1,\n    b = 2,\n    c = 3;', options: [4] },
    { code: 'var a = 1\n    ,b = 2\n    ,c = 3;', options: [4] },
    { code: "while (1 < 2) console.log('hi')", options: [2] },
    {
      code: "function salutation () {\n  switch (1) {\n    case 0: return console.log('hi')\n    case 1: return console.log('hey')\n  }\n}",
      options: [2, { SwitchCase: 1 }],
    },
    {
      code: "var items = [\n  {\n    foo: 'bar'\n  }\n];",
      options: [2, { VariableDeclarator: 2 }],
    },
    {
      code: "const a = 1,\n      b = 2;\nconst items1 = [\n  {\n    foo: 'bar'\n  }\n];\nconst items2 = Items(\n  {\n    foo: 'bar'\n  }\n);",
      options: [2, { VariableDeclarator: 3 }],
    },
    {
      code: 'const geometry = 2,\n      rotate = 3;\nvar a = 1,\n  b = 2;\nlet light = true,\n    shadow = false;',
      options: [2, { VariableDeclarator: { const: 3, let: 2 } }],
    },
    {
      code: 'const abc = 5,\n      c = 2,\n      xyz =\n      {\n        a: 1,\n        b: 2\n      };\nlet abc2 = 5,\n  c2 = 2,\n  xyz2 =\n  {\n    a: 1,\n    b: 2\n  };\nvar abc3 = 5,\n    c3 = 2,\n    xyz3 =\n    {\n      a: 1,\n      b: 2\n    };',
      options: [
        2,
        {
          VariableDeclarator: { var: 2, const: 3 },
          SwitchCase: 1,
          assignmentOperator: 0,
        },
      ],
    },
    {
      code: "module.exports = {\n  'Unit tests':\n  {\n    rootPath: './',\n    environment: 'node',\n    tests:\n    [\n      'test/test-*.js'\n    ],\n    sources:\n    [\n      '*.js',\n      'test/**.js'\n    ]\n  }\n};",
      options: [2],
    },
    { code: 'foo =\n  bar;', options: [2] },
    { code: 'foo = (\n  bar\n);', options: [2] },
    {
      code: "var path     = require('path')\n  , crypto    = require('crypto')\n  ;",
      options: [2],
    },
    'var a = 1\n    ,b = 2\n    ;',
    {
      code: 'export function create (some,\n                        argument) {\n  return Object.create({\n    a: some,\n    b: argument\n  });\n};',
      options: [2, { FunctionDeclaration: { parameters: 'first' } }],
    },
    {
      code: 'export function create (id, xfilter, rawType,\n                        width=defaultWidth, height=defaultHeight,\n                        footerHeight=defaultFooterHeight,\n                        padding=defaultPadding) {\n  // ... function body, indented two spaces\n}',
      options: [2, { FunctionDeclaration: { parameters: 'first' } }],
    },
    {
      code: 'var obj = {\n  foo: function () {\n    return new p()\n      .then(function (ok) {\n        return ok;\n      }, function () {\n        // ignore things\n      });\n  }\n};',
      options: [2],
    },
    { code: 'a.b()\n  .c(function(){\n    var a;\n  }).d.e;', options: [2] },
    {
      code: "const YO = 'bah',\n      TE = 'mah'\n\nvar res,\n    a = 5,\n    b = 4",
      options: [2, { VariableDeclarator: { var: 2, let: 2, const: 3 } }],
    },
    {
      code: "const YO = 'bah',\n      TE = 'mah'\n\nvar res,\n    a = 5,\n    b = 4\n\nif (YO) console.log(TE)",
      options: [2, { VariableDeclarator: { var: 2, let: 2, const: 3 } }],
    },
    {
      code: "var foo = 'foo',\n  bar = 'bar',\n  baz = function() {\n\n  }\n\nfunction hello () {\n\n}",
      options: [2],
    },
    {
      code: "var obj = {\n  send: function () {\n    return P.resolve({\n      type: 'POST'\n    })\n      .then(function () {\n        return true;\n      }, function () {\n        return false;\n      });\n  }\n};",
      options: [2],
    },
    {
      code: "var obj = {\n  send: function () {\n    return P.resolve({\n      type: 'POST'\n    })\n    .then(function () {\n      return true;\n    }, function () {\n      return false;\n    });\n  }\n};",
      options: [2, { MemberExpression: 0 }],
    },
    "const someOtherFunction = argument => {\n        console.log(argument);\n    },\n    someOtherValue = 'someOtherValue';",
    {
      code: "[\n  'a',\n  'b'\n].sort().should.deepEqual([\n  'x',\n  'y'\n]);",
      options: [2],
    },
    {
      code: 'var a = 1,\n    B = class {\n      constructor(){}\n      a(){}\n      get b(){}\n    };',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
    },
    {
      code: 'var a = 1,\n    B =\n    class {\n      constructor(){}\n      a(){}\n      get b(){}\n    },\n    c = 3;',
      options: [
        2,
        { VariableDeclarator: 2, SwitchCase: 1, assignmentOperator: 0 },
      ],
    },
    {
      code: 'class A{\n    constructor(){}\n    a(){}\n    get b(){}\n}',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    {
      code: 'var A = class {\n    constructor(){}\n    a(){}\n    get b(){}\n}',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
    },
    { code: 'var a = {\n  some: 1\n  , name: 2\n};', options: [2] },
    {
      code: "a.c = {\n    aa: function() {\n        'test1';\n        return 'aa';\n    }\n    , bb: function() {\n        return this.bb();\n    }\n};",
      options: [4],
    },
    {
      code: "var a =\n{\n    actions:\n    [\n        {\n            name: 'compile'\n        }\n    ]\n};",
      options: [
        4,
        { VariableDeclarator: 0, SwitchCase: 1, assignmentOperator: 0 },
      ],
    },
    {
      code: "var a =\n[\n    {\n        name: 'compile'\n    }\n];",
      options: [
        4,
        { VariableDeclarator: 0, SwitchCase: 1, assignmentOperator: 0 },
      ],
    },
    '[[\n], function(\n    foo\n) {}\n]',
    "define([\n    'foo'\n], function(\n    bar\n) {\n    baz;\n}\n)",
    {
      code: "const func = function (opts) {\n    return Promise.resolve()\n    .then(() => {\n        [\n            'ONE', 'TWO'\n        ].forEach(command => { doSomething(); });\n    });\n};",
      options: [4, { MemberExpression: 0 }],
    },
    {
      code: "const func = function (opts) {\n    return Promise.resolve()\n        .then(() => {\n            [\n                'ONE', 'TWO'\n            ].forEach(command => { doSomething(); });\n        });\n};",
      options: [4],
    },
    {
      code: 'var haveFun = function () {\n    SillyFunction(\n        {\n            value: true,\n        },\n        {\n            _id: true,\n        }\n    );\n};',
      options: [4],
    },
    {
      code: 'var haveFun = function () {\n    new SillyFunction(\n        {\n            value: true,\n        },\n        {\n            _id: true,\n        }\n    );\n};',
      options: [4],
    },
    {
      code: 'let object1 = {\n  doThing() {\n    return _.chain([])\n      .map(v => (\n        {\n          value: true,\n        }\n      ))\n      .value();\n  }\n};',
      options: [2],
    },
    {
      code: 'var foo = {\n    bar: 1,\n    baz: {\n      qux: 2\n    }\n  },\n  bar = 1;',
      options: [2],
    },
    { code: 'class Foo\n  extends Bar {\n  baz() {}\n}', options: [2] },
    { code: 'class Foo extends\n  Bar {\n  baz() {}\n}', options: [2] },
    {
      code: 'class Foo extends\n  (\n    Bar\n  ) {\n  baz() {}\n}',
      options: [2],
    },
    { code: 'var Foo = class\n  extends Bar {\n  baz() {}\n}', options: [2] },
    { code: 'var Foo = class extends\n  Bar {\n  baz() {}\n}', options: [2] },
    {
      code: 'var Foo = class extends\n  (\n    Bar\n  ) {\n  baz() {}\n}',
      options: [2],
    },
    {
      code: "fs.readdirSync(path.join(__dirname, '../rules')).forEach(name => {\n  files[name] = foo;\n});",
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: '(function(){\nfunction foo(x) {\n  return x + 1;\n}\n})();',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: '(function(){\n        function foo(x) {\n            return x + 1;\n        }\n})();',
      options: [4, { outerIIFEBody: 2 }],
    },
    {
      code: '(function(x, y){\nfunction foo(x) {\n  return x + 1;\n}\n})(1, 2);',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: '(function(){\nfunction foo(x) {\n  return x + 1;\n}\n}());',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: '!function(){\nfunction foo(x) {\n  return x + 1;\n}\n}();',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: '!function(){\n\t\t\tfunction foo(x) {\n\t\t\t\treturn x + 1;\n\t\t\t}\n}();',
      options: ['tab', { outerIIFEBody: 3 }],
    },
    {
      code: 'var out = function(){\n  function fooVar(x) {\n    return x + 1;\n  }\n};',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: 'var ns = function(){\nfunction fooVar(x) {\n  return x + 1;\n}\n}();',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: 'ns = function(){\nfunction fooVar(x) {\n  return x + 1;\n}\n}();',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: 'var ns = (function(){\nfunction fooVar(x) {\n  return x + 1;\n}\n}(x));',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: 'var ns = (function(){\n        function fooVar(x) {\n            return x + 1;\n        }\n}(x));',
      options: [4, { outerIIFEBody: 2 }],
    },
    {
      code: 'var obj = {\n  foo: function() {\n    return true;\n  }\n};',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: 'while (\n  function() {\n    return true;\n  }()) {\n\n  x = x + 1;\n};',
      options: [2, { outerIIFEBody: 20 }],
    },
    {
      code: '(() => {\nfunction foo(x) {\n  return x + 1;\n}\n})();',
      options: [2, { outerIIFEBody: 0 }],
    },
    { code: 'function foo() {\n}', options: ['tab', { outerIIFEBody: 0 }] },
    {
      code: ';(() => {\nfunction foo(x) {\n  return x + 1;\n}\n})();',
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: "if(data) {\n  console.log('hi');\n}",
      options: [2, { outerIIFEBody: 0 }],
    },
    {
      code: '(function(x) {\n    return x + 1;\n})();',
      options: [4, { outerIIFEBody: 'off' }],
    },
    {
      code: '(function(x) {\nreturn x + 1;\n})();',
      options: [4, { outerIIFEBody: 'off' }],
    },
    {
      code: ';(() => {\n    function x(y) {\n        return y + 1;\n    }\n})();',
      options: [4, { outerIIFEBody: 'off' }],
    },
    {
      code: ';(() => {\nfunction x(y) {\n    return y + 1;\n}\n})();',
      options: [4, { outerIIFEBody: 'off' }],
    },
    { code: 'function foo() {\n}', options: [4, { outerIIFEBody: 'off' }] },
    { code: 'Buffer.length', options: [4, { MemberExpression: 1 }] },
    {
      code: "Buffer\n    .indexOf('a')\n    .toString()",
      options: [4, { MemberExpression: 1 }],
    },
    { code: 'Buffer.\n    length', options: [4, { MemberExpression: 1 }] },
    {
      code: 'Buffer\n    .foo\n    .bar',
      options: [4, { MemberExpression: 1 }],
    },
    {
      code: 'Buffer\n\t.foo\n\t.bar',
      options: ['tab', { MemberExpression: 1 }],
    },
    {
      code: 'Buffer\n    .foo\n    .bar',
      options: [2, { MemberExpression: 2 }],
    },
    '(\n    foo\n        .bar\n)',
    '(\n    (\n        foo\n            .bar\n    )\n)',
    '(\n    foo\n)\n    .bar',
    '(\n    (\n        foo\n    )\n        .bar\n)',
    '(\n    (\n        foo\n    )\n        [\n            (\n                bar\n            )\n        ]\n)',
    '(\n    foo[bar]\n)\n    .baz',
    '(\n    (foo.bar)\n)\n    .baz',
    {
      code: 'MemberExpression\n.can\n  .be\n    .turned\n .off();',
      options: [4, { MemberExpression: 'off' }],
    },
    {
      code: 'foo = bar.baz()\n    .bip();',
      options: [4, { MemberExpression: 1 }],
    },
    'function foo() {\n    new\n        .target\n}',
    'function foo() {\n    new.\n        target\n}',
    {
      code: 'if (foo) {\n  bar();\n} else if (baz) {\n  foobar();\n} else if (qux) {\n  qux();\n}',
      options: [2],
    },
    {
      code: 'function foo(aaa,\n  bbb, ccc, ddd) {\n    bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 1, body: 2 } }],
    },
    {
      code: 'function foo(aaa, bbb,\n      ccc, ddd) {\n  bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 3, body: 1 } }],
    },
    {
      code: 'function foo(aaa,\n    bbb,\n    ccc) {\n            bar();\n}',
      options: [4, { FunctionDeclaration: { parameters: 1, body: 3 } }],
    },
    {
      code: 'function foo(aaa,\n             bbb, ccc,\n             ddd, eee, fff) {\n  bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 'first', body: 1 } }],
    },
    {
      code: 'function foo(aaa, bbb)\n{\n      bar();\n}',
      options: [2, { FunctionDeclaration: { body: 3 } }],
    },
    {
      code: 'function foo(\n  aaa,\n  bbb) {\n    bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 'first', body: 2 } }],
    },
    {
      code: 'var foo = function(aaa,\n    bbb,\n    ccc,\n    ddd) {\nbar();\n}',
      options: [2, { FunctionExpression: { parameters: 2, body: 0 } }],
    },
    {
      code: 'var foo = function(aaa,\n  bbb,\n  ccc) {\n                    bar();\n}',
      options: [2, { FunctionExpression: { parameters: 1, body: 10 } }],
    },
    {
      code: 'var foo = function(aaa,\n                   bbb, ccc, ddd,\n                   eee, fff) {\n    bar();\n}',
      options: [4, { FunctionExpression: { parameters: 'first', body: 1 } }],
    },
    {
      code: 'var foo = function(\n  aaa, bbb, ccc,\n  ddd, eee) {\n      bar();\n}',
      options: [2, { FunctionExpression: { parameters: 'first', body: 3 } }],
    },
    {
      code: 'foo.bar(\n      baz, qux, function() {\n            qux;\n      }\n);',
      options: [
        2,
        { FunctionExpression: { body: 3 }, CallExpression: { arguments: 3 } },
      ],
    },
    {
      code: 'function foo() {\n  function bar() {\n    baz();\n  }\n}',
      options: [2, { FunctionDeclaration: { body: 1 } }],
    },
    {
      code: 'function foo() {\n  function bar(baz,\n      qux) {\n    foobar();\n  }\n}',
      options: [2, { FunctionDeclaration: { body: 1, parameters: 2 } }],
    },
    { code: '((\n    foo\n))', options: [4] },
    { code: 'foo\n  ? bar\n  : baz', options: [2] },
    { code: 'foo = (bar ?\n  baz :\n  qux\n);', options: [2] },
    {
      code: 'condition\n  ? () => {\n    return true\n  }\n  : condition2\n    ? () => {\n      return true\n    }\n    : () => {\n      return false\n    }',
      options: [2],
    },
    {
      code: 'condition\n  ? () => {\n    return true\n  }\n  : condition2\n    ? () => {\n      return true\n    }\n    : () => {\n      return false\n    }',
      options: [2, { offsetTernaryExpressions: false }],
    },
    {
      code: 'condition\n  ? new Foo({\n    })\n  : condition2\n    ? new Bar({\n      })\n    : new Baz({\n      })',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: 'condition\n  ? new Foo({\n  })\n  : condition2\n    ? new Bar({\n    })\n    : new Baz({\n    })',
      options: [2, { offsetTernaryExpressions: { NewExpression: false } }],
    },
    {
      code: 'condition\n  ? () => {\n      return true\n    }\n  : condition2\n    ? () => {\n        return true\n      }\n    : () => {\n        return false\n      }',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: 'condition\n    ? () => {\n            return true\n        }\n    : condition2\n        ? () => {\n                return true\n            }\n        : () => {\n                return false\n            }',
      options: [4, { offsetTernaryExpressions: true }],
    },
    {
      code: 'condition1\n  ? condition2\n    ? Promise.resolve(1)\n    : Promise.resolve(2)\n  : Promise.resolve(3)',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: 'condition1\n  ? Promise.resolve(1)\n  : condition2\n    ? Promise.resolve(2)\n    : Promise.resolve(3)',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: 'condition\n\t? () => {\n\t\t\treturn true\n\t\t}\n\t: condition2\n\t\t? () => {\n\t\t\t\treturn true\n\t\t\t}\n\t\t: () => {\n\t\t\t\treturn false\n\t\t\t}',
      options: ['tab', { offsetTernaryExpressions: true }],
    },
    {
      code: "const _obj = {\n  condition:\n    list.length > 3\n      ? t('string', {\n          num: list.length,\n        })\n      : '',\n};",
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: "const _obj = {\n  condition:\n    list.length > 3\n      ? t('string', {\n        num: list.length,\n      })\n      : '',\n};",
      options: [2, { offsetTernaryExpressions: false }],
    },
    {
      code: 'condition1\n  ? condition2\n    ? t()\n    : t({\n        foo,\n      })\n  : () => {\n      t()\n    }',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: "const _obj = {\n  condition:\n    list.length > 3\n      ? t('string', {\n        num: list.length,\n      })\n      : '',\n};",
      options: [2, { offsetTernaryExpressions: { CallExpression: false } }],
    },
    {
      code: 'condition1\n  ? condition2\n    ? t()\n    : t({\n      foo,\n    })\n  : () => {\n      t()\n    }',
      options: [2, { offsetTernaryExpressions: { CallExpression: false } }],
    },
    {
      code: 'isHeader(1)\n  ? renderSectionHeader?.(\n    typeof item === "string" ? item : "",\n    virtualRow.size,\n  )\n  : renderItem(\n    item,\n    virtualRow.size,\n  )',
      options: [2, { offsetTernaryExpressions: { CallExpression: false } }],
    },
    {
      code: 'isHeader(1)\n  ? renderSectionHeader?.(\n      typeof item === "string" ? item : "",\n      virtualRow.size,\n    )\n  : renderItem(\n      item,\n      virtualRow.size,\n    )',
      options: [2, { offsetTernaryExpressions: { CallExpression: true } }],
    },
    {
      code: 'menus\n  ? await Promise.all(\n    menus.map(async (menu) => ({\n      menuName: menu.name,\n      menu: await resolveUrlToFile(menu.fileUrl),\n    })),\n  )\n  : []',
      options: [2, { offsetTernaryExpressions: { AwaitExpression: false } }],
    },
    {
      code: 'menus\n  ? await Promise.all(\n      menus.map(async (menu) => ({\n        menuName: menu.name,\n        menu: await resolveUrlToFile(menu.fileUrl),\n      })),\n    )\n  : []',
      options: [2, { offsetTernaryExpressions: { AwaitExpression: true } }],
    },
    '[\n    foo ?\n        bar :\n        baz,\n    qux\n];',
    {
      code: 'foo();\n// Line\n/* multiline\n  Line */\nbar();\n// trailing comment',
      options: [2],
    },
    {
      code: 'switch (foo) {\n  case bar:\n    baz();\n    // call the baz function\n}',
      options: [2, { SwitchCase: 1 }],
    },
    {
      code: 'switch (foo) {\n  case bar:\n    baz();\n  // no default\n}',
      options: [2, { SwitchCase: 1 }],
    },
    '[\n    // no elements\n]',
    {
      code: 'var {\n  foo,\n  bar,\n  baz: qux,\n  foobar: baz = foobar\n} = qux;',
      options: [2],
    },
    {
      code: 'var [\n  foo,\n  bar,\n  baz,\n  foobar = baz\n] = qux;',
      options: [2],
    },
    {
      code: 'const {\n  a\n}\n=\n{\n  a: 1\n}',
      options: [2, { assignmentOperator: 0 }],
    },
    { code: 'const {\n  a\n} = {\n  a: 1\n}', options: [2] },
    { code: 'const\n  {\n    a\n  } = {\n    a: 1\n  };', options: [2] },
    { code: 'const\n  foo = {\n    bar: 1\n  }', options: [2] },
    { code: 'const [\n  a\n] = [\n  1\n]', options: [2] },
    {
      code: 'var folder = filePath\n    .foo()\n    .bar;',
      options: [2, { MemberExpression: 2 }],
    },
    { code: 'for (const foo of bar)\n  baz();', options: [2] },
    { code: 'var x = () =>\n  5;', options: [2] },
    '(\n    foo\n)(\n    bar\n)',
    '(() =>\n    foo\n)(\n    bar\n)',
    '(() => {\n    foo();\n})(\n    bar\n)',
    { code: '({code:\n  "foo.bar();"})', options: [2] },
    { code: '({code:\n"foo.bar();"})', options: [2] },
    '({\n    foo:\n        bar\n})',
    '({\n    [foo]:\n        bar\n})',
    {
      code: 'switch (foo) {\n  // comment\n  case study:\n    // comment\n    bar();\n  case closed:\n    /* multiline comment\n    */\n}',
      options: [2, { SwitchCase: 1 }],
    },
    {
      code: 'switch (foo) {\n  // comment\n  case study:\n  // the comment can also be here\n  case closed:\n}',
      options: [2, { SwitchCase: 1 }],
    },
    { code: 'foo && (\n    bar\n)', options: [4] },
    { code: 'foo && ((\n    bar\n))', options: [4] },
    { code: 'foo &&\n    (\n        bar\n    )', options: [4] },
    'foo &&\n    !bar(\n    )',
    'foo &&\n    ![].map(() => {\n        bar();\n    })',
    { code: 'foo =\n    bar;', options: [4] },
    {
      code: 'function foo() {\n  var bar = function(baz,\n        qux) {\n    foobar();\n  };\n}',
      options: [2, { FunctionExpression: { parameters: 3 } }],
    },
    'function foo() {\n    return (bar === 1 || bar === 2 &&\n        (/Function/.test(grandparent.type))) &&\n        directives(parent).indexOf(node) >= 0;\n}',
    {
      code: 'function foo() {\n    return (foo === bar || (\n        baz === qux && (\n            foo === foo ||\n            bar === bar ||\n            baz === baz\n        )\n    ))\n}',
      options: [4],
    },
    'if (\n    foo === 1 ||\n    bar === 1 ||\n    // comment\n    (baz === 1 && qux === 1)\n) {}',
    { code: 'foo =\n  (bar + baz);', options: [2] },
    {
      code: 'function foo() {\n  return (bar === 1 || bar === 2) &&\n    (z === 3 || z === 4);\n}',
      options: [2],
    },
    { code: '/* comment */ if (foo) {\n  bar();\n}', options: [2] },
    {
      code: 'if (foo) {\n  bar();\n// Otherwise, if foo is false, do baz.\n// baz is very important.\n} else {\n  baz();\n}',
      options: [2],
    },
    {
      code: 'function foo() {\n  return ((bar === 1 || bar === 2) &&\n    (z === 3 || z === 4));\n}',
      options: [2],
    },
    {
      code: 'foo(\n  bar,\n  baz,\n  qux\n);',
      options: [2, { CallExpression: { arguments: 1 } }],
    },
    {
      code: 'foo(\n\tbar,\n\tbaz,\n\tqux\n);',
      options: ['tab', { CallExpression: { arguments: 1 } }],
    },
    {
      code: 'foo(bar,\n        baz,\n        qux);',
      options: [4, { CallExpression: { arguments: 2 } }],
    },
    {
      code: 'foo(\nbar,\nbaz,\nqux\n);',
      options: [2, { CallExpression: { arguments: 0 } }],
    },
    {
      code: 'foo(bar,\n    baz,\n    qux\n);',
      options: [2, { CallExpression: { arguments: 'first' } }],
    },
    {
      code: 'foo(bar, baz,\n    qux, barbaz,\n    barqux, bazqux);',
      options: [2, { CallExpression: { arguments: 'first' } }],
    },
    {
      code: "foo(bar,\n        1 + 2,\n        !baz,\n        new Car('!')\n);",
      options: [2, { CallExpression: { arguments: 4 } }],
    },
    'foo(\n    (bar)\n);',
    {
      code: 'foo(\n    (bar)\n);',
      options: [4, { CallExpression: { arguments: 1 } }],
    },
    {
      code: 'var foo = function() {\n  return bar(\n    [{\n    }].concat(baz)\n  );\n};',
      options: [2],
    },
    'return (\n    foo\n);',
    'return (\n    foo\n)',
    'var foo = [\n    bar,\n    baz\n]',
    'var foo = [bar,\n    baz,\n    qux\n]',
    {
      code: 'var foo = [bar,\nbaz,\nqux\n]',
      options: [2, { ArrayExpression: 0 }],
    },
    {
      code: 'var foo = [bar,\n                baz,\n                qux\n]',
      options: [2, { ArrayExpression: 8 }],
    },
    {
      code: 'var foo = [bar,\n           baz,\n           qux\n]',
      options: [2, { ArrayExpression: 'first' }],
    },
    {
      code: 'var foo = [bar,\n           baz, qux\n]',
      options: [2, { ArrayExpression: 'first' }],
    },
    {
      code: 'var foo = [\n        { bar: 1,\n          baz: 2 },\n        { bar: 3,\n          baz: 4 }\n]',
      options: [4, { ArrayExpression: 2, ObjectExpression: 'first' }],
    },
    {
      code: 'var foo = {\nbar: 1,\nbaz: 2\n};',
      options: [2, { ObjectExpression: 0 }],
    },
    {
      code: 'var foo = { foo: 1, bar: 2,\n            baz: 3 }',
      options: [2, { ObjectExpression: 'first' }],
    },
    {
      code: 'var foo = [\n        {\n            foo: 1\n        }\n]',
      options: [4, { ArrayExpression: 2 }],
    },
    {
      code: 'function foo() {\n  [\n          foo\n  ]\n}',
      options: [2, { ArrayExpression: 4 }],
    },
    { code: '[\n]', options: [2, { ArrayExpression: 'first' }] },
    { code: '[\n]', options: [2, { ArrayExpression: 1 }] },
    { code: '{\n}', options: [2, { ObjectExpression: 'first' }] },
    { code: '{\n}', options: [2, { ObjectExpression: 1 }] },
    {
      code: 'var foo = [\n  [\n    1\n  ]\n]',
      options: [2, { ArrayExpression: 'first' }],
    },
    {
      code: 'var foo = [ 1,\n            [\n              2\n            ]\n];',
      options: [2, { ArrayExpression: 'first' }],
    },
    {
      code: 'var foo = bar(1,\n              [ 2,\n                3\n              ]\n);',
      options: [
        4,
        { ArrayExpression: 'first', CallExpression: { arguments: 'first' } },
      ],
    },
    {
      code: 'var foo =\n    [\n    ]()',
      options: [
        4,
        { CallExpression: { arguments: 'first' }, ArrayExpression: 'first' },
      ],
    },
    {
      code: 'const lambda = foo => {\n  Object.assign({},\n    filterName,\n    {\n      display\n    }\n  );\n}',
      options: [2, { ObjectExpression: 1 }],
    },
    {
      code: 'const lambda = foo => {\n  Object.assign({},\n    filterName,\n    {\n      display\n    }\n  );\n}',
      options: [2, { ObjectExpression: 'first' }],
    },
    {
      code: "var foo = function() {\n\twindow.foo('foo',\n\t\t{\n\t\t\tfoo: 'bar',\n\t\t\tbar: {\n\t\t\t\tfoo: 'bar'\n\t\t\t}\n\t\t}\n\t);\n}",
      options: ['tab'],
    },
    {
      code: "echo = spawn('cmd.exe',\n             ['foo', 'bar',\n              'baz']);",
      options: [
        2,
        { ArrayExpression: 'first', CallExpression: { arguments: 'first' } },
      ],
    },
    {
      code: 'if (foo)\n  bar();\n// Otherwise, if foo is false, do baz.\n// baz is very important.\nelse {\n  baz();\n}',
      options: [2],
    },
    {
      code: 'if (\n    foo && bar ||\n    baz && qux // This line is ignored because BinaryExpressions are not checked.\n) {\n    qux();\n}',
      options: [4],
    },
    '[\n] || [\n]',
    '(\n    [\n    ] || [\n    ]\n)',
    '1\n+ (\n    1\n)',
    '(\n    foo && (\n        bar ||\n        baz\n    )\n)',
    'foo\n    || (\n        bar\n    )',
    'foo\n                || (\n                    bar\n                )',
    {
      code: 'var foo =\n        1;',
      options: [4, { VariableDeclarator: 2, assignmentOperator: 'off' }],
    },
    {
      code: 'var foo = 1,\n    bar =\n    2;',
      options: [4, { assignmentOperator: 0 }],
    },
    {
      code: 'switch (foo) {\n  case bar:\n  {\n    baz();\n  }\n}',
      options: [2, { SwitchCase: 1 }],
    },
    { code: '`foo${\n  bar}`', options: [2] },
    { code: '`foo${\n  `bar${\n    baz}`}`', options: [2] },
    { code: '`foo${\n  `bar${\n    baz\n  }`\n}`', options: [2] },
    { code: '`foo${\n  (\n    bar\n  )\n}`', options: [2] },
    'foo(`\n    bar\n`, {\n    baz: 1\n});',
    'function foo() {\n    `foo${bar}baz${\n        qux}foo${\n        bar}baz`\n}',
    'JSON\n    .stringify(\n        {\n            ok: true\n        }\n    );',
    {
      code: 'foo =\n    bar =\n    baz;',
      options: [2, { assignmentOperator: 'off' }],
    },
    {
      code: 'foo =\nbar =\n    baz;',
      options: [4, { assignmentOperator: 'off' }],
    },
    "function foo() {\n    const template = `this indentation is not checked\nbecause it's part of a template literal.`;\n}",
    'function foo() {\n    const template = `the indentation of a ${\n        node.type\n    } node is checked.`;\n}',
    {
      code: "JSON\n    .stringify(\n        {\n            test: 'test'\n        }\n    );",
      options: [4, { CallExpression: { arguments: 1 } }],
    },
    '[\n    foo,\n    // comment\n    // another comment\n    bar\n]',
    'if (foo) {\n    /* comment */ bar();\n}',
    'function foo() {\n    return (\n        1\n    );\n}',
    'function foo() {\n    return (\n        1\n    )\n}',
    'if (\n    foo &&\n    !(\n        bar\n    )\n) {}',
    { code: "var abc = [\n  (\n    ''\n  ),\n  def,\n]", options: [2] },
    {
      code: "var abc = [\n  (\n    ''\n  ),\n  (\n    'bar'\n  )\n]",
      options: [2],
    },
    "function f() {\n    return asyncCall()\n        .then(\n            'some string',\n            [\n                1,\n                2,\n                3\n            ]\n        );\n}",
    {
      code: "function f() {\n    return asyncCall()\n        .then(\n            'some string',\n            [\n                1,\n                2,\n                3\n            ]\n        );\n}",
      options: [4, { MemberExpression: 1 }],
    },
    'var x = [\n    [1],\n    [2]\n]',
    'var y = [\n    {a: 1},\n    {b: 2}\n]',
    'foo(\n)',
    {
      code: 'foo(\n    bar,\n    {\n        baz: 1\n    }\n)',
      options: [4, { CallExpression: { arguments: 'first' } }],
    },
    'new Foo',
    'new (Foo)',
    'if (Foo) {\n    new Foo\n}',
    'var foo = 0, bar = 0, baz = 0;\nexport {\n    foo,\n    bar,\n    baz\n}',
    {
      code: 'foo\n    ? bar\n    : baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo ?\n    bar :\n    baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo ?\n    bar\n    : baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo\n    ? bar :\n    baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo\n    ? bar\n    : baz\n        ? qux\n        : foobar\n            ? boop\n            : beep',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo ?\n    bar :\n    baz ?\n        qux :\n        foobar ?\n            boop :\n            beep',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'var a =\n    foo ? bar :\n    baz ? qux :\n    foobar ? boop :\n    /*else*/ beep',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'var a = foo\n    ? bar\n    : baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'var a =\n    foo\n        ? bar\n        : baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'a =\n    foo ? bar :\n    baz ? qux :\n    foobar ? boop :\n    /*else*/ beep',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'a = foo\n    ? bar\n    : baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'a =\n    foo\n        ? bar\n        : baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo(\n    foo ? bar :\n    baz ? qux :\n    foobar ? boop :\n    /*else*/ beep\n)',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'function wrap() {\n    return (\n        foo ? bar :\n        baz ? qux :\n        foobar ? boop :\n        /*else*/ beep\n    )\n}',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'function wrap() {\n    return foo\n        ? bar\n        : baz\n}',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'function wrap() {\n    return (\n        foo\n            ? bar\n            : baz\n    )\n}',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo(\n    foo\n        ? bar\n        : baz\n)',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo(foo\n    ? bar\n    : baz\n)',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo\n    ? bar\n    : baz\n        ? qux\n        : foobar\n            ? boop\n            : beep',
      options: [4, { flatTernaryExpressions: false }],
    },
    {
      code: 'foo ?\n    bar :\n    baz ?\n        qux :\n        foobar ?\n            boop :\n            beep',
      options: [4, { flatTernaryExpressions: false }],
    },
    { code: '[,]', options: [2, { ArrayExpression: 'first' }] },
    { code: '[,]', options: [2, { ArrayExpression: 'off' }] },
    {
      code: '[\n    ,\n    foo\n]',
      options: [4, { ArrayExpression: 'first' }],
    },
    { code: '[sparse, , array];', options: [2, { ArrayExpression: 'first' }] },
    {
      code: "foo.bar('baz', function(err) {\n  qux;\n});",
      options: [2, { CallExpression: { arguments: 'first' } }],
    },
    {
      code: 'foo.bar(function() {\n  cookies;\n}).baz(function() {\n  cookies;\n});',
      options: [2, { MemberExpression: 1 }],
    },
    {
      code: 'foo.bar().baz(function() {\n  cookies;\n}).qux(function() {\n  cookies;\n});',
      options: [2, { MemberExpression: 1 }],
    },
    {
      code: '(\n  {\n    foo: 1,\n    baz: 2\n  }\n);',
      options: [2, { ObjectExpression: 'first' }],
    },
    {
      code: 'foo(() => {\n    bar;\n}, () => {\n    baz;\n})',
      options: [4, { CallExpression: { arguments: 'first' } }],
    },
    {
      code: '[ foo,\n  bar ].forEach(function() {\n  baz;\n})',
      options: [2, { ArrayExpression: 'first', MemberExpression: 1 }],
    },
    'foo = bar[\n    baz\n];',
    { code: 'foo[\n    bar\n];', options: [4, { MemberExpression: 1 }] },
    {
      code: 'foo[\n    (\n        bar\n    )\n];',
      options: [4, { MemberExpression: 1 }],
    },
    'if (foo)\n    bar;\nelse if (baz)\n    qux;',
    'if (foo) bar()\n\n; [1, 2, 3].map(baz)',
    'if (foo)\n    ;',
    'x => {}',
    "import {foo}\n    from 'bar';",
    "import 'foo'",
    {
      code: "import { foo,\n    bar,\n    baz,\n} from 'qux';",
      options: [4, { ImportDeclaration: 1 }],
    },
    {
      code: "import {\n    foo,\n    bar,\n    baz,\n} from 'qux';",
      options: [4, { ImportDeclaration: 1 }],
    },
    {
      code: "import { apple as a,\n         banana as b } from 'fruits';\nimport { cat } from 'animals';",
      options: [4, { ImportDeclaration: 'first' }],
    },
    {
      code: "import { declaration,\n                 can,\n                  be,\n              turned } from 'off';",
      options: [4, { ImportDeclaration: 'off' }],
    },
    '(\n    a\n) => b => {\n    c\n}',
    '(\n    a\n) => b => c => d => {\n    e\n}',
    '(\n    a\n) =>\n    (\n        b\n    ) => {\n        c\n    }',
    'if (\n    foo\n) bar(\n    baz\n);',
    'if (foo)\n{\n    bar();\n}',
    'function foo(bar)\n{\n    baz();\n}',
    '() =>\n    ({})',
    '() =>\n    (({}))',
    '(\n    () =>\n        ({})\n)',
    'var x = function foop(bar)\n{\n    baz();\n}',
    'var x = (bar) =>\n{\n    baz();\n}',
    'class Foo\n{\n    constructor()\n    {\n        foo();\n    }\n\n    bar()\n    {\n        baz();\n    }\n}',
    'class Foo\n    extends Bar\n{\n    constructor()\n    {\n        foo();\n    }\n\n    bar()\n    {\n        baz();\n    }\n}',
    '(\n    class Foo\n    {\n        constructor()\n        {\n            foo();\n        }\n\n        bar()\n        {\n            baz();\n        }\n    }\n)',
    {
      code: 'switch (foo)\n{\n    case 1:\n        bar();\n}',
      options: [4, { SwitchCase: 1 }],
    },
    'foo\n    .bar(function() {\n        baz\n    })',
    {
      code: 'foo\n        .bar(function() {\n            baz\n        })',
      options: [4, { MemberExpression: 2 }],
    },
    'foo\n    [bar](function() {\n        baz\n    })',
    'foo.\n    bar.\n    baz',
    {
      code: 'foo\n    .bar(function() {\n        baz\n    })',
      options: [4, { MemberExpression: 'off' }],
    },
    {
      code: 'foo\n                .bar(function() {\n                    baz\n                })',
      options: [4, { MemberExpression: 'off' }],
    },
    {
      code: 'foo\n                [bar](function() {\n                    baz\n                })',
      options: [4, { MemberExpression: 'off' }],
    },
    {
      code: 'foo.\n        bar.\n                    baz',
      options: [4, { MemberExpression: 'off' }],
    },
    {
      code: 'foo = bar(\n).baz(\n)',
      options: [4, { MemberExpression: 'off' }],
    },
    {
      code: 'foo[\n    bar ? baz :\n    qux\n]',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'function foo() {\n    return foo ? bar :\n        baz\n}',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'throw foo ? bar :\n    baz',
      options: [4, { flatTernaryExpressions: true }],
    },
    {
      code: 'foo(\n    bar\n) ? baz :\n    qux',
      options: [4, { flatTernaryExpressions: true }],
    },
    'foo\n    [\n        bar\n    ]\n    .baz(function() {\n        quz();\n    })',
    '[\n    foo\n][\n    "map"](function() {\n    qux();\n})',
    '(\n    a.b(function() {\n        c;\n    })\n)',
    '(\n    foo\n).bar(function() {\n    baz();\n})',
    'new Foo(\n    bar\n        .baz\n        .qux\n)',
    "const foo = a.b(),\n    longName =\n        (baz(\n            'bar',\n            'bar'\n        ));",
    {
      code: "const foo = a.b(),\n    longName =\n    (baz(\n        'bar',\n        'bar'\n    ));",
      options: [4, { assignmentOperator: 0 }],
    },
    "const foo = a.b(),\n    longName =\n        baz(\n            'bar',\n            'bar'\n        );",
    {
      code: "const foo = a.b(),\n    longName =\n    baz(\n        'bar',\n        'bar'\n    );",
      options: [4, { assignmentOperator: 0 }],
    },
    "const foo = a.b(),\n    longName\n        = baz(\n            'bar',\n            'bar'\n        );",
    {
      code: "const foo = a.b(),\n    longName\n    = baz(\n        'bar',\n        'bar'\n    );",
      options: [4, { assignmentOperator: 0 }],
    },
    "const foo = a.b(),\n    longName =\n        ('fff');",
    {
      code: "const foo = a.b(),\n    longName =\n    ('fff');",
      options: [4, { assignmentOperator: 0 }],
    },
    "const foo = a.b(),\n    longName\n        = ('fff');",
    {
      code: "const foo = a.b(),\n    longName\n    = ('fff');",
      options: [4, { assignmentOperator: 0 }],
    },
    "const foo = a.b(),\n    longName =\n        (\n            'fff'\n        );",
    {
      code: "const foo = a.b(),\n    longName =\n    (\n        'fff'\n    );",
      options: [4, { assignmentOperator: 0 }],
    },
    "const foo = a.b(),\n    longName\n        =(\n            'fff'\n        );",
    {
      code: "const foo = a.b(),\n    longName\n    =(\n        'fff'\n    );",
      options: [4, { assignmentOperator: 0 }],
    },
    {
      code: "type httpMethod = 'GET'\n  | 'POST'\n  | 'PUT';",
      options: [2, { VariableDeclarator: 0 }],
    },
    {
      code: "type httpMethod = 'GET'\n| 'POST'\n| 'PUT';",
      options: [2, { VariableDeclarator: 1 }],
    },
    'foo(`foo\n        `, {\n    ok: true\n},\n{\n    ok: false\n})',
    'foo(tag`foo\n        `, {\n    ok: true\n},\n{\n    ok: false\n}\n)',
    'async function test() {\n    const {\n        foo,\n        bar,\n    } = await doSomethingAsync(\n        1,\n        2,\n        3,\n    );\n}',
    'function* test() {\n    const {\n        foo,\n        bar,\n    } = yield doSomethingAsync(\n        1,\n        2,\n        3,\n    );\n}',
    '({\n    a: b\n} = +foo(\n    bar\n));',
    'const {\n    foo,\n    bar,\n} = typeof foo(\n    1,\n    2,\n    3,\n);',
    'const {\n    foo,\n    bar,\n} = +(\n    foo\n);',
    '<Foo a="b" c="d"/>;',
    '<Foo\n    a="b"\n    c="d"\n/>;',
    'var foo = <Bar a="b" c="d"/>;',
    'var foo = <Bar\n    a="b"\n    c="d"\n/>;',
    'var foo = (<Bar\n    a="b"\n    c="d"\n/>);',
    'var foo = (\n    <Bar\n        a="b"\n        c="d"\n    />\n);',
    '<\n    Foo\n    a="b"\n    c="d"\n/>;',
    '<Foo\n    a="b"\n    c="d"/>;',
    '<\n    Foo\n    a="b"\n    c="d"/>;',
    '<a href="foo">bar</a>;',
    '<a href="foo">\n    bar\n</a>;',
    '<a\n    href="foo"\n>\n    bar\n</a>;',
    '<a\n    href="foo">\n    bar\n</a>;',
    '<\n    a\n    href="foo">\n    bar\n</a>;',
    '<a\n    href="foo">\n    bar\n</a\n>;',
    'var foo = <a href="bar">\n    baz\n</a>;',
    'var foo = <a\n    href="bar"\n>\n    baz\n</a>;',
    'var foo = <a\n    href="bar">\n    baz\n</a>;',
    'var foo = <\n    a\n    href="bar">\n    baz\n</a>;',
    'var foo = <a\n    href="bar">\n    baz\n</a\n>',
    'var foo = (<a\n    href="bar">\n    baz\n</a>);',
    'var foo = (\n    <a href="bar">baz</a>\n);',
    'var foo = (\n    <a href="bar">\n        baz\n    </a>\n);',
    'var foo = (\n    <a\n        href="bar">\n        baz\n    </a>\n);',
    'var foo = <a href="bar">baz</a>;',
    '<a>\n    {\n    }\n</a>',
    '<a>\n    {\n        foo\n    }\n</a>',
    'function foo() {\n    return (\n        <a>\n            {\n                b.forEach(() => {\n                    // comment\n                    a = c\n                        .d()\n                        .e();\n                })\n            }\n        </a>\n    );\n}',
    '<App foo\n/>',
    { code: '<App\n  foo\n/>', options: [2] },
    { code: '<App\nfoo\n/>', options: [0] },
    { code: '<App\n\tfoo\n/>', options: ['tab'] },
    '<App\n    foo\n/>',
    '<App\n    foo\n></App>',
    {
      code: "<App\n  foo={function() {\n    console.log('bar');\n  }}\n/>",
      options: [2],
    },
    {
      code: "<App foo={function() {\n  console.log('bar');\n}}\n/>",
      options: [2],
    },
    {
      code: "var x = function() {\n  return <App\n    foo={function() {\n      console.log('bar');\n    }}\n  />\n}",
      options: [2],
    },
    {
      code: "var x = <App\n  foo={function() {\n    console.log('bar');\n  }}\n/>",
      options: [2],
    },
    {
      code: "<Provider\n  store\n>\n  <App\n    foo={function() {\n      console.log('bar');\n    }}\n  />\n</Provider>",
      options: [2],
    },
    {
      code: "<Provider\n  store\n>\n  {baz && <App\n    foo={function() {\n      console.log('bar');\n    }}\n  />}\n</Provider>",
      options: [2],
    },
    { code: '<App\n\tfoo\n/>', options: ['tab'] },
    { code: '<App\n\tfoo\n></App>', options: ['tab'] },
    {
      code: "<App foo={function() {\n\tconsole.log('bar');\n}}\n/>",
      options: ['tab'],
    },
    {
      code: "var x = <App\n\tfoo={function() {\n\t\tconsole.log('bar');\n\t}}\n/>",
      options: ['tab'],
    },
    '<App\n    foo />',
    '<div>\n    unrelated{\n        foo\n    }\n</div>',
    '<div>unrelated{\n    foo\n}\n</div>',
    '<\n    input\n    type=\n        "number"\n/>',
    "<\n    input\n    type=\n        {'number'}\n/>",
    '<\n    input\n    type\n        ="number"\n/>',
    'foo ? (\n    bar\n) : (\n    baz\n)',
    'foo ? (\n    <div>\n    </div>\n) : (\n    <span>\n    </span>\n)',
    '<div>\n    {\n        /* foo */\n    }\n</div>',
    '<>\n    <A />\n</>',
    '<\n>\n    <A />\n</>',
    '<>\n    <A />\n</\n>',
    '<\n>\n    <A />\n</\n>',
    '< // Comment\n>\n    <A />\n</>',
    '<\n    // Comment\n>\n    <A />\n</>',
    '<\n// Comment\n>\n    <A />\n</>',
    '<>\n    <A />\n</ // Comment\n>',
    '<>\n    <A />\n</\n    // Comment\n>',
    '<>\n    <A />\n</\n// Comment\n>',
    '< /* Comment */\n>\n    <A />\n</>',
    '<\n    /* Comment */\n>\n    <A />\n</>',
    '<\n/* Comment */\n>\n    <A />\n</>',
    '<\n    /**                 * Comment\n     */\n>\n    <A />\n</>',
    '<\n/**             * Comment\n */\n>\n    <A />\n</>',
    '<>\n    <A />\n</ /* Comment */\n>',
    '<>\n    <A />\n</\n    /* Comment */ >',
    '<>\n    <A />\n</\n/* Comment */ >',
    '<>\n    <A />\n</\n    /* Comment */\n>',
    '<>\n    <A />\n</\n/* Comment */\n>',
    '<div>\n    {\n        (\n            1\n        )\n    }\n</div>',
    'function A() {\n    return (\n        <div>\n            {\n                b && (\n                    <div>\n                    </div>\n                )\n            }\n        </div>\n    );\n}',
    '<div>foo\n    <div>bar</div>\n</div>',
    '<small>Foo bar&nbsp;\n    <a>baz qux</a>.\n</small>',
    '<div\n    {...props}\n/>',
    '<div\n    {\n        ...props\n    }\n/>',
    {
      code: 'a(b\n  , c\n)',
      options: [2, { CallExpression: { arguments: 'off' } }],
    },
    {
      code: 'a(\n  new B({\n    c,\n  })\n);',
      options: [2, { CallExpression: { arguments: 'off' } }],
    },
    {
      code: 'foo\n? bar\n            : baz',
      options: [4, { ignoredNodes: ['ConditionalExpression'] }],
    },
    {
      code: 'class Foo {\nfoo() {\n    bar();\n}\n}',
      options: [4, { ignoredNodes: ['ClassBody'] }],
    },
    {
      code: 'class Foo {\nfoo() {\nbar();\n}\n}',
      options: [4, { ignoredNodes: ['ClassBody', 'BlockStatement'] }],
    },
    {
      code: 'foo({\n        bar: 1\n    },\n    {\n        baz: 2\n    },\n    {\n        qux: 3\n})',
      options: [4, { ignoredNodes: ['CallExpression > ObjectExpression'] }],
    },
    {
      code: 'foo\n                            .bar',
      options: [4, { ignoredNodes: ['MemberExpression'] }],
    },
    {
      code: '$(function() {\n\nfoo();\nbar();\n\n});',
      options: [
        4,
        {
          ignoredNodes: [
            "Program > ExpressionStatement > CallExpression[callee.name='$'] > FunctionExpression > BlockStatement",
          ],
        },
      ],
    },
    {
      code: '<Foo\n            bar="1" />',
      options: [4, { ignoredNodes: ['JSXOpeningElement'] }],
    },
    {
      code: 'foo &&\n<Bar\n>\n</Bar>',
      options: [4, { ignoredNodes: ['JSXElement', 'JSXOpeningElement'] }],
    },
    {
      code: '(function($) {\n$(function() {\n    foo;\n});\n}())',
      options: [
        4,
        {
          ignoredNodes: [
            'ExpressionStatement > CallExpression > FunctionExpression.callee > BlockStatement',
          ],
        },
      ],
    },
    {
      code: 'const value = (\n    condition ?\n    valueIfTrue :\n    valueIfFalse\n);',
      options: [4, { ignoredNodes: ['ConditionalExpression'] }],
    },
    {
      code: 'var a = 0, b = 0, c = 0;\nexport default foo(\n    a,\n    b, {\n    c\n    }\n)',
      options: [
        4,
        {
          ignoredNodes: [
            'ExportDefaultDeclaration > CallExpression > ObjectExpression',
          ],
        },
      ],
    },
    {
      code: 'foobar = baz\n       ? qux\n       : boop',
      options: [4, { ignoredNodes: ['ConditionalExpression'] }],
    },
    {
      code: '`\n    SELECT\n        ${\n            foo\n        } FROM THE_DATABASE\n`',
      options: [4, { ignoredNodes: ['TemplateLiteral'] }],
    },
    {
      code: "<foo\n    prop='bar'\n    >\n    Text\n</foo>",
      options: [4, { ignoredNodes: ['JSXOpeningElement'] }],
    },
    {
      code: 'var x = 1,\n    y = 2;\nvar z;',
      options: ['tab', { ignoredNodes: ['VariableDeclarator'] }],
    },
    {
      code: '[\n    foo(),\n    bar\n]',
      options: [
        'tab',
        { ArrayExpression: 'first', ignoredNodes: ['CallExpression'] },
      ],
    },
    {
      code: 'if (foo) {\n    doSomething();\n\n// Intentionally unindented comment\n    doSomethingElse();\n}',
      options: [4, { ignoreComments: true }],
    },
    {
      code: 'if (foo) {\n    doSomething();\n\n/* Intentionally unindented comment */\n    doSomethingElse();\n}',
      options: [4, { ignoreComments: true }],
    },
    'const obj = {\n    foo () {\n        return condition ? // comment\n            1 :\n            2\n    }\n}',
    'if (foo) {\n// Comment can align with code immediately above even if "incorrect" alignment\n    doSomething();\n}',
    'if (foo) {\n    doSomething();\n// Comment can align with code immediately below even if "incorrect" alignment\n}',
    'if (foo) {\n    // Comment can be in correct alignment even if not aligned with code above/below\n}',
    'if (foo) {\n\n    // Comment can be in correct alignment even if gaps between (and not aligned with) code above/below\n\n}',
    '[{\n    foo\n},\n\n// Comment between nodes\n\n{\n    bar\n}];',
    '[{\n    foo\n},\n\n// Comment between nodes\n\n{ // comment\n    bar\n}];',
    'let foo\n\n// comment\n\n;(async () => {})()',
    'let foo\n// comment\n\n;(async () => {})()',
    'let foo\n\n// comment\n;(async () => {})()',
    'let foo\n// comment\n;(async () => {})()',
    'let foo\n\n    /* comment */;\n\n(async () => {})()',
    'let foo\n    /* comment */;\n\n(async () => {})()',
    'let foo\n\n    /* comment */;\n(async () => {})()',
    'let foo\n    /* comment */;\n(async () => {})()',
    'let foo\n/* comment */;\n\n(async () => {})()',
    'let foo\n/* comment */;\n(async () => {})()',
    '// comment\n\n;(async () => {})()',
    '// comment\n;(async () => {})()',
    '{\n    let foo\n\n    // comment\n\n    ;(async () => {})()\n}',
    '{\n    let foo\n    // comment\n    ;(async () => {})()\n}',
    '{\n    // comment\n\n    ;(async () => {})()\n}',
    '{\n    // comment\n    ;(async () => {})()\n}',
    'const foo = 1\nconst bar = foo\n\n/* comment */\n\n;[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n/* comment */\n\n;[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n\n/* comment */\n;[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n/* comment */\n;[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n\n    /* comment */;\n\n[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n    /* comment */;\n\n[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n\n    /* comment */;\n[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n    /* comment */;\n[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n/* comment */;\n\n[1, 2, 3].forEach(() => {})',
    'const foo = 1\nconst bar = foo\n/* comment */;\n[1, 2, 3].forEach(() => {})',
    '/* comment */\n\n;[1, 2, 3].forEach(() => {})',
    '/* comment */\n;[1, 2, 3].forEach(() => {})',
    '{\n    const foo = 1\n    const bar = foo\n\n    /* comment */\n\n    ;[1, 2, 3].forEach(() => {})\n}',
    '{\n    const foo = 1\n    const bar = foo\n    /* comment */\n    ;[1, 2, 3].forEach(() => {})\n}',
    '{\n    /* comment */\n\n    ;[1, 2, 3].forEach(() => {})\n}',
    '{\n    /* comment */\n    ;[1, 2, 3].forEach(() => {})\n}',
    'import(\n    // before\n    source\n    // after\n)',
    'foo(() => {\n    tag`\n    multiline\n    template\n    literal\n    `(() => {\n        bar();\n    });\n});',
    '{\n    tag`\n    multiline\n    template\n    ${a} ${b}\n    literal\n    `(() => {\n        bar();\n    });\n}',
    'foo(() => {\n    tagOne`\n    multiline\n    template\n    literal\n    ${a} ${b}\n    `(() => {\n        tagTwo`\n        multiline\n        template\n        literal\n        `(() => {\n            bar();\n        });\n\n        baz();\n    });\n});',
    '{\n    tagOne`\n    ${a} ${b}\n    multiline\n    template\n    literal\n    `(() => {\n        tagTwo`\n        multiline\n        template\n        literal\n        `(() => {\n            bar();\n        });\n\n        baz();\n    });\n};',
    'tagOne`multiline\n        ${a} ${b}\n        template\n        literal\n        `(() => {\n    foo();\n\n    tagTwo`multiline\n            template\n            literal\n        `({\n        bar: 1,\n        baz: 2\n    });\n});',
    'tagOne`multiline\n    template\n    literal\n    ${a} ${b}`({\n    foo: 1,\n    bar: tagTwo`multiline\n        template\n        literal`(() => {\n\n        baz();\n    })\n});',
    'foo.bar` template literal `(() => {\n    baz();\n})',
    'foo.bar.baz` template literal `(() => {\n    baz();\n})',
    'foo\n    .bar` template\n        literal `(() => {\n        baz();\n    })',
    'foo\n    .bar\n    .baz` template\n        literal `(() => {\n        baz();\n    })',
    'foo.bar`\n    ${a} ${b}\n    `(() => {\n    baz();\n})',
    'foo.bar1.bar2`\n    ${a} ${b}\n    `(() => {\n    baz();\n})',
    'foo\n    .bar1\n    .bar2`\n    ${a} ${b}\n    `(() => {\n        baz();\n    })',
    'foo\n    .bar`\n    ${a} ${b}\n    `(() => {\n        baz();\n    })',
    {
      code: 'foo\n.test`\n    ${a} ${b}\n    `(() => {\n    baz();\n})',
      options: [4, { MemberExpression: 0 }],
    },
    {
      code: 'foo\n        .test`\n    ${a} ${b}\n    `(() => {\n            baz();\n        })',
      options: [4, { MemberExpression: 2 }],
    },
    {
      code: 'const foo = async (arg1,\n                   arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
    },
    {
      code: 'const foo = async /* some comments */(arg1,\n                                      arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
    },
    { code: 'const a = async\nb => {}', options: [2] },
    {
      code: 'const foo = (arg1,\n             arg2) => async (arr1,\n                             arr2) =>\n{\n  return arg1 + arg2;\n}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
    },
    {
      code: 'const foo = async (arg1,\n  arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [2],
    },
    {
      code: 'const foo = async /*comments*/(arg1,\n  arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [2],
    },
    {
      code: 'const foo = async (arg1,\n        arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 4 },
          FunctionExpression: { parameters: 4 },
        },
      ],
    },
    {
      code: 'const foo = (arg1,\n        arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 4 },
          FunctionExpression: { parameters: 4 },
        },
      ],
    },
    {
      code: 'async function fn(ar1,\n                  ar2){}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
    },
    {
      code: 'async function /* some comments */ fn(ar1,\n                                      ar2){}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
    },
    {
      code: 'async  /* some comments */  function fn(ar1,\n                                        ar2){}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
    },
    {
      code: 'class C {\n  static {\n    foo();\n    bar();\n  }\n}',
      options: [2],
    },
    {
      code: 'class C {\n    static {\n        foo();\n        bar();\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    static {\n            foo();\n            bar();\n    }\n}',
      options: [4, { StaticBlock: { body: 2 } }],
    },
    {
      code: 'class C {\n    static {\n    foo();\n    bar();\n    }\n}',
      options: [4, { StaticBlock: { body: 0 } }],
    },
    {
      code: 'class C {\n\tstatic {\n\t\tfoo();\n\t\tbar();\n\t}\n}',
      options: ['tab'],
    },
    {
      code: 'class C {\n\tstatic {\n\t\t\tfoo();\n\t\t\tbar();\n\t}\n}',
      options: ['tab', { StaticBlock: { body: 2 } }],
    },
    {
      code: 'class C {\n    static\n    {\n        foo();\n        bar();\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    static {\n        var x,\n            y;\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    static\n    {\n        var x,\n            y;\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    static {\n        if (foo) {\n            bar;\n        }\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    static {\n        {\n            bar;\n        }\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    static {}\n\n    static {\n    }\n\n    static\n    {\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n\n    static {\n        foo;\n    }\n\n    static {\n        bar;\n    }\n\n}',
      options: [4],
    },
    {
      code: 'class C {\n\n    x = 1;\n\n    static {\n        foo;\n    }\n\n    y = 2;\n\n}',
      options: [4],
    },
    {
      code: 'class C {\n\n    method1(param) {\n        foo;\n    }\n\n    static {\n        bar;\n    }\n\n    method2(param) {\n        foo;\n    }\n\n}',
      options: [4],
    },
    {
      code: 'function f() {\n    class C {\n        static {\n            foo();\n            bar();\n        }\n    }\n}',
      options: [4],
    },
    {
      code: 'class C {\n    method() {\n            foo;\n    }\n    static {\n            bar;\n    }\n}',
      options: [
        4,
        { FunctionExpression: { body: 2 }, StaticBlock: { body: 2 } },
      ],
    },
    {
      code: "if (2 > 1)\n\tconsole.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: ['tab'],
    },
    {
      code: "if (2 > 1)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    { code: 'if (foo) bar();\nbaz()', options: [4] },
    { code: 'if (foo) bar()\n;baz()', options: [4] },
    { code: 'if (foo)\n    bar();\nbaz();', options: [4] },
    { code: 'if (foo)\n    bar()\n; baz()', options: [4] },
    { code: 'if (foo)\n    bar()\n;baz()\nqux()', options: [4] },
    { code: 'if (foo)\n    bar()\n;else\n    baz()', options: [4] },
    { code: 'if (foo)\n    bar()\nelse\n    baz()\n;qux()', options: [4] },
    { code: 'if (foo)\n    if (bar)\n        baz()\n;qux()', options: [4] },
    {
      code: 'if (foo)\n    bar()\nelse if (baz)\n    qux()\n;quux()',
      options: [4],
    },
    {
      code: 'if (foo)\n    if (bar)\n        baz()\n    else\n        qux()\n;quux()',
      options: [4],
    },
    { code: 'if (foo)\n    bar()\n    ;\nbaz()', options: [4] },
    { code: 'if (foo)\n    ;\nbaz()', options: [4] },
    { code: 'if (foo)\n;baz()', options: [4] },
    { code: 'if (foo);\nelse\n    baz()', options: [4] },
    { code: 'if (foo)\n    ;\nelse\n    baz()', options: [4] },
    { code: 'if (foo)\n;else\n    baz()', options: [4] },
    { code: 'do foo();\nwhile (bar)', options: [4] },
    { code: 'do foo()\n;while (bar)', options: [4] },
    { code: 'do\n    foo();\nwhile (bar)', options: [4] },
    { code: 'do\n    foo()\n;while (bar)', options: [4] },
    { code: 'do;\nwhile (foo)', options: [4] },
    { code: 'do\n    ;\nwhile (foo)', options: [4] },
    { code: 'do\n;while (foo)', options: [4] },
    {
      code: "while (2 > 1)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    {
      code: "for (;;)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    {
      code: "for (a in b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    {
      code: "for (a of b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    {
      code: 'with (a)\n    console.log(b)\n;[1, 2, 3].forEach(x=>console.log(x))',
      options: [4],
    },
    {
      code: "label: for (a of b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    {
      code: "label:\nfor (a of b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
    },
    {
      code: 'if (foo)\n\tif (bar) doSomething();\n\telse doSomething();\nelse\n\tif (bar) doSomething();\n\telse doSomething();',
      options: ['tab'],
    },
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar) doSomething();\n    else doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar)\n        doSomething();\n    else doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar) doSomething();\n    else\n        doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar)\n        doSomething();\n    else\n        doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar) doSomething();\nelse doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar)\n    doSomething();\nelse doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar) doSomething();\nelse\n    doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar)\n    doSomething();\nelse\n    doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo)\n        if (bar) doSomething();\n        else doSomething();\n    else\n        if (bar) doSomething();\n        else doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo)\n        if (bar) doSomething();\n        else\n            if (bar) doSomething();\n            else doSomething();\n    else doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo) doSomething();\nelse doSomething();',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo) {\n    doSomething();\n}',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo)\n{\n    doSomething();\n}',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo) {\n        doSomething();\n    }',
    'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo)\n    {\n        doSomething();\n    }',
    'import json\n    from "./foo.json"\n    with {\n        type: "json"\n    };',
    'import json from "./foo.json"\n    with\n    {\n        type\n        :\n        "json"\n    };',
    'import "./foo.json"\n    with {\n        type: "json"\n    };',
    'import "./foo.json"\n    with\n    {\n        type\n        :\n        "json"\n    };',
    'import(\n    "./foo.json",\n    {\n        type: "json"\n    },\n)',
    'import(\n    "./foo.json"\n    ,\n    {\n        type\n        :\n        "json"\n    }\n    ,\n)',
    '// should correctly recognize a `form` token\nimport {\n    from\n    as\n    foo\n}\n    from "./foo"\n    with {\n        from: "foo"\n    };',

    // ---- from indent._ts_.test.ts ----
    "\n// ClassDeclaration\nabstract class Foo {\n    constructor() {}\n    method() {\n        console.log('hi');\n    }\n}",
    '\n// TSAbstractPropertyDefinition\nclass Foo {\n    abstract bar : baz;\n    abstract foo : {\n        a : number\n        b : number\n    };\n}',
    '\n// TSAbstractMethodDefinition\nclass Foo {\n    abstract bar() : baz;\n    abstract foo() : {\n        a : number\n        b : number\n    };\n}',
    '\n// TSArrayType\ntype foo = ArrType[];',
    '\n// TSAsExpression\nconst foo = {} as {\n    foo: string,\n    bar: number,\n};',
    '\n// TSAsExpression\nconst foo = {} as\n{\n    foo: string,\n    bar: number,\n};',
    '\n// TSConditionalType\ntype Foo<T> = T extends string\n    ? {\n        a: number,\n        b: boolean\n    }\n    : {\n        c: string\n    };',
    '\n// TSConditionalType\ntype Foo<T> = T extends string ? {\n    a: number,\n    b: boolean\n} : string;',
    '\n// TSConstructorType\ntype Constructor<T> = new (\n    ...args: any[]\n) => T;',
    '\n// TSConstructSignature\ninterface Foo {\n    new () : Foo\n    new () : {\n        bar : string\n        baz : string\n    }\n}',
    '\n// TSDeclareFunction\ndeclare function foo() : {\n    bar : number,\n    baz : string,\n};',
    '\n// TSEmptyBodyFunctionExpression\nclass Foo {\n    constructor(\n        a : string,\n        b : {\n            c : number\n        }\n    )\n}',
    '\n// TSEnumDeclaration, TSEnumMember\nenum Foo {\n    bar = 1,\n    baz = 1,\n}',
    '\n// TSEnumDeclaration, TSEnumMember\nenum Foo\n{\n    bar = 1,\n    baz = 1,\n}',
    '\n// TSExportAssignment\nexport = {\n    a: 1,\n    b: 2,\n}',
    '\n// TSFunctionType\nconst foo: () => void = () => ({\n    a: 1,\n    b: 2,\n});',
    '\n// TSFunctionType\nconst foo: () => {\n    a: number,\n    b: number,\n} = () => ({\n    a: 1,\n    b: 2,\n});',
    '\n// TSFunctionType\nconst foo: ({\n    a: number,\n    b: number,\n}) => void = (arg) => ({\n    a: 1,\n    b: 2,\n});',
    '\n// TSFunctionType\nconst foo: ({\n    a: number,\n    b: number,\n}) => {\n    a: number,\n    b: number,\n} = (arg) => ({\n    a: arg.a,\n    b: arg.b,\n});',
    '\n// TSImportType\nconst foo: import("bar") = {\n    a: 1,\n    b: 2,\n};',
    '\n// TSImportType\nconst foo: import(\n    "bar"\n) = {\n    a: 1,\n    b: 2,\n};',
    "\n// TSIndexedAccessType\ntype Foo = Bar[\n    'asdf'\n];",
    '\n// TSIndexSignature\ntype Foo = {\n    [a : string] : {\n        x : foo\n        [b : number] : boolean\n    }\n}',
    '\n// TSInferType\ntype Foo<T> = T extends string\n    ? infer U\n    : {\n        a : string\n    };',
    '\n// TSInterfaceBody, TSInterfaceDeclaration\ninterface Foo {\n    a : string\n    b : {\n        c : number\n        d : boolean\n    }\n}',
    '\n// TSInterfaceHeritage\ninterface Foo extends Bar {\n    a : string\n    b : {\n        c : number\n        d : boolean\n    }\n}',
    '\n// TSIntersectionType\ntype Foo = "string" & {\n    a : number\n} & number;',
    "\n// TSImportEqualsDeclaration, TSExternalModuleReference\nimport foo = require(\n    'asdf'\n);",
    '\n// TSMappedType\ntype Partial<T> = {\n    [P in keyof T];\n}',
    '\n// TSMappedType\ntype Partial<T> = {\n    [P in keyof T]: T[P];\n}',
    '\n// TSMappedType\ntype Partial<T> = {\n    readonly [P in keyof T]: T[P];\n}',
    '\n// TSMappedType\n// TSQuestionToken\ntype Partial<T> = {\n    [P in keyof T]?: T[P];\n}',
    '\n// TSMappedType\n// TSPlusToken\ntype Partial<T> = {\n    [P in keyof T]+?: T[P];\n}',
    '\n// TSMappedType\n// TSMinusToken\ntype Partial<T> = {\n    [P in keyof T]-?: T[P];\n}',
    '\n// TSMethodSignature\ninterface Foo {\n    method() : string\n    method2() : {\n        a : number\n        b : string\n    }\n}',
    '\n// TSModuleBlock, TSModuleDeclaration\ndeclare module "foo" {\n    export const bar : {\n        a : string,\n        b : number,\n    }\n}',
    '\n// TSNonNullExpression\nconst foo = a!\n    .b!.\n    c;',
    "\n// TSParameterProperty\nclass Foo {\n    constructor(\n        private foo : string,\n        public bar : {\n            a : string,\n            b : number,\n        }\n    ) {\n        console.log('foo')\n    }\n}",
    '\n// TSPropertySignature\ninterface Foo {\n    bar : string\n    baz : {\n        a : string\n        b : number\n    }\n}',
    '\n// TSQualifiedName\nconst a: Foo.bar = {\n    a: 1,\n    b: 2,\n};',
    '\n// TSQualifiedName\nconst a: Foo.\n    bar\n    .baz = {\n        a: 1,\n        b: 2,\n    };',
    '\n// TSRestType\ntype foo = [\n    string,\n    ...string[],\n];',
    '\n// TSThisType\ndeclare class MyArray<T> extends Array<T> {\n    sort(compareFn?: (a: T, b: T) => number): this;\n    meth() : {\n        a: number,\n    }\n}',
    '\n// TSTupleType\ntype foo = [\n    string,\n    number,\n];',
    '\n// TSTupleType\ntype foo = [\n    [\n        string,\n        number,\n    ],\n];',
    '\n// TSTypeOperator\ntype T = keyof {\n    a: 1,\n    b: 2,\n};',
    '\n// TSTypeParameter, TSTypeParameterDeclaration\ntype Foo<T> = {\n    a : unknown,\n    b : never,\n}',
    "\n// TSTypeParameter, TSTypeParameterDeclaration\nfunction foo<\n    T,\n    U\n>() {\n    console.log('');\n}",
    '\n// TSUnionType\ntype Foo = string | {\n    a : number\n} | number;',
    "@Component({\n    components: {\n        ErrorPage: () => import('@/components/ErrorPage.vue'),\n    },\n    head: {\n        titleTemplate(title) {\n            if (title) {\n                return `test`\n            }\n            return 'Title'\n        },\n        htmlAttrs: {\n            lang: 'en',\n        },\n        meta: [\n            { charset: 'utf-8' },\n            { name: 'viewport', content: 'width=device-width, initial-scale=1' },\n        ],\n    },\n})\nexport default class App extends Vue\n{\n    get error()\n    {\n        return this.$store.state.errorHandler.error\n    }\n}",
    '/**\n * @param {string} name\n * @param {number} age\n * @returns {string}\n */\nfunction foo(name: string, age: number): string {}',
    'const firebaseApp = firebase.apps.length\n    ? firebase.app()\n    : firebase.initializeApp({\n        apiKey: __FIREBASE_API_KEY__,\n        authDomain: __FIREBASE_AUTH_DOMAIN__,\n        databaseURL: __FIREBASE_DATABASE_URL__,\n        projectId: __FIREBASE_PROJECT_ID__,\n        storageBucket: __FIREBASE_STORAGE_BUCKET__,\n        messagingSenderId: __FIREBASE_MESSAGING_SENDER_ID__,\n    })',
    {
      code: 'const foo = {\n                a: 1,\n                b: 2\n            },\n            bar = 1;',
      options: [4, { VariableDeclarator: { const: 3 } }],
    },
    {
      code: 'const foo : Foo = {\n                a: 1,\n                b: 2\n            },\n            bar = 1;',
      options: [4, { VariableDeclarator: { const: 3 } }],
    },
    {
      code: 'const name: string = \'  Typescript  \'\n        .toUpperCase()\n        .trim(),\n\n      greeting: string = (" Hello " + name)\n        .toUpperCase()\n        .trim();',
      options: [2, { VariableDeclarator: { const: 3 } }],
    },
    {
      code: "const div: JQuery<HTMLElement> = $('<div>')\n        .addClass('some-class')\n        .appendTo($('body')),\n\n      button: JQuery<HTMLElement> = $('<button>')\n        .text('Cancel')\n        .appendTo(div);",
      options: [2, { VariableDeclarator: { const: 3 } }],
    },
    {
      code: '@Bar()\nexport class Foo {\n  @a\n  id: string;\n\n  @a @b()\n  age: number;\n\n  @a @b() username: string;\n}',
      options: [2],
    },
    {
      code: 'const map2 = Object.keys(map)\n  .filter((key) => true)\n  .reduce<Record<string, string>>((result, key) => {\n    result[key] = map[key];\n    return result;\n  }, {});',
      options: [2],
    },
    {
      code: 'class Some {\n  range: {\n    index?: number\n    length?: number\n  } = {\n    index: 0,\n    length: 0\n  }\n}',
      options: [2],
    },
    {
      code: 'const some: {\n  index?: number\n  length?: number\n} = {\n  index: 0,\n  length: 0\n}',
      options: [2],
    },
    {
      code: 'const some: {\n  index?: number\n  length?: number\n} = {\n  index: 0,\n  length: 0\n} as {\n  index?: number\n  length?: number\n}',
      options: [2],
    },
    'type Foo = string\ndeclare type Foo = number\nnamespace Foo {\n    type Bar = boolean\n}',
    {
      code: 'class Foo {\n  accessor [bar]: string\n  accessor baz: number\n}',
      options: [2],
    },
    {
      code: 'using a = foo(),\n  b = bar();\nawait using c = baz(),\n  d = qux();',
      options: [2, { VariableDeclarator: 1 }],
    },
    {
      code: 'using a = foo(),\n      b = bar();\nawait using c = baz(),\n            d = qux();',
      options: [2, { VariableDeclarator: { using: 'first' } }],
    },
    {
      code: 'async function foo(bar: number): Promise<\n  number\n> {\n  return 2;\n}',
      options: [2],
    },
    {
      code: 'async function foo(\n  bar: number,\n): Promise<\n  number\n> {\n  return 2;\n}',
      options: [2],
    },
    {
      code: 'function foo(bar: number): (\n  number\n) {\n  return 2;\n}',
      options: [2],
    },
    {
      code: 'function foo(\n  bar: number,\n): (\n  number\n) {\n  return 2;\n}',
      options: [2],
    },
    {
      code: 'const a = (\n  param: 2 | 3,\n): Promise<\n  (\n    2 | 3\n  )\n> => {\n  return Promise.resolve(param)\n}',
      options: [2],
    },
    'export function isAuthenticated(authResult: AuthenticationResult | null | undefined)\n    : authResult is SuccessAuthenticationResult {\n    return !! authResult && authResult.isAuthenticated();\n}',
    "type SomeType =\n  'one'\n  | 'two'\n  | 'four'\n;",
    'genericFunction<\n    () => void\n>(\n    () => {\n        console.log("Test");\n    }\n);',
    'genericFunction<() => void>(\n    () => {\n        console.log("Test");\n    }\n)',
    'function foo<\n    T\n        =\n            Foo\n>() {}',
    "import foo\n    =\n        require('source')",

    // ---- from indent._jsx_.test.ts (@typescript-eslint/parser variant) ----
    '<App></App>\n// features: [], parser: @typescript-eslint/parser, , , ',
    '<></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    '<App>\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
    '<>\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    {
      code: '<App>\n  <Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '<App>\n  <></>\n</App>\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '<>\n  <Foo />\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '<App>\n<Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: [0], ',
      options: [0],
    },
    {
      code: '<App>\n\t<Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
    },
    {
      code: 'function App() {\n  return <App>\n    <Foo />\n  </App>;\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function App() {\n  return <App>\n    <></>\n  </App>;\n}\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function App() {\n  return (<App>\n    <Foo />\n  </App>);\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function App() {\n  return (<App>\n    <></>\n  </App>);\n}\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function App() {\n  return (\n    <App>\n      <Foo />\n    </App>\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function App() {\n  return (\n    <App>\n      <></>\n    </App>\n  );\n}\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'it(\n  (\n    <div>\n      <span />\n    </div>\n  )\n)\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'it(\n  (\n    <div>\n      <></>\n    </div>\n  )\n)\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'it(\n  (<div>\n    <span />\n    <span />\n    <span />\n  </div>)\n)\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '(\n  <div>\n    <span />\n  </div>\n)\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '{\n  head.title &&\n  <h1>\n    {head.title}\n  </h1>\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '{\n  head.title &&\n    <>\n      {head.title}\n    </>\n}\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '{\n  head.title &&\n    <h1>\n      {head.title}\n    </h1>\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '{\n  head.title && (\n    <h1>\n      {head.title}\n    </h1>)\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '{\n  head.title && (\n    <h1>\n      {head.title}\n    </h1>\n  )\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '[\n  <div />,\n  <div />\n]\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '[\n  <></>,\n  <></>\n]\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    '<div>\n    {\n        [\n            <Foo />,\n            <Bar />\n        ]\n    }\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
    '<div>\n    {foo &&\n        [\n            <Foo />,\n            <Bar />\n        ]\n    }\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
    '<div>\n    {foo &&\n        [\n            <></>,\n            <></>\n        ]\n    }\n</div>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    '<div>\n    bar <div>\n        bar\n        bar {foo}\n        bar </div>\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
    '<>\n    bar <>\n        bar\n        bar {foo}\n        bar </>\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <Foo /> :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <></> :\n    <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <Foo />\n    : <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <></>\n    : <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <Foo />\n    :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <></>\n    :\n    <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    '{!foo ?\n    <Foo\n        onClick={this.onClick}\n    />\n    :\n    <Bar\n        onClick={this.onClick}\n    />\n}\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <Foo /> :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <></> :\n    <></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
    'foo ? <Foo />\n    : <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <></>\n    : <></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
    'foo ? <Foo />\n    :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <></>\n    :\n    <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <Foo />\n) :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <></>\n) :\n    <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <Foo />\n)\n    : <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <></>\n)\n    : <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <Foo />\n)\n    :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <></>\n)\n    :\n    <></>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <Foo /> : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <></> : (\n        <></>\n    )\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <Foo />\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <></>\n    : (\n        <></>\n    )\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <Foo />\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ?\n    <></>\n    : (\n        <></>\n    )\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <Foo />\n) : (\n    <Bar />\n)\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <></>\n) : (\n    <></>\n)\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <Foo />\n)\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <></>\n)\n    : (\n        <></>\n    )\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? (\n    <Foo />\n)\n    :\n    (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <Foo /> : (\n    <Bar />\n)\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <></> : (\n    <></>\n)\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? <Foo />\n    : (<Bar />)\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <></>\n    : (<></>)\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    'foo ? <Foo />\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
    'foo ? <></>\n    : (\n        <></>\n    )\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
    {
      code: '<span>\n  {condition ?\n    <Thing\n      foo={`bar`}\n    /> :\n    <Thing/>\n  }\n</span>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: '<span>\n  {condition ?\n    <Thing\n      foo={"bar"}\n    /> :\n    <Thing/>\n  }\n</span>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function foo() {\n  <span>\n    {condition ?\n      <Thing\n        foo={superFoo}\n      /> :\n      <Thing/>\n    }\n  </span>\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function foo() {\n  <span>\n    {condition ?\n      <Thing\n        foo={superFoo}\n      /> :\n      <></>\n    }\n  </span>\n}\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'class Test extends React.Component {\n  render() {\n    return (\n      <div>\n        <div />\n        <div />\n      </div>\n    );\n  }\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'class Test extends React.Component {\n  render() {\n    return (\n      <>\n        <></>\n        <></>\n      </>\n    );\n  }\n}\n// features: [fragment], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'const Component = () => (\n  <View\n    ListFooterComponent={(\n      <View\n        rowSpan={3}\n        placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n      />\n    )}\n  />\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'const Component = () => (\n  <View\n    ListFooterComponent={(\n      <View\n        rowSpan={3}\n        placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n      />\n    )}\n  />\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'const Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n\t\t\t/>\n\t\t)}\n\t/>\n);\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
    },
    {
      code: 'function Foo() {\n  return (\n    <input\n      type="radio"\n      defaultChecked\n    />\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function Foo() {\n  return (\n    <div>\n      {condition && (\n        <p>Bar</p>\n      )}\n    </div>\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    '<App>\n    text\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
    '<App>\n    text\n    text\n    text\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
    {
      code: '<App>\n\ttext\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
    },
    {
      code: '<App>\n\t{undefined}\n\t{null}\n\t{true}\n\t{false}\n\t{42}\n\t{NaN}\n\t{"foo"}\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
    },
    {
      code: 'function App() {\n  return (\n    <App />\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'function App() {\n  return <App>\n    <Foo />\n  </App>;\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'const myFunction = () => (\n  [\n    <Tag\n      {...properties}\n    />,\n    <Tag\n      {...properties}\n    />,\n    <Tag\n      {...properties}\n    />,\n  ]\n)\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: 'const Item = ({ id, name, onSelect }) => <div onClick={onSelect}>\n  {id}: {name}\n</div>;\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: "<a role={'button'}\n  className={`navbar-burger ${open ? 'is-active' : ''}`}\n  href={'#'}\n  aria-label={'menu'}\n  aria-expanded={false}\n  onClick={openMenu}>\n  <span aria-hidden={'true'}/>\n  <span aria-hidden={'true'}/>\n  <span aria-hidden={'true'}/>\n</a>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ",
      options: [2],
    },
    {
      code: "export default class App extends React.Component {\n  state = {\n    name: '',\n  }\n\n  componentDidMount() {\n    this.fetchName()\n      .then(name => {\n        this.setState({name})\n      });\n  }\n\n  fetchName = () => {\n    const url = 'https://api.github.com/users/job13er'\n    return fetch(url)\n      .then(resp => resp.json())\n      .then(json => json.name)\n  }\n\n  render() {\n    const {name} = this.state\n    return (\n      <h1>Hello, {name}</h1>\n    )\n  }\n}\n// features: [class fields], parser: @typescript-eslint/parser, , options: [2], ",
      options: [2],
    },
    {
      code: 'function test (foo) {\n  return foo != null\n    ? <div>foo</div>\n    : <div>bar</div>\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
    {
      code: "<>\n  <div\n    foo={\n      condition\n        ? [\n          'bar'\n        ]\n        : [\n          'baz',\n          'qux'\n        ]\n    }\n  />\n  <div\n    style={\n      true\n        ? {\n          color: 'red',\n        }\n        : {\n          height: 1,\n        }\n    }\n  />\n</>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ",
      options: [2],
    },
    {
      code: '<App\n  foo\n    =\n      "bar"\n/>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
  ],

  invalid: [
    // ---- from indent._js_.test.ts ----
    {
      code: 'var a = b;\nif (a) {\nb();\n}',
      output: 'var a = b;\nif (a) {\n  b();\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: "require('http').request({hostname: 'localhost',\n                  port: 80}, function(res) {\n    res.end();\n  });",
      output:
        "require('http').request({hostname: 'localhost',\n  port: 80}, function(res) {\n  res.end();\n});",
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 18 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (array.some(function(){\n  return true;\n})) {\na++; // ->\n  b++;\n    c++; // <-\n}',
      output:
        'if (array.some(function(){\n  return true;\n})) {\n  a++; // ->\n  b++;\n  c++; // <-\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (a){\n\tb=c;\n\t\tc=d;\ne=f;\n}',
      output: 'if (a){\n\tb=c;\n\tc=d;\n\te=f;\n}',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 2 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (a){\n    b=c;\n      c=d;\n e=f;\n}',
      output: 'if (a){\n    b=c;\n    c=d;\n    e=f;\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 1 },
          line: 4,
        },
      ],
    },
    {
      code: "if (a) {\n  var b = c;\n  var d = e\n    * f;\n    var e = f; // <-\n// ->\n  function g() {\n    if (h) {\n      var i = j;\n      } // <-\n    } // <-\n\n  while (k) l++;\n  while (m) {\n  n--; // ->\n    } // <-\n\n  do {\n    o = p +\n  q; // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n    o = p +\n    q;\n    } while(r); // <-\n\n  for (var s in t) {\n    u++;\n  }\n\n    for (;;) {\n      v++; // <-\n  }\n\n  if ( w ) {\n    x++;\n  } else if (y) {\n      z++; // <-\n    aa++;\n    } else { // <-\n  bb++; // ->\n} // ->\n}\n\n/**/var b; // NO ERROR: single line multi-line comments followed by code is OK\n/*\n *\n */ var b; // NO ERROR: multi-line comments followed by code is OK\n\nvar arr = [\n  a,\n  b,\n  c,\n  function (){\n    d\n    }, // <-\n  {},\n  {\n    a: b,\n    c: d,\n    d: e\n  },\n  [\n    f,\n    g,\n    h,\n    i\n  ],\n  [j]\n];\n\nvar obj = {\n  a: {\n    b: {\n      c: d,\n      e: f,\n      g: h +\n    i // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n    }\n  },\n  g: [\n    h,\n    i,\n    j,\n    k\n  ]\n};\n\nvar arrObject = {a:[\n  a,\n  b, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c\n]};\n\nvar objArray = [{\n  a: b,\n  b: c, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c: d\n}];\n\nvar arrArray = [[\n  a,\n  b, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c\n]];\n\nvar objObject = {a:{\n  a: b,\n  b: c, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c: d\n}};\n\n\nswitch (a) {\n  case 'a':\n  var a = 'b'; // ->\n    break;\n  case 'b':\n    var a = 'b';\n    break;\n  case 'c':\n      var a = 'b'; // <-\n    break;\n  case 'd':\n    var a = 'b';\n  break; // ->\n  case 'f':\n    var a = 'b';\n    break;\n  case 'g':     {\n    var a = 'b';\n    break;\n  }\n  case 'z':\n  default:\n      break; // <-\n}\n\na.b('hi')\n   .c(a.b()) // <-\n   .d(); // <-\n\nif ( a ) {\n  if ( b ) {\nd.e(f) // ->\n  .g() // ->\n  .h(); // ->\n\n    i.j(m)\n      .k() // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n      .l(); // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n\n      n.o(p) // <-\n        .q() // <-\n        .r(); // <-\n  }\n}\n\nvar a = b,\n  c = function () {\n  h = i; // ->\n    j = k;\n      l = m; // <-\n  },\n  e = {\n    f: g,\n    n: o,\n    p: q\n  },\n  r = [\n    s,\n    t,\n    u\n  ];\n\nvar a = function () {\nb = c; // ->\n  d = e;\n    f = g; // <-\n};\n\nfunction c(a, b) {\n  if (a || (a &&\n            b)) { // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n    return d;\n  }\n}\n\nif ( a\n  || b ) {\nvar x; // ->\n  var c,\n    d = function(a,\n                  b) { // <-\n    a; // ->\n      b;\n        c; // <-\n    }\n}\n\n\na({\n  d: 1\n});\n\na(\n1\n);\n\na(\n  b({\n    d: 1\n  })\n);\n\na(\n  b(\n    c({\n      d: 1,\n      e: 1,\n      f: 1\n    })\n  )\n);\n\na({ d: 1 });\n\naa(\n   b({ // NO ERROR: CallExpression args not linted by default\n    c: d, // ->\n     e: f,\n     f: g\n  }) // ->\n);\n\naaaaaa(\n  b,\n  c,\n  {\n    d: a\n  }\n);\n\na(b, c,\n  d, e,\n    f, g  // NO ERROR: alignment of arguments of callExpression not checked\n  );  // <-\n\na(\n  ); // <-\n\naaaaaa(\n  b,\n  c, {\n    d: a\n  }, {\n    e: f\n  }\n);\n\na.b()\n  .c(function(){\n    var a;\n  }).d.e;\n\nif (a == 'b') {\n  if (c && d) e = f\n  else g('h').i('j')\n}\n\na = function (b, c) {\n  return a(function () {\n    var d = e\n    var f = g\n    var h = i\n\n    if (!j) k('l', (m = n))\n    if (o) p\n    else if (q) r\n  })\n}\n\nvar a = function() {\n  \"b\"\n    .replace(/a/, \"a\")\n    .replace(/bc?/, function(e) {\n      return \"b\" + (e.f === 2 ? \"c\" : \"f\");\n    })\n    .replace(/d/, \"d\");\n};\n\n$(b)\n  .on('a', 'b', function() { $(c).e('f'); })\n  .on('g', 'h', function() { $(i).j('k'); });\n\na\n  .b('c',\n           'd'); // NO ERROR: CallExpression args not linted by default\n\na\n  .b('c', [ 'd', function(e) {\n    e++;\n  }]);\n\nvar a = function() {\n      a++;\n    b++; // <-\n        c++; // <-\n    },\n    b;\n\nvar b = [\n      a,\n      b,\n      c\n    ],\n    c;\n\nvar c = {\n      a: 1,\n      b: 2,\n      c: 3\n    },\n    d;\n\n// holes in arrays indentation\nx = [\n 1,\n 1,\n 1,\n 1,\n 1,\n 1,\n 1,\n 1,\n 1,\n 1\n];\n\ntry {\n  a++;\n    b++; // <-\nc++; // ->\n} catch (d) {\n  e++;\n    f++; // <-\ng++; // ->\n} finally {\n  h++;\n    i++; // <-\nj++; // ->\n}\n\nif (array.some(function(){\n  return true;\n})) {\na++; // ->\n  b++;\n    c++; // <-\n}\n\nvar a = b.c(function() {\n      d++;\n    }),\n    e;\n\nswitch (true) {\n  case (a\n  && b):\ncase (c // ->\n&& d):\n    case (e // <-\n    && f):\n  case (g\n&& h):\n      var i = j; // <-\n    var k = l;\n  var m = n; // ->\n}\n\nif (a) {\n  b();\n}\nelse {\nc(); // ->\n  d();\n    e(); // <-\n}\n\nif (a) b();\nelse {\nc(); // ->\n  d();\n    e(); // <-\n}\n\nif (a) {\n  b();\n} else c();\n\nif (a) {\n  b();\n}\nelse c();\n\na();\n\nif( \"very very long multi line\" +\n      \"with weird indentation\" ) {\n  b();\na(); // ->\n    c(); // <-\n}\n\na( \"very very long multi line\" +\n    \"with weird indentation\", function() {\n  b();\na(); // ->\n    c(); // <-\n    }); // <-\n\na = function(content, dom) {\n  b();\n    c(); // <-\nd(); // ->\n};\n\na = function(content, dom) {\n      b();\n        c(); // <-\n    d(); // ->\n    };\n\na = function(content, dom) {\n    b(); // ->\n    };\n\na = function(content, dom) {\nb(); // ->\n    };\n\na('This is a terribly long description youll ' +\n  'have to read', function () {\n    b(); // <-\n    c(); // <-\n  }); // <-\n\nif (\n  array.some(function(){\n    return true;\n  })\n) {\na++; // ->\n  b++;\n    c++; // <-\n}\n\nfunction c(d) {\n  return {\n    e: function(f, g) {\n    }\n  };\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1:\n      if (foo) {\n        return 5;\n      }\n  }\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1:\n      c;\n  }\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1: c;\n  }\n}\n\nfunction test() {\n  var a = 1;\n  {\n    a();\n  }\n}\n\n{\n  a();\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1:\n        { // <-\n      a(); // ->\n      }\n      break;\n    default:\n      {\n        b();\n        }\n  }\n}\n\nswitch (a) {\n  default:\n    if (b)\n      c();\n}\n\nfunction test(x) {\n  switch (x) {\n    case 1:\n      return function() {\n        var a = 5;\n        return a;\n      };\n  }\n}\n\nswitch (a) {\n  default:\n    if (b)\n      c();\n}\n",
      output:
        "if (a) {\n  var b = c;\n  var d = e\n    * f;\n  var e = f; // <-\n  // ->\n  function g() {\n    if (h) {\n      var i = j;\n    } // <-\n  } // <-\n\n  while (k) l++;\n  while (m) {\n    n--; // ->\n  } // <-\n\n  do {\n    o = p +\n  q; // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n    o = p +\n    q;\n  } while(r); // <-\n\n  for (var s in t) {\n    u++;\n  }\n\n  for (;;) {\n    v++; // <-\n  }\n\n  if ( w ) {\n    x++;\n  } else if (y) {\n    z++; // <-\n    aa++;\n  } else { // <-\n    bb++; // ->\n  } // ->\n}\n\n/**/var b; // NO ERROR: single line multi-line comments followed by code is OK\n/*\n *\n */ var b; // NO ERROR: multi-line comments followed by code is OK\n\nvar arr = [\n  a,\n  b,\n  c,\n  function (){\n    d\n  }, // <-\n  {},\n  {\n    a: b,\n    c: d,\n    d: e\n  },\n  [\n    f,\n    g,\n    h,\n    i\n  ],\n  [j]\n];\n\nvar obj = {\n  a: {\n    b: {\n      c: d,\n      e: f,\n      g: h +\n    i // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n    }\n  },\n  g: [\n    h,\n    i,\n    j,\n    k\n  ]\n};\n\nvar arrObject = {a:[\n  a,\n  b, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c\n]};\n\nvar objArray = [{\n  a: b,\n  b: c, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c: d\n}];\n\nvar arrArray = [[\n  a,\n  b, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c\n]];\n\nvar objObject = {a:{\n  a: b,\n  b: c, // NO ERROR: INDENT ONCE WHEN MULTIPLE INDENTED EXPRESSIONS ARE ON SAME LINE\n  c: d\n}};\n\n\nswitch (a) {\n  case 'a':\n    var a = 'b'; // ->\n    break;\n  case 'b':\n    var a = 'b';\n    break;\n  case 'c':\n    var a = 'b'; // <-\n    break;\n  case 'd':\n    var a = 'b';\n    break; // ->\n  case 'f':\n    var a = 'b';\n    break;\n  case 'g':     {\n    var a = 'b';\n    break;\n  }\n  case 'z':\n  default:\n    break; // <-\n}\n\na.b('hi')\n  .c(a.b()) // <-\n  .d(); // <-\n\nif ( a ) {\n  if ( b ) {\n    d.e(f) // ->\n      .g() // ->\n      .h(); // ->\n\n    i.j(m)\n      .k() // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n      .l(); // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n\n    n.o(p) // <-\n      .q() // <-\n      .r(); // <-\n  }\n}\n\nvar a = b,\n  c = function () {\n    h = i; // ->\n    j = k;\n    l = m; // <-\n  },\n  e = {\n    f: g,\n    n: o,\n    p: q\n  },\n  r = [\n    s,\n    t,\n    u\n  ];\n\nvar a = function () {\n  b = c; // ->\n  d = e;\n  f = g; // <-\n};\n\nfunction c(a, b) {\n  if (a || (a &&\n            b)) { // NO ERROR: DON'T VALIDATE MULTILINE STATEMENTS\n    return d;\n  }\n}\n\nif ( a\n  || b ) {\n  var x; // ->\n  var c,\n    d = function(a,\n      b) { // <-\n      a; // ->\n      b;\n      c; // <-\n    }\n}\n\n\na({\n  d: 1\n});\n\na(\n1\n);\n\na(\n  b({\n    d: 1\n  })\n);\n\na(\n  b(\n    c({\n      d: 1,\n      e: 1,\n      f: 1\n    })\n  )\n);\n\na({ d: 1 });\n\naa(\n   b({ // NO ERROR: CallExpression args not linted by default\n     c: d, // ->\n     e: f,\n     f: g\n   }) // ->\n);\n\naaaaaa(\n  b,\n  c,\n  {\n    d: a\n  }\n);\n\na(b, c,\n  d, e,\n    f, g  // NO ERROR: alignment of arguments of callExpression not checked\n);  // <-\n\na(\n); // <-\n\naaaaaa(\n  b,\n  c, {\n    d: a\n  }, {\n    e: f\n  }\n);\n\na.b()\n  .c(function(){\n    var a;\n  }).d.e;\n\nif (a == 'b') {\n  if (c && d) e = f\n  else g('h').i('j')\n}\n\na = function (b, c) {\n  return a(function () {\n    var d = e\n    var f = g\n    var h = i\n\n    if (!j) k('l', (m = n))\n    if (o) p\n    else if (q) r\n  })\n}\n\nvar a = function() {\n  \"b\"\n    .replace(/a/, \"a\")\n    .replace(/bc?/, function(e) {\n      return \"b\" + (e.f === 2 ? \"c\" : \"f\");\n    })\n    .replace(/d/, \"d\");\n};\n\n$(b)\n  .on('a', 'b', function() { $(c).e('f'); })\n  .on('g', 'h', function() { $(i).j('k'); });\n\na\n  .b('c',\n           'd'); // NO ERROR: CallExpression args not linted by default\n\na\n  .b('c', [ 'd', function(e) {\n    e++;\n  }]);\n\nvar a = function() {\n    a++;\n    b++; // <-\n    c++; // <-\n  },\n  b;\n\nvar b = [\n    a,\n    b,\n    c\n  ],\n  c;\n\nvar c = {\n    a: 1,\n    b: 2,\n    c: 3\n  },\n  d;\n\n// holes in arrays indentation\nx = [\n  1,\n  1,\n  1,\n  1,\n  1,\n  1,\n  1,\n  1,\n  1,\n  1\n];\n\ntry {\n  a++;\n  b++; // <-\n  c++; // ->\n} catch (d) {\n  e++;\n  f++; // <-\n  g++; // ->\n} finally {\n  h++;\n  i++; // <-\n  j++; // ->\n}\n\nif (array.some(function(){\n  return true;\n})) {\n  a++; // ->\n  b++;\n  c++; // <-\n}\n\nvar a = b.c(function() {\n    d++;\n  }),\n  e;\n\nswitch (true) {\n  case (a\n  && b):\n  case (c // ->\n&& d):\n  case (e // <-\n    && f):\n  case (g\n&& h):\n    var i = j; // <-\n    var k = l;\n    var m = n; // ->\n}\n\nif (a) {\n  b();\n}\nelse {\n  c(); // ->\n  d();\n  e(); // <-\n}\n\nif (a) b();\nelse {\n  c(); // ->\n  d();\n  e(); // <-\n}\n\nif (a) {\n  b();\n} else c();\n\nif (a) {\n  b();\n}\nelse c();\n\na();\n\nif( \"very very long multi line\" +\n      \"with weird indentation\" ) {\n  b();\n  a(); // ->\n  c(); // <-\n}\n\na( \"very very long multi line\" +\n    \"with weird indentation\", function() {\n  b();\n  a(); // ->\n  c(); // <-\n}); // <-\n\na = function(content, dom) {\n  b();\n  c(); // <-\n  d(); // ->\n};\n\na = function(content, dom) {\n  b();\n  c(); // <-\n  d(); // ->\n};\n\na = function(content, dom) {\n  b(); // ->\n};\n\na = function(content, dom) {\n  b(); // ->\n};\n\na('This is a terribly long description youll ' +\n  'have to read', function () {\n  b(); // <-\n  c(); // <-\n}); // <-\n\nif (\n  array.some(function(){\n    return true;\n  })\n) {\n  a++; // ->\n  b++;\n  c++; // <-\n}\n\nfunction c(d) {\n  return {\n    e: function(f, g) {\n    }\n  };\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1:\n      if (foo) {\n        return 5;\n      }\n  }\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1:\n      c;\n  }\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1: c;\n  }\n}\n\nfunction test() {\n  var a = 1;\n  {\n    a();\n  }\n}\n\n{\n  a();\n}\n\nfunction a(b) {\n  switch(x) {\n    case 1:\n      { // <-\n        a(); // ->\n      }\n      break;\n    default:\n    {\n      b();\n    }\n  }\n}\n\nswitch (a) {\n  default:\n    if (b)\n      c();\n}\n\nfunction test(x) {\n  switch (x) {\n    case 1:\n      return function() {\n        var a = 5;\n        return a;\n      };\n  }\n}\n\nswitch (a) {\n  default:\n    if (b)\n      c();\n}\n",
      options: [
        2,
        {
          SwitchCase: 1,
          MemberExpression: 1,
          CallExpression: { arguments: 'off' },
        },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 11,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 15,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 16,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 23,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 29,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 30,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 36,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 38,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 39,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 40,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 54,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 114,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 120,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 124,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 134,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 3 },
          line: 138,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 3 },
          line: 139,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 143,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 144,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 145,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 151,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 152,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 153,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 159,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 161,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 175,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 177,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 189,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 18 },
          line: 192,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 4 },
          line: 193,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 195,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '5 spaces', actual: 4 },
          line: 228,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '3 spaces', actual: 2 },
          line: 231,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 245,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 248,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 304,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 306,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 307,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 308,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 311,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 312,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 313,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 314,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 315,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 318,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 319,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 320,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 321,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 322,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 326,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 327,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 328,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 329,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 330,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 331,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 332,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 333,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 334,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 335,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 340,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 341,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 344,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 345,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 348,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 349,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 355,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 357,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 361,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 362,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 363,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 368,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 370,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 374,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 376,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 383,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 385,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 390,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 392,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 409,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 410,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 416,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 417,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 418,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 422,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 423,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 6 },
          line: 427,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 8 },
          line: 428,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 429,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 430,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 433,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 434,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 437,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 438,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 442,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 443,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 444,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 451,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 453,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 499,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
          line: 500,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 504,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 505,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 506,
        },
      ],
    },
    {
      code: 'switch(value){\n    case "1":\n        a();\n    break;\n    case "2":\n        a();\n    break;\n    default:\n        a();\n        break;\n}',
      output:
        'switch(value){\n    case "1":\n        a();\n        break;\n    case "2":\n        a();\n        break;\n    default:\n        a();\n        break;\n}',
      options: [4, { SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 7,
        },
      ],
    },
    {
      code: 'var x = 0 &&\n    {\n       a: 1,\n          b: 2\n    };',
      output: 'var x = 0 &&\n    {\n        a: 1,\n        b: 2\n    };',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 7 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 10 },
          line: 4,
        },
      ],
    },
    {
      code: 'switch(value){\n    case "1":\n        a();\n        break;\n    case "2":\n        a();\n        break;\n    default:\n    break;\n}',
      output:
        'switch(value){\n    case "1":\n        a();\n        break;\n    case "2":\n        a();\n        break;\n    default:\n        break;\n}',
      options: [4, { SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 9,
        },
      ],
    },
    {
      code: 'switch(value){\n    case "1":\n    case "2":\n        a();\n        break;\n    default:\n        break;\n}\nswitch(value){\n    case "1":\n    break;\n    case "2":\n        a();\n    break;\n    default:\n        a();\n    break;\n}',
      output:
        'switch(value){\n    case "1":\n    case "2":\n        a();\n        break;\n    default:\n        break;\n}\nswitch(value){\n    case "1":\n        break;\n    case "2":\n        a();\n        break;\n    default:\n        a();\n        break;\n}',
      options: [4, { SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 11,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 14,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 17,
        },
      ],
    },
    {
      code: 'switch(value){\ncase "1":\n        a();\n        break;\n    case "2":\n        break;\n    default:\n        break;\n}',
      output:
        'switch(value){\n    case "1":\n        a();\n        break;\n    case "2":\n        break;\n    default:\n        break;\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var obj = {foo: 1, bar: 2};\nwith (obj) {\nconsole.log(foo + bar);\n}',
      output:
        'var obj = {foo: 1, bar: 2};\nwith (obj) {\n    console.log(foo + bar);\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: "switch (a) {\ncase '1':\nb();\nbreak;\ndefault:\nc();\nbreak;\n}",
      output:
        "switch (a) {\n    case '1':\n        b();\n        break;\n    default:\n        c();\n        break;\n}",
      options: [4, { SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'var foo = function(){\n    foo\n          .bar\n}',
      output: 'var foo = function(){\n    foo\n        .bar\n}',
      options: [4, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 10 },
          line: 3,
        },
      ],
    },
    {
      code: '(\n    foo\n    .bar\n)',
      output: '(\n    foo\n        .bar\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'var foo = function(){\n    foo\n             .bar\n}',
      output: 'var foo = function(){\n    foo\n            .bar\n}',
      options: [4, { MemberExpression: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 13 },
          line: 3,
        },
      ],
    },
    {
      code: 'var foo = () => {\n    foo\n             .bar\n}',
      output: 'var foo = () => {\n    foo\n            .bar\n}',
      options: [4, { MemberExpression: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 13 },
          line: 3,
        },
      ],
    },
    {
      code: 'TestClass.prototype.method = function () {\n  return Promise.resolve(3)\n      .then(function (x) {\n      return x;\n    });\n};',
      output:
        'TestClass.prototype.method = function () {\n  return Promise.resolve(3)\n    .then(function (x) {\n      return x;\n    });\n};',
      options: [2, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: 'while (a)\nb();',
      output: 'while (a)\n    b();',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'lmn = [{\n        a: 1\n    },\n    {\n        b: 2\n    },\n    {\n        x: 2\n}];',
      output: 'lmn = [{\n    a: 1\n},\n{\n    b: 2\n},\n{\n    x: 2\n}];',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 8,
        },
      ],
    },
    {
      code: 'for (var foo = 1;\nfoo < 10;\nfoo++) {}',
      output: 'for (var foo = 1;\n    foo < 10;\n    foo++) {}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'for (\nvar foo = 1;\nfoo < 10;\nfoo++\n    ) {}',
      output: 'for (\n    var foo = 1;\n    foo < 10;\n    foo++\n) {}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'for (;;)\nb();',
      output: 'for (;;)\n    b();',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'for (a in x)\nb();',
      output: 'for (a in x)\n    b();',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'do\nb();\nwhile(true)',
      output: 'do\n    b();\nwhile(true)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'with(a)\nb();',
      output: 'with(a)\n    b();',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'if(true)\nb();',
      output: 'if(true)\n    b();',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var test = {\n      a: 1,\n    b: 2\n    };',
      output: 'var test = {\n  a: 1,\n  b: 2\n};',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 6 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'var a = function() {\n      a++;\n    b++;\n          c++;\n    },\n    b;',
      output:
        'var a = function() {\n        a++;\n        b++;\n        c++;\n    },\n    b;',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 10 },
          line: 4,
        },
      ],
    },
    {
      code: 'var a = 1,\nb = 2,\nc = 3;',
      output: 'var a = 1,\n    b = 2,\n    c = 3;',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '[a, b,\n    c].forEach((index) => {\n        index;\n    });',
      output: '[a, b,\n    c].forEach((index) => {\n    index;\n});',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: '[a, b,\nc].forEach(function(index){\n  return index;\n});',
      output: '[a, b,\n    c].forEach(function(index){\n    return index;\n});',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: '[a, b, c].forEach(function(index){\n  return index;\n});',
      output: '[a, b, c].forEach(function(index){\n    return index;\n});',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: '(foo)\n    .bar([\n    baz\n]);',
      output: '(foo)\n    .bar([\n        baz\n    ]);',
      options: [4, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: "var x = ['a',\n         'b',\n         'c'\n];",
      output: "var x = ['a',\n    'b',\n    'c'\n];",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 3,
        },
      ],
    },
    {
      code: "var x = [\n         'a',\n         'b',\n         'c'\n];",
      output: "var x = [\n    'a',\n    'b',\n    'c'\n];",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 4,
        },
      ],
    },
    {
      code: "var x = [\n         'a',\n         'b',\n         'c',\n'd'];",
      output: "var x = [\n    'a',\n    'b',\n    'c',\n    'd'];",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: "var x = [\n         'a',\n         'b',\n         'c'\n  ];",
      output: "var x = [\n    'a',\n    'b',\n    'c'\n];",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: '[[\n], function(\n        foo\n    ) {}\n]',
      output: '[[\n], function(\n    foo\n) {}\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: "define([\n    'foo'\n], function(\n        bar\n    ) {\n    baz;\n}\n)",
      output: "define([\n    'foo'\n], function(\n    bar\n) {\n    baz;\n}\n)",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: "while (1 < 2)\nconsole.log('foo')\n  console.log('bar')",
      output: "while (1 < 2)\n  console.log('foo')\nconsole.log('bar')",
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: "function salutation () {\n  switch (1) {\n  case 0: return console.log('hi')\n    case 1: return console.log('hey')\n  }\n}",
      output:
        "function salutation () {\n  switch (1) {\n    case 0: return console.log('hi')\n    case 1: return console.log('hey')\n  }\n}",
      options: [2, { SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'var geometry, box, face1, face2, colorT, colorB, sprite, padding, maxWidth,\nheight, rotate;',
      output:
        'var geometry, box, face1, face2, colorT, colorB, sprite, padding, maxWidth,\n  height, rotate;',
      options: [2, { SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: "switch (a) {\ncase '1':\nb();\nbreak;\ndefault:\nc();\nbreak;\n}",
      output:
        "switch (a) {\n        case '1':\n            b();\n            break;\n        default:\n            c();\n            break;\n}",
      options: [4, { SwitchCase: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'var geometry,\nrotate;',
      output: 'var geometry,\n  rotate;',
      options: [2, { VariableDeclarator: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var geometry,\n  rotate;',
      output: 'var geometry,\n    rotate;',
      options: [2, { VariableDeclarator: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: 'var geometry,\n\trotate;',
      output: 'var geometry,\n\t\trotate;',
      options: ['tab', { VariableDeclarator: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 tabs', actual: 1 },
          line: 2,
        },
      ],
    },
    {
      code: 'let geometry,\n  rotate;',
      output: 'let geometry,\n    rotate;',
      options: [2, { VariableDeclarator: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: "let foo = 'foo',\n  bar = bar;\nconst a = 'a',\n  b = 'b';",
      output:
        "let foo = 'foo',\n    bar = bar;\nconst a = 'a',\n      b = 'b';",
      options: [2, { VariableDeclarator: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'var abc =\n  {\n    a: 1,\n     b: 2\n  };',
      output: 'var abc =\n  {\n    a: 1,\n    b: 2\n  };',
      options: [2, { VariableDeclarator: 'first', SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 5 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = {\n      bar: 1,\n      baz: {\n        qux: 2\n      }\n    },\n  bar = 1;',
      output:
        'var foo = {\n      bar: 1,\n      baz: {\n        qux: 2\n      }\n    },\n    bar = 1;',
      options: [2, { VariableDeclarator: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 7,
        },
      ],
    },
    {
      code: 'var a = 1,\n    B = class {\n    constructor(){}\n      a(){}\n      get b(){}\n    };',
      output:
        'var a = 1,\n    B = class {\n      constructor(){}\n      a(){}\n      get b(){}\n    };',
      options: [2, { VariableDeclarator: 'first', SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "var foo = 'foo',\n  bar = bar;",
      output: "var foo = 'foo',\n    bar = bar;",
      options: [2, { VariableDeclarator: { var: 'first' } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: 'if(true)\n  if (true)\n    if (true)\n    console.log(val);',
      output: 'if(true)\n  if (true)\n    if (true)\n      console.log(val);',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'var a = {\n    a: 1,\n    b: 2\n}',
      output: 'var a = {\n  a: 1,\n  b: 2\n}',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'var a = [\n    a,\n    b\n]',
      output: 'var a = [\n  a,\n  b\n]',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'let a = [\n    a,\n    b\n]',
      output: 'let a = [\n  a,\n  b\n]',
      options: [2, { VariableDeclarator: { let: 2 }, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'var a = new Test({\n      a: 1\n  }),\n    b = 4;',
      output: 'var a = new Test({\n        a: 1\n    }),\n    b = 4;',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'var a = new Test({\n      a: 1\n    }),\n    b = 4;\nconst c = new Test({\n      a: 1\n    }),\n    d = 4;',
      output:
        'var a = new Test({\n      a: 1\n    }),\n    b = 4;\nconst c = new Test({\n    a: 1\n  }),\n  d = 4;',
      options: [2, { VariableDeclarator: { var: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 8,
        },
      ],
    },
    {
      code: 'var abc = 5,\n    c = 2,\n    xyz =\n    {\n      a: 1,\n       b: 2\n    };',
      output:
        'var abc = 5,\n    c = 2,\n    xyz =\n    {\n      a: 1,\n      b: 2\n    };',
      options: [
        2,
        { VariableDeclarator: 2, SwitchCase: 1, assignmentOperator: 0 },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 7 },
          line: 6,
        },
      ],
    },
    {
      code: 'var abc =\n     {\n       a: 1,\n        b: 2\n     };',
      output: 'var abc =\n     {\n       a: 1,\n       b: 2\n     };',
      options: [
        2,
        { VariableDeclarator: 2, SwitchCase: 1, assignmentOperator: 'off' },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '7 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = {\n    bar: 1,\n    baz: {\n        qux: 2\n      }\n  },\n  bar = 1;',
      output:
        'var foo = {\n    bar: 1,\n    baz: {\n      qux: 2\n    }\n  },\n  bar = 1;',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 5,
        },
      ],
    },
    {
      code: "var path     = require('path')\n , crypto    = require('crypto')\n;",
      output:
        "var path     = require('path')\n  , crypto    = require('crypto')\n;",
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 2,
        },
      ],
    },
    {
      code: 'var a = 1\n   ,b = 2\n;',
      output: 'var a = 1\n    ,b = 2\n;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 3 },
          line: 2,
        },
      ],
    },
    {
      code: 'class A{\n  constructor(){}\n    a(){}\n    get b(){}\n}',
      output: 'class A{\n    constructor(){}\n    a(){}\n    get b(){}\n}',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: 'var A = class {\n  constructor(){}\n    a(){}\n  get b(){}\n};',
      output:
        'var A = class {\n    constructor(){}\n    a(){}\n    get b(){}\n};',
      options: [4, { VariableDeclarator: 1, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'class A\nextends B {\n}',
      output: 'class A\n    extends B {\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var A = class\nextends B {\n};',
      output: 'var A = class\n    extends B {\n};',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var a = 1,\n    B = class {\n    constructor(){}\n      a(){}\n      get b(){}\n    };',
      output:
        'var a = 1,\n    B = class {\n      constructor(){}\n      a(){}\n      get b(){}\n    };',
      options: [
        2,
        { VariableDeclarator: 2, SwitchCase: 1, assignmentOperator: 'off' },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '{\n    if(a){\n        foo();\n    }\n  else{\n        bar();\n    }\n}',
      output:
        '{\n    if(a){\n        foo();\n    }\n    else{\n        bar();\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: '{\n    if(a){\n        foo();\n    }\n  else\n        bar();\n\n}',
      output:
        '{\n    if(a){\n        foo();\n    }\n    else\n        bar();\n\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: '{\n    if(a)\n        foo();\n  else\n        bar();\n}',
      output: '{\n    if(a)\n        foo();\n    else\n        bar();\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: '(function(){\n  function foo(x) {\n    return x + 1;\n  }\n})();',
      output: '(function(){\nfunction foo(x) {\n  return x + 1;\n}\n})();',
      options: [2, { outerIIFEBody: 0 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: '(function(){\n    function foo(x) {\n        return x + 1;\n    }\n})();',
      output:
        '(function(){\n        function foo(x) {\n            return x + 1;\n        }\n})();',
      options: [4, { outerIIFEBody: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: "if(data) {\nconsole.log('hi');\n}",
      output: "if(data) {\n  console.log('hi');\n}",
      options: [2, { outerIIFEBody: 0 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var ns = function(){\n    function fooVar(x) {\n        return x + 1;\n    }\n}(x);',
      output:
        'var ns = function(){\n        function fooVar(x) {\n            return x + 1;\n        }\n}(x);',
      options: [4, { outerIIFEBody: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'var obj = {\n  foo: function() {\n  return true;\n  }()\n};',
      output: 'var obj = {\n  foo: function() {\n    return true;\n  }()\n};',
      options: [2, { outerIIFEBody: 0 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'typeof function() {\n    function fooVar(x) {\n      return x + 1;\n    }\n}();',
      output:
        'typeof function() {\n  function fooVar(x) {\n    return x + 1;\n  }\n}();',
      options: [2, { outerIIFEBody: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: '{\n\t!function(x) {\n\t\t\t\treturn x + 1;\n\t}()\n};',
      output: '{\n\t!function(x) {\n\t\treturn x + 1;\n\t}()\n};',
      options: ['tab', { outerIIFEBody: 3 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 tabs', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '(function(){\n    function foo(x) {\n    return x + 1;\n    }\n})();',
      output:
        '(function(){\n    function foo(x) {\n        return x + 1;\n    }\n})();',
      options: [4, { outerIIFEBody: 'off' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '(function(){\nfunction foo(x) {\nreturn x + 1;\n}\n})();',
      output: '(function(){\nfunction foo(x) {\n    return x + 1;\n}\n})();',
      options: [4, { outerIIFEBody: 'off' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '(() => {\n    function foo(x) {\n    return x + 1;\n    }\n})();',
      output:
        '(() => {\n    function foo(x) {\n        return x + 1;\n    }\n})();',
      options: [4, { outerIIFEBody: 'off' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '(() => {\nfunction foo(x) {\nreturn x + 1;\n}\n})();',
      output: '(() => {\nfunction foo(x) {\n    return x + 1;\n}\n})();',
      options: [4, { outerIIFEBody: 'off' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'Buffer\n.toString()',
      output: 'Buffer\n    .toString()',
      options: [4, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: "Buffer\n    .indexOf('a')\n.toString()",
      output: "Buffer\n    .indexOf('a')\n    .toString()",
      options: [4, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'Buffer.\nlength',
      output: 'Buffer.\n    length',
      options: [4, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'Buffer.\n\t\tlength',
      output: 'Buffer.\n\tlength',
      options: ['tab', { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: 'Buffer\n  .foo\n  .bar',
      output: 'Buffer\n    .foo\n    .bar',
      options: [2, { MemberExpression: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo() {\n    new\n    .target\n}',
      output: 'function foo() {\n    new\n        .target\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo() {\n    new.\n    target\n}',
      output: 'function foo() {\n    new.\n        target\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) bar();\nelse if (baz) foobar();\n  else if (qux) qux();',
      output: 'if (foo) bar();\nelse if (baz) foobar();\nelse if (qux) qux();',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) bar();\nelse if (baz) foobar();\n  else qux();',
      output: 'if (foo) bar();\nelse if (baz) foobar();\nelse qux();',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo();\n  if (baz) foobar();\n  else qux();',
      output: 'foo();\nif (baz) foobar();\nelse qux();',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) bar();\nelse if (baz) foobar();\n     else if (bip) {\n       qux();\n     }',
      output:
        'if (foo) bar();\nelse if (baz) foobar();\nelse if (bip) {\n  qux();\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 5 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 7 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 5 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (foo) bar();\nelse if (baz) {\n    foobar();\n     } else if (boop) {\n       qux();\n     }',
      output:
        'if (foo) bar();\nelse if (baz) {\n  foobar();\n} else if (boop) {\n  qux();\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 5 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 7 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 5 },
          line: 6,
        },
      ],
    },
    {
      code: 'function foo(aaa,\n    bbb, ccc, ddd) {\n      bar();\n}',
      output: 'function foo(aaa,\n  bbb, ccc, ddd) {\n    bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 1, body: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo(aaa, bbb,\n  ccc, ddd) {\nbar();\n}',
      output: 'function foo(aaa, bbb,\n      ccc, ddd) {\n  bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 3, body: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo(aaa,\n        bbb,\n  ccc) {\n      bar();\n}',
      output: 'function foo(aaa,\n    bbb,\n    ccc) {\n            bar();\n}',
      options: [4, { FunctionDeclaration: { parameters: 1, body: 3 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 6 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo(aaa,\n  bbb, ccc,\n                   ddd, eee, fff) {\n   bar();\n}',
      output:
        'function foo(aaa,\n             bbb, ccc,\n             ddd, eee, fff) {\n  bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 'first', body: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '13 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '13 spaces', actual: 19 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 3 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo(aaa, bbb)\n{\nbar();\n}',
      output: 'function foo(aaa, bbb)\n{\n      bar();\n}',
      options: [2, { FunctionDeclaration: { body: 3 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo(\naaa,\n    bbb) {\nbar();\n}',
      output: 'function foo(\n  aaa,\n  bbb) {\n    bar();\n}',
      options: [2, { FunctionDeclaration: { parameters: 'first', body: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = function(aaa,\n  bbb,\n    ccc,\n      ddd) {\n  bar();\n}',
      output:
        'var foo = function(aaa,\n    bbb,\n    ccc,\n    ddd) {\nbar();\n}',
      options: [2, { FunctionExpression: { parameters: 2, body: 0 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: 'var foo = function(aaa,\n   bbb,\n ccc) {\n  bar();\n}',
      output:
        'var foo = function(aaa,\n  bbb,\n  ccc) {\n                    bar();\n}',
      options: [2, { FunctionExpression: { parameters: 1, body: 10 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 3 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 1 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '20 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = function(aaa,\n  bbb, ccc, ddd,\n                        eee, fff) {\n        bar();\n}',
      output:
        'var foo = function(aaa,\n                   bbb, ccc, ddd,\n                   eee, fff) {\n    bar();\n}',
      options: [4, { FunctionExpression: { parameters: 'first', body: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '19 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '19 spaces', actual: 24 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = function(\naaa, bbb, ccc,\n    ddd, eee) {\n  bar();\n}',
      output:
        'var foo = function(\n  aaa, bbb, ccc,\n  ddd, eee) {\n      bar();\n}',
      options: [2, { FunctionExpression: { parameters: 'first', body: 3 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = bar;\n\t\t\tvar baz = qux;',
      output: 'var foo = bar;\nvar baz = qux;',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: '3 tabs' },
          line: 2,
        },
      ],
    },
    {
      code: 'function foo() {\n\tbar();\n  baz();\n              qux();\n}',
      output: 'function foo() {\n\tbar();\n\tbaz();\n\tqux();\n}',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: '2 spaces' },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: '14 spaces' },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo() {\n  bar();\n\t\t}',
      output: 'function foo() {\n  bar();\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: '2 tabs' },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo() {\n  function bar() {\n        baz();\n  }\n}',
      output: 'function foo() {\n  function bar() {\n    baz();\n  }\n}',
      options: [2, { FunctionDeclaration: { body: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo() {\n  function bar(baz,\n    qux) {\n    foobar();\n  }\n}',
      output:
        'function foo() {\n  function bar(baz,\n      qux) {\n    foobar();\n  }\n}',
      options: [2, { FunctionDeclaration: { body: 1, parameters: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'function foo() {\n  var bar = function(baz,\n          qux) {\n    foobar();\n  };\n}',
      output:
        'function foo() {\n  var bar = function(baz,\n        qux) {\n    foobar();\n  };\n}',
      options: [2, { FunctionExpression: { parameters: 3 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 10 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo.bar(\n      baz, qux, function() {\n        qux;\n      }\n);',
      output:
        'foo.bar(\n      baz, qux, function() {\n            qux;\n      }\n);',
      options: [
        2,
        { FunctionExpression: { body: 3 }, CallExpression: { arguments: 3 } },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 3,
        },
      ],
    },
    {
      code: '{\n    try {\n    }\ncatch (err) {\n    }\nfinally {\n    }\n}',
      output:
        '{\n    try {\n    }\n    catch (err) {\n    }\n    finally {\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: '{\n    do {\n    }\nwhile (true)\n}',
      output: '{\n    do {\n    }\n    while (true)\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo() {\n  return (\n    1\n    )\n}',
      output: 'function foo() {\n  return (\n    1\n  )\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo() {\n  return (\n    1\n    );\n}',
      output: 'function foo() {\n  return (\n    1\n  );\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'function test(){\n  switch(length){\n    case 1: return function(a){\n    return fn.call(that, a);\n    };\n  }\n}',
      output:
        'function test(){\n  switch(length){\n    case 1: return function(a){\n      return fn.call(that, a);\n    };\n  }\n}',
      options: [2, { VariableDeclarator: 2, SwitchCase: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo() {\n   return 1\n}',
      output: 'function foo() {\n  return 1\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 3 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo(\nbar,\n  baz,\n    qux);',
      output: 'foo(\n  bar,\n  baz,\n  qux);',
      options: [2, { CallExpression: { arguments: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo(\n\tbar,\n\tbaz);',
      output: 'foo(\n    bar,\n    baz);',
      options: [2, { CallExpression: { arguments: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: '1 tab' },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: '1 tab' },
          line: 3,
        },
      ],
    },
    {
      code: 'foo(bar,\n\t\tbaz,\n\t\tqux);',
      output: 'foo(bar,\n\tbaz,\n\tqux);',
      options: ['tab', { CallExpression: { arguments: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo(bar, baz,\n         qux);',
      output: 'foo(bar, baz,\n    qux);',
      options: [2, { CallExpression: { arguments: 'first' } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo(\n          bar,\n    baz);',
      output: 'foo(\n  bar,\n  baz);',
      options: [2, { CallExpression: { arguments: 'first' } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 10 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "foo(bar,\n  1 + 2,\n              !baz,\n        new Car('!')\n);",
      output: "foo(bar,\n      1 + 2,\n      !baz,\n      new Car('!')\n);",
      options: [2, { CallExpression: { arguments: 3 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 14 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: 'return (\n    foo\n    );',
      output: 'return (\n    foo\n);',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'return (\n    foo\n    )',
      output: 'return (\n    foo\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) {\n        /* comment */bar();\n}',
      output: 'if (foo) {\n    /* comment */bar();\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
      ],
    },
    {
      code: "foo('bar',\n        /** comment */{\n        ok: true\n    });",
      output: "foo('bar',\n    /** comment */{\n        ok: true\n    });",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo(\n(bar)\n);',
      output: 'foo(\n    (bar)\n);',
      options: [4, { CallExpression: { arguments: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '((\nfoo\n))',
      output: '((\n    foo\n))',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo\n? bar\n    : baz',
      output: 'foo\n  ? bar\n  : baz',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '[\n    foo ?\n        bar :\n        baz,\n        qux\n]',
      output: '[\n    foo ?\n        bar :\n        baz,\n    qux\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 5,
        },
      ],
    },
    {
      code: 'condition\n? () => {\nreturn true\n}\n: condition2\n? () => {\nreturn true\n}\n: () => {\nreturn false\n}',
      output:
        'condition\n  ? () => {\n      return true\n    }\n  : condition2\n    ? () => {\n        return true\n      }\n    : () => {\n        return false\n      }',
      options: [2, { offsetTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 11,
        },
      ],
    },
    {
      code: 'condition\n? () => {\nreturn true\n}\n: condition2\n? () => {\nreturn true\n}\n: () => {\nreturn false\n}',
      output:
        'condition\n  ? () => {\n    return true\n  }\n  : condition2\n    ? () => {\n      return true\n    }\n    : () => {\n      return false\n    }',
      options: [2, { offsetTernaryExpressions: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 11,
        },
      ],
    },
    {
      code: 'foo();\n  // comment\n    /* multiline\n  comment */\nbar();\n // trailing comment',
      output:
        'foo();\n// comment\n/* multiline\n  comment */\nbar();\n// trailing comment',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 1 },
          line: 6,
        },
      ],
    },
    {
      code: '  // comment',
      output: '// comment',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 1,
        },
      ],
    },
    {
      code: 'foo\n  // comment',
      output: 'foo\n// comment',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: '  // comment\nfoo',
      output: '// comment\nfoo',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 1,
        },
      ],
    },
    {
      code: '[\n        // no elements\n]',
      output: '[\n    // no elements\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
      ],
    },
    {
      code: 'var {\nfoo,\n  bar,\n    baz: qux,\n      foobar: baz = foobar\n  } = qux;',
      output:
        'var {\n  foo,\n  bar,\n  baz: qux,\n  foobar: baz = foobar\n} = qux;',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 6 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 6,
        },
      ],
    },
    {
      code: 'const {\n  a\n} = {\n    a: 1\n  }',
      output: 'const {\n  a\n} = {\n  a: 1\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: 'var foo = [\n           bar,\n  baz\n          ]',
      output: 'var foo = [\n    bar,\n    baz\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 11 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 10 },
          line: 4,
        },
      ],
    },
    {
      code: 'var foo = [bar,\nbaz,\n    qux\n]',
      output: 'var foo = [bar,\n    baz,\n    qux\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'var foo = [bar,\n  baz,\n  qux\n]',
      output: 'var foo = [bar,\nbaz,\nqux\n]',
      options: [2, { ArrayExpression: 0 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'var foo = [bar,\n  baz,\n  qux\n]',
      output: 'var foo = [bar,\n                baz,\n                qux\n]',
      options: [2, { ArrayExpression: 8 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'var foo = [bar,\n    baz,\n    qux\n]',
      output: 'var foo = [bar,\n           baz,\n           qux\n]',
      options: [2, { ArrayExpression: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '11 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '11 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'var foo = [bar,\n    baz, qux\n]',
      output: 'var foo = [bar,\n           baz, qux\n]',
      options: [2, { ArrayExpression: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '11 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'var foo = [\n        { bar: 1,\n            baz: 2 },\n        { bar: 3,\n            qux: 4 }\n]',
      output:
        'var foo = [\n        { bar: 1,\n          baz: 2 },\n        { bar: 3,\n          qux: 4 }\n]',
      options: [4, { ArrayExpression: 2, ObjectExpression: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 12 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 12 },
          line: 5,
        },
      ],
    },
    {
      code: 'var foo = {\n  bar: 1,\n  baz: 2\n};',
      output: 'var foo = {\nbar: 1,\nbaz: 2\n};',
      options: [2, { ObjectExpression: 0 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'var quux = { foo: 1, bar: 2,\nbaz: 3 }',
      output: 'var quux = { foo: 1, bar: 2,\n             baz: 3 }',
      options: [2, { ObjectExpression: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '13 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'function foo() {\n    [\n            foo\n    ]\n}',
      output: 'function foo() {\n  [\n          foo\n  ]\n}',
      options: [2, { ArrayExpression: 4 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 12 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'var [\nfoo,\n  bar,\n    baz,\n      foobar = baz\n  ] = qux;',
      output: 'var [\n  foo,\n  bar,\n  baz,\n  foobar = baz\n] = qux;',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 6 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 6,
        },
      ],
    },
    {
      code: "import {\nfoo,\n  bar,\n    baz\n} from 'qux';",
      output: "import {\n    foo,\n    bar,\n    baz\n} from 'qux';",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: "import { foo,\n         bar,\n          baz,\n} from 'qux';",
      output: "import { foo,\n         bar,\n         baz,\n} from 'qux';",
      options: [4, { ImportDeclaration: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '9 spaces', actual: 10 },
          line: 3,
        },
      ],
    },
    {
      code: "import { foo,\n    bar,\n     baz,\n} from 'qux';",
      output: "import { foo,\n    bar,\n    baz,\n} from 'qux';",
      options: [2, { ImportDeclaration: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 5 },
          line: 3,
        },
      ],
    },
    {
      code: 'var foo = 0, bar = 0, baz = 0;\nexport {\nfoo,\n  bar,\n    baz\n};',
      output:
        'var foo = 0, bar = 0, baz = 0;\nexport {\n    foo,\n    bar,\n    baz\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: "var foo = 0, bar = 0, baz = 0;\nexport {\nfoo,\n  bar,\n    baz\n} from 'qux';",
      output:
        "var foo = 0, bar = 0, baz = 0;\nexport {\n    foo,\n    bar,\n    baz\n} from 'qux';",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'var folder = filePath\n  .foo()\n      .bar;',
      output: 'var folder = filePath\n    .foo()\n    .bar;',
      options: [2, { MemberExpression: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: 'for (const foo of bar)\n    baz();',
      output: 'for (const foo of bar)\n  baz();',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'var x = () =>\n    5;',
      output: 'var x = () =>\n  5;',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo && (\n        bar\n)',
      output: 'foo && (\n    bar\n)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo &&\n    !bar(\n)',
      output: 'foo &&\n    !bar(\n    )',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo &&\n    ![].map(() => {\n    bar();\n})',
      output: 'foo &&\n    ![].map(() => {\n        bar();\n    })',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: '[\n] || [\n    ]',
      output: '[\n] || [\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo\n        || (\n                bar\n            )',
      output: 'foo\n        || (\n            bar\n        )',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 16 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 4,
        },
      ],
    },
    {
      code: '1\n+ (\n        1\n    )',
      output: '1\n+ (\n    1\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: '`foo${\nbar}`',
      output: '`foo${\n  bar}`',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '`foo${\n    `bar${\nbaz}`}`',
      output: '`foo${\n  `bar${\n    baz}`}`',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '`foo${\n    `bar${\n  baz\n    }`\n  }`',
      output: '`foo${\n  `bar${\n    baz\n  }`\n}`',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: '`foo${\n(\n  bar\n)\n}`',
      output: '`foo${\n  (\n    bar\n  )\n}`',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo() {\n    `foo${bar}baz${\nqux}foo${\n  bar}baz`\n}',
      output:
        'function foo() {\n    `foo${bar}baz${\n        qux}foo${\n        bar}baz`\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'function foo() {\n    const template = `the indentation of\na curly element in a ${\n        node.type\n    } node is checked.`;\n}',
      output:
        'function foo() {\n    const template = `the indentation of\na curly element in a ${\n    node.type\n} node is checked.`;\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'function outerFunctionForExtraIndent() {\n    function foo() {\n        const template = `the indentation of\n    a curly element in a ${\n            node.type\n        } node is checked.`;\n    }\n}',
      output:
        'function outerFunctionForExtraIndent() {\n    function foo() {\n        const template = `the indentation of\n    a curly element in a ${\n        node.type\n    } node is checked.`;\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 6,
        },
      ],
    },
    {
      code: "function foo() {\n    const template = `this time the\nclosing curly is at the end of the line ${\n            foo}\n        so the spaces before this line aren't removed.`;\n}",
      output:
        "function foo() {\n    const template = `this time the\nclosing curly is at the end of the line ${\n    foo}\n        so the spaces before this line aren't removed.`;\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 12 },
          line: 4,
        },
      ],
    },
    {
      code: '`\n    SELECT ${\n                  foo\n        } FROM THE_DATABASE\n`',
      output: '`\n    SELECT ${\n      foo\n    } FROM THE_DATABASE\n`',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 18 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: '`\n  some code here {\n    {\n      {\n        ${\nnew Array(3).fill(0).map(() => {\n\t\t\t\treturn Math.random()\n\t})\n\t\t\t}\n      }\n    }\n  }\n`',
      output:
        '`\n  some code here {\n    {\n      {\n        ${\n\t\t\t\t\tnew Array(3).fill(0).map(() => {\n\t\t\t\t\t\treturn Math.random()\n\t\t\t\t\t})\n\t\t\t\t}\n      }\n    }\n  }\n`',
      options: ['tab', { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '5 tabs', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 tabs', actual: 4 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '5 tabs', actual: 1 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 tabs', actual: 3 },
          line: 9,
        },
      ],
    },
    {
      code: '`\n  some code here {\n    {\n      {\n        ${\nnew Array(3).fill(0).map(() => {\n        return Math.random()\n  })\n      }\n      }\n    }\n  }\n`',
      output:
        '`\n  some code here {\n    {\n      {\n        ${\n          new Array(3).fill(0).map(() => {\n            return Math.random()\n          })\n        }\n      }\n    }\n  }\n`',
      options: [2, { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 2 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
          line: 9,
        },
      ],
    },
    {
      code: '`\n  some code here {\n    {\n      {${\n  new Array(3).fill(0).map(() => {\n    return Math.random()\n  })\n}\n      }\n    }\n  }\n`',
      output:
        '`\n  some code here {\n    {\n      {${\n        new Array(3).fill(0).map(() => {\n          return Math.random()\n        })\n      }\n      }\n    }\n  }\n`',
      options: [2, { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 2 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 2 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 0 },
          line: 8,
        },
      ],
    },
    {
      code: "`\n  some code here {\n    [\n      {\n        'attr with spaces': ${\n    new Array(3).fill(0).map(() => {\n      return Math.random()\n    })\n                  }\n      }\n    ]\n  }\n`",
      output:
        "`\n  some code here {\n    [\n      {\n        'attr with spaces': ${\n          new Array(3).fill(0).map(() => {\n            return Math.random()\n          })\n        }\n      }\n    ]\n  }\n`",
      options: [2, { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 6 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 4 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 18 },
          line: 9,
        },
      ],
    },
    {
      code: '`\n  start on same line, end with new line {\n    [\n      {\n        attr: ${fn(\n                      1,\n                    2,\n                  3,\n                )\n              }\n      }\n    ]\n  }\n`',
      output:
        '`\n  start on same line, end with new line {\n    [\n      {\n        attr: ${fn(\n            1,\n            2,\n            3,\n          )\n        }\n      }\n    ]\n  }\n`',
      options: [2, { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 22 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 20 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 18 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 16 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 14 },
          line: 10,
        },
      ],
    },
    {
      code: "`\n  <div>\n    <div>\n      <div>\n        <article\n          test=\"test\"\n          /* expression is wrapped with curlies */\n          class=${clsx('container', {\n    hide: 'shouldHide',\nempty: 'isCollapsed',\n                            })}\n        >\n          <div class=\"header\"></div>\n        </article>\n      </div>\n    </div>\n  </div>\n`",
      output:
        "`\n  <div>\n    <div>\n      <div>\n        <article\n          test=\"test\"\n          /* expression is wrapped with curlies */\n          class=${clsx('container', {\n            hide: 'shouldHide',\n            empty: 'isCollapsed',\n          })}\n        >\n          <div class=\"header\"></div>\n        </article>\n      </div>\n    </div>\n  </div>\n`",
      options: [2, { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 28 },
          line: 11,
        },
      ],
    },
    {
      code: "`\n  <div>\n    <div>\n      <div>\n        <article\n          test=\"test\"\n          /* expression is wrapped with new lines */\n          class=${\n          clsx('container', {\n    hide: 'shouldHide',\nempty: 'isCollapsed',\n                  })\n                          }\n        >\n          <div class=\"header\"></div>\n        </article>\n      </div>\n    </div>\n  </div>\n`",
      output:
        "`\n  <div>\n    <div>\n      <div>\n        <article\n          test=\"test\"\n          /* expression is wrapped with new lines */\n          class=${\n            clsx('container', {\n              hide: 'shouldHide',\n              empty: 'isCollapsed',\n            })\n          }\n        >\n          <div class=\"header\"></div>\n        </article>\n      </div>\n    </div>\n  </div>\n`",
      options: [2, { tabLength: 2 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 10 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '14 spaces', actual: 4 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '14 spaces', actual: 0 },
          line: 11,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 18 },
          line: 12,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '10 spaces', actual: 26 },
          line: 13,
        },
      ],
    },
    {
      code: 'if (true) {\n    a = (\n1 +\n        2);\n}',
      output: 'if (true) {\n    a = (\n        1 +\n        2);\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (true) {\n    for (;;) {\n      b();\n  }\n}',
      output: 'if (true) {\n  for (;;) {\n    b();\n  }\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: "function f() {\n    return asyncCall()\n    .then(\n               'some string',\n              [\n              1,\n         2,\n                                   3\n                      ]\n);\n }",
      output:
        "function f() {\n    return asyncCall()\n        .then(\n            'some string',\n            [\n                1,\n                2,\n                3\n            ]\n        );\n}",
      options: [4, { MemberExpression: 1, CallExpression: { arguments: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 15 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 14 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 14 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 9 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 35 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 22 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 1 },
          line: 11,
        },
      ],
    },
    {
      code: 'var x = [\n      [1],\n  [2]\n]',
      output: 'var x = [\n    [1],\n    [2]\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'var y = [\n      {a: 1},\n  {b: 2}\n]',
      output: 'var y = [\n    {a: 1},\n    {b: 2}\n]',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: "echo = spawn('cmd.exe',\n            ['foo', 'bar',\n             'baz']);",
      output:
        "echo = spawn('cmd.exe',\n             ['foo', 'bar',\n              'baz']);",
      options: [
        2,
        { ArrayExpression: 'first', CallExpression: { arguments: 'first' } },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '13 spaces', actual: 12 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '14 spaces', actual: 13 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo(\n  )',
      output: 'foo(\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo(\n        bar,\n    {\n        baz: 1\n    }\n)',
      output: 'foo(\n    bar,\n    {\n        baz: 1\n    }\n)',
      options: [4, { CallExpression: { arguments: 'first' } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
      ],
    },
    {
      code: '  new Foo',
      output: 'new Foo',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 1,
        },
      ],
    },
    {
      code: 'var foo = 0, bar = 0, baz = 0;\nexport {\nfoo,\n        bar,\n  baz\n}',
      output:
        'var foo = 0, bar = 0, baz = 0;\nexport {\n    foo,\n    bar,\n    baz\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo\n    ? bar\n: baz',
      output: 'foo\n    ? bar\n    : baz',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ?\n    bar :\nbaz',
      output: 'foo ?\n    bar :\n    baz',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ?\n    bar\n  : baz',
      output: 'foo ?\n    bar\n    : baz',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo\n    ? bar :\nbaz',
      output: 'foo\n    ? bar :\n    baz',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ? bar\n    : baz ? qux\n        : foobar ? boop\n            : beep',
      output: 'foo ? bar\n    : baz ? qux\n    : foobar ? boop\n    : beep',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 12 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? bar :\n    baz ? qux :\n        foobar ? boop :\n            beep',
      output: 'foo ? bar :\n    baz ? qux :\n    foobar ? boop :\n    beep',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 12 },
          line: 4,
        },
      ],
    },
    {
      code: 'var a =\n    foo ? bar :\n      baz ? qux :\n  foobar ? boop :\n    /*else*/ beep',
      output:
        'var a =\n    foo ? bar :\n    baz ? qux :\n    foobar ? boop :\n    /*else*/ beep',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'var a =\n    foo\n    ? bar\n    : baz',
      output: 'var a =\n    foo\n        ? bar\n        : baz',
      options: [4, { flatTernaryExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? bar\n    : baz ? qux\n    : foobar ? boop\n    : beep',
      output:
        'foo ? bar\n    : baz ? qux\n        : foobar ? boop\n            : beep',
      options: [4, { flatTernaryExpressions: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? bar :\n    baz ? qux :\n    foobar ? boop :\n    beep',
      output:
        'foo ? bar :\n    baz ? qux :\n        foobar ? boop :\n            beep',
      options: [4, { flatTernaryExpressions: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo\n    ? bar\n    : baz\n    ? qux\n    : foobar\n    ? boop\n    : beep',
      output:
        'foo\n    ? bar\n    : baz\n        ? qux\n        : foobar\n            ? boop\n            : beep',
      options: [4, { flatTernaryExpressions: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 7,
        },
      ],
    },
    {
      code: 'foo ?\n    bar :\n    baz ?\n    qux :\n    foobar ?\n    boop :\n    beep',
      output:
        'foo ?\n    bar :\n    baz ?\n        qux :\n        foobar ?\n            boop :\n            beep',
      options: [4, { flatTernaryExpressions: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
          line: 7,
        },
      ],
    },
    {
      code: "foo.bar('baz', function(err) {\n          qux;\n});",
      output: "foo.bar('baz', function(err) {\n  qux;\n});",
      options: [2, { CallExpression: { arguments: 'first' } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 10 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo.bar(function() {\n  cookies;\n}).baz(function() {\n    cookies;\n  });',
      output:
        'foo.bar(function() {\n  cookies;\n}).baz(function() {\n  cookies;\n});',
      options: [2, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo.bar().baz(function() {\n  cookies;\n}).qux(function() {\n    cookies;\n  });',
      output:
        'foo.bar().baz(function() {\n  cookies;\n}).qux(function() {\n  cookies;\n});',
      options: [2, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 5,
        },
      ],
    },
    {
      code: '[ foo,\n  bar ].forEach(function() {\n    baz;\n  })',
      output: '[ foo,\n  bar ].forEach(function() {\n  baz;\n})',
      options: [2, { ArrayExpression: 'first', MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo[\n    bar\n    ];',
      output: 'foo[\n    bar\n];',
      options: [4, { MemberExpression: 1 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo({\nbar: 1,\nbaz: 2\n})',
      output: 'foo({\n    bar: 1,\n    baz: 2\n})',
      options: [4, { ObjectExpression: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo(\n                        bar, baz,\n                        qux);',
      output: 'foo(\n  bar, baz,\n  qux);',
      options: [2, { CallExpression: { arguments: 'first' } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 24 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 24 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) bar()\n\n    ; [1, 2, 3].map(baz)',
      output: 'if (foo) bar()\n\n; [1, 2, 3].map(baz)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo)\n;',
      output: 'if (foo)\n    ;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: "import {foo}\nfrom 'bar';",
      output: "import {foo}\n    from 'bar';",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: "export {foo}\nfrom 'bar';",
      output: "export {foo}\n    from 'bar';",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '(\n    a\n) => b => {\n        c\n    }',
      output: '(\n    a\n) => b => {\n    c\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: '(\n    a\n) => b => c => d => {\n        e\n    }',
      output: '(\n    a\n) => b => c => d => {\n    e\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (\n    foo\n) bar(\n        baz\n    );',
      output: 'if (\n    foo\n) bar(\n    baz\n);',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: '(\n    foo\n)(\n        bar\n    )',
      output: '(\n    foo\n)(\n    bar\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: '(() =>\n    foo\n)(\n        bar\n    )',
      output: '(() =>\n    foo\n)(\n    bar\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: '(() => {\n    foo();\n})(\n        bar\n    )',
      output: '(() => {\n    foo();\n})(\n    bar\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo.\n  bar.\n      baz',
      output: 'foo.\n    bar.\n    baz',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: "const foo = a.b(),\n    longName\n        = (baz(\n                'bar',\n                'bar'\n            ));",
      output:
        "const foo = a.b(),\n    longName\n        = (baz(\n            'bar',\n            'bar'\n        ));",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 16 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 16 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 6,
        },
      ],
    },
    {
      code: "const foo = a.b(),\n    longName =\n        (baz(\n                'bar',\n                'bar'\n            ));",
      output:
        "const foo = a.b(),\n    longName =\n        (baz(\n            'bar',\n            'bar'\n        ));",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 16 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 16 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 6,
        },
      ],
    },
    {
      code: "const foo = a.b(),\n    longName\n        =baz(\n            'bar',\n            'bar'\n    );",
      output:
        "const foo = a.b(),\n    longName\n        =baz(\n            'bar',\n            'bar'\n        );",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 6,
        },
      ],
    },
    {
      code: "const foo = a.b(),\n    longName\n        =(\n        'fff'\n        );",
      output:
        "const foo = a.b(),\n    longName\n        =(\n            'fff'\n        );",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: '<App\n  foo\n/>',
      output: '<App\n    foo\n/>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: '<App\n  foo\n  />',
      output: '<App\n  foo\n/>',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: '<App\n  foo\n  ></App>',
      output: '<App\n  foo\n></App>',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 3,
        },
      ],
    },
    {
      code: 'const Button = function(props) {\n  return (\n    <Button\n      size={size}\n      onClick={onClick}\n                                    >\n                                      Button Text\n    </Button>\n  );\n};',
      output:
        'const Button = function(props) {\n  return (\n    <Button\n      size={size}\n      onClick={onClick}\n    >\n      Button Text\n    </Button>\n  );\n};',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 36 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 38 },
          line: 6,
        },
      ],
    },
    {
      code: 'var x = function() {\n  return <App\n    foo\n         />\n}',
      output: 'var x = function() {\n  return <App\n    foo\n  />\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 9 },
          line: 4,
        },
      ],
    },
    {
      code: 'var x = <App\n  foo\n        />',
      output: 'var x = <App\n  foo\n/>',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 8 },
          line: 3,
        },
      ],
    },
    {
      code: 'var x = (\n  <Something\n    />\n)',
      output: 'var x = (\n  <Something\n  />\n)',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '<App\n\tfoo\n\t/>',
      output: '<App\n\tfoo\n/>',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 tabs', actual: 1 },
          line: 3,
        },
      ],
    },
    {
      code: '<App\n\tfoo\n\t></App>',
      output: '<App\n\tfoo\n></App>',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 tabs', actual: 1 },
          line: 3,
        },
      ],
    },
    {
      code: '<\n    input\n    type=\n    "number"\n/>',
      output: '<\n    input\n    type=\n        "number"\n/>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: "<\n    input\n    type=\n    {'number'}\n/>",
      output: "<\n    input\n    type=\n        {'number'}\n/>",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: '<\n    input\n    type\n    ="number"\n/>',
      output: '<\n    input\n    type\n        ="number"\n/>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? (\n    bar\n) : (\n        baz\n    )',
      output: 'foo ? (\n    bar\n) : (\n    baz\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo ? (\n    <div>\n    </div>\n) : (\n        <span>\n        </span>\n    )',
      output:
        'foo ? (\n    <div>\n    </div>\n) : (\n    <span>\n    </span>\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 7,
        },
      ],
    },
    {
      code: '<div>\n    {\n    (\n        1\n    )\n    }\n</div>',
      output:
        '<div>\n    {\n        (\n            1\n        )\n    }\n</div>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: '<div>\n    {\n      /* foo */\n    }\n</div>',
      output: '<div>\n    {\n        /* foo */\n    }\n</div>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: '<div\n{...props}\n/>',
      output: '<div\n    {...props}\n/>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '<div\n    {\n      ...props\n    }\n/>',
      output: '<div\n    {\n        ...props\n    }\n/>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: '<div>foo\n<div>bar</div>\n</div>',
      output: '<div>foo\n    <div>bar</div>\n</div>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '<small>Foo bar&nbsp;\n<a>baz qux</a>.\n</small>',
      output: '<small>Foo bar&nbsp;\n    <a>baz qux</a>.\n</small>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '<>\n<A />\n</>',
      output: '<>\n    <A />\n</>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '<\n    >\n    <A />\n</>',
      output: '<\n>\n    <A />\n</>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: '<>\n    <A />\n</\n    >',
      output: '<>\n    <A />\n</\n>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: '<\n    >\n    <A />\n</\n    >',
      output: '<\n>\n    <A />\n</\n>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: '< // Comment\n    >\n    <A />\n</>',
      output: '< // Comment\n>\n    <A />\n</>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: '<>\n    <A />\n</ // Comment\n    >',
      output: '<>\n    <A />\n</ // Comment\n>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: '< /* Comment */\n    >\n    <A />\n</>',
      output: '< /* Comment */\n>\n    <A />\n</>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: '<>\n    <A />\n</ /* Comment */\n    >',
      output: '<>\n    <A />\n</ /* Comment */\n>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'class Foo {\nfoo() {\nbar();\n}\n}',
      output: 'class Foo {\nfoo() {\n    bar();\n}\n}',
      options: [4, { ignoredNodes: ['ClassBody'] }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '$(function() {\n\nfoo();\nbar();\n\nfoo(function() {\nbaz();\n});\n\n});',
      output:
        '$(function() {\n\nfoo();\nbar();\n\nfoo(function() {\n    baz();\n});\n\n});',
      options: [
        4,
        {
          ignoredNodes: [
            "ExpressionStatement > CallExpression[callee.name='$'] > FunctionExpression > BlockStatement",
          ],
        },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: '(function($) {\n$(function() {\nfoo;\n});\n})()',
      output: '(function($) {\n$(function() {\n    foo;\n});\n})()',
      options: [
        4,
        {
          ignoredNodes: [
            'ExpressionStatement > CallExpression > FunctionExpression.callee > BlockStatement',
          ],
        },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) {\n    doSomething();\n\n// Intentionally unindented comment\n    doSomethingElse();\n}',
      output:
        'if (foo) {\n    doSomething();\n\n    // Intentionally unindented comment\n    doSomethingElse();\n}',
      options: [4, { ignoreComments: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (foo) {\n    doSomething();\n\n/* Intentionally unindented comment */\n    doSomethingElse();\n}',
      output:
        'if (foo) {\n    doSomething();\n\n    /* Intentionally unindented comment */\n    doSomethingElse();\n}',
      options: [4, { ignoreComments: false }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'const obj = {\n    foo () {\n        return condition ? // comment\n        1 :\n            2\n    }\n}',
      output:
        'const obj = {\n    foo () {\n        return condition ? // comment\n            1 :\n            2\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (foo) {\n\n// Comment cannot align with code immediately above if there is a whitespace gap\n    doSomething();\n}',
      output:
        'if (foo) {\n\n    // Comment cannot align with code immediately above if there is a whitespace gap\n    doSomething();\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) {\n    foo(\n        bar);\n// Comment cannot align with code immediately below if there is a whitespace gap\n\n}',
      output:
        'if (foo) {\n    foo(\n        bar);\n    // Comment cannot align with code immediately below if there is a whitespace gap\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: '[{\n    foo\n},\n\n    // Comment between nodes\n\n{\n    bar\n}];',
      output: '[{\n    foo\n},\n\n// Comment between nodes\n\n{\n    bar\n}];',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'let foo\n\n    // comment\n\n;(async () => {})()',
      output: 'let foo\n\n// comment\n\n;(async () => {})()',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'let foo\n    // comment\n;(async () => {})()',
      output: 'let foo\n// comment\n;(async () => {})()',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'let foo\n\n/* comment */;\n\n(async () => {})()',
      output: 'let foo\n\n    /* comment */;\n\n(async () => {})()',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '    // comment\n\n;(async () => {})()',
      output: '// comment\n\n;(async () => {})()',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 1,
        },
      ],
    },
    {
      code: '    // comment\n;(async () => {})()',
      output: '// comment\n;(async () => {})()',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 1,
        },
      ],
    },
    {
      code: '{\n    let foo\n\n        // comment\n\n    ;(async () => {})()\n\n}',
      output:
        '{\n    let foo\n\n    // comment\n\n    ;(async () => {})()\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: '{\n    let foo\n        // comment\n    ;(async () => {})()\n\n}',
      output: '{\n    let foo\n    // comment\n    ;(async () => {})()\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
      ],
    },
    {
      code: '{\n    let foo\n\n    /* comment */;\n\n    (async () => {})()\n\n}',
      output:
        '{\n    let foo\n\n        /* comment */;\n\n    (async () => {})()\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'const foo = 1\nconst bar = foo\n\n    /* comment */\n\n;[1, 2, 3].forEach(() => {})',
      output:
        'const foo = 1\nconst bar = foo\n\n/* comment */\n\n;[1, 2, 3].forEach(() => {})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'const foo = 1\nconst bar = foo\n    /* comment */\n;[1, 2, 3].forEach(() => {})',
      output:
        'const foo = 1\nconst bar = foo\n/* comment */\n;[1, 2, 3].forEach(() => {})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'const foo = 1\nconst bar = foo\n\n/* comment */;\n\n[1, 2, 3].forEach(() => {})',
      output:
        'const foo = 1\nconst bar = foo\n\n    /* comment */;\n\n[1, 2, 3].forEach(() => {})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: '    /* comment */\n\n;[1, 2, 3].forEach(() => {})',
      output: '/* comment */\n\n;[1, 2, 3].forEach(() => {})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 1,
        },
      ],
    },
    {
      code: '    /* comment */\n;[1, 2, 3].forEach(() => {})',
      output: '/* comment */\n;[1, 2, 3].forEach(() => {})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 1,
        },
      ],
    },
    {
      code: '{\n    const foo = 1\n    const bar = foo\n\n        /* comment */\n\n    ;[1, 2, 3].forEach(() => {})\n\n}',
      output:
        '{\n    const foo = 1\n    const bar = foo\n\n    /* comment */\n\n    ;[1, 2, 3].forEach(() => {})\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 5,
        },
      ],
    },
    {
      code: '{\n    const foo = 1\n    const bar = foo\n        /* comment */\n    ;[1, 2, 3].forEach(() => {})\n\n}',
      output:
        '{\n    const foo = 1\n    const bar = foo\n    /* comment */\n    ;[1, 2, 3].forEach(() => {})\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
      ],
    },
    {
      code: '{\n    const foo = 1\n    const bar = foo\n\n    /* comment */;\n\n    [1, 2, 3].forEach(() => {})\n\n}',
      output:
        '{\n    const foo = 1\n    const bar = foo\n\n        /* comment */;\n\n    [1, 2, 3].forEach(() => {})\n\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'import(\nsource\n    )',
      output: 'import(\n    source\n)',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo(() => {\n    tag`\n    multiline\n    template${a} ${b}\n    literal\n    `(() => {\n    bar();\n    });\n});',
      output:
        'foo(() => {\n    tag`\n    multiline\n    template${a} ${b}\n    literal\n    `(() => {\n        bar();\n    });\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 7,
        },
      ],
    },
    {
      code: '{\n        tag`\n    multiline\n    template\n    literal\n    ${a} ${b}`(() => {\n            bar();\n        });\n}',
      output:
        '{\n    tag`\n    multiline\n    template\n    literal\n    ${a} ${b}`(() => {\n        bar();\n    });\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 8,
        },
      ],
    },
    {
      code: 'foo(() => {\n    tagOne`${a} ${b}\n    multiline\n    template\n    literal\n    `(() => {\n            tagTwo`\n        multiline\n        template\n        literal\n        `(() => {\n            bar();\n        });\n\n            baz();\n});\n});',
      output:
        'foo(() => {\n    tagOne`${a} ${b}\n    multiline\n    template\n    literal\n    `(() => {\n        tagTwo`\n        multiline\n        template\n        literal\n        `(() => {\n            bar();\n        });\n\n        baz();\n    });\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 15,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 16,
        },
      ],
    },
    {
      code: '{\n    tagOne`\n    multiline\n    template\n    literal\n    ${a} ${b}`(() => {\n            tagTwo`\n        multiline\n        template\n        literal\n        `(() => {\n            bar();\n        });\n\n            baz();\n});\n}',
      output:
        '{\n    tagOne`\n    multiline\n    template\n    literal\n    ${a} ${b}`(() => {\n        tagTwo`\n        multiline\n        template\n        literal\n        `(() => {\n            bar();\n        });\n\n        baz();\n    });\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 15,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 16,
        },
      ],
    },
    {
      code: 'tagOne`multiline ${a} ${b}\n        template\n        literal\n        `(() => {\nfoo();\n\n    tagTwo`multiline\n            template\n            literal\n        `({\n    bar: 1,\n        baz: 2\n    });\n});',
      output:
        'tagOne`multiline ${a} ${b}\n        template\n        literal\n        `(() => {\n    foo();\n\n    tagTwo`multiline\n            template\n            literal\n        `({\n        bar: 1,\n        baz: 2\n    });\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 11,
        },
      ],
    },
    {
      code: 'tagOne`multiline\n    template ${a} ${b}\n    literal`({\n        foo: 1,\nbar: tagTwo`multiline\n        template\n        literal`(() => {\n\nbaz();\n    })\n});',
      output:
        'tagOne`multiline\n    template ${a} ${b}\n    literal`({\n    foo: 1,\n    bar: tagTwo`multiline\n        template\n        literal`(() => {\n\n        baz();\n    })\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 9,
        },
      ],
    },
    {
      code: 'foo.bar` template literal `(() => {\n        baz();\n})',
      output: 'foo.bar` template literal `(() => {\n    baz();\n})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo.bar.baz` template literal `(() => {\nbaz();\n    })',
      output: 'foo.bar.baz` template literal `(() => {\n    baz();\n})',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo\n    .bar` template\n        literal `(() => {\n        baz();\n})',
      output:
        'foo\n    .bar` template\n        literal `(() => {\n        baz();\n    })',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo\n    .test`\n    ${a} ${b}\n    `(() => {\nbar();\n    })',
      output:
        'foo\n    .test`\n    ${a} ${b}\n    `(() => {\n        bar();\n    })',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo\n    .test`\n    ${a} ${b}\n    `(() => {\nbar();\n    })',
      output: 'foo\n.test`\n    ${a} ${b}\n    `(() => {\n    bar();\n})',
      options: [4, { MemberExpression: 0 }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 6,
        },
      ],
    },
    {
      code: 'obj\n?.prop\n?.[key]\n?.\n[key]',
      output: 'obj\n    ?.prop\n    ?.[key]\n    ?.\n        [key]',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: '(\n    longSomething\n        ?.prop\n        ?.[key]\n)\n?.prop\n?.[key]',
      output:
        '(\n    longSomething\n        ?.prop\n        ?.[key]\n)\n    ?.prop\n    ?.[key]',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'obj\n?.(arg)\n?.\n(arg)',
      output: 'obj\n    ?.(arg)\n    ?.\n    (arg)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: '(\n    longSomething\n        ?.(arg)\n        ?.(arg)\n)\n?.(arg)\n?.(arg)',
      output:
        '(\n    longSomething\n        ?.(arg)\n        ?.(arg)\n)\n    ?.(arg)\n    ?.(arg)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'const foo = async (arg1,\n                    arg2) =>\n{\n  return arg1 + arg2;\n}',
      output:
        'const foo = async (arg1,\n                   arg2) =>\n{\n  return arg1 + arg2;\n}',
      options: [
        2,
        {
          FunctionDeclaration: { parameters: 'first' },
          FunctionExpression: { parameters: 'first' },
        },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '19 spaces', actual: 20 },
          line: 2,
        },
      ],
    },
    {
      code: 'const a = async\n b => {}',
      output: 'const a = async\nb => {}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 1 },
          line: 2,
        },
      ],
    },
    {
      code: 'class C {\nfield1;\nstatic field2;\n}',
      output: 'class C {\n    field1;\n    static field2;\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'class C {\nfield1\n=\n0\n;\nstatic\nfield2\n=\n0\n;\n}',
      output:
        'class C {\n    field1\n        =\n            0\n            ;\n    static\n        field2\n            =\n                0\n                ;\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 0 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '16 spaces', actual: 0 },
          line: 10,
        },
      ],
    },
    {
      code: 'class C {\n[\nfield1\n]\n=\n0\n;\nstatic\n[\nfield2\n]\n=\n0\n;\n[\nfield3\n] =\n0;\n[field4] =\n0;\n}',
      output:
        'class C {\n    [\n        field1\n    ]\n        =\n            0\n            ;\n    static\n    [\n        field2\n    ]\n        =\n            0\n            ;\n    [\n        field3\n    ] =\n        0;\n    [field4] =\n        0;\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 10,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 11,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 12,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 13,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 14,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 15,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 16,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 17,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 18,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 19,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 20,
        },
      ],
    },
    {
      code: 'class C {\nfield1 = (\nfoo\n+ bar\n);\n}',
      output: 'class C {\n    field1 = (\n        foo\n+ bar\n    );\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\n#aaa\nfoo() {\nreturn this.#aaa\n}\n}',
      output:
        'class C {\n    #aaa\n    foo() {\n        return this.#aaa\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nfoo();\nbar();\n}\n}',
      output: 'class C {\n  static {\n    foo();\n    bar();\n  }\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nfoo();\nbar();\n}\n}',
      output:
        'class C {\n    static {\n        foo();\n        bar();\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\n        static {\n    foo();\nbar();\n        }\n}',
      output:
        'class C {\n    static {\n        foo();\n        bar();\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nfoo();\nbar();\n}\n}',
      output:
        'class C {\n    static {\n            foo();\n            bar();\n    }\n}',
      options: [4, { StaticBlock: { body: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nfoo();\nbar();\n}\n}',
      output: 'class C {\n    static {\n    foo();\n    bar();\n    }\n}',
      options: [4, { StaticBlock: { body: 0 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nfoo();\nbar();\n}\n}',
      output: 'class C {\n\tstatic {\n\t\tfoo();\n\t\tbar();\n\t}\n}',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 tabs', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 tabs', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nfoo();\nbar();\n}\n}',
      output: 'class C {\n\tstatic {\n\t\t\tfoo();\n\t\t\tbar();\n\t}\n}',
      options: ['tab', { StaticBlock: { body: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '3 tabs', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '3 tabs', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic\n{\nfoo();\nbar();\n}\n}',
      output:
        'class C {\n    static\n    {\n        foo();\n        bar();\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'class C {\n    static\n        {\n        foo();\n        bar();\n        }\n}',
      output:
        'class C {\n    static\n    {\n        foo();\n        bar();\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 6,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nvar x,\ny;\n}\n}',
      output:
        'class C {\n    static {\n        var x,\n            y;\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'class C {\nstatic\n{\nvar x,\ny;\n}\n}',
      output:
        'class C {\n    static\n    {\n        var x,\n            y;\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\nif (foo) {\nbar;\n}\n}\n}',
      output:
        'class C {\n    static {\n        if (foo) {\n            bar;\n        }\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'class C {\nstatic {\n{\nbar;\n}\n}\n}',
      output:
        'class C {\n    static {\n        {\n            bar;\n        }\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'class C {\nstatic {}\n\nstatic {\n}\n\nstatic\n{\n}\n}',
      output:
        'class C {\n    static {}\n\n    static {\n    }\n\n    static\n    {\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
      ],
    },
    {
      code: 'class C {\n\nstatic {\n    foo;\n}\n\nstatic {\n    bar;\n}\n\n}',
      output:
        'class C {\n\n    static {\n        foo;\n    }\n\n    static {\n        bar;\n    }\n\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
      ],
    },
    {
      code: 'class C {\n\nx = 1;\n\nstatic {\n    foo;\n}\n\ny = 2;\n\n}',
      output:
        'class C {\n\n    x = 1;\n\n    static {\n        foo;\n    }\n\n    y = 2;\n\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
      ],
    },
    {
      code: 'class C {\n\nmethod1(param) {\n    foo;\n}\n\nstatic {\n    bar;\n}\n\nmethod2(param) {\n    foo;\n}\n\n}',
      output:
        'class C {\n\n    method1(param) {\n        foo;\n    }\n\n    static {\n        bar;\n    }\n\n    method2(param) {\n        foo;\n    }\n\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 11,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 12,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 13,
        },
      ],
    },
    {
      code: 'function f() {\nclass C {\nstatic {\nfoo();\nbar();\n}\n}\n}',
      output:
        'function f() {\n    class C {\n        static {\n            foo();\n            bar();\n        }\n    }\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'class C {\nmethod() {\nfoo;\n}\nstatic {\nbar;\n}\n}',
      output:
        'class C {\n    method() {\n            foo;\n    }\n    static {\n            bar;\n    }\n}',
      options: [
        4,
        { FunctionExpression: { body: 2 }, StaticBlock: { body: 2 } },
      ],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'class C {\nfoo =\n"bar";\n}',
      output: 'class C {\n    foo =\n        "bar";\n}',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: "if (2 > 1)\n\tconsole.log('a')\n\t;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "if (2 > 1)\n\tconsole.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 tabs', actual: 1 },
          line: 3,
        },
      ],
    },
    {
      code: "if (2 > 1)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "if (2 > 1)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo) bar();\n    baz()',
      output: 'if (foo) bar();\nbaz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'if (foo) bar()\n    ;baz()',
      output: 'if (foo) bar()\n;baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar();\n    baz();',
      output: 'if (foo)\n    bar();\nbaz();',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar()\n    ; baz()',
      output: 'if (foo)\n    bar()\n; baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar()\n    ;baz()\n    qux()',
      output: 'if (foo)\n    bar()\n;baz()\nqux()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar()\n    ;else\n    baz()',
      output: 'if (foo)\n    bar()\n;else\n    baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar()\nelse\n    baz()\n    ;qux()',
      output: 'if (foo)\n    bar()\nelse\n    baz()\n;qux()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar)\n        baz()\n    ;qux()',
      output: 'if (foo)\n    if (bar)\n        baz()\n;qux()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar()\nelse if (baz)\n    qux()\n    ;quux()',
      output: 'if (foo)\n    bar()\nelse if (baz)\n    qux()\n;quux()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar)\n        baz()\n    else\n        qux()\n    ;quux()',
      output:
        'if (foo)\n    if (bar)\n        baz()\n    else\n        qux()\n;quux()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (foo)\n    bar()\n;\nbaz()',
      output: 'if (foo)\n    bar()\n    ;\nbaz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'if (foo)\n;\nbaz()',
      output: 'if (foo)\n    ;\nbaz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'if (foo)\n    ;baz()',
      output: 'if (foo)\n;baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'if (foo);\n    else\n    baz()',
      output: 'if (foo);\nelse\n    baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'if (foo)\n;\nelse\n    baz()',
      output: 'if (foo)\n    ;\nelse\n    baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'if (foo)\n    ;else\n    baz()',
      output: 'if (foo)\n;else\n    baz()',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'do foo();\n    while (bar)',
      output: 'do foo();\nwhile (bar)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'do foo()\n    ;while (bar)',
      output: 'do foo()\n;while (bar)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'do\n    foo();\n    while (bar)',
      output: 'do\n    foo();\nwhile (bar)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'do\n    foo()\n    ;while (bar)',
      output: 'do\n    foo()\n;while (bar)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'do;\n    while (foo)',
      output: 'do;\nwhile (foo)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: 'do\n;\nwhile (foo)',
      output: 'do\n    ;\nwhile (foo)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'do\n    ;while (foo)',
      output: 'do\n;while (foo)',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: "while (2 > 1)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "while (2 > 1)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "for (;;)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "for (;;)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "for (a in b)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "for (a in b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "for (a of b)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "for (a of b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'with (a)\n    console.log(b)\n    ;[1, 2, 3].forEach(x=>console.log(x))',
      output:
        'with (a)\n    console.log(b)\n;[1, 2, 3].forEach(x=>console.log(x))',
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "label: for (a of b)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "label: for (a of b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: "label:\nfor (a of b)\n    console.log('a')\n    ;[1, 2, 3].forEach(x=>console.log(x))",
      output:
        "label:\nfor (a of b)\n    console.log('a')\n;[1, 2, 3].forEach(x=>console.log(x))",
      options: [4],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'if (foo)\n\tif (bar) doSomething();\n\telse doSomething();\nelse\nif (bar) doSomething();\nelse doSomething();',
      output:
        'if (foo)\n\tif (bar) doSomething();\n\telse doSomething();\nelse\n\tif (bar) doSomething();\n\telse doSomething();',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (foo)\n\tif (bar) doSomething();\n\telse doSomething();\nelse\n\t\tif (bar) doSomething();\n\t\telse doSomething();',
      output:
        'if (foo)\n\tif (bar) doSomething();\n\telse doSomething();\nelse\n\tif (bar) doSomething();\n\telse doSomething();',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 2 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 2 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (bar) doSomething();\nelse doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar) doSomething();\n    else doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (bar)\ndoSomething();\nelse doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar)\n        doSomething();\n    else doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (bar) doSomething();\nelse\ndoSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar) doSomething();\n    else\n        doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (bar)\n    doSomething();\nelse\ndoSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar)\n        doSomething();\n    else\n        doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 8,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar) doSomething();\n    else doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar) doSomething();\nelse doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\n    else if (bar)\n        doSomething();\n    else doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar)\n    doSomething();\nelse doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar) doSomething();\n     else\n         doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar) doSomething();\nelse\n    doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 5 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 9 },
          line: 6,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar)\ndoSomething();\nelse\ndoSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (bar)\n    doSomething();\nelse\n    doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (bar) doSomething();\n    else doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo)\n        if (bar) doSomething();\n        else doSomething();\n    else\n        if (bar) doSomething();\n        else doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 10,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (foo)\nif (bar) doSomething();\nelse\nif (bar) doSomething();\nelse doSomething();\nelse doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo)\n        if (bar) doSomething();\n        else\n            if (bar) doSomething();\n            else doSomething();\n    else doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 8,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 9,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 10,
        },
      ],
    },
    {
      code: 'if (foo)\nif (bar) doSomething();\nelse doSomething();\nelse if (foo) doSomething();\n    else doSomething();',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo) doSomething();\nelse doSomething();',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo) {\ndoSomething();\n}',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo) {\n    doSomething();\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo)\n    {\n        doSomething();\n    }',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse if (foo)\n{\n    doSomething();\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 7,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (foo) {\n    doSomething();\n}',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo) {\n        doSomething();\n    }',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
        },
      ],
    },
    {
      code: 'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\nif (foo)\n{\n    doSomething();\n}',
      output:
        'if (foo)\n    if (bar) doSomething();\n    else doSomething();\nelse\n    if (foo)\n    {\n        doSomething();\n    }',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 7,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
        },
      ],
    },
    {
      code: 'function foo() {\n  bar();\n  \tbaz();\n\t   \t\t\t  \t\t\t  \t   \tqux();\n}',
      output: 'function foo() {\n  bar();\n  baz();\n  qux();\n}',
      options: [2],
    },
    {
      code: 'function foo() {\n  bar();\n   \t\t}',
      output: 'function foo() {\n  bar();\n}',
      options: [2],
    },
    {
      code: '{\n\tvar x = 1,\n\t    y = 2;\n}',
      output: '{\n\tvar x = 1,\n\t\ty = 2;\n}',
      options: ['tab'],
    },
    {
      code: 'import json from "./foo.json"\nwith\n{\ntype\n:\n"json"\n};',
      output:
        'import json from "./foo.json"\n    with\n    {\n        type\n        :\n        "json"\n    };',
    },
    {
      code: 'import "./foo.json"\nwith\n{\ntype\n:\n"json"\n};',
      output:
        'import "./foo.json"\n    with\n    {\n        type\n        :\n        "json"\n    };',
    },
    {
      code: 'import(\n"./foo.json"\n,\n{\ntype\n:\n"json"\n}\n,\n)',
      output:
        'import(\n    "./foo.json"\n    ,\n    {\n        type\n        :\n"json"\n    }\n    ,\n)',
    },
    {
      code: 'isHeader(1)\n  ? renderSectionHeader?.(\n    typeof item === "string" ? item : "",\n    virtualRow.size,\n  )\n  : renderItem(\n    item,\n    virtualRow.size,\n  )',
      output:
        'isHeader(1)\n  ? renderSectionHeader?.(\n      typeof item === "string" ? item : "",\n      virtualRow.size,\n    )\n  : renderItem(\n      item,\n      virtualRow.size,\n    )',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: 'menus\n  ? await Promise.all(\n    menus.map(async (menu) => ({\n      menuName: menu.name,\n      menu: await resolveUrlToFile(menu.fileUrl),\n    })),\n  )\n  : []',
      output:
        'menus\n  ? await Promise.all(\n      menus.map(async (menu) => ({\n        menuName: menu.name,\n        menu: await resolveUrlToFile(menu.fileUrl),\n      })),\n    )\n  : []',
      options: [2, { offsetTernaryExpressions: true }],
    },
    {
      code: 'menus\n  ? new abc({\n    a: 1,\n    b: 2\n  })\n  : undefined',
      output:
        'menus\n  ? new abc({\n      a: 1,\n      b: 2\n    })\n  : undefined',
      options: [2, { offsetTernaryExpressions: true }],
    },

    // ---- from indent._ts_.test.ts ----
    {
      code: "\n// ClassDeclaration\nabstract class Foo {\nconstructor() {}\nmethod() {\nconsole.log('hi');\n}\n}",
      output:
        "\n// ClassDeclaration\nabstract class Foo {\n    constructor() {}\n    method() {\n        console.log('hi');\n    }\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSAbstractPropertyDefinition\nclass Foo {\nabstract bar : baz;\nabstract foo : {\na : number\nb : number\n};\n}',
      output:
        '\n// TSAbstractPropertyDefinition\nclass Foo {\n    abstract bar : baz;\n    abstract foo : {\n        a : number\n        b : number\n    };\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSAbstractMethodDefinition\nclass Foo {\nabstract bar() : baz;\nabstract foo() : {\na : number\nb : number\n};\n}',
      output:
        '\n// TSAbstractMethodDefinition\nclass Foo {\n    abstract bar() : baz;\n    abstract foo() : {\n        a : number\n        b : number\n    };\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSAsExpression\nconst foo = {} as {\nfoo: string,\nbar: number,\n};',
      output:
        '\n// TSAsExpression\nconst foo = {} as {\n    foo: string,\n    bar: number,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSAsExpression\nconst foo = {} as\n{\nfoo: string,\nbar: number,\n};',
      output:
        '\n// TSAsExpression\nconst foo = {} as\n{\n    foo: string,\n    bar: number,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSConditionalType\ntype Foo<T> = T extends string\n? {\na: number,\nb: boolean\n}\n: {\nc: string\n};',
      output:
        '\n// TSConditionalType\ntype Foo<T> = T extends string\n    ? {\n        a: number,\n        b: boolean\n    }\n    : {\n        c: string\n    };',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 9,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 10,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSConditionalType\ntype Foo<T> = T extends string ? {\na: number,\nb: boolean\n} : string;',
      output:
        '\n// TSConditionalType\ntype Foo<T> = T extends string ? {\n    a: number,\n    b: boolean\n} : string;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSConstructorType\ntype Constructor<T> = new (\n...args: any[]\n) => T;',
      output:
        '\n// TSConstructorType\ntype Constructor<T> = new (\n    ...args: any[]\n) => T;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSConstructSignature\ninterface Foo {\nnew () : Foo\nnew () : {\nbar : string\nbaz : string\n}\n}',
      output:
        '\n// TSConstructSignature\ninterface Foo {\n    new () : Foo\n    new () : {\n        bar : string\n        baz : string\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSDeclareFunction\ndeclare function foo() : {\nbar : number,\nbaz : string,\n};',
      output:
        '\n// TSDeclareFunction\ndeclare function foo() : {\n    bar : number,\n    baz : string,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSEmptyBodyFunctionExpression\nclass Foo {\nconstructor(\na : string,\nb : {\nc : number\n}\n)\n}',
      output:
        '\n// TSEmptyBodyFunctionExpression\nclass Foo {\n    constructor(\n        a : string,\n        b : {\n            c : number\n        }\n    )\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSEnumDeclaration, TSEnumMember\nenum Foo {\nbar = 1,\nbaz = 1,\n}',
      output:
        '\n// TSEnumDeclaration, TSEnumMember\nenum Foo {\n    bar = 1,\n    baz = 1,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSEnumDeclaration, TSEnumMember\nenum Foo\n{\nbar = 1,\nbaz = 1,\n}',
      output:
        '\n// TSEnumDeclaration, TSEnumMember\nenum Foo\n{\n    bar = 1,\n    baz = 1,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSExportAssignment\nexport = {\na: 1,\nb: 2,\n}',
      output: '\n// TSExportAssignment\nexport = {\n    a: 1,\n    b: 2,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSFunctionType\nconst foo: () => void = () => ({\na: 1,\nb: 2,\n});',
      output:
        '\n// TSFunctionType\nconst foo: () => void = () => ({\n    a: 1,\n    b: 2,\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSFunctionType\nconst foo: () => {\na: number,\nb: number,\n} = () => ({\na: 1,\nb: 2,\n});',
      output:
        '\n// TSFunctionType\nconst foo: () => {\n    a: number,\n    b: number,\n} = () => ({\n    a: 1,\n    b: 2,\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSFunctionType\nconst foo: ({\na: number,\nb: number,\n}) => void = (arg) => ({\na: 1,\nb: 2,\n});',
      output:
        '\n// TSFunctionType\nconst foo: ({\n    a: number,\n    b: number,\n}) => void = (arg) => ({\n    a: 1,\n    b: 2,\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSFunctionType\nconst foo: ({\na: number,\nb: number,\n}) => {\na: number,\nb: number,\n} = (arg) => ({\na: arg.a,\nb: arg.b,\n});',
      output:
        '\n// TSFunctionType\nconst foo: ({\n    a: number,\n    b: number,\n}) => {\n    a: number,\n    b: number,\n} = (arg) => ({\n    a: arg.a,\n    b: arg.b,\n});',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 10,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 11,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSImportType\nconst foo: import("bar") = {\na: 1,\nb: 2,\n};',
      output:
        '\n// TSImportType\nconst foo: import("bar") = {\n    a: 1,\n    b: 2,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSImportType\nconst foo: import(\n"bar"\n) = {\na: 1,\nb: 2,\n};',
      output:
        '\n// TSImportType\nconst foo: import(\n    "bar"\n) = {\n    a: 1,\n    b: 2,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: "\n// TSIndexedAccessType\ntype Foo = Bar[\n'asdf'\n];",
      output: "\n// TSIndexedAccessType\ntype Foo = Bar[\n    'asdf'\n];",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSIndexSignature\ntype Foo = {\n[a : string] : {\nx : foo\n[b : number] : boolean\n}\n}',
      output:
        '\n// TSIndexSignature\ntype Foo = {\n    [a : string] : {\n        x : foo\n        [b : number] : boolean\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSInferType\ntype Foo<T> = T extends string\n? infer U\n: {\na : string\n};',
      output:
        '\n// TSInferType\ntype Foo<T> = T extends string\n    ? infer U\n    : {\n        a : string\n    };',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSInterfaceBody, TSInterfaceDeclaration\ninterface Foo {\na : string\nb : {\nc : number\nd : boolean\n}\n}',
      output:
        '\n// TSInterfaceBody, TSInterfaceDeclaration\ninterface Foo {\n    a : string\n    b : {\n        c : number\n        d : boolean\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSInterfaceHeritage\ninterface Foo extends Bar {\na : string\nb : {\nc : number\nd : boolean\n}\n}',
      output:
        '\n// TSInterfaceHeritage\ninterface Foo extends Bar {\n    a : string\n    b : {\n        c : number\n        d : boolean\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSIntersectionType\ntype Foo = "string" & {\na : number\n} & number;',
      output:
        '\n// TSIntersectionType\ntype Foo = "string" & {\n    a : number\n} & number;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: "\n// TSImportEqualsDeclaration, TSExternalModuleReference\nimport foo = require(\n'asdf'\n);",
      output:
        "\n// TSImportEqualsDeclaration, TSExternalModuleReference\nimport foo = require(\n    'asdf'\n);",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMappedType\ntype Partial<T> = {\n[P in keyof T];\n}',
      output: '\n// TSMappedType\ntype Partial<T> = {\n    [P in keyof T];\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMappedType\ntype Partial<T> = {\n[P in keyof T]: T[P];\n}',
      output:
        '\n// TSMappedType\ntype Partial<T> = {\n    [P in keyof T]: T[P];\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMappedType\ntype Partial<T> = {\nreadonly [P in keyof T]: T[P];\n}',
      output:
        '\n// TSMappedType\ntype Partial<T> = {\n    readonly [P in keyof T]: T[P];\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMappedType\n// TSQuestionToken\ntype Partial<T> = {\n[P in keyof T]?: T[P];\n}',
      output:
        '\n// TSMappedType\n// TSQuestionToken\ntype Partial<T> = {\n    [P in keyof T]?: T[P];\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMappedType\n// TSPlusToken\ntype Partial<T> = {\n[P in keyof T]+?: T[P];\n}',
      output:
        '\n// TSMappedType\n// TSPlusToken\ntype Partial<T> = {\n    [P in keyof T]+?: T[P];\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMappedType\n// TSMinusToken\ntype Partial<T> = {\n[P in keyof T]-?: T[P];\n}',
      output:
        '\n// TSMappedType\n// TSMinusToken\ntype Partial<T> = {\n    [P in keyof T]-?: T[P];\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSMethodSignature\ninterface Foo {\nmethod() : string\nmethod2() : {\na : number\nb : string\n}\n}',
      output:
        '\n// TSMethodSignature\ninterface Foo {\n    method() : string\n    method2() : {\n        a : number\n        b : string\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSModuleBlock, TSModuleDeclaration\ndeclare module "foo" {\nexport const bar : {\na : string,\nb : number,\n}\n}',
      output:
        '\n// TSModuleBlock, TSModuleDeclaration\ndeclare module "foo" {\n    export const bar : {\n        a : string,\n        b : number,\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSNonNullExpression\nconst foo = a!\n.b!.\nc;',
      output: '\n// TSNonNullExpression\nconst foo = a!\n    .b!.\n    c;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: "\n// TSParameterProperty\nclass Foo {\nconstructor(\nprivate foo : string,\npublic bar : {\na : string,\nb : number,\n}\n) {\nconsole.log('foo')\n}\n}",
      output:
        "\n// TSParameterProperty\nclass Foo {\n    constructor(\n        private foo : string,\n        public bar : {\n            a : string,\n            b : number,\n        }\n    ) {\n        console.log('foo')\n    }\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 9,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 10,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 11,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 12,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSPropertySignature\ninterface Foo {\nbar : string\nbaz : {\na : string\nb : number\n}\n}',
      output:
        '\n// TSPropertySignature\ninterface Foo {\n    bar : string\n    baz : {\n        a : string\n        b : number\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSQualifiedName\nconst a: Foo.bar = {\na: 1,\nb: 2,\n};',
      output:
        '\n// TSQualifiedName\nconst a: Foo.bar = {\n    a: 1,\n    b: 2,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSQualifiedName\nconst a: Foo.\nbar\n.baz = {\na: 1,\nb: 2,\n};',
      output:
        '\n// TSQualifiedName\nconst a: Foo.\n    bar\n    .baz = {\n        a: 1,\n        b: 2,\n    };',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSRestType\ntype foo = [\nstring,\n...string[],\n];',
      output:
        '\n// TSRestType\ntype foo = [\n    string,\n    ...string[],\n];',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSThisType\ndeclare class MyArray<T> extends Array<T> {\nsort(compareFn?: (a: T, b: T) => number): this;\nmeth() : {\na: number,\n}\n}',
      output:
        '\n// TSThisType\ndeclare class MyArray<T> extends Array<T> {\n    sort(compareFn?: (a: T, b: T) => number): this;\n    meth() : {\n        a: number,\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSTupleType\ntype foo = [\nstring,\nnumber,\n];',
      output: '\n// TSTupleType\ntype foo = [\n    string,\n    number,\n];',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSTupleType\ntype foo = [\n[\nstring,\nnumber,\n],\n];',
      output:
        '\n// TSTupleType\ntype foo = [\n    [\n        string,\n        number,\n    ],\n];',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSTypeOperator\ntype T = keyof {\na: 1,\nb: 2,\n};',
      output: '\n// TSTypeOperator\ntype T = keyof {\n    a: 1,\n    b: 2,\n};',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSTypeParameter, TSTypeParameterDeclaration\ntype Foo<T> = {\na : unknown,\nb : never,\n}',
      output:
        '\n// TSTypeParameter, TSTypeParameterDeclaration\ntype Foo<T> = {\n    a : unknown,\n    b : never,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: "\n// TSTypeParameter, TSTypeParameterDeclaration\nfunction foo<\nT,\nU\n>() {\nconsole.log('');\n}",
      output:
        "\n// TSTypeParameter, TSTypeParameterDeclaration\nfunction foo<\n    T,\n    U\n>() {\n    console.log('');\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: '\n// TSUnionType\ntype Foo = string | {\na : number\n} | number;',
      output:
        '\n// TSUnionType\ntype Foo = string | {\n    a : number\n} | number;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: 'type Foo = {\nbar : string,\nage : number,\n}',
      output: 'type Foo = {\n    bar : string,\n    age : number,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: 'interface Foo {\nbar : string,\nage : number,\nfoo(): boolean,\nbaz(\nasdf: string,\n): boolean,\nnew(): Foo,\nnew(\nasdf: string,\n): Foo,\n}',
      output:
        'interface Foo {\n    bar : string,\n    age : number,\n    foo(): boolean,\n    baz(\n        asdf: string,\n    ): boolean,\n    new(): Foo,\n    new(\n        asdf: string,\n    ): Foo,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 10,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 11,
          column: 1,
        },
      ],
    },
    {
      code: 'interface Foo {\nbar : {\nbaz : string,\n},\nage : number,\n}',
      output:
        'interface Foo {\n    bar : {\n        baz : string,\n    },\n    age : number,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
      ],
    },
    {
      code: 'interface Foo extends Bar {\nbar : string,\nage : number,\n}',
      output:
        'interface Foo extends Bar {\n    bar : string,\n    age : number,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: 'class Foo\nextends Bar {\nbar : string = "asdf";\nage : number = 1;\n}',
      output:
        'class Foo\n    extends Bar {\n    bar : string = "asdf";\n    age : number = 1;\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: 'interface Foo\nextends Bar {\nbar : string,\nage : number,\n}',
      output:
        'interface Foo\n    extends Bar {\n    bar : string,\n    age : number,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: 'const foo : Foo<{\nbar : string,\nage : number,\n}>',
      output: 'const foo : Foo<{\n    bar : string,\n    age : number,\n}>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: 'type FooAlias = Foo<\nBar,\nBaz\n>',
      output: 'type FooAlias = Foo<\n    Bar,\n    Baz\n>',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: 'type T = {\nbar : string,\nage : number,\n} | {\nbar : string,\nage : number,\n}',
      output:
        'type T = {\n    bar : string,\n    age : number,\n} | {\n    bar : string,\n    age : number,\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
      ],
    },
    {
      code: 'type T =\n    | {\nbar : string,\nage : number,\n}\n    | {\n    bar : string,\n    age : number,\n}',
      output:
        'type T =\n    | {\n        bar : string,\n        age : number,\n    }\n    | {\n        bar : string,\n        age : number,\n    }',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
          column: 1,
        },
      ],
    },
    {
      code: '    import Dialogs = require("widgets/Dialogs");',
      output: 'import Dialogs = require("widgets/Dialogs");',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 1,
          column: 1,
        },
      ],
    },
    {
      code: '\n    import Dialogs =\n      require("widgets/Dialogs");\n      ',
      output: '\nimport Dialogs =\n    require("widgets/Dialogs");\n      ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: 'class Foo {\npublic bar : string;\nprivate bar : string;\nprotected bar : string;\nabstract bar : string;\nfoo : string;\nconstructor() {\nconst foo = "";\n}\nconstructor(\nasdf : number,\nprivate test : boolean,\n) {}\n}',
      output:
        'class Foo {\n    public bar : string;\n    private bar : string;\n    protected bar : string;\n    abstract bar : string;\n    foo : string;\n    constructor() {\n        const foo = "";\n    }\n    constructor(\n        asdf : number,\n        private test : boolean,\n    ) {}\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 7,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 8,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 9,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 10,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 11,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 12,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 13,
          column: 1,
        },
      ],
    },
    {
      code: '\n    abstract class Foo {}\n    class Foo {}\n      ',
      output: '\nabstract class Foo {}\nclass Foo {}\n      ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: "enum Foo {\nbar,\nbaz = 1,\nbuzz = '',\n}",
      output: "enum Foo {\n    bar,\n    baz = 1,\n    buzz = '',\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: "enum Foo\n    {\n    bar,\n    baz = 1,\n    buzz = '',\n    }",
      output: "enum Foo\n{\n    bar,\n    baz = 1,\n    buzz = '',\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 6,
          column: 1,
        },
      ],
    },
    {
      code: "const enum Foo {\nbar,\nbaz = 1,\nbuzz = '',\n}",
      output: "const enum Foo {\n    bar,\n    baz = 1,\n    buzz = '',\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '\n    export = Foo;\n      ',
      output: '\nexport = Foo;\n      ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: '\n    declare function h(x: number): number;\n      ',
      output: '\ndeclare function h(x: number): number;\n      ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: 'declare function h(\nx: number,\n): number;',
      output: 'declare function h(\n    x: number,\n): number;',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
      ],
    },
    {
      code: 'namespace Validation {\nexport interface StringValidator {\nisAcceptable(s: string): boolean;\n}\n}',
      output:
        'namespace Validation {\n    export interface StringValidator {\n        isAcceptable(s: string): boolean;\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: 'declare module "Validation" {\nexport interface StringValidator {\nisAcceptable(s: string): boolean;\n}\n}',
      output:
        'declare module "Validation" {\n    export interface StringValidator {\n        isAcceptable(s: string): boolean;\n    }\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: '    @Decorator()\nclass Foo {\n    @a\n        foo: any;\n\n@b @c()\n    bar: any;\n\n        @d baz: any;\n}',
      output:
        '@Decorator()\nclass Foo {\n    @a\n    foo: any;\n\n    @b @c()\n    bar: any;\n\n    @d baz: any;\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 4 },
          line: 1,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 8 },
          line: 9,
          column: 1,
        },
      ],
    },
    {
      code: 'class Foo {\n    @a\n      func() {}\n  @b\n    get bar() { return }\n  @c\n  baz: () => 1\n}',
      output:
        'class Foo {\n    @a\n    func() {}\n    @b\n    get bar() { return }\n    @c\n    baz: () => 1\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 6,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 7,
          column: 1,
        },
      ],
    },
    {
      code: 'class Foo {\n    bar =\n"baz";\n}',
      output: 'class Foo {\n    bar =\n        "baz";\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
          column: 1,
        },
      ],
    },
    {
      code: "class Foo {\n    func(\n            @Param('foo') foo: string,\n    @Param('bar') bar: string,\n        @Param('baz') baz: string\n    ) {\n        return { foo, bar, baz };\n    }\n}",
      output:
        "class Foo {\n    func(\n        @Param('foo') foo: string,\n        @Param('bar') bar: string,\n        @Param('baz') baz: string\n    ) {\n        return { foo, bar, baz };\n    }\n}",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 12 },
          line: 3,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: 'class Foo {\naccessor [bar]: string\n    accessor baz: number\n}',
      output:
        'class Foo {\n  accessor [bar]: string\n  accessor baz: number\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
        },
      ],
    },
    {
      code: 'abstract class Foo {\nabstract protected bar: number\n    abstract accessor [baz]: string\n}',
      output:
        'abstract class Foo {\n  abstract protected bar: number\n  abstract accessor [baz]: string\n}',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
        },
      ],
    },
    {
      code: ' type A = number\n  declare type B = number\nnamespace Foo {\n      declare type C = number\n}',
      output:
        'type A = number\ndeclare type B = number\nnamespace Foo {\n    declare type C = number\n}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 1 },
          line: 1,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
          line: 2,
          column: 1,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 4,
          column: 1,
        },
      ],
    },
    {
      code: 'using a = foo(),\n  b = bar();\nawait using c = baz(),\n  d = qux();',
      output:
        'using a = foo(),\n      b = bar();\nawait using c = baz(),\n            d = qux();',
      options: [2, { VariableDeclarator: 'first' }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 2 },
        },
      ],
    },
    {
      code: 'using a = foo(),\n      b = bar();\nawait using c = baz(),\n            d = qux();',
      output:
        'using a = foo(),\n  b = bar();\nawait using c = baz(),\n  d = qux();',
      options: [2, { VariableDeclarator: { using: 1 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 6 },
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 12 },
        },
      ],
    },
    {
      code: 'export function isAuthenticated(authResult: AuthenticationResult | null | undefined)\n: authResult is SuccessAuthenticationResult {\n  return !! authResult && authResult.isAuthenticated();\n}',
      output:
        'export function isAuthenticated(authResult: AuthenticationResult | null | undefined)\n    : authResult is SuccessAuthenticationResult {\n  return !! authResult && authResult.isAuthenticated();\n}',
      options: [2, { FunctionDeclaration: { returnType: 2 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
        },
      ],
    },
    {
      code: "const foo = function(a: string)\n  : a is 'a' {\n}",
      output: "const foo = function(a: string)\n: a is 'a' {\n}",
      options: [2, { FunctionExpression: { returnType: 0 } }],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '0 spaces', actual: 2 },
        },
      ],
    },
    {
      code: 'function foo<\n    T\n    =\n    Foo\n>() {}',
      output: 'function foo<\n    T\n        =\n            Foo\n>() {}',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 4 },
        },
      ],
    },
    {
      code: "import foo\n=\nrequire('source')",
      output: "import foo\n    =\n        require('source')",
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
        },
      ],
    },

    // ---- from indent._jsx_.test.ts (@typescript-eslint/parser variant) ----
    {
      code: '<App>\n  <Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<App>\n    <Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: '<App>\n  <></>\n</App>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        '<App>\n    <></>\n</App>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: '<>\n  <Foo />\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
      output:
        '<>\n    <Foo />\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 2,
        },
      ],
    },
    {
      code: '<App>\n    <Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        '<App>\n  <Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 2,
        },
      ],
    },
    {
      code: '<App>\n    <Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        '<App>\n\t<Foo />\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: '4 spaces' },
          line: 2,
        },
      ],
    },
    {
      code: 'function App() {\n  return <App>\n    <Foo />\n         </App>;\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'function App() {\n  return <App>\n    <Foo />\n  </App>;\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 9 },
          line: 4,
        },
      ],
    },
    {
      code: 'function App() {\n  return (<App>\n    <Foo />\n    </App>);\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'function App() {\n  return (<App>\n    <Foo />\n  </App>);\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'function App() {\n  return (\n<App>\n  <Foo />\n</App>\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'function App() {\n  return (\n    <App>\n      <Foo />\n    </App>\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '6 spaces', actual: 2 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: '<App>\n   {test}\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<App>\n    {test}\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 3 },
          line: 2,
        },
      ],
    },
    {
      code: '<App>\n    {options.map((option, index) => (\n        <option key={index} value={option.key}>\n           {option.name}\n        </option>\n    ))}\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<App>\n    {options.map((option, index) => (\n        <option key={index} value={option.key}>\n            {option.name}\n        </option>\n    ))}\n</App>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 11 },
          line: 4,
        },
      ],
    },
    {
      code: '<App>\n{test}\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        '<App>\n\t{test}\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: '<App>\n\t{options.map((option, index) => (\n\t\t<option key={index} value={option.key}>\n\t\t{option.name}\n\t\t</option>\n\t))}\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        '<App>\n\t{options.map((option, index) => (\n\t\t<option key={index} value={option.key}>\n\t\t\t{option.name}\n\t\t</option>\n\t))}\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '3 tabs', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: '<App>\n\n<Foo />\n\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        '<App>\n\n\t<Foo />\n\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '[\n  <div />,\n    <div />\n]\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        '[\n  <div />,\n  <div />\n]\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '[\n  <div />,\n    <></>\n]\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        '[\n  <div />,\n  <></>\n]\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: '<App>\n\n <Foo />\n\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        '<App>\n\n\t<Foo />\n\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '1 tab', actual: '1 space' },
          line: 3,
        },
      ],
    },
    {
      code: '<App>\n\n\t<Foo />\n\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        '<App>\n\n  <Foo />\n\n</App>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: '1 tab' },
          line: 3,
        },
      ],
    },
    {
      code: '<div>\n    {\n        [\n            <Foo />,\n        <Bar />\n        ]\n    }\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<div>\n    {\n        [\n            <Foo />,\n            <Bar />\n        ]\n    }\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 5,
        },
      ],
    },
    {
      code: '<div>\n    {foo &&\n        [\n            <Foo />,\n        <Bar />\n        ]\n    }\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<div>\n    {foo &&\n        [\n            <Foo />,\n            <Bar />\n        ]\n    }\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '12 spaces', actual: 8 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo /> :\n<Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo /> :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo /> :\n<></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo /> :\n    <></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo />\n    :\n<Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo />\n    :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo />\n    :\n<></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo />\n    :\n    <></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? <Foo />\n    :\n      <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? <Foo />\n    :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 6 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ? (\n    <Foo />\n) :\n<Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <Foo />\n) :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? (\n    <Foo />\n) :\n<></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <Foo />\n) :\n    <></>\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? (\n    <Foo />\n)\n    :\n<Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <Foo />\n)\n    :\n    <Bar />\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo /> : (\n    <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo /> : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo /> : (\n    <></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo /> : (\n        <></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo />\n    : (\n<Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo />\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo />\n    : (\n    <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo />\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ?\n    <Foo />\n    : (\n    <></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ?\n    <Foo />\n    : (\n        <></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 4 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? (\n<Foo />\n) : (\n<Bar />\n)\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <Foo />\n) : (\n    <Bar />\n)\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? (\n<></>\n) : (\n<></>\n)\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <></>\n) : (\n    <></>\n)\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: 'foo ? (\n<Foo />\n)\n    : (\n<Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <Foo />\n)\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'foo ? (\n<Foo />\n)\n    :\n    (\n<Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <Foo />\n)\n    :\n    (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'foo ? (\n<></>\n)\n    :\n    (\n<></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? (\n    <></>\n)\n    :\n    (\n        <></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 6,
        },
      ],
    },
    {
      code: 'foo ? <Foo /> : (\n<Bar />\n)\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? <Foo /> : (\n    <Bar />\n)\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo ? <Foo /> : (\n<></>\n)\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? <Foo /> : (\n    <></>\n)\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 2,
        },
      ],
    },
    {
      code: 'foo ? <Foo />\n    : (\n<Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? <Foo />\n    : (\n        <Bar />\n    )\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: 'foo ? <Foo />\n    : (\n<></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      output:
        'foo ? <Foo />\n    : (\n        <></>\n    )\n// features: [fragment,no-ts-old], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 0 },
          line: 3,
        },
      ],
    },
    {
      code: '<p>\n    <div>\n        <SelfClosingTag />Text\n  </div>\n</p>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<p>\n    <div>\n        <SelfClosingTag />Text\n    </div>\n</p>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 2 },
          line: 4,
        },
      ],
    },
    {
      code: 'const Component = () => (\n  <View\n    ListFooterComponent={(\n      <View\n        rowSpan={3}\n        placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n      />\n)}\n  />\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'const Component = () => (\n  <View\n    ListFooterComponent={(\n      <View\n        rowSpan={3}\n        placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n      />\n    )}\n  />\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 8,
        },
      ],
    },
    {
      code: 'const Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n\t\t\t/>\n)}\n\t/>\n);\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        'const Component = () => (\n\t<View\n\t\tListFooterComponent={(\n\t\t\t<View\n\t\t\t\trowSpan={3}\n\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"\n\t\t\t/>\n\t\t)}\n\t/>\n);\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 tabs', actual: 0 },
          line: 8,
        },
      ],
    },
    {
      code: 'function Foo() {\n  return (\n    <div>\n      {condition && (\n      <p>Bar</p>\n      )}\n    </div>\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'function Foo() {\n  return (\n    <div>\n      {condition && (\n        <p>Bar</p>\n      )}\n    </div>\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '8 spaces', actual: 6 },
        },
      ],
    },
    {
      code: '<div>\ntext\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        '<div>\n    text\n</div>\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 1,
        },
      ],
    },
    {
      code: '<div>\n\t\ttext\n</div>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      output:
        '<div>\n\ttext\n</div>\n// features: [], parser: @typescript-eslint/parser, , options: ["tab"], ',
      options: ['tab'],
      errors: [{ messageId: 'wrongIndentation' }],
    },
    {
      code: '<>\naaa\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
      output:
        '<>\n    aaa\n</>\n// features: [fragment], parser: @typescript-eslint/parser, , , ',
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 1,
        },
      ],
    },
    {
      code: 'const StatelessComponent = () => {\n    if (new Date() % 2) {\n        return (\n  <div>Hello</div>\n        );\n    }\n    return null;\n};\n// features: [], parser: @typescript-eslint/parser, , , ',
      output:
        'const StatelessComponent = () => {\n    if (new Date() % 2) {\n        return (\n            <div>Hello</div>\n        );\n    }\n    return null;\n};\n// features: [], parser: @typescript-eslint/parser, , , ',
      errors: [{ messageId: 'wrongIndentation' }],
    },
    {
      code: 'function App() {\n  return (\n    <App />\n    );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'function App() {\n  return (\n    <App />\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [{ message: 'Expected indentation of 2 spaces but found 4.' }],
    },
    {
      code: 'function App() {\n  return (\n    <App />\n);\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'function App() {\n  return (\n    <App />\n  );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [{ message: 'Expected indentation of 2 spaces but found 0.' }],
    },
    {
      code: '{condition && [\n    <Tag key="a" onClick={() => {\n      // some code\n    }} />,\n    <Tag key="b" onClick={() => {\n      // some code\n    }} />,\n]}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        '{condition && [\n  <Tag key="a" onClick={() => {\n    // some code\n  }} />,\n  <Tag key="b" onClick={() => {\n    // some code\n  }} />,\n]}\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        { message: 'Expected indentation of 2 spaces but found 4.', line: 2 },
        { message: 'Expected indentation of 4 spaces but found 6.', line: 3 },
        { message: 'Expected indentation of 2 spaces but found 4.', line: 4 },
        { message: 'Expected indentation of 2 spaces but found 4.', line: 5 },
        { message: 'Expected indentation of 4 spaces but found 6.', line: 6 },
        { message: 'Expected indentation of 2 spaces but found 4.', line: 7 },
      ],
    },
    {
      code: 'const IndexPage = () => (\n  <h1>\n{"Hi people"}\n<button/>\n</h1>\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'const IndexPage = () => (\n  <h1>\n    {"Hi people"}\n    <button/>\n  </h1>\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 3,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
        {
          messageId: 'wrongIndentation',
          data: { expected: '2 spaces', actual: 0 },
          line: 5,
        },
      ],
    },
    {
      code: 'const IndexPage = () => (\n  <h1>\n    Hi people\n<button/>\n  </h1>\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        'const IndexPage = () => (\n  <h1>\n    Hi people\n    <button/>\n  </h1>\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
      errors: [
        {
          messageId: 'wrongIndentation',
          data: { expected: '4 spaces', actual: 0 },
          line: 4,
        },
      ],
    },
    {
      code: '<App\n  foo\n  =\n  "bar"\n/>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      output:
        '<App\n  foo\n    =\n      "bar"\n/>\n// features: [], parser: @typescript-eslint/parser, , options: [2], ',
      options: [2],
    },
  ],
});

// ============================ indent — KNOWN GAPS ============================
//
// The cases below are ported verbatim from upstream but are NOT run through the green
// ruleTester.run above. Each is a real rslint<->upstream divergence, not a silenced
// failure. The expected upstream behaviour is preserved for the record.
//
// ---- GAP 1: ts-go syntax errors (parser-level) ----
// rslint parses every fixture with the ts-go parser, which rejects these forms as
// SYNTAX ERRORS (and the rslint CLI aborts JSONL for the whole batch on any syntax
// error, which is why they must live outside the green set). The indent rule logic is
// not at fault; the input is unparseable under ts-go.
//
// (a) JSX with a line break / comment splitting an opening or closing tag, or a
//     multi-line fragment `<>...<\n/>`. ts-go reports TS17014 / TS1003 / TS1005 /
//     TS1109 / TS1161 (JSX fragment / element has no corresponding closing tag, etc.).
//     Upstream parses these with espree/@typescript-eslint where the newline-in-tag is
//     tolerated. 18 cases:
//       {"code":"<a\n    href=\"foo\">\n    bar\n</\n    a>;"}
//       {"code":"var foo = <a\n    href=\"bar\">\n    baz\n</\n    a>;"}
//       {"code":"<\n    foo\n        .bar\n        .baz\n>\n    foo\n</\n    foo.\n        bar.\n        baz\n>"}
//       {"code":"<>\n    <A />\n<\n/>"}
//       {"code":"<\n>\n    <A />\n<\n/>"}
//       {"code":"<>\n    <A />\n< // Comment\n/>"}
//       {"code":"<>\n    <A />\n<\n    // Comment\n/>"}
//       {"code":"<>\n    <A />\n<\n// Comment\n/>"}
//       {"code":"<>\n    <A />\n< /* Comment */\n/>"}
//       {"code":"<>\n    <A />\n<\n    /* Comment */ />"}
//       {"code":"<>\n    <A />\n<\n/* Comment */ />"}
//       {"code":"<>\n    <A />\n<\n    /* Comment */\n/>"}
//       {"code":"<>\n    <A />\n<\n/* Comment */\n/>"}
//       {"code":"<\n    foo\n    .bar\n    .baz\n>\n    foo\n</\n    foo.\n    bar.\n    baz\n>","output":"<\n    foo\n        .bar\n        .baz\n>\n    foo\n</\n    foo.\n        bar.\n        baz\n>","errors":[{"messageId":"wrongIndentation","data":{"expected":"8 spaces","actual":4},"line":3},{"messageId":"wrongIndentation","data":{"expected":"8 spaces","actual":4},"line":4},{"messageId":"wrongIndentation","data":{"expected":"8 spaces","actual":4},"line":9},{"messageId":"wrongIndentation","data":{"expected":"8 spaces","actual":4},"line":10}]}
//       {"code":"<>\n    <A />\n<\n    />","output":"<>\n    <A />\n<\n/>","errors":[{"messageId":"wrongIndentation","data":{"expected":"0 spaces","actual":4},"line":4}]}
//       {"code":"<\n    >\n    <A />\n<\n    />","output":"<\n>\n    <A />\n<\n/>","errors":[{"messageId":"wrongIndentation","data":{"expected":"0 spaces","actual":4},"line":2},{"messageId":"wrongIndentation","data":{"expected":"0 spaces","actual":4},"line":5}]}
//       {"code":"<>\n    <A />\n< // Comment\n    />","output":"<>\n    <A />\n< // Comment\n/>","errors":[{"messageId":"wrongIndentation","data":{"expected":"0 spaces","actual":4},"line":4}]}
//       {"code":"<>\n    <A />\n< /* Comment */\n    />","output":"<>\n    <A />\n< /* Comment */\n/>","errors":[{"messageId":"wrongIndentation","data":{"expected":"0 spaces","actual":4},"line":4}]}
//
// (b) Import attributes on an `export ... from` statement (`export {...} from '...'
//     with { type: 'json' }`) and the `export * as from` / `export { v as from }`
//     "form"-keyword edge cases. ts-go reports TS1005. 8 cases:
//       {"code":"export {v}\n    from \"./foo.json\"\n    with {\n        type: \"json\"\n    };"}
//       {"code":"export {v} from \"./foo.json\"\n    with\n    {\n        type\n        :\n        \"json\"\n    };"}
//       {"code":"export * as json\n    from \"./foo.json\"\n    with {\n        type: \"json\"\n    };"}
//       {"code":"export * as json from \"./foo.json\"\n    with\n    {\n        type\n        :\n        \"json\"\n    };"}
//       {"code":"// should correctly recognize a `form` token\nexport {\n    v\n    as\n    from\n}\n    from \"./foo\"\n    with {\n        from: \"foo\"\n    };"}
//       {"code":"// should correctly recognize a `form` token\nexport\n*\nas\nfrom\n    from \"./foo\"\n    with {\n        from: \"foo\"\n    };"}
//       {"code":"export {v} from \"./foo.json\"\nwith\n{\ntype\n:\n\"json\"\n};","output":"export {v} from \"./foo.json\"\n    with\n    {\n        type\n        :\n        \"json\"\n    };"}
//       {"code":"export * as json from \"./foo.json\"\nwith\n{\ntype\n:\n\"json\"\n};","output":"export * as json from \"./foo.json\"\n    with\n    {\n        type\n        :\n        \"json\"\n    };"}
//
// ---- GAP 2: multi-pass fix vs single-pass output ----
// Diagnostics match upstream exactly (same count, messages, lines), but the autofix
// OUTPUT differs: ESLint's RuleTester pins a SINGLE fix pass, whereas rslint --fix runs
// to a stable fixpoint (multi-pass). For these JSX cases the first pass shifts the
// context of a later line, so rslint corrects lines upstream leaves untouched. 2 cases
// (errors are identical and would pass; only the output diverges):
//   code:     "const IndexPage = () => (\n  <h1>\nHi people\n<button/>\n</h1>\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], "
//   upstream: "const IndexPage = () => (\n  <h1>\n    Hi people\n<button/>\n  </h1>\n);\n// features: [], parser: @typescript-eslint/parser, , options: [2], "
//   errors:   [{"messageId":"wrongIndentation","data":{"expected":"4 spaces","actual":0},"line":2},{"messageId":"wrongIndentation","data":{"expected":"4 spaces","actual":0},"line":4},{"messageId":"wrongIndentation","data":{"expected":"2 spaces","actual":0},"line":5}]
//   code:     "import React from 'react';\n\nexport default function () {\n    return (\n        <div>\n                    Test1\n\n              <p>Test2</p>\n        </div>\n    );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [4], "
//   upstream: "import React from 'react';\n\nexport default function () {\n    return (\n        <div>\n            Test1\n\n              <p>Test2</p>\n        </div>\n    );\n}\n// features: [], parser: @typescript-eslint/parser, , options: [4], "
//   errors:   [{"messageId":"wrongIndentation","line":5},{"messageId":"wrongIndentation","line":8}]
//
// ---- GAP 3: Babel/Flow (upstream skips these too under ESLint >= 10) ----
// The upstream `if (!skipBabel) run({ name: 'indent_babel', ... })` block (Flow-typed
// arrow params via @babel/eslint-parser + languageOptionsForBabelFlow). `skipBabel` is
// true on the installed ESLint, so upstream never runs them; rslint has no Babel/Flow
// parser either. 3 invalid cases:
//   { code: '({\n    foo\n    }: bar) => baz', output: '({\n    foo\n}: bar) => baz', errors: [{ messageId: 'wrongIndentation', data: { expected: '0 spaces', actual: 4 }, line: 3 }] }
//   { code: '([\n    foo\n    ]: bar) => baz', output: '([\n    foo\n]: bar) => baz', errors: [{ messageId: 'wrongIndentation', data: { expected: '0 spaces', actual: 4 }, line: 3 }] }
//   { code: '({\n    foo\n    }: {}) => baz',  output: '({\n    foo\n}: {}) => baz',  errors: [{ messageId: 'wrongIndentation', data: { expected: '0 spaces', actual: 4 }, line: 3 }] }
//
// ============================================================================
