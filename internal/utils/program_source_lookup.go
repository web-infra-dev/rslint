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

func exactProgramSourcePathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func (lookup *ProgramSourceLookup) canonicalPathID(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if lookup.fs != nil {
		if realPath := lookup.fs.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return exactProgramSourcePathID(filePath)
}

// SourceFileForCandidate validates a Program lookup against the target's exact
// canonical identity. This rejects a case-folded lookup that returned a
// different physical file.
func (lookup *ProgramSourceLookup) SourceFileForCandidate(candidate string, canonicalTarget string) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || candidate == "" {
		return nil
	}
	candidate = tspath.NormalizePath(candidate)
	sourceFile := lookup.program.GetSourceFile(candidate)
	if sourceFile == nil {
		return nil
	}
	if exactProgramSourcePathID(sourceFile.FileName()) == exactProgramSourcePathID(candidate) {
		return sourceFile
	}
	if canonicalTarget == "" {
		return nil
	}
	if lookup.canonicalPathID(sourceFile.FileName()) == lookup.canonicalPathID(canonicalTarget) {
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
	return lookup.canonicalSourceFileByID(lookup.canonicalPathID(canonicalPath))
}

func (lookup *ProgramSourceLookup) canonicalSourceFileByID(canonicalID string) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || lookup.fs == nil || canonicalID == "" {
		return nil
	}
	if !lookup.canonicalIndexBuilt {
		lookup.canonicalIndexBuilt = true
		lookup.canonicalSources = make(map[string]*ast.SourceFile)
		for _, sourceFile := range lookup.program.GetSourceFiles() {
			canonicalID := lookup.canonicalPathID(sourceFile.FileName())
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
	canonicalID := exactProgramSourcePathID(canonicalPath)
	if sourceFile := lookup.program.GetSourceFile(filePath); sourceFile != nil &&
		lookup.canonicalPathID(sourceFile.FileName()) == canonicalID {
		return sourceFile
	}
	if exactProgramSourcePathID(canonicalPath) != exactProgramSourcePathID(filePath) {
		if sourceFile := lookup.program.GetSourceFile(canonicalPath); sourceFile != nil &&
			lookup.canonicalPathID(sourceFile.FileName()) == canonicalID {
			return sourceFile
		}
	}
	return lookup.canonicalSourceFileByID(canonicalID)
}

// SourceFileForPath resolves a lexical filesystem path through exact and
// canonical Program source identities.
func (lookup *ProgramSourceLookup) SourceFileForPath(filePath string) *ast.SourceFile {
	if lookup == nil || lookup.program == nil || filePath == "" {
		return nil
	}
	filePath = tspath.NormalizePath(filePath)
	if sourceFile := lookup.SourceFileForCandidate(filePath, ""); sourceFile != nil {
		return sourceFile
	}

	canonicalPath := filePath
	if lookup.fs != nil {
		if realPath := lookup.fs.Realpath(filePath); realPath != "" {
			canonicalPath = tspath.NormalizePath(realPath)
		}
	}
	if sourceFile := lookup.SourceFileForCandidate(filePath, canonicalPath); sourceFile != nil {
		return sourceFile
	}
	if exactProgramSourcePathID(canonicalPath) != exactProgramSourcePathID(filePath) {
		if sourceFile := lookup.SourceFileForCandidate(canonicalPath, canonicalPath); sourceFile != nil {
			return sourceFile
		}
	}
	return lookup.CanonicalSourceFile(canonicalPath)
}
