package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
)

func configFile(p *program.Program, name string) *ast.SourceFile {
	q := p.CompilerQueries()
	if q == nil {
		return nil
	}
	path := tspath.ToPath(name, p.CurrentDirectory(), p.FS().UseCaseSensitiveFileNames())
	config := q.CommandLine()
	if config.ConfigFile == nil || config.ConfigFile.SourceFile.Path() != path {
		config = q.ResolvedProjectReference(path)
	}
	if config != nil && config.ConfigFile != nil {
		return config.ConfigFile.SourceFile
	}
	return nil
}

func programQuery(ctx context.Context, p *program.Program, r *ProgramRequest, node *ast.Node) (Result, error) {
	q := p.CompilerQueries()
	if q == nil {
		return nil, errors.New("program query requires a typed project")
	}
	var file *ast.SourceFile
	if node != nil && r.Op != "modeForUsageLocation" {
		if node.Kind != ast.KindSourceFile {
			return nil, errors.New("expected a Program source file")
		}
		file = node.AsSourceFile()
	}
	switch r.Op {
	case "rootFileNames":
		return Strings(p.RootFileNames()), nil
	case "nodeCount":
		count := 0
		for _, f := range p.SourceFiles() {
			count += f.NodeCount
		}
		return Handle(count), nil
	case "identifierCount":
		return Handle(q.IdentifierCount()), nil
	case "typeCount":
		return Handle(q.TypeCount()), nil
	case "symbolCount":
		return Handle(q.SymbolCount()), nil
	case "instantiationCount":
		return Handle(q.InstantiationCount()), nil
	case "relationCacheSizes":
		sizes := q.RelationCacheSizes()
		return Handles(sizes[:]), nil
	case "modeForUsageLocation":
		source := p.GetSourceFile(r.File)
		if source == nil || node == nil {
			return nil, errors.New("expected source and module specifier")
		}
		return Handle(p.GetModeForUsageLocation(source, node)), nil
	case "modeForResolutionAtIndex":
		if file == nil || r.Index < 0 || r.Index >= len(file.Imports()) {
			return nil, errors.New("invalid module resolution index")
		}
		return Handle(p.GetModeForUsageLocation(file, file.Imports()[r.Index])), nil
	case "projectReferences":
		data, err := json.Marshal(projectReferences(q.CommandLine()))
		return Text(data), err
	case "resolvedProjectReferences":
		return resolvedReferences(p)
	}
	kind, ok := diagnosticKinds[r.Op]
	if !ok {
		return nil, errors.New("unknown Program query")
	}
	return diagnosticFacts(q.Diagnostics(ctx, kind, file)), ctx.Err()
}

var diagnosticKinds = map[Operation]program.DiagnosticKind{
	"configFileParsingDiagnostics": program.ConfigFileParsingDiagnostics,
	"optionsDiagnostics":           program.OptionsDiagnostics,
	"globalDiagnostics":            program.GlobalDiagnostics,
	"syntacticDiagnostics":         program.SyntacticDiagnostics,
	"semanticDiagnostics":          program.SemanticDiagnostics,
	"declarationDiagnostics":       program.DeclarationDiagnostics,
}

type projectReference struct {
	Path         string `json:"path"`
	OriginalPath string `json:"originalPath"`
	Circular     *bool  `json:"circular,omitempty"`
	Prepend      *bool  `json:"prepend,omitempty"`
}

func projectReferences(c *tsoptions.ParsedCommandLine) []projectReference {
	refs := c.ProjectReferences()
	if refs == nil {
		return nil
	}
	var raw struct {
		References []projectReference `json:"references"`
	}
	if b, err := json.Marshal(c.Raw); err == nil {
		_ = json.Unmarshal(b, &raw)
	}
	result := make([]projectReference, len(refs))
	for i, r := range refs {
		v := projectReference{Path: r.Path, OriginalPath: r.OriginalPath}
		for _, input := range raw.References {
			if input.Path == r.OriginalPath {
				v.Prepend, v.Circular = input.Prepend, input.Circular
				break
			}
		}
		if r.Circular {
			v.Circular = &r.Circular
		}
		result[i] = v
	}
	return result
}

