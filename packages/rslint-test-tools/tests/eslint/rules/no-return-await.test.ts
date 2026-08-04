import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-return-await', {
  valid: [
    '\nasync function foo() {\n\tawait bar(); return;\n}\n',
    '\nasync function foo() {\n\tconst x = await bar(); return x;\n}\n',
    '\nasync () => { return bar(); }\n',
    '\nasync () => bar()\n',
    '\nasync function foo() {\nif (a) {\n\t\tif (b) {\n\t\t\treturn bar();\n\t\t}\n\t}\n}\n',
    '\nasync () => {\nif (a) {\n\t\tif (b) {\n\t\t\treturn bar();\n\t\t}\n\t}\n}\n',
    '\nasync function foo() {\n\treturn (await bar() && a);\n}\n',
    '\nasync function foo() {\n\treturn (await bar() || a);\n}\n',
    '\nasync function foo() {\n\treturn (a && await baz() && b);\n}\n',
    '\nasync function foo() {\n\treturn (await bar(), a);\n}\n',
    '\nasync function foo() {\n\treturn (await baz(), await bar(), a);\n}\n',
    '\nasync function foo() {\n\treturn (a, b, (await bar(), c));\n}\n',
    '\nasync function foo() {\n\treturn (await bar() ? a : b);\n}\n',
    '\nasync function foo() {\n\treturn ((a && await bar()) ? b : c);\n}\n',
    '\nasync function foo() {\n\treturn (baz() ? (await bar(), a) : b);\n}\n',
    '\nasync function foo() {\n\treturn (baz() ? (await bar() && a) : b);\n}\n',
    '\nasync function foo() {\n\treturn (baz() ? a : (await bar(), b));\n}\n',
    '\nasync function foo() {\n\treturn (baz() ? a : (await bar() && b));\n}\n',
    '\nasync () => (await bar(), a)\n',
    '\nasync () => (await bar() && a)\n',
    '\nasync () => (await bar() || a)\n',
    '\nasync () => (a && await bar() && b)\n',
    '\nasync () => (await baz(), await bar(), a)\n',
    '\nasync () => (a, b, (await bar(), c))\n',
    '\nasync () => (await bar() ? a : b)\n',
    '\nasync () => ((a && await bar()) ? b : c)\n',
    '\nasync () => (baz() ? (await bar(), a) : b)\n',
    '\nasync () => (baz() ? (await bar() && a) : b)\n',
    '\nasync () => (baz() ? a : (await bar(), b))\n',
    '\nasync () => (baz() ? a : (await bar() && b))\n',
    '\n          async function foo() {\n            try {\n              return await bar();\n            } catch (e) {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {\n              return await bar();\n            } finally {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {}\n            catch (e) {\n              return await bar();\n            } finally {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {\n              try {}\n              finally {\n                return await bar();\n              }\n            } finally {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {\n              try {}\n              catch (e) {\n                return await bar();\n              }\n            } finally {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {\n              return (a, await bar());\n            } catch (e) {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {\n              return (qux() ? await bar() : b);\n            } catch (e) {\n              baz();\n            }\n          }\n        ',
    '\n          async function foo() {\n            try {\n              return (a && await bar());\n            } catch (e) {\n              baz();\n            }\n          }\n        ',
  ],
  invalid: [
    {
      code: '\nasync function foo() {\n\treturn await bar();\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 9 }],
    },
    {
      code: '\nasync function foo() {\n\treturn await(bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 9 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a, await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 13 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a, b, await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 16 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a && await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 15 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a && b && await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 20 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a || await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 15 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a, b, (c, d, await bar()));\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 23 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (a, b, (c && await bar()));\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 22 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (await baz(), b, await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 26 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (baz() ? await bar() : b);\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 18 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (baz() ? a : await bar());\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 22 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (baz() ? (a, await bar()) : b);\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 22 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (baz() ? a : (b, await bar()));\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 26 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (baz() ? (a && await bar()) : b);\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 24 }],
    },
    {
      code: '\nasync function foo() {\n\treturn (baz() ? a : (b && await bar()));\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 28 }],
    },
    {
      code: '\nasync () => { return await bar(); }\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 22 }],
    },
    {
      code: '\nasync () => await bar()\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 13 }],
    },
    {
      code: '\nasync () => (a, b, await bar())\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 20 }],
    },
    {
      code: '\nasync () => (a && await bar())\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 19 }],
    },
    {
      code: '\nasync () => (baz() ? await bar() : b)\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 22 }],
    },
    {
      code: '\nasync () => (baz() ? a : (b, await bar()))\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 30 }],
    },
    {
      code: '\nasync () => (baz() ? a : (b && await bar()))\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 2, column: 32 }],
    },
    {
      code: '\nasync function foo() {\nif (a) {\n\t\tif (b) {\n\t\t\treturn await bar();\n\t\t}\n\t}\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 5, column: 11 }],
    },
    {
      code: '\nasync () => {\nif (a) {\n\t\tif (b) {\n\t\t\treturn await bar();\n\t\t}\n\t}\n}\n',
      errors: [{ messageId: 'redundantUseOfAwait', line: 5, column: 11 }],
    },
    {
      code: '\n              async function foo() {\n                try {}\n                finally {\n                  return await bar();\n                }\n              }\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 5, column: 26 }],
    },
    {
      code: '\n              async function foo() {\n                try {}\n                catch (e) {\n                  return await bar();\n                }\n              }\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 5, column: 26 }],
    },
    {
      code: '\n              try {\n                async function foo() {\n                  return await bar();\n                }\n              } catch (e) {}\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 4, column: 26 }],
    },
    {
      code: '\n              try {\n                async () => await bar();\n              } catch (e) {}\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 29 }],
    },
    {
      code: '\n              async function foo() {\n                try {}\n                catch (e) {\n                  try {}\n                  catch (e) {\n                    return await bar();\n                  }\n                }\n              }\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 7, column: 28 }],
    },
    {
      code: '\n              async function foo() {\n                return await new Promise(resolve => {\n                  resolve(5);\n                });\n              }\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 24 }],
    },
    {
      code: '\n              async () => {\n                return await (\n                  foo()\n                )\n              };\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 24 }],
    },
    {
      code: '\n              async function foo() {\n                return await // Test\n                  5;\n              }\n            ',
      errors: [{ messageId: 'redundantUseOfAwait', line: 3, column: 24 }],
    },
  ],
});
