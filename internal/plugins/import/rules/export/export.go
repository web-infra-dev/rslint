package export

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// See https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/export.js.
var ExportRule = rule.Rule{
	Name:   "import/export",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		c := collector{ctx: ctx, byName: make(map[exportKey]int)}
		for _, statement := range ctx.SourceFile.Statements.Nodes {
			c.collect(statement)
		}
		c.report()
		return nil
	},
}

type exportKey struct {
	scope *ast.Node
	name  string
}

type exportGroup struct {
	name  string
	first *ast.Node
	// Allocate the occurrence list only once a name is repeated.
	nodes []*ast.Node
}

// Duplicate detection belongs to this file's declarations. Cross-file export
// discovery stays in the import plugin's shared, read-only ExportMap.
type collector struct {
	ctx         rule.RuleContext
	byName      map[exportKey]int
	groups      []exportGroup
	diagnostics []exportDiagnostic
}

type exportDiagnostic struct {
	node    *ast.Node
	message rule.RuleMessage
}

func (c *collector) add(statement, node *ast.Node, name string, typeOnly bool) {
	if typeOnly {
		name = "type:" + name
	}
	var scope *ast.Node
	if statement.Parent != nil && statement.Parent.Kind == ast.KindModuleBlock {
		scope = statement.Parent.Parent
	}
	key := exportKey{scope: scope, name: name}
	index, ok := c.byName[key]
	if !ok {
		c.byName[key] = len(c.groups)
		c.groups = append(c.groups, exportGroup{name: name, first: node})
		return
	}
	// ExportMap.Names already deduplicates names reached through several paths.
	group := &c.groups[index]
	if group.nodes == nil {
		group.nodes = []*ast.Node{group.first, node}
	} else {
		group.nodes = append(group.nodes, node)
	}
}

func (c *collector) collect(node *ast.Node) {
	switch node.Kind {
	case ast.KindExportAssignment:
		if !node.AsExportAssignment().IsExportEquals {
			c.add(node, node, "default", false)
		}
	case ast.KindExportDeclaration:
		c.collectExport(node)
	default:
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
			c.collectDeclaration(node)
		}
	}

	// Each namespace or ambient module declaration has its own export scope,
	// even when TypeScript merges declarations with the same name.
	if node.Kind == ast.KindModuleDeclaration {
		body := node.AsModuleDeclaration().Body
		if body == nil {
			return
		}
		switch body.Kind {
		case ast.KindModuleBlock:
			for _, statement := range body.AsModuleBlock().Statements.Nodes {
				c.collect(statement)
			}
		case ast.KindModuleDeclaration:
			c.collect(body)
		}
	}
}

func (c *collector) collectDeclaration(node *ast.Node) {
	// Upstream removes overload signatures before considering duplicates.
	if node.Kind == ast.KindFunctionDeclaration && node.Body() == nil {
		return
	}
	// Qualified namespace names and string module names have no identifier
	// name upstream. They only merge with each other and cannot conflict.
	if node.Kind == ast.KindModuleDeclaration && (node.Name().Kind != ast.KindIdentifier ||
		node.Body() != nil && node.Body().Kind == ast.KindModuleDeclaration) {
		return
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) {
		c.add(node, node, "default", false)
		return
	}
	switch node.Kind {
	case ast.KindVariableStatement:
		utils.ForEachVariableDeclarationBinding(node.AsVariableStatement().DeclarationList, func(_ *ast.Node, identifier *ast.Node, name string) {
			c.add(node, identifier, name, false)
		})
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration, ast.KindImportEqualsDeclaration:
		if name := node.Name(); name != nil {
			c.add(node, name, name.Text(), node.Kind == ast.KindTypeAliasDeclaration || node.Kind == ast.KindInterfaceDeclaration)
		}
	}
}

func (c *collector) collectExport(node *ast.Node) {
	declaration := node.AsExportDeclaration()
	if clause := declaration.ExportClause; clause != nil {
		if clause.Kind == ast.KindNamedExports {
			for _, specifier := range clause.AsNamedExports().Elements.Nodes {
				name := specifier.Name()
				if text, ok := utils.GetStaticPropertyName(name); ok {
					// Type-only specifiers share the value export bucket upstream.
					c.add(node, name, text, false)
				}
			}
			return
		}
		// Upstream excludes identifier namespace aliases, but treats string
		// aliases like ordinary star exports (exported.name is absent).
		if clause.Kind == ast.KindNamespaceExport && clause.Name().Kind == ast.KindIdentifier {
			return
		}
	}
	exports, ok := import_utils.GetExportMap(c.ctx, declaration.ModuleSpecifier)
	if !ok {
		return
	}
	anyNamed := false
	for name := range exports.Names() {
		if name != "default" {
			anyNamed = true
			c.add(node, node, name, false)
		}
	}
	if !anyNamed {
		c.diagnostics = append(c.diagnostics, exportDiagnostic{declaration.ModuleSpecifier, rule.RuleMessage{
			Id:          "noNamed",
			Description: fmt.Sprintf("No named exports found in module '%s'.", declaration.ModuleSpecifier.Text()),
		}})
	}
}

func (c *collector) report() {
	for _, group := range c.groups {
		nodes := group.nodes
		if len(nodes) == 0 {
			continue
		}
		kinds := make(map[ast.Kind]int)
		for _, node := range nodes {
			kinds[node.Parent.Kind]++
		}
		hasNamespace := kinds[ast.KindModuleDeclaration] > 0
		if hasNamespace && (len(kinds) == 1 || len(kinds) == 2 &&
			(kinds[ast.KindFunctionDeclaration] > 0 || kinds[ast.KindClassDeclaration] == 1 || kinds[ast.KindEnumDeclaration] == 1)) {
			continue
		}
		message := rule.RuleMessage{
			Id:          "multipleNamed",
			Description: fmt.Sprintf("Multiple exports of name '%s'.", strings.Replace(group.name, "type:", "", 1)),
		}
		if group.name == "default" {
			message = rule.RuleMessage{Id: "multipleDefault", Description: "Multiple default exports."}
		}
		for _, node := range nodes {
			if node.Parent.Kind == ast.KindModuleDeclaration &&
				(kinds[ast.KindFunctionDeclaration] > 0 || kinds[ast.KindClassDeclaration] > 0 || kinds[ast.KindEnumDeclaration] > 0) {
				continue
			}
			c.diagnostics = append(c.diagnostics, exportDiagnostic{node: node, message: message})
		}
	}
	slices.SortStableFunc(c.diagnostics, func(a, b exportDiagnostic) int { return cmp.Compare(a.node.Pos(), b.node.Pos()) })
	for _, diagnostic := range c.diagnostics {
		reportRange := utils.GetESTreeBindingIdentifierRange(c.ctx.SourceFile, diagnostic.node)
		// ESTree starts a default export at `export`, excluding decorators
		// written before that keyword. tsgo includes them in the declaration.
		if modifiers := diagnostic.node.Modifiers(); modifiers != nil {
			for _, modifier := range modifiers.Nodes {
				if modifier.Kind == ast.KindExportKeyword {
					reportRange = core.NewTextRange(utils.TrimNodeTextRange(c.ctx.SourceFile, modifier).Pos(), reportRange.End())
					break
				}
			}
		}
		c.ctx.ReportRange(reportRange, diagnostic.message)
	}
}
