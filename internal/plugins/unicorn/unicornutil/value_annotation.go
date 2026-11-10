package unicornutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ValueTypeAnnotation reads authored annotations through tsgo's AST, retaining
// JSDoc unknown types and optional parameter metadata so static folding cannot
// bypass them.
func ValueTypeAnnotation(ctx rule.RuleContext, declaration *ast.Node) (annotation *ast.Node, optional bool) {
	annotation = ast.GetTypeAnnotationNode(declaration)
	optional = ast.HasQuestionToken(declaration)
	if annotation != nil && (declaration.Kind != ast.KindParameter || !utils.IsJSDocSyntaxNode(annotation)) {
		return annotation, optional
	}
	host := declaration
	if declaration.Kind == ast.KindParameter {
		host = declaration.Parent
	}
	for host != nil {
		for _, doc := range host.JSDoc(ctx.SourceFile) {
			if tags := doc.AsJSDoc().Tags; tags != nil {
				for _, tag := range tags.Nodes {
					var expression *ast.Node
					switch tag.Kind {
					case ast.KindJSDocParameterTag:
						parameter := tag.AsJSDocParameterOrPropertyTag()
						if declaration.Kind == ast.KindParameter && ast.IsIdentifier(declaration.Name()) && parameter.Name().Text() == declaration.Name().Text() {
							expression = parameter.TypeExpression
							optional = optional || parameter.IsBracketed
						}
					case ast.KindJSDocTypeTag:
						if declaration.Kind == ast.KindVariableDeclaration {
							expression = tag.AsJSDocTypeTag().TypeExpression
						}
					case ast.KindJSDocReturnTag:
						if ast.IsFunctionLike(declaration) {
							expression = tag.AsJSDocReturnTag().TypeExpression
						}
					}
					if expression != nil {
						if annotation == nil {
							annotation = expression.Type()
						}
						return annotation, optional
					}
				}
			}
		}
		host = ast.GetNextJSDocCommentLocation(host)
	}
	return annotation, optional
}

// HasJSDocTypeReference reports an authored JSDoc type through an alias or a
// larger expression. Traverse runtime children using tsgo and the reference
// store, excluding property names, type syntax, and unevaluated function bodies.
func HasJSDocTypeReference(ctx rule.RuleContext, node *ast.Node, visited map[*ast.Symbol]bool) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) || ast.IsFunctionLike(node) || ast.IsClassLike(node) {
		return false
	}
	if ast.IsIdentifier(node) {
		if ctx.Refs == nil || utils.IsNonReferenceIdentifier(node) {
			return false
		}
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if annotation, _ := ValueTypeAnnotation(ctx, declaration); annotation != nil && utils.IsJSDocSyntaxNode(annotation) {
			return true
		}
		if visited[symbol] {
			return false
		}
		visited[symbol] = true
		return declaration.Kind == ast.KindVariableDeclaration && HasJSDocTypeReference(ctx, declaration.AsVariableDeclaration().Initializer, visited)
	}
	return node.ForEachChild(func(child *ast.Node) bool { return HasJSDocTypeReference(ctx, child, visited) })
}

// CallableReturnTypeAnnotation reads the return type authored on a function
// or on a binding annotated as a function.
func CallableReturnTypeAnnotation(ctx rule.RuleContext, declaration *ast.Node) *ast.Node {
	annotation := declaration.Type()
	if (ast.IsVariableDeclaration(declaration) || ast.IsParameterDeclaration(declaration)) && annotation != nil {
		annotation = ast.SkipTypeParentheses(annotation)
		if ast.IsFunctionTypeNode(annotation) {
			annotation = annotation.Type()
		} else {
			annotation = nil
		}
	}
	if annotation == nil && ast.IsParameterDeclaration(declaration) {
		if documented, optional := ValueTypeAnnotation(ctx, declaration); documented != nil && !optional {
			if function := ast.SkipTypeParentheses(documented); ast.IsFunctionTypeNode(function) {
				return function.Type()
			}
		}
	}
	if annotation == nil {
		for host := declaration; host != nil && annotation == nil; host = ast.GetNextJSDocCommentLocation(host) {
			for _, doc := range host.JSDoc(ctx.SourceFile) {
				if doc.AsJSDoc().Tags == nil {
					continue
				}
				for _, tag := range doc.AsJSDoc().Tags.Nodes {
					if tag.TypeExpression() == nil {
						continue
					}
					tagType := tag.TypeExpression().Type()
					if ast.IsJSDocReturnTag(tag) {
						annotation = tagType
					} else if ast.IsJSDocTypeTag(tag) && ast.IsFunctionTypeNode(tagType) {
						annotation = tagType.Type()
					} else if ast.IsJSDocTypeTag(tag) && ast.IsTypeReferenceNode(tagType) && ast.IsIdentifier(tagType.AsTypeReferenceNode().TypeName) && tagType.AsTypeReferenceNode().TypeName.Text() == "function" {
						annotation = closureCallableReturnTypeAnnotation(ctx, tag, declaration)
					}
					if annotation != nil {
						break
					}
				}
			}
		}
	}
	return annotation
}

func closureCallableReturnTypeAnnotation(ctx rule.RuleContext, tag, declaration *ast.Node) *ast.Node {
	// tsgo parses Closure's `function(): T` only as the name `function`.
	// Wrapping it as a type-literal method lets its JSDoc parser read T,
	// including multiline comments, without implementing a type parser here.
	text := "/** @type {" + ctx.SourceFile.Text()[tag.TypeExpression().Pos():tag.End()] + "} */ const factory = undefined;"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/annotation.js", Path: "/annotation.js"}, text, core.ScriptKindJS)
	annotation := file.Statements.Nodes[0].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Type()
	if annotation == nil || annotation.Kind != ast.KindTypeLiteral {
		return nil
	}
	members := annotation.AsTypeLiteralNode().Members.Nodes
	if len(members) != 1 || members[0].Kind != ast.KindMethodSignature || members[0].Name().Text() != "function" {
		return nil
	}
	result := members[0].Type()
	if result != nil {
		// Resolve return-type names in the original factory's lexical scope.
		result.Parent = declaration
	}
	return result
}
