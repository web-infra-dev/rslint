package utils

import (
	"encoding/json"

	"github.com/tailscale/hujson"
)

// ParseJSONC parses JSONC (JSON with Comments) using hujson library
func ParseJSONC(data []byte, v interface{}) error {
	// Parse with hujson first to handle comments and trailing commas
	ast, err := hujson.Parse(data)
	if err != nil {
		return err
	}

	// Standardize to valid JSON
	ast.Standardize()

	// Convert back to bytes and parse with standard JSON
	standardJSON := ast.Pack()
	return json.Unmarshal(standardJSON, v)
}

// StripJSONComments removes comments from JSONC and returns clean JSON string
func StripJSONComments(jsoncString string) string {
	ast, err := hujson.Parse([]byte(jsoncString))
	if err != nil {
		return jsoncString // Return original on error
	}

	ast.Standardize()
	return string(ast.Pack())
}
