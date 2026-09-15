// Package vast is the abstract syntax tree of a Vue template.
//
// # Why a second tree
//
// rslint's rule engine is built on one AST, TypeScript's, and every one of its
// rules is keyed on an ast.Kind. A Vue template is not TypeScript and cannot be
// represented in that tree, so a rule about `<template>` needs a tree of its
// own.
//
// The shape here follows TypeScript's rather than ESTree's (a node carries a
// Kind, a Parent, a range and its children), even though upstream
// eslint-plugin-vue rules are written against vue-eslint-parser's ESTree-shaped
// nodes. Matching the surrounding engine is worth more than matching upstream's
// field names: the traversal, the disable manager, and the whole ReportRange
// family already speak this shape, and every rule this repository has ever
// ported already translates from ESTree names to a different tree.
//
// # Ranges
//
// Every range is a byte offset into the *whole component's* text, not into the
// template block. That is possible because the projection the TypeScript parser
// reads preserves every offset (see internal/vue/vuesfc), so one set of offsets
// addresses the script, the template and the file on disk alike. A template
// rule therefore reports through ctx.ReportRange like any other, and its
// autofixes need no translation.
//
// Note that utils.TrimNodeTextRange must not be used on these nodes: it asks
// TypeScript's scanner to skip trivia, and a template has no such notion. A
// node's Loc is already exact.
package vast

import "github.com/microsoft/typescript-go/shim/core"

// Kind identifies what a node is.
type Kind uint8

const (
	// KindRoot is the template block's content. It is the node a traversal
	// starts from and has no tag of its own.
	KindRoot Kind = iota
	// KindElement is an element, whether an HTML element, a Vue built-in like
	// `<template>`, or a component.
	KindElement
	// KindText is character data between tags. Interpolation is not broken out
	// of it yet.
	KindText
	// KindComment is an HTML comment.
	KindComment
	// KindAttribute is one entry of a start tag: either a plain attribute or a
	// Vue directive.
	KindAttribute
)

// Node is one node of a template.
//
// The kind-specific payloads hang off dedicated pointers rather than an
// interface, so a rule reads them without a type assertion and a node of
// another kind cannot be misread: exactly one of them is non-nil, and it is the
// one Kind names.
type Node struct {
	Kind Kind
	// Loc is the node's exact range, absolute in the component's text.
	Loc    core.TextRange
	Parent *Node
	// Children are element and root children in source order. An attribute is
	// not a child; it hangs off Element.Attributes.
	Children []*Node

	// Element is non-nil exactly when Kind is KindElement.
	Element *Element
	// Attribute is non-nil exactly when Kind is KindAttribute.
	Attribute *Attribute
}

// Element is the payload of an element node.
type Element struct {
	// Tag is the tag name folded to ASCII lower case, which is how HTML
	// compares two tag names. Use it to ask what an element is.
	Tag string
	// RawTag is the tag name as written. A component referenced as `<MyThing>`
	// keeps its capitals here, which is what a rule about naming needs.
	RawTag string
	// StartTag spans `<tag ...>` and EndTag spans `</tag>`. EndTag is the empty
	// range for a self-closing or void element, and for one the author never
	// closed.
	StartTag core.TextRange
	EndTag   core.TextRange
	// Attributes are the start tag's entries in source order.
	Attributes  []*Node
	SelfClosing bool
}

// Attribute is the payload of an attribute node, covering both a plain
// attribute and a Vue directive.
//
// A directive is written `v-name:argument.modifier="value"`, with `:arg` and
// `@arg` as the shorthands for `v-bind:arg` and `v-on:arg`. The parts are kept
// separate because that is how rules ask about them: `no-duplicate-attributes`
// compares a plain attribute's Name against a `v-bind` directive's Argument,
// and neither reading works if the two are one string.
type Attribute struct {
	// Directive reports whether this entry is a Vue directive rather than a
	// plain attribute.
	Directive bool
	// Name is the plain attribute's name, or the directive's name without its
	// `v-` prefix, such as `bind`, `on`, `if`, `for` or `slot`. A shorthand resolves to
	// the directive it stands for.
	Name string
	// Key spans the whole `v-name:argument.modifier` or plain name.
	Key core.TextRange
	// Argument is a directive's argument, empty when it has none. It is not
	// folded: `:fooBar` binds a prop whose case matters.
	Argument string
	// ArgumentLoc is the argument's range, empty when there is no argument.
	ArgumentLoc core.TextRange
	// DynamicArgument reports a bracketed argument, `v-bind:[name]`, whose
	// value is only known at run time. Argument then holds the expression
	// source rather than a static name.
	DynamicArgument bool
	// Modifiers are the `.`-separated modifiers, without their dots.
	Modifiers []string
	// HasValue distinguishes `disabled` from `disabled=""`.
	HasValue bool
	// Value is the attribute value with quotes removed and no entity decoding.
	// For a directive this is the expression source, which is not parsed yet.
	Value string
	// ValueLoc is the value's range excluding quotes, empty when there is no
	// value.
	ValueLoc core.TextRange
}

// NewRoot returns an empty root node spanning loc.
func NewRoot(loc core.TextRange) *Node {
	return &Node{Kind: KindRoot, Loc: loc}
}

// AppendChild adds child to parent and records the parent link.
func (n *Node) AppendChild(child *Node) {
	if n == nil || child == nil {
		return
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}

// AppendAttribute adds an attribute node to an element and records the parent
// link. It does nothing for a node that is not an element.
func (n *Node) AppendAttribute(attribute *Node) {
	if n == nil || attribute == nil || n.Element == nil {
		return
	}
	attribute.Parent = n
	n.Element.Attributes = append(n.Element.Attributes, attribute)
}

// Tag returns an element node's folded tag name, or the empty string for any
// other node. It saves a rule the nil check on every question about a tag.
func (n *Node) Tag() string {
	if n == nil || n.Element == nil {
		return ""
	}
	return n.Element.Tag
}

// Attributes returns an element node's attributes, or nil for any other node.
func (n *Node) Attributes() []*Node {
	if n == nil || n.Element == nil {
		return nil
	}
	return n.Element.Attributes
}

// HasDirective reports whether an element carries the named directive, given
// without its `v-` prefix.
func (n *Node) HasDirective(name string) bool {
	for _, attribute := range n.Attributes() {
		if payload := attribute.Attribute; payload != nil &&
			payload.Directive && payload.Name == name {
			return true
		}
	}
	return false
}

// Walk visits n and then its children, depth first, in source order. Returning
// false from visit skips that node's children.
//
// Attributes are visited after their element and before its children, so a
// listener sees a start tag complete before anything inside it.
func Walk(node *Node, visit func(*Node) bool) {
	if node == nil || !visit(node) {
		return
	}
	for _, attribute := range node.Attributes() {
		visit(attribute)
	}
	for _, child := range node.Children {
		Walk(child, visit)
	}
}
