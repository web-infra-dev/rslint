package no_unused_modules

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// isPackageEntry reports whether the nearest package.json publishes fileName
// as an entry point of a non-private package. Upstream checks `bin`,
// `browser` and `main`, each naming the exact file; this port also accepts
// every target listed in `exports`, which modern packages use instead.
func isPackageEntry(sourceProgram *program.Program, fileName string) bool {
	pkg := packagejson.FindNearest(sourceProgram, fileName)
	if pkg == nil || pkg.Field("private") == true {
		return false
	}
	directory := pkg.Directory()
	names := func(field any) bool {
		switch value := field.(type) {
		case string:
			return tspath.ResolvePath(directory, value) == fileName
		case map[string]any:
			for _, target := range value {
				if target, ok := target.(string); ok && tspath.ResolvePath(directory, target) == fileName {
					return true
				}
			}
		}
		return false
	}
	if names(pkg.Field("bin")) || names(pkg.Field("browser")) {
		return true
	}
	if main, ok := pkg.Field("main").(string); ok && tspath.ResolvePath(directory, main) == fileName {
		return true
	}
	return exportsTarget(pkg.Field("exports"), directory, fileName)
}

// exportsTarget reports whether any target of a package.json `exports` value,
// under any subpath and any condition, names fileName. A `*` in a target
// matches any substring, as in a subpath pattern; `null` targets publish
// nothing.
func exportsTarget(value any, directory, fileName string) bool {
	switch value := value.(type) {
	case string:
		if !strings.HasPrefix(value, "./") {
			return false
		}
		target := tspath.ResolvePath(directory, value)
		prefix, suffix, pattern := strings.Cut(target, "*")
		if !pattern {
			return target == fileName
		}
		return len(fileName) >= len(prefix)+len(suffix) && strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, suffix)
	case []any:
		for _, entry := range value {
			if exportsTarget(entry, directory, fileName) {
				return true
			}
		}
	case map[string]any:
		for _, entry := range value {
			if exportsTarget(entry, directory, fileName) {
				return true
			}
		}
	}
	return false
}
