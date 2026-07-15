package output

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
)

type githubFormatter struct{}

func (githubFormatter) begin(_ *bufio.Writer, _ Report, _ bool) error { return nil }

func (githubFormatter) diagnostic(w *bufio.Writer, view diagnosticView) error {
	severity := "notice"
	switch view.raw.Severity {
	case rule.SeverityError:
		severity = "error"
	case rule.SeverityWarning:
		severity = "warning"
	}

	fmt.Fprintf(w,
		"::%s file=%s,line=%d,endLine=%d,col=%d,endColumn=%d,title=%s::%s\n",
		severity,
		escapeProperty(view.relativePath),
		view.start.line+1,
		view.end.line+1,
		view.start.column+1,
		view.end.column+1,
		escapeProperty(view.raw.RuleName),
		escapeData(view.raw.Message.Description),
	)
	return nil
}

func (githubFormatter) fileWarning(w *bufio.Writer, view fileWarningView) error {
	fmt.Fprintf(w, "::warning file=%s::%s\n", escapeProperty(view.relativePath), escapeData(view.raw.Message))
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
