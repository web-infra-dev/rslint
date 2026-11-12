package no_unnecessary_string_trim

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// hasNonStringJSDocReceiver checks nullable and Closure-style annotations on
// receivers or factory return types.
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
	return nonStringJSDocType(annotation) || hasClosureCallableValue(ctx, declaration, annotation)
}

func hasClosureCallableValue(ctx rule.RuleContext, declaration, annotation *ast.Node) bool {
	source := ctx.SourceFile.Text()
	if ast.IsVariableDeclaration(declaration) {
		return isClosureCallableType(source, annotation, "")
	}
	parameter := declaration.Name()
	if !ast.IsIdentifier(parameter) {
		return false
	}
	name := parameter.Text()
	for host := declaration.Parent; host != nil; host = ast.GetNextJSDocCommentLocation(host) {
		for _, doc := range host.JSDoc(ctx.SourceFile) {
			tags := doc.AsJSDoc().Tags
			if tags == nil {
				continue
			}
			for _, tag := range tags.Nodes {
				if !ast.IsJSDocParameterTag(tag) || tag.TypeExpression() == nil || !ast.IsIdentifier(tag.Name()) {
					continue
				}
				switch tag.Name().Text() {
				case name:
					return isClosureCallableType(source[:tag.End()], tag.TypeExpression().Type(), "")
				case "":
					// A truncated type-first tag can lose the parameter name.
					if isClosureCallableType(source[:tag.End()], tag.TypeExpression().Type(), name) {
						return true
					}
				}
			}
		}
	}
	return false
}

func isClosureCallableType(source string, annotation *ast.Node, trailingName string) bool {
	// tsgo can stop reading the annotation at "function". Read the rest far
	// enough to distinguish a callable value from a union that includes strings.
	if !utils.IsJSDocSyntaxNode(annotation) {
		return false
	}
	groups := 0
unwrap:
	for annotation != nil {
		switch annotation.Kind {
		case ast.KindParenthesizedType:
			groups++
		case ast.KindJSDocNullableType, ast.KindJSDocOptionalType, ast.KindJSDocNonNullableType:
		default:
			break unwrap
		}
		annotation = annotation.Type()
	}
	if !ast.IsTypeReferenceNode(annotation) {
		return false
	}
	name := annotation.AsTypeReferenceNode().TypeName
	if !ast.IsIdentifier(name) || name.Text() != "function" {
		return false
	}
	tail, _, _ := strings.Cut(source[annotation.End():], "*/")
	scan := scanner.NewScanner()
	scan.SetText(tail)
	scan.SetSkipJSDocLeadingAsterisks(true)
	if scan.Scan() != ast.KindOpenParenToken {
		return false
	}
	depth := 1
	for depth > 0 {
		switch scan.Scan() {
		case ast.KindEndOfFile:
			return false
		case ast.KindOpenParenToken:
			depth++
		case ast.KindCloseParenToken:
			depth--
		}
	}
	token := scan.Scan()
	if token == ast.KindColonToken {
		// The return type extends to the enclosing group or annotation.
		for token = scan.Scan(); token != ast.KindEndOfFile; token = scan.Scan() {
			if depth == 0 && (token == ast.KindCloseParenToken || token == ast.KindCloseBraceToken) {
				break
			}
			switch token {
			case ast.KindOpenParenToken, ast.KindOpenBraceToken:
				depth++
			case ast.KindCloseParenToken, ast.KindCloseBraceToken:
				depth--
			}
		}
	}
	for groups > 0 && token == ast.KindCloseParenToken {
		groups--
		token = scan.Scan()
	}
	if token == ast.KindEqualsToken {
		token = scan.Scan()
	}
	if groups != 0 || token != ast.KindCloseBraceToken {
		return false
	}
	if trailingName != "" {
		if scan.Scan() == ast.KindOpenBracketToken {
			scan.Scan()
		}
		return scan.TokenValue() == trailingName
	}
	return true
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
