package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func Resolve(moduleSpecifier *ast.StringLiteralLike, ctx rule.RuleContext) (string, bool) {
	return ResolveFromSourceFile(ctx, ctx.SourceFile, moduleSpecifier)
}

// ResolveFromSourceFile resolves a module specifier to the file it names, from
// the perspective of the provided origin file.
func ResolveFromSourceFile(ctx rule.RuleContext, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, bool) {
	resolvedPath, _, ok := resolveModule(ctx, sourceFile, moduleSpecifier)
	return resolvedPath, ok
}

// ResolveSourceFileFromSourceFile resolves a module specifier and returns the
// Program's source file for the resolved path.
func ResolveSourceFileFromSourceFile(ctx rule.RuleContext, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, *ast.SourceFile, bool) {
	resolvedPath, target, ok := resolveModule(ctx, sourceFile, moduleSpecifier)
	if !ok || target == nil {
		return "", nil, false
	}
	return resolvedPath, target, true
}

// resolveModule resolves a module specifier the way eslint-plugin-import's
// resolver does, and reports the Program's source file for the result when it
// has one.
//
// TypeScript's own resolution is the first source of truth, but it only covers
// specifiers TypeScript collected while building the Program: every `import`
// and `export`, plus `import()` and — in JavaScript files only — `require()`
// calls written as a bare `require` identifier with exactly one argument. A
// `require('./x')` in a TypeScript file, or a `(require)('./x')` anywhere, is
// therefore absent from that cache even though upstream resolves both. Those
// fall through to a probe of the relative specifier against the files already
// loaded into the Program, which also covers the extension-substitution cases
// upstream still treats as edges. A package specifier carries no relative path
// to probe, so a `require('some-package')` in a TypeScript file stays unresolved.
func resolveModule(ctx rule.RuleContext, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, *ast.SourceFile, bool) {
	if ctx.Program == nil || sourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return "", nil, false
	}

	resolvedPath := ""
	if module := ctx.Program.GetResolvedModuleFromModuleSpecifier(sourceFile, moduleSpecifier); module != nil {
		resolvedPath = module.ResolvedFileName
	}
	if resolvedPath != "" {
		if target := ctx.Program.GetSourceFileForResolvedModule(resolvedPath); target != nil {
			return resolvedPath, target, true
		}
	}

	if target := resolveRelativeFromLoadedFiles(ctx, sourceFile, moduleSpecifier.Text()); target != nil {
		return target.FileName(), target, true
	}

	// TypeScript named a file the Program never loaded, such as one outside the
	// project. Callers that only need the path can still use it.
	if resolvedPath != "" {
		return resolvedPath, nil, true
	}
	return "", nil, false
}

// resolveRelativeFromLoadedFiles finds the file a relative specifier names by
// probing the candidates TypeScript's own resolution would try, against the
// files the Program has already loaded.
func resolveRelativeFromLoadedFiles(ctx rule.RuleContext, sourceFile *ast.SourceFile, specifier string) *ast.SourceFile {
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
		if target := ctx.Program.GetSourceFile(candidate); target != nil {
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
