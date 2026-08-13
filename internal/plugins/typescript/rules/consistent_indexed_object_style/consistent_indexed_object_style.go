package consistent_indexed_object_style

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_indexed_object_style.schema.json
var schemaJSON []byte

type ConsistentIndexedObjectStyleOptions struct {
	Style string `json:"style"`
}

var (
	preferRecordMessage = rule.RuleMessage{
		Id:          "preferRecord",
		Description: "A record is preferred over an index signature.",
	}
	preferRecordSuggestionMessage = rule.RuleMessage{
		Id:          "preferRecordSuggestion",
		Description: "Change into a record instead of an index signature.",
	}
	preferIndexSignatureMessage = rule.RuleMessage{
		Id:          "preferIndexSignature",
		Description: "An index signature is preferred over a record.",
	}
	preferIndexSignatureSuggestionMessage = rule.RuleMessage{
		Id:          "preferIndexSignatureSuggestion",
		Description: "Change into an index signature instead of a record.",
	}
)

// ConsistentIndexedObjectStyleRule requires a consistent indexed-object syntax.
var ConsistentIndexedObjectStyleRule = rule.CreateRule(rule.Rule{
	Name:   "consistent-indexed-object-style",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func parseOptions(options []any) ConsistentIndexedObjectStyleOptions {
	opts := ConsistentIndexedObjectStyleOptions{Style: "record"}
	if len(options) > 0 {
		if style, ok := options[0].(string); ok {
			opts.Style = style
		}
	}
	return opts
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	if parseOptions(options).Style == "index-signature" {
		return rule.RuleListeners{
			ast.KindTypeReference: func(node *ast.Node) {
				reportRecordAsIndexSignature(&ctx, node)
			},
		}
	}

	circularReferences := circularReferenceAnalyzer{refs: ctx.Refs}
	return rule.RuleListeners{
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			declaration := node.AsInterfaceDeclaration()
			if declaration == nil || !hasSingleIndexSignatureMember(declaration.Members) {
				return
			}
			indexSignature := declaration.Members.Nodes[0]
			keyType, valueType, ok := indexSignatureTypes(indexSignature)
			if ok {
				reportInterfaceAsRecord(&ctx, &circularReferences, node, declaration, indexSignature, keyType, valueType)
			}
		},
		ast.KindMappedType: func(node *ast.Node) {
			mappedType := node.AsMappedTypeNode()
			if mappedType == nil || mappedType.TypeParameter == nil {
				return
			}
			typeParameter := mappedType.TypeParameter.AsTypeParameterDeclaration()
			if typeParameter == nil || typeParameter.Name() == nil || typeParameter.Name().Kind != ast.KindIdentifier ||
				typeParameter.Constraint == nil {
				return
			}
			constraint := typeParameter.Constraint
			if constraint.Kind == ast.KindTypeOperator {
				typeOperator := constraint.AsTypeOperatorNode()
				if typeOperator != nil && typeOperator.Operator == ast.KindKeyOfKeyword {
					return
				}
			}
			reportMappedTypeAsRecord(&ctx, &circularReferences, node, mappedType, typeParameter, constraint)
		},
		ast.KindTypeLiteral: func(node *ast.Node) {
			typeLiteral := node.AsTypeLiteralNode()
			if typeLiteral == nil || !hasSingleIndexSignatureMember(typeLiteral.Members) {
				return
			}
			indexSignature := typeLiteral.Members.Nodes[0]
			keyType, valueType, ok := indexSignatureTypes(indexSignature)
			if ok {
				reportTypeLiteralAsRecord(&ctx, &circularReferences, node, indexSignature, keyType, valueType)
			}
		},
	}
}

func hasSingleIndexSignatureMember(members *ast.NodeList) bool {
	return members != nil && len(members.Nodes) == 1 && members.Nodes[0] != nil &&
		members.Nodes[0].Kind == ast.KindIndexSignature
}

func reportInterfaceAsRecord(
	ctx *rule.RuleContext,
	circularReferences *circularReferenceAnalyzer,
	node *ast.Node,
	declaration *ast.InterfaceDeclaration,
	indexSignature *ast.Node,
	keyType *ast.Node,
	valueType *ast.Node,
) {
	if circularReferences.referencesDeclaration(node, valueType) {
		return
	}

	reportRange := interfaceDeclarationRange(ctx.SourceFile, node)
	safeFix := (declaration.HeritageClauses == nil || len(declaration.HeritageClauses.Nodes) == 0) &&
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault)
	reportIndexSignatureAsRecord(
		ctx,
		reportRange,
		indexSignature,
		keyType,
		valueType,
		declaration,
		";",
		safeFix,
	)
}

