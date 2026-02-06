package naming_convention

import (
	"fmt"
	"math/bits"
	"regexp"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NamingConventionRule is the exported rule for registration.
var NamingConventionRule = rule.CreateRule(rule.Rule{
	Name: "naming-convention",
	Run:  run,
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

var formatNames = map[predefinedFormat]string{
	formatCamelCase:        "camelCase",
	formatStrictCamelCase:  "strictCamelCase",
	formatPascalCase:       "PascalCase",
	formatStrictPascalCase: "StrictPascalCase",
	formatSnakeCase:        "snake_case",
	formatUpperCase:        "UPPER_CASE",
}

func parseFormatName(s string) (predefinedFormat, bool) {
	for k, v := range formatNames {
		if v == s {
			return k, true
		}
	}
	return 0, false
}

type underscoreOption int

const (
	underscoreForbid underscoreOption = iota
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
	selectorClassProperty
	selectorObjectLiteralProperty
	selectorTypeProperty
	selectorParameterProperty
	selectorEnumMember
	selectorClassMethod
	selectorObjectLiteralMethod
	selectorTypeMethod
	selectorClassicAccessor
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
	selectorTypeLike = selectorClass | selectorInterface | selectorTypeAlias | selectorEnum | selectorTypeParameter
	selectorMethod   = selectorClassMethod | selectorObjectLiteralMethod | selectorTypeMethod
	selectorProperty = selectorClassProperty | selectorObjectLiteralProperty | selectorTypeProperty
	selectorAccessor = selectorClassicAccessor | selectorAutoAccessor
	selectorDefault  = selectorVariableLike | selectorMemberLike | selectorTypeLike | selectorImport
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
	regex *regexp.Regexp
	match bool
}

type normalizedSelector struct {
	selector          selectorKind
	modifiers         modifierKind
	types             typeModifierKind
	filter            *matchRegex
	format            []predefinedFormat // nil means "no format check" (format: null)
	formatNull        bool
	custom            *matchRegex
	leadingUnderscore underscoreOption
	trailingUnderscore underscoreOption
	prefix            []string
	suffix            []string
	modifierWeight    int
}

// ---- Format checking functions ----
// These use regex patterns matching the official typescript-eslint implementation,
// with an additional consecutive-uppercase check for strict variants.

var (
	reCamelCase    = regexp.MustCompile(`^[a-z][\da-zA-Z]*$`)
	rePascalCase   = regexp.MustCompile(`^[A-Z][\da-zA-Z]*$`)
	reUpperCase    = regexp.MustCompile(`^[A-Z][\dA-Z_]*$`)
	reIdentifier   = regexp.MustCompile(`^[\p{L}_$][\p{L}\p{N}_$]*$`)
	reSnakeCase  = regexp.MustCompile(`^[a-z][\da-z_]*$`)
)

func checkCamelCase(name string) bool {
	return reCamelCase.MatchString(name)
}

func checkStrictCamelCase(name string) bool {
	if !reCamelCase.MatchString(name) {
		return false
	}
	return !hasConsecutiveUppercase(name, 0)
}

func checkPascalCase(name string) bool {
	return rePascalCase.MatchString(name)
}

func checkStrictPascalCase(name string) bool {
	if !rePascalCase.MatchString(name) {
		return false
	}
	// Skip the first character for PascalCase consecutive check
	return !hasConsecutiveUppercase(name, 1)
}

func checkSnakeCase(name string) bool {
	return reSnakeCase.MatchString(name)
}

func checkUpperCase(name string) bool {
	return reUpperCase.MatchString(name)
}

// hasConsecutiveUppercase checks if there are two or more consecutive uppercase
// ASCII letters starting from the given index.
func hasConsecutiveUppercase(name string, startIdx int) bool {
	for i := startIdx; i < len(name)-1; i++ {
		if name[i] >= 'A' && name[i] <= 'Z' && name[i+1] >= 'A' && name[i+1] <= 'Z' {
			return true
		}
	}
	return false
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

func parseOptions(rawOpts any) []normalizedSelector {
	if rawOpts == nil {
		return getDefaultConfig()
	}

	var optsList []interface{}
	if arr, ok := rawOpts.([]interface{}); ok {
		optsList = arr
	} else {
		return getDefaultConfig()
	}

	if len(optsList) == 0 {
		return getDefaultConfig()
	}

	var selectors []normalizedSelector
	for _, opt := range optsList {
		optMap, ok := opt.(map[string]interface{})
		if !ok {
			continue
		}
		selectors = append(selectors, parseOneSelector(optMap)...)
	}

	// Sort by specificity (more specific first)
	sort.SliceStable(selectors, func(i, j int) bool {
		return selectors[i].modifierWeight > selectors[j].modifierWeight
	})

	return selectors
}

func getDefaultConfig() []normalizedSelector {
	return parseOptions([]interface{}{
		map[string]interface{}{
			"selector":          "default",
			"format":            []interface{}{"camelCase"},
			"leadingUnderscore": "allow",
			"trailingUnderscore": "allow",
		},
		map[string]interface{}{
			"selector": "import",
			"format":   []interface{}{"camelCase", "PascalCase"},
		},
		map[string]interface{}{
			"selector":          "variable",
			"format":            []interface{}{"camelCase", "UPPER_CASE"},
			"leadingUnderscore": "allow",
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

	// Parse underscores
	leadingUnderscore := underscoreAllow
	if v, ok := optMap["leadingUnderscore"].(string); ok {
		leadingUnderscore = parseUnderscoreOption(v)
	} else if _, exists := optMap["leadingUnderscore"]; !exists {
		leadingUnderscore = underscoreForbid
	}

	trailingUnderscore := underscoreAllow
	if v, ok := optMap["trailingUnderscore"].(string); ok {
		trailingUnderscore = parseUnderscoreOption(v)
	} else if _, exists := optMap["trailingUnderscore"]; !exists {
		trailingUnderscore = underscoreForbid
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
		weight := calculateWeight(mods, selectorTypes, filter, sk)
		result = append(result, normalizedSelector{
			selector:          sk,
			modifiers:         mods,
			types:             selectorTypes,
			filter:            filter,
			format:            formats,
			formatNull:        formatNull,
			custom:            custom,
			leadingUnderscore: leadingUnderscore,
			trailingUnderscore: trailingUnderscore,
			prefix:            prefix,
			suffix:            suffix,
			modifierWeight:    weight,
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
		re, err := regexp.Compile(v)
		if err != nil {
			return nil
		}
		return &matchRegex{regex: re, match: true}
	case map[string]interface{}:
		regexStr, _ := v["regex"].(string)
		matchVal := true
		if m, ok := v["match"].(bool); ok {
			matchVal = m
		}
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil
		}
		return &matchRegex{regex: re, match: matchVal}
	}
	return nil
}

func calculateWeight(mods modifierKind, types typeModifierKind, filter *matchRegex, sk selectorKind) int {
	weight := 0

	// Individual selector (bitCount=1) is most specific
	pc := bitCount(int(sk))
	if pc == 1 {
		weight |= 1 << 27
	} else {
		// Smaller group = more specific. Invert so fewer bits = higher weight.
		// Max possible bit count for selectorDefault is ~19, so 20-pc gives us 1-20 range
		weight |= (20 - pc) << 22
	}

	if mods != 0 {
		weight |= 1 << 28
		weight += bitCount(int(mods))
	}
	if types != 0 {
		weight |= 1 << 29
		weight += bitCount(int(types))
	}
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
		fmtNames = append(fmtNames, formatNames[f])
	}
	return rule.RuleMessage{
		Id:          "doesNotMatchFormat",
		Description: fmt.Sprintf("%s name `%s` must match one of the following formats: %s", typeName, name, strings.Join(fmtNames, ", ")),
	}
}

func doesNotMatchFormatTrimmedMessage(typeName, name, processedName string, formats []predefinedFormat) rule.RuleMessage {
	var fmtNames []string
	for _, f := range formats {
		fmtNames = append(fmtNames, formatNames[f])
	}
	return rule.RuleMessage{
		Id:          "doesNotMatchFormatTrimmed",
		Description: fmt.Sprintf("%s name `%s` must match one of the following formats: %s. The format was tested against `%s`", typeName, name, strings.Join(fmtNames, ", "), processedName),
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
		Description: fmt.Sprintf("%s name `%s` must %s the RegExp: %s", typeName, name, matchStr, regex),
	}
}

func unexpectedUnderscoreMessage(typeName, name, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedUnderscore",
		Description: fmt.Sprintf("%s name `%s` must not have a %s underscore.", typeName, name, position),
	}
}

func missingUnderscoreMessage(typeName, name string, count int, position string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingUnderscore",
		Description: fmt.Sprintf("%s name `%s` must have %d %s underscore(s).", typeName, name, count, position),
	}
}

// ---- Validation logic ----

type validationResult struct {
	valid   bool
	message *rule.RuleMessage
}

func validate(name string, sel normalizedSelector, idMods modifierKind) validationResult {
	typeName := selectorTypeToMessageString(sel.selector)

	// 1. Check filter
	if sel.filter != nil {
		matches := sel.filter.regex.MatchString(name)
		if sel.filter.match != matches {
			// Filter does not match - skip this selector
			return validationResult{valid: true}
		}
	}

	// 2. (requiresQuotes + formatNull is handled by step 8's formatNull check)

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
		matches := sel.custom.regex.MatchString(processedName)
		if sel.custom.match != matches {
			msg := satisfyCustomMessage(typeName, name, sel.custom.match, sel.custom.regex.String())
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
	if types == 0 || ch == nil {
		return true
	}

	t := ch.GetTypeAtLocation(node)
	if t == nil {
		return false
	}

	return checkTypeMatch(ch, t, types)
}

func checkTypeMatch(ch *checker.Checker, t *checker.Type, types typeModifierKind) bool {
	if types&typeModBoolean != 0 {
		if isBooleanLikeType(t) {
			return true
		}
	}
	if types&typeModString != 0 {
		if isStringLikeType(t) {
			return true
		}
	}
	if types&typeModNumber != 0 {
		if isNumberLikeType(t) {
			return true
		}
	}
	if types&typeModFunction != 0 {
		if isFunctionLikeType(ch, t) {
			return true
		}
	}
	if types&typeModArray != 0 {
		if isArrayLikeType(ch, t) {
			return true
		}
	}
	return false
}

func isBooleanLikeType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range utils.UnionTypeParts(t) {
			partFlags := checker.Type_flags(part)
			if partFlags&(checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
				return false
			}
		}
		return true
	}
	return false
}

func isStringLikeType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsStringLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range utils.UnionTypeParts(t) {
			partFlags := checker.Type_flags(part)
			if partFlags&(checker.TypeFlagsStringLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
				return false
			}
		}
		return true
	}
	return false
}

func isNumberLikeType(t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsNumberLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range utils.UnionTypeParts(t) {
			partFlags := checker.Type_flags(part)
			if partFlags&(checker.TypeFlagsNumberLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
				return false
			}
		}
		return true
	}
	return false
}

func isFunctionLikeType(ch *checker.Checker, t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range utils.UnionTypeParts(t) {
			if !isFunctionLikeType(ch, part) {
				return false
			}
		}
		return true
	}
	sigs := checker.Checker_getSignaturesOfType(ch, t, checker.SignatureKindCall)
	return len(sigs) > 0
}

