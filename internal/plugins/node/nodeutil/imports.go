// cspell:ignore jsxdev
package nodeutil

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

var npmSpecifier = esregexp.MustCompile(`^(@[\w~-][\w.~-]*/)?[\w~-][\w.~-]*`, "")

// ImportModuleName returns the npm package root after removing loader params.
// Builtins, relative/absolute paths, import maps and URL imports have no npm name.
func ImportModuleName(specifier string) (name, resource string) {
	resource, _, _ = strings.Cut(specifier, "!")
	if core.NodeCoreModules()[resource] || strings.HasPrefix(resource, "data:") ||
		strings.HasPrefix(resource, "http://") || strings.HasPrefix(resource, "https://") || !npmSpecifier.Test(resource) {
		return "", resource
	}
	name, _ = module.ParsePackageName(resource)
	return name, resource
}

func isTypeScript(fileName string) bool {
	switch path.Ext(fileName) {
	case ".ts", ".tsx", ".mts", ".cts":
		return true
	}
	return false
}

// HasTypeScriptAlias preserves upstream's prefix exemption for compiler paths.
// Config parsing, including extends and the nearest-file search, belongs to Program.
func HasTypeScriptAlias(p *program.Program, fileName, name string) bool {
	if !isTypeScript(fileName) {
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
	if isTypeScript(fileName) {
		config := p.NearestCompilerOptions(fileName)
		if config != nil && config.AllowImportingTsExtensions == core.TSTrue {
			if result.Extensions == nil {
				result.Extensions = []string{".js", ".ts", ".mjs", ".mts", ".cjs", ".cts", ".json", ".node"}
			}
		} else {
			result.ExtensionAliases = map[string][]string{".js": {".ts"}, ".mjs": {".mts"}, ".cjs": {".cts"}, ".jsx": {".tsx"}}
			if config != nil && config.Jsx != core.JsxEmitPreserve && config.Jsx != core.JsxEmitNone {
				result.ExtensionAliases[".js"] = append(result.ExtensionAliases[".js"], ".tsx")
				delete(result.ExtensionAliases, ".jsx")
			}
		}
		// Unlike the independent shared lists, upstream selects settings.n as
		// a whole before settings.node for these extension-mapping settings.
		shared, ok := settings["n"].(map[string]any)
		if !ok {
			shared, _ = settings["node"].(map[string]any)
		}
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
				result.ExtensionAliases = map[string][]string{".js": {".ts"}, ".mjs": {".mts"}, ".cjs": {".cts"}}
				if preset == "preserve" {
					result.ExtensionAliases[".jsx"] = []string{".tsx"}
				} else {
					result.ExtensionAliases[".js"] = append(result.ExtensionAliases[".js"], ".tsx")
				}
			}
		}
	}
	return result
}
