package estree

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func source(path, text string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, text, core.GetScriptKindFromFileName(path))
}

type nativeSyntaxState struct {
	kind       ast.Kind
	flags      ast.NodeFlags
	parent     *ast.Node
	start, end int
	children   []*ast.Node
}

func snapshotNativeSyntax(file *ast.SourceFile) map[*ast.Node]nativeSyntaxState {
	states := make(map[*ast.Node]nativeSyntaxState)
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		state := nativeSyntaxState{kind: node.Kind, flags: node.Flags, parent: node.Parent, start: node.Pos(), end: node.End()}
		node.ForEachChild(func(child *ast.Node) bool {
			state.children = append(state.children, child)
			return visit(child)
		})
		states[node] = state
		return false
	}
	visit(file.AsNode())
	return states
}

func TestSchemaMatchesNodeDecoder(t *testing.T) {
	data, err := os.ReadFile("../../../../packages/rslint/src/languages/js/frame-schema.ts")
	if err != nil {
		t.Fatal(err)
	}
	start := bytes.Index(data, []byte("= [")) + 2
	if start < 2 {
		t.Fatal("missing Node schema table")
	}
	end := bytes.Index(data[start:], []byte(";")) + start
	text := strings.ReplaceAll(string(data[start:end]), "'", "\"")
	text = regexp.MustCompile(`,\s*]`).ReplaceAllString(text, "]")
	var actual [][]json.RawMessage
	if err := json.Unmarshal([]byte(text), &actual); err != nil {
		t.Fatal(err)
	}
	if len(actual) != len(schemas) {
		t.Fatalf("schema count %d != %d", len(actual), len(schemas))
	}
	fingerprint := fnv.New32a()
	for index, schema := range schemas {
		var name string
		var fields []string
		if err := json.Unmarshal(actual[index][0], &name); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(actual[index][1], &fields); err != nil {
			t.Fatal(err)
		}
		if name != schema.name || !reflect.DeepEqual(fields, schema.fields) {
			t.Fatalf("schema %d drifted: %s %v", index, name, fields)
		}
		for _, word := range append([]string{name}, fields...) {
			fingerprint.Write([]byte(word))
			fingerprint.Write([]byte{0})
		}
	}
	if fingerprint.Sum32() != schemaHash {
		t.Fatal("schema fingerprint drifted")
	}
	match := regexp.MustCompile(`FRAME_SCHEMA_HASH = (0x[0-9a-f]+)`).FindSubmatch(data)
	if len(match) != 2 {
		t.Fatal("missing Node schema fingerprint")
	}
	nodeHash, err := strconv.ParseUint(string(match[1]), 0, 32)
	if err != nil || uint32(nodeHash) != schemaHash {
		t.Fatalf("Node schema fingerprint drifted: %s", match[1])
	}
}

