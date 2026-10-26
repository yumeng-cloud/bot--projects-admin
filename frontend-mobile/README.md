# frontend-mobile · 育蒙协作移动端壳（MV2）

正常手机 App 新壳（Vite + React + TS）。与 PC `frontend/` 同账号、同 API；**浅色安静**画布 `#F2F4F8` / 主色 `#5E6AD2`。

本波 **MV2-R0**：脚手架 + 登录 + `GET /auth/me` 绿。
**MV2-R1**：进壳 hydrate `GET /me/session-context`（有值 → `/resume` 续现场；空 → 安静默认）；可 `PUT` 更新；失败**不踢下线**。R2（底栏 Tab · 列表详情）另刀。

## 启动

需本机 API（默认 `http://127.0.0.1:8080`）。端口 **5174**，不与 PC Vite `5173` 冲突。

在仓库根（推荐，workspace）：

```bash
pnpm install
pnpm --filter ai-collab-board-frontend-mobile dev
```

或进入本目录：

```bash
cd frontend-mobile
pnpm install   # 根已装过可跳过
pnpm dev       # http://127.0.0.1:5174
```

构建：

```bash
cd frontend-mobile && pnpm build
```

可选环境变量：`VITE_API_BASE` / `VITE_API_BASE_URL`（默认同源 `/api/v1`，开发由 Vite 代理到 `:8080`）。

## 手验（R0）

1. 起后端 `:8080` + 本壳 `pnpm dev`。
2. 打开 http://127.0.0.1:5174 → 登录页。
3. 用已有账号（或 seed `admin@yumeng.tech` / `Admin123!`）登录。
4. 进「我的」页应见姓名 / 邮箱；徽章 **/auth/me 绿**。可点「再打 /auth/me」复核。
5. **R1**：若此前 PC/移动写过现场，登录后应进 `/resume` 看到 wsId 等；空现场则留在「我的」安静默认。可「PUT 写一条现场」后再登录验证续现场；断 API 时不应被踢回登录。

## 共享客户端

依赖 workspace 包 `@acb/api-client`（`packages/api-client`）：只鉴权头 / 错误 / 类型。禁拷贝 PC layouts / antd 密表 / 看板 DnD。

## 非目标（本目录 V1 首刀）

Roku/货架/D-pad；Expo/RN；Capacitor；动 Soft / LANDING / INVITE / PC 主路径。

## 依赖说明

根目录 `pnpm-workspace.yaml` 含 `packages/*` 与 `frontend-mobile`。正常：`pnpm install`（根）后 `pnpm --filter ai-collab-board-frontend-mobile dev|build`。

若 registry 暂不可达，可临时复用 PC 前端已装依赖（版本对齐）：对 `react` / `react-dom` / `react-router-dom` / `vite` / `typescript` / `@types/*` / `@vitejs/plugin-react` 做 symlink 到 `frontend/node_modules`，`@acb/api-client` → `packages/api-client`，再 `pnpm build`。
