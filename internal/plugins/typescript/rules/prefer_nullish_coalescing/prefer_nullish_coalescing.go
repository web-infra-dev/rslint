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
        // Default: do not ignore ternary tests unless explicitly configured.
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
    // Only consider types explicitly nullable (contain null or undefined)
    // This prevents flagging expressions like `bar || baz` when `bar` is `any`/`unknown`.
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
    // (Defensive: after the nullable check above, `any` alone won't reach here.)
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

	// Walk up the parent chain to check if this node is used as a test condition
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		
		// Skip over parenthesized expressions
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		
		// If parent is a logical expression, continue checking up the tree
		if parent.Kind == ast.KindBinaryExpression {
			binExpr := parent.AsBinaryExpression()
			if binExpr != nil && (binExpr.OperatorToken.Kind == ast.KindBarBarToken || 
			                       binExpr.OperatorToken.Kind == ast.KindAmpersandAmpersandToken) {
				current = parent
				continue
			}
		}

		// Check if we're in a test/condition position
		switch parent.Kind {
		case ast.KindConditionalExpression:
			// In a ternary expression, check if we're the test condition
			condExpr := parent.AsConditionalExpression()
			if condExpr != nil && isNodeWithin(condExpr.Condition, current) {
				return true
			}
			// If we're in the consequent or alternate, stop checking
			return false
			
		case ast.KindIfStatement:
			ifStmt := parent.AsIfStatement()
			if ifStmt != nil && isNodeWithin(ifStmt.Expression, current) {
				return true
			}
			return false
			
		case ast.KindWhileStatement:
			whileStmt := parent.AsWhileStatement()
			if whileStmt != nil && isNodeWithin(whileStmt.Expression, current) {
				return true
			}
			return false
			
		case ast.KindDoStatement:
			doStmt := parent.AsDoStatement()
			if doStmt != nil && isNodeWithin(doStmt.Expression, current) {
				return true
			}
			return false
			
		case ast.KindForStatement:
			forStmt := parent.AsForStatement()
			if forStmt != nil && isNodeWithin(forStmt.Condition, current) {
				return true
			}
			return false
		}

		// If we hit any other kind of parent, stop checking
		break
	}

	return false
}

// isWithinStatementCondition checks if a node is within a statement condition

// isDirectlyInStatementCondition returns true only when the binary expression is directly
// part of an if/while/do/for condition without passing through function/call/new boundaries.
func isDirectlyInStatementCondition(node *ast.Node) bool {
    if node == nil {
        return false
    }
    current := node
    for current != nil && current.Parent != nil {
        parent := current.Parent
        // Skip simple wrappers
        if parent.Kind == ast.KindParenthesizedExpression {
            current = parent
            continue
        }
        if parent.Kind == ast.KindBinaryExpression {
            if pb := parent.AsBinaryExpression(); pb != nil {
                switch pb.OperatorToken.Kind {
                case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken, ast.KindCommaToken:
                    current = parent
                    continue
                }
            }
        }
        // Disqualify when wrapped by unary operators, function/call/new boundaries
        switch parent.Kind {
        case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
            return false
        case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration,
            ast.KindCallExpression, ast.KindNewExpression:
            return false
        }
        // Check condition containers
        switch parent.Kind {
        case ast.KindIfStatement:
            if ifs := parent.AsIfStatement(); ifs != nil {
                return isNodeOrParentOf(ifs.Expression, current) || isNodeWithin(ifs.Expression, current)
            }
            return false
        case ast.KindWhileStatement:
            if ws := parent.AsWhileStatement(); ws != nil {
                return isNodeOrParentOf(ws.Expression, current) || isNodeWithin(ws.Expression, current)
            }
            return false
        case ast.KindDoStatement:
            if ds := parent.AsDoStatement(); ds != nil {
                return isNodeOrParentOf(ds.Expression, current) || isNodeWithin(ds.Expression, current)
            }
            return false
        case ast.KindForStatement:
            if fs := parent.AsForStatement(); fs != nil {
                return isNodeOrParentOf(fs.Condition, current) || isNodeWithin(fs.Condition, current)
            }
            return false
        }
        break
    }
    return false
}

