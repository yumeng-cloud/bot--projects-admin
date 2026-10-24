# 看板 UI 挂载说明

负责人：看板 UI。包管理：**pnpm**。

## 导出

| 模块 | 路径 |
|------|------|
| 看板页 | `src/pages/board` → `BoardPage` |
| 列表页 | `src/pages/list` → `IssueListPage` |
| Cycle 页 | `src/pages/cycles` → `CyclesPage` |
| Issue 抽屉 | `src/components/issue` → `IssueDrawer` |
| 看板组件 | `src/components/board` → `KanbanBoard` |

## Props

三页均需：`{ workspaceId: string; projectId: string }`

## 路由示例（React 前端挂载）

前缀：`/w/:wsId/...`

已由 `App.tsx` + `ProjectScoped.tsx` 挂载：

- `/w/:wsId/projects/:projectId/board|issues|cycles` → 真页面
- `/w/:wsId/board|list|cycles` → `ProjectRedirect` 到首个项目（无项目则演示 `MOCK_PROJECT`）


## Mock

- 默认打真接口（`/workspaces/{ws}/issues|cycles`）；**仅失败**回退 mock，顶栏「演示数据」。
- 强制 mock：`localStorage ai-collab-board:force-mock=1` 或 `VITE_USE_MOCK=1`。
- 失败回退不粘滞：下次请求仍先打真接口。

## API 预留（workspace 作用域，与 api/issues|cycles 一致）

- `GET /workspaces/{ws}/issues?project_id=`
- `PATCH /workspaces/{ws}/issues/{id}`（含 `version`）
- `GET /workspaces/{ws}/cycles?project_id=`
- `PATCH /workspaces/{ws}/cycles/{id}`

未改：`layouts/`、`LoginPage`、`HomePage`、`App.tsx`、`api/client.ts` 核心。

## Issue 详情布局

对标 Linear 的抽屉信息架构与验收标准见 **[ISSUE_DETAIL_LAYOUT.md](./ISSUE_DETAIL_LAYOUT.md)**（产品体验助手）。定稿要点：桌面抽屉默认 **800×100%**（宽≥720）；描述编辑器选定 **ByteMD ^1.22**（见该文档 §11）。实施前与 @UI 视觉 / @React 前端同步。
