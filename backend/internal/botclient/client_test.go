package botclient

import (
	"strings"
	"testing"
)

func TestResolveIDKeyPriority(t *testing.T) {
	t.Setenv("YMBOT_IDEMPOTENCY_KEY", "from-env")
	if got := ResolveIDKey("explicit-key"); got != "explicit-key" {
		t.Fatalf("explicit: got %q", got)
	}
	if got := ResolveIDKey(""); got != "from-env" {
		t.Fatalf("env: got %q", got)
	}
	t.Setenv("YMBOT_IDEMPOTENCY_KEY", "")
	got := ResolveIDKey("")
	if got == "" || got == "from-env" || len(got) != 32 {
		t.Fatalf("auto uuid hex: got %q", got)
	}
}

func TestAPIErrorChinese(t *testing.T) {
	cases := []struct {
		name string
		err  *APIError
		want []string
	}{
		{"401 revoked", &APIError{Status: 401, Code: "unauthorized", Message: "API Key 已吊销"}, []string{"认证失败", "已吊销", "HTTP 401"}},
		{"401 invalid", &APIError{Status: 401, Code: "unauthorized", Message: "无效的 API Key"}, []string{"认证失败", "无效", "HTTP 401"}},
		{"401 empty", &APIError{Status: 401, Code: "unauthorized", Message: ""}, []string{"认证失败", "已吊销", "无效", "HTTP 401"}},
		{"403 done", &APIError{Status: 403, Code: "forbidden", Message: "Bot 不能将 Issue 设为 done，需人类确认（可改为 in_review）"}, []string{"无权限", "done", "HTTP 403"}},
		{"404", &APIError{Status: 404, Code: "not_found", Message: "Issue 不存在"}, []string{"未找到", "Issue", "HTTP 404"}},
		{"409", &APIError{Status: 409, Code: "version_conflict", Message: "乐观锁冲突"}, []string{"冲突", "乐观锁", "HTTP 409"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.err.Error()
			if strings.Contains(s, "\n") {
				t.Fatalf("must be one line: %q", s)
			}
			for _, w := range tc.want {
				if !strings.Contains(s, w) {
					t.Fatalf("Error()=%q missing %q", s, w)
				}
			}
		})
	}
}

func TestAPIErrorExitCodes(t *testing.T) {
	if (&APIError{Status: 401}).ExitCode() != 10 {
		t.Fatal("401")
	}
	if (&APIError{Status: 403}).ExitCode() != 11 {
		t.Fatal("403")
	}
	if (&APIError{Status: 404}).ExitCode() != 12 {
		t.Fatal("404")
	}
	if (&APIError{Status: 409}).ExitCode() != 13 {
		t.Fatal("409")
	}
}

func TestGuardBotDoneChinese(t *testing.T) {
	err := GuardBotDone("done")
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatal("want APIError")
	}
	s := ae.Error()
	if !strings.Contains(s, "done") || !strings.Contains(s, "人类确认") {
		t.Fatalf("guard: %q", s)
	}
}

func TestGuardBotOrphanCreate(t *testing.T) {
	if err := GuardBotOrphanCreate("parent-1"); err != nil {
		t.Fatalf("with parent: %v", err)
	}
	err := GuardBotOrphanCreate("")
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 403 {
		t.Fatalf("want 403 APIError, got %v", err)
	}
	s := ae.Error()
	if !strings.Contains(s, "游离") || !strings.Contains(s, "父") {
		t.Fatalf("orphan guard: %q", s)
	}
}
