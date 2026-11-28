package typeservice

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/microsoft/TypeScript/tsc/shim/locale"
	"github.com/microsoft/TypeScript/tsc/shim/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/shim/printer"
)

// Public TypeScript ObjectFlags have different bit positions from Go's flags.
// Do not expose checker-private cache state through the compatibility surface.
var publicObjectFlags = [...]struct {
	flag checker.ObjectFlags
	name string
}{
	{checker.ObjectFlagsClass, "Class"},
	{checker.ObjectFlagsInterface, "Interface"},
	{checker.ObjectFlagsReference, "Reference"},
	{checker.ObjectFlagsTuple, "Tuple"},
	{checker.ObjectFlagsAnonymous, "Anonymous"},
	{checker.ObjectFlagsMapped, "Mapped"},
	{checker.ObjectFlagsInstantiated, "Instantiated"},
	{checker.ObjectFlagsObjectLiteral, "ObjectLiteral"},
	{checker.ObjectFlagsEvolvingArray, "EvolvingArray"},
	{checker.ObjectFlagsObjectLiteralPatternWithComputedProperties, "ObjectLiteralPatternWithComputedProperties"},
	{checker.ObjectFlagsReverseMapped, "ReverseMapped"},
	{checker.ObjectFlagsJsxAttributes, "JsxAttributes"},
	{checker.ObjectFlagsJSLiteral, "JSLiteral"},
	{checker.ObjectFlagsFreshLiteral, "FreshLiteral"},
	{checker.ObjectFlagsArrayLiteral, "ArrayLiteral"},
	{checker.ObjectFlagsSingleSignatureType, "SingleSignatureType"},
	{checker.ObjectFlagsContainsSpread, "ContainsSpread"},
	{checker.ObjectFlagsObjectRestType, "ObjectRestType"},
	{checker.ObjectFlagsInstantiationExpressionType, "InstantiationExpressionType"},
}

func directRecordKind(op Operation) Operation {
	switch op {
	case "type", "fromTypeNode", "context", "constraint", "apparent", "baseLiteral", "nonNullable",
		"errorType", "anyType", "unknownType", "stringType", "numberType", "booleanType",
		"bigIntType", "eSSymbolType", "nonPrimitiveType", "neverType", "nullType", "undefinedType",
		"trueType", "falseType", "stringLiteralType", "numberLiteralType", "bigIntLiteralType",
		"nullableType", "void",
		"declaredType", "symbolType", "returnType", "numberIndexType", "typeArguments",
		"stringIndexType", "awaited", "widened", "baseTypes", "defaultType", "typeOfSymbol",
		"indexInfo", "indexInfos", "declaredIndexInfos", "typePredicate", "parameterType", "propertyType",
		"mappedFields", "classThisType", "contextualArgument":
		return "typeRecord"
	case "symbol", "declarationSymbol", "aliasedSymbol", "property", "properties", "resolveName",
		"exportSymbol", "immediateAliasedSymbol", "shorthandValueSymbol",
		"exportSpecifierLocalTargetSymbol", "propertySymbolOfDestructuringAssignment",
		"mergedSymbol", "moduleMember", "parameterPropertySymbols":
		return "symbolRecord"
	case "signature", "resolvedSignature", "signatures":
		return "signatureRecord"
	default:
		return ""
	}
}

