package moduleresolver

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/tailscale/hujson"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// A TS-ending marker reaches the package field before tsgo's primary pass
// can fall back to an unrelated index.ts while waiting for its JS pass.
const implementationMainSuffix = ".__rslint_node_target__.ts"

type implementationKey struct {
	options    *core.CompilerOptions
	name, file string
	mode       core.ResolutionMode
}

type implementationMain struct {
	directory string
	indexOnly bool
}

// ResolveImplementation follows a compiler-resolved request to its runtime
// implementation. Source aliases and root directories retain TypeScript's file
// probes; package entries retain their authored targets, including explicit TS
// files, instead of substituting a same-named source for an existing JS entry.
func ResolveImplementation(p *program.Program, sourceFile *ast.SourceFile, source modules.Source) string {
	if !p.IsValid() || !p.OwnsSourceFile(sourceFile) || source.Specifier() == nil || !ast.IsStringLiteralLike(source.Specifier()) {
		return ""
	}
	options := p.OptionsForFile(sourceFile)
	mode := p.GetModeForUsageLocation(sourceFile, source.Specifier())
	name, file := source.Specifier().Text(), sourceFile.FileName()
	return program.Cached(p, implementationKey{options, name, file, mode}, func() string {
		options := options.Clone()
		options.NoDtsResolution = core.TSTrue
		options.ResolveJsonModule = core.TSTrue
		options.CustomConditions = slices.DeleteFunc(slices.Clone(options.CustomConditions), func(condition string) bool {
			return condition == "types" || strings.HasPrefix(condition, "types@")
		})
		paths := &collections.OrderedMap[string, []string]{}
		for name, targets := range options.Paths.Entries() {
			targets := slices.DeleteFunc(slices.Clone(targets), tspath.IsDeclarationFileName)
			if len(targets) != 0 {
				paths.Set(name, targets)
			}
		}
		options.Paths = paths
		runtimeOptions := Options{
			Extensions:       []string{".js", ".json", ".node"},
			Conditions:       module.GetConditions(options, mode),
			NodeExports:      true,
			PreserveSymlinks: options.PreserveSymlinks == core.TSTrue,
		}
		resolver := &nodeResolver{program: p, fileName: file, options: runtimeOptions}
		view := &nodeResolutionFS{
			FS: p.FS(), folder: "node_modules", base: tspath.GetDirectoryPath(file),
			options: runtimeOptions, resolver: resolver, request: name,
			resolved: map[string]string{}, compilerOptions: options,
			implementationMains: map[string]implementationMain{},
		}
		resolved, _ := view.newResolver(p.CurrentDirectory()).ResolveModuleName(name, file, mode, nil)
		if view.unresolved || view.failure.Error != "" || resolved == nil || !resolved.IsResolved() || view.builtin {
			return ""
		}
		return view.Realpath(resolved.ResolvedFileName)
	})
}

// tsgo inserts moduleSuffixes immediately before the synthetic marker's
// extension. Recover the package target; its suffixes are probed in order
// after URL decoding, before trying another package target or legacy index.
func (f *nodeResolutionFS) implementationMarker(name string) string {
	if f.compilerOptions == nil {
		return name
	}
	suffixes := f.compilerOptions.ModuleSuffixes
	if len(suffixes) == 0 {
		suffixes = []string{""}
	}
	for _, marker := range []string{implementationMainSuffix, nodeTargetSuffix, nodeExportSuffix, nodeDirectoryExportSuffix} {
		extension := tspath.TryGetExtensionFromPath(marker)
		stem := strings.TrimSuffix(marker, extension)
		canonical := marker
		if marker == implementationMainSuffix {
			canonical = nodeTargetSuffix
		}
		for _, suffix := range suffixes {
			ending := stem + suffix + extension
			if strings.HasSuffix(name, ending) {
				return strings.TrimSuffix(name, ending) + canonical
			}
		}
	}
	return name
}

func (f *nodeResolutionFS) probeImplementationTarget(name string, exact, applyAlias bool) string {
	if f.compilerOptions == nil {
		if exact {
			return f.probeFile(name)
		}
		return f.probe(name, applyAlias)
	}
	suffixes := f.compilerOptions.ModuleSuffixes
	if len(suffixes) == 0 {
		suffixes = []string{""}
	}
	for _, suffix := range suffixes {
		if result := f.probeFile(implementationSuffix(name, suffix)); result != "" {
			return result
		}
	}
	if !exact && !f.fullySpecified {
		// Like the compiler's ordinary probes, finish a runtime extension's
		// suffixes before trying the next extension.
		for _, extension := range f.options.Extensions {
			if result := f.probeImplementationTarget(name+extension, true, applyAlias); result != "" {
				return result
			}
		}
	}
	return ""
}

// A package main directory uses Node's index lookup, not another source
// resolution. In particular it neither replaces index.js with index.ts nor
// follows a nested package.json main. Node does not probe an extensionless index.
func (f *nodeResolutionFS) probeImplementationIndex(directory string) string {
	for _, extension := range f.options.Extensions {
		if result := f.probeImplementationTarget(tspath.ResolvePath(directory, "index"+extension), true, false); result != "" {
			return result
		}
	}
	return ""
}

func implementationSuffix(name, suffix string) string {
	extension := tspath.TryGetExtensionFromPath(name)
	return strings.TrimSuffix(name, extension) + suffix + extension
}

// A synthetic main also owns legacy index fallback. Otherwise tsgo would
// prefer index.ts over the runtime index.js after a missing package main.
func (f *nodeResolutionFS) prepareImplementationMain(object *hujson.Object, directory string) {
	if f.compilerOptions == nil {
		return
	}
	target := ""
	for i := range object.Members {
		member := &object.Members[i]
		if nodePackageMemberName(*member) != "main" {
			continue
		}
		if literal, ok := member.Value.Value.(hujson.Literal); ok && literal.Kind() == '"' {
			target = literal.String()
		}
		indexOnly := target == ""
		if indexOnly {
			target = "./index"
			member.Value.Value = hujson.String(target)
		}
		f.implementationMains[tspath.ResolvePath(directory, target)] = implementationMain{directory, indexOnly}
		return
	}
	object.Members = append(object.Members, hujson.ObjectMember{
		Name:  hujson.Value{Value: hujson.String("main")},
		Value: hujson.Value{Value: hujson.String("./index")},
	})
	f.implementationMains[tspath.ResolvePath(directory, "index")] = implementationMain{directory, true}
}
