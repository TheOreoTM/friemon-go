package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	// Trivia state keys
	TriviaStateKey           = "trivia:state"
	TriviaUsersKey           = "trivia:users"
	TriviaScoresKey          = "trivia:scores"
	TriviaQuestionsKey       = "trivia:questions"
	TriviaCurrentQuestionKey = "trivia:current_question"

	// Trivia states
	TriviaStateStarting   = "starting"
	TriviaStateWaiting    = "waiting"
	TriviaStateInProgress = "in_progress"
	TriviaStateEnded      = "ended"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

// Trivia state management functions

func (r *RedisClient) SetTriviaState(ctx context.Context, state string) error {
	return r.client.Set(ctx, TriviaStateKey, state, 0).Err()
}

func (r *RedisClient) GetTriviaState(ctx context.Context) (string, error) {
	return r.client.Get(ctx, TriviaStateKey).Result()
}

func (r *RedisClient) RegisterTriviaUser(ctx context.Context, userID string) error {
	return r.client.SAdd(ctx, TriviaUsersKey, userID).Err()
}

func (r *RedisClient) UnregisterTriviaUser(ctx context.Context, userID string) error {
	return r.client.SRem(ctx, TriviaUsersKey, userID).Err()
}

func (r *RedisClient) GetRegisteredTriviaUsers(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, TriviaUsersKey).Result()
}

func (r *RedisClient) ClearTriviaUsers(ctx context.Context) error {
	return r.client.Del(ctx, TriviaUsersKey).Err()
}

func (r *RedisClient) UpdateTriviaScore(ctx context.Context, userID string, score int) error {
	return r.client.ZIncrBy(ctx, TriviaScoresKey, float64(score), userID).Err()
}

func (r *RedisClient) GetTriviaScores(ctx context.Context) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(ctx, TriviaScoresKey, 0, -1).Result()
}

func (r *RedisClient) ClearTriviaScores(ctx context.Context) error {
	return r.client.Del(ctx, TriviaScoresKey).Err()
}

func (r *RedisClient) SetTriviaQuestions(ctx context.Context, questions []string) error {
	// Store questions as a list
	pipe := r.client.Pipeline()
	pipe.Del(ctx, TriviaQuestionsKey)
	for _, q := range questions {
		pipe.RPush(ctx, TriviaQuestionsKey, q)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisClient) GetNextTriviaQuestion(ctx context.Context) (string, error) {
	// Pop the next question from the list
	question, err := r.client.LPop(ctx, TriviaQuestionsKey).Result()
	if err != nil {
		return "", err
	}

	// Store current question
	err = r.client.Set(ctx, TriviaCurrentQuestionKey, question, 0).Err()
	return question, err
}

func (r *RedisClient) GetCurrentQuestion(ctx context.Context) (string, error) {
	return r.client.Get(ctx, TriviaCurrentQuestionKey).Result()
}

func (r *RedisClient) GetRemainingQuestionCount(ctx context.Context) (int64, error) {
	return r.client.LLen(ctx, TriviaQuestionsKey).Result()
}

// Cleanup functions

func (r *RedisClient) CleanupTrivia(ctx context.Context) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, TriviaStateKey)
	pipe.Del(ctx, TriviaUsersKey)
	pipe.Del(ctx, TriviaScoresKey)
	pipe.Del(ctx, TriviaQuestionsKey)
	pipe.Del(ctx, TriviaCurrentQuestionKey)
	_, err := pipe.Exec(ctx)
	return err
}

// Close connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
