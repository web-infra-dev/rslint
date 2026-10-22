# 修复迁移性能问题后的 Rsbuild 整仓对照

2026-10-22：已修复迁移实现中的重复工作，并重新通过完整工作量校验。**固定 Rsbuild 整仓的完整 lint 耗时由 5.056 s 降至 3.299 s，减少 1.757 s（34.75%，约 1.53 倍速度）；采样进程树峰值 RSS 由 887.16 MiB 降至 596.33 MiB，减少 290.83 MiB（32.78%）。** 原版基线仍保留默认优化和解析缓存。

此前 13.111 s 的退化属于初版迁移实现，不能据此推断 tsrs 本身的性能。新结果仍然是整条 rslint 路径的实测，包含规则移植、编译器、JS 适配与启动关闭，不是 tsrs checker 的独立理论值。

## 新一轮结果

每端每范围预热 2 次，再做 7 次 AB/BA 交错正式测量，共 36 次运行。全部退出 0，stdout/stderr 与正确性证明一致；所有时钟检查通过，没有删除慢样本。下表为 **中位数 ± MAD [最小值, 最大值]**；变化为候选减基线，倍率为候选/基线。

| 范围             | 指标                 | 原版 tsgo                           | Rust/tsrs                           | 变化 / 百分比 / 倍率        |
| ---------------- | -------------------- | ----------------------------------- | ----------------------------------- | --------------------------- |
| 完整 rs lint     | 耗时 (s)             | 5.056 ± 0.074 [4.874, 5.213]        | 3.299 ± 0.157 [2.933, 3.538]        | -1.757 / -34.75% / 0.653×   |
| 完整 rs lint     | 采样树峰值 RSS (MiB) | 887.156 ± 38.781 [793.734, 992.797] | 596.328 ± 16.984 [579.344, 644.625] | -290.828 / -32.78% / 0.672× |
| 独立 rslint 阶段 | 耗时 (s)             | 5.119 ± 0.184 [4.668, 5.416]        | 3.273 ± 0.120 [3.073, 3.493]        | -1.845 / -36.05% / 0.639×   |
| 独立 rslint 阶段 | 采样树峰值 RSS (MiB) | 823.469 ± 17.719 [786.156, 895.859] | 592.484 ± 6.750 [576.078, 609.656]  | -230.984 / -28.05% / 0.719× |

独立 rslint 阶段耗时减少 1.845 s（36.05%，约 1.56 倍速度），RSS 减少 230.98 MiB（28.05%）。两范围的时间和 RSS 最小值至最大值均不重叠；中位数差也超过 `3 × (两端 MAD 之和)` 的启发式门槛，该门槛不是置信区间。完整命令和独立阶段是不同批次的样本，不能相减推导包装层耗时。

## 实际修复

- **跨项目解析与绑定复用缺失**：补齐单次 lint generation 的共享缓存。键包含完整 parse options、script kind 及精确源码内容；每份共享 source 拥有独立 region，所有 Program/checker 释放后才回收，checker 状态仍按 Program 隔离。新增共享、选项隔离、首个 owner 释放、overlay 更新与内存释放回归。
- **文件系统重复探测**：同一 generation 缓存 canonical path、文件/目录探测与成功读入的文本；失败读取仍允许重试。新 generation/fix 清空缓存，避免陈旧文本与永久失败缓存。
- **每节点重复规则分派及全文扫描**：一次构造 AST Kind listener 桶并保持配置顺序；未审计元数据保守访问所有节点。转义相关全文预检移至每文件一次。严格 oracle 曾抓到 `await-thenable` 漏 `CallExpression`、`no-unexpected-multiline` 漏 `BinaryExpression`，已补齐并加正反例，保留失败记录。
- **JS admission 重复路径计算**：完整 V8 profile 中 admission 约 3.199 s，重复 ignore 路径坐标计算是主要开销。改为 invocation 内复用相同 path-relative/within 结果，保留 ignore/negation、配置边界与符号链接的原语义；不跨调用缓存文件系统状态。
- **CLI 重复发现整仓文件**：捕获 source snapshot 和执行阶段原先各遍历一次，现在经同一验证路径只发现一次。保留 lint 前 source overlay，以及写盘 fix 的 canonical path、inode、原始字节防竞态检查。对正常 CLI 只扫描一次及别名/配置变化添加回归。

