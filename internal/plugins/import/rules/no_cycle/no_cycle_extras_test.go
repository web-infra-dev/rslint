package no_cycle_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
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
		withDefaultNoCycleValidFileName([]rule_tester.ValidTestCase{
			// Excluding a target, an intermediate file, or the importer breaks
			// every route into that file, including duplicate references.
			{Code: `import "./no-cycle/depth-one"; import "./no-cycle/depth-one"; ` + rootExports, Settings: map[string]interface{}{"import/ignore": []interface{}{`/depth-one\.ts$`}}},
			{Code: `import "./no-cycle/depth-two"; ` + rootExports, Settings: map[string]interface{}{"import/ignore": []interface{}{`/depth-one\.ts$`}}},
			{Code: `import "./no-cycle/depth-one"; ` + rootExports, Settings: map[string]interface{}{"import/ignore": []interface{}{`/file\.ts$`}}},
			{Code: `import "./no-cycle/depth-one"; ` + rootExports, Options: map[string]interface{}{"ignoreExternal": true}, Settings: map[string]interface{}{"import/external-module-folders": []interface{}{"no-cycle"}}},
			{Code: `import "./no-cycle/depth-one"; ` + rootExports, Options: map[string]interface{}{"ignoreExternal": true}, Settings: map[string]interface{}{"import/external-module-folders": []interface{}{""}}},
			{Code: `import "./no-cycle/depth-one"; import("./no-cycle/depth-one"); ` + rootExports, Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}, Settings: map[string]interface{}{"import/ignore": []interface{}{`/depth-one\.ts$`}}},

			// Suppression still applies to reports after graph filtering.
			{Code: "// eslint-disable-next-line test\n" + `import "./no-cycle/depth-one"; ` + rootExports},
			{Code: `/* eslint-disable test */ import "./no-cycle/depth-one"; ` + rootExports},

			// ---- JSDoc import types are comments, not module-graph edges ----
			// @typescript-eslint/parser exposes these through comments / parser
			// services rather than ESTree import nodes, so import/no-cycle ignores
			// them even though TypeScript-Go parses them again into synthetic syntax.
			{
				Code:     `/** @type {import("./no-cycle/depth-one").depthOne} */ const value = {}; export const rootValue = 1;`,
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code:     `/** @typedef {import("./no-cycle/depth-one").depthOne} DepthOne */ export const rootValue = 1;`,
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code:     `/** @import {depthOne} from "./no-cycle/depth-one" */ export const rootValue = 1;`,
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
			},

			// ---- Dimension 4: empty and non-string call arguments are not module edges ----
			{Code: `const name = "./no-cycle/depth-one"; import(name); require(); define([name]); ` + rootExports, Options: map[string]interface{}{"commonjs": true, "amd": true}},

			// ---- Dimension 4: object keys are ordinary string literals, not module source literals ----
			{Code: `const keyed = { "./no-cycle/depth-one": true, ["./no-cycle/depth-two"]: false }; ` + rootExports},

			// ---- Dimension 4: string literals inside declarations are not module source literals ----
			{Code: `class Local { method() { return "./no-cycle/depth-one"; } } ` + rootExports},

			// ---- Fast ESM collection must not widen the rule to TypeScript-only module syntax ----
			{Code: `import depthOne = require("./no-cycle/depth-one"); void depthOne; ` + rootExports},
			{Code: `type DepthOneModule = typeof import("./no-cycle/depth-one"); ` + rootExports},

			// ---- Dimension 4: `export type * from` stays type-only and does not form a graph edge ----
			{Code: `export type * from "./no-cycle/reexport-type-only"; ` + rootExports},

			// ---- Dimension 4: CommonJS require() must have exactly one argument to be a module edge ----
			{Code: `const common = require("./no-cycle/commonjs-depth-one", "extra"); ` + rootExports, Options: map[string]interface{}{"commonjs": true}},

			// ---- Real-user: #2265 dynamic import cycles are permitted when explicitly allowed ----
			{Code: `const lazy = () => import("./no-cycle/dynamic-depth-one"); ` + rootExports, Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}},

			// ---- Real-user: #1647 ignoreExternal skips external folders reached through a relative intermediate ----
			{Code: `import { externalDepthTwo } from "./no-cycle/external-depth-two"; export const rootValue = externalDepthTwo; export type RootType = string;`, Options: map[string]interface{}{"ignoreExternal": true}, Settings: map[string]interface{}{"import/external-module-folders": []interface{}{"no-cycle/external"}}},

			// ---- Dimension 3: reaching a cycle is not joining one ----
			// The linted file imports into a pair that imports each other. It is
			// not part of that component, so nothing about it is this file's
			// cycle — the rule reports membership, never reachability.
			{Code: `import { reachedPairA } from "./no-cycle/reached-pair-a"; export const rootValue = reachedPairA; export type RootType = string;`},
			{Code: `import { reachedPairA } from "./no-cycle/reached-pair-a"; export const rootValue = reachedPairA; export type RootType = string;`, Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}},

			// Locks in upstream checkSourceValue() arm 1: unresolved imports return without reporting.
			{Code: `import missing from "./no-cycle/does-not-exist"; ` + rootExports},

			// Locks in upstream checkSourceValue() arm 2: direct self imports are delegated to import/no-self-import.
			{Code: `import { rootValue as self } from "./file"; ` + rootExports},

			// Locks in upstream checkSourceValue() arm 3: import type declarations are ignored.
			{Code: `import type { RootType as LocalType } from "./no-cycle/type-only"; ` + rootExports},

			// Locks in upstream detectCycle() maxDepth branch: depth-three is beyond maxDepth 2.
			{Code: `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"maxDepth": 2}}},

			// A json.Number maxDepth is schema-valid and must limit traversal like a plain number.
			{Code: `import { depthThree } from "./no-cycle/depth-three"; export const rootValue = depthThree; export type RootType = string;`, Options: []interface{}{map[string]interface{}{"maxDepth": json.Number("2")}}},
		}),
		withDefaultNoCycleInvalidFileName([]rule_tester.InvalidTestCase{
			// Local bare aliases are not external when they resolve to local files.
			{
				Code:     `import { aliasB } from "@cycles/alias-b"; export const aliasA = aliasB;`,
				FileName: "no-cycle/alias-a.ts",
				TSConfig: "tsconfig.no-cycle-paths.json",
				Options:  map[string]interface{}{"ignoreExternal": true},
				Errors:   []rule_tester.InvalidTestCaseError{cycleError(messageDetected)},
			},
			// An empty external folder list and an unrelated ignore pattern
			// keep the original diagnostic and its complete source range.
			{
				Code:     "/* leading comment */\n" + `import "./no-cycle/depth-two"; ` + rootExports,
				Options:  map[string]interface{}{"ignoreExternal": true},
				Settings: map[string]interface{}{"import/external-module-folders": []interface{}{}, "import/ignore": []interface{}{`/missing/`}},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "cycle", Message: messageViaDepthOne, Line: 2, Column: 1, EndLine: 2, EndColumn: 31}},
			},
			// A missing target must not become an edge to the first graph node.
			{
				Code:   "import './no-cycle/does-not-exist';\n" + `import "./no-cycle/depth-one"; ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "cycle", Message: messageDetected, Line: 2, Column: 1}},
			},
			// Excluding one cyclic target must not exclude its remaining siblings.
			{
				Code:     "import './no-cycle/depth-one';\n" + `import "./no-cycle/independent-b"; ` + rootExports,
				Settings: map[string]interface{}{"import/ignore": []interface{}{`/depth-one\.ts$`}},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "cycle", Message: messageDetected, Line: 2, Column: 1}},
			},
			{
				Code:   "/* eslint-disable test */\nimport './no-cycle/depth-one';\n/* eslint-enable test */\n" + `import "./no-cycle/independent-b"; ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "cycle", Message: messageDetected, Line: 4, Column: 1}},
			},

			// Control for the JSDoc cases above: with the same JavaScript filename,
			// tsconfig, and target, an authored import is a graph edge and closes
			// the cycle through no-cycle/depth-one.ts.
			{
				Code:     `import { depthOne } from "./no-cycle/depth-one"; export const rootValue = depthOne;`,
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageDetected),
				},
			},

			// Module-augmentation bodies are not present in SourceFile.Imports and must use the full collector.
			{
				Code: `declare module "virtual" { import { depthOne } from "./no-cycle/depth-one"; export { depthOne }; } ` + rootExports,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "cycle", Message: messageDetected, Line: 1},
				},
			},

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

			// Locks in upstream detectCycle() dynamic arm: allowUnsafeDynamicCyclicDependency only skips the dynamic path, not unrelated static cycles.
			{
				Code:    `import { unrelatedDynamic } from "./no-cycle/unrelated-dynamic"; export const rootValue = unrelatedDynamic; export type RootType = string;`,
				Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError(messageViaDepthOne),
				},
			},

			// ---- Dimension 3: the route survives a file that also imports off the cycle ----
			// tail-mid imports a leaf before it imports its way back here, so the
			// search past it has to step over a file that leads nowhere. Both
			// configurations take the same confined search, so both are locked in.
			{
				Code: `import { tailMid } from "./no-cycle/tail-mid"; export const rootValue = tailMid; export type RootType = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError("Dependency cycle via ./tail-back:2"),
				},
			},
			{
				Code:    `import { tailMid } from "./no-cycle/tail-mid"; export const rootValue = tailMid; export type RootType = string;`,
				Options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
				Errors: []rule_tester.InvalidTestCaseError{
					cycleError("Dependency cycle via ./tail-back:2"),
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
		}),
	)
}

