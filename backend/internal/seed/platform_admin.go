package seed

import (
	"database/sql"
	"log"
	"strings"
	"time"
)

// BootstrapPlatformAdmins elevates users whose emails appear in PLATFORM_ADMIN_EMAILS
// (comma-separated). Missing emails are skipped (user must register first). Idempotent.
func BootstrapPlatformAdmins(db *sql.DB, emails []string) error {
	if len(emails) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, raw := range emails {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" {
			continue
		}
		res, err := db.Exec(`UPDATE users SET is_platform_admin = 1, updated_at = ? WHERE lower(email) = ?`, now, email)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			log.Printf("platform admin bootstrap: 用户尚未注册，跳过 %s", email)
			continue
		}
		log.Printf("platform admin bootstrap: 已授予超管 %s", email)
	}
	return nil
}
