package prefer_readonly

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func messagePreferReadonly(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferReadonly",
		Description: fmt.Sprintf("Member '%s' is never reassigned; mark it as `readonly`.", name),
	}
}

type options struct {
	onlyInlineLambdas bool
}

func parseOptions(rawOpts any) options {
	opts := options{onlyInlineLambdas: false}
	if rawOpts == nil {
		return opts
	}

	var optsMap map[string]interface{}
	if arr, ok := rawOpts.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = rawOpts.(map[string]interface{})
	}

	if optsMap != nil {
		if v, ok := optsMap["onlyInlineLambdas"].(bool); ok {
			opts.onlyInlineLambdas = v
		}
	}
	return opts
}

const (
	outsideConstructor       = -1
	directlyInsideConstructor = 0
)

type typeToClassRelation int

const (
	relationNone             typeToClassRelation = iota
	relationClass            typeToClassRelation = iota
	relationInstance         typeToClassRelation = iota
	relationClassAndInstance typeToClassRelation = iota
)

type classScope struct {
	checker                                  *checker.Checker
	classType                                *checker.Type
	onlyInlineLambdas                        bool
	constructorScopeDepth                    int
	memberVariableModifications              *utils.Set[string]
	memberVariableWithConstructorModifications *utils.Set[string]
	staticVariableModifications              *utils.Set[string]
	privateModifiableMembers                 map[string]*ast.Node
	privateModifiableStatics                 map[string]*ast.Node
}

func newClassScope(ch *checker.Checker, classNode *ast.Node, onlyInlineLambdas bool) *classScope {
	classType := ch.GetTypeAtLocation(classNode)
	if classType != nil && utils.IsIntersectionType(classType) {
		parts := utils.IntersectionTypeParts(classType)
		if len(parts) > 0 {
			classType = parts[0]
		}
	}

	cs := &classScope{
		checker:                                  ch,
		classType:                                classType,
		onlyInlineLambdas:                        onlyInlineLambdas,
		constructorScopeDepth:                    outsideConstructor,
		memberVariableModifications:              utils.NewSetWithSizeHint[string](4),
		memberVariableWithConstructorModifications: utils.NewSetWithSizeHint[string](4),
		staticVariableModifications:              utils.NewSetWithSizeHint[string](4),
		privateModifiableMembers:                 make(map[string]*ast.Node),
		privateModifiableStatics:                 make(map[string]*ast.Node),
	}

	// Scan class members for property declarations
	members := classNode.Members()
	for _, member := range members {
		if ast.IsPropertyDeclaration(member) {
			cs.addDeclaredVariable(member)
		}
	}

	return cs
}

func (cs *classScope) addDeclaredVariable(node *ast.Node) {
	flags := ast.GetCombinedModifierFlags(node)

	// Must be private (either `private` keyword or `#` private field)
	nameNode := getPropertyName(node)
	isPrivateField := nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier
	isPrivateKeyword := flags&ast.ModifierFlagsPrivate != 0

	if !isPrivateKeyword && !isPrivateField {
		return
	}

	// Skip if already readonly
	if flags&ast.ModifierFlagsReadonly != 0 {
		return
	}

	// Skip accessor properties
	if flags&ast.ModifierFlagsAccessor != 0 {
		return
	}

	// Skip computed property names (non-private identifiers)
	if nameNode != nil && nameNode.Kind == ast.KindComputedPropertyName {
		return
	}

	// Check onlyInlineLambdas option
	if cs.onlyInlineLambdas {
		initializer := getInitializer(node)
		if initializer != nil && !ast.IsArrowFunction(initializer) {
			return
		}
	}

	name := getNameText(nameNode)
	if name == "" {
		return
	}

	if flags&ast.ModifierFlagsStatic != 0 {
		cs.privateModifiableStatics[name] = node
	} else {
		cs.privateModifiableMembers[name] = node
	}
}

func (cs *classScope) addVariableModification(node *ast.Node) {
	// node is a PropertyAccessExpression
	propAccess := node.AsPropertyAccessExpression()
	modifierType := cs.checker.GetTypeAtLocation(propAccess.Expression)

	relation := cs.getTypeToClassRelation(modifierType)

	name := propAccess.Name().Text()

	if relation == relationInstance && cs.constructorScopeDepth == directlyInsideConstructor {
		cs.memberVariableWithConstructorModifications.Add(name)
		return
	}

	if relation == relationInstance || relation == relationClassAndInstance {
		cs.memberVariableModifications.Add(name)
	}
	if relation == relationClass || relation == relationClassAndInstance {
		cs.staticVariableModifications.Add(name)
	}
}