func isArrayLikeType(ch *checker.Checker, t *checker.Type) bool {
	flags := checker.Type_flags(t)
	if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return true
	}
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range utils.UnionTypeParts(t) {
			if !isArrayLikeType(ch, part) {
				return false
			}
		}
		return true
	}
	return checker.Checker_isArrayType(ch, t) || checker.Checker_isArrayOrTupleType(ch, t)
}

// ---- Modifier detection helpers ----

func getModifiers(ctx rule.RuleContext, node *ast.Node, sel selectorKind, name string, reExportedNames map[string]bool, destructured bool, referencedSymbols map[*ast.Symbol]bool) modifierKind {
	var mods modifierKind

	flags := ast.GetCombinedModifierFlags(node)

	// Access modifiers
	if flags&ast.ModifierFlagsPublic != 0 {
		mods |= modifierPublic
	}
	if flags&ast.ModifierFlagsProtected != 0 {
		mods |= modifierProtected
	}
	if flags&ast.ModifierFlagsPrivate != 0 {
		mods |= modifierPrivate
	}

	// Other modifiers
	if flags&ast.ModifierFlagsStatic != 0 {
		mods |= modifierStatic
	}
	if flags&ast.ModifierFlagsReadonly != 0 {
		mods |= modifierReadonly
	}
	if flags&ast.ModifierFlagsAbstract != 0 {
		mods |= modifierAbstract
	}
	if flags&ast.ModifierFlagsAsync != 0 {
		mods |= modifierAsync
	}

	// Detect async modifier from function expression/arrow function initializer.
	// For `AsyncBar = async () => {}` (class property or object literal property)
	// and `const asyncFoo = async () => {}` (variable), the async flag is on
	// the initializer, not on the declaration itself.
	if mods&modifierAsync == 0 {
		switch node.Kind {
		case ast.KindPropertyDeclaration:
			if init := node.AsPropertyDeclaration().Initializer; init != nil {
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					if ast.GetCombinedModifierFlags(init)&ast.ModifierFlagsAsync != 0 {
						mods |= modifierAsync
					}
				}
			}
		case ast.KindPropertyAssignment:
			if init := node.AsPropertyAssignment().Initializer; init != nil {
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					if ast.GetCombinedModifierFlags(init)&ast.ModifierFlagsAsync != 0 {
						mods |= modifierAsync
					}
				}
			}
		case ast.KindVariableDeclaration:
			if init := node.AsVariableDeclaration().Initializer; init != nil {
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					if ast.GetCombinedModifierFlags(init)&ast.ModifierFlagsAsync != 0 {
						mods |= modifierAsync
					}
				}
			}
		}
	}

	if flags&ast.ModifierFlagsConst != 0 {
		mods |= modifierConst
	}
	if flags&ast.ModifierFlagsOverride != 0 {
		mods |= modifierOverride
	}
	// Accessor keyword check is handled by the selector system, not as a modifier

	// Hash private (ECMAScript private)
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
		mods |= modifierHashPrivate
	}

	// Default accessibility: if no access modifier on a class member, it's public
	if sel&selectorMemberLike != 0 && mods&(modifierPublic|modifierProtected|modifierPrivate|modifierHashPrivate) == 0 {
		mods |= modifierPublic
	}

	// Export check (direct export or re-export via export { ... })
	if isExported(node) || reExportedNames[name] {
		mods |= modifierExported
	}

	// Global check
	if isGlobalScope(node) {
		mods |= modifierGlobal
	}

	// Destructured check (per-identifier, set during extraction)
	if destructured {
		mods |= modifierDestructured
	}

	// Unused check - exported symbols are never considered unused
	if mods&modifierExported == 0 && isUnused(ctx, node, referencedSymbols) {
		mods |= modifierUnused
	}

	// Const check for variables
	if sel&selectorVariable != 0 {
		if isConstVariable(node) {
			mods |= modifierConst
		}
	}

	// Const check for enum
	if sel&selectorEnum != 0 {
		if flags&ast.ModifierFlagsConst != 0 {
			mods |= modifierConst
		}
	}

	// requiresQuotes check for members
	if sel&(selectorProperty|selectorMethod|selectorAccessor|selectorEnumMember) != 0 {
		if requiresQuoting(node) {
			mods |= modifierRequiresQuotes
		}
	}

	// Import modifier detection (default vs named vs namespace)
	if sel&selectorImport != 0 {
		switch node.Kind {
		case ast.KindImportClause:
			// `import Foo from ...` is a default import
			mods |= modifierDefault
		case ast.KindNamespaceImport:
			// `import * as Foo from ...` is a namespace import
			mods |= modifierNamespace
		case ast.KindImportSpecifier:
			// `import { default as Foo } from ...` is also a default import
			importSpec := node.AsImportSpecifier()
			if importSpec.PropertyName != nil && importSpec.PropertyName.Kind == ast.KindIdentifier {
				if importSpec.PropertyName.AsIdentifier().Text == "default" {
					mods |= modifierDefault
				}
			}
		}
	}

	return mods
}

