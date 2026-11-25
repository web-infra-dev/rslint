package no_unnecessary_condition

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_unnecessary_condition.schema.json
var schemaJSON []byte

// Only null and undefined are always nullish. Void is intentionally excluded:
// it participates in possibly-nullish checks, but TypeScript permits functions
// returning other values to satisfy a void return type.
const alwaysNullishFlag = checker.TypeFlagsUndefined | checker.TypeFlagsNull

// Upstream treats void as possibly nullish because a void-returning expression
// can evaluate to undefined at runtime.
const possiblyNullishFlag = alwaysNullishFlag | checker.TypeFlagsVoid

// isNullishType checks if a type is strictly null or undefined.
func isNullishType(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, alwaysNullishFlag)
}

// isAlwaysNullish checks if ALL union constituents are null or undefined.
func isAlwaysNullish(t *checker.Type) bool {
	return utils.Every(utils.UnionTypeParts(t), func(part *checker.Type) bool {
		return isNullishType(part)
	})
}

// isPossiblyNullish checks if any union constituent is null, undefined, or void.
func isPossiblyNullish(t *checker.Type) bool {
	return utils.IsTypeFlagSetWithUnion(t, possiblyNullishFlag)
}

func isBoolOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindLessThanToken,
		ast.KindGreaterThanToken,
		ast.KindLessThanEqualsToken,
		ast.KindGreaterThanEqualsToken,
		ast.KindEqualsEqualsToken,
		ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken,
		ast.KindExclamationEqualsEqualsToken:
		return true
	default:
		return false
	}
}

type allowConstantLoopConditionsOption string

const (
	loopConditionNever              allowConstantLoopConditionsOption = "never"
	loopConditionAlways             allowConstantLoopConditionsOption = "always"
	loopConditionOnlyAllowedLiteral allowConstantLoopConditionsOption = "only-allowed-literals"
)

func normalizeAllowConstantLoopConditions(v interface{}) allowConstantLoopConditionsOption {
	switch val := v.(type) {
	case bool:
		if val {
			return loopConditionAlways
		}
		return loopConditionNever
	case string:
		switch val {
		case "always":
			return loopConditionAlways
		case "only-allowed-literals":
			return loopConditionOnlyAllowedLiteral
		default:
			return loopConditionNever
		}
	default:
		return loopConditionNever
	}
}

type ruleOptions struct {
	allowConstantLoopConditions                            allowConstantLoopConditionsOption
	allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing bool
	checkTypePredicates                                    bool
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{
		allowConstantLoopConditions: loopConditionNever,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if v, ok := optsMap["allowConstantLoopConditions"]; ok {
		opts.allowConstantLoopConditions = normalizeAllowConstantLoopConditions(v)
	}
	if v, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
		opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = v
	}
	if v, ok := optsMap["checkTypePredicates"].(bool); ok {
		opts.checkTypePredicates = v
	}
	return opts
}

// Messages
func buildAlwaysTruthyMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "alwaysTruthy", Description: "Unnecessary conditional, value is always truthy."}
}
func buildAlwaysFalsyMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "alwaysFalsy", Description: "Unnecessary conditional, value is always falsy."}
}
func buildAlwaysTruthyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "alwaysTruthyFunc", Description: "This callback should return a conditional, but return is always truthy."}
}
func buildAlwaysFalsyFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "alwaysFalsyFunc", Description: "This callback should return a conditional, but return is always falsy."}
}
func buildNeverMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "never", Description: "Unnecessary conditional, value is `never`."}
}
func buildNeverNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "neverNullish", Description: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined."}
}
func buildAlwaysNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "alwaysNullish", Description: "Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`."}
}
func buildNeverOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "neverOptionalChain", Description: "Unnecessary optional chain on a non-nullish value."}
}
func buildSuggestRemoveOptionalChainMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "suggestRemoveOptionalChain", Description: "Remove unnecessary optional chain"}
}
func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "noStrictNullCheck", Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly."}
}
func buildComparisonBetweenLiteralTypesMessage(left, operator, right, trueOrFalse string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "comparisonBetweenLiteralTypes",
		Description: fmt.Sprintf("Unnecessary conditional, comparison is always %s, since `%s %s %s` is %s.", trueOrFalse, left, operator, right, trueOrFalse),
	}
}
func buildNoOverlapBooleanExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "noOverlapBooleanExpression", Description: "Unnecessary conditional, the types have no overlap."}
}
func buildTypeGuardAlreadyIsTypeMessage(typeGuardOrAssertionFunction string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "typeGuardAlreadyIsType",
		Description: fmt.Sprintf("Unnecessary conditional, expression already has the type being checked by the %s.", typeGuardOrAssertionFunction),
	}
}

