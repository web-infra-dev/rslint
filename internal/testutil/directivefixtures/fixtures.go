// Package directivefixtures loads the pinned, generated directive conformance
// corpus shared by matcher and production-fixer tests. It contains no matcher.
package directivefixtures

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

const Revision = "bdd4b1da21dca81d50933edef0a015216059d877"

type Fix struct {
	Range [2]int `json:"range"`
	Text  string `json:"text"`
}

var lineTerminators = regexp.MustCompile("\r\n|[\r\n\u2028\u2029]")

// LinePreservingFix keeps the upstream range while adapting its replacement to
// rslint's line-preservation policy. The checked-in upstream oracle stays intact.
func LinePreservingFix(t *testing.T, source string, fix Fix) Fix {
	t.Helper()
	deleted := source[ByteOffset(t, source, fix.Range[0]):ByteOffset(t, source, fix.Range[1])]
	if breaks := lineTerminators.FindAllString(deleted, -1); len(breaks) > 0 {
		fix.Text = strings.Join(breaks, "")
	}
	return fix
}

type Diagnostic struct {
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity int    `json:"severity"`
	Fix      *Fix   `json:"fix"`
}

type Report struct {
	Fix        *Fix   `json:"fix"`
	RuleID     string `json:"ruleId"`
	Pos        int    `json:"pos"`
	Suppressed bool   `json:"suppressed"`
}

type Variant struct {
	Severity           string       `json:"severity"`
	DisableFixes       bool         `json:"disableFixes"`
	Diagnostics        []Diagnostic `json:"diagnostics"`
	Fixed              string       `json:"fixed"`
	RslintFixed        string       `json:"rslintFixed"`
	BatchedFixed       *string      `json:"batchedFixed"`
	RslintBatchedFixed *string      `json:"rslintBatchedFixed"`
}

type Case struct {
	BOM          bool      `json:"bom"`
	Name         string    `json:"name"`
	Source       string    `json:"source"`
	RslintSource string    `json:"rslintSource"`
	SourceLine   int       `json:"sourceLine"`
	Reports      []Report  `json:"reports"`
	Ignored      []string  `json:"ignored"`
	InvalidRules []string  `json:"invalidRules"`
	Variants     []Variant `json:"variants"`
}

type Corpus struct {
	Manifest struct {
		Revision      string `json:"revision"`
		UpstreamTests int    `json:"upstreamTests"`
		ImportedCases int    `json:"importedCases"`
		VariantCount  int    `json:"variantCount"`
		Exceptions    []struct {
			Name   string `json:"name"`
			Reason string `json:"reason"`
		} `json:"exceptions"`
	} `json:"manifest"`
	Cases []Case `json:"cases"`
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
