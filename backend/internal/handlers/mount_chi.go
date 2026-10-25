package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// chiWS 最小路由面：挂在 /api/v1/workspaces/{workspaceID} 之下。
type chiWS interface {
	Get(pattern string, h http.HandlerFunc)
	Post(pattern string, h http.HandlerFunc)
	Patch(pattern string, h http.HandlerFunc)
	Put(pattern string, h http.HandlerFunc)
	Delete(pattern string, h http.HandlerFunc)
}

func asChi(r chi.Router) chiWS { return chiAdapter{r} }

type chiAdapter struct{ chi.Router }

func (c chiAdapter) Get(pattern string, h http.HandlerFunc)    { c.Router.Get(pattern, h) }
func (c chiAdapter) Post(pattern string, h http.HandlerFunc)   { c.Router.Post(pattern, h) }
func (c chiAdapter) Patch(pattern string, h http.HandlerFunc)  { c.Router.Patch(pattern, h) }
func (c chiAdapter) Put(pattern string, h http.HandlerFunc)    { c.Router.Put(pattern, h) }
func (c chiAdapter) Delete(pattern string, h http.HandlerFunc) { c.Router.Delete(pattern, h) }

// MountWikiChi 挂载 Wiki（路径相对 workspace）。
func MountWikiChi(r chi.Router, db *sql.DB) {
	h := NewWikiHandler(db)
	m := asChi(r)
	m.Get("/wiki/spaces", h.ListSpaces)
	m.Post("/wiki/spaces", h.CreateSpace)
	m.Get("/wiki/spaces/{spaceID}", h.GetSpace)
	m.Patch("/wiki/spaces/{spaceID}", h.UpdateSpace)
	m.Delete("/wiki/spaces/{spaceID}", h.DeleteSpace)
	m.Get("/wiki/spaces/{spaceID}/pages", h.ListPages)
	m.Post("/wiki/spaces/{spaceID}/pages", h.CreatePage)
	m.Get("/wiki/pages/{pageID}", h.GetPage)
	m.Patch("/wiki/pages/{pageID}", h.UpdatePage)
	m.Delete("/wiki/pages/{pageID}", h.DeletePage)
	m.Get("/wiki/search", h.Search)
	m.Get("/wiki/tree", h.Tree)
}

// MountTimeChi 挂载工时与计费。
func MountTimeChi(r chi.Router, db *sql.DB) {
	h := NewTimeHandler(db)
	m := asChi(r)
	m.Get("/time-entries", h.ListEntries)
	m.Post("/time-entries", h.CreateEntry)
	m.Get("/time-entries/{entryID}", h.GetEntry)
	m.Patch("/time-entries/{entryID}", h.UpdateEntry)
	m.Delete("/time-entries/{entryID}", h.DeleteEntry)
	m.Get("/issues/{issueID}/time-entries", h.ListIssueEntries)
	m.Post("/issues/{issueID}/time-entries", h.CreateIssueEntry)
	m.Get("/billing-rates", h.ListRates)
	m.Post("/billing-rates", h.UpsertRate)
	m.Put("/billing-rates", h.UpsertRate)
	m.Delete("/billing-rates/{rateID}", h.DeleteRate)
	m.Get("/billing/summary", h.Summary)
	m.Get("/billing/export.csv", h.ExportCSV)
	m.Get("/time-entries/export.csv", h.ExportCSV)
	m.Get("/time-entries/summary", h.Summary)
}

// MountBotsChi 挂载 Bot 角色 / 记忆 / 待办 / handoff。
func MountBotsChi(r chi.Router, db *sql.DB) {
	h := NewBotsHandler(db)
	m := asChi(r)
	m.Get("/bot-roles", h.ListRoles)
	m.Post("/bot-roles", h.CreateRole)
	m.Post("/bot-roles/{roleID}/claim", h.ClaimRole)
	m.Post("/bot-roles/{roleID}/release", h.ReleaseRole)
	m.Get("/bots", h.Roster)
	// /bots/me 须在 /bots/{memberID}/... 之前
	m.Get("/bots/me", h.Me)
	m.Get("/bots/{memberID}/memories", h.ListMemories)
	m.Post("/bots/{memberID}/memories", h.CreateMemory)
	m.Patch("/bots/{memberID}/memories/{memoryID}", h.PatchMemory)
	m.Delete("/bots/{memberID}/memories/{memoryID}", h.DeleteMemory)
	m.Get("/bots/{memberID}/todos", h.ListTodos)
	m.Post("/bots/{memberID}/todos", h.CreateTodo)
	m.Patch("/bots/{memberID}/todos/{todoID}", h.PatchTodo)
	m.Get("/bots/{memberID}/handoff", h.Handoff)
	m.Post("/bots/{memberID}/handoff", h.ImportHandoff)
}

// MountTeamsChi 挂载 Teams（路径相对 workspace）。
func MountTeamsChi(r chi.Router, db *sql.DB) {
	h := NewTeamHandler(db)
	m := asChi(r)
	m.Post("/teams", h.CreateTeam)
	m.Post("/teams/from-template", h.CreateTeamFromTemplate)
	m.Get("/team-templates", h.ListTeamTemplates)
	m.Get("/teams", h.ListTeams)
	m.Get("/teams/{teamID}", h.GetTeam)
	m.Patch("/teams/{teamID}", h.PatchTeam)
	m.Get("/teams/{teamID}/roster", h.GetRoster)
	m.Post("/teams/{teamID}/roster/ensure", h.EnsureRoster)
	m.Post("/teams/{teamID}/seats", h.CreateSeat)
	m.Post("/teams/{teamID}/seats/{seatID}/bind", h.BindSeat)
	m.Post("/teams/{teamID}/seats/{seatID}/unbind", h.UnbindSeat)
	m.Post("/teams/{teamID}/seats/{seatID}/stop", h.StopSeat)
	m.Post("/teams/{teamID}/seats/expand", h.ExpandSeats)
	m.Post("/teams/{teamID}/expand", h.ExpandSeats)
	m.Get("/teams/{teamID}/invites", h.ListInvites)
	m.Post("/teams/{teamID}/invites", h.CreateInvite)
	m.Post("/teams/{teamID}/invites/{inviteID}/revoke", h.RevokeInvite)
	m.Get("/teams/{teamID}/pins", h.ListPins)
	m.Get("/teams/{teamID}/pins/{key}", h.GetPin)
	m.Put("/teams/{teamID}/pins", h.PutPin)
	m.Put("/teams/{teamID}/pins/{key}", h.PutPin)
	m.Delete("/teams/{teamID}/pins/{key}", h.DeletePin)
}