func isExported(node *ast.Node) bool {
	// Check if node itself has export keyword
	flags := ast.GetCombinedModifierFlags(node)
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

	// Check if parent is an export statement
	if ast.IsFunctionDeclaration(node) || ast.IsClassDeclaration(node) ||
		node.Kind == ast.KindInterfaceDeclaration || node.Kind == ast.KindTypeAliasDeclaration ||
		node.Kind == ast.KindEnumDeclaration || node.Kind == ast.KindModuleDeclaration {
		parentFlags := ast.GetCombinedModifierFlags(node)
		if parentFlags&ast.ModifierFlagsExport != 0 {
			return true
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
	return node.Kind == ast.KindSourceFile || node.Kind == ast.KindModuleBlock
}

// collectReferencedSymbols walks the source file and collects all symbols that are referenced
// (i.e., used in a non-declaration context). This is used to detect the "unused" modifier.
func collectReferencedSymbols(ctx rule.RuleContext) map[*ast.Symbol]bool {
	if ctx.TypeChecker == nil {
		return nil
	}

	referenced := make(map[*ast.Symbol]bool)
	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier {
			// Skip if this identifier is the name of a declaration
			parent := node.Parent
			if parent != nil && ast.GetNameOfDeclaration(parent) == node {
				// This is a declaration name, not a reference
			} else {
				sym := ctx.TypeChecker.GetSymbolAtLocation(node)
				if sym != nil {
					referenced[sym] = true
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	walk(&ctx.SourceFile.Node)
	return referenced
}

func isUnused(ctx rule.RuleContext, node *ast.Node, referencedSymbols map[*ast.Symbol]bool) bool {
	if referencedSymbols == nil {
		return false
	}

	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return false
	}

	if ctx.TypeChecker == nil {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if symbol == nil {
		return false
	}

	return !referencedSymbols[symbol]
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

func requiresQuoting(node *ast.Node) bool {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return false
	}

	if nameNode.Kind == ast.KindComputedPropertyName {
		return true
	}

	if nameNode.Kind == ast.KindStringLiteral {
		// Only requires quoting if the string value is not a valid JS identifier
		text := nameNode.AsStringLiteral().Text
		return !isIdentifierText(text)
	}

	if nameNode.Kind == ast.KindNumericLiteral {
		// Numeric literals are never valid identifiers
		return true
	}

	return false
}

// isIdentifierText checks if a string is a valid JavaScript/TypeScript identifier.
func isIdentifierText(name string) bool {
	if len(name) == 0 {
		return false
	}
	return reIdentifier.MatchString(name)
}

// ---- Name extraction helper ----

// getNameFromNode extracts the name string from a member name node, handling string/numeric literals
// that GetNameFromMember doesn't handle directly.
func getNameFromNode(ctx rule.RuleContext, nameNode *ast.Node) (string, bool) {
	switch nameNode.Kind {
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

func getIdentifierFromNode(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	switch node.Kind {
	case ast.KindVariableStatement:
		return getIdentifiersFromVariableStatement(ctx, node)
	case ast.KindFunctionDeclaration:
		return getIdentifiersFromFunctionDeclaration(node)
	case ast.KindFunctionExpression:
		return getIdentifiersFromFunctionExpression(node)
	case ast.KindParameter:
		return getIdentifiersFromParameter(ctx, node)
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return getIdentifiersFromClassDeclaration(node)
	case ast.KindInterfaceDeclaration:
		return getIdentifiersFromInterfaceDeclaration(node)
	case ast.KindTypeAliasDeclaration:
		return getIdentifiersFromTypeAliasDeclaration(node)
	case ast.KindEnumDeclaration:
		return getIdentifiersFromEnumDeclaration(node)
	case ast.KindEnumMember:
		return getIdentifiersFromEnumMember(ctx, node)
	case ast.KindTypeParameter:
		return getIdentifiersFromTypeParameter(node)
	case ast.KindPropertyDeclaration:
		return getIdentifiersFromPropertyDeclaration(ctx, node)
	case ast.KindMethodDeclaration:
		return getIdentifiersFromMethodDeclaration(ctx, node)
	case ast.KindGetAccessor, ast.KindSetAccessor:
		return getIdentifiersFromAccessorDeclaration(ctx, node)
	case ast.KindPropertySignature:
		return getIdentifiersFromPropertySignature(ctx, node)
	case ast.KindMethodSignature:
		return getIdentifiersFromMethodSignature(ctx, node)
	case ast.KindPropertyAssignment:
		return getIdentifiersFromPropertyAssignment(ctx, node)
	case ast.KindShorthandPropertyAssignment:
		return getIdentifiersFromShorthandPropertyAssignment(ctx, node)
	case ast.KindImportClause:
		return getIdentifiersFromImportClause(node)
	case ast.KindImportSpecifier:
		return getIdentifiersFromImportSpecifier(node)
	case ast.KindNamespaceImport:
		return getIdentifiersFromNamespaceImport(node)
	default:
		return nil
	}
}

func getIdentifiersFromVariableStatement(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	varStmt := node.AsVariableStatement()
	if varStmt.DeclarationList == nil {
		return nil
	}

	declList := varStmt.DeclarationList.AsVariableDeclarationList()
	var result []identifierInfo

	for _, decl := range declList.Declarations.Nodes {
		nameNode := decl.Name()
		if nameNode == nil {
			continue
		}

		switch nameNode.Kind {
		case ast.KindIdentifier:
			// Check if this is a function expression assignment
			sel := selectorVariable
			if decl.AsVariableDeclaration().Initializer != nil {
				init := decl.AsVariableDeclaration().Initializer
				if ast.IsArrowFunction(init) || ast.IsFunctionExpression(init) {
					sel = selectorFunction
				}
			}
			result = append(result, identifierInfo{
				name:     nameNode.AsIdentifier().Text,
				node:     nameNode,
				declNode: decl,
				selector: sel,
			})
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			ids := getIdentifiersFromBindingPattern(nameNode)
			for i := range ids {
				ids[i].declNode = decl
			}
			result = append(result, ids...)
		}
	}
	return result
}

func getIdentifiersFromBindingPattern(pattern *ast.Node) []identifierInfo {
	var result []identifierInfo

	elements := ast.GetElementsOfBindingOrAssignmentPattern(pattern)
	for _, elem := range elements {
		if elem.Kind == ast.KindBindingElement {
			nameNode := elem.Name()
			if nameNode == nil {
				continue
			}
			// A binding element is "destructured" only if it doesn't have a property name (rename).
			// e.g., `{ foo }` is destructured, `{ bar: foo }` is NOT destructured (renamed).
			isDestructuredBinding := elem.AsBindingElement().PropertyName == nil
			switch nameNode.Kind {
			case ast.KindIdentifier:
				result = append(result, identifierInfo{
					name:         nameNode.AsIdentifier().Text,
					node:         nameNode,
					selector:     selectorVariable,
					destructured: isDestructuredBinding,
				})
			case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
				result = append(result, getIdentifiersFromBindingPattern(nameNode)...)
			}
		}
	}

	return result
}

func getIdentifiersFromFunctionDeclaration(node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorFunction,
	}}
}

func getIdentifiersFromFunctionExpression(node *ast.Node) []identifierInfo {
	// Only named function expressions have a name to check
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		declNode: node, // Use the FunctionExpression for modifier detection (async, etc.)
		selector: selectorFunction,
	}}
}

func getIdentifiersFromParameter(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	// Check if this is a parameter property (constructor parameter with access modifier or readonly)
	flags := ast.GetCombinedModifierFlags(node)
	isParamProp := flags&(ast.ModifierFlagsPublic|ast.ModifierFlagsProtected|ast.ModifierFlagsPrivate|ast.ModifierFlagsReadonly) != 0

	nameNode := node.Name()
	if nameNode == nil {
		return nil
	}

	sel := selectorParameter
	if isParamProp {
		sel = selectorParameterProperty
	}

	switch nameNode.Kind {
	case ast.KindIdentifier:
		return []identifierInfo{{
			name:     nameNode.AsIdentifier().Text,
			node:     nameNode,
			selector: sel,
		}}
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		ids := getIdentifiersFromBindingPattern(nameNode)
		for i := range ids {
			ids[i].selector = sel
		}
		return ids
	}
	return nil
}

func getIdentifiersFromClassDeclaration(node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorClass,
	}}
}

func getIdentifiersFromInterfaceDeclaration(node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorInterface,
	}}
}

