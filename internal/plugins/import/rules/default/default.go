package default_rule

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// defaultIgnorePatterns mirrors upstream eslint-plugin-import's
// ExportMapBuilder fallback when `settings["import/ignore"]` is unset.
// These match file shapes the upstream tool refuses to parse for an
// export map (CoffeeScript, ESLint config files, native bindings, JSON, …).
// rslint's TypeChecker can resolve some of these (e.g. `.json` with
// resolveJsonModule), but matching the upstream contract is what users
// configuring `import/ignore` expect.
var defaultIgnorePatterns = []string{
	`\.coffee$`,
	`\.eslintrc(\.[a-z]+)?$`,
	`\.(es6|exs|json|node)$`,
}

// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/default.js
var DefaultRule = rule.Rule{
	Name: "import/default",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}
		// Reuse the plugin-shared module visitor so module-specifier extraction
		// stays consistent across import rules. ExportNamedDeclaration is in
		// the visitor's listener set; the upstream rule's
		// `ExportDefaultSpecifier` branch (Babel-only `export X from "..."`)
		// is unreachable through the TS parser, so the export-side check is a
		// no-op here.
		return utils.VisitModules(func(source *ast.StringLiteralLike, node *ast.Node) {
			checkDefault(ctx, source, node)
		}, utils.VisitModulesOptions{
			ESModule: true,
		})
	},
}

// checkDefault reports when an `import X from "..."` requests a default that
// the resolved module does not actually expose.
func checkDefault(ctx rule.RuleContext, source *ast.StringLiteralLike, node *ast.Node) {
	defaultName := getDefaultSpecifier(node)
	if defaultName == nil {
		return
	}

	// Match upstream's "core / non-analyzable modules always have a default"
	// allowance: anything we cannot resolve to a project source file is
	// skipped so synthetic / declaration-only modules don't false-positive.
	resolvedPath, ok := utils.Resolve(source, ctx)
	if !ok || resolvedPath == "" {
		return
	}

	// `settings["import/ignore"]` — array of regex patterns (or upstream's
	// default list) tested against the resolved path. A match marks the
	// module as un-analyzable, mirroring `ExportMapBuilder.get(...) === null`.
	if isIgnoredPath(ctx, resolvedPath) {
		return
	}

	// External-library imports (node_modules, ambient @types) — when the user
	// has NOT overridden `import/ignore`, we keep the upstream-equivalent
	// "core modules always have a default" allowance. Once the user opts into
	// custom `import/ignore`, this fallback steps aside so they get full
	// control.
	if !hasExplicitIgnoreSetting(ctx) {
		if module := ctx.Program.GetResolvedModuleFromModuleSpecifier(ctx.SourceFile, source); module != nil && module.IsExternalLibraryImport {
			return
		}
	}

	moduleSymbol := ctx.TypeChecker.GetSymbolAtLocation(source)
	if moduleSymbol == nil {
		return
	}

	if moduleHasDefaultExport(ctx, moduleSymbol) {
		return
	}

	ctx.ReportNode(defaultName, rule.RuleMessage{
		Id:          "default",
		Description: fmt.Sprintf("No default export found in imported module \"%s\".", source.AsStringLiteral().Text),
	})
}

// hasExplicitIgnoreSetting reports whether the user has set
// `settings["import/ignore"]` to a non-nil value. An empty array is honoured
// as "do not ignore anything" — once the user takes control, we don't layer
// our own fallback. A `null` value is treated as absent (matches JSON-null
// semantics: the key is present in serialized settings but carries no value).
func hasExplicitIgnoreSetting(ctx rule.RuleContext) bool {
	if ctx.Settings == nil {
		return false
	}
	v, ok := ctx.Settings["import/ignore"]
	if !ok {
		return false
	}
	return v != nil
}

// isIgnoredPath returns true when the resolved path matches any of the
// configured ignore patterns. Mirrors upstream eslint-plugin-import's
// `ExportMapBuilder.exportMap()` which tests patterns against the resolved
// file path, NOT the source value as written in the import statement.
//
// When `import/ignore` is unset, upstream's default list is applied. Note
// that several of those defaults (`\.coffee$`, `\.(es6|exs|node)$`,
// `\.eslintrc...`) are effectively inert in rslint because TypeScript's
// resolver itself refuses to load those extensions — the rule's
// "resolve fails → skip" branch fires before the ignore check ever runs.
// `\.json$` is the one default that bites under `resolveJsonModule: true`.
func isIgnoredPath(ctx rule.RuleContext, resolvedPath string) bool {
	patterns := getIgnorePatterns(ctx)
	if len(patterns) == 0 {
		return false
	}
	for _, p := range patterns {
		if p == nil {
			continue
		}
		if p.MatchString(resolvedPath) {
			return true
		}
	}
	return false
}

