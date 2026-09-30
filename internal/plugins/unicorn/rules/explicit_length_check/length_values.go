package explicit_length_check

import (
	"math"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The exclusion policy belongs to this rule. Symbol references and JavaScript
// evaluation come from the existing services; this is not a general flow engine.
type lengthValues struct {
	ctx       rule.RuleContext
	evaluator *utils.StaticStringEvaluator
}

func (v *lengthValues) eval() *utils.StaticStringEvaluator {
	if v.evaluator == nil {
		v.evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(v.ctx.TypeChecker, v.ctx.SourceFile, v.ctx.Refs)
		v.evaluator.GlobalAccess = v.ctx.Globals.Access
	}
	return v.evaluator
}

func isCardinality(value any) bool {
	number, ok := value.(jsnum.Number)
	return ok && number >= 0 && number <= 9007199254740991 && math.Trunc(float64(number)) == float64(number)
}

func (v *lengthValues) knownNumber(node *ast.Node) bool {
	value, known := v.eval().EvalControlFlowValue(node)
	return known && isCardinality(value)
}

func (v *lengthValues) knownNonCollection(member *ast.Node) bool {
	if value, known := v.eval().EvalSideEffectFreeValue(member); known {
		return !isCardinality(value)
	}
	object := utils.ESTreeRuntimeExpression(member.Expression())
	if !ast.IsIdentifier(object) || v.ctx.Refs == nil {
		return false
	}
	symbol := v.ctx.Refs.ResolveInFile(object)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindVariableDeclaration {
		return false
	}
	initializer := utils.ESTreeRuntimeExpression(symbol.Declarations[0].AsVariableDeclaration().Initializer)
	if initializer == nil || initializer.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	property := member.Name().Text()
	execution := executionContext(member)
	unknown, numeric, accessor := false, false, false
	for _, reference := range v.ctx.Refs.References(symbol) {
		effect := v.propertyEffect(reference, property)
		if effect == effectRead {
			continue
		}
		uncertain := executionContext(reference) != execution || v.conditional(reference)
		if !uncertain && reference.Pos() > member.Pos() {
			continue
		}
		switch effect {
		case effectNumeric:
			if uncertain {
				unknown = true
			} else {
				numeric = true
			}
		case effectUnknown:
			unknown = true
		case effectAccessor:
			if uncertain {
				unknown = true
			} else {
				accessor = true
			}
		case effectEscape:
			if uncertain {
				unknown = true
			} else {
				return false
			}
		}
	}
	if unknown {
		return true
	}
	if numeric {
		return false
	}
	if accessor {
		return true
	}
	value, known := v.eval().EvalSideEffectFreePropertyValue(initializer, property)
	return known && !isCardinality(value)
}

type propertyEffect uint8

const (
	effectRead propertyEffect = iota
	effectNumeric
	effectUnknown
	effectAccessor
	effectEscape
)

func (v *lengthValues) numericEffect(value *ast.Node) propertyEffect {
	if v.knownNumber(value) {
		return effectNumeric
	}
	return effectUnknown
}

func (v *lengthValues) propertyEffect(reference *ast.Node, property string) propertyEffect {
	if call := v.objectCall(reference); call != nil {
		args := call.Arguments()
		method, _ := v.eval().EvalAccessExpressionName(utils.ESTreeRuntimeExpression(call.Expression()))
		var descriptor *ast.Node
		switch method {
		case "defineProperty":
			if len(args) < 2 {
				return effectUnknown
			}
			key, known := v.eval().EvalControlFlowValue(args[1])
			if !known {
				return effectUnknown
			}
			if key != property {
				return effectRead
			}
			if len(args) > 2 {
				descriptor = utils.ESTreeRuntimeExpression(args[2])
			}
			if descriptor == nil {
				return effectUnknown
			}
		case "defineProperties":
			if len(args) < 2 {
				return effectUnknown
			}
			descriptor = v.lastPropertyValue(utils.ESTreeRuntimeExpression(args[1]), property)
			if descriptor == nil {
				if v.unknownProperties(args[1]) {
					return effectUnknown
				}
				return effectRead
			}
		case "assign":
			if len(args) < 2 {
				return effectRead
			}
			if len(args) != 2 || hasSpread(args[1]) {
				return effectUnknown
			}
			value := v.lastPropertyValue(utils.ESTreeRuntimeExpression(args[1]), property)
			if value != nil {
				return v.numericEffect(value)
			}
			if v.unknownProperties(args[1]) {
				return effectUnknown
			}
			return effectRead
		default:
			return effectEscape
		}
		if descriptor.Kind == ast.KindObjectLiteralExpression && !hasSpread(descriptor) {
			for _, property := range descriptor.AsObjectLiteralExpression().Properties.Nodes {
				name, _ := v.eval().EvalPropertyName(property.Name())
				if name == "get" || name == "set" {
					return effectAccessor
				}
			}
		}
		return v.numericEffect(v.lastPropertyValue(descriptor, "value"))
	}
	member := reference
	for parent := utils.ESTreeParent(member); parent != nil && ast.IsAccessExpression(parent) && utils.ESTreeRuntimeExpression(parent.Expression()) == member; parent = utils.ESTreeParent(member) {
		member = parent
	}
	if !ast.IsAccessExpression(member) {
		return effectEscape
	}
	name, known := v.eval().EvalAccessExpressionName(member)
	if !known || name != property {
		return effectEscape
	}
	parent := utils.ESTreeParent(member)
	if parent == nil {
		return effectRead
	}
	if parent.Kind == ast.KindForInStatement && utils.ESTreeRuntimeExpression(parent.AsForInOrOfStatement().Initializer) == member {
		return effectUnknown
	}
	if parent.Kind == ast.KindForOfStatement && utils.ESTreeRuntimeExpression(parent.AsForInOrOfStatement().Initializer) == member {
		values, known := v.eval().EvalControlFlowArrayElements(parent.AsForInOrOfStatement().Expression)
		if !known || len(values) == 0 {
			return effectUnknown
		}
		for _, value := range values {
			if !isCardinality(value) {
				return effectUnknown
			}
		}
		return effectNumeric
	}
	if ast.IsAssignmentTarget(utils.OutermostParenthesizedExpression(member)) {
		return v.numericEffect(v.assignmentValue(member))
	}
	if parent.Kind == ast.KindTaggedTemplateExpression && utils.ESTreeRuntimeExpression(parent.AsTaggedTemplateExpression().Tag) == member {
		return effectEscape
	}
	if (parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression) &&
		utils.ESTreeRuntimeExpression(parent.Expression()) == member {
		return effectEscape
	}
	return effectRead
}

func (v *lengthValues) objectCall(reference *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(reference)
	if parent == nil || parent.Kind != ast.KindCallExpression || ast.IsOptionalChain(parent) ||
		len(parent.Arguments()) == 0 || utils.ESTreeRuntimeExpression(parent.Arguments()[0]) != reference {
		return nil
	}
	callee := utils.ESTreeCallCallee(parent.Expression())
	if callee == nil || !ast.IsAccessExpression(callee) || ast.IsOptionalChain(callee) {
		return nil
	}
	object := utils.ESTreeRuntimeExpression(callee.Expression())
	if !unicornutil.IsGlobalReference(v.ctx, object) || !v.ctx.Globals.Access(object.Text()).IsDeclared() {
		return nil
	}
	return parent
}

func hasSpread(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	return slices.ContainsFunc(node.AsObjectLiteralExpression().Properties.Nodes, ast.IsSpreadAssignment)
}

func (v *lengthValues) unknownProperties(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return true
	}
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindSpreadAssignment {
			return true
		}
		if _, known := v.eval().EvalPropertyName(property.Name()); !known && ast.IsComputedPropertyName(property.Name()) {
			return true
		}
	}
	return false
}

