package file_extension_in_import

import (
	_ "embed"
	"path/filepath"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed file_extension_in_import.schema.json
var schemaJSON []byte

// extension follows Node's path.extname: trailing separators are ignored and
// a leading dot alone does not introduce an extension (e.g. .env and ..).
func extension(name string) string {
	base := filepath.Base(name)
	if base == ".." || strings.LastIndexByte(base, '.') <= 0 {
		return ""
	}
	return filepath.Ext(base)
}

type directoryEntriesKey string

func existingExtensions(p *program.Program, filePath, basename string) []string {
	directory := filepath.Dir(filePath)
	if filepath.Separator == '/' && strings.Contains(directory, `\`) {
		return nil
	}
	directory = tspath.NormalizePath(directory)
	names := program.Cached(p, directoryEntriesKey(directory), func() []string {
		fs := p.FS()
		entries := fs.GetAccessibleEntries(directory)
		names := append(slices.Clone(entries.Files), entries.Directories...)
		// Include broken symlinks and special entries, as readdir does.
		// Accessible entries also supply files present only in editor overlays.
		_ = fs.WalkDir(directory, func(path string, entry vfs.DirEntry, err error) error {
			if err != nil || entry == nil || path == directory {
				return err
			}
			names = append(names, entry.Name())
			if entry.IsDir() {
				return vfs.SkipDir
			}
			return nil
		})
		slices.Sort(names)
		return slices.Compact(names)
	})
	prefix := basename + "."
	start, _ := slices.BinarySearch(names, prefix)
	var extensions []string
	for _, name := range names[start:] {
		if !strings.HasPrefix(name, prefix) {
			break
		}
		extensions = append(extensions, extension(name))
	}
	return extensions
}

func preferredExtension(existing, preferred []string) string {
	for _, ext := range preferred {
		if slices.Contains(existing, ext) {
			return ext
		}
	}
	if len(existing) > 0 {
		return existing[0]
	}
	return ""
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/file-extension-in-import.js
var FileExtensionInImportRule = rule.Rule{
	Name:   "node/file-extension-in-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		p := ctx.Program()
		if p == nil || ctx.SourceFile == nil || strings.HasPrefix(ctx.SourceFile.FileName(), "<") {
			return nil
		}
		defaultStyle := "always"
		if len(options) > 0 {
			defaultStyle, _ = options[0].(string)
		}
		var overrides map[string]any
		if len(options) > 1 {
			overrides, _ = options[1].(map[string]any)
		}
		var resolutions [2]*nodeutil.ResolutionOptions
		return nodeutil.VisitImports(nodeutil.ImportVisitorOptions{}, func(source *ast.Node, name string, typeOnly bool) {
			if !tspath.PathIsRelative(name) && !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, `\`) {
				return
			}
			index := 0
			if typeOnly {
				index = 1
			}
			if resolutions[index] == nil {
				resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, nil)
				resolutions[index] = &resolution
			}
			resolution := resolutions[index]
			filePath := nodeutil.ImportFilePath(p, name, ctx.SourceFile.FileName(), typeOnly, *resolution)
			if filePath == "" {
				return
			}
			currentExt, actualExt := extension(name), extension(filePath)
			fs := p.FS()
			normalizedPath := tspath.NormalizePath(filePath)
			// A literal POSIX backslash must not alias a slash-separated VFS path.
			canProbe := filepath.Separator == '\\' || !strings.Contains(filePath, `\`)
			if actualExt != "" && (!canProbe || !fs.FileExists(normalizedPath)) {
				// A fallback such as utils.client may actually name utils.client.ts.
				basename := filepath.Base(name)
				if found := preferredExtension(existingExtensions(p, filePath, basename), resolution.Extensions); found != "" {
					actualExt = found
				}
			}
			isDirectory := false
			if canProbe && fs.DirectoryExists(normalizedPath) {
				if found := preferredExtension(existingExtensions(p, filepath.Join(filePath, "index"), "index"), resolution.Extensions); found != "" {
					isDirectory = true
					actualExt = found
				}
			}
			style := defaultStyle
			if override, ok := overrides[actualExt].(string); ok {
				style = override
			}
			expectedExt := nodeutil.ImportExtension(ctx, actualExt)
			if style == "always" && currentExt != expectedExt {
				ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
					Id: "requireExt", Description: "require file extension '" + expectedExt + "'.",
					Data: map[string]string{"ext": expectedExt},
				}, func() []rule.RuleFix {
					if source.Kind != ast.KindStringLiteral {
						return nil
					}
					quote := ctx.SourceFile.Text()[source.End()-1]
					// Custom mappings can contain characters unsafe to insert verbatim.
					if (quote != '\'' && quote != '"') || strings.ContainsAny(expectedExt, "\\\r\n") || strings.ContainsRune(expectedExt, rune(quote)) {
						return nil
					}
					text := expectedExt
					if isDirectory {
						text = "index" + text
						if !strings.HasSuffix(name, "/") && !strings.HasSuffix(name, `\`) {
							text = "/" + text
						}
					}
					position := source.End() - 1
					return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(position, position), text)}
				})
			}
			if style == "never" && currentExt != "" && currentExt == expectedExt {
				ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
					Id: "forbidExt", Description: "forbid file extension '" + currentExt + "'.",
					Data: map[string]string{"ext": currentExt},
				}, func() []rule.RuleFix {
					basename := strings.TrimSuffix(filepath.Base(filePath), extension(filePath))
					if len(existingExtensions(p, filePath, basename)) != 1 {
						return nil
					}
					span := utils.TrimNodeTextRange(ctx.SourceFile, source)
					// Upstream offsets decoded strings into raw tokens, which can
					// remove unrelated characters when the path contains escapes.
					// Keep the diagnostic but omit that unsafe fix.
					if source.Kind != ast.KindStringLiteral || ctx.SourceFile.Text()[span.Pos()+1:span.End()-1] != source.Text() {
						return nil
					}
					start := span.Pos() + 1 + strings.LastIndex(name, currentExt)
					return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, start+len(currentExt)))}
				})
			}
		})
	},
}
