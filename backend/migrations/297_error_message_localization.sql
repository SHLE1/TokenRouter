-- 错误规则的自定义响应文本支持译文，匹配条件和业务错误码保持原值。
ALTER TABLE error_passthrough_rules ADD COLUMN IF NOT EXISTS message_localization JSONB;
UPDATE error_passthrough_rules SET message_localization = jsonb_build_object(
    'source_locale', NULL, 'source', COALESCE(custom_message, ''),
    'translations', '{}'::jsonb, 'revision', 1, 'source_revision', 1
) WHERE message_localization IS NULL;
