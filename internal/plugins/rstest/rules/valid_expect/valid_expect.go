package valid_expect

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed valid_expect.schema.json
var schemaJSON []byte

type validExpectOptions struct {
	AlwaysAwait   bool
	AsyncMatchers []string
	MinArgs       int
	MaxArgs       int
}

type asyncDescriptor struct {
	node           *ast.Node
	promiseWrapped bool
}

func pluralSuffix(amount int) string {
	if amount == 1 {
		return ""
	}
	return "s"
}

// Message builders. Ids and text match jest/valid-expect exactly.

func buildErrorTooManyArgsMessage(amount int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tooManyArgs",
		Description: fmt.Sprintf("Expect takes at most %d argument%s", amount, pluralSuffix(amount)),
		Data: map[string]string{
			"amount": strconv.Itoa(amount),
			"s":      pluralSuffix(amount),
		},
	}
}

func buildErrorNotEnoughArgsMessage(amount int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notEnoughArgs",
		Description: fmt.Sprintf("Expect requires at least %d argument%s", amount, pluralSuffix(amount)),
		Data: map[string]string{
			"amount": strconv.Itoa(amount),
			"s":      pluralSuffix(amount),
		},
	}
}

func buildErrorModifierUnknownMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "modifierUnknown",
		Description: "Expect has an unknown modifier",
	}
}

func buildErrorMatcherNotFoundMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "matcherNotFound",
		Description: "Expect must have a corresponding matcher call",
	}
}

func buildErrorMatcherNotCalledMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "matcherNotCalled",
		Description: "Matchers must be called to assert",
	}
}

func buildErrorAsyncMustBeAwaitedMessage(alwaysAwait bool) rule.RuleMessage {
	orReturned := " or returned"
	if alwaysAwait {
		orReturned = ""
	}
	return rule.RuleMessage{
		Id:          "asyncMustBeAwaited",
		Description: "Async assertions must be awaited" + orReturned,
		Data:        map[string]string{"orReturned": orReturned},
	}
}

func buildErrorPromisesWithAsyncAssertionsMustBeAwaitedMessage(alwaysAwait bool) rule.RuleMessage {
	orReturned := " or returned"
	if alwaysAwait {
		orReturned = ""
	}
	return rule.RuleMessage{
		Id:          "promisesWithAsyncAssertionsMustBeAwaited",
		Description: "Promises which return async assertions must be awaited" + orReturned,
		Data:        map[string]string{"orReturned": orReturned},
	}
}

func buildAsyncDescriptorMessage(descriptor asyncDescriptor, alwaysAwait bool) rule.RuleMessage {
	if descriptor.promiseWrapped {
		return buildErrorPromisesWithAsyncAssertionsMustBeAwaitedMessage(alwaysAwait)
	}
	return buildErrorAsyncMustBeAwaitedMessage(alwaysAwait)
}

func parseOptions(options []any) validExpectOptions {
	out := validExpectOptions{
		AlwaysAwait:   false,
		AsyncMatchers: []string{"toReject", "toResolve"},
		MinArgs:       1,
		MaxArgs:       1,
	}

	if len(options) == 0 {
		return out
	}
	m, _ := options[0].(map[string]interface{})

	if raw, ok := m["alwaysAwait"].(bool); ok {
		out.AlwaysAwait = raw
	}
	out.MinArgs = readIntOption(m, "minArgs", out.MinArgs)
	out.MaxArgs = readIntOption(m, "maxArgs", out.MaxArgs)
	if raw, ok := m["asyncMatchers"].([]interface{}); ok {
		out.AsyncMatchers = out.AsyncMatchers[:0]
		for _, value := range raw {
			if s, ok := value.(string); ok {
				out.AsyncMatchers = append(out.AsyncMatchers, s)
			}
		}
	}

	return out
}

