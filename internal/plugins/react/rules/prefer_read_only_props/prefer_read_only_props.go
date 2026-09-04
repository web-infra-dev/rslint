package prefer_read_only_props

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
	pragma := reactutil.GetReactPragmaFromContext(ctx)
	typeAliases := map[string][]*ast.Node{}
	genericImports := map[string]string{}
	classExpressions := make([]*ast.Node, 0)
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

	// eslint-plugin-react resolves TypeScript declarations by searching the
	// complete Program body, so a component may refer to a type declared later
	// in the file. Imports remain source-order sensitive and are collected by
	// their listener below.
	if ctx.SourceFile != nil {
		ctx.SourceFile.Node.ForEachChild(func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindTypeAliasDeclaration:
				if isTopLevelDeclaration(node) {
					collectTypeAlias(node, typeAliases)
				}
			case ast.KindInterfaceDeclaration:
				if isTopLevelDeclaration(node) && ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsDefault == 0 {
					collectInterface(node, typeAliases)
				}
			}
			return false
		})
	}

	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			if isTopLevelDeclaration(node) {
				collectReactGenericImports(node, genericImports)
			}
		},
		ast.KindClassDeclaration: func(node *ast.Node) {
			if !isReactComponentClass(node, pragma, ctx) {
				return
			}
			if propsType := classPropsTypeArgument(node); propsType != nil {
				validateType(node, propsType)
			}
		},
		ast.KindClassExpression: func(node *ast.Node) {
			classExpressions = append(classExpressions, node)
		},
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			pd := node.AsPropertyDeclaration()
			if pd == nil || pd.Type == nil || reactutil.EnclosingClass(node) == nil {
				return
			}
			classNode := reactutil.EnclosingClass(node)
			if !isReactComponentClass(classNode, pragma, ctx) {
				return
			}
			name := pd.Name()
			if !isPropsClassField(node, name) {
				return
			}
			validateType(classNode, pd.Type)
		},
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			checkFunction(node, ctx.TypeChecker, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindFunctionExpression: func(node *ast.Node) {
			checkFunction(node, ctx.TypeChecker, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindArrowFunction: func(node *ast.Node) {
			checkFunction(node, ctx.TypeChecker, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindMethodDeclaration: func(node *ast.Node) {
			checkFunction(node, ctx.TypeChecker, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindGetAccessor: func(node *ast.Node) {
			checkFunction(node, ctx.TypeChecker, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindSetAccessor: func(node *ast.Node) {
			checkFunction(node, ctx.TypeChecker, pragma, genericImports, wrapperFunctions, validateType)
		},
		ast.KindEndOfFile: func(_ *ast.Node) {
			for _, node := range classExpressions {
				if !isReactComponentClass(node, pragma, ctx) {
					continue
				}
				if propsType := classPropsTypeArgument(node); propsType != nil {
					validateType(node, propsType)
				}
			}
			flushReports()
		},
	}
}

func classPropsTypeArgument(node *ast.Node) *ast.Node {
	heritage := ast.GetClassExtendsHeritageElement(node)
	if heritage == nil {
		return nil
	}
	extends := heritage.AsExpressionWithTypeArguments()
	if extends == nil || extends.TypeArguments == nil || len(extends.TypeArguments.Nodes) == 0 {
		return nil
	}
	propsIndex := 0
	if len(extends.TypeArguments.Nodes) > 2 {
		propsIndex = 1
	}
	propsType := extends.TypeArguments.Nodes[propsIndex]
	for propsType != nil && propsType.Kind == ast.KindParenthesizedType {
		parenthesized := propsType.AsParenthesizedTypeNode()
		if parenthesized == nil {
			return nil
		}
		propsType = parenthesized.Type
	}
	if propsType == nil || propsType.Kind != ast.KindTypeReference {
		return nil
	}
	return propsType
}

func checkFunction(node *ast.Node, tc *checker.Checker, pragma string, genericImports map[string]string, wrappers []reactutil.ComponentWrapperEntry, validateType func(*ast.Node, *ast.Node)) {
	if isAsyncGenerator(node) {
		return
	}
	// The forwardRef arm is checked before component-return heuristics in the
	// upstream collector. Its second type argument is the props type.
	if call := parentCall(node); call != nil && call.TypeArguments != nil &&
		isForwardRefWrapper(call, node, wrappers, pragma, tc) {
		if len(call.TypeArguments.Nodes) >= 2 {
			validateType(node, call.TypeArguments.Nodes[1])
		}
		return
	}

	if !reactutil.IsStatelessReactComponentWithWrappers(node, pragma, tc, wrappers) {
		return
	}
	params := reactutil.FunctionParameters(node)
	if len(params) == 0 {
		return
	}
	hasParameterType := false
	if param := params[0].AsParameterDeclaration(); param != nil && param.Type != nil {
		name := param.Name()
		if isDestructuredParameter(name) || isPropsParameter(name) {
			hasParameterType = true
			validateType(node, param.Type)
		}
	}

	// For `const Component: React.FC<Props> = (...) => ...` without a parameter
	// annotation, the upstream collector uses the variable's annotation.
	parent := reactutil.SkipExpressionWrappersUp(node)
	if !hasParameterType && parent != nil && parent.Kind == ast.KindVariableDeclaration {
		vd := parent.AsVariableDeclaration()
		if vd != nil && vd.Type != nil {
			if propsType := reactGenericPropsType(vd.Type, genericImports); propsType != nil {
				validateType(node, propsType)
			}
		}
	}
}

func isForwardRefWrapper(call *ast.CallExpression, fn *ast.Node, wrappers []reactutil.ComponentWrapperEntry, pragma string, tc *checker.Checker) bool {
	return call != nil && fn != nil && reactutil.MatchesAnyComponentWrapperWithChecker(call.AsNode(), fn, wrappers, pragma, tc) &&
		forwardRefCallName(call, pragma) == "forwardRef"
}

func forwardRefCallName(call *ast.CallExpression, pragma string) string {
	callee := reactutil.SkipExpressionWrappers(call.Expression)
	if callee == nil {
		return ""
	}
	if callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "forwardRef" {
		return "forwardRef"
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	property := callee.AsPropertyAccessExpression()
	object := reactutil.SkipExpressionWrappers(property.Expression)
	name := property.Name()
	if object == nil || object.Kind != ast.KindIdentifier || object.AsIdentifier().Text != pragma ||
		name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "forwardRef" {
		return ""
	}
	return "forwardRef"
}

func isAsyncGenerator(node *ast.Node) bool {
	if node == nil || !ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	default:
		return false
	}
}

func isDestructuredParameter(name *ast.Node) bool {
	return name != nil && name.Kind == ast.KindObjectBindingPattern
}

func isPropsParameter(name *ast.Node) bool {
	if name == nil {
		return false
	}
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text == "props"
	}
	if name.Kind != ast.KindBindingElement {
		return false
	}
	return isPropsParameter(name.AsBindingElement().Name())
}

func isReactComponentClass(node *ast.Node, pragma string, ctx rule.RuleContext) bool {
	return reactutil.ExtendsReactComponent(node, pragma) || hasReactComponentJSDoc(node, ctx)
}

func isPropsClassField(node *ast.Node, name *ast.Node) bool {
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "props" {
		return false
	}
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil {
		return false
	}
	tokens := utils.TokensOfNode(sourceFile, node)
	for i := 0; i < len(tokens) && i < 2; i++ {
		if tokens[i].Kind == ast.KindIdentifier && tokens[i].Text == "props" {
			return true
		}
	}
	return false
}

func hasReactComponentJSDoc(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil {
		return false
	}
	for _, doc := range node.JSDoc(nil) {
		jsDoc := doc.AsJSDoc()
		if jsDoc == nil || jsDoc.Tags == nil {
			continue
		}
		for _, tag := range jsDoc.Tags.Nodes {
			if !ast.IsJSDocAugmentsTag(tag) {
				continue
			}
			augments := tag.AsJSDocAugmentsTag()
			if augments != nil && jsDocTypeRefMatchesComponent(augments.ClassName) {
				return true
			}
		}
	}
	if ctx.SourceFile == nil || ctx.Comments == nil {
		return false
	}
	comments := ctx.Comments.All()
	tokens := utils.TokensOfNode(ctx.SourceFile, node)
	if len(comments) == 0 || len(tokens) == 0 {
		return false
	}
	tokenStart := tokens[0].Start
	index := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= tokenStart })
	if index == 0 {
		return false
	}
	comment := comments[index-1]
	text := ctx.SourceFile.Text()
	if comment.Kind != ast.KindMultiLineCommentTrivia || comment.End() > tokenStart ||
		!strings.HasPrefix(text[comment.Pos():comment.End()], "/**") ||
		!ecmascript.IsBlank(text[comment.End():tokenStart]) {
		return false
	}
	fields := strings.Fields(utils.CommentValue(text, comment))
	for i := 0; i+1 < len(fields); i++ {
		if (fields[i] == "@extends" || fields[i] == "@augments") &&
			(fields[i+1] == "React.Component" || fields[i+1] == "React.PureComponent") {
			return true
		}
	}
	return false
}

func jsDocTypeRefMatchesComponent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "Component" || node.AsIdentifier().Text == "PureComponent"
	case ast.KindExpressionWithTypeArguments:
		return jsDocTypeRefMatchesComponent(node.AsExpressionWithTypeArguments().Expression)
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression().Name()
		return property != nil && property.Kind == ast.KindIdentifier &&
			(property.AsIdentifier().Text == "Component" || property.AsIdentifier().Text == "PureComponent")
	default:
		return false
	}
}

func parentCall(node *ast.Node) *ast.CallExpression {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent != nil && parent.Kind == ast.KindCallExpression {
		return parent.AsCallExpression()
	}
	return nil
}

func isTopLevelDeclaration(node *ast.Node) bool {
	return node != nil && node.Parent != nil && node.Parent.Kind == ast.KindSourceFile
}

func collectTypeAlias(node *ast.Node, types map[string][]*ast.Node) {
	decl := node.AsTypeAliasDeclaration()
	if decl == nil || decl.Name() == nil || decl.Name().Kind != ast.KindIdentifier {
		return
	}
	name := decl.Name().AsIdentifier().Text
	types[name] = append(types[name], decl.Type)
}

func collectInterface(node *ast.Node, types map[string][]*ast.Node) {
	decl := node.AsInterfaceDeclaration()
	if decl == nil || decl.Name() == nil || decl.Name().Kind != ast.KindIdentifier {
		return
	}
	name := decl.Name().AsIdentifier().Text
	types[name] = append(types[name], node)
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
	if qualified.Left.Kind != ast.KindIdentifier || qualified.Right.Kind != ast.KindIdentifier {
		return "", false
	}
	left := qualified.Left
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
		if decl.Members != nil {
			for _, member := range decl.Members.Nodes {
				checkPropertySignature(member, report)
			}
		}
		if decl.HeritageClauses != nil {
			for _, clause := range decl.HeritageClauses.Nodes {
				heritage := clause.AsHeritageClause()
				if heritage == nil || heritage.Types == nil {
					continue
				}
				for _, item := range heritage.Types.Nodes {
					extends := item.AsExpressionWithTypeArguments()
					if extends == nil || extends.Expression == nil {
						continue
					}
					if extends.Expression.Kind == ast.KindIdentifier &&
						extends.Expression.AsIdentifier().Text == "ReturnType" {
						if extends.TypeArguments != nil && len(extends.TypeArguments.Nodes) == 1 {
							validateReturnTypeArgument(extends.TypeArguments.Nodes[0], aliases, genericImports, report, seen)
						}
						continue
					}
					if extends.Expression.Kind != ast.KindIdentifier {
						continue
					}
					for _, target := range aliases[extends.Expression.AsIdentifier().Text] {
						validateTypeNode(target, aliases, genericImports, report, seen)
					}
				}
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
		if ref.TypeName != nil && ref.TypeName.Kind == ast.KindIdentifier &&
			ref.TypeName.AsIdentifier().Text == "ReturnType" &&
			ref.TypeArguments != nil && len(ref.TypeArguments.Nodes) == 1 {
			validateReturnTypeArgument(ref.TypeArguments.Nodes[0], aliases, genericImports, report, seen)
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

func validateReturnTypeArgument(node *ast.Node, aliases map[string][]*ast.Node, genericImports map[string]string, report func(*ast.Node, string, bool), seen map[*ast.Node]bool) {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
	}
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindFunctionType:
		if function := node.AsFunctionTypeNode(); function != nil {
			validateTypeNode(function.Type, aliases, genericImports, report, seen)
		}
	case ast.KindTypeQuery:
		query := node.AsTypeQueryNode()
		if query == nil || query.ExprName == nil {
			return
		}
		if query.ExprName.Kind != ast.KindIdentifier {
			return
		}
		name := query.ExprName.AsIdentifier().Text
		for _, value := range topLevelVariableValues(ast.GetSourceFileOfNode(node), name) {
			value = reactutil.SkipExpressionWrappers(value)
			if !ast.IsFunctionLike(value) {
				continue
			}
			validateReturnValue(value, aliases, genericImports, report, seen)
		}
	}
}

func validateReturnValue(node *ast.Node, aliases map[string][]*ast.Node, genericImports map[string]string, report func(*ast.Node, string, bool), seen map[*ast.Node]bool) {
	if node == nil {
		return
	}
	node = reactutil.SkipExpressionWrappers(node)
	if ast.IsFunctionLike(node) {
		node = functionReturnExpression(node)
	}
	if node == nil {
		return
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		if call != nil && call.TypeArguments != nil {
			for _, typeArgument := range call.TypeArguments.Nodes {
				validateTypeNode(typeArgument, aliases, genericImports, report, seen)
			}
		}
		return
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		object := node.AsObjectLiteralExpression()
		if object == nil || object.Properties == nil {
			return
		}
		for _, property := range object.Properties.Nodes {
			if property == nil || property.Kind != ast.KindSpreadAssignment {
				continue
			}
			spread := property.AsSpreadAssignment()
			if spread != nil {
				validateReturnValue(spread.Expression, aliases, genericImports, report, seen)
			}
		}
	}
}

func functionReturnExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindArrowFunction:
		body := node.AsArrowFunction().Body
		if body != nil && body.Kind == ast.KindBlock {
			return functionBodyReturnExpression(body)
		}
		return reactutil.SkipExpressionWrappers(body)
	case ast.KindFunctionDeclaration:
		return functionBodyReturnExpression(node.AsFunctionDeclaration().Body)
	case ast.KindFunctionExpression:
		return functionBodyReturnExpression(node.AsFunctionExpression().Body)
	}
	return nil
}

func functionBodyReturnExpression(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	block := body.AsBlock()
	if block == nil || block.Statements == nil {
		return nil
	}
	return lastReturnExpression(block.Statements.Nodes)
}

// lastReturnExpression mirrors eslint-plugin-react's ast.loopNodes helper:
// scan statements backwards, and when the trailing relevant statement is a
// switch, recurse only into its final case.
func lastReturnExpression(statements []*ast.Node) *ast.Node {
	for i := len(statements) - 1; i >= 0; i-- {
		statement := statements[i]
		if statement == nil {
			continue
		}
		if statement.Kind == ast.KindReturnStatement {
			return statement.AsReturnStatement().Expression
		}
		if statement.Kind != ast.KindSwitchStatement {
			continue
		}
		switchStatement := statement.AsSwitchStatement()
		if switchStatement == nil || switchStatement.CaseBlock == nil {
			continue
		}
		caseBlock := switchStatement.CaseBlock.AsCaseBlock()
		if caseBlock == nil || caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
			continue
		}
		lastClause := caseBlock.Clauses.Nodes[len(caseBlock.Clauses.Nodes)-1].AsCaseOrDefaultClause()
		if lastClause == nil || lastClause.Statements == nil {
			return nil
		}
		return lastReturnExpression(lastClause.Statements.Nodes)
	}
	return nil
}

func topLevelVariableValues(source *ast.SourceFile, name string) []*ast.Node {
	if source == nil || name == "" {
		return nil
	}
	var values []*ast.Node
	source.Node.ForEachChild(func(node *ast.Node) bool {
		if node.Kind != ast.KindVariableStatement {
			return false
		}
		statement := node.AsVariableStatement()
		if statement == nil || statement.DeclarationList == nil {
			return false
		}
		declarationList := statement.DeclarationList.AsVariableDeclarationList()
		if declarationList == nil || declarationList.Declarations == nil {
			return false
		}
		matched := false
		for _, declaration := range declarationList.Declarations.Nodes {
			if declaration != nil && declaration.Name() != nil && declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Name().AsIdentifier().Text == name {
				matched = true
				break
			}
		}
		if matched {
			for _, declaration := range declarationList.Declarations.Nodes {
				if declaration != nil {
					values = append(values, declaration.Initializer())
				}
			}
		}
		return false
	})
	return values
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
	if !ok || name == "__proto__" {
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
