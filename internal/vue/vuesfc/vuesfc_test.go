package vuesfc

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
)

// scriptText returns the projected text of every script block, which is what a
// parser reads out of the projection.
func scriptText(result Result) string {
	var parts []string
	for _, block := range result.Scripts() {
		parts = append(parts, result.Text[block.Content.Pos():block.Content.End()])
	}
	return strings.Join(parts, "")
}

// TestExtractPreservesOffsets locks in the one property everything else in
// .vue support is built on: the projection is the input's length, and script
// content sits at exactly the offsets it had in the file.
func TestExtractPreservesOffsets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "template before script",
			source: "<template>\n  <div>{{ msg }}</div>\n</template>\n\n<script>\nexport default { data: () => ({ msg: 'hi' }) };\n</script>\n",
		},
		{
			name:   "script before template",
			source: "<script setup lang=\"ts\">\nconst count = 0;\n</script>\n\n<template>\n  <p>{{ count }}</p>\n</template>\n",
		},
		{
			name:   "style and custom block",
			source: "<template><b/></template>\n<script>const a = 1;</script>\n<style scoped>.a { color: red }</style>\n<i18n>{ \"en\": {} }</i18n>\n",
		},
		{
			// Multi-byte bytes outside a script block become as many spaces,
			// which is what keeps every offset after them exact.
			name:   "multibyte template text",
			source: "<template>\n  <p>✓ ☂ 🎉 日本語</p>\n</template>\n<script>const a = 1;</script>\n",
		},
		{
			name:   "crlf line endings",
			source: "<template>\r\n  <div/>\r\n</template>\r\n<script>\r\nconst a = 1;\r\n</script>\r\n",
		},
		{
			name:   "no script block at all",
			source: "<template><div/></template>\n<style>.a{}</style>\n",
		},
		{
			name:   "empty source",
			source: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := Extract(test.source)

			if len(result.Text) != len(test.source) {
				t.Fatalf("projection length = %d, want %d", len(result.Text), len(test.source))
			}

			for _, block := range result.Scripts() {
				projected := result.Text[block.Content.Pos():block.Content.End()]
				original := test.source[block.Content.Pos():block.Content.End()]
				if projected != original {
					t.Errorf("script content at [%d,%d) = %q, want %q",
						block.Content.Pos(), block.Content.End(), projected, original)
				}
			}

			// Every byte outside a script block is a space or a line
			// terminator, so no template or style text can reach the parser.
			for index := range len(result.Text) {
				if inAnyScript(result, index) {
					continue
				}
				switch result.Text[index] {
				case ' ', '\n', '\r':
				default:
					t.Fatalf("byte %d outside a script block is %q, want blank",
						index, result.Text[index])
				}
			}
		})
	}
}

func inAnyScript(result Result, offset int) bool {
	for _, block := range result.Scripts() {
		if offset >= block.Content.Pos() && offset < block.Content.End() {
			return true
		}
	}
	return false
}

// TestExtractLineNumbersSurvive checks that the projection keeps every line
// break, so a diagnostic's reported line matches the .vue file.
func TestExtractLineNumbersSurvive(t *testing.T) {
	t.Parallel()

	source := "<template>\n  <div/>\n</template>\n<script>\nconst a = 1;\n</script>\n"
	result := Extract(source)

	if got, want := strings.Count(result.Text, "\n"), strings.Count(source, "\n"); got != want {
		t.Fatalf("projection has %d newlines, want %d", got, want)
	}

	offset := strings.Index(source, "const a")
	if got := strings.Index(result.Text, "const a"); got != offset {
		t.Fatalf("`const a` at offset %d in projection, want %d", got, offset)
	}
	if got, want := strings.Count(source[:offset], "\n"), 4; got != want {
		t.Fatalf("fixture drifted: `const a` on line index %d, want %d", got, want)
	}
}

func TestExtractScriptKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   core.ScriptKind
	}{
		{"no lang is javascript", "<script>const a = 1;</script>", core.ScriptKindJS},
		{"lang js", "<script lang=\"js\">const a = 1;</script>", core.ScriptKindJS},
		{"lang ts", "<script lang=\"ts\">const a: number = 1;</script>", core.ScriptKindTS},
		{"lang tsx", "<script lang=\"tsx\">const a = <div/>;</script>", core.ScriptKindTSX},
		{"lang jsx", "<script lang=\"jsx\">const a = <div/>;</script>", core.ScriptKindJSX},
		// Upstream compares `lang` exactly, so an uppercase value names no
		// language Vue compiles and the projection stays empty.
		{"lang uppercase is not typescript", "<script lang=\"TS\">const a: number = 1;</script>", core.ScriptKindJS},
		{"single quoted lang", "<script lang='ts'>const a: number = 1;</script>", core.ScriptKindTS},
		{"unquoted lang", "<script lang=ts>const a: number = 1;</script>", core.ScriptKindTS},
		{"no script is javascript", "<template><div/></template>", core.ScriptKindJS},
		{
			name:   "both blocks agree on ts",
			source: "<script lang=\"ts\">const a: number = 1;</script>\n<script setup lang=\"ts\">const b: number = 2;</script>",
			want:   core.ScriptKindTS,
		},
		{
			name:   "disagreeing blocks take the permissive kind",
			source: "<script lang=\"ts\">const a: number = 1;</script>\n<script setup lang=\"tsx\">const b = <div/>;</script>",
			want:   core.ScriptKindTSX,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := Extract(test.source).ScriptKind; got != test.want {
				t.Errorf("ScriptKind = %v, want %v", got, test.want)
			}
		})
	}
}

func TestExtractBlocks(t *testing.T) {
	t.Parallel()

	t.Run("both script blocks are found", func(t *testing.T) {
		t.Parallel()
		source := "<script>const a = 1;</script>\n<script setup>const b = 2;</script>"
		result := Extract(source)

		scripts := result.Scripts()
		if len(scripts) != 2 {
			t.Fatalf("got %d script blocks, want 2", len(scripts))
		}
		if scripts[0].Setup {
			t.Error("first block must not be a setup block")
		}
		if !scripts[1].Setup {
			t.Error("second block must be a setup block")
		}
		if got := scriptText(result); got != "const a = 1;const b = 2;" {
			t.Errorf("projected script text = %q", got)
		}
	})

	t.Run("block kinds are classified", func(t *testing.T) {
		t.Parallel()
		source := "<template><div/></template><script>x</script><style>.a{}</style><docs>hi</docs>"
		result := Extract(source)

		want := []BlockKind{BlockTemplate, BlockScript, BlockStyle, BlockOther}
		if len(result.Blocks) != len(want) {
			t.Fatalf("got %d blocks, want %d", len(result.Blocks), len(want))
		}
		for index, kind := range want {
			if result.Blocks[index].Kind != kind {
				t.Errorf("block %d kind = %v, want %v", index, result.Blocks[index].Kind, kind)
			}
		}
	})

	t.Run("uppercase tag name folds", func(t *testing.T) {
		t.Parallel()
		result := Extract("<SCRIPT>const a = 1;</SCRIPT>")
		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q, want the script content", got)
		}
	})

	t.Run("content range excludes the tags", func(t *testing.T) {
		t.Parallel()
		source := "<script>abc</script>"
		result := Extract(source)
		block := result.Blocks[0]
		if got := source[block.Content.Pos():block.Content.End()]; got != "abc" {
			t.Errorf("content = %q, want \"abc\"", got)
		}
		if got := source[block.Outer.Pos():block.Outer.End()]; got != source {
			t.Errorf("outer = %q, want the whole element", got)
		}
	})
}