func TestSupportedSyntaxAndFrameOwnership(t *testing.T) {
	t.Run("token covers preserve source order and repeated positions", func(t *testing.T) {
		file := source("/covers.ts", "first; second;")
		covers := []tokenCover{
			{start: 7, end: 13, kind: ast.KindIdentifier, typeCode: tokenJSXIdentifier},
			{start: 0, end: 5, kind: ast.KindIdentifier, typeCode: tokenIdentifier},
			{start: 7, end: 13, kind: ast.KindIdentifier, typeCode: tokenKeyword},
		}
		tokens, _ := scanTokensAndCommentsInto(context.Background(), scanner.NewScanner(), file, utf16Offsets{file.GetPositionMap()}, covers, nil, nil, maxFrameSize)
		if len(tokens) != 4 || tokens[0].typeCode != tokenIdentifier || tokens[2].typeCode != tokenKeyword {
			t.Fatalf("token cover precedence changed: %v", tokens)
		}
	})
	cases := map[string]string{
		"/basic.js":     "'use strict'; const [, value = fallback, ...rest] = tuple; ({ value = 1, ...other } = input); (a?.b).c; a?.b.c?.();",
		"/functions.js": "async function* f(...items) { yield* items; } const o = { get x() { return 1 }, set x(v) {}, async f(v) { return v } };",
		"/templates.js": "const x = `a${ `b${c}` }d`; tag`\\unicode`; const r = /a[\\/]/giu; const n = 0x20000000000001; const b = 0x10n; const s = '\\uD800';",
		"/control.js":   "outer: for (let i = 0; i < 2; i++) { if (i) continue outer; } try { throw 1 } catch ({a}) { while (a) break; } finally {} switch(x) { case 1: break; default: break; }",
		"/class.ts":     "@deco abstract class C<T> extends Base<T> implements I<T> { abstract f(x?: T): T; accessor x = 1; readonly y!: T; constructor(public a: T) {} static { init() } }",
		"/types.ts":     "export interface I<T extends string = string> extends Base<T> { readonly x?: T; f<U>(x: U): T; [key: string]: T; new(x: T): I<T> } type A<T> = { -readonly [K in keyof T as K]-?: T[K] }; type B = (x: number) => string; type C = [a: string, ...rest: number[]]; type D<T> = T extends infer U ? U : never;",
		"/modules.ts":   "import type { Foo as F } from 'a'; export type { F }; export * as ns from 'a'; export as namespace N; import X = require('x'); export = X; namespace A.B { export const x = 1 } enum E { A, B = 2 }",
		"/jsx.tsx":      "const x = <A.B ns:x='&amp;amp;' {...p}>x\r\ny&amp;{/* hi */}{value}<><input /></></A.B>;",
		"/jsdoc.js":     "/** @template T\n * @this {Object}\n * @param {string} x\n * @returns {number} */ function f(x) { return 1 } /** @type {number} */ (value); /** @satisfies {string} */ (other);",
		"/unicode.tsx":  "/* 世界😀 */ const café = <A 字段='😃'>{`世界${value}`}</A>; const lone = '\\uD800';",
	}
	var builder Builder
	for path, text := range cases {
		t.Run(path, func(t *testing.T) {
			file := source(path, text)
			original := snapshotNativeSyntax(file)
			data, err := builder.Build(context.Background(), file, Options{SourceType: "module"})
			if err != nil {
				t.Fatal(err)
			}
			if file.Text() != text || !reflect.DeepEqual(original, snapshotNativeSyntax(file)) {
				t.Fatal("projection changed native source syntax")
			}
			saved := bytes.Clone(data)
			if builder.p.file != nil || builder.p.text != "" || builder.p.sourceType != "" || builder.p.offsets.positions != nil || len(builder.p.stringIDs) != 0 {
				t.Fatal("builder retained source graph")
			}
			assertClearedProjectionStrings(t, &builder)
			builder.Reset()
			if len(builder.p.stringIDs) != 0 || len(builder.p.tokenCovers) != 0 || builder.p.scan.Text() != "" {
				t.Fatal("reset builder retained source indexes")
			}
			next, err := builder.Build(context.Background(), source("/next.js", "next()"), Options{SourceType: "script"})
			if err != nil {
				t.Fatal(err)
			}
			next[len(next)-1] ^= 1
			if !bytes.Equal(data, saved) {
				t.Fatal("builder output aliases reusable scratch")
			}
			if binary.LittleEndian.Uint32(data[16:]) != uint32(len(data)) || string(data[binary.LittleEndian.Uint32(data[60:]):]) != text {
				t.Fatal("frame source or size changed")
			}
		})
	}
}

