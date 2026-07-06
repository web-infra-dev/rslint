package no_cycle_test

import (
	"math"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	rootExports        = `export const rootValue = 1; export type RootType = string;`
	messageDetected    = "Dependency cycle detected."
	messageViaDepthOne = "Dependency cycle via ./depth-one:1"
	messageViaTwoOne   = "Dependency cycle via ./depth-two:1=>./depth-one:1"
)

// TestNoCycleUpstream migrates the full supported valid/invalid suite from
// upstream tests/src/rules/no-cycle.js 1:1 by semantic group. Unsupported
// parser/resolver framework cases stay present as skipped cases. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases live in no_cycle_extras_test.go.
func TestNoCycleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_cycle.NoCycleRule,
		withDisableSccValid(noCycleUpstreamValidCases()),
		withDisableSccInvalid(noCycleUpstreamInvalidCases()),
	)
}

func noCycleUpstreamValidCases() []rule_tester.ValidTestCase {
	return []rule_tester.ValidTestCase{
		// ---- upstream valid: this rule does not report self imports or unresolved imports ----
		{Code: `import { rootValue as other } from "./file"; ` + rootExports},
		{Code: `import foo from "./foo.js"; ` + rootExports},

		// ---- upstream valid: unresolved and external modules are no-unresolved territory ----
		{Code: `import _ from "lodash"; ` + rootExports},
		{Code: `import foo from "@scope/foo"; ` + rootExports},
		{Code: `import { noCycle } from "./no-cycle/no-cycle"; export const rootValue = noCycle; export type RootType = string;`},

		// ---- upstream valid: commonjs and amd are off by default ----
		{Code: `var _ = require("lodash"); ` + rootExports},
		{Code: `var find = require("lodash.find"); ` + rootExports},
		{Code: `var foo = require("./foo"); ` + rootExports},
		{Code: `var foo = require("../foo"); ` + rootExports},
		{Code: `var foo = require("foo"); ` + rootExports},
		{Code: `var foo = require("./"); ` + rootExports},
		{Code: `var foo = require("@scope/foo"); ` + rootExports},
		{Code: `var bar = require("./bar/index"); ` + rootExports},
		{Code: `var bar = require("./bar"); ` + rootExports},
		{Code: `const common = require("./no-cycle/commonjs-depth-one"); ` + rootExports},
		{Code: `define(["./no-cycle/amd-depth-one"], () => ({})); ` + rootExports},

		// SKIP: rule_tester always maps test files into the fixture VFS, so it cannot produce ESLint's literal "<text>" physical filename.
		{Code: `var bar = require("./bar");`, FileName: "<text>", Skip: true},
		// SKIP: rslint Go rule tests do not use upstream's webpack resolver settings.
		{Code: `import { foo } from "cycles/external/depth-one";`, Options: []interface{}{map[string]interface{}{"ignoreExternal": true}}, Skip: true},
		// SKIP: rslint Go rule tests do not use upstream's webpack resolver settings.
		{Code: `import { foo } from "./external-depth-two";`, Options: []interface{}{map[string]interface{}{"ignoreExternal": true}}, Skip: true},

		// ---- upstream valid: maxDepth limits traversal beyond direct cycles ----
		{Code: `import { depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthTwo; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"maxDepth": 1}}},
		{Code: `import { depthOne, depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthOne || depthTwo; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"maxDepth": 1}}},
		{Code: `import("./no-cycle/depth-two").then(({ depthTwo }) => depthTwo); ` + rootExports, Options: []interface{}{map[string]interface{}{"maxDepth": 1}}},
		{Code: `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`, Options: map[string]interface{}{"maxDepth": 2}},

		// ---- upstream valid: type-only imports do not form runtime cycles ----
		{Code: `import type { RootType as LocalType } from "./no-cycle/type-only"; ` + rootExports},
		{Code: `import type { RootType as LocalType, TypeOnly } from "./no-cycle/type-only"; ` + rootExports},
		{Code: `import { typeOnly } from "./no-cycle/type-only"; export const rootValue = typeOnly; export type RootType = string;`},
		{Code: `import { inlineTypeOnly } from "./no-cycle/inline-type-only"; export const rootValue = inlineTypeOnly; export type RootType = string;`},

		// ---- upstream valid: dynamic cycle allowed by option (#2265) ----
		{Code: `function bar(){ return import("./no-cycle/depth-one"); } ` + rootExports, Options: []interface{}{map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}}},
		{Code: `import { dynamicDepthOne } from "./no-cycle/depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}}},
		{Code: `import { loadRoot } from "./no-cycle/dynamic-depth-one"; export const rootValue = loadRoot; export type RootType = string;`, Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}},
		{Code: `import { loadDepthOne } from "./no-cycle/dynamic-middle"; export const rootValue = loadDepthOne; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}}},

		// ---- upstream valid: ignoreExternal skips external-looking specifiers ----
		{Code: `import { anything } from "external-package"; ` + rootExports, Options: map[string]interface{}{"ignoreExternal": true}},

		// ---- upstream valid: esmodule=false disables import/export source checks ----
		{Code: `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne; export type RootType = string;`, Options: map[string]interface{}{"esmodule": false}},

		// SKIP: rslint does not parse Flow import type/typeof syntax with the Babel parser.
		{Code: `import { bar } from "./flow-types"`, Skip: true},
		// SKIP: rslint does not parse Flow import type/typeof syntax with the Babel parser.
		{Code: `import { bar } from "./flow-types-only-importing-type"`, Skip: true},
		// SKIP: rslint does not parse Flow import type/typeof syntax with the Babel parser.
		{Code: `import { bar } from "./flow-types-only-importing-multiple-types"`, Skip: true},
		// SKIP: rslint does not parse Flow `import typeof`.
		{Code: `import { bar } from "./flow-typeof"`, Skip: true},
	}
}

