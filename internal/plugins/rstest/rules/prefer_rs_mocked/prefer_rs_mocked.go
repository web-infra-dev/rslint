package prefer_rs_mocked

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const (
	namespaceRs     = "rs"
	namespaceRstest = "rstest"
)

var restrictedMockTypes = map[string]bool{
	"Mock":         true,
	"Mocked":       true,
	"MockInstance": true,
}

func preferRsMockedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferRsMocked",
		Description: "Prefer `rs.mocked()` over type assertions",
	}
}

func isRestrictedMockType(ctx rule.RuleContext, typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference || ctx.Refs == nil {
		return false
	}
	typeName := typeNode.AsTypeReferenceNode().TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier {
		return false
	}

	localName := typeName.AsIdentifier().Text
	originalName, _, mode := testFramework.ResolveTypeIdentifierReferenceFromSymbolModules(
		localName,
		typeName,
		ctx.Refs.Resolve(typeName),
		ctx.SourceFile,
		rstestUtils.RstestCoreImportModules,
	)
	return mode == testFramework.ReferenceModeImport && restrictedMockTypes[originalName]
}

type namespaceState struct {
	available bool
	imported  bool
	used      bool
	// localName is the name the file binds the namespace to, which an aliased
	// import makes something other than the canonical spelling.
	localName string
}

// fileNamespaces computes only the file-wide facts needed to choose a safe
// edit. The scan is deferred until a consumer asks for fixes or suggestions.
type fileNamespaces struct {
	ctx rule.RuleContext

	scanned       bool
	rs            namespaceState
	rstest        namespaceState
	hasImportMeta bool
}

func (file *fileNamespaces) scan() {
	if file.scanned {
		return
	}
	file.scanned = true
	file.rs = namespaceState{available: true, localName: namespaceRs}
	file.rstest = namespaceState{available: true, localName: namespaceRstest}
	if file.ctx.SourceFile == nil {
		return
	}
	file.collectImports()

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if rstestUtils.IsImportMetaRstest(node) {
			file.hasImportMeta = true
		}
		if node.Kind == ast.KindIdentifier {
			// The names watched are the ones the file binds, so an aliased
			// import is followed under its alias rather than under the
			// spelling it was imported from.
			text := node.AsIdentifier().Text
			if text == file.rs.localName {
				file.observeIdentifier(&file.rs, node)
			}
			if text == file.rstest.localName {
				file.observeIdentifier(&file.rstest, node)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(file.ctx.SourceFile.AsNode())
}

// collectImports records which namespaces the file imports from Rstest core and
// under which local name, because an edit has to write the name the file
// actually binds rather than the spelling the namespace is exported under.
func (file *fileNamespaces) collectImports() {
	for _, statement := range file.ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration == nil ||
			declaration.ModuleSpecifier == nil ||
			!rstestUtils.IsRstestCoreImportModule(declaration.ModuleSpecifier.Text()) ||
			declaration.ImportClause == nil ||
			declaration.ImportClause.IsTypeOnly() {
			continue
		}
		clause := declaration.ImportClause.AsImportClause()
		if clause == nil ||
			clause.NamedBindings == nil ||
			clause.NamedBindings.Kind != ast.KindNamedImports {
			continue
		}
		named := clause.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			continue
		}
		for _, element := range named.Elements.Nodes {
			file.collectImportSpecifier(element)
		}
	}
}

func (file *fileNamespaces) collectImportSpecifier(element *ast.Node) {
	specifier := element.AsImportSpecifier()
	if specifier == nil || ast.IsTypeOnlyImportDeclaration(element) {
		return
	}
	local := specifier.Name()
	if local == nil || local.Kind != ast.KindIdentifier {
		return
	}
	switch importedName(specifier) {
	case namespaceRs:
		file.rs.imported = true
		file.rs.localName = local.AsIdentifier().Text
	case namespaceRstest:
		file.rstest.imported = true
		file.rstest.localName = local.AsIdentifier().Text
	}
}

func (file *fileNamespaces) observeIdentifier(state *namespaceState, identifier *ast.Node) {
	if ast.IsDeclarationName(identifier) {
		// An Rstest utilities import binds the very namespace an edit reaches
		// for, so it does not make the name unavailable.
		if isRstestUtilitiesImportName(identifier) {
			return
		}
		if declaresBinding(identifier.Parent) {
			state.available = false
		}
		return
	}
	if !ast.IsIdentifierName(identifier) && !ast.IsDeclarationNameOrImportPropertyName(identifier) {
		state.used = true
	}
}

