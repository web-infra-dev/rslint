package no_internal_modules

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed no_internal_modules.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-internal-modules.js
var NoInternalModulesRule = rule.Rule{
	Name:   "import/no-internal-modules",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		_, forbid := opts["forbid"]
		key := "allow"
		if forbid {
			key = "forbid"
		}
		var patterns []*minimatch3.RegexpMatcher
		for _, pattern := range utils.ToStringSlice(opts[key]) {
			patterns = append(patterns, minimatch3.New(pattern, minimatch3.Options{}).MakeRe())
		}
		matches := func(path string) bool {
			for _, pattern := range patterns {
				if pattern.Test(path) {
					return true
				}
			}
			return false
		}
		settings := import_utils.SettingsFor(ctx)
		resolver := import_utils.NewImportResolver(ctx)
		reportedResolverError := false
		return import_utils.VisitModules(func(ref modules.Source) {
			source := ref.Specifier()
			name := source.Text()
			// importType resolves even when a normalized raw path matches.
			resolved, _, resolveError := resolver.Resolve(ref)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			if !potentialViolation(settings, name, resolved) {
				return
			}
			steps := pathSteps(name)
			if !forbid {
				nonScope := 0
				for _, step := range steps {
					if !strings.HasPrefix(step, "@") {
						nonScope++
					}
				}
				if nonScope <= 1 {
					return
				}
			}
			joined := strings.Join(steps, "/")
			matched := matches(joined) || matches("/"+joined)
			if !matched && resolved != "" {
				matched = matches(strings.ReplaceAll(resolved, `\`, "/"))
			}
			violation := matched
			if !forbid {
				violation = !matched && resolved != ""
			}
			if violation {
				ctx.ReportNode(source, rule.RuleMessage{Description: fmt.Sprintf(`Reaching to "%s" is not allowed.`, name)})
			}
		}, import_utils.VisitModulesOptions{Commonjs: true, ESModule: true})
	},
}

// The rule checks the union of parent/index/sibling/external/internal, so
// package-boundary distinctions need no separate filesystem classification.
func potentialViolation(settings *import_utils.ModuleSettings, name, resolved string) bool {
	if settings.IsInternalSpecifier(name) {
		return true
	}
	if nodeutil.IsAbsolutePath(name) || settings.IsBuiltinSpecifier(name, resolved) {
		return false
	}
	if resolved != "" || name == "." || name == ".." || strings.HasPrefix(name, "./") ||
		strings.HasPrefix(name, "../") || strings.HasPrefix(name, `.\`) || strings.HasPrefix(name, `..\`) {
		return true
	}
	if name == "" {
		return false
	}
	c := name[0]
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
		import_utils.IsScopedModuleSpecifier(name)
}

func pathSteps(name string) []string {
	var steps []string
	for _, step := range strings.Split(strings.ReplaceAll(name, `\`, "/"), "/") {
		switch step {
		case "", ".":
		case "..":
			if len(steps) > 0 {
				steps = steps[:len(steps)-1]
			}
		default:
			steps = append(steps, step)
		}
	}
	return steps
}