// isWithinStatementCondition returns true if the given node appears within the condition
// of an if/while/do/for statement, regardless of intermediary operators and parentheses.
func isWithinStatementCondition(node *ast.Node) bool {
    if node == nil {
        return false
    }
    original := node
    current := node
    for current != nil && current.Parent != nil {
        parent := current.Parent
        // Crossing a function/call boundary means it's not directly within a statement condition
        switch parent.Kind {
        case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration,
            ast.KindCallExpression, ast.KindNewExpression:
            return false
        }
        switch parent.Kind {
        case ast.KindIfStatement:
            if ifStmt := parent.AsIfStatement(); ifStmt != nil {
                return isNodeOrParentOf(ifStmt.Expression, original) || isNodeWithin(ifStmt.Expression, original)
            }
            return false
        case ast.KindWhileStatement:
            if whileStmt := parent.AsWhileStatement(); whileStmt != nil {
                return isNodeOrParentOf(whileStmt.Expression, original) || isNodeWithin(whileStmt.Expression, original)
            }
            return false
        case ast.KindDoStatement:
            if doStmt := parent.AsDoStatement(); doStmt != nil {
                return isNodeOrParentOf(doStmt.Expression, original) || isNodeWithin(doStmt.Expression, original)
            }
            return false
        case ast.KindForStatement:
            if forStmt := parent.AsForStatement(); forStmt != nil {
                return isNodeOrParentOf(forStmt.Condition, original) || isNodeWithin(forStmt.Condition, original)
            }
            return false
        }
        current = parent
    }
    return false
}

// isNodeOrParentOf checks if target is the same as or a parent of node
func isNodeOrParentOf(target, node *ast.Node) bool {
	if target == nil || node == nil {
		return false
	}
	current := node
	// Traverse up from node to see if we reach target
	for current != nil {
		if current == target {
			return true
		}
		current = current.Parent
	}
	return false
}

// isNodeWithin checks if `node`'s range lies within `target`'s range (inclusive)
func isNodeWithin(target, node *ast.Node) bool {
    if target == nil || node == nil {
        return false
    }
    tStart, tEnd := target.Pos(), target.End()
    nStart, nEnd := node.Pos(), node.End()
    return nStart >= tStart && nEnd <= tEnd
}

// isBooleanConstructorContext checks if the node is within a context where it's being coerced to boolean (e.g., Boolean(expr))
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

// isWithinTernaryTestCondition checks if a node is within the test condition of a ternary expression
func isWithinTernaryTestCondition(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Walk up the parent chain to find a ternary expression
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent

		// Check if parent is a ternary expression
		if parent.Kind == ast.KindConditionalExpression {
			condExpr := parent.AsConditionalExpression()
			if condExpr != nil {
				// Check if the current node (which could be a parent of the original) is the condition
				// or if the original node is within the condition branch
				if condExpr.Condition == current || isNodeWithin(condExpr.Condition, node) {
					return true
				}
				// If we're in whenTrue or whenFalse, we're not in the condition
				if isNodeWithin(condExpr.WhenTrue, node) || isNodeWithin(condExpr.WhenFalse, node) {
					return false
				}
			}
		}

		current = parent
	}

	return false
}

