package prefer_optional_chain

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ChainAnalyzer struct {
	ctx  rule.RuleContext
	opts PreferOptionalChainOptions
}

func NewChainAnalyzer(ctx rule.RuleContext, opts PreferOptionalChainOptions) *ChainAnalyzer {
	return &ChainAnalyzer{
		ctx:  ctx,
		opts: opts,
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

		// Pre-compute which operands are "paired" strict checks.
		// Paired means: two consecutive operands with Equal ComparedNodes and
		// complementary strict checks (e.g., !== undefined + !== null).
		// Paired operands together cover both nullish values.
		pairedOperands := make(map[int]bool)
		for j := range len(operands) - 1 {
			a, b := operands[j], operands[j+1]
			if a.ComparedNode != nil && b.ComparedNode != nil &&
				compareNodesUncached(a.ComparedNode, b.ComparedNode) == NodeComparisonEqual &&
				isComplementaryGuard(a, b) {
				pairedOperands[j] = true
				pairedOperands[j+1] = true
			}
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

			// Check if the current guard is sufficient for optional chaining.
			// Strict equality (!== null, !== undefined) only covers one nullish value.
			// Skip the check if this operand is part of a complementary pair.
			if !pairedOperands[chainEnd-1] && !ca.isGuardSufficientForChain(curr) {
				break
			}

			cmp := compareNodesUncached(curr.ComparedNode, next.ComparedNode)
			if cmp != NodeComparisonSubset && cmp != NodeComparisonEqual {
				break
			}

			// When extending to a deeper property (Subset), verify the next guard
			// is sufficient (or paired). Prevents extending through a lone strict
			// check (e.g., !== undefined after != null).
			if cmp == NodeComparisonSubset && !pairedOperands[chainEnd] && !ca.isGuardSufficientForChain(next) {
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

		// The last operand must extend beyond the second-to-last operand (Subset, not Equal).
		// If the last operand is Equal to the second-to-last, trim the chain back.
		// E.g., `foo && foo.bar() && foo.bar()` → chain should be `foo && foo.bar()` only.
		for chainEnd-chainStart >= 2 {
			secondToLast := operands[chainEnd-2]
			lastOp := operands[chainEnd-1]
			if secondToLast.ComparedNode != nil && lastOp.ComparedNode != nil {
				if compareNodesUncached(secondToLast.ComparedNode, lastOp.ComparedNode) == NodeComparisonEqual {
					chainEnd--
					continue
				}
			}
			break
		}
		if chainEnd-chainStart < 2 {
			i++
			continue
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

		// typeof check on a bare identifier that's NOT nullable (e.g., `typeof globalThis`)
		// is a "is defined?" check, not a null guard. Skip it from the chain.
		// But typeof on a nullable parameter (e.g., `globalThis?: ...`) IS a valid guard.
		if operands[chainStart].IsTypeof && firstCompared != nil {
			fc := ast.SkipParentheses(firstCompared)
			if fc.Kind == ast.KindIdentifier && !ca.nodeTypeHasFlags(fc, checker.TypeFlagsUndefined|checker.TypeFlagsVoid) {
				chainStart++
				if chainEnd-chainStart < 2 {
					i = chainEnd
					continue
				}
			}
		}

		// When the chain trails with a bare truthy/negation check on a call expression
		// AND any guard uses strict UNDEFINED equality on a call expression result,
		// truncate the chain for fix generation but keep the tail unchanged.
		// Matches TS-ESLint: the rule is conservative for `!== undefined` chains with
		// impure calls because the comparison result and call count both change.
		originalChainEnd := chainEnd
		if chainEnd-chainStart >= 2 {
			last := operands[chainEnd-1]
			if (last.ComparisonType == ComparisonBoolean || last.ComparisonType == ComparisonNotBoolean) &&
				last.ComparedNode != nil &&
				ast.IsCallExpression(ast.SkipParentheses(last.ComparedNode)) {
				for gi := chainStart; gi < chainEnd-1; gi++ {
					g := operands[gi]
					if g.ComparedNode != nil &&
						(g.ComparisonType == ComparisonNotStrictEqualUndefined ||
							g.ComparisonType == ComparisonStrictEqualUndefined) &&
						ast.IsCallExpression(ast.SkipParentheses(g.ComparedNode)) {
						chainEnd = gi + 1
						break
					}
				}
			}
		}
		if chainEnd-chainStart < 2 {
			i = originalChainEnd
			continue
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
			i = originalChainEnd
			continue
		}

		if chainEnd < originalChainEnd {
			ca.reportChainWithTail(operands[chainStart:chainEnd], operands[chainEnd:originalChainEnd], operator, parentNode)
		} else {
			ca.reportChain(operands[chainStart:chainEnd], operator, parentNode)
		}
		i = originalChainEnd
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

// reportChainWithTail reports a chain where the fix only covers the truncated
// portion; the tail operands (which were not safe to fold, e.g., because they
// depend on repeated impure call results) are preserved as-is in the source.
// The fix range only covers operand[0] through the last truncated operand.
func (ca *ChainAnalyzer) reportChainWithTail(
	chainOps []Operand,
	tailOps []Operand,
	operator ast.Kind,
	parentNode *ast.Node,
) {
	if len(chainOps) < 2 || len(tailOps) == 0 {
		return
	}

	if ca.shouldSkipForRequireNullish(chainOps) {
		return
	}
	// Check vacuousness across the FULL original chain (chain + tail), since the
	// truncation was only for fix generation. Later guards in the tail may be the
	// meaningful ones (tsgo narrows earlier guards' types away from nullish).
	fullOps := make([]Operand, 0, len(chainOps)+len(tailOps))
	fullOps = append(fullOps, chainOps...)
	fullOps = append(fullOps, tailOps...)
	if ca.hasOnlyVacuousStrictGuards(fullOps[:len(fullOps)-1]) {
		return
	}

	fixCode := ca.buildOptionalChainCode(chainOps, operator)
	if fixCode == "" {
		return
	}
	fixCode = wrapChainCode(fixCode, chainOps[len(chainOps)-1])

	// Report range covers the full expression (chain + tail) for display purposes,
	// but fix range only covers the truncated chain — tail source text stays as-is.
	firstNode := chainOps[0].Node
	lastChainNode := chainOps[len(chainOps)-1].Node
	lastTailNode := tailOps[len(tailOps)-1].Node
	reportNode := findBinaryExpressionCovering(parentNode, firstNode, lastTailNode)
	if reportNode == nil {
		reportNode = parentNode
	}

	startNode := firstNode
	n := firstNode.Parent
	for n != nil && n != reportNode.Parent {
		if ast.IsParenthesizedExpression(n) {
			startNode = n
		}
		n = n.Parent
	}
	// Fix range: from start to the last chain operand's end (tail preserved as-is).
	fixRange := utils.TrimNodeTextRange(ca.ctx.SourceFile, startNode).WithEnd(lastChainNode.End())

	// For truncated chains, force suggestion for && chains (matches TS-ESLint).
	useFix := ca.shouldUseFix(chainOps)
	if operator == ast.KindAmpersandAmpersandToken {
		useFix = false
	}

	msg := buildPreferOptionalChainMessage()
	sugMsg := buildOptionalChainSuggestMessage()

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(fixRange, fixCode),
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
					if t != nil && utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny) {
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

	case ComparisonNotEqualNullOrUndefined:
		// != null/undefined: undefined != null is false. Matches original short-circuit. Safe.
		return true

	case ComparisonNotStrictEqualUndefined,
		ComparisonStrictEqualUndefined:
		// !== undefined / === undefined: the comparison result is the same
		// whether applied to undefined or original short-circuit value. Safe.
		return true

	case ComparisonNotStrictEqualNull:
		// !== null: undefined !== null is true, but original && chain
		// short-circuits to false when guard fails. Semantics change! Not safe.
		return false

	case ComparisonStrictEqualNull:
		// === null: undefined === null is false, but original || chain
		// short-circuits to true. Semantics change! Not safe.
		return false
	}

	// Fallback: check if the result type already includes undefined
	if ca.ctx.TypeChecker != nil && last.ComparedNode != nil {
		t := ca.ctx.TypeChecker.GetTypeAtLocation(last.ComparedNode)
		if t != nil && utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny) {
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
		if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) {
			return false // has nullish -> don't skip
		}
	}
	return true // no nullish found in any operand -> skip
}

// isGuardSufficientForChain checks whether a strict equality guard fully covers
// the nullish types of the operand. Loose equality (!= / ==) catches both null
// and undefined, so it's always sufficient. Strict equality (!== / ===) only
// catches one, so we need to verify the type doesn't include the other.
// isComplementaryGuard checks if two guards together cover both null and undefined.
// E.g., (!== undefined, !== null) or (!== null, !== undefined) are complementary.
func isComplementaryGuard(a, b Operand) bool {
	coversNull := func(ct NullishComparisonType) bool {
		return ct == ComparisonNotStrictEqualNull || ct == ComparisonStrictEqualNull
	}
	coversUndefined := func(ct NullishComparisonType) bool {
		return ct == ComparisonNotStrictEqualUndefined || ct == ComparisonStrictEqualUndefined
	}
	return (coversNull(a.ComparisonType) && coversUndefined(b.ComparisonType)) ||
		(coversUndefined(a.ComparisonType) && coversNull(b.ComparisonType))
}

func (ca *ChainAnalyzer) isGuardSufficientForChain(guard Operand) bool {
	switch guard.ComparisonType {
	case ComparisonBoolean, ComparisonNotBoolean:
		return true
	case ComparisonNotEqualNullOrUndefined, ComparisonEqualNullOrUndefined:
		return true
	case ComparisonNotStrictEqualNull, ComparisonStrictEqualNull:
		if ca.ctx.TypeChecker == nil {
			return false
		}
		return !ca.nodeTypeHasFlags(guard.ComparedNode, checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny|checker.TypeFlagsUnknown)
	case ComparisonNotStrictEqualUndefined, ComparisonStrictEqualUndefined:
		if ca.ctx.TypeChecker == nil {
			return false
		}
		return !ca.nodeTypeHasFlags(guard.ComparedNode, checker.TypeFlagsNull|checker.TypeFlagsAny|checker.TypeFlagsUnknown)
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
			if ca.nodeTypeHasFlags(guard.ComparedNode, checker.TypeFlagsNull|checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return false
			}
		case ComparisonNotStrictEqualUndefined, ComparisonStrictEqualUndefined:
			if ca.nodeTypeHasFlags(guard.ComparedNode, checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
				return false
			}
		}
	}
	return true // all guards are vacuous
}

// nodeTypeHasFlags checks if a node's type includes any of the given flags,
// iterating through union constituents. Returns false if type checker or node is nil.
func (ca *ChainAnalyzer) nodeTypeHasFlags(node *ast.Node, flags checker.TypeFlags) bool {
	if ca.ctx.TypeChecker == nil || node == nil {
		return false
	}
	t := ca.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	return utils.IsTypeFlagSetWithUnion(t, flags)
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

	// Track the deepest guard match so we can use its parts for output
	var deepestGuardParts []chainPart
	deepestGuardPartIdx := -1

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
				// Track the deepest guard for output generation
				if j > deepestGuardPartIdx {
					deepestGuardPartIdx = j
					deepestGuardParts = flattenChainExpression(guard.ComparedNode)
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
					deepestGuardPartIdx = j
					deepestGuardParts = flattenChainExpression(firstGuard.ComparedNode)
					break
				}
			}
		}
	}

	// Build the output with ?. inserted at appropriate positions
	var sb strings.Builder

	// Merge ?. and ! annotations from guards into the output.
	// tsgo's AST QuestionDotToken flags may be on the wrong operand's nodes,
	// so we scan the guard's source text (via GetSourceTextOfNodeFromSourceFile)
	// to determine the correct ?. positions.
	guardOptionals := make(map[int]bool)
	guardNonNulls := make(map[int]bool)
	if deepestGuardPartIdx >= 0 {
		// Scan each guard's source text for ?. positions
		for gi := range len(operands) - 1 {
			guard := operands[gi]
			if guard.ComparedNode == nil {
				continue
			}
			guardText := ca.getNodeText(guard.ComparedNode)
			depth := 0
			for ci := 0; ci < len(guardText); ci++ {
				if ci+1 < len(guardText) && guardText[ci] == '?' && guardText[ci+1] == '.' {
					depth++
					guardOptionals[depth] = true
					ci++ // skip the '.'
				} else if guardText[ci] == '.' {
					depth++
				}
			}
		}
		guardNonNulls = collectNonNullPositions(operands[0 : len(operands)-1])
		// Align base node NonNull with the guard:
		// - If guard has NonNull base and target doesn't: use guard's base
		// - If target has NonNull base but guard doesn't: strip it (use inner expression)
		if len(deepestGuardParts) > 0 {
			gp := deepestGuardParts[0]
			if ast.IsNonNullExpression(gp.node) && !ast.IsNonNullExpression(parts[0].node) {
				parts[0].node = gp.node
			} else if !ast.IsNonNullExpression(gp.node) && ast.IsNonNullExpression(parts[0].node) {
				// Guard has no NonNull but target does — strip it
				parts[0].node = ast.SkipParentheses(parts[0].node.Expression())
			}
		}
	}

	// Write the base expression
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
		// Use ?. from optionalIndices (new insertions) OR from guard source text
		// (preserving existing ?. from the original code)
		useOptional := optionalIndices[i] || (i <= deepestGuardPartIdx && guardOptionals[i])

		// Check NonNull from guard (skip position 0 if base is already NonNull)
		hasNonNull := false
		if i-1 < deepestGuardPartIdx && (i-1 != 0 || !ast.IsNonNullExpression(parts[0].node)) {
			hasNonNull = guardNonNulls[i-1]
		}

		ca.writeChainPart(&sb, parts[i], useOptional, hasNonNull)
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
	hasNonNullAfter   bool // the chain result up to this part is wrapped in NonNullExpression (!)
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
			prevLen := len(*parts)
			flattenChainExpressionRec(inner, parts)
			// Mark the last added part as having NonNull after it (e.g., foo.bar! → .bar has NonNull)
			if len(*parts) > prevLen {
				(*parts)[len(*parts)-1].hasNonNullAfter = true
			}
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
			accessName:        prop.Name().Text(),
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

