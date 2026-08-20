package no_duplicate_enum_values

import (
	"math"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

type enumValueKind uint8

const (
	enumValueString enumValueKind = 1 << iota
	enumValueNumber
	enumValueNegativeNumber
	enumValueNegativeZero
)

type enumValue struct {
	text string
	kind enumValueKind
}

const inlineEnumValueCount = 8

// Most enums are small. Keep their values inline so checking them does not
// allocate a map; larger enums promote the same state to overflow.
type seenEnumValues struct {
	inline      [inlineEnumValueCount]enumValue
	inlineCount int
	valueCount  int
	overflow    map[string]enumValueKind
}

func (seen *seenEnumValues) add(value enumValue, capacity int) bool {
	seen.valueCount++
	if seen.overflow != nil {
		seenKinds := seen.overflow[value.text]
		if seenKinds&value.kind != 0 {
			return true
		}
		seen.overflow[value.text] = seenKinds | value.kind
		return false
	}

	for index := range seen.inlineCount {
		if seen.inline[index].text != value.text {
			continue
		}
		if seen.inline[index].kind&value.kind != 0 {
			return true
		}
		seen.inline[index].kind |= value.kind
		return false
	}

	if seen.inlineCount < len(seen.inline) {
		seen.inline[seen.inlineCount] = value
		seen.inlineCount++
		return false
	}

	uniqueCount := seen.inlineCount + 1
	// Preserve the full-capacity fast path for unique enums, but scale the hint
	// down when the values seen so far contain duplicates.
	mapCapacity := capacity * uniqueCount / seen.valueCount
	if mapCapacity < uniqueCount {
		mapCapacity = uniqueCount
	}
	seen.overflow = make(map[string]enumValueKind, mapCapacity)
	for _, inlineValue := range seen.inline {
		seen.overflow[inlineValue.text] = inlineValue.kind
	}
	seen.overflow[value.text] = value.kind
	return false
}

func getEnumValue(initializer *ast.Node) (enumValue, bool) {
	if initializer.Kind == ast.KindParenthesizedExpression {
		initializer = ast.SkipParentheses(initializer)
	}
	switch initializer.Kind {
	case ast.KindNumericLiteral:
		return enumValue{text: initializer.AsNumericLiteral().Text, kind: enumValueNumber}, true
	case ast.KindStringLiteral:
		return enumValue{text: initializer.AsStringLiteral().Text, kind: enumValueString}, true
	case ast.KindNoSubstitutionTemplateLiteral:
		template := initializer.AsNoSubstitutionTemplateLiteral()
		if template.TemplateFlags&ast.TokenFlagsContainsInvalidEscape != 0 {
			return enumValue{}, false
		}
		return enumValue{text: template.Text, kind: enumValueString}, true
	case ast.KindPrefixUnaryExpression:
		unaryExpression := initializer.AsPrefixUnaryExpression()
		if unaryExpression.Operator != ast.KindMinusToken && unaryExpression.Operator != ast.KindPlusToken {
			return enumValue{}, false
		}
		value, ok := getEnumValue(unaryExpression.Operand)
		if !ok {
			return enumValue{}, false
		}
		return applyUnaryOperator(value, unaryExpression.Operator)
	default:
		return enumValue{}, false
	}
}

func applyUnaryOperator(value enumValue, operator ast.Kind) (enumValue, bool) {
	if value.kind == enumValueString {
		number, ok := ecmascript.StringToNumber(value.text)
		if !ok {
			return enumValue{}, false
		}
		if operator == ast.KindMinusToken {
			number = -number
		}
		return numberEnumValue(number), true
	}

	if operator == ast.KindPlusToken {
		return value, true
	}
	if value.kind == enumValueNegativeZero {
		return enumValue{text: "0", kind: enumValueNumber}, true
	}
	if value.kind == enumValueNegativeNumber {
		return enumValue{text: value.text, kind: enumValueNumber}, true
	}
	if value.text == "0" {
		return enumValue{text: "0", kind: enumValueNegativeZero}, true
	}
	return enumValue{text: value.text, kind: enumValueNegativeNumber}, true
}

func numberEnumValue(number float64) enumValue {
	if number == 0 && math.Signbit(number) {
		return enumValue{text: "0", kind: enumValueNegativeZero}
	}
	if number < 0 {
		return enumValue{text: ecmascript.NumberToString(-number), kind: enumValueNegativeNumber}
	}
	return enumValue{text: ecmascript.NumberToString(number), kind: enumValueNumber}
}

func duplicateValueMessage(value enumValue) rule.RuleMessage {
	if value.kind == enumValueNegativeNumber {
		return rule.RuleMessage{Id: "duplicateValue", Description: "Duplicate enum member value -" + value.text + "."}
	}
	return rule.RuleMessage{Id: "duplicateValue", Description: "Duplicate enum member value " + value.text + "."}
}

var NoDuplicateEnumValuesRule = rule.CreateRule(rule.Rule{
	Name:   "no-duplicate-enum-values",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				var seenValues seenEnumValues

				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil || member.Initializer == nil {
						continue
					}

					initializer := member.Initializer
					value, ok := getEnumValue(initializer)
					if !ok {
						continue
					}

					if seenValues.add(value, len(enumDecl.Members.Nodes)) {
						ctx.ReportNode(memberNode, duplicateValueMessage(value))
					}
				}
			},
		}
	},
})
