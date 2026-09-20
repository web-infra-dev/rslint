package require_hook

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed require_hook.schema.json
var schemaJSON []byte

func buildUseHookMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useHook",
		Description: "This should be done within a hook",
	}
}

type Options struct {
	AllowedFunctionCalls []string
}

func parseAllowedFunctionCalls(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func parseOptions(options []any) Options {
	opts := Options{}
	if len(options) == 0 {
		return opts
	}

	optsMap, ok := options[0].(map[string]any)
	if !ok {
		return opts
	}
	if raw, ok := optsMap["allowedFunctionCalls"]; ok {
		opts.AllowedFunctionCalls = parseAllowedFunctionCalls(raw)
	}
	return opts
}

func isNullOrUndefined(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "undefined"
	default:
		return false
	}
}

// isRootedInUtilitiesObject reports whether a callee chain is read off Rstest's
// utilities object, however many members deep: `rs.mock('./m')`,
// `rstest.spyOn(o, 'm').mockReturnValue(1)`, `import.meta.rstest.rs.fn()`.
//
// This is where the Rstest port deliberately parts company with
// @vitest/eslint-plugin, which spells the same exemption as
// `getNodeName(node)?.startsWith('vi')` — with no separator, so `video.play()`,
// `viewport.set()` and `visit('/home')` are all silently treated as framework
// calls. eslint-plugin-jest has the separator (`'jest.'`) and does not have the
// bug. Reproducing it under Rstest's `rs` prefix would swallow `response.*`,
// `result.*`, `resetDatabase()` and most other setup spellings, so the receiver
// is resolved instead of pattern-matched.
//
// Resolving also buys the spellings a prefix test cannot reach: a renamed
// import, a namespace import and `import.meta.rstest`.
func isRootedInUtilitiesObject(ctx rule.RuleContext, expr *ast.Node) bool {
	for {
		expr = internalUtils.SkipAssertionsAndParens(expr)
		if expr == nil {
			return false
		}
		switch expr.Kind {
		case ast.KindCallExpression:
			expr = expr.AsCallExpression().Expression
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			receiver := expr.Expression()
			if rstestUtils.IsUtilitiesObject(ctx, receiver) {
				return true
			}
			expr = receiver
		default:
			return false
		}
	}
}

// isRstestFnCall reports whether a call is part of Rstest's own API surface,
// which is what the rule takes as "not setup code that belongs in a hook".
//
// The four sources are separate because the plugin models them separately:
// registrations and hooks come from the shared call analysis, expect chains
// have their own parser, and the utilities object is reached either as the
// build rewrites it (by the receiver as written) or through ordinary bindings.
//
// They are asked in increasing cost: a candidate-gated parse, then pure
// syntax, then a receiver resolution, and only last the expect parser, whose
// candidate set is completed by indexing every callback in the file.
func isRstestFnCall(
	node *ast.Node,
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
) bool {
	if analysis.ParseFnCall(node) != nil {
		return true
	}
	// A plugin-managed member is matched by the receiver as written, because
	// the build rewrites it that way even in a file that declares its own `rs`.
	if rstestUtils.ParseRstestPluginManagedCall(node) != nil {
		return true
	}
	if isRootedInUtilitiesObject(ctx, node.AsCallExpression().Expression) {
		return true
	}
	return analysis.IsExpectCall(node)
}

func shouldBeInHook(
	node *ast.Node,
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
	allowedFunctionCalls []string,
) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindExpressionStatement:
		return shouldBeInHook(
			ast.SkipParentheses(node.AsExpressionStatement().Expression),
			ctx,
			analysis,
			allowedFunctionCalls,
		)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		// A dynamic import is an ImportExpression upstream rather than a
		// CallExpression, and an optional call is wrapped in a ChainExpression,
		// so neither reaches the CallExpression branch there.
		if call.Expression.Kind == ast.KindImportKeyword ||
			call.QuestionDotToken != nil ||
			ast.IsOptionalChain(node) {
			return false
		}
		if isRstestFnCall(node, ctx, analysis) {
			return false
		}
		return !slices.Contains(allowedFunctionCalls, testFramework.CalleeChainName(node))
	case ast.KindVariableStatement:
		// `export let value = setup()` is an ExportNamedDeclaration upstream,
		// so the VariableDeclaration is never a block-body statement there.
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
			return false
		}
		declList := node.AsVariableStatement().DeclarationList
		if declList == nil || declList.Flags&ast.NodeFlagsBlockScoped == ast.NodeFlagsConst {
			return false
		}
		decls := declList.AsVariableDeclarationList().Declarations
		if decls == nil {
			return false
		}
		for _, decl := range decls.Nodes {
			declaration := decl.AsVariableDeclaration()
			if declaration.Initializer != nil && !isNullOrUndefined(declaration.Initializer) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func hasCallExpressionParent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindCallExpression
}

// runsDuringCollection reports whether statement stands directly in code Rstest
// executes while it collects the file: the module body, or the block a
// `describe` callback written at the call site runs.
//
// Only a function written at the call site counts, which is upstream's
// boundary: a callback passed by name is not descended into. A named function
// can be registered more than once and called from elsewhere, so attributing
// its body to one suite would report statements that are not suite setup.
func runsDuringCollection(
	statement *ast.Node,
	analysis *rstestUtils.RstestCallAnalysis,
) bool {
	parent := statement.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindSourceFile {
		return true
	}
	if parent.Kind != ast.KindBlock {
		return false
	}

	callback := parent.Parent
	if callback == nil || !testFramework.IsFunction(callback) {
		return false
	}
	call := callback.Parent
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	// A `describe` nested in another call is reported as that outer statement
	// instead, so its body is left alone.
	if hasCallExpressionParent(call) {
		return false
	}
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) < 2 || args.Nodes[1] != callback {
		return false
	}
	parsed := analysis.ParseFnCall(call)
	return parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeDescribe
}

var RequireHookRule = rule.Rule{
	Name:   "rstest/require-hook",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)

		// Statements are checked where the traversal reaches them rather than
		// by scanning each block, so diagnostics come out in source order even
		// when a suite body sits between two module-level statements.
		check := func(statement *ast.Node) {
			if !runsDuringCollection(statement, analysis) {
				return
			}
			if !shouldBeInHook(statement, ctx, analysis, opts.AllowedFunctionCalls) {
				return
			}
			ctx.ReportNode(statement, buildUseHookMessage())
		}

		return rule.RuleListeners{
			ast.KindExpressionStatement: check,
			ast.KindVariableStatement:   check,
		}
	},
}