func readIntOption(options map[string]interface{}, key string, defaultValue int) int {
	raw, ok := options[key]
	if !ok {
		return defaultValue
	}

	switch value := raw.(type) {
	case float64:
		return int(value)
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	default:
		return defaultValue
	}
}

// --- Async assertion machinery ---
//
// The Promise-chain walking, acceptable-return detection, array de-duplication
// and async/await fixers below are ported verbatim from
// internal/plugins/jest/rules/valid_expect/valid_expect.go, which is itself
// aligned with eslint-plugin-vitest. Kept as a local copy rather than shared:
// sharing them would require touching the jest rule and re-running its
// regression, and there is no third consumer yet.

func isPromiseMethodCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}

	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !isMemberAccessNode(callee) {
		return false
	}

	return testFramework.CalleeChainName(internalUtils.AccessExpressionObject(callee)) == "Promise"
}

func getPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	if node.Kind == ast.KindArrayLiteralExpression && node.Parent != nil && node.Parent.Kind == ast.KindCallExpression {
		node = node.Parent
	}

	if isPromiseMethodCall(node) {
		return node
	}

	return nil
}

func findPromiseCallExpressionNode(node *ast.Node) *ast.Node {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	if node.Parent.Kind != ast.KindCallExpression && node.Parent.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	return getPromiseCallExpressionNode(node.Parent)
}

func getParentIfPromiseChained(node *ast.Node) *ast.Node {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return node
	}

	grandParent := node.Parent.Parent
	if grandParent.Kind != ast.KindCallExpression || !isMemberAccessNode(grandParent.AsCallExpression().Expression) {
		return node
	}

	member := grandParent.AsCallExpression().Expression
	entries := testFramework.GetMemberEntries(member)
	if len(entries) == 0 {
		return node
	}

	last := entries[len(entries)-1].Name
	if last == "then" || last == "catch" {
		return getParentIfPromiseChained(grandParent)
	}

	return node
}

func isAcceptableReturnNode(node *ast.Node, allowReturn bool) bool {
	if node == nil {
		return false
	}

	if allowReturn && node.Kind == ast.KindReturnStatement {
		return true
	}
	if node.Kind == ast.KindConditionalExpression {
		return isAcceptableReturnNode(node.Parent, allowReturn)
	}

	return node.Kind == ast.KindArrowFunction || node.Kind == ast.KindAwaitExpression
}

func promiseArrayExceptionKey(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	return fmt.Sprintf("%d:%d", r.Pos(), r.End())
}

func asyncInsertFix(sourceFile *ast.SourceFile, fn *ast.Node) rule.RuleFix {
	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		head := internalUtils.GetFunctionHeadLoc(sourceFile, fn)
		return rule.RuleFixReplaceRange(core.NewTextRange(head.Pos(), head.Pos()), "async ")
	default:
		return rule.RuleFixInsertBefore(sourceFile, fn, "async ")
	}
}

func awaitFix(sourceFile *ast.SourceFile, node *ast.Node, alwaysAwait bool) rule.RuleFix {
	if alwaysAwait && node.Parent != nil && node.Parent.Kind == ast.KindReturnStatement {
		ret := node.Parent
		retRange := internalUtils.TrimNodeTextRange(sourceFile, ret)
		nodeRange := internalUtils.TrimNodeTextRange(sourceFile, node)
		return rule.RuleFixReplaceRange(core.NewTextRange(retRange.Pos(), nodeRange.Pos()), "await ")
	}
	return rule.RuleFixInsertBefore(sourceFile, node, "await ")
}

func reportAsyncDescriptor(
	ctx rule.RuleContext,
	descriptor asyncDescriptor,
	alwaysAwait bool,
	asyncInserted map[*ast.Node]bool,
) {
	msg := buildAsyncDescriptorMessage(descriptor, alwaysAwait)

	var fixes []rule.RuleFix
	if fn := ast.GetContainingFunction(descriptor.node); fn != nil {
		if !ast.IsAsyncFunction(fn) && !asyncInserted[fn] {
			fixes = append(fixes, asyncInsertFix(ctx.SourceFile, fn))
			asyncInserted[fn] = true
		}
		fixes = append(fixes, awaitFix(ctx.SourceFile, descriptor.node, alwaysAwait))
	}

	if len(fixes) > 0 {
		ctx.ReportNodeWithFixes(descriptor.node, msg, fixes...)
		return
	}
	ctx.ReportNode(descriptor.node, msg)
}

