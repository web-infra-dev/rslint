package prefer_called_exactly_once_with

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Rstest ships both matcher spellings on one Chai Assertion prototype:
// @vitest/expect's ChaiStyleAssertions registers calledOnce / calledWith /
// calledOnceWith with defProperty and defMethod, each delegating to the jest
// method of the same meaning, and packages/core/src/runtime/api/expect.ts
// installs JestChaiExpect and ChaiStyleAssertions side by side. Both spellings
// therefore assert on the same spy and merge the same way, so the rule pairs
// them regardless of which spelling each assertion uses.
//
// combinedMatchers maps every "called with these arguments" matcher to the
// matcher that also pins the call count. The merged assertion keeps the
// argument-side call, so its spelling decides the spelling of the result.
var combinedMatchers = map[string]string{
	"toHaveBeenCalledWith": "toHaveBeenCalledExactlyOnceWith",
	"calledWith":           "calledOnceWith",
}

// onceMatchers lists the "called exactly one time" matchers, whose assertion
// the merge folds into the argument-side call and then deletes.
var onceMatchers = map[string]bool{
	"toHaveBeenCalledOnce": true,
	"calledOnce":           true,
}

// mockResetMethods are the per-mock resets: each clears the call history of
// the mock it is called on, so only a call whose receiver is the asserted
// target splits the pair.
var mockResetMethods = map[string]bool{
	"mockClear":   true,
	"mockReset":   true,
	"mockRestore": true,
}

// globalMockResetMethods are the suite-wide resets. They clear every mock's
// call history, so one call anywhere between the two assertions splits them
// just as a receiver-matched reset does and no receiver has to match. Matching
// on the method name alone also accepts an unrelated object's clearAllMocks;
// that only ever suppresses a report, which is the safe direction. Upstream
// knows none of the three and merges across them.
var globalMockResetMethods = map[string]bool{
	"clearAllMocks":   true,
	"resetAllMocks":   true,
	"restoreAllMocks": true,
}

// mockResetKind classifies a call as a reset of one mock, a reset of every
// mock, or neither.
type mockResetKind uint8

const (
	resetNone mockResetKind = iota
	resetTarget
	resetGlobal
)

type assertionRole uint8

const (
	roleOnce assertionRole = iota
	roleWith
)

// chainAssertion is one matcher this rule can act on, found in an assertion
// chain.
type chainAssertion struct {
	role assertionRole
	// matcher is the matcher name as written, which the report message echoes
	// so a Chai-style chain is not described with jest-style names.
	matcher string
	// combined is the merged matcher name; only meaningful for roleWith.
	combined string
	node     *ast.Node
}

// mergeCandidate is one expression statement this rule can act on. A Chai
// chain may assert several times, so a statement carries either one matcher of
// interest, which merges with a matching statement elsewhere in the block, or
// one of each role, which already states both halves and merges with itself.
type mergeCandidate struct {
	hits      []chainAssertion
	statement *ast.Node
	// position is the trimmed start offset of statement, used to order the
	// pair and to bound the mock-reset search between them.
	position int
	// pairKey groups the assertions that may merge; see pairKey().
	pairKey string
	// targetNode is the first factory argument. It only identifies the mock for
	// the reset barrier and never pairs assertions, and it is compared
	// structurally so an equivalent spelling is still recognised.
	targetNode *ast.Node
	// fixable is false when the chain asserts more than the rule understands,
	// so the merge is reported but left to the author: folding a statement away
	// would drop the assertions sharing its chain, and rewriting the chain in
	// place would have to splice out a segment. Both are more than an autofix
	// should attempt, and a wrong one silently deletes a live assertion.
	fixable bool
}

func (candidate *mergeCandidate) isSelfContained() bool {
	return len(candidate.hits) == 2
}

func preferCalledExactlyOnceWithMessage(once, with, combined string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "preferCalledExactlyOnceWith",
		Description: fmt.Sprintf(
			"Using `%s` and `%s` on the same target; prefer `%s` instead.",
			once, with, combined,
		),
	}
}

func nodeText(sourceFile *ast.SourceFile, node *ast.Node) (string, bool) {
	if sourceFile == nil || node == nil {
		return "", false
	}
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()
	if r.Pos() < 0 || r.End() > len(text) || r.Pos() >= r.End() {
		return "", false
	}
	return text[r.Pos():r.End()], true
}

// transparentWrappers are the wrappers that change neither the value an
// expression denotes nor whether evaluating it can be observed: parentheses
// and the type-only assertions (`as`, `satisfies`, `!`).
const transparentWrappers = ast.OEKParentheses | ast.OEKAssertions

