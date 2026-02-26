package prefer_regexp_exec

import (
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferRegExpExecMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "regExpExecOverStringMatch",
		Description: "Use the `RegExp#exec()` method instead.",
	}
}

func isGlobalRegexLiteral(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindRegularExpressionLiteral {
		return false
	}
	regex := node.AsRegularExpressionLiteral()
	if regex == nil {
		return false
	}
	text := regex.Text
	lastSlash := strings.LastIndex(text, "/")
	if lastSlash < 0 || lastSlash+1 >= len(text) {
		return false
	}
	flags := text[lastSlash+1:]
	return strings.Contains(flags, "g")
}

type staticArgInfo struct {
	known  bool
	global bool
}

const (
	argumentTypeOther  = 0
	argumentTypeString = 1 << iota
	argumentTypeRegExp
)

func unwrapExpression(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	for node != nil {
		switch node.Kind {
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
			node = ast.SkipParentheses(node.Expression())
		default:
			return node
		}
	}
	return nil
}

func isNodeParenthesized(node *ast.Node) bool {
	if node == nil || node.Parent == nil || !ast.IsParenthesizedExpression(node.Parent) {
		return false
	}
	parent := node.Parent.AsParenthesizedExpression()
	return parent != nil && parent.Expression == node
}

func isWeakPrecedenceParent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPostfixUnaryExpression,
		ast.KindPrefixUnaryExpression,
		ast.KindBinaryExpression,
		ast.KindConditionalExpression,
		ast.KindAwaitExpression:
		return true
	}
	if ast.IsPropertyAccessExpression(parent) {
		return parent.AsPropertyAccessExpression().Expression == node
	}
	if ast.IsElementAccessExpression(parent) {
		return parent.AsElementAccessExpression().Expression == node
	}
	if ast.IsCallExpression(parent) || ast.IsNewExpression(parent) {
		return parent.Expression() == node
	}
	if ast.IsTaggedTemplateExpression(parent) {
		return parent.AsTaggedTemplateExpression().Tag == node
	}
	return false
}

func getWrappedNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	text := strings.TrimSpace(scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, node, false))
	if text == "" {
		return ""
	}
	if !utils.IsStrongPrecedenceNode(node) {
		text = "(" + text + ")"
	}
	return text
}

func collectArgumentTypes(ctx rule.RuleContext, argument *ast.Node) int {
	argument = unwrapExpression(argument)
	if argument == nil {
		return argumentTypeOther
	}
	switch argument.Kind {
	case ast.KindStringLiteral:
		return argumentTypeString
	case ast.KindRegularExpressionLiteral:
		return argumentTypeRegExp
	}
	if ctx.TypeChecker == nil {
		return argumentTypeOther
	}
	argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
	result := argumentTypeOther
	for _, part := range utils.UnionTypeParts(argType) {
		switch utils.GetTypeName(ctx.TypeChecker, part) {
		case "RegExp":
			result |= argumentTypeRegExp
		case "string":
			result |= argumentTypeString
		default:
			return argumentTypeOther
		}
	}
	return result
}

func regExpFlagInfo(ctx rule.RuleContext, args []*ast.Node) (known bool, global bool) {
	patternKnown := len(args) == 0 || args[0] == nil
	patternGlobal := false

	// Pattern validity/flags only matter when statically known.
	if len(args) > 0 && args[0] != nil {
		patternArg := unwrapExpression(args[0])
		switch patternArg.Kind {
		case ast.KindStringLiteral:
			if _, err := regexp2.Compile(patternArg.AsStringLiteral().Text, regexp2.ECMAScript); err != nil {
				return false, false
			}
			patternKnown = true
		case ast.KindRegularExpressionLiteral:
			patternKnown = true
			patternGlobal = isGlobalRegexLiteral(patternArg)
		default:
			// For dynamic patterns, only mark as known when type info proves string-only.
			if collectArgumentTypes(ctx, patternArg) == argumentTypeString {
				patternKnown = true
			}
		}
	}

	if len(args) < 2 || args[1] == nil || isUndefinedLiteral(ctx, args[1]) {
		if !patternKnown {
			return false, false
		}
		return true, patternGlobal
	}
	flagsArg := ast.SkipParentheses(args[1])
	if flagsArg.Kind != ast.KindStringLiteral {
		return false, false
	}
	return true, strings.Contains(flagsArg.AsStringLiteral().Text, "g")
}

