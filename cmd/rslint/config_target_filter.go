package main

import (
	"fmt"

	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

// projectDiscoveredTargets converts discovery's already-final ConfigArray
// decisions into the shared lint-target plan shape. It deliberately has no config,
// filesystem, or matcher dependency: a missing merged result is an invariant
// violation instead of permission to perform a second selection.
func projectDiscoveredTargets(targets []discovery.DiscoveredTarget) ([]string, []discovery.DiscoveredTarget, error) {
	if len(targets) == 0 {
		return nil, nil, nil
	}

	allowedFiles := make([]string, 0, len(targets))
	for _, target := range targets {
		if target.MergedConfig == nil {
			return nil, nil, fmt.Errorf("discovered target %q has no merged config", target.Path)
		}
		allowedFiles = append(allowedFiles, target.Path)
	}
	return allowedFiles, append([]discovery.DiscoveredTarget(nil), targets...), nil
}