// Rule definition
var NoUnnecessaryConditionRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-condition",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		tc := ctx.TypeChecker

		compilerOptions := ctx.Program().Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)
		isNoUncheckedIndexedAccess := compilerOptions.NoUncheckedIndexedAccess.IsTrue()

		if !isStrictNullChecks && !opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		// --- Helper functions ---

		isArrayIndexExpression := func(node *ast.Node) bool {
			node = ast.SkipParentheses(node)
			if !ast.IsElementAccessExpression(node) {
				return false
			}
			elem := node.AsElementAccessExpression()
			property := ast.SkipParentheses(elem.ArgumentExpression)
			// ESTree represents template literals separately from literal keys.
			isDynamicIndex := !ast.IsLiteralExpression(property) || property.Kind == ast.KindNoSubstitutionTemplateLiteral
			objectType := utils.GetConstrainedTypeAtLocation(tc, elem.Expression)
			for _, part := range utils.UnionTypeParts(objectType) {
				if checker.Checker_isArrayType(tc, part) || (isDynamicIndex && checker.IsTupleType(part)) {
					return true
				}
			}
			return false
		}

		// Conditional is always necessary if it involves any, unknown, or a naked type variable
		isConditionalAlwaysNecessary := func(t *checker.Type) bool {
			return utils.Some(utils.UnionTypeParts(t), func(part *checker.Type) bool {
				return utils.IsTypeAnyType(part) ||
					utils.IsTypeUnknownType(part) ||
					utils.IsTypeFlagSet(part, checker.TypeFlagsTypeVariable|checker.TypeFlagsConditional)
			})
		}

		isNullableMemberExpression := func(node *ast.Node) bool {
			if !ast.IsAccessExpression(node) {
				return false
			}
			objectType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
			if ast.IsElementAccessExpression(node) {
				propertyType := ctx.TypeChecker.GetTypeAtLocation(node.AsElementAccessExpression().ArgumentExpression)
				return isNullablePropertyType(tc, objectType, propertyType, false)
			}
			// PropertyAccessExpression — use Checker_getAccessedPropertyName for
			// correct handling of private fields (#prop) and computed properties
			propName, ok := checker.Checker_getAccessedPropertyName(tc, node)
			if !ok {
				return false
			}
			propSymbol := checker.Checker_getPropertyOfType(tc, objectType, propName)
			if propSymbol == nil {
				// Fallback for private fields (#prop) — getPropertyOfType uses
				// the display name but private fields have internal escaped names.
				// GetSymbolAtLocation resolves private fields correctly.
				propSymbol = tc.GetSymbolAtLocation(node)
			}
			if propSymbol != nil && utils.IsSymbolFlagSet(propSymbol, ast.SymbolFlagsOptional) {
				return true
			}
			return false
		}

		// --- Core check functions ---

		var checkNode func(expression *ast.Node, isUnaryNotArgument bool, reportNode *ast.Node, check truthinessCheck)
		checkNode = func(expression *ast.Node, isUnaryNotArgument bool, reportNode *ast.Node, check truthinessCheck) {
			expression = ast.SkipParentheses(expression)

			if reportNode == nil {
				reportNode = expression
			}

			// Handle unary negation
			if ast.IsPrefixUnaryExpression(expression) && expression.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
				checkNode(expression.AsPrefixUnaryExpression().Operand, !isUnaryNotArgument, reportNode, checkBoth)
				return
			}

			// Skip array index expressions (unsound typing)
			if !isNoUncheckedIndexedAccess && isArrayIndexExpression(expression) {
				return
			}

			// The left operand has its own listener. In a value context, only
			// absorbing right operands (`a || true`, `a && false`) make the
			// whole expression constant. Preserve this restriction when nested.
			if ast.IsBinaryExpression(expression) {
				binExpr := expression.AsBinaryExpression()
				op := binExpr.OperatorToken.Kind
				if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
					if check == 0 && isOnlyUsedForTruthiness(expression, tc, opts.checkTypePredicates) {
						check = checkBoth
					}
					if check != checkBoth {
						absorbing := checkTruthy
						if op == ast.KindAmpersandAmpersandToken {
							absorbing = checkFalsy
						}
						if check == 0 {
							check = absorbing
						} else {
							check &= absorbing
						}
						if check == 0 {
							return
						}
					}
					checkNode(binExpr.Right, false, nil, check)
					return
				}
			}

			t := utils.GetConstrainedTypeAtLocation(tc, expression)
			if isConditionalAlwaysNecessary(t) {
				return
			}

			if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsNever) {
				ctx.ReportNode(reportNode, buildNeverMessage())
				return
			}
			if !utils.IsPossiblyTruthy(t) && check != checkTruthy {
				if !isUnaryNotArgument {
					ctx.ReportNode(reportNode, buildAlwaysFalsyMessage())
				} else {
					ctx.ReportNode(reportNode, buildAlwaysTruthyMessage())
				}
				return
			}
			if !utils.IsPossiblyFalsy(t) && check != checkFalsy {
				if !isUnaryNotArgument {
					ctx.ReportNode(reportNode, buildAlwaysTruthyMessage())
				} else {
					ctx.ReportNode(reportNode, buildAlwaysFalsyMessage())
				}
				return
			}
		}

		checkNodeForNullish := func(node *ast.Node, isAssignment bool) {
			// Match upstream: the rule operates on the ESTree node, which has no
			// parenthesized-expression wrapper, so `(x) ?? y` reports on `x`.
			// tsgo's AST keeps ParenthesizedExpression nodes, so skip them to
			// align the reported range with typescript-eslint (mirrors checkNode).
			node = ast.SkipParentheses(node)
			t := utils.GetConstrainedTypeAtLocation(tc, node)

			// Conditional is always necessary if it involves any, unknown, or type parameter
			if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsTypeVariable|checker.TypeFlagsConditional) {
				return
			}

			if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsNever) {
				ctx.ReportNode(node, buildNeverMessage())
				return
			}

			if !isPossiblyNullish(t) {
				// Skip unsound array index expressions before nullable-member analysis;
				// the latter performs additional property and index-signature queries.
				if !isNoUncheckedIndexedAccess && isArrayIndexExpression(node) {
					return
				}
				if ast.IsAccessExpression(node) && isNullableMemberExpression(node) {
					return
				}
				if !isNoUncheckedIndexedAccess &&
					isChainExpressionWithOptionalArrayIndex(node, isArrayIndexExpression) {
					return
				}
				ctx.ReportNode(node, buildNeverNullishMessage())
				return
			}

			if isAlwaysNullish(t) && (!isAssignment || !hasPossiblyNonNullishComputedMemberProperty(tc, node)) {
				ctx.ReportNode(node, buildAlwaysNullishMessage())
				return
			}
		}

		checkIfBoolExpressionIsNecessaryConditional := func(node *ast.Node, left *ast.Node, right *ast.Node, operatorToken ast.Kind) {
			leftType := utils.GetConstrainedTypeAtLocation(tc, left)
			rightType := utils.GetConstrainedTypeAtLocation(tc, right)
			operator := scanner.TokenToString(operatorToken)

			if conditionIsTrue, known := utils.CompareLiteralTypes(leftType, rightType, operatorToken); known {
				trueOrFalse := "false"
				if conditionIsTrue {
					trueOrFalse = "true"
				}
				ctx.ReportNode(node, buildComparisonBetweenLiteralTypesMessage(
					tc.TypeToString(leftType),
					operator,
					tc.TypeToString(rightType),
					trueOrFalse,
				))
				return
			}

			// Workaround for TypeScript issue #37160
			if isStrictNullChecks {
				isComparable := func(t *checker.Type, flag checker.TypeFlags) bool {
					flag |= checker.TypeFlagsAny | checker.TypeFlagsUnknown | checker.TypeFlagsTypeParameter | checker.TypeFlagsTypeVariable | checker.TypeFlagsConditional
					if operator == "==" || operator == "!=" {
						flag |= checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsVoid
					}
					return utils.IsTypeFlagSetWithUnion(t, flag)
				}

				leftFlags := checker.Type_flags(leftType)
				rightFlags := checker.Type_flags(rightType)

				if (leftFlags == checker.TypeFlagsUndefined && !isComparable(rightType, checker.TypeFlagsUndefined|checker.TypeFlagsVoid)) ||
					(rightFlags == checker.TypeFlagsUndefined && !isComparable(leftType, checker.TypeFlagsUndefined|checker.TypeFlagsVoid)) ||
					(leftFlags == checker.TypeFlagsNull && !isComparable(rightType, checker.TypeFlagsNull)) ||
					(rightFlags == checker.TypeFlagsNull && !isComparable(leftType, checker.TypeFlagsNull)) {
					ctx.ReportNode(node, buildNoOverlapBooleanExpressionMessage())
					return
				}
			}
		}

		checkLogicalExpressionForUnnecessaryConditionals := func(node *ast.Node) {
			binExpr := node.AsBinaryExpression()
			if binExpr.OperatorToken.Kind == ast.KindQuestionQuestionToken {
				checkNodeForNullish(binExpr.Left, false)
				return
			}
			// Only check the left side
			checkNode(binExpr.Left, false, nil, 0)
		}

		checkIfLoopIsNecessaryConditional := func(node *ast.Node) {
			var test *ast.Node
			switch node.Kind {
			case ast.KindWhileStatement:
				test = node.AsWhileStatement().Expression
			case ast.KindDoStatement:
				test = node.AsDoStatement().Expression
			case ast.KindForStatement:
				test = node.AsForStatement().Condition
			}
			if test == nil {
				return
			}
			test = ast.SkipParentheses(test)

			if opts.allowConstantLoopConditions == loopConditionOnlyAllowedLiteral {
				if test.Kind == ast.KindTrueKeyword || test.Kind == ast.KindFalseKeyword {
					return
				}
				if ast.IsNumericLiteral(test) {
					text := test.Text()
					if text == "0" || text == "1" {
						return
					}
				}
			}

			if opts.allowConstantLoopConditions == loopConditionAlways {
				testType := utils.GetConstrainedTypeAtLocation(tc, test)
				if utils.IsTrueLiteralType(testType) {
					return
				}
			}

			checkNode(test, false, nil, checkBoth)
		}

		// optionChainContainsOptionArrayIndex recursively walks the optional chain
		// looking for array/tuple index accesses. Array index accesses are unsound
		// without noUncheckedIndexedAccess, and unlike object index signatures,
		// they remain unsound even after function calls (the result type is still
		// from the array element type).
		var optionChainContainsOptionArrayIndex func(node *ast.Node) bool
		optionChainContainsOptionArrayIndex = func(node *ast.Node) bool {
			if ast.IsCallExpression(node) {
				callExpr := node.AsCallExpression()
				lhsNode := callExpr.Expression
				if ast.IsOptionalChain(node) && isArrayIndexExpression(lhsNode) {
					return true
				}
				if ast.IsAccessExpression(lhsNode) || ast.IsCallExpression(lhsNode) {
					return optionChainContainsOptionArrayIndex(lhsNode)
				}
			} else if ast.IsAccessExpression(node) {
				obj := node.Expression()
				if ast.IsOptionalChain(node) && isArrayIndexExpression(obj) {
					return true
				}
				if ast.IsAccessExpression(obj) || ast.IsCallExpression(obj) {
					return optionChainContainsOptionArrayIndex(obj)
				}
			}
			return false
		}

		isMemberExpressionNullableOriginFromObject := func(node *ast.Node) bool {
			isComputed := ast.IsElementAccessExpression(node)
			// Match upstream's `isIdentifier(node.property)` guard: a computed
			// access is only attributed to its object when the key is an
			// identifier (`a[k]`). For literal keys (`a[0]`, `a['x']`) the element
			// type, e.g. `T | undefined` under noUncheckedIndexedAccess, is the
			// access's own nullability and must not be blamed on the object.
			// ESTree removes parentheses around identifiers.
			if isComputed && !ast.IsIdentifier(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)) {
				return false
			}
			prevType := utils.GetConstrainedTypeAtLocation(tc, node.Expression())
			if !prevType.IsUnion() {
				return false
			}
			var propertyType *checker.Type
			if isComputed {
				propertyType = utils.GetConstrainedTypeAtLocation(tc, node.AsElementAccessExpression().ArgumentExpression)
			}
			propName := ""
			if !isComputed {
				// Use Checker_getAccessedPropertyName for correct handling of
				// private fields (#prop) and computed properties
				var ok bool
				propName, ok = checker.Checker_getAccessedPropertyName(tc, node)
				if !ok {
					return false
				}
			}

			isOwnNullable := utils.Some(prevType.Types(), func(partType *checker.Type) bool {
				if isComputed {
					return isNullablePropertyType(tc, partType, propertyType, isNoUncheckedIndexedAccess)
				}
				propType := tc.GetTypeOfPropertyOfType(partType, propName)
				if propType != nil {
					return utils.IsNullableType(propType)
				}
				indexInfos := tc.GetIndexInfosOfType(partType)
				return utils.Some(indexInfos, func(info *checker.IndexInfo) bool {
					keyTypeName := utils.GetTypeName(tc, info.KeyType())
					isStringOrNumber := keyTypeName == "string" || keyTypeName == "number"
					return isStringOrNumber && (isNoUncheckedIndexedAccess || utils.IsNullableType(info.ValueType()))
				})
			})
			return !isOwnNullable && utils.IsNullableType(prevType)
		}

		isCallExpressionNullableOriginFromCallee := func(node *ast.Node) bool {
			callExpr := node.AsCallExpression()
			prevType := utils.GetConstrainedTypeAtLocation(tc, callExpr.Expression)
			if !prevType.IsUnion() {
				return false
			}
			isOwnNullable := utils.Some(prevType.Types(), func(partType *checker.Type) bool {
				signatures := utils.GetCallSignatures(tc, partType)
				return utils.Some(signatures, func(sig *checker.Signature) bool {
					return utils.IsNullableType(checker.Checker_getReturnTypeOfSignature(tc, sig))
				})
			})
			return !isOwnNullable && utils.IsNullableType(prevType)
		}

		isOptionableExpression := func(node *ast.Node) bool {
			t := utils.GetConstrainedTypeAtLocation(tc, node)
			if isConditionalAlwaysNecessary(t) {
				return true
			}
			if !utils.IsNullableType(t) {
				return false
			}
			if ast.IsAccessExpression(node) {
				return !isMemberExpressionNullableOriginFromObject(node)
			}
			if ast.IsCallExpression(node) {
				return !isCallExpressionNullableOriginFromCallee(node)
			}
			return true
		}

		checkOptionalChain := func(node *ast.Node, fix string) {
			if !ast.IsOptionalChain(node) {
				return
			}

			// Skip unsound array/tuple index expressions
			if !isNoUncheckedIndexedAccess && optionChainContainsOptionArrayIndex(node) {
				return
			}

			var nodeToCheck *ast.Node
			if ast.IsCallExpression(node) {
				nodeToCheck = node.AsCallExpression().Expression
			} else {
				nodeToCheck = node.Expression()
			}
			if isOptionableExpression(nodeToCheck) {
				return
			}

			// Get the ?. token range directly from the AST node
			questionDotToken := node.QuestionDotToken()
			if questionDotToken == nil {
				return
			}
			questionDotRange := utils.TrimNodeTextRange(ctx.SourceFile, questionDotToken)

			ctx.ReportRangeWithSuggestions(
				questionDotRange,
				buildNeverOptionalChainMessage(),
				rule.RuleSuggestion{
					Message: buildSuggestRemoveOptionalChainMessage(),
					FixesArr: []rule.RuleFix{
						rule.RuleFixReplaceRange(questionDotRange, fix),
					},
				},
			)
		}

		checkCallExpression := func(node *ast.Node) {
			callExpr := node.AsCallExpression()

			if opts.checkTypePredicates {
				// Check for truthiness assertion functions
				truthinessArg := typescriptutil.FindTruthinessAssertedArgument(tc, callExpr)
				if truthinessArg != nil {
					checkNode(truthinessArg, false, nil, checkBoth)
				}

				// Check for type guard assertion functions
				typeGuardResult := findTypeGuardAssertedArgument(tc, callExpr)
				if typeGuardResult != nil {
					typeOfArgument := utils.GetConstrainedTypeAtLocation(tc, typeGuardResult.argument)
					predType := typeGuardResult.predicateType
					// Match original: skip any/unknown, check mutual assignability or union predicate
					if !utils.IsTypeAnyType(typeOfArgument) && !utils.IsTypeUnknownType(typeOfArgument) &&
						checker.Checker_isTypeAssignableTo(tc, typeOfArgument, predType) &&
						(checker.Checker_isTypeAssignableTo(tc, predType, typeOfArgument) || predType.IsUnion()) {
						label := "type guard"
						if typeGuardResult.asserts {
							label = "assertion function"
						}
						ctx.ReportNode(typeGuardResult.argument, buildTypeGuardAlreadyIsTypeMessage(label))
					}
				}
			}

			// Check array method calls with predicates
			if utils.IsArrayMethodCallWithPredicate(tc, callExpr) && len(callExpr.Arguments.Nodes) > 0 {
				callback := ast.SkipParentheses(callExpr.Arguments.Nodes[0])

				if ast.IsArrowFunction(callback) || ast.IsFunctionExpression(callback) {
					body := callback.Body()
					if body != nil && !ast.IsBlock(body) {
						// Arrow function with expression body: () => something
						checkNode(body, false, nil, checkBoth)
						return
					}
					if body != nil && ast.IsBlock(body) {
						statements := body.AsBlock().Statements.Nodes
						if len(statements) == 1 && ast.IsReturnStatement(statements[0]) {
							returnStmt := statements[0].AsReturnStatement()
							if returnStmt.Expression != nil {
								checkNode(returnStmt.Expression, false, nil, checkBoth)
								return
							}
						}
					}
				}

				// Check callback return types
				callbackType := utils.GetConstrainedTypeAtLocation(tc, callback)
				returnTypes := collectReturnTypes(tc, callbackType)
				if len(returnTypes) == 0 {
					return
				}

				hasFalsyReturnTypes := false
				hasTruthyReturnTypes := false
				for _, rt := range returnTypes {
					constraintType, _ := utils.GetConstraintInfo(tc, rt)
					if constraintType == nil || utils.IsTypeAnyType(constraintType) || utils.IsTypeUnknownType(constraintType) {
						return
					}
					if utils.IsPossiblyFalsy(constraintType) {
						hasFalsyReturnTypes = true
					}
					if utils.IsPossiblyTruthy(constraintType) {
						hasTruthyReturnTypes = true
					}
					if hasFalsyReturnTypes && hasTruthyReturnTypes {
						return
					}
				}

				if !hasFalsyReturnTypes {
					ctx.ReportNode(callback, buildAlwaysTruthyFuncMessage())
					return
				}
				if !hasTruthyReturnTypes {
					ctx.ReportNode(callback, buildAlwaysFalsyFuncMessage())
					return
				}
			}
		}

		checkAssignmentExpression := func(node *ast.Node) {
			assignExpr := node.AsBinaryExpression()
			switch assignExpr.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken:
				checkNode(assignExpr.Left, false, nil, checkBoth)
			case ast.KindQuestionQuestionEqualsToken:
				checkNodeForNullish(assignExpr.Left, true)
			}
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				checkNode(node.AsIfStatement().Expression, false, nil, checkBoth)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkNode(node.AsConditionalExpression().Condition, false, nil, checkBoth)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				checkIfLoopIsNecessaryConditional(node)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				checkIfLoopIsNecessaryConditional(node)
			},
			ast.KindForStatement: func(node *ast.Node) {
				checkIfLoopIsNecessaryConditional(node)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				op := binExpr.OperatorToken.Kind

				// Logical expressions: &&, ||, ??
				if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken || op == ast.KindQuestionQuestionToken {
					checkLogicalExpressionForUnnecessaryConditionals(node)
					return
				}

				// Assignment expressions: &&=, ||=, ??=
				if op == ast.KindAmpersandAmpersandEqualsToken || op == ast.KindBarBarEqualsToken || op == ast.KindQuestionQuestionEqualsToken {
					checkAssignmentExpression(node)
					return
				}

				// Boolean comparison operators
				if isBoolOperator(op) {
					checkIfBoolExpressionIsNecessaryConditional(node, binExpr.Left, binExpr.Right, op)
					return
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				checkCallExpression(node)

				// Check optional call chain
				callExpr := node.AsCallExpression()
				if callExpr.QuestionDotToken != nil {
					checkOptionalChain(node, "")
				}
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				propAccess := node.AsPropertyAccessExpression()
				if propAccess.QuestionDotToken != nil {
					checkOptionalChain(node, ".")
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				elemAccess := node.AsElementAccessExpression()
				if elemAccess.QuestionDotToken != nil {
					checkOptionalChain(node, "")
				}
			},
			ast.KindCaseClause: func(node *ast.Node) {
				caseClause := node.AsCaseOrDefaultClause()
				if caseClause == nil || caseClause.Expression == nil {
					return
				}
				// Navigate up from CaseClause -> CaseBlock -> SwitchStatement
				caseBlock := node.Parent
				if caseBlock == nil {
					return
				}
				switchNode := caseBlock.Parent
				if switchNode == nil || !ast.IsSwitchStatement(switchNode) {
					return
				}
				switchStmt := switchNode.AsSwitchStatement()
				checkIfBoolExpressionIsNecessaryConditional(
					caseClause.Expression,
					switchStmt.Expression,
					caseClause.Expression,
					ast.KindEqualsEqualsEqualsToken,
				)
			},
		}
	},
})

