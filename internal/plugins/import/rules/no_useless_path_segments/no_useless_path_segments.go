package no_useless_path_segments

import (
	_ "embed"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed no_useless_path_segments.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-useless-path-segments.js
var NoUselessPathSegmentsRule = rule.Rule{
	Name:   "import/no-useless-path-segments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var commonjs, noUselessIndex bool
		if len(options) > 0 {
			option, _ := options[0].(map[string]any)
			commonjs, _ = option["commonjs"].(bool)
			noUselessIndex, _ = option["noUselessIndex"].(bool)
		}
		var resolver *import_utils.ImportResolver
		reportedResolverError := false
		resolve := func(name string, source *ast.Node) (string, bool) {
			if resolver == nil {
				resolver = import_utils.NewImportResolver(ctx)
			}
			resolved, found, err := resolver.ResolveName(name, source)
			if err != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + err})
				reportedResolverError = true
			}
			return resolved, found
		}
		var extensions []string
		var indexPattern *esregexp.RegExp
		indexPatternInitialized := false
		return import_utils.VisitModules(func(source, _ *ast.Node) {
			importPath := source.Text()
			if !strings.HasPrefix(importPath, ".") {
				return
			}
			report := func(proposed string) {
				ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
					Description: `Useless path segments for "` + importPath + `", should be "` + proposed + `"`,
				}, func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, source, quotePath(proposed))}
				})
			}
			resolved, found := resolve(importPath, source)
			normalized := toRelativePath(path.Clean(importPath))
			if normalized != importPath {
				resolvedNormalized, foundNormalized := resolve(normalized, source)
				// Two unresolved paths compare equal upstream as well.
				if found == foundNormalized && resolved == resolvedNormalized {
					report(normalized)
					return
				}
			}
			if noUselessIndex && !indexPatternInitialized {
				extensions = import_utils.FileExtensions(ctx.Settings)
				// Match the upstream pattern, including settings-authored regex syntax.
				indexPattern, _ = esregexp.Compile(indexPatternSource(extensions), "")
				indexPatternInitialized = true
			}
			unnecessaryIndex := false
			if indexPattern != nil {
				matched, err := indexPattern.Unwrap().MatchRunes(ecmascript.StringCodeUnitRunes(importPath))
				// A timeout cannot justify a diagnostic or an autofix.
				unnecessaryIndex = err == nil && matched
			}
			if unnecessaryIndex {
				// Split preserves the written ./ and parent segments, as Node's
				// dirname does; filepath.Dir would clean them first.
				parent, _ := filepath.Split(strings.TrimRightFunc(importPath, func(r rune) bool {
					return r == '/' || r == filepath.Separator
				}))
				if parent == "" {
					parent = "."
				} else {
					parent = parent[:len(parent)-1]
				}
				if parent != "." && parent != ".." {
					for _, extension := range extensions {
						if candidate, ok := resolve(parent+extension, source); ok && candidate != "" {
							report(parent + "/")
							return
						}
					}
				}
				report(parent)
				return
			}
			if strings.HasPrefix(importPath, "./") || !found || resolved == "" {
				return
			}
			expected, err := filepath.Rel(filepath.Dir(ctx.SourceFile.FileName()), resolved)
			if err != nil {
				return
			}
			segments := strings.Split(strings.TrimPrefix(importPath, "./"), "/")
			parents := countParents(segments)
			expectedParents := countParents(strings.Split(filepath.ToSlash(expected), "/"))
			diff := parents - expectedParents
			if diff > 0 {
				// JavaScript slice permits an offset beyond the end.
				suffix := segments[min(parents+diff, len(segments)):]
				report(toRelativePath(strings.Join(append(segments[:expectedParents:expectedParents], suffix...), "/")))
			}
		}, import_utils.VisitModulesOptions{ESModule: true, Commonjs: commonjs})
	},
}

// The upstream regexp has no Unicode flag. Encode literal surrogates in its
// source too, so character classes and quantifiers see the same UTF-16 units
// as the subject. Keep escaped characters together when replacing a unit.
func indexPatternSource(extensions []string) string {
	units := ecmascript.StringCodeUnitRunes(`.*/index(\` + strings.Join(extensions, `|\`) + `)?$`)
	var source strings.Builder
	for i := 0; i < len(units); i++ {
		unit := units[i]
		if unit == '\\' && i+1 < len(units) {
			i++
			unit = units[i]
			if unit < 0xD800 || unit > 0xDFFF {
				source.WriteByte('\\')
			}
		}
		if unit >= 0xD800 && unit <= 0xDFFF {
			fmt.Fprintf(&source, `\u%04x`, unit)
		} else {
			source.WriteRune(unit)
		}
	}
	return source.String()
}

func toRelativePath(value string) string {
	// Unlike tspath's relative-path helpers, upstream treats backslashes as
	// literal characters here, even on Windows (path.posix.normalize).
	value = strings.TrimSuffix(value, "/")
	if value == "." || value == ".." || strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") {
		return value
	}
	return "./" + value
}

func countParents(segments []string) int {
	count := 0
	for _, segment := range segments {
		if segment == ".." {
			count++
		}
	}
	return count
}

// Use tsgo's JSON quoting, preserving lone UTF-16 surrogates that its JSON
// encoder would replace with U+FFFD. AST string values encode these as WTF-8.
func quotePath(value string) string {
	if utf8.ValidString(value) {
		return utils.Must(core.StringifyJson(value, "", ""))
	}
	var out strings.Builder
	out.WriteByte('"')
	start := 0
	for offset := 0; offset < len(value); {
		r, size := ecmascript.DecodeStringRune(value[offset:])
		if r >= 0xD800 && r <= 0xDFFF {
			quoted := utils.Must(core.StringifyJson(value[start:offset], "", ""))
			out.WriteString(quoted[1 : len(quoted)-1])
			fmt.Fprintf(&out, `\u%04x`, r)
			start = offset + size
		}
		offset += size
	}
	quoted := utils.Must(core.StringifyJson(value[start:], "", ""))
	out.WriteString(quoted[1:])
	return out.String()
}
