//go:build !js && !darwin && !linux && !windows

package main

import "errors"

func mapPluginSources(pluginSourceMapping) ([]byte, func() error, error) {
	return nil, nil, errors.New("shared sources unavailable on this platform")
}
