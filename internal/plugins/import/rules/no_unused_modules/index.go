package no_unused_modules

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// projectIndex is the usage of every analyzed module in one source generation,
// built once for one rule configuration before any file reports from it.
//
// The analyzed files are fixed by `src` alone. Unlike upstream, linting a file
// outside `src` never adds its imports, so the result does not depend on which
// files are linted or in which order. Every file is read from the pass's own
// filesystem view, so API overlays and autofix rounds are seen consistently.
type projectIndex struct {
	caseSensitive bool
	src           map[string]struct{}
	ignore        *fileSelection
	usage         map[string]*usage
	realpath      func(string) string
}

// indexKey identifies one index of a pass. Every input that changes which
// files are analyzed or how their requests resolve is part of it.
type indexKey struct {
	cwd           string
	src           string
	ignoreExports string
	extensions    string
	settings      string
}

func newIndexKey(cwd string, opts options, extensions []string, settings map[string]any) indexKey {
	encodedSettings, err := json.Marshal(settings)
	if err != nil {
		// Settings JSON cannot encode never come from configuration; keep such a
		// rule test from sharing an index with any other configuration.
		encodedSettings = []byte{0}
	}
	return indexKey{
		cwd:           cwd,
		src:           strings.Join(opts.src, "\x00"),
		ignoreExports: strings.Join(opts.ignoreExports, "\x00"),
		extensions:    strings.Join(extensions, "\x00"),
		settings:      string(encodedSettings),
	}
}

func (index *projectIndex) key(fileName string) string {
	return string(tspath.ToPath(index.realpath(fileName), "", index.caseSensitive))
}

func (index *projectIndex) analyzes(fileName string) bool {
	_, ok := index.src[index.key(fileName)]
	return ok
}

func (index *projectIndex) ignored(fileName string) bool {
	return index.ignore.matches(fileName)
}

func (index *projectIndex) usageOf(fileName string) *usage {
	return index.usage[index.key(fileName)]
}

type indexInput struct {
	fs         vfs.FS
	cwd        string
	opts       options
	extensions []string
	settings   map[string]any
	compiled   *import_utils.ModuleSettings
	// fallback provides filesystem lookup for importers no Program owns.
	fallback *program.Program
}

func buildProjectIndex(scope *rule.GenerationScope, input indexInput) *projectIndex {
	fs := input.fs
	index := &projectIndex{
		caseSensitive: fs.UseCaseSensitiveFileNames(),
		src:           make(map[string]struct{}),
		ignore:        newFileSelection(fs, input.cwd, input.opts.ignoreExports, input.extensions),
		usage:         make(map[string]*usage),
		realpath:      fs.Realpath,
	}
	files := newFileSelection(fs, input.cwd, input.opts.src, input.extensions).enumerate()
	for _, fileName := range files {
		index.src[index.key(fileName)] = struct{}{}
	}

	settings := input.settings
	moduleSettings := input.compiled
	results := make([]map[string]*usage, len(files))
	analyze := func(position int) {
		fileName := files[position]
		// Upstream builds no export map, and so records no imports, for a
		// file matching import/ignore or outside the analyzed extensions.
		if moduleSettings.IsIgnoredPath(fileName) || !hasAnyExtension(fileName, input.extensions) {
			return
		}
		sourceProgram, file, owned := importerSource(scope, input.fallback, fs, fileName)
		if file == nil {
			return
		}
		references := collectReferences(file)
		if len(references) == 0 {
			return
		}
		resolver := import_utils.NewImportResolverForFile(sourceProgram, file, owned, settings, input.cwd)
		found := make(map[string]*usage)
		for _, reference := range references {
			specifier := reference.source.Specifier()
			if specifier == nil {
				continue
			}
			resolved, ok, _ := resolver.ResolveName(specifier.Text(), reference.source)
			if !ok || resolved == "" || strings.Contains(resolved, "/node_modules/") {
				continue
			}
			key := index.key(resolved)
			target := found[key]
			if target == nil {
				target = &usage{}
				found[key] = target
			}
			target.merge(&reference.usage)
		}
		results[position] = found
	}
	if scope.SingleThreaded() || len(files) < 2 {
		for position := range files {
			analyze(position)
		}
	} else {
		var wait sync.WaitGroup
		next := make(chan int)
		for range min(runtime.GOMAXPROCS(0), len(files)) {
			wait.Go(func() {
				for position := range next {
					analyze(position)
				}
			})
		}
		for position := range files {
			next <- position
		}
		close(next)
		wait.Wait()
	}
	for _, found := range results {
		for key, used := range found {
			target := index.usage[key]
			if target == nil {
				target = &usage{}
				index.usage[key] = target
			}
			target.merge(used)
		}
	}
	return index
}

// importerSource returns the AST of an analyzed file and the Program to
// resolve its requests with. A Program of the pass that owns the file supplies
// both, so its tsconfig applies to the file's requests; the first one wins.
// Any other file is parsed alone and resolved without a project.
func importerSource(scope *rule.GenerationScope, fallback *program.Program, fs vfs.FS, fileName string) (*program.Program, *ast.SourceFile, bool) {
	for _, sourceProgram := range scope.Programs() {
		if file := sourceProgram.GetSourceFile(fileName); file != nil {
			return sourceProgram, file, true
		}
	}
	text, ok := fs.ReadFile(fileName)
	if !ok {
		return nil, nil, false
	}
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.ToPath(fileName, "", fs.UseCaseSensitiveFileNames()),
	}, text, core.GetScriptKindFromFileName(fileName))
	return fallback, file, false
}

func hasAnyExtension(fileName string, extensions []string) bool {
	for _, extension := range extensions {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}
