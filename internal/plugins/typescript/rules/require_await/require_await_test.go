package require_await

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRequireAwaitRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireAwaitRule, []rule_tester.ValidTestCase{
		{Code: `
function numberOne(): number {
  return 1;
}
    `},
		{Code: `
const numberOne = function (): number {
  return 1;
};
    `},
		{Code: `
      const numberOne = (): number => 1;
    `},
		{Code: `
const numberOne = (): number => {
  return 1;
};
    `},
		{Code: `
function delay() {
  return Promise.resolve();
}
    `},
		{Code: `
const delay = () => {
  return Promise.resolve();
};
    `},
		{Code: "const delay = () => Promise.resolve();"},
		{Code: `
async function numberOne(): Promise<number> {
  return await 1;
}
    `},
		{Code: `
const numberOne = async function (): Promise<number> {
  return await 1;
};
    `},
		{Code: "const numberOne = async (): Promise<number> => await 1;"},
		{Code: `
const numberOne = async (): Promise<number> => {
  return await 1;
};
    `},
		{Code: `
async function numberOne(): Promise<number> {
  return Promise.resolve(1);
}
    `},
		{Code: `
const numberOne = async function (): Promise<number> {
  return Promise.resolve(1);
};
    `},
		{Code: "const numberOne = async (): Promise<number> => Promise.resolve(1);"},
		{Code: `
const numberOne = async (): Promise<number> => {
  return Promise.resolve(1);
};
    `},
		{Code: `
async function numberOne(): Promise<number> {
  return getAsyncNumber(1);
}
async function getAsyncNumber(x: number): Promise<number> {
  return Promise.resolve(x);
}
    `},
		{Code: `
const numberOne = async function (): Promise<number> {
  return getAsyncNumber(1);
};
const getAsyncNumber = async function (x: number): Promise<number> {
  return Promise.resolve(x);
};
    `},
		{Code: `
const numberOne = async (): Promise<number> => getAsyncNumber(1);
const getAsyncNumber = async function (x: number): Promise<number> {
  return Promise.resolve(x);
};
    `},
		{Code: `
const numberOne = async (): Promise<number> => {
  return getAsyncNumber(1);
};
const getAsyncNumber = async function (x: number): Promise<number> {
  return Promise.resolve(x);
};
    `},
		{Code: `
async function testFunction(): Promise<void> {
  await Promise.all(
    [1, 2, 3].map(
      // this should not trigger an error on the parent function
      async value => Promise.resolve(value),
    ),
  );
}
    `},
		{Code: `
function* test6() {
  yield* syncGenerator();
}
    `},
		{Code: `
function* syncGenerator() {
  yield 1;
}
    `},
		{Code: `
async function* asyncGenerator() {
  await Promise.resolve();
  yield 1;
}
async function* test1() {
  yield* asyncGenerator();
}
    `},
		{Code: `
async function* asyncGenerator() {
  await Promise.resolve();
  yield 1;
}
async function* test1() {
  yield* asyncGenerator();
  yield* 2;
}
    `},
		{Code: `
async function* test(source: AsyncIterable<any>) {
  yield* source;
}
    `},
		{Code: `
async function* test(source: Iterable<any> & AsyncIterable<any>) {
  yield* source;
}
    `},
		{Code: `
async function* test(source: Iterable<any> | AsyncIterable<any>) {
  yield* source;
}
    `},
		{Code: `
type MyType = {
  [Symbol.iterator](): Iterator<any>;
  [Symbol.asyncIterator](): AsyncIterator<any>;
};
async function* test(source: MyType) {
  yield* source;
}
    `},
		{Code: `
type MyType = {
  [Symbol.asyncIterator]: () => AsyncIterator<any>;
};
async function* test(source: MyType) {
  yield* source;
}
    `},
		{Code: `
type MyFunctionType = () => AsyncIterator<any>;
type MyType = {
  [Symbol.asyncIterator]: MyFunctionType;
};
async function* test(source: MyType) {
  yield* source;
}
    `},
		{Code: `
async function* foo(): Promise<string> {
  return new Promise(res => res(` + "`" + `hello` + "`" + `));
}
    `},
		{Code: `
      async function* f() {
        let x!: Omit<
          {
            [Symbol.asyncIterator](): AsyncIterator<any>;
          },
          'z'
        >;
        yield* x;
      }
    `},
		{Code: `
      const fn = async () => {
        await using foo = new Bar();
      };
    `},
		{Code: `
      async function* test1() {
        yield Promise.resolve(1);
      }
    `},
		{Code: `
      function asyncFunction() {
        return Promise.resolve(1);
      }
      async function* test1() {
        yield asyncFunction();
      }
    `},
		{Code: `
      declare const asyncFunction: () => Promise<void>;
      async function* test1() {
        yield asyncFunction();
      }
    `},
		{Code: `
      async function* test1() {
        yield new Promise(() => {});
      }
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
async function numberOne(): Promise<number> {
  return 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function numberOne(): number {
					//   return 1;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
const numberOne = async function (): Promise<number> {
  return 1;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// const numberOne = function (): number {
					//   return 1;
					// };
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: "const numberOne = async (): Promise<number> => 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//   {
					//     MessageId: "removeAsync",
					//     Output: "const numberOne = (): number => 1;",
					//   },
					// },
				},
			},
		},
		{
			Code: `
async function values(): Promise<Array<number>> {
  return [1];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function values(): Array<number> {
					//   return [1];
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
        async function foo() {
          function nested() {
            await doSomething();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function foo() {
					//     function nested() {
					//       await doSomething();
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
async function* foo(): void {
  doSomething();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* foo(): void {
					//   doSomething();
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* foo() {
  yield 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* foo() {
					//   yield 1;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
const foo = async function* () {
  console.log('bar');
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// const foo = function* () {
					//   console.log('bar');
					// };
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* asyncGenerator() {
  yield 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* asyncGenerator() {
					//   yield 1;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* asyncGenerator(source: Iterable<any>) {
  yield* source;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* asyncGenerator(source: Iterable<any>) {
					//   yield* source;
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
function isAsyncIterable(value: unknown): value is AsyncIterable<any> {
  return true;
}
async function* asyncGenerator(source: Iterable<any> | AsyncIterable<any>) {
  if (!isAsyncIterable(source)) {
    yield* source;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function isAsyncIterable(value: unknown): value is AsyncIterable<any> {
					//   return true;
					// }
					// function* asyncGenerator(source: Iterable<any> | AsyncIterable<any>) {
					//   if (!isAsyncIterable(source)) {
					//     yield* source;
					//   }
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
function* syncGenerator() {
  yield 1;
}
async function* asyncGenerator() {
  yield* syncGenerator();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* syncGenerator() {
					//   yield 1;
					// }
					// function* asyncGenerator() {
					//   yield* syncGenerator();
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
async function* asyncGenerator() {
  yield* anotherAsyncGenerator(); // Unknown function.
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//             Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//               {
					//                 MessageId: "removeAsync",
					//                 Output: `
					// function* asyncGenerator() {
					//   yield* anotherAsyncGenerator(); // Unknown function.
					// }
					//       `,
					//               },
					//             },
				},
			},
		},
		{
			Code: `
        const fn = async () => {
          using foo = new Bar();
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   const fn = () => {
					//     using foo = new Bar();
					//   };
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        // intentional TS error
        async function* foo(): Promise<number> {
          yield 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   // intentional TS error
					//   function* foo(): Promise<number> {
					//     yield 1;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function* foo(): AsyncGenerator {
          yield 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function* foo(): Generator {
					//     yield 1;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function* foo(): AsyncGenerator<number> {
          yield 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function* foo(): Generator<number> {
					//     yield 1;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
	})
}

func TestRequireAwaitRuleEslintBase(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireAwaitRule, []rule_tester.ValidTestCase{
		{Code: `
async function foo() {
  await doSomething();
}
    `},
		{Code: `
(async function () {
  await doSomething();
});
    `},
		{Code: `
async () => {
  await doSomething();
};
    `},
		{Code: "async () => await doSomething();"},
		{Code: `
({
  async foo() {
    await doSomething();
  },
});
    `},
		{Code: `
class A {
  async foo() {
    await doSomething();
  }
}
    `},
		{Code: `
(class {
  async foo() {
    await doSomething();
  }
});
    `},
		{Code: `
async function foo() {
  await (async () => {
    await doSomething();
  });
}
    `},
		{Code: "async function foo() {}"},
		{Code: "async () => {};"},
		{Code: `
function foo() {
  doSomething();
}
    `},
		{Code: `
async function foo() {
  for await (x of xs);
}
    `},
		{
			Code: "await foo();",
		},
		{
			Code: `
for await (let num of asyncIterable) {
  console.log(num);
}
      `,
		},
		{
			Code: `
        async function* run() {
          await new Promise(resolve => setTimeout(resolve, 100));
          yield 'Hello';
          console.log('World');
        }
      `,
		},
		{
			Code: "async function* run() {}",
		},
		{
			Code: "const foo = async function* () {};",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        async function foo() {
          doSomething();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function foo() {
					//     doSomething();
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        (async function () {
          doSomething();
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   (function () {
					//     doSomething();
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async () => {
          doSomething();
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   () => {
					//     doSomething();
					//   };
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: "async () => doSomething();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					// Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//   {
					//     MessageId: "removeAsync",
					//     Output: "() => doSomething();",
					//   },
					// },
				},
			},
		},
		{
			Code: `
        ({
          async foo() {
            doSomething();
          },
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   ({
					//     foo() {
					//       doSomething();
					//     },
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          async foo() {
            doSomething();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     foo() {
					//       doSomething();
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          public async foo() {
            doSomething();
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     public foo() {
					//       doSomething();
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        (class {
          async foo() {
            doSomething();
          }
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   (class {
					//     foo() {
					//       doSomething();
					//     }
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        (class {
          async ''() {
            doSomething();
          }
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   (class {
					//     ''() {
					//       doSomething();
					//     }
					//   });
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function foo() {
          async () => {
            await doSomething();
          };
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   function foo() {
					//     async () => {
					//       await doSomething();
					//     };
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async function foo() {
          await (async () => {
            doSomething();
          });
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   async function foo() {
					//     await (() => {
					//       doSomething();
					//     });
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        const obj = {
          async: async function foo() {
            bar();
          },
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   const obj = {
					//     async: function foo() {
					//       bar();
					//     },
					//   };
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        async    /* test */ function foo() {
          doSomething();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   /* test */ function foo() {
					//     doSomething();
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          a = 0
          async [b]() {
            return 0;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     a = 0
					//     ;[b]() {
					//       return 0;
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        foo
        async () => {
          return 0;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   foo
					//   ;() => {
					//     return 0;
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
		{
			Code: `
        class A {
          foo() {}
          async [bar]() {
            baz;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingAwait",
					//       Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					//         {
					//           MessageId: "removeAsync",
					//           Output: `
					//   class A {
					//     foo() {}
					//     [bar]() {
					//       baz;
					//     }
					//   }
					// `,
					//         },
					//       },
				},
			},
		},
	})
}

func TestRequireAwaitDeferredTypeChecks(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		[]rule_tester.ValidTestCase{
			{Code: `
declare function consume(value: number): number;
async function returnsAwait() {
  return consume(await Promise.resolve(1));
}
      `},
			{Code: `
declare function consume(value: number): number;
const arrow = async () => consume(await Promise.resolve(1));
      `},
			{Code: `
declare function consume(value: number): number;
async function* yieldsAwait() {
  yield consume(await Promise.resolve(1));
}
      `},
			{Code: `
async function* awaitsBeforeYield() {
  await Promise.resolve();
  yield Promise.resolve(1);
}
      `},
			{Code: `
interface StructuralThenable {
  then(onfulfilled: (value: number) => unknown): unknown;
}
declare const value: StructuralThenable;
async function returnsThenable() {
  return value;
}
      `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
interface FakeThenable {
  then(value: number): unknown;
}
declare const value: FakeThenable;
async function returnsFakeThenable() {
  return value;
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
interface FakeThenable {
  then(value: number): unknown;
}
declare const value: FakeThenable;
const returnsFakeThenable = async () => value;
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
interface FakeThenable {
  then(value: number): unknown;
}
declare const value: FakeThenable;
async function* yieldsFakeThenable() {
  yield value;
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
async function outer() {
  function* syncGenerator() {
    yield Promise.resolve(1);
  }
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
async function outer() {
  return async () => await Promise.resolve(1);
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
async function* outer() {
  yield async () => await Promise.resolve(1);
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
		},
	)
}

func TestRequireAwaitDeepScopeStack(t *testing.T) {
	const depth = 40
	var source strings.Builder
	for i := range depth {
		source.WriteString("async function f")
		source.WriteString(strconv.Itoa(i))
		source.WriteString("() {\n")
	}
	source.WriteString("await Promise.resolve();\n")
	for range depth {
		source.WriteString("}\n")
	}

	errors := make([]rule_tester.InvalidTestCaseError, depth-1)
	for i := range errors {
		errors[i].MessageId = "missingAwait"
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		[]rule_tester.InvalidTestCase{{
			Code:   source.String(),
			Errors: errors,
		}},
	)
}

func TestRequireAwaitThenableFastPathParity(t *testing.T) {
	const source = `
interface GoodThenable {
  then(onfulfilled: (value: number) => unknown): unknown;
}
interface RestThenable {
  then(...callbacks: Array<(value: number) => unknown>): unknown;
}
interface UnionCallbackThenable {
  then(onfulfilled: ((value: number) => unknown) | { tag: string }): unknown;
}
interface BadThenable {
  then(value: number): unknown;
}
interface NoArgThenable {
  then(): unknown;
}
declare const good: GoodThenable;
declare const rest: RestThenable;
declare const unionCallback: UnionCallbackThenable;
declare const bad: BadThenable;
declare const noArg: NoArgThenable;

const caseNumber = 1;
const casePromise = Promise.resolve(1);
const caseGood = good;
const caseRest = rest;
const caseUnionCallback = unionCallback;
const caseBad = bad;
const caseNoArg = noArg;
const caseGoodUnion = null as number | GoodThenable;
const caseBadUnion = null as number | BadThenable;
const caseIntersection = null as GoodThenable & { tag: string };
`
	expected := map[string]bool{
		"caseNumber":        false,
		"casePromise":       true,
		"caseGood":          true,
		"caseRest":          true,
		"caseUnionCallback": true,
		"caseBad":           false,
		"caseNoArg":         false,
		"caseGoodUnion":     true,
		"caseBadUnion":      false,
		"caseIntersection":  true,
	}

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "require-await-thenable-parity.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	seen := make(map[string]bool, len(expected))
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsExpressionNode(node) {
			typ := typeChecker.GetTypeAtLocation(node)
			got := isThenableType(typeChecker, node, typ)
			shared := utils.IsThenableType(typeChecker, node, typ)
			if got != shared {
				t.Errorf(
					"expression kind %v at [%d,%d): fast path = %t, shared helper = %t",
					node.Kind,
					node.Pos(),
					node.End(),
					got,
					shared,
				)
			}
		}
		if node.Kind == ast.KindVariableDeclaration {
			name := node.Name()
			initializer := node.Initializer()
			if name != nil && initializer != nil {
				if want, ok := expected[name.Text()]; ok {
					typ := typeChecker.GetTypeAtLocation(initializer)
					got := isThenableType(typeChecker, initializer, typ)
					shared := utils.IsThenableType(typeChecker, initializer, typ)
					if got != shared {
						t.Errorf("%s: fast path = %t, shared helper = %t", name.Text(), got, shared)
					}
					if got != want {
						t.Errorf("%s: isThenableType = %t, want %t", name.Text(), got, want)
					}
					seen[name.Text()] = true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)

	for name := range expected {
		if !seen[name] {
			t.Errorf("test case %s was not visited", name)
		}
	}
}

func TestRequireAwaitRecoveryAST(t *testing.T) {
	for index, source := range []string{
		`async function value() {`,
		`const value = async () => {`,
		`async function* value() {`,
		`async function outer() { function inner() {`,
		`class Value { async method() {`,
		`async function outer() { const inner = async () => {`,
	} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			fileName := "/require-await-recovery-" + strconv.Itoa(index) + ".ts"
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: fileName,
				Path:     tspath.Path(fileName),
			}, source, core.ScriptKindTS)
			ctx := rule.RuleContext{
				SourceFile:     sourceFile,
				DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
			}.WithReporter(
				RequireAwaitRule.Name,
				rule.SeverityError,
				func(diagnostic rule.RuleDiagnostic) {
					if diagnostic.Range.Pos() < 0 ||
						diagnostic.Range.End() < diagnostic.Range.Pos() ||
						diagnostic.Range.End() > len(source) {
						t.Errorf(
							"out-of-bounds diagnostic range [%d,%d) for source length %d",
							diagnostic.Range.Pos(),
							diagnostic.Range.End(),
							len(source),
						)
					}
				},
			)
			listeners := RequireAwaitRule.Run(ctx, nil)

			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if listener := listeners[node.Kind]; listener != nil {
					listener(node)
				}
				node.ForEachChild(visit)
				if listener := listeners[rule.ListenerOnExit(node.Kind)]; listener != nil {
					listener(node)
				}
				return false
			}
			sourceFile.AsNode().ForEachChild(visit)
		})
	}
}
