package moduleresolver

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/tailscale/hujson"
	"github.com/web-infra-dev/rslint/internal/program"
)

// Only these options affect the JSON handed to tsgo. MainFields has three
// modes: retain main, remove main, or install the lazy main-field target.
// The actual field selection still belongs to the request's resolver.
type packageJSONKey struct {
	path, conditions            string
	ignoreExports, nodeExports  bool
	mainFields, mainFieldTarget bool
}

// packageJSON is immutable and owns no Program or filesystem. Its imports tree
// and exact keys are read-only metadata for replaying each request's bookkeeping.
type packageJSON struct {
	text                string
	invalid             bool
	name                string
	hasName, hasExports bool
	imports             *hujson.Object
	exportsKeys         map[string]bool
}

func (f *nodeResolutionFS) ReadFile(name string) (string, bool) {
	physical := f.physical(name)
	text, ok := f.FS.ReadFile(physical)
	if !ok || !strings.HasSuffix(name, "/package.json") {
		return text, ok
	}
	// Keep the exact underlying read path: realpath or case folding could
	// merge distinct overlay entries. Failed reads remain uncached.
	conditions, _ := json.Marshal(f.options.Conditions) //nolint:errchkjson // A string slice always has a JSON encoding.
	key := packageJSONKey{
		path: physical, conditions: string(conditions),
		ignoreExports: f.options.IgnoreExports, nodeExports: f.options.NodeExports,
		mainFields: f.options.MainFields != nil, mainFieldTarget: len(f.options.MainFields) != 0,
	}
	pkg := program.Cached(f.resolver.program, key, func() *packageJSON {
		return transformPackageJSON(text, f.options)
	})
	if pkg.invalid {
		f.unresolved = true
		f.failure = Result{Error: "Invalid package.json at " + physical, terminal: true}
		return "", false
	}
	if pkg.imports != nil && f.importsFile == "" && strings.HasPrefix(f.request, "#") {
		f.recordImports(pkg.imports, physical)
	}
	request := f.request
	if f.importsTarget != "" {
		request = f.importsTarget
	}
	packageName, _ := module.ParsePackageName(request)
	selfReference := pkg.hasName && pkg.name == packageName
	if pkg.hasExports && (selfReference || strings.HasSuffix(name, "/"+packageName+"/package.json")) {
		f.exportsFile, f.exportsKeys = physical, pkg.exportsKeys
	}
	return pkg.text, true
}

// Transform once per path and options in a Program generation. Matching an
// imports request and recording its exports owner happen only in ReadFile.
func transformPackageJSON(text string, options Options) *packageJSON {
	if !json.Valid([]byte(text)) {
		return &packageJSON{invalid: true}
	}
	value, err := hujson.Parse([]byte(text))
	if err != nil {
		return &packageJSON{text: text}
	}
	result := &packageJSON{}
	if object, ok := value.Value.(*hujson.Object); ok {
		if options.IgnoreExports {
			object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
				key := nodePackageMemberName(member)
				return key == "exports" || key == "imports"
			})
		}
		if imports := value.Find("/imports"); imports != nil {
			filterNodeConditions(imports, options.Conditions)
			markNodeImportTargets(imports, options)
			result.imports, _ = imports.Value.(*hujson.Object)
		}
		if name := value.Find("/name"); name != nil {
			if literal, ok := name.Value.(hujson.Literal); ok && literal.Kind() == '"' {
				result.name, result.hasName = literal.String(), true
			}
		}
		// Runtime lookup must not select TypeScript's versioned declarations.
		object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
			return nodePackageMemberName(member) == "typesVersions"
		})
		for i := range object.Members {
			member := &object.Members[i]
			key := nodePackageMemberName(*member)
			if key == "main" || key == "exports" {
				if key == "exports" {
					if literal, ok := member.Value.Value.(hujson.Literal); ok && (string(literal) == "null" || string(literal) == "false" || string(literal) == `""`) {
						continue
					}
					result.hasExports = true
					result.exportsKeys = nodeExactKeys(member.Value)
					filterNodeConditions(&member.Value, options.Conditions)
					if !prepareNodeExports(&member.Value, options) {
						// A blocked selection must retain an exports map: a bare
						// null field permits tsgo's legacy main/index fallback.
						member.Value.Value = &hujson.Object{Members: []hujson.ObjectMember{{
							Name: hujson.Value{Value: hujson.String(".")}, Value: hujson.Value{Value: hujson.Literal("null")},
						}}}
					}
				}
				suffix := nodeTargetSuffix
				if key == "exports" {
					suffix = nodeExportSuffix
				}
				markNodeTargets(&member.Value, suffix)
			}
		}
		if options.MainFields != nil {
			object.Members = slices.DeleteFunc(object.Members, func(member hujson.ObjectMember) bool {
				return nodePackageMemberName(member) == "main"
			})
			if len(options.MainFields) != 0 {
				object.Members = append(object.Members, hujson.ObjectMember{
					Name:  hujson.Value{Value: hujson.String("main")},
					Value: hujson.Value{Value: hujson.String("./" + nodeMainTarget + nodeTargetSuffix)},
				})
			}
		}
	}
	result.text = string(value.Pack())
	return result
}