// writeChainPart writes a single accessor part to the string builder,
// handling ?., !., and regular . delimiters.
func (ca *ChainAnalyzer) writeChainPart(sb *strings.Builder, part chainPart, needsOptional bool, prevHasNonNull bool) {
	switch part.accessKind {
	case accessKindProperty:
		if needsOptional {
			sb.WriteString("?.")
		} else if prevHasNonNull {
			sb.WriteString("!.")
		} else {
			sb.WriteString(".")
		}
		sb.WriteString(part.accessName)

	case accessKindElement:
		if needsOptional {
			sb.WriteString("?.")
		} else if prevHasNonNull {
			sb.WriteString("!")
		}
		sb.WriteString("[")
		sb.WriteString(ca.getNodeText(part.accessArgument))
		sb.WriteString("]")

	case accessKindCall:
		if needsOptional {
			sb.WriteString("?.")
		} else if prevHasNonNull {
			sb.WriteString("!")
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
		for j, arg := range part.callArgs {
			if j > 0 {
				sb.WriteString(",")
			}
			// Use trivia-preserving text to keep comments (e.g., /* comment */a)
			sb.WriteString(ca.getNodeTextWithTrivia(arg))
		}
		sb.WriteString(")")
	}
}

// collectNonNullPositions walks the guard operands' ASTs to determine at which
// chain depths a NonNullExpression (!) wraps the chain result.
func collectNonNullPositions(guards []Operand) map[int]bool {
	result := make(map[int]bool)
	for _, guard := range guards {
		if guard.ComparedNode == nil {
			continue
		}
		collectNonNullRec(ast.SkipParentheses(guard.ComparedNode), result)
	}
	return result
}

func collectNonNullRec(node *ast.Node, result map[int]bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		prop := node.AsPropertyAccessExpression()
		collectNonNullRec(ast.SkipParentheses(prop.Expression), result)
	case ast.KindElementAccessExpression:
		elem := node.AsElementAccessExpression()
		collectNonNullRec(ast.SkipParentheses(elem.Expression), result)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		collectNonNullRec(ast.SkipParentheses(call.Expression), result)
	case ast.KindNonNullExpression:
		inner := ast.SkipParentheses(node.Expression())
		depth := chainDepth(inner)
		result[depth] = true
		collectNonNullRec(inner, result)
	}
}

