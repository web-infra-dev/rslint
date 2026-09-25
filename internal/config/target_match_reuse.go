package config

import "reflect"

func (r *FileConfigResolver) reuseTargetMatch(target PathIdentity, match *TargetMatch) (configTargetDecision, bool) {
	if match == nil || match.source == nil || target.Path == "" ||
		target.CanonicalPath == "" || target.CanonicalParentPath == "" || target != match.target {
		return configTargetDecision{}, false
	}
	source := match.source
	current := r.matchSource
	if current == nil || source.pathSpaces != current.pathSpaces ||
		source.configDirectory != current.configDirectory || source.useCaseSensitive != current.useCaseSensitive ||
		source.hasFS != current.hasFS {
		return configTargetDecision{}, false
	}
	if !r.matchSources.getOrInit(source, func() bool {
		return hasSameSelectionPrefix(source.config, r.config, current.configDirectory)
	}) {
		return configTargetDecision{}, false
	}

	decision := match.decision
	if decision.globallyIgnored || len(source.config) == len(r.config) {
		return decision, true
	}
	// CLI rule overrides append unconditional entries after target discovery.
	// They contribute values without widening the original target selection.
	var tail []byte
	if len(r.config) > 64 {
		tail = make([]byte, (len(r.config)-64+7)/8)
		copy(tail, decision.key.tail)
	}
	for index := len(source.config); index < len(r.config); index++ {
		decision.key.add(index, tail)
		decision.matched = true
	}
	if tail != nil {
		decision.key.tail = string(tail)
	}
	return decision, true
}

// hasSameSelectionPrefix compares matching inputs only. Rule validation and
// execution overlays may change values without changing which entries match.
// Like both resolvers, the input config arrays are immutable for their lifetime.
func hasSameSelectionPrefix(before, after RslintConfig, owner string) bool {
	if len(before) > len(after) {
		return false
	}
	for index, left := range before {
		right := after[index]
		if isGlobalIgnoreEntry(left) != isGlobalIgnoreEntry(right) ||
			configEntryPathOrigin(left, owner) != configEntryPathOrigin(right, owner) ||
			!reflect.DeepEqual(left.Files, right.Files) ||
			!reflect.DeepEqual(left.FilePatternGroups, right.FilePatternGroups) ||
			!reflect.DeepEqual(left.Ignores, right.Ignores) ||
			!reflect.DeepEqual(left.collectedGitignore, right.collectedGitignore) {
			return false
		}
	}
	for _, entry := range after[len(before):] {
		if hasFileSelectors(entry) || len(entry.Ignores) != 0 || entry.collectedGitignore != nil ||
			configEntryPathOrigin(entry, owner).basePathScoped {
			return false
		}
	}
	return true
}
