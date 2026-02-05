package prefer_optional_chain

import (
	"reflect"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type OperandValidity int

const (
	OperandValid OperandValidity = iota
	OperandInvalid
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
}

type OperandAnalyzer struct {
	ctx  rule.RuleContext
	opts PreferOptionalChainOptions
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

	operands := make([]Operand, 0, 4)
	a.flattenLogicalOperands(node, operator, &operands)
	return operands, operator
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

	return Operand{
		Node:     node,
		Validity: OperandInvalid,
	}
}

func (a *OperandAnalyzer) classifyComparisonOperand(bin *ast.BinaryExpression, chainOperator ast.Kind) (Operand, bool) {
	left := ast.SkipParentheses(bin.Left)
	right := ast.SkipParentheses(bin.Right)
	opKind := bin.OperatorToken.Kind

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

		var compType NullishComparisonType
		isNegated := isNegatedComparison(opKind, chainOperator)

		switch {
		case opKind == ast.KindExclamationEqualsToken || opKind == ast.KindEqualsEqualsToken:
			if isNegated {
				compType = ComparisonNotEqualNullOrUndefined
			} else {
				compType = ComparisonEqualNullOrUndefined
			}
		case (opKind == ast.KindExclamationEqualsEqualsToken || opKind == ast.KindEqualsEqualsEqualsToken) && isNull:
			if isNegated {
				compType = ComparisonNotStrictEqualNull
			} else {
				compType = ComparisonStrictEqualNull
			}
		case (opKind == ast.KindExclamationEqualsEqualsToken || opKind == ast.KindEqualsEqualsEqualsToken) && isUndefined:
			if isNegated {
				compType = ComparisonNotStrictEqualUndefined
			} else {
				compType = ComparisonStrictEqualUndefined
			}
		default:
			continue
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

		isNegated := isNegatedComparison(opKind, chainOperator)
		var compType NullishComparisonType
		if isNegated {
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

func isNegatedComparison(opKind ast.Kind, chainOperator ast.Kind) bool {
	isNegatedOp := opKind == ast.KindExclamationEqualsToken || opKind == ast.KindExclamationEqualsEqualsToken
	if chainOperator == ast.KindAmpersandAmpersandToken {
		return isNegatedOp
	}
	// For || chains, the sense is reversed
	return !isNegatedOp
}

func invertComparisonType(ct NullishComparisonType) NullishComparisonType {
	switch ct {
	case ComparisonBoolean:
		return ComparisonNotBoolean
	case ComparisonNotBoolean:
		return ComparisonBoolean
	case ComparisonNotEqualNullOrUndefined:
		return ComparisonEqualNullOrUndefined
	case ComparisonEqualNullOrUndefined:
		return ComparisonNotEqualNullOrUndefined
	case ComparisonNotStrictEqualNull:
		return ComparisonStrictEqualNull
	case ComparisonStrictEqualNull:
		return ComparisonNotStrictEqualNull
	case ComparisonNotStrictEqualUndefined:
		return ComparisonStrictEqualUndefined
	case ComparisonStrictEqualUndefined:
		return ComparisonNotStrictEqualUndefined
	}
	return ct
}

func (a *OperandAnalyzer) isValidBooleanCheckType(node *ast.Node) bool {
	if a.ctx.TypeChecker == nil {
		return true
	}

	t := a.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return true
	}

	// Check for falsy literal unions: if the type is a union containing a falsy
	// literal (false, 0, '', 0n) alongside an object type, but NO null/undefined/void,
	// then the truthiness check is being used as a type discriminator
	// (e.g., `false | { a: string }`), not as a null guard.
	// Don't suggest optional chaining in this case.
	// Note: we require an object type in the union to distinguish discriminated unions
	// (like `false | { a: string }`) from plain primitive types (like `boolean` = `true | false`).
	parts := utils.UnionTypeParts(t)
	if len(parts) > 1 {
		hasFalsyLiteral := false
		hasNullUndefined := false
		hasObjectType := false
		for _, part := range parts {
			pFlags := checker.Type_flags(part)
			if pFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
				hasNullUndefined = true
			}
			if pFlags&checker.TypeFlagsObject != 0 {
				hasObjectType = true
			}
			if isFalsyLiteralType(part) {
				hasFalsyLiteral = true
			}
		}
		if hasFalsyLiteral && !hasNullUndefined && hasObjectType {
			return false
		}
	}

	opts := a.opts
	for _, part := range parts {
		flags := checker.Type_flags(part)

		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			continue
		}

		if derefBoolDefault(opts.CheckAny, true) && flags&checker.TypeFlagsAny != 0 {
			continue
		}
		if derefBoolDefault(opts.CheckUnknown, true) && flags&checker.TypeFlagsUnknown != 0 {
			continue
		}
		if derefBoolDefault(opts.CheckString, true) && flags&checker.TypeFlagsStringLike != 0 {
			continue
		}
		if derefBoolDefault(opts.CheckNumber, true) && flags&checker.TypeFlagsNumberLike != 0 {
			continue
		}
		if derefBoolDefault(opts.CheckBoolean, true) && flags&checker.TypeFlagsBooleanLike != 0 {
			continue
		}
		if derefBoolDefault(opts.CheckBigInt, true) && flags&checker.TypeFlagsBigIntLike != 0 {
			continue
		}
		if flags&checker.TypeFlagsTypeParameter != 0 {
			constraint := checker.Checker_getBaseConstraintOfType(a.ctx.TypeChecker, part)
			if constraint == nil {
				continue
			}
			// Recurse into constraint
			constraintValid := true
			for _, cPart := range utils.UnionTypeParts(constraint) {
				cFlags := checker.Type_flags(cPart)
				if cFlags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
					continue
				}
				if !a.isTypePartValidForBoolean(cFlags) {
					constraintValid = false
					break
				}
			}
			if constraintValid {
				continue
			}
		}

		if flags&checker.TypeFlagsObject != 0 {
			// object types are always truthy (when not null/undefined),
			// so boolean coercion is safe for null/undefined guards
			continue
		}

		if flags&checker.TypeFlagsNever != 0 {
			continue
		}

		if flags&checker.TypeFlagsEnum != 0 || flags&checker.TypeFlagsEnumLiteral != 0 {
			continue
		}

		return false
	}

	return true
}

