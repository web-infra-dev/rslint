package utils

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
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
	// Negated reports whether Chai's persistent `not` flag is active when this
	// matcher runs. Modifiers can appear between matchers in one assertion.
	Negated bool
}

// ParsedRstestExpectCall describes one Rstest expect call.
//
// Contract: RstestCallAnalysis.ParseExpectCall returns nil if and only if the node is not a
// Rstest expect call (which includes calls in the middle of a larger chain);
// once a node is recognized as an expect call the parse always succeeds and
// broken chains are reported through Reason instead of a nil result. This
// deliberately differs from JestCallAnalysis.ParseExpectCall, which returns nil
// for broken chains and exposes the reason only through
// ParseExpectCallWithReason.
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
	// FromTestContext reports whether the expect root is the local expect
	// provided by a test callback parameter, including renamed destructuring
	// and receiver forms such as ctx.expect.
	FromTestContext bool

	// rootInvoked and trailingMembers feed IsStaticRstestExpectCall, which
	// needs the syntactic shape rather than the chain semantics.
	rootInvoked     bool
	trailingMembers int
}

// PromiseModifierEntry includes Chai modifiers after matchers, but not members of an assertion's result.
func (parsed *ParsedRstestExpectCall) PromiseModifierEntry() *ParsedRstestFnMemberEntry {
	for i := range parsed.MemberEntries {
		entry := &parsed.MemberEntries[i]
		if entry.Call == parsed.Head {
			continue
		}
		if entry.Node == nil || isComputedDynamicMemberName(entry.Node) {
			return nil
		}
		if entry.Name == "then" || entry.Name == "catch" || entry.Name == "finally" {
			return nil
		}
		chain := classifyRstestExpectChainEntry(*entry, false)
		if chain.Kind == rstestExpectChainUnknown {
			return nil
		}
		if chain.Kind == rstestExpectChainModifier && (entry.Name == "resolves" || entry.Name == "rejects") {
			return entry
		}
	}
	return nil
}

func isRstestExpectCall(
	node *ast.Node,
	analysis *RstestCallAnalysis,
) bool {
	if node == nil || node.Kind != ast.KindCallExpression || FindTopMostCallExpression(node) != node {
		return false
	}
	expression := findTopMostRstestExpectExpressionOptions(node, false)
	entries := firstRstestExpectEntries(expression)
	if entries.count > 0 && entries.first != nil && entries.first.Kind == ast.KindIdentifier {
		return rstestExpectRootMatch(
			entries.first,
			entries.secondNode,
			entries.secondName,
			entries.count > 1,
			entries.secondDynamic,
			entries.firstInvoked,
			analysis,
		).ok
	}
	metaEntries := importMetaRstestMemberEntries(expression, false)
	return isImportMetaRstestMemberCall(metaEntries, "expect", false, analysis)
}

type rstestExpectEntries struct {
	first         *ast.Node
	secondNode    *ast.Node
	secondName    string
	count         uint8
	secondDynamic bool
	firstInvoked  bool
}

func (entries *rstestExpectEntries) append(name string, node *ast.Node, dynamic bool) {
	if name == "" {
		return
	}
	switch entries.count {
	case 0:
		entries.first = node
	case 1:
		entries.secondNode = node
		entries.secondName = name
		entries.secondDynamic = dynamic
	}
	if entries.count < 2 {
		entries.count++
	}
}

