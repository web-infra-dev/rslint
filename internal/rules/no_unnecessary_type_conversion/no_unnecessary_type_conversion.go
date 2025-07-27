package no_unnecessary_type_conversion

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

var NoUnnecessaryTypeConversionRule = rule.Rule{
	Name: "no-unnecessary-type-conversion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return getListeners(ctx)
	},
}

type Options struct{}

func getListeners(ctx rule.RuleContext) rule.RuleListeners {
	return rule.RuleListeners{
		ast.KindCallExpression: func(node *ast.Node) {
			handleCallExpression(ctx, node)
		},

		ast.KindBinaryExpression: func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr.OperatorToken.Kind == ast.KindPlusToken {
				handleStringConcatenation(ctx, node)
			}
		},

		ast.KindPrefixUnaryExpression: func(node *ast.Node) {
			unaryExpr := node.AsPrefixUnaryExpression()
			if unaryExpr.Operator == ast.KindPlusToken {
				handleUnaryPlus(ctx, node)
			}
		},

	}
}

func doesUnderlyingTypeMatchFlag(ctx rule.RuleContext, typ *checker.Type, typeFlag checker.TypeFlags) bool {
	if typ == nil {
		return false
	}
	
	return utils.Every(utils.UnionTypeParts(typ), func(t *checker.Type) bool {
		return utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, typeFlag)
		})
	})
}

func handleCallExpression(ctx rule.RuleContext, node *ast.Node) {
	callExpr := node.AsCallExpression()
	callee := callExpr.Expression
	if callee.Kind != ast.KindIdentifier {
		return
	}
	
	calleeName := string(ctx.SourceFile.Text()[callee.Pos():callee.End()])
	
	// Map of built-in type constructors to their type flags
	builtInTypeFlags := map[string]checker.TypeFlags{
		"String":  checker.TypeFlagsStringLike,
		"Number":  checker.TypeFlagsNumberLike,
		"Boolean": checker.TypeFlagsBooleanLike,
		"BigInt":  checker.TypeFlagsBigIntLike,
	}
	
	typeFlag, ok := builtInTypeFlags[calleeName]
	if !ok {
		return
	}
	
	// For now, skip symbol checking to get basic functionality working
	// TODO: Add proper shadowing detection later
	_ = ctx.TypeChecker.GetSymbolAtLocation(callee)
	
	arguments := callExpr.Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return
	}
	
	argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arguments.Nodes[0])
	if !doesUnderlyingTypeMatchFlag(ctx, argType, typeFlag) {
		return
	}
	
	typeString := strings.ToLower(calleeName)
	message := fmt.Sprintf("Passing a %s to %s() does not change the type or value of the %s.", typeString, calleeName, typeString)
	
	argText := string(ctx.SourceFile.Text()[arguments.Nodes[0].Pos():arguments.Nodes[0].End()])
	ctx.ReportNodeWithSuggestions(callee, rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: message,
	}, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "suggestRemove",
			Description: "Remove the type conversion.",
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(node.Pos(), node.End()), argText),
		},
	})
}

func handleToStringCall(ctx rule.RuleContext, node *ast.Node) {
	memberExpr := node.Parent
	if memberExpr.Kind != ast.KindPropertyAccessExpression {
		return
	}
	
	propAccess := memberExpr.AsPropertyAccessExpression()
	object := propAccess.Expression
	objType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, object)
	
	if !doesUnderlyingTypeMatchFlag(ctx, objType, checker.TypeFlagsString) {
		return
	}
	
	callExpr := memberExpr.Parent
	message := "Calling a string's .toString() method does not change the type or value of the string."
	
	objText := string(ctx.SourceFile.Text()[object.Pos():object.End()])
	ctx.ReportRangeWithSuggestions(core.NewTextRange(propAccess.Name().Pos(), callExpr.End()), rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: message,
	}, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "suggestRemove",
			Description: "Remove the type conversion.",
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(callExpr.Pos(), callExpr.End()), objText),
		},
	})
}

func handleStringConcatenation(ctx rule.RuleContext, node *ast.Node) {
	binExpr := node.AsBinaryExpression()
	left := binExpr.Left
	right := binExpr.Right
	
	// Check if right is ''
	if right.Kind == ast.KindStringLiteral {
		strLit := right.AsStringLiteral()
		if strLit.Text == "" {
			leftType := ctx.TypeChecker.GetTypeAtLocation(left)
			if doesUnderlyingTypeMatchFlag(ctx, leftType, checker.TypeFlagsString) {
				message := "Concatenating a string with '' does not change the type or value of the string."
				reportStringConcatenation(ctx, node, left, core.NewTextRange(left.End(), node.End()), message, "Concatenating a string with ''")
			}
		}
	}
	
	// Check if left is ''
	if left.Kind == ast.KindStringLiteral {
		strLit := left.AsStringLiteral()
		if strLit.Text == "" {
			rightType := ctx.TypeChecker.GetTypeAtLocation(right)
			if doesUnderlyingTypeMatchFlag(ctx, rightType, checker.TypeFlagsString) {
				message := "Concatenating '' with a string does not change the type or value of the string."
				reportStringConcatenation(ctx, node, right, core.NewTextRange(node.Pos(), right.Pos()), message, "Concatenating '' with a string")
			}
		}
	}
}

