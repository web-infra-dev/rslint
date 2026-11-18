package naming_convention

import (
	_ "embed"
	"fmt"
	"math/bits"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed naming_convention.schema.json
var schemaJSON []byte

// NamingConventionRule is the exported rule for registration.
var NamingConventionRule = rule.CreateRule(rule.Rule{
	Name:   "naming-convention",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

// ---- Enums ----

type predefinedFormat int

const (
	formatCamelCase predefinedFormat = iota
	formatStrictCamelCase
	formatPascalCase
	formatStrictPascalCase
	formatSnakeCase
	formatUpperCase
)

func parseFormatName(s string) (predefinedFormat, bool) {
	switch s {
	case "camelCase":
		return formatCamelCase, true
	case "strictCamelCase":
		return formatStrictCamelCase, true
	case "PascalCase":
		return formatPascalCase, true
	case "StrictPascalCase":
		return formatStrictPascalCase, true
	case "snake_case":
		return formatSnakeCase, true
	case "UPPER_CASE":
		return formatUpperCase, true
	default:
		return 0, false
	}
}

func formatName(format predefinedFormat) string {
	switch format {
	case formatCamelCase:
		return "camelCase"
	case formatStrictCamelCase:
		return "strictCamelCase"
	case formatPascalCase:
		return "PascalCase"
	case formatStrictPascalCase:
		return "StrictPascalCase"
	case formatSnakeCase:
		return "snake_case"
	case formatUpperCase:
		return "UPPER_CASE"
	default:
		return ""
	}
}

type underscoreOption int

const (
	underscoreNone underscoreOption = iota // not specified — no check, no stripping
	underscoreForbid
	underscoreAllow
	underscoreRequire
	underscoreRequireDouble
	underscoreAllowDouble
	underscoreAllowSingleOrDouble
)

func parseUnderscoreOption(s string) underscoreOption {
	switch s {
	case "forbid":
		return underscoreForbid
	case "allow":
		return underscoreAllow
	case "require":
		return underscoreRequire
	case "requireDouble":
		return underscoreRequireDouble
	case "allowDouble":
		return underscoreAllowDouble
	case "allowSingleOrDouble":
		return underscoreAllowSingleOrDouble
	default:
		return underscoreAllow
	}
}

type selectorKind int

const (
	// Individual selectors
	selectorVariable selectorKind = 1 << iota
	selectorFunction
	selectorParameter
	selectorParameterProperty
	selectorClassicAccessor
	selectorEnumMember
	selectorClassMethod
	selectorObjectLiteralMethod
	selectorTypeMethod
	selectorClassProperty
	selectorObjectLiteralProperty
	selectorTypeProperty
	selectorAutoAccessor
	selectorClass
	selectorInterface
	selectorTypeAlias
	selectorEnum
	selectorTypeParameter
	selectorImport
)

// Group selectors
const (
	selectorVariableLike = selectorVariable | selectorFunction | selectorParameter
	selectorMemberLike   = selectorClassProperty | selectorObjectLiteralProperty | selectorTypeProperty |
		selectorParameterProperty | selectorEnumMember | selectorClassMethod | selectorObjectLiteralMethod |
		selectorTypeMethod | selectorClassicAccessor | selectorAutoAccessor
	selectorTypeLike                     = selectorClass | selectorInterface | selectorTypeAlias | selectorEnum | selectorTypeParameter
	selectorMethod                       = selectorClassMethod | selectorObjectLiteralMethod | selectorTypeMethod
	selectorProperty                     = selectorClassProperty | selectorObjectLiteralProperty | selectorTypeProperty
	selectorAccessor                     = selectorClassicAccessor | selectorAutoAccessor
	selectorDefault         selectorKind = -1
	selectorUnusedSupported              = selectorVariableLike | selectorTypeLike
)

func parseSelectorKind(s string) (selectorKind, bool) {
	switch s {
	case "variable":
		return selectorVariable, true
	case "function":
		return selectorFunction, true
	case "parameter":
		return selectorParameter, true
	case "classProperty":
		return selectorClassProperty, true
	case "objectLiteralProperty":
		return selectorObjectLiteralProperty, true
	case "typeProperty":
		return selectorTypeProperty, true
	case "parameterProperty":
		return selectorParameterProperty, true
	case "enumMember":
		return selectorEnumMember, true
	case "classMethod":
		return selectorClassMethod, true
	case "objectLiteralMethod":
		return selectorObjectLiteralMethod, true
	case "typeMethod":
		return selectorTypeMethod, true
	case "classicAccessor":
		return selectorClassicAccessor, true
	case "autoAccessor":
		return selectorAutoAccessor, true
	case "class":
		return selectorClass, true
	case "interface":
		return selectorInterface, true
	case "typeAlias":
		return selectorTypeAlias, true
	case "enum":
		return selectorEnum, true
	case "typeParameter":
		return selectorTypeParameter, true
	case "import":
		return selectorImport, true
	// Group selectors
	case "default":
		return selectorDefault, true
	case "variableLike":
		return selectorVariableLike, true
	case "memberLike":
		return selectorMemberLike, true
	case "typeLike":
		return selectorTypeLike, true
	case "method":
		return selectorMethod, true
	case "property":
		return selectorProperty, true
	case "accessor":
		return selectorAccessor, true
	default:
		return 0, false
	}
}

type modifierKind int

const (
	modifierConst modifierKind = 1 << iota
	modifierReadonly
	modifierStatic
	modifierPublic
	modifierProtected
	modifierPrivate
	modifierHashPrivate
	modifierAbstract
	modifierDestructured
	modifierGlobal
	modifierExported
	modifierUnused
	modifierRequiresQuotes
	modifierOverride
	modifierAsync
	modifierDefault
	modifierNamespace
)

func parseModifier(s string) (modifierKind, bool) {
	switch s {
	case "const":
		return modifierConst, true
	case "readonly":
		return modifierReadonly, true
	case "static":
		return modifierStatic, true
	case "public":
		return modifierPublic, true
	case "protected":
		return modifierProtected, true
	case "private":
		return modifierPrivate, true
	case "#private":
		return modifierHashPrivate, true
	case "abstract":
		return modifierAbstract, true
	case "destructured":
		return modifierDestructured, true
	case "global":
		return modifierGlobal, true
	case "exported":
		return modifierExported, true
	case "unused":
		return modifierUnused, true
	case "requiresQuotes":
		return modifierRequiresQuotes, true
	case "override":
		return modifierOverride, true
	case "async":
		return modifierAsync, true
	case "default":
		return modifierDefault, true
	case "namespace":
		return modifierNamespace, true
	default:
		return 0, false
	}
}

type typeModifierKind int

const (
	typeModBoolean typeModifierKind = 1 << iota
	typeModString
	typeModNumber
	typeModFunction
	typeModArray
)

func parseTypeModifier(s string) (typeModifierKind, bool) {
	switch s {
	case "boolean":
		return typeModBoolean, true
	case "string":
		return typeModString, true
	case "number":
		return typeModNumber, true
	case "function":
		return typeModFunction, true
	case "array":
		return typeModArray, true
	default:
		return 0, false
	}
}

// ---- Normalized config types ----

type matchRegex struct {
	regex *esregexp.RegExp
	match bool
}

const maxCachedNamingRegexps = 256

var namingRegexpCache = struct {
	sync.RWMutex
	entries map[string]*esregexp.RegExp
}{
	entries: make(map[string]*esregexp.RegExp),
}

type normalizedSelector struct {
	selector           selectorKind
	modifiers          modifierKind
	types              typeModifierKind
	typesConfigured    bool
	filter             *matchRegex
	format             []predefinedFormat // nil means "no format check" (format: null)
	formatNull         bool
	custom             *matchRegex
	leadingUnderscore  underscoreOption
	trailingUnderscore underscoreOption
	prefix             []string
	suffix             []string
	modifierWeight     int
}

type namingOptionsCacheKey struct {
	// A Go pointer in the map key keeps the slice backing array alive, so the
	// allocator cannot reuse this identity for another config.
	first  *any
	length int
}

const individualSelectorCount = 19

type parsedNamingOptions struct {
	selectors       []normalizedSelector
	neededModifiers [individualSelectorCount]modifierKind
	allModifiers    modifierKind
}

const maxCachedNamingOptions = 64

var namingOptionsCache = struct {
	sync.RWMutex
	entries map[namingOptionsCacheKey]*parsedNamingOptions
	keys    [maxCachedNamingOptions]namingOptionsCacheKey
	next    int
}{
	entries: make(map[namingOptionsCacheKey]*parsedNamingOptions),
}

var (
	defaultNamingOptionsOnce sync.Once
	defaultNamingOptions     *parsedNamingOptions
)

// getParsedNamingOptions parses each immutable config-resolver option slice once.
// Files with the same matched config shape share that exact slice, so a project
// normally has only a handful of entries even when the rule runs on thousands
// of files. The bounded cache also prevents repeated regexp compilation.
func getParsedNamingOptions(options []any) *parsedNamingOptions {
	if len(options) == 0 {
		defaultNamingOptionsOnce.Do(func() {
			defaultNamingOptions = buildParsedNamingOptions(options)
		})
		return defaultNamingOptions
	}

	key := namingOptionsCacheKey{first: &options[0], length: len(options)}
	namingOptionsCache.RLock()
	parsed, ok := namingOptionsCache.entries[key]
	namingOptionsCache.RUnlock()
	if ok {
		return parsed
	}

	parsed = buildParsedNamingOptions(options)
	namingOptionsCache.Lock()
	defer namingOptionsCache.Unlock()
	if existing, ok := namingOptionsCache.entries[key]; ok {
		return existing
	}
	if len(namingOptionsCache.entries) == maxCachedNamingOptions {
		delete(namingOptionsCache.entries, namingOptionsCache.keys[namingOptionsCache.next])
	}
	namingOptionsCache.entries[key] = parsed
	namingOptionsCache.keys[namingOptionsCache.next] = key
	namingOptionsCache.next = (namingOptionsCache.next + 1) % maxCachedNamingOptions
	return parsed
}

func buildParsedNamingOptions(options []any) *parsedNamingOptions {
	parsed := &parsedNamingOptions{selectors: parseOptions(options)}
	const quotingSelectors = selectorProperty | selectorMethod | selectorAccessor | selectorEnumMember
	for index := range individualSelectorCount {
		individual := selectorKind(1 << index)
		var needed modifierKind
		for _, sel := range parsed.selectors {
			if sel.selector&individual == 0 {
				continue
			}
			needed |= sel.modifiers
			if individual&quotingSelectors != 0 && !sel.formatNull && len(sel.format) != 0 {
				needed |= modifierRequiresQuotes
			}
		}
		// Exported declarations never receive the unused modifier, even when
		// the selector does not explicitly mention exported.
		if needed&modifierUnused != 0 {
			needed |= modifierExported
		}
		parsed.neededModifiers[index] = needed
		parsed.allModifiers |= needed
	}
	return parsed
}

// ---- Format checking functions ----
// These port typescript-eslint's character-based implementation (derived from
// tslint-consistent-codestyle), NOT a regex. The difference is load-bearing:
//   - empty strings are considered valid for every format (supports names that
//     trim down to "" after leadingUnderscore: "allow", e.g. a lone `_`)
//   - characters that are neither upper nor lowercase letters (e.g. `$`, digits)
//     are treated neutrally and pass both camelCase and PascalCase first-char
//     checks, matching JS `c === c.toLowerCase() && c === c.toUpperCase()`.
// See: https://github.com/typescript-eslint/typescript-eslint/blob/v6.21.0/packages/eslint-plugin/src/rules/naming-convention-utils/format.ts

func firstRune(name string) rune {
	for _, r := range name {
		return r
	}
	return 0
}

func isUppercaseRune(r rune) bool {
	// Upstream reads one UTF-16 code unit, and a lone surrogate has no case,
	// so a character outside the BMP counts as neither upper nor lower case.
	if r > 0xFFFF {
		return false
	}
	c := string(r)
	return ecmascript.StringToUpperCase(c) == c && ecmascript.StringToLowerCase(c) != c
}

// Matches JS `name[0] === name[0].toUpperCase()`: returns true for uppercase
// letters AND for characters with no case distinction (e.g. `$`, digits) —
// a character outside the BMP among them, since upstream indexes one UTF-16
// code unit and a lone surrogate is its own uppercase.
func firstIsUpper(name string) bool {
	r := firstRune(name)
	if r > 0xFFFF {
		return true
	}
	c := string(r)
	return ecmascript.StringToUpperCase(c) == c
}

// Matches JS `name[0] === name[0].toLowerCase()`: symmetric to firstIsUpper.
func firstIsLower(name string) bool {
	r := firstRune(name)
	if r > 0xFFFF {
		return true
	}
	c := string(r)
	return ecmascript.StringToLowerCase(c) == c
}

func checkPascalCase(name string) bool {
	if len(name) == 0 {
		return true
	}
	return firstIsUpper(name) && !strings.Contains(name, "_")
}

func checkStrictPascalCase(name string) bool {
	if len(name) == 0 {
		return true
	}
	return firstIsUpper(name) && hasStrictCamelHumps(name, true)
}

func checkCamelCase(name string) bool {
	if len(name) == 0 {
		return true
	}
	return firstIsLower(name) && !strings.Contains(name, "_")
}

func checkStrictCamelCase(name string) bool {
	if len(name) == 0 {
		return true
	}
	return firstIsLower(name) && hasStrictCamelHumps(name, false)
}

// hasStrictCamelHumps ports the typescript-eslint algorithm: no leading `_`,
// no internal `_`, and no two consecutive uppercase letters (case-distinct
// runs must alternate). `isUpper` tracks whether the *next* run is expected
// to be uppercase; on a mismatch we flip, on a match of two uppercase runs
// we reject.
func hasStrictCamelHumps(name string, isUpper bool) bool {
	runes := []rune(name)
	if len(runes) == 0 {
		return true
	}
	if runes[0] == '_' {
		return false
	}
	for i := 1; i < len(runes); i++ {
		if runes[i] == '_' {
			return false
		}
		charIsUpper := isUppercaseRune(runes[i])
		if isUpper == charIsUpper {
			if isUpper {
				return false
			}
		} else {
			isUpper = !isUpper
		}
	}
	return true
}

func checkSnakeCase(name string) bool {
	if len(name) == 0 {
		return true
	}
	return name == ecmascript.StringToLowerCase(name) && validateSnakeUnderscores(name)
}

func checkUpperCase(name string) bool {
	if len(name) == 0 {
		return true
	}
	return name == ecmascript.StringToUpperCase(name) && validateSnakeUnderscores(name)
}

// validateSnakeUnderscores rejects leading `_`, adjacent `__`, and trailing `_`.
func validateSnakeUnderscores(name string) bool {
	runes := []rune(name)
	if len(runes) == 0 {
		return true
	}
	if runes[0] == '_' {
		return false
	}
	wasUnderscore := false
	for i := 1; i < len(runes); i++ {
		if runes[i] == '_' {
			if wasUnderscore {
				return false
			}
			wasUnderscore = true
		} else {
			wasUnderscore = false
		}
	}
	return !wasUnderscore
}

func checkFormat(name string, format predefinedFormat) bool {
	switch format {
	case formatCamelCase:
		return checkCamelCase(name)
	case formatStrictCamelCase:
		return checkStrictCamelCase(name)
	case formatPascalCase:
		return checkPascalCase(name)
	case formatStrictPascalCase:
		return checkStrictPascalCase(name)
	case formatSnakeCase:
		return checkSnakeCase(name)
	case formatUpperCase:
		return checkUpperCase(name)
	default:
		return true
	}
}

// ---- Options parsing ----

func parseOptions(options []any) []normalizedSelector {
	if len(options) == 0 {
		return getDefaultConfig()
	}

	var selectors []normalizedSelector
	for _, opt := range options {
		optMap, ok := opt.(map[string]interface{})
		if !ok {
			continue
		}
		selectors = append(selectors, parseOneSelector(optMap)...)
	}

	// Sort by specificity (more specific first). Selector-kind specificity
	// (individual > method/property meta > other meta groups) always takes
	// precedence over modifierWeight — modifierWeight (modifiers/types/filter)
	// only breaks ties between selectors that target the exact same selector
	// kind. Otherwise a `filter` on a broad, earlier selector (e.g. memberLike)
	// would wrongly outrank a narrower, later one (e.g. property) that has no
	// filter at all.
	sort.SliceStable(selectors, func(i, j int) bool {
		a, b := selectors[i], selectors[j]
		if a.selector == b.selector {
			return a.modifierWeight > b.modifierWeight
		}

		aIndividual := bitCount(int(a.selector)) == 1
		bIndividual := bitCount(int(b.selector)) == 1
		if aIndividual != bIndividual {
			return aIndividual
		}

		// For backward compatibility, method and property have higher
		// precedence than other meta selectors (matches upstream
		// @typescript-eslint/eslint-plugin behavior).
		aMethodOrProperty := a.selector == selectorMethod || a.selector == selectorProperty
		bMethodOrProperty := b.selector == selectorMethod || b.selector == selectorProperty
		if aMethodOrProperty != bMethodOrProperty {
			return aMethodOrProperty
		}

		// Upstream orders remaining selectors by their numeric bitmask.
		return a.selector > b.selector
	})

	return selectors
}

func getDefaultConfig() []normalizedSelector {
	return parseOptions([]any{
		map[string]interface{}{
			"selector":           "default",
			"format":             []interface{}{"camelCase"},
			"leadingUnderscore":  "allow",
			"trailingUnderscore": "allow",
		},
		map[string]interface{}{
			"selector": "import",
			"format":   []interface{}{"camelCase", "PascalCase"},
		},
		map[string]interface{}{
			"selector":           "variable",
			"format":             []interface{}{"camelCase", "UPPER_CASE"},
			"leadingUnderscore":  "allow",
			"trailingUnderscore": "allow",
		},
		map[string]interface{}{
			"selector": "typeLike",
			"format":   []interface{}{"PascalCase"},
		},
	})
}

func parseOneSelector(optMap map[string]interface{}) []normalizedSelector {
	// Parse selector(s) - can be a string or array of strings
	var selectorKinds []selectorKind
	switch v := optMap["selector"].(type) {
	case string:
		if sk, ok := parseSelectorKind(v); ok {
			selectorKinds = append(selectorKinds, sk)
		}
	case []interface{}:
		for _, s := range v {
			if str, ok := s.(string); ok {
				if sk, ok := parseSelectorKind(str); ok {
					selectorKinds = append(selectorKinds, sk)
				}
			}
		}
	}

	if len(selectorKinds) == 0 {
		return nil
	}

	// Parse format
	var formats []predefinedFormat
	formatNull := false
	if formatVal, exists := optMap["format"]; exists {
		if formatVal == nil {
			formatNull = true
		} else if arr, ok := formatVal.([]interface{}); ok {
			for _, f := range arr {
				if str, ok := f.(string); ok {
					if pf, ok := parseFormatName(str); ok {
						formats = append(formats, pf)
					}
				}
			}
		}
	}

	// Parse modifiers
	var mods modifierKind
	if modsVal, ok := optMap["modifiers"].([]interface{}); ok {
		for _, m := range modsVal {
			if str, ok := m.(string); ok {
				if mk, ok := parseModifier(str); ok {
					mods |= mk
				}
			}
		}
	}

	// Parse types
	var types typeModifierKind
	if typesVal, ok := optMap["types"].([]interface{}); ok {
		for _, t := range typesVal {
			if str, ok := t.(string); ok {
				if tk, ok := parseTypeModifier(str); ok {
					types |= tk
				}
			}
		}
	}

	// Parse filter
	var filter *matchRegex
	if filterVal, exists := optMap["filter"]; exists {
		filter = parseMatchRegex(filterVal)
	}

	// Parse custom
	var custom *matchRegex
	if customVal, exists := optMap["custom"]; exists {
		custom = parseMatchRegex(customVal)
	}

	// Parse underscores — when not specified, use underscoreNone (no check, no strip),
	// matching the original ESLint rule behavior where omitted options mean no processing.
	leadingUnderscore := underscoreNone
	if v, ok := optMap["leadingUnderscore"].(string); ok {
		leadingUnderscore = parseUnderscoreOption(v)
	}

	trailingUnderscore := underscoreNone
	if v, ok := optMap["trailingUnderscore"].(string); ok {
		trailingUnderscore = parseUnderscoreOption(v)
	}

	// Parse prefix/suffix
	var prefix, suffix []string
	if arr, ok := optMap["prefix"].([]interface{}); ok {
		for _, p := range arr {
			if s, ok := p.(string); ok {
				prefix = append(prefix, s)
			}
		}
	}
	if arr, ok := optMap["suffix"].([]interface{}); ok {
		for _, s := range arr {
			if str, ok := s.(string); ok {
				suffix = append(suffix, str)
			}
		}
	}

	var result []normalizedSelector
	for _, sk := range selectorKinds {
		// Only apply types filter to selectors that support it
		selectorTypes := types
		if !selectorSupportsTypes(sk) {
			selectorTypes = 0
		}
		weight := calculateWeight(mods, selectorTypes, filter)
		result = append(result, normalizedSelector{
			selector:           sk,
			modifiers:          mods,
			types:              selectorTypes,
			typesConfigured:    optMap["types"] != nil,
			filter:             filter,
			format:             formats,
			formatNull:         formatNull,
			custom:             custom,
			leadingUnderscore:  leadingUnderscore,
			trailingUnderscore: trailingUnderscore,
			prefix:             prefix,
			suffix:             suffix,
			modifierWeight:     weight,
		})
	}
	return result
}

// selectorSupportsTypes returns true if the given individual selector supports the types option.
func selectorSupportsTypes(sk selectorKind) bool {
	const typeSupportedSelectors = selectorVariable | selectorParameter |
		selectorClassProperty | selectorObjectLiteralProperty | selectorTypeProperty |
		selectorParameterProperty | selectorClassicAccessor | selectorAutoAccessor
	return sk&typeSupportedSelectors != 0
}

func parseMatchRegex(val interface{}) *matchRegex {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		re := compileNamingRegexp(v)
		if re == nil {
			return nil
		}
		return &matchRegex{regex: re, match: true}
	case map[string]interface{}:
		regexStr, _ := v["regex"].(string)
		matchVal := true
		if m, ok := v["match"].(bool); ok {
			matchVal = m
		}
		re := compileNamingRegexp(regexStr)
		if re == nil {
			return nil
		}
		return &matchRegex{regex: re, match: matchVal}
	}
	return nil
}

// compileNamingRegexp shares immutable compiled programs across files using
// the same config. regexp2.Regexp is safe for concurrent use. The bounded
// cache avoids retaining an unbounded stream of patterns in long-lived LSP
// processes while removing repeated compilation for ordinary project configs.
func compileNamingRegexp(pattern string) *esregexp.RegExp {
	namingRegexpCache.RLock()
	re, ok := namingRegexpCache.entries[pattern]
	namingRegexpCache.RUnlock()
	if ok {
		return re
	}

	namingRegexpCache.Lock()
	defer namingRegexpCache.Unlock()

	if re, ok = namingRegexpCache.entries[pattern]; ok {
		return re
	}
	re, err := esregexp.Compile(pattern, "u")
	if err != nil {
		return nil
	}
	if len(namingRegexpCache.entries) < maxCachedNamingRegexps {
		namingRegexpCache.entries[pattern] = re
	}
	return re
}

// calculateWeight mirrors the original ESLint rule's modifierWeight
// calculation. It is used only to break ties between selectors that target
// the exact same selector kind — selector-kind specificity itself is decided
// separately (see parseOptions' sort), never by this weight.
func calculateWeight(mods modifierKind, types typeModifierKind, filter *matchRegex) int {
	weight := 0

	// Raw bitmask value (not bitcount): higher-valued modifiers (e.g.,
	// unused=2048) take priority over lower-valued ones (e.g.,
	// destructured=256). Shift types past the modifier bit range so the two
	// don't collide (both modifierKind and typeModifierKind start at 1<<0).
	weight |= int(mods)
	weight |= int(types) << 17

	// Give selectors with a filter the highest priority among same-kind
	// selectors.
	if filter != nil {
		weight |= 1 << 30
	}
	return weight
}

func bitCount(x int) int {
	return bits.OnesCount(uint(x))
}

// ---- Message helpers ----

func selectorTypeToMessageString(sk selectorKind) string {
	switch sk {
	case selectorVariable:
		return "Variable"
	case selectorFunction:
		return "Function"
	case selectorParameter:
		return "Parameter"
	case selectorClassProperty:
		return "Class Property"
	case selectorObjectLiteralProperty:
		return "Object Literal Property"
	case selectorTypeProperty:
		return "Type Property"
	case selectorParameterProperty:
		return "Parameter Property"
	case selectorEnumMember:
		return "Enum Member"
	case selectorClassMethod:
		return "Class Method"
	case selectorObjectLiteralMethod:
		return "Object Literal Method"
	case selectorTypeMethod:
		return "Type Method"
	case selectorClassicAccessor:
		return "Classic Accessor"
	case selectorAutoAccessor:
		return "Auto Accessor"
	case selectorClass:
		return "Class"
	case selectorInterface:
		return "Interface"
	case selectorTypeAlias:
		return "Type Alias"
	case selectorEnum:
		return "Enum"
	case selectorTypeParameter:
		return "Type Parameter"
	case selectorImport:
		return "Import"
	default:
		return "Identifier"
	}
}

func doesNotMatchFormatMessage(typeName, name string, formats []predefinedFormat) rule.RuleMessage {
	var fmtNames []string
	for _, f := range formats {
		fmtNames = append(fmtNames, formatName(f))
	}
	return rule.RuleMessage{
		Id:          "doesNotMatchFormat",
		Description: fmt.Sprintf("%s name `%s` must match one of the following formats: %s", typeName, name, strings.Join(fmtNames, ", ")),
	}
}

func doesNotMatchFormatTrimmedMessage(typeName, name, processedName string, formats []predefinedFormat) rule.RuleMessage {
	var fmtNames []string
	for _, f := range formats {
		fmtNames = append(fmtNames, formatName(f))
	}
	return rule.RuleMessage{
		Id:          "doesNotMatchFormatTrimmed",
		Description: fmt.Sprintf("%s name `%s` trimmed as `%s` must match one of the following formats: %s", typeName, name, processedName, strings.Join(fmtNames, ", ")),
	}
}

func missingAffixMessage(typeName, name, position string, affixes []string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAffix",
		Description: fmt.Sprintf("%s name `%s` must have one of the following %ses: %s", typeName, name, position, strings.Join(affixes, ", ")),
	}
}

func satisfyCustomMessage(typeName, name string, regexMatch bool, regex string) rule.RuleMessage {
	matchStr := "match"
	if !regexMatch {
		matchStr = "not match"
	}
	return rule.RuleMessage{
		Id:          "satisfyCustom",
		Description: fmt.Sprintf("%s name `%s` must %s the RegExp: /%s/u", typeName, name, matchStr, regexpSourceForMessage(regex)),
	}
}

// RegExp.prototype.source escapes literal slashes and line terminators even
// when the option was supplied as a string rather than a regexp literal.
func regexpSourceForMessage(pattern string) string {
	if pattern == "" {
		return "(?:)"
	}
	var result strings.Builder
	escaped := false
	inClass := false
	for _, char := range pattern {
		switch char {
		case '[':
			if !escaped {
				inClass = true
			}
			result.WriteRune(char)
		case ']':
			if !escaped {
				inClass = false
			}
			result.WriteRune(char)
		case '/':
			if !escaped && !inClass {
				result.WriteByte('\\')
			}
			result.WriteRune(char)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\u2028':
			result.WriteString(`\u2028`)
		case '\u2029':
			result.WriteString(`\u2029`)
		default:
			result.WriteRune(char)
		}
		escaped = char == '\\' && !escaped
	}
	return result.String()
}

func unexpectedUnderscoreMessage(typeName, name, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedUnderscore",
		Description: fmt.Sprintf("%s name `%s` must not have a %s underscore.", typeName, name, position),
	}
}

