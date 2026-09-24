package extensions

import (
	_ "embed"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

//go:embed extensions.schema.json
var schemaJSON []byte

var querySuffix = esregexp.MustCompile(`\?(.*)$`, "")

type pathOverride struct {
	matcher *minimatch3.Matcher
	action  string
}

type options struct {
	defaultMode      string
	pattern          map[string]string
	ignorePackages   bool
	checkTypeImports bool
	overrides        []pathOverride
}

func parseOptions(raw []any) options {
	// Extension maps use own keys, without upstream's accidental exemptions for
	// Object.prototype names such as "constructor"; see Differences from upstream.
	opts := options{defaultMode: "never", pattern: map[string]string{}}
	for _, value := range raw {
		if mode, ok := value.(string); ok {
			opts.defaultMode = mode
			continue
		}
		object, _ := value.(map[string]any)
		pattern := object
		if object["pattern"] != nil || object["ignorePackages"] != nil || object["checkTypeImports"] != nil {
			pattern, _ = object["pattern"].(map[string]any)
			opts.ignorePackages, _ = object["ignorePackages"].(bool)
			opts.checkTypeImports, _ = object["checkTypeImports"].(bool)
			if overrides, ok := object["pathGroupOverrides"].([]any); ok {
				for _, entry := range overrides {
					group, _ := entry.(map[string]any)
					glob, _ := group["pattern"].(string)
					action, _ := group["action"].(string)
					opts.overrides = append(opts.overrides, pathOverride{
						matcher: import_utils.NewPathGroupMatcher(glob, group["patternOptions"]), action: action,
					})
				}
			}
		}
		// Upstream treats an object containing only pathGroupOverrides as the
		// legacy extension map. Keep that distinction despite the permissive schema.
		for extension, mode := range pattern {
			if mode, ok := mode.(string); ok {
				opts.pattern[extension] = mode
			}
		}
	}
	if opts.defaultMode == "ignorePackages" {
		opts.defaultMode, opts.ignorePackages = "always", true
	}
	return opts
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/extensions.js
// Upstream uses literal messages and provides no fixes or suggestions.
var ExtensionsRule = rule.Rule{
	Name:   "import/extensions",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, raw []any) rule.RuleListeners {
		opts := parseOptions(raw)
		settings := import_utils.SettingsFor(ctx)
		var resolver *import_utils.ImportResolver
		reportedResolverError := false
		resolve := func(source *ast.Node, name string) (string, bool) {
			if resolver == nil {
				resolver = import_utils.NewImportResolver(ctx)
			}
			path, found, resolveError := resolver.ResolveName(name, source)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			return path, found
		}
		return import_utils.VisitModules(func(source *ast.Node, declaration *ast.Node) {
			written := source.Text()
			if written == "" {
				return
			}
			action := ""
			for _, override := range opts.overrides {
				if override.matcher.Match(written) {
					action = override.action
					break
				}
			}
			if action == "ignore" || action == "" && settings.IsCoreModuleSpecifier(written) {
				return
			}
			name := written
			if strings.ContainsRune(written, '?') {
				name, _ = querySuffix.ReplaceFirst(written, "")
			}
			if action == "" && isExternalRoot(name) {
				return
			}
			resolved, found := resolve(source, name)
			extension := fileExtension(resolved)
			if resolved == "" {
				extension = fileExtension(name)
			}
			mode := opts.pattern[extension]
			if mode == "" {
				mode = opts.defaultMode
			}
			if extension == "" || !strings.HasSuffix(name, "."+extension) {
				if mode != "always" || !opts.checkTypeImports && ast.IsExclusivelyTypeOnlyImportOrExport(declaration) {
					return
				}
				if action == "" && opts.ignorePackages {
					if import_utils.IsScopedModuleSpecifier(name) || settings.IsExternalModule(ctx, name, resolved) {
						return
					}
				}
				message := "Missing file extension "
				if extension != "" {
					message += "\"" + extension + "\" "
				}
				ctx.ReportNode(source, rule.RuleMessage{Description: message + "for \"" + written + "\""})
			} else if mode == "never" {
				withoutExtension := name[:len(name)-len(extension)-1]
				otherPath, otherFound := resolve(source, withoutExtension)
				// Both unresolved is also equality upstream (undefined === undefined).
				// Builtins are a distinct successful result with no filesystem path.
				if resolved == otherPath && found == otherFound {
					ctx.ReportNode(source, rule.RuleMessage{Description: "Unexpected use of file extension \"" + extension + "\" for \"" + written + "\""})
				}
			}
		}, import_utils.VisitModulesOptions{ESModule: true, Commonjs: true})
	},
}

func isExternalRoot(name string) bool {
	return name != "." && name != ".." && (strings.Count(name, "/") == 0 || import_utils.IsScopedModuleSpecifier(name) && strings.Count(name, "/") <= 1)
}

// Node's path.extname ignores leading dots and trailing separators. Go's Ext
// supplies the last-dot scan, with these two small semantic adjustments.
func fileExtension(name string) string {
	base := filepath.Base(name)
	extension := filepath.Ext(base)
	if base == ".." || extension == base {
		return ""
	}
	return strings.TrimPrefix(extension, ".")
}