func noCycleUpstreamInvalidCases() []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		// SKIP: rslint does not parse Flow import type syntax with the Babel parser.
		{Code: `import { bar } from "./flow-types-some-type-imports"`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{cycleError(messageDetected)}},
		// SKIP: rslint Go rule tests do not use upstream's webpack resolver settings.
		{Code: `import { foo } from "cycles/external/depth-one"`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{cycleError(messageDetected)}},
		// SKIP: rslint Go rule tests do not use upstream's webpack resolver settings.
		{Code: `import { foo } from "./external-depth-two"`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{cycleError("Dependency cycle via cycles/external/depth-one:1")}},

		// ---- upstream invalid: direct ES module cycle ----
		{
			Code: `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		{
			Code:    `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne; export type RootType = string;`,
			Options: []interface{}{map[string]interface{}{}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		{
			Code:    `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne; export type RootType = string;`,
			Options: map[string]interface{}{"maxDepth": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		{
			Code:    `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne; export type RootType = string;`,
			Options: []interface{}{map[string]interface{}{"maxDepth": 1}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},

		// ---- upstream invalid: CommonJS and AMD sources when enabled ----
		{
			Code:    `const { common } = require("./no-cycle/commonjs-depth-one"); ` + rootExports,
			Options: []interface{}{map[string]interface{}{"commonjs": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "cycle", Message: messageDetected, Line: 1, Column: 20},
			},
		},
		{
			Code:    `require(["./no-cycle/amd-depth-one"], d1 => {}); ` + rootExports,
			Options: []interface{}{map[string]interface{}{"amd": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		{
			Code:    `define(["./no-cycle/amd-depth-one"], d1 => {}); ` + rootExports,
			Options: []interface{}{map[string]interface{}{"amd": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},

		// ---- upstream invalid: re-export cycles are import graph edges ----
		{
			Code: `import { rootValue as reexported } from "./no-cycle/reexport-depth-one"; export const rootValue = reexported; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},

		// ---- upstream invalid: dependency path route is reported ----
		{
			Code: `import { depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthTwo; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaDepthOne),
			},
		},
		{
			Code:    `import { depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthTwo; export type RootType = string;`,
			Options: []interface{}{map[string]interface{}{"maxDepth": 2}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaDepthOne),
			},
		},
		{
			Code:    `const { depthTwo } = require("./no-cycle/depth-two"); export const rootValue = depthTwo; export type RootType = string;`,
			Options: []interface{}{map[string]interface{}{"commonjs": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "cycle", Message: messageViaDepthOne, Line: 1, Column: 22},
			},
		},
		{
			Code: `import { two } from "./no-cycle/depth-three-star"; export const rootValue = two.depthTwo; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaTwoOne),
			},
		},
		{
			Code: `import one, { two, three } from "./no-cycle/depth-three-star"; export const rootValue = one || two || three; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaTwoOne),
			},
		},
		{
			Code: `import { depthThreeIndirect } from "./no-cycle/depth-three-indirect"; export const rootValue = depthThreeIndirect; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaTwoOne),
			},
		},
		{
			Code: `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaTwoOne),
			},
		},
		{
			Code:    `import { depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthTwo; export type RootType = string;`,
			Options: []interface{}{map[string]interface{}{"maxDepth": math.Inf(1)}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaDepthOne),
			},
		},
		{
			Code:    `import { depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthTwo; export type RootType = string;`,
			Options: map[string]interface{}{"maxDepth": "∞"},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaDepthOne),
			},
		},

		// ---- upstream invalid: dynamic cycles are reported by default ----
		{
			Code: `import("./no-cycle/depth-three-star"); ` + rootExports,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaTwoOne),
			},
		},
		{
			Code: `import("./no-cycle/depth-three-indirect"); ` + rootExports,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageViaTwoOne),
			},
		},
		{
			Code: `function bar(){ return import("./no-cycle/depth-one"); } ` + rootExports,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "cycle", Message: messageDetected, Line: 1, Column: 24},
			},
		},
		{
			Code: `import { dynamicDepthOne } from "./no-cycle/depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		{
			Code:    `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne; export type RootType = string;`,
			Options: []interface{}{map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},

		// ---- upstream invalid: mixed value/type imports remain runtime imports ----
		{
			Code: `import { mixed } from "./no-cycle/mixed-type"; export const rootValue = mixed; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		// ---- upstream invalid: named type re-exports still participate in the export-map dependency graph ----
		{
			Code: `import "./no-cycle/reexport-type-only"; ` + rootExports,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
		// SKIP: rslint does not parse Flow import type syntax with the Babel parser.
		{Code: `import { bar } from "./flow-types-depth-one"`, Skip: true, Errors: []rule_tester.InvalidTestCaseError{cycleError("Dependency cycle via ./flow-types-depth-two:4=>./es6/depth-one:1")}},

		// ---- upstream invalid: disabled nested rule config does not remove graph edges ----
		{
			Code: `import { intermediateIgnoredValue } from "./no-cycle/intermediate-ignore"; export const rootValue = intermediateIgnoredValue; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError("Dependency cycle via ./ignore:1"),
			},
		},
		{
			Code: `import { ignoredValue } from "./no-cycle/ignore"; export const rootValue = ignoredValue; export type RootType = string;`,
			Errors: []rule_tester.InvalidTestCaseError{
				cycleError(messageDetected),
			},
		},
	}
}