func reportTypeLiteralAsRecord(
	ctx *rule.RuleContext,
	circularReferences *circularReferenceAnalyzer,
	node *ast.Node,
	indexSignature *ast.Node,
	keyType *ast.Node,
	valueType *ast.Node,
) {
	if canDeeplyReferenceType(valueType.Kind) {
		if parentDeclaration := findParentTypeAlias(node); parentDeclaration != nil &&
			circularReferences.referencesDeclaration(parentDeclaration, valueType) {
			return
		}
	}

	reportIndexSignatureAsRecord(
		ctx,
		bracedTypeRange(ctx.SourceFile, node),
		indexSignature,
		keyType,
		valueType,
		nil,
		"",
		true,
	)
}

func indexSignatureTypes(indexSignature *ast.Node) (
	keyType *ast.Node,
	valueType *ast.Node,
	ok bool,
) {
	declaration := indexSignature.AsIndexSignatureDeclaration()
	if declaration == nil || declaration.Parameters == nil || len(declaration.Parameters.Nodes) == 0 || declaration.Type == nil {
		return nil, nil, false
	}
	parameter := declaration.Parameters.Nodes[0].AsParameterDeclaration()
	if parameter == nil || parameter.DotDotDotToken != nil || parameter.Name() == nil ||
		parameter.Name().Kind != ast.KindIdentifier || parameter.Type == nil {
		return nil, nil, false
	}
	return parameter.Type, declaration.Type, true
}

func reportIndexSignatureAsRecord(
	ctx *rule.RuleContext,
	reportRange core.TextRange,
	indexSignature *ast.Node,
	keyType *ast.Node,
	valueType *ast.Node,
	interfaceDeclaration *ast.InterfaceDeclaration,
	postfix string,
	safeFix bool,
) {
	if !safeFix {
		ctx.ReportRange(reportRange, preferRecordMessage)
		return
	}

	keyType = skipParenthesizedType(keyType)
	valueType = skipParenthesizedType(valueType)
	ctx.ReportRangeWithDeferredFixesAndSuggestions(
		reportRange,
		preferRecordMessage,
		func() []rule.RuleFix {
			if hasUnpreservedComments(ctx, reportRange, keyType, valueType) {
				return nil
			}
			return []rule.RuleFix{rule.RuleFixReplaceRange(
				reportRange,
				indexSignatureRecordReplacement(ctx, interfaceDeclaration, indexSignature, keyType, valueType, postfix),
			)}
		},
		func() []rule.RuleSuggestion {
			if !hasUnpreservedComments(ctx, reportRange, keyType, valueType) {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message: preferRecordSuggestionMessage,
				FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(
					reportRange,
					indexSignatureRecordReplacement(ctx, interfaceDeclaration, indexSignature, keyType, valueType, postfix),
				)},
			}}
		},
	)
}

func indexSignatureRecordReplacement(
	ctx *rule.RuleContext,
	interfaceDeclaration *ast.InterfaceDeclaration,
	indexSignature *ast.Node,
	keyType *ast.Node,
	valueType *ast.Node,
	postfix string,
) string {
	key := utils.TrimmedNodeText(ctx.SourceFile, keyType)
	value := utils.TrimmedNodeText(ctx.SourceFile, valueType)
	record := "Record<" + key + ", " + value + ">"
	if ast.HasSyntacticModifier(indexSignature, ast.ModifierFlagsReadonly) {
		record = "Readonly<" + record + ">"
	}
	if interfaceDeclaration == nil {
		return record + postfix
	}
	return interfaceTypeAliasPrefix(ctx.SourceFile, interfaceDeclaration) + record + postfix
}

func interfaceTypeAliasPrefix(sourceFile *ast.SourceFile, declaration *ast.InterfaceDeclaration) string {
	var text strings.Builder
	text.WriteString("type ")
	text.WriteString(declaration.Name().Text())
	if declaration.TypeParameters != nil && len(declaration.TypeParameters.Nodes) > 0 {
		text.WriteByte('<')
		for index, typeParameter := range declaration.TypeParameters.Nodes {
			if index > 0 {
				text.WriteString(", ")
			}
			text.WriteString(utils.TrimmedNodeText(sourceFile, typeParameter))
		}
		text.WriteByte('>')
	}
	text.WriteString(" = ")
	return text.String()
}

func interfaceDeclarationRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	modifiers := node.Modifiers()
	if modifiers == nil || len(modifiers.Nodes) == 0 {
		return nodeRange
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind != ast.KindExportKeyword && modifier.Kind != ast.KindDefaultKeyword {
			return core.NewTextRange(utils.TrimNodeTextRange(sourceFile, modifier).Pos(), node.End())
		}
	}
	if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		return nodeRange
	}

	s := scanner.GetScannerForSourceFile(sourceFile, nodeRange.Pos())
	for s.TokenStart() < node.End() {
		if s.Token() == ast.KindInterfaceKeyword {
			return core.NewTextRange(s.TokenStart(), node.End())
		}
		s.Scan()
	}
	return nodeRange
}

func bracedTypeRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	sourceText := sourceFile.Text()
	start := node.Pos()
	end := node.End()
	if start >= 0 && end <= len(sourceText) && start < end {
		if sourceText[start] == '{' {
			return core.NewTextRange(start, end)
		}
		if start+1 < end && isASCIIWhitespace(sourceText[start]) && sourceText[start+1] == '{' {
			return core.NewTextRange(start+1, end)
		}
		for start < end && isASCIIWhitespace(sourceText[start]) {
			start++
		}
		if start < end && sourceText[start] == '{' {
			return core.NewTextRange(start, end)
		}
	}
	return utils.TrimNodeTextRange(sourceFile, node)
}

func isASCIIWhitespace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}

func reportMappedTypeAsRecord(
	ctx *rule.RuleContext,
	circularReferences *circularReferenceAnalyzer,
	node *ast.Node,
	mappedType *ast.MappedTypeNode,
	typeParameter *ast.TypeParameterDeclaration,
	constraint *ast.Node,
) {
	if mappedTypeReferencesKey(ctx, mappedType, typeParameter) {
		return
	}
	if node.Parent != nil {
		circularStart := node.Parent
		if circularStart.Kind == ast.KindTypeAliasDeclaration {
			circularStart = mappedType.Type
		}
		if circularStart != nil && canDeeplyReferenceType(circularStart.Kind) {
			if parentDeclaration := findParentTypeAlias(node); parentDeclaration != nil &&
				circularReferences.referencesDeclaration(parentDeclaration, circularStart) {
				return
			}
		}
	}

	reportRange := bracedTypeRange(ctx.SourceFile, node)
	if mappedType.ReadonlyToken != nil && mappedType.ReadonlyToken.Kind == ast.KindMinusToken {
		ctx.ReportRange(reportRange, preferRecordMessage)
		return
	}

	constraint = skipParenthesizedType(constraint)
	valueType := skipParenthesizedType(mappedType.Type)
	ctx.ReportRangeWithDeferredFixesAndSuggestions(
		reportRange,
		preferRecordMessage,
		func() []rule.RuleFix {
			if hasUnpreservedComments(ctx, reportRange, constraint, valueType) {
				return nil
			}
			return []rule.RuleFix{rule.RuleFixReplaceRange(
				reportRange,
				mappedTypeRecordReplacement(ctx, mappedType, constraint, valueType),
			)}
		},
		func() []rule.RuleSuggestion {
			if !hasUnpreservedComments(ctx, reportRange, constraint, valueType) {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message: preferRecordSuggestionMessage,
				FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(
					reportRange,
					mappedTypeRecordReplacement(ctx, mappedType, constraint, valueType),
				)},
			}}
		},
	)
}

func mappedTypeRecordReplacement(
	ctx *rule.RuleContext,
	mappedType *ast.MappedTypeNode,
	constraint *ast.Node,
	valueType *ast.Node,
) string {
	key := utils.TrimmedNodeText(ctx.SourceFile, constraint)
	value := "any"
	if valueType != nil {
		value = utils.TrimmedNodeText(ctx.SourceFile, valueType)
	}
	record := "Record<" + key + ", " + value + ">"
	if mappedType.QuestionToken != nil {
		switch mappedType.QuestionToken.Kind {
		case ast.KindQuestionToken, ast.KindPlusToken:
			record = "Partial<" + record + ">"
		case ast.KindMinusToken:
			record = "Required<" + record + ">"
		}
	}
	if mappedType.ReadonlyToken != nil {
		switch mappedType.ReadonlyToken.Kind {
		case ast.KindReadonlyKeyword, ast.KindPlusToken:
			record = "Readonly<" + record + ">"
		}
	}
	return record
}

