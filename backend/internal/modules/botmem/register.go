package botmem

import "github.com/yumeng-tech/ai-collab-board/backend/internal/modules"

func init() {
	// 路由已由 server.NewRouter → handlers.MountBotsChi 挂在 workspace 子路由上。
	// 保留非 nil 赋值，避免巡检误判「未注册」；此处不再二次 Mount，防止与 chi 冲突。
	modules.RegisterBotMem = func(mux modules.Mux, deps modules.Deps) {
		_ = mux
		_ = deps
	}
}
