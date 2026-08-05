package utils_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// describeParsedExpect canonicalizes a parse result so the probe below can
// assert the full contract — entry kind, head presence, modifiers, matcher,
// reason and the static-call classification — through exact message matches.
func describeParsedExpect(parsed *rstestUtils.ParsedRstestExpectCall) string {
	reason := string(parsed.Reason)
	if reason == "" {
		reason = "none"
	}
	return fmt.Sprintf(
		"entry=%s head=%t modifiers=[%s] matcher=%s reason=%s static=%t",
		parsed.Entry,
		parsed.Head != nil,
		strings.Join(parsed.Modifiers, " "),
		parsed.Matcher,
		reason,
		rstestUtils.IsStaticRstestExpectCall(parsed),
	)
}

// expectParseProbe reports every parsed expect call with its canonical
// description, so invalid test cases assert exact parse results and valid test
// cases assert that ParseRstestExpectCall returned nil.
var expectParseProbe = rule.Rule{
	Name:             "rstest/expect-parse-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		callbacks := rstestUtils.CollectRstestTestCallbacks(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestExpectCall(node, ctx, callbacks)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("parsedExpect", describeParsedExpect(parsed)))
			},
		}
	},
}

func parsedExpectError(message string) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{MessageId: "parsedExpect", Message: message}}
}