func isUndefinedLiteral(ctx rule.RuleContext, node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if node.Kind != ast.KindIdentifier {
		return false
	}
	id := node.AsIdentifier()
	if id == nil || id.Text != "undefined" {
		return false
	}
	if ctx.TypeChecker == nil || ctx.Program == nil {
		return true
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(node)
	if sym == nil || sym.Declarations == nil || len(sym.Declarations) == 0 {
		return true
	}
	for _, decl := range sym.Declarations {
		if decl == nil {
			return false
		}
		sourceFile := ast.GetSourceFileOfNode(decl)
		if sourceFile == nil || !sourceFile.IsDeclarationFile {
			return false
		}
	}
	return true
}

func resolveStaticArgumentInfo(ctx rule.RuleContext, node *ast.Node, seen map[*ast.Symbol]bool) staticArgInfo {
	node = unwrapExpression(node)
	if node == nil {
		return staticArgInfo{}
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return staticArgInfo{known: true}
	case ast.KindRegularExpressionLiteral:
		return staticArgInfo{known: true, global: isGlobalRegexLiteral(node)}
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call != nil && call.Expression != nil && call.Expression.Kind == ast.KindIdentifier && call.Expression.AsIdentifier().Text == "RegExp" && call.Arguments != nil {
			known, global := regExpFlagInfo(ctx, call.Arguments.Nodes)
			return staticArgInfo{known: known, global: global}
		}
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		if newExpr != nil && newExpr.Expression != nil && newExpr.Expression.Kind == ast.KindIdentifier && newExpr.Expression.AsIdentifier().Text == "RegExp" && newExpr.Arguments != nil {
			known, global := regExpFlagInfo(ctx, newExpr.Arguments.Nodes)
			return staticArgInfo{known: known, global: global}
		}
	case ast.KindIdentifier:
		if ctx.TypeChecker == nil {
			return staticArgInfo{}
		}
		sym := ctx.TypeChecker.GetSymbolAtLocation(node)
		if sym == nil {
			return staticArgInfo{}
		}
		if seen[sym] {
			return staticArgInfo{}
		}
		seen[sym] = true
		defer delete(seen, sym)
		if sym.Declarations == nil {
			return staticArgInfo{}
		}
		for _, decl := range sym.Declarations {
			if decl == nil || decl.Kind != ast.KindVariableDeclaration {
				continue
			}
			if !ast.IsVariableDeclarationList(decl.Parent) || decl.Parent.Flags&ast.NodeFlagsConst == 0 {
				return staticArgInfo{}
			}
			varDecl := decl.AsVariableDeclaration()
			if varDecl == nil || varDecl.Initializer == nil {
				continue
			}
			info := resolveStaticArgumentInfo(ctx, varDecl.Initializer, seen)
			if info.known {
				return info
			}
		}
	}
	return staticArgInfo{}
}

func definitelyDoesNotContainGlobalFlag(ctx rule.RuleContext, node *ast.Node) bool {
	node = unwrapExpression(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier || call.Expression.AsIdentifier().Text != "RegExp" || call.Arguments == nil {
			return false
		}
		known, global := regExpFlagInfo(ctx, call.Arguments.Nodes)
		return known && !global
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		if newExpr == nil || newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier || newExpr.Expression.AsIdentifier().Text != "RegExp" || newExpr.Arguments == nil {
			return false
		}
		known, global := regExpFlagInfo(ctx, newExpr.Arguments.Nodes)
		return known && !global
	default:
		return false
	}
}

func isStringMatchCall(node *ast.Node) (*ast.CallExpression, bool) {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil, false
	}
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return nil, false
	}

	switch call.Expression.Kind {
	case ast.KindPropertyAccessExpression:
		access := call.Expression.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Text() != "match" {
			return nil, false
		}
		return call, true
	case ast.KindElementAccessExpression:
		access := call.Expression.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil || access.ArgumentExpression.Kind != ast.KindStringLiteral {
			return nil, false
		}
		if access.ArgumentExpression.AsStringLiteral().Text != "match" {
			return nil, false
		}
		return call, true
	}
	return nil, false
}