// Helper functions

// Zero asks a logical expression to infer whether its value is consumed.
// Other values restrict which constant right operands can be reported.
type truthinessCheck uint8

const (
	checkTruthy truthinessCheck = 1 << iota
	checkFalsy
	checkBoth = checkTruthy | checkFalsy
)

func hasPossiblyNonNullishComputedMemberProperty(tc *checker.Checker, node *ast.Node) bool {
	if !ast.IsElementAccessExpression(node) || ast.IsOptionalChain(node) {
		return false
	}
	element := node.AsElementAccessExpression()
	objectType := utils.GetConstrainedTypeAtLocation(tc, element.Expression)
	propertyType := utils.GetConstrainedTypeAtLocation(tc, element.ArgumentExpression)
	// Writes through a union key use the intersection of the property types.
	// A read or a single-key write still has meaningful flow narrowing.
	if !propertyType.IsUnion() {
		return false
	}
	possiblyNonNullish := false
	for _, keyType := range utils.UnionTypeParts(propertyType) {
		if !checker.IsTypeUsableAsPropertyName(keyType) {
			return false
		}
		selectedType := tc.GetTypeOfPropertyOfType(objectType, checker.GetPropertyNameFromType(keyType))
		if selectedType == nil {
			return false
		}
		possiblyNonNullish = possiblyNonNullish || !utils.Every(utils.UnionTypeParts(selectedType), func(part *checker.Type) bool {
			return utils.IsTypeFlagSet(part, possiblyNullishFlag)
		})
	}
	return possiblyNonNullish
}

