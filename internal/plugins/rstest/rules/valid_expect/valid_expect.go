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
	sharedValidExpect "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_expect"
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
// and async/await fixers live in the shared package: they are pure syntax, with
// no jest or rstest semantics, and keeping two copies is what let the same
// parenthesized-assertion false positive exist twice.

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
			fixes = append(fixes, sharedValidExpect.AsyncInsertFix(ctx.SourceFile, fn))
			asyncInserted[fn] = true
		}
		fixes = append(fixes, sharedValidExpect.AwaitFix(ctx.SourceFile, descriptor.node, alwaysAwait))
	}

	if len(fixes) > 0 {
		ctx.ReportNodeWithFixes(descriptor.node, msg, fixes...)
		return
	}
	ctx.ReportNode(descriptor.node, msg)
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
				if parsed.Head == nil || parsed.Head.Kind != ast.KindCallExpression {
					return
				}

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

				if len(parsed.Matchers) == 0 || !shouldBeAwaited(parsed, opts.AsyncMatchers) {
					return
				}

				reportNode, promiseWrapped, insideAssertionArray, shouldReport := sharedValidExpect.ResolveAsyncAssertionReportNode(
					parsed.Expression,
					opts.AlwaysAwait,
				)
				if reportNode == nil {
					return
				}

				reportNodeKey := sharedValidExpect.PromiseArrayExceptionKey(ctx.SourceFile, reportNode)
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
