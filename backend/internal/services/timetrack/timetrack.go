package timetrack

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type Service struct {
	DB *sql.DB
}

func New(db *sql.DB) *Service { return &Service{DB: db} }

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func scanEntry(row interface{ Scan(dest ...any) error }) (models.TimeEntry, error) {
	var e models.TimeEntry
	var billable int
	var rate sql.NullFloat64
	err := row.Scan(&e.ID, &e.WorkspaceID, &e.IssueID, &e.MemberID, &e.Hours, &e.Note, &billable, &rate, &e.WorkedAt, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return e, err
	}
	e.Billable = billable != 0
	if rate.Valid {
		v := rate.Float64
		e.Rate = &v
	}
	return e, nil
}

const entryCols = `id, workspace_id, issue_id, member_id, hours, note, billable, rate, worked_at, created_at, updated_at`


// enrichEntries fills issue_identifier / issue_title / member_name (S7: no raw id as primary label).
func (s *Service) enrichEntries(entries []models.TimeEntry) {
	if len(entries) == 0 {
		return
	}
	issueIDs := make([]string, 0, len(entries))
	memberIDs := make([]string, 0, len(entries))
	seenI, seenM := map[string]struct{}{}, map[string]struct{}{}
	for _, e := range entries {
		if e.IssueID != "" {
			if _, ok := seenI[e.IssueID]; !ok {
				seenI[e.IssueID] = struct{}{}
				issueIDs = append(issueIDs, e.IssueID)
			}
		}
		if e.MemberID != "" {
			if _, ok := seenM[e.MemberID]; !ok {
				seenM[e.MemberID] = struct{}{}
				memberIDs = append(memberIDs, e.MemberID)
			}
		}
	}
	idents, titles := map[string]string{}, map[string]string{}
	if len(issueIDs) > 0 {
		ph, args := inPlaceholders(issueIDs)
		rows, err := s.DB.Query(`SELECT id, identifier, title FROM issues WHERE id IN (`+ph+`)`, args...)
		if err == nil {
			for rows.Next() {
				var id, ident, title string
				if rows.Scan(&id, &ident, &title) == nil {
					idents[id] = ident
					titles[id] = title
				}
			}
			_ = rows.Close()
		}
	}
	names := map[string]string{}
	if len(memberIDs) > 0 {
		ph, args := inPlaceholders(memberIDs)
		rows, err := s.DB.Query(`SELECT id, display_name FROM members WHERE id IN (`+ph+`)`, args...)
		if err == nil {
			for rows.Next() {
				var id, name string
				if rows.Scan(&id, &name) == nil {
					names[id] = name
				}
			}
			_ = rows.Close()
		}
	}
	for i := range entries {
		if v, ok := idents[entries[i].IssueID]; ok {
			entries[i].IssueIdent = v
		}
		if v, ok := titles[entries[i].IssueID]; ok {
			entries[i].IssueTitle = v
		}
		if v, ok := names[entries[i].MemberID]; ok {
			entries[i].MemberName = v
		}
	}
}

func inPlaceholders(ids []string) (string, []any) {
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args[i] = id
	}
	return strings.Join(ph, ","), args
}

type CreateEntryInput struct {
	IssueID  string
	MemberID string
	Hours    float64
	Note     string
	Billable *bool
	Rate     *float64
	WorkedAt string
}

func (s *Service) ensureIssue(workspaceID, issueID string) error {
	var id string
	err := s.DB.QueryRow(`SELECT id FROM issues WHERE id = ? AND workspace_id = ?`, issueID, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.NotFound("Issue 不存在")
	}
	return err
}

