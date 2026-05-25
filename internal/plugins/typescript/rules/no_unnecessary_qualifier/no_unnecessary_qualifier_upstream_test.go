// TestNoUnnecessaryQualifierUpstream migrates the full valid/invalid suite
// from upstream
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-unnecessary-qualifier.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// no_unnecessary_qualifier_extras_test.go.
package no_unnecessary_qualifier

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryQualifierUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryQualifierRule, []rule_tester.ValidTestCase{
		{Code: `
namespace X {
  export type T = number;
}

namespace Y {
  export const x: X.T = 3;
}
    `},
		{Code: `
namespace A {}
namespace A.B {
  export type Z = 1;
}
    `},
		{Code: `
enum A {
  X,
  Y,
}

enum B {
  Z = A.X,
}
    `},
		{Code: `
namespace X {
  export type T = number;
  namespace Y {
    type T = string;
    const x: X.T = 0;
  }
}
    `},
		{Code: `const x: A.B = 3;`},
		{Code: `
namespace X {
  const z = X.y;
}
    `},
		{Code: `
enum Foo {
  One,
}

namespace Foo {
  export function bar() {
    return Foo.One;
  }
}
    `},
		{Code: `
namespace Foo {
  export enum Foo {
    One,
  }
}

namespace Foo {
  export function bar() {
    return Foo.One;
  }
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
namespace A {
  export type B = number;
  const x: A.B = 3;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 12},
			},
			Output: []string{`
namespace A {
  export type B = number;
  const x: B = 3;
}
      `},
		},
		{
			Code: `
namespace A {
  export const x = 3;
  export const y = A.x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 20},
			},
			Output: []string{`
namespace A {
  export const x = 3;
  export const y = x;
}
      `},
		},
		{
			Code: `
namespace A {
  export type T = number;
  export namespace B {
    const x: A.T = 3;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 14},
			},
			Output: []string{`
namespace A {
  export type T = number;
  export namespace B {
    const x: T = 3;
  }
}
      `},
		},
		{
			Code: `
namespace A {
  export namespace B {
    export type T = number;
    const x: A.B.T = 3;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 14},
			},
			Output: []string{`
namespace A {
  export namespace B {
    export type T = number;
    const x: T = 3;
  }
}
      `},
		},
		{
			Code: `
namespace A {
  export namespace B.C {
    export type D = number;
    const x: A.B.C.D = 3;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 14},
			},
			Output: []string{`
namespace A {
  export namespace B.C {
    export type D = number;
    const x: D = 3;
  }
}
      `},
		},
		{
			Code: `
namespace A {
  export namespace B {
    export const x = 3;
    const y = A.B.x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 15},
			},
			Output: []string{`
namespace A {
  export namespace B {
    export const x = 3;
    const y = x;
  }
}
      `},
		},
		{
			Code: `
enum A {
  B,
  C = A.B,
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 7},
			},
			Output: []string{`
enum A {
  B,
  C = B,
}
      `},
		},
		{
			Code: `
namespace Foo {
  export enum A {
    B,
    C = Foo.A.B,
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 9},
			},
			Output: []string{`
namespace Foo {
  export enum A {
    B,
    C = B,
  }
}
      `},
		},
		{
			Code: `
import * as Foo from './foo';
declare module './foo' {
  const x: Foo.T = 3;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 12},
			},
			Output: []string{`
import * as Foo from './foo';
declare module './foo' {
  const x: T = 3;
}
      `},
		},
	})
}
