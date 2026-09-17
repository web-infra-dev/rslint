// cspell:ignore jsxdev
package nodeutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Keep both directions: resolution can try several source extensions, while
// emitted import paths use the last configured mapping for each source extension.
type extensionMapping struct {
	forward map[string]string
	aliases map[string][]string
}

var (
	preservedExtensions = typescriptExtensions(core.JsxEmitPreserve)
	emittedExtensions   = typescriptExtensions(core.JsxEmitReact)
)

func typescriptExtensions(jsx core.JsxEmit) extensionMapping {
	options := &core.CompilerOptions{Jsx: jsx}
	mapping := extensionMapping{map[string]string{"": ".js"}, map[string][]string{}}
	for _, extension := range tspath.SupportedTSImplementationExtensions {
		emitted := module.TryGetJSExtensionForFile("index"+extension, options)
		mapping.forward[extension] = emitted
		mapping.aliases[emitted] = append(mapping.aliases[emitted], extension)
	}
	return mapping
}

func importExtensionMapping(ctx rule.RuleContext, options map[string]any) extensionMapping {
	p := ctx.Program()
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = p.CurrentDirectory()
	}
	// Upstream chooses settings.n as a whole before settings.node.
	sharedValue := ctx.Settings["n"]
	if sharedValue == nil {
		sharedValue = ctx.Settings["node"]
	}
	shared, _ := sharedValue.(map[string]any)
	for _, value := range []map[string]any{options, shared} {
		if mapping, ok := configuredExtensions(p, cwd, value); ok {
			return mapping
		}
	}
	if mapping, ok := compilerExtensions(nearestCompilerOptions(p, ctx.SourceFile.FileName())); ok {
		return mapping
	}
	return preservedExtensions
}

func compilerExtensions(config *core.CompilerOptions) (extensionMapping, bool) {
	if config != nil {
		if config.AllowImportingTsExtensions == core.TSTrue {
			return extensionMapping{}, true
		}
		if config.Jsx == core.JsxEmitPreserve {
			return preservedExtensions, true
		}
		if config.Jsx != core.JsxEmitNone {
			return emittedExtensions, true
		}
	}
	return extensionMapping{}, false
}

func configuredExtensions(p *program.Program, cwd string, options map[string]any) (extensionMapping, bool) {
	if pairs, ok := options["typescriptExtensionMap"].([]any); ok {
		mapping := extensionMapping{map[string]string{}, map[string][]string{}}
		for _, pair := range pairs {
			values := stringArray(pair)
			if len(values) == 2 {
				mapping.forward[values[0]] = values[1]
				if values[0] != "" {
					mapping.aliases[values[1]] = append(mapping.aliases[values[1]], values[0])
				}
			}
		}
		return mapping, true
	}
	preset, _ := options["typescriptExtensionMap"].(string)
	switch preset {
	case "preserve":
		return preservedExtensions, true
	case "react", "react-jsx", "react-jsxdev", "react-native":
		return emittedExtensions, true
	}
	if configPath, ok := options["tsconfigPath"].(string); ok && configPath != "" {
		return compilerExtensions(readCompilerOptions(p, tspath.ResolvePath(cwd, configPath)))
	}
	return extensionMapping{}, false
}

// ImportExtension maps a target's source extension to the emitted import
// extension. Like upstream, JavaScript importers keep the original extension.
func ImportExtension(ctx rule.RuleContext, extension string) string {
	if tspath.HasTSFileExtension(ctx.SourceFile.FileName()) {
		if emitted, ok := importExtensionMapping(ctx, nil).forward[extension]; ok {
			return emitted
		}
	}
	return extension
}
