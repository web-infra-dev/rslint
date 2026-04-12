package no_unnecessary_condition

import (
	"fmt"
	"math"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// strictNullishFlag matches the original typescript-eslint behavior:
// only null and undefined are considered "nullish" for isAlwaysNullish.
var strictNullishFlag = checker.TypeFlagsUndefined | checker.TypeFlagsNull

// isNullishType checks if a type is strictly null or undefined.
func isNullishType(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, strictNullishFlag)
}

// isAlwaysNullish checks if ALL union constituents are null or undefined.
func isAlwaysNullish(t *checker.Type) bool {
	return utils.Every(utils.UnionTypeParts(t), func(part *checker.Type) bool {
		return isNullishType(part)
	})
}

// isPossiblyNullish checks if ANY union constituent is null or undefined.
// Matches original typescript-eslint: only Null and Undefined, not Void.
func isPossiblyNullish(t *checker.Type) bool {
	return utils.Some(utils.UnionTypeParts(t), func(part *checker.Type) bool {
		return utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined|checker.TypeFlagsNull)
	})
}

// Sentinel types for distinguishing null and undefined in comparisons.
type jsNullType struct{}
type jsUndefinedType struct{}

var jsNull = jsNullType{}
var jsUndefined = jsUndefinedType{}

type staticValue struct {
	value interface{}
	ok    bool
}

func toStaticValue(t *checker.Type) staticValue {
	if utils.IsTrueLiteralType(t) {
		return staticValue{value: true, ok: true}
	}
	if utils.IsFalseLiteralType(t) {
		return staticValue{value: false, ok: true}
	}
	flags := checker.Type_flags(t)
	if flags == checker.TypeFlagsUndefined {
		return staticValue{value: jsUndefined, ok: true}
	}
	if flags == checker.TypeFlagsNull {
		return staticValue{value: jsNull, ok: true}
	}
	if t.IsStringLiteral() || t.IsNumberLiteral() || t.IsBigIntLiteral() {
		return staticValue{value: t.AsLiteralType().Value(), ok: true}
	}
	return staticValue{ok: false}
}

var boolOperators = map[string]bool{
	"<": true, ">": true, "<=": true, ">=": true,
	"==": true, "===": true, "!=": true, "!==": true,
}

func isBoolOperator(op string) bool {
	return boolOperators[op]
}

// booleanComparison mimics JavaScript comparison semantics for literal types.
func booleanComparison(left interface{}, operator string, right interface{}) bool {
	switch operator {
	case "===":
		return strictEqual(left, right)
	case "!==":
		return !strictEqual(left, right)
	case "==":
		return looseEqual(left, right)
	case "!=":
		return !looseEqual(left, right)
	case "<", "<=", ">", ">=":
		cmp, ok := relationalCompare(left, right)
		if !ok {
			return false
		}
		switch operator {
		case "<":
			return cmp < 0
		case "<=":
			return cmp <= 0
		case ">":
			return cmp > 0
		case ">=":
			return cmp >= 0
		}
	}
	return false
}

func isJSNullish(v interface{}) bool {
	_, isNull := v.(jsNullType)
	_, isUndef := v.(jsUndefinedType)
	return isNull || isUndef
}

// valueString returns a canonical string for comparing literal values.
// Uses fmt.Stringer for named types (Number, PseudoBigInt).
func valueString(v interface{}) string {
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%v", v)
}

func strictEqual(a, b interface{}) bool {
	// null === null, undefined === undefined, but null !== undefined
	_, aNull := a.(jsNullType)
	_, bNull := b.(jsNullType)
	_, aUndef := a.(jsUndefinedType)
	_, bUndef := b.(jsUndefinedType)
	if aNull && bNull {
		return true
	}
	if aUndef && bUndef {
		return true
	}
	if aNull || bNull || aUndef || bUndef {
		return false
	}
	// JS strict equality requires same type — different Go types means different JS types
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		return false
	}
	return valueString(a) == valueString(b)
}

func looseEqual(a, b interface{}) bool {
	// null == undefined is true in JS
	if isJSNullish(a) && isJSNullish(b) {
		return true
	}
	// null/undefined != anything else
	if isJSNullish(a) || isJSNullish(b) {
		return false
	}
	return valueString(a) == valueString(b)
}

