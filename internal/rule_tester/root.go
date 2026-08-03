package rule_tester

import "github.com/web-infra-dev/rslint/internal/testutil/embedfs"

// Root identifies where RunRuleTester and NewProgramHelper read fixture
// files from.
type Root = embedfs.Root
