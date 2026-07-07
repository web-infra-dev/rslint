package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func Resolve(moduleSpecifier *ast.StringLiteralLike, ctx rule.RuleContext) (string, bool) {
	return ResolveFromSourceFile(ctx, ctx.SourceFile, moduleSpecifier)
}

// ResolveFromSourceFile resolves a module specifier using TypeScript's cached
// resolution result for the provided origin file.
func ResolveFromSourceFile(ctx rule.RuleContext, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, bool) {
	if ctx.Program == nil || sourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return "", false
	}

	module := ctx.Program.GetResolvedModuleFromModuleSpecifier(sourceFile, moduleSpecifier)
	if module != nil && module.ResolvedFileName != "" {
		return module.ResolvedFileName, true
	}

	return "", false
}

// ResolveSourceFileFromSourceFile resolves a module specifier and returns the
// source file TypeScript associated with the resolved path.
func ResolveSourceFileFromSourceFile(ctx rule.RuleContext, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, *ast.SourceFile, bool) {
	resolvedPath, ok := ResolveFromSourceFile(ctx, sourceFile, moduleSpecifier)
	if !ok {
		return "", nil, false
	}

	target := ctx.Program.GetSourceFileForResolvedModule(resolvedPath)
	if target == nil {
		return "", nil, false
	}
	return resolvedPath, target, true
}

// ResolveModuleReferenceFromSourceFile resolves lint graph edges. It first uses
// TypeScript resolution, then falls back to already-loaded relative source files
// for extension-substitution cases eslint-plugin-import still treats as edges.
func ResolveModuleReferenceFromSourceFile(ctx rule.RuleContext, sourceFile *ast.SourceFile, moduleSpecifier *ast.StringLiteralLike) (string, *ast.SourceFile, bool) {
	if ctx.Program == nil || sourceFile == nil || moduleSpecifier == nil || !ast.IsStringLiteralLike(moduleSpecifier) {
		return "", nil, false
	}

	if resolvedPath, target, ok := ResolveSourceFileFromSourceFile(ctx, sourceFile, moduleSpecifier); ok {
		return resolvedPath, target, true
	}

	specifier := moduleSpecifier.Text()
	if specifier == "" || !tspath.IsExternalModuleNameRelative(specifier) {
		return "", nil, false
	}

	basePath := specifier
	if tspath.PathIsAbsolute(basePath) {
		basePath = tspath.NormalizePath(basePath)
	} else {
		basePath = tspath.ResolvePath(tspath.GetDirectoryPath(sourceFile.FileName()), specifier)
	}

	for _, candidate := range moduleResolutionFallbackCandidates(basePath) {
		if target := ctx.Program.GetSourceFile(candidate); target != nil {
			return target.FileName(), target, true
		}
	}

	return "", nil, false
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
