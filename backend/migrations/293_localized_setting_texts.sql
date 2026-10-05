-- 普通用户可见的独立文案保存原文和译文，内部字段继续使用现有格式。
WITH fields(key, fallback) AS (
    VALUES ('balance_unit_name', 'USD'), ('balance_low_notify_recharge_url', ''),
           ('oidc_connect_provider_name', ''), ('PAYMENT_HELP_TEXT', ''),
           ('PAYMENT_HELP_IMAGE_URL', ''), ('PRODUCT_NAME_PREFIX', ''),
           ('PRODUCT_NAME_SUFFIX', ''), ('smtp_from_name', '')
)
INSERT INTO settings(key, value, updated_at)
SELECT f.key || '_localized', jsonb_build_object(
    'source_locale', NULL, 'source', COALESCE(NULLIF(s.value, ''), f.fallback),
    'translations', '{}'::jsonb, 'revision', 1, 'source_revision', 1
)::text, now()
FROM fields f LEFT JOIN settings s ON s.key = f.key
ON CONFLICT(key) DO NOTHING;
