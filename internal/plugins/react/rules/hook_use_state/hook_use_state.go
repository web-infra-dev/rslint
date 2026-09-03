package hook_use_state

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed hook_use_state.schema.json
var schemaJSON []byte

const (
	useStateErrorText          = "useState call is not destructured into value + setter pair"
	destructuredStateErrorText = "useState call is not destructured into value + setter pair (you can allow destructuring by enabling \"allowDestructuredState\" option)"
)

func useStateErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "useStateErrorMessage", Description: useStateErrorText}
}

func destructuredStateErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "useStateErrorMessageOrAddOption", Description: destructuredStateErrorText}
}

func suggestPairMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "suggestPair", Description: "Destructure useState call into value + setter pair"}
}

func suggestMemoMessage() rule.RuleMessage {
	return rule.RuleMessage{Id: "suggestMemo", Description: "Replace useState call with useMemo"}
}

// importInfo is the small part of Components.detect's React import cache that
// hook-use-state uses for its suggestions and call recognition.
type importInfo struct {
	defaultName       string
	defaultSpecifier  *ast.Node
	useStateSpecifier *ast.Node
	useMemoSpecifier  *ast.Node
	useMemoName       string
	namedHookNames    map[string]string
}

func importedName(spec *ast.Node) string {
	if spec == nil || spec.Kind != ast.KindImportSpecifier {
		return ""
	}
	s := spec.AsImportSpecifier()
	name := s.Name()
	if s.PropertyName != nil {
		name = s.PropertyName
	}
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func localName(spec *ast.Node) string {
	if spec == nil {
		return ""
	}
	name := spec.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func reactImportInfo(node *ast.Node) importInfo {
	var info importInfo
	if node == nil || node.Kind != ast.KindImportDeclaration {
		return info
	}
	decl := node.AsImportDeclaration()
	if decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral ||
		decl.ModuleSpecifier.AsStringLiteral().Text != "react" || decl.ImportClause == nil {
		return info
	}
	clause := decl.ImportClause.AsImportClause()
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		info.defaultName = clause.Name().AsIdentifier().Text
		info.defaultSpecifier = clause.Name()
	}
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return info
	}
	info.namedHookNames = make(map[string]string)
	for _, spec := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
		imported := importedName(spec)
		local := localName(spec)
		if len(imported) >= 4 && imported[:3] == "use" && imported[3] >= 'A' && imported[3] <= 'Z' && local != "" {
			info.namedHookNames[local] = imported
		}
		switch imported {
		case "useState":
			if info.useStateSpecifier == nil {
				info.useStateSpecifier = spec
			}
		case "useMemo":
			if info.useMemoSpecifier == nil {
				info.useMemoSpecifier = spec
				info.useMemoName = localName(spec)
			}
		}
	}
	return info
}

// hookReferenceIndex keeps the original scope references intact and adds only
// references whose upstream ordering is unambiguous. This matters because
// Components does not resolve the call itself: it inspects the first
// hook-shaped reference in SourceCode#getScope(call).
type hookReferenceIndex struct {
	referenceScopes map[*ast.Node]*scope.Scope
	scopeByBlock    map[*ast.Node]*scope.Scope
	extraReferences map[*scope.Scope][]hookReferenceCandidate
}

type hookReferenceKey struct {
	from *scope.Scope
	name string
}

type hookReferenceCandidate struct {
	from      *scope.Scope
	reference *scope.Reference
	position  int
	tieBreak  int
}

