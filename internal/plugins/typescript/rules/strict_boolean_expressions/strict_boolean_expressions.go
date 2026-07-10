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
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options mirrors typescript-eslint's option shape. Pointer-typed booleans
// distinguish "user explicitly set false" from "user did not set, fall back to
// upstream default" — important because upstream defaults are split: string /
// number / nullable-object default to `true`, everything else defaults to
// `false`.
type Options struct {
	AllowAny                                               *bool `json:"allowAny"`
	AllowNullableBoolean                                   *bool `json:"allowNullableBoolean"`
	AllowNullableEnum                                      *bool `json:"allowNullableEnum"`
	AllowNullableNumber                                    *bool `json:"allowNullableNumber"`
	AllowNullableObject                                    *bool `json:"allowNullableObject"`
	AllowNullableString                                    *bool `json:"allowNullableString"`
	AllowNumber                                            *bool `json:"allowNumber"`
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
	AllowString                                            *bool `json:"allowString"`
}

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

func parseOptions(options any) resolvedOptions {
	// Defaults match upstream:
	//   allowString=true, allowNumber=true, allowNullableObject=true,
	//   all other booleans default to false.
	opts := resolvedOptions{
		allowString:         true,
		allowNumber:         true,
		allowNullableObject: true,
	}

	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}

	// Round-trip via JSON so the call accepts both the CLI shape (bare object)
	// and the rule_tester shape (array-wrapped), matching every other
	// typescript-eslint rule in this repo.
	jsonBytes, err := json.Marshal(optsMap)
	if err != nil {
		return opts
	}
	var parsed Options
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		return opts
	}

	if parsed.AllowAny != nil {
		opts.allowAny = *parsed.AllowAny
	}
	if parsed.AllowNullableBoolean != nil {
		opts.allowNullableBoolean = *parsed.AllowNullableBoolean
	}
	if parsed.AllowNullableEnum != nil {
		opts.allowNullableEnum = *parsed.AllowNullableEnum
	}
	if parsed.AllowNullableNumber != nil {
		opts.allowNullableNumber = *parsed.AllowNullableNumber
	}
	if parsed.AllowNullableObject != nil {
		opts.allowNullableObject = *parsed.AllowNullableObject
	}
	if parsed.AllowNullableString != nil {
		opts.allowNullableString = *parsed.AllowNullableString
	}
	if parsed.AllowNumber != nil {
		opts.allowNumber = *parsed.AllowNumber
	}
	if parsed.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing != nil {
		opts.allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = *parsed.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing
	}
	if parsed.AllowString != nil {
		opts.allowString = *parsed.AllowString
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
	return rule.RuleMessage{Id: "conditionFixDefaultEmptyString", Description: "Explicitly treat nullish value the same as an empty string (`value ?? \"\")`"}
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
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _optionsRaw []any) rule.RuleListeners {
		optionsRaw := rule.LegacyUnwrapOptions(_optionsRaw)
		opts := parseOptions(optionsRaw)
		tc := ctx.TypeChecker
		sf := ctx.SourceFile

		// strictNullChecks gate. Matches upstream: when strictNullChecks is
		// off and the user has not opted in via the long-named "I know what
		// I am doing" flag, the rule reports a single noStrictNullCheck
		// diagnostic anchored at the start of the file and still walks the
		// rest of the file.
		compilerOptions := ctx.Program.Options()
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
				if utils.IsArrayMethodCallWithPredicate(tc, callExpr) && len(callExpr.Arguments.Nodes) > 0 {
					checkArrayMethodPredicate(ctx, sf, tc, callExpr.Arguments.Nodes[0], opts)
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
		if neg != nil {
			return []rule.RuleSuggestion{
				wrappingFix(sf, sugCompareStringLength(), neg, node, func(code string) string {
					return code + ".length === 0"
				}),
				wrappingFix(sf, sugCompareEmptyString(), neg, node, func(code string) string {
					return code + ` === ""`
				}),
				wrappingFix(sf, sugCastBoolean(), neg, node, func(code string) string {
					return "!Boolean(" + code + ")"
				}),
			}
		}
		return []rule.RuleSuggestion{
			wrappingFix(sf, sugCompareStringLength(), node, node, func(code string) string {
				return code + ".length > 0"
			}),
			wrappingFix(sf, sugCompareEmptyString(), node, node, func(code string) string {
				return code + ` !== ""`
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

func checkArrayMethodPredicate(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	predicateNode *ast.Node,
	opts resolvedOptions,
) {
	isFunctionExpression := ast.IsArrowFunction(predicateNode) || ast.IsFunctionExpression(predicateNode)

	// Async arrow / function expression → Promise<…>, never a boolean.
	if isFunctionExpression && (ast.GetFunctionFlags(predicateNode)&ast.FunctionFlagsAsync) != 0 {
		ctx.ReportNode(predicateNode, msgPredicateCannotBeAsync())
		return
	}

	predicateType := utils.GetConstrainedTypeAtLocation(tc, predicateNode)
	sigs := utils.GetCallSignatures(tc, predicateType)
	if len(sigs) == 0 {
		return
	}

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
			suggestions = append(suggestions, getSuggestionsForConditionError(sf, tc, body, reportID, opts)...)
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
	src := sf.Text()
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

	if len(params.Nodes) == 0 {
		// `() => …` / `function () {}` — scan from the function start for the
		// first `)` and insert `: boolean` after it.
		closeParen := findFirstCharAt(src, fn.Pos(), ')')
		if closeParen < 0 {
			return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType()}
		}
		fixes = append(fixes, rule.RuleFix{
			Text:  ": boolean",
			Range: core.NewTextRange(closeParen+1, closeParen+1),
		})
		return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType(), FixesArr: fixes}
	}

	lastParam := params.Nodes[len(params.Nodes)-1]
	closeParen := findFirstCharAt(src, lastParam.End(), ')')
	if closeParen < 0 {
		return rule.RuleSuggestion{Message: msgExplicitBooleanReturnType()}
	}
	fixes = append(fixes, rule.RuleFix{
		Text:  ": boolean",
		Range: core.NewTextRange(closeParen+1, closeParen+1),
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

// findFirstCharAt scans `src` forward from `from`, returning the index of the
// first occurrence of `ch`. Returns -1 when not found. Sufficient for the
// narrow `(`/`)` scans the predicate-return-type fixer performs because
// neither character can appear inside a parameter name or default value
// without an enclosing paren the AST already tracks.
func findFirstCharAt(src string, from int, ch byte) int {
	for i := from; i < len(src); i++ {
		if src[i] == ch {
			return i
		}
	}
	return -1
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
	if !leadingSemiTrigger.MatchString(code) {
		return false
	}
	src := sf.Text()
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
			if block != nil && (block.Kind == ast.KindSourceFile || block.Kind == ast.KindBlock) {
				statements := getBlockStatements(block)
				idx := indexOfNode(statements, parent)
				if idx > 0 {
					prev := statements[idx-1]
					last := lastSignificantChar(src, prev.End())
					if last != ';' {
						return true
					}
				}
			}
			return false
		}
		if !isLeftHandSide(current) {
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

// lastSignificantChar returns the last non-whitespace character at or before
// `endExclusive`, skipping common trivia. Used by the ASI hazard check.
func lastSignificantChar(src string, endExclusive int) byte {
	for i := endExclusive - 1; i >= 0; i-- {
		ch := src[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			continue
		}
		return ch
	}
	return 0
}

// isLeftHandSide mirrors upstream's `isLeftHandSide` (the private helper used
// by the wrapping fixer's ASI check). Upstream only treats `UpdateExpression`
// (`a++` / `++a`) as a left-hand parent on the unary side, not the broader
// `UnaryExpression` (`!a`, `~a`, `typeof a`, ...). tsgo collapses both shapes
// into PrefixUnary / PostfixUnary, so we differentiate by operator: only the
// update operators `++` and `--` should return true.
func isLeftHandSide(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindPostfixUnaryExpression {
		// `a++` / `a--` — operand is the leftmost token. Upstream UpdateExpression.
		return true
	}
	if parent.Kind == ast.KindPrefixUnaryExpression {
		op := parent.AsPrefixUnaryExpression().Operator
		// Only `++a` / `--a` count as UpdateExpression upstream-side; `!a`,
		// `~a`, `+a`, `-a` are UnaryExpression and upstream returns false.
		return op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken
	}
	if ast.IsBinaryExpression(parent) {
		return parent.AsBinaryExpression().Left == node
	}
	if ast.IsConditionalExpression(parent) {
		return parent.AsConditionalExpression().Condition == node
	}
	if ast.IsCallExpression(parent) {
		return parent.AsCallExpression().Expression == node
	}
	if ast.IsTaggedTemplateExpression(parent) {
		return parent.AsTaggedTemplateExpression().Tag == node
	}
	return false
}

// ----- Array length detection -------------------------------------------

// isArrayLengthExpression mirrors upstream's `isArrayLengthExpression`: a
// non-computed `.length` access on a value whose constrained type is an array
// or a union of arrays.
func isArrayLengthExpression(tc *checker.Checker, node *ast.Node) bool {
	if node == nil || !ast.IsPropertyAccessExpression(node) {
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
