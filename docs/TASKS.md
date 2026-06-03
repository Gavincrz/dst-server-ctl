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
- 已补仓库级开发入口：`make dev` 可同时启动 Go 后端和 Vite 前端，`make check` 汇总常用检查。
- 已完成一轮“公开资料 + 当前实现”的轻量参数基线核对，并确认最新安装产物/脚本的实物核对留到后续通过受管安装链路验证。

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
  `world_size`、`branching`、`loop`、`start_location`、`season_start`、`day`、`weather`、`lightning`、`wildfires`、`petrification`、`hounds`、`autumn`、`winter`、`spring`、`summer`、`spawnmode`、`ghostenabled`、`resettime`、`krampus`、`roads`、`touchstone`、`boons`、`cave_ponds`、`earthquakes`、`wormattacks`
- 2026-06-02 已继续补入一批高价值 worldsettings 字段：
  `winterhounds`、`summerhounds`、`wormattacks_boss`、`atriumgate`、`spawnprotection`、`dropeverythingondespawn`、`healthpenalty`、`temperaturedamage`、`hunger`、`darkness`、`specialevent`。
  当前仍沿用现有 `worldGenOverrides` 映射链路，其中 Master 专属 / master-controlled 字段先只在 Master 表单暴露，Caves 专属字段只在 Caves 表单暴露。
- 2026-06-02 已继续补入 `portalresurection` 和 `events` 组节庆开关：
  `crow_carnival`、`hallowed_nights`、`winters_feast`、`year_of_the_gobbler`、`year_of_the_varg`、`year_of_the_pig`、`year_of_the_carrat`、`year_of_the_beefalo`、`year_of_the_catcoon`、`year_of_the_bunnyman`、`year_of_the_dragonfly`、`year_of_the_snake`、`year_of_the_knight`。
  这些字段当前同样先挂在 Master 表单，并继续写回现有 shard `worldGenOverrides`。
- 2026-06-02 已继续补入其余一批 master-controlled global 字段：
  `ghostsanitydrain`、`beefaloheat`。
  当前 `global` / `survivors` / `events` 中已结构化的大部分 master-controlled 字段仍先集中在 Master 表单呈现。
- 2026-06-02 已开始补 `WORLDGEN_GROUP / misc` 字段：
  `task_set`、`cavelight`、`prefabswaps_start`、`terrariumchest`、`stageplays`、`junkyard`。
  其中 `task_set` 目前按 shard 提供不同候选项：Master 使用 forest task set，Caves 使用 cave task set，其余字段按 world 归属分别挂到 Master 或 Caves 表单。
- 2026-06-02 已继续补 `WORLDGEN_GROUP / misc` 中森林专属字段：
  `moon_fissure`、`balatro`。
  其中 `moon_fissure` 使用真实 `worldgen_frequency_descriptions` 值域，而不是复用普通 runtime frequency 枚举。
- 2026-06-02 已先把前端世界字段的语义边界收敛一轮：
  不再仅靠 `masterWorldSettingFields` / `cavesWorldSettingFields` 的排除式过滤维持表单，而是显式区分
  “Master shard worldgen/world rules” 与 “master-controlled survivor/event rules”。
  当前仍保持 API、存储和 writer 兼容，master-controlled 字段继续写入 Master 的 `worldgenoverride.lua`。
- 2026-06-03 已继续把上述边界收敛到后端模型：
  `cluster` API / domain 现已显式增加 cluster 级 `masterWorldSettings`，用于承载
  `master-controlled` worldsettings；SQLite 读写保持对现有表结构兼容，writer 仍在生成
  `Master/worldgenoverride.lua` 时把这组设置并入落盘内容。
  当前前端表单仍沿用现有 `masterWorldSettings` 表单状态，但请求构造时已把
  “Master shard worldgen/world rules” 与 “cluster-wide master-controlled rules” 拆成不同 API 字段。
