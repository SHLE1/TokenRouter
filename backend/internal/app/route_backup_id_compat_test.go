package app

import backuphttp "github.com/TokenFlux/TokenRouter/internal/backup/httpapi"

// 兼容路径测试使用备份 HTTP 适配器的 ID 校验。
var requireCanonicalBackupID = backuphttp.RequireCanonicalBackupID
