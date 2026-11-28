package utils_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
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
	t.Run("directory-root-projection", func(t *testing.T) {
		for _, test := range []struct {
			root, directory, boundary string
		}{
			{root: "/repo/", directory: "/repo/pkg/", boundary: "/repo"},
			{root: "/", directory: "//server/share/pkg"},
		} {
			boundary, within := utils.DirectorySearchBoundary(utils.DirectoryIdentity{LexicalPath: test.directory}, utils.DirectoryIdentity{LexicalPath: test.root}, nil)
			if boundary != test.boundary || within != (test.boundary != "") {
				t.Fatalf("root=%q directory=%q: boundary=%q within=%v", test.root, test.directory, boundary, within)
			}
		}
	})
	t.Run("directory-alias-boundary", func(t *testing.T) {
		root := tspath.NormalizePath(archive.Materialize(t, "nearest-excludes-parent-includes"))
		pkg, alias := tspath.ResolvePath(root, "pkg"), tspath.ResolvePath(root, "alias")
		if err := os.Symlink(pkg, alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
		for _, paths := range [][2]string{{alias, pkg}, {pkg, alias}} {
			host := newRecordingHost(bundled.WrapFS(osvfs.FS()))
			selected, err := utils.NewTypeScriptProjectDiscovery(host.fs, host.host().ParseConfig).Find(paths[0]+"/src/file.ts", paths[1])
			if err != nil || selected != nil || host.parsed[root+"/tsconfig.json"] != 0 {
				t.Fatalf("directory alias escaped its boundary: selected=%s parsed=%v error=%v", selected.ConfigName(), host.parsed, err)
			}
		}
	})
	t.Run("nested-directory-alias-boundary", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		realDir, alias := tspath.ResolvePath(root, "real"), tspath.ResolvePath(root, "alias")
		for _, directory := range []string{realDir + "/src", realDir + "/group"} {
			if err := os.MkdirAll(directory, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(realDir+"/src/file.ts", []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(realDir+"/tsconfig.json", []byte(`{"files":["group/link/file.ts"],"compilerOptions":{"noLib":true}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, link := range [][2]string{{realDir + "/src", realDir + "/group/link"}, {realDir, alias}} {
			if err := os.Symlink(link[0], link[1]); err != nil {
				t.Skipf("directory symlink unavailable: %v", err)
			}
		}
		host := newRecordingHost(bundled.WrapFS(osvfs.FS()))
		selected, err := utils.NewTypeScriptProjectDiscovery(host.fs, host.host().ParseConfig).Find(alias+"/group/link/file.ts", realDir)
		if err != nil || selected == nil || selected.ConfigName() != alias+"/tsconfig.json" {
			t.Fatalf("nested directory alias lost its project: selected=%s parsed=%v error=%v", selected.ConfigName(), host.parsed, err)
		}
	})
	t.Run("directory-alias-into-subdirectory", func(t *testing.T) {
		root := tspath.NormalizePath(archive.Materialize(t, "nearest-excludes-parent-includes"))
		pkg, alias := tspath.ResolvePath(root, "pkg"), tspath.ResolvePath(root, "alias")
		if err := os.Symlink(pkg+"/src", alias); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}
		host := newRecordingHost(bundled.WrapFS(osvfs.FS()))
		selected, err := utils.NewTypeScriptProjectDiscovery(host.fs, host.host().ParseConfig).Find(alias+"/file.ts", pkg)
		if err != nil || selected != nil || host.parsed[root+"/tsconfig.json"] != 0 {
			t.Fatalf("directory alias escaped its boundary: selected=%s parsed=%v error=%v", selected.ConfigName(), host.parsed, err)
		}
	})
	t.Run("native-alias-boundary", func(t *testing.T) {
		root := tspath.NormalizePath(archive.Materialize(t, "nearest-excludes-parent-includes"))
		pkg, alias := root+"/pkg", root+"/PKG"
		actual, err := os.Stat(pkg)
		if err != nil {
			t.Fatal(err)
		}
		aliased, err := os.Stat(alias)
		if err != nil || !os.SameFile(actual, aliased) {
			t.Skip("native directory case aliases unavailable")
		}
		for _, caseSensitive := range []bool{false, true} {
			fsys := &caseSensitivityFS{FS: bundled.WrapFS(osvfs.FS()), caseSensitive: caseSensitive}
			host := newRecordingHost(fsys)
			selected, err := utils.NewTypeScriptProjectDiscovery(fsys, host.host().ParseConfig).Find(pkg+"/src/file.ts", alias)
			if err != nil || selected != nil || host.parsed[root+"/tsconfig.json"] != 0 {
				t.Fatalf("caseSensitive=%v: native alias escaped its boundary: selected=%s parsed=%v error=%v", caseSensitive, selected.ConfigName(), host.parsed, err)
			}
		}
	})
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

type countingDiscoveryFS struct {
	vfs.FS
	fileExists int
	realpath   int
	stat       int
}

func (fsys *countingDiscoveryFS) FileExists(path string) bool {
	fsys.fileExists++
	return fsys.FS.FileExists(path)
}

func (fsys *countingDiscoveryFS) Realpath(path string) string {
	fsys.realpath++
	return fsys.FS.Realpath(path)
}

func (fsys *countingDiscoveryFS) Stat(path string) vfs.FileInfo {
	fsys.stat++
	return fsys.FS.Stat(path)
}

func writeDiscoveryFiles(t testing.TB, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := tspath.ResolvePath(root, rel)
		if err := os.MkdirAll(tspath.GetDirectoryPath(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertSelectedConfig(t *testing.T, discovery *utils.TypeScriptProjectDiscovery, fileName, rootDir, want string) {
	t.Helper()
	selected, err := discovery.Find(fileName, rootDir)
	if err != nil {
		t.Fatalf("Find(%q, %q) unexpected error: %v", fileName, rootDir, err)
	}
	got := ""
	if selected != nil {
		got = selected.ConfigName()
	}
	if got != want {
		t.Fatalf("Find(%q, %q) = %q, want %q", fileName, rootDir, got, want)
	}
}

func TestTypeScriptProjectDiscoverySiblingAncestorCaching(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	writeDiscoveryFiles(t, root, map[string]string{
		"tsconfig.json":     `{"compilerOptions":{"noLib":true},"include":["a/b/c/d/*.ts"]}`,
		"a/b/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["c/d/f1.ts","c/d/f2.ts"]}`,
		"a/b/jsconfig.json": `{"compilerOptions":{"noLib":true},"files":["c/d/f3.ts"]}`,
		"a/b/c/d/f1.ts":     "export const v = 1;\n",
		"a/b/c/d/f2.ts":     "export const v = 2;\n",
		"a/b/c/d/f3.ts":     "export const v = 3;\n",
		"a/b/c/d/f4.ts":     "export const v = 4;\n",
		"a/b/c/d/f5.ts":     "export const v = 5;\n",
	})

	counted := &countingDiscoveryFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS()))}
	host := newRecordingHost(counted)
	discovery := utils.NewTypeScriptProjectDiscovery(counted, host.host().ParseConfig)

	abTsconfig := tspath.ResolvePath(root, "a/b/tsconfig.json")
	abJsconfig := tspath.ResolvePath(root, "a/b/jsconfig.json")
	rootTsconfig := tspath.ResolvePath(root, "tsconfig.json")

	// First file matches a/b/tsconfig.json; jsconfig.json and root tsconfig.json remain unparsed.
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f1.ts"), root, abTsconfig)
	if host.parsed[abJsconfig] != 0 || host.parsed[rootTsconfig] != 0 {
		t.Fatalf("unexpected eager config parsing: %v", host.parsed)
	}

	// Sibling file reuses cached boundary and ancestor chain with zero FS probes or parses.
	beforeExists, beforeRealpath := counted.fileExists, counted.realpath
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f2.ts"), root, abTsconfig)
	if counted.fileExists != beforeExists || counted.realpath != beforeRealpath {
		t.Fatalf("sibling lookup hit FS: fileExists %d->%d, realpath %d->%d",
			beforeExists, counted.fileExists, beforeRealpath, counted.realpath)
	}

	// File excluded by a/b/tsconfig.json lazily loads a/b/jsconfig.json; root remains unparsed.
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f3.ts"), root, abJsconfig)
	if host.parsed[rootTsconfig] != 0 {
		t.Fatalf("unexpected eager root config parsing: %v", host.parsed)
	}

	// File excluded by both a/b configs walks up to root tsconfig.json.
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f4.ts"), root, rootTsconfig)

	// Same directory with tighter rootDirectory stops at a/b boundary.
	abDir := tspath.ResolvePath(root, "a/b")
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f4.ts"), abDir, "")

	// Warm lookups for both root boundaries perform zero FS probes and zero re-parses.
	beforeExists, beforeRealpath, beforeStat := counted.fileExists, counted.realpath, counted.stat
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f5.ts"), root, rootTsconfig)
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a/b/c/d/f5.ts"), abDir, "")
	if counted.fileExists != beforeExists || counted.realpath != beforeRealpath || counted.stat != beforeStat {
		t.Fatalf("warm boundary/ancestor lookup hit FS: fileExists %d->%d, realpath %d->%d, stat %d->%d",
			beforeExists, counted.fileExists, beforeRealpath, counted.realpath, beforeStat, counted.stat)
	}
	for cfg, count := range host.parsed {
		if count != 1 {
			t.Fatalf("config %s parsed %d times, want 1", cfg, count)
		}
	}

	t.Run("symlink-directory-alias-boundary-cache", func(t *testing.T) {
		aliasRoot := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, aliasRoot, map[string]string{
			"tsconfig.json":     `{"compilerOptions":{"noLib":true},"include":["**/*.ts"]}`,
			"pkg/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["src/f1.ts","src/f2.ts"]}`,
			"pkg/src/f1.ts":     "export const v = 1;\n",
			"pkg/src/f2.ts":     "export const v = 2;\n",
			"pkg/src/f3.ts":     "export const v = 3;\n",
		})
		pkgDir := tspath.ResolvePath(aliasRoot, "pkg")
		aliasDir := tspath.ResolvePath(aliasRoot, "alias")
		if err := os.Symlink(pkgDir, aliasDir); err != nil {
			t.Skipf("directory symlink unavailable: %v", err)
		}

		aliasCounted := &countingDiscoveryFS{FS: bundled.WrapFS(osvfs.FS())}
		aliasHost := newRecordingHost(aliasCounted)
		aliasDiscovery := utils.NewTypeScriptProjectDiscovery(aliasCounted, aliasHost.host().ParseConfig)

		aliasTsconfig := tspath.ResolvePath(aliasDir, "tsconfig.json")
		assertSelectedConfig(t, aliasDiscovery, tspath.ResolvePath(aliasDir, "src/f1.ts"), pkgDir, aliasTsconfig)
		if aliasCounted.realpath == 0 || aliasCounted.stat == 0 {
			t.Fatalf("cold symlink boundary lookup did not probe FS: realpath=%d, stat=%d",
				aliasCounted.realpath, aliasCounted.stat)
		}

		// Sibling lookup through the same symlinked directory alias reuses cached boundary and dirs.
		beforeAliasExists, beforeAliasRealpath, beforeAliasStat := aliasCounted.fileExists, aliasCounted.realpath, aliasCounted.stat
		assertSelectedConfig(t, aliasDiscovery, tspath.ResolvePath(aliasDir, "src/f2.ts"), pkgDir, aliasTsconfig)
		if aliasCounted.fileExists != beforeAliasExists || aliasCounted.realpath != beforeAliasRealpath || aliasCounted.stat != beforeAliasStat {
			t.Fatalf("symlink sibling lookup hit FS: fileExists %d->%d, realpath %d->%d, stat %d->%d",
				beforeAliasExists, aliasCounted.fileExists, beforeAliasRealpath, aliasCounted.realpath, beforeAliasStat, aliasCounted.stat)
		}

		// Excluded sibling stops at the aliased root boundary without re-probing Realpath/Stat.
		assertSelectedConfig(t, aliasDiscovery, tspath.ResolvePath(aliasDir, "src/f3.ts"), pkgDir, "")
		if aliasCounted.realpath != beforeAliasRealpath || aliasCounted.stat != beforeAliasStat {
			t.Fatalf("symlink boundary check re-probed FS: realpath %d->%d, stat %d->%d",
				beforeAliasRealpath, aliasCounted.realpath, beforeAliasStat, aliasCounted.stat)
		}
		if aliasHost.parsed[tspath.ResolvePath(aliasRoot, "tsconfig.json")] != 0 {
			t.Fatalf("symlink boundary escaped to outer root: %v", aliasHost.parsed)
		}
	})
}

