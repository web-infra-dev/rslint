package modules

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// IsNodeBuiltin recognizes runtime builtin specifiers, including supported internal names.
// Use IsNodeBuiltinAtVersion for version-specific availability.
func IsNodeBuiltin(specifier string) bool {
	if core.NodeCoreModules()[specifier] {
		return true
	}
	// tsgo intentionally filters out underscore-prefixed internal modules.
	// Keep supported internals; the _stream_* aliases were removed in Node 26.
	switch strings.TrimPrefix(specifier, "node:") {
	case "_http_agent", "_http_client", "_http_common", "_http_incoming", "_http_outgoing", "_http_server",
		"_tls_common", "_tls_wrap":
		return true
	}
	return false
}