func missingUnderscoreMessage(typeName, name string, count int, position string) rule.RuleMessage {
	countWord := "one"
	if count == 2 {
		countWord = "two"
	}
	return rule.RuleMessage{
		Id:          "missingUnderscore",
		Description: fmt.Sprintf("%s name `%s` must have %s %s underscore(s).", typeName, name, countWord, position),
	}
}

// ---- Validation logic ----

type validationResult struct {
	valid   bool
	message *rule.RuleMessage
}

func validate(name string, sel normalizedSelector, idMods modifierKind, idSelector selectorKind) validationResult {
	typeName := selectorTypeToMessageString(idSelector)

	// Filter is checked in validateIdentifier as part of selector matching,
	// not here, so that non-matching filters cause the selector to be skipped
	// (allowing the next selector in specificity order to match).

	processedName := name

	// 3. Validate leading underscore
	processedName, result := validateUnderscore("leading", processedName, typeName, name, sel.leadingUnderscore)
	if !result.valid {
		return result
	}

	// 4. Validate trailing underscore
	processedName, result = validateUnderscore("trailing", processedName, typeName, name, sel.trailingUnderscore)
	if !result.valid {
		return result
	}

	// 5. Validate prefix
	if len(sel.prefix) > 0 {
		found := false
		for _, p := range sel.prefix {
			if strings.HasPrefix(processedName, p) {
				processedName = processedName[len(p):]
				found = true
				break
			}
		}
		if !found {
			msg := missingAffixMessage(typeName, name, "prefix", sel.prefix)
			return validationResult{valid: false, message: &msg}
		}
	}

	// 6. Validate suffix
	if len(sel.suffix) > 0 {
		found := false
		for _, s := range sel.suffix {
			if strings.HasSuffix(processedName, s) {
				processedName = processedName[:len(processedName)-len(s)]
				found = true
				break
			}
		}
		if !found {
			msg := missingAffixMessage(typeName, name, "suffix", sel.suffix)
			return validationResult{valid: false, message: &msg}
		}
	}

	// 7. Validate custom regex (against processed name, after stripping underscores and affixes)
	if sel.custom != nil {
		matches, err := sel.custom.regex.TestOrError(processedName)
		if err == nil && sel.custom.match != matches {
			msg := satisfyCustomMessage(typeName, name, sel.custom.match, sel.custom.regex.Source())
			return validationResult{valid: false, message: &msg}
		}
	}

	// 8. Validate format
	if sel.formatNull {
		return validationResult{valid: true}
	}

	if len(sel.format) == 0 {
		return validationResult{valid: true}
	}

	// If the identifier doesn't require quoting, check format normally.
	// If it does require quoting, skip format checks (name can never match any predefined format)
	// and fall through to report an error.
	if idMods&modifierRequiresQuotes == 0 {
		for _, f := range sel.format {
			if checkFormat(processedName, f) {
				return validationResult{valid: true}
			}
		}
	}

	// Format check failed
	if processedName != name {
		msg := doesNotMatchFormatTrimmedMessage(typeName, name, processedName, sel.format)
		return validationResult{valid: false, message: &msg}
	}
	msg := doesNotMatchFormatMessage(typeName, name, sel.format)
	return validationResult{valid: false, message: &msg}
}