func TestTypeScriptProjectDiscoveryCachingPreservesErrors(t *testing.T) {
	t.Run("ancestor-parse-error-not-cached", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"pkg/tsconfig.json": `{"compilerOptions":{"noLib":true}}`,
			"pkg/a.ts":          "export const a = 1;\n",
			"pkg/b.ts":          "export const b = 2;\n",
		})
		wantErr := errors.New("parse boom")
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		calls := 0
		discovery := utils.NewTypeScriptProjectDiscovery(fsys, func(string) (*tsoptions.ParsedCommandLine, error) {
			calls++
			return nil, wantErr
		})
		for _, rel := range []string{"pkg/a.ts", "pkg/b.ts"} {
			if _, err := discovery.Find(tspath.ResolvePath(root, rel), root); !errors.Is(err, wantErr) {
				t.Fatalf("Find(%s) error = %v, want %v", rel, err, wantErr)
			}
		}
		if calls != 2 {
			t.Fatalf("parse calls = %d, want 2", calls)
		}
	})

	t.Run("reference-error-stays-lazy-and-repeatable", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["a.ts","b.ts","c.ts"],"references":[{"path":"./good.json"},{"path":"./broken.json"}]}`,
			"good.json":     `{"compilerOptions":{"noLib":true},"files":["a.ts"]}`,
			"broken.json":   `{"compilerOptions":{"noLib":true},"files":["b.ts"]}`,
			"a.ts":          "export const a = 1;\n",
			"b.ts":          "export const b = 2;\n",
			"c.ts":          "export const c = 3;\n",
		})
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		host := newRecordingHost(fsys)
		baseParse := host.host().ParseConfig
		brokenPath := tspath.ResolvePath(root, "broken.json")
		goodPath := tspath.ResolvePath(root, "good.json")
		wantErr := errors.New("broken ref")
		brokenCalls := 0
		discovery := utils.NewTypeScriptProjectDiscovery(fsys, func(name string) (*tsoptions.ParsedCommandLine, error) {
			if name == brokenPath {
				brokenCalls++
				return nil, wantErr
			}
			return baseParse(name)
		})

		// a.ts hits good.json before broken.json is ever parsed.
		assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a.ts"), root, goodPath)
		if brokenCalls != 0 {
			t.Fatalf("broken.json parsed eagerly: %d calls", brokenCalls)
		}
		// b.ts misses good.json and hits broken.json error.
		if _, err := discovery.Find(tspath.ResolvePath(root, "b.ts"), root); !errors.Is(err, wantErr) {
			t.Fatalf("Find(b.ts) error = %v, want %v", err, wantErr)
		}
		// Cached prefix still resolves a.ts without hitting broken.json.
		assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "a.ts"), root, goodPath)
		// c.ts still propagates the error from broken.json.
		if _, err := discovery.Find(tspath.ResolvePath(root, "c.ts"), root); !errors.Is(err, wantErr) {
			t.Fatalf("Find(c.ts) error = %v, want %v", err, wantErr)
		}
		if brokenCalls != 2 {
			t.Fatalf("broken.json parse calls = %d, want 2", brokenCalls)
		}
	})

	t.Run("nested-reference-error-resumes-stack", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["f1.ts","f2.ts","f3.ts"],"references":[{"path":"./b.json"}]}`,
			"b.json":        `{"compilerOptions":{"noLib":true,"disableSourceOfProjectReferenceRedirect":true},"files":["f1.ts"],"references":[{"path":"./broken.json"}]}`,
			"broken.json":   `{"compilerOptions":{"noLib":true},"files":["f2.ts"]}`,
			"f1.ts":         "export const f1 = 1;\n",
			"f2.ts":         "export const f2 = 2;\n",
			"f3.ts":         "export const f3 = 3;\n",
		})
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		host := newRecordingHost(fsys)
		baseParse := host.host().ParseConfig
		bPath := tspath.ResolvePath(root, "b.json")
		brokenPath := tspath.ResolvePath(root, "broken.json")
		wantErr := errors.New("nested broken ref")
		brokenCalls := 0
		discovery := utils.NewTypeScriptProjectDiscovery(fsys, func(name string) (*tsoptions.ParsedCommandLine, error) {
			if name == brokenPath {
				brokenCalls++
				return nil, wantErr
			}
			return baseParse(name)
		})

		// First f1.ts lookup runs uncached; second f1.ts lookup populates root's
		// redirect cache and pauses its DFS stack at depth 2 (B -> broken.json).
		assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "f1.ts"), root, bPath)
		assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "f1.ts"), root, bPath)
		if brokenCalls != 0 {
			t.Fatalf("broken.json parsed eagerly: %d calls", brokenCalls)
		}
		// f2.ts misses B and resumes the depth-2 stack frame, hitting broken.json.
		if _, err := discovery.Find(tspath.ResolvePath(root, "f2.ts"), root); !errors.Is(err, wantErr) {
			t.Fatalf("Find(f2.ts) error = %v, want %v", err, wantErr)
		}
		if brokenCalls != 1 {
			t.Fatalf("broken.json parse calls = %d, want 1", brokenCalls)
		}
		// f1.ts still resolves from visited prefix without retrying broken.json.
		assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "f1.ts"), root, bPath)
		if brokenCalls != 1 {
			t.Fatalf("f1.ts retry unexpectedly hit broken.json: %d calls", brokenCalls)
		}
		// f3.ts resumes the same depth-2 frame and propagates broken.json's error again.
		if _, err := discovery.Find(tspath.ResolvePath(root, "f3.ts"), root); !errors.Is(err, wantErr) {
			t.Fatalf("Find(f3.ts) error = %v, want %v", err, wantErr)
		}
		if brokenCalls != 2 {
			t.Fatalf("broken.json parse calls = %d, want 2", brokenCalls)
		}
	})
}