func TestProjectionCancellationAndFailure(t *testing.T) {
	var builder Builder
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if data, err := builder.Build(ctx, source("/file.ts", "const x = 1"), Options{SourceType: "module"}); data != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled build = %v/%v", data, err)
	}
	if data, err := builder.Build(context.Background(), nil, Options{SourceType: "module"}); data != nil || err == nil {
		t.Fatal("nil syntax accepted")
	}
	for _, delta := range []int{-1, 1} {
		calls := 0
		data, err := builder.BuildInto(context.Background(), source("/file.ts", "const x = 1"), Options{SourceType: "module"}, func(size int) []byte {
			calls++
			return make([]byte, size+delta)
		})
		if data != nil || err == nil || calls != 1 {
			t.Fatalf("incorrect allocator size %d = %v/%v, calls=%d", delta, data, err, calls)
		}
		assertClearedProjectionStrings(t, &builder)
	}
	for _, mode := range []string{"panic", "cancel"} {
		t.Run("allocator "+mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			data, err := builder.BuildInto(ctx, source("/file.ts", "const text = 'staged string';"), Options{SourceType: "module"}, func(size int) []byte {
				if mode == "panic" {
					panic("output allocator failure")
				}
				cancel()
				return make([]byte, size)
			})
			if data != nil || err == nil || (mode == "cancel" && !errors.Is(err, context.Canceled)) {
				t.Fatalf("allocator %s = %v/%v", mode, data, err)
			}
			assertClearedProjectionStrings(t, &builder)
		})
	}
	t.Run("cancel while encoding strings", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		probe := &cancelOnProjectionCheck{Context: ctx, cancel: cancel, remaining: 1 << 30}
		text := "const text = '" + strings.Repeat(`\u4e16\u754c\uD83D\uDE00`, 8192) + "';"
		data, err := builder.BuildInto(probe, source("/file.ts", text), Options{SourceType: "module"}, func(size int) []byte {
			probe.remaining = 8
			return make([]byte, size)
		})
		if data != nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("string encoding cancellation = %v/%v", data, err)
		}
		assertClearedProjectionStrings(t, &builder)
	})
	file := source("/file.ts", "const x = 1")
	file.Statements.Nodes[0].Kind = ast.KindUnknown
	if data, err := builder.Build(context.Background(), file, Options{SourceType: "module"}); data != nil || err == nil {
		t.Fatal("unsupported syntax became a successful tree")
	}
	if data, err := builder.Build(context.Background(), source("/file.ts", "const x = 1"), Options{SourceType: "commonjs", HasBOM: true}); err != nil || binary.LittleEndian.Uint32(data[12:]) != 1 {
		t.Fatal("builder cannot recover after a failure")
	}
	for _, after := range []int{2, 50, 700} {
		ctx, cancel := context.WithCancel(context.Background())
		ctx = &cancelOnProjectionCheck{Context: ctx, cancel: cancel, remaining: after}
		file := source("/file.ts", strings.Repeat("const value = { key: 'text' };\n", 100))
		data, err := builder.Build(ctx, file, Options{SourceType: "module"})
		cancel()
		if data != nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("mid-projection cancellation at %d = %v/%v", after, data, err)
		}
		if builder.p.file != nil || builder.p.ctx != nil || builder.p.text != "" || builder.p.offsets.positions != nil || len(builder.p.stringIDs) != 0 || len(builder.p.tokenCovers) != 0 {
			t.Fatal("canceled projection retained generation references")
		}
		assertClearedProjectionStrings(t, &builder)
	}
}

