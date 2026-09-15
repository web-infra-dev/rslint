package no_empty_function

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_empty_function.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-empty-function
var NoEmptyFunctionRule = rule.Rule{
	Name:   "no-empty-function",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	return runWithMessageData(ctx, options, true)
}

// RunTSESLint runs the shared implementation for the typescript-eslint
// extension. ESLint uses report data only to interpolate the already formatted
// message and does not expose it in lint results, so the extension avoids a
// per-diagnostic Go map that its previous implementation also did not provide.
func RunTSESLint(ctx rule.RuleContext, options []any) rule.RuleListeners {
	return runWithMessageData(ctx, options, false)
}

func runWithMessageData(ctx rule.RuleContext, options []any, includeMessageData bool) rule.RuleListeners {
	opts := parseOptions(options)

	check := func(node *ast.Node) {
		body := node.Body()
		if body == nil || body.Kind != ast.KindBlock {
			return
		}
		if len(body.Statements()) != 0 {
			return
		}
		bodyRange := core.NewTextRange(scanner.SkipTrivia(ctx.SourceFile.Text(), body.Pos()), body.End())
		innerRange := core.NewTextRange(bodyRange.Pos(), bodyRange.Pos())
		if bodyRange.End() > bodyRange.Pos()+1 {
			innerRange = core.NewTextRange(bodyRange.Pos()+1, bodyRange.End()-1)
		}
		hasComment := false
		if ctx.Comments != nil {
			hasComment = utils.HasCommentInSpan(ctx.Comments.All(), innerRange.Pos(), innerRange.End())
		} else {
			hasComment = utils.HasCommentInsideNode(ctx.SourceFile, body)
		}
		if hasComment {
			return
		}
		if isAllowedEmptyFunction(node, opts) {
			return
		}

		name := emptyFunctionDisplayName(node)
		message := rule.RuleMessage{
			Id:          "unexpected",
			Description: "Unexpected empty " + name + ".",
		}
		if includeMessageData {
			message.Data = map[string]string{"name": name}
		}
		ctx.ReportRangeWithDeferredSuggestions(bodyRange, message, func() []rule.RuleSuggestion {
			suggestionMessage := rule.RuleMessage{
				Id:          "suggestComment",
				Description: "Add comment inside empty " + name + ".",
			}
			if includeMessageData {
				suggestionMessage.Data = map[string]string{"name": name}
			}
			return []rule.RuleSuggestion{{
				Message: suggestionMessage,
				FixesArr: []rule.RuleFix{
					rule.RuleFixReplaceRange(innerRange, " /* empty */ "),
				},
			}}
		})
	}

	return rule.RuleListeners{
		ast.KindArrowFunction:       check,
		ast.KindConstructor:         check,
		ast.KindFunctionDeclaration: check,
		ast.KindFunctionExpression:  check,
		ast.KindGetAccessor:         check,
		ast.KindMethodDeclaration:   check,
		ast.KindSetAccessor:         check,
	}
}

type noEmptyFunctionOptions struct {
	allow map[string]bool
}

func parseOptions(options []any) noEmptyFunctionOptions {
	var opts noEmptyFunctionOptions
	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]any)
	allow, _ := m["allow"].([]any)
	if len(allow) != 0 {
		opts.allow = make(map[string]bool, len(allow))
	}
	for _, item := range allow {
		if s, ok := item.(string); ok {
			opts.allow[s] = true
		}
	}
	return opts
}

func isAllowedEmptyFunction(node *ast.Node, opts noEmptyFunctionOptions) bool {
	kind := emptyFunctionKind(node)
	if opts.allow[kind] {
		return true
	}
	modifierFlags := utils.ESTreeModifierFlags(node)

	if kind == "constructors" {
		if modifierFlags&ast.ModifierFlagsPrivate != 0 && opts.allow["privateConstructors"] {
			return true
		}
		if modifierFlags&ast.ModifierFlagsProtected != 0 && opts.allow["protectedConstructors"] {
			return true
		}
		if hasParameterProperty(node) {
			return true
		}
	}

	if isMethodOrAccessorKind(kind) {
		if ast.HasDecorators(node) && opts.allow["decoratedFunctions"] {
			return true
		}
		if modifierFlags&ast.ModifierFlagsOverride != 0 && opts.allow["overrideMethods"] {
			return true
		}
	}

	return false
}

