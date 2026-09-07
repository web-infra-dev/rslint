package padding_around_describe_blocks

import rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"

var PaddingAroundDescribeBlocksRule = rstestUtils.MakePaddingRule("rstest/padding-around-describe-blocks", rstestUtils.PaddingDescribe)
