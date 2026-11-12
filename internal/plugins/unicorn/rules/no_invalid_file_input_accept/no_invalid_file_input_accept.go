// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package no_invalid_file_input_accept

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/transformers/jsxtransforms"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	invalidMessageID   = "no-invalid-file-input-accept/invalid"
	staticMessageID    = "no-invalid-file-input-accept/static"
	normalizeMessageID = "no-invalid-file-input-accept/normalize"
)

var (
	escapeDoubleQuoted = strings.NewReplacer("&", "&amp;", `"`, "&quot;", "\x00", "&#0;")
	escapeSingleQuoted = strings.NewReplacer("&", "&amp;", "'", "&apos;", "\x00", "&#0;")
)

var NoInvalidFileInputAcceptRule = rule.Rule{
	Name:   "unicorn/no-invalid-file-input-accept",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		staticValue := func(initializer *ast.Node) (string, bool) {
			if initializer == nil {
				return "", false
			}
			switch initializer.Kind {
			case ast.KindStringLiteral:
				literal := initializer.AsStringLiteral()
				if literal.TokenFlags&ast.TokenFlagsUnterminated == 0 {
					return jsxtransforms.DecodeEntities(literal.Text), true
				}
			case ast.KindJsxExpression:
				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}
				value, _ := evaluator.EvalValueIfNoSideEffects(initializer.Expression())
				text, ok := value.(string)
				return text, ok
			}
			return "", false
		}

		// Native rules receive JS/TS ASTs. Upstream's html-eslint Tag listener
		// has no counterpart for standalone HTML here.
		checkInput := func(node *ast.Node) {
			tag := node.TagName()
			if !ast.IsIdentifier(tag) || tag.Text() != "input" {
				return
			}
			attributes := node.Attributes().Properties()
			acceptAttribute := firstJSXAttribute(attributes, "accept")
			typeAttribute := firstJSXAttribute(attributes, "type")
			if acceptAttribute == nil || typeAttribute == nil {
				return
			}
			inputType, known := staticValue(typeAttribute.Initializer())
			if !known || !ecmascript.EqualsWhenLowercased(inputType, "file") {
				return
			}

			initializer := acceptAttribute.Initializer()
			if initializer == nil {
				reportInvalid(ctx, acceptAttribute, "The `accept` attribute must have a value.")
				return
			}
			value, known := staticValue(initializer)
			if !known {
				ctx.ReportNode(initializer, rule.RuleMessage{
					Id: staticMessageID, Description: "The file input `accept` value must be static.",
				})
				return
			}
			checkAccept(ctx, initializer, value)
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     checkInput,
			ast.KindJsxSelfClosingElement: checkInput,
		}
	},
}

func firstJSXAttribute(attributes []*ast.Node, name string) *ast.Node {
	for _, attribute := range attributes {
		if ast.IsJsxAttribute(attribute) && ast.IsIdentifier(attribute.Name()) && attribute.Name().Text() == name {
			return attribute
		}
	}
	return nil
}

func reportInvalid(ctx rule.RuleContext, node *ast.Node, reason string) {
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          invalidMessageID,
		Description: "Invalid file input `accept` value: " + reason,
		Data:        map[string]string{"reason": reason},
	})
}

func checkAccept(ctx rule.RuleContext, node *ast.Node, value string) {
	replacement, reason := normalizeAccept(value)
	if reason != "" {
		reportInvalid(ctx, node, reason)
		return
	}
	if replacement == value {
		return
	}

	message := rule.RuleMessage{
		Id:          normalizeMessageID,
		Description: "Prefer `" + replacement + "` as the file input `accept` value.",
		Data:        map[string]string{"replacement": replacement},
	}
	if !ast.IsStringLiteral(node) {
		ctx.ReportNode(node, message)
		return
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	ctx.ReportRangeWithDeferredFixes(textRange, message, func() []rule.RuleFix {
		raw := node.Text()
		// Different JSX readers can assign different values to numeric references.
		if !utf8.ValidString(replacement) ||
			strings.Contains(raw, "&") && ecmascript.DecodeJSXAttribute(raw) != value {
			return nil
		}
		// Keep the JSX delimiter and escape characters that would change the
		// parsed value or insert a NUL byte into the source.
		escape := escapeDoubleQuoted
		if node.AsStringLiteral().TokenFlags&ast.TokenFlagsSingleQuote != 0 {
			escape = escapeSingleQuoted
		}
		content := textRange.WithPos(textRange.Pos() + 1).WithEnd(textRange.End() - 1)
		return []rule.RuleFix{rule.RuleFixReplaceRange(content, escape.Replace(replacement))}
	})
}