func isStringLikeReceiver(ctx rule.RuleContext, receiver *ast.Node) bool {
	if receiver == nil || ctx.TypeChecker == nil {
		return false
	}
	receiverType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)
	return utils.GetTypeName(ctx.TypeChecker, receiverType) == "string"
}

func buildRegexLiteralFromString(pattern string) (string, bool) {
	// Validate using ECMAScript semantics (not Go regexp/RE2).
	if _, err := regexp2.Compile(pattern, regexp2.ECMAScript); err != nil {
		return "", false
	}
	var b strings.Builder
	b.WriteByte('/')
	for _, ch := range pattern {
		switch ch {
		case '/':
			b.WriteString(`\/`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(ch)
		}
	}
	b.WriteByte('/')
	return b.String(), true
}

func buildPreferRegExpExecReplacement(ctx rule.RuleContext, callNode *ast.Node, receiver *ast.Node, arg *ast.Node, argumentTypes int) (string, bool) {
	if ctx.SourceFile == nil || callNode == nil || receiver == nil || arg == nil {
		return "", false
	}
	receiverText := getWrappedNodeText(ctx.SourceFile, receiver)
	argText := getWrappedNodeText(ctx.SourceFile, arg)
	if receiverText == "" || argText == "" {
		return "", false
	}

	var replacement string
	if arg.Kind == ast.KindStringLiteral {
		regexLiteral, ok := buildRegexLiteralFromString(arg.AsStringLiteral().Text)
		if !ok {
			return "", false
		}
		replacement = regexLiteral + ".exec(" + receiverText + ")"
	} else {
		switch argumentTypes {
		case argumentTypeRegExp:
			replacement = argText + ".exec(" + receiverText + ")"
		case argumentTypeString:
			replacement = "RegExp(" + argText + ").exec(" + receiverText + ")"
		default:
			return "", false
		}
	}
	if isWeakPrecedenceParent(callNode) && !isNodeParenthesized(callNode) {
		replacement = "(" + replacement + ")"
	}
	return replacement, true
}

var PreferRegExpExecRule = rule.CreateRule(rule.Rule{
	Name: "prefer-regexp-exec",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := isStringMatchCall(node)
				if !ok || call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
					return
				}

				var receiver *ast.Node
				var reportNode *ast.Node
				switch call.Expression.Kind {
				case ast.KindPropertyAccessExpression:
					access := call.Expression.AsPropertyAccessExpression()
					receiver = access.Expression
					reportNode = access.Name().AsNode()
				case ast.KindElementAccessExpression:
					access := call.Expression.AsElementAccessExpression()
					receiver = access.Expression
					reportNode = access.ArgumentExpression
				}
				if receiver == nil || reportNode == nil {
					return
				}
				if !isStringLikeReceiver(ctx, receiver) {
					return
				}

				arg := call.Arguments.Nodes[0]
				argumentTypes := collectArgumentTypes(ctx, arg)
				if argumentTypes == argumentTypeOther || argumentTypes == argumentTypeString|argumentTypeRegExp {
					return
				}
				staticInfo := resolveStaticArgumentInfo(ctx, arg, map[*ast.Symbol]bool{})
				if staticInfo.known && staticInfo.global {
					return
				}
				if !staticInfo.known && argumentTypes&argumentTypeRegExp != 0 && !definitelyDoesNotContainGlobalFlag(ctx, arg) {
					return
				}
				if arg.Kind == ast.KindStringLiteral {
					if _, err := regexp2.Compile(arg.AsStringLiteral().Text, regexp2.ECMAScript); err != nil {
						return
					}
				}

				msg := buildPreferRegExpExecMessage()
				if replacement, ok := buildPreferRegExpExecReplacement(ctx, node, receiver, arg, argumentTypes); ok {
					if ctx.SourceFile == nil {
						ctx.ReportNode(reportNode, msg)
						return
					}
					ctx.ReportNodeWithFixes(reportNode, msg, rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, node), replacement))
					return
				}
				ctx.ReportNode(reportNode, msg)
			},
		}
	},
})
