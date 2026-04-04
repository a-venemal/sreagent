package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/repository"
)

// ---- typed config structs (replaces config.AIConfig / config.LarkConfig) ----

// AIConfig holds AI/LLM integration configuration stored in the DB.
type AIConfig struct {
	Provider string `json:"provider"` // openai, azure, ollama, custom
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	Enabled  bool   `json:"enabled"`
}

// LarkConfig holds Lark/Feishu bot configuration stored in the DB.
type LarkConfig struct {
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	DefaultWebhook    string `json:"default_webhook"`
	VerificationToken string `json:"verification_token"`
	EncryptKey        string `json:"encrypt_key"`
	BotEnabled        bool   `json:"bot_enabled"`
}

const (
	groupAI   = "ai"
	groupLark = "lark"

	// cacheTTL is how long a cached config entry is considered fresh.
	cacheTTL = 30 * time.Second

	// encPrefix is the prefix prepended to AES-GCM ciphertext stored in the DB.
	// Format: "enc:" + base64(12-byte nonce + ciphertext)
	encPrefix = "enc:"
)

// sensitiveKeys lists the setting keys that must be encrypted at rest.
// Key format: "group.key_name".
var sensitiveKeys = map[string]bool{
	"ai.api_key":              true,
	"lark.app_secret":         true,
	"lark.verification_token": true,
	"lark.encrypt_key":        true,
}

// cachedConfig is a generic TTL-cached value.
type cachedConfig[T any] struct {
	value     T
	expiresAt time.Time
}

func (c *cachedConfig[T]) valid() bool {
	return !c.expiresAt.IsZero() && time.Now().Before(c.expiresAt)
}

// SystemSettingService manages platform-level key-value settings stored in DB.
// AI and Lark configs are cached in memory for cacheTTL (30 s) to avoid a DB
// round-trip on every LLM/Lark call. Writes invalidate the cache immediately.
//
// Sensitive fields (api_key, app_secret, etc.) are encrypted with AES-256-GCM
// using the master key loaded from the SREAGENT_SECRET_KEY environment variable
// (32-byte hex string). If the env var is absent, values are stored plaintext
// and a warning is logged at startup.
type SystemSettingService struct {
	repo      *repository.SystemSettingRepository
	logger    *zap.Logger
	masterKey []byte // 32-byte AES key; nil if not configured

	aiMu    sync.RWMutex
	aiCache cachedConfig[AIConfig]

	larkMu    sync.RWMutex
	larkCache cachedConfig[LarkConfig]
}

// NewSystemSettingService creates a new SystemSettingService.
// It attempts to load the master encryption key from SREAGENT_SECRET_KEY.
func NewSystemSettingService(repo *repository.SystemSettingRepository, logger *zap.Logger) *SystemSettingService {
	svc := &SystemSettingService{repo: repo, logger: logger}

	keyHex := os.Getenv("SREAGENT_SECRET_KEY")
	if keyHex == "" {
		logger.Warn("SREAGENT_SECRET_KEY not set — sensitive settings will be stored in plaintext")
	} else {
		key, err := hex.DecodeString(keyHex)
		if err != nil || len(key) != 32 {
			logger.Error("SREAGENT_SECRET_KEY must be a 64-character hex string (32 bytes); falling back to plaintext storage",
				zap.Error(err),
			)
		} else {
			svc.masterKey = key
			logger.Info("encryption key loaded for sensitive settings")
		}
	}

	return svc
}