// relationalCompare compares two literal values for <, <=, >, >=.
// JS uses numeric comparison for numbers/bigints/booleans, and
// lexicographic comparison when both operands are strings.
func relationalCompare(a, b interface{}) (int, bool) {
	// String vs string: lexicographic comparison
	sa, aStr := a.(string)
	sb, bStr := b.(string)
	if aStr && bStr {
		if sa < sb {
			return -1, true
		}
		if sa > sb {
			return 1, true
		}
		return 0, true
	}

	na, aOK := toNumber(a)
	nb, bOK := toNumber(b)
	if !aOK || !bOK || math.IsNaN(na) || math.IsNaN(nb) {
		return 0, false
	}
	if na < nb {
		return -1, true
	}
	if na > nb {
		return 1, true
	}
	return 0, true
}

// toNumber converts a literal type value to float64 for relational comparison.
// tsgo stores number literals as Number (a named float64 type) and bigint
// literals as PseudoBigInt — both implement fmt.Stringer.
func toNumber(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case bool:
		if val {
			return 1, true
		}
		return 0, true
	case string:
		// Strings can't be numerically compared — use stringCompare instead
		return 0, false
	default:
		// Handle Number and PseudoBigInt via fmt.Stringer
		if s, ok := v.(fmt.Stringer); ok {
			str := s.String()
			f, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return 0, false
			}
			return f, true
		}
		return 0, false
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

