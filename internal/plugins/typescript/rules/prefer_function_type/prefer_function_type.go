package prefer_function_type

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func phrase(kind ast.Kind) string {
	switch kind {
	case ast.KindInterfaceDeclaration:
		return "Interface"
	case ast.KindTypeLiteral:
		return "Type literal"
	}
	return ""
}

func functionTypeOverCallableTypeMessage(literalOrInterface string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "functionTypeOverCallableType",
		Description: literalOrInterface + " only has a call signature, you should use a function type instead.",
		Data: map[string]string{
			"literalOrInterface": literalOrInterface,
		},
	}
}

func unexpectedThisMessage(interfaceName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unexpectedThisOnFunctionOnlyInterface",
		Description: "`this` refers to the function type '" + interfaceName +
			"', did you intend to use a generic `this` parameter like `<Self>(this: Self, ...) => Self` instead?",
		Data: map[string]string{
			"interfaceName": interfaceName,
		},
	}
}

var PreferFunctionTypeRule = rule.CreateRule(rule.Rule{
	Name: "prefer-function-type",
	Run:  run,
})

// collectCommentsForward walks `text` starting at `start`, skipping whitespace
// and collecting every `//` or `/* ... */` comment it sees. It stops at the
// first non-trivia character or at `limit`. Used in place of
// scanner.GetLeadingCommentRanges because that helper only collects comments
// adjacent to line breaks, missing same-line inline blocks like
// `{ /* c */ (): void }` that the rule needs to relocate.
func collectCommentsForward(text string, start, limit int) []core.TextRange {
	var result []core.TextRange
	pos := start
	for pos < limit {
		ch := text[pos]
		switch ch {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			pos++
			continue
		}
		if pos+1 < limit && text[pos] == '/' && text[pos+1] == '/' {
			cStart := pos
			pos += 2
			for pos < limit && text[pos] != '\n' && text[pos] != '\r' {
				pos++
			}
			result = append(result, core.NewTextRange(cStart, pos))
			continue
		}
		if pos+1 < limit && text[pos] == '/' && text[pos+1] == '*' {
			cStart := pos
			pos += 2
			for pos < limit {
				if pos+1 < limit && text[pos] == '*' && text[pos+1] == '/' {
					pos += 2
					break
				}
				pos++
			}
			result = append(result, core.NewTextRange(cStart, pos))
			continue
		}
		return result
	}
	return result
}

type preferFunctionTypeFixer struct {
	sourceFile *ast.SourceFile
	sourceText string
	lineStarts []core.TextPos
}

func shouldWrapFunctionTypeSuggestion(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindUnionType, ast.KindIntersectionType, ast.KindArrayType:
		return true
	}
	return false
}

// interfaceHeader returns the verbatim Name<TypeParams> text used to prefix
// the rewritten type alias. Constraints, defaults, and variance markers are
// preserved by slicing the source.
func (f preferFunctionTypeFixer) interfaceHeader(
	declaration *ast.InterfaceDeclaration,
) string {
	nameText := utils.TrimmedNodeText(f.sourceFile, declaration.Name())
	if declaration.TypeParameters == nil || len(declaration.TypeParameters.Nodes) == 0 {
		return nameText
	}
	parameters := declaration.TypeParameters.Nodes
	lastParameter := parameters[len(parameters)-1]
	lastParameterEnd := utils.TrimNodeTextRange(f.sourceFile, lastParameter).End()
	greaterThanPosition := lastParameterEnd
	for greaterThanPosition < len(f.sourceText) && f.sourceText[greaterThanPosition] != '>' {
		greaterThanPosition++
	}
	if greaterThanPosition < len(f.sourceText) {
		greaterThanPosition++
	}
	nameStart := utils.TrimNodeTextRange(f.sourceFile, declaration.Name()).Pos()
	return f.sourceText[nameStart:greaterThanPosition]
}

