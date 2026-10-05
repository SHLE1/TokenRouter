-- 将历史中文模板与邮箱偏好迁入共用语言代码。
INSERT INTO settings(key, value, updated_at)
SELECT regexp_replace(key, ':zh$', ':zh-Hans'), value, updated_at
FROM settings WHERE key LIKE 'notification_email_template:%:zh'
ON CONFLICT(key) DO NOTHING;
DELETE FROM settings WHERE key LIKE 'notification_email_template:%:zh';
UPDATE settings SET value = 'zh-Hans'
WHERE key LIKE 'notification_email_locale:%' AND value = 'zh';
