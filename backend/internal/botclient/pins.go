package botclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// PinList GET …/teams/{teamID}/pins
func (c *Client) PinList(teamID string) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return nil, fmt.Errorf("需要 team uuid")
	}
	return c.Do(http.MethodGet, c.ws("/teams/"+teamID+"/pins"), nil, nil, false)
}

// PinGet GET …/teams/{teamID}/pins/{key}
func (c *Client) PinGet(teamID, key string) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	key = strings.TrimSpace(key)
	if teamID == "" || key == "" {
		return nil, fmt.Errorf("需要 team uuid 与 key")
	}
	return c.Do(http.MethodGet, c.ws("/teams/"+teamID+"/pins/"+url.PathEscape(key)), nil, nil, false)
}

// PinSet PUT …/teams/{teamID}/pins body {key,value}（契约主路径；handler 亦接受 …/pins/{key}）。
func (c *Client) PinSet(teamID, key, value string) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	key = strings.TrimSpace(key)
	if teamID == "" || key == "" {
		return nil, fmt.Errorf("需要 team uuid 与 key")
	}
	body := map[string]string{"key": key, "value": value}
	return c.Do(http.MethodPut, c.ws("/teams/"+teamID+"/pins"), nil, body, true)
}

// PinDelete DELETE …/teams/{teamID}/pins/{key}
func (c *Client) PinDelete(teamID, key string) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	key = strings.TrimSpace(key)
	if teamID == "" || key == "" {
		return nil, fmt.Errorf("需要 team uuid 与 key")
	}
	return c.Do(http.MethodDelete, c.ws("/teams/"+teamID+"/pins/"+url.PathEscape(key)), nil, nil, false)
}

// ProjectPins 瘦投影 pins[] / pin → [{key,value,updated_at}]；--id 附加 id。
func ProjectPins(raw json.RawMessage, opts ViewOpts) ([]map[string]any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	var list []map[string]any
	if arr, ok := top["pins"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				list = append(list, m)
			}
		}
	} else if one, ok := top["pin"].(map[string]any); ok {
		list = []map[string]any{one}
	} else if okVal, has := top["ok"]; has {
		return []map[string]any{{"ok": okVal}}, nil
	} else {
		return nil, fmt.Errorf("unexpected pin JSON shape")
	}
	out := make([]map[string]any, 0, len(list))
	for _, m := range list {
		if opts.Full {
			cp := map[string]any{}
			for k, v := range m {
				cp[k] = v
			}
			out = append(out, cp)
			continue
		}
		item := map[string]any{
			"key":        strField(m, "key"),
			"value":      trunc(strField(m, "value"), 80),
			"updated_at": shortTime(strField(m, "updated_at")),
		}
		if opts.IDs {
			if id := strField(m, "id"); id != "" {
				item["id"] = id
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// FormatPinListText 默认瘦文本：key\tupdated\tvalue（无裸 UUID）。
func FormatPinListText(items []map[string]any) string {
	if len(items) == 0 {
		return "(无 pin)\n"
	}
	if len(items) == 1 {
		if v, ok := items[0]["ok"]; ok {
			return fmt.Sprintf("ok=%v\n", v)
		}
	}
	var b strings.Builder
	for _, it := range items {
		fmt.Fprintf(&b, "%s\t%s\t%s",
			dash(strField(it, "key")),
			dash(strField(it, "updated_at")),
			dash(strField(it, "value")),
		)
		if id := strField(it, "id"); id != "" {
			fmt.Fprintf(&b, "\tid=%s", id)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
