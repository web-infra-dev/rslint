//go:build !darwin && !linux && !windows

package sharedsource

import "errors"

func mapSources(Descriptor) (mappedSources, error) {
	return mappedSources{}, errors.New("shared sources unavailable on this platform")
}
