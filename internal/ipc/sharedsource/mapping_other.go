//go:build !darwin && !linux && !windows

package sharedsource

import "errors"

func mapSources(Descriptor) ([]byte, func() error, error) {
	return nil, nil, errors.New("shared sources unavailable on this platform")
}
