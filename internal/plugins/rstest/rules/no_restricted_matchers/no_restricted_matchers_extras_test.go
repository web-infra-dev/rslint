// TestNoRestrictedMatchersExtras covers Rstest sources, assertion factories,
// first-matcher semantics, computed members, ordering, and tsgo edge shapes.
// The upstream Vitest suite lives in no_restricted_matchers_upstream_test.go.
package no_restricted_matchers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedMatchersExtras(t *testing.T) {
	if NoRestrictedMatchersRule.RequiresTypeInfo {
		t.Fatal("rstest/no-restricted-matchers must run in source-only programs")
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedMatchersRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(value).toBe(1)`},
			{Code: `expect(value).toBe(1)`, Options: []any{map[string]any{}}},

			// ---- First matcher boundary ----
			{Code: `expect(value).to.be.a("string").and.contain("x")`, Options: restrictions("contain", nil)},
			{Code: `expect(value).equal(other).and.toBe(1)`, Options: restrictions("toBe", nil)},

			// ---- Foreign and shadowed sources ----
			{Code: `import { expect } from 'vitest'; expect(value).toBe(1)`, Options: restrictions("toBe", nil)},
			{Code: `import { expect } from '@jest/globals'; expect(value).toBe(1)`, Options: restrictions("toBe", nil)},
			{Code: `import { expect } from '@playwright/test'; expect(value).toBe(1)`, Options: restrictions("toBe", nil)},
			{Code: `const expect = createAssertionLibrary(); expect(value).toBe(1)`, Options: restrictions("toBe", nil)},
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(value).toBe(1) }`, Options: restrictions("toBe", nil)},

			// ---- Dynamic computed members cannot form a static chain ----
			{Code: `const matcher = "toBe"; expect(value)[matcher](1)`, Options: restrictions("toBe", nil)},
			{Code: `expect(value)[getMatcher()](1)`, Options: restrictions("toBe", nil)},

			// ---- Dimension 4: wrappers are parser boundaries ----
			{Code: `expect(value)!.toBe(1)`, Options: restrictions("toBe", nil)},
			{Code: `(expect(value) as any).toBe(1)`, Options: restrictions("toBe", nil)},
			{Code: `(expect(value) satisfies Assertion).toBe(1)`, Options: restrictions("toBe", nil)},

			// N/A: declaration/container forms, function kinds, class members,
			// spread/rest, destructuring and overload signatures do not change
			// this rule's call-chain listener.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user: soft, polling, and browser element assertions ----
			invalid(`expect.soft(value).toBe(1)`, restrictions("soft.toBe", nil), "restrictedChain", "", 1, 8, 1, 24),
			invalid(`expect.poll(() => value).toBe(1)`, restrictions("poll.toBe", nil), "restrictedChain", "", 1, 8, 1, 30),
			invalid(`expect.element(locator).toBeVisible()`, restrictions("element.toBeVisible", nil), "restrictedChain", "", 1, 8, 1, 36),

			// ---- Real-user: static asymmetric matcher factories ----
			invalid(`expect.stringContaining("x")`, restrictions("stringContaining", nil), "restrictedChain", "", 1, 8, 1, 24),
			invalid(`expect.not.arrayContaining([])`, restrictions("not", nil), "restrictedChain", "", 1, 8, 1, 27),

			// ---- Chai property, call, and multi-matcher chains ----
			invalid(`expect(value).to.be.ok`, restrictions("to.be.ok", nil), "restrictedChain", "", 1, 15, 1, 23),
			invalid(`expect(value).to.have.property("name")`, restrictions("to.have.property", nil), "restrictedChain", "", 1, 15, 1, 31),
			invalid(`expect(value).to.be.a("string").and.contain("x")`, restrictions("to.be.a", nil), "restrictedChain", "", 1, 15, 1, 22),

			// ---- Complete Rstest source matrix ----
			invalid(`import { expect } from '@rstest/core'; expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 54, 1, 58),
			invalid(`import { expect as check } from '@rstest/core'; check(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 62, 1, 66),
			invalid(`const { expect } = require('@rstest/core'); expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 59, 1, 63),
			invalid(`const { expect: check } = require('@rstest/core'); check(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 65, 1, 69),
			invalid(`import * as core from '@rstest/core'; core.expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 58, 1, 62),
			invalid(`const core = require('@rstest/core'); core.expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 58, 1, 62),
			invalid(`import.meta.rstest.expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 34, 1, 38),
			invalid(`const { expect } = import.meta.rstest; expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 54, 1, 58),
			invalid(`const api = import.meta.rstest; api.expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 51, 1, 55),
			invalid(`import { expect } from '@rstest/playwright'; expect(value).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 60, 1, 64),
			invalid(`test("works", (ctx) => { ctx.expect(value).toBe(1) })`, restrictions("toBe", nil), "restrictedChain", "", 1, 44, 1, 48),
			invalid(`test("works", ({ expect }) => { expect(value).toBe(1) })`, restrictions("toBe", nil), "restrictedChain", "", 1, 47, 1, 51),
			invalid(`test("works", ({ expect: check }) => { check(value).toBe(1) })`, restrictions("toBe", nil), "restrictedChain", "", 1, 53, 1, 57),

			// ---- Deterministic overlap selection and one-report contract ----
			{
				Code: `expect(value).resolves.not.toBe(1)`,
				Options: []any{map[string]any{
					"resolves":          "broad",
					"resolves.not":      "specific",
					"resolves.not.toBe": "most specific",
				}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "restrictedChainWithMessage",
					Message:   "most specific",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 32,
				}},
			},

			// ---- Range trims leading trivia and stops at the first matcher ----
			invalid(`expect(value). /* comment */ to.be.a("string").and.contain("x")`, restrictions("to.be.a", nil), "restrictedChain", "", 1, 30, 1, 37),
			invalid("expect(value).\n  not.toBe(1)", restrictions("not.toBe", nil), "restrictedChain", "", 2, 3, 2, 11),
			invalid(`(expect(value)).toBe(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 17, 1, 21),

			// ---- Dimension 4: optional and computed static access ----
			invalid(`expect?.(value)?.toBe?.(1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 18, 1, 22),
			invalid(`expect(value)["toBe"](1)`, restrictions("toBe", nil), "restrictedChain", "", 1, 15, 1, 21),
			invalid("expect(value)[`toBe`](1)", restrictions("toBe", nil), "restrictedChain", "", 1, 15, 1, 21),
		},
	)
}
