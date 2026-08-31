package main

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// These live-filesystem helpers reconcile compiler-owned paths from TypeScript
// diagnostics and Program roots. Frozen target/config paths use
// config.ExactPathID directly.
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
	return rslintconfig.ExactPathID(authoritativeFilesystemPath(filePath, fsys))
}
