package lang

import (
	"strings"

	"golang.org/x/text/language"
)

// isValidBCP47Tag mirrors `language-tags`' `tags.check(value)` (used by the
// upstream `eslint-plugin-jsx-a11y/lang` rule) byte-for-byte on the cases the
// rule actually exercises. The check is intentionally NOT a thin wrapper
// around `language.Parse`: the Go and JS libraries diverge in several
// directions that matter for upstream parity. Each rule below has been
// verified empirically against `language-tags@1.x` via side-by-side
// comparison.
//
// Steps:
//
//  1. Reject underscore separator — BCP-47 ABNF only allows hyphen.
//     `language.Parse` silently rewrites `_` to `-`; `language-tags` rejects.
//
//  2. Trim leading/trailing whitespace — `language-tags` accepts ` en `,
//     `language.Parse` rejects with "not well-formed".
//
//  3. Reject internal whitespace — both libraries reject `en US`.
//
//  4. Reject deprecated grandfathered tags (`i-klingon`, `en-gb-oed`, etc.)
//     — `language-tags` returns false because `.deprecated()` is set;
//     `language.Parse` accepts and normalizes to the preferred-value form
//     (`tlh`, `en-GB-oxendict`). We carry the IANA grandfathered table to
//     match upstream's rejection.
//
//  5. Reject private-use-only tags (`x-foo`, `x`) — `language-tags` errors
//     "Missing language tag"; `language.Parse` accepts.
//
//  6. Reject leading hyphen — both reject.
//
//  7. Strip empty subtags from trailing/double hyphens (`en-` → `en`,
//     `en--US` → `en-US`) — `language-tags`' subtag parser skips empty
//     codes, `language.Parse` rejects them.
//
//  8. Run `language.Parse` for IANA-registered subtag membership and
//     remaining well-formedness (length limits, alpha/digit class, valid
//     extension subtags, etc.).
//
//  9. Reject Suppress-Script violations (`hi-Deva`, `ja-Jpan`, `en-Latn`,
//     …) — per RFC 5646 §4.1, a script subtag should be omitted when it
//     matches the language's IANA Suppress-Script. `language-tags`
//     enforces this; `language.Parse` accepts.
//
// The IANA tables (deprecatedGrandfatheredTags, suppressScriptByLanguage)
// were extracted from `language-subtag-registry@0.3.x` and are kept in
// sync with the upstream registry — manual refresh is required if rslint
// adopts a newer registry snapshot.
func isValidBCP47Tag(value string) bool {
	if strings.ContainsRune(value, '_') {
		return false
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f' {
			return false
		}
	}
	if deprecatedGrandfatheredTags[strings.ToLower(trimmed)] {
		return false
	}
	if isPrivateUseOnly(trimmed) {
		return false
	}
	if strings.HasPrefix(trimmed, "-") {
		return false
	}
	cleaned := stripEmptySubtags(trimmed)
	if hasLegacyPOSIXVariant(cleaned) {
		return false
	}
	tag, err := language.Parse(cleaned)
	if err != nil {
		return false
	}
	if hasSuppressScriptViolation(tag) {
		return false
	}
	return true
}

// hasLegacyPOSIXVariant reports whether any subtag of the tag is `posix`
// (case-insensitive). `language.Parse` silently rewrites `<lang>-<region>-POSIX`
// into the Unicode extension form `<lang>-<region>-u-va-posix`, but
// `language-tags` reports "Unknown code 'posix'" because POSIX is not a
// registered variant subtag in the IANA Subtag Registry. Rejecting the
// `posix` subtag up front keeps the acceptance set aligned with upstream.
func hasLegacyPOSIXVariant(s string) bool {
	for _, p := range strings.Split(s, "-") {
		if strings.EqualFold(p, "posix") {
			return true
		}
	}
	return false
}

// isPrivateUseOnly reports whether the tag has no primary language subtag,
// i.e. it starts with the private-use singleton `x` (`x`, `x-...`). Per
// `language-tags` the tag is invalid because there is no language portion.
func isPrivateUseOnly(s string) bool {
	if s == "" {
		return false
	}
	if s[0] != 'x' && s[0] != 'X' {
		return false
	}
	if len(s) == 1 {
		return true
	}
	return s[1] == '-'
}

// stripEmptySubtags removes empty entries produced by trailing or repeated
// hyphens. `language-tags`' subtag iterator silently skips empty codes, so
// `en-` and `en--US` parse as if they were `en` and `en-US`. `language.Parse`
// would otherwise reject these as "not well-formed".
func stripEmptySubtags(s string) string {
	parts := strings.Split(s, "-")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "-")
}

// hasSuppressScriptViolation reports whether the parsed tag explicitly
// carries a script subtag that matches the IANA Suppress-Script for its
// primary language. Per RFC 5646 §4.1 such a tag is considered invalid by
// `language-tags`, even though it is syntactically well-formed.
//
// `language.Parse` records the script's confidence: `Exact` means the
// caller wrote the script subtag explicitly; `High`/`Low` mean it was
// inferred from the language alone. We check only the `Exact` case so
// that bare `fr` (whose script Latn is inferred) is not over-rejected.
func hasSuppressScriptViolation(tag language.Tag) bool {
	script, conf := tag.Script()
	if conf != language.Exact {
		return false
	}
	base, _ := tag.Base()
	suppress, ok := suppressScriptByLanguage[strings.ToLower(base.String())]
	if !ok {
		return false
	}
	return strings.EqualFold(script.String(), suppress)
}

