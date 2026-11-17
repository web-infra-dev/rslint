package no_deprecated

// cspell:ignore inheritdoc

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_deprecated.schema.json
var schemaJSON []byte

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
	for callee != nil && callee.Parent != nil {
		parent := callee.Parent
		if parent.Kind == ast.KindParenthesizedExpression ||
			(parent.Kind == ast.KindPropertyAccessExpression && parent.Name() == callee) {
			callee = parent
			continue
		}
		break
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

func parseAllowSpecifiers(options []any) []utils.TypeOrValueSpecifier {
	if len(options) == 0 {
		return nil
	}
	optsMap, _ := options[0].(map[string]any)
	return utils.ParseTypeOrValueSpecifiers(optsMap["allow"])
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

// TypeScript attaches documentation to a declaration's JSDoc host. In
// particular, only the first declarator inherits its variable statement's
// documentation. Parse the actual tags instead of searching nearby source.
// Tag presence matters: an ordinary local tag blocks inherited tags, while
// @inheritDoc explicitly places inherited tags before the local ones.
type deprecationInfo struct {
	deprecated bool
	reason     string
	hasTags    bool
	inherit    bool
}

func ownDeprecation(c *checker.Checker, node *ast.Node) deprecationInfo {
	var result deprecationInfo
	for current := node; current != nil; current = ast.GetNextJSDocCommentLocation(current) {
		docs := current.JSDoc(nil)
		if len(docs) == 0 {
			continue
		}
		// Only the last JSDoc block owns ordinary tags; earlier blocks can carry
		// JSDoc overloads, but cannot contribute a deprecated tag to this host.
		tags := docs[len(docs)-1].AsJSDoc().Tags
		if tags == nil {
			continue
		}
		result.hasTags = result.hasTags || len(tags.Nodes) > 0
		for _, tag := range tags.Nodes {
			if tag.TagName().Text() == "inheritDoc" || tag.TagName().Text() == "inheritdoc" {
				result.inherit = true
			}
			if !result.deprecated && ast.IsJSDocDeprecatedTag(tag) {
				result.deprecated = true
				result.reason = jsDocCommentText(c, tag.AsJSDocDeprecatedTag().Comment)
			}
		}
	}
	return result
}

func inheritedDeprecation(c *checker.Checker, declaration *ast.Node, seen map[*ast.Node]bool) deprecationInfo {
	parent := declaration.Parent
	if parent != nil && parent.Kind == ast.KindConstructor {
		parent = parent.Parent
	}
	if parent == nil || declaration.Symbol() == nil {
		return deprecationInfo{}
	}
	var clauses *ast.NodeList
	switch parent.Kind {
	case ast.KindClassDeclaration:
		clauses = parent.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		clauses = parent.AsClassExpression().HeritageClauses
	case ast.KindInterfaceDeclaration:
		clauses = parent.AsInterfaceDeclaration().HeritageClauses
	}
	if clauses == nil {
		return deprecationInfo{}
	}
	for _, clause := range clauses.Nodes {
		for _, baseNode := range clause.AsHeritageClause().Types.Nodes {
			typ := c.GetTypeAtLocation(baseNode)
			if typ == nil {
				continue
			}
			if ast.HasStaticModifier(declaration) && typ.Symbol() != nil {
				typ = c.GetTypeOfSymbolAtLocation(typ.Symbol(), declaration)
			}
			symbol := checker.Checker_getPropertyOfType(c, typ, declaration.Symbol().Name)
			if symbol == nil {
				continue
			}
			declarations := symbol.Declarations
			if declaration.Kind == ast.KindGetAccessor || declaration.Kind == ast.KindSetAccessor {
				var contextual []*ast.Node
				for _, d := range declarations {
					if d.Kind == declaration.Kind {
						contextual = append(contextual, d)
					}
				}
				info := declarationsDeprecation(c, contextual, seen)
				if info.hasTags {
					return info
				}
			} else if len(declarations) != 1 {
				continue
			}
			return declarationsDeprecation(c, declarations, seen)
		}
	}
	return deprecationInfo{}
}

func declarationsDeprecation(c *checker.Checker, declarations []*ast.Node, seen map[*ast.Node]bool) deprecationInfo {
	var result deprecationInfo
	var active []*ast.Node
	for _, declaration := range declarations {
		if declaration == nil || seen[declaration] {
			continue
		}
		seen[declaration] = true
		active = append(active, declaration)
		info := ownDeprecation(c, declaration)
		result.hasTags = result.hasTags || info.hasTags
		result.inherit = result.inherit || info.inherit
		if !result.deprecated && info.deprecated {
			result.deprecated, result.reason = true, info.reason
		}
	}
	if !result.hasTags || result.inherit {
		for _, declaration := range active {
			info := inheritedDeprecation(c, declaration, seen)
			result.hasTags = result.hasTags || info.hasTags
			if info.deprecated {
				result.deprecated, result.reason = true, info.reason
			}
		}
	}
	for _, declaration := range active {
		delete(seen, declaration)
	}
	return result
}

func declarationDeprecation(c *checker.Checker, node *ast.Node) (bool, string) {
	info := declarationsDeprecation(c, []*ast.Node{node}, map[*ast.Node]bool{})
	return info.deprecated, info.reason
}

func symbolDeprecation(c *checker.Checker, symbol *ast.Symbol) (bool, string) {
	if symbol == nil {
		return false, ""
	}
	info := declarationsDeprecation(c, symbol.Declarations, map[*ast.Node]bool{})
	return info.deprecated, info.reason
}

// An export alias can carry its own tag, independently of the final target.
// Use cycle detection rather than an arbitrary depth limit for re-exports.
func aliasDeprecation(c *checker.Checker, symbol *ast.Symbol, includeTarget bool) (bool, string) {
	seen := map[*ast.Symbol]bool{}
	for symbol != nil && !seen[symbol] {
		seen[symbol] = true
		if symbol.Flags&ast.SymbolFlagsAlias == 0 {
			if includeTarget {
				return symbolDeprecation(c, symbol)
			}
			return false, ""
		}
		if deprecated, reason := symbolDeprecation(c, symbol); deprecated {
			return true, reason
		}
		if len(symbol.Declarations) == 0 {
			break
		}
		symbol = c.GetImmediateAliasedSymbol(symbol)
	}
	return false, ""
}

func callDeprecation(c *checker.Checker, callee *ast.Node) (bool, string) {
	signature := checker.Checker_getResolvedSignature(c, callee.Parent, nil, checker.CheckModeNormal)
	symbol := c.GetSymbolAtLocation(ast.SkipParentheses(callee))
	target := symbol
	if target != nil && target.Flags&ast.SymbolFlagsAlias != 0 {
		target = c.GetAliasedSymbol(target)
	}
	var kind ast.Kind
	if target != nil && len(target.Declarations) > 0 {
		kind = target.Declarations[0].Kind
	}
	// Function and method declarations collect tags across all overloads.
	// Calls must instead use the selected signature, after export alias tags.
	includeTarget := kind != ast.KindMethodDeclaration && kind != ast.KindFunctionDeclaration && kind != ast.KindMethodSignature
	if deprecated, reason := aliasDeprecation(c, symbol, includeTarget); deprecated {
		return true, reason
	}
	if signature != nil {
		if deprecated, reason := declarationDeprecation(c, signature.Declaration()); deprecated {
			return true, reason
		}
	}
	if includeTarget {
		return symbolDeprecation(c, target)
	}
	return false, ""
}

func isDeclaration(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindBindingElement:
		binding := parent.AsBindingElement()
		if parent.Parent.Kind == ast.KindArrayBindingPattern || binding.DotDotDotToken != nil {
			return binding.Name() == node
		}
		// The single name in a shorthand binding also represents the property
		// reference; a renamed binding's local name is only a declaration.
		return binding.Name() == node && binding.PropertyName != nil
	case ast.KindClassExpression, ast.KindClassDeclaration, ast.KindVariableDeclaration, ast.KindEnumMember,
		ast.KindMethodDeclaration, ast.KindPropertyDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return parent.Name() == node
	case ast.KindPropertyAssignment:
		return parent.Name() == node
	case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindEnumDeclaration, ast.KindInterfaceDeclaration, ast.KindMethodSignature,
		ast.KindModuleDeclaration, ast.KindParameter, ast.KindPropertySignature,
		ast.KindTypeAliasDeclaration, ast.KindTypeParameter:
		return true
	case ast.KindImportEqualsDeclaration:
		return parent.Name() == node
	}
	return false
}

func isInsideImport(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindImportDeclaration:
			return true
		case ast.KindSourceFile, ast.KindBlock, ast.KindFunctionDeclaration,
			ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration, ast.KindVariableDeclaration,
			ast.KindExportDeclaration, ast.KindUnionType:
			return false
		}
	}
	return false
}

