// Package queue provides a Redis-backed task queue adapter for background media processing.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	queueAssetProcess = "aura:queue:asset_processing"
	queueAIAnalysis   = "aura:queue:ai_analysis"
	queueTranscode    = "aura:queue:transcode"
)

// JobType classifies a background task.
type JobType string

const (
	JobTypeProcessAsset JobType = "process_asset"
	JobTypeAIAnalyze    JobType = "ai_analyze"
	JobTypeTranscode    JobType = "transcode"
)

// Job is the envelope pushed onto a Redis list.
type Job struct {
	ID        string    `json:"id"`
	Type      JobType   `json:"type"`
	AssetID   uuid.UUID `json:"asset_id"`
	CreatedAt time.Time `json:"created_at"`
	Retries   int       `json:"retries"`
}

// RedisQueue is the Redis-backed implementation of the task queue ports.
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a new RedisQueue.
func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// EnqueueAssetProcessing pushes an asset-processing job onto the queue.
func (q *RedisQueue) EnqueueAssetProcessing(ctx context.Context, assetID uuid.UUID) error {
	return q.push(ctx, queueAssetProcess, Job{
		ID:        uuid.New().String(),
		Type:      JobTypeProcessAsset,
		AssetID:   assetID,
		CreatedAt: time.Now(),
	})
}

// EnqueueAIAnalysis pushes an AI-tagging job.
func (q *RedisQueue) EnqueueAIAnalysis(ctx context.Context, assetID uuid.UUID) error {
	return q.push(ctx, queueAIAnalysis, Job{
		ID:        uuid.New().String(),
		Type:      JobTypeAIAnalyze,
		AssetID:   assetID,
		CreatedAt: time.Now(),
	})
}

// EnqueueTranscode pushes a video transcoding job.
func (q *RedisQueue) EnqueueTranscode(ctx context.Context, assetID uuid.UUID) error {
	return q.push(ctx, queueTranscode, Job{
		ID:        uuid.New().String(),
		Type:      JobTypeTranscode,
		AssetID:   assetID,
		CreatedAt: time.Now(),
	})
}

// Dequeue pops the next job from the given queue (blocking, 5 s timeout).
func (q *RedisQueue) Dequeue(ctx context.Context, queueName string) (*Job, error) {
	result, err := q.client.BLPop(ctx, 5*time.Second, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // timeout — no job available
		}
		return nil, fmt.Errorf("BLPop(%q): %w", queueName, err)
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("unmarshalling job: %w", err)
	}
	return &job, nil
}

func (q *RedisQueue) push(ctx context.Context, queueName string, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshalling job: %w", err)
	}
	return q.client.RPush(ctx, queueName, data).Err()
}
