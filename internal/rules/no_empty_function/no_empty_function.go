package no_empty_function

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-empty-function
var NoEmptyFunctionRule = rule.Rule{
	Name: "no-empty-function",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		check := func(node *ast.Node) {
			body := node.Body()
			if body == nil || body.Kind != ast.KindBlock {
				return
			}
			if len(body.Statements()) != 0 {
				return
			}
			if utils.HasCommentInsideNode(ctx.SourceFile, body) {
				return
			}
			if isAllowedEmptyFunction(node, opts) {
				return
			}

			name := emptyFunctionDisplayName(node)
			bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
			message := rule.RuleMessage{
				Id:          "unexpected",
				Description: fmt.Sprintf("Unexpected empty %s.", name),
				Data:        map[string]string{"name": name},
			}
			suggestionMessage := rule.RuleMessage{
				Id:          "suggestComment",
				Description: fmt.Sprintf("Add comment inside empty %s.", name),
				Data:        map[string]string{"name": name},
			}
			ctx.ReportRangeWithSuggestions(bodyRange, message, rule.RuleSuggestion{
				Message: suggestionMessage,
				FixesArr: []rule.RuleFix{
					rule.RuleFixReplaceRange(core.NewTextRange(bodyRange.Pos()+1, bodyRange.End()-1), " /* empty */ "),
				},
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
	},
}

type noEmptyFunctionOptions struct {
	allow map[string]bool
}

func parseOptions(options any) noEmptyFunctionOptions {
	result := noEmptyFunctionOptions{allow: map[string]bool{}}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return result
	}

	switch allow := optsMap["allow"].(type) {
	case []interface{}:
		for _, item := range allow {
			if s, ok := item.(string); ok {
				result.allow[s] = true
			}
		}
	case []string:
		for _, item := range allow {
			result.allow[item] = true
		}
	}

	return result
}

func isAllowedEmptyFunction(node *ast.Node, opts noEmptyFunctionOptions) bool {
	kind := emptyFunctionKind(node)
	if opts.allow[kind] {
		return true
	}

	if kind == "constructors" {
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) && opts.allow["privateConstructors"] {
			return true
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) && opts.allow["protectedConstructors"] {
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
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsOverride) && opts.allow["overrideMethods"] {
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
		return "constructors"
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
		return "constructor"
	}

	tokens := make([]string, 0, 5)
	parent := parentSkippingParens(node)

	if isClassMemberFunction(node) {
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
			tokens = append(tokens, "static")
		}
		if name := node.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			tokens = append(tokens, "private")
		}
	} else if parent != nil && parent.Kind == ast.KindPropertyDeclaration && parent.Parent != nil && ast.IsClassLike(parent.Parent) {
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
	case parent != nil && parent.Kind == ast.KindPropertyDeclaration && parent.Parent != nil && ast.IsClassLike(parent.Parent):
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
	if node == nil || node.Parent == nil {
		return nil
	}
	// tsgo keeps ParenthesizedExpression nodes that ESLint does not expose
	// here, so recover the property/class-field container around `(fn)`.
	return ast.WalkUpParenthesizedExpressions(node.Parent)
}

func isClassMemberFunction(node *ast.Node) bool {
	return node != nil && node.Parent != nil && ast.IsClassLike(node.Parent) && ast.IsMethodOrAccessor(node)
}

func functionDisplayName(node *ast.Node) string {
	if parent := parentSkippingParens(node); parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
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
		return fmt.Sprintf("'%s'", displayName)
	}
	return ""
}

func ownFunctionDisplayName(node *ast.Node) string {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return fmt.Sprintf("'%s'", name.AsIdentifier().Text)
}