func identifierDeprecation(c *checker.Checker, node *ast.Node) (bool, string) {
	if callee := getCallLikeNode(node); callee != nil {
		return callDeprecation(c, callee)
	}
	parent := node.Parent
	if parent.Kind == ast.KindJsxAttribute {
		opening := parent.Parent.Parent
		contextual := checker.Checker_getContextualType(c, opening.TagName(), checker.ContextFlagsNone)
		if contextual != nil {
			return symbolDeprecation(c, checker.Checker_getPropertyOfType(c, contextual, node.Text()))
		}
		return false, ""
	}
	if parent.Kind == ast.KindBindingElement && parent.Parent.Kind == ast.KindObjectBindingPattern {
		// Binding pattern types retain the property symbols of the source,
		// including nested patterns and contextual parameter types.
		binding := parent.AsBindingElement()
		key := binding.PropertyName
		if key == nil {
			key = binding.Name()
		}
		if key == node {
			typ := c.GetTypeAtLocation(parent.Parent)
			if typ != nil {
				if deprecated, reason := symbolDeprecation(c, checker.Checker_getPropertyOfType(c, typ, node.Text())); deprecated {
					return true, reason
				}
			}
		}
	}
	symbol := c.GetSymbolAtLocation(node)
	if deprecated, reason := aliasDeprecation(c, symbol, true); deprecated {
		return true, reason
	}
	if parent.Kind == ast.KindShorthandPropertyAssignment {
		return aliasDeprecation(c, c.GetShorthandAssignmentValueSymbol(parent), true)
	}
	return false, ""
}