func validateUnderscore(position string, processedName string, typeName string, originalName string, option underscoreOption) (string, validationResult) {
	isLeading := position == "leading"

	switch option {
	case underscoreNone:
		// Not specified — no check, no stripping. Pass through unchanged.
		return processedName, validationResult{valid: true}
	case underscoreForbid:
		if isLeading {
			if len(processedName) > 0 && processedName[0] == '_' {
				msg := unexpectedUnderscoreMessage(typeName, originalName, position)
				return processedName, validationResult{valid: false, message: &msg}
			}
		} else {
			if len(processedName) > 0 && processedName[len(processedName)-1] == '_' {
				msg := unexpectedUnderscoreMessage(typeName, originalName, position)
				return processedName, validationResult{valid: false, message: &msg}
			}
		}
	case underscoreRequire:
		if isLeading {
			if len(processedName) == 0 || processedName[0] != '_' {
				msg := missingUnderscoreMessage(typeName, originalName, 1, position)
				return processedName, validationResult{valid: false, message: &msg}
			}
			processedName = processedName[1:]
		} else {
			if len(processedName) == 0 || processedName[len(processedName)-1] != '_' {
				msg := missingUnderscoreMessage(typeName, originalName, 1, position)
				return processedName, validationResult{valid: false, message: &msg}
			}
			processedName = processedName[:len(processedName)-1]
		}
	case underscoreRequireDouble:
		if isLeading {
			if !strings.HasPrefix(processedName, "__") {
				msg := missingUnderscoreMessage(typeName, originalName, 2, position)
				return processedName, validationResult{valid: false, message: &msg}
			}
			processedName = processedName[2:]
		} else {
			if !strings.HasSuffix(processedName, "__") {
				msg := missingUnderscoreMessage(typeName, originalName, 2, position)
				return processedName, validationResult{valid: false, message: &msg}
			}
			processedName = processedName[:len(processedName)-2]
		}
	case underscoreAllow:
		// Strip single underscore if present
		if isLeading {
			if len(processedName) > 0 && processedName[0] == '_' {
				processedName = processedName[1:]
			}
		} else {
			if len(processedName) > 0 && processedName[len(processedName)-1] == '_' {
				processedName = processedName[:len(processedName)-1]
			}
		}
	case underscoreAllowDouble:
		// Strip double underscore if present
		if isLeading {
			processedName = strings.TrimPrefix(processedName, "__")
		} else {
			processedName = strings.TrimSuffix(processedName, "__")
		}
	case underscoreAllowSingleOrDouble:
		// Strip up to two underscores if present
		if isLeading {
			if strings.HasPrefix(processedName, "__") {
				processedName = processedName[2:]
			} else if len(processedName) > 0 && processedName[0] == '_' {
				processedName = processedName[1:]
			}
		} else {
			if strings.HasSuffix(processedName, "__") {
				processedName = processedName[:len(processedName)-2]
			} else if len(processedName) > 0 && processedName[len(processedName)-1] == '_' {
				processedName = processedName[:len(processedName)-1]
			}
		}
	}

	return processedName, validationResult{valid: true}
}

