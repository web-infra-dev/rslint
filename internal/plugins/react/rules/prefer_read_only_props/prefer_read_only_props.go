package prefer_read_only_props

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const readOnlyPropMessage = "Prop '%s' should be read-only."

var PreferReadOnlyPropsRule = rule.Rule{
	Name:   "react/prefer-read-only-props",
	Schema: rule.EmptyArraySchema,
	Run:    runRule,
}

// runRule mirrors eslint-plugin-react's propTypes collector for the
// TypeScript shapes understood by tsgo.
func runRule(ctx rule.RuleContext, _ []any) rule.RuleListeners {
	pragma := reactutil.GetReactPragma(ctx.Settings)
	typeAliases, genericImports := collectTopLevelTypes(ctx.SourceFile)
	wrapperFunctions := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
	type reportKey struct {
		owner *ast.Node
		name  string
	}
	type pendingReport struct {
		node     *ast.Node
		name     string
		readonly bool
	}
	pending := make(map[reportKey]int)
	reports := make([]pendingReport, 0)

	report := func(owner, node *ast.Node, name string, readonly bool) {
		if owner == nil || node == nil {
			return
		}
		key := reportKey{owner: owner, name: name}
		if index, ok := pending[key]; ok {
			reports[index] = pendingReport{node: node, name: name, readonly: readonly}
			return
		}
		pending[key] = len(reports)
		reports = append(reports, pendingReport{node: node, name: name, readonly: readonly})
	}

	flushReports := func() {
		sort.SliceStable(reports, func(i, j int) bool {
			return reports[i].node.Pos() < reports[j].node.Pos()
		})
		for _, item := range reports {
			if item.readonly {
				continue
			}
			node, name := item.node, item.name
			ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
				Id:          "readOnlyProp",
				Description: fmt.Sprintf(readOnlyPropMessage, name),
				Data:        map[string]string{"name": name},
			}, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, node, "readonly ")}
			})
		}
	}

	validateType := func(owner, node *ast.Node) {
		validateTypeNode(node, typeAliases, genericImports, func(node *ast.Node, name string, readonly bool) {
			report(owner, node, name, readonly)
		}, map[*ast.Node]bool{})
	}

	return rule.RuleListeners{
		ast.KindClassDeclaration: func(node *ast.Node) {
			if !reactutil.ExtendsReactComponent(node, pragma) {
				return
			}
			if heritage := ast.GetClassExtendsHeritageElement(node); heritage != nil {
				if extends := heritage.AsExpressionWithTypeArguments(); extends != nil &&
					extends.TypeArguments != nil && len(extends.TypeArguments.Nodes) > 0 {
					if extends.TypeArguments.Nodes[0].Kind == ast.KindTypeReference {
						validateType(node, extends.TypeArguments.Nodes[0])
					}
				}
			}
		},
		ast.KindClassExpression: func(node *ast.Node) {
			if !reactutil.ExtendsReactComponent(node, pragma) {
				return
			}
			if heritage := ast.GetClassExtendsHeritageElement(node); heritage != nil {
				if extends := heritage.AsExpressionWithTypeArguments(); extends != nil &&
					extends.TypeArguments != nil && len(extends.TypeArguments.Nodes) > 0 {
					if extends.TypeArguments.Nodes[0].Kind == ast.KindTypeReference {
						validateType(node, extends.TypeArguments.Nodes[0])
					}
				}
			}
		},
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			pd := node.AsPropertyDeclaration()
			if pd == nil || pd.Type == nil || reactutil.EnclosingClass(node) == nil {
				return
			}
			classNode := reactutil.EnclosingClass(node)
			if !reactutil.ExtendsReactComponent(classNode, pragma) {
				return
			}
			name := pd.Name()
			if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "props" {
				return
			}
			validateType(classNode, pd.Type)
		},
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			checkFunction(node, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindFunctionExpression: func(node *ast.Node) {
			checkFunction(node, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindArrowFunction: func(node *ast.Node) {
			checkFunction(node, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindMethodDeclaration: func(node *ast.Node) {
			checkFunction(node, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindGetAccessor: func(node *ast.Node) {
			checkFunction(node, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindSetAccessor: func(node *ast.Node) {
			checkFunction(node, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindEndOfFile: func(_ *ast.Node) {
			flushReports()
		},
	}
}

func checkFunction(node *ast.Node, pragma string, genericImports map[string]string, wrappers []reactutil.ComponentWrapperEntry, validateType func(*ast.Node, *ast.Node)) {
	// The forwardRef arm is checked before component-return heuristics in the
	// upstream collector. Its second type argument is the props type.
	if call := parentCall(node); call != nil && call.TypeArguments != nil && len(call.TypeArguments.Nodes) >= 2 && isForwardRefCall(call) {
		if !isImportedForwardRef(call, genericImports) {
			return
		}
		validateType(node, call.TypeArguments.Nodes[1])
		return
	}

	if !reactutil.IsStatelessReactComponentWithWrappers(node, pragma, nil, wrappers) {
		return
	}
	params := reactutil.FunctionParameters(node)
	if len(params) == 0 {
		return
	}
	if param := params[0].AsParameterDeclaration(); param != nil && param.Type != nil {
		validateType(node, param.Type)
	}

	// For `const Component: React.FC<Props> = (...) => ...`, the upstream
	// collector uses the variable's annotation rather than the parameter's.
	// Validate it as well; the per-node cycle guard prevents recursive aliases.
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindVariableDeclaration {
		vd := parent.AsVariableDeclaration()
		if vd != nil && vd.Type != nil {
			if propsType := reactGenericPropsType(vd.Type, genericImports); propsType != nil {
				validateType(node, propsType)
			}
		}
	}
}

func parentCall(node *ast.Node) *ast.CallExpression {
	parent := reactutil.SkipExpressionWrappersUp(node)
	if parent != nil && parent.Kind == ast.KindCallExpression {
		return parent.AsCallExpression()
	}
	return nil
}

func isForwardRefCall(call *ast.CallExpression) bool {
	if call == nil {
		return false
	}
	callee := reactutil.SkipExpressionWrappers(call.Expression)
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		return callee.AsIdentifier().Text == "forwardRef"
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := callee.AsPropertyAccessExpression()
	obj := reactutil.SkipExpressionWrappers(pa.Expression)
	name := pa.Name()
	return obj != nil && obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "React" &&
		name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "forwardRef"
}

func isImportedForwardRef(call *ast.CallExpression, genericImports map[string]string) bool {
	if call == nil {
		return false
	}
	callee := reactutil.SkipExpressionWrappers(call.Expression)
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		pa := callee.AsPropertyAccessExpression()
		obj := reactutil.SkipExpressionWrappers(pa.Expression)
		name := pa.Name()
		return obj != nil && obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "React" &&
			name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "forwardRef"
	}
	if callee.Kind == ast.KindIdentifier {
		return genericImports[callee.AsIdentifier().Text] == "forwardRef"
	}
	return false
}

func collectTopLevelTypes(source *ast.SourceFile) (map[string][]*ast.Node, map[string]string) {
	types := map[string][]*ast.Node{}
	imports := map[string]string{}
	if source == nil {
		return types, imports
	}
	source.Node.ForEachChild(func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindTypeAliasDeclaration:
			decl := node.AsTypeAliasDeclaration()
			if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
				types[decl.Name().AsIdentifier().Text] = append(types[decl.Name().AsIdentifier().Text], decl.Type)
			}
		case ast.KindInterfaceDeclaration:
			decl := node.AsInterfaceDeclaration()
			if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
				types[decl.Name().AsIdentifier().Text] = append(types[decl.Name().AsIdentifier().Text], node)
			}
		case ast.KindImportDeclaration:
			collectReactGenericImports(node, imports)
		}
		return false
	})
	return types, imports
}

func collectReactGenericImports(node *ast.Node, imports map[string]string) {
	decl := node.AsImportDeclaration()
	if decl == nil || decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral ||
		decl.ModuleSpecifier.AsStringLiteral().Text != "react" || decl.ImportClause == nil {
		return
	}
	clause := decl.ImportClause.AsImportClause()
	if clause == nil {
		return
	}
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		imports[clause.Name().AsIdentifier().Text] = "*"
	}
	if clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		ns := clause.NamedBindings.AsNamespaceImport()
		if ns != nil && ns.Name() != nil {
			imports[ns.Name().AsIdentifier().Text] = "*"
		}
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			return
		}
		for _, spec := range named.Elements.Nodes {
			item := spec.AsImportSpecifier()
			if item == nil || item.Name() == nil {
				continue
			}
			local := item.Name()
			imported := item.PropertyName
			if imported == nil {
				imported = local
			}
			if local.Kind == ast.KindIdentifier && imported.Kind == ast.KindIdentifier && isGenericTypeName(imported.AsIdentifier().Text) {
				imports[local.AsIdentifier().Text] = imported.AsIdentifier().Text
			}
		}
	}
}

func reactGenericPropsType(node *ast.Node, imports map[string]string) *ast.Node {
	ref := node.AsTypeReferenceNode()
	if ref == nil || ref.TypeArguments == nil || len(ref.TypeArguments.Nodes) == 0 || !isReactGenericType(ref.TypeName, imports) {
		return nil
	}
	index := 0
	if genericName, _ := reactGenericTypeName(ref.TypeName, imports); genericName == "forwardRef" || genericName == "ForwardRefRenderFunction" {
		index = 1
	}
	if len(ref.TypeArguments.Nodes) <= index {
		return nil
	}
	return ref.TypeArguments.Nodes[index]
}

func isReactGenericType(name *ast.Node, imports map[string]string) bool {
	_, ok := reactGenericTypeName(name, imports)
	return ok
}

func reactGenericTypeName(name *ast.Node, imports map[string]string) (string, bool) {
	if name == nil {
		return "", false
	}
	if name.Kind == ast.KindIdentifier {
		imported, ok := imports[name.AsIdentifier().Text]
		return imported, ok && imported != "" && imported != "*" && isGenericTypeName(imported)
	}
	if name.Kind != ast.KindQualifiedName {
		return "", false
	}
	qualified := name.AsQualifiedName()
	if qualified == nil || qualified.Left == nil || qualified.Right == nil {
		return "", false
	}
	left := reactutil.EntityNameRightmost(qualified.Left)
	right := reactutil.EntityNameRightmost(qualified.Right)
	if left == nil || right == nil || !isGenericTypeName(right.AsIdentifier().Text) {
		return "", false
	}
	_, ok := imports[left.AsIdentifier().Text]
	return right.AsIdentifier().Text, ok
}

func isGenericTypeName(name string) bool {
	switch name {
	case "ComponentProps", "ComponentPropsWithRef", "ComponentPropsWithoutRef", "forwardRef", "ForwardRefRenderFunction", "VFC", "VoidFunctionComponent", "PropsWithChildren", "SFC", "StatelessComponent", "FunctionComponent", "FC":
		return true
	default:
		return false
	}
}

func validateTypeNode(node *ast.Node, aliases map[string][]*ast.Node, genericImports map[string]string, report func(*ast.Node, string, bool), seen map[*ast.Node]bool) {
	if node == nil || seen[node] {
		return
	}
	seen[node] = true
	switch node.Kind {
	case ast.KindTypeLiteral:
		literal := node.AsTypeLiteralNode()
		if literal != nil && literal.Members != nil {
			for _, member := range literal.Members.Nodes {
				checkPropertySignature(member, report)
			}
		}
	case ast.KindInterfaceDeclaration:
		decl := node.AsInterfaceDeclaration()
		if decl == nil {
			return
		}
		if decl.HeritageClauses != nil {
			for _, clause := range decl.HeritageClauses.Nodes {
				heritage := clause.AsHeritageClause()
				if heritage == nil || heritage.Types == nil {
					continue
				}
				for _, item := range heritage.Types.Nodes {
					extends := item.AsExpressionWithTypeArguments()
					if extends == nil {
						continue
					}
					if extends.Expression == nil || extends.Expression.Kind != ast.KindIdentifier {
						continue
					}
					for _, target := range aliases[extends.Expression.AsIdentifier().Text] {
						validateTypeNode(target, aliases, genericImports, report, seen)
					}
				}
			}
		}
		if decl.Members != nil {
			for _, member := range decl.Members.Nodes {
				checkPropertySignature(member, report)
			}
		}
	case ast.KindTypeReference:
		ref := node.AsTypeReferenceNode()
		if ref == nil {
			return
		}
		if propsType := reactGenericPropsType(node, genericImports); propsType != nil {
			validateTypeNode(propsType, aliases, genericImports, report, seen)
			return
		}
		if ref.TypeName == nil || ref.TypeName.Kind != ast.KindIdentifier {
			return
		}
		for _, target := range aliases[ref.TypeName.AsIdentifier().Text] {
			validateTypeNode(target, aliases, genericImports, report, seen)
		}
	case ast.KindParenthesizedType:
		parenthesized := node.AsParenthesizedTypeNode()
		if parenthesized != nil {
			validateTypeNode(parenthesized.Type, aliases, genericImports, report, seen)
		}
	case ast.KindIntersectionType:
		intersection := node.AsIntersectionTypeNode()
		if intersection != nil && intersection.Types != nil {
			for _, item := range intersection.Types.Nodes {
				validateTypeNode(item, aliases, genericImports, report, seen)
			}
		}
	}
}

func checkPropertySignature(node *ast.Node, report func(*ast.Node, string, bool)) {
	if node == nil || node.Kind != ast.KindPropertySignature {
		return
	}
	property := node.AsPropertySignatureDeclaration()
	if property == nil || property.Name() == nil {
		return
	}
	nameNode := property.Name()
	name, ok := propertyName(nameNode)
	if !ok {
		return
	}
	report(node, name, ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly))
}

