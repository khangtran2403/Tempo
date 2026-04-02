package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiter sử dụng Redis để implement rate limiting
type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter() (*RateLimiter, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis-14702.c252.ap-southeast-1-1.ec2.cloud.redislabs.com:14702",
		Username: "ktran",
		Password: "0899154297kK@",
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RateLimiter{
		client: client,
	}, nil
}

// Allow kiểm tra xem request có được phép không
func (rl *RateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := now.Truncate(window)

	redisKey := fmt.Sprintf("ratelimit:%s:%d", key, windowStart.Unix())

	pipe := rl.client.Pipeline()
	incrCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	count := int(incrCmd.Val())
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := windowStart.Add(window)
	allowed := count <= limit

	return allowed, remaining, resetTime, nil
}

// AllowWithBurst cho phép burst requests
func (rl *RateLimiter) AllowWithBurst(
	ctx context.Context,
	key string,
	limit int,
	burstLimit int,
	window time.Duration,
) (bool, int, time.Time, error) {
	// Use token bucket algorithm
	now := time.Now()
	redisKey := fmt.Sprintf("ratelimit:burst:%s", key)

	luaScript := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local rate = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])
		
		local bucket = redis.call('HMGET', key, 'tokens', 'last_update')
		local tokens = tonumber(bucket[1]) or burst
		local last_update = tonumber(bucket[2]) or now
		
		-- Calculate tokens to add
		local elapsed = now - last_update
		local new_tokens = tokens + (elapsed * rate)
		if new_tokens > burst then
			new_tokens = burst
		end
		
		-- Try to consume 1 token
		if new_tokens >= 1 then
			new_tokens = new_tokens - 1
			redis.call('HMSET', key, 'tokens', new_tokens, 'last_update', now)
			redis.call('EXPIRE', key, 3600)
			return {1, math.floor(new_tokens)}
		else
			redis.call('HMSET', key, 'tokens', new_tokens, 'last_update', now)
			return {0, 0}
		end
	`

	rate := float64(limit) / window.Seconds()

	result, err := rl.client.Eval(
		ctx,
		luaScript,
		[]string{redisKey},
		limit, burstLimit, rate, now.Unix(),
	).Result()

	if err != nil {
		return false, 0, time.Time{}, err
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remaining := int(resultSlice[1].(int64))
	resetTime := now.Add(window)

	return allowed, remaining, resetTime, nil
}

// Close đóng Redis connection
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}
