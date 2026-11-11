package constructor_super

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"sort"
)

func buildMissingAll() rule.RuleMessage {
	return rule.RuleMessage{Id: "missingAll", Description: "Expected to call 'super()'."}
}
func buildMissingSome() rule.RuleMessage {
	return rule.RuleMessage{Id: "missingSome", Description: "Lacked a call of 'super()' in some code paths."}
}
func buildDuplicate() rule.RuleMessage {
	return rule.RuleMessage{Id: "duplicate", Description: "Unexpected duplicate 'super()'."}
}
func buildBadSuper() rule.RuleMessage {
	return rule.RuleMessage{Id: "badSuper", Description: "Unexpected 'super()' because 'super' is not a constructor."}
}

func superClass(node *ast.Node) *ast.Node {
	clauses := utils.GetHeritageClauses(node)
	if clauses != nil {
		for _, c := range clauses.Nodes {
			h := c.AsHeritageClause()
			if h.Token == ast.KindExtendsKeyword && h.Types != nil && len(h.Types.Nodes) > 0 {
				return h.Types.Nodes[0].Expression()
			}
		}
	}
	return nil
}

// Match ESLint's syntactic possibility check, without resolving values or types.
func isPossibleConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	switch node.Kind {
	case ast.KindClassExpression, ast.KindFunctionExpression, ast.KindThisKeyword,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression,
		ast.KindNewExpression, ast.KindYieldExpression, ast.KindTaggedTemplateExpression, ast.KindMetaProperty:
		return true
	case ast.KindIdentifier:
		return node.Text() != "undefined"
	case ast.KindBinaryExpression:
		b := node.AsBinaryExpression()
		switch b.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			return isPossibleConstructor(b.Right)
		case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return isPossibleConstructor(b.Left) || isPossibleConstructor(b.Right)
		}
	case ast.KindConditionalExpression:
		c := node.AsConditionalExpression()
		return isPossibleConstructor(c.WhenTrue) || isPossibleConstructor(c.WhenFalse)
	}
	return false
}

// ConstructorSuperRule tracks completed super() calls separately from returned
// values. A finally completion can replace a return without changing whether
// this was initialized.
var ConstructorSuperRule = rule.Rule{
	Name: "constructor-super", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindConstructor: func(n *ast.Node) {
			if n.Body() == nil || n.Parent == nil {
				return
			}
			base := superClass(n.Parent)
			if base == nil && n.Body().SubtreeFacts()&ast.SubtreeContainsLexicalSuper == 0 {
				return
			}
			a := analyzer{possible: isPossibleConstructor(base), instanceElements: hasInstanceElements(n.Parent)}
			state := uncalled
			for _, param := range n.Parameters() {
				state = a.bindingDefault(param.Name(), param.Initializer(), state, true)
			}
			result := a.statement(n.Body(), state)
			returned := result.normal | result.returned
			if result.returnedValue != 0 {
				returned |= called
			}
			if base != nil && returned&uncalled != 0 {
				if returned&called != 0 {
					ctx.ReportNode(n, buildMissingSome())
				} else {
					ctx.ReportNode(n, buildMissingAll())
				}
			}
			calls := make([]*ast.Node, 0, len(a.reports))
			for call := range a.reports {
				calls = append(calls, call)
			}
			sort.Slice(calls, func(i, j int) bool { return calls[i].Pos() < calls[j].Pos() })
			for _, call := range calls {
				ctx.ReportNode(call, a.reports[call])
			}
		}}
	},
}

// Instance element initialization occurs after super() binds this, and can
// throw with this already initialized. Static members execute outside it.
func hasInstanceElements(class *ast.Node) bool {
	for _, member := range class.Members() {
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if member.Kind == ast.KindPropertyDeclaration {
			return true
		}
		if name := member.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			return true
		}
	}
	return false
}
