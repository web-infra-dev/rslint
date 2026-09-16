package nodeutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
)

type moduleAlias struct {
	Name       string   `json:"name"`
	Targets    []string `json:"targets"`
	OnlyModule bool     `json:"onlyModule"`
	Ignore     bool     `json:"ignore"`
}

type moduleAliasResolver struct {
	program  *program.Program
	fileName string
	options  ResolutionOptions
	active   map[string]bool
}

func (resolver *moduleAliasResolver) resolve(name string) (result nodeResolution) {
	// A target independently overrides query and fragment. Keep both separate
	// from the filesystem path used by dependency and existence checks.
	if index := strings.IndexAny(name, "?#"); index > 0 {
		defer func() {
			if result.path != "" {
				query, fragment, hasFragment := strings.Cut(result.resourceSuffix, "#")
				originalQuery, originalFragment, hasOriginalFragment := strings.Cut(name[index:], "#")
				if query == "" {
					query = originalQuery
				}
				if !hasFragment && hasOriginalFragment {
					fragment, hasFragment = originalFragment, true
				}
				result.resourceSuffix = query
				if hasFragment {
					result.resourceSuffix += "#" + fragment
				}
			}
		}()
	}
	if len(resolver.options.Aliases) == 0 {
		if isNodeBuiltin(name) {
			return nodeResolution{}
		}
		return resolveModuleRequest(resolver.program, name, resolver.fileName, resolver.options, nil)
	}
	if resolver.active[name] || len(resolver.active) >= 100 {
		return nodeResolution{resolveError: "Recursive alias while resolving '" + name + "'"}
	}
	if resolver.active == nil {
		resolver.active = map[string]bool{}
	}
	resolver.active[name] = true
	defer delete(resolver.active, name)
	request := name
	if index := strings.IndexAny(request, "?#"); index > 0 {
		request = request[:index]
	}
	if result, matched := resolver.alias(request); matched {
		return result
	}

	if isNodeBuiltin(request) {
		return nodeResolution{}
	}
	result = resolveModuleRequest(resolver.program, name, resolver.fileName, resolver.options, resolver.alias)
	return result
}

func (resolver *moduleAliasResolver) alias(request string) (nodeResolution, bool) {
	for _, alias := range resolver.options.Aliases {
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
			return nodeResolution{}, true
		}
		var last nodeResolution
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
			last = resolver.resolve(target)
			if last.resolveError == "" {
				return last, true
			}
		}
		if matched {
			if strings.HasPrefix(last.resolveError, "Can't resolve '") {
				last.resolveError = "Can't resolve '" + request + "' in '" + tspath.GetDirectoryPath(resolver.fileName) + "'"
			}
			return last, true
		}
	}
	return nodeResolution{}, false
}
