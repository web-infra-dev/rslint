package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const modulePredicateKey = "$rslintPredicate"

type configMatcher struct {
	pattern     string
	predicateID string
}

func (matcher configMatcher) isPredicate() bool {
	return matcher.predicateID != ""
}

type configFileSelector struct {
	matchers []configMatcher
}

// DecodeModuleConfig decodes the trusted normalizeConfig payload emitted by
// ConfigModuleHost. Unlike RslintConfig.UnmarshalJSON, this boundary accepts
// opaque predicate descriptors in files/ignores and retains their exact order.
// JSON/JSONC and the low-level API deliberately continue to accept strings
// only, so descriptor objects cannot be supplied by an ordinary config file.
func DecodeModuleConfig(data []byte) (RslintConfig, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil, nil
	}
	var rawEntries []json.RawMessage
	if err := json.Unmarshal(data, &rawEntries); err != nil {
		return nil, err
	}

	entries := make(RslintConfig, 0, len(rawEntries))
	for entryIndex, rawEntry := range rawEntries {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rawEntry, &raw); err != nil {
			return nil, err
		}

		withoutMatchers := make(map[string]json.RawMessage, len(raw))
		for key, value := range raw {
			if key != "files" && key != "ignores" {
				withoutMatchers[key] = value
			}
		}
		plainJSON, err := json.Marshal(withoutMatchers)
		if err != nil {
			return nil, err
		}
		var entry ConfigEntry
		if err := json.Unmarshal(plainJSON, &entry); err != nil {
			return nil, err
		}

		if rawFiles, present := raw["files"]; present {
			selectors, err := decodeModuleFileSelectors(rawFiles, entryIndex)
			if err != nil {
				return nil, err
			}
			entry.moduleFileSelectors = selectors
			for _, selector := range selectors {
				if len(selector.matchers) == 1 && !selector.matchers[0].isPredicate() {
					entry.Files = append(entry.Files, selector.matchers[0].pattern)
					continue
				}
				group := make([]string, 0, len(selector.matchers))
				allStrings := true
				for _, matcher := range selector.matchers {
					if matcher.isPredicate() {
						allStrings = false
						break
					}
					group = append(group, matcher.pattern)
				}
				if allStrings {
					entry.FilePatternGroups = append(entry.FilePatternGroups, group)
				}
			}
		}
		if rawIgnores, present := raw["ignores"]; present {
			matchers, err := decodeModuleMatcherArray(rawIgnores, entryIndex, "ignores", false)
			if err != nil {
				return nil, err
			}
			entry.moduleIgnoreMatchers = matchers
			for _, matcher := range matchers {
				if !matcher.isPredicate() {
					entry.Ignores = append(entry.Ignores, matcher.pattern)
				}
			}
		}

		// Preserve global-ignore object-shape semantics for unsupported or
		// undefined-authored fields in exactly the same way as ordinary decoding.
		hasNonGlobalKey := false
		for key := range raw {
			if key != "ignores" && key != "name" && key != "basePath" {
				hasNonGlobalKey = true
				break
			}
		}
		if hasNonGlobalKey && !hasFileSelectors(entry) && entry.Rules == nil && entry.Plugins == nil && entry.Settings == nil && entry.LanguageOptions == nil {
			entry.Settings = Settings{}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func decodeModuleFileSelectors(raw json.RawMessage, entryIndex int) ([]configFileSelector, error) {
	var selectors []json.RawMessage
	if err := json.Unmarshal(raw, &selectors); err != nil || len(selectors) == 0 {
		if err != nil {
			return nil, fmt.Errorf("config entry at index %d: key \"files\": expected value to be a non-empty array: %w", entryIndex, err)
		}
		return nil, fmt.Errorf("config entry at index %d: key \"files\": expected value to be a non-empty array", entryIndex)
	}

	result := make([]configFileSelector, 0, len(selectors))
	for selectorIndex, selector := range selectors {
		var nested []json.RawMessage
		if err := json.Unmarshal(selector, &nested); err == nil {
			matchers, err := decodeModuleMatcherItems(nested, entryIndex, "files", selectorIndex)
			if err != nil {
				return nil, err
			}
			result = append(result, configFileSelector{matchers: matchers})
			continue
		}
		matcher, err := decodeModuleMatcher(selector)
		if err != nil {
			return nil, fmt.Errorf(
				"config entry at index %d: key \"files\": item at index %d must be a string, a function, or an array of strings and functions",
				entryIndex,
				selectorIndex,
			)
		}
		result = append(result, configFileSelector{matchers: []configMatcher{matcher}})
	}
	return result, nil
}

func decodeModuleMatcherArray(raw json.RawMessage, entryIndex int, key string, requireNonEmpty bool) ([]configMatcher, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || (requireNonEmpty && len(items) == 0) {
		if err != nil {
			return nil, fmt.Errorf("config entry at index %d: key %q: expected an array: %w", entryIndex, key, err)
		}
		return nil, fmt.Errorf("config entry at index %d: key %q: expected a non-empty array", entryIndex, key)
	}
	return decodeModuleMatcherItems(items, entryIndex, key, -1)
}

func decodeModuleMatcherItems(items []json.RawMessage, entryIndex int, key string, selectorIndex int) ([]configMatcher, error) {
	matchers := make([]configMatcher, 0, len(items))
	for itemIndex, item := range items {
		matcher, err := decodeModuleMatcher(item)
		if err != nil {
			location := fmt.Sprintf("item at index %d", itemIndex)
			if selectorIndex >= 0 {
				location = fmt.Sprintf("nested item at files[%d][%d]", selectorIndex, itemIndex)
			}
			return nil, fmt.Errorf("config entry at index %d: key %q: %s must be a string or a function", entryIndex, key, location)
		}
		matchers = append(matchers, matcher)
	}
	return matchers, nil
}

func decodeModuleMatcher(raw json.RawMessage) (configMatcher, error) {
	var pattern string
	if err := json.Unmarshal(raw, &pattern); err == nil {
		return configMatcher{pattern: pattern}, nil
	}
	var descriptor map[string]json.RawMessage
	if err := json.Unmarshal(raw, &descriptor); err != nil || len(descriptor) != 1 {
		return configMatcher{}, errors.New("invalid matcher")
	}
	predicateRaw, present := descriptor[modulePredicateKey]
	if !present {
		return configMatcher{}, errors.New("invalid matcher")
	}
	var predicateID string
	if err := json.Unmarshal(predicateRaw, &predicateID); err != nil || predicateID == "" {
		return configMatcher{}, errors.New("invalid predicate id")
	}
	return configMatcher{predicateID: predicateID}, nil
}

func fileSelectors(entry ConfigEntry) []configFileSelector {
	if entry.moduleFileSelectors != nil {
		return entry.moduleFileSelectors
	}
	selectors := make([]configFileSelector, 0, len(entry.Files)+len(entry.FilePatternGroups))
	for _, pattern := range entry.Files {
		selectors = append(selectors, configFileSelector{matchers: []configMatcher{{pattern: pattern}}})
	}
	for _, group := range entry.FilePatternGroups {
		matchers := make([]configMatcher, 0, len(group))
		for _, pattern := range group {
			matchers = append(matchers, configMatcher{pattern: pattern})
		}
		selectors = append(selectors, configFileSelector{matchers: matchers})
	}
	return selectors
}

func ignoreMatchers(entry ConfigEntry) []configMatcher {
	if entry.moduleIgnoreMatchers != nil {
		return entry.moduleIgnoreMatchers
	}
	matchers := make([]configMatcher, 0, len(entry.Ignores))
	for _, pattern := range entry.Ignores {
		matchers = append(matchers, configMatcher{pattern: pattern})
	}
	return matchers
}

func hasIgnoreMatchers(entry ConfigEntry) bool {
	return len(entry.Ignores) > 0 || len(entry.moduleIgnoreMatchers) > 0
}

// HasConfigPredicates reports whether a trusted JavaScript module config
// contains any live files/ignores predicate descriptors.
func HasConfigPredicates(config RslintConfig) bool {
	for _, entry := range config {
		for _, selector := range fileSelectors(entry) {
			for _, matcher := range selector.matchers {
				if matcher.isPredicate() {
					return true
				}
			}
		}
		for _, matcher := range ignoreMatchers(entry) {
			if matcher.isPredicate() {
				return true
			}
		}
	}
	return false
}
