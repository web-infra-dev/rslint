package catch_error_name

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "catch-error-name"

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
	runes := []rune(value)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
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

func availableName(identifier *ast.Node, references []*ast.Node, expected string) string {
	for candidate := expected; ; candidate += "_" {
		available := !utils.IsShadowed(identifier, candidate)
		if available {
			for _, reference := range references {
				if utils.IsShadowed(reference, candidate) {
					available = false
					break
				}
			}
		}
		if available {
			return candidate
		}
		if len(candidate) > len(expected)+100 {
			return ""
		}
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
				fixedName := availableName(identifier, references, opts.name)
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
