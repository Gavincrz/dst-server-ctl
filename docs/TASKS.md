# 任务

规则：

- 每次只处理一个任务。
- 任意时刻只能有一个任务是 `doing`。
- 如果当前任务跨度过大，先拆成更小的可交付任务，再开始其中一个。

## 当前状态

项目当前已经具备：

- harness、Go 后端骨架、Svelte 前端骨架和基础工程流程。
- managed root 路径布局、共享 command runner、SQLite 状态存储基础层和启动时 managed root 初始化。
- 安装状态 API、安装任务 API、任务模型，以及由任务驱动的 SteamCMD / DST 安装执行流程。
- 初始化状态页、运行态状态页、运行控制入口，以及安装/更新任务和 shard 日志查看能力。
- 受管 cluster 的结构化配置状态、`GET/PUT /api/v1/cluster` 读写 API，以及由该状态生成的 `cluster.ini`、`server.ini` 与 `worldgenoverride.lua`。
- 扩展后的 cluster/shared 与 shard/network 结构化参数、每 shard 的 world preset + overrides 状态、对应的 Web UI 表单和前端测试。
- 基于 managed root `clusters/primary` 布局的 DST shard 启动命令生成、运行时启动 service、`/api/v1/runtime` 系列接口和基础运行历史。
- 安装/更新/版本检查/runtime 日志 SSE、dashboard 汇总 SSE、按 offset 增量读取的日志推流，以及前端断线/回退测试。
- 带本地/远端版本比较、手动检查、手动更新、运行中更新保护和日志落盘/排查入口的 DST 更新流程。
- Master / Caves 世界配置表单已补入一批高频世界生成 / 世界规则字段，并继续保留 raw overrides 兜底。

## 任务列表

### T-001 | done | 初始化仓库与基础工程骨架

目标：
建立 git 仓库、harness 文档、Go 后端骨架和 Svelte 前端骨架。

完成标准：
- git 仓库已初始化。
- harness 文档已添加。
- Go 和 Svelte 工程骨架可用。

下一步：
补齐路径布局、command runner 与基础测试。

### T-002 | done | 打通基础运行支撑

目标：
建立 managed root、共享 command runner、SQLite migration 和安装状态存储基础层。

完成标准：
- managed root 路径布局已落地。
- command runner 已提供基础测试。
- SQLite migration 与安装状态仓储接口可用。

下一步：
定义安装状态 API 和受管安装执行链路。

### T-003 | done | 打通受管安装与初始化 UI

目标：
让控制器可先启动，再通过 Web UI 驱动受管安装。

完成标准：
- 安装状态 API、安装任务 API 和任务模型已接通。
- SteamCMD / DST 安装流程可由后端任务执行。
- 状态页已接入 `/api/v1/status` 与 `/api/v1/installation`。

下一步：
补齐安装任务列表、轮询和错误展示细节。

### T-004 | done | 打通 cluster 基础配置链路

目标：
建立受管 cluster 的结构化配置状态、API、文件生成和基础 UI。

完成标准：
- `GET/PUT /api/v1/cluster` 可读写结构化配置。
- `cluster.ini` 和 shard `server.ini` 可从状态生成。
- Web UI 可编辑、保存和重置基础 cluster 配置。

下一步：
把生成后的 cluster 目录接入启动流程。

### T-005 | done | 打通运行态控制与日志查看

目标：
让控制器可启动/停止/重启受管 shard，并提供基础日志与错误反馈。

完成标准：
- `runtime` 系列 API 和运行态控制面板可用。
- Master / Caves 日志可落盘并查看。
- 进程退出状态、自动清理和基础错误展示已接通。

下一步：
补齐运行历史、自动重试和配置变更后的重启判定。

### T-006 | done | 收敛日志与状态刷新体验

目标：
减少日志和 dashboard 相关的重复轮询与前端重复实现。

完成标准：
- 安装/更新/runtime/版本检查日志已统一到共享日志面板抽象。
- 日志流已迁移到 SSE，并为不支持 EventSource 的场景保留回退。
- dashboard 汇总 SSE 和分级状态轮询已生效。

下一步：
评估日志 tail 与 dashboard 事件边界，避免过早增加复杂度。

### T-007 | done | 梳理 DST 配置分层并打通世界配置主链路

目标：
明确 DST 配置边界，并先打通 `worldgenoverride.lua` 的结构化生成链路。

完成标准：
- `docs/CONFIGURATION.md` 已记录 cluster、shard、世界和模组配置边界。
- cluster/shared 与 shard/network 结构化参数已扩展到主要 `cluster.ini` / `server.ini` 范围。
- 每 shard 的 world preset + overrides 已持久化、生成并接入 Web UI。

下一步：
继续扩展世界配置字段覆盖面，并保持“最终全量参数可配置”的目标。

### T-008 | done | 规范仓库治理文档护栏

目标：
把仓库协作规则、任务文档、架构文档和决策文档整理成更稳定的 agent 友好形态。

