package rule

// LanguageOptions is the normalized, per-file subset of ESLint language
// options consumed by the native rule framework. Its zero value intentionally
// selects the moving `latest` ECMAScript edition.
//
// ECMAVersion is stored as 3, 5, or a four-digit year. Config aliases such as
// 6 (ES2015), 17 (ES2026), and "latest" are normalized before this reaches a
// rule.
type LanguageOptions struct {
	ECMAVersion int `json:"ecmaVersion"`
	// SourceType is the normalized source goal selected for the file. An empty
	// value has ESLint's default module semantics; file-specific defaults are
	// resolved before the options reach a RuleContext.
	SourceType string `json:"sourceType"`
}

// IsValidSourceType reports whether value is one of ESLint's accepted source
// goals. Empty represents an omitted option and is valid for normalized Go
// values, but it is not a valid explicitly authored config value.
func IsValidSourceType(value string) bool {
	return value == "module" || value == "script" || value == "commonjs"
}

// NormalizeECMAScriptVersion normalizes an Espree/ESLint-compatible numeric
// edition. Edition aliases 6 and above map to their four-digit years.
func NormalizeECMAScriptVersion(version int) (int, bool) {
	switch {
	case version == 3 || version == 5:
		return version, true
	case version >= 6 && version <= LatestECMAScriptVersion-2009:
		return version + 2009, true
	case version >= 2015 && version <= LatestECMAScriptVersion:
		return version, true
	default:
		return 0, false
	}
}

// EffectiveECMAVersion returns the normalized edition selected for this file.
func (o LanguageOptions) EffectiveECMAVersion() int {
	if o.ECMAVersion == 0 {
		return LatestECMAScriptVersion
	}
	return o.ECMAVersion
}

// EffectiveSourceType returns the source goal selected for this file. The
// zero value follows ESLint flat config's module default.
// ResolveLanguageDefaults fills .js/.mjs to module and .cjs to commonjs
// before rules receive these options; other extensions may still be empty.
// For TypeScript-flavoured extensions the scope model may still derive
// module/script state from import/export syntax rather than from this value;
// consult ctx.Refs when scope facts matter.
func (o LanguageOptions) EffectiveSourceType() string {
	if o.SourceType == "" {
		return "module"
	}
	return o.SourceType
}
