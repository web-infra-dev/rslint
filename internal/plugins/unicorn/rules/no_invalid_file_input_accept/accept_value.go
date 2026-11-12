package no_invalid_file_input_accept

import (
	"regexp"
	"strings"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const httpToken = "[A-Za-z0-9!#$%&'*+.^_`|~-]+"

var mimeEssence = regexp.MustCompile("^" + httpToken + "/" + httpToken + "$")

var commonMIMETypeMistakes = map[string]string{
	"application/x-rar-compressed": "application/vnd.rar",
	"application/x-zip-compressed": "application/zip",
	"image/jpg":                    "image/jpeg",
	"image/svg":                    "image/svg+xml",
	"image/x-icon":                 "image/vnd.microsoft.icon",
}

func normalizeAccept(value string) (replacement, reason string) {
	if !strings.Contains(value, ",") {
		return normalizeAcceptToken(value)
	}

	var tokens []string
	seen := make(map[string]struct{})
	for raw := range strings.SplitSeq(value, ",") {
		normalized, reason := normalizeAcceptToken(raw)
		if reason != "" {
			return "", reason
		}
		if _, duplicate := seen[normalized]; !duplicate {
			seen[normalized] = struct{}{}
			tokens = append(tokens, normalized)
		}
	}
	return strings.Join(tokens, ", "), ""
}

func normalizeAcceptToken(raw string) (replacement, reason string) {
	token := ecmascript.StringTrim(raw)
	if token == "" {
		return "", "Empty entries are not allowed."
	}
	hasDot := strings.HasPrefix(token, ".")
	if hasDot || !strings.Contains(token, "/") {
		extension := strings.TrimPrefix(token, ".")
		if extension == "" || strings.ContainsAny(extension, ",/*${}") ||
			strings.ContainsFunc(extension, ecmascript.IsWhiteSpaceOrLineTerminator) {
			return "", "File extensions must start with `.` and cannot contain whitespace, commas, slashes, wildcards, or template markers."
		}
		lower := ecmascript.StringToLowerCase(token)
		if hasDot {
			return lower, ""
		}
		return "." + lower, ""
	}
	if strings.Contains(token, "*") {
		lower := ecmascript.StringToLowerCase(token)
		switch lower {
		case "audio/*", "image/*", "video/*":
			return lower, ""
		default:
			return "", "Wildcard MIME types must be `audio/*`, `image/*`, or `video/*`."
		}
	}

	// Upstream uses the WHATWG MIME essence. Parameters may be malformed or
	// repeated without affecting the type/subtype, so only parse that prefix.
	essence, _, _ := strings.Cut(token, ";")
	essence = strings.TrimRight(essence, " \t\r\n")
	if !mimeEssence.MatchString(essence) {
		return "", "MIME types must use valid `type/subtype` syntax."
	}
	essence = ecmascript.StringToLowerCase(essence)
	if corrected, ok := commonMIMETypeMistakes[essence]; ok {
		return corrected, ""
	}
	return essence, ""
}
