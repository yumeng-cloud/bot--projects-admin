package seed

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/auth"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

const metaSampleDataSeeded = "sample_data_seeded"

// Run is idempotent: creates admin user, workspace, admin member, Grok Bot, bot roles.
// Demo project / labels / sample issues run only when cfg.SeedSampleData is true
// (env SEED_SAMPLE_DATA=1|true), and sample issues are gated by seed_meta so clearing
// issues never triggers a re-seed.
func Run(db *sql.DB, cfg config.Config) error {
	email := strings.ToLower(strings.TrimSpace(cfg.SeedAdminEmail))
	pass := cfg.SeedAdminPass
	wsName := cfg.SeedWSName

	var userID string
	err := db.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		hash, err := auth.HashPassword(pass)
		if err != nil {
			return err
		}
		userID = idgen.New()
		now := nowUTC()
		_, err = db.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?,?,?,?,?,?)`,
			userID, email, hash, "管理员", now, now)
		if err != nil {
			return fmt.Errorf("seed user: %w", err)
		}
		log.Printf("seed: created admin user %s", email)
	} else if err != nil {
		return err
	}

	slug := "yumeng"
	var wsID string
	err = db.QueryRow(`SELECT id FROM workspaces WHERE slug = ?`, slug).Scan(&wsID)
	if err == sql.ErrNoRows {
		wsID = idgen.New()
		now := nowUTC()
		_, err = db.Exec(`INSERT INTO workspaces (id, name, slug, created_at, updated_at) VALUES (?,?,?,?,?)`,
			wsID, wsName, slug, now, now)
		if err != nil {
			return fmt.Errorf("seed workspace: %w", err)
		}
		log.Printf("seed: created workspace %s", wsName)
	} else if err != nil {
		return err
	}

	var memID string
	err = db.QueryRow(`SELECT id FROM members WHERE workspace_id = ? AND user_id = ?`, wsID, userID).Scan(&memID)
	if err == sql.ErrNoRows {
		memID = idgen.New()
		_, err = db.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,?,?,?,?,?)`,
			memID, wsID, userID, "admin", "管理员", 0, 1)
		if err != nil {
			return fmt.Errorf("seed admin member: %w", err)
		}
		log.Printf("seed: linked admin member")
	} else if err != nil {
		return err
	}

	var botCount int
	_ = db.QueryRow(`SELECT COUNT(1) FROM members WHERE workspace_id = ? AND is_bot = 1 AND display_name = ?`, wsID, "Grok Bot").Scan(&botCount)
	if botCount == 0 {
		botID := idgen.New()
		_, err = db.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,NULL,?,?,?,?)`,
			botID, wsID, "bot", "Grok Bot", 1, 1)
		if err != nil {
			return fmt.Errorf("seed bot: %w", err)
		}
		log.Printf("seed: created Grok Bot member")
	}

	if err := SeedBotRoles(db, wsID); err != nil {
		return fmt.Errorf("seed bot roles: %w", err)
	}

	if cfg.SeedSampleData {
		if err := seedSampleData(db, wsID, memID); err != nil {
			return err
		}
	}

	log.Printf("seed: ok workspace_id=%s admin=%s sample_data=%v", wsID, email, cfg.SeedSampleData)
	return nil
}

func seedSampleData(db *sql.DB, wsID, memID string) error {
	var projID string
	err := db.QueryRow(`SELECT id FROM projects WHERE workspace_id=? AND key_prefix=?`, wsID, "YM").Scan(&projID)
	if err == sql.ErrNoRows {
		projID = idgen.New()
		now := nowUTC()
		_, err = db.Exec(`INSERT INTO projects (id, workspace_id, name, key_prefix, description, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?)`, projID, wsID, "示例项目", "YM", "育蒙科技示例项目", now, now)
		if err != nil {
			return fmt.Errorf("seed project: %w", err)
		}
		_, _ = db.Exec(`INSERT INTO issue_counters (project_id, next_number) VALUES (?, 1)`, projID)
		log.Printf("seed: created sample project YM")
	} else if err != nil {
		return err
	}

	var labelCount int
	_ = db.QueryRow(`SELECT COUNT(1) FROM labels WHERE workspace_id=? AND name=?`, wsID, "前端").Scan(&labelCount)
	if labelCount == 0 {
		now := nowUTC()
		for _, pair := range [][2]string{{"前端", "#1677ff"}, {"后端", "#52c41a"}, {"紧急", "#ff4d4f"}} {
			_, _ = db.Exec(`INSERT INTO labels (id, workspace_id, name, color, created_at) VALUES (?,?,?,?,?)`,
				idgen.New(), wsID, pair[0], pair[1], now)
		}
		log.Printf("seed: created sample labels")
	}

	seeded, err := hasSeedMeta(db, metaSampleDataSeeded)
	if err != nil {
		return fmt.Errorf("seed meta read: %w", err)
	}
	if seeded {
		log.Printf("seed: sample data already marked seeded; skip issues re-seed")
		return nil
	}

	if err := SeedSampleIssues(db, wsID, projID, memID); err != nil {
		return fmt.Errorf("seed sample issues: %w", err)
	}
	if err := setSeedMeta(db, metaSampleDataSeeded, "1"); err != nil {
		return fmt.Errorf("seed meta write: %w", err)
	}
	return nil
}

func hasSeedMeta(db *sql.DB, key string) (bool, error) {
	var v string
	err := db.QueryRow(`SELECT value FROM seed_meta WHERE key = ?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return v == "1" || strings.EqualFold(v, "true"), nil
}

