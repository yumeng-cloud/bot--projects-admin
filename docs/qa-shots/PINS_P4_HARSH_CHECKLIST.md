# P4 pins 苛刻复测清单（QA）

> 对 Go `6915db4` + FE `4459d01` · 权威 `docs/NEXT_WAVE_DEBT_PINS.md`  
> 等 Bot 记忆箱验证据后立刻独立复验；禁盲测关单。

## 环境
- API `:8080` 含 6915db4；FE `:5174` 含 4459d01
- WS `ba1481c49148aba9553ba3811406b69f`；双队新鲜 CreateTeam
- Admin JWT + 两队成员/Bot key（同 T6-R KEYA/KEYB 可复用若仍为本队成员）

## API（P1–P3 复验 + P4 证据）
1. 队 A `PUT …/teams/{A}/pins` key=`qa.pin.cross` value=`from-A`
2. 队 A `GET …/pins` 见该 key；队 B `GET` **无**该 key → P1
3. 队 B 对 A：`PUT/DELETE` 同 path 用 B 身份或错 team 路径 → **404** → P2
4. 同 Bot 写 `bot_memories` 同名 key 与 pin 互不覆盖 → P3
5. 写证据 `docs/qa-shots/pins-team-isolate-YYYYMMDD.txt`（或记忆已交后 QA 独立文件 `pins-qa-retest-*.txt`）

## FE（P5 / 换队看墙）
1. 设置 → Bot 席位页「记忆墙」：当前队列表=本队 pins
2. 换/切 team：列表清空再载入新队；**无跨队残留** key
3. upsert / 删除一行可见；API 404/501 友好降级不崩
4. 截图 `docs/qa-shots/pins-fe-switch-*.png`（切队前后各一张更佳）

## 苛刻红线
- 错队可见 / 可改 / 可删 → FAIL
- 切队后仍显示旧队 pin → FAIL
- pin 与 bot_memories 串写 → FAIL
- 不做 Issue 池 / T7

## 关报
短报产品 + 总指挥：Overall PASS/FAIL · P1–P5 · 证据路径 · HEAD
