package ban_types

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type BanTypesOptions struct {
	Types          map[string]interface{} `json:"types"`
	ExtendDefaults bool                   `json:"extendDefaults"`
}

type BannedTypeConfig struct {
	Message string      `json:"message"`
	FixWith interface{} `json:"fixWith"`
}

var defaultBannedTypes = map[string]BannedTypeConfig{
	"String": {
		Message: "Use string instead",
		FixWith: "string",
	},
	"Boolean": {
		Message: "Use boolean instead",
		FixWith: "boolean",
	},
	"Number": {
		Message: "Use number instead",
		FixWith: "number",
	},
	"Symbol": {
		Message: "Use symbol instead",
		FixWith: "symbol",
	},
	"BigInt": {
		Message: "Use bigint instead",
		FixWith: "bigint",
	},
	"Function": {
		Message: "The `Function` type accepts any function-like value. It provides no type safety when calling the function, which can be a common source of bugs. It also accepts things like class declarations, which will throw at runtime as they will not be called with `new`. If you are expecting the function to accept certain arguments, you should explicitly define the function shape.",
		FixWith: nil,
	},
	"Object": {
		Message: "The `Object` type actually means \"any non-nullish value\", so it is marginally better than `unknown`. If you want a type meaning \"any object\", you probably want `object` instead. If you want a type meaning \"any value\", you probably want `unknown` instead.",
		FixWith: nil,
	},
	"{}": {
		Message: "`{}` actually means \"any non-nullish value\". If you want a type meaning \"any object\", you probably want `object` instead. If you want a type meaning \"empty object\", you probably want `Record<string, never>` instead.",
		FixWith: "object",
	},
}

var BanTypesRule = rule.CreateRule(rule.Rule{
	Name: "ban-types",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := BanTypesOptions{
			Types:          make(map[string]interface{}),
			ExtendDefaults: true,
		}

		// Parse options with dual-format support
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if types, ok := optsMap["types"].(map[string]interface{}); ok {
					opts.Types = types
				}
				if extendDefaults, ok := optsMap["extendDefaults"].(bool); ok {
					opts.ExtendDefaults = extendDefaults
				}
			}
		}

		// Build banned types map
		bannedTypes := make(map[string]BannedTypeConfig)
		if opts.ExtendDefaults {
			for k, v := range defaultBannedTypes {
				bannedTypes[k] = v
			}
		}

		// Merge custom types
		for typeName, typeConfig := range opts.Types {
			if typeConfig == nil || typeConfig == false {
				// Explicit null or false disables the ban
				delete(bannedTypes, typeName)
				continue
			}

			config := BannedTypeConfig{}
			switch v := typeConfig.(type) {
			case string:
				config.Message = v
			case map[string]interface{}:
				if msg, ok := v["message"].(string); ok {
					config.Message = msg
				}
				if fix, ok := v["fixWith"]; ok {
					config.FixWith = fix
				}
			}
			bannedTypes[typeName] = config
		}

		checkTypeReference := func(node *ast.Node) {
			typeRef := node.AsTypeReferenceNode()
			if typeRef == nil || typeRef.TypeName == nil {
				return
			}

			// Get the type name text
			typeName := getTypeNameText(ctx, typeRef.TypeName)
			if typeName == "" {
				return
			}

			// Check if this type is banned
			config, isBanned := bannedTypes[typeName]
			if !isBanned {
				return
			}

			message := config.Message
			if message == "" {
				message = fmt.Sprintf("Don't use `%s` as a type.", typeName)
			}

			// Create fix if available
			if config.FixWith != nil {
				if fixStr, ok := config.FixWith.(string); ok {
					ctx.ReportNodeWithFixes(node, rule.RuleMessage{
						Id:          "bannedTypeMessage",
						Description: message,
					}, rule.RuleFixReplace(ctx.SourceFile, node, fixStr))
					return
				}
			}

			// Report without fix
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "bannedTypeMessage",
				Description: message,
			})
		}

		return rule.RuleListeners{
			ast.KindTypeReference: checkTypeReference,
		}
	},
})

func getTypeNameText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		ident := node.AsIdentifier()
		if ident != nil {
			return ident.Text
		}
	case ast.KindQualifiedName:
		// For qualified names like A.B.C, we only care about the root identifier
		qual := node.AsQualifiedName()
		if qual != nil && qual.Left != nil {
			return getTypeNameText(ctx, qual.Left)
		}
	}

	// Fallback: get text from source
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
	return strings.TrimSpace(text)
}
