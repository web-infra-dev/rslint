// Ported from eslint-plugin-unicorn v77.0.0 (MIT).
package import_style

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	styleUnassigned = "unassigned"
	styleDefault    = "default"
	styleNamespace  = "namespace"
	styleNamed      = "named"

	messageID       = "importStyle"
	messageIDBanned = "importStyleBanned"
)

var allStyles = [...]string{styleUnassigned, styleDefault, styleNamespace, styleNamed}

//go:embed import_style.schema.json
var schemaJSON []byte

type moduleStyles struct {
	allowed map[string]bool
	order   []string
}

func (styles *moduleStyles) set(name string, allowed bool) {
	if _, exists := styles.allowed[name]; !exists {
		styles.order = append(styles.order, name)
	}
	styles.allowed[name] = allowed
}

type options struct {
	styles             map[string]moduleStyles
	bannedModules      map[string]bool
	checkImport        bool
	checkDynamicImport bool
	checkExportFrom    bool
	checkRequire       bool
}

func parseOptions(raw []any) options {
	result := options{
		styles:             map[string]moduleStyles{},
		bannedModules:      map[string]bool{},
		checkImport:        true,
		checkDynamicImport: true,
		checkRequire:       true,
	}

	values := map[string]any{}
	if len(raw) > 0 {
		values, _ = raw[0].(map[string]any)
	}
	extendDefaults := true
	if value, ok := values["extendDefaultStyles"].(bool); ok {
		extendDefaults = value
	}
	if value, ok := values["checkImport"].(bool); ok {
		result.checkImport = value
	}
	if value, ok := values["checkDynamicImport"].(bool); ok {
		result.checkDynamicImport = value
	}
	if value, ok := values["checkExportFrom"].(bool); ok {
		result.checkExportFrom = value
	}
	if value, ok := values["checkRequire"].(bool); ok {
		result.checkRequire = value
	}

	configured, _ := values["styles"].(map[string]any)
	if extendDefaults {
		result.styles["chalk"] = newModuleStyles(styleDefault)
		result.styles["path"] = newModuleStyles(styleDefault)
		result.styles["util"] = newModuleStyles(styleNamed)
	}
	for moduleName, rawModuleStyles := range configured {
		if rawModuleStyles == false {
			result.styles[moduleName] = moduleStyles{allowed: map[string]bool{}}
			continue
		}

		module, exists := result.styles[moduleName]
		if !exists {
			module = moduleStyles{allowed: map[string]bool{}}
		}
		styleValues, _ := rawModuleStyles.(map[string]any)
		if allStylesFalse(styleValues) {
			result.bannedModules[moduleName] = true
		}
		// Go's decoded JSON objects do not retain property insertion order. This
		// ordering matches upstream's own defaults and its documented/tested
		// multi-style examples while keeping diagnostics deterministic.
		for _, style := range [...]string{styleNamed, styleNamespace, styleDefault, styleUnassigned} {
			if allowed, ok := styleValues[style].(bool); ok {
				module.set(style, allowed)
			}
		}
		customStyles := make([]string, 0, len(styleValues))
		for style := range styleValues {
			if style != styleUnassigned && style != styleDefault && style != styleNamespace && style != styleNamed {
				customStyles = append(customStyles, style)
			}
		}
		sort.Strings(customStyles)
		for _, style := range customStyles {
			if allowed, ok := styleValues[style].(bool); ok {
				module.set(style, allowed)
			}
		}
		result.styles[moduleName] = module
	}

	return result
}

func newModuleStyles(styles ...string) moduleStyles {
	result := moduleStyles{allowed: map[string]bool{}}
	for _, style := range styles {
		result.set(style, true)
	}
	return result
}

func allStylesFalse(values map[string]any) bool {
	for _, style := range allStyles {
		value, ok := values[style].(bool)
		if !ok || value {
			return false
		}
	}
	return true
}

