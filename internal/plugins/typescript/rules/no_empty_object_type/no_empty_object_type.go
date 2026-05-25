package no_empty_object_type

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoEmptyObjectTypeOptions struct {
	AllowInterfaces  string
	AllowObjectTypes string
	AllowWithName    string
}

func parseOptions(options any) NoEmptyObjectTypeOptions {
	opts := NoEmptyObjectTypeOptions{
		AllowInterfaces:  "never",
		AllowObjectTypes: "never",
	}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowInterfaces"].(string); ok {
		opts.AllowInterfaces = v
	}
	if v, ok := optsMap["allowObjectTypes"].(string); ok {
		opts.AllowObjectTypes = v
	}
	if v, ok := optsMap["allowWithName"].(string); ok {
		opts.AllowWithName = v
	}
	return opts
}

func buildNoEmptyInterfaceMessage(option string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "noEmptyInterface",
		Description: "An empty interface declaration allows any non-nullish value, including literals like `0` and `\"\"`.\n" +
			"- If that's what you want, disable this lint rule with an inline comment or configure the '" + option + "' rule option.\n" +
			"- If you want a type meaning \"any object\", you probably want `object` instead.\n" +
			"- If you want a type meaning \"any value\", you probably want `unknown` instead.",
		Data: map[string]string{"option": option},
	}
}

func buildNoEmptyInterfaceWithSuperMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noEmptyInterfaceWithSuper",
		Description: "An interface declaring no members is equivalent to its supertype.",
	}
}

func buildNoEmptyObjectMessage(option string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "noEmptyObject",
		Description: "The `{}` (\"empty object\") type allows any non-nullish value, including literals like `0` and `\"\"`.\n" +
			"- If that's what you want, disable this lint rule with an inline comment or configure the '" + option + "' rule option.\n" +
			"- If you want a type meaning \"any object\", you probably want `object` instead.\n" +
			"- If you want a type meaning \"any value\", you probably want `unknown` instead.",
		Data: map[string]string{"option": option},
	}
}

func buildReplaceEmptyInterfaceMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceEmptyInterface",
		Description: fmt.Sprintf("Replace empty interface with `%s`.", replacement),
		Data:        map[string]string{"replacement": replacement},
	}
}

func buildReplaceEmptyInterfaceWithSuperMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceEmptyInterfaceWithSuper",
		Description: "Replace empty interface with a type alias.",
	}
}

func buildReplaceEmptyObjectTypeMessage(replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaceEmptyObjectType",
		Description: fmt.Sprintf("Replace `{}` with `%s`.", replacement),
		Data:        map[string]string{"replacement": replacement},
	}
}

// getExtendsClause returns the single `extends` heritage clause of an
// interface (or nil if absent).
func getExtendsClause(interfaceDecl *ast.InterfaceDeclaration) *ast.HeritageClause {
	if interfaceDecl.HeritageClauses == nil {
		return nil
	}
	for _, clause := range interfaceDecl.HeritageClauses.Nodes {
		heritageClause := clause.AsHeritageClause()
		if heritageClause == nil {
			continue
		}
		if heritageClause.Token == ast.KindExtendsKeyword {
			return heritageClause
		}
	}
	return nil
}

// isMergedWithClassDeclaration reports whether the interface's name resolves
// to a symbol that also has a class declaration in the same scope (TypeScript
// declaration merging). Mirrors upstream's
// `scope.set.get(name).defs.some(def => def.type === 'ClassDeclaration')`
// check via the tsgo TypeChecker.
func isMergedWithClassDeclaration(ctx rule.RuleContext, nameNode *ast.Node) bool {
	if ctx.TypeChecker == nil || nameNode == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	if symbol == nil {
		return false
	}
	for _, decl := range symbol.Declarations {
		if decl.Kind == ast.KindClassDeclaration {
			return true
		}
	}
	return false
}

// typeParametersText returns the bracketed type-parameter clause text
// (`<T, U>`) or empty string when no type parameters are present. The
// brackets are located via tsgo's scanner rather than by ±1 on the inner
// element ranges so that whitespace / line breaks / trailing commas
// between `<`, the parameters, and `>` are preserved verbatim.
//
// tsgo's `parseDelimitedList` consumes the trailing comma (when present) and
// returns a NodeList whose End() points at `>`'s `TokenFullStart`, so we can
// safely scan from there to find `>` itself — there is no risk of the
// scanner returning the comma's range.
func typeParametersText(ctx rule.RuleContext, interfaceDecl *ast.InterfaceDeclaration) string {
	if interfaceDecl.TypeParameters == nil || len(interfaceDecl.TypeParameters.Nodes) == 0 {
		return ""
	}
	ltRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, interfaceDecl.Name().End())
	gtRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, interfaceDecl.TypeParameters.End())
	return ctx.SourceFile.Text()[ltRange.Pos():gtRange.End()]
}