func (a *OperandAnalyzer) isTypePartValidForBoolean(flags checker.TypeFlags) bool {
	opts := a.opts
	if derefBoolDefault(opts.CheckAny, true) && flags&checker.TypeFlagsAny != 0 {
		return true
	}
	if derefBoolDefault(opts.CheckUnknown, true) && flags&checker.TypeFlagsUnknown != 0 {
		return true
	}
	if derefBoolDefault(opts.CheckString, true) && flags&checker.TypeFlagsStringLike != 0 {
		return true
	}
	if derefBoolDefault(opts.CheckNumber, true) && flags&checker.TypeFlagsNumberLike != 0 {
		return true
	}
	if derefBoolDefault(opts.CheckBoolean, true) && flags&checker.TypeFlagsBooleanLike != 0 {
		return true
	}
	if derefBoolDefault(opts.CheckBigInt, true) && flags&checker.TypeFlagsBigIntLike != 0 {
		return true
	}
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
		return true
	}
	return false
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
// false, 0, '', or 0n. These appear in discriminated unions like `false | { a: string }`
// where truthiness is used as a type discriminator rather than a null guard.
func isFalsyLiteralType(t *checker.Type) bool {
	flags := checker.Type_flags(t)

	// Check BooleanLiteral (false), NumberLiteral (0), StringLiteral (''), BigIntLiteral (0n)
	if flags&(checker.TypeFlagsBooleanLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsStringLiteral|checker.TypeFlagsBigIntLiteral) != 0 {
		literal := t.AsLiteralType()
		if literal != nil {
			val := checker.LiteralType_value(literal)
			if val != nil && reflect.ValueOf(val).IsZero() {
				return true
			}
		}
	}

	return false
}