func (s *Session) directRecords(ctx context.Context, c *checker.Checker, op Operation, result Result) ([]Record, error) {
	kind := directRecordKind(op)
	if kind == "" {
		return nil, nil
	}
	var ids []uint32
	switch value := result.(type) {
	case Handle:
		ids = []uint32{uint32(value)}
	case Handles:
		ids = value
	case NamedSymbols:
		ids = make([]uint32, 0, len(value))
		for _, item := range value {
			ids = append(ids, item.ID)
		}
	case IndexInfos:
		ids = make([]uint32, 0, 2*len(value))
		for _, info := range value {
			ids = append(ids, info.KeyType, info.Type)
		}
	case OptionalIndexInfos:
		if value.Present {
			ids = make([]uint32, 0, 2*len(value.Infos))
			for _, info := range value.Infos {
				ids = append(ids, info.KeyType, info.Type)
			}
		}
	case *TypePredicateFact:
		if value != nil {
			ids = []uint32{value.Type}
		}
	}
	records := []Record{}
	seen := make(map[[2]uint32]bool)
	var add func(Operation, uint32) error
	add = func(kind Operation, id uint32) error {
		if id == 0 {
			return nil
		}
		tag := uint32(1)
		if kind == "symbolRecord" {
			tag = 2
		}
		if kind == "signatureRecord" {
			tag = 3
		}
		key := [2]uint32{tag, id}
		if seen[key] {
			return nil
		}
		seen[key] = true
		var request Request
		switch kind {
		case "typeRecord":
			request = &TypeRequest{Op: kind, ID: id}
		case "symbolRecord":
			request = &SymbolRequest{Op: kind, ID: id}
		case "signatureRecord":
			request = &SignatureRequest{Op: kind, ID: id}
		}
		fact, err := s.execute(ctx, c, request, nil)
		if err != nil {
			return err
		}
		records = append(records, Record{kind, id, fact})
		switch value := fact.(type) {
		case *TypeFact:
			if err := add("symbolRecord", value.Symbol); err != nil {
				return err
			}
			return add("symbolRecord", value.AliasSymbol)
		case *SignatureFact:
			for _, parameter := range value.Parameters {
				if err := add("symbolRecord", parameter); err != nil {
					return err
				}
			}
			return add("symbolRecord", value.ThisParameter)
		}
		return nil
	}
	for _, id := range ids {
		if err := add(kind, id); err != nil {
			return nil, err
		}
	}
	return records, nil
}

