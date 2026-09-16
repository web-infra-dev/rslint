package exports_style

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	scopeAnalysis "github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed exports_style.schema.json
var schema []byte

var unexpectedExports = rule.RuleMessage{
	Id: "unexpectedExports", Description: "Unexpected access to 'exports'. Use 'module.exports' instead.",
}
var unexpectedModuleExports = rule.RuleMessage{
	Id: "unexpectedModuleExports", Description: "Unexpected access to 'module.exports'. Use 'exports' instead.",
}
var unexpectedAssignment = rule.RuleMessage{
	Id: "unexpectedAssignment", Description: "Unexpected assignment to 'exports'. Don't modify 'exports' itself.",
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/exports-style.js
var ExportsStyleRule = rule.Rule{
	Name:   "node/exports-style",
	Schema: rule.NewSchema(schema),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !ctx.SourceFile.HasIdentifier("exports") && !ctx.SourceFile.HasIdentifier("module") {
			return nil
		}
		mode := "module.exports"
		if len(options) > 0 {
			mode, _ = options[0].(string)
		}
		allowBatchAssign := false
		if len(options) > 1 {
			config, _ := options[1].(map[string]any)
			allowBatchAssign, _ = config["allowBatchAssign"].(bool)
		}

		scopes := scopeAnalysis.Get(ctx, scope.Options{
			CollectReferences: true,
			ReferenceNames:    map[string]struct{}{"exports": {}, "module": {}},
		})
		exports, moduleExports := exportReferences(ctx, scopes)
		assignments := map[*ast.Node]int{}
		if allowBatchAssign {
			nodes := moduleExports
			if mode == "exports" {
				nodes = exports
			}
			for _, node := range nodes {
				if top := topAssignment(node); top != nil {
					assignments[top]++
				}
			}
		}
		report := func(node *ast.Node, message rule.RuleMessage) {
			span := utils.TrimNodeTextRange(ctx.SourceFile, node)
			if next, ok := utils.TokenAtOrAfter(ctx.SourceFile, node.End()); ok {
				span = span.WithEnd(next.End)
			}
			ctx.ReportRange(span, message)
		}
		if mode == "module.exports" {
			for _, node := range exports {
				if assignments[topAssignment(node)] == 0 {
					report(node, unexpectedExports)
				}
			}
			return nil
		}

		batchAssignments := map[*ast.Node]bool{}
		var reports []*ast.Node
		for _, node := range moduleExports {
			top := topAssignment(node)
			if assignments[top] > 0 {
				// Upstream consumes one exports assignment for each match.
				assignments[top]--
				batchAssignments[top] = true
				continue
			}
			reports = append(reports, node)
		}
		for _, node := range exports {
			if isAssignee(node) && !batchAssignments[topAssignment(node)] {
				reports = append(reports, node)
			}
		}
		slices.SortStableFunc(reports, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
		for _, node := range reports {
			if node.Kind == ast.KindIdentifier {
				report(node, unexpectedAssignment)
			} else {
				report(node, unexpectedModuleExports)
			}
		}
		return nil
	},
}

// Upstream reads references from the program's outer global scope. In script
// files that includes authored top-level bindings; module-local bindings and
// declarations in nested scopes are excluded.
func exportReferences(ctx rule.RuleContext, scopes *scope.Manager) (exports, moduleExports []*ast.Node) {
	var identifiers []*ast.Node
	for _, ref := range scopes.References {
		if utils.IsInJsxTagName(ref.Identifier) {
			continue
		}
		if resolved := ref.Resolved(); resolved != nil {
			if resolved.Scope == scopes.Global && !ctx.Refs.HasNonGlobalProgramScope() {
				identifiers = append(identifiers, ref.Identifier)
			}
		} else {
			access := ctx.Globals.Access(ref.Identifier.Text())
			if access == utils.GlobalAccessReadonly || access == utils.GlobalAccessWritable {
				identifiers = append(identifiers, ref.Identifier)
			}
		}
	}
	// The scope reference index excludes declaration initializers. ESLint also
	// records their writes, which matter when exports is declared in a script.
	if !ctx.Refs.HasNonGlobalProgramScope() {
		for _, decl := range scopes.Global.Declarations("exports") {
			if decl.Kind != scope.DefVariable {
				continue
			}
			for range declarationWrites(decl.ID) {
				identifiers = append(identifiers, decl.ID)
			}
		}
	}
	slices.SortStableFunc(identifiers, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	for _, id := range identifiers {
		if id.Text() == "exports" {
			exports = append(exports, id)
			// A pattern default and the containing assignment each write the
			// target. ESLint retains both references, even at the same location.
			for node := id; node != nil; node = utils.ESTreeParent(node) {
				parent := utils.ESTreeParent(node)
				if parent == nil {
					break
				}
				switch parent.Kind {
				case ast.KindShorthandPropertyAssignment:
					property := parent.AsShorthandPropertyAssignment()
					if property.Name() == node && utils.IsInDestructuringAssignment(parent) {
						if property.ObjectAssignmentInitializer != nil {
							exports = append(exports, id)
						}
						continue
					}
				case ast.KindBinaryExpression:
					if utils.IsDefaultValueInDestructuringAssignment(parent) && utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Left) == node {
						exports = append(exports, id)
						continue
					}
				case ast.KindPropertyAssignment, ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindSpreadAssignment, ast.KindSpreadElement:
					continue
				}
				break
			}
			continue
		}
		parent := utils.ESTreeParent(id)
		_, property := utils.MemberExpressionParts(parent)
		if property == nil || utils.IsInJsxTagName(parent) {
			continue
		}
		property = utils.ESTreeRuntimeExpression(property)
		name := utils.GetStaticStringValue(property)
		if property.Kind == ast.KindIdentifier && parent.Kind != ast.KindElementAccessExpression {
			name = property.Text()
		}
		if name == "exports" {
			moduleExports = append(moduleExports, parent)
		}
	}
	return exports, moduleExports
}

func declarationWrites(id *ast.Node) int {
	writes := 0
	for node := id.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindBindingElement:
			if node.AsBindingElement().Initializer != nil {
				writes++
			}
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			continue
		case ast.KindVariableDeclaration:
			if node.AsVariableDeclaration().Initializer != nil {
				writes++
			}
			if node.Parent.Parent.Kind == ast.KindForInStatement || node.Parent.Parent.Kind == ast.KindForOfStatement {
				writes++
			}
			return writes
		default:
			return writes
		}
	}
	return writes
}

func isAssignment(node *ast.Node) bool {
	return ast.IsAssignmentExpression(node, false) && !utils.IsDefaultValueInDestructuringAssignment(node)
}

func isAssignee(node *ast.Node) bool {
	parent := utils.ESTreeParent(node)
	return isAssignment(parent) && utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Left) == node
}

func memberParent(node *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(node)
	object, _ := utils.MemberExpressionParts(parent)
	if object == nil || utils.ESTreeRuntimeExpression(object) != node {
		return nil
	}
	// Parentheses around an optional chain introduce an ESTree ChainExpression.
	if ast.IsOptionalChain(node) && (object != node || !ast.IsOptionalChain(parent)) {
		return nil
	}
	return parent
}

func topAssignment(node *ast.Node) *ast.Node {
	for parent := memberParent(node); parent != nil; parent = memberParent(node) {
		node = parent
	}
	if !isAssignee(node) {
		return nil
	}
	for parent := utils.ESTreeParent(node); isAssignment(parent); parent = utils.ESTreeParent(node) {
		node = parent
	}
	return node
}
