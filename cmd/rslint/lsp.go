// modified based on https://github.com/microsoft/typescript-go/blob/cedc0cbe6c188f9bfe6a51af00c79be48c9ab74d/cmd/tsgo/lsp.go#L1
package main

import (
	"os"
	"runtime"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/lsp"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func runLSP(args []string) int {

	fs := bundled.WrapFS(osvfs.FS())
	defaultLibraryPath := bundled.LibPath()
	typingsLocation := getGlobalTypingsCacheLocation()

	s := lsp.NewServer(&lsp.ServerOptions{
		In:                 lsp.ToReader(os.Stdin),
		Out:                lsp.ToWriter(os.Stdout),
		Err:                os.Stderr,
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
