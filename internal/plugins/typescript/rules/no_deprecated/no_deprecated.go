package no_deprecated

import (
	"context"
	_ "embed"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_deprecated.schema.json
var schemaJSON []byte

var deprecatedReasonPattern = regexp.MustCompile(`(?s)@deprecated\s*([\s\S]*?)\*/`)

const (
	diagnosticCodeSecondEntityName     = 6387
	declarationReasonSearchWindowBytes = 512
	maxAliasDeprecationDepth           = 64
)

type sourceDeprecationLookupKind uint8

const (
	sourceDeprecationAny sourceDeprecationLookupKind = iota
	sourceDeprecationVariable
	sourceDeprecationStructuralProperty
	sourceDeprecationMethod
	sourceDeprecationFunction
	sourceDeprecationProperty
)

type sourceDeprecationLookupKey struct {
	kind          sourceDeprecationLookupKind
	name          string
	argCount      int
	matchArgCount bool
}

type sourceDeprecationInfo struct {
	isDeprecated bool
	reason       string
}

type callLikeDeprecationInfo struct {
	checked                        bool
	callLikeNode                   *ast.Node
	isDeprecated                   bool
	reason                         string
	resolvedNonDeprecatedSignature bool
}

func buildDeprecatedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "deprecated",
		Description: "`" + name + "` is deprecated.",
	}
}

func buildDeprecatedWithReasonMessage(name string, reason string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "deprecatedWithReason",
		Description: "`" + name + "` is deprecated. " + reason,
	}
}

func isNodeCalleeOfParent(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindNewExpression:
		newExpr := node.Parent.AsNewExpression()
		return newExpr != nil && newExpr.Expression == node
	case ast.KindCallExpression:
		callExpr := node.Parent.AsCallExpression()
		return callExpr != nil && callExpr.Expression == node
	case ast.KindTaggedTemplateExpression:
		taggedTemplate := node.Parent.AsTaggedTemplateExpression()
		return taggedTemplate != nil && taggedTemplate.Tag == node
	case ast.KindJsxOpeningElement:
		jsxOpening := node.Parent.AsJsxOpeningElement()
		return jsxOpening != nil && jsxOpening.TagName == node
	case ast.KindJsxSelfClosingElement:
		jsxSelfClosing := node.Parent.AsJsxSelfClosingElement()
		return jsxSelfClosing != nil && jsxSelfClosing.TagName == node
	default:
		return false
	}
}

func getCallLikeNode(node *ast.Node) *ast.Node {
	callee := node
	for callee != nil && callee.Parent != nil && callee.Parent.Kind == ast.KindPropertyAccessExpression {
		parentAccess := callee.Parent.AsPropertyAccessExpression()
		if parentAccess == nil || parentAccess.Name() == nil || parentAccess.Name().AsNode() != callee {
			break
		}
		callee = callee.Parent
	}
	if isNodeCalleeOfParent(callee) {
		return callee
	}
	return nil
}

func getReportedNodeName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindSuperKeyword {
		return "super"
	}
	if node.Kind == ast.KindPrivateIdentifier {
		privateIdentifier := node.AsPrivateIdentifier()
		if privateIdentifier != nil {
			return "#" + strings.TrimPrefix(privateIdentifier.Text, "#")
		}
	}
	return node.Text()
}

func reportedIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindSuperKeyword:
		return getReportedNodeName(node)
	default:
		return ""
	}
}

func getJsDocDeprecationFromNode(typeChecker *checker.Checker, node *ast.Node) string {
	if node == nil {
		return ""
	}
	jsDocs := node.JSDoc(nil)
	for _, jsdoc := range jsDocs {
		tags := jsdoc.AsJSDoc().Tags
		if tags == nil {
			continue
		}
		for _, tagNode := range tags.Nodes {
			if !ast.IsJSDocDeprecatedTag(tagNode) {
				continue
			}
			deprecatedTag := tagNode.AsJSDocDeprecatedTag()
			if deprecatedTag != nil {
				return ecmascript.StringTrim(jsDocCommentText(typeChecker, deprecatedTag.Comment))
			}
			return ""
		}
	}
	return ""
}

func jsDocCommentText(typeChecker *checker.Checker, comment *ast.NodeList) string {
	if comment == nil {
		return ""
	}
	if len(comment.Nodes) == 1 && comment.Nodes[0].Kind == ast.KindJSDocText {
		return comment.Nodes[0].Text()
	}
	var text strings.Builder
	for _, commentNode := range comment.Nodes {
		switch commentNode.Kind {
		case ast.KindJSDocText:
			text.WriteString(commentNode.Text())
		case ast.KindJSDocLink, ast.KindJSDocLinkCode, ast.KindJSDocLinkPlain:
			text.WriteString(formatJSDocLink(typeChecker, commentNode))
		}
	}
	return text.String()
}

// formatJSDocLink mirrors TypeScript's buildLinkParts followed by
// displayPartsToString, including its intentionally observable spacing.
func formatJSDocLink(typeChecker *checker.Checker, link *ast.Node) string {
	if link == nil {
		return ""
	}
	linkKind := "link"
	switch link.Kind {
	case ast.KindJSDocLinkCode:
		linkKind = "linkcode"
	case ast.KindJSDocLinkPlain:
		linkKind = "linkplain"
	}
	prefix := "{@" + linkKind + " "

	linkText := link.Text()
	nameNode := link.Name()
	if nameNode == nil {
		return prefix + linkText + "}"
	}

	suffix := jsDocLinkNameEnd(linkText)
	name := scanner.GetTextOfNode(nameNode) + linkText[:suffix]
	remainingText := skipJSDocLinkSeparator(linkText[suffix:])
	if jsDocLinkTargetExists(typeChecker, nameNode) {
		return prefix + name + remainingText + "}"
	}

	separator := ""
	if suffix == 0 || (suffix < len(linkText) && linkText[suffix] == '|' && !strings.HasSuffix(name, " ")) {
		separator = " "
	}
	return prefix + name + separator + remainingText + "}"
}

func jsDocLinkTargetExists(typeChecker *checker.Checker, name *ast.Node) bool {
	if typeChecker == nil || name == nil {
		return false
	}
	symbol := typeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return false
	}
	symbol = typeChecker.SkipAlias(symbol)
	if symbol == nil {
		return false
	}
	for _, target := range typeChecker.GetRootSymbols(symbol) {
		if target != nil && (target.ValueDeclaration != nil || len(target.Declarations) > 0) {
			return true
		}
	}
	return false
}

func jsDocLinkNameEnd(text string) int {
	if strings.HasPrefix(text, "://") {
		if separator := strings.IndexByte(text, '|'); separator >= 0 {
			return separator
		}
		return len(text)
	}
	if strings.HasPrefix(text, "()") {
		return 2
	}
	if len(text) == 0 || text[0] != '<' {
		return 0
	}
	depth := 0
	for index, character := range text {
		switch character {
		case '<':
			depth++
		case '>':
			depth--
		}
		if depth == 0 {
			return index + 1
		}
	}
	return 0
}

func skipJSDocLinkSeparator(text string) string {
	if len(text) == 0 || text[0] != '|' {
		return text
	}
	index := 1
	for index < len(text) && text[index] == ' ' {
		index++
	}
	return text[index:]
}

func hasDeprecatedTag(node *ast.Node) bool {
	if node == nil {
		return false
	}
	jsDocs := node.JSDoc(nil)
	for _, jsdoc := range jsDocs {
		tags := jsdoc.AsJSDoc().Tags
		if tags == nil {
			continue
		}
		for _, tagNode := range tags.Nodes {
			if ast.IsJSDocDeprecatedTag(tagNode) {
				return true
			}
		}
	}
	return false
}

func hasDeprecatedTagInSource(node *ast.Node) bool {
	if node == nil {
		return false
	}
	declarationNode := node
	if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
		variableDeclaration := node.Parent.AsVariableDeclaration()
		if variableDeclaration != nil && variableDeclaration.Name() == node {
			declarationNode = node.Parent
		}
	}
	sourceFile := ast.GetSourceFileOfNode(declarationNode)
	if sourceFile == nil {
		return false
	}
	text := sourceFile.Text()
	if text == "" {
		return false
	}

	anchor := declarationNode.Pos()
	if declarationNode.Kind == ast.KindVariableDeclaration {
		variableDeclaration := declarationNode.AsVariableDeclaration()
		if variableDeclaration != nil && variableDeclaration.Name() != nil {
			anchor = variableDeclaration.Name().Pos()
		}
		if declarationNode.Parent != nil && declarationNode.Parent.Kind == ast.KindVariableDeclarationList && declarationNode.Parent.Parent != nil {
			declList := declarationNode.Parent.AsVariableDeclarationList()
			if declList != nil && declList.Declarations != nil && len(declList.Declarations.Nodes) > 0 && declList.Declarations.Nodes[0] == declarationNode {
				anchor = declarationNode.Parent.Parent.Pos()
			}
		}
	}
	if anchor < 0 || anchor > len(text) {
		return false
	}
	if declarationNode.Kind == ast.KindVariableDeclaration && hasDeprecatedLeadingCommentAt(text, anchor) {
		return true
	}
	// Walk backwards over whitespace to find a directly preceding JSDoc comment.
	i := anchor - 1
	for i >= 0 {
		ch := text[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i--
			continue
		}
		break
	}
	if i < 1 || text[i] != '/' || text[i-1] != '*' {
		return false
	}
	commentEnd := i + 1
	commentStart := strings.LastIndex(text[:commentEnd], "/**")
	if commentStart == -1 {
		return false
	}
	comment := text[commentStart:commentEnd]
	if strings.Contains(comment[:len(comment)-2], "*/") {
		return false
	}
	return deprecatedReasonPattern.MatchString(comment)
}

func hasDeprecatedLeadingCommentAt(text string, anchor int) bool {
	if anchor < 0 || anchor >= len(text) {
		return false
	}
	snippet := text[anchor:]
	trimmed := strings.TrimLeft(snippet, " 	\n\r")
	if !strings.HasPrefix(trimmed, "/**") {
		return false
	}
	commentEnd := strings.Index(trimmed, "*/")
	if commentEnd < 0 {
		return false
	}
	comment := trimmed[:commentEnd+2]
	return deprecatedReasonPattern.MatchString(comment)
}

