// cspell:ignore jsxdev
package nodeutil

import (
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

var npmSpecifier = esregexp.MustCompile(`^(@[\w~-][\w.~-]*/)?[\w~-][\w.~-]*`, "")

// ImportVisitorOptions selects which literal imports the rule checks.
// IgnoreTypeImport applies to import declarations, not type-only re-exports.
type ImportVisitorOptions struct {
	IncludeCore      bool
	IgnoreTypeImport bool
}

// VisitImports shares the literal import/export shapes used by the Node rules.
func VisitImports(options ImportVisitorOptions, check func(*ast.Node, string, bool)) rule.RuleListeners {
	visit := func(source *ast.Node, typeOnly bool) {
		if source == nil {
			return
		}
		switch source.Kind {
		case ast.KindStringLiteral, ast.KindBigIntLiteral, ast.KindNumericLiteral,
			ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindRegularExpressionLiteral:
			// Only ESTree literals qualify. The shared helper preserves JS
			// number rounding, BigInt values and canonical regexp flags.
		default:
			return
		}
		specifier, _ := utils.GetStaticExpressionValue(source)
		specifier, _, _ = strings.Cut(specifier, "!")
		if !options.IncludeCore && isNodeBuiltin(specifier) {
			return
		}
		check(source, specifier, typeOnly)
	}
	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			typeOnly := declaration.ImportClause != nil && declaration.ImportClause.AsImportClause().IsTypeOnly()
			if !options.IgnoreTypeImport || !typeOnly {
				visit(declaration.ModuleSpecifier, typeOnly)
			}
		},
		ast.KindExportDeclaration: func(node *ast.Node) {
			declaration := node.AsExportDeclaration()
			visit(declaration.ModuleSpecifier, declaration.IsTypeOnly)
		},
		ast.KindCallExpression: func(node *ast.Node) {
			call := node.AsCallExpression()
			if call.Expression.Kind == ast.KindImportKeyword && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				// ESTree strips parentheses, but retains templates and TS wrappers.
				visit(utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0]), false)
			}
		},
	}
}

// isImportURL identifies the URL imports exempted by eslint-plugin-n.
func isImportURL(specifier string) bool {
	return strings.HasPrefix(specifier, "data:") || strings.HasPrefix(specifier, "http://") || strings.HasPrefix(specifier, "https://")
}

// These inverse maps are immutable. tsgo owns the emitted extension table;
// the Node plugin selects preserve mode when no JSX setting is configured.
var (
	preservedExtensionAliases = typescriptExtensionAliases(core.JsxEmitPreserve)
	emittedExtensionAliases   = typescriptExtensionAliases(core.JsxEmitReact)
)

func typescriptExtensionAliases(jsx core.JsxEmit) map[string][]string {
	options := &core.CompilerOptions{Jsx: jsx}
	aliases := map[string][]string{}
	for _, extension := range tspath.SupportedTSImplementationExtensions {
		emitted := module.TryGetJSExtensionForFile("index"+extension, options)
		aliases[emitted] = append(aliases[emitted], extension)
	}
	return aliases
}

