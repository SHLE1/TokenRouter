# 模型价格补充

远程模型价格与展示属性来自 models.dev 的 `https://models.dev/catalog.json`。程序内嵌 `backend/internal/modelcatalog/catalog.json.gz` 离线快照；首次启动先加载磁盘缓存或离线快照，后台使用 ETag 条件请求同步。完整验证成功后才一起发布价格和属性，失败时保留现有版本。

本目录中的 `model_pricing_supplements.json` 随发布包分发，用于补齐目录缺少的媒体单价及生图模型文本输出价。已存在的目录价格优先；独立补充模型仍可使用现有价格字段。补充文件不提供模型展示属性。

`pricing.fallback_file` 指定补充文件，`pricing.override_file` 指定优先级更高的本地覆盖文件。覆盖按模型键逐字段浅合并，`null` 删除字段。两种文件都必须是 JSON 对象。修改和删除会在下一次目录检查时生效，也可通过管理员“更新目录”立即重载；非法内容保留当前版本并报告错误。

新文件使用 `provider` 表示提供方，读取时兼容 `litellm_provider`。两者同时存在时以 `provider` 为准，包括空字符串和 `null`。旧文件中的价格字段、单位和覆盖规则保持兼容。

默认每 10 分钟检查，配置键为 `pricing.check_interval_minutes`。不再下载独立 hash 文件，也不分发旧完整价格 JSON。远程地址为空时仍可读取离线目录，并监测本地补充及覆盖文件。

离线快照及其 MIT 许可位于 `backend/internal/modelcatalog/`。更新快照时应验证目录解析和默认价格契约，保留许可文件；不要用补充文件替换完整目录缓存。
