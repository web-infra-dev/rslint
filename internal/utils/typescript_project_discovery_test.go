package utils_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type selectionObservation struct {
	Config      string   `json:"config"`
	Roots       []string `json:"roots"`
	Program     int      `json:"program"`
	Strict      *bool    `json:"strict"`
	Diagnostics []int    `json:"diagnostics"`
	Error       bool     `json:"error"`
	Unowned     bool     `json:"unowned"`
}

type selectionExpectation struct {
	RootDirectory string      `json:"rootDirectory"`
	Symlinks      [][2]string `json:"symlinks"`
	Steps         []struct {
		File          string `json:"file"`
		RootDirectory string `json:"rootDirectory"`
		selectionObservation
		Rslint *selectionObservation `json:"rslint"`
	} `json:"steps"`
}

type recordingHost struct {
	fs     vfs.FS
	parsed map[string]int
	built  map[string]int
}

type caseSensitivityFS struct {
	vfs.FS
	caseSensitive bool
}

func (fsys *caseSensitivityFS) UseCaseSensitiveFileNames() bool {
	return fsys.caseSensitive
}

func newRecordingHost(fsys vfs.FS) *recordingHost {
	return &recordingHost{fs: fsys, parsed: make(map[string]int), built: make(map[string]int)}
}

func (h *recordingHost) compilerHost(configPath string) compiler.CompilerHost {
	return compiler.NewCompilerHost(tspath.GetDirectoryPath(configPath), h.fs, bundled.LibPath(), nil, nil, nil)
}

type recordingProjectHost struct {
	FS            vfs.FS
	ParseConfig   func(string) (*tsoptions.ParsedCommandLine, error)
	CreateProgram func(string, *tsoptions.ParsedCommandLine) (*compiler.Program, error)
}

func (h *recordingHost) host() recordingProjectHost {
	return recordingProjectHost{
		FS: h.fs,
		ParseConfig: func(configPath string) (*tsoptions.ParsedCommandLine, error) {
			h.parsed[configPath]++
			parsed, _ := tsoptions.GetParsedCommandLineOfConfigFile(configPath, &core.CompilerOptions{}, nil, h.compilerHost(configPath), nil)
			return parsed, nil
		},
		CreateProgram: func(configPath string, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
			h.built[configPath]++
			return utils.CreateProgramFromParsedConfigLenientWithProjectReferences(true, parsed, h.compilerHost(configPath))
		},
	}
}

