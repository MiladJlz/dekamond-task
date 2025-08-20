package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisOTP struct {
	client         *redis.Client
	otpTTL         time.Duration
	rateLimit      int
	rateLimitWindow time.Duration
}

func NewRedisClient(addr string, otpTTL time.Duration, rateLimit int, rateLimitWindow time.Duration) *RedisOTP {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	})
	return &RedisOTP{
		client:         rdb,
		otpTTL:         otpTTL,
		rateLimit:      rateLimit,
		rateLimitWindow: rateLimitWindow,
	}
}

// generateRandomOTP generates a random 6-digit OTP
func generateRandomOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// Generate creates and stores an OTP in Redis
func (r *RedisOTP) Generate(phone string) (string, error) {
	code := generateRandomOTP()
	key := fmt.Sprintf("otp:%s", phone)

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.client.Set(timeoutCtx, key, code, r.otpTTL).Err(); err != nil {
		return "", err
	}

	return code, nil
}

// Validate verifies an OTP from Redis and deletes it on success
func (r *RedisOTP) Validate(phone, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s", phone)

	// Use a timeout context for Redis operations
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	val, err := r.client.Get(timeoutCtx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	if val == code {
		_ = r.client.Del(timeoutCtx, key).Err()
		return true, nil
	}
	return false, nil
}

// RateLimit increments and checks rate limiting counters in Redis
func (r *RedisOTP) RateLimit(phone string) (bool, error) {
	key := fmt.Sprintf("rate:%s", phone)

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cnt, err := r.client.Incr(timeoutCtx, key).Result()
	if err != nil {
		return false, err
	}

	if cnt == 1 {
		_ = r.client.Expire(timeoutCtx, key, r.rateLimitWindow).Err()
	}
	return int(cnt) <= r.rateLimit, nil 
}

// PingRedis tests Redis connectivity
func (r *RedisOTP) PingRedis() error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return r.client.Ping(timeoutCtx).Err()
}
