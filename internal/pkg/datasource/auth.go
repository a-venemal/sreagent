package datasource

import (
	"encoding/json"
	"net/http"
)

type basicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type bearerAuthConfig struct {
	Token string `json:"token"`
}

type apiKeyAuthConfig struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

// applyAuth adds authentication headers to the request based on auth type.
func applyAuth(req *http.Request, authType, authConfig string) {
	if authType == "" || authType == "none" || authConfig == "" {
		return
	}

	switch authType {
	case "basic":
		var cfg basicAuthConfig
		if err := json.Unmarshal([]byte(authConfig), &cfg); err == nil {
			req.SetBasicAuth(cfg.Username, cfg.Password)
		}
	case "bearer":
		var cfg bearerAuthConfig
		if err := json.Unmarshal([]byte(authConfig), &cfg); err == nil {
			req.Header.Set("Authorization", "Bearer "+cfg.Token)
		}
	case "api_key":
		var cfg apiKeyAuthConfig
		if err := json.Unmarshal([]byte(authConfig), &cfg); err == nil {
			headerName := cfg.Header
			if headerName == "" {
				headerName = "X-API-Key"
			}
			req.Header.Set(headerName, cfg.Value)
		}
	}
}