// These small Programs exercise repeated queries against the same generation;
// ordinary rule-tester cases each construct a new Program and cannot expose
// configuration-dependent answers leaking through the direct-cycle caches.
func TestNoCycleDepthOneReferenceSemantics(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		target  string
		options map[string]interface{}
		want    []string
	}{
		{
			name:   "type-only references do not consume duplicate runtime targets",
			source: "import type { B } from './b';\nimport './b';\nimport './b.ts';",
			target: "import './a'; export type B = string;",
			want:   []string{"import './b';"},
		},
		{
			name:    "a skipped dynamic reference does not consume the static target",
			source:  "import('./b');\nimport './b';",
			target:  "import './a';",
			options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
			want:    []string{"import './b';"},
		},
		{
			name:    "a later dynamic source reference does not suppress the static one",
			source:  "import './b';\nimport('./b');",
			target:  "import './a';",
			options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
			want:    []string{"import './b';"},
		},
		{
			name:   "the first dynamic source reports when unsafe cycles are forbidden",
			source: "import('./b');\nimport './b';",
			target: "import './a';",
			want:   []string{"import('./b')"},
		},
		{
			name:   "self imports do not consume the return edge",
			source: "import './a';\nimport './b';",
			target: "import './a';",
			want:   []string{"import './b';"},
		},
		{
			name:    "a later dynamic back edge withholds the static back edge",
			source:  "import './b';",
			target:  "import './a';\nimport('./a');",
			options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
		},
		{
			name:    "a later static back edge does not erase the dynamic back edge",
			source:  "import './b';",
			target:  "import('./a');\nimport './a';",
			options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
		},
		{
			name:   "mixed back edges report when unsafe cycles are forbidden",
			source: "import './b';",
			target: "import('./a');\nimport './a';",
			want:   []string{"import './b';"},
		},
		{
			name:    "unrelated and unresolved dynamic imports do not withhold self",
			source:  "import './b';",
			target:  "import('./c');\nimport('./missing');\nimport './a';",
			options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true},
			want:    []string{"import './b';"},
		},
		{
			name:   "a type-only return edge is not a cycle",
			source: "import './b'; export type A = string;",
			target: "import type { A } from './a';",
		},
		{
			name:   "named type reexports retain upstream dependency semantics",
			source: "import './b'; export type A = string;",
			target: "export type { A } from './a';",
			want:   []string{"import './b';"},
		},
		{
			name:    "dynamic references outside selected kinds do not withhold CommonJS",
			source:  "require('./b');",
			target:  "require('./a');\nimport('./a');",
			options: map[string]interface{}{"esmodule": false, "commonjs": true, "allowUnsafeDynamicCyclicDependency": true},
			want:    []string{"require('./b')"},
		},
		{
			name:    "AMD references form direct cycles when selected",
			source:  "define(['./b'], () => {});",
			target:  "require(['./a'], () => {});",
			options: map[string]interface{}{"esmodule": false, "amd": true},
			want:    []string{"define(['./b'], () => {})"},
		},
		{
			name:   "unselected call references cannot close a cycle",
			source: "require('./b'); define(['./b'], () => {});",
			target: "require('./a'); define(['./a'], () => {});",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, roots := noCycleDepthOneRoot(t, map[string]string{
				"a.ts": tt.source,
				"b.ts": tt.target,
				"c.ts": "export {};",
			})
			p := noCycleDepthOneProgram(t, root, roots)
			options := map[string]interface{}{"maxDepth": 1}
			for key, value := range tt.options {
				options[key] = value
			}
			// Both cold and cached queries must report the same first source.
			for range 2 {
				reports := noCycleDepthOneReports(t, p, root, "a.ts", options, nil, rule.EditDemandNone)
				assertNoCycleDepthOneReports(t, reports, tt.want)
			}
		})
	}
}