// unwrapExpression strips those wrappers, so `obj /* keep */ .fn` and
// `(obj.fn as Mock)` reduce to the same shape as `obj.fn`.
func unwrapExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, transparentWrappers)
}

// unwrapMocked additionally looks through `mocked(x)` and `rstest.mocked(x)`.
// mocked is identity at runtime — `mocked: ((item: any) => item)` in rstest's
// runtime/api/utilities.ts — and exists only to narrow the type, so
// `rstest.mocked(x).mockClear()` really does reset `x` and the two spellings
// have to compare equal. Only the reset barrier unwraps it: nothing here
// promises the call is that mocked, and a wrong guess costs a missed report,
// never a bad fix.
func unwrapMocked(node *ast.Node) *ast.Node {
	for {
		unwrapped := unwrapExpression(node)
		if unwrapped == nil || unwrapped.Kind != ast.KindCallExpression {
			return unwrapped
		}
		call := unwrapped.AsCallExpression()
		arguments := unwrapped.Arguments()
		if call == nil || len(arguments) != 1 {
			return unwrapped
		}
		entries := test_framework.GetMemberEntries(ast.SkipParentheses(call.Expression))
		if len(entries) == 0 {
			return unwrapped
		}
		callee := entries[len(entries)-1]
		if callee.Name != "mocked" || test_framework.IsComputedIdentifierAccessor(callee.Node) {
			return unwrapped
		}
		node = arguments[0]
	}
}

// isStableExpression reports whether evaluating the expression twice is
// indistinguishable from evaluating it once: no call, no `new`, no `await`, no
// assignment or update. Property and element access are accepted — a getter
// could in principle observe the second read, but treating `obj.fn` as
// unstable would refuse a fix for the shape the rule mostly exists for.
// Element access is limited to literal keys, so `mocks[i++]` stays out.
func isStableExpression(node *ast.Node) bool {
	node = unwrapExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		return access != nil && isStableExpression(access.Expression)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if access == nil || !isStableExpression(access.Expression) {
			return false
		}
		argument := unwrapExpression(access.ArgumentExpression)
		return argument != nil &&
			(argument.Kind == ast.KindStringLiteral ||
				argument.Kind == ast.KindNumericLiteral ||
				argument.Kind == ast.KindNoSubstitutionTemplateLiteral)
	default:
		return false
	}
}

// sameTargetExpression compares two expressions structurally, so a receiver
// that differs from the target only in parentheses, type assertions or trivia
// still counts as the same mock. Comparing the source text instead would let
// `obj /* keep */ .fn.mockClear()` slip past the reset barrier and hand the
// author a merge across two call histories.
func sameTargetExpression(left, right *ast.Node) bool {
	left, right = unwrapExpression(left), unwrapExpression(right)
	if left == nil || right == nil || left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindIdentifier:
		return left.AsIdentifier().Text == right.AsIdentifier().Text
	case ast.KindThisKeyword, ast.KindSuperKeyword, ast.KindNullKeyword,
		ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return left.Text() == right.Text()
	case ast.KindPropertyAccessExpression:
		leftAccess, rightAccess := left.AsPropertyAccessExpression(), right.AsPropertyAccessExpression()
		if leftAccess == nil || rightAccess == nil ||
			leftAccess.Name() == nil || rightAccess.Name() == nil ||
			leftAccess.Name().Text() != rightAccess.Name().Text() {
			return false
		}
		return sameTargetExpression(leftAccess.Expression, rightAccess.Expression)
	case ast.KindElementAccessExpression:
		leftAccess, rightAccess := left.AsElementAccessExpression(), right.AsElementAccessExpression()
		if leftAccess == nil || rightAccess == nil ||
			!sameTargetExpression(leftAccess.ArgumentExpression, rightAccess.ArgumentExpression) {
			return false
		}
		return sameTargetExpression(leftAccess.Expression, rightAccess.Expression)
	default:
		return false
	}
}

// assertionRootCall returns the call ParseExpectCall must be handed for an
// expression statement, and whether the statement awaits it. Method-style
// assertions are already that call, while a Chai property assertion such as
// expect(x).to.have.been.calledOnce ends on a member access whose chain is
// rooted in the expect(x) call.
//
// A promise modifier makes the assertion awaitable, and such an assertion is
// normally awaited, so a rule that stopped at the await would only ever see
// the form that is already wrong. Upstream stops there.
func assertionRootCall(expression *ast.Node) (*ast.Node, bool) {
	node := ast.SkipParentheses(expression)
	awaited := false
	for node != nil {
		switch node.Kind {
		case ast.KindCallExpression:
			return node, awaited
		case ast.KindAwaitExpression:
			awaited = true
			node = ast.SkipParentheses(node.AsAwaitExpression().Expression)
		case ast.KindPropertyAccessExpression:
			node = ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			node = ast.SkipParentheses(node.AsElementAccessExpression().Expression)
		default:
			return nil, false
		}
	}
	return nil, false
}

