package async_server_action

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Source: https://github.com/jsx-eslint/eslint-plugin-react/blob/5d4bf128fb506f604f1bb61bcf76fa8caeea6a1a/tests/lib/rules/async-server-action.js
// All 52 valid and 22 invalid upstream cases, with ranges checked against ESLint.
func TestAsyncServerActionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AsyncServerActionRule,
		[]rule_tester.ValidTestCase{{Code: `
        async function addToCart(data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        async function requestUsername(formData) {
          'use server';
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        async function addToCart(data) {
          "use server";
        }
      `, Tsx: true},
			{Code: `
        async function requestUsername(formData) {
          "use server";
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        function addToCart(data) {
          console.log("test");
          'use server';
        }
      `, Tsx: true},
			{Code: `
        function requestUsername(formData) {
          const username = formData.get('username');
          'use server';
        }
      `, Tsx: true},
			{Code: `
        function addToCart(data) {
          console.log("use server");
        }
      `, Tsx: true},
			{Code: `
        function requestUsername(formData) {
          console.log("use server");
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        const addToCart = async (data) => {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = async (formData) => {
          'use server';
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        const addToCart = async (data) => {
          "use server";
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = async (formData) => {
          "use server";
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        const addToCart = (data) => {
          console.log("test");
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = (formData) => {
          const username = formData.get('username');
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const addToCart = (data) => {
          console.log("use server");
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = (formData) => {
          console.log("use server");
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        const addToCart = async function (data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = async function (formData) {
          'use server';
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        const addToCart = async function (data) {
          "use server";
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = async function (formData) {
          "use server";
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: `
        const addToCart = function (data) {
          console.log("test");
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = function (formData) {
          const username = formData.get('username');
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const addToCart = function (data) {
          console.log("use server");
        }
      `, Tsx: true},
			{Code: `
        const requestUsername = function (formData) {
          console.log("use server");
          const username = formData.get('username');
        }
      `, Tsx: true},
			{Code: "\n        async function addToCart(data) {\n          `use server`;\n        }\n      ", Tsx: true},
			{Code: "\n        function addToCart(data) {\n          `use server`;\n        }\n      ", Tsx: true},
			{Code: "\n        const addToCart = async (data) => {\n          `use server`;\n        }\n      ", Tsx: true},
			{Code: "\n        const addToCart = (data) => {\n          `use server`;\n        }\n      ", Tsx: true},
			{Code: "\n        const addToCart = async function (data) {\n          `use server`;\n        }\n      ", Tsx: true},
			{Code: "\n        const addToCart = function (data) {\n          `use server`;\n        }\n      ", Tsx: true},
			{Code: `
        const addToCart = async function* (data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const addToCart = async function* (data) {
          "use server";
        }
      `, Tsx: true},
			{Code: `
        const addToCart = function* (data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const addToCart = function* (data) {
          "use server";
        }
      `, Tsx: true},
			{Code: `
        function* addToCart(data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        async function* addToCart(data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        export async function addToCart(data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        export default async function addToCart(data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        export default async function (data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        const obj = {
          async action() {
            'use server';
          }
        };
      `, Tsx: true},
			{Code: `
        const obj = {
          async action() {
            'use server';
            const x = 1;
          }
        };
      `, Tsx: true},
			{Code: `
        class Foo {
          async action() {
            'use server';
          }
        }
      `, Tsx: true},
			{Code: `
        class Foo {
          static async action() {
            'use server';
          }
        }
      `, Tsx: true},
			{Code: `
        function outer() {
          async function inner() {
            'use server';
          }
        }
      `, Tsx: true},
			{Code: `
        const action = async function named(data) {
          'use server';
        }
      `, Tsx: true},
			{Code: `
        function addToCart(data) {
          'use strict';
          console.log('use server');
        }
      `, Tsx: true},
			{Code: `
        function empty() {}
      `, Tsx: true},
			{Code: `
        const fn = () => 'use server';
      `, Tsx: true},
			{Code: `
        <form action={async () => { 'use server'; }} />
      `, Tsx: true},
			{Code: `
        <button onClick={async () => { 'use server'; doSomething(); }} />
      `, Tsx: true},
			{Code: `
        async function action() {
          'use strict';
          'use server';
        }
      `, Tsx: true},
			{Code: `
        function action() {
          'use strict';
          'use server';
        }
      `, Tsx: true}},
		[]rule_tester.InvalidTestCase{{Code: `
        function addToCart(data) {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
			Line: 2, Column: 9, EndLine: 4, EndColumn: 10,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				Output: `
        async function addToCart(data) {
          'use server';
        }
      `,
			}},
		}}},
			{Code: `
        function requestUsername(formData) {
          'use server';
          const username = formData.get('username');
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 9, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        async function requestUsername(formData) {
          'use server';
          const username = formData.get('username');
        }
      `,
				}},
			}}},
			{Code: `
        function addToCart(data) {
          "use server";
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 9, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        async function addToCart(data) {
          "use server";
        }
      `,
				}},
			}}},
			{Code: `
        function requestUsername(formData) {
          "use server";
          const username = formData.get('username');
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 9, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        async function requestUsername(formData) {
          "use server";
          const username = formData.get('username');
        }
      `,
				}},
			}}},
			{Code: `
        const addToCart = (data) => {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 27, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const addToCart = async (data) => {
          'use server';
        }
      `,
				}},
			}}},
			{Code: `
        const requestUsername = (formData) => {
          'use server';
          const username = formData.get('username');
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 33, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const requestUsername = async (formData) => {
          'use server';
          const username = formData.get('username');
        }
      `,
				}},
			}}},
			{Code: `
        const addToCart = (data) => {
          "use server";
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 27, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const addToCart = async (data) => {
          "use server";
        }
      `,
				}},
			}}},
			{Code: `
        const requestUsername = (formData) => {
          "use server";
          const username = formData.get('username');
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 33, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const requestUsername = async (formData) => {
          "use server";
          const username = formData.get('username');
        }
      `,
				}},
			}}},
			{Code: `
        const addToCart = function (data) {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 27, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const addToCart = async function (data) {
          'use server';
        }
      `,
				}},
			}}},
			{Code: `
        const requestUsername = function (formData) {
          'use server';
          const username = formData.get('username');
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 33, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const requestUsername = async function (formData) {
          'use server';
          const username = formData.get('username');
        }
      `,
				}},
			}}},
			{Code: `
        const addToCart = function (data) {
          "use server";
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 27, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const addToCart = async function (data) {
          "use server";
        }
      `,
				}},
			}}},
			{Code: `
        const requestUsername = function (formData) {
          "use server";
          const username = formData.get('username');
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 33, EndLine: 5, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const requestUsername = async function (formData) {
          "use server";
          const username = formData.get('username');
        }
      `,
				}},
			}}},
			{Code: `
        export function addToCart(data) {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 16, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        export async function addToCart(data) {
          'use server';
        }
      `,
				}},
			}}},
			{Code: `
        export default function addToCart(data) {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 24, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        export default async function addToCart(data) {
          'use server';
        }
      `,
				}},
			}}},
			{Code: `
        export default function (data) {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 24, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        export default async function (data) {
          'use server';
        }
      `,
				}},
			}}},
			{Code: `
        const obj = {
          action() {
            'use server';
          }
        };
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 3, Column: 17, EndLine: 5, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const obj = {
          async action() {
            'use server';
          }
        };
      `,
				}},
			}}},
			{Code: `
        class Foo {
          action() {
            'use server';
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 3, Column: 17, EndLine: 5, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        class Foo {
          async action() {
            'use server';
          }
        }
      `,
				}},
			}}},
			{Code: `
        class Foo {
          static action() {
            'use server';
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 3, Column: 24, EndLine: 5, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        class Foo {
          static async action() {
            'use server';
          }
        }
      `,
				}},
			}}},
			{Code: `
        function outer() {
          function inner() {
            'use server';
          }
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 3, Column: 11, EndLine: 5, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        function outer() {
          async function inner() {
            'use server';
          }
        }
      `,
				}},
			}}},
			{Code: `
        const action = function named(data) {
          'use server';
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 24, EndLine: 4, EndColumn: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        const action = async function named(data) {
          'use server';
        }
      `,
				}},
			}}},
			{Code: `
        <form action={() => { 'use server'; }} />
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 2, Column: 23, EndLine: 2, EndColumn: 46,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        <form action={async () => { 'use server'; }} />
      `,
				}},
			}}},
			{Code: `
        <form
          action={function () {
            'use server';
            doSomething();
          }}
        />
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 3, Column: 19, EndLine: 6, EndColumn: 12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `
        <form
          action={async function () {
            'use server';
            doSomething();
          }}
        />
      `,
				}},
			}}}})
}

// The four pinned documentation examples replace placeholder ellipses with calls.
func TestAsyncServerActionDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &AsyncServerActionRule,
		[]rule_tester.ValidTestCase{{Code: `<form
  action={async () => {
    'use server';
    doSomething();
  }}
>
  doSomething();
</form>
`, Tsx: true},
			{Code: `async function action() {
  'use server';
  doSomething();
}
`, Tsx: true}},
		[]rule_tester.InvalidTestCase{{Code: `<form
  action={() => {
    'use server';
    doSomething();
  }}
>
  doSomething();
</form>
`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
			Line: 2, Column: 11, EndLine: 5, EndColumn: 4,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				Output: `<form
  action={async () => {
    'use server';
    doSomething();
  }}
>
  doSomething();
</form>
`,
			}},
		}}},
			{Code: `function action() {
  'use server';
  doSomething();
}
`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: `asyncServerAction`, Message: `Server Actions must be async`,
				Line: 1, Column: 1, EndLine: 4, EndColumn: 2,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					Output: `async function action() {
  'use server';
  doSomething();
}
`,
				}},
			}}}})
}