func TestParseRstestExpectCallEntries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(x).toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect.soft(x).toBe(1);`,
				Errors: parsedExpectError("entry=soft head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect['soft'](x).toBe(1);`,
				Errors: parsedExpectError("entry=soft head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect.poll(fn).toBe(1);`,
				Errors: parsedExpectError("entry=poll head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect.element(l).toBeVisible();`,
				Errors: parsedExpectError("entry=element head=true modifiers=[] matcher=toBeVisible reason=none static=false"),
			},
			{
				Code:   `expect.unreachable();`,
				Errors: parsedExpectError("entry=unreachable head=false modifiers=[] matcher=unreachable reason=none static=true"),
			},
			// assertion-count declarations are the one static form
			// no-standalone-expect still reports, hence static=false.
			{
				Code:   `expect.assertions(1);`,
				Errors: parsedExpectError("entry=assertions head=false modifiers=[] matcher=assertions reason=none static=false"),
			},
			{
				Code:   `expect.hasAssertions();`,
				Errors: parsedExpectError("entry=assertions head=false modifiers=[] matcher=hasAssertions reason=none static=false"),
			},
			{
				Code:   `expect.stringContaining("a");`,
				Errors: parsedExpectError("entry=asymmetric head=false modifiers=[] matcher=stringContaining reason=none static=true"),
			},
			{
				Code:   `expect.anything();`,
				Errors: parsedExpectError("entry=asymmetric head=false modifiers=[] matcher=anything reason=none static=true"),
			},
			// ExpectStatic.not.<asymmetric> is the static negation, not the
			// Assertion.not modifier; upstream parses it as modifier + matcher
			// and its two members keep it reportable by no-standalone-expect.
			{
				Code:   `expect.not.stringContaining("a");`,
				Errors: parsedExpectError("entry=asymmetric head=false modifiers=[not] matcher=stringContaining reason=none static=false"),
			},
			{
				Code:   `expect.addEqualityTesters([]);`,
				Errors: parsedExpectError("entry=config head=false modifiers=[] matcher=addEqualityTesters reason=none static=true"),
			},
			{
				Code:   `expect.extend({});`,
				Errors: parsedExpectError("entry=config head=false modifiers=[] matcher=extend reason=none static=true"),
			},
			// Unknown static members may be custom asymmetric matchers
			// registered through expect.extend.
			{
				Code:   `expect.toBeFoo(1);`,
				Errors: parsedExpectError("entry=asymmetric head=false modifiers=[] matcher=toBeFoo reason=none static=true"),
			},
			// A bare modifier chain parses like a plain chain with no head,
			// mirroring vitest's null expectCall.
			{
				Code:   `expect.resolves.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=false modifiers=[resolves] matcher=toBe reason=none static=false"),
			},
		},
	)
}

func TestParseRstestExpectCallModifiersAndMatcher(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(x).not.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[not] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect(x).resolves.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[resolves] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect(x).rejects.toThrow();`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[rejects] matcher=toThrow reason=none static=false"),
			},
			{
				Code:   `expect(x).resolves.not.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[resolves not] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect(x).rejects.not.toThrow();`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[rejects not] matcher=toThrow reason=none static=false"),
			},
			{
				Code:   `expect(x).not.not.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			{
				Code:   `expect(x).not.resolves.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			{
				Code:   `expect(x).resolves.rejects.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			{
				Code:   `expect(x).foo.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			// The matcher is the first invoked member, never simply the last
			// member of the chain.
			{
				Code:   `expect(x).toBe(1).toString();`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect(x);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=matcher-not-found static=false"),
			},
			{
				Code:   `expect(x).toBe;`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=matcher-not-called static=false"),
			},
			{
				Code:   `expect(x).resolves;`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=matcher-not-called static=false"),
			},
			{
				Code:   `(expect(x)).not.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[not] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect(x)['toBe'](1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			// A computed member with an identifier key cannot be resolved
			// statically.
			{
				Code:   `expect(x)[matcherName](1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=matcher-not-found static=false"),
			},
			{
				Code:   `expect(f).toHaveBeenCalledBefore(g);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher=toHaveBeenCalledBefore reason=none static=false"),
			},
			{
				Code:   `expect(x).matchSnapshot();`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher=matchSnapshot reason=none static=false"),
			},
			// A factory call without a matcher is a broken chain, but still a
			// single static member syntactically, so no-standalone-expect skips
			// it — matching upstream, where the failed parse hides the call.
			{
				Code:   `expect.soft(x);`,
				Errors: parsedExpectError("entry=soft head=true modifiers=[] matcher= reason=matcher-not-found static=true"),
			},
			// A factory member that is never invoked cannot start a chain.
			{
				Code:   `expect.soft.foo(1);`,
				Errors: parsedExpectError("entry=soft head=false modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
		},
	)
}

func TestParseRstestExpectCallSources(t *testing.T) {
	chain := "entry=expect head=true modifiers=[] matcher=toBe reason=none static=false"
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `import { expect } from '@rstest/core'; expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import { expect as check } from '@rstest/core'; check(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `const { expect } = require('@rstest/core'); expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import * as rstest from '@rstest/core'; rstest.expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import * as rstest from '@rstest/core'; rstest['expect'](1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `const rstest = require('@rstest/core'); rstest.expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import.meta.rstest.expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `const { expect } = import.meta.rstest; expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `const api = import.meta.rstest; api.expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `test("t", (ctx) => { ctx.expect(1).toBe(1); });`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `test("t", ({ expect }) => { expect(1).toBe(1); });`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `test("t", ({ expect: check }) => { check(1).toBe(1); });`,
				Errors: parsedExpectError(chain),
			},
			{
				Code:   `import { expect } from '@rstest/playwright'; expect(1).toBe(1);`,
				Errors: parsedExpectError(chain),
			},
			// The entry kinds resolve through receiver roots too.
			{
				Code:   `import * as rstest from '@rstest/core'; rstest.expect.soft(1).toBe(1);`,
				Errors: parsedExpectError("entry=soft head=true modifiers=[] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `import.meta.rstest.expect.assertions(1);`,
				Errors: parsedExpectError("entry=assertions head=false modifiers=[] matcher=assertions reason=none static=false"),
			},
		},
	)
}

func TestParseRstestExpectCallIgnoresForeignExpect(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectParseProbe,
		[]rule_tester.ValidTestCase{
			{Code: `import { expect } from 'vitest'; expect(1).toBe(1);`},
			{Code: `import { expect } from '@jest/globals'; expect(1).toBe(1);`},
			{Code: `import { test, expect } from '@playwright/test'; expect(1).toBe(1);`},
			{Code: `const expect = createAssertionLibrary(); expect(1).toBe(1);`},
			{Code: `custom.expect(1).toBe(1);`},
			{Code: `notExpect(1).toBe(1);`},
			{Code: `test("t", () => {});`},
			{Code: `describe("suite", () => {});`},
			{Code: `Math.max(1, 2);`},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

// expectAwaitProbe reports ShouldRstestExpectBeAwaited with a fixed
// asyncMatchers option of ["toResolve"].
var expectAwaitProbe = rule.Rule{
	Name:             "rstest/expect-await-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		callbacks := rstestUtils.CollectRstestTestCallbacks(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := rstestUtils.ParseRstestExpectCall(node, ctx, callbacks)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("awaited", fmt.Sprintf(
					"await=%t",
					rstestUtils.ShouldRstestExpectBeAwaited(parsed, []string{"toResolve"}),
				)))
			},
		}
	},
}

