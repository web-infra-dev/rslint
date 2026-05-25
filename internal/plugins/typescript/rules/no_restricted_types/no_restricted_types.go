package no_restricted_types

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// keywordTypeNames maps the tsgo keyword `ast.Kind` for a primitive type
// annotation to the lower-case name upstream uses as the bannedTypes map key
// (see TYPE_KEYWORDS in upstream). Listeners for these kinds are only wired up
// when the corresponding name is actually present in the user's `types`
// option, matching upstream's `objectReduceKey` behavior.
var keywordTypeNames = map[ast.Kind]string{
	ast.KindBigIntKeyword:    "bigint",
	ast.KindBooleanKeyword:   "boolean",
	ast.KindNeverKeyword:     "never",
	ast.KindNullKeyword:      "null",
	ast.KindNumberKeyword:    "number",
	ast.KindObjectKeyword:    "object",
	ast.KindStringKeyword:    "string",
	ast.KindSymbolKeyword:    "symbol",
	ast.KindUndefinedKeyword: "undefined",
	ast.KindUnknownKeyword:   "unknown",
	ast.KindVoidKeyword:      "void",
}

// bannedTypeConfig holds the parsed per-type configuration. The four shapes
// upstream supports (`true`, `false`, `null`, `string`, `object`) collapse
// here into: `Disabled` (null/false → don't report), or a record with the
// fields below. `kind` distinguishes the unset / string / object shapes so we
// can mirror upstream's getCustomMessage branching exactly.
type bannedTypeConfig struct {
	Disabled bool
	// "true" | "string" | "object"
	Kind    string
	Message string
	FixWith string
	HasFix  bool
	Suggest []string
}

func parseBannedTypes(raw map[string]interface{}) map[string]bannedTypeConfig {
	out := make(map[string]bannedTypeConfig, len(raw))
	for key, value := range raw {
		normalized := removeSpaces(key)
		cfg := bannedTypeConfig{}
		switch v := value.(type) {
		case nil:
			cfg.Disabled = true
		case bool:
			if !v {
				cfg.Disabled = true
			} else {
				cfg.Kind = "true"
			}
		case string:
			cfg.Kind = "string"
			cfg.Message = v
		case map[string]interface{}:
			cfg.Kind = "object"
			if msg, ok := v["message"].(string); ok {
				cfg.Message = msg
			}
			if fix, ok := v["fixWith"].(string); ok {
				cfg.FixWith = fix
				cfg.HasFix = true
			}
			if suggest, ok := v["suggest"].([]interface{}); ok {
				for _, s := range suggest {
					if str, ok := s.(string); ok {
						cfg.Suggest = append(cfg.Suggest, str)
					}
				}
			}
		default:
			// Unknown shape — upstream would pass through to `getCustomMessage`
			// and treat it as truthy; we treat it as a bare ban (no message).
			cfg.Kind = "true"
		}
		out[normalized] = cfg
	}
	return out
}

func removeSpaces(s string) string {
	// Match upstream's `replaceAll(/\s/g, '')` — strip every Unicode whitespace
	// codepoint, not just ASCII space. tsgo's source text is the raw input, so
	// tabs and newlines inside `Banned<\n  A\n>` must collapse the same way
	// the option key `Banned<A>` does.
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isWhitespace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\f', '\v', 0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	if r >= 0x2000 && r <= 0x200A {
		return true
	}
	return false
}

// customMessageFor mirrors upstream's getCustomMessage. Upstream short-circuits
// any falsy value (including the empty string ` Banned: '' `) at the
// `!bannedType` check and returns ''. So the rule for every shape collapses to:
// non-empty Message → " " + Message; otherwise "".
func customMessageFor(cfg bannedTypeConfig) string {
	if cfg.Message == "" {
		return ""
	}
	return " " + cfg.Message
}