// ---- Type checking helpers ----

func isCorrectType(ch *checker.Checker, node *ast.Node, types typeModifierKind) bool {
	if ch == nil {
		return false
	}
	t := ch.GetNonNullableType(ch.GetTypeAtLocation(node))
	if t == nil {
		return false
	}
	return checkTypeMatch(ch, t, types)
}

func checkTypeMatch(ch *checker.Checker, t *checker.Type, types typeModifierKind) bool {
	if types&(typeModBoolean|typeModString|typeModNumber) != 0 {
		base := checker.Checker_getWidenedType(ch, checker.Checker_getBaseTypeOfLiteralType(ch, t))
		switch ch.TypeToString(base) {
		case "boolean":
			if types&typeModBoolean != 0 {
				return true
			}
		case "string":
			if types&typeModString != 0 {
				return true
			}
		case "number":
			if types&typeModNumber != 0 {
				return true
			}
		}
	}
	allPartsMatch := func(matches func(*checker.Type) bool) bool {
		for _, part := range utils.UnionTypeParts(t) {
			if !matches(part) {
				return false
			}
		}
		return true
	}
	if types&typeModFunction != 0 && allPartsMatch(func(part *checker.Type) bool {
		return len(utils.GetCallSignatures(ch, part)) > 0
	}) {
		return true
	}
	return types&typeModArray != 0 && allPartsMatch(func(part *checker.Type) bool {
		return checker.Checker_isArrayOrTupleType(ch, part)
	})
}

// ---- Modifier detection helpers ----

