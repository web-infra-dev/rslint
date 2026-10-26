package main

import (
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/lsp"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func runLSP(args []string) int {
	s := lsp.NewServer(&lsp.ServerOptions{
		In:                 lsp.ToReader(os.Stdin),
		Out:                lsp.ToWriter(os.Stdout),
		Err:                os.Stderr,
		Cwd:                utils.Must(os.Getwd()),
		FS:                 bundled.WrapFS(osvfs.FS()),
		DefaultLibraryPath: bundled.LibPath(),
	})
	if err := s.Run(); err != nil {
		return 1
	}
	return 0
}
