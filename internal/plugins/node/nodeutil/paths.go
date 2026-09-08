// cspell:ignore globrex licen npmignore
// Package nodeutil contains the package metadata and publication policies shared
// by rules ported from eslint-plugin-n. It does not resolve JavaScript imports.
package nodeutil

import (
	"encoding/json"
	"maps"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/gitignore"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

// PackageJSON is immutable package metadata from one Program generation.
type PackageJSON struct {
	directory string
	data      map[string]any
}

type packageJSONKey string

// FindPackage finds the nearest valid package object using the Program resolver.
// Malformed package metadata falls back to the next parent package, as upstream does.
func FindPackage(p *program.Program, fileName string) *PackageJSON {
	if p.FS() == nil {
		return nil
	}
	for directory := tspath.GetDirectoryPath(fileName); directory != ""; {
		directory = p.NearestPackageJSONDirectory(directory)
		if directory == "" {
			return nil
		}
		// Package contents are immutable within this Program generation. Share
		// decoding across its files without retaining the Program in the value.
		pkg := program.Cached(p, packageJSONKey(directory), func() *PackageJSON {
			if text, ok := p.FS().ReadFile(tspath.ResolvePath(directory, "package.json")); ok {
				var data map[string]any
				if json.Unmarshal([]byte(text), &data) == nil && data != nil {
					return &PackageJSON{directory: directory, data: data}
				}
			}
			return nil
		})
		if pkg != nil {
			return pkg
		}
		parent := tspath.GetDirectoryPath(directory)
		if parent == directory {
			return nil
		}
		directory = parent
	}
	return nil
}

// Directory returns the normalized directory containing package.json.
func (pkg *PackageJSON) Directory() string { return pkg.directory }

// IsBinFile applies eslint-plugin-n's syntactic bin aliases (.js and /index).
// This deliberately does not perform TypeScript module or exports resolution.
func (pkg *PackageJSON) IsBinFile(fileName string) bool {
	return isBinFile(fileName, pkg.data["bin"], pkg.directory)
}

func isBinFile(fileName string, bin any, directory string) bool {
	match := func(value any) bool {
		name, ok := value.(string)
		if !ok {
			return false
		}
		resolved := tspath.ResolvePath(directory, name)
		withoutJS := strings.TrimSuffix(fileName, ".js")
		return resolved == fileName || resolved == withoutJS || resolved == strings.TrimSuffix(withoutJS, "/index")
	}
	switch value := bin.(type) {
	case string:
		return value != "" && match(value)
	case map[string]any:
		for _, name := range value {
			if match(name) {
				return true
			}
		}
	case []any:
		return slices.ContainsFunc(value, match)
	}
	return false
}

var conversionGlobLiterals = strings.NewReplacer(`\`, `\\`, "?", `\?`, "[", `\[`, "]", `\]`)

// ConvertPath applies the configured source-to-published path mapping.
// The boolean is false if the conversion cannot be evaluated safely.
func ConvertPath(fileName string, options, settings map[string]any) (string, bool) {
	conversion := options["convertPath"]
	for _, name := range []string{"n", "node"} {
		if conversion == nil {
			if shared, ok := settings[name].(map[string]any); ok {
				conversion = shared["convertPath"]
			}
		}
	}
	var entries []any
	switch value := conversion.(type) {
	case []any:
		entries = value
	case map[string]any:
		// Go config objects have no property order. The array form retains
		// explicit priority; overlapping object patterns use lexical order.
		for _, pattern := range slices.Sorted(maps.Keys(value)) {
			entries = append(entries, map[string]any{"include": []any{pattern}, "replace": value[pattern]})
		}
	}
	for _, entry := range entries {
		value, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		replacement := utils.ToStringSlice(value["replace"])
		if len(replacement) != 2 {
			return fileName, false
		}
		expression, err := esregexp.Compile(replacement[0], "")
		if err != nil {
			return fileName, false
		}
		matches := func(pattern string) bool {
			// globrex's convertPath mode disables extended syntax. Reuse
			// minimatch with literal ?, classes and braces in this mode.
			pattern = conversionGlobLiterals.Replace(pattern)
			return minimatch3.Match(pattern, fileName, minimatch3.Options{Dot: true, NoBrace: true, NoExt: true, NoNegate: true, NoComment: true, PreserveWhitespace: true})
		}
		if slices.ContainsFunc(utils.ToStringSlice(value["include"]), matches) && !slices.ContainsFunc(utils.ToStringSlice(value["exclude"]), matches) {
			converted, err := expression.ReplaceFirst(fileName, replacement[1])
			return converted, err == nil
		}
	}
	return fileName, true
}

var extendedGlob = esregexp.MustCompile(`[{}]|[@+?!*]\(`, "")

var neverIgnored = esregexp.MustCompile(`^(?:readme\.[^.]*|(?:licen[cs]e|changes|changelog|history)(?:\.[^.]*)?)$`, "iu")

type publicationKey string

type publication struct {
	hasFiles, hasIgnore                bool
	ignore, includes, excludes         *gitignore.Matcher
	extendedIncludes, extendedExcludes []*minimatch3.Matcher
}

// IsUnpublished applies eslint-plugin-n publication policy, including files,
// npmignore/gitignore precedence, exclusions, and always-published metadata.
// convertedPath is the mapping relative to the source package; absolute is its
// resolved target. A nested target package supplies its own publication paths.
func IsUnpublished(p *program.Program, absolute, convertedPath string) bool {
	// Resolve interior . and .. before testing the ancestor component.
	convertedPath = tspath.NormalizePath(convertedPath)
	if convertedPath == ".." || strings.HasPrefix(convertedPath, "../") {
		return true
	}
	pkg := FindPackage(p, absolute)
	if pkg == nil {
		return false
	}
	// Publication patterns belong to the package containing the converted
	// target. Rebase the path if conversion entered a nested package.
	relative := tspath.GetRelativePathFromDirectory(pkg.directory, absolute, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
	published := program.Cached(p, publicationKey(pkg.directory), func() *publication {
		return compilePublication(p, pkg)
	})
	if !published.hasFiles && !published.hasIgnore {
		return false
	}
	if main, ok := pkg.data["main"].(string); ok && tspath.ResolvePath(pkg.directory, main) == tspath.ResolvePath(pkg.directory, relative) {
		return false
	}
	if relative == "package.json" || neverIgnored.Test(relative) {
		return false
	}
	if published.ignore.Match(relative) {
		return true
	}
	if !published.hasFiles {
		return false
	}
	matches := func(pattern *minimatch3.Matcher) bool { return pattern.Match(relative) }
	included := published.includes.Match(relative) || slices.ContainsFunc(published.extendedIncludes, matches)
	excluded := published.excludes.Match(relative) || slices.ContainsFunc(published.extendedExcludes, matches)
	return !included || excluded
}

func compilePublication(p *program.Program, pkg *PackageJSON) *publication {
	files, hasFiles := pkg.data["files"].([]any)
	ignoreText, hasIgnore := p.FS().ReadFile(tspath.ResolvePath(pkg.directory, ".npmignore"))
	if !hasIgnore && !hasFiles {
		ignoreText, hasIgnore = p.FS().ReadFile(tspath.ResolvePath(pkg.directory, ".gitignore"))
	}
	published := &publication{hasFiles: hasFiles, hasIgnore: hasIgnore}
	if hasIgnore {
		published.ignore = gitignore.NewMatcherFromText(ignoreText, false)
	}
	var includes, excludes []string
	var extendedIncludes, extendedExcludes []string
	for _, file := range files {
		pattern, ok := file.(string)
		if !ok || pattern == "" {
			continue
		}
		negated := strings.HasPrefix(pattern, "!")
		body := pattern
		// Remove only one marker: !/foo.js must retain /foo.js.
		if negated || strings.HasPrefix(body, "/") {
			body = body[1:]
		}
		body = strings.TrimRight(path.Clean(body), "/")
		if extendedGlob.Test(body) {
			patterns := []string{body, body + "/**"}
			if negated {
				if !strings.Contains(body, "/") {
					patterns = append(patterns, "**/"+body)
				}
				extendedExcludes = append(extendedExcludes, patterns...)
			} else {
				extendedIncludes = append(extendedIncludes, patterns...)
			}
		} else if negated {
			excludes = append(excludes, body, body+"/**")
		} else {
			includes = append(includes, "/"+body, "/"+body+"/**")
		}
	}
	// The files field is a pattern list. Its entries stay atomic, just as
	// additionalExecutables entries do; only ignore file text is line-delimited.
	published.includes = gitignore.NewMatcher(includes, false)
	published.excludes = gitignore.NewMatcher(excludes, false)
	compile := func(patterns []string) []*minimatch3.Matcher {
		var result []*minimatch3.Matcher
		for _, pattern := range patterns {
			result = append(result, minimatch3.New(pattern, minimatch3.Options{Dot: true, NoCase: true, NoNegate: true, NoComment: true, PreserveWhitespace: true}))
		}
		return result
	}
	published.extendedIncludes = compile(extendedIncludes)
	published.extendedExcludes = compile(extendedExcludes)
	return published
}

type ignorePatternsKey string

// MatchIgnorePatterns shares compiled explicit pattern lists within a Program.
// Length-prefixed entries preserve list boundaries, embedded newlines, and order.
func MatchIgnorePatterns(p *program.Program, patterns []string, relative string) bool {
	if len(patterns) == 0 {
		return false
	}
	var key strings.Builder
	size := 0
	for _, pattern := range patterns {
		size += len(pattern) + 8
	}
	key.Grow(size)
	for _, pattern := range patterns {
		key.WriteString(strconv.Itoa(len(pattern)))
		key.WriteByte(':')
		key.WriteString(pattern)
	}
	matcher := program.Cached(p, ignorePatternsKey(key.String()), func() *gitignore.Matcher {
		return gitignore.NewMatcher(patterns, false)
	})
	return matcher.Match(relative)
}
