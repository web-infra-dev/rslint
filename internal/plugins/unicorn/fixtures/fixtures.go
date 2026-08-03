package fixtures

import (
	"embed"

	"github.com/web-infra-dev/rslint/internal/testutil/embedfs"
)

//go:embed *
var files embed.FS

func GetRootDir() embedfs.Root {
	return embedfs.EmbedRoot(files)
}