func firstRstestExpectEntries(node *ast.Node) rstestExpectEntries {
	if node == nil {
		return rstestExpectEntries{}
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return rstestExpectEntries{}
	}

	switch node.Kind {
	case ast.KindIdentifier:
		if node.AsIdentifier().Text == "" {
			return rstestExpectEntries{}
		}
		return rstestExpectEntries{
			first: node,
			count: 1,
		}
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		entries := firstRstestExpectEntries(property.Expression)
		if entries.count == 0 {
			return rstestExpectEntries{}
		}
		name := property.Name()
		if name == nil {
			return rstestExpectEntries{}
		}
		switch name.Kind {
		case ast.KindIdentifier:
			if text := name.AsIdentifier().Text; text != "" {
				entries.append(text, name, false)
				return entries
			}
		case ast.KindPrivateIdentifier:
			if text := name.AsPrivateIdentifier().Text; text != "" {
				entries.append(text, name, false)
				return entries
			}
		}
		return rstestExpectEntries{}
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		entries := firstRstestExpectEntries(element.Expression)
		if entries.count == 0 || element.ArgumentExpression == nil {
			return rstestExpectEntries{}
		}
		name := ast.SkipParentheses(element.ArgumentExpression)
		if name == nil {
			return rstestExpectEntries{}
		}
		switch name.Kind {
		case ast.KindIdentifier:
			if text := name.AsIdentifier().Text; text != "" {
				entries.append(text, name, true)
				return entries
			}
		case ast.KindStringLiteral:
			if text := name.AsStringLiteral().Text; text != "" {
				entries.append(text, name, false)
				return entries
			}
		case ast.KindNoSubstitutionTemplateLiteral:
			if text := name.AsNoSubstitutionTemplateLiteral().Text; text != "" {
				entries.append(text, name, false)
				return entries
			}
		}
		return rstestExpectEntries{}
	case ast.KindCallExpression:
		entries := firstRstestExpectEntries(node.AsCallExpression().Expression)
		if entries.count == 1 {
			entries.firstInvoked = true
		}
		return entries
	case ast.KindTaggedTemplateExpression:
		entries := firstRstestExpectEntries(node.AsTaggedTemplateExpression().Tag)
		if entries.count == 1 {
			entries.firstInvoked = true
		}
		return entries
	default:
		return rstestExpectEntries{}
	}
}

func parseRstestExpectCall(
	node *ast.Node,
	analysis *RstestCallAnalysis,
) *ParsedRstestExpectCall {
	return parseRstestExpectCallOptions(node, analysis, false)
}

func parseRstestExpectCallThroughTransparentExpressions(
	node *ast.Node,
	analysis *RstestCallAnalysis,
) *ParsedRstestExpectCall {
	return parseRstestExpectCallOptions(node, analysis, true)
}

func parseRstestExpectCallOptions(
	node *ast.Node,
	analysis *RstestCallAnalysis,
	throughTransparentExpressions bool,
) *ParsedRstestExpectCall {
	if node == nil || node.Kind != ast.KindCallExpression || findTopMostCallExpression(node, throughTransparentExpressions) != node {
		return nil
	}
	expression := findTopMostRstestExpectExpressionOptions(node, throughTransparentExpressions)
	var entries []testFramework.MemberEntry
	if throughTransparentExpressions {
		entries = testFramework.GetMemberEntriesThroughTransparentExpressions(expression)
	} else {
		entries = testFramework.GetMemberEntries(expression)
	}
	var match rstestExpectMatch
	if len(entries) > 0 {
		match = rstestExpectMemberMatch(entries, analysis, throughTransparentExpressions)
	} else if metaEntries := importMetaRstestMemberEntries(expression, throughTransparentExpressions); isImportMetaRstestMemberCall(metaEntries, "expect", throughTransparentExpressions, analysis) {
		entries = metaEntries
		match = rstestExpectMatch{index: 1, ok: true}
	}
	if !match.ok {
		return nil
	}
	return classifyRstestExpectCall(expression, entries, match, throughTransparentExpressions)
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
	return findTopMostCallExpression(node, false)
}

func findTopMostCallExpression(node *ast.Node, throughTransparentExpressions bool) *ast.Node {
	top := node
	current := node
	for parent := current.Parent; parent != nil; {
		if parent.Kind == ast.KindParenthesizedExpression {
			// Parentheses are always transparent to call-chain parsing.
		} else if expression, ok := internalUtils.TransparentExpression(parent); ok {
			if !throughTransparentExpressions || expression != current {
				return top
			}
		} else {
			switch parent.Kind {
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
				elem := parent.AsElementAccessExpression()
				if elem.Expression != current || !isSupportedRstestExpectElementKey(elem.ArgumentExpression, throughTransparentExpressions) {
					return top
				}
			default:
				return top
			}
		}
		current, parent = parent, parent.Parent
	}
	return top
}

