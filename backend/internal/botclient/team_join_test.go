package botclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newJoinTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		Cfg: Config{
			BaseURL:     srv.URL,
			APIKey:      "ymk_test",
			WorkspaceID: "ws-old",
			MemberID:    "member-me",
		},
		HTTP: srv.Client(),
	}
}

func TestClassifySeatOccupancy(t *testing.T) {
	empty := map[string]any{"role_slug": "frontend", "claim": nil}
	if g := ClassifySeatOccupancy(empty, "me"); g != "empty" {
		t.Fatalf("empty got %s", g)
	}
	selfBound := map[string]any{"role_slug": "pm", "bound_member_id": "me"}
	if g := ClassifySeatOccupancy(selfBound, "me"); g != "self" {
		t.Fatalf("self got %s", g)
	}
	activeClaim := map[string]any{
		"role_slug": "qa",
		"claim":     map[string]any{"member_id": "other", "claimed_at": "2026-10-24T01:00:00Z"},
	}
	if g := ClassifySeatOccupancy(activeClaim, "me"); g != "active" {
		t.Fatalf("active got %s", g)
	}
	stopped := map[string]any{
		"role_slug":          "backend",
		"status":             "stopped",
		"bound_member_id":    "other",
		"bound_display_name": "OldBot",
	}
	if g := ClassifySeatOccupancy(stopped, "me"); g != "stopped" {
		t.Fatalf("stopped got %s", g)
	}
	if SeatIsActiveOccupancy("active") != true || SeatIsActiveOccupancy("stopped") {
		t.Fatal("active occupancy helper")
	}
}

