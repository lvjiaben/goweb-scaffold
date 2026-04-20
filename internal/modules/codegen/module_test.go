package codegen

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-core/validate"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type responseEnvelope struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

func newModuleTestRuntime(t *testing.T) *bootstrap.Runtime {
	t.Helper()
	engine := httpx.NewEngine(slog.New(slog.NewTextHandler(io.Discard, nil)))
	repoRoot := t.TempDir()
	for _, dir := range []string{
		"internal/modules",
		"internal/gen",
		"vben-admin/apps/admin/src/generated",
	} {
		if err := os.MkdirAll(filepath.Join(repoRoot, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	runtime := &bootstrap.Runtime{
		RepoRoot:  repoRoot,
		Engine:    engine,
		Validator: validate.New(),
	}
	runtime.AdminProtectedGroup = engine.Group("/admin-api")
	if err := (Module{}).Register(runtime); err != nil {
		t.Fatalf("register codegen module: %v", err)
	}
	return runtime
}

func performJSONRequest(t *testing.T, runtime *bootstrap.Runtime, method string, path string, body string) responseEnvelope {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	runtime.Engine.ServeHTTP(recorder, req)

	var resp responseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, recorder.Body.String())
	}
	return resp
}

func TestPreviewHandlerRejectsMissingTableName(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodPost, "/admin-api/codegen/preview", `{"module_name":"demo_article"}`)
	if resp.Code == 0 {
		t.Fatalf("expected preview validation error, got %+v", resp)
	}
}

func TestDiffHandlerRejectsMissingTableName(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodPost, "/admin-api/codegen/diff", `{"module_name":"demo_article"}`)
	if resp.Code == 0 {
		t.Fatalf("expected diff validation error, got %+v", resp)
	}
}

func TestGenerateHandlerRejectsMissingModuleName(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodPost, "/admin-api/codegen/generate", `{"table_name":"demo_article"}`)
	if resp.Code == 0 {
		t.Fatalf("expected generate validation error, got %+v", resp)
	}
}

func TestRegenerateHandlerRejectsEmptyRequest(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodPost, "/admin-api/codegen/regenerate", `{}`)
	if resp.Code == 0 {
		t.Fatalf("expected regenerate validation error, got %+v", resp)
	}
}

func TestModulesHandlerReturnsEmptyListWithoutLocks(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodGet, "/admin-api/codegen/modules", ``)
	if resp.Code != 0 {
		t.Fatalf("expected modules success, got %+v", resp)
	}
	if _, ok := resp.Data["list"]; !ok {
		t.Fatalf("expected modules response to contain list, got %+v", resp)
	}
}

func TestExportHandlerRejectsMissingModuleNameAndHistoryID(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodGet, "/admin-api/codegen/export", ``)
	if resp.Code == 0 {
		t.Fatalf("expected export validation error, got %+v", resp)
	}
}

func TestRemoveHandlerWarnsWhenModuleDoesNotExist(t *testing.T) {
	runtime := newModuleTestRuntime(t)
	resp := performJSONRequest(t, runtime, http.MethodPost, "/admin-api/codegen/remove", `{"module_name":"missing_demo","remove_files":true,"unregister_module":true,"remove_lock":true}`)
	if resp.Code != 0 {
		t.Fatalf("expected remove to return warning instead of error, got %+v", resp)
	}
}
