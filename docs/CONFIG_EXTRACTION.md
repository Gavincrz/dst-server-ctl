# 配置字段提取

这份文档说明如何从最新受管 DST 安装产物里重新提取配置字段清单，避免每次手工翻 `scripts.zip`。

## 目标

提取流程主要服务于两类任务：

- 核对最新 DST 版本里有哪些 worldgen / worldsettings 字段。
- 为 `T-009` / `T-010` 选择下一批值得结构化支持的高价值字段。

它不是自动更新业务模型的机制。

- 控制器仍以“结构化已支持字段 + raw overrides 兜底”为主。
- 提取结果是候选清单，不应直接自动改 API、存储或 UI。

## 数据来源

当前脚本默认读取受管安装目录下这些产物：

- `dst/steamapps/appmanifest_343050.acf`
- `dst/data/databundles/scripts.zip`

其中重点关注：

- `scripts/map/customize.lua`
- `scripts/worldsettings_overrides.lua`
- `scripts/languages/*.po`

说明：

- `cluster.ini`、`server.ini`、`worldgenoverride.lua` 是控制器生成产物，不在原始 server 安装包里。
- `scripts/map/customize.lua` 更适合提取字段分组、默认值、世界适用范围和选项来源。
- `scripts/worldsettings_overrides.lua` 更适合人工判断某个字段背后是否只是简单 override，还是会映射到更复杂的行为逻辑。

## 用法

默认直接读取：

```sh
make extract-config-fields
```

等价于：

```sh
python3 ./scripts/extract_dst_config_fields.py
```

如果 managed root 不在默认位置，可以显式传入：

```sh
python3 ./scripts/extract_dst_config_fields.py \
  --managed-root /path/to/dst-server-ctl
```

如果想保存结果到文件：

```sh
python3 ./scripts/extract_dst_config_fields.py \
  --output /tmp/dst-config-fields.md
```

如果需要 JSON 结果做后续处理：

```sh
python3 ./scripts/extract_dst_config_fields.py \
  --format json \
  --output /tmp/dst-config-fields.json
```

## 输出内容

脚本当前会输出：

- 当前受管安装的 `buildid`
- `scripts/languages/*.po` 语言文件列表
- `WORLDGEN_GROUP`
- `WORLDSETTINGS_GROUP`

每个字段项会带上这些信息：

- `key`
- `default`
- `worlds`
- `desc`
- `options`
- `flags`

其中：

- `desc` 是 `customize.lua` 里的选项来源标识，例如 `frequency_descriptions`
- `options` 是脚本能静态解析出的 `data = "..."` 值列表
- `flags` 会标出 `masteroption`、`master_controlled`、`master_sync`

这些标记很重要，因为它们能帮助判断字段更适合作为：

- 每 shard 独立 world override
- 主世界控制项
- cluster 级共享设置

## 推荐工作流

拿到新版本 DST 后，建议按这个顺序做：

1. 先通过网页端或受管更新流程拿到最新安装产物。
2. 跑 `make extract-config-fields`，保存一份本次版本报告。
3. 从 `WORLDSETTINGS_GROUP` 和 `WORLDGEN_GROUP` 里筛选高价值字段。
4. 结合 `scripts/worldsettings_overrides.lua` 判断字段是否只是简单枚举，还是带复杂行为映射。
5. 把候选字段回填到 `docs/TASKS.md` 或具体实现任务中。

优先筛选标准建议是：

- 玩家常改
- 服务端运维能清晰理解
- 选项集合稳定
- 不需要立刻重构领域模型也能安全接入

## 局限

当前脚本是静态提取工具，不做这些事：

- 不自动修改控制器结构化配置模型
- 不自动生成 Go / TS 字段定义
- 不自动判断哪些字段必须做 cluster 级抽象
- 不自动解析 `worldsettings_overrides.lua` 里的所有行为语义

因此，脚本结果应该被当成“最新版本字段库存”，不是“可直接上线的字段设计”。
