package prefer_exact_props

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const propTypesMessage = "Component propTypes should be exact by using %s."

var PreferExactPropsRule = rule.Rule{
	Name:   "react/prefer-exact-props",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		exactWrappers := getExactPropWrappers(ctx.Settings)
		if len(exactWrappers) == 0 {
			return rule.RuleListeners{}
		}

		message := rule.RuleMessage{
			Id:          "propTypes",
			Description: fmt.Sprintf(propTypesMessage, formatExactPropWrappers(exactWrappers)),
		}
		aliasInitializers := make(map[*ast.Symbol]*ast.Node)

		reportIfNonExact := func(reportNode, value *ast.Node, resolveIdentifier bool) {
			if isNonExactPropTypesValue(ctx, value, exactWrappers, resolveIdentifier, aliasInitializers) {
				ctx.ReportNode(reportNode, message)
			}
		}

		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				if property == nil || !isPropTypesPropertyDeclaration(node) {
					return
				}
				// The upstream class-property listener only inspects direct object
				// and call initializers; identifier resolution belongs solely to
				// its MemberExpression assignment path.
				reportIfNonExact(node, property.Initializer, false)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil {
					return
				}
				if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return
				}
				left := utils.ESTreeRuntimeExpression(binary.Left)
				if !isPropTypesMember(left) {
					return
				}
				// Optional members sit inside an ESTree ChainExpression rather than
				// directly under the binary expression inspected upstream.
				if ast.IsOptionalChain(left) {
					return
				}
				reportIfNonExact(left, binary.Right, true)
			},
		}
	},
}

func getExactPropWrappers(settings map[string]interface{}) []reactutil.PropWrapperEntry {
	wrappers := reactutil.GetPropWrapperFunctions(settings)
	exact := make([]reactutil.PropWrapperEntry, 0, len(wrappers))
	for _, wrapper := range wrappers {
		if wrapper.Exact {
			exact = append(exact, wrapper)
		}
	}
	return exact
}

func formatExactPropWrappers(wrappers []reactutil.PropWrapperEntry) string {
	formatted := make([]string, 0, len(wrappers))
	for _, wrapper := range wrappers {
		name := wrapper.Property
		if wrapper.Object != "" {
			name = wrapper.Object + "." + wrapper.Property
		}
		formatted = append(formatted, "'"+name+"'")
	}
	joined := strings.Join(formatted, ", ")
	if len(formatted) > 1 {
		return "one of " + joined
	}
	return joined
}

func isNonExactPropTypesValue(
	ctx rule.RuleContext,
	value *ast.Node,
	exactWrappers []reactutil.PropWrapperEntry,
	resolveIdentifier bool,
	aliasInitializers map[*ast.Symbol]*ast.Node,
) bool {
	value = utils.ESTreeRuntimeExpression(value)
	if value == nil {
		return false
	}
	switch value.Kind {
	case ast.KindObjectLiteralExpression:
		return len(value.AsObjectLiteralExpression().Properties.Nodes) > 0
	case ast.KindCallExpression:
		// ESLint exposes any optional call/property chain as a ChainExpression,
		// so it never reaches the upstream CallExpression predicate.
		if ast.IsOptionalChain(value) {
			return false
		}
		return !isExactPropWrapperCall(ctx.SourceFile, value, exactWrappers)
	case ast.KindIdentifier:
		if !resolveIdentifier {
			return false
		}
		initializer := resolveDirectConstInitializer(ctx, value, aliasInitializers)
		return isNonExactPropTypesInitializer(ctx.SourceFile, initializer, exactWrappers)
	default:
		return false
	}
}

