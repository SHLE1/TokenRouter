//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/TokenFlux/TokenRouter/internal/infra/postgres"
	"github.com/TokenFlux/TokenRouter/internal/pkg/locale"
	"github.com/TokenFlux/TokenRouter/migrations"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// TestUserLocalizationMigrations 检查空库升级、历史内容、语言偏好及事务失败回滚。
func TestUserLocalizationMigrations(t *testing.T) {
	ctx := context.Background()
	pg, err := tcpostgres.Run(ctx, "postgres:18.1-alpine3.23", tcpostgres.WithDatabase("localization"), tcpostgres.WithUsername("postgres"), tcpostgres.WithPassword("postgres"), tcpostgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pg.Terminate(ctx)) })
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, postgres.ApplyMigrations(ctx, db, migrations.FS))
	require.NoError(t, postgres.ApplyMigrations(ctx, db, migrations.FS))
	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.columns WHERE table_name='users' AND column_name='preferred_locale'`).Scan(&count))
	require.Equal(t, 1, count)
	_, err = db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public`)
	require.NoError(t, err)
	prior := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	require.NoError(t, err)
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() >= "290" {
			continue
		}
		data, err := migrations.FS.ReadFile(entry.Name())
		require.NoError(t, err)
		prior[entry.Name()] = &fstest.MapFile{Data: data}
	}
	require.NoError(t, postgres.ApplyMigrations(ctx, db, prior))
	_, err = db.Exec(`INSERT INTO users(id,email,password_hash) VALUES(991,'locale@example.com','hash');
 INSERT INTO announcements(id,title,content) VALUES(991,'公告','正文');
 INSERT INTO announcement_reads(announcement_id,user_id) VALUES(991,991);
 INSERT INTO groups(id,name) VALUES(991,'business-group');
 INSERT INTO settings(key,value) VALUES
 ('site_name','Original'),('site_name_zh','中文站点'),('site_name_en','English site'),
 ('site_title_zh','中文标题'),('site_subtitle','  '),('site_subtitle_zh',''),('site_subtitle_en',''),('notification_email_locale:user:991','zh'),
 ('notification_email_template:auth.verify_code:zh','{"subject":"验证码","html":"<p>正文</p>"}');`)
	require.NoError(t, err)
	// 单个迁移失败时同事务中新增的字段与内容都回滚。
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	migration, err := migrations.FS.ReadFile("290_user_locale.sql")
	require.NoError(t, err)
	_, err = tx.Exec(string(migration))
	require.NoError(t, err)
	_, err = tx.Exec(`SELECT localization_intentional_failure()`)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM information_schema.columns WHERE table_name='users' AND column_name='preferred_locale'`).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, postgres.ApplyMigrations(ctx, db, migrations.FS))
	var language, raw string
	require.NoError(t, db.QueryRow(`SELECT preferred_locale FROM users WHERE id=991`).Scan(&language))
	require.Equal(t, "zh-Hans", language)
	require.NoError(t, db.QueryRow(`SELECT value FROM settings WHERE key='site_texts'`).Scan(&raw))
	var texts map[string]locale.Content[string]
	require.NoError(t, json.Unmarshal([]byte(raw), &texts))
	require.Nil(t, texts["site_name"].SourceLocale)
	require.Equal(t, "Original", texts["site_name"].Source)
	for code, expected := range map[string]string{"en": "English site", "zh-Hans": "中文站点"} {
		value, info := texts["site_name"].Resolve(code)
		require.Equal(t, expected, value)
		require.False(t, info.Fallback)
	}
	require.NotContains(t, texts, "site_subtitle")
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM announcement_reads WHERE user_id=991 AND announcement_id=991`).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, db.QueryRow(`SELECT localization->'source'->>'title' FROM announcements WHERE id=991`).Scan(&raw))
	require.Equal(t, "公告", raw)
	require.NoError(t, db.QueryRow(`SELECT localization->'source'->>'display_name' FROM groups WHERE id=991`).Scan(&raw))
	require.Equal(t, "business-group", raw)
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM settings WHERE key IN ('site_name_zh','site_name_en','notification_email_template:auth.verify_code:zh')`).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM settings WHERE key='notification_email_template:auth.verify_code:zh-Hans'`).Scan(&count))
	require.Equal(t, 1, count)
}

// TestRepairEmptyHomeTextMigration 检查已升级站点的空标题修复及管理员保存值的保护。
func TestRepairEmptyHomeTextMigration(t *testing.T) {
	ctx := context.Background()
	pg, err := tcpostgres.Run(ctx, "postgres:18.1-alpine3.23", tcpostgres.WithDatabase("home_text"), tcpostgres.WithUsername("postgres"), tcpostgres.WithPassword("postgres"), tcpostgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pg.Terminate(ctx)) })
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	_, err = db.Exec(`CREATE TABLE settings(key text PRIMARY KEY, value text NOT NULL, updated_at timestamptz DEFAULT now()); INSERT INTO settings(key,value) VALUES('unrelated','plain text')`)
	require.NoError(t, err)
	migration, err := migrations.FS.ReadFile("298_restore_empty_home_text_defaults.sql")
	require.NoError(t, err)
	en := "en"
	for _, tc := range []struct {
		name, source string
		language     *string
		revision     int64
		translations map[string]locale.Translation[string]
		keep         bool
	}{
		{name: "empty", revision: 1},
		{name: "whitespace", source: " \t\n", revision: 1},
		{name: "custom", source: "Custom title", revision: 1, keep: true},
		{name: "translated", revision: 1, translations: map[string]locale.Translation[string]{"en": {Value: "English title", SourceRevision: 1}}, keep: true},
		{name: "intentional blank", language: &en, revision: 1, keep: true},
		{name: "already edited", revision: 2, keep: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.translations == nil {
				tc.translations = map[string]locale.Translation[string]{}
			}
			document := locale.Content[string]{Source: tc.source, SourceLocale: tc.language, Revision: tc.revision, SourceRevision: 1, Translations: tc.translations}
			input := map[string]any{"site_title": document, "site_subtitle": document, "contact_info": map[string]string{"source": "Contact us"}, "site_name": map[string]string{"source": "Brand"}}
			encoded, err := json.Marshal(input)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO settings(key,value) VALUES('site_texts',$1) ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value`, string(encoded))
			require.NoError(t, err)
			for range 2 {
				_, err = db.Exec(string(migration))
				require.NoError(t, err)
			}
			var actual string
			require.NoError(t, db.QueryRow(`SELECT value FROM settings WHERE key='site_texts'`).Scan(&actual))
			if !tc.keep {
				delete(input, "site_title")
				delete(input, "site_subtitle")
			}
			expected, err := json.Marshal(input)
			require.NoError(t, err)
			require.JSONEq(t, string(expected), actual)
		})
	}
	// 元数据不完整时，迁移保持现值供人工处理。
	_, err = db.Exec(`UPDATE settings SET value='{"site_title":{"source":""},"site_subtitle":{"source":"existing"}}' WHERE key='site_texts'`)
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	var actual string
	require.NoError(t, db.QueryRow(`SELECT value FROM settings WHERE key='site_texts'`).Scan(&actual))
	require.JSONEq(t, `{"site_title":{"source":""},"site_subtitle":{"source":"existing"}}`, actual)
}
