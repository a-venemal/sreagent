package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/engine"
)

const stateKeyPrefix = "sreagent:state:"

// stateKey returns the Redis Hash key for a rule's alert states.
// Format: sreagent:state:{ruleID}
func stateKey(ruleID uint) string {
	return fmt.Sprintf("%s%d", stateKeyPrefix, ruleID)
}

// RedisStateStore implements engine.StateStore using Redis Hashes.
// Each rule's states are stored in a single Hash: sreagent:state:{ruleID}
// with fingerprint as field and JSON-serialized StateEntry as value.
// The Hash key has a TTL of 3x the evaluation interval so stale data
// is automatically cleaned up if the evaluator stops.
type RedisStateStore struct {
	client *Client
	logger *zap.Logger
}

// NewRedisStateStore creates a RedisStateStore backed by the given Client.
func NewRedisStateStore(client *Client, logger *zap.Logger) *RedisStateStore {
	return &RedisStateStore{
		client: client,
		logger: logger.Named("redis_state_store"),
	}
}

// SaveState persists a single alert state entry for a rule.
// It sets the field in the rule's Hash and refreshes the Hash TTL.
func (s *RedisStateStore) SaveState(ctx context.Context, ruleID uint, fp string, entry *engine.StateEntry, ttl time.Duration) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal state entry: %w", err)
	}

	key := stateKey(ruleID)
	pipe := s.client.rdb.Pipeline()
	pipe.HSet(ctx, key, fp, data)
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("save state to redis: %w", err)
	}
	return nil
}

// DeleteState removes a single alert state entry from the rule's Hash.
func (s *RedisStateStore) DeleteState(ctx context.Context, ruleID uint, fp string) error {
	key := stateKey(ruleID)
	err := s.client.rdb.HDel(ctx, key, fp).Err()
	if err != nil {
		return fmt.Errorf("delete state from redis: %w", err)
	}
	return nil
}

// LoadStates loads all persisted alert states for a given rule.
// Returns an empty map (not error) if the key does not exist.
func (s *RedisStateStore) LoadStates(ctx context.Context, ruleID uint) (map[string]*engine.StateEntry, error) {
	key := stateKey(ruleID)
	result, err := s.client.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return make(map[string]*engine.StateEntry), nil
		}
		return nil, fmt.Errorf("load states from redis: %w", err)
	}

	states := make(map[string]*engine.StateEntry, len(result))
	for fp, raw := range result {
		var entry engine.StateEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			s.logger.Warn("failed to unmarshal state entry, skipping",
				zap.Uint("rule_id", ruleID),
				zap.String("fingerprint", fp),
				zap.Error(err),
			)
			continue
		}
		states[fp] = &entry
	}
	return states, nil
}

// DeleteRuleStates removes all persisted states for a rule (when rule is stopped).
func (s *RedisStateStore) DeleteRuleStates(ctx context.Context, ruleID uint) error {
	key := stateKey(ruleID)
	err := s.client.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("delete rule states from redis: %w", err)
	}
	return nil
}