func (f preferFunctionTypeFixer) fixes(member *ast.Node, node *ast.Node) []rule.RuleFix {
	var returnType *ast.Node
	switch member.Kind {
	case ast.KindCallSignature:
		returnType = member.AsCallSignatureDeclaration().Type
	case ast.KindConstructSignature:
		returnType = member.AsConstructSignatureDeclaration().Type
	}
	if returnType == nil {
		return nil
	}

	isInterface := node.Kind == ast.KindInterfaceDeclaration
	hasExport := isInterface && ast.HasSyntacticModifier(node, ast.ModifierFlagsExport)

	memberRange := utils.TrimNodeTextRange(f.sourceFile, member)
	memberStart := memberRange.Pos()
	memberEnd := memberRange.End()
	text := f.sourceText[memberStart:memberEnd]

	returnTypeStart := utils.TrimNodeTextRange(f.sourceFile, returnType).Pos()
	colonPosition := returnTypeStart - 1
	for colonPosition >= memberStart && f.sourceText[colonPosition] != ':' {
		colonPosition--
	}
	if colonPosition < memberStart {
		return nil
	}
	colonOffset := colonPosition - memberStart

	suggestion := text[:colonOffset] + " =>" + text[colonOffset+1:]
	lastCharacter := ""
	if strings.HasSuffix(suggestion, ";") {
		lastCharacter = ";"
		suggestion = suggestion[:len(suggestion)-1]
	}

	if shouldWrapFunctionTypeSuggestion(node.Parent) {
		suggestion = "(" + suggestion + ")"
	}

	if isInterface {
		declaration := node.AsInterfaceDeclaration()
		suggestion = "type " + f.interfaceHeader(declaration) + " = " + suggestion + lastCharacter
	}

	nodeRange := utils.TrimNodeTextRange(f.sourceFile, node)
	comments := append(
		collectCommentsForward(f.sourceText, member.Pos(), memberStart),
		collectCommentsForward(f.sourceText, memberEnd, nodeRange.End())...,
	)
	memberLine := scanner.ComputeLineOfPosition(f.lineStarts, memberStart)

	var fixes []rule.RuleFix
	if isInterface && hasExport {
		var commentsText strings.Builder
		for _, comment := range comments {
			commentsText.WriteString(f.sourceText[comment.Pos():comment.End()])
			commentsText.WriteByte('\n')
		}
		if commentsText.Len() > 0 {
			fixes = append(fixes, rule.RuleFix{
				Text:  commentsText.String(),
				Range: core.NewTextRange(nodeRange.Pos(), nodeRange.Pos()),
			})
		}
	} else {
		for _, comment := range comments {
			commentText := f.sourceText[comment.Pos():comment.End()]
			if scanner.ComputeLineOfPosition(f.lineStarts, comment.Pos()) == memberLine {
				commentText += " "
			} else {
				commentText += "\n"
			}
			suggestion = commentText + suggestion
		}
	}

	replaceStart := nodeRange.Pos()
	if isInterface {
		modifiers := node.AsInterfaceDeclaration().Modifiers()
		var scanFrom int
		if modifiers != nil && len(modifiers.Nodes) > 0 {
			scanFrom = modifiers.Nodes[len(modifiers.Nodes)-1].End()
		} else {
			scanFrom = nodeRange.Pos()
		}
		sourceScanner := scanner.GetScannerForSourceFile(f.sourceFile, scanFrom)
		for sourceScanner.TokenStart() < nodeRange.End() {
			if sourceScanner.Token() == ast.KindInterfaceKeyword {
				replaceStart = sourceScanner.TokenStart()
				break
			}
			sourceScanner.Scan()
		}
	}

	fixes = append(fixes, rule.RuleFix{
		Text:  suggestion,
		Range: core.NewTextRange(replaceStart, nodeRange.End()),
	})
	return fixes
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	fixer := preferFunctionTypeFixer{
		sourceFile: ctx.SourceFile,
		sourceText: ctx.SourceFile.Text(),
		lineStarts: ctx.SourceFile.ECMALineMap(),
	}

	// hasOneNonFunctionSupertype mirrors upstream's hasOneSupertype(): an
	// interface with `extends X[, Y, ...]` is skipped unless the only supertype
	// is literally the `Function` identifier.
	hasOneNonFunctionSupertype := func(iface *ast.InterfaceDeclaration) bool {
		if iface.HeritageClauses == nil {
			return false
		}
		var extends []*ast.Node
		for _, clause := range iface.HeritageClauses.Nodes {
			hc := clause.AsHeritageClause()
			if hc == nil || hc.Token != ast.KindExtendsKeyword || hc.Types == nil {
				continue
			}
			extends = append(extends, hc.Types.Nodes...)
		}
		if len(extends) == 0 {
			return false
		}
		if len(extends) != 1 {
			return true
		}
		expr := extends[0].AsExpressionWithTypeArguments().Expression
		if expr == nil || expr.Kind != ast.KindIdentifier || expr.AsIdentifier().Text != "Function" {
			return true
		}
		return false
	}

	// collectThisTypes walks the interface for `this` type references that are
	// not nested inside a TypeLiteral. Mirrors upstream's
	// 'TSInterfaceDeclaration TSThisType' visitor with literalNesting tracking.
	collectThisTypes := func(root *ast.Node) []*ast.Node {
		var result []*ast.Node
		nesting := 0
		var visit func(n *ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if n.Kind == ast.KindTypeLiteral {
				nesting++
				n.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
				nesting--
				return
			}
			if n.Kind == ast.KindThisType {
				if nesting == 0 {
					result = append(result, n)
				}
				return
			}
			n.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return false
			})
		}
		visit(root)
		return result
	}

	checkMember := func(member *ast.Node, node *ast.Node, tsThisTypes []*ast.Node) {
		var returnType *ast.Node
		switch member.Kind {
		case ast.KindCallSignature:
			returnType = member.AsCallSignatureDeclaration().Type
		case ast.KindConstructSignature:
			returnType = member.AsConstructSignatureDeclaration().Type
		default:
			return
		}
		if returnType == nil {
			return
		}

		isInterface := node.Kind == ast.KindInterfaceDeclaration

		if isInterface && len(tsThisTypes) > 0 {
			interfaceName := node.AsInterfaceDeclaration().Name().AsIdentifier().Text
			ctx.ReportNode(tsThisTypes[0], unexpectedThisMessage(interfaceName))
			return
		}

		msg := functionTypeOverCallableTypeMessage(phrase(node.Kind))

		hasExport := isInterface && ast.HasSyntacticModifier(node, ast.ModifierFlagsExport)
		hasDefault := isInterface && ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault)

		// Upstream marks `export default interface ...` as non-fixable because
		// the fix would need to split it into a type alias plus a separate
		// `export default` statement.
		if hasExport && hasDefault {
			ctx.ReportNode(member, msg)
			return
		}

		ctx.ReportNodeWithDeferredFixes(member, msg, func() []rule.RuleFix {
			return fixer.fixes(member, node)
		})
	}

	return rule.RuleListeners{
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			iface := node.AsInterfaceDeclaration()
			if hasOneNonFunctionSupertype(iface) {
				return
			}
			if iface.Members == nil || len(iface.Members.Nodes) != 1 {
				return
			}
			member := iface.Members.Nodes[0]
			tsThisTypes := collectThisTypes(member)
			checkMember(member, node, tsThisTypes)
		},
		ast.KindTypeLiteral: func(node *ast.Node) {
			typeLit := node.AsTypeLiteralNode()
			if typeLit.Members == nil || len(typeLit.Members.Nodes) != 1 {
				return
			}
			checkMember(typeLit.Members.Nodes[0], node, nil)
		},
	}
}
