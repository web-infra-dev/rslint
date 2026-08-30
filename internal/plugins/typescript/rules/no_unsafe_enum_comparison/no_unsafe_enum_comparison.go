package no_unsafe_enum_comparison

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMismatchedCaseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatchedCase",
		Description: "The case statement does not have a shared enum type with the switch predicate.",
	}
}
func buildMismatchedConditionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatchedCondition",
		Description: "The two values in this comparison do not have a shared enum type.",
	}
}

func buildReplaceValueWithEnumMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceValueWithEnum",
		Description: "Replace with an enum value comparison.",
	}
}

type staticNumber interface {
	IsNaN() bool
}

func enumLiteralMatchesStaticValue(
	enumLiteral *checker.Type,
	staticValue any,
	staticValueText string,
) bool {
	literalValue := enumLiteral.AsLiteralType().Value()
	if utils.IsTypeFlagSet(enumLiteral, checker.TypeFlagsStringLiteral) {
		enumString, enumStringOK := literalValue.(string)
		staticString, staticStringOK := staticValue.(string)
		return enumStringOK && staticStringOK && enumString == staticString
	}
	if utils.IsTypeFlagSet(enumLiteral, checker.TypeFlagsNumberLiteral) {
		staticNumber, staticNumberOK := staticValue.(staticNumber)
		return staticNumberOK && !staticNumber.IsNaN() && checker.ValueToString(literalValue) == staticValueText
	}
	return false
}

func getEnumKeyForLiteral(
	sourceFile *ast.SourceFile,
	enumLiterals []*checker.Type,
	literal *ast.Node,
	staticEvaluator *utils.StaticStringEvaluator,
) (string, bool) {
	staticValue, ok := staticEvaluator.EvalValue(literal)
	if !ok {
		return "", false
	}

	staticValueText := ""
	if _, isNumber := staticValue.(staticNumber); isNumber {
		staticValueText, ok = staticEvaluator.EvalToString(literal)
		if !ok {
			return "", false
		}
	}

	for _, enumLiteral := range enumLiterals {
		if !enumLiteralMatchesStaticValue(enumLiteral, staticValue, staticValueText) {
			continue
		}

		symbol := checker.Type_symbol(enumLiteral)
		if symbol == nil || symbol.ValueDeclaration == nil || !ast.IsEnumMember(symbol.ValueDeclaration) {
			continue
		}
		enumMember := symbol.ValueDeclaration
		enumDeclaration := enumMember.Parent
		if enumDeclaration == nil || !ast.IsEnumDeclaration(enumDeclaration) {
			continue
		}

		enumName := enumDeclaration.Name().Text()
		memberName := enumMember.Name()
		switch memberName.Kind {
		case ast.KindIdentifier:
			return enumName + "." + memberName.Text(), true
		case ast.KindStringLiteral:
			escapedName := strings.ReplaceAll(memberName.Text(), "'", "\\'")
			return enumName + "['" + escapedName + "']", true
		case ast.KindComputedPropertyName:
			expression := memberName.AsComputedPropertyName().Expression
			return enumName + "[" + utils.TrimmedNodeText(sourceFile, expression) + "]", true
		}
	}

	return "", false
}

/**
 * @returns What type a type's enum value is (number or string), if either.
 */
func getEnumValueType(t *checker.Type) checker.TypeFlags {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
		if utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLiteral) {
			return checker.TypeFlagsNumber
		}
		return checker.TypeFlagsString
	}
	return checker.TypeFlagsNone
}

/**
 * @returns Whether the right type is an unsafe comparison against any left type.
 */
func typeViolates(leftTypeParts []*checker.Type, rightType *checker.Type) bool {
	hasNumberEnum, hasStringEnum := false, false
	for _, typePart := range leftTypeParts {
		switch getEnumValueType(typePart) {
		case checker.TypeFlagsNumber:
			hasNumberEnum = true
		case checker.TypeFlagsString:
			hasStringEnum = true
		}
	}

	return (hasNumberEnum && isTypeLike(rightType, checker.TypeFlagsNumberLike)) ||
		(hasStringEnum && isTypeLike(rightType, checker.TypeFlagsStringLike))
}