func getIdentifiersFromTypeAliasDeclaration(node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorTypeAlias,
	}}
}

func getIdentifiersFromEnumDeclaration(node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorEnum,
	}}
}

func getIdentifiersFromEnumMember(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: selectorEnumMember,
	}}
}

func getIdentifiersFromTypeParameter(node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorTypeParameter,
	}}
}

func getIdentifiersFromPropertyDeclaration(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	// Check if this is an accessor property
	flags := ast.GetCombinedModifierFlags(node)
	if flags&ast.ModifierFlagsAccessor != 0 {
		return []identifierInfo{{
			name:     name,
			node:     nameNode,
			selector: selectorAutoAccessor,
		}}
	}

	// Check if the property has a function value (making it a method-like property)
	propDecl := node.AsPropertyDeclaration()
	if propDecl.Initializer != nil {
		if ast.IsArrowFunction(propDecl.Initializer) || ast.IsFunctionExpression(propDecl.Initializer) {
			return []identifierInfo{{
				name:     name,
				node:     nameNode,
				selector: selectorClassMethod,
			}}
		}
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: selectorClassProperty,
	}}
}

func getIdentifiersFromMethodDeclaration(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	// Determine if this is class method or object literal method
	parent := node.Parent
	sel := selectorClassMethod
	if parent != nil && parent.Kind == ast.KindObjectLiteralExpression {
		sel = selectorObjectLiteralMethod
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: sel,
	}}
}

