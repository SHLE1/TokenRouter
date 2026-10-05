-- 用户主动选择的语言供界面和后台通知共用，历史账户从通知偏好迁入。
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_locale VARCHAR(35);
UPDATE users u SET preferred_locale = CASE s.value
    WHEN 'zh' THEN 'zh-Hans'
    WHEN 'zh-Hans' THEN 'zh-Hans'
    WHEN 'en' THEN 'en'
END
FROM settings s
WHERE s.key = 'notification_email_locale:user:' || u.id::text
  AND s.value IN ('zh', 'zh-Hans', 'en') AND u.preferred_locale IS NULL;
