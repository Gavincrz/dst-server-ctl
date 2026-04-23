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
