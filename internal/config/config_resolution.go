package config

import (
	"strconv"

	"github.com/web-infra-dev/rslint/internal/utils/jsonorder"
)

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

// matchConfigEntries applies the flat-config selection policy once and returns
// the exact set of contributing entries. A nil entryIgnorePatterns asks the
// direct compatibility path to parse entry ignores on demand; run-scoped
// resolvers pass their precomputed immutable patterns.
func (config RslintConfig) matchConfigEntries(
	filePath string,
	cwd string,
	globalIgnorePatterns []IgnorePattern,
	entryIgnorePatterns [][]IgnorePattern,
	directoryBlocks *directoryBlockMatcher,
) (configMatchKey, bool) {
	if len(globalIgnorePatterns) > 0 {
		var blocked bool
		if directoryBlocks == nil {
			blocked = isDirBlockedByIgnores(filePath, globalIgnorePatterns, cwd)
		} else {
			blocked = directoryBlocks.blocksFileDirectory(filePath)
		}
		if blocked {
			return configMatchKey{}, false
		}
	}
	matchPath := newFileMatchPath(filePath, cwd)
	if len(globalIgnorePatterns) > 0 && matchPath.isIgnored(globalIgnorePatterns) {
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
		if hasFileSelectors(entry) && !matchPath.matchesConfigEntry(entry) {
			continue
		}

		ignores := entryIgnorePatternsAt(entryIgnorePatterns, index, entry)
		if len(ignores) > 0 && matchPath.isIgnored(ignores) {
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

// mergeConfigEntries merges the matched entries and retains the authored base
// of the effective project value. It does not expand or read project paths.
// key must come from matchConfigEntries for this config.
func (config RslintConfig) mergeConfigEntries(key configMatchKey, configDirectory string) *MergedConfig {
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
				next.optionKeyOrder = previous.optionKeyOrder
			} else if len(next.Options) != 0 && entry.propertyOrder != nil {
				next.optionKeyOrder = make([]*jsonorder.Order, len(next.Options))
				for index := range next.Options {
					next.optionKeyOrder[index] = entry.propertyOrder.At("rules", ruleName, strconv.Itoa(index+1))
				}
			}
			merged.Rules[ruleName] = next
		}

		for _, plugin := range entry.Plugins {
			merged.Plugins[NormalizePluginName(plugin)] = struct{}{}
		}

		if entry.Settings != nil {
			merged.settingsKeyOrder = mergeConfigObjectOrder(map[string]any(merged.Settings), map[string]any(entry.Settings), merged.settingsKeyOrder, entry.propertyOrder.At("settings"))
			merged.Settings = Settings(deepMergeConfigObjects(
				map[string]any(merged.Settings),
				map[string]any(entry.Settings),
			))
		}

		merged.LanguageOptions = mergeLanguageOptions(merged.LanguageOptions, entry.LanguageOptions)
		if entry.LanguageOptions != nil && entry.LanguageOptions.ParserOptions != nil {
			options := entry.LanguageOptions.ParserOptions
			if options.Project != nil {
				merged.project = &ProjectDeclaration{
					Patterns: options.Project, BaseDirectory: configEntryBaseDirectory(entry, configDirectory),
				}
			} else if options.ProjectDisabled || options.projectAutomatic {
				merged.project = nil
			}
		}
	}

	return merged
}

// Overriding an existing property keeps its position; new properties append.
// Object children merge recursively, while arrays and scalars replace them.
func mergeConfigObjectOrder(base, override map[string]any, baseOrder, overrideOrder *jsonorder.Order) *jsonorder.Order {
	if baseOrder == nil && overrideOrder == nil {
		return nil
	}
	merged := &jsonorder.Order{Keys: baseOrder.PropertyKeys(base), Children: map[string]*jsonorder.Order{}}
	for key := range base {
		merged.Children[key] = baseOrder.At(key)
	}
	for _, key := range overrideOrder.PropertyKeys(override) {
		if _, exists := base[key]; !exists {
			merged.Keys = append(merged.Keys, key)
		}
		baseObject, baseIsObject := configObject(base[key])
		overrideObject, overrideIsObject := configObject(override[key])
		if baseIsObject && overrideIsObject {
			merged.Children[key] = mergeConfigObjectOrder(baseObject, overrideObject, baseOrder.At(key), overrideOrder.At(key))
		} else {
			merged.Children[key] = overrideOrder.At(key)
		}
	}
	return merged
}
