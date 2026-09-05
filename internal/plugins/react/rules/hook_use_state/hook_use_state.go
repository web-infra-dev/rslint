package hook_use_state

import (
	_ "embed"
	"slices"

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

func firstReactHookReference(referenceScopes map[*ast.Node]*scope.Scope, imports importInfo, ident *ast.Node) *scope.Reference {
	if ident == nil || imports.namedHookNames == nil {
		return nil
	}
	from := referenceScopes[ident]
	if from == nil {
		return nil
	}
	for _, reference := range from.References {
		if _, ok := imports.namedHookNames[reference.Identifier.Text()]; ok {
			return reference
		}
	}
	return nil
}

func resolvesToNamedUseState(referenceScopes map[*ast.Node]*scope.Scope, imports importInfo, ident *ast.Node) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	name := ident.AsIdentifier().Text
	reference := firstReactHookReference(referenceScopes, imports, ident)
	if reference == nil {
		return false
	}
	for _, declaration := range reference.Declarations {
		if declaration != nil && declaration.Name == name && declaration.Kind != scope.DefImport {
			return false
		}
	}
	return true
}

// Components applies the same local-to-imported mapping to bare callees and
// member properties. Its prefix excludes digits, unlike reactutil.IsHookName.
func isUseStateHookName(imports importInfo, name string) bool {
	if len(name) < 4 || name[:3] != "use" || name[3] < 'A' || name[3] > 'Z' {
		return false
	}
	if imported := imports.namedHookNames[name]; imported != "" {
		name = imported
	}
	return name == "useState"
}

func firstReference(referenceScopes map[*ast.Node]*scope.Scope, ident *ast.Node, name string) *scope.Reference {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	from := referenceScopes[ident]
	if from == nil {
		return nil
	}
	for _, reference := range from.References {
		if reference.Identifier != nil && reference.Identifier.Text() == name {
			return reference
		}
	}
	return nil
}

func resolvesToDefaultReactImport(referenceScopes map[*ast.Node]*scope.Scope, imports importInfo, ident *ast.Node) bool {
	if imports.defaultName == "" {
		return false
	}
	reference := firstReference(referenceScopes, ident, imports.defaultName)
	if reference == nil || len(reference.Declarations) == 0 {
		return false
	}
	// Components checks the definition kind, even when a nested import
	// binds the same spelling to a different module.
	for _, declaration := range reference.Declarations {
		if declaration == nil || declaration.Kind != scope.DefImport {
			return false
		}
	}
	return true
}

// isReactUseStateCall ports Components#isReactHookCall(node, ["useState"]).
// Source parentheses and JSDoc casts are transparent; authored TypeScript
// assertions remain explicit nodes in the upstream parser.
func isReactUseStateCall(getReferenceScopes func() map[*ast.Node]*scope.Scope, imports importInfo, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		return len(imports.namedHookNames) > 0 && isUseStateHookName(imports, callee.Text()) &&
			resolvesToNamedUseState(getReferenceScopes(), imports, callee)
	}
	if !ast.IsAccessExpression(callee) {
		return false
	}
	var name *ast.Node
	if ast.IsPropertyAccessExpression(callee) {
		name = callee.Name()
	} else {
		// ESTree also accepts a computed Identifier property, but not a
		// literal or a TS assertion around that identifier.
		name = utils.ESTreeRuntimeExpression(callee.AsElementAccessExpression().ArgumentExpression)
	}
	if name == nil || name.Kind != ast.KindIdentifier || !isUseStateHookName(imports, name.Text()) {
		return false
	}
	receiver := utils.ESTreeRuntimeExpression(utils.AccessExpressionObject(callee))
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.AsIdentifier().Text != imports.defaultName {
		return false
	}
	return resolvesToDefaultReactImport(getReferenceScopes(), imports, receiver)
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
		// The upstream hook matcher can fall back to a bare `useState` spelling
		// after finding another hook-shaped React reference in the scope. Keep
		// that otherwise-unimported reference available to the first-reference
		// lookup without collecting every identifier in the file.
		var referenceScopes map[*ast.Node]*scope.Scope
		getReferenceScopes := func() map[*ast.Node]*scope.Scope {
			if referenceScopes == nil {
				referenceNames := map[string]struct{}{"useState": {}}
				// Collect every import once, including nested and later imports,
				// without making it visible to the hook matcher before its visit.
				// This keeps interleaved imports/calls from rebuilding the analysis.
				if ctx.SourceFile != nil {
					utils.VisitDescendants(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
						if node.Kind != ast.KindImportDeclaration {
							return true
						}
						info := reactImportInfo(node)
						if info.defaultName != "" {
							referenceNames[info.defaultName] = struct{}{}
						}
						for local := range info.namedHookNames {
							referenceNames[local] = struct{}{}
						}
						return false
					})
				}
				manager := scope.Build(ctx.SourceFile, scope.Options{CollectReferences: true, ReferenceNames: referenceNames})
				referenceScopes = make(map[*ast.Node]*scope.Scope, len(manager.References))
				for _, reference := range manager.References {
					referenceScopes[reference.Identifier] = reference.From
				}
			}
			return referenceScopes
		}
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
				if !isReactUseStateCall(getReferenceScopes, imports, node) {
					return
				}
				// ESTree wraps an optional call in ChainExpression, so it is not
				// directly initialized by the variable declarator upstream.
				callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
				if ast.IsOptionalChain(node) || ast.IsOptionalChain(callee) {
					ctx.ReportNode(node, useStateErrorMessage())
					return
				}
				parent := utils.ESTreeParent(node)
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
				if value != nil && setter != nil && len(elements) == 2 && slices.Contains(expected, setterName) {
					return
				}
				// TSESTree includes the declaration's type annotation in the
				// ArrayPattern range used by both reports and suggestions.
				patternRange := utils.TrimNodeTextRange(ctx.SourceFile, pattern)
				if annotation := utils.ESTreeType(parent); annotation != nil {
					patternRange = patternRange.WithEnd(annotation.End())
				}
				if onlyDestructuredValue {
					ctx.ReportRange(patternRange, destructuredStateErrorMessage())
					return
				}

				ctx.ReportRangeWithDeferredSuggestions(patternRange, useStateErrorMessage(), func() []rule.RuleSuggestion {
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
							rule.RuleFixReplaceRange(patternRange, memoValueName),
							rule.RuleFixReplace(ctx.SourceFile, node, memoName+"(() => "+utils.TrimmedNodeText(ctx.SourceFile, utils.ESTreeRuntimeExpression(argument))+
								", [])"),
						)
						suggestions = append(suggestions, rule.RuleSuggestion{Message: suggestMemoMessage(), FixesArr: fixes})
					}
					if len(expected) > 0 {
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message:  suggestPairMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(patternRange, "["+valueName+", "+expected[0]+"]")},
						})
					}
					return suggestions
				})
			},
		}
	},
}
