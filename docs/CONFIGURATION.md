# 配置分层

## 目标

把 DST 服务器配置拆成稳定层次，避免继续把“少量已实现字段”误当成最终模型。

当前控制器已经覆盖：

- `cluster.ini`：基础 gameplay/network/misc/shard 字段
- `server.ini`：每 shard 的 game/steam/auth 端口和主从联机骨架
- `worldgenoverride.lua`：每 shard 的 preset + overrides 主源

当前仍未覆盖：

- `cluster_token.txt`
- admin/allow/block 列表
- `leveldataoverride.lua` 兼容导出
- 模组配置

## 文件边界

### 1. cluster 共享配置

文件：

- `cluster.ini`
- `cluster_token.txt`
- `adminlist.txt`
- `whitelist.txt`
- `blocklist.txt`

职责：

- 对外可见的服务器身份与接入策略
- cluster 级玩法与联机模式
- shard 间共享的联机参数
- 敏感信息和访问控制列表

建议字段分组：

- Gameplay：`game_mode`、`max_players`、`pvp`、`pause_when_empty`
- Network/Public：`cluster_name`、`cluster_description`、`cluster_language`、`cluster_password`、`cluster_intention`
- Network/Transport：`offline_cluster`、`lan_only_cluster`、`tick_rate`
- Misc：`console_enabled`
- Shard Cluster：`shard_enabled`、`bind_ip`、`master_port`、`cluster_key`
- Secrets/Lists：token、admin、allow、block

说明：

- `cluster_token.txt`、密码、admin 凭据都必须继续视为敏感信息。
- Web 默认仍监听 `127.0.0.1`，但 DST 服务器对外监听和 shard 内部联机端口应作为独立配置处理，不能和控制器监听地址混在一起。

### 2. shard 专属配置

文件：

- `Master/server.ini`
- `Caves/server.ini`

职责：

- shard 身份
- shard 对外游戏端口
- shard steam 端口
- 多 shard 拓扑中的主从连通

建议字段分组：

- Shard Identity：`is_master`、`name`
- Network：`server_port`
- Steam：`master_server_port`、`authentication_port`
- Secondary Shard Link：`master_ip`、`master_port`

说明：

- 现有实现只写了 `is_master`、`name`、`master_ip`、`master_port` 的一部分骨架。
- `server_port`、`master_server_port`、`authentication_port` 后续需要进入结构化状态，否则多实例和端口冲突无法可靠管理。

### 3. 世界配置

文件：

- `Master/worldgenoverride.lua`
- `Caves/worldgenoverride.lua`
- `Master/leveldataoverride.lua`
- `Caves/leveldataoverride.lua`

职责：

- shard 世界生成 preset
- shard 世界参数覆盖
- 洞穴/地表差异化设置

当前结论：

- 优先把 `worldgenoverride.lua` 视为人类可编辑主源。
- `leveldataoverride.lua` 更接近客户端内部保存格式，不应作为第一版核心编辑目标。
- 首轮世界配置先支持“preset + 少量关键 overrides + Master/Caves 分离”，但领域模型必须为全量参数扩展预留结构。

建议状态模型：

- World Preset：如地表默认、洞穴默认
- World Overrides：键值映射，保持未知字段可透传
- World Meta：位置、required prefabs、版本等仅在确有必要时结构化

### 4. 模组配置

文件：

- `dedicated_server_mods_setup.lua`
- `Master/modoverrides.lua`
- `Caves/modoverrides.lua`

职责：

- 模组下载清单
- shard 维度启用/禁用
- 模组选项配置

说明：

- 这是独立于世界配置的第四层，不应和 `cluster.ini` 或世界 preset 混在一个表单模型里。
- mod 配置仍排在世界和服务器配置之后。

## 建议领域拆分

后端不应继续把所有配置长期塞在单个 `domain.ClusterConfig` 里。建议按下列子模型演进：

- `ClusterSettings`：共享玩法、公开展示、接入策略、token/list 元数据
- `ShardSettings`：每个 shard 的启用状态、端口、主从拓扑字段
- `WorldSettings`：每个 shard 的 preset 与 overrides
- `ModSettings`：模组下载项、启用项和配置项

这样可以保持 `service` 编排清晰，同时让 `adapter/dstconfig` 继续负责最终文件生成。

## 实施优先级

1. 先扩展 cluster/shared + shard/network 的结构化状态，补齐 `cluster.ini` 和 `server.ini` 的主要参数边界。
2. 再引入每 shard 世界配置模型，主源采用 `worldgenoverride.lua`。
3. 然后补 token、密码和访问控制列表的写入与脱敏展示。
4. 最后进入模组下载清单与 `modoverrides.lua`。

## 待确认项

- `cluster_language` 是否存在 Klei 官方维护的完整语言代码表；当前只确认英语可作为默认首个支持目标。
- 是否需要在后续提供 `leveldataoverride.lua` 兼容导出，或只保留 `worldgenoverride.lua` 主写入路径。
