# 调研

## DST 专用服务器

关键事实：

- DST 专用服务器通过 SteamCMD app `343050` 安装。
- 一个启用洞穴的普通世界会运行两个 shard 进程：`Master` 和 `Caves`。
- 共享服务器设置在 `cluster.ini`。
- shard 专属网络设置在各自的 `server.ini`。
- cluster token 存在 `cluster_token.txt`。
- 服务端模组通过 `dedicated_server_mods_setup.lua` 下载，例如 `ServerModSetup("workshop_id")`。
- 启用的模组和模组配置在 `modoverrides.lua`。

本开发 VPS 上观察到的本机参考：

- Master 进程：`./dontstarve_dedicated_server_nullrenderer -persistent_storage_root /home/dontstarve/dst-server/dontstarve-config -conf_dir server_dir -cluster lhy_server -console -shard Master`
- Caves 进程：`./dontstarve_dedicated_server_nullrenderer -persistent_storage_root /home/dontstarve/dst-server/dontstarve-config -conf_dir server_dir -cluster lhy_server -console -shard Caves`
- 该手动部署仅作参考。项目不能导入、修改或假定拥有 `/home/dontstarve/dst-server`。

参考链接：

- Klei Linux 专用服务器指南：https://forums.kleientertainment.com/forums/topic/64441-dedicated-server-quick-setup-guide-linux/
- DST 专用服务器指南：https://dontstarve.wiki.gg/wiki/Guides/Don%E2%80%99t_Starve_Together_Dedicated_Servers
- SteamCMD 文档：https://developer.valvesoftware.com/wiki/SteamCMD

## 同类项目

已有项目主要强在 Docker 部署和 CLI 自动化。本项目刻意差异化：非 Docker 单二进制部署，并提供世界和模组的可视化 Web UI。

示例：

- https://github.com/Jamesits/docker-dst-server
- https://hub.docker.com/r/superjump22/dontstarvetogether

## Agent Harness

harness 采用短 `AGENTS.md`、当前架构文档、决策记录和任务跟踪。目标是给 coding agent 持久上下文，同时避免每次对话塞入过多陈旧或低信号内容。

参考链接：

- OpenAI harness engineering：https://openai.com/index/harness-engineering/
- OpenAI Codex guide：https://openai.com/business/guides-and-resources/how-openai-uses-codex/

## 日志 watcher 方案

当前实现：

- 日志 API 首次连接时读取最近窗口作为 `snapshot`。
- 后续 SSE 连接每秒检查文件状态，并按文件 `offset` 只读取新增内容。
- 文件被轮转、替换、删除重建时，服务端会退回发送新的 `snapshot`。

候选方案对比：

- 保持当前 `poll + offset`：
  优点是实现简单、跨平台行为稳定、和现有 `domain.LogStream`/`adapter/logtail` 边界一致。
  缺点是空闲连接仍有周期性 `stat`/`poll` 开销，日志推送延迟下限仍受轮询周期影响。
- 改成 `file watcher + offset`：
  优点是新增日志可更快推送，并把高频轮询降为低频兜底。
  缺点是复杂度明显上升，需要处理 watcher 丢事件、队列溢出、文件轮转、不同平台语义差异和额外生命周期管理。

当前结论：

- 在本项目现阶段，`file watcher` 不是优先项。
- 已经通过 `offset` 增量读取消除了最主要的重复 IO；继续上 watcher 的收益，主要是更低延迟和更少空转，而不是数量级上的架构改进。
- 如果后续出现日志连接数明显增加、空转 IO 成本可见，或用户明确要求接近实时的子秒级日志推送，再考虑实现 watcher 驱动。

若后续实现 watcher，建议边界：

- `domain.LogStream` 接口保持不变，不把 watcher 细节泄漏到 `service` 或 `http`。
- watcher 实现仍放在 `adapter/logtail`，和当前基于 `offset` 的文件读取共用同一份增量拼行逻辑。
- 保留低频 fallback poll，处理 watcher 丢事件和文件替换场景。
- 文件轮转、截断、删除重建统一视为日志源重置，对 SSE 发送新的 `snapshot`，不要尝试跨文件拼接旧上下文。

## Dashboard 细粒度事件

当前 dashboard SSE 是整包 `snapshot` 模型，适合替换页面级普通轮询，但还不是细粒度事件总线。

候选内容拆分后，性质不同：

- `runtimeHistory`：
  这是持久化状态，天然适合继续细化成 append-only 事件流。后端已有运行历史仓储和列表 API，后续若要减少整包 snapshot 传输，优先从这里开始最合理。
- 安装/更新任务列表与版本状态：
  也属于后端真实状态，可以事件化，但需要明确“task changed”“update state changed”这类稳定事件类型，否则只是把整包 snapshot 切碎，收益有限。
- cluster 保存结果、运行控制成功提示、安装/更新按钮后的成功提示：
  这些当前主要是前端根据单次 HTTP 操作结果生成的临时回执，不是后端持久状态。若直接塞进 dashboard SSE，会把 UI 文案、提示时机和后端协议耦合起来。

当前结论：

- 暂不把 `clusterMessage`、`runtimeMessage`、`actionMessage`、`updateMessage` 这一类临时回执并入 dashboard SSE。
- 若后续确实需要跨标签页、跨刷新保留“最近操作结果”，应先在后端定义独立的操作事件模型或通知模型，而不是复用 dashboard 状态流。
- 下一步若继续细化 dashboard，优先顺序应是：
  1. `runtimeHistory` append 事件。
  2. 安装/更新任务状态变更事件。
  3. 版本检查状态变更事件。
