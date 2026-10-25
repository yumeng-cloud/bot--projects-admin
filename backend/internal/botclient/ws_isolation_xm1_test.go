package botclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newWSIsoTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		Cfg: Config{
			BaseURL:     srv.URL,
			APIKey:      "ymk_test",
			WorkspaceID: "ws-2",
			MemberID:    "mid-ws1", // 错仓 mid 场景
		},
		HTTP: srv.Client(),
	}
}

func TestAnnotateBotWSErrHuman(t *testing.T) {
	e403 := annotateBotWSErr(&APIError{Status: 403, Code: "forbidden", Message: ""})
	s := e403.Error()
	if !strings.Contains(s, "工作区隔离") || !strings.Contains(s, "HTTP 403") {
		t.Fatalf("403: %q", s)
	}
	e404 := annotateBotWSErr(&APIError{Status: 404, Code: "not_found", Message: ""})
	s = e404.Error()
	if !strings.Contains(s, "错仓") || !strings.Contains(s, "404") || !strings.Contains(s, "禁止当空列表") {
		t.Fatalf("404: %q", s)
	}
	// 幂等：已含隔离字样不再叠字
	again := annotateBotWSErr(e404)
	if again.Error() != e404.Error() {
		t.Fatalf("re-annotate changed: %q vs %q", again, e404)
	}
}

func TestMemoryListPropagates403NotEmptySuccess(t *testing.T) {
	c := newWSIsoTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/memories") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "forbidden", "message": "不是工作区成员"})
	})
	raw, err := c.MemoryList("", "", 0)
	if raw != nil {
		t.Fatalf("want nil raw on 403, got %s", raw)
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 403 {
		t.Fatalf("want APIError 403, got %v", err)
	}
	s := ae.Error()
	if strings.Contains(s, "(无记忆)") {
		t.Fatalf("must not look like empty success: %q", s)
	}
	if !strings.Contains(s, "工作区") && !strings.Contains(s, "隔离") {
		t.Fatalf("want WS isolation human text: %q", s)
	}
}

func TestMemoryListPropagates404NotEmptySuccess(t *testing.T) {
	c := newWSIsoTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "not_found", "message": "成员不存在"})
	})
	raw, err := c.MemoryList("", "", 0)
	if raw != nil {
		t.Fatalf("want nil raw on 404, got %s", raw)
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 404 {
		t.Fatalf("want APIError 404, got %v", err)
	}
	s := ae.Error()
	if strings.Contains(s, "(无记忆)") {
		t.Fatalf("must not print 无记忆 on error: %q", s)
	}
	if !strings.Contains(s, "错仓") && !strings.Contains(s, "禁止当空列表") {
		t.Fatalf("want 错仓 hint: %q", s)
	}
}

func TestMemorySetPropagates404(t *testing.T) {
	c := newWSIsoTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "not_found", "message": "成员不存在"})
	})
	raw, err := c.MemorySet("k", "v", "")
	if raw != nil || err == nil {
		t.Fatalf("want 404 err, raw=%s err=%v", raw, err)
	}
	ae := err.(*APIError)
	if ae.Status != 404 || !strings.Contains(ae.Error(), "404") {
		t.Fatalf("%v", err)
	}
}

func TestHandoffGetPropagates404(t *testing.T) {
	c := newWSIsoTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "not_found", "message": "成员不存在"})
	})
	_, err := c.HandoffGet()
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 404 {
		t.Fatalf("got %v", err)
	}
}

func TestMemoryList200EmptyStillOKUntilGoGate(t *testing.T) {
	// Go X-G1 落地前服务端仍可能 200+[]；通道不伪造 404，只保证真 403/404 不软吞。
	c := newWSIsoTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"memories": []any{}})
	})
	raw, err := c.MemoryList("", "", 0)
	if err != nil {
		t.Fatalf("200 empty must still succeed until Go gate: %v", err)
	}
	items, err := ProjectMemories(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	txt := FormatMemoryListText(items)
	if txt != "(无记忆)\n" {
		t.Fatalf("got %q", txt)
	}
}

func TestFormatMemoryListTextOnlyForSuccessPath(t *testing.T) {
	// 契约：CLI/MCP 仅在 err==nil 后调用 FormatMemoryListText；错误走人话 APIError。
	err := annotateBotWSErr(&APIError{Status: 404, Message: "成员不存在"})
	if strings.Contains(err.Error(), "(无记忆)") {
		t.Fatal("error text must not contain 无记忆")
	}
}