func mappedTypeReferencesKey(ctx *rule.RuleContext, mappedType *ast.MappedTypeNode, typeParameter *ast.TypeParameterDeclaration) bool {
	target := mappedType.TypeParameter.Symbol()
	name := typeParameter.Name().Text()
	return subtreeReferencesSymbol(ctx, mappedType.NameType, target, name) ||
		subtreeReferencesSymbol(ctx, mappedType.Type, target, name)
}

func subtreeReferencesSymbol(ctx *rule.RuleContext, node *ast.Node, target *ast.Symbol, fallbackName string) bool {
	if node == nil || isIdentifierFreeType(node.Kind) {
		return false
	}
	sourceText := ctx.SourceFile.Text()
	if node.Pos() >= 0 && node.End() <= len(sourceText) && node.Pos() <= node.End() {
		nodeText := sourceText[node.Pos():node.End()]
		if !strings.Contains(nodeText, "\\") && !strings.Contains(nodeText, fallbackName) {
			return false
		}
	}
	return subtreeReferencesSymbolWorker(ctx.Refs, node, target, fallbackName)
}

func isIdentifierFreeType(kind ast.Kind) bool {
	switch kind {
	case ast.KindAnyKeyword,
		ast.KindBigIntKeyword,
		ast.KindBooleanKeyword,
		ast.KindNeverKeyword,
		ast.KindNumberKeyword,
		ast.KindObjectKeyword,
		ast.KindStringKeyword,
		ast.KindSymbolKeyword,
		ast.KindUndefinedKeyword,
		ast.KindUnknownKeyword,
		ast.KindVoidKeyword:
		return true
	default:
		return false
	}
}

func subtreeReferencesSymbolWorker(refs *rule.RefStore, node *ast.Node, target *ast.Symbol, fallbackName string) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier && node.Text() == fallbackName {
		if refs == nil {
			return true
		}
		if symbol := refs.ResolveInFile(node); symbol != nil && symbol == target {
			return true
		}
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if subtreeReferencesSymbolWorker(refs, child, target, fallbackName) {
			found = true
			return true
		}
		return false
	})
	return found
}

func reportRecordAsIndexSignature(ctx *rule.RuleContext, node *ast.Node) {
	typeReference := node.AsTypeReferenceNode()
	if typeReference == nil || typeReference.TypeName == nil || typeReference.TypeName.Kind != ast.KindIdentifier ||
		typeReference.TypeName.Text() != "Record" || typeReference.TypeArguments == nil || len(typeReference.TypeArguments.Nodes) != 2 {
		return
	}

	keyType := skipParenthesizedType(typeReference.TypeArguments.Nodes[0])
	valueType := skipParenthesizedType(typeReference.TypeArguments.Nodes[1])
	reportRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	ctx.ReportRangeWithDeferredFixesAndSuggestions(
		reportRange,
		preferIndexSignatureMessage,
		func() []rule.RuleFix {
			if !isAutofixableRecordKey(keyType.Kind) || hasUnpreservedComments(ctx, reportRange, keyType, valueType) {
				return nil
			}
			return []rule.RuleFix{rule.RuleFixReplaceRange(
				reportRange,
				recordIndexSignatureReplacement(ctx, keyType, valueType),
			)}
		},
		func() []rule.RuleSuggestion {
			if isAutofixableRecordKey(keyType.Kind) && !hasUnpreservedComments(ctx, reportRange, keyType, valueType) {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message: preferIndexSignatureSuggestionMessage,
				FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(
					reportRange,
					recordIndexSignatureReplacement(ctx, keyType, valueType),
				)},
			}}
		},
	)
}

func isAutofixableRecordKey(kind ast.Kind) bool {
	return kind == ast.KindStringKeyword || kind == ast.KindNumberKeyword || kind == ast.KindSymbolKeyword
}

