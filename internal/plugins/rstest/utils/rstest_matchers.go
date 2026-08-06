package utils

// Constant tables for Rstest expect parsing. Every table is a snapshot of the
// audited baselines below; re-verify each one when upgrading Rstest or
// @vitest/expect.
//
// Baselines:
//   - rstest commit c4b67c72
//   - @vitest/expect 4.1.10 (rstest packages/core/package.json declares ^4.1.10)
//   - Chai 6.2.2

// RSTEST_EXPECT_MODIFIER_NAMES lists the assertion chain modifiers.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:72-143 (Assertion
// exposes not / resolves / rejects). Legal combinations follow Vitest's
// count-based validation: at most one not and one promise modifier.
var RSTEST_EXPECT_MODIFIER_NAMES = map[string]bool{
	"not":      true,
	"rejects":  true,
	"resolves": true,
}

// rstestChaiLanguageChains lists Chai's no-op language chains plus `itself`,
// which sets a flag consumed by respondTo. They connect assertions but are
// never matchers by themselves.
// Source: chai@6.2.2 lib/chai/core/assertions.js (language chains and itself),
// cross-checked against eslint-plugin-vitest@1.6.26
// src/utils/parse-vitest-fn-call.ts.
var rstestChaiLanguageChains = map[string]bool{
	"also":   true,
	"and":    true,
	"at":     true,
	"be":     true,
	"been":   true,
	"but":    true,
	"does":   true,
	"has":    true,
	"have":   true,
	"is":     true,
	"itself": true,
	"of":     true,
	"same":   true,
	"still":  true,
	"that":   true,
	"to":     true,
	"which":  true,
	"with":   true,
}

// rstestChaiFlagChains lists Chai properties that alter later matchers without
// performing an assertion themselves.
// Source: chai@6.2.2 lib/chai/core/assertions.js, cross-checked against
// eslint-plugin-vitest@1.6.26 src/utils/parse-vitest-fn-call.ts.
var rstestChaiFlagChains = map[string]bool{
	"all":     true,
	"any":     true,
	"deep":    true,
	"nested":  true,
	"ordered": true,
	"own":     true,
}

// rstestChaiDualRoleChains lists Chai chainable methods: an uncalled property
// access configures or connects a later assertion, while invoking the same
// member makes it the matcher.
// Source: chai@6.2.2 lib/chai/core/assertions.js, cross-checked against
// eslint-plugin-vitest@1.6.26 src/utils/parse-vitest-fn-call.ts.
var rstestChaiDualRoleChains = map[string]bool{
	"a":         true,
	"an":        true,
	"contain":   true,
	"contains":  true,
	"include":   true,
	"includes":  true,
	"length":    true,
	"lengthOf":  true,
	"respondTo": true,
}

// rstestChaiChainableMembers is the union of the language, flag and dual-role
// tables. The source tables remain separate because classification depends on
// whether a dual-role member is invoked.
var rstestChaiChainableMembers = mergeRstestExpectNameSets(
	rstestChaiLanguageChains,
	rstestChaiFlagChains,
	rstestChaiDualRoleChains,
)

// rstestChaiPropertyMatchers lists official getter assertions that execute
// without a call expression. Unknown terminal properties are deliberately not
// accepted because they may be misspelled method matchers.
// Sources:
//   - chai@6.2.2 lib/chai/core/assertions.js (core property assertions)
//   - @vitest/expect@4.1.10 src/chai-style-assertions.ts (called properties)
var rstestChaiPropertyMatchers = map[string]bool{
	"Arguments":    true,
	"NaN":          true,
	"arguments":    true,
	"callable":     true,
	"called":       true,
	"calledOnce":   true,
	"calledThrice": true,
	"calledTwice":  true,
	"empty":        true,
	"exist":        true,
	"exists":       true,
	"extensible":   true,
	"false":        true,
	"finite":       true,
	"frozen":       true,
	"iterable":     true,
	"null":         true,
	"numeric":      true,
	"ok":           true,
	"sealed":       true,
	"true":         true,
	"undefined":    true,
}