func (v *lengthValues) lastPropertyValue(object *ast.Node, name string) *ast.Node {
	object = utils.ESTreeRuntimeExpression(object)
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	for _, property := range slices.Backward(object.AsObjectLiteralExpression().Properties.Nodes) {
		if property.Kind == ast.KindSpreadAssignment {
			return nil
		}
		key, known := v.eval().EvalPropertyName(property.Name())
		if !known && ast.IsComputedPropertyName(property.Name()) {
			return nil
		}
		if key == name {
			if _, value, ok := unicornutil.ObjectDataProperty(property); ok {
				return utils.ESTreeRuntimeExpression(value)
			}
			return nil
		}
	}
	return nil
}

func (v *lengthValues) assignmentValue(node *ast.Node) *ast.Node {
	var path []*ast.Node
	for current := node; current != nil; current = utils.ESTreeParent(current) {
		parent := utils.ESTreeParent(current)
		if parent == nil {
			return nil
		}
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken.Kind != ast.KindEqualsToken || utils.ESTreeRuntimeExpression(binary.Left) != current || ast.IsAssignmentTarget(parent) {
				return nil
			}
			value := utils.ESTreeRuntimeExpression(binary.Right)
			for _, part := range slices.Backward(path) {
				if value == nil {
					return nil
				}
				if part.Parent.Kind == ast.KindArrayLiteralExpression {
					if value.Kind != ast.KindArrayLiteralExpression {
						return nil
					}
					index := slices.Index(part.Parent.AsArrayLiteralExpression().Elements.Nodes, part)
					values := value.AsArrayLiteralExpression().Elements.Nodes
					if index < 0 || index >= len(values) {
						return nil
					}
					value = utils.ESTreeRuntimeExpression(values[index])
				} else {
					name, known := v.eval().EvalPropertyName(part.Name())
					if !known {
						return nil
					}
					value = v.lastPropertyValue(value, name)
				}
			}
			return value
		case ast.KindArrayLiteralExpression:
			path = append(path, utils.OutermostParenthesizedExpression(current))
		case ast.KindPropertyAssignment:
			path = append(path, parent)
		case ast.KindSpreadElement, ast.KindSpreadAssignment:
			return nil
		}
	}
	return nil
}

