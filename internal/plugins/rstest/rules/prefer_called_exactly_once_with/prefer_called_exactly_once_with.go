package prefer_called_exactly_once_with

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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

// invokingMatchers execute the value they assert on, so an assertion using one
// can run arbitrary code — `expect(callTheMock).toThrow()` calls the mock. The
// set is closed: it is fixed by the matchers @vitest/expect installs, unlike
// the open question of which function eventually reaches the mock.
var invokingMatchers = map[string]bool{
	"toThrow":      true,
	"toThrowError": true,
	"toSatisfy":    true,
	"throw":        true,
	"throws":       true,
	"Throw":        true,
}

// suspendingModifiers hand control to the microtask queue, which lets whatever
// is already pending run before the assertion resumes.
var suspendingModifiers = map[string]bool{
	"resolves": true,
	"rejects":  true,
}

// controlTransfer reports whether a node hands control to code this rule
// cannot see. Anything that runs user code qualifies, as does any write: a
// write needs no target check, because a statement between the assertions has
// no business assigning at all.
func controlTransfer(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindAwaitExpression, ast.KindYieldExpression, ast.KindDeleteExpression:
		return true
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		return binary != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		return unary != nil &&
			(unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken)
	case ast.KindPostfixUnaryExpression:
		unary := node.AsPostfixUnaryExpression()
		return unary != nil &&
			(unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken)
	default:
		return false
	}
}

// expectRootName is the local name the assertion factory is reached through,
// so `expect.any(String)` can be told apart from an arbitrary call. Empty when
// the root is not a plain member chain, in which case no helper is allowed.
func expectRootName(head *ast.Node) string {
	if head == nil || head.Kind != ast.KindCallExpression {
		return ""
	}
	entries := test_framework.GetMemberEntries(ast.SkipParentheses(head.AsCallExpression().Expression))
	if len(entries) == 0 {
		return ""
	}
	return entries[0].Name
}

// isExpectHelperCall reports a call such as expect.any(String) or
// expect.objectContaining({}), reached through the same root as the assertion
// itself. These are framework functions with known semantics: they build a
// value to compare against and run nothing of the author's.
func isExpectHelperCall(node *ast.Node, rootName string) bool {
	if rootName == "" || node.Kind != ast.KindCallExpression {
		return false
	}
	entries := test_framework.GetMemberEntries(ast.SkipParentheses(node.AsCallExpression().Expression))
	return len(entries) == 2 && entries[0].Name == rootName && entries[0].Call == nil
}

// betweenScanner decides whether the statements separating a pair leave its
// merge intact. It carries what that needs: the parse cache, the checker
// through ctx, and the names the two assertions read.
type betweenScanner struct {
	ctx            rule.RuleContext
	analysis       *rstestUtils.RstestCallAnalysis
	assertionNames map[string]bool
}

// isCallable reports whether an argument could be invoked by whoever receives
// it. Anything the checker cannot resolve counts as callable, so an unknown
// argument keeps its call out of the inert set.
func (scanner *betweenScanner) isCallable(node *ast.Node) bool {
	node = unwrapExpression(node)
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression:
		return true
	}
	if scanner.ctx.TypeChecker == nil {
		return true
	}
	argumentType := scanner.ctx.TypeChecker.GetTypeAtLocation(node)
	if argumentType == nil ||
		internalUtils.IsTypeFlagSet(argumentType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		// `any` has no call signatures and answers nothing about the value, so
		// an argument the checker could not pin down stays callable.
		return true
	}
	return len(internalUtils.GetCallSignatures(scanner.ctx.TypeChecker, argumentType)) > 0 ||
		len(internalUtils.GetConstructSignatures(scanner.ctx.TypeChecker, argumentType)) > 0
}

// isLibraryCall reports a call that runs nothing of the author's: its callee is
// declared in TypeScript's default library, and nothing callable is handed to
// it, so the library function has nothing of the author's to invoke.
// `console.log('checkpoint')` and `JSON.stringify(value)` qualify;
// `items.forEach(cb)`, `setTimeout(cb)` and `promise.then(cb)` do not, and
// neither does a `console` the author declared themselves, which resolves to
// their own binding rather than the library's. Without a checker nothing
// qualifies, which leaves the pair unreported — the safe direction.
func (scanner *betweenScanner) isLibraryCall(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression || scanner.ctx.TypeChecker == nil {
		return false
	}
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	symbol := scanner.ctx.TypeChecker.GetSymbolAtLocation(ast.SkipParentheses(call.Expression))
	if symbol == nil || !internalUtils.IsSymbolFromDefaultLibrary(scanner.ctx.Program(), symbol) {
		return false
	}
	for _, argument := range node.Arguments() {
		if scanner.isCallable(argument) {
			return false
		}
	}
	return true
}

// inertSubtree reports whether a subtree transfers control nowhere, treating
// the given framework calls, `<root>.helper(...)` calls and default-library
// calls as known-safe. Their arguments are still walked, so a call nested in
// one of them still counts.
func (scanner *betweenScanner) inertSubtree(node *ast.Node, frameworkCalls map[*ast.Node]bool, rootName string) bool {
	inert := true
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if !frameworkCalls[node] && !isExpectHelperCall(node, rootName) &&
			!scanner.isLibraryCall(node) && controlTransfer(node) {
			inert = false
			return true
		}
		return node.ForEachChild(visit)
	}
	visit(node)
	return inert
}

