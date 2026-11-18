package checker

import (
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

type PluginNodeBuilderFlags = nodebuilder.Flags

// CachedSymbolType reads the same transient-symbol slot exposed as links.type
// by the JavaScript checker. Merely inspecting it must not resolve a type.
func CachedSymbolType(c *Checker, symbol *ast.Symbol) *Type {
	links := (*extra_Checker)(unsafe.Pointer(c)).valueSymbolLinks.store.TryGet(uint64(ast.GetSymbolId(symbol)))
	if links == nil || *links == nil {
		return nil
	}
	return ValueSymbolLinks_resolvedType(*links)
}

// TypeNodeForPlugin accepts the JavaScript NodeBuilderFlags bitmask.
func TypeNodeForPlugin(c *Checker, typ *Type, enclosing *ast.Node, flags uint32) *ast.Node {
	return c.TypeToTypeNode(typ, enclosing, nodebuilder.Flags(flags), nil)
}

// ResolveSignatureWithOptions uses the same overload mode and cached
// resolution as TypeScript's getResolvedSignature API.
func ResolveSignatureWithOptions(c *Checker, node *ast.Node, candidates *[]*Signature, argumentCount *int) *Signature {
	checker := (*extra_Checker)(unsafe.Pointer(c))
	savedCount := checker.apparentArgumentCount
	checker.apparentArgumentCount = argumentCount
	defer func() { checker.apparentArgumentCount = savedCount }()
	return Checker_getResolvedSignature(c, node, candidates, CheckModeNormal)
}

// TypeToStringWithJavaScriptLayout uses the compiler emitter with the empty
// newline and four-space indentation of the JavaScript API's default writer.
// The emitter still decides where lines and indentation occur.
func TypeToStringWithJavaScriptLayout(c *Checker, typ *Type, enclosing *ast.Node, flags TypeFormatFlags) string {
	checker := (*extra_Checker)(unsafe.Pointer(c))
	const maxSerializationLevel = 2
	if checker.serializationLevel >= maxSerializationLevel {
		return "?"
	}
	noTruncation := checker.compilerOptions.NoErrorTruncation == core.TSTrue || flags&TypeFormatFlagsNoTruncation != 0
	builderFlags := nodebuilder.Flags(flags&TypeFormatFlagsNodeBuilderFlagsMask) | nodebuilder.FlagsIgnoreErrors
	if noTruncation {
		builderFlags |= nodebuilder.FlagsNoTruncation
	}
	builder, release := Checker_getNodeBuilder(c)
	defer release()
	checker.serializationLevel++
	defer func() { checker.serializationLevel-- }()
	node := builder.TypeToTypeNode(typ, enclosing, builderFlags, nodebuilder.InternalFlagsNone, nil)
	writer := printer.NewTextWriter("", 0)
	p := printer.NewPrinter(printer.PrinterOptions{RemoveComments: typ != checker.unresolvedType}, printer.PrintHandlers{}, builder.EmitContext())
	p.Write(node, ast.GetSourceFileOfNode(enclosing), writer, nil)
	return writer.String()
}

// SignatureToStringWithJavaScriptLayout keeps the JavaScript API's single-line
// writer even when the builder requests multiline type literals.
func SignatureToStringWithJavaScriptLayout(c *Checker, signature *Signature, enclosing *ast.Node, flags TypeFormatFlags, kind SignatureKind) string {
	var nodeKind ast.Kind
	if flags&TypeFormatFlagsWriteArrowStyleSignature != 0 {
		nodeKind = core.IfElse(kind == SignatureKindConstruct, ast.KindConstructorType, ast.KindFunctionType)
	} else {
		nodeKind = core.IfElse(kind == SignatureKindConstruct, ast.KindConstructSignature, ast.KindCallSignature)
	}
	builder, release := Checker_getNodeBuilder(c)
	defer release()
	builderFlags := nodebuilder.Flags(flags&TypeFormatFlagsNodeBuilderFlagsMask) | nodebuilder.FlagsIgnoreErrors | nodebuilder.FlagsWriteTypeParametersInQualifiedName
	node := builder.SignatureToSignatureDeclaration(signature, nodeKind, enclosing, builderFlags, nodebuilder.InternalFlagsNone, nil)
	writer, releaseWriter := printer.GetSingleLineStringWriter()
	defer releaseWriter()
	textWriter := &signatureTextWriter{
		EmitTextWriter: writer,
		useSingleQuote: flags&TypeFormatFlagsUseSingleQuotesForStringLiteralType != 0,
	}
	handlers := printer.PrintHandlers{
		OnBeforeEmitNode: func(n *ast.Node) {
			if n != nil && ast.IsBindingPattern(n) {
				textWriter.bindingDepth++
			}
		},
		OnAfterEmitNode: func(n *ast.Node) {
			if n != nil && ast.IsBindingPattern(n) {
				textWriter.bindingDepth--
			}
		},
	}
	p := printer.NewPrinter(printer.PrinterOptions{RemoveComments: true, OmitTrailingSemicolon: true}, handlers, builder.EmitContext())
	p.Write(node, ast.GetSourceFileOfNode(enclosing), textWriter, nil)
	return writer.String()
}

//go:linkname escapeSignatureString github.com/microsoft/TypeScript/tsc/internal/printer.escapeNonAsciiString
func escapeSignatureString(value string, quote printer.QuoteChar) string

// The printer can reuse the source spelling of a literal. Escape its decoded
// value so a source escape cannot change meaning in the formatted signature.
type signatureTextWriter struct {
	printer.EmitTextWriter
	useSingleQuote bool
	bindingDepth   int
}

func (w *signatureTextWriter) WriteStringLiteral(text string) {
	if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
		s := scanner.NewScanner()
		s.SetText(text)
		if s.Scan() == ast.KindStringLiteral && s.TokenEnd() == len(text) && s.TokenFlags()&ast.TokenFlagsUnterminated == 0 {
			quote, escape := printer.QuoteCharDoubleQuote, escapeSignatureString
			if w.bindingDepth != 0 {
				// The JS printer keeps native characters in parameter names.
				escape = printer.EscapeString
			} else if w.useSingleQuote {
				quote = printer.QuoteChar(text[0])
			}
			text = string(quote) + escape(s.TokenValue(), quote) + string(quote)
		}
	}
	w.EmitTextWriter.WriteStringLiteral(text)
}

