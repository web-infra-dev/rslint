package modules

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// IsNodeBuiltin recognizes runtime builtin specifiers, including legacy internal names.
// Rules may apply narrower version or replacement policies to this set.
func IsNodeBuiltin(specifier string) bool {
	if core.NodeCoreModules()[specifier] {
		return true
	}
	// tsgo intentionally filters out underscore-prefixed internal modules.
	// Node's isBuiltin still includes these legacy names (verified on Node 22).
	switch strings.TrimPrefix(specifier, "node:") {
	case "_http_agent", "_http_client", "_http_common", "_http_incoming", "_http_outgoing", "_http_server",
		"_stream_duplex", "_stream_passthrough", "_stream_readable", "_stream_transform", "_stream_wrap", "_stream_writable",
		"_tls_common", "_tls_wrap":
		return true
	}
	return false
}
