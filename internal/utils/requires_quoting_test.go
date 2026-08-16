package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
)

func TestRequiresQuotingHonorsScriptTarget(t *testing.T) {
	tests := []struct {
		name   string
		target core.ScriptTarget
		want   bool
	}{
		{name: "plain", target: core.ScriptTargetES5, want: false},
		{name: "a1", target: core.ScriptTargetES5, want: false},
		{name: "1a", target: core.ScriptTargetES5, want: true},
		{name: "a-b", target: core.ScriptTargetESNext, want: true},
		// U+037F and U+00B7 entered the identifier tables after ES5's
		// Unicode 6.2 snapshot. The latter is valid only as a part.
		{name: "Ϳ", target: core.ScriptTargetES5, want: true},
		{name: "Ϳ", target: core.ScriptTargetES2015, want: false},
		{name: "Ϳ", target: core.ScriptTargetNone, want: false},
		{name: "a·", target: core.ScriptTargetES5, want: true},
		{name: "a·", target: core.ScriptTargetES2015, want: false},
		// TypeScript's helper examines UTF-16 code units, so supplementary
		// identifier code points are quoted even for modern targets.
		{name: "𐀀", target: core.ScriptTargetESNext, want: true},
		{name: "", target: core.ScriptTargetESNext, want: true},
	}

	for _, test := range tests {
		t.Run(test.name+"/"+targetName(test.target), func(t *testing.T) {
			if got := RequiresQuoting(test.name, test.target); got != test.want {
				t.Fatalf("RequiresQuoting(%q, %v) = %v, want %v", test.name, test.target, got, test.want)
			}
		})
	}
}

func TestES5IdentifierTablesMatchTypeScript593(t *testing.T) {
	tests := []struct {
		name string
		data []uint16
		len  int
		hash string
	}{
		{"start", unicodeES5IdentifierStart[:], 740, "af52df1b36a602f4c5e5565a72024a98b0c97c358d475e75fcf6bf615f155f2f"},
		{"part", unicodeES5IdentifierPart[:], 856, "26a7d4503cfcef4b9f165098c0e2be4cbeee3e09cde858eb6bef561212656c63"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if len(test.data) != test.len {
				t.Fatalf("table length = %d, want %d", len(test.data), test.len)
			}
			encoded, err := json.Marshal(test.data)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(encoded)
			if got := hex.EncodeToString(sum[:]); got != test.hash {
				t.Fatalf("table SHA-256 = %s, want %s", got, test.hash)
			}
		})
	}
}

func targetName(target core.ScriptTarget) string {
	switch target {
	case core.ScriptTargetNone:
		return "none"
	case core.ScriptTargetES5:
		return "es5"
	case core.ScriptTargetES2015:
		return "es2015"
	default:
		return "other"
	}
}
