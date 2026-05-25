package reactutil

import (
	"regexp"
	"strconv"
	"strings"
)

// DefaultReactPragma is the fallback object name for createElement calls
// when `settings.react.pragma` is not configured, matching eslint-plugin-react.
const DefaultReactPragma = "React"

// DefaultReactCreateClass is the fallback ES5 factory name when
// `settings.react.createClass` is not configured, matching
// eslint-plugin-react.
const DefaultReactCreateClass = "createReactClass"

// GetReactPragma reads `settings.react.pragma` from the config settings map.
// Returns DefaultReactPragma when the setting is absent, not a string, or empty.
func GetReactPragma(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactPragma
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactPragma
	}
	pragma, ok := reactSettings["pragma"].(string)
	if !ok || pragma == "" {
		return DefaultReactPragma
	}
	return pragma
}

// DefaultReactFragment is the fallback fragment name for JSX shorthand
// fragment diagnostics when `settings.react.fragment` is not configured,
// matching eslint-plugin-react.
const DefaultReactFragment = "Fragment"

// GetReactFragmentPragma reads `settings.react.fragment` from the config
// settings map. Returns DefaultReactFragment when the setting is absent,
// not a string, or empty.
func GetReactFragmentPragma(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactFragment
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactFragment
	}
	v, ok := reactSettings["fragment"].(string)
	if !ok || v == "" {
		return DefaultReactFragment
	}
	return v
}

// GetReactCreateClass reads `settings.react.createClass` from the config
// settings map. Returns DefaultReactCreateClass when the setting is absent,
// not a string, or empty.
func GetReactCreateClass(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactCreateClass
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactCreateClass
	}
	v, ok := reactSettings["createClass"].(string)
	if !ok || v == "" {
		return DefaultReactCreateClass
	}
	return v
}

// reactVersionRe captures the leading major[.minor[.patch]] numeric triple of
// a semver-ish string. Prerelease / build metadata / range qualifiers are
// ignored — matching eslint-plugin-react's `semver.coerce`-like behavior for
// the simple comparisons this package performs.
var reactVersionRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// ParseReactVersion returns the (major, minor, patch) triple of
// `settings.react.version`. When the setting is missing, not a string, empty,
// or not recognizable as a version, it defaults to (999, 999, 999) — matching
// eslint-plugin-react's `getReactVersionFromContext`, which treats an absent
// version as "latest".
func ParseReactVersion(settings map[string]interface{}) (int, int, int) {
	if settings == nil {
		return 999, 999, 999
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return 999, 999, 999
	}
	raw, _ := reactSettings["version"].(string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 999, 999, 999
	}
	m := reactVersionRe.FindStringSubmatch(raw)
	if m == nil {
		return 999, 999, 999
	}
	toInt := func(s string) int {
		if s == "" {
			return 0
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return n
	}
	return toInt(m[1]), toInt(m[2]), toInt(m[3])
}

// ReactVersionLessThan reports whether `settings.react.version` is strictly
// less than the given major.minor.patch. See ParseReactVersion for the default
// when the setting is missing.
func ReactVersionLessThan(settings map[string]interface{}, major, minor, patch int) bool {
	a, b, c := ParseReactVersion(settings)
	if a != major {
		return a < major
	}
	if b != minor {
		return b < minor
	}
	return c < patch
}