func findTopMostRstestExpectExpressionOptions(node *ast.Node, throughTransparentExpressions bool) *ast.Node {
	top := node
	current := node
	for parent := current.Parent; parent != nil; {
		if parent.Kind == ast.KindParenthesizedExpression {
			// Parentheses are always transparent to expect-chain parsing.
		} else if expression, ok := internalUtils.TransparentExpression(parent); ok {
			if !throughTransparentExpressions || expression != current {
				return top
			}
		} else {
			switch parent.Kind {
			case ast.KindPropertyAccessExpression:
				if parent.AsPropertyAccessExpression().Expression != current {
					return top
				}
				top = parent
			case ast.KindElementAccessExpression:
				elem := parent.AsElementAccessExpression()
				if elem.Expression != current || !isSupportedRstestExpectElementKey(elem.ArgumentExpression, throughTransparentExpressions) {
					return top
				}
				top = parent
			default:
				return top
			}
		}
		current, parent = parent, parent.Parent
	}
	return top
}

func isSupportedRstestExpectElementKey(arg *ast.Node, throughTransparentExpressions bool) bool {
	if arg == nil {
		return false
	}
	var key *ast.Node
	if throughTransparentExpressions {
		key = internalUtils.SkipAssertionsAndParens(arg)
	} else {
		key = ast.SkipParentheses(arg)
	}
	if key == nil {
		return false
	}
	switch key.Kind {
	case ast.KindIdentifier:
		return key.AsIdentifier().Text != ""
	case ast.KindStringLiteral:
		return key.AsStringLiteral().Text != ""
	case ast.KindNoSubstitutionTemplateLiteral:
		return key.AsNoSubstitutionTemplateLiteral().Text != ""
	default:
		return false
	}
}

// rstestExpectMatch locates the expect member inside an assertion chain.
// Keeping the three results together stops call sites from having to remember
// the order of two adjacent booleans that mean different things.
type rstestExpectMatch struct {
	// index is where the expect member sits in the chain entries: 0 when the
	// root identifier itself is expect (imports, globals, destructured
	// context), 1 when the root is a receiver (namespace object,
	// import.meta.rstest, test context).
	index int
	// ok reports whether the chain is an Rstest expect chain at all.
	ok bool
	// fromTestContext reports whether the root is the local expect handed to a
	// test callback rather than a global or imported one.
	fromTestContext bool
}

// importMetaRstestMemberEntries extracts a call/member chain rooted at
// `import.meta.rstest` with the same accessor and transparent-wrapper rules as
// identifier-rooted chains.
func importMetaRstestMemberEntries(
	node *ast.Node,
	throughTransparentExpressions bool,
) []testFramework.MemberEntry {
	return testFramework.GetMemberEntriesFromRoot(
		node,
		throughTransparentExpressions,
		func(candidate *ast.Node) (testFramework.MemberEntry, bool) {
			return importMetaRstestRootEntryOptions(candidate, throughTransparentExpressions)
		},
	)
}

// resolveReceiverMemberName returns the runtime property name selected by a
// member access on an Rstest receiver (namespace import, require binding,
// test context, or `import.meta.rstest`). Static dot/string/template access
// returns its text directly; a computed identifier key (such as `r[key]`) is
// folded through the file's static string evaluator so `const key = 'expect'`
// resolves to `"expect"` while `const expect = 'test'` resolves to `"test"`.
func resolveReceiverMemberName(
	memberNode *ast.Node,
	defaultName string,
	dynamic bool,
	analysis *RstestCallAnalysis,
) (string, bool) {
	if !dynamic {
		return defaultName, defaultName != ""
	}
	if memberNode == nil || analysis == nil {
		return "", false
	}
	return analysis.staticStringEvaluator().EvalToString(memberNode)
}

func isImportMetaRstestMemberCall(
	entries []testFramework.MemberEntry,
	memberName string,
	throughTransparentExpressions bool,
	analysis *RstestCallAnalysis,
) bool {
	if len(entries) <= 1 || isReceiverEntryInvoked(entries, throughTransparentExpressions) {
		return false
	}
	resolvedName, ok := resolveReceiverMemberName(
		entries[1].Node,
		entries[1].Name,
		isComputedDynamicMemberName(entries[1].Node),
		analysis,
	)
	return ok && resolvedName == memberName
}

func isReceiverEntryInvoked(
	entries []testFramework.MemberEntry,
	throughTransparentExpressions bool,
) bool {
	if len(entries) < 2 {
		return false
	}
	if entries[0].Call != nil {
		return true
	}
	receiver, _ := testFramework.AccessorReceiverAndParent(&entries[1])
	for receiver != nil {
		receiver = ast.SkipParentheses(receiver)
		if receiver == nil {
			return false
		}
		if throughTransparentExpressions {
			if expr, ok := internalUtils.TransparentExpression(receiver); ok {
				receiver = expr
				continue
			}
		}
		return receiver.Kind == ast.KindCallExpression ||
			receiver.Kind == ast.KindTaggedTemplateExpression
	}
	return false
}

