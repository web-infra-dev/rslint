package getter_return

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

//go:embed getter_return.schema.json
var schemaJSON []byte

// Options for getter-return rule
type Options struct {
	AllowImplicit bool `json:"allowImplicit"`
}

func parseOptions(options []any) Options {
	opts := Options{
		AllowImplicit: false,
	}

	if len(options) == 0 {
		return opts
	}

	optsMap, _ := options[0].(map[string]any)
	if v, ok := optsMap["allowImplicit"].(bool); ok {
		opts.AllowImplicit = v
	}
	return opts
}

func buildExpectedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expected",
		Description: "Expected to return a value in " + name + ".",
		Data:        map[string]string{"name": name},
	}
}

func buildExpectedAlwaysMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedAlways",
		Description: "Expected " + name + " to always return a value.",
		Data:        map[string]string{"name": name},
	}
}

func reportEmptyReturns(ctx rule.RuleContext, funcNode *ast.Node, body *ast.Node) (string, rule.RuleMessage) {
	name := utils.GetFunctionNameWithKindCore(funcNode)
	message := buildExpectedMessage(name)
	ast.ForEachReturnStatement(body, func(stmt *ast.Node) bool {
		if stmt.Expression() == nil {
			ctx.ReportNode(stmt, message)
		}
		return false
	})
	return name, message
}

func eslintEndReachable(funcNode *ast.Node, body *ast.Node) bool {
	statements := body.Statements()
	if len(statements) == 0 {
		return true
	}
	switch statements[len(statements)-1].Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return false
	}
	return cfg.Build(funcNode, cfg.Hooks[struct{}]{}).EndReachable
}

// reportGetterReturn mirrors ESLint's code-path checks for one getter. A bare
// return is reported at the return itself unless allowImplicit is enabled. If
// the function end is reachable, the missing path is reported at the function
// head and any kind of return selects expectedAlways.
func reportGetterReturn(ctx rule.RuleContext, funcNode *ast.Node, opts Options) {
	if funcNode == nil {
		return
	}

	body := funcNode.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return
	}

	var name string
	var expected rule.RuleMessage
	hasExpected := false
	analysis := utils.AnalyzeFunctionReturns(funcNode)
	// The binder supplies the return kinds cheaply, while ESLint's code-path
	// shape is authoritative for whether the getter can reach its end.
	analysis.EndReachable = eslintEndReachable(funcNode, body)
	if !opts.AllowImplicit && analysis.HasEmptyReturn {
		name, expected = reportEmptyReturns(ctx, funcNode, body)
		hasExpected = true
	}

	if !analysis.EndReachable {
		return
	}

	if name == "" {
		name = utils.GetFunctionNameWithKindCore(funcNode)
	}
	message := expected
	if analysis.HasReturnWithValue || analysis.HasEmptyReturn {
		message = buildExpectedAlwaysMessage(name)
	} else if !hasExpected {
		message = buildExpectedMessage(name)
	}
	ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, funcNode), message)
}

// isGlobalReference checks the effective global scope rather than TypeScript's
// default libraries, so Reflect follows ecmaVersion and authored overrides.
func isGlobalReference(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	name := node.AsIdentifier().Text
	if !ctx.Globals.Access(name).IsDeclared() {
		return false
	}
	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(node); symbol != nil {
			return !utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return !utils.IsShadowed(node, name)
}

// matchGlobalMethodCall checks for an unshadowed global object method call.
// It accepts dot/bracket access, optional chaining, and parentheses, matching
// ESLint's isSpecificMemberAccess and SourceCode#isGlobalReference checks.
func matchGlobalMethodCall(callNode *ast.Node, ctx rule.RuleContext, objectName, methodName string) bool {
	callee := ast.SkipParentheses(callNode.Expression())
	if callee == nil {
		return false
	}

	var object *ast.Node
	var name string
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		object = callee.Expression()
		if property := callee.Name(); property != nil && property.Kind == ast.KindIdentifier {
			name = property.AsIdentifier().Text
		}
	case ast.KindElementAccessExpression:
		object = callee.Expression()
		argument := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if value, ok := utils.GetStaticExpressionValue(argument); ok {
			name = value
		}
	default:
		return false
	}

	object = ast.SkipParentheses(object)
	return name == methodName && object != nil && object.Kind == ast.KindIdentifier &&
		object.AsIdentifier().Text == objectName && isGlobalReference(object, ctx)
}

func isPropertyDescriptor(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil {
		return false
	}

	// Direct descriptor: the third argument of Object.defineProperty or
	// Reflect.defineProperty.
	if parent.Kind == ast.KindCallExpression {
		args := parent.Arguments()
		if len(args) >= 3 && ast.SkipParentheses(args[2]) == node &&
			(matchGlobalMethodCall(parent, ctx, "Object", "defineProperty") ||
				matchGlobalMethodCall(parent, ctx, "Reflect", "defineProperty")) {
			return true
		}
	}

	// Nested descriptor: a property value in the second argument of
	// Object.create or Object.defineProperties.
	if parent.Kind != ast.KindPropertyAssignment || ast.SkipParentheses(parent.Initializer()) != node {
		return false
	}
	properties := ast.WalkUpParenthesizedExpressions(parent.Parent)
	if properties == nil || properties.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	call := ast.WalkUpParenthesizedExpressions(properties.Parent)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	args := call.Arguments()
	return len(args) >= 2 && ast.SkipParentheses(args[1]) == properties &&
		(matchGlobalMethodCall(call, ctx, "Object", "create") ||
			matchGlobalMethodCall(call, ctx, "Object", "defineProperties"))
}

func checkPropertyDescriptor(ctx rule.RuleContext, node *ast.Node, opts Options) {
	if !isPropertyDescriptor(node, ctx) {
		return
	}

	for _, property := range node.Properties() {
		if property == nil || property.Name() == nil {
			continue
		}
		name, ok := utils.GetStaticPropertyName(property.Name())
		if !ok || name != "get" {
			continue
		}

		var getter *ast.Node
		switch property.Kind {
		case ast.KindPropertyAssignment:
			getter = ast.SkipParentheses(property.Initializer())
			if getter == nil || (getter.Kind != ast.KindFunctionExpression && getter.Kind != ast.KindArrowFunction) {
				continue
			}
		case ast.KindMethodDeclaration:
			getter = property
		default:
			continue
		}
		reportGetterReturn(ctx, getter, opts)
	}
}

// GetterReturnRule enforces return statements in getters.
var GetterReturnRule = rule.Rule{
	Name:   "getter-return",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindGetAccessor: func(node *ast.Node) {
				reportGetterReturn(ctx, node, opts)
			},
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				checkPropertyDescriptor(ctx, node, opts)
			},
		}
	},
}
