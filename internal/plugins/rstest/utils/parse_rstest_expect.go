package utils

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// RstestExpectEntry classifies which ExpectStatic surface a parsed expect call
// goes through.
type RstestExpectEntry string

const (
	// RstestExpectEntryCall covers plain expect(x) chains, and also bare
	// modifier chains such as expect.resolves.toBe(1) where the expect head was
	// never invoked; the latter parse with Head == nil.
	RstestExpectEntryCall RstestExpectEntry = "expect"
	// RstestExpectEntrySoft covers expect.soft(x) chains.
	RstestExpectEntrySoft RstestExpectEntry = "soft"
	// RstestExpectEntryPoll covers expect.poll(fn) chains.
	RstestExpectEntryPoll RstestExpectEntry = "poll"
	// RstestExpectEntryElement covers expect.element(locator) chains.
	RstestExpectEntryElement RstestExpectEntry = "element"
	// RstestExpectEntryUnreachable covers expect.unreachable(...).
	RstestExpectEntryUnreachable RstestExpectEntry = "unreachable"
	// RstestExpectEntryAssertions covers expect.assertions(n) and
	// expect.hasAssertions().
	RstestExpectEntryAssertions RstestExpectEntry = "assertions"
	// RstestExpectEntryAsymmetric covers static matcher value constructors:
	// expect.stringContaining(...), expect.not.arrayContaining(...), and
	// unknown static members, which may be custom asymmetric matchers
	// registered through expect.extend.
	RstestExpectEntryAsymmetric RstestExpectEntry = "asymmetric"
	// RstestExpectEntryConfig covers runtime configuration members such as
	// expect.extend(...) or expect.getState().
	RstestExpectEntryConfig RstestExpectEntry = "config"
)

// RstestExpectParseReason explains why a parsed expect call carries no
// matcher. The values and their triggers mirror eslint-plugin-jest and
// eslint-plugin-vitest.
type RstestExpectParseReason string

const (
	RstestExpectParseReasonNone            RstestExpectParseReason = ""
	RstestExpectParseReasonMatcherNotFound RstestExpectParseReason = "matcher-not-found"
	// RstestExpectParseReasonMatcherNotCalled reports chains that end on an
	// member access that is never invoked, e.g. expect(x).toBe; — the mapping is
	// the one jest/valid-expect and vitest apply: the matcher walk found
	// nothing and the call itself sits under a member access.
	RstestExpectParseReasonMatcherNotCalled RstestExpectParseReason = "matcher-not-called"
	RstestExpectParseReasonModifierUnknown  RstestExpectParseReason = "modifier-unknown"
)

// ParsedRstestExpectCall describes one Rstest expect call.
//
// Contract: ParseRstestExpectCall returns nil if and only if the node is not a
// Rstest expect call (which includes calls in the middle of a larger chain);
// once a node is recognized as an expect call the parse always succeeds and
// broken chains are reported through Reason instead of a nil result. This
// deliberately improves on jest's ParseJestFnCall, which returns nil for
// broken chains and forces valid-expect to re-derive the reason.
type ParsedRstestExpectCall struct {
	// Head is the assertion factory call: expect(x), expect.soft(x),
	// expect.poll(fn) or expect.element(loc). It is nil for static member
	// calls such as expect.assertions(1) and for bare modifier chains such as
	// expect.resolves.toBe(1) — the same shapes eslint-plugin-vitest reports
	// with a null expectCall. Consumers that need the factory arguments must
	// check for nil.
	Head *ast.Node
	// Entry classifies the ExpectStatic surface the call goes through.
	Entry RstestExpectEntry
	// Modifiers holds the modifier names before the matcher, in source order.
	Modifiers []string
	// ModifierEntries parallels Modifiers and carries the AST nodes.
	ModifierEntries []ParsedRstestFnMemberEntry
	// Matcher is the resolved matcher name; empty when Reason is set. Static
	// member calls resolve their member as the matcher (expect.assertions(1)
	// has Matcher "assertions"), mirroring jest and vitest.
	Matcher string
	// MatcherEntry carries the matcher's AST node; nil when Reason is set.
	MatcherEntry *ParsedRstestFnMemberEntry
	// Reason explains why no matcher was resolved.
	Reason RstestExpectParseReason

	// rootInvoked and trailingMembers feed IsStaticRstestExpectCall, which
	// needs the syntactic shape rather than the chain semantics.
	rootInvoked     bool
	trailingMembers int
}