// getIgnorePatterns reads `settings["import/ignore"]` as `[]string`, falling
// back to upstream's default list when unset or null. Invalid patterns are
// skipped silently to match upstream's behaviour of dropping malformed regexes.
func getIgnorePatterns(ctx rule.RuleContext) []*regexp.Regexp {
	raw := defaultIgnorePatterns
	if ctx.Settings != nil {
		if v, ok := ctx.Settings["import/ignore"]; ok && v != nil {
			parsed, applied := coerceStringSlice(v)
			if applied {
				raw = parsed
			}
		}
	}
	out := make([]*regexp.Regexp, 0, len(raw))
	for _, p := range raw {
		if re, err := regexp.Compile(p); err == nil {
			out = append(out, re)
		}
	}
	return out
}

// coerceStringSlice accepts the JSON shapes ESLint settings can deliver:
// `[]interface{}` of strings, `[]string`, or a single string. Returns the
// extracted patterns and whether any value (even an empty array) was found.
func coerceStringSlice(v any) ([]string, bool) {
	switch raw := v.(type) {
	case []string:
		return raw, true
	case []interface{}:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out, true
	case string:
		return []string{raw}, true
	}
	return nil, false
}

// getDefaultSpecifier returns the binding node carrying a default import — the
// place to anchor diagnostics — or nil when the construct does not introduce
// one. Mirrors upstream's `specifiers.find(s => s.type === 'ImportDefaultSpecifier')`.
//
// Returns nil for: bare imports, named-only imports, namespace imports,
// dynamic `import("...")`, plain `require("...")`, and `export ... from "..."`
// re-exports.
func getDefaultSpecifier(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindImportDeclaration && node.Kind != ast.KindJSImportDeclaration {
		return nil
	}
	decl := node.AsImportDeclaration()
	if decl == nil || decl.ImportClause == nil {
		return nil
	}
	clause := decl.ImportClause.AsImportClause()
	if clause == nil {
		return nil
	}
	return clause.Name()
}

// moduleHasDefaultExport reports whether the resolved module exposes a default
// export. We accept either:
//   - `default`  — ES `export default X`
//   - `export=`  — TS-only `export = X`, but only counted as default when the
//     program enables esModuleInterop or allowSyntheticDefaultImports. Without
//     either, `import X from "..."` against a CJS-shaped module is a
//     compile-time error in TypeScript and upstream eslint-plugin-import
//     reports it as missing default.
//
// Each candidate entry is followed through SkipAlias because re-exports such
// as `export { default } from "./other"` register a `default` entry on the
// re-exporting module's own export table even when the source module lacks a
// default — SkipAlias collapses the chain to the unknown symbol in that case,
// matching upstream's broken-trampoline semantics.
func moduleHasDefaultExport(ctx rule.RuleContext, moduleSymbol *ast.Symbol) bool {
	if findResolvedExport(ctx, moduleSymbol, ast.InternalSymbolNameDefault) {
		return true
	}
	if !canSynthesizeDefaultFromExportEquals(ctx) {
		return false
	}
	return findResolvedExport(ctx, moduleSymbol, ast.InternalSymbolNameExportEquals)
}

// canSynthesizeDefaultFromExportEquals returns true when the program's
// compiler options permit treating `export = X` as a default import target.
// TypeScript turns either `esModuleInterop: true` or
// `allowSyntheticDefaultImports: true` into "you can `import X from` a CJS
// module"; without either, the default specifier doesn't bind.
func canSynthesizeDefaultFromExportEquals(ctx rule.RuleContext) bool {
	opts := ctx.Program.Options()
	if opts == nil {
		return true // defensive: behave as before when options are unavailable
	}
	return opts.ESModuleInterop.IsTrue() || opts.AllowSyntheticDefaultImports.IsTrue()
}

func findResolvedExport(ctx rule.RuleContext, moduleSymbol *ast.Symbol, name string) bool {
	if moduleSymbol == nil || moduleSymbol.Exports == nil {
		return false
	}
	entry, ok := moduleSymbol.Exports[name]
	if !ok || entry == nil {
		return false
	}
	if !isAliasSymbol(entry) {
		return true
	}
	resolved := ctx.TypeChecker.SkipAlias(entry)
	if resolved == nil {
		return false
	}
	return !ctx.TypeChecker.IsUnknownSymbol(resolved)
}

func isAliasSymbol(symbol *ast.Symbol) bool {
	return symbol.Flags&ast.SymbolFlagsAlias != 0
}
