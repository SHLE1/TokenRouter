-- 公告的所有语言共用同一条公告和已读记录。
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS localization JSONB;
UPDATE announcements SET localization = jsonb_build_object(
    'source_locale', NULL, 'source', jsonb_build_object('title', title, 'content', content),
    'translations', '{}'::jsonb, 'revision', 1, 'source_revision', 1
) WHERE localization IS NULL;