func isOnlyUsedForTruthiness(node *ast.Node, tc *checker.Checker, checkTypePredicates bool) bool {
	node = ast.SkipParentheses(node)
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	var callback *ast.Node
	switch parent.Kind {
	case ast.KindIfStatement:
		return ast.SkipParentheses(parent.AsIfStatement().Expression) == node
	case ast.KindConditionalExpression:
		return ast.SkipParentheses(parent.AsConditionalExpression().Condition) == node
	case ast.KindWhileStatement:
		return ast.SkipParentheses(parent.AsWhileStatement().Expression) == node
	case ast.KindDoStatement:
		return ast.SkipParentheses(parent.AsDoStatement().Expression) == node
	case ast.KindForStatement:
		condition := parent.AsForStatement().Condition
		return condition != nil && ast.SkipParentheses(condition) == node
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			return ast.SkipParentheses(binary.Left) == node || isOnlyUsedForTruthiness(parent, tc, checkTypePredicates)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return isOnlyUsedForTruthiness(parent, tc, checkTypePredicates)
		}
	case ast.KindCallExpression:
		if checkTypePredicates {
			argument := typescriptutil.FindTruthinessAssertedArgument(tc, parent.AsCallExpression())
			return argument != nil && ast.SkipParentheses(argument) == node
		}
	case ast.KindArrowFunction:
		callback = parent
	case ast.KindReturnStatement:
		callback = ast.GetContainingFunction(parent)
	}
	if callback != nil && (ast.IsArrowFunction(callback) || ast.IsFunctionExpression(callback)) {
		call := ast.WalkUpParenthesizedExpressions(callback.Parent)
		if ast.IsCallExpression(call) {
			expression := call.AsCallExpression()
			return len(expression.Arguments.Nodes) > 0 &&
				ast.SkipParentheses(expression.Arguments.Nodes[0]) == callback &&
				utils.IsArrayMethodCallWithPredicate(tc, expression)
		}
	}
	return false
}