type referenceConfig struct {
	Name                string                `json:"name"`
	Path                tspath.Path           `json:"path"`
	Options             map[string]any        `json:"options"`
	FileNames           []string              `json:"fileNames"`
	ProjectReferences   []projectReference    `json:"projectReferences,omitempty"`
	References          []int                 `json:"references,omitempty"`
	ExtendedSourceFiles []string              `json:"extendedSourceFiles,omitempty"`
	WildcardDirectories map[string]uint8      `json:"wildcardDirectories"`
	Raw                 any                   `json:"raw,omitempty"`
	CompileOnSave       *bool                 `json:"compileOnSave,omitempty"`
	WatchOptions        *core.WatchOptions    `json:"watchOptions,omitempty"`
	TypeAcquisition     *core.TypeAcquisition `json:"typeAcquisition,omitempty"`
	Diagnostics         []byte                `json:"diagnostics,omitempty"`
}

// A flat graph preserves repeated references, cycles, and unresolved slots.
func resolvedReferences(p *program.Program) (Result, error) {
	q := p.CompilerQueries()
	roots := q.ResolvedProjectReferences()
	if roots == nil {
		return Text("null"), nil
	}
	graph := struct {
		Roots   []int              `json:"roots"`
		Configs []*referenceConfig `json:"configs"`
	}{Roots: make([]int, len(roots)), Configs: []*referenceConfig{}}
	seen := make(map[*tsoptions.ParsedCommandLine]int)
	var visit func(*tsoptions.ParsedCommandLine) (int, error)
	visit = func(c *tsoptions.ParsedCommandLine) (int, error) {
		if c == nil || c.ConfigFile == nil {
			return 0, nil
		}
		if id := seen[c]; id != 0 {
			return id, nil
		}
		source := c.ConfigFile.SourceFile
		options, err := compilerOptions(c.CompilerOptions())
		if err != nil {
			return 0, err
		}
		fact := &referenceConfig{Name: source.FileName(), Path: source.Path(), Options: options, FileNames: slices.Concat([]string{}, c.FileNames()), ProjectReferences: projectReferences(c), Raw: c.Raw, CompileOnSave: c.CompileOnSave, WatchOptions: c.ParsedConfig.WatchOptions, TypeAcquisition: c.TypeAcquisition()}
		fact.ExtendedSourceFiles = c.ExtendedSourceFiles()
		dirs := c.WildcardDirectories()
		fact.WildcardDirectories = make(map[string]uint8, len(dirs))
		for path, recursive := range dirs {
			// TypeScript WatchDirectoryFlags.None and Recursive are 0 and 1.
			var flags uint8
			if recursive {
				flags = 1
			}
			fact.WildcardDirectories[path] = flags
		}
		graph.Configs = append(graph.Configs, fact)
		id := len(graph.Configs)
		seen[c] = id
		if len(c.Errors) != 0 {
			fact.Diagnostics, err = EncodePacket(diagnosticFacts(c.Errors))
			if err != nil {
				return 0, err
			}
		}
		for _, child := range c.ResolvedProjectReferencePaths() {
			target := q.ResolvedProjectReference(tspath.ToPath(child, p.CurrentDirectory(), p.FS().UseCaseSensitiveFileNames()))
			childID, err := visit(target)
			if err != nil {
				return 0, err
			}
			fact.References = append(fact.References, childID)
		}
		return id, nil
	}
	for i, c := range roots {
		id, err := visit(c)
		if err != nil {
			return nil, err
		}
		graph.Roots[i] = id
	}
	data, err := json.Marshal(graph)
	return Text(data), err
}
