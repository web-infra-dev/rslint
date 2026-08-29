package prefer_optional_chain

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ChainAnalyzer struct {
	ctx        rule.RuleContext
	opts       PreferOptionalChainOptions
	chainParts []chainPart
	partFlags  []uint8
}

func NewChainAnalyzer(ctx rule.RuleContext, opts PreferOptionalChainOptions) *ChainAnalyzer {
	return &ChainAnalyzer{
		ctx:  ctx,
		opts: opts,
	}
}

// reportRangeWithDeferredFixesOrSuggestions reports a diagnostic immediately,
// but only decides its edit category and builds replacement text when the
// diagnostic consumer requests edits.
func reportRangeWithDeferredFixesOrSuggestions(
	ctx rule.RuleContext,
	textRange core.TextRange,
	msg rule.RuleMessage,
	suggestionMsg rule.RuleMessage,
	shouldUseFix func() bool,
	buildFixes func() []rule.RuleFix,
) {
	var (
		decisionReady bool
		useFix        bool
		fixesReady    bool
		fixes         []rule.RuleFix
	)

	getUseFix := func() bool {
		if !decisionReady {
			useFix = shouldUseFix()
			decisionReady = true
		}
		return useFix
	}
	getFixes := func() []rule.RuleFix {
		if !fixesReady {
			fixes = buildFixes()
			fixesReady = true
		}
		return fixes
	}

	ctx.ReportRangeWithDeferredFixesAndSuggestions(
		textRange,
		msg,
		func() []rule.RuleFix {
			if !getUseFix() {
				return nil
			}
			return getFixes()
		},
		func() []rule.RuleSuggestion {
			if getUseFix() {
				return nil
			}
			builtFixes := getFixes()
			if len(builtFixes) == 0 {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message:  suggestionMsg,
				FixesArr: builtFixes,
			}}
		},
	)
}

func trimNodeRangeWithEnd(sourceFile *ast.SourceFile, node *ast.Node, end int) core.TextRange {
	return core.NewTextRange(scanner.SkipTrivia(sourceFile.Text(), node.Pos()), end)
}

func (ca *ChainAnalyzer) AnalyzeChain(
	operands []Operand,
	operator ast.Kind,
	parentNode *ast.Node,
) {
	// An invalid operand breaks the chain that precedes it; a last-chain
	// operand closes it. Everything between those boundaries is a run of
	// guards that may collapse into one optional chain.
	runStart := 0
	for i := 0; i <= len(operands); i++ {
		if i == len(operands) {
			ca.analyzeRun(operands[runStart:i], nil, operator, parentNode)
			break
		}
		switch operands[i].Validity {
		case OperandInvalid:
			ca.analyzeRun(operands[runStart:i], nil, operator, parentNode)
			runStart = i + 1
		case OperandLast:
			ca.analyzeRun(operands[runStart:i], &operands[i], operator, parentNode)
			runStart = i + 1
		}
	}
}

