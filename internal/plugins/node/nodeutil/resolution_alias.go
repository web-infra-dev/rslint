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

type nodeResolutionRequest struct{ name, fileName string }

type nodeResolver struct {
	program    *program.Program
	fileName   string
	options    ResolutionOptions
	active     map[nodeResolutionRequest]bool
	mainTarget bool
}

func (resolver *nodeResolver) resolve(name string) (result nodeResolution) {
	// A target independently overrides query and fragment. Keep both separate
	// from the filesystem path used by dependency and existence checks.
	if index := strings.IndexAny(name, "?#"); index > 0 && !resolver.mainTarget {
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
	if len(resolver.options.Aliases) == 0 && len(resolver.options.AliasFields) == 0 && len(resolver.options.MainFields) == 0 {
		if isNodeBuiltin(name) {
			return nodeResolution{}
		}
		return resolver.resolveRequest(name)
	}
	key := nodeResolutionRequest{name, resolver.fileName}
	if resolver.active[key] || len(resolver.active) >= 100 {
		return nodeResolution{resolveError: "Recursive alias while resolving '" + name + "'", recursive: true}
	}
	if resolver.active == nil {
		resolver.active = map[nodeResolutionRequest]bool{}
	}
	resolver.active[key] = true
	defer delete(resolver.active, key)
	request := name
	if index := strings.IndexAny(request, "?#"); index > 0 && !resolver.mainTarget {
		request = request[:index]
	}
	if !resolver.mainTarget {
		if result, matched := resolver.alias(request); matched {
			return result
		}
	}

	if isNodeBuiltin(request) && len(resolver.options.AliasFields) == 0 {
		return nodeResolution{}
	}
	result = resolver.resolveRequest(name)
	return result
}

func (resolver *nodeResolver) alias(request string) (nodeResolution, bool) {
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
			child := *resolver
			child.mainTarget = false
			last = child.resolve(target)
			if last.resolveError == "" || last.recursive {
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