完成标准：
- `AGENTS.md` 已补充代码健康和小步重构规则。
- `docs/TASKS.md` 已改为固定任务模板，并反映完成内容、下一步和阻塞点。
- `docs/ARCHITECTURE.md` 已补充外部接口与测试切入点。
- `docs/DECISIONS.md` 已采用稳定的“id + date + 决定 + 原因”格式。

下一步：
回到服务器配置、世界配置和 mod 管理主线。

### T-009 | todo | 继续扩展服务器配置的全量参数覆盖

目标：
在服务器配置实现上以“最终全量参数可配置”为目标推进；首轮可以先挑少量参数验证链路，但不能把当前少量字段当成最终范围。

完成标准：
- 明确 `cluster.ini`、`server.ini` 仍缺失的高价值参数范围。
- 新增一批结构化字段，并接通 API、存储、writer 和基础 UI。
- 不把当前已支持字段误当成最终配置面。

实现备注：
- 优先沿现有结构化配置链路扩展，不要绕回原地编辑 DST 文本文件。
- 如果发现字段边界不清楚，先补文档或记录未决项，再继续实现。

下一步：
回到 cluster shared secrets/lists 和其余高价值 `cluster.ini` / `server.ini` 参数。

### T-010 | doing | 扩展世界配置字段覆盖面

目标：
在已支持 preset + overrides 的前提下，继续把常用世界参数做成结构化、可发现的表单项。

完成标准：
- 补齐一批高频世界配置项的结构化状态。
- 文件生成、持久化和 Web UI 保持一致。
- 对仍不适合结构化的项保留 overrides 兜底。

实现备注：
- 继续以 `worldgenoverride.lua` 为主源，不提前切回 `leveldataoverride.lua`。
- 本轮已把下列字段接成结构化表单，并继续映射回 shard overrides：
  `world_size`、`branching`、`loop`、`start_location`、`season_start`、`day`、`weather`、`autumn`、`winter`、`spring`、`summer`、`roads`、`touchstone`、`boons`、`cave_ponds`、`wormattacks`
- 仍保留 “extra world overrides” 文本框，用于透传尚未结构化的世界项。

下一步：
继续补资源、生物、事件类高频世界项，并评估是否需要把世界字段从通用 overrides map 进一步演进成独立子模型。

### T-011 | todo | 核实语言配置边界

目标：
先以英语为默认和首个完整支持目标，并确认语言切换所需的可靠来源。

完成标准：
- 明确首版是否只提供英语默认值还是允许有限语言切换。
- 若缺少 Klei 官方维护的完整语言代码对照表，记录为待确认项而不阻塞配置主线。

实现备注：
- 不为了语言问题打断服务器配置与世界配置主线。

下一步：
继续规划 mod 管理的接入范围。

### T-012 | todo | 延后低优先级体验优化

目标：
把实时刷新、dashboard 事件细化和其他体验优化降级优先级，避免分散当前主线。

完成标准：
- 当前主线优先级明确回到服务器配置、世界配置和 mod 管理。
- 新的体验优化需求先进入任务列表，不直接挤占主线任务。

下一步：
在配置主线收敛后再回头挑选高价值体验优化。

## 暂时不要做

- 不要添加 Docker 支持。
- 不要导入、迁移或修改 `/home/dontstarve/dst-server` 下的手动 DST 部署。
- 在 managed install、状态存储、进程生命周期完成前，不要提前做完整模组管理 UI。
- 默认不要把 Web UI 暴露到公网网卡。

## 最近完成检查

- [x] 已运行相关后端/前端检查。
- [x] 本次未改 Go 文件，无需运行 `gofmt`。
- [x] 若边界或技术决策变化，已更新架构或决策文档。
- [x] 本文件已反映完成内容和下一任务。
- [x] 本次改动未提交 commit。

## 未决问题

- 第一版公开发布时，世界配置是否只写 `worldgenoverride.lua`，还是还要补 `leveldataoverride.lua` 兼容导出。
- 启动流程接入后，cluster 配置变更与运行中 shard 的重载策略要不要区分“需重启”与“即时生效”。
- 当前日志 SSE 已改成按秒检查文件并按 offset 读取新增内容；当前结论是暂不升级为 file watcher，但如果后续日志连接数、空转 IO 或实时性要求明显提高，再重新评估。
- 当前 `restartRequired` 只基于 cluster 配置快照；后续如果 token、admin 列表、模组或世界设置接入运行态，也要纳入重启判定。
- 当前自动重试策略只做每个 shard 一次立即重试，没有退避、上限策略或外部告警通道。
- 当前 dashboard SSE 仍以整包 snapshot 为主；虽然已经明显减少页面轮询，但如果后续要进一步细化，应该优先事件化 runtime history 和任务状态，而不是把临时成功提示词塞进后端协议。
- 当前还没有确认 Klei 官方维护的完整 `cluster_language` 代码对照表；语言切换能力后续需要单独核实来源，现阶段不阻塞英语配置支持。
