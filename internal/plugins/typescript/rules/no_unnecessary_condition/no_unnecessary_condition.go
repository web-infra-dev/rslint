package no_unnecessary_condition

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type AllowConstantLoopConditions string

const (
	AllowConstantLoopConditionsNever              AllowConstantLoopConditions = "never"
	AllowConstantLoopConditionsAlways             AllowConstantLoopConditions = "always"
	AllowConstantLoopConditionsOnlyAllowedLiterals AllowConstantLoopConditions = "only-allowed-literals"
)

type Options struct {
	AllowConstantLoopConditions AllowConstantLoopConditions `json:"allowConstantLoopConditions"`
	CheckTypePredicates         bool                        `json:"checkTypePredicates"`
	AllowRuleToRunWithoutStrictNullChecks bool              `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
}

func parseOptions(options any) Options {
	opts := Options{
		AllowConstantLoopConditions: AllowConstantLoopConditionsNever,
		CheckTypePredicates:         false,
		AllowRuleToRunWithoutStrictNullChecks: false,
	}

	if options == nil {
		return opts
	}

	optionsMap, ok := options.(map[string]any)
	if !ok {
		return opts
	}

	if val, ok := optionsMap["allowConstantLoopConditions"]; ok {
		if strVal, ok := val.(string); ok {
			opts.AllowConstantLoopConditions = AllowConstantLoopConditions(strVal)
		} else if boolVal, ok := val.(bool); ok {
			if boolVal {
				opts.AllowConstantLoopConditions = AllowConstantLoopConditionsAlways
			} else {
				opts.AllowConstantLoopConditions = AllowConstantLoopConditionsNever
			}
		}
	}

	if val, ok := optionsMap["checkTypePredicates"]; ok {
		if boolVal, ok := val.(bool); ok {
			opts.CheckTypePredicates = boolVal
		}
	}

	if val, ok := optionsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"]; ok {
		if boolVal, ok := val.(bool); ok {
			opts.AllowRuleToRunWithoutStrictNullChecks = boolVal
		}
	}

	return opts
}

func buildAlwaysTruthyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthy",
		Description: "Unnecessary conditional, value is always truthy.",
	}
}

func buildAlwaysFalsyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsy",
		Description: "Unnecessary conditional, value is always falsy.",
	}
}

func buildAlwaysTruthyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysTruthyFunc",
		Description: "This callback should return a conditional, but return is always truthy.",
	}
}

func buildAlwaysFalsyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "alwaysFalsyFunc",
		Description: "This callback should return a conditional, but return is always falsy.",
	}
}

func buildLiteralBooleanExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "literalBooleanExpression",
		Description: "Unnecessary conditional, both sides are literal values.",
	}
}

func buildNeverNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverNullish",
		Description: "Unnecessary conditional, left-hand side is never nullish.",
	}
}

func buildNeverOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "neverOptionalChain",
		Description: "Unnecessary optional chain on a non-nullish value.",
	}
}

func buildNoOverlapMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noOverlap",
		Description: "This comparison will always result in the same boolean value.",
	}
}

type Truthiness int

const (
	TruthinessTruthy Truthiness = iota
	TruthinessFalsy
	TruthinessMaybeTruthy
)

var NoUnnecessaryConditionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-condition",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		compilerOptions := ctx.Program.Options()
		strictNullChecksEnabled := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !strictNullChecksEnabled && !opts.AllowRuleToRunWithoutStrictNullChecks {
			return rule.RuleListeners{}
		}

		// Helper to check if node is a comparison binary expression
		isComparisonExpression := func(node *ast.Node) bool {
			if !ast.IsBinaryExpression(node) {
				return false
			}
			op := node.AsBinaryExpression().OperatorToken.Kind
			return op == ast.KindEqualsEqualsToken ||
				op == ast.KindExclamationEqualsToken ||
				op == ast.KindEqualsEqualsEqualsToken ||
				op == ast.KindExclamationEqualsEqualsToken ||
				op == ast.KindLessThanToken ||
				op == ast.KindGreaterThanToken ||
				op == ast.KindLessThanEqualsToken ||
				op == ast.KindGreaterThanEqualsToken
		}

		// Declare checkTypeIsTruthy first so it can be used recursively
		var checkTypeIsTruthy func(t *checker.Type) Truthiness

		// Check if a type is always truthy, always falsy, or may be either
		checkTypeIsTruthy = func(t *checker.Type) Truthiness {
			// Handle any/unknown
			if utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) {
				return TruthinessMaybeTruthy
			}

			flags := checker.Type_flags(t)

			// Handle unions - must check all parts
			if utils.IsUnionType(t) {
				var hasTruthy, hasFalsy bool
				for _, part := range utils.UnionTypeParts(t) {
					truthiness := checkTypeIsTruthy(part)
					if truthiness == TruthinessMaybeTruthy {
						return TruthinessMaybeTruthy
					}
					if truthiness == TruthinessTruthy {
						hasTruthy = true
					} else {
						hasFalsy = true
					}
				}
				if hasTruthy && hasFalsy {
					return TruthinessMaybeTruthy
				}
				if hasTruthy {
					return TruthinessTruthy
				}
				return TruthinessFalsy
			}

			// Handle intersection types (e.g., string & { __brand: 'Brand' })
			// For branded types and other intersections, check each constituent
			if utils.IsIntersectionType(t) {
				// Get the constituents of the intersection
				parts := utils.IntersectionTypeParts(t)
				if len(parts) > 0 {
					// Check each part
					for _, part := range parts {
						partFlags := checker.Type_flags(part)
						// If any part is a base primitive type (string/number/boolean/bigint),
						// the intersection can have the same truthiness variability
						if partFlags&(checker.TypeFlagsString|checker.TypeFlagsNumber|checker.TypeFlagsBoolean|checker.TypeFlagsBigInt) != 0 {
							return TruthinessMaybeTruthy
						}
					}
				}
			}

			// Falsy types
			if flags&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				return TruthinessFalsy
			}

			// Check boolean literals first (before generic boolean check)
			if flags&checker.TypeFlagsBooleanLiteral != 0 {
				if utils.IsTrueLiteralType(ctx.TypeChecker, t) {
					return TruthinessTruthy
				}
				if utils.IsFalseLiteralType(ctx.TypeChecker, t) {
					return TruthinessFalsy
				}
				// If we have the boolean literal flag but can't determine which one,
				// treat as maybe truthy (shouldn't happen in practice)
				return TruthinessMaybeTruthy
			}

			// Check generic boolean - could be true or false
			if flags&checker.TypeFlagsBoolean != 0 {
				return TruthinessMaybeTruthy
			}

			// Check for empty string literal
			if flags&checker.TypeFlagsStringLiteral != 0 {
				// Get the string value by converting type to string
				typeStr := ctx.TypeChecker.TypeToString(t)
				// Empty string literals are represented as ""
				if typeStr == `""` || typeStr == "" {
					return TruthinessFalsy
				}
				// Non-empty string literal is truthy
				return TruthinessTruthy
			}

			// Check for 0 or -0 numeric literal
			if flags&checker.TypeFlagsNumberLiteral != 0 {
				typeStr := ctx.TypeChecker.TypeToString(t)
				if typeStr == "0" || typeStr == "-0" {
					return TruthinessFalsy
				}
				// Non-zero number literal is truthy
				return TruthinessTruthy
			}

			// Check for bigint 0
			if flags&checker.TypeFlagsBigIntLiteral != 0 {
				typeStr := ctx.TypeChecker.TypeToString(t)
				if typeStr == "0n" {
					return TruthinessFalsy
				}
				// Non-zero bigint literal is truthy
				return TruthinessTruthy
			}

			// Check generic string - could be empty or non-empty
			if flags&checker.TypeFlagsString != 0 {
				return TruthinessMaybeTruthy
			}

			// Check generic number - could be 0 or non-zero
			if flags&checker.TypeFlagsNumber != 0 {
				return TruthinessMaybeTruthy
			}

			// Check generic bigint - could be 0n or non-zero
			if flags&checker.TypeFlagsBigInt != 0 {
				return TruthinessMaybeTruthy
			}

			// Type parameters need special handling
			if utils.IsTypeParameter(t) {
				constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
				if constraint == nil {
					return TruthinessMaybeTruthy
				}
				return checkTypeIsTruthy(constraint)
			}

			// Everything else is truthy (objects, functions, etc.)
			return TruthinessTruthy
		}

		// Check if a node represents a condition
		checkNode := func(node *ast.Node, isRoot bool) {
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			constrainedType, isTypeParam := utils.GetConstraintInfo(ctx.TypeChecker, t)

			// If it's an unconstrained generic, we can't determine truthiness
			if isTypeParam && constrainedType == nil {
				return
			}

			typeToCheck := constrainedType
			if typeToCheck == nil {
				typeToCheck = t
			}

			truthiness := checkTypeIsTruthy(typeToCheck)

			if truthiness == TruthinessTruthy {
				ctx.ReportNode(node, buildAlwaysTruthyMessage())
			} else if truthiness == TruthinessFalsy {
				ctx.ReportNode(node, buildAlwaysFalsyMessage())
			}
		}

		// Check binary expressions like === or !==
		checkBinaryExpression := func(node *ast.Node) {
			expr := node.AsBinaryExpression()
			op := expr.OperatorToken.Kind

			// Only check equality/inequality operators
			if op != ast.KindEqualsEqualsToken &&
				op != ast.KindExclamationEqualsToken &&
				op != ast.KindEqualsEqualsEqualsToken &&
				op != ast.KindExclamationEqualsEqualsToken &&
				op != ast.KindLessThanToken &&
				op != ast.KindGreaterThanToken &&
				op != ast.KindLessThanEqualsToken &&
				op != ast.KindGreaterThanEqualsToken {
				return
			}

			left := expr.Left
			right := expr.Right

			leftType := ctx.TypeChecker.GetTypeAtLocation(left)
			rightType := ctx.TypeChecker.GetTypeAtLocation(right)

			// Check if comparing literal types that can never be equal
			leftFlags := checker.Type_flags(leftType)
			rightFlags := checker.Type_flags(rightType)

			// Check for literal comparisons where both are the same literal
			if leftFlags&checker.TypeFlagsLiteral != 0 && rightFlags&checker.TypeFlagsLiteral != 0 {
				// Get the type strings to compare
				leftStr := ctx.TypeChecker.TypeToString(leftType)
				rightStr := ctx.TypeChecker.TypeToString(rightType)

				// For equality/inequality, check if they're the same literal
				if op == ast.KindEqualsEqualsEqualsToken || op == ast.KindEqualsEqualsToken {
					// If comparing same literal with ===, always true
					if leftStr == rightStr {
						ctx.ReportNode(node, buildNoOverlapMessage())
					}
				}
			}
		}

		// Check nullish coalescing operator (??)
		checkNullishCoalescing := func(node *ast.Node) {
			expr := node.AsBinaryExpression()
			left := expr.Left

			leftType := ctx.TypeChecker.GetTypeAtLocation(left)
			constrainedType, isTypeParam := utils.GetConstraintInfo(ctx.TypeChecker, leftType)

			if isTypeParam && constrainedType == nil {
				return
			}

			typeToCheck := constrainedType
			if typeToCheck == nil {
				typeToCheck = leftType
			}

			// Check if left side can ever be nullish
			canBeNullish := false
			for _, part := range utils.UnionTypeParts(typeToCheck) {
				flags := checker.Type_flags(part)
				if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
					canBeNullish = true
					break
				}
			}

			if !canBeNullish && !utils.IsTypeAnyType(typeToCheck) && !utils.IsTypeUnknownType(typeToCheck) {
				ctx.ReportNode(left, buildNeverNullishMessage())
			}
		}

		// Check optional chaining
		checkOptionalChain := func(node *ast.Node) {
			var expr *ast.Node

			if ast.IsPropertyAccessExpression(node) {
				if node.AsPropertyAccessExpression().QuestionDotToken == nil {
					return
				}
				expr = node.Expression()
			} else if ast.IsElementAccessExpression(node) {
				if node.AsElementAccessExpression().QuestionDotToken == nil {
					return
				}
				expr = node.Expression()
			} else if ast.IsCallExpression(node) {
				if node.AsCallExpression().QuestionDotToken == nil {
					return
				}
				expr = node.Expression()
			} else {
				return
			}

			exprType := ctx.TypeChecker.GetTypeAtLocation(expr)
			constrainedType, isTypeParam := utils.GetConstraintInfo(ctx.TypeChecker, exprType)

			if isTypeParam && constrainedType == nil {
				return
			}

			typeToCheck := constrainedType
			if typeToCheck == nil {
				typeToCheck = exprType
			}

			// Check if expression can ever be nullish
			canBeNullish := false
			for _, part := range utils.UnionTypeParts(typeToCheck) {
				flags := checker.Type_flags(part)
				if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
					canBeNullish = true
					break
				}
			}

			if !canBeNullish && !utils.IsTypeAnyType(typeToCheck) && !utils.IsTypeUnknownType(typeToCheck) {
				ctx.ReportNode(node, buildNeverOptionalChainMessage())
			}
		}

		// Check loop conditions
		checkLoopCondition := func(node *ast.Node, condition *ast.Node) {
			if condition == nil {
				return
			}

			// Check the option setting
			if opts.AllowConstantLoopConditions == AllowConstantLoopConditionsAlways {
				return
			}

			if opts.AllowConstantLoopConditions == AllowConstantLoopConditionsOnlyAllowedLiterals {
				// Allow literal true/false and numeric literals 0 and 1
				if condition.Kind == ast.KindTrueKeyword || condition.Kind == ast.KindFalseKeyword {
					return
				}
				if ast.IsNumericLiteral(condition) {
					text := condition.Text()
					if text == "0" || text == "1" {
						return
					}
				}
			}

			checkNode(condition, true)
		}

		// Helper to check if expression should skip condition checking
		shouldSkipConditionCheck := func(expr *ast.Node) bool {
			if isComparisonExpression(expr) {
				return true
			}
			if ast.IsPrefixUnaryExpression(expr) {
				return true
			}
			if ast.IsBinaryExpression(expr) {
				op := expr.AsBinaryExpression().OperatorToken.Kind
				// Skip logical operators - they handle their own checks
				if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
					return true
				}
			}
			// Skip if the expression type is already boolean (not a boolean literal)
			// Boolean types are meant to be used in conditionals
			exprType := ctx.TypeChecker.GetTypeAtLocation(expr)
			flags := checker.Type_flags(exprType)
			if flags&checker.TypeFlagsBoolean != 0 {
				// Make sure it's not a boolean literal (true/false), just generic boolean
				if !utils.IsTrueLiteralType(ctx.TypeChecker, exprType) && !utils.IsFalseLiteralType(ctx.TypeChecker, exprType) {
					return true
				}
			}
			return false
		}

		// Check array predicate callbacks
		checkArrayPredicate := func(node *ast.Node) {
			callExpr := node.AsCallExpression()

			if !utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, callExpr) {
				return
			}

			if len(callExpr.Arguments.Nodes) == 0 {
				return
			}

			callback := callExpr.Arguments.Nodes[0]

			// Get the return type of the callback
			var returnNode *ast.Node
			if ast.IsArrowFunction(callback) {
				body := callback.Body()
				if !ast.IsBlock(body) {
					returnNode = body
				}
			}

			if returnNode == nil {
				return
			}

			// Don't check if return value should be skipped (comparisons, booleans, etc.)
			if shouldSkipConditionCheck(returnNode) {
				return
			}

			returnType := ctx.TypeChecker.GetTypeAtLocation(returnNode)
			truthiness := checkTypeIsTruthy(returnType)

			if truthiness == TruthinessTruthy {
				ctx.ReportNode(returnNode, buildAlwaysTruthyFuncMessage())
			} else if truthiness == TruthinessFalsy {
				ctx.ReportNode(returnNode, buildAlwaysFalsyFuncMessage())
			}
		}

		return rule.RuleListeners{
			// Check conditions in if statements
			ast.KindIfStatement: func(node *ast.Node) {
				expr := node.AsIfStatement().Expression
				if !shouldSkipConditionCheck(expr) {
					checkNode(expr, true)
				}
			},

			// Check ternary conditions
			ast.KindConditionalExpression: func(node *ast.Node) {
				condition := node.AsConditionalExpression().Condition
				if !shouldSkipConditionCheck(condition) {
					checkNode(condition, true)
				}
			},

			// Check logical AND/OR
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				op := expr.OperatorToken.Kind

				if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
					// Only check the left operand for truthiness
					// Don't check comparison expressions or other binary expressions
					left := expr.Left
					if !shouldSkipConditionCheck(left) {
						checkNode(left, false)
					}
				} else if op == ast.KindQuestionQuestionToken {
					checkNullishCoalescing(node)
				} else {
					checkBinaryExpression(node)
				}
			},

			// Check prefix unary ! operator
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				expr := node.AsPrefixUnaryExpression()
				if expr.Operator == ast.KindExclamationToken {
					checkNode(expr.Operand, false)
				}
			},

			// Check while loop conditions
			ast.KindWhileStatement: func(node *ast.Node) {
				stmt := node.AsWhileStatement()
				checkLoopCondition(node, stmt.Expression)
			},

			// Check do-while loop conditions
			ast.KindDoStatement: func(node *ast.Node) {
				stmt := node.AsDoStatement()
				checkLoopCondition(node, stmt.Expression)
			},

			// Check for loop conditions
			ast.KindForStatement: func(node *ast.Node) {
				stmt := node.AsForStatement()
				checkLoopCondition(node, stmt.Condition)
			},

			// Check optional chaining
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkOptionalChain(node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkOptionalChain(node)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				checkOptionalChain(node)
				checkArrayPredicate(node)
			},
		}
	},
})