func (cs *classScope) enterConstructor(node *ast.Node) {
	cs.constructorScopeDepth = directlyInsideConstructor

	// Add parameter properties
	funcData := node.FunctionLikeData()
	if funcData != nil && funcData.Parameters != nil {
		for _, param := range funcData.Parameters.Nodes {
			flags := ast.GetCombinedModifierFlags(param)
			if flags&ast.ModifierFlagsPrivate != 0 {
				cs.addDeclaredVariable(param)
			}
		}
	}
}

func (cs *classScope) enterNonConstructor() {
	if cs.constructorScopeDepth != outsideConstructor {
		cs.constructorScopeDepth++
	}
}

func (cs *classScope) exitConstructor() {
	cs.constructorScopeDepth = outsideConstructor
}

func (cs *classScope) exitNonConstructor() {
	if cs.constructorScopeDepth != outsideConstructor {
		cs.constructorScopeDepth--
	}
}

func (cs *classScope) finalizeUnmodifiedPrivateNonReadonlys() []*ast.Node {
	for name := range cs.memberVariableModifications.Keys() {
		delete(cs.privateModifiableMembers, name)
	}
	for name := range cs.staticVariableModifications.Keys() {
		delete(cs.privateModifiableStatics, name)
	}

	result := make([]*ast.Node, 0, len(cs.privateModifiableMembers)+len(cs.privateModifiableStatics))
	for _, node := range cs.privateModifiableMembers {
		result = append(result, node)
	}
	for _, node := range cs.privateModifiableStatics {
		result = append(result, node)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Pos() < result[j].Pos() })
	return result
}

func (cs *classScope) memberHasConstructorModifications(name string) bool {
	return cs.memberVariableWithConstructorModifications.Has(name)
}

func (cs *classScope) getTypeToClassRelation(t *checker.Type) typeToClassRelation {
	if t == nil {
		return relationNone
	}

	// Handle TypeParameter (e.g., 'this' type in class methods)
	if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := checker.Checker_getBaseConstraintOfType(cs.checker, t)
		if constraint != nil {
			return cs.getTypeToClassRelation(constraint)
		}
		return relationNone
	}

	if utils.IsIntersectionType(t) {
		result := relationNone
		for _, subType := range utils.IntersectionTypeParts(t) {
			subResult := cs.getTypeToClassRelation(subType)
			switch subResult {
			case relationClass:
				if result == relationInstance {
					return relationClassAndInstance
				}
				result = relationClass
			case relationInstance:
				if result == relationClass {
					return relationClassAndInstance
				}
				result = relationInstance
			}
		}
		return result
	}

	if utils.IsUnionType(t) {
		parts := utils.UnionTypeParts(t)
		if len(parts) > 0 {
			return cs.getTypeToClassRelation(parts[0])
		}
		return relationNone
	}

	if t.Symbol() == nil || !typeIsOrHasBaseType(cs.checker, t, cs.classType) {
		return relationNone
	}

	if utils.IsObjectType(t) && t.ObjectFlags()&checker.ObjectFlagsClassOrInterface != 0 && !isObjectFlagAnonymous(t) {
		return relationInstance
	}

	return relationClass
}

func typeIsOrHasBaseType(ch *checker.Checker, t *checker.Type, baseType *checker.Type) bool {
	if t == nil || baseType == nil {
		return false
	}
	if t == baseType {
		return true
	}
	// Compare by symbol identity - handles cases where 'this' type
	// and class type are different objects representing the same class
	tSym := t.Symbol()
	baseSym := baseType.Symbol()
	if tSym != nil && baseSym != nil && tSym == baseSym {
		return true
	}
	// Only class/interface types have resolved base types
	if !utils.IsObjectType(t) || t.ObjectFlags()&checker.ObjectFlagsClassOrInterface == 0 {
		return false
	}
	baseTypes := ch.GetBaseTypes(t)
	for _, bt := range baseTypes {
		if typeIsOrHasBaseType(ch, bt, baseType) {
			return true
		}
	}
	return false
}

func isObjectFlagAnonymous(t *checker.Type) bool {
	return t.ObjectFlags()&checker.ObjectFlagsAnonymous != 0
}

// Helper functions

