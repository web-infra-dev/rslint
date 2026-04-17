package jsx_no_bind

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Violation kinds correspond to ESLint messageIds. The `bindExpression` kind
// (ES `::` proposal) is omitted because TypeScript does not parse it.
const (
	kindBindCall  = "bindCall"
	kindArrowFunc = "arrowFunc"
	kindFunc      = "func"
)

// violationOrder controls lookup priority when a JSX attribute value is an
// identifier that appears under multiple tracked kinds in the same block.
var violationOrder = []string{kindArrowFunc, kindBindCall, kindFunc}

// unwrap strips parentheses, which ESTree does not represent as nodes but the
// TypeScript AST does. Other TS-only wrappers (`as T`, `expr!`, `satisfies T`,
// `<T>expr`, partially-emitted) are intentionally left visible to match the
// original ESLint rule's behavior under typescript-eslint.
func unwrap(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipParentheses(node)
}

type options struct {
	allowArrowFunctions bool
	allowBind           bool
	allowFunctions      bool
	ignoreRefs          bool
	ignoreDOMComponents bool
}

func parseOptions(raw any) options {
	opts := options{}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowArrowFunctions"].(bool); ok {
		opts.allowArrowFunctions = v
	}
	if v, ok := optsMap["allowBind"].(bool); ok {
		opts.allowBind = v
	}
	if v, ok := optsMap["allowFunctions"].(bool); ok {
		opts.allowFunctions = v
	}
	if v, ok := optsMap["ignoreRefs"].(bool); ok {
		opts.ignoreRefs = v
	}
	if v, ok := optsMap["ignoreDOMComponents"].(bool); ok {
		opts.ignoreDOMComponents = v
	}
	return opts
}

func message(id string) rule.RuleMessage {
	var desc string
	switch id {
	case kindBindCall:
		desc = "JSX props should not use .bind()"
	case kindArrowFunc:
		desc = "JSX props should not use arrow functions"
	case kindFunc:
		desc = "JSX props should not use functions"
	}
	return rule.RuleMessage{Id: id, Description: desc}
}

var JsxNoBindRule = rule.Rule{
	Name: "react/jsx-no-bind",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// blockVariableNameSets tracks, per enclosing Block (keyed by node
		// position), the names of const-bound variables and local function
		// declarations whose value is a banned expression kind.
		blockVariableNameSets := map[int]map[string]map[string]bool{}

		initBlock := func(blockPos int) {
			if _, ok := blockVariableNameSets[blockPos]; ok {
				return
			}
			blockVariableNameSets[blockPos] = map[string]map[string]bool{
				kindArrowFunc: {},
				kindBindCall:  {},
				kindFunc:      {},
			}
		}

		var getNodeViolationType func(node *ast.Node) string
		getNodeViolationType = func(node *ast.Node) string {
			node = unwrap(node)
			if node == nil {
				return ""
			}
			if !opts.allowBind && node.Kind == ast.KindCallExpression {
				callee := unwrap(node.AsCallExpression().Expression)
				if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
					nameNode := callee.AsPropertyAccessExpression().Name()
					if nameNode != nil && nameNode.Kind == ast.KindIdentifier && nameNode.AsIdentifier().Text == "bind" {
						return kindBindCall
					}
				}
			}
			if node.Kind == ast.KindConditionalExpression {
				ce := node.AsConditionalExpression()
				if t := getNodeViolationType(ce.Condition); t != "" {
					return t
				}
				if t := getNodeViolationType(ce.WhenTrue); t != "" {
					return t
				}
				return getNodeViolationType(ce.WhenFalse)
			}
			if !opts.allowArrowFunctions && node.Kind == ast.KindArrowFunction {
				return kindArrowFunc
			}
			if !opts.allowFunctions && (node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindFunctionDeclaration) {
				return kindFunc
			}
			return ""
		}

		// blockAncestors returns the enclosing Block nodes from innermost to
		// outermost, matching ESLint's getBlockStatementAncestors().
		blockAncestors := func(node *ast.Node) []*ast.Node {
			var result []*ast.Node
			for cur := node.Parent; cur != nil; cur = cur.Parent {
				if cur.Kind == ast.KindBlock {
					result = append(result, cur)
				}
			}
			return result
		}

		reportVariableViolation := func(attrNode *ast.Node, name string, blockPos int) bool {
			sets, ok := blockVariableNameSets[blockPos]
			if !ok {
				return false
			}
			for _, violationType := range violationOrder {
				if sets[violationType][name] {
					ctx.ReportNode(attrNode, message(violationType))
					return true
				}
			}
			return false
		}

		findVariableViolation := func(attrNode *ast.Node, name string) {
			for _, block := range blockAncestors(attrNode) {
				if reportVariableViolation(attrNode, name, block.Pos()) {
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				initBlock(node.Pos())
			},

			ast.KindFunctionDeclaration: func(node *ast.Node) {
				ancestors := blockAncestors(node)
				if len(ancestors) == 0 {
					return
				}
				violation := getNodeViolationType(node)
				if violation == "" {
					return
				}
				nameNode := node.AsFunctionDeclaration().Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				blockPos := ancestors[0].Pos()
				initBlock(blockPos)
				blockVariableNameSets[blockPos][violation][nameNode.AsIdentifier().Text] = true
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				varDecl := node.AsVariableDeclaration()
				if varDecl.Initializer == nil {
					return
				}
				// Only `const` bindings, matching ESLint.
				declList := node.Parent
				if declList == nil || declList.Kind != ast.KindVariableDeclarationList {
					return
				}
				if declList.Flags&ast.NodeFlagsConst == 0 {
					return
				}
				// `for (const x of ...)` and friends: the declaration lives inside a
				// `For*Statement`, not a `VariableStatement`. Skip, because ESLint only
				// tracks true block-scoped const bindings.
				if declList.Parent == nil || declList.Parent.Kind != ast.KindVariableStatement {
					return
				}
				ancestors := blockAncestors(node)
				if len(ancestors) == 0 {
					return
				}
				violation := getNodeViolationType(varDecl.Initializer)
				if violation == "" {
					return
				}
				nameNode := varDecl.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				blockPos := ancestors[0].Pos()
				initBlock(blockPos)
				blockVariableNameSets[blockPos][violation][nameNode.AsIdentifier().Text] = true
			},

			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				nameNode := attr.Name()
				if opts.ignoreRefs && nameNode != nil && nameNode.Kind == ast.KindIdentifier && nameNode.AsIdentifier().Text == "ref" {
					return
				}
				initializer := attr.Initializer
				if initializer == nil || initializer.Kind != ast.KindJsxExpression {
					return
				}
				expr := unwrap(initializer.AsJsxExpression().Expression)
				if expr == nil {
					return
				}
				if opts.ignoreDOMComponents && reactutil.IsDOMComponent(reactutil.GetJsxParentElement(node)) {
					return
				}
				if expr.Kind == ast.KindIdentifier {
					findVariableViolation(node, expr.AsIdentifier().Text)
					return
				}
				if violation := getNodeViolationType(expr); violation != "" {
					ctx.ReportNode(node, message(violation))
				}
			},
		}
	},
}
