// Package media_has_caption ports eslint-plugin-jsx-a11y's
// `media-has-caption` rule. The rule enforces that `<audio>` and `<video>`
// elements include a `<track kind="captions">` child for accessibility,
// unless the media element is muted (`muted={true}` or boolean form).
//
// Custom wrappers can be registered via the `audio`, `video`, and `track`
// option arrays. Like the other jsx-a11y rules, the polymorphic prop /
// components-map settings are honored via `jsxa11yutil.GetElementType`.
package media_has_caption

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed media_has_caption.schema.json
var schemaJSON []byte

const errorMessage = "Media elements such as <audio> and <video> must have a <track> for captions."

// mediaTypes mirrors upstream's `MEDIA_TYPES` constant — the always-on base
// set for the `isMediaType` check, regardless of `options.audio` /
// `options.video`.
var mediaTypes = []string{"audio", "video"}

type options struct {
	// audio / video / track mirror upstream's `options.{audio,video,track}`.
	// Each is an array of additional component names that should be treated
	// as the corresponding native element. Compared AS-IS against the
	// resolved nodeType (after `getElementType`'s polymorphic / components-map
	// resolution).
	audio []string
	video []string
	track []string
}

func parseOptions(raw []any) options {
	opts := options{}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]interface{})
	opts.audio = jsxa11yutil.StringSliceOption(m["audio"])
	opts.video = jsxa11yutil.StringSliceOption(m["video"])
	opts.track = jsxa11yutil.StringSliceOption(m["track"])
	return opts
}

var MediaHasCaptionRule = rule.Rule{
	Name:   "jsx-a11y/media-has-caption",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Mirrors upstream's `isMediaType`:
		//   MEDIA_TYPES.concat(flatMap(MEDIA_TYPES, m => options[m]))
		//              .some(t => t === type)
		// flatMap concatenates BOTH options.audio AND options.video into one
		// pool, then strict-equality checks against the resolved type.
		isMediaType := func(t string) bool {
			if t == "" {
				return false
			}
			if slices.Contains(mediaTypes, t) {
				return true
			}
			return slices.Contains(opts.audio, t) || slices.Contains(opts.video, t)
		}

		// Mirrors upstream's `isTrackType`:
		//   ['track'].concat(options.track || []).some(t => t === type)
		isTrackType := func(t string) bool {
			if t == "track" {
				return true
			}
			return slices.Contains(opts.track, t)
		}

		check := func(node *ast.Node) {
			elementType := jsxa11yutil.GetElementType(node, ctx.Settings)
			if !isMediaType(elementType) {
				return
			}

			attrs := reactutil.GetJsxElementAttributes(node)
			mutedProp := jsxa11yutil.FindAttributeByName(attrs, "muted")
			// Upstream's `mutedPropVal === true` is the strict literal check —
			// only the actual boolean `true` (or coerced `"true"` string)
			// silences the rule. `null`, `false`, identifiers, conditionals
			// all fall through to the report path.
			if jsxa11yutil.LiteralPropIsExactlyTrue(mutedProp) {
				return
			}

			// Children only exist for the paired form. tsgo splits ESTree's
			// `JSXElement` into KindJsxElement (paired, owns Children) and
			// KindJsxSelfClosingElement (no children). Upstream's
			// `node.children.filter(...)` returns [] for the self-closing
			// shape since ESTree's JSXElement of a self-closing element has
			// `children: []`.
			var children []*ast.Node
			if node.Kind == ast.KindJsxOpeningElement && node.Parent != nil &&
				node.Parent.Kind == ast.KindJsxElement {
				children = reactutil.GetJsxChildren(node.Parent)
			}

			// Filter for track-type element children. Upstream filters by
			// `child.type === 'JSXElement'`, which in ESTree covers BOTH
			// paired and self-closing JSX child elements (both share the
			// JSXElement type). tsgo splits them, so we accept both
			// KindJsxElement and KindJsxSelfClosingElement.
			//
			// JsxFragment children are skipped to match upstream — JSXFragment
			// is a distinct ESTree type and fails the `=== 'JSXElement'` filter.
			var trackChildren []*ast.Node
			for _, child := range children {
				var trackOpening *ast.Node
				switch child.Kind {
				case ast.KindJsxElement:
					trackOpening = child.AsJsxElement().OpeningElement
				case ast.KindJsxSelfClosingElement:
					trackOpening = child
				default:
					continue
				}
				if trackOpening == nil {
					continue
				}
				if isTrackType(jsxa11yutil.GetElementType(trackOpening, ctx.Settings)) {
					trackChildren = append(trackChildren, trackOpening)
				}
			}

			if len(trackChildren) == 0 {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "mediaHasCaption",
					Description: errorMessage,
				})
				return
			}

			// Upstream's caption check:
			//   const kindPropValue = getLiteralPropValue(kindProp) || '';
			//   return kindPropValue.toLowerCase() === 'captions';
			// `LiteralPropStringValue` returns ("", false) for non-string
			// literal extractions (Identifier, missing prop, boolean form,
			// "true"/"false" coerced to bool). The `|| ''` fallback then
			// turns it into "" — which case-insensitive-fails the captions
			// match. Net effect: only an actual literal string "captions"
			// (case-insensitive) clears the rule.
			hasCaption := false
			for _, trackOpening := range trackChildren {
				trackAttrs := reactutil.GetJsxElementAttributes(trackOpening)
				kindProp := jsxa11yutil.FindAttributeByName(trackAttrs, "kind")
				if kind, ok := jsxa11yutil.LiteralPropStringValue(kindProp); ok {
					if strings.EqualFold(kind, "captions") {
						hasCaption = true
						break
					}
				}
			}

			if !hasCaption {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "mediaHasCaption",
					Description: errorMessage,
				})
			}
		}

		// tsgo splits ESTree's `JSXOpeningElement` into KindJsxOpeningElement
		// (paired tags `<audio>...</audio>`) and KindJsxSelfClosingElement
		// (`<audio />`). Upstream's single `JSXElement` listener fires once
		// per element regardless of form; we mirror by registering on both
		// opening kinds. The JsxOpeningElement case walks up to its parent
		// JsxElement to read children.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