func TestTypeScriptProjectDiscoveryCyclicReferences(t *testing.T) {
	t.Run("two-node-cycle", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"tsconfig.json":       `{"compilerOptions":{"noLib":true},"files":["only-a.ts","both.ts"],"references":[{"path":"./b/tsconfig.app.json"}]}`,
			"b/tsconfig.app.json": `{"compilerOptions":{"noLib":true},"files":["../only-b.ts","../both.ts"],"references":[{"path":"../tsconfig.json"}]}`,
			"only-a.ts":           "export const a = 1;\n",
			"only-b.ts":           "export const b = 2;\n",
			"both.ts":             "export const both = 3;\n",
		})
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		host := newRecordingHost(fsys)
		discovery := utils.NewTypeScriptProjectDiscovery(fsys, host.host().ParseConfig)

		aCfg := tspath.ResolvePath(root, "tsconfig.json")
		bCfg := tspath.ResolvePath(root, "b/tsconfig.app.json")
		for range 3 {
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "only-b.ts"), root, bCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "both.ts"), root, aCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "only-a.ts"), root, aCfg)
		}
	})

	t.Run("three-node-and-self-cycle", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"tsconfig.json":      `{"compilerOptions":{"noLib":true},"files":["only-a.ts","in-ac.ts"],"references":[{"path":"./b.json"}]}`,
			"b.json":             `{"compilerOptions":{"noLib":true},"files":[],"references":[{"path":"./c.json"}]}`,
			"c.json":             `{"compilerOptions":{"noLib":true},"files":["only-c.ts","in-ac.ts"],"references":[{"path":"./tsconfig.json"}]}`,
			"self/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["self.ts"],"references":[{"path":"./tsconfig.json"}]}`,
			"only-a.ts":          "export const a = 1;\n",
			"only-c.ts":          "export const c = 2;\n",
			"in-ac.ts":           "export const ac = 3;\n",
			"self/self.ts":       "export const self = 4;\n",
		})
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		host := newRecordingHost(fsys)
		discovery := utils.NewTypeScriptProjectDiscovery(fsys, host.host().ParseConfig)

		aCfg := tspath.ResolvePath(root, "tsconfig.json")
		cCfg := tspath.ResolvePath(root, "c.json")
		selfCfg := tspath.ResolvePath(root, "self/tsconfig.json")
		for range 3 {
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "only-a.ts"), root, aCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "only-c.ts"), root, cCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "in-ac.ts"), root, aCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "self/self.ts"), root, selfCfg)
		}
	})

	t.Run("non-lexical-config-name-cycle", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"tsconfig.json":      `{"compilerOptions":{"noLib":true},"files":["pkg/only-a.ts"]}`,
			"pkg/tsconfig.json":  `{}`,
			"real/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["../pkg/only-a.ts"],"references":[{"path":"../pkg/b.json"}]}`,
			"pkg/b.json":         `{"compilerOptions":{"noLib":true},"files":["only-b.ts"],"references":[{"path":"./tsconfig.json"}]}`,
			"pkg/only-a.ts":      "export const a = 1;\n",
			"pkg/only-b.ts":      "export const b = 2;\n",
			"pkg/unowned.ts":     "export const u = 3;\n",
		})
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		host := newRecordingHost(fsys)
		baseParse := host.host().ParseConfig
		lexicalPkgCfg := tspath.ResolvePath(root, "pkg/tsconfig.json")
		realPkgCfg := tspath.ResolvePath(root, "real/tsconfig.json")
		bCfg := tspath.ResolvePath(root, "pkg/b.json")
		discovery := utils.NewTypeScriptProjectDiscovery(fsys, func(name string) (*tsoptions.ParsedCommandLine, error) {
			if name == lexicalPkgCfg {
				return baseParse(realPkgCfg)
			}
			return baseParse(name)
		})

		// Back-reference from b.json to pkg/tsconfig.json (lookupPath) must not redirect
		// pkg/only-a.ts away from real/tsconfig.json (configPath) to the outer root tsconfig.json.
		for range 2 {
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/only-a.ts"), root, realPkgCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/only-b.ts"), root, bCfg)
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/unowned.ts"), root, "")
		}
	})
}

func TestTypeScriptProjectDiscoveryCaseInsensitiveCache(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	writeDiscoveryFiles(t, root, map[string]string{
		"tsconfig.json":     `{"compilerOptions":{"noLib":true},"files":["pkg/src/a.ts","pkg/src/b.ts"]}`,
		"pkg/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["src/a.ts","src/b.ts"],"references":[{"path":"../ref/tsconfig.json"}]}`,
		"pkg/src/a.ts":      "export const a = 1;\n",
		"pkg/src/b.ts":      "export const b = 2;\n",
		"ref/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["other.ts"],"references":[{"path":"../pkg/TSCONFIG.json"}]}`,
		"ref/other.ts":      "export const o = 3;\n",
	})

	fsys := &caseSensitivityFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS())), caseSensitive: false}
	host := newRecordingHost(fsys)
	discovery := utils.NewTypeScriptProjectDiscovery(fsys, host.host().ParseConfig)

	rootCfg := tspath.ResolvePath(root, "tsconfig.json")
	pkgCfg := tspath.ResolvePath(root, "pkg/tsconfig.json")
	refCfg := tspath.ResolvePath(root, "ref/tsconfig.json")
	parsedRef, err := host.host().ParseConfig(refCfg)
	if err != nil {
		t.Fatal(err)
	}
	wantRefTarget := tspath.ResolvePath(root, "pkg/TSCONFIG.json")
	if got := parsedRef.ResolvedProjectReferencePaths(); len(got) != 1 || got[0] != wantRefTarget {
		t.Fatalf("ref resolved references = %v, want [%s]", got, wantRefTarget)
	}
	if !strings.EqualFold(wantRefTarget, pkgCfg) {
		t.Fatalf("expected case-insensitive match between %q and %q", wantRefTarget, pkgCfg)
	}
	host.parsed[refCfg] = 0

	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/src/a.ts"), root, pkgCfg)
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/src/B.TS"), root, pkgCfg)
	// Different lexical directory casing reuses the same parsed config via case-insensitive path key.
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "PKG/src/a.ts"), root, pkgCfg)
	// Querying ref/other.ts twice exercises both uncached and cached redirect paths through ../pkg/TSCONFIG.json.
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "ref/other.ts"), root, refCfg)
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "ref/other.ts"), root, refCfg)
	// Repeated queries confirm mixed-case cyclic back-reference ../pkg/TSCONFIG.json did not redirect
	// pkg/tsconfig.json away to the outer root tsconfig.json.
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/src/a.ts"), root, pkgCfg)
	assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/src/b.ts"), root, pkgCfg)

	if host.parsed[rootCfg] != 0 {
		t.Fatalf("false self-redirect escaped to outer root tsconfig.json (%d parses)", host.parsed[rootCfg])
	}
	if host.parsed[pkgCfg] != 1 {
		t.Fatalf("pkg/tsconfig.json parsed %d times, want 1", host.parsed[pkgCfg])
	}
	if host.parsed[refCfg] != 1 {
		t.Fatalf("ref/tsconfig.json parsed %d times, want 1", host.parsed[refCfg])
	}
}

