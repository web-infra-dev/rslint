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

func describeParsedExpectChain(parsed *rstestUtils.ParsedRstestExpectCall) string {
	reason := string(parsed.Reason)
	if reason == "" {
		reason = "none"
	}
	matchers := make([]string, len(parsed.Matchers))
	for i, matcher := range parsed.Matchers {
		matchers[i] = fmt.Sprintf("%s:%s", matcher.Name, matcher.Kind)
	}
	return fmt.Sprintf(
		"entry=%s head=%t expression=%s members=[%s] modifiers=[%s] matcher=%s matchers=[%s] reason=%s static=%t",
		parsed.Entry,
		parsed.Head != nil,
		expectExpressionKind(parsed.Expression),
		strings.Join(parsed.Members, " "),
		strings.Join(parsed.Modifiers, " "),
		parsed.Matcher,
		strings.Join(matchers, " "),
		reason,
		rstestUtils.IsStaticRstestExpectCall(parsed),
	)
}

func expectExpressionKind(node *ast.Node) string {
	if node == nil {
		return "nil"
	}
	switch node.Kind {
	case ast.KindCallExpression:
		return "call"
	case ast.KindPropertyAccessExpression:
		return "property"
	case ast.KindElementAccessExpression:
		return "element"
	default:
		return "other"
	}
}

// expectParseProbe reports every parsed expect call with its canonical
// description, so invalid test cases assert exact parse results and valid test
// cases assert that ParseRstestExpectCall returned nil.
var expectParseProbe = rule.Rule{
	Name:             "rstest/expect-parse-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("parsedExpect", describeParsedExpect(parsed)))
			},
		}
	},
}

var expectChainParseProbe = rule.Rule{
	Name:             "rstest/expect-chain-parse-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return
				}
				ctx.ReportNode(node, probeMessage("parsedExpect", describeParsedExpectChain(parsed)))
			},
		}
	},
}

func parsedExpectError(message string) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{MessageId: "parsedExpect", Message: message}}
}

// expectEntryInvariantProbe reports only when Entry and Head disagree, so the
// corpus below passes as valid test cases and any violation surfaces as an
// unexpected error rather than as a diff a reader has to spot.
var expectEntryInvariantProbe = rule.Rule{
	Name: "rstest/probe-expect-entry-invariant",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return
				}
				isStatic := parsed.Entry == rstestUtils.RstestExpectEntryStatic
				if isStatic == (parsed.Head == nil) {
					return
				}
				ctx.ReportNode(node, probeMessage("parsedExpect", fmt.Sprintf(
					"invariant broken: entry=%s head=%t", parsed.Entry, parsed.Head != nil,
				)))
			},
		}
	},
}

