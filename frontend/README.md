# AI 协作看板 · Frontend

## 启动

```bash
cd frontend
pnpm install
pnpm dev
```

- 地址：http://localhost:5173
- API：`VITE_API_BASE_URL`（默认 `http://localhost:8080/api/v1`）；也可走 Vite proxy `/api` → `:8080`
- 种子：`admin@yumeng.tech` / `Admin123!`

## 技术

- React 19 + TypeScript + Vite 6
- Ant Design **6.6.7**
- react-router-dom 7、dayjs（Asia/Shanghai）

## 页面清单

| 路由 | 说明 |
|------|------|
| `/login` | 登录 / 注册 |
| `/` | 工作区列表 |
| `/w/:wsId/projects/:projectId/board` | 看板（看板 UI） |
| `/w/:wsId/projects/:projectId/issues` | 列表 |
| `/w/:wsId/projects/:projectId/cycles` | Cycles |
| `/w/:wsId/board|list|cycles` | 重定向到首个项目 |
| `/w/:wsId/wiki` | Wiki（空间/树/搜索/读写 Markdown） |
| `/w/:wsId/time` | 工时（列表/记工/汇总/费率/CSV） |
| `/w/:wsId/settings` | 成员 / 标签 / API Key / Webhook |
| `/w/:wsId/settings/bots` | Bot 名册 |

## 缺口

- 看板拖拽 / Issue 抽屉 / 列表筛选：等「看板 UI」
- Issues API list/move：等后端 Issue 域
