package prefer_optional_chain

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type OperandValidity int

const (
	OperandValid OperandValidity = iota
	OperandInvalid
	// OperandLast is a `x.y !== <value>` comparison against something other
	// than a nullish literal. It cannot guard a chain, only close one.
	OperandLast
)

// LastComparisonType is the source-level operator of an OperandLast operand.
type LastComparisonType int

const (
	LastComparisonNotEqual LastComparisonType = iota
	LastComparisonEqual
	LastComparisonNotStrictEqual
	LastComparisonStrictEqual
)

// YodaKind records which side of an OperandLast comparison holds the chain.
// It is Unknown when both sides could, i.e. `a.b !== a.c`.
type YodaKind int

const (
	YodaNo YodaKind = iota
	YodaYes
	YodaUnknown
)

type NullishComparisonType int

const (
	ComparisonNotEqualNullOrUndefined NullishComparisonType = iota // x != null, x != undefined
	ComparisonNotStrictEqualNull                                   // x !== null
	ComparisonNotStrictEqualUndefined                              // x !== undefined, typeof x !== 'undefined'
	ComparisonBoolean                                              // x (truthy check)
	ComparisonNotBoolean                                           // !x (falsy check, for || chains)
	ComparisonEqualNullOrUndefined                                 // x == null, x == undefined
	ComparisonStrictEqualNull                                      // x === null
	ComparisonStrictEqualUndefined                                 // x === undefined, typeof x === 'undefined'
)

type Operand struct {
	Node           *ast.Node
	ComparedNode   *ast.Node // the node being compared (e.g., `foo` in `foo != null`)
	ComparisonType NullishComparisonType
	Validity       OperandValidity
	IsYoda         bool
	IsTypeof       bool // whether this was a typeof check (typeof x !== 'undefined')
	UsesNull       bool // for != / == comparisons, whether null was used (vs undefined)

	// For OperandLast only: the expression ComparedNode is compared against,
	// and the operator joining them.
	ComparisonValue *ast.Node
	LastComparison  LastComparisonType
	Yoda            YodaKind
}

type OperandAnalyzer struct {
	ctx      rule.RuleContext
	opts     PreferOptionalChainOptions
	operands []Operand
}

func NewOperandAnalyzer(ctx rule.RuleContext, opts PreferOptionalChainOptions) *OperandAnalyzer {
	return &OperandAnalyzer{ctx: ctx, opts: opts}
}

func (a *OperandAnalyzer) GatherLogicalOperands(node *ast.Node) ([]Operand, ast.Kind) {
	bin := node.AsBinaryExpression()
	operator := bin.OperatorToken.Kind

	if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken {
		return nil, operator
	}

	operandCount := countLogicalOperands(node, operator)
	if cap(a.operands) < operandCount {
		a.operands = make([]Operand, 0, operandCount)
	} else {
		a.operands = a.operands[:0]
	}
	a.flattenLogicalOperands(node, operator, &a.operands)
	operands := a.operands

	// The last operand in the chain is not used as a guard — it's the chain
	// target — so re-classify it without the boolean-check type restriction
	// that only makes sense for a guard.
	if len(operands) >= 2 {
		last := &operands[len(operands)-1]
		if last.Validity == OperandInvalid {
			raw := ast.SkipParentheses(last.Node)
			if operator == ast.KindBarBarToken && ast.IsPrefixUnaryExpression(raw) {
				prefix := raw.AsPrefixUnaryExpression()
				if prefix.Operator == ast.KindExclamationToken {
					inner := ast.SkipParentheses(prefix.Operand)
					if isValidChainTarget(inner, true) {
						last.ComparedNode = inner
						last.ComparisonType = ComparisonNotBoolean
						last.Validity = OperandValid
					}
				}
			} else if operator == ast.KindAmpersandAmpersandToken && isValidChainTarget(raw, true) {
				last.ComparedNode = raw
				last.ComparisonType = ComparisonBoolean
				last.Validity = OperandValid
			}
		}
	}

	return operands, operator
}

func countLogicalOperands(node *ast.Node, operator ast.Kind) int {
	unwrapped := ast.SkipParentheses(node)
	if ast.IsBinaryExpression(unwrapped) {
		bin := unwrapped.AsBinaryExpression()
		if bin.OperatorToken.Kind == operator {
			return countLogicalOperands(bin.Left, operator) + countLogicalOperands(bin.Right, operator)
		}
	}
	return 1
}

func (a *OperandAnalyzer) flattenLogicalOperands(node *ast.Node, operator ast.Kind, operands *[]Operand) {
	// Skip parentheses to handle cases like `a && (a.b && a.b.c)`
	unwrapped := ast.SkipParentheses(node)
	if ast.IsBinaryExpression(unwrapped) {
		bin := unwrapped.AsBinaryExpression()
		if bin.OperatorToken.Kind == operator {
			a.flattenLogicalOperands(bin.Left, operator, operands)
			a.flattenLogicalOperands(bin.Right, operator, operands)
			return
		}
	}

	operand := a.classifyOperand(node, operator)
	*operands = append(*operands, operand)
}

