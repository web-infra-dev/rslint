package checker

import (
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

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
