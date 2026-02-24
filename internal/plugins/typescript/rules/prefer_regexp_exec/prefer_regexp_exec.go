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

func regExpFlagInfo(args []*ast.Node) (known bool, global bool) {
	// Pattern validity only matters when it is statically known.
	// If the pattern is dynamic, we can still reason about the global flag from the flags argument.
	if len(args) > 0 && args[0] != nil {
		patternArg := ast.SkipParentheses(args[0])
		if patternArg.Kind == ast.KindStringLiteral {
			if _, err := regexp2.Compile(patternArg.AsStringLiteral().Text, regexp2.ECMAScript); err != nil {
				return false, false
			}
		}
	}

	if len(args) < 2 || args[1] == nil || isUndefinedLiteral(args[1]) {
		return true, false
	}
	flagsArg := ast.SkipParentheses(args[1])
	if flagsArg.Kind != ast.KindStringLiteral {
		return false, false
	}
	return true, strings.Contains(flagsArg.AsStringLiteral().Text, "g")
}

func isUndefinedLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if node.Kind != ast.KindIdentifier {
		return false
	}
	id := node.AsIdentifier()
	return id != nil && id.Text == "undefined"
}

func resolveStaticArgumentInfo(ctx rule.RuleContext, node *ast.Node, seen map[*ast.Symbol]bool) staticArgInfo {
	node = ast.SkipParentheses(node)
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
			known, global := regExpFlagInfo(call.Arguments.Nodes)
			return staticArgInfo{known: known, global: global}
		}
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		if newExpr != nil && newExpr.Expression != nil && newExpr.Expression.Kind == ast.KindIdentifier && newExpr.Expression.AsIdentifier().Text == "RegExp" && newExpr.Arguments != nil {
			known, global := regExpFlagInfo(newExpr.Arguments.Nodes)
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

func definitelyDoesNotContainGlobalFlag(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindIdentifier || call.Expression.AsIdentifier().Text != "RegExp" || call.Arguments == nil {
			return false
		}
		known, global := regExpFlagInfo(call.Arguments.Nodes)
		return known && !global
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		if newExpr == nil || newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier || newExpr.Expression.AsIdentifier().Text != "RegExp" || newExpr.Arguments == nil {
			return false
		}
		known, global := regExpFlagInfo(newExpr.Arguments.Nodes)
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

func isRegExpOrStringArgument(ctx rule.RuleContext, argument *ast.Node) bool {
	if argument == nil {
		return false
	}
	if argument.Kind == ast.KindRegularExpressionLiteral || argument.Kind == ast.KindStringLiteral {
		return true
	}
	if ctx.TypeChecker == nil {
		return false
	}
	argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
	typeName := utils.GetTypeName(ctx.TypeChecker, argType)
	if typeName == "RegExp" || typeName == "string" {
		return true
	}
	if utils.IsUnionType(argType) {
		regExpSeen := false
		stringSeen := false
		for _, part := range utils.UnionTypeParts(argType) {
			partName := utils.GetTypeName(ctx.TypeChecker, part)
			if partName == "RegExp" {
				regExpSeen = true
			} else if partName == "string" {
				stringSeen = true
			} else {
				return false
			}
		}
		// If both string and RegExp are possible, avoid reporting.
		if regExpSeen && stringSeen {
			return false
		}
		return regExpSeen || stringSeen
	}
	return false
}

func buildRegexLiteralFromString(pattern string) (string, bool) {
	// Validate using ECMAScript semantics (not Go regexp/RE2).
	if _, err := regexp2.Compile(pattern, regexp2.ECMAScript); err != nil {
		return "", false
	}
	pattern = strings.ReplaceAll(pattern, `\`, `\\`)
	pattern = strings.ReplaceAll(pattern, "\n", `\n`)
	pattern = strings.ReplaceAll(pattern, "\r", `\r`)
	pattern = strings.ReplaceAll(pattern, "/", `\/`)
	return "/" + pattern + "/", true
}

func buildPreferRegExpExecReplacement(ctx rule.RuleContext, receiver *ast.Node, arg *ast.Node) (string, bool) {
	if ctx.SourceFile == nil || receiver == nil || arg == nil {
		return "", false
	}
	receiverText := strings.TrimSpace(scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, receiver, false))
	argText := strings.TrimSpace(scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false))
	if receiverText == "" || argText == "" {
		return "", false
	}

	if arg.Kind == ast.KindStringLiteral {
		regexLiteral, ok := buildRegexLiteralFromString(arg.AsStringLiteral().Text)
		if !ok {
			return "", false
		}
		return regexLiteral + ".exec(" + receiverText + ")", true
	}

	if arg.Kind == ast.KindRegularExpressionLiteral {
		return argText + ".exec(" + receiverText + ")", true
	}

	if ctx.TypeChecker == nil {
		return "", false
	}
	argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arg)
	typeName := utils.GetTypeName(ctx.TypeChecker, argType)
	if typeName == "RegExp" {
		return argText + ".exec(" + receiverText + ")", true
	}
	if typeName == "string" {
		return "RegExp(" + argText + ").exec(" + receiverText + ")", true
	}
	return "", false
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
				staticInfo := resolveStaticArgumentInfo(ctx, arg, map[*ast.Symbol]bool{})
				if staticInfo.known && staticInfo.global {
					return
				}
				if !staticInfo.known && !definitelyDoesNotContainGlobalFlag(arg) {
					return
				}
				if !isRegExpOrStringArgument(ctx, arg) {
					return
				}
				if arg.Kind == ast.KindStringLiteral {
					if _, err := regexp2.Compile(arg.AsStringLiteral().Text, regexp2.ECMAScript); err != nil {
						return
					}
				}

				msg := buildPreferRegExpExecMessage()
				if replacement, ok := buildPreferRegExpExecReplacement(ctx, receiver, arg); ok {
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