func importMetaRstestRootEntryOptions(
	node *ast.Node,
	throughTransparentExpressions bool,
) (testFramework.MemberEntry, bool) {
	// `@rstest/core` types `import.meta.rstest` as optional for in-source
	// testing, so TypeScript assertions on the `import.meta.rstest` receiver
	// itself (such as `import.meta.rstest!` or `(import.meta.rstest as ...)`)
	// are erased at runtime and still name the Rstest namespace. Unwrap them
	// here while keeping later wrapper boundaries in the assertion chain
	// controlled by `throughTransparentExpressions`.
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return testFramework.MemberEntry{}, false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		nameNode := property.Name()
		if nameNode == nil || nameNode.Text() != "rstest" ||
			!isImportMetaOptions(property.Expression, throughTransparentExpressions) {
			return testFramework.MemberEntry{}, false
		}
		return testFramework.MemberEntry{
			Name: "rstest",
			Node: nameNode,
		}, true
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		if element.ArgumentExpression == nil {
			return testFramework.MemberEntry{}, false
		}
		nameNode := ast.SkipParentheses(element.ArgumentExpression)
		name, ok := internalUtils.GetStaticStringLiteralValue(nameNode)
		if !ok || name != "rstest" ||
			!isImportMetaOptions(element.Expression, throughTransparentExpressions) {
			return testFramework.MemberEntry{}, false
		}
		return testFramework.MemberEntry{
			Name: name,
			Node: nameNode,
		}, true
	default:
		return testFramework.MemberEntry{}, false
	}
}

// rstestExpectMemberMatch resolves whether an identifier-rooted chain is an
// expect call and where the expect member sits in entries.
func rstestExpectMemberMatch(
	entries []testFramework.MemberEntry,
	analysis *RstestCallAnalysis,
	throughTransparentExpressions bool,
) rstestExpectMatch {
	if len(entries) == 0 || entries[0].Node == nil || entries[0].Node.Kind != ast.KindIdentifier {
		return rstestExpectMatch{}
	}
	var secondNode *ast.Node
	secondName := ""
	secondDynamic := false
	if len(entries) > 1 {
		secondNode = entries[1].Node
		secondName = entries[1].Name
		secondDynamic = isComputedDynamicMemberName(entries[1].Node)
	}
	return rstestExpectRootMatch(
		entries[0].Node,
		secondNode,
		secondName,
		len(entries) > 1,
		secondDynamic,
		isReceiverEntryInvoked(entries, throughTransparentExpressions),
		analysis,
	)
}

func rstestExpectRootMatch(
	root *ast.Node,
	secondNode *ast.Node,
	secondName string,
	hasSecond bool,
	secondDynamic bool,
	firstInvoked bool,
	analysis *RstestCallAnalysis,
) rstestExpectMatch {
	if root == nil || root.Kind != ast.KindIdentifier {
		return rstestExpectMatch{}
	}
	localName := root.AsIdentifier().Text
	ctx := analysis.ctx
	symbol := resolveRstestRootSymbol(ctx, root)
	if symbol == nil {
		return rstestExpectMatch{
			ok: localName == "expect" && !analysis.globalExpectWritten,
		}
	}
	rootInfo, ok := analysis.expectRoots[symbol]
	if !ok {
		rootInfo = classifyRstestExpectRoot(
			localName,
			root,
			symbol,
			analysis,
		)
		analysis.expectRoots[symbol] = rootInfo
	}
	switch rootInfo.Kind {
	case rstestExpectRootDirect:
		return rstestExpectMatch{ok: true, fromTestContext: rootInfo.FromTestContext}
	case rstestExpectRootReceiver:
		if firstInvoked || !hasSecond {
			return rstestExpectMatch{}
		}
		memberName, ok := resolveReceiverMemberName(secondNode, secondName, secondDynamic, analysis)
		return rstestExpectMatch{
			index:           1,
			ok:              ok && memberName == "expect",
			fromTestContext: rootInfo.FromTestContext,
		}
	default:
		return rstestExpectMatch{}
	}
}

