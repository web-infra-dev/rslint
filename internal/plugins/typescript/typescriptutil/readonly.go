package typescriptutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ReadonlynessOptions mirrors the options shared by typescript-eslint rules
// that ask whether a TypeScript type is deeply readonly.
type ReadonlynessOptions struct {
	Allow                  []utils.TypeOrValueSpecifier
	TreatMethodsAsReadonly bool
}

type readonlyness uint8

const (
	readonlynessUnknown readonlyness = iota
	readonlynessMutable
	readonlynessReadonly
)

// IsTypeReadonly reports whether t is deeply readonly using the same rules as
// @typescript-eslint/type-utils' isTypeReadonly helper.
func IsTypeReadonly(
	typeChecker *checker.Checker,
	t *checker.Type,
	options ReadonlynessOptions,
	sourceProgram *program.Program,
) bool {
	return isTypeReadonly(typeChecker, t, options, sourceProgram, make(map[*checker.Type]struct{})) == readonlynessReadonly
}

func isTypeReadonly(
	typeChecker *checker.Checker,
	t *checker.Type,
	options ReadonlynessOptions,
	sourceProgram *program.Program,
	seen map[*checker.Type]struct{},
) readonlyness {
	if t == nil {
		return readonlynessMutable
	}
	seen[t] = struct{}{}

	// Upstream checks alias names before provenance-aware specifier matching.
	// This preserves aliases around unions, which otherwise lose their identity
	// while their constituents are inspected.
	if alias := checker.Type_alias(t); alias != nil && alias.Symbol() != nil {
		aliasName := alias.Symbol().Name
		for _, specifier := range options.Allow {
			if utils.Some(specifier.Name, func(name string) bool { return name == aliasName }) {
				return readonlynessReadonly
			}
		}
	}
	if utils.TypeMatchesSomeSpecifier(t, options.Allow, nil, sourceProgram) {
		return readonlynessReadonly
	}

	flags := checker.Type_flags(t)
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range t.Types() {
			if _, ok := seen[part]; ok {
				continue
			}
			if isTypeReadonly(typeChecker, part, options, sourceProgram, seen) != readonlynessReadonly {
				return readonlynessMutable
			}
		}
		return readonlynessReadonly
	}

	if flags&checker.TypeFlagsIntersection != 0 {
		parts := t.Types()
		containsArrayOrTuple := utils.Some(parts, func(part *checker.Type) bool {
			return typeChecker.IsArrayType(part) || checker.IsTupleType(part)
		})
		if containsArrayOrTuple {
			for _, part := range parts {
				if _, ok := seen[part]; ok {
					continue
				}
				if isTypeReadonly(typeChecker, part, options, sourceProgram, seen) != readonlynessReadonly {
					return readonlynessMutable
				}
			}
			return readonlynessReadonly
		}
		if result := isReadonlyObject(typeChecker, t, options, sourceProgram, seen); result != readonlynessUnknown {
			return result
		}
	}

	if flags&checker.TypeFlagsConditional != 0 {
		for _, part := range []*checker.Type{
			typeChecker.GetTrueTypeOfConditionalType(t),
			typeChecker.GetFalseTypeOfConditionalType(t),
		} {
			if _, ok := seen[part]; ok {
				continue
			}
			if isTypeReadonly(typeChecker, part, options, sourceProgram, seen) != readonlynessReadonly {
				return readonlynessMutable
			}
		}
		return readonlynessReadonly
	}

	if flags&checker.TypeFlagsObject == 0 {
		return readonlynessReadonly
	}

	if len(typeChecker.GetSignaturesOfType(t, checker.SignatureKindCall)) > 0 &&
		len(typeChecker.GetPropertiesOfType(t)) == 0 {
		return readonlynessReadonly
	}

	if result := isReadonlyArrayOrTuple(typeChecker, t, options, sourceProgram, seen); result != readonlynessUnknown {
		return result
	}
	return isReadonlyObject(typeChecker, t, options, sourceProgram, seen)
}

func isReadonlyArrayOrTuple(
	typeChecker *checker.Checker,
	t *checker.Type,
	options ReadonlynessOptions,
	sourceProgram *program.Program,
	seen map[*checker.Type]struct{},
) readonlyness {
	checkTypeArguments := func() readonlyness {
		for _, argument := range typeChecker.GetTypeArguments(t) {
			if _, ok := seen[argument]; ok {
				continue
			}
			if isTypeReadonly(typeChecker, argument, options, sourceProgram, seen) == readonlynessMutable {
				return readonlynessMutable
			}
		}
		return readonlynessReadonly
	}

	if typeChecker.IsArrayType(t) {
		if symbol := checker.Type_symbol(t); symbol != nil && symbol.Name == "Array" {
			return readonlynessMutable
		}
		return checkTypeArguments()
	}
	if checker.IsTupleType(t) {
		if !t.TargetTupleType().IsReadonly() {
			return readonlynessMutable
		}
		return checkTypeArguments()
	}
	return readonlynessUnknown
}

