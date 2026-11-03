// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package no_useless_length_check

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var NoUselessLengthCheckRule = rule.Rule{
	Name:   "unicorn/no-useless-length-check",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !unicornutil.IsLogicalExpression(node) {
					return
				}
				operator := node.AsBinaryExpression().OperatorToken.Kind
				if parent := utils.ESTreeParent(node); unicornutil.IsLogicalExpression(parent) && parent.AsBinaryExpression().OperatorToken.Kind == operator {
					return
				}
				var buffer [8]*ast.Node
				conditions := buffer[:0]
				flattenLogical(node, operator, &conditions)
				for index, condition := range conditions {
					member := lengthCheck(ctx.SourceFile, condition)
					if member == nil {
						continue
					}
					comparison := condition.AsBinaryExpression().OperatorToken.Kind
					method, messageID := "", "non-zero"
					if operator == ast.KindBarBarToken && comparison == ast.KindEqualsEqualsEqualsToken {
						method, messageID = "every", "zero"
					} else if operator == ast.KindAmpersandAmpersandToken && isNonZeroOperator(comparison) {
						method = "some"
					} else {
						continue
					}
					var previous, next *ast.Node
					if index > 0 {
						previous = conditions[index-1]
					}
					if index+1 < len(conditions) {
						next = conditions[index+1]
					}
					leading := isMatchingCall(ctx, next, method, member.Expression())
					if !leading && !isMatchingCall(ctx, previous, method, member.Expression()) {
						continue
					}
					// Arguments of a leading call start running for empty arrays;
					// a trailing callback must not change the array's length.
					var checked, safe bool
					checkEdits := func() {
						if checked {
							return
						}
						checked = true
						if evaluator == nil {
							evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
							evaluator.GlobalAccess = ctx.Globals.Access
						}
						if leading {
							safe = true
							for _, argument := range next.Arguments() {
								// Tags and spreads can run user code for an empty
								// array after removing the guard.
								if evaluator.HasSideEffect(argument, false) || hasEvaluatedSpread(argument) {
									safe = false
									break
								}
							}
						} else if arguments := previous.Arguments(); len(arguments) > 0 {
							safe = isSideEffectFreeFunction(evaluator, arguments[0])
						}
					}
					reportRange := checkRange(ctx.SourceFile, condition, member)
					message := checkMessage(messageID)
					ctx.ReportRangeWithDeferredFixesAndSuggestions(reportRange, message,
						func() []rule.RuleFix {
							checkEdits()
							if !safe {
								return nil
							}
							return removeLengthCheck(ctx, condition)
						},
						func() []rule.RuleSuggestion {
							checkEdits()
							if safe {
								return nil
							}
							fixes := removeLengthCheck(ctx, condition)
							if len(fixes) == 0 {
								return nil
							}
							return []rule.RuleSuggestion{{Message: rule.RuleMessage{Id: "remove-length-check", Description: "Remove the length check."}, FixesArr: fixes}}
						})
				}
			},
			ast.KindIfStatement: func(node *ast.Node) {
				statement := node.AsIfStatement()
				if statement.ElseStatement != nil {
					return
				}
				loop := statement.ThenStatement
				if ast.IsBlock(loop) {
					statements := loop.AsBlock().Statements.Nodes
					if len(statements) != 1 {
						return
					}
					loop = statements[0]
				}
				check := utils.ESTreeRuntimeExpression(statement.Expression)
				member := lengthCheck(ctx.SourceFile, check)
				if !ast.IsForOfStatement(loop) || member == nil || !isNonZeroOperator(check.AsBinaryExpression().OperatorToken.Kind) {
					return
				}
				forOf := loop.AsForInOrOfStatement()
				if forOf.AwaitModifier != nil || !unicornutil.IsSameReference(member.Expression(), forOf.Expression) ||
					hasLoopBindingReferenceInRight(ctx, loop) || unicornutil.IsKnownNonIndexedCollection(ctx, forOf.Expression) {
					return
				}
				ctx.ReportRangeWithDeferredFixes(checkRange(ctx.SourceFile, check, member), checkMessage("for-of"), func() []rule.RuleFix {
					return unwrapLoop(ctx, node, loop)
				})
			},
		}
	},
}

func lengthCheck(sourceFile *ast.SourceFile, node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil
	}
	comparison := node.AsBinaryExpression()
	left, right := utils.ESTreeRuntimeExpression(comparison.Left), utils.ESTreeRuntimeExpression(comparison.Right)
	if left.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChain(left) ||
		!ast.IsIdentifier(left.Name()) || left.Name().Text() != "length" ||
		right.Kind != ast.KindNumericLiteral || sourceFile.Text()[scanner.GetTokenPosOfNode(right, sourceFile, false):right.End()] != "0" {
		return nil
	}
	return left
}

