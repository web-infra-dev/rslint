package unicornutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
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
		if result := callableBindingReturnType(annotation); result != nil {
			return result
		}
		if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) {
			return nil
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
		docs, source := callableJSDoc(ctx, host)
		for _, doc := range docs {
			if doc.AsJSDoc().Tags == nil {
				continue
			}
			for _, tag := range doc.AsJSDoc().Tags.Nodes {
				if tag.TypeExpression() == nil {
					continue
				}
				if parameter {
					if tag.Kind != ast.KindJSDocParameterTag {
						continue
					}
					if name := tag.Name().Text(); name != "" && name != declaration.Name().Text() {
						continue
					}
					tag = adaptClosureCallableTag(tag, source, declaration)
					param := tag.AsJSDocParameterOrPropertyTag()
					if !param.IsBracketed && param.Name().Text() == declaration.Name().Text() {
						return callableBindingReturnType(param.TypeExpression.Type())
					}
				} else if ast.IsJSDocReturnTag(tag) {
					return tag.TypeExpression().Type()
				} else if ast.IsJSDocTypeTag(tag) {
					tag = adaptClosureCallableTag(tag, source, declaration)
					if result := callableBindingReturnType(tag.TypeExpression().Type()); result != nil {
						return result
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

func callableJSDoc(ctx rule.RuleContext, host *ast.Node) ([]*ast.Node, string) {
	source := ctx.SourceFile.Text()
	if docs := host.JSDoc(ctx.SourceFile); len(docs) != 0 {
		return docs, source
	}
	// tsgo may attach a same-line JSDoc as trailing trivia of the previous
	// statement. Use the closest comment in this node's leading trivia.
	start, end := host.Pos(), scanner.SkipTrivia(source, host.Pos())
	if !strings.Contains(source[start:end], "/**") {
		return nil, ""
	}
	comments := utils.CommentsInSpan(ctx.Comments.All(), start, end)
	if len(comments) == 0 {
		return nil, ""
	}
	comment := comments[len(comments)-1]
	if comment.Pos() < start || comment.End() > end || comment.Kind != ast.KindMultiLineCommentTrivia {
		return nil, ""
	}
	text := source[comment.Pos():comment.End()]
	return parseCallableJSDoc(text, host), text
}

func parseCallableJSDoc(comment string, scope *ast.Node) []*ast.Node {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/annotation.js", Path: "/annotation.js"},
		comment+"\nconst documented = 0;", core.ScriptKindJS)
	docs := file.Statements.Nodes[0].JSDoc(file)
	for _, doc := range docs {
		// Resolve names in the original lexical scope.
		doc.Parent = scope
	}
	return docs
}

func adaptClosureCallableTag(tag *ast.Node, source string, scope *ast.Node) *ast.Node {
	expression := tag.TypeExpression()
	text := source[expression.Pos():tag.End()]
	if !strings.Contains(text, "function") || !containsClosureCallableType(expression.Type()) {
		return tag
	}
	if rewritten, ok := rewriteClosureCallables(text); ok {
		comment := "/** " + source[tag.Pos():expression.Pos()] + rewritten + " */"
		for _, doc := range parseCallableJSDoc(comment, scope) {
			if tags := doc.AsJSDoc().Tags; tags != nil && len(tags.Nodes) != 0 {
				if parsed := tags.Nodes[0]; parsed.Kind == tag.Kind && parsed.TypeExpression() != nil {
					return parsed
				}
			}
		}
	}
	return tag
}

func containsClosureCallableType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsTypeReferenceNode(node) {
		name := node.AsTypeReferenceNode().TypeName
		if ast.IsIdentifier(name) && name.Text() == "function" {
			return true
		}
	}
	return node.ForEachChild(containsClosureCallableType)
}

// tsgo stops at Closure's `function` keyword in a type. Adapt callable
// members to arrow syntax and let tsgo parse the enclosing type and tag.
func rewriteClosureCallables(text string) (string, bool) {
	tokens := scanner.NewScanner()
	tokens.SetText(text)
	tokens.SetSkipJSDocLeadingAsterisks(true)
	if tokens.Scan() != ast.KindOpenBraceToken {
		return "", false
	}
	var result strings.Builder
	depth, written := 1, 0
	for kind := tokens.Scan(); kind != ast.KindEndOfFile; kind = tokens.Scan() {
		switch kind {
		case ast.KindOpenBraceToken:
			depth++
		case ast.KindCloseBraceToken:
			depth--
			if depth == 0 {
				if written == 0 {
					return "", false
				}
				result.WriteString(text[written:])
				return result.String(), true
			}
		case ast.KindFunctionKeyword:
			// A function-named method inside a record is ordinary type syntax.
			if depth != 1 {
				continue
			}
			start, state := tokens.TokenStart(), tokens.Mark()
			if tokens.Scan() != ast.KindOpenParenToken {
				tokens.Rewind(state)
				continue
			}
			for parentheses := 1; parentheses > 0; {
				switch tokens.Scan() {
				case ast.KindOpenParenToken:
					parentheses++
				case ast.KindCloseParenToken:
					parentheses--
				case ast.KindEndOfFile:
					return "", false
				}
			}
			if tokens.Scan() != ast.KindColonToken {
				return "", false
			}
			result.WriteString(text[written:start])
			result.WriteString("() =>")
			written = tokens.TokenEnd()
		}
	}
	return "", false
}
