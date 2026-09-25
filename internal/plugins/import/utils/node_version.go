package utils

import (
	"errors"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/semver"
)

// This repository-authored pattern validates a version, not a user regexp.
var nodeVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
var defaultNodeVersion = semver.MustParse("22.0.0")

// NodeVersion reads import/node-version independently of module resolution settings.
// Native linting has no Node runtime; use a stable Node 22 baseline when omitted.
func NodeVersion(settings map[string]any) (semver.Version, error) {
	raw, configured := settings["import/node-version"]
	if !configured {
		return defaultNodeVersion, nil
	}
	text, ok := raw.(string)
	if ok && nodeVersionPattern.MatchString(text) {
		if version, err := semver.TryParseVersion(text); err == nil {
			return version, nil
		}
		// Upstream accepts leading zeroes despite its error message.
		parts := strings.Split(text, ".")
		for i, part := range parts {
			parts[i] = strings.TrimLeft(part, "0")
			if parts[i] == "" {
				parts[i] = "0"
			}
		}
		if version, err := semver.TryParseVersion(strings.Join(parts, ".")); err == nil {
			return version, nil
		}
	}
	return semver.Version{}, errors.New("`import/node-version` setting must be a string in the format \"10.23.45\" (a semver version, with no leading zero)")
}
