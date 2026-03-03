package no_deprecated

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var deprecatedReasonPattern = regexp.MustCompile(`(?s)@deprecated\s*([\s\S]*?)\*/`)

const (
	diagnosticCodeSecondEntityName     = 6387
	declarationReasonSearchWindowBytes = 512
	maxConstantPropertyResolveDepth    = 8
)

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
	for {
		if callee == nil || callee.Parent == nil || callee.Parent.Kind != ast.KindPropertyAccessExpression {
			break
		}
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
			return "#" + privateIdentifier.Text
		}
	}
	return node.Text()
}

func getJsDocDeprecationFromNode(node *ast.Node) string {
	if node == nil {
		return ""
	}
	jsdocs := node.JSDoc(nil)
	for _, jsdoc := range jsdocs {
		tags := jsdoc.AsJSDoc().Tags
		if tags == nil {
			continue
		}
		for _, tagNode := range tags.Nodes {
			if !ast.IsJSDocDeprecatedTag(tagNode) {
				continue
			}
			deprecatedTag := tagNode.AsJSDocDeprecatedTag()
			if deprecatedTag != nil && deprecatedTag.Comment != nil && len(deprecatedTag.Comment.Nodes) > 0 {
				var text strings.Builder
				for _, commentNode := range deprecatedTag.Comment.Nodes {
					text.WriteString(commentNode.Text())
				}
				return strings.TrimSpace(text.String())
			}
			return ""
		}
	}
	return ""
}

