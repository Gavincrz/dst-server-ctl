# 任务

## 当前状态

- [x] 初始化 git 仓库。
- [x] 添加 harness 文档。
- [x] 添加 Go 后端骨架。
- [x] 添加 Svelte 前端骨架。
- [x] 添加路径布局和 command runner 的基础测试。
- [x] 配置本机 Go 工具链，并验证后端测试。
- [x] 添加 SQLite migration 层和安装状态仓储接口。
- [x] 定义安装状态 API 和 managed root 初始化流程。
- [x] 实现 SteamCMD 安装计划和任务模型。
- [x] 添加第一个 Svelte 状态页，并接入后端 `/api/v1/status` 和 `/api/v1/installation`。
- [x] 添加安装任务 API，并把任务模型接入 SteamCMD 安装执行流程。
- [x] 在 Svelte 状态页显示安装任务列表和安装操作入口。
- [x] 为安装任务状态增加前端轮询和错误展示细节。
- [x] 定义受管 cluster 的基础结构化配置状态和读写 API。
- [x] 为 cluster 配置生成基础 `cluster.ini` 和 shard `server.ini` 输出。
- [x] 把 cluster 配置 API 接入 Web UI 表单和保存流程。
- [x] 为受管 DST 启动流程接入生成后的 cluster 配置目录和 shard 布局。
- [x] 在启动流程接通后补充运行态状态页和基础进程控制入口。
- [x] 为 Master 和 Caves 补充日志流读取与展示。
- [x] 为运行中进程补充退出状态跟踪、自动清理和更细的错误呈现。
- [x] 在运行态稳定后补充重启入口，并区分配置变更是否需要重启。
- [x] 持久化更细的运行历史，并为意外退出补重试/告警策略。
- [x] 修复安装任务轮询期间的 SQLite 并发锁问题，并按真实 DST Linux 64 位产物修正 shard 启动二进制路径。
- [x] 为安装任务补充 Web UI 日志查看入口，和更新流程的排查体验保持一致。
- [x] 为安装任务和更新任务的日志展开态增加自动刷新，减少排查时手动点击 Refresh Logs。
- [x] 为任务日志自动刷新补充活跃任务过滤，避免已结束任务面板持续重复请求。
- [x] 为安装任务、更新任务和版本检查日志面板统一抽象刷新逻辑，减少前端重复状态和请求代码。
- [x] 把 runtime shard 日志面板也并入统一的前端日志面板抽象，减少日志 UI 的分散实现。
- [x] 为 runtime shard 日志增加基于 SSE 的持续推流接口，并让前端日志面板在展开时改用 EventSource 订阅增量输出。
- [x] 把安装任务、更新任务和版本检查日志也迁移到 SSE，并为不支持 EventSource 的场景保留轮询回退。
- [x] 为日志 SSE 补充前端交互测试和断线/回退场景覆盖，防止后续重构破坏连接管理。
- [x] 把日志 SSE 的服务端读取改成基于文件 offset 的增量读取，避免连接期间每秒重读整段最近日志窗口。

当前项目已有 harness、工程骨架、managed root 路径布局、共享 command runner、SQLite 状态存储基础层、启动时 managed root 初始化、安装状态 API、安装任务 API、任务模型、由任务驱动的 SteamCMD/DST 安装执行流程、初始化状态页、可反映控制器启动时间的基础运行状态、受管 cluster 的结构化配置状态和 `GET/PUT /api/v1/cluster` 读写 API、由该状态生成的基础 `cluster.ini` 与 shard `server.ini` 文件输出、接入 Web UI 的 cluster 配置表单/保存/重置和前端测试、基于 managed root `clusters/primary` 布局的 DST shard 启动命令生成与运行时启动 service、`GET /api/v1/runtime`、`POST /api/v1/runtime/start`、`POST /api/v1/runtime/stop`、`POST /api/v1/runtime/restart` 与对应的运行态 Web 控制面板、写入 `logs/master.log`/`logs/caves.log` 的 shard 日志落盘、最近日志读取 API 和日志展示面板、shard 异常退出后的自动状态清理与错误回传、基于启动时配置快照的 `restartRequired` 判定、持久化到 SQLite 的 runtime history 与一次自动重试策略，以及带有本地/远端版本比较、手动检查、手动更新、启动后定时检查、运行中更新保护、停服确认、更新任务日志读取/失败排查入口、安装任务日志查看入口、安装/更新任务展开日志自动刷新、活跃任务过滤、统一日志面板刷新抽象、runtime shard 日志面板统一展开/刷新交互、runtime/安装/更新/版本检查日志 SSE 持续推流、基于文件 offset 的服务端增量日志读取、前端 SSE 连接/断线/回退测试和版本检查日志落盘/排查入口的 DST 更新流程。

## 下一任务

- [ ] 评估是否需要把当前基于定时 poll + offset 增量读取的日志 SSE 再演进为 file watcher 驱动的 tail，并明确文件轮转/重建时的处理策略。

## 暂时不要做

- 不要添加 Docker 支持。
- 不要导入、迁移或修改 `/home/dontstarve/dst-server` 下的手动 DST 部署。
- 在 managed install、状态存储、进程生命周期完成前，不要提前做完整模组管理 UI。
- 默认不要把 Web UI 暴露到公网网卡。

## 最近完成检查

- [x] 已运行相关后端/前端检查。
- [x] 本次改过的 Go 文件已格式化。
- [x] 若边界或技术决策变化，已更新架构或决策文档。
- [x] 本文件已反映完成内容和下一任务。
- [ ] 本次改动未提交 commit。

## 未决问题

- 第一版公开发布时，`leveldataoverride.lua` 要做到多完整的可视化。
- 启动流程接入后，cluster 配置变更与运行中 shard 的重载策略要不要区分“需重启”与“即时生效”。
- 当前日志 SSE 已改成按秒检查文件并按 offset 读取新增内容，不再每次重读整段 tail；但它仍不是基于 file watcher 的真正事件驱动，后续可继续评估是否值得再复杂化。
- 当前 `restartRequired` 只基于 cluster 配置快照；后续如果 token、admin 列表、模组或世界设置接入运行态，也要纳入重启判定。
- 当前自动重试策略只做每个 shard 一次立即重试，没有退避、上限策略或外部告警通道。
