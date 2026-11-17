package unicornutil

import (
	"slices"
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
func ValueTypeAnnotation(ctx rule.RuleContext, declaration *ast.Node) (*ast.Node, bool) {
	return valueTypeAnnotation(ctx, declaration, false)
}

func valueTypeAnnotation(ctx rule.RuleContext, declaration *ast.Node, unicornJSDoc bool) (annotation *ast.Node, optional bool) {
	annotation = ast.GetTypeAnnotationNode(declaration)
	optional = ast.HasQuestionToken(declaration)
	jsdocAnnotation := utils.IsJSDocSyntaxNode(annotation)
	if unicornJSDoc && (annotation == nil || jsdocAnnotation) {
		return unicornJSDocAnnotation(ctx, declaration, annotation, optional)
	}
	if annotation != nil && (declaration.Kind != ast.KindParameter || !jsdocAnnotation) {
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

// unicornJSDocAnnotation reads a variable's @type or a simple parameter's
// inline @type or named @param as Unicorn does: only the closest attached
// documentation counts, and the applicable tag must be unique. reparsed is
// tsgo's annotation, kept when it comes from the selected tag.
func unicornJSDocAnnotation(ctx rule.RuleContext, declaration, reparsed *ast.Node, optional bool) (*ast.Node, bool) {
	var tag *ast.Node
	var source string
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		var tags []*ast.Node
		tags, source = closestJSDocTags(ctx, declaration)
		tag = uniqueJSDocTag(tags, ast.IsJSDocTypeTag)
	case ast.KindParameter:
		name := declaration.Name()
		if !ast.IsIdentifier(name) {
			return nil, optional
		}
		tags, namedSource := closestJSDocTags(ctx, declaration.Parent)
		named := uniqueJSDocTag(tags, func(tag *ast.Node) bool {
			return ast.IsJSDocParameterTag(tag) && jsDocParameterTagName(tag, namedSource, declaration) == name.Text()
		})
		if named != nil {
			tag, source = named, namedSource
			optional = optional || adaptClosureCallableTag(named, namedSource, declaration).AsJSDocParameterOrPropertyTag().IsBracketed
		}
		tags, inlineSource := closestJSDocTags(ctx, declaration)
		if inline := uniqueJSDocTag(tags, ast.IsJSDocTypeTag); inline != nil && inline.TypeExpression() != nil {
			tag, source = inline, inlineSource
		}
	}
	if tag == nil || tag.TypeExpression() == nil {
		return nil, optional
	}
	if adapted := adaptClosureCallableTag(tag, source, declaration); adapted != tag {
		return adapted.TypeExpression().Type(), optional
	}
	expression := tag.TypeExpression().Type()
	if reparsed != nil && source == ctx.SourceFile.Text() && reparsed.Pos() == expression.Pos() {
		return reparsed, optional
	}
	return expression, optional
}

// jsDocParameterTagName reads a @param name, which tsgo leaves empty after a
// type-first Closure function type.
func jsDocParameterTagName(tag *ast.Node, source string, declaration *ast.Node) string {
	name := tag.Name()
	if name.Text() == "" && tag.TypeExpression() != nil {
		name = adaptClosureCallableTag(tag, source, declaration).Name()
	}
	if !ast.IsIdentifier(name) {
		return ""
	}
	return name.Text()
}

// UnmodifiedFunctionValue finds a local callable whose authored return type can
// be used without a checker. A nil function with ok set can still have a JSDoc
// callable annotation on the binding. Pass a symbol with one declaration.
func UnmodifiedFunctionValue(ctx rule.RuleContext, symbol *ast.Symbol) (function *ast.Node, ok bool) {
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindFunctionDeclaration && declaration.Type() != nil && !utils.IsJSDocSyntaxNode(declaration.Type()) {
		return nil, false
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if reference != declaration.Name() && utils.IsWriteReference(reference) {
			return nil, false
		}
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		function = declaration
	case ast.KindVariableDeclaration:
		initializer := utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer)
		for initializer != nil && (initializer.Kind == ast.KindSatisfiesExpression || initializer.Kind == ast.KindNonNullExpression) {
			initializer = utils.ESTreeRuntimeExpression(initializer.Expression())
		}
		if initializer != nil && (initializer.Kind == ast.KindFunctionExpression || initializer.Kind == ast.KindArrowFunction) {
			function = initializer
		}
	}
	if function != nil && (isAsyncOrGenerator(function) || len(function.TypeParameters()) > 0) {
		return nil, false
	}
	return function, true
}

func isAsyncOrGenerator(function *ast.Node) bool {
	body := function.BodyData()
	return ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) || (body != nil && body.AsteriskToken != nil)
}

