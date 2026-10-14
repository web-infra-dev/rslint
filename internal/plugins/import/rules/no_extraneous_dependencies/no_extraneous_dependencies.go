package no_extraneous_dependencies

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed no_extraneous_dependencies.schema.json
var schemaJSON []byte

var dependencyFields = [...]string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies", "bundledDependencies"}

type dependencySet [5]map[string]bool

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-extraneous-dependencies.js
// Upstream uses literal messages, without message IDs or edits. Read manifests
// from this invocation's Program so edits and separate VFSs stay visible.
var NoExtraneousDependenciesRule = rule.Rule{
	Name:   "import/no-extraneous-dependencies",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !ctx.Program().IsValid() {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		cwd := ctx.ProcessCurrentDirectory()
		if cwd == "" {
			cwd = ctx.Program().CurrentDirectory()
		}
		deps := readDependencies(ctx, opts["packageDir"], cwd)
		var allowed [5]bool
		allowed[0] = true
		for i := 1; i < 5; i++ {
			allowed[i] = allowDependency(opts[dependencyFields[i]], ctx.SourceFile.FileName(), cwd)
		}
		settings := import_utils.SettingsFor(ctx)
		resolver := import_utils.NewImportResolver(ctx)
		reportedResolverError := false
		return import_utils.VisitModules(func(source modules.Source) {
			typeOnly := source.TypeOnly() || source.Declaration().Kind == ast.KindExportDeclaration && source.Declaration().IsTypeOnly()
			if opts["includeTypes"] != true && typeOnly {
				return
			}
			name := source.Specifier().Text()
			resolved, found, resolveError := resolver.Resolve(source)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			if !found || resolved == "" {
				return
			}
			external := settings.IsExternalModule(ctx, name, resolved)
			// importType calls a bare, resolved alias internal after excluding builtin,
			// absolute and relative spellings. includeInternal does not include siblings.
			internal := settings.IsInternalSpecifier(name) || !tspath.IsExternalModuleNameRelative(name) && !tspath.IsRootedDiskPath(name) && !external
			if !external && (opts["includeInternal"] != true || !internal) {
				return
			}
			original := packageName(name)
			status := declared(deps, original)
			if permitted(status, allowed) {
				return
			}
			realName := realPackageName(ctx, resolved)
			if realName != "" && realName != original {
				extra := declared(deps, realName)
				for i := range status {
					status[i] = status[i] || extra[i]
				}
				if permitted(status, allowed) {
					return
				}
			}
			if realName == "" {
				realName = original
			}
			message := fmt.Sprintf("'%s' should be listed in the project's dependencies. Run 'npm i -S %s' to add it", realName, realName)
			if status[1] && !allowed[1] {
				message = fmt.Sprintf("'%s' should be listed in the project's dependencies, not devDependencies.", realName)
			} else if status[2] && !allowed[2] {
				message = fmt.Sprintf("'%s' should be listed in the project's dependencies, not optionalDependencies.", realName)
			}
			ctx.ReportNode(source.Declaration(), rule.RuleMessage{Description: message})
		}, import_utils.VisitModulesOptions{Commonjs: true, ESModule: true})
	},
}

func allowDependency(value any, filename, cwd string) bool {
	if value == nil {
		return true
	}
	if b, ok := value.(bool); ok {
		return b
	}
	for _, pattern := range stringList(value) {
		if minimatch3.Match(pattern, filename, minimatch3.Options{}) || minimatch3.Match(tspath.NormalizePath(filepath.Join(cwd, pattern)), filename, minimatch3.Options{}) {
			return true
		}
	}
	return false
}
func stringList(value any) []string {
	if s, ok := value.(string); ok {
		return []string{s}
	}
	var result []string
	if values, ok := value.([]any); ok {
		for _, v := range values {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
	}
	return result
}
func packageName(name string) string {
	parts := strings.Split(name, "/")
	if strings.HasPrefix(name, "@") && len(parts) > 1 {
		return strings.Join(parts[:2], "/")
	}
	return parts[0]
}
func declared(deps dependencySet, name string) (status [5]bool) {
	parts := strings.Split(name, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, "@") {
			continue
		}
		ancestor := strings.Join(parts[:i+1], "/")
		for field := range status {
			status[field] = status[field] || deps[field][ancestor]
		}
	}
	return
}
func permitted(status, allowed [5]bool) bool {
	for i := range status {
		if status[i] && allowed[i] {
			return true
		}
	}
	return false
}
func readDependencies(ctx rule.RuleContext, configured any, cwd string) dependencySet {
	var result dependencySet
	for i := range result {
		result[i] = map[string]bool{}
	}
	dirs := stringList(configured)
	if configured == "" {
		dirs = nil
	}
	bundled := map[int]string{}
	explicit := len(dirs) > 0
	if !explicit {
		dirs = []string{ctx.Program().NearestPackageJSONDirectory(tspath.GetDirectoryPath(ctx.SourceFile.FileName()))}
	}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		file := tspath.ResolvePath(tspath.ResolvePath(cwd, dir), "package.json")
		text, ok := ctx.Program().FS().ReadFile(file)
		if !ok {
			if explicit && len(dirs) == 1 {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "The package.json file could not be found."})
			}
			continue
		}
		var pkg map[string]any
		if err := json.Unmarshal([]byte(text), &pkg); err != nil {
			// readJSON suppresses malformed manifests for automatic discovery and
			// multiple directories. An explicit single directory reports the error.
			if explicit && len(dirs) == 1 {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "The package.json file could not be parsed: " + err.Error()})
			}
			continue
		}
		var raw map[string]json.RawMessage
		_ = json.Unmarshal([]byte(text), &raw)
		for i, field := range dependencyFields {
			if i == 4 {
				value := raw["bundleDependencies"]
				if !truthy(pkg["bundleDependencies"]) {
					value = raw["bundledDependencies"]
				}
				for index, name := range bundleNames(value) {
					bundled[index] = name
				}
			} else if obj, ok := pkg[field].(map[string]any); ok {
				for name := range obj {
					result[i][name] = true
				}
			}
		}
	}
	for _, name := range bundled {
		result[4][name] = true
	}
	return result
}
func realPackageName(ctx rule.RuleContext, resolved string) string {
	dir := tspath.GetDirectoryPath(resolved)
	for dir != "" {
		root := ctx.Program().NearestPackageJSONDirectory(dir)
		if root == "" {
			return ""
		}
		if text, ok := ctx.Program().FS().ReadFile(tspath.ResolvePath(root, "package.json")); ok {
			var pkg struct {
				Name string `json:"name"`
			}
			if json.Unmarshal([]byte(text), &pkg) == nil && pkg.Name != "" {
				return pkg.Name
			}
		}
		parent := tspath.GetDirectoryPath(root)
		if parent == root {
			return ""
		}
		dir = parent
	}
	return ""
}

// JSON object key order matters because upstream converts object bundles to
// arrays and merges multiple manifests by array index. Decode keys in order.
func bundleNames(raw json.RawMessage) []string {
	var values []string
	if json.Unmarshal(raw, &values) == nil {
		return values
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil
	}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil
		}
		var ignored json.RawMessage
		if decoder.Decode(&ignored) != nil {
			return nil
		}
		name, ok := key.(string)
		if !ok {
			return nil
		}
		values = append(values, name)
	}
	return values
}
func truthy(value any) bool {
	switch value := value.(type) {
	case nil:
		return false
	case bool:
		return value
	case string:
		return value != ""
	case float64:
		return value != 0
	default:
		return true
	}
}
