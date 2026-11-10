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
	parameter := ast.IsParameterDeclaration(declaration)
	host := declaration
	if parameter {
		var optional bool
		annotation, optional = ValueTypeAnnotation(ctx, declaration)
		if optional {
			return nil
		}
		if annotation != nil && !isClosureCallableType(annotation) {
			return callableBindingReturnType(annotation)
		}
		host = declaration.Parent
	} else if ast.IsVariableDeclaration(declaration) {
		if result := callableBindingReturnType(annotation); result != nil {
			return result
		}
	} else if annotation != nil {
		return annotation
	}
	for ; host != nil; host = ast.GetNextJSDocCommentLocation(host) {
		for _, doc := range host.JSDoc(ctx.SourceFile) {
			if doc.AsJSDoc().Tags == nil {
				continue
			}
			for _, tag := range doc.AsJSDoc().Tags.Nodes {
				if tag.TypeExpression() == nil {
					continue
				}
				tagType := tag.TypeExpression().Type()
				if parameter {
					if tag.Kind != ast.KindJSDocParameterTag || !isClosureCallableType(tagType) {
						continue
					}
					if name := tag.Name().Text(); name != "" && name != declaration.Name().Text() {
						continue
					}
					if result := closureCallableReturnTypeAnnotation(ctx, tag, declaration); result != nil {
						return result
					}
				} else if ast.IsJSDocReturnTag(tag) {
					return tagType
				} else if ast.IsJSDocTypeTag(tag) {
					if result := callableBindingReturnType(tagType); result != nil {
						return result
					}
					if isClosureCallableType(tagType) {
						if result := closureCallableReturnTypeAnnotation(ctx, tag, declaration); result != nil {
							return result
						}
					}
				}
			}
		}
	}
	return nil
}

func callableBindingReturnType(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipTypeParentheses(node)
	if ast.IsFunctionTypeNode(node) {
		return node.Type()
	}
	if !utils.IsJSDocSyntaxNode(node) {
		return nil
	}
	var returns []*ast.Node
	if !collectJSDocCallableReturns(node, &returns, false) || len(returns) == 0 {
		return nil
	}
	if len(returns) == 1 {
		return returns[0]
	}
	// Any overload may supply the result, so classify these as a union.
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	return factory.NewUnionTypeNode(factory.NewNodeList(returns))
}

func collectJSDocCallableReturns(node *ast.Node, returns *[]*ast.Node, inIntersection bool) bool {
	if node == nil {
		return false
	}
	node = ast.SkipTypeParentheses(node)
	switch node.Kind {
	case ast.KindFunctionType:
		if result := node.Type(); result != nil {
			*returns = append(*returns, result)
			return true
		}
	case ast.KindUnionType, ast.KindIntersectionType:
		intersection := node.Kind == ast.KindIntersectionType
		start := len(*returns)
		invalid := node.ForEachChild(func(child *ast.Node) bool {
			return !collectJSDocCallableReturns(child, returns, intersection)
		})
		return !invalid && (len(*returns) > start || (intersection && inIntersection))
	case ast.KindTypeLiteral:
		members := node.AsTypeLiteralNode().Members.Nodes
		start := len(*returns)
		for _, member := range members {
			if member.Kind == ast.KindCallSignature {
				if member.Type() == nil {
					return false
				}
				*returns = append(*returns, member.Type())
			}
		}
		return len(*returns) > start || (inIntersection && len(members) > 0)
	}
	return false
}

func isClosureCallableType(node *ast.Node) bool {
	return node != nil && ast.IsTypeReferenceNode(node) && ast.IsIdentifier(node.AsTypeReferenceNode().TypeName) &&
		node.AsTypeReferenceNode().TypeName.Text() == "function"
}

func closureCallableReturnTypeAnnotation(ctx rule.RuleContext, tag, declaration *ast.Node) *ast.Node {
	// tsgo parses Closure's `function(): T` only as the name `function`.
	// Add an opening brace so tsgo reads it as a type-literal method. The
	// original closing brace ends the literal and JSDoc parsing recovers the
	// rest of the tag, including the parameter name.
	source := ctx.SourceFile.Text()
	start := tag.TypeExpression().Pos()
	text := "/** " + source[tag.Pos():start] + "{" + source[start:tag.End()] + " */ const factory = undefined;"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/annotation.js", Path: "/annotation.js"}, text, core.ScriptKindJS)
	docs := file.Statements.Nodes[0].JSDoc(file)
	if len(docs) == 0 || docs[0].AsJSDoc().Tags == nil || len(docs[0].AsJSDoc().Tags.Nodes) == 0 {
		return nil
	}
	parsed := docs[0].AsJSDoc().Tags.Nodes[0]
	if parsed.Kind != tag.Kind || parsed.TypeExpression() == nil {
		return nil
	}
	if parsed.Kind == ast.KindJSDocParameterTag {
		param := parsed.AsJSDocParameterOrPropertyTag()
		if param.IsBracketed || param.Name().Text() != declaration.Name().Text() {
			return nil
		}
	}
	annotation := parsed.TypeExpression().Type()
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
