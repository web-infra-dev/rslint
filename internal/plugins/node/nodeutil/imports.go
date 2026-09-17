// cspell:ignore jsxdev
package nodeutil

import (
	"maps"
	"path/filepath"
	"slices"
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
	return resolveImport(p, name, fileName, typeOnly, options).resolveError
}

// ImportFilePath supplies the target for absolute-path restrictions. Missing
// local imports retain their lexical path; unresolved packages have no path.
func ImportFilePath(p *program.Program, name, fileName string, typeOnly bool, options ResolutionOptions) string {
	resolved := resolveImport(p, name, fileName, typeOnly, options)
	return moduleFilePath(name, fileName, resolved)
}

// RequireFilePath uses CommonJS directory and alias resolution for restrictions.
// Missing local targets share the import rule's lexical-path fallback.
func RequireFilePath(p *program.Program, name, fileName string, options ResolutionOptions) string {
	resolved := resolveWithTypeScriptAliases(p, name, fileName, options)
	return moduleFilePath(name, fileName, resolved)
}

func moduleFilePath(name, fileName string, resolved nodeResolution) string {
	if resolved.path != "" {
		if isImportURL(name) {
			return resolved.path
		}
		// Only diagnostic matching uses host paths, never VFS lookups.
		return filepath.FromSlash(resolved.path) + resolved.resourceSuffix
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

func resolveImport(p *program.Program, name, fileName string, typeOnly bool, options ResolutionOptions) nodeResolution {
	if !typeOnly && isImportURL(name) {
		return nodeResolution{path: name}
	}
	moduleName, _ := ImportModuleName(name)
	if moduleName == "" {
		if options.MainFields == nil {
			options.MainFields = []nodeMainField{}
		}
		if options.MainFiles == nil {
			options.MainFiles = []string{}
		}
		options.NoDirectory = len(options.MainFields) == 0 && len(options.MainFiles) == 0
	}
	return resolveWithTypeScriptAliases(p, name, fileName, options)
}

func resolveWithTypeScriptAliases(p *program.Program, name, fileName string, options ResolutionOptions) nodeResolution {
	if !options.AliasesConfigured && tspath.HasTSFileExtension(fileName) {
		if config := nearestCompilerOptions(p, fileName); config != nil && config.Paths != nil {
			for name, targets := range config.Paths.Entries() {
				alias := moduleAlias{Name: strings.TrimRight(name, `/\*`)}
				for _, target := range targets {
					alias.Targets = append(alias.Targets, tspath.ResolvePath(tspath.GetDirectoryPath(config.ConfigFilePath), strings.TrimRight(target, `/\*`)))
				}
				options.Aliases = append(options.Aliases, alias)
			}
		}
	}
	return resolveModuleCached(p, name, fileName, options)
}

// ImportResolutionOptions implements the documented resolverConfig.modules
// option and the shared Node extension/lookup settings. convertPath is accepted
// by the extraneous rules' schemas but, as upstream, does not affect these checks.
func ImportResolutionOptions(ctx rule.RuleContext, typeOnly bool, options map[string]any) ResolutionOptions {
	conditions := []string{"node", "require", "import"}
	if typeOnly {
		conditions = append(conditions, "types")
	}
	return importResolutionOptions(ctx, conditions, options)
}

func importResolutionOptions(ctx rule.RuleContext, conditions []string, options map[string]any) ResolutionOptions {
	p, fileName, settings := ctx.Program(), ctx.SourceFile.FileName(), ctx.Settings
	result := ResolutionOptions{
		Extensions: StringListSetting("tryExtensions", options, settings),
		Paths:      StringListSetting("resolvePaths", options, settings),
		Conditions: conditions,
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
	for _, value := range settingValues("resolverConfig", options, settings) {
		if config, ok := value.(map[string]any); ok {
			applyResolverConfig(&result, config)
			break
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
	return importResolutionOptions(ctx, []string{"node", "require"}, options)
}

// Explicit resolver options replace the corresponding defaults, including
// TypeScript aliases and extension mappings. An empty list remains meaningful.
func applyResolverConfig(options *ResolutionOptions, config map[string]any) {
	for key, destination := range map[string]*[]string{
		"modules": &options.Modules, "extensions": &options.Extensions, "conditionNames": &options.Conditions,
		"mainFiles": &options.MainFiles,
	} {
		switch value := config[key].(type) {
		case []any, []string:
			*destination = stringArray(value)
		case string:
			if key == "modules" && value != "" {
				*destination = []string{value}
			}
		}
	}
	mainFields := config["mainFields"]
	if names, ok := mainFields.([]string); ok {
		mainFields = utils.Map(names, func(name string) any { return name })
	}
	if fields, ok := mainFields.([]any); ok {
		options.MainFields = []nodeMainField{}
		for _, value := range fields {
			field := nodeMainField{ForceRelative: true}
			if object, ok := value.(map[string]any); ok {
				value = object["name"]
				field.ForceRelative, _ = object["forceRelative"].(bool)
			}
			if name, ok := value.(string); ok {
				field.Name = []string{name}
			} else {
				field.Name = stringArray(value)
			}
			if len(field.Name) != 0 {
				options.MainFields = append(options.MainFields, field)
			}
		}
	}
	aliasFields := config["aliasFields"]
	if names, ok := aliasFields.([]string); ok {
		aliasFields = utils.Map(names, func(name string) any { return name })
	}
	if fields, ok := aliasFields.([]any); ok {
		options.AliasFields = nil
		for _, value := range fields {
			if name, ok := value.(string); ok {
				options.AliasFields = append(options.AliasFields, []string{name})
			} else if names := stringArray(value); len(names) != 0 {
				options.AliasFields = append(options.AliasFields, names)
			}
		}
	}
	if aliases, ok := config["extensionAlias"].(map[string]any); ok {
		options.ExtensionAliases = make(map[string][]string, len(aliases))
		for extension, targets := range aliases {
			if target, ok := targets.(string); ok {
				options.ExtensionAliases[extension] = []string{target}
			} else {
				options.ExtensionAliases[extension] = stringArray(targets)
			}
		}
	}
	value, present := config["alias"]
	options.AliasesConfigured = present
	appendAlias := func(name string, targets any, exact bool) {
		alias := moduleAlias{Name: name, OnlyModule: exact}
		switch targets := targets.(type) {
		case string:
			alias.Targets = []string{targets}
		case []any, []string:
			alias.Targets = stringArray(targets)
		case bool:
			alias.Ignore = !targets
		}
		options.Aliases = append(options.Aliases, alias)
	}
	switch aliases := value.(type) {
	case map[string]any:
		// Settings maps have no declaration order. Array form preserves an
		// explicit priority when keys overlap; object form is deterministic.
		for _, name := range slices.Sorted(maps.Keys(aliases)) {
			key, exact := strings.CutSuffix(name, "$")
			appendAlias(key, aliases[name], exact)
		}
	case []any:
		for _, entry := range aliases {
			if alias, ok := entry.(map[string]any); ok {
				if name, ok := alias["name"].(string); ok {
					exact, _ := alias["onlyModule"].(bool)
					appendAlias(name, alias["alias"], exact)
				}
			}
		}
	}
}
