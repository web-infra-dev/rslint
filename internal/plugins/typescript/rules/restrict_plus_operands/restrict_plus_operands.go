package restrict_plus_operands

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed restrict_plus_operands.schema.json
var schemaJSON []byte

func buildBigintAndNumberMessage(left, right string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bigintAndNumber",
		Description: "Numeric '+' operations must either be both bigints or both numbers. Got `" + left + "` + `" + right + "`.",
	}
}
func buildInvalidMessage(stringLike, t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalid",
		Description: "Invalid operand for a '+' operation. Operands must each be a number or " + stringLike + ". Got `" + t + "`.",
	}
}
func buildMismatchedMessage(stringLike, left, right string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mismatched",
		Description: "Operands of '+' operations must be a number or " + stringLike + ". Got `" + left + "` + `" + right + "`.",
	}
}

type RestrictPlusOperandsOptions struct {
	AllowAny                bool
	AllowBoolean            bool
	AllowNullish            bool
	AllowNumberAndString    bool
	AllowRegExp             bool
	SkipCompoundAssignments bool
}

func isDeeplyObjectType(t *checker.Type) bool {
	if utils.IsIntersectionType(t) {
		return utils.Every(t.Types(), utils.IsObjectType)
	}
	return utils.IsObjectType(t)
}

func invalidObjectOperandType(
	typeChecker *checker.Checker,
	part *checker.Type,
	otherType *checker.Type,
	allowAny bool,
	allowRegExp bool,
) (string, bool) {
	isAny := utils.IsTypeAnyType(part)
	isTypeParameter := utils.IsTypeParameter(part)
	deeplyObject := isDeeplyObjectType(part)
	if !isTypeParameter && !deeplyObject && (allowAny || !isAny) {
		return "", false
	}

	typeName := utils.GetTypeName(typeChecker, part)
	renderedType := typeName
	if isTypeParameter {
		renderedType = typeChecker.TypeToString(part)
	}
	if typeName == "RegExp" {
		if allowRegExp && !utils.IsTypeFlagSet(otherType, checker.TypeFlagsNumberLike) {
			return "", false
		}
		return renderedType, true
	}
	if (!allowAny && isAny) || deeplyObject {
		return renderedType, true
	}
	return "", false
}