func TestTypeScriptProjectDiscoveryLazyRefsAndBoundedRedirects(t *testing.T) {
	t.Run("single-file-lookup-initializes-refs-lazily", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		const numRefs = 512
		refs := make([]string, numRefs)
		files := make(map[string]string, numRefs+3)
		for i := range numRefs {
			// Use uppercase characters with caseSensitive=false so eager d.path(name)
			// would perform >= 512 heap allocations on every OS (including Linux).
			refs[i] = fmt.Sprintf(`{"path":"./P%03d/tsconfig.app.json"}`, i)
		}
		files["tsconfig.json"] = fmt.Sprintf(`{"files":[],"references":[%s]}`, strings.Join(refs, ","))
		files["P000/tsconfig.app.json"] = `{"compilerOptions":{"noLib":true},"files":["index.ts"]}`
		files["P000/index.ts"] = "export const v = 1;\n"
		writeDiscoveryFiles(t, root, files)

		fsys := &caseSensitivityFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS())), caseSensitive: false}
		baseHost := newRecordingHost(fsys).host()
		parsedConfigs := make(map[string]*tsoptions.ParsedCommandLine, 2)
		for _, rel := range []string{"tsconfig.json", "P000/tsconfig.app.json"} {
			cfgPath := tspath.ResolvePath(root, rel)
			parsed, err := baseHost.ParseConfig(cfgPath)
			if err != nil {
				t.Fatal(err)
			}
			// Pre-populate cached ResolvedProjectReferencePaths on ParsedCommandLine so
			// allocations below measure only discovery's own per-ref work.
			_ = parsed.ResolvedProjectReferencePaths()
			parsedConfigs[cfgPath] = parsed
		}
		parseCached := func(name string) (*tsoptions.ParsedCommandLine, error) {
			return parsedConfigs[name], nil
		}

		target := tspath.ResolvePath(root, "P000/index.ts")
		wantCfg := tspath.ResolvePath(root, "P000/tsconfig.app.json")
		// Warm cachedvfs so FileExists/Realpath make no first-touch allocations.
		assertSelectedConfig(t, utils.NewTypeScriptProjectDiscovery(fsys, parseCached), target, root, wantCfg)

		// Eagerly converting all 512 references would perform >512 allocations;
		// lazy refAt initializes only ref 0.
		allocs := testing.AllocsPerRun(20, func() {
			discovery := utils.NewTypeScriptProjectDiscovery(fsys, parseCached)
			selected, err := discovery.Find(target, root)
			if err != nil || selected == nil || selected.ConfigName() != wantCfg {
				t.Fatalf("Find = (%v, %v), want %q", selected, err, wantCfg)
			}
		})
		if allocs > 100 {
			t.Fatalf("first-of-512 single-file lookup performed %.0f allocs, want <= 100 (refs were eagerly initialized)", allocs)
		}
	})

	t.Run("redirect-budget-overflow-disables-cache-without-thrashing-and-preserves-redirects-and-errors", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		// Filler chain f000..f099: querying f000..f057 twice populates
		// sum(99..42) = 58 * 141 / 2 = 4089 cached redirect entries, leaving
		// exactly 7 slots below maxDiscoveryRedirectEntries (4096).
		const fillerLen = 100
		const fillerQueryCount = 58
		const ovChainLen = 10
		files := make(map[string]string, fillerLen*2+ovChainLen+8)
		for i := range fillerLen {
			dir := fmt.Sprintf("f%03d", i)
			if i+1 < fillerLen {
				files[dir+"/tsconfig.json"] = fmt.Sprintf(
					`{"compilerOptions":{"noLib":true},"files":["index.ts"],"references":[{"path":"../f%03d/tsconfig.json"}]}`,
					i+1,
				)
			} else {
				files[dir+"/tsconfig.json"] = `{"compilerOptions":{"noLib":true},"files":["index.ts"]}`
			}
			files[dir+"/index.ts"] = fmt.Sprintf("export const f = %d;\n", i)
		}
		files["ov/tsconfig.json"] = `{"compilerOptions":{"noLib":true},"files":["own.ts","hit-at-cap.ts","hit-after-cap.ts","err-after-cap.ts"],"references":[{"path":"../chain0.json"}]}`
		files["ov/own.ts"] = "export const own = 1;\n"
		files["ov/hit-at-cap.ts"] = "export const hitAtCap = 2;\n"
		files["ov/hit-after-cap.ts"] = "export const hitAfterCap = 3;\n"
		files["ov/err-after-cap.ts"] = "export const errAfterCap = 4;\n"
		for i := range ovChainLen {
			var fileList, refList, extraOpts string
			switch i {
			case 7:
				// 8th node (after 7 slots chain0..chain6 fill 4089->4096): hits exact budget cap!
				// disableSourceOfProjectReferenceRedirect lets chain7 claim hit-at-cap.ts directly
				// without chasing chain8..tail-broken.json during Subtest C's priming lookup.
				fileList = `"ov/hit-at-cap.ts"`
				extraOpts = `,"disableSourceOfProjectReferenceRedirect":true`
			case ovChainLen - 1:
				fileList = `"ov/hit-after-cap.ts"`
			}
			if i+1 < ovChainLen {
				refList = fmt.Sprintf(`{"path":"./chain%d.json"}`, i+1)
			} else {
				refList = `{"path":"./tail-broken.json"}`
			}
			files[fmt.Sprintf("chain%d.json", i)] = fmt.Sprintf(
				`{"compilerOptions":{"noLib":true%s},"files":[%s],"references":[%s]}`,
				extraOpts, fileList, refList,
			)
		}
		files["tail-broken.json"] = `{"compilerOptions":{"noLib":true},"files":["tail.ts"]}`
		files["tail.ts"] = "export const tail = 5;\n"
		writeDiscoveryFiles(t, root, files)

		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		baseHost := newRecordingHost(fsys).host()
		parsedConfigs := make(map[string]*tsoptions.ParsedCommandLine, fillerLen+ovChainLen+2)
		preParse := func(rel string) string {
			cfgPath := tspath.ResolvePath(root, rel)
			parsed, err := baseHost.ParseConfig(cfgPath)
			if err != nil {
				t.Fatal(err)
			}
			_ = parsed.ResolvedProjectReferencePaths()
			parsedConfigs[cfgPath] = parsed
			return cfgPath
		}
		for i := range fillerLen {
			preParse(fmt.Sprintf("f%03d/tsconfig.json", i))
		}
		ovCfg := preParse("ov/tsconfig.json")
		chain7Cfg := preParse("chain7.json")
		chain9Cfg := preParse("chain9.json")
		for i := range ovChainLen {
			if i != 7 && i != ovChainLen-1 {
				preParse(fmt.Sprintf("chain%d.json", i))
			}
		}
		brokenPath := preParse("tail-broken.json")
		wantErr := errors.New("tail broken ref")

		newFilledDiscovery := func(failBroken bool) *utils.TypeScriptProjectDiscovery {
			d := utils.NewTypeScriptProjectDiscovery(fsys, func(name string) (*tsoptions.ParsedCommandLine, error) {
				if name == brokenPath && failBroken {
					return nil, wantErr
				}
				return parsedConfigs[name], nil
			})
			// Two passes over f000..f057 fill exactly 4089 slots in d.redirectEntries.
			for range 2 {
				for i := range fillerQueryCount {
					dir := fmt.Sprintf("f%03d", i)
					assertSelectedConfig(t, d, tspath.ResolvePath(root, dir+"/index.ts"), root, tspath.ResolvePath(root, dir+"/tsconfig.json"))
				}
			}
			return d
		}

		// Subtest A: redirect hit on the exact child that hits the 4096 budget cap (chain7),
		// followed by verifying zero thrashing on subsequent repeat queries to ov.
		dA := newFilledDiscovery(false)
		assertSelectedConfig(t, dA, tspath.ResolvePath(root, "ov/own.ts"), root, ovCfg)
		assertSelectedConfig(t, dA, tspath.ResolvePath(root, "ov/hit-at-cap.ts"), root, chain7Cfg)
		if gotEntries := mustStructField(t, reflect.ValueOf(dA), "redirectEntries").Int(); gotEntries != 4089 {
			t.Fatalf("after overflow rollback: dA.redirectEntries = %d, want 4089 (reclaimed 7 ov entries)", gotEntries)
		}
		ovOwn := tspath.ResolvePath(root, "ov/own.ts")
		// Warm once so refAt slices on chain0..chain9 are already populated before measuring repeat allocs.
		assertSelectedConfig(t, dA, ovOwn, root, ovCfg)
		allocs := testing.AllocsPerRun(20, func() {
			selected, err := dA.Find(ovOwn, root)
			if err != nil || selected == nil || selected.ConfigName() != ovCfg {
				t.Fatalf("overflowed Find = (%v, %v), want %q", selected, err, ovCfg)
			}
		})
		if allocs > 8 {
			t.Fatalf("repeat query on overflowed config performed %.0f allocs, want <= 8 (redirect cache thrashing)", allocs)
		}

		// Subtest B: redirect hit discovered after cap overflow inside drainRedirectSearch (chain9).
		dB := newFilledDiscovery(false)
		assertSelectedConfig(t, dB, tspath.ResolvePath(root, "ov/own.ts"), root, ovCfg)
		assertSelectedConfig(t, dB, tspath.ResolvePath(root, "ov/hit-after-cap.ts"), root, chain9Cfg)
		assertSelectedConfig(t, dB, tspath.ResolvePath(root, "ov/hit-after-cap.ts"), root, chain9Cfg)

		// Subtest C: broken ref error encountered during drainRedirectSearch after cap overflow,
		// and repeated cleanly on subsequent queries once redirectDisabled is set.
		brokenCalls := 0
		dC := utils.NewTypeScriptProjectDiscovery(fsys, func(name string) (*tsoptions.ParsedCommandLine, error) {
			if name == brokenPath {
				brokenCalls++
				return nil, wantErr
			}
			return parsedConfigs[name], nil
		})
		for range 2 {
			for i := range fillerQueryCount {
				dir := fmt.Sprintf("f%03d", i)
				assertSelectedConfig(t, dC, tspath.ResolvePath(root, dir+"/index.ts"), root, tspath.ResolvePath(root, dir+"/tsconfig.json"))
			}
		}
		// Prime ov.redirectQueried via hit-at-cap.ts (stops at chain7 before tail-broken.json).
		assertSelectedConfig(t, dC, tspath.ResolvePath(root, "ov/hit-at-cap.ts"), root, chain7Cfg)
		if brokenCalls != 0 {
			t.Fatalf("tail-broken.json called %d times during priming, want 0", brokenCalls)
		}
		for attempt := range 2 {
			if _, err := dC.Find(tspath.ResolvePath(root, "ov/err-after-cap.ts"), root); !errors.Is(err, wantErr) {
				t.Fatalf("attempt %d: Find(err-after-cap.ts) error = %v, want %v", attempt, err, wantErr)
			}
			if brokenCalls != attempt+1 {
				t.Fatalf("attempt %d: tail-broken.json called %d times, want %d", attempt, brokenCalls, attempt+1)
			}
		}
	})

	t.Run("intermediate-config-with-both-disable-flags-still-traverses-transitive-redirects", func(t *testing.T) {
		root := tspath.NormalizePath(t.TempDir())
		writeDiscoveryFiles(t, root, map[string]string{
			"tsconfig.json":     `{"compilerOptions":{"noLib":true},"files":["pkg/x.ts"]}`,
			"pkg/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["x.ts","err.ts"],"references":[{"path":"./mid.json"}]}`,
			"pkg/mid.json":      `{"compilerOptions":{"noLib":true,"disableSourceOfProjectReferenceRedirect":true,"disableReferencedProjectLoad":true},"files":[],"references":[{"path":"./leaf.json"},{"path":"./broken.json"}]}`,
			"pkg/leaf.json":     `{"compilerOptions":{"noLib":true},"files":["x.ts"]}`,
			"pkg/broken.json":   `{"compilerOptions":{"noLib":true},"files":["other.ts"]}`,
			"pkg/x.ts":          "export const x = 1;\n",
			"pkg/err.ts":        "export const e = 2;\n",
		})
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		baseHost := newRecordingHost(fsys).host()
		rootCfg := tspath.ResolvePath(root, "tsconfig.json")
		brokenCfg := tspath.ResolvePath(root, "pkg/broken.json")
		wantErr := errors.New("transitive broken ref")

		discovery := utils.NewTypeScriptProjectDiscovery(fsys, func(name string) (*tsoptions.ParsedCommandLine, error) {
			if name == brokenCfg {
				return nil, wantErr
			}
			return baseHost.ParseConfig(name)
		})
		// pkg/tsconfig.json -> pkg/mid.json (both flags true) -> pkg/leaf.json owns x.ts,
		// so pkg/tsconfig.json must redirect and walk up to outer root tsconfig.json.
		for range 2 {
			assertSelectedConfig(t, discovery, tspath.ResolvePath(root, "pkg/x.ts"), root, rootCfg)
		}
		// Missing in leaf.json must continue past mid.json into broken.json and propagate its error.
		if _, err := discovery.Find(tspath.ResolvePath(root, "pkg/err.ts"), root); !errors.Is(err, wantErr) {
			t.Fatalf("Find(pkg/err.ts) error = %v, want %v", err, wantErr)
		}
	})

	t.Run("single-file-late-of-512-refs-avoids-repeated-copies", func(t *testing.T) {
		const root = "/Workspace"
		const numRefs = 512
		const maxChunkSize = 32
		refs := make([]string, numRefs)
		files := map[string]string{
			root + "/first.ts":  "export const x = 1;\n",
			root + "/ref001.ts": "export const x = 1;\n",
			root + "/ref032.ts": "export const x = 1;\n",
			root + "/ref033.ts": "export const x = 1;\n",
			root + "/last.ts":   "export const x = 1;\n",
			root + "/other.ts":  "export const x = 1;\n",
		}
		for i := range numRefs {
			name := fmt.Sprintf("Project%03d.json", i)
			refs[i] = fmt.Sprintf(`{"path":"./%s"}`, name)
			target := "other.ts"
			switch i {
			case 0:
				target = "first.ts"
			case 1:
				target = "ref001.ts"
			case 32:
				target = "ref032.ts"
			case 33:
				target = "ref033.ts"
			case numRefs - 1:
				target = "last.ts"
			}
			files[root+"/"+name] = fmt.Sprintf(`{"compilerOptions":{"composite":true},"files":[%q]}`, target)
		}
		files[root+"/tsconfig.json"] = fmt.Sprintf(`{"files":[],"references":[%s]}`, strings.Join(refs, ","))

		for _, caseSensitive := range []bool{true, false} {
			t.Run(fmt.Sprintf("caseSensitive=%t", caseSensitive), func(t *testing.T) {
				fsys, parse := inMemoryDiscoveryHost(t, files, caseSensitive)
				firstFile := root + "/first.ts"
				firstCfg := root + "/Project000.json"
				lastFile := root + "/last.ts"
				lastCfg := root + "/Project511.json"

				// Verify the non-copying fixed-size chunk invariant directly as refAt grows
				// across boundaries: 0 -> 1 -> 32 -> 33 -> 511.
				d := utils.NewTypeScriptProjectDiscovery(fsys, parse)
				rootCfgKey := tspath.ToPath(root+"/tsconfig.json", "", caseSensitive)
				inspectRootRefs := func() (refCount int, firstRefAddr uintptr, refChunks reflect.Value) {
					t.Helper()
					dv := reflect.ValueOf(d).Elem()
					configs := mustStructField(t, dv, "configs")
					cfg := configs.MapIndex(reflect.ValueOf(rootCfgKey))
					if !cfg.IsValid() || cfg.IsNil() {
						t.Fatalf("missing root config entry %q in discovery.configs", rootCfgKey)
					}
					refState := mustStructField(t, cfg, "refs")
					if refState.IsNil() {
						t.Fatalf("root config entry %q has nil refs", rootCfgKey)
					}
					refCount = int(mustStructField(t, refState, "refCount").Int())
					firstRefAddr = mustStructField(t, refState, "firstRef").Addr().Pointer()
					refChunks = mustStructField(t, refState, "refChunks")
					return refCount, firstRefAddr, refChunks
				}

				var firstRefAddr, chunk0Addr, chunk1Addr uintptr
				for _, step := range []struct {
					index      int
					file       string
					wantCfg    string
					wantCount  int
					wantChunks int
				}{
					{index: 0, file: firstFile, wantCfg: firstCfg, wantCount: 1, wantChunks: 0},
					{index: 1, file: root + "/ref001.ts", wantCfg: root + "/Project001.json", wantCount: 2, wantChunks: 1},
					{index: 32, file: root + "/ref032.ts", wantCfg: root + "/Project032.json", wantCount: 33, wantChunks: 1},
					{index: 33, file: root + "/ref033.ts", wantCfg: root + "/Project033.json", wantCount: 34, wantChunks: 2},
					{index: numRefs - 1, file: lastFile, wantCfg: lastCfg, wantCount: numRefs, wantChunks: 16},
				} {
					assertSelectedConfig(t, d, step.file, root, step.wantCfg)
					refCount, gotFirstAddr, refChunks := inspectRootRefs()
					if refCount != step.wantCount {
						t.Fatalf("after index %d: refCount = %d, want %d", step.index, refCount, step.wantCount)
					}
					if refChunks.Len() != step.wantChunks {
						t.Fatalf("after index %d: len(refChunks) = %d, want %d", step.index, refChunks.Len(), step.wantChunks)
					}
					if gotFirstAddr == 0 {
						t.Fatalf("after index %d: firstRef address is 0", step.index)
					}
					if firstRefAddr == 0 {
						firstRefAddr = gotFirstAddr
					} else if gotFirstAddr != firstRefAddr {
						t.Fatalf("after index %d: firstRef moved from %#x to %#x", step.index, firstRefAddr, gotFirstAddr)
					}
					for ci := range refChunks.Len() {
						chunk := refChunks.Index(ci)
						if chunk.Len() == 0 || chunk.Len() > maxChunkSize || chunk.Cap() > maxChunkSize {
							t.Fatalf("after index %d: refChunks[%d] has len=%d cap=%d, want 1 <= len <= cap <= %d",
								step.index, ci, chunk.Len(), chunk.Cap(), maxChunkSize)
						}
					}
					if refChunks.Len() >= 1 {
						gotChunk0Addr := refChunks.Index(0).Index(0).Addr().Pointer()
						if gotChunk0Addr == 0 {
							t.Fatalf("after index %d: refChunks[0][0] address is 0", step.index)
						}
						if chunk0Addr == 0 {
							chunk0Addr = gotChunk0Addr
						} else if gotChunk0Addr != chunk0Addr {
							t.Fatalf("after index %d: refChunks[0][0] moved from %#x to %#x", step.index, chunk0Addr, gotChunk0Addr)
						}
					}
					if refChunks.Len() >= 2 {
						gotChunk1Addr := refChunks.Index(1).Index(0).Addr().Pointer()
						if gotChunk1Addr == 0 {
							t.Fatalf("after index %d: refChunks[1][0] address is 0", step.index)
						}
						if chunk1Addr == 0 {
							chunk1Addr = gotChunk1Addr
						} else if gotChunk1Addr != chunk1Addr {
							t.Fatalf("after index %d: refChunks[1][0] moved from %#x to %#x", step.index, chunk1Addr, gotChunk1Addr)
						}
					}
					if step.index == numRefs-1 {
						totalChunkSlots := 0
						for ci := range refChunks.Len() {
							wantLen := maxChunkSize
							if ci == refChunks.Len()-1 {
								wantLen = (numRefs - 1) % maxChunkSize
								if wantLen == 0 {
									wantLen = maxChunkSize
								}
							}
							chunk := refChunks.Index(ci)
							if chunk.Len() != wantLen {
								t.Fatalf("final refChunks[%d] len = %d, want %d", ci, chunk.Len(), wantLen)
							}
							totalChunkSlots += chunk.Len()
						}
						if 1+totalChunkSlots != numRefs {
							t.Fatalf("1 + totalChunkSlots = %d, want %d", 1+totalChunkSlots, numRefs)
						}
					}
				}

				firstAllocs := testing.AllocsPerRun(20, func() {
					fresh := utils.NewTypeScriptProjectDiscovery(fsys, parse)
					selected, err := fresh.Find(firstFile, root)
					if err != nil || selected == nil || selected.ConfigName() != firstCfg {
						t.Fatalf("Find(first) = (%v, %v), want %q", selected, err, firstCfg)
					}
				})
				if firstAllocs > 80 {
					t.Fatalf("first-of-512 single-file lookup performed %.0f allocs, want <= 80", firstAllocs)
				}
			})
		}
	})

	t.Run("shared-missing-and-aliased-refs-stay-within-redirect-budget", func(t *testing.T) {
		type retainedRedirectStats struct {
			retainedRedirects int
			visitedEntries    int
			seenKeys          int
			stackFrames       int
			missingSeenKeys   int
			aliasSeenKeys     int
			globalMissingKeys int
		}
		const redirectBudget = 4096
		inspectRetainedMaps := func(t *testing.T, d *utils.TypeScriptProjectDiscovery) retainedRedirectStats {
			t.Helper()
			var stats retainedRedirectStats
			dv := reflect.ValueOf(d).Elem()
			for _, mapName := range []string{"seenGen", "redirectSeenGen"} {
				m := mustStructField(t, dv, mapName)
				if !m.IsNil() {
					iter := m.MapRange()
					for iter.Next() {
						if strings.Contains(iter.Key().String(), "/absent-") {
							stats.globalMissingKeys++
						}
					}
				}
			}
			configs := mustStructField(t, dv, "configs")
			entries := configs.MapRange()
			for entries.Next() {
				cfg := entries.Value()
				if cfg.IsNil() {
					continue
				}
				refState := mustStructField(t, cfg, "refs")
				if refState.IsNil() {
					continue
				}
				cache := mustStructField(t, refState, "redirect")
				if cache.IsNil() {
					continue
				}
				stats.retainedRedirects++
				visited := mustStructField(t, cache, "visited").Len()
				seen := mustStructField(t, cache, "seen")
				stack := mustStructField(t, cache, "stack").Len()
				seenLen := 0
				if !seen.IsNil() {
					seenLen = seen.Len()
				}
				if visited < 1 {
					t.Fatalf("retained redirectSearch for %q has visited=%d, want >= 1", entries.Key().String(), visited)
				}
				if seenLen > 2+2*visited {
					t.Fatalf("retained redirectSearch for %q has seen=%d > 2+2*visited (%d)", entries.Key().String(), seenLen, 2+2*visited)
				}
				if stack > 1+visited {
					t.Fatalf("retained redirectSearch for %q has stack=%d > 1+visited (%d)", entries.Key().String(), stack, 1+visited)
				}
				stats.visitedEntries += visited
				stats.seenKeys += seenLen
				stats.stackFrames += stack
				if !seen.IsNil() {
					keys := seen.MapRange()
					for keys.Next() {
						k := keys.Key().String()
						if strings.Contains(k, "/absent-") {
							stats.missingSeenKeys++
						}
						if strings.Contains(k, "/alias-") {
							stats.aliasSeenKeys++
						}
					}
				}
			}
			redirectEntries := int(mustStructField(t, dv, "redirectEntries").Int())
			if redirectEntries != stats.visitedEntries {
				t.Fatalf("discovery.redirectEntries = %d, want sum(visited) = %d", redirectEntries, stats.visitedEntries)
			}
			if stats.visitedEntries > redirectBudget {
				t.Fatalf("total retained redirect visited entries = %d, want <= %d", stats.visitedEntries, redirectBudget)
			}
			return stats
		}

		for _, withTrailingRef := range []bool{false, true} {
			name := "exhausted-after-match"
			if withTrailingRef {
				name = "paused-before-trailing-ref"
			}
			t.Run(name, func(t *testing.T) {
				const root = "/workspace"
				const parents = 128
				const missing = 512
				files := map[string]string{}
				var targetFiles, ownerFiles, sharedRefs []string
				for i := range parents {
					dir := fmt.Sprintf("p%03d", i)
					files[root+"/"+dir+"/tsconfig.json"] = `{"files":["one.ts","two.ts"],"references":[{"path":"../shared.json"}]}`
					for _, leaf := range []string{"one.ts", "two.ts"} {
						rel := dir + "/" + leaf
						files[root+"/"+rel] = "export const x = 1;\n"
						ownerFiles = append(ownerFiles, fmt.Sprintf("%q", rel))
						targetFiles = append(targetFiles, root+"/"+rel)
					}
				}
				for i := range missing {
					sharedRefs = append(sharedRefs, fmt.Sprintf(`{"path":"./absent-%03d.json"}`, i))
				}
				sharedRefs = append(sharedRefs, `{"path":"./owner.json"}`)
				if withTrailingRef {
					sharedRefs = append(sharedRefs, `{"path":"./trailing.json"}`)
					files[root+"/trailing.json"] = `{"compilerOptions":{"composite":true},"files":[]}`
				}
				files[root+"/owner.json"] = fmt.Sprintf(`{"compilerOptions":{"composite":true},"files":[%s]}`, strings.Join(ownerFiles, ","))
				files[root+"/shared.json"] = fmt.Sprintf(`{"files":[],"references":[%s]}`, strings.Join(sharedRefs, ","))

				fsys, baseParse := inMemoryDiscoveryHost(t, files, true)
				trailingParsed := 0
				parse := func(path string) (*tsoptions.ParsedCommandLine, error) {
					if path == root+"/trailing.json" {
						trailingParsed++
					}
					return baseParse(path)
				}

				d := utils.NewTypeScriptProjectDiscovery(fsys, parse)
				for _, f := range targetFiles {
					assertSelectedConfig(t, d, f, root, root+"/owner.json")
				}

				stats := inspectRetainedMaps(t, d)
				t.Logf("withTrailingRef=%t redirects=%d visited=%d seen_keys=%d stack_frames=%d missing_seen_keys=%d alias_seen_keys=%d global_missing_keys=%d",
					withTrailingRef, stats.retainedRedirects, stats.visitedEntries, stats.seenKeys, stats.stackFrames,
					stats.missingSeenKeys, stats.aliasSeenKeys, stats.globalMissingKeys)

				if trailingParsed != 0 {
					t.Fatalf("trailing.json parsed eagerly %d times, want 0", trailingParsed)
				}
				if stats.retainedRedirects != parents {
					t.Fatalf("retained redirect caches = %d, want %d", stats.retainedRedirects, parents)
				}
				if stats.visitedEntries != parents*2 {
					t.Fatalf("cached redirect visited entries = %d, want %d", stats.visitedEntries, parents*2)
				}
				if stats.missingSeenKeys != 0 {
					t.Fatalf("retained %d missing-path entries in per-project redirect seen maps, want 0", stats.missingSeenKeys)
				}
				if stats.globalMissingKeys != 0 {
					t.Fatalf("retained %d missing-path entries in discovery seenGen maps, want 0", stats.globalMissingKeys)
				}
				if stats.aliasSeenKeys != 0 {
					t.Fatalf("retained %d alias-path entries in redirect seen maps, want 0", stats.aliasSeenKeys)
				}
				if withTrailingRef {
					if stats.seenKeys != parents*3 {
						t.Fatalf("retained %d seen keys across %d paused redirect maps, want %d", stats.seenKeys, parents, parents*3)
					}
					if stats.stackFrames != parents*2 {
						t.Fatalf("retained %d stack frames across %d paused redirect searches, want %d", stats.stackFrames, parents, parents*2)
					}
				} else if stats.seenKeys != 0 || stats.stackFrames != 0 {
					t.Fatalf("retained seen=%d stack=%d after exhausted redirect searches, want 0 and 0", stats.seenKeys, stats.stackFrames)
				}
			})
		}

		t.Run("duplicate-alias-refs-do-not-grow-redirect-seen-maps", func(t *testing.T) {
			const root = "/workspace"
			const parents = 32
			const aliases = 128
			files := map[string]string{
				root + "/canonical.json": `{"compilerOptions":{"composite":true},"files":[]}`,
				root + "/trailing.json":  `{"compilerOptions":{"composite":true},"files":[]}`,
			}
			var targetFiles, ownerFiles, sharedRefs []string
			for i := range parents {
				dir := fmt.Sprintf("p%03d", i)
				files[root+"/"+dir+"/tsconfig.json"] = `{"files":["one.ts","two.ts"],"references":[{"path":"../shared.json"}]}`
				for _, leaf := range []string{"one.ts", "two.ts"} {
					rel := dir + "/" + leaf
					files[root+"/"+rel] = "export const x = 1;\n"
					ownerFiles = append(ownerFiles, fmt.Sprintf("%q", rel))
					targetFiles = append(targetFiles, root+"/"+rel)
				}
			}
			sharedRefs = append(sharedRefs, `{"path":"./canonical.json"}`)
			for i := range aliases {
				aliasRel := fmt.Sprintf("alias-%03d.json", i)
				files[root+"/"+aliasRel] = `{}`
				sharedRefs = append(sharedRefs, fmt.Sprintf(`{"path":"./%s"}`, aliasRel))
			}
			sharedRefs = append(sharedRefs, `{"path":"./owner.json"}`, `{"path":"./trailing.json"}`)
			files[root+"/owner.json"] = fmt.Sprintf(`{"compilerOptions":{"composite":true},"files":[%s]}`, strings.Join(ownerFiles, ","))
			files[root+"/shared.json"] = fmt.Sprintf(`{"files":[],"references":[%s]}`, strings.Join(sharedRefs, ","))

			fsys, baseParse := inMemoryDiscoveryHost(t, files, true)
			canonicalCfg, err := baseParse(root + "/canonical.json")
			if err != nil {
				t.Fatal(err)
			}
			parse := func(path string) (*tsoptions.ParsedCommandLine, error) {
				if strings.Contains(path, "/alias-") {
					return canonicalCfg, nil
				}
				return baseParse(path)
			}
			d := utils.NewTypeScriptProjectDiscovery(fsys, parse)
			for _, f := range targetFiles {
				assertSelectedConfig(t, d, f, root, root+"/owner.json")
			}
			stats := inspectRetainedMaps(t, d)
			if stats.retainedRedirects != parents {
				t.Fatalf("retained redirect caches = %d, want %d", stats.retainedRedirects, parents)
			}
			if stats.visitedEntries != parents*3 {
				t.Fatalf("visited = %d, want %d", stats.visitedEntries, parents*3)
			}
			if stats.missingSeenKeys != 0 {
				t.Fatalf("retained %d missing-path entries in redirect seen maps, want 0", stats.missingSeenKeys)
			}
			if stats.globalMissingKeys != 0 {
				t.Fatalf("retained %d missing-path entries in discovery seenGen maps, want 0", stats.globalMissingKeys)
			}
			if stats.aliasSeenKeys != 0 {
				t.Fatalf("retained %d duplicate alias paths in redirect seen maps, want 0", stats.aliasSeenKeys)
			}
			if stats.seenKeys != parents*4 {
				t.Fatalf("retained %d seen keys across %d paused redirect maps, want %d", stats.seenKeys, parents, parents*4)
			}
			if stats.stackFrames != parents*2 {
				t.Fatalf("retained %d stack frames across %d paused redirect searches, want %d", stats.stackFrames, parents, parents*2)
			}
		})

		t.Run("diamond-to-leaf-deduplicates-visited", func(t *testing.T) {
			const root = "/workspace"
			files := map[string]string{
				root + "/a/tsconfig.json": `{"files":["one.ts","two.ts","leaf.ts"],"references":[{"path":"../b.json"},{"path":"../c.json"}]}`,
				root + "/a/one.ts":        "export const x = 1;\n",
				root + "/a/two.ts":        "export const x = 2;\n",
				root + "/a/leaf.ts":       "export const x = 3;\n",
				root + "/b.json":          `{"compilerOptions":{"composite":true},"files":[],"references":[{"path":"./leaf.json"}]}`,
				root + "/c.json":          `{"compilerOptions":{"composite":true},"files":[],"references":[{"path":"./leaf.json"}]}`,
				root + "/leaf.json":       `{"compilerOptions":{"composite":true},"files":["a/leaf.ts"]}`,
			}
			fsys, parse := inMemoryDiscoveryHost(t, files, true)
			d := utils.NewTypeScriptProjectDiscovery(fsys, parse)
			assertSelectedConfig(t, d, root+"/a/one.ts", root, root+"/a/tsconfig.json")
			assertSelectedConfig(t, d, root+"/a/two.ts", root, root+"/a/tsconfig.json")
			assertSelectedConfig(t, d, root+"/a/leaf.ts", root, root+"/leaf.json")

			stats := inspectRetainedMaps(t, d)
			if stats.retainedRedirects != 1 {
				t.Fatalf("retained redirect caches = %d, want 1", stats.retainedRedirects)
			}
			if stats.visitedEntries != 3 {
				t.Fatalf("diamond visited = %d, want 3 (B, C, and deduplicated Leaf)", stats.visitedEntries)
			}
			if stats.seenKeys != 0 || stats.stackFrames != 0 {
				t.Fatalf("exhausted diamond retained seen=%d stack=%d, want 0 and 0", stats.seenKeys, stats.stackFrames)
			}
			if stats.missingSeenKeys != 0 || stats.globalMissingKeys != 0 || stats.aliasSeenKeys != 0 {
				t.Fatalf("unexpected retained keys: missing=%d globalMissing=%d alias=%d",
					stats.missingSeenKeys, stats.globalMissingKeys, stats.aliasSeenKeys)
			}
		})

		t.Run("redirect-match-skips-resolving-unvisited-grandchild-refs", func(t *testing.T) {
			const root = "/Workspace"
			const numGrandchildRefs = 512
			refs := make([]string, numGrandchildRefs)
			files := map[string]string{
				root + "/first.ts":  "export const a = 1;\n",
				root + "/second.ts": "export const b = 2;\n",
				root + "/other.ts":  "export const c = 3;\n",
				root + "/deep.ts":   "export const d = 4;\n",
				root + "/tsconfig.json": `{
					"files": ["first.ts", "second.ts", "deep.ts"],
					"references": [{"path": "./branch.json"}, {"path": "./owner.json"}]
				}`,
				root + "/branch.json": `{
					"compilerOptions": {"composite": true},
					"files": [],
					"references": [{"path": "./grandchild.json"}]
				}`,
				root + "/owner.json": `{
					"compilerOptions": {"composite": true},
					"files": ["first.ts", "second.ts"]
				}`,
			}
			for i := range numGrandchildRefs {
				name := fmt.Sprintf("dep%03d.json", i)
				refs[i] = fmt.Sprintf(`{"path":"./%s"}`, name)
				if i == numGrandchildRefs-1 {
					files[root+"/"+name] = `{"compilerOptions":{"composite":true},"files":["deep.ts"]}`
				} else {
					files[root+"/"+name] = `{"compilerOptions":{"composite":true},"files":["other.ts"]}`
				}
			}
			files[root+"/grandchild.json"] = fmt.Sprintf(
				`{"compilerOptions":{"composite":true},"files":["first.ts","second.ts"],"references":[%s]}`,
				strings.Join(refs, ","),
			)

			// Use a real recording host without pre-warming ResolvedProjectReferencePaths
			// so we can assert grandchild.json's rawRefPaths stays nil/unresolved.
			input := make(fstest.MapFS, len(files))
			for name, contents := range files {
				input[strings.TrimPrefix(name, "/")] = &fstest.MapFile{Data: []byte(contents)}
			}
			fsys := bundled.WrapFS(cachedvfs.From(iovfs.From(input, true)))
			parse := newRecordingHost(fsys).host().ParseConfig

			d := utils.NewTypeScriptProjectDiscovery(fsys, parse)
			assertSelectedConfig(t, d, root+"/first.ts", root, root+"/owner.json")
			assertSelectedConfig(t, d, root+"/second.ts", root, root+"/owner.json")

			configsMap := mustStructField(t, reflect.ValueOf(d), "configs")
			grandchildKey := reflect.ValueOf(tspath.ToPath(root+"/grandchild.json", "", true))
			entry := configsMap.MapIndex(grandchildKey)
			if !entry.IsValid() || entry.IsNil() {
				t.Fatal("expected grandchild.json in discovery configs cache")
			}
			refsField := mustStructField(t, entry, "refs")
			if refsField.IsNil() {
				t.Fatal("expected grandchild.json to have non-nil refs state")
			}
			rawRefs := mustStructField(t, refsField, "rawRefs")
			if !rawRefs.IsNil() {
				t.Fatalf("grandchild rawRefs resolved %d paths on an unvisited stack frame, want nil (0)", rawRefs.Len())
			}

			// A subsequent miss must resume the paused {grandchild, index: 0} frame
			// and redirect to dep511.json.
			assertSelectedConfig(t, d, root+"/deep.ts", root, root+"/dep511.json")
			rawRefsAfter := mustStructField(t, refsField, "rawRefs")
			if rawRefsAfter.Len() != numGrandchildRefs {
				t.Fatalf("grandchild rawRefs after resuming DFS = %d, want %d", rawRefsAfter.Len(), numGrandchildRefs)
			}
		})
	})
}

