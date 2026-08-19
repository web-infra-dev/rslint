package unicornutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TypeClass is Unicorn's three-state view of an expression type. Unknown is
// deliberately distinct from NonTarget: rules generally keep reporting when
// a receiver might still be the target type.
type TypeClass uint8

const (
	TypeUnknown TypeClass = iota
	TypeTarget
	TypeNonTarget
)

// TypeClassifierOptions configures the shared recursive portion of Unicorn's
// createTypeCheckers helper. Syntactic target/non-target recognition remains
// with each rule because those shapes are domain-specific.
type TypeClassifierOptions struct {
	TargetTypeNames            *utils.Set[string]
	NonTargetTypeNames         *utils.Set[string]
	IsTargetType               func(*checker.Type) bool
	HeritageSymbolFlags        ast.SymbolFlags
	AllowNullishInMixedUnion   bool
	TreatMixedUnionAsTarget    bool
	TreatMixedUnionAsNonTarget bool
}

// ClassifyType mirrors eslint-plugin-unicorn's getTypeScriptType plus the
// normalizeType step used by createTypeCheckers.
func ClassifyType(ctx rule.RuleContext, t *checker.Type, options TypeClassifierOptions) TypeClass {
	if t == nil {
		return TypeUnknown
	}
	if utils.IsTypeAnyType(t) || utils.IsTypeUnknownType(t) || utils.IsIntrinsicErrorType(t) {
		return TypeUnknown
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
		return TypeNonTarget
	}

	if utils.IsTypeParameter(t) {
		constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
		if constraint == nil {
			return TypeUnknown
		}
		return ClassifyType(ctx, constraint, options)
	}

	if utils.IsUnionType(t) {
		return classifyUnion(ctx, utils.UnionTypeParts(t), options)
	}
	if utils.IsIntersectionType(t) {
		return classifyIntersection(ctx, utils.IntersectionTypeParts(t), options)
	}

	if options.IsTargetType != nil && options.IsTargetType(t) {
		return TypeTarget
	}

	constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
	if constraint != nil && constraint != t {
		return ClassifyType(ctx, constraint, options)
	}

	name, ok := TypeSymbolName(t)
	if !ok {
		if utils.IsTypeFlagSet(t, checker.TypeFlagsPrimitive|checker.TypeFlagsIntrinsic) {
			return TypeNonTarget
		}
		return TypeUnknown
	}
	if options.TargetTypeNames != nil && options.TargetTypeNames.Has(name) {
		return TypeTarget
	}
	if options.NonTargetTypeNames != nil && options.NonTargetTypeNames.Has(name) {
		return TypeNonTarget
	}
	if options.HeritageSymbolFlags != 0 &&
		classifyHeritage(ctx, t, options) == TypeTarget {
		return TypeTarget
	}
	return TypeNonTarget
}

func classifyUnion(ctx rule.RuleContext, parts []*checker.Type, options TypeClassifierOptions) TypeClass {
	classes := make([]TypeClass, 0, len(parts))
	for _, part := range parts {
		if options.AllowNullishInMixedUnion &&
			utils.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		classes = append(classes, ClassifyType(ctx, part, options))
	}
	if len(classes) == 0 {
		return TypeNonTarget
	}

	anyTarget := false
	allTarget := true
	anyNonTarget := false
	allNonTarget := true
	for _, class := range classes {
		anyTarget = anyTarget || class == TypeTarget
		allTarget = allTarget && class == TypeTarget
		anyNonTarget = anyNonTarget || class == TypeNonTarget
		allNonTarget = allNonTarget && class == TypeNonTarget
	}

	if (options.TreatMixedUnionAsTarget && anyTarget) ||
		(!options.TreatMixedUnionAsTarget && allTarget) {
		return TypeTarget
	}
	if (options.TreatMixedUnionAsNonTarget && anyNonTarget) ||
		(!options.TreatMixedUnionAsNonTarget && allNonTarget) {
		return TypeNonTarget
	}
	return TypeUnknown
}

func classifyIntersection(ctx rule.RuleContext, parts []*checker.Type, options TypeClassifierOptions) TypeClass {
	allNonTarget := true
	for _, part := range parts {
		class := ClassifyType(ctx, part, options)
		if class == TypeTarget {
			return TypeTarget
		}
		allNonTarget = allNonTarget && class == TypeNonTarget
	}
	if allNonTarget {
		return TypeNonTarget
	}
	return TypeUnknown
}

func classifyHeritage(ctx rule.RuleContext, t *checker.Type, options TypeClassifierOptions) TypeClass {
	symbol := checker.Type_symbol(t)
	if symbol == nil || symbol.Flags&options.HeritageSymbolFlags == 0 {
		return TypeUnknown
	}
	declared := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, symbol)
	if declared == nil {
		return TypeUnknown
	}
	for _, base := range checker.Checker_getBaseTypes(ctx.TypeChecker, declared) {
		if ClassifyType(ctx, base, options) == TypeTarget {
			return TypeTarget
		}
	}
	return TypeUnknown
}

// TypeSymbolName returns a type's direct symbol name, falling back to its
// alias symbol, just like Unicorn's getTypeSymbol helper.
func TypeSymbolName(t *checker.Type) (string, bool) {
	if sym := checker.Type_symbol(t); sym != nil && sym.Name != "" {
		return sym.Name, true
	}
	if alias := checker.Type_alias(t); alias != nil {
		if sym := alias.Symbol(); sym != nil && sym.Name != "" {
			return sym.Name, true
		}
	}
	return "", false
}