- 2026-06-03 已继续补一批更贴近客户端语义分组的真实脚本字段：
  - `WORLDGEN_GROUP / misc`：补齐 `balatro`
  - `WORLDSETTINGS_GROUP / misc`：`frograin`、`meteorshowers`、`hunt`、`alternatehunt`、`disease_delay`、`wanderingtrader_enabled`、`acidrain_enabled`
  - `WORLDSETTINGS_GROUP / survivors`：`extrastartingitems`、`seasonalstartingitems`、`lessdamagetaken`、`shadowcreatures`、`brightmarecreatures`
  - `WORLDSETTINGS_GROUP / resources`：`basicresource_regrowth`
  其中 master-controlled 字段现已进入 cluster 级 `masterWorldSettings` 白名单与表单映射，其余字段继续按 shard 写回 `worldgenoverride.lua`。
- 2026-06-03 已继续补一批 shard 级后期世界进度相关字段：
  - `WORLDSETTINGS_GROUP / misc`：`rifts_frequency`、`rifts_enabled`、`lunarhail_frequency`、`rifts_frequency_cave`、`rifts_enabled_cave`
  - `WORLDSETTINGS_GROUP / portal_resources`：`portal_spawnrate`、`bananabush_portalrate`、`lightcrab_portalrate`、`monkeytail_portalrate`、`palmcone_seed_portalrate`、`powder_monkey_portalrate`
  这批字段当前都保持 shard 级建模，继续通过现有 world settings 表单映射回对应 shard 的 `worldgenoverride.lua`。
- 2026-06-03 已基于当前受管安装产物 `buildid 23206748` 继续补入一小批
  `WORLDSETTINGS_GROUP / lunar_mutations` shard 级字段：
  `mutated_hounds`、`penguins_moon`、`moon_spider`、`mutated_birds`、`mutated_merm`、`mutated_spiderqueen`。
  其中 `mutated_hounds` / `penguins_moon` 仅在 Master 表单暴露，其余按真实 world 归属同时支持
  Master / Caves，并继续沿现有 shard `worldGenOverrides` 写回 `worldgenoverride.lua`。
- 仍保留 “extra world overrides” 文本框，用于透传尚未结构化的世界项。
- 2026-06-02 已基于最新受管安装产物中的 `scripts/map/customize.lua` 与 `scripts/worldsettings_overrides.lua` 完成二次实物核对；下一批应优先从真实 `WORLDSETTINGS_GROUP` / `WORLDGEN_GROUP` 中挑高价值字段，而不是继续凭印象补项。

下一步：
在已完成前后端分组收敛的基础上，继续把世界配置模型向“客户端语义分组”推进，优先补完
`lunar_mutations` 余下 gestalt / boss 相关字段，或切到 `giants` 组挑一小批高频 boss 开关继续落地。

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

### T-013 | done | 补齐本地开发统一入口

目标：
给仓库提供统一的本地开发启动命令，减少手动分别启动前后端的摩擦。

完成标准：
- 提供仓库级单命令入口同时启动 Go 后端和前端开发服务器。
- 退出时能一并清理子进程，避免残留后台任务。
- README 记录开发期访问地址和常用检查命令。

实现备注：
- 保持依赖简单，优先使用 `Makefile` + shell 脚本，不为开发编排额外引入 Node 或其他进程管理依赖。

下一步：
确认是否需要继续补 `make test`、发布构建入口或前端嵌入后的统一产物流程。

### T-014 | done | 先做轻量配置参数基线核对

目标：
在不额外下载安装最新 DST 的前提下，先用公开资料和当前实现收敛配置边界，避免后续继续盲目补字段。

完成标准：
- 明确当前控制器已经覆盖的 `cluster.ini`、`server.ini`、`worldgenoverride.lua` 主链路范围。
- 记录高优先级缺口，例如 token、权限列表、模组文件与 `leveldataoverride.lua` 兼容导出。
- 明确这仍不是“按最新安装产物/脚本逐项核实完成”的结论，并把实物核对留给下一步。

