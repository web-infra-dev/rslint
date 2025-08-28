package prefer_nullish_coalescing

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type PreferNullishCoalescingOptions struct {
	AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing *bool                          `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
	IgnoreBooleanCoercion                                  *bool                          `json:"ignoreBooleanCoercion"`
	IgnoreConditionalTests                                 *bool                          `json:"ignoreConditionalTests"`
	IgnoreIfStatements                                     *bool                          `json:"ignoreIfStatements"`
	IgnoreMixedLogicalExpressions                          *bool                          `json:"ignoreMixedLogicalExpressions"`
	IgnorePrimitives                                       *PreferNullishPrimitivesOption `json:"ignorePrimitives"`
	IgnoreTernaryTests                                     *bool                          `json:"ignoreTernaryTests"`
}

type PreferNullishPrimitivesOption struct {
	Boolean *bool `json:"boolean"`
	String  *bool `json:"string"`
	Number  *bool `json:"number"`
	Bigint  *bool `json:"bigint"`
}

func parseOptions(options any) PreferNullishCoalescingOptions {
	opts := PreferNullishCoalescingOptions{
		AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(false),
		IgnoreBooleanCoercion:         utils.Ref(false),
		IgnoreConditionalTests:        utils.Ref(true),
		IgnoreIfStatements:            utils.Ref(false),
		IgnoreMixedLogicalExpressions: utils.Ref(true),
		IgnorePrimitives: &PreferNullishPrimitivesOption{
			Boolean: utils.Ref(false),
			String:  utils.Ref(false),
			Number:  utils.Ref(false),
			Bigint:  utils.Ref(false),
		},
		IgnoreTernaryTests: utils.Ref(false),
	}

	if options == nil {
		return opts
	}

	// Handle array format: [{ option: value }]
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				parseOptionsFromMap(m, &opts)
			}
		}
		return opts
	}

	// Handle direct object format
	if m, ok := options.(map[string]interface{}); ok {
		parseOptionsFromMap(m, &opts)
	}

	return opts
}

func parseOptionsFromMap(m map[string]interface{}, opts *PreferNullishCoalescingOptions) {
	if v, ok := m["allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"].(bool); ok {
		opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing = &v
	}
	if v, ok := m["ignoreBooleanCoercion"].(bool); ok {
		opts.IgnoreBooleanCoercion = &v
	}
	if v, ok := m["ignoreConditionalTests"].(bool); ok {
		opts.IgnoreConditionalTests = &v
	}
	if v, ok := m["ignoreIfStatements"].(bool); ok {
		opts.IgnoreIfStatements = &v
	}
	if v, ok := m["ignoreMixedLogicalExpressions"].(bool); ok {
		opts.IgnoreMixedLogicalExpressions = &v
	}
	if v, ok := m["ignoreTernaryTests"].(bool); ok {
		opts.IgnoreTernaryTests = &v
	}

	// Handle ignorePrimitives option
	if primitives, ok := m["ignorePrimitives"]; ok {
		if primitivesBool, ok := primitives.(bool); ok && primitivesBool {
			// If true, ignore all primitives
			opts.IgnorePrimitives.Boolean = utils.Ref(true)
			opts.IgnorePrimitives.String = utils.Ref(true)
			opts.IgnorePrimitives.Number = utils.Ref(true)
			opts.IgnorePrimitives.Bigint = utils.Ref(true)
		} else if primitivesMap, ok := primitives.(map[string]interface{}); ok {
			if v, ok := primitivesMap["boolean"].(bool); ok {
				opts.IgnorePrimitives.Boolean = &v
			}
			if v, ok := primitivesMap["string"].(bool); ok {
				opts.IgnorePrimitives.String = &v
			}
			if v, ok := primitivesMap["number"].(bool); ok {
				opts.IgnorePrimitives.Number = &v
			}
			if v, ok := primitivesMap["bigint"].(bool); ok {
				opts.IgnorePrimitives.Bigint = &v
			}
		}
	}
}

func buildNoStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

func buildPreferNullishOverOrMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverOr",
		Description: "Prefer using nullish coalescing operator (`??`) instead of a logical or (`||`), as it is a safer operator.",
	}
}

func buildPreferNullishOverTernaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverTernary",
		Description: "Prefer using nullish coalescing operator (`??`) instead of a ternary expression testing for null/undefined.",
	}
}

func buildPreferNullishOverAssignmentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverAssignment",
		Description: "Prefer using nullish coalescing assignment (`??=`) instead of an if statement with a nullish check.",
	}
}

func buildSuggestNullishMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestNullish",
		Description: "Fix to nullish coalescing operator (`??`).",
	}
}

// isNullableType checks if a type includes null or undefined
func isNullableType(t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, unionType := range t.Types() {
			flags := checker.Type_flags(unionType)
			if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
				return true
			}
		}
	}
	flags := checker.Type_flags(t)
	return flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0
}

// isTypeEligibleForPreferNullish checks if a type is eligible for nullish coalescing conversion
func isTypeEligibleForPreferNullish(t *checker.Type, opts PreferNullishCoalescingOptions) bool {
	if !isNullableType(t) {
		return false
	}

	// Check for ignorable flags based on options
	var ignorableFlags checker.TypeFlags
	if opts.IgnorePrimitives.Boolean != nil && *opts.IgnorePrimitives.Boolean {
		ignorableFlags |= checker.TypeFlagsBooleanLike
	}
	if opts.IgnorePrimitives.String != nil && *opts.IgnorePrimitives.String {
		ignorableFlags |= checker.TypeFlagsStringLike
	}
	if opts.IgnorePrimitives.Number != nil && *opts.IgnorePrimitives.Number {
		ignorableFlags |= checker.TypeFlagsNumberLike
	}
	if opts.IgnorePrimitives.Bigint != nil && *opts.IgnorePrimitives.Bigint {
		ignorableFlags |= checker.TypeFlagsBigIntLike
	}

	// If the type is any or unknown and we're ignoring primitives, return false
	// This matches TypeScript ESLint behavior
	flags := checker.Type_flags(t)
	if ignorableFlags != 0 && flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return false
	}

	// Check if any type constituents have intersection types with ignored primitives
	// This handles branded types like (string & { __brand?: any })
	if ignorableFlags != 0 {
		constituents := utils.UnionTypeParts(t)
		for _, constituent := range constituents {
			// Skip null and undefined types
			constituentFlags := checker.Type_flags(constituent)
			if constituentFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
				continue
			}

			// Check if this constituent is an intersection type
			if utils.IsIntersectionType(constituent) {
				// Check intersection constituents for ignored primitives
				// This matches the TypeScript ESLint implementation
				intersectionParts := utils.IntersectionTypeParts(constituent)
				for _, part := range intersectionParts {
					partFlags := checker.Type_flags(part)
					if partFlags&ignorableFlags != 0 {
						// Found an intersection type containing an ignored primitive
						return false
					}
				}
			}
		}
	}

	// Check if any non-intersection constituents match the ignorable flags
	if ignorableFlags != 0 {
		constituents := utils.UnionTypeParts(t)
		for _, constituent := range constituents {
			// Skip null and undefined types
			constituentFlags := checker.Type_flags(constituent)
			if constituentFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
				continue
			}

			// If not an intersection type, check if it matches ignorable flags directly
			if !utils.IsIntersectionType(constituent) {
				if constituentFlags&ignorableFlags != 0 {
					return false
				}
			}
		}
	}

	return true
}

// isMemberAccessLike checks if a node is a member access-like expression
func isMemberAccessLike(node *ast.Node) bool {
	return node.Kind == ast.KindIdentifier ||
		node.Kind == ast.KindPropertyAccessExpression ||
		node.Kind == ast.KindElementAccessExpression ||
		node.Kind == ast.KindCallExpression ||
		node.Kind == ast.KindNewExpression
}

// isConditionalTest checks if a node is within a conditional test context
func isConditionalTest(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Walk up the parent chain and check if we eventually find a conditional statement
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			break
		}

		// Skip over parenthesized expressions
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}

		// If we find a conditional statement, check if we're in its condition
		switch parent.Kind {
		case ast.KindConditionalExpression:
			// We found a ternary expression, check if we're in its test part
			condExpr := parent.AsConditionalExpression()
			if condExpr != nil && condExpr.Condition != nil {
				// Check if current (which may be a parenthesized expression) is the condition
				if current == condExpr.Condition {
					return true
				}
				// Also walk up from our original node to see if we reach the test condition
				temp := node
				for temp != nil {
					if temp == condExpr.Condition {
						return true
					}
					temp = temp.Parent
				}
			}
		case ast.KindIfStatement:
			// We found an if statement, check if we're in its condition part
			ifStmt := parent.AsIfStatement()
			if ifStmt != nil && ifStmt.Expression != nil {
				// Walk up from our original node to see if we reach the if condition
				temp := node
				for temp != nil {
					if temp == ifStmt.Expression {
						return true
					}
					temp = temp.Parent
				}
			}
		case ast.KindWhileStatement:
			whileStmt := parent.AsWhileStatement()
			if whileStmt != nil && whileStmt.Expression != nil {
				temp := node
				for temp != nil {
					if temp == whileStmt.Expression {
						return true
					}
					temp = temp.Parent
				}
			}
		case ast.KindDoStatement:
			doStmt := parent.AsDoStatement()
			if doStmt != nil && doStmt.Expression != nil {
				temp := node
				for temp != nil {
					if temp == doStmt.Expression {
						return true
					}
					temp = temp.Parent
				}
			}
		case ast.KindForStatement:
			forStmt := parent.AsForStatement()
			if forStmt != nil && forStmt.Condition != nil {
				temp := node
				for temp != nil {
					if temp == forStmt.Condition {
						return true
					}
					temp = temp.Parent
				}
			}
		}

		current = parent
	}

	return false
}

// containsNode checks if container contains the target node as a descendant

// isBooleanConstructorContext checks if a node is within a Boolean constructor context
func isBooleanConstructorContext(node *ast.Node) bool {
	// Check up to 10 levels to avoid infinite recursion
	for i := 0; i < 10 && node != nil; i++ {
		parent := node.Parent
		if parent == nil {
			return false
		}

		// Check if we've reached a Boolean constructor call
		if parent.Kind == ast.KindCallExpression {
			callExpr := parent.AsCallExpression()
			if callExpr != nil && callExpr.Expression.Kind == ast.KindIdentifier {
				identifier := callExpr.Expression.AsIdentifier()
				if identifier != nil && identifier.Text == "Boolean" {
					return true
				}
			}
		}

		// Continue traversing through these node types
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			// Continue through parentheses
		case ast.KindBinaryExpression:
			binExpr := parent.AsBinaryExpression()
			if binExpr != nil {
				// Allow traversal through logical, nullish coalescing, and comma operators
				switch binExpr.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken,
					ast.KindBarBarToken,
					ast.KindQuestionQuestionToken,
					ast.KindCommaToken:
					// Continue checking
				default:
					return false
				}
			}
		case ast.KindConditionalExpression:
			// Continue checking through ternary
		default:
			// Stop traversal for other node types
			return false
		}

		node = parent
	}
	return false
}

// isMixedLogicalExpression checks if a logical expression is mixed with && operators
func isMixedLogicalExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check if this is a || expression
	if node.Kind != ast.KindBinaryExpression {
		return false
	}

	binExpr := node.AsBinaryExpression()
	if binExpr == nil || binExpr.OperatorToken.Kind != ast.KindBarBarToken {
		return false
	}

	// Look for && operators in the entire expression tree
	// Start from the topmost || expression by finding the root of the expression chain
	root := node
	for root.Parent != nil && root.Parent.Kind == ast.KindBinaryExpression {
		parentBin := root.Parent.AsBinaryExpression()
		if parentBin != nil && parentBin.OperatorToken.Kind == ast.KindBarBarToken {
			root = root.Parent
		} else {
			break
		}
	}

	// Now check if this entire expression chain contains any && operators
	return hasAndOperator(root)
}

// hasAndOperator recursively checks if a node contains && operators
func hasAndOperator(node *ast.Node) bool {
	if node == nil {
		return false
	}

	if node.Kind == ast.KindBinaryExpression {
		binExpr := node.AsBinaryExpression()
		if binExpr != nil {
			if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
				return true
			}
			// Recursively check both sides
			return hasAndOperator(binExpr.Left) || hasAndOperator(binExpr.Right)
		}
	}

	return false
}

// getNodeText extracts the text corresponding to a node from the given source file.
//
// Safety mechanisms:
// - Checks if either sourceFile or node is nil, returning an empty string if so.
// - Retrieves the start and end positions of the node and ensures they are within the bounds of the source text.
// - If the start position is negative, the end position exceeds the length of the text, or start > end, returns an empty string.
// - Only returns the substring if all checks pass, preventing panics or out-of-bounds errors.
func getNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	text := sourceFile.Text()
	start := node.Pos()
	end := node.End()
	if start < 0 || end > len(text) || start > end {
		return ""
	}
	return text[start:end]
}

// needsParentheses checks if an expression needs parentheses when used as the right operand of ??
func needsParentheses(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBinaryExpression:
		binExpr := node.AsBinaryExpression()
		if binExpr != nil {
			// Lower precedence operators need parentheses
			switch binExpr.OperatorToken.Kind {
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
				return true
			}
		}
	case ast.KindConditionalExpression:
		return true
	}
	return false
}

// areNodesTextuallyEqual checks if two nodes have the same text content
func areNodesTextuallyEqual(sourceFile *ast.SourceFile, a, b *ast.Node) bool {
	if a == nil || b == nil {
		return false
	}
	return getNodeText(sourceFile, a) == getNodeText(sourceFile, b)
}

// isSingleNullishCheck checks if the condition is a single check for null or undefined (strict equality)
// and whether the type allows this to be converted to nullish coalescing
func isSingleNullishCheck(condition, whenTrue, whenFalse *ast.Node, ctx rule.RuleContext) (bool, *ast.Node) {
	condition = unwrapParentheses(condition)
	if condition == nil || condition.Kind != ast.KindBinaryExpression {
		return false, nil
	}

	binExpr := condition.AsBinaryExpression()
	if binExpr == nil {
		return false, nil
	}

	var target *ast.Node
	var checkingFor string // "null" or "undefined"

	// Check for x !== undefined or x !== null (strict inequality)
	if binExpr.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
		if binExpr.Right.Kind == ast.KindIdentifier && binExpr.Right.AsIdentifier() != nil &&
			binExpr.Right.AsIdentifier().Text == "undefined" {
			checkingFor = "undefined"
			if areNodesSemanticallyEqual(binExpr.Left, whenTrue) {
				target = binExpr.Left
			}
		} else if binExpr.Left.Kind == ast.KindIdentifier && binExpr.Left.AsIdentifier() != nil &&
			binExpr.Left.AsIdentifier().Text == "undefined" {
			checkingFor = "undefined"
			if areNodesSemanticallyEqual(binExpr.Right, whenTrue) {
				target = binExpr.Right
			}
		} else if binExpr.Right.Kind == ast.KindNullKeyword {
			checkingFor = "null"
			if areNodesSemanticallyEqual(binExpr.Left, whenTrue) {
				target = binExpr.Left
			}
		} else if binExpr.Left.Kind == ast.KindNullKeyword {
			checkingFor = "null"
			if areNodesSemanticallyEqual(binExpr.Right, whenTrue) {
				target = binExpr.Right
			}
		}
		// Check for x === undefined or x === null (strict equality)
	} else if binExpr.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken {
		if binExpr.Right.Kind == ast.KindIdentifier && binExpr.Right.AsIdentifier() != nil &&
			binExpr.Right.AsIdentifier().Text == "undefined" {
			checkingFor = "undefined"
			if areNodesSemanticallyEqual(binExpr.Left, whenFalse) {
				target = binExpr.Left
			}
		} else if binExpr.Left.Kind == ast.KindIdentifier && binExpr.Left.AsIdentifier() != nil &&
			binExpr.Left.AsIdentifier().Text == "undefined" {
			checkingFor = "undefined"
			if areNodesSemanticallyEqual(binExpr.Right, whenFalse) {
				target = binExpr.Right
			}
		} else if binExpr.Right.Kind == ast.KindNullKeyword {
			checkingFor = "null"
			if areNodesSemanticallyEqual(binExpr.Left, whenFalse) {
				target = binExpr.Left
			}
		} else if binExpr.Left.Kind == ast.KindNullKeyword {
			checkingFor = "null"
			if areNodesSemanticallyEqual(binExpr.Right, whenFalse) {
				target = binExpr.Right
			}
		}
	}

	if target == nil || checkingFor == "" {
		return false, nil
	}

	// Now check the type to see if checking for only one nullish value is sufficient
	targetType := ctx.TypeChecker.GetTypeAtLocation(target)
	if targetType == nil {
		return false, nil
	}

	// Check what nullish types are present in the union
	hasNull := false
	hasUndefined := false

	if utils.IsUnionType(targetType) {
		for _, unionType := range targetType.Types() {
			flags := checker.Type_flags(unionType)
			if flags&checker.TypeFlagsNull != 0 {
				hasNull = true
			}
			if flags&checker.TypeFlagsUndefined != 0 {
				hasUndefined = true
			}
		}
	} else {
		flags := checker.Type_flags(targetType)
		if flags&checker.TypeFlagsNull != 0 {
			hasNull = true
		}
		if flags&checker.TypeFlagsUndefined != 0 {
			hasUndefined = true
		}
	}

	// The check is valid for nullish coalescing if:
	// 1. We're checking for undefined and the type only has undefined (not null)
	// 2. We're checking for null and the type only has null (not undefined)
	if checkingFor == "undefined" && hasUndefined && !hasNull {
		return true, target
	}
	if checkingFor == "null" && hasNull && !hasUndefined {
		return true, target
	}

	// If the type has both null and undefined, checking for only one is not sufficient
	return false, nil
}

// isExplicitNullishCheck checks if the condition is an explicit null/undefined check pattern
// like: x !== undefined && x !== null ? x : y or x === undefined || x === null ? y : x
func unwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		paren := node.AsParenthesizedExpression()
		if paren != nil {
			node = paren.Expression
		} else {
			break
		}
	}
	return node
}

// checksAreDifferentNullishTypes verifies that two check expressions check for different nullish values (one for null, one for undefined)
func checksAreDifferentNullishTypes(check1, check2 *ast.Node) bool {
	if check1 == nil || check2 == nil || check1.Kind != ast.KindBinaryExpression || check2.Kind != ast.KindBinaryExpression {
		return false
	}

	bin1 := check1.AsBinaryExpression()
	bin2 := check2.AsBinaryExpression()
	if bin1 == nil || bin2 == nil {
		return false
	}

	// Determine what each check is checking for
	check1Type := getNullishType(bin1)
	check2Type := getNullishType(bin2)

	// They must be checking for different nullish types
	// If both are checking for the same thing (e.g., both "null"), it's not a proper nullish check
	if check1Type == "" || check2Type == "" || check1Type == check2Type {
		return false
	}

	// Return true only if one checks for null and the other checks for undefined
	return (check1Type == "null" && check2Type == "undefined") || (check1Type == "undefined" && check2Type == "null")
}

// getNullishType returns "null", "undefined", or "" based on what the binary expression checks for
func getNullishType(bin *ast.BinaryExpression) string {
	if bin.Right.Kind == ast.KindNullKeyword {
		return "null"
	}
	if bin.Right.Kind == ast.KindIdentifier && bin.Right.AsIdentifier() != nil && bin.Right.AsIdentifier().Text == "undefined" {
		return "undefined"
	}
	if bin.Left.Kind == ast.KindNullKeyword {
		return "null"
	}
	if bin.Left.Kind == ast.KindIdentifier && bin.Left.AsIdentifier() != nil && bin.Left.AsIdentifier().Text == "undefined" {
		return "undefined"
	}
	return ""
}

func isExplicitNullishCheck(condition, whenTrue, whenFalse *ast.Node, sourceFile *ast.SourceFile) (bool, *ast.Node) {
	condition = unwrapParentheses(condition)

	if condition == nil {
		return false, nil
	}

	// Pattern 1: x !== undefined && x !== null ? x : y (allow both orderings)
	if condition.Kind == ast.KindBinaryExpression {
		binExpr := condition.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			// Check that the two checks are for different nullish types
			if !checksAreDifferentNullishTypes(binExpr.Left, binExpr.Right) {
				return false, nil
			}

			leftTarget := getNullishCheckTarget(binExpr.Left, sourceFile, false)
			rightTarget := getNullishCheckTarget(binExpr.Right, sourceFile, false)

			if leftTarget != nil && rightTarget != nil &&
				areNodesSemanticallyEqual(leftTarget, rightTarget) &&
				areNodesSemanticallyEqual(leftTarget, whenTrue) {
				return true, leftTarget
			}
		}
	}

	// Pattern 2: x === undefined || x === null ? y : x
	if condition.Kind == ast.KindBinaryExpression {
		binExpr := condition.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindBarBarToken {
			// Check that the two checks are for different nullish types
			different := checksAreDifferentNullishTypes(binExpr.Left, binExpr.Right)
			if !different {
				// Don't match if both checks are for the same nullish type
				return false, nil
			}

			leftTarget := getNullishCheckTarget(binExpr.Left, sourceFile, true)
			rightTarget := getNullishCheckTarget(binExpr.Right, sourceFile, true)
			if leftTarget != nil && rightTarget != nil &&
				areNodesSemanticallyEqual(leftTarget, rightTarget) &&
				areNodesSemanticallyEqual(leftTarget, whenFalse) {
				return true, leftTarget
			}
		}
	}

	// Pattern 3: x != null or x != undefined (loose equality with either covers both null and undefined)
	// This pattern ONLY matches loose equality (== or !=), not strict equality (=== or !==)
	if condition.Kind == ast.KindBinaryExpression {
		binExpr := condition.AsBinaryExpression()
		// Only match simple nullish checks, not compound ones (&&, ||)
		if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
			binExpr.OperatorToken.Kind == ast.KindBarBarToken {
			return false, nil
		}

		// Check for x != null/undefined or x == null/undefined (loose equality ONLY)
		if binExpr.OperatorToken.Kind == ast.KindExclamationEqualsToken {
			// x != null or x != undefined ? x : y pattern
			isNullOrUndefined := binExpr.Right.Kind == ast.KindNullKeyword ||
				(binExpr.Right.Kind == ast.KindIdentifier && binExpr.Right.AsIdentifier() != nil && binExpr.Right.AsIdentifier().Text == "undefined")
			if isNullOrUndefined && areNodesSemanticallyEqual(binExpr.Left, whenTrue) {
				return true, binExpr.Left
			}
			// null/undefined != x ? x : y pattern (reversed)
			isNullOrUndefined = binExpr.Left.Kind == ast.KindNullKeyword ||
				(binExpr.Left.Kind == ast.KindIdentifier && binExpr.Left.AsIdentifier() != nil && binExpr.Left.AsIdentifier().Text == "undefined")
			if isNullOrUndefined && areNodesSemanticallyEqual(binExpr.Right, whenTrue) {
				return true, binExpr.Right
			}
		} else if binExpr.OperatorToken.Kind == ast.KindEqualsEqualsToken {
			// x == null or x == undefined ? y : x pattern
			isNullOrUndefined := binExpr.Right.Kind == ast.KindNullKeyword ||
				(binExpr.Right.Kind == ast.KindIdentifier && binExpr.Right.AsIdentifier() != nil && binExpr.Right.AsIdentifier().Text == "undefined")
			if isNullOrUndefined && areNodesSemanticallyEqual(binExpr.Left, whenFalse) {
				return true, binExpr.Left
			}
			// null/undefined == x ? y : x pattern (reversed)
			isNullOrUndefined = binExpr.Left.Kind == ast.KindNullKeyword ||
				(binExpr.Left.Kind == ast.KindIdentifier && binExpr.Left.AsIdentifier() != nil && binExpr.Left.AsIdentifier().Text == "undefined")
			if isNullOrUndefined && areNodesSemanticallyEqual(binExpr.Right, whenFalse) {
				return true, binExpr.Right
			}
		}
		// Don't match strict equality (!==, ===)
	}

	return false, nil
}

// areNodesSemanticallyEqual checks if two nodes are structurally the same identifier/member access
func areNodesSemanticallyEqual(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ast.KindIdentifier:
		return a.AsIdentifier().Text == b.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		pa := a.AsPropertyAccessExpression()
		pb := b.AsPropertyAccessExpression()
		return areNodesSemanticallyEqual(pa.Expression, pb.Expression) && pa.Name().AsIdentifier().Text == pb.Name().AsIdentifier().Text
	case ast.KindElementAccessExpression:
		ea := a.AsElementAccessExpression()
		eb := b.AsElementAccessExpression()
		return areNodesSemanticallyEqual(ea.Expression, eb.Expression) && areNodesSemanticallyEqual(ea.ArgumentExpression, eb.ArgumentExpression)
	case ast.KindThisKeyword:
		// Both are 'this' keywords
		return true
	case ast.KindStringLiteral:
		// Compare string literals
		return a.AsStringLiteral().Text == b.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		// Compare numeric literals
		return a.AsNumericLiteral().Text == b.AsNumericLiteral().Text
	default:
		return false
	}
}

// getNullishCheckTarget returns the target node if the expression is a nullish check, else nil
// If reverse is true, checks for ===, else for !==
func getNullishCheckTarget(expr *ast.Node, sourceFile *ast.SourceFile, reverse bool) *ast.Node {
	if expr == nil || expr.Kind != ast.KindBinaryExpression {
		return nil
	}
	bin := expr.AsBinaryExpression()
	if bin == nil {
		return nil
	}
	if reverse {
		if bin.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && bin.OperatorToken.Kind != ast.KindEqualsEqualsToken {
			return nil
		}
	} else {
		if bin.OperatorToken.Kind != ast.KindExclamationEqualsToken && bin.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
			return nil
		}
	}
	// Check if right side is null or undefined
	if bin.Right.Kind == ast.KindNullKeyword ||
		(bin.Right.Kind == ast.KindIdentifier && bin.Right.AsIdentifier() != nil && bin.Right.AsIdentifier().Text == "undefined") {
		return bin.Left
	}
	// Check if left side is null or undefined (handle reversed order)
	if bin.Left.Kind == ast.KindNullKeyword ||
		(bin.Left.Kind == ast.KindIdentifier && bin.Left.AsIdentifier() != nil && bin.Left.AsIdentifier().Text == "undefined") {
		return bin.Right
	}
	return nil
}

var PreferNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name: "prefer-nullish-coalescing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {

		opts := parseOptions(options)

		// Check for strict null checks
		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		if !isStrictNullChecks && !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			// Handle logical OR expressions: a || b
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil {
					return
				}

				// Handle logical OR expressions: a || b
				if binExpr.OperatorToken.Kind == ast.KindBarBarToken {
					// Check if left operand is eligible for nullish coalescing
					leftType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Left)
					if !isTypeEligibleForPreferNullish(leftType, opts) {
						return
					}

					// Check various ignore conditions
					if *opts.IgnoreConditionalTests && isConditionalTest(node) {
						// Debug: This OR is in a conditional test context, skipping
						return
					}

					// Check if this is a test in a ternary expression
					if *opts.IgnoreTernaryTests && node.Parent != nil {
						// Check if direct parent is conditional expression
						if node.Parent.Kind == ast.KindConditionalExpression {
							if condExpr := node.Parent.AsConditionalExpression(); condExpr != nil && condExpr.Condition == node {
								return
							}
						}
						// Check if parent is parenthesized expression inside conditional test
						if node.Parent.Kind == ast.KindParenthesizedExpression && node.Parent.Parent != nil {
							if node.Parent.Parent.Kind == ast.KindConditionalExpression {
								if condExpr := node.Parent.Parent.AsConditionalExpression(); condExpr != nil && condExpr.Condition == node.Parent {
									return
								}
							}
						}
					}

					if *opts.IgnoreBooleanCoercion && isBooleanConstructorContext(node) {
						return
					}

					if *opts.IgnoreMixedLogicalExpressions && isMixedLogicalExpression(node) {
						return
					}

					// Create fix suggestion
					leftText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Left))
					rightText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Right))

					var fixedRightText string
					if needsParentheses(binExpr.Right) {
						fixedRightText = fmt.Sprintf("(%s)", rightText)
					} else {
						fixedRightText = rightText
					}

					replacement := fmt.Sprintf("%s ?? %s", leftText, fixedRightText)

					// Check if the entire expression needs parentheses
					if node.Parent != nil && node.Parent.Kind == ast.KindBinaryExpression {
						parentBinExpr := node.Parent.AsBinaryExpression()
						if parentBinExpr != nil && parentBinExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
							replacement = fmt.Sprintf("(%s)", replacement)
						}
					}

					ctx.ReportNodeWithSuggestions(binExpr.OperatorToken, buildPreferNullishOverOrMessage(),
						rule.RuleSuggestion{
							Message:  buildSuggestNullishMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
						},
					)
					return
				}

				// Handle logical OR assignment expressions: a ||= b
				if binExpr.OperatorToken.Kind == ast.KindBarBarEqualsToken {
					// Check if left operand is eligible for nullish coalescing
					leftType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Left)
					if !isTypeEligibleForPreferNullish(leftType, opts) {
						return
					}

					// Check various ignore conditions
					if *opts.IgnoreConditionalTests && isConditionalTest(node) {
						return
					}

					// Check if this is a test in a ternary expression
					if *opts.IgnoreTernaryTests && node.Parent != nil {
						// Check if direct parent is conditional expression
						if node.Parent.Kind == ast.KindConditionalExpression {
							if condExpr := node.Parent.AsConditionalExpression(); condExpr != nil && condExpr.Condition == node {
								return
							}
						}
						// Check if parent is parenthesized expression inside conditional test
						if node.Parent.Kind == ast.KindParenthesizedExpression && node.Parent.Parent != nil {
							if node.Parent.Parent.Kind == ast.KindConditionalExpression {
								if condExpr := node.Parent.Parent.AsConditionalExpression(); condExpr != nil && condExpr.Condition == node.Parent {
									return
								}
							}
						}
					}

					if *opts.IgnoreBooleanCoercion && isBooleanConstructorContext(node) {
						return
					}

					// Create fix suggestion
					leftText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Left))
					rightText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Right))
					replacement := fmt.Sprintf("%s ??= %s", leftText, rightText)

					ctx.ReportNodeWithSuggestions(binExpr.OperatorToken, buildPreferNullishOverOrMessage(),
						rule.RuleSuggestion{
							Message:  buildSuggestNullishMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
						},
					)
				}
			},

			// Handle ternary expressions: a ? a : b
			ast.KindConditionalExpression: func(node *ast.Node) {
				// Do not skip ternary checks for simple pattern; only respect option outside simple detection.

				condExpr := node.AsConditionalExpression()
				if condExpr == nil {
					return
				}

				// Check if this is a nullish check pattern
				// Two cases to check:
				// 1. Simple case: a ? a : b (where condition and consequent are the same)
				// 2. Explicit null check: x !== undefined && x !== null ? x : y

				var targetNode *ast.Node
				var skipTypeCheck bool

				// First check for simple pattern (where condition is exactly whenTrue)
				// This would catch cases like: x ? x : y  -> x ?? y
				// Compare after unwrapping parentheses and trimming text
				condUnwrapped := unwrapParentheses(condExpr.Condition)
				whenTrueUnwrapped := unwrapParentheses(condExpr.WhenTrue)
				isSimplePattern := areNodesTextuallyEqual(ctx.SourceFile, condUnwrapped, whenTrueUnwrapped)

				if isSimplePattern {
					// Use the unwrapped whenTrue node for type queries to avoid AST shape issues
					targetNode = unwrapParentheses(condExpr.WhenTrue)
					// Only proceed for identifier or member access like targets
					if !isMemberAccessLike(targetNode) && targetNode.Kind != ast.KindIdentifier {
						return
					}
					// If member/element access, reduce to the root object for type nullability checks
					if targetNode.Kind == ast.KindPropertyAccessExpression {
						pa := targetNode.AsPropertyAccessExpression()
						if pa != nil {
							targetNode = pa.Expression
						}
					} else if targetNode.Kind == ast.KindElementAccessExpression {
						ea := targetNode.AsElementAccessExpression()
						if ea != nil {
							targetNode = ea.Expression
						}
					}
					// For simple a ? a : b, prefer to report like TS-ESLint.
					// Skip full primitive-based eligibility; only ensure the type is potentially nullable when available.
					skipTypeCheck = true
					finalTarget := targetNode
					for finalTarget != nil && finalTarget.Kind == ast.KindParenthesizedExpression {
						paren := finalTarget.AsParenthesizedExpression()
						if paren == nil {
							break
						}
						finalTarget = paren.Expression
					}
					if tt := ctx.TypeChecker.GetTypeAtLocation(finalTarget); tt != nil {
						// Allow union with null/undefined to pass; if type info degrades, still report like TS-ESLint
						if !isNullableType(tt) {
							// Heuristic disabled: align with TS-ESLint by reporting simple pattern regardless of primitive ignores
							// Proceed without blocking
						}
					}
				} else {
					// Check for explicit null/undefined check patterns
					isExplicit, _ := isExplicitNullishCheck(condExpr.Condition, condExpr.WhenTrue, condExpr.WhenFalse, ctx.SourceFile)
					if isExplicit {
						// Combined explicit checks should be reported when ignoreTernaryTests is false
						skipTypeCheck = true
						targetNode = condExpr.WhenTrue
					} else {
						// Check for single nullish check (x !== undefined when type is string | undefined)
						isSingle, t := isSingleNullishCheck(condExpr.Condition, condExpr.WhenTrue, condExpr.WhenFalse, ctx)
						if isSingle {
							targetNode = t
							skipTypeCheck = true // Type check already done in isSingleNullishCheck
						} else {
							// Not a pattern we can convert to nullish coalescing
							return
						}
					}
				}

				// Check if the target is eligible for nullish coalescing
				if !skipTypeCheck {
					targetType := ctx.TypeChecker.GetTypeAtLocation(targetNode)
					if !isTypeEligibleForPreferNullish(targetType, opts) {
						// Debug: log why type check failed
						// fmt.Printf("Type check failed for: %s\n", getNodeText(ctx.SourceFile, targetNode))
						return
					}
				}

				// Check various ignore conditions
				if *opts.IgnoreConditionalTests && isConditionalTest(node) {
					return
				}

				// Check if this is a test in a ternary expression
				if *opts.IgnoreTernaryTests && node.Parent != nil && node.Parent.Kind == ast.KindConditionalExpression {
					if condExpr := node.Parent.AsConditionalExpression(); condExpr != nil && condExpr.Condition == node {
						return
					}
				}

				if *opts.IgnoreBooleanCoercion && isBooleanConstructorContext(node) {
					return
				}

				// Create fix suggestion
				targetText := strings.TrimSpace(getNodeText(ctx.SourceFile, targetNode))
				alternateText := strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.WhenFalse))

				var fixedAlternateText string
				if needsParentheses(condExpr.WhenFalse) {
					fixedAlternateText = fmt.Sprintf("(%s)", alternateText)
				} else {
					fixedAlternateText = alternateText
				}

				replacement := fmt.Sprintf("%s ?? %s", targetText, fixedAlternateText)

				ctx.ReportNodeWithSuggestions(node, buildPreferNullishOverTernaryMessage(),
					rule.RuleSuggestion{
						Message:  buildSuggestNullishMessage(),
						FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
					},
				)
			},

			// Handle if statements: if (!a) a = b;
			ast.KindIfStatement: func(node *ast.Node) {
				if *opts.IgnoreIfStatements {
					return
				}

				ifStmt := node.AsIfStatement()
				if ifStmt == nil || ifStmt.ElseStatement != nil {
					return
				}

				// Check if the if statement body is a simple assignment
				var assignmentExpr *ast.BinaryExpression
				switch ifStmt.ThenStatement.Kind {
				case ast.KindBlock:
					block := ifStmt.ThenStatement.AsBlock()
					if block == nil || block.Statements == nil || len(block.Statements.Nodes) != 1 {
						return
					}
					if block.Statements.Nodes[0].Kind == ast.KindExpressionStatement {
						exprStmt := block.Statements.Nodes[0].AsExpressionStatement()
						if exprStmt != nil && exprStmt.Expression.Kind == ast.KindBinaryExpression {
							binExpr := exprStmt.Expression.AsBinaryExpression()
							if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
								assignmentExpr = binExpr
							}
						}
					}
				case ast.KindExpressionStatement:
					exprStmt := ifStmt.ThenStatement.AsExpressionStatement()
					if exprStmt != nil && exprStmt.Expression.Kind == ast.KindBinaryExpression {
						binExpr := exprStmt.Expression.AsBinaryExpression()
						if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
							assignmentExpr = binExpr
						}
					}
				}

				if assignmentExpr == nil || !isMemberAccessLike(assignmentExpr.Left) {
					return
				}

				// Check if the condition is a simple nullish check for the same variable
				var conditionTarget *ast.Node
				if ifStmt.Expression.Kind == ast.KindPrefixUnaryExpression {
					prefixExpr := ifStmt.Expression.AsPrefixUnaryExpression()
					if prefixExpr != nil && prefixExpr.Operator == ast.KindExclamationToken {
						conditionTarget = prefixExpr.Operand
					}
				} else {
					// Handle other nullish check patterns like: if (a == null || a == undefined)
					conditionTarget = ifStmt.Expression
				}

				if conditionTarget == nil || !areNodesTextuallyEqual(ctx.SourceFile, conditionTarget, assignmentExpr.Left) {
					return
				}

				// Check if left operand is eligible for nullish coalescing
				leftType := ctx.TypeChecker.GetTypeAtLocation(assignmentExpr.Left)
				if !isTypeEligibleForPreferNullish(leftType, opts) {
					return
				}

				// Create fix suggestion
				leftText := strings.TrimSpace(getNodeText(ctx.SourceFile, assignmentExpr.Left))
				rightText := strings.TrimSpace(getNodeText(ctx.SourceFile, assignmentExpr.Right))
				replacement := fmt.Sprintf("%s ??= %s;", leftText, rightText)

				ctx.ReportNodeWithSuggestions(node, buildPreferNullishOverAssignmentMessage(),
					rule.RuleSuggestion{
						Message:  buildSuggestNullishMessage(),
						FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
					},
				)
			},
		}
	},
})