func parseOptions(options any) ruleOptions {
	opts := ruleOptions{
		allowConstantLoopConditions: loopConditionNever,
	}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["allowConstantLoopConditions"]; ok {
			opts.allowConstantLoopConditions = normalizeAllowConstantLoopConditions(v)
		}
		if v, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
			opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = v
		}
		if v, ok := optsMap["checkTypePredicates"].(bool); ok {
			opts.checkTypePredicates = v
		}
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
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		tc := ctx.TypeChecker

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)
		isNoUncheckedIndexedAccess := compilerOptions.NoUncheckedIndexedAccess.IsTrue()

		if !isStrictNullChecks && !opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
		}

		// --- Helper functions ---

		nodeIsArrayType := func(node *ast.Node) bool {
			nodeType := utils.GetConstrainedTypeAtLocation(tc, node)
			return utils.Some(utils.UnionTypeParts(nodeType), func(part *checker.Type) bool {
				return checker.Checker_isArrayType(tc, part)
			})
		}

		nodeIsTupleType := func(node *ast.Node) bool {
			nodeType := utils.GetConstrainedTypeAtLocation(tc, node)
			return utils.Some(utils.UnionTypeParts(nodeType), func(part *checker.Type) bool {
				return checker.IsTupleType(part)
			})
		}

		isArrayIndexExpression := func(node *ast.Node) bool {
			if !ast.IsElementAccessExpression(node) {
				return false
			}
			elem := node.AsElementAccessExpression()
			obj := elem.Expression
			return nodeIsArrayType(obj) ||
				(nodeIsTupleType(obj) && !ast.IsLiteralExpression(elem.ArgumentExpression))
		}

		// Conditional is always necessary if it involves any, unknown, or a naked type variable
		isConditionalAlwaysNecessary := func(t *checker.Type) bool {
			return utils.Some(utils.UnionTypeParts(t), func(part *checker.Type) bool {
				return utils.IsTypeAnyType(part) ||
					utils.IsTypeUnknownType(part) ||
					utils.IsTypeFlagSet(part, checker.TypeFlagsTypeVariable)
			})
		}

		isNullableMemberExpression := func(node *ast.Node) bool {
			if !ast.IsAccessExpression(node) {
				return false
			}
			objectType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
			if ast.IsElementAccessExpression(node) {
				propertyType := ctx.TypeChecker.GetTypeAtLocation(node.AsElementAccessExpression().ArgumentExpression)
				return isNullablePropertyType(tc, objectType, propertyType)
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

		var checkNode func(expression *ast.Node, isUnaryNotArgument bool, reportNode *ast.Node)
		checkNode = func(expression *ast.Node, isUnaryNotArgument bool, reportNode *ast.Node) {
			expression = ast.SkipParentheses(expression)

			if reportNode == nil {
				reportNode = expression
			}

			// Handle unary negation
			if ast.IsPrefixUnaryExpression(expression) && expression.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
				checkNode(expression.AsPrefixUnaryExpression().Operand, !isUnaryNotArgument, reportNode)
				return
			}

			// Skip array index expressions (unsound typing)
			if !isNoUncheckedIndexedAccess && isArrayIndexExpression(expression) {
				return
			}

			// For logical expressions (except ??), only check the right side
			if ast.IsBinaryExpression(expression) {
				binExpr := expression.AsBinaryExpression()
				op := binExpr.OperatorToken.Kind
				if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
					checkNode(binExpr.Right, false, nil)
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
			if !utils.IsPossiblyTruthy(t) {
				if !isUnaryNotArgument {
					ctx.ReportNode(reportNode, buildAlwaysFalsyMessage())
				} else {
					ctx.ReportNode(reportNode, buildAlwaysTruthyMessage())
				}
				return
			}
			if !utils.IsPossiblyFalsy(t) {
				if !isUnaryNotArgument {
					ctx.ReportNode(reportNode, buildAlwaysTruthyMessage())
				} else {
					ctx.ReportNode(reportNode, buildAlwaysFalsyMessage())
				}
				return
			}
		}

		checkNodeForNullish := func(node *ast.Node) {
			t := utils.GetConstrainedTypeAtLocation(tc, node)

			// Conditional is always necessary if it involves any, unknown, or type parameter
			if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsTypeVariable) {
				return
			}

			if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsNever) {
				ctx.ReportNode(node, buildNeverMessage())
				return
			}

			isMemberExpr := ast.IsAccessExpression(node)
			if !isPossiblyNullish(t) && (!isMemberExpr || !isNullableMemberExpression(node)) {
				// Skip array index expressions without noUncheckedIndexedAccess
				if isNoUncheckedIndexedAccess ||
					(!isArrayIndexExpression(node) &&
						!isChainExpressionWithOptionalArrayIndex(node, isArrayIndexExpression)) {
					ctx.ReportNode(node, buildNeverNullishMessage())
					return
				}
			} else if isAlwaysNullish(t) {
				ctx.ReportNode(node, buildAlwaysNullishMessage())
				return
			}
		}

		checkIfBoolExpressionIsNecessaryConditional := func(node *ast.Node, left *ast.Node, right *ast.Node, operator string) {
			leftType := utils.GetConstrainedTypeAtLocation(tc, left)
			rightType := utils.GetConstrainedTypeAtLocation(tc, right)

			leftStatic := toStaticValue(leftType)
			rightStatic := toStaticValue(rightType)

			if leftStatic.ok && rightStatic.ok {
				conditionIsTrue := booleanComparison(leftStatic.value, operator, rightStatic.value)
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
					flag |= checker.TypeFlagsAny | checker.TypeFlagsUnknown | checker.TypeFlagsTypeParameter | checker.TypeFlagsTypeVariable
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
			checkNode(binExpr.Left, false, nil)
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

			checkNode(test, false, nil)
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
			prevType := utils.GetConstrainedTypeAtLocation(tc, node.Expression())
			if !prevType.IsUnion() {
				return false
			}

			isComputed := ast.IsElementAccessExpression(node)
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
					propertyType := utils.GetConstrainedTypeAtLocation(tc, node.AsElementAccessExpression().ArgumentExpression)
					return isNullablePropertyType(tc, partType, propertyType)
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
			isOwnNullable := true
			if ast.IsAccessExpression(node) {
				isOwnNullable = !isMemberExpressionNullableOriginFromObject(node)
			} else if ast.IsCallExpression(node) {
				isOwnNullable = !isCallExpressionNullableOriginFromCallee(node)
			}
			return isConditionalAlwaysNecessary(t) || (isOwnNullable && utils.IsNullableType(t))
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
				truthinessArg := findTruthinessAssertedArgument(tc, callExpr)
				if truthinessArg != nil {
					checkNode(truthinessArg, false, nil)
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
				callback := callExpr.Arguments.Nodes[0]

				if ast.IsArrowFunction(callback) || ast.IsFunctionExpression(callback) {
					body := callback.Body()
					if body != nil && !ast.IsBlock(body) {
						// Arrow function with expression body: () => something
						checkNode(body, false, nil)
						return
					}
					if body != nil && ast.IsBlock(body) {
						statements := body.AsBlock().Statements.Nodes
						if len(statements) == 1 && ast.IsReturnStatement(statements[0]) {
							returnStmt := statements[0].AsReturnStatement()
							if returnStmt.Expression != nil {
								checkNode(returnStmt.Expression, false, nil)
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
				checkNode(assignExpr.Left, false, nil)
			case ast.KindQuestionQuestionEqualsToken:
				checkNodeForNullish(assignExpr.Left)
			}
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				checkNode(node.AsIfStatement().Expression, false, nil)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				checkNode(node.AsConditionalExpression().Condition, false, nil)
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
				opStr := scanner.TokenToString(op)
				if isBoolOperator(opStr) {
					checkIfBoolExpressionIsNecessaryConditional(node, binExpr.Left, binExpr.Right, opStr)
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
					"===",
				)
			},
		}
	},
})

// Helper functions

func isNullablePropertyType(tc *checker.Checker, objType *checker.Type, propertyType *checker.Type) bool {
	if propertyType.IsUnion() {
		return utils.Some(propertyType.Types(), func(part *checker.Type) bool {
			return isNullablePropertyType(tc, objType, part)
		})
	}
	if propertyType.IsNumberLiteral() || propertyType.IsStringLiteral() {
		propName := valueString(propertyType.AsLiteralType().Value())
		propType := tc.GetTypeOfPropertyOfType(objType, propName)
		if propType != nil {
			return utils.IsNullableType(propType)
		}
		// tsgo's GetTypeOfPropertyOfType doesn't resolve index signatures for
		// non-array object types (e.g., { [key: number]: T }). For these types,
		// if the literal key matches an index signature, treat the property as
		// potentially missing (matching original behavior).
		// Skip arrays — they are handled by the fallthrough getTypeName comparison.
		if !checker.Checker_isArrayType(tc, objType) && !checker.IsTupleType(objType) {
			var keyType *checker.Type
			if propertyType.IsNumberLiteral() {
				keyType = tc.GetNumberType()
			} else {
				keyType = tc.GetStringType()
			}
			indexType := checker.Checker_getIndexTypeOfType(tc, objType, keyType)
			if indexType != nil {
				return true
			}
		}
	}
	typeName := utils.GetTypeName(tc, propertyType)
	indexInfos := tc.GetIndexInfosOfType(objType)
	return utils.Some(indexInfos, func(info *checker.IndexInfo) bool {
		return utils.GetTypeName(tc, info.KeyType()) == typeName
	})
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

// firstSpreadIndex returns the index of the first spread element argument,
// or -1 if none. Arguments before the first spread can still be reliably
// mapped to parameters by index.
func firstSpreadIndex(callExpr *ast.CallExpression) int {
	for i, arg := range callExpr.Arguments.Nodes {
		if ast.IsSpreadElement(arg) {
			return i
		}
	}
	return -1
}

func findTruthinessAssertedArgument(tc *checker.Checker, callExpr *ast.CallExpression) *ast.Node {
	// Get the resolved signature
	sig := checker.Checker_getResolvedSignature(tc, callExpr.AsNode(), nil, checker.CheckModeNormal)
	if sig == nil {
		return nil
	}

	predicate := tc.GetTypePredicateOfSignature(sig)
	if predicate == nil {
		return nil
	}

	// Truthiness assertions: asserts param (no type) or param is truthy (no type)
	if predicate.Type() != nil {
		return nil
	}

	// Must be an asserts predicate
	if predicate.Kind() != checker.TypePredicateKindAssertsIdentifier {
		return nil
	}

	paramIndex := predicate.ParameterIndex()
	// Skip if parameter index is at or past a spread element (unreliable mapping)
	spreadIdx := firstSpreadIndex(callExpr)
	if spreadIdx >= 0 && int(paramIndex) >= spreadIdx {
		return nil
	}
	if int(paramIndex) >= len(callExpr.Arguments.Nodes) {
		return nil
	}
	return callExpr.Arguments.Nodes[paramIndex]
}

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
	spreadIdx := firstSpreadIndex(callExpr)
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