func TestNoCycleDepthOneCacheIsolation(t *testing.T) {
	files := map[string]string{
		"a.ts":          "import './local';\nimport './external/b';\nimport './dynamic';\nrequire('./common');\ndefine(['./amd'], () => {});",
		"local.ts":      "import './a'; import './other';",
		"external/b.ts": "import '../a';",
		"dynamic.ts":    "import './a'; import('./a');",
		"common.ts":     "require('./a'); import('./a');",
		"amd.ts":        "define(['./a'], () => {});",
		"other.ts":      "import './local';",
		"unrelated.ts":  "import './local';",
	}
	local, external, dynamic := "import './local';", "import './external/b';", "import './dynamic';"
	tests := []struct {
		name     string
		options  map[string]interface{}
		settings map[string]interface{}
		want     []string
	}{
		{name: "ignored target", settings: map[string]interface{}{"import/ignore": []interface{}{`/local\.ts$`}}, want: []string{external, dynamic}},
		{name: "external target", options: map[string]interface{}{"ignoreExternal": true}, settings: map[string]interface{}{"import/external-module-folders": []interface{}{"external"}}, want: []string{local, dynamic}},
		{name: "unsafe enabled", options: map[string]interface{}{"allowUnsafeDynamicCyclicDependency": true}, want: []string{local, external}},
		{name: "no filters", want: []string{local, external, dynamic}},
		{name: "external option disabled", settings: map[string]interface{}{"import/external-module-folders": []interface{}{"external"}}, want: []string{local, external, dynamic}},
		{name: "empty external folders", options: map[string]interface{}{"ignoreExternal": true}, settings: map[string]interface{}{"import/external-module-folders": []interface{}{}}, want: []string{local, external, dynamic}},
		{name: "ignored self", settings: map[string]interface{}{"import/ignore": []interface{}{`/a\.ts$`}}},
		{name: "CommonJS only", options: map[string]interface{}{"esmodule": false, "commonjs": true, "allowUnsafeDynamicCyclicDependency": true}, want: []string{"require('./common')"}},
		{name: "ESM and CommonJS unsafe", options: map[string]interface{}{"commonjs": true, "allowUnsafeDynamicCyclicDependency": true}, want: []string{local, external}},
		{name: "AMD only", options: map[string]interface{}{"esmodule": false, "amd": true}, want: []string{"define(['./amd'], () => {})"}},
		{name: "all syntax kinds", options: map[string]interface{}{"commonjs": true, "amd": true}, want: []string{local, external, dynamic, "require('./common')", "define(['./amd'], () => {})"}},
		{name: "no syntax kinds", options: map[string]interface{}{"esmodule": false}},
	}
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse=%t", reverse), func(t *testing.T) {
			root, roots := noCycleDepthOneRoot(t, files)
			p := noCycleDepthOneProgram(t, root, roots)
			for index := range tests {
				if reverse {
					index = len(tests) - 1 - index
				}
				tt := tests[index]
				t.Run(tt.name, func(t *testing.T) {
					options := map[string]interface{}{"maxDepth": 1}
					for key, value := range tt.options {
						options[key] = value
					}
					for range 2 {
						reports := noCycleDepthOneReports(t, p, root, "a.ts", options, tt.settings, rule.EditDemandNone)
						assertNoCycleDepthOneReports(t, reports, tt.want)
					}
				})
			}
			// The same target's cached return set must answer separately for
			// another importer it reaches and an importer it does not reach.
			options := map[string]interface{}{"maxDepth": 1}
			assertNoCycleDepthOneReports(t, noCycleDepthOneReports(t, p, root, "other.ts", options, nil, rule.EditDemandNone), []string{local})
			assertNoCycleDepthOneReports(t, noCycleDepthOneReports(t, p, root, "unrelated.ts", options, nil, rule.EditDemandNone), nil)
		})
	}
}