func isDeclarationDeprecated(typeChecker *checker.Checker, node *ast.Node) bool {
	if node == nil {
		return false
	}
	if hasDeprecatedTag(node) || hasDeprecatedTagInSource(node) {
		return true
	}
	if typeChecker == nil {
		return false
	}
	if node.Kind == ast.KindVariableDeclaration {
		return false
	}
	return typeChecker.IsDeprecatedDeclaration(node)
}

func getJsDocDeprecation(typeChecker *checker.Checker, symbol *ast.Symbol) (bool, string) {
	if typeChecker == nil || symbol == nil {
		return false, ""
	}
	for _, decl := range symbol.Declarations {
		if decl == nil {
			continue
		}
		if decl.Kind == ast.KindBindingElement {
			continue
		}
		if isDeclarationDeprecated(typeChecker, decl) {
			reason := deprecatedReasonFromDeclaration(typeChecker, decl)
			return true, reason
		}
	}
	if symbol.ValueDeclaration != nil {
		if symbol.ValueDeclaration.Kind != ast.KindBindingElement &&
			isDeclarationDeprecated(typeChecker, symbol.ValueDeclaration) {
			reason := deprecatedReasonFromDeclaration(typeChecker, symbol.ValueDeclaration)
			return true, reason
		}
	}
	return false, ""
}

func searchForDeprecationInAliasesChain(
	typeChecker *checker.Checker,
	symbol *ast.Symbol,
	checkDeprecationsOfAliasedSymbol bool,
) (bool, string) {
	if typeChecker == nil || symbol == nil {
		return false, ""
	}
	if symbol.Flags&ast.SymbolFlagsAlias == 0 {
		if checkDeprecationsOfAliasedSymbol {
			return getJsDocDeprecation(typeChecker, symbol)
		}
		return false, ""
	}
	targetSymbol := typeChecker.GetAliasedSymbol(symbol)
	for depth := 0; symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 && depth < maxAliasDeprecationDepth; depth++ {
		if symbol == targetSymbol {
			if checkDeprecationsOfAliasedSymbol {
				return getJsDocDeprecation(typeChecker, symbol)
			}
			break
		}
		if isDeprecated, reason := getJsDocDeprecation(typeChecker, symbol); isDeprecated {
			return true, reason
		}
		if len(symbol.Declarations) == 0 {
			break
		}
		immediateAliasedSymbol := typeChecker.GetImmediateAliasedSymbol(symbol)
		if immediateAliasedSymbol == nil {
			break
		}
		if immediateAliasedSymbol == symbol {
			break
		}
		symbol = immediateAliasedSymbol
		if checkDeprecationsOfAliasedSymbol && symbol == targetSymbol {
			return getJsDocDeprecation(typeChecker, symbol)
		}
	}
	return false, ""
}

func stripQuotes(text string) string {
	text = ecmascript.StringTrim(text)
	if len(text) >= 2 {
		if (strings.HasPrefix(text, "'") && strings.HasSuffix(text, "'")) ||
			(strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"")) ||
			(strings.HasPrefix(text, "`") && strings.HasSuffix(text, "`")) {
			return text[1 : len(text)-1]
		}
	}
	return text
}

func normalizeComparableName(text string) string {
	return strings.TrimPrefix(stripQuotes(text), "#")
}

func diagnosticEntityName(diagnostic *ast.Diagnostic) string {
	if diagnostic == nil {
		return ""
	}
	args := diagnostic.MessageArgs()
	if len(args) == 0 {
		return ""
	}
	// Diagnostic 6387 reports the relevant entity name in the second argument.
	if diagnostic.Code() == diagnosticCodeSecondEntityName && len(args) >= 2 {
		return stripQuotes(args[1])
	}
	return stripQuotes(args[0])
}

func sourceSpanText(sourceFile *ast.SourceFile, pos int, end int) string {
	if sourceFile == nil {
		return ""
	}
	text := sourceFile.Text()
	if pos < 0 || end > len(text) || pos >= end {
		return ""
	}
	return text[pos:end]
}

func cleanupDeprecatedReason(text string) string {
	text = ecmascript.StringTrim(text)
	if text == "" {
		return ""
	}
	text = strings.TrimPrefix(text, ":")
	text = ecmascript.StringTrim(text)
	text = strings.TrimSuffix(text, "*/")
	text = ecmascript.StringTrim(text)
	lines := strings.Split(text, "\n")
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := ecmascript.StringTrim(line)
		trimmed = strings.TrimPrefix(trimmed, "*")
		trimmed = ecmascript.StringTrim(trimmed)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return ecmascript.StringTrim(strings.Join(parts, "\n"))
}

func deprecatedReasonFromDiagnostic(diagnostic *ast.Diagnostic) string {
	if diagnostic == nil {
		return ""
	}
	for _, related := range diagnostic.RelatedInformation() {
		if related == nil || related.File() == nil {
			continue
		}
		text := sourceSpanText(related.File(), related.Pos(), related.End())
		if text == "" {
			continue
		}
		at := strings.Index(text, "@deprecated")
		if at < 0 {
			continue
		}
		reason := cleanupDeprecatedReason(text[at+len("@deprecated"):])
		if reason != "" {
			return reason
		}
	}
	return ""
}

func deprecatedReasonFromDeclaration(typeChecker *checker.Checker, declaration *ast.Node) string {
	if declaration == nil {
		return ""
	}
	if reason := getJsDocDeprecationFromNode(typeChecker, declaration); reason != "" {
		return reason
	}
	sourceFile := ast.GetSourceFileOfNode(declaration)
	if sourceFile == nil {
		return ""
	}
	text := sourceFile.Text()
	if text == "" {
		return ""
	}

	start := declaration.Pos()
	if start < 0 || start > len(text) {
		return ""
	}
	windowStart := start - declarationReasonSearchWindowBytes
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := declaration.End()
	if windowEnd < start {
		windowEnd = start
	}
	if windowEnd > len(text) {
		windowEnd = len(text)
	}
	snippet := text[windowStart:windowEnd]
	matches := deprecatedReasonPattern.FindAllStringSubmatch(snippet, -1)
	if len(matches) == 0 {
		return ""
	}
	lastMatch := matches[len(matches)-1]
	if len(lastMatch) < 2 {
		return ""
	}
	return cleanupDeprecatedReason(lastMatch[1])
}

// parseAllowSpecifiers decodes the `allow` option — type specifiers written
// either in the string shorthand or in object form.
func parseAllowSpecifiers(options []any) []utils.TypeOrValueSpecifier {
	if len(options) == 0 {
		return nil
	}
	optsMap, _ := options[0].(map[string]any)
	return utils.ParseTypeOrValueSpecifiers(optsMap["allow"])
}

func mostSpecificNodeContainingRange(node *ast.Node, position int, end int) *ast.Node {
	if node == nil || position < node.Pos() || end > node.End() {
		return nil
	}
	best := node
	node.ForEachChild(func(child *ast.Node) bool {
		if child == nil || position < child.Pos() || end > child.End() {
			return false
		}
		if nested := mostSpecificNodeContainingRange(child, position, end); nested != nil {
			best = nested
			return true
		}
		return false
	})
	return best
}

func diagnosticNode(sourceFile *ast.SourceFile, position int, end int) *ast.Node {
	if sourceFile == nil {
		return nil
	}
	candidates := []int{
		position,
		end - 1,
		(position + end) / 2,
		position + 1,
		position - 1,
	}
	for _, candidate := range candidates {
		if candidate < 0 || candidate >= len(sourceFile.Text()) {
			continue
		}
		if node := ast.GetNodeAtPosition(sourceFile, candidate, true); node != nil {
			if specific := mostSpecificNodeContainingRange(node, position, end); specific != nil {
				return specific
			}
			return node
		}
	}
	return nil
}

func declarationInCurrentFile(symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	if symbol == nil || sourceFile == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration:
			continue
		}
		if ast.GetSourceFileOfNode(declaration) == sourceFile {
			return true
		}
	}
	return false
}

func walkAst(node *ast.Node, visitor func(*ast.Node) bool) bool {
	if node == nil {
		return false
	}
	if visitor(node) {
		return true
	}
	return node.ForEachChild(func(child *ast.Node) bool {
		return walkAst(child, visitor)
	})
}

func nodeNameText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		identifier := node.AsIdentifier()
		if identifier == nil {
			return ""
		}
		return identifier.Text
	case ast.KindPrivateIdentifier:
		privateIdentifier := node.AsPrivateIdentifier()
		if privateIdentifier == nil {
			return ""
		}
		return "#" + privateIdentifier.Text
	case ast.KindStringLiteral:
		stringLiteral := node.AsStringLiteral()
		if stringLiteral == nil {
			return ""
		}
		return stringLiteral.Text
	case ast.KindNumericLiteral:
		numericLiteral := node.AsNumericLiteral()
		if numericLiteral == nil {
			return ""
		}
		return numericLiteral.Text
	case ast.KindComputedPropertyName:
		computedPropertyName := node.AsComputedPropertyName()
		if computedPropertyName == nil || computedPropertyName.Expression == nil {
			return ""
		}
		return nodeNameText(computedPropertyName.Expression)
	case ast.KindBindingElement:
		bindingElement := node.AsBindingElement()
		if bindingElement == nil || bindingElement.Name() == nil {
			return ""
		}
		return nodeNameText(bindingElement.Name())
	case ast.KindSuperKeyword:
		return "super"
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		// Binding patterns do not represent a single declaration name.
		return ""
	default:
		// Avoid calling Node.Text on unsupported kinds (for example binding patterns),
		// which can panic in typescript-go internals.
		return ""
	}
}