func newHookReferenceIndex(manager *scope.Manager, referenceNames map[string]struct{}) *hookReferenceIndex {
	index := &hookReferenceIndex{
		referenceScopes: make(map[*ast.Node]*scope.Scope, len(manager.References)),
		scopeByBlock:    make(map[*ast.Node]*scope.Scope, len(manager.Scopes)),
	}
	var parameterDecoratorScopes map[*scope.Scope]struct{}
	for _, current := range manager.Scopes {
		// Named function expressions have two scopes for one tsgo node; the
		// later scope is the inner function scope SourceCode#getScope acquires.
		index.scopeByBlock[current.Block] = current
		if current.Kind == scope.KindFunction && len(current.References) == 0 &&
			current.Block.SubtreeFacts()&ast.SubtreeContainsDecorators != 0 {
			for _, parameter := range current.Block.Parameters() {
				if !ast.HasDecorators(parameter) {
					continue
				}
				if parameterDecoratorScopes == nil {
					parameterDecoratorScopes = make(map[*scope.Scope]struct{})
				}
				parameterDecoratorScopes[current] = struct{}{}
				break
			}
		}
	}
	for _, reference := range manager.References {
		index.referenceScopes[reference.Identifier] = reference.From
	}

	// Unlike the shared scope model, typescript-eslint does not create a scope
	// for a non-generic alias or interface. Projecting a partial reference
	// timeline is unsafe, though: computed keys and typeof operands are visited
	// by its value referencer, whose order is context-sensitive. If a target
	// scope contains any such candidate (or a conditional check candidate),
	// leave all aliases for that target untouched instead of recreating its
	// TypeVisitor/PatternVisitor in this rule.
	var firstByKey map[hookReferenceKey]hookReferenceCandidate
	var blockedTypeTargets map[*scope.Scope]bool
	for _, reference := range manager.References {
		from, ok := transparentTypeScopeParent(reference.From)
		if !ok || !scopeNeedsHookExtras(from, parameterDecoratorScopes) {
			continue
		}
		if reference.IsValueReference() || isConditionalCheckReference(reference.Identifier, reference.From.Block) {
			if blockedTypeTargets == nil {
				blockedTypeTargets = make(map[*scope.Scope]bool)
			}
			blockedTypeTargets[from] = true
			continue
		}
		if firstByKey == nil {
			firstByKey = make(map[hookReferenceKey]hookReferenceCandidate)
		}
		keepFirstHookReference(firstByKey, hookReferenceCandidate{
			from:      from,
			reference: reference,
			position:  reference.Identifier.Pos(),
			tieBreak:  reference.Identifier.Pos(),
		})
	}
	for key := range firstByKey {
		if blockedTypeTargets[key.from] {
			delete(firstByKey, key)
		}
	}

	// eslint-scope emits an initialization write for declarations. The shared
	// scope intentionally omits declaration-shaped references, so synthesize
	// only candidate Hook names and retain the earliest merged declaration.
	for _, current := range manager.Scopes {
		for _, variable := range current.Vars {
			if variable == nil || variable.ID == nil || variable.DeclareModifier {
				continue
			}
			if _, relevant := referenceNames[variable.Name]; !relevant {
				continue
			}
			position, initialized := hookInitializationPosition(variable)
			if !initialized {
				continue
			}
			from := index.scopeForDeclaration(variable.ID)
			if !scopeNeedsHookExtras(from, parameterDecoratorScopes) {
				continue
			}
			if firstByKey == nil {
				firstByKey = make(map[hookReferenceKey]hookReferenceCandidate)
			}
			keepFirstHookReference(firstByKey, hookReferenceCandidate{
				from: from,
				reference: &scope.Reference{
					Identifier:   variable.ID,
					From:         from,
					Declarations: variable.Scope.Declarations(variable.Name),
				},
				position: position,
				tieBreak: variable.ID.Pos(),
			})
		}
	}

	if len(firstByKey) != 0 {
		index.extraReferences = make(map[*scope.Scope][]hookReferenceCandidate)
		for _, candidate := range firstByKey {
			index.extraReferences[candidate.from] = append(index.extraReferences[candidate.from], candidate)
		}
	}
	return index
}