// ImportModuleName returns the npm package root after removing loader params.
// Builtins, relative/absolute paths, import maps and URL imports have no npm name.
func ImportModuleName(specifier string) (name, resource string) {
	resource, _, _ = strings.Cut(specifier, "!")
	if strings.HasPrefix(resource, ".") || strings.HasPrefix(resource, "/") || strings.HasPrefix(resource, `\`) || isNodeBuiltin(resource) || isImportURL(resource) || !npmSpecifier.Test(resource) {
		return "", resource
	}
	name, _ = module.ParsePackageName(resource)
	return name, resource
}

func isNodeBuiltin(specifier string) bool {
	if core.NodeCoreModules()[specifier] {
		return true
	}
	// tsgo intentionally filters out underscore-prefixed internal modules.
	// Node's isBuiltin still includes these legacy names (verified on Node 22).
	switch strings.TrimPrefix(specifier, "node:") {
	case "_http_agent", "_http_client", "_http_common", "_http_incoming", "_http_outgoing", "_http_server",
		"_stream_duplex", "_stream_passthrough", "_stream_readable", "_stream_transform", "_stream_wrap", "_stream_writable",
		"_tls_common", "_tls_wrap":
		return true
	}
	return false
}

// HasTypeScriptAlias preserves upstream's prefix exemption for compiler paths.
// Config parsing reuses tsgo and reads through the Program's filesystem.
func HasTypeScriptAlias(p *program.Program, fileName, name string) bool {
	if !tspath.HasTSFileExtension(fileName) {
		return false
	}
	options := nearestCompilerOptions(p, fileName)
	if options == nil || options.Paths == nil {
		return false
	}
	for alias := range options.Paths.Keys() {
		if strings.HasPrefix(name, strings.TrimRight(alias, `/\*`)) {
			return true
		}
	}
	return false
}

// ImportResolveError checks actual targets, including TypeScript path aliases.
// Extraneous-dependency rules intentionally use HasTypeScriptAlias instead:
// their upstream contract exempts an alias even when its target is missing.
func ImportResolveError(p *program.Program, name, fileName string, typeOnly bool, options ResolutionOptions) string {
	_, resolveError := resolveImport(p, name, fileName, typeOnly, options)
	return resolveError
}

// ImportFilePath supplies the target for absolute-path restrictions. Missing
// local imports retain their lexical path; unresolved packages have no path.
func ImportFilePath(p *program.Program, name, fileName string, typeOnly bool, options ResolutionOptions) string {
	resolved, _ := resolveImport(p, name, fileName, typeOnly, options)
	return moduleFilePath(name, fileName, resolved)
}

// RequireFilePath uses CommonJS directory and alias resolution for restrictions.
// Missing local targets share the import rule's lexical-path fallback.
func RequireFilePath(p *program.Program, name, fileName string, options ResolutionOptions) string {
	if isNodeBuiltin(name) {
		return ""
	}
	resolved, _ := resolveWithTypeScriptAliases(p, name, fileName, options)
	return moduleFilePath(name, fileName, resolved)
}

func moduleFilePath(name, fileName, resolved string) string {
	if resolved != "" {
		if isImportURL(name) {
			return resolved
		}
		// The resolver uses tsgo paths; upstream matches host filesystem paths.
		resolved = filepath.FromSlash(resolved)
		// Runtime lookup removes resource queries/fragments, but restrictions
		// compare the complete resource returned by enhanced-resolve.
		if index := strings.IndexAny(name, "?#"); index > 0 {
			resolved += name[index:]
		}
		return resolved
	}
	if tspath.PathIsRelative(name) || strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) {
		// Only the lexical fallback uses host paths, never VFS lookup keys.
		// A rooted Windows path without a drive inherits the importer's volume.
		directory := filepath.Dir(filepath.FromSlash(fileName))
		if filepath.Separator == '\\' && filepath.VolumeName(name) == "" && !tspath.PathIsRelative(name) {
			name = filepath.VolumeName(directory) + name
		}
		if IsAbsolutePath(name) {
			return filepath.Clean(name)
		}
		return filepath.Join(directory, name)
	}
	return ""
}

func resolveImport(p *program.Program, name, fileName string, typeOnly bool, options ResolutionOptions) (string, string) {
	if isNodeBuiltin(name) {
		return "", ""
	}
	if !typeOnly && isImportURL(name) {
		return name, ""
	}
	moduleName, _ := ImportModuleName(name)
	options.NoDirectory = moduleName == ""
	return resolveWithTypeScriptAliases(p, name, fileName, options)
}

func resolveWithTypeScriptAliases(p *program.Program, name, fileName string, options ResolutionOptions) (string, string) {
	if tspath.HasTSFileExtension(fileName) {
		if config := nearestCompilerOptions(p, fileName); config != nil && config.Paths != nil {
			for alias, targets := range config.Paths.Entries() {
				alias = strings.TrimRight(alias, `/\*`)
				pattern := core.TryParsePattern(alias)
				wildcard := pattern.IsValid() && pattern.StarIndex >= 0 && pattern.Matches(name)
				if !wildcard && name != alias && !strings.HasPrefix(name, alias+"/") {
					continue
				}
				for _, target := range targets {
					target = strings.TrimRight(target, `/\*`)
					if wildcard {
						target = strings.Replace(target, "*", pattern.MatchedText(name), 1)
					} else {
						target += strings.TrimPrefix(name, alias)
					}
					target = tspath.ResolvePath(tspath.GetDirectoryPath(config.ConfigFilePath), target)
					if resolved := ResolveModule(p, target, fileName, options); resolved != "" {
						return resolved, ""
					}
				}
				return "", "Can't resolve '" + name + "' in '" + tspath.GetDirectoryPath(fileName) + "'"
			}
		}
	}
	return ResolveModuleWithError(p, name, fileName, options)
}

// ImportResolutionOptions implements the documented resolverConfig.modules
// option and the shared Node extension/lookup settings. convertPath is accepted
// by the extraneous rules' schemas but, as upstream, does not affect these checks.
func ImportResolutionOptions(ctx rule.RuleContext, typeOnly bool, options map[string]any) ResolutionOptions {
	p, fileName, settings := ctx.Program(), ctx.SourceFile.FileName(), ctx.Settings
	result := ResolutionOptions{
		Extensions: StringListSetting("tryExtensions", options, settings),
		Paths:      StringListSetting("resolvePaths", options, settings),
		Conditions: []string{"node", "require", "import"},
	}
	if typeOnly {
		result.Conditions = append(result.Conditions, "types")
	}
	for _, value := range settingValues("resolverConfig", options, settings) {
		if config, ok := value.(map[string]any); ok {
			result.Modules = stringArray(config["modules"])
			if directory, ok := config["modules"].(string); ok && directory != "" {
				result.Modules = []string{directory}
			}
			break
		}
	}
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = p.CurrentDirectory()
	}
	processDirectory := cwd
	if configured, ok := settings["cwd"].(string); ok {
		cwd = tspath.ResolvePath(cwd, configured)
	}
	for i, base := range result.Paths {
		result.Paths[i] = tspath.ResolvePath(cwd, base)
	}
	if tspath.HasTSFileExtension(fileName) {
		config := nearestCompilerOptions(p, fileName)
		if config != nil && config.AllowImportingTsExtensions == core.TSTrue {
			if result.Extensions == nil {
				result.Extensions = []string{".js", ".ts", ".mjs", ".mts", ".cjs", ".cts", ".json", ".node"}
			}
		} else {
			result.ExtensionAliases = preservedExtensionAliases
			if config != nil && config.Jsx != core.JsxEmitPreserve && config.Jsx != core.JsxEmitNone {
				result.ExtensionAliases = emittedExtensionAliases
			}
		}
		// Unlike the independent shared lists, upstream selects settings.n as
		// a whole before settings.node for these extension-mapping settings.
		sharedValue := settings["n"]
		if sharedValue == nil {
			sharedValue = settings["node"]
		}
		shared, _ := sharedValue.(map[string]any)
		for _, value := range []map[string]any{options, shared} {
			if aliases, ok := configuredExtensionAliases(p, processDirectory, value); ok {
				result.ExtensionAliases = aliases
				break
			}
		}
	}
	return result
}

func configuredExtensionAliases(p *program.Program, cwd string, options map[string]any) (map[string][]string, bool) {
	if pairs, ok := options["typescriptExtensionMap"].([]any); ok {
		aliases := map[string][]string{}
		for _, pair := range pairs {
			values := stringArray(pair)
			if len(values) == 2 && values[0] != "" {
				aliases[values[1]] = append(aliases[values[1]], values[0])
			}
		}
		return aliases, true
	}
	preset, _ := options["typescriptExtensionMap"].(string)
	switch preset {
	case "preserve":
		return preservedExtensionAliases, true
	case "react", "react-jsx", "react-jsxdev", "react-native":
		return emittedExtensionAliases, true
	}
	if configPath, ok := options["tsconfigPath"].(string); ok && configPath != "" {
		if config := readCompilerOptions(p, tspath.ResolvePath(cwd, configPath)); config != nil {
			if config.AllowImportingTsExtensions == core.TSTrue {
				return nil, true
			}
			if config.Jsx == core.JsxEmitPreserve {
				return preservedExtensionAliases, true
			}
			if config.Jsx != core.JsxEmitNone {
				return emittedExtensionAliases, true
			}
		}
	}
	return nil, false
}

// RequireResolutionOptions uses the same lookup and TypeScript settings as
// imports, but CommonJS does not activate the import or types export conditions.
func RequireResolutionOptions(ctx rule.RuleContext, options map[string]any) ResolutionOptions {
	result := ImportResolutionOptions(ctx, false, options)
	result.Conditions = []string{"node", "require"}
	return result
}
