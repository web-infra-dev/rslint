package config

import (
	"encoding/json"
	"fmt"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// UnusedDirectiveSeverity accepts ESLint flat-config severities. Its boolean
// compatibility differs from the historical CLI flag: true means warn.
type UnusedDirectiveSeverity string

func (s *UnusedDirectiveSeverity) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	switch value {
	case false, float64(0), "off":
		*s = "off"
	case true, float64(1), "warn":
		*s = "warn"
	case float64(2), "error":
		*s = "error"
	default:
		return fmt.Errorf("linterOptions.reportUnusedDisableDirectives: expected off, warn, error, 0, 1, 2, or a boolean")
	}
	return nil
}

type LinterOptions struct {
	ReportUnusedDisableDirectives *UnusedDirectiveSeverity `json:"reportUnusedDisableDirectives,omitempty"`
}

func (config *MergedConfig) UnusedDirectiveSeverity() rule.DiagnosticSeverity {
	if config == nil || config.LinterOptions == nil || config.LinterOptions.ReportUnusedDisableDirectives == nil {
		return rule.SeverityOff
	}
	return rule.ParseSeverity(string(*config.LinterOptions.ReportUnusedDisableDirectives))
}

func (options *LinterOptions) UnmarshalJSON(data []byte) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	if object == nil {
		return fmt.Errorf("linterOptions: expected an object")
	}
	if value, present := object["reportUnusedDisableDirectives"]; present {
		var severity UnusedDirectiveSeverity
		if err := json.Unmarshal(value, &severity); err != nil {
			return err
		}
		options.ReportUnusedDisableDirectives = &severity
	}
	return nil
}