func scopeNeedsHookExtras(current *scope.Scope, parameterDecoratorScopes map[*scope.Scope]struct{}) bool {
	if current == nil {
		return false
	}
	if len(current.References) != 0 {
		return true
	}
	if current.Kind == scope.KindClass {
		return ast.HasDecorators(current.Block)
	}
	if current.Kind == scope.KindFunction {
		_, ok := parameterDecoratorScopes[current]
		return ok
	}
	return false
}

func transparentTypeScopeParent(current *scope.Scope) (*scope.Scope, bool) {
	if current == nil || current.Kind != scope.KindType || current.Block == nil || current.Parent == nil ||
		(current.Block.Kind != ast.KindTypeAliasDeclaration && current.Block.Kind != ast.KindInterfaceDeclaration) ||
		current.Block.TypeParameterList() != nil {
		return nil, false
	}
	return current.Parent, true
}

func isConditionalCheckReference(identifier, boundary *ast.Node) bool {
	var child *ast.Node
	for current := identifier; current != nil && current != boundary; current = current.Parent {
		if current.Kind == ast.KindConditionalType &&
			current.AsConditionalTypeNode().CheckType == child {
			return true
		}
		child = current
	}
	return false
}

func keepFirstHookReference(firstByKey map[hookReferenceKey]hookReferenceCandidate, candidate hookReferenceCandidate) {
	key := hookReferenceKey{from: candidate.from, name: candidate.reference.Identifier.Text()}
	if previous, exists := firstByKey[key]; !exists || hookReferenceCandidateLess(candidate, previous) {
		firstByKey[key] = candidate
	}
}

func hookReferenceCandidateLess(left, right hookReferenceCandidate) bool {
	return left.position < right.position || left.position == right.position && left.tieBreak < right.tieBreak
}

func hookInitializationPosition(variable *scope.Variable) (int, bool) {
	hasBindingDefault := false
	for current := variable.ID; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBindingElement:
			hasBindingDefault = hasBindingDefault || current.Initializer() != nil
		case ast.KindVariableDeclaration:
			if variable.Kind != scope.DefVariable {
				if hasBindingDefault {
					return current.Pos(), true
				}
				return 0, false
			}
			declaration := current.AsVariableDeclaration()
			if hasBindingDefault || declaration != nil && declaration.Initializer != nil {
				return current.Pos(), true
			}
			if declaration != nil && utils.IsVarDeclInForInOrOf(current) {
				// Pattern defaults and computed keys precede the loop write.
				list := current.Parent.AsVariableDeclarationList()
				return current.End(), list != nil && list.Declarations != nil &&
					len(list.Declarations.Nodes) != 0 && list.Declarations.Nodes[0] == current
			}
			return 0, false
		case ast.KindParameter:
			return current.Pos(), variable.Kind == scope.DefParameter &&
				(hasBindingDefault || current.Initializer() != nil)
		}
	}
	if hasBindingDefault {
		// A catch binding has no Parameter/VariableDeclaration wrapper.
		return variable.ID.Pos(), true
	}
	return 0, false
}

func (index *hookReferenceIndex) scopeForDeclaration(node *ast.Node) *scope.Scope {
	for current := node; current != nil; current = current.Parent {
		if candidate := index.scopeByBlock[current]; candidate != nil {
			return candidate
		}
	}
	return nil
}

func isDescendantOf(node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func (index *hookReferenceIndex) decoratorAcquiredScope(node *ast.Node) *scope.Scope {
	var decorator *ast.Node
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindDecorator {
			decorator = current
			break
		}
	}
	if decorator == nil || decorator.Parent == nil {
		return nil
	}
	for current := node; current != decorator; current = current.Parent {
		if candidate := index.scopeByBlock[current]; candidate != nil {
			// ESTree keeps a method's computed key outside the FunctionExpression
			// that owns its parameters and body.
			if name := current.Name(); candidate.Kind == scope.KindFunction && name != nil &&
				name.Kind == ast.KindComputedPropertyName && isDescendantOf(node, name) {
				continue
			}
			return candidate
		}
	}
	target := decorator.Parent
	if target.Kind == ast.KindParameter {
		for target = target.Parent; target != nil && !ast.IsFunctionLikeDeclaration(target); target = target.Parent {
		}
		if target == nil || target.Body() == nil || utils.IsInAmbientContext(target) ||
			ast.HasSyntacticModifier(target, ast.ModifierFlagsAbstract) {
			return nil
		}
		return index.scopeByBlock[target]
	}
	for current := target; current != nil; current = current.Parent {
		if ast.IsClassLike(current) {
			return index.scopeByBlock[current]
		}
	}
	return nil
}