func isNonZeroOperator(operator ast.Kind) bool {
	return operator == ast.KindGreaterThanToken || operator == ast.KindExclamationEqualsEqualsToken
}

func flattenLogical(node *ast.Node, operator ast.Kind, conditions *[]*ast.Node) {
	node = utils.ESTreeRuntimeExpression(node)
	if unicornutil.IsLogicalExpression(node) && node.AsBinaryExpression().OperatorToken.Kind == operator {
		flattenLogical(node.AsBinaryExpression().Left, operator, conditions)
		flattenLogical(node.AsBinaryExpression().Right, operator, conditions)
	} else {
		*conditions = append(*conditions, node)
	}
}

func isMatchingCall(ctx rule.RuleContext, node *ast.Node, method string, receiver *ast.Node) bool {
	// A continued optional chain is an ESTree ChainExpression, even when
	// neither the method access nor this call introduces the `?.` link.
	if node == nil || ast.IsOptionalChain(node) {
		return false
	}
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{Method: method})
	return ok && unicornutil.IsSameReference(receiver, call.Object) && !unicornutil.IsKnownNonIndexedCollection(ctx, call.Object)
}

func checkRange(sourceFile *ast.SourceFile, check, member *ast.Node) core.TextRange {
	return core.NewTextRange(scanner.GetTokenPosOfNode(member.Name(), sourceFile, false), check.End())
}

func checkMessage(id string) rule.RuleMessage {
	description := "The non-empty check is useless as `Array#some()` returns `false` for an empty array."
	switch id {
	case "zero":
		description = "The empty check is useless as `Array#every()` returns `true` for an empty array."
	case "for-of":
		description = "The non-empty check is useless as `for…of` does not iterate over an empty array."
	}
	return rule.RuleMessage{Id: id, Description: description}
}

func removeLengthCheck(ctx rule.RuleContext, check *ast.Node) []rule.RuleFix {
	parent := utils.ESTreeParent(check).AsBinaryExpression()
	left := core.NewTextRange(scanner.GetTokenPosOfNode(parent.Left, ctx.SourceFile, false), parent.Left.End())
	right := core.NewTextRange(scanner.GetTokenPosOfNode(parent.Right, ctx.SourceFile, false), parent.Right.End())
	removed := core.NewTextRange(left.End(), right.End())
	leading := utils.ESTreeRuntimeExpression(parent.Left) == check
	if leading {
		removed = core.NewTextRange(left.Pos(), right.Pos())
	}
	if utils.HasCommentInSpan(ctx.Comments.All(), removed.Pos(), removed.End()) {
		return nil
	}
	replacement := ""
	if leading && unicornutil.NeedsSemicolonBefore(ctx.SourceFile, parent.Left, ctx.SourceFile.Text()[removed.End():]) {
		replacement = ";"
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(removed, replacement)}
}

func isSideEffectFreeFunction(evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || (node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression) {
		return false
	}
	for _, parameter := range node.Parameters() {
		if evaluator.HasSideEffect(parameter, false) || containsSpreadOrTag(parameter) {
			return false
		}
	}
	return !evaluator.HasSideEffect(node.Body(), false) && !containsSpreadOrTag(node.Body())
}

func hasEvaluatedSpread(node *ast.Node) bool {
	found := false
	utils.VisitDescendants(node, func(child *ast.Node) bool {
		if found {
			return false
		}
		if ast.IsFunctionLike(child) {
			// Creating a callback or method does not evaluate its body or
			// parameters, but a computed method name is evaluated immediately.
			found = hasEvaluatedSpread(child.Name())
			return false
		}
		switch child.Kind {
		case ast.KindSpreadElement, ast.KindSpreadAssignment, ast.KindJsxSpreadAttribute:
			found = true
			return false
		}
		return !ast.IsTypeNode(child) || (child.Kind == ast.KindExpressionWithTypeArguments && !ast.IsPartOfTypeNode(child))
	})
	return found
}

func containsSpreadOrTag(node *ast.Node) bool {
	// A runtime instantiation shares its kind with type-only heritage nodes.
	if node != nil && node.Kind == ast.KindExpressionWithTypeArguments && !ast.IsPartOfTypeNode(node) {
		return containsSpreadOrTag(node.Expression())
	}
	if node == nil || ast.IsTypeNode(node) {
		return false
	}
	return node.Kind == ast.KindSpreadElement || node.Kind == ast.KindSpreadAssignment || node.Kind == ast.KindJsxSpreadAttribute || node.Kind == ast.KindTaggedTemplateExpression ||
		node.ForEachChild(containsSpreadOrTag)
}

