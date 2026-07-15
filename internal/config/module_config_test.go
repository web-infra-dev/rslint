package config

import (
	"encoding/json"
	"testing"
)

func TestPredicateDescriptorsAreRestrictedToTrustedModuleDecoder(t *testing.T) {
	source := []byte(`[{"files":[{"$rslintPredicate":"file-1"}],"ignores":[{"$rslintPredicate":"ignore-1"}]}]`)
	var ordinary RslintConfig
	if err := json.Unmarshal(source, &ordinary); err == nil {
		t.Fatal("ordinary JSON config accepted a predicate descriptor")
	}

	trusted, err := DecodeModuleConfig(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(trusted) != 1 || !HasConfigPredicates(trusted) {
		t.Fatalf("trusted config = %+v, want retained predicates", trusted)
	}
}

func TestDecodeModuleConfigRejectsForgedPredicateShapes(t *testing.T) {
	for _, source := range []string{
		`[{"files":[{"$rslintPredicate":""}]}]`,
		`[{"files":[{"$rslintPredicate":"a","extra":true}]}]`,
		`[{"ignores":[{"predicate":"a"}]}]`,
		`[{"ignores":[["nested-is-invalid"]]}]`,
	} {
		if _, err := DecodeModuleConfig([]byte(source)); err == nil {
			t.Fatalf("DecodeModuleConfig accepted %s", source)
		}
	}
}
