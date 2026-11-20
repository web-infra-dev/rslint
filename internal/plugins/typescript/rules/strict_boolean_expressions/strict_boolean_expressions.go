// Package strict_boolean_expressions ports typescript-eslint's
// `@typescript-eslint/strict-boolean-expressions` rule to rslint.
//
// The rule restricts what values may appear in a boolean position (test
// expressions, `!` arguments, `&&` / `||` operands, truthiness-asserted call
// arguments, and the return value of array predicate callbacks). The set of
// "variant" types accumulated from a value's union constituents is matched
// against the user's `allow*` options to decide whether to report and which
// `messageId` / suggestion fixes to emit.
package strict_boolean_expressions

import (
	_ "embed"
	"fmt"
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed strict_boolean_expressions.schema.json
var schemaJSON []byte

type resolvedOptions struct {
	allowAny                                               bool
	allowNullableBoolean                                   bool
	allowNullableEnum                                      bool
	allowNullableNumber                                    bool
	allowNullableObject                                    bool
	allowNullableString                                    bool
	allowNumber                                            bool
	allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing bool
	allowString                                            bool
}

func parseOptions(options []any) resolvedOptions {
	// Defaults match upstream:
	//   allowString=true, allowNumber=true, allowNullableObject=true,
	//   all other booleans default to false.
	opts := resolvedOptions{
		allowString:         true,
		allowNumber:         true,
		allowNullableObject: true,
	}

	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)

	if value, ok := optsMap["allowAny"].(bool); ok {
		opts.allowAny = value
	}
	if value, ok := optsMap["allowNullableBoolean"].(bool); ok {
		opts.allowNullableBoolean = value
	}
	if value, ok := optsMap["allowNullableEnum"].(bool); ok {
		opts.allowNullableEnum = value
	}
	if value, ok := optsMap["allowNullableNumber"].(bool); ok {
		opts.allowNullableNumber = value
	}
	if value, ok := optsMap["allowNullableObject"].(bool); ok {
		opts.allowNullableObject = value
	}
	if value, ok := optsMap["allowNullableString"].(bool); ok {
		opts.allowNullableString = value
	}
	if value, ok := optsMap["allowNumber"].(bool); ok {
		opts.allowNumber = value
	}
	if value, ok := optsMap["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
		opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = value
	}
	if value, ok := optsMap["allowString"].(bool); ok {
		opts.allowString = value
	}
	return opts
}

// VariantType labels a single "variant" we collect from a type's union
// constituents. The exact set names match upstream's `VariantType` union so
// determineReportType's combinatorial logic stays readable when cross-checked.
type variantType string

const (
	vtAny           variantType = "any"
	vtBoolean       variantType = "boolean"
	vtEnum          variantType = "enum"
	vtNever         variantType = "never"
	vtNullish       variantType = "nullish"
	vtNumber        variantType = "number"
	vtObject        variantType = "object"
	vtString        variantType = "string"
	vtTruthyBoolean variantType = "truthy boolean"
	vtTruthyNumber  variantType = "truthy number"
	vtTruthyString  variantType = "truthy string"
)

type variantSet map[variantType]struct{}

func newVariantSet() variantSet             { return variantSet{} }
func (s variantSet) add(v variantType)      { s[v] = struct{}{} }
func (s variantSet) has(v variantType) bool { _, ok := s[v]; return ok }
func (s variantSet) size() int              { return len(s) }

// ----- Messages ---------------------------------------------------------

func msgConditionErrorAny(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorAny",
		Description: fmt.Sprintf("Unexpected any value in %s. An explicit comparison or type conversion is required.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorNullableBoolean(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableBoolean",
		Description: fmt.Sprintf("Unexpected nullable boolean value in %s. Please handle the nullish case explicitly.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorNullableEnum(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableEnum",
		Description: fmt.Sprintf("Unexpected nullable enum value in %s. Please handle the nullish/zero/NaN cases explicitly.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorNullableNumber(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableNumber",
		Description: fmt.Sprintf("Unexpected nullable number value in %s. Please handle the nullish/zero/NaN cases explicitly.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorNullableObject(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableObject",
		Description: fmt.Sprintf("Unexpected nullable object value in %s. An explicit null check is required.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorNullableString(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullableString",
		Description: fmt.Sprintf("Unexpected nullable string value in %s. Please handle the nullish/empty cases explicitly.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorNullish() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNullish",
		Description: "Unexpected nullish value in conditional. The condition is always false.",
	}
}
func msgConditionErrorNumber(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorNumber",
		Description: fmt.Sprintf("Unexpected number value in %s. An explicit zero/NaN check is required.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorObject(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorObject",
		Description: fmt.Sprintf("Unexpected object value in %s. The condition is always true.", context),
		Data:        map[string]string{"context": context},
	}
}
func msgConditionErrorOther() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorOther",
		Description: "Unexpected value in conditional. A boolean expression is required.",
	}
}
func msgConditionErrorString(context string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditionErrorString",
		Description: fmt.Sprintf("Unexpected string value in %s. An explicit empty string check is required.", context),
		Data:        map[string]string{"context": context},
	}
}

func msgNoStrictNullCheck() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func msgPredicateCannotBeAsync() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "predicateCannotBeAsync",
		Description: "Predicate function should not be 'async'; expected a boolean return type.",
	}
}

func msgExplicitBooleanReturnType() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "explicitBooleanReturnType",
		Description: "Add an explicit `boolean` return type annotation.",
	}
}