func isNullablePropertyType(tc *checker.Checker, objType *checker.Type, propertyType *checker.Type, includeUncheckedIndex bool) bool {
	if propertyType.IsUnion() {
		return utils.Some(propertyType.Types(), func(part *checker.Type) bool {
			return isNullablePropertyType(tc, objType, part, includeUncheckedIndex)
		})
	}
	if checker.IsTypeUsableAsPropertyName(propertyType) {
		propName := checker.GetPropertyNameFromType(propertyType)
		propType := tc.GetTypeOfPropertyOfType(objType, propName)
		if propType != nil {
			return utils.IsNullableType(propType)
		}
	}
	// Let the checker handle numeric strings, template keys, symbols, and
	// overlapping signatures instead of matching printed key type names.
	info := checker.Checker_getApplicableIndexInfo(tc, objType, propertyType)
	return info != nil && (includeUncheckedIndex || utils.IsNullableType(info.ValueType()) ||
		// Preserve upstream's conservative treatment of matching index types.
		utils.GetTypeName(tc, info.KeyType()) == utils.GetTypeName(tc, propertyType))
}

func collectReturnTypes(tc *checker.Checker, callbackType *checker.Type) []*checker.Type {
	sigs := utils.CollectAllCallSignatures(tc, callbackType)
	result := make([]*checker.Type, 0, len(sigs))
	for _, sig := range sigs {
		result = append(result, checker.Checker_getReturnTypeOfSignature(tc, sig))
	}
	return result
}

