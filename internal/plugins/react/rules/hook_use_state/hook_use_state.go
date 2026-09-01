package hook_use_state

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
		info.defaultSpecifier = decl.ImportClause
	}
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return info
	}
	for _, spec := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
		switch importedName(spec) {
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

func isReactImportDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind != ast.KindImportDeclaration {
			continue
		}
		decl := parent.AsImportDeclaration()
		return decl.ModuleSpecifier != nil && decl.ModuleSpecifier.Kind == ast.KindStringLiteral &&
			decl.ModuleSpecifier.AsStringLiteral().Text == "react"
	}
	return false
}

func resolvesToNamedUseState(ctx rule.RuleContext, ident *ast.Node) bool {
	if ctx.Refs == nil || ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	// Components#isReactHookCall requires the local hook spelling to match
	// /^use[A-Z]/, even when it is imported under another name.
	name := ident.AsIdentifier().Text
	if len(name) < 4 || name[:3] != "use" || name[3] < 'A' || name[3] > 'Z' {
		return false
	}
	symbol := ctx.Refs.Resolve(ident)
	if symbol == nil {
		return false
	}
	for _, decl := range symbol.Declarations {
		if decl != nil && decl.Kind == ast.KindImportSpecifier && decl.Pos() < ident.Pos() &&
			importedName(decl) == "useState" && isReactImportDeclaration(decl) {
			return true
		}
	}
	return false
}

func resolvesToDefaultReactImport(ctx rule.RuleContext, defaultSpecifier, ident *ast.Node) bool {
	if ctx.Refs == nil || defaultSpecifier == nil || ident == nil || ident.Kind != ast.KindIdentifier ||
		defaultSpecifier.Pos() >= ident.Pos() {
		return false
	}
	symbol := ctx.Refs.Resolve(ident)
	if symbol == nil {
		return false
	}
	for _, decl := range symbol.Declarations {
		if decl == defaultSpecifier {
			return true
		}
	}
	return false
}

// isReactUseStateCall ports Components#isReactHookCall(node, ["useState"]).
// Parentheses are transparent because ESTree drops them; TS assertion wrappers
// deliberately are not, matching the upstream parser's explicit TS nodes.
func isReactUseStateCall(ctx rule.RuleContext, imports importInfo, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	if ast.IsOptionalChain(callee) {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		return resolvesToNamedUseState(ctx, callee)
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "useState" {
		return false
	}
	return resolvesToDefaultReactImport(ctx, imports.defaultSpecifier, ast.SkipParentheses(access.Expression))
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
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				info := reactImportInfo(node)
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
				if !isReactUseStateCall(ctx, imports, node) {
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
					if value != nil && len(elements) == 1 && valueName != "" && node.AsCallExpression().Arguments != nil && len(node.AsCallExpression().Arguments.Nodes) == 1 {
						argument := node.AsCallExpression().Arguments.Nodes[0]
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
							rule.RuleFixReplace(ctx.SourceFile, pattern, valueName),
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