// analyzeRun walks a run of guards, growing the longest chain of subset
// accesses it can and reporting each chain it completes. lastChainOperand,
// when present, closes the run's final chain with a comparison against an
// arbitrary value.
func (ca *ChainAnalyzer) analyzeRun(
	chain []Operand,
	lastChainOperand *Operand,
	operator ast.Kind,
	parentNode *ast.Node,
) {
	total := len(chain)
	if lastChainOperand != nil {
		total++
	}
	if total <= 1 {
		return
	}

	// A pair of strict checks on the same access, such as
	// `x !== null && x !== undefined`, is one link of the chain even though it
	// contributes two operands, so links are counted separately from operands.
	subChain := make([]Operand, 0, total)
	links := 0
	var lastChain *Operand

	reportThenReset := func(seed []Operand) {
		if links+boolToInt(lastChain != nil) > 1 {
			ca.reportChain(subChain, lastChain, operator, parentNode)
		}
		// The operand that ended this chain may start the next one, as in
		// `unrelated != null && foo != null && foo.bar`, so keep it.
		subChain = append(subChain[:0], seed...)
		links = boolToInt(len(seed) > 0)
		lastChain = nil
	}

	for i := 0; i < len(chain); i++ {
		var prev *Operand
		if len(subChain) > 0 {
			prev = &subChain[len(subChain)-1]
		}
		operand := chain[i]

		consumed := ca.analyzeGuard(chain, i, operator)
		if consumed == 0 {
			// A check against `undefined` that reaches deeper than the chain so
			// far can still close it, leaving the rest of the expression alone.
			if prev != nil &&
				(operand.ComparisonType == ComparisonStrictEqualUndefined ||
					operand.ComparisonType == ComparisonNotStrictEqualUndefined) &&
				compareNodesUncached(prev.ComparedNode, operand.ComparedNode) == NodeComparisonSubset {
				lastChain = &chain[i]
			}
			reportThenReset(nil)
			continue
		}

		group := chain[i : i+consumed]
		i += consumed - 1

		if prev == nil {
			subChain = append(subChain, group[0])
			links++
			continue
		}
		// Compare against the last operand of the group: the earlier ones check
		// the same access, so `foo !== null && foo !== undefined` stays one unit.
		switch compareNodesUncached(prev.ComparedNode, group[len(group)-1].ComparedNode) {
		case NodeComparisonSubset:
			subChain = append(subChain, group[0])
			links++
		case NodeComparisonInvalid:
			reportThenReset(group)
		}
		// An equal access is a no-op, so `foo && foo` is left alone.
	}

	if len(subChain) > 0 && lastChainOperand != nil {
		if resolved, ok := ca.resolveLastChainOperand(subChain, *lastChainOperand, operator); ok {
			lastChain = &resolved
		}
	}
	reportThenReset(nil)
}

// analyzeGuard reports how many operands the guard at index consumes, or 0 when
// it cannot narrow the chain. A pair of strict checks that together cover both
// null and undefined consumes two.
func (ca *ChainAnalyzer) analyzeGuard(chain []Operand, index int, operator ast.Kind) int {
	operand := chain[index]
	hasMore := index+1 < len(chain)
	pairsWith := func(want NullishComparisonType) bool {
		return hasMore &&
			chain[index+1].ComparisonType == want &&
			compareNodesUncached(operand.ComparedNode, chain[index+1].ComparedNode) == NodeComparisonEqual
	}

	if operator == ast.KindAmpersandAmpersandToken {
		switch operand.ComparisonType {
		case ComparisonBoolean, ComparisonNotEqualNullOrUndefined:
			return 1
		case ComparisonNotStrictEqualNull:
			if pairsWith(ComparisonNotStrictEqualUndefined) {
				return 2
			}
			// `x !== null` leaves undefined in play, so it only guards a chain
			// when x cannot be undefined and a later operand carries the chain.
			if hasMore && !ca.includesType(operand.ComparedNode, checker.TypeFlagsUndefined) {
				return 1
			}
		case ComparisonNotStrictEqualUndefined:
			if pairsWith(ComparisonNotStrictEqualNull) {
				return 2
			}
			if !ca.includesType(operand.ComparedNode, checker.TypeFlagsNull) {
				return 1
			}
		}
		return 0
	}

	switch operand.ComparisonType {
	case ComparisonNotBoolean, ComparisonEqualNullOrUndefined:
		return 1
	case ComparisonStrictEqualNull:
		if pairsWith(ComparisonStrictEqualUndefined) {
			return 2
		}
		if !ca.includesType(operand.ComparedNode, checker.TypeFlagsUndefined) {
			return 1
		}
	case ComparisonStrictEqualUndefined:
		if pairsWith(ComparisonStrictEqualNull) {
			return 2
		}
		if !ca.includesType(operand.ComparedNode, checker.TypeFlagsNull) {
			return 1
		}
	}
	return 0
}

