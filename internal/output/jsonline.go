package output

import (
	"bufio"
	"encoding/json"
)

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

type jsonLineFormatter struct {
	row jsonDiagnostic
}

func (*jsonLineFormatter) begin(_ *bufio.Writer, _ Report, _ bool) error { return nil }

func (f *jsonLineFormatter) diagnostic(w *bufio.Writer, diagnostic *Diagnostic, relativePath string) error {
	f.row = jsonDiagnostic{
		RuleName: diagnostic.RuleName,
		Message:  diagnostic.Message,
		FilePath: relativePath,
		Range: jsonRange{
			Start: jsonLocation{Line: diagnostic.Start.Line + 1, Column: diagnostic.Start.Column + 1},
			End:   jsonLocation{Line: diagnostic.End.Line + 1, Column: diagnostic.End.Column + 1},
		},
		Severity: diagnostic.Severity.String(),
	}

	encoded, err := json.Marshal(&f.row)
	if err != nil {
		return err
	}
	w.Write(encoded)
	w.WriteByte('\n')
	return nil
}

func (*jsonLineFormatter) finish(_ *bufio.Writer, _ Report) error { return nil }
