package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func providerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("PROVIDER_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open provider database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skipf("provider database is not reachable: %v", err)
	}
	return db
}

func TestPostgresProviderRepositoryReadsSeededEligibleCitizen(t *testing.T) {
	db := providerTestDB(t)
	defer db.Close()

	repo := NewPostgresProviderRepository(db)
	ctx := context.Background()

	dopa, err := repo.GetDOPAVerification(ctx, "1101700203451")
	if err != nil {
		t.Fatalf("GetDOPAVerification returned error: %v", err)
	}
	if !dopa.Valid || dopa.Age <= 18 || !dopa.Alive || !dopa.CardActive {
		t.Fatalf("expected seeded citizen to be valid, got %+v", dopa)
	}

	sso, err := repo.GetSSOStatus(ctx, "1101700203451")
	if err != nil {
		t.Fatalf("GetSSOStatus returned error: %v", err)
	}
	if sso.Section == nil || *sso.Section != "40" {
		t.Fatalf("expected SSO section 40, got %+v", sso)
	}

	ktb, err := repo.GetKTBFinancialCheck(ctx, "1101700203451")
	if err != nil {
		t.Fatalf("GetKTBFinancialCheck returned error: %v", err)
	}
	if ktb.DepositTotal != 12000 || ktb.AverageMonthlyIncome != 15000 {
		t.Fatalf("unexpected KTB financial data: %+v", ktb)
	}
}
