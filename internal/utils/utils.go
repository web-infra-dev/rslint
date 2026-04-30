package utils

import (
	"iter"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

func TrimNodeTextRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	return scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos()).WithEnd(node.End())
}

// TrimmedNodeText returns the source text for node over the same span as TrimNodeTextRange.
func TrimmedNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	r := TrimNodeTextRange(sourceFile, node)
	return sourceFile.Text()[r.Pos():r.End()]
}

func GetCommentsInRange(sourceFile *ast.SourceFile, inRange core.TextRange) iter.Seq[ast.CommentRange] {
	nodeFactory := ast.NewNodeFactory(ast.NodeFactoryHooks{})

	return func(yield func(ast.CommentRange) bool) {
		for commentRange := range scanner.GetTrailingCommentRanges(nodeFactory, sourceFile.Text(), inRange.Pos()) {
			if commentRange.Pos() >= inRange.End() {
				break
			}
			if !yield(commentRange) {
				return
			}
		}

		for commentRange := range scanner.GetLeadingCommentRanges(nodeFactory, sourceFile.Text(), inRange.Pos()) {
			if commentRange.Pos() >= inRange.End() {
				break
			}
			if !yield(commentRange) {
				return
			}
		}
	}
}

func HasCommentsInRange(sourceFile *ast.SourceFile, inRange core.TextRange) bool {
	for range GetCommentsInRange(sourceFile, inRange) {
		return true
	}
	return false
}

func TypeRecurser(t *checker.Type, predicate func(t *checker.Type) /* should stop */ bool) bool {
	if IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, subtype := range t.Types() {
			if TypeRecurser(subtype, predicate) {
				return true
			}
		}
		return false
	} else {
		return predicate(t)
	}
}

// SUPER DIRTY HACK FOR OPTIONAL FIELDS :(
func Ref[T any](a T) *T {
	return &a
}

func GetNumberIndexType(typeChecker *checker.Checker, t *checker.Type) *checker.Type {
	return checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
}
func GetHeritageClauses(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		return node.AsClassExpression().HeritageClauses
	case ast.KindInterfaceDeclaration:
		return node.AsInterfaceDeclaration().HeritageClauses
	}
	return nil
}

