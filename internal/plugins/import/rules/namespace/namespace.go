package namespace

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

type ruleOptions struct {
	allowComputed bool
}

// NamespaceRule ensures imported namespaces contain dereferenced properties.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/namespace.js
var NamespaceRule = rule.Rule{
	Name: "import/namespace",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		namespaces := collectNamespaces(ctx)

		checkAccess := func(node *ast.Node) {
			checkNamespaceAccess(ctx, namespaces, opts, node)
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: checkAccess,
			ast.KindElementAccessExpression:  checkAccess,
			ast.KindVariableDeclaration: func(node *ast.Node) {
				checkNamespaceDestructuring(ctx, namespaces, node)
			},
			ast.KindNamespaceExport: func(node *ast.Node) {
				checkNamespaceExport(ctx, node)
			},
		}
	},
}

func parseOptions(options any) ruleOptions {
	opts := ruleOptions{}
	if optsMap := rslint_utils.GetOptionsMap(options); optsMap != nil {
		if allowComputed, ok := optsMap["allowComputed"].(bool); ok {
			opts.allowComputed = allowComputed
		}
	}
	return opts
}

func collectNamespaces(ctx rule.RuleContext) map[string]*import_utils.ExportMap {
	namespaces := make(map[string]*import_utils.ExportMap)
	if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil {
		return namespaces
	}

	// SourceFile listeners are not fired by the linter, so seed namespace
	// imports up front. This also preserves ESLint's import-hoisting behavior.
	for _, stmt := range ctx.SourceFile.Statements.Nodes {
		if stmt == nil || stmt.Kind != ast.KindImportDeclaration {
			continue
		}
		processImportDeclaration(ctx, namespaces, stmt.AsImportDeclaration())
	}

	return namespaces
}

func processImportDeclaration(ctx rule.RuleContext, namespaces map[string]*import_utils.ExportMap, importDecl *ast.ImportDeclaration) {
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}

	imports, ok := import_utils.GetExportMap(ctx, importDecl.ModuleSpecifier)
	if !ok {
		return
	}

	importClause := importDecl.ImportClause.AsImportClause()
	if importClause == nil {
		return
	}

	if defaultImport := importClause.Name(); defaultImport != nil {
		if meta := imports.Get(defaultExportName); meta != nil && meta.Namespace != nil {
			namespaces[defaultImport.Text()] = meta.Namespace
		}
	}

	if importClause.NamedBindings == nil {
		return
	}

	switch importClause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		namespaceImport := importClause.NamedBindings.AsNamespaceImport()
		if namespaceImport == nil || namespaceImport.Name() == nil {
			return
		}
		if imports.Size() == 0 {
			ctx.ReportNode(namespaceImport.Name(), messageNoExports(importDecl.ModuleSpecifier.Text()))
		}
		namespaces[namespaceImport.Name().Text()] = imports
	case ast.KindNamedImports:
		namedImports := importClause.NamedBindings.AsNamedImports()
		if namedImports == nil || namedImports.Elements == nil {
			return
		}
		for _, specifierNode := range namedImports.Elements.Nodes {
			if specifierNode == nil || specifierNode.Kind != ast.KindImportSpecifier {
				continue
			}
			specifier := specifierNode.AsImportSpecifier()
			importedName, ok := importSpecifierImportedName(specifier)
			if !ok {
				continue
			}
			if meta := imports.Get(importedName); meta != nil && meta.Namespace != nil {
				namespaces[specifier.Name().Text()] = meta.Namespace
			}
		}
	}
}

func importSpecifierImportedName(specifier *ast.ImportSpecifier) (string, bool) {
	if specifier == nil {
		return "", false
	}
	imported := specifier.PropertyName
	if imported == nil {
		imported = specifier.Name()
	}
	if imported == nil {
		return "", false
	}
	if ast.ModuleExportNameIsDefault(imported) {
		return defaultExportName, true
	}
	return rslint_utils.GetStaticPropertyName(imported)
}

func checkNamespaceExport(ctx rule.RuleContext, namespaceExport *ast.Node) {
	exportDecl := findParentExportDeclaration(namespaceExport)
	if exportDecl == nil || exportDecl.ModuleSpecifier == nil {
		return
	}

	imports, ok := import_utils.GetExportMap(ctx, exportDecl.ModuleSpecifier)
	if !ok || imports.Size() > 0 {
		return
	}

	ctx.ReportNode(namespaceExport, messageNoExports(exportDecl.ModuleSpecifier.Text()))
}

