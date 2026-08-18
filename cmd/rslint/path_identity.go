package main

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// These helpers define command-layer path identities used by diagnostic
// remapping and config routing. Program loading owns its own private identity
// index; keeping these presentation/routing helpers here avoids coupling cmd
// consumers to loader internals.
func exactFilesystemPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func authoritativeFilesystemPath(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return filePath
}

func canonicalFilesystemPathID(filePath string, fsys vfs.FS) string {
	return exactFilesystemPathID(authoritativeFilesystemPath(filePath, fsys))
}