func isTypeLike(t *checker.Type, flags checker.TypeFlags) bool {
	for _, unionPart := range utils.UnionTypeParts(t) {
		matches := false
		for _, intersectionPart := range utils.IntersectionTypeParts(unionPart) {
			if utils.IsTypeFlagSet(intersectionPart, flags) {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}
	return true
}

var NoUnsafeEnumComparisonRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-enum-comparison",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		staticEvaluator := utils.NewStaticStringEvaluatorWithoutScope()
		isMismatchedComparison := func(
			leftType *checker.Type,
			rightType *checker.Type,
		) bool {
			// Allow comparisons that don't have anything to do with enums:
			//
			// ```ts
			// 1 === 2;
			// ```
			leftEnumTypes := utils.GetEnumTypes(ctx.TypeChecker, leftType)
			rightEnumTypes := utils.NewSetFromItems(utils.GetEnumTypes(ctx.TypeChecker, rightType)...)
			if len(leftEnumTypes) == 0 && rightEnumTypes.Len() == 0 {
				return false
			}

			// Allow comparisons that share an enum type:
			//
			// ```ts
			// Fruit.Apple === Fruit.Banana;
			// ```
			for _, leftEnumType := range leftEnumTypes {
				if rightEnumTypes.Has(leftEnumType) {
					return false
				}
			}

			// We need to split the type into the union type parts in order to find
			// valid enum comparisons like:
			//
			// ```ts
			// declare const something: Fruit | Vegetable;
			// something === Fruit.Apple;
			// ```
			leftTypeParts := utils.UnionTypeParts(leftType)
			rightTypeParts := utils.UnionTypeParts(rightType)

			// If a type exists in both sides, we consider this comparison safe:
			//
			// ```ts
			// declare const fruit: Fruit.Apple | 0;
			// fruit === 0;
			// ```
			for _, leftTypePart := range leftTypeParts {
				if slices.Contains(rightTypeParts, leftTypePart) {
					return false
				}
			}

			l := typeViolates(leftTypeParts, rightType)

			return (l || typeViolates(rightTypeParts, leftType))
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				opKind := expr.OperatorToken.Kind
				if opKind != ast.KindLessThanToken && opKind != ast.KindLessThanEqualsToken && opKind != ast.KindGreaterThanToken && opKind != ast.KindGreaterThanEqualsToken && opKind != ast.KindEqualsEqualsToken && opKind != ast.KindEqualsEqualsEqualsToken && opKind != ast.KindExclamationEqualsToken && opKind != ast.KindExclamationEqualsEqualsToken {
					return
				}

				leftType := ctx.TypeChecker.GetTypeAtLocation(expr.Left)
				rightType := ctx.TypeChecker.GetTypeAtLocation(expr.Right)

				if isMismatchedComparison(leftType, rightType) {
					ctx.ReportNodeWithDeferredSuggestions(node, buildMismatchedConditionMessage(), func() []rule.RuleSuggestion {
						leftEnumKey, ok := getEnumKeyForLiteral(
							ctx.SourceFile,
							utils.GetEnumLiterals(leftType),
							expr.Right,
							staticEvaluator,
						)
						if ok {
							return []rule.RuleSuggestion{{
								Message:  buildReplaceValueWithEnumMessage(),
								FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, ast.SkipParentheses(expr.Right), leftEnumKey)},
							}}
						}

						rightEnumKey, ok := getEnumKeyForLiteral(
							ctx.SourceFile,
							utils.GetEnumLiterals(rightType),
							expr.Left,
							staticEvaluator,
						)
						if ok {
							return []rule.RuleSuggestion{{
								Message:  buildReplaceValueWithEnumMessage(),
								FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, ast.SkipParentheses(expr.Left), rightEnumKey)},
							}}
						}

						return nil
					})
				}
			},

			ast.KindCaseClause: func(node *ast.Node) {
				leftType := ctx.TypeChecker.GetTypeAtLocation(node.Parent.Parent.Expression())
				rightType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())

				if isMismatchedComparison(leftType, rightType) {
					ctx.ReportNode(node, buildMismatchedCaseMessage())
				}
			},
		}
	},
})
