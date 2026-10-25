package utils

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Templates belong to one exact SourceFile. Their steps, documentation and
// containers are immutable and contain no resolved links or Program state.
// Only statements relevant to the export index occupy a slot.
type exportStatement struct {
	statement *ast.Node
	steps     []exportStep
}

var exportSyntaxKey = ast.NewSourceFileDataKey[[]exportStatement]()

func exportSyntaxFor(file *ast.SourceFile) []exportStatement {
	return ast.GetOrComputeSourceFileData(file, exportSyntaxKey, func(file *ast.SourceFile) []exportStatement {
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
			if len(entry.steps) != 0 || stmt.Kind == ast.KindImportDeclaration ||
				stmt.Kind == ast.KindNamespaceExportDeclaration ||
				stmt.Kind == ast.KindExportAssignment && stmt.AsExportAssignment().IsExportEquals {
				syntax = append(syntax, entry)
			}
		}
		return syntax
	})
}

func (local *localExports) appendSyntaxStatement(stmt *ast.Node) {
	if isExportedDeclaration(stmt) {
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