// isWithinConditionalExpressionInStatementCondition returns true if `node` is inside
// a ConditionalExpression (ternary) that itself is used as the condition of
// an if/while/do/for statement, regardless of nesting depth.
func isWithinConditionalExpressionInStatementCondition(node *ast.Node) bool {
    if node == nil {
        return false
    }
    foundConditional := false
    current := node
    for current != nil && current.Parent != nil {
        parent := current.Parent
        if parent.Kind == ast.KindConditionalExpression {
            foundConditional = true
        }
        switch parent.Kind {
        case ast.KindIfStatement:
            if foundConditional {
                return true
            }
            return false
        case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement:
            if foundConditional {
                return true
            }
            return false
        }
        current = parent
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

// findRootOrExpression climbs up through chained `||` binary expressions to the topmost one.
func findRootOrExpression(node *ast.Node) *ast.Node {
    root := node
    for root.Parent != nil && root.Parent.Kind == ast.KindBinaryExpression {
        if parentBin := root.Parent.AsBinaryExpression(); parentBin != nil && parentBin.OperatorToken.Kind == ast.KindBarBarToken {
            root = root.Parent
        } else {
            break
        }
    }
    return root
}

// findLeftmostOrExpression walks down the left side of a chained `||` expression
// to locate the leftmost `||` operator in the chain.
func findLeftmostOrExpression(node *ast.Node) *ast.Node {
    // Start from the root `||` expression
    cur := findRootOrExpression(node)
    for cur != nil && cur.Kind == ast.KindBinaryExpression {
        bin := cur.AsBinaryExpression()
        if bin == nil || bin.OperatorToken.Kind != ast.KindBarBarToken {
            break
        }
        // If the left child is also a `||` expression, continue descending left
        if bin.Left != nil && bin.Left.Kind == ast.KindBinaryExpression {
            leftBin := bin.Left.AsBinaryExpression()
            if leftBin != nil && leftBin.OperatorToken.Kind == ast.KindBarBarToken {
                cur = bin.Left
                continue
            }
        }
        break
    }
    return cur
}

// hasOrAncestor returns true if there is a parent `||` above the given node
func hasOrAncestor(node *ast.Node) bool {
    for p := node.Parent; p != nil; p = p.Parent {
        if p.Kind != ast.KindBinaryExpression {
            continue
        }
        if pb := p.AsBinaryExpression(); pb != nil && pb.OperatorToken.Kind == ast.KindBarBarToken {
            return true
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

// findOperatorStart finds the start index of the given operator string following the left node,
// skipping over any whitespace/trivia between left and the operator. Falls back to a small window search.
func findOperatorStart(sourceFile *ast.SourceFile, left *ast.Node, operator string) int {
    if sourceFile == nil || left == nil {
        return 0
    }
    text := sourceFile.Text()
    start := left.End()
    i := start
    for i < len(text) {
        switch text[i] {
        case ' ', '\t', '\n', '\r':
            i++
        default:
            goto CHECK
        }
    }
CHECK:
    if i+len(operator) <= len(text) && text[i:i+len(operator)] == operator {
        return i
    }
    // Fallback: bounded search window
    windowEnd := start + 64
    if windowEnd > len(text) {
        windowEnd = len(text)
    }
    if idx := strings.Index(text[start:windowEnd], operator); idx >= 0 {
        return start + idx
    }
    return start
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

	switch binExpr.OperatorToken.Kind {
	case ast.KindExclamationEqualsEqualsToken:
		// Check for x !== undefined or x !== null (strict inequality)
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
	case ast.KindEqualsEqualsEqualsToken:
		// Check for x === undefined or x === null (strict equality)
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
			diffTypes := checksAreDifferentNullishTypes(binExpr.Left, binExpr.Right)
			if !diffTypes {
				return false, nil
			}

			leftTarget := getNullishCheckTarget(binExpr.Left, sourceFile, false)
			rightTarget := getNullishCheckTarget(binExpr.Right, sourceFile, false)

			// Unwrap parentheses from whenTrue for comparison
			whenTrueUnwrapped := unwrapParentheses(whenTrue)
			
			if leftTarget != nil && rightTarget != nil {
				// Check if both targets refer to the same identifier/member access
				targetsEqual := areNodesSemanticallyEqual(leftTarget, rightTarget)
				targetEqualsWhenTrue := areNodesSemanticallyEqual(leftTarget, whenTrueUnwrapped)
				
				// Also try comparing text directly as a fallback
				if (!targetsEqual || !targetEqualsWhenTrue) && sourceFile != nil {
					leftText := getNodeText(sourceFile, leftTarget)
					rightText := getNodeText(sourceFile, rightTarget)
					whenTrueText := getNodeText(sourceFile, whenTrueUnwrapped)
					
					if leftText == rightText && leftText == whenTrueText && leftText != "" {
						return true, leftTarget
					}
				}
				
				if targetsEqual && targetEqualsWhenTrue {
					return true, leftTarget
				}
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
			
			// Unwrap parentheses from whenFalse for comparison
			whenFalseUnwrapped := unwrapParentheses(whenFalse)
			
			if leftTarget != nil && rightTarget != nil {
				targetsEqual := areNodesSemanticallyEqual(leftTarget, rightTarget)
				targetEqualsWhenFalse := areNodesSemanticallyEqual(leftTarget, whenFalseUnwrapped)
				
				// Also try comparing text directly as a fallback
				if (!targetsEqual || !targetEqualsWhenFalse) && sourceFile != nil {
					leftText := getNodeText(sourceFile, leftTarget)
					rightText := getNodeText(sourceFile, rightTarget)
					whenFalseText := getNodeText(sourceFile, whenFalseUnwrapped)
					
					if leftText == rightText && leftText == whenFalseText && leftText != "" {
						return true, leftTarget
					}
				}
				
				if targetsEqual && targetEqualsWhenFalse {
					return true, leftTarget
				}
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
		switch binExpr.OperatorToken.Kind {
		case ast.KindExclamationEqualsToken:
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
		case ast.KindEqualsEqualsToken:
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
		// Strict equality checks (!==, ===) for single null/undefined should NOT be matched
		// They should only be matched when both null AND undefined are checked together
		case ast.KindExclamationEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
			// Don't match single strict equality checks - these need both null AND undefined checks
			return false, nil
		}
	}

	return false, nil
}

// areNodesSemanticallyEqual checks if two nodes are structurally the same identifier/member access
// This function is more lenient than exact equality - it considers nodes equal if they represent
// the same member access path, even if one uses optional chaining and the other doesn't.
// For example: a.b.c and a?.b.c would be considered semantically equal.
func areNodesSemanticallyEqual(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return false
	}

	// Unwrap parenthesized expressions first
	a = unwrapParentheses(a)
	b = unwrapParentheses(b)

	// For property access, check if both access the same property regardless of optional chaining
	if a.Kind == ast.KindPropertyAccessExpression && b.Kind == ast.KindPropertyAccessExpression {
		pa := a.AsPropertyAccessExpression()
		pb := b.AsPropertyAccessExpression()
		if pa != nil && pb != nil {
			// Check if the property names match
			if pa.Name() != nil && pb.Name() != nil {
				nameA := pa.Name().AsIdentifier()
				nameB := pb.Name().AsIdentifier()
				if nameA != nil && nameB != nil && nameA.Text == nameB.Text {
					// Recursively check the base expressions
					return areNodesSemanticallyEqual(pa.Expression, pb.Expression)
				}
			}
		}
		return false
	}

	// For element access, check if both access with the same key
	if a.Kind == ast.KindElementAccessExpression && b.Kind == ast.KindElementAccessExpression {
		ea := a.AsElementAccessExpression()
		eb := b.AsElementAccessExpression()
		if ea != nil && eb != nil {
			// Check if the keys match and base expressions match
			return areNodesSemanticallyEqual(ea.Expression, eb.Expression) &&
				areNodesSemanticallyEqual(ea.ArgumentExpression, eb.ArgumentExpression)
		}
		return false
	}

	// Allow property access and element access to match if they access the same member
	// e.g., obj["prop"] and obj.prop
	if (a.Kind == ast.KindPropertyAccessExpression && b.Kind == ast.KindElementAccessExpression) ||
		(a.Kind == ast.KindElementAccessExpression && b.Kind == ast.KindPropertyAccessExpression) {
		var propName string
		var baseA, baseB *ast.Node

		if a.Kind == ast.KindPropertyAccessExpression {
			pa := a.AsPropertyAccessExpression()
			if pa != nil && pa.Name() != nil && pa.Name().AsIdentifier() != nil {
				propName = pa.Name().AsIdentifier().Text
				baseA = pa.Expression
			}
			eb := b.AsElementAccessExpression()
			if eb != nil && eb.ArgumentExpression != nil && eb.ArgumentExpression.Kind == ast.KindStringLiteral {
				str := eb.ArgumentExpression.AsStringLiteral()
				if str != nil && str.Text == propName {
					baseB = eb.Expression
				}
			}
		} else {
			ea := a.AsElementAccessExpression()
			if ea != nil && ea.ArgumentExpression != nil && ea.ArgumentExpression.Kind == ast.KindStringLiteral {
				str := ea.ArgumentExpression.AsStringLiteral()
				if str != nil {
					propName = str.Text
					baseA = ea.Expression
				}
			}
			pb := b.AsPropertyAccessExpression()
			if pb != nil && pb.Name() != nil && pb.Name().AsIdentifier() != nil {
				if pb.Name().AsIdentifier().Text == propName {
					baseB = pb.Expression
				}
			}
		}

		if baseA != nil && baseB != nil {
			return areNodesSemanticallyEqual(baseA, baseB)
		}
		return false
	}

	// For other node types, require exact kind match
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case ast.KindIdentifier:
		aIdent := a.AsIdentifier()
		bIdent := b.AsIdentifier()
		if aIdent == nil || bIdent == nil {
			return false
		}
		return aIdent.Text == bIdent.Text
	case ast.KindThisKeyword:
		return true
	case ast.KindStringLiteral:
		aStr := a.AsStringLiteral()
		bStr := b.AsStringLiteral()
		if aStr == nil || bStr == nil {
			return false
		}
		return aStr.Text == bStr.Text
	case ast.KindNumericLiteral:
		aNum := a.AsNumericLiteral()
		bNum := b.AsNumericLiteral()
		if aNum == nil || bNum == nil {
			return false
		}
		return aNum.Text == bNum.Text
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
        // Fallback: if the global `strict` option is explicitly false, treat
        // strictNullChecks as disabled for the purposes of this rule. This
        // matches the intent of upstream tests that set only `strict: false`.
        if !isStrictNullChecks && compilerOptions.Strict.IsFalse() {
            isStrictNullChecks = false
        }

		if !isStrictNullChecks && (opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing == nil || !*opts.AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing) {
			ctx.ReportRange(core.NewTextRange(0, 0), buildNoStrictNullCheckMessage())
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			// Handle logical OR and logical OR assignment expressions
			ast.KindBinaryExpression: func(node *ast.Node) {
				binExpr := node.AsBinaryExpression()
				if binExpr == nil {
					return
				}

				// Handle logical OR assignment: a ||= b -> a ??= b
				if binExpr.OperatorToken.Kind == ast.KindBarBarEqualsToken {
					// Check if left operand is eligible for nullish coalescing
					leftType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Left)
					// Fallback to declared type if the location type isn't nullable
					if !isNullableType(leftType) && (binExpr.Left.Kind == ast.KindIdentifier || binExpr.Left.Kind == ast.KindPropertyAccessExpression) {
						if sym := ctx.TypeChecker.GetSymbolAtLocation(binExpr.Left); sym != nil {
							if declType := ctx.TypeChecker.GetTypeOfSymbol(sym); declType != nil {
								leftType = declType
							}
						}
					}

					if !isTypeEligibleForPreferNullish(leftType, opts) {
						return
					}

					leftText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Left))
					rightText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Right))
					replacement := fmt.Sprintf("%s ??= %s", leftText, rightText)

					// Report precisely on the '||=' operator token
					opStart := findOperatorStart(ctx.SourceFile, binExpr.Left, "||=")
					opRange := core.NewTextRange(opStart, opStart+3)
					// Match typescript-eslint: use the same message id as logical OR case
					ctx.ReportRangeWithSuggestions(opRange, buildPreferNullishOverOrMessage(),
						rule.RuleSuggestion{
							Message:  buildSuggestNullishMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
						},
					)
					return
				}

				// Handle logical OR expressions: a || b
				if binExpr.OperatorToken.Kind == ast.KindBarBarToken {
					// If this OR is the left child of a parent OR whose right subtree contains '&&'
					// and that parent has no OR ancestor (i.e., it's the top of the chain like `a || b || c && d`),
					// skip reporting here because the parent will emit both diagnostics in order.
					if node.Parent != nil && node.Parent.Kind == ast.KindBinaryExpression {
						if parentBin := node.Parent.AsBinaryExpression(); parentBin != nil && parentBin.OperatorToken.Kind == ast.KindBarBarToken {
							if !hasOrAncestor(node.Parent) && (hasAndOperator(parentBin.Right) || hasAndOperator(parentBin.Left)) {
								return
							}
						}
					}
                    // If this `||` is nested anywhere under a nullish coalescing expression, skip.
                    for p := node.Parent; p != nil; p = p.Parent {
                        if p.Kind == ast.KindBinaryExpression {
                            if pb := p.AsBinaryExpression(); pb != nil && pb.OperatorToken.Kind == ast.KindQuestionQuestionToken {
                                return
                            }
                        }
                    // Stop climbing at statement boundaries, and do a textual fallback
                    shouldBreak := false
                    switch p.Kind {
                    case ast.KindIfStatement:
                        if ifs := p.AsIfStatement(); ifs != nil {
                            if strings.Contains(getNodeText(ctx.SourceFile, ifs.Expression), "??") {
                                return
                            }
                        }
                        shouldBreak = true
                    case ast.KindWhileStatement:
                        if ws := p.AsWhileStatement(); ws != nil {
                            if strings.Contains(getNodeText(ctx.SourceFile, ws.Expression), "??") {
                                return
                            }
                        }
                        shouldBreak = true
                    case ast.KindDoStatement:
                        if ds := p.AsDoStatement(); ds != nil {
                            if strings.Contains(getNodeText(ctx.SourceFile, ds.Expression), "??") {
                                return
                            }
                        }
                        shouldBreak = true
                    case ast.KindForStatement:
                        if fs := p.AsForStatement(); fs != nil {
                            if strings.Contains(getNodeText(ctx.SourceFile, fs.Condition), "??") {
                                return
                            }
                        }
                        shouldBreak = true
                    }
                    if shouldBreak {
                        break
                    }
					}
					// For column/range accuracy, anchor reports to the current node's `||`.
					// The traversal is pre-order (parent before child), so the left child `||`
					// will be reported first, then the parent `||`, matching upstream ordering.
					anchorNode := node

                    // Operate (for ranges and text) on the anchor node, not necessarily the current node
                    anchorBin := anchorNode.AsBinaryExpression()
                    if anchorBin == nil {
                        return
                    }

                    // Check if left operand is eligible for nullish coalescing
                    leftType := ctx.TypeChecker.GetTypeAtLocation(anchorBin.Left)
                    // Fallback to declared type if the location type isn't nullable (can happen due to control flow)
                    if !isNullableType(leftType) && (anchorBin.Left.Kind == ast.KindIdentifier || anchorBin.Left.Kind == ast.KindPropertyAccessExpression) {
                        if sym := ctx.TypeChecker.GetSymbolAtLocation(anchorBin.Left); sym != nil {
                            if declType := ctx.TypeChecker.GetTypeOfSymbol(sym); declType != nil {
                                leftType = declType
                            }
                        }
                    }

					if !isTypeEligibleForPreferNullish(leftType, opts) {
						return
					}

                    // Check various ignore conditions with precedence rules:
                    // - If ignoreConditionalTests is true: ignore both statement and ternary tests,
                    //   unless ignoreTernaryTests is explicitly false (override to enable ternary checks).
                    // - If ignoreConditionalTests is false: do not ignore either context regardless of ignoreTernaryTests.
                    // - Otherwise (no explicit setting): respect ignoreTernaryTests for ternary-only ignoring.
                    inTernary := isWithinTernaryTestCondition(node)
                    inStmtCondDirect := isDirectlyInStatementCondition(node)
                    inStmtCondViaTernary := isWithinConditionalExpressionInStatementCondition(node)
                    if opts.IgnoreConditionalTests != nil && *opts.IgnoreConditionalTests {
                        // Always ignore statement conditions when enabled
                        if inStmtCondDirect || inStmtCondViaTernary {
                            return
                        }
                        // For ternary test positions: allow an explicit override via ignoreTernaryTests: false
                        if inTernary {
                            if !(opts.IgnoreTernaryTests != nil && !*opts.IgnoreTernaryTests) {
                                return
                            }
                        }
                    }

					if opts.IgnoreBooleanCoercion != nil && *opts.IgnoreBooleanCoercion && isBooleanConstructorContext(node) {
						return
					}

					if opts.IgnoreMixedLogicalExpressions != nil && *opts.IgnoreMixedLogicalExpressions && isMixedLogicalExpression(node) {
						return
					}

					// Special cases emitting two diagnostics in order for mixed expressions:
					// - `a || b || (c && d)` (right contains &&)
					// - `(a && b) || c || d` (left contains &&)
					if !hasOrAncestor(node) && (hasAndOperator(binExpr.Right) || hasAndOperator(binExpr.Left)) {
						if binExpr.Left != nil && binExpr.Left.Kind == ast.KindBinaryExpression {
							leftOr := binExpr.Left.AsBinaryExpression()
							if leftOr != nil && leftOr.OperatorToken.Kind == ast.KindBarBarToken {
								opStartLeft := findOperatorStart(ctx.SourceFile, leftOr.Left, "||")
								opRangeLeft := core.NewTextRange(opStartLeft, opStartLeft+2)
								ctx.ReportRange(opRangeLeft, buildPreferNullishOverOrMessage())
							}
						}
					}

					// Create fix suggestion
					leftText := strings.TrimSpace(getNodeText(ctx.SourceFile, anchorBin.Left))
					rightText := strings.TrimSpace(getNodeText(ctx.SourceFile, anchorBin.Right))

					var fixedRightText string
					if needsParentheses(binExpr.Right) {
						fixedRightText = fmt.Sprintf("(%s)", rightText)
					} else {
						fixedRightText = rightText
					}

					replacement := fmt.Sprintf("%s ?? %s", leftText, fixedRightText)

                    // Check if the entire expression needs parentheses
                    // When introducing '??' into a larger logical expression (either parent '&&' or '||'),
                    // parentheses are required to avoid mixing errors and to preserve evaluation order.
                    if node.Parent != nil && node.Parent.Kind == ast.KindBinaryExpression {
                        parentBinExpr := node.Parent.AsBinaryExpression()
                        if parentBinExpr != nil {
                            switch parentBinExpr.OperatorToken.Kind {
                            case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
                                replacement = fmt.Sprintf("(%s)", replacement)
                            }
                        }
                    }

                    // Report precisely on the '||' operator token, skipping leading trivia
                    opStart := findOperatorStart(ctx.SourceFile, anchorBin.Left, "||")
                    opRange := core.NewTextRange(opStart, opStart+2)
                    ctx.ReportRangeWithSuggestions(opRange, buildPreferNullishOverOrMessage(),
                        rule.RuleSuggestion{
                            Message:  buildSuggestNullishMessage(),
                            FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
                        },
                    )
					return
				}

				// Handle logical OR assignment expressions: a ||= b
				if binExpr.OperatorToken.Kind == ast.KindBarBarEqualsToken {
					// Check various ignore conditions (precedence handled below)

                    // Same ignore precedence for logical OR assignment in conditions
                    inTernary := isWithinTernaryTestCondition(node)
                    inStmtCondDirect := isDirectlyInStatementCondition(node)
                    inStmtCondViaTernary := isWithinConditionalExpressionInStatementCondition(node)
                    if opts.IgnoreConditionalTests != nil && *opts.IgnoreConditionalTests {
                        if inTernary || inStmtCondDirect || inStmtCondViaTernary {
                            return
                        }
                    }

					// After ignore handling, check eligibility. In ternary-test context we err on the
					// side of reporting (upstream does) because control-flow may narrow the type.
					leftType := ctx.TypeChecker.GetTypeAtLocation(binExpr.Left)
					if !isNullableType(leftType) && (binExpr.Left.Kind == ast.KindIdentifier || binExpr.Left.Kind == ast.KindPropertyAccessExpression) {
						if sym := ctx.TypeChecker.GetSymbolAtLocation(binExpr.Left); sym != nil {
							if declType := ctx.TypeChecker.GetTypeOfSymbol(sym); declType != nil {
								leftType = declType
							}
						}
					}
					if !isTypeEligibleForPreferNullish(leftType, opts) && !inTernary {
						return
					}

					if opts.IgnoreBooleanCoercion != nil && *opts.IgnoreBooleanCoercion && isBooleanConstructorContext(node) {
						return
					}

					// Create fix suggestion
					leftText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Left))
					rightText := strings.TrimSpace(getNodeText(ctx.SourceFile, binExpr.Right))
					replacement := fmt.Sprintf("%s ??= %s", leftText, rightText)

                    opStart := findOperatorStart(ctx.SourceFile, binExpr.Left, "||=")
                    opRange := core.NewTextRange(opStart, opStart+3)
                    ctx.ReportRangeWithSuggestions(opRange, buildPreferNullishOverOrMessage(),
                        rule.RuleSuggestion{
                            Message:  buildSuggestNullishMessage(),
                            FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
                        },
                    )
				}
			},

			// Handle ternary expressions: a ? a : b
			ast.KindConditionalExpression: func(node *ast.Node) {
				condExpr := node.AsConditionalExpression()
				if condExpr == nil {
					return
				}

				// Check if this is a nullish check pattern
				// Three cases to check:
				// 1. Simple case: a ? a : b (where condition and consequent are the same)
				// 2. Negated simple case: !a ? b : a (where negated condition and alternate are the same)
				// 3. Explicit null check: x !== undefined && x !== null ? x : y

				var targetNode *ast.Node
				var skipTypeCheck bool
				var isNegatedPattern bool
				var isExplicitPattern bool
				var isSimplePattern bool

				// Compare after unwrapping parentheses and trimming text
				condUnwrapped := unwrapParentheses(condExpr.Condition)
				whenTrueUnwrapped := unwrapParentheses(condExpr.WhenTrue)
				whenFalseUnwrapped := unwrapParentheses(condExpr.WhenFalse)

				// First check for explicit null/undefined check patterns
				// These should always be reported regardless of ignoreTernaryTests
				isExplicit, explicitTarget := isExplicitNullishCheck(condExpr.Condition, condExpr.WhenTrue, condExpr.WhenFalse, ctx.SourceFile)
				if isExplicit {
					isExplicitPattern = true
					skipTypeCheck = true
					targetNode = explicitTarget
				} else {
					// Check for single nullish check (x !== undefined when type is string | undefined)
					isSingle, t := isSingleNullishCheck(condExpr.Condition, condExpr.WhenTrue, condExpr.WhenFalse, ctx)
					if isSingle {
						isExplicitPattern = true
						targetNode = t
						skipTypeCheck = true // Type check already done in isSingleNullishCheck
					}
				}

				// If not an explicit pattern, check for simple patterns
				if !isExplicitPattern {
					isSimplePattern = areNodesSemanticallyEqual(condUnwrapped, whenTrueUnwrapped)

					// Check for negated pattern: !x ? y : x
					if !isSimplePattern && condUnwrapped != nil && condUnwrapped.Kind == ast.KindPrefixUnaryExpression {
						prefixUnary := condUnwrapped.AsPrefixUnaryExpression()
						if prefixUnary != nil && prefixUnary.Operator == ast.KindExclamationToken {
							negatedOperand := unwrapParentheses(prefixUnary.Operand)
							if areNodesSemanticallyEqual(negatedOperand, whenFalseUnwrapped) {
								isNegatedPattern = true
								targetNode = whenFalseUnwrapped
								// Only proceed for identifier or member access like targets
								if !isMemberAccessLike(targetNode) {
									return
								}
								skipTypeCheck = false
							}
						}
					}

					if isSimplePattern {
						// Use the unwrapped whenTrue node for type queries to avoid AST shape issues
						targetNode = unwrapParentheses(condExpr.WhenTrue)
						// Only proceed for identifier or member access like targets
						if !isMemberAccessLike(targetNode) {
							return
						}
						// For simple a ? a : b, check if the type is nullable
						skipTypeCheck = false
					} else if !isNegatedPattern {
						// Not a pattern we can convert to nullish coalescing
						return
					}
				}

				// Apply ignore conditions only for non-explicit patterns
				if !isExplicitPattern {
					// If we should ignore conditional tests, and this ternary is used inside
					// a statement condition (if/while/do/for), then skip reporting.
					if opts.IgnoreConditionalTests != nil && *opts.IgnoreConditionalTests && (isWithinStatementCondition(node) || isConditionalTest(node)) {
						return
					}
					// Check if we should ignore ternary tests
					// When ignoreTernaryTests is true, don't report on ternary expressions
					if opts.IgnoreTernaryTests != nil && *opts.IgnoreTernaryTests {
						return
					}
				}

				// Check if the target is eligible for nullish coalescing
				if !skipTypeCheck || isSimplePattern {
					targetType := ctx.TypeChecker.GetTypeAtLocation(targetNode)

					// For identifiers and property access, also try to get the declared type if the location type isn't nullable
					// This handles cases where TypeScript might optimize the type
					if (targetNode.Kind == ast.KindIdentifier || targetNode.Kind == ast.KindPropertyAccessExpression) && !isNullableType(targetType) {
						// Try getting the symbol's type
						symbol := ctx.TypeChecker.GetSymbolAtLocation(targetNode)
						if symbol != nil {
							declaredType := ctx.TypeChecker.GetTypeOfSymbol(symbol)
							if declaredType != nil {
								targetType = declaredType
							}
						}
					}

					if !isTypeEligibleForPreferNullish(targetType, opts) {
						return
					}
				}

				// Check various ignore conditions
				if opts.IgnoreBooleanCoercion != nil && *opts.IgnoreBooleanCoercion && isBooleanConstructorContext(node) {
					return
				}


				// Guard: only report on ternary patterns when the subject type is clearly nullable,
				// not when it's 'any' or 'unknown'. Determine the subject based on the pattern.
				{
					var subject *ast.Node
					if isSimplePattern {
						subject = condUnwrapped
					} else if isNegatedPattern {
						subject = whenFalseUnwrapped
					} else if targetNode != nil {
						subject = targetNode
					}
					if subject != nil {
						typ := ctx.TypeChecker.GetTypeAtLocation(subject)
						if typ != nil {
							flags := checker.Type_flags(typ)
							if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
								return
							}
						}
					}
				}

				// Create fix suggestion
				// For simple pattern, use the condition/whenTrue for text
				// For negated pattern, swap the order to get x ?? y
				var targetText, alternateText string
				if isSimplePattern {
					targetText = strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.Condition))
					alternateText = strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.WhenFalse))
				} else if isNegatedPattern {
					// For !x ? y : x pattern, we want to generate x ?? y
					targetText = strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.WhenFalse))
					alternateText = strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.WhenTrue))
				} else {
					targetText = strings.TrimSpace(getNodeText(ctx.SourceFile, targetNode))
					alternateText = strings.TrimSpace(getNodeText(ctx.SourceFile, condExpr.WhenFalse))
				}

				var fixedAlternateText string
				if isNegatedPattern {
					// For negated pattern, whenTrue becomes the alternate
					if needsParentheses(condExpr.WhenTrue) {
						fixedAlternateText = fmt.Sprintf("(%s)", alternateText)
					} else {
						fixedAlternateText = alternateText
					}
				} else {
					// For other patterns, whenFalse is the alternate
					if needsParentheses(condExpr.WhenFalse) {
						fixedAlternateText = fmt.Sprintf("(%s)", alternateText)
					} else {
						fixedAlternateText = alternateText
					}
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
				if opts.IgnoreIfStatements != nil && *opts.IgnoreIfStatements {
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
							if binExpr != nil && (binExpr.OperatorToken.Kind == ast.KindEqualsToken || binExpr.OperatorToken.Kind == ast.KindBarBarEqualsToken || binExpr.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken) {
								assignmentExpr = binExpr
							}
						}
					}
				case ast.KindExpressionStatement:
					exprStmt := ifStmt.ThenStatement.AsExpressionStatement()
					if exprStmt != nil && exprStmt.Expression.Kind == ast.KindBinaryExpression {
						binExpr := exprStmt.Expression.AsBinaryExpression()
						if binExpr != nil && (binExpr.OperatorToken.Kind == ast.KindEqualsToken || binExpr.OperatorToken.Kind == ast.KindBarBarEqualsToken || binExpr.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken) {
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