func mergeCandidateForStatement(
	analysis *rstestUtils.RstestCallAnalysis,
	sourceFile *ast.SourceFile,
	statement *ast.Node,
) *mergeCandidate {
	if statement == nil || statement.Kind != ast.KindExpressionStatement {
		return nil
	}
	expressionStatement := statement.AsExpressionStatement()
	if expressionStatement == nil {
		return nil
	}
	rootCall, awaited := assertionRootCall(expressionStatement.Expression)
	if rootCall == nil {
		return nil
	}
	parsed := analysis.ParseExpectCall(rootCall)
	if parsed == nil {
		return nil
	}
	// `not` never merges, even on both assertions: "not called once" and "not
	// called with these arguments" is `¬once ∧ ¬with`, while the combined
	// matcher negated is `¬(once ∧ with)`. Upstream drops `not` chains for the
	// same reason.
	//
	// Modifiers only holds what precedes the first matcher, and Chai permits a
	// modifier between matchers, so `calledOnce.and.not.calledWith('a')` would
	// otherwise reach the merge with an empty Modifiers and both matchers
	// present. Scanning the whole chain is what actually disqualifies it.
	if slices.Contains(parsed.Members, "not") {
		return nil
	}
	if parsed.Head == nil || parsed.Head.Kind != ast.KindCallExpression {
		return nil
	}
	hits := chainAssertions(parsed)
	if hits == nil {
		return nil
	}

	headText, ok := nodeText(sourceFile, parsed.Head)
	if !ok {
		return nil
	}
	arguments := parsed.Head.Arguments()
	if len(arguments) == 0 {
		return nil
	}
	if arguments[0] == nil {
		return nil
	}
	// The fix keeps one `expect(...)` call and deletes the other, so every
	// argument of the surviving call is evaluated once instead of twice. Equal
	// source text does not make that safe: `expect(getMock())` may return a
	// different mock each time, and dropping an evaluation drops whatever the
	// expression did. Such a pair is still worth reporting, but the merge is
	// the author's to make.
	fixable := len(parsed.Matchers) == 1
	for _, argument := range arguments {
		if !isStableExpression(argument) {
			fixable = false
			break
		}
	}
	candidate := &mergeCandidate{hits: hits, fixable: fixable}

	candidate.statement = statement
	candidate.position = internalUtils.TrimNodeTextRange(sourceFile, statement).Pos()
	candidate.pairKey = pairKey(headText, parsed.Modifiers, awaited)
	candidate.targetNode = arguments[0]
	return candidate
}

// pairKey composes what has to match before two assertions can be treated as
// two halves of one claim: they must assert on the same value, in the same
// way, in the same execution context. The fix keeps one statement and deletes
// the other, so anything that differs here would be silently dropped — an
// `await` most of all, whose loss turns a working assertion into a floating
// promise whose failure escapes as an unhandled rejection.
//
// Upstream gets the first two from the text preceding the matcher, and misses
// the third because it never looks through an await. That text also carries
// Chai's language chains, which assert nothing, so it would keep
// `expect(x).to.have.been.calledOnce` from pairing with
// `expect(x).calledWith('a')`. Composing the key from the parse instead keeps
// what changes meaning and drops the sugar.
//
// The separator cannot occur in source text, so no factory text can collide
// with a modified chain.
func pairKey(headText string, modifiers []string, awaited bool) string {
	key := headText + "\x00" + strings.Join(modifiers, ".")
	if awaited {
		return key + "\x00await"
	}
	return key
}

// chainAssertions picks the matchers this rule acts on out of an assertion
// chain, in source order. It returns one matcher for a statement that states
// half of the merge, one of each role for a Chai chain that already states
// both halves, and nil for anything else — including a chain that repeats a
// role, where which occurrence to merge is not decidable.
func chainAssertions(parsed *rstestUtils.ParsedRstestExpectCall) []chainAssertion {
	var hits []chainAssertion
	seen := map[assertionRole]bool{}
	for _, matcher := range parsed.Matchers {
		hit := chainAssertion{matcher: matcher.Name, node: matcher.Entry.Node}
		if combined, ok := combinedMatchers[matcher.Name]; ok {
			hit.role = roleWith
			hit.combined = combined
		} else if onceMatchers[matcher.Name] {
			hit.role = roleOnce
		} else {
			continue
		}
		if seen[hit.role] || hit.node == nil {
			return nil
		}
		seen[hit.role] = true
		hits = append(hits, hit)
	}
	if len(hits) == 0 {
		return nil
	}
	return hits
}