// encryptValue encrypts a plaintext string using AES-256-GCM.
// Returns "enc:<base64(nonce+ciphertext)>" or the original value if no key is set.
func (s *SystemSettingService) encryptValue(plaintext string) (string, error) {
	if len(s.masterKey) == 0 || plaintext == "" {
		return plaintext, nil
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptValue decrypts a value encrypted by encryptValue.
// Values not starting with encPrefix are returned as-is (backward compatible).
func (s *SystemSettingService) decryptValue(value string) (string, error) {
	if len(s.masterKey) == 0 || !strings.HasPrefix(value, encPrefix) {
		return value, nil
	}

	data, err := base64.StdEncoding.DecodeString(value[len(encPrefix):])
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", io.ErrUnexpectedEOF
	}

	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// setEncrypted encrypts a value for a given group+key if it is sensitive.
func (s *SystemSettingService) setEncrypted(group, key, value string) (string, error) {
	if sensitiveKeys[group+"."+key] {
		return s.encryptValue(value)
	}
	return value, nil
}

// getDecrypted decrypts a value for a given group+key if it is sensitive.
func (s *SystemSettingService) getDecrypted(group, key, value string) string {
	if !sensitiveKeys[group+"."+key] {
		return value
	}
	plain, err := s.decryptValue(value)
	if err != nil {
		s.logger.Error("failed to decrypt sensitive setting",
			zap.String("group", group),
			zap.String("key", key),
			zap.Error(err),
		)
		return ""
	}
	return plain
}

// ---- AI config ---------------------------------------------------------------

// GetAIConfig loads the AI configuration from cache or DB.
// Cache TTL is cacheTTL (30 s); writes invalidate the cache immediately.
func (s *SystemSettingService) GetAIConfig(ctx context.Context) (AIConfig, error) {
	// Fast path: read from cache.
	s.aiMu.RLock()
	if s.aiCache.valid() {
		cfg := s.aiCache.value
		s.aiMu.RUnlock()
		return cfg, nil
	}
	s.aiMu.RUnlock()

	// Slow path: load from DB and repopulate cache.
	kv, err := s.repo.ListByGroup(ctx, groupAI)
	if err != nil {
		return AIConfig{}, err
	}
	cfg := AIConfig{
		Provider: strDef(kv["provider"], "openai"),
		APIKey:   s.getDecrypted(groupAI, "api_key", kv["api_key"]),
		BaseURL:  strDef(kv["base_url"], "https://api.openai.com/v1"),
		Model:    strDef(kv["model"], "gpt-4o"),
		Enabled:  parseBool(kv["enabled"]),
	}

	s.aiMu.Lock()
	s.aiCache = cachedConfig[AIConfig]{value: cfg, expiresAt: time.Now().Add(cacheTTL)}
	s.aiMu.Unlock()

	return cfg, nil
}

// SaveAIConfig persists all AI configuration keys to the DB and invalidates cache.
// Empty api_key means "do not overwrite the existing key".
func (s *SystemSettingService) SaveAIConfig(ctx context.Context, cfg AIConfig) error {
	kv := map[string]string{
		"provider": cfg.Provider,
		"base_url": cfg.BaseURL,
		"model":    cfg.Model,
		"enabled":  strconv.FormatBool(cfg.Enabled),
	}
	// Only save api_key when caller provided a non-empty value (avoids clearing
	// a stored key when the frontend sends back the masked placeholder).
	if cfg.APIKey != "" {
		enc, err := s.setEncrypted(groupAI, "api_key", cfg.APIKey)
		if err != nil {
			s.logger.Error("failed to encrypt ai.api_key", zap.Error(err))
			return err
		}
		kv["api_key"] = enc
	}
	if err := s.repo.SetGroup(ctx, groupAI, kv); err != nil {
		return err
	}
	// Invalidate cache so the next read fetches fresh data.
	s.aiMu.Lock()
	s.aiCache = cachedConfig[AIConfig]{}
	s.aiMu.Unlock()
	return nil
}

// ---- Lark config -------------------------------------------------------------

// GetLarkConfig loads the Lark bot configuration from cache or DB.
func (s *SystemSettingService) GetLarkConfig(ctx context.Context) (LarkConfig, error) {
	// Fast path: read from cache.
	s.larkMu.RLock()
	if s.larkCache.valid() {
		cfg := s.larkCache.value
		s.larkMu.RUnlock()
		return cfg, nil
	}
	s.larkMu.RUnlock()

	// Slow path: load from DB and repopulate cache.
	kv, err := s.repo.ListByGroup(ctx, groupLark)
	if err != nil {
		return LarkConfig{}, err
	}
	cfg := LarkConfig{
		AppID:             kv["app_id"],
		AppSecret:         s.getDecrypted(groupLark, "app_secret", kv["app_secret"]),
		DefaultWebhook:    kv["default_webhook"],
		VerificationToken: s.getDecrypted(groupLark, "verification_token", kv["verification_token"]),
		EncryptKey:        s.getDecrypted(groupLark, "encrypt_key", kv["encrypt_key"]),
		BotEnabled:        parseBool(kv["bot_enabled"]),
	}

	s.larkMu.Lock()
	s.larkCache = cachedConfig[LarkConfig]{value: cfg, expiresAt: time.Now().Add(cacheTTL)}
	s.larkMu.Unlock()

	return cfg, nil
}

// SaveLarkConfig persists all Lark bot configuration keys to the DB and invalidates cache.
// Empty secret fields are not overwritten (same pattern as AI).
func (s *SystemSettingService) SaveLarkConfig(ctx context.Context, cfg LarkConfig) error {
	kv := map[string]string{
		"app_id":          cfg.AppID,
		"default_webhook": cfg.DefaultWebhook,
		"bot_enabled":     strconv.FormatBool(cfg.BotEnabled),
	}

	encryptField := func(group, key, value string) (string, error) {
		if value == "" {
			return "", nil
		}
		enc, err := s.setEncrypted(group, key, value)
		if err != nil {
			s.logger.Error("failed to encrypt lark field",
				zap.String("key", key),
				zap.Error(err),
			)
			return "", err
		}
		return enc, nil
	}

	if cfg.AppSecret != "" {
		enc, err := encryptField(groupLark, "app_secret", cfg.AppSecret)
		if err != nil {
			return err
		}
		kv["app_secret"] = enc
	}
	if cfg.EncryptKey != "" {
		enc, err := encryptField(groupLark, "encrypt_key", cfg.EncryptKey)
		if err != nil {
			return err
		}
		kv["encrypt_key"] = enc
	}
	if cfg.VerificationToken != "" {
		enc, err := encryptField(groupLark, "verification_token", cfg.VerificationToken)
		if err != nil {
			return err
		}
		kv["verification_token"] = enc
	}

	if err := s.repo.SetGroup(ctx, groupLark, kv); err != nil {
		return err
	}
	// Invalidate cache so the next read fetches fresh data.
	s.larkMu.Lock()
	s.larkCache = cachedConfig[LarkConfig]{}
	s.larkMu.Unlock()
	return nil
}

// ---- helpers -----------------------------------------------------------------

func strDef(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func parseBool(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}
