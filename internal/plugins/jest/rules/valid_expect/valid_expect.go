package valid_expect

import (
	_ "embed"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	sharedValidExpect "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_expect"
	"slices"
	"strconv"
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

// Message Builders

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
		Data: map[string]string{
			"orReturned": orReturned,
		},
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
		Data: map[string]string{
			"orReturned": orReturned,
		},
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

func shouldBeAwaited(parsed *utils.ParsedJestFnCall, asyncMatchers []string) bool {
	for _, modifier := range parsed.Modifiers {
		if modifier != "not" {
			return true
		}
	}
	return slices.Contains(asyncMatchers, parsed.Matcher)
}

func expectOpenParenRange(sourceFile *ast.SourceFile, call *ast.Node) core.TextRange {
	if sourceFile == nil || call == nil || call.Kind != ast.KindCallExpression {
		return internalUtils.TrimNodeTextRange(sourceFile, call)
	}

	callExpr := call.AsCallExpression()
	start := internalUtils.TrimNodeTextRange(sourceFile, callExpr.Expression).End()
	text := sourceFile.Text()
	for i := start; i < len(text) && i < call.End(); i++ {
		if text[i] == '(' {
			return core.NewTextRange(i, i+1)
		}
	}

	return internalUtils.TrimNodeTextRange(sourceFile, call).WithEnd(start)
}

func tooManyArgsRange(sourceFile *ast.SourceFile, args []*ast.Node, maxArgs int) core.TextRange {
	start := internalUtils.TrimNodeTextRange(sourceFile, args[maxArgs]).Pos()
	end := internalUtils.TrimNodeTextRange(sourceFile, args[len(args)-1]).End()
	if end > start {
		end--
	}
	return core.NewTextRange(start, end)
}

// resolveAsyncAssertionReportNode locates the assertion expression a matcher
// belongs to — jest matchers are always called, so it is the call wrapping the
// matcher's member access — and hands it to the shared resolver.
func resolveAsyncAssertionReportNode(
	resolver *sharedValidExpect.AsyncAssertionResolver,
	matcherEntry *utils.ParsedJestFnMemberEntry,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	if matcherEntry == nil {
		return nil, false, false, false
	}
	assertionExpr := testFramework.DirectMemberCall(matcherEntry)
	if assertionExpr == nil {
		assertionExpr = matcherEntry.Call
	}
	if assertionExpr == nil {
		return nil, false, false, false
	}

	return resolver.ResolveAsyncAssertionReportNode(assertionExpr, alwaysAwait)
}

func reportAsyncDescriptor(
	ctx rule.RuleContext,
	resolver *sharedValidExpect.AsyncAssertionResolver,
	descriptor asyncDescriptor,
	alwaysAwait bool,
	fnFixable map[*ast.Node]bool,
) {
	msg := buildAsyncDescriptorMessage(descriptor, alwaysAwait)
	resolver.ReportAsyncDescriptor(ctx, descriptor.node, msg, alwaysAwait, fnFixable)
}

func findTopLevelMemberAccess(node *ast.Node) *ast.Node {
	current := node
	for current != nil && utils.IsMemberAccessNode(current.Parent) {
		current = current.Parent
	}
	return current
}

var ValidExpectRule = rule.Rule{
	Name:   "jest/valid-expect",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		analysis := utils.GetJestCallAnalysis(ctx)
		asyncResolver := sharedValidExpect.NewAsyncAssertionResolver(ctx)
		arrayExceptions := map[string]bool{}
		fnFixable := map[*ast.Node]bool{}
		var descriptors []asyncDescriptor

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed, reason := analysis.ParseExpectCallWithReason(node)
				if parsed == nil {
					if reason == "" {
						return
					}

					reportNode := node
					if utils.IsMemberAccessNode(node.Parent) {
						topMember := findTopLevelMemberAccess(node.Parent)
						if topMember != nil {
							reportNode = topMember
						}
					}

					switch reason {
					case utils.ExpectParseReasonMatcherNotFound:
						ctx.ReportNode(reportNode, buildErrorMatcherNotFoundMessage())
					case utils.ExpectParseReasonMatcherNotCalled:
						propNode, propName := testFramework.AccessorPropertyAndName(reportNode)
						if propNode == nil {
							ctx.ReportNode(reportNode, buildErrorMatcherNotCalledMessage())
							return
						}
						if propName != "" && !testFramework.IsComputedIdentifierAccessor(propNode) && utils.EXPECT_MODIFIER_NAMES[propName] {
							ctx.ReportNode(propNode, buildErrorMatcherNotFoundMessage())
							return
						}
						ctx.ReportNode(propNode, buildErrorMatcherNotCalledMessage())
					case utils.ExpectParseReasonModifierUnknown:
						ctx.ReportNode(reportNode, buildErrorModifierUnknownMessage())
					}
					return
				}
				if parsed.Kind != utils.JestFnTypeExpect || parsed.Head.Local.Node == nil {
					return
				}

				expectCall := ast.WalkUpParenthesizedExpressions(parsed.Head.Local.Node.Parent)
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				args := expectCall.AsCallExpression().Arguments.Nodes
				if len(args) < opts.MinArgs {
					ctx.ReportRange(
						expectOpenParenRange(ctx.SourceFile, expectCall),
						buildErrorNotEnoughArgsMessage(opts.MinArgs),
					)
				}
				if len(args) > opts.MaxArgs {
					ctx.ReportRange(
						tooManyArgsRange(ctx.SourceFile, args, opts.MaxArgs),
						buildErrorTooManyArgsMessage(opts.MaxArgs),
					)
				}

				if parsed.MatcherEntry == nil || !shouldBeAwaited(parsed, opts.AsyncMatchers) {
					return
				}

				reportNode, promiseWrapped, insideAssertionArray, shouldReport := resolveAsyncAssertionReportNode(
					asyncResolver,
					parsed.MatcherEntry,
					opts.AlwaysAwait,
				)
				if reportNode == nil {
					return
				}

				reportNodeKey := sharedValidExpect.PromiseArrayExceptionKey(ctx.SourceFile, reportNode)
				if shouldReport && !arrayExceptions[reportNodeKey] {
					arrayExceptions[reportNodeKey] = true
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
					reportAsyncDescriptor(ctx, asyncResolver, descriptor, opts.AlwaysAwait, fnFixable)
				}
			},
		}
	},
}