func mustStructField(tb testing.TB, v reflect.Value, name string) reflect.Value {
	tb.Helper()
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			tb.Fatalf("nil %s pointer while reading field %q", v.Type(), name)
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		tb.Fatalf("cannot read field %q from non-struct %s", name, v.Type())
	}
	field := v.FieldByName(name)
	if !field.IsValid() {
		tb.Fatalf("struct %s is missing expected field %q", v.Type(), name)
	}
	return field
}

func inMemoryDiscoveryHost(
	tb testing.TB,
	files map[string]string,
	caseSensitive bool,
) (vfs.FS, func(string) (*tsoptions.ParsedCommandLine, error)) {
	tb.Helper()
	input := make(fstest.MapFS, len(files))
	for name, contents := range files {
		input[strings.TrimPrefix(name, "/")] = &fstest.MapFile{Data: []byte(contents)}
	}
	fsys := bundled.WrapFS(cachedvfs.From(iovfs.From(input, caseSensitive)))
	parse := newRecordingHost(fsys).host().ParseConfig
	configs := make(map[string]*tsoptions.ParsedCommandLine)
	for path := range files {
		if strings.HasSuffix(path, ".json") {
			cfg, err := parse(path)
			if err != nil || cfg == nil {
				tb.Fatalf("parse %s: %v", path, err)
			}
			_ = cfg.ResolvedProjectReferencePaths()
			_ = cfg.FileNamesByPath()
			configs[path] = cfg
		}
	}
	return fsys, func(path string) (*tsoptions.ParsedCommandLine, error) {
		if cfg, ok := configs[path]; ok {
			return cfg, nil
		}
		return nil, fmt.Errorf("unexpected parse %q", path)
	}
}

