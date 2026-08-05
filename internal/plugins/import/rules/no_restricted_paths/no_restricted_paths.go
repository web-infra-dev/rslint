package no_restricted_paths

import (
	_ "embed"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_restricted_paths.schema.json
var schemaJSON []byte

type zone struct {
	targets   []string
	fromPaths []string
	except    []string
	message   string
}

type ruleOptions struct {
	zones    []zone
	basePath string
}

// pathValidator decides, for one `from` entry of a zone, whether an import is
// restricted and whether the zone's `except` list applies to it.
type pathValidator struct {
	isPathRestricted   func(absoluteImportPath string) bool
	hasValidExceptions bool
	isPathException    func(absoluteImportPath string) bool
	invalidException   rule.RuleMessage
}

// NoRestrictedPathsRule enforces which files may be imported inside a zone.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/no-restricted-paths.js
var NoRestrictedPathsRule = rule.Rule{
	Name:   "import/no-restricted-paths",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		if len(opts.zones) == 0 || ctx.Program == nil {
			return rule.RuleListeners{}
		}

		basePath := resolveBasePath(ctx, opts.basePath)
		// `path.relative` compares case-insensitively only in its win32 flavor,
		// so this follows the platform rather than the host file system: a
		// case-insensitive volume on macOS still gets a case-sensitive compare.
		caseSensitive := runtime.GOOS != "windows"

		currentFilename := import_utils.GetPhysicalFilename(ctx)
		matchingZones := make([]zone, 0, len(opts.zones))
		for _, z := range opts.zones {
			if isMatchingZone(z, basePath, currentFilename, caseSensitive) {
				matchingZones = append(matchingZones, z)
			}
		}
		if len(matchingZones) == 0 {
			return rule.RuleListeners{}
		}

		validators := make([][]pathValidator, len(matchingZones))
		built := make([]bool, len(matchingZones))
		applicable := make([]int, 0, 4)

		return import_utils.VisitModules(func(source *ast.StringLiteralLike, node *ast.Node) {
			absoluteImportPath, ok := import_utils.Resolve(source, ctx)
			if !ok {
				return
			}

			for i := range matchingZones {
				if !built[i] {
					validators[i] = makePathValidators(matchingZones[i].fromPaths, matchingZones[i].except, basePath, caseSensitive)
					built[i] = true
				}

				applicable = applicable[:0]
				for j := range validators[i] {
					if validators[i][j].isPathRestricted(absoluteImportPath) {
						applicable = append(applicable, j)
					}
				}

				for _, j := range applicable {
					if !validators[i][j].hasValidExceptions {
						ctx.ReportNode(source, validators[i][j].invalidException)
					}
				}
				for _, j := range applicable {
					if validators[i][j].hasValidExceptions && !validators[i][j].isPathException(absoluteImportPath) {
						ctx.ReportNode(source, unexpectedPathMessage(source.Text(), matchingZones[i].message))
					}
				}
			}
		}, import_utils.VisitModulesOptions{
			Commonjs: true,
			ESModule: true,
		})
	},
}

// resolveBasePath mirrors upstream's `options.basePath || process.cwd()` fed
// into `path.resolve`: a configured relative `basePath` is made absolute
// against the working directory, so zone paths resolved from it can be
// compared against absolute file names. Rule tests and any other caller
// without a process directory fall back to the Program's own directory.
func resolveBasePath(ctx rule.RuleContext, configured string) string {
	cwd := ctx.Cwd
	if cwd == "" {
		cwd = ctx.Program.Host().GetCurrentDirectory()
	}
	if configured == "" {
		return tspath.NormalizePath(cwd)
	}
	return tspath.GetNormalizedAbsolutePath(configured, cwd)
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}

	opts.basePath, _ = optsMap["basePath"].(string)

	rawZones, _ := optsMap["zones"].([]interface{})
	opts.zones = make([]zone, 0, len(rawZones))
	for _, rawZone := range rawZones {
		zoneMap, ok := rawZone.(map[string]interface{})
		if !ok {
			continue
		}
		z := zone{
			targets:   toStringSlice(zoneMap["target"]),
			fromPaths: toStringSlice(zoneMap["from"]),
			except:    toStringSlice(zoneMap["except"]),
		}
		z.message, _ = zoneMap["message"].(string)
		opts.zones = append(opts.zones, z)
	}

	return opts
}

// toStringSlice normalizes upstream's `[].concat(value)` handling of options
// that accept either a single string or an array of strings.
func toStringSlice(raw any) []string {
	switch typed := raw.(type) {
	case string:
		return []string{typed}
	case []interface{}:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if value, ok := item.(string); ok {
				values = append(values, value)
			}
		}
		return values
	}
	return nil
}

func isMatchingZone(z zone, basePath string, currentFilename string, caseSensitive bool) bool {
	for _, target := range z.targets {
		if isMatchingTargetPath(currentFilename, tspath.ResolvePath(basePath, target), caseSensitive) {
			return true
		}
	}
	return false
}

