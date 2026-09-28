package jsx_no_constructed_context_values

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

// Upstream disables option validation and ignores every supplied option.
//
//go:embed jsx_no_constructed_context_values.schema.json
var schemaJSON []byte

type construction struct {
	kind  string
	node  *ast.Node
	usage *ast.Node
}

// scope stores destructured bindings at their BindingElement; ESLint stores
// the containing VariableDeclarator as the definition node instead.
func variableDeclarator(def *scope.Variable) *ast.Node {
	if def.Kind != scope.DefVariable {
		return nil
	}
	return ast.GetRootDeclaration(def.DefNode)
}

// All recursive lookups deliberately use the invocation scope, without
// searching its parents. Values declared outside that scope are stable for
// this rule, even when another scope belongs to the same component.
func findConstruction(node *ast.Node, invocation *scope.Scope, active map[*ast.Node]bool) construction {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || ast.IsOptionalChain(node) {
		return construction{}
	}
	result := construction{node: node}
	switch node.Kind {
	case ast.KindRegularExpressionLiteral:
		result.kind = "regular expression"
	case ast.KindObjectLiteralExpression:
		result.kind = "object"
	case ast.KindArrayLiteralExpression:
		result.kind = "array"
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		result.kind = "function expression"
	case ast.KindClassExpression:
		result.kind = "class expression"
	case ast.KindNewExpression:
		result.kind = "new expression"
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		result.kind = "JSX element"
	case ast.KindJsxFragment:
		result.kind = "JSX fragment"
	case ast.KindIdentifier:
		defs := invocation.Declarations(node.Text())
		if len(defs) == 0 {
			return construction{}
		}
		def := defs[len(defs)-1]
		if def.Kind == scope.DefFunctionName && def.DefNode.Kind == ast.KindFunctionDeclaration && def.DefNode.Body() != nil {
			return construction{kind: "function declaration", node: def.DefNode, usage: node}
		}
		declaration := variableDeclarator(def)
		if declaration == nil || active[declaration] {
			return construction{}
		}
		// Cyclic aliases can occur in incomplete or unreachable source.
		active[declaration] = true
		result = findConstruction(declaration.AsVariableDeclaration().Initializer, invocation, active)
		delete(active, declaration)
		result.usage = node
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		result = findConstruction(conditional.WhenTrue, invocation, active)
		if result.kind == "" {
			result = findConstruction(conditional.WhenFalse, invocation, active)
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind) {
			result = findConstruction(binary.Left, invocation, active)
			if result.kind == "" {
				result = findConstruction(binary.Right, invocation, active)
			}
		} else if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			result = findConstruction(binary.Right, invocation, active)
			if result.kind != "" {
				result.kind = "assignment expression"
				result.usage = node
			}
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		object := utils.ESTreeRuntimeExpression(node.Expression())
		result = findConstruction(object, invocation, active)
		result.usage = object
	case ast.KindAsExpression:
		return findConstruction(node.AsAsExpression().Expression, invocation, active)
	}
	return result
}

func isContextTag(tag *ast.Node, scopes scopeanalysis.Provider) bool {
	if tag.Kind == ast.KindPropertyAccessExpression {
		return tag.AsPropertyAccessExpression().Name().Text() == "Provider"
	}
	if tag.Kind != ast.KindIdentifier {
		return false
	}
	// React 19 context tags search outward, unlike constructed value lookups.
	for current := scopes.Declarations().Acquire(tag); current != nil; current = current.Parent {
		defs := current.Declarations(tag.Text())
		if len(defs) == 0 {
			continue
		}
		declaration := variableDeclarator(defs[0])
		if declaration == nil {
			return false
		}
		init := utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer)
		if init == nil || init.Kind != ast.KindCallExpression || ast.IsOptionalChain(init) {
			return false
		}
		callee := utils.ESTreeCallCallee(init.AsCallExpression().Expression)
		if callee == nil {
			return false
		}
		if callee.Kind == ast.KindIdentifier {
			return callee.Text() == "createContext"
		}
		object, property := utils.MemberExpressionParts(callee)
		object = utils.ESTreeRuntimeExpression(object)
		property = utils.ESTreeRuntimeExpression(property)
		return object != nil && object.Kind == ast.KindIdentifier && object.Text() == "React" && property != nil &&
			((property.Kind == ast.KindIdentifier && property.Text() == "createContext") ||
				(property.Kind == ast.KindPrivateIdentifier && property.Text() == "#createContext"))
	}
	return false
}

func constructionMessage(ctx rule.RuleContext, value construction) rule.RuleMessage {
	line := func(node *ast.Node) int {
		return scanner.GetECMALineOfPosition(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()) + 1
	}
	id, hook := "defaultMsg", "useMemo"
	var description string
	if value.usage == nil {
		description = fmt.Sprintf("The %s passed as the value prop to the Context provider (at line %d) changes every render.", value.kind, line(value.node))
	} else {
		id = "withIdentifierMsg"
		name := "undefined"
		if value.usage.Kind == ast.KindIdentifier {
			name = value.usage.Text()
		}
		description = fmt.Sprintf("The '%s' %s (at line %d) passed as the value prop to the Context provider (at line %d) changes every render.", name, value.kind, line(value.node), line(value.usage))
	}
	if value.kind == "function expression" || value.kind == "function declaration" {
		id += "Func"
		hook = "useCallback"
	}
	return rule.RuleMessage{Id: id, Description: description + " To fix this consider wrapping it in a " + hook + " hook."}
}

var JsxNoConstructedContextValuesRule = rule.Rule{
	Name:   "react/jsx-no-constructed-context-values",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		scopes := scopeanalysis.For(ctx)
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		check := func(node *ast.Node) {
			for _, attribute := range reactutil.GetJsxElementAttributes(node) {
				if attribute.Kind != ast.KindJsxAttribute || reactutil.GetJsxPropName(attribute) != "value" {
					continue
				}
				value := attribute.AsJsxAttribute().Initializer
				if value == nil || value.Kind != ast.KindJsxExpression {
					return
				}
				// Ordinary JSX without a value expression needs no context lookup.
				if !isContextTag(reactutil.GetJsxTagName(node), scopes) {
					return
				}
				constructed := findConstruction(value.AsJsxExpression().Expression, scopes.Declarations().Acquire(node), make(map[*ast.Node]bool))
				if constructed.kind != "" && reactutil.GetParentReactComponentScopeBasedOrStateless(node, pragma, createClass, wrappers, scopes) != nil {
					ctx.ReportNode(constructed.node, constructionMessage(ctx, constructed))
				}
				// Upstream inspects only the first value attribute.
				return
			}
		}
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
