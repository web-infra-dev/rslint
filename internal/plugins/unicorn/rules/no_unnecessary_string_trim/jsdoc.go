package no_unnecessary_string_trim

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// hasNonStringJSDocReceiver checks explicit JSDoc types that rule out a string
// even when a receiver annotation is nullable or a factory uses Closure syntax.
func hasNonStringJSDocReceiver(ctx rule.RuleContext, receiver *ast.Node) bool {
	if !ast.IsInJSFile(receiver) {
		return false
	}
	call := ast.IsCallExpression(receiver)
	reference := receiver
	if call {
		reference = utils.ESTreeCallCallee(receiver.Expression())
	}
	symbol := ctx.Refs.ResolveInFile(reference)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	binding := ast.IsParameterDeclaration(declaration) || ast.IsVariableDeclaration(declaration)
	if call {
		if !binding && !ast.IsFunctionDeclaration(declaration) {
			return false
		}
		return nonStringJSDocType(unicornutil.CallableReturnTypeAnnotation(ctx, declaration)) &&
			!slices.ContainsFunc(ctx.Refs.References(symbol), utils.IsWriteReference)
	}
	if !binding {
		return false
	}
	annotation, _ := unicornutil.ValueTypeAnnotation(ctx, declaration)
	return nonStringJSDocType(annotation)
}

func nonStringJSDocType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipTypeParentheses(node)
	switch node.Kind {
	case ast.KindJSDocNullableType, ast.KindJSDocOptionalType, ast.KindJSDocNonNullableType:
		return nonStringJSDocType(node.Type())
	case ast.KindUnionType:
		return utils.Every(node.AsUnionTypeNode().Types.Nodes, nonStringJSDocType)
	case ast.KindLiteralType:
		literal := node.AsLiteralTypeNode().Literal
		return literal != nil && !ast.IsStringLiteralLike(literal)
	case ast.KindTypeLiteral, ast.KindJSDocTypeLiteral, ast.KindObjectKeyword,
		ast.KindArrayType, ast.KindTupleType, ast.KindFunctionType, ast.KindConstructorType,
		ast.KindNumberKeyword, ast.KindBooleanKeyword, ast.KindBigIntKeyword, ast.KindSymbolKeyword,
		ast.KindNullKeyword, ast.KindUndefinedKeyword, ast.KindVoidKeyword, ast.KindNeverKeyword:
		return true
	default:
		return false
	}
}
