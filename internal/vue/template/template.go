// Package template parses the content of a Vue component's `<template>` block
// into a [vast] tree.
//
// # Scope
//
// This is a tree parser, not a conforming HTML parser. It covers what template
// rules ask about: nested elements, self-closing and void elements, the
// raw-text content model, comments, plain attributes, and Vue directives with
// their arguments and modifiers.
//
// It deliberately does not yet cover: parsing a directive's expression (the
// value is carried as source text), breaking `{{ }}` interpolation out of text,
// HTML entity decoding, implicit end tags for elements like `<p>` and `<li>`,
// or the SVG and MathML namespaces. Each of those is wanted by rules beyond the
// ones ported so far, and each will arrive with the first rule that needs it
// rather than speculatively.
//
// # Error recovery
//
// Markup a user is midway through editing must still produce a tree, because a
// rule that reports nothing on a broken file is more useful than one that
// crashes on it. An unterminated element closes at the end of the template, a
// stray end tag is ignored, and a `<` that begins no tag is text.
package template

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/vue/htmlscan"
	"github.com/web-infra-dev/rslint/internal/vue/vast"
)

// Parse parses the template content in text[content.Pos():content.End()] and
// returns the root node.
//
// text must be the whole component's text and content the template block's
// content range, so that every range in the returned tree is absolute in the
// component. That is the property that lets a template rule report and fix through the
// same offsets as a script rule.
func Parse(text string, content core.TextRange) *vast.Node {
	root := vast.NewRoot(content)
	parser := &parser{text: text, end: content.End()}
	parser.parseChildren(root, content.Pos(), "")
	return root
}

type parser struct {
	text string
	// end bounds every scan, so a template block can never consume the
	// `<style>` that follows it.
	end int
}

// parseChildren fills parent with the nodes between index and either the end of
// the template or the end tag for closeTag, and returns where it stopped.
//
// closeTag is empty for the root, which has no end tag to look for.
func (p *parser) parseChildren(parent *vast.Node, index int, closeTag string) int {
	textStart := index
	flushText := func(upTo int) {
		if upTo > textStart {
			parent.AppendChild(&vast.Node{
				Kind: vast.KindText,
				Loc:  core.NewTextRange(textStart, upTo),
			})
		}
	}

	for index < p.end {
		next := strings.IndexByte(p.text[index:p.end], '<')
		if next < 0 {
			break
		}
		index += next

		switch {
		case p.hasPrefix(index, "<!--"):
			flushText(index)
			commentEnd := min(htmlscan.CommentEnd(p.text, index), p.end)
			parent.AppendChild(&vast.Node{
				Kind: vast.KindComment,
				Loc:  core.NewTextRange(index, commentEnd),
			})
			index = commentEnd
			textStart = index

		case p.hasPrefix(index, "</"):
			nameEnd := htmlscan.TagNameEnd(p.text, index+2)
			name := htmlscan.Fold(p.text[index+2 : nameEnd])
			if closeTag != "" && name == closeTag {
				flushText(index)
				return index
			}
			// A stray end tag closes nothing and is not text either.
			flushText(index)
			index = min(htmlscan.DeclarationEnd(p.text, nameEnd), p.end)
			textStart = index

		case p.hasPrefix(index, "<!"), p.hasPrefix(index, "<?"):
			flushText(index)
			index = min(htmlscan.DeclarationEnd(p.text, index), p.end)
			textStart = index

		default:
			nameEnd := htmlscan.TagNameEnd(p.text, index+1)
			if nameEnd == index+1 {
				// A bare `<` in text, as in `a < b`.
				index++
				continue
			}
			flushText(index)
			element, after := p.parseElement(index, nameEnd)
			parent.AppendChild(element)
			index = after
			textStart = index
		}
	}

	flushText(p.end)
	return p.end
}

// parseElement parses one element whose `<` is at start and whose name ends at
// nameEnd, returning it and the offset just past it.
func (p *parser) parseElement(start, nameEnd int) (*vast.Node, int) {
	rawTag := p.text[start+1 : nameEnd]
	tag := htmlscan.Fold(rawTag)
	startTag := htmlscan.ReadStartTag(p.text, nameEnd)
	tagEnd := min(startTag.End, p.end)

	element := &vast.Node{
		Kind: vast.KindElement,
		Element: &vast.Element{
			Tag:         tag,
			RawTag:      rawTag,
			StartTag:    core.NewTextRange(start, tagEnd),
			SelfClosing: startTag.SelfClosing,
		},
	}
	for _, attribute := range startTag.Attributes {
		element.AppendAttribute(newAttribute(attribute))
	}

	if startTag.SelfClosing || htmlscan.IsVoid(tag) {
		element.Loc = core.NewTextRange(start, tagEnd)
		return element, tagEnd
	}

	if htmlscan.IsRawText(tag) {
		// Raw text ends at the first matching end tag, markup or not.
		contentEnd, elementEnd := p.findRawTextEnd(tag, tagEnd)
		if contentEnd > tagEnd {
			element.AppendChild(&vast.Node{
				Kind: vast.KindText,
				Loc:  core.NewTextRange(tagEnd, contentEnd),
			})
		}
		element.Element.EndTag = core.NewTextRange(contentEnd, elementEnd)
		element.Loc = core.NewTextRange(start, elementEnd)
		return element, elementEnd
	}

	contentEnd := p.parseChildren(element, tagEnd, tag)
	elementEnd := contentEnd
	if contentEnd < p.end && p.hasPrefix(contentEnd, "</") {
		nameEnd := htmlscan.TagNameEnd(p.text, contentEnd+2)
		elementEnd = min(htmlscan.DeclarationEnd(p.text, nameEnd), p.end)
		element.Element.EndTag = core.NewTextRange(contentEnd, elementEnd)
	}
	element.Loc = core.NewTextRange(start, elementEnd)
	return element, elementEnd
}

