package loader

import (
	"sync"

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
// loader satisfies that contract while assembling each requested load result.
//
// The result is immutable load metadata. It is not retained by the Session or
// Programs because target hints and scheduling policy are operation-specific.
func (s *Session) FreezeProgramRootIdentities(
	programs []*lintprogram.Program,
	knownTargets []target.File,
	singleThreaded bool,
) ([]string, error) {
	projector, err := s.NewProgramRootIdentityProjector(knownTargets, singleThreaded)
	if err != nil {
		return nil, err
	}
	return projector.Freeze(programs), nil
}

// ProgramRootIdentityProjector freezes actual Program roots across one
// operation's admitted loads. Target hints are immutable and shared; unknown
// roots use one bounded identity-resolution window. It retains no Programs,
// source files, or resolved-root cache between calls.
type ProgramRootIdentityProjector struct {
	resolver *projectRootIdentityResolver
	unknown  sync.Mutex
}

// NewProgramRootIdentityProjector starts an operation over this Session's
// immutable VFS generation and the complete operation's target hints.
func (s *Session) NewProgramRootIdentityProjector(knownTargets []target.File, singleThreaded bool) (*ProgramRootIdentityProjector, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	return &ProgramRootIdentityProjector{resolver: newProjectRootIdentityResolver(knownTargets, s.FS(), singleThreaded)}, nil
}

// Freeze returns physical root identities in first-declaration order. Calls
// may overlap, and every Program must belong to the projector's source/VFS
// generation. Each caller owns its result independently of other loads.
func (p *ProgramRootIdentityProjector) Freeze(programs []*lintprogram.Program) []string {
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

	resolver := p.resolver
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
	resolvedUnknown := p.freezeUnknown(unknownPaths)
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
	return identities
}

func (p *ProgramRootIdentityProjector) freezeUnknown(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	p.unknown.Lock()
	defer p.unknown.Unlock()
	return p.resolver.canonicalRootPathIDs(paths)
}
