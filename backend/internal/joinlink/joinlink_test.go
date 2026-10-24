package joinlink

import (
	"strings"
	"testing"
)

func TestBuildJoinURL_Golden(t *testing.T) {
	got, err := BuildJoinURL(DefaultAPIBase, Params{
		Code: "ymt_deadbeefcafebabe",
		Role: "frontend",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "https://bot.api.sn.cn/join?code=ymt_deadbeefcafebabe&role=frontend"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildJoinURL_SeatAndOmitProductionAPI(t *testing.T) {
	got, err := BuildJoinURL("https://bot.api.sn.cn/", Params{
		Code: "ymt_abc123",
		Role: "wiki",
		Seat: "seat-uuid-1",
		API:  DefaultAPIBase, // same as production → omit api=
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "api=") {
		t.Fatalf("production api= should be omitted: %s", got)
	}
	if !strings.Contains(got, "seat=seat-uuid-1") || !strings.Contains(got, "role=wiki") {
		t.Fatalf("missing seat/role: %s", got)
	}
}

func TestBuildJoinURL_CustomAPIQuery(t *testing.T) {
	got, err := BuildJoinURL("https://bot.api.sn.cn", Params{
		Code: "ymt_abc123",
		API:  "http://127.0.0.1:8080",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "api=http%3A%2F%2F127.0.0.1%3A8080") && !strings.Contains(got, "api=http://127.0.0.1:8080") {
		// url.Values encodes
		if !strings.Contains(got, "api=") {
			t.Fatalf("expected api= override: %s", got)
		}
	}
	uHost := "https://bot.api.sn.cn/join?"
	if !strings.HasPrefix(got, uHost) {
		t.Fatalf("host should stay production page: %s", got)
	}
}

func TestBuildJoinURL_SelfHost(t *testing.T) {
	got, err := BuildJoinURL("http://127.0.0.1:8080", Params{Code: "ymt_xxyyzz", Role: "go"})
	if err != nil {
		t.Fatal(err)
	}
	want := "http://127.0.0.1:8080/join?code=ymt_xxyyzz&role=go"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestValidateInviteCode(t *testing.T) {
	cases := []struct {
		code string
		ok   bool
		sub  string
	}{
		{"", false, "缺少入场码"},
		{"ymw_xxx", false, "ymt_"},
		{"ymk_secret", false, "ymt_"},
		{"ymt_", false, "不完整"},
		{"ymt_ab", true, ""},
	}
	for _, tc := range cases {
		err := ValidateInviteCode(tc.code)
		if tc.ok {
			if err != nil {
				t.Fatalf("code=%q: %v", tc.code, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.sub) {
			t.Fatalf("code=%q err=%v want sub %q", tc.code, err, tc.sub)
		}
	}
}

func TestBuildBotPrompt(t *testing.T) {
	p, err := BuildBotPrompt("", Params{Code: "ymt_cafe", Role: "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	installCurl := "curl -fsSL https://bot.api.sn.cn/install.sh | bash"
	if !strings.HasPrefix(strings.TrimSpace(p), "curl -fsSL ") || !strings.Contains(p, installCurl) {
		t.Fatalf("prompt must lead with install curl|bash:\n%s", p)
	}
	for _, need := range []string{
		installCurl,
		"install.sh",
		GoInstallCmd,
		"go install github.com/yumeng-tech/ai-collab-board/backend/cmd/ymbot@latest",
		"https://bot.api.sn.cn",
		"ymt_cafe",
		"frontend",
		"whoami",
		"team join --code ymt_cafe --role frontend",
		"handoff get",
		"禁止",
		"mcp.json",
		"一键安装",
		"身份两态",
		"自动发一把",
		"已有钥",
		"不重发",
	} {
		if !strings.Contains(p, need) {
			t.Fatalf("prompt missing %q:\n%s", need, p)
		}
	}
	for _, ban := range []string{
		"YMBOT_API_KEY=",
		"ymk_abcdef",
		"使用本机已有 Bot Key",
	} {
		if strings.Contains(p, ban) {
			t.Fatalf("prompt must not contain %q", ban)
		}
	}
}

func TestInstallCurlLine(t *testing.T) {
	got := InstallCurlLine("")
	want := "curl -fsSL https://bot.api.sn.cn/install.sh | bash"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = InstallCurlLine("http://127.0.0.1:8080/")
	if got != "curl -fsSL http://127.0.0.1:8080/install.sh | bash" {
		t.Fatal(got)
	}
}

func TestCLIJoinLine(t *testing.T) {
	got, err := CLIJoinLine(Params{Code: "ymt_1", Role: "qa"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "ymbot team join --code ymt_1 --role qa" {
		t.Fatal(got)
	}
}