func getIdentifiersFromAccessorDeclaration(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: selectorClassicAccessor,
	}}
}

func getIdentifiersFromPropertySignature(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	// Check if the property has a function type annotation (e.g., `method: () => void`)
	// If so, classify as typeMethod instead of typeProperty
	sel := selectorTypeProperty
	propSig := node.AsPropertySignatureDeclaration()
	if propSig.Type != nil && propSig.Type.Kind == ast.KindFunctionType {
		sel = selectorTypeMethod
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: sel,
	}}
}

func getIdentifiersFromMethodSignature(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: selectorTypeMethod,
	}}
}

func getIdentifiersFromPropertyAssignment(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil {
		return nil
	}

	name, ok := getNameFromNode(ctx, nameNode)
	if !ok {
		return nil
	}

	propAssignment := node.AsPropertyAssignment()
	if propAssignment.Initializer != nil {
		if ast.IsArrowFunction(propAssignment.Initializer) || ast.IsFunctionExpression(propAssignment.Initializer) {
			return []identifierInfo{{
				name:     name,
				node:     nameNode,
				selector: selectorObjectLiteralMethod,
			}}
		}
	}

	return []identifierInfo{{
		name:     name,
		node:     nameNode,
		selector: selectorObjectLiteralProperty,
	}}
}