// --- Rstest-specific helpers ---

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

// expectFactoryOpenParenRange locates the `(` of the assertion factory call so
// notEnoughArgs points at the empty argument list, mirroring jest's
// expectOpenParenRange. head is the factory call (expect(x) / expect.soft(x)).
func expectFactoryOpenParenRange(sourceFile *ast.SourceFile, head *ast.Node) core.TextRange {
	if sourceFile == nil || head == nil || head.Kind != ast.KindCallExpression {
		return internalUtils.TrimNodeTextRange(sourceFile, head)
	}

	callExpr := head.AsCallExpression()
	start := internalUtils.TrimNodeTextRange(sourceFile, callExpr.Expression).End()
	text := sourceFile.Text()
	for i := start; i < len(text) && i < head.End(); i++ {
		if text[i] == '(' {
			return core.NewTextRange(i, i+1)
		}
	}

	return internalUtils.TrimNodeTextRange(sourceFile, head).WithEnd(start)
}

func tooManyArgsRange(sourceFile *ast.SourceFile, args []*ast.Node, maxArgs int) core.TextRange {
	start := internalUtils.TrimNodeTextRange(sourceFile, args[maxArgs]).Pos()
	end := internalUtils.TrimNodeTextRange(sourceFile, args[len(args)-1]).End()
	if end > start {
		end--
	}
	return core.NewTextRange(start, end)
}

// isAllowedExtraExpectArg reports whether a second expect argument is a valid
// message overload, matching eslint-plugin-vitest: expect(actual, message) and
// expect(actual, `template`) are legal, and expect.poll / expect.element take
// an options object as their second argument.
func isAllowedExtraExpectArg(arg *ast.Node, entry rstestUtils.RstestExpectEntry) bool {
	if entry == rstestUtils.RstestExpectEntryPoll || entry == rstestUtils.RstestExpectEntryElement {
		return true
	}
	if arg == nil {
		return false
	}
	switch ast.SkipParentheses(arg).Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	default:
		return false
	}
}

func shouldBeAwaited(parsed *rstestUtils.ParsedRstestExpectCall, asyncMatchers []string) bool {
	return rstestUtils.ShouldRstestExpectBeAwaited(parsed, asyncMatchers)
}

// resolveAsyncAssertionReportNode mirrors the jest rule: from the matcher
// member it walks out through chained .then/.catch, notes whether the assertion
// sits inside a Promise.all([...]) array, and decides whether the resulting node
// must be awaited or returned.
func resolveAsyncAssertionReportNode(
	matcherEntry *rstestUtils.ParsedRstestFnMemberEntry,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	if matcherEntry == nil || matcherEntry.Node == nil || matcherEntry.Node.Parent == nil {
		return nil, false, false, false
	}

	matcherMemberNode := matcherEntry.Node.Parent
	if matcherMemberNode.Parent == nil {
		return nil, false, false, false
	}

	promiseChainedAssertionNode := getParentIfPromiseChained(matcherMemberNode.Parent)
	insideAssertionArray = promiseChainedAssertionNode.Parent != nil && promiseChainedAssertionNode.Parent.Kind == ast.KindArrayLiteralExpression
	reportNode = promiseChainedAssertionNode
	if promiseCallNode := findPromiseCallExpressionNode(promiseChainedAssertionNode); promiseCallNode != nil {
		reportNode = promiseCallNode
		promiseWrapped = true
	}

	if reportNode.Parent == nil || isAcceptableReturnNode(reportNode.Parent, !alwaysAwait) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	return reportNode, promiseWrapped, insideAssertionArray, true
}