func symbolAtLocation(typeChecker *checker.Checker, node *ast.Node) *ast.Symbol {
	if typeChecker == nil || node == nil {
		return nil
	}
	if symbol := typeChecker.GetSymbolAtLocation(node); symbol != nil {
		return symbol
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		if access != nil && access.Name() != nil {
			if symbol := typeChecker.GetSymbolAtLocation(access.Name()); symbol != nil {
				return symbol
			}
		}
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if access != nil && access.ArgumentExpression != nil {
			if symbol := typeChecker.GetSymbolAtLocation(access.ArgumentExpression); symbol != nil {
				return symbol
			}
		}
	case ast.KindJsxOpeningElement:
		opening := node.AsJsxOpeningElement()
		if opening != nil && opening.TagName != nil {
			if symbol := typeChecker.GetSymbolAtLocation(opening.TagName); symbol != nil {
				return symbol
			}
		}
	case ast.KindJsxClosingElement:
		closing := node.AsJsxClosingElement()
		if closing != nil && closing.TagName != nil {
			if symbol := typeChecker.GetSymbolAtLocation(closing.TagName); symbol != nil {
				return symbol
			}
		}
	case ast.KindJsxSelfClosingElement:
		selfClosing := node.AsJsxSelfClosingElement()
		if selfClosing != nil && selfClosing.TagName != nil {
			if symbol := typeChecker.GetSymbolAtLocation(selfClosing.TagName); symbol != nil {
				return symbol
			}
		}
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if symbol := typeChecker.GetSymbolAtLocation(parent); symbol != nil {
			return symbol
		}
		if parent.Kind == ast.KindPropertyAccessExpression {
			access := parent.AsPropertyAccessExpression()
			if access != nil && access.Name() != nil {
				if symbol := typeChecker.GetSymbolAtLocation(access.Name()); symbol != nil {
					return symbol
				}
			}
		}
	}
	return nil
}

func getCallLikeDeprecation(ctx rule.RuleContext, node *ast.Node) (bool, string, bool) {
	if ctx.TypeChecker == nil || node == nil || node.Parent == nil {
		return false, "", false
	}
	signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, node.Parent, nil, checker.CheckModeNormal)
	if signature == nil {
		return false, "", false
	}
	signatureDecl := signature.Declaration()
	signatureDeprecated := signatureDecl != nil && isDeclarationDeprecated(ctx.TypeChecker, signatureDecl)
	resolvedNonDeprecatedSignature := signatureDecl != nil &&
		(signatureDecl.Kind == ast.KindFunctionDeclaration ||
			signatureDecl.Kind == ast.KindMethodDeclaration ||
			signatureDecl.Kind == ast.KindMethodSignature) &&
		!signatureDeprecated
	if signatureDeprecated {
		reason := getJsDocDeprecationFromNode(ctx.TypeChecker, signatureDecl)
		return true, reason, false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false, "", resolvedNonDeprecatedSignature
	}
	aliasedSymbol := symbol
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		aliasedSymbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
	}
	var symbolDeclarationKind ast.Kind
	if aliasedSymbol != nil && len(aliasedSymbol.Declarations) > 0 && aliasedSymbol.Declarations[0] != nil {
		symbolDeclarationKind = aliasedSymbol.Declarations[0].Kind
	}
	if symbolDeclarationKind != ast.KindMethodDeclaration &&
		symbolDeclarationKind != ast.KindFunctionDeclaration &&
		symbolDeclarationKind != ast.KindMethodSignature {
		isDeprecated, reason := searchForDeprecationInAliasesChain(ctx.TypeChecker, symbol, true)
		return isDeprecated, reason, resolvedNonDeprecatedSignature
	}
	isDeprecated, reason := searchForDeprecationInAliasesChain(ctx.TypeChecker, symbol, false)
	if isDeprecated {
		return true, reason, false
	}
	if signatureDecl == nil && aliasedSymbol != nil {
		isDeprecated, reason := getJsDocDeprecation(ctx.TypeChecker, aliasedSymbol)
		return isDeprecated, reason, false
	}
	return false, "", resolvedNonDeprecatedSignature
}

func getCallLikeDeprecationInfo(ctx rule.RuleContext, node *ast.Node) callLikeDeprecationInfo {
	callLikeNode := getCallLikeNode(node)
	if callLikeNode == nil {
		return callLikeDeprecationInfo{checked: true}
	}
	isDeprecated, reason, resolvedNonDeprecatedSignature := getCallLikeDeprecation(ctx, callLikeNode)
	return callLikeDeprecationInfo{
		checked:                        true,
		callLikeNode:                   callLikeNode,
		isDeprecated:                   isDeprecated,
		reason:                         reason,
		resolvedNonDeprecatedSignature: resolvedNonDeprecatedSignature,
	}
}

func getJsxAttributeDeprecation(ctx rule.RuleContext, elementNode *ast.Node, propertyName string) (bool, string) {
	if ctx.TypeChecker == nil || elementNode == nil || propertyName == "" {
		return false, ""
	}
	var tagName *ast.Node
	switch elementNode.Kind {
	case ast.KindJsxSelfClosingElement:
		tagName = elementNode.AsJsxSelfClosingElement().TagName
	case ast.KindJsxOpeningElement:
		tagName = elementNode.AsJsxOpeningElement().TagName
	}
	if tagName == nil {
		return false, ""
	}
	contextualType := checker.Checker_getContextualType(ctx.TypeChecker, tagName, checker.ContextFlagsNone)
	if contextualType == nil {
		return false, ""
	}
	symbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, contextualType, propertyName)
	return getJsDocDeprecation(ctx.TypeChecker, symbol)
}

func getBindingPatternSourceType(ctx rule.RuleContext, bindingPattern *ast.Node, seen map[*ast.Node]bool) *checker.Type {
	if ctx.TypeChecker == nil || bindingPattern == nil {
		return nil
	}
	current := bindingPattern
	if current.Kind == ast.KindArrayBindingPattern {
		parentSourceType := getBindingPatternSourceType(ctx, current.Parent, seen)
		if parentSourceType != nil && checker.Checker_isArrayOrTupleType(ctx.TypeChecker, parentSourceType) {
			typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, parentSourceType)
			if len(typeArgs) > 0 && typeArgs[0] != nil {
				return typeArgs[0]
			}
		}
	}
	for current != nil {
		if seen[current] {
			return nil
		}
		seen[current] = true
		switch current.Kind {
		case ast.KindVariableDeclaration:
			varDecl := current.AsVariableDeclaration()
			if varDecl != nil && varDecl.Initializer != nil {
				return utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, varDecl.Initializer)
			}
			return nil
		case ast.KindParameter:
			parameter := current.AsParameterDeclaration()
			if parameter == nil {
				return nil
			}
			if parameter.Type != nil {
				return ctx.TypeChecker.GetTypeAtLocation(parameter.Type)
			}
			if parameter.Initializer != nil {
				return utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, parameter.Initializer)
			}
			return nil
		case ast.KindBindingElement:
			bindingElem := current.AsBindingElement()
			if bindingElem == nil {
				return nil
			}
			if current.Parent == nil {
				return nil
			}
			parentPattern := current.Parent
			parentSourceType := getBindingPatternSourceType(ctx, parentPattern, seen)
			if parentSourceType == nil {
				return nil
			}
			propertyName := bindingElementPropertyName(bindingElem)
			if propertyName == "" {
				return nil
			}
			property := checker.Checker_getPropertyOfType(ctx.TypeChecker, parentSourceType, propertyName)
			if property == nil {
				return nil
			}
			return ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, current)
		case ast.KindArrayBindingPattern:
			parentSourceType := getBindingPatternSourceType(ctx, current.Parent, seen)
			if parentSourceType == nil {
				return nil
			}
			property := checker.Checker_getPropertyOfType(ctx.TypeChecker, parentSourceType, "0")
			if property != nil {
				return ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, current)
			}
			return parentSourceType
		case ast.KindObjectBindingPattern:
			current = current.Parent
			continue
		}
		current = current.Parent
	}
	return nil
}

func getDeprecationReason(
	ctx rule.RuleContext,
	node *ast.Node,
	callLikeDeprecation callLikeDeprecationInfo,
) (bool, string) {
	if ctx.TypeChecker == nil || node == nil {
		return false, ""
	}
	if !callLikeDeprecation.checked {
		callLikeDeprecation = getCallLikeDeprecationInfo(ctx, node)
	}
	if callLikeDeprecation.callLikeNode != nil {
		return callLikeDeprecation.isDeprecated, callLikeDeprecation.reason
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute && node.Kind != ast.KindSuperKeyword {
		if node.Parent.Parent != nil && node.Parent.Parent.Parent != nil {
			return getJsxAttributeDeprecation(ctx, node.Parent.Parent.Parent, node.Text())
		}
	}
	if node.Parent != nil && node.Kind != ast.KindSuperKeyword {
		parent := node.Parent
		if parent.Kind == ast.KindBindingElement {
			bindingElement := parent.AsBindingElement()
			if bindingElement != nil {
				isBindingTarget := bindingElement.Name() == node ||
					(bindingElement.PropertyName != nil && bindingElement.PropertyName == node)
				if isBindingTarget {
					bindingPattern := parent.Parent
					if bindingPattern != nil && (bindingPattern.Kind == ast.KindObjectBindingPattern || bindingPattern.Kind == ast.KindArrayBindingPattern) {
						sourceType := getBindingPatternSourceType(ctx, bindingPattern, map[*ast.Node]bool{})
						if sourceType == nil && bindingPattern.Kind == ast.KindObjectBindingPattern {
							sourceType = ctx.TypeChecker.GetTypeAtLocation(bindingPattern)
						}
						if sourceType == nil {
							sourceType = ctx.TypeChecker.GetTypeAtLocation(bindingPattern)
						}
						if sourceType != nil {
							bindingNode := node
							if bindingElement.PropertyName != nil {
								bindingNode = bindingElement.PropertyName
							}
							propertyName := ""
							if bindingPattern.Kind == ast.KindArrayBindingPattern {
								if index, ok := bindingElementIndex(bindingElement); ok {
									propertyName = strconv.Itoa(index)
								}
							} else {
								if bindingName := bindingElementPropertyName(bindingElement); bindingName != "" {
									propertyName = bindingName
								}
								if propertyName == "" && bindingElement.PropertyName != nil {
									if resolvedName, ok := elementAccessPropertyName(ctx, bindingElement.PropertyName); ok {
										propertyName = resolvedName
									}
								}
								if propertyName == "" {
									propertyName = node.Text()
								}
							}
							if propertyName != "" {
								property := checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceType, propertyName)
								if isDeprecated, reason := getJsDocDeprecation(ctx.TypeChecker, property); isDeprecated {
									return true, reason
								}
								if propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(bindingNode); propertySymbol != nil {
									if propertySymbol.ValueDeclaration != nil && propertySymbol.ValueDeclaration.Kind == ast.KindBindingElement {
										propertySymbol = nil
									}
									if isDeprecated, reason := searchForDeprecationInAliasesChain(ctx.TypeChecker, propertySymbol, true); isDeprecated {
										return true, reason
									}
									if isDeprecated, reason := getJsDocDeprecation(ctx.TypeChecker, propertySymbol); isDeprecated {
										return true, reason
									}
								}
								if property != nil {
									if isDeprecated, reason := searchForDeprecationInAliasesChain(ctx.TypeChecker, property, true); isDeprecated {
										return true, reason
									}
								}
							}
						}
					}
				}
			}
		}
		if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Parent != nil {
			parentType := ctx.TypeChecker.GetTypeAtLocation(parent.Parent)
			if parentType != nil {
				propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				property := checker.Checker_getPropertyOfType(ctx.TypeChecker, parentType, node.Text())
				if isDeprecated, reason := searchForDeprecationInAliasesChain(ctx.TypeChecker, propertySymbol, true); isDeprecated {
					return true, reason
				}
				if isDeprecated, reason := getJsDocDeprecation(ctx.TypeChecker, property); isDeprecated {
					return true, reason
				}
				if isDeprecated, reason := getJsDocDeprecation(ctx.TypeChecker, propertySymbol); isDeprecated {
					return true, reason
				}
			}
		}
	}
	return searchForDeprecationInAliasesChain(ctx.TypeChecker, ctx.TypeChecker.GetSymbolAtLocation(node), true)
}