func classifyRstestExpectCall(
	expression *ast.Node,
	entries []testFramework.MemberEntry,
	match rstestExpectMatch,
	throughTransparentExpressions bool,
) *ParsedRstestExpectCall {
	expectEntry := entries[match.index]
	expectCall := testFramework.DirectMemberCall(&expectEntry)
	if expectCall == nil {
		expectCall = expectEntry.Call
	}
	rest := entries[match.index+1:]
	parsed := &ParsedRstestExpectCall{
		Expression:      expression,
		Entry:           RstestExpectEntryCall,
		Head:            expectCall,
		rootInvoked:     expectCall != nil,
		trailingMembers: len(rest),
		FromTestContext: match.fromTestContext,
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
		factoryCall := testFramework.DirectMemberCall(&rest[0])
		if factoryCall == nil {
			factoryCall = rest[0].Call
		}
		if entry, ok := rstestExpectFactoryEntries[rest[0].Name]; ok &&
			!isComputedDynamicMemberName(rest[0].Node) &&
			factoryCall != nil {
			// expect.soft(x) / expect.poll(fn) / expect.element(loc): the factory
			// call replaces expect(x) as the assertion head and the chain resumes
			// after it.
			parsed.Entry = entry
			parsed.Head = factoryCall
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
	if reason == RstestExpectParseReasonMatcherNotFound {
		// If chain ascent stopped before an unsupported computed element
		// access that IS immediately invoked (such as `expect(1)['to' + 'Be'](1)`
		// or `expect(p).resolves[getMatcher()](1)`), treat that call as a
		// dynamic matcher call — unless the preceding chain already ended on an
		// unknown member (such as `expect(1).foo['to' + 'Be'](1)`), which is an
		// unknown modifier.
		if elemCall, elemArg := directTrailingDynamicMatcherCall(expression, throughTransparentExpressions); elemCall != nil {
			if endsWithUnknownChainMember(chain) {
				parsed.Reason = RstestExpectParseReasonModifierUnknown
				parsed.Expression = elemCall
			} else {
				parsed.Reason = RstestExpectParseReasonNone
				parsed.Expression = elemCall
				matchers = []ParsedRstestExpectMatcher{{
					Name: "",
					Entry: ParsedRstestFnMemberEntry{
						Node: elemArg,
						Call: elemCall,
					},
					Kind: RstestExpectMatcherCall,
				}}
			}
		} else if len(chain) > 0 &&
			isMemberAccessNode(expression) &&
			!hasComputedDynamicRstestExpectMember(chain) {
			parsed.Reason = RstestExpectParseReasonMatcherNotCalled
		}
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

func endsWithUnknownChainMember(chain []testFramework.MemberEntry) bool {
	if len(chain) == 0 {
		return false
	}
	last := chain[len(chain)-1]
	if last.Node == nil || isComputedDynamicMemberName(last.Node) {
		return true
	}
	return classifyRstestExpectChainEntry(last, false).Kind == rstestExpectChainUnknown
}

func directTrailingDynamicMatcherCall(expression *ast.Node, throughTransparentExpressions bool) (*ast.Node, *ast.Node) {
	outer := expression
	for outer != nil && outer.Parent != nil {
		parent := outer.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			outer = parent
			continue
		}
		if throughTransparentExpressions {
			if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == outer {
				outer = parent
				continue
			}
		}
		break
	}
	if outer == nil || outer.Parent == nil || outer.Parent.Kind != ast.KindElementAccessExpression {
		return nil, nil
	}
	elem := outer.Parent
	if elem.AsElementAccessExpression().Expression != outer {
		return nil, nil
	}
	cur := elem
	for cur.Parent != nil {
		parent := cur.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			cur = parent
			continue
		}
		if throughTransparentExpressions {
			if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == cur {
				cur = parent
				continue
			}
		}
		if parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == cur {
			arg := elem.AsElementAccessExpression().ArgumentExpression
			if arg != nil {
				arg = ast.SkipParentheses(arg)
			}
			return parent, arg
		}
		return nil, nil
	}
	return nil, nil
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
			// The runtime value of a computed identifier key is unknowable, so
			// the member is neither a modifier nor a matcher. Classifying it
			// unknown rather than failing the whole parse keeps a statically
			// resolved matcher earlier in the chain visible, which is what
			// eslint-plugin-vitest and eslint-plugin-jest both do.
			chains[i] = rstestExpectChainEntry{
				Entry: member,
				Kind:  rstestExpectChainUnknown,
			}
			continue
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
		return modifiers, nil, RstestExpectParseReasonMatcherNotFound
	}

	matchers := make([]ParsedRstestExpectMatcher, 0, len(chains)-matcherIndex)
	negated := notCount > 0
	for i := matcherIndex; i < len(chains); i++ {
		chain := chains[i]
		switch chain.Kind {
		case rstestExpectChainMatcher:
			matchers = append(matchers, ParsedRstestExpectMatcher{
				Name:    chain.Entry.Name,
				Entry:   chain.Entry,
				Kind:    chain.MatcherKind,
				Negated: negated,
			})
		case rstestExpectChainLanguage:
			if chain.Entry.Call != nil {
				return nil, nil, RstestExpectParseReasonModifierUnknown
			}
		case rstestExpectChainModifier:
			// Chai permits modifiers between assertions in a multi-matcher
			// chain, e.g. .a("string").that.does.not.contain("x").
			if chain.Entry.Name == "not" {
				negated = true
			}
		case rstestExpectChainUnknown:
			// A member the chain grammar does not recognise ends the
			// assertion: everything after it reads a property of the
			// assertion's own result, e.g. expect(a).toBe(1).message or
			// expect(a).toBe(1)[key]. The assertion itself already parsed, so
			// the chain is truncated here instead of failing.
			return modifiers, matchers, RstestExpectParseReasonNone
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
// computed element access with an identifier key, e.g. expect(x)[matcherName]
// or expect(x)[matcherName as any]. GetMemberEntries records the identifier's
// text as the member name, but the runtime value is unknowable statically.
func isComputedDynamicMemberName(node *ast.Node) bool {
	return testFramework.IsComputedIdentifierAccessor(node)
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

type rstestExpectRoot struct {
	Kind            rstestExpectRootKind
	FromTestContext bool
}

func classifyRstestExpectRoot(
	localName string,
	root *ast.Node,
	symbol *ast.Symbol,
	analysis *RstestCallAnalysis,
) rstestExpectRoot {
	callbacks := analysis.callbacksRef()
	ctx := analysis.ctx
	initializations := 0
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindVariableDeclaration, ast.KindBindingElement:
			if internalUtils.IsVariableWriteReference(declaration.Name()) {
				initializations++
			}
		case ast.KindParameter, ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration, ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			initializations++
		}
		if initializations > 1 {
			return rstestExpectRoot{Kind: rstestExpectRootNone}
		}
	}
	if ctx.Refs != nil {
		for _, reference := range ctx.Refs.References(symbol) {
			if internalUtils.IsWriteReference(reference) {
				return rstestExpectRoot{Kind: rstestExpectRootNone}
			}
		}
	}
	if callbacks.ContextExpectNames[symbol] {
		return rstestExpectRoot{Kind: rstestExpectRootDirect, FromTestContext: true}
	}
	if callbacks.ContextReceivers[symbol] {
		return rstestExpectRoot{Kind: rstestExpectRootReceiver, FromTestContext: true}
	}

	if name, _, ok := resolveImportMetaRstestBinding(symbol); ok {
		if name == "expect" {
			return rstestExpectRoot{Kind: rstestExpectRootDirect}
		}
		return rstestExpectRoot{Kind: rstestExpectRootNone}
	}
	if isImportMetaRstestObjectAlias(symbol) {
		return rstestExpectRoot{Kind: rstestExpectRootReceiver}
	}

	name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
		localName,
		root,
		symbol,
		ctx.SourceFile,
		RstestAllImportModules,
	)
	if name == "expect" {
		return rstestExpectRoot{Kind: rstestExpectRootDirect}
	}
	if testFramework.IsModuleNamespaceSymbolModules(symbol, RstestAllImportModules) {
		return rstestExpectRoot{Kind: rstestExpectRootReceiver}
	}
	return rstestExpectRoot{Kind: rstestExpectRootNone}
}

func isImportMetaRstestObjectAlias(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		initializer := internalUtils.SkipAssertionsAndParens(declaration.AsVariableDeclaration().Initializer)
		if isImportMetaRstest(initializer) {
			return true
		}
	}
	return false
}
