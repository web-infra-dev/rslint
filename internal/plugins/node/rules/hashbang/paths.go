// cspell:ignore globrex licen npmignore
package hashbang

import (
	"encoding/json"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

type packageJSON struct {
	directory string
	data      map[string]any
}

type packageJSONKey string

func findPackage(p *program.Program, fileName string) *packageJSON {
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
		pkg := program.Cached(p, packageJSONKey(directory), func() *packageJSON {
			if text, ok := p.FS().ReadFile(tspath.ResolvePath(directory, "package.json")); ok {
				var data map[string]any
				if json.Unmarshal([]byte(text), &data) == nil && data != nil {
					return &packageJSON{directory: directory, data: data}
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

func fileExtension(fileName string) string {
	base := path.Base(fileName)
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return ""
	}
	return path.Ext(fileName)
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

var conversionGlobLiterals = strings.NewReplacer("?", `\?`, "[", `\[`, "]", `\]`)

func convertPath(fileName string, options, settings map[string]any) (string, bool) {
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
			return minimatch3.Match(pattern, fileName, minimatch3.Options{Dot: true, NoBrace: true, NoExt: true, NoNegate: true, NoComment: true})
		}
		if slices.ContainsFunc(utils.ToStringSlice(value["include"]), matches) && !slices.ContainsFunc(utils.ToStringSlice(value["exclude"]), matches) {
			converted, err := expression.ReplaceFirst(fileName, replacement[1])
			return converted, err == nil
		}
	}
	return fileName, true
}

var neverIgnored = esregexp.MustCompile(`^(?:readme\.[^.]*|(?:licen[cs]e|changes|changelog|history)(?:\.[^.]*)?)$`, "iu")

func isUnpublished(program *program.Program, absolute, relative string) bool {
	pkg := findPackage(program, absolute)
	if pkg == nil {
		return strings.HasPrefix(relative, "..")
	}
	files, hasFiles := pkg.data["files"].([]any)
	ignoreText, hasIgnore := program.FS().ReadFile(tspath.ResolvePath(pkg.directory, ".npmignore"))
	if !hasIgnore && !hasFiles {
		ignoreText, hasIgnore = program.FS().ReadFile(tspath.ResolvePath(pkg.directory, ".gitignore"))
	}
	if !hasFiles && !hasIgnore {
		return strings.HasPrefix(relative, "..")
	}
	if main, ok := pkg.data["main"].(string); ok && tspath.ResolvePath(pkg.directory, main) == tspath.ResolvePath(pkg.directory, relative) {
		return false
	}
	// Upstream resolves this particular exemption against the process cwd.
	fromCWD := tspath.GetRelativePathFromDirectory(pkg.directory, tspath.ResolvePath(program.CurrentDirectory(), relative), tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
	if relative == "package.json" || neverIgnored.Test(fromCWD) {
		return false
	}
	if strings.HasPrefix(relative, "..") || hasIgnore && gitignore.NewMatcher([]string{ignoreText}, false).Match(relative) {
		return true
	}
	if !hasFiles {
		return false
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
		if strings.ContainsAny(body, "{}") || strings.Contains(body, "(") {
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
	matchesExtended := func(pattern string) bool {
		// Reuse minimatch's brace expansion, extended groups and UTF-16 matching.
		return minimatch3.Match(pattern, relative, minimatch3.Options{Dot: true, NoCase: true, NoNegate: true, NoComment: true})
	}
	included := gitignore.NewMatcher(includes, false).Match(relative) || slices.ContainsFunc(extendedIncludes, matchesExtended)
	excluded := gitignore.NewMatcher(excludes, false).Match(relative) || slices.ContainsFunc(extendedExcludes, matchesExtended)
	return !included || excluded
}
