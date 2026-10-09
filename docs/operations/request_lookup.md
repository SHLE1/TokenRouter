# 请求 ID 与请求查询

本文记录请求 ID、诊断记录、计费关联和请求详情查询。用量聚合见[使用记录与运维预聚合](pre_aggregation.md)，升级步骤见[版本升级说明](upgrade_notes.md#request_id_unification)。

<a id="request_identity"></a>
## 请求标识

全局 HTTP 入口生成本地 UUID，网关中间件复用它。`X-Request-ID`、`X-TokenRouter-Request-ID` 和兼容的 `X-Sub2API-Request-ID` 返回同一个值。响应 writer 在提交正文、刷新和连接升级前恢复这些头，上游响应头因此无法替换本地请求 ID。浏览器可以通过 CORS 读取这些响应头。

调用方的 `X-Request-ID` 和 `X-Client-Request-ID` 保存为调用方别名。合法的 `X-Client-Request-ID` 继续回显，缺失时使用本地 ID。别名用于查询，权限根据认证主体、记录归属和团队关系判断。

提供商重试和切换共用入口 ID。WebSocket 在每轮开始时生成请求 ID，提供商切换和重试复用该轮 ID，并保存连接请求的 ID。异步任务使用从稳定业务键派生的统一格式 ID，重放得到同一个值。上游请求 ID 和响应 ID 保存为外部别名。

<a id="request_storage"></a>
## 记录与恢复

`requestlog` 保存到达正常运行服务的请求摘要，包括成功、拒绝、失败、取消和未完成请求。健康检查和 404 也经过记录入口。请求到达应用中间件之前的网络失败没有应用请求记录。

`request_records` 保存起止时间、状态、路径、认证后的归属、模型、阶段耗时、上游尝试和关联标识。完整提示词与报文按各业务既有的留存规则处理。未知凭据的请求保留管理员可查的摘要。已匹配 Key 的请求按 Key 的行为用户和团队归属查询，认证权限仍由认证中间件判断。

请求快照先写入本地待写目录，再由后台按批写入 PostgreSQL。短窗口内的更新共用一次追加日志和磁盘同步，同一 ID 的更新按时间合并。检查点通过原子替换回收已确认数据。数据库写入成功后确认对应版本，写入期间到达的新版本继续等待下一批。旧快照重放不会覆盖更新的状态。进程重启后重放完整日志帧，写到一半的末帧会被移除。待写快照的内存预算为 64 MiB，超过预算的更新会报告持久化失败。

待写目录使用进程文件锁。同一台主机上的多个实例需要配置各自的目录；容器需要为目录挂载持久存储。目录不可写、磁盘不足或数据库故障会写日志并增加失败计数。管理员通过 `/api/v1/admin/requests/health` 查看待写数量和累计故障次数。数据库暂不可用时，本机按本地请求 ID 查询可以返回归属匹配的待写摘要。

`running` 表示尚未记录终态，可能仍在执行，也可能在进程退出时中断。请求详情对仍存在的创作台和批量图片任务读取任务表的状态。历史记录缺少元数据时使用 `legacy` 标记，各类原始记录分别展示；单独的用量记录不能证明 HTTP 请求成功。

## 配置与留存

```yaml
request_log:
  retention_days: 30
  spool_dir: ""
```

`retention_days` 默认 30 天，值为 0 时使用默认值。`spool_dir` 为空时使用 `${DATA_DIR}/request-records`；没有 `DATA_DIR` 时使用价格资源数据目录下的 `request-records`。环境变量分别为 `REQUEST_LOG_RETENTION_DAYS` 和 `REQUEST_LOG_SPOOL_DIR`。修改配置后重启服务。

请求摘要有独立的清理任务，Ops 开关和采样配置控制各自的监控数据。用量、审计和计费数据按各自规则留存，摘要过期后仍可查询尚存的业务记录。备份的 Ops 日志选项同时控制 `request_records`，本地待写目录需要随数据卷保存。

<a id="billing_identity"></a>
## 计费兼容

新使用记录的 `request_id` 保存统一请求 ID，`billing_key` 保存资金操作的去重值。资金账本和去重表里的既有 `request_id` 字段继续表示计费键。普通请求、WS 轮次和任务使用各自既有的结算规则，外部别名不会替代计费键。

使用记录按 `(COALESCE(billing_key, request_id), api_key_id)` 去重。历史行的 `billing_key` 为空时使用原请求 ID，因此任务跨版本重放共用同一条用量记录。预占、扣款、释放的动作键分别处理资金操作。

`client:`、`local:` 和 `generated:` 等历史形式由查询兼容处理。历史 ID 之间的对应关系从尚存的系统日志和错误记录读取，缺失的对应关系不会通过字符串相似度补全。

<a id="request_query"></a>
## 查询与权限

- 用户入口：`GET /api/v1/requests?request_id=...`，或 `/api/v1/requests/:request_id`。
- 管理员入口：`GET /api/v1/admin/requests?request_id=...`，或 `/api/v1/admin/requests/:request_id`。
- 用量、错误、系统日志、审计和审核列表支持 `request_id` 筛选。精确查询省略时间条件时读取仍在保留期内的数据。

响应中的 `items` 包含请求摘要、用量及错误引用，父请求查询包含其任务和轮次。同一个外部 ID 可以匹配多次请求，超过 100 条时返回 `has_more`。页面提供请求 ID 搜索、复制和详情跳转，使用记录的汇总与导出采用相同筛选条件。

用户查询根据行为用户和当前 Owner 团队范围过滤。供应商标识、外部别名和审计引用由管理员查看，用户错误引用受用户错误展示设置控制。请求 ID 本身不赋予访问权限。
