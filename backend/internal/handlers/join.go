package handlers

import (
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/joinlink"
)

// JoinHandler serves GET /join — one-tap landing (API host).
// Does NOT redeem or claim; guides to existing ymt_ + ymbot team join / redeem.
type JoinHandler struct {
	// PublicAPIBase is the configured API root (PUBLIC_API_BASE); default production.
	PublicAPIBase string
	// BoardBase optional FE origin for deep-link hint; empty → text-only guidance.
	BoardBase string
}

// Page handles GET /join?code=&role=&seat=&api=
func (h *JoinHandler) Page(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	code := strings.TrimSpace(q.Get("code"))
	role := strings.TrimSpace(q.Get("role"))
	seat := strings.TrimSpace(q.Get("seat"))
	apiQ := strings.TrimSpace(q.Get("api"))

	if err := joinlink.ValidateInviteCode(code); err != nil {
		writeJoinHTML(w, http.StatusBadRequest, joinErrorPage(err.Error()))
		return
	}

	p := joinlink.Params{Code: code, Role: role, Seat: seat, API: apiQ}
	apiBase := joinlink.EffectiveAPIBase(h.PublicAPIBase, p)
	prompt, err := joinlink.BuildBotPrompt(apiBase, p)
	if err != nil {
		writeJoinHTML(w, http.StatusBadRequest, joinErrorPage(err.Error()))
		return
	}
	cli, err := joinlink.CLIJoinLine(p)
	if err != nil {
		writeJoinHTML(w, http.StatusBadRequest, joinErrorPage(err.Error()))
		return
	}

	boardHint := "打开看板 → 设置 → Bot 席 →「加入团队」，粘贴入场码完成（人类路径；Bot 请用下方提示词或 CLI）。"
	if b := strings.TrimRight(strings.TrimSpace(h.BoardBase), "/"); b != "" {
		boardHint = "看板深链（同参预填，若前端已实现 /join）：" + b + "/join?code=" + urlQueryEscape(code)
		if role != "" {
			boardHint += "&role=" + urlQueryEscape(role)
		}
		boardHint += "。或：设置 → Bot 席 →「加入团队」。"
	}

	writeJoinHTML(w, http.StatusOK, joinOKPage(joinOKData{
		APIBase:     apiBase,
		Code:        code,
		Role:        role,
		Seat:        seat,
		CLI:         cli,
		Prompt:      prompt,
		BoardHint:   boardHint,
		InstallCurl: joinlink.InstallCurlLine(apiBase),
	}))
}

func urlQueryEscape(s string) string {
	return url.QueryEscape(s)
}

type joinOKData struct {
	APIBase     string
	Code        string
	Role        string
	Seat        string
	CLI         string
	Prompt      string
	BoardHint   string
	InstallCurl string
}

func writeJoinHTML(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func joinErrorPage(msg string) string {
	esc := html.EscapeString(msg)
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>入场链接无效</title>
<style>
body{font-family:system-ui,-apple-system,sans-serif;max-width:40rem;margin:2rem auto;padding:0 1rem;line-height:1.55;color:#1a1a1a}
h1{font-size:1.25rem} .err{background:#fff1f0;border:1px solid #ffa39e;padding:1rem;border-radius:8px}
a{color:#1677ff}
</style>
</head>
<body>
<h1>无法打开入场页</h1>
<div class="err">` + esc + `</div>
<p>正确格式示例：<code>https://bot.api.sn.cn/join?code=ymt_…&amp;role=frontend</code></p>
<p>入场仍走既有团队入场码兑换（<code>ymt_</code> + <code>ymbot team join</code>），本页不另造 claim。</p>
</body>
</html>`
}

func joinOKPage(d joinOKData) string {
	roleRow := "（未指定，join 时自动选席）"
	if d.Role != "" {
		roleRow = html.EscapeString(d.Role)
	}
	seatRow := ""
	if d.Seat != "" {
		seatRow = `<tr><th>席位 ID</th><td><code>` + html.EscapeString(d.Seat) + `</code></td></tr>`
	}
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>一键入场 · AI 协作看板</title>
<style>
body{font-family:system-ui,-apple-system,sans-serif;max-width:44rem;margin:2rem auto;padding:0 1rem;line-height:1.55;color:#1a1a1a}
h1{font-size:1.35rem;margin-bottom:.25rem} .sub{color:#666;margin-top:0}
table{border-collapse:collapse;width:100%;margin:1rem 0}
th,td{text-align:left;padding:.4rem .5rem;border-bottom:1px solid #eee;vertical-align:top}
th{width:7rem;color:#555;font-weight:600}
pre,textarea{width:100%;box-sizing:border-box;background:#f6f8fa;border:1px solid #d0d7de;border-radius:8px;padding:.75rem;font-size:.85rem;overflow:auto}
textarea{min-height:14rem;resize:vertical}
.actions{display:flex;flex-wrap:wrap;gap:.5rem;margin:1rem 0}
button{cursor:pointer;background:#1677ff;color:#fff;border:0;border-radius:6px;padding:.45rem .9rem;font-size:.9rem}
button.secondary{background:#fff;color:#1677ff;border:1px solid #1677ff}
.note{background:#f0f5ff;border:1px solid #adc6ff;padding:.75rem 1rem;border-radius:8px;font-size:.9rem}
.ok{color:#389e0d;font-size:.85rem;margin-left:.5rem}
</style>
</head>
<body>
<h1>一键入场</h1>
<p class="sub">先装 <code>ymbot</code>（若无）→ 再贴提示词或 CLI 入队。本页<strong>不会</strong>消耗入场码（兑换仍由 <code>team join</code> / redeem 完成）。</p>
<table>
<tr><th>API 根</th><td><code>` + html.EscapeString(d.APIBase) + `</code></td></tr>
<tr><th>入场码</th><td><code>` + html.EscapeString(d.Code) + `</code></td></tr>
<tr><th>角色</th><td>` + roleRow + `</td></tr>
` + seatRow + `
</table>
<div class="note">` + html.EscapeString(d.BoardHint) + `</div>
<h2>① 一键安装 ymbot（若本机还没有）</h2>
<p class="sub">禁止假设已装 CLI。主路径 curl|bash；备路径见提示词内 <code>go install …</code>。</p>
<pre id="install">` + html.EscapeString(d.InstallCurl) + `</pre>
<div class="actions">
<button type="button" onclick="copyText('install',this)">复制安装命令</button>
</div>
<h2>② CLI 入队（装好后再跑）</h2>
<pre id="cli">` + html.EscapeString(d.CLI) + `</pre>
<div class="actions">
<button type="button" onclick="copyText('cli',this)">复制 CLI</button>
</div>
<h2>Bot 提示词</h2>
<textarea id="prompt" readonly>` + html.EscapeString(d.Prompt) + `</textarea>
<div class="actions">
<button type="button" onclick="copyText('prompt',this)">复制提示词</button>
<button type="button" class="secondary" onclick="copyText('cli',this)">复制 CLI</button>
</div>
<p class="sub">安全：链接含入场码明文，勿发公开群；勿把 <code>ymk_</code> 写入可转发内容。禁区：不要教改 mcp.json / 长串 YMBOT_*。</p>
<script>
function copyText(id, btn){
  var el = document.getElementById(id);
  var t = el.value !== undefined ? el.value : el.textContent;
  navigator.clipboard.writeText(t).then(function(){
    var s = document.createElement('span'); s.className='ok'; s.textContent='已复制';
    btn.after(s); setTimeout(function(){ s.remove(); }, 1600);
  });
}
</script>
</body>
</html>`
}
