package no_relative_packages

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// Native schema validation rejects malformed ignore regexps before traversal;
// upstream throws when its module visitor constructs RegExp for the same input.
//
//go:embed no_relative_packages.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-relative-packages.js
var NoRelativePackagesRule = rule.Rule{
	Name:   "import/no-relative-packages",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		visitorOptions := import_utils.VisitModulesOptions{
			ESModule: opts["esmodule"] != false, Commonjs: opts["commonjs"] == true,
			AMD: opts["amd"] == true, Ignore: utils.ToStringSlice(opts["ignore"]),
		}
		settings := import_utils.SettingsFor(ctx)
		var resolver *import_utils.ImportResolver
		reportedResolverError := false
		return import_utils.VisitModules(func(source modules.Source) {
			imported := source.Specifier().Text()
			// importType resolves even non-relative names before classifying them.
			if resolver == nil {
				resolver = import_utils.NewImportResolver(ctx)
			}
			resolved, found, resolveError := resolver.Resolve(source)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			relative := imported == "." || imported == ".." || strings.HasPrefix(imported, "./") || strings.HasPrefix(imported, "../") || strings.HasPrefix(imported, `.\`) || strings.HasPrefix(imported, `..\`)
			if !relative || settings.IsInternalSpecifier(imported) || settings.IsBuiltinSpecifier(imported, resolved) || !found || resolved == "" {
				return
			}
			importName, importRoot := findNamedPackage(ctx.Program(), resolved)
			contextName, _ := findNamedPackage(ctx.Program(), ctx.SourceFile.FileName())
			if importName == "" || contextName == "" || importName == contextName {
				return
			}
			properPath, err := filepath.Rel(importRoot, resolved)
			if err != nil {
				return
			}
			base := filepath.Base(imported)
			if base == filepath.Base(importRoot) {
				base = ""
			}
			properImport := filepath.Join(importName, filepath.Dir(properPath), base)
			ctx.ReportNodeWithDeferredFixes(source.Specifier(), rule.RuleMessage{Description: fmt.Sprintf(
				"Relative import from another package is not allowed. Use `%s` instead of `%s`", properImport, imported,
			)}, func() []rule.RuleFix {
				quoted := utils.Must(core.StringifyJson(strings.ReplaceAll(properImport, `\`, "/"), "", ""))
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, source.Specifier(), quoted)}
			})
		}, visitorOptions)
	},
}

// Read through the Program's filesystem so overlays and disk files share the
// same authority. Unnamed packages inherit their nearest named ancestor;
// Malformed JSON stops the search, as readPkgUp does upstream. The shared
// reader also rejects non-object manifests; upstream can inherit through those.
func findNamedPackage(p *program.Program, fileName string) (string, string) {
	directory := tspath.GetDirectoryPath(fileName)
	for {
		root := p.NearestPackageJSONDirectory(directory)
		if root == "" {
			return "", ""
		}
		var nameValue any
		if pkg := packagejson.Read(p, root); pkg != nil {
			nameValue = pkg.Field("name")
		} else {
			// readPkgUp strips a leading BOM. The shared metadata reader
			// deliberately accepts only strict JSON, so adapt just that edge.
			content, ok := p.FS().ReadFile(tspath.ResolvePath(root, "package.json"))
			if !ok || !strings.HasPrefix(content, "\ufeff") {
				return "", ""
			}
			var pkg map[string]any
			if json.Unmarshal([]byte(strings.TrimPrefix(content, "\ufeff")), &pkg) != nil || pkg == nil {
				return "", ""
			}
			nameValue = pkg["name"]
		}
		if name, ok := nameValue.(string); ok && name != "" {
			return name, root
		}
		parent := tspath.GetDirectoryPath(root)
		if parent == root {
			return "", ""
		}
		directory = parent
	}
}
