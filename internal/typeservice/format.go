package typeservice

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/nodebuilder"
	"github.com/microsoft/TypeScript/tsc/shim/printer"
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
	options := printer.PrinterOptions{RemoveComments: typ != checker.Checker_unresolvedType(c)}
	text := printDisplay(node, enclosing, flags, options, builder.EmitContext(), writer)
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
	options := printer.PrinterOptions{RemoveComments: true, OmitTrailingSemicolon: true}
	return printDisplay(node, enclosing, flags, options, builder.EmitContext(), writer)
}

// The native printer may reuse the source spelling of a quoted literal.
// JavaScript's display API quotes the decoded value on the emitted syntax node.
type displayWriter struct {
	printer.EmitTextWriter
	context      *printer.EmitContext
	literal      *ast.Node
	singleQuotes bool
	bindingDepth int
}

func printDisplay(node, enclosing *ast.Node, flags checker.TypeFormatFlags, options printer.PrinterOptions, context *printer.EmitContext, writer printer.EmitTextWriter) string {
	w := &displayWriter{EmitTextWriter: writer, context: context, singleQuotes: flags&checker.TypeFormatFlagsUseSingleQuotesForStringLiteralType != 0}
	handlers := printer.PrintHandlers{
		OnBeforeEmitNode: func(n *ast.Node) {
			if n != nil && n.Kind == ast.KindStringLiteral {
				w.literal = n
			} else if n != nil && ast.IsBindingPattern(n) {
				w.bindingDepth++
			}
		},
		OnAfterEmitNode: func(n *ast.Node) {
			if n != nil && n.Kind == ast.KindStringLiteral {
				w.literal = nil
			} else if n != nil && ast.IsBindingPattern(n) {
				w.bindingDepth--
			}
		},
	}
	p := printer.NewPrinter(options, handlers, context)
	p.Write(node, ast.GetSourceFileOfNode(enclosing), w, nil)
	return writer.String()
}

func (w *displayWriter) WriteStringLiteral(text string) {
	if n := w.literal; n != nil {
		quote := printer.QuoteCharDoubleQuote
		if w.bindingDepth == 0 && w.singleQuotes && n.AsStringLiteral().TokenFlags&ast.TokenFlagsSingleQuote != 0 {
			quote = printer.QuoteCharSingleQuote
		}
		escape := printer.EscapeNonAsciiString
		if w.bindingDepth != 0 || (w.context.EmitFlags(n)&printer.EFNoAsciiEscaping != 0 && !reusedDisplayLiteral(w.context.ParseNode(n))) {
			escape = printer.EscapeString
		}
		text = string(quote) + escape(n.Text(), quote) + string(quote)
	}
	w.EmitTextWriter.WriteStringLiteral(text)
}

// TypeScript reuses the spelling policy of annotations and expressions that
// directly supply a return type. Literals inferred inside a composite type
// are synthesized with NoAsciiEscaping.
func reusedDisplayLiteral(source *ast.Node) bool {
	if source == nil || source.Parent == nil {
		return false
	}
	parent := source.Parent
	if ast.IsLiteralTypeNode(parent) || ast.IsTypeElement(parent) {
		return true
	}
	for parent != nil && ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) && parent.Expression() == source {
		source, parent = parent, parent.Parent
	}
	return parent != nil && (ast.IsReturnStatement(parent) || ast.IsArrowFunction(parent) && parent.Body() == source)
}
