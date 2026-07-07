package no_cycle

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

const unlimitedDepth = int(^uint(0) >> 1)

type ruleOptions struct {
	maxDepth                           int
	ignoreExternal                     bool
	allowUnsafeDynamicCyclicDependency bool
	moduleReferences                   import_utils.ModuleReferenceOptions
}

type routeStep struct {
	value string
	line  int
}

type queuedModule struct {
	sourceFile *ast.SourceFile
	route      []routeStep
}

// NoCycleRule forbids dependency paths that resolve back to the linted module.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/no-cycle.js
var NoCycleRule = rule.Rule{
	Name: "import/no-cycle",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		checkSourceFile(ctx, opts)
		return rule.RuleListeners{}
	},
}

func parseOptions(options any) ruleOptions {
	opts := ruleOptions{
		maxDepth: unlimitedDepth,
		moduleReferences: import_utils.ModuleReferenceOptions{
			ESModule: true,
		},
	}

	optsMap := rslint_utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}

	if maxDepth, ok := parseMaxDepth(optsMap["maxDepth"]); ok {
		opts.maxDepth = maxDepth
	}
	if ignoreExternal, ok := optsMap["ignoreExternal"].(bool); ok {
		opts.ignoreExternal = ignoreExternal
	}
	if allow, ok := optsMap["allowUnsafeDynamicCyclicDependency"].(bool); ok {
		opts.allowUnsafeDynamicCyclicDependency = allow
	}
	if commonJS, ok := optsMap["commonjs"].(bool); ok {
		opts.moduleReferences.CommonJS = commonJS
	}
	if amd, ok := optsMap["amd"].(bool); ok {
		opts.moduleReferences.AMD = amd
	}
	if esmodule, ok := optsMap["esmodule"].(bool); ok {
		opts.moduleReferences.ESModule = esmodule
	}

	return opts
}

func parseMaxDepth(raw any) (int, bool) {
	switch value := raw.(type) {
	case nil:
		return 0, false
	case int:
		return normalizeDepth(value)
	case int64:
		return normalizeDepth(int(value))
	case float64:
		if math.IsInf(value, 1) {
			return unlimitedDepth, true
		}
		if math.Trunc(value) != value {
			return 0, false
		}
		return normalizeDepth(int(value))
	case float32:
		if math.IsInf(float64(value), 1) {
			return unlimitedDepth, true
		}
		if math.Trunc(float64(value)) != float64(value) {
			return 0, false
		}
		return normalizeDepth(int(value))
	case json.Number:
		if i, err := value.Int64(); err == nil {
			return normalizeDepth(int(i))
		}
		if f, err := value.Float64(); err == nil {
			return normalizeDepth(int(f))
		}
	case string:
		if value == "∞" {
			return unlimitedDepth, true
		}
	}
	return 0, false
}

func normalizeDepth(depth int) (int, bool) {
	if depth < 1 {
		return 0, false
	}
	return depth, true
}

func checkSourceFile(ctx rule.RuleContext, opts ruleOptions) {
	if ctx.SourceFile == nil || ctx.Program == nil {
		return
	}

	myPath := ctx.SourceFile.FileName()
	if myPath == "" || myPath == "<text>" {
		return
	}

	traversed := make(map[string]bool)
	for _, ref := range import_utils.CollectModuleReferences(ctx, ctx.SourceFile, opts.moduleReferences) {
		checkReference(ctx, opts, myPath, traversed, ref)
	}
}

func checkReference(ctx rule.RuleContext, opts ruleOptions, myPath string, traversed map[string]bool, ref import_utils.ModuleReference) {
	if ref.OnlyTypes || ref.Target == nil || shouldIgnoreExternal(ctx, opts, ref) {
		return
	}

	if opts.allowUnsafeDynamicCyclicDependency && ref.Dynamic {
		return
	}

	if moduleReferencePath(ref) == myPath {
		return
	}

	route, ok := detectCycle(ctx, opts, myPath, traversed, ref.Target)
	if !ok {
		return
	}

	reportNode := ref.Importer
	if reportNode == nil {
		reportNode = ref.Source
	}
	ctx.ReportNode(reportNode, messageCycle(route))
}

func detectCycle(ctx rule.RuleContext, opts ruleOptions, myPath string, traversed map[string]bool, start *ast.SourceFile) ([]routeStep, bool) {
	queue := []queuedModule{{sourceFile: start}}
	if traversed == nil {
		traversed = make(map[string]bool)
	}

	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		if next.sourceFile == nil {
			continue
		}

		sourcePath := next.sourceFile.FileName()
		if traversed[sourcePath] {
			continue
		}
		traversed[sourcePath] = true

		for _, ref := range referencesToTraverse(ctx, opts, import_utils.CollectModuleReferences(ctx, next.sourceFile, opts.moduleReferences)) {
			targetPath := moduleReferencePath(ref)
			if targetPath == "" || traversed[targetPath] {
				continue
			}
			if targetPath == myPath {
				return next.route, true
			}
			if ref.Target != nil && len(next.route)+1 < opts.maxDepth {
				queue = append(queue, queuedModule{
					sourceFile: ref.Target,
					route:      appendRoute(next.route, ref),
				})
			}
		}
	}

	return nil, false
}

func referencesToTraverse(ctx rule.RuleContext, opts ruleOptions, refs []import_utils.ModuleReference) []import_utils.ModuleReference {
	filtered := make([]import_utils.ModuleReference, 0, len(refs))
	dynamicPaths := make(map[string]bool)
	for _, ref := range refs {
		if ref.OnlyTypes || ref.Target == nil || shouldIgnoreExternal(ctx, opts, ref) {
			continue
		}
		if opts.allowUnsafeDynamicCyclicDependency && ref.Dynamic {
			dynamicPaths[moduleReferencePath(ref)] = true
		}
		filtered = append(filtered, ref)
	}

	if len(dynamicPaths) == 0 {
		return filtered
	}

	traversable := filtered[:0]
	for _, ref := range filtered {
		if !dynamicPaths[moduleReferencePath(ref)] {
			traversable = append(traversable, ref)
		}
	}
	return traversable
}

func appendRoute(route []routeStep, ref import_utils.ModuleReference) []routeStep {
	next := make([]routeStep, 0, len(route)+1)
	next = append(next, route...)
	next = append(next, routeStep{
		value: ref.Specifier,
		line:  sourceLine(ref),
	})
	return next
}

func sourceLine(ref import_utils.ModuleReference) int {
	if ref.SourceFile == nil || ref.Source == nil {
		return 1
	}
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(ref.SourceFile, ref.Source.Pos())
	return line + 1
}

func messageCycle(route []routeStep) rule.RuleMessage {
	description := "Dependency cycle detected."
	if len(route) > 0 {
		parts := make([]string, 0, len(route))
		for _, step := range route {
			parts = append(parts, fmt.Sprintf("%s:%d", step.value, step.line))
		}
		description = "Dependency cycle via " + strings.Join(parts, "=>")
	}
	return rule.RuleMessage{
		Id:          "cycle",
		Description: description,
	}
}

func moduleReferencePath(ref import_utils.ModuleReference) string {
	if ref.Target != nil {
		return ref.Target.FileName()
	}
	return ref.ResolvedPath
}

func shouldIgnoreExternal(ctx rule.RuleContext, opts ruleOptions, ref import_utils.ModuleReference) bool {
	if !opts.ignoreExternal {
		return false
	}
	return import_utils.IsExternalModulePath(ctx.Settings, ref.Specifier, moduleReferencePath(ref))
}
