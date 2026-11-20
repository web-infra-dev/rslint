//go:build !darwin

package main

import "os"

func getCLIWorkingDirectory() (string, error) {
	return os.Getwd()
}
