-- 提供商改名只修改历史大表的元数据；不回填用量或重建索引。
SET LOCAL lock_timeout = '10s';
SET LOCAL statement_timeout = '120s';

DO $$
DECLARE
    item RECORD;
    target_tables OID[];
BEGIN
    -- 名称来自当前应用 schema，限定在 public，避免触及扩展与其它租户 schema。
    FOR item IN
        SELECT c.oid, c.relname FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relkind = 'r'
          AND c.relname IN ('accounts', 'account_groups',
              'pricing_config_account_stats_pricing_rules',
              'pricing_config_account_stats_model_pricing',
              'pricing_config_account_stats_pricing_intervals')
        ORDER BY c.relname
    LOOP
        EXECUTE format('ALTER TABLE public.%I RENAME TO %I', item.relname, replace(item.relname, 'account', 'provider'));
    END LOOP;

    -- 元数据范围只包含本应用的提供商相关表，不修改其它 public 表的同名词汇。
    SELECT array_agg(c.oid) INTO target_tables FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relname IN ('providers', 'provider_groups', 'batch_image_jobs',
              'content_moderation_cyber_warnings', 'creative_runs',
              'group_availability_probe_results', 'ops_error_logs', 'ops_system_logs',
              'ops_system_metrics', 'pricing_config_provider_stats_pricing_rules',
              'scheduled_test_plans', 'scheduler_outbox', 'usage_analytics_daily',
              'usage_analytics_hourly', 'usage_dashboard_daily', 'usage_dashboard_hourly', 'usage_logs');
    SELECT target_tables || COALESCE(array_agg(c.oid), '{}'::oid[]) INTO target_tables
    FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'public' AND c.relname IN ('pricing_config_provider_stats_model_pricing', 'pricing_config_provider_stats_pricing_intervals');

    FOR item IN
        SELECT c.relname, a.attname FROM pg_attribute a
        JOIN pg_class c ON c.oid = a.attrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relkind = 'r'
          AND a.attnum > 0 AND NOT a.attisdropped AND a.attname LIKE '%account%'
          AND c.relname IN ('providers', 'provider_groups', 'batch_image_jobs',
              'content_moderation_cyber_warnings', 'creative_runs',
              'group_availability_probe_results', 'ops_error_logs', 'ops_system_logs',
              'ops_system_metrics', 'pricing_config_provider_stats_pricing_rules',
              'scheduled_test_plans', 'scheduler_outbox', 'usage_analytics_daily',
              'usage_analytics_hourly', 'usage_dashboard_daily', 'usage_dashboard_hourly', 'usage_logs')
        ORDER BY c.relname, a.attnum
    LOOP
        EXECUTE format('ALTER TABLE public.%I RENAME COLUMN %I TO %I', item.relname, item.attname, replace(item.attname, 'account', 'provider'));
    END LOOP;

    -- 创作与批量图片中的旧 provider 字段表示平台，不是提供商实体。
    FOR item IN
        SELECT table_name FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name IN ('creative_runs', 'batch_image_jobs') AND column_name = 'provider'
    LOOP
        EXECUTE format('ALTER TABLE public.%I RENAME COLUMN provider TO platform', item.table_name);
    END LOOP;

    -- 外键自动跟随表和列；约束及索引改名保留物理索引与验证状态。
    FOR item IN
        SELECT c.relname, con.conname FROM pg_constraint con
        JOIN pg_class c ON c.oid = con.conrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND con.conname LIKE '%account%' AND con.conrelid = ANY(target_tables)
        ORDER BY c.relname, con.conname
    LOOP
        EXECUTE format('ALTER TABLE public.%I RENAME CONSTRAINT %I TO %I', item.relname, item.conname, replace(item.conname, 'account', 'provider'));
    END LOOP;
    FOR item IN
        SELECT c.relname, c.relkind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relkind IN ('i', 'S') AND c.relname LIKE '%account%'
          AND ((c.relkind = 'i' AND EXISTS (SELECT 1 FROM pg_index i WHERE i.indexrelid = c.oid AND i.indrelid = ANY(target_tables)))
            OR (c.relkind = 'S' AND EXISTS (SELECT 1 FROM pg_depend d WHERE d.classid = 'pg_class'::regclass
                AND d.objid = c.oid AND d.refclassid = 'pg_class'::regclass AND d.refobjid = ANY(target_tables) AND d.deptype IN ('a', 'i'))))
        ORDER BY c.relname
    LOOP
        EXECUTE format('ALTER %s public.%I RENAME TO %I', CASE WHEN item.relkind = 'i' THEN 'INDEX' ELSE 'SEQUENCE' END, item.relname, replace(item.relname, 'account', 'provider'));
    END LOOP;
