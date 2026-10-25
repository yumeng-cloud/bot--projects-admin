package botclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newSeatLeaveTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		Cfg: Config{
			BaseURL:     srv.URL,
			APIKey:      "ymk_test",
			WorkspaceID: "ws-1",
			MemberID:    "member-me",
		},
		HTTP: srv.Client(),
	}
}

func TestFormatSeatLeaveTextTransmitsStaleNote(t *testing.T) {
	txt := FormatSeatLeaveText(map[string]any{
		"role": "frontend",
		"via":  "stale_cleared",
		"note": MsgSeatBindStaleCleared,
	})
	if !strings.Contains(txt, MsgSeatBindStaleCleared) {
		t.Fatalf("missing note: %s", txt)
	}
	if !strings.Contains(txt, "left role=frontend via=stale_cleared") {
		t.Fatalf("text=%s", txt)
	}
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID leaked: %s", txt)
	}
}

func TestSeatLeaveUnbindSuccessClearsCacheAndPassesMessage(t *testing.T) {
	c := newSeatLeaveTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/roster"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "seat-1", "role_slug": "frontend", "bound_member_id": "member-me"},
				},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unbind"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"seat":    map[string]any{"id": "seat-1", "role_slug": "frontend", "status": "left"},
				"message": MsgSeatBindStaleCleared,
			})
		default:
			http.NotFound(w, r)
		}
	})
	c.ensureCaches()
	c.cacheMu.Lock()
	c.memberNames["member-me"] = "Me"
	c.membersLoaded = true
	c.cacheMu.Unlock()

	out, err := c.SeatLeave("frontend", SeatLeaveOpts{TeamID: "team-1"})
	if err != nil {
		t.Fatal(err)
	}
	if out["via"] != "unbind" {
		t.Fatalf("via=%v", out["via"])
	}
	if out["note"] != MsgSeatBindStaleCleared {
		t.Fatalf("note=%v", out["note"])
	}
	c.cacheMu.Lock()
	loaded := c.membersLoaded
	c.cacheMu.Unlock()
	if loaded {
		t.Fatal("expected ClearLocalTeamSeatCache to reset membersLoaded")
	}
	txt := FormatSeatLeaveText(out)
	if !strings.Contains(txt, MsgSeatBindStaleCleared) {
		t.Fatalf("text=%s", txt)
	}
}

func TestSeatLeaveForbiddenDirtyBindTreatedAsStaleCleared(t *testing.T) {
	c := newSeatLeaveTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/roster"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{
						"id":                 "seat-1",
						"role_slug":          "frontend",
						"bound_member_id":    "member-gone",
						"bound_display_name": "OldBot",
					},
				},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unbind"):
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": "forbidden", "message": "只能解绑自己绑定的席位",
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/members"):
			// 脏绑：成员表已无 member-gone
			_ = json.NewEncoder(w).Encode(map[string]any{
				"members": []map[string]any{
					{"id": "member-me", "display_name": "Me"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	})

	out, err := c.SeatLeave("frontend", SeatLeaveOpts{TeamID: "team-1"})
	if err != nil {
		t.Fatalf("want success got %v", err)
	}
	if out["via"] != "stale_cleared" {
		t.Fatalf("via=%v", out["via"])
	}
	if out["note"] != MsgSeatBindStaleCleared {
		t.Fatalf("note=%v", out["note"])
	}
	txt := FormatSeatLeaveText(out)
	if strings.Contains(txt, "只能解绑自己") {
		t.Fatalf("dead-end 403 leaked: %s", txt)
	}
	if !strings.Contains(txt, MsgSeatBindStaleCleared) {
		t.Fatalf("text=%s", txt)
	}
}

func TestSeatLeaveForbiddenLiveOtherKeepsU6(t *testing.T) {
	c := newSeatLeaveTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/roster"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{
						"id":                 "seat-1",
						"role_slug":          "frontend",
						"bound_member_id":    "member-other",
						"bound_display_name": "OtherBot",
					},
				},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unbind"):
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": "forbidden", "message": "只能解绑自己绑定的席位",
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/members"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"members": []map[string]any{
					{"id": "member-me", "display_name": "Me"},
					{"id": "member-other", "display_name": "OtherBot"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	})

	_, err := c.SeatLeave("frontend", SeatLeaveOpts{TeamID: "team-1"})
	if err == nil {
		t.Fatal("expected 403 for live other")
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 403 {
		t.Fatalf("want 403 got %v", err)
	}
	if strings.Contains(ae.Message, "只能解绑自己") {
		t.Fatalf("old dead-end msg: %s", ae.Message)
	}
	if !strings.Contains(ae.Message, "OtherBot") || !strings.Contains(ae.Message, "占用") {
		t.Fatalf("want U6 with holder: %s", ae.Message)
	}
}

func TestSeatLeaveAlreadyUnboundViaServer200(t *testing.T) {
	c := newSeatLeaveTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/roster"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"roster": []map[string]any{
					{"id": "seat-1", "role_slug": "frontend", "status": "left"},
				},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/unbind"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"seat": map[string]any{"id": "seat-1", "role_slug": "frontend", "status": "left"},
			})
		default:
			http.NotFound(w, r)
		}
	})
	out, err := c.SeatLeave("frontend", SeatLeaveOpts{TeamID: "team-1"})
	if err != nil {
		t.Fatal(err)
	}
	if out["via"] != "unbind" {
		t.Fatalf("via=%v", out["via"])
	}
}

func TestIsUnbindSelfOnlyForbidden(t *testing.T) {
	if !isUnbindSelfOnlyForbidden(&APIError{Status: 403, Message: "只能解绑自己绑定的席位"}) {
		t.Fatal("expected match")
	}
	if isUnbindSelfOnlyForbidden(&APIError{Status: 403, Message: "席位由「X」占用；请对方离席"}) {
		t.Fatal("U6 must not match self-only")
	}
	if isUnbindSelfOnlyForbidden(&APIError{Status: 409, Message: "只能解绑自己绑定的席位"}) {
		t.Fatal("wrong status")
	}
}