// Suggestion-fix message builders.
func sugCastBoolean() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCastBoolean", Description: "Explicitly convert value to a boolean (`Boolean(value)`)"}
}
func sugCompareArrayLengthNonzero() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareArrayLengthNonzero", Description: "Change condition to check array's length (`value.length > 0`)"}
}
func sugCompareArrayLengthZero() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareArrayLengthZero", Description: "Change condition to check array's length (`value.length === 0`)"}
}
func sugCompareEmptyString() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareEmptyString", Description: "Change condition to check for empty string (`value !== \"\"`)"}
}
func sugCompareFalse() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareFalse", Description: "Change condition to check if false (`value === false`)"}
}
func sugCompareNaN() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareNaN", Description: "Change condition to check for NaN (`!Number.isNaN(value)`)"}
}
func sugCompareNullish() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareNullish", Description: "Change condition to check for null/undefined (`value != null`)"}
}
func sugCompareStringLength() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareStringLength", Description: "Change condition to check string's length (`value.length !== 0`)"}
}
func sugCompareTrue() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareTrue", Description: "Change condition to check if true (`value === true`)"}
}
func sugCompareZero() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixCompareZero", Description: "Change condition to check for 0 (`value !== 0`)"}
}
func sugDefaultEmptyString() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixDefaultEmptyString", Description: "Explicitly treat nullish value the same as an empty string (`value ?? \"\"`)"}
}
func sugDefaultFalse() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixDefaultFalse", Description: "Explicitly treat nullish value the same as false (`value ?? false`)"}
}
func sugDefaultZero() rule.RuleMessage {
	return rule.RuleMessage{Id: "conditionFixDefaultZero", Description: "Explicitly treat nullish value the same as 0 (`value ?? 0`)"}
}

// ----- Rule definition ---------------------------------------------------

var StrictBooleanExpressionsRule = rule.CreateRule(rule.Rule{
	Name:             "strict-boolean-expressions",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		tc := ctx.TypeChecker
		sf := ctx.SourceFile
		staticEvaluator := utils.NewStaticStringEvaluatorWithCallEscapes(tc, sf, ctx.Refs)

		// strictNullChecks gate. Matches upstream: when strictNullChecks is
		// off and the user has not opted in via the long-named "I know what
		// I am doing" flag, the rule reports a single noStrictNullCheck
		// diagnostic anchored at the start of the file and still walks the
		// rest of the file.
		compilerOptions := ctx.Program().Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)
		if !isStrictNullChecks && !opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), msgNoStrictNullCheck())
		}

		// traversedNodes dedupes traverseNode entries so the listener for an
		// inner LogicalExpression doesn't re-check operands that the outer
		// `if`/`while`/... listener already processed.
		traversed := map[*ast.Node]struct{}{}

		var traverseNode func(node *ast.Node, isCondition bool)
		var traverseLogical func(node *ast.Node, isCondition bool)
		var checkNode func(node *ast.Node)

		traverseLogical = func(node *ast.Node, isCondition bool) {
			bin := node.AsBinaryExpression()
			if bin == nil {
				return
			}
			// Left operand is always a condition.
			traverseNode(bin.Left, true)
			// Right operand is a condition only when the logical expression
			// itself sits in a condition; otherwise it is used for its side
			// effects and may legally be any value.
			traverseNode(bin.Right, isCondition)
		}

		traverseNode = func(node *ast.Node, isCondition bool) {
			if node == nil {
				return
			}
			node = ast.SkipParentheses(node)
			if _, seen := traversed[node]; seen {
				return
			}
			traversed[node] = struct{}{}

			// tsgo models parens as explicit ParenthesizedExpression nodes
			// that ESTree drops at parse time. Unwrap here so downstream code
			// — listener dispatch AND text-based suggestion fixes — observes
			// the same shape upstream sees.
			inner := ast.SkipParentheses(node)

			if ast.IsBinaryExpression(inner) {
				op := inner.AsBinaryExpression().OperatorToken.Kind
				if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
					traverseLogical(inner, isCondition)
					return
				}
			}

			if !isCondition {
				return
			}
			checkNode(inner)
		}

		checkNode = func(node *ast.Node) {
			nodeType := utils.GetConstrainedTypeAtLocation(tc, node)
			parts := utils.UnionTypeParts(nodeType)
			variants := inspectVariantTypes(tc, parts)
			reportID := determineReportType(variants, opts)
			if reportID == "" {
				return
			}
			reportConditional(ctx, sf, tc, node, reportID, opts)
		}

		traverseTest := func(test *ast.Node) {
			if test == nil {
				return
			}
			traverseNode(test, true)
		}

		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				traverseTest(node.AsIfStatement().Expression)
			},
			ast.KindWhileStatement: func(node *ast.Node) {
				traverseTest(node.AsWhileStatement().Expression)
			},
			ast.KindDoStatement: func(node *ast.Node) {
				traverseTest(node.AsDoStatement().Expression)
			},
			ast.KindForStatement: func(node *ast.Node) {
				traverseTest(node.AsForStatement().Condition)
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				traverseTest(node.AsConditionalExpression().Condition)
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				u := node.AsPrefixUnaryExpression()
				if u == nil || u.Operator != ast.KindExclamationToken {
					return
				}
				traverseNode(u.Operand, true)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				op := bin.OperatorToken.Kind
				if op != ast.KindAmpersandAmpersandToken && op != ast.KindBarBarToken {
					return
				}
				// Top-level entry into the logical chain: by default the
				// expression is NOT a condition itself. Inner LogicalExpression
				// listeners will be deduped by traversed map.
				if _, seen := traversed[node]; seen {
					return
				}
				traversed[node] = struct{}{}
				traverseLogical(node, false)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}
				// (1) truthiness assertion: traverse the asserted argument
				// as if it were a condition.
				if assertedArg := typescriptutil.FindTruthinessAssertedArgument(tc, callExpr); assertedArg != nil {
					traverseNode(assertedArg, true)
				}
				// (2) array predicate: inspect the predicate function's return
				// type against the same VariantType matrix.
				if utils.IsArrayMethodCallWithPredicate(tc, callExpr, staticEvaluator) && len(callExpr.Arguments.Nodes) > 0 {
					predicate, predicateType := firstArrayPredicateArgument(ctx, callExpr.Arguments.Nodes)
					if predicate != nil {
						checkArrayMethodPredicate(ctx, sf, tc, predicate, predicateType, opts)
					}
				}
			},
		}
	},
})

// ----- Variant inspection -----------------------------------------------

