package switch_exhaustiveness_check

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSwitchExhaustivenessCheckPrefersReachableValueAlias(t *testing.T) {
	const bracketOutput = `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import { QuotedEnum as ValueQuoted } from './switch-exhaustiveness-check-quoted';
declare const value: ValueQuoted;
switch (value) {
  case ValueQuoted.z:
    break;
  case ValueQuoted['x-y']: { throw new Error('Not implemented yet: ValueQuoted[\'x-y\'] case') }
}
`
	const dotOutput = `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import { QuotedEnum as ValueQuoted } from './switch-exhaustiveness-check-quoted';
declare const value: ValueQuoted;
switch (value) {
  case ValueQuoted['x-y']:
    break;
  case ValueQuoted.z: { throw new Error('Not implemented yet: ValueQuoted.z case') }
}
`
	const symbolOutput = `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import { uniqueA as ValueA, uniqueB as ValueB } from './switch-exhaustiveness-check-quoted';
declare const value: typeof ValueA | typeof ValueB;
switch (value) {
  case ValueA:
    break;
  case ValueB: { throw new Error('Not implemented yet: ValueB case') }
}
`
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				FileName: "mixed-import-alias-order.ts",
				Code: `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import { QuotedEnum as ValueQuoted } from './switch-exhaustiveness-check-quoted';
declare const value: ValueQuoted;
switch (value) {
  case ValueQuoted.z:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: bracketOutput},
						},
					},
				},
			},
			{
				FileName: "mixed-import-dot-alias-order.ts",
				Code: `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import { QuotedEnum as ValueQuoted } from './switch-exhaustiveness-check-quoted';
declare const value: ValueQuoted;
switch (value) {
  case ValueQuoted['x-y']:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: dotOutput},
						},
					},
				},
			},
			{
				FileName: "mixed-import-symbol-alias-order.ts",
				Code: `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import { uniqueA as ValueA, uniqueB as ValueB } from './switch-exhaustiveness-check-quoted';
declare const value: typeof ValueA | typeof ValueB;
switch (value) {
  case ValueA:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: symbolOutput},
						},
					},
				},
			},
		},
	)
}

func TestSwitchExhaustivenessCheckUsesTargetAwareEnumAccess(t *testing.T) {
	const output = `
enum E {
  Ϳ = 1,
  x = 2,
}
declare const value: E;
switch (value) {
  case E.x:
    break;
  case E.Ϳ: { throw new Error('Not implemented yet: E.Ϳ case') }
}
`
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				Code: `
enum E {
  Ϳ = 1,
  x = 2,
}
declare const value: E;
switch (value) {
  case E.x:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: output},
						},
					},
				},
			},
		},
	)
}

func TestSwitchExhaustivenessCheckUsesNamespaceValueStarExports(t *testing.T) {
	const enumOutput = `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-value-star';
declare const value: barrel.QuotedEnum;
switch (value) {
  case barrel.QuotedEnum.z:
    break;
  case barrel.QuotedEnum['x-y']: { throw new Error('Not implemented yet: barrel.QuotedEnum[\'x-y\'] case') }
}
`
	const symbolOutput = `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-value-star';
declare const value: typeof barrel.uniqueA | typeof barrel.uniqueB;
switch (value) {
  case barrel.uniqueA:
    break;
  case barrel.uniqueB: { throw new Error('Not implemented yet: barrel.uniqueB case') }
}
`
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				FileName: "namespace-value-star-enum.ts",
				Code: `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-value-star';
declare const value: barrel.QuotedEnum;
switch (value) {
  case barrel.QuotedEnum.z:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: enumOutput},
						},
					},
				},
			},
			{
				FileName: "namespace-value-star-symbol.ts",
				Code: `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-value-star';
declare const value: typeof barrel.uniqueA | typeof barrel.uniqueB;
switch (value) {
  case barrel.uniqueA:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: symbolOutput},
						},
					},
				},
			},
		},
	)
}

func TestSwitchExhaustivenessCheckRejectsNamespaceTypeOnlyStarExports(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				FileName: "namespace-type-only-star-enum.ts",
				Code: `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-type-only-star';
declare const value: barrel.QuotedEnum;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
			{
				FileName: "namespace-type-only-star-symbol.ts",
				Code: `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-type-only-star';
declare const value: typeof barrel.uniqueA | typeof barrel.uniqueB;
switch (value) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchIsNotExhaustive"},
				},
			},
		},
	)
}

func TestSwitchExhaustivenessCheckUsesRenamedNamespaceExports(t *testing.T) {
	const enumOutput = `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-renamed';
declare const value: barrel.RenamedEnum;
switch (value) {
  case barrel.RenamedEnum.z:
    break;
  case barrel.RenamedEnum['x-y']: { throw new Error('Not implemented yet: barrel.RenamedEnum[\'x-y\'] case') }
}
`
	const symbolOutput = `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-renamed';
declare const value: typeof barrel.renamedA | typeof barrel.renamedB;
switch (value) {
  case barrel.renamedA:
    break;
  case barrel.renamedB: { throw new Error('Not implemented yet: barrel.renamedB case') }
}
`
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		nil,
		[]rule_tester.InvalidTestCase{
			{
				FileName: "namespace-renamed-enum.ts",
				Code: `
import type { QuotedEnum as TypeQuoted } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-renamed';
declare const value: barrel.RenamedEnum;
switch (value) {
  case barrel.RenamedEnum.z:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: enumOutput},
						},
					},
				},
			},
			{
				FileName: "namespace-renamed-symbol.ts",
				Code: `
import type { uniqueA as TypeA, uniqueB as TypeB } from './switch-exhaustiveness-check-quoted';
import * as barrel from './switch-exhaustiveness-check-renamed';
declare const value: typeof barrel.renamedA | typeof barrel.renamedB;
switch (value) {
  case barrel.renamedA:
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addMissingCases", Output: symbolOutput},
						},
					},
				},
			},
		},
	)
}
