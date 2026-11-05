package output

import (
	"bufio"
	"fmt"
	"strings"
)

type githubFormatter struct{}

func (githubFormatter) begin(_ *bufio.Writer, _ Report, _ bool) error { return nil }

func (githubFormatter) diagnostic(w *bufio.Writer, diagnostic *Diagnostic, relativePath string) error {
	severity := "notice"
	switch diagnostic.Severity {
	case SeverityError:
		severity = "error"
	case SeverityWarning:
		severity = "warning"
	}

	fmt.Fprintf(w,
		"::%s file=%s,line=%d,endLine=%d,col=%d,endColumn=%d,title=%s::%s\n",
		severity,
		escapeProperty(relativePath),
		diagnostic.Start.Line+1,
		diagnostic.End.Line+1,
		diagnostic.Start.Column+1,
		diagnostic.End.Column+1,
		escapeProperty(diagnostic.RuleName),
		escapeData(diagnostic.Message),
	)
	return nil
}

func (githubFormatter) finish(_ *bufio.Writer, _ Report) error { return nil }

func escapeData(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}

func escapeProperty(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	value = strings.ReplaceAll(value, ":", "%3A")
	value = strings.ReplaceAll(value, ",", "%2C")
	return value
}