func actualImportDeclarationStyles(node *ast.Node) []string {
	declaration := node.AsImportDeclaration()
	if declaration.ImportClause == nil {
		return []string{styleUnassigned}
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause.IsTypeOnly() {
		return nil
	}

	styles := newStyleSet()
	if clause.Name() != nil {
		styles.add(styleDefault)
	}
	bindings := clause.NamedBindings
	if bindings == nil {
		return styles.values()
	}
	if bindings.Kind == ast.KindNamespaceImport {
		styles.add(styleNamespace)
		return styles.values()
	}
	elements := bindings.AsNamedImports().Elements.Nodes
	for _, specifierNode := range elements {
		specifier := specifierNode.AsImportSpecifier()
		if specifier.IsTypeOnly {
			continue
		}
		imported := specifier.PropertyName
		if imported == nil {
			imported = specifier.Name()
		}
		if imported.Kind == ast.KindIdentifier && imported.Text() == "default" {
			styles.add(styleDefault)
		} else {
			styles.add(styleNamed)
		}
	}
	if len(elements) == 0 {
		return []string{styleUnassigned}
	}
	return styles.values()
}

func actualExportDeclarationStyles(node *ast.Node) []string {
	declaration := node.AsExportDeclaration()
	if declaration.ExportClause == nil {
		return []string{styleNamespace}
	}
	if declaration.ExportClause.Kind == ast.KindNamespaceExport {
		return []string{styleNamespace}
	}
	named := declaration.ExportClause.AsNamedExports()
	if named.Elements == nil || len(named.Elements.Nodes) == 0 {
		return []string{styleUnassigned}
	}
	styles := newStyleSet()
	for _, specifierNode := range named.Elements.Nodes {
		exported := specifierNode.AsExportSpecifier().Name()
		if exported.Kind == ast.KindIdentifier && exported.Text() == "default" {
			styles.add(styleDefault)
		} else {
			styles.add(styleNamed)
		}
	}
	return styles.values()
}

func actualAssignmentTargetStyles(target *ast.Node) []string {
	if target == nil {
		return nil
	}
	switch target.Kind {
	case ast.KindIdentifier, ast.KindArrayBindingPattern:
		return []string{styleNamespace}
	case ast.KindObjectBindingPattern:
		pattern := target.AsBindingPattern()
		if pattern.Elements == nil || len(pattern.Elements.Nodes) == 0 {
			return []string{styleUnassigned}
		}
		styles := newStyleSet()
		for _, elementNode := range pattern.Elements.Nodes {
			element := elementNode.AsBindingElement()
			if element.DotDotDotToken != nil {
				styles.add(styleNamed)
				continue
			}
			key := element.PropertyName
			if key == nil {
				key = element.Name()
			}
			if key != nil && key.Kind == ast.KindIdentifier {
				if key.Text() == "default" {
					styles.add(styleDefault)
				} else {
					styles.add(styleNamed)
				}
			}
		}
		return styles.values()
	default:
		return nil
	}
}

type styleSet struct {
	seen  map[string]bool
	order []string
}

func newStyleSet() *styleSet { return &styleSet{seen: map[string]bool{}} }

func (set *styleSet) add(style string) {
	if !set.seen[style] {
		set.seen[style] = true
		set.order = append(set.order, style)
	}
}

func (set *styleSet) values() []string { return set.order }

func directIdentifierCall(node *ast.Node, name string, minimumArguments int, exactArguments *int, rejectOptional bool) bool {
	if node == nil || !ast.IsCallExpression(node) || ast.IsImportCall(node) {
		return false
	}
	if rejectOptional && ast.IsOptionalChainRoot(node) {
		return false
	}
	callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != name {
		return false
	}
	arguments := node.Arguments()
	if exactArguments != nil && len(arguments) != *exactArguments {
		return false
	}
	return len(arguments) >= minimumArguments
}

func assignedDynamicImport(node *ast.Node) (*ast.Node, bool) {
	parent := utils.ESTreeParent(node)
	if parent == nil || parent.Kind != ast.KindAwaitExpression || utils.ESTreeRuntimeExpression(parent.Expression()) != node {
		return nil, false
	}
	declaration := utils.ESTreeParent(parent)
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration ||
		utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer) != parent {
		return nil, false
	}
	return declaration, true
}

func moduleStyleName(moduleName string) string {
	return strings.TrimPrefix(moduleName, "node:")
}

func disjunction(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
	}
}