func propertyAccessForDiagnosticRange(node *ast.Node, pos int, end int) *ast.PropertyAccessExpression {
	for current := node; current != nil; current = current.Parent {
		if current.Kind != ast.KindPropertyAccessExpression {
			continue
		}
		access := current.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil {
			continue
		}
		nameNode := access.Name()
		if nameNode.Pos() == pos && nameNode.End() == end {
			return access
		}
		if pos <= nameNode.Pos() && end >= nameNode.End() {
			return access
		}
		accessNode := access.AsNode()
		if accessNode != nil && pos <= accessNode.Pos() && end >= accessNode.End() {
			return access
		}
	}
	return nil
}

func elementAccessForDiagnosticRange(node *ast.Node, pos int, end int) *ast.ElementAccessExpression {
	for current := node; current != nil; current = current.Parent {
		if current.Kind != ast.KindElementAccessExpression {
			continue
		}
		access := current.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			continue
		}
		argNode := access.ArgumentExpression
		if argNode.Pos() == pos && argNode.End() == end {
			return access
		}
		if pos <= argNode.Pos() && end >= argNode.End() {
			return access
		}
		accessNode := access.AsNode()
		if accessNode != nil && pos <= accessNode.Pos() && end >= accessNode.End() {
			return access
		}
	}
	return nil
}

func isDynamicImportResultIdentifier(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		variableDeclaration := declaration.AsVariableDeclaration()
		if variableDeclaration == nil || variableDeclaration.Initializer == nil {
			continue
		}
		initializer := ast.SkipParentheses(variableDeclaration.Initializer)
		if initializer == nil {
			continue
		}
		if initializer.Kind == ast.KindAwaitExpression {
			awaitExpression := initializer.AsAwaitExpression()
			if awaitExpression == nil {
				continue
			}
			initializer = ast.SkipParentheses(awaitExpression.Expression)
		}
		if initializer == nil || initializer.Kind != ast.KindCallExpression {
			continue
		}
		callExpression := initializer.AsCallExpression()
		if callExpression == nil || callExpression.Expression == nil {
			continue
		}
		callee := ast.SkipParentheses(callExpression.Expression)
		if callee != nil && callee.Kind == ast.KindImportKeyword {
			return true
		}
	}
	return false
}

func isDynamicImportDefaultAccess(node *ast.Node, typeChecker *checker.Checker) bool {
	if node == nil || typeChecker == nil {
		return false
	}
	for current := node; current != nil; current = current.Parent {
		if current.Kind != ast.KindPropertyAccessExpression {
			continue
		}
		access := current.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Expression == nil {
			continue
		}
		if access.Name().Text() != "default" {
			continue
		}
		target := ast.SkipParentheses(access.Expression)
		if target == nil || target.Kind != ast.KindIdentifier {
			continue
		}
		if !isDynamicImportResultIdentifier(typeChecker.GetSymbolAtLocation(target)) {
			continue
		}
		return true
	}
	return false
}

func isPromotedDynamicImportDefaultAccess(access *ast.PropertyAccessExpression, typeChecker *checker.Checker) bool {
	if access == nil || typeChecker == nil || access.Name() == nil || access.Expression == nil {
		return false
	}
	if access.Name().Text() != "default" {
		return false
	}
	target := ast.SkipParentheses(access.Expression)
	if target == nil || target.Kind != ast.KindIdentifier {
		return false
	}
	if !isDynamicImportResultIdentifier(typeChecker.GetSymbolAtLocation(target)) {
		return false
	}
	parent := access.AsNode().Parent
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	parentAccess := parent.AsPropertyAccessExpression()
	return parentAccess != nil &&
		parentAccess.Expression == access.AsNode() &&
		parentAccess.Name() != nil &&
		parentAccess.Name().Text() == "default"
}

func shouldIgnoreDynamicImportDefault(node *ast.Node, pos int, end int, entityName string, typeChecker *checker.Checker) bool {
	if entityName != "default" || node == nil || typeChecker == nil {
		return false
	}
	access := propertyAccessForDiagnosticRange(node, pos, end)
	if access == nil || access.Name() == nil || access.Expression == nil {
		return false
	}
	if access.Name().Text() != "default" {
		return false
	}
	target := ast.SkipParentheses(access.Expression)
	if target == nil || target.Kind != ast.KindIdentifier {
		return false
	}
	if isDynamicImportResultIdentifier(typeChecker.GetSymbolAtLocation(target)) {
		parent := access.AsNode().Parent
		if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
			parentAccess := parent.AsPropertyAccessExpression()
			if parentAccess != nil && parentAccess.Expression == access.AsNode() &&
				parentAccess.Name() != nil && parentAccess.Name().Text() == "default" {
				return false
			}
		}
		return true
	}
	if access.Name() != nil && isDynamicImportResultIdentifier(typeChecker.GetSymbolAtLocation(access.Name())) {
		return true
	}
	return isDynamicImportDefaultAccess(node, typeChecker)
}

func promotedDynamicImportDefaultRange(node *ast.Node, pos int, end int, entityName string, typeChecker *checker.Checker) *core.TextRange {
	if entityName != "default" || node == nil || typeChecker == nil {
		return nil
	}
	access := propertyAccessForDiagnosticRange(node, pos, end)
	if access == nil || access.Expression == nil {
		return nil
	}
	target := ast.SkipParentheses(access.Expression)
	if target == nil || target.Kind != ast.KindIdentifier {
		return nil
	}
	if !isDynamicImportResultIdentifier(typeChecker.GetSymbolAtLocation(target)) {
		return nil
	}
	if access.AsNode().Parent == nil || access.AsNode().Parent.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	parentAccess := access.AsNode().Parent.AsPropertyAccessExpression()
	if parentAccess == nil || parentAccess.Expression != access.AsNode() || parentAccess.Name() == nil {
		return nil
	}
	if parentAccess.Name().Text() != "default" {
		return nil
	}
	grandParent := parentAccess.AsNode().Parent
	if grandParent != nil {
		if grandParent.Kind == ast.KindPropertyAccessExpression || grandParent.Kind == ast.KindElementAccessExpression {
			return nil
		}
	}
	promoted := core.NewTextRange(parentAccess.Name().Pos(), parentAccess.Name().End())
	return &promoted
}

func isWithinJsxClosingElement(node *ast.Node, pos int, end int) bool {
	for current := node; current != nil; current = current.Parent {
		if current.Kind != ast.KindJsxClosingElement {
			continue
		}
		closingElement := current.AsJsxClosingElement()
		if closingElement == nil || closingElement.TagName == nil {
			continue
		}
		if closingElement.TagName.Pos() == pos && closingElement.TagName.End() == end {
			return true
		}
	}
	return false
}

func isImportBindingAtRange(node *ast.Node, pos int, end int) bool {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindImportSpecifier:
			specifier := current.AsImportSpecifier()
			if specifier != nil && specifier.Name() != nil {
				nameNode := specifier.Name()
				if (nameNode.Pos() == pos && nameNode.End() == end) || (pos >= current.Pos() && end <= current.End()) {
					return true
				}
			}
		case ast.KindImportClause:
			clause := current.AsImportClause()
			if clause != nil && clause.Name() != nil {
				nameNode := clause.Name()
				if (nameNode.Pos() == pos && nameNode.End() == end) || (pos >= current.Pos() && end <= current.End()) {
					return true
				}
			}
		case ast.KindNamespaceImport:
			namespaceImport := current.AsNamespaceImport()
			if namespaceImport != nil && namespaceImport.Name() != nil {
				nameNode := namespaceImport.Name()
				if (nameNode.Pos() == pos && nameNode.End() == end) || (pos >= current.Pos() && end <= current.End()) {
					return true
				}
			}
		case ast.KindImportEqualsDeclaration:
			importEquals := current.AsImportEqualsDeclaration()
			if importEquals != nil && importEquals.Name() != nil {
				nameNode := importEquals.Name()
				if (nameNode.Pos() == pos && nameNode.End() == end) || (pos >= current.Pos() && end <= current.End()) {
					return true
				}
			}
		}
	}
	return false
}