func BenchmarkTypeScriptProjectDiscoverySingleFileWideRefs(b *testing.B) {
	const root = "/Workspace"
	const numRefs = 512
	refs := make([]string, numRefs)
	files := map[string]string{
		root + "/first.ts": "export const x = 1;\n",
		root + "/last.ts":  "export const x = 1;\n",
		root + "/other.ts": "export const x = 1;\n",
	}
	for i := range numRefs {
		name := fmt.Sprintf("Project%03d.json", i)
		refs[i] = fmt.Sprintf(`{"path":"./%s"}`, name)
		target := "other.ts"
		switch i {
		case 0:
			target = "first.ts"
		case numRefs - 1:
			target = "last.ts"
		}
		files[root+"/"+name] = fmt.Sprintf(`{"compilerOptions":{"composite":true},"files":[%q]}`, target)
	}
	files[root+"/tsconfig.json"] = fmt.Sprintf(`{"files":[],"references":[%s]}`, strings.Join(refs, ","))

	for _, caseSensitive := range []bool{true, false} {
		fsys, parse := inMemoryDiscoveryHost(b, files, caseSensitive)
		for _, tc := range []struct {
			name string
			file string
			want string
		}{
			{name: "first", file: root + "/first.ts", want: root + "/Project000.json"},
			{name: "last", file: root + "/last.ts", want: root + "/Project511.json"},
		} {
			b.Run(fmt.Sprintf("caseSensitive=%t/%s", caseSensitive, tc.name), func(b *testing.B) {
				run := func() {
					d := utils.NewTypeScriptProjectDiscovery(fsys, parse)
					cfg, err := d.Find(tc.file, root)
					if err != nil || cfg == nil || cfg.ConfigName() != tc.want {
						b.Fatalf("Find(%q) = (%v, %v), want %q", tc.file, cfg, err, tc.want)
					}
				}
				run()
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					run()
				}
			})
		}
	}
}