func isMatchingTargetPath(fileName string, targetPath string, caseSensitive bool) bool {
	if isGlob(targetPath) {
		return utils.MatchGlob(targetPath, fileName)
	}
	return containsPath(fileName, targetPath, caseSensitive)
}

func makePathValidators(fromPaths []string, except []string, basePath string, caseSensitive bool) []pathValidator {
	anyGlob := false
	anyNonGlob := false
	for _, from := range fromPaths {
		if isGlob(from) {
			anyGlob = true
		} else {
			anyNonGlob = true
		}
	}

	if anyGlob && anyNonGlob {
		return []pathValidator{{
			isPathRestricted:   func(string) bool { return true },
			hasValidExceptions: false,
			invalidException: rule.RuleMessage{
				Id:          "mixedGlobAndNonGlob",
				Description: "Restricted path `from` must contain either only glob patterns or none",
			},
		}}
	}

	validators := make([]pathValidator, 0, len(fromPaths))
	for _, from := range fromPaths {
		absoluteFrom := tspath.ResolvePath(basePath, from)
		if anyGlob {
			validators = append(validators, computeGlobPatternPathValidator(absoluteFrom, except))
		} else {
			validators = append(validators, computeAbsolutePathValidator(absoluteFrom, except, caseSensitive))
		}
	}
	return validators
}

// computeGlobPatternPathValidator matches both `from` and `except` as glob
// patterns. Note that upstream matches `except` patterns verbatim rather than
// resolving them against `basePath` first, so only absolute exception patterns
// can ever match a resolved import path.
func computeGlobPatternPathValidator(absoluteFrom string, except []string) pathValidator {
	validator := pathValidator{
		isPathRestricted: func(absoluteImportPath string) bool {
			return utils.MatchGlob(absoluteFrom, absoluteImportPath)
		},
		hasValidExceptions: true,
		invalidException: rule.RuleMessage{
			Id:          "invalidExceptionGlob",
			Description: "Restricted path exceptions must be glob patterns when `from` contains glob patterns",
		},
	}

	for _, exception := range except {
		if !isGlob(exception) {
			validator.hasValidExceptions = false
			return validator
		}
	}

	exceptions := except
	validator.isPathException = func(absoluteImportPath string) bool {
		for _, exception := range exceptions {
			if utils.MatchGlob(exception, absoluteImportPath) {
				return true
			}
		}
		return false
	}
	return validator
}

func computeAbsolutePathValidator(absoluteFrom string, except []string, caseSensitive bool) pathValidator {
	validator := pathValidator{
		isPathRestricted: func(absoluteImportPath string) bool {
			return containsPath(absoluteImportPath, absoluteFrom, caseSensitive)
		},
		hasValidExceptions: true,
		invalidException: rule.RuleMessage{
			Id:          "invalidExceptionPath",
			Description: "Restricted path exceptions must be descendants of the configured `from` path for that zone.",
		},
	}

	absoluteExceptions := make([]string, 0, len(except))
	for _, exception := range except {
		absoluteException := tspath.ResolvePath(absoluteFrom, exception)
		// Upstream classifies the `from`-relative exception path with
		// importType and rejects it when it comes out as "parent", whose regex
		// `^\.\.$|^\.\.[\\/]` matches only a leading `..` component. A name such
		// as `..secret` therefore stays a valid exception, even though
		// containsPath reads the same leftover as escaping `from`.
		if !isPathWithin(absoluteException, absoluteFrom, caseSensitive) {
			validator.hasValidExceptions = false
			return validator
		}
		absoluteExceptions = append(absoluteExceptions, absoluteException)
	}

	validator.isPathException = func(absoluteImportPath string) bool {
		for _, absoluteException := range absoluteExceptions {
			if containsPath(absoluteImportPath, absoluteException, caseSensitive) {
				return true
			}
		}
		return false
	}
	return validator
}

// containsPath reports whether filePath is target itself or one of its
// descendants, mirroring upstream's `relative === '' || !relative.startsWith('..')`
// check on `path.relative(target, filePath)`.
func containsPath(filePath string, target string, caseSensitive bool) bool {
	if !isPathWithin(filePath, target, caseSensitive) {
		return false
	}
	// `path.relative` joins the components left over, and upstream's prefix test
	// reads a leading `..` in that relative path as "outside target". A first
	// leftover component such as `..secret.js` therefore keeps the file out of
	// target even though it names a real child of it.
	fileComponents := pathComponents(filePath)
	targetComponents := pathComponents(target)
	return len(fileComponents) == len(targetComponents) ||
		!strings.HasPrefix(fileComponents[len(targetComponents)], "..")
}

// isPathWithin reports whether filePath resolves inside target, i.e. whether
// `path.relative(target, filePath)` walks no `..` component upwards. Components
// are compared the way the platform's `path.relative` does, so a leftover such
// as `..secret.js` still counts as inside target.
func isPathWithin(filePath string, target string, caseSensitive bool) bool {
	fileComponents := pathComponents(filePath)
	targetComponents := pathComponents(target)
	if len(targetComponents) > len(fileComponents) {
		return false
	}
	for i, component := range targetComponents {
		if !equalPathComponent(component, fileComponents[i], caseSensitive) {
			return false
		}
	}
	return true
}

