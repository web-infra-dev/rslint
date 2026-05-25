package no_unescaped_entities

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type entity struct {
	// char is the original entity character for use in error messages.
	char string
	// charRune is the decoded single rune when hasRune is true. Both fields
	// are needed because the rune NUL (U+0000) is a legitimate match target
	// and cannot be distinguished from "unset" by value alone.
	charRune rune
	hasRune  bool
	// alternatives is the list of suggested replacements. An empty list
	// means the entity is reported without suggestions (the "string" form
	// of the forbid option).
	alternatives []string
}

// defaultEntities mirrors the ESLint rule's DEFAULTS. `<` and `{` are
// intentionally omitted because they cause syntax errors when left
// unescaped in JSX.
var defaultEntities = []entity{
	mustSingleRune(">", []string{"&gt;"}),
	mustSingleRune("\"", []string{"&quot;", "&ldquo;", "&#34;", "&rdquo;"}),
	mustSingleRune("'", []string{"&apos;", "&lsquo;", "&#39;", "&rsquo;"}),
	mustSingleRune("}", []string{"&#125;"}),
}

func mustSingleRune(char string, alts []string) entity {
	r, size := utf8.DecodeRuneInString(char)
	if size == 0 || size != len(char) || r == utf8.RuneError {
		panic("default entity must be a single valid rune: " + char)
	}
	return entity{char: char, charRune: r, hasRune: true, alternatives: alts}
}

// newEntity builds an entity from a user-configured forbid item.
// Multi-rune or empty strings are accepted for parity with ESLint (which
// never matches them on a per-char scan).
func newEntity(char string, alts []string) entity {
	if char == "" {
		return entity{char: char, alternatives: alts}
	}
	r, size := utf8.DecodeRuneInString(char)
	if size != len(char) || r == utf8.RuneError {
		return entity{char: char, alternatives: alts}
	}
	return entity{char: char, charRune: r, hasRune: true, alternatives: alts}
}

// parseEntities extracts the `forbid` option. When the option is missing
// (or malformed) defaults are used; when it is explicitly provided — even
// as an empty array — the caller's list is respected verbatim, matching
// ESLint's `configuration.forbid || DEFAULTS` semantics.
func parseEntities(options any) []entity {
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return defaultEntities
	}
	rawForbid, ok := optsMap["forbid"]
	if !ok {
		return defaultEntities
	}
	forbidArr, ok := rawForbid.([]interface{})
	if !ok {
		return defaultEntities
	}
	result := make([]entity, 0, len(forbidArr))
	for _, item := range forbidArr {
		switch v := item.(type) {
		case string:
			result = append(result, newEntity(v, nil))
		case map[string]interface{}:
			char, _ := v["char"].(string)
			var alts []string
			if altsRaw, ok := v["alternatives"].([]interface{}); ok {
				alts = make([]string, 0, len(altsRaw))
				for _, alt := range altsRaw {
					if s, ok := alt.(string); ok {
						alts = append(alts, s)
					}
				}
			}
			result = append(result, newEntity(char, alts))
		}
	}
	return result
}

var NoUnescapedEntitiesRule = rule.Rule{
	Name: "react/no-unescaped-entities",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		entities := parseEntities(options)
		if len(entities) == 0 {
			return rule.RuleListeners{}
		}

		return rule.RuleListeners{
			// JsxText is the only AST node that represents literal text appearing
			// as a direct child of a JsxElement/JsxFragment. String literals,
			// attribute values, and expression-container contents live on other
			// node kinds and are intentionally not checked (matches ESLint's
			// `'Literal, JSXText'` selector + `isJSX(parent)` predicate, since
			// TypeScript never produces a Literal whose parent is JSX).
			//
			// NOTE: TypeScript's JSX parser rejects unescaped `>` and `}` in JSX
			// text with TS1381/TS1382 before this rule runs, so those defaults
			// only meaningfully fire on characters the parser still accepts
			// (`'`, `"`, and any custom `forbid` chars).
			ast.KindJsxText: func(node *ast.Node) {
				// Intentionally do not skip whitespace-only JsxText: a user may
				// configure `forbid` to include a whitespace character, which
				// ESLint would still flag.
				source := ctx.SourceFile.Text()
				startPos, endPos := node.Pos(), node.End()
				if startPos < 0 || endPos > len(source) || startPos >= endPos {
					return
				}
				content := source[startPos:endPos]

				// Iterate rune-by-rune so multi-byte Unicode forbid chars work.
				// `i` is the byte offset within `content`; the absolute source
				// position is `startPos + i`.
				for i := 0; i < len(content); {
					r, size := utf8.DecodeRuneInString(content[i:])
					runeStart := startPos + i
					runeEnd := runeStart + size
					for idx := range entities {
						e := &entities[idx]
						if !e.hasRune || e.charRune != r {
							continue
						}
						reportEntity(ctx, e, runeStart, runeEnd)
					}
					i += size
				}
			},
		}
	},
}

func reportEntity(ctx rule.RuleContext, e *entity, start, end int) {
	charRange := core.NewTextRange(start, end)
	if len(e.alternatives) == 0 {
		ctx.ReportRange(charRange, rule.RuleMessage{
			Id:          "unescapedEntity",
			Description: "HTML entity, `" + e.char + "` , must be escaped.",
		})
		return
	}

	suggestions := make([]rule.RuleSuggestion, len(e.alternatives))
	altsQuoted := make([]string, len(e.alternatives))
	for i, alt := range e.alternatives {
		altsQuoted[i] = "`" + alt + "`"
		suggestions[i] = rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          "replaceWithAlt",
				Description: "Replace with `" + alt + "`.",
			},
			FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(charRange, alt)},
		}
	}
	ctx.ReportRangeWithSuggestions(charRange, rule.RuleMessage{
		Id:          "unescapedEntityAlts",
		Description: "`" + e.char + "` can be escaped with " + strings.Join(altsQuoted, ", ") + ".",
	}, suggestions...)
}