// inspectVariantTypes partitions the union constituents of a type into the
// VariantType set determineReportType cares about. Mirrors upstream's
// `inspectVariantTypes` modulo flag-name differences in tsgo.
func inspectVariantTypes(tc *checker.Checker, types []*checker.Type) variantSet {
	out := newVariantSet()

	if anyOf(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoidLike)
	}) {
		out.add(vtNullish)
	}

	booleans := filterTypes(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsBooleanLike)
	})
	// "boolean" decomposes into `true | false` in tsgo's union constituents
	// (same as ts-api-utils `unionConstituents`). One literal → narrow it;
	// two literals → original was the full `boolean` type.
	if len(booleans) == 1 {
		if utils.IsTrueLiteralType(booleans[0]) {
			out.add(vtTruthyBoolean)
		} else {
			out.add(vtBoolean)
		}
	} else if len(booleans) >= 2 {
		out.add(vtBoolean)
	}

	strings := filterTypes(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
	})
	if len(strings) > 0 {
		if everyType(strings, func(t *checker.Type) bool {
			if !utils.IsTypeFlagSet(t, checker.TypeFlagsStringLiteral) {
				return false
			}
			if s, ok := t.AsLiteralType().Value().(string); ok {
				return s != ""
			}
			return false
		}) {
			out.add(vtTruthyString)
		} else {
			out.add(vtString)
		}
	}

	numbers := filterTypes(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike)
	})
	if len(numbers) > 0 {
		if everyType(numbers, func(t *checker.Type) bool {
			if !utils.IsTypeFlagSet(t, checker.TypeFlagsNumberLiteral) {
				return false
			}
			val := t.AsLiteralType().Value()
			return !utils.IsNumberLiteralZeroOrNaN(val)
		}) {
			out.add(vtTruthyNumber)
		} else {
			out.add(vtNumber)
		}
	}

	if anyOf(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsEnumLike)
	}) {
		out.add(vtEnum)
	}

	// "object" bucket: anything that is none of the primitive/nullish/never/
	// type-parameter/any/unknown buckets. A branded boolean (e.g.
	// `boolean & { __brand: 'X' }`) is conventionally treated as `boolean`,
	// not `object`.
	if anyOf(types, func(t *checker.Type) bool {
		return !utils.IsTypeFlagSet(t, checker.TypeFlagsNull|
			checker.TypeFlagsUndefined|
			checker.TypeFlagsVoidLike|
			checker.TypeFlagsBooleanLike|
			checker.TypeFlagsStringLike|
			checker.TypeFlagsNumberLike|
			checker.TypeFlagsBigIntLike|
			checker.TypeFlagsTypeParameter|
			checker.TypeFlagsAny|
			checker.TypeFlagsUnknown|
			checker.TypeFlagsNever)
	}) {
		if anyOf(types, isBrandedBoolean) {
			out.add(vtBoolean)
		} else {
			out.add(vtObject)
		}
	}

	if anyOf(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsTypeParameter|checker.TypeFlagsAny|checker.TypeFlagsUnknown)
	}) {
		out.add(vtAny)
	}

	if anyOf(types, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsNever)
	}) {
		out.add(vtNever)
	}

	return out
}

// determineReportType maps a VariantType set to the messageId we should report,
// honoring the user's `allow*` options. Returns "" when the value is acceptable.
// The branch order mirrors upstream so audits map 1:1.
func determineReportType(types variantSet, opts resolvedOptions) string {
	is := func(wanted ...variantType) bool {
		if types.size() != len(wanted) {
			return false
		}
		for _, w := range wanted {
			if !types.has(w) {
				return false
			}
		}
		return true
	}

	switch {
	case is(vtBoolean), is(vtTruthyBoolean):
		return ""
	case is(vtNever):
		return ""
	case is(vtNullish):
		return "conditionErrorNullish"
	case is(vtNullish, vtTruthyBoolean):
		return ""
	case is(vtNullish, vtBoolean):
		if !opts.allowNullableBoolean {
			return "conditionErrorNullableBoolean"
		}
		return ""
	}

	// Truthy-primitive + nullish: only valid when the corresponding `allow*`
	// option is on (the nullish branch is always false, so the expression
	// reduces to a truthy value).
	if opts.allowNumber && is(vtNullish, vtTruthyNumber) {
		return ""
	}
	if opts.allowString && is(vtNullish, vtTruthyString) {
		return ""
	}

	switch {
	case is(vtString), is(vtTruthyString):
		if !opts.allowString {
			return "conditionErrorString"
		}
		return ""
	case is(vtNullish, vtString):
		if !opts.allowNullableString {
			return "conditionErrorNullableString"
		}
		return ""
	case is(vtNumber), is(vtTruthyNumber):
		if !opts.allowNumber {
			return "conditionErrorNumber"
		}
		return ""
	case is(vtNullish, vtNumber):
		if !opts.allowNullableNumber {
			return "conditionErrorNullableNumber"
		}
		return ""
	case is(vtObject):
		return "conditionErrorObject"
	case is(vtNullish, vtObject):
		if !opts.allowNullableObject {
			return "conditionErrorNullableObject"
		}
		return ""
	}

	// Nullable enum variants — including the mixed-enum shapes where the enum
	// members straddle string/number literal types.
	switch {
	case is(vtNullish, vtNumber, vtEnum),
		is(vtNullish, vtString, vtEnum),
		is(vtNullish, vtTruthyNumber, vtEnum),
		is(vtNullish, vtTruthyString, vtEnum),
		is(vtNullish, vtTruthyNumber, vtTruthyString, vtEnum),
		is(vtNullish, vtTruthyNumber, vtString, vtEnum),
		is(vtNullish, vtTruthyString, vtNumber, vtEnum),
		is(vtNullish, vtNumber, vtString, vtEnum):
		if !opts.allowNullableEnum {
			return "conditionErrorNullableEnum"
		}
		return ""
	case is(vtAny):
		if !opts.allowAny {
			return "conditionErrorAny"
		}
		return ""
	}

	return "conditionErrorOther"
}

// ----- Reporting --------------------------------------------------------

