package rule

// LanguageOptions is the normalized, per-file subset of ESLint language
// options consumed by the native rule framework. Its zero value intentionally
// selects the moving `latest` ECMAScript edition.
//
// ECMAVersion is stored as 3, 5, or a four-digit year. Config aliases such as
// 6 (ES2015), 17 (ES2026), and "latest" are normalized before this reaches a
// rule.
type LanguageOptions struct {
	ECMAVersion int
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