Profile 用于定位，未混入正式样本；其中 native binding 边界的耗时也包含桥接与结果处理，不能标作纯编译器时间。未通过关闭规则、少检查文件或关闭类型信息制造收益。

## 工作量与功能核验

- Rsbuild `e7d65898ce92d5f5d78afb0d93b4670b5e2a469d`；原版 rslint `e0a29f751d4def86d27931759098f27748fa3003`。保持实际继承配置、overrides、选项和级别：**2394 文件、33 计划、5 配置、129 启用规则、221014 规则文件对、31461 类型规则文件对、1232 类型文件**，含无诊断文件。工作量 manifest SHA256 为 `c09ae5c1d3a77875817f3f1e8dfd35fc788f9921f324699240476f73c9d697f6`。
- 新快照完整 Rstack 命令、独立 CLI、API lint/fix、实际 applied rules/options/severity/type capability 和项目选择对照通过；负向探针仍拒绝错误项目、配置、cwd、overlay 与 compiler pin。Program 全部 source 成员、文本、字节、默认库及 external 标记、cwd 元数据差异为 0。
- 129 规则的实际 options sentinel：258 输入、516 次 fix/非 fix 对照，诊断、消息、位置、建议与 fix 一致。128 条有真实 public ruleId 正命中；`no-octal` 违规输入被双方 parser 先拒绝，单列为 parser 见证，不冒充规则命中。
- Rust 本轮受影响的 **44 项**缓存/生命周期/分发/类型/fix/配置及 await/lexical 测试通过；完整 await oracle 419 输入/838 检查、lexical oracle 490 输入/980 检查通过。最终 JS 修复不改变这些 Rust 源码、编译器资源、构建输入及 addon，按内容 hash 复用该证据，未伪称再次执行。
- 最终 JS 重建后 **97 项消费者测试、Native 8 项、实际 CLI 类型命中/不命中/写盘 fix、Program universe 与无 Go 门禁重新运行通过**。完整 proof 与逐项 hash 绑定的 supplement 通过后才开始测量。没有删断言或忽略原失败。

## 架构、版本与条件

正常运行由 Rust/tsrs lint 核心通过 N-API 直接连接 JS/TS，候选运行 guard 拒绝所有 child_process；未回退 Go/tsgo/桥接/WASM。历史 Go 源码与构建依赖已逐字节归档到仓库外，原版仅在外部冻结目录作为 oracle/A/B。当前源和实际 addon 的无 Go 门禁通过。此前隔离目录的 Rust/Node install/build 证据只适用于其旧快照；没有把它冒充这轮 fresh cold build。

Apple M5 Max 18 核、64 GiB，Darwin 25.6.0，Node 24.20.0；Go 1.27.1 原生优化构建，Rust 1.99.1 release/opt-level 3/LTO/codegen-units 1。tsrs pin `280b426b89cd4ecd6153cf48228826bd60a99c61`；原版 compiler 与匹配声明资源 pin `1f70213d4922b434345f639b441681e470c7cfc1`，兼容 fingerprint `67ddffebeaa754ec2e481caa64501f9c4a8c1ba0924574bc4e664deb0b93957b`。

双方 `--singleThreaded`、`GOMAXPROCS=RAYON_NUM_THREADS=1`，原版 `GOMEMLIMIT=4GiB`，保留原版默认解析缓存。每次新进程，暖 OS/dependency cache，Node compile cache 禁用；Rust request cache 每次重建。完整范围为 Rstack 0.8.4 的 Node 展开命令 `rs lint --singleThreaded --format jsonline`，并单测其 rslint 阶段。排除双方 pnpm launcher，包含 Node、原版 Go 子进程/IPC、候选 N-API 和退出；此命令不包含另行配置的格式或拼写脚本。

全批遵守共享机器锁，无本任务并行重型构建；机器仍有桌面等后台活动，一分钟 load 在样本启动时为 2.75–6.86。前后无已记录 thermal/performance warning；电源日志匹配的 40 行全是 Assertions，包含名为 MaintenanceWake 的 assertion，**没有实际 Sleep/Wake/Thermal 转换**。所有样本独立时钟一致。普通 idle sleep 防止工具未绕过热保护。