// includesType reports whether the node's type has any of the given flags,
// counting `any` and `unknown` as possibly including them.
func (ca *ChainAnalyzer) includesType(node *ast.Node, flags checker.TypeFlags) bool {
	return ca.nodeTypeHasFlags(node, flags|checker.TypeFlagsAny|checker.TypeFlagsUnknown)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// resolveLastChainOperand decides whether `last` closes the chain built from
// `chain`, and returns it with the chain-target side resolved.
func (ca *ChainAnalyzer) resolveLastChainOperand(
	chain []Operand,
	last Operand,
	operator ast.Kind,
) (Operand, bool) {
	prev := chain[len(chain)-1]
	if prev.ComparedNode == nil || last.ComparedNode == nil || last.ComparisonValue == nil {
		return Operand{}, false
	}

	comparedNode, comparisonValue, isYoda, ok := resolveOperandSubset(prev, last)
	if !ok {
		return Operand{}, false
	}
	if !ca.isValidLastChainComparison(comparisonValue, last.LastComparison, operator) {
		return Operand{}, false
	}

	last.ComparedNode = comparedNode
	last.ComparisonValue = comparisonValue
	last.IsYoda = isYoda
	return last, true
}

// resolveOperandSubset picks which side of the trailing comparison extends the
// chain. When both sides are member-based the ambiguity is resolved by which
// one is a subset of the preceding guard — if both are, or neither is, the
// chain cannot be folded.
func resolveOperandSubset(prev, last Operand) (comparedNode, comparisonValue *ast.Node, isYoda, ok bool) {
	nameIsSubset := compareNodesUncached(prev.ComparedNode, last.ComparedNode) == NodeComparisonSubset
	if last.Yoda != YodaUnknown {
		return last.ComparedNode, last.ComparisonValue, last.Yoda == YodaYes, nameIsSubset
	}

	valueIsSubset := compareNodesUncached(prev.ComparedNode, last.ComparisonValue) == NodeComparisonSubset
	switch {
	case nameIsSubset && !valueIsSubset:
		return last.ComparedNode, last.ComparisonValue, false, true
	case !nameIsSubset && valueIsSubset:
		return last.ComparisonValue, last.ComparedNode, true, true
	}
	return nil, nil, false, false
}

// isValidLastChainComparison reports whether folding the chain leaves the
// trailing comparison's result unchanged. An optional chain yields `undefined`
// where the original expression short-circuited, so the compared value decides:
// the comparison must keep the same answer for `undefined` as the short circuit
// gave.
func (ca *ChainAnalyzer) isValidLastChainComparison(
	value *ast.Node,
	comparison LastComparisonType,
	operator ast.Kind,
) bool {
	if ca.ctx.TypeChecker == nil {
		return false
	}
	t := ca.ctx.TypeChecker.GetTypeAtLocation(value)
	if t == nil {
		return false
	}
	parts := utils.UnionTypeParts(t)

	someHasFlags := func(flags checker.TypeFlags) bool {
		for _, part := range parts {
			if checker.Type_flags(part)&flags != 0 {
				return true
			}
		}
		return false
	}
	everyHasFlags := func(flags checker.TypeFlags) bool {
		for _, part := range parts {
			if checker.Type_flags(part)&flags == 0 {
				return false
			}
		}
		return true
	}

	const anyOrUnknown = checker.TypeFlagsAny | checker.TypeFlagsUnknown
	loose := comparison == LastComparisonEqual || comparison == LastComparisonNotEqual
	negated := comparison == LastComparisonNotEqual || comparison == LastComparisonNotStrictEqual

	// The comparison whose sense matches the chain operator must never match a
	// nullish value; the opposite one must always match it.
	if negated == (operator == ast.KindBarBarToken) {
		if loose {
			return !someHasFlags(anyOrUnknown | checker.TypeFlagsNull | checker.TypeFlagsUndefined)
		}
		return !someHasFlags(anyOrUnknown | checker.TypeFlagsUndefined)
	}
	if loose {
		return everyHasFlags(checker.TypeFlagsNull | checker.TypeFlagsUndefined)
	}
	return everyHasFlags(checker.TypeFlagsUndefined)
}

// reportChain reports one folded chain: the guards in subChain plus, when
// present, the operand that closes it.
func (ca *ChainAnalyzer) reportChain(
	subChain []Operand,
	lastChain *Operand,
	operator ast.Kind,
	parentNode *ast.Node,
) {
	// The fixes are built lazily, after subChain's backing array has been
	// reused for the next chain, so keep a copy.
	chain := make([]Operand, 0, len(subChain)+1)
	chain = append(chain, subChain...)
	if lastChain != nil {
		chain = append(chain, *lastChain)
	}
	if len(chain) < 2 {
		return
	}

	if ca.shouldSkipForRequireNullish(chain) {
		return
	}

	reportRange := ca.chainReportRange(chain[0].Node, chain[len(chain)-1].Node, parentNode)
	hasLastChain := lastChain != nil

	reportRangeWithDeferredFixesOrSuggestions(
		ca.ctx,
		reportRange,
		buildPreferOptionalChainMessage(),
		buildOptionalChainSuggestMessage(),
		func() bool {
			return ca.shouldUseFix(chain, hasLastChain, operator)
		},
		func() []rule.RuleFix {
			fixCode := ca.buildOptionalChainCode(chain, operator)
			if fixCode == "" {
				return nil
			}
			fixCode = ca.wrapChainCode(fixCode, chain[len(chain)-1])
			fixes := []rule.RuleFix{rule.RuleFixReplaceRange(reportRange, fixCode)}
			for _, pos := range ca.orphanedCloseParens(parentNode, reportRange) {
				fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(pos, pos+1)))
			}
			return fixes
		},
	)
}