// reportConditional reports a `messageId` against `node` together with the
// appropriate suggestion fixes. `node` is the inner (paren-stripped) operand,
// not the surrounding `!` or test context.
func reportConditional(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	node *ast.Node,
	messageID string,
	opts resolvedOptions,
) {
	context := "conditional"
	var msg rule.RuleMessage
	switch messageID {
	case "conditionErrorAny":
		msg = msgConditionErrorAny(context)
	case "conditionErrorNullableBoolean":
		msg = msgConditionErrorNullableBoolean(context)
	case "conditionErrorNullableEnum":
		msg = msgConditionErrorNullableEnum(context)
	case "conditionErrorNullableNumber":
		msg = msgConditionErrorNullableNumber(context)
	case "conditionErrorNullableObject":
		msg = msgConditionErrorNullableObject(context)
	case "conditionErrorNullableString":
		msg = msgConditionErrorNullableString(context)
	case "conditionErrorNullish":
		msg = msgConditionErrorNullish()
	case "conditionErrorNumber":
		msg = msgConditionErrorNumber(context)
	case "conditionErrorObject":
		msg = msgConditionErrorObject(context)
	case "conditionErrorOther":
		msg = msgConditionErrorOther()
	case "conditionErrorString":
		msg = msgConditionErrorString(context)
	default:
		return
	}

	sugs := getSuggestionsForConditionError(sf, tc, node, messageID, opts)
	if len(sugs) == 0 {
		ctx.ReportNode(node, msg)
		return
	}
	ctx.ReportNodeWithSuggestions(node, msg, sugs...)
}

// enclosingNegation returns the outermost `!` UnaryExpression that wraps
// `node`, transparently walking past any ParenthesizedExpression layers.
// Returns nil when `node` is not the operand of a logical negation.
//
// tsgo preserves ParenthesizedExpression as an explicit AST node whereas
// ESTree drops it at parse time. Upstream's `isLogicalNegationExpression`
// runs on ESTree, where `!(x)` has `x.parent === UnaryExpression` directly.
// To match that semantics we walk through paren wrappers before deciding.
//
// This also returns the actual `!` node that the wrapping fixer should
// target for replacement — for `!(x)` the negation fix should replace
// the whole `!(x)`, not just the inner `(x)`.
func enclosingNegation(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return nil
	}
	if !ast.IsPrefixUnaryExpression(parent) {
		return nil
	}
	if parent.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
		return nil
	}
	return parent
}

func getSuggestionsForConditionError(
	sf *ast.SourceFile,
	tc *checker.Checker,
	node *ast.Node,
	messageID string,
	opts resolvedOptions,
) []rule.RuleSuggestion {
	// Resolve the enclosing `!` once, walking through paren wrappers tsgo
	// preserves but ESTree drops. `neg != nil` mirrors upstream's
	// `isLogicalNegationExpression(node.parent)`; when non-nil we target the
	// outermost `!` for replacement, not the immediate paren-wrapper.
	neg := enclosingNegation(node)

	switch messageID {
	case "conditionErrorAny":
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCastBoolean(), node, node, func(code string) string {
				return "Boolean(" + code + ")"
			}),
		}

	case "conditionErrorNullableBoolean":
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugDefaultFalse(), node, node, func(code string) string {
					return code + " ?? false"
				}),
				wrappingFix(sf, sugCompareFalse(), neg, node, func(code string) string {
					return code + " === false"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugDefaultFalse(), node, node, func(code string) string {
				return code + " ?? false"
			}),
			wrappingFix(sf, sugCompareTrue(), node, node, func(code string) string {
				return code + " === true"
			}),
		}

	case "conditionErrorNullableEnum":
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareNullish(), neg, node, func(code string) string {
					return code + " == null"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareNullish(), node, node, func(code string) string {
				return code + " != null"
			}),
		}

	case "conditionErrorNullableNumber":
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareNullish(), neg, node, func(code string) string {
					return code + " == null"
				}),
				wrappingFix(sf, sugDefaultZero(), node, node, func(code string) string {
					return code + " ?? 0"
				}),
				wrappingFix(sf, sugCastBoolean(), neg, node, func(code string) string {
					return "!Boolean(" + code + ")"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareNullish(), node, node, func(code string) string {
				return code + " != null"
			}),
			wrappingFix(sf, sugDefaultZero(), node, node, func(code string) string {
				return code + " ?? 0"
			}),
			wrappingFix(sf, sugCastBoolean(), node, node, func(code string) string {
				return "Boolean(" + code + ")"
			}),
		}

	case "conditionErrorNullableObject":
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareNullish(), neg, node, func(code string) string {
					return code + " == null"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareNullish(), node, node, func(code string) string {
				return code + " != null"
			}),
		}

	case "conditionErrorNullableString":
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareNullish(), neg, node, func(code string) string {
					return code + " == null"
				}),
				wrappingFix(sf, sugDefaultEmptyString(), node, node, func(code string) string {
					return code + ` ?? ""`
				}),
				wrappingFix(sf, sugCastBoolean(), neg, node, func(code string) string {
					return "!Boolean(" + code + ")"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareNullish(), node, node, func(code string) string {
				return code + " != null"
			}),
			wrappingFix(sf, sugDefaultEmptyString(), node, node, func(code string) string {
				return code + ` ?? ""`
			}),
			wrappingFix(sf, sugCastBoolean(), node, node, func(code string) string {
				return "Boolean(" + code + ")"
			}),
		}

	case "conditionErrorNumber":
		if isArrayLengthExpression(tc, node) {
			if neg != nil {
				return []rule.RuleSuggestion{
					wrappingFix(sf, sugCompareArrayLengthZero(), neg, node, func(code string) string {
						return code + " === 0"
					}),
				}
			}
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareArrayLengthNonzero(), node, node, func(code string) string {
					return code + " > 0"
				}),
			}
		}
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareZero(), neg, node, func(code string) string {
					return code + " === 0"
				}),
				wrappingFix(sf, sugCompareNaN(), neg, node, func(code string) string {
					return "Number.isNaN(" + code + ")"
				}),
				wrappingFix(sf, sugCastBoolean(), neg, node, func(code string) string {
					return "!Boolean(" + code + ")"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareZero(), node, node, func(code string) string {
				return code + " !== 0"
			}),
			wrappingFix(sf, sugCompareNaN(), node, node, func(code string) string {
				return "!Number.isNaN(" + code + ")"
			}),
			wrappingFix(sf, sugCastBoolean(), node, node, func(code string) string {
				return "Boolean(" + code + ")"
			}),
		}

	case "conditionErrorString":
		// Assertions can hide nullish values from the final expression type.
		// Default their result before accessing length,
		// preserving short-circuiting and evaluating the original expression once.
		guardNullish := hasOptionalChain(node) || assertionCanBeNullish(tc, node)
		valueCode := func(code string) string {
			if guardNullish {
				return "(" + code + ` ?? "")`
			}
			return code
		}
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareStringLength(), neg, node, func(code string) string {
					return valueCode(code) + ".length === 0"
				}),
				wrappingFix(sf, sugCompareEmptyString(), neg, node, func(code string) string {
					return valueCode(code) + ` === ""`
				}),
				wrappingFix(sf, sugCastBoolean(), neg, node, func(code string) string {
					return "!Boolean(" + code + ")"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareStringLength(), node, node, func(code string) string {
				return valueCode(code) + ".length > 0"
			}),
			wrappingFix(sf, sugCompareEmptyString(), node, node, func(code string) string {
				return valueCode(code) + ` !== ""`
			}),
			wrappingFix(sf, sugCastBoolean(), node, node, func(code string) string {
				return "Boolean(" + code + ")"
			}),
		}
	}

	// conditionErrorObject / conditionErrorNullish / conditionErrorOther are
	// reported without suggestions — upstream returns [].
	_ = opts
	return nil
}