func isReadonlyObject(
	typeChecker *checker.Checker,
	t *checker.Type,
	options ReadonlynessOptions,
	sourceProgram *program.Program,
	seen map[*checker.Type]struct{},
) readonlyness {
	properties := typeChecker.GetPropertiesOfType(t)
	for _, property := range properties {
		if options.TreatMethodsAsReadonly && property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		if isPrivateProperty(property) {
			continue
		}
		if !isPropertyReadonly(typeChecker, t, property.Name) {
			return readonlynessMutable
		}
	}

	for _, property := range properties {
		// tsgo materializes the array's well-known symbol property as a mapped
		// object whose inherited method keys appear mutable. Upstream treats the same
		// well-known-symbol member as readonly; it cannot be assigned through a
		// normal property name, so keep it out of the deep value walk.
		if isWellKnownSymbolName(property.Name) {
			continue
		}
		propertyType := typeChecker.GetTypeOfPropertyOfType(t, property.Name)
		if propertyType == nil {
			return readonlynessMutable
		}
		if _, ok := seen[propertyType]; ok {
			continue
		}
		if isTypeReadonly(typeChecker, propertyType, options, sourceProgram, seen) == readonlynessMutable {
			return readonlynessMutable
		}
	}

	for _, indexInfo := range typeChecker.GetIndexInfosOfType(t) {
		if !indexInfo.IsReadonly() {
			return readonlynessMutable
		}
		indexType := indexInfo.ValueType()
		if indexType == t {
			continue
		}
		if _, ok := seen[indexType]; ok {
			continue
		}
		if isTypeReadonly(typeChecker, indexType, options, sourceProgram, seen) == readonlynessMutable {
			return readonlynessMutable
		}
	}

	return readonlynessReadonly
}

func isPrivateProperty(property *ast.Symbol) bool {
	declaration := property.ValueDeclaration
	if declaration == nil {
		return false
	}
	name := ast.GetNameOfDeclaration(declaration)
	return name != nil && ast.IsPrivateIdentifier(name)
}

func isPropertyReadonly(typeChecker *checker.Checker, t *checker.Type, name string) bool {
	if checker.Type_flags(t)&checker.TypeFlagsIntersection != 0 {
		for _, part := range t.Types() {
			if typeChecker.GetPropertyOfType(part, name) != nil && isPropertyReadonly(typeChecker, part, name) {
				return true
			}
		}
		return false
	}

	// Mapped types can synthesize transient property symbols whose declarations
	// point back to the original (mutable) member. Read the mapped declaration's
	// readonly modifier before falling back to those declarations. As upstream
	// does, leave well-known symbol properties to their original declarations.
	if checker.Type_objectFlags(t)&checker.ObjectFlagsMapped != 0 && !isWellKnownSymbolName(name) {
		if symbol := checker.Type_symbol(t); symbol != nil && len(symbol.Declarations) > 0 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindMappedType {
				readonlyToken := declaration.AsMappedTypeNode().ReadonlyToken
				if readonlyToken != nil {
					return readonlyToken.Kind != ast.KindMinusToken
				}
			}
		}
	}

	property := typeChecker.GetPropertyOfType(t, name)
	if property == nil {
		return false
	}
	if property.CheckFlags&ast.CheckFlagsReadonly != 0 || property.Flags&ast.SymbolFlagsValueModule != 0 {
		return true
	}
	if property.Flags&ast.SymbolFlagsGetAccessor != 0 && property.Flags&ast.SymbolFlagsSetAccessor == 0 {
		return true
	}
	if property.Flags&ast.SymbolFlagsEnumMember != 0 {
		return true
	}
	for _, declaration := range property.Declarations {
		if declaration.ModifierFlags()&ast.ModifierFlagsReadonly != 0 {
			return true
		}
	}
	return checker.GetDeclarationModifierFlagsFromSymbol(property)&ast.ModifierFlagsReadonly != 0
}

func isWellKnownSymbolName(name string) bool {
	if !strings.HasPrefix(name, ast.InternalSymbolNamePrefix+"@") {
		return false
	}
	rest := strings.TrimPrefix(name, ast.InternalSymbolNamePrefix+"@")
	return strings.Contains(rest, "@")
}

// IsTypeBrandedLiteralLike reports whether t is a branded primitive/literal
// intersection, or a union made entirely from such intersections.
func IsTypeBrandedLiteralLike(typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}
	if checker.Type_flags(t)&checker.TypeFlagsUnion != 0 {
		return utils.Every(t.Types(), func(part *checker.Type) bool {
			return isTypeBrandedLiteral(typeChecker, part)
		})
	}
	return isTypeBrandedLiteral(typeChecker, t)
}

func isTypeBrandedLiteral(typeChecker *checker.Checker, t *checker.Type) bool {
	if checker.Type_flags(t)&checker.TypeFlagsIntersection == 0 {
		return false
	}
	hadObjectLike := false
	hadPrimitiveLike := false
	for _, part := range t.Types() {
		flags := checker.Type_flags(part)
		if flags&checker.TypeFlagsObject != 0 &&
			len(typeChecker.GetSignaturesOfType(part, checker.SignatureKindCall)) == 0 &&
			len(typeChecker.GetSignaturesOfType(part, checker.SignatureKindConstruct)) == 0 {
			hadObjectLike = true
		} else if flags&(checker.TypeFlagsLiteral|checker.TypeFlagsBigInt|checker.TypeFlagsNumber|checker.TypeFlagsString|checker.TypeFlagsTemplateLiteral) != 0 {
			hadPrimitiveLike = true
		} else {
			return false
		}
	}
	return hadObjectLike && hadPrimitiveLike
}
