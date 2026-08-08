package redisqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ayush/supportiq/internal/queue"
	"github.com/ayush/supportiq/internal/utils"
	goredis "github.com/redis/go-redis/v9"
)

const (
	eventChannel = "events:notifications"
)

// Client implements queue.Queue using Redis lists and a sorted set for delayed retry.
type Client struct {
	rdb          *goredis.Client
	mainKey      string // Redis list — main work queue
	retryKey     string // Redis sorted set — delayed retry (score = exec timestamp)
	deadKey      string // Redis list — dead letter queue
}

// New connects to Redis and returns a Client.
func New(redisURL, queueName string) (*Client, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	rdb := goredis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	utils.Logger.WithField("queue", queueName).Info("Queue: Redis connected")

	return &Client{
		rdb:      rdb,
		mainKey:  "queue:" + queueName,
		retryKey: "queue:" + queueName + ":retry",
		deadKey:  "queue:" + queueName + ":dead",
	}, nil
}

// Enqueue pushes a job to the left of the main Redis list (LPUSH).
func (c *Client) Enqueue(ctx context.Context, job queue.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return c.rdb.LPush(ctx, c.mainKey, data).Err()
}

// Dequeue blocks on BRPOP until a job is available or ctx is cancelled.
func (c *Client) Dequeue(ctx context.Context) (*queue.Job, error) {
	result, err := c.rdb.BRPop(ctx, 0, c.mainKey).Result()
	if err != nil {
		return nil, err
	}
	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected BRPop result")
	}

	var job queue.Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}
	return &job, nil
}

// EnqueueDelayed places a job in the retry sorted set with score = execution time.
func (c *Client) EnqueueDelayed(ctx context.Context, job queue.Job, delaySeconds int) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	score := float64(time.Now().Add(time.Duration(delaySeconds) * time.Second).Unix())
	return c.rdb.ZAdd(ctx, c.retryKey, goredis.Z{Score: score, Member: string(data)}).Err()
}

// MoveToDeadLetter archives a permanently failed job.
func (c *Client) MoveToDeadLetter(ctx context.Context, job queue.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	return c.rdb.LPush(ctx, c.deadKey, data).Err()
}

// QueueLen returns the length of the main work queue.
func (c *Client) QueueLen(ctx context.Context) (int64, error) {
	return c.rdb.LLen(ctx, c.mainKey).Result()
}

// RetryQueueLen returns the number of delayed retry jobs.
func (c *Client) RetryQueueLen(ctx context.Context) (int64, error) {
	return c.rdb.ZCard(ctx, c.retryKey).Result()
}

// DeadLetterLen returns the number of permanently failed jobs.
func (c *Client) DeadLetterLen(ctx context.Context) (int64, error) {
	return c.rdb.LLen(ctx, c.deadKey).Result()
}

// MoveDueRetryJobs moves delayed jobs whose time has come back to the main queue.
// Should be called periodically by the worker's retry poller goroutine.
func (c *Client) MoveDueRetryJobs(ctx context.Context) (int, error) {
	now := fmt.Sprintf("%d", time.Now().Unix())
	members, err := c.rdb.ZRangeByScore(ctx, c.retryKey, &goredis.ZRangeBy{
		Min: "-inf",
		Max: now,
	}).Result()
	if err != nil {
		return 0, err
	}

	for _, m := range members {
		pipe := c.rdb.TxPipeline()
		pipe.LPush(ctx, c.mainKey, m)
		pipe.ZRem(ctx, c.retryKey, m)
		if _, err := pipe.Exec(ctx); err != nil {
			utils.Logger.WithError(err).Warn("Queue: Failed to move retry job")
		}
	}
	return len(members), nil
}

// PublishEvent publishes a JSON-encoded event to the Redis pub/sub channel
// so the API server can broadcast it to connected WebSocket clients.
func (c *Client) PublishEvent(ctx context.Context, payload []byte) error {
	return c.rdb.Publish(ctx, eventChannel, payload).Err()
}

// Subscribe returns a pub/sub subscription on the events channel.
// The API server uses this to receive events and forward them to WebSocket clients.
func (c *Client) Subscribe(ctx context.Context) *goredis.PubSub {
	return c.rdb.Subscribe(ctx, eventChannel)
}

// Close releases the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// RawClient returns the underlying go-redis client for direct operations.
func (c *Client) RawClient() *goredis.Client {
	return c.rdb
}
