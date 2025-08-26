package no_unnecessary_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoUnnecessaryConditionOptions struct {
	AllowConstantLoopConditions                            *string `json:"allowConstantLoopConditions,omitempty"`
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool   `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing,omitempty"`
	CheckTypePredicates                                    *bool   `json:"checkTypePredicates,omitempty"`
}

func parseOptions(options any) NoUnnecessaryConditionOptions {
	opts := NoUnnecessaryConditionOptions{
		AllowConstantLoopConditions:                            utils.Ref("never"),
		AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(false),
		CheckTypePredicates:                                    utils.Ref(false),
	}

	if options == nil {
		return opts
	}

	// Handle array format: [{ option: value }]
	if arr, ok := options.([]any); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				if v, ok := m["allowConstantLoopConditions"].(string); ok {
					opts.AllowConstantLoopConditions = utils.Ref(v)
				}
				if v, ok := m["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
					opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(v)
				}
				if v, ok := m["checkTypePredicates"].(bool); ok {
					opts.CheckTypePredicates = utils.Ref(v)
				}
			}
		}
	}

	return opts
}

// Rule message builders
func buildAlwaysFalsyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsy",
		Description: "Unnecessary conditional, value is always falsy.",
	}
}

func buildAlwaysTruthyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthy",
		Description: "Unnecessary conditional, value is always truthy.",
	}
}

func buildNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "never",
		Description: "Unnecessary conditional, value is `never`.",
	}
}

func buildAlwaysNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysNullish",
		Description: "Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.",
	}
}

func buildNeverNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverNullish",
		Description: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildComparisonBetweenLiteralTypesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparisonBetweenLiteralTypes",
		Description: "Unnecessary conditional, comparison is always true or false.",
	}
}

// Type checking utilities using the correct RSLint APIs
func isNeverType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsNever)
}

func isAnyType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsAny)
}

func isUnknownType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsUnknown)
}

func isNullType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsNull)
}

func isUndefinedType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsUndefined)
}

func isVoidType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsVoid)
}

func isBooleanLiteralType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsBooleanLiteral)
}

func isNumberLiteralType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsNumberLiteral)
}

func isStringLiteralType(typeOfNode *checker.Type) bool {
	return utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsStringLiteral)
}

// Check if type could be nullish (null | undefined)
func isPossiblyNullish(typeOfNode *checker.Type) bool {
	if isNullType(typeOfNode) || isUndefinedType(typeOfNode) || isVoidType(typeOfNode) {
		return true
	}

	// For union types, check if any constituent could be nullish
	if utils.IsUnionType(typeOfNode) {
		for _, unionType := range utils.UnionTypeParts(typeOfNode) {
			if isPossiblyNullish(unionType) {
				return true
			}
		}
		return false
	}

	return false
}

// Check if type is always nullish
func isAlwaysNullish(typeOfNode *checker.Type) bool {
	if isNullType(typeOfNode) || isUndefinedType(typeOfNode) || isVoidType(typeOfNode) {
		return true
	}

	// For union types, check if all constituents are nullish
	if utils.IsUnionType(typeOfNode) {
		for _, unionType := range utils.UnionTypeParts(typeOfNode) {
			if !isAlwaysNullish(unionType) {
				return false
			}
		}
		return true
	}

	return false
}

// Check if type could be truthy
func isPossiblyTruthy(typeOfNode *checker.Type) bool {
	// Always falsy types: null, undefined, void, false, 0, "", NaN
	if isNullType(typeOfNode) || isUndefinedType(typeOfNode) || isVoidType(typeOfNode) {
		return false
	}

	// For literal types, we conservatively assume they could be truthy
	// A more complete implementation would check the actual literal values
	if isBooleanLiteralType(typeOfNode) || isNumberLiteralType(typeOfNode) || isStringLiteralType(typeOfNode) {
		// This is a simplification - would need to check if it's specifically false, 0, or ""
		// For now, assume it could be truthy
		return true
	}

	// For union types, check if any constituent could be truthy
	if utils.IsUnionType(typeOfNode) {
		for _, unionType := range utils.UnionTypeParts(typeOfNode) {
			if isPossiblyTruthy(unionType) {
				return true
			}
		}
		return false
	}

	// For other types, assume they could be truthy
	return true
}

// Check if type could be falsy
func isPossiblyFalsy(typeOfNode *checker.Type) bool {
	// Always falsy types
	if isNullType(typeOfNode) || isUndefinedType(typeOfNode) || isVoidType(typeOfNode) {
		return true
	}

	// Literal types could be falsy values
	if isBooleanLiteralType(typeOfNode) || isNumberLiteralType(typeOfNode) || isStringLiteralType(typeOfNode) {
		// This is a simplification - would need to check if it's specifically false, 0, or ""
		return true
	}

	// Union types might contain falsy values
	if utils.IsUnionType(typeOfNode) {
		for _, unionType := range utils.UnionTypeParts(typeOfNode) {
			if isPossiblyFalsy(unionType) {
				return true
			}
		}
		return false
	}

	// Conservative: most types could potentially be falsy
	return true
}

