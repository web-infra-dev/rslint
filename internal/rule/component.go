package rule

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/vue/vuesfc"
)

// Component lazily provides the Vue single file component structure behind one
// linted file.
//
// A component's parsed text is a projection of its <script> blocks with every
// other byte blanked, so ctx.SourceFile.Text() cannot say anything about the
// component itself: which blocks it has, and even their tags, are gone from it
// by the time a rule runs. The component's own text is read back from the file
// system instead, the same way [SourceBOM] reads a file's bytes for the one
// question its text cannot answer.
//
// One store per file is shared by every rule on it, and the read happens on the
// first question asked. A file that is not a component, and a component no rule
// asks about, both pay nothing.
type Component struct {
	fs   vfs.FS
	path string
	once sync.Once
	sfc  vuesfc.Result
	ok   bool
}

// NewComponent returns the store for one file. A nil file system, or a path
// that is not a component, yields a store that reports it is not one.
func NewComponent(fileSystem vfs.FS, path string) *Component {
	return &Component{fs: fileSystem, path: path}
}

func (c *Component) load() {
	if c == nil {
		return
	}
	c.once.Do(func() {
		if c.fs == nil || !vuesfc.IsFile(c.path) {
			return
		}
		text, ok := c.fs.ReadFile(c.path)
		if !ok {
			return
		}
		c.sfc = vuesfc.Extract(text)
		c.ok = true
	})
}

// IsComponent reports whether this file is a Vue single file component whose
// text could be read.
func (c *Component) IsComponent() bool {
	c.load()
	return c != nil && c.ok
}

// ScriptSetupRange returns the content range of the component's `<script setup>`
// block. A component with no such block reports false.
func (c *Component) ScriptSetupRange() (core.TextRange, bool) {
	c.load()
	if c == nil || !c.ok {
		return core.TextRange{}, false
	}
	for _, block := range c.sfc.Blocks {
		if block.Kind == vuesfc.BlockScript && block.Setup {
			return block.Content, true
		}
	}
	return core.TextRange{}, false
}

// IsExposedToTemplate reports whether a Vue component's template can read or
// write a binding without the script ever mentioning it again.
//
// A component with a `<script setup>` block exposes every top-level binding to
// its template. That includes the bindings of a plain `<script>` beside it,
// because compileScript merges both blocks' top-level bindings into the
// metadata the template is compiled against, and the template may assign to a
// `let` as well as read it. For a rule that reasons about how a binding is used,
// the template is the same kind of consumer an `/* exported */` directive
// describes: code the rule cannot see.
//
// A binding that emits nothing at run time, such as an interface, a type alias
// or a type-only import, is never exposed, so it is still judged normally.
//
// The answer is deliberately coarse. It does not look at what the template
// references, so a top-level binding the template never uses is not reported
// either. Missing that case is safer than reporting a binding the template does
// use, and a precise answer needs the template parsed.
//
// symbol must be the raw binder symbol attached to the declaration, as for
// [RuleContext.IsExportedGlobalBinding].
//
// https://github.com/vuejs/core/blob/main/packages/compiler-sfc/src/compileScript.ts
func (ctx *RuleContext) IsExposedToTemplate(symbol *ast.Symbol, name string) bool {
	if ctx == nil || symbol == nil || ctx.SourceFile == nil {
		return false
	}
	if _, ok := ctx.Component.ScriptSetupRange(); !ok {
		return false
	}
	if symbol.Flags&(ast.SymbolFlagsValue|ast.SymbolFlagsAlias) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if ast.IsPartOfTypeOnlyImportOrExportDeclaration(declaration) {
			return false
		}
	}
	return ctx.isSourceFileBinding(symbol, name)
}
