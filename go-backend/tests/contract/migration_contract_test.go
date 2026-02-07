package contract_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"go-backend/internal/auth"
	httpserver "go-backend/internal/http"
	"go-backend/internal/http/handler"
	"go-backend/internal/http/response"
	"go-backend/internal/store/sqlite"
)

func TestOpenAPISubStoreContracts(t *testing.T) {
	router, repo := setupContractRouter(t, "contract-jwt-secret")

	const tunnelFlowGB = int64(500)
	const tunnelInFlow = int64(123)
	const tunnelOutFlow = int64(456)
	const tunnelExpTimeMs = int64(2727251700000)

	now := time.Now().UnixMilli()
	res, err := repo.DB().Exec(`INSERT INTO tunnel(name, traffic_ratio, type, protocol, flow, created_time, updated_time, status, in_ip, inx) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"contract-tunnel", 1.0, 1, "tls", 1, now, now, 1, nil, 0)
	if err != nil {
		t.Fatalf("insert tunnel: %v", err)
	}
	tunnelID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	if _, err := repo.DB().Exec(`INSERT INTO user_tunnel(user_id, tunnel_id, speed_id, num, flow, in_flow, out_flow, flow_reset_time, exp_time, status) VALUES(?, ?, NULL, ?, ?, ?, ?, ?, ?, ?)`,
		1, tunnelID, 99999, tunnelFlowGB, tunnelInFlow, tunnelOutFlow, 1, tunnelExpTimeMs, 1); err != nil {
		t.Fatalf("insert user_tunnel: %v", err)
	}

	t.Run("default user subscription payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/open_api/sub_store?user=admin_user&pwd=admin_user", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		expected := "upload=0; download=0; total=107373108658176; expire=2727251700"
		if string(body) != expected {
			t.Fatalf("expected body %q, got %q", expected, string(body))
		}
		if got := resp.Header().Get("subscription-userinfo"); got != expected {
			t.Fatalf("expected subscription-userinfo %q, got %q", expected, got)
		}
		if !strings.Contains(resp.Header().Get("Content-Type"), "text/plain") {
			t.Fatalf("expected text/plain content type, got %q", resp.Header().Get("Content-Type"))
		}
	})

	t.Run("tunnel scoped subscription payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/open_api/sub_store?user=admin_user&pwd=admin_user&tunnel="+strconv.FormatInt(tunnelID, 10), nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		expected := "upload=123; download=456; total=536870912000; expire=2727251700"
		if string(body) != expected {
			t.Fatalf("expected body %q, got %q", expected, string(body))
		}
		if got := resp.Header().Get("subscription-userinfo"); got != expected {
			t.Fatalf("expected subscription-userinfo %q, got %q", expected, got)
		}
	})

	t.Run("invalid credentials returns contract error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/open_api/sub_store?user=admin_user&pwd=wrong", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assertCodeMsg(t, resp, -1, "鉴权失败")
	})

	t.Run("missing tunnel returns contract error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/open_api/sub_store?user=admin_user&pwd=admin_user&tunnel=999999", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assertCodeMsg(t, resp, -1, "隧道不存在")
	})
}

func TestSpeedLimitTunnelsRouteAlias(t *testing.T) {
	secret := "contract-jwt-secret"
	router, _ := setupContractRouter(t, secret)

	t.Run("missing token blocked", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/speed-limit/tunnels", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assertCodeMsg(t, resp, 401, "未登录或token已过期")
	})

	t.Run("admin token receives success envelope", func(t *testing.T) {
		token, err := auth.GenerateToken(1, "admin_user", 0, secret)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/speed-limit/tunnels", nil)
		req.Header.Set("Authorization", token)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		var out response.R
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if out.Code != 0 {
			t.Fatalf("expected code 0, got %d (%s)", out.Code, out.Msg)
		}
	})
}

func setupContractRouter(t *testing.T, jwtSecret string) (http.Handler, *sqlite.Repository) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "contract.db")
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	h := handler.New(repo, jwtSecret)
	return httpserver.NewRouter(h, jwtSecret), repo
}