func isInImportStatementRange(sourceFile *ast.SourceFile, pos int) bool {
	if sourceFile == nil {
		return false
	}
	text := sourceFile.Text()
	if pos < 0 || pos >= len(text) {
		return false
	}
	lineStart := strings.LastIndex(text[:pos], "\n") + 1
	lineEndRelative := strings.Index(text[pos:], "\n")
	lineEnd := len(text)
	if lineEndRelative >= 0 {
		lineEnd = pos + lineEndRelative
	}
	lineText := text[lineStart:lineEnd]
	trimmedLine := ecmascript.StringTrim(lineText)
	if !strings.HasPrefix(trimmedLine, "import ") {
		return false
	}
	if fromIndex := strings.Index(lineText, " from "); fromIndex >= 0 {
		return pos < lineStart+fromIndex
	}
	return true
}

func symbolIsDeprecated(typeChecker *checker.Checker, symbol *ast.Symbol) bool {
	if typeChecker == nil || symbol == nil {
		return false
	}
	isDeprecated, _ := getJsDocDeprecation(typeChecker, symbol)
	return isDeprecated
}

func resolvedSymbolIsDeprecated(typeChecker *checker.Checker, symbol *ast.Symbol) bool {
	isDeprecated, _ := searchForDeprecationInAliasesChain(typeChecker, symbol, true)
	return isDeprecated
}

func bindingElementPropertyName(bindingElement *ast.BindingElement) string {
	if bindingElement == nil {
		return ""
	}
	if bindingElement.PropertyName != nil {
		switch bindingElement.PropertyName.Kind {
		case ast.KindIdentifier:
			return bindingElement.PropertyName.AsIdentifier().Text
		case ast.KindStringLiteral:
			return bindingElement.PropertyName.AsStringLiteral().Text
		case ast.KindNumericLiteral:
			return bindingElement.PropertyName.Text()
		}
	}
	if bindingElement.Name() != nil && bindingElement.Name().Kind == ast.KindIdentifier {
		return bindingElement.Name().AsIdentifier().Text
	}
	return ""
}

func bindingElementIndex(bindingElement *ast.BindingElement) (int, bool) {
	if bindingElement == nil || bindingElement.Parent == nil || bindingElement.Parent.Kind != ast.KindArrayBindingPattern {
		return 0, false
	}
	pattern := bindingElement.Parent.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return 0, false
	}
	for i, element := range pattern.Elements.Nodes {
		if element == nil {
			continue
		}
		if element == bindingElement.AsNode() {
			return i, true
		}
	}
	return 0, false
}

func elementAccessPropertyName(ctx rule.RuleContext, argument *ast.Node) (string, bool) {
	if ctx.TypeChecker == nil || argument == nil {
		return "", false
	}
	propertyType := ctx.TypeChecker.GetTypeAtLocation(argument)
	if propertyType == nil {
		return "", false
	}
	flags := checker.Type_flags(propertyType)
	if flags&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) == 0 {
		return "", false
	}
	literalType := propertyType.AsLiteralType()
	if literalType == nil {
		return "", false
	}
	if flags&checker.TypeFlagsStringLiteral != 0 {
		value, ok := literalType.Value().(string)
		return value, ok
	}
	if flags&checker.TypeFlagsNumberLiteral != 0 {
		return checker.ValueToString(literalType.Value()), true
	}
	if flags&checker.TypeFlagsBigIntLiteral != 0 {
		// Upstream applies String() to TypeScript's PseudoBigInt object.
		return "[object Object]", true
	}
	return "", false
}

func isPropertyLikeDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyDeclaration,
		ast.KindPropertySignature,
		ast.KindMethodDeclaration,
		ast.KindMethodSignature,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return true
	default:
		return false
	}
}

func declarationNameNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindVariableDeclaration {
		variableDeclaration := node.AsVariableDeclaration()
		if variableDeclaration == nil {
			return nil
		}
		return variableDeclaration.Name()
	}
	return node.Name()
}

func deprecatedInfoByNameInSource(ctx rule.RuleContext, name string, propertyOnly bool) (bool, string) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == "" {
		return false, ""
	}
	targetName := normalizeComparableName(name)
	if targetName == "" {
		return false, ""
	}
	found := false
	reason := ""
	walkAst(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if propertyOnly && !isPropertyLikeDeclaration(node) {
			return false
		}
		nameNode := declarationNameNode(node)
		if nameNode == nil || normalizeComparableName(nodeNameText(nameNode)) != targetName {
			return false
		}
		if !isDeclarationDeprecated(ctx.TypeChecker, node) {
			return false
		}
		found = true
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(ctx.TypeChecker, node)
		}
		return reason != ""
	})
	return found, reason
}

func deprecatedVariableInfoByNameInSource(ctx rule.RuleContext, name string) (bool, string) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == "" {
		return false, ""
	}
	targetName := normalizeComparableName(name)
	if targetName == "" {
		return false, ""
	}
	found := false
	reason := ""
	walkAst(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if node == nil || node.Kind != ast.KindVariableDeclaration {
			return false
		}
		nameNode := declarationNameNode(node)
		if nameNode == nil || normalizeComparableName(nodeNameText(nameNode)) != targetName {
			return false
		}
		if !isDeclarationDeprecated(ctx.TypeChecker, node) {
			return false
		}
		found = true
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(ctx.TypeChecker, node)
		}
		return reason != ""
	})
	return found, reason
}

func deprecatedStructuralPropertyInfoByNameInSource(ctx rule.RuleContext, name string) (bool, string) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == "" {
		return false, ""
	}
	targetName := normalizeComparableName(name)
	if targetName == "" {
		return false, ""
	}
	found := false
	allDeprecated := true
	reason := ""
	walkAst(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind != ast.KindPropertyAssignment &&
			node.Kind != ast.KindPropertySignature &&
			node.Kind != ast.KindPropertyDeclaration {
			return false
		}
		nameNode := declarationNameNode(node)
		if nameNode == nil || normalizeComparableName(nodeNameText(nameNode)) != targetName {
			return false
		}
		found = true
		if !isDeclarationDeprecated(ctx.TypeChecker, node) {
			allDeprecated = false
			return false
		}
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(ctx.TypeChecker, node)
		}
		return reason != ""
	})
	return found && allDeprecated, reason
}

func deprecatedMethodInfoByNameInSource(ctx rule.RuleContext, name string, argCount int, matchArgCount bool) (bool, string) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == "" {
		return false, ""
	}
	targetName := normalizeComparableName(name)
	if targetName == "" {
		return false, ""
	}
	candidates := []*ast.Node{}
	hasSignature := false
	walkAst(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if node == nil || (node.Kind != ast.KindMethodDeclaration && node.Kind != ast.KindMethodSignature) {
			return false
		}
		nameNode := declarationNameNode(node)
		if nameNode == nil || normalizeComparableName(nodeNameText(nameNode)) != targetName {
			return false
		}
		if matchArgCount && len(node.Parameters()) != argCount {
			return false
		}
		candidates = append(candidates, node)
		if node.Body() == nil {
			hasSignature = true
		}
		return false
	})
	found := false
	allDeprecated := true
	reason := ""
	for _, node := range candidates {
		if hasSignature && node.Body() != nil {
			continue
		}
		found = true
		if !isDeclarationDeprecated(ctx.TypeChecker, node) {
			allDeprecated = false
			continue
		}
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(ctx.TypeChecker, node)
		}
	}
	return found && allDeprecated, reason
}

func deprecatedFunctionInfoByNameInSource(ctx rule.RuleContext, name string, argCount int, matchArgCount bool) (bool, string) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == "" {
		return false, ""
	}
	targetName := normalizeComparableName(name)
	if targetName == "" {
		return false, ""
	}
	candidates := []*ast.Node{}
	hasSignature := false
	walkAst(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if node == nil || node.Kind != ast.KindFunctionDeclaration {
			return false
		}
		nameNode := declarationNameNode(node)
		if nameNode == nil || normalizeComparableName(nodeNameText(nameNode)) != targetName {
			return false
		}
		if matchArgCount && len(node.Parameters()) != argCount {
			return false
		}
		candidates = append(candidates, node)
		if node.Body() == nil {
			hasSignature = true
		}
		return false
	})
	found := false
	allDeprecated := true
	reason := ""
	for _, node := range candidates {
		if hasSignature && node.Body() != nil {
			continue
		}
		found = true
		if !isDeclarationDeprecated(ctx.TypeChecker, node) {
			allDeprecated = false
			continue
		}
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(ctx.TypeChecker, node)
		}
	}
	return found && allDeprecated, reason
}

func deprecatedPropertyInfoByNameInSource(ctx rule.RuleContext, name string) (bool, string) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == "" {
		return false, ""
	}
	targetName := normalizeComparableName(name)
	if targetName == "" {
		return false, ""
	}
	found := false
	allDeprecated := true
	reason := ""
	walkAst(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if node == nil || !isPropertyLikeDeclaration(node) {
			return false
		}
		nameNode := declarationNameNode(node)
		if nameNode == nil || normalizeComparableName(nodeNameText(nameNode)) != targetName {
			return false
		}
		found = true
		if !isDeclarationDeprecated(ctx.TypeChecker, node) {
			allDeprecated = false
			return false
		}
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(ctx.TypeChecker, node)
		}
		return reason != ""
	})
	return found && allDeprecated, reason
}

func shouldUseDeprecatedVariableSourceFallback(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.Parent.AsPropertyAccessExpression()
		return access == nil || access.Name() == nil || access.Name().AsNode() != node
	case ast.KindElementAccessExpression, ast.KindJsxAttribute:
		return false
	default:
		return true
	}
}

