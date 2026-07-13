package no_cycle_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoCycleExtras locks in branches, real-user shapes, and tsgo AST edge
// cases that the upstream test suite doesn't exercise. Upstream-migrated cases
// live in no_cycle_upstream_test.go.
func TestNoCycleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_cycle.NoCycleRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: empty and non-string call arguments are not module edges ----
			{Code: `const name = "./no-cycle/depth-one"; import(name); require(); define([name]); ` + rootExports, Options: map[string]interface{}{"commonjs": true, "amd": true}},

			// ---- Dimension 4: object keys are ordinary string literals, not module source literals ----
			{Code: `const keyed = { "./no-cycle/depth-one": true, ["./no-cycle/depth-two"]: false }; ` + rootExports},

			// ---- Dimension 4: string literals inside declarations are not module source literals ----
			{Code: `class Local { method() { return "./no-cycle/depth-one"; } } ` + rootExports},

			// ---- Dimension 4: `export type * from` stays type-only and does not form a graph edge ----
			{Code: `export type * from "./no-cycle/reexport-type-only"; ` + rootExports},

			// ---- Dimension 4: CommonJS require() must have exactly one argument to be a module edge ----
			{Code: `const common = require("./no-cycle/commonjs-depth-one", "extra"); ` + rootExports, Options: map[string]interface{}{"commonjs": true}},

			// ---- Real-user: #2265 dynamic import cycles are permitted when explicitly allowed ----
			{Code: `const lazy = () => import("./no-cycle/dynamic-depth-one"); ` + rootExports, Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}},

			// ---- Real-user: #1647 ignoreExternal skips external folders reached through a relative intermediate ----
			{Code: `import { externalDepthTwo } from "./no-cycle/external-depth-two"; export const rootValue = externalDepthTwo; export type RootType = string;`, Options: map[string]interface{}{"ignoreExternal": true}, Settings: map[string]interface{}{"import/external-module-folders": []interface{}{"no-cycle/external"}}},

			// Locks in upstream checkSourceValue() arm 1: unresolved imports return without reporting.
			{Code: `import missing from "./no-cycle/does-not-exist"; ` + rootExports},

			// Locks in upstream checkSourceValue() arm 2: direct self imports are delegated to import/no-self-import.
			{Code: `import { rootValue as self } from "./file"; ` + rootExports},

			// Locks in upstream checkSourceValue() arm 3: import type declarations are ignored.
			{Code: `import type { RootType as LocalType } from "./no-cycle/type-only"; ` + rootExports},

			// Locks in upstream detectCycle() maxDepth branch: depth-three is beyond maxDepth 2.
			{Code: `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"maxDepth": 2}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized CommonJS source literals still resolve ----
			{
				Code:    `const common = require(("./no-cycle/commonjs-depth-one")); ` + rootExports,
				Options: map[string]interface{}{"commonjs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "cycle", Message: messageDetected, Line: 1, Column: 16},
				},
			},

			// ---- Dimension 4: require() inside nested code still contributes a graph edge ----
			{
				Code:    `function load() { return require("./no-cycle/commonjs-depth-one"); } ` + rootExports,
				Options: map[string]interface{}{"commonjs": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "cycle", Message: messageDetected, Line: 1},
				},
			},

			// ---- Dimension 4: dynamic import calls are checked by default ----
			{
				Code: `import("./no-cycle/dynamic-depth-one"); ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},
			{
				Code: `import(("./no-cycle/dynamic-depth-one")); ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// ---- Real-user: #1554 TS path aliases still participate in cycle detection ----
			{
				Code:     `import { aliasB } from "@cycles/alias-b"; export const aliasA = aliasB;`,
				FileName: "no-cycle/alias-a.ts",
				TSConfig: "tsconfig.no-cycle-paths.json",
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// ---- Real-user: #3147 side-effect-only imports participate in indirect cycle detection ----
			{
				Code: `import "./no-cycle/depth-three-indirect"; ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaTwoOne),
				},
			},

			// Graph-level cycles report even when exports are independent.
			{
				Code: `import { independent } from "./no-cycle/independent-b"; export const rootValue = independent; export type RootType = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// Locks in upstream detectCycle() arm 1: route strings include intermediate source and line.
			{
				Code: `import { depthTwo } from "./no-cycle/depth-two"; export const rootValue = depthTwo; export type RootType = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaDepthOne),
				},
			},

			// Locks in upstream detectCycle() arm 2: maxDepth "∞" is treated as unlimited.
			{
				Code:    `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`,
				Options: map[string]interface{}{"maxDepth": "∞"},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaTwoOne),
				},
			},
			{
				Code:    `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`,
				Options: map[string]interface{}{"maxDepth": "2"},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaTwoOne),
				},
			},
			{
				Code:    `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`,
				Options: map[string]interface{}{"maxDepth": 2.5},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaTwoOne),
				},
			},

			// Locks in upstream detectCycle() dynamic arm: allowUnsafeDynamicCyclicDependency only skips the dynamic path, not unrelated static cycles.
			{
				Code:    `import { unrelatedDynamic } from "./no-cycle/unrelated-dynamic"; export const rootValue = unrelatedDynamic; export type RootType = string;`,
				Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaDepthOne),
				},
			},

			// Locks in upstream ignoreModule() arm: ignoreExternal=false keeps external-folder paths in the graph.
			{
				Code:     `import { externalDepthTwo } from "./no-cycle/external-depth-two"; export const rootValue = externalDepthTwo; export type RootType = string;`,
				Settings: map[string]interface{}{"import/external-module-folders": []interface{}{"no-cycle/external"}},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError("Dependency cycle via ./external/depth-one:1"),
				},
			},

			// Locks in upstream checkSourceValue() arm 4: mixed type/value imports are runtime edges.
			{
				Code: `import { mixed } from "./no-cycle/mixed-type"; export const rootValue = mixed; export type RootType = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// Inline type specifiers do not make a mixed value re-export type-only.
			{
				Code: `import { mixedInlineReexport } from "./no-cycle/mixed-inline-type-reexport"; export const rootValue = mixedInlineReexport; export type RootType = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// ---- Dimension 4: named type re-exports match upstream export-map dependency edges ----
			{
				Code: `export type { RootType } from "./no-cycle/reexport-type-only"; ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},
			{
				Code: `import "./no-cycle/inline-type-reexport"; ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// ---- Real-user: rspack-style barrel reports each cyclic import and keeps route text aligned ----
			{
				Code: `import * as realBarrel from "./no-cycle/barrel-real";
import { run } from "./no-cycle/barrel-real-runtime";
export type RealEntry = string;
export const realValue = realBarrel.run || run;`,
				FileName: "barrel-real-entry.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "cycle", Message: messageDetected, Line: 1, Column: 1},
					{MessageId: "cycle", Message: "Dependency cycle via ./barrel-real-config:1", Line: 2, Column: 1},
				},
			},

			// Upstream keeps one traversal set per linted file, so duplicate imports of the same cyclic target report once.
			{
				Code: `import { depthOne } from "./no-cycle/depth-one"; import { depthOne as again } from "./no-cycle/depth-one"; export const rootValue = depthOne || again; export type RootType = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},
		},
	)
}