func (index *hookReferenceIndex) firstMatchingReference(from *scope.Scope, augmented bool, matches func(string) bool) *scope.Reference {
	if from == nil {
		return nil
	}
	var first *scope.Reference
	firstCandidate := hookReferenceCandidate{}
	for _, reference := range from.References {
		if reference.Identifier != nil && matches(reference.Identifier.Text()) {
			first = reference
			firstCandidate = hookReferenceCandidate{position: reference.Identifier.Pos(), tieBreak: reference.Identifier.Pos()}
			break
		}
	}
	if augmented {
		for _, candidate := range index.extraReferences[from] {
			if (first == nil || hookReferenceCandidateLess(candidate, firstCandidate)) && matches(candidate.reference.Identifier.Text()) {
				first = candidate.reference
				firstCandidate = candidate
			}
		}
	}
	return first
}

func firstReactHookReference(index *hookReferenceIndex, imports importInfo, from *scope.Scope, augmented bool) *scope.Reference {
	if index == nil || imports.namedHookNames == nil {
		return nil
	}
	return index.firstMatchingReference(from, augmented, func(name string) bool {
		return imports.namedHookNames[name] != ""
	})
}

func resolvesToNamedUseState(index *hookReferenceIndex, imports importInfo, ident *ast.Node, from *scope.Scope, augmented bool) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	// Components#isReactHookCall requires the local hook spelling to match
	// /^use[A-Z]/, even when it is imported under another name.
	name := ident.AsIdentifier().Text
	if len(name) < 4 || name[:3] != "use" || name[3] < 'A' || name[3] > 'Z' {
		return false
	}
	reference := firstReactHookReference(index, imports, from, augmented)
	if reference == nil {
		return false
	}
	for _, declaration := range reference.Declarations {
		if declaration != nil && declaration.Name == name && declaration.Kind != scope.DefImport {
			return false
		}
	}
	// Components#isReactHookCall falls back to the local spelling when the
	// first hook-shaped reference is not mapped to a different imported name.
	return imports.namedHookNames[name] == "" || imports.namedHookNames[name] == "useState"
}

func firstReference(index *hookReferenceIndex, from *scope.Scope, name string, augmented bool) *scope.Reference {
	if index == nil {
		return nil
	}
	return index.firstMatchingReference(from, augmented, func(candidate string) bool {
		return candidate == name
	})
}

func resolvesToDefaultReactImport(index *hookReferenceIndex, imports importInfo, from *scope.Scope, augmented bool) bool {
	if imports.defaultName == "" || imports.defaultSpecifier == nil {
		return false
	}
	reference := firstReference(index, from, imports.defaultName, augmented)
	if reference == nil {
		return false
	}
	matchedImport := false
	for _, declaration := range reference.Declarations {
		if declaration == nil || declaration.Kind != scope.DefImport {
			return false
		}
		if declaration.ID == imports.defaultSpecifier {
			matchedImport = true
		}
	}
	return matchedImport
}

