# 模型价格补充

默认价格与展示属性来自 models.dev 的 `https://models.dev/catalog.json`。程序内嵌经过校验的离线快照，启动时加载磁盘缓存或离线目录；后台使用 ETag 同步，完整验证成功后才一起发布价格、操作价格和属性。更新失败保留最近有效版本。

`model_pricing_supplements.json` 随发布包分发，只填补当前计费需要且目录缺失的字段。每个模型条目记录 `source_url` 和 `verified_at`。目录已经有值的字段，包括零价，均优先于补充。不要重复登记普通 token 价格、复制展示属性、借用另一型号价格或保存未经核实的估算值。

## 文件与价格配置

- `pricing.fallback_file` 指定补充文件；缺失字段保持缺价。
- 自定义售价通过管理端「价格配置」设置，再关联相应分组；显式零价仍然有效。
- `pricing.override_file` 和 `PRICING_OVERRIDE_FILE` 已退役。旧文件不读取、不自动迁移，启动时提示弃用；升级前迁移需要保留的售价，提供商成本另行配置。
- 保留旧 JSON 字段及 `litellm_provider` 的读取兼容；同时存在时 `provider` 优先，包括空值和 `null`。程序不改写部署者的文件。
- 默认每 10 分钟检查；管理员更新目录可立即重载。远程返回 304 或远程地址为空时，本地文件修改和删除仍生效。非法单价、规则或非对象条目拒绝本次更新。

同一原厂记录的裸名和供应商限定名共享补充，精确键优先；其他供应商、日期版本及档位后缀不会自动继承。保留 `source` 和 `price_sources` 作为实际字段来源，来源类别与 `source_url` 分开保存。

## 计价维度

价格统一使用美元。旧 token 字段单位为美元/token；`image_prices` 的 1K/2K/4K 键为美元/张，`video_prices` 的 480p/720p/1080p 键为美元/秒。不同单位不能替代；缺失尺寸不按固定倍率补价。旧 `output_cost_per_image` 表示不分尺寸的单张价，仍可读取。

`fast_multiplier`、`flex_multiplier`、`max_reasoning_effort_multiplier` 描述明确倍率；`cache_write_multiplier` 和 `cache_write_1h_multiplier` 仅在对应缓存单价缺失时由输入价派生，显式零价优先。`time_pricing` 使用现有价格配置的时区、每日时段和倍率结构。未声明规则时不根据型号产生加价或折扣。

`_billing_defaults` 是操作价格保留节点，不是模型。可用字段为 `web_search_price_per_call`、`search_price_per_1k`、`audio_realtime_price_per_min`、`audio_tts_price_per_million_chars`、`audio_stt_price_per_hour`。单位分别是美元/次、美元/千次、美元/分钟、美元/百万字符、美元/小时。管理员价格配置优先；`sources` 逐字段记录依据，`verified_at` 记录核验日期。

当前分发数据保留 OpenAI Web Search、xAI TTS 和 REST STT 的操作价。统一搜索次数无法表达 X Search 按帖子和档案的收费方式，通用实时音频时长也不足以覆盖不同语音型号，因此这两类操作不提供统一默认金额。部署者可以明确配置自己的售价。

Gemini 生图目录的 `cost.output` 是图片 token 价，补充只提供独立文本输出价及公布的尺寸单价。GPT Image 1/mini/1.5 补齐目录缺少的媒体价格桶。xAI 图片和视频只登记已核实的输出尺寸单价；`grok-imagine-image-2.0` 的自动质量随生成/编辑变化，现有计价输入没有质量维度，因此不分发该型号的统一默认单价。输入媒体、质量等尚未进入现有用量结构的维度不会伪装成输出价格。

## 维护

新增补充前，使用实际目录解析结果检查完整供应商身份和价格桶。确认数据源缺口后再登记官方依据；无法核实则保持缺价。目录补齐后删除重复字段。离线快照与 MIT 许可保存在 `backend/internal/modelcatalog/`，不能用补充文件替换完整缓存。

价格清理不重算历史账单或任务资金快照，也不改变既有缺价处理及提供商协议规则。
