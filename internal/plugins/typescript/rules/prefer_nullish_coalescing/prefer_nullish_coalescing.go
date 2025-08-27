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
		IgnoreMixedLogicalExpressions: utils.Ref(false),
		IgnorePrimitives: &PreferNullishPrimitivesOption{
			Boolean: utils.Ref(false),
			String:  utils.Ref(false),
			Number:  utils.Ref(false),
			Bigint:  utils.Ref(false),
		},
		IgnoreTernaryTests: utils.Ref(true),
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

func buildPreferNullishOverAssignmentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverAssignment",
		Description: "Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read.",
	}
}

func buildPreferNullishOverTernaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNullishOverTernary",
		Description: "Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read.",
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

	if ignorableFlags == 0 {
		return true // Any types are eligible for conversion
	}

	// If the type is any or unknown, we can't make assumptions about the value
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return false
	}

	// Check if any type constituents match the ignorable flags
	if utils.IsUnionType(t) {
		for _, unionType := range t.Types() {
			typeFlags := checker.Type_flags(unionType)
			if typeFlags&ignorableFlags != 0 {
				return false
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
	return isConditionalTestRecursive(node, make(map[*ast.Node]bool), 0)
}

func isConditionalTestRecursive(node *ast.Node, visited map[*ast.Node]bool, depth int) bool {
	// Prevent infinite recursion
	if depth > 10 || node == nil {
		return false
	}

	parent := node.Parent
	if parent == nil || visited[parent] {
		return false
	}
	visited[parent] = true

	switch parent.Kind {
	case ast.KindConditionalExpression:
		condExpr := parent.AsConditionalExpression()
		if condExpr != nil && condExpr.Condition == node {
			return true
		}
	case ast.KindIfStatement:
		ifStmt := parent.AsIfStatement()
		if ifStmt != nil && ifStmt.Expression == node {
			return true
		}
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement:
		return true
	case ast.KindParenthesizedExpression:
		// Check if the parenthesized expression is in a conditional context
		return isConditionalTestRecursive(parent, visited, depth+1)
	case ast.KindBinaryExpression:
		// Check if this is part of a logical expression that leads to a conditional
		binExpr := parent.AsBinaryExpression()
		if binExpr != nil && (binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
			binExpr.OperatorToken.Kind == ast.KindBarBarToken) {
			return isConditionalTestRecursive(parent, visited, depth+1)
		}
	case ast.KindPrefixUnaryExpression:
		prefixExpr := parent.AsPrefixUnaryExpression()
		if prefixExpr != nil && prefixExpr.Operator == ast.KindExclamationToken {
			return isConditionalTestRecursive(parent, visited, depth+1)
		}
	}

	return false
}

// isBooleanConstructorContext checks if a node is within a Boolean constructor context
func isBooleanConstructorContext(node *ast.Node) bool {
	visited := make(map[*ast.Node]bool)
	return isBooleanConstructorContextHelper(node, visited)
}

func isBooleanConstructorContextHelper(node *ast.Node, visited map[*ast.Node]bool) bool {
	parent := node.Parent
	if parent == nil || visited[parent] {
		return false
	}
	visited[parent] = true

	if parent.Kind == ast.KindCallExpression {
		callExpr := parent.AsCallExpression()
		if callExpr != nil && callExpr.Expression.Kind == ast.KindIdentifier {
			identifier := callExpr.Expression.AsIdentifier()
			if identifier != nil && identifier.Text == "Boolean" {
				return true
			}
		}
	}

	// Check parent contexts recursively
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binExpr := parent.AsBinaryExpression()
		if binExpr != nil && (binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken ||
			binExpr.OperatorToken.Kind == ast.KindBarBarToken) {
			return isBooleanConstructorContextHelper(parent, visited)
		}
	case ast.KindConditionalExpression:
		return isBooleanConstructorContextHelper(parent, visited)
	}

	return false
}

// isMixedLogicalExpression checks if a logical expression is mixed with && operators
func isMixedLogicalExpression(node *ast.Node) bool {
	seen := make(map[*ast.Node]bool)
	queue := []*ast.Node{node.Parent}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == nil || seen[current] {
			continue
		}
		seen[current] = true

		if current.Kind == ast.KindBinaryExpression {
			binExpr := current.AsBinaryExpression()
			if binExpr != nil {
				if binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
					return true
				}
				if binExpr.OperatorToken.Kind == ast.KindBarBarToken {
					queue = append(queue, current.Parent, binExpr.Left, binExpr.Right)
				}
			}
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

					ctx.ReportNodeWithSuggestions(binExpr.OperatorToken, buildPreferNullishOverAssignmentMessage(),
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
				// Simple case: a ? a : b (where condition and consequent are the same)
				if !areNodesTextuallyEqual(ctx.SourceFile, condExpr.Condition, condExpr.WhenTrue) {
					return
				}

				// Check if condition is eligible for nullish coalescing
				conditionType := ctx.TypeChecker.GetTypeAtLocation(condExpr.Condition)
				if !isTypeEligibleForPreferNullish(conditionType, opts) {
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
				conditionText := strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.Condition))
				alternateText := strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.WhenFalse))

				var fixedAlternateText string
				if needsParentheses(condExpr.WhenFalse) {
					fixedAlternateText = fmt.Sprintf("(%s)", alternateText)
				} else {
					fixedAlternateText = alternateText
				}

				replacement := fmt.Sprintf("%s ?? %s", conditionText, fixedAlternateText)

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
