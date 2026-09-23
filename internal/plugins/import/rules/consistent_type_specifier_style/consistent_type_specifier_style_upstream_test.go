package consistent_type_specifier_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/consistent_type_specifier_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Pinned upstream tests and documentation, including every Flow case as an explained skip.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/consistent-type-specifier-style.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/consistent-type-specifier-style.md

func TestConsistentTypeSpecifierStyleUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule,
		[]rule_tester.ValidTestCase{

			// COMMON_TESTS (includes the regression for upstream issue #2753).
			{
				Code:    `import Foo from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type Foo from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import { Foo } from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import { Foo as Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import * as Foo from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import {} from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type {} from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type { Foo } from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type { Foo as Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type { Foo, Bar, Baz, Bam } from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import Foo from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import type Foo from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import { Foo } from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import { Foo as Bar } from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import * as Foo from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import {} from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import type {} from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import { type Foo } from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import { type Foo as Bar } from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import { type Foo, type Bar, Baz, Bam } from 'Foo';`,
				Options: []any{"prefer-inline"},
			},

			// TS_ONLY
			{
				Code: `import type * as Foo from 'Foo';`,
			},
		},
		[]rule_tester.InvalidTestCase{

			// COMMON_TESTS (includes the regression for upstream issue #2753).
			{
				Code:    `import { type Foo } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output:  []string{`import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    `import { type Foo as Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output:  []string{`import type {Foo as Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    `import { type Foo, type Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output:  []string{`import type {Foo, Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code:    `import { Foo, type Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output: []string{`import { Foo  } from 'Foo';
import type {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `import { type Foo, Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output: []string{`import {  Bar } from 'Foo';
import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `import Foo, { type Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output: []string{`import Foo from 'Foo';
import type {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `import Foo, { type Bar, Baz } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output: []string{`import Foo, {  Baz } from 'Foo';
import type {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `import { Component, type ComponentProps } from "package-1";
import {
  Component1,
  Component2,
  Component3,
  Component4,
  Component5,
} from "package-2";`,
				Options: []any{"prefer-top-level"},
				Output: []string{`import { Component  } from "package-1";
import type {ComponentProps} from "package-1";
import {
  Component1,
  Component2,
  Component3,
  Component4,
  Component5,
} from "package-2";`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 21, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    `import type { Foo } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Output:  []string{`import  { type Foo } from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    `import type { Foo, Bar, Baz } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Output:  []string{`import  { type Foo, type Bar, type Baz } from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42},
				},
			},
		},
	)
}

func TestConsistentTypeSpecifierStyleFlow(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule,
		[]rule_tester.ValidTestCase{

			// FLOW_ONLY
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof Foo from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof { Foo, Bar, Baz, Bam } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof Foo from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { typeof Foo } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { typeof Foo, typeof Bar, typeof Baz, typeof Bam } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { type Foo, type Bar, typeof Baz, typeof Bam } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
		},
		[]rule_tester.InvalidTestCase{

			// FLOW_ONLY
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { typeof Foo } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output:  []string{`import typeof {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { typeof Foo as Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output:  []string{`import typeof {Foo as Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { type Foo, typeof Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output: []string{`import type {Foo} from 'Foo';
import typeof {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type/typeof-only import instead of inline type/typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { typeof Foo, typeof Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output:  []string{`import typeof {Foo, Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { Foo, typeof Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output: []string{`import { Foo  } from 'Foo';
import typeof {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { typeof Foo, Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output: []string{`import {  Bar } from 'Foo';
import typeof {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import { Foo, type Bar, typeof Baz } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output: []string{`import { Foo   } from 'Foo';
import type {Bar} from 'Foo';
import typeof {Baz} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers."},
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import Foo, { typeof Bar } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output: []string{`import Foo from 'Foo';
import typeof {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import Foo, { typeof Bar, Baz } from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
				Output: []string{`import Foo, {  Baz } from 'Foo';
import typeof {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level typeof-only import instead of inline typeof specifiers."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof { Foo } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
				Output:  []string{`import  { typeof Foo } from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline typeof specifiers instead of a top-level typeof-only import."},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof { Foo, Bar, Baz } from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
				Output:  []string{`import  { typeof Foo, typeof Bar, typeof Baz } from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline typeof specifiers instead of a top-level typeof-only import."},
				},
			},
		},
	)
}

func TestConsistentTypeSpecifierStyleDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule,
		[]rule_tester.ValidTestCase{

			// Documentation example 1
			{
				Code:    `import type Foo from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type {Bar} from 'Bar';`,
				Options: []any{"prefer-top-level"},
			},
			{
				Code:    `import type * as Bam from 'Bam';`,
				Options: []any{"prefer-top-level"},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof Baz from 'Baz';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
			},

			// Documentation example 2
			{
				Code:    `import {type Foo} from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import {typeof Bar} from 'Bar';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},

			// Documentation example 4
			{
				Code:    `import type {Foo} from 'Foo';`,
				Options: []any{"prefer-top-level"},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import type Foo, {Bar} from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof {Foo} from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
			},

			// Documentation example 6
			{
				Code:    `import {type Foo} from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				Code:    `import Foo, {type Bar} from 'Foo';`,
				Options: []any{"prefer-inline"},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import {typeof Foo} from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
		},
		[]rule_tester.InvalidTestCase{

			// Documentation example 3
			{
				Code:    `import {type Foo} from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output:  []string{`import type {Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    `import Foo, {type Bar} from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Output: []string{`import Foo from 'Foo';
import type {Bar} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTopLevel", Message: "Prefer using a top-level type-only import instead of inline type specifiers.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import {typeof Foo} from 'Foo';`,
				Options: []any{"prefer-top-level"},
				Skip:    true,
			},

			// Documentation example 5
			{
				Code:    `import type {Foo} from 'Foo';`,
				Options: []any{"prefer-inline"},
				Output:  []string{`import  {type Foo} from 'Foo';`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferInline", Message: "Prefer using inline type specifiers instead of a top-level type-only import.", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import type Foo, {Bar} from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
			{
				// Skipped: tsgo does not support Flow import syntax.
				Code:    `import typeof {Foo} from 'Foo';`,
				Options: []any{"prefer-inline"},
				Skip:    true,
			},
		},
	)
}