func bindingElementHasResolvedNonDeprecatedProperty(ctx rule.RuleContext, declaration *ast.Node) bool {
	if ctx.TypeChecker == nil || declaration == nil || declaration.Kind != ast.KindBindingElement || declaration.Parent == nil {
		return false
	}
	bindingElement := declaration.AsBindingElement()
	if bindingElement == nil {
		return false
	}
	bindingPattern := declaration.Parent
	if bindingPattern.Kind != ast.KindObjectBindingPattern && bindingPattern.Kind != ast.KindArrayBindingPattern {
		return false
	}
	sourceType := getBindingPatternSourceType(ctx, bindingPattern, map[*ast.Node]bool{})
	if sourceType == nil {
		return false
	}
	propertyName := bindingElementPropertyName(bindingElement)
	if bindingPattern.Kind == ast.KindArrayBindingPattern {
		index, ok := bindingElementIndex(bindingElement)
		if !ok {
			return false
		}
		propertyName = strconv.Itoa(index)
	}
	if propertyName == "" {
		return false
	}
	property := checker.Checker_getPropertyOfType(ctx.TypeChecker, sourceType, propertyName)
	if property == nil {
		return false
	}
	// Module namespace properties can be export aliases. Upstream checks the
	// export alias's own JSDoc here rather than merging tags from the aliased
	// function's overload declarations. A deprecated re-export still reports,
	// while an unannotated re-export of mixed overloads does not.
	if property.Flags&ast.SymbolFlagsAlias != 0 {
		return !symbolIsDeprecated(ctx.TypeChecker, property)
	}
	return !resolvedSymbolIsDeprecated(ctx.TypeChecker, property)
}

func propertyAccessHasResolvedNonDeprecatedProperty(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.TypeChecker == nil || node == nil {
		return false
	}
	var access *ast.PropertyAccessExpression
	if node.Kind == ast.KindPropertyAccessExpression {
		access = node.AsPropertyAccessExpression()
	} else if isPropertyAccessName(node) {
		access = node.Parent.AsPropertyAccessExpression()
	}
	if access == nil || access.Expression == nil || access.Name() == nil {
		return false
	}
	objectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, access.Expression)
	if objectType == nil {
		return false
	}
	property := checker.Checker_getPropertyOfType(ctx.TypeChecker, objectType, access.Name().Text())
	if property == nil {
		return false
	}
	return !resolvedSymbolIsDeprecated(ctx.TypeChecker, property)
}

func hasResolvedNonDeprecatedAccessTarget(
	ctx rule.RuleContext,
	node *ast.Node,
	callLikeDeprecation callLikeDeprecationInfo,
) bool {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || node == nil {
		return false
	}
	if !callLikeDeprecation.checked {
		callLikeDeprecation = getCallLikeDeprecationInfo(ctx, node)
	}
	if callLikeDeprecation.callLikeNode != nil {
		if callLikeDeprecation.isDeprecated {
			return false
		}
		if callLikeDeprecation.resolvedNonDeprecatedSignature {
			return true
		}
	}
	if propertyAccessHasResolvedNonDeprecatedProperty(ctx, node) {
		return true
	}
	bindingDeclaration := node
	if bindingDeclaration.Kind != ast.KindBindingElement && bindingDeclaration.Parent != nil && bindingDeclaration.Parent.Kind == ast.KindBindingElement {
		bindingDeclaration = bindingDeclaration.Parent
	}
	if bindingElementHasResolvedNonDeprecatedProperty(ctx, bindingDeclaration) {
		return true
	}
	// References to a destructured variable resolve to the local binding symbol,
	// not to the source object's property. Follow that symbol back to its binding
	// element so an unrelated deprecated property with the same name cannot win
	// through the source-wide fallback.
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration != nil && declaration.Kind == ast.KindBindingElement &&
				bindingElementHasResolvedNonDeprecatedProperty(ctx, declaration) {
				return true
			}
		}
		if symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindBindingElement &&
			bindingElementHasResolvedNonDeprecatedProperty(ctx, symbol.ValueDeclaration) {
			return true
		}
	}
	return false
}

// hasResolvedNonDeprecatedTarget returns true when the checker has resolved the
// identifier to a concrete non-deprecated declaration. A source-wide lookup by
// name is only safe when resolution failed: otherwise an unrelated type with a
// deprecated member of the same name can be mistaken for the resolved target.
func hasResolvedNonDeprecatedTarget(ctx rule.RuleContext, node *ast.Node) bool {
	if propertyAccessHasResolvedNonDeprecatedProperty(ctx, node) {
		return true
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}
	allowCrossFileDeclaration := isPropertyAccessName(node)
	checkDecl := func(decl *ast.Node) bool {
		if decl == nil || decl.Kind == ast.KindShorthandPropertyAssignment {
			return false
		}
		if decl.Kind == ast.KindBindingElement {
			return bindingElementHasResolvedNonDeprecatedProperty(ctx, decl)
		}
		declSourceFile := ast.GetSourceFileOfNode(decl)
		if declSourceFile == nil {
			return false
		}
		if declSourceFile != ctx.SourceFile {
			// Virtual-file programs can expose a second SourceFile instance for
			// a declaration in the current logical file. Preserve the existing
			// source fallback for that case; only a property access resolved to a
			// genuinely different file is conclusive cross-file evidence.
			if !allowCrossFileDeclaration || declSourceFile.FileName() == ctx.SourceFile.FileName() {
				return false
			}
		}
		// If the resolved declaration is deprecated, let the normal detection
		// path (or, if necessary, its source fallback) report it.
		if isDeclarationDeprecated(ctx.TypeChecker, decl) {
			return false
		}
		return true
	}
	for _, decl := range symbol.Declarations {
		if checkDecl(decl) {
			return true
		}
	}
	return checkDecl(symbol.ValueDeclaration)
}

func isPropertyAccessName(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := node.Parent.AsPropertyAccessExpression()
	return access != nil && access.Name() != nil && access.Name().AsNode() == node
}

func isBindingElementNameOrProperty(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindBindingElement {
		return false
	}
	bindingElement := node.Parent.AsBindingElement()
	if bindingElement == nil {
		return false
	}
	return bindingElement.Name() == node || bindingElement.PropertyName == node
}

func isPropertyAccessCalleeName(node *ast.Node) bool {
	if !isPropertyAccessName(node) || node.Parent.Parent == nil {
		return false
	}
	if node.Parent.Parent.Kind != ast.KindCallExpression {
		return false
	}
	callExpression := node.Parent.Parent.AsCallExpression()
	return callExpression != nil && callExpression.Expression == node.Parent
}

func propertyAccessCallArgCount(node *ast.Node) (int, bool) {
	if !isPropertyAccessCalleeName(node) {
		return 0, false
	}
	callExpression := node.Parent.Parent.AsCallExpression()
	if callExpression == nil || callExpression.Arguments == nil {
		return 0, true
	}
	return len(callExpression.Arguments.Nodes), true
}

func callArgCountForIdentifierCallee(node *ast.Node) (int, bool) {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindCallExpression {
		return 0, false
	}
	callExpression := node.Parent.AsCallExpression()
	if callExpression == nil || callExpression.Expression != node {
		return 0, false
	}
	if callExpression.Arguments == nil {
		return 0, true
	}
	return len(callExpression.Arguments.Nodes), true
}

func isIdentifierTaggedTemplateTag(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindTaggedTemplateExpression {
		return false
	}
	taggedTemplate := node.Parent.AsTaggedTemplateExpression()
	return taggedTemplate != nil && taggedTemplate.Tag == node
}

