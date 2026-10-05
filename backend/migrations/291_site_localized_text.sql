-- 将站点文案迁入同一结构，语言键成为数据，历史原文和两个语言版本分别保存。
WITH fields(name, fallback) AS (
    VALUES ('site_name', 'TokenRouter'), ('site_title', ''), ('site_subtitle', ''),
           ('contact_info', ''), ('doc_url', ''), ('home_content', ''),
           ('purchase_subscription_url', ''), ('footer_text', '')
), copies AS (
    SELECT f.name,
           COALESCE(NULLIF(b.value, ''), NULLIF(z.value, ''), NULLIF(e.value, ''), f.fallback) AS original,
           z.value AS zh, e.value AS en,
           (b.key IS NOT NULL OR z.key IS NOT NULL OR e.key IS NOT NULL) AS configured
    FROM fields f
    LEFT JOIN settings b ON b.key = f.name
    LEFT JOIN settings z ON z.key = f.name || '_zh'
    LEFT JOIN settings e ON e.key = f.name || '_en'
), payload AS (
    SELECT jsonb_object_agg(name, jsonb_build_object(
        'source_locale', NULL, 'source', original, 'revision', 1, 'source_revision', 1,
        'translations',
            CASE WHEN COALESCE(trim(zh), '') <> '' THEN jsonb_build_object('zh-Hans', jsonb_build_object('value', zh, 'source_revision', 1)) ELSE '{}'::jsonb END ||
            CASE WHEN COALESCE(trim(en), '') <> '' THEN jsonb_build_object('en', jsonb_build_object('value', en, 'source_revision', 1)) ELSE '{}'::jsonb END
    )) FILTER (WHERE configured) AS value FROM copies
)
INSERT INTO settings (key, value, updated_at)
SELECT 'site_texts', COALESCE(value, '{}'::jsonb)::text, now() FROM payload
ON CONFLICT (key) DO NOTHING;

INSERT INTO settings (key, value, updated_at) VALUES ('default_locale', 'en', now()) ON CONFLICT (key) DO NOTHING;
DELETE FROM settings WHERE key IN ('site_name_zh', 'site_name_en', 'site_title_zh', 'site_title_en', 'site_subtitle_zh', 'site_subtitle_en');
