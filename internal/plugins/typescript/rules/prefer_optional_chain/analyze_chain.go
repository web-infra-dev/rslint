package prefer_optional_chain

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ChainAnalyzer struct {
	ctx        rule.RuleContext
	opts       PreferOptionalChainOptions
	sourceText string
}

func NewChainAnalyzer(ctx rule.RuleContext, opts PreferOptionalChainOptions) *ChainAnalyzer {
	return &ChainAnalyzer{
		ctx:        ctx,
		opts:       opts,
		sourceText: ctx.SourceFile.Text(),
	}
}

func (ca *ChainAnalyzer) AnalyzeChain(
	operands []Operand,
	operator ast.Kind,
	parentNode *ast.Node,
) {
	if len(operands) < 2 {
		return
	}

	// For && chains: each operand should be a valid guard, and consecutive operands
	// should form subset relationships.
	// For || chains: each operand (negated) should form subset relationships.

	i := 0
	for i < len(operands)-1 {
		// Find the start of a potential chain
		chainStart := i
		if operands[i].Validity != OperandValid {
			i++
			continue
		}

		// Try to extend the chain
		chainEnd := i + 1
		for chainEnd < len(operands) {
			curr := operands[chainEnd-1]
			next := operands[chainEnd]

			if next.Validity != OperandValid {
				break
			}

			if curr.ComparedNode == nil || next.ComparedNode == nil {
				break
			}

			// Strict equality checks (!== null, !== undefined) only guard against ONE
			// nullish value. If the type includes the OTHER nullish value, the guard
			// is insufficient for optional chaining.
			if !ca.isGuardSufficientForChain(curr) {
				break
			}

			cmp := compareNodesUncached(curr.ComparedNode, next.ComparedNode)
			if cmp != NodeComparisonSubset && cmp != NodeComparisonEqual {
				break
			}

			chainEnd++
		}

		if chainEnd-chainStart < 2 {
			// Need at least 2 operands to form a chain
			i++
			continue
		}

		// Check the last operand - it should actually access a property
		lastOperand := operands[chainEnd-1]
		if lastOperand.ComparedNode == nil {
			i++
			continue
		}

		// For && chains, only process if the last operand is an actual chain target (not just an equality check)
		// The last operand must extend beyond the second-to-last operand
		secondToLast := operands[chainEnd-2]
		if secondToLast.ComparedNode != nil {
			cmp := compareNodesUncached(secondToLast.ComparedNode, lastOperand.ComparedNode)
			if cmp == NodeComparisonEqual && chainEnd-chainStart == 2 {
				// Two equal nodes don't form a chain
				i++
				continue
			}
		}

		// Skip if the first operand is a bare this/super keyword (not a property access)
		// e.g., `this && this.foo` should not be flagged
		firstCompared := operands[chainStart].ComparedNode
		if firstCompared != nil {
			fc := ast.SkipParentheses(firstCompared)
			if fc.Kind == ast.KindThisKeyword || fc.Kind == ast.KindSuperKeyword {
				i = chainEnd
				continue
			}
		}

		// For || chains, restore the last operand's original ComparisonType
		// for comparison operators. These were inverted by isNegatedComparison()
		// during DeMorgan classification, but output generation (wrapChainCode)
		// and fix/suggest decision need the original source operator.
		// Boolean types (ComparisonBoolean/ComparisonNotBoolean) are set directly
		// and should NOT be de-inverted.
		if operator == ast.KindBarBarToken {
			ct := operands[chainEnd-1].ComparisonType
			if ct != ComparisonBoolean && ct != ComparisonNotBoolean {
				operands[chainEnd-1].ComparisonType = invertComparisonType(ct)
			}
		}

		// Skip chains where the last operand's strict comparison would change
		// truthiness when the optional chain returns undefined.
		// e.g., `data && data.value !== null` → `data?.value !== null` changes
		// from falsy to true when data is null/undefined.
		if wouldChangeTruthiness(operands[chainStart:chainEnd], operator) {
			i = chainEnd
			continue
		}

		ca.reportChain(operands[chainStart:chainEnd], operator, parentNode)
		i = chainEnd
	}
}