func IsRstestExpectCall(
	node *ast.Node,
	ctx rule.RuleContext,
	callbacks RstestTestCallbacks,
) bool {
	if node == nil || node.Kind != ast.KindCallExpression || FindTopMostCallExpression(node) != node {
		return false
	}
	_, ok := rstestExpectMemberIndex(node, testFramework.GetMemberEntries(node), ctx, callbacks)
	return ok
}

// ParseRstestExpectCall parses node as a Rstest expect call, resolving the
// entry kind, the modifier chain and the matcher. See ParsedRstestExpectCall
// for the nil contract.
func ParseRstestExpectCall(
	node *ast.Node,
	ctx rule.RuleContext,
	callbacks RstestTestCallbacks,
) *ParsedRstestExpectCall {
	if node == nil || node.Kind != ast.KindCallExpression || FindTopMostCallExpression(node) != node {
		return nil
	}
	entries := testFramework.GetMemberEntries(node)
	expectIndex, ok := rstestExpectMemberIndex(node, entries, ctx, callbacks)
	if !ok {
		return nil
	}
	return classifyRstestExpectCall(node, entries, expectIndex)
}

// IsStaticRstestExpectCall reports calls of a single static member on the
// expect object itself, e.g. expect.anything() or expect.getState(), which
// rules like no-standalone-expect must not treat as standalone assertions.
// expect.assertions(n) and expect.hasAssertions() are deliberately excluded —
// they are misplaced outside test blocks — mirroring eslint-plugin-jest's
// no-standalone-expect (src/rules/no-standalone-expect.ts:89-98). The member
// count condition is upstream's members.length === 1: two-member forms such as
// expect.not.stringContaining(...) or expect.soft(x).toBe(1) stay reportable.
func IsStaticRstestExpectCall(parsed *ParsedRstestExpectCall) bool {
	return parsed != nil &&
		!parsed.rootInvoked &&
		parsed.trailingMembers == 1 &&
		parsed.Entry != RstestExpectEntryAssertions
}

// ShouldRstestExpectBeAwaited reports whether the assertion produces a promise
// that valid-expect requires to be awaited or returned: any modifier other
// than not (resolves / rejects), or a matcher the rule's asyncMatchers option
// declares async. Mirrors jest's shouldBeAwaited (valid_expect.go).
func ShouldRstestExpectBeAwaited(parsed *ParsedRstestExpectCall, asyncMatchers []string) bool {
	if parsed == nil {
		return false
	}
	for _, modifier := range parsed.Modifiers {
		if modifier != "not" {
			return true
		}
	}
	return slices.Contains(asyncMatchers, parsed.Matcher)
}

// FindTopMostCallExpression walks up the call/member chain that node is the
// head of. Ascending only continues while node stays on the callee side of its
// parent: a call in argument or computed-key position heads its own chain.
func FindTopMostCallExpression(node *ast.Node) *ast.Node {
	top := node
	current := node
	for parent := current.Parent; parent != nil; {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression != current {
				return top
			}
			top = parent
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return top
			}
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return top
			}
		default:
			return top
		}
		current, parent = parent, parent.Parent
	}
	return top
}

// rstestExpectMemberIndex resolves whether node is an expect call and where
// the expect member sits in entries: 0 when the root identifier itself is
// expect (imports, globals, destructured context), 1 when the root is a
// receiver (namespace object, import.meta.rstest, test context).
func rstestExpectMemberIndex(
	node *ast.Node,
	entries []testFramework.MemberEntry,
	ctx rule.RuleContext,
	callbacks RstestTestCallbacks,
) (int, bool) {
	if isImportMetaRstestExpectCall(node) {
		// GetMemberEntries cannot represent the import.meta prefix and starts
		// the chain at the rstest property, so expect sits at index 1.
		if len(entries) > 1 && entries[1].Name == "expect" {
			return 1, true
		}
		return 0, false
	}

	if len(entries) == 0 || entries[0].Node == nil || entries[0].Node.Kind != ast.KindIdentifier {
		return 0, false
	}
	root := entries[0].Node
	if index, ok := contextExpectMemberIndex(root, entries, ctx, callbacks); ok {
		return index, true
	}
	return importedOrGlobalExpectMemberIndex(root, entries, ctx)
}

