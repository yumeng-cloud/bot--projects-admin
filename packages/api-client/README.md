# @acb/api-client

共享 HTTP 客户端：**只**鉴权头 / 错误 / 类型 / session-context。

- 供 `frontend-mobile`（及后续端）复用
- **不含** PC layouts、antd、看板 DnD
- 契约：`/auth/login`、`/auth/me`、Bearer token、`GET|PUT /me/session-context`

```ts
import { login, me, getSessionContext, putSessionContext } from '@acb/api-client'
```
