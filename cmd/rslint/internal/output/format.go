package output

import "fmt"

// Format identifies one CLI stdout protocol. Its zero value is the default
// human-readable format so internal callers that construct options directly
// retain the CLI default.
type Format uint8

// Keep this list in sync with OUTPUT_FORMATS in
// packages/rslint/src/utils/args.ts.
const (
	FormatDefault Format = iota
	FormatJSONLine
	FormatGitHub
	FormatGitLab
)

func ParseFormat(value string) (Format, error) {
	switch value {
	case "default":
		return FormatDefault, nil
	case "jsonline":
		return FormatJSONLine, nil
	case "github":
		return FormatGitHub, nil
	case "gitlab":
		return FormatGitLab, nil
	default:
		return FormatDefault, fmt.Errorf("invalid output format %q (expected default, jsonline, github, or gitlab)", value)
	}
}

func (f Format) String() string {
	switch f {
	case FormatDefault:
		return "default"
	case FormatJSONLine:
		return "jsonline"
	case FormatGitHub:
		return "github"
	case FormatGitLab:
		return "gitlab"
	default:
		return fmt.Sprintf("Format(%d)", f)
	}
}
