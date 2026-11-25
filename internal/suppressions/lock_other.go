//go:build (!unix || aix) && !windows

package suppressions

import (
	"context"
	"os"
)

func lockBaselineUpdate(ctx context.Context, _ string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return func() {}, nil
}

func replaceBaselineFile(_ context.Context, tempName, path string) error {
	return os.Rename(tempName, path)
}