// Check if conditional is always necessary (any, unknown, type variables)
func isConditionalAlwaysNecessary(typeOfNode *checker.Type) bool {
	return isAnyType(typeOfNode) || isUnknownType(typeOfNode) ||
		utils.IsTypeFlagSet(typeOfNode, checker.TypeFlagsTypeParameter)
}

// Check boolean operators
func isBooleanOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindLessThanToken, ast.KindGreaterThanToken,
		ast.KindLessThanEqualsToken, ast.KindGreaterThanEqualsToken,
		ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return true
	default:
		return false
	}
}

// Main condition checking logic
func checkCondition(ctx rule.RuleContext, node *ast.Node, isUnaryNotArgument bool) {
	// Handle unary not expressions
	if node.Kind == ast.KindPrefixUnaryExpression {
		unaryExpr := node.AsPrefixUnaryExpression()
		if unaryExpr != nil && unaryExpr.Operator == ast.KindExclamationToken {
			checkCondition(ctx, unaryExpr.Operand, !isUnaryNotArgument)
			return
		}
	}

	// Get type of the expression
	typeOfNode := ctx.TypeChecker.GetTypeAtLocation(node)
	if typeOfNode == nil {
		return
	}

	// Skip if conditional is always necessary
	if isConditionalAlwaysNecessary(typeOfNode) {
		return
	}

	var messageBuilder func() rule.RuleMessage

	if isNeverType(typeOfNode) {
		messageBuilder = buildNeverMessage
	} else if !isPossiblyTruthy(typeOfNode) {
		if isUnaryNotArgument {
			messageBuilder = buildAlwaysTruthyMessage
		} else {
			messageBuilder = buildAlwaysFalsyMessage
		}
	} else if !isPossiblyFalsy(typeOfNode) {
		if isUnaryNotArgument {
			messageBuilder = buildAlwaysFalsyMessage
		} else {
			messageBuilder = buildAlwaysTruthyMessage
		}
	}

	if messageBuilder != nil {
		ctx.ReportNode(node, messageBuilder())
	}
}

// Check nullish coalescing expressions
func checkNullishCoalescing(ctx rule.RuleContext, expr *ast.BinaryExpression) {
	leftType := ctx.TypeChecker.GetTypeAtLocation(expr.Left)
	if leftType == nil {
		return
	}

	var messageBuilder func() rule.RuleMessage

	if isNeverType(leftType) {
		messageBuilder = buildNeverMessage
	} else if !isPossiblyNullish(leftType) {
		messageBuilder = buildNeverNullishMessage
	} else if isAlwaysNullish(leftType) {
		messageBuilder = buildAlwaysNullishMessage
	}

	if messageBuilder != nil {
		ctx.ReportNode(expr.Left, messageBuilder())
	}
}

var NoUnnecessaryConditionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Check for strict null checks
		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks && !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			// Report at the beginning of the file
			ctx.ReportNode(&ast.Node{}, buildNoStrictNullCheckMessage())
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			// If statement conditions
			ast.KindIfStatement: func(node *ast.Node) {
				ifStmt := node.AsIfStatement()
				if ifStmt != nil {
					checkCondition(ctx, ifStmt.Expression, false)
				}
			},

			// While loop conditions
			ast.KindWhileStatement: func(node *ast.Node) {
				whileStmt := node.AsWhileStatement()
				if whileStmt != nil {
					// Handle constant loop conditions
					if *opts.AllowConstantLoopConditions == "always" {
						typeOfCondition := ctx.TypeChecker.GetTypeAtLocation(whileStmt.Expression)
						if typeOfCondition != nil {
							// Skip if it's a constant true condition
							// This would require checking for true literal type
							return
						}
					}
					checkCondition(ctx, whileStmt.Expression, false)
				}
			},

			// For loop conditions
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt != nil && forStmt.Condition != nil {
					checkCondition(ctx, forStmt.Condition, false)
				}
			},

			// Do-while loop conditions
			ast.KindDoStatement: func(node *ast.Node) {
				doStmt := node.AsDoStatement()
				if doStmt != nil {
					checkCondition(ctx, doStmt.Expression, false)
				}
			},

			// Conditional expressions (ternary)
			ast.KindConditionalExpression: func(node *ast.Node) {
				condExpr := node.AsConditionalExpression()
				if condExpr != nil {
					checkCondition(ctx, condExpr.Condition, false)
				}
			},

			// Binary expressions (comparisons and logical expressions)
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr != nil {
					// Handle nullish coalescing
					if binExpr.OperatorToken.Kind == ast.KindQuestionQuestionToken {
						checkNullishCoalescing(ctx, binExpr)
						return
					}

					// Handle logical AND/OR
					if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
						binExpr.OperatorToken.Kind == ast.KindBarBarToken {
						checkCondition(ctx, binExpr.Left, false)
						return
					}

					// Handle boolean comparisons
					if isBooleanOperator(binExpr.OperatorToken.Kind) {
						leftType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Left)
						rightType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Right)

						if leftType != nil && rightType != nil {
							// Check if both sides are literal types
							// This would require extracting literal values from types
							// For now, skip this complex comparison logic
							// A full implementation would check for literal type comparisons like:
							// if (true === true) -> always true
							// if (1 === 2) -> always false
						}
					}
				}
			},
		}
	},
})