func TestProjectionStringEncoding(t *testing.T) {
	cases := []struct {
		text  string
		units []uint16
	}{
		{"", nil},
		{"abc", []uint16{'a', 'b', 'c'}},
		{"世界", []uint16{0x4e16, 0x754c}},
		{"😀", []uint16{0xd83d, 0xde00}},
		{"\xed\xa0\x80", []uint16{0xd800}},
		{"\xed\xb0\x80", []uint16{0xdc00}},
		{"\xed\xa0\x80\xed\xb0\x80", []uint16{0xd800, 0xdc00}},
		{"\xff", []uint16{0xfffd}},
		{strings.Repeat("a", 4095) + "😀\xed\xa0\x80世界", append(slices.Repeat([]uint16{'a'}, 4095), 0xd83d, 0xde00, 0xd800, 0x4e16, 0x754c)},
	}
	for _, test := range cases {
		t.Run(fmt.Sprintf("%x", test.text[:min(len(test.text), 32)]), func(t *testing.T) {
			p := estreeProjector{ctx: context.Background(), stringIDs: make(map[string]uint32)}
			value := p.stringValue(test.text)
			data := bytes.Repeat([]byte{0xff}, p.stringBytes)
			p.encodeStrings(data)
			if value != taggedValue(tagString, 0) || int(binary.LittleEndian.Uint32(data)) != len(test.units) {
				t.Fatal("encoded string length or tag changed")
			}
			for index, expected := range test.units {
				if actual := binary.LittleEndian.Uint16(data[4+index*2:]); actual != expected {
					t.Fatalf("unit %d: got %x, want %x", index, actual, expected)
				}
			}
			if p.stringBytes%4 != 0 || p.stringValue(test.text) != value || len(p.strings) != 1 {
				t.Fatal("string alignment or deduplication changed")
			}
			for _, padding := range data[4+len(test.units)*2:] {
				if padding != 0 {
					t.Fatal("string alignment retained previous output bytes")
				}
			}
			if next := p.stringValue("next:" + test.text); next != taggedValue(tagString, 1) || p.stringValue(test.text) != value {
				t.Fatal("string references must be sequential IDs independent of record byte size")
			}
		})
	}
}

func TestProjectionSourceStringEncoding(t *testing.T) {
	text := strings.Repeat("prefix", 3) + "世界😀source text" + strings.Repeat("suffix", 3)
	value := text[18 : len(text)-18]
	for _, aliased := range []bool{true, false} {
		t.Run(strconv.FormatBool(aliased), func(t *testing.T) {
			input := value
			if !aliased {
				input = strings.Clone(value)
			}
			p := estreeProjector{ctx: context.Background(), text: text, stringIDs: make(map[string]uint32)}
			id := p.stringValue(input)
			data := make([]byte, p.stringBytes)
			p.encodeStrings(data)
			encoded := binary.LittleEndian.Uint32(data)
			if aliased {
				if len(data) != 8 || encoded != ^uint32(len(value)) || binary.LittleEndian.Uint32(data[4:]) != 18 {
					t.Fatalf("source span changed: %x", data)
				}
			} else if encoded>>31 != 0 || encoded != uint32(ecmascript.StringCodeUnitCount(value)) {
				t.Fatal("independent values must retain UTF-16 encoding")
			}
			if id != p.stringValue(strings.Clone(value)) || len(p.strings) != 1 {
				t.Fatal("representation changed string deduplication")
			}
			p.clearSource()
			if p.text != "" || len(p.strings) != 0 || len(p.stringIDs) != 0 {
				t.Fatal("source spans retained source references")
			}
		})
	}
	p := estreeProjector{text: text}
	if _, ok := p.sourceStringOffset(text[19:22]); ok {
		t.Fatal("partial UTF-8 scalars cannot use source spans")
	}
}

