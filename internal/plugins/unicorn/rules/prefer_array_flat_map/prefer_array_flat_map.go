package prefer_array_flat_map

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "prefer-array-flat-map"

var preferArrayFlatMapMessage = rule.RuleMessage{
	Id:          messageID,
	Description: "Prefer `.flatMap(…)` over `.map(…).flat()`.",
}

type dotMethodCall struct {
	call      *ast.Node
	rawCallee *ast.Node
	callee    *ast.Node
	object    *ast.Node
	property  *ast.Node
}

// matchDotMethodCall mirrors unicorn's isMethodCall with computed:false. The
// wider utils.IsSpecificMemberAccess helper also matches static bracket access,
// which would make foo["map"](...).flat() a false positive for this rule.
func matchDotMethodCall(node *ast.Node, method string, optionalCallFalse bool, optionalMemberFalse bool) (dotMethodCall, bool) {
	if node == nil || !ast.IsCallExpression(node) {
		return dotMethodCall{}, false
	}

	call := node.AsCallExpression()
	// Check the direct optional token, not ast.IsOptionalChain: in tsgo,
	// foo?.map(...).flat() marks later links as optional-chain nodes too, but
	// upstream still reports because only map's member access is optional.
	if optionalCallFalse && call.QuestionDotToken != nil {
		return dotMethodCall{}, false
	}

	rawCallee := call.Expression
	callee := ast.SkipParentheses(rawCallee)
	// Parentheses around an optional-chain callee are not transparent to
	// upstream isMethodCall: (foo?.map)(cb) is not treated as foo?.map(cb).
	if callee == nil || (rawCallee != callee && ast.IsOptionalChain(callee)) || !ast.IsPropertyAccessExpression(callee) {
		return dotMethodCall{}, false
	}

	propAccess := callee.AsPropertyAccessExpression()
	if optionalMemberFalse && propAccess.QuestionDotToken != nil {
		return dotMethodCall{}, false
	}

	property := propAccess.Name()
	if property == nil || !ast.IsIdentifier(property) || property.AsIdentifier().Text != method {
		return dotMethodCall{}, false
	}

	return dotMethodCall{
		call:      node,
		rawCallee: rawCallee,
		callee:    callee,
		object:    propAccess.Expression,
		property:  property,
	}, true
}

func hasNoDepthOrRawDepthOne(sf *ast.SourceFile, call *ast.Node) bool {
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) == 0 {
		return true
	}
	if len(args.Nodes) != 1 {
		return false
	}

	arg := ast.SkipParentheses(args.Nodes[0])
	if arg == nil || arg.Kind != ast.KindNumericLiteral {
		return false
	}
	return scanner.GetSourceTextOfNodeFromSourceFile(sf, arg, false) == "1"
}

func isIgnoredMapObject(node *ast.Node) bool {
	return nodeMatchesPath(node, "Children") || nodeMatchesPath(node, "React.Children")
}

// nodeMatchesPath mirrors unicorn's isNodeMatches for the React.Children
// exception: parentheses are transparent, while computed and optional members
// are not. That keeps React["Children"].map(...).flat() reportable.
func nodeMatchesPath(node *ast.Node, path string) bool {
	parts := strings.Split(path, ".")
	return nodeMatchesPathParts(ast.SkipParentheses(node), parts)
}

func nodeMatchesPathParts(node *ast.Node, parts []string) bool {
	if node == nil || len(parts) == 0 {
		return false
	}
	if len(parts) == 1 {
		return ast.IsIdentifier(node) && node.AsIdentifier().Text == parts[0]
	}
	if !ast.IsPropertyAccessExpression(node) {
		return false
	}

	propAccess := node.AsPropertyAccessExpression()
	if propAccess.QuestionDotToken != nil {
		return false
	}
	name := propAccess.Name()
	if name == nil || !ast.IsIdentifier(name) || name.AsIdentifier().Text != parts[len(parts)-1] {
		return false
	}
	return nodeMatchesPathParts(ast.SkipParentheses(propAccess.Expression), parts[:len(parts)-1])
}

func buildFixes(sf *ast.SourceFile, flatCall dotMethodCall, mapCall dotMethodCall) []rule.RuleFix {
	mapPropertyRange := utils.TrimNodeTextRange(sf, mapCall.property)
	// Remove the .flat member and the following call separately, preserving any
	// parentheses around the callee: (foo.map(cb).flat)() -> (foo.flatMap(cb)).
	removeFlatPropertyRange := core.NewTextRange(flatCall.object.End(), flatCall.callee.End())
	removeFlatArgumentsRange := core.NewTextRange(flatCall.rawCallee.End(), flatCall.call.End())
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(mapPropertyRange, "flatMap"),
		rule.RuleFixRemoveRange(removeFlatPropertyRange),
		rule.RuleFixRemoveRange(removeFlatArgumentsRange),
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-array-flat-map.md
var PreferArrayFlatMapRule = rule.Rule{
	Name: "unicorn/prefer-array-flat-map",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				flatCall, ok := matchDotMethodCall(node, "flat", true, true)
				if !ok || !hasNoDepthOrRawDepthOne(ctx.SourceFile, node) {
					return
				}

				mapCallObject := ast.SkipParentheses(flatCall.object)
				mapCall, ok := matchDotMethodCall(mapCallObject, "map", true, false)
				if !ok || isIgnoredMapObject(mapCall.object) {
					return
				}

				reportRange := utils.TrimNodeTextRange(ctx.SourceFile, mapCall.property).WithEnd(node.End())
				ctx.ReportRangeWithFixes(
					reportRange,
					preferArrayFlatMapMessage,
					buildFixes(ctx.SourceFile, flatCall, mapCall)...,
				)
			},
		}
	},
}