// mockResetCall classifies a call as a reset, and for a per-mock reset also
// returns the object it resets.
func mockResetCall(callNode *ast.Node) (mockResetKind, *ast.Node) {
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return resetNone, nil
	}
	callee := ast.SkipParentheses(callNode.AsCallExpression().Expression)
	if callee == nil {
		return resetNone, nil
	}
	entries := test_framework.GetMemberEntries(callee)
	if len(entries) == 0 {
		return resetNone, nil
	}
	method := entries[len(entries)-1]
	if test_framework.IsComputedIdentifierAccessor(method.Node) {
		return resetNone, nil
	}
	if globalMockResetMethods[method.Name] {
		return resetGlobal, nil
	}
	if !mockResetMethods[method.Name] {
		return resetNone, nil
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		return resetTarget, unwrapMocked(callee.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return resetTarget, unwrapMocked(callee.AsElementAccessExpression().Expression)
	default:
		return resetNone, nil
	}
}

// hasMockResetBetween reports a reset between the two assertions, which splits
// them into claims about two different call histories: either a reset of the
// shared target or a suite-wide reset, which clears that target along with
// everything else. The search descends into the statements that sit between
// the pair, so a reset nested in an if or a loop still blocks the merge;
// upstream only scans the sibling statements and merges across such a reset.
func hasMockResetBetween(
	sourceFile *ast.SourceFile,
	statements []*ast.Node,
	first, second *mergeCandidate,
) bool {
	// The pair was already keyed on identical `expect(...)` text, which assumes
	// both spellings denote the same mock. A target the structural comparison
	// cannot decide — `expect(getMock())`, say — must not silently disable the
	// barrier as well, or `getMock().mockClear()` between the two assertions
	// would pass unnoticed. For such a target any reset counts, whoever its
	// receiver is: the report costs a false negative when the reset was aimed
	// at some other mock, which is the direction to err in.
	target := unwrapMocked(first.targetNode)
	targetIsComparable := sameTargetExpression(target, unwrapMocked(second.targetNode))
	minPosition, maxPosition := first.position, second.position
	if minPosition > maxPosition {
		minPosition, maxPosition = maxPosition, minPosition
	}

	found := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		switch kind, receiver := mockResetCall(node); kind {
		case resetGlobal:
			found = true
			return true
		case resetTarget:
			if !targetIsComparable || sameTargetExpression(receiver, target) {
				found = true
				return true
			}
		}
		return node.ForEachChild(visit)
	}

	for _, statement := range statements {
		if statement == nil {
			continue
		}
		position := internalUtils.TrimNodeTextRange(sourceFile, statement).Pos()
		if position <= minPosition || position >= maxPosition {
			continue
		}
		if visit(statement); found {
			return true
		}
	}
	return false
}

// statementRemovalRange is the range the fix deletes for the folded assertion.
// A statement that owns its line is removed with that line's indentation and
// line break, because leaving them behind turns the merge into a stray blank
// line. When anything else shares the line — another statement, a comment —
// only the statement itself goes, so the fix never swallows code it did not
// look at. Upstream always deletes whole lines and does swallow it.
func statementRemovalRange(sourceFile *ast.SourceFile, statement *ast.Node) core.TextRange {
	statementRange := internalUtils.TrimNodeTextRange(sourceFile, statement)
	text := sourceFile.Text()
	start, end := statementRange.Pos(), statementRange.End()

	lineStart := start
	for lineStart > 0 && text[lineStart-1] != '\n' {
		if text[lineStart-1] != ' ' && text[lineStart-1] != '\t' {
			return statementRange
		}
		lineStart--
	}

	lineEnd := end
	for lineEnd < len(text) && text[lineEnd] != '\n' {
		if text[lineEnd] != ' ' && text[lineEnd] != '\t' && text[lineEnd] != '\r' {
			return statementRange
		}
		lineEnd++
	}
	if lineEnd < len(text) {
		lineEnd++
	}
	return core.NewTextRange(lineStart, lineEnd)
}

func statementsForBlock(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindSourceFile:
		sourceFile := node.AsSourceFile()
		if sourceFile != nil && sourceFile.Statements != nil {
			return sourceFile.Statements.Nodes
		}
	case ast.KindBlock:
		block := node.AsBlock()
		if block != nil && block.Statements != nil {
			return block.Statements.Nodes
		}
	}
	return nil
}

