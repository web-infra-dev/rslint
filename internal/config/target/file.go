package target

import rslintconfig "github.com/web-infra-dev/rslint/internal/config"

// File preserves the caller-visible path, physical file and
// parent identities, and config owner established at the target boundary.
// CanonicalParentPath distinguishes a leaf file symlink from a directory
// alias without asking the live filesystem again during config matching.
type File struct {
	rslintconfig.PathIdentity
	ConfigDirectory string
}

// Identity returns the config matching input carried by this lint file.
func (file File) Identity() rslintconfig.PathIdentity {
	return file.PathIdentity
}