type typeGuardResult struct {
	argument      *ast.Node
	predicateType *checker.Type
	asserts       bool
}

// firstSpreadIndex / findTruthinessAssertedArgument were moved to
// typescriptutil because strict_boolean_expressions needs the same predicate
// mapping. Callers should use typescriptutil.FirstSpreadIndex /
// typescriptutil.FindTruthinessAssertedArgument directly.

func findTypeGuardAssertedArgument(tc *checker.Checker, callExpr *ast.CallExpression) *typeGuardResult {
	sig := checker.Checker_getResolvedSignature(tc, callExpr.AsNode(), nil, checker.CheckModeNormal)
	if sig == nil {
		return nil
	}

	predicate := tc.GetTypePredicateOfSignature(sig)
	if predicate == nil {
		return nil
	}

	// Type guard: param is Type / asserts param is Type
	if predicate.Type() == nil {
		return nil
	}

	if predicate.Kind() != checker.TypePredicateKindIdentifier && predicate.Kind() != checker.TypePredicateKindAssertsIdentifier {
		return nil
	}

	paramIndex := predicate.ParameterIndex()
	// Skip if parameter index is at or past a spread element (unreliable mapping)
	spreadIdx := typescriptutil.FirstSpreadIndex(callExpr)
	if spreadIdx >= 0 && int(paramIndex) >= spreadIdx {
		return nil
	}
	if int(paramIndex) >= len(callExpr.Arguments.Nodes) {
		return nil
	}

	return &typeGuardResult{
		argument:      callExpr.Arguments.Nodes[paramIndex],
		predicateType: predicate.Type(),
		asserts:       predicate.Kind() == checker.TypePredicateKindAssertsIdentifier,
	}
}

func isChainExpressionWithOptionalArrayIndex(node *ast.Node, isArrayIndexExpr func(n *ast.Node) bool) bool {
	if !ast.IsAccessExpression(node) && !ast.IsCallExpression(node) {
		return false
	}
	// Walk the chain, recursing through both access and call expressions
	current := node
	for current != nil {
		if ast.IsAccessExpression(current) {
			obj := current.Expression()
			if ast.IsOptionalChain(current) && isArrayIndexExpr(obj) {
				return true
			}
			current = obj
		} else if ast.IsCallExpression(current) {
			callExpr := current.AsCallExpression()
			if ast.IsOptionalChain(current) && isArrayIndexExpr(callExpr.Expression) {
				return true
			}
			current = callExpr.Expression
		} else {
			break
		}
	}
	return false
}
