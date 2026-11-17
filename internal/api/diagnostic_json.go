package api

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Diagnostic text may contain compiler WTF-8 for a lone JavaScript surrogate.
// encoding/json replaces those bytes. Adapt only responses containing such
// text, at the API boundary; the generic IPC transport remains task-agnostic.
func lintResponseForWire(response *LintResponse) any {
	if response == nil {
		return response
	}
	needsAdaptation := false
	for _, diagnostic := range response.Diagnostics {
		if !utf8.ValidString(diagnostic.Message) {
			needsAdaptation = true
			break
		}
		for _, suggestion := range diagnostic.Suggestions {
			if !utf8.ValidString(suggestion.Message) {
				needsAdaptation = true
				break
			}
		}
		if needsAdaptation {
			break
		}
	}
	if !needsAdaptation {
		return response
	}
	wrapped := make([]wireDiagnostic, len(response.Diagnostics))
	for i, diagnostic := range response.Diagnostics {
		wrapped[i] = wireDiagnostic{Diagnostic: diagnostic, Message: jsonMessage(diagnostic.Message)}
		if diagnostic.Suggestions != nil {
			wrapped[i].Suggestions = make([]wireSuggestion, len(diagnostic.Suggestions))
			for j, suggestion := range diagnostic.Suggestions {
				wrapped[i].Suggestions[j] = wireSuggestion{Suggestion: suggestion, Message: jsonMessage(suggestion.Message)}
			}
		}
	}
	return struct {
		*LintResponse
		Diagnostics []wireDiagnostic `json:"diagnostics"`
	}{response, wrapped}
}

type wireDiagnostic struct {
	Diagnostic
	Message     jsonMessage      `json:"message"`
	Suggestions []wireSuggestion `json:"suggestions,omitempty"`
}

type wireSuggestion struct {
	Suggestion
	Message jsonMessage `json:"message"`
}

// jsonMessage preserves UTF-16 values while producing ordinary JSON text.
// HTML escaping stays disabled, as in the API's IPC encoder. Other fields and
// the ordinary UTF-8 response path keep their existing serialization.
type jsonMessage string

func (message jsonMessage) MarshalJSON() ([]byte, error) {
	text := string(message)
	var out bytes.Buffer
	out.Grow(len(text) + 2)
	out.WriteByte('"')
	for i := 0; i < len(text); {
		ch, size := ecmascript.DecodeStringRune(text[i:])
		i += size
		switch ch {
		case '"', '\\':
			out.WriteByte('\\')
			out.WriteRune(ch)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if ch < 0x20 || ch >= 0xd800 && ch <= 0xdfff || ch == 0x2028 || ch == 0x2029 {
				fmt.Fprintf(&out, `\u%04x`, ch)
			} else {
				out.WriteRune(ch)
			}
		}
	}
	out.WriteByte('"')
	return out.Bytes(), nil
}
