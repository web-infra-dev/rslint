package moduleresolver

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// Alias maps a name to ordered fallbacks, or ignores it. OnlyModule requires
// an exact name; otherwise package paths can match the name as a prefix.
type Alias struct {
	Name       string   `json:"name"`
	Targets    []string `json:"targets"`
	OnlyModule bool     `json:"onlyModule"`
	Ignore     bool     `json:"ignore"`
}

type nodeResolutionRequest struct{ name, fileName string }

type nodeResolver struct {
	program    *program.Program
	fileName   string
	options    Options
	active     map[nodeResolutionRequest]bool
	mainTarget bool
}

func (resolver *nodeResolver) resolve(name string) (result Result) {
	// A target independently overrides query and fragment. Keep both separate
	// from the filesystem path used by dependency and existence checks.
	if index := strings.IndexAny(name, "?#"); index > 0 && !resolver.mainTarget && !resolver.options.LiteralPaths {
		defer func() {
			if result.Path != "" {
				query, fragment, hasFragment := strings.Cut(result.ResourceSuffix, "#")
				originalQuery, originalFragment, hasOriginalFragment := strings.Cut(name[index:], "#")
				if query == "" {
					query = originalQuery
				}
				if !hasFragment && hasOriginalFragment {
					fragment, hasFragment = originalFragment, true
				}
				result.ResourceSuffix = query
				if hasFragment {
					result.ResourceSuffix += "#" + fragment
				}
			}
		}()
	}
	if len(resolver.options.Aliases) == 0 && len(resolver.options.Fallbacks) == 0 && len(resolver.options.AliasFields) == 0 && len(resolver.options.MainFields) == 0 {
		if modules.IsNodeBuiltin(name) {
			return Result{}
		}
		return resolver.resolveRequest(name)
	}
	key := nodeResolutionRequest{name, resolver.fileName}
	if resolver.active[key] || len(resolver.active) >= 100 {
		return Result{Error: "Recursive alias while resolving '" + name + "'", recursive: true}
	}
	if resolver.active == nil {
		resolver.active = map[nodeResolutionRequest]bool{}
	}
	resolver.active[key] = true
	defer delete(resolver.active, key)
	request := name
	if index := strings.IndexAny(request, "?#"); index > 0 && !resolver.mainTarget && !resolver.options.LiteralPaths {
		request = request[:index]
	}
	if !resolver.mainTarget {
		defer func() {
			if result.Error != "" && !result.recursive && !result.terminal {
				if fallback, matched := resolver.alias(request, resolver.options.Fallbacks); matched {
					result = fallback
				}
			}
		}()
		if result, matched := resolver.alias(request, resolver.options.Aliases); matched {
			return result
		}
	}

	if modules.IsNodeBuiltin(request) && len(resolver.options.AliasFields) == 0 && len(resolver.options.Fallbacks) == 0 {
		return Result{}
	}
	result = resolver.resolveRequest(name)
	return result
}

func (resolver *nodeResolver) alias(request string, aliases []Alias) (Result, bool) {
	for _, alias := range aliases {
		name, candidate := alias.Name, request
		if tspath.IsRootedDiskPath(name) && tspath.IsRootedDiskPath(candidate) {
			name, candidate = tspath.NormalizePath(name), tspath.NormalizePath(candidate)
		}
		match := candidate == name || !alias.OnlyModule && strings.HasPrefix(candidate, name+"/")
		pattern := core.TryParsePattern(name)
		wildcard := !alias.OnlyModule && pattern.IsValid() && pattern.StarIndex >= 0 && pattern.Matches(candidate)
		if !match && !wildcard {
			continue
		}
		if alias.Ignore {
			return Result{}, true
		}
		var last Result
		matched := false
		for _, target := range alias.Targets {
			if match && (candidate == target || strings.HasPrefix(candidate, target+"/")) {
				continue
			}
			if wildcard {
				target = strings.Replace(target, "*", pattern.MatchedText(candidate), 1)
			} else {
				target += strings.TrimPrefix(candidate, name)
			}
			matched = true
			child := *resolver
			child.mainTarget = false
			child.options.FullySpecified = false
			last = child.resolve(target)
			if last.Error == "" || last.recursive || last.terminal {
				return last, true
			}
		}
		if matched {
			if strings.HasPrefix(last.Error, "Can't resolve '") {
				last.Error = "Can't resolve '" + request + "' in '" + tspath.GetDirectoryPath(resolver.fileName) + "'"
			}
			return last, true
		}
	}
	return Result{}, false
}