var NoDeprecatedRule = rule.CreateRule(rule.Rule{
	Name:             "no-deprecated",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		allowSpecifiers := parseAllowSpecifiers(options)
		var sourceDeprecationCache map[sourceDeprecationLookupKey]sourceDeprecationInfo
		lookupSourceDeprecation := func(kind sourceDeprecationLookupKind, name string, argCount int, matchArgCount bool) (bool, string) {
			key := sourceDeprecationLookupKey{
				kind:          kind,
				name:          name,
				argCount:      argCount,
				matchArgCount: matchArgCount,
			}
			if info, ok := sourceDeprecationCache[key]; ok {
				return info.isDeprecated, info.reason
			}
			info := sourceDeprecationInfo{}
			switch kind {
			case sourceDeprecationAny:
				info.isDeprecated, info.reason = deprecatedInfoByNameInSource(ctx, name, false)
			case sourceDeprecationVariable:
				info.isDeprecated, info.reason = deprecatedVariableInfoByNameInSource(ctx, name)
			case sourceDeprecationStructuralProperty:
				info.isDeprecated, info.reason = deprecatedStructuralPropertyInfoByNameInSource(ctx, name)
			case sourceDeprecationMethod:
				info.isDeprecated, info.reason = deprecatedMethodInfoByNameInSource(ctx, name, argCount, matchArgCount)
			case sourceDeprecationFunction:
				info.isDeprecated, info.reason = deprecatedFunctionInfoByNameInSource(ctx, name, argCount, matchArgCount)
			case sourceDeprecationProperty:
				info.isDeprecated, info.reason = deprecatedPropertyInfoByNameInSource(ctx, name)
			}
			if sourceDeprecationCache == nil {
				sourceDeprecationCache = make(map[sourceDeprecationLookupKey]sourceDeprecationInfo)
			}
			sourceDeprecationCache[key] = info
			return info.isDeprecated, info.reason
		}
		lastCallLikeDeprecation := callLikeDeprecationInfo{}
		callLikeDeprecationForNode := func(node *ast.Node) callLikeDeprecationInfo {
			callLikeNode := getCallLikeNode(node)
			if lastCallLikeDeprecation.checked && lastCallLikeDeprecation.callLikeNode == callLikeNode {
				return lastCallLikeDeprecation
			}
			lastCallLikeDeprecation = getCallLikeDeprecationInfo(ctx, node)
			return lastCallLikeDeprecation
		}
		// A reference is allowed when the type at that location matches a
		// specifier, or when the referenced value itself does — an export whose
		// type carries a different name is only reachable through the latter.
		isAllowed := func(node *ast.Node) bool {
			if len(allowSpecifiers) == 0 || node == nil {
				return false
			}
			nodeType := ctx.TypeChecker.GetTypeAtLocation(node)
			return utils.TypeMatchesSomeSpecifier(nodeType, allowSpecifiers, nil, ctx.Program()) ||
				utils.ValueMatchesSomeSpecifier(node, allowSpecifiers, ctx.Program(), nodeType)
		}
		// A computed access names no value of its own, so only the type carrying
		// the deprecated property can be allowed there. That type is the one
		// written at the receiver: a specifier matches the type parameter of a
		// generic receiver, not the constraint the property comes from.
		isObjectTypeAllowed := func(expression *ast.Node) bool {
			if len(allowSpecifiers) == 0 || expression == nil {
				return false
			}
			objectType := ctx.TypeChecker.GetTypeAtLocation(expression)
			return utils.TypeMatchesSomeSpecifier(objectType, allowSpecifiers, nil, ctx.Program())
		}
		sourceFile := ctx.SourceFile
		if sourceFile.AsNode().Flags&ast.NodeFlagsAmbient != 0 {
			// Avoid crashing TypeScript's suggestion diagnostics on ambient source files.
			sourceFile = nil
		}
		// Determine if an identifier is part of a declaration (not a usage).
		isDeclaration := func(node *ast.Node) bool {
			parent := node.Parent
			if parent == nil {
				return false
			}
			switch parent.Kind {
			case ast.KindBindingElement:
				bindingElement := parent.AsBindingElement()
				if bindingElement == nil {
					return false
				}
				if bindingElement.PropertyName != nil && bindingElement.PropertyName == node {
					return false
				}
				if bindingElement.Name() == node {
					if bindingElement.PropertyName == nil {
						return false
					}
					return bindingElement.PropertyName.Text() != node.Text()
				}
				return false
			case ast.KindClassExpression:
				fallthrough
			case ast.KindVariableDeclaration:
				fallthrough
			case ast.KindEnumMember:
				fallthrough
			case ast.KindClassDeclaration:
				return parent.Name() == node
			case ast.KindMethodDeclaration:
				fallthrough
			case ast.KindPropertyDeclaration:
				fallthrough
			case ast.KindGetAccessor:
				fallthrough
			case ast.KindSetAccessor:
				fallthrough
			case ast.KindFunctionDeclaration:
				fallthrough
			case ast.KindInterfaceDeclaration:
				fallthrough
			case ast.KindTypeAliasDeclaration:
				return parent.Name() == node
			case ast.KindPropertyAssignment:
				propAssign := parent.AsPropertyAssignment()
				if propAssign == nil {
					return false
				}
				if propAssign.Initializer == node {
					return false
				}
				return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectLiteralExpression
			case ast.KindArrowFunction:
				fallthrough
			case ast.KindFunctionExpression:
				fallthrough
			case ast.KindEnumDeclaration:
				fallthrough
			case ast.KindModuleDeclaration:
				fallthrough
			case ast.KindMethodSignature:
				fallthrough
			case ast.KindPropertySignature:
				fallthrough
			case ast.KindTypeParameter:
				fallthrough
			case ast.KindParameter:
				return true
			case ast.KindImportEqualsDeclaration:
				return parent.Name() == node
			default:
				return false
			}
		}
		isInsideImport := func(node *ast.Node) bool {
			for current := node; current != nil; current = current.Parent {
				kind := current.Kind
				if kind == ast.KindImportDeclaration {
					return true
				}
				if kind == ast.KindSourceFile ||
					kind == ast.KindBlock ||
					kind == ast.KindFunctionDeclaration ||
					kind == ast.KindFunctionExpression ||
					kind == ast.KindArrowFunction ||
					kind == ast.KindClassDeclaration ||
					kind == ast.KindClassExpression {
					return false
				}
			}
			return false
		}
		var reportedRanges map[core.TextRange]struct{}
		claimRange := func(diagnosticRange core.TextRange) bool {
			if _, reported := reportedRanges[diagnosticRange]; reported {
				return false
			}
			if reportedRanges == nil {
				reportedRanges = make(map[core.TextRange]struct{})
			}
			reportedRanges[diagnosticRange] = struct{}{}
			return true
		}
		// Two parallel detection paths: (1) TypeScript's GetSuggestionDiagnostics for
		// symbol/type-driven deprecations, and (2) AST listeners for cases the type
		// checker misses (e.g., JSX attributes, local declarations). Duplicates are
		// deduplicated by text range via reportedRanges. This is intentionally
		// heuristic-based to maintain parity with upstream typescript-eslint behavior.
		// TODO: Consolidate to a single symbol-driven path once external .d.ts
		// deprecations are resolved via type checker.
		if sourceFile != nil {
			diagnostics := ctx.TypeChecker.GetSuggestionDiagnostics(context.Background(), sourceFile)
			for _, diagnostic := range diagnostics {
				if diagnostic == nil || !diagnostic.ReportsDeprecated() || diagnostic.File() != sourceFile {
					continue
				}
				name := diagnosticEntityName(diagnostic)
				node := diagnosticNode(sourceFile, diagnostic.Pos(), diagnostic.End())
				if shouldIgnoreDynamicImportDefault(node, diagnostic.Pos(), diagnostic.End(), name, ctx.TypeChecker) {
					continue
				}
				elementAccess := elementAccessForDiagnosticRange(node, diagnostic.Pos(), diagnostic.End())
				if elementAccess != nil {
					if isObjectTypeAllowed(elementAccess.Expression) {
						continue
					}
				} else if isAllowed(node) {
					continue
				}
				resolvedNode := node
				reportName := name
				if identifierName := reportedIdentifierName(resolvedNode); identifierName != "" {
					reportName = identifierName
				}
				if access := propertyAccessForDiagnosticRange(node, diagnostic.Pos(), diagnostic.End()); access != nil && access.Name() != nil {
					resolvedNode = access.Name()
					if identifierName := reportedIdentifierName(resolvedNode); identifierName != "" {
						reportName = identifierName
					}
				}
				callDeprecation := callLikeDeprecationForNode(resolvedNode)
				if hasResolvedNonDeprecatedAccessTarget(ctx, resolvedNode, callDeprecation) {
					continue
				}
				symbol := symbolAtLocation(ctx.TypeChecker, node)
				if symbol != nil && declarationInCurrentFile(symbol, ctx.SourceFile) {
					if isDeprecatedLocal, _ := getJsDocDeprecation(ctx.TypeChecker, symbol); !isDeprecatedLocal {
						continue
					}
				}
				if isWithinJsxClosingElement(node, diagnostic.Pos(), diagnostic.End()) {
					continue
				}
				if isImportBindingAtRange(node, diagnostic.Pos(), diagnostic.End()) {
					continue
				}
				if isInImportStatementRange(sourceFile, diagnostic.Pos()) {
					continue
				}
				if elementAccess != nil && elementAccess.Expression != nil {
					rawObjectType := ctx.TypeChecker.GetTypeAtLocation(elementAccess.Expression)
					if rawObjectType != nil && (utils.IsTypeAnyType(rawObjectType) || utils.IsTypeUnknownType(rawObjectType)) {
						continue
					}
				}
				diagnosticRange := core.NewTextRange(diagnostic.Pos(), diagnostic.End())
				if promotedRange := promotedDynamicImportDefaultRange(node, diagnostic.Pos(), diagnostic.End(), name, ctx.TypeChecker); promotedRange != nil {
					diagnosticRange = *promotedRange
				}
				if !claimRange(diagnosticRange) {
					continue
				}
				message := buildDeprecatedMessage(reportName)
				reason := ""
				if symbol != nil {
					_, reason = searchForDeprecationInAliasesChain(ctx.TypeChecker, symbol, true)
				}
				if reason == "" {
					reason = deprecatedReasonFromDiagnostic(diagnostic)
				}
				if reason == "" && symbol != nil {
					for _, declaration := range symbol.Declarations {
						if declaration == nil {
							continue
						}
						if reasonFromDeclaration := deprecatedReasonFromDeclaration(ctx.TypeChecker, declaration); reasonFromDeclaration != "" {
							reason = reasonFromDeclaration
							break
						}
					}
				}
				if reason != "" {
					message = buildDeprecatedWithReasonMessage(reportName, reason)
				}
				deprecationNode := node
				if node != nil && node.Kind == ast.KindVariableDeclaration {
					variableDeclaration := node.AsVariableDeclaration()
					if variableDeclaration != nil && variableDeclaration.Initializer != nil {
						deprecationNode = ast.SkipParentheses(variableDeclaration.Initializer)
					}
				}
				if message.Id != "deprecatedWithReason" {
					callDeprecation := callLikeDeprecationForNode(deprecationNode)
					if _, reason := getDeprecationReason(ctx, deprecationNode, callDeprecation); reason != "" {
						message = buildDeprecatedWithReasonMessage(reportName, reason)
					}
				}
				if message.Id != "deprecatedWithReason" {
					if _, reason := lookupSourceDeprecation(sourceDeprecationAny, reportName, 0, false); reason != "" {
						message = buildDeprecatedWithReasonMessage(reportName, reason)
					}
				}
				ctx.ReportRange(diagnosticRange, message)
			}
		}
		checkIdentifier := func(node *ast.Node) {
			if node == nil {
				return
			}
			if isDeclaration(node) || isInsideImport(node) {
				return
			}
			name := getReportedNodeName(node)
			if name == "default" {
				if access := propertyAccessForDiagnosticRange(node, node.Pos(), node.End()); isPromotedDynamicImportDefaultAccess(access, ctx.TypeChecker) {
					return
				}
			}
			if shouldIgnoreDynamicImportDefault(node, node.Pos(), node.End(), name, ctx.TypeChecker) {
				return
			}
			reportName := name
			callDeprecation := callLikeDeprecationForNode(node)
			isDeprecated, reason := getDeprecationReason(ctx, node, callDeprecation)
			allowanceChecked := isDeprecated
			if allowanceChecked && isAllowed(node) {
				return
			}
			resolvedNonDeprecated := !isDeprecated && callDeprecation.resolvedNonDeprecatedSignature
			if isDeprecated && isBindingElementNameOrProperty(node) {
				resolvedNonDeprecated = hasResolvedNonDeprecatedAccessTarget(ctx, node, callDeprecation)
				if resolvedNonDeprecated {
					isDeprecated = false
					reason = ""
				}
			}
			if !isDeprecated && !resolvedNonDeprecated {
				resolvedNonDeprecated = hasResolvedNonDeprecatedTarget(ctx, node)
			}
			if !isDeprecated && shouldUseDeprecatedVariableSourceFallback(node) && !resolvedNonDeprecated {
				isDeprecated, reason = lookupSourceDeprecation(sourceDeprecationVariable, name, 0, false)
			}
			if !isDeprecated && isPropertyAccessName(node) && !resolvedNonDeprecated {
				isDeprecated, reason = lookupSourceDeprecation(sourceDeprecationStructuralProperty, name, 0, false)
				if !isDeprecated {
					argCount, matchArgCount := propertyAccessCallArgCount(node)
					isDeprecated, reason = lookupSourceDeprecation(sourceDeprecationMethod, name, argCount, matchArgCount)
				}
			}
			if !isDeprecated && isBindingElementNameOrProperty(node) && !resolvedNonDeprecated {
				isDeprecated, reason = lookupSourceDeprecation(sourceDeprecationProperty, name, 0, false)
			}
			if !isDeprecated && !resolvedNonDeprecated {
				argCount, matchArgCount := callArgCountForIdentifierCallee(node)
				if matchArgCount {
					isDeprecated, reason = lookupSourceDeprecation(sourceDeprecationFunction, name, argCount, true)
				} else if isIdentifierTaggedTemplateTag(node) {
					isDeprecated, reason = lookupSourceDeprecation(sourceDeprecationFunction, name, 0, false)
				}
			}
			if !isDeprecated {
				return
			}
			if !allowanceChecked && isAllowed(node) {
				return
			}
			trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			diagnosticRange := core.NewTextRange(trimmedRange.Pos(), trimmedRange.End())
			if !claimRange(diagnosticRange) {
				return
			}
			symbol := symbolAtLocation(ctx.TypeChecker, node)
			if symbol == nil {
				symbol = ctx.TypeChecker.GetSymbolAtLocation(node)
			}
			if symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindBindingElement {
				symbol = nil
			}
			message := buildDeprecatedMessage(reportName)
			if reason != "" {
				message = buildDeprecatedWithReasonMessage(reportName, reason)
			} else if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if declaration == nil {
						continue
					}
					if reasonFromDecl := deprecatedReasonFromDeclaration(ctx.TypeChecker, declaration); reasonFromDecl != "" {
						message = buildDeprecatedWithReasonMessage(reportName, reasonFromDecl)
						break
					}
				}
			}
			if message.Id != "deprecatedWithReason" {
				if _, reasonByName := lookupSourceDeprecation(sourceDeprecationAny, name, 0, false); reasonByName != "" {
					message = buildDeprecatedWithReasonMessage(reportName, reasonByName)
				}
			}
			ctx.ReportRange(diagnosticRange, message)
		}
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				identifier := node.AsIdentifier()
				if identifier == nil {
					return
				}
				if node.Parent == nil {
					return
				}
				// Avoid duplicate reports on declarations and import/export boilerplate.
				if isWithinJsxClosingElement(node, node.Pos(), node.End()) {
					return
				}
				if node.Parent.Kind == ast.KindExportDeclaration || node.Parent.Kind == ast.KindNamespaceExport {
					return
				}
				if node.Parent.Kind == ast.KindExportSpecifier {
					exportSpec := node.Parent.AsExportSpecifier()
					if exportSpec != nil {
						isPropertyName := exportSpec.PropertyName != nil && exportSpec.PropertyName.AsNode() == node
						if isPropertyName {
							return
						}
						jsDocs := node.Parent.JSDoc(nil)
						for _, jsdoc := range jsDocs {
							tags := jsdoc.AsJSDoc().Tags
							if tags == nil {
								continue
							}
							for _, tagNode := range tags.Nodes {
								if ast.IsJSDocDeprecatedTag(tagNode) {
									return
								}
							}
						}
					}
				}
				checkIdentifier(node)
			},
			ast.KindSuperKeyword: checkIdentifier,
			ast.KindJsxAttribute: func(node *ast.Node) {
				jsxAttribute := node.AsJsxAttribute()
				if jsxAttribute == nil || jsxAttribute.Name() == nil {
					return
				}
				nameNode := jsxAttribute.Name()
				nameText := nameNode.Text()
				if nameText == "" {
					return
				}
				var attributesType *checker.Type
				propertySymbol := symbolAtLocation(ctx.TypeChecker, nameNode)
				if propertySymbol == nil && node.Parent != nil && node.Parent.Kind == ast.KindJsxAttributes {
					attributesType = ctx.TypeChecker.GetTypeAtLocation(node.Parent)
					if attributesType != nil {
						propertySymbol = checker.Checker_getPropertyOfType(ctx.TypeChecker, attributesType, nameText)
					}
				}
				isDeprecated := symbolIsDeprecated(ctx.TypeChecker, propertySymbol)
				sourceDeprecated, sourceReason := false, ""
				if !isDeprecated && propertySymbol == nil && attributesType != nil && !utils.IsTypeAnyType(attributesType) && !utils.IsTypeUnknownType(attributesType) {
					sourceDeprecated, sourceReason = lookupSourceDeprecation(sourceDeprecationProperty, nameText, 0, false)
				}
				if !isDeprecated && !sourceDeprecated {
					return
				}
				if isAllowed(nameNode) {
					return
				}
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				diagnosticRange := core.NewTextRange(trimmedRange.Pos(), trimmedRange.End())
				if !claimRange(diagnosticRange) {
					return
				}
				message := buildDeprecatedMessage(nameText)
				if propertySymbol != nil {
					for _, declaration := range propertySymbol.Declarations {
						if reason := deprecatedReasonFromDeclaration(ctx.TypeChecker, declaration); reason != "" {
							message = buildDeprecatedWithReasonMessage(nameText, reason)
							break
						}
					}
				}
				if message.Id != "deprecatedWithReason" && sourceReason != "" {
					message = buildDeprecatedWithReasonMessage(nameText, sourceReason)
				}
				ctx.ReportRange(diagnosticRange, message)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				elementAccess := node.AsElementAccessExpression()
				if elementAccess == nil || elementAccess.Expression == nil || elementAccess.ArgumentExpression == nil {
					return
				}
				propertyName, ok := elementAccessPropertyName(ctx, elementAccess.ArgumentExpression)
				if !ok || propertyName == "" {
					return
				}
				rawObjectType := ctx.TypeChecker.GetTypeAtLocation(elementAccess.Expression)
				objectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, elementAccess.Expression)
				var propertySymbol *ast.Symbol
				if objectType != nil {
					propertySymbol = checker.Checker_getPropertyOfType(ctx.TypeChecker, objectType, propertyName)
				}
				isDeprecated := symbolIsDeprecated(ctx.TypeChecker, propertySymbol)
				sourceDeprecated, sourceReason := false, ""
				if !isDeprecated && propertySymbol == nil && objectType != nil && rawObjectType != nil && !utils.IsTypeAnyType(rawObjectType) && !utils.IsTypeUnknownType(rawObjectType) && !utils.IsTypeAnyType(objectType) && !utils.IsTypeUnknownType(objectType) {
					sourceDeprecated, sourceReason = lookupSourceDeprecation(sourceDeprecationStructuralProperty, propertyName, 0, false)
				}
				if !isDeprecated && !sourceDeprecated {
					return
				}
				if isObjectTypeAllowed(elementAccess.Expression) {
					return
				}
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, elementAccess.ArgumentExpression)
				diagnosticRange := core.NewTextRange(trimmedRange.Pos(), trimmedRange.End())
				if !claimRange(diagnosticRange) {
					return
				}
				message := buildDeprecatedMessage(propertyName)
				if propertySymbol != nil {
					for _, declaration := range propertySymbol.Declarations {
						if reason := deprecatedReasonFromDeclaration(ctx.TypeChecker, declaration); reason != "" {
							message = buildDeprecatedWithReasonMessage(propertyName, reason)
							break
						}
					}
				}
				if message.Id != "deprecatedWithReason" && sourceReason != "" {
					message = buildDeprecatedWithReasonMessage(propertyName, sourceReason)
				}
				ctx.ReportRange(diagnosticRange, message)
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				if access == nil || access.Name() == nil {
					return
				}
				if isPromotedDynamicImportDefaultAccess(access, ctx.TypeChecker) {
					return
				}
				// Report the property name if it is deprecated.
				nameNode := access.Name()
				name := nameNode.Text()
				if shouldIgnoreDynamicImportDefault(nameNode, nameNode.Pos(), nameNode.End(), name, ctx.TypeChecker) {
					return
				}
				if isAllowed(nameNode) {
					return
				}
				callDeprecation := callLikeDeprecationForNode(nameNode)
				if !callDeprecation.isDeprecated && callDeprecation.resolvedNonDeprecatedSignature {
					return
				}
				propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				if !symbolIsDeprecated(ctx.TypeChecker, propertySymbol) {
					return
				}
				if hasResolvedNonDeprecatedAccessTarget(ctx, nameNode, callDeprecation) {
					return
				}
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				diagnosticRange := core.NewTextRange(trimmedRange.Pos(), trimmedRange.End())
				if !claimRange(diagnosticRange) {
					return
				}
				_, reason := getDeprecationReason(ctx, nameNode, callDeprecation)
				message := buildDeprecatedMessage(name)
				if reason != "" {
					message = buildDeprecatedWithReasonMessage(name, reason)
				} else if propertySymbol != nil {
					for _, declaration := range propertySymbol.Declarations {
						if reasonFromDecl := deprecatedReasonFromDeclaration(ctx.TypeChecker, declaration); reasonFromDecl != "" {
							message = buildDeprecatedWithReasonMessage(name, reasonFromDecl)
							break
						}
					}
				}
				if message.Id != "deprecatedWithReason" {
					if _, reasonByName := lookupSourceDeprecation(sourceDeprecationAny, name, 0, false); reasonByName != "" {
						message = buildDeprecatedWithReasonMessage(name, reasonByName)
					}
				}
				ctx.ReportRange(diagnosticRange, message)
			},
		}
	},
})
