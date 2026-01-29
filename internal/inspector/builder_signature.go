package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// BuildShallowSignatureInfo builds minimal SignatureInfo for nested signatures
func (b *Builder) BuildShallowSignatureInfo(sig *checker.Signature) *SignatureInfo {
	if sig == nil {
		return nil
	}

	// Get flags
	flags := checker.Signature_flags(sig)

	info := &SignatureInfo{
		Flags:     uint32(flags),
		FlagNames: GetSignatureFlagNames(flags),
	}

	// Calculate min argument count
	params := checker.Signature_parameters(sig)
	minArgCount := 0
	for i, param := range params {
		isRest := flags&checker.SignatureFlagsHasRestParameter != 0 && i == len(params)-1
		isOptional := param.Flags&ast.SymbolFlagsOptional != 0
		if !isRest && !isOptional {
			minArgCount++
		}
	}
	info.MinArgumentCount = minArgCount

	// Get position from declaration for on-demand fetch
	if decl := checker.Signature_declaration(sig); decl != nil {
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

	return info
}

// BuildSignatureInfo builds SignatureInfo from a Signature
func (b *Builder) BuildSignatureInfo(sig *checker.Signature) *SignatureInfo {
	if sig == nil {
		return nil
	}

	// Get flags
	flags := checker.Signature_flags(sig)

	// Get parameters
	params := checker.Signature_parameters(sig)

	// Calculate minimum argument count from parameters
	// Count required parameters (non-optional, non-rest)
	minArgCount := 0
	for i, param := range params {
		isRest := flags&checker.SignatureFlagsHasRestParameter != 0 && i == len(params)-1
		isOptional := param.Flags&ast.SymbolFlagsOptional != 0
		if !isRest && !isOptional {
			minArgCount++
		}
	}

	info := &SignatureInfo{
		Flags:            uint32(flags),
		FlagNames:        GetSignatureFlagNames(flags),
		MinArgumentCount: minArgCount,
	}

	// Position for on-demand fetch
	if decl := checker.Signature_declaration(sig); decl != nil {
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

	// Parameters as symbols (full for detailed info)
	if len(params) > 0 {
		info.Parameters = make([]*SymbolInfo, 0, len(params))
		for _, param := range params {
			info.Parameters = append(info.Parameters, b.BuildSymbolInfo(param))
		}
	}

	// This parameter as symbol (full for detailed info)
	if thisSym := checker.Signature_thisParameter(sig); thisSym != nil {
		info.ThisParameter = b.BuildSymbolInfo(thisSym)
	}

	// Type parameters as full types
	typeParams := checker.Signature_typeParameters(sig)
	if len(typeParams) > 0 {
		info.TypeParameters = make([]*TypeInfo, 0, len(typeParams))
		for _, tp := range typeParams {
			info.TypeParameters = append(info.TypeParameters, b.BuildTypeInfo(tp))
		}
	}

	// Return type (full type info)
	returnType := b.checker.GetReturnTypeOfSignature(sig)
	if returnType != nil {
		info.ReturnType = b.BuildTypeInfo(returnType)
	}

	// Type predicate (with full type)
	if pred := b.checker.GetTypePredicateOfSignature(sig); pred != nil {
		info.TypePredicate = b.buildTypePredicateInfo(pred)
	}

	// Declaration (shallow node info for lazy loading)
	if decl := checker.Signature_declaration(sig); decl != nil {
		info.Declaration = b.BuildShallowNodeInfo(decl)
	}

	return info
}

// buildTypePredicateInfo builds TypePredicateInfo with full type info
func (b *Builder) buildTypePredicateInfo(pred *checker.TypePredicate) *TypePredicateInfo {
	if pred == nil {
		return nil
	}

	// Get kind and convert to string with checker. prefix
	kind := checker.TypePredicate_kind(pred)
	var kindName string
	switch kind {
	case checker.TypePredicateKindThis:
		kindName = "checker.TypePredicateKindThis"
	case checker.TypePredicateKindIdentifier:
		kindName = "checker.TypePredicateKindIdentifier"
	case checker.TypePredicateKindAssertsThis:
		kindName = "checker.TypePredicateKindAssertsThis"
	case checker.TypePredicateKindAssertsIdentifier:
		kindName = "checker.TypePredicateKindAssertsIdentifier"
	default:
		kindName = "Unknown"
	}

	info := &TypePredicateInfo{
		Kind:           int(kind),
		KindName:       kindName,
		ParameterName:  checker.TypePredicate_parameterName(pred),
		ParameterIndex: int(checker.TypePredicate_parameterIndex(pred)),
	}

	// Predicate type (full)
	if t := pred.Type(); t != nil {
		info.Type = b.BuildTypeInfo(t)
	}

	return info
}
