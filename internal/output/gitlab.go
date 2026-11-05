package output

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type gitlabFormatter struct {
	wrote        int
	fingerprints gitlabFingerprintState
}

func newGitLabFormatter() *gitlabFormatter {
	return &gitlabFormatter{fingerprints: newGitLabFingerprintState()}
}

func (f *gitlabFormatter) begin(w *bufio.Writer, _ Report, _ bool) error {
	return w.WriteByte('[')
}

func (f *gitlabFormatter) diagnostic(w *bufio.Writer, diagnostic *Diagnostic, relativePath string) error {
	type gitlabPosition struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}
	type gitlabPositions struct {
		Begin gitlabPosition `json:"begin"`
		End   gitlabPosition `json:"end"`
	}
	type gitlabLines struct {
		Begin int `json:"begin"`
		End   int `json:"end"`
	}
	type gitlabLocation struct {
		Path      string          `json:"path"`
		Lines     gitlabLines     `json:"lines"`
		Positions gitlabPositions `json:"positions"`
	}
	type gitlabIssue struct {
		Description string         `json:"description"`
		CheckName   string         `json:"check_name"`
		Fingerprint string         `json:"fingerprint"`
		Severity    string         `json:"severity"`
		Location    gitlabLocation `json:"location"`
	}

	severity := "info"
	switch diagnostic.Severity {
	case SeverityError:
		severity = "major"
	case SeverityWarning:
		severity = "minor"
	}

	beginLine, beginColumn := diagnostic.Start.Line+1, diagnostic.Start.Column+1
	endLine, endColumn := diagnostic.End.Line+1, diagnostic.End.Column+1
	issue := gitlabIssue{
		Description: diagnostic.Message,
		CheckName:   diagnostic.RuleName,
		Fingerprint: f.fingerprints.fingerprint(relativePath, diagnostic.RuleName, diagnostic.Message, beginLine, beginColumn, endLine, endColumn),
		Severity:    severity,
		Location: gitlabLocation{
			Path:  relativePath,
			Lines: gitlabLines{Begin: beginLine, End: endLine},
			Positions: gitlabPositions{
				Begin: gitlabPosition{Line: beginLine, Column: beginColumn},
				End:   gitlabPosition{Line: endLine, Column: endColumn},
			},
		},
	}

	encoded, err := json.Marshal(issue)
	if err != nil {
		return err
	}
	if f.wrote > 0 {
		w.WriteByte(',')
	}
	w.Write(encoded)
	f.wrote++
	return nil
}

func (f *gitlabFormatter) finish(w *bufio.Writer, _ Report) error {
	w.WriteString("]\n")
	return nil
}

type gitlabFingerprintState struct {
	nextSalt map[string]int
	emitted  map[[md5.Size]byte]struct{}
}

func newGitLabFingerprintState() gitlabFingerprintState {
	return gitlabFingerprintState{
		emitted: make(map[[md5.Size]byte]struct{}),
	}
}

func (s *gitlabFingerprintState) fingerprint(filePath, ruleName, message string, startLine, startColumn, endLine, endColumn int) string {
	input := fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d", filePath, ruleName, message, startLine, startColumn, endLine, endColumn)
	// An ordinary first occurrence needs only its emitted digest. The first
	// duplicate rediscovers that digest and starts at the same historical :1
	// salt. Every candidate, including a salted one, checks all emitted IDs.
	for salt := s.nextSalt[input]; ; salt++ {
		value := input
		if salt > 0 {
			value += ":" + strconv.Itoa(salt)
		}
		sum := md5.Sum([]byte(value)) //nolint:gosec // opaque identifier, not a security boundary
		if _, exists := s.emitted[sum]; exists {
			continue
		}
		s.emitted[sum] = struct{}{}
		if salt > 0 {
			if s.nextSalt == nil {
				s.nextSalt = make(map[string]int)
			}
			s.nextSalt[input] = salt + 1
		}
		return hex.EncodeToString(sum[:])
	}
}