// reportBrokenChain reports a chain that the parser resolved without a matcher.
// The reason comes directly from the shared analysis, so unlike jest this rule
// does not re-derive it. matcher-not-called points at the trailing member; the
// other reasons point at the outermost expression.
func reportBrokenChain(ctx rule.RuleContext, parsed *rstestUtils.ParsedRstestExpectCall) {
	switch parsed.Reason {
	case rstestUtils.RstestExpectParseReasonMatcherNotFound:
		ctx.ReportNode(parsed.Expression, buildErrorMatcherNotFoundMessage())
	case rstestUtils.RstestExpectParseReasonMatcherNotCalled:
		if len(parsed.MemberEntries) == 0 {
			ctx.ReportNode(parsed.Expression, buildErrorMatcherNotCalledMessage())
			return
		}
		last := parsed.MemberEntries[len(parsed.MemberEntries)-1]
		// A trailing modifier that was never followed by a matcher is a missing
		// matcher, not an uncalled one, matching jest/vitest.
		if rstestUtils.RSTEST_EXPECT_MODIFIER_NAMES[last.Name] {
			ctx.ReportNode(last.Node, buildErrorMatcherNotFoundMessage())
			return
		}
		ctx.ReportNode(last.Node, buildErrorMatcherNotCalledMessage())
	case rstestUtils.RstestExpectParseReasonModifierUnknown:
		ctx.ReportNode(parsed.Expression, buildErrorModifierUnknownMessage())
	}
}

var ValidExpectRule = rule.Rule{
	Name:   "rstest/valid-expect",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		arrayExceptions := map[string]bool{}
		asyncInserted := map[*ast.Node]bool{}
		var descriptors []asyncDescriptor

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil {
					return
				}

				if parsed.Reason != rstestUtils.RstestExpectParseReasonNone {
					reportBrokenChain(ctx, parsed)
					return
				}

				// Static members (expect.assertions(1), expect.any(...)) carry no
				// assertion factory, so argument and await checks do not apply.
				if parsed.Head != nil && parsed.Head.Kind == ast.KindCallExpression {
					args := parsed.Head.AsCallExpression().Arguments.Nodes
					if len(args) < opts.MinArgs {
						ctx.ReportRange(
							expectFactoryOpenParenRange(ctx.SourceFile, parsed.Head),
							buildErrorNotEnoughArgsMessage(opts.MinArgs),
						)
					}
					if len(args) > opts.MaxArgs {
						// vitest allowance: a lone message string/template, or the
						// options object of poll/element, is not an excess argument.
						if len(args) != opts.MaxArgs+1 || !isAllowedExtraExpectArg(args[opts.MaxArgs], parsed.Entry) {
							ctx.ReportRange(
								tooManyArgsRange(ctx.SourceFile, args, opts.MaxArgs),
								buildErrorTooManyArgsMessage(opts.MaxArgs),
							)
						}
					}
				}

				if parsed.MatcherEntry == nil || !shouldBeAwaited(parsed, opts.AsyncMatchers) {
					return
				}

				reportNode, promiseWrapped, insideAssertionArray, shouldReport := resolveAsyncAssertionReportNode(
					parsed.MatcherEntry,
					opts.AlwaysAwait,
				)
				if reportNode == nil {
					return
				}

				reportNodeKey := promiseArrayExceptionKey(ctx.SourceFile, reportNode)
				if shouldReport && !arrayExceptions[reportNodeKey] {
					descriptors = append(descriptors, asyncDescriptor{
						node:           reportNode,
						promiseWrapped: promiseWrapped,
					})
				}

				if insideAssertionArray {
					arrayExceptions[reportNodeKey] = true
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) {
				_ = node
				for _, descriptor := range descriptors {
					reportAsyncDescriptor(ctx, descriptor, opts.AlwaysAwait, asyncInserted)
				}
			},
		}
	},
}
