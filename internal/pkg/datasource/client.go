package datasource

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthChecker is the interface for checking datasource health.
type HealthChecker interface {
	CheckHealth(ctx context.Context, endpoint, authType, authConfig string) error
}

// httpClient is a shared HTTP client with reasonable timeouts.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// NewChecker creates the appropriate health checker for a datasource type.
func NewChecker(dsType string) (HealthChecker, error) {
	switch dsType {
	case "prometheus":
		return &PrometheusChecker{}, nil
	case "victoriametrics":
		return &VictoriaMetricsChecker{}, nil
	case "zabbix":
		return &ZabbixChecker{}, nil
	case "victorialogs":
		return &VictoriaLogsChecker{}, nil
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", dsType)
	}
}
