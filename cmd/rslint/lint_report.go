package main

import (
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/output"
)

// lintReportOutcome is the command-owned exit policy. The returned value is
// also passed to the output renderer so the status line and exit code consume
// one decision.
func lintReportOutcome(counts output.Counts, maxWarnings int) output.Outcome {
	switch {
	case counts.Errors > 0:
		return output.Outcome{Kind: output.OutcomeDiagnosticsFailed}
	case maxWarnings >= 0 && counts.Warnings > maxWarnings:
		return output.Outcome{
			Kind:         output.OutcomeWarningLimitExceeded,
			WarningLimit: maxWarnings,
		}
	default:
		return output.Outcome{Kind: output.OutcomePassed}
	}
}

// lintReportFileCount returns the user-visible number of distinct files. Lint
// targets and compiler roots describe different phases, so combined mode uses
// their canonical union rather than either execution count or max(set sizes).
func lintReportFileCount(
	mode output.Mode,
	lintTargets []target.File,
	typeCheckedRoots []string,
	fsys vfs.FS,
) int {
	seen := make(map[string]struct{}, len(lintTargets)+len(typeCheckedRoots))
	if mode != output.ModeTypeCheckOnly {
		for _, lintTarget := range lintTargets {
			path := lintTarget.CanonicalPath
			if path == "" {
				path = lintTarget.Path
			}
			// A target.Plan has already frozen physical identity for this run.
			// Re-resolving it against the live filesystem here could make the
			// reported count disagree with the files the pipeline actually used.
			seen[rslintconfig.ExactPathID(path)] = struct{}{}
		}
	}
	if mode != output.ModeLint {
		for _, root := range typeCheckedRoots {
			seen[canonicalFilesystemPathID(root, fsys)] = struct{}{}
		}
	}
	return len(seen)
}