func executionContext(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, func(current *ast.Node) bool {
		return ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) || current.Kind == ast.KindPropertyDeclaration
	})
}

func (v *lengthValues) conditional(node *ast.Node) bool {
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		var test *ast.Node
		consequent := false
		switch parent.Kind {
		case ast.KindIfStatement:
			statement := parent.AsIfStatement()
			if current == statement.ThenStatement || current == statement.ElseStatement {
				test, consequent = statement.Expression, current == statement.ThenStatement
			}
		case ast.KindConditionalExpression:
			expression := parent.AsConditionalExpression()
			if current == expression.WhenTrue || current == expression.WhenFalse {
				test, consequent = expression.Condition, current == expression.WhenTrue
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.Right == current && (unicornutil.IsLogicalExpression(parent) || binary.OperatorToken.Kind == ast.KindQuestionQuestionToken ||
				(binary.OperatorToken.Kind == ast.KindEqualsToken && ast.IsAssignmentTarget(parent))) {
				return true
			}
		case ast.KindForStatement:
			statement := parent.AsForStatement()
			if statement.Statement == current || statement.Incrementor == current {
				return true
			}
		case ast.KindWhileStatement, ast.KindForInStatement, ast.KindForOfStatement:
			if parent.Statement() == current {
				return true
			}
		case ast.KindCaseClause:
			if slices.Contains(parent.AsCaseOrDefaultClause().Statements.Nodes, current) {
				return true
			}
		case ast.KindDefaultClause:
			if len(parent.Parent.AsCaseBlock().Clauses.Nodes) != 1 {
				return true
			}
		case ast.KindCatchClause:
			if parent.AsCatchClause().Block == current {
				return true
			}
		case ast.KindTryStatement:
			if parent.AsTryStatement().TryBlock == current {
				return true
			}
		case ast.KindBindingElement:
			if parent.AsBindingElement().Initializer == current {
				return true
			}
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().ArgumentExpression == current && unicornutil.HasOptionalChainElement(parent) {
				return true
			}
		case ast.KindCallExpression:
			if slices.Contains(parent.Arguments(), current) && unicornutil.HasOptionalChainElement(parent) {
				return true
			}
		}
		if test != nil {
			truthy, known := v.eval().EvalControlFlowTruthiness(test)
			if !known || truthy != consequent {
				return true
			}
		}
	}
	return false
}