func classifyRstestExpectCall(
	node *ast.Node,
	entries []testFramework.MemberEntry,
	expectIndex int,
) *ParsedRstestExpectCall {
	expectEntry := entries[expectIndex]
	rest := entries[expectIndex+1:]
	parsed := &ParsedRstestExpectCall{
		Entry:           RstestExpectEntryCall,
		Head:            expectEntry.Call,
		rootInvoked:     expectEntry.Call != nil,
		trailingMembers: len(rest),
	}

	chain := rest
	if !parsed.rootInvoked {
		if len(rest) == 0 {
			// GetMemberEntries marks the last entry of a call's callee chain as
			// invoked, so a top-most call whose expect root is not invoked always
			// has trailing members. Defensive only.
			return nil
		}
		if entry, ok := rstestExpectFactoryEntries[rest[0].Name]; ok && rest[0].Call != nil {
			// expect.soft(x) / expect.poll(fn) / expect.element(loc): the factory
			// call replaces expect(x) as the assertion head and the chain resumes
			// after it.
			parsed.Entry = entry
			parsed.Head = rest[0].Call
			chain = rest[1:]
		} else {
			// Static member access on the expect object. The chain keeps the
			// member: the walk resolves it as the matcher (expect.assertions(1)
			// yields Matcher "assertions"), matching jest and vitest.
			parsed.Entry = classifyStaticExpectEntry(rest)
		}
	}

	modifierEntries, matcher, reason := findRstestExpectModifiersAndMatcher(chain)
	parsed.Reason = reason
	if reason == RstestExpectParseReasonMatcherNotFound && isMemberAccessNode(node.Parent) {
		parsed.Reason = RstestExpectParseReasonMatcherNotCalled
	}
	if matcher != nil {
		parsed.Matcher = matcher.Name
		parsed.MatcherEntry = matcher
		parsed.ModifierEntries = modifierEntries
		if len(modifierEntries) > 0 {
			parsed.Modifiers = make([]string, len(modifierEntries))
			for i, entry := range modifierEntries {
				parsed.Modifiers[i] = entry.Name
			}
		}
	}
	return parsed
}

// classifyStaticExpectEntry names the ExpectStatic surface behind a not-invoked
// expect root. rest is never empty here.
func classifyStaticExpectEntry(rest []testFramework.MemberEntry) RstestExpectEntry {
	name := rest[0].Name
	switch {
	case name == "unreachable":
		return RstestExpectEntryUnreachable
	case name == "assertions", name == "hasAssertions":
		return RstestExpectEntryAssertions
	case rstestExpectConfigMembers[name]:
		return RstestExpectEntryConfig
	case name == "not" && len(rest) > 1 && RSTEST_ASYMMETRIC_MATCHERS[rest[1].Name]:
		// ExpectStatic.not.stringContaining(...): the static negation of an
		// asymmetric matcher, unrelated to the Assertion.not modifier.
		return RstestExpectEntryAsymmetric
	case RSTEST_EXPECT_MODIFIER_NAMES[name]:
		// expect.resolves.toBe(1) and friends: a bare modifier chain whose
		// expect head was never invoked. It parses like a plain chain with a
		// nil Head, mirroring vitest's null expectCall.
		return RstestExpectEntryCall
	case rstestExpectFactoryEntries[name] != "":
		// A factory member that was not invoked, e.g. expect.soft.foo(1); the
		// walk rejects the chain with modifier-unknown.
		return rstestExpectFactoryEntries[name]
	default:
		// Known asymmetric matchers, plus unknown static members that may be
		// custom asymmetric matchers registered through expect.extend.
		return RstestExpectEntryAsymmetric
	}
}