func TestPickJoinSeatPreferEmptyNoForceActive(t *testing.T) {
	seats := []map[string]any{
		{"id": "s1", "role_slug": "frontend", "bound_member_id": "other", "claim": map[string]any{"member_id": "other", "claimed_at": "t"}},
		{"id": "s2", "role_slug": "pm", "claim": nil},
	}
	seat, slug, self, create, err := pickJoinSeat(seats, "me", TeamJoinOpts{}, false, "")
	if err != nil || create || self || slug != "pm" || strField(seat, "id") != "s2" {
		t.Fatalf("prefer empty: seat=%v slug=%s self=%v create=%v err=%v", seat, slug, self, create, err)
	}

	_, _, _, _, err = pickJoinSeat(seats, "me", TeamJoinOpts{Role: "frontend"}, false, "")
	if err == nil {
		t.Fatal("expected 409 on active role")
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 409 || !strings.Contains(ae.Message, "禁止顶替") {
		t.Fatalf("want 409 禁止顶替 got %v", err)
	}
}

func TestPickJoinSeatAllowExpand(t *testing.T) {
	seats := []map[string]any{
		{"id": "s1", "role_slug": "frontend", "bound_member_id": "a", "claim": map[string]any{"member_id": "a"}},
	}
	_, _, _, _, err := pickJoinSeat(seats, "me", TeamJoinOpts{}, false, "")
	if err == nil || !strings.Contains(err.Error(), "未允许加新席") {
		t.Fatalf("want expand denied: %v", err)
	}
	_, slug, _, create, err := pickJoinSeat(seats, "me", TeamJoinOpts{NewSeat: "frontend-2"}, true, "hint")
	if err != nil || !create || slug != "frontend-2" {
		t.Fatalf("expand new-seat: slug=%s create=%v err=%v", slug, create, err)
	}
	_, slug, _, create, err = pickJoinSeat(seats, "me", TeamJoinOpts{}, true, "role_hint_x")
	if err != nil || !create || slug != "role_hint_x" {
		t.Fatalf("expand hint: slug=%s create=%v err=%v", slug, create, err)
	}
	_, slug, _, create, err = pickJoinSeat(seats, "me", TeamJoinOpts{}, true, "")
	if err != nil || !create || slug != "seat-1" {
		t.Fatalf("expand auto: slug=%s create=%v err=%v", slug, create, err)
	}
}

func TestPickJoinSeatIdempotentSelf(t *testing.T) {
	seats := []map[string]any{
		{"id": "s1", "role_slug": "wiki", "bound_member_id": "me"},
	}
	seat, slug, self, create, err := pickJoinSeat(seats, "me", TeamJoinOpts{Role: "wiki"}, false, "")
	if err != nil || !self || create || slug != "wiki" || seat == nil {
		t.Fatalf("self role: self=%v create=%v slug=%s err=%v", self, create, slug, err)
	}
	// 无 --role 且无空席但已绑本 member
	seats2 := []map[string]any{
		{"id": "s1", "role_slug": "wiki", "bound_member_id": "me"},
		{"id": "s2", "role_slug": "pm", "bound_member_id": "other", "claim": map[string]any{"member_id": "other"}},
	}
	_, slug, self, create, err = pickJoinSeat(seats2, "me", TeamJoinOpts{}, false, "")
	if err != nil || !self || create || slug != "wiki" {
		t.Fatalf("self no-role: self=%v create=%v slug=%s err=%v", self, create, slug, err)
	}
}

func TestTeamInviteRedeemAndJoinHappy(t *testing.T) {
	var sawForce bool
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/team-invites/redeem":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "ymt_abc") {
				t.Errorf("redeem body=%s", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workspace_id":      "ws-new",
				"team_id":           "team-1",
				"allow_expand_seat": false,
				"role_hint":         "frontend",
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/roster/ensure"):
			if !strings.Contains(r.URL.Path, "/workspaces/ws-new/") {
				t.Errorf("workspace not switched: %s", r.URL.Path)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "seat-fe", "role_slug": "frontend", "display_name": "前端", "claim": nil},
					{"id": "seat-pm", "role_slug": "pm", "bound_member_id": "other", "claim": map[string]any{"member_id": "other", "claimed_at": "t"}},
				},
			})
		case strings.Contains(r.URL.Path, "/bots/me"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"member_id": "member-me", "display_name": "BotMe", "actor_kind": "bot", "workspace_id": "ws-new",
			})
		case strings.HasSuffix(r.URL.Path, "/handoff") && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"handoff": map[string]any{
					"claimed_role": map[string]any{"name": "前端", "slug": "frontend"},
					"memories":     []any{},
					"todos":        []any{},
					"issues":       []any{},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/bind") && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), `"force":true`) {
				sawForce = true
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"seat": map[string]any{
					"id": "seat-fe", "role_slug": "frontend",
					"bound_member_id": "member-me", "bound_at": "2026-10-24T02:00:00Z",
					"bound_display_name": "BotMe",
				},
			})
		case strings.HasSuffix(r.URL.Path, "/roster") && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "seat-fe", "role_slug": "frontend", "claim": nil},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/members"):
			_ = json.NewEncoder(w).Encode(map[string]any{"members": []any{}})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	out, err := c.TeamJoin(TeamJoinOpts{Code: "ymt_abc", Role: "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Cfg.WorkspaceID != "ws-new" {
		t.Fatalf("workspace=%s", c.Cfg.WorkspaceID)
	}
	if out["role"] != "frontend" || out["skipped_enter"] == true {
		t.Fatalf("%#v", out)
	}
	if sawForce {
		t.Fatal("join must not send force on empty/active path")
	}
	txt := FormatTeamJoinText(out)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID in join text: %s", txt)
	}
	if !strings.Contains(txt, "join role=frontend") {
		t.Fatalf("text=%s", txt)
	}
}

