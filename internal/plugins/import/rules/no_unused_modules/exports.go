package no_unused_modules

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// localExport is one name the linted module exports, and the node a diagnostic
// for it reports.
type localExport struct {
	name   string
	node   *ast.Node
	isType bool
}

// moduleExports describes what the linted module itself exports. checked holds
// the names the rule may report as unused; count also covers exports the rule
// never reports on but which still make the module export something.
type moduleExports struct {
	checked []localExport
	count   int
}

// collectModuleExports reads the top-level statements of file. Only those can
// export from the module: an `export` inside a namespace body exports from the
// namespace, so it is neither a module export nor checked.
func collectModuleExports(file *ast.SourceFile) moduleExports {
	var result moduleExports
	statements := file.Statements.Nodes
	implemented := implementedFunctions(statements)
	declaredSignature := make(map[string]bool)
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindExportAssignment:
			result.count++
			// `export = value` (TypeScript CommonJS export) is never checked
			// upstream; only `export default` is.
			if !statement.AsExportAssignment().IsExportEquals {
				result.checked = append(result.checked, localExport{name: "default", node: statement})
			}
		case ast.KindExportDeclaration:
			collectExportDeclaration(statement, statements, &result)
		case ast.KindNamespaceExportDeclaration:
			// `export as namespace X` only names a UMD global.
		default:
			if !ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
				continue
			}
			if ast.HasSyntacticModifier(statement, ast.ModifierFlagsDefault) {
				result.count++
				result.checked = append(result.checked, localExport{name: "default", node: statement})
				continue
			}
			collectDeclaration(statement, implemented, declaredSignature, &result)
		}
	}
	return result
}

func collectExportDeclaration(statement *ast.Node, statements []*ast.Node, result *moduleExports) {
	declaration := statement.AsExportDeclaration()
	clause := declaration.ExportClause
	if clause == nil {
		// `export * from 'x'` re-exports the target's names. Upstream counts
		// none of them, which makes a pure barrel report "No exports found".
		result.count++
		return
	}
	if clause.Kind == ast.KindNamespaceExport {
		// `export * as ns from 'x'` exports ns. Upstream does not check it.
		result.count++
		return
	}
	typeOnlyDeclaration := declaration.IsTypeOnly
	reexport := declaration.ModuleSpecifier != nil
	for _, element := range clause.AsNamedExports().Elements.Nodes {
		specifier := element.AsExportSpecifier()
		name := moduleExportNameText(specifier.Name())
		result.count++
		isType := typeOnlyDeclaration || specifier.IsTypeOnly
		if !isType && !reexport {
			local := specifier.PropertyName
			if local == nil {
				local = specifier.Name()
			}
			isType = local.Kind == ast.KindIdentifier && declaresOnlyTypes(statements, local.Text())
		}
		result.checked = append(result.checked, localExport{name: name, node: element, isType: isType})
	}
}

func collectDeclaration(statement *ast.Node, implemented map[string]bool, declaredSignature map[string]bool, result *moduleExports) {
	switch statement.Kind {
	case ast.KindVariableStatement:
		for _, declarator := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			// Every bound name counts, including those nested in array and
			// object patterns with defaults; holes bind nothing.
			utils.CollectBindingNames(declarator.Name(), func(_ *ast.Node, name string) {
				result.count++
				result.checked = append(result.checked, localExport{name: name, node: statement})
			})
		}
	case ast.KindFunctionDeclaration:
		name := statement.Name()
		if name == nil {
			return
		}
		text := name.Text()
		if statement.Body() == nil {
			// An overload signature shares its name with the implementation,
			// which is the one reported. An ambient `declare function` has
			// none; report its first signature once.
			if implemented[text] || declaredSignature[text] {
				result.count++
				return
			}
			declaredSignature[text] = true
		}
		result.count++
		result.checked = append(result.checked, localExport{name: text, node: statement})
	case ast.KindClassDeclaration:
		if name := statement.Name(); name != nil {
			result.count++
			result.checked = append(result.checked, localExport{name: name.Text(), node: statement})
		}
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
		result.count++
		result.checked = append(result.checked, localExport{name: statement.Name().Text(), node: statement, isType: true})
	case ast.KindModuleDeclaration:
		name := statement.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return
		}
		result.count++
		result.checked = append(result.checked, localExport{name: name.Text(), node: statement})
	case ast.KindImportEqualsDeclaration:
		// `export import A = N.B` exports A; upstream does not check it.
		result.count++
	}
}

// implementedFunctions returns the top-level function names that have a body.
func implementedFunctions(statements []*ast.Node) map[string]bool {
	implemented := make(map[string]bool)
	for _, statement := range statements {
		if statement.Kind == ast.KindFunctionDeclaration && statement.Body() != nil {
			if name := statement.Name(); name != nil {
				implemented[name.Text()] = true
			}
		}
	}
	return implemented
}

// declaresOnlyTypes reports whether the top-level declarations of name are all
// interfaces, type aliases or enums: the kinds whose direct export upstream
// treats as a type export for ignoreUnusedTypeExports.
func declaresOnlyTypes(statements []*ast.Node, name string) bool {
	found := false
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
			if statement.Name().Text() == name {
				found = true
			}
		case ast.KindVariableStatement:
			declared := false
			for _, declarator := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				utils.CollectBindingNames(declarator.Name(), func(_ *ast.Node, bound string) {
					declared = declared || bound == name
				})
			}
			if declared {
				return false
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindModuleDeclaration, ast.KindImportEqualsDeclaration:
			if declared := statement.Name(); declared != nil && declared.Kind == ast.KindIdentifier && declared.Text() == name {
				return false
			}
		case ast.KindImportDeclaration:
			// An imported binding re-exported locally may be a value.
			if importDeclares(statement, name) {
				return false
			}
		}
	}
	return found
}

func importDeclares(statement *ast.Node, name string) bool {
	clause := statement.AsImportDeclaration().ImportClause
	if clause == nil {
		return false
	}
	importClause := clause.AsImportClause()
	if local := importClause.Name(); local != nil && local.Text() == name {
		return true
	}
	bindings := importClause.NamedBindings
	if bindings == nil {
		return false
	}
	if bindings.Kind == ast.KindNamespaceImport {
		return bindings.Name().Text() == name
	}
	for _, element := range bindings.AsNamedImports().Elements.Nodes {
		if element.Name().Text() == name {
			return true
		}
	}
	return false
}

// moduleExportNameText returns an export or import name as JavaScript sees
// it: an identifier's text, or a string literal's value for ES2022 arbitrary
// module namespace names.
func moduleExportNameText(name *ast.Node) string {
	if name == nil {
		return ""
	}
	return name.Text()
}
