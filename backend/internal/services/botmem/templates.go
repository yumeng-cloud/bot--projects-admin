package botmem

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"embed"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

//go:embed templates/*.json
var templateFS embed.FS

// PlaybookRef 席位手册挂载（T7-2）：repo + ref + path；落库为 repo@ref:path。
type PlaybookRef struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
	Path string `json:"path"`
}

// TemplateSeat 内置模板席位行。
type TemplateSeat struct {
	Slug        string      `json:"slug"`
	DisplayName string      `json:"display_name"`
	PlaybookRef PlaybookRef `json:"playbook_ref"`
}

// TeamTemplate 内置班底模板。
type TeamTemplate struct {
	ID          string         `json:"id"`
	Key         string         `json:"key"` // API 对外 = template_key；与 id 同
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Version     int            `json:"version,omitempty"`
	Seats       []TemplateSeat `json:"seats"`
}

var (
	templatesOnce sync.Once
	templatesByID map[string]*TeamTemplate
	templatesList []*TeamTemplate
	templatesErr  error
)

func loadTemplates() {
	templatesOnce.Do(func() {
		templatesByID = map[string]*TeamTemplate{}
		entries, err := templateFS.ReadDir("templates")
		if err != nil {
			templatesErr = err
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			raw, err := templateFS.ReadFile("templates/" + e.Name())
			if err != nil {
				templatesErr = err
				return
			}
			var t TeamTemplate
			if err := json.Unmarshal(raw, &t); err != nil {
				templatesErr = fmt.Errorf("模板 %s: %w", e.Name(), err)
				return
			}
			t.ID = strings.TrimSpace(strings.ToLower(t.ID))
			if t.ID == "" {
				templatesErr = fmt.Errorf("模板 %s 缺少 id", e.Name())
				return
			}
			t.Key = t.ID
			if err := validateTemplate(&t); err != nil {
				templatesErr = err
				return
			}
			cp := t
			templatesByID[t.ID] = &cp
			templatesList = append(templatesList, &cp)
		}
	})
}

func validateTemplate(t *TeamTemplate) error {
	if len(t.Seats) == 0 {
		return fmt.Errorf("模板 %s 席位表为空", t.ID)
	}
	seen := map[string]struct{}{}
	for i, s := range t.Seats {
		slug := strings.TrimSpace(strings.ToLower(s.Slug))
		if slug == "" {
			return fmt.Errorf("模板 %s 第 %d 席 slug 为空", t.ID, i+1)
		}
		if _, ok := seen[slug]; ok {
			return fmt.Errorf("模板 %s 席位 slug 重复: %s", t.ID, slug)
		}
		seen[slug] = struct{}{}
		t.Seats[i].Slug = slug
		if strings.TrimSpace(s.DisplayName) == "" {
			t.Seats[i].DisplayName = slug
		}
		if err := validatePlaybookRef(s.PlaybookRef); err != nil {
			return fmt.Errorf("模板 %s 席位 %s: %w", t.ID, slug, err)
		}
	}
	return nil
}

func validatePlaybookRef(p PlaybookRef) error {
	if strings.TrimSpace(p.Repo) == "" || strings.TrimSpace(p.Ref) == "" || strings.TrimSpace(p.Path) == "" {
		return apperr.BadRequest("playbook_ref 须含 repo、ref、path")
	}
	if strings.Contains(p.Path, "..") {
		return apperr.BadRequest("playbook_ref.path 非法")
	}
	return nil
}

// FormatPlaybookRef 序列化为 repo@ref:path（API.md / 手册索引约定）。
func FormatPlaybookRef(p PlaybookRef) string {
	return strings.TrimSpace(p.Repo) + "@" + strings.TrimSpace(p.Ref) + ":" + strings.TrimSpace(p.Path)
}

