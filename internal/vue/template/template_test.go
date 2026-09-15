package template

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/vue/vast"
	"github.com/web-infra-dev/rslint/internal/vue/vuesfc"
)

// parseComponent parses the template block of a whole component, which is how
// every caller uses this package.
func parseComponent(t *testing.T, component string) (*vast.Node, string) {
	t.Helper()
	sfc := vuesfc.Extract(component)
	for _, block := range sfc.Blocks {
		if block.Kind == vuesfc.BlockTemplate {
			return Parse(component, block.Content), component
		}
	}
	t.Fatalf("component has no template block:\n%s", component)
	return nil, ""
}

// elements collects every element in the tree, depth first.
func elements(root *vast.Node) []*vast.Node {
	var found []*vast.Node
	vast.Walk(root, func(node *vast.Node) bool {
		if node.Kind == vast.KindElement {
			found = append(found, node)
		}
		return true
	})
	return found
}

func attributeByKey(t *testing.T, element *vast.Node, source string, key string) *vast.Attribute {
	t.Helper()
	for _, attribute := range element.Attributes() {
		if source[attribute.Attribute.Key.Pos():attribute.Attribute.Key.End()] == key {
			return attribute.Attribute
		}
	}
	t.Fatalf("element <%s> has no attribute written %q", element.Tag(), key)
	return nil
}

// TestParseRangesAreAbsoluteInTheComponent locks in the property the whole
// design rests on: a range in the template tree indexes the component's text,
// so a template rule reports and fixes through the same offsets as a script
// rule.
func TestParseRangesAreAbsoluteInTheComponent(t *testing.T) {
	t.Parallel()

	component := "<script>\nconst a = 1;\n</script>\n\n<template>\n  <div class=\"box\">hi</div>\n</template>\n"
	root, source := parseComponent(t, component)

	found := elements(root)
	if len(found) != 1 {
		t.Fatalf("got %d elements, want 1", len(found))
	}
	div := found[0]

	if got := source[div.Loc.Pos():div.Loc.End()]; got != `<div class="box">hi</div>` {
		t.Errorf("element range covers %q", got)
	}
	if got := source[div.Element.StartTag.Pos():div.Element.StartTag.End()]; got != `<div class="box">` {
		t.Errorf("start tag range covers %q", got)
	}
	if got := source[div.Element.EndTag.Pos():div.Element.EndTag.End()]; got != `</div>` {
		t.Errorf("end tag range covers %q", got)
	}

	class := attributeByKey(t, div, source, "class")
	if got := source[class.ValueLoc.Pos():class.ValueLoc.End()]; got != "box" {
		t.Errorf("attribute value range covers %q", got)
	}

	// The template starts well past the script, so a tree built from
	// template-relative offsets would land inside the script block here.
	if div.Loc.Pos() < len("<script>\nconst a = 1;\n</script>") {
		t.Errorf("element starts at %d, which is inside the script block", div.Loc.Pos())
	}
}