func checkNamespaceDestructuring(ctx rule.RuleContext, namespaces map[string]*import_utils.ExportMap, node *ast.Node) {
	variableDecl := node.AsVariableDeclaration()
	if variableDecl == nil || variableDecl.Initializer == nil {
		return
	}

	init := ast.SkipParentheses(variableDecl.Initializer)
	if init == nil || init.Kind != ast.KindIdentifier {
		return
	}

	namespaceName := init.AsIdentifier().Text
	namespace, ok := namespaces[namespaceName]
	if !ok {
		return
	}
	if isShadowedBeforeModule(init, namespaceName) {
		return
	}

	testBindingPattern(ctx, variableDecl.Name(), namespace, []string{namespaceName})
}

func testBindingPattern(ctx rule.RuleContext, pattern *ast.Node, namespace *import_utils.ExportMap, namePath []string) {
	if namespace == nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return
	}

	for _, elementNode := range pattern.AsBindingPattern().Elements.Nodes {
		if elementNode == nil || elementNode.Kind != ast.KindBindingElement {
			continue
		}
		element := elementNode.AsBindingElement()
		if element.DotDotDotToken != nil {
			continue
		}

		key := element.PropertyName
		if key == nil {
			key = element.Name()
		}
		if key == nil {
			continue
		}
		if key.Kind != ast.KindIdentifier {
			ctx.ReportNode(elementNode, messageOnlyTopLevel())
			continue
		}

		keyName := key.AsIdentifier().Text
		if !namespace.Has(keyName) {
			ctx.ReportNode(key, messageNotFound(keyName, namePath))
			continue
		}

		exported := namespace.Get(keyName)
		if exported == nil {
			continue
		}
		testBindingPattern(ctx, element.Name(), exported.Namespace, append(namePath, keyName))
	}
}

func findParentExportDeclaration(node *ast.Node) *ast.ExportDeclaration {
	parent := ast.FindAncestorKind(node, ast.KindExportDeclaration)
	if parent == nil {
		return nil
	}
	return parent.AsExportDeclaration()
}

func checkNamespaceAccess(ctx rule.RuleContext, namespaces map[string]*import_utils.ExportMap, opts ruleOptions, access *ast.Node) {
	if access == nil {
		return
	}

	object := ast.SkipParentheses(rslint_utils.AccessExpressionObject(access))
	if object == nil || object.Kind != ast.KindIdentifier {
		return
	}

	rootName := object.AsIdentifier().Text
	namespace, ok := namespaces[rootName]
	if !ok {
		return
	}
	if isShadowedBeforeModule(object, rootName) {
		return
	}

	if isAssignmentTarget(access) {
		ctx.ReportNode(access.Parent, messageAssignment(rootName))
	}

	validateNamespaceAccess(ctx, opts, access, namespace, []string{rootName}, rootName)
}

func validateNamespaceAccess(ctx rule.RuleContext, opts ruleOptions, access *ast.Node, namespace *import_utils.ExportMap, namePath []string, rootName string) {
	current := access
	for namespace != nil && current != nil && ast.IsAccessExpression(current) {
		if current.Kind == ast.KindElementAccessExpression {
			argument := current.AsElementAccessExpression().ArgumentExpression
			if !opts.allowComputed && argument != nil {
				ctx.ReportNode(argument, messageComputed(rootName))
			}
			return
		}

		property := current.AsPropertyAccessExpression().Name()
		if property == nil {
			return
		}
		propertyName := property.Text()
		if !namespace.Has(propertyName) {
			ctx.ReportNode(property, messageNotFound(propertyName, namePath))
			return
		}

		exported := namespace.Get(propertyName)
		if exported == nil {
			return
		}

		namePath = append(namePath, propertyName)
		namespace = exported.Namespace

		parent := current.Parent
		if parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = ast.WalkUpParenthesizedExpressions(parent)
		}
		if parent == nil || !ast.IsAccessExpression(parent) {
			return
		}
		if ast.SkipParentheses(rslint_utils.AccessExpressionObject(parent)) != current {
			return
		}
		current = parent
	}
}

func isAssignmentTarget(access *ast.Node) bool {
	if access == nil || access.Parent == nil || !ast.IsAssignmentExpression(access.Parent, false) {
		return false
	}
	return ast.SkipParentheses(access.Parent.AsBinaryExpression().Left) == access
}