func (s *Session) execute(ctx context.Context, c *checker.Checker, request Request, native *ast.Node) (Result, error) {
	switch r := request.(type) {
	case *NodeRequest:
		switch r.Op {
		case "exportSpecifierLocalTargetSymbol":
			if (native.Kind != ast.KindExportSpecifier || native.Parent == nil || native.Parent.Parent == nil) &&
				native.Kind != ast.KindIdentifier {
				return nil, errors.New("expected export specifier or identifier")
			}
			return Handle(s.symbolID(c.GetExportSpecifierLocalTargetSymbol(native))), nil
		case "propertySymbolOfDestructuringAssignment":
			if native.Kind != ast.KindIdentifier || native.Parent == nil || native.Parent.Parent == nil {
				return Handle(0), nil
			}
			return Handle(s.symbolID(c.GetPropertySymbolOfDestructuringAssignment(native))), nil
		case "jsxIntrinsicTagNamesAt":
			return s.namedSymbols(c.GetJsxIntrinsicTagNamesAt(native)), nil
		case "constantValue":
			return constantValue(c.GetConstantValue(native)), nil
		case "isOptionalParameter":
			return Boolean(checker.Checker_isOptionalParameter(c, native)), nil
		case "isImplementationOfOverload":
			return overloadImplementation(c, native), nil
		case "type":
			switch native.Kind {
			case ast.KindJSDocPropertyTag, ast.KindJSDocTypedefTag, ast.KindJSDocCallbackTag:
				if bound := boundJSDocDeclaration(native); bound != nil {
					native = bound
				}
			}
			return Handle(s.typeID(c.GetTypeAtLocation(native))), nil
		case "fromTypeNode":
			return Handle(s.typeID(c.GetTypeFromTypeNode(native))), nil
		case "context":
			if !ast.IsExpression(native) {
				return Handle(0), nil
			}
			return Handle(s.typeID(c.GetContextualType(native, checker.ContextFlagsNone))), nil
		case "declarationSymbol":
			return Handle(s.symbolID(native.Symbol())), nil
		case "symbol":
			if native.Kind == ast.KindSourceFile && !ast.IsExternalModule(native.AsSourceFile()) {
				return Handle(0), nil
			}
			return Handle(s.symbolID(c.GetSymbolAtLocation(native))), nil
		case "signature":
			if native.Kind == ast.KindJSDocSignature {
				if bound := boundJSDocDeclaration(native); bound != nil {
					native = bound
				}
			}
			if !ast.IsFunctionLike(native) {
				return Handle(0), nil
			}
			return Handle(s.signatureID(c.GetSignatureFromDeclaration(native))), nil
		case "shorthandValueSymbol":
			return Handle(s.symbolID(c.GetShorthandAssignmentValueSymbol(native))), nil
		case "suggestionDiagnostics":
			if native.Kind != ast.KindSourceFile {
				return nil, errors.New("expected a source file for suggestion diagnostics")
			}
			return diagnosticFacts(c.GetSuggestionDiagnostics(ctx, native.AsSourceFile())), ctx.Err()
		}
	case *ResolvedSignatureRequest:
		if !ast.IsCallLikeExpression(native) {
			return Handles{0}, nil
		}
		var candidates []*checker.Signature
		var candidatesOut *[]*checker.Signature
		if r.WithCandidates {
			candidatesOut = &candidates
		}
		sig := checker.Checker_getResolvedSignatureWithArgumentCount(
			c, native, candidatesOut, r.ArgumentCount, checker.CheckModeNormal,
		)
		ids := make(Handles, 1, 1+len(candidates))
		ids[0] = s.signatureID(sig)
		for _, candidate := range candidates {
			ids = append(ids, s.signatureID(candidate))
		}
		return ids, nil
	case *NamedNodeRequest:
		switch r.Op {
		case "validPropertyAccess":
			switch native.Kind {
			case ast.KindPropertyAccessExpression, ast.KindQualifiedName, ast.KindImportType:
				return Boolean(c.IsValidPropertyAccess(native, string(r.Name))), nil
			default:
				return nil, errors.New("expected property access, qualified name, or import type")
			}
		case "parameterPropertySymbols":
			if !ast.IsParameterPropertyDeclaration(native, native.Parent) ||
				native.Parent == nil || native.Parent.Parent == nil ||
				native.Parent.Locals() == nil || native.Parent.Parent.Symbol() == nil {
				return nil, errors.New("parameter is not a parameter property declaration")
			}
			param, prop := c.GetSymbolsOfParameterPropertyDeclaration(native, string(r.Name))
			ids := Handles{}
			if param != nil {
				ids = append(ids, s.symbolID(param))
			}
			if prop != nil {
				ids = append(ids, s.symbolID(prop))
			}
			return ids, nil
		}
	case *ContextualArgumentRequest:
		if !ast.IsCallLikeExpression(native) {
			return Handle(0), nil
		}
		return Handle(s.typeID(c.GetContextualTypeForArgumentAtIndex(native, r.Index))), nil
	case *TypeRequest:
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		switch r.Op {
		case "augmentedPropertiesOfType":
			return s.namedSymbols(checker.Checker_getAugmentedPropertiesOfType(c, t)), nil
		case "constraint":
			return Handle(s.typeID(c.GetBaseConstraintOfType(t))), nil
		case "apparent":
			return Handle(s.typeID(c.GetApparentType(t))), nil
		case "baseLiteral":
			return Handle(s.typeID(c.GetBaseTypeOfLiteralType(t))), nil
		case "nonNullable":
			return Handle(s.typeID(c.GetNonNullableType(t))), nil
		case "awaited":
			return Handle(s.typeID(checker.Checker_getAwaitedType(c, t))), nil
		case "widened":
			return Handle(s.typeID(c.GetWidenedType(t))), nil
		case "baseTypes":
			if t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
				t = t.Target()
			}
			return s.typeIDs(c.GetBaseTypes(t)), nil
		case "defaultType":
			return Handle(s.typeID(c.GetDefaultFromTypeParameter(t))), nil
		case "declaredIndexInfos":
			return s.declaredIndexInfos(t), nil
		case "typeDetails":
			return s.typeDetails(t), nil
		case "classThisType":
			iface := t.AsInterfaceType()
			if iface == nil {
				return Handle(0), nil
			}
			return Handle(s.typeID(checker.InterfaceType_thisType(iface))), nil
		case "mappedFields":
			if t.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
				return Handles(nil), nil
			}
			mapped := t.AsMappedType()
			return Handles{
				s.typeID(checker.MappedType_typeParameter(mapped)),
				s.typeID(checker.MappedType_constraintType(mapped)),
				s.typeID(checker.MappedType_templateType(mapped)),
				s.typeID(checker.MappedType_modifiersType(mapped)),
			}, nil
		case "properties":
			return s.namedSymbols(c.GetPropertiesOfType(t)), nil
		case "typeArguments":
			return s.typeIDs(c.GetTypeArguments(t)), nil
		case "isArray":
			return Boolean(c.IsArrayType(t)), nil
		case "isArrayLike":
			return Boolean(c.IsArrayLikeType(t)), nil
		case "isTuple":
			return Boolean(checker.IsTupleType(t)), nil
		case "numberIndexType":
			return Handle(s.typeID(c.GetNumberIndexType(t))), nil
		case "stringIndexType":
			return Handle(s.typeID(c.GetStringIndexType(t))), nil
		case "indexInfos":
			infos := c.GetIndexInfosOfType(t)
			result := make(IndexInfos, len(infos))
			for i, info := range infos {
				result[i] = s.indexInfo(info)
			}
			return result, nil
		case "typeRecord":
			return s.typeFact(t)
		}
	case *SymbolRequest:
		sym, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		switch r.Op {
		case "rootSymbols":
			return s.namedSymbols(c.GetRootSymbols(sym)), nil
		case "mergedSymbol":
			return Handle(s.symbolID(c.GetMergedSymbol(sym))), nil
		case "isArgumentsSymbol":
			return Boolean(c.GetArgumentsSymbol() == sym), nil
		case "isUndefinedSymbol":
			return Boolean(c.GetUndefinedSymbol() == sym), nil
		case "aliasedSymbol":
			return Handle(s.symbolID(c.GetAliasedSymbol(sym))), nil
		case "immediateAliasedSymbol":
			return Handle(s.symbolID(c.GetImmediateAliasedSymbol(sym))), nil
		case "isUnknownSymbol":
			return Boolean(c.IsUnknownSymbol(sym)), nil
		case "exportsOfModule":
			exports := c.GetExportsOfModule(sym)
			slices.SortFunc(exports, c.CompareSymbols)
			return s.namedSymbols(exports), nil
		case "fullyQualifiedName":
			return Text(c.GetFullyQualifiedName(sym)), nil
		case "symbolMembers", "symbolExports":
			return s.symbolTable(c, sym, r.Op == "symbolExports"), nil
		case "exportSymbol":
			return Handle(s.symbolID(c.GetExportSymbolOfSymbol(sym))), nil
		case "declaredType":
			return Handle(s.typeID(c.GetDeclaredTypeOfSymbol(sym))), nil
		case "typeOfSymbol":
			return Handle(s.typeID(c.GetTypeOfSymbol(sym))), nil
		case "symbolRecord":
			declarations, value := symbolDeclarations(sym, true)
			return &SymbolFact{
				Name:        symbolName(sym),
				EscapedName: ast.EscapeSymbolName(sym.Name),
				Flags:       uint32(sym.Flags),
				CheckFlags:  jsCheckFlags(sym.CheckFlags),
				nodes:       declarations,
				valueNode:   value,
			}, nil
		}
	case *SymbolDocsRequest:
		if !r.WithChecker {
			c = nil
		}
		symbol, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		active := map[*ast.Symbol]bool{symbol: true}
		declarations, _ := symbolDeclarations(symbol, false)
		if r.Op == "symbolDocumentation" {
			return documentation(c, declarations, active), nil
		}
		return jsDocTags(c, declarations, active), nil
	case *SignatureRequest:
		sig, err := s.signatureByID(r.ID)
		if err != nil {
			return nil, err
		}
		switch r.Op {
		case "returnType":
			return Handle(s.typeID(c.GetReturnTypeOfSignature(sig))), nil
		case "signatureDocumentation", "signatureJSDocTags":
			declarations := []*ast.Node{sig.Declaration()}
			active := make(map[*ast.Symbol]bool)
			if r.Op == "signatureDocumentation" {
				return documentation(c, declarations, active), nil
			}
			return jsDocTags(c, declarations, active), nil
		case "typePredicate":
			predicate := c.GetTypePredicateOfSignature(sig)
			if predicate == nil {
				return (*TypePredicateFact)(nil), nil
			}
			fact := &TypePredicateFact{Kind: uint32(predicate.Kind()), Type: s.typeID(predicate.Type())}
			if predicate.Kind() == checker.TypePredicateKindIdentifier || predicate.Kind() == checker.TypePredicateKindAssertsIdentifier {
				name := predicate.ParameterName()
				fact.ParameterName = &name
				fact.ParameterIndex = predicate.ParameterIndex()
			}
			return fact, nil
		case "signatureRecord":
			return &SignatureFact{
				Parameters:       s.symbolIDs(sig.Parameters()),
				TypeParameters:   s.typeIDs(sig.TypeParameters()),
				ThisParameter:    s.symbolID(sig.ThisParameter()),
				MinArgumentCount: sig.MinArgumentCount(),
				node:             sig.Declaration(),
			}, nil
		}
	case *CheckerRequest:
		switch r.Op {
		case "errorType":
			return Handle(s.typeID(c.GetErrorType())), nil
		case "anyType":
			return Handle(s.typeID(c.GetAnyType())), nil
		case "unknownType":
			return Handle(s.typeID(c.GetUnknownType())), nil
		case "stringType":
			return Handle(s.typeID(c.GetStringType())), nil
		case "numberType":
			return Handle(s.typeID(c.GetNumberType())), nil
		case "booleanType":
			return Handle(s.typeID(c.GetBooleanType())), nil
		case "bigIntType":
			return Handle(s.typeID(c.GetBigIntType())), nil
		case "eSSymbolType":
			return Handle(s.typeID(c.GetESSymbolType())), nil
		case "nonPrimitiveType":
			return Handle(s.typeID(c.GetNonPrimitiveType())), nil
		case "neverType":
			return Handle(s.typeID(c.GetNeverType())), nil
		case "nullType":
			return Handle(s.typeID(c.GetNullType())), nil
		case "undefinedType":
			return Handle(s.typeID(c.GetUndefinedType())), nil
		case "trueType":
			return Handle(s.typeID(c.GetBooleanType().Types()[1])), nil
		case "falseType":
			return Handle(s.typeID(c.GetBooleanType().Types()[0])), nil
		case "void":
			return Handle(s.typeID(c.GetVoidType())), nil
		case "ambientModules":
			modules := slices.Clone(c.GetAmbientModules())
			slices.SortFunc(modules, c.CompareSymbols)
			return s.namedSymbols(modules), nil
		}
	case *PropertyRequest:
		switch r.Op {
		case "moduleMember":
			sym, err := s.symbolByID(r.ID)
			if err != nil {
				return nil, err
			}
			return Handle(s.symbolID(c.TryGetMemberInModuleExports(string(r.Name), sym))), nil
		case "property":
			t, err := s.typeByID(r.ID)
			if err != nil {
				return nil, err
			}
			return Handle(s.symbolID(c.GetPropertyOfType(t, string(r.Name)))), nil
		case "propertyType":
			t, err := s.typeByID(r.ID)
			if err != nil {
				return nil, err
			}
			return Handle(s.typeID(c.GetTypeOfPropertyOfType(t, string(r.Name)))), nil
		}
	case *SymbolTypeRequest:
		sym, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(c.GetTypeOfSymbolAtLocation(sym, native))), nil
	case *ResolveNameRequest:
		return Handle(s.symbolID(resolveName(c, string(r.Name), native, ast.SymbolFlags(r.Kind), r.ExcludeGlobals))), nil
	case *AssignableRequest:
		source, err := s.typeByID(r.Source)
		if err != nil {
			return nil, err
		}
		target, err := s.typeByID(r.Target)
		if err != nil {
			return nil, err
		}
		return Boolean(c.IsTypeAssignableTo(source, target)), nil
	case *IndexInfoRequest:
		if r.Kind > 1 {
			return nil, errors.New("invalid index kind")
		}
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		key := c.GetNumberType()
		if r.Kind == 0 {
			key = c.GetStringType()
		}
		info := c.GetIndexInfoOfType(t, key)
		if info == nil {
			return IndexInfos{}, nil
		}
		return IndexInfos{s.indexInfo(info)}, nil
	case *ScopeSymbolsRequest:
		return s.namedSymbols(symbolsInScope(c, native, ast.SymbolFlags(r.Meaning))), nil
	case *ParameterTypeRequest:
		sig, err := s.signatureByID(r.ID)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(c.GetTypeParameterAtPosition(sig, r.Position))), nil
	case *StringLiteralRequest:
		return Handle(s.typeID(checker.Checker_getStringLiteralType(c, string(r.Value)))), nil
	case *NumberLiteralRequest:
		value := r.Value
		switch r.NonFinite {
		case "":
		case "NaN":
			value = math.NaN()
		case "Infinity":
			value = math.Inf(1)
		case "-Infinity":
			value = math.Inf(-1)
		default:
			return nil, errors.New("invalid non-finite number literal")
		}
		return Handle(s.typeID(checker.Checker_getNumberLiteralType(c, jsnum.Number(value)))), nil
	case *BigIntLiteralRequest:
		text := r.Base10Value
		if text == "" {
			text = "0"
		}
		negative := r.Negative && text != "0"
		return Handle(s.typeID(checker.Checker_getBigIntLiteralType(c, jsnum.PseudoBigInt{Negative: negative, Base10Value: text}))), nil
	case *NullableTypeRequest:
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		return Handle(s.typeID(checker.Checker_getNullableType(c, t, checker.TypeFlags(r.Flags)))), nil
	case *TypePredicateStringRequest:
		if r.Kind > uint32(checker.TypePredicateKindAssertsIdentifier) {
			return nil, errors.New("invalid type predicate kind")
		}
		var name string
		if r.ParameterName != nil {
			name = string(*r.ParameterName)
		}
		var predType *checker.Type
		if r.Type != 0 {
			var err error
			predType, err = s.typeByID(r.Type)
			if err != nil {
				return nil, err
			}
		}
		pred := checker.Checker_newTypePredicate(c, checker.TypePredicateKind(r.Kind), name, r.ParameterIndex, predType)
		flags := checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		}
		return Text(checker.Checker_typePredicateToStringEx(c, pred, native, flags)), nil
	case *PrintRequest:
		return s.executePrint(c, r, native)
	case *SignaturesRequest:
		if r.Kind > 1 {
			return nil, errors.New("invalid signature kind")
		}
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		ids := Handles{}
		for _, v := range c.GetSignaturesOfType(t, checker.SignatureKind(r.Kind)) {
			ids = append(ids, s.signatureID(v))
		}
		return ids, nil
	}
	return nil, fmt.Errorf("unsupported type service operation %q", request.operation())
}

