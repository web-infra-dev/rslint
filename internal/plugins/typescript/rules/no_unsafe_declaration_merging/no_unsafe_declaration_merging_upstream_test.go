// TestNoUnsafeDeclarationMergingUpstream migrates the full valid/invalid
// suite from upstream
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-unsafe-declaration-merging.test.ts
// 1:1. Position assertions cover line/column for every invalid case. rslint
// specific lock-in cases live in
// no_unsafe_declaration_merging_extras_test.go.
package no_unsafe_declaration_merging

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeDeclarationMergingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeDeclarationMergingRule, []rule_tester.ValidTestCase{
		// ---- valid ----
		{Code: `
interface Foo {}
class Bar implements Foo {}
`},
		{Code: `
namespace Foo {}
namespace Foo {}
`},
		{Code: `
enum Foo {}
namespace Foo {}
`},
		{Code: `
namespace Fooo {}
function Foo() {}
`},
		{Code: `
const Foo = class {};
`},
		{Code: `
interface Foo {
  props: string;
}

function bar() {
  return class Foo {};
}
`},
		{Code: `
interface Foo {
  props: string;
}

(function bar() {
  class Foo {}
})();
`},
		{Code: `
declare global {
  interface Foo {}
}

class Foo {}
`},
	}, []rule_tester.InvalidTestCase{
		// ---- invalid ----
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
	})
}
