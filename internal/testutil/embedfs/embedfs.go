// Package embedfs builds vfs roots backed entirely in memory by an embedded
// filesystem, for locating test fixtures without runtime.Caller (which
// returns a non-absolute, module-relative path when built with -trimpath).
//
// It's kept separate from internal/rule_tester, which the fixtures packages
// consuming Root also import from their _test.go files: a fixtures package
// importing internal/rule_tester directly would create an import cycle for
// any package (like internal/rule and internal/utils) whose own tests both
// import fixtures and are imported by internal/rule_tester itself.
package embedfs

import (
	"embed"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/iovfs"
)

// Root identifies where a TypeScript program's files should be read from.
type Root struct {
	Dir string
	FS  vfs.FS
}

// EmbedRoot builds a Root backed entirely in memory by fsys.
func EmbedRoot(fsys embed.FS) Root {
	return Root{
		Dir: "/",
		FS:  bundled.WrapFS(iovfs.From(fsys, true)),
	}
}
