package projectservice_test

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
	"github.com/web-infra-dev/rslint/internal/program/projectservice"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type selectionExpectation struct {
	RootDirectory string      `json:"rootDirectory"`
	Symlinks      [][2]string `json:"symlinks"`
	Steps         []struct {
		File          string   `json:"file"`
		RootDirectory string   `json:"rootDirectory"`
		Config        string   `json:"config"`
		Roots         []string `json:"roots"`
		Program       int      `json:"program"`
		Strict        *bool    `json:"strict"`
		Diagnostics   []int    `json:"diagnostics"`
		Error         bool     `json:"error"`
		Unowned       bool     `json:"unowned"`
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

func (h *recordingHost) host() projectservice.Host {
	return projectservice.Host{
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
	root := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/upstream.txtar").Materialize(t, "js-listed-allowjs-false"))
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

func TestUpstreamProjectSelection(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
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
			selector := projectservice.New(host.host())
			programs := make(map[int]*compiler.Program)
			for _, step := range expected.Steps {
				fileName := tspath.ResolvePath(root, step.File)
				rootDirectory := expected.RootDirectory
				if step.RootDirectory != "" {
					rootDirectory = step.RootDirectory
				}
				selected, err := selector.Select(fileName, tspath.ResolvePath(root, rootDirectory))
				if step.Error {
					if err == nil {
						t.Errorf("%s: expected rejection, selected %s", step.File, selected.ConfigPath)
					}
					continue
				}
				if err != nil {
					t.Errorf("%s: %v", step.File, err)
					continue
				}
				if step.Unowned {
					if selected.Program != nil || selected.ConfigPath != "" {
						t.Errorf("%s: unowned target selected %s", step.File, selected.ConfigPath)
					}
					continue
				}
				if selected.ConfigPath != tspath.ResolvePath(root, step.Config) {
					t.Errorf("%s: selected %s, want %s", step.File, selected.ConfigPath, step.Config)
				}
				if selected.Program.GetSourceFile(fileName) == nil {
					t.Errorf("%s: selected Program does not contain its source", step.File)
				}
				wantRoots := make([]string, len(step.Roots))
				for i, file := range step.Roots {
					wantRoots[i] = tspath.ResolvePath(root, file)
				}
				gotRoots := slices.Clone(selected.Program.CommandLine().FileNames())
				slices.Sort(wantRoots)
				slices.Sort(gotRoots)
				if !slices.Equal(gotRoots, wantRoots) {
					t.Errorf("%s: complete roots = %v, want %v", step.File, gotRoots, wantRoots)
				}
				if step.Strict != nil && selected.Program.Options().Strict.IsTrue() != *step.Strict {
					t.Errorf("%s: strict = %v, want %v", step.File, selected.Program.Options().Strict.IsTrue(), *step.Strict)
				}
				var diagnosticCodes []int
				for _, diagnostic := range selected.Program.GetSemanticDiagnostics(context.Background(), selected.Program.GetSourceFile(fileName)) {
					diagnosticCodes = append(diagnosticCodes, int(diagnostic.Code()))
				}
				slices.Sort(diagnosticCodes)
				slices.Sort(step.Diagnostics)
				if !slices.Equal(diagnosticCodes, step.Diagnostics) {
					t.Errorf("%s: semantic diagnostics = %v, want %v", step.File, diagnosticCodes, step.Diagnostics)
				}
				if previous := programs[step.Program]; previous != nil && previous != selected.Program {
					t.Errorf("%s: project Program was not reused", step.File)
				}
				for id, previous := range programs {
					if id != step.Program && previous == selected.Program {
						t.Errorf("%s: distinct TypeScript projects shared a Program", step.File)
					}
				}
				programs[step.Program] = selected.Program
			}
			for configPath, count := range host.parsed {
				if count != 1 {
					t.Errorf("parsed %s %d times", configPath, count)
				}
			}
			for configPath, count := range host.built {
				if count != 1 {
					t.Errorf("built %s %d times", configPath, count)
				}
			}
		})
	}
}

func TestSelectionBuildsOnlyRelevantProjects(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	for _, test := range []struct {
		name   string
		target string
		want   string
	}{
		{name: "nearest-nested", target: "pkg/src/file.ts", want: "pkg/tsconfig.json"},
		{name: "solution-first-deep-versus-second-shallow", target: "target/file.ts", want: "b/custom.json"},
		{name: "no-config", target: "src/file.ts"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := tspath.NormalizePath(archive.Materialize(t, test.name))
			host := newRecordingHost(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
			selector := projectservice.New(host.host())
			selected, err := selector.Select(tspath.ResolvePath(root, test.target), root)
			if err != nil {
				t.Fatal(err)
			}
			if test.want == "" {
				if selected.Program != nil || len(host.built) != 0 {
					t.Fatalf("no-config target constructed a project: %v", host.built)
				}
				return
			}
			want := tspath.ResolvePath(root, test.want)
			if selected.ConfigPath != want || len(host.built) != 1 || host.built[want] != 1 {
				t.Fatalf("selection = %s, built = %v, want only %s", selected.ConfigPath, host.built, want)
			}
		})
	}
}

func TestSelectionPropagatesHostFailure(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	root := tspath.NormalizePath(archive.Materialize(t, "nearest-root"))
	want := errors.New("fixture host failure")
	for _, stage := range []string{"parse", "build", "nil-parse", "nil-build"} {
		t.Run(stage, func(t *testing.T) {
			host := newRecordingHost(bundled.WrapFS(osvfs.FS())).host()
			switch stage {
			case "parse":
				host.ParseConfig = func(string) (*tsoptions.ParsedCommandLine, error) { return nil, want }
			case "build":
				host.CreateProgram = func(string, *tsoptions.ParsedCommandLine) (*compiler.Program, error) { return nil, want }
			case "nil-parse":
				host.ParseConfig = func(string) (*tsoptions.ParsedCommandLine, error) {
					return nil, nil //nolint:nilnil // Model a failed config read reported without an error.
				}
			case "nil-build":
				host.CreateProgram = func(string, *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
					return nil, nil //nolint:nilnil // Model a host that cannot construct the configured Program.
				}
			}
			_, err := projectservice.New(host).Select(tspath.ResolvePath(root, "src/file.ts"), root)
			if err == nil || !strings.HasPrefix(stage, "nil-") && !errors.Is(err, want) {
				t.Fatalf("error = %v, want host failure", err)
			}
		})
	}
}

func TestSelectionKeepsLexicalConfigIdentity(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	for _, caseSensitive := range []bool{false, true} {
		name := "case-insensitive-node-modules"
		if caseSensitive {
			name = "case-sensitive-node-modules"
		}
		t.Run(name, func(t *testing.T) {
			root := tspath.NormalizePath(archive.Materialize(t, "node-modules-case-boundary"))
			fsys := &caseSensitivityFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS())), caseSensitive: caseSensitive}
			host := newRecordingHost(fsys)
			selected, err := projectservice.New(host.host()).Select(tspath.ResolvePath(root, "NODE_MODULES/pkg/file.ts"), root)
			if !caseSensitive {
				if err != nil || selected.Program != nil || len(host.built) != 0 {
					t.Fatalf("node_modules boundary was crossed: selected %s, built %v", selected.ConfigPath, host.built)
				}
			} else if err != nil || selected.ConfigPath != tspath.ResolvePath(root, "tsconfig.json") {
				t.Fatalf("uppercase directory must remain distinct: selection %s, error %v", selected.ConfigPath, err)
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
		selector := projectservice.New(host.host())
		selected, err := selector.Select(upper, root)
		if err != nil {
			t.Fatal(err)
		}
		if selected.ConfigPath != tspath.ResolvePath(root, "tsconfig.json") {
			t.Fatalf("case alias selected %s", selected.ConfigPath)
		}
	})
}