func (s *Session) symbolTable(c *checker.Checker, sym *ast.Symbol, exports bool) OptionalNamedSymbols {
	table := sym.Members
	expectedFlags := ast.SymbolFlagsClass | ast.SymbolFlagsInterface | ast.SymbolFlagsTypeLiteral | ast.SymbolFlagsObjectLiteral
	if exports {
		table = sym.Exports
		expectedFlags = ast.SymbolFlagsClass | ast.SymbolFlagsEnum | ast.SymbolFlagsModule | ast.SymbolFlagsVariable
	}
	present := table != nil
	if !present && len(sym.Declarations) != 0 {
		// TypeScript's binder creates empty member/export maps for container symbols.
		bound := sym.Declarations[0].Symbol()
		if (bound == sym || c.GetMergedSymbol(bound) == sym) && sym.Flags&expectedFlags != 0 {
			present = true
		}
	}
	if !present {
		return OptionalNamedSymbols{}
	}
	members := slices.Collect(maps.Values(table))
	slices.SortFunc(members, c.CompareSymbols)
	result := make(NamedSymbols, len(members))
	for i, member := range members {
		result[i] = NamedSymbol{ID: s.symbolID(member), Name: ast.EscapeSymbolName(member.Name)}
	}
	return OptionalNamedSymbols{Present: true, Symbols: result}
}

