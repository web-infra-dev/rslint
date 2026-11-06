// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package prefer_query_selector

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed prefer_query_selector.schema.json
var schemaJSON []byte

// This repository-authored pattern is equivalent to upstream's ASCII-only
// /^-?[A-Z_a-z][\w-]*$/u, without interpreting user-supplied patterns.
var simpleCSSIdentifier = regexp.MustCompile(`^-?[A-Z_a-z][A-Za-z0-9_-]*$`)

var methods = []string{"getElementById", "getElementsByClassName", "getElementsByTagName", "getElementsByName"}

var PreferQuerySelectorRule = rule.Rule{
	Name:   "unicorn/prefer-query-selector",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowWithVariables := false
		if len(options) > 0 {
			settings, _ := options[0].(map[string]any)
			allowWithVariables, _ = settings["allowWithVariables"].(bool)
		}
		oneArgument := 1
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Methods: methods, ArgumentsLength: &oneArgument, RejectSpreadElement: true,
			})
			if !ok || unicornutil.IsNonDOMNodeExpression(utils.ESTreeRuntimeExpression(call.Object)) {
				return
			}
			method := call.Property.Text()
			argument := ast.SkipOuterExpressions(node.Arguments()[0], ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
			if allowWithVariables && (method == "getElementById" || method == "getElementsByClassName") && isNonLiteralArgument(argument) {
				return
			}

			replacement := "querySelectorAll"
			var firstElement *ast.Node
			if method == "getElementById" {
				replacement = "querySelector"
			} else {
				firstElement = getFirstElementAccess(ctx, node)
				if firstElement != nil {
					replacement = "querySelector"
				}
			}
			ctx.ReportNodeWithDeferredFixes(call.Property, rule.RuleMessage{
				Id: "prefer-query-selector", Description: "Prefer `." + replacement + "()` over `." + method + "()`.",
				Data: map[string]string{"replacement": replacement, "method": method},
			}, func() []rule.RuleFix {
				return selectorFixes(ctx, call, argument, method, replacement, firstElement)
			})
		}}
	},
}

func isNonLiteralArgument(node *ast.Node) bool {
	if node.Kind == ast.KindNullKeyword || ast.IsStringLiteralLike(node) {
		return false
	}
	if ast.IsBinaryExpression(node) {
		op := node.AsBinaryExpression().OperatorToken.Kind
		if op != ast.KindCommaToken && !ast.IsAssignmentOperator(op) && !ast.IsLogicalOrCoalescingBinaryOperator(op) {
			return false
		}
	}
	if node.Kind == ast.KindTemplateExpression {
		template := node.AsTemplateExpression()
		if !ecmascript.IsBlank(template.Head.Text()) {
			return false
		}
		for _, span := range template.TemplateSpans.Nodes {
			if !ecmascript.IsBlank(span.AsTemplateSpan().Literal.Text()) {
				return false
			}
		}
	}
	return true
}

func isDocumentObject(node *ast.Node) bool {
	// Parentheses terminating an optional chain expose a ChainExpression;
	// an optional member inside a continuing chain remains a MemberExpression.
	raw := node
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || (raw != node && ast.IsOptionalChain(node)) {
		return false
	}
	if ast.IsIdentifier(node) {
		return node.Text() == "document"
	}
	if !ast.IsPropertyAccessExpression(node) {
		return false
	}
	member := node.AsPropertyAccessExpression()
	object := utils.ESTreeRuntimeExpression(member.Expression)
	return ast.IsIdentifier(object) && (object.Text() == "globalThis" || object.Text() == "window") &&
		ast.IsIdentifier(member.Name()) && member.Name().Text() == "document"
}

func selectorFixes(ctx rule.RuleContext, call unicornutil.DotMethodCall, argument *ast.Node, method, replacement string, firstElement *ast.Node) []rule.RuleFix {
	value, static := utils.GetStaticStringLiteralValue(argument)
	if argument.Kind != ast.KindNullKeyword {
		if !static || ecmascript.IsBlank(value) {
			return nil
		}
		switch method {
		case "getElementById":
			if !simpleCSSIdentifier.MatchString(value) {
				return nil
			}
		case "getElementsByTagName":
			if value != "*" && !simpleCSSIdentifier.MatchString(value) {
				return nil
			}
		case "getElementsByClassName":
			classes := strings.FieldsFunc(value, ecmascript.IsWhiteSpaceOrLineTerminator)
			for _, class := range classes {
				if !simpleCSSIdentifier.MatchString(class) {
					return nil
				}
			}
			value = "." + strings.Join(classes, ".")
		case "getElementsByName":
			raw := utils.TrimmedNodeText(ctx.SourceFile, argument)
			if strings.ContainsAny(value, "\n\r\"'\\\u2028\u2029") || strings.ContainsAny(raw, "\\\n\r\u2028\u2029") {
				return nil
			}
		}
	}

	fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, replacement)}
	scopeSelector := static && !isDocumentObject(call.Object)
	if static && (scopeSelector || method != "getElementsByTagName") {
		raw := utils.TrimmedNodeText(ctx.SourceFile, argument)
		quote := raw[:1]
		switch method {
		case "getElementById":
			value = "#" + value
		case "getElementsByName":
			innerQuote := "'"
			if quote == "'" {
				innerQuote = "\""
			}
			value = "[name=" + innerQuote + value + innerQuote + "]"
		}
		if scopeSelector {
			value = ":scope " + value
		}
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, argument, quote+value+quote))
	}
	if firstElement != nil {
		if ast.IsElementAccessExpression(firstElement) {
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(firstElement.Expression().End(), firstElement.End())))
		} else if access, ok := unicornutil.MatchDotMethodCall(firstElement, unicornutil.DotMethodCallOptions{}); ok {
			fixes = append(fixes, unicornutil.RemoveMethodCallFixes(access)...)
		}
	}
	return fixes
}

func getFirstElementAccess(ctx rule.RuleContext, call *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(call)
	if parent == nil || !ast.IsAccessExpression(parent) {
		return nil
	}
	object := parent.Expression()
	if utils.ESTreeRuntimeExpression(object) != call || (object != call && ast.IsOptionalChain(call)) {
		return nil
	}
	var access *ast.Node
	if ast.IsElementAccessExpression(parent) && !ast.IsOptionalChainRoot(parent) && isZeroLiteral(parent.AsElementAccessExpression().ArgumentExpression) {
		access = parent
	} else if ast.IsPropertyAccessExpression(parent) {
		oneArgument := 1
		candidate := utils.ESTreeParent(parent)
		if methodCall, ok := unicornutil.MatchDotMethodCall(candidate, unicornutil.DotMethodCallOptions{
			Methods: []string{"at", "item"}, ArgumentsLength: &oneArgument, RejectSpreadElement: true,
		}); ok && methodCall.Callee == parent && isZeroLiteral(candidate.Arguments()[0]) {
			access = candidate
		}
	}
	if access == nil || isReferenceTarget(access) || utils.HasCommentInSpan(ctx.Comments.All(), utils.OutermostParenthesizedExpression(call).End(), access.End()) {
		return nil
	}
	return access
}

func isZeroLiteral(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return ast.IsNumericLiteral(node) && utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) == "0"
}

func isReferenceTarget(node *ast.Node) bool {
	for parent := node.Parent; parent != nil && ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments); parent = node.Parent {
		node = parent
	}
	return node.Parent != nil && (ast.IsAssignmentTarget(node) || node.Parent.Kind == ast.KindDeleteExpression)
}
