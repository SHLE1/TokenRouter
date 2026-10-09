-- 兼容升级前的去重值，重放任务时共用同一条用量记录。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_billing_key_api_key
ON usage_logs ((COALESCE(billing_key, request_id)), api_key_id);
