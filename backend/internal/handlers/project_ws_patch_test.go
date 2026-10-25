package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// S-G2：生产 chi 须挂 PATCH /workspaces/{ws}/projects/{id}（FE patchProject 不改 path）。
func TestPatchProjectUnderWorkspaceChi(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	tok := loginUser(t, h, "admin@yumeng.tech", "Admin123!")

	rr := doJSON(t, h, http.MethodGet, "/api/v1/workspaces", tok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("workspaces %d %s", rr.Code, rr.Body.String())
	}
	var wsRes struct {
		Workspaces []struct {
			ID string `json:"id"`
		} `json:"workspaces"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &wsRes)
	if len(wsRes.Workspaces) == 0 {
		t.Fatal("no workspace")
	}
	wsID := wsRes.Workspaces[0].ID

	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspaces/"+wsID+"/projects", tok, map[string]string{
		"name": "改名前", "key_prefix": "SG2",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	if created.Project.ID == "" {
		t.Fatalf("no project id: %s", rr.Body.String())
	}

	// 核心：ws 作用域 PATCH（曾漏挂 → 404）
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/workspaces/"+wsID+"/projects/"+created.Project.ID, tok, map[string]string{
		"name": "改名后",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("ws patch %d %s", rr.Code, rr.Body.String())
	}
	var patched struct {
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &patched)
	if patched.Project.Name != "改名后" {
		t.Fatalf("name=%q body=%s", patched.Project.Name, rr.Body.String())
	}

	// 顶层兼容仍可用
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/projects/"+created.Project.ID, tok, map[string]string{
		"name": "顶层仍可",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("top patch %d %s", rr.Code, rr.Body.String())
	}

	// 顺手：ws Get
	rr = doJSON(t, h, http.MethodGet, "/api/v1/workspaces/"+wsID+"/projects/"+created.Project.ID, tok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("ws get %d %s", rr.Code, rr.Body.String())
	}

}