func recordIndexSignatureReplacement(ctx *rule.RuleContext, keyType *ast.Node, valueType *ast.Node) string {
	key := utils.TrimmedNodeText(ctx.SourceFile, keyType)
	value := utils.TrimmedNodeText(ctx.SourceFile, valueType)
	return "{ [key: " + key + "]: " + value + " }"
}

func hasUnpreservedComments(ctx *rule.RuleContext, nodeRange core.TextRange, preserved ...*ast.Node) bool {
	if ctx.Comments == nil {
		return false
	}
	for _, comment := range ctx.Comments.All() {
		if comment.End() <= nodeRange.Pos() {
			continue
		}
		if comment.Pos() >= nodeRange.End() {
			break
		}
		if comment.Pos() < nodeRange.Pos() || comment.End() > nodeRange.End() {
			continue
		}
		isPreserved := false
		for _, target := range preserved {
			if target == nil {
				continue
			}
			targetRange := utils.TrimNodeTextRange(ctx.SourceFile, target)
			if comment.Pos() >= targetRange.Pos() && comment.End() <= targetRange.End() {
				isPreserved = true
				break
			}
		}
		if !isPreserved {
			return true
		}
	}
	return false
}

func skipParenthesizedType(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		parenthesizedType := node.AsParenthesizedTypeNode()
		if parenthesizedType == nil || parenthesizedType.Type == nil {
			break
		}
		node = parenthesizedType.Type
	}
	return node
}

func findParentTypeAlias(node *ast.Node) *ast.Node {
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		if parent.Kind == ast.KindTypeAliasDeclaration {
			return parent
		}
		if isTypeAnnotationBoundary(parent) {
			return nil
		}
		current = parent
	}
	return nil
}

func isTypeAnnotationBoundary(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindConstructorType,
		ast.KindFunctionType,
		ast.KindIndexSignature,
		ast.KindMethodSignature,
		ast.KindParameter,
		ast.KindPropertySignature:
		return true
	default:
		return false
	}
}

type circularReferenceAnalyzer struct {
	refs *rule.RefStore
}

func (a *circularReferenceAnalyzer) referencesDeclaration(targetDeclaration *ast.Node, start *ast.Node) bool {
	if start == nil || !canDeeplyReferenceType(start.Kind) || targetDeclaration == nil || targetDeclaration.Name() == nil ||
		targetDeclaration.Name().Kind != ast.KindIdentifier {
		return false
	}
	checker := circularReferenceChecker{
		refs:              a.refs,
		targetDeclaration: targetDeclaration,
		targetSymbol:      circularTargetSymbol(targetDeclaration),
		targetName:        targetDeclaration.Name().Text(),
	}
	return checker.check(start)
}

func circularTargetSymbol(declaration *ast.Node) *ast.Symbol {
	var typeParameters *ast.NodeList
	switch declaration.Kind {
	case ast.KindTypeAliasDeclaration:
		if typed := declaration.AsTypeAliasDeclaration(); typed != nil {
			typeParameters = typed.TypeParameters
		}
	case ast.KindInterfaceDeclaration:
		if typed := declaration.AsInterfaceDeclaration(); typed != nil {
			typeParameters = typed.TypeParameters
		}
	}
	if typeParameters != nil {
		declarationName := declaration.Name().Text()
		for _, parameter := range typeParameters.Nodes {
			if name := parameter.Name(); name != nil && name.Kind == ast.KindIdentifier && name.Text() == declarationName {
				return parameter.Symbol()
			}
		}
	}
	return declaration.Symbol()
}

func canDeeplyReferenceType(kind ast.Kind) bool {
	switch kind {
	case ast.KindIdentifier,
		ast.KindTypeLiteral,
		ast.KindTypeAliasDeclaration,
		ast.KindIndexedAccessType,
		ast.KindMappedType,
		ast.KindConditionalType,
		ast.KindUnionType,
		ast.KindIntersectionType,
		ast.KindInterfaceDeclaration,
		ast.KindIndexSignature,
		ast.KindTypeReference,
		ast.KindParenthesizedType:
		return true
	default:
		return false
	}
}

type circularReferenceChecker struct {
	refs              *rule.RefStore
	targetDeclaration *ast.Node
	targetSymbol      *ast.Symbol
	targetName        string
	visitedSmall      [8]*ast.Node
	visitedCount      int
	visitedLarge      map[*ast.Node]struct{}
}

