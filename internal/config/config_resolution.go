package config

// configMatchKey is the exact resolver-local identity of the ordered config
// entries that apply to one file. The first 64 entries avoid allocation; the
// complete tail bitset keeps configs with more entries collision-free.
type configMatchKey struct {
	first uint64
	tail  string
}

func (key *configMatchKey) add(index int, tail []byte) {
	if index < 64 {
		key.first |= uint64(1) << index
		return
	}
	tailIndex := index - 64
	tail[tailIndex/8] |= byte(1 << (tailIndex % 8))
}

func (key configMatchKey) contains(index int) bool {
	if index < 64 {
		return key.first&(uint64(1)<<index) != 0
	}
	tailIndex := index - 64
	return key.tail[tailIndex/8]&byte(1<<(tailIndex%8)) != 0
}

func parseEntryIgnorePatterns(config RslintConfig) [][]IgnorePattern {
	patterns := make([][]IgnorePattern, len(config))
	for index, entry := range config {
		if !isGlobalIgnoreEntry(entry) {
			patterns[index] = ParseIgnorePatterns(entry.Ignores)
		}
	}
	return patterns
}

// matchConfigEntries applies the flat-config selection policy once and returns
// the exact set of contributing entries. A nil entryIgnorePatterns asks the
// direct compatibility path to parse entry ignores on demand; run-scoped
// resolvers pass their precomputed immutable patterns.
func (config RslintConfig) matchConfigEntries(
	filePath string,
	cwd string,
	globalIgnorePatterns []IgnorePattern,
	entryIgnorePatterns [][]IgnorePattern,
) (configMatchKey, bool) {
	if len(globalIgnorePatterns) > 0 &&
		(isDirBlockedByIgnores(filePath, globalIgnorePatterns, cwd) ||
			isFileIgnored(filePath, globalIgnorePatterns, cwd)) {
		return configMatchKey{}, false
	}

	selected := isDefaultLintFile(filePath)
	matched := false
	var key configMatchKey
	var tail []byte
	if tailEntryCount := len(config) - 64; tailEntryCount > 0 {
		tail = make([]byte, (tailEntryCount+7)/8)
	}

	for index, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}
		if hasFileSelectors(entry) && !isFileMatchedByConfigEntry(filePath, entry, cwd) {
			continue
		}

		ignores := entryIgnorePatternsAt(entryIgnorePatterns, index, entry)
		if isFileIgnored(filePath, ignores, cwd) {
			continue
		}

		matched = true
		if hasFileSelectors(entry) {
			selected = true
		}
		key.add(index, tail)
	}

	if !selected || !matched {
		return configMatchKey{}, false
	}
	if tail != nil {
		key.tail = string(tail)
	}
	return key, true
}

func entryIgnorePatternsAt(patterns [][]IgnorePattern, index int, entry ConfigEntry) []IgnorePattern {
	if patterns != nil {
		return patterns[index]
	}
	return ParseIgnorePatterns(entry.Ignores)
}

// mergeConfigEntries performs only the path-independent half of flat-config
// resolution. key must come from matchConfigEntries for this config.
func (config RslintConfig) mergeConfigEntries(key configMatchKey) *MergedConfig {
	merged := &MergedConfig{
		Rules:   make(map[string]*RuleConfig),
		Plugins: make(map[string]struct{}),
	}

	for index, entry := range config {
		if !key.contains(index) {
			continue
		}

		// Later entries override earlier rules. A severity-only override keeps
		// the previous options, matching ESLint flat-config behavior.
		for ruleName, ruleValue := range entry.Rules {
			next, hasOptions, err := parseRuleConfigValue(ruleValue)
			if err != nil {
				// Config ingress validates rule values before merge. Keep this guard
				// for callers that construct RslintConfig directly.
				continue
			}
			if previous := merged.Rules[ruleName]; !hasOptions && previous != nil {
				next.Options = append([]interface{}(nil), previous.Options...)
			}
			merged.Rules[ruleName] = next
		}

		for _, plugin := range entry.Plugins {
			merged.Plugins[NormalizePluginName(plugin)] = struct{}{}
		}

		if entry.Settings != nil {
			merged.Settings = Settings(deepMergeConfigObjects(
				map[string]any(merged.Settings),
				map[string]any(entry.Settings),
			))
		}

		merged.LanguageOptions = mergeLanguageOptions(merged.LanguageOptions, entry.LanguageOptions)
	}

	return merged
}