// TestParseRstestExpectCallEntryInvariant pins the contract the entry kinds
// exist to carry: Entry names a factory that actually ran, so it is static
// exactly when there is no head. Rules rely on either field alone to decide
// whether they are looking at a real assertion.
func TestParseRstestExpectCallEntryInvariant(t *testing.T) {
	shapes := []string{
		`expect(x).toBe(1);`,
		`expect(x).not.toBe(1);`,
		`expect(x).resolves.not.toBe(1);`,
		`expect(x).to.be.ok;`,
		`expect(x);`,
		`expect(x).not;`,
		`expect.soft(x).toBe(1);`,
		`expect.poll(fn).toBe(1);`,
		`expect.element(l).toBeVisible();`,
		`expect.soft.foo(1);`,
		`expect.assertions(1);`,
		`expect.hasAssertions();`,
		`expect.unreachable();`,
		`expect.getState();`,
		`expect.extend({});`,
		`expect.stringContaining("a");`,
		`expect.not.stringContaining("a");`,
		`expect.not.toBeDividedBy(5);`,
		`expect.toBeFoo(1);`,
		`expect.resolves.toBe(1);`,
		`expect.not.resolves.toBe(1);`,
		`expect.not.not.toBe(1);`,
		`expect.not;`,
	}
	cases := make([]rule_tester.ValidTestCase, len(shapes))
	for i, shape := range shapes {
		cases[i] = rule_tester.ValidTestCase{Code: shape}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectEntryInvariantProbe,
		cases,
		[]rule_tester.InvalidTestCase{},
	)
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
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=unreachable reason=none static=true"),
			},
			// assertion-count declarations are the one static form
			// no-standalone-expect still reports, hence static=false.
			{
				Code:   `expect.assertions(1);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=assertions reason=none static=false"),
			},
			{
				Code:   `expect.hasAssertions();`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=hasAssertions reason=none static=false"),
			},
			{
				Code:   `expect.stringContaining("a");`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=stringContaining reason=none static=true"),
			},
			{
				Code:   `expect.anything();`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=anything reason=none static=true"),
			},
			{
				Code:   `expect.any(String);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=any reason=none static=true"),
			},
			// ExpectStatic.not.<asymmetric> is the static negation, not the
			// Assertion.not modifier; upstream parses it as modifier + matcher
			// and its two members keep it reportable by no-standalone-expect.
			{
				Code:   `expect.not.stringContaining("a");`,
				Errors: parsedExpectError("entry=static head=false modifiers=[not] matcher=stringContaining reason=none static=false"),
			},
			// A negated custom matcher parses identically to a negated
			// built-in one. The parser does not consult
			// RSTEST_ASYMMETRIC_MATCHERS, so a name it has never seen is not
			// classified differently from one it has.
			{
				Code:   `expect.not.toBeDividedBy(5);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[not] matcher=toBeDividedBy reason=none static=false"),
			},
			{
				Code:   `expect.addEqualityTesters([]);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=addEqualityTesters reason=none static=true"),
			},
			{
				Code:   `expect.extend({});`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=extend reason=none static=true"),
			},
			// Unknown static members may be custom asymmetric matchers
			// registered through expect.extend.
			{
				Code:   `expect.toBeFoo(1);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=toBeFoo reason=none static=true"),
			},
			// A bare modifier chain is static too: no assertion factory ran,
			// mirroring vitest's null expectCall. Entry never claims a factory
			// that was not invoked.
			{
				Code:   `expect.resolves.toBe(1);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[resolves] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect.not.resolves.toBe(1);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[not resolves] matcher=toBe reason=none static=false"),
			},
			// A factory member that is never called does not make the chain a
			// factory chain either.
			{
				Code:   `expect.soft.toBe(1);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher= reason=modifier-unknown static=false"),
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
				Code:   `expect(x).not.resolves.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[not resolves] matcher=toBe reason=none static=false"),
			},
			{
				Code:   `expect(x).not.rejects.toThrow();`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[not rejects] matcher=toThrow reason=none static=false"),
			},
			{
				Code:   `expect(x).not.not.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			{
				Code:   `expect(x).resolves.rejects.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			{
				Code:   `expect(x).resolves.resolves.toBe(1);`,
				Errors: parsedExpectError("entry=expect head=true modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
			{
				Code:   `expect(x).rejects.rejects.toThrow();`,
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
			// A factory member that is never invoked cannot start a chain, so
			// the entry stays static rather than claiming a factory that never
			// ran.
			{
				Code:   `expect.soft.foo(1);`,
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher= reason=modifier-unknown static=false"),
			},
		},
	)
}

func TestParseRstestExpectCallChaiMethods(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectChainParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(x).to.equal(y);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to equal] modifiers=[] matcher=equal matchers=[equal:call] reason=none static=false",
				),
			},
			{
				Code: `expect(x).to.have.property("name");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to have property] modifiers=[] matcher=property matchers=[property:call] reason=none static=false",
				),
			},
			{
				Code: `expect(x).to.deep.equal(y);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to deep equal] modifiers=[] matcher=equal matchers=[equal:call] reason=none static=false",
				),
			},
			{
				Code: `expect(x).to.have.any.keys("a", "b");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to have any keys] modifiers=[] matcher=keys matchers=[keys:call] reason=none static=false",
				),
			},
			{
				Code: `expect(x).to.include("x");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to include] modifiers=[] matcher=include matchers=[include:call] reason=none static=false",
				),
			},
			{
				Code: `expect(xs).to.include.members([x]);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to include members] modifiers=[] matcher=members matchers=[members:call] reason=none static=false",
				),
			},
			{
				Code: `expect(xs).to.have.lengthOf(3);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to have lengthOf] modifiers=[] matcher=lengthOf matchers=[lengthOf:call] reason=none static=false",
				),
			},
			{
				Code: `expect(xs).to.have.lengthOf.above(2);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to have lengthOf above] modifiers=[] matcher=above matchers=[above:call] reason=none static=false",
				),
			},
			{
				Code: `expect(Foo).itself.to.respondTo("bar");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[itself to respondTo] modifiers=[] matcher=respondTo matchers=[respondTo:call] reason=none static=false",
				),
			},
			{
				Code: `expect("hello").to.be.a("string").and.contain("hell");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to be a and contain] modifiers=[] matcher=a matchers=[a:call contain:call] reason=none static=false",
				),
			},
			{
				Code: `expect("hello").to.be.a("string").that.does.not.contain("world");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to be a that does not contain] modifiers=[] matcher=a matchers=[a:call contain:call] reason=none static=false",
				),
			},
		},
	)
}

