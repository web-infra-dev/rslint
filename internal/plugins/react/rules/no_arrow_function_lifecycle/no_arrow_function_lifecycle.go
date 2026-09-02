package no_arrow_function_lifecycle

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	return reactutil.IdentifierOrPrivateName(member.Name())
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

func isLifecycleMethod(member *ast.Node, name string) bool {
	methods := instanceLifecycleMethods
	if utils.IncludesModifier(member, ast.KindStaticKeyword) {
		methods = staticLifecycleMethods
	}
	return methods[name]
}

func parameterNames(fn *ast.Node) string {
	params := reactutil.FunctionParameters(fn)
	names := make([]string, len(params))
	for i, param := range params {
		if param != nil {
			name := param.AsParameterDeclaration().Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				names[i] = name.AsIdentifier().Text
			}
		}
	}
	return strings.Join(names, ", ")
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
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
	bodyStart, bodyEnd := bodyRange.Pos(), bodyRange.End()
	previousToken, hasPreviousToken := utils.TokenBeforePosition(ctx.SourceFile, bodyStart)
	commentsBefore := make([]*ast.CommentRange, 0)
	commentsAfter := make([]*ast.CommentRange, 0)
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

	params := parameterNames(fn)
	methodText := "(" + params + ") "
	if member.Kind == ast.KindPropertyAssignment {
		methodText = ": function(" + params + ") "
	}
	fixes := []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(name.End(), headEnd), methodText)}
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
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(node, pragma) {
					reportComponent(ctx, node)
				}
			case ast.KindObjectLiteralExpression:
				if reactutil.IsCreateReactClassObjectArg(node, pragma, createClass) {
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