func equalPathComponent(a string, b string, caseSensitive bool) bool {
	if caseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}

// pathComponents splits a path into its root plus segments, dropping the empty
// trailing segment a path written with a trailing separator would produce.
func pathComponents(p string) []string {
	components := tspath.GetPathComponents(p, "")
	for len(components) > 1 && components[len(components)-1] == "" {
		components = components[:len(components)-1]
	}
	return components
}

func unexpectedPathMessage(importPath string, customMessage string) rule.RuleMessage {
	description := fmt.Sprintf("Unexpected path \"%s\" imported in restricted zone.", importPath)
	if customMessage != "" {
		description += " " + customMessage
	}
	return rule.RuleMessage{
		Id:          "unexpectedPath",
		Description: description,
	}
}

var extglobPattern = regexp.MustCompile(`(\\).|([@?!+*]\(.*\))`)

// isGlob ports is-glob@4.0.3 in its default strict mode, which upstream uses to
// decide whether a zone path is a glob pattern or a plain directory path.
func isGlob(str string) bool {
	if str == "" {
		return false
	}
	if isExtglob(str) {
		return true
	}
	return strictCheck(str)
}

func isExtglob(str string) bool {
	for str != "" {
		match := extglobPattern.FindStringSubmatchIndex(str)
		if match == nil {
			return false
		}
		if match[4] != -1 {
			return true
		}
		str = str[match[1]:]
	}
	return false
}

func strictCheck(str string) bool {
	if str[0] == '!' {
		return true
	}

	index := 0
	pipeIndex := -2
	closeSquareIndex := -2
	closeCurlyIndex := -2
	closeParenIndex := -2
	backSlashIndex := -2

	for index < len(str) {
		if str[index] == '*' {
			return true
		}

		if charAt(str, index+1) == '?' && strings.IndexByte("].+)", str[index]) >= 0 {
			return true
		}

		if closeSquareIndex != -1 && str[index] == '[' && charAt(str, index+1) != ']' {
			if closeSquareIndex < index {
				closeSquareIndex = indexOfFrom(str, ']', index)
			}
			if closeSquareIndex > index {
				if backSlashIndex == -1 || backSlashIndex > closeSquareIndex {
					return true
				}
				backSlashIndex = indexOfFrom(str, '\\', index)
				if backSlashIndex == -1 || backSlashIndex > closeSquareIndex {
					return true
				}
			}
		}

		if closeCurlyIndex != -1 && str[index] == '{' && charAt(str, index+1) != '}' {
			closeCurlyIndex = indexOfFrom(str, '}', index)
			if closeCurlyIndex > index {
				backSlashIndex = indexOfFrom(str, '\\', index)
				if backSlashIndex == -1 || backSlashIndex > closeCurlyIndex {
					return true
				}
			}
		}

		if closeParenIndex != -1 && str[index] == '(' && charAt(str, index+1) == '?' &&
			strings.IndexByte(":!=", charAt(str, index+2)) >= 0 && charAt(str, index+3) != ')' {
			closeParenIndex = indexOfFrom(str, ')', index)
			if closeParenIndex > index {
				backSlashIndex = indexOfFrom(str, '\\', index)
				if backSlashIndex == -1 || backSlashIndex > closeParenIndex {
					return true
				}
			}
		}

		if pipeIndex != -1 && str[index] == '(' && charAt(str, index+1) != '|' {
			if pipeIndex < index {
				pipeIndex = indexOfFrom(str, '|', index)
			}
			if pipeIndex != -1 && charAt(str, pipeIndex+1) != ')' {
				closeParenIndex = indexOfFrom(str, ')', pipeIndex)
				if closeParenIndex > pipeIndex {
					backSlashIndex = indexOfFrom(str, '\\', pipeIndex)
					if backSlashIndex == -1 || backSlashIndex > closeParenIndex {
						return true
					}
				}
			}
		}

		if str[index] == '\\' {
			open := charAt(str, index+1)
			index += 2
			if closer, ok := closingChar(open); ok {
				if n := indexOfFrom(str, closer, index); n != -1 {
					index = n + 1
				}
			}
			if charAt(str, index) == '!' {
				return true
			}
		} else {
			index++
		}
	}

	return false
}

// charAt returns 0 for out-of-range positions, standing in for the `undefined`
// JavaScript reads past the end of a string.
func charAt(str string, index int) byte {
	if index < 0 || index >= len(str) {
		return 0
	}
	return str[index]
}

func indexOfFrom(str string, char byte, from int) int {
	if from < 0 {
		from = 0
	}
	if from > len(str) {
		return -1
	}
	offset := strings.IndexByte(str[from:], char)
	if offset == -1 {
		return -1
	}
	return from + offset
}

func closingChar(open byte) (byte, bool) {
	switch open {
	case '{':
		return '}', true
	case '(':
		return ')', true
	case '[':
		return ']', true
	}
	return 0, false
}