func (s *Service) CreateEntry(workspaceID string, in CreateEntryInput) (*models.TimeEntry, error) {
	if in.IssueID == "" {
		return nil, apperr.BadRequest("issue_id 不能为空")
	}
	if in.MemberID == "" {
		return nil, apperr.BadRequest("member_id 不能为空")
	}
	if in.Hours <= 0 {
		return nil, apperr.BadRequest("工时必须大于 0")
	}
	if err := s.ensureIssue(workspaceID, in.IssueID); err != nil {
		return nil, err
	}
	workedAt := strings.TrimSpace(in.WorkedAt)
	if workedAt == "" {
		workedAt = time.Now().UTC().Format("2006-01-02")
	}
	billable := 1
	if in.Billable != nil && !*in.Billable {
		billable = 0
	}
	id := idgen.New()
	now := nowUTC()
	var rate any
	if in.Rate != nil {
		rate = *in.Rate
	}
	_, err := s.DB.Exec(`INSERT INTO time_entries
		(id, workspace_id, issue_id, member_id, hours, note, billable, rate, worked_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		id, workspaceID, in.IssueID, in.MemberID, in.Hours, in.Note, billable, rate, workedAt, now, now)
	if err != nil {
		return nil, err
	}
	return s.GetEntry(workspaceID, id)
}

func (s *Service) GetEntry(workspaceID, entryID string) (*models.TimeEntry, error) {
	e, err := scanEntry(s.DB.QueryRow(`SELECT `+entryCols+` FROM time_entries WHERE id = ? AND workspace_id = ?`, entryID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("工时记录不存在")
	}
	if err != nil {
		return nil, err
	}
	out := []models.TimeEntry{e}
	s.enrichEntries(out)
	return &out[0], nil
}

type ListFilter struct {
	IssueID  string
	MemberID string
	From     string
	To       string
	Limit    int
}

func (s *Service) ListEntries(workspaceID string, f ListFilter) ([]models.TimeEntry, error) {
	q := `SELECT ` + entryCols + ` FROM time_entries WHERE workspace_id = ?`
	args := []any{workspaceID}
	if f.IssueID != "" {
		q += ` AND issue_id = ?`
		args = append(args, f.IssueID)
	}
	if f.MemberID != "" {
		q += ` AND member_id = ?`
		args = append(args, f.MemberID)
	}
	if f.From != "" {
		q += ` AND worked_at >= ?`
		args = append(args, f.From)
	}
	if f.To != "" {
		q += ` AND worked_at <= ?`
		args = append(args, f.To)
	}
	q += ` ORDER BY worked_at DESC, created_at DESC`
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	q += fmt.Sprintf(` LIMIT %d`, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.TimeEntry{}
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.enrichEntries(out)
	return out, nil
}

type UpdateEntryInput struct {
	Hours    *float64
	Note     *string
	Billable *bool
	Rate     *float64
	WorkedAt *string
	ClearRate bool
}

func (s *Service) UpdateEntry(workspaceID, entryID string, in UpdateEntryInput) (*models.TimeEntry, error) {
	e, err := s.GetEntry(workspaceID, entryID)
	if err != nil {
		return nil, err
	}
	if in.Hours != nil {
		if *in.Hours <= 0 {
			return nil, apperr.BadRequest("工时必须大于 0")
		}
		e.Hours = *in.Hours
	}
	if in.Note != nil {
		e.Note = *in.Note
	}
	if in.Billable != nil {
		e.Billable = *in.Billable
	}
	if in.ClearRate {
		e.Rate = nil
	} else if in.Rate != nil {
		e.Rate = in.Rate
	}
	if in.WorkedAt != nil {
		e.WorkedAt = strings.TrimSpace(*in.WorkedAt)
	}
	billable := 0
	if e.Billable {
		billable = 1
	}
	var rate any
	if e.Rate != nil {
		rate = *e.Rate
	}
	now := nowUTC()
	_, err = s.DB.Exec(`UPDATE time_entries SET hours=?, note=?, billable=?, rate=?, worked_at=?, updated_at=? WHERE id=? AND workspace_id=?`,
		e.Hours, e.Note, billable, rate, e.WorkedAt, now, entryID, workspaceID)
	if err != nil {
		return nil, err
	}
	return s.GetEntry(workspaceID, entryID)
}

func (s *Service) DeleteEntry(workspaceID, entryID string) error {
	res, err := s.DB.Exec(`DELETE FROM time_entries WHERE id = ? AND workspace_id = ?`, entryID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("工时记录不存在")
	}
	return nil
}

// --- Billing rates ---

func scanRate(row interface{ Scan(dest ...any) error }) (models.BillingRate, error) {
	var r models.BillingRate
	var mid sql.NullString
	err := row.Scan(&r.ID, &r.WorkspaceID, &mid, &r.HourlyRate, &r.Currency, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return r, err
	}
	if mid.Valid {
		r.MemberID = &mid.String
	}
	return r, nil
}

const rateCols = `id, workspace_id, member_id, hourly_rate, currency, created_at, updated_at`

func (s *Service) ListRates(workspaceID string) ([]models.BillingRate, error) {
	rows, err := s.DB.Query(`SELECT `+rateCols+` FROM billing_rates WHERE workspace_id = ? ORDER BY member_id IS NOT NULL, member_id`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.BillingRate{}
	for rows.Next() {
		r, err := scanRate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type UpsertRateInput struct {
	MemberID   *string // nil = workspace default
	HourlyRate float64
	Currency   string
}

func (s *Service) UpsertRate(workspaceID string, in UpsertRateInput) (*models.BillingRate, error) {
	if in.HourlyRate < 0 {
		return nil, apperr.BadRequest("费率不能为负数")
	}
	currency := strings.TrimSpace(in.Currency)
	if currency == "" {
		currency = "CNY"
	}
	now := nowUTC()
	var member any
	if in.MemberID != nil && *in.MemberID != "" {
		member = *in.MemberID
	}

	var existingID string
	var err error
	if member == nil {
		err = s.DB.QueryRow(`SELECT id FROM billing_rates WHERE workspace_id = ? AND member_id IS NULL`, workspaceID).Scan(&existingID)
	} else {
		err = s.DB.QueryRow(`SELECT id FROM billing_rates WHERE workspace_id = ? AND member_id = ?`, workspaceID, member).Scan(&existingID)
	}
	if err == sql.ErrNoRows {
		id := idgen.New()
		_, err = s.DB.Exec(`INSERT INTO billing_rates (id, workspace_id, member_id, hourly_rate, currency, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?)`, id, workspaceID, member, in.HourlyRate, currency, now, now)
		if err != nil {
			return nil, err
		}
		existingID = id
	} else if err != nil {
		return nil, err
	} else {
		_, err = s.DB.Exec(`UPDATE billing_rates SET hourly_rate=?, currency=?, updated_at=? WHERE id=?`,
			in.HourlyRate, currency, now, existingID)
		if err != nil {
			return nil, err
		}
	}
	r, err := scanRate(s.DB.QueryRow(`SELECT `+rateCols+` FROM billing_rates WHERE id = ?`, existingID))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) DeleteRate(workspaceID, rateID string) error {
	res, err := s.DB.Exec(`DELETE FROM billing_rates WHERE id = ? AND workspace_id = ?`, rateID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("费率不存在")
	}
	return nil
}

func (s *Service) memberRate(workspaceID, memberID string) (float64, string) {
	var rate float64
	var currency string
	err := s.DB.QueryRow(`SELECT hourly_rate, currency FROM billing_rates WHERE workspace_id = ? AND member_id = ?`, workspaceID, memberID).
		Scan(&rate, &currency)
	if err == nil {
		return rate, currency
	}
	err = s.DB.QueryRow(`SELECT hourly_rate, currency FROM billing_rates WHERE workspace_id = ? AND member_id IS NULL`, workspaceID).
		Scan(&rate, &currency)
	if err == nil {
		return rate, currency
	}
	return 0, "CNY"
}

func (s *Service) effectiveRate(e models.TimeEntry) (float64, string) {
	if e.Rate != nil {
		_, cur := s.memberRate(e.WorkspaceID, e.MemberID)
		return *e.Rate, cur
	}
	return s.memberRate(e.WorkspaceID, e.MemberID)
}

// Summary 按 project / member / date 汇总。
func (s *Service) Summary(workspaceID, from, to, groupBy string) (*models.BillingSummary, error) {
	groupBy = strings.ToLower(strings.TrimSpace(groupBy))
	if groupBy == "" {
		groupBy = "member"
	}
	if groupBy != "project" && groupBy != "member" && groupBy != "date" {
		return nil, apperr.BadRequest("group_by 须为 project、member 或 date")
	}
	entries, err := s.ListEntries(workspaceID, ListFilter{From: from, To: to, Limit: 10000})
	if err != nil {
		return nil, err
	}

	type agg struct {
		label string
		hours float64
		billableHours float64
		amount float64
		currency string
		n int
	}
	m := map[string]*agg{}
	var totalH, totalA float64
	currency := "CNY"

	for _, e := range entries {
		rate, cur := s.effectiveRate(e)
		currency = cur
		key := ""
		label := ""
		switch groupBy {
		case "member":
			key = e.MemberID
			var name string
			_ = s.DB.QueryRow(`SELECT display_name FROM members WHERE id = ?`, e.MemberID).Scan(&name)
			if name == "" {
				name = e.MemberID
			}
			label = name
		case "date":
			key = e.WorkedAt
			if len(key) > 10 {
				key = key[:10]
			}
			label = key
		case "project":
			var pid, pname string
			_ = s.DB.QueryRow(`SELECT p.id, p.name FROM issues i JOIN projects p ON p.id = i.project_id WHERE i.id = ?`, e.IssueID).
				Scan(&pid, &pname)
			if pid == "" {
				key = "_none"
				label = "(无项目)"
			} else {
				key = pid
				label = pname
			}
		}
		a := m[key]
		if a == nil {
			a = &agg{label: label, currency: cur}
			m[key] = a
		}
		a.hours += e.Hours
		a.n++
		totalH += e.Hours
		if e.Billable {
			amt := e.Hours * rate
			a.billableHours += e.Hours
			a.amount += amt
			totalA += amt
		}
	}

	rows := make([]models.BillingSummaryRow, 0, len(m))
	for k, a := range m {
		rows = append(rows, models.BillingSummaryRow{
			GroupKey:      k,
			GroupLabel:    a.label,
			Hours:         a.hours,
			BillableHours: a.billableHours,
			Amount:        a.amount,
			Currency:      a.currency,
			Entries:       a.n,
		})
	}
	return &models.BillingSummary{
		From:        from,
		To:          to,
		GroupBy:     groupBy,
		Rows:        rows,
		TotalHours:  totalH,
		TotalAmount: totalA,
		Currency:    currency,
	}, nil
}

// ExportCSV 写出工时明细 CSV。
func (s *Service) ExportCSV(w io.Writer, workspaceID, from, to string) error {
	entries, err := s.ListEntries(workspaceID, ListFilter{From: from, To: to, Limit: 10000})
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "worked_at", "issue_id", "issue_identifier", "member_id", "member_name", "hours", "billable", "rate", "amount", "currency", "note"})
	for _, e := range entries {
		rate, cur := s.effectiveRate(e)
		var ident, mname string
		_ = s.DB.QueryRow(`SELECT identifier FROM issues WHERE id = ?`, e.IssueID).Scan(&ident)
		_ = s.DB.QueryRow(`SELECT display_name FROM members WHERE id = ?`, e.MemberID).Scan(&mname)
		amt := 0.0
		if e.Billable {
			amt = e.Hours * rate
		}
		bill := "0"
		if e.Billable {
			bill = "1"
		}
		_ = cw.Write([]string{
			e.ID, e.WorkedAt, e.IssueID, ident, e.MemberID, mname,
			fmt.Sprintf("%.2f", e.Hours), bill, fmt.Sprintf("%.2f", rate), fmt.Sprintf("%.2f", amt), cur, e.Note,
		})
	}
	cw.Flush()
	return cw.Error()
}
