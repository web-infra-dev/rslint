package typeservice

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
)

type configFilesKey struct{}

// Separate parses of the same config can appear in different project references.
func configFiles(p *program.Program) *sync.Map {
	return program.Cached(p, configFilesKey{}, func() *sync.Map { return new(sync.Map) })
}

func configID(file *ast.SourceFile) string {
	return strconv.FormatUint(uint64(ast.GetNodeId(file.AsNode())), 10)
}

func configFile(p *program.Program, id string) *ast.SourceFile {
	value, _ := configFiles(p).Load(id)
	file, _ := value.(*ast.SourceFile)
	return file
}

func parsedConfig(q program.CompilerQueries, path tspath.Path) *tsoptions.ParsedCommandLine {
	c := q.CommandLine()
	if c.ConfigFile != nil && c.ConfigFile.SourceFile.Path() == path {
		return c
	}
	return q.ResolvedProjectReference(path)
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
		if file != nil && r.Index >= 0 {
			imports := file.Imports()
			if r.Index < len(imports) {
				return Handle(p.GetModeForUsageLocation(file, imports[r.Index])), nil
			}
			index := len(imports)
			for _, name := range file.ModuleAugmentations {
				if name.Kind == ast.KindStringLiteral {
					if index == r.Index {
						return Handle(p.GetModeForUsageLocation(file, name)), nil
					}
					index++
				}
			}
		}
		return nil, errors.New("invalid module resolution index")
	case "configSyntax":
		source := configFile(p, r.Config)
		if source == nil {
			return nil, errors.New("unknown project config")
		}
		tree, _, err := encoder.EncodeSourceFile(source)
		return &EncodedTypeNode{tree: tree}, err
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
	return diagnosticFacts(p, q.Diagnostics(ctx, kind, file)), ctx.Err()
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
	Circular     *any   `json:"circular,omitempty"`
	OriginalPath string `json:"originalPath"`
	Path         string `json:"path"`
	Prepend      *any   `json:"prepend,omitempty"`
}

func projectReferences(c *tsoptions.ParsedCommandLine) []projectReference {
	refs := c.ProjectReferences()
	if len(refs) == 0 {
		return nil
	}
	var inputs []any
	if config, ok := c.Raw.(*jsonObject); ok {
		inputs, _ = config.GetOrZero("references").([]any)
	}
	result := make([]projectReference, len(refs))
	for i, r := range refs {
		ref := projectReference{Path: r.Path, OriginalPath: r.OriginalPath}
		// The parser skips invalid entries and preserves duplicate paths in order.
		for len(inputs) != 0 {
			input, _ := inputs[0].(*jsonObject)
			inputs = inputs[1:]
			if input == nil || input.GetOrZero("path") != r.OriginalPath {
				continue
			}
			if value, ok := input.Get("prepend"); ok {
				ref.Prepend = &value
			}
			if value, ok := input.Get("circular"); ok {
				ref.Circular = &value
			}
			break
		}
		if r.Circular && ref.Circular == nil {
			ref.Circular = new(any(true))
		}
		result[i] = ref
	}
	return result
}

type configSourceInfo struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	Path                tspath.Path `json:"path"`
	ExtendedSourceFiles []string    `json:"extendedSourceFiles,omitempty"`
}

func configSource(p *program.Program, c *tsoptions.ParsedCommandLine) *configSourceInfo {
	if c.ConfigFile == nil {
		return nil
	}
	source := c.ConfigFile.SourceFile
	info := registerConfig(p, source)
	info.Path = source.Path()
	info.ExtendedSourceFiles = c.ExtendedSourceFiles()
	return info
}

func registerConfig(p *program.Program, source *ast.SourceFile) *configSourceInfo {
	id := configID(source)
	configFiles(p).LoadOrStore(id, source)
	return &configSourceInfo{ID: id, Name: source.FileName()}
}

type referenceConfig struct {
	*configSourceInfo
	Options             map[string]any        `json:"options"`
	FileNames           []string              `json:"fileNames"`
	ProjectReferences   []projectReference    `json:"projectReferences,omitempty"`
	References          []int                 `json:"references,omitempty"`
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
		options, err := compilerOptions(c)
		if err != nil {
			return 0, err
		}
		fact := &referenceConfig{configSourceInfo: configSource(p, c), Options: options, FileNames: slices.Concat([]string{}, c.FileNames()), ProjectReferences: projectReferences(c), Raw: c.Raw, CompileOnSave: c.CompileOnSave, WatchOptions: c.ParsedConfig.WatchOptions, TypeAcquisition: c.TypeAcquisition()}
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
			fact.Diagnostics, err = EncodePacket(diagnosticFacts(p, c.Errors))
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