// Source: typescript-go/internal/core/core.go
func Filter[T any](slice []T, f func(T) bool) []T {
	for i, value := range slice {
		if !f(value) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Source: typescript-go/internal/core/core.go
func FilterIndex[T any](slice []T, f func(T, int, []T) bool) []T {
	for i, value := range slice {
		if !f(value, i, slice) {
			result := slices.Clone(slice[:i])
			for i++; i < len(slice); i++ {
				value = slice[i]
				if f(value, i, slice) {
					result = append(result, value)
				}
			}
			return result
		}
	}
	return slice
}

// Source: typescript-go/internal/core/core.go
func Map[T, U any](slice []T, f func(T) U) []U {
	if len(slice) == 0 {
		return nil
	}
	result := make([]U, len(slice))
	for i, value := range slice {
		result[i] = f(value)
	}
	return result
}

// Source: typescript-go/internal/core/core.go
func Some[T any](slice []T, f func(T) bool) bool {
	for _, value := range slice {
		if f(value) {
			return true
		}
	}
	return false
}

// Source: typescript-go/internal/core/core.go
func Every[T any](slice []T, f func(T) bool) bool {
	for _, value := range slice {
		if !f(value) {
			return false
		}
	}
	return true
}

// Source: typescript-go/internal/core/core.go
func Flatten[T any](array [][]T) []T {
	var result []T
	for _, subArray := range array {
		result = append(result, subArray...)
	}
	return result
}

// IsConstructorName reports whether `name` follows the ESLint constructor
// naming convention: the first character that is not `_`, `$`, or an ASCII
// digit is uppercase. Names consisting only of `_`, `$` and ASCII digits
// (e.g. `_`, `$$`, `_8`) are not treated as constructors.
//
// Matches the `isConstructor` helper used by ESLint's `new-cap` and
// `object-shorthand` rules, including Unicode identifier characters
// (e.g. `Π`). ESLint's regex `/[^_$0-9]/u` pairs an ASCII-only digit range
// with a Unicode-aware `toUpperCase()` check — we mirror that: the digit
// prefix is strictly ASCII while the case test is `unicode.IsUpper`.
func IsConstructorName(name string) bool {
	for _, r := range name {
		if r == '_' || r == '$' || (r >= '0' && r <= '9') {
			continue
		}
		// First non-prefix rune: constructor iff uppercase.
		return unicode.IsUpper(r)
	}
	return false
}

// IsStringLiteralOrTemplate reports whether node is a string literal or a
// template literal (with or without substitutions). Matches the semantics of
// ESLint's `astUtils.isStringLiteral`, which treats `Literal{string}` and
// `TemplateLiteral` as equivalent. The shim's `ast.IsStringLiteralLike` only
// covers `StringLiteral` and `NoSubstitutionTemplateLiteral`, so we also
// include `TemplateExpression` (templates with `${}`).
func IsStringLiteralOrTemplate(node *ast.Node) bool {
	return node != nil && (ast.IsStringLiteralLike(node) || node.Kind == ast.KindTemplateExpression)
}

// IsPlusBinaryExpression reports whether node is a `+` binary expression.
// Covers both string concatenation and numeric addition — callers that only
// care about concatenation must additionally inspect the operands.
func IsPlusBinaryExpression(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := node.AsBinaryExpression()
	return bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindPlusToken
}

func IncludesModifier(node interface{ Modifiers() *ast.ModifierList }, modifier ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	return Some(modifiers.Nodes, func(m *ast.Node) bool {
		return m.Kind == modifier
	})
}

// Source: https://github.com/microsoft/typescript-go/blob/5652e65d5ae944375676d3955f9755e554576d41/internal/jsnum/string.go#L99
func IsStrWhiteSpace(r rune) bool {
	// This is different than stringutil.IsWhiteSpaceLike.

	// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-LineTerminator
	// https://tc39.es/ecma262/2024/multipage/ecmascript-language-lexical-grammar.html#prod-WhiteSpace

	switch r {
	// LineTerminator
	case '\n', '\r', 0x2028, 0x2029:
		return true
	// WhiteSpace
	case '\t', '\v', '\f', 0xFEFF:
		return true
	}

	// WhiteSpace
	return unicode.Is(unicode.Zs, r)
}

// IsECMABlankLine reports whether s contains only ECMAScript WhiteSpace /
// LineTerminator runes — matching JavaScript's `"".trim() === ""` check used
// by rules like max-lines / max-lines-per-function for `skipBlankLines`.
// Go's strings.TrimSpace diverges on U+FEFF (BOM) and U+0085 (NEL), so we
// can't use it directly.
func IsECMABlankLine(s string) bool {
	for _, r := range s {
		if !IsStrWhiteSpace(r) {
			return false
		}
	}
	return true
}

// LineContentEnd returns the byte position just past the last character of the
// line whose successor starts at nextLineStart — i.e. nextLineStart with its
// immediately-preceding ECMA line terminator (LF, CR, CRLF, LS, PS) stripped.
// Useful when slicing a single line out of source text without its terminator,
// matching the behavior of ESLint's SourceCode.lines entries.
func LineContentEnd(text string, nextLineStart int) int {
	if nextLineStart >= 2 && text[nextLineStart-2] == '\r' && text[nextLineStart-1] == '\n' {
		return nextLineStart - 2
	}
	if nextLineStart >= 1 {
		c := text[nextLineStart-1]
		if c == '\r' || c == '\n' {
			return nextLineStart - 1
		}
		// U+2028 / U+2029 encode as 0xE2 0x80 0xA8 / 0xA9.
		if nextLineStart >= 3 &&
			text[nextLineStart-3] == 0xE2 &&
			text[nextLineStart-2] == 0x80 &&
			(text[nextLineStart-1] == 0xA8 || text[nextLineStart-1] == 0xA9) {
			return nextLineStart - 3
		}
	}
	return nextLineStart
}

// ExcludePaths contains path substrings that should be excluded from linting.
// Used by RunLinterInProgram to skip files during program source file iteration.
var ExcludePaths = []string{"/node_modules/", "bundled:"}

// DefaultExcludeDirNames contains directory names that are always excluded
// from file scanning. This is the single source of truth for default directory
// exclusions, used by DiscoverGapFiles and the no-tsconfig fallback.
// Aligned with JS-side SCAN_EXCLUDE_DIRS: new Set(['node_modules', '.git']).
var DefaultExcludeDirNames = []string{"node_modules", ".git"}

// DefaultIgnoreDirGlobs returns glob patterns derived from DefaultExcludeDirNames,
// suitable for use with ignore pattern matching (e.g., DiscoverGapFiles).
func DefaultIgnoreDirGlobs() []string {
	globs := make([]string, len(DefaultExcludeDirNames))
	for i, name := range DefaultExcludeDirNames {
		globs[i] = name + "/**"
	}
	return globs
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// GetOptionsMap extracts a map[string]interface{} from rule options.
// It handles both array format [{ option: value }] and direct object format { option: value }.
// ExtractRegexPatternAndFlags splits a RegularExpressionLiteral's text (e.g. "/pattern/gi")
// into the pattern and flags portions. Returns ("", "") for malformed input.
func ExtractRegexPatternAndFlags(text string) (pattern string, flags string) {
	if len(text) < 2 || text[0] != '/' {
		return "", ""
	}
	lastSlash := strings.LastIndex(text[1:], "/")
	if lastSlash == -1 {
		return text[1:], ""
	}
	return text[1 : lastSlash+1], text[lastSlash+2:]
}

func GetOptionsMap(opts any) map[string]interface{} {
	if opts == nil {
		return nil
	}

	var optsMap map[string]interface{}
	if arr, ok := opts.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = opts.(map[string]interface{})
	}

	return optsMap
}

// GetOptionsString extracts a string option from the weakly-typed options parameter.
// It handles both direct string format ("value") and array format (["value"]).
func GetOptionsString(opts any) string {
	if opts == nil {
		return ""
	}
	if s, ok := opts.(string); ok {
		return s
	}
	if arr, ok := opts.([]interface{}); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			return s
		}
	}
	return ""
}