func TestTeamJoinRefuseForceOnActive(t *testing.T) {
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/team-invites/redeem":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workspace_id": "ws-new", "team_id": "team-1", "allow_expand_seat": false,
			})
		case strings.HasSuffix(r.URL.Path, "/roster/ensure"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "s1", "role_slug": "frontend", "bound_member_id": "other",
						"bound_display_name": "Busy", "claim": map[string]any{"member_id": "other", "claimed_at": "t"}},
				},
			})
		case strings.Contains(r.URL.Path, "/bots/me"):
			_ = json.NewEncoder(w).Encode(map[string]any{"member_id": "member-me", "display_name": "Me", "actor_kind": "bot"})
		case strings.HasSuffix(r.URL.Path, "/bind"):
			t.Error("must not call bind on active seat")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"code":"conflict","message":"should not reach"}`))
		default:
			http.NotFound(w, r)
		}
	})
	_, err := c.TeamJoin(TeamJoinOpts{Code: "ymt_x", Role: "frontend", Force: true})
	if err == nil {
		t.Fatal("expected conflict")
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 409 || !strings.Contains(ae.Message, "禁止顶替") {
		t.Fatalf("got %v", err)
	}
}

func TestTeamJoinAllowExpandCreatesSeat(t *testing.T) {
	var createdSlug string
	var forceOnBind bool
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/team-invites/redeem":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workspace_id": "ws-new", "team_id": "team-1", "allow_expand_seat": true, "role_hint": "frontend-2",
			})
		case strings.HasSuffix(r.URL.Path, "/roster/ensure"):
			roster := []map[string]any{
				{"id": "s1", "role_slug": "frontend", "bound_member_id": "other",
					"claim": map[string]any{"member_id": "other", "claimed_at": "t"}},
			}
			if createdSlug != "" {
				roster = append(roster, map[string]any{"id": "s-new", "role_slug": createdSlug, "claim": nil})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"roster": roster})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/seats") && !strings.Contains(r.URL.Path, "/seats/"):
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			createdSlug = strField(m, "role_slug", "slug")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"seat": map[string]any{"id": "s-new", "role_slug": createdSlug},
			})
		case strings.Contains(r.URL.Path, "/bots/me"):
			_ = json.NewEncoder(w).Encode(map[string]any{"member_id": "member-me", "display_name": "Me", "actor_kind": "bot"})
		case strings.HasSuffix(r.URL.Path, "/handoff"):
			_ = json.NewEncoder(w).Encode(map[string]any{"handoff": map[string]any{"memories": []any{}, "todos": []any{}, "issues": []any{}}})
		case strings.HasSuffix(r.URL.Path, "/roster") && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "s1", "role_slug": "frontend", "bound_member_id": "other", "claim": map[string]any{"member_id": "other"}},
					{"id": "s-new", "role_slug": createdSlug, "claim": nil},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/bind"):
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), `"force":true`) {
				forceOnBind = true
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"seat": map[string]any{"id": "s-new", "role_slug": createdSlug, "bound_member_id": "member-me", "bound_at": "t"},
			})
		case strings.HasSuffix(r.URL.Path, "/members"):
			_ = json.NewEncoder(w).Encode(map[string]any{"members": []any{}})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	})
	out, err := c.TeamJoin(TeamJoinOpts{Code: "ymt_exp"})
	if err != nil {
		t.Fatal(err)
	}
	if createdSlug != "frontend-2" {
		t.Fatalf("created=%s", createdSlug)
	}
	if out["seat_created"] != true || out["role"] != "frontend-2" {
		t.Fatalf("%#v", out)
	}
	if forceOnBind {
		t.Fatal("create+enter must not force")
	}
}

func TestTeamInviteRedeem404Human(t *testing.T) {
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"not found"}`))
	})
	_, err := c.TeamInviteRedeem("ymt_missing", "")
	if err == nil {
		t.Fatal("expected 404")
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 404 {
		t.Fatalf("%v", err)
	}
	if !strings.Contains(ae.Error(), "等服务端 G3") {
		t.Fatalf("msg=%s", ae.Error())
	}
}

func TestTeamSeatCreate404Human(t *testing.T) {
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found"}`))
	})
	_, err := c.TeamSeatCreate("team-1", TeamSeatCreateOpts{RoleSlug: "x"})
	if err == nil {
		t.Fatal("expected 404")
	}
	if !strings.Contains(err.Error(), "等服务端 G2") {
		t.Fatalf("msg=%s", err.Error())
	}
}

