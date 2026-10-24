package wiki

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"context"
	"database/sql"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type Service struct {
	DB    *sql.DB
	Embed embedder.Embedder // optional; set via WithEmbedder
	Store *embedder.Store   // optional; Postgres+pgvector
}

func New(db *sql.DB) *Service { return &Service{DB: db} }

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339Nano) }

var nonSlug = regexp.MustCompile(`[^a-z0-9\p{Han}-]+`)

func Slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return '-'
		}
		return r
	}, s)
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" {
		s = "page-" + idgen.New()[:8]
	}
	return s
}

func nullStr(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

func scanSpace(row interface{ Scan(dest ...any) error }) (models.WikiSpace, error) {
	var s models.WikiSpace
	err := row.Scan(&s.ID, &s.WorkspaceID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func scanPage(row interface{ Scan(dest ...any) error }) (models.WikiPage, error) {
	var p models.WikiPage
	var parent, project, issue, createdBy, updatedBy sql.NullString
	err := row.Scan(
		&p.ID, &p.SpaceID, &p.WorkspaceID, &parent, &p.Title, &p.Slug, &p.BodyMD,
		&project, &issue, &createdBy, &updatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}
	if parent.Valid {
		p.ParentID = &parent.String
	}
	if project.Valid {
		p.ProjectID = &project.String
	}
	if issue.Valid {
		p.IssueID = &issue.String
	}
	if createdBy.Valid {
		p.CreatedBy = &createdBy.String
	}
	if updatedBy.Valid {
		p.UpdatedBy = &updatedBy.String
	}
	return p, nil
}

const spaceCols = `id, workspace_id, name, slug, description, created_at, updated_at`
const pageCols = `id, space_id, workspace_id, parent_id, title, slug, body_md, project_id, issue_id, created_by, updated_by, created_at, updated_at`

// --- Spaces ---

type CreateSpaceInput struct {
	Name        string
	Slug        string
	Description string
}

func (s *Service) CreateSpace(workspaceID string, in CreateSpaceInput) (*models.WikiSpace, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, apperr.BadRequest("空间名称不能为空")
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = Slugify(in.Name)
	} else {
		slug = Slugify(slug)
	}
	id := idgen.New()
	now := nowUTC()
	_, err := s.DB.Exec(`INSERT INTO wiki_spaces (id, workspace_id, name, slug, description, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?)`, id, workspaceID, in.Name, slug, in.Description, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该 slug 已存在")
		}
		return nil, err
	}
	return s.GetSpace(workspaceID, id)
}

func (s *Service) ListSpaces(workspaceID string) ([]models.WikiSpace, error) {
	rows, err := s.DB.Query(`SELECT `+spaceCols+` FROM wiki_spaces WHERE workspace_id = ? ORDER BY name ASC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WikiSpace{}
	for rows.Next() {
		sp, err := scanSpace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *Service) GetSpace(workspaceID, spaceID string) (*models.WikiSpace, error) {
	sp, err := scanSpace(s.DB.QueryRow(`SELECT `+spaceCols+` FROM wiki_spaces WHERE id = ? AND workspace_id = ?`, spaceID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("知识库空间不存在")
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

type UpdateSpaceInput struct {
	Name        *string
	Slug        *string
	Description *string
}

func (s *Service) UpdateSpace(workspaceID, spaceID string, in UpdateSpaceInput) (*models.WikiSpace, error) {
	sp, err := s.GetSpace(workspaceID, spaceID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return nil, apperr.BadRequest("空间名称不能为空")
		}
		sp.Name = n
	}
	if in.Slug != nil {
		sp.Slug = Slugify(*in.Slug)
	}
	if in.Description != nil {
		sp.Description = *in.Description
	}
	now := nowUTC()
	_, err = s.DB.Exec(`UPDATE wiki_spaces SET name=?, slug=?, description=?, updated_at=? WHERE id=? AND workspace_id=?`,
		sp.Name, sp.Slug, sp.Description, now, spaceID, workspaceID)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该 slug 已存在")
		}
		return nil, err
	}
	return s.GetSpace(workspaceID, spaceID)
}

func (s *Service) DeleteSpace(workspaceID, spaceID string) error {
	if s.Store != nil {
		rows, qerr := s.DB.Query(`SELECT id FROM wiki_pages WHERE workspace_id = ? AND space_id = ?`, workspaceID, spaceID)
		if qerr == nil {
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					s.dropPageIndex(context.Background(), workspaceID, id)
				}
			}
			_ = rows.Close()
		}
	}
	res, err := s.DB.Exec(`DELETE FROM wiki_spaces WHERE id = ? AND workspace_id = ?`, spaceID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("知识库空间不存在")
	}
	return nil
}

// --- Pages ---

type CreatePageInput struct {
	Title     string
	Slug      string
	BodyMD    string
	ParentID  *string
	ProjectID *string
	IssueID   *string
	MemberID  string
}

func (s *Service) pageSlugTaken(spaceID, slug string) (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(1) FROM wiki_pages WHERE space_id = ? AND slug = ?`, spaceID, slug).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// uniqueAutoPageSlug makes an auto-generated slug unique within a space by appending
// "-" + the first 8 chars of the page id when the base slug is already taken.
func (s *Service) uniqueAutoPageSlug(spaceID, base, pageID string) (string, error) {
	taken, err := s.pageSlugTaken(spaceID, base)
	if err != nil {
		return "", err
	}
	if !taken {
		return base, nil
	}
	suf := pageID
	if len(suf) > 8 {
		suf = suf[:8]
	}
	cand := base + "-" + suf
	taken, err = s.pageSlugTaken(spaceID, cand)
	if err != nil {
		return "", err
	}
	if !taken {
		return cand, nil
	}
	// Extremely rare: full id suffix.
	return base + "-" + pageID, nil
}

func (s *Service) CreatePage(workspaceID, spaceID string, in CreatePageInput) (*models.WikiPage, error) {
	if _, err := s.GetSpace(workspaceID, spaceID); err != nil {
		return nil, err
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		in.Title = "无标题"
	}
	explicitSlug := strings.TrimSpace(in.Slug) != ""
	id := idgen.New()
	var slug string
	if explicitSlug {
		slug = Slugify(in.Slug)
	} else {
		var err error
		slug, err = s.uniqueAutoPageSlug(spaceID, Slugify(in.Title), id)
		if err != nil {
			return nil, err
		}
	}
	if in.ParentID != nil && *in.ParentID != "" {
		if _, err := s.GetPage(workspaceID, *in.ParentID); err != nil {
			return nil, apperr.BadRequest("父页面不存在")
		}
	}
	now := nowUTC()
	var member any
	if in.MemberID != "" {
		member = in.MemberID
	}
	_, err := s.DB.Exec(`INSERT INTO wiki_pages
		(id, space_id, workspace_id, parent_id, title, slug, body_md, project_id, issue_id, created_by, updated_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, spaceID, workspaceID, nullStr(in.ParentID), in.Title, slug, in.BodyMD,
		nullStr(in.ProjectID), nullStr(in.IssueID), member, member, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该页面 slug 已存在")
		}
		return nil, err
	}
	s.indexPage(context.Background(), workspaceID, id, in.Title, in.BodyMD)
	return s.GetPage(workspaceID, id)
}

func (s *Service) ListPages(workspaceID, spaceID string) ([]models.WikiPage, error) {
	if _, err := s.GetSpace(workspaceID, spaceID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`SELECT `+pageCols+` FROM wiki_pages WHERE workspace_id = ? AND space_id = ? ORDER BY title ASC`,
		workspaceID, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WikiPage{}
	for rows.Next() {
		p, err := scanPage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) GetPage(workspaceID, pageID string) (*models.WikiPage, error) {
	p, err := scanPage(s.DB.QueryRow(`SELECT `+pageCols+` FROM wiki_pages WHERE id = ? AND workspace_id = ?`, pageID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("页面不存在")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type UpdatePageInput struct {
	Title        *string
	Slug         *string
	BodyMD       *string
	ParentID     *string // empty string clears
	ProjectID    *string
	IssueID      *string
	MemberID     string
	ClearParent  bool
	ClearProject bool
	ClearIssue   bool
}

func (s *Service) UpdatePage(workspaceID, pageID string, in UpdatePageInput) (*models.WikiPage, error) {
	p, err := s.GetPage(workspaceID, pageID)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			t = "无标题"
		}
		p.Title = t
	}
	if in.Slug != nil {
		p.Slug = Slugify(*in.Slug)
	}
	if in.BodyMD != nil {
		p.BodyMD = *in.BodyMD
	}
	if in.ClearParent {
		p.ParentID = nil
	} else if in.ParentID != nil {
		if *in.ParentID == pageID {
			return nil, apperr.BadRequest("不能将页面设为自己的父页面")
		}
		if *in.ParentID != "" {
			if _, err := s.GetPage(workspaceID, *in.ParentID); err != nil {
				return nil, apperr.BadRequest("父页面不存在")
			}
			p.ParentID = in.ParentID
		} else {
			p.ParentID = nil
		}
	}
	if in.ClearProject {
		p.ProjectID = nil
	} else if in.ProjectID != nil {
		p.ProjectID = in.ProjectID
	}
	if in.ClearIssue {
		p.IssueID = nil
	} else if in.IssueID != nil {
		p.IssueID = in.IssueID
	}
	now := nowUTC()
	var member any
	if in.MemberID != "" {
		member = in.MemberID
	} else if p.UpdatedBy != nil {
		member = *p.UpdatedBy
	}
	_, err = s.DB.Exec(`UPDATE wiki_pages SET title=?, slug=?, body_md=?, parent_id=?, project_id=?, issue_id=?, updated_by=?, updated_at=?
		WHERE id=? AND workspace_id=?`,
		p.Title, p.Slug, p.BodyMD, nullStr(p.ParentID), nullStr(p.ProjectID), nullStr(p.IssueID), member, now, pageID, workspaceID)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该页面 slug 已存在")
		}
		return nil, err
	}
	s.indexPage(context.Background(), workspaceID, pageID, p.Title, p.BodyMD)
	return s.GetPage(workspaceID, pageID)
}

func (s *Service) DeletePage(workspaceID, pageID string) error {
	res, err := s.DB.Exec(`DELETE FROM wiki_pages WHERE id = ? AND workspace_id = ?`, pageID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("页面不存在")
	}
	s.dropPageIndex(context.Background(), workspaceID, pageID)
	return nil
}

// Search 按 title/body 模糊搜索。
// LIKE 为开发兜底（SQLite/Postgres 均可）；Postgres + pgvector 语义召回由知识库经 embedder.Store 接入，接口可后续换后端。
func (s *Service) Search(workspaceID, q string, limit int) ([]models.WikiPage, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, apperr.BadRequest("搜索关键词不能为空")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	like := "%" + q + "%"
	rows, err := s.DB.Query(`SELECT `+pageCols+` FROM wiki_pages
		WHERE workspace_id = ? AND (title LIKE ? OR body_md LIKE ?)
		ORDER BY updated_at DESC LIMIT ?`, workspaceID, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WikiPage{}
	for rows.Next() {
		p, err := scanPage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Tree 返回某 space 的树形目录。
func (s *Service) Tree(workspaceID, spaceID string) ([]models.WikiTreeNode, error) {
	if spaceID == "" {
		return nil, apperr.BadRequest("需要 space_id")
	}
	if _, err := s.GetSpace(workspaceID, spaceID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`SELECT id, title, slug, parent_id FROM wiki_pages WHERE workspace_id = ? AND space_id = ? ORDER BY title ASC`,
		workspaceID, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type flat struct {
		id, title, slug string
		parent          sql.NullString
	}
	var flats []flat
	byID := map[string]*models.WikiTreeNode{}
	for rows.Next() {
		var f flat
		if err := rows.Scan(&f.id, &f.title, &f.slug, &f.parent); err != nil {
			return nil, err
		}
		flats = append(flats, f)
		node := &models.WikiTreeNode{ID: f.id, Title: f.title, Slug: f.slug, Children: []models.WikiTreeNode{}}
		if f.parent.Valid {
			node.ParentID = &f.parent.String
		}
		byID[f.id] = node
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	roots := []models.WikiTreeNode{}
	// Attach children; build in second pass using pointers then convert roots
	childMap := map[string][]string{}
	var rootIDs []string
	for _, f := range flats {
		if f.parent.Valid && f.parent.String != "" {
			if _, ok := byID[f.parent.String]; ok {
				childMap[f.parent.String] = append(childMap[f.parent.String], f.id)
				continue
			}
		}
		rootIDs = append(rootIDs, f.id)
	}
	var build func(id string) models.WikiTreeNode
	build = func(id string) models.WikiTreeNode {
		n := *byID[id]
		n.Children = []models.WikiTreeNode{}
		for _, cid := range childMap[id] {
			n.Children = append(n.Children, build(cid))
		}
		return n
	}
	for _, id := range rootIDs {
		roots = append(roots, build(id))
	}
	return roots, nil
}
