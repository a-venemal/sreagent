package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AlertActionClaims holds the JWT claims for no-auth alert action tokens.
type AlertActionClaims struct {
	EventID uint `json:"event_id"`
	jwt.RegisteredClaims
}

// GenerateAlertActionToken creates a JWT token for no-auth alert actions.
// The token embeds the event ID and expires in 24 hours.
func GenerateAlertActionToken(eventID uint, secret string) (string, error) {
	claims := AlertActionClaims{
		EventID: eventID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sreagent-alert-action",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseAlertActionToken validates and extracts the event ID from an alert action token.
func ParseAlertActionToken(tokenStr, secret string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AlertActionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid or expired token: %w", err)
	}

	claims, ok := token.Claims.(*AlertActionClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	return claims.EventID, nil
}
