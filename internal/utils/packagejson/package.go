// Package packagejson reads package metadata through an immutable Program.
// Dependency, publication and runtime-version policies belong to its callers.
package packagejson

import (
	"encoding/json"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
)

// Package is an immutable JSON object from one Program generation.
type Package struct {
	directory string
	text      string
	data      map[string]any
}

type packageKey string

// Read reads package.json in directory. Missing files, malformed JSON and
// non-object roots return nil. The result, including failures, is cached.
func Read(p *program.Program, directory string) *Package {
	if p.FS() == nil {
		return nil
	}
	directory = tspath.ResolvePath(p.CurrentDirectory(), directory)
	return program.Cached(p, packageKey(directory), func() *Package {
		text, ok := p.FS().ReadFile(tspath.ResolvePath(directory, "package.json"))
		if !ok {
			return nil
		}
		var data map[string]any
		if json.Unmarshal([]byte(text), &data) != nil || data == nil {
			return nil
		}
		return &Package{directory: directory, text: text, data: data}
	})
}

// FindNearest reads the closest package.json to a source file. An invalid
// nearest file returns nil; it does not change which package owns the source.
func FindNearest(p *program.Program, fileName string) *Package {
	return findNearest(p, fileName, false)
}

// FindNearestValid skips invalid package objects and continues to parent
// packages. This is the lookup policy used by eslint-plugin-n.
func FindNearestValid(p *program.Program, fileName string) *Package {
	return findNearest(p, fileName, true)
}

func findNearest(p *program.Program, fileName string, skipInvalid bool) *Package {
	if p.FS() == nil {
		return nil
	}
	fileName = tspath.ResolvePath(p.CurrentDirectory(), fileName)
	for directory := tspath.GetDirectoryPath(fileName); directory != ""; {
		directory = p.NearestPackageJSONDirectory(directory)
		if directory == "" {
			return nil
		}
		pkg := Read(p, directory)
		if pkg != nil || !skipInvalid {
			return pkg
		}
		parent := tspath.GetDirectoryPath(directory)
		if parent == directory {
			return nil
		}
		directory = parent
	}
	return nil
}

// Directory returns the normalized directory containing package.json.
func (pkg *Package) Directory() string { return pkg.directory }

// Text returns the original JSON for consumers that decode their own schema.
func (pkg *Package) Text() string { return pkg.text }

// Field reads a property path. Missing properties and non-object parents
// return nil. Returned objects and arrays are shared and must not be modified.
func (pkg *Package) Field(names ...string) any {
	var value any = pkg.data
	for _, name := range names {
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = object[name]
	}
	return value
}
