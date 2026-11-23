//go:build unix || js || wasip1

package main

import "syscall"

// Stat_t.Dev is signed on Darwin and unsigned on Linux.
type deviceID uint64

type fixFileID struct {
	device deviceID
	inode  uint64
}

func statFixFile(path string) (fixFileID, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		return fixFileID{}, err
	}
	return fixFileID{deviceID(stat.Dev), stat.Ino}, nil
}
