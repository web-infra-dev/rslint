package no_arrow_function_lifecycle

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var instanceLifecycleMethods = map[string]bool{
	"getDefaultProps":                  true,
	"getInitialState":                  true,
	"getChildContext":                  true,
	"componentWillMount":               true,
	"UNSAFE_componentWillMount":        true,
	"componentDidMount":                true,
	"componentWillReceiveProps":        true,
	"UNSAFE_componentWillReceiveProps": true,
	"shouldComponentUpdate":            true,
	"componentWillUpdate":              true,
	"UNSAFE_componentWillUpdate":       true,
	"getSnapshotBeforeUpdate":          true,
	"componentDidUpdate":               true,
	"componentDidCatch":                true,
	"componentWillUnmount":             true,
	"render":                           true,
}

var staticLifecycleMethods = map[string]bool{
	"getDerivedStateFromProps": true,
}

const lifecycleMessage = "{{propertyName}} is a React lifecycle method, and should not be an arrow function or in a class field. Use an instance method instead."

func memberName(member *ast.Node) string {
	if member == nil {
		return ""
	}
	name := member.Name()
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		// ESTree keeps an Identifier key for `[render]`, so upstream's
		// `getPropertyName` still returns "render". String and non-identifier
		// computed keys remain unnamed.
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	return reactutil.IdentifierOrPrivateName(name)
}

func memberValue(member *ast.Node) *ast.Node {
	if member == nil {
		return nil
	}
	switch member.Kind {
	case ast.KindPropertyAssignment:
		return member.AsPropertyAssignment().Initializer
	case ast.KindPropertyDeclaration:
		return member.AsPropertyDeclaration().Initializer
	}
	return nil
}

func componentProperties(component *ast.Node) []*ast.Node {
	if component == nil {
		return nil
	}
	switch component.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return component.Members()
	case ast.KindObjectLiteralExpression:
		return component.AsObjectLiteralExpression().Properties.Nodes
	}
	return nil
}

// isCreateReactClassObjectComponent mirrors eslint-plugin-react's
// isES5Component check. The upstream check only inspects the object's direct
// call parent and callee; it does not require the object to be the first
// argument. The shared reactutil helper is intentionally stricter because
// other rules use the first argument as their component-shape boundary.
func isCreateReactClassObjectComponent(obj *ast.Node, pragma, createClass string) bool {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	arg := obj
	for arg.Parent != nil && arg.Parent.Kind == ast.KindParenthesizedExpression {
		arg = arg.Parent
	}
	if arg.Parent == nil {
		return false
	}
	var arguments []*ast.Node
	switch arg.Parent.Kind {
	case ast.KindCallExpression:
		call := arg.Parent.AsCallExpression()
		if !reactutil.IsCreateClassCall(call, pragma, createClass) || call.Arguments == nil {
			return false
		}
		arguments = call.Arguments.Nodes
	case ast.KindNewExpression:
		newExpression := arg.Parent.AsNewExpression()
		if newExpression == nil || !isCreateClassCallee(newExpression.Expression, pragma, createClass) || newExpression.Arguments == nil {
			return false
		}
		arguments = newExpression.Arguments.Nodes
	default:
		return false
	}
	for _, candidate := range arguments {
		if candidate == arg {
			return true
		}
	}
	return false
}

func isCreateClassCallee(callee *ast.Node, pragma, createClass string) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = reactutil.DefaultReactPragma
	}
	if createClass == "" {
		createClass = reactutil.DefaultReactCreateClass
	}
	callee = ast.SkipParentheses(callee)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == createClass
	case ast.KindPropertyAccessExpression:
		propertyAccess := callee.AsPropertyAccessExpression()
		object := ast.SkipParentheses(propertyAccess.Expression)
		name := propertyAccess.Name()
		return object != nil && object.Kind == ast.KindIdentifier && object.AsIdentifier().Text == pragma &&
			name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == createClass
	default:
		return false
	}
}

func isLifecycleMethod(member *ast.Node, name string) bool {
	methods := instanceLifecycleMethods
	if utils.IncludesModifier(member, ast.KindStaticKeyword) {
		methods = staticLifecycleMethods
	}
	return methods[name]
}

// hasExplicitReactComponentJSDoc reports whether a class has an unbraced JSDoc
// `@extends React.Component` / `@augments React.Component` tag. This matches
// eslint-plugin-react's doctrine-based check, whose tag.name is nil for the
// braced `@extends {React.Component}` form.
func hasExplicitReactComponentJSDoc(sf *ast.SourceFile, comments []*ast.CommentRange, classNode *ast.Node) bool {
	if classNode == nil {
		return false
	}
	comment := nearestAttachedJSDocComment(sf, comments, classNode)
	if comment == nil {
		return false
	}
	docs := classNode.JSDoc(nil)
	if len(docs) == 0 {
		return false
	}
	jd := docs[len(docs)-1].AsJSDoc()
	if jd == nil || jd.Tags == nil {
		return false
	}
	for _, tag := range jd.Tags.Nodes {
		if !ast.IsJSDocAugmentsTag(tag) || !isUnbracedJSDocAugmentsTag(sf, tag) {
			continue
		}
		className := tag.AsJSDocAugmentsTag().ClassName
		if className == nil {
			continue
		}
		if isExplicitReactComponentType(className) {
			return true
		}
	}
	return false
}

