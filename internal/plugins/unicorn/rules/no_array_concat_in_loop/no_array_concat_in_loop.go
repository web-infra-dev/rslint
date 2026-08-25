package no_array_concat_in_loop

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-array-concat-in-loop"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not use `Array#concat()` to accumulate an array in a loop.",
}

var NoArrayConcatInLoopRule = rule.Rule{
	Name:   "unicorn/no-array-concat-in-loop",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				checkAssignment(ctx, node)
			},
		}
	},
}

func checkAssignment(ctx rule.RuleContext, node *ast.Node) {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil ||
		binary.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	left := ast.SkipParentheses(binary.Left)
	if left == nil || !ast.IsIdentifier(left) {
		return
	}

	right := utils.SkipAssertionsAndParens(binary.Right)
	minimumArguments := 1
	call, ok := unicornutil.MatchDotMethodCall(right, unicornutil.DotMethodCallOptions{
		Method:           "concat",
		MinimumArguments: &minimumArguments,
	})
	if !ok {
		return
	}

	receiver := utils.SkipAssertionsAndParens(call.Object)
	if receiver == nil || !ast.IsIdentifier(receiver) || ctx.Refs == nil {
		return
	}

	variable := ctx.Refs.Resolve(left)
	if variable == nil || variable != ctx.Refs.Resolve(receiver) {
		return
	}
	if isGlobalScopeVariable(ctx, variable) {
		return
	}

	declaration := mutableEmptyArrayDeclaration(variable)
	if declaration == nil {
		return
	}

	loop := nearestLoop(node)
	if loop == nil || nodeInside(declaration, loopBody(loop)) {
		return
	}

	ctx.ReportNode(call.Property, message)
}

func isGlobalScopeVariable(ctx rule.RuleContext, variable *ast.Symbol) bool {
	return ctx.SourceFile != nil && !ctx.Refs.HasNonGlobalProgramScope() &&
		ctx.SourceFile.Locals[variable.Name] == variable
}

func mutableEmptyArrayDeclaration(variable *ast.Symbol) *ast.Node {
	if variable == nil || len(variable.Declarations) != 1 {
		return nil
	}

	declarationNode := variable.Declarations[0]
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return nil
	}
	declaration := declarationNode.AsVariableDeclaration()
	if declaration == nil || declaration.Name() == nil ||
		!ast.IsIdentifier(declaration.Name()) || declaration.Initializer == nil {
		return nil
	}

	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	kind := utils.GetVarDeclListKind(declarationList)
	if kind != "let" && kind != "var" {
		return nil
	}

	initializer := utils.SkipAssertionsAndParens(declaration.Initializer)
	if initializer == nil || !ast.IsEmptyArrayLiteral(initializer) {
		return nil
	}
	return declarationNode
}

func nearestLoop(node *ast.Node) *ast.Node {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ast.IsFunctionLike(ancestor) {
			return nil
		}
		if ast.IsIterationStatement(ancestor, false) {
			return ancestor
		}
	}
	return nil
}

func loopBody(loop *ast.Node) *ast.Node {
	if loop == nil {
		return nil
	}
	switch loop.Kind {
	case ast.KindForStatement:
		return loop.AsForStatement().Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		return loop.AsForInOrOfStatement().Statement
	case ast.KindWhileStatement:
		return loop.AsWhileStatement().Statement
	case ast.KindDoStatement:
		return loop.AsDoStatement().Statement
	default:
		return nil
	}
}

func nodeInside(node *ast.Node, ancestor *ast.Node) bool {
	return node != nil && ancestor != nil &&
		node.Pos() >= ancestor.Pos() && node.End() <= ancestor.End()
}