// chainReportRange spans the chain's first through last operand, widened over
// any parentheses that sit between them and the enclosing logical expression.
func (ca *ChainAnalyzer) chainReportRange(first, last, boundary *ast.Node) core.TextRange {
	text := ca.ctx.SourceFile.Text()
	boundaryStart := scanner.SkipTrivia(text, boundary.Pos())
	boundaryEnd := boundary.End()

	start := scanner.SkipTrivia(text, first.Pos())
	for n := first.Parent; n != nil; n = n.Parent {
		nodeStart := scanner.SkipTrivia(text, n.Pos())
		if ast.IsParenthesizedExpression(n) && nodeStart >= boundaryStart &&
			scanner.SkipTrivia(text, n.AsParenthesizedExpression().Expression.Pos()) == start {
			start = nodeStart
			continue
		}
		if nodeStart != start {
			break
		}
	}

	end := last.End()
	for n := last.Parent; n != nil; n = n.Parent {
		if ast.IsParenthesizedExpression(n) && n.End() <= boundaryEnd &&
			n.AsParenthesizedExpression().Expression.End() == end {
			end = n.End()
			continue
		}
		if n.End() != end {
			break
		}
	}

	return core.NewTextRange(start, end)
}

// orphanedCloseParens returns the positions of `)` tokens whose `(` the fix
// replaces, as in `foo && (foo.bar && baz)` folding to `foo?.bar && baz`.
func (ca *ChainAnalyzer) orphanedCloseParens(boundary *ast.Node, reportRange core.TextRange) []int {
	text := ca.ctx.SourceFile.Text()
	var positions []int
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if ast.IsParenthesizedExpression(n) && n.End() > reportRange.End() {
			open := scanner.SkipTrivia(text, n.Pos())
			if open >= reportRange.Pos() && open < reportRange.End() {
				positions = append(positions, n.End()-1)
			}
		}
		n.ForEachChild(visit)
		return false
	}
	boundary.ForEachChild(visit)
	return positions
}

// shouldUseFix reports whether the rewrite is safe to apply automatically. An
// optional chain evaluates to `undefined` where the original short-circuited,
// so it is only an auto-fix when that cannot change the expression's value.
func (ca *ChainAnalyzer) shouldUseFix(chain []Operand, hasLastChain bool, operator ast.Kind) bool {
	if ca.opts.AllowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing {
		return true
	}
	// A trailing comparison against an arbitrary value can start answering
	// differently once the chain feeds it `undefined`.
	if hasLastChain {
		return false
	}

	switch chain[len(chain)-1].ComparisonType {
	case ComparisonEqualNullOrUndefined, ComparisonNotEqualNullOrUndefined,
		ComparisonStrictEqualUndefined, ComparisonNotStrictEqualUndefined:
		// The comparison answers the same for `undefined` as for whatever the
		// original expression short-circuited to.
		return true
	case ComparisonNotBoolean:
		if operator == ast.KindBarBarToken {
			return true
		}
	}

	// Otherwise the chain widens the result type with `undefined`, which is
	// only safe when some operand could already produce it.
	for _, op := range chain {
		if ca.includesType(op.Node, checker.TypeFlagsUndefined) {
			return true
		}
	}
	return false
}

