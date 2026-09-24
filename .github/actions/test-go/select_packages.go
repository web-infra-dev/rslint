// Select tests from changed repository files and a go list -test -json stream.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type goPackage struct {
	ImportPath      string   `json:"ImportPath"`
	ForTest         string   `json:"ForTest"`
	Dir             string   `json:"Dir"`
	Match           []string `json:"Match"`
	Deps            []string `json:"Deps"`
	TestGoFiles     []string `json:"TestGoFiles"`
	XTestGoFiles    []string `json:"XTestGoFiles"`
	EmbedFiles      []string `json:"EmbedFiles"`
	TestEmbedFiles  []string `json:"TestEmbedFiles"`
	XTestEmbedFiles []string `json:"XTestEmbedFiles"`
	Module          *struct {
		Dir string `json:"Dir"`
	} `json:"Module"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) != 2 {
		return errors.New("usage: select_packages changed-files metadata-json")
	}
	changed, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	metadata, err := os.Open(args[1])
	if err != nil {
		return err
	}
	defer metadata.Close()
	files := strings.Split(strings.ReplaceAll(string(changed), "\r\n", "\n"), "\n")
	packages, err := selectPackages(files, metadata)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if _, err := fmt.Fprintln(output, pkg); err != nil {
			return err
		}
	}
	return nil
}

func selectPackages(changed []string, metadata io.Reader) ([]string, error) {
	directories := map[string]string{}
	fixtures := map[string]string{}
	embedded := map[string][]string{}
	dependencies := map[string][]string{}
	decoder := json.NewDecoder(metadata)
	for {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read Go package metadata: %w", err)
		}
		// A generated test binary has neither a matching pattern nor ForTest.
		// Its name can collide with a real package ending in .test.
		if pkg.ForTest == "" && len(pkg.Match) == 0 {
			continue
		}
		owner := pkg.ImportPath
		if pkg.ForTest != "" {
			owner = pkg.ForTest
		}
		dependencies[owner] = append(dependencies[owner], pkg.Deps...)
		// Test variants contribute dependencies. Only the original packages
		// matched by ./cmd/... and ./internal/... own files; test binaries do not.
		if pkg.ForTest != "" {
			continue
		}
		if pkg.Module == nil || pkg.Module.Dir == "" || pkg.Dir == "" || pkg.ImportPath == "" {
			return nil, errors.New("missing Go package path or directory in metadata")
		}
		directory, err := filepath.Rel(pkg.Module.Dir, pkg.Dir)
		if err != nil {
			return nil, err
		}
		directory = filepath.ToSlash(directory)
		if directory != "cmd" && directory != "internal" && !isGoPath(directory) {
			return nil, fmt.Errorf("package is outside the Go test suite: %s", directory)
		}
		directories[directory] = pkg.ImportPath
		if len(pkg.TestGoFiles)+len(pkg.XTestGoFiles) > 0 {
			fixtures[directory+"/testdata/"] = pkg.ImportPath
		}
		for _, file := range slices.Concat(pkg.EmbedFiles, pkg.TestEmbedFiles, pkg.XTestEmbedFiles) {
			file = directory + "/" + filepath.ToSlash(file)
			embedded[file] = append(embedded[file], pkg.ImportPath)
		}
	}
	if len(directories) == 0 {
		return nil, errors.New("no Go packages found in metadata")
	}

	seeds := map[string]bool{}
	hasChanges := false
	for _, file := range changed {
		if file == "" {
			continue
		}
		hasChanges = true
		if !isGoPath(file) {
			return nil, fmt.Errorf("invalid changed Go input: %s", file)
		}
		found := false
		if strings.HasSuffix(file, ".go") {
			if owner := directories[path.Dir(file)]; owner != "" {
				seeds[owner], found = true, true
			}
		}
		for _, owner := range embedded[file] {
			seeds[owner], found = true, true
		}
		// Fixtures belong to the package beside testdata, including deleted
		// fixtures and .go files that are data for a test project.
		for prefix, owner := range fixtures {
			if strings.HasPrefix(file, prefix) {
				seeds[owner], found = true, true
			}
		}
		if found {
			continue
		}
		if isDocumentation(file) {
			// Existing unowned docs are excluded. Removed docs may previously
			// have been embedded, which current metadata cannot establish.
			if info, err := os.Stat(filepath.FromSlash(file)); err == nil && info.Mode().IsRegular() {
				continue
			}
		}
		return nil, fmt.Errorf("no Go package owns changed file: %s", file)
	}
	if !hasChanges {
		return nil, errors.New("changed file list is empty")
	}

	selected := []string{}
	for _, pkg := range directories {
		affected := seeds[pkg]
		for _, dep := range dependencies[pkg] {
			// Go may name a test dependency "package [owner.test]".
			dep, _, _ = strings.Cut(dep, " [")
			affected = affected || seeds[dep]
		}
		if affected {
			selected = append(selected, pkg)
		}
	}
	slices.Sort(selected)
	return selected, nil
}

func isGoPath(file string) bool {
	return path.Clean(file) == file && (strings.HasPrefix(file, "cmd/") || strings.HasPrefix(file, "internal/"))
}

func isDocumentation(file string) bool {
	return strings.HasSuffix(file, ".md") || strings.HasSuffix(file, ".mdx") ||
		path.Base(file) == "_meta.json" || path.Base(file) == "dictionary.txt"
}