func (ca *ChainAnalyzer) reportChain(
	operands []Operand,
	operator ast.Kind,
	parentNode *ast.Node,
) {
	if len(operands) < 2 {
		return
	}

	// Check requireNullish option
	if ca.shouldSkipForRequireNullish(operands) {
		return
	}

	// For chains with strict equality guards, ensure at least one guard is
	// meaningful (its compared node's type actually includes the nullish value
	// being checked). This prevents reporting chains where all guards are
	// vacuous (e.g., `foo.bar !== undefined && foo.bar()` where foo.bar is
	// always a function and never undefined).
	if ca.hasOnlyVacuousStrictGuards(operands[:len(operands)-1]) {
		return
	}

	// Build the optional chain expression
	fixCode := ca.buildOptionalChainCode(operands, operator)
	if fixCode == "" {
		return
	}

	// Wrap the chain code based on the last operand's comparison type
	lastOperand := operands[len(operands)-1]
	fixCode = wrapChainCode(fixCode, lastOperand)

	// Find the binary expression that encompasses all operands, covering any
	// enclosing parentheses (e.g., `a && (a.b && a.b.c)` → the outer `&&`).
	firstNode := operands[0].Node
	reportNode := findBinaryExpressionCovering(parentNode, firstNode, operands[len(operands)-1].Node)
	if reportNode == nil {
		reportNode = parentNode
	}

	// Compute the fix range start by walking up from firstNode through any
	// ParenthesizedExpressions to include opening parens in the range.
	// E.g., `(a && a.b) && a.b.c` → startNode is the ParenthesizedExpression `(a && a.b)`.
	// But for `foo?.a && bar && bar.a` → startNode stays as `bar` (no wrapping parens).
	startNode := firstNode
	n := firstNode.Parent
	for n != nil && n != reportNode.Parent {
		if ast.IsParenthesizedExpression(n) {
			startNode = n
		}
		n = n.Parent
	}
	reportRange := utils.TrimNodeTextRange(ca.ctx.SourceFile, startNode).WithEnd(reportNode.End())

	// Determine if we should use fix or suggestion
	useFix := ca.shouldUseFix(operands)

	msg := buildPreferOptionalChainMessage()
	sugMsg := buildOptionalChainSuggestMessage()

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(reportRange, fixCode),
	}

	rule.ReportNodeWithFixesOrSuggestions(ca.ctx, reportNode, useFix, msg, sugMsg, fixes...)
}

func (ca *ChainAnalyzer) shouldUseFix(operands []Operand) bool {
	if derefBoolDefault(ca.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing, false) {
		return true
	}

	last := operands[len(operands)-1]

	switch last.ComparisonType {
	case ComparisonBoolean:
		// For bare truthy checks, auto-fix is safe when the original expression
		// can already return undefined (meaning the fix doesn't add a new undefined).
		// This requires a guard that's a bare truthy/negation check (ComparisonBoolean/
		// ComparisonNotBoolean) whose type includes undefined. Only bare truthy guards
		// let the expression short-circuit to the guard's actual value (which may be
		// undefined). Comparison guards (!= null, !== undefined, etc.) short-circuit
		// to boolean (true/false), so even if the guard's type includes undefined,
		// the expression can never return undefined.
		if ca.ctx.TypeChecker != nil {
			for _, guard := range operands[:len(operands)-1] {
				if guard.ComparisonType != ComparisonBoolean && guard.ComparisonType != ComparisonNotBoolean {
					continue
				}
				if guard.ComparedNode != nil {
					t := ca.ctx.TypeChecker.GetTypeAtLocation(guard.ComparedNode)
					if t != nil && typeIncludesUndefined(t) {
						return true
					}
				}
			}
		}
		// Fall through to type check on last operand

	case ComparisonNotBoolean:
		// !expr always returns boolean. !undefined = true, same as
		// original || short-circuit. Always safe.
		return true

	case ComparisonEqualNullOrUndefined:
		// == null/undefined: undefined == null is true. Matches original. Safe.
		return true

	case ComparisonNotEqualNullOrUndefined,
		ComparisonNotStrictEqualNull,
		ComparisonNotStrictEqualUndefined,
		ComparisonStrictEqualUndefined:
		// These all produce the same result when applied to undefined
		// as the original short-circuit behavior. Safe.
		return true

	case ComparisonStrictEqualNull:
		// === null: undefined === null is false, but original || chain
		// short-circuits to true. Semantics change! Not safe.
		return false
	}

	// Fallback: check if the result type already includes undefined
	if ca.ctx.TypeChecker != nil && last.ComparedNode != nil {
		t := ca.ctx.TypeChecker.GetTypeAtLocation(last.ComparedNode)
		if t != nil && typeIncludesUndefined(t) {
			return true
		}
	}

	return false
}