// ----- Array predicate checking ----------------------------------------

// Inspect local result-producing expressions for erased nullish values.
// Follow conditional branches and logical results, without evaluating calls,
// resolving aliases or treating a nullable receiver as a nullable result.
func assertionCanBeNullish(tc *checker.Checker, node *ast.Node) bool {
	for node = ast.SkipParentheses(node); node != nil; node = ast.SkipParentheses(node.Expression()) {
		awaited := false
		switch node.Kind {
		case ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
		case ast.KindAwaitExpression:
			awaited = true
		case ast.KindConditionalExpression:
			conditional := node.AsConditionalExpression()
			return assertionCanBeNullish(tc, conditional.WhenTrue) || hasOptionalChain(conditional.WhenTrue) ||
				assertionCanBeNullish(tc, conditional.WhenFalse) || hasOptionalChain(conditional.WhenFalse)
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken:
				return assertionCanBeNullish(tc, binary.Left) || hasOptionalChain(binary.Left) ||
					assertionCanBeNullish(tc, binary.Right) || hasOptionalChain(binary.Right)
			case ast.KindBarBarToken, ast.KindQuestionQuestionToken, ast.KindCommaToken:
				// A nullish left operand cannot be the result of || or ??.
				return assertionCanBeNullish(tc, binary.Right) || hasOptionalChain(binary.Right)
			}
			return false
		default:
			return false
		}
		operand := ast.SkipOuterExpressions(node.Expression(), ast.OEKParentheses|ast.OEKAssertions|ast.OEKNonNullAssertions)
		operandType := utils.GetConstrainedTypeAtLocation(tc, operand)
		if awaited {
			operandType = checker.Checker_getAwaitedType(tc, operandType)
			if operandType == nil {
				return true
			}
		}
		for _, part := range utils.UnionTypeParts(operandType) {
			if utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoidLike|checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter) {
				return true
			}
		}
	}
	return false
}

// Protect every authored optional chain consistently, including assertions
// and await wrappers. Non-nullish static receiver types do not remove the
// runtime optional syntax or change how the length suggestion is constructed.
func hasOptionalChain(node *ast.Node) bool {
	const wrappers = ast.OEKNonNullAssertions | ast.OEKAssertions | ast.OEKParentheses
	for node = ast.SkipOuterExpressions(node, wrappers); node != nil; node = ast.SkipOuterExpressions(node.Expression(), wrappers) {
		optional := false
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			optional = node.AsPropertyAccessExpression().QuestionDotToken != nil
		case ast.KindElementAccessExpression:
			optional = node.AsElementAccessExpression().QuestionDotToken != nil
		case ast.KindCallExpression:
			optional = node.AsCallExpression().QuestionDotToken != nil
		case ast.KindAwaitExpression:
			// Await preserves the nullish result of a short-circuited call.
		default:
			return false
		}
		if optional {
			return true
		}
	}
	return false
}

// firstArrayPredicateArgument selects possible first arguments from static spread types.
// A possibly empty spread also admits the next argument. Keep the authored
// spread as the diagnostic node: its element types do not identify a safe
// inline function to edit.
func firstArrayPredicateArgument(ctx rule.RuleContext, arguments []*ast.Node) (*ast.Node, *checker.Type) {
	tc := ctx.TypeChecker
	var firstTypes []*checker.Type
	var reportNode *ast.Node
	for _, argument := range arguments {
		if !ast.IsSpreadElement(argument) {
			if reportNode == nil {
				reportNode = ast.SkipParentheses(argument)
			}
			firstTypes = append(firstTypes, utils.GetConstrainedTypeAtLocation(tc, argument))
			break
		}
		spreadType := utils.GetConstrainedTypeAtLocation(tc, argument.Expression())
		types, mayBeEmpty, known := firstSpreadElementTypes(ctx, spreadType, map[*checker.Type]bool{})
		if !known {
			// Keep unresolved mixed collection unions conservative instead of
			// losing tuple argument positions in an iterated-type union.
			if spreadType.IsUnion() {
				return nil, nil
			}
			// A spread node exposes the compiler's iterated element type for
			// Set/Iterable even when the collection has no numeric index.
			// Error/any/unknown types do not establish a callback type.
			element := utils.GetConstrainedTypeAtLocation(tc, argument)
			if element == nil || utils.IsTypeFlagSet(element, checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsNever) {
				return nil, nil
			}
			types, mayBeEmpty = []*checker.Type{element}, true
		}
		if len(types) != 0 && reportNode == nil {
			reportNode = argument
		}
		firstTypes = append(firstTypes, types...)
		if !mayBeEmpty {
			break
		}
	}
	if len(firstTypes) == 0 {
		return nil, nil
	}
	return reportNode, tc.GetUnionType(firstTypes)
}

