// https://github.com/jsx-eslint/eslint-plugin-react/blob/5d4bf128fb506f604f1bb61bcf76fa8caeea6a1a/tests/lib/rules/async-server-action.js
import { RuleTester, type ExpectedSuggestion } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectAsyncServerActionError(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'asyncServerAction',
    message: 'Server Actions must be async',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('async-server-action', {
  valid: [
    {
      code: `
        async function addToCart(data) {
          'use server';
        }
      `,
    },
    {
      code: `
        async function requestUsername(formData) {
          'use server';
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        async function addToCart(data) {
          "use server";
        }
      `,
    },
    {
      code: `
        async function requestUsername(formData) {
          "use server";
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        function addToCart(data) {
          console.log("test");
          'use server';
        }
      `,
    },
    {
      code: `
        function requestUsername(formData) {
          const username = formData.get('username');
          'use server';
        }
      `,
    },
    {
      code: `
        function addToCart(data) {
          console.log("use server");
        }
      `,
    },
    {
      code: `
        function requestUsername(formData) {
          console.log("use server");
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        const addToCart = async (data) => {
          'use server';
        }
      `,
    },
    {
      code: `
        const requestUsername = async (formData) => {
          'use server';
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        const addToCart = async (data) => {
          "use server";
        }
      `,
    },
    {
      code: `
        const requestUsername = async (formData) => {
          "use server";
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        const addToCart = (data) => {
          console.log("test");
          'use server';
        }
      `,
    },
    {
      code: `
        const requestUsername = (formData) => {
          const username = formData.get('username');
          'use server';
        }
      `,
    },
    {
      code: `
        const addToCart = (data) => {
          console.log("use server");
        }
      `,
    },
    {
      code: `
        const requestUsername = (formData) => {
          console.log("use server");
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        const addToCart = async function (data) {
          'use server';
        }
      `,
    },
    {
      code: `
        const requestUsername = async function (formData) {
          'use server';
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        const addToCart = async function (data) {
          "use server";
        }
      `,
    },
    {
      code: `
        const requestUsername = async function (formData) {
          "use server";
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        const addToCart = function (data) {
          console.log("test");
          'use server';
        }
      `,
    },
    {
      code: `
        const requestUsername = function (formData) {
          const username = formData.get('username');
          'use server';
        }
      `,
    },
    {
      code: `
        const addToCart = function (data) {
          console.log("use server");
        }
      `,
    },
    {
      code: `
        const requestUsername = function (formData) {
          console.log("use server");
          const username = formData.get('username');
        }
      `,
    },
    {
      code: `
        async function addToCart(data) {
          \`use server\`;
        }
      `,
    },
    {
      code: `
        function addToCart(data) {
          \`use server\`;
        }
      `,
    },
    {
      code: `
        const addToCart = async (data) => {
          \`use server\`;
        }
      `,
    },
    {
      code: `
        const addToCart = (data) => {
          \`use server\`;
        }
      `,
    },
    {
      code: `
        const addToCart = async function (data) {
          \`use server\`;
        }
      `,
    },
    {
      code: `
        const addToCart = function (data) {
          \`use server\`;
        }
      `,
    },
    {
      code: `
        const addToCart = async function* (data) {
          'use server';
        }
      `,
    },
    {
      code: `
        const addToCart = async function* (data) {
          "use server";
        }
      `,
    },
    {
      code: `
        const addToCart = function* (data) {
          'use server';
        }
      `,
    },
    {
      code: `
        const addToCart = function* (data) {
          "use server";
        }
      `,
    },
    {
      code: `
        function* addToCart(data) {
          'use server';
        }
      `,
    },
    {
      code: `
        async function* addToCart(data) {
          'use server';
        }
      `,
    },
    {
      code: `
        export async function addToCart(data) {
          'use server';
        }
      `,
    },
    {
      code: `
        export default async function addToCart(data) {
          'use server';
        }
      `,
    },
    {
      code: `
        export default async function (data) {
          'use server';
        }
      `,
    },
    {
      code: `
        const obj = {
          async action() {
            'use server';
          }
        };
      `,
    },
    {
      code: `
        const obj = {
          async action() {
            'use server';
            const x = 1;
          }
        };
      `,
    },
    {
      code: `
        class Foo {
          async action() {
            'use server';
          }
        }
      `,
    },
    {
      code: `
        class Foo {
          static async action() {
            'use server';
          }
        }
      `,
    },
    {
      code: `
        function outer() {
          async function inner() {
            'use server';
          }
        }
      `,
    },
    {
      code: `
        const action = async function named(data) {
          'use server';
        }
      `,
    },
    {
      code: `
        function addToCart(data) {
          'use strict';
          console.log('use server');
        }
      `,
    },
    {
      code: `
        function empty() {}
      `,
    },
    {
      code: `
        const fn = () => 'use server';
      `,
    },
    {
      code: `
        <form action={async () => { 'use server'; }} />
      `,
    },
    {
      code: `
        <button onClick={async () => { 'use server'; doSomething(); }} />
      `,
    },
    {
      code: `
        async function action() {
          'use strict';
          'use server';
        }
      `,
    },
    {
      code: `
        function action() {
          'use strict';
          'use server';
        }
      `,
    },
  ],

  invalid: [
    {
      code: `
        function addToCart(data) {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 9, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make `addToCart` an `async` function',
              output:
                "\n        async function addToCart(data) {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        function requestUsername(formData) {
          'use server';
          const username = formData.get('username');
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 9, 5, 10],
          [
            {
              messageId: '',
              desc: 'Make `requestUsername` an `async` function',
              output:
                "\n        async function requestUsername(formData) {\n          'use server';\n          const username = formData.get('username');\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        function addToCart(data) {
          "use server";
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 9, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make `addToCart` an `async` function',
              output:
                '\n        async function addToCart(data) {\n          "use server";\n        }\n      ',
            },
          ],
        ),
      ],
    },
    {
      code: `
        function requestUsername(formData) {
          "use server";
          const username = formData.get('username');
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 9, 5, 10],
          [
            {
              messageId: '',
              desc: 'Make `requestUsername` an `async` function',
              output:
                '\n        async function requestUsername(formData) {\n          "use server";\n          const username = formData.get(\'username\');\n        }\n      ',
            },
          ],
        ),
      ],
    },
    {
      code: `
        const addToCart = (data) => {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 27, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        const addToCart = async (data) => {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        const requestUsername = (formData) => {
          'use server';
          const username = formData.get('username');
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 33, 5, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        const requestUsername = async (formData) => {\n          'use server';\n          const username = formData.get('username');\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        const addToCart = (data) => {
          "use server";
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 27, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                '\n        const addToCart = async (data) => {\n          "use server";\n        }\n      ',
            },
          ],
        ),
      ],
    },
    {
      code: `
        const requestUsername = (formData) => {
          "use server";
          const username = formData.get('username');
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 33, 5, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                '\n        const requestUsername = async (formData) => {\n          "use server";\n          const username = formData.get(\'username\');\n        }\n      ',
            },
          ],
        ),
      ],
    },
    {
      code: `
        const addToCart = function (data) {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 27, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        const addToCart = async function (data) {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        const requestUsername = function (formData) {
          'use server';
          const username = formData.get('username');
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 33, 5, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        const requestUsername = async function (formData) {\n          'use server';\n          const username = formData.get('username');\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        const addToCart = function (data) {
          "use server";
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 27, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                '\n        const addToCart = async function (data) {\n          "use server";\n        }\n      ',
            },
          ],
        ),
      ],
    },
    {
      code: `
        const requestUsername = function (formData) {
          "use server";
          const username = formData.get('username');
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 33, 5, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                '\n        const requestUsername = async function (formData) {\n          "use server";\n          const username = formData.get(\'username\');\n        }\n      ',
            },
          ],
        ),
      ],
    },
    {
      code: `
        export function addToCart(data) {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 16, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make `addToCart` an `async` function',
              output:
                "\n        export async function addToCart(data) {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        export default function addToCart(data) {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 24, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make `addToCart` an `async` function',
              output:
                "\n        export default async function addToCart(data) {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        export default function (data) {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 24, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        export default async function (data) {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        const obj = {
          action() {
            'use server';
          }
        };
      `,
      errors: [
        expectAsyncServerActionError(
          [3, 17, 5, 12],
          [
            {
              messageId: '',
              desc: 'Make `action` an `async` function',
              output:
                "\n        const obj = {\n          async action() {\n            'use server';\n          }\n        };\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        class Foo {
          action() {
            'use server';
          }
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [3, 17, 5, 12],
          [
            {
              messageId: '',
              desc: 'Make `action` an `async` function',
              output:
                "\n        class Foo {\n          async action() {\n            'use server';\n          }\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        class Foo {
          static action() {
            'use server';
          }
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [3, 24, 5, 12],
          [
            {
              messageId: '',
              desc: 'Make `action` an `async` function',
              output:
                "\n        class Foo {\n          static async action() {\n            'use server';\n          }\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        function outer() {
          function inner() {
            'use server';
          }
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [3, 11, 5, 12],
          [
            {
              messageId: '',
              desc: 'Make `inner` an `async` function',
              output:
                "\n        function outer() {\n          async function inner() {\n            'use server';\n          }\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        const action = function named(data) {
          'use server';
        }
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 24, 4, 10],
          [
            {
              messageId: '',
              desc: 'Make `named` an `async` function',
              output:
                "\n        const action = async function named(data) {\n          'use server';\n        }\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        <form action={() => { 'use server'; }} />
      `,
      errors: [
        expectAsyncServerActionError(
          [2, 23, 2, 46],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        <form action={async () => { 'use server'; }} />\n      ",
            },
          ],
        ),
      ],
    },
    {
      code: `
        <form
          action={function () {
            'use server';
            doSomething();
          }}
        />
      `,
      errors: [
        expectAsyncServerActionError(
          [3, 19, 6, 12],
          [
            {
              messageId: '',
              desc: 'Make this function `async`',
              output:
                "\n        <form\n          action={async function () {\n            'use server';\n            doSomething();\n          }}\n        />\n      ",
            },
          ],
        ),
      ],
    },
  ],
});