func TestTeamJoinSkippedEnterIdempotent(t *testing.T) {
	var bindCalls int
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/team-invites/redeem":
			_ = json.NewEncoder(w).Encode(map[string]any{"workspace_id": "ws-new", "team_id": "team-1"})
		case strings.HasSuffix(r.URL.Path, "/roster/ensure"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "s1", "role_slug": "wiki", "bound_member_id": "member-me"},
				},
			})
		case strings.Contains(r.URL.Path, "/bots/me"):
			_ = json.NewEncoder(w).Encode(map[string]any{"member_id": "member-me", "display_name": "Me", "actor_kind": "bot"})
		case strings.HasSuffix(r.URL.Path, "/handoff"):
			_ = json.NewEncoder(w).Encode(map[string]any{"handoff": map[string]any{"memories": []any{map[string]any{"k": "v"}}, "todos": []any{}, "issues": []any{}}})
		case strings.HasSuffix(r.URL.Path, "/bind"):
			bindCalls++
			w.WriteHeader(http.StatusConflict)
		default:
			http.NotFound(w, r)
		}
	})
	out, err := c.TeamJoin(TeamJoinOpts{Code: "ymt_again", Role: "wiki"})
	if err != nil {
		t.Fatal(err)
	}
	if bindCalls != 0 {
		t.Fatalf("bind called %d", bindCalls)
	}
	if out["skipped_enter"] != true {
		t.Fatalf("%#v", out)
	}
}

func TestNextAutoSeatSlug(t *testing.T) {
	if g := nextAutoSeatSlug(nil); g != "seat-1" {
		t.Fatalf("%s", g)
	}
	seats := []map[string]any{{"role_slug": "seat-1"}, {"role_slug": "seat-3"}}
	if g := nextAutoSeatSlug(seats); g != "seat-4" {
		t.Fatalf("%s", g)
	}
}

func TestFormatTeamJoinTextIdentityReused(t *testing.T) {
	txt := FormatTeamJoinText(map[string]any{
		"role": "frontend", "step": "redeem→team pull→seat enter→handoff get",
		"identity_reused": true,
	})
	if !strings.Contains(txt, "有 K 用") || !strings.Contains(txt, "未重发") {
		t.Fatalf("want 有 K 用 messaging, got %s", txt)
	}
	if strings.Contains(txt, "token=") {
		t.Fatalf("must not print token when reused: %s", txt)
	}
}

func TestTeamJoinIdentityReusedWhenKeyPresent(t *testing.T) {
	c := newJoinTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/team-invites/redeem":
			if ah := r.Header.Get("Authorization"); !strings.HasPrefix(ah, "Bearer ymk_") {
				t.Errorf("expected Bearer ymk_ on redeem, got %q", ah)
			}
			// 有 K 用：服务端不重发
			_ = json.NewEncoder(w).Encode(map[string]any{
				"workspace_id": "ws-new", "team_id": "team-1", "allow_expand_seat": false,
				"member_id": "member-me", "token_issued": false,
			})
		case strings.HasSuffix(r.URL.Path, "/roster/ensure"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "seat-fe", "role_slug": "frontend", "bound_member_id": "member-me"},
				},
			})
		case strings.Contains(r.URL.Path, "/handoff"):
			_ = json.NewEncoder(w).Encode(map[string]any{"handoff": map[string]any{"memories": []any{}, "todos": []any{}, "issues": []any{}}})
		default:
			http.NotFound(w, r)
		}
	})
	out, err := c.TeamJoin(TeamJoinOpts{Code: "ymt_reuse", Role: "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	if out["identity_reused"] != true {
		t.Fatalf("want identity_reused, got %#v", out)
	}
	if out["token_issued"] == true || strField(out, "token") != "" {
		t.Fatalf("must not remint: %#v", out)
	}
	txt := FormatTeamJoinText(out)
	if !strings.Contains(txt, "有 K 用") {
		t.Fatalf("text=%s", txt)
	}
}