func nearestAttachedJSDocComment(sf *ast.SourceFile, comments []*ast.CommentRange, node *ast.Node) *ast.CommentRange {
	if sf == nil || node == nil {
		return nil
	}
	text := sf.Text()
	nodeStart := node.Pos()
	sourceScanner := scanner.GetScannerForSourceFile(sf, node.Pos())
	for sourceScanner.Token() != ast.KindEndOfFile && sourceScanner.TokenStart() < node.End() {
		if sourceScanner.Token() != ast.KindOpenParenToken {
			nodeStart = sourceScanner.TokenStart()
			break
		}
		sourceScanner.Scan()
	}
	index := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= nodeStart })
	if index == 0 {
		return nil
	}
	comment := comments[index-1]
	if comment.End() > nodeStart || !ecmascript.IsBlank(text[comment.End():nodeStart]) ||
		comment.Kind != ast.KindMultiLineCommentTrivia || !strings.HasPrefix(text[comment.Pos():comment.End()], "/**") {
		return nil
	}
	lineMap := sf.ECMALineMap()
	if scanner.ComputeLineOfPosition(lineMap, nodeStart)-scanner.ComputeLineOfPosition(lineMap, comment.End()) > 1 {
		return nil
	}
	return comment
}

func isUnbracedJSDocAugmentsTag(sf *ast.SourceFile, tag *ast.Node) bool {
	if sf == nil || tag == nil || !ast.IsJSDocAugmentsTag(tag) {
		return false
	}
	tagName := tag.TagName()
	className := tag.ClassName()
	if tagName == nil || className == nil {
		return false
	}
	between := sourceSlice(sf, tagName.End(), className.Pos())
	return !strings.Contains(between, "{")
}

func isExplicitReactComponentType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindExpressionWithTypeArguments {
		expressionWithTypeArguments := node.AsExpressionWithTypeArguments()
		if expressionWithTypeArguments.TypeArguments != nil && len(expressionWithTypeArguments.TypeArguments.Nodes) != 0 {
			return false
		}
		return isExplicitReactComponentType(expressionWithTypeArguments.Expression)
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	property := node.AsPropertyAccessExpression()
	object := ast.SkipParentheses(property.Expression)
	name := property.Name()
	return object != nil && object.Kind == ast.KindIdentifier && object.AsIdentifier().Text == "React" &&
		name != nil && name.Kind == ast.KindIdentifier &&
		(name.AsIdentifier().Text == "Component" || name.AsIdentifier().Text == "PureComponent")
}

func parameterText(sf *ast.SourceFile, fn *ast.Node) string {
	// NOTE: Unlike ESLint v7.37.5, preserve the complete parameter source.
	// The upstream fixer emits empty names for default, rest, and destructured
	// parameters, which produces invalid JavaScript.
	params := reactutil.FunctionParameters(fn)
	if len(params) == 0 {
		return ""
	}
	first := utils.TrimNodeTextRange(sf, params[0])
	last := utils.TrimNodeTextRange(sf, params[len(params)-1])
	return sourceSlice(sf, first.Pos(), last.End())
}

func typeParameterText(sf *ast.SourceFile, fn *ast.Node) string {
	if len(utils.ESTreeTypeParameters(fn)) == 0 {
		return ""
	}
	list := fn.TypeParameterList()
	if list == nil {
		return ""
	}
	// TypeParameterList's range covers the entries, not the surrounding angle
	// brackets, so restore the syntax consumed by the class-field replacement.
	return "<" + sourceSlice(sf, list.Pos(), list.End()) + ">"
}

func sourceSlice(sf *ast.SourceFile, start, end int) string {
	text := sf.Text()
	if start < 0 || end < start || end > len(text) {
		return ""
	}
	return text[start:end]
}