实现备注：
- 本轮不修改用户现有手工 DST 部署。
- 轻量核对结论记录在 `docs/CONFIGURATION.md`，作为 `T-009` / `T-010` 的前置边界。

下一步：
通过 Web 前端触发受管 DST 安装或更新，在 managed root 下拿最新受管副本做端到端链路验证，并补“最新安装产物/脚本实物核对”。

### T-017 | done | 基于最新受管安装产物做二次实物核对

目标：
在受管 DST 已实际安装完成后，基于最新安装产物和真实脚本核对配置参数边界，收敛 `T-009` / `T-010` 的下一批实现目标。

完成标准：
- 通过 Web 前端完成受管安装与版本检查链路验证。
- 基于 `scripts.zip`、生成产物和受管安装目录确认当前主链路覆盖范围。
- 明确一批适合优先结构化支持的高价值字段，而不是继续盲目补项。

实现备注：
- 本轮已核对 `scripts/map/customize.lua`、`scripts/worldsettings_overrides.lua`、`scripts/languages/language.lua` 以及受管生成的 `cluster.ini` / `server.ini` / `worldgenoverride.lua`。
- 结论已回填到 `docs/CONFIGURATION.md`，并确认下一批世界字段应优先来自真实 `WORLDSETTINGS_GROUP` / `WORLDGEN_GROUP`。

下一步：
回到 `T-010` 和 `T-009`，按这次实物核对结果实现一批高价值世界字段与服务器关键文件边界。

### T-018 | done | 固化配置字段提取流程

目标：
把“从最新受管 DST 安装产物提取配置字段”的过程沉淀成可复用脚本和文档，避免每次手工翻 `scripts.zip`。

完成标准：
- 提供可直接运行的提取脚本。
- 文档说明数据来源、使用方式、输出含义和局限。
- 让后续版本核对能通过固定流程重复执行。

实现备注：
- 当前脚本从受管 `appmanifest_343050.acf`、`scripts.zip`、`scripts/map/customize.lua` 和语言文件中提取字段清单。
- 该流程用于生成候选库存，不自动修改控制器结构化模型。

下一步：
在每次 DST 更新或安装新版本后重新运行字段提取，再据结果选择下一批要结构化支持的字段。

### T-015 | todo | 重做 Web UI 的信息架构与页面分层

目标：
解决当前 Web UI 把初始化、更新、运行态、日志和配置都堆在同一页面上的问题，重新设计更清晰的导航、页面拆分和操作流。

完成标准：
- 明确首版 Web UI 的页面结构、导航层级和主操作路径。
- 把初始化、更新、运行控制、日志查看、cluster 配置等高频任务拆到更清晰的页面或分区，而不是继续堆叠在单页中。
- 新设计在不牺牲当前功能覆盖的前提下，降低首次使用和日常维护时的理解成本。

实现备注：
- 这是独立体验改造任务，暂不打断当前 DST 安装验证和配置参数主线。
- 重设计前应先结合真实使用流程，记录当前单页在发现性、反馈和状态切换上的具体问题点。

下一步：
先完成受管安装/更新端到端验证与最新安装产物核对，再回头整理 UI 重设计范围和原型方向。

### T-016 | done | 修复更新检查读取本地版本路径错误

目标：
修复 `Check Now` 在 DST 已安装后仍然失败的问题，确保更新检查从受管 DST 安装目录读取本地 build id。

完成标准：
- 本地版本读取路径指向受管 `dst/steamapps/appmanifest_343050.acf`。
- 更新检查不再错误读取 `steamcmd/steamapps`。
- 补上覆盖该路径选择的最小测试。

实现备注：
- 这是安装验证阶段暴露出的链路 bug，和“继续扩展配置参数”主线并行修复。

下一步：
继续执行网页端安装/更新端到端验证，并据最新受管安装产物核对参数面。

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
- 当前“最新 DST 参数面已逐项核对”的结论尚未成立；还需要在 managed root 下安装或更新最新受管 DST，并通过前端链路完成端到端验证后再补实物核对。
