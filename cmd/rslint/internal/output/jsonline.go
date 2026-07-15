package output

import (
	"bufio"
	"encoding/json"
)

type jsonLineFormatter struct{}

func (jsonLineFormatter) begin(_ *bufio.Writer, _ Report, _ bool) error { return nil }

func (jsonLineFormatter) diagnostic(w *bufio.Writer, view diagnosticView) error {
	type jsonLocation struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}
	type jsonRange struct {
		Start jsonLocation `json:"start"`
		End   jsonLocation `json:"end"`
	}
	type jsonDiagnostic struct {
		RuleName string    `json:"ruleName"`
		Message  string    `json:"message"`
		FilePath string    `json:"filePath"`
		Range    jsonRange `json:"range"`
		Severity string    `json:"severity"`
	}

	diagnostic := jsonDiagnostic{
		RuleName: view.raw.RuleName,
		Message:  view.raw.Message.Description,
		FilePath: view.relativePath,
		Range: jsonRange{
			Start: jsonLocation{Line: view.start.line + 1, Column: view.start.column + 1},
			End:   jsonLocation{Line: view.end.line + 1, Column: view.end.column + 1},
		},
		Severity: view.raw.Severity.String(),
	}

	encoded, err := json.Marshal(diagnostic)
	if err != nil {
		return err
	}
	w.Write(encoded)
	w.WriteByte('\n')
	return nil
}

func (jsonLineFormatter) fileWarning(w *bufio.Writer, view fileWarningView) error {
	encoded, err := json.Marshal(struct {
		Message  string `json:"message"`
		FilePath string `json:"filePath"`
		Severity string `json:"severity"`
	}{
		Message: view.raw.Message, FilePath: view.relativePath, Severity: "warning",
	})
	if err != nil {
		return err
	}
	w.Write(encoded)
	return w.WriteByte('\n')
}

func (jsonLineFormatter) finish(_ *bufio.Writer, _ Report) error { return nil }