// isReactUseStateCall ports Components#isReactHookCall(node, ["useState"]).
// Parentheses are transparent because ESTree drops them; TS assertion wrappers
// deliberately are not, matching the upstream parser's explicit TS nodes.
func isReactUseStateCall(index *hookReferenceIndex, imports importInfo, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	rawCallee := node.AsCallExpression().Expression
	callee := ast.SkipParentheses(rawCallee)
	if callee == nil {
		return false
	}
	// A direct optional member call is still a MemberExpression to ESTree,
	// while parentheses around the optional member make the callee a
	// ChainExpression and prevent the upstream hook matcher from recognizing it.
	if ast.IsOptionalChain(callee) && rawCallee != callee {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		from := index.referenceScopes[callee]
		if acquired := index.decoratorAcquiredScope(node); acquired != nil && acquired != from {
			return firstReactHookReference(index, imports, acquired, true) != nil &&
				resolvesToNamedUseState(index, imports, callee, from, false)
		}
		return resolvesToNamedUseState(index, imports, callee, from, true)
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "useState" {
		return false
	}
	receiver := ast.SkipParentheses(access.Expression)
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.AsIdentifier().Text != imports.defaultName {
		return false
	}
	from := index.referenceScopes[receiver]
	if acquired := index.decoratorAcquiredScope(node); acquired != nil && acquired != from {
		return firstReference(index, acquired, imports.defaultName, true) != nil &&
			resolvesToDefaultReactImport(index, imports, from, false)
	}
	return resolvesToDefaultReactImport(index, imports, from, true)
}

func isDestructuringBinding(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBindingElement {
		return false
	}
	element := node.AsBindingElement()
	if element.Initializer != nil || element.DotDotDotToken != nil {
		return false
	}
	name := element.Name()
	return name != nil && (name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern)
}

func identifierBindingName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindBindingElement {
		return ""
	}
	element := node.AsBindingElement()
	if element.Initializer != nil || element.DotDotDotToken != nil {
		return ""
	}
	name := element.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func isPresentBinding(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBindingElement {
		return false
	}
	return node.AsBindingElement().Name() != nil
}

func expectedSetterNames(value string) []string {
	end := 0
	for end < len(value) && value[end] >= 'a' && value[end] <= 'z' {
		end++
	}
	if end == 0 {
		return nil
	}
	prefix, suffix := value[:end], value[end:]
	upperFirst := string(prefix[0]-'a'+'A') + prefix[1:]
	upperAll := ecmascript.StringToUpperCase(prefix)
	return []string{"set" + upperFirst + suffix, "set" + upperAll + suffix}
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// HookUseStateRule ensures that a React useState call is destructured into a
// symmetric [value, setValue] pair.
var HookUseStateRule = rule.Rule{
	Name:   "react/hook-use-state",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowDestructuredState := false
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				allowDestructuredState, _ = option["allowDestructuredState"].(bool)
			}
		}

		imports := importInfo{}
		namedHookNames := make(map[string]struct{})
		if ctx.SourceFile != nil && ctx.SourceFile.Statements != nil {
			for _, statement := range ctx.SourceFile.Statements.Nodes {
				info := reactImportInfo(statement)
				if info.defaultName != "" {
					namedHookNames[info.defaultName] = struct{}{}
				}
				for local := range info.namedHookNames {
					namedHookNames[local] = struct{}{}
				}
			}
		}
		// The upstream hook matcher can fall back to a bare `useState` spelling
		// after finding another hook-shaped React reference in the scope. Keep
		// that otherwise-unimported reference available to the first-reference
		// lookup without collecting every identifier in the file.
		namedHookNames["useState"] = struct{}{}
		manager := scope.Build(ctx.SourceFile, scope.Options{CollectReferences: true, ReferenceNames: namedHookNames})
		references := newHookReferenceIndex(manager, namedHookNames)
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				info := reactImportInfo(node)
				if imports.namedHookNames == nil {
					imports.namedHookNames = make(map[string]string)
				}
				for local, imported := range info.namedHookNames {
					imports.namedHookNames[local] = imported
				}
				if imports.defaultName == "" && info.defaultName != "" {
					imports.defaultName = info.defaultName
					imports.defaultSpecifier = info.defaultSpecifier
				}
				if imports.useStateSpecifier == nil && info.useStateSpecifier != nil {
					imports.useStateSpecifier = info.useStateSpecifier
				}
				if imports.useMemoSpecifier == nil && info.useMemoSpecifier != nil {
					imports.useMemoSpecifier = info.useMemoSpecifier
					imports.useMemoName = info.useMemoName
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if !isReactUseStateCall(references, imports, node) {
					return
				}
				// ESTree wraps an optional call in ChainExpression, so it is not
				// directly initialized by the variable declarator upstream.
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if ast.IsOptionalChain(node) || ast.IsOptionalChain(callee) {
					ctx.ReportNode(node, useStateErrorMessage())
					return
				}
				parent := node.Parent
				for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
					parent = parent.Parent
				}
				if parent != nil && parent.Kind == ast.KindReturnStatement {
					return
				}
				if parent == nil || parent.Kind != ast.KindVariableDeclaration {
					ctx.ReportNode(node, useStateErrorMessage())
					return
				}
				decl := parent.AsVariableDeclaration()
				pattern := decl.Name()
				if pattern == nil || pattern.Kind != ast.KindArrayBindingPattern {
					ctx.ReportNode(node, useStateErrorMessage())
					return
				}
				elements := pattern.AsBindingPattern().Elements.Nodes
				var value, setter *ast.Node
				if len(elements) > 0 {
					value = elements[0]
				}
				if len(elements) > 1 {
					setter = elements[1]
				}
				onlyDestructuredValue := isDestructuringBinding(value) && !isDestructuringBinding(setter)
				if allowDestructuredState && onlyDestructuredValue {
					return
				}
				valueName := identifierBindingName(value)
				setterName := identifierBindingName(setter)
				expected := expectedSetterNames(valueName)
				if value != nil && setter != nil && len(elements) == 2 && contains(expected, setterName) {
					return
				}
				if onlyDestructuredValue {
					ctx.ReportNode(pattern, destructuredStateErrorMessage())
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(pattern, useStateErrorMessage(), func() []rule.RuleSuggestion {
					suggestions := make([]rule.RuleSuggestion, 0, 2)
					if isPresentBinding(value) && len(elements) == 1 && node.AsCallExpression().Arguments != nil && len(node.AsCallExpression().Arguments.Nodes) == 1 {
						argument := node.AsCallExpression().Arguments.Nodes[0]
						memoValueName := valueName
						if memoValueName == "" {
							// AssignmentPattern and RestElement have an undefined
							// ESTree `name`, which becomes the string "undefined"
							// when the upstream fixer interpolates it.
							memoValueName = "undefined"
						}
						memoName := imports.useMemoName
						var importFix *rule.RuleFix
						if memoName == "" {
							if imports.defaultName != "" {
								memoName = imports.defaultName + ".useMemo"
							} else {
								memoName = "useMemo"
							}
						}
						if imports.useStateSpecifier != nil && (imports.useMemoSpecifier == nil || imports.defaultName != "") {
							fix := rule.RuleFixInsertAfter(imports.useStateSpecifier, ", useMemo")
							importFix = &fix
						}
						fixes := make([]rule.RuleFix, 0, 3)
						if importFix != nil {
							fixes = append(fixes, *importFix)
						}
						fixes = append(fixes,
							rule.RuleFixReplace(ctx.SourceFile, pattern, memoValueName),
							rule.RuleFixReplace(ctx.SourceFile, node, memoName+"(() => "+utils.TrimmedNodeText(ctx.SourceFile, ast.SkipParentheses(argument))+
								", [])"),
						)
						suggestions = append(suggestions, rule.RuleSuggestion{Message: suggestMemoMessage(), FixesArr: fixes})
					}
					if len(expected) > 0 {
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message:  suggestPairMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, pattern, "["+valueName+", "+expected[0]+"]")},
						})
					}
					return suggestions
				})
			},
		}
	},
}