func getModifiers(
	ctx rule.RuleContext,
	node *ast.Node,
	nameNode *ast.Node,
	sel selectorKind,
	name string,
	reExportedNames map[*ast.Symbol]bool,
	destructured bool,
	refInfo *referencedInfo,
	needed modifierKind,
) modifierKind {
	// Group selectors may request modifiers absent from some member kinds.
	// Preserve the modifiers attached by each upstream visitor, rather than
	// deriving additional ones from unrelated TypeScript declaration flags.
	switch sel {
	case selectorVariable:
		needed &= modifierConst | modifierGlobal | modifierExported | modifierUnused | modifierAsync | modifierDestructured
	case selectorFunction:
		needed &= modifierGlobal | modifierExported | modifierUnused | modifierAsync
	case selectorParameter:
		needed &= modifierUnused | modifierDestructured
	case selectorClass:
		needed &= modifierAbstract | modifierExported | modifierUnused
	case selectorInterface, selectorTypeAlias, selectorEnum:
		needed &= modifierExported | modifierUnused
	case selectorTypeParameter:
		needed &= modifierUnused
	case selectorEnumMember:
		needed &= modifierRequiresQuotes
	case selectorTypeMethod:
		needed &= modifierPublic | modifierRequiresQuotes
	case selectorTypeProperty:
		needed &= modifierPublic | modifierReadonly | modifierRequiresQuotes
	case selectorObjectLiteralProperty:
		needed &= modifierPublic | modifierRequiresQuotes
	case selectorObjectLiteralMethod:
		needed &= modifierPublic | modifierRequiresQuotes | modifierAsync
	case selectorImport:
		needed &= modifierDefault | modifierNamespace
	}

	var mods modifierKind

	const flagModifiers = modifierPublic | modifierProtected | modifierPrivate |
		modifierStatic | modifierReadonly | modifierAbstract | modifierAsync |
		modifierConst | modifierOverride | modifierExported
	var flags ast.ModifierFlags
	if needed&flagModifiers != 0 {
		flags = ast.GetCombinedModifierFlags(node)
	}

	// Access modifiers
	if needed&modifierPublic != 0 && flags&ast.ModifierFlagsPublic != 0 {
		mods |= modifierPublic
	}
	if needed&modifierProtected != 0 && flags&ast.ModifierFlagsProtected != 0 {
		mods |= modifierProtected
	}
	if needed&modifierPrivate != 0 && flags&ast.ModifierFlagsPrivate != 0 {
		mods |= modifierPrivate
	}

	// Other modifiers
	if needed&modifierStatic != 0 && flags&ast.ModifierFlagsStatic != 0 {
		mods |= modifierStatic
	}
	if needed&modifierReadonly != 0 && flags&ast.ModifierFlagsReadonly != 0 {
		mods |= modifierReadonly
	}
	if needed&modifierAbstract != 0 && flags&ast.ModifierFlagsAbstract != 0 {
		mods |= modifierAbstract
	}
	if needed&modifierAsync != 0 && flags&ast.ModifierFlagsAsync != 0 {
		mods |= modifierAsync
	}

	// Detect async modifier from function expression/arrow function initializer.
	// For `AsyncBar = async () => {}` (class property or object literal property)
	// and `const asyncFoo = async () => {}` (variable), the async flag is on
	// the initializer, not on the declaration itself.
	if needed&modifierAsync != 0 && mods&modifierAsync == 0 {
		switch node.Kind {
		case ast.KindPropertyDeclaration:
			if init := node.AsPropertyDeclaration().Initializer; init != nil {
				init = ast.SkipParentheses(init)
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					if ast.GetCombinedModifierFlags(init)&ast.ModifierFlagsAsync != 0 {
						mods |= modifierAsync
					}
				}
			}
		case ast.KindPropertyAssignment:
			if init := node.AsPropertyAssignment().Initializer; init != nil {
				init = ast.SkipParentheses(init)
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					if ast.GetCombinedModifierFlags(init)&ast.ModifierFlagsAsync != 0 {
						mods |= modifierAsync
					}
				}
			}
		case ast.KindVariableDeclaration:
			if init := node.AsVariableDeclaration().Initializer; init != nil {
				init = ast.SkipParentheses(init)
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					if ast.GetCombinedModifierFlags(init)&ast.ModifierFlagsAsync != 0 {
						mods |= modifierAsync
					}
				}
			}
		}
	}

	if needed&modifierConst != 0 && flags&ast.ModifierFlagsConst != 0 {
		mods |= modifierConst
	}
	if needed&modifierOverride != 0 && flags&ast.ModifierFlagsOverride != 0 {
		mods |= modifierOverride
	}
	// Accessor keyword check is handled by the selector system, not as a modifier

	// Hash private (ECMAScript private)
	if needed&modifierHashPrivate != 0 && nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
		mods |= modifierHashPrivate
	}

	// Default accessibility: if no access modifier on a class member, it's public
	if needed&modifierPublic != 0 && sel&selectorMemberLike != 0 &&
		flags&(ast.ModifierFlagsPublic|ast.ModifierFlagsProtected|ast.ModifierFlagsPrivate) == 0 &&
		(nameNode == nil || nameNode.Kind != ast.KindPrivateIdentifier) {
		mods |= modifierPublic
	}

	// Export check (direct export or re-export via export { ... })
	if needed&modifierExported != 0 && (isExported(node, flags) || reExportedNames[utils.BindingNameSymbol(nameNode)]) {
		mods |= modifierExported
	}

	// Global check
	if needed&modifierGlobal != 0 && isGlobalScope(node) {
		mods |= modifierGlobal
	}

	// Destructured check (per-identifier, set during extraction)
	if needed&modifierDestructured != 0 && destructured {
		mods |= modifierDestructured
	}

	// Unused check - exported symbols are never considered unused, and only
	// scope-participating declarations (not members or imports) can be unused.
	// Use nameNode (the specific identifier) for symbol lookup, not the declaration node,
	// so that individual destructured bindings are correctly detected as unused.
	if needed&modifierUnused != 0 && sel&selectorUnusedSupported != 0 &&
		mods&modifierExported == 0 && isUnusedByNameNode(ctx, nameNode, refInfo) {
		mods |= modifierUnused
	}

	// Const check for variables
	if needed&modifierConst != 0 && sel&selectorVariable != 0 {
		if isConstVariable(node) {
			mods |= modifierConst
		}
	}

	// Const check for enum
	if needed&modifierConst != 0 && sel&selectorEnum != 0 {
		if flags&ast.ModifierFlagsConst != 0 {
			mods |= modifierConst
		}
	}

	// requiresQuotes check for members
	if needed&modifierRequiresQuotes != 0 && sel&(selectorProperty|selectorMethod|selectorAccessor|selectorEnumMember) != 0 {
		if requiresQuoting(nameNode) {
			mods |= modifierRequiresQuotes
		}
	}

	// Import modifier detection (default vs named vs namespace)
	if sel&selectorImport != 0 && needed&(modifierDefault|modifierNamespace) != 0 {
		switch node.Kind {
		case ast.KindImportClause:
			// `import Foo from ...` is a default import
			if needed&modifierDefault != 0 {
				mods |= modifierDefault
			}
		case ast.KindNamespaceImport:
			// `import * as Foo from ...` is a namespace import
			if needed&modifierNamespace != 0 {
				mods |= modifierNamespace
			}
		case ast.KindImportSpecifier:
			// `import { default as Foo } from ...` is also a default import
			importSpec := node.AsImportSpecifier()
			if needed&modifierDefault != 0 && importSpec.PropertyName != nil {
				if importSpec.PropertyName.Kind == ast.KindStringLiteral || importSpec.PropertyName.Text() == "default" {
					mods |= modifierDefault
				}
			}
		}
	}

	return mods
}

func isExported(node *ast.Node, flags ast.ModifierFlags) bool {
	// Check if node itself has export keyword
	if flags&ast.ModifierFlagsExport != 0 {
		return true
	}

	parent := node.Parent
	if parent == nil {
		return false
	}

	// Check variable statement: `export const x = 1`
	if ast.IsVariableDeclaration(node) {
		declList := node.Parent
		if declList != nil && declList.Kind == ast.KindVariableDeclarationList {
			varStmt := declList.Parent
			if varStmt != nil && varStmt.Kind == ast.KindVariableStatement {
				stmtFlags := ast.GetCombinedModifierFlags(varStmt)
				if stmtFlags&ast.ModifierFlagsExport != 0 {
					return true
				}
			}
		}
	}

	return false
}

func isGlobalScope(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Variable: check if the VariableStatement is at top level
	if ast.IsVariableDeclaration(node) {
		declList := node.Parent
		if declList != nil && declList.Kind == ast.KindVariableDeclarationList {
			varStmt := declList.Parent
			if varStmt != nil {
				if (ast.IsForInOrOfStatement(varStmt) || varStmt.Kind == ast.KindForStatement) && declList.Flags&ast.NodeFlagsBlockScoped != 0 {
					return false
				}
				return isTopLevelScope(varStmt.Parent)
			}
		}
		return false
	}

	return isTopLevelScope(parent)
}

func isTopLevelScope(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.KindSourceFile
}

type referencedInfo struct {
	refs           *rule.RefStore
	symbols        map[*ast.Symbol]bool
	shorthandNames map[string]bool // names used in shorthand property assignments
}

func getReferencedInfo(ctx rule.RuleContext) *referencedInfo {
	if ctx.Refs != nil {
		return &referencedInfo{refs: ctx.Refs}
	}
	return collectReferencedSymbols(ctx)
}

// collectReferencedSymbols walks the source file and collects all symbols that are referenced
// (i.e., used in a non-declaration context). This is used to detect the "unused" modifier.
//
// The symbol set is only ever queried for symbols of declaration names in this file,
// and any reference resolving to such a symbol is an identifier with the same text as
// the declared name. So the walk first gathers all declaration-name texts in one
// pure-syntax pass, then consults the checker only for identifiers matching one of
// them — GetSymbolAtLocation is expensive (property accesses trigger lazy type
// checking; failed resolutions trigger spelling suggestions).
func collectReferencedSymbols(ctx rule.RuleContext) *referencedInfo {
	if ctx.TypeChecker == nil {
		return nil
	}

	declNames := make(map[string]bool)
	var collectDeclNames func(node *ast.Node)
	collectDeclNames = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && node.Parent != nil &&
			ast.GetNameOfDeclaration(node.Parent) == node {
			declNames[node.AsIdentifier().Text] = true
		}
		node.ForEachChild(func(child *ast.Node) bool {
			collectDeclNames(child)
			return false
		})
	}
	collectDeclNames(&ctx.SourceFile.Node)

	info := &referencedInfo{
		symbols:        make(map[*ast.Symbol]bool),
		shorthandNames: make(map[string]bool),
	}
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier {
			parent := node.Parent
			isDeclarationName := parent != nil && ast.GetNameOfDeclaration(parent) == node
			if !isDeclarationName && declNames[node.AsIdentifier().Text] {
				sym := ctx.TypeChecker.GetSymbolAtLocation(node)
				if sym != nil {
					info.symbols[sym] = true
				}
			}
		}
		// ShorthandPropertyAssignment `{ foo }` references the variable `foo`,
		// but GetSymbolAtLocation returns the property symbol, not the variable
		// symbol. Track names so isUnused can check by name as fallback.
		if node.Kind == ast.KindShorthandPropertyAssignment {
			nameNode := ast.GetNameOfDeclaration(node)
			if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
				info.shorthandNames[nameNode.AsIdentifier().Text] = true
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	walk(&ctx.SourceFile.Node)
	return info
}

