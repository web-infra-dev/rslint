package utils

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// RstestExpectEntry names the assertion factory a parsed expect call went
// through. The four factory values mirror eslint-plugin-vitest's
// expectApiMethods plus its default expect(x) case.
//
// Everything the expect object exposes that is not an assertion factory —
// expect.assertions(n), expect.unreachable(), expect.stringContaining(...),
// expect.getState() — is RstestExpectEntryStatic. The parser deliberately does
// not sub-classify those: telling a custom asymmetric matcher registered
// through expect.extend apart from a forgotten expect(x) is not decidable from
// one file, so the distinction is left to consumers, which have the member
// name in Matcher and the tables in rstest_matchers.go to query. Jest and
// vitest make the same call: neither classifies static members, both detect
// them structurally (vitest's expectCall == null, our Head == nil).
type RstestExpectEntry string

const (
	// RstestExpectEntryCall covers plain expect(x) chains.
	RstestExpectEntryCall RstestExpectEntry = "expect"
	// RstestExpectEntrySoft covers expect.soft(x) chains.
	RstestExpectEntrySoft RstestExpectEntry = "soft"
	// RstestExpectEntryPoll covers expect.poll(fn) chains.
	RstestExpectEntryPoll RstestExpectEntry = "poll"
	// RstestExpectEntryElement covers expect.element(locator) chains.
	RstestExpectEntryElement RstestExpectEntry = "element"
	// RstestExpectEntryStatic covers every chain where no assertion factory was
	// invoked: static members of the expect object such as
	// expect.assertions(1), expect.stringContaining("a") and
	// expect.not.arrayContaining([]), and bare modifier chains such as
	// expect.resolves.toBe(1) whose expect(x) call the author forgot.
	//
	// Invariant: Entry == RstestExpectEntryStatic if and only if Head == nil.
	RstestExpectEntryStatic RstestExpectEntry = "static"
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

// RstestExpectMatcherKind records whether an assertion executes through a
// method call or through a Chai property getter.
type RstestExpectMatcherKind string

const (
	RstestExpectMatcherCall     RstestExpectMatcherKind = "call"
	RstestExpectMatcherProperty RstestExpectMatcherKind = "property"
)

// ParsedRstestExpectMatcher describes one matcher in an assertion chain.
// Chai permits multiple assertions in one chain, so callers that need the
// complete chain should use ParsedRstestExpectCall.Matchers.
type ParsedRstestExpectMatcher struct {
	Name  string
	Entry ParsedRstestFnMemberEntry
	Kind  RstestExpectMatcherKind
}

// ParsedRstestExpectCall describes one Rstest expect call.
//
// Contract: RstestCallAnalysis.ParseExpectCall returns nil if and only if the node is not a
// Rstest expect call (which includes calls in the middle of a larger chain);
// once a node is recognized as an expect call the parse always succeeds and
// broken chains are reported through Reason instead of a nil result. This
// deliberately improves on jest's ParseJestFnCall, which returns nil for
// broken chains and forces valid-expect to re-derive the reason.
type ParsedRstestExpectCall struct {
	// Expression is the outermost assertion expression. It is a call for
	// method-style chains and a member access for property-style Chai
	// assertions such as expect(value).to.be.ok.
	Expression *ast.Node
	// Head is the assertion factory call: expect(x), expect.soft(x),
	// expect.poll(fn) or expect.element(loc). It is nil for static member
	// calls such as expect.assertions(1) and for bare modifier chains such as
	// expect.resolves.toBe(1) — the same shapes eslint-plugin-vitest reports
	// with a null expectCall. Consumers that need the factory arguments must
	// check for nil.
	Head *ast.Node
	// Entry names the assertion factory, or RstestExpectEntryStatic when none
	// was invoked. Head is nil exactly when Entry is RstestExpectEntryStatic,
	// so either field answers "is this a real assertion"; Entry additionally
	// says which factory.
	Entry RstestExpectEntry
	// Members contains the complete source chain after the resolved expect
	// root, including assertion factories, Chai language chains, modifiers and
	// every matcher.
	Members []string
	// MemberEntries parallels Members and carries the AST nodes.
	MemberEntries []ParsedRstestFnMemberEntry
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
	// Matchers contains every matcher in source order. Matcher and MatcherEntry
	// mirror the first element for compatibility with matcher-consuming rules.
	Matchers []ParsedRstestExpectMatcher
	// Reason explains why no matcher was resolved.
	Reason RstestExpectParseReason

	// rootInvoked and trailingMembers feed IsStaticRstestExpectCall, which
	// needs the syntactic shape rather than the chain semantics.
	rootInvoked     bool
	trailingMembers int
}

func isRstestExpectCall(
	node *ast.Node,
	analysis *RstestCallAnalysis,
) bool {
	if node == nil || node.Kind != ast.KindCallExpression || FindTopMostCallExpression(node) != node {
		return false
	}
	if isImportMetaRstestExpectCall(node) {
		return true
	}
	entries := firstRstestExpectEntries(node)
	if entries.count == 0 || entries.first == nil || entries.first.Kind != ast.KindIdentifier {
		return false
	}
	_, ok := rstestExpectRootIndex(
		entries.first,
		entries.secondName,
		entries.count > 1,
		analysis,
	)
	return ok
}

type rstestExpectEntries struct {
	first      *ast.Node
	secondName string
	count      uint8
}

func (entries *rstestExpectEntries) append(name string, node *ast.Node) {
	if name == "" {
		return
	}
	switch entries.count {
	case 0:
		entries.first = node
	case 1:
		entries.secondName = name
	}
	if entries.count < 2 {
		entries.count++
	}
}

func firstRstestExpectEntries(node *ast.Node) rstestExpectEntries {
	node = ast.SkipParentheses(node)
	if node == nil {
		return rstestExpectEntries{}
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return rstestExpectEntries{
			first: node,
			count: 1,
		}
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		entries := firstRstestExpectEntries(property.Expression)
		name := property.Name()
		if name != nil {
			switch name.Kind {
			case ast.KindIdentifier:
				entries.append(name.AsIdentifier().Text, name)
			case ast.KindPrivateIdentifier:
				entries.append(name.AsPrivateIdentifier().Text, name)
			}
		}
		return entries
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		entries := firstRstestExpectEntries(element.Expression)
		name := ast.SkipParentheses(element.ArgumentExpression)
		if name == nil {
			return rstestExpectEntries{}
		}
		switch name.Kind {
		case ast.KindIdentifier:
			entries.append(name.AsIdentifier().Text, name)
		case ast.KindStringLiteral:
			entries.append(name.AsStringLiteral().Text, name)
		case ast.KindNoSubstitutionTemplateLiteral:
			entries.append(name.AsNoSubstitutionTemplateLiteral().Text, name)
		default:
			return rstestExpectEntries{}
		}
		return entries
	case ast.KindCallExpression:
		return firstRstestExpectEntries(node.AsCallExpression().Expression)
	case ast.KindTaggedTemplateExpression:
		return firstRstestExpectEntries(node.AsTaggedTemplateExpression().Tag)
	default:
		return rstestExpectEntries{}
	}
}

func parseRstestExpectCall(
	node *ast.Node,
	analysis *RstestCallAnalysis,
) *ParsedRstestExpectCall {
	if node == nil || node.Kind != ast.KindCallExpression || FindTopMostCallExpression(node) != node {
		return nil
	}
	expression := findTopMostRstestExpectExpression(node)
	entries := testFramework.GetMemberEntries(expression)
	expectIndex, ok := rstestExpectMemberIndex(node, entries, analysis)
	if !ok {
		return nil
	}
	return classifyRstestExpectCall(expression, entries, expectIndex)
}

// IsStaticRstestExpectCall reports calls of a single static member on the
// expect object itself, e.g. expect.anything() or expect.getState(), which
// rules like no-standalone-expect must not treat as standalone assertions.
// expect.assertions(n) and expect.hasAssertions() are deliberately excluded —
// they are misplaced outside test blocks — mirroring eslint-plugin-jest's
// no-standalone-expect (src/rules/no-standalone-expect.ts:89-98). The member
// count condition is upstream's members.length === 1: two-member forms such as
// expect.not.stringContaining(...) or expect.soft(x).toBe(1) stay reportable.
//
// The check is structural on purpose and never consults a matcher-name table,
// so a custom asymmetric matcher registered through expect.extend behaves like
// a built-in one.
func IsStaticRstestExpectCall(parsed *ParsedRstestExpectCall) bool {
	return parsed != nil &&
		!parsed.rootInvoked &&
		parsed.trailingMembers == 1 &&
		!RSTEST_EXPECT_ASSERTION_COUNT_MEMBERS[parsed.Matcher]
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

// findTopMostRstestExpectExpression extends the outermost call through a
// trailing static member chain. This lets the CallExpression-only parser see
// property-style Chai assertions such as expect(value).to.be.ok without
// requiring rules to listen for member-access nodes.
func findTopMostRstestExpectExpression(node *ast.Node) *ast.Node {
	top := node
	current := node
	for parent := current.Parent; parent != nil; {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return top
			}
			top = parent
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return top
			}
			top = parent
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
	analysis *RstestCallAnalysis,
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
	secondName := ""
	if len(entries) > 1 {
		secondName = entries[1].Name
	}
	return rstestExpectRootIndex(
		entries[0].Node,
		secondName,
		len(entries) > 1,
		analysis,
	)
}

func rstestExpectRootIndex(
	root *ast.Node,
	secondName string,
	hasSecond bool,
	analysis *RstestCallAnalysis,
) (int, bool) {
	if root == nil || root.Kind != ast.KindIdentifier {
		return 0, false
	}
	localName := root.AsIdentifier().Text
	ctx := analysis.ctx
	if ctx.TypeChecker == nil {
		return 0, localName == "expect"
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(root)
	if symbol == nil {
		return 0, localName == "expect"
	}
	kind, ok := analysis.expectRoots[symbol]
	if !ok {
		kind = classifyRstestExpectRoot(
			localName,
			root,
			symbol,
			analysis,
		)
		analysis.expectRoots[symbol] = kind
	}
	switch kind {
	case rstestExpectRootDirect:
		return 0, true
	case rstestExpectRootReceiver:
		return 1, hasSecond && secondName == "expect"
	default:
		return 0, false
	}
}

func classifyRstestExpectCall(
	expression *ast.Node,
	entries []testFramework.MemberEntry,
	expectIndex int,
) *ParsedRstestExpectCall {
	expectEntry := entries[expectIndex]
	rest := entries[expectIndex+1:]
	parsed := &ParsedRstestExpectCall{
		Expression:      expression,
		Entry:           RstestExpectEntryCall,
		Head:            expectEntry.Call,
		rootInvoked:     expectEntry.Call != nil,
		trailingMembers: len(rest),
	}
	if len(rest) > 0 {
		parsed.Members = make([]string, len(rest))
		parsed.MemberEntries = make([]ParsedRstestFnMemberEntry, len(rest))
		for i, entry := range rest {
			parsed.Members[i] = entry.Name
			parsed.MemberEntries[i] = entry
		}
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
			parsed.Entry = RstestExpectEntryStatic
		}
	}

	modifierEntries, matchers, reason := findRstestExpectModifiersAndMatchers(
		chain,
		!parsed.rootInvoked && parsed.Head == nil,
	)
	parsed.Reason = reason
	if reason == RstestExpectParseReasonMatcherNotFound &&
		len(chain) > 0 &&
		isMemberAccessNode(expression) &&
		!hasComputedDynamicRstestExpectMember(chain) {
		parsed.Reason = RstestExpectParseReasonMatcherNotCalled
	}
	if len(matchers) > 0 {
		parsed.Matchers = matchers
		parsed.Matcher = matchers[0].Name
		parsed.MatcherEntry = &parsed.Matchers[0].Entry
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

func hasComputedDynamicRstestExpectMember(entries []testFramework.MemberEntry) bool {
	for _, entry := range entries {
		if entry.Node != nil && isComputedDynamicMemberName(entry.Node) {
			return true
		}
	}
	return false
}

type rstestExpectChainKind uint8

const (
	rstestExpectChainUnknown rstestExpectChainKind = iota
	rstestExpectChainLanguage
	rstestExpectChainModifier
	rstestExpectChainMatcher
)

type rstestExpectChainEntry struct {
	Entry       ParsedRstestFnMemberEntry
	Kind        rstestExpectChainKind
	MatcherKind RstestExpectMatcherKind
}

// findRstestExpectModifiersAndMatchers classifies the complete assertion
// chain, returns every matcher in source order and validates the language and
// modifier members around them. Matcher and MatcherEntry on the public result
// continue to expose the first matcher for compatibility.
func findRstestExpectModifiersAndMatchers(
	entries []testFramework.MemberEntry,
	isStaticExpectChain bool,
) (
	[]ParsedRstestFnMemberEntry,
	[]ParsedRstestExpectMatcher,
	RstestExpectParseReason,
) {
	if len(entries) == 0 {
		return nil, nil, RstestExpectParseReasonMatcherNotFound
	}

	chains := make([]rstestExpectChainEntry, len(entries))
	matcherIndex := -1
	for i, member := range entries {
		if member.Node == nil {
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}
		if isComputedDynamicMemberName(member.Node) {
			return nil, nil, RstestExpectParseReasonMatcherNotFound
		}
		chains[i] = classifyRstestExpectChainEntry(
			member,
			isStaticExpectChain && i == 0,
		)
		if matcherIndex == -1 && chains[i].Kind == rstestExpectChainMatcher {
			matcherIndex = i
		}
	}

	modifierEnd := matcherIndex
	if modifierEnd == -1 {
		modifierEnd = len(chains)
	}
	modifiers := make([]ParsedRstestFnMemberEntry, 0, modifierEnd)
	notCount := 0
	promiseModifierCount := 0
	for i := range modifierEnd {
		chain := chains[i]
		if chain.Kind == rstestExpectChainUnknown {
			// With no matcher, an unknown terminal member may simply be a
			// matcher whose call was omitted. Earlier unknown members make the
			// chain invalid.
			if matcherIndex == -1 && i == len(chains)-1 {
				continue
			}
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}
		if chain.Kind == rstestExpectChainLanguage {
			if chain.Entry.Call != nil {
				return nil, nil, RstestExpectParseReasonModifierUnknown
			}
			continue
		}
		if chain.Kind != rstestExpectChainModifier {
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}

		if chain.Entry.Name == "not" {
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

		modifiers = append(modifiers, chain.Entry)
	}

	if matcherIndex == -1 {
		return nil, nil, RstestExpectParseReasonMatcherNotFound
	}

	matchers := make([]ParsedRstestExpectMatcher, 0, len(chains)-matcherIndex)
	for i := matcherIndex; i < len(chains); i++ {
		chain := chains[i]
		switch chain.Kind {
		case rstestExpectChainMatcher:
			matchers = append(matchers, ParsedRstestExpectMatcher{
				Name:  chain.Entry.Name,
				Entry: chain.Entry,
				Kind:  chain.MatcherKind,
			})
		case rstestExpectChainLanguage:
			if chain.Entry.Call != nil {
				return nil, nil, RstestExpectParseReasonModifierUnknown
			}
		case rstestExpectChainModifier:
			// Chai permits modifiers between assertions in a multi-matcher
			// chain, e.g. .a("string").that.does.not.contain("x").
		case rstestExpectChainUnknown:
			return nil, nil, RstestExpectParseReasonModifierUnknown
		}
	}

	return modifiers, matchers, RstestExpectParseReasonNone
}

func classifyRstestExpectChainEntry(
	entry ParsedRstestFnMemberEntry,
	isFirstStaticMember bool,
) rstestExpectChainEntry {
	name := entry.Name
	if entry.Call != nil {
		if isFirstStaticMember && name == "any" {
			return rstestExpectChainEntry{
				Entry:       entry,
				Kind:        rstestExpectChainMatcher,
				MatcherKind: RstestExpectMatcherCall,
			}
		}
		if rstestChaiPropertyMatchers[name] ||
			RSTEST_EXPECT_MODIFIER_NAMES[name] ||
			(rstestChaiChainableMembers[name] && !rstestChaiDualRoleChains[name]) {
			return rstestExpectChainEntry{
				Entry: entry,
				Kind:  rstestExpectChainLanguage,
			}
		}
		return rstestExpectChainEntry{
			Entry:       entry,
			Kind:        rstestExpectChainMatcher,
			MatcherKind: RstestExpectMatcherCall,
		}
	}

	if rstestChaiPropertyMatchers[name] {
		return rstestExpectChainEntry{
			Entry:       entry,
			Kind:        rstestExpectChainMatcher,
			MatcherKind: RstestExpectMatcherProperty,
		}
	}
	if rstestChaiChainableMembers[name] {
		return rstestExpectChainEntry{
			Entry: entry,
			Kind:  rstestExpectChainLanguage,
		}
	}
	if RSTEST_EXPECT_MODIFIER_NAMES[name] {
		return rstestExpectChainEntry{
			Entry: entry,
			Kind:  rstestExpectChainModifier,
		}
	}
	return rstestExpectChainEntry{
		Entry: entry,
		Kind:  rstestExpectChainUnknown,
	}
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

type rstestExpectRootKind uint8

const (
	rstestExpectRootNone rstestExpectRootKind = iota
	rstestExpectRootDirect
	rstestExpectRootReceiver
)

func classifyRstestExpectRoot(
	localName string,
	root *ast.Node,
	symbol *ast.Symbol,
	analysis *RstestCallAnalysis,
) rstestExpectRootKind {
	callbacks := analysis.callbacksRef()
	ctx := analysis.ctx
	if callbacks.ContextExpectNames[symbol] {
		return rstestExpectRootDirect
	}
	if callbacks.ContextReceivers[symbol] {
		return rstestExpectRootReceiver
	}

	if name, _, ok := resolveImportMetaRstestBinding(symbol); ok {
		if name == "expect" {
			return rstestExpectRootDirect
		}
		return rstestExpectRootNone
	}
	if isImportMetaRstestObjectAlias(symbol) {
		return rstestExpectRootReceiver
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
			return rstestExpectRootDirect
		}
		if testFramework.IsModuleNamespaceSymbol(symbol, module) {
			return rstestExpectRootReceiver
		}
	}
	return rstestExpectRootNone
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