func parseOptions(options []any) RestrictPlusOperandsOptions {
	opts := RestrictPlusOperandsOptions{
		AllowAny:             true,
		AllowBoolean:         true,
		AllowNullish:         true,
		AllowNumberAndString: true,
		AllowRegExp:          true,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["allowAny"].(bool); ok {
		opts.AllowAny = value
	}
	if value, ok := optsMap["allowBoolean"].(bool); ok {
		opts.AllowBoolean = value
	}
	if value, ok := optsMap["allowNullish"].(bool); ok {
		opts.AllowNullish = value
	}
	if value, ok := optsMap["allowNumberAndString"].(bool); ok {
		opts.AllowNumberAndString = value
	}
	if value, ok := optsMap["allowRegExp"].(bool); ok {
		opts.AllowRegExp = value
	}
	if value, ok := optsMap["skipCompoundAssignments"].(bool); ok {
		opts.SkipCompoundAssignments = value
	}
	return opts
}

var RestrictPlusOperandsRule = rule.CreateRule(rule.Rule{
	Name:             "restrict-plus-operands",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		var stringLikes [5]string
		stringLikeCount := 0
		if opts.AllowAny {
			stringLikes[stringLikeCount] = "`any`"
			stringLikeCount++
		}
		if opts.AllowBoolean {
			stringLikes[stringLikeCount] = "`boolean`"
			stringLikeCount++
		}
		if opts.AllowNullish {
			stringLikes[stringLikeCount] = "`null`"
			stringLikeCount++
		}
		if opts.AllowRegExp {
			stringLikes[stringLikeCount] = "`RegExp`"
			stringLikeCount++
		}
		if opts.AllowNullish {
			stringLikes[stringLikeCount] = "`undefined`"
			stringLikeCount++
		}
		var stringLike string
		switch stringLikeCount {
		case 0:
			stringLike = "string"
		case 1:
			stringLike = "string, allowing a string + " + stringLikes[0]
		default:
			stringLike = "string, allowing a string + any of: " + strings.Join(stringLikes[:stringLikeCount], ", ")
		}

		getTypeConstrained := func(node *ast.Node) *checker.Type {
			return checker.Checker_getBaseTypeOfLiteralType(ctx.TypeChecker, utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node))
		}

		invalidFlags := checker.TypeFlagsESSymbolLike |
			checker.TypeFlagsNever |
			checker.TypeFlagsUnknown
		if !opts.AllowAny {
			invalidFlags |= checker.TypeFlagsAny
		}
		if !opts.AllowBoolean {
			invalidFlags |= checker.TypeFlagsBooleanLike
		}
		if !opts.AllowNullish {
			invalidFlags |= checker.TypeFlagsNullable
		}

		checkInvalidPlusOperand := func(baseNode *ast.Node, baseType, otherType *checker.Type) (checker.TypeFlags, bool) {
			if !utils.IsUnionType(baseType) {
				flags := checker.Type_flags(baseType)
				if flags&invalidFlags != 0 {
					ctx.ReportNode(baseNode, buildInvalidMessage(stringLike, ctx.TypeChecker.TypeToString(baseType)))
					return flags, true
				}
				if typeName, invalid := invalidObjectOperandType(
					ctx.TypeChecker,
					baseType,
					otherType,
					opts.AllowAny,
					opts.AllowRegExp,
				); invalid {
					ctx.ReportNode(baseNode, buildInvalidMessage(stringLike, typeName))
					return flags, true
				}
				return flags, false
			}

			parts := baseType.Types()
			var flags checker.TypeFlags
			for _, part := range parts {
				flags |= checker.Type_flags(part)
			}
			if flags&invalidFlags != 0 {
				ctx.ReportNode(baseNode, buildInvalidMessage(stringLike, ctx.TypeChecker.TypeToString(baseType)))
				return flags, true
			}

			reported := false
			for _, part := range parts {
				if typeName, invalid := invalidObjectOperandType(
					ctx.TypeChecker,
					part,
					otherType,
					opts.AllowAny,
					opts.AllowRegExp,
				); invalid {
					ctx.ReportNode(baseNode, buildInvalidMessage(stringLike, typeName))
					reported = true
				}
			}
			return flags, reported
		}

		checkPlusOperands := func(
			node *ast.BinaryExpression,
		) {
			leftType := getTypeConstrained(node.Left)
			rightType := getTypeConstrained(node.Right)

			if leftType == rightType &&
				utils.IsTypeFlagSet(
					leftType,
					checker.TypeFlagsBigIntLike|
						checker.TypeFlagsNumberLike|
						checker.TypeFlagsStringLike,
				) {
				return
			}

			leftNode := node.Left
			if leftNode.Kind == ast.KindParenthesizedExpression {
				leftNode = ast.SkipParentheses(leftNode)
			}
			rightNode := node.Right
			if rightNode.Kind == ast.KindParenthesizedExpression {
				rightNode = ast.SkipParentheses(rightNode)
			}
			leftTypeFlags, leftInvalid := checkInvalidPlusOperand(leftNode, leftType, rightType)
			rightTypeFlags, rightInvalid := checkInvalidPlusOperand(rightNode, rightType, leftType)
			if leftInvalid || rightInvalid {
				return
			}

			if !opts.AllowNumberAndString &&
				(leftTypeFlags&checker.TypeFlagsStringLike != 0 &&
					rightTypeFlags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) != 0 ||
					rightTypeFlags&checker.TypeFlagsStringLike != 0 &&
						leftTypeFlags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) != 0) {
				ctx.ReportNode(&node.Node, buildMismatchedMessage(stringLike, ctx.TypeChecker.TypeToString(leftType), ctx.TypeChecker.TypeToString(rightType)))
				return
			}

			if leftTypeFlags&checker.TypeFlagsNumberLike != 0 && rightTypeFlags&checker.TypeFlagsBigIntLike != 0 ||
				rightTypeFlags&checker.TypeFlagsNumberLike != 0 && leftTypeFlags&checker.TypeFlagsBigIntLike != 0 {
				ctx.ReportNode(&node.Node, buildBigintAndNumberMessage(ctx.TypeChecker.TypeToString(leftType), ctx.TypeChecker.TypeToString(rightType)))
			}
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				if expr.OperatorToken.Kind == ast.KindPlusToken || (!opts.SkipCompoundAssignments && expr.OperatorToken.Kind == ast.KindPlusEqualsToken) {
					checkPlusOperands(expr)
				}
			},
		}
	},
})