// isUnusedByNameNode checks if a specific identifier node is unused by looking up
// its symbol directly, rather than going through the declaration node.
// This correctly handles individual bindings in destructuring patterns.
func isUnusedByNameNode(ctx rule.RuleContext, nameNode *ast.Node, refInfo *referencedInfo) bool {
	if refInfo == nil || nameNode == nil {
		return false
	}
	if refInfo.refs != nil {
		symbol := utils.BindingNameSymbol(nameNode)
		if symbol == nil {
			return false
		}
		declaration := nameNode.Parent
		if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && declaration.Parent.Parent != nil &&
			ast.IsForInOrOfStatement(declaration.Parent.Parent) && namingLoopReturns(declaration.Parent.Parent) {
			return false
		}
		if declaration.Kind == ast.KindFunctionExpression || declaration.Kind == ast.KindClassExpression {
			return false
		}
		for parent := declaration; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindParameter {
				owner := parent.Parent
				if nameNode.Text() == "this" || owner.Kind == ast.KindSetAccessor || owner.Body() == nil {
					return false
				}
				break
			}
			if ast.IsFunctionLikeDeclaration(parent) {
				break
			}
		}
		for _, reference := range refInfo.refs.References(symbol) {
			if isNamingRead(ctx.SourceFile, reference, symbol, refInfo.refs) {
				return false
			}
		}
		return true
	}
	if ctx.TypeChecker == nil {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if symbol == nil {
		return false
	}

	if refInfo.symbols[symbol] {
		return false
	}

	// Shorthand property fallback: `{ foo }` references variable `foo` but
	// GetSymbolAtLocation on the shorthand returns the property symbol.
	// Check by name to catch this case.
	if nameNode.Kind == ast.KindIdentifier && refInfo.shorthandNames[nameNode.AsIdentifier().Text] {
		return false
	}

	return true
}

// isNamingRead mirrors collectVariables' distinction between a reference
// and a use: writes, discarded self-updates, type queries and recursive
// references inside the binding's own definition do not make it used.
func isNamingRead(sourceFile *ast.SourceFile, reference *ast.Node, symbol *ast.Symbol, refs *rule.RefStore) bool {
	if ast.IsPartOfTypeQuery(reference) {
		return false
	}
	for node := reference; node != nil && (ast.IsIdentifier(node) || node.Kind == ast.KindQualifiedName); node = node.Parent {
		if node.Parent != nil && node.Parent.Kind == ast.KindTypePredicate {
			return false
		}
	}
	for _, declaration := range symbol.Declarations {
		container := declaration
		if declaration.Kind == ast.KindVariableDeclaration {
			initializer := declaration.Initializer()
			if initializer == nil {
				continue
			}
			container = ast.SkipParentheses(initializer)
			if !ast.IsArrowFunction(container) && !ast.IsFunctionExpression(container) {
				continue
			}
		} else {
			switch declaration.Kind {
			case ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration, ast.KindClassDeclaration:
			default:
				continue
			}
		}
		start := container.Pos()
		if container.Kind == ast.KindClassDeclaration {
			// A class's own decorators execute outside its class scope. tsgo
			// includes them in the declaration range, unlike the upstream AST.
			start = utils.NodeTextRangeSkippingDecorators(sourceFile, container).Pos()
		}
		if reference.Pos() >= start && reference.End() <= container.End() {
			return false
		}
	}
	if utils.IsWriteReference(reference) {
		target := ast.GetAssignmentTarget(reference)
		if target != nil {
			if target.Kind == ast.KindBinaryExpression {
				op := target.AsBinaryExpression().OperatorToken.Kind
				if op == ast.KindEqualsToken {
					return false
				}
				if op == ast.KindAmpersandAmpersandEqualsToken || op == ast.KindBarBarEqualsToken || op == ast.KindQuestionQuestionEqualsToken {
					return true
				}
			}
			if ast.IsForInOrOfStatement(target) {
				return target.Initializer() == reference && namingLoopReturns(target)
			}
			if unusedNamingExpression(target) {
				return false
			}
		}
	}
	// In an ordinary discarded assignment, reading the binding only to
	// compute its new value is a self-use. A read in another function or
	// a loop can consume a value written by an earlier execution.
	if len(symbol.Declarations) == 0 || namingExecutionScope(reference) != namingExecutionScope(symbol.Declarations[0]) {
		return true
	}
	for parent := reference.Parent; parent != nil; parent = parent.Parent {
		if ast.IsIterationStatement(parent, false) {
			return true
		}
		if ast.IsFunctionLikeDeclaration(parent) {
			break
		}
	}

	for parent := reference.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLikeDeclaration(parent) || ast.IsIterationStatement(parent, false) {
			break
		}
		if parent.Kind != ast.KindBinaryExpression {
			continue
		}
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken || !unusedNamingExpression(parent) {
			continue
		}
		left := ast.SkipParentheses(binary.Left)
		if ast.IsIdentifier(left) && refs.ResolveInFile(left) == symbol && reference.Pos() >= binary.Right.Pos() && reference.End() <= binary.Right.End() {
			return false
		}
	}
	return true
}

func namingExecutionScope(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLikeDeclaration(parent) || parent.Kind == ast.KindSourceFile || parent.Kind == ast.KindClassStaticBlockDeclaration || parent.Kind == ast.KindModuleBlock {
			return parent
		}
	}
	return nil
}

func namingLoopReturns(loop *ast.Node) bool {
	body := loop.AsForInOrOfStatement().Statement
	if body != nil && body.Kind == ast.KindBlock {
		statements := body.AsBlock().Statements
		if statements == nil || len(statements.Nodes) != 1 {
			return false
		}
		body = statements.Nodes[0]
	}
	return body != nil && body.Kind == ast.KindReturnStatement
}

func unusedNamingExpression(node *ast.Node) bool {
	for node.Parent != nil {
		parent := node.Parent
		if ast.IsParenthesizedExpression(parent) {
			node = parent
			continue
		}
		if parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
			if parent.AsBinaryExpression().Left == node {
				return true
			}
			node = parent
			continue
		}
		return parent.Kind == ast.KindExpressionStatement
	}
	return false
}

func isConstVariable(node *ast.Node) bool {
	if !ast.IsVariableDeclaration(node) {
		return false
	}

	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclarationList {
		return false
	}

	// Check if the declaration list uses `const`
	declList := parent.AsVariableDeclarationList()
	return declList.Flags&ast.NodeFlagsConst != 0
}

func requiresQuoting(nameNode *ast.Node) bool {
	if nameNode == nil {
		return false
	}

	if nameNode.Kind == ast.KindComputedPropertyName {
		return true
	}

	if nameNode.Kind == ast.KindStringLiteral {
		// Only requires quoting if the string value is not a valid JS identifier
		text := nameNode.AsStringLiteral().Text
		return !scanner.IsValidIdentifier(text)
	}

	if nameNode.Kind == ast.KindNumericLiteral {
		// Numeric literals are never valid identifiers
		return true
	}

	return false
}

// ---- Name extraction helper ----

// getNameFromNode extracts the name string from a member name node, handling string/numeric literals
// that GetNameFromMember doesn't handle directly.
// Returns ("", false) for computed property names, matching the original ESLint rule
// which skips all computed members via [computed = false] selector filters.
func getNameFromNode(ctx rule.RuleContext, nameNode *ast.Node) (string, bool) {
	switch nameNode.Kind {
	case ast.KindComputedPropertyName:
		// The original ESLint rule uses [computed = false] selectors to skip
		// all computed members — even string literal computed names like ["foo"].
		return "", false
	case ast.KindStringLiteral:
		return nameNode.AsStringLiteral().Text, true
	case ast.KindNumericLiteral:
		return nameNode.Text(), true
	case ast.KindPrivateIdentifier:
		// Strip the '#' prefix from private identifiers
		return strings.TrimPrefix(nameNode.Text(), "#"), true
	default:
		name, memberType := utils.GetNameFromMember(ctx.SourceFile, nameNode)
		if memberType == utils.MemberNameTypeExpression {
			return "", false
		}
		return name, true
	}
}

// ---- Identifier extraction from different node types ----

type identifierInfo struct {
	name         string
	node         *ast.Node // the name node for reporting
	declNode     *ast.Node // the declaration node for modifier detection
	selector     selectorKind
	destructured bool // whether this identifier comes from a destructured binding (without rename)
}

