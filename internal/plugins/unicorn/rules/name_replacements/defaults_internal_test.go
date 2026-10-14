package name_replacements

import "testing"

func TestDefaultReplacementTable(t *testing.T) {
	if len(defaultReplacementPairs) != 104 {
		t.Fatalf("default replacement pairs = %d, want 104", len(defaultReplacementPairs))
	}

	names := make(map[string]struct{}, 92)
	seen := make(map[defaultReplacementPair]struct{}, len(defaultReplacementPairs))
	for _, pair := range defaultReplacementPairs {
		if pair.name == "" || pair.replacement == "" {
			t.Fatalf("empty default replacement pair: %#v", pair)
		}
		if _, duplicate := seen[pair]; duplicate {
			t.Fatalf("duplicate default replacement pair: %#v", pair)
		}
		seen[pair] = struct{}{}
		names[pair.name] = struct{}{}
	}
	if len(names) != 92 {
		t.Fatalf("default replacement names = %d, want 92", len(names))
	}

	decoded := newDefaultReplacements()
	if len(decoded) != len(names) {
		t.Fatalf("decoded replacements = %d, want %d", len(decoded), len(names))
	}
	for _, pair := range defaultReplacementPairs {
		if !decoded[pair.name][pair.replacement] {
			t.Fatalf("decoded replacements for %q omit %q", pair.name, pair.replacement)
		}
	}
}
