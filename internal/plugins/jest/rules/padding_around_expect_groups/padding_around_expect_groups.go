package padding_around_expect_groups

import jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"

var PaddingAroundExpectGroupsRule = jestUtils.MakePaddingRule("jest/padding-around-expect-groups", jestUtils.PaddingExpect)