// ToStringSlice converts a weakly-typed JSON array ([]interface{}) to []string,
// extracting only the string elements. Returns nil if the input is nil, not an array,
// or contains no strings. Useful for parsing rule options from JSON config.
func ToStringSlice(val interface{}) []string {
	if val == nil {
		return nil
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// NeedsLeadingSpaceForReplacement reports whether inserting `replacement`
// at `insertPos` in `src` would merge with the preceding character into a
// single identifier token. Callers use this when synthesizing a fix whose
// text starts with an identifier (e.g. `Boolean(foo)`, `Number(foo)`,
// `String(foo)`) to decide whether a leading space is required.
//
// Mirrors the identifier/keyword case of ESLint's `canTokensBeAdjacent`:
// `typeof+foo` replaced with `Number(foo)` would otherwise become
// `typeofNumber(foo)` (a single identifier). Multi-byte identifier chars
// are handled via `scanner.IsIdentifierPart` / `scanner.IsIdentifierStart`.
func NeedsLeadingSpaceForReplacement(src string, insertPos int, replacement string) bool {
	if insertPos <= 0 || insertPos > len(src) || replacement == "" {
		return false
	}
	firstRune, _ := utf8.DecodeRuneInString(replacement)
	if firstRune == utf8.RuneError || !scanner.IsIdentifierStart(firstRune) {
		return false
	}
	prevRune, _ := utf8.DecodeLastRuneInString(src[:insertPos])
	if prevRune == utf8.RuneError {
		return false
	}
	return scanner.IsIdentifierPart(prevRune)
}

// NaturalCompare compares two strings using natural sort order,
// where embedded numeric segments are compared by their numeric value
// (e.g., "a2" < "a10" instead of "a10" < "a2").
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func NaturalCompare(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	ai, bi := 0, 0
	for ai < len(ra) && bi < len(rb) {
		ca, cb := ra[ai], rb[bi]

		if unicode.IsDigit(ca) && unicode.IsDigit(cb) {
			na, nextA := extractRuneDigits(ra, ai)
			nb, nextB := extractRuneDigits(rb, bi)
			naTrimmed := strings.TrimLeft(na, "0")
			nbTrimmed := strings.TrimLeft(nb, "0")
			if naTrimmed == "" {
				naTrimmed = "0"
			}
			if nbTrimmed == "" {
				nbTrimmed = "0"
			}
			if len(naTrimmed) != len(nbTrimmed) {
				if len(naTrimmed) < len(nbTrimmed) {
					return -1
				}
				return 1
			}
			if naTrimmed < nbTrimmed {
				return -1
			}
			if naTrimmed > nbTrimmed {
				return 1
			}
			ai = nextA
			bi = nextB
		} else {
			if ca < cb {
				return -1
			}
			if ca > cb {
				return 1
			}
			ai++
			bi++
		}
	}

	if ai < len(ra) {
		return 1
	}
	if bi < len(rb) {
		return -1
	}
	return 0
}

func extractRuneDigits(runes []rune, start int) (string, int) {
	end := start
	for end < len(runes) && unicode.IsDigit(runes[end]) {
		end++
	}
	return string(runes[start:end]), end
}