// ResolvedSignatureTypeArguments applies the instantiation mapper of this call.
func ResolvedSignatureTypeArguments(c *Checker, sig *Signature) []*Type {
	mapper := Signature_mapper(sig)
	if mapper == nil {
		return nil
	}
	target := sig.Target()
	if target == nil {
		target = sig
	}
	return Checker_instantiateTypes(c, target.TypeParameters(), mapper)
}

func NewPluginIndexInfo(key, value *Type, readonly bool, declaration *ast.Node) *IndexInfo {
	return (*IndexInfo)(unsafe.Pointer(&extra_IndexInfo{keyType: key, valueType: value, isReadonly: readonly, declaration: declaration}))
}

func PluginPredicateString(c *Checker, kind TypePredicateKind, name string, index int32, typ *Type, enclosing *ast.Node, flags TypeFormatFlags) string {
	predicate := &extra_TypePredicate{kind: kind, parameterName: name, parameterIndex: index, t: typ}
	return Checker_typePredicateToStringEx(c, (*TypePredicate)(unsafe.Pointer(predicate)), enclosing, flags)
}

// PluginRelationCacheSizes reads the four tables while the caller owns the checker.
func PluginRelationCacheSizes(c *Checker) [4]uint32 {
	ch := (*extra_Checker)(unsafe.Pointer(c))
	return [4]uint32{
		uint32(len(Relation_results(ch.assignableRelation))),
		uint32(len(Relation_results(ch.identityRelation))),
		uint32(len(Relation_results(ch.subtypeRelation))),
		uint32(len(Relation_results(ch.strictSubtypeRelation))),
	}
}

// CachedIndexedAccess reads the simplification without triggering a new one.
func CachedIndexedAccess(c *Checker, t *Type, writing bool) *Type {
	kind := CachedTypeKindIndexedAccessForReading
	if writing {
		kind = CachedTypeKindIndexedAccessForWriting
	}
	k := extra_CachedTypeKey{kind: kind, typeId: t.Id()}
	return Checker_cachedTypes(c)[*(*CachedTypeKey)(unsafe.Pointer(&k))]
}
