//go:build darwin || linux

package ipc

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestMemoryBootstrapTransfersAndClosesDescriptors(t *testing.T) {
	for _, variant := range []string{"valid", "missing-fd", "extra-fd", "short-mapping"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			bootstrap, err := listenMemoryBootstrap(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer bootstrap.close()
			file, err := os.CreateTemp(t.TempDir(), "mapping")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			if variant != "short-mapping" {
				if err := file.Truncate(MemoryCapacity); err != nil {
					t.Fatal(err)
				}
				if _, err := file.WriteAt([]byte("shared bytes"), MemoryHeaderSize); err != nil {
					t.Fatal(err)
				}
			}
			conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: bootstrap.path(), Net: "unix"})
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			fds := []int{int(file.Fd())}
			if variant == "missing-fd" {
				fds = nil
			}
			if variant == "extra-fd" {
				fds = append(fds, int(file.Fd()))
			}
			if _, _, err := conn.WriteMsgUnix([]byte{0}, unix.UnixRights(fds...), nil); err != nil {
				t.Fatal(err)
			}
			mapping, err := bootstrap.receive(ctx, MemoryMapping{Version: MemoryVersion})
			if variant != "valid" {
				if err == nil {
					_ = mapping.unmap()
					t.Fatal("invalid transfer accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer mapping.unmap()
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(file.Name()); err != nil {
				t.Fatal(err)
			}
			if got := string(mapping.data[MemoryHeaderSize : MemoryHeaderSize+12]); got != "shared bytes" {
				t.Fatalf("received mapping changed after sender closed: %q", got)
			}
			bootstrap.close()
			if _, err := os.Stat(bootstrap.path()); !os.IsNotExist(err) {
				t.Fatalf("socket remains: %v", err)
			}
		})
	}
}

func TestMemoryBootstrapCancellationClosesWaitingSocket(t *testing.T) {
	for _, connected := range []bool{false, true} {
		t.Run(map[bool]string{false: "accept", true: "receive"}[connected], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bootstrap, err := listenMemoryBootstrap(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer bootstrap.close()
			if connected {
				conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: bootstrap.path(), Net: "unix"})
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
			}
			done := make(chan error, 1)
			go func() { _, err := bootstrap.receive(ctx, MemoryMapping{Version: MemoryVersion}); done <- err }()
			cancel()
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("cancelled bootstrap accepted storage")
				}
			case <-time.After(time.Second):
				t.Fatal("cancelled bootstrap is still waiting")
			}
		})
	}
}
