// modified based on https://github.com/microsoft/typescript-go/blob/cedc0cbe6c188f9bfe6a51af00c79be48c9ab74d/cmd/tsgo/lsp.go#L1
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/lsp"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func runLSP(args []string) int {

	fs := bundled.WrapFS(osvfs.FS())
	defaultLibraryPath := bundled.LibPath()
	typingsLocation := getGlobalTypingsCacheLocation()

	var runtimeRequest lsp.RuntimeRequestFunc
	var runtimeChannel *ipc.Channel
	var runtimePnpPath string
	if len(args) > 0 && args[0] == "--runtime-ipc" {
		// The core sidecar creates a one-shot authenticated loopback listener.
		// A TCP stream is used instead of inherited fd 3/4 because Go's Windows
		// os.NewFile expects a native HANDLE, while Node exposes only a CRT-style
		// stdio index to non-Node children.
		runtimeConn, err := connectEditorRuntime()
		if err != nil {
			fmt.Fprintf(os.Stderr, "rslint: connect editor runtime: %v\n", err)
			return 1
		}
		runtimeChannel = ipc.NewChannel(runtimeConn, runtimeConn)
		runtimePnpPath = os.Getenv("RSLINT_RUNTIME_PNP_PATH")
		runtimeChannel.Start()
		runtimeRequest = func(ctx context.Context, method string, params any) (any, error) {
			kind := ipc.MessageKind(strings.TrimPrefix(method, "rslint/"))
			response, err := runtimeChannel.SendRequest(ctx, kind, params)
			if err != nil {
				return nil, err
			}
			var result any
			if err := response.Decode(&result); err != nil {
				return nil, err
			}
			return result, nil
		}
		defer runtimeChannel.Close()
	}

	s := lsp.NewServer(&lsp.ServerOptions{
		In:             lsp.ToReader(os.Stdin),
		Out:            lsp.ToWriter(os.Stdout),
		Err:            os.Stderr,
		RuntimeRequest: runtimeRequest,
		RuntimePnpPath: runtimePnpPath,
		RuntimeDone: func() <-chan struct{} {
			if runtimeChannel == nil {
				return nil
			}
			return runtimeChannel.Done()
		}(),
		Cwd:                utils.Must(os.Getwd()),
		FS:                 fs,
		DefaultLibraryPath: defaultLibraryPath,
		TypingsLocation:    typingsLocation,
	})

	if err := s.Run(); err != nil {
		return 1
	}
	return 0
}

var runtimeTokenPattern = regexp.MustCompile("^[0-9a-f]{64}$")

func connectEditorRuntime() (net.Conn, error) {
	address := os.Getenv("RSLINT_RUNTIME_IPC_ADDRESS")
	token := os.Getenv("RSLINT_RUNTIME_IPC_TOKEN")
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return nil, fmt.Errorf("runtime address must be numeric loopback, got %q", host)
	}
	if !runtimeTokenPattern.MatchString(token) {
		return nil, errors.New("runtime token is missing or malformed")
	}
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return nil, err
	}
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if _, err := io.WriteString(conn, token+"\n"); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("write authentication token: %w", err)
	}
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func getGlobalTypingsCacheLocation() string {
	switch runtime.GOOS {
	case "windows":
		return tspath.CombinePaths(tspath.CombinePaths(getWindowsCacheLocation(), "Microsoft/TypeScript"), core.VersionMajorMinor())
	case "openbsd", "freebsd", "netbsd", "darwin", "linux", "android":
		return tspath.CombinePaths(tspath.CombinePaths(getNonWindowsCacheLocation(), "typescript"), core.VersionMajorMinor())
	default:
		panic("unsupported platform: " + runtime.GOOS)
	}
}

func getWindowsCacheLocation() string {
	basePath, err := os.UserCacheDir()
	if err != nil {
		if basePath, err = os.UserConfigDir(); err != nil {
			if basePath, err = os.UserHomeDir(); err != nil {
				if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
					basePath = userProfile
				} else if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); homeDrive != "" && homePath != "" {
					basePath = homeDrive + homePath
				} else {
					basePath = os.TempDir()
				}
			}
		}
	}
	return basePath
}

func getNonWindowsCacheLocation() string {
	if xdgCacheHome := os.Getenv("XDG_CACHE_HOME"); xdgCacheHome != "" {
		return xdgCacheHome
	}
	const platformIsDarwin = runtime.GOOS == "darwin"
	var usersDir string
	if platformIsDarwin {
		usersDir = "Users"
	} else {
		usersDir = "home"
	}
	homePath, err := os.UserHomeDir()
	if err != nil {
		if home := os.Getenv("HOME"); home != "" {
			homePath = home
		} else {
			var userName string
			if logName := os.Getenv("LOGNAME"); logName != "" {
				userName = logName
			} else if user := os.Getenv("USER"); user != "" {
				userName = user
			}
			if userName != "" {
				homePath = "/" + usersDir + "/" + userName
			} else {
				homePath = os.TempDir()
			}
		}
	}
	var cacheFolder string
	if platformIsDarwin {
		cacheFolder = "Library/Caches"
	} else {
		cacheFolder = ".cache"
	}
	return tspath.CombinePaths(homePath, cacheFolder)
}
