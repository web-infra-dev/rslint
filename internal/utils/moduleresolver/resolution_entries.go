package moduleresolver

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// MainField selects a package property path for an entry point.
// ForceRelative interprets a bare target relative to its package directory.
type MainField struct {
	Name          []string `json:"name"`
	ForceRelative bool     `json:"forceRelative"`
}

// A lazy main target lets tsgo choose the package and apply exports before
// consulting configured entry fields. It never enters the Program filesystem.
const nodeMainTarget = ".__rslint_node_main__"

func (resolver *nodeResolver) resolveAt(name, directory string, mainTarget bool) Result {
	child := *resolver
	child.mainTarget = mainTarget
	child.fileName = tspath.ResolvePath(directory, "__import__.js")
	// Package fields resolve from their owning package, not resolvePaths.
	child.options.Paths = nil
	return child.resolve(name)
}

func (resolver *nodeResolver) mainEntry(directory string) Result {
	pkg := packagejson.Read(resolver.program, directory)
	if pkg != nil {
		for _, field := range resolver.options.MainFields {
			target, ok := pkg.Field(field.Name...).(string)
			if !ok || target == "" || target == "." || target == "./" {
				continue
			}
			if field.ForceRelative && !strings.HasPrefix(target, "./") && !strings.HasPrefix(target, "../") {
				target = "./" + target
			}
			// Backslashes cannot become directory separators in POSIX entries.
			if strings.Contains(target, `\`) && tspath.GetRootLength(directory) == 1 && !tspath.IsRootedDiskPath(target) {
				continue
			}
			result := resolver.resolveAt(tspath.ResolvePath(directory, target), directory, true)
			if result.Error == "" || result.recursive {
				return result
			}
		}
	}
	return Result{Error: "No matching package entry"}
}

func (resolver *nodeResolver) aliasField(request, directory string, file bool) (Result, bool) {
	if len(resolver.options.AliasFields) == 0 {
		return Result{}, false
	}
	pkg := packagejson.FindNearest(resolver.program, tspath.ResolvePath(directory, "__import__.js"))
	if pkg == nil {
		return Result{}, false
	}
	inner := request
	if file || tspath.PathIsRelative(request) {
		if !file && strings.Contains(request, `\`) && tspath.GetRootLength(directory) == 1 {
			return Result{}, false
		}
		absolute := tspath.ResolvePath(directory, request)
		inner = tspath.GetRelativePathFromDirectory(pkg.Directory(), absolute, tspath.ComparePathsOptions{
			UseCaseSensitiveFileNames: resolver.program.FS().UseCaseSensitiveFileNames(),
		})
		if file {
			inner = "./" + inner
		}
	}
	for _, field := range resolver.options.AliasFields {
		aliases, ok := pkg.Field(field...).(map[string]any)
		if !ok {
			continue
		}
		target, present := aliases[inner]
		if !present && strings.HasPrefix(inner, "./") {
			target, present = aliases[inner[2:]]
		}
		if !present || target == inner {
			continue
		}
		if target == false {
			return Result{}, true
		}
		if name, ok := target.(string); ok {
			if name == "" {
				name = "."
			}
			return resolver.resolveAt(name, pkg.Directory(), false), true
		}
		return Result{Error: "Invalid package alias for '" + inner + "'"}, true
	}
	return Result{}, false
}

// Only implicit index probes represent mainFiles. A request for index or
// index.js must keep its filename, including when the package is named index.
func (f *nodeResolutionFS) isDirectoryIndex(logical, physical string) bool {
	if !strings.HasSuffix(physical, "/index.js") {
		return false
	}
	if len(f.activeDirectories) != 0 {
		return true
	}
	canonical := func(name string) string {
		return tspath.GetCanonicalFileName(tspath.NormalizePath(name), f.UseCaseSensitiveFileNames())
	}
	request := canonical(f.request)
	if tspath.IsExternalModuleNameRelative(f.request) {
		request = canonical(tspath.ResolvePath(f.base, f.request))
		physical = canonical(physical)
		return physical != request && physical != request+".js"
	}
	logical = canonical(logical)
	return !strings.HasSuffix(logical, "/node_modules/"+request) && !strings.HasSuffix(logical, "/node_modules/"+request+".js")
}
