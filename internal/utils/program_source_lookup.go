package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// ProgramSourceLookup resolves a filesystem target to the exact source file in
// one Program. Its canonical source index is built lazily and lives for the
// owning Program generation.
type ProgramSourceLookup struct {
	program             *compiler.Program
	fs                  vfs.FS
	canonicalIndexBuilt bool
	canonicalSources    map[string]*ast.SourceFile
}

// NewProgramSourceLookup creates an exact source lookup for one Program.
func NewProgramSourceLookup(program *compiler.Program, fs vfs.FS) *ProgramSourceLookup {
	return &ProgramSourceLookup{program: program, fs: fs}
}

// UpdateProgramForReusedGraph preserves path identities after a source edit.
// The caller must have received a reused graph from Program.UpdateProgram:
// membership and resolution are unchanged, but source objects may be replaced.
func (lookup *ProgramSourceLookup) UpdateProgramForReusedGraph(program *compiler.Program, fs vfs.FS) {
	lookup.program = program
	lookup.fs = fs
	for canonicalID, sourceFile := range lookup.canonicalSources {
		lookup.canonicalSources[canonicalID] = program.GetSourceFileByPath(sourceFile.Path())
	}
}

func canonicalSourcePathID(filePath string, fs vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fs != nil {
		if realPath := fs.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return ExactPathID(filePath)
}

// SourceFileForCandidate validates a Program lookup against the target's exact
// canonical identity. This rejects a case-folded lookup that returned a
// different physical file.
func (lookup *ProgramSourceLookup) SourceFileForCandidate(candidate string, canonicalTarget string) *ast.SourceFile {
	if lookup == nil {
		return nil
	}
	return lookup.sourceFileForCandidate(candidate, canonicalTarget, lookup.fs)
}

func (lookup *ProgramSourceLookup) sourceFileForCandidate(candidate string, canonicalTarget string, fs vfs.FS) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || candidate == "" {
		return nil
	}
	candidate = tspath.NormalizePath(candidate)
	sourceFile := lookup.program.GetSourceFile(candidate)
	if sourceFile == nil {
		return nil
	}
	if ExactPathID(sourceFile.FileName()) == ExactPathID(candidate) {
		return sourceFile
	}
	if canonicalTarget == "" {
		return nil
	}
	if canonicalSourcePathID(sourceFile.FileName(), fs) == canonicalSourcePathID(canonicalTarget, fs) {
		return sourceFile
	}
	return nil
}

// CanonicalSourceFile finds a Program source by physical filesystem identity.
// The index is created only after direct candidate lookups miss.
func (lookup *ProgramSourceLookup) CanonicalSourceFile(canonicalPath string) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || lookup.fs == nil || canonicalPath == "" {
		return nil
	}
	return lookup.canonicalSourceFileByID(canonicalSourcePathID(canonicalPath, lookup.fs), lookup.fs)
}

func (lookup *ProgramSourceLookup) canonicalSourceFileByID(canonicalID string, fs vfs.FS) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || fs == nil || canonicalID == "" {
		return nil
	}
	if !lookup.canonicalIndexBuilt {
		lookup.canonicalIndexBuilt = true
		lookup.canonicalSources = make(map[string]*ast.SourceFile)
		for _, sourceFile := range lookup.program.GetSourceFiles() {
			canonicalID := canonicalSourcePathID(sourceFile.FileName(), fs)
			existing := lookup.canonicalSources[canonicalID]
			if existing == nil || sourceFile.FileName() < existing.FileName() {
				lookup.canonicalSources[canonicalID] = sourceFile
			}
		}
	}
	return lookup.canonicalSources[canonicalID]
}

// SourceFileForTarget resolves a caller-visible path using a canonical target
// identity that was frozen by an earlier discovery/snapshot boundary. It never
// performs another filesystem identity lookup for either spelling, so source
// selection cannot observe a different generation from config selection.
func (lookup *ProgramSourceLookup) SourceFileForTarget(filePath string, canonicalPath string) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || filePath == "" {
		return nil
	}
	filePath = tspath.NormalizePath(filePath)
	if sourceFile := lookup.SourceFileForCandidate(filePath, ""); sourceFile != nil {
		return sourceFile
	}
	if canonicalPath == "" {
		return nil
	}
	canonicalPath = tspath.NormalizePath(canonicalPath)
	canonicalID := ExactPathID(canonicalPath)
	if sourceFile := lookup.program.GetSourceFile(filePath); sourceFile != nil &&
		canonicalSourcePathID(sourceFile.FileName(), lookup.fs) == canonicalID {
		return sourceFile
	}
	if ExactPathID(canonicalPath) != ExactPathID(filePath) {
		if sourceFile := lookup.program.GetSourceFile(canonicalPath); sourceFile != nil &&
			canonicalSourcePathID(sourceFile.FileName(), lookup.fs) == canonicalID {
			return sourceFile
		}
	}
	return lookup.canonicalSourceFileByID(canonicalID, lookup.fs)
}

// SourceFileForPath resolves a lexical filesystem path through exact and
// canonical Program source identities. fs supplies this observation's identity
// queries and is not retained; it must preserve the Program's overlay mapping.
func (lookup *ProgramSourceLookup) SourceFileForPath(filePath string, fs vfs.FS) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || filePath == "" {
		return nil
	}
	filePath = tspath.NormalizePath(filePath)
	if sourceFile := lookup.sourceFileForCandidate(filePath, "", fs); sourceFile != nil {
		return sourceFile
	}

	canonicalPath := filePath
	if fs != nil {
		if realPath := fs.Realpath(filePath); realPath != "" {
			canonicalPath = tspath.NormalizePath(realPath)
		}
	}
	if sourceFile := lookup.sourceFileForCandidate(filePath, canonicalPath, fs); sourceFile != nil {
		return sourceFile
	}
	if ExactPathID(canonicalPath) != ExactPathID(filePath) {
		if sourceFile := lookup.sourceFileForCandidate(canonicalPath, canonicalPath, fs); sourceFile != nil {
			return sourceFile
		}
	}
	return lookup.canonicalSourceFileByID(canonicalSourcePathID(canonicalPath, fs), fs)
}
