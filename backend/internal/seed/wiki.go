package seed

import (
	"database/sql"
	"log"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/wiki"
)

// Wiki 幂等写入默认知识库空间与约定页，减少 AI 协作跑偏。
func Wiki(db *sql.DB, workspaceID, memberID string) error {
	svc := wiki.New(db)
	spaces, err := svc.ListSpaces(workspaceID)
	if err != nil {
		return err
	}
	var spaceID string
	for _, sp := range spaces {
		if sp.Slug == "handbook" {
			spaceID = sp.ID
			break
		}
	}
	if spaceID == "" {
		sp, err := svc.CreateSpace(workspaceID, wiki.CreateSpaceInput{
			Name:        "协作手册",
			Slug:        "handbook",
			Description: "产品约定、状态机与 Bot 规范（种子）",
		})
		if err != nil {
			return err
		}
		spaceID = sp.ID
		log.Printf("seed wiki: created space handbook")
	}

	pages := []struct {
		slug  string
		title string
		body  string
	}{
		{"product-conventions", "产品约定", productConventionsMD},
		{"issue-state-machine", "Issue 状态机", issueStateMachineMD},
		{"bot-guidelines", "Bot 协作规范", botGuidelinesMD},
	}
	existing, err := svc.ListPages(workspaceID, spaceID)
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, p := range existing {
		have[p.Slug] = true
	}
	for _, pg := range pages {
		if have[pg.slug] {
			continue
		}
		_, err := svc.CreatePage(workspaceID, spaceID, wiki.CreatePageInput{
			Title: pg.title, Slug: pg.slug, BodyMD: strings.TrimSpace(pg.body) + "\n", MemberID: memberID,
		})
		if err != nil {
			return err
		}
		log.Printf("seed wiki: page %s", pg.slug)
	}
	return nil
}

// EnsureWikiForSlug 供外部在已知 workspace slug 时调用。
func EnsureWikiForSlug(db *sql.DB, wsSlug string) error {
	var wsID, memID string
	err := db.QueryRow(`SELECT w.id, m.id FROM workspaces w
		JOIN members m ON m.workspace_id = w.id AND m.role = 'admin'
		WHERE w.slug = ? LIMIT 1`, wsSlug).Scan(&wsID, &memID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	return Wiki(db, wsID, memID)
}


const productConventionsMD = `# 产品约定

## 目标
育蒙「AI 协作看板」支持多 Bot + 人类在同一工作空间协作 Issue、Cycle、知识库与工时。

## 原则
1. **同一套 API**：人与 Bot 均走 Bearer（JWT 或 API Key），返回 JSON；Wiki body 为 Markdown。
2. **工作区隔离**：所有资源挂在 workspace 下；跨空间操作必须显式授权。
3. **可追溯**：写操作尽量带活动日志；Bot 写操作支持 Idempotency-Key。
4. **乐观协作**：Issue 更新带 version；冲突时返回 409，由客户端重试。
5. **中文优先**：用户可见错误信息使用中文（apperr）。

## 字段风格
- JSON：snake_case
- 时间：UTC ISO-8601（存库）；展示层按 Asia/Shanghai
`

const issueStateMachineMD = `# Issue 状态机

合法状态：

| 状态 | 含义 |
|------|------|
| backlog | 待规划 |
| todo | 已排期待做 |
| in_progress | 进行中 |
| in_review | 评审/验收中 |
| done | 完成 |
| canceled | 取消 |

## 推荐流转
backlog → todo → in_progress → in_review → done

也可：
- todo / in_progress / in_review → canceled
- done 一般不回流；特殊情况经人工改回 in_progress

## Bot 注意
- 改状态时写清原因（comment 或 activity payload）
- 不要跳过 in_review 直接 done（除非工作区约定允许）
- 认领 Issue 前检查 assignee 是否空闲
`

const botGuidelinesMD = `# Bot 协作规范

## 身份
- Bot 在 members 表中 is_bot=1，role=bot
- 使用 API Key（Bearer）调用与人类相同的 REST

## 行为准则
1. **先读 Wiki**：动手前搜索「产品约定」「状态机」「Bot 规范」及项目相关页
2. **小步提交**：一次只改一个 Issue 或一小批相关变更；写操作带 Idempotency-Key
3. **不越权**：不删他人数据；不修改 billing_rates（除非角色允许）
4. **记工时**：完成可计费工作后写 time_entries（hours、note、worked_at）
5. **交接**：长任务更新 bot_todos / memories，并在 handoff 中留下上下文

## 禁止
- 伪造用户身份
- 绕过鉴权或泄露密钥
- 大面积改 schema / 核心看板文件（按 COORDINATION.md 归属）
`