func TestSelectionReusesLoadedProjects(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	for _, scenario := range []struct {
		name, config, seed, target string
	}{
		{
			name:   "solution-disabled-reference-load-warm",
			config: "pkg/tsconfig.json",
			target: "outside/file.ts",
		},
		{
			name:   "cache-disabled-ref-parsed-only",
			config: "custom.json",
			seed:   "trigger.ts",
			target: "outside/file.ts",
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := tspath.NormalizePath(archive.Materialize(t, scenario.name))
			configPath := tspath.ResolvePath(root, scenario.config)
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			residentHost := newRecordingHost(fsys).host()
			parsed, err := residentHost.ParseConfig(configPath)
			if err != nil {
				t.Fatal(err)
			}
			resident, err := residentHost.CreateProgram(configPath, parsed)
			if err != nil {
				t.Fatal(err)
			}
			for _, state := range []string{"loaded", "missing", "failed"} {
				t.Run(state, func(t *testing.T) {
					recording := newRecordingHost(fsys)
					host := recording.host()
					wantFailure := errors.New("resident project unavailable")
					var lookups []string
					host.LoadedProgram = func(path string) (*compiler.Program, error) {
						lookups = append(lookups, path)
						if state == "failed" {
							return nil, wantFailure
						}
						if state == "missing" {
							return nil, nil //nolint:nilnil // Model the optional callback reporting a cache miss.
						}
						return resident, nil
					}
					selector := projectservice.New(host)
					if scenario.seed != "" {
						if _, err := selector.Select(tspath.ResolvePath(root, scenario.seed), root); err != nil {
							t.Fatal(err)
						}
						if len(lookups) != 0 {
							t.Fatalf("ordinary project discovery consulted loaded-only callback: %v", lookups)
						}
					}
					selected, err := selector.Select(tspath.ResolvePath(root, scenario.target), root)
					switch state {
					case "loaded":
						if err != nil || selected.Program != resident || selected.ConfigPath != configPath {
							t.Fatalf("selection = %v, error = %v, want resident %s", selected, err, configPath)
						}
					case "missing":
						if err != nil || selected.Program != nil {
							t.Fatalf("cold reference selected %s, error = %v", selected.ConfigPath, err)
						}
					case "failed":
						if !errors.Is(err, wantFailure) {
							t.Fatalf("error = %v, want resident host failure", err)
						}
					}
					if !slices.Equal(lookups, []string{configPath}) {
						t.Errorf("loaded-only lookups = %v, want %s", lookups, configPath)
					}
					if recording.built[configPath] != 0 {
						t.Errorf("disabled reference loading created %s", configPath)
					}
				})
			}
		})
	}
}
