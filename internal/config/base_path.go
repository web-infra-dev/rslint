package config

import (
	"path"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
)

// ResolveBasePaths desugars per-entry BasePath into ordinary relative
// files / ignores / parserOptions.project patterns, then clears BasePath.
//
// configDirectory is the match root the rest of the engine already uses:
// config-file directory for auto-discovered configs, cwd for --config.
// Relative BasePath values resolve against it. Absolute BasePath values are
// rebased via ConvertToRelativePath.
//
// JS/TS configs normally desugar basePath in Node before the payload reaches
// Go; this path covers JSON/JSONC configs and any raw ConfigEntry constructed
// with BasePath still set.
func ResolveBasePaths(config RslintConfig, configDirectory string) RslintConfig {
	if len(config) == 0 {
		return config
	}
	for i := range config {
		resolveBasePathForEntry(&config[i], configDirectory)
	}
	return config
}

func resolveBasePathForEntry(entry *ConfigEntry, configDirectory string) {
	if entry.BasePath == "" {
		return
	}

	relativeBase := resolveRelativeBasePath(entry.BasePath, configDirectory)
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
}

func entryScopesToSubtree(entry ConfigEntry) bool {
	return entry.Rules != nil ||
		entry.Plugins != nil ||
		entry.Settings != nil ||
		entry.LanguageOptions != nil
}

// resolveRelativeBasePath returns a POSIX-ish path relative to configDirectory.
// Empty means the config root itself (no prefix).
func resolveRelativeBasePath(basePath string, configDirectory string) string {
	if basePath == "" {
		return ""
	}
	absolute := tspath.ResolvePath(configDirectory, basePath)
	relative := tspath.ConvertToRelativePath(absolute, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: true,
		CurrentDirectory:          configDirectory,
	})
	relative = strings.ReplaceAll(tspath.NormalizePath(relative), "\\", "/")
	relative = strings.TrimPrefix(relative, "./")
	if relative == "." || relative == "" {
		return ""
	}
	return relative
}

// rebasePattern prefixes a relative glob/path with relativeBase.
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

func isAbsolutePattern(pattern string) bool {
	if pattern == "" {
		return false
	}
	// POSIX absolute, Windows drive, or UNC.
	if strings.HasPrefix(pattern, "/") || strings.HasPrefix(pattern, "\\") {
		return true
	}
	if len(pattern) >= 3 {
		drive := pattern[0]
		if (drive >= 'A' && drive <= 'Z') || (drive >= 'a' && drive <= 'z') {
			if pattern[1] == ':' && (pattern[2] == '/' || pattern[2] == '\\') {
				return true
			}
		}
	}
	return tspath.GetEncodedRootLength(pattern) > 0
}