func TestNoCycleDepthOneDepthBoundary(t *testing.T) {
	root, roots := noCycleDepthOneRoot(t, map[string]string{
		"a.ts": "import './b';",
		"b.ts": "import './c';",
		"c.ts": "import './a';",
	})
	p := noCycleDepthOneProgram(t, root, roots)
	for _, depth := range []int{1, 2, 1} {
		reports := noCycleDepthOneReports(t, p, root, "a.ts", map[string]interface{}{"maxDepth": depth}, nil, rule.EditDemandNone)
		if depth == 1 {
			assertNoCycleDepthOneReports(t, reports, nil)
		} else if len(reports) != 1 || reports[0].Message.Description != "Dependency cycle via ./c:1" {
			t.Fatalf("maxDepth 2 reports = %v, want the indirect cycle", reports)
		}
	}
}

func TestNoCycleDepthOneConcurrentQueries(t *testing.T) {
	t.Parallel()
	root, roots := noCycleDepthOneRoot(t, map[string]string{
		"a.ts": "import './b';",
		"b.ts": "import './a'; import('./a');",
	})
	p := noCycleDepthOneProgram(t, root, roots)
	for index := range 24 {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			unsafe := index%2 == 0
			demand := []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAutofix | rule.EditDemandSuggestion}[(index/2)%4]
			options := map[string]interface{}{"maxDepth": 1, "allowUnsafeDynamicCyclicDependency": unsafe}
			var want []string
			if !unsafe {
				want = []string{"import './b';"}
			}
			for range 3 {
				assertNoCycleDepthOneReports(t, noCycleDepthOneReports(t, p, root, "a.ts", options, nil, demand), want)
			}
		})
	}
}