var NoRestrictedTypesRule = rule.CreateRule(rule.Rule{
	Name: "no-restricted-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		bannedTypes := map[string]bannedTypeConfig{}
		if optsMap := utils.GetOptionsMap(options); optsMap != nil {
			if types, ok := optsMap["types"].(map[string]interface{}); ok {
				bannedTypes = parseBannedTypes(types)
			}
		}

		nodeText := func(node *ast.Node) string {
			r := utils.TrimNodeTextRange(ctx.SourceFile, node)
			return ctx.SourceFile.Text()[r.Pos():r.End()]
		}

		checkBannedTypes := func(typeNode *ast.Node, presetName string) {
			if typeNode == nil {
				return
			}
			name := presetName
			if name == "" {
				name = removeSpaces(nodeText(typeNode))
			}
			cfg, found := bannedTypes[name]
			if !found || cfg.Disabled {
				return
			}

			customMessage := customMessageFor(cfg)
			msg := rule.RuleMessage{
				Id:          "bannedTypeMessage",
				Description: fmt.Sprintf("Don't use `%s` as a type.%s", name, customMessage),
				Data: map[string]string{
					"name":          name,
					"customMessage": customMessage,
				},
			}

			var fixes []rule.RuleFix
			if cfg.Kind == "object" && cfg.HasFix {
				fixes = []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, typeNode, cfg.FixWith)}
			}

			var suggestions []rule.RuleSuggestion
			for _, replacement := range cfg.Suggest {
				suggestions = append(suggestions, rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          "bannedTypeReplacement",
						Description: fmt.Sprintf("Replace `%s` with `%s`.", name, replacement),
						Data: map[string]string{
							"name":        name,
							"replacement": replacement,
						},
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, typeNode, replacement)},
				})
			}

			switch {
			case len(fixes) > 0 && len(suggestions) > 0:
				ctx.ReportNodeWithFixesAndSuggestions(typeNode, msg, fixes, suggestions)
			case len(fixes) > 0:
				ctx.ReportNodeWithFixes(typeNode, msg, fixes...)
			case len(suggestions) > 0:
				ctx.ReportNodeWithSuggestions(typeNode, msg, suggestions...)
			default:
				ctx.ReportNode(typeNode, msg)
			}
		}

		listeners := rule.RuleListeners{}

		// Mirror upstream's `objectReduceKey(TYPE_KEYWORDS, ...)`: only
		// register a keyword listener when the corresponding name is
		// actually present in the user's banned-types map. This keeps the
		// traversal cost proportional to the configured set rather than
		// firing on every primitive annotation in the file.
		for kind, name := range keywordTypeNames {
			if _, banned := bannedTypes[name]; !banned {
				continue
			}
			kw := name
			kindCopy := kind
			listeners[kindCopy] = func(node *ast.Node) {
				checkBannedTypes(node, kw)
			}
		}

		listeners[ast.KindTypeReference] = func(node *ast.Node) {
			ref := node.AsTypeReferenceNode()
			if ref == nil {
				return
			}
			checkBannedTypes(ref.TypeName, "")
			if ref.TypeArguments != nil {
				checkBannedTypes(node, "")
			}
		}

		listeners[ast.KindExpressionWithTypeArguments] = func(node *ast.Node) {
			if !typescriptutil.IsClassImplementsOrInterfaceExtends(node) {
				return
			}
			expr := node.AsExpressionWithTypeArguments()
			if expr == nil {
				return
			}
			checkBannedTypes(expr.Expression, "")
			if expr.TypeArguments != nil {
				checkBannedTypes(node, "")
			}
		}

		listeners[ast.KindTupleType] = func(node *ast.Node) {
			tuple := node.AsTupleTypeNode()
			if tuple == nil {
				return
			}
			if tuple.Elements == nil || len(tuple.Elements.Nodes) == 0 {
				checkBannedTypes(node, "")
			}
		}

		listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
			literal := node.AsTypeLiteralNode()
			if literal == nil {
				return
			}
			if literal.Members == nil || len(literal.Members.Nodes) == 0 {
				checkBannedTypes(node, "")
			}
		}

		return listeners
	},
})
