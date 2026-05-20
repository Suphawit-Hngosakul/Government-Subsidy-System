package adapter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"government-subsidy-system/backend/domain"
)

func TestHTTPOrchestratorAdapterReturnsDecodedDecision(t *testing.T) {
	expected := domain.EligibilityResult{
		ClaimID: "claim-001",
		Status:  "approved",
		Reasons: []string{"eligible under project criteria rules"},
		Sources: domain.EligibilitySources{
			DOPA: domain.EligibilityDOPA{Valid: true, Age: 35, Alive: true, CardActive: true},
			SSO:  domain.EligibilitySSO{Section: "40", ContributionMonths: 12},
			KTB:  domain.EligibilityKTB{DepositTotal: 12000, AverageMonthlyIncome: 15000, PromptPayLinked: true},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/decision" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		var body map[string]string
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["claimId"] != "claim-001" {
			t.Errorf("expected claimId=claim-001 in body, got %q", body["claimId"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	adapter := NewHTTPOrchestratorAdapter(srv.URL)
	result, err := adapter.GetDecision(context.Background(), "claim-001")
	if err != nil {
		t.Fatalf("GetDecision returned error: %v", err)
	}
	if result.Status != "approved" {
		t.Fatalf("expected status approved, got %s", result.Status)
	}
	if result.Sources.DOPA.Age != 35 {
		t.Fatalf("expected DOPA age 35, got %d", result.Sources.DOPA.Age)
	}
}

func TestHTTPOrchestratorAdapterReturnsEmptyResultOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	adapter := NewHTTPOrchestratorAdapter(srv.URL)
	result, err := adapter.GetDecision(context.Background(), "claim-missing")
	if err != nil {
		t.Fatalf("expected nil error on 404, got %v", err)
	}
	if result.ClaimID != "" || result.Status != "" {
		t.Fatalf("expected empty result, got %+v", result)
	}
}

func TestHTTPOrchestratorAdapterErrorsOn5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()

	adapter := NewHTTPOrchestratorAdapter(srv.URL)
	_, err := adapter.GetDecision(context.Background(), "claim-001")
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}
