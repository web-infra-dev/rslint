package only_throw_error

import (
	"encoding/json"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type OnlyThrowErrorOptions struct {
	Allow                []utils.TypeOrValueSpecifier `json:"allow"`
	AllowRethrowing      *bool                        `json:"allowRethrowing"`
	AllowThrowingAny     *bool                        `json:"allowThrowingAny"`
	AllowThrowingUnknown *bool                        `json:"allowThrowingUnknown"`
}

func buildObjectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "object",
		Description: "Expected an error object to be thrown.",
	}
}
func buildUndefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "undef",
		Description: "Do not throw undefined.",
	}
}

// isRethrownError checks if the thrown expression is a variable caught in a
// catch clause or as a parameter of a Promise .catch()/.then() rejection handler.
func isRethrownError(ctx rule.RuleContext, expr *ast.Node) bool {
	if !ast.IsIdentifier(expr) {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(expr)
	if symbol == nil {
		return false
	}

	decl := symbol.ValueDeclaration
	if decl == nil {
		return false
	}

	// Case 1: try { ... } catch (e) { throw e; }
	// In TS compiler AST, the catch variable's ValueDeclaration is a
	// VariableDeclaration whose parent is the CatchClause.
	if decl.Parent != nil && decl.Parent.Kind == ast.KindCatchClause {
		return true
	}

	// Case 2: promise.catch(e => { throw e; })
	//         promise.then(onFulfilled, e => { throw e; })
	// The parameter's ValueDeclaration is a Parameter node.
	if decl.Kind != ast.KindParameter {
		return false
	}

	paramDecl := decl.AsParameterDeclaration()
	// Rest parameters are not simple rethrows: (...e) => { throw e; }
	if paramDecl.DotDotDotToken != nil {
		return false
	}

	// The parameter's parent should be an arrow function
	funcNode := decl.Parent
	if funcNode == nil || funcNode.Kind != ast.KindArrowFunction {
		return false
	}

	// The parameter must be the first parameter of the arrow function
	params := funcNode.Parameters()
	if len(params) == 0 || params[0] != decl {
		return false
	}

	// The arrow function must be a direct argument to a call expression.
	// Parent pointers: argument.Parent = CallExpression (set by parser's finishNode).
	callNode := funcNode.Parent
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return false
	}

	callExpr := callNode.AsCallExpression()
	callee := callExpr.Expression

	if !ast.IsAccessExpression(callee) {
		return false
	}

	propertyName, found := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
	if !found {
		return false
	}

	if callExpr.Arguments == nil {
		return false
	}
	args := callExpr.Arguments.Nodes
	var onRejected *ast.Node

	switch propertyName {
	case "catch":
		// promise.catch(handler)
		// If first arg is spread, we can't determine the handler
		if len(args) >= 1 && !ast.IsSpreadElement(args[0]) {
			onRejected = args[0]
		}
	case "then":
		// promise.then(onFulfilled, handler)
		// Need at least 2 args, and neither of the first two can be spread
		if len(args) >= 2 && !ast.IsSpreadElement(args[0]) && !ast.IsSpreadElement(args[1]) {
			onRejected = args[1]
		}
	default:
		return false
	}

	// The arrow function must be the onRejected handler
	if onRejected != funcNode {
		return false
	}

	// Verify the object is actually thenable (has a `then` method)
	objectNode := callee.Expression()
	objectType := ctx.TypeChecker.GetTypeAtLocation(objectNode)
	return utils.IsThenableType(ctx.TypeChecker, objectNode, objectType)
}

var OnlyThrowErrorRule = rule.CreateRule(rule.Rule{
	Name: "only-throw-error",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(OnlyThrowErrorOptions)
		if !ok {
			opts = OnlyThrowErrorOptions{}
			// When options come from JSON (API/TS tests), they arrive as []interface{}
			if optionsArray, arrayOk := options.([]interface{}); arrayOk && len(optionsArray) > 0 {
				if optsJSON, err := json.Marshal(optionsArray[0]); err == nil {
					json.Unmarshal(optsJSON, &opts)
				}
			}
		}
		if opts.Allow == nil {
			opts.Allow = []utils.TypeOrValueSpecifier{}
		}
		if opts.AllowRethrowing == nil {
			opts.AllowRethrowing = utils.Ref(true)
		}
		if opts.AllowThrowingAny == nil {
			opts.AllowThrowingAny = utils.Ref(true)
		}
		if opts.AllowThrowingUnknown == nil {
			opts.AllowThrowingUnknown = utils.Ref(true)
		}

		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				expr := node.Expression()

				if *opts.AllowRethrowing && isRethrownError(ctx, expr) {
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(expr)

				if utils.TypeMatchesSomeSpecifier(t, opts.Allow, nil, ctx.Program) {
					return
				}

				if utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
					ctx.ReportNode(expr, buildUndefMessage())
					return
				}

				if *opts.AllowThrowingAny && utils.IsTypeAnyType(t) {
					return
				}

				if *opts.AllowThrowingUnknown && utils.IsTypeUnknownType(t) {
					return
				}

				if utils.IsErrorLike(ctx.Program, ctx.TypeChecker, t) {
					return
				}

				ctx.ReportNode(expr, buildObjectMessage())
			},
		}
	},
})
