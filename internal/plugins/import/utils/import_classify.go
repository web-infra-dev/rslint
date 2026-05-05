// Package utils — import classification helpers shared across import-plugin
// rules.
//
// The classification mirrors eslint-plugin-import's `importType.js` (returning
// one of `"builtin"`, `"external"`, `"internal"`, `"parent"`, `"sibling"`,
// `"index"`, `"absolute"` or `"unknown"`), but uses tsgo's module resolver and
// path utilities so we get tsconfig `paths`, monorepo symlinks, and the
// `IsExternalLibraryImport` flag for free.
package utils

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// scopedModuleRegexp matches `@scope/pkg` style names.
var scopedModuleRegexp = regexp.MustCompile(`^@[^/]+/?[^/]+`)

// moduleRegexp matches names that "look like" packages (start with a word char).
var moduleRegexp = regexp.MustCompile(`^\w`)

// IsScoped reports whether name is an `@scope/pkg` style specifier.
func IsScoped(name string) bool { return scopedModuleRegexp.MatchString(name) }

// IsExternalLooking reports whether the bare specifier looks like a package
// name (starts with a word character, or is a scoped name).
func IsExternalLooking(name string) bool {
	return moduleRegexp.MatchString(name) || IsScoped(name)
}

// BaseModule returns the package portion of a specifier:
//   - "lodash"            -> "lodash"
//   - "lodash/fp"         -> "lodash"
//   - "@scope/pkg"        -> "@scope/pkg"
//   - "@scope/pkg/sub"    -> "@scope/pkg"
//   - "fs/promises"       -> "fs/promises" (when the joined form is itself a builtin; see IsBuiltinModule)
//
// Note: `fs/promises` is preserved in the IsBuiltinModule lookup because the
// joined name is itself in the Node builtin table.
func BaseModule(name string) string {
	if IsScoped(name) {
		// Walk to the second `/` if present.
		first := strings.IndexByte(name, '/')
		if first < 0 {
			return name
		}
		rest := name[first+1:]
		next := strings.IndexByte(rest, '/')
		if next < 0 {
			return name
		}
		return name[:first+1+next]
	}
	if i := strings.IndexByte(name, '/'); i >= 0 {
		return name[:i]
	}
	return name
}

// IsBuiltinModule returns true if name is a Node.js core module (with or
// without the `node:` prefix), or if the user listed it under
// `settings["import/core-modules"]`.
//
// The core list is sourced from tsgo's `core.NodeCoreModules()` (a
// `sync.OnceValue` keyed by both bare and `node:`-prefixed forms), so it
// stays in sync with whatever Node.js builtins the TypeScript compiler we're
// linked against recognizes — no hand-maintained table to drift.
func IsBuiltinModule(name string, settings map[string]any) bool {
	if name == "" {
		return false
	}
	if core.NodeCoreModules()[name] {
		return true
	}
	if settings != nil {
		if extras, ok := settings["import/core-modules"].([]any); ok {
			stripped := strings.TrimPrefix(name, "node:")
			base := BaseModule(stripped)
			for _, e := range extras {
				s, ok := e.(string)
				if !ok {
					continue
				}
				if s == name || s == stripped || s == base {
					return true
				}
			}
		}
	}
	return false
}

// IsAbsolutePath returns true if name is an OS-absolute path (`/foo`, `C:\foo`,
// etc). Delegates to tsgo's `tspath.IsRootedDiskPath` for cross-platform parity.
func IsAbsolutePath(name string) bool {
	return tspath.IsRootedDiskPath(name)
}