func handleStringConcatenationAssignment(ctx rule.RuleContext, node *ast.Node) {
	assignExpr := node.AsBinaryExpression()
	left := assignExpr.Left
	right := assignExpr.Right
	
	if right.Kind != ast.KindStringLiteral {
		return
	}
	strLit := right.AsStringLiteral()
	if strLit.Text != "" {
		return
	}
	
	leftType := ctx.TypeChecker.GetTypeAtLocation(left)
	if !doesUnderlyingTypeMatchFlag(ctx, leftType, checker.TypeFlagsString) {
		return
	}
	
	message := "Concatenating a string with '' does not change the type or value of the string."
	
	// Check if this is in an expression statement
	isExpressionStatement := node.Parent != nil && node.Parent.Kind == ast.KindExpressionStatement
	
	suggestion := rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "suggestRemove",
			Description: "Remove the type conversion.",
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(
				core.NewTextRange(
					func() int {
					if isExpressionStatement {
						return node.Parent.Pos()
					}
					return node.Pos()
				}(),
				func() int {
					if isExpressionStatement {
						return node.Parent.End()
					}
					return node.End()
				}(),
				),
				func() string {
					if isExpressionStatement {
						return ""
					}
					return string(ctx.SourceFile.Text()[left.Pos():left.End()])
				}(),
			),
		},
	}

	ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: message,
	}, suggestion)
}

func reportStringConcatenation(ctx rule.RuleContext, node, innerNode *ast.Node, reportRange core.TextRange, message, violation string) {
	innerText := string(ctx.SourceFile.Text()[innerNode.Pos():innerNode.End()])
	ctx.ReportRangeWithSuggestions(reportRange, rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: message,
	}, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "suggestRemove",
			Description: "Remove the type conversion.",
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(node.Pos(), node.End()), innerText),
		},
	})
}

func handleUnaryPlus(ctx rule.RuleContext, node *ast.Node) {
	unaryExpr := node.AsPrefixUnaryExpression()
	operand := unaryExpr.Operand
	operandType := ctx.TypeChecker.GetTypeAtLocation(operand)
	
	if !doesUnderlyingTypeMatchFlag(ctx, operandType, checker.TypeFlagsNumber) {
		return
	}
	
	handleUnaryOperator(ctx, node, "number", "Using the unary + operator on a number", false)
}

func handleDoubleNegation(ctx rule.RuleContext, node *ast.Node) {
	unaryExpr := node.AsPrefixUnaryExpression()
	operand := unaryExpr.Operand
	operandType := ctx.TypeChecker.GetTypeAtLocation(operand)
	
	if !doesUnderlyingTypeMatchFlag(ctx, operandType, checker.TypeFlagsBoolean) {
		return
	}
	
	handleUnaryOperator(ctx, node, "boolean", "Using !! on a boolean", true)
}

func handleDoubleTilde(ctx rule.RuleContext, node *ast.Node) {
	unaryExpr := node.AsPrefixUnaryExpression()
	operand := unaryExpr.Operand
	operandType := ctx.TypeChecker.GetTypeAtLocation(operand)
	
	if !doesUnderlyingTypeMatchFlag(ctx, operandType, checker.TypeFlagsNumber) {
		return
	}
	
	handleUnaryOperator(ctx, node, "number", "Using ~~ on a number", true)
}

func handleUnaryOperator(ctx rule.RuleContext, node *ast.Node, typeString, violation string, isDoubleOperator bool) {
	outerNode := node
	if isDoubleOperator && node.Parent != nil {
		outerNode = node.Parent
	}
	
	unaryExpr := node.AsPrefixUnaryExpression()
	operand := unaryExpr.Operand
	
	message := fmt.Sprintf("%s does not change the type or value of the %s.", violation, typeString)
	
	reportRange := core.NewTextRange(
		outerNode.Pos(),
		func() int {
			if isDoubleOperator {
				// For double operators, highlight up to the second operator
				return node.Pos() + 1
			}
			// For single operators, highlight just the operator
			return node.Pos() + 1
		}(),
	)
	
	operandText := string(ctx.SourceFile.Text()[operand.Pos():operand.End()])
	ctx.ReportRangeWithSuggestions(reportRange, rule.RuleMessage{
		Id:          "unnecessaryTypeConversion",
		Description: message,
	}, rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          "suggestRemove",
			Description: "Remove the type conversion.",
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(outerNode.Pos(), outerNode.End()), operandText),
		},
	})
}