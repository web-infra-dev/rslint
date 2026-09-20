package no_unpublished_import

import (
	_ "embed"
	"path/filepath"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

//go:embed no_unpublished_import.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-import.js
var NoUnpublishedImportRule = rule.Rule{
	Name:   "node/no-unpublished-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		p, fileName := ctx.Program(), ctx.SourceFile.FileName()
		if p == nil || fileName == "<input>" {
			return nil
		}
		pkg := packagejson.FindNearestValid(p, fileName)
		if pkg == nil {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		if opts["ignorePrivate"] != false && pkg.Field("private") == true {
			return nil
		}
		if len(modules.Collect(ctx.SourceFile, modules.ESModuleReferences)) == 0 {
			return nil
		}
		cwd := ctx.ProcessCurrentDirectory()
		if cwd == "" {
			cwd = p.CurrentDirectory()
		}
		toRelative := func(target string) (string, bool) {
			// Like Node's path.relative, unresolved targets and URL spellings
			// are relative to the working directory. Normalize at the VFS boundary.
			if !nodeutil.IsAbsolutePath(target) {
				target = filepath.Join(filepath.FromSlash(cwd), target)
			}
			relative := tspath.GetRelativePathFromDirectory(pkg.Directory(), tspath.NormalizePath(target), tspath.ComparePathsOptions{UseCaseSensitiveFileNames: p.FS().UseCaseSensitiveFileNames()})
			return nodeutil.ConvertPath(relative, opts, ctx.Settings)
		}
		converted, ok := toRelative(fileName)
		// Share nodeutil's publication policy, including its documented root
		// metadata exemptions, nested ignore files and conversion behavior.
		if !ok || nodeutil.IsUnpublished(p, pkg, tspath.ResolvePath(pkg.Directory(), converted)) {
			return nil
		}
		allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
		development, _ := pkg.Field("devDependencies").(map[string]any)
		ignoreTypeImport, _ := opts["ignoreTypeImport"].(bool)
		var resolutionOptions [2]*moduleresolver.Options
		var unpublished [2]map[string]bool
		return nodeutil.VisitImports(ctx, nodeutil.ImportVisitorOptions{IgnoreTypeImport: ignoreTypeImport}, func(source *ast.Node, specifier string, typeOnly bool) {
			name, resource := nodeutil.ImportModuleName(specifier)
			if name != "" {
				if _, exists := development[name]; !exists || slices.Contains(allowed, name) {
					return
				}
				// Publication uses this package's own dependency declarations,
				// without inheriting development tools from workspace ancestors.
				for _, field := range []string{"dependencies", "peerDependencies", "optionalDependencies"} {
					dependencies, _ := pkg.Field(field).(map[string]any)
					if _, exists := dependencies[name]; exists {
						return
					}
				}
			} else {
				index := 0
				if typeOnly {
					index = 1
				}
				ignored, checked := unpublished[index][resource]
				if !checked {
					if resolutionOptions[index] == nil {
						resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, opts)
						resolutionOptions[index] = &resolution
						unpublished[index] = map[string]bool{}
					}
					target := nodeutil.ImportFilePath(p, resource, fileName, typeOnly, *resolutionOptions[index])
					relative, ok := toRelative(target)
					ignored = ok && relative != "" && nodeutil.IsUnpublished(p, pkg, tspath.ResolvePath(pkg.Directory(), relative))
					// Options and package metadata stay fixed within this file.
					// Type imports can resolve differently; each occurrence still reports.
					unpublished[index][resource] = ignored
				}
				if !ignored {
					return
				}
				name = specifier
			}
			ctx.ReportNode(source, rule.RuleMessage{
				Id: "notPublished", Description: `"` + name + `" is not published.`,
				Data: map[string]string{"name": name},
			})
		})
	},
}