func hasDeprecatedTag(node *ast.Node) bool {
	if node == nil {
		return false
	}
	jsdocs := node.JSDoc(nil)
	for _, jsdoc := range jsdocs {
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
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil {
		return false
	}
	text := sourceFile.Text()
	if text == "" {
		return false
	}

	anchor := node.Pos()
	if node.Kind == ast.KindVariableDeclaration && node.Parent != nil && node.Parent.Parent != nil {
		anchor = node.Parent.Parent.Pos()
	}
	if anchor <= 0 || anchor > len(text) {
		return false
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
		if typeChecker.IsDeprecatedDeclaration(decl) || hasDeprecatedTag(decl) || hasDeprecatedTagInSource(decl) {
			reason := getJsDocDeprecationFromNode(decl)
			if reason == "" {
				reason = deprecatedReasonFromDeclaration(decl)
			}
			return true, reason
		}
	}
	if symbol.ValueDeclaration != nil {
		if symbol.ValueDeclaration.Kind != ast.KindBindingElement &&
			(typeChecker.IsDeprecatedDeclaration(symbol.ValueDeclaration) || hasDeprecatedTag(symbol.ValueDeclaration) || hasDeprecatedTagInSource(symbol.ValueDeclaration)) {
			reason := getJsDocDeprecationFromNode(symbol.ValueDeclaration)
			if reason == "" {
				reason = deprecatedReasonFromDeclaration(symbol.ValueDeclaration)
			}
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
	if isDeprecated, reason := getJsDocDeprecation(typeChecker, symbol); isDeprecated {
		return true, reason
	}
	if !checkDeprecationsOfAliasedSymbol {
		return false, ""
	}
	aliasedSymbol := typeChecker.GetAliasedSymbol(symbol)
	if aliasedSymbol == nil {
		return false, ""
	}
	return getJsDocDeprecation(typeChecker, aliasedSymbol)
}

func stripQuotes(text string) string {
	text = strings.TrimSpace(text)
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
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	text = strings.TrimPrefix(text, ":")
	text = strings.TrimSpace(text)
	text = strings.TrimSuffix(text, "*/")
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "*")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
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

func deprecatedReasonFromDeclaration(declaration *ast.Node) string {
	if declaration == nil {
		return ""
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

type noDeprecatedAllowEntry struct {
	From    string
	Name    string
	Package string
}

func parseAllowEntries(options any) []noDeprecatedAllowEntry {
	entries := []noDeprecatedAllowEntry{}
	var optionMap map[string]interface{}
	switch value := options.(type) {
	case map[string]interface{}:
		optionMap = value
	case []interface{}:
		if len(value) > 0 {
			if parsedMap, ok := value[0].(map[string]interface{}); ok {
				optionMap = parsedMap
			}
		}
	}
	if optionMap == nil {
		return entries
	}
	rawAllowValue, exists := optionMap["allow"]
	if !exists {
		return entries
	}
	rawAllow, ok := rawAllowValue.([]interface{})
	if !ok {
		return entries
	}
	for _, raw := range rawAllow {
		switch value := raw.(type) {
		case string:
			entries = append(entries, noDeprecatedAllowEntry{
				Name: value,
			})
		case map[string]interface{}:
			entry := noDeprecatedAllowEntry{}
			if name, ok := value["name"].(string); ok {
				entry.Name = name
			}
			if from, ok := value["from"].(string); ok {
				entry.From = from
			}
			if pkg, ok := value["package"].(string); ok {
				entry.Package = pkg
			}
			if entry.Name != "" {
				entries = append(entries, entry)
			}
		}
	}
	return entries
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
			return node
		}
	}
	return nil
}

func symbolHierarchyNames(symbol *ast.Symbol) map[string]bool {
	names := map[string]bool{}
	for current := symbol; current != nil; current = current.Parent {
		if current.Name == "" {
			continue
		}
		names[current.Name] = true
		unquoted := strings.Trim(current.Name, "\"'")
		if unquoted != "" {
			names[unquoted] = true
		}
	}
	return names
}

func declarationInCurrentFile(symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	if symbol == nil || sourceFile == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		if ast.GetSourceFileOfNode(declaration) == sourceFile {
			return true
		}
	}
	return false
}

func packageMatchesSymbol(entryPackage string, symbol *ast.Symbol) bool {
	if entryPackage == "" || symbol == nil {
		return false
	}
	hierarchy := symbolHierarchyNames(symbol)
	if hierarchy[entryPackage] || hierarchy["@types/"+entryPackage] {
		return true
	}
	for current := symbol; current != nil; current = current.Parent {
		for _, declaration := range current.Declarations {
			sourceFile := ast.GetSourceFileOfNode(declaration)
			if sourceFile == nil {
				continue
			}
			fileName := sourceFile.FileName()
			if strings.Contains(fileName, "/node_modules/"+entryPackage+"/") ||
				strings.Contains(fileName, "/node_modules/@types/"+entryPackage+"/") {
				return true
			}
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
	shouldStop := false
	node.ForEachChild(func(child *ast.Node) bool {
		if walkAst(child, visitor) {
			shouldStop = true
			return true
		}
		return false
	})
	return shouldStop
}

func moduleSpecifierText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		stringLiteral := node.AsStringLiteral()
		if stringLiteral != nil {
			return stringLiteral.Text
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		templateLiteral := node.AsNoSubstitutionTemplateLiteral()
		if templateLiteral != nil {
			return templateLiteral.Text
		}
	}
	return stripQuotes(strings.TrimSpace(node.Text()))
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

func importedNameMatches(targetName string, node *ast.Node) bool {
	if targetName == "" || node == nil {
		return false
	}
	return normalizeComparableName(targetName) == normalizeComparableName(nodeNameText(node))
}

func staticImportContainsName(importDeclaration *ast.ImportDeclaration, targetName string) bool {
	if importDeclaration == nil || importDeclaration.ImportClause == nil || targetName == "" {
		return false
	}
	importClause := importDeclaration.ImportClause.AsImportClause()
	if importClause == nil {
		return false
	}
	if importClause.Name() != nil && importedNameMatches(targetName, importClause.Name()) {
		return true
	}
	if importClause.NamedBindings == nil {
		return false
	}
	if importClause.NamedBindings.Kind == ast.KindNamespaceImport {
		namespaceImport := importClause.NamedBindings.AsNamespaceImport()
		if namespaceImport != nil && namespaceImport.Name() != nil && importedNameMatches(targetName, namespaceImport.Name()) {
			return true
		}
		return false
	}
	if importClause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	namedImports := importClause.NamedBindings.AsNamedImports()
	if namedImports == nil || namedImports.Elements == nil {
		return false
	}
	for _, elementNode := range namedImports.Elements.Nodes {
		importSpecifier := elementNode.AsImportSpecifier()
		if importSpecifier == nil {
			continue
		}
		if importSpecifier.Name() != nil && importedNameMatches(targetName, importSpecifier.Name()) {
			return true
		}
		if importSpecifier.PropertyName != nil && importedNameMatches(targetName, importSpecifier.PropertyName) {
			return true
		}
	}
	return false
}

func isImportCallFromPackage(initializer *ast.Node, pkg string) bool {
	if initializer == nil || pkg == "" {
		return false
	}
	current := ast.SkipParentheses(initializer)
	if current == nil {
		return false
	}
	if current.Kind == ast.KindAwaitExpression {
		awaitExpression := current.AsAwaitExpression()
		if awaitExpression == nil {
			return false
		}
		current = ast.SkipParentheses(awaitExpression.Expression)
	}
	if current == nil || current.Kind != ast.KindCallExpression {
		return false
	}
	callExpression := current.AsCallExpression()
	if callExpression == nil || callExpression.Expression == nil || callExpression.Arguments == nil || len(callExpression.Arguments.Nodes) == 0 {
		return false
	}
	callee := ast.SkipParentheses(callExpression.Expression)
	if callee == nil || callee.Kind != ast.KindImportKeyword {
		return false
	}
	return moduleSpecifierText(callExpression.Arguments.Nodes[0]) == pkg
}

func dynamicImportBindingContainsName(nameNode *ast.Node, targetName string) bool {
	if nameNode == nil || targetName == "" || nameNode.Kind != ast.KindObjectBindingPattern {
		return false
	}
	objectBindingPattern := nameNode.AsBindingPattern()
	if objectBindingPattern == nil || objectBindingPattern.Elements == nil {
		return false
	}
	for _, elementNode := range objectBindingPattern.Elements.Nodes {
		bindingElement := elementNode.AsBindingElement()
		if bindingElement == nil {
			continue
		}
		propertyName := bindingElementPropertyName(bindingElement)
		if propertyName != "" && normalizeComparableName(propertyName) == normalizeComparableName(targetName) {
			return true
		}
		name := bindingElement.Name()
		if name != nil && importedNameMatches(targetName, name) {
			return true
		}
	}
	return false
}

func nameImportedFromPackage(sourceFile *ast.SourceFile, name string, pkg string) bool {
	if sourceFile == nil || name == "" || pkg == "" {
		return false
	}
	if sourceFile.Statements == nil {
		return false
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if statement == nil {
			continue
		}
		if statement.Kind == ast.KindImportDeclaration {
			importDeclaration := statement.AsImportDeclaration()
			if importDeclaration == nil || importDeclaration.ModuleSpecifier == nil {
				continue
			}
			if moduleSpecifierText(importDeclaration.ModuleSpecifier) != pkg {
				continue
			}
			if staticImportContainsName(importDeclaration, name) {
				return true
			}
			continue
		}
		if walkAst(statement, func(node *ast.Node) bool {
			if node == nil || node.Kind != ast.KindVariableDeclaration {
				return false
			}
			variableDeclaration := node.AsVariableDeclaration()
			if variableDeclaration == nil || variableDeclaration.Initializer == nil || variableDeclaration.Name() == nil {
				return false
			}
			if !isImportCallFromPackage(variableDeclaration.Initializer, pkg) {
				return false
			}
			return dynamicImportBindingContainsName(variableDeclaration.Name(), name)
		}) {
			return true
		}
	}
	// Fallback for parser edge cases where import binding nodes are not exposed as expected.
	sourceText := sourceFile.Text()
	if sourceText == "" {
		return false
	}
	namePattern := regexp.QuoteMeta(name)
	pkgPattern := regexp.QuoteMeta(pkg)
	staticImportPattern := regexp.MustCompile(`(?s)import\s*\{[^}]*\b` + namePattern + `\b[^}]*\}\s*from\s*['"]` + pkgPattern + `['"]`)
	if staticImportPattern.MatchString(sourceText) {
		return true
	}
	dynamicImportPattern := regexp.MustCompile(`(?s)\{[^}]*\b` + namePattern + `\b[^}]*\}\s*=\s*(?:await\s+)?import\(\s*['"]` + pkgPattern + `['"]\s*\)`)
	return dynamicImportPattern.MatchString(sourceText)
}

func allowEntryMatches(entry noDeprecatedAllowEntry, diagnosticName string, symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	if entry.Name != "" {
		if normalizeComparableName(diagnosticName) == normalizeComparableName(entry.Name) {
			// direct match
		} else if symbol != nil {
			hierarchy := symbolHierarchyNames(symbol)
			nameMatched := false
			for hierarchyName := range hierarchy {
				if normalizeComparableName(hierarchyName) == normalizeComparableName(entry.Name) {
					nameMatched = true
					break
				}
			}
			if !nameMatched {
				return false
			}
		} else {
			return false
		}
	}

	switch entry.From {
	case "":
		return true
	case "file":
		if symbol == nil {
			return entry.Name != "" && normalizeComparableName(diagnosticName) == normalizeComparableName(entry.Name)
		}
		return declarationInCurrentFile(symbol, sourceFile)
	case "package":
		if packageMatchesSymbol(entry.Package, symbol) {
			return true
		}
		return nameImportedFromPackage(sourceFile, entry.Name, entry.Package)
	default:
		return false
	}
}

func shouldAllowDiagnostic(entries []noDeprecatedAllowEntry, diagnosticName string, symbol *ast.Symbol, sourceFile *ast.SourceFile) bool {
	for _, entry := range entries {
		if allowEntryMatches(entry, diagnosticName, symbol, sourceFile) {
			return true
		}
	}
	return false
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

func getCallLikeDeprecation(ctx rule.RuleContext, node *ast.Node) (bool, string) {
	if ctx.TypeChecker == nil || node == nil || node.Parent == nil {
		return false, ""
	}
	signature := checker.Checker_getResolvedSignature(ctx.TypeChecker, node.Parent, nil, checker.CheckModeNormal)
	if signature == nil {
		return false, ""
	}
	signatureDecl := signature.Declaration()
	if signatureDecl != nil && (ctx.TypeChecker.IsDeprecatedDeclaration(signatureDecl) || hasDeprecatedTag(signatureDecl)) {
		reason := getJsDocDeprecationFromNode(signatureDecl)
		return true, reason
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false, ""
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
		return searchForDeprecationInAliasesChain(ctx.TypeChecker, symbol, true)
	}
	isDeprecated, reason := searchForDeprecationInAliasesChain(ctx.TypeChecker, symbol, false)
	if isDeprecated {
		return true, reason
	}
	if signatureDecl == nil && aliasedSymbol != nil {
		return getJsDocDeprecation(ctx.TypeChecker, aliasedSymbol)
	}
	return false, ""
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

func bindingElementPropertyNameFromNode(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindBindingElement {
		return ""
	}
	bindingElement := node.AsBindingElement()
	return bindingElementPropertyName(bindingElement)
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

func getDeprecationReason(ctx rule.RuleContext, node *ast.Node) (bool, string) {
	if ctx.TypeChecker == nil || node == nil {
		return false, ""
	}
	callLikeNode := getCallLikeNode(node)
	if callLikeNode != nil {
		return getCallLikeDeprecation(ctx, callLikeNode)
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute && node.Kind != ast.KindSuperKeyword {
		if node.Parent.Parent != nil && node.Parent.Parent.Parent != nil {
			return getJsxAttributeDeprecation(ctx, node.Parent.Parent.Parent, node.Text())
		}
	}
	if node.Parent != nil && node.Kind != ast.KindSuperKeyword {
		parent := node.Parent
		if parent.Kind == ast.KindBindingElement {
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
					bindingElement := parent.AsBindingElement()
					bindingNode := node
					if bindingElement != nil && bindingElement.PropertyName != nil {
						bindingNode = bindingElement.PropertyName
					}
					propertyName := ""
					if bindingPattern.Kind == ast.KindArrayBindingPattern {
						if bindingElement != nil {
							if index, ok := bindingElementIndex(bindingElement); ok {
								propertyName = strconv.Itoa(index)
							}
						}
					} else {
						if bindingName := bindingElementPropertyNameFromNode(parent); bindingName != "" {
							propertyName = bindingName
						}
						if propertyName == "" && bindingElement != nil && bindingElement.PropertyName != nil {
							if resolvedName, ok := resolveConstantPropertyName(ctx, bindingElement.PropertyName, 0, map[*ast.Symbol]bool{}); ok {
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
		// Ignore only direct default access on the import result.
		parent := access.AsNode().Parent
		if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
			parentAccess := parent.AsPropertyAccessExpression()
			if parentAccess != nil && parentAccess.Expression == access.AsNode() && parentAccess.Name() != nil && parentAccess.Name().Text() == "default" {
				return false
			}
		}
		return true
	}
	return false
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
		return true
	}
	return isDynamicImportResultIdentifier(typeChecker.GetSymbolAtLocation(access.Name()))
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
				if nameNode.Pos() == pos && nameNode.End() == end {
					return true
				}
			}
		case ast.KindImportClause:
			clause := current.AsImportClause()
			if clause != nil && clause.Name() != nil {
				nameNode := clause.Name()
				if nameNode.Pos() == pos && nameNode.End() == end {
					return true
				}
			}
		case ast.KindNamespaceImport:
			namespaceImport := current.AsNamespaceImport()
			if namespaceImport != nil && namespaceImport.Name() != nil {
				nameNode := namespaceImport.Name()
				if nameNode.Pos() == pos && nameNode.End() == end {
					return true
				}
			}
		case ast.KindImportEqualsDeclaration:
			importEquals := current.AsImportEqualsDeclaration()
			if importEquals != nil && importEquals.Name() != nil {
				nameNode := importEquals.Name()
				if nameNode.Pos() == pos && nameNode.End() == end {
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
	trimmedLine := strings.TrimSpace(lineText)
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
	if symbol.ValueDeclaration != nil && (typeChecker.IsDeprecatedDeclaration(symbol.ValueDeclaration) || hasDeprecatedTag(symbol.ValueDeclaration) || hasDeprecatedTagInSource(symbol.ValueDeclaration)) {
		return true
	}
	if len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || (!typeChecker.IsDeprecatedDeclaration(declaration) && !hasDeprecatedTag(declaration) && !hasDeprecatedTagInSource(declaration)) {
			return false
		}
	}
	return true
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

func resolveConstantPropertyName(ctx rule.RuleContext, node *ast.Node, depth int, seen map[*ast.Symbol]bool) (string, bool) {
	if ctx.TypeChecker == nil || node == nil || depth > maxConstantPropertyResolveDepth {
		return "", false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		stringLiteral := node.AsStringLiteral()
		if stringLiteral == nil {
			return "", false
		}
		return stringLiteral.Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		templateLiteral := node.AsNoSubstitutionTemplateLiteral()
		if templateLiteral == nil {
			return "", false
		}
		return templateLiteral.Text, true
	case ast.KindNumericLiteral:
		numericLiteral := node.AsNumericLiteral()
		if numericLiteral == nil {
			return "", false
		}
		return numericLiteral.Text, true
	case ast.KindAsExpression:
		asExpression := node.AsAsExpression()
		if asExpression == nil {
			return "", false
		}
		return resolveConstantPropertyName(ctx, asExpression.Expression, depth+1, seen)
	case ast.KindTypeAssertionExpression:
		typeAssertion := node.AsTypeAssertion()
		if typeAssertion == nil {
			return "", false
		}
		return resolveConstantPropertyName(ctx, typeAssertion.Expression, depth+1, seen)
	case ast.KindPropertyAccessExpression:
		propertyAccess := node.AsPropertyAccessExpression()
		if propertyAccess == nil || propertyAccess.Name() == nil {
			return "", false
		}
		symbol := ctx.TypeChecker.GetSymbolAtLocation(propertyAccess.Name())
		if symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindEnumMember {
			enumMember := symbol.ValueDeclaration.AsEnumMember()
			if enumMember != nil {
				return resolveConstantPropertyName(ctx, enumMember.Initializer, depth+1, seen)
			}
		}
	case ast.KindIdentifier:
		symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
		if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration {
			return "", false
		}
		if seen[symbol] {
			return "", false
		}
		seen[symbol] = true
		defer delete(seen, symbol)

		variableDeclaration := symbol.ValueDeclaration.AsVariableDeclaration()
		if variableDeclaration == nil || variableDeclaration.Initializer == nil {
			return "", false
		}
		return resolveConstantPropertyName(ctx, variableDeclaration.Initializer, depth+1, seen)
	}
	if constantValue := ctx.TypeChecker.GetConstantValue(node); constantValue != nil {
		if text, ok := constantValue.(string); ok {
			return text, true
		}
		switch value := constantValue.(type) {
		case float64:
			return strconv.FormatFloat(value, 'f', -1, 64), true
		case int:
			return strconv.Itoa(value), true
		case int32:
			return strconv.Itoa(int(value)), true
		case int64:
			return strconv.FormatInt(value, 10), true
		}
	}
	return "", false
}

func elementAccessPropertyName(ctx rule.RuleContext, argument *ast.Node) (string, bool) {
	return resolveConstantPropertyName(ctx, argument, 0, map[*ast.Symbol]bool{})
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
		if !ctx.TypeChecker.IsDeprecatedDeclaration(node) && !hasDeprecatedTag(node) && !hasDeprecatedTagInSource(node) {
			return false
		}
		found = true
		if reason == "" {
			reason = deprecatedReasonFromDeclaration(node)
		}
		return reason != ""
	})
	return found, reason
}

func deprecatedReasonByNameInSource(ctx rule.RuleContext, name string) string {
	_, reason := deprecatedInfoByNameInSource(ctx, name, false)
	return reason
}

func deprecatedPropertyInfoByNameInSource(ctx rule.RuleContext, name string) (bool, string) {
	return deprecatedInfoByNameInSource(ctx, name, true)
}

var NoDeprecatedRule = rule.CreateRule(rule.Rule{
	Name: "no-deprecated",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		allowEntries := parseAllowEntries(options)
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
					return bindingElement.PropertyName != nil
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
		reported := map[string]bool{}
		reportRange := func(diagnosticRange core.TextRange, message rule.RuleMessage) {
			key := strconv.Itoa(diagnosticRange.Pos()) + ":" + strconv.Itoa(diagnosticRange.End()) + ":" + message.Id
			if reported[key] {
				return
			}
			reported[key] = true
			ctx.ReportRange(diagnosticRange, message)
		}
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
				symbol := symbolAtLocation(ctx.TypeChecker, node)
				if isWithinJsxClosingElement(node, diagnostic.Pos(), diagnostic.End()) {
					continue
				}
				if isImportBindingAtRange(node, diagnostic.Pos(), diagnostic.End()) {
					continue
				}
				if isInImportStatementRange(sourceFile, diagnostic.Pos()) {
					continue
				}
				diagnosticRange := core.NewTextRange(diagnostic.Pos(), diagnostic.End())
				if promotedRange := promotedDynamicImportDefaultRange(node, diagnostic.Pos(), diagnostic.End(), name, ctx.TypeChecker); promotedRange != nil {
					diagnosticRange = *promotedRange
				}
				if shouldAllowDiagnostic(allowEntries, name, symbol, ctx.SourceFile) {
					continue
				}
				message := buildDeprecatedMessage(name)
				if reason := deprecatedReasonFromDiagnostic(diagnostic); reason != "" {
					message = buildDeprecatedWithReasonMessage(name, reason)
				} else if symbol != nil {
					for _, declaration := range symbol.Declarations {
						if declaration == nil {
							continue
						}
						if reason := deprecatedReasonFromDeclaration(declaration); reason != "" {
							message = buildDeprecatedWithReasonMessage(name, reason)
							break
						}
					}
				}
				if message.Id != "deprecatedWithReason" {
					if reason := deprecatedReasonByNameInSource(ctx, name); reason != "" {
						message = buildDeprecatedWithReasonMessage(name, reason)
					}
				}
				reportRange(diagnosticRange, message)
			}
		}
		checkIdentifier := func(node *ast.Node) {
			if node == nil {
				return
			}
			if isDeclaration(node) || isInsideImport(node) {
				return
			}
			isDeprecated, reason := getDeprecationReason(ctx, node)
			if !isDeprecated {
				return
			}
			symbol := symbolAtLocation(ctx.TypeChecker, node)
			if symbol == nil {
				symbol = ctx.TypeChecker.GetSymbolAtLocation(node)
			}
			if symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindBindingElement {
				symbol = nil
			}
			// Only report if the deprecated symbol isn't just a local declaration in the current file.
			// Allow reporting for deprecated locals (for example, symbol usage).
			name := getReportedNodeName(node)
			if shouldAllowDiagnostic(allowEntries, name, symbol, ctx.SourceFile) {
				return
			}
			message := buildDeprecatedMessage(name)
			if reason != "" {
				message = buildDeprecatedWithReasonMessage(name, reason)
			} else if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if declaration == nil {
						continue
					}
					if reasonFromDecl := deprecatedReasonFromDeclaration(declaration); reasonFromDecl != "" {
						message = buildDeprecatedWithReasonMessage(name, reasonFromDecl)
						break
					}
				}
			}
			if message.Id != "deprecatedWithReason" {
				if reasonByName := deprecatedReasonByNameInSource(ctx, name); reasonByName != "" {
					message = buildDeprecatedWithReasonMessage(name, reasonByName)
				}
			}
			trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			reportRange(core.NewTextRange(trimmedRange.Pos(), trimmedRange.End()), message)
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
						jsdocs := node.Parent.JSDoc(nil)
						for _, jsdoc := range jsdocs {
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
			ast.KindPrivateIdentifier: checkIdentifier,
			ast.KindSuperKeyword:      checkIdentifier,
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
				propertySymbol := symbolAtLocation(ctx.TypeChecker, nameNode)
				if propertySymbol == nil && node.Parent != nil && node.Parent.Kind == ast.KindJsxAttributes {
					attributesType := ctx.TypeChecker.GetTypeAtLocation(node.Parent)
					if attributesType != nil {
						propertySymbol = checker.Checker_getPropertyOfType(ctx.TypeChecker, attributesType, nameText)
					}
				}
				isDeprecated := symbolIsDeprecated(ctx.TypeChecker, propertySymbol)
				sourceDeprecated, sourceReason := deprecatedPropertyInfoByNameInSource(ctx, nameText)
				if !isDeprecated && !sourceDeprecated {
					return
				}
				if shouldAllowDiagnostic(allowEntries, nameText, propertySymbol, ctx.SourceFile) {
					return
				}
				message := buildDeprecatedMessage(nameText)
				if propertySymbol != nil {
					for _, declaration := range propertySymbol.Declarations {
						if reason := deprecatedReasonFromDeclaration(declaration); reason != "" {
							message = buildDeprecatedWithReasonMessage(nameText, reason)
							break
						}
					}
				}
				if message.Id != "deprecatedWithReason" && sourceReason != "" {
					message = buildDeprecatedWithReasonMessage(nameText, sourceReason)
				}
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				reportRange(core.NewTextRange(trimmedRange.Pos(), trimmedRange.End()), message)
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
				objectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, elementAccess.Expression)
				if objectType == nil {
					return
				}
				propertySymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, objectType, propertyName)
				if !symbolIsDeprecated(ctx.TypeChecker, propertySymbol) {
					return
				}
				if shouldAllowDiagnostic(allowEntries, propertyName, propertySymbol, ctx.SourceFile) {
					return
				}
				message := buildDeprecatedMessage(propertyName)
				for _, declaration := range propertySymbol.Declarations {
					if reason := deprecatedReasonFromDeclaration(declaration); reason != "" {
						message = buildDeprecatedWithReasonMessage(propertyName, reason)
						break
					}
				}
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, elementAccess.ArgumentExpression)
				reportRange(core.NewTextRange(trimmedRange.Pos(), trimmedRange.End()), message)
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				if access == nil || access.Name() == nil {
					return
				}
				// Report the property name if it is deprecated.
				nameNode := access.Name()
				propertySymbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				if !symbolIsDeprecated(ctx.TypeChecker, propertySymbol) {
					return
				}
				name := nameNode.Text()
				if shouldAllowDiagnostic(allowEntries, name, propertySymbol, ctx.SourceFile) {
					return
				}
				message := buildDeprecatedMessage(name)
				for _, declaration := range propertySymbol.Declarations {
					if reason := deprecatedReasonFromDeclaration(declaration); reason != "" {
						message = buildDeprecatedWithReasonMessage(name, reason)
						break
					}
				}
				trimmedRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				reportRange(core.NewTextRange(trimmedRange.Pos(), trimmedRange.End()), message)
			},
		}
	},
})