// findRawTextEnd locates the end tag closing a raw-text element, returning
// where its content ends and where the element ends.
func (p *parser) findRawTextEnd(tag string, contentStart int) (int, int) {
	for index := contentStart; index < p.end; {
		next := strings.IndexByte(p.text[index:p.end], '<')
		if next < 0 {
			break
		}
		index += next
		if p.hasPrefix(index, "</") {
			nameEnd := htmlscan.TagNameEnd(p.text, index+2)
			if htmlscan.Fold(p.text[index+2:nameEnd]) == tag {
				return index, min(htmlscan.DeclarationEnd(p.text, nameEnd), p.end)
			}
		}
		index++
	}
	return p.end, p.end
}

func (p *parser) hasPrefix(index int, prefix string) bool {
	if index+len(prefix) > p.end {
		return false
	}
	return p.text[index:index+len(prefix)] == prefix
}

// newAttribute turns one scanned start-tag entry into an attribute node,
// splitting a Vue directive into its name, argument and modifiers.
func newAttribute(scanned htmlscan.Attribute) *vast.Node {
	payload := &vast.Attribute{
		Key:      core.NewTextRange(scanned.NamePos, scanned.NameEnd),
		HasValue: scanned.HasValue,
		Value:    scanned.Value,
	}
	if scanned.HasValue {
		payload.ValueLoc = core.NewTextRange(scanned.ValuePos, scanned.ValueEnd)
	}

	parseDirectiveKey(scanned, payload)

	return &vast.Node{
		Kind:      vast.KindAttribute,
		Loc:       core.NewTextRange(scanned.NamePos, attributeEnd(scanned)),
		Attribute: payload,
	}
}

func attributeEnd(scanned htmlscan.Attribute) int {
	if scanned.HasValue {
		return scanned.ValueEnd
	}
	return scanned.NameEnd
}

// parseDirectiveKey fills in the directive half of payload from the written
// attribute name.
//
// The shapes are `v-name`, `v-name:arg`, `v-name:[dynamic]`, `:arg` and
// `.arg` for `v-bind`, `@arg` for `v-on`, and `#arg` for `v-slot`, each
// optionally followed by `.modifier` parts.
func parseDirectiveKey(scanned htmlscan.Attribute, payload *vast.Attribute) {
	name := scanned.Name
	offset := scanned.NamePos
	rest := ""

	switch {
	case strings.HasPrefix(name, "v-"):
		payload.Directive = true
		afterPrefix := name[len("v-"):]
		if cut := strings.IndexAny(afterPrefix, ":.["); cut >= 0 {
			payload.Name = htmlscan.Fold(afterPrefix[:cut])
			rest = afterPrefix[cut:]
			offset += len("v-") + cut
		} else {
			payload.Name = htmlscan.Fold(afterPrefix)
			offset += len(name)
		}
	case strings.HasPrefix(name, ":"), strings.HasPrefix(name, "."):
		// `.prop` is the `.prop` modifier shorthand of v-bind.
		payload.Directive = true
		payload.Name = "bind"
		rest = name
		if name[0] == '.' {
			payload.Modifiers = append(payload.Modifiers, "prop")
			rest = ":" + name[1:]
		}
	case strings.HasPrefix(name, "@"):
		payload.Directive = true
		payload.Name = "on"
		rest = ":" + name[len("@"):]
	case strings.HasPrefix(name, "#"):
		payload.Directive = true
		payload.Name = "slot"
		rest = ":" + name[len("#"):]
	default:
		// A plain attribute. HTML folds its name.
		payload.Name = htmlscan.Fold(name)
		return
	}

	if rest == "" {
		return
	}

	// Modifiers first, so an argument never swallows them.
	modifierPart := ""
	if cut := strings.IndexByte(rest, '.'); cut >= 0 && !strings.HasPrefix(rest, ".") {
		modifierPart = rest[cut+1:]
		rest = rest[:cut]
	} else if strings.HasPrefix(rest, ".") {
		modifierPart = rest[1:]
		rest = ""
	}
	for _, modifier := range strings.Split(modifierPart, ".") {
		if modifier != "" {
			payload.Modifiers = append(payload.Modifiers, modifier)
		}
	}

	if !strings.HasPrefix(rest, ":") {
		return
	}
	argument := rest[1:]
	argumentPos := offset + 1
	if strings.HasPrefix(argument, "[") {
		payload.DynamicArgument = true
		argument = strings.TrimSuffix(strings.TrimPrefix(argument, "["), "]")
		argumentPos++
	}
	if argument == "" {
		return
	}
	payload.Argument = argument
	payload.ArgumentLoc = core.NewTextRange(argumentPos, argumentPos+len(argument))
}