func TestParseRstestExpectCallChaiProperties(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectChainParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(value).to.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to be ok] modifiers=[] matcher=ok matchers=[ok:property] reason=none static=false",
				),
			},
			{
				Code: `expect(value).to.not.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to not be ok] modifiers=[not] matcher=ok matchers=[ok:property] reason=none static=false",
				),
			},
			{
				Code: `expect(value).not.to.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[not to be ok] modifiers=[not] matcher=ok matchers=[ok:property] reason=none static=false",
				),
			},
			{
				Code: `expect(value).to.be.true;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to be true] modifiers=[] matcher=true matchers=[true:property] reason=none static=false",
				),
			},
			{
				Code: `expect(value).to.exist;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to exist] modifiers=[] matcher=exist matchers=[exist:property] reason=none static=false",
				),
			},
			{
				Code: `expect(spy).to.have.been.called;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to have been called] modifiers=[] matcher=called matchers=[called:property] reason=none static=false",
				),
			},
			{
				Code: `expect(spy).to.have.been.calledOnce;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to have been calledOnce] modifiers=[] matcher=calledOnce matchers=[calledOnce:property] reason=none static=false",
				),
			},
			{
				Code: `expect(value)["to"]["be"]["ok"];`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=element members=[to be ok] modifiers=[] matcher=ok matchers=[ok:property] reason=none static=false",
				),
			},
			{
				Code: `expect(1).to.equal(1).and.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[to equal and be ok] modifiers=[] matcher=equal matchers=[equal:call ok:property] reason=none static=false",
				),
			},
		},
	)
}

func TestParseRstestExpectCallChaiPromiseModifiers(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectChainParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(p).resolves.to.equal(x);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[resolves to equal] modifiers=[resolves] matcher=equal matchers=[equal:call] reason=none static=false",
				),
			},
			{
				Code: `expect(p).not.resolves.to.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[not resolves to be ok] modifiers=[not resolves] matcher=ok matchers=[ok:property] reason=none static=false",
				),
			},
			{
				Code: `expect(p).resolves.to.not.equal(x);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[resolves to not equal] modifiers=[resolves not] matcher=equal matchers=[equal:call] reason=none static=false",
				),
			},
			{
				Code: `expect(p).to.resolves.equal(x);`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[to resolves equal] modifiers=[resolves] matcher=equal matchers=[equal:call] reason=none static=false",
				),
			},
			{
				Code: `expect(p).rejects.to.have.property("message");`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=call members=[rejects to have property] modifiers=[rejects] matcher=property matchers=[property:call] reason=none static=false",
				),
			},
		},
	)
}

func TestParseRstestExpectCallChaiInvalidChains(t *testing.T) {
	invalid := "entry=expect head=true expression=call members=[%s] modifiers=[] matcher= matchers=[] reason=modifier-unknown static=false"
	propertyInvalid := "entry=expect head=true expression=property members=[%s] modifiers=[] matcher= matchers=[] reason=modifier-unknown static=false"
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectChainParseProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(x).to.be("string").that.does.not.contain("world");`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "to be that does not contain")),
			},
			{
				Code:   `expect(x).to.have.foo.equal(y);`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "to have foo equal")),
			},
			{
				Code:   `expect(x).to.be.a("string").that.foo.contain("x");`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "to be a that foo contain")),
			},
			{
				Code:   `expect(x).not().to.equal(y);`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "not to equal")),
			},
			{
				Code:   `expect(x).to();`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "to")),
			},
			{
				Code:   `expect(x).any(y);`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "any")),
			},
			{
				Code:   `expect(x).to.be.ok();`,
				Errors: parsedExpectError(fmt.Sprintf(invalid, "to be ok")),
			},
			{
				Code:   `expect(x).not.not.to.be.ok;`,
				Errors: parsedExpectError(fmt.Sprintf(propertyInvalid, "not not to be ok")),
			},
			{
				Code:   `expect(x).resolves.rejects.to.be.ok;`,
				Errors: parsedExpectError(fmt.Sprintf(propertyInvalid, "resolves rejects to be ok")),
			},
			{
				Code: `expect(x).unknownMatcher;`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=property members=[unknownMatcher] modifiers=[] matcher= matchers=[] reason=matcher-not-called static=false",
				),
			},
			{
				Code:   `expect(x).foo.bar;`,
				Errors: parsedExpectError(fmt.Sprintf(propertyInvalid, "foo bar")),
			},
			{
				Code: `expect(x)[matcherName];`,
				Errors: parsedExpectError(
					"entry=expect head=true expression=element members=[matcherName] modifiers=[] matcher= matchers=[] reason=matcher-not-found static=false",
				),
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
				Errors: parsedExpectError("entry=static head=false modifiers=[] matcher=assertions reason=none static=false"),
			},
			{
				Code: `import { expect as check } from '@rstest/core'; check(1).to.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true modifiers=[] matcher=ok reason=none static=false",
				),
			},
			{
				Code: `import * as rstest from '@rstest/core'; rstest.expect(1).to.equal(1);`,
				Errors: parsedExpectError(
					"entry=expect head=true modifiers=[] matcher=equal reason=none static=false",
				),
			},
			{
				Code: `import.meta.rstest.expect(1).to.be.ok;`,
				Errors: parsedExpectError(
					"entry=expect head=true modifiers=[] matcher=ok reason=none static=false",
				),
			},
			{
				Code: `test("t", ({ expect }) => { expect(1).to.be.ok; });`,
				Errors: parsedExpectError(
					"entry=expect head=true modifiers=[] matcher=ok reason=none static=false",
				),
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
			{Code: `import { expect } from 'vitest'; expect(1).to.be.ok;`},
			{Code: `import { expect } from '@jest/globals'; expect(1).to.equal(1);`},
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
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
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

var expectIdentityProbe = rule.Rule{
	Name:             "rstest/expect-identity-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if analysis.IsExpectCall(node) {
					ctx.ReportNode(node, probeMessage("expectCall", "expect=true"))
				}
			},
		}
	},
}

func TestIsRstestExpectCallRecognizesChaiPropertyStyleOnce(t *testing.T) {
	expectCall := []rule_tester.InvalidTestCaseError{{MessageId: "expectCall", Message: "expect=true"}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectIdentityProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: `expect(value).to.be.ok;`, Errors: expectCall},
			{Code: `expect(spy).to.have.been.calledOnce;`, Errors: expectCall},
			{Code: `expect(value).to.equal(expected);`, Errors: expectCall},
		},
	)
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
			{Code: `expect(x).not.resolves.toBe(1);`, Errors: awaited},
			{Code: `expect(x).not.rejects.toThrow();`, Errors: awaited},
			{Code: `expect(x).resolves.to.equal(1);`, Errors: awaited},
			{Code: `expect(x).not.resolves.to.be.ok;`, Errors: awaited},
			{Code: `expect(x).rejects.to.have.property("message");`, Errors: awaited},
			{Code: `expect(x).toResolve();`, Errors: awaited},
			{Code: `expect(x).to.be.ok;`, Errors: notAwaited},
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
	// The asymmetric table answers "is this static member a known built-in
	// value constructor". A count assertion would only ever be updated to
	// match whatever the table says, so pin membership on both sides instead:
	// value constructors in, assertion matchers and modifiers out.
	for _, name := range []string{"stringContaining", "arrayContaining", "any", "anything"} {
		if !rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[name] {
			t.Errorf("asymmetric matchers must contain %q", name)
		}
	}
	for _, name := range []string{"toBe", "toEqual", "not", "resolves", "assertions", "soft"} {
		if rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[name] {
			t.Errorf("asymmetric matchers must not contain %q", name)
		}
	}

	for _, name := range []string{"assertions", "hasAssertions"} {
		if !rstestUtils.RSTEST_EXPECT_ASSERTION_COUNT_MEMBERS[name] {
			t.Errorf("assertion-count members must contain %q", name)
		}
	}
	if len(rstestUtils.RSTEST_EXPECT_ASSERTION_COUNT_MEMBERS) != 2 {
		t.Errorf("expected 2 assertion-count members, got %d", len(rstestUtils.RSTEST_EXPECT_ASSERTION_COUNT_MEMBERS))
	}

	for _, name := range []string{"extend", "getState", "setState", "addEqualityTesters", "addSnapshotSerializer"} {
		if !rstestUtils.RSTEST_EXPECT_CONFIG_MEMBERS[name] {
			t.Errorf("config members must contain %q", name)
		}
	}

	if len(rstestUtils.RSTEST_POLL_EXCLUDED_MEMBERS) != 11 {
		t.Errorf("expected 11 poll-excluded members, got %d", len(rstestUtils.RSTEST_POLL_EXCLUDED_MEMBERS))
	}
	for _, name := range []string{"resolves", "rejects", "toMatchInlineSnapshot"} {
		if !rstestUtils.RSTEST_POLL_EXCLUDED_MEMBERS[name] {
			t.Errorf("poll-excluded members must contain %q", name)
		}
	}

	propertyMatchers := []string{
		"ok",
		"true",
		"numeric",
		"callable",
		"false",
		"null",
		"undefined",
		"NaN",
		"exist",
		"exists",
		"empty",
		"arguments",
		"Arguments",
		"iterable",
		"extensible",
		"sealed",
		"frozen",
		"finite",
		"called",
		"calledOnce",
		"calledTwice",
		"calledThrice",
	}
	for _, matcher := range propertyMatchers {
		t.Run("property-"+matcher, func(t *testing.T) {
			code := fmt.Sprintf("expect(value).to.%s;", matcher)
			rule_tester.RunRuleTester(
				fixtures.GetRootDir(), "tsconfig.json", t, &expectChainParseProbe,
				[]rule_tester.ValidTestCase{},
				[]rule_tester.InvalidTestCase{{
					Code: code,
					Errors: parsedExpectError(fmt.Sprintf(
						"entry=expect head=true expression=property members=[to %s] modifiers=[] matcher=%s matchers=[%s:property] reason=none static=false",
						matcher,
						matcher,
						matcher,
					)),
				}},
			)
		})
	}
}
