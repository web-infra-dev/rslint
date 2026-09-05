package padding_around_expect_groups

import rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"

var PaddingAroundExpectGroupsRule = rstestUtils.MakePaddingRule("rstest/padding-around-expect-groups", rstestUtils.PaddingExpect)
