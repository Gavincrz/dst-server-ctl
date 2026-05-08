# 架构

## 概览

`dst-server-ctl` 是一个单二进制 DST 服务器控制器，用来管理一个受管的 Don't Starve Together 专用服务器实例。它通过 SteamCMD 安装和更新 DST，从结构化状态生成 DST 配置文件，启动 Master/Caves 进程，读取日志，并提供本地 Web UI。

控制器不接管、不修改用户已有的手动 DST 部署。所有受管数据都放在独立 managed root 中，默认使用 `$XDG_DATA_HOME/dst-server-ctl`，没有该环境变量时使用 `~/.local/share/dst-server-ctl`。

## 后端分层

- `domain`：核心概念，例如安装布局、世界、shard、mod、任务、服务器状态。
- `service`：用例编排，例如安装、更新、启动、停止、世界选择、配置渲染。
- `adapter`：外部集成，例如文件系统、进程 runner、SteamCMD、SQLite、日志 tail、DST 配置 writer。
- `http`：API 路由、请求校验、认证、响应格式。

依赖方向向内：HTTP 调 service，service 使用 domain 类型和 adapter 接口，adapter 实现外部副作用。

## 运行时数据

控制器状态存 SQLite。DST 原生文件从结构化状态生成到 managed root。

managed root 使用稳定子目录：

- `steamcmd/`：SteamCMD 安装目录。
- `dst/`：DST 专用服务器安装目录。
- `clusters/`：生成的 cluster、世界、存档、token、shard 配置。
- `logs/`：控制器日志和任务日志。
- `state/`：SQLite 数据库和本地控制器元数据。

## DST 文件策略

控制器拥有受管服务器的生成文件：

- `cluster.ini`
- `adminlist.txt`
- `whitelist.txt`
- `blocklist.txt`
- `Master/server.ini`
- `Caves/server.ini`
- `Master/worldgenoverride.lua`
- `Caves/worldgenoverride.lua`
- `Master/modoverrides.lua`
- `Caves/modoverrides.lua`
- `cluster_token.txt`

DST 文件 writer 放在 adapter 层。handler 和 UI 不能手写这些格式。

配置边界按四层拆分：

- cluster 共享配置：`cluster.ini`、token、admin/allow/block 列表
- shard 专属配置：各 shard `server.ini`
- 世界配置：各 shard `worldgenoverride.lua`，必要时兼容 `leveldataoverride.lua`
- 模组配置：`dedicated_server_mods_setup.lua` 与各 shard `modoverrides.lua`

## 进程策略

控制器直接启动并监督单个受管服务器实例的 Master 和 Caves 进程，负责启动、停止、重启、状态检测、更新安全和日志流。

首版不生成 systemd unit。后续 installer 可以选择让 systemd 托管控制器自身，但 shard 生命周期仍由控制器内部管理。

## 安全

默认监听 `127.0.0.1`。控制器首次运行生成 admin token，并要求修改类 API 携带该 token。日志和普通 API 响应中必须隐藏敏感信息。
