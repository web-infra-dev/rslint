// Package directivefixtures loads the pinned, generated directive conformance
// corpus shared by matcher and production-fixer tests. It contains no matcher.
package directivefixtures

import (
	"encoding/json"
	"os"
	"testing"
)

const Revision = "bdd4b1da21dca81d50933edef0a015216059d877"

type Fix struct {
	Range [2]int
	Text  string
}

type Diagnostic struct {
	Message                string
	Line, Column, Severity int
	Fix                    *Fix
}

type Report struct {
	Fix        *Fix
	RuleID     string `json:"ruleId"`
	Pos        int
	Suppressed bool
}

type Variant struct {
	Severity                         string
	DisableFixes                     bool
	Diagnostics                      []Diagnostic
	Fixed, RslintFixed               string
	BatchedFixed, RslintBatchedFixed *string
}

type Case struct {
	BOM                        bool
	Name, Source, RslintSource string
	SourceLine                 int
	Reports                    []Report
	Ignored                    []string
	InvalidRules               []string
	Variants                   []Variant
}

type Corpus struct {
	Manifest struct {
		Revision                                   string
		UpstreamTests, ImportedCases, VariantCount int
		Exceptions                                 []struct{ Name, Reason string }
	}
	Cases []Case
}

func Load(t *testing.T, path string) Corpus {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var corpus Corpus
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Manifest.Revision != Revision {
		t.Fatalf("unexpected revision %q", corpus.Manifest.Revision)
	}
	if len(corpus.Cases) == 0 || len(corpus.Cases) != corpus.Manifest.ImportedCases {
		t.Fatal("missing or incomplete conformance cases")
	}
	variants := 0
	for _, c := range corpus.Cases {
		if c.Name == "" || len(c.Variants) == 0 {
			t.Fatalf("incomplete conformance case %q", c.Name)
		}
		variants += len(c.Variants)
	}
	if variants != corpus.Manifest.VariantCount {
		t.Fatal("missing conformance variants")
	}
	return corpus
}

// ByteOffset converts an oracle's UTF-16 position, rejecting split surrogates
// and out-of-range fixture data instead of silently changing the assertion.
func ByteOffset(t *testing.T, text string, offset int) int {
	t.Helper()
	units := 0
	for pos, r := range text {
		if units == offset {
			return pos
		}
		units++
		if r > 0xffff {
			units++
		}
	}
	if units == offset {
		return len(text)
	}
	t.Fatalf("invalid UTF-16 fixture offset %d (length %d)", offset, units)
	return 0
}
