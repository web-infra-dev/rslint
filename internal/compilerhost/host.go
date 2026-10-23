// Package compilerhost retains the base directory needed by rslint's program
// builders alongside TypeScript's directory-independent compiler host.
package compilerhost

import (
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

type Host interface {
	compiler.CompilerHost
	GetCurrentDirectory() tspath.RootedDirectoryPath
}

type host struct {
	compiler.CompilerHost
	directory tspath.RootedDirectoryPath
}

func WithDirectory(base compiler.CompilerHost, directory tspath.RootedDirectoryPath) Host {
	return &host{CompilerHost: base, directory: directory}
}

func (h *host) GetCurrentDirectory() tspath.RootedDirectoryPath { return h.directory }
