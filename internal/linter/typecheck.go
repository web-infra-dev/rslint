package linter

import (
	"context"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// typeCheckRequest bundles the inputs for runTypeCheckAcrossPrograms.
type typeCheckRequest struct {
	Sources        []*program.DiagnosticSource
	SingleThreaded bool
	OnDiagnostic   DiagnosticHandler
}

// runTypeCheckAcrossPrograms runs program-scoped semantic diagnostics
// through independent capabilities for complete source generations. Source
// construction remains private; the linter schedules only the capability it needs.
//
// Aggregates diagnostics via collectNoEmitDiagnostics, which mirrors
// compiler.GetDiagnosticsOfAnyProgram(file=nil) but enforces tsc --noEmit
// semantics regardless of whether the user's tsconfig sets noEmit.
// It covers config-parsing, syntactic, program, global, semantic, and
// (when GetEmitDeclarations() is set) declaration diagnostics.
// @ts-ignore / @ts-expect-error suppression and TS2578 unused-directive
// errors are handled inside typescript-go.
//
// Cross-program dedupe: shared declaration files (typical with project
// references and pulled-in node_modules .d.ts) appear in multiple programs.
// We dedupe on canonical (filePath, code, pos, end, final message), including
// message chains and related information.
//
// req.Sources is a private capability slice transferred by lintExecution.
// Dispatch consumes each slot; completed tasks retain only diagnostics.
// req.OnDiagnostic must be non-nil.
func runTypeCheckAcrossPrograms(req typeCheckRequest) {
	collected := make([][]collectedTypeCheckDiagnostic, len(req.Sources))
	wg := newProgramWorkGroup(req.SingleThreaded, len(req.Sources))
	for i, source := range req.Sources {
		req.Sources[i] = nil
		if source == nil {
			continue
		}
		programIndex := i
		diagnosticSource := source
		wg.Queue(func() {
			collected[programIndex] = runTypeCheckForProgram(diagnosticSource)
		})
	}
	wg.RunAndWait()

	// Program diagnostics are computed in parallel above, but survivor choice
	// follows stable Program input order rather than goroutine completion order.
	emitDeduplicatedTypeCheckDiagnostics(collected, req.OnDiagnostic)
}

func emitDeduplicatedTypeCheckDiagnostics(collected [][]collectedTypeCheckDiagnostic, onDiagnostic DiagnosticHandler) {
	seen := make(map[typeCheckDedupeKey]struct{})
	for _, programDiagnostics := range collected {
		for _, diagnostic := range programDiagnostics {
			if _, duplicate := seen[diagnostic.key]; duplicate {
				continue
			}
			seen[diagnostic.key] = struct{}{}
			onDiagnostic(diagnostic.ruleDiagnostic)
		}
	}
}

type collectedTypeCheckDiagnostic struct {
	key            typeCheckDedupeKey
	ruleDiagnostic rule.RuleDiagnostic
}

func runTypeCheckForProgram(source *program.DiagnosticSource) []collectedTypeCheckDiagnostic {
	ctx := context.Background()
	diags := source.NoEmitDiagnostics(ctx)
	collected := make([]collectedTypeCheckDiagnostic, 0, len(diags))

	for _, d := range diags {
		file := d.File()
		// Diagnostics without an attached file have no source location to
		// render against. Examples: TS18003 ("No inputs were found in
		// config file"), TS5108 ("Option ... has been removed"), TS2318
		// ("Cannot find global type") — all surface from
		// GetDiagnosticsOfAnyProgram with d.File()==nil.
		//
		// rslint --type-check intentionally drops these — see
		// "Diagnostics without a source location" in
		// website/docs/en/guide/type-checking.md. The behaviour is
		// pinned by TestBoundary_NoSourceLocationDiagnosticsDropped.
		if file == nil {
			continue
		}

		// Suppression for @ts-ignore / @ts-expect-error / @ts-nocheck and
		// the corresponding TS2578 unused-directive errors is already
		// handled inside typescript-go (GetDiagnosticsOfAnyProgram). We do
		// not add an rslint-specific disable channel here on purpose:
		// type-check is meant to be a transparent passthrough of tsc.

		message := flattenDiagnosticMessage(d)
		collected = append(collected, collectedTypeCheckDiagnostic{
			key:            typeCheckDedupeKeyForDiagnostic(source, d),
			ruleDiagnostic: newTypeScriptDiagnostic(file, d, message),
		})
	}
	return collected
}

type typeCheckDedupeKey struct {
	path    string
	code    int32
	pos     int
	end     int
	message string
}

func typeCheckDedupeKeyForDiagnostic(source *program.DiagnosticSource, d *ast.Diagnostic) typeCheckDedupeKey {
	loc := d.Loc()
	return typeCheckDedupeKey{
		path:    typeCheckFilesystemPathID(source, d.File().FileName()),
		code:    d.Code(),
		pos:     loc.Pos(),
		end:     loc.End(),
		message: flattenDiagnosticMessageForIdentity(source, d),
	}
}

func typeCheckFilesystemPathID(source *program.DiagnosticSource, filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if source == nil {
		return filePath
	}
	filePath = tspath.GetNormalizedAbsolutePath(filePath, source.CurrentDirectory())
	fsys := source.FS()
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return string(tspath.ToPath(filePath, "", true))
}

// flattenDiagnosticMessage builds a human-readable message from a
// TypeScript diagnostic, including its MessageChain and
// RelatedInformation. Format mirrors tsc's output.
func flattenDiagnosticMessage(d *ast.Diagnostic) string {
	return flattenDiagnosticMessageWithRelatedPath(d, func(file *ast.SourceFile) string {
		return file.FileName()
	})
}

func flattenDiagnosticMessageForIdentity(source *program.DiagnosticSource, d *ast.Diagnostic) string {
	return flattenDiagnosticMessageWithRelatedPath(d, func(file *ast.SourceFile) string {
		return typeCheckFilesystemPathID(source, file.FileName())
	})
}

func flattenDiagnosticMessageWithRelatedPath(d *ast.Diagnostic, relatedPath func(*ast.SourceFile) string) string {
	var b strings.Builder
	b.WriteString(d.String())
	for _, chain := range d.MessageChain() {
		flattenMessageChain(&b, chain, 1)
	}
	for _, related := range d.RelatedInformation() {
		if related.File() != nil {
			line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(related.File(), related.Pos())
			fmt.Fprintf(&b, "\n  %s:%d: %s", relatedPath(related.File()), line+1, related.String())
		}
	}
	return b.String()
}

func flattenMessageChain(b *strings.Builder, chain *ast.Diagnostic, level int) {
	b.WriteByte('\n')
	for range level {
		b.WriteString("  ")
	}
	b.WriteString(chain.String())
	for _, child := range chain.MessageChain() {
		flattenMessageChain(b, child, level+1)
	}
}
