package no_unresolved

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed no_unresolved.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-unresolved.js
// Upstream reports literal messages, without message IDs, fixes or suggestions.
var NoUnresolvedRule = rule.Rule{
	Name:   "import/no-unresolved",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		kinds := modules.ESModuleReferences
		if opts["esmodule"] == false {
			kinds = 0
		}
		if opts["commonjs"] == true {
			kinds |= modules.CommonJSReferences
		}
		if opts["amd"] == true {
			kinds |= modules.AMDReferences
		}
		var ignored []*esregexp.RegExp
		if patterns, ok := opts["ignore"].([]any); ok {
			for _, pattern := range patterns {
				text, _ := pattern.(string)
				if expression, err := esregexp.Compile(text, ""); err == nil {
					ignored = append(ignored, expression)
				}
			}
		}
		var resolver *import_utils.ImportResolver
		reportedResolverError := false
		strict := opts["caseSensitiveStrict"] == true
		checkCase := !ctx.Program().FS().UseCaseSensitiveFileNames() && (strict || opts["caseSensitive"] != false)
		for _, ref := range modules.Collect(ctx.SourceFile, kinds) {
			source := unresolvedSource(ref)
			if source == nil || slices.ContainsFunc(ignored, func(re *esregexp.RegExp) bool { return re.TestOrTimeout(source.Text()) }) {
				continue
			}
			if resolver == nil {
				resolver = import_utils.NewImportResolver(ctx)
			}
			path, found, resolveError := resolver.Resolve(source)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			if !found {
				ctx.ReportNode(source, rule.RuleMessage{Description: "Unable to resolve path to module '" + source.Text() + "'."})
			} else if checkCase && !pathHasExactCase(ctx, path, strict) {
				ctx.ReportNode(source, rule.RuleMessage{Description: "Casing of " + source.Text() + " does not match the underlying filesystem."})
			}
		}
		return nil
	},
}

// The shared collector keeps all expressions and emitted type-only edges.
// This rule checks only literal strings and skips explicit import/export type
// declarations, matching moduleVisitor rather than emitted dependencies.
func unresolvedSource(ref modules.Source) *ast.Node {
	node := ref.Declaration
	switch ref.Kind {
	case modules.ModuleReferenceImport:
		if clause := node.AsImportDeclaration().ImportClause; clause != nil && clause.IsTypeOnly() {
			return nil
		}
	case modules.ModuleReferenceExport:
		if node.AsExportDeclaration().IsTypeOnly {
			return nil
		}
	}
	return import_utils.LiteralModuleSource(ref)
}

type exactCaseKey struct {
	path, cwd string
	strict    bool
}

// Check the resolved spelling through the Program's filesystem, including
// overlays. Only strict mode checks the working directory and its ancestors.
func pathHasExactCase(ctx rule.RuleContext, path string, strict bool) bool {
	if path == "" {
		return true
	}
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = ctx.Program().CurrentDirectory()
	}
	return rule.CachedByProgram(ctx, exactCaseKey{path, cwd, strict}, func() bool {
		for path != "" {
			if !strict && ecmascript.EqualsWhenLowercased(path, cwd) {
				return true
			}
			directory := tspath.GetDirectoryPath(path)
			if directory == path || directory == "" {
				return true
			}
			entries := ctx.Program().FS().GetAccessibleEntries(directory)
			name := tspath.GetBaseFileName(path)
			if !slices.Contains(entries.Files, name) && !slices.Contains(entries.Directories, name) {
				return false
			}
			path = directory
		}
		return true
	})
}
