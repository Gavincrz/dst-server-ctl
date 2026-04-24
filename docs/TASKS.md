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

当前项目已有 harness、工程骨架、managed root 路径布局、共享 command runner、SQLite 状态存储基础层、启动时 managed root 初始化、安装状态 API、安装任务 API、任务模型、由任务驱动的 SteamCMD/DST 安装执行流程、初始化状态页，以及可反映控制器启动时间的基础运行状态。

## 下一任务

- [ ] 定义受管 cluster 的基础结构化配置状态和读写 API。

## 后续任务

- [ ] 为 cluster 配置生成基础 `cluster.ini` 和 shard `server.ini` 输出。

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
- [x] 未经用户明确要求，没有提交 commit。

## 未决问题

- 第一版公开发布时，`leveldataoverride.lua` 要做到多完整的可视化。