func isRstestUtilitiesImportName(identifier *ast.Node) bool {
	if identifier == nil || identifier.Parent == nil || identifier.Parent.Kind != ast.KindImportSpecifier {
		return false
	}
	specifier := identifier.Parent.AsImportSpecifier()
	if specifier == nil ||
		ast.IsTypeOnlyImportDeclaration(identifier.Parent) ||
		specifier.Name() != identifier {
		return false
	}
	declaration := testFramework.FindImportDeclaration(identifier.Parent)
	if declaration == nil || declaration.ModuleSpecifier == nil ||
		!rstestUtils.IsRstestCoreImportModule(declaration.ModuleSpecifier.Text()) {
		return false
	}
	name := importedName(specifier)
	return name == namespaceRs || name == namespaceRstest
}

// importedName is the name the specifier reaches for in the module it imports
// from. It is spelled by `PropertyName` when the import is aliased, and an
// alias may name the export with a string literal, so the text is read off the
// node rather than the node being required to be an identifier.
func importedName(specifier *ast.ImportSpecifier) string {
	if specifier == nil {
		return ""
	}
	if specifier.PropertyName != nil {
		return specifier.PropertyName.Text()
	}
	name := specifier.Name()
	if name == nil {
		return ""
	}
	return name.Text()
}

func declaresBinding(declaration *ast.Node) bool {
	if declaration == nil {
		return false
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration,
		ast.KindParameter,
		ast.KindBindingElement,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration,
		ast.KindImportClause,
		ast.KindImportSpecifier,
		ast.KindNamespaceImport,
		ast.KindImportEqualsDeclaration:
		return true
	default:
		return false
	}
}

func (file *fileNamespaces) fixNamespace() string {
	file.scan()
	rsImported := file.rs.available && file.rs.imported
	rstestImported := file.rstest.available && file.rstest.imported
	if rsImported && !rstestImported {
		return file.rs.localName
	}
	if rstestImported && !rsImported {
		return file.rstest.localName
	}
	if rsImported && rstestImported {
		return file.rs.localName
	}

	// With no utilities import, a free global spelling is usable only when both
	// standard names are unoccupied. If either is declared by unrelated code,
	// the file gives no evidence that globals mode is enabled, so introducing
	// the other spelling would be a guess.
	if !file.rs.available || !file.rstest.available {
		return ""
	}
	if file.rs.used && !file.rstest.used {
		return namespaceRs
	}
	if file.rstest.used && !file.rs.used {
		return namespaceRstest
	}
	return namespaceRs
}

func (file *fileNamespaces) canSuggestImportMeta() bool {
	file.scan()
	return file.fixNamespace() == "" && file.hasImportMeta
}

func replacementText(ctx rule.RuleContext, expression *ast.Node, namespace string) string {
	innerExpression := ast.SkipOuterExpressions(expression, ast.OEKParentheses|ast.OEKTypeAssertions)
	return namespace + ".mocked(" + utils.TrimmedNodeText(ctx.SourceFile, innerExpression) + ")"
}

func reportAssertion(ctx rule.RuleContext, file *fileNamespaces, assertionNode, expression *ast.Node) {
	message := preferRsMockedMessage()
	ctx.ReportNodeWithDeferredFixesAndSuggestions(
		assertionNode,
		message,
		func() []rule.RuleFix {
			namespace := file.fixNamespace()
			if namespace == "" {
				return nil
			}
			return []rule.RuleFix{
				rule.RuleFixReplace(ctx.SourceFile, assertionNode, replacementText(ctx, expression, namespace)),
			}
		},
		func() []rule.RuleSuggestion {
			if !file.canSuggestImportMeta() {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message: message,
				FixesArr: []rule.RuleFix{
					rule.RuleFixReplace(
						ctx.SourceFile,
						assertionNode,
						replacementText(ctx, expression, "import.meta.rstest.rs"),
					),
				},
			}}
		},
	)
}

func checkAssertion(ctx rule.RuleContext, file *fileNamespaces, node, typeNode, expression *ast.Node) {
	if node.Kind == ast.KindAsExpression {
		if parent := ast.WalkUpParenthesizedExpressions(node.Parent); parent != nil && parent.Kind == ast.KindAsExpression {
			return
		}
	}
	if !isRestrictedMockType(ctx, typeNode) {
		return
	}
	reportAssertion(ctx, file, node, expression)
}

var PreferRsMockedRule = rule.Rule{
	Name:   "rstest/prefer-rs-mocked",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		file := &fileNamespaces{ctx: ctx}
		return rule.RuleListeners{
			ast.KindAsExpression: func(node *ast.Node) {
				asExpression := node.AsAsExpression()
				if asExpression != nil {
					checkAssertion(ctx, file, node, asExpression.Type, asExpression.Expression)
				}
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				typeAssertion := node.AsTypeAssertion()
				if typeAssertion != nil {
					checkAssertion(ctx, file, node, typeAssertion.Type, typeAssertion.Expression)
				}
			},
		}
	},
}
