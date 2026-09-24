package utils_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
)

func TestVisitModulesAMDAndIgnore(t *testing.T) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/visitor.ts", Path: "/visitor.ts"}, `
import "ignored";
export * from "ignored";
import("ignored");
require("ignored");
define(["ignored"], cb);
define(["require", "exports", "module", , 1, null, `+"`template`"+`, ...["spread"], ("first")], cb);
(require)((["second"]), cb);
define?.(["third"], cb);
define(["wrong arity"]);
define("named", ["wrong arity"], cb);
define(["asserted"] as string[], cb);
obj.define(["member"], cb);
`, core.ScriptKindTS)
	var sources []string
	listeners := import_utils.VisitModules(func(source, importer *ast.Node) {
		sources = append(sources, source.Text())
		if source != importer {
			t.Error("AMD importer must be the dependency literal, matching upstream")
		}
	}, import_utils.VisitModulesOptions{ESModule: true, Commonjs: true, AMD: true, Ignore: []string{"^ignored$"}})
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if listener := listeners[node.Kind]; listener != nil {
			listener(node)
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(file.AsNode())
	if want := []string{"module", "first", "second", "third"}; !reflect.DeepEqual(sources, want) {
		t.Fatalf("sources = %v, want %v", sources, want)
	}
}

func TestVisitModulesInvalidIgnore(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid regex must fail instead of silently removing the filter")
		}
	}()
	import_utils.VisitModules(func(_, _ *ast.Node) {}, import_utils.VisitModulesOptions{Ignore: []string{"["}})
}
