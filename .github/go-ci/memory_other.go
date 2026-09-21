//go:build !windows

package main

import "errors"

func availableCommitMiB() (int64, error) {
	return 0, errors.New("Windows commit accounting is required when memory_Limit is set")
}

// Allow wrapper integration tests without memory admission on other platforms.
func trackTree() (func() (usage, error), error) {
	return func() (usage, error) { return usage{}, nil }, nil
}