var NoDeprecatedRule = rule.CreateRule(rule.Rule{
	Name:             "no-deprecated",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		c := ctx.TypeChecker
		if c == nil {
			return rule.RuleListeners{}
		}
		allow := parseAllowSpecifiers(options)
		report := func(node *ast.Node, name, reason string) {
			message := buildDeprecatedMessage(name)
			if reason != "" {
				message = buildDeprecatedWithReasonMessage(name, reason)
			}
			ctx.ReportNode(node, message)
		}
		typeAllowed := func(typ *checker.Type) bool {
			return len(allow) > 0 && utils.TypeMatchesSomeSpecifier(typ, allow, nil, ctx.Program())
		}
		checkIdentifier := func(node *ast.Node) {
			if isDeclaration(node) || isInsideImport(node) {
				return
			}
			deprecated, reason := identifierDeprecation(c, node)
			if !deprecated {
				return
			}
			if len(allow) > 0 {
				typ := c.GetTypeAtLocation(node)
				if typeAllowed(typ) || utils.ValueMatchesSomeSpecifier(node, allow, ctx.Program(), typ) {
					return
				}
			}
			report(node, getReportedNodeName(node), reason)
		}
		checkProperty := func(node *ast.Node) {
			if node.Parent.Kind != ast.KindObjectLiteralExpression {
				return
			}
			key := node.Name()
			if key == nil {
				return
			}
			var name string
			var known bool
			if key.Kind == ast.KindComputedPropertyName {
				key = key.AsComputedPropertyName().Expression
				typ := c.GetTypeAtLocation(key)
				if typ != nil && checker.Type_flags(typ)&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral) != 0 {
					name, known = elementAccessPropertyName(ctx, key)
				}
			} else {
				name, known = utils.GetStaticPropertyName(key)
			}
			if !known {
				return
			}
			contextual := checker.Checker_getContextualType(c, node.Parent, checker.ContextFlagsNone)
			if contextual == nil {
				return
			}
			deprecated, reason := symbolDeprecation(c, checker.Checker_getPropertyOfType(c, contextual, name))
			if deprecated && !typeAllowed(contextual) {
				report(key, name, reason)
			}
		}
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil {
					return
				}
				if parent.Kind == ast.KindExportDeclaration || parent.Kind == ast.KindNamespaceExport || parent.Kind == ast.KindJsxClosingElement {
					return
				}
				if parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().ArgumentExpression == node {
					return
				}
				if parent.Kind == ast.KindExportSpecifier {
					if parent.AsExportSpecifier().PropertyName == node {
						return
					}
					if deprecated, _ := symbolDeprecation(c, c.GetSymbolAtLocation(node)); deprecated {
						return
					}
				}
				checkIdentifier(node)
			},
			ast.KindPrivateIdentifier: checkIdentifier,
			ast.KindSuperKeyword:      checkIdentifier,
			ast.KindElementAccessExpression: func(node *ast.Node) {
				access := node.AsElementAccessExpression()
				name, known := elementAccessPropertyName(ctx, access.ArgumentExpression)
				if !known {
					return
				}
				typ := c.GetTypeAtLocation(access.Expression)
				if typ == nil {
					return
				}
				deprecated, reason := symbolDeprecation(c, checker.Checker_getPropertyOfType(c, typ, name))
				if deprecated && !typeAllowed(typ) {
					report(access.ArgumentExpression, name, reason)
				}
			},
			ast.KindPropertyAssignment:          checkProperty,
			ast.KindShorthandPropertyAssignment: checkProperty,
			ast.KindMethodDeclaration:           checkProperty,
			ast.KindGetAccessor:                 checkProperty,
			ast.KindSetAccessor:                 checkProperty,
		}
	},
})
