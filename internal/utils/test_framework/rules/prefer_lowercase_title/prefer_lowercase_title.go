package prefer_lowercase_title

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed prefer_lowercase_title.schema.json
var schemaJSON []byte

// Config configures a prefer-lowercase-title rule for one test framework.
type Config struct {
	Name            string
	DescribeAliases []string
	TestAliases     []string
	ItAliases       []string
	Prepare         func(rule.RuleContext) Runtime
}

// Runtime supplies framework-specific call semantics to the shared rule.
type Runtime struct {
	Parse         func(*ast.Node) *testFramework.ParsedCall
	IsTodo        func(*ast.Node) bool
	DescribeDepth func(*ast.Node) int
	Skip          bool
}

type resolvedOptions struct {
	ignoredNames           map[string]struct{}
	allowedPrefixes        []string
	ignoreTopLevelDescribe bool
	ignoreTodos            bool
}

func firstOptionMap(options []any) map[string]interface{} {
	if len(options) == 0 {
		return nil
	}
	m, ok := options[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}

func boolFromMap(m map[string]interface{}, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func stringSliceFromMap(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func resolveOptions(options []any, cfg Config) resolvedOptions {
	m := firstOptionMap(options)
	if m == nil {
		return resolvedOptions{ignoredNames: map[string]struct{}{}}
	}
	o := resolvedOptions{
		ignoredNames:           map[string]struct{}{},
		allowedPrefixes:        stringSliceFromMap(m, "allowedPrefixes"),
		ignoreTopLevelDescribe: boolFromMap(m, "ignoreTopLevelDescribe", false),
		ignoreTodos:            boolFromMap(m, "ignoreTodos", false),
	}
	for _, ig := range stringSliceFromMap(m, "ignore") {
		switch ig {
		case "describe":
			for _, name := range cfg.DescribeAliases {
				o.ignoredNames[name] = struct{}{}
			}
		case "test":
			for _, name := range cfg.TestAliases {
				o.ignoredNames[name] = struct{}{}
			}
		case "it":
			for _, name := range cfg.ItAliases {
				o.ignoredNames[name] = struct{}{}
			}
		}
	}
	return o
}

// NewRule creates a prefer-lowercase-title rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip {
				return rule.RuleListeners{}
			}
			opts := resolveOptions(options, config)
			var numberOfDescribeBlocks int

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil {
						return
					}

					if parsed.Kind == testFramework.FnKindDescribe {
						numberOfDescribeBlocks++
						if opts.ignoreTopLevelDescribe {
							describeDepth := numberOfDescribeBlocks
							if runtime.DescribeDepth != nil {
								describeDepth = runtime.DescribeDepth(node)
							}
							if describeDepth == 1 {
								return
							}
						}
					} else if parsed.Kind != testFramework.FnKindTest {
						return
					}

					if opts.ignoreTodos {
						isTodo := slices.Contains(parsed.Members, "todo")
						if runtime.IsTodo != nil {
							isTodo = isTodo || runtime.IsTodo(node)
						}
						if isTodo {
							return
						}
					}

					call := node.AsCallExpression()
					if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
						return
					}
					arg0 := call.Arguments.Nodes[0]
					inner := internalUtils.SkipAssertionsAndParens(arg0)
					description, ok := internalUtils.GetStaticStringLiteralValue(inner)
					if !ok {
						return
					}

					runes := []rune(description)
					if len(runes) == 0 {
						return
					}

					for _, prefix := range opts.allowedPrefixes {
						if strings.HasPrefix(description, prefix) {
							return
						}
					}

					firstChar := string(runes[0])
					if ecmascript.StringToLowerCase(firstChar) == firstChar {
						return
					}

					if _, ignored := opts.ignoredNames[parsed.Name]; ignored {
						return
					}

					lowercaseFirstChar := ecmascript.StringToLowerCase(firstChar)

					ctx.ReportNodeWithDeferredFixes(inner, rule.RuleMessage{
						Id:          "unexpectedCase",
						Description: fmt.Sprintf("`%s`s should begin with lowercase", parsed.Name),
						Data:        map[string]string{"method": parsed.Name},
					}, func() []rule.RuleFix {
						r := internalUtils.TrimNodeTextRange(ctx.SourceFile, inner)
						raw := ctx.SourceFile.Text()[r.Pos():r.End()]
						units := internalUtils.ParseJSStringLiteralSource(raw)
						if inner.Kind == ast.KindNoSubstitutionTemplateLiteral {
							units = internalUtils.ParseJSTemplateLiteralSource(raw)
						}
						if len(units) == 0 {
							return nil
						}
						fixRange := r.WithPos(r.Pos() + units[0].Start).WithEnd(r.Pos() + units[0].End)
						return []rule.RuleFix{
							rule.RuleFixReplaceRange(fixRange, lowercaseFirstChar),
						}
					})
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil {
						return
					}
					if parsed.Kind == testFramework.FnKindDescribe {
						numberOfDescribeBlocks--
					}
				},
			}
		},
	}
}
