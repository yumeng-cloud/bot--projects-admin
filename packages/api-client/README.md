# @acb/api-client

共享 HTTP 客户端：**鉴权头 / 错误 / 类型 / session-context / 看板读改最小面**。

- 供 `frontend-mobile` 复用
- **不含** PC layouts、antd、看板 DnD
- 契约：`/auth/*`、`/me/session-context`、workspaces / teams·roster / issues·comments / wiki spaces

```ts
import { login, me, listIssues, getIssue, patchIssue, putSessionContext } from '@acb/api-client'
```
