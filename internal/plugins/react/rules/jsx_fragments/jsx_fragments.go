package jsx_fragments

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	modeSyntax  = "syntax"
	modeElement = "element"
)

var JsxFragmentsRule = rule.Rule{
	Name: "react/jsx-fragments",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		mode := parseMode(rawOptions)
		reactPragma := reactutil.GetReactPragma(ctx.Settings)
		fragmentPragma := reactutil.GetReactFragmentPragma(ctx.Settings)
		openFragLong := "<" + reactPragma + "." + fragmentPragma + ">"
		closeFragLong := "</" + reactPragma + "." + fragmentPragma + ">"
		matcher := newFragmentMatcher(ctx, reactPragma, fragmentPragma)

		reportOnReactVersion := func(node *ast.Node) bool {
			if !reactutil.ReactVersionLessThan(ctx.Settings, 16, 2, 0) {
				return false
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id: "fragmentsNotSupported",
				Description: "Fragments are only supported starting from React v16.2. " +
					"Please disable the `react/jsx-fragments` rule in `eslint` settings or upgrade your version of React.",
			})
			return true
		}

		preferPragmaMsg := func() rule.RuleMessage {
			return rule.RuleMessage{
				Id:          "preferPragma",
				Description: "Prefer " + reactPragma + "." + fragmentPragma + " over fragment shorthand",
				Data: map[string]string{
					"react":    reactPragma,
					"fragment": fragmentPragma,
				},
			}
		}

		preferFragmentMsg := func() rule.RuleMessage {
			return rule.RuleMessage{
				Id:          "preferFragment",
				Description: "Prefer fragment shorthand over " + reactPragma + "." + fragmentPragma,
				Data: map[string]string{
					"react":    reactPragma,
					"fragment": fragmentPragma,
				},
			}
		}

		checkFragment := func(node *ast.Node) {
			if reportOnReactVersion(node) {
				return
			}
			if mode != modeElement {
				return
			}
			fragment := node.AsJsxFragment()
			if fragment.OpeningFragment == nil || fragment.ClosingFragment == nil {
				ctx.ReportNode(node, preferPragmaMsg())
				return
			}
			ctx.ReportNodeWithFixes(node, preferPragmaMsg(),
				rule.RuleFixReplace(ctx.SourceFile, fragment.OpeningFragment, openFragLong),
				rule.RuleFixReplace(ctx.SourceFile, fragment.ClosingFragment, closeFragLong),
			)
		}

		checkStandardElement := func(node *ast.Node) {
			opening := openingElement(node)
			if opening == nil || !matcher.isReactFragment(opening) {
				return
			}
			if reportOnReactVersion(node) {
				return
			}
			if mode != modeSyntax || len(reactutil.GetJsxElementAttributes(opening)) > 0 {
				return
			}
			fixes := fixesToShort(ctx.SourceFile, node)
			if len(fixes) == 0 {
				ctx.ReportNode(node, preferFragmentMsg())
				return
			}
			ctx.ReportNodeWithFixes(node, preferFragmentMsg(), fixes...)
		}

		return rule.RuleListeners{
			ast.KindJsxFragment:           checkFragment,
			ast.KindJsxElement:            checkStandardElement,
			ast.KindJsxSelfClosingElement: checkStandardElement,
		}
	},
}

func parseMode(raw any) string {
	options := rule.NormalizeOptions(raw)
	if len(options) == 0 {
		return modeSyntax
	}
	if mode, ok := options[0].(string); ok && (mode == modeSyntax || mode == modeElement) {
		return mode
	}
	return modeSyntax
}

func openingElement(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindJsxElement:
		return node.AsJsxElement().OpeningElement
	case ast.KindJsxSelfClosingElement:
		return node
	default:
		return nil
	}
}

func fixesToShort(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	switch node.Kind {
	case ast.KindJsxElement:
		element := node.AsJsxElement()
		if element.OpeningElement == nil || element.ClosingElement == nil {
			return nil
		}
		return []rule.RuleFix{
			rule.RuleFixReplace(sourceFile, element.OpeningElement, "<>"),
			rule.RuleFixReplace(sourceFile, element.ClosingElement, "</>"),
		}
	case ast.KindJsxSelfClosingElement:
		return []rule.RuleFix{
			rule.RuleFixReplace(sourceFile, node, "<></>"),
		}
	default:
		return nil
	}
}

type fragmentMatcher struct {
	ctx            rule.RuleContext
	reactPragma    string
	fragmentPragma string
	fragmentNames  map[string]bool
}

func newFragmentMatcher(ctx rule.RuleContext, reactPragma, fragmentPragma string) fragmentMatcher {
	m := fragmentMatcher{
		ctx:            ctx,
		reactPragma:    reactPragma,
		fragmentPragma: fragmentPragma,
		fragmentNames:  map[string]bool{reactPragma + "." + fragmentPragma: true},
	}
	// Mirrors upstream's file-level `fragmentNames` set. It deliberately
	// tracks import aliases by text and not by scope, so a later shadowing
	// binding with the same JSX tag name still matches the import path.
	if ctx.SourceFile != nil {
		m.collectImportFragmentNames(ctx.SourceFile.AsNode())
	}
	return m
}

