package catch_error_name

import (
	_ "embed"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "catch-error-name"

var reservedWords = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "enum": true, "export": true, "extends": true, "false": true,
	"finally": true, "for": true, "function": true, "if": true, "import": true,
	"in": true, "instanceof": true, "new": true, "null": true, "return": true,
	"super": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "while": true,
	"with": true, "as": true, "implements": true, "interface": true, "let": true,
	"package": true, "private": true, "protected": true, "public": true, "static": true,
	"yield": true, "any": true, "boolean": true, "constructor": true, "declare": true,
	"get": true, "module": true, "require": true, "number": true, "set": true,
	"string": true, "symbol": true, "type": true, "from": true, "of": true,
}

//go:embed catch_error_name.schema.json
var schemaJSON []byte

type options struct {
	name   string
	ignore []*regexp2.Regexp
}

func parseOptions(raw []any) options {
	result := options{name: "error"}
	if len(raw) == 0 {
		return result
	}
	values, _ := raw[0].(map[string]any)
	if name, ok := values["name"].(string); ok {
		result.name = name
	}
	if patterns, ok := values["ignore"].([]any); ok {
		for _, value := range patterns {
			pattern, ok := value.(string)
			if !ok {
				continue
			}
			if re, err := utils.CompileRegexp2(pattern, utils.JSUnicodeRegexOptions); err == nil {
				result.ignore = append(result.ignore, re)
			}
		}
	}
	return result
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	first, size := utf8.DecodeRuneInString(value)
	// Upstream uppercases `charAt(0)`, a single UTF-16 code unit. For an
	// astral character that unit is only the high surrogate and has no case
	// mapping, so the string stays unchanged.
	if first > 0xffff {
		return value
	}
	return strings.ToUpper(string(first)) + value[size:]
}

func (o options) allows(name string) bool {
	if name == o.name || strings.HasSuffix(name, o.name) || strings.HasSuffix(name, upperFirst(o.name)) {
		return true
	}
	for _, re := range o.ignore {
		if utils.Regexp2MatchString(re, name) {
			return true
		}
	}
	return false
}

func bindingDeclaration(identifier *ast.Node) *ast.Node {
	if identifier == nil || identifier.Parent == nil ||
		(identifier.Parent.Kind != ast.KindParameter && identifier.Parent.Kind != ast.KindVariableDeclaration) {
		return nil
	}
	declaration := identifier.Parent
	if declaration.Name() != identifier {
		return nil
	}
	return declaration
}

func isCatchParameter(identifier *ast.Node) bool {
	declaration := bindingDeclaration(identifier)
	if declaration == nil || declaration.Parent == nil || declaration.Parent.Kind != ast.KindCatchClause {
		return false
	}
	clause := declaration.Parent.AsCatchClause()
	return clause != nil && clause.VariableDeclaration == declaration
}

func isPromiseCatchParameter(identifier *ast.Node) bool {
	declaration := bindingDeclaration(identifier)
	if declaration == nil || declaration.Kind != ast.KindParameter || declaration.Parent == nil {
		return false
	}
	function := declaration.Parent
	if function.Kind != ast.KindArrowFunction && function.Kind != ast.KindFunctionExpression {
		return false
	}
	parameters := function.Parameters()
	if len(parameters) == 0 || unicornutil.PlainParameterIdentifier(parameters[0]) != identifier || function.Parent == nil {
		return false
	}

	argumentsLength := 1
	method := "catch"
	if match, ok := unicornutil.MatchDotMethodCall(function.Parent, unicornutil.DotMethodCallOptions{
		Method:              method,
		ArgumentsLength:     &argumentsLength,
		AllowOptionalMember: true,
	}); ok && match.Call.Arguments()[0] == function {
		return true
	}

	argumentsLength = 2
	method = "then"
	match, ok := unicornutil.MatchDotMethodCall(function.Parent, unicornutil.DotMethodCallOptions{
		Method:              method,
		ArgumentsLength:     &argumentsLength,
		AllowOptionalMember: true,
	})
	return ok && match.Call.Arguments()[1] == function
}

func isValidVariableName(name string) bool {
	return scanner.IsValidIdentifier(name) && !reservedWords[name]
}

