package prefer_array_flat_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferArrayFlatExtrasJSDoc locks in JavaScript JSDoc wrappers across
// callees, receivers, configured paths, and reduce patterns while keeping
// authored TypeScript wrappers and optional/computed calls visible. Upstream
// cases and other augmentation live in the sibling upstream and extras files.
func TestPreferArrayFlatExtrasJSDoc(t *testing.T) {
	var suite upstreamSuite

	for _, testCase := range []struct {
		fileName string
		code     string
		target   string
		output   string
	}{
		{
			fileName: "file.js",
			code:     `/** @type {any} */ ([].concat)(...array)`,
			target:   `([].concat)(...array)`,
			output:   `/** @type {any} */ array.flat()`,
		},
		{
			fileName: "file.js",
			code:     `/** @satisfies {any} */ ([].concat)(...array)`,
			target:   `([].concat)(...array)`,
			output:   `/** @satisfies {any} */ array.flat()`,
		},
		{
			fileName: "file.jsx",
			code:     `const view = <div>{/** @type {any} */ ([].concat)(...array)}</div>;`,
			target:   `([].concat)(...array)`,
			output:   `const view = <div>{/** @type {any} */ array.flat()}</div>;`,
		},
	} {
		suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
			Code:     testCase.code,
			FileName: testCase.fileName,
			Output:   []string{testCase.output},
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamError(testCase.code, testCase.target, `[].concat()`, 0),
			},
		})
	}

	const internalJSDoc = `((/** @type {any} */ ([].concat)))(...array)`
	suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
		Code:     internalJSDoc,
		FileName: "file.js",
		Errors: []rule_tester.InvalidTestCaseError{
			upstreamError(internalJSDoc, internalJSDoc, `[].concat()`, 0),
		},
	})

	// Receiver/root wrappers are transparent to ESTree matching, but their
	// comments sit outside the selected replacement expression and suppress
	// the fix.
	for _, testCase := range []struct {
		code        string
		description string
		options     any
	}{
		{
			code:        `(/** @type {any[]} */ ([])).concat(...array)`,
			description: `[].concat()`,
		},
		{
			code:        `(/** @satisfies {any[]} */ []).concat(...array)`,
			description: `[].concat()`,
		},
		{
			code:        `(/** @type {any} */ Array.prototype.concat).apply([], array)`,
			description: `Array.prototype.concat()`,
		},
		{
			code:        `(/** @type {typeof Array} */ Array).prototype.concat.apply([], array)`,
			description: `Array.prototype.concat()`,
		},
		{
			code:        `(/** @type {any} */ _).flatten(array)`,
			description: `_.flatten()`,
		},
		{
			code:        `(/** @type {any} */ utils).flat(array)`,
			description: `utils.flat()`,
			options: map[string]any{
				"functions": []any{"utils.flat"},
			},
		},
		{
			code:        `array.reduce((a,b)=>(/** @type {any[]} */ a).concat(b),[])`,
			description: `Array#reduce()`,
		},
		{
			code:        `array.reduce((a,b)=>a.concat(/** @type {any[]} */ b),[])`,
			description: `Array#reduce()`,
		},
		{
			code:        `(/** @type {Set<number[]>} */ foo).reduce((a,b)=>a.concat(b),[])`,
			description: `Array#reduce()`,
		},
		{
			code:        `(/** @type {any[]} */ []).concat.apply([], array)`,
			description: `Array.prototype.concat()`,
		},
	} {
		suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
			Code:     testCase.code,
			FileName: "file.js",
			Options:  testCase.options,
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamError(testCase.code, testCase.code, testCase.description, 0),
			},
		})
	}

	// A JSDoc wrapper around the whole callee/call is outside the reported
	// CallExpression range, so the safe fix remains available.
	for _, testCase := range []struct {
		code        string
		target      string
		output      string
		description string
	}{
		{
			code:        `/** @type {any} */ (_.flatten)(array)`,
			target:      `(_.flatten)(array)`,
			output:      `/** @type {any} */ array.flat()`,
			description: `_.flatten()`,
		},
		{
			code:        `/** @type {any} */ (([].concat(...array)))`,
			target:      `[].concat(...array)`,
			output:      `/** @type {any} */ ((array.flat()))`,
			description: `[].concat()`,
		},
	} {
		suite.invalid = append(suite.invalid, rule_tester.InvalidTestCase{
			Code:     testCase.code,
			FileName: "file.js",
			Output:   []string{testCase.output},
			Errors: []rule_tester.InvalidTestCaseError{
				upstreamError(testCase.code, testCase.target, testCase.description, 0),
			},
		})
	}

	// JSDoc types do not turn syntactically known flatMap non-arrays into
	// arrays. Authored TypeScript wrappers remain visible and unmatched.
	suite.valid = append(suite.valid,
		rule_tester.ValidTestCase{
			Code:     `(/** @type {any[]} */ Effects).flatMap(x => x)`,
			FileName: "file.js",
		},
		rule_tester.ValidTestCase{
			Code:     `const value = {}; (/** @type {any[]} */ value).flatMap(x => x)`,
			FileName: "file.js",
		},
		rule_tester.ValidTestCase{
			Code:     `([] as unknown[]).concat(...array)`,
			FileName: "file.ts",
		},
		rule_tester.ValidTestCase{
			Code:     `(_.flatten as any)(array)`,
			FileName: "file.ts",
		},
	)

	for _, code := range []string{
		`/** @type {any} */ (value.concat)(...array)`,
		`/** @type {any} */ ([]?.concat)(...array)`,
		`/** @type {any} */ ([].concat)?.(...array)`,
		`/** @type {any} */ ([]["concat"])(...array)`,
	} {
		suite.valid = append(suite.valid, rule_tester.ValidTestCase{
			Code:     code,
			FileName: "file.js",
		})
	}

	for _, code := range []string{
		`([].concat as any)(...array)`,
		`([].concat satisfies any)(...array)`,
		`([].concat!)(...array)`,
	} {
		suite.valid = append(suite.valid, rule_tester.ValidTestCase{
			Code:     code,
			FileName: "file.ts",
		})
	}

	suite.run(t)
}
