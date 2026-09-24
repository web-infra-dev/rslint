package no_unassigned_import

import (
	_ "embed"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

//go:embed no_unassigned_import.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-unassigned-import.js
var NoUnassignedImportRule = rule.Rule{
	Name:   "import/no-unassigned-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var option map[string]any
		if len(options) > 0 {
			option, _ = options[0].(map[string]any)
		}
		patterns, _ := option["allow"].([]any)
		var allowed []*minimatch3.Matcher
		isAllowed := func(source string) bool {
			if len(patterns) == 0 {
				return false
			}
			// Files without a bare import or require never need glob matchers.
			if allowed == nil {
				cwd := ctx.ProcessCurrentDirectory()
				if cwd == "" {
					cwd = ctx.Program().CurrentDirectory()
				}
				allowed = make([]*minimatch3.Matcher, 0, 2*len(patterns))
				for _, value := range patterns {
					pattern, _ := value.(string)
					pattern = filepath.ToSlash(pattern)
					cwdPattern := filepath.ToSlash(filepath.Join(cwd, pattern))
					// Node's path.join retains a trailing separator in glob patterns.
					if strings.HasSuffix(pattern, "/") && !strings.HasSuffix(cwdPattern, "/") {
						cwdPattern += "/"
					}
					allowed = append(allowed,
						minimatch3.New(pattern, minimatch3.Options{}),
						minimatch3.New(cwdPattern, minimatch3.Options{}),
					)
				}
			}
			// Only relative and slash-prefixed specifiers are file paths.
			// Resolve them lexically; this rule does not resolve modules on disk.
			if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") {
				directory := filepath.Dir(ctx.SourceFile.FileName())
				if strings.HasPrefix(source, "/") {
					directory = filepath.VolumeName(directory)
				}
				source = filepath.Join(directory, source)
			}
			source = filepath.ToSlash(source)
			for _, pattern := range allowed {
				if pattern.Match(source) {
					return true
				}
			}
			return false
		}
		// Upstream reports a literal message, without fixes or suggestions.
		message := rule.RuleMessage{Description: "Imported module should be assigned"}
		listeners := import_utils.VisitModules(func(source *ast.StringLiteralLike, node *ast.Node) {
			// A ChainExpression or a used require result is not a bare call
			// expression statement in ESTree.
			parent := utils.ESTreeParent(node)
			if !ast.IsOptionalChain(node) && parent != nil && parent.Kind == ast.KindExpressionStatement && !isAllowed(source.Text()) {
				ctx.ReportNode(node, message)
			}
		}, import_utils.VisitModulesOptions{Commonjs: true})
		listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			if declaration.ImportClause != nil {
				clause := declaration.ImportClause.AsImportClause()
				if clause.Name() != nil || clause.NamedBindings != nil &&
					(clause.NamedBindings.Kind != ast.KindNamedImports || len(clause.NamedBindings.AsNamedImports().Elements.Nodes) > 0) {
					return
				}
			}
			if declaration.ModuleSpecifier != nil && !isAllowed(declaration.ModuleSpecifier.Text()) {
				ctx.ReportNode(node, message)
			}
		}
		return listeners
	},
}