// IsRelativeToParent returns true for `..` / `../...` / `..\\...`.
func IsRelativeToParent(name string) bool {
	return name == ".." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, `..\`)
}

// IsRelativeToSibling returns true for `./...` / `.\\...`.
func IsRelativeToSibling(name string) bool {
	return strings.HasPrefix(name, "./") || strings.HasPrefix(name, `.\`)
}

// IsIndexImport returns true for `.`, `./`, `./index`, `./index.js`.
//
// (Mirrors the canonical list from `importType.js`. We keep the literal set —
// any other extension is `sibling`, not `index`.)
func IsIndexImport(name string) bool {
	switch name {
	case ".", "./", "./index", "./index.js":
		return true
	}
	return false
}

// ClassifyImport returns the import-type bucket for the given specifier.
//
// Resolution policy (mirrors upstream `importType.js`):
//
//  1. If `import/internal-regex` is set and matches the name → "internal".
//  2. Absolute path (`/foo`, `C:\\foo`) → "absolute".
//  3. Node builtin → "builtin".
//  4. Starts with `..` → "parent".
//  5. Index marker (`.`, `./`, `./index`, `./index.js`) → "index".
//  6. Starts with `./` → "sibling".
//  7. The resolver succeeded with `IsExternalLibraryImport: true`
//     OR the path lies under a directory listed in
//     `settings["import/external-module-folders"]` (default `["node_modules"]`)
//     → "external".
//  8. Resolver succeeded but not external → "internal".
//  9. Looks like a bare package name → "external".
//  10. Otherwise → "unknown".
//
// `internalRegex` is precompiled by the caller (passing nil disables that step).
// `spec` may be nil — in that case we skip resolution and fall through.
func ClassifyImport(ctx rule.RuleContext, name string, spec *ast.Node, internalRegex *regexp.Regexp) string {
	if internalRegex != nil && internalRegex.MatchString(name) {
		return "internal"
	}
	if IsAbsolutePath(name) {
		return "absolute"
	}
	if IsBuiltinModule(name, ctx.Settings) {
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

	resolvedPath, isExternal, resolved := resolveModule(ctx, spec)
	if resolved {
		if isExternal {
			return "external"
		}
		// Honour `import/external-module-folders`: even if tsgo says the path
		// resolved internally, a user-listed folder forces "external".
		if matchesExternalFolder(resolvedPath, ctx.Settings) {
			return "external"
		}
		return "internal"
	}

	if IsExternalLooking(name) {
		return "external"
	}
	return "unknown"
}

// resolveModule returns the resolved file path, the IsExternalLibraryImport
// flag, and whether resolution succeeded. Uses tsgo's own resolver so we
// inherit tsconfig `paths`, `baseUrl`, monorepo symlinks, and exports/imports
// maps. Returns (_, false, false) when the program/spec is nil or unresolvable.
func resolveModule(ctx rule.RuleContext, spec *ast.Node) (string, bool, bool) {
	if spec == nil || ctx.Program == nil || ctx.SourceFile == nil {
		return "", false, false
	}
	mod := ctx.Program.GetResolvedModuleFromModuleSpecifier(ctx.SourceFile, spec)
	if mod == nil || mod.ResolvedFileName == "" {
		return "", false, false
	}
	return mod.ResolvedFileName, mod.IsExternalLibraryImport, true
}

// matchesExternalFolder returns true if path lives under any folder listed in
// `settings["import/external-module-folders"]`. The settings entry is treated
// as path segments to be searched anywhere in the resolved absolute path —
// matching upstream's permissive behaviour for hoisted/symlinked deps.
func matchesExternalFolder(path string, settings map[string]any) bool {
	if path == "" || settings == nil {
		return false
	}
	folders, ok := settings["import/external-module-folders"].([]any)
	if !ok {
		return false
	}
	normalized := strings.ReplaceAll(path, `\`, "/")
	for _, f := range folders {
		s, ok := f.(string)
		if !ok || s == "" {
			continue
		}
		clean := strings.TrimSuffix(strings.TrimPrefix(s, "/"), "/")
		if clean == "" {
			continue
		}
		if strings.Contains(normalized, "/"+clean+"/") || strings.HasPrefix(normalized, clean+"/") {
			return true
		}
	}
	return false
}