// findRstestExpectModifiersAndMatcher locates the matcher — the first chain
// member whose enclosing member access is invoked — and validates the
// modifiers before it, mirroring Vitest's expect-chain validation: not may
// appear at most once, and resolves / rejects may appear at most once in
// total, regardless of their order.
func findRstestExpectModifiersAndMatcher(entries []testFramework.MemberEntry) (
	[]ParsedRstestFnMemberEntry,
	*ParsedRstestFnMemberEntry,
	RstestExpectParseReason,
) {
	if len(entries) == 0 {
		return nil, nil, RstestExpectParseReasonMatcherNotFound
	}

	modifiers := make([]ParsedRstestFnMemberEntry, 0, len(entries))
	notCount := 0
	promiseModifierCount := 0
	for _, member := range entries {
		if member.Node == nil {
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}
		if isComputedDynamicMemberName(member.Node) {
			// expect(x)[matcherName](1): the member name cannot be resolved
			// statically, so no matcher can be located.
			return nil, nil, RstestExpectParseReasonMatcherNotFound
		}
		parent := member.Node.Parent
		if parent == nil {
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}

		grandparent := ast.WalkUpParenthesizedExpressions(parent.Parent)
		if grandparent != nil && grandparent.Kind == ast.KindCallExpression {
			return modifiers, &member, RstestExpectParseReasonNone
		}

		if !RSTEST_EXPECT_MODIFIER_NAMES[member.Name] {
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}

		if member.Name == "not" {
			notCount++
			if notCount > 1 {
				return nil, nil, RstestExpectParseReasonModifierUnknown
			}
		} else {
			promiseModifierCount++
			if promiseModifierCount > 1 {
				return nil, nil, RstestExpectParseReasonModifierUnknown
			}
		}

		modifiers = append(modifiers, member)
	}

	return nil, nil, RstestExpectParseReasonMatcherNotFound
}

// isComputedDynamicMemberName reports member name nodes that came from a
// computed element access with an identifier key, e.g. expect(x)[matcherName].
// GetMemberEntries records the identifier's text as the member name, but the
// runtime value is unknowable statically. String literal keys are fine.
func isComputedDynamicMemberName(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil &&
		parent.Kind == ast.KindElementAccessExpression &&
		parent.AsElementAccessExpression().Expression != node
}

func isMemberAccessNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return true
	default:
		return false
	}
}

// contextExpectMemberIndex resolves expect calls that come from a test context
// parameter: ({ expect }) => expect(x)... at index 0, ctx => ctx.expect(x)...
// at index 1.
func contextExpectMemberIndex(
	root *ast.Node,
	entries []testFramework.MemberEntry,
	ctx rule.RuleContext,
	callbacks RstestTestCallbacks,
) (int, bool) {
	if ctx.TypeChecker == nil {
		return 0, false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(root)
	if symbol == nil {
		return 0, false
	}
	if callbacks.ContextExpectNames[symbol] {
		return 0, true
	}
	if callbacks.ContextReceivers[symbol] &&
		len(entries) > 1 &&
		entries[1].Name == "expect" {
		return 1, true
	}
	return 0, false
}

// importedOrGlobalExpectMemberIndex resolves expect calls whose root is an
// import, a global, a require binding, an import.meta.rstest alias, or a
// module namespace: direct bindings put expect at index 0, receiver objects at
// index 1.
func importedOrGlobalExpectMemberIndex(
	root *ast.Node,
	entries []testFramework.MemberEntry,
	ctx rule.RuleContext,
) (int, bool) {
	localName := root.AsIdentifier().Text
	var symbol *ast.Symbol
	if ctx.TypeChecker != nil {
		symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
	}

	if name, _, ok := resolveImportMetaRstestBinding(symbol); ok {
		return 0, name == "expect"
	}
	if isImportMetaRstestObjectAlias(symbol) {
		if len(entries) > 1 && entries[1].Name == "expect" {
			return 1, true
		}
		return 0, false
	}

	for _, module := range []string{RstestImportModule, RstestPlaywrightImportModule} {
		name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
			localName,
			root,
			symbol,
			ctx.SourceFile,
			module,
		)
		if name == "expect" {
			return 0, true
		}
		if testFramework.IsModuleNamespaceSymbol(symbol, module) &&
			len(entries) > 1 &&
			entries[1].Name == "expect" {
			return 1, true
		}
	}
	return 0, false
}

func isImportMetaRstestExpectCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	_, parts, _, ok := parseImportMetaRstestChain(call.Expression)
	return ok && len(parts) > 0 && parts[0].name == "expect"
}

func isImportMetaRstestObjectAlias(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		if isImportMetaRstest(initializer) {
			return true
		}
	}
	return false
}
