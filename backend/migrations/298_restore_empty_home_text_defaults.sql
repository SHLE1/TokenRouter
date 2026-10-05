-- 291 将历史空标题保存为内容块，这类设置需要继续使用首页内置文案。
-- 修复未编辑、原文语言未知且没有译文的标题和副标题。
WITH repairs AS (
    SELECT s.key,
           jsonb_object_agg(entry.key, entry.value) FILTER (
               WHERE NOT COALESCE((
                   entry.key IN ('site_title', 'site_subtitle')
                   AND entry.value->'source_locale' = 'null'::jsonb
                   AND entry.value->'revision' = '1'::jsonb
                   AND entry.value->'source_revision' = '1'::jsonb
                   AND jsonb_typeof(entry.value->'source') = 'string'
                   AND entry.value->>'source' ~ '^[[:space:]]*$'
                   AND entry.value->'translations' = '{}'::jsonb
               ), false)
           ) AS content
    FROM settings s
    CROSS JOIN LATERAL jsonb_each(s.value::jsonb) entry
    WHERE s.key = 'site_texts'
    GROUP BY s.key
)
UPDATE settings s
SET value = COALESCE(repairs.content, '{}'::jsonb)::text,
    updated_at = now()
FROM repairs
WHERE s.key = repairs.key
  AND s.value::jsonb IS DISTINCT FROM COALESCE(repairs.content, '{}'::jsonb);