func setSeedMeta(db *sql.DB, key, value string) error {
	now := nowUTC()
	_, err := db.Exec(`INSERT INTO seed_meta (key, value, updated_at) VALUES (?,?,?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, now)
	return err
}

// SeedBotRoles 幂等写入默认协作角色（结构种子，不受 SEED_SAMPLE_DATA 控制）。
func SeedBotRoles(db *sql.DB, workspaceID string) error {
	defaults := []struct {
		name, slug, resp string
	}{
		{"前端", "frontend", "UI/前端实现、组件与交互"},
		{"产品经理", "pm", "需求澄清、优先级与验收标准"},
		{"QA", "qa", "测试计划、回归与缺陷跟踪"},
		{"后端", "backend", "API/服务实现与数据模型"},
		{"知识库", "wiki", "Wiki 文档整理与知识沉淀"},
	}
	for _, d := range defaults {
		var n int
		_ = db.QueryRow(`SELECT COUNT(1) FROM bot_roles WHERE workspace_id = ? AND slug = ?`, workspaceID, d.slug).Scan(&n)
		if n > 0 {
			continue
		}
		id := idgen.New()
		now := nowUTC()
		_, err := db.Exec(`INSERT INTO bot_roles (id, workspace_id, name, slug, responsibilities, skills, default_project_ids, created_at, updated_at)
			VALUES (?,?,?,?,?,'[]','[]',?,?)`, id, workspaceID, d.name, d.slug, d.resp, now, now)
		if err != nil {
			return err
		}
		log.Printf("seed: bot role %s", d.slug)
	}
	return nil
}

// SeedSampleIssues writes cross-column demo issues when called.
// Callers must gate with SEED_SAMPLE_DATA and seed_meta; this no longer re-seeds
// merely because issue count is zero.
func SeedSampleIssues(db *sql.DB, workspaceID, projectID, assigneeID string) error {
	now := nowUTC()
	type row struct {
		num                 int
		title, status, prio string
		est                 int
		labelNames          []string
	}
	rows := []row{
		{1, "看板拖拽改状态", "in_progress", "high", 5, []string{"前端", "紧急"}},
		{2, "列表筛选与空状态", "todo", "medium", 3, []string{"前端"}},
		{3, "Issue 抽屉乐观锁", "in_review", "high", 2, []string{"前端"}},
		{4, "Cycle 容量过载展示", "backlog", "urgent", 8, []string{"后端"}},
		{5, "登录页中文文案", "done", "low", 1, []string{"前端"}},
		{6, "Webhook 重试策略", "todo", "medium", 5, []string{"后端"}},
		{7, "Wiki 搜索 API", "backlog", "high", 5, []string{"后端"}},
		{8, "工时报表导出", "in_progress", "medium", 3, []string{"前端"}},
	}
	labelID := map[string]string{}
	lrows, err := db.Query(`SELECT id, name FROM labels WHERE workspace_id = ?`, workspaceID)
	if err != nil {
		return err
	}
	defer lrows.Close()
	for lrows.Next() {
		var id, name string
		if err := lrows.Scan(&id, &name); err != nil {
			return err
		}
		labelID[name] = id
	}
	for _, r := range rows {
		id := idgen.New()
		ident := fmt.Sprintf("YM-%d", r.num)
		_, err := db.Exec(`INSERT INTO issues (id, project_id, workspace_id, number, identifier, title, description,
			status, priority, estimate, assignee_id, reporter_id, cycle_id, parent_id, board_order, version, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,NULL,NULL,?,?,?,?)`,
			id, projectID, workspaceID, r.num, ident, r.title, "", r.status, r.prio, r.est,
			assigneeID, assigneeID, r.num*100, 1, now, now)
		if err != nil {
			return err
		}
		for _, ln := range r.labelNames {
			if lid, ok := labelID[ln]; ok {
				_, _ = db.Exec(`INSERT INTO issue_labels (issue_id, label_id) VALUES (?,?)`, id, lid)
			}
		}
	}
	_, _ = db.Exec(`UPDATE issue_counters SET next_number = ? WHERE project_id = ?`, len(rows)+1, projectID)
	log.Printf("seed: created %d sample issues across columns", len(rows))
	return nil
}

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339Nano) }
