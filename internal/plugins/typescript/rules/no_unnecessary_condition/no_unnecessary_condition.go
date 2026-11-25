package no_unnecessary_condition

import (
	_ "embed"
	"fmt"
	"slices"

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
		checkTypePredicates := opts.checkTypePredicates
		tc := ctx.TypeChecker
		conditionTypes := conditionTypeChecker{context: &ctx}

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

		isPossiblyMissingRead := func(node *ast.Node) bool {
			node = ast.SkipParentheses(node)
			if !isStrictNullChecks && ast.IsIdentifier(node) {
				if symbol := tc.GetSymbolAtLocation(node); symbol != nil && symbol.ValueDeclaration != nil {
					declaration := symbol.ValueDeclaration
					if ast.IsParameterDeclaration(declaration) && declaration.AsParameterDeclaration().QuestionToken != nil {
						return !hasKnownNonNullishAssignment(tc, node, conditionTypes.staticEvaluator())
					}
				}
			}
			if !ast.IsAccessExpression(node) {
				return false
			}
			var objectType *checker.Type
			if conditionTypes.hasAccessorRead(node.Expression()) {
				objectType = tc.GetNonNullableType(conditionTypes.getConstrainedConditionType(node.Expression()))
			}
			declaredType, propertyFlags := conditionTypes.getDeclaredMemberReadType(node, objectType)
			// Optional symbols retain information erased from types when
			// strictNullChecks is disabled.
			if !isStrictNullChecks && propertyFlags&ast.SymbolFlagsOptional != 0 {
				if objectType != nil || propertyFlags&ast.SymbolFlagsAccessor != 0 {
					return true
				}
				// Non-strict types erase undefined from assignment sources too.
				// Require a statically known value at the bound assignment point.
				return !hasKnownNonNullishAssignment(tc, node, conditionTypes.staticEvaluator())
			}
			if !ast.IsElementAccessExpression(node) {
				return false
			}
			// A nullable declared read type and a non-nullish flow type mean
			// the access was narrowed. Do not widen it back to its declaration.
			if declaredType == nil || isPossiblyNullish(declaredType) {
				return false
			}
			element := node.AsElementAccessExpression()
			if objectType == nil {
				objectType = utils.GetConstrainedTypeAtLocation(tc, element.Expression)
			}
			propertyType := utils.GetConstrainedTypeAtLocation(tc, element.ArgumentExpression)
			return isNullablePropertyType(tc, objectType, propertyType, false)
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
					if check == 0 && isOnlyUsedForTruthiness(expression, tc, checkTypePredicates) {
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

			t := conditionTypes.getConstrainedConditionType(expression)
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

		checkNodeForNullish := func(node *ast.Node) {
			// Match upstream: the rule operates on the ESTree node, which has no
			// parenthesized-expression wrapper, so `(x) ?? y` reports on `x`.
			// tsgo's AST keeps ParenthesizedExpression nodes, so skip them to
			// align the reported range with typescript-eslint (mirrors checkNode).
			node = ast.SkipParentheses(node)
			t := conditionTypes.getConstrainedConditionType(node)

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
				if (!isNoUncheckedIndexedAccess || !isStrictNullChecks) && isPossiblyMissingRead(node) {
					return
				}
				if !isNoUncheckedIndexedAccess &&
					isChainExpressionWithOptionalArrayIndex(node, isArrayIndexExpression) {
					return
				}
				ctx.ReportNode(node, buildNeverNullishMessage())
				return
			}

			if isAlwaysNullish(t) {
				ctx.ReportNode(node, buildAlwaysNullishMessage())
				return
			}
		}

		checkIfBoolExpressionIsNecessaryConditional := func(node *ast.Node, left *ast.Node, right *ast.Node, operatorToken ast.Kind) {
			leftType := conditionTypes.getConstrainedConditionType(left)
			rightType := conditionTypes.getConstrainedConditionType(right)
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
				checkNodeForNullish(binExpr.Left)
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
				testType := conditionTypes.getConstrainedConditionType(test)
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
			prevType := conditionTypes.getConstrainedConditionType(node.Expression())
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
			prevType := conditionTypes.getConstrainedConditionType(callExpr.Expression)
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
			t := conditionTypes.getConstrainedConditionType(node)
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

			if checkTypePredicates {
				// Check for truthiness assertion functions
				truthinessArg := typescriptutil.FindTruthinessAssertedArgument(tc, callExpr)
				if truthinessArg != nil {
					checkNode(truthinessArg, false, nil, checkBoth)
				}

				// Check for type guard assertion functions
				typeGuardResult := findTypeGuardAssertedArgument(tc, callExpr)
				if typeGuardResult != nil {
					typeOfArgument := conditionTypes.getConstrainedConditionType(typeGuardResult.argument)
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

				if (ast.IsArrowFunction(callback) || ast.IsFunctionExpression(callback)) && ast.GetFunctionFlags(callback) == ast.FunctionFlagsNormal {
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
				callbackType := conditionTypes.getConstrainedConditionType(callback)
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
				checkNodeForNullish(assignExpr.Left)
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

// Conditions test read values. Accessors can produce a different value on each
// read, so previous reads and setter arguments do not narrow the getter's type.
type conditionTypeChecker struct {
	context           *rule.RuleContext
	evaluator         *utils.StaticStringEvaluator
	mutationEvaluator *utils.StaticStringEvaluator
	initializers      map[*ast.Symbol]initializerOrigin
	dataCopyLiterals  map[*ast.Node]bool
	capturedReadTypes map[*ast.Symbol]*checker.Type
}

func (conditionTypes *conditionTypeChecker) staticEvaluator() *utils.StaticStringEvaluator {
	if conditionTypes.evaluator == nil {
		ctx := conditionTypes.context
		var references utils.StaticReferenceResolver
		if ctx.Refs != nil {
			references = ctx.Refs
		}
		conditionTypes.evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, references)
	}
	return conditionTypes.evaluator
}

func (conditionTypes *conditionTypeChecker) getConstrainedConditionType(node *ast.Node) *checker.Type {
	return conditionTypes.getConditionType(node, nil)
}

// Carry the nearest accessor while descending its receiver chain, so each
// recursive step does not scan the same prefix again.
func (conditionTypes *conditionTypeChecker) getConditionType(node, accessor *ast.Node) *checker.Type {
	tc := conditionTypes.context.TypeChecker
	node = ast.SkipParentheses(node)
	if ast.IsSatisfiesExpression(node) {
		return conditionTypes.getConditionType(node.Expression(), accessor)
	}
	if ast.IsNonNullExpression(node) {
		expressionType := conditionTypes.getConditionType(node.Expression(), accessor)
		if ast.IsOptionalChain(node) {
			// The assertion excludes a nullish value on the reached path,
			// but it does not cancel an earlier optional-chain short circuit.
			nonOptionalType := checker.Checker_getOptionalExpressionType(tc, expressionType, node.Expression())
			return checker.Checker_propagateOptionalTypeMarker(tc, tc.GetNonNullableType(nonOptionalType), node, expressionType != nonOptionalType)
		}
		return tc.GetNonNullableType(expressionType)
	}
	// Resolve the expression first so the checker supplies its selected property
	// symbol. Ordinary reads can then keep their existing flow type directly.
	originalType := utils.GetConstrainedTypeAtLocation(tc, node)
	if ast.IsIdentifier(node) {
		if captured := conditionTypes.getCapturedReadType(node); captured != nil {
			// Start flow analysis from unknown so the stale initializer cannot
			// narrow this read again. Intersect its guards with the captured type.
			guardType := getConditionFlowType(tc, node, tc.GetUnknownType())
			return checker.Checker_getIntersectionType(tc, []*checker.Type{captured, guardType})
		}
	}
	flowType := originalType
	var expressionType *checker.Type
	needsOptionalPropagation := false
	if ast.IsAccessExpression(node) {
		target := ast.GetAssignmentTarget(node)
		isLogicalAssignment := target != nil && ast.IsBinaryExpression(target) && ast.IsLogicalOrCoalescingAssignmentOperator(target.AsBinaryExpression().OperatorToken.Kind)
		needsReadType := isLogicalAssignment
		if ast.IsElementAccessExpression(node) {
			key := node.AsElementAccessExpression().ArgumentExpression
			needsReadType = needsReadType || ast.SkipParentheses(key) != key
		}
		if accessor == nil || accessor == node {
			accessor = conditionTypes.findAccessorRead(node.Expression())
		}
		varyingReceiver := accessor != nil && !ast.IsIdentifier(accessor)
		var readType *checker.Type
		var propertyFlags ast.SymbolFlags
		if accessor != nil {
			expressionType = conditionTypes.getConditionType(node.Expression(), accessor)
			objectType := tc.GetNonNullableType(expressionType)
			readType, propertyFlags = conditionTypes.getDeclaredMemberReadType(node, objectType)
			if readType == nil {
				readType = tc.GetUnknownType()
			}
		} else if symbol := checker.Checker_getResolvedSymbolOrNil(tc, node); symbol != nil {
			readSymbol, readContract, flags := conditionTypes.memberReadSymbol(node, symbol, nil)
			propertyFlags = flags
			if needsReadType || propertyFlags&ast.SymbolFlagsAccessor != 0 || symbol.Flags&ast.SymbolFlagsAccessor != 0 || symbolHasAccessor(symbol) {
				readType = getMemberReadType(tc, symbol, readSymbol, readContract)
			}
		} else if needsReadType {
			readType, propertyFlags = conditionTypes.getDeclaredMemberReadType(node, nil)
		} else if hasUnionOrConstrainedKey(tc, node) {
			readType, propertyFlags = conditionTypes.getDeclaredMemberReadType(node, nil)
		} else if origin := conditionTypes.receiverObjectOrigin(node); origin.node != nil && !origin.dataCopy {
			readType, propertyFlags = conditionTypes.getDeclaredMemberReadType(node, nil)
		}
		isAccessor := propertyFlags&ast.SymbolFlagsAccessor != 0
		needsOptionalPropagation = accessor != nil || readType != nil
		if readType != nil {
			if constraint := checker.Checker_getBaseConstraintOfType(tc, readType); constraint != nil {
				readType = constraint
			}
			if !isAccessor && !varyingReceiver {
				readType = getConditionFlowType(tc, node, readType)
			}
			flowType = readType
		}
	}
	if ast.IsCallExpression(node) && accessor == nil {
		accessor = conditionTypes.findAccessorRead(node.Expression())
	}
	if ast.IsCallExpression(node) && accessor != nil {
		needsOptionalPropagation = true
		expressionType = conditionTypes.getConditionType(node.Expression(), accessor)
		flowCalleeType := utils.GetConstrainedTypeAtLocation(tc, node.Expression())
		// Resolve the actual arguments against the getter's call signatures.
		// The checker handles overload selection and generic instantiation.
		readCalleeType := tc.GetNonNullableType(expressionType)
		flowCalleeType = tc.GetNonNullableType(flowCalleeType)
		// Unchanged unions of single signatures already have the checker's
		// argument and generic inference, including functions of unequal arity.
		hasOverloads := readCalleeType.IsUnion() && slices.ContainsFunc(readCalleeType.Types(), func(part *checker.Type) bool {
			return len(checker.Checker_getSignaturesOfType(tc, part, checker.SignatureKindCall)) > 1
		})
		if readCalleeType != flowCalleeType || hasOverloads {
			calleeType := checker.Checker_getApparentType(tc, tc.GetNonNullableType(expressionType))
			flowType = getReadCallType(tc, node, calleeType)
			if constraint := checker.Checker_getBaseConstraintOfType(tc, flowType); constraint != nil {
				flowType = constraint
			}
		}
	}
	if needsOptionalPropagation && ast.IsOptionalChain(node) {
		// The short circuit also reaches ordinary properties and calls after
		// an accessor, even when their cached flow type lost that possibility.
		if expressionType == nil {
			expressionType = conditionTypes.getConditionType(node.Expression(), accessor)
		}
		if flowType != originalType || expressionType != utils.GetConstrainedTypeAtLocation(tc, node.Expression()) {
			nonOptionalType := checker.Checker_getOptionalExpressionType(tc, expressionType, node.Expression())
			flowType = checker.Checker_propagateOptionalTypeMarker(tc, flowType, node, expressionType != nonOptionalType)
		}
	}
	return flowType
}

func getReadCallType(tc *checker.Checker, node *ast.Node, calleeType *checker.Type) *checker.Type {
	// Let the checker reject unions without compatible call signatures,
	// including generic functions with different type-parameter constraints.
	if len(checker.Checker_getSignaturesOfType(tc, calleeType, checker.SignatureKindCall)) == 0 {
		return tc.GetUnknownType()
	}
	var returnTypes []*checker.Type
	for _, part := range utils.UnionTypeParts(calleeType) {
		// Each possible function selects its own overload and generic arguments.
		// Assignability does not prove that two overloaded calls return alike.
		signatures := checker.Checker_getSignaturesOfType(tc, part, checker.SignatureKindCall)
		if len(signatures) == 0 {
			return tc.GetUnknownType()
		}
		var factory ast.NodeFactory
		call := node.AsCallExpression().Clone(&factory)
		call.Parent = node.Parent
		// Keep speculative signatures and errors off the parsed call.
		var candidates []*checker.Signature
		signature := checker.Checker_resolveCall(tc, call, signatures, &candidates, checker.CheckModeNormal, checker.SignatureFlagsNone, nil)
		arguments := checker.Checker_getEffectiveCallArguments(tc, call)
		if len(signatures) == 1 && !signature.HasRestParameter() && len(arguments) > len(signature.Parameters()) {
			// A single-signature union member ignores surplus arguments.
			// Keep checking the arguments it accepts and its receiver normally.
			arguments = arguments[:len(signature.Parameters())]
		}
		if !checker.Checker_hasCorrectArity(tc, call, arguments, signature, false) ||
			!checker.Checker_isSignatureApplicable(tc, call, arguments, signature, checker.Checker_assignableRelation(tc), checker.CheckModeNormal, false, nil) {
			return tc.GetUnknownType()
		}
		returnTypes = append(returnTypes, checker.Checker_getReturnTypeOfSignature(tc, signature))
	}
	return tc.GetUnionType(returnTypes)
}

func (conditionTypes *conditionTypeChecker) getDeclaredMemberReadType(node *ast.Node, objectType *checker.Type) (*checker.Type, ast.SymbolFlags) {
	tc := conditionTypes.context.TypeChecker
	if symbol := checker.Checker_getResolvedSymbolOrNil(tc, node); symbol != nil {
		if objectType != nil {
			symbol = checker.Checker_getPropertyOfType(tc, objectType, symbol.Name)
		}
		if symbol != nil {
			readSymbol, readContract, flags := conditionTypes.memberReadSymbol(node, symbol, objectType)
			return getMemberReadType(tc, symbol, readSymbol, readContract), flags
		}
	}
	var keyType *checker.Type
	if ast.IsElementAccessExpression(node) {
		key := node.AsElementAccessExpression().ArgumentExpression
		keyType = utils.GetConstrainedTypeAtLocation(tc, key)
		if checker.Checker_isForInVariableForNumericPropertyNames(tc, key) {
			keyType = checker.Checker_numberType(tc)
		}
	} else {
		// Symbol lookup also resolves private fields and inherited accessors.
		if symbol := tc.GetSymbolAtLocation(node); symbol != nil {
			if objectType != nil {
				symbol = checker.Checker_getPropertyOfType(tc, objectType, symbol.Name)
			}
			if symbol != nil {
				readSymbol, readContract, flags := conditionTypes.memberReadSymbol(node, symbol, objectType)
				return getMemberReadType(tc, symbol, readSymbol, readContract), flags
			}
		}
		name, ok := checker.Checker_getAccessedPropertyName(tc, node)
		if !ok {
			return nil, ast.SymbolFlagsNone
		}
		keyType = checker.Checker_getStringLiteralType(tc, name)
	}
	if objectType == nil {
		objectType = tc.GetNonNullableType(utils.GetConstrainedTypeAtLocation(tc, node.Expression()))
	}
	var propertyFlags ast.SymbolFlags
	var readTypes []*checker.Type
	allPropertiesKnown := true
	hasReadOverride := false
	for _, key := range utils.UnionTypeParts(keyType) {
		if checker.IsTypeUsableAsPropertyName(key) {
			if symbol := checker.Checker_getPropertyOfType(tc, objectType, checker.GetPropertyNameFromType(key)); symbol != nil {
				readSymbol, readContract, flags := conditionTypes.memberReadSymbol(node, symbol, objectType)
				propertyFlags |= flags
				hasReadOverride = hasReadOverride || readSymbol != symbol || readContract != nil
				readTypes = append(readTypes, getMemberReadType(tc, symbol, readSymbol, readContract))
				continue
			}
		}
		allPropertiesKnown = false
	}
	if allPropertiesKnown && hasReadOverride {
		// Object.assign can intersect away an existing target getter's
		// return type. Preserve its actual read type for union keys too.
		return tc.GetUnionType(readTypes), propertyFlags
	}
	// With no access node, this query computes a read type without using the
	// assignment target's write context or changing its cached resolution.
	readType := checker.Checker_getIndexedAccessTypeOrUndefined(tc, objectType, keyType, checker.AccessFlagsExpressionPosition, nil, nil)
	if origin := conditionTypes.receiverObjectOrigin(node); origin.node != nil && !origin.dataCopy {
		// Index signatures have no property symbol. The target's read contract
		// still applies, and an undeclared target property may invoke a getter.
		if origin.unknownRead {
			return tc.GetUnknownType(), propertyFlags | ast.SymbolFlagsAccessor
		}
		targetType := tc.GetNonNullableType(utils.GetConstrainedTypeAtLocation(tc, origin.node))
		targetReadType := checker.Checker_getIndexedAccessTypeOrUndefined(tc, targetType, keyType, checker.AccessFlagsExpressionPosition, nil, nil)
		if targetReadType == nil {
			return tc.GetUnknownType(), propertyFlags | ast.SymbolFlagsAccessor
		}
		if origin.readTarget != nil {
			contractType := tc.GetNonNullableType(utils.GetConstrainedTypeAtLocation(tc, origin.readTarget))
			if contractRead := checker.Checker_getIndexedAccessTypeOrUndefined(tc, contractType, keyType, checker.AccessFlagsExpressionPosition, nil, nil); contractRead != nil {
				targetReadType = tc.GetUnionType([]*checker.Type{targetReadType, contractRead})
			}
		}
		if readType == nil {
			return targetReadType, propertyFlags | ast.SymbolFlagsAccessor
		}
		return tc.GetUnionType([]*checker.Type{readType, targetReadType}), propertyFlags | ast.SymbolFlagsAccessor
	}
	return readType, propertyFlags
}

func getMemberReadType(tc *checker.Checker, symbol, readSymbol, readContract *ast.Symbol) *checker.Type {
	if readSymbol == nil {
		// An opaque target can retain a getter absent from its structural type.
		return tc.GetUnknownType()
	}
	declaredType := checker.Checker_getTypeOfSymbol(tc, symbol)
	if readContract != nil {
		// A result annotation must not replace the target getter's actual read
		// type; retain both the implementation and its public read contract.
		return tc.GetUnionType([]*checker.Type{declaredType, checker.Checker_getTypeOfSymbol(tc, readSymbol), checker.Checker_getTypeOfSymbol(tc, readContract)})
	}
	if readSymbol == symbol {
		return declaredType
	}
	return tc.GetUnionType([]*checker.Type{declaredType, checker.Checker_getTypeOfSymbol(tc, readSymbol)})
}

func (conditionTypes *conditionTypeChecker) memberReadSymbol(node *ast.Node, symbol *ast.Symbol, objectType *checker.Type) (*ast.Symbol, *ast.Symbol, ast.SymbolFlags) {
	tc := conditionTypes.context.TypeChecker
	flags := symbol.Flags &^ ast.SymbolFlagsAccessor
	isAccessor := symbolHasAccessor(symbol)
	possibleAccessor := isAccessor
	if !possibleAccessor {
		for _, declaration := range symbol.Declarations {
			if ast.IsAccessor(declaration) {
				possibleAccessor = true
				break
			}
		}
	}
	receiver := ast.SkipOuterExpressions(node.Expression(), ast.OEKAll)
	if !possibleAccessor && ast.IsIdentifier(receiver) {
		receiverSymbol := checker.Checker_getResolvedSymbolOrNil(tc, receiver)
		if receiverSymbol == nil {
			receiverSymbol = tc.GetSymbolAtLocation(receiver)
		}
		if receiverSymbol == nil || receiverSymbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(receiverSymbol.ValueDeclaration) || receiverSymbol.ValueDeclaration.Initializer() == nil {
			return symbol, nil, flags
		}
	}
	if result := conditionTypes.receiverObjectOrigin(node); result.node != nil {
		origin, dataCopy, readTarget := result.node, result.dataCopy, result.readTarget
		// Intersections returned by Object.assign may narrow the public
		// property type even though the target still uses its own getter.
		if result.unknownRead {
			return nil, nil, flags | ast.SymbolFlagsAccessor
		}
		if !ast.IsObjectBindingPattern(origin) {
			if dataCopy && readTarget == nil && !ast.IsObjectLiteralExpression(origin) &&
				(ast.IsArrayLiteralExpression(origin) || ast.IsFunctionLike(origin)) {
				return symbol, nil, flags
			}
			// The target may be an instance, parameter, or call result. Its
			// declared getter still applies after Object.assign invokes a setter.
			var targetType *checker.Type
			if ast.IsObjectLiteralExpression(origin) {
				// Reuse the checker's stable initializer cache. Rechecking the
				// whole literal for each field read would be quadratic.
				targetType = checker.Checker_checkExpressionCached(tc, origin)
			} else {
				targetType = tc.GetNonNullableType(utils.GetConstrainedTypeAtLocation(tc, origin))
			}
			property := checker.Checker_getPropertyOfType(tc, targetType, symbol.Name)
			resolvedOrigin, _ := conditionTypes.resolveInitializer(origin)
			hasAccessor := false
			if resolvedOrigin != nil && ast.IsObjectLiteralExpression(resolvedOrigin) {
				// Generic spreads can retain the source's getter symbols. Only
				// accessors declared on this literal remain getters in the copy.
				literalProperty := property
				if resolvedOrigin != origin {
					literalType := checker.Checker_checkExpressionCached(tc, resolvedOrigin)
					literalProperty = checker.Checker_getPropertyOfType(tc, literalType, symbol.Name)
				}
				hasAccessor = literalProperty != nil && originHasAccessorProperty(resolvedOrigin, literalProperty)
			} else if property != nil {
				hasAccessor = typeHasAccessorProperty(tc, targetType, property)
			}
			if readTarget != nil {
				// Nested assignments must not replace a public getter contract
				// with the narrower type of its original initializer.
				readTargetType := tc.GetNonNullableType(utils.GetConstrainedTypeAtLocation(tc, readTarget))
				readProperty := checker.Checker_getPropertyOfType(tc, readTargetType, symbol.Name)
				if readProperty != nil && (hasAccessor || typeHasAccessorProperty(tc, readTargetType, readProperty)) {
					if hasAccessor {
						return property, readProperty, flags | ast.SymbolFlagsAccessor
					}
					return readProperty, nil, flags | ast.SymbolFlagsAccessor
				}
			}
			if hasAccessor {
				return property, nil, flags | ast.SymbolFlagsAccessor
			}
			if dataCopy {
				return symbol, nil, flags
			}
			origin = resolvedOrigin
			if origin != nil && ast.IsObjectLiteralExpression(origin) && property != nil {
				declaration := property.ValueDeclaration
				if declaration != nil && declaration.Parent == origin &&
					(ast.IsShorthandPropertyAssignment(declaration) || ast.IsPropertyAssignment(declaration) && !ast.IsProtoSetter(declaration.Name())) {
					return symbol, nil, flags
				}
			}
			// A structural target type does not prove a data field. Even new
			// expressions may create proxies or call an unknown constructor.
			if origin != nil && ast.IsNewExpression(origin) {
				constructor, _ := conditionTypes.resolveInitializer(origin.Expression())
				if constructor != nil {
					if ast.IsClassExpression(constructor) {
						return symbol, nil, flags
					}
					if class := tc.GetSymbolAtLocation(constructor); class != nil && checker.SkipAlias(class, tc).Flags&ast.SymbolFlagsClass != 0 {
						return symbol, nil, flags
					}
				}
			}
			return property, nil, flags | ast.SymbolFlagsAccessor
		}
		return symbol, nil, flags
	}
	if isAccessor {
		if objectType == nil {
			objectType = tc.GetNonNullableType(utils.GetConstrainedTypeAtLocation(tc, node.Expression()))
		}
		if typeHasAccessorProperty(tc, objectType, symbol) {
			flags |= ast.SymbolFlagsAccessor
		}
	}
	return symbol, nil, flags
}

func symbolHasAccessor(symbol *ast.Symbol) bool {
	if len(symbol.Declarations) == 1 && ast.IsAutoAccessorPropertyDeclaration(symbol.Declarations[0]) {
		declaration := symbol.Declarations[0]
		isPrivate := ast.IsPrivateIdentifier(declaration.Name()) || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsPrivate)
		return !isPrivate || ast.HasDecorators(declaration)
	}
	if symbol.Flags&ast.SymbolFlagsAccessor != 0 {
		return true
	}
	for _, declaration := range symbol.Declarations {
		if ast.IsAutoAccessorPropertyDeclaration(declaration) {
			// An undecorated private auto-accessor cannot be overridden and
			// reads exactly the value written to its backing field.
			isPrivate := ast.IsPrivateIdentifier(declaration.Name()) || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsPrivate)
			if !isPrivate || ast.HasDecorators(declaration) {
				return true
			}
		}
		// Synthetic union and mapped properties can retain real accessors.
		if symbol.CheckFlags&(ast.CheckFlagsSyntheticProperty|ast.CheckFlagsMapped) != 0 && ast.IsAccessor(declaration) {
			return true
		}
	}
	return false
}

func typeHasAccessorProperty(tc *checker.Checker, objectType *checker.Type, symbol *ast.Symbol) bool {
	if objectType.ObjectFlags()&checker.ObjectFlagsMapped != 0 {
		// Resolve mapped properties back to the checker-provided source type,
		// preserving the distinction between a getter and a copied data field.
		sourceType := checker.Checker_getModifiersTypeFromMappedType(tc, objectType)
		if sourceType != objectType && !utils.IsTypeUnknownType(sourceType) {
			found := false
			for _, root := range tc.GetRootSymbols(symbol) {
				if property := checker.Checker_getPropertyOfType(tc, sourceType, root.Name); property != nil {
					found = true
					if typeHasAccessorProperty(tc, sourceType, property) {
						return true
					}
				}
			}
			if found {
				return false
			}
		}
	}
	if objectType.IsUnion() || objectType.IsIntersection() {
		for _, part := range objectType.Types() {
			property := checker.Checker_getPropertyOfType(tc, part, symbol.Name)
			if property != nil && typeHasAccessorProperty(tc, part, property) {
				return true
			}
		}
		return false
	}
	if owner := objectType.Symbol(); owner != nil && owner.ValueDeclaration != nil && ast.IsObjectLiteralExpression(owner.ValueDeclaration) {
		// Spreading invokes source getters and stores their values. The
		// checker can reuse source symbols, so retain only accessors declared
		// on this object literal itself, including after type widening.
		return originHasAccessorProperty(owner.ValueDeclaration, symbol)
	}
	return symbolHasAccessor(symbol)
}

func originHasAccessorProperty(literal *ast.Node, symbol *ast.Symbol) bool {
	for _, declaration := range symbol.Declarations {
		if declaration.Parent == literal && ast.IsAccessor(declaration) {
			return true
		}
	}
	return false
}

type initializerOrigin struct {
	node, readTarget *ast.Node
}

func (conditionTypes *conditionTypeChecker) resolveInitializer(node *ast.Node) (origin *ast.Node, readTarget *ast.Node) {
	assertion := ast.SkipOuterExpressions(node, ast.OEKAll&^ast.OEKAssertions)
	if ast.IsAssertionExpression(assertion) || ast.IsNonNullExpression(assertion) {
		readTarget = assertion
	}
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	if !ast.IsIdentifier(node) {
		return node, readTarget
	}
	tc := conditionTypes.context.TypeChecker
	symbol := checker.Checker_getResolvedSymbolOrNil(tc, node)
	if symbol == nil {
		symbol = tc.GetSymbolAtLocation(node)
	}
	if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsVariableDeclaration(symbol.ValueDeclaration) || symbol.ValueDeclaration.Initializer() == nil {
		return node, readTarget
	}
	if readTarget == nil && symbol.ValueDeclaration.Type() != nil {
		readTarget = node
	}
	resolved, cached := conditionTypes.initializers[symbol]
	if !cached {
		initializer, ok := conditionTypes.staticEvaluator().ResolveIdentifierInitializer(node)
		if !ok {
			return node, readTarget
		}
		if conditionTypes.initializers == nil {
			conditionTypes.initializers = make(map[*ast.Symbol]initializerOrigin)
		}
		// The empty entry terminates cycles. Cache the stable initializer and
		// its own contract, never an assertion or flow type from this use.
		conditionTypes.initializers[symbol] = initializerOrigin{}
		resolved.node, resolved.readTarget = conditionTypes.resolveInitializer(initializer)
		conditionTypes.initializers[symbol] = resolved
	}
	if readTarget == nil {
		readTarget = resolved.readTarget
	}
	return resolved.node, readTarget
}

type receiverOrigin struct {
	node, readTarget      *ast.Node
	dataCopy, unknownRead bool
}

func (conditionTypes *conditionTypeChecker) receiverObjectOrigin(access *ast.Node) receiverOrigin {
	tc := conditionTypes.context.TypeChecker
	node := access.Expression()
	var readTarget *ast.Node
	var seen map[*ast.Node]bool
	var assignedTarget *ast.Node
	finish := func(origin, allocation *ast.Node, dataCopy bool) receiverOrigin {
		unknownRead := assignedTarget != nil && allocation != nil && conditionTypes.copyHasDescriptorMutation(allocation, access, seen)
		return receiverOrigin{node: origin, readTarget: readTarget, dataCopy: dataCopy && !unknownRead, unknownRead: unknownRead}
	}
	for {
		resolved, declaredReadTarget := conditionTypes.resolveInitializer(node)
		node = resolved
		if readTarget == nil {
			readTarget = declaredReadTarget
		}
		if node == nil {
			return finish(nil, nil, false)
		}
		if ast.IsIdentifier(node) {
			symbol := checker.Checker_getResolvedSymbolOrNil(tc, node)
			if symbol == nil {
				symbol = tc.GetSymbolAtLocation(node)
			}
			if symbol != nil && symbol.ValueDeclaration != nil {
				declaration := symbol.ValueDeclaration
				if assignedTarget != nil && ast.IsFunctionDeclaration(declaration) && declaration.Body() != nil {
					return finish(declaration, declaration, true)
				}
				if ast.IsBindingElement(declaration) && declaration.AsBindingElement().DotDotDotToken != nil && ast.IsObjectBindingPattern(declaration.Parent) && checker.Checker_isConstantReference(tc, node) {
					return finish(declaration.Parent, declaration, true)
				}
			}
			return finish(assignedTarget, nil, false)
		}
		if assignedTarget != nil && (ast.IsArrayLiteralExpression(node) || ast.IsFunctionExpressionOrArrowFunction(node)) {
			return finish(node, node, true)
		}
		if ast.IsObjectLiteralExpression(node) {
			origin := node
			if assignedTarget != nil {
				origin = assignedTarget
			}
			return finish(origin, node, conditionTypes.literalCreatesDataProperties(node))
		}
		if assignedTarget != nil && ast.IsNewExpression(node) {
			// Keep the target's public read contract before inspecting its
			// initializer: a constructor may return a different object.
			return finish(assignedTarget, nil, false)
		}
		if !ast.IsCallExpression(node) || len(node.Arguments()) == 0 || seen[node] {
			return finish(assignedTarget, nil, false)
		}
		callee, _ := conditionTypes.resolveInitializer(node.Expression())
		if callee == nil || !ast.IsAccessExpression(callee) {
			return finish(assignedTarget, nil, false)
		}
		name, ok := checker.Checker_getAccessedPropertyName(tc, callee)
		if !ok || name != "assign" {
			return finish(assignedTarget, nil, false)
		}
		object, _ := conditionTypes.resolveInitializer(callee.Expression())
		if object == nil {
			return finish(assignedTarget, nil, false)
		}
		symbol := tc.GetSymbolAtLocation(object)
		if symbol == nil || symbol.Name != "Object" || !utils.IsSymbolFromDefaultLibrary(conditionTypes.context.Program(), symbol) {
			return finish(assignedTarget, nil, false)
		}
		// Object.assign writes values into its target; source accessors do
		// not become accessors there. Existing target accessors still apply.
		if seen == nil {
			seen = make(map[*ast.Node]bool)
		}
		seen[node] = true
		assignedTarget = node.Arguments()[0]
		node = assignedTarget
	}
}

func (conditionTypes *conditionTypeChecker) copyHasDescriptorMutation(origin, access *ast.Node, copies map[*ast.Node]bool) bool {
	tc := conditionTypes.context.TypeChecker
	for call := range copies {
		for _, source := range call.Arguments()[1:] {
			if ast.IsSpreadElement(source) {
				return true
			}
			// An own __proto__ key is copied through the target's inherited
			// setter, which can install accessors before later source writes.
			for _, sourceType := range utils.UnionTypeParts(utils.GetConstrainedTypeAtLocation(tc, source)) {
				if checker.Checker_getPropertyOfType(tc, sourceType, "__proto__") != nil ||
					checker.Checker_getIndexTypeOfType(tc, sourceType, tc.GetStringType()) != nil {
					return true
				}
			}
		}
	}
	if conditionTypes.mutationEvaluator == nil {
		ctx := conditionTypes.context
		var references utils.StaticReferenceResolver
		if ctx.Refs != nil {
			references = ctx.Refs
		}
		conditionTypes.mutationEvaluator = utils.NewStaticStringEvaluatorWithCallEscapes(ctx.TypeChecker, ctx.SourceFile, references)
	}
	evaluator := conditionTypes.mutationEvaluator
	allocation := origin
	receiver := ast.SkipOuterExpressions(access.Expression(), ast.OEKAll)
	if ast.IsIdentifier(receiver) && evaluator.HasPropertyShapeMutation(receiver, allocation, access, copies) {
		return true
	}
	if (ast.IsFunctionLike(origin) || ast.IsBindingElement(origin)) && origin.Name() != nil && evaluator.HasPropertyShapeMutation(origin.Name(), allocation, access, copies) {
		return true
	}
	// Follow only transparent wrappers and this copy chain to the binding
	// owning the fresh target. The evaluator already follows its other aliases.
	for origin.Parent != nil && (ast.IsOuterExpression(origin.Parent, ast.OEKAll) || copies[origin.Parent]) {
		origin = origin.Parent
	}
	if origin.Parent != nil && ast.IsVariableDeclaration(origin.Parent) {
		if name := ast.GetAssignedName(origin); name != nil {
			return evaluator.HasPropertyShapeMutation(name, allocation, access, copies)
		}
	}
	return false
}

// This checker belongs to one file. Cache only the immutable literal shape,
// so reading every field of a large object does not repeatedly scan its peers.
func (conditionTypes *conditionTypeChecker) literalCreatesDataProperties(node *ast.Node) bool {
	if result, ok := conditionTypes.dataCopyLiterals[node]; ok {
		return result
	}
	result := !slices.ContainsFunc(node.AsObjectLiteralExpression().Properties.Nodes, func(property *ast.Node) bool {
		return ast.IsPropertyAssignment(property) && ast.IsProtoSetter(property.Name()) &&
			ast.SkipOuterExpressions(property.Initializer(), ast.OEKAll).Kind != ast.KindNullKeyword
	})
	if conditionTypes.dataCopyLiterals == nil {
		conditionTypes.dataCopyLiterals = make(map[*ast.Node]bool)
	}
	conditionTypes.dataCopyLiterals[node] = result
	return result
}

func (conditionTypes *conditionTypeChecker) getCapturedReadType(node *ast.Node) (captured *checker.Type) {
	tc := conditionTypes.context.TypeChecker
	symbol := checker.Checker_getResolvedSymbolOrNil(tc, node)
	if symbol == nil {
		symbol = tc.GetSymbolAtLocation(node)
	}
	if symbol == nil || symbol.ValueDeclaration == nil {
		return nil
	}
	declaration := symbol.ValueDeclaration
	if ast.IsVariableDeclaration(declaration) {
		if declaration.Initializer() == nil {
			return nil
		}
		initializer := ast.SkipParentheses(declaration.Initializer())
		if !ast.IsIdentifier(initializer) && !ast.IsAccessExpression(initializer) && !ast.IsCallExpression(initializer) && !ast.IsSatisfiesExpression(initializer) {
			return nil
		}
	} else if !ast.IsBindingElement(declaration) {
		return nil
	}
	if result, ok := conditionTypes.capturedReadTypes[symbol]; ok {
		return result
	}
	if conditionTypes.capturedReadTypes == nil {
		conditionTypes.capturedReadTypes = make(map[*ast.Symbol]*checker.Type)
	}
	// Cache the stable initializer read, not the type at a use site. Guards
	// are still evaluated by the checker for each individual reference.
	conditionTypes.capturedReadTypes[symbol] = nil
	defer func() {
		if captured != nil {
			root := ast.GetRootDeclaration(declaration)
			if ast.IsVariableDeclaration(root) && root.Type() != nil {
				contract := checker.Checker_getTypeOfSymbol(tc, symbol)
				if utils.IsTypeAnyType(contract) || utils.IsTypeUnknownType(contract) {
					captured = contract
				} else {
					// Capturing an object does not freeze its fields. Preserve
					// annotated object contracts without widening the captured
					// primitive value or adding nullish values from an annotation.
					objects := checker.Checker_getTypeWithFacts(tc, tc.GetNonNullableType(contract), checker.TypeFactsTypeofEQObject|checker.TypeFactsTypeofEQFunction)
					captured = tc.GetUnionType([]*checker.Type{captured, objects})
				}
			}
		}
		conditionTypes.capturedReadTypes[symbol] = captured
	}()
	if ast.IsBindingElement(declaration) {
		root := ast.GetRootDeclaration(declaration)
		if !ast.IsVariableDeclaration(root) || root.Initializer() == nil ||
			declaration.AsBindingElement().DotDotDotToken != nil || !checker.Checker_isConstantReference(tc, node) {
			return nil
		}
		// Use the checker's equivalent read instead of reimplementing binding
		// names, nesting, and their original control-flow points.
		reference := getBindingReadReference(tc, declaration)
		if reference == nil {
			return nil
		}
		originalType := utils.GetConstrainedTypeAtLocation(tc, reference)
		readType := conditionTypes.getConditionType(reference, nil)
		if ast.IsArrayBindingPattern(declaration.Parent) {
			if readType != originalType && utils.IsTypeUnknownType(readType) {
				// A visible descriptor mutation can also replace the iterator.
				if origin := conditionTypes.receiverObjectOrigin(reference); origin.unknownRead {
					return readType
				}
			}
			parentType := utils.GetConstrainedTypeAtLocation(tc, reference.Expression())
			hasCustomIterator := func(t *checker.Type) bool {
				iterator := utils.GetWellKnownSymbolPropertyOfType(t, "iterator", tc)
				return iterator != nil && slices.ContainsFunc(iterator.Declarations, func(declaration *ast.Node) bool {
					return !conditionTypes.context.Program().IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(declaration))
				})
			}
			if !tc.IsArrayLikeType(parentType) || hasCustomIterator(parentType) {
				// Iterable destructuring reads the iterator, not an indexed getter.
				// An Object.assign intersection can hide a replacement iterator's
				// nullable yield type, just as it can hide a getter's read type.
				var iteratedTypes []*checker.Type
				for _, variant := range utils.UnionTypeParts(parentType) {
					parts := utils.IntersectionTypeParts(variant)
					custom := slices.ContainsFunc(parts, hasCustomIterator)
					for _, part := range parts {
						if custom && !hasCustomIterator(part) {
							continue
						}
						if iterated := checker.Checker_getIterationTypeOfIterable(tc, checker.IterationUseDestructuring, checker.IterationTypeKindYield, part, nil); iterated != nil {
							iteratedTypes = append(iteratedTypes, iterated)
						} else if custom {
							iteratedTypes = append(iteratedTypes, tc.GetUnknownType())
						}
					}
				}
				if len(iteratedTypes) != 0 {
					readType = tc.GetUnionType(iteratedTypes)
				}
			}
		}
		if readType == originalType {
			return nil
		}
		for parent := declaration.Parent.Parent; ast.IsBindingElement(parent); parent = parent.Parent.Parent {
			if parent.Initializer() != nil {
				parentRead := getBindingReadReference(tc, parent)
				if parentRead != nil && checker.Checker_getTypeWithFacts(tc, conditionTypes.getConditionType(parentRead, nil), checker.TypeFactsEQUndefined).Flags()&checker.TypeFlagsNever != 0 {
					continue
				}
				// The synthetic access omits enclosing defaults. Keep the
				// checker's binding contract, which includes those alternatives.
				readType = tc.GetUnionType([]*checker.Type{readType, checker.Checker_getTypeOfSymbol(tc, symbol)})
				break
			}
		}
		if declaration.Initializer() != nil &&
			checker.Checker_getTypeWithFacts(tc, readType, checker.TypeFactsEQUndefined).Flags()&checker.TypeFlagsNever == 0 {
			readType = tc.GetUnionType([]*checker.Type{
				checker.Checker_getTypeWithFacts(tc, readType, checker.TypeFactsNEUndefined),
				conditionTypes.getConditionType(declaration.Initializer(), nil),
			})
		}
		return readType
	}
	// Follow one alias at a time so each symbol reuses its cached initializer.
	initializer, ok := conditionTypes.staticEvaluator().ResolveIdentifierInitializer(node)
	if !ok {
		return nil
	}
	initializer = ast.SkipParentheses(initializer)
	if !ast.IsIdentifier(initializer) && !ast.IsAccessExpression(initializer) && !ast.IsCallExpression(initializer) && !ast.IsSatisfiesExpression(initializer) {
		return nil
	}
	var accessor *ast.Node
	if ast.IsCallExpression(initializer) {
		accessor = conditionTypes.findAccessorRead(initializer.Expression())
		if accessor == nil {
			return nil
		}
	}
	t := utils.GetConstrainedTypeAtLocation(tc, initializer)
	readType := conditionTypes.getConditionType(initializer, accessor)
	if readType == t {
		return nil
	}
	return readType
}

func getBindingReadReference(tc *checker.Checker, declaration *ast.Node) *ast.Node {
	if reference := checker.Checker_getSyntheticElementAccess(tc, declaration); reference != nil {
		return reference
	}
	// The checker's flow helper omits calls and other initializers without a
	// flow node. Reuse its binding paths and keys for a plain declared read.
	parent := checker.Checker_getParentElementAccess(tc, declaration)
	if parent == nil && ast.IsBindingElement(declaration.Parent.Parent) {
		parent = getBindingReadReference(tc, declaration.Parent.Parent)
	}
	if parent == nil {
		return nil
	}
	var factory ast.NodeFactory
	var key *ast.Node
	if name, known := checker.Checker_getDestructuringPropertyName(tc, declaration); known {
		key = factory.NewStringLiteral(name, ast.TokenFlagsNone)
		key.Loc = declaration.Loc
	} else if name := declaration.PropertyName(); name != nil && ast.IsComputedPropertyName(name) {
		key = name.Expression()
	} else {
		return nil
	}
	if !ast.IsLeftHandSideExpression(parent) {
		parent = factory.NewParenthesizedExpression(parent)
		parent.Loc = declaration.Loc
	}
	reference := factory.NewElementAccessExpression(parent, nil, key, ast.NodeFlagsNone)
	reference.Loc, reference.Parent = declaration.Loc, declaration
	// Preserve the parsed expression parents and their lexical bindings.
	if key.Parent == nil {
		key.Parent = reference
	}
	if parent.Parent == nil {
		parent.Parent = reference
	}
	return reference
}

// A getter in the receiver can return a different object on the next read.
// Assignments to a property of the previous object do not narrow the new one.
func (conditionTypes *conditionTypeChecker) hasAccessorRead(node *ast.Node) bool {
	return conditionTypes.findAccessorRead(node) != nil
}

func (conditionTypes *conditionTypeChecker) findAccessorRead(node *ast.Node) *ast.Node {
	tc := conditionTypes.context.TypeChecker
	for {
		node = ast.SkipOuterExpressions(node, ast.OEKAll)
		if ast.IsCallExpression(node) {
			node = node.Expression()
			continue
		}
		if !ast.IsAccessExpression(node) {
			if ast.IsIdentifier(node) && conditionTypes.getCapturedReadType(node) != nil {
				return node
			}
			return nil
		}
		symbol := checker.Checker_getResolvedSymbolOrNil(tc, node)
		if symbol == nil {
			tc.GetTypeAtLocation(node)
			symbol = checker.Checker_getResolvedSymbolOrNil(tc, node)
		}
		var flags ast.SymbolFlags
		if symbol != nil {
			_, _, flags = conditionTypes.memberReadSymbol(node, symbol, nil)
		} else if hasUnionOrConstrainedKey(tc, node) {
			_, flags = conditionTypes.getDeclaredMemberReadType(node, nil)
		} else if origin := conditionTypes.receiverObjectOrigin(node); origin.node != nil && !origin.dataCopy {
			_, flags = conditionTypes.getDeclaredMemberReadType(node, nil)
		}
		if flags&ast.SymbolFlagsAccessor != 0 {
			return node
		}
		node = node.Expression()
	}
}

func hasUnionOrConstrainedKey(tc *checker.Checker, node *ast.Node) bool {
	if !ast.IsElementAccessExpression(node) {
		return false
	}
	// These keys can select accessors without a single resolved symbol.
	keyType := tc.GetTypeAtLocation(node.AsElementAccessExpression().ArgumentExpression)
	constraint := checker.Checker_getBaseConstraintOfType(tc, keyType)
	return keyType.IsUnion() || constraint != nil && constraint != keyType
}

func getConditionFlowType(tc *checker.Checker, node *ast.Node, declaredType *checker.Type) *checker.Type {
	return checker.Checker_getFlowTypeOfReference(tc, getConditionReference(node), declaredType)
}

func getConditionReference(node *ast.Node) *ast.Node {
	if ast.IsElementAccessExpression(node) {
		element := node.AsElementAccessExpression()
		key := ast.SkipParentheses(element.ArgumentExpression)
		if key != element.ArgumentExpression {
			// Query an equivalent reference without changing the parsed tree.
			// Keep its original flow point and surrounding assertion context.
			var factory ast.NodeFactory
			reference := factory.UpdateElementAccessExpression(element, element.Expression, element.QuestionDotToken, key, node.Flags)
			reference.Parent = node.Parent
			reference.FlowNodeData().FlowNode = element.FlowNode
			if reference.FlowNodeData().FlowNode == nil {
				// The binder may not attach flow to a parenthesized key. Its
				// identifier, or the receiver before a literal key, has that point.
				for _, part := range []*ast.Node{key, ast.SkipParentheses(element.Expression)} {
					if flow := part.FlowNodeData(); flow != nil && flow.FlowNode != nil {
						reference.FlowNodeData().FlowNode = flow.FlowNode
						break
					}
				}
			}
			node = reference
		}
	}
	return node
}

func hasKnownNonNullishAssignment(tc *checker.Checker, node *ast.Node, evaluator *utils.StaticStringEvaluator) bool {
	reference := getConditionReference(node)
	flow := reference.FlowNodeData().FlowNode
	for flow != nil && flow.Flags&ast.FlowFlagsAssignment != 0 && flow.Node != nil &&
		!checker.Checker_isMatchingReference(tc, reference, flow.Node) {
		// Only cross simple declarations. Destructuring, spread, coercion and
		// resource initialization can invoke user code without an explicit call.
		if !ast.IsVariableDeclaration(flow.Node) || !ast.IsIdentifier(flow.Node.Name()) ||
			ast.IsVarUsing(flow.Node.Parent) || ast.IsVarAwaitUsing(flow.Node.Parent) ||
			checker.Checker_containsMatchingReference(tc, reference, flow.Node) {
			return false
		}
		initializer := utils.SkipAssertionsAndParens(flow.Node.Initializer())
		if initializer != nil && !ast.IsIdentifier(initializer) && initializer.Kind != ast.KindNullKeyword &&
			!ast.IsPrimitiveLiteralValue(initializer, true) && !ast.IsEmptyObjectLiteral(initializer) &&
			!ast.IsEmptyArrayLiteral(initializer) && !ast.IsFunctionExpression(initializer) && !ast.IsArrowFunction(initializer) {
			return false
		}
		flow = flow.Antecedent
	}
	if flow == nil || flow.Flags&ast.FlowFlagsAssignment == 0 || flow.Node == nil {
		return false
	}
	assignment := ast.GetAssignmentTarget(flow.Node)
	if assignment == nil || !ast.IsBinaryExpression(assignment) || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}
	nullish, known := evaluator.EvalControlFlowNullish(assignment.AsBinaryExpression().Right)
	return known && !nullish
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
	if callback != nil && (ast.IsArrowFunction(callback) || ast.IsFunctionExpression(callback)) && ast.GetFunctionFlags(callback) == ast.FunctionFlagsNormal {
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