func (ca *ChainAnalyzer) shouldSkipForRequireNullish(operands []Operand) bool {
	if !derefBoolDefault(ca.opts.RequireNullish, false) {
		return false
	}
	if ca.ctx.TypeChecker == nil {
		return false
	}

	// With requireNullish, at least one guard operand must have null/undefined in its type.
	// Only check guard operands (all except the last), since the last operand is the
	// chain target, not a guard. E.g., in `foo && foo.bar`, foo is the guard.
	guards := operands
	if len(guards) > 1 {
		guards = operands[:len(operands)-1]
	}
	for _, op := range guards {
		if op.ComparedNode == nil {
			continue
		}
		t := ca.ctx.TypeChecker.GetTypeAtLocation(op.ComparedNode)
		if t == nil {
			continue
		}
		for _, part := range utils.UnionTypeParts(t) {
			flags := checker.Type_flags(part)
			if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
				return false // has nullish -> don't skip
			}
		}
	}
	return true // no nullish found in any operand -> skip
}

// isGuardSufficientForChain checks whether a strict equality guard fully covers
// the nullish types of the operand. Loose equality (!= / ==) catches both null
// and undefined, so it's always sufficient. Strict equality (!== / ===) only
// catches one, so we need to verify the type doesn't include the other.
func (ca *ChainAnalyzer) isGuardSufficientForChain(guard Operand) bool {
	switch guard.ComparisonType {
	case ComparisonBoolean, ComparisonNotBoolean:
		return true
	case ComparisonNotEqualNullOrUndefined, ComparisonEqualNullOrUndefined:
		return true
	case ComparisonNotStrictEqualNull, ComparisonStrictEqualNull:
		// !== null / === null only covers null. If type also has undefined, not sufficient.
		return !ca.nodeTypeIncludesUndefined(guard.ComparedNode)
	case ComparisonNotStrictEqualUndefined, ComparisonStrictEqualUndefined:
		// !== undefined / === undefined only covers undefined. If type also has null, not sufficient.
		return !ca.nodeTypeIncludesNull(guard.ComparedNode)
	}
	return true
}

// hasOnlyVacuousStrictGuards returns true if all guards in the given operands
// use strict equality checks and none of them have a type that actually includes
// the nullish value being checked. A vacuous guard like `foo !== undefined` where
// `foo` is always a non-nullable type should not trigger the rule.
func (ca *ChainAnalyzer) hasOnlyVacuousStrictGuards(guards []Operand) bool {
	if ca.ctx.TypeChecker == nil {
		return false // can't determine without type info
	}
	for _, guard := range guards {
		switch guard.ComparisonType {
		case ComparisonBoolean, ComparisonNotBoolean,
			ComparisonNotEqualNullOrUndefined, ComparisonEqualNullOrUndefined:
			return false // non-strict guards are always meaningful
		case ComparisonNotStrictEqualNull, ComparisonStrictEqualNull:
			if ca.guardTypeIsMeaningfulForNull(guard.ComparedNode) {
				return false
			}
		case ComparisonNotStrictEqualUndefined, ComparisonStrictEqualUndefined:
			if ca.guardTypeIsMeaningfulForUndefined(guard.ComparedNode) {
				return false
			}
		}
	}
	return true // all guards are vacuous
}