func (m fragmentMatcher) isReactFragment(opening *ast.Node) bool {
	elementName := reactutil.GetJsxElementTypeString(opening)
	if m.fragmentNames[elementName] {
		return true
	}

	// The fallback mirrors upstream's `refersToReactFragment` branch for bare
	// JSX identifiers whose variable initializer points at the React fragment.
	tagName := reactutil.GetJsxTagName(opening)
	if tagName == nil || tagName.Kind != ast.KindIdentifier {
		return false
	}
	return m.refersToReactFragment(tagName, tagName.AsIdentifier().Text)
}

func (m fragmentMatcher) collectImportFragmentNames(root *ast.Node) {
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindImportDeclaration {
			m.collectImportDeclaration(node)
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(root)
}

func (m fragmentMatcher) collectImportDeclaration(node *ast.Node) {
	decl := node.AsImportDeclaration()
	if decl.ModuleSpecifier == nil ||
		decl.ModuleSpecifier.Kind != ast.KindStringLiteral ||
		decl.ModuleSpecifier.AsStringLiteral().Text != "react" ||
		decl.ImportClause == nil {
		return
	}
	importClause := decl.ImportClause.AsImportClause()
	if importClause.NamedBindings == nil || importClause.NamedBindings.Kind != ast.KindNamedImports {
		return
	}
	namedImports := importClause.NamedBindings.AsNamedImports()
	if namedImports.Elements == nil {
		return
	}
	for _, specNode := range namedImports.Elements.Nodes {
		spec := specNode.AsImportSpecifier()
		imported := spec.Name()
		if spec.PropertyName != nil {
			imported = spec.PropertyName
		}
		if imported == nil || imported.Kind != ast.KindIdentifier || imported.AsIdentifier().Text != m.fragmentPragma {
			continue
		}
		local := spec.Name()
		if local != nil && local.Kind == ast.KindIdentifier {
			m.fragmentNames[local.AsIdentifier().Text] = true
		}
	}
}

func (m fragmentMatcher) refersToReactFragment(ident *ast.Node, name string) bool {
	if ident == nil || name == "" {
		return false
	}
	if m.ctx.TypeChecker != nil {
		if decl := declarationForIdentifier(utils.GetReferenceSymbol(ident, m.ctx.TypeChecker)); decl != nil {
			return m.declarationRefersToReactFragment(decl)
		}
	}
	if m.ctx.SourceFile == nil {
		return false
	}
	return m.syntaxDeclarationRefersToReactFragment(m.ctx.SourceFile.AsNode(), name)
}

func declarationForIdentifier(symbol *ast.Symbol) *ast.Node {
	if symbol == nil {
		return nil
	}
	if symbol.ValueDeclaration != nil {
		return symbol.ValueDeclaration
	}
	if len(symbol.Declarations) > 0 {
		return symbol.Declarations[0]
	}
	return nil
}

func (m fragmentMatcher) declarationRefersToReactFragment(decl *ast.Node) bool {
	switch decl.Kind {
	case ast.KindVariableDeclaration:
		return m.initializerRefersToReactFragment(decl.AsVariableDeclaration().Initializer)
	case ast.KindBindingElement:
		root := ast.GetRootDeclaration(decl)
		if root == nil || root.Kind != ast.KindVariableDeclaration {
			return false
		}
		return m.initializerRefersToReactFragment(root.AsVariableDeclaration().Initializer)
	default:
		return false
	}
}

func (m fragmentMatcher) syntaxDeclarationRefersToReactFragment(root *ast.Node, name string) bool {
	var found bool
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found || node == nil {
			return
		}
		if node.Kind == ast.KindVariableDeclaration && variableDeclarationBindsName(node, name) {
			found = m.initializerRefersToReactFragment(node.AsVariableDeclaration().Initializer)
			if found {
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(root)
	return found
}

func variableDeclarationBindsName(node *ast.Node, name string) bool {
	declName := node.AsVariableDeclaration().Name()
	if declName == nil {
		return false
	}
	found := false
	utils.CollectBindingNames(declName, func(_ *ast.Node, bindingName string) {
		if bindingName == name {
			found = true
		}
	})
	return found
}

func (m fragmentMatcher) initializerRefersToReactFragment(init *ast.Node) bool {
	init = ast.SkipParentheses(init)
	if init == nil {
		return false
	}

	if ast.IsOptionalChain(init) {
		return false
	}

	if init.Kind == ast.KindIdentifier && init.AsIdentifier().Text == m.reactPragma {
		return true
	}

	// Match only `<pragma>.<fragment>`, not `require('react').Fragment` or an
	// arbitrary pragma member. Only parentheses are transparent here; upstream
	// does not unwrap TS-only expression wrappers like `as`, `satisfies`, or `!`.
	if init.Kind == ast.KindPropertyAccessExpression {
		access := init.AsPropertyAccessExpression()
		property := access.Name()
		object := ast.SkipParentheses(access.Expression)
		if object != nil &&
			object.Kind == ast.KindIdentifier &&
			object.AsIdentifier().Text == m.reactPragma &&
			property != nil &&
			property.Kind == ast.KindIdentifier &&
			property.AsIdentifier().Text == m.fragmentPragma {
			return true
		}
	}

	return isRequireReactCall(init)
}

func isRequireReactCall(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	if ast.IsOptionalChain(node) {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "require" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	// Do not use ast.IsRequireCall here: upstream only checks the first
	// argument and therefore accepts unusual calls like require('react', x).
	arg := ast.SkipParentheses(call.Arguments.Nodes[0])
	return arg != nil && arg.Kind == ast.KindStringLiteral && arg.AsStringLiteral().Text == "react"
}