func (s *Session) typeFact(t *checker.Type) (*TypeFact, error) {
	flags := t.Flags()
	fact := &TypeFact{
		Flags:              checker.FormatTypeFlags(flags),
		Symbol:             s.symbolID(t.Symbol()),
		AliasSymbol:        s.symbolID(t.Alias().Symbol()),
		AliasTypeArguments: s.typeIDs(t.Alias().TypeArguments()),
	}
	if flags&checker.TypeFlagsIntrinsic != 0 {
		fact.IntrinsicName = t.AsIntrinsicType().IntrinsicName()
	} else if flags&checker.TypeFlagsBoolean != 0 {
		fact.IntrinsicName = "boolean"
	}
	if flags&checker.TypeFlagsUniqueESSymbol != 0 {
		fact.EscapedName = ast.EscapeSymbolName(checker.GetPropertyNameFromType(t))
	}
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsTypeParameter) != 0 {
		fact.Target = s.typeID(t.Target())
	}
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsUnionOrIntersection|checker.TypeFlagsIntrinsic|checker.TypeFlagsBooleanLiteral) != 0 {
		fact.ObjectFlags = []string{}
		// Public ObjectFlags bits (including 22..25, which tsgo overloads on
		// unions/intersections for internal flags such as ContainsIntersections)
		// are only meaningful on TypeFlagsObject.
		if flags&checker.TypeFlagsObject != 0 {
			objectFlags := t.ObjectFlags()
			for _, flag := range publicObjectFlags {
				if objectFlags&flag.flag != 0 {
					fact.ObjectFlags = append(fact.ObjectFlags, flag.name)
				}
			}
		}
	}
	if flags&checker.TypeFlagsObject != 0 {
		if definition := t.AsInterfaceType(); definition != nil {
			if parameters := definition.TypeParameters(); len(parameters) > 0 {
				fact.TypeParameters = s.typeIDs(parameters)
			}
			if t.ObjectFlags()&checker.ObjectFlagsTuple != 0 {
				tuple := t.AsTupleType()
				fact.Tuple = &TupleFact{
					Readonly:      tuple.IsReadonly(),
					CombinedFlags: uint32(checker.TupleType_combinedFlags(tuple)),
				}
			}
		}
	}
	switch {
	case flags&(checker.TypeFlagsUnionOrIntersection|checker.TypeFlagsTemplateLiteral) != 0:
		fact.TypeOperands = s.typeIDs(t.Types())
	case flags&(checker.TypeFlagsIndex|checker.TypeFlagsStringMapping) != 0:
		fact.TypeOperands = []uint32{s.typeID(t.Target())}
	case flags&checker.TypeFlagsIndexedAccess != 0:
		indexed := t.AsIndexedAccessType()
		fact.TypeOperands = []uint32{s.typeID(indexed.ObjectType()), s.typeID(indexed.IndexType())}
	case flags&checker.TypeFlagsConditional != 0:
		conditional := t.AsConditionalType()
		fact.TypeOperands = []uint32{s.typeID(conditional.CheckType()), s.typeID(conditional.ExtendsType())}
	}
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		fact.IntrinsicName = fmt.Sprint(t.AsLiteralType().Value())
	}
	if flags&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) != 0 {
		switch value := t.AsLiteralType().Value().(type) {
		case string:
			fact.Literal = Literal{Kind: 1, Text: value}
		case jsnum.Number:
			fact.Literal = Literal{Kind: 2, Number: float64(value)}
		case jsnum.PseudoBigInt:
			text := value.Base10Value
			if text == "" {
				text = "0"
			}
			fact.Literal = Literal{Kind: 3, Text: text, Negative: value.Negative}
		case nil: // Computed enum members have no literal value.
		default:
			return nil, fmt.Errorf("unsupported compiler literal %T", value)
		}
	}
	return fact, nil
}