func TestNoCycleDepthOneExcludedTargetsAreNotCollected(t *testing.T) {
	for _, ignoreExternal := range []bool{false, true} {
		t.Run(fmt.Sprintf("ignoreExternal=%t", ignoreExternal), func(t *testing.T) {
			root, roots := noCycleDepthOneRoot(t, map[string]string{
				"a.ts":          "import './external/b';",
				"external/b.ts": "import '../a'; define(['no-cycle-demand-probe'], () => {});",
			})
			root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
				tspath.ResolvePath(root.Dir, "node_modules/no-cycle-demand-probe/package.json"): `{"name":"no-cycle-demand-probe","types":"index.d.ts"}`,
				tspath.ResolvePath(root.Dir, "node_modules/no-cycle-demand-probe/index.d.ts"):   "export {};",
			})
			probe := &noCycleDepthOneProbeFS{FS: root.FS}
			root.FS = probe
			p := noCycleDepthOneProgram(t, root, roots)
			probe.probes.Store(0)
			options := map[string]interface{}{"maxDepth": 1, "amd": true, "ignoreExternal": ignoreExternal}
			settings := map[string]interface{}{"import/ignore": []interface{}{`/external/`}}
			if ignoreExternal {
				settings = map[string]interface{}{"import/external-module-folders": []interface{}{"external"}}
			}
			assertNoCycleDepthOneReports(t, noCycleDepthOneReports(t, p, root, "a.ts", options, settings, rule.EditDemandNone), nil)
			if got := probe.probes.Load(); got != 0 {
				t.Fatalf("excluded target caused %d module-resolution probes", got)
			}
			// The positive control makes this fail if the probe cannot see
			// AMD resolution, and also checks a cached negative answer
			// does not survive into the unfiltered configuration.
			assertNoCycleDepthOneReports(t, noCycleDepthOneReports(t, p, root, "a.ts", options, nil, rule.EditDemandNone), []string{"import './external/b';"})
			if probe.probes.Load() == 0 {
				t.Fatal("unfiltered target did not resolve the AMD probe")
			}
		})
	}
}

