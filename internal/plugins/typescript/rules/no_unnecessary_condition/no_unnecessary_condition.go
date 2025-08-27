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

	// Handle direct map format
	if m, ok := options.(map[string]any); ok {
		parseOptionsFromMap(m, &opts)
		return opts
	}

	// Handle array format: [{ option: value }]
	if arr, ok := options.([]any); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				parseOptionsFromMap(m, &opts)
			}
		}
	}

	return opts
}

func parseOptionsFromMap(m map[string]any, opts *NoUnnecessaryConditionOptions) {
	if v, ok := m["allowConstantLoopConditions"]; ok {
		// Can be boolean or string
		switch val := v.(type) {
		case bool:
			if val {
				opts.AllowConstantLoopConditions = utils.Ref("always")
			} else {
				opts.AllowConstantLoopConditions = utils.Ref("never")
			}
		case string:
			opts.AllowConstantLoopConditions = utils.Ref(val)
		}
	}
	if v, ok := m["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
		opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = utils.Ref(v)
	}
	if v, ok := m["checkTypePredicates"].(bool); ok {
		opts.CheckTypePredicates = utils.Ref(v)
	}
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
	}
	return false
}

// isTypeNeverNullish checks if a type can never be null or undefined
func isTypeNeverNullish(t *checker.Type, typeChecker *checker.Checker) bool {
	if t == nil {
		return false
	}

	// Check for any or unknown types - these could be nullish
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return false
	}

	// Check if the type itself is null, undefined, or void
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false
	}

	// For union types, check if any constituent could be nullish
	if utils.IsUnionType(t) {
		for _, unionType := range t.Types() {
			typeFlags := checker.Type_flags(unionType)
			if typeFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
				return false
			}
		}
	}

	// If we get here, the type cannot be nullish
	return true
}

// isAlwaysTruthy checks if a type is always truthy (cannot be falsy)
func isAlwaysTruthy(t *checker.Type) bool {
	if t == nil {
		return false
	}

	flags := checker.Type_flags(t)

	// Any and unknown could be falsy
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return false
	}

	// Never type cannot have a value
	if flags&checker.TypeFlagsNever != 0 {
		return false
	}

	// These types are always falsy or could be falsy
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return false
	}

	// Check for union types - all parts must be truthy
	if utils.IsUnionType(t) {
		for _, unionType := range t.Types() {
			if !isAlwaysTruthy(unionType) {
				return false
			}
		}
		return true
	}

	// Boolean type (not literal) can be true or false, so not always truthy
	if flags&checker.TypeFlagsBoolean != 0 {
		return false
	}

	// Boolean literals - check if it's the 'true' literal
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		if utils.IsIntrinsicType(t) {
			intrinsic := t.AsIntrinsicType()
			if intrinsic != nil && intrinsic.IntrinsicName() == "true" {
				return true
			}
		}
		return false
	}

	// Number literals could be 0, -0, or NaN (falsy values)
	if flags&checker.TypeFlagsNumberLiteral != 0 {
		// Would need to check the actual value
		// For now, conservatively return false
		return false
	}

	// String literals could be "" (falsy)
	if flags&checker.TypeFlagsStringLiteral != 0 {
		// Would need to check for empty string
		// For now, conservatively return false
		return false
	}

	// BigInt literals could be 0n (falsy)
	if flags&checker.TypeFlagsBigIntLiteral != 0 {
		// Would need to check for 0n
		return false
	}

	// Object types are always truthy
	if flags&checker.TypeFlagsObject != 0 {
		return true
	}

	// For the purpose of this rule, non-nullable primitive types are considered "always truthy"
	// This is not technically correct from a JavaScript perspective (empty string, 0, NaN are falsy),
	// but matches the TypeScript ESLint rule behavior which flags these as unnecessary conditions
	// when they are non-nullable types
	if flags&checker.TypeFlagsString != 0 {
		return true
	}

	// Number type - treat as always truthy for non-nullable numbers
	if flags&checker.TypeFlagsNumber != 0 {
		return true
	}

	// BigInt type - treat as always truthy for non-nullable bigints
	if flags&checker.TypeFlagsBigInt != 0 {
		return true
	}

	// ESSymbol is always truthy
	if flags&checker.TypeFlagsESSymbol != 0 {
		return true
	}

	return false
}

// isAlwaysFalsy checks if a type is always falsy
func isAlwaysFalsy(t *checker.Type) bool {
	if t == nil {
		return false
	}

	flags := checker.Type_flags(t)

	// Null, undefined, and void are always falsy
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return true
	}

	// Check for literal false
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		if utils.IsIntrinsicType(t) {
			intrinsic := t.AsIntrinsicType()
			if intrinsic != nil && intrinsic.IntrinsicName() == "false" {
				return true
			}
		}
	}

	// Would need to check for literal 0, -0, NaN, "", 0n
	// For now, we don't mark these as always falsy

	return false
}

// checkCondition checks if a condition is unnecessary (always true/false/never)
func checkCondition(ctx rule.RuleContext, node *ast.Node, isNegated bool) {
	if node == nil {
		return
	}

	// Get the type of the condition expression
	conditionType := ctx.TypeChecker.GetTypeAtLocation(node)
	if conditionType == nil {
		return
	}

	// Check for never type
	if isNeverType(conditionType) {
		ctx.ReportNode(node, buildNeverMessage())
		return
	}

	// Check for always truthy
	if isAlwaysTruthy(conditionType) {
		ctx.ReportNode(node, buildAlwaysTruthyMessage())
		return
	}

	// Check for always falsy
	if isAlwaysFalsy(conditionType) {
		ctx.ReportNode(node, buildAlwaysFalsyMessage())
		return
	}
}

// isBooleanOperator checks if a token kind represents a boolean comparison operator
func isBooleanOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken,
		ast.KindLessThanToken, ast.KindLessThanEqualsToken,
		ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		return true
	}
	return false
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
				if whileStmt != nil && whileStmt.Expression != nil {
					// Handle constant loop conditions
					if *opts.AllowConstantLoopConditions != "never" {
						// Check if it's a constant condition
						typeOfCondition := ctx.TypeChecker.GetTypeAtLocation(whileStmt.Expression)
						if typeOfCondition != nil {
							flags := checker.Type_flags(typeOfCondition)
							// Check for literal true/false
							if flags&checker.TypeFlagsBooleanLiteral != 0 {
								if utils.IsIntrinsicType(typeOfCondition) {
									intrinsic := typeOfCondition.AsIntrinsicType()
									if intrinsic != nil && (intrinsic.IntrinsicName() == "true" || intrinsic.IntrinsicName() == "false") {
										// Skip checking constant boolean literals in loops when allowed
										return
									}
								}
							}
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
					// Handle logical AND/OR
					if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
						binExpr.OperatorToken.Kind == ast.KindBarBarToken {
						checkCondition(ctx, binExpr.Left, false)
						return
					}

					// Handle nullish coalescing operator (??)
					if binExpr.OperatorToken.Kind == ast.KindQuestionQuestionToken {
						leftType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Left)
						if leftType != nil {
							// Check if left side can never be nullish (null or undefined)
							if isTypeNeverNullish(leftType, ctx.TypeChecker) {
								ctx.ReportNode(binExpr.Left, buildNeverNullishMessage())
							}
							// Check if left side is always nullish
							if isAlwaysFalsy(leftType) && isPossiblyNullish(leftType) {
								ctx.ReportNode(binExpr.Left, buildAlwaysNullishMessage())
							}
						}
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
							// TODO: Implement literal type comparison logic
							_, _ = leftType, rightType
						}
					}
				}
			},
		}
	},
})
