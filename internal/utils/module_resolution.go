package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// ModuleResolutionRuntime is the non-type-aware subset of a TypeScript
// Program needed to resolve module references. Both real Programs and the
// standalone CLI gap runtime implement it.
type ModuleResolutionRuntime interface {
	Options() *core.CompilerOptions
	GetModeForUsageLocation(sourceFile ast.HasFileName, location *ast.StringLiteralLike) core.ResolutionMode
	GetResolvedModule(sourceFile ast.HasFileName, moduleReference string, mode core.ResolutionMode) *module.ResolvedModule
	ResolveModuleName(moduleName string, containingFile string, mode core.ResolutionMode) *module.ResolvedModule
	GetSourceFileForResolvedModule(fileName string) *ast.SourceFile
	GetSourceFile(fileName string) *ast.SourceFile
}

// ResolveModulePath resolves a module specifier to the path of the file it
// names, from the perspective of the provided origin file.
func ResolveModulePath(runtime ModuleResolutionRuntime, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, bool) {
	resolvedPath, _, ok := resolveModule(runtime, sourceFile, moduleSpecifier)
	return resolvedPath, ok
}

// ResolveModuleFile resolves a module specifier and returns the runtime's
// source file for the resolved path.
func ResolveModuleFile(runtime ModuleResolutionRuntime, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, *ast.SourceFile, bool) {
	resolvedPath, target, ok := resolveModule(runtime, sourceFile, moduleSpecifier)
	if !ok || target == nil {
		return "", nil, false
	}
	return resolvedPath, target, true
}

// resolveModule resolves a module specifier the way eslint-plugin-import's
// resolver does, and reports the runtime's source file for the result when it
// has one.
//
// TypeScript's own resolution is the source of truth, but the runtime's cache
// only covers specifiers TypeScript collected while building it: every `import`
// and `export`, plus `import()` and — in JavaScript files only — `require()`
// calls written as a bare `require` identifier with exactly one argument. A
// `require('./x')` in a TypeScript file, or a `(require)('./x')` anywhere, is
// therefore absent from that cache even though upstream resolves both, so those
// run through TypeScript's resolver here instead, under the same compiler
// options the runtime uses. Only a specifier TypeScript resolves
// nowhere falls back to probing the relative specifier against the files
// already in the runtime, which covers the extension-substitution cases
// upstream still treats as edges.
func resolveModule(runtime ModuleResolutionRuntime, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, *ast.SourceFile, bool) {
	if runtime == nil || sourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return "", nil, false
	}

	specifier := moduleSpecifier.Text()
	mode := resolutionMode(runtime, sourceFile, moduleSpecifier)

	resolvedPath := ""
	if cached := runtime.GetResolvedModule(sourceFile, specifier, mode); cached != nil {
		resolvedPath = cached.ResolvedFileName
	} else {
		// A missing cache entry means TypeScript never collected this specifier,
		// rather than that it tried and failed, so resolving it costs a walk
		// TypeScript has not already done. Specifiers TypeScript did try and fail on
		// stay failed.
		resolvedPath = resolveWithModuleResolver(runtime, sourceFile, specifier, mode)
	}

	if resolvedPath != "" {
		if target := runtime.GetSourceFileForResolvedModule(resolvedPath); target != nil {
			return resolvedPath, target, true
		}
		// TypeScript named a file the runtime never loaded, such as one outside the
		// project. Callers that only need the path can still use it.
		return resolvedPath, nil, true
	}

	if target := resolveRelativeFromLoadedFiles(runtime, sourceFile, specifier); target != nil {
		return target.FileName(), target, true
	}

	return "", nil, false
}

// resolutionMode answers with the mode a specifier resolves under. TypeScript
// reads a call as a `require` only when its callee is written as a bare
// `require` identifier, and answers `(require)('pkg')` with the format of the
// file holding it instead. The module accessor reads both spellings as a
// `require`, so an ES module's parenthesized call resolves as CommonJS here too,
// and a package's `require` condition stays selected for both.
func resolutionMode(runtime ModuleResolutionRuntime, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) core.ResolutionMode {
	mode := runtime.GetModeForUsageLocation(sourceFile, moduleSpecifier)
	if mode == core.ResolutionModeESM && isRequireCall(moduleSpecifier.Parent) {
		return core.ResolutionModeCommonJS
	}
	return mode
}