// RSTEST_MATCHER_ALIASES maps alias matchers to their canonical names.
// Source: @vitest/expect@4.1.10 dist/index.d.ts (JestAssertion), cross-checked
// against eslint-plugin-vitest@1.6.26 no-alias-methods; the 11 pairs are
// identical to Jest's.
var RSTEST_MATCHER_ALIASES = map[string]string{
	"toBeCalled":       "toHaveBeenCalled",
	"toBeCalledTimes":  "toHaveBeenCalledTimes",
	"toBeCalledWith":   "toHaveBeenCalledWith",
	"lastCalledWith":   "toHaveBeenLastCalledWith",
	"nthCalledWith":    "toHaveBeenNthCalledWith",
	"toReturn":         "toHaveReturned",
	"toReturnTimes":    "toHaveReturnedTimes",
	"toReturnWith":     "toHaveReturnedWith",
	"lastReturnedWith": "toHaveLastReturnedWith",
	"nthReturnedWith":  "toHaveNthReturnedWith",
	"toThrowError":     "toThrow",
}

// RSTEST_INLINE_SNAPSHOT_MATCHERS lists the snapshot matchers whose expected
// value lives inline in the test file.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:86-127; identical
// to Jest's INLINE_SNAPSHOT_MATCHERS.
var RSTEST_INLINE_SNAPSHOT_MATCHERS = map[string]bool{
	"toMatchInlineSnapshot":              true,
	"toThrowErrorMatchingInlineSnapshot": true,
}

// RSTEST_SNAPSHOT_MATCHERS lists every snapshot matcher, inline ones included.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:86-127. Rstest
// extends the Jest set with matchSnapshot (short alias of toMatchSnapshot) and
// toMatchFileSnapshot.
var RSTEST_SNAPSHOT_MATCHERS = map[string]bool{
	"matchSnapshot":                      true,
	"toMatchSnapshot":                    true,
	"toMatchInlineSnapshot":              true,
	"toThrowErrorMatchingSnapshot":       true,
	"toThrowErrorMatchingInlineSnapshot": true,
	"toMatchFileSnapshot":                true,
}

// RSTEST_BROWSER_ELEMENT_MATCHERS lists the matchers available on
// expect.element(locator) chains in Browser Mode. They form a set independent
// from Assertion; the parser resolves them through the same matcher walk and
// this table only exists for consumers that need to query membership.
// Source: rstest c4b67c72 packages/browser/src/augmentExpect.ts:3-58
// (BrowserElementExpect).
var RSTEST_BROWSER_ELEMENT_MATCHERS = map[string]bool{
	"toBeVisible":      true,
	"toBeHidden":       true,
	"toBeEnabled":      true,
	"toBeDisabled":     true,
	"toBeChecked":      true,
	"toBeUnchecked":    true,
	"toBeAttached":     true,
	"toBeDetached":     true,
	"toBeEditable":     true,
	"toBeFocused":      true,
	"toBeEmpty":        true,
	"toBeInViewport":   true,
	"toHaveText":       true,
	"toContainText":    true,
	"toHaveValue":      true,
	"toHaveId":         true,
	"toHaveAttribute":  true,
	"toHaveClass":      true,
	"toHaveCount":      true,
	"toHaveCSS":        true,
	"toHaveJSProperty": true,
}