// guardTypeIsMeaningfulForNull checks if a node's type includes null, any, or unknown.
// Used by hasOnlyVacuousStrictGuards to determine if a strict null guard is meaningful.
func (ca *ChainAnalyzer) guardTypeIsMeaningfulForNull(node *ast.Node) bool {
	if ca.ctx.TypeChecker == nil || node == nil {
		return false
	}
	t := ca.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)
		if flags&(checker.TypeFlagsNull|checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return true
		}
	}
	return false
}

// guardTypeIsMeaningfulForUndefined checks if a node's type includes undefined/void, any, or unknown.
// Used by hasOnlyVacuousStrictGuards to determine if a strict undefined guard is meaningful.
func (ca *ChainAnalyzer) guardTypeIsMeaningfulForUndefined(node *ast.Node) bool {
	if ca.ctx.TypeChecker == nil || node == nil {
		return false
	}
	t := ca.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)
		if flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return true
		}
	}
	return false
}

func (ca *ChainAnalyzer) nodeTypeIncludesUndefined(node *ast.Node) bool {
	if ca.ctx.TypeChecker == nil || node == nil {
		return false
	}
	t := ca.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)
		if flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return true
		}
	}
	return false
}

func (ca *ChainAnalyzer) nodeTypeIncludesNull(node *ast.Node) bool {
	if ca.ctx.TypeChecker == nil || node == nil {
		return false
	}
	t := ca.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)
		if flags&checker.TypeFlagsNull != 0 {
			return true
		}
	}
	return false
}

// wouldChangeTruthiness checks if the last operand's strict comparison wrapper
// would evaluate to `true` when given `undefined` (which optional chain returns
// for null/undefined), AND at least one guard is a truthy check (which means
// the original short-circuits to a falsy value). This would change the
// expression's truthiness, making the transformation unsafe.
func wouldChangeTruthiness(operands []Operand, operator ast.Kind) bool {
	last := operands[len(operands)-1]

	// Only applicable when the last operand has a comparison wrapper
	if last.ComparisonType == ComparisonBoolean || last.ComparisonType == ComparisonNotBoolean {
		return false
	}

	// Check if any guard is a truthy check
	hasTruthyGuard := false
	for _, op := range operands[:len(operands)-1] {
		if op.ComparisonType == ComparisonBoolean || op.ComparisonType == ComparisonNotBoolean {
			hasTruthyGuard = true
			break
		}
	}
	if !hasTruthyGuard {
		return false
	}

	// For && chains: check if `undefined <op> value` would be `true`
	if operator == ast.KindAmpersandAmpersandToken {
		switch last.ComparisonType {
		case ComparisonNotStrictEqualNull:
			// undefined !== null → true (BAD: changes from falsy to truthy)
			return true
		case ComparisonStrictEqualUndefined:
			// undefined === undefined → true (BAD: changes from falsy to truthy)
			return true
		}
	}

	// For || chains (last operand has been de-inverted to its original source operator):
	// check if `undefined <op> value` would be `false`
	// (since || short-circuits on truthy, changing to falsy is problematic)
	if operator == ast.KindBarBarToken {
		switch last.ComparisonType {
		case ComparisonStrictEqualNull:
			// foo === null || ...: undefined === null → false (BAD: original is truthy)
			return true
		case ComparisonNotStrictEqualUndefined:
			// foo !== undefined || ...: undefined !== undefined → false (BAD: original is truthy)
			return true
		}
	}

	return false
}

func typeIncludesUndefined(t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)
		// any subsumes all types including undefined
		if flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny) != 0 {
			return true
		}
	}
	return false
}

