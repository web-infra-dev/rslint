// TestConsistentTupleLabelsExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. Each case points at its branch,
// Dimension 4 row, or representative real-user shape. The direct v74.0.0
// migration lives in consistent_tuple_labels_upstream_test.go.
package consistent_tuple_labels_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	consistent_tuple_labels "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_tuple_labels"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTupleLabelsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_tuple_labels.ConsistentTupleLabelsRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream create() arm 1: tuples with fewer than two
			// elements cannot be inconsistent.
			tsValid(`type Empty = []; type Single = [value: string];`),

			// Locks in upstream create() arm 2a: a tuple with no unlabeled
			// elements is consistent in a declaration container.
			tsValid(`interface API { result: [value: string, error: Error] }`),

			// Locks in upstream create() arm 2b: a tuple with no labels is
			// consistent even when nested inside a generic type argument.
			tsValid(`type Result = Promise<[string, Error]>;`),

			// Locks in upstream isLabeledElement() arms 1 and 2: tsgo represents
			// both an ordinary labeled element and a labeled rest element as
			// NamedTupleMember nodes.
			tsValid(`type Params = [head: string, ...tail: number[]];`),

			// ---- Dimension 4: same-kind nesting remains independent ----
			tsValid(`type Nested = [outer: [inner: string, count: number], flag: boolean];`),

			// N/A: runtime receiver wrappers, optional chains, member-key forms,
			// class/function variants, and destructuring shapes are not tuple
			// element forms inspected by this type-syntax-only rule.
			// N/A: the rule has no autofix, suggestions, or options.
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream create() final arm: every unlabeled element in a
			// mixed tuple receives its own diagnostic.
			tsInvalid(
				`type Mixed = [label: string, number, boolean];`,
				expectedError(`type Mixed = [label: string, number, boolean];`, `number`, 0),
				expectedError(`type Mixed = [label: string, number, boolean];`, `boolean`, 0),
			),

			// ---- Dimension 4: parenthesized and union element types stay
			// unlabeled and are reported across their complete type ranges. ----
			tsInvalid(
				`type Parenthesized = [label: string, (number)];`,
				expectedError(`type Parenthesized = [label: string, (number)];`, `(number)`, 0),
			),
			tsInvalid(
				`type UnionElement = [label: string, number | undefined];`,
				expectedError(`type UnionElement = [label: string, number | undefined];`, `number | undefined`, 0),
			),

			// ---- Dimension 4: nested mixed tuples report at both tuple levels
			// without bleeding sibling state. ----
			tsInvalid(
				`type Nested = [outer: string, [inner: number, boolean]];`,
				expectedError(`type Nested = [outer: string, [inner: number, boolean]];`, `[inner: number, boolean]`, 0),
				expectedError(`type Nested = [outer: string, [inner: number, boolean]];`, `boolean`, 0),
			),

			// ---- Real-user: upstream #3382 partially labeled event-listener argument tuple ----
			tsInvalid(
				`type Listener = (...args: [event: string, number]) => void;`,
				expectedError(`type Listener = (...args: [event: string, number]) => void;`, `number`, 0),
			),

			// ---- Real-user: upstream #3382 partially labeled tuple in a public declaration ----
			tsInvalid(
				`declare function load(): Promise<[
	value: string,
	Error,
]>;`,
				expectedError(`declare function load(): Promise<[
	value: string,
	Error,
]>;`, `Error`, 0),
			),
		},
	)
}
