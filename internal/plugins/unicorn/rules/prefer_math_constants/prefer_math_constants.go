// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package prefer_math_constants

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var approximations = sync.OnceValue(buildApproximations)

var PreferMathConstantsRule = rule.Rule{
	Name:   "unicorn/prefer-math-constants",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		constants := approximations()
		source := ctx.SourceFile.Text()
		return rule.RuleListeners{ast.KindNumericLiteral: func(node *ast.Node) {
			flags := node.AsNumericLiteral().TokenFlags
			if flags&(ast.TokenFlagsScientific|ast.TokenFlagsWithSpecifier|ast.TokenFlagsIsInvalid) != 0 {
				return
			}
			span := node.Loc.WithPos(scanner.GetTokenPosOfNode(node, ctx.SourceFile, false))
			raw := source[span.Pos():span.End()]
			text := raw
			if flags&ast.TokenFlagsContainsSeparator != 0 {
				text = strings.ReplaceAll(text, "_", "")
			}
			leadingDot := strings.HasPrefix(text, ".")
			if leadingDot {
				text = "0" + text
			}
			constant := constants[text]
			if constant == "" || isLiteralOnly(node) {
				return
			}
			data := map[string]string{"constant": constant, "value": raw}
			ctx.ReportRangeWithDeferredSuggestions(span, rule.RuleMessage{
				Id:          "prefer-math-constants/error",
				Description: "Prefer `" + constant + "` over the approximate value `" + raw + "`.",
				Data:        data,
			}, func() []rule.RuleSuggestion {
				// Keep the diagnostic if the replacement would reference a different
				// Math object or violate a TypeScript literal/constant requirement.
				if !ctx.Globals.Access("Math").IsDeclared() ||
					!ctx.Refs.IsGlobalNameReference(node, "Math", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) ||
					requiresLiteralOrConstant(node) {
					return nil
				}
				replacement := constant
				if leadingDot {
					// A spelling such as return.6931 can touch a preceding keyword.
					replacement = utils.SafeReplacementText(ctx.SourceFile, node, replacement)
				}
				return []rule.RuleSuggestion{{
					Message: rule.RuleMessage{
						Id:          "prefer-math-constants/suggestion",
						Description: "Replace `" + raw + "` with `" + constant + "`.",
						Data:        data,
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(span, replacement)},
				}}
			})
		}}
	},
}

func isLiteralOnly(node *ast.Node) bool {
	parent := node.Parent
	if parent.PropertyNameOrName() == node {
		return true
	}
	if ast.IsPrefixUnaryExpression(parent) {
		parent = parent.Parent
	}
	return ast.IsLiteralTypeNode(parent)
}

func requiresLiteralOrConstant(node *ast.Node) bool {
	if utils.IsInAmbientContext(node) {
		return true
	}
	if member := ast.FindAncestor(node.Parent, ast.IsEnumMember); member != nil {
		declaration := member.Parent
		if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) {
			return true
		}
		// An implicit enum value must follow a member with a constant initializer.
		members := declaration.AsEnumDeclaration().Members.Nodes
		next := slices.Index(members, member) + 1
		if next < len(members) && members[next].AsEnumMember().Initializer == nil {
			return true
		}
	}
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if ast.IsPrefixUnaryExpression(parent) && parent.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
		parent = ast.WalkUpParenthesizedExpressions(parent.Parent)
	}
	if ast.IsAssertionExpression(parent) && ast.IsConstTypeReference(ast.SkipTypeParentheses(parent.Type())) {
		return true
	}
	return ast.IsComputedPropertyName(parent) &&
		(ast.IsInterfaceDeclaration(parent.Parent.Parent) || ast.IsTypeLiteralNode(parent.Parent.Parent))
}

func buildApproximations() map[string]string {
	constants := [...]struct {
		name          string
		value         float64
		minimumDigits int
	}{
		{"Math.PI", math.Pi, 2},
		{"Math.E", math.E, 3},
		{"Math.LN2", math.Ln2, 4},
		{"Math.LN10", math.Ln10, 4},
		{"Math.LOG2E", math.Log2E, 4},
		{"Math.LOG10E", math.Log10E, 4},
		{"Math.SQRT2", math.Sqrt2, 4},
		{"Math.SQRT1_2", math.Sqrt2 / 2, 4},
	}
	matches := make(map[string]string)
	for _, constant := range constants {
		exact := ecmascript.NumberToString(constant.value)
		fractionStart := strings.IndexByte(exact, '.') + 1
		matches[exact] = constant.name
		for digits := constant.minimumDigits; digits < len(exact)-fractionStart; digits++ {
			matches[exact[:fractionStart+digits]] = constant.name
			// These constants cannot fall exactly halfway at any precision used here.
			matches[strconv.FormatFloat(constant.value, 'f', digits, 64)] = constant.name
		}
	}
	return matches
}