func (a *OperandAnalyzer) classifyOperand(node *ast.Node, chainOperator ast.Kind) Operand {
	raw := ast.SkipParentheses(node)

	// Handle `!expr` for || chains (DeMorgan)
	if chainOperator == ast.KindBarBarToken && ast.IsPrefixUnaryExpression(raw) {
		prefix := raw.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindExclamationToken {
			inner := ast.SkipParentheses(prefix.Operand)
			if isValidChainTarget(inner, true) {
				if a.isValidBooleanCheckType(inner) {
					return Operand{
						Node:           node,
						ComparedNode:   inner,
						ComparisonType: ComparisonNotBoolean,
						Validity:       OperandValid,
					}
				}
			}
		}
	}

	// Handle comparison expressions: x != null, x !== undefined, etc.
	if ast.IsBinaryExpression(raw) {
		bin := raw.AsBinaryExpression()
		result, ok := a.classifyComparisonOperand(bin, chainOperator)
		if ok {
			return result
		}
	}

	// Handle typeof x !== 'undefined'
	if ast.IsBinaryExpression(raw) {
		bin := raw.AsBinaryExpression()
		result, ok := a.classifyTypeofOperand(bin, chainOperator)
		if ok {
			return result
		}
	}

	// Handle bare truthy check: `x` in && chain
	if chainOperator == ast.KindAmpersandAmpersandToken && isValidChainTarget(raw, true) {
		if a.isValidBooleanCheckType(raw) {
			return Operand{
				Node:           node,
				ComparedNode:   raw,
				ComparisonType: ComparisonBoolean,
				Validity:       OperandValid,
			}
		}
	}

	// Handle `x.y != something` — a comparison against an arbitrary value.
	// A comparison against `'undefined'` is only a nullish check with typeof,
	// which was already handled above, so it never reaches this point valid.
	if ast.IsBinaryExpression(raw) {
		bin := raw.AsBinaryExpression()
		if !isUndefinedStringLiteral(ast.SkipParentheses(bin.Left)) &&
			!isUndefinedStringLiteral(ast.SkipParentheses(bin.Right)) {
			result, ok := classifyLastChainOperand(bin)
			if ok {
				return result
			}
		}
	}

	return Operand{
		Node:     node,
		Validity: OperandInvalid,
	}
}

// classifyLastChainOperand handles comparisons like `a.b !== value`, where the
// compared value is an arbitrary expression rather than a nullish literal. At
// least one side must be member-based for a chain to be foldable into it.
func classifyLastChainOperand(bin *ast.BinaryExpression) (Operand, bool) {
	var comparison LastComparisonType
	switch bin.OperatorToken.Kind {
	case ast.KindEqualsEqualsToken:
		comparison = LastComparisonEqual
	case ast.KindEqualsEqualsEqualsToken:
		comparison = LastComparisonStrictEqual
	case ast.KindExclamationEqualsToken:
		comparison = LastComparisonNotEqual
	case ast.KindExclamationEqualsEqualsToken:
		comparison = LastComparisonNotStrictEqual
	default:
		return Operand{}, false
	}

	left := ast.SkipParentheses(bin.Left)
	right := ast.SkipParentheses(bin.Right)
	leftIsMember := isMemberBasedExpression(left)
	rightIsMember := isMemberBasedExpression(right)

	operand := Operand{
		Node:           bin.AsNode(),
		Validity:       OperandLast,
		LastComparison: comparison,
	}
	switch {
	case leftIsMember && !rightIsMember:
		operand.ComparedNode, operand.ComparisonValue, operand.Yoda = left, right, YodaNo
	case !leftIsMember && rightIsMember:
		operand.ComparedNode, operand.ComparisonValue, operand.Yoda = right, left, YodaYes
	case leftIsMember && rightIsMember:
		operand.ComparedNode, operand.ComparisonValue, operand.Yoda = left, right, YodaUnknown
	default:
		return Operand{}, false
	}
	return operand, true
}

// isMemberBasedExpression reports whether the node is a property access or a
// call on one, i.e. `a.b`, `a[b]`, `a.b()`.
func isMemberBasedExpression(node *ast.Node) bool {
	// ESTree wraps the complete optional expression in a ChainExpression,
	// which is not a MemberExpression or CallExpression. Match that boundary
	// explicitly because tsgo represents the same expression as its outermost
	// optional-chain node without a wrapper.
	if ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node) {
		return false
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return true
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		return callee.Kind == ast.KindPropertyAccessExpression ||
			callee.Kind == ast.KindElementAccessExpression
	}
	return false
}

