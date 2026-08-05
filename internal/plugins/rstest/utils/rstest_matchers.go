package utils

// Constant tables for Rstest expect parsing. Every table is a snapshot of the
// audited baselines below; re-verify each one when upgrading Rstest or
// @vitest/expect.
//
// Baselines:
//   - rstest commit c4b67c72
//   - @vitest/expect 4.1.10 (rstest packages/core/package.json declares ^4.1.10)

// RSTEST_EXPECT_MODIFIER_NAMES lists the assertion chain modifiers.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:72-143 (Assertion
// exposes not / resolves / rejects); the set and the legal combinations are
// identical to Jest's.
var RSTEST_EXPECT_MODIFIER_NAMES = map[string]bool{
	"not":      true,
	"rejects":  true,
	"resolves": true,
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

// RSTEST_ASYMMETRIC_MATCHERS lists the static value constructors on the expect
// object: expect.<name>(...) builds a matcher value to pass into another
// assertion instead of asserting by itself.
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

// rstestExpectConfigMembers lists the ExpectStatic members that configure the
// expect runtime rather than asserting or building matcher values.
// Source: rstest c4b67c72 packages/core/src/types/expect.ts:147-175
// (addEqualityTesters / addSnapshotSerializer / getState / setState) and
// @vitest/expect@4.1.10 dist/index.d.ts:183-191 (extend, inherited through
// VitestExpectProperties).
var rstestExpectConfigMembers = map[string]bool{
	"addEqualityTesters":    true,
	"addSnapshotSerializer": true,
	"getState":              true,
	"setState":              true,
	"extend":                true,
}
