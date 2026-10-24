# 主题（UI 视觉）

- **默认浅色（强制产品默认）**：layout `#fafafa`、容器白、细边框、主色 `#5E6AD2`
- **storage**：`localStorage` key `ai-collab-board:theme-mode-v2`（`light`|`dark`）；无值或旧 key → **light**
- **文件**：`src/theme/linearTheme.ts`；挂载 `src/main.tsx`；`index.html` 预置 `data-theme="light"`
- **暗色**：仅用户点顶栏切换并写入 v2 key 后生效；首次进入必为浅色
- **空态**：`.acb-empty-panel`；**活动**：普通人话时间线
- **勿擅自改 token**：壳层切换 UI 归 React 前端
