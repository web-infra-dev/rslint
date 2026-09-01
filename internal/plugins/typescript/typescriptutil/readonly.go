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
		conditional := t.AsConditionalType()
		root := checker.ConditionalType_root(conditional)
		node := checker.ConditionalRoot_node(root)
		for _, part := range []*checker.Type{
			typeChecker.GetTypeFromTypeNode(node.TrueType),
			typeChecker.GetTypeFromTypeNode(node.FalseType),
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
		if options.TreatMethodsAsReadonly && isMethodProperty(property) {
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
		if isPrivateProperty(property) {
			continue
		}
		// Upstream's computed-symbol lookup reads the property's declared type,
		// which is `any` for these escaped property symbols. tsgo's equivalent API
		// exposes the authored value type instead, so omit the deep value walk while
		// retaining the readonly-property check above.
		if strings.HasPrefix(property.Name, ast.InternalSymbolNamePrefix+"@") {
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
		keyFlags := checker.Type_flags(indexInfo.KeyType())
		if keyFlags&(checker.TypeFlagsString|checker.TypeFlagsNumber) == 0 {
			continue
		}
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

func isMethodProperty(property *ast.Symbol) bool {
	if property.Flags&ast.SymbolFlagsMethod != 0 {
		return true
	}
	if declaration := property.ValueDeclaration; declaration != nil {
		if symbol := declaration.Symbol(); symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 {
			return true
		}
	}
	if len(property.Declarations) > 0 {
		if symbol := property.Declarations[len(property.Declarations)-1].Symbol(); symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0 {
			return true
		}
	}
	return false
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
	// does, apply explicit modifiers to every key spelling and inherit a source
	// property's modifier only when the name type filters rather than remaps.
	if checker.Type_objectFlags(t)&checker.ObjectFlagsMapped != 0 {
		if symbol := checker.Type_symbol(t); symbol != nil && len(symbol.Declarations) > 0 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindMappedType {
				mapped := declaration.AsMappedTypeNode()
				readonlyToken := mapped.ReadonlyToken
				if readonlyToken != nil {
					return readonlyToken.Kind != ast.KindMinusToken
				}
				nameType := checker.Checker_getNameTypeFromMappedType(typeChecker, t)
				typeParameter := checker.Checker_getTypeParameterFromMappedType(typeChecker, t)
				if nameType != nil && (typeParameter == nil ||
					!checker.Checker_isTypeAssignableTo(typeChecker, nameType, typeParameter)) {
					return false
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
		if declaration.ModifierFlags()&ast.ModifierFlagsReadonly != 0 ||
			isConstVariableDeclaration(declaration) ||
			isReadonlyAssignmentDeclaration(typeChecker, declaration) {
			return true
		}
	}
	return checker.GetDeclarationModifierFlagsFromSymbol(property)&ast.ModifierFlagsReadonly != 0
}

func isConstVariableDeclaration(declaration *ast.Node) bool {
	return declaration.Kind == ast.KindVariableDeclaration &&
		declaration.Parent != nil &&
		declaration.Parent.Flags&ast.NodeFlagsConstant != 0
}

func isReadonlyAssignmentDeclaration(typeChecker *checker.Checker, declaration *ast.Node) bool {
	if declaration.Kind != ast.KindCallExpression || !ast.IsBindableObjectDefinePropertyCall(declaration) {
		return false
	}
	arguments := declaration.Arguments()
	descriptorType := typeChecker.GetTypeAtLocation(arguments[2])
	if typeChecker.GetPropertyOfType(descriptorType, "value") != nil {
		writableProperty := typeChecker.GetPropertyOfType(descriptorType, "writable")
		if writableProperty == nil {
			return false
		}
		var writableType *checker.Type
		if valueDeclaration := writableProperty.ValueDeclaration; valueDeclaration != nil && valueDeclaration.Kind == ast.KindPropertyAssignment {
			writableType = typeChecker.GetTypeAtLocation(valueDeclaration.Initializer())
		} else {
			writableType = checker.Checker_getTypeOfSymbol(typeChecker, writableProperty)
		}
		return utils.IsFalseLiteralType(writableType)
	}
	return typeChecker.GetPropertyOfType(descriptorType, "set") == nil
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