func getPropertyName(node *ast.Node) *ast.Node {
	if ast.IsPropertyDeclaration(node) {
		return node.AsPropertyDeclaration().Name()
	}
	if ast.IsParameter(node) {
		return node.AsParameterDeclaration().Name()
	}
	return nil
}

func getInitializer(node *ast.Node) *ast.Node {
	if ast.IsPropertyDeclaration(node) {
		return node.AsPropertyDeclaration().Initializer
	}
	if ast.IsParameter(node) {
		return node.AsParameterDeclaration().Initializer
	}
	return nil
}

func getTypeAnnotation(node *ast.Node) *ast.Node {
	if ast.IsPropertyDeclaration(node) {
		return node.AsPropertyDeclaration().Type
	}
	if ast.IsParameter(node) {
		return node.AsParameterDeclaration().Type
	}
	return nil
}

func getNameText(nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}
	if nameNode.Kind == ast.KindIdentifier {
		return nameNode.AsIdentifier().Text
	}
	if nameNode.Kind == ast.KindPrivateIdentifier {
		return nameNode.Text()
	}
	return ""
}

func getNameDisplayText(ctx rule.RuleContext, nameNode *ast.Node) string {
	if nameNode == nil {
		return ""
	}
	text := strings.TrimSpace(ctx.SourceFile.Text()[nameNode.Pos():nameNode.End()])
	return text
}

func isConstructor(node *ast.Node) bool {
	return ast.IsConstructorDeclaration(node)
}

func isFunctionScopeBoundary(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionDeclaration,
		ast.KindFunctionExpression, ast.KindMethodDeclaration,
		ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	}
	return false
}

func isDestructuringAssignment(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		parent := current.Parent
		if parent == nil {
			break
		}

		if ast.IsObjectLiteralExpression(parent) ||
			ast.IsArrayLiteralExpression(parent) ||
			ast.IsSpreadAssignment(parent) ||
			(ast.IsSpreadElement(parent) && parent.Parent != nil && ast.IsArrayLiteralExpression(parent.Parent)) {
			current = parent
		} else if ast.IsBinaryExpression(parent) && !ast.IsPropertyAccessExpression(current) {
			bin := parent.AsBinaryExpression()
			return bin.Left == current && bin.OperatorToken.Kind == ast.KindEqualsToken
		} else {
			break
		}
	}
	return false
}

var PreferReadonlyRule = rule.CreateRule(rule.Rule{
	Name: "prefer-readonly",
	Run: func(ctx rule.RuleContext, rawOpts any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		opts := parseOptions(rawOpts)
		var classScopeStack []*classScope

		handlePropertyAccessExpression := func(node *ast.Node, parent *ast.Node, cs *classScope) {
			if ast.IsBinaryExpression(parent) {
				bin := parent.AsBinaryExpression()
				if bin.Left == node && ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
					cs.addVariableModification(node)
				}
				return
			}

			if parent.Kind == ast.KindDeleteExpression || isDestructuringAssignment(node) {
				cs.addVariableModification(node)
				return
			}

			switch parent.Kind {
			case ast.KindPostfixUnaryExpression:
				unary := parent.AsPostfixUnaryExpression()
				if unary != nil && (unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken) {
					cs.addVariableModification(node)
				}
			case ast.KindPrefixUnaryExpression:
				unary := parent.AsPrefixUnaryExpression()
				if unary != nil && (unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken) {
					cs.addVariableModification(node)
				}
			}
		}

		isFunctionScopeBoundaryInStack := func(node *ast.Node) bool {
			if len(classScopeStack) == 0 {
				return false
			}
			if isConstructor(node) {
				return false
			}
			return isFunctionScopeBoundary(node)
		}

		listeners := rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classScopeStack = append(classScopeStack, newClassScope(ctx.TypeChecker, node, opts.onlyInlineLambdas))
			},
			ast.KindClassExpression: func(node *ast.Node) {
				classScopeStack = append(classScopeStack, newClassScope(ctx.TypeChecker, node, opts.onlyInlineLambdas))
			},
			rule.ListenerOnExit(ast.KindClassDeclaration): func(node *ast.Node) {
				if len(classScopeStack) == 0 {
					return
				}
				cs := classScopeStack[len(classScopeStack)-1]
				classScopeStack = classScopeStack[:len(classScopeStack)-1]
				reportViolations(ctx, cs, node)
			},
			rule.ListenerOnExit(ast.KindClassExpression): func(node *ast.Node) {
				if len(classScopeStack) == 0 {
					return
				}
				cs := classScopeStack[len(classScopeStack)-1]
				classScopeStack = classScopeStack[:len(classScopeStack)-1]
				reportViolations(ctx, cs, node)
			},

			// Enter/exit constructors and functions
			ast.KindConstructor: func(node *ast.Node) {
				if len(classScopeStack) > 0 {
					classScopeStack[len(classScopeStack)-1].enterConstructor(node)
				}
			},
			rule.ListenerOnExit(ast.KindConstructor): func(node *ast.Node) {
				if len(classScopeStack) > 0 {
					classScopeStack[len(classScopeStack)-1].exitConstructor()
				}
			},

			ast.KindArrowFunction: func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].enterNonConstructor()
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].exitNonConstructor()
				}
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].enterNonConstructor()
				}
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].exitNonConstructor()
				}
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].enterNonConstructor()
				}
			},
			rule.ListenerOnExit(ast.KindFunctionExpression): func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].exitNonConstructor()
				}
			},
			ast.KindMethodDeclaration: func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].enterNonConstructor()
				}
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].exitNonConstructor()
				}
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].enterNonConstructor()
				}
			},
			rule.ListenerOnExit(ast.KindGetAccessor): func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].exitNonConstructor()
				}
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].enterNonConstructor()
				}
			},
			rule.ListenerOnExit(ast.KindSetAccessor): func(node *ast.Node) {
				if isFunctionScopeBoundaryInStack(node) {
					classScopeStack[len(classScopeStack)-1].exitNonConstructor()
				}
			},

			// Track member access expressions
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if len(classScopeStack) == 0 {
					return
				}
				propAccess := node.AsPropertyAccessExpression()
				if propAccess == nil {
					return
				}
				handlePropertyAccessExpression(node, node.Parent, classScopeStack[len(classScopeStack)-1])
			},
		}

		return listeners
	},
})