func isNonExactPropTypesInitializer(
	sourceFile *ast.SourceFile,
	initializer *ast.Node,
	exactWrappers []reactutil.PropWrapperEntry,
) bool {
	initializer = utils.ESTreeRuntimeExpression(initializer)
	if initializer == nil {
		return false
	}
	switch initializer.Kind {
	case ast.KindObjectLiteralExpression:
		return len(initializer.AsObjectLiteralExpression().Properties.Nodes) > 0
	case ast.KindCallExpression:
		if ast.IsOptionalChain(initializer) {
			return false
		}
		return !isExactPropWrapperCall(sourceFile, initializer, exactWrappers)
	default:
		return false
	}
}

// resolveDirectConstInitializer resolves only an identifier's own, directly
// initialized const declaration. Object-literal aliases remain known only
// while every runtime reference is another propTypes assignment RHS; mutable,
// destructured, mutated, and escaped bindings are kept unknown.
func resolveDirectConstInitializer(
	ctx rule.RuleContext,
	identifier *ast.Node,
	aliasInitializers map[*ast.Symbol]*ast.Node,
) *ast.Node {
	if ctx.Refs == nil || identifier == nil || identifier.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := ctx.Refs.Resolve(identifier)
	if symbol == nil {
		return nil
	}
	if initializer, ok := aliasInitializers[symbol]; ok {
		return initializer
	}
	initializer := classifyDirectConstInitializer(ctx, identifier, symbol)
	aliasInitializers[symbol] = initializer
	return initializer
}

func classifyDirectConstInitializer(ctx rule.RuleContext, identifier *ast.Node, symbol *ast.Symbol) *ast.Node {
	declarationNode := symbol.ValueDeclaration
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration ||
		ast.GetSourceFileOfNode(declarationNode) != ctx.SourceFile {
		return nil
	}
	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList ||
		!ast.IsVarConst(declarationList) {
		return nil
	}
	declaration := declarationNode.AsVariableDeclaration()
	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != identifier.AsIdentifier().Text {
		return nil
	}
	initializer := declaration.Initializer
	runtimeInitializer := utils.ESTreeRuntimeExpression(initializer)
	if runtimeInitializer != nil && runtimeInitializer.Kind == ast.KindObjectLiteralExpression &&
		!hasOnlyPropTypesAssignmentReferences(ctx, symbol) {
		return nil
	}
	return initializer
}

// hasOnlyPropTypesAssignmentReferences keeps object-literal aliases
// conservative without introducing a second reference index. RefStore builds
// and caches the file's identifier buckets once; this rule only classifies the
// cached references for each symbol once through aliasInitializers.
func hasOnlyPropTypesAssignmentReferences(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	for _, reference := range ctx.Refs.References(symbol) {
		if !utils.IsReadReference(reference) {
			continue
		}
		if !isPropTypesAssignmentValueReference(reference) {
			return false
		}
	}
	return true
}

func isPropTypesAssignmentValueReference(reference *ast.Node) bool {
	parent := utils.ESTreeParent(reference)
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) ||
		utils.ESTreeRuntimeExpression(binary.Right) != reference {
		return false
	}
	left := utils.ESTreeRuntimeExpression(binary.Left)
	return left != nil && !ast.IsOptionalChain(left) && isPropTypesMember(left)
}

func isExactPropWrapperCall(sourceFile *ast.SourceFile, node *ast.Node, wrappers []reactutil.PropWrapperEntry) bool {
	if sourceFile == nil || node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	name := utils.TrimmedNodeText(sourceFile, callee)
	for _, wrapper := range wrappers {
		if name == wrapper.Property {
			return true
		}
		if strings.Count(name, ".") == 1 && wrapper.Object != "" && name == wrapper.Object+"."+wrapper.Property {
			return true
		}
	}
	return false
}

func isPropTypesPropertyDeclaration(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyDeclaration ||
		!ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
		return false
	}
	name, ok := utils.GetStaticPropertyName(node.AsPropertyDeclaration().Name())
	return ok && name == "propTypes"
}

func isPropTypesMember(node *ast.Node) bool {
	if node == nil {
		return false
	}
	name, ok := utils.AccessExpressionStaticName(node)
	return ok && name == "propTypes"
}