func (a *OperandAnalyzer) classifyComparisonOperand(bin *ast.BinaryExpression, chainOperator ast.Kind) (Operand, bool) {
	left := ast.SkipParentheses(bin.Left)
	right := ast.SkipParentheses(bin.Right)
	opKind := bin.OperatorToken.Kind

	if !isEqualityOperator(opKind) {
		return Operand{}, false
	}

	// Try both orientations for Yoda-style: `null != foo`
	for _, yoda := range []bool{false, true} {
		var testExpr, checkExpr *ast.Node
		if yoda {
			testExpr, checkExpr = right, left
		} else {
			testExpr, checkExpr = left, right
		}

		if !isValidChainTarget(testExpr, true) {
			continue
		}

		isNull := checkExpr.Kind == ast.KindNullKeyword
		isUndefined := ast.IsIdentifier(checkExpr) && checkExpr.Text() == "undefined"

		if !isNull && !isUndefined {
			continue
		}

		// A guard only narrows the chain when its sense matches the operator:
		// `x != null && x.y` and `x == null || x.y`.
		negated := isNegatedOperator(opKind)
		if negated == (chainOperator == ast.KindBarBarToken) {
			continue
		}

		var compType NullishComparisonType
		switch {
		case opKind == ast.KindExclamationEqualsToken || opKind == ast.KindEqualsEqualsToken:
			if negated {
				compType = ComparisonNotEqualNullOrUndefined
			} else {
				compType = ComparisonEqualNullOrUndefined
			}
		case isNull:
			if negated {
				compType = ComparisonNotStrictEqualNull
			} else {
				compType = ComparisonStrictEqualNull
			}
		default:
			if negated {
				compType = ComparisonNotStrictEqualUndefined
			} else {
				compType = ComparisonStrictEqualUndefined
			}
		}

		return Operand{
			Node:           bin.AsNode(),
			ComparedNode:   testExpr,
			ComparisonType: compType,
			Validity:       OperandValid,
			IsYoda:         yoda,
			UsesNull:       isNull,
		}, true
	}

	return Operand{}, false
}

func (a *OperandAnalyzer) classifyTypeofOperand(bin *ast.BinaryExpression, chainOperator ast.Kind) (Operand, bool) {
	left := ast.SkipParentheses(bin.Left)
	right := ast.SkipParentheses(bin.Right)
	opKind := bin.OperatorToken.Kind

	if opKind != ast.KindExclamationEqualsEqualsToken && opKind != ast.KindEqualsEqualsEqualsToken {
		return Operand{}, false
	}

	for _, yoda := range []bool{false, true} {
		var typeofExpr, stringExpr *ast.Node
		if yoda {
			typeofExpr, stringExpr = right, left
		} else {
			typeofExpr, stringExpr = left, right
		}

		if !ast.IsTypeOfExpression(typeofExpr) {
			continue
		}

		if !ast.IsStringLiteral(stringExpr) || stringExpr.Text() != "undefined" {
			continue
		}

		inner := typeofExpr.AsTypeOfExpression().Expression
		if !isValidChainTarget(inner, true) {
			continue
		}

		// `typeof globalThis !== 'undefined'` asks whether the global exists at
		// all, which an optional chain cannot express.
		if ast.IsIdentifier(inner) {
			isGlobal, resolvedInCurrentFile := a.classifyCurrentFileTypeofIdentifier(inner)
			if !resolvedInCurrentFile {
				isGlobal = typescriptutil.IsReferenceToGlobalIdentifier(a.ctx, inner)
			}
			if isGlobal {
				return Operand{Node: bin.AsNode(), Validity: OperandInvalid}, true
			}
		}

		var compType NullishComparisonType
		if isNegatedOperator(opKind) {
			compType = ComparisonNotStrictEqualUndefined
		} else {
			compType = ComparisonStrictEqualUndefined
		}

		return Operand{
			Node:           bin.AsNode(),
			ComparedNode:   inner,
			ComparisonType: compType,
			Validity:       OperandValid,
			IsTypeof:       true,
			IsYoda:         yoda,
		}, true
	}

	return Operand{}, false
}