func (ca *ChainAnalyzer) buildOptionalChainCode(operands []Operand, operator ast.Kind) string {
	if len(operands) < 2 {
		return ""
	}

	// The last operand is the actual expression we want to convert to optional chain
	lastOperand := operands[len(operands)-1]
	if lastOperand.ComparedNode == nil {
		return ""
	}

	// Flatten the last operand's node into a chain of accesses
	parts := flattenChainExpression(lastOperand.ComparedNode)
	if len(parts) == 0 {
		return ca.getNodeText(lastOperand.ComparedNode)
	}

	// Determine which part indices need ?. insertion
	// A guard matching at part[j] means: the access at part[j+1] should use ?.
	optionalIndices := make(map[int]bool)
	for i := range len(operands) - 1 {
		guard := operands[i]
		if guard.ComparedNode == nil {
			continue
		}

		for j, part := range parts {
			cmp := compareNodesUncached(guard.ComparedNode, part.node)
			if cmp == NodeComparisonEqual {
				// Guard matches this part exactly, so the NEXT access needs ?.
				if j+1 < len(parts) {
					optionalIndices[j+1] = true
				}
				break
			}
		}
	}

	// If no optional indices were found, try a simpler approach:
	// the first guard is a prefix of the last operand
	if len(optionalIndices) == 0 {
		firstGuard := operands[0]
		if firstGuard.ComparedNode != nil {
			for j, part := range parts {
				cmp := compareNodesUncached(firstGuard.ComparedNode, part.node)
				if cmp == NodeComparisonEqual && j+1 < len(parts) {
					optionalIndices[j+1] = true
					break
				}
			}
		}
	}

	// Build the output with ?. inserted at appropriate positions
	var sb strings.Builder

	// Write the base expression (the root of the chain)
	if len(parts) > 0 {
		baseNode := parts[0].node
		baseText := ca.getNodeText(baseNode)

		if needsParensAsBase(baseNode) {
			sb.WriteString("(")
			sb.WriteString(baseText)
			sb.WriteString(")")
		} else {
			sb.WriteString(baseText)
		}
	}

	// Write each accessor, inserting ?. where needed
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		needsOptional := optionalIndices[i] && !part.isAlreadyOptional

		switch part.accessKind {
		case accessKindProperty:
			if needsOptional || part.isAlreadyOptional {
				sb.WriteString("?.")
			} else {
				sb.WriteString(".")
			}
			sb.WriteString(part.accessName)

		case accessKindElement:
			if needsOptional || part.isAlreadyOptional {
				sb.WriteString("?.")
			}
			sb.WriteString("[")
			sb.WriteString(ca.getNodeText(part.accessArgument))
			sb.WriteString("]")

		case accessKindCall:
			if needsOptional || part.isAlreadyOptional {
				sb.WriteString("?.")
			}
			if len(part.typeArgs) > 0 {
				sb.WriteString("<")
				for j, ta := range part.typeArgs {
					if j > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(ca.getNodeText(ta))
				}
				sb.WriteString(">")
			}
			sb.WriteString("(")
			args := part.callArgs
			for j, arg := range args {
				if j > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(ca.getNodeText(arg))
			}
			sb.WriteString(")")
		}
	}

	return sb.String()
}

type accessKind int

const (
	accessKindProperty accessKind = iota
	accessKindElement
	accessKindCall
)

type chainPart struct {
	node              *ast.Node
	accessKind        accessKind
	accessName        string
	accessArgument    *ast.Node
	callArgs          []*ast.Node
	typeArgs          []*ast.Node
	isAlreadyOptional bool
}

func flattenChainExpression(node *ast.Node) []chainPart {
	n := ast.SkipParentheses(node)
	var parts []chainPart
	flattenChainExpressionRec(n, &parts)
	return parts
}