type noCycleDepthOneProbeFS struct {
	vfs.FS
	probes atomic.Int64
}

func (fs *noCycleDepthOneProbeFS) FileExists(path string) bool {
	if strings.Contains(path, "no-cycle-demand-probe") {
		fs.probes.Add(1)
	}
	return fs.FS.FileExists(path)
}

func noCycleDepthOneRoot(t *testing.T, files map[string]string) (rule_tester.Root, []string) {
	t.Helper()
	if len(files) == 0 {
		t.Fatal("no-cycle depth-one fixture has no files")
	}
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-cycle-depth-one-extras")
	normalized := make(map[string]string, len(files))
	roots := make([]string, 0, len(files))
	for name, code := range files {
		path := tspath.ResolvePath(root.Dir, name)
		normalized[path] = code
		roots = append(roots, path)
	}
	slices.Sort(roots)
	root.FS = utils.NewOverlayVFS(root.FS, normalized)
	return root, roots
}

func noCycleDepthOneProgram(t *testing.T, root rule_tester.Root, roots []string) *program.Program {
	t.Helper()
	p, err := program.NewFromRoots(program.RootOptions{
		RootFileNames: roots,
		Host:          utils.CreateCompilerHost(root.Dir, root.FS),
		CompilerOptions: &core.CompilerOptions{
			NoLib:            core.TSTrue,
			Module:           core.ModuleKindESNext,
			ModuleResolution: core.ModuleResolutionKindBundler,
		},
		SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.SourceFiles()) != len(roots) {
		t.Fatalf("loaded %d source files, want %d", len(p.SourceFiles()), len(roots))
	}
	return p
}

func noCycleDepthOneReports(t *testing.T, p *program.Program, root rule_tester.Root, name string, options, settings map[string]interface{}, demand rule.EditDemand) []rule.RuleDiagnostic {
	t.Helper()
	file := p.GetSourceFile(tspath.ResolvePath(root.Dir, name))
	if file == nil {
		t.Fatalf("no-cycle depth-one fixture is missing %s", name)
	}
	var reports []rule.RuleDiagnostic
	ctx := (rule.RuleContext{SourceFile: file, Settings: settings}).WithProgram(p).WithDiagnosticConsumer(
		no_cycle.NoCycleRule.Name, rule.SeverityWarning,
		rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { reports = append(reports, diagnostic) }},
	)
	no_cycle.NoCycleRule.Run(ctx, []any{options})
	for _, diagnostic := range reports {
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
			t.Fatal("import/no-cycle offered an edit")
		}
	}
	return reports
}

func assertNoCycleDepthOneReports(t *testing.T, reports []rule.RuleDiagnostic, want []string) {
	t.Helper()
	var got []string
	for _, diagnostic := range reports {
		if diagnostic.Message.Id != "cycle" || diagnostic.Message.Description != messageDetected {
			t.Fatalf("unexpected direct-cycle message: %v", diagnostic.Message)
		}
		got = append(got, diagnostic.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	if !slices.Equal(got, want) {
		t.Fatalf("reported source ranges = %q, want %q", got, want)
	}
}
