package output

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/web-infra-dev/rslint/internal/rule"
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

func (f *gitlabFormatter) diagnostic(w *bufio.Writer, view diagnosticView) error {
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
	switch view.raw.Severity {
	case rule.SeverityError:
		severity = "major"
	case rule.SeverityWarning:
		severity = "minor"
	}

	beginLine, beginColumn := view.start.line+1, view.start.column+1
	endLine, endColumn := view.end.line+1, view.end.column+1
	issue := gitlabIssue{
		Description: view.raw.Message.Description,
		CheckName:   view.raw.RuleName,
		Fingerprint: f.fingerprints.fingerprint(view.relativePath, view.raw.RuleName, view.raw.Message.Description, beginLine, beginColumn, endLine, endColumn),
		Severity:    severity,
		Location: gitlabLocation{
			Path:  view.relativePath,
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
	emitted  map[string]struct{}
}

func newGitLabFingerprintState() gitlabFingerprintState {
	return gitlabFingerprintState{
		nextSalt: make(map[string]int),
		emitted:  make(map[string]struct{}),
	}
}

func (s *gitlabFingerprintState) fingerprint(filePath, ruleName, message string, startLine, startColumn, endLine, endColumn int) string {
	input := fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d", filePath, ruleName, message, startLine, startColumn, endLine, endColumn)
	digest := func(value string) string {
		sum := md5.Sum([]byte(value)) //nolint:gosec // opaque identifier, not a security boundary
		return hex.EncodeToString(sum[:])
	}

	// Preserve the historical hash for the first ordinary occurrence and the
	// historical :1, :2, ... salts for duplicates. Track every fingerprint that
	// was actually emitted as well: a salted fingerprint can otherwise collide
	// with another diagnostic whose unsalted tuple contains the same colon-
	// separated suffix.
	salt := s.nextSalt[input]
	for {
		value := input
		if salt > 0 {
			value += ":" + strconv.Itoa(salt)
		}
		salt++
		s.nextSalt[input] = salt

		fingerprint := digest(value)
		if _, exists := s.emitted[fingerprint]; exists {
			continue
		}
		s.emitted[fingerprint] = struct{}{}
		return fingerprint
	}
}