func getIdentifierFromNode(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	switch node.Kind {
	case ast.KindVariableStatement:
		return getIdentifiersFromVariableStatement(node, result)
	case ast.KindForOfStatement, ast.KindForInStatement, ast.KindForStatement:
		return getIdentifiersFromForLoopInitializer(node, result)
	case ast.KindFunctionDeclaration:
		return getIdentifiersFromFunctionDeclaration(node, result)
	case ast.KindFunctionExpression:
		return getIdentifiersFromFunctionExpression(node, result)
	case ast.KindParameter:
		return getIdentifiersFromParameter(node, result)
	case ast.KindClassDeclaration:
		return getIdentifiersFromClassDeclaration(node, result)
	case ast.KindClassExpression:
		return getIdentifiersFromClassExpression(node, result)
	case ast.KindInterfaceDeclaration:
		return getIdentifiersFromInterfaceDeclaration(node, result)
	case ast.KindTypeAliasDeclaration:
		return getIdentifiersFromTypeAliasDeclaration(node, result)
	case ast.KindEnumDeclaration:
		return getIdentifiersFromEnumDeclaration(node, result)
	case ast.KindEnumMember:
		return getIdentifiersFromEnumMember(ctx, node, result)
	case ast.KindTypeParameter:
		return getIdentifiersFromTypeParameter(node, result)
	case ast.KindPropertyDeclaration:
		return getIdentifiersFromPropertyDeclaration(ctx, node, result)
	case ast.KindMethodDeclaration:
		return getIdentifiersFromMethodDeclaration(ctx, node, result)
	case ast.KindGetAccessor, ast.KindSetAccessor:
		return getIdentifiersFromAccessorDeclaration(ctx, node, result)
	case ast.KindPropertySignature:
		return getIdentifiersFromPropertySignature(ctx, node, result)
	case ast.KindMethodSignature:
		return getIdentifiersFromMethodSignature(ctx, node, result)
	case ast.KindPropertyAssignment:
		return getIdentifiersFromPropertyAssignment(ctx, node, result)
	case ast.KindShorthandPropertyAssignment:
		return getIdentifiersFromShorthandPropertyAssignment(node, result)
	case ast.KindImportClause:
		return getIdentifiersFromImportClause(node, result)
	case ast.KindImportSpecifier:
		return getIdentifiersFromImportSpecifier(node, result)
	case ast.KindNamespaceImport:
		return getIdentifiersFromNamespaceImport(node, result)
	default:
		return result
	}
}

func getIdentifiersFromVariableStatement(node *ast.Node, result []identifierInfo) []identifierInfo {
	varStmt := node.AsVariableStatement()
	if varStmt.DeclarationList == nil {
		return result
	}

	declList := varStmt.DeclarationList.AsVariableDeclarationList()

	for _, decl := range declList.Declarations.Nodes {
		nameNode := decl.Name()
		if nameNode == nil {
			continue
		}

		switch nameNode.Kind {
		case ast.KindIdentifier:
			result = append(result, identifierInfo{
				name:     nameNode.AsIdentifier().Text,
				node:     nameNode,
				declNode: decl,
				selector: selectorVariable,
			})
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			start := len(result)
			result = getIdentifiersFromBindingPattern(nameNode, result)
			for i := start; i < len(result); i++ {
				result[i].declNode = decl
			}
		}
	}
	return result
}

// getIdentifiersFromForLoopInitializer extracts variable identifiers from
// for-of, for-in, and for statement initializers (e.g., `for (const x of arr)`).
// These are VariableDeclarationLists that are NOT inside a VariableStatement.
func getIdentifiersFromForLoopInitializer(node *ast.Node, result []identifierInfo) []identifierInfo {
	init := node.Initializer()
	if init == nil || init.Kind != ast.KindVariableDeclarationList {
		return result
	}

	declList := init.AsVariableDeclarationList()

	for _, decl := range declList.Declarations.Nodes {
		nameNode := decl.Name()
		if nameNode == nil {
			continue
		}

		switch nameNode.Kind {
		case ast.KindIdentifier:
			result = append(result, identifierInfo{
				name:     nameNode.AsIdentifier().Text,
				node:     nameNode,
				declNode: decl,
				selector: selectorVariable,
			})
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			start := len(result)
			result = getIdentifiersFromBindingPattern(nameNode, result)
			for i := start; i < len(result); i++ {
				result[i].declNode = decl
			}
		}
	}
	return result
}

func getIdentifiersFromBindingPattern(pattern *ast.Node, result []identifierInfo) []identifierInfo {
	elements := ast.GetElementsOfBindingOrAssignmentPattern(pattern)
	for _, elem := range elements {
		if elem.Kind == ast.KindBindingElement {
			nameNode := elem.Name()
			if nameNode == nil {
				continue
			}
			// A binding element is "destructured" only if it's in an OBJECT pattern
			// without a rename (shorthand property). Array pattern elements are NOT
			// considered destructured, matching the original ESLint rule behavior.
			// e.g., `{ foo }` → destructured, `{ bar: foo }` → NOT, `[foo]` → NOT.
			isDestructuredBinding := pattern.Kind == ast.KindObjectBindingPattern &&
				elem.AsBindingElement().PropertyName == nil && elem.AsBindingElement().DotDotDotToken == nil
			switch nameNode.Kind {
			case ast.KindIdentifier:
				result = append(result, identifierInfo{
					name:         nameNode.AsIdentifier().Text,
					node:         nameNode,
					selector:     selectorVariable,
					destructured: isDestructuredBinding,
				})
			case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
				result = getIdentifiersFromBindingPattern(nameNode, result)
			}
		}
	}

	return result
}

func getIdentifiersFromFunctionDeclaration(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorFunction,
	})
}

func getIdentifiersFromFunctionExpression(node *ast.Node, result []identifierInfo) []identifierInfo {
	// Only named function expressions have a name to check.
	// Use the function expression's own Name() rather than GetNameOfDeclaration(),
	// because GetNameOfDeclaration() for anonymous function expressions returns the
	// parent assignment target (e.g., the variable name), which would cause duplicates
	// with the VariableStatement/PropertyAssignment handlers.
	nameNode := node.AsFunctionExpression().Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		declNode: node, // Use the FunctionExpression for modifier detection (async, etc.)
		selector: selectorFunction,
	})
}

func getIdentifiersFromParameter(node *ast.Node, result []identifierInfo) []identifierInfo {
	// Parameters of type-only signatures (function types, call/construct/method
	// signatures, index signatures, ...) create no binding, so real
	// @typescript-eslint/naming-convention never validates their names. Only
	// parameters of an actual function-like declaration/expression are checked.
	if !ast.IsFunctionLikeDeclaration(node.Parent) {
		return result
	}

	// Check if this is a parameter property (constructor parameter with access modifier, readonly, or override)
	isParamProp := ast.IsParameterPropertyDeclaration(node, node.Parent)

	nameNode := node.Name()
	if nameNode == nil {
		return result
	}

	sel := selectorParameter
	if isParamProp {
		sel = selectorParameterProperty
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		return append(result, identifierInfo{
			name:     nameNode.AsIdentifier().Text,
			node:     nameNode,
			selector: sel,
		})
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		start := len(result)
		result = getIdentifiersFromBindingPattern(nameNode, result)
		for i := start; i < len(result); i++ {
			result[i].selector = sel
		}
		return result
	}
	return result
}

func getIdentifiersFromClassDeclaration(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorClass,
	})
}

func getIdentifiersFromClassExpression(node *ast.Node, result []identifierInfo) []identifierInfo {
	// Only named class expressions have a name to check.
	// Use the class expression's own Name() rather than GetNameOfDeclaration(),
	// because GetNameOfDeclaration() for anonymous class expressions returns the
	// parent assignment target (e.g., the variable name), which would incorrectly
	// apply the `class` selector to variable-inferred names such as
	// `const clz = class extends Base {}` (real @typescript-eslint/naming-convention
	// never flags this — it only validates the class's own identifier, if any).
	nameNode := node.AsClassExpression().Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorClass,
	})
}

func getIdentifiersFromInterfaceDeclaration(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorInterface,
	})
}

func getIdentifiersFromTypeAliasDeclaration(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorTypeAlias,
	})
}

func getIdentifiersFromEnumDeclaration(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorEnum,
	})
}

func getIdentifiersFromEnumMember(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: selectorEnumMember,
	})
}

func getIdentifiersFromTypeParameter(node *ast.Node, result []identifierInfo) []identifierInfo {
	// Only check type parameters inside declaration containers
	// (e.g., `function fn<T>`, `class C<T>`, `interface I<T>`, `type T<U>`).
	// Skip type parameters in mapped types (`{ [K in T]: V }`) and infer
	// types (`infer U`), matching the original ESLint rule's
	// `TSTypeParameterDeclaration > TSTypeParameter` selector.
	if node.Parent != nil && (node.Parent.Kind == ast.KindMappedType || node.Parent.Kind == ast.KindInferType) {
		return result
	}

	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorTypeParameter,
	})
}

func getIdentifiersFromPropertyDeclaration(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	// Check if this is an accessor property
	flags := ast.GetCombinedModifierFlags(node)
	if flags&ast.ModifierFlagsAccessor != 0 {
		return append(result, identifierInfo{
			name:     name,
			node:     nameNode,
			selector: selectorAutoAccessor,
		})
	}

	// Check if the property has a function value (making it a method-like property)
	propDecl := node.AsPropertyDeclaration()
	if propDecl.Initializer != nil {
		if ast.IsArrowFunction(ast.SkipParentheses(propDecl.Initializer)) || ast.IsFunctionExpression(ast.SkipParentheses(propDecl.Initializer)) {
			return append(result, identifierInfo{
				name:     name,
				node:     nameNode,
				selector: selectorClassMethod,
			})
		}
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: selectorClassProperty,
	})
}