// isInertDeclaration accepts a declaration whose initializers transfer control
// nowhere — `const hoge = 'foo';` between the two assertions changes nothing
// about either. A `var` is additionally required to declare names neither
// assertion reads: it is hoisted, so unlike `const` and `let` it can rebind a
// name the first assertion already used without a temporal dead zone error.
func (scanner *betweenScanner) isInertDeclaration(statement *ast.Node) bool {
	if statement == nil || statement.Kind != ast.KindVariableStatement {
		return false
	}
	variableStatement := statement.AsVariableStatement()
	if variableStatement == nil || variableStatement.DeclarationList == nil {
		return false
	}
	declarationList := variableStatement.DeclarationList
	list := declarationList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return false
	}
	hoisted := declarationList.Flags&(ast.NodeFlagsConst|ast.NodeFlagsLet) == 0
	for _, declaration := range list.Declarations.Nodes {
		if declaration == nil {
			return false
		}
		if hoisted && scanner.shadowsAssertion(declaration.Name()) {
			return false
		}
		initializer := declaration.Initializer()
		if initializer == nil {
			continue
		}
		if !scanner.inertSubtree(initializer, nil, "") {
			return false
		}
	}
	return true
}

// shadowsAssertion reports whether a binding name occurs in either assertion,
// which is the only way a hoisted declaration can change what they assert on.
func (scanner *betweenScanner) shadowsAssertion(name *ast.Node) bool {
	shadows := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindIdentifier && scanner.assertionNames[node.AsIdentifier().Text] {
			shadows = true
			return true
		}
		return node.ForEachChild(visit)
	}
	if name != nil {
		visit(name)
	}
	return shadows
}

// isInertStatement reports whether a statement can sit between the two
// assertions without invalidating their merge. The merge claims both halves
// describe one call history, so nothing in between may call the asserted mock
// or rebind it. Neither question is decidable from the syntax, so the rule
// answers a narrower one it can decide: the statement must be an assertion
// that transfers control nowhere.
//
// That means an expect chain whose factory, matcher and default-library calls
// are the only calls in it — their semantics are known, they read the spy's
// record and compare values — with no other call, no write, and no await
// anywhere in the arguments. A matcher that executes its subject, and a
// modifier that awaits one, are excluded by name.
//
// What survives is the common shape: assertions on other mocks, or on other
// aspects of this one, grouped between the two halves. What does not is
// anything that could reach the target, and anything the rule cannot resolve.
//
// A property read is the one thing accepted without proof, since a getter
// could run code. The rule already accepts that when it treats `obj.fn` as
// stable under a second evaluation; refusing it here would reject the shape
// the rule mostly exists for.
func (scanner *betweenScanner) isInertStatement(statement *ast.Node) bool {
	if statement == nil || statement.Kind != ast.KindExpressionStatement {
		return false
	}
	expressionStatement := statement.AsExpressionStatement()
	if expressionStatement == nil {
		return false
	}
	rootCall, awaited := assertionRootCall(expressionStatement.Expression)
	if awaited {
		return false
	}
	var parsed *rstestUtils.ParsedRstestExpectCall
	if rootCall != nil {
		parsed = scanner.analysis.ParseExpectCall(rootCall)
	}
	if parsed == nil || parsed.Head == nil {
		// Not an assertion. It can still be inert on its own terms — a
		// `console.log('checkpoint')` calls nothing of the author's — with no
		// framework call to exempt.
		return scanner.inertSubtree(expressionStatement.Expression, nil, "")
	}
	for _, modifier := range parsed.Modifiers {
		if suspendingModifiers[modifier] {
			return false
		}
	}
	frameworkCalls := map[*ast.Node]bool{parsed.Head: true}
	for _, matcher := range parsed.Matchers {
		if invokingMatchers[matcher.Name] {
			return false
		}
		if matcher.Entry.Call != nil {
			frameworkCalls[matcher.Entry.Call] = true
		}
	}
	return scanner.inertSubtree(expressionStatement.Expression, frameworkCalls, expectRootName(parsed.Head))
}

// collectIdentifierNames gathers every identifier a subtree mentions.
func collectIdentifierNames(node *ast.Node, into map[string]bool) {
	if node == nil {
		return
	}
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindIdentifier {
			into[node.AsIdentifier().Text] = true
		}
		return node.ForEachChild(visit)
	}
	visit(node)
}

// onlyInertStatementsBetween reports whether every statement between the two
// assertions leaves their merge intact. Upstream asks a much narrower
// question — whether a mockClear on a plain identifier sits between them — and
// merges across a reassignment, a re-invocation, or a reset it does not
// recognise.
func onlyInertStatementsBetween(
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
	statements []*ast.Node,
	first, second *mergeCandidate,
) bool {
	minPosition, maxPosition := first.position, second.position
	if minPosition > maxPosition {
		minPosition, maxPosition = maxPosition, minPosition
	}
	scanner := &betweenScanner{ctx: ctx, analysis: analysis, assertionNames: map[string]bool{}}
	collectIdentifierNames(first.statement, scanner.assertionNames)
	collectIdentifierNames(second.statement, scanner.assertionNames)

	for _, statement := range statements {
		if statement == nil {
			continue
		}
		position := internalUtils.TrimNodeTextRange(ctx.SourceFile, statement).Pos()
		if position <= minPosition || position >= maxPosition {
			continue
		}
		if !scanner.isInertStatement(statement) && !scanner.isInertDeclaration(statement) {
			return false
		}
	}
	return true
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

func reportPair(
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
	statements []*ast.Node,
	candidates []*mergeCandidate,
) {
	first, second := candidates[0], candidates[1]
	if first.hits[0].role == second.hits[0].role ||
		!onlyInertStatementsBetween(ctx, analysis, statements, first, second) {
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
			report:   func() { reportPair(ctx, analysis, statements, candidates) },
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
