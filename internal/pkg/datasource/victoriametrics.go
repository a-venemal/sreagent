package datasource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// VictoriaMetricsChecker checks VictoriaMetrics health.
type VictoriaMetricsChecker struct{}

func (c *VictoriaMetricsChecker) CheckHealth(ctx context.Context, endpoint, authType, authConfig string) error {
	url := strings.TrimRight(endpoint, "/") + "/health"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	applyAuth(req, authType, authConfig)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
