package modules

import (
	"database/sql"
	"net/http"
)

// Optional domain registrars — other bots assign these at init time.
// Main calls them when non-nil so missing modules never break the build.

type Deps struct {
	DB      *sql.DB
	AuthMW  func(http.Handler) http.Handler // JWT or API Key
	WSMemMW func(http.Handler) http.Handler // RequireWorkspaceMember
}

// RegisterFunc mounts domain routes under /api/v1.
type RegisterFunc func(mux Mux, deps Deps)

// Mux is the minimal surface used by optional modules (chi or std compatible).
type Mux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

var (
	RegisterWiki     RegisterFunc
	RegisterTime     RegisterFunc
	RegisterBotMem   RegisterFunc
	RegisterIssues   RegisterFunc // Issue/Board domain
	RegisterCollab   RegisterFunc // comments/activity extras if separated
)