// interfaceFixRange returns the source range to replace when rewriting an
// `interface` declaration to a `type` alias. The range starts at the
// `interface` keyword (not the first modifier and not any leading trivia),
// which mirrors typescript-eslint's TSInterfaceDeclaration.range[0]
// convention: modifiers AND any trivia between modifiers and the `interface`
// keyword (e.g. `export /* doc */ interface Foo`) are preserved verbatim by
// being outside the replace range.
func interfaceFixRange(ctx rule.RuleContext, interfaceDecl *ast.InterfaceDeclaration, node *ast.Node) core.TextRange {
	pos := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
	if modifiers := interfaceDecl.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
		last := modifiers.Nodes[len(modifiers.Nodes)-1]
		pos = scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, last.End()).Pos()
	}
	return core.NewTextRange(pos, node.End())
}

var NoEmptyObjectTypeRule = rule.CreateRule(rule.Rule{
	Name: "no-empty-object-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		var allowWithNameTester *regexp.Regexp
		if opts.AllowWithName != "" {
			allowWithNameTester, _ = regexp.Compile(opts.AllowWithName)
		}

		listeners := rule.RuleListeners{}

		if opts.AllowInterfaces != "always" {
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()
				if interfaceDecl == nil {
					return
				}
				nameNode := interfaceDecl.Name()
				if nameNode == nil {
					return
				}
				if allowWithNameTester != nil {
					if id := nameNode.AsIdentifier(); id != nil && allowWithNameTester.MatchString(id.Text) {
						return
					}
				}

				if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
					return
				}

				extendClause := getExtendsClause(interfaceDecl)
				extendCount := 0
				if extendClause != nil && extendClause.Types != nil {
					extendCount = len(extendClause.Types.Nodes)
				}

				if extendCount == 1 && opts.AllowInterfaces == "with-single-extends" {
					return
				}
				if extendCount > 1 {
					return
				}

				mergedWithClass := isMergedWithClassDeclaration(ctx, nameNode)

				nameText := utils.TrimmedNodeText(ctx.SourceFile, nameNode)
				typeParam := typeParametersText(ctx, interfaceDecl)
				fixRange := interfaceFixRange(ctx, interfaceDecl, node)

				if extendCount == 0 {
					if mergedWithClass {
						ctx.ReportNode(nameNode, buildNoEmptyInterfaceMessage("allowInterfaces"))
						return
					}
					suggestions := make([]rule.RuleSuggestion, 0, 2)
					for _, replacement := range []string{"object", "unknown"} {
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message: buildReplaceEmptyInterfaceMessage(replacement),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplaceRange(fixRange, fmt.Sprintf("type %s%s = %s", nameText, typeParam, replacement)),
							},
						})
					}
					ctx.ReportNodeWithSuggestions(nameNode, buildNoEmptyInterfaceMessage("allowInterfaces"), suggestions...)
					return
				}

				// extendCount == 1 here.
				if mergedWithClass {
					ctx.ReportNode(nameNode, buildNoEmptyInterfaceWithSuperMessage())
					return
				}

				extendedTypeText := utils.TrimmedNodeText(ctx.SourceFile, extendClause.Types.Nodes[0])
				ctx.ReportNodeWithSuggestions(nameNode, buildNoEmptyInterfaceWithSuperMessage(), rule.RuleSuggestion{
					Message: buildReplaceEmptyInterfaceWithSuperMessage(),
					FixesArr: []rule.RuleFix{
						rule.RuleFixReplaceRange(fixRange, fmt.Sprintf("type %s%s = %s", nameText, typeParam, extendedTypeText)),
					},
				})
			}
		}

		if opts.AllowObjectTypes != "always" {
			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				typeLiteral := node.AsTypeLiteralNode()
				if typeLiteral == nil {
					return
				}
				if typeLiteral.Members != nil && len(typeLiteral.Members.Nodes) > 0 {
					return
				}
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindIntersectionType {
					return
				}
				if allowWithNameTester != nil && parent != nil && parent.Kind == ast.KindTypeAliasDeclaration {
					typeAlias := parent.AsTypeAliasDeclaration()
					if typeAlias != nil {
						aliasName := typeAlias.Name()
						if aliasName != nil {
							if id := aliasName.AsIdentifier(); id != nil && allowWithNameTester.MatchString(id.Text) {
								return
							}
						}
					}
				}

				suggestions := make([]rule.RuleSuggestion, 0, 2)
				for _, replacement := range []string{"object", "unknown"} {
					suggestions = append(suggestions, rule.RuleSuggestion{
						Message: buildReplaceEmptyObjectTypeMessage(replacement),
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, replacement),
						},
					})
				}
				ctx.ReportNodeWithSuggestions(node, buildNoEmptyObjectMessage("allowObjectTypes"), suggestions...)
			}
		}

		return listeners
	},
})
