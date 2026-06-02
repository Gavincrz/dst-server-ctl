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

## 2026-05-27 轻量核对结论

本轮先做“公开资料 + 当前仓库实现”的轻量核对，不在本地额外下载安装最新 DST。目标是先收敛高价值差异，再把“最新安装产物/脚本实物核对”留到后续通过受管安装链路验证。

本轮主要依据：

- Klei 官方论坛 Linux dedicated server quick setup guide，确认 cluster 基础文件布局、`cluster_token.txt`、`cluster.ini` 与 shard `server.ini` 的主链路。
- Klei 官方论坛 dedicated server 相关讨论，确认 `worldgenoverride.lua` 仍是当前主线运维语境中的世界配置入口，`dedicated_server_mods_setup.lua` 与 shard `modoverrides.lua` 仍是模组文件边界。
- 当前仓库 `internal/adapter/dstconfig/writer.go`、`internal/service/cluster.go`、`internal/http/router.go` 的实际字段与文件生成逻辑。

当前可先视为“已覆盖且链路完整”的范围：

- `cluster.ini`
  - `[GAMEPLAY]`：`game_mode`、`max_players`、`pvp`、`pause_when_empty`
  - `[NETWORK]`：`cluster_name`、`cluster_description`、`cluster_password`、`cluster_intention`、`cluster_language`、`offline_cluster`、`lan_only_cluster`、`tick_rate`
  - `[MISC]`：`console_enabled`
  - `[SHARD]`：`shard_enabled`，以及多 shard 时的 `bind_ip`、`master_port`、`cluster_key`
- `server.ini`
  - `[NETWORK]`：`server_port`，以及非 Master shard 的 `master_ip`、`master_port`
  - `[STEAM]`：`master_server_port`、`authentication_port`
  - `[SHARD]`：多 shard 时的 `is_master`，以及非 Master shard 的 `name`
- `worldgenoverride.lua`
  - 每 shard 的 `preset`
  - 结构化字段 + raw overrides 透传并存
  - 未结构化键仍可通过 overrides map 落盘

本轮确认仍缺失、且优先级高于继续零散补字段的范围：

- `cluster_token.txt`
- `adminlist.txt`、`whitelist.txt`、`blocklist.txt`
- `dedicated_server_mods_setup.lua`
- `Master/modoverrides.lua`、`Caves/modoverrides.lua`
- `leveldataoverride.lua` 兼容导出

本轮对 `cluster.ini` / `server.ini` 的结论：

- 以 Klei 官方 quick setup guide 能直接确认的基础参数看，当前控制器已经覆盖主干启动链路所需的大部分核心字段。
- 但这还不是“最新版本参数面已逐项穷尽核对”的结论。
- 后续如果要宣称“已按最新 DST 服务端和脚本源码核实”，仍需要在 managed root 下拿最新受管安装产物做二次实物核对。

本轮对世界参数的结论：

- 当前控制器已经把一批高频世界生成/规则字段做成结构化输入，并保留 raw overrides 兜底，这条模型方向是对的。
- 公开官方资料并没有提供一份稳定、完备、面向服务端维护者的“最新世界参数全集”清单，至少当前仓库里还没有足够可靠的来源可直接固化为最终基线。
- 因此，资源、生物、事件类字段继续扩展前，最好先基于最新受管安装中的真实脚本或模板做二次核对。

## 2026-06-02 最新受管安装产物实物核对

本轮已经通过 Web 前端完成受管安装与版本检查链路验证，并基于 managed root 下的最新 DST 安装产物做二次实物核对。

本轮主要依据：

- `~/.local/share/dst-server-ctl/dst/data/databundles/scripts.zip`
- `scripts/map/customize.lua`
- `scripts/worldsettings_overrides.lua`
- `scripts/languages/language.lua`
- 受管生成产物：
  `clusters/primary/cluster.ini`、
  `Master/server.ini`、
  `Caves/server.ini`、
  `Master/worldgenoverride.lua`、
  `Caves/worldgenoverride.lua`

本轮确认：

- 当前受管安装链路可用，DST 已成功安装到 managed root。
- `Check Now` 能正确读到本地与远端 build id，并完成版本比较。
- 当前生成文件仍与现有结构化模型一致：
  `cluster.ini`、`server.ini`、`worldgenoverride.lua` 主链路正常。
- `cluster_token.txt`、admin/allow/block 列表、`dedicated_server_mods_setup.lua`、各 shard `modoverrides.lua`、`leveldataoverride.lua` 兼容导出仍未接入。

本轮对世界配置的新增结论：

- 最新脚本中，世界参数已经明显分成两层：
  - worldgen 类：`WORLDGEN_GROUP`
  - runtime/worldsettings 类：`WORLDSETTINGS_GROUP`
- 当前控制器已结构化的 16 个字段，主要只覆盖了 `global` / `misc` 中一小部分基础项：
  `world_size`、`branching`、`loop`、`start_location`、`season_start`、`day`、`weather`、`autumn`、`winter`、`spring`、`summer`、`roads`、`touchstone`、`boons`、`cave_ponds`、`wormattacks`
- 真实脚本里仍有大量高价值字段尚未结构化，但很适合作为下一批候选：
  - worldgen / misc：`task_set`、`cavelight`、`prefabswaps_start`、`terrariumchest`、`stageplays`、`junkyard`
  - worldsettings / misc：`hounds`、`winterhounds`、`summerhounds`、`lightning`、`wildfires`、`petrification`、`earthquakes`、`wormattacks_boss`、`atriumgate`
  - worldsettings / global：`spawnmode`、`ghostenabled`、`portalresurection`、`resettime`、`krampus`
  - worldsettings / survivors：`spawnprotection`、`dropeverythingondespawn`、`healthpenalty`、`temperaturedamage`、`hunger`、`darkness`
  - worldsettings / events：`specialevent` 与各个节庆开关
- 其中一批字段带有 `masteroption`、`master_controlled` 或 `master_sync` 标记，说明它们更适合作为 cluster 级或主世界控制项建模，而不是简单继续塞回每 shard 独立 overrides。
- `scripts/worldsettings_overrides.lua` 说明很多字段并不只是“落盘一个键”，而是会映射到多组 `TUNING` / 世界行为；这进一步证明继续走“结构化状态 + overrides 兜底”是对的，但字段分组需要更贴近脚本真实语义。

本轮对语言配置的新增结论：

- `scripts/languages/language.lua` 确认 dedicated server 的语言来源仍是 `TheNet:GetDefaultServerLanguage()`，也就是 cluster 语言配置确实是服务端主源之一。
- 但完整语言代码对照表不在这个入口文件本身里，`T-011` 仍然不能只凭这次核对直接判定为完成。

基于本轮实物核对，当前更合理的优先级是：

1. 先补一批高价值 worldsettings 字段，而不是继续盲目扩展任意 overrides。
2. 同步推进 `cluster_token.txt`、admin/allow/block 列表与模组文件边界，避免“世界字段越来越多，但服务器管理关键文件仍缺失”。
3. 之后再评估是否把世界配置从通用 overrides map 演进为更清晰的子模型或分组表单。

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