func (s *Session) executePrint(c *checker.Checker, r *PrintRequest, native *ast.Node) (Result, error) {
	switch r.Op {
	case "symbolString":
		sym, err := s.symbolByID(r.ID)
		if err != nil {
			return nil, err
		}
		flags := checker.SymbolFormatFlagsAllowAnyNodeKind
		if r.Flags != nil {
			flags = checker.SymbolFormatFlags(*r.Flags)
		}
		return Text(c.SymbolToStringEx(sym, native, ast.SymbolFlags(r.Kind), flags)), nil
	case "typeNode":
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		var builderFlags nodebuilder.Flags
		if r.Flags != nil {
			builderFlags = nodebuilder.Flags(*r.Flags)
		}
		builder := checker.NewNodeBuilder(c, printer.NewEmitContext())
		node := builder.TypeToTypeNode(t, native, builderFlags, nodebuilder.InternalFlagsNone, nil)
		return encodeGeneratedNode(node, builder, native)
	case "typeString":
		t, err := s.typeByID(r.ID)
		if err != nil {
			return nil, err
		}
		flags := checker.TypeFormatFlagsAllowUniqueESSymbolType | checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		}
		return Text(sanitizeTruncatedUTF8(c.TypeToStringEx(t, native, flags, nil))), nil
	case "signatureString":
		if r.Kind > uint32(checker.SignatureKindConstruct) {
			return nil, errors.New("invalid signature kind")
		}
		sig, err := s.signatureByID(r.ID)
		if err != nil {
			return nil, err
		}
		flags := checker.TypeFormatFlagsNone
		if r.Flags != nil {
			flags = checker.TypeFormatFlags(*r.Flags)
		}
		if r.Kind == uint32(checker.SignatureKindCall) {
			flags |= checker.TypeFormatFlagsWriteCallStyleSignature
			return Text(c.SignatureToStringEx(sig, native, flags, nil)), nil
		}
		// TypeScript's signatureToString lets callers format any signature with
		// SignatureKind.Construct; tsgo's SignatureToStringEx only inspects sig.flags.
		sigOutput := ast.KindConstructSignature
		if flags&checker.TypeFormatFlagsWriteArrowStyleSignature != 0 {
			sigOutput = ast.KindConstructorType
		}
		emitCtx := printer.NewEmitContext()
		builder := checker.NewNodeBuilder(c, emitCtx)
		combinedFlags := nodebuilder.Flags(flags&checker.TypeFormatFlagsNodeBuilderFlagsMask) |
			nodebuilder.FlagsIgnoreErrors |
			nodebuilder.FlagsWriteTypeParametersInQualifiedName
		node := builder.SignatureToSignatureDeclaration(sig, sigOutput, native, combinedFlags, nodebuilder.InternalFlagsNone, nil)
		p := printer.NewPrinter(printer.PrinterOptions{
			RemoveComments:        true,
			OmitTrailingSemicolon: true,
			NeverAsciiEscape:      true,
		}, printer.PrintHandlers{}, emitCtx)
		var sf *ast.SourceFile
		if native != nil {
			sf = ast.GetSourceFileOfNode(native)
		}
		if flags&checker.TypeFormatFlagsMultilineObjectLiterals != 0 {
			w := printer.NewTextWriter("\n", 0)
			p.Write(node, sf, w, nil)
			return Text(w.String()), nil
		}
		w, put := printer.GetSingleLineStringWriter()
		defer put()
		p.Write(node, sf, w, nil)
		return Text(w.String()), nil
	default:
		return nil, fmt.Errorf("unsupported print operation %q", r.Op)
	}
}

