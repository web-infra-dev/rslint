// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package no_typeof_undefined

import (
	_ "embed"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_typeof_undefined.schema.json
var schemaJSON []byte

const (
	messageIDError      = "no-typeof-undefined/error"
	messageIDSuggestion = "no-typeof-undefined/suggestion"
)

var errorMessage = rule.RuleMessage{
	Id:          messageIDError,
	Description: "Compare with `undefined` directly instead of using `typeof`.",
}

func suggestionMessage(operator ast.Kind) rule.RuleMessage {
	strict := "==="
	if operator == ast.KindExclamationEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
		strict = "!=="
	}
	return rule.RuleMessage{
		Id:          messageIDSuggestion,
		Description: "Switch to `… " + strict + " undefined`.",
		Data:        map[string]string{"operator": strict},
	}
}

var NoTypeofUndefinedRule = rule.Rule{
	Name:   "unicorn/no-typeof-undefined",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		checkGlobals := checkGlobalVariables(rawOptions)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil || !isSupportedComparison(binary.OperatorToken.Kind) {
					return
				}

				typeofNode := ast.SkipParentheses(binary.Left)
				if typeofNode == nil || typeofNode.Kind != ast.KindTypeOfExpression {
					return
				}
				undefinedString := ast.SkipParentheses(binary.Right)
				if undefinedString == nil || undefinedString.Kind != ast.KindStringLiteral ||
					undefinedString.AsStringLiteral().Text != "undefined" {
					return
				}

				valueNode := typeofNode.AsTypeOfExpression().Expression
				runtimeValue := ast.SkipOuterExpressions(valueNode, ast.OEKParentheses|ast.OEKAssertions)
				isGlobal := runtimeValue != nil && ast.IsIdentifier(runtimeValue) &&
					ctx.Refs != nil && ctx.Refs.IsGlobalReference(runtimeValue)
				if isGlobal && !checkGlobals {
					return
				}

				safeReplacement := ctx.Globals.Access("undefined").IsDeclared() &&
					!utils.IsShadowed(node, "undefined") &&
					!isGlobalDocumentAll(ctx, runtimeValue)

				// TypeOfExpression always owns the `typeof` token followed by its operand.
				tokens := utils.TokensOfNode(ctx.SourceFile, typeofNode)
				typeofToken := tokens[0]
				fixes := func() []rule.RuleFix {
					if !safeReplacement {
						return nil
					}
					return buildFixes(ctx.SourceFile, node, typeofNode, undefinedString, binary.OperatorToken, typeofToken, tokens[1])
				}

				if isGlobal {
					ctx.ReportRangeWithDeferredSuggestions(typeofToken.Range(), errorMessage, func() []rule.RuleSuggestion {
						if !safeReplacement {
							return nil
						}
						return []rule.RuleSuggestion{{
							Message:  suggestionMessage(binary.OperatorToken.Kind),
							FixesArr: fixes(),
						}}
					})
					return
				}

				ctx.ReportRangeWithDeferredFixes(typeofToken.Range(), errorMessage, fixes)
			},
		}
	},
}

func isGlobalDocumentAll(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || !utils.IsSpecificMemberAccess(node, "", "all") {
		return false
	}
	object := utils.AccessExpressionObject(node)
	object = ast.SkipOuterExpressions(object, ast.OEKParentheses|ast.OEKAssertions)
	return object != nil && ast.IsIdentifier(object) && object.Text() == "document" &&
		ctx.Refs != nil && ctx.Refs.IsGlobalReference(object)
}

func checkGlobalVariables(rawOptions []any) bool {
	if len(rawOptions) == 0 {
		return false
	}
	options, _ := rawOptions[0].(map[string]any)
	value, _ := options["checkGlobalVariables"].(bool)
	return value
}

func isSupportedComparison(operator ast.Kind) bool {
	switch operator {
	case ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsEqualsToken,
		ast.KindEqualsEqualsToken,
		ast.KindExclamationEqualsToken:
		return true
	default:
		return false
	}
}