func propertyName(nameNode *ast.Node) (string, bool) {
	if nameNode == nil {
		return "", false
	}
	if nameNode.Kind == ast.KindComputedPropertyName {
		expr := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
		if expr == nil {
			return "undefined", true
		}
		if expr.Kind == ast.KindIdentifier {
			return expr.AsIdentifier().Text, true
		}
		if expr.Kind == ast.KindNumericLiteral {
			return authoredNumericLiteralText(expr), true
		}
		if expr.Kind != ast.KindNoSubstitutionTemplateLiteral {
			if name, ok := utils.GetStaticPropertyName(nameNode); ok {
				return name, true
			}
		}
		return "undefined", true
	}
	var name string
	switch nameNode.Kind {
	case ast.KindIdentifier:
		name = nameNode.AsIdentifier().Text
	case ast.KindStringLiteral:
		name = nameNode.AsStringLiteral().Text
	case ast.KindNumericLiteral:
		return authoredNumericLiteralText(nameNode), true
	default:
		if name, ok := utils.GetStaticPropertyName(nameNode); ok {
			return name, true
		}
		return "", false
	}
	return name, true
}

func authoredNumericLiteralText(node *ast.Node) string {
	if sourceFile := ast.GetSourceFileOfNode(node); sourceFile != nil {
		if text := scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, node, false); text != "" {
			return text
		}
	}
	return node.AsNumericLiteral().Text
}
