package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/joinlink"
)

func TestJoinPage_OK(t *testing.T) {
	h := &JoinHandler{PublicAPIBase: joinlink.DefaultAPIBase}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/join?code=ymt_deadbeef&role=frontend", nil)
	h.Page(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type %s", ct)
	}
	body := rr.Body.String()
	for _, need := range []string{
		"ymt_deadbeef",
		"frontend",
		"ymbot team join --code ymt_deadbeef --role frontend",
		"https://bot.api.sn.cn",
		"复制提示词",
		"whoami",
	} {
		if !strings.Contains(body, need) {
			t.Fatalf("missing %q in body", need)
		}
	}
	// must not redeem / invent claim language that implies second claim
	if strings.Contains(body, "ymk_") && strings.Contains(body, "ymk_abcdef") {
		t.Fatal("must not embed a concrete ymk_ secret")
	}
}

func TestJoinPage_MissingCode(t *testing.T) {
	h := &JoinHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/join", nil)
	h.Page(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "缺少入场码") {
		t.Fatalf("body %s", rr.Body.String())
	}
}

func TestJoinPage_BadPrefix(t *testing.T) {
	h := &JoinHandler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/join?code=ymk_not_invite", nil)
	h.Page(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "ymt_") {
		t.Fatalf("body %s", body)
	}
}

func TestJoinPage_APIOverride(t *testing.T) {
	h := &JoinHandler{PublicAPIBase: joinlink.DefaultAPIBase}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/join?code=ymt_zz&api=http://127.0.0.1:8080", nil)
	h.Page(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "http://127.0.0.1:8080") {
		t.Fatalf("expected override api in page: %s", rr.Body.String())
	}
}