func flattenChainExpressionRec(node *ast.Node, parts *[]chainPart) {
	n := ast.SkipParentheses(node)

	switch n.Kind {
	case ast.KindNonNullExpression:
		// For foo!, check if the inner expression is a chain access like (foo.bar)!
		inner := ast.SkipParentheses(n.Expression())
		if inner.Kind == ast.KindPropertyAccessExpression ||
			inner.Kind == ast.KindElementAccessExpression ||
			inner.Kind == ast.KindCallExpression {
			// The NonNullExpression wraps a chain access - recurse into it
			flattenChainExpressionRec(inner, parts)
		} else {
			// Base-level NonNullExpression like foo! - preserve it as the base node
			*parts = append(*parts, chainPart{node: n})
		}
		return

	case ast.KindPropertyAccessExpression:
		prop := n.AsPropertyAccessExpression()
		flattenChainExpressionRec(prop.Expression, parts)
		*parts = append(*parts, chainPart{
			node:              n,
			accessKind:        accessKindProperty,
			accessName:        nodeText(prop.Name()),
			isAlreadyOptional: prop.QuestionDotToken != nil,
		})
		return

	case ast.KindElementAccessExpression:
		elem := n.AsElementAccessExpression()
		flattenChainExpressionRec(elem.Expression, parts)
		*parts = append(*parts, chainPart{
			node:              n,
			accessKind:        accessKindElement,
			accessArgument:    elem.ArgumentExpression,
			isAlreadyOptional: elem.QuestionDotToken != nil,
		})
		return

	case ast.KindCallExpression:
		call := n.AsCallExpression()
		flattenChainExpressionRec(call.Expression, parts)
		var callArgs []*ast.Node
		if call.Arguments != nil {
			callArgs = call.Arguments.Nodes
		}
		var typeArgs []*ast.Node
		if call.TypeArguments != nil {
			typeArgs = call.TypeArguments.Nodes
		}
		*parts = append(*parts, chainPart{
			node:              n,
			accessKind:        accessKindCall,
			callArgs:          callArgs,
			typeArgs:          typeArgs,
			isAlreadyOptional: call.QuestionDotToken != nil,
		})
		return
	}

	// Base case: identifier, this, etc.
	*parts = append(*parts, chainPart{
		node: n,
	})
}

func needsParensAsBase(node *ast.Node) bool {
	n := skipDownwards(node)
	return ast.IsAwaitExpression(n)
}

func findBinaryExpressionCovering(root *ast.Node, first *ast.Node, last *ast.Node) *ast.Node {
	// Walk up from first to find a binary expression that covers both first and last
	node := first
	for node != nil {
		if ast.IsBinaryExpression(node) {
			bin := node.AsBinaryExpression()
			op := bin.OperatorToken.Kind
			if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
				if node.Pos() <= first.Pos() && node.End() >= last.End() {
					return node
				}
			}
		}
		if node == root {
			break
		}
		node = node.Parent
	}
	return nil
}

func (ca *ChainAnalyzer) getNodeText(node *ast.Node) string {
	trimmed := utils.TrimNodeTextRange(ca.ctx.SourceFile, node)
	return ca.sourceText[trimmed.Pos():trimmed.End()]
}

func (ca *ChainAnalyzer) CheckNullishAndReport(node *ast.Node) bool {
	if !derefBoolDefault(ca.opts.RequireNullish, false) {
		return false
	}

	if ca.ctx.TypeChecker == nil {
		return false
	}

	// With requireNullish, we need at least one part of the chain to include null/undefined in its type
	t := ca.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}

	for _, part := range utils.UnionTypeParts(t) {
		flags := checker.Type_flags(part)
		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			return false // has nullish -> don't skip
		}
	}

	return true // no nullish found -> skip the report
}