func reportViolations(ctx rule.RuleContext, cs *classScope, classNode *ast.Node) {
	violations := cs.finalizeUnmodifiedPrivateNonReadonlys()

	for _, violatingNode := range violations {
		nameNode := getPropertyName(violatingNode)
		if nameNode == nil {
			continue
		}

		nameText := getNameDisplayText(ctx, nameNode)

		// Build the report range: from start of member to end of name
		memberRange := getMemberHeadRange(ctx, violatingNode, nameNode)

		// Build fixes
		fixes := buildFixes(ctx, cs, violatingNode, nameNode)

		ctx.ReportRangeWithFixes(memberRange, messagePreferReadonly(nameText), fixes...)
	}
}

func getMemberHeadRange(ctx rule.RuleContext, node *ast.Node, nameNode *ast.Node) core.TextRange {
	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
	nameEnd := nameNode.End()
	return trimmed.WithEnd(nameEnd)
}

func buildFixes(ctx rule.RuleContext, cs *classScope, node *ast.Node, nameNode *ast.Node) []rule.RuleFix {
	var fixes []rule.RuleFix

	// Insert 'readonly ' before the name
	fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, nameNode, "readonly "))

	// For property declarations without type annotation but with an initializer
	// that has constructor modifications, we may need to add a type annotation
	if ast.IsPropertyDeclaration(node) {
		typeAnno := getTypeAnnotation(node)
		initializer := getInitializer(node)

		if typeAnno == nil && initializer != nil && nameNode.Kind == ast.KindIdentifier {
			identName := nameNode.AsIdentifier().Text
			if cs.memberHasConstructorModifications(identName) {
				violatingType := ctx.TypeChecker.GetTypeAtLocation(node)
				initializerType := ctx.TypeChecker.GetTypeAtLocation(initializer)

				if violatingType != nil && initializerType != nil &&
					isLiteralType(initializerType) && !isLiteralType(violatingType) {
					annotation := ctx.TypeChecker.TypeToString(violatingType)
					if annotation != "" {
						fixes = append(fixes, rule.RuleFixInsertAfter(nameNode, ": "+annotation))
					}
				}
			}
		}
	}

	return fixes
}

func isLiteralType(t *checker.Type) bool {
	if t == nil {
		return false
	}
	flags := t.Flags()
	return flags&checker.TypeFlagsStringLiteral != 0 ||
		flags&checker.TypeFlagsNumberLiteral != 0 ||
		flags&checker.TypeFlagsBooleanLiteral != 0 ||
		flags&checker.TypeFlagsEnumLiteral != 0
}
