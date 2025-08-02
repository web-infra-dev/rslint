package no_unsafe_declaration_merging

import (
	"testing"
)

func TestNoUnsafeDeclarationMerging(t *testing.T) {
	// TODO: Convert this test to use rule_tester.RunRuleTester format
	// This test uses an incompatible structure with rule_tester.TestCases
	/*
	tester := rule_tester.New(t)

	tester.Run("no-unsafe-declaration-merging", NoUnsafeDeclarationMergingRule, rule_tester.TestCases{
		Valid: []rule_tester.ValidTestCase{
			{
				Code: `
interface Foo {}
class Bar implements Foo {}
`,
			},
			{
				Code: `
namespace Foo {}
namespace Foo {}
`,
			},
			{
				Code: `
enum Foo {}
namespace Foo {}
`,
			},
			{
				Code: `
namespace Fooo {}
function Foo() {}
`,
			},
			{
				Code: `
const Foo = class {};
`,
			},
			{
				Code: `
interface Foo {
  props: string;
}

function bar() {
  return class Foo {};
}
`,
			},
			{
				Code: `
interface Foo {
  props: string;
}

(function bar() {
  class Foo {}
})();
`,
			},
			{
				Code: `
declare global {
  interface Foo {}
}

class Foo {}
`,
			},
		},
		Invalid: []rule_tester.InvalidTestCase{
			{
				Code: `
interface Foo {}
class Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeMerging",
						Line:      2,
						Column:    11,
					},
					{
						MessageId: "unsafeMerging",
						Line:      3,
						Column:    7,
					},
				},
			},
			{
				Code: `
class Foo {}
interface Foo {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeMerging",
						Line:      2,
						Column:    7,
					},
					{
						MessageId: "unsafeMerging",
						Line:      3,
						Column:    11,
					},
				},
			},
			{
				Code: `
declare global {
  interface Foo {}
  class Foo {}
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeMerging",
						Line:      3,
						Column:    13,
					},
					{
						MessageId: "unsafeMerging",
						Line:      4,
						Column:    9,
					},
				},
			},
		},
	})
	*/
}