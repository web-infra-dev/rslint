package no_absolute_path

import (
	_ "embed"
	"encoding/json"
	"path"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_absolute_path.schema.json
var schemaJSON []byte

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-absolute-path.js
var NoAbsolutePathRule = rule.Rule{
	Name:   "import/no-absolute-path",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		var fromParts []string
		var cwd string
		return import_utils.VisitModules(func(source, _ *ast.Node) {
			if !nodeutil.IsAbsolutePath(source.Text()) {
				return
			}
			ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
				Description: "Do not import modules using an absolute path",
			}, func() []rule.RuleFix {
				if fromParts == nil {
					cwd = ctx.ProcessCurrentDirectory()
					if cwd == "" && ctx.Program().IsValid() {
						cwd = ctx.Program().CurrentDirectory()
					}
					// Node's path.posix resolves relative inputs against the cwd
					// without its Windows drive prefix.
					if runtime.GOOS == "windows" {
						cwd = strings.ReplaceAll(cwd, `\`, "/")
						if slash := strings.IndexByte(cwd, '/'); slash >= 0 {
							cwd = cwd[slash:]
						}
					}
					fromParts = posixPathComponents(tspath.GetDirectoryPath(ctx.SourceFile.FileName()), cwd)
				}
				relative := relativeImportPath(fromParts, source.Text(), cwd)
				// JSON encoding must not replace an unpaired UTF-16 surrogate.
				if !utf8.ValidString(relative) {
					return nil
				}
				text, _ := json.Marshal(relative)
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, source, string(text))}
			})
		}, import_utils.VisitModulesOptions{
			ESModule: opts["esmodule"] != false,
			Commonjs: opts["commonjs"] != false,
			AMD:      opts["amd"] == true,
			Ignore:   utils.ToStringSlice(opts["ignore"]),
		})
	},
}

// The fix uses POSIX separators on every host. tspath also normalizes literal
// backslashes and drive letters, which would change these module specifiers.
func posixPathComponents(value, cwd string) []string {
	if !path.IsAbs(value) {
		value = path.Join(cwd, value)
	}
	return strings.FieldsFunc(path.Clean(value), func(r rune) bool { return r == '/' })
}

func relativeImportPath(fromParts []string, to, cwd string) string {
	toParts := posixPathComponents(to, cwd)
	common := 0
	for common < len(fromParts) && common < len(toParts) && fromParts[common] == toParts[common] {
		common++
	}
	relative := strings.TrimSuffix(strings.Repeat("../", len(fromParts)-common)+strings.Join(toParts[common:], "/"), "/")
	if !strings.HasPrefix(relative, ".") {
		relative = "./" + relative
	}
	return relative
}
