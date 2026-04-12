package utils

import (
	"iter"
	"slices"
	"strings"
	"unicode"

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