END $$;

-- 仅迁移现行配置中的自有键，第三方身份键保持原始协议。
CREATE OR REPLACE FUNCTION pg_temp.provider_config_keys(value JSONB) RETURNS JSONB
LANGUAGE plpgsql AS $$
DECLARE
    result JSONB;
    entry RECORD;
    new_key TEXT;
BEGIN
    IF jsonb_typeof(value) = 'object' THEN
        result := '{}'::jsonb;
        FOR entry IN SELECT key, val FROM jsonb_each(value) AS e(key, val) LOOP
            new_key := CASE WHEN entry.key IN ('account_id', 'account_uuid', 'account_structure', 'account_residency_region')
                OR entry.key LIKE 'chatgpt_account_%' OR entry.key LIKE 'crs_account_%'
                OR entry.key LIKE 'aws_account_%' OR strpos(entry.key, 'service_account') > 0
                THEN entry.key ELSE replace(entry.key, 'account', 'provider') END;
            IF new_key <> entry.key AND value ? new_key THEN
                RAISE EXCEPTION 'provider configuration key conflict: %', new_key;
            END IF;
            result := result || jsonb_build_object(new_key, pg_temp.provider_config_keys(entry.val));
        END LOOP;
        RETURN result;
    ELSIF jsonb_typeof(value) = 'array' THEN
        RETURN COALESCE((SELECT jsonb_agg(pg_temp.provider_config_keys(v)) FROM jsonb_array_elements(value) v), '[]'::jsonb);
    END IF;
    RETURN value;
END $$;

UPDATE providers SET credentials = pg_temp.provider_config_keys(credentials), extra = pg_temp.provider_config_keys(extra)
WHERE credentials::text LIKE '%account%' OR extra::text LIKE '%account%';

-- 设置键冲突让事务失败，避免覆盖已经存在的新配置。
UPDATE settings SET key = replace(key, 'account', 'provider') WHERE key LIKE '%account%' AND key NOT LIKE 'notification_email_template:content_moderation.account_disabled:%';
DO $$
DECLARE item RECORD;
BEGIN
    FOR item IN SELECT id, value FROM settings
        WHERE key IN ('openai_oauth_import_defaults', 'provider_scheduling_thresholds', 'ops_email_notification_config', 'ops_advanced_settings', 'ops_metric_thresholds', 'ops_alert_runtime_settings') AND NULLIF(btrim(value), '') IS NOT NULL LOOP
        UPDATE settings SET value = pg_temp.provider_config_keys(item.value::jsonb)::text WHERE id = item.id;
    END LOOP;
END $$;

-- 模板解析器允许占位符内包含空白；先解码 JSON，再处理 subject/html，保留其它设置。
DO $$
DECLARE
    item RECORD;
    payload JSONB;
    field_name TEXT;
BEGIN
    FOR item IN SELECT id, value FROM settings
        WHERE key LIKE 'notification_email_template:%' AND value LIKE '%account_%' LOOP
        payload := item.value::jsonb;
        FOREACH field_name IN ARRAY ARRAY['subject', 'html'] LOOP
            IF jsonb_typeof(payload -> field_name) = 'string' THEN
                payload := jsonb_set(payload, ARRAY[field_name], to_jsonb(regexp_replace(
                    payload ->> field_name, '\{\{\s*account_(id|name)\s*\}\}', '{{provider_\1}}', 'g')), false);
            END IF;
        END LOOP;
        UPDATE settings SET value = payload::text WHERE id = item.id;
    END LOOP;
END $$;

-- 告警规则是现行配置，保留阈值与启用状态，只改指标名和过滤键。
UPDATE ops_alert_rules SET metric_type = replace(metric_type, 'account', 'provider'), filters = pg_temp.provider_config_keys(filters)
WHERE metric_type LIKE '%account%' OR filters::text LIKE '%account%';
