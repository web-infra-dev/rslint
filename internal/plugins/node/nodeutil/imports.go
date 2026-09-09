// cspell:ignore jsxdev
package nodeutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

var npmSpecifier = esregexp.MustCompile(`^(@[\w~-][\w.~-]*/)?[\w~-][\w.~-]*`, "")

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
	if isNodeBuiltin(resource) || strings.HasPrefix(resource, "data:") ||
		strings.HasPrefix(resource, "http://") || strings.HasPrefix(resource, "https://") || !npmSpecifier.Test(resource) {
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
// Config parsing, including extends and the nearest-file search, belongs to Program.
func HasTypeScriptAlias(p *program.Program, fileName, name string) bool {
	if !tspath.HasTSFileExtension(fileName) {
		return false
	}
	options := p.NearestCompilerOptions(fileName)
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

// ImportResolutionOptions implements the documented resolverConfig.modules
// option and the shared Node extension/lookup settings. convertPath is accepted
// by no-extraneous-import's schema but, as upstream, does not affect this check.
func ImportResolutionOptions(p *program.Program, fileName string, typeOnly bool, options, settings map[string]any) program.NodeResolutionOptions {
	result := program.NodeResolutionOptions{
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
	cwd := p.CurrentDirectory()
	if configured, ok := settings["cwd"].(string); ok {
		cwd = tspath.ResolvePath(cwd, configured)
	}
	for i, base := range result.Paths {
		result.Paths[i] = tspath.ResolvePath(cwd, base)
	}
	if tspath.HasTSFileExtension(fileName) {
		config := p.NearestCompilerOptions(fileName)
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
		if pairs, ok := shared["typescriptExtensionMap"].([]any); ok {
			result.ExtensionAliases = map[string][]string{}
			for _, pair := range pairs {
				values := stringArray(pair)
				if len(values) == 2 && values[0] != "" {
					result.ExtensionAliases[values[1]] = append(result.ExtensionAliases[values[1]], values[0])
				}
			}
		} else {
			preset, _ := shared["typescriptExtensionMap"].(string)
			switch preset {
			case "react", "react-jsx", "react-jsxdev", "react-native", "preserve":
			default:
				if configPath, ok := shared["tsconfigPath"].(string); ok && configPath != "" {
					config = p.ReadCompilerOptions(configPath)
					if config != nil && config.AllowImportingTsExtensions == core.TSTrue {
						result.ExtensionAliases = nil
					} else if config != nil {
						if config.Jsx == core.JsxEmitPreserve {
							preset = "preserve"
						} else if config.Jsx != core.JsxEmitNone {
							preset = "react"
						}
					}
				}
			}
			switch preset {
			case "react", "react-jsx", "react-jsxdev", "react-native", "preserve":
				result.ExtensionAliases = emittedExtensionAliases
				if preset == "preserve" {
					result.ExtensionAliases = preservedExtensionAliases
				}
			}
		}
	}
	return result
}