func getIdentifiersFromMethodDeclaration(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	// Determine if this is class method or object literal method
	parent := node.Parent
	sel := selectorClassMethod
	if parent != nil && parent.Kind == ast.KindObjectLiteralExpression {
		sel = selectorObjectLiteralMethod
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: sel,
	})
}

func getIdentifiersFromAccessorDeclaration(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: selectorClassicAccessor,
	})
}

func getIdentifiersFromPropertySignature(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	// Check if the property has a function type annotation (e.g., `method: () => void`)
	// If so, classify as typeMethod instead of typeProperty
	sel := selectorTypeProperty
	propSig := node.AsPropertySignatureDeclaration()
	if propSig.Type != nil && ast.SkipTypeParentheses(propSig.Type).Kind == ast.KindFunctionType {
		sel = selectorTypeMethod
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: sel,
	})
}

func getIdentifiersFromMethodSignature(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: selectorTypeMethod,
	})
}

func getIdentifiersFromPropertyAssignment(ctx rule.RuleContext, node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return result
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return result
	}

	propAssignment := node.AsPropertyAssignment()
	if propAssignment.Initializer != nil {
		if ast.IsArrowFunction(ast.SkipParentheses(propAssignment.Initializer)) || ast.IsFunctionExpression(ast.SkipParentheses(propAssignment.Initializer)) {
			return append(result, identifierInfo{
				name:     name,
				node:     nameNode,
				selector: selectorObjectLiteralMethod,
			})
		}
	}

	return append(result, identifierInfo{
		name:     name,
		node:     nameNode,
		selector: selectorObjectLiteralProperty,
	})
}

func getIdentifiersFromShorthandPropertyAssignment(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}

	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorObjectLiteralProperty,
	})
}

func getIdentifiersFromImportClause(node *ast.Node, result []identifierInfo) []identifierInfo {
	// Default import: `import Foo from ...`
	importClause := node.AsImportClause()
	if importClause.Name() != nil && importClause.Name().Kind == ast.KindIdentifier {
		return append(result, identifierInfo{
			name:     importClause.Name().AsIdentifier().Text,
			node:     importClause.Name(),
			selector: selectorImport,
		})
	}
	return result
}

func getIdentifiersFromImportSpecifier(node *ast.Node, result []identifierInfo) []identifierInfo {
	importSpec := node.AsImportSpecifier()

	// The import selector only matches default and namespace imports, NOT named imports.
	// `import { default as foo }` is a default import (propertyName is "default").
	// `import { foo }` is a named import and should NOT be matched by the import selector.
	isDefaultImport := importSpec.PropertyName != nil &&
		(importSpec.PropertyName.Kind == ast.KindStringLiteral ||
			(importSpec.PropertyName.Kind == ast.KindIdentifier && importSpec.PropertyName.AsIdentifier().Text == "default"))
	if !isDefaultImport {
		return result
	}

	localName := importSpec.Name()
	if localName == nil || localName.Kind != ast.KindIdentifier {
		return result
	}

	return append(result, identifierInfo{
		name:     localName.AsIdentifier().Text,
		node:     localName,
		selector: selectorImport,
	})
}

func getIdentifiersFromNamespaceImport(node *ast.Node, result []identifierInfo) []identifierInfo {
	nameNode := node.AsNamespaceImport().Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return result
	}
	return append(result, identifierInfo{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorImport,
	})
}

// ---- Re-export name collection ----

func collectReExportedNames(ctx rule.RuleContext) map[*ast.Symbol]bool {
	exported := make(map[*ast.Symbol]bool)
	if ctx.Refs == nil {
		return exported
	}
	mark := func(name *ast.Node) {
		if name != nil && name.Kind == ast.KindIdentifier {
			if symbol := ctx.Refs.ResolveInFile(name); symbol != nil {
				exported[symbol] = true
			}
		}
	}
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		switch node.Kind {
		case ast.KindExportAssignment:
			assignment := node.AsExportAssignment()
			if !assignment.IsExportEquals {
				mark(ast.SkipParentheses(assignment.Expression))
			}
		case ast.KindExportDeclaration:
			declaration := node.AsExportDeclaration()
			if declaration.ModuleSpecifier == nil && declaration.ExportClause != nil && ast.IsNamedExports(declaration.ExportClause) {
				for _, specifier := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
					spec := specifier.AsExportSpecifier()
					name := spec.Name()
					if spec.PropertyName != nil {
						name = spec.PropertyName
					}
					mark(name)
				}
			}
		case ast.KindSourceFile, ast.KindModuleDeclaration, ast.KindModuleBlock:
			node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
	}
	visit(ctx.SourceFile.AsNode())
	return exported
}

// ---- Main run function ----

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	parsed := getParsedNamingOptions(options)
	selectors := parsed.selectors

	if len(selectors) == 0 {
		return nil
	}

	var reExportedNames map[*ast.Symbol]bool
	if parsed.allModifiers&modifierExported != 0 {
		reExportedNames = collectReExportedNames(ctx)
	}
	var refInfo *referencedInfo
	if parsed.allModifiers&modifierUnused != 0 {
		refInfo = getReferencedInfo(ctx)
	}
	identifierBuffer := make([]identifierInfo, 0, 8)

	handleNode := func(node *ast.Node) {
		identifierBuffer = getIdentifierFromNode(ctx, node, identifierBuffer[:0])
		for _, id := range identifierBuffer {
			neededMods := parsed.neededModifiers[bits.TrailingZeros(uint(id.selector))]
			validateIdentifier(ctx, id, selectors, node, reExportedNames, refInfo, neededMods)
		}
	}

	return rule.RuleListeners{
		ast.KindVariableStatement:           handleNode,
		ast.KindForOfStatement:              handleNode,
		ast.KindForInStatement:              handleNode,
		ast.KindForStatement:                handleNode,
		ast.KindFunctionDeclaration:         handleNode,
		ast.KindFunctionExpression:          handleNode,
		ast.KindParameter:                   handleNode,
		ast.KindClassDeclaration:            handleNode,
		ast.KindClassExpression:             handleNode,
		ast.KindInterfaceDeclaration:        handleNode,
		ast.KindTypeAliasDeclaration:        handleNode,
		ast.KindEnumDeclaration:             handleNode,
		ast.KindEnumMember:                  handleNode,
		ast.KindTypeParameter:               handleNode,
		ast.KindPropertyDeclaration:         handleNode,
		ast.KindMethodDeclaration:           handleNode,
		ast.KindGetAccessor:                 handleNode,
		ast.KindSetAccessor:                 handleNode,
		ast.KindPropertySignature:           handleNode,
		ast.KindMethodSignature:             handleNode,
		ast.KindPropertyAssignment:          handleNode,
		ast.KindShorthandPropertyAssignment: handleNode,
		ast.KindImportClause:                handleNode,
		ast.KindImportSpecifier:             handleNode,
		ast.KindNamespaceImport:             handleNode,
	}
}

func validateIdentifier(
	ctx rule.RuleContext,
	id identifierInfo,
	selectors []normalizedSelector,
	originalNode *ast.Node,
	reExportedNames map[*ast.Symbol]bool,
	referencedInfo *referencedInfo,
	neededMods modifierKind,
) {
	// Use the declaration node for modifier detection if available
	modNode := originalNode
	if id.declNode != nil {
		modNode = id.declNode
	}

	var idMods modifierKind
	modsReady := neededMods == 0
	for _, sel := range selectors {
		// Check if this selector matches the identifier's selector kind
		if id.selector&sel.selector == 0 {
			continue
		}

		// Preserve the cheap modifier gate before configured regexes when the
		// selector actually asks for modifiers. Otherwise defer modifier work
		// until a selector has passed its filter and type predicates.
		if sel.modifiers != 0 {
			if !modsReady {
				idMods = getModifiers(ctx, modNode, id.node, id.selector, id.name, reExportedNames, id.destructured, referencedInfo, neededMods)
				modsReady = true
			}
			if (idMods & sel.modifiers) != sel.modifiers {
				continue
			}
		}

		// Check filter match — if filter doesn't match the name, skip this
		// selector entirely so the next one in specificity order can match.
		if sel.filter != nil {
			matches, err := sel.filter.regex.TestOrError(id.name)
			if err != nil || sel.filter.match != matches {
				continue
			}
		}

		// Check type match
		if sel.typesConfigured && selectorSupportsTypes(id.selector) {
			if !isCorrectType(ctx.TypeChecker, id.node, sel.types) {
				continue
			}
		}

		if !modsReady {
			idMods = getModifiers(ctx, modNode, id.node, id.selector, id.name, reExportedNames, id.destructured, referencedInfo, neededMods)
		}

		// This selector matches - validate the name
		result := validate(id.name, sel, idMods, id.selector)
		if !result.valid && result.message != nil {
			ctx.ReportRange(utils.GetESTreeBindingIdentifierRange(ctx.SourceFile, id.node), *result.message)
		}

		// First matching selector wins (most specific first due to sorting)
		return
	}
}