// mergeMessage names the two matchers in the order the merge reads — the
// call-count half first — whichever order they appear in.
func mergeMessage(first, second chainAssertion) (rule.RuleMessage, chainAssertion) {
	once, with := first, second
	if once.role != roleOnce {
		once, with = second, first
	}
	return preferCalledExactlyOnceWithMessage(once.matcher, with.matcher, with.combined), with
}

// reportSelfMerge handles a Chai chain that already asserts both halves, e.g.
// expect(x).to.have.been.calledOnce.and.calledWith('a'). There is nothing to
// pair it with and no statement to delete; the merge rewrites the chain in
// place, which this rule leaves to the author.
func reportSelfMerge(ctx rule.RuleContext, candidate *mergeCandidate) {
	message, _ := mergeMessage(candidate.hits[0], candidate.hits[1])
	ctx.ReportNode(candidate.hits[1].node, message)
}

func reportPair(ctx rule.RuleContext, statements []*ast.Node, candidates []*mergeCandidate) {
	first, second := candidates[0], candidates[1]
	if first.hits[0].role == second.hits[0].role ||
		hasMockResetBetween(ctx.SourceFile, statements, first, second) {
		return
	}
	message, with := mergeMessage(first.hits[0], second.hits[0])
	once := first
	if once.hits[0].role != roleOnce {
		once = second
	}

	// The later assertion carries the report, as upstream does; the earlier one
	// alone would point at code that reads fine until the second one shows up.
	reportNode := second.hits[0].node
	if !first.fixable || !second.fixable {
		ctx.ReportNode(reportNode, message)
		return
	}
	matcherRange, replacement, ok := test_framework.AccessorReplacement(
		ctx.SourceFile,
		with.node,
		with.combined,
	)
	if !ok {
		ctx.ReportNode(reportNode, message)
		return
	}
	removeRange := statementRemovalRange(ctx.SourceFile, once.statement)
	if removeRange.Overlaps(matcherRange) {
		ctx.ReportNode(reportNode, message)
		return
	}
	ctx.ReportNodeWithFixes(
		reportNode,
		message,
		rule.RuleFixReplaceRange(matcherRange, replacement),
		rule.RuleFixRemoveRange(removeRange),
	)
}

// pendingReport is one resolved merge, held until the whole block is analysed
// so reports can be emitted in source order.
type pendingReport struct {
	position int
	report   func()
}

func checkBlock(
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
	node *ast.Node,
) {
	statements := statementsForBlock(node)
	if len(statements) == 0 {
		return
	}

	// Merges are resolved per target but reported in source order, so the
	// diagnostic list reads down the file and never depends on map iteration
	// order. Each entry is keyed by the statement that carries its report.
	var pending []pendingReport
	var pairKeys []string
	byPairKey := map[string][]*mergeCandidate{}
	for _, statement := range statements {
		candidate := mergeCandidateForStatement(analysis, ctx.SourceFile, statement)
		if candidate == nil {
			continue
		}
		// A chain that states both halves needs no partner, and counting it
		// among the candidates would only suppress it.
		if candidate.isSelfContained() {
			pending = append(pending, pendingReport{
				position: candidate.position,
				report:   func() { reportSelfMerge(ctx, candidate) },
			})
			continue
		}
		if _, seen := byPairKey[candidate.pairKey]; !seen {
			pairKeys = append(pairKeys, candidate.pairKey)
		}
		byPairKey[candidate.pairKey] = append(byPairKey[candidate.pairKey], candidate)
	}

	for _, pairKey := range pairKeys {
		// A third assertion on the same target makes the merge ambiguous, so
		// upstream reports only an exact pair and this rule follows.
		candidates := byPairKey[pairKey]
		if len(candidates) != 2 {
			continue
		}
		pending = append(pending, pendingReport{
			position: candidates[1].position,
			report:   func() { reportPair(ctx, statements, candidates) },
		})
	}

	slices.SortFunc(pending, func(a, b pendingReport) int { return a.position - b.position })
	for _, entry := range pending {
		entry.report()
	}
}

var PreferCalledExactlyOnceWithRule = rule.Rule{
	Name:   "rstest/prefer-called-exactly-once-with",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		// The linter traverses SourceFile children but does not dispatch a
		// SourceFile listener, so inspect the top-level statement list eagerly.
		if ctx.SourceFile != nil {
			checkBlock(ctx, analysis, ctx.SourceFile.AsNode())
		}
		check := func(node *ast.Node) { checkBlock(ctx, analysis, node) }
		return rule.RuleListeners{
			ast.KindBlock: check,
		}
	},
}