func sanitizeTruncatedUTF8(s string) string {
	const ellipsis = "..."
	if !strings.HasSuffix(s, ellipsis) || utf8.ValidString(s) {
		return s
	}
	prefix := s[:len(s)-len(ellipsis)]
	for len(prefix) > 0 && !utf8.ValidString(prefix) {
		prefix = prefix[:len(prefix)-1]
	}
	return prefix + ellipsis
}

func overloadImplementation(c *checker.Checker, node *ast.Node) Boolean {
	if !ast.IsFunctionLike(node) || !ast.NodeIsPresent(node.Body()) ||
		ast.IsGetAccessorDeclaration(node) || ast.IsSetAccessorDeclaration(node) {
		return false
	}
	signatures := checker.Checker_getSignaturesOfSymbol(c, checker.Checker_getSymbolOfDeclaration(c, node))
	if len(signatures) > 1 {
		return true
	}
	if len(signatures) == 1 {
		decl := signatures[0].Declaration()
		return Boolean(decl != nil && decl != node && decl.Flags&ast.NodeFlagsJSDoc == 0)
	}
	return false
}

func constantValue(value any) Literal {
	switch v := value.(type) {
	case string:
		return Literal{Kind: 1, Text: v}
	case jsnum.Number:
		return Literal{Kind: 2, Number: float64(v)}
	default:
		return Literal{}
	}
}

