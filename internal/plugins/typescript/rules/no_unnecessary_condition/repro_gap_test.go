package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Alignment tests for patterns that differ between typescript-eslint v6.21.0
// and v8.x (latest). A consumer project comparison (portal, on v6.21.0) flagged
// these as "Rslint-only" reports. Verified by running latest typescript-eslint
// (v8.59.0) on the same snippets — the stricter behavior below matches latest.

// Pattern A: closure-captured `let` variable whose narrowing is preserved into
// the closure by TypeScript's CFA. `(x || [])` LHS is always truthy when x was
// narrowed to a non-nullable value before the closure was created.
func TestAlign_PatternA_ClosureNarrowedLet(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
declare const map: Map<string, Set<number>>;
declare function callback(opts: { onEvent: () => void }): void;

function run(id: string) {
  let taskSessionList = map.get(id);
  if (!taskSessionList) {
    taskSessionList = new Set<number>();
    map.set(id, taskSessionList);
  }
  callback({
    onEvent: () => {
      const sessionInfo = [...(taskSessionList || [])].find((s) => s === 1);
      return sessionInfo;
    },
  });
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "alwaysTruthy"},
				},
			},
		},
	)
}

// Pattern B: `?.` on index access.
//   - When the value type is genuinely nullable (`T[] | undefined`), the chain
//     is necessary — no report.
//   - When a preceding `if (!map[k]) map[k] = []` guard persists narrowing,
//     the value is non-nullable at the call site — report (matches latest).
func TestAlign_PatternB_IndexAccessOptionalChain(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{
			{Code: `
declare const scmProjectMap: Record<string, string[] | undefined>;
scmProjectMap['x']?.push('entry');
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
const scmProjectMap: Record<string, string[] | undefined> = {};
function run(scmName: string, entry: string) {
  if (!scmProjectMap[scmName]) {
    scmProjectMap[scmName] = [];
  }
  scmProjectMap[scmName]?.push(entry);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "neverOptionalChain",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveOptionalChain",
								Output: `
const scmProjectMap: Record<string, string[] | undefined> = {};
function run(scmName: string, entry: string) {
  if (!scmProjectMap[scmName]) {
    scmProjectMap[scmName] = [];
  }
  scmProjectMap[scmName].push(entry);
}
`,
							},
						},
					},
				},
			},
		},
	)
}

// Pattern C: optional property whose nullability comes from the object itself,
// not from an intrinsic property type. `items` on `RegionConfig` is
// optional — chain is necessary.
func TestAlign_PatternC_OptionalPropChain(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{
			{Code: `
type Region = 'I18N' | 'US-TTP';
interface RegionConfig {
  items?: { id: string }[];
}
declare const privateContext: Record<Region, RegionConfig>;
declare const regionSite: Region;
function run() {
  privateContext[regionSite].items?.push({ id: 'a' });
}
`},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

// Pattern D: optional call on index-accessed function. With a non-nullable
// value type (`Record<string, (m: M) => string>`), the `?.()` is unnecessary —
// latest reports.
func TestAlign_PatternD_OptionalCallOnIndex(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
type M = { a: number };
declare const bizScenarioStatusMap: Record<string, (map: M) => string>;
declare const bizScenario: string;
declare const map: M;
declare const statusMapHas: boolean;
function run(): string | undefined {
  if (!statusMapHas) return undefined;
  const res = bizScenarioStatusMap[bizScenario]?.(map);
  return res;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "neverOptionalChain",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "suggestRemoveOptionalChain",
								Output: `
type M = { a: number };
declare const bizScenarioStatusMap: Record<string, (map: M) => string>;
declare const bizScenario: string;
declare const map: M;
declare const statusMapHas: boolean;
function run(): string | undefined {
  if (!statusMapHas) return undefined;
  const res = bizScenarioStatusMap[bizScenario](map);
  return res;
}
`,
							},
						},
					},
				},
			},
		},
	)
}