func (ca *ChainAnalyzer) shouldSkipForRequireNullish(operands []Operand) bool {
	if !ca.opts.RequireNullish {
		return false
	}
	if ca.ctx.TypeChecker == nil {
		return false
	}

	// With requireNullish, at least one guard operand must have null/undefined in its type.
	// Only check guard operands (all except the last), since the last operand is the
	// chain target, not a guard. E.g., in `foo && foo.bar`, foo is the guard.
	//
	// The type comes from the operand as written, not from the expression it
	// compares: a guard spelled `foo != null` is a boolean, so only a bare
	// truthy check like `foo` can carry a nullish type.
	guards := operands
	if len(guards) > 1 {
		guards = operands[:len(operands)-1]
	}
	for _, op := range guards {
		if op.Node == nil {
			continue
		}
		t := ca.ctx.TypeChecker.GetTypeAtLocation(op.Node)
		if t == nil {
			continue
		}
		if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			return false // has nullish -> don't skip
		}
	}
	return true // no nullish found in any operand -> skip
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
	parts := ca.flattenChainExpression(lastOperand.ComparedNode)
	if len(parts) == 0 {
		return ca.getNodeText(lastOperand.ComparedNode)
	}

	// A guard matching at part[j] means the access at part[j+1] should use ?.
	// Keep all per-part output state in one compact slice.
	var partFlags []uint8
	if cap(ca.partFlags) < len(parts) {
		ca.partFlags = make([]uint8, len(parts))
		partFlags = ca.partFlags
	} else {
		ca.partFlags = ca.partFlags[:len(parts)]
		clear(ca.partFlags)
		partFlags = ca.partFlags
	}
	hasOptionalIndex := false

	// Track the deepest guard match so we can use its base for output.
	var deepestGuardNode *ast.Node
	deepestGuardPartIdx := -1

	for i := range len(operands) - 1 {
		guard := operands[i]
		if guard.ComparedNode == nil {
			continue
		}

		// Equal chain expressions have the same access depth, so only the
		// corresponding part can match. Avoid comparing the guard against
		// every shallower prefix.
		j := chainDepth(guard.ComparedNode)
		if j < len(parts) &&
			compareNodesUncached(guard.ComparedNode, parts[j].node) == NodeComparisonEqual {
			if j+1 < len(parts) {
				partFlags[j+1] |= chainPartFlagInsertedOptional
				hasOptionalIndex = true
			}
			// Track the deepest guard for output generation
			if j > deepestGuardPartIdx {
				deepestGuardPartIdx = j
				deepestGuardNode = guard.ComparedNode
			}
		}
	}

	// If no optional indices were found, try a simpler approach:
	// the first guard is a prefix of the last operand
	if !hasOptionalIndex {
		firstGuard := operands[0]
		if firstGuard.ComparedNode != nil {
			j := chainDepth(firstGuard.ComparedNode)
			if j+1 < len(parts) &&
				compareNodesUncached(firstGuard.ComparedNode, parts[j].node) == NodeComparisonEqual {
				partFlags[j+1] |= chainPartFlagInsertedOptional
				deepestGuardPartIdx = j
				deepestGuardNode = firstGuard.ComparedNode
			}
		}
	}

	// Build the output with ?. inserted at appropriate positions
	var sb strings.Builder
	sb.Grow(len(ca.getNodeText(lastOperand.ComparedNode)) + len(parts))

	// Merge ?. and ! annotations from guards into the output.
	// tsgo's AST QuestionDotToken flags may be on the wrong operand's nodes,
	// so we scan the guard's source text (via GetSourceTextOfNodeFromSourceFile)
	// to determine the correct ?. positions.
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
					if depth < len(partFlags) {
						partFlags[depth] |= chainPartFlagPreservedOptional
					}
					ci++ // skip the '.'
				} else if guardText[ci] == '.' {
					depth++
				}
			}
		}
		collectNonNullPositions(operands[0:len(operands)-1], partFlags)
		// Align base node NonNull with the guard:
		// - If guard has NonNull base and target doesn't: use guard's base
		// - If target has NonNull base but guard doesn't: strip it (use inner expression)
		if deepestGuardNode != nil {
			guardBase := chainBase(deepestGuardNode)
			if ast.IsNonNullExpression(guardBase) && !ast.IsNonNullExpression(parts[0].node) {
				parts[0].node = guardBase
			} else if !ast.IsNonNullExpression(guardBase) && ast.IsNonNullExpression(parts[0].node) {
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
		// Keep optional accesses carried by guard operands through the guarded
		// prefix. Beyond that prefix, the target operand owns the syntax. This
		// deliberately ignores target-only ?. tokens inside the guarded prefix,
		// matching upstream's "randomly placed optional chain tokens" cases.
		guardOptional := i <= deepestGuardPartIdx && partFlags[i]&chainPartFlagPreservedOptional != 0
		targetOptional := i > deepestGuardPartIdx && parts[i].isAlreadyOptional
		useOptional := partFlags[i]&chainPartFlagInsertedOptional != 0 || guardOptional || targetOptional

		// Check NonNull from guard (skip position 0 if base is already NonNull)
		hasNonNull := false
		if i-1 < deepestGuardPartIdx && (i-1 != 0 || !ast.IsNonNullExpression(parts[0].node)) {
			hasNonNull = partFlags[i-1]&chainPartFlagPreservedNonNull != 0
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

const (
	chainPartFlagInsertedOptional uint8 = 1 << iota
	chainPartFlagPreservedOptional
	chainPartFlagPreservedNonNull
)

type chainPart struct {
	node              *ast.Node
	accessKind        accessKind
	accessName        string
	accessArgument    *ast.Node
	callArgsText      string
	typeArgs          []*ast.Node
	isAlreadyOptional bool
	hasNonNullAfter   bool // the chain result up to this part is wrapped in NonNullExpression (!)
}

func (ca *ChainAnalyzer) flattenChainExpression(node *ast.Node) []chainPart {
	n := ast.SkipParentheses(node)
	partCount := chainDepth(n) + 1
	if cap(ca.chainParts) < partCount {
		ca.chainParts = make([]chainPart, 0, partCount)
	} else {
		ca.chainParts = ca.chainParts[:0]
	}
	ca.flattenChainExpressionRec(n, &ca.chainParts)
	return ca.chainParts
}

func chainBase(node *ast.Node) *ast.Node {
	n := ast.SkipParentheses(node)
	for {
		switch n.Kind {
		case ast.KindNonNullExpression:
			inner := ast.SkipParentheses(n.Expression())
			if inner.Kind != ast.KindPropertyAccessExpression &&
				inner.Kind != ast.KindElementAccessExpression &&
				inner.Kind != ast.KindCallExpression {
				return n
			}
			n = inner
		case ast.KindPropertyAccessExpression:
			n = ast.SkipParentheses(n.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			n = ast.SkipParentheses(n.AsElementAccessExpression().Expression)
		case ast.KindCallExpression:
			n = ast.SkipParentheses(n.AsCallExpression().Expression)
		default:
			return n
		}
	}
}

func (ca *ChainAnalyzer) flattenChainExpressionRec(node *ast.Node, parts *[]chainPart) {
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
			ca.flattenChainExpressionRec(inner, parts)
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
		ca.flattenChainExpressionRec(prop.Expression, parts)
		*parts = append(*parts, chainPart{
			node:              n,
			accessKind:        accessKindProperty,
			accessName:        prop.Name().Text(),
			isAlreadyOptional: prop.QuestionDotToken != nil,
		})
		return

	case ast.KindElementAccessExpression:
		elem := n.AsElementAccessExpression()
		ca.flattenChainExpressionRec(elem.Expression, parts)
		*parts = append(*parts, chainPart{
			node:              n,
			accessKind:        accessKindElement,
			accessArgument:    elem.ArgumentExpression,
			isAlreadyOptional: elem.QuestionDotToken != nil,
		})
		return

	case ast.KindCallExpression:
		call := n.AsCallExpression()
		ca.flattenChainExpressionRec(call.Expression, parts)
		var typeArgs []*ast.Node
		if call.TypeArguments != nil {
			typeArgs = call.TypeArguments.Nodes
		}
		*parts = append(*parts, chainPart{
			node:              n,
			accessKind:        accessKindCall,
			callArgsText:      ca.callArgumentsText(call),
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
		sb.WriteString(part.callArgsText)
	}
}

// collectNonNullPositions walks the guard operands' ASTs to determine at which
// chain depths a NonNullExpression (!) wraps the chain result.
func collectNonNullPositions(guards []Operand, partFlags []uint8) {
	for _, guard := range guards {
		if guard.ComparedNode == nil {
			continue
		}
		collectNonNullRec(ast.SkipParentheses(guard.ComparedNode), partFlags)
	}
}

func collectNonNullRec(node *ast.Node, partFlags []uint8) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		prop := node.AsPropertyAccessExpression()
		collectNonNullRec(ast.SkipParentheses(prop.Expression), partFlags)
	case ast.KindElementAccessExpression:
		elem := node.AsElementAccessExpression()
		collectNonNullRec(ast.SkipParentheses(elem.Expression), partFlags)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		collectNonNullRec(ast.SkipParentheses(call.Expression), partFlags)
	case ast.KindNonNullExpression:
		inner := ast.SkipParentheses(node.Expression())
		depth := chainDepth(inner)
		if depth < len(partFlags) {
			partFlags[depth] |= chainPartFlagPreservedNonNull
		}
		collectNonNullRec(inner, partFlags)
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
	return ast.GetExpressionPrecedence(skipDownwards(node)) < ast.OperatorPrecedenceMember
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

// callArgumentsText returns the argument list verbatim, from `(` through `)`,
// so that comments and trailing commas survive the rewrite.
func (ca *ChainAnalyzer) callArgumentsText(call *ast.CallExpression) string {
	text := ca.ctx.SourceFile.Text()
	pos := call.Expression.End()
	if call.TypeArguments != nil {
		pos = call.TypeArguments.End()
	}
	// Only `?.`, the type argument list's `>` and trivia can sit between the
	// callee and the argument list, so the next `(` opens it.
	for pos < call.End() {
		pos = scanner.SkipTrivia(text, pos)
		if text[pos] == '(' {
			break
		}
		pos++
	}
	return text[pos:call.End()]
}

func (ca *ChainAnalyzer) getNodeText(node *ast.Node) string {
	return scanner.GetSourceTextOfNodeFromSourceFile(ca.ctx.SourceFile, node, false)
}

func (ca *ChainAnalyzer) CheckNullishAndReport(node *ast.Node) bool {
	if !ca.opts.RequireNullish {
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

	msg := buildPreferOptionalChainMessage()
	sugMsg := buildOptionalChainSuggestMessage()
	reportRange := trimNodeRangeWithEnd(ca.ctx.SourceFile, node, node.End())

	// (foo || {}).bar is always a suggestion, not a fix
	ca.ctx.ReportRangeWithDeferredSuggestions(reportRange, msg, func() []rule.RuleSuggestion {
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

		return []rule.RuleSuggestion{{
			Message: sugMsg,
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(reportRange, fixCode),
			},
		}}
	})
}

// wrapChainCode wraps the generated optional chain code with the last operand's
// comparison wrapper (e.g., `!= null`, `!`, `typeof ... !== 'undefined'`).
func (ca *ChainAnalyzer) wrapChainCode(chainCode string, lastOperand Operand) string {
	if lastOperand.Validity == OperandLast {
		bin := lastOperand.Node.AsBinaryExpression()
		operator := lastComparisonText(lastOperand.LastComparison)
		if lastOperand.IsYoda {
			return ca.getNodeText(bin.Left) + " " + operator + " " + chainCode
		}
		return chainCode + " " + operator + " " + ca.getNodeText(bin.Right)
	}

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

func lastComparisonText(comparison LastComparisonType) string {
	switch comparison {
	case LastComparisonEqual:
		return "=="
	case LastComparisonStrictEqual:
		return "==="
	case LastComparisonNotEqual:
		return "!="
	}
	return "!=="
}
