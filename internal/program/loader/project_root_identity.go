package loader

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

// FreezeProgramRootIdentities returns the stable physical identities of every
// root owned by a Program capable of program-wide diagnostics. Known
// lint-target identities seed the operation, and all remaining roots are
// resolved in one bounded per-file pass. The returned exact IDs are
// deduplicated in first-declaration order.
//
// Programs must belong to this Session's loader generation and VFS. The
// command satisfies that contract by requesting the projection immediately
// after loading and before execution.
//
// The result belongs to this call only. It is not retained by the Session,
// Programs, or their backing projects because target hints and scheduling
// policy are operation-specific rather than project-generation state.
func (s *Session) FreezeProgramRootIdentities(
	programs []*lintprogram.Program,
	knownTargets []target.File,
	singleThreaded bool,
) ([]string, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}

	rootPaths := make([]string, 0)
	seenRootPaths := make(map[string]struct{})
	for _, program := range programs {
		if !program.CanProvideProgramDiagnostics() {
			continue
		}
		for _, rootPath := range program.RootFileNames() {
			rootPath = tspath.NormalizePath(rootPath)
			rootPathID := exactPathID(rootPath)
			if _, seen := seenRootPaths[rootPathID]; seen {
				continue
			}
			seenRootPaths[rootPathID] = struct{}{}
			rootPaths = append(rootPaths, rootPath)
		}
	}

	resolver := newProjectRootIdentityResolver(knownTargets, s.FS(), singleThreaded)
	canonicalIDs := make([]string, len(rootPaths))
	unknownPaths := make([]string, 0, len(rootPaths))
	unknownIndexes := make([]int, 0, len(rootPaths))
	for index, rootPath := range rootPaths {
		rootPathID := exactPathID(rootPath)
		if canonicalID, known := resolver.knownCanonicalID(rootPathID); known {
			canonicalIDs[index] = canonicalID
			continue
		}
		unknownPaths = append(unknownPaths, rootPath)
		unknownIndexes = append(unknownIndexes, index)
	}
	resolvedUnknown := resolver.canonicalRootPathIDs(unknownPaths)
	for index, canonicalID := range resolvedUnknown {
		canonicalIDs[unknownIndexes[index]] = canonicalID
	}

	identities := make([]string, 0, len(canonicalIDs))
	seenIdentities := make(map[string]struct{}, len(canonicalIDs))
	for _, canonicalID := range canonicalIDs {
		if _, seen := seenIdentities[canonicalID]; seen {
			continue
		}
		seenIdentities[canonicalID] = struct{}{}
		identities = append(identities, canonicalID)
	}
	return identities, nil
}