// ParsePlaybookRef 解析 playbook_ref；非法 → 400 中文（R2）。
// 接受：repo@ref:path · JSON {repo,ref,path} · 等价纯路径。
func ParsePlaybookRef(raw string) (PlaybookRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PlaybookRef{}, apperr.BadRequest("playbook_ref 不能为空")
	}
	if strings.HasPrefix(raw, "{") {
		var p PlaybookRef
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return PlaybookRef{}, apperr.BadRequest("playbook_ref 格式无效，须为 {repo,ref,path} JSON")
		}
		if err := validatePlaybookRef(p); err != nil {
			return PlaybookRef{}, err
		}
		return p, nil
	}
	// repo@ref:path
	if at := strings.Index(raw, "@"); at > 0 {
		rest := raw[at+1:]
		if colon := strings.Index(rest, ":"); colon > 0 {
			p := PlaybookRef{
				Repo: raw[:at],
				Ref:  rest[:colon],
				Path: rest[colon+1:],
			}
			if err := validatePlaybookRef(p); err != nil {
				return PlaybookRef{}, err
			}
			return p, nil
		}
	}
	// 等价纯路径
	if strings.Contains(raw, "..") || strings.ContainsRune(raw, 0) || strings.ContainsAny(raw, "\n\r") {
		return PlaybookRef{}, apperr.BadRequest("playbook_ref 非法")
	}
	if !strings.Contains(raw, "/") && !strings.HasSuffix(raw, ".md") {
		return PlaybookRef{}, apperr.BadRequest("playbook_ref 无法解析 path")
	}
	return PlaybookRef{Path: raw}, nil
}

// CanonicalPlaybookRefString 解析并规范为落库/出口字符串；空或非法 → 4xx 中文（T7-3）。
// 禁止调用方在失败时默默换成别的硬编码 playbook。
func CanonicalPlaybookRefString(raw string) (string, error) {
	p, err := ParsePlaybookRef(raw)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(p.Repo) != "" && strings.TrimSpace(p.Ref) != "" {
		return FormatPlaybookRef(p), nil
	}
	return strings.TrimSpace(p.Path), nil
}

// ListTeamTemplates 返回内置模板清单（只读）。
func ListTeamTemplates() ([]TeamTemplate, error) {
	loadTemplates()
	if templatesErr != nil {
		return nil, templatesErr
	}
	out := make([]TeamTemplate, 0, len(templatesList))
	for _, t := range templatesList {
		out = append(out, *t)
	}
	return out, nil
}

// GetTeamTemplate 按 id/key 取模板；未知 → 400 中文（R2）。
func GetTeamTemplate(id string) (*TeamTemplate, error) {
	loadTemplates()
	if templatesErr != nil {
		return nil, templatesErr
	}
	key := strings.TrimSpace(strings.ToLower(id))
	if key == "" {
		return nil, apperr.BadRequest("模板 id 不能为空")
	}
	t, ok := templatesByID[key]
	if !ok {
		return nil, apperr.BadRequest(fmt.Sprintf("未知模板：%s", key))
	}
	cp := *t
	return &cp, nil
}

func templateSeatsToEnsure(t *TeamTemplate) []EnsureSeatInput {
	out := make([]EnsureSeatInput, 0, len(t.Seats))
	for _, s := range t.Seats {
		out = append(out, EnsureSeatInput{
			Slug:        s.Slug,
			Name:        s.DisplayName,
			PlaybookRef: FormatPlaybookRef(s.PlaybookRef),
		})
	}
	return out
}

// TemplateSeatDTO API 清单席位（playbook_ref 为约定字符串）。
type TemplateSeatDTO struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	PlaybookRef string `json:"playbook_ref"`
}

// TemplateDTO API 清单项。
type TemplateDTO struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Seats       []TemplateSeatDTO `json:"seats"`
}

// ListTeamTemplateDTOs 清单 API 用 DTO。
func ListTeamTemplateDTOs() ([]TemplateDTO, error) {
	list, err := ListTeamTemplates()
	if err != nil {
		return nil, err
	}
	out := make([]TemplateDTO, 0, len(list))
	for _, t := range list {
		dto := TemplateDTO{Key: t.Key, Name: t.Name, Description: t.Description}
		for _, s := range t.Seats {
			dto.Seats = append(dto.Seats, TemplateSeatDTO{
				Slug: s.Slug, DisplayName: s.DisplayName, PlaybookRef: FormatPlaybookRef(s.PlaybookRef),
			})
		}
		out = append(out, dto)
	}
	return out, nil
}
