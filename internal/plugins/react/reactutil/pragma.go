package reactutil

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
	raw = ecmascript.StringTrim(raw)
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

// GetReactPragmaFromContext mirrors eslint-plugin-react's
// `pragmaUtil.getFromContext`, which resolves the React pragma from the file
// itself before consulting configuration: the first comment carrying a
// `@jsx <name>` annotation wins over `settings.react.pragma`, and only the
// head of a dotted annotation counts (`@jsx Preact.h` yields `Preact`).
// A name that is not a valid JavaScript identifier falls back to
// DefaultReactPragma, matching upstream's warn-and-default path.
//
// Rules that classify JSX by pragma (`<React.Fragment>`, `React.createElement`,
// deprecated `React.*` members) must go through this rather than
// GetReactPragma, or a classic-runtime file compiled with a non-React pragma
// both misses its own fragments and reports React's as if they were the
// configured ones. GetReactPragma remains correct for the settings-only
// lookups upstream performs the same way, such as `getFragmentFromContext`.
func GetReactPragmaFromContext(ctx rule.RuleContext) string {
	if annotated, found := jsxAnnotationPragma(ctx); found {
		if isJavaScriptIdentifier(annotated) {
			return annotated
		}
		return DefaultReactPragma
	}
	pragma := GetReactPragma(ctx.Settings)
	if !isJavaScriptIdentifier(pragma) {
		return DefaultReactPragma
	}
	return pragma
}

// jsxAnnotationPragma returns the head of the first `@jsx <name>` annotation
// found in a comment, and whether any annotation was present at all. The
// caller needs both: upstream stops looking at `settings.react.pragma` as soon
// as an annotation exists, even one it later rejects as a non-identifier.
func jsxAnnotationPragma(ctx rule.RuleContext) (string, bool) {
	if ctx.SourceFile == nil {
		return "", false
	}
	text := ctx.SourceFile.Text()
	// Most files carry no classic-runtime pragma; keep the comment scan off
	// the hot path for them.
	if !strings.Contains(text, "@jsx") {
		return "", false
	}

	// ESLint's SourceCode#getAllComments exposes the hashbang as its first
	// comment. tsgo keeps it separate from ordinary comment ranges, so inspect
	// it first to preserve both inclusion and source-order precedence.
	if shebang := scanner.GetShebang(text); shebang != "" {
		if name, found := jsxPragmaFromCommentValue(shebang[2:]); found {
			return name, true
		}
	}

	for _, comment := range ctx.Comments.All() {
		if name, found := jsxPragmaFromCommentValue(utils.CommentValue(text, comment)); found {
			return name, true
		}
	}
	return "", false
}

func jsxPragmaFromCommentValue(value string) (string, bool) {
	name, found := jsxAnnotationInComment(value)
	if !found {
		return "", false
	}
	// Upstream reads `matches[1].split('.')[0]`.
	if dot := strings.IndexByte(name, '.'); dot >= 0 {
		name = name[:dot]
	}
	return name, true
}

// jsxAnnotationInComment ports `/@jsx\s+([^\s]+)/` over one comment body.
// It is hand-rolled rather than compiled because the character class is
// JavaScript's `\s` — which covers more than Go's RE2 `\s` — and the pattern
// is applied to the source under lint.
func jsxAnnotationInComment(value string) (string, bool) {
	const marker = "@jsx"
	for offset := 0; ; {
		index := strings.Index(value[offset:], marker)
		if index < 0 {
			return "", false
		}
		afterMarker := offset + index + len(marker)
		nameStart := ecmascript.SkipLeadingWhitespace(value, afterMarker, len(value))
		// `\s+` demands at least one whitespace character, so `@jsxFrag Foo`
		// is not a `@jsx` annotation.
		if nameStart > afterMarker {
			// `[^\s]+` demands at least one character.
			if nameEnd := skipNonWhitespace(value, nameStart); nameEnd > nameStart {
				return value[nameStart:nameEnd], true
			}
		}
		offset = afterMarker
	}
}

// skipNonWhitespace returns the index one past the run of non-whitespace
// characters starting at `start` — the reverse of
// ecmascript.SkipLeadingWhitespace, over the same JavaScript `\s` definition.
func skipNonWhitespace(text string, start int) int {
	position := start
	for position < len(text) {
		if text[position] < 0x80 {
			if ecmascript.IsTriviaWhitespaceByte(text[position]) {
				return position
			}
			position++
			continue
		}
		r, size := utf8.DecodeRuneInString(text[position:])
		if size == 0 || ecmascript.IsTriviaWhitespaceRune(r) {
			return position
		}
		position += size
	}
	return position
}

// isJavaScriptIdentifier ports upstream's
// `/^[_$a-zA-Z][_$a-zA-Z0-9]*$/` guard. Like upstream it is deliberately
// ASCII-only and does not reject reserved words.
func isJavaScriptIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i := range len(name) {
		c := name[i]
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '$'
		if isLetter {
			continue
		}
		if i > 0 && c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}