// Go inserts three write-access bits before ContainsStatic and has private
// cache flags above Unresolved. Neither group exists in JS CheckFlags.
func jsCheckFlags(flags ast.CheckFlags) uint32 {
	const low = ast.CheckFlagsContainsWritePublic - 1
	const high = (ast.CheckFlagsUnresolved*2 - 1) &^ (ast.CheckFlagsContainsStatic - 1)
	return uint32(flags&low | (flags&high)>>3)
}

func diagnosticFacts(diagnostics []*ast.Diagnostic) SuggestionDiagnostics {
	if len(diagnostics) == 0 {
		return nil
	}
	result := make(SuggestionDiagnostics, len(diagnostics))
	for i, diagnostic := range diagnostics {
		d := DiagnosticFact{
			Code:        uint32(diagnostic.Code()),
			Category:    uint32(diagnostic.Category()),
			Message:     diagnostic.Localize(locale.Default),
			Source:      diagnostic.Source(),
			Deprecated:  diagnostic.ReportsDeprecated(),
			Unnecessary: diagnostic.ReportsUnnecessary(),
			Chain:       diagnosticFacts(diagnostic.MessageChain()),
			Related:     diagnosticFacts(diagnostic.RelatedInformation()),
		}
		if f := diagnostic.File(); f != nil {
			d.File = f.FileName()
			d.FilePath = string(f.Path())
			positions := f.GetPositionMap()
			start := positions.UTF8ToUTF16(diagnostic.Pos())
			d.Start = uint32(start)
			d.Length = uint32(positions.UTF8ToUTF16(diagnostic.End()) - start)
		}
		result[i] = d
	}
	return result
}
