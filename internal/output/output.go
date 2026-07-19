package output

import (
	"bufio"
	"errors"
	"io"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const outputBufferSize = 4096 * 100

type Options struct {
	Format       Format
	ComparePaths tspath.ComparePathsOptions
	Quiet        bool
	ColorEnabled bool
}

type formatter interface {
	begin(w *bufio.Writer, report Report, hasVisibleDiagnostics bool) error
	diagnostic(w *bufio.Writer, view diagnosticView) error
	finish(w *bufio.Writer, report Report) error
}

func Render(dst io.Writer, report Report, options Options) error {
	selected, err := newFormatter(options)
	if err != nil {
		return err
	}

	// Build and validate every visible view before writing anything. A malformed
	// diagnostic must not leave a partial machine-readable report (for example,
	// an unterminated GitLab JSON array) in the destination.
	views := make([]diagnosticView, 0, len(report.diagnostics))
	for _, diagnostic := range report.diagnostics {
		if !isVisible(diagnostic, options.Quiet) {
			continue
		}
		view, err := newDiagnosticView(diagnostic, options.ComparePaths)
		if err != nil {
			return err
		}
		views = append(views, view)
	}

	w := bufio.NewWriterSize(dst, outputBufferSize)
	if err := selected.begin(w, report, len(views) > 0); err != nil {
		return errors.Join(err, w.Flush())
	}

	for _, view := range views {
		if err := selected.diagnostic(w, view); err != nil {
			return errors.Join(err, w.Flush())
		}
		if w.Available() < 4096 {
			if err := w.Flush(); err != nil {
				return err
			}
		}
	}

	if err := selected.finish(w, report); err != nil {
		return errors.Join(err, w.Flush())
	}
	return w.Flush()
}

func isVisible(diagnostic rule.RuleDiagnostic, quiet bool) bool {
	return !quiet || diagnostic.Severity == rule.SeverityError
}

func newFormatter(options Options) (formatter, error) {
	switch options.Format {
	case FormatDefault:
		return &defaultFormatter{colors: newColorScheme(options.ColorEnabled)}, nil
	case FormatJSONLine:
		return jsonLineFormatter{}, nil
	case FormatGitHub:
		return githubFormatter{}, nil
	case FormatGitLab:
		return newGitLabFormatter(), nil
	default:
		return nil, errors.New("unsupported output format " + options.Format.String())
	}
}