func getIdentifiersFromShorthandPropertyAssignment(ctx rule.RuleContext, node *ast.Node) []identifierInfo {
	nameNode := ast.GetNameOfDeclaration(node)
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}

	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorObjectLiteralProperty,
	}}
}

func getIdentifiersFromImportClause(node *ast.Node) []identifierInfo {
	// Default import: `import Foo from ...`
	importClause := node.AsImportClause()
	if importClause.Name() != nil && importClause.Name().Kind == ast.KindIdentifier {
		return []identifierInfo{{
			name:     importClause.Name().AsIdentifier().Text,
			node:     importClause.Name(),
			selector: selectorImport,
		}}
	}
	return nil
}

func getIdentifiersFromImportSpecifier(node *ast.Node) []identifierInfo {
	importSpec := node.AsImportSpecifier()

	// The import selector only matches default and namespace imports, NOT named imports.
	// `import { default as foo }` is a default import (propertyName is "default").
	// `import { foo }` is a named import and should NOT be matched by the import selector.
	isDefaultImport := importSpec.PropertyName != nil &&
		importSpec.PropertyName.Kind == ast.KindIdentifier &&
		importSpec.PropertyName.AsIdentifier().Text == "default"
	if !isDefaultImport {
		return nil
	}

	localName := importSpec.Name()
	if localName == nil || localName.Kind != ast.KindIdentifier {
		return nil
	}

	return []identifierInfo{{
		name:     localName.AsIdentifier().Text,
		node:     localName,
		selector: selectorImport,
	}}
}

