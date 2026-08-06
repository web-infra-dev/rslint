package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// ResolveBasePaths desugars per-entry BasePath into ordinary relative
// files / ignores / parserOptions.project patterns, then clears BasePath.
//
// configDirectory is the match root the rest of the engine already uses:
// config-file directory for auto-discovered configs, cwd for --config.
// Relative BasePath values resolve against it. Absolute BasePath values are
// rebased via ConvertToRelativePath. fsys supplies filesystem case sensitivity
// (as the rest of config matching does) so an absolute basePath whose casing
// differs from the match root still resolves on case-insensitive filesystems.
//
// A BasePath that resolves outside configDirectory is rejected: its desugared
// patterns would be ../-prefixed and could never match under the engine's
// single-match-root model, indistinguishable from a typo. ESLint reports such
// files as "ignored because outside of base path"; rslint fails fast instead
// so the misconfiguration is actionable.
//
// JS/TS configs normally desugar basePath in Node before the payload reaches
// Go; this path covers JSON/JSONC configs and any raw ConfigEntry constructed
// with BasePath still set.
func ResolveBasePaths(config RslintConfig, configDirectory string, fsys vfs.FS) (RslintConfig, error) {
	if len(config) == 0 {
		return config, nil
	}
	useCaseSensitive := fsys == nil || fsys.UseCaseSensitiveFileNames()
	for i := range config {
		if err := resolveBasePathForEntry(&config[i], configDirectory, useCaseSensitive); err != nil {
			return nil, fmt.Errorf("config entry at index %d: %w", i, err)
		}
	}
	return config, nil
}

func resolveBasePathForEntry(entry *ConfigEntry, configDirectory string, useCaseSensitive bool) error {
	if entry.BasePath == "" {
		return nil
	}

	relativeBase, err := resolveRelativeBasePath(entry.BasePath, configDirectory, useCaseSensitive)
	if err != nil {
		return err
	}
	entry.BasePath = ""

	if len(entry.Files) > 0 {
		rebases := make([]string, len(entry.Files))
		for i, pattern := range entry.Files {
			rebases[i] = rebasePattern(pattern, relativeBase)
		}
		entry.Files = rebases
	}
	if len(entry.FilePatternGroups) > 0 {
		groups := make([][]string, len(entry.FilePatternGroups))
		for i, group := range entry.FilePatternGroups {
			rebases := make([]string, len(group))
			for j, pattern := range group {
				rebases[j] = rebasePattern(pattern, relativeBase)
			}
			groups[i] = rebases
		}
		entry.FilePatternGroups = groups
	}
	if len(entry.Ignores) > 0 {
		rebases := make([]string, len(entry.Ignores))
		for i, pattern := range entry.Ignores {
			rebases[i] = rebasePattern(pattern, relativeBase)
		}
		entry.Ignores = rebases
	}
	if entry.LanguageOptions != nil && entry.LanguageOptions.ParserOptions != nil {
		projects := entry.LanguageOptions.ParserOptions.Project
		if len(projects) > 0 {
			rebases := make([]string, len(projects))
			for i, project := range projects {
				rebases[i] = rebasePattern(project, relativeBase)
			}
			entry.LanguageOptions.ParserOptions.Project = rebases
		}
	}

	// Scoped entry with basePath but no files → limit to the base subtree,
	// matching ESLint's "bare basePath applies under that directory" intent.
	if !hasFileSelectors(*entry) && entryScopesToSubtree(*entry) {
		catchAll := "**"
		if relativeBase != "" {
			catchAll = path.Join(relativeBase, "**")
		}
		entry.Files = []string{catchAll}
	}
	return nil
}

func entryScopesToSubtree(entry ConfigEntry) bool {
	return entry.Rules != nil ||
		entry.Plugins != nil ||
		entry.Settings != nil ||
		entry.LanguageOptions != nil
}

// resolveRelativeBasePath returns a glob-escaped POSIX-ish path relative to
// configDirectory. Empty means the config root itself (no prefix). It rejects
// a basePath that resolves outside configDirectory, since the resulting
// ../-prefixed patterns could never match under the engine's single-match-root
// model.
func resolveRelativeBasePath(basePath string, configDirectory string, useCaseSensitive bool) (string, error) {
	if basePath == "" {
		return "", nil
	}
	// tspath.ResolvePath normalizes to forward slashes, but callers (and test
	// fixtures) may hand in a configDirectory with native separators. Normalize
	// once so the containment check and relative rebase compare like for like.
	if configDirectory != "" {
		configDirectory = tspath.NormalizePath(configDirectory)
	}
	absolute := tspath.ResolvePath(configDirectory, basePath)
	if _, ok := RelativePathWithinConfigRoot(absolute, configDirectory, useCaseSensitive); !ok {
		return "", fmt.Errorf(
			"key \"basePath\": value %q resolves outside the config match root %q; basePath must be a subdirectory of the config root",
			basePath,
			configDirectory,
		)
	}
	relative := tspath.ConvertToRelativePath(absolute, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: useCaseSensitive,
		CurrentDirectory:          configDirectory,
	})
	relative = strings.ReplaceAll(tspath.NormalizePath(relative), "\\", "/")
	relative = strings.TrimPrefix(relative, "./")
	if relative == "." || relative == "" {
		return "", nil
	}
	return escapeGlobBasePath(relative), nil
}

// escapeGlobBasePath escapes glob metacharacters in a resolved basePath so the
// directory is treated literally when spliced into glob patterns. ESLint treats
// basePath as a literal path; without escaping, a basePath such as
// "packages/[locale]" would be interpreted as a character class and never match.
//
// Escaping uses character classes ([*], [?], [[] ...), which the doublestar and
// picomatch matchers understand on every platform and which survive the
// tspath.NormalizePath applied to patterns before matching. A bare ']' or '}'
// outside a class/brace pair is already literal to the matchers, so only the
// characters that open a class/brace or introduce a wildcard need escaping. A
// leading '!' (basePath "!vendor") has no class form the matchers accept as
// literal, so it remains a known limitation.
func escapeGlobBasePath(s string) string {
	return strings.NewReplacer(
		"[", "[[]",
		"*", "[*]",
		"?", "[?]",
		"{", "[{]",
	).Replace(s)
}

// rebasePattern prefixes a relative glob/path with the glob-escaped
// relativeBase produced by resolveRelativeBasePath.
// Leading ! negation is preserved; absolute patterns are left unchanged;
// empty relativeBase only strips a leading ./.
func rebasePattern(pattern string, relativeBase string) string {
	negated := false
	body := pattern
	for strings.HasPrefix(body, "!") {
		negated = !negated
		body = strings.TrimPrefix(body, "!")
	}
	body = strings.ReplaceAll(body, "\\", "/")
	for strings.HasPrefix(body, "./") {
		body = strings.TrimPrefix(body, "./")
	}
	if body == "" || isAbsolutePattern(body) {
		if negated {
			return "!" + body
		}
		return body
	}
	if relativeBase == "" {
		if negated {
			return "!" + body
		}
		return body
	}
	joined := path.Join(relativeBase, body)
	if negated {
		return "!" + joined
	}
	return joined
}

// isAbsolutePattern reports whether pattern is an absolute path in any form
// tspath recognizes (POSIX, UNC, Windows drive-letter).
func isAbsolutePattern(pattern string) bool {
	return tspath.GetEncodedRootLength(pattern) > 0
}