// TestExtractHostileMarkup covers the shapes where a naive scanner silently
// takes the wrong span: the cases that would leak template text into the
// parser or truncate real code.
func TestExtractHostileMarkup(t *testing.T) {
	t.Parallel()

	t.Run("angle bracket in an attribute value does not end the tag", func(t *testing.T) {
		t.Parallel()
		result := Extract(`<script lang="ts" data-x=">">const a = 1;</script>`)
		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q", got)
		}
		if got := result.ScriptKind; got != core.ScriptKindTS {
			t.Errorf("ScriptKind = %v, want TS", got)
		}
	})

	t.Run("nested template does not close the outer one", func(t *testing.T) {
		t.Parallel()
		source := "<template>\n  <template v-if=\"a\"><b/></template>\n</template>\n<script>const a = 1;</script>"
		result := Extract(source)

		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q; the outer template swallowed the script", got)
		}
		if len(result.Blocks) != 2 {
			t.Fatalf("got %d top-level blocks, want 2", len(result.Blocks))
		}
	})

	t.Run("commented-out end tag does not close the block", func(t *testing.T) {
		t.Parallel()
		source := "<template>\n  <!-- </template> -->\n  <div/>\n</template>\n<script>const a = 1;</script>"
		result := Extract(source)

		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q", got)
		}
		if len(result.Blocks) != 2 {
			t.Fatalf("got %d top-level blocks, want 2", len(result.Blocks))
		}
	})

	t.Run("self-closing template with src has no content", func(t *testing.T) {
		t.Parallel()
		source := `<template src="./tpl.html"/><script>const a = 1;</script>`
		result := Extract(source)

		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q", got)
		}
	})

	t.Run("external script contributes no content", func(t *testing.T) {
		t.Parallel()
		result := Extract(`<script src="./main.ts" lang="ts"></script>`)

		if len(result.Blocks) != 1 || !result.Blocks[0].External {
			t.Fatalf("block not marked external: %+v", result.Blocks)
		}
		if got := scriptText(result); got != "" {
			t.Errorf("projected script text = %q, want empty", got)
		}
	})

	t.Run("uppercase lang contributes no content", func(t *testing.T) {
		t.Parallel()
		result := Extract(`<script lang="TS">const a: number = 1;</script>`)

		if got := scriptText(result); got != "" {
			t.Errorf("projected script text = %q, want empty because upstream reads `lang` exactly", got)
		}
	})

	t.Run("unsupported lang contributes no content", func(t *testing.T) {
		t.Parallel()
		source := `<script lang="coffee">a = 1 unless b</script>`
		result := Extract(source)

		if got := scriptText(result); got != "" {
			t.Errorf("projected script text = %q, want empty because coffee is not JavaScript", got)
		}
		if strings.TrimSpace(result.Text) != "" {
			t.Errorf("projection = %q, want all blank", result.Text)
		}
	})

	t.Run("script end tag inside a string closes the block", func(t *testing.T) {
		t.Parallel()
		// HTML's raw-text content model ends the block at the first `</script`,
		// string literal or not. Vue's own parser agrees, so the projection
		// must too rather than "fixing" the author's markup.
		source := `<script>const s = "</script>";</script>`
		result := Extract(source)

		if got := scriptText(result); got != `const s = "` {
			t.Errorf("projected script text = %q, want the text up to the first end tag", got)
		}
	})

	t.Run("unterminated script runs to end of file", func(t *testing.T) {
		t.Parallel()
		source := "<script>\nconst a = 1;\n"
		result := Extract(source)

		if got := scriptText(result); got != "\nconst a = 1;\n" {
			t.Errorf("projected script text = %q", got)
		}
		if len(result.Text) != len(source) {
			t.Errorf("projection length = %d, want %d", len(result.Text), len(source))
		}
	})

	t.Run("html comment before the blocks is recorded", func(t *testing.T) {
		t.Parallel()
		source := "<!-- eslint-disable no-console -->\n<script>console.log(1);</script>"
		result := Extract(source)

		if len(result.Comments) != 1 {
			t.Fatalf("got %d comments, want 1", len(result.Comments))
		}
		comment := result.Comments[0]
		if got := source[comment.Pos():comment.End()]; got != "<!-- eslint-disable no-console -->" {
			t.Errorf("comment = %q", got)
		}
	})

	t.Run("stray text and end tags are skipped", func(t *testing.T) {
		t.Parallel()
		source := "oops </div> a < b\n<script>const a = 1;</script>"
		result := Extract(source)

		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q", got)
		}
	})

	t.Run("style holding a script end tag does not confuse the scan", func(t *testing.T) {
		t.Parallel()
		source := "<style>.a::before { content: \"</script>\" }</style>\n<script>const a = 1;</script>"
		result := Extract(source)

		if got := scriptText(result); got != "const a = 1;" {
			t.Errorf("projected script text = %q", got)
		}
	})
}