func getIdentifiersFromNamespaceImport(node *ast.Node) []identifierInfo {
	nameNode := node.AsNamespaceImport().Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	return []identifierInfo{{
		name:     nameNode.AsIdentifier().Text,
		node:     nameNode,
		selector: selectorImport,
	}}
}

// ---- Re-export name collection ----

func collectReExportedNames(ctx rule.RuleContext) map[string]bool {
	reExported := make(map[string]bool)
	if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil {
		return reExported
	}
	for _, stmt := range ctx.SourceFile.Statements.Nodes {
		if stmt.Kind != ast.KindExportDeclaration {
			continue
		}
		ed := stmt.AsExportDeclaration()
		if ed == nil || ed.ExportClause == nil {
			continue
		}
		// Only handle named exports without a module specifier (re-exports of local names)
		if ed.ModuleSpecifier != nil {
			continue
		}
		if ed.ExportClause.Kind != ast.KindNamedExports {
			continue
		}
		namedExports := ed.ExportClause.AsNamedExports()
		if namedExports == nil || namedExports.Elements == nil {
			continue
		}
		for _, element := range namedExports.Elements.Nodes {
			spec := element.AsExportSpecifier()
			if spec == nil {
				continue
			}
			// The local name is PropertyName if aliased, otherwise Name
			var localName *ast.Node
			if spec.PropertyName != nil {
				localName = spec.PropertyName
			} else {
				localName = spec.Name()
			}
			if localName != nil && localName.Kind == ast.KindIdentifier {
				reExported[localName.AsIdentifier().Text] = true
			}
		}
	}
	return reExported
}

