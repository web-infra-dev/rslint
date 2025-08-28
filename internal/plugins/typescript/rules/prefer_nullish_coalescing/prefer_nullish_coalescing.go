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

	// If the type is any or unknown, we can't make assumptions about the value
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
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

	// Check for complex types that should be ignored (intersection types, branded types, etc.)
	// Do this check first, as intersection types may not have the expected primitive flags
	if utils.IsUnionType(t) {
		for _, unionType := range t.Types() {
			// Skip null and undefined types
			unionFlags := checker.Type_flags(unionType)
			if unionFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
				continue
			}

			// Check if this union constituent is an intersection type (branded type)
			// Use multiple methods to detect intersection/branded types
			if utils.IsIntersectionType(unionType) {
				return false // Branded/intersection types should be ignored
			}

			// Alternative check: if the type doesn't have standard primitive flags but has complex structure
			if unionFlags == 0 || (unionFlags&^(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike)) != 0 {
				// This might be a complex type like an intersection type that should be ignored
				return false
			}
		}
	} else {
		// Check if the non-union type is an intersection type
		if utils.IsIntersectionType(t) {
			return false
		}

		// Alternative check for non-union complex types
		if flags == 0 || (flags&^(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined)) != 0 {
			return false
		}
	}

	// Check if any type constituents match the ignorable flags
	if ignorableFlags != 0 {
		if utils.IsUnionType(t) {
			for _, unionType := range t.Types() {
				// Skip null and undefined types
				unionFlags := checker.Type_flags(unionType)
				if unionFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
					continue
				}

				// Check if this constituent matches ignorable flags
				if unionFlags&ignorableFlags != 0 {
					return false
				}

				// Special handling for intersection types that may contain ignored primitives
				// Note: intersection types (branded types) may not have the expected primitive flags,
				// so we need to check for them explicitly when primitive types are being ignored
				if utils.IsIntersectionType(unionType) {
					// If we're ignoring any primitive types and this is an intersection type,
					// it's likely a branded type that should be ignored
					return false
				}
			}
		} else {
			// For non-union types, check if they match ignorable flags
			if flags&ignorableFlags != 0 {
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
func containsNode(container, target *ast.Node) bool {
	if container == nil || target == nil {
		return false
	}
	if container == target {
		return true
	}

	// Simple traversal up from target to see if container is an ancestor
	current := target
	for current != nil {
		if current == container {
			return true
		}
		current = current.Parent
	}

	return false
}

// nodeContainsRecursive performs a depth-first search to find target within container
func nodeContainsRecursive(container, target *ast.Node, depth int) bool {
	if depth > 10 || container == nil {
		return false
	}

	if container == target {
		return true
	}

	// Check all children - this is a simplified check,
	// in a real implementation we'd traverse all child nodes
	// For now, let's check key structural elements
	switch container.Kind {
	case ast.KindBinaryExpression:
		binExpr := container.AsBinaryExpression()
		if binExpr != nil {
			return nodeContainsRecursive(binExpr.Left, target, depth+1) ||
				nodeContainsRecursive(binExpr.Right, target, depth+1)
		}
	case ast.KindParenthesizedExpression:
		parenExpr := container.AsParenthesizedExpression()
		if parenExpr != nil {
			return nodeContainsRecursive(parenExpr.Expression, target, depth+1)
		}
	}

	return false
}

// isAssignmentContext checks if a node is in an assignment context
func isAssignmentContext(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check up to 3 levels to avoid infinite recursion
	for i := 0; i < 3 && node != nil; i++ {
		switch node.Kind {
		case ast.KindVariableDeclaration, ast.KindVariableStatement:
			return true
		case ast.KindBinaryExpression:
			binExpr := node.AsBinaryExpression()
			if binExpr != nil && binExpr.OperatorToken.Kind == ast.KindEqualsToken {
				return true
			}
		}
		node = node.Parent
	}

	return false
}

// isBooleanConstructorContext checks if a node is within a Boolean constructor context
func isBooleanConstructorContext(node *ast.Node) bool {
	// Check up to 5 levels to avoid infinite recursion
	for i := 0; i < 5 && node != nil; i++ {
		parent := node.Parent
		if parent == nil {
			return false
		}

		if parent.Kind == ast.KindCallExpression {
			callExpr := parent.AsCallExpression()
			if callExpr != nil && callExpr.Expression.Kind == ast.KindIdentifier {
				identifier := callExpr.Expression.AsIdentifier()
				if identifier != nil && identifier.Text == "Boolean" {
					return true
				}
			}
		}

		// Only traverse through logical expressions and conditionals
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binExpr := parent.AsBinaryExpression()
			if binExpr == nil || (binExpr.OperatorToken.Kind != ast.KindAmpersandAmpersandToken &&
				binExpr.OperatorToken.Kind != ast.KindBarBarToken) {
				return false
			}
		case ast.KindConditionalExpression:
			// Continue checking
		default:
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
