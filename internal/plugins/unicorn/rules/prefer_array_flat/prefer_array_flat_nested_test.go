// cspell:ignore tems
package prefer_array_flat_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferArrayFlatNestedAndScope covers parenthesized callees, TypeScript
// parameter shapes, lexical symbol resolution, configured deep paths, and
// overlapping fixes at more than two nesting levels.
func TestPreferArrayFlatNestedAndScope(t *testing.T) {
	var suite upstreamSuite

	suite.addFixed(
		`(array.flatMap)(x => x)`,
		`(array.flatMap)(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`((array.flatMap(x => x)))`,
		`array.flatMap(x => x)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`array.flatMap((value?: unknown) => value)`,
		`array.flatMap((value?: unknown) => value)`,
		`array.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addValid(nil,
		`array.flatMap((value = []) => value)`,
		`array.flatMap((...values) => values)`,
		`array.flatMap(([value]) => value)`,
	)

	// Receiver classification follows the exact symbol in the current scope
	// and only unwraps parentheses around a const initializer.
	suite.addValid(nil,
		`const collection = ({}); collection.flatMap(value => value);`,
		`const Items = [] as unknown[]; Items.flatMap(value => value);`,
	)
	suite.addFixed(
		`const collection = {} as unknown; collection.flatMap(value => value);`,
		`collection.flatMap(value => value)`,
		`collection.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`let collection = {}; collection.flatMap(value => value);`,
		`collection.flatMap(value => value)`,
		`collection.flat()`,
		`Array#flatMap()`,
		nil,
	)
	// U+2160 is uppercase-like but belongs to Unicode category Nl, not Lu.
	// Upstream's /^\p{Lu}/u therefore does not suppress this receiver.
	suite.addFixed(
		`const Ⅰtems = getValues(); Ⅰtems.flatMap(value => value);`,
		`Ⅰtems.flatMap(value => value)`,
		`Ⅰtems.flat()`,
		`Array#flatMap()`,
		nil,
	)
	shadowedCode := `{ const Items = {}; Items.flatMap(value => value); }
const Items = [];
Items.flatMap(value => value);`
	suite.addFixedOutput(
		shadowedCode,
		`{ const Items = {}; Items.flatMap(value => value); }
const Items = [];
Items.flat();`,
		nil,
		expectedDiagnostic{
			target:      `Items.flatMap(value => value)`,
			description: `Array#flatMap()`,
			occurrence:  1,
		},
	)
	suite.addFixed(
		`Items.flatMap(value => value); const Items = [];`,
		`Items.flatMap(value => value)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)
	// ESLint's variable definition points each destructured binding at the
	// enclosing declarator, so classification uses that declarator's complete
	// initializer rather than the extracted property's runtime value.
	suite.addValid(nil,
		`const {items} = {}; items.flatMap(value => value);`,
	)
	suite.addFixed(
		`const [items] = []; items.flatMap(value => value);`,
		`items.flatMap(value => value)`,
		`items.flat()`,
		`Array#flatMap()`,
		nil,
	)
	suite.addFixed(
		`const {Items} = []; Items.flatMap(value => value);`,
		`Items.flatMap(value => value)`,
		`Items.flat()`,
		`Array#flatMap()`,
		nil,
	)

	pathOptions := map[string]interface{}{
		"functions": []interface{}{"utils.deep.flat", "utils.flat", "flat"},
	}
	suite.addFixed(
		`(utils.deep).flat(array)`,
		`(utils.deep).flat(array)`,
		`array.flat()`,
		`utils.deep.flat()`,
		pathOptions,
	)
	suite.addFixed(
		`((utils.flat))(array)`,
		`((utils.flat))(array)`,
		`array.flat()`,
		`utils.flat()`,
		pathOptions,
	)
	suite.addValid(pathOptions,
		`utils?.deep.flat(array)`,
		`utils.deep?.flat(array)`,
		`utils["deep"].flat(array)`,
	)
	suite.addFixed(
		`flat(array.flatMap(value => other))`,
		`flat(array.flatMap(value => other))`,
		`array.flatMap(value => other).flat()`,
		`flat()`,
		pathOptions,
	)

	// A configured function can intentionally overlap a built-in matcher.
	// Both diagnostics are retained, while the first case's overlapping fix
	// wins just as it does in ESLint.
	overlapCode := `array.flatMap(value => value)`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code:    overlapCode,
		Output:  []string{`array.flat()`},
		Options: map[string]interface{}{"functions": []interface{}{"array.flatMap"}},
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(overlapCode, overlapCode, `Array#flatMap()`, 0),
			upstreamError(overlapCode, overlapCode, `array.flatMap()`, 0),
		},
	})

	// All three calls report on the first pass. Their overlapping replacements
	// are then applied one enclosing call at a time.
	nestedCode := `flat(flat(_.flatten(array)))`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code: nestedCode,
		Output: []string{
			`flat(_.flatten(array)).flat()`,
			`_.flatten(array).flat().flat()`,
			`array.flat().flat().flat()`,
		},
		Options: pathOptions,
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(nestedCode, nestedCode, `flat()`, 0),
			upstreamError(nestedCode, `flat(_.flatten(array))`, `flat()`, 0),
			upstreamError(nestedCode, `_.flatten(array)`, `_.flatten()`, 0),
		},
	})

	suite.run(t)
}