func TestProjectionScratchResetDropsLargeBuffers(t *testing.T) {
	t.Run("evicts large columns without discarding small ones", func(t *testing.T) {
		builder := Builder{p: estreeProjector{
			nodes:  make([]nodeRecord, 0, 1<<19),
			values: make([]uint32, 0, 128),
			lists:  make([]uint32, 0, 64),
		}}
		builder.Reset()
		if cap(builder.p.nodes) != 0 || cap(builder.p.values) != 128 || cap(builder.p.lists) != 64 {
			t.Fatal("one oversized column discarded unrelated small scratch")
		}
	})
	t.Run("clears scanner state and bounds dictionary reuse", func(t *testing.T) {
		var builder Builder
		if _, err := builder.Build(context.Background(), source("/small.ts", "// @ts-ignore\nconst value = 0xabcdef;"), Options{SourceType: "module"}); err != nil {
			t.Fatal(err)
		}
		index := builder.p.stringIDs
		builder.Reset()
		if builder.p.scan.Text() != "" || builder.p.scan.TokenValue() != "" || len(builder.p.scan.CommentDirectives()) != 0 || len(index) != 0 {
			t.Fatal("scratch retained scanner or dictionary source references")
		}
		builder.p.stringIDs["reuse probe"] = 1
		if _, ok := index["reuse probe"]; !ok {
			t.Fatal("small cleared dictionary was not reused")
		}
		delete(builder.p.stringIDs, "reuse probe")
		var text strings.Builder
		for index := range 4100 {
			fmt.Fprintf(&text, "name%d;", index)
		}
		if _, err := builder.Build(context.Background(), source("/dictionary.ts", text.String()), Options{SourceType: "module"}); err != nil {
			t.Fatal(err)
		}
		builder.Reset()
		if builder.p.stringIDs != nil || len(index) != 0 {
			t.Fatal("oversized dictionary survived reset")
		}
	})
	text := strings.Repeat("const value = { key: 'text' };\n", 20_000)
	var builder Builder
	data, err := builder.Build(context.Background(), source("/large.ts", text), Options{SourceType: "module"})
	if err != nil {
		t.Fatal(err)
	}
	saved := bytes.Clone(data)
	builder.Reset()
	retained := uintptr(0)
	for _, buffer := range []any{builder.p.nodes, builder.p.values, builder.p.lists, builder.p.strings, builder.p.numbers, builder.tokens, builder.comments, builder.p.tokenCovers} {
		value := reflect.ValueOf(buffer)
		retained += uintptr(value.Cap()) * value.Type().Elem().Size()
	}
	if retained > 4<<20 {
		t.Fatalf("oversized scratch survived reset: %d bytes", retained)
	}
	assertClearedProjectionStrings(t, &builder)
	if _, err := builder.Build(context.Background(), source("/next.ts", "next()"), Options{SourceType: "script"}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, saved) {
		t.Fatal("reset changed an independently owned frame")
	}
}

func assertClearedProjectionStrings(t *testing.T, builder *Builder) {
	t.Helper()
	for _, value := range builder.p.strings[:cap(builder.p.strings)] {
		if value != "" {
			t.Fatal("projection scratch retained a staged source string")
		}
	}
}

func TestProjectionCallerOwnedOutput(t *testing.T) {
	var builder Builder
	buffer := bytes.Repeat([]byte{0xff}, 4096)
	for index, text := range []string{"const café = <A x='😀' />;", "next()", "const value = '\\uD800';"} {
		file := source("/file.tsx", text)
		options := Options{SourceType: "module", HasBOM: index == 0}
		want, err := builder.Build(context.Background(), file, options)
		if err != nil {
			t.Fatal(err)
		}
		got, err := builder.BuildInto(context.Background(), file, options, func(size int) []byte {
			return buffer[:size]
		})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) || &got[0] != &buffer[0] || cap(got) != len(got) {
			t.Fatal("caller buffer was not reused with exact frame bytes")
		}
		builder.Reset()
		if !bytes.Equal(got, want) {
			t.Fatal("reset changed caller-owned output")
		}
	}
}

func BenchmarkProjectionWindow(b *testing.B) {
	cases := map[string]string{
		"ASCII":      strings.Repeat("function f(a) { return { value: a + 2, text: 'abc' }; }\n", 1000),
		"Unicode":    strings.Repeat("const café = { 字段: '\\uD800', text: '😃' };\n", 1000),
		"LongString": "const text = '" + strings.Repeat("a", 1<<20) + "';",
	}
	for name, text := range cases {
		file := source("/input.ts", text)
		for _, window := range []int{1, 32} {
			b.Run(fmt.Sprintf("%s/%d", name, window), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text) * window))
				for b.Loop() {
					var builder Builder
					for range window {
						if _, err := builder.Build(context.Background(), file, Options{SourceType: "module"}); err != nil {
							b.Fatal(err)
						}
					}
				}
			})
		}
	}
}

// Cancel at a deterministic producer checkpoint, without wall-clock races.
type cancelOnProjectionCheck struct {
	context.Context
	cancel    context.CancelFunc
	remaining int
}

func (ctx *cancelOnProjectionCheck) Err() error {
	ctx.remaining--
	if ctx.remaining == 0 {
		ctx.cancel()
	}
	return ctx.Context.Err()
}
