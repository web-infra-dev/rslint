package new_cap

import (
	_ "embed"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed new_cap.schema.json
var schemaJSON []byte

var defaultCapIsNewExceptions = []string{
	"Array",
	"Boolean",
	"Date",
	"Error",
	"Function",
	"Number",
	"Object",
	"RegExp",
	"String",
	"Symbol",
	"BigInt",
}

type options struct {
	newIsCap                 bool
	capIsNew                 bool
	newIsCapExceptions       map[string]struct{}
	newIsCapExceptionPattern *esregexp.RegExp
	capIsNewExceptions       map[string]struct{}
	capIsNewExceptionPattern *esregexp.RegExp
	properties               bool
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func addStringOptions(set map[string]struct{}, value any) {
	items, _ := value.([]any)
	for _, item := range items {
		if text, ok := item.(string); ok {
			set[text] = struct{}{}
		}
	}
}

func compilePattern(value any) *esregexp.RegExp {
	pattern, _ := value.(string)
	if pattern == "" {
		return nil
	}
	compiled, err := esregexp.Compile(pattern, "u")
	if err != nil {
		return nil
	}
	return compiled
}

func parseOptions(raw []any) options {
	opts := options{
		newIsCap:           true,
		capIsNew:           true,
		newIsCapExceptions: make(map[string]struct{}),
		capIsNewExceptions: stringSet(defaultCapIsNewExceptions),
		properties:         true,
	}
	if len(raw) == 0 {
		return opts
	}
	object, _ := raw[0].(map[string]any)
	if value, ok := object["newIsCap"].(bool); ok {
		opts.newIsCap = value
	}
	if value, ok := object["capIsNew"].(bool); ok {
		opts.capIsNew = value
	}
	if value, ok := object["properties"].(bool); ok {
		opts.properties = value
	}
	addStringOptions(opts.newIsCapExceptions, object["newIsCapExceptions"])
	addStringOptions(opts.capIsNewExceptions, object["capIsNewExceptions"])
	opts.newIsCapExceptionPattern = compilePattern(object["newIsCapExceptionPattern"])
	opts.capIsNewExceptionPattern = compilePattern(object["capIsNewExceptionPattern"])
	return opts
}

type capitalization uint8

const (
	capitalizationNonAlpha capitalization = iota
	capitalizationLower
	capitalizationUpper
)

// getCap mirrors upstream's `str.charAt(0)` classification. JavaScript reads
// one UTF-16 code unit here, so an astral letter begins with a lone high
// surrogate and is consequently non-alphabetic for this rule.
func getCap(name string) capitalization {
	firstRune, size := utf8.DecodeRuneInString(name)
	if size == 0 || firstRune > 0xFFFF {
		return capitalizationNonAlpha
	}
	first := string(firstRune)
	lower := ecmascript.StringToLowerCase(first)
	upper := ecmascript.StringToUpperCase(first)
	if lower == upper {
		return capitalizationNonAlpha
	}
	if first == lower {
		return capitalizationLower
	}
	return capitalizationUpper
}

func calleeName(callee *ast.Node) (string, bool) {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return "", false
	}
	if callee.Kind == ast.KindIdentifier {
		return callee.AsIdentifier().Text, true
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name != nil {
			return name.Text(), true
		}
	case ast.KindElementAccessExpression:
		argument := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		return utils.GetStaticExpressionValue(argument)
	}
	return "", false
}

func calleeReportNode(callee *ast.Node) *ast.Node {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return nil
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		return callee.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		return ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
	default:
		return callee
	}
}

// hasEmptyTypeArguments identifies ts-go's recoverable AST for an empty type
// argument list. typescript-eslint rejects this syntax before rules run, so an
// ESTree-compatible rule must not inspect the recovered call or constructor.
func hasEmptyTypeArguments(node *ast.Node) bool {
	var typeArguments *ast.NodeList
	switch node.Kind {
	case ast.KindCallExpression:
		typeArguments = node.AsCallExpression().TypeArguments
	case ast.KindNewExpression:
		typeArguments = node.AsNewExpression().TypeArguments
	}
	return typeArguments != nil && len(typeArguments.Nodes) == 0
}

func isAllowed(
	ctx rule.RuleContext,
	allowed map[string]struct{},
	callee *ast.Node,
	name string,
	pattern *esregexp.RegExp,
	skipProperties bool,
) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}
	sourceText := utils.TrimmedNodeText(ctx.SourceFile, callee)
	if _, ok := allowed[name]; ok {
		return true
	}
	if _, ok := allowed[sourceText]; ok {
		return true
	}
	if pattern != nil && pattern.TestOrTimeout(sourceText) {
		return true
	}
	if name == "UTC" && ast.IsAccessExpression(callee) {
		object := ast.SkipParentheses(utils.AccessExpressionObject(callee))
		return object != nil && object.Kind == ast.KindIdentifier && object.AsIdentifier().Text == "Date"
	}
	return skipProperties && ast.IsAccessExpression(callee)
}

var lowerMessage = rule.RuleMessage{
	Id:          "lower",
	Description: "A constructor name should not start with a lowercase letter.",
}

var upperMessage = rule.RuleMessage{
	Id:          "upper",
	Description: "A function with a name starting with an uppercase letter should only be used as a constructor.",
}

// https://eslint.org/docs/latest/rules/new-cap
var NewCapRule = rule.Rule{
	Name:   "new-cap",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		listeners := rule.RuleListeners{}

		if opts.newIsCap {
			listeners[ast.KindNewExpression] = func(node *ast.Node) {
				if hasEmptyTypeArguments(node) {
					return
				}
				callee := node.AsNewExpression().Expression
				name, ok := calleeName(callee)
				if !ok || getCap(name) != capitalizationLower ||
					isAllowed(ctx, opts.newIsCapExceptions, callee, name, opts.newIsCapExceptionPattern, !opts.properties) {
					return
				}
				if reportNode := calleeReportNode(callee); reportNode != nil {
					ctx.ReportNode(reportNode, lowerMessage)
				}
			}
		}

		if opts.capIsNew {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				if hasEmptyTypeArguments(node) {
					return
				}
				callee := node.AsCallExpression().Expression
				name, ok := calleeName(callee)
				if !ok || getCap(name) != capitalizationUpper ||
					isAllowed(ctx, opts.capIsNewExceptions, callee, name, opts.capIsNewExceptionPattern, !opts.properties) {
					return
				}
				if reportNode := calleeReportNode(callee); reportNode != nil {
					ctx.ReportNode(reportNode, upperMessage)
				}
			}
		}

		return listeners
	},
}