// classifyCurrentFileTypeofIdentifier decides identifiers whose checker symbol
// has a declaration in this file. A declaration in `declare global` is global:
// it creates no runtime lexical binding, so removing a typeof guard could make
// an absent host global throw. Any ordinary same-file declaration shadows it.
// Identifiers with no same-file declaration are left to the shared scope
// fallback, which covers lib globals and checker resolution edge cases.
func (a *OperandAnalyzer) classifyCurrentFileTypeofIdentifier(ident *ast.Node) (isGlobal, resolved bool) {
	if a.ctx.TypeChecker == nil {
		return false, false
	}
	symbol := a.ctx.TypeChecker.GetSymbolAtLocation(ident)
	if symbol == nil {
		return false, false
	}

	foundGlobalAugmentation := false
	for _, decl := range symbol.Declarations {
		if decl == nil || ast.GetSourceFileOfNode(decl) != a.ctx.SourceFile {
			continue
		}
		if decl.Flags&ast.NodeFlagsAmbient == 0 {
			return false, true
		}
		inGlobalAugmentation := false
		for parent := decl.Parent; parent != nil; parent = parent.Parent {
			if ast.IsGlobalScopeAugmentation(parent) {
				inGlobalAugmentation = true
				break
			}
		}
		if !inGlobalAugmentation {
			return false, true
		}
		foundGlobalAugmentation = true
	}
	return foundGlobalAugmentation, foundGlobalAugmentation
}

func isEqualityOperator(opKind ast.Kind) bool {
	switch opKind {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return true
	}
	return false
}

func isNegatedOperator(opKind ast.Kind) bool {
	return opKind == ast.KindExclamationEqualsToken || opKind == ast.KindExclamationEqualsEqualsToken
}

// isUndefinedStringLiteral reports whether one side of a comparison is the
// string `'undefined'`, which only forms a nullish check together with typeof.
func isUndefinedStringLiteral(node *ast.Node) bool {
	return ast.IsStringLiteral(node) && node.Text() == "undefined"
}

// isValidBooleanCheckType reports whether a truthiness check on this node can
// stand in for a nullish guard. It rejects types that a chain would not narrow
// the same way, and types the check* options have opted out of.
func (a *OperandAnalyzer) isValidBooleanCheckType(node *ast.Node) bool {
	if a.ctx.TypeChecker == nil {
		return true
	}

	t := a.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return true
	}

	// Intersection members are compared individually, so `({a: string} & {b: string}) | null`
	// is judged by its object parts rather than by the intersection as a whole.
	var parts []*checker.Type
	for _, part := range utils.UnionTypeParts(t) {
		parts = append(parts, utils.IntersectionTypeParts(part)...)
	}

	// A falsy literal (false, 0, '', 0n) means the check narrows out a
	// non-nullish falsy value rather than guarding against null/undefined, so
	// `declare const x: false | {a: string}; x && x.a` must stay as it is.
	for _, part := range parts {
		if isFalsyLiteralType(part) {
			return false
		}
	}

	allowed := checker.TypeFlagsNull | checker.TypeFlagsUndefined | checker.TypeFlagsObject
	opts := a.opts
	if opts.CheckAny {
		allowed |= checker.TypeFlagsAny
	}
	if opts.CheckUnknown {
		allowed |= checker.TypeFlagsUnknown
	}
	if opts.CheckString {
		allowed |= checker.TypeFlagsStringLike
	}
	if opts.CheckNumber {
		allowed |= checker.TypeFlagsNumberLike
	}
	if opts.CheckBoolean {
		allowed |= checker.TypeFlagsBooleanLike
	}
	if opts.CheckBigInt {
		allowed |= checker.TypeFlagsBigIntLike
	}

	for _, part := range parts {
		if checker.Type_flags(part)&allowed == 0 {
			return false
		}
	}
	return true
}

func isValidChainTarget(node *ast.Node, allowIdentifier bool) bool {
	n := ast.SkipParentheses(node)
	switch n.Kind {
	case ast.KindIdentifier:
		return allowIdentifier
	case ast.KindThisKeyword:
		return allowIdentifier
	case ast.KindPropertyAccessExpression:
		// Private identifiers (#foo) cannot use optional chaining
		prop := n.AsPropertyAccessExpression()
		if ast.IsPrivateIdentifier(prop.Name()) {
			return false
		}
		return true
	case ast.KindElementAccessExpression:
		return true
	case ast.KindCallExpression:
		return true
	case ast.KindMetaProperty:
		return true
	case ast.KindNonNullExpression:
		return isValidChainTarget(n.Expression(), allowIdentifier)
	}
	return false
}

// isFalsyLiteralType checks if a type is a specific falsy literal value:
// false, 0, ”, or 0n. These appear in discriminated unions like `false | { a: string }`
// where truthiness is used as a type discriminator rather than a null guard.
func isFalsyLiteralType(t *checker.Type) bool {
	flags := checker.Type_flags(t)

	if utils.IsFalseLiteralType(t) {
		return true
	}
	if flags&checker.TypeFlagsStringLiteral != 0 {
		if s, ok := t.AsLiteralType().Value().(string); ok {
			return s == ""
		}
	}
	if flags&checker.TypeFlagsNumberLiteral != 0 {
		return utils.IsNumberLiteralZeroOrNaN(t.AsLiteralType().Value())
	}
	if flags&checker.TypeFlagsBigIntLiteral != 0 {
		return checker.ValueToString(t.AsLiteralType().Value()) == "0n"
	}

	return false
}