// Read tuple elements in order and stop at the first guaranteed argument. For
// [...Callback[], Callback, object], the object can never be the callback, even
// though it appears in the tuple's numeric index type. Variadic elements recurse
// through their constraints, with a cycle guard for recursive generic types.
func firstSpreadElementTypes(ctx rule.RuleContext, t *checker.Type, visiting map[*checker.Type]bool) ([]*checker.Type, bool, bool) {
	tc := ctx.TypeChecker
	if t == nil || visiting[t] {
		return nil, false, false
	}
	visiting[t] = true
	defer delete(visiting, t)
	if utils.IsTypeParameter(t) {
		return firstSpreadElementTypes(ctx, checker.Checker_getBaseConstraintOfType(tc, t), visiting)
	}
	if t.IsUnion() {
		var types []*checker.Type
		mayBeEmpty := false
		for _, part := range t.Types() {
			first, empty, known := firstSpreadElementTypes(ctx, part, visiting)
			if !known {
				return nil, false, false
			}
			types = append(types, first...)
			mayBeEmpty = mayBeEmpty || empty
		}
		return types, mayBeEmpty, true
	}
	iterator := utils.GetWellKnownSymbolPropertyOfType(t, "iterator", tc)
	standardIterator := iterator == nil || hasStandardIteratorSignature(ctx, iterator)
	// User iterator declarations define the spread values independently of
	// numeric properties or tuple-like metadata. Their length is not inferred.
	if !standardIterator {
		return iterableElementTypes(tc, t)
	}
	if !checker.IsTupleType(t) {
		// A required zero property on an array-like type preserves its
		// callback and nonempty information through branding/intersections.
		// Do not infer iterator order for an arbitrary indexed object.
		arrayLike := tc.IsArrayLikeType(t)
		if zero := tc.GetPropertyOfType(t, "0"); arrayLike && zero != nil && zero.Flags&ast.SymbolFlagsOptional == 0 {
			if first := tc.GetTypeOfPropertyOfType(t, "0"); first != nil {
				return []*checker.Type{first}, false, true
			}
		}
		if first, empty, known := iterableElementTypes(tc, t); known {
			return first, empty, true
		}
		// Downlevel arrays can lack the iterator library declaration.
		if iterator == nil && arrayLike {
			if first := tc.GetNumberIndexType(t); first != nil {
				return []*checker.Type{first}, true, true
			}
		}
		return nil, false, false
	}
	infos := t.TargetTupleType().ElementInfos()
	elements := tc.GetTypeArguments(t)
	if len(infos) != len(elements) {
		return nil, false, false
	}
	var types []*checker.Type
	for i, info := range infos {
		flags := info.TupleElementFlags()
		element := elements[i]
		if flags&checker.ElementFlagsVariadic != 0 {
			first, empty, known := firstSpreadElementTypes(ctx, element, visiting)
			if !known {
				return nil, false, false
			}
			types = append(types, first...)
			if !empty {
				return types, false, true
			}
			continue
		}
		if flags&checker.ElementFlagsOptional != 0 {
			element = tc.GetUnionType([]*checker.Type{element, tc.GetUndefinedType()})
		}
		types = append(types, element)
		if flags&checker.ElementFlagsRequired != 0 {
			return types, false, true
		}
	}
	return types, true, true
}