// classifyCallableReturnType handles JSDoc function types and call signatures.
func classifyCallableReturnType(annotation *ast.Node, classify func(*ast.Node) TypeClass) TypeClass {
	if annotation == nil {
		return TypeUnknown
	}
	annotation = ast.SkipTypeParentheses(annotation)
	switch annotation.Kind {
	case ast.KindJSDocNonNullableType:
		return classifyCallableReturnType(annotation.Type(), classify)
	case ast.KindFunctionType:
		return classify(annotation.Type())
	case ast.KindTypeLiteral:
		var classes []TypeClass
		for _, member := range annotation.AsTypeLiteralNode().Members.Nodes {
			if member.Kind == ast.KindCallSignature {
				classes = append(classes, classify(member.Type()))
			}
		}
		if len(classes) > 0 {
			return combineTypeClassUnion(classes)
		}
	case ast.KindUnionType:
		var classes []TypeClass
		for _, member := range annotation.AsUnionTypeNode().Types.Nodes {
			classes = append(classes, classifyCallableReturnType(member, classify))
		}
		if len(classes) > 0 {
			return combineTypeClassUnion(classes)
		}
	case ast.KindIntersectionType:
		var classes []TypeClass
		var visit func(*ast.Node)
		visit = func(member *ast.Node) {
			member = ast.SkipTypeParentheses(member)
			if member.Kind == ast.KindJSDocNonNullableType {
				visit(member.Type())
				return
			}
			if member.Kind == ast.KindIntersectionType {
				for _, part := range member.AsIntersectionTypeNode().Types.Nodes {
					visit(part)
				}
				return
			}
			if member.Kind == ast.KindTypeLiteral && !slices.ContainsFunc(member.AsTypeLiteralNode().Members.Nodes, func(part *ast.Node) bool { return part.Kind == ast.KindCallSignature }) {
				return
			}
			classes = append(classes, classifyCallableReturnType(member, classify))
		}
		visit(annotation)
		if len(classes) > 0 {
			return combineTypeClassUnion(classes)
		}
	}
	return TypeUnknown
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
	return callableReturnTypeAnnotation(ctx, declaration, false)
}

// callableReturnTypeAnnotation applies Unicorn's closest-documentation and
// unique-tag policies to JSDoc when unicornJSDoc is set.
func callableReturnTypeAnnotation(ctx rule.RuleContext, declaration *ast.Node, unicornJSDoc bool) *ast.Node {
	if unicornJSDoc {
		return unicornCallableReturnType(ctx, declaration)
	}
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

// unicornCallableReturnType reads a callable's documented return type from the
// unique applicable tag in its closest documentation: a function's @returns,
// or a binding's or parameter's callable annotation.
func unicornCallableReturnType(ctx rule.RuleContext, declaration *ast.Node) *ast.Node {
	if ast.IsParameterDeclaration(declaration) || ast.IsVariableDeclaration(declaration) {
		annotation, optional := valueTypeAnnotation(ctx, declaration, true)
		if optional {
			return nil
		}
		return callableBindingReturnType(annotation)
	}
	reparsed := declaration.Type()
	if reparsed != nil && !utils.IsJSDocSyntaxNode(reparsed) {
		return reparsed
	}
	tags, source := closestJSDocTags(ctx, declaration)
	tag := uniqueJSDocTag(tags, ast.IsJSDocReturnTag)
	if tag == nil || tag.TypeExpression() == nil {
		return nil
	}
	expression := tag.TypeExpression().Type()
	if reparsed != nil && source == ctx.SourceFile.Text() && reparsed.Pos() == expression.Pos() {
		return reparsed
	}
	return expression
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
	case ast.KindJSDocNonNullableType:
		return collectJSDocCallableReturns(node.Type(), returns, inIntersection)
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
	raw := source[expression.Pos():tag.End()]
	if !strings.Contains(raw, "function") || !containsClosureCallableType(expression.Type()) {
		return tag
	}
	end := closureTypeEnd(raw)
	if end == 0 {
		return tag
	}
	// Call signatures preserve the grouping of Closure callables under ? and !.
	normalized, changed := normalizeClosureFunctionTypes(raw[1 : end-1])
	if !changed {
		return tag
	}
	comment := "/** " + source[tag.Pos():expression.Pos()] + "{" + normalized + raw[end-1:] + " */"
	for _, doc := range parseCallableJSDoc(comment, scope) {
		if tags := doc.AsJSDoc().Tags; tags != nil && len(tags.Nodes) != 0 {
			if parsed := tags.Nodes[0]; parsed.Kind == tag.Kind && parsed.TypeExpression() != nil {
				return parsed
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
