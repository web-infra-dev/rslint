package typeservice

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/shim/printer"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// TypeScript's default writer uses an empty newline, even for multiline object
// literals. Its final display limit counts UTF-16 code units.
func typeToString(c *checker.Checker, typ *checker.Type, enclosing *ast.Node, flags checker.TypeFormatFlags, noErrorTruncation bool) string {
	limit := 320
	builderFlags := nodebuilder.Flags(flags&checker.TypeFormatFlagsNodeBuilderFlagsMask) | nodebuilder.FlagsIgnoreErrors
	if noErrorTruncation || flags&checker.TypeFormatFlagsNoTruncation != 0 {
		limit = 2_000_000
		builderFlags |= nodebuilder.FlagsNoTruncation
	}
	builder, release := checker.Checker_getNodeBuilder(c)
	defer release()
	// Building a node can report diagnostics that recursively format types.
	// Count this outer call in the native checker's recursion guard.
	level := checker.Checker_serializationLevelPtr(c)
	*level++
	defer func() { *level-- }()
	node := builder.TypeToTypeNode(typ, enclosing, builderFlags, nodebuilder.InternalFlagsNone, nil)
	writer := printer.NewTextWriter("", 0)
	p := printer.NewPrinter(printer.PrinterOptions{RemoveComments: typ != checker.Checker_unresolvedType(c)}, printer.PrintHandlers{}, builder.EmitContext())
	p.Write(node, ast.GetSourceFileOfNode(enclosing), writer, nil)
	text := writer.String()
	if len(text) >= limit {
		if units := ecmascript.StringCodeUnits(text); len(units) >= limit {
			return ecmascript.StringFromCodeUnits(units[:limit-3]) + "..."
		}
	}
	return text
}

// TypeScript uses a single-line writer for signatures; the requested kind
// determines whether it prints a call or a constructor.
func signatureToString(c *checker.Checker, signature *checker.Signature, enclosing *ast.Node, flags checker.TypeFormatFlags, kind checker.SignatureKind) string {
	nodeKind, arrowKind := ast.KindCallSignature, ast.KindFunctionType
	if kind == checker.SignatureKindConstruct {
		nodeKind, arrowKind = ast.KindConstructSignature, ast.KindConstructorType
	}
	if flags&checker.TypeFormatFlagsWriteArrowStyleSignature != 0 {
		nodeKind = arrowKind
	}
	builder, release := checker.Checker_getNodeBuilder(c)
	defer release()
	builderFlags := nodebuilder.Flags(flags&checker.TypeFormatFlagsNodeBuilderFlagsMask) | nodebuilder.FlagsIgnoreErrors | nodebuilder.FlagsWriteTypeParametersInQualifiedName
	node := builder.SignatureToSignatureDeclaration(signature, nodeKind, enclosing, builderFlags, nodebuilder.InternalFlagsNone, nil)
	writer, releaseWriter := printer.GetSingleLineStringWriter()
	defer releaseWriter()
	textWriter := &signatureTextWriter{
		EmitTextWriter: writer,
		useSingleQuote: flags&checker.TypeFormatFlagsUseSingleQuotesForStringLiteralType != 0,
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
			quote, escape := printer.QuoteCharDoubleQuote, printer.EscapeNonAsciiString
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
