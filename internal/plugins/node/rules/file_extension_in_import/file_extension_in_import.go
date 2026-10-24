package file_extension_in_import

import (
	_ "embed"
	"os"
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
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
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

type directoryEntries struct {
	names    []string
	complete bool
}

func hostDirectoryNames(fs vfs.FS, directory string) ([]string, bool) {
	// Stat through the Program first to preserve its path mapping and network
	// rejection. Accessible entries alone may omit broken links or special files.
	info := fs.Stat(pathutil.Path(directory))
	if info == nil || info.Sys() == nil {
		return nil, false
	}
	handle, err := os.Open(fs.Realpath(pathutil.Path(directory)).AsString())
	if err != nil {
		return nil, false
	}
	defer handle.Close()
	hostInfo, err := handle.Stat()
	if err != nil || !os.SameFile(info, hostInfo) {
		return nil, false
	}
	// Validate and enumerate the same opened directory, even if its alias moves.
	names, err := handle.Readdirnames(-1)
	return names, err == nil
}

func existingExtensions(p *program.Program, filePath, basename string) ([]string, bool) {
	directory := filepath.Dir(filePath)
	if filepath.Separator == '/' && strings.Contains(directory, `\`) {
		return nil, false
	}

	directory = tspath.NormalizePath(directory)
	entries := program.Cached(p, directoryEntriesKey(directory), func() directoryEntries {
		fs := p.FS()
		entries := fs.GetAccessibleEntries(pathutil.Directory(directory))
		names := append(slices.Clone(entries.Files), entries.Directories...)
		hostNames, complete := hostDirectoryNames(fs, directory)
		if complete {
			names = append(names, hostNames...)
		}
		slices.Sort(names)
		return directoryEntries{names: slices.Compact(names), complete: complete}
	})
	prefix := basename + "."
	start, _ := slices.BinarySearch(entries.names, prefix)
	var extensions []string
	for _, name := range entries.names[start:] {
		if !strings.HasPrefix(name, prefix) {
			break
		}
		extensions = append(extensions, extension(name))
	}
	return extensions, entries.complete
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
		if p == nil || ctx.SourceFile == nil || strings.HasPrefix(ctx.SourceFile.FileName().AsString(), "<") {
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
		var resolutions [2]*moduleresolver.Options
		return nodeutil.VisitImports(ctx, nodeutil.ImportVisitorOptions{}, func(source *ast.Node, name string, typeOnly bool) {
			if !tspath.PathIsRelative(name) && !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, `\`) {
				return
			}
			// VisitImports already removes loader suffixes. Resolve and check only
			// the path, keeping queries and fragments out of extension decisions.
			if index := strings.IndexAny(name, "?#"); index >= 0 {
				name = name[:index]
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
			filePath := nodeutil.ImportFilePath(p, name, ctx.SourceFile.FileName().AsString(), typeOnly, *resolution)
			if filePath == "" {
				return
			}
			currentExt, actualExt := extension(name), extension(filePath)
			fs := p.FS()
			normalizedPath := tspath.NormalizePath(filePath)
			// A literal POSIX backslash must not alias a slash-separated VFS path.
			canProbe := filepath.Separator == '\\' || !strings.Contains(filePath, `\`)
			if actualExt != "" && (!canProbe || !fs.FileExists(pathutil.File(normalizedPath))) {
				// A fallback such as utils.client may actually name utils.client.ts.
				basename := filepath.Base(name)
				existing, _ := existingExtensions(p, filePath, basename)
				if found := preferredExtension(existing, resolution.Extensions); found != "" {
					actualExt = found
				}
			}
			isDirectory := false
			if canProbe && fs.DirectoryExists(pathutil.Directory(normalizedPath)) {
				existing, _ := existingExtensions(p, filepath.Join(filePath, "index"), "index")
				if found := preferredExtension(existing, resolution.Extensions); found != "" {
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
					if !canProbe || source.Kind != ast.KindStringLiteral || ast.IsUnterminatedLiteral(source) {
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
					if len(name) < len(source.Text()) {
						span := utils.TrimNodeTextRange(ctx.SourceFile, source)
						raw := ctx.SourceFile.Text()[span.Pos():span.End()]
						offset := len(name) + 1
						if raw[1:len(raw)-1] != source.Text() {
							units := utils.ParseJSStringLiteralSource(raw)
							index := ecmascript.StringCodeUnitCount(name)
							if index >= len(units) {
								return nil
							}
							offset = units[index].Start
						}
						position = span.Pos() + offset
					}
					return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(position, position), text)}
				})
			}
			if style == "never" && currentExt != "" && currentExt == expectedExt {
				ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
					Id: "forbidExt", Description: "forbid file extension '" + currentExt + "'.",
					Data: map[string]string{"ext": currentExt},
				}, func() []rule.RuleFix {
					if !canProbe || source.Kind != ast.KindStringLiteral || ast.IsUnterminatedLiteral(source) {
						return nil
					}
					basename := strings.TrimSuffix(filepath.Base(filePath), extension(filePath))
					existing, complete := existingExtensions(p, filePath, basename)
					if !complete || len(existing) != 1 {
						return nil
					}
					span := utils.TrimNodeTextRange(ctx.SourceFile, source)
					raw := ctx.SourceFile.Text()[span.Pos():span.End()]
					index := strings.LastIndex(name, currentExt)
					start, end := index+1, index+1+len(currentExt)
					if raw[1:len(raw)-1] != source.Text() {
						// Decoded positions count UTF-16 units; edits address source bytes.
						units := utils.ParseJSStringLiteralSource(raw)
						first := ecmascript.StringCodeUnitCount(name[:index])
						last := first + ecmascript.StringCodeUnitCount(currentExt)
						if last > len(units) {
							return nil
						}
						start, end = units[first].Start, units[last-1].End
					}
					return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(span.Pos()+start, span.Pos()+end))}
				})
			}
		})
	},
}
