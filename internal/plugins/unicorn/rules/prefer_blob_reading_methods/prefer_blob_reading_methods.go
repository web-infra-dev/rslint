// Package prefer_blob_reading_methods ports eslint-plugin-unicorn's
// `prefer-blob-reading-methods` rule.
package prefer_blob_reading_methods

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "error"

var fileReaderMethods = []string{"readAsText", "readAsArrayBuffer"}

func preferBlobMethodMessage(method string) rule.RuleMessage {
	replacement := "text"
	if method == "readAsArrayBuffer" {
		replacement = "arrayBuffer"
	}

	return rule.RuleMessage{
		Id:          messageID,
		Description: fmt.Sprintf("Prefer `Blob#%s()` over `FileReader#%s(blob)`.", replacement, method),
		Data: map[string]string{
			"method":      method,
			"replacement": replacement,
		},
	}
}

// PreferBlobReadingMethodsRule prefers Blob's promise-based reading methods
// over the callback-based FileReader methods with equivalent results.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/prefer-blob-reading-methods.js
var PreferBlobReadingMethodsRule = rule.Rule{
	Name:   "unicorn/prefer-blob-reading-methods",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		oneArgument := 1

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods:             fileReaderMethods,
					ArgumentsLength:     &oneArgument,
					RejectSpreadElement: true,
				})
				if !ok {
					return
				}

				method := call.Property.AsIdentifier().Text
				ctx.ReportNode(call.Property, preferBlobMethodMessage(method))
			},
		}
	},
}
