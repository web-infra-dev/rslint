//go:build !darwin && !linux && !windows

package ipc

import "errors"

func mapSourceMapping(SourceDescriptor) (sourceMapping, error) {
	return sourceMapping{}, errors.New("shared sources unavailable on this platform")
}