func TestParseTree(t *testing.T) {
	t.Parallel()

	t.Run("nesting and text", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t,
			"<template>\n  <ul>\n    <li>one</li>\n    <li>two</li>\n  </ul>\n</template>\n")

		found := elements(root)
		if len(found) != 3 {
			t.Fatalf("got %d elements, want 3", len(found))
		}
		if found[0].Tag() != "ul" || found[1].Tag() != "li" || found[2].Tag() != "li" {
			t.Fatalf("tags = %q, %q, %q", found[0].Tag(), found[1].Tag(), found[2].Tag())
		}
		if found[1].Parent != found[0] {
			t.Error("the first <li> is not a child of the <ul>")
		}
	})

	t.Run("nested template does not close the outer one", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t,
			"<template>\n  <template v-if=\"a\"><b/></template>\n  <i/>\n</template>\n")

		found := elements(root)
		if len(found) != 3 {
			t.Fatalf("got %d elements, want 3: %v", len(found), tagsOf(found))
		}
		if got := tagsOf(found); got[0] != "template" || got[1] != "b" || got[2] != "i" {
			t.Errorf("tags = %v", got)
		}
		// The <i/> is a sibling of the inner template, not its child.
		if found[2].Parent != root {
			t.Error("<i/> did not land at the template root")
		}
	})

	t.Run("self-closing and void elements have no children", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t, "<template><img src=\"a.png\"><br><hr/>after</template>")

		found := elements(root)
		if got := tagsOf(found); len(got) != 3 || got[0] != "img" || got[1] != "br" || got[2] != "hr" {
			t.Fatalf("tags = %v", got)
		}
		for _, element := range found {
			if len(element.Children) != 0 {
				t.Errorf("<%s> swallowed %d children", element.Tag(), len(element.Children))
			}
		}
	})

	t.Run("comment is a node, not text", func(t *testing.T) {
		t.Parallel()
		root, source := parseComponent(t, "<template><!-- eslint-disable -->  <div/></template>")

		var comments []*vast.Node
		vast.Walk(root, func(node *vast.Node) bool {
			if node.Kind == vast.KindComment {
				comments = append(comments, node)
			}
			return true
		})
		if len(comments) != 1 {
			t.Fatalf("got %d comments, want 1", len(comments))
		}
		if got := source[comments[0].Loc.Pos():comments[0].Loc.End()]; got != "<!-- eslint-disable -->" {
			t.Errorf("comment range covers %q", got)
		}
	})

	t.Run("a commented-out end tag does not close an element", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t, "<template><div><!-- </div> --><b/></div></template>")

		found := elements(root)
		if got := tagsOf(found); len(got) != 2 || got[0] != "div" || got[1] != "b" {
			t.Fatalf("tags = %v", got)
		}
		if found[1].Parent != found[0] {
			t.Error("<b/> escaped the <div> through a commented-out end tag")
		}
	})

	t.Run("the template never reads past its own block", func(t *testing.T) {
		t.Parallel()
		root, source := parseComponent(t,
			"<template><div/></template>\n<style>.a { color: red }</style>\n")

		found := elements(root)
		if got := tagsOf(found); len(got) != 1 || got[0] != "div" {
			t.Fatalf("tags = %v: the parser reached into the style block", got)
		}
		if root.Loc.End() > len(source)-len("\n<style>.a { color: red }</style>\n") {
			t.Errorf("root ends at %d, past the template block", root.Loc.End())
		}
	})

	t.Run("unterminated element closes at the end of the template", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t, "<template><div><span>text</template>")

		found := elements(root)
		if got := tagsOf(found); len(got) != 2 || got[0] != "div" || got[1] != "span" {
			t.Fatalf("tags = %v", got)
		}
	})

	t.Run("a bare angle bracket is text", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t, "<template><p>a < b</p></template>")

		found := elements(root)
		if got := tagsOf(found); len(got) != 1 || got[0] != "p" {
			t.Fatalf("tags = %v", got)
		}
	})

	t.Run("tag name folds but the raw name is kept", func(t *testing.T) {
		t.Parallel()
		root, _ := parseComponent(t, "<template><MyThing/></template>")

		found := elements(root)
		if len(found) != 1 {
			t.Fatalf("got %d elements, want 1", len(found))
		}
		if found[0].Tag() != "mything" {
			t.Errorf("folded tag = %q, want \"mything\"", found[0].Tag())
		}
		if found[0].Element.RawTag != "MyThing" {
			t.Errorf("raw tag = %q, want \"MyThing\"", found[0].Element.RawTag)
		}
	})
}