func buildFixes(
	sourceFile *ast.SourceFile,
	binaryNode *ast.Node,
	typeofNode *ast.Node,
	undefinedString *ast.Node,
	operatorToken *ast.Node,
	typeofToken utils.SourceToken,
	secondToken utils.SourceToken,
) []rule.RuleFix {
	fixes := []rule.RuleFix{
		rule.RuleFixReplace(sourceFile, undefinedString, "undefined"),
		rule.RuleFixRemoveRange(typeofToken.Range()),
	}

	switch operatorToken.Kind {
	case ast.KindEqualsEqualsToken:
		fixes = append(fixes, rule.RuleFixReplace(sourceFile, operatorToken, "==="))
	case ast.KindExclamationEqualsToken:
		fixes = append(fixes, rule.RuleFixReplace(sourceFile, operatorToken, "!=="))
	}

	if whitespace := whitespaceAfterToken(sourceFile.Text(), typeofToken.End); whitespace.End() > whitespace.Pos() {
		fixes = append(fixes, rule.RuleFixRemoveRange(whitespace))
	}

	if needsReturnOrThrowParentheses(sourceFile, binaryNode, typeofNode, typeofToken, secondToken) {
		fixes = append(fixes, returnOrThrowParenthesesFixes(sourceFile, binaryNode.Parent)...)
	} else if operandNeedsExpressionStatementParentheses(binaryNode, typeofNode) {
		operand := typeofNode.AsTypeOfExpression().Expression
		r := utils.TrimNodeTextRange(sourceFile, operand)
		fixes = append(fixes,
			rule.RuleFixReplaceRange(core.NewTextRange(r.Pos(), r.Pos()), "("),
			rule.RuleFixReplaceRange(core.NewTextRange(r.End(), r.End()), ")"),
		)
	} else if unicornutil.NeedsSemicolonBefore(sourceFile, binaryNode, secondToken.Text) {
		start := utils.TrimNodeTextRange(sourceFile, binaryNode).Pos()
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, start), ";"))
	}

	return fixes
}

func operandNeedsExpressionStatementParentheses(binaryNode, typeofNode *ast.Node) bool {
	if binaryNode == nil || binaryNode.Parent == nil || !ast.IsExpressionStatement(binaryNode.Parent) ||
		typeofNode == nil || typeofNode.Kind != ast.KindTypeOfExpression {
		return false
	}
	operand := typeofNode.AsTypeOfExpression().Expression
	operand = ast.SkipOuterExpressions(operand, ast.OEKAssertions)
	switch operand.Kind {
	case ast.KindObjectLiteralExpression, ast.KindFunctionExpression, ast.KindClassExpression:
		return true
	default:
		return false
	}
}

func whitespaceAfterToken(source string, start int) core.TextRange {
	end := start
	for end < len(source) {
		r, size := utf8.DecodeRuneInString(source[end:])
		if (r == utf8.RuneError && size == 0) || !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
			break
		}
		end += size
	}
	return core.NewTextRange(start, end)
}

func isParenthesized(node *ast.Node) bool {
	return node != nil && node.Parent != nil &&
		node.Parent.Kind == ast.KindParenthesizedExpression &&
		node.Parent.AsParenthesizedExpression().Expression == node
}

func needsReturnOrThrowParentheses(
	sourceFile *ast.SourceFile,
	binaryNode *ast.Node,
	typeofNode *ast.Node,
	typeofToken utils.SourceToken,
	secondToken utils.SourceToken,
) bool {
	if binaryNode == nil || binaryNode.Parent == nil ||
		isParenthesized(binaryNode) || isParenthesized(typeofNode) ||
		utils.IsSameLine(sourceFile, typeofToken.Start, secondToken.Start) {
		return false
	}

	parent := binaryNode.Parent
	switch parent.Kind {
	case ast.KindReturnStatement:
		return parent.AsReturnStatement().Expression == binaryNode
	case ast.KindThrowStatement:
		return parent.AsThrowStatement().Expression == binaryNode
	default:
		return false
	}
}

func returnOrThrowParenthesesFixes(sourceFile *ast.SourceFile, statement *ast.Node) []rule.RuleFix {
	tokens := utils.TokensOfNode(sourceFile, statement)
	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(tokens[0].End, tokens[0].End), " ("),
	}
	last := tokens[len(tokens)-1]
	end := utils.TrimNodeTextRange(sourceFile, statement).End()
	if last.Kind == ast.KindSemicolonToken {
		end = last.Start
	}
	return append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(end, end), ")"))
}
