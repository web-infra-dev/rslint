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

		// If we find a conditional statement, check if we're in its condition
		switch parent.Kind {
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

func isExplicitNullishCheck(condition, whenTrue, whenFalse *ast.Node, sourceFile *ast.SourceFile) (bool, *ast.Node) {
	condition = unwrapParentheses(condition)

	if condition == nil {
		return false, nil
	}

	// Pattern 1: x !== undefined && x !== null ? x : y
	if condition.Kind == ast.KindBinaryExpression {
		binExpr := condition.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			leftTarget := getNullishCheckTarget(binExpr.Left, sourceFile, false)
			rightTarget := getNullishCheckTarget(binExpr.Right, sourceFile, false)
			if leftTarget != nil && rightTarget != nil &&
				areNodesTextuallyEqual(sourceFile, leftTarget, rightTarget) &&
				areNodesTextuallyEqual(sourceFile, leftTarget, whenTrue) {
				return true, leftTarget
			}
		}
	}

	// Pattern 2: x === undefined || x === null ? y : x
	if condition.Kind == ast.KindBinaryExpression {
		binExpr := condition.AsBinaryExpression()
		if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindBarBarToken {
			leftTarget := getNullishCheckTarget(binExpr.Left, sourceFile, true)
			rightTarget := getNullishCheckTarget(binExpr.Right, sourceFile, true)
			if leftTarget != nil && rightTarget != nil &&
				areNodesTextuallyEqual(sourceFile, leftTarget, rightTarget) &&
				areNodesTextuallyEqual(sourceFile, leftTarget, whenFalse) {
				return true, leftTarget
			}
		}
	}

	// Pattern 3: x != null or x !== null
	if condition.Kind == ast.KindBinaryExpression {
		binExpr := condition.AsBinaryExpression()
		if getNullishCheckTarget(condition, sourceFile, false) != nil &&
			areNodesTextuallyEqual(sourceFile, binExpr.Left, whenTrue) {
			return true, binExpr.Left
		}
		if getNullishCheckTarget(condition, sourceFile, true) != nil &&
			areNodesTextuallyEqual(sourceFile, binExpr.Left, whenFalse) {
			return true, binExpr.Left
		}
	}

	return false, nil
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
	if bin.Right.Kind == ast.KindNullKeyword ||
		(bin.Right.Kind == ast.KindIdentifier && bin.Right.AsIdentifier() != nil && bin.Right.AsIdentifier().Text == "undefined") {
		return bin.Left
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
				if *opts.IgnoreTernaryTests {
					return
				}

				condExpr := node.AsConditionalExpression()
				if condExpr == nil {
					return
				}

				// Check if this is a nullish check pattern
				// Two cases to check:
				// 1. Simple case: a ? a : b (where condition and consequent are the same)
				// 2. Explicit null check: x !== undefined && x !== null ? x : y

				var targetNode *ast.Node
				isSimplePattern := areNodesTextuallyEqual(ctx.SourceFile, condExpr.Condition, condExpr.WhenTrue)

				if isSimplePattern {
					targetNode = condExpr.Condition
				} else {
					// Check for explicit null/undefined check pattern (both normal and reverse)
					if ok, t := isExplicitNullishCheck(condExpr.Condition, condExpr.WhenTrue, condExpr.WhenFalse, ctx.SourceFile); ok {
						targetNode = t
					} else {
						return
					}
				}

				// Check if the target is eligible for nullish coalescing
				targetType := ctx.TypeChecker.GetTypeAtLocation(targetNode)
				if !isTypeEligibleForPreferNullish(targetType, opts) {
					return
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
