package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Templates belong to one exact SourceFile. Their steps, documentation and
// containers are immutable and contain no resolved links or Program state.
// Only statements relevant to the export index occupy a slot.
type exportStatement struct {
	statement *ast.Node
	steps     []exportStep
}

type fileExportSyntax struct {
	statements         []exportStatement
	localExportedNames map[string]bool
}

var exportSyntaxKey = ast.NewSourceFileDataKey[fileExportSyntax]()

func exportSyntaxFor(file *ast.SourceFile) fileExportSyntax {
	return ast.GetOrComputeSourceFileData(file, exportSyntaxKey, func(file *ast.SourceFile) fileExportSyntax {
		var localExportedNames map[string]bool
		for _, stmt := range file.Statements.Nodes {
			if stmt == nil || stmt.Kind != ast.KindExportDeclaration {
				continue
			}
			exportDecl := stmt.AsExportDeclaration()
			if exportDecl == nil || exportDecl.ModuleSpecifier != nil ||
				exportDecl.ExportClause == nil || exportDecl.ExportClause.Kind != ast.KindNamedExports {
				continue
			}
			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports == nil || namedExports.Elements == nil {
				continue
			}
			for _, spec := range namedExports.Elements.Nodes {
				if spec == nil || spec.Kind != ast.KindExportSpecifier {
					continue
				}
				specifier := spec.AsExportSpecifier()
				sourceName := specifier.PropertyName
				if sourceName == nil {
					sourceName = specifier.Name()
				}
				if ident := referencedIdentifierText(sourceName); ident != "" {
					if localExportedNames == nil {
						localExportedNames = make(map[string]bool)
					}
					localExportedNames[ident] = true
				}
			}
		}

		var syntax []exportStatement
		for _, stmt := range file.Statements.Nodes {
			if stmt == nil {
				continue
			}
			entry := exportStatement{statement: stmt}
			local := localExports{}
			local.appendSyntaxStatement(stmt)
			for i := range local.Steps {
				captureExportDocSources(file, &local.Steps[i], stmt)
			}
			entry.steps = local.Steps
			if len(entry.steps) != 0 ||
				stmt.Kind == ast.KindImportDeclaration && importDeclRelevantForExports(stmt.AsImportDeclaration(), localExportedNames) ||
				stmt.Kind == ast.KindNamespaceExportDeclaration ||
				stmt.Kind == ast.KindExportAssignment && stmt.AsExportAssignment().IsExportEquals {
				syntax = append(syntax, entry)
			}
		}
		return fileExportSyntax{
			statements:         syntax,
			localExportedNames: localExportedNames,
		}
	})
}

func (local *localExports) appendSyntaxStatement(stmt *ast.Node) {
	if isExportedDeclaration(stmt) {
		if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
			var localName string
			switch stmt.Kind {
			case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindInterfaceDeclaration:
				if name := stmt.Name(); name != nil {
					localName = name.Text()
				}
			}
			local.Steps = append(local.Steps, exportStep{
				Kind:  exportStepNames,
				Names: []string{defaultExportName},
				Local: localName,
			})
			return
		}
		if names := exportedDeclarationNames(stmt); len(names) > 0 {
			local.Steps = append(local.Steps, exportStep{Kind: exportStepNames, Names: names})
		}
		return
	}

	switch stmt.Kind {
	case ast.KindExportAssignment:
		assignment := stmt.AsExportAssignment()
		if !assignment.IsExportEquals {
			local.Steps = append(local.Steps, exportStep{Kind: exportStepLocalDefault, Local: referencedIdentifierText(assignment.Expression)})
		}
	case ast.KindExportDeclaration:
		local.appendExportDeclaration(stmt.AsExportDeclaration())
	}
}
