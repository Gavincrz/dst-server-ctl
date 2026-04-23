# 路线图

## 阶段 0：Harness 和骨架

- 建立 agent 指令和架构文档。
- 创建带 package 边界的 Go 后端骨架。
- 创建 Svelte 前端骨架。
- 添加路径安全和 command runner 基础测试。

## 阶段 1：受管安装 MVP

- 初始化 managed root。
- 下载/安装 SteamCMD。
- 安装 DST app `343050`。
- 将安装状态持久化到 SQLite。
- 在 Web UI 显示安装状态。

## 阶段 2：服务器生命周期 MVP

- 创建一个包含 Master 和 Caves 的受管 cluster。
- 生成核心 `cluster.ini` 和 `server.ini`。
- 启动、停止、重启，并查看 shard 状态。
- 在 UI 中流式查看 Master 和 Caves 日志。

## 阶段 3：更新

- 手动执行 SteamCMD 更新。
- 检查本地和远端版本状态。
- 添加每日定时更新检查。
- 更新运行中的世界前必须显式确认停服。

## 阶段 4：世界和服务器配置 UI

- 配置服务器名、描述、密码、语言、人数、PvP、暂停行为、游戏模式。
- 管理 token 输入，但不回显 token。
- 管理 admin、block、allow 列表。
- 添加世界模板和生成的 `leveldataoverride.lua`。

## 阶段 5：模组管理

- 添加 Workshop ID，并生成 `dedicated_server_mods_setup.lua`。
- 更新/下载模组。
- 下载后读取本地 `modinfo.lua` 元数据。
- 为支持的 `configuration_options` 生成可视化表单。
- 为 Master 和 Caves 生成 `modoverrides.lua`。

## 阶段 6：发布打包

- 添加 Linux amd64/arm64 release 构建。
- 添加可选安装脚本。
- 添加干净卸载流程，并提示用户选择保留或删除世界、存档、模组和 DST 安装文件。
