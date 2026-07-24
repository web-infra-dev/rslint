package no_restricted_types

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestNoRestrictedTypesEditDemand(t *testing.T) {
	t.Parallel()

	const source = "let value: Banned;\n"

	configs := []struct {
		name            string
		bannedType      any
		wantFix         bool
		fixText         string
		suggestionTexts []string
	}{
		{
			name:       "diagnostic only",
			bannedType: true,
		},
		{
			name: "autofix only",
			bannedType: map[string]interface{}{
				"fixWith": "Fixed",
			},
			wantFix: true,
			fixText: "Fixed",
		},
		{
			name: "empty autofix replacement",
			bannedType: map[string]interface{}{
				"fixWith": "",
			},
			wantFix: true,
		},
		{
			name: "suggestions only",
			bannedType: map[string]interface{}{
				"suggest": []interface{}{"SuggestedOne", ""},
			},
			suggestionTexts: []string{"SuggestedOne", ""},
		},
		{
			name: "autofix and suggestions",
			bannedType: map[string]interface{}{
				"fixWith": "Fixed",
				"suggest": []interface{}{"SuggestedOne", "SuggestedTwo"},
			},
			wantFix:         true,
			fixText:         "Fixed",
			suggestionTexts: []string{"SuggestedOne", "SuggestedTwo"},
		},
	}

	for index, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createNoRestrictedTypesProgram(
				t,
				fmt.Sprintf("edit-demand-%d.ts", index),
				source,
			)
			options := rule.NormalizeOptions(map[string]interface{}{
				"types": map[string]interface{}{
					"Banned": config.bannedType,
				},
			})
			diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
			for _, demand := range []rule.EditDemand{
				rule.EditDemandNone,
				rule.EditDemandAutofix,
				rule.EditDemandSuggestion,
				rule.EditDemandAll,
			} {
				got := lintNoRestrictedTypesWithDemand(program, sourceFile, options, demand)
				if len(got) != 1 {
					t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(got))
				}
				if got[0].Message.Description != "Don't use `Banned` as a type." {
					t.Errorf("demand %d: unexpected message %q", demand, got[0].Message.Description)
				}
				diagnostics[demand] = got[0]
			}

			diagnosticsOnly := diagnostics[rule.EditDemandNone]
			for demand, diagnostic := range diagnostics {
				requireSameNoRestrictedTypesDiagnostic(t, diagnosticsOnly, diagnostic, demand)
			}
			requireNoRestrictedTypeEdits(t, diagnostics[rule.EditDemandNone], rule.EditDemandNone)
			if diagnostics[rule.EditDemandAutofix].Suggestions != nil {
				t.Errorf("autofix-only demand unexpectedly materialized suggestions")
			}
			if diagnostics[rule.EditDemandSuggestion].FixesPtr != nil {
				t.Errorf("suggestion-only demand unexpectedly materialized autofixes")
			}

			autofixOnly := diagnostics[rule.EditDemandAutofix].FixesPtr
			allFixes := diagnostics[rule.EditDemandAll].FixesPtr
			if !config.wantFix {
				if autofixOnly != nil || allFixes != nil {
					t.Fatalf("unexpected autofixes: autofix=%#v all=%#v", autofixOnly, allFixes)
				}
			} else {
				if autofixOnly == nil || allFixes == nil || !reflect.DeepEqual(*autofixOnly, *allFixes) {
					t.Fatalf("autofix artifacts differ between autofix-only and all demand")
				}
				if len(*autofixOnly) != 1 || (*autofixOnly)[0].Text != config.fixText {
					t.Fatalf("autofixes = %#v, want replacement %q", *autofixOnly, config.fixText)
				}
			}

			suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
			allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
			if len(config.suggestionTexts) == 0 {
				if suggestionOnly != nil || allSuggestions != nil {
					t.Fatalf("unexpected suggestions: suggestion=%#v all=%#v", suggestionOnly, allSuggestions)
				}
				return
			}
			if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
				t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
			}
			if len(*suggestionOnly) != len(config.suggestionTexts) {
				t.Fatalf("suggestions = %#v, want %d", *suggestionOnly, len(config.suggestionTexts))
			}
			for index, suggestion := range *suggestionOnly {
				replacement := config.suggestionTexts[index]
				if suggestion.Message.Description != "Replace `Banned` with `"+replacement+"`." {
					t.Errorf("suggestion %d message = %q", index, suggestion.Message.Description)
				}
				if !reflect.DeepEqual(suggestion.Message.Data, map[string]string{
					"name":        "Banned",
					"replacement": replacement,
				}) {
					t.Errorf("suggestion %d data = %#v", index, suggestion.Message.Data)
				}
				if len(suggestion.FixesArr) != 1 || suggestion.FixesArr[0].Text != replacement {
					t.Errorf("suggestion %d fixes = %#v", index, suggestion.FixesArr)
				}
			}
		})
	}
}

func lintNoRestrictedTypesWithDemand(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	options []any,
	demand rule.EditDemand,
) []rule.RuleDiagnostic {
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         program,
		File:            sourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: noRestrictedTypesConfiguredRules(options),
		ExcludePaths:    []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	return diagnostics
}

func requireSameNoRestrictedTypesDiagnostic(
	t *testing.T,
	want rule.RuleDiagnostic,
	got rule.RuleDiagnostic,
	demand rule.EditDemand,
) {
	t.Helper()
	want.FixesPtr = nil
	want.Suggestions = nil
	got.FixesPtr = nil
	got.Suggestions = nil
	if !reflect.DeepEqual(got, want) {
		t.Errorf("demand %d changed diagnostic metadata:\ngot:  %#v\nwant: %#v", demand, got, want)
	}
}

func requireNoRestrictedTypeEdits(
	t *testing.T,
	diagnostic rule.RuleDiagnostic,
	demand rule.EditDemand,
) {
	t.Helper()
	if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
		t.Errorf(
			"demand %d unexpectedly materialized edits: fixes=%#v suggestions=%#v",
			demand,
			diagnostic.FixesPtr,
			diagnostic.Suggestions,
		)
	}
}