func isRequireCall(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}

	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return false
	}

	callee := ast.SkipParentheses(call.Expression)
	return ast.IsIdentifier(callee) && callee.Text() == "require"
}

// resolveWithModuleResolver runs TypeScript's own module resolution for a
// specifier the runtime never collected, so that it answers under the same
// compiler options — `moduleSuffixes`, `paths`, `rootDirs` — every specifier
// TypeScript did collect was resolved under. It goes through the runtime's own
// resolver, so a specifier several rules ask about is walked once.
func resolveWithModuleResolver(runtime ModuleResolutionRuntime, sourceFile *ast.SourceFile, specifier string, mode core.ResolutionMode) string {
	resolved := runtime.ResolveModuleName(specifier, sourceFile.FileName(), mode)
	if resolved == nil || !resolved.IsResolved() {
		return ""
	}
	return resolved.ResolvedFileName
}

// resolveRelativeFromLoadedFiles finds the file a relative specifier names by
// probing the candidates TypeScript's own resolution would try, against the
// files the runtime has already loaded.
func resolveRelativeFromLoadedFiles(runtime ModuleResolutionRuntime, sourceFile *ast.SourceFile, specifier string) *ast.SourceFile {
	if specifier == "" || !tspath.IsExternalModuleNameRelative(specifier) {
		return nil
	}

	basePath := specifier
	if tspath.PathIsAbsolute(basePath) {
		basePath = tspath.NormalizePath(basePath)
	} else {
		basePath = tspath.ResolvePath(tspath.GetDirectoryPath(sourceFile.FileName()), specifier)
	}

	for _, candidate := range moduleResolutionFallbackCandidates(basePath) {
		if target := runtime.GetSourceFile(candidate); target != nil {
			return target
		}
	}

	return nil
}

func moduleResolutionFallbackCandidates(basePath string) []string {
	ext := tspath.TryGetExtensionFromPath(basePath)
	if ext != "" {
		candidates := []string{basePath}
		withoutExt := tspath.RemoveFileExtension(basePath)
		switch ext {
		case tspath.ExtensionJs:
			candidates = append(candidates, withoutExt+tspath.ExtensionTs, withoutExt+tspath.ExtensionTsx, withoutExt+tspath.ExtensionDts)
		case tspath.ExtensionJsx:
			candidates = append(candidates, withoutExt+tspath.ExtensionTsx, withoutExt+tspath.ExtensionTs, withoutExt+tspath.ExtensionDts)
		case tspath.ExtensionMjs:
			candidates = append(candidates, withoutExt+tspath.ExtensionMts, withoutExt+tspath.ExtensionDmts, withoutExt+tspath.ExtensionTs)
		case tspath.ExtensionCjs:
			candidates = append(candidates, withoutExt+tspath.ExtensionCts, withoutExt+tspath.ExtensionDcts, withoutExt+tspath.ExtensionTs)
		}
		return candidates
	}

	return []string{
		basePath,
		basePath + tspath.ExtensionTs,
		basePath + tspath.ExtensionTsx,
		basePath + tspath.ExtensionMts,
		basePath + tspath.ExtensionCts,
		basePath + tspath.ExtensionJs,
		basePath + tspath.ExtensionJsx,
		basePath + tspath.ExtensionMjs,
		basePath + tspath.ExtensionCjs,
		basePath + tspath.ExtensionDts,
		basePath + tspath.ExtensionDmts,
		basePath + tspath.ExtensionDcts,
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionTs),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionTsx),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionMts),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionCts),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionJs),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionJsx),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionMjs),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionCjs),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionDts),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionDmts),
		tspath.CombinePaths(basePath, "index"+tspath.ExtensionDcts),
	}
}