func TestParseDirectives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		written        string
		wantDirective  bool
		wantName       string
		wantArgument   string
		wantModifiers  []string
		wantDynamicArg bool
	}{
		{name: "plain attribute", written: "class", wantName: "class"},
		{name: "plain attribute folds", written: "CLASS", wantName: "class"},
		{name: "bare directive", written: "v-if", wantDirective: true, wantName: "if"},
		{
			name: "directive with argument", written: "v-bind:value",
			wantDirective: true, wantName: "bind", wantArgument: "value",
		},
		{
			name: "colon shorthand", written: ":value",
			wantDirective: true, wantName: "bind", wantArgument: "value",
		},
		{
			name: "argument case is kept", written: ":fooBar",
			wantDirective: true, wantName: "bind", wantArgument: "fooBar",
		},
		{
			name: "at shorthand is v-on", written: "@click",
			wantDirective: true, wantName: "on", wantArgument: "click",
		},
		{
			name: "hash shorthand is v-slot", written: "#header",
			wantDirective: true, wantName: "slot", wantArgument: "header",
		},
		{
			name: "modifiers", written: "@click.stop.prevent",
			wantDirective: true, wantName: "on", wantArgument: "click",
			wantModifiers: []string{"stop", "prevent"},
		},
		{
			name: "bare directive with modifier", written: "v-model.trim",
			wantDirective: true, wantName: "model", wantModifiers: []string{"trim"},
		},
		{
			name: "dot shorthand is a v-bind prop", written: ".value",
			wantDirective: true, wantName: "bind", wantArgument: "value",
			wantModifiers: []string{"prop"},
		},
		{
			name: "dynamic argument", written: "v-bind:[name]",
			wantDirective: true, wantName: "bind", wantArgument: "name",
			wantDynamicArg: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			root, source := parseComponent(t,
				"<template><div "+test.written+"=\"x\"></div></template>")

			found := elements(root)
			if len(found) != 1 {
				t.Fatalf("got %d elements, want 1", len(found))
			}
			attribute := attributeByKey(t, found[0], source, test.written)

			if attribute.Directive != test.wantDirective {
				t.Errorf("Directive = %v, want %v", attribute.Directive, test.wantDirective)
			}
			if attribute.Name != test.wantName {
				t.Errorf("Name = %q, want %q", attribute.Name, test.wantName)
			}
			if attribute.Argument != test.wantArgument {
				t.Errorf("Argument = %q, want %q", attribute.Argument, test.wantArgument)
			}
			if attribute.DynamicArgument != test.wantDynamicArg {
				t.Errorf("DynamicArgument = %v, want %v", attribute.DynamicArgument, test.wantDynamicArg)
			}
			if len(attribute.Modifiers) != len(test.wantModifiers) {
				t.Fatalf("Modifiers = %v, want %v", attribute.Modifiers, test.wantModifiers)
			}
			for index, modifier := range test.wantModifiers {
				if attribute.Modifiers[index] != modifier {
					t.Errorf("Modifiers[%d] = %q, want %q", index, attribute.Modifiers[index], modifier)
				}
			}
			if attribute.Argument != "" && !test.wantDynamicArg {
				if got := source[attribute.ArgumentLoc.Pos():attribute.ArgumentLoc.End()]; got != test.wantArgument {
					t.Errorf("argument range covers %q, want %q", got, test.wantArgument)
				}
			}
		})
	}
}

func TestParseAttributeValues(t *testing.T) {
	t.Parallel()

	t.Run("a valueless attribute is distinguishable", func(t *testing.T) {
		t.Parallel()
		root, source := parseComponent(t, `<template><input disabled required=""></template>`)

		element := elements(root)[0]
		disabled := attributeByKey(t, element, source, "disabled")
		if disabled.HasValue {
			t.Error("`disabled` reported a value")
		}
		required := attributeByKey(t, element, source, "required")
		if !required.HasValue || required.Value != "" {
			t.Errorf("`required=\"\"` = %+v, want an empty value", required)
		}
	})

	t.Run("an angle bracket in a value does not end the tag", func(t *testing.T) {
		t.Parallel()
		root, source := parseComponent(t, `<template><div v-if="a > b"><b/></div></template>`)

		found := elements(root)
		if got := tagsOf(found); len(got) != 2 || got[0] != "div" || got[1] != "b" {
			t.Fatalf("tags = %v", got)
		}
		condition := attributeByKey(t, found[0], source, "v-if")
		if condition.Value != "a > b" {
			t.Errorf("value = %q, want \"a > b\"", condition.Value)
		}
	})

	t.Run("single quoted value", func(t *testing.T) {
		t.Parallel()
		root, source := parseComponent(t, `<template><div title='x y'/></template>`)
		title := attributeByKey(t, elements(root)[0], source, "title")
		if title.Value != "x y" {
			t.Errorf("value = %q", title.Value)
		}
	})

	t.Run("unquoted value", func(t *testing.T) {
		t.Parallel()
		root, source := parseComponent(t, `<template><div id=main></div></template>`)
		id := attributeByKey(t, elements(root)[0], source, "id")
		if id.Value != "main" {
			t.Errorf("value = %q", id.Value)
		}
	})

	t.Run("a slash belongs to an unquoted value", func(t *testing.T) {
		t.Parallel()
		// HTML's unquoted attribute value state ends only at whitespace or
		// `>`, so the `/` in `id=main/` is part of the value and the element
		// is not self-closing. Vue's own tokenizer reads it the same way, and
		// this is why `<div id=main/>` surprises people.
		root, source := parseComponent(t, `<template><div id=main/></template>`)
		element := elements(root)[0]
		id := attributeByKey(t, element, source, "id")
		if id.Value != "main/" {
			t.Errorf("value = %q, want \"main/\"", id.Value)
		}
		if element.Element.SelfClosing {
			t.Error("the element reported itself self-closing")
		}
	})
}

func tagsOf(nodes []*vast.Node) []string {
	tags := make([]string, len(nodes))
	for index, node := range nodes {
		tags[index] = node.Tag()
	}
	return tags
}
