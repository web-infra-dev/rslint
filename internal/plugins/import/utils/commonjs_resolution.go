package utils

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
)

// Source mappings select source implementations; package entries select runtime
// files. In particular, moduleSuffixes cannot rewrite a package's JS entry.
func resolveRuntimeImplementation(p *program.Program, name, containingFile string, mode core.ResolutionMode) string {
	options := p.Options().Clone()
	options.NoDtsResolution = core.TSTrue
	options.ResolveJsonModule = core.TSTrue
	options.CustomConditions = slices.DeleteFunc(slices.Clone(options.CustomConditions), func(condition string) bool {
		return condition == "types" || strings.HasPrefix(condition, "types@")
	})
	if path := resolveImplementationSource(p, options, name, containingFile, mode); path != "" {
		return path
	}
	return moduleresolver.ResolveModule(p, name, containingFile, moduleresolver.Options{
		Extensions:       []string{".js", ".json", ".node"},
		MainFiles:        []string{"index.js", "index.json", "index.node"},
		Conditions:       module.GetConditions(options, mode),
		NodeExports:      true,
		PreserveSymlinks: options.PreserveSymlinks == core.TSTrue,
	})
}

func resolveImplementationSource(p *program.Program, options *core.CompilerOptions, name, containingFile string, mode core.ResolutionMode) string {
	type candidate struct {
		path  string
		exact bool
	}
	var candidates []candidate
	if tspath.IsExternalModuleNameRelative(name) {
		if len(options.RootDirs) != 0 {
			candidates = []candidate{{path: name}}
		}
	} else if options.Paths != nil && options.Paths.Size() != 0 {
		pattern := module.MatchPatternOrExact(module.TryParsePatterns(options.Paths), name)
		if pattern.IsValid() {
			base := options.GetPathsBasePath(p.CurrentDirectory())
			for _, target := range options.Paths.GetOrZero(pattern.Text) {
				exact := tspath.TryGetExtensionFromPath(target) != ""
				target = strings.Replace(target, "*", pattern.MatchedText(name), 1)
				// A declaration-only alias does not name a runtime file. Do not
				// substitute its neighboring JS file for an unrelated package.
				if !tspath.IsDeclarationFileName(target) {
					candidates = append(candidates, candidate{tspath.ResolvePath(base, target), exact})
				}
			}
		}
		options.RootDirs = nil
	}
	if len(candidates) == 0 {
		return ""
	}
	options.Paths = nil
	host := rslint_utils.CreateCompilerHost(p.CurrentDirectory(), p.FS())
	resolver := module.NewResolver(host, options, "", "", nil)
	for _, candidate := range candidates {
		// tsgo's paths loader first tries an explicitly suffixed substitution
		// as a file. Ordinary relative resolution substitutes TS before JS.
		if candidate.exact {
			suffixes := options.ModuleSuffixes
			if len(suffixes) == 0 {
				suffixes = []string{""}
			}
			extension := tspath.TryGetExtensionFromPath(candidate.path)
			for _, suffix := range suffixes {
				path := tspath.RemoveExtension(candidate.path, extension) + suffix + extension
				if p.FS().FileExists(path) {
					if options.PreserveSymlinks != core.TSTrue && strings.Contains(path, "/node_modules/") {
						path = p.FS().Realpath(path)
					}
					return path
				}
			}
		}
		resolved, _ := resolver.ResolveModuleName(candidate.path, containingFile, mode, nil)
		if resolved != nil && resolved.IsResolved() && !tspath.IsDeclarationFileName(resolved.ResolvedFileName) {
			return resolved.ResolvedFileName
		}
	}
	return ""
}
