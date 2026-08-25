package main

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// applyFixPass applies auto-fixes for all files in diagnosticsByFile,
// writes fixed content to disk, and returns the number of issues fixed. Write
// failures are returned after all independent files have been attempted.
func applyFixPass(diagnosticsByFile map[string][]rule.RuleDiagnostic, fsys vfs.FS) (int, error) {
	fixed := 0
	fileNames := make([]string, 0, len(diagnosticsByFile))
	for fileName := range diagnosticsByFile {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	var writeErrors []error
	for _, fileName := range fileNames {
		fileDiagnostics := diagnosticsByFile[fileName]
		var diagnosticsWithFixes []rule.RuleDiagnostic
		for _, d := range fileDiagnostics {
			if len(d.Fixes()) > 0 {
				diagnosticsWithFixes = append(diagnosticsWithFixes, d)
			}
		}
		if len(diagnosticsWithFixes) == 0 {
			continue
		}

		// The parsed text has no byte order mark — decoding the file consumed
		// it — so put it back before fixing. ApplyRuleFixes carries it through
		// to the bytes written here, and drops it only for a fix that asks.
		originalContent := diagnosticsWithFixes[0].SourceFile.Text()
		if utils.SourceHasBOM(fsys, fileName) {
			originalContent = utils.BOM + originalContent
		}
		fixedContent, unapplied, wasFixed := linter.ApplyRuleFixes(originalContent, diagnosticsWithFixes)

		if wasFixed {
			err := os.WriteFile(fileName, []byte(fixedContent), 0644)
			if err != nil {
				writeErrors = append(writeErrors, fmt.Errorf("write fixed file %q: %w", fileName, err))
			} else {
				fixed += len(diagnosticsWithFixes) - len(unapplied)
			}
		}
	}
	return fixed, errors.Join(writeErrors...)
}