// Matching ambient signatures preserve the standard static tuple model.
// This does not prove that the runtime iterator has not been replaced.
// Different signatures use the declared yield type instead.
func hasStandardIteratorSignature(ctx rule.RuleContext, iterator *ast.Symbol) bool {
	tc := ctx.TypeChecker
	if len(iterator.Declarations) == 0 {
		return false
	}
	allLibrary := true
	for _, declaration := range iterator.Declarations {
		if !ctx.Program().IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(declaration)) {
			allLibrary = false
			if declaration.Kind != ast.KindMethodSignature {
				return false
			}
		}
	}
	if allLibrary {
		return true
	}
	signatures := utils.GetCallSignatures(tc, tc.GetTypeOfSymbolAtLocation(iterator, ctx.SourceFile.AsNode()))
	var library []*checker.Signature
	for _, signature := range signatures {
		if declaration := signature.Declaration(); declaration != nil && ctx.Program().IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(declaration)) {
			library = append(library, signature)
		}
	}
	if len(library) == 0 {
		return false
	}
	for _, signature := range signatures {
		declaration := signature.Declaration()
		if declaration == nil {
			return false
		}
		if ctx.Program().IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(declaration)) {
			continue
		}
		matched := false
		for _, standard := range library {
			if len(signature.Parameters()) == len(standard.Parameters()) && len(signature.TypeParameters()) == 0 && len(standard.TypeParameters()) == 0 &&
				checker.Checker_isTypeIdenticalTo(tc, tc.GetReturnTypeOfSignature(signature), tc.GetReturnTypeOfSignature(standard)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func iterableElementTypes(tc *checker.Checker, t *checker.Type) ([]*checker.Type, bool, bool) {
	first := checker.Checker_getIterationTypeOfIterable(tc, checker.IterationUseSpread, checker.IterationTypeKindYield, t, nil)
	if first == nil || utils.IsTypeFlagSet(first, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return nil, false, false
	}
	if utils.IsTypeFlagSet(first, checker.TypeFlagsNever) {
		return nil, true, true
	}
	return []*checker.Type{first}, true, true
}

func checkArrayMethodPredicate(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	predicateNode *ast.Node,
	predicateType *checker.Type,
	opts resolvedOptions,
) {
	predicateNode = ast.SkipParentheses(predicateNode)
	isFunctionExpression := ast.IsArrowFunction(predicateNode) || ast.IsFunctionExpression(predicateNode)

	// Async arrow / function expression → Promise<…>, never a boolean.
	if isFunctionExpression && (ast.GetFunctionFlags(predicateNode)&ast.FunctionFlagsAsync) != 0 {
		ctx.ReportNode(predicateNode, msgPredicateCannotBeAsync())
		return
	}

	sigs := utils.GetCallSignatures(tc, predicateType)
	var collected []*checker.Type
	for _, sig := range sigs {
		rt := checker.Checker_getReturnTypeOfSignature(tc, sig)
		if utils.IsTypeParameter(rt) {
			if constraint := checker.Checker_getBaseConstraintOfType(tc, rt); constraint != nil {
				rt = constraint
			}
		}
		collected = append(collected, rt)
	}

	// Flatten through union constituents and dedupe.
	flat := dedupeTypes(flatMapTypes(collected, utils.UnionTypeParts))
	variants := inspectVariantTypes(tc, flat)
	reportID := determineReportType(variants, opts)
	if reportID == "" {
		return
	}

	contextLabel := "array predicate return type"
	var msg rule.RuleMessage
	switch reportID {
	case "conditionErrorAny":
		msg = msgConditionErrorAny(contextLabel)
	case "conditionErrorNullableBoolean":
		msg = msgConditionErrorNullableBoolean(contextLabel)
	case "conditionErrorNullableEnum":
		msg = msgConditionErrorNullableEnum(contextLabel)
	case "conditionErrorNullableNumber":
		msg = msgConditionErrorNullableNumber(contextLabel)
	case "conditionErrorNullableObject":
		msg = msgConditionErrorNullableObject(contextLabel)
	case "conditionErrorNullableString":
		msg = msgConditionErrorNullableString(contextLabel)
	case "conditionErrorNullish":
		msg = msgConditionErrorNullish()
	case "conditionErrorNumber":
		msg = msgConditionErrorNumber(contextLabel)
	case "conditionErrorObject":
		msg = msgConditionErrorObject(contextLabel)
	case "conditionErrorOther":
		msg = msgConditionErrorOther()
	case "conditionErrorString":
		msg = msgConditionErrorString(contextLabel)
	}

	var suggestions []rule.RuleSuggestion
	if isFunctionExpression {
		body := predicateNode.Body()
		// Expression-bodied function: emit the same suggestion fixes as a
		// conditional, anchored at the body expression.
		if body != nil && !ast.IsBlock(body) {
			suggestions = append(suggestions, getSuggestionsForConditionError(sf, tc, ast.SkipParentheses(body), reportID, opts)...)
		}

		// Add explicit boolean return-type annotation suggestion when the
		// function does not already have one.
		if !hasReturnTypeAnnotation(predicateNode) {
			suggestions = append(suggestions, explicitBooleanReturnTypeSuggestion(sf, predicateNode))
		}
	}

	if len(suggestions) == 0 {
		ctx.ReportNode(predicateNode, msg)
		return
	}
	ctx.ReportNodeWithSuggestions(predicateNode, msg, suggestions...)
}

func hasReturnTypeAnnotation(fn *ast.Node) bool {
	switch fn.Kind {
	case ast.KindArrowFunction:
		return fn.AsArrowFunction().Type != nil
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Type != nil
	}
	return false
}

// explicitBooleanReturnTypeSuggestion produces a textual fix that inserts a
// `: boolean` return-type annotation on the predicate function. The exact
// insertion point depends on whether the predicate is a parenless arrow, a
// no-arg arrow / function expression, or a normal parameter-listed function.
func explicitBooleanReturnTypeSuggestion(sf *ast.SourceFile, fn *ast.Node) rule.RuleSuggestion {
	var fixes []rule.RuleFix

	if ast.IsArrowFunction(fn) && utils.IsParenlessArrowFunction(fn) {
		af := fn.AsArrowFunction()
		params := af.Parameters
		if params != nil && len(params.Nodes) > 0 {
			p := params.Nodes[0]
			fixes = append(fixes,
				rule.RuleFixInsertBefore(sf, p, "("),
				rule.RuleFixInsertAfter(p, "): boolean"),
			)
			return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType(), FixesArr: fixes}
		}
	}

	params := getFunctionParameters(fn)
	if params == nil {
		return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType()}
	}

	// The parser records the parameter-list end immediately before its closing
	// token. Start there so comments and parentheses in generic constraints or
	// default values cannot redirect the annotation into unrelated source.
	closeParen, ok := utils.TokenAtOrAfter(sf, params.End())
	if !ok || closeParen.Kind != ast.KindCloseParenToken {
		return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType()}
	}
	fixes = append(fixes, rule.RuleFix{
		Text:  ": boolean",
		Range: core.NewTextRange(closeParen.End, closeParen.End),
	})
	return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType(), FixesArr: fixes}
}

func getFunctionParameters(fn *ast.Node) *ast.NodeList {
	switch fn.Kind {
	case ast.KindArrowFunction:
		return fn.AsArrowFunction().Parameters
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Parameters
	}
	return nil
}

// ----- Wrapping fixer ---------------------------------------------------

// wrappingFix is a port of upstream's `getWrappingFixer`. It replaces `node`
// with `wrap(text-of-innerNode)`, adding parens around the inner code when the
// inner node is not strong-precedence, and around the outer result when the
// surrounding parent is weak-precedence.
//
// The semicolon-prefix hazard (`;(expr)` to keep ASI honest when the previous
// line had no trailing `;`) is replicated for the cases where the resulting
// code starts with `(`, `[`, or a backtick.
func wrappingFix(
	sf *ast.SourceFile,
	msg rule.RuleMessage,
	node *ast.Node,
	innerNode *ast.Node,
	wrap func(code string) string,
) rule.RuleSuggestion {
	innerText := utils.TrimmedNodeText(sf, innerNode)
	if !isStrongInnerNode(innerNode) || isObjectExpressionInOneLineReturn(node, innerNode) {
		innerText = "(" + innerText + ")"
	}

	code := wrap(innerText)

	if typescriptutil.IsWeakPrecedenceParent(node) && !isAlreadyParenthesized(node) {
		code = "(" + code + ")"
	}

	if needsLeadingSemicolon(sf, node, code) {
		code = ";" + code
	}

	return rule.RuleSuggestion{
		Message:  msg,
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(sf, node, code)},
	}
}

