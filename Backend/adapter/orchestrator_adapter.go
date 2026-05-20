package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"government-subsidy-system/backend/domain"
)

type HTTPOrchestratorAdapter struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPOrchestratorAdapter(baseURL string) *HTTPOrchestratorAdapter {
	return &HTTPOrchestratorAdapter{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (a *HTTPOrchestratorAdapter) Orchestrate(ctx context.Context, claimID string, nationalID string, projectID string) error {
	payload := map[string]string{
		"claimId":    claimID,
		"nationalId": nationalID,
		"projectId":  projectID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode orchestrate request: %w", err)
	}

	endpoint, err := url.JoinPath(a.baseURL, "/internal/orchestrate")
	if err != nil {
		return fmt.Errorf("build orchestrate url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build orchestrate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call orchestrator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("orchestrator returned %d: %s", resp.StatusCode, raw)
	}

	return nil
}

func (a *HTTPOrchestratorAdapter) GetDecision(ctx context.Context, claimID string) (domain.EligibilityResult, error) {
	body, err := json.Marshal(map[string]string{"claimId": claimID})
	if err != nil {
		return domain.EligibilityResult{}, fmt.Errorf("encode decision request: %w", err)
	}

	endpoint, err := url.JoinPath(a.baseURL, "/internal/decision")
	if err != nil {
		return domain.EligibilityResult{}, fmt.Errorf("build decision url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return domain.EligibilityResult{}, fmt.Errorf("build decision request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return domain.EligibilityResult{}, fmt.Errorf("call orchestrator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.EligibilityResult{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return domain.EligibilityResult{}, fmt.Errorf("orchestrator returned %d: %s", resp.StatusCode, raw)
	}

	var result domain.EligibilityResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.EligibilityResult{}, fmt.Errorf("decode orchestrator response: %w", err)
	}
	return result, nil
}