func report(ctx rule.RuleContext, opts options, node *ast.Node, moduleName string, actual []string, isRequire bool) {
	configured, exists := opts.styles[moduleStyleName(moduleName)]
	if !exists {
		return
	}

	allowed := make([]string, 0, len(configured.order))
	allowedSet := make(map[string]bool, len(configured.allowed)+1)
	for _, style := range configured.order {
		if configured.allowed[style] {
			allowed = append(allowed, style)
			allowedSet[style] = true
		}
	}
	if len(allowed) == 0 {
		if opts.bannedModules[moduleStyleName(moduleName)] {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          messageIDBanned,
				Description: "All import styles are disabled for module `" + moduleName + "`. Use the `no-restricted-imports` rule to disallow a module.",
				Data:        map[string]string{"moduleName": moduleName},
			})
		}
		return
	}
	if isRequire && allowedSet[styleDefault] {
		allowedSet[styleNamespace] = true
	}
	for _, style := range actual {
		if !allowedSet[style] {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          messageID,
				Description: "Use " + disjunction(allowed) + " import for module `" + moduleName + "`.",
				Data: map[string]string{
					"allowedStyles": disjunction(allowed),
					"moduleName":    moduleName,
				},
			})
			return
		}
	}
}

var ImportStyleRule = rule.Rule{
	Name:   "unicorn/import-style",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		staticStrings := utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
		staticStrings.GlobalAccess = ctx.Globals.Access
		evaluate := func(node *ast.Node) (string, bool) {
			return staticStrings.EvalToString(node)
		}

		listeners := rule.RuleListeners{}
		if opts.checkImport {
			listeners[ast.KindImportDeclaration] = func(node *ast.Node) {
				moduleName, ok := evaluate(node.AsImportDeclaration().ModuleSpecifier)
				if ok {
					report(ctx, opts, node, moduleName, actualImportDeclarationStyles(node), false)
				}
			}
		}
		if opts.checkExportFrom {
			listeners[ast.KindExportDeclaration] = func(node *ast.Node) {
				declaration := node.AsExportDeclaration()
				if declaration.ModuleSpecifier == nil {
					return
				}
				moduleName, ok := evaluate(declaration.ModuleSpecifier)
				if ok {
					report(ctx, opts, node, moduleName, actualExportDeclarationStyles(node), false)
				}
			}
		}

		if opts.checkDynamicImport || opts.checkRequire {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				if opts.checkDynamicImport && ast.IsImportCall(node) {
					if _, assigned := assignedDynamicImport(node); assigned {
						return
					}
					moduleName, ok := evaluate(ast.GetExternalModuleName(node))
					if ok {
						report(ctx, opts, node, moduleName, []string{styleUnassigned}, false)
					}
					return
				}
				if !opts.checkRequire {
					return
				}
				exact := 1
				if !directIdentifierCall(node, "require", 0, &exact, true) {
					return
				}
				parent := utils.ESTreeParent(node)
				if parent == nil || parent.Kind != ast.KindExpressionStatement ||
					utils.ESTreeRuntimeExpression(parent.Expression()) != node {
					return
				}
				moduleName, ok := evaluate(node.Arguments()[0])
				if ok {
					report(ctx, opts, node, moduleName, []string{styleUnassigned}, true)
				}
			}

			listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				initializer := utils.ESTreeRuntimeExpression(declaration.Initializer)
				if opts.checkDynamicImport && initializer != nil && initializer.Kind == ast.KindAwaitExpression {
					importCall := utils.ESTreeRuntimeExpression(initializer.Expression())
					if importCall != nil && ast.IsImportCall(importCall) {
						moduleName, ok := evaluate(ast.GetExternalModuleName(importCall))
						if ok && moduleName != "" {
							report(ctx, opts, node, moduleName, actualAssignmentTargetStyles(declaration.Name()), false)
						}
						return
					}
				}
				if !opts.checkRequire || !directIdentifierCall(initializer, "require", 1, nil, true) {
					return
				}
				moduleName, ok := evaluate(initializer.Arguments()[0])
				if ok && moduleName != "" {
					report(ctx, opts, node, moduleName, actualAssignmentTargetStyles(declaration.Name()), true)
				}
			}
		}
		return listeners
	},
}