// chainDepth returns the number of access steps in a chain expression.
// foo = 0, foo.bar = 1, foo.bar.baz = 2, etc.
func chainDepth(node *ast.Node) int {
	n := ast.SkipParentheses(node)
	depth := 0
	for {
		switch n.Kind {
		case ast.KindPropertyAccessExpression:
			depth++
			n = ast.SkipParentheses(n.AsPropertyAccessExpression().Expression)
			continue
		case ast.KindElementAccessExpression:
			depth++
			n = ast.SkipParentheses(n.AsElementAccessExpression().Expression)
			continue
		case ast.KindCallExpression:
			depth++
			n = ast.SkipParentheses(n.AsCallExpression().Expression)
			continue
		case ast.KindNonNullExpression:
			n = ast.SkipParentheses(n.Expression())
			continue
		}
		break
	}
	return depth
}

func needsParensAsBase(node *ast.Node) bool {
	n := skipDownwards(node)
	return ast.IsAwaitExpression(n)
}

// needsParensForOptionalBase checks if an expression needs parentheses when
// used as the LHS of ?. in (expr || {}).bar → expr?.bar transforms.
// E.g., `foo || undefined` needs parens: `(foo || undefined)?.bar`, not `foo || undefined?.bar`.
func needsParensForOptionalBase(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBinaryExpression,
		ast.KindConditionalExpression,
		ast.KindAwaitExpression,
		ast.KindVoidExpression,
		ast.KindTypeOfExpression,
		ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindAsExpression,
		ast.KindSatisfiesExpression:
		return true
	}
	return false
}