func emptyFunctionKind(node *ast.Node) string {
	var kind string
	switch node.Kind {
	case ast.KindArrowFunction:
		return "arrowFunctions"
	case ast.KindConstructor:
		if !ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
			return "constructors"
		}
		// ESTree represents `static constructor()` as an ordinary method.
		// tsgo keeps ConstructorDeclaration as the node kind, so classify the
		// static spelling explicitly before applying async/generator prefixes.
		kind = "methods"
	case ast.KindGetAccessor:
		kind = "getters"
	case ast.KindSetAccessor:
		kind = "setters"
	case ast.KindMethodDeclaration:
		kind = "methods"
	default:
		kind = "functions"
	}

	flags := ast.GetFunctionFlags(node)
	switch {
	case flags&ast.FunctionFlagsGenerator != 0:
		return "generator" + utils.UpperCaseFirstASCII(kind)
	case flags&ast.FunctionFlagsAsync != 0:
		return "async" + utils.UpperCaseFirstASCII(kind)
	default:
		return kind
	}
}

func isMethodOrAccessorKind(kind string) bool {
	switch kind {
	case "getters", "setters", "methods", "generatorMethods", "asyncMethods":
		return true
	default:
		return false
	}
}

func hasParameterProperty(node *ast.Node) bool {
	if node.Kind != ast.KindConstructor {
		return false
	}
	for _, param := range node.Parameters() {
		if ast.IsParameterPropertyDeclaration(param, node) {
			return true
		}
	}
	return false
}

// emptyFunctionDisplayName mirrors ESLint's getFunctionNameWithKind for this
// rule's message. It intentionally does not recover outer variable names for
// anonymous function expressions because upstream reports those as plain
// "function" / "arrow function".
func emptyFunctionDisplayName(node *ast.Node) string {
	if node.Kind == ast.KindConstructor {
		if !ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
			return "constructor"
		}
		return utils.GetFunctionNameWithKindCore(node)
	}

	tokens := make([]string, 0, 6)
	parent := parentSkippingParens(node)

	if isClassMemberFunction(node) {
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
			tokens = append(tokens, "static")
		}
		if name := node.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			tokens = append(tokens, "private")
		}
	} else if isClassFieldFunction(parent) {
		if ast.HasSyntacticModifier(parent, ast.ModifierFlagsStatic) {
			tokens = append(tokens, "static")
		}
		if name := parent.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			tokens = append(tokens, "private")
		}
	}

	flags := ast.GetFunctionFlags(node)
	if flags&ast.FunctionFlagsAsync != 0 {
		tokens = append(tokens, "async")
	}
	if flags&ast.FunctionFlagsGenerator != 0 {
		tokens = append(tokens, "generator")
	}

	switch {
	case node.Kind == ast.KindGetAccessor:
		tokens = append(tokens, "getter")
	case node.Kind == ast.KindSetAccessor:
		tokens = append(tokens, "setter")
	case node.Kind == ast.KindMethodDeclaration:
		tokens = append(tokens, "method")
	case parent != nil && parent.Kind == ast.KindPropertyAssignment:
		tokens = append(tokens, "method")
	case isClassFieldFunction(parent):
		tokens = append(tokens, "method")
	default:
		if node.Kind == ast.KindArrowFunction {
			tokens = append(tokens, "arrow")
		}
		tokens = append(tokens, "function")
	}

	if name := functionDisplayName(node); name != "" {
		tokens = append(tokens, name)
	}

	return strings.Join(tokens, " ")
}

func parentSkippingParens(node *ast.Node) *ast.Node {
	return utils.ESTreeParent(node)
}

func isClassMemberFunction(node *ast.Node) bool {
	return node != nil && node.Parent != nil && ast.IsClassLike(node.Parent) && ast.IsMethodOrAccessor(node)
}

func isClassFieldFunction(parent *ast.Node) bool {
	return parent != nil &&
		parent.Kind == ast.KindPropertyDeclaration &&
		parent.Parent != nil &&
		ast.IsClassLike(parent.Parent) &&
		!ast.HasSyntacticModifier(parent, ast.ModifierFlagsAccessor)
}

func functionDisplayName(node *ast.Node) string {
	if parent := parentSkippingParens(node); parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
			if parent.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(parent, ast.ModifierFlagsAccessor) {
				if node.Kind == ast.KindFunctionExpression {
					return ownFunctionDisplayName(node)
				}
				return ""
			}
			if name := memberDisplayName(parent.Name()); name != "" {
				return name
			}
			if node.Kind == ast.KindFunctionExpression {
				return ownFunctionDisplayName(node)
			}
			return ""
		}
	}

	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if display := memberDisplayName(node.Name()); display != "" {
			return display
		}
		if node.Kind == ast.KindMethodDeclaration {
			return ownFunctionDisplayName(node)
		}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return ownFunctionDisplayName(node)
	}

	return ""
}

func memberDisplayName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	if name.Kind == ast.KindPrivateIdentifier {
		return utils.GetPropertyDisplayName(name)
	}
	if displayName := utils.GetPropertyDisplayName(name); displayName != "" {
		return "'" + displayName + "'"
	}
	return ""
}

func ownFunctionDisplayName(node *ast.Node) string {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return "'" + name.AsIdentifier().Text + "'"
}
