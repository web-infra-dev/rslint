// Package utils contains import classification helpers shared by
// eslint-plugin-import rules.
package utils

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var (
	scopedModuleRegexp = regexp.MustCompile(`^@[^/]+/?[^/]+`)
	moduleRegexp       = regexp.MustCompile(`^\w`)
)

// IsScoped reports whether name starts with an npm-style scoped package name.
func IsScoped(name string) bool {
	return name != "" && scopedModuleRegexp.MatchString(name)
}

// IsExternalLooking reports whether name has the shape of a bare package
// specifier. This deliberately follows import-js rather than Node's complete
// package-name grammar.
func IsExternalLooking(name string) bool {
	return name != "" && (moduleRegexp.MatchString(name) || IsScoped(name))
}

// BaseModule returns the package portion of a bare specifier.
func BaseModule(name string) string {
	if IsScoped(name) {
		parts := strings.Split(name, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		// importType.js destructures `[scope, pkg] = name.split('/')` and
		// interpolates the missing package as JavaScript's `undefined`.
		return parts[0] + "/undefined"
	}
	if slash := strings.IndexByte(name, '/'); slash >= 0 {
		return name[:slash]
	}
	return name
}

// IsBuiltinModule mirrors importType.js's isBuiltIn. A successfully resolved
// path suppresses builtin classification so a project-local module named like
// a Node builtin remains internal.
func IsBuiltinModule(name string, settings map[string]interface{}, resolvedPath string) bool {
	if name == "" || resolvedPath != "" {
		return false
	}
	base := BaseModule(name)
	if core.NodeCoreModules()[base] {
		return true
	}
	if settings == nil {
		return false
	}
	switch extras := settings["import/core-modules"].(type) {
	case []string:
		for _, extra := range extras {
			if extra == base {
				return true
			}
		}
	case []interface{}:
		for _, raw := range extras {
			if extra, ok := raw.(string); ok && extra == base {
				return true
			}
		}
	}
	return false
}

// IsAbsolutePath reports whether name is an absolute disk path.
func IsAbsolutePath(name string) bool {
	return name != "" && tspath.IsRootedDiskPath(name)
}

// IsRelativeToParent reports whether name is .. or starts with ../ or ..\.
func IsRelativeToParent(name string) bool {
	return name == ".." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, `..\`)
}

// IsRelativeToSibling reports whether name starts with ./ or .\.
func IsRelativeToSibling(name string) bool {
	return strings.HasPrefix(name, "./") || strings.HasPrefix(name, `.\`)
}

// IsIndexImport reports whether name is one of import-js's canonical index
// spellings. Other extensions intentionally remain sibling imports.
func IsIndexImport(name string) bool {
	switch name {
	case ".", "./", "./index", "./index.js":
		return true
	default:
		return false
	}
}

// ClassifyImport returns import-js's import type for a module specifier.
func ClassifyImport(ctx rule.RuleContext, name string, specifier *ast.Node) string {
	settings := SettingsFor(ctx)
	// Internal-regex and absolute classifications win without consulting the
	// resolved target. Builtins come next in upstream, and a successful
	// resolution suppresses even a user-configured core-module name, so resolve
	// before classifying relative specifiers as well.
	if settings.IsInternalSpecifier(name) {
		return "internal"
	}
	if IsAbsolutePath(name) {
		return "absolute"
	}

	resolvedPath := ""
	sourceProgram := ctx.Program()
	if sourceProgram != nil && specifier != nil && ast.IsStringLiteralLike(specifier) {
		resolvedPath, _, _ = sourceProgram.ResolveModule(ctx.SourceFile, specifier)
	}

	if IsBuiltinModule(name, ctx.Settings, resolvedPath) {
		return "builtin"
	}
	if IsRelativeToParent(name) {
		return "parent"
	}
	if IsIndexImport(name) {
		return "index"
	}
	if IsRelativeToSibling(name) {
		return "sibling"
	}

	packagePath := contextPackagePath(ctx)
	caseSensitive := true
	if sourceProgram != nil && sourceProgram.FS() != nil {
		caseSensitive = sourceProgram.FS().UseCaseSensitiveFileNames()
	}
	if settings.IsExternalPathFromPackage(packagePath, resolvedPath, caseSensitive) {
		return "external"
	}
	if resolvedPath != "" {
		return "internal"
	}
	if IsExternalLooking(name) {
		return "external"
	}
	return "unknown"
}

func contextPackagePath(ctx rule.RuleContext) string {
	start := ""
	if ctx.SourceFile != nil {
		start = tspath.GetDirectoryPath(ctx.SourceFile.FileName())
	}
	sourceProgram := ctx.Program()
	if sourceProgram != nil && start != "" {
		// Reuse TypeScript's package-scope cache instead of walking and probing
		// package.json ancestors for every classified import.
		if packageDir := sourceProgram.NearestPackageJSONDirectory(start); packageDir != "" {
			return tspath.NormalizePath(packageDir)
		}
	}
	if cwd := ctx.ProcessCurrentDirectory(); cwd != "" {
		return tspath.NormalizePath(cwd)
	}
	if sourceProgram != nil && sourceProgram.CurrentDirectory() != "" {
		return tspath.NormalizePath(sourceProgram.CurrentDirectory())
	}
	return tspath.NormalizePath(start)
}
