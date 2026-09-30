package require_artifact_dependency_entry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var RequireArtifactDependencyEntryRule = rule.Rule{
	Name:   "rslim/require-artifact-dependency-entry",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		program := ctx.Program()
		file := ctx.SourceFile.FileName()
		artifactKind := ""
		if pairedDeclaration(program.FileExists, file) != "" {
			artifactKind = "paired JavaScript"
		} else if isTypeScriptSource(file) && program.Options() != nil && program.Options().NoCheck.IsTrue() {
			artifactKind = "noCheck TypeScript"
		}
		if artifactKind == "" {
			return nil
		}

		report := func(specifier *ast.Node, candidates string) {
			if specifier == nil || !ast.IsStringLiteralLike(specifier) {
				return
			}
			resolved, target, found := program.ResolveModule(ctx.SourceFile, specifier)
			if !found {
				return
			}
			source := sourceBuiltTarget(program, resolved, target)
			if source == "" {
				return
			}
			ctx.ReportNode(specifier, rule.RuleMessage{
				Id:          "artifactDependency",
				Description: fmt.Sprintf("This %s module imports %q, which resolves to source-built module %q. Candidate runtime uses: %s. Rslim may not see references across this artifact boundary. Review the emitted JavaScript and dependent source modules; mark required symbols with @entry, or add source files to Rslim entries in Lib Mode.", artifactKind, specifier.Text(), source, candidates),
				Data:        map[string]string{"artifactKind": artifactKind, "specifier": specifier.Text(), "resolvedPath": resolved, "targetSource": source, "candidateUses": candidates},
			})
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if !hasRuntimeImport(declaration.ImportClause) {
					return
				}
				report(declaration.ModuleSpecifier, importCandidates(declaration.ImportClause))
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				declaration := node.AsExportDeclaration()
				if hasRuntimeExport(declaration) {
					report(declaration.ModuleSpecifier, exportCandidates(declaration))
				}
			},
		}
	},
}

func sourceBuiltTarget(program *lintprogram.Program, resolved string, target *ast.SourceFile) string {
	if source, noCheck, referenced := program.ProjectReferenceSource(resolved); referenced {
		if !noCheck && isTypeScriptSource(source) {
			return source
		}
		return ""
	}
	if target == nil {
		return ""
	}
	if options := program.Options(); options != nil && !options.NoCheck.IsTrue() && !target.IsDeclarationFile && isTypeScriptSource(target.FileName()) {
		return target.FileName()
	}
	return ""
}

func hasRuntimeImport(clause *ast.ImportClauseNode) bool {
	// Inline `type` specifiers may still emit an empty import under
	// verbatimModuleSyntax, preserving the target module's side effects.
	return clause == nil || !clause.IsTypeOnly()
}

func hasRuntimeExport(declaration *ast.ExportDeclaration) bool {
	return !declaration.IsTypeOnly
}

func importCandidates(clause *ast.ImportClauseNode) string {
	if clause == nil {
		return "module side effects"
	}
	names := []string{}
	if clause.Name() != nil {
		names = append(names, "default")
	}
	bindings := clause.AsImportClause().NamedBindings
	if bindings != nil {
		if bindings.Kind == ast.KindNamespaceImport {
			return "unknown namespace exports; inspect the emitted JavaScript"
		}
		for _, element := range bindings.AsNamedImports().Elements.Nodes {
			imported := element.AsImportSpecifier()
			if imported.IsTypeOnly {
				continue
			}
			name := imported.PropertyName
			if name == nil {
				name = imported.Name()
			}
			names = append(names, name.Text())
		}
	}
	return candidateNames(names)
}

func exportCandidates(declaration *ast.ExportDeclaration) string {
	if declaration.ExportClause == nil || declaration.ExportClause.Kind == ast.KindNamespaceExport {
		return "unknown re-exported values; inspect the emitted JavaScript"
	}
	names := []string{}
	for _, element := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
		exported := element.AsExportSpecifier()
		if exported.IsTypeOnly {
			continue
		}
		name := exported.PropertyName
		if name == nil {
			name = exported.Name()
		}
		names = append(names, name.Text())
	}
	return candidateNames(names)
}

func candidateNames(names []string) string {
	if len(names) == 0 {
		return "module side effects"
	}
	slices.Sort(names)
	names = slices.Compact(names)
	return strings.Join(names, ", ")
}

func isTypeScriptSource(file string) bool {
	if strings.HasSuffix(file, ".d.ts") || strings.HasSuffix(file, ".d.mts") || strings.HasSuffix(file, ".d.cts") {
		return false
	}
	return strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx") || strings.HasSuffix(file, ".mts") || strings.HasSuffix(file, ".cts")
}

func pairedDeclaration(exists func(string) bool, runtime string) string {
	if strings.HasSuffix(runtime, ".js") {
		path := strings.TrimSuffix(runtime, ".js") + ".d.ts"
		if exists(path) {
			return path
		}
	}
	return ""
}