func handlerBody(identifier *ast.Node) *ast.Node {
	declaration := bindingDeclaration(identifier)
	if declaration == nil || declaration.Parent == nil {
		return nil
	}
	parent := declaration.Parent
	if parent.Kind == ast.KindCatchClause {
		return parent.AsCatchClause().Block
	}
	return parent.Body()
}

func isLexicalScope(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsFunctionLikeDeclaration(node) {
		return true
	}
	switch node.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindCatchClause, ast.KindCaseBlock,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	default:
		return false
	}
}

func nearestLexicalScope(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, isLexicalScope)
}

func relevantScopes(identifier *ast.Node, references []*ast.Node, body *ast.Node) map[*ast.Node]bool {
	result := map[*ast.Node]bool{}
	declaration := bindingDeclaration(identifier)
	var boundary *ast.Node
	if declaration != nil {
		boundary = declaration.Parent
	}
	if boundary != nil {
		result[boundary] = true
	}
	if body != nil {
		result[body] = true
	}
	for _, reference := range references {
		for current := nearestLexicalScope(reference); current != nil; current = nearestLexicalScope(current) {
			result[current] = true
			if current == boundary || current == body {
				break
			}
		}
	}
	return result
}

func bodyHasNameConflict(ctx rule.RuleContext, body *ast.Node, scopes map[*ast.Node]bool, originalSymbol *ast.Symbol, candidate string) bool {
	conflict := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || conflict {
			return
		}
		if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == candidate {
			if scopes[nearestLexicalScope(node)] {
				if symbol := utils.BindingNameSymbol(node); symbol != nil {
					conflict = symbol != originalSymbol
				} else if !utils.IsNonReferenceIdentifier(node) {
					conflict = ctx.Refs.Resolve(node) != originalSymbol
				}
			}
			if conflict {
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return conflict
		})
	}
	walk(body)
	return conflict
}

func availableName(ctx rule.RuleContext, identifier *ast.Node, references []*ast.Node, expected string) string {
	candidate := expected
	if !isValidVariableName(candidate) {
		candidate += "_"
		if !isValidVariableName(candidate) {
			return ""
		}
	}
	originalSymbol := bindingDeclaration(identifier).Symbol()
	body := handlerBody(identifier)
	scopes := relevantScopes(identifier, references, body)
	for {
		available := !utils.IsShadowed(identifier, candidate)
		if available {
			for _, reference := range references {
				if utils.IsShadowed(reference, candidate) {
					available = false
					break
				}
			}
		}
		if available && bodyHasNameConflict(ctx, body, scopes, originalSymbol, candidate) {
			available = false
		}
		if available {
			return candidate
		}
		candidate += "_"
	}
}

func renameFixes(ctx rule.RuleContext, identifier *ast.Node, references []*ast.Node, name string) []rule.RuleFix {
	fixes := make([]rule.RuleFix, 0, len(references)+1)
	fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, identifier, name))
	for _, reference := range references {
		replacement := name
		if parent := reference.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == reference {
			replacement = reference.Text() + ": " + name
		}
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, reference, replacement))
	}
	return fixes
}

var CatchErrorNameRule = rule.Rule{
	Name:   "unicorn/catch-error-name",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindIdentifier: func(identifier *ast.Node) {
				if !isCatchParameter(identifier) && !isPromiseCatchParameter(identifier) {
					return
				}
				originalName := identifier.AsIdentifier().Text
				trimmedName := strings.TrimRight(originalName, "_")
				if opts.allows(originalName) || opts.allows(trimmedName) {
					return
				}
				declaration := bindingDeclaration(identifier)
				if declaration == nil || ctx.Refs == nil || declaration.Symbol() == nil {
					return
				}
				references := ctx.Refs.References(declaration.Symbol())
				if originalName == "_" && len(references) == 0 {
					return
				}
				fixedName := availableName(ctx, identifier, references, opts.name)
				messageName := fixedName
				if messageName == "" {
					messageName = opts.name
				}
				message := rule.RuleMessage{
					Id:          messageID,
					Description: fmt.Sprintf("The catch parameter `%s` should be named `%s`.", originalName, messageName),
					Data:        map[string]string{"originalName": originalName, "fixedName": messageName},
				}
				ctx.ReportNodeWithDeferredFixes(identifier, message, func() []rule.RuleFix {
					if fixedName == "" {
						return nil
					}
					return renameFixes(ctx, identifier, references, fixedName)
				})
			},
		}
	},
}