// deprecatedGrandfatheredTags is the lowercased set of IANA Subtag Registry
// entries with Type=grandfathered AND Deprecated≠"". `language-tags`'
// `tags.check()` returns false for these because the underlying `Tag`
// object's `.deprecated()` evaluates truthy. Extracted from
// `language-subtag-registry@0.3.x`. The list is stable (no new
// grandfathered tags are added after RFC 5646); refresh only on rare
// IANA Deprecated-flag changes.
var deprecatedGrandfatheredTags = map[string]bool{
	"art-lojban":  true,
	"cel-gaulish": true,
	"en-gb-oed":   true,
	"i-ami":       true,
	"i-bnn":       true,
	"i-enochian":  true,
	"i-hak":       true,
	"i-klingon":   true,
	"i-lux":       true,
	"i-navajo":    true,
	"i-pwn":       true,
	"i-tao":       true,
	"i-tay":       true,
	"i-tsu":       true,
	"no-bok":      true,
	"no-nyn":      true,
	"sgn-be-fr":   true,
	"sgn-be-nl":   true,
	"sgn-ch-de":   true,
	"zh-guoyu":    true,
	"zh-hakka":    true,
	"zh-min":      true,
	"zh-min-nan":  true,
	"zh-xiang":    true,
}

// suppressScriptByLanguage maps a lowercased primary language subtag to the
// script subtag that the IANA Subtag Registry records as its
// `Suppress-Script`. When the input tag's explicit script subtag matches
// the suppress value, `language-tags` reports "the script subtag 'X' is
// the same as the language suppress-script" and the tag becomes invalid.
//
// Extracted from `language-subtag-registry@0.3.x` (Type=language entries
// carrying a Suppress-Script field). 134 entries as of the registry
// snapshot used during the port.
var suppressScriptByLanguage = map[string]string{
	"ab": "Cyrl", "af": "Latn", "am": "Ethi", "ar": "Arab", "as": "Beng",
	"ay": "Latn", "be": "Cyrl", "bg": "Cyrl", "bn": "Beng", "bs": "Latn",
	"ca": "Latn", "ch": "Latn", "cs": "Latn", "cy": "Latn", "da": "Latn",
	"de": "Latn", "dsb": "Latn", "dv": "Thaa", "dz": "Tibt", "el": "Grek",
	"en": "Latn", "eo": "Latn", "es": "Latn", "et": "Latn", "eu": "Latn",
	"fa": "Arab", "fi": "Latn", "fj": "Latn", "fo": "Latn", "fr": "Latn",
	"frr": "Latn", "frs": "Latn", "fy": "Latn", "ga": "Latn", "gl": "Latn",
	"gn": "Latn", "gsw": "Latn", "gu": "Gujr", "gv": "Latn", "he": "Hebr",
	"hi": "Deva", "hr": "Latn", "hsb": "Latn", "ht": "Latn", "hu": "Latn",
	"hy": "Armn", "id": "Latn", "in": "Latn", "is": "Latn", "it": "Latn",
	"iw": "Hebr", "ja": "Jpan", "ka": "Geor", "kk": "Cyrl", "kl": "Latn",
	"km": "Khmr", "kn": "Knda", "ko": "Kore", "kok": "Deva", "la": "Latn",
	"lb": "Latn", "ln": "Latn", "lo": "Laoo", "lt": "Latn", "lv": "Latn",
	"mai": "Deva", "men": "Latn", "mg": "Latn", "mh": "Latn", "mk": "Cyrl",
	"ml": "Mlym", "mo": "Latn", "mr": "Deva", "ms": "Latn", "mt": "Latn",
	"my": "Mymr", "na": "Latn", "nb": "Latn", "nd": "Latn", "nds": "Latn",
	"ne": "Deva", "niu": "Latn", "nl": "Latn", "nn": "Latn", "no": "Latn",
	"nqo": "Nkoo", "nr": "Latn", "nso": "Latn", "ny": "Latn", "om": "Latn",
	"or": "Orya", "pa": "Guru", "pl": "Latn", "ps": "Arab", "pt": "Latn",
	"qu": "Latn", "rm": "Latn", "rn": "Latn", "ro": "Latn", "ru": "Cyrl",
	"rw": "Latn", "sg": "Latn", "si": "Sinh", "sk": "Latn", "sl": "Latn",
	"sm": "Latn", "so": "Latn", "sq": "Latn", "ss": "Latn", "st": "Latn",
	"sv": "Latn", "sw": "Latn", "ta": "Taml", "te": "Telu", "tem": "Latn",
	"th": "Thai", "ti": "Ethi", "tkl": "Latn", "tl": "Latn", "tmh": "Latn",
	"tn": "Latn", "to": "Latn", "tpi": "Latn", "tr": "Latn", "ts": "Latn",
	"tvl": "Latn", "uk": "Cyrl", "ur": "Arab", "ve": "Latn", "vi": "Latn",
	"xh": "Latn", "yi": "Hebr", "zbl": "Blis", "zu": "Latn",
}
