package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// BuildShallowTypeInfo builds minimal TypeInfo for nested types
// Note: ID is not included in shallow info as it changes on each request
func (b *Builder) BuildShallowTypeInfo(t *checker.Type) *TypeInfo {
	if t == nil {
		return nil
	}
	info := &TypeInfo{
		Flags:      uint32(t.Flags()),
		FlagNames:  GetTypeFlagNames(t.Flags()),
		TypeString: b.checker.TypeToString(t),
	}
	// Get position from symbol's declaration for fetching full info later
	if symbol := t.Symbol(); symbol != nil && !IsInternalSymbol(symbol) {
		var decl *ast.Node
		if symbol.ValueDeclaration != nil {
			decl = symbol.ValueDeclaration
		} else if len(symbol.Declarations) > 0 {
			decl = symbol.Declarations[0]
		}
		if decl != nil {
			if name := decl.Name(); name != nil {
				info.Pos = b.GetTokenPos(name.AsNode())
			} else {
				info.Pos = b.GetTokenPos(decl)
			}
			// Check if declaration is from external file
			declFile := ast.GetSourceFileOfNode(decl)
			if declFile != nil && declFile != b.sourceFile {
				info.FileName = declFile.FileName()
			}
		}
	}
	return info
}

// BuildTypeInfo builds TypeInfo from a Type
func (b *Builder) BuildTypeInfo(t *checker.Type) *TypeInfo {
	if t == nil {
		return nil
	}

	info := &TypeInfo{
		Id:         uint32(t.Id()),
		Flags:      uint32(t.Flags()),
		FlagNames:  GetTypeFlagNames(t.Flags()),
		TypeString: b.checker.TypeToString(t),
	}

	// Get position from symbol's declaration for fetching full info later
	if symbol := t.Symbol(); symbol != nil {
		var decl *ast.Node
		if symbol.ValueDeclaration != nil {
			decl = symbol.ValueDeclaration
		} else if len(symbol.Declarations) > 0 {
			decl = symbol.Declarations[0]
		}
		if decl != nil {
			if name := decl.Name(); name != nil {
				info.Pos = b.GetTokenPos(name.AsNode())
			} else {
				info.Pos = b.GetTokenPos(decl)
			}
			// Check if declaration is from external file
			declFile := ast.GetSourceFileOfNode(decl)
			if declFile != nil && declFile != b.sourceFile {
				info.FileName = declFile.FileName()
			}
		}
	}

	// Object flags
	if t.Flags()&checker.TypeFlagsObject != 0 {
		info.ObjectFlags = uint32(t.ObjectFlags())
		info.ObjectFlagNames = GetObjectFlagNames(t.ObjectFlags())
	}

	// Intrinsic name for primitive types
	if t.Flags()&checker.TypeFlagsIntrinsic != 0 {
		if intrinsic := t.AsIntrinsicType(); intrinsic != nil {
			info.IntrinsicName = intrinsic.IntrinsicName()
		}
	}

	// Literal type properties (value, freshType, regularType)
	// Note: freshType and regularType are literal types without symbols, so they can't be lazy loaded
	// We use buildLiteralTypeShallow to get info with id but without recursion
	if t.Flags()&checker.TypeFlagsLiteral != 0 {
		if literal := t.AsLiteralType(); literal != nil {
			info.Value = checker.LiteralType_value(literal)
			// Build shallow info for freshType/regularType to avoid infinite recursion
			// but include id to indicate it's complete data
			if freshType := checker.LiteralType_freshType(literal); freshType != nil {
				info.FreshType = b.buildLiteralTypeShallow(freshType)
			}
			if regularType := checker.LiteralType_regularType(literal); regularType != nil {
				info.RegularType = b.buildLiteralTypeShallow(regularType)
			}
		}
	}

	// Symbol (full for detailed info)
	if symbol := t.Symbol(); symbol != nil {
		info.Symbol = b.BuildSymbolInfo(symbol)
	}

	// Alias symbol (full for detailed info)
	if alias := checker.Type_alias(t); alias != nil {
		if aliasSymbol := alias.Symbol(); aliasSymbol != nil {
			info.AliasSymbol = b.BuildSymbolInfo(aliasSymbol)
		}
	}

	// Type arguments (full for detailed info - these are often primitives without valid positions)
	if t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		typeArgs := b.checker.GetTypeArguments(t)
		if len(typeArgs) > 0 {
			info.TypeArguments = make([]*TypeInfo, 0, len(typeArgs))
			for _, arg := range typeArgs {
				info.TypeArguments = append(info.TypeArguments, b.BuildTypeInfo(arg))
			}
		}
	}

	// Base types (full for detailed info)
	if t.Flags()&checker.TypeFlagsObject != 0 {
		objFlags := t.ObjectFlags()
		if objFlags&checker.ObjectFlagsClassOrInterface != 0 {
			baseTypes := b.checker.GetBaseTypes(t)
			if len(baseTypes) > 0 {
				info.BaseTypes = make([]*TypeInfo, 0, len(baseTypes))
				for _, base := range baseTypes {
					info.BaseTypes = append(info.BaseTypes, b.BuildTypeInfo(base))
				}
			}
		}
	}

	// Union/Intersection types - use full info since members are typically leaf types
	// and intrinsic types (string, number, etc.) don't have positions to fetch from
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		var types []*checker.Type
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			if union := t.AsUnionType(); union != nil {
				types = union.Types()
			}
		} else if t.Flags()&checker.TypeFlagsIntersection != 0 {
			if inter := t.AsIntersectionType(); inter != nil {
				types = inter.Types()
			}
		}
		if len(types) > 0 {
			info.Types = make([]*TypeInfo, 0, len(types))
			for _, member := range types {
				// Use full type info to avoid fetch issues with intrinsic types
				info.Types = append(info.Types, b.BuildTypeInfo(member))
			}
		}
	}

	// Properties (full symbols for detailed info)
	if t.Flags()&checker.TypeFlagsObject != 0 {
		props := b.checker.GetPropertiesOfType(t)
		if len(props) > 0 {
			info.Properties = make([]*SymbolInfo, 0, len(props))
			for _, prop := range props {
				info.Properties = append(info.Properties, b.BuildSymbolInfo(prop))
			}
		}
	}

	// Call signatures (full info)
	callSigs := b.checker.GetSignaturesOfType(t, checker.SignatureKindCall)
	if len(callSigs) > 0 {
		info.CallSignatures = make([]*SignatureInfo, 0, len(callSigs))
		for _, sig := range callSigs {
			info.CallSignatures = append(info.CallSignatures, b.BuildSignatureInfo(sig))
		}
	}

	// Construct signatures (full info)
	constructSigs := b.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)
	if len(constructSigs) > 0 {
		info.ConstructSignatures = make([]*SignatureInfo, 0, len(constructSigs))
		for _, sig := range constructSigs {
			info.ConstructSignatures = append(info.ConstructSignatures, b.BuildSignatureInfo(sig))
		}
	}

	// Index signatures (full types for detailed info)
	if t.Flags()&checker.TypeFlagsObject != 0 {
		indexInfos := checker.Checker_getIndexInfosOfType(b.checker, t)
		if len(indexInfos) > 0 {
			info.IndexInfos = make([]*IndexInfoType, 0, len(indexInfos))
			for _, ii := range indexInfos {
				indexInfo := &IndexInfoType{
					IsReadonly: checker.IndexInfo_isReadonly(ii),
				}
				if keyType := checker.IndexInfo_keyType(ii); keyType != nil {
					indexInfo.KeyType = b.BuildTypeInfo(keyType)
				}
				if valueType := checker.IndexInfo_valueType(ii); valueType != nil {
					indexInfo.ValueType = b.BuildTypeInfo(valueType)
				}
				info.IndexInfos = append(info.IndexInfos, indexInfo)
			}
		}
	}

	// Type parameter constraint (full for detailed info)
	if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := b.checker.GetConstraintOfTypeParameter(t)
		if constraint != nil {
			info.Constraint = b.BuildTypeInfo(constraint)
		}
	}

	// Target type for type references, type parameters, index types, etc. (full for detailed info)
	// Target() returns the underlying type for generic references like Array<string> -> Array<T>
	target := b.SafeGetTarget(t)
	if target != nil && target != t {
		info.Target = b.BuildTypeInfo(target)
	}

	return info
}

// buildLiteralTypeShallow builds TypeInfo for literal types with id (to indicate complete data)
// but without recursing into freshType/regularType to avoid infinite loops
func (b *Builder) buildLiteralTypeShallow(t *checker.Type) *TypeInfo {
	if t == nil {
		return nil
	}
	info := &TypeInfo{
		Id:         uint32(t.Id()),
		Flags:      uint32(t.Flags()),
		FlagNames:  GetTypeFlagNames(t.Flags()),
		TypeString: b.checker.TypeToString(t),
	}

	// Include value for literal types
	if t.Flags()&checker.TypeFlagsLiteral != 0 {
		if literal := t.AsLiteralType(); literal != nil {
			info.Value = checker.LiteralType_value(literal)
		}
	}

	return info
}