// RSTEST_ASYMMETRIC_MATCHERS lists the built-in static value constructors on
// the expect object: expect.<name>(...) builds a matcher value to pass into
// another assertion instead of asserting by itself.
//
// The parser does not consult this table — a static member it does not
// recognize could be a custom asymmetric matcher registered through
// expect.extend or a forgotten expect(x), and one file cannot tell them apart.
// The table is how a consumer asks the question the parser refuses to guess
// at: given Entry == RstestExpectEntryStatic, is Matcher a *known* value
// constructor? A miss means "unknown", not "not asymmetric" — including for
// the negated form, where Modifiers is ["not"] and Matcher is the constructor
// name (expect.not.arrayContaining([])).
// Source: @vitest/expect@4.1.10 dist/index.d.ts:183-262 — ExpectStatic
// (anything / any), AsymmetricMatchersContaining (stringContaining /
// stringMatching / objectContaining / arrayContaining / closeTo /
// schemaMatching) and CustomMatcher (toSatisfy / toBeOneOf). toSatisfy and
// toBeOneOf double as instance matchers; this table covers their static
// asymmetric form.
var RSTEST_ASYMMETRIC_MATCHERS = map[string]bool{
	"anything":         true,
	"any":              true,
	"stringContaining": true,
	"stringMatching":   true,
	"objectContaining": true,
	"arrayContaining":  true,
	"closeTo":          true,
	"schemaMatching":   true,
	"toSatisfy":        true,
	"toBeOneOf":        true,
}

// RSTEST_POLL_EXCLUDED_MEMBERS lists the Assertion members that the type of
// expect.poll(...) removes. The parser records the set but does not enforce
// it: writing expect.poll(fn).resolves.toBe(1) is a type error TypeScript
// already reports, so the parser stays descriptive and future rules that want
// the check query this table instead of re-auditing the source.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:151-167 (the
// Omit<...> on the poll return type).
var RSTEST_POLL_EXCLUDED_MEMBERS = map[string]bool{
	"rejects":                            true,
	"resolves":                           true,
	"toThrow":                            true,
	"toThrowError":                       true,
	"throw":                              true,
	"throws":                             true,
	"matchSnapshot":                      true,
	"toMatchSnapshot":                    true,
	"toMatchInlineSnapshot":              true,
	"toThrowErrorMatchingSnapshot":       true,
	"toThrowErrorMatchingInlineSnapshot": true,
}

// rstestExpectFactoryEntries maps the ExpectStatic members that, when invoked,
// start an assertion chain of their own, replacing expect(x) as the head.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:147-175 (soft /
// poll) and packages/browser/src/augmentExpect.ts:59-62 (element).
var rstestExpectFactoryEntries = map[string]RstestExpectEntry{
	"soft":    RstestExpectEntrySoft,
	"poll":    RstestExpectEntryPoll,
	"element": RstestExpectEntryElement,
}

// RSTEST_EXPECT_ASSERTION_COUNT_MEMBERS lists the ExpectStatic members that
// declare how many assertions a test must run. They are the one group of
// static members that is misplaced outside a test block, which is why
// IsStaticRstestExpectCall excludes them and why they are a table rather than
// two string literals at the one call site.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:147-175.
var RSTEST_EXPECT_ASSERTION_COUNT_MEMBERS = map[string]bool{
	"assertions":    true,
	"hasAssertions": true,
}

// RSTEST_EXPECT_CONFIG_MEMBERS lists the ExpectStatic members that configure
// the expect runtime rather than asserting or building matcher values.
// Consumers reach it through Matcher when Entry is RstestExpectEntryStatic.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:147-175
// (addEqualityTesters / addSnapshotSerializer / getState / setState) and
// @vitest/expect@4.1.10 dist/index.d.ts:183-191 (extend, inherited through
// VitestExpectProperties).
var RSTEST_EXPECT_CONFIG_MEMBERS = map[string]bool{
	"addEqualityTesters":    true,
	"addSnapshotSerializer": true,
	"getState":              true,
	"setState":              true,
	"extend":                true,
}

func mergeRstestExpectNameSets(sets ...map[string]bool) map[string]bool {
	size := 0
	for _, set := range sets {
		size += len(set)
	}
	merged := make(map[string]bool, size)
	for _, set := range sets {
		for name := range set {
			merged[name] = true
		}
	}
	return merged
}
