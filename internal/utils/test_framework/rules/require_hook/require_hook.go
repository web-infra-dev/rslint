package require_hook

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed require_hook.schema.json
var schemaJSON []byte

// Runtime is the framework-specific half of the rule: what counts as the
// framework's own API, and which call is a suite registration.
type Runtime struct {
	// IsFrameworkCall reports whether a call is part of the framework's own
	// API surface, which the rule takes as "not setup code that belongs in a
	// hook".
	IsFrameworkCall func(call *ast.Node) bool
	// IsDescribe reports whether a call registers a suite.
	IsDescribe func(call *ast.Node) bool
	// ExecutionTimeAPI names the API a call reaches when it may only run inside
	// a test (so a hook is the wrong repair), or returns "". Optional.
	ExecutionTimeAPI func(call *ast.Node) string
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

type Options struct {
	AllowedFunctionCalls []string
}

func buildUseHookMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useHook",
		Description: "This should be done within a hook",
	}
}

func buildUseTestMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useTest",
		Description: name + "() can only be called inside a test",
		Data:        map[string]string{"name": name},
	}
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

func shouldBeInHook(node *ast.Node, runtime Runtime, allowedFunctionCalls []string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindExpressionStatement:
		return shouldBeInHook(
			ast.SkipParentheses(node.AsExpressionStatement().Expression),
			runtime,
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
		if runtime.IsFrameworkCall(node) {
			return false
		}
		// An unnameable callee — a conditional expression, a computed member
		// whose key is not a literal — has no name to match, and upstream's
		// getNodeName yields undefined rather than the empty string there.
		// Comparing the empty name would let `allowedFunctionCalls: [""]`
		// exempt every such call.
		name := testFramework.CalleeChainName(node)
		return name == "" || !slices.Contains(allowedFunctionCalls, name)
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

// messageFor picks the repair the reported statement can actually take. Almost
// every statement belongs in a lifecycle hook, but a framework's execution-time
// APIs belong inside a test instead.
func messageFor(runtime Runtime, statement *ast.Node) rule.RuleMessage {
	if runtime.ExecutionTimeAPI == nil || statement.Kind != ast.KindExpressionStatement {
		return buildUseHookMessage()
	}
	call := ast.SkipParentheses(statement.AsExpressionStatement().Expression)
	if call == nil || call.Kind != ast.KindCallExpression {
		return buildUseHookMessage()
	}
	if name := runtime.ExecutionTimeAPI(call); name != "" {
		return buildUseTestMessage(name)
	}
	return buildUseHookMessage()
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

// runsDuringCollection reports whether statement stands directly in code the
// framework executes while it collects the file: the module body, or the block
// a `describe` callback written at the call site runs.
//
// Only a function written at the call site counts, which is upstream's
// boundary: a callback passed by name is not descended into. A named function
// can be registered more than once and called from elsewhere, so attributing
// its body to one suite would report statements that are not suite setup.
func runsDuringCollection(statement *ast.Node, runtime Runtime) bool {
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
	// ts-go keeps parentheses and TypeScript's type-only syntax as nodes where
	// ESTree exposes the function directly, so `describe('s', (() => { … }))`
	// reaches the call one or more wrappers above the callback.
	call := outermostTransparentWrapper(callback).Parent
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	// A `describe` nested in another call is reported as that outer statement
	// instead, so its body is left alone.
	if hasCallExpressionParent(call) {
		return false
	}
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) < 2 ||
		internalUtils.SkipAssertionsAndParens(args.Nodes[1]) != callback {
		return false
	}
	return runtime.IsDescribe(call)
}

// outermostTransparentWrapper returns the outermost expression wrapping node
// that leaves what a surrounding call receives unchanged. It is the upward
// counterpart of SkipAssertionsAndParens.
func outermostTransparentWrapper(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression:
			node = node.Parent
		default:
			return node
		}
	}
	return node
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			opts := parseOptions(options)
			runtime := config.Prepare(ctx)

			// Statements are checked where the traversal reaches them rather
			// than by scanning each block, so diagnostics come out in source
			// order even when a suite body sits between two module-level
			// statements.
			check := func(statement *ast.Node) {
				if !runsDuringCollection(statement, runtime) {
					return
				}
				if !shouldBeInHook(statement, runtime, opts.AllowedFunctionCalls) {
					return
				}
				ctx.ReportNode(statement, messageFor(runtime, statement))
			}

			return rule.RuleListeners{
				ast.KindExpressionStatement: check,
				ast.KindVariableStatement:   check,
			}
		},
	}
}
