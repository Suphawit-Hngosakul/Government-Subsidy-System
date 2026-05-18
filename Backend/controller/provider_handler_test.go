package controller

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"government-subsidy-system/backend/repository"
	"government-subsidy-system/backend/service"

	_ "github.com/lib/pq"
)

func providerHandlerTestMux(t *testing.T) *http.ServeMux {
	t.Helper()

	dsn := os.Getenv("PROVIDER_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open provider database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("provider database is not reachable: %v", err)
	}

	mux := http.NewServeMux()
	NewProviderHandler(service.NewProviderService(repository.NewPostgresProviderRepository(db))).RegisterRoutes(mux)
	return mux
}

func TestProviderHandlerDOPAVerify(t *testing.T) {
	mux := providerHandlerTestMux(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dopa/verify/1101700203451", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if body := rec.Body.String(); !containsAll(body, `"nationalId":"1101700203451"`, `"valid":true`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestProviderHandlerRejectsInvalidNationalID(t *testing.T) {
	mux := providerHandlerTestMux(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sso/status/123", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func containsAll(text string, needles ...string) bool {
	for _, needle := range needles {
		if !contains(text, needle) {
			return false
		}
	}
	return true
}

func contains(text, needle string) bool {
	return strings.Contains(text, needle)
}
