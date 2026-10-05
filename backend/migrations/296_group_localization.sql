-- 用户展示名称独立于分组的唯一业务名称。
ALTER TABLE groups ADD COLUMN IF NOT EXISTS localization JSONB;
UPDATE groups SET localization = jsonb_build_object(
    'source_locale', NULL,
    'source', jsonb_build_object('display_name', name, 'description', COALESCE(description, '')),
    'translations', '{}'::jsonb, 'revision', 1, 'source_revision', 1
) WHERE localization IS NULL;
