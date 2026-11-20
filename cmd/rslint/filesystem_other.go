//go:build !darwin

package main

import (
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

func newCLIFileSystem() vfs.FS {
	return bundled.WrapFS(cachedvfs.From(osvfs.FS()))
}

func newLSPFileSystem() vfs.FS {
	return bundled.WrapFS(osvfs.FS())
}

func getCLIWorkingDirectory() (string, error) {
	return os.Getwd()
}