// AnalyzeOrEmptyObjectPattern checks for (foo || {}).bar / (foo ?? {}).bar patterns
func (ca *ChainAnalyzer) AnalyzeOrEmptyObjectPattern(node *ast.Node) {
	// node is a PropertyAccessExpression or ElementAccessExpression
	var exprNode *ast.Node
	var accessName string
	var accessArgument *ast.Node
	var isElementAccess bool

	if ast.IsPropertyAccessExpression(node) {
		prop := node.AsPropertyAccessExpression()
		if prop.QuestionDotToken != nil {
			return // already optional chain
		}
		exprNode = prop.Expression
		accessName = nodeText(prop.Name())
	} else if ast.IsElementAccessExpression(node) {
		elem := node.AsElementAccessExpression()
		if elem.QuestionDotToken != nil {
			return // already optional chain
		}
		exprNode = elem.Expression
		accessArgument = elem.ArgumentExpression
		isElementAccess = true
	} else {
		return
	}

	// The expression should be a parenthesized || or ?? with {} on the right
	inner := ast.SkipParentheses(exprNode)
	if !ast.IsBinaryExpression(inner) {
		return
	}

	bin := inner.AsBinaryExpression()
	if bin.OperatorToken.Kind != ast.KindBarBarToken && bin.OperatorToken.Kind != ast.KindQuestionQuestionToken {
		return
	}

	right := ast.SkipParentheses(bin.Right)
	if !isEmptyObjectLiteral(right) {
		return
	}

	// Check requireNullish option
	if ca.CheckNullishAndReport(bin.Left) {
		return
	}

	// Build the fix
	leftText := ca.getNodeText(bin.Left)

	// Check if left needs parens
	leftNode := ast.SkipParentheses(bin.Left)
	if needsParensAsBase(leftNode) {
		leftText = "(" + leftText + ")"
	}

	var fixCode string
	if isElementAccess {
		argText := ca.getNodeText(accessArgument)
		fixCode = leftText + "?.[" + argText + "]"
	} else {
		fixCode = leftText + "?." + accessName
	}

	reportRange := utils.TrimNodeTextRange(ca.ctx.SourceFile, node).WithEnd(node.End())

	msg := buildPreferOptionalChainMessage()
	sugMsg := buildOptionalChainSuggestMessage()

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(reportRange, fixCode),
	}

	// (foo || {}).bar is always a suggestion, not a fix
	rule.ReportNodeWithFixesOrSuggestions(ca.ctx, node, false, msg, sugMsg, fixes...)
}

func isEmptyObjectLiteral(node *ast.Node) bool {
	if node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	obj := node.AsObjectLiteralExpression()
	return obj.Properties == nil || len(obj.Properties.Nodes) == 0
}

// wrapChainCode wraps the generated optional chain code with the last operand's
// comparison wrapper (e.g., `!= null`, `!`, `typeof ... !== 'undefined'`).
func wrapChainCode(chainCode string, lastOperand Operand) string {
	switch lastOperand.ComparisonType {
	case ComparisonBoolean:
		return chainCode

	case ComparisonNotBoolean:
		return "!" + chainCode

	case ComparisonNotEqualNullOrUndefined:
		if lastOperand.IsYoda {
			if lastOperand.UsesNull {
				return "null != " + chainCode
			}
			return "undefined != " + chainCode
		}
		if lastOperand.UsesNull {
			return chainCode + " != null"
		}
		return chainCode + " != undefined"

	case ComparisonNotStrictEqualNull:
		if lastOperand.IsYoda {
			return "null !== " + chainCode
		}
		return chainCode + " !== null"

	case ComparisonNotStrictEqualUndefined:
		if lastOperand.IsTypeof {
			if lastOperand.IsYoda {
				return "'undefined' !== typeof " + chainCode
			}
			return "typeof " + chainCode + " !== 'undefined'"
		}
		if lastOperand.IsYoda {
			return "undefined !== " + chainCode
		}
		return chainCode + " !== undefined"

	case ComparisonEqualNullOrUndefined:
		if lastOperand.IsYoda {
			if lastOperand.UsesNull {
				return "null == " + chainCode
			}
			return "undefined == " + chainCode
		}
		if lastOperand.UsesNull {
			return chainCode + " == null"
		}
		return chainCode + " == undefined"

	case ComparisonStrictEqualNull:
		if lastOperand.IsYoda {
			return "null === " + chainCode
		}
		return chainCode + " === null"

	case ComparisonStrictEqualUndefined:
		if lastOperand.IsTypeof {
			if lastOperand.IsYoda {
				return "'undefined' === typeof " + chainCode
			}
			return "typeof " + chainCode + " === 'undefined'"
		}
		if lastOperand.IsYoda {
			return "undefined === " + chainCode
		}
		return chainCode + " === undefined"
	}

	return chainCode
}