// buildFixes mirrors eslint-plugin-react's two-part fix. Source text and
// comment/range work stay in this deferred builder so diagnostics-only runs
// do not pay for autofix construction.
func buildFixes(ctx rule.RuleContext, member, fn, rawValue *ast.Node) []rule.RuleFix {
	body := ast.SkipParentheses(reactutil.FunctionBody(fn))
	if body == nil {
		return nil
	}

	name := member.Name()
	if name == nil {
		return nil
	}
	if member.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly|ast.ModifierFlagsAccessor) {
		return nil
	}
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
	bodyStart, bodyEnd := bodyRange.Pos(), bodyRange.End()
	previousToken, hasPreviousToken := utils.TokenBeforePosition(ctx.SourceFile, bodyStart)
	commentsBefore := make([]*ast.CommentRange, 0)
	commentsAfter := make([]*ast.CommentRange, 0)
	if body.Kind != ast.KindBlock {
		for _, comment := range ctx.Comments.All() {
			if comment == nil {
				continue
			}
			if hasPreviousToken && comment.Pos() >= previousToken.End && comment.End() <= bodyStart {
				commentsBefore = append(commentsBefore, comment)
			} else if nextToken, ok := utils.TokenAtOrAfter(ctx.SourceFile, bodyEnd); ok && comment.Pos() >= bodyEnd && comment.End() <= nextToken.Start {
				commentsAfter = append(commentsAfter, comment)
			}
		}
	}

	headEnd := bodyStart
	if len(commentsBefore) != 0 {
		headEnd = commentsBefore[0].Pos()
		bodyStart = commentsBefore[0].Pos()
	}
	if len(commentsAfter) != 0 {
		bodyEnd = commentsAfter[len(commentsAfter)-1].End()
	}
	if body.Kind == ast.KindObjectLiteralExpression {
		// Arrow functions returning object literals have a wrapping `)` outside
		// the ObjectLiteralExpression node: `() => ({})`.
		bodyEnd++
	}
	// ESTree removes parentheses around the arrow function itself. tsgo keeps
	// them as ParenthesizedExpression nodes, so consume their closing suffix
	// together with an expression body when converting the member.
	if rawValue != nil {
		rawValueEnd := utils.TrimNodeTextRange(ctx.SourceFile, rawValue).End()
		if rawValueEnd > bodyEnd {
			bodyEnd = rawValueEnd
		}
	}

	params := parameterText(ctx.SourceFile, fn)
	typeParams := typeParameterText(ctx.SourceFile, fn)
	methodText := "(" + params + ") "
	if member.Kind == ast.KindPropertyAssignment {
		methodText = ": "
		if ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync) {
			methodText += "async "
		}
		methodText += "function" + typeParams + "(" + params + ") "
	} else {
		methodText = typeParams + methodText
	}
	// NOTE: Unlike ESLint v7.37.5, replace after the complete computed key.
	// The upstream range can remove the closing bracket and produce invalid
	// JavaScript for keys such as `[render]`.
	fixes := make([]rule.RuleFix, 0, 2)
	if member.Kind != ast.KindPropertyAssignment && ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync) {
		fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, name, "async "))
	}
	fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(name.End(), headEnd), methodText))
	if body.Kind == ast.KindBlock {
		return fixes
	}

	hasSemi := sourceSlice(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, rawValue).End(), member.End()) == ";"
	if hasSemi {
		bodyEnd++
	}
	var beforeText strings.Builder
	for _, comment := range commentsBefore {
		beforeText.WriteString(sourceSlice(ctx.SourceFile, comment.Pos(), comment.End()))
	}
	var afterText strings.Builder
	for _, comment := range commentsAfter {
		afterText.WriteString(sourceSlice(ctx.SourceFile, comment.Pos(), comment.End()))
	}
	replacement := "{ return " + beforeText.String() + sourceSlice(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, body).Pos(), utils.TrimNodeTextRange(ctx.SourceFile, body).End()) + afterText.String() + "; }"
	fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(bodyStart, bodyEnd), replacement))
	return fixes
}

func reportComponent(ctx rule.RuleContext, component *ast.Node) {
	for _, member := range componentProperties(component) {
		name := memberName(member)
		value := memberValue(member)
		if name == "" || value == nil {
			continue
		}
		value = ast.SkipParentheses(value)
		if value == nil || value.Kind != ast.KindArrowFunction || !isLifecycleMethod(member, name) {
			continue
		}
		ctx.ReportNodeWithDeferredFixes(member, rule.RuleMessage{
			Id:          "lifecycle",
			Description: strings.Replace(lifecycleMessage, "{{propertyName}}", name, 1),
			Data:        map[string]string{"propertyName": name},
		}, func() []rule.RuleFix {
			return buildFixes(ctx, member, value, memberValue(member))
		})
	}
}

// NoArrowFunctionLifecycleRule reports arrow-valued lifecycle members in
// React class components and createReactClass object components.
var NoArrowFunctionLifecycleRule = rule.Rule{
	Name:   "react/no-arrow-function-lifecycle",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(node, pragma) || hasExplicitReactComponentJSDoc(ctx.SourceFile, ctx.Comments.All(), node) {
					reportComponent(ctx, node)
				}
			case ast.KindObjectLiteralExpression:
				if isCreateReactClassObjectComponent(node, pragma, createClass) {
					reportComponent(ctx, node)
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(ctx.SourceFile.AsNode())
		return rule.RuleListeners{}
	},
}
