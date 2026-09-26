package schemagen

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/web-infra-dev/rslint/internal/config"
)

func TestGeneratedSchemaIsUpToDate(t *testing.T) {
	path := filepath.Join(repoRoot(t), "rslint-schema.json")
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	generated, _, err := Merge(existing)
	if err != nil {
		t.Fatal(err)
	}
	equal, err := CanonicalEqual(existing, generated)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("rslint-schema.json is out of date; run pnpm generate:config-schema")
	}
}

func TestMergeAddsModelBackedField(t *testing.T) {
	existing, err := os.ReadFile(filepath.Join(repoRoot(t), "rslint-schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	root, err := Parse(existing)
	if err != nil {
		t.Fatal(err)
	}
	properties, err := objectAt(rootObject(t, root), "definitions", "ParserOptions", "properties")
	if err != nil {
		t.Fatal(err)
	}
	properties.Delete("tsconfigRootDir")
	withoutField, err := root.MarshalIndent()
	if err != nil {
		t.Fatal(err)
	}

	generated, result, err := Merge(withoutField)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Added, "/definitions/ParserOptions/properties/tsconfigRootDir") {
		t.Fatalf("expected tsconfigRootDir to be reported as added, got %v", result.Added)
	}
	updated, err := Parse(generated)
	if err != nil {
		t.Fatal(err)
	}
	updatedProperties, err := objectAt(rootObject(t, updated), "definitions", "ParserOptions", "properties")
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := updatedProperties.Get("tsconfigRootDir"); !exists {
		t.Fatal("generated schema does not contain tsconfigRootDir")
	}
}

func TestMergeRepairsModelOwnedType(t *testing.T) {
	existing, err := os.ReadFile(filepath.Join(repoRoot(t), "rslint-schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	root, err := Parse(existing)
	if err != nil {
		t.Fatal(err)
	}
	properties, err := objectAt(rootObject(t, root), "definitions", "ParserOptions", "properties")
	if err != nil {
		t.Fatal(err)
	}
	projectService, ok := properties.Get("projectService")
	if !ok {
		t.Fatal("schema is missing projectService")
	}
	projectServiceObject, ok := projectService.AsObject()
	if !ok {
		t.Fatal("projectService is not an object")
	}
	projectServiceObject.Set("type", newScalar("boolean"))
	broken, err := root.MarshalIndent()
	if err != nil {
		t.Fatal(err)
	}

	generated, result, err := Merge(broken)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Updated, "/definitions/ParserOptions/properties/projectService") {
		t.Fatalf("expected projectService type repair, got %v", result.Updated)
	}
	updated, err := Parse(generated)
	if err != nil {
		t.Fatal(err)
	}
	updatedProperties, err := objectAt(rootObject(t, updated), "definitions", "ParserOptions", "properties")
	if err != nil {
		t.Fatal(err)
	}
	updatedService, _ := updatedProperties.Get("projectService")
	updatedObject, _ := updatedService.AsObject()
	updatedType, _ := updatedObject.Get("type")
	if got := updatedType.toAny(); !reflect.DeepEqual(got, []any{"boolean", "null"}) {
		t.Fatalf("projectService type = %#v, want boolean/null union", got)
	}
}

func TestMergePreservesOpenLanguageOptionsFields(t *testing.T) {
	existing, err := os.ReadFile(filepath.Join(repoRoot(t), "rslint-schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	generated, _, err := Merge(existing)
	if err != nil {
		t.Fatal(err)
	}
	root, err := Parse(generated)
	if err != nil {
		t.Fatal(err)
	}
	properties, err := objectAt(rootObject(t, root), "definitions", "LanguageOptions", "properties")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ecmaVersion", "sourceType", "globals", "ecmaFeatures"} {
		if _, exists := properties.Get(name); !exists {
			t.Errorf("open language option %q was removed", name)
		}
	}
}

func TestModelFieldsIgnorePrivateAndJSONIgnoredFields(t *testing.T) {
	fields := modelFields(reflect.TypeOf(config.ConfigEntry{}))
	for _, field := range fields {
		if field.Name == "FilePatternGroups" || field.Name == "collectedGitignore" {
			t.Fatalf("internal field %q should not be modeled", field.Name)
		}
	}

	parserFields := modelFields(reflect.TypeOf(config.ParserOptions{}))
	if !containsField(parserFields, "tsconfigRootDir") {
		t.Fatal("ParserOptions model should include tsconfigRootDir")
	}
}

func TestCanonicalIgnoresFormatting(t *testing.T) {
	formatted := []byte("{\n  \"b\": 1,\n  \"a\": [true, null]\n}\n")
	compact := []byte(`{"a":[true,null],"b":1}`)
	left, err := Canonical(formatted)
	if err != nil {
		t.Fatal(err)
	}
	right, err := Canonical(compact)
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) {
		t.Fatalf("canonical forms differ: %s != %s", left, right)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func rootObject(t *testing.T, node Node) *Object {
	t.Helper()
	object, ok := node.AsObject()
	if !ok {
		t.Fatal("schema root is not an object")
	}
	return object
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsField(fields []fieldInfo, target string) bool {
	for _, field := range fields {
		if field.Name == target {
			return true
		}
	}
	return false
}
