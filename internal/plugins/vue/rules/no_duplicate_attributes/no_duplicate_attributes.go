// Package no_duplicate_attributes ports eslint-plugin-vue's
// `no-duplicate-attributes` rule.
package no_duplicate_attributes

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/vue/vast"
)

//go:embed no_duplicate_attributes.schema.json
var schemaJSON []byte

const messageID = "duplicateAttribute"

func duplicateMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: "Duplicate attribute '" + name + "'.",
		Data:        map[string]string{"name": name},
	}
}

// NoDuplicateAttributesRule disallows naming the same attribute twice on one
// element.
//
// Vue keeps the first of two identical attributes and silently drops the
// second, so a duplicate is always either dead code or a bug. The comparison
// spans both spellings: `foo` and `:foo` name the same thing, because
// `v-bind:foo` binds the attribute `foo`.
//
// `class` and `style` are the exception, and by default an allowed one. Vue
// merges a static `class` with a bound `:class` rather than dropping either,
// which is the common idiom for a fixed class plus conditional ones, so the two
// may coexist, but two static `class` attributes still cannot.
//
// https://github.com/vuejs/eslint-plugin-vue/blob/master/lib/rules/no-duplicate-attributes.js
var NoDuplicateAttributesRule = rule.Rule{
	Name:   "vue/no-duplicate-attributes",
	Schema: rule.NewSchema(schemaJSON),
	RunTemplate: func(ctx rule.RuleContext, options []any) rule.TemplateListeners {
		settings := parseOptions(options)

		return rule.TemplateListeners{
			vast.KindElement: func(element *vast.Node) {
				checkElement(ctx, element, settings)
			},
		}
	},
}

// options are the rule's two coexistence allowances.
type options struct {
	allowCoexistClass bool
	allowCoexistStyle bool
}

// parseOptions reads the rule's options. Both allowances default to true, which
// the schema fills in, so an absent option object still arrives here as one.
func parseOptions(raw []any) options {
	settings := options{allowCoexistClass: true, allowCoexistStyle: true}
	if len(raw) == 0 {
		return settings
	}
	object, ok := raw[0].(map[string]any)
	if !ok {
		return settings
	}
	if value, ok := object["allowCoexistClass"].(bool); ok {
		settings.allowCoexistClass = value
	}
	if value, ok := object["allowCoexistStyle"].(bool); ok {
		settings.allowCoexistStyle = value
	}
	return settings
}

// checkElement reports every attribute of one element whose name was already
// taken.
//
// Directives and plain attributes are counted separately so that a name allowed
// to coexist is a duplicate only within its own half, while every other name
// collides across both.
func checkElement(ctx rule.RuleContext, element *vast.Node, settings options) {
	directives := make(map[string]struct{})
	plain := make(map[string]struct{})

	for _, attribute := range element.Attributes() {
		payload := attribute.Attribute
		if payload == nil {
			continue
		}
		name, ok := attributeName(payload)
		if !ok {
			continue
		}

		seen := plain
		other := directives
		if payload.Directive {
			seen, other = directives, plain
		}

		duplicate := false
		if _, taken := seen[name]; taken {
			duplicate = true
		} else if !mayCoexist(name, settings) {
			_, duplicate = other[name]
		}

		seen[name] = struct{}{}
		if duplicate {
			ctx.ReportRange(attribute.Loc, duplicateMessage(name))
		}
	}
}

// attributeName is the name two attributes are compared by: a plain
// attribute's own name, or the argument a `v-bind` binds.
//
// Every other directive reports no name. `v-if` and `v-for` are not attributes
// of the rendered element, and a `v-bind` with no argument spreads an object
// whose keys are unknown until run time. Neither can be said to collide with
// anything.
func attributeName(attribute *vast.Attribute) (string, bool) {
	if !attribute.Directive {
		return attribute.Name, attribute.Name != ""
	}
	if attribute.Name != "bind" {
		return "", false
	}
	if attribute.Argument == "" || attribute.DynamicArgument {
		return "", false
	}
	return attribute.Argument, true
}

// mayCoexist reports whether a name is allowed to appear once as a plain
// attribute and once as a directive.
func mayCoexist(name string, settings options) bool {
	switch name {
	case "class":
		return settings.allowCoexistClass
	case "style":
		return settings.allowCoexistStyle
	}
	return false
}