func BenchmarkTypeScriptProjectDiscoveryFind(b *testing.B) {
	for _, tc := range []struct {
		name           string
		rootConfigBody string
	}{
		{
			name:           "solution",
			rootConfigBody: `{"files":[],"references":[%s]}`,
		},
		{
			name:           "root-and-references",
			rootConfigBody: `{"compilerOptions":{"noLib":true},"include":["**/*.ts"],"references":[%s]}`,
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			root := tspath.NormalizePath(b.TempDir())
			const numPackages = 20
			const filesPerPackage = 20

			var refs []string
			files := make(map[string]string, numPackages*(filesPerPackage+1)+2)
			targetFiles := make([]string, 0, numPackages*filesPerPackage+1)
			for i := range numPackages {
				refs = append(refs, fmt.Sprintf(`{"path":"./pkg%02d/tsconfig.app.json"}`, i))
				var pkgFiles []string
				for j := range filesPerPackage {
					rel := fmt.Sprintf("pkg%02d/src/file%02d.ts", i, j)
					files[rel] = "export const v = 1;\n"
					pkgFiles = append(pkgFiles, fmt.Sprintf(`"src/file%02d.ts"`, j))
					targetFiles = append(targetFiles, tspath.ResolvePath(root, rel))
				}
				var subRefs []string
				if i+1 < numPackages {
					subRefs = append(subRefs, fmt.Sprintf(`{"path":"../pkg%02d/tsconfig.app.json"}`, i+1))
				}
				files[fmt.Sprintf("pkg%02d/tsconfig.app.json", i)] = fmt.Sprintf(
					`{"compilerOptions":{"noLib":true},"files":[%s],"references":[%s]}`,
					strings.Join(pkgFiles, ","),
					strings.Join(subRefs, ","),
				)
			}
			files["root-only.ts"] = "export const r = 1;\n"
			targetFiles = append(targetFiles, tspath.ResolvePath(root, "root-only.ts"))
			files["tsconfig.json"] = fmt.Sprintf(tc.rootConfigBody, strings.Join(refs, ","))
			writeDiscoveryFiles(b, root, files)

			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			host := newRecordingHost(fsys).host()
			parsedConfigs := make(map[string]*tsoptions.ParsedCommandLine, numPackages+1)
			parseConfig := func(configPath string) (*tsoptions.ParsedCommandLine, error) {
				if parsed, ok := parsedConfigs[configPath]; ok {
					return parsed, nil
				}
				parsed, err := host.ParseConfig(configPath)
				if err != nil {
					return nil, err
				}
				parsedConfigs[configPath] = parsed
				return parsed, nil
			}
			// Pre-parse configs and warm cachedvfs once so the benchmark isolates single-pass discovery.
			warmup := utils.NewTypeScriptProjectDiscovery(fsys, parseConfig)
			for _, f := range targetFiles {
				if _, err := warmup.Find(f, root); err != nil {
					b.Fatal(err)
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				discovery := utils.NewTypeScriptProjectDiscovery(fsys, parseConfig)
				for _, f := range targetFiles {
					if _, err := discovery.Find(f, root); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
