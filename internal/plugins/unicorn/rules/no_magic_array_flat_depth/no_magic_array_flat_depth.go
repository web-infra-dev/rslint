package no_magic_array_flat_depth

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-magic-array-flat-depth"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Magic number as depth is not allowed.",
}

// NoMagicArrayFlatDepthRule flags `Array#flat(<numeric literal>)` calls where
// the depth argument is any numeric literal other than `1` (the default). A
// "magic" depth such as `flat(2)` or `flat(99)` is almost always a sign that
// the caller doesn't know the array's structure; preferring `flat(Infinity)`
// or restructuring the data is usually clearer.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/main/docs/rules/no-magic-array-flat-depth.md
var NoMagicArrayFlatDepthRule = rule.Rule{
	Name:   "unicorn/no-magic-array-flat-depth",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "flat",
					ArgumentsLength:     &[]int{1}[0],
					AllowOptionalMember: true,
				})
				if !ok {
					return
				}

				depth := call.Call.Arguments()[0]
				// ESTree / upstream transparently unwraps parens around the
				// argument; tsgo keeps ParenthesizedExpression nodes, so the
				// depth may be wrapped. SkipParentheses matches the upstream
				// semantic.
				depth = ast.SkipParentheses(depth)
				if !isNumericLiteral(depth) {
					return
				}
				// `depth.value === 1` in upstream. tsgo's NumericLiteral exposes
				// the raw text rather than a typed value; NormalizeNumericLiteral
				// decimal-canonicalises the text so `1`, `1.0`, and `0x01` all
				// round-trip to the string `"1"` and reject this branch.
				if utils.NormalizeNumericLiteral(depth.AsNumericLiteral().Text) == "1" {
					return
				}

				// Skip if any comment sits between the callee end (= the `(`) and
				// the closing paren — a commented explanation is exactly the kind
				// of context upstream wants to preserve. Upstream checks
				// `commentsExistBetween(openingParenthesisToken, closingParenthesisToken)`,
				// which covers both sides of the depth argument.
				if hasCommentBetweenParens(ctx, call.Call, depth) {
					return
				}

				// Skip when the receiver is statically known NOT to be an array
				// (e.g. a typed `Set`). Mirrors upstream's
				// `shouldSkipKnownNonArrayReceiver`: typed arrays stay
				// reportable, but receivers like `new Set()` are skipped, and
				// directly-visible mismatches like `"x".flat(2)` or
				// `({flat(){}}).flat(2)` are always reported.
				if unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
					return
				}

				ctx.ReportNode(depth, message)
			},
		}
	},
}

// isNumericLiteral reports whether node is a numeric literal. Mirrors
// upstream's `isNumericLiteral` helper.
func isNumericLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindNumericLiteral
}

// hasCommentBetweenParens reports whether any comment lives in the source
// span from the end of the callee (right after `flat`, i.e. the position of
// the opening `(`) up to the end of the call expression. Mirrors upstream's
// `sourceCode.commentsExistBetween(openingParenthesisToken, closingParenthesisToken)`.
func hasCommentBetweenParens(ctx rule.RuleContext, call *ast.Node, depth *ast.Node) bool {
	callee := call.AsCallExpression().Expression
	if callee == nil {
		return false
	}
	openParenPos := utils.TrimNodeTextRange(ctx.SourceFile, callee).End()
	callEnd := utils.TrimNodeTextRange(ctx.SourceFile, call).End()
	depthStart := utils.TrimNodeTextRange(ctx.SourceFile, depth).Pos()
	if openParenPos >= callEnd {
		return false
	}
	// Avoid wasting a comment scan on the trivial case where the depth is the
	// only thing in the parens.
	if depthStart == openParenPos+1 || depthStart == openParenPos+2 {
		// depth is immediately after `(`, optionally with whitespace.
		depthEnd := depth.End()
		return utils.HasCommentInSpan(ctx.Comments.All(), depthEnd, callEnd)
	}
	return utils.HasCommentInSpan(ctx.Comments.All(), openParenPos, callEnd)
}