func hasLoopBindingReferenceInRight(ctx rule.RuleContext, loop *ast.Node) bool {
	statement := loop.AsForInOrOfStatement()
	initializer := statement.Initializer
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
		return false
	}
	found := false
	for _, declaration := range initializer.AsVariableDeclarationList().Declarations.Nodes {
		utils.CollectBindingNames(declaration.Name(), func(identifier *ast.Node, _ string) {
			for _, reference := range ctx.Refs.References(utils.BindingNameSymbol(identifier)) {
				if reference.Pos() >= statement.Expression.Pos() && reference.End() <= statement.Expression.End() {
					found = true
				}
			}
		})
	}
	return found
}

func unwrapLoop(ctx rule.RuleContext, node, loop *ast.Node) []rule.RuleFix {
	sourceFile := ctx.SourceFile
	outerRange := core.NewTextRange(scanner.GetTokenPosOfNode(node, sourceFile, false), node.End())
	loopRange := core.NewTextRange(scanner.GetTokenPosOfNode(loop, sourceFile, false), loop.End())
	outerIndent, outerStartsAtIndent := indentAt(sourceFile, outerRange.Pos())
	loopIndent, loopStartsAtIndent := indentAt(sourceFile, loopRange.Pos())
	canUnindent := utils.IsSameLine(sourceFile, loopRange.Pos(), loopRange.End()) || (outerStartsAtIndent && loopStartsAtIndent)
	if !canUnindent {
		return nil
	}
	if utils.HasCommentInSpan(ctx.Comments.All(), outerRange.Pos(), loopRange.Pos()) ||
		utils.HasCommentInSpan(ctx.Comments.All(), loopRange.End(), outerRange.End()) {
		return nil
	}
	hasMultilineToken := false
	utils.ForEachToken(loop, func(token *ast.Node) {
		if hasMultilineToken {
			return
		}
		start := token.Pos()
		// Whitespace inside JSX text belongs to the token, including lines
		// containing only indentation before an expression or closing tag.
		if token.Kind != ast.KindJsxText {
			start = scanner.GetTokenPosOfNode(token, sourceFile, false)
		}
		hasMultilineToken = !utils.IsSameLine(sourceFile, start, token.End())
	}, sourceFile)
	if hasMultilineToken {
		return nil
	}
	text := sourceFile.Text()[loopRange.Pos():loopRange.End()]
	replacement := adjustLoopIndent(text, loopIndent, outerIndent)
	if !ast.IsBlock(loop.AsForInOrOfStatement().Statement) && text[len(text)-1] != ';' {
		if next, hasNext := utils.TokenAtOrAfter(sourceFile, outerRange.End()); hasNext {
			last, hasLast := utils.TokenBeforePosition(sourceFile, loopRange.End())
			if hasLast && unicornutil.NeedsSemicolonAfter(sourceFile, last, next.Text, utils.IsSameLine(sourceFile, outerRange.End(), next.Start)) {
				replacement += ";"
			}
		}
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(outerRange, replacement)}
}

func indentAt(sourceFile *ast.SourceFile, pos int) (string, bool) {
	line := scanner.ComputeLineOfPosition(sourceFile.ECMALineMap(), pos)
	prefix := sourceFile.Text()[sourceFile.ECMALineMap()[line]:pos]
	index := ecmascript.SkipTrailingWhitespace(prefix, 0, len(prefix))
	return prefix[index:], index == 0
}

func adjustLoopIndent(text, sourceIndent, targetIndent string) string {
	var result strings.Builder
	// Split after each ECMA line terminator so CRLF and mixed endings survive.
	lineStart, first := 0, true
	for index, character := range text {
		if !ecmascript.IsLineTerminator(character) {
			continue
		}
		line := text[lineStart:index]
		if !first {
			if ecmascript.IsBlank(line) {
				line = ""
			} else if strings.HasPrefix(line, sourceIndent) {
				line = targetIndent + strings.TrimPrefix(line, sourceIndent)
			}
		}
		result.WriteString(line)
		result.WriteRune(character)
		lineStart, first = index+utf8.RuneLen(character), false
	}
	line := text[lineStart:]
	if !first && strings.HasPrefix(line, sourceIndent) {
		line = targetIndent + strings.TrimPrefix(line, sourceIndent)
	}
	result.WriteString(line)
	return result.String()
}
