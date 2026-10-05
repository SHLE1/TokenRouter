-- 套餐译文与价格和权益存放在同一业务对象中。
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS localization JSONB;
UPDATE subscription_plans SET localization = jsonb_build_object(
    'source_locale', NULL,
    'source', jsonb_build_object('name', name, 'description', description, 'features', features, 'product_name', product_name),
    'translations', '{}'::jsonb, 'revision', 1, 'source_revision', 1
) WHERE localization IS NULL;
