package rule

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// SourceBOM lazily answers whether one file's source text began with a Unicode
// byte order mark. The text a rule sees never contains the mark — see
// [utils.BOMSource] — so this is the only witness, and answering costs a read
// of the file's own bytes whenever the text came off disk. One store per file
// is shared by every rule on it, and a file no rule asks about pays nothing.
type SourceBOM struct {
	fs   vfs.FS
	path string
	once sync.Once
	has  bool
}

// NewSourceBOM returns the store for one file. A nil file system yields a
// store that answers false, which is what a context with no program has to
// say.
func NewSourceBOM(fileSystem vfs.FS, path string) *SourceBOM {
	return &SourceBOM{fs: fileSystem, path: path}
}

// Value reports whether the file's source text began with a byte order mark.
func (s *SourceBOM) Value() bool {
	if s == nil || s.fs == nil {
		return false
	}
	s.once.Do(func() {
		s.has = utils.SourceHasBOM(s.fs, s.path)
	})
	return s.has
}

// HasBOM reports whether this file's source text began with a Unicode byte
// order mark, U+FEFF. It is ESLint's `SourceCode#hasBOM`: SourceFile.Text()
// never contains the mark, so a rule that cares about one asks here.
func (ctx RuleContext) HasBOM() bool {
	return ctx.BOM.Value()
}