func findBinaryExpressionCovering(root *ast.Node, first *ast.Node, last *ast.Node) *ast.Node {
	firstPos, lastEnd := first.Pos(), last.End()
	for node := first; node != nil; node = node.Parent {
		if ast.IsBinaryExpression(node) {
			op := node.AsBinaryExpression().OperatorToken.Kind
			if (op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken) &&
				node.Pos() <= firstPos && node.End() >= lastEnd {
				return node
			}
		}
		if node == root {
			break
		}
	}
	return nil
}

func (ca *ChainAnalyzer) getNodeText(node *ast.Node) string {
	return scanner.GetSourceTextOfNodeFromSourceFile(ca.ctx.SourceFile, node, false)
}

// getNodeTextWithTrivia returns the source text for a node including leading
// trivia (comments, whitespace). Used for call arguments where comments like
// /* comment */ should be preserved in the output.
func (ca *ChainAnalyzer) getNodeTextWithTrivia(node *ast.Node) string {
	return scanner.GetSourceTextOfNodeFromSourceFile(ca.ctx.SourceFile, node, true)
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

	if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) {
		return false // has nullish -> don't skip
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
		accessName = prop.Name().Text()
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
	if !ast.IsEmptyObjectLiteral(right) {
		return
	}

	// Check requireNullish option
	if ca.CheckNullishAndReport(bin.Left) {
		return
	}

	// Build the fix
	leftText := ca.getNodeText(bin.Left)

	// Check if left needs parens when used as the LHS of ?.
	// Skip if the source already has parens around the left expression.
	leftNode := ast.SkipParentheses(bin.Left)
	if needsParensForOptionalBase(leftNode) && !ast.IsParenthesizedExpression(bin.Left) {
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