func TestShouldRstestExpectBeAwaited(t *testing.T) {
	awaited := []rule_tester.InvalidTestCaseError{{MessageId: "awaited", Message: "await=true"}}
	notAwaited := []rule_tester.InvalidTestCaseError{{MessageId: "awaited", Message: "await=false"}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectAwaitProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: `expect(x).resolves.toBe(1);`, Errors: awaited},
			{Code: `expect(x).rejects.not.toThrow();`, Errors: awaited},
			{Code: `expect(x).toResolve();`, Errors: awaited},
			{Code: `expect(x).not.toBe(1);`, Errors: notAwaited},
			{Code: `expect(x).toBe(1);`, Errors: notAwaited},
		},
	)
}

func TestRstestMatcherTables(t *testing.T) {
	if len(rstestUtils.RSTEST_MATCHER_ALIASES) != 11 {
		t.Errorf("expected 11 matcher aliases, got %d", len(rstestUtils.RSTEST_MATCHER_ALIASES))
	}
	if got := rstestUtils.RSTEST_MATCHER_ALIASES["toBeCalled"]; got != "toHaveBeenCalled" {
		t.Errorf("toBeCalled should map to toHaveBeenCalled, got %q", got)
	}
	for alias, canonical := range rstestUtils.RSTEST_MATCHER_ALIASES {
		if _, isAlias := rstestUtils.RSTEST_MATCHER_ALIASES[canonical]; isAlias {
			t.Errorf("canonical name %q (for alias %q) must not itself be an alias", canonical, alias)
		}
	}

	if len(rstestUtils.RSTEST_INLINE_SNAPSHOT_MATCHERS) != 2 {
		t.Errorf("expected 2 inline snapshot matchers, got %d", len(rstestUtils.RSTEST_INLINE_SNAPSHOT_MATCHERS))
	}
	for name := range rstestUtils.RSTEST_INLINE_SNAPSHOT_MATCHERS {
		if !rstestUtils.RSTEST_SNAPSHOT_MATCHERS[name] {
			t.Errorf("inline snapshot matcher %q missing from RSTEST_SNAPSHOT_MATCHERS", name)
		}
	}
	if len(rstestUtils.RSTEST_SNAPSHOT_MATCHERS) != 6 {
		t.Errorf("expected 6 snapshot matchers, got %d", len(rstestUtils.RSTEST_SNAPSHOT_MATCHERS))
	}

	if len(rstestUtils.RSTEST_BROWSER_ELEMENT_MATCHERS) != 21 {
		t.Errorf("expected 21 browser element matchers, got %d", len(rstestUtils.RSTEST_BROWSER_ELEMENT_MATCHERS))
	}
	if len(rstestUtils.RSTEST_ASYMMETRIC_MATCHERS) != 10 {
		t.Errorf("expected 10 asymmetric matchers, got %d", len(rstestUtils.RSTEST_ASYMMETRIC_MATCHERS))
	}

	if len(rstestUtils.RSTEST_POLL_EXCLUDED_MEMBERS) != 11 {
		t.Errorf("expected 11 poll-excluded members, got %d", len(rstestUtils.RSTEST_POLL_EXCLUDED_MEMBERS))
	}
	for _, name := range []string{"resolves", "rejects", "toMatchInlineSnapshot"} {
		if !rstestUtils.RSTEST_POLL_EXCLUDED_MEMBERS[name] {
			t.Errorf("poll-excluded members must contain %q", name)
		}
	}
}