// ---- Main run function ----

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	selectors := parseOptions(options)

	if len(selectors) == 0 {
		return nil
	}

	reExportedNames := collectReExportedNames(ctx)
	referencedSymbols := collectReferencedSymbols(ctx)

	handleNode := func(node *ast.Node) {
		identifiers := getIdentifierFromNode(ctx, node)
		for _, id := range identifiers {
			validateIdentifier(ctx, id, selectors, node, reExportedNames, referencedSymbols)
		}
	}

	return rule.RuleListeners{
		ast.KindVariableStatement:            handleNode,
		ast.KindFunctionDeclaration:          handleNode,
		ast.KindFunctionExpression:           handleNode,
		ast.KindParameter:                    handleNode,
		ast.KindClassDeclaration:             handleNode,
		ast.KindClassExpression:              handleNode,
		ast.KindInterfaceDeclaration:         handleNode,
		ast.KindTypeAliasDeclaration:         handleNode,
		ast.KindEnumDeclaration:              handleNode,
		ast.KindEnumMember:                   handleNode,
		ast.KindTypeParameter:                handleNode,
		ast.KindPropertyDeclaration:          handleNode,
		ast.KindMethodDeclaration:            handleNode,
		ast.KindGetAccessor:                  handleNode,
		ast.KindSetAccessor:                  handleNode,
		ast.KindPropertySignature:            handleNode,
		ast.KindMethodSignature:              handleNode,
		ast.KindPropertyAssignment:           handleNode,
		ast.KindShorthandPropertyAssignment:  handleNode,
		ast.KindImportClause:                 handleNode,
		ast.KindImportSpecifier:              handleNode,
		ast.KindNamespaceImport:              handleNode,
	}
}

func validateIdentifier(ctx rule.RuleContext, id identifierInfo, selectors []normalizedSelector, originalNode *ast.Node, reExportedNames map[string]bool, referencedSymbols map[*ast.Symbol]bool) {
	// Use the declaration node for modifier detection if available
	modNode := originalNode
	if id.declNode != nil {
		modNode = id.declNode
	}

	for _, sel := range selectors {
		// Check if this selector matches the identifier's selector kind
		if id.selector&sel.selector == 0 {
			continue
		}

		// Check modifiers match
		idMods := getModifiers(ctx, modNode, id.selector, id.name, reExportedNames, id.destructured, referencedSymbols)
		if sel.modifiers != 0 && (idMods&sel.modifiers) != sel.modifiers {
			continue
		}

		// Check type match
		if sel.types != 0 {
			if !isCorrectType(ctx.TypeChecker, id.node, sel.types) {
				continue
			}
		}

		// This selector matches - validate the name
		result := validate(id.name, sel, idMods)
		if !result.valid && result.message != nil {
			ctx.ReportNode(id.node, *result.message)
		}

		// First matching selector wins (most specific first due to sorting)
		return
	}
}