// isStrongInnerNode extends utils.IsStrongPrecedenceNode with kinds the
// shared helper does not cover but matter for this rule's wrapping fixer.
//
// Intentional divergence from upstream's `isStrongPrecedenceNode`
// (Phase 1 Step 6.A): upstream's list omits these because ESTree doesn't
// distinguish them or they don't apply to ESTree. tsgo has them as first-class
// AST kinds and they ARE strong-precedence in JS grammar, so wrapping them in
// extra parens would produce noisier (less idiomatic) fix output without
// changing semantics:
//   - `this` / `super` — bind exactly as tightly as Identifier; ESTree models
//     them as `ThisExpression` / `Super` and upstream just doesn't enumerate
//     them, but the rule wouldn't actually hit them through a fix path on
//     ESTree either.
//   - `x!` non-null assertion (TSNonNullExpression) — single-token postfix
//     that binds tighter than any binary operator, so the outer wrap-or-not
//     decision is identical to `x` alone. Locked in by the Dimension-4
//     extras test for `x!` operand of `!` and `if (x!)` outputs.
func isStrongInnerNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	// ESTree exposes an optional chain as ChainExpression and even a template
	// without substitutions as TemplateLiteral. Neither is a strong node in
	// the upstream wrapping fixer.
	if ast.IsOptionalChain(ast.SkipOuterExpressions(node, ast.OEKNonNullAssertions)) || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return false
	}
	if utils.IsStrongPrecedenceNode(node) {
		return true
	}
	switch node.Kind {
	case ast.KindThisKeyword, ast.KindSuperKeyword, ast.KindNonNullExpression:
		return true
	}
	return false
}

func isAlreadyParenthesized(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	return parent.Kind == ast.KindParenthesizedExpression
}

// isObjectExpressionInOneLineReturn matches upstream's same-named helper: an
// arrow function whose body IS the candidate inner node, where the inner node
// is an object literal — needed to avoid `() => { … }` being mistaken for a
// block when emitting `() => ({ … })` style replacements.
func isObjectExpressionInOneLineReturn(node, innerNode *ast.Node) bool {
	if node == nil || innerNode == nil {
		return false
	}
	parent := node.Parent
	if parent == nil || !ast.IsArrowFunction(parent) {
		return false
	}
	if parent.AsArrowFunction().Body != node {
		return false
	}
	return ast.IsObjectLiteralExpression(innerNode)
}

var leadingSemiTrigger = regexp.MustCompile("^[`(\\[]")

func needsLeadingSemicolon(sf *ast.SourceFile, node *ast.Node, code string) bool {
	if !leadingSemiTrigger.MatchString(code) || !utils.IsStartOfExpressionStatement(sf, node) {
		return false
	}
	// Walk up: while the node is in a left-hand position relative to its
	// parent, keep going. Then check whether the previous statement (if any)
	// in the surrounding block / program terminated with `;`.
	current := node
	for {
		parent := current.Parent
		if parent == nil {
			return false
		}
		if ast.IsExpressionStatement(parent) {
			block := parent.Parent
			if block != nil {
				statements := getBlockStatements(block)
				idx := indexOfNode(statements, parent)
				if idx > 0 {
					prev := statements[idx-1]
					last, ok := utils.TokenBeforePosition(sf, prev.End())
					if ok && last.Kind != ast.KindSemicolonToken {
						return true
					}
				}
			}
			return false
		}
		current = parent
	}
}

func getBlockStatements(block *ast.Node) []*ast.Node {
	switch block.Kind {
	case ast.KindSourceFile:
		return block.AsSourceFile().Statements.Nodes
	case ast.KindBlock:
		return block.AsBlock().Statements.Nodes
	case ast.KindModuleBlock:
		return block.AsModuleBlock().Statements.Nodes
	case ast.KindCaseClause, ast.KindDefaultClause:
		return block.AsCaseOrDefaultClause().Statements.Nodes
	}
	return nil
}

func indexOfNode(nodes []*ast.Node, target *ast.Node) int {
	for i, n := range nodes {
		if n == target {
			return i
		}
	}
	return -1
}

// ----- Array length detection -------------------------------------------

// isArrayLengthExpression mirrors upstream's `isArrayLengthExpression`: a
// non-computed `.length` access on a value whose constrained type is an array
// or a union of arrays.
func isArrayLengthExpression(tc *checker.Checker, node *ast.Node) bool {
	if node == nil || !ast.IsPropertyAccessExpression(node) || ast.IsOptionalChain(node) {
		return false
	}
	access := node.AsPropertyAccessExpression()
	if access.Name() == nil || access.Name().Text() != "length" {
		return false
	}
	objType := utils.GetConstrainedTypeAtLocation(tc, access.Expression)
	return isArrayTypeOrUnionOfArrayTypes(tc, objType)
}

func isArrayTypeOrUnionOfArrayTypes(tc *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(t) {
		if !checker.Checker_isArrayType(tc, part) {
			return false
		}
	}
	return len(utils.UnionTypeParts(t)) > 0
}

// ----- Branded boolean detection ----------------------------------------

func isBooleanType(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral)
}

// isBrandedBoolean returns true for `boolean & { __brand: 'X' }` style
// intersections, which upstream classifies as "boolean" rather than "object".
func isBrandedBoolean(t *checker.Type) bool {
	if !utils.IsTypeFlagSet(t, checker.TypeFlagsIntersection) {
		return false
	}
	for _, part := range t.Types() {
		if isBooleanType(part) {
			return true
		}
	}
	return false
}

// ----- Type helpers (small slice utilities, kept local so they read on the
// page next to the variant-inspection code).

func anyOf(types []*checker.Type, pred func(*checker.Type) bool) bool {
	for _, t := range types {
		if pred(t) {
			return true
		}
	}
	return false
}

func filterTypes(types []*checker.Type, pred func(*checker.Type) bool) []*checker.Type {
	out := types[:0:0]
	for _, t := range types {
		if pred(t) {
			out = append(out, t)
		}
	}
	return out
}

func everyType(types []*checker.Type, pred func(*checker.Type) bool) bool {
	for _, t := range types {
		if !pred(t) {
			return false
		}
	}
	return true
}

func flatMapTypes(types []*checker.Type, mapper func(*checker.Type) []*checker.Type) []*checker.Type {
	var out []*checker.Type
	for _, t := range types {
		out = append(out, mapper(t)...)
	}
	return out
}

func dedupeTypes(types []*checker.Type) []*checker.Type {
	seen := map[*checker.Type]struct{}{}
	out := types[:0:0]
	for _, t := range types {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
