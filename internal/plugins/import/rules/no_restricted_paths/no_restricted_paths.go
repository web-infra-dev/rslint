package no_restricted_paths

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/isglob"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
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
		if len(opts.zones) == 0 || !ctx.Program().IsValid() {
			return rule.RuleListeners{}
		}

		basePath := resolveBasePath(ctx, opts.basePath)
		// `path.relative` and Minimatch both pick their win32 flavor from the
		// platform rather than the host file system, so a case-insensitive
		// volume on macOS still gets a case-sensitive compare.
		windows := runtime.GOOS == "windows"

		currentFilename := ctx.SourceFile.FileName()
		matchingZones := make([]zone, 0, len(opts.zones))
		for _, z := range opts.zones {
			if isMatchingZone(z, basePath, currentFilename, windows) {
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
			absoluteImportPath, _, ok := ctx.Program().ResolveModule(ctx.SourceFile, source)
			if !ok {
				return
			}

			for i := range matchingZones {
				if !built[i] {
					validators[i] = makePathValidators(matchingZones[i].fromPaths, matchingZones[i].except, basePath, windows)
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
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = ctx.Program().CurrentDirectory()
	}
	if configured == "" {
		return tspath.NormalizePath(cwd)
	}
	return tspath.GetNormalizedAbsolutePath(configured, cwd)
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)

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

func isMatchingZone(z zone, basePath string, currentFilename string, windows bool) bool {
	for _, target := range z.targets {
		if isMatchingTargetPath(currentFilename, tspath.ResolvePath(basePath, target), windows) {
			return true
		}
	}
	return false
}

func isMatchingTargetPath(fileName string, targetPath string, windows bool) bool {
	if isglob.Is(targetPath) {
		return minimatch3.Match(targetPath, fileName, minimatch3.Options{})
	}
	return containsPath(fileName, targetPath, windows)
}

func makePathValidators(fromPaths []string, except []string, basePath string, windows bool) []pathValidator {
	anyGlob := false
	anyNonGlob := false
	for _, from := range fromPaths {
		if isglob.Is(from) {
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
			validators = append(validators, computeGlobPatternPathValidator(absoluteFrom, except, windows))
		} else {
			validators = append(validators, computeAbsolutePathValidator(absoluteFrom, except, windows))
		}
	}
	return validators
}

// computeGlobPatternPathValidator matches both `from` and `except` as glob
// patterns. Note that upstream matches `except` patterns verbatim rather than
// resolving them against `basePath` first, so only absolute exception patterns
// can ever match a resolved import path.
func computeGlobPatternPathValidator(absoluteFrom string, except []string, windows bool) pathValidator {
	fromMatcher := minimatch3.New(absoluteFrom, minimatch3.Options{})
	validator := pathValidator{
		isPathRestricted: func(absoluteImportPath string) bool {
			return fromMatcher.Match(absoluteImportPath)
		},
		hasValidExceptions: true,
		invalidException: rule.RuleMessage{
			Id:          "invalidExceptionGlob",
			Description: "Restricted path exceptions must be glob patterns when `from` contains glob patterns",
		},
	}

	for _, exception := range except {
		if !isglob.Is(exception) {
			validator.hasValidExceptions = false
			return validator
		}
	}

	// The exception patterns stay unresolved, so a Windows-native one such as
	// `C:\repo\server\allowed\**\*` still carries its backslashes. Upstream's
	// Minimatch constructor rewrites the platform separator to `/` before it
	// compiles a pattern; without the same rewrite the backslashes would be
	// read as escapes and could never match an import path normalized to `/`.
	matchers := make([]*minimatch3.Matcher, 0, len(except))
	for _, exception := range except {
		if windows {
			exception = strings.ReplaceAll(exception, `\`, "/")
		}
		matchers = append(matchers, minimatch3.New(exception, minimatch3.Options{}))
	}
	validator.isPathException = func(absoluteImportPath string) bool {
		for _, matcher := range matchers {
			if matcher.Match(absoluteImportPath) {
				return true
			}
		}
		return false
	}
	return validator
}

func computeAbsolutePathValidator(absoluteFrom string, except []string, windows bool) pathValidator {
	validator := pathValidator{
		isPathRestricted: func(absoluteImportPath string) bool {
			return containsPath(absoluteImportPath, absoluteFrom, windows)
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
		if !isPathWithin(absoluteException, absoluteFrom, windows) {
			validator.hasValidExceptions = false
			return validator
		}
		absoluteExceptions = append(absoluteExceptions, absoluteException)
	}

	validator.isPathException = func(absoluteImportPath string) bool {
		for _, absoluteException := range absoluteExceptions {
			if containsPath(absoluteImportPath, absoluteException, windows) {
				return true
			}
		}
		return false
	}
	return validator
}

// containsPath reports whether filePath is target itself or one of its
// descendants, mirroring upstream's `relative === ” || !relative.startsWith('..')`
// check on `path.relative(target, filePath)`.
func containsPath(filePath string, target string, windows bool) bool {
	if !isPathWithin(filePath, target, windows) {
		return false
	}
	if rootsDiverge(filePath, target, windows) {
		return true
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
func isPathWithin(filePath string, target string, windows bool) bool {
	if rootsDiverge(filePath, target, windows) {
		return true
	}
	fileComponents := pathComponents(filePath)
	targetComponents := pathComponents(target)
	if len(targetComponents) > len(fileComponents) {
		return false
	}
	for i, component := range targetComponents {
		if !equalPathComponent(component, fileComponents[i], windows) {
			return false
		}
	}
	return true
}

// rootsDiverge reports whether `path.win32.relative` gives up on the pair and
// hands back its absolute second argument, which it does once the two paths
// differ before their first separator: different drive letters, or different
// UNC servers. That answer is neither empty nor `..`-prefixed, so upstream ends
// up reading filePath as inside target. Two shares on one UNC server still
// share that first segment and keep the ordinary component comparison.
func rootsDiverge(filePath string, target string, windows bool) bool {
	return windows && !ecmascript.EqualsWhenLowercased(firstPathSegment(filePath), firstPathSegment(target))
}

// firstPathSegment returns the leading segment `path.win32.relative` compares
// before it reaches a separator, skipping the separators a UNC path starts
// with: the drive of `C:/repo/a`, or the server of `//server/share/a`.
func firstPathSegment(p string) string {
	p = strings.TrimLeft(p, "/")
	if index := strings.IndexByte(p, '/'); index >= 0 {
		return p[:index]
	}
	return p
}

func equalPathComponent(a string, b string, windows bool) bool {
	if windows {
		return ecmascript.EqualsWhenLowercased(a, b)
	}
	return a == b
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
