package nodeutil

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

type dependenciesKey string

// AllowsDependency includes the package itself, all four dependency fields and
// the nearest ancestor workspace that includes this package (eslint-plugin-n
// v18.3.0). A negative workspace pattern takes precedence over positive ones.
func (pkg *PackageJSON) AllowsDependency(p *program.Program, name string) bool {
	names := program.Cached(p, dependenciesKey(pkg.directory), func() map[string]bool {
		names := map[string]bool{}
		add := func(current *PackageJSON) {
			if name, ok := current.data["name"].(string); ok {
				names[name] = true
			}
			for _, field := range []string{"dependencies", "devDependencies", "peerDependencies", "optionalDependencies"} {
				if dependencies, ok := current.data[field].(map[string]any); ok {
					for name := range dependencies {
						names[name] = true
					}
				}
			}
		}
		add(pkg)
		for directory := tspath.GetDirectoryPath(pkg.directory); directory != ""; {
			ancestor := FindPackage(p, tspath.ResolvePath(directory, "__workspace__.js"))
			if ancestor == nil {
				break
			}
			relative := tspath.GetRelativePathFromDirectory(ancestor.directory, pkg.directory, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
			if matchesWorkspace(relative, ancestor.data["workspaces"]) {
				add(ancestor)
				break
			}
			parent := tspath.GetDirectoryPath(ancestor.directory)
			if parent == ancestor.directory {
				break
			}
			directory = parent
		}
		return names
	})
	return names[name]
}

func matchesWorkspace(relative string, workspaces any) bool {
	if value, ok := workspaces.(map[string]any); ok {
		workspaces = value["packages"]
	}
	matched := false
	for _, raw := range stringArray(workspaces) {
		negative := strings.HasPrefix(raw, "!")
		pattern := strings.TrimSuffix(strings.TrimPrefix(raw, "!"), "/")
		pattern = strings.TrimRight(strings.TrimPrefix(strings.ReplaceAll(pattern, `\`, "/"), "./"), "/")
		// Reuse rslint's extended glob grammar. Numeric brace ranges are a
		// documented extension over the upstream grammar.
		if pattern != "" && minimatch3.Match(pattern, relative, minimatch3.Options{Dot: true, NoNegate: true, NoComment: true, PreserveWhitespace: true}) {
			if negative {
				return false
			}
			matched = true
		}
	}
	return matched
}

// StringListSetting uses options, settings.n, then settings.node. Invalid
// shared values fall through; an explicitly empty array overrides defaults.
func StringListSetting(name string, options, settings map[string]any) []string {
	for _, value := range settingValues(name, options, settings) {
		if list := stringArray(value); list != nil {
			return list
		}
	}
	return nil
}

func settingValues(name string, options, settings map[string]any) []any {
	values := []any{options[name]}
	for _, namespace := range []string{"n", "node"} {
		shared, _ := settings[namespace].(map[string]any)
		values = append(values, shared[name])
	}
	return values
}

func stringArray(value any) []string {
	switch value := value.(type) {
	case []string:
		return slices.Clone(value)
	case []any:
		result := make([]string, len(value))
		for i, item := range value {
			result[i] = settingString(item)
		}
		return result
	}
	return nil
}

// Shared settings and package JSON are JSON values. Preserve JavaScript's
// String conversion, including Array#join's treatment of null elements.
func settingString(value any) string {
	switch value := value.(type) {
	case nil:
		return "null"
	case string:
		return value
	case bool:
		return strconv.FormatBool(value)
	case float64:
		return ecmascript.NumberToString(value)
	case []any:
		parts := make([]string, len(value))
		for i, item := range value {
			if item != nil {
				parts[i] = settingString(item)
			}
		}
		return strings.Join(parts, ",")
	default:
		return "[object Object]"
	}
}
