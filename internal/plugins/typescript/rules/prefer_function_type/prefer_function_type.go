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

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	sourceText := ctx.SourceFile.Text()
	lineStarts := ctx.SourceFile.ECMALineMap()

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

	shouldWrapSuggestion := func(parent *ast.Node) bool {
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindUnionType, ast.KindIntersectionType, ast.KindArrayType:
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

	// buildInterfaceHeader returns the verbatim `Name<TypeParams>` text used to
	// prefix the rewritten type alias. Includes the angle brackets when type
	// parameters exist; constraints / defaults / variance markers are picked up
	// for free because we slice raw source.
	buildInterfaceHeader := func(iface *ast.InterfaceDeclaration) string {
		nameText := utils.TrimmedNodeText(ctx.SourceFile, iface.Name())
		if iface.TypeParameters == nil || len(iface.TypeParameters.Nodes) == 0 {
			return nameText
		}
		params := iface.TypeParameters.Nodes
		lastParam := params[len(params)-1]
		lastParamEnd := utils.TrimNodeTextRange(ctx.SourceFile, lastParam).End()
		gtPos := lastParamEnd
		for gtPos < len(sourceText) && sourceText[gtPos] != '>' {
			gtPos++
		}
		if gtPos < len(sourceText) {
			gtPos++
		}
		nameStart := utils.TrimNodeTextRange(ctx.SourceFile, iface.Name()).Pos()
		return sourceText[nameStart:gtPos]
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

		memberRange := utils.TrimNodeTextRange(ctx.SourceFile, member)
		memberStart := memberRange.Pos()
		memberEnd := memberRange.End()
		text := sourceText[memberStart:memberEnd]

		// Find the ':' before the return type. The return type node starts at
		// the type token itself (whitespace skipped); the ':' immediately
		// precedes it (modulo whitespace).
		returnTypeStart := utils.TrimNodeTextRange(ctx.SourceFile, returnType).Pos()
		colonPos := returnTypeStart - 1
		for colonPos >= memberStart && sourceText[colonPos] != ':' {
			colonPos--
		}
		if colonPos < memberStart {
			ctx.ReportNode(member, msg)
			return
		}
		colonOffset := colonPos - memberStart

		suggestion := text[:colonOffset] + " =>" + text[colonOffset+1:]
		lastChar := ""
		if strings.HasSuffix(suggestion, ";") {
			lastChar = ";"
			suggestion = suggestion[:len(suggestion)-1]
		}

		if shouldWrapSuggestion(node.Parent) {
			suggestion = "(" + suggestion + ")"
		}

		if isInterface {
			suggestion = "type " + buildInterfaceHeader(node.AsInterfaceDeclaration()) + " = " + suggestion + lastChar
		}

		nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)

		// Collect leading and trailing comments by scanning the trivia regions
		// directly. tsgo's scanner.GetLeadingCommentRanges only collects line-
		// boundary-adjacent comments, so it misses same-line inline blocks like
		// `{ /* c */ (): void }` that the upstream rule needs to relocate.
		// Trailing scan is capped at the containing node's end so comments
		// outside the type-literal / interface body never leak in.
		comments := append(
			collectCommentsForward(sourceText, member.Pos(), memberStart),
			collectCommentsForward(sourceText, memberEnd, nodeRange.End())...,
		)

		memberLine := scanner.ComputeLineOfPosition(lineStarts, memberStart)

		var fixes []rule.RuleFix

		// `export interface Foo` has comments moved before `export` so they
		// don't end up between `export` and the resulting `type` declaration.
		if isInterface && hasExport && !hasDefault {
			var commentsText strings.Builder
			for _, c := range comments {
				commentsText.WriteString(sourceText[c.Pos():c.End()])
				commentsText.WriteByte('\n')
			}
			if commentsText.Len() > 0 {
				fixes = append(fixes, rule.RuleFix{
					Text:  commentsText.String(),
					Range: core.NewTextRange(nodeRange.Pos(), nodeRange.Pos()),
				})
			}
		} else {
			for _, c := range comments {
				cText := sourceText[c.Pos():c.End()]
				if scanner.ComputeLineOfPosition(lineStarts, c.Pos()) == memberLine {
					cText += " "
				} else {
					cText += "\n"
				}
				suggestion = cText + suggestion
			}
		}

		// For interfaces, only replace from the `interface` keyword onwards.
		// Preserving the modifier prefix verbatim avoids dropping any trivia
		// (e.g. comments) that lives between modifiers, exactly matching
		// upstream's `replaceTextRange([declaration.range[0], ...])` semantics
		// — ESTree's `declaration.range[0]` is the `interface` keyword, not
		// the first modifier.
		replaceStart := nodeRange.Pos()
		if isInterface {
			modifiers := node.AsInterfaceDeclaration().Modifiers()
			var scanFrom int
			if modifiers != nil && len(modifiers.Nodes) > 0 {
				scanFrom = modifiers.Nodes[len(modifiers.Nodes)-1].End()
			} else {
				scanFrom = nodeRange.Pos()
			}
			s := scanner.GetScannerForSourceFile(ctx.SourceFile, scanFrom)
			for s.TokenStart() < nodeRange.End() {
				if s.Token() == ast.KindInterfaceKeyword {
					replaceStart = s.TokenStart()
					break
				}
				s.Scan()
			}
		}

		fixes = append(fixes, rule.RuleFix{
			Text:  suggestion,
			Range: core.NewTextRange(replaceStart, nodeRange.End()),
		})

		ctx.ReportNodeWithFixes(member, msg, fixes...)
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