func withDisableSccValid(cases []rule_tester.ValidTestCase) []rule_tester.ValidTestCase {
	withVariants := make([]rule_tester.ValidTestCase, 0, len(cases)*2)
	for _, testCase := range cases {
		withVariants = append(withVariants, testCase)
		variant := testCase
		variant.Code += ` // disableScc=true`
		variant.Options = withOption(testCase.Options, "disableScc", true)
		withVariants = append(withVariants, variant)
	}
	return withVariants
}

func withDisableSccInvalid(cases []rule_tester.InvalidTestCase) []rule_tester.InvalidTestCase {
	withVariants := make([]rule_tester.InvalidTestCase, 0, len(cases)*2)
	for _, testCase := range cases {
		withVariants = append(withVariants, testCase)
		variant := testCase
		variant.Code += ` // disableScc=true`
		variant.Options = withOption(testCase.Options, "disableScc", true)
		withVariants = append(withVariants, variant)
	}
	return withVariants
}

func withOption(options any, name string, value any) any {
	switch typed := options.(type) {
	case nil:
		return map[string]interface{}{name: value}
	case map[string]interface{}:
		next := make(map[string]interface{}, len(typed)+1)
		for key, item := range typed {
			next[key] = item
		}
		next[name] = value
		return next
	case []interface{}:
		if len(typed) == 0 {
			return []interface{}{map[string]interface{}{name: value}}
		}
		if first, ok := typed[0].(map[string]interface{}); ok {
			nextFirst := make(map[string]interface{}, len(first)+1)
			for key, item := range first {
				nextFirst[key] = item
			}
			nextFirst[name] = value
			next := append([]interface{}{}, typed...)
			next[0] = nextFirst
			return next
		}
	}

	return map[string]interface{}{name: value}
}

func cycleError(message string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "cycle",
		Message:   message,
		Line:      1,
		Column:    1,
	}
}