RSS 是同次采样中 root 与其全部已知子孙进程 resident bytes 的和，再取整次最大值，不是父进程 RSS 或各进程独立峰值相加。请求间隔 2 ms，本批实际最大采样间隔 **54.35 ms**，最大 wall 观察不确定性 **41.07 ms**。共享页可能重复计数、进程读取非原子、极短子进程及采样间峰值可能遗漏，因此称为采样进程树峰值，不称内核真实整树峰值。全部采样与后台负载记录保留。

## 身份、原始数据与复跑

分支 `perf/tsrs-minimal-poc`，HEAD `dfa3ac4d054678550f4ce5af94fd74188f1d53d8` 上的未提交迁移快照。HEAD 不能单独标识所测代码；Native addon SHA256 为 `4b167a98e3b6162511c49b5dc8a79ebf05965475af00176bb4c07755443ae08c`，完整 measurement identity 为 `66bf0a58f305887df3778d014f23c28ebf5d0fe57c745a8fc73827bac0a35c67`。Rust 产物与 perf-fix-02 相同，JS bundle 和其 source 身份已在这轮重建并重新固定。

[统计与全部紧凑样本](rsbuild-performance-fixed.json)、[36 次原始样本表](rsbuild-performance-fixed.csv)、[工作量清单](rsbuild-workload.json)保留可检查数据。本地产物 `rust-js-migration/perf-fix-03/` 包含 proof、supplement、measure 的完整 stdout/stderr/RSS、进程与负载日志、源/产物 hash、测试日志、两阶段 patch 和复跑脚本；Rust 原始测试/oracle/构建证据保留在 `perf-fix-02/`，JS patch 与结果保留在 `perf-fix-03/`。

在已经准备好的本任务冻结 baseline、Rsbuild、依赖和 evidence bundle 布局中，运行：

```sh
caffeinate -di bash "$ARTIFACTS/rust-js-migration/perf-fix-03/reproduce.sh" \
  "$ARTIFACTS/rust-js-migration/rsbuild-repeat"
```

`ARTIFACTS` 指本任务的 artifacts 目录。输出前缀必须全新且在源码树外；入口先锁定验证当前 JS/CLI/Native/universe/no-Go，Rust 证据只允许在源码/构建输入及实际产物完全相同时复用，再重新 proof、生成 supplement、执行 2+7 A/B。不要给 benchmark 再套同一锁。入口固定本机 darwin-arm64 布局，不能声称仅凭统计 JSON 在任意干净 checkout 一键复现；通用测量驱动为 `scripts/tsrs/rust-rsbuild-benchmark.py --help`。

历史初版 4.891→13.111 s、826.61→1566.70 MiB 的数据继续保留在 [旧原型结果](rsbuild-performance.json)；中间 Rust 修复版 5.122→5.898 s 与最终 JS 修复版属于不同批次，不合并样本、不做逐项精确消融。旧两规则约 46% 收益与此完整工作量没有可比性。

本轮只验收固定 Rsbuild 的规则和所需 CLI/API/config/fix 路径。默认多核、其他平台、LSP、浏览器 WASM、外部 JS 插件及历史全规则入口未因此验收。未声称生产/CI 全绿；本轮定向文档格式检查和 `git diff --check` 通过，性能阶段定向拼写检查曾报告 8 个词典问题；提交准备阶段登记技术名称/测试标记并排除保留原字节的上游数据后，全部显式变更文件拼写检查通过；旧完整 formatting/Clippy 门禁结果不能替代本轮定向功能证据。原性能测量阶段未提交或推送；后续按用户要求更新原 PR2443 的分支、说明和表格评论，保留其关闭状态。原生旧 goal 的 paused/unfinished 条目无法编辑或替换；没有伪标旧 813 规则目标完成，以最新用户范围及持久验收文件为准。

## 提交前验证

19 处已有格式差异已按仓库规范整理；格式化阶段对照确认 13 个 JS/TS 文件结构语法树、5 份 JSON 内容及 Markdown 代码块一致。随后只补充文档说明及词典配置。重新构建 JS 后，**71 项测量产物均逐字节不变**，Rust 产品源码及实际 addon 未改动；Native 8、JS 97、构建/打包 13 项测试与无 Go 门禁重新通过。全仓 `pnpm run format:check`、显式变更文件拼写检查及 `git diff --check` 通过。初始失败记录保留；这些提交前整理没有新增或替换性能样本。见[提交前验证与产物哈希](publication-validation.json)。
