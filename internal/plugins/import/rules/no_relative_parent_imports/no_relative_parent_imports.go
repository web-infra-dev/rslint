package no_relative_parent_imports

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed no_relative_parent_imports.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-relative-parent-imports.js
// Upstream reports a literal message without a message ID, fixes or suggestions.
var NoRelativeParentImportsRule = rule.Rule{
	Name:   "import/no-relative-parent-imports",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Native SourceFiles always have an absolute filename, unlike ESLint's
		// special <text> input.
		fileName := ctx.SourceFile.FileName()
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
		for _, pattern := range utils.ToStringSlice(opts["ignore"]) {
			if expression, err := esregexp.Compile(pattern, ""); err == nil {
				ignored = append(ignored, expression)
			}
		}
		var settings *import_utils.ModuleSettings
		var resolver *import_utils.ImportResolver
		reportedResolverError := false
		for _, ref := range modules.Collect(ctx.SourceFile, kinds) {
			source := import_utils.LiteralModuleSource(ref)
			if source == nil || slices.ContainsFunc(ignored, func(re *esregexp.RegExp) bool { return re.TestOrTimeout(source.Text()) }) {
				continue
			}
			if resolver == nil {
				settings = import_utils.SettingsFor(ctx)
				resolver = import_utils.NewImportResolver(ctx)
			}
			resolvedPath, found, resolveError := resolver.Resolve(source)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			if !found || resolvedPath == "" || settings.IsExternalResolvedImport(ctx, source.Text(), resolvedPath) {
				continue
			}
			relativePath, err := filepath.Rel(tspath.GetDirectoryPath(fileName), resolvedPath)
			if err != nil || settings.IsInternalSpecifier(relativePath) ||
				!strings.HasPrefix(relativePath, "../") && !strings.HasPrefix(relativePath, `..\`) {
				continue
			}
			ctx.ReportNode(source, rule.RuleMessage{Description: fmt.Sprintf(
				"Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `%s` to same directory as `%s` or consider making `%s` a package.",
				tspath.GetBaseFileName(fileName), source.Text(), source.Text(),
			)})
		}
		return nil
	},
}