func (c *circularReferenceChecker) check(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return c.checkIdentifier(node)
	case ast.KindTypeLiteral:
		typeLiteral := node.AsTypeLiteralNode()
		return typeLiteral != nil && c.checkNodes(typeLiteral.Members)
	case ast.KindTypeAliasDeclaration:
		declaration := node.AsTypeAliasDeclaration()
		return declaration != nil && c.check(declaration.Type)
	case ast.KindIndexedAccessType:
		indexedAccessType := node.AsIndexedAccessTypeNode()
		return indexedAccessType != nil && (c.check(indexedAccessType.IndexType) || c.check(indexedAccessType.ObjectType))
	case ast.KindMappedType:
		mappedType := node.AsMappedTypeNode()
		return mappedType != nil && c.check(mappedType.Type)
	case ast.KindConditionalType:
		conditionalType := node.AsConditionalTypeNode()
		return conditionalType != nil &&
			(c.check(conditionalType.CheckType) ||
				c.check(conditionalType.ExtendsType) ||
				c.check(conditionalType.FalseType) ||
				c.check(conditionalType.TrueType))
	case ast.KindUnionType:
		unionType := node.AsUnionTypeNode()
		return unionType != nil && c.checkNodes(unionType.Types)
	case ast.KindIntersectionType:
		intersectionType := node.AsIntersectionTypeNode()
		return intersectionType != nil && c.checkNodes(intersectionType.Types)
	case ast.KindInterfaceDeclaration:
		declaration := node.AsInterfaceDeclaration()
		return declaration != nil && c.checkNodes(declaration.Members)
	case ast.KindIndexSignature:
		indexSignature := node.AsIndexSignatureDeclaration()
		return indexSignature != nil && c.check(indexSignature.Type)
	case ast.KindTypeReference:
		typeReference := node.AsTypeReferenceNode()
		return typeReference != nil &&
			(c.check(typeReference.TypeName) || c.checkNodes(typeReference.TypeArguments))
	case ast.KindParenthesizedType:
		parenthesizedType := node.AsParenthesizedTypeNode()
		return parenthesizedType != nil && c.check(parenthesizedType.Type)
	default:
		return false
	}
}

func (c *circularReferenceChecker) checkNodes(nodes *ast.NodeList) bool {
	if nodes == nil {
		return false
	}
	for _, node := range nodes.Nodes {
		if c.check(node) {
			return true
		}
	}
	return false
}

func (c *circularReferenceChecker) checkIdentifier(identifier *ast.Node) bool {
	var symbol *ast.Symbol
	if c.refs != nil {
		symbol = c.refs.ResolveInFile(identifier)
	}
	if symbol == nil {
		return identifier.Text() == c.targetName
	}
	if symbol == c.targetSymbol || symbolDeclaresNode(symbol, c.targetDeclaration) {
		return true
	}

	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindTypeAliasDeclaration && declaration.Kind != ast.KindInterfaceDeclaration {
			continue
		}
		if !c.markDeclarationVisited(declaration) {
			continue
		}
		if c.check(declaration) {
			return true
		}
	}
	return false
}

func (c *circularReferenceChecker) markDeclarationVisited(declaration *ast.Node) bool {
	for index := 0; index < c.visitedCount && index < len(c.visitedSmall); index++ {
		if c.visitedSmall[index] == declaration {
			return false
		}
	}
	if c.visitedLarge != nil {
		if _, seen := c.visitedLarge[declaration]; seen {
			return false
		}
		c.visitedLarge[declaration] = struct{}{}
		c.visitedCount++
		return true
	}
	if c.visitedCount < len(c.visitedSmall) {
		c.visitedSmall[c.visitedCount] = declaration
		c.visitedCount++
		return true
	}

	c.visitedLarge = make(map[*ast.Node]struct{}, len(c.visitedSmall)*2)
	for _, visited := range c.visitedSmall {
		c.visitedLarge[visited] = struct{}{}
	}
	c.visitedLarge[declaration] = struct{}{}
	c.visitedCount++
	return true
}

func symbolDeclaresNode(symbol *ast.Symbol, declaration *ast.Node) bool {
	if symbol == nil || declaration == nil {
		return false
	}
	for _, candidate := range symbol.Declarations {
		if candidate == declaration {
			return true
		}
	}
	return false
}
