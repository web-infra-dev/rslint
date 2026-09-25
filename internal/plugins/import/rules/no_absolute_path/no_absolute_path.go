package no_absolute_path

import (
	_ "embed"
	"encoding/json"
	"path"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
				text, ok := quoteModulePath(relative)
				if !ok {
					return nil
				}
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, source, text)}
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
	// Without a parent-directory segment, even a name like .hidden needs ./.
	if common == len(fromParts) {
		relative = "./" + relative
	}
	return relative
}

// JSON handles ordinary text; preserve lone surrogates as JS Unicode escapes
// instead of letting the encoder replace their compiler WTF-8 representation.
func quoteModulePath(value string) (string, bool) {
	if utf8.ValidString(value) {
		text, err := json.Marshal(value)
		return string(text), err == nil
	}
	var quoted strings.Builder
	quoted.WriteByte('"')
	start := 0
	for offset := 0; offset < len(value); {
		r, size := ecmascript.DecodeStringRune(value[offset:])
		if r == utf8.RuneError && size == 1 {
			return "", false
		}
		if utf16.IsSurrogate(r) {
			text, err := json.Marshal(value[start:offset])
			if err != nil {
				return "", false
			}
			quoted.Write(text[1 : len(text)-1])
			quoted.WriteString(`\u`)
			quoted.WriteString(strconv.FormatInt(int64(r), 16))
			start = offset + size
		}
		offset += size
	}
	text, err := json.Marshal(value[start:])
	if err != nil {
		return "", false
	}
	quoted.Write(text[1:])
	return quoted.String(), true
}