func TestServiceProgramKeepsParsedConfigModesIsolated(t *testing.T) {
	root := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/typescript_project_discovery.txtar").Materialize(t, "js-listed-allowjs-false"))
	host := newRecordingHost(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
	configPath := tspath.ResolvePath(root, "tsconfig.json")
	parsed, err := host.host().ParseConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	originalOptions := parsed.CompilerOptions()
	originalExtensions := originalOptions.AllowNonTsExtensions
	service, err := host.host().CreateProgram(configPath, parsed)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.CompilerOptions() != originalOptions || originalOptions.AllowNonTsExtensions != originalExtensions {
		t.Fatal("service construction mutated the caller's parsed options")
	}
	if !service.Options().AllowNonTsExtensions.IsTrue() || service.Options().AllowJs != core.TSFalse {
		t.Fatal("service must allow listed JS sources without changing allowJs")
	}
	if service.CommandLine().ConfigFile != parsed.ConfigFile || !slices.Equal(service.CommandLine().FileNames(), parsed.FileNames()) {
		t.Fatal("service construction changed config identity or complete roots")
	}
	file := tspath.ResolvePath(root, "gap.js")
	if service.GetSourceFile(file) == nil {
		t.Fatal("service must retain the listed JavaScript source")
	}
	legacy, err := utils.CreateProgramFromParsedConfigLenient(true, parsed, host.compilerHost(configPath))
	if err != nil {
		t.Fatal(err)
	}
	if legacy.GetSourceFile(file) != nil || legacy.Options().AllowNonTsExtensions != originalExtensions {
		t.Fatal("legacy construction borrowed the service extension policy")
	}
}

func fixtureScenarios(t *testing.T, archive *txtarfs.Archive) []string {
	t.Helper()
	files, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, file := range files {
		if strings.HasSuffix(file, "/expect.json") {
			names = append(names, strings.TrimSuffix(file, "/expect.json"))
		}
	}
	if len(names) == 0 {
		t.Fatal("no project-service scenarios selected")
	}
	return names
}

func TestTypeScriptProjectDiscoveryObservations(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/typescript_project_discovery.txtar")
	for _, name := range fixtureScenarios(t, archive) {
		t.Run(name, func(t *testing.T) {
			root := tspath.NormalizePath(archive.Materialize(t, name))
			data, err := archive.ReadFile(name + "/expect.json")
			if err != nil {
				t.Fatal(err)
			}
			var expected selectionExpectation
			if err := json.Unmarshal(data, &expected); err != nil {
				t.Fatal(err)
			}
			if len(expected.Steps) == 0 {
				t.Fatal("scenario contains no target observations")
			}
			for _, link := range expected.Symlinks {
				linkPath := tspath.ResolvePath(root, link[0])
				if err := os.MkdirAll(tspath.GetDirectoryPath(linkPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(tspath.ResolvePath(root, link[1]), linkPath); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			host := newRecordingHost(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
			discovery := utils.NewTypeScriptProjectDiscovery(host.fs, host.host().ParseConfig)
			// Upstream Program IDs identify configured projects. Discovery only owns
			// parsed metadata; loader integration tests verify actual Program reuse.
			projects := make(map[int]*tsoptions.ParsedCommandLine)
			for _, step := range expected.Steps {
				if step.Rslint != nil {
					step.selectionObservation = *step.Rslint
				}
				fileName := tspath.ResolvePath(root, step.File)
				rootDirectory := expected.RootDirectory
				if step.RootDirectory != "" {
					rootDirectory = step.RootDirectory
				}
				selected, err := discovery.Find(fileName, tspath.ResolvePath(root, rootDirectory))
				var program *compiler.Program
				if err == nil && selected != nil {
					program, err = host.host().CreateProgram(selected.ConfigName(), selected)
					if err == nil && program.GetSourceFile(fileName) == nil {
						err = errors.New("selected config cannot supply its source file")
					}
				}
				if step.Error {
					if err == nil {
						t.Errorf("%s: expected rejection, selected %s", step.File, selected.ConfigName())
					}
					continue
				}
				if err != nil {
					t.Errorf("%s: %v", step.File, err)
					continue
				}
				if step.Unowned {
					if selected != nil {
						t.Errorf("%s: unowned target selected %s", step.File, selected.ConfigName())
					}
					continue
				}
				if selected == nil {
					t.Errorf("%s: no selected config, want %s", step.File, step.Config)
					continue
				}
				if selected.ConfigName() != tspath.ResolvePath(root, step.Config) {
					t.Errorf("%s: selected %s, want %s", step.File, selected.ConfigName(), step.Config)
				}
				if program.GetSourceFile(fileName) == nil {
					t.Errorf("%s: selected Program does not contain its source", step.File)
				}
				wantRoots := make([]string, len(step.Roots))
				for i, file := range step.Roots {
					wantRoots[i] = tspath.ResolvePath(root, file)
				}
				gotRoots := slices.Clone(program.CommandLine().FileNames())
				slices.Sort(wantRoots)
				slices.Sort(gotRoots)
				if !slices.Equal(gotRoots, wantRoots) {
					t.Errorf("%s: complete roots = %v, want %v", step.File, gotRoots, wantRoots)
				}
				if step.Strict != nil && program.Options().Strict.IsTrue() != *step.Strict {
					t.Errorf("%s: strict = %v, want %v", step.File, program.Options().Strict.IsTrue(), *step.Strict)
				}
				var diagnosticCodes []int
				for _, diagnostic := range program.GetSemanticDiagnostics(context.Background(), program.GetSourceFile(fileName)) {
					diagnosticCodes = append(diagnosticCodes, int(diagnostic.Code()))
				}
				slices.Sort(diagnosticCodes)
				slices.Sort(step.Diagnostics)
				if !slices.Equal(diagnosticCodes, step.Diagnostics) {
					t.Errorf("%s: semantic diagnostics = %v, want %v", step.File, diagnosticCodes, step.Diagnostics)
				}
				if previous := projects[step.Program]; previous != nil && previous != selected {
					t.Errorf("%s: parsed project metadata was not reused", step.File)
				}
				for id, previous := range projects {
					if id != step.Program && previous == selected {
						t.Errorf("%s: distinct TypeScript projects shared metadata", step.File)
					}
				}
				projects[step.Program] = selected
			}
			for configPath, count := range host.parsed {
				if count != 1 {
					t.Errorf("parsed %s %d times", configPath, count)
				}
			}
		})
	}
}

// readTrackingFS catches source reads during discovery, including candidates
// that would have admitted the target only after resolving an import.
type discoveryReadTrackingFS struct {
	vfs.FS
	sources []string
}

func (fsys *discoveryReadTrackingFS) ReadFile(name string) (string, bool) {
	if !strings.HasSuffix(name, ".json") {
		fsys.sources = append(fsys.sources, name)
	}
	return fsys.FS.ReadFile(name)
}

func TestTypeScriptProjectDiscoveryReadsOnlyMetadata(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/typescript_project_discovery.txtar")
	for _, test := range []struct{ name, target, want string }{
		{"nearest-nested", "pkg/src/file.ts", "pkg/tsconfig.json"},
		{"solution-first-deep-versus-second-shallow", "target/file.ts", "b/custom.json"},
		{"no-config", "src/file.ts", ""},
		{"excluded-imported", "outside/file.ts", ""},
		{"composite-excluded-imported", "outside/file.ts", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := tspath.NormalizePath(archive.Materialize(t, test.name))
			fsys := &discoveryReadTrackingFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS()))}
			host := newRecordingHost(fsys)
			selected, err := utils.NewTypeScriptProjectDiscovery(fsys, host.host().ParseConfig).Find(tspath.ResolvePath(root, test.target), root)
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if test.want != "" {
				want = tspath.ResolvePath(root, test.want)
			}
			if selected.ConfigName() != want {
				t.Fatalf("selected %s, want %s", selected.ConfigName(), want)
			}
			if len(fsys.sources) != 0 {
				t.Fatalf("discovery read source files: %v", fsys.sources)
			}
		})
	}
}

func TestTypeScriptProjectDiscoveryPropagatesParseFailure(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/typescript_project_discovery.txtar")
	root := tspath.NormalizePath(archive.Materialize(t, "nearest-root"))
	want := errors.New("fixture host failure")
	for _, malformedHost := range []bool{false, true} {
		host := newRecordingHost(bundled.WrapFS(osvfs.FS())).host()
		parse := func(string) (*tsoptions.ParsedCommandLine, error) {
			if malformedHost {
				return nil, nil //nolint:nilnil // Model a parser that violated its result contract.
			}
			return nil, want
		}
		_, err := utils.NewTypeScriptProjectDiscovery(host.FS, parse).Find(tspath.ResolvePath(root, "src/file.ts"), root)
		if err == nil || !malformedHost && !errors.Is(err, want) {
			t.Fatalf("error = %v, want parser failure", err)
		}
	}
}

func TestSelectionKeepsLexicalConfigIdentity(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/typescript_project_discovery.txtar")
	for _, caseSensitive := range []bool{false, true} {
		name := "case-insensitive-node-modules"
		if caseSensitive {
			name = "case-sensitive-node-modules"
		}
		t.Run(name, func(t *testing.T) {
			root := tspath.NormalizePath(archive.Materialize(t, "node-modules-case-boundary"))
			fsys := &caseSensitivityFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS())), caseSensitive: caseSensitive}
			host := newRecordingHost(fsys)
			selected, err := utils.NewTypeScriptProjectDiscovery(host.fs, host.host().ParseConfig).Find(tspath.ResolvePath(root, "NODE_MODULES/pkg/file.ts"), root)
			if !caseSensitive {
				if err != nil || selected != nil {
					t.Fatalf("node_modules boundary was crossed: selected %s, built %v", selected.ConfigName(), host.built)
				}
			} else if err != nil || selected.ConfigName() != tspath.ResolvePath(root, "tsconfig.json") {
				t.Fatalf("uppercase directory must remain distinct: selection %s, error %v", selected.ConfigName(), err)
			}
		})
	}

	t.Run("case-alias", func(t *testing.T) {
		root := tspath.NormalizePath(archive.Materialize(t, "nearest-root"))
		upper := tspath.NormalizePath(filepath.Join(root, "src", "FILE.ts"))
		info, err := os.Stat(upper)
		if err != nil || info.IsDir() {
			t.Skip("case aliases unavailable on this filesystem")
		}
		host := newRecordingHost(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
		discovery := utils.NewTypeScriptProjectDiscovery(host.fs, host.host().ParseConfig)
		selected, err := discovery.Find(upper, root)
		if err != nil {
			t.Fatal(err)
		}
		if selected.ConfigName() != tspath.ResolvePath(root, "tsconfig.json") {
			t.Fatalf("case alias selected %s", selected.ConfigName())
		}
	})
}