func isShadowedBeforeModule(identifier *ast.Node, name string) bool {
	if identifier == nil {
		return false
	}
	boundary := ast.FindAncestorKind(identifier, ast.KindSourceFile)
	nestedFunctionSeen := false
	nestedSignatureNamespaceRef := false
	for current := identifier.Parent; current != nil && current != boundary; current = current.Parent {
		if ast.IsFunctionLikeDeclaration(current) {
			signatureNamespaceRef := functionSignatureReferencesNamespaceMember(current, name)
			if rslint_utils.HasShadowingParameter(current, name) {
				if nestedSignatureNamespaceRef || (!nestedFunctionSeen && signatureNamespaceRef) {
					// eslint-plugin-import loses the value shadow when the
					// active function signature uses the same namespace in a
					// qualified type, e.g. `fn(ns: ns.Type) { ns.missing }`.
				} else {
					return true
				}
			}
			if current.Kind == ast.KindFunctionDeclaration || current.Kind == ast.KindFunctionExpression {
				if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
					return true
				}
			}
			if body := current.Body(); body != nil && rslint_utils.HasHoistedVarDeclaration(body, name) {
				return true
			}
			if signatureNamespaceRef {
				nestedSignatureNamespaceRef = true
			}
			nestedFunctionSeen = true
		}
		switch current.Kind {
		case ast.KindBlock:
			if rslint_utils.HasShadowingDeclaration(current, name) {
				return true
			}
		case ast.KindCatchClause:
			cc := current.AsCatchClause()
			if cc != nil && cc.VariableDeclaration != nil {
				vd := cc.VariableDeclaration.AsVariableDeclaration()
				if vd != nil && vd.Name() != nil && rslint_utils.HasNameInBindingPattern(vd.Name(), name) {
					return true
				}
			}
		case ast.KindForStatement:
			forStmt := current.AsForStatement()
			if forStmt != nil && forStmt.Initializer != nil &&
				forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				rslint_utils.HasVarDeclListWithName(forStmt.Initializer, name) {
				return true
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := current.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer != nil &&
				stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				rslint_utils.HasVarDeclListWithName(stmt.Initializer, name) {
				return true
			}
		case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindEnumDeclaration:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
	}
	return false
}

func functionSignatureReferencesNamespaceMember(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		if parameterTypeReferencesNamespaceMember(param, name) {
			return true
		}
	}
	return typeNodeReferencesNamespaceMember(node.Type(), name)
}

func parameterTypeReferencesNamespaceMember(param *ast.Node, name string) bool {
	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil || paramDecl.Type == nil {
		return false
	}
	return typeNodeReferencesNamespaceMember(paramDecl.Type.AsNode(), name)
}

func typeNodeReferencesNamespaceMember(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		if current == nil || found {
			return found
		}
		if current.Kind == ast.KindQualifiedName && qualifiedNameRootName(current) == name {
			found = true
			return true
		}
		current.ForEachChild(visit)
		return found
	}
	visit(node)
	return found
}

func qualifiedNameRootName(node *ast.Node) string {
	current := node
	for current != nil && current.Kind == ast.KindQualifiedName {
		current = current.AsQualifiedName().Left
	}
	if current != nil && current.Kind == ast.KindIdentifier {
		return current.AsIdentifier().Text
	}
	return ""
}

func messageNoExports(moduleName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noExports",
		Description: fmt.Sprintf("No exported names found in module '%s'.", moduleName),
	}
}

func messageComputed(namespaceName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "computed",
		Description: fmt.Sprintf("Unable to validate computed reference to imported namespace '%s'.", namespaceName),
	}
}

func messageAssignment(namespaceName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "assignment",
		Description: fmt.Sprintf("Assignment to member of namespace '%s'.", namespaceName),
	}
}

func messageNotFound(name string, namePath []string) rule.RuleMessage {
	importKind := "imported"
	if len(namePath) > 1 {
		importKind = "deeply imported"
	}
	return rule.RuleMessage{
		Id:          "notFound",
		Description: fmt.Sprintf("'%s' not found in %s namespace '%s'.", name, importKind, strings.Join(namePath, ".")),
	}
}

func messageOnlyTopLevel() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "onlyTopLevel",
		Description: "Only destructure top-level names.",
	}
}

const defaultExportName = "default"
