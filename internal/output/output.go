package output

import (
	"bufio"
	"errors"
	"io"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
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
	diagnostic(w *bufio.Writer, diagnostic *Diagnostic, relativePath string) error
	finish(w *bufio.Writer, report Report) error
}

func Render(dst io.Writer, report Report, options Options) error {
	if options.Format == FormatDefault && !report.hasSummary {
		return errors.New("default output requires a report summary")
	}
	selected, err := newFormatter(options)
	if err != nil {
		return err
	}

	// Validate every visible diagnostic before writing anything. A malformed
	// diagnostic must not leave a partial machine-readable report (for example,
	// an unterminated GitLab JSON array) in the destination.
	hasVisible := false
	for i := range report.diagnostics {
		diagnostic := &report.diagnostics[i]
		if !isVisible(diagnostic, options.Quiet) {
			continue
		}
		if err := validateDiagnostic(diagnostic, options.Format == FormatDefault); err != nil {
			return err
		}
		hasVisible = true
	}

	w := bufio.NewWriterSize(dst, outputBufferSize)
	if err := selected.begin(w, report, hasVisible); err != nil {
		return errors.Join(err, w.Flush())
	}

	var lastPath, relativePath string
	for i := range report.diagnostics {
		diagnostic := &report.diagnostics[i]
		if !isVisible(diagnostic, options.Quiet) {
			continue
		}
		if diagnostic.FilePath != lastPath {
			lastPath = diagnostic.FilePath
			relativePath = tspath.ConvertToRelativePath(lastPath, options.ComparePaths)
		}

		if err := selected.diagnostic(w, diagnostic, relativePath); err != nil {
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

func isVisible(diagnostic *Diagnostic, quiet bool) bool {
	return !quiet || diagnostic.Severity == SeverityError
}

func newFormatter(options Options) (formatter, error) {
	switch options.Format {
	case FormatDefault:
		return &defaultFormatter{colors: newColorScheme(options.ColorEnabled)}, nil
	case FormatJSONLine:
		return &jsonLineFormatter{}, nil
	case FormatGitHub:
		return githubFormatter{}, nil
	case FormatGitLab:
		return newGitLabFormatter(), nil
	default:
		return nil, errors.New("unsupported output format " + options.Format.String())
	}
}
